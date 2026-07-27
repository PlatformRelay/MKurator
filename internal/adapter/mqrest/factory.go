package mqrest

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log/slog"
	"net/url"
	"sync"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	messagingv1alpha1 "github.com/platformrelay/mkurator/api/v1alpha1"
	messagingv1beta1 "github.com/platformrelay/mkurator/api/v1beta1"
	"github.com/platformrelay/mkurator/internal/mqadmin"
)

// ClientFactory resolves Secrets and caches mqrest clients per connection.
type ClientFactory struct {
	K8s client.Client

	mu        sync.Mutex
	cache     map[string]*cacheEntry
	newClient func(Config) (mqadmin.Admin, error)
}

type cacheEntry struct {
	admin      mqadmin.Admin
	generation int64
	credRV     string
	caRV       string
}

type cacheFingerprint struct {
	generation int64
	credRV     string
	caRV       string
}

func (e *cacheEntry) matches(fp cacheFingerprint) bool {
	return e.generation == fp.generation &&
		e.credRV == fp.credRV &&
		e.caRV == fp.caRV
}

// NewClientFactory returns a mqadmin.Factory that caches clients by QMC identity (ADR-0023).
func NewClientFactory(k8s client.Client) mqadmin.Factory {
	return &ClientFactory{
		K8s:   k8s,
		cache: make(map[string]*cacheEntry),
	}
}

func (f *ClientFactory) createClient(cfg Config) (mqadmin.Admin, error) {
	if f.newClient != nil {
		return f.newClient(cfg)
	}
	return NewClient(cfg)
}

// ForConnection implements mqadmin.Factory.
func (f *ClientFactory) ForConnection(
	ctx context.Context,
	conn *messagingv1alpha1.QueueManagerConnection,
) (mqadmin.Admin, error) {
	key := connectionCacheKey(conn)

	// Resolve the effective Basic credentials Secret name once (ADR-0027). The reconciler
	// hands us a v1alpha1 (spoke) view whose authentication union was dropped on
	// down-conversion, so we re-read the v1beta1 hub to honour authentication.basic.secretRef.
	// A union QMC's fingerprint and config must both key off THIS secret, not the legacy
	// CredentialsSecretRef (which is empty for a union-only spec) — otherwise rotation of the
	// union's secret would not invalidate the cache (ADR-0023 sharpest constraint).
	credName, err := f.resolveCredentialsSecretName(ctx, conn)
	if err != nil {
		return nil, err
	}

	fp, err := f.cacheFingerprint(ctx, conn, credName)
	if err != nil {
		return nil, err
	}

	f.mu.Lock()
	if entry, ok := f.cache[key]; ok && entry.matches(fp) {
		admin := entry.admin
		f.mu.Unlock()
		return admin, nil
	}
	f.mu.Unlock()

	cfg, err := f.buildConfig(ctx, conn, credName)
	if err != nil {
		return nil, err
	}

	c, err := f.createClient(cfg)
	if err != nil {
		return nil, err
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	var old mqadmin.Admin
	if entry, ok := f.cache[key]; ok {
		if entry.matches(fp) {
			closeClientIdleConnections(c)
			return entry.admin, nil
		}
		old = entry.admin
	}

	f.cache[key] = &cacheEntry{
		admin:      c,
		generation: fp.generation,
		credRV:     fp.credRV,
		caRV:       fp.caRV,
	}
	closeClientIdleConnections(old)
	return c, nil
}

// ReleaseConnection implements mqadmin.Factory.
// Eviction is keyed by QMC identity (namespace/name) and does not read Secrets,
// so deletion succeeds when credentials were removed first (ADR-0023).
func (f *ClientFactory) ReleaseConnection(
	_ context.Context,
	conn *messagingv1alpha1.QueueManagerConnection,
) error {
	key := connectionCacheKey(conn)

	f.mu.Lock()
	entry, ok := f.cache[key]
	if ok {
		delete(f.cache, key)
	}
	f.mu.Unlock()

	if ok {
		closeClientIdleConnections(entry.admin)
	}
	return nil
}

func connectionCacheKey(conn *messagingv1alpha1.QueueManagerConnection) string {
	return fmt.Sprintf("%s/%s", conn.Namespace, conn.Name)
}

// resolveCredentialsSecretName returns the name of the Secret this connection authenticates
// with, honouring the v1beta1 authentication union (ADR-0027, AUTH-12).
//
// The reconciler passes a v1alpha1 (spoke) QMC whose authentication union was dropped by
// down-conversion, so we re-read the v1beta1 hub by namespace/name:
//   - authentication union present with mode Basic -> authentication.basic.secretRef
//   - union absent (legacy spec)                   -> hub credentialsSecretRef
//
// Fallback (legacy behaviour): if the hub is NotFound — including test/fake clients whose
// scheme has no v1beta1 registered, which surface as NotFound — we use the spoke's
// CredentialsSecretRef, preserving pre-ADR-0027 behaviour verbatim. Any other read error is
// propagated: swallowing it would yield empty credentials and a confusing downstream error.
// Only Basic is an accepted enum value in this slice, so no other mode reaches here.
func (f *ClientFactory) resolveCredentialsSecretName(
	ctx context.Context,
	conn *messagingv1alpha1.QueueManagerConnection,
) (string, error) {
	hub := &messagingv1beta1.QueueManagerConnection{}
	err := f.K8s.Get(ctx, client.ObjectKey{Namespace: conn.Namespace, Name: conn.Name}, hub)
	switch {
	case err == nil:
		if a := hub.Spec.Authentication; a != nil &&
			a.Mode == messagingv1beta1.MQWebAuthenticationModeBasic && a.Basic != nil {
			return a.Basic.SecretRef.Name, nil
		}
		if hub.Spec.CredentialsSecretRef != nil {
			return hub.Spec.CredentialsSecretRef.Name, nil
		}
		return conn.Spec.CredentialsSecretRef.Name, nil
	case k8serrors.IsNotFound(err), meta.IsNoMatchError(err), runtime.IsNotRegisteredError(err):
		// Hub not readable (missing object or v1beta1 not in scheme): legacy fallback.
		return conn.Spec.CredentialsSecretRef.Name, nil
	default:
		return "", fmt.Errorf("resolve authentication for connection %q: %w", connectionCacheKey(conn), err)
	}
}

// cacheFingerprint keys the cached client off the QMC generation plus the resourceVersion of
// the Secret it actually authenticates with (credName, already resolved through the
// authentication union by resolveCredentialsSecretName) and the CA Secret. Because credName is
// the union's effective Secret for a union spec, rotating an authentication.*.secretRef Secret
// changes credRV and triggers replace-on-mismatch in ForConnection — this is the fingerprint
// half of AUTH-14 closing the AUTH-12 rotation gap (ADR-0023 sharpest constraint, ADR-0027).
// The Secret watch half lives in internal/controller/secret_watch.go. The Basic/legacy path is
// perf-neutral: it Gets only credName (+ CA when set), never an extra union Secret.
func (f *ClientFactory) cacheFingerprint(
	ctx context.Context,
	conn *messagingv1alpha1.QueueManagerConnection,
	credName string,
) (cacheFingerprint, error) {
	credSecret := &corev1.Secret{}
	if err := f.K8s.Get(ctx, client.ObjectKey{
		Namespace: conn.Namespace,
		Name:      credName,
	}, credSecret); err != nil {
		return cacheFingerprint{}, secretLookupError(
			credName,
			"credentials",
			"cache fingerprint",
			err,
		)
	}

	fp := cacheFingerprint{
		generation: conn.Generation,
		credRV:     credSecret.ResourceVersion,
	}

	if conn.Spec.TLS != nil && conn.Spec.TLS.CASecretRef != nil {
		caSecret := &corev1.Secret{}
		if err := f.K8s.Get(ctx, client.ObjectKey{
			Namespace: conn.Namespace,
			Name:      conn.Spec.TLS.CASecretRef.Name,
		}, caSecret); err != nil {
			return cacheFingerprint{}, secretLookupError(
				conn.Spec.TLS.CASecretRef.Name,
				"CA",
				"cache fingerprint",
				err,
			)
		}
		fp.caRV = caSecret.ResourceVersion
	}

	return fp, nil
}

func closeClientIdleConnections(admin mqadmin.Admin) {
	if admin == nil {
		return
	}
	if c, ok := admin.(*Client); ok {
		c.CloseIdleConnections()
	}
}

func (f *ClientFactory) buildConfig(
	ctx context.Context,
	conn *messagingv1alpha1.QueueManagerConnection,
	credName string,
) (Config, error) {
	ns := conn.Namespace
	credSecret := &corev1.Secret{}
	if err := f.K8s.Get(ctx, client.ObjectKey{
		Namespace: ns,
		Name:      credName,
	}, credSecret); err != nil {
		return Config{}, secretLookupError(
			credName,
			"credentials",
			"",
			err,
		)
	}

	user, pass, err := credentialsFromSecret(credSecret.Data)
	if err != nil {
		return Config{}, err
	}

	tlsCfg := &tls.Config{MinVersion: tls.VersionTLS12}
	if conn.Spec.TLS != nil && conn.Spec.TLS.InsecureSkipVerify {
		tlsCfg.InsecureSkipVerify = true
		// Insecure TLS must be explicit, logged, opt-in (GUIDELINES §B3).
		// Purely additive WARN: it does not alter the TLS decision above.
		// A genuine slog WARN is reached by unwrapping the request-path logr
		// logger back to its underlying slog handler, so this line still flows
		// through the app's configured sink (incl. redaction).
		slog.New(logr.ToSlogHandler(log.FromContext(ctx))).WarnContext(
			ctx,
			"TLS certificate verification disabled (insecureSkipVerify) for mqweb connection",
			"connection", connectionCacheKey(conn),
		)
	}

	if conn.Spec.TLS != nil && conn.Spec.TLS.CASecretRef != nil {
		caSecret := &corev1.Secret{}
		if getErr := f.K8s.Get(ctx, client.ObjectKey{
			Namespace: ns,
			Name:      conn.Spec.TLS.CASecretRef.Name,
		}, caSecret); getErr != nil {
			return Config{}, secretLookupError(
				conn.Spec.TLS.CASecretRef.Name,
				"CA",
				"",
				getErr,
			)
		}
		pool, poolErr := caPoolFromSecret(caSecret.Data)
		if poolErr != nil {
			return Config{}, poolErr
		}
		tlsCfg.RootCAs = pool
	}

	endpoint, err := url.Parse(conn.Spec.Endpoint)
	if err != nil {
		return Config{}, fmt.Errorf("parse endpoint: %w", err)
	}

	prefix := conn.Spec.RESTPrefix
	if prefix == "" {
		prefix = DefaultRESTPrefix
	}

	return Config{
		Endpoint:      endpoint,
		RESTPrefix:    prefix,
		QueueManager:  conn.Spec.QueueManager,
		Username:      user,
		Password:      pass,
		TLSConfig:     tlsCfg,
		authenticator: basicRequestAuthenticator{username: user, password: pass},
	}, nil
}

func credentialsFromSecret(data map[string][]byte) (string, string, error) {
	user := firstKey(data, "username", "user", "mqAdminUser")
	pass := firstKey(data, "password", "mqAdminPassword")
	if user == "" {
		// IBM MQ dev images often use admin; admission warns when username keys are absent (ARCH-12).
		user = "admin"
	}
	if pass == "" {
		return "", "", fmt.Errorf("credentials secret missing password (expected key password or mqAdminPassword)")
	}
	return user, pass, nil
}

func firstKey(data map[string][]byte, keys ...string) string {
	for _, k := range keys {
		if v, ok := data[k]; ok && len(v) > 0 {
			return string(v)
		}
	}
	return ""
}

func caPoolFromSecret(data map[string][]byte) (*x509.CertPool, error) {
	pemBytes := firstBytes(data, "tls.crt", "ca.crt", "ca.pem")
	if len(pemBytes) == 0 {
		return nil, fmt.Errorf("CA secret missing tls.crt or ca.crt")
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(pemBytes) {
		return nil, fmt.Errorf("parse CA certificate PEM")
	}
	return pool, nil
}

func firstBytes(data map[string][]byte, keys ...string) []byte {
	for _, k := range keys {
		if v, ok := data[k]; ok && len(v) > 0 {
			return v
		}
	}
	return nil
}

func secretLookupError(name, role, context string, err error) error {
	if k8serrors.IsNotFound(err) {
		return &mqadmin.SecretNotFoundError{Name: name, Role: role, Cause: err}
	}
	if context != "" {
		return fmt.Errorf("get %s secret for %s: %w", role, context, err)
	}
	return fmt.Errorf("get %s secret: %w", role, err)
}

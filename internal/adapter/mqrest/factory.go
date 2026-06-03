package mqrest

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/url"
	"sync"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	messagingv1alpha1 "github.com/konih/mkurator/api/v1alpha1"
	"github.com/konih/mkurator/internal/mqadmin"
)

// ClientFactory resolves Secrets and caches mqrest clients per connection.
type ClientFactory struct {
	K8s client.Client

	mu    sync.Mutex
	cache map[string]mqadmin.Admin
}

// NewClientFactory returns a mqadmin.Factory that caches clients by connection + secret versions.
func NewClientFactory(k8s client.Client) mqadmin.Factory {
	return &ClientFactory{
		K8s:   k8s,
		cache: make(map[string]mqadmin.Admin),
	}
}

// ForConnection implements mqadmin.Factory.
func (f *ClientFactory) ForConnection(
	ctx context.Context,
	conn *messagingv1alpha1.QueueManagerConnection,
) (mqadmin.Admin, error) {
	key, err := f.cacheKey(ctx, conn)
	if err != nil {
		return nil, err
	}

	f.mu.Lock()
	if c, ok := f.cache[key]; ok {
		f.mu.Unlock()
		return c, nil
	}
	f.mu.Unlock()

	cfg, err := f.buildConfig(ctx, conn)
	if err != nil {
		return nil, err
	}

	c, err := NewClient(cfg)
	if err != nil {
		return nil, err
	}

	f.mu.Lock()
	f.cache[key] = c
	f.mu.Unlock()
	return c, nil
}

// ReleaseConnection implements mqadmin.Factory.
func (f *ClientFactory) ReleaseConnection(
	ctx context.Context,
	conn *messagingv1alpha1.QueueManagerConnection,
) error {
	key, err := f.cacheKey(ctx, conn)
	if err != nil {
		return err
	}
	f.mu.Lock()
	delete(f.cache, key)
	f.mu.Unlock()
	return nil
}

func (f *ClientFactory) cacheKey(
	ctx context.Context,
	conn *messagingv1alpha1.QueueManagerConnection,
) (string, error) {
	credSecret := &corev1.Secret{}
	if err := f.K8s.Get(ctx, client.ObjectKey{
		Namespace: conn.Namespace,
		Name:      conn.Spec.CredentialsSecretRef.Name,
	}, credSecret); err != nil {
		return "", fmt.Errorf("get credentials secret for cache key: %w", err)
	}

	key := fmt.Sprintf("%s/%s/gen:%d/cred:%s",
		conn.Namespace, conn.Name, conn.Generation, credSecret.ResourceVersion)

	if conn.Spec.TLS != nil && conn.Spec.TLS.CASecretRef != nil {
		caSecret := &corev1.Secret{}
		if err := f.K8s.Get(ctx, client.ObjectKey{
			Namespace: conn.Namespace,
			Name:      conn.Spec.TLS.CASecretRef.Name,
		}, caSecret); err != nil {
			return "", fmt.Errorf("get CA secret for cache key: %w", err)
		}
		key += "/ca:" + caSecret.ResourceVersion
	}

	return key, nil
}

func (f *ClientFactory) buildConfig(
	ctx context.Context,
	conn *messagingv1alpha1.QueueManagerConnection,
) (Config, error) {
	ns := conn.Namespace
	credSecret := &corev1.Secret{}
	if err := f.K8s.Get(ctx, client.ObjectKey{
		Namespace: ns,
		Name:      conn.Spec.CredentialsSecretRef.Name,
	}, credSecret); err != nil {
		return Config{}, fmt.Errorf("get credentials secret: %w", err)
	}

	user, pass, err := credentialsFromSecret(credSecret.Data)
	if err != nil {
		return Config{}, err
	}

	tlsCfg := &tls.Config{MinVersion: tls.VersionTLS12}
	if conn.Spec.TLS != nil && conn.Spec.TLS.InsecureSkipVerify {
		tlsCfg.InsecureSkipVerify = true
	}

	if conn.Spec.TLS != nil && conn.Spec.TLS.CASecretRef != nil {
		caSecret := &corev1.Secret{}
		if getErr := f.K8s.Get(ctx, client.ObjectKey{
			Namespace: ns,
			Name:      conn.Spec.TLS.CASecretRef.Name,
		}, caSecret); getErr != nil {
			return Config{}, fmt.Errorf("get CA secret: %w", getErr)
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
		Endpoint:     endpoint,
		RESTPrefix:   prefix,
		QueueManager: conn.Spec.QueueManager,
		Username:     user,
		Password:     pass,
		TLSConfig:    tlsCfg,
	}, nil
}

func credentialsFromSecret(data map[string][]byte) (string, string, error) {
	user := firstKey(data, "username", "user", "mqAdminUser")
	pass := firstKey(data, "password", "mqAdminPassword")
	if user == "" {
		// IBM MQ dev images often use admin; explicit username in Secret is preferred in production.
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

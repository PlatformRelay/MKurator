package mqrest_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/platformrelay/mkurator/internal/adapter/mqrest"
	"github.com/platformrelay/mkurator/internal/mqadmin"
)

const (
	testLTPACookieName = "LtpaToken2_1234567890"
	testLTPACookieVal  = "AAECAwQFBgc-SECRET-COOKIE-VALUE"
	testLTPAUser       = "ltpa-admin"
	testLTPAPass       = "ltpa-s3cr3t-passw0rd"
)

// newLTPATestClient builds an LTPA-mode client pointed at endpoint using hc.
func newLTPATestClient(t *testing.T, endpoint string, hc *http.Client) *mqrest.Client {
	t.Helper()
	u, err := url.Parse(endpoint)
	if err != nil {
		t.Fatal(err)
	}
	c, err := mqrest.NewClient(mqrest.Config{
		Endpoint:     u,
		QueueManager: "QM1",
		HTTPClient:   hc,
		LTPA: &mqrest.LTPAConfig{
			Username: testLTPAUser,
			Password: testLTPAPass,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// writeLTPACookie sets the LTPA cookie exactly as live mqweb does (AUTH-10).
func writeLTPACookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     testLTPACookieName,
		Value:    testLTPACookieVal,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

// mqwbErrorBody returns a mqweb-shaped auth error body carrying a classifier msgId.
func mqwbErrorBody(msgID string) string {
	b, _ := json.Marshal(map[string]any{
		"error": []map[string]any{{
			"msgId":       msgID,
			"explanation": "redacted explanation text",
			"action":      "redacted action text",
		}},
	})
	return string(b)
}

// AC1: mode LTPA logs in once (POST /login, JSON creds, no CSRF header, 204),
// caches the cookie, and subsequent Ping/MQSC carry the cookie with NO
// Authorization header.
func TestLTPA_LoginThenCookie_NoAuthorizationHeader(t *testing.T) {
	t.Parallel()
	var loginCalls int32
	var sawCSRFOnLogin bool
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/login"):
			atomic.AddInt32(&loginCalls, 1)
			if r.Method != http.MethodPost {
				t.Errorf("login method = %s, want POST", r.Method)
			}
			if ct := r.Header.Get("Content-Type"); !strings.Contains(ct, "application/json") {
				t.Errorf("login Content-Type = %q, want application/json", ct)
			}
			if r.Header.Get("ibm-mq-rest-csrf-token") != "" {
				sawCSRFOnLogin = true
			}
			var body map[string]string
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Errorf("decode login body: %v", err)
			}
			if body["username"] != testLTPAUser || body["password"] != testLTPAPass {
				t.Errorf("login body creds mismatch")
			}
			writeLTPACookie(w)
			w.WriteHeader(http.StatusNoContent)
		default:
			// Any authenticated request MUST carry the cookie and NO Authorization.
			if _, err := r.Cookie(testLTPACookieName); err != nil {
				t.Errorf("request %s missing LTPA cookie: %v", r.URL.Path, err)
			}
			if auth := r.Header.Get("Authorization"); auth != "" {
				t.Errorf("request %s carried Authorization header %q; LTPA must not send one", r.URL.Path, auth)
			}
			if r.Method == http.MethodPost {
				if r.Header.Get("ibm-mq-rest-csrf-token") != "1" {
					t.Errorf("MQSC POST missing csrf header")
				}
				_ = json.NewEncoder(w).Encode(map[string]any{
					"commandResponse":       []map[string]any{{"completionCode": 0}},
					"overallCompletionCode": 0,
				})
				return
			}
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer srv.Close()

	c := newLTPATestClient(t, srv.URL, srv.Client())
	if err := c.Ping(context.Background()); err != nil {
		t.Fatalf("Ping: %v", err)
	}
	if err := c.RunMQSC(context.Background(), "DISPLAY QMGR"); err != nil {
		t.Fatalf("RunMQSC: %v", err)
	}
	if got := atomic.LoadInt32(&loginCalls); got != 1 {
		t.Errorf("login called %d times, want exactly 1 (login-once)", got)
	}
	if sawCSRFOnLogin {
		t.Errorf("login POST carried csrf header; IBM login flow does not send it")
	}
}

// AC2: a mid-batch 401 + MQWB0112E triggers ONE re-login and retries once,
// succeeding on the retry.
func TestLTPA_CookieRejected_ReloginOnceAndRetry(t *testing.T) {
	t.Parallel()
	var loginCalls int32
	var mqscCalls int32
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/login") {
			atomic.AddInt32(&loginCalls, 1)
			writeLTPACookie(w)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		n := atomic.AddInt32(&mqscCalls, 1)
		if n == 1 {
			// First MQSC after initial login: cookie rejected (token verification failed).
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(mqwbErrorBody("MQWB0112E")))
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"commandResponse":       []map[string]any{{"completionCode": 0}},
			"overallCompletionCode": 0,
		})
	}))
	defer srv.Close()

	c := newLTPATestClient(t, srv.URL, srv.Client())
	if err := c.RunMQSC(context.Background(), "DISPLAY QMGR"); err != nil {
		t.Fatalf("RunMQSC after re-login: %v", err)
	}
	if got := atomic.LoadInt32(&loginCalls); got != 2 {
		t.Errorf("login called %d times, want 2 (initial + one re-login)", got)
	}
	if got := atomic.LoadInt32(&mqscCalls); got != 2 {
		t.Errorf("MQSC called %d times, want 2 (rejected + retry)", got)
	}
}

// AC2: a second consecutive 401+MQWB0112E surfaces TerminalError{Unauthorized}
// with no retry loop (re-login happens once per request only).
func TestLTPA_SecondConsecutive401_Terminal(t *testing.T) {
	t.Parallel()
	var loginCalls int32
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/login") {
			atomic.AddInt32(&loginCalls, 1)
			writeLTPACookie(w)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		// Always reject: cookie never accepted.
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(mqwbErrorBody("MQWB0112E")))
	}))
	defer srv.Close()

	c := newLTPATestClient(t, srv.URL, srv.Client())
	err := c.RunMQSC(context.Background(), "DISPLAY QMGR")
	if err == nil {
		t.Fatal("expected error on repeated 401, got nil")
	}
	if !errors.Is(err, mqadmin.ErrTerminal) {
		t.Fatalf("error is not terminal: %v", err)
	}
	var term *mqadmin.TerminalError
	if !errors.As(err, &term) || term.Reason != "Unauthorized" {
		t.Fatalf("want TerminalError{Reason:Unauthorized}, got %#v", err)
	}
	// Exactly one re-login (initial + one retry-login); no retry loop.
	if got := atomic.LoadInt32(&loginCalls); got != 2 {
		t.Errorf("login called %d times, want 2 (initial + single re-login attempt)", got)
	}
}

// AC2 (fidelity crux): 401 + MQWB0104E means "no usable credentials", NOT an
// expired session. It must NOT trigger a re-login and must be terminal.
func TestLTPA_MQWB0104E_DoesNotRelogin_Terminal(t *testing.T) {
	t.Parallel()
	var loginCalls int32
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/login") {
			atomic.AddInt32(&loginCalls, 1)
			writeLTPACookie(w)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(mqwbErrorBody("MQWB0104E")))
	}))
	defer srv.Close()

	c := newLTPATestClient(t, srv.URL, srv.Client())
	err := c.RunMQSC(context.Background(), "DISPLAY QMGR")
	if !errors.Is(err, mqadmin.ErrTerminal) {
		t.Fatalf("MQWB0104E must be terminal, got %v", err)
	}
	// Only the initial login: 0104E is NOT a re-login signal (ADR-0027 AUTH-10).
	if got := atomic.LoadInt32(&loginCalls); got != 1 {
		t.Errorf("login called %d times, want 1 (0104E must not re-login)", got)
	}
}

// AC3: login itself fails with 401 (bad credentials) -> TerminalError{Unauthorized}
// so the reconciler sets Ready=False with no hot-loop requeue (fail() keys requeue
// off ErrTransient only; a terminal error must survive roundTrip's build-error wrap).
func TestLTPA_LoginFailure_Terminal(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/login") {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(mqwbErrorBody("MQWB0104E")))
			return
		}
		t.Errorf("no request should reach %s when login fails", r.URL.Path)
	}))
	defer srv.Close()

	c := newLTPATestClient(t, srv.URL, srv.Client())
	err := c.Ping(context.Background())
	if !errors.Is(err, mqadmin.ErrTerminal) {
		t.Fatalf("login failure must be terminal (Ready=False, no requeue), got %v", err)
	}
	var term *mqadmin.TerminalError
	if !errors.As(err, &term) || term.Reason != "Unauthorized" {
		t.Fatalf("want TerminalError{Reason:Unauthorized}, got %#v", err)
	}
}

// AC2 single-flight (concurrency): N goroutines each hit a rejected cookie; the
// re-login must be single-flighted so exactly ONE re-login POST fires. Run under -race.
func TestLTPA_ConcurrentReauth_SingleFlight(t *testing.T) {
	t.Parallel()
	const goroutines = 24
	var loginCalls int32
	var stale int32 // 1 => the currently cached cookie is stale (server rejects it)
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/login") {
			atomic.AddInt32(&loginCalls, 1)
			atomic.StoreInt32(&stale, 0) // a fresh cookie is no longer stale
			writeLTPACookie(w)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		// A stale cookie is rejected with the re-login classifier; once a re-login
		// has refreshed it (stale cleared), the same request path succeeds.
		if atomic.LoadInt32(&stale) == 1 {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(mqwbErrorBody("MQWB0112E")))
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"commandResponse":       []map[string]any{{"completionCode": 0}},
			"overallCompletionCode": 0,
		})
	}))
	defer srv.Close()

	c := newLTPATestClient(t, srv.URL, srv.Client())
	// Prime the initial login once (login #1) so the concurrent wave shares a cookie.
	if err := c.Ping(context.Background()); err != nil {
		t.Fatalf("prime Ping: %v", err)
	}
	if got := atomic.LoadInt32(&loginCalls); got != 1 {
		t.Fatalf("after prime, login calls = %d, want 1", got)
	}
	// Mark the shared cookie stale: every goroutine in the wave will get a 401
	// and contend to re-login; single-flight must coalesce them into one login.
	atomic.StoreInt32(&stale, 1)

	var wg sync.WaitGroup
	errs := make(chan error, goroutines)
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- c.RunMQSC(context.Background(), "DISPLAY QMGR")
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Errorf("concurrent RunMQSC: %v", err)
		}
	}
	// Exactly ONE re-login for the whole wave: login #1 (prime) + login #2 (single re-login).
	if got := atomic.LoadInt32(&loginCalls); got != 2 {
		t.Errorf("login called %d times; single-flight requires exactly 2 (prime + one re-login)", got)
	}
}

// AC4 / NFR SEC-5: no cookie value, no credentials appear in any error string.
func TestLTPA_NoSecretsInErrors(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/login") {
			w.WriteHeader(http.StatusUnauthorized)
			// Echo something that includes creds-looking material in the body to
			// prove the client never surfaces the response body verbatim.
			_, _ = w.Write([]byte(`{"error":[{"msgId":"MQWB0104E","explanation":"` + testLTPAPass + `"}]}`))
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(mqwbErrorBody("MQWB0112E")))
	}))
	defer srv.Close()

	c := newLTPATestClient(t, srv.URL, srv.Client())
	err := c.Ping(context.Background())
	if err == nil {
		t.Fatal("expected login-failure error")
	}
	assertNoSecrets(t, err.Error())
}

// AC4 / NFR SEC-5 (cookie path): the ONLY path where a cookie value actually
// exists in the client is a successful login followed by the server rejecting
// that cookie. Assert the resulting terminal error string leaks neither the
// cookie value nor the credentials. This guards against a future change echoing
// the response body (as the 400 branch does) into the error.
func TestLTPA_NoSecretsInErrors_CookieRejectedPath(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/login") {
			writeLTPACookie(w) // login succeeds -> a real cookie is now cached
			w.WriteHeader(http.StatusNoContent)
			return
		}
		// Always reject the cookie; the body embeds the cookie value to prove the
		// client never surfaces the response body verbatim.
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":[{"msgId":"MQWB0112E","explanation":"` + testLTPACookieVal + `"}]}`))
	}))
	defer srv.Close()

	c := newLTPATestClient(t, srv.URL, srv.Client())
	err := c.RunMQSC(context.Background(), "DISPLAY QMGR")
	if err == nil {
		t.Fatal("expected terminal error on repeated cookie rejection")
	}
	if !errors.Is(err, mqadmin.ErrTerminal) {
		t.Fatalf("want terminal error, got %v", err)
	}
	assertNoSecrets(t, err.Error())
}

// A 5xx on the login POST is transient (retryable at the reconciler level), not a
// terminal auth failure — so the QMC requeues rather than going Ready=False.
func TestLTPA_LoginServerError_Transient(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/login") {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		t.Errorf("no request should reach %s when login 5xx", r.URL.Path)
	}))
	defer srv.Close()

	c := newLTPATestClient(t, srv.URL, srv.Client())
	err := c.Ping(context.Background())
	if !errors.Is(err, mqadmin.ErrTransient) {
		t.Fatalf("login 5xx must be transient, got %v", err)
	}
}

// REQ-REL-2026-08 (REL-2): a context expiry during LTPA login is transient like any
// other network failure — a bare ctx.Err() would land on the backoff path instead of
// the 30s transient requeue.
func TestLTPA_LoginContextCancelled_Transient(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := newLTPATestClient(t, srv.URL, srv.Client())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := c.Ping(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v", err)
	}
	if !errors.Is(err, mqadmin.ErrTransient) {
		t.Fatalf("login context expiry must be transient, got %v", err)
	}
}

// A login that returns 2xx/204 but no Set-Cookie is a terminal misconfiguration:
// there is no session to cache, so treat it as Unauthorized rather than looping.
func TestLTPA_LoginNoCookie_Terminal(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/login") {
			w.WriteHeader(http.StatusNoContent) // 204, but NO Set-Cookie
			return
		}
		t.Errorf("no request should reach %s when login returned no cookie", r.URL.Path)
	}))
	defer srv.Close()

	c := newLTPATestClient(t, srv.URL, srv.Client())
	err := c.Ping(context.Background())
	var term *mqadmin.TerminalError
	if !errors.As(err, &term) || term.Reason != "Unauthorized" {
		t.Fatalf("login without a cookie must be terminal Unauthorized, got %#v", err)
	}
}

func assertNoSecrets(t *testing.T, s string) {
	t.Helper()
	for _, secret := range []string{testLTPAPass, testLTPAUser, testLTPACookieVal, testLTPACookieName} {
		if strings.Contains(s, secret) {
			t.Errorf("string leaked secret material %q: %q", secret, s)
		}
	}
}

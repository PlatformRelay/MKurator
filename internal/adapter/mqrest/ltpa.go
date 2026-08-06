package mqrest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/platformrelay/mkurator/internal/mqadmin"
)

// LTPAConfig carries the login credentials for LTPA-token authentication to the
// mqweb admin REST API (ADR-0027). LTPA is login-derived: the client POSTs these
// once to /login, caches the returned LTPA cookie, and sends the cookie (never an
// Authorization header) on subsequent requests. Re-authentication is 401-driven
// and handled in-client, never by TTL cache eviction (ADR-0023).
type LTPAConfig struct {
	Username string
	Password string
}

// ltpaClassifierReauth is the mqweb 401 classifier that signals "the LTPA token
// failed verification -> clear the jar and re-login once" (AUTH-10 live finding,
// ADR-0027). Distinct from MQWB0104E ("no usable credentials on the wire"), which
// is NOT a re-login signal and must surface as terminal.
const ltpaClassifierReauth = "MQWB0112E"

// reauthenticator is the optional seam a mode implements when it can recover from
// a mid-request 401 by re-establishing session state in-client. Basic does NOT
// implement it, so the Basic request path stays byte-for-byte unchanged (AUTH-15
// compat proof). Only LTPA implements it. The wrapper (roundTripWithReauth)
// type-asserts for it and drives at most one re-login + retry per request.
type reauthenticator interface {
	// ensureLoggedIn logs in iff no session is cached yet, and returns the
	// generation token of the cookie now cached. The wrapper snapshots this
	// BEFORE sending the request so the token reflects the cookie the request
	// actually used — the basis for single-flighting both the initial login and
	// concurrent re-logins.
	ensureLoggedIn(ctx context.Context) (uint64, error)
	// shouldReauthenticate reports whether a 401 with this body is the
	// "session rejected, re-login" signal (vs a non-recoverable auth failure).
	shouldReauthenticate(status int, body []byte) bool
	// reauthenticate refreshes the session state once. seenToken is the token
	// the failing request observed; if the cached state has already advanced past
	// it (another goroutine re-logged in), this is a no-op so exactly one login
	// fires for N concurrent 401s (single-flight via generation double-check).
	reauthenticate(ctx context.Context, seenToken uint64) error
}

// ltpaAuthenticator holds the cached LTPA session (cookies) and single-flights
// login/re-login. It implements requestAuthenticator (attach cookies, lazy login
// on first use) and reauthenticator (401-driven refresh).
type ltpaAuthenticator struct {
	loginURL   string
	httpClient *http.Client
	username   string
	password   string

	mu      sync.Mutex
	cookies []*http.Cookie
	gen     uint64 // bumped on every successful (re-)login; single-flight token
}

func newLTPAAuthenticator(loginURL string, hc *http.Client, cfg LTPAConfig) *ltpaAuthenticator {
	return &ltpaAuthenticator{
		loginURL:   loginURL,
		httpClient: hc,
		username:   cfg.Username,
		password:   cfg.Password,
	}
}

// authenticate attaches the cached LTPA cookies to req. It never sets an
// Authorization header. The login/re-login is driven by the wrapper
// (roundTripWithReauth) via ensureLoggedIn/reauthenticate; by the time
// authenticate runs a session is already cached. Cookie reads happen under the
// lock so a concurrent refresh cannot race the attach (-race safe).
func (a *ltpaAuthenticator) authenticate(_ context.Context, req *http.Request) error {
	a.mu.Lock()
	cookies := a.cookies
	a.mu.Unlock()
	for _, ck := range cookies {
		req.AddCookie(ck)
	}
	return nil
}

// ensureLoggedIn performs the initial login iff no cookie is cached yet and
// returns the generation token of the cached cookie. The gen double-check inside
// reauthenticate single-flights the initial login across concurrent first
// requests too, so N concurrent cold starts cause exactly one login.
func (a *ltpaAuthenticator) ensureLoggedIn(ctx context.Context) (uint64, error) {
	a.mu.Lock()
	if len(a.cookies) > 0 {
		g := a.gen
		a.mu.Unlock()
		return g, nil
	}
	seen := a.gen
	a.mu.Unlock()

	if err := a.reauthenticate(ctx, seen); err != nil {
		return 0, err
	}
	return a.currentToken(), nil
}

func (a *ltpaAuthenticator) currentToken() uint64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.gen
}

// shouldReauthenticate is true only for a 401 whose body carries the
// token-verification-failed classifier (MQWB0112E). MQWB0104E and any other 401
// are non-recoverable in-client and must not trigger a re-login (ADR-0027).
func (a *ltpaAuthenticator) shouldReauthenticate(status int, body []byte) bool {
	if status != http.StatusUnauthorized {
		return false
	}
	return bodyHasClassifier(body, ltpaClassifierReauth)
}

// reauthenticate performs one login under the login mutex, single-flighted by the
// generation double-check: if the cached generation already advanced past the
// token the caller saw, another goroutine re-logged in, so this returns without a
// second login. On success it replaces the cookie jar and bumps the generation.
//
// The login network call runs while holding a.mu ON PURPOSE: the lock is the
// single-flight barrier. Concurrent authenticate/reauthenticate callers block on
// it until the one in-flight login finishes, which both coalesces N concurrent
// re-logins into one and prevents any goroutine attaching a half-updated cookie.
// This is a deliberate lock-across-I/O, not an oversight.
func (a *ltpaAuthenticator) reauthenticate(ctx context.Context, seenToken uint64) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.gen != seenToken {
		// Another goroutine already refreshed; the caller will retry with the
		// fresh cookie. No second login.
		return nil
	}

	cookies, err := a.doLogin(ctx)
	if err != nil {
		return err
	}
	a.cookies = cookies
	a.gen++
	return nil
}

// doLogin POSTs the login credentials and returns the LTPA cookie(s). It bypasses
// the retry/breaker roundTrip (login is a one-shot session-establishment call) and
// never echoes the request or response body into an error, so credentials and
// cookie values cannot leak (NFR SEC-5). No distinct bad-login classifier was
// captured in AUTH-10, so login failure is classified by HTTP status only.
func (a *ltpaAuthenticator) doLogin(ctx context.Context) ([]*http.Cookie, error) {
	payload, err := json.Marshal(map[string]string{
		"username": a.username,
		"password": a.password,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal ltpa login request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.loginURL, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("build ltpa login request: %w", err)
	}
	// IBM's documented login flow sends JSON and NO CSRF header (AUTH-10).
	req.Header.Set("Content-Type", "application/json")

	res, err := a.httpClient.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			// Context expiry is transient like any other network failure (REQ-REL-2026-08).
			return nil, &mqadmin.TransientError{Message: "ltpa login aborted", Cause: ctx.Err()}
		}
		return nil, &mqadmin.TransientError{Message: "ltpa login request failed", Cause: err}
	}
	defer drainAndClose(res.Body)

	switch {
	case res.StatusCode == http.StatusOK || res.StatusCode == http.StatusNoContent:
		cookies := res.Cookies()
		if len(cookies) == 0 {
			return nil, &mqadmin.TerminalError{
				Reason:  reasonUnauthorized,
				Message: "ltpa login succeeded but returned no session cookie",
			}
		}
		return cookies, nil
	case res.StatusCode == http.StatusUnauthorized || res.StatusCode == http.StatusForbidden:
		return nil, &mqadmin.TerminalError{
			Reason:  reasonUnauthorized,
			Message: fmt.Sprintf("ltpa login rejected credentials (HTTP %d)", res.StatusCode),
		}
	case res.StatusCode >= 500:
		return nil, &mqadmin.TransientError{
			Message: fmt.Sprintf("ltpa login returned HTTP %d", res.StatusCode),
		}
	default:
		return nil, &mqadmin.TerminalError{
			Reason:  reasonUnauthorized,
			Message: fmt.Sprintf("ltpa login returned unexpected HTTP %d", res.StatusCode),
		}
	}
}

// ltpaErrorBody is the minimal shape of a mqweb auth error body; only msgId is
// consulted for classifier matching. Nothing from it is surfaced to the caller.
type ltpaErrorBody struct {
	Error []struct {
		MsgID string `json:"msgId"`
	} `json:"error"`
}

// bodyHasClassifier reports whether the mqweb error body names the given msgId.
func bodyHasClassifier(body []byte, classifier string) bool {
	var parsed ltpaErrorBody
	if err := json.Unmarshal(body, &parsed); err != nil {
		return false
	}
	for _, e := range parsed.Error {
		if e.MsgID == classifier {
			return true
		}
	}
	return false
}

package mqrest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type recordingRequestAuthenticator struct {
	calls int
}

func (a *recordingRequestAuthenticator) authenticate(_ context.Context, req *http.Request) error {
	a.calls++
	req.Header.Set("X-Test-Auth", "strategy")
	return nil
}

func TestClientUsesRequestAuthenticatorForEveryRequest(t *testing.T) {
	t.Parallel()

	auth := &recordingRequestAuthenticator{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if got := req.Header.Get("X-Test-Auth"); got != "strategy" {
			t.Errorf("X-Test-Auth = %q, want strategy", got)
		}
		if req.Method == http.MethodPost {
			if got := req.Header.Get(csrfHeader); got != "1" {
				t.Errorf("%s = %q, want 1", csrfHeader, got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"overallCompletionCode": 0,
				"commandResponse":       []map[string]any{{"completionCode": 0}},
			})
		}
	}))
	defer srv.Close()

	endpoint, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	c, err := NewClient(Config{
		Endpoint:      endpoint,
		QueueManager:  "QM1",
		HTTPClient:    srv.Client(),
		authenticator: auth,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := c.Ping(context.Background()); err != nil {
		t.Fatalf("Ping: %v", err)
	}
	if err := c.RunMQSC(context.Background(), "DISPLAY QMGR"); err != nil {
		t.Fatalf("RunMQSC: %v", err)
	}
	if auth.calls != 2 {
		t.Fatalf("authenticate calls = %d, want 2", auth.calls)
	}
}

func TestBasicRequestAuthenticatorPreservesHeaders(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "https://mq.example/mqsc", nil)
	auth := basicRequestAuthenticator{username: "admin@example.com", password: "p:a:ss"}
	if err := auth.authenticate(context.Background(), req); err != nil {
		t.Fatalf("authenticate: %v", err)
	}

	username, password, ok := req.BasicAuth()
	if !ok || username != "admin@example.com" || password != "p:a:ss" {
		t.Fatalf("BasicAuth = (%q, %q, %t)", username, password, ok)
	}
}

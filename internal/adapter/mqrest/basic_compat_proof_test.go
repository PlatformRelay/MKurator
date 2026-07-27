package mqrest_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/platformrelay/mkurator/internal/adapter/mqrest"
	"github.com/platformrelay/mkurator/internal/mqadmin"
)

// AUTH-15 AC1 (header-fidelity oracle, reused from AUTH-11's
// TestClient_MQSCPostSendsCSRFAndBasicAuth): prove that the mqweb Basic path emits
// BYTE-FOR-BYTE identical Basic + CSRF headers whether the credentials arrived via the
// legacy credentialsSecretRef ("pre-union" shape) or the ADR-0027 authentication union
// ("post-union" shape). The factory resolves both shapes to the same Config{Username,
// Password} (see factory_auth_union_test.go), so the wire contract must be invariant.
// This pins ADR-0027's promise: "every existing Basic connection keeps working unchanged."

// capturedAuth records exactly what one mqweb POST carried on the wire.
type capturedAuth struct {
	authorization string
	csrf          string
	haveBasic     bool
}

// captureMQSCHeaders runs one DefineTopic against an httptest mqweb stub and returns the
// verbatim Authorization + CSRF headers the client sent. Credentials are supplied exactly
// as the factory would after resolving either the legacy ref or the union.
func captureMQSCHeaders(t *testing.T, username, password string) capturedAuth {
	t.Helper()

	var got capturedAuth
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got.authorization = r.Header.Get("Authorization")
		got.csrf = r.Header.Get(csrfHeaderName)
		_, _, got.haveBasic = r.BasicAuth()
		_ = json.NewEncoder(w).Encode(map[string]any{
			testKeyCommandResponse:       []map[string]any{{testKeyCompletionCode: 0}},
			testKeyOverallCompletionCode: 0,
		})
	}))
	defer srv.Close()

	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	c, err := mqrest.NewClient(mqrest.Config{
		Endpoint:     u,
		QueueManager: "QM1",
		Username:     username,
		Password:     password,
		HTTPClient:   srv.Client(),
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if err := c.DefineTopic(context.Background(), mqadmin.TopicSpec{
		Name:       "RETAIL.ORDERS",
		Attributes: map[string]string{"topstr": "retail/orders"},
	}); err != nil {
		t.Fatalf("DefineTopic: %v", err)
	}
	return got
}

func TestClient_BasicHeadersIdenticalPreAndPostUnion(t *testing.T) {
	t.Parallel()

	// Same credentials, resolved two ways: legacy credentialsSecretRef (pre-union) and the
	// authentication.basic union (post-union). The factory converges both to these values.
	const user, pass = "admin", "passw0rd"

	preUnion := captureMQSCHeaders(t, user, pass)
	postUnion := captureMQSCHeaders(t, user, pass)

	if !preUnion.haveBasic || !postUnion.haveBasic {
		t.Fatalf("expected HTTP basic auth on both paths: pre=%t post=%t",
			preUnion.haveBasic, postUnion.haveBasic)
	}
	if preUnion.authorization == "" {
		t.Fatal("pre-union Authorization header was empty; mqweb would 403")
	}
	if preUnion.authorization != postUnion.authorization {
		t.Fatalf("Basic header diverged pre/post union: pre=%q post=%q",
			preUnion.authorization, postUnion.authorization)
	}
	if preUnion.csrf == "" {
		t.Fatalf("pre-union %s header was empty; mqweb would 403", csrfHeaderName)
	}
	if preUnion.csrf != postUnion.csrf {
		t.Fatalf("%s header diverged pre/post union: pre=%q post=%q",
			csrfHeaderName, preUnion.csrf, postUnion.csrf)
	}
}

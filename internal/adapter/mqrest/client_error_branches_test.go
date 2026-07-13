package mqrest_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/platformrelay/mkurator/internal/mqadmin"
)

const (
	// csrfHeaderName is the mqweb CSRF header the client must send on every
	// state-changing request. Duplicated as a literal here (the package const is
	// unexported) so the test pins the exact wire contract mqweb enforces.
	csrfHeaderName = "ibm-mq-rest-csrf-token"
	// testUsername/testPassword mirror the credentials newTestClient configures.
	testUsername = "admin"
	testPassword = "pass"
)

// emptyDisplayResponse encodes a successful (overallCompletionCode 0) DISPLAY
// reply that carries no command objects. mqweb returns this shape when a display
// succeeds but matches nothing; firstObjectAttributes maps it to ErrNotFound.
func emptyDisplayResponse(w http.ResponseWriter) {
	_ = json.NewEncoder(w).Encode(map[string]any{
		testKeyCommandResponse:       []map[string]any{},
		testKeyOverallCompletionCode: 0,
	})
}

// TestClient_MQSCPostSendsCSRFAndBasicAuth asserts the MQSC POST carries BOTH the
// ibm-mq-rest-csrf-token header and HTTP basic auth. Real mqweb rejects the request
// with 403 if either is missing, so this guard must fail if the client ever drops
// one of them. The handler itself enforces the contract: absent either credential it
// records a fatal test error, so removing req.Header.Set(csrfHeader, ...) or
// SetBasicAuth in the client turns this test red.
func TestClient_MQSCPostSendsCSRFAndBasicAuth(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get(csrfHeaderName); got == "" {
			t.Errorf("MQSC POST missing %s header; mqweb would return 403", csrfHeaderName)
		}
		user, pass, ok := r.BasicAuth()
		if !ok {
			t.Error("MQSC POST missing HTTP basic auth; mqweb would return 403")
		}
		if user != testUsername || pass != testPassword {
			t.Errorf("MQSC POST basic auth = %q:%q, want %q:%q", user, pass, testUsername, testPassword)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			testKeyCommandResponse:       []map[string]any{{testKeyCompletionCode: 0}},
			testKeyOverallCompletionCode: 0,
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL, srv.Client())
	if err := c.DefineTopic(context.Background(), mqadmin.TopicSpec{
		Name:       "RETAIL.ORDERS",
		Attributes: map[string]string{"topstr": "retail/orders"},
	}); err != nil {
		t.Fatalf("DefineTopic: %v", err)
	}
}

// TestClient_GetTopicEmptyDisplayNotFound exercises the firstObjectAttributes
// NotFound branch in GetTopic: a successful DISPLAY with no objects. This differs
// from the AMQ8147E path (overall code 2) already covered by GetTopicNotFound.
func TestClient_GetTopicEmptyDisplayNotFound(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		emptyDisplayResponse(w)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL, srv.Client())
	_, err := c.GetTopic(context.Background(), "RETAIL.ORDERS")
	if err == nil {
		t.Fatal("expected not found error for empty display response")
	}
	nf := &mqadmin.NotFoundError{}
	if !errors.As(err, &nf) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
	if nf.Object != "RETAIL.ORDERS" {
		t.Fatalf("NotFoundError.Object = %q, want RETAIL.ORDERS", nf.Object)
	}
}

// TestClient_GetChannelEmptyDisplayNotFound exercises the firstObjectAttributes
// NotFound branch in GetChannel via a successful, object-less DISPLAY.
func TestClient_GetChannelEmptyDisplayNotFound(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		emptyDisplayResponse(w)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL, srv.Client())
	spec := mqadmin.ChannelSpec{Name: "ORDERS.APP", Type: mqadmin.ChannelTypeSvrconn}
	_, err := c.GetChannel(context.Background(), spec)
	if err == nil {
		t.Fatal("expected not found error for empty display response")
	}
	nf := &mqadmin.NotFoundError{}
	if !errors.As(err, &nf) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
	if nf.Object != "ORDERS.APP" {
		t.Fatalf("NotFoundError.Object = %q, want ORDERS.APP", nf.Object)
	}
}

// TestClient_GetChannelUnsupportedType exercises the validateChannelType guard
// inside GetChannel: an unsupported channel type must short-circuit with a
// terminal error before any HTTP call is made.
func TestClient_GetChannelUnsupportedType(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Error("GetChannel must not call mqweb for an unsupported channel type")
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL, srv.Client())
	spec := mqadmin.ChannelSpec{Name: "ORDERS.APP", Type: mqadmin.ChannelType("BOGUS")}
	_, err := c.GetChannel(context.Background(), spec)
	if err == nil {
		t.Fatal("expected terminal error for unsupported channel type")
	}
	if !errors.Is(err, mqadmin.ErrTerminal) {
		t.Fatalf("expected terminal error, got %v", err)
	}
}

// TestClient_GetAliasQueueEmptyDisplayNotFound exercises the firstObjectAttributes
// NotFound branch in GetQueue. An alias spec avoids the local-queue display probe,
// so the object-less DISPLAY flows straight to firstObjectAttributes.
func TestClient_GetAliasQueueEmptyDisplayNotFound(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		emptyDisplayResponse(w)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL, srv.Client())
	spec := mqadmin.QueueSpec{Name: "APP.ALIAS", Type: mqadmin.QueueTypeAlias}
	_, err := c.GetQueue(context.Background(), spec)
	if err == nil {
		t.Fatal("expected not found error for empty display response")
	}
	nf := &mqadmin.NotFoundError{}
	if !errors.As(err, &nf) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
	if nf.Object != "APP.ALIAS" {
		t.Fatalf("NotFoundError.Object = %q, want APP.ALIAS", nf.Object)
	}
}

// TestClient_GetLocalQueueDisplayProbeError exercises the error branch in GetQueue
// where the local-queue display-parameter probe itself fails (mqweb 5xx). GetQueue
// must surface the transient error before attempting the real DISPLAY.
func TestClient_GetLocalQueueDisplayProbeError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL, srv.Client())
	spec := mqadmin.QueueSpec{Name: "APP.LOCAL", Type: mqadmin.QueueTypeLocal}
	_, err := c.GetQueue(context.Background(), spec)
	if err == nil {
		t.Fatal("expected error when the local display probe fails")
	}
	if !errors.Is(err, mqadmin.ErrTransient) {
		t.Fatalf("expected transient error, got %v", err)
	}
}

// TestClient_GetAuthorityDisplayError exercises the postMQSC non-NotFound error
// branch in runDisplayMQSC: mqweb returns HTTP 500 for the DISPLAY AUTHREC call,
// which must surface as a transient error (not swallowed as NotFound).
func TestClient_GetAuthorityDisplayError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL, srv.Client())
	spec := mqadmin.AuthoritySpec{
		Profile:    "APP.ORDERS",
		ObjectType: mqadmin.AuthorityObjectTypeQueue,
		Principal:  "app",
	}
	_, err := c.GetAuthority(context.Background(), spec)
	if err == nil {
		t.Fatal("expected error when DISPLAY AUTHREC returns HTTP 500")
	}
	if !errors.Is(err, mqadmin.ErrTransient) {
		t.Fatalf("expected transient error, got %v", err)
	}
}

// TestClient_GetAuthorityEmptyDisplayNotFound exercises the firstObjectAttributes
// NotFound branch inside runDisplayMQSC: a successful DISPLAY AUTHREC with no
// objects must be reported as not found against the requested profile.
func TestClient_GetAuthorityEmptyDisplayNotFound(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		emptyDisplayResponse(w)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL, srv.Client())
	spec := mqadmin.AuthoritySpec{
		Profile:    "MISSING.PROFILE",
		ObjectType: mqadmin.AuthorityObjectTypeQueue,
		Principal:  "app",
	}
	_, err := c.GetAuthority(context.Background(), spec)
	if err == nil {
		t.Fatal("expected not found error for empty DISPLAY AUTHREC response")
	}
	if !errors.Is(err, mqadmin.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

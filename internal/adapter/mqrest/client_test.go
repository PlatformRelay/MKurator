package mqrest_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/konradheimel/kurator/internal/adapter/mqrest"
	"github.com/konradheimel/kurator/internal/mqadmin"
)

const (
	testKeyCommandResponse       = "commandResponse"
	testKeyCompletionCode        = "completionCode"
	testKeyOverallCompletionCode = "overallCompletionCode"
)

func TestClient_PingSuccess(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ibmmq/rest/v3/admin/qmgr/QM1" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL, srv.Client())
	if err := c.Ping(context.Background()); err != nil {
		t.Fatalf("Ping: %v", err)
	}
}

func TestClient_DefineAndGetQueue(t *testing.T) {
	t.Parallel()
	var lastBody map[string]any
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&lastBody); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if lastBody["command"] == "display" {
			rp, _ := lastBody["responseParameters"].([]any)
			for _, p := range rp {
				if p == "maxmsglen" {
					t.Error("display must not request maxmsglen")
				}
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				testKeyCommandResponse: []map[string]any{{
					testKeyCompletionCode: 0,
					"parameters":          map[string]any{"maxdepth": "5000", "descr": "orders"},
				}},
				testKeyOverallCompletionCode: 0,
			})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			testKeyCommandResponse:       []map[string]any{{testKeyCompletionCode: 0}},
			testKeyOverallCompletionCode: 0,
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL, srv.Client())
	spec := mqadmin.QueueSpec{
		Name: "APP.ORDERS",
		Type: mqadmin.QueueTypeLocal,
		Attributes: map[string]string{
			"maxdepth": "5000",
			"descr":    "orders",
		},
	}
	if err := c.DefineQueue(context.Background(), spec); err != nil {
		t.Fatalf("DefineQueue: %v", err)
	}
	if lastBody["type"] != "runCommandJSON" {
		t.Fatalf("define type = %v", lastBody["type"])
	}
	params, _ := lastBody["parameters"].(map[string]any)
	if params["maxdepth"] != float64(5000) && params["maxdepth"] != 5000 {
		t.Fatalf("maxdepth param = %T(%v)", params["maxdepth"], params["maxdepth"])
	}
	state, err := c.GetQueue(context.Background(), "APP.ORDERS")
	if err != nil {
		t.Fatalf("GetQueue: %v", err)
	}
	if state.Attributes["maxdepth"] != "5000" {
		t.Fatalf("maxdepth = %q", state.Attributes["maxdepth"])
	}
}

func TestClient_GetQueueNotFound(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			testKeyCommandResponse: []map[string]any{{
				testKeyCompletionCode: 2,
				"message":             []string{"AMQ8147E: IBM MQ object APP.MISSING not found."},
			}},
			testKeyOverallCompletionCode: 2,
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL, srv.Client())
	_, err := c.GetQueue(context.Background(), "APP.MISSING")
	if err == nil {
		t.Fatal("expected not found error")
	}
	if !errors.Is(err, mqadmin.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestClient_RunMQSC(t *testing.T) {
	t.Parallel()
	var lastBody map[string]any
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&lastBody); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			testKeyOverallCompletionCode: 0,
			testKeyCommandResponse:       []map[string]any{{testKeyCompletionCode: 0}},
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL, srv.Client())
	cmd := "DEFINE CHANNEL('APP.CH') CHLTYPE(SVRCONN) REPLACE"
	if err := c.RunMQSC(context.Background(), cmd); err != nil {
		t.Fatalf("RunMQSC: %v", err)
	}
	if lastBody["type"] != "runCommand" {
		t.Fatalf("type = %v", lastBody["type"])
	}
	params, _ := lastBody["parameters"].(map[string]any)
	if params["command"] != cmd {
		t.Fatalf("command = %v", params["command"])
	}
}

func newTestClient(t *testing.T, endpoint string, hc *http.Client) *mqrest.Client {
	t.Helper()
	u, err := url.Parse(endpoint)
	if err != nil {
		t.Fatal(err)
	}
	c, err := mqrest.NewClient(mqrest.Config{
		Endpoint:     u,
		QueueManager: "QM1",
		Username:     "admin",
		Password:     "pass",
		HTTPClient:   hc,
	})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestNewClientValidation(t *testing.T) {
	t.Parallel()
	_, err := mqrest.NewClient(mqrest.Config{})
	if err == nil {
		t.Fatal("expected error for missing endpoint")
	}
	u, _ := url.Parse("https://mq.example:9443")
	_, err = mqrest.NewClient(mqrest.Config{Endpoint: u})
	if err == nil {
		t.Fatal("expected error for missing queue manager")
	}
}

func TestClient_PingUnauthorized(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()
	c := newTestClient(t, srv.URL, srv.Client())
	err := c.Ping(context.Background())
	if !errors.Is(err, mqadmin.ErrTerminal) {
		t.Fatalf("expected terminal error, got %v", err)
	}
}

func TestClient_DeleteQueue(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			testKeyOverallCompletionCode: 0,
			testKeyCommandResponse:       []map[string]any{{testKeyCompletionCode: 0}},
		})
	}))
	defer srv.Close()
	c := newTestClient(t, srv.URL, srv.Client())
	if err := c.DeleteQueue(context.Background(), "APP.ORDERS"); err != nil {
		t.Fatalf("DeleteQueue: %v", err)
	}
}

func TestClient_DeleteQueueNotFound(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			testKeyCommandResponse: []map[string]any{{
				testKeyCompletionCode: 2,
				"message":             []string{"AMQ8147E: IBM MQ object APP.GONE not found."},
			}},
			testKeyOverallCompletionCode: 2,
		})
	}))
	defer srv.Close()
	c := newTestClient(t, srv.URL, srv.Client())
	if err := c.DeleteQueue(context.Background(), "APP.GONE"); err != nil {
		t.Fatalf("DeleteQueue not found should succeed: %v", err)
	}
}

func TestClient_DefineQueueUnsupportedType(t *testing.T) {
	t.Parallel()
	u, _ := url.Parse("https://mq.example:9443")
	c, err := mqrest.NewClient(mqrest.Config{
		Endpoint: u, QueueManager: "QM1", Username: "a", Password: "b",
	})
	if err != nil {
		t.Fatal(err)
	}
	err = c.DefineQueue(context.Background(), mqadmin.QueueSpec{
		Name: "X", Type: mqadmin.QueueType("remote"),
	})
	if !errors.Is(err, mqadmin.ErrTerminal) {
		t.Fatalf("expected terminal error, got %v", err)
	}
}

func TestClient_PostMQSCServerError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()
	c := newTestClient(t, srv.URL, srv.Client())
	err := c.RunMQSC(context.Background(), "DISPLAY QMGR")
	if !errors.Is(err, mqadmin.ErrTransient) {
		t.Fatalf("expected transient error, got %v", err)
	}
}

func TestClient_PostMQSCBadRequestLongBody(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(strings.Repeat("x", 300)))
	}))
	defer srv.Close()
	c := newTestClient(t, srv.URL, srv.Client())
	err := c.RunMQSC(context.Background(), "DISPLAY QMGR")
	if !errors.Is(err, mqadmin.ErrTerminal) {
		t.Fatalf("expected terminal error, got %v", err)
	}
}

func TestClient_PostMQSCInvalidJSON(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not-json"))
	}))
	defer srv.Close()
	c := newTestClient(t, srv.URL, srv.Client())
	err := c.RunMQSC(context.Background(), "DISPLAY QMGR")
	if !errors.Is(err, mqadmin.ErrTerminal) {
		t.Fatalf("expected terminal error, got %v", err)
	}
}

func TestClient_PingServerError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()
	c := newTestClient(t, srv.URL, srv.Client())
	err := c.Ping(context.Background())
	if !errors.Is(err, mqadmin.ErrTransient) {
		t.Fatalf("expected transient error, got %v", err)
	}
}

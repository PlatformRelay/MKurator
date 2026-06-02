package mqrest_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
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

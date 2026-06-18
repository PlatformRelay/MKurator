package mqrest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/conduit-ops/mkurator/internal/mqadmin"
)

func TestResolveQueueDriftCheckKeys_LocalIncludesShareWhenProbeSucceeds(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if body["command"] != "display" || body["name"] != defaultQueueLocalDisplayProbeObject {
			t.Fatalf("unexpected request: %+v", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"commandResponse": []map[string]any{{
				"completionCode": 0,
				"parameters":     map[string]any{"share": "yes"},
			}},
			"overallCompletionCode": 0,
		})
	}))
	defer srv.Close()

	c := newProbeTestClient(t, srv.URL, srv.Client())
	keys, err := ResolveQueueDriftCheckKeys(context.Background(), c, mqadmin.QueueTypeLocal)
	if err != nil {
		t.Fatalf("ResolveQueueDriftCheckKeys: %v", err)
	}
	if !containsString(keys, attrShare) {
		t.Fatalf("expected share in drift keys, got %v", keys)
	}
}

func TestResolveQueueDriftCheckKeys_LocalOmitsShareWhenProbeRejected(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":[{"msgId":"MQWB0120E","message":"Attribute not valid"}]}`))
	}))
	defer srv.Close()

	c := newProbeTestClient(t, srv.URL, srv.Client())
	keys, err := ResolveQueueDriftCheckKeys(context.Background(), c, mqadmin.QueueTypeLocal)
	if err != nil {
		t.Fatalf("ResolveQueueDriftCheckKeys: %v", err)
	}
	if containsString(keys, attrShare) {
		t.Fatalf("expected share omitted from drift keys, got %v", keys)
	}
}

func TestResolveQueueDriftCheckKeys_AliasStatic(t *testing.T) {
	t.Parallel()
	c := newProbeTestClient(t, "https://example.invalid", http.DefaultClient)
	keys, err := ResolveQueueDriftCheckKeys(context.Background(), c, mqadmin.QueueTypeAlias)
	if err != nil {
		t.Fatalf("ResolveQueueDriftCheckKeys: %v", err)
	}
	if len(keys) == 0 || keys[0] != attrTargq {
		t.Fatalf("alias keys = %v", keys)
	}
}

func TestGetQueue_LocalRequestsShareWhenProbeAllows(t *testing.T) {
	t.Parallel()
	var displayParams []any
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if body["command"] != "display" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"commandResponse":       []map[string]any{{"completionCode": 0}},
				"overallCompletionCode": 0,
			})
			return
		}
		if body["name"] == defaultQueueLocalDisplayProbeObject {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"commandResponse": []map[string]any{{
					"completionCode": 0,
					"parameters":     map[string]any{"share": "yes"},
				}},
				"overallCompletionCode": 0,
			})
			return
		}
		displayParams, _ = body["responseParameters"].([]any)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"commandResponse": []map[string]any{{
				"completionCode": 0,
				"parameters":     map[string]any{"maxdepth": "100", "share": "yes"},
			}},
			"overallCompletionCode": 0,
		})
	}))
	defer srv.Close()

	c := newProbeTestClient(t, srv.URL, srv.Client())
	state, err := c.GetQueue(context.Background(), mqadmin.QueueSpec{
		Name: "APP.Q",
		Type: mqadmin.QueueTypeLocal,
	})
	if err != nil {
		t.Fatalf("GetQueue: %v", err)
	}
	if !containsAny(displayParams, attrShare) {
		t.Fatalf("display responseParameters = %v, want share", displayParams)
	}
	if state.Attributes["share"] != "yes" {
		t.Fatalf("share = %q", state.Attributes["share"])
	}
}

func newProbeTestClient(t *testing.T, baseURL string, hc *http.Client) *Client {
	t.Helper()
	u, err := url.Parse(baseURL)
	if err != nil {
		t.Fatal(err)
	}
	c, err := NewClient(Config{
		Endpoint:     u,
		QueueManager: "QM1",
		HTTPClient:   hc,
	})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func containsString(slice []string, want string) bool {
	for _, s := range slice {
		if s == want {
			return true
		}
	}
	return false
}

func containsAny(slice []any, want string) bool {
	for _, s := range slice {
		if s == want {
			return true
		}
	}
	return false
}

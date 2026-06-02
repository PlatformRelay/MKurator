package mqrest

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/konradheimel/kurator/internal/metrics"
	"github.com/konradheimel/kurator/internal/mqadmin"
)

const (
	// DefaultRESTPrefix is the mqweb REST API path for IBM MQ 9.3+.
	DefaultRESTPrefix = "/ibmmq/rest/v3"
	csrfHeader        = "ibm-mq-rest-csrf-token"
	mqscType          = "runCommandJSON"
	qualifierQLocal   = "qlocal"
)

// Config holds connection parameters for mqweb.
type Config struct {
	Endpoint     *url.URL
	RESTPrefix   string
	QueueManager string
	Username     string
	Password     string
	TLSConfig    *tls.Config
	HTTPClient   *http.Client
}

// Client implements mqadmin.Admin over the mqweb /mqsc endpoint.
type Client struct {
	mqscURL      string
	adminQMURL   string
	queueManager string
	httpClient   *http.Client
	username     string
	password     string
}

// NewClient builds an mqrest client from Config.
func NewClient(cfg Config) (*Client, error) {
	if cfg.Endpoint == nil {
		return nil, fmt.Errorf("endpoint is required")
	}
	if cfg.QueueManager == "" {
		return nil, fmt.Errorf("queue manager name is required")
	}
	prefix := cfg.RESTPrefix
	if prefix == "" {
		prefix = DefaultRESTPrefix
	}

	base := strings.TrimSuffix(cfg.Endpoint.String(), "/") + prefix
	qm := url.PathEscape(cfg.QueueManager)
	mqscURL := fmt.Sprintf("%s/admin/action/qmgr/%s/mqsc", base, qm)
	adminQMURL := fmt.Sprintf("%s/admin/qmgr/%s", base, qm)

	hc := cfg.HTTPClient
	if hc == nil {
		tr := http.DefaultTransport.(*http.Transport).Clone()
		if cfg.TLSConfig != nil {
			tr.TLSClientConfig = cfg.TLSConfig.Clone()
		}
		hc = &http.Client{Timeout: 60 * time.Second, Transport: tr}
	}

	return &Client{
		mqscURL:      mqscURL,
		adminQMURL:   adminQMURL,
		queueManager: cfg.QueueManager,
		httpClient:   hc,
		username:     cfg.Username,
		password:     cfg.Password,
	}, nil
}

// Ping verifies mqweb can reach the queue manager.
func (c *Client) Ping(ctx context.Context) error {
	var err error
	defer func() { metrics.RecordMQOperation(metrics.MQOpPing, err) }()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.adminQMURL, nil)
	if err != nil {
		return fmt.Errorf("build ping request: %w", err)
	}
	req.SetBasicAuth(c.username, c.password)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return &mqadmin.TransientError{Message: "mqweb ping failed", Cause: err}
	}
	defer closeBody(res.Body)

	if res.StatusCode == http.StatusUnauthorized || res.StatusCode == http.StatusForbidden {
		return &mqadmin.TerminalError{
			Reason:  "Unauthorized",
			Message: fmt.Sprintf("mqweb ping returned HTTP %d", res.StatusCode),
		}
	}
	if res.StatusCode >= 500 {
		return &mqadmin.TransientError{Message: fmt.Sprintf("mqweb ping returned HTTP %d", res.StatusCode)}
	}
	if res.StatusCode >= 400 {
		return &mqadmin.TerminalError{
			Reason:  "Unreachable",
			Message: fmt.Sprintf("mqweb ping returned HTTP %d", res.StatusCode),
		}
	}
	return nil
}

// GetQueue returns observed attributes for a local queue.
func (c *Client) GetQueue(ctx context.Context, name string) (*mqadmin.QueueState, error) {
	var err error
	defer func() { metrics.RecordMQOperation(metrics.MQOpGetQueue, err) }()

	resp, err := c.runCommandJSON(ctx, runCommandJSONRequest{
		Type:               mqscType,
		Command:            "display",
		Qualifier:          qualifierQLocal,
		Name:               name,
		ResponseParameters: append([]string(nil), queueDisplayParameters...),
	})
	if err != nil {
		return nil, err
	}
	if len(resp.CommandResponse) == 0 {
		err = &mqadmin.NotFoundError{Object: name}
		return nil, err
	}
	attrs := map[string]string{}
	for _, cr := range resp.CommandResponse {
		for k, v := range cr.Parameters {
			attrs[strings.ToLower(k)] = fmt.Sprint(v)
		}
	}
	if len(attrs) == 0 && resp.overallFailed() {
		if resp.isObjectMissing() {
			err = &mqadmin.NotFoundError{Object: name}
			return nil, err
		}
		err = resp.terminalError("display queue")
		return nil, err
	}
	return &mqadmin.QueueState{Name: name, Attributes: attrs}, nil
}

// DefineQueue creates or updates a local queue (REPLACE).
func (c *Client) DefineQueue(ctx context.Context, spec mqadmin.QueueSpec) error {
	var err error
	defer func() { metrics.RecordMQOperation(metrics.MQOpDefineQueue, err) }()

	if spec.Type != "" && spec.Type != mqadmin.QueueTypeLocal {
		err = &mqadmin.TerminalError{
			Reason:  "UnsupportedQueueType",
			Message: fmt.Sprintf("queue type %q is not supported yet", spec.Type),
		}
		return err
	}
	params := defineQueueParameters(spec)
	_, err = c.runCommandJSON(ctx, runCommandJSONRequest{
		Type:       mqscType,
		Command:    "define",
		Qualifier:  qualifierQLocal,
		Name:       spec.Name,
		Parameters: params,
	})
	return err
}

// RunMQSC executes a single MQSC command string via the runCommand API.
func (c *Client) RunMQSC(ctx context.Context, command string) error {
	var err error
	defer func() { metrics.RecordMQOperation(metrics.MQOpRunMQSC, err) }()

	body := runCommandRequest{Type: "runCommand"}
	body.Parameters.Command = command
	_, err = c.postMQSC(ctx, body)
	return err
}

// DeleteQueue removes a local queue.
func (c *Client) DeleteQueue(ctx context.Context, name string) error {
	var err error
	defer func() { metrics.RecordMQOperation(metrics.MQOpDeleteQueue, err) }()

	_, err = c.runCommandJSON(ctx, runCommandJSONRequest{
		Type:      mqscType,
		Command:   "delete",
		Qualifier: qualifierQLocal,
		Name:      name,
	})
	if err != nil && errors.Is(err, mqadmin.ErrNotFound) {
		err = nil
		return nil
	}
	return err
}

func (c *Client) runCommandJSON(ctx context.Context, body runCommandJSONRequest) (*mqscResponse, error) {
	return c.postMQSC(ctx, body)
}

func (c *Client) postMQSC(ctx context.Context, body any) (*mqscResponse, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal mqsc request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.mqscURL, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("build mqsc request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set(csrfHeader, "1")
	req.SetBasicAuth(c.username, c.password)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, &mqadmin.TransientError{Message: "mqweb request failed", Cause: err}
	}
	defer closeBody(res.Body)

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, &mqadmin.TransientError{Message: "read mqweb response", Cause: err}
	}

	if res.StatusCode == http.StatusUnauthorized || res.StatusCode == http.StatusForbidden {
		return nil, &mqadmin.TerminalError{
			Reason:  "Unauthorized",
			Message: fmt.Sprintf("mqweb returned HTTP %d", res.StatusCode),
		}
	}
	if res.StatusCode >= 500 || res.StatusCode == http.StatusServiceUnavailable {
		return nil, &mqadmin.TransientError{
			Message: fmt.Sprintf("mqweb returned HTTP %d", res.StatusCode),
		}
	}
	if res.StatusCode >= 400 {
		return nil, &mqadmin.TerminalError{
			Reason:  "BadRequest",
			Message: fmt.Sprintf("mqweb returned HTTP %d: %s", res.StatusCode, truncate(string(raw), 200)),
		}
	}

	var parsed mqscResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, &mqadmin.TerminalError{
			Reason:  "InvalidResponse",
			Message: "failed to parse mqsc response",
			Cause:   err,
		}
	}
	if parsed.overallFailed() {
		if parsed.isObjectMissing() {
			obj := ""
			if req, ok := body.(runCommandJSONRequest); ok {
				obj = req.Name
			}
			return &parsed, &mqadmin.NotFoundError{Object: obj}
		}
		return &parsed, parsed.terminalError("mqsc command failed")
	}
	return &parsed, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func closeBody(body io.ReadCloser) {
	_ = body.Close()
}

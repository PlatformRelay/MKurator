package mqrest

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/conduit-ops/mkurator/internal/mqadmin"
)

// mqwbAttributeNotDisplayable is returned by mqweb when a keyword is valid on
// DEFINE but not allowed in runCommandJSON responseParameters for DISPLAY.
const mqwbAttributeNotDisplayable = "MQWB0120"

// QueueLocalDefineOnlyCandidates lists QLOCAL attributes known to be DEFINE-only
// on IBM MQ 9.4.x mqweb (DISPLAY via responseParameters returns MQWB0120E).
// Used by capability-probe spikes and future QMC-ready probing (ADR-0024 §4).
var QueueLocalDefineOnlyCandidates = []string{
	attrShare, attrDefopts, attrBothresh, attrBoqname, attrUsage, attrMaxMsgLen,
}

func queueAttributeDisplayProbeRequest(queueName, attribute string) runCommandJSONRequest {
	return runCommandJSONRequest{
		Type:               mqscType,
		Command:            mqscCommandDisplay,
		Qualifier:          qualifierQLocal,
		Name:               queueName,
		ResponseParameters: []string{strings.ToLower(attribute)},
	}
}

// responseIndicatesAttributeNotDisplayable reports whether mqweb rejected a
// single-attribute DISPLAY with MQWB0120E (attribute not displayable).
func responseIndicatesAttributeNotDisplayable(resp *mqscResponse) bool {
	if resp == nil {
		return false
	}
	for _, cr := range resp.CommandResponse {
		for _, msg := range append(cr.Message, cr.Text...) {
			if strings.Contains(strings.ToUpper(msg), mqwbAttributeNotDisplayable) {
				return true
			}
		}
	}
	for _, e := range resp.Error {
		for _, msg := range []string{e.Message, e.Explanation} {
			if strings.Contains(strings.ToUpper(msg), mqwbAttributeNotDisplayable) {
				return true
			}
		}
	}
	return false
}

func errIndicatesAttributeNotDisplayable(err error) bool {
	if err == nil {
		return false
	}
	var term *mqadmin.TerminalError
	if errors.As(err, &term) {
		combined := strings.ToUpper(term.Message + " " + term.Reason)
		if strings.Contains(combined, mqwbAttributeNotDisplayable) {
			return true
		}
	}
	return strings.Contains(strings.ToUpper(err.Error()), mqwbAttributeNotDisplayable)
}

// ProbeQueueLocalAttributeDisplayable issues DISPLAY QLOCAL with a single
// responseParameter and reports whether mqweb returns the attribute.
//
// Requires an existing local queue (probe object). Returns (true, nil) when
// DISPLAY succeeds, (false, nil) when mqweb responds with MQWB0120E, or an
// error for missing queues and other MQ failures. Used by ADR-0024 §4 display
// capability cache when building local-queue DISPLAY for drift.
func (c *Client) ProbeQueueLocalAttributeDisplayable(
	ctx context.Context,
	queueName, attribute string,
) (bool, error) {
	queueName = strings.TrimSpace(queueName)
	attribute = strings.ToLower(strings.TrimSpace(attribute))
	if queueName == "" {
		return false, &mqadmin.TerminalError{Reason: "InvalidArgument", Message: "queue name is required"}
	}
	if attribute == "" {
		return false, &mqadmin.TerminalError{Reason: "InvalidArgument", Message: "attribute is required"}
	}

	resp, err := c.runCommandJSON(ctx, queueAttributeDisplayProbeRequest(queueName, attribute))
	if err == nil {
		return true, nil
	}
	if responseIndicatesAttributeNotDisplayable(resp) || errIndicatesAttributeNotDisplayable(err) {
		return false, nil
	}
	if resp != nil && resp.isObjectMissing() {
		return false, &mqadmin.NotFoundError{Object: queueName}
	}
	return false, fmt.Errorf("probe display qlocal %q attribute %q: %w", queueName, attribute, err)
}

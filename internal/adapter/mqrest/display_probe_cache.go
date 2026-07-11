package mqrest

import (
	"context"
	"fmt"

	"github.com/platformrelay/mkurator/internal/mqadmin"
)

// defaultQueueLocalDisplayProbeObject is a stable QLOCAL used to probe DISPLAY
// capability without depending on the workload queue existing yet.
const defaultQueueLocalDisplayProbeObject = "SYSTEM.DEFAULT.LOCAL.QUEUE"

// queueLocalProbedDisplayCandidates lists define-only candidates probed at runtime
// when building local-queue DISPLAY for drift (ADR-0024 §4).
var queueLocalProbedDisplayCandidates = QueueLocalDefineOnlyCandidates

// ResolveQueueDriftCheckKeys returns DISPLAY-safe drift keys for the queue type.
// For local queues backed by *Client, probed define-only candidates (e.g. share)
// are included when mqweb reports them displayable; other Admin implementations
// receive the static base list only.
func ResolveQueueDriftCheckKeys(
	ctx context.Context,
	admin mqadmin.Admin,
	qType mqadmin.QueueType,
) ([]string, error) {
	if mqadmin.NormalizeQueueType(qType) != mqadmin.QueueTypeLocal {
		return queueDisplayParameters(qType), nil
	}
	c, ok := admin.(*Client)
	if !ok {
		return queueDisplayParameters(qType), nil
	}
	return c.queueLocalDisplayParametersResolved(ctx)
}

func (c *Client) queueLocalDisplayParametersResolved(ctx context.Context) ([]string, error) {
	base := append([]string(nil), queueLocalDisplayParameters...)
	for _, attr := range queueLocalProbedDisplayCandidates {
		displayable, err := c.probeQueueLocalDisplayableCached(ctx, attr)
		if err != nil {
			return nil, fmt.Errorf("probe qlocal display %q: %w", attr, err)
		}
		if displayable {
			base = append(base, attr)
		}
	}
	return base, nil
}

func (c *Client) probeQueueLocalDisplayableCached(ctx context.Context, attribute string) (bool, error) {
	c.displayProbeMu.Lock()
	if c.displayProbeCache != nil {
		if displayable, ok := c.displayProbeCache[attribute]; ok {
			c.displayProbeMu.Unlock()
			return displayable, nil
		}
	}
	c.displayProbeMu.Unlock()

	displayable, err := c.ProbeQueueLocalAttributeDisplayable(
		ctx,
		defaultQueueLocalDisplayProbeObject,
		attribute,
	)
	if err != nil {
		return false, err
	}

	c.displayProbeMu.Lock()
	if c.displayProbeCache == nil {
		c.displayProbeCache = make(map[string]bool)
	}
	c.displayProbeCache[attribute] = displayable
	c.displayProbeMu.Unlock()
	return displayable, nil
}

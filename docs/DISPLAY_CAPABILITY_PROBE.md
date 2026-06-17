# DISPLAY capability probing (spike)

Spike for [ADR-0024 §4](adr/0024-mqsc-command-construction-hygiene.md): probe whether
mqweb allows a queue attribute in `runCommandJSON` DISPLAY `responseParameters`
instead of maintaining static safe lists in `internal/adapter/mqrest/mqsc_params.go`.

## Problem

Some QLOCAL attributes are valid on **DEFINE** but rejected when requested in
DISPLAY `responseParameters`. On IBM MQ 9.4.x mqweb this surfaces as
`MQWB0120E`. MKurator omits those keys from drift checks today; see
[ATTRIBUTE_RECONCILIATION.md](ATTRIBUTE_RECONCILIATION.md) (e.g. `share`,
`defopts`, `bothresh`, `boqname`, `usage`, `maxmsglen`).

Hand-maintained slices (`queueLocalDisplayParameters`) do not adapt when a newer
mqweb starts supporting DISPLAY for a formerly define-only keyword.

## Probe method

Issue DISPLAY for an **existing** local queue with a single `responseParameter`:

```json
{
  "type": "runCommandJSON",
  "command": "display",
  "qualifier": "qlocal",
  "name": "<probe-queue>",
  "responseParameters": ["share"]
}
```

Interpretation:

| Outcome | Meaning |
|---------|---------|
| `overallCompletionCode` 0 and parameters returned | Attribute is **displayable** — safe to add to drift checks |
| Message contains `MQWB0120E` | Attribute is **define-only** on this mqweb/QM |
| `AMQ8147E` / not found | Probe queue missing — fix probe setup, not attribute capability |

Implementation: `Client.ProbeQueueLocalAttributeDisplayable` in
`internal/adapter/mqrest/display_probe.go`.

## Spike result: `share` on mqweb 9.4

Pilot attribute: **`share`** (representative of the DEFINE-only group in
ATTRIBUTE_RECONCILIATION).

| Environment | DISPLAY `share` | Notes |
|-------------|-----------------|-------|
| IBM MQ 9.4.x mqweb (Docker integration) | **Not displayable** (`MQWB0120E`) | Confirms static omission in `queueLocalDisplayParameters` |
| Unit httptest | Simulated `MQWB0120E` | `TestClient_ProbeQueueLocalAttributeDisplayable` |

`define share` on QLOCAL still succeeds; only DISPLAY via `responseParameters`
is blocked. Drift for `share` remains deferred until a probe (or manual test)
shows DISPLAY support on the target mqweb level.

## Future wiring (not in this spike)

Per ADR-0024 §4, a full implementation would:

1. Run probes once per `QueueManagerConnection` at Ready (using a stable probe
   queue, e.g. `SYSTEM.DEFAULT.LOCAL.QUEUE` or a dedicated operator probe object).
2. Cache displayable attribute sets on QMC status (or in-memory on the adapter
   factory).
3. Build DISPLAY `responseParameters` dynamically from desired keys ∩ displayable
   set instead of `queueLocalDisplayParameters`.

Candidates for bulk probe: `QueueLocalDefineOnlyCandidates` in `display_probe.go`.

## Verification

```bash
# Unit (always)
go test ./internal/adapter/mqrest/... -run Probe -count=1

# Live mqweb (optional)
KURATOR_INTEGRATION_MQ=1 task mq:integration:up
go test -tags=integration ./test/integration/mq/... -run ProbeQueueLocalAttribute_share -count=1
```

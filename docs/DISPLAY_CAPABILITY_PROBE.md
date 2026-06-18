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

## Spike result: `share` and `maxmsglen` on mqweb 9.4.x

Pilot attributes:

| Attribute | Static table (9.4.x) | Probe outcome |
|-----------|---------------------|---------------|
| **`maxmsglen`** | DEFINE-only (`MQWB0120E`) | **Not displayable** on Docker MQ `9.4.5.1-r1` (integration CI) — confirms static omission |
| **`share`** | DEFINE-only (`MQWB0120E`) | **Displayable** on Docker MQ `9.4.5.1-r1` — static safe list is stale for this fix level |

The `share` result is why ADR-0024 §4 favours runtime probing over hand-maintained
`queueLocalDisplayParameters`: mqweb fix packs can enable DISPLAY for keywords
that older docs and tables still list as define-only.

`define share` / `define maxmsglen` on QLOCAL still succeed when DISPLAY via
`responseParameters` is blocked. Drift for blocked keys remains deferred until
probe (or manual test) shows DISPLAY support on the target mqweb level.

## Future wiring

Per ADR-0024 §4, remaining work:

1. Probe additional candidates from `QueueLocalDefineOnlyCandidates` (`defopts`,
   `bothresh`, `boqname`, `usage`, `maxmsglen`) using the same client cache.
2. Optionally run probes once per `QueueManagerConnection` at Ready and surface
   displayable sets on QMC status (today: lazy probe on first local `GetQueue`).
3. Build DISPLAY `responseParameters` from desired keys ∩ displayable set.

Candidates for bulk probe: `QueueLocalDefineOnlyCandidates` in `display_probe.go`.

## Verification

```bash
# Unit (always)
go test ./internal/adapter/mqrest/... -run Probe -count=1

# Live mqweb (optional)
KURATOR_INTEGRATION_MQ=1 task mq:integration:up
go test -tags=integration ./test/integration/mq/... -run ProbeQueueLocalAttribute_DisplayProbeMechanism -count=1
```

# Sample resources (Helm)

Same manifests as [`config/samples/`](../../../config/samples/README.md), synced via
`task samples:sync`. Applied by **`task deploy:samples`** after `task deploy:helm`
(or as part of `task local:up`).

See [docs/INSTALL_AND_USE.md](../../../docs/INSTALL_AND_USE.md) for the full install
and usage guide.

| File | Kind |
|------|------|
| `mq-credentials-secret.yaml` | `Secret` (chart-only; not synced from config) |
| `queuemanagerconnection.yaml` | `QueueManagerConnection` |
| `queue.yaml` | `Queue` (local) |
| `queue-alias.yaml` | `Queue` (alias) |
| `queue-remote.yaml` | `Queue` (remote) |
| `topic.yaml` | `Topic` |
| `channel.yaml` | `Channel` |
| `channelauthrule.yaml` | `ChannelAuthRule` (`ADDRESSMAP`) |
| `channelauthrule-blockuser.yaml` | `ChannelAuthRule` (`BLOCKUSER`, optional) |
| `authorityrecord.yaml` | `AuthorityRecord` |

**Preferred apply:** `task deploy:samples` (creates `kurator-system` if needed, then
`kubectl apply --server-side -k` this directory).

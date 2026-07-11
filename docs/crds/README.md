# Custom resources

API group: **`messaging.mkurator.dev`** — both **`v1alpha1`** and **`v1beta1`**
are served (all six kinds), with a conversion webhook between them. New manifests
should use **`v1beta1`** (the samples below default to it); `v1alpha1` remains the
etcd storage version until the hub migration is proven in CI
([ADR-0026](../adr/0026-v1beta1-graduation-plan.md), [API_STABILITY.md](../API_STABILITY.md)).

Field-level reference pages (every `spec`/`status` field) are being added per
kind; the field tables are generated from the CRD OpenAPI schema by
`task docs:crd-ref` and augmented with hand-written prose. All six kinds are
covered: **[QueueManagerConnection](queuemanagerconnection.md)**,
**[Queue](queue.md)**, **[Topic](topic.md)**, **[Channel](channel.md)**,
**[ChannelAuthRule](channelauthrule.md)**, **[AuthorityRecord](authorityrecord.md)**.

| Kind | Short name | Purpose | Sample YAML |
| --- | --- | --- | --- |
| [`QueueManagerConnection`](queuemanagerconnection.md) | `qmc` | mqweb endpoint, TLS, credential `Secret` reference | [queuemanagerconnection.yaml](https://github.com/platformrelay/MKurator/blob/main/config/samples/messaging_v1beta1_queuemanagerconnection.yaml) |
| [`Queue`](queue.md) | `mq` | Local, alias, or remote queue | [queue.yaml](https://github.com/platformrelay/MKurator/blob/main/config/samples/messaging_v1beta1_queue.yaml) |
| [`Topic`](topic.md) | `tp` | Administrative topic | [topic.yaml](https://github.com/platformrelay/MKurator/blob/main/config/samples/messaging_v1beta1_topic.yaml) |
| [`Channel`](channel.md) | `chl` | Server-connection, sender, or receiver channel | [channel.yaml](https://github.com/platformrelay/MKurator/blob/main/config/samples/messaging_v1beta1_channel.yaml) |
| [`ChannelAuthRule`](channelauthrule.md) | `car` | Channel authentication rule | [channelauthrule.yaml](https://github.com/platformrelay/MKurator/blob/main/config/samples/messaging_v1beta1_channelauthrule.yaml) |
| [`AuthorityRecord`](authorityrecord.md) | `auth` | OAM authority record | [authorityrecord.yaml](https://github.com/platformrelay/MKurator/blob/main/config/samples/messaging_v1beta1_authorityrecord.yaml) |

## Queue variants

| `spec.type` | MQ object | Sample |
| --- | --- | --- |
| `local` (default) | `QLOCAL` | [queue.yaml](https://github.com/platformrelay/MKurator/blob/main/config/samples/messaging_v1beta1_queue.yaml) |
| `alias` | `QALIAS` | [queue_alias.yaml](https://github.com/platformrelay/MKurator/blob/main/config/samples/messaging_v1beta1_queue_alias.yaml) |
| `remote` | `QREMOTE` | [queue_remote.yaml](https://github.com/platformrelay/MKurator/blob/main/config/samples/messaging_v1beta1_queue_remote.yaml) |

## Further reading

- Field-level install guide: [INSTALL_AND_USE.md — Resource reference](../INSTALL_AND_USE.md#resource-reference)
- Drift and attribute policy: [ATTRIBUTE_RECONCILIATION.md](../ATTRIBUTE_RECONCILIATION.md)
- Annotated samples: [config/samples/README.md](https://github.com/platformrelay/MKurator/blob/main/config/samples/README.md)
- OpenAPI golden contracts: [test/schema/README.md](https://github.com/platformrelay/MKurator/blob/main/test/schema/README.md)

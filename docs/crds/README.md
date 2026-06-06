# Custom resources (v1alpha1)

API group: **`messaging.mkurator.dev/v1alpha1`**

| Kind | Short name | Purpose | Sample YAML |
| --- | --- | --- | --- |
| `QueueManagerConnection` | `qmc` | mqweb endpoint, TLS, credential `Secret` reference | [queuemanagerconnection.yaml](../../config/samples/messaging_v1alpha1_queuemanagerconnection.yaml) |
| `Queue` | `mq` | Local, alias, or remote queue | [queue.yaml](../../config/samples/messaging_v1alpha1_queue.yaml) |
| `Topic` | `tp` | Administrative topic | [topic.yaml](../../config/samples/messaging_v1alpha1_topic.yaml) |
| `Channel` | `chl` | Server-connection channel | [channel.yaml](../../config/samples/messaging_v1alpha1_channel.yaml) |
| `ChannelAuthRule` | `car` | Channel authentication rule | [channelauthrule.yaml](../../config/samples/messaging_v1alpha1_channelauthrule.yaml) |
| `AuthorityRecord` | `auth` | OAM authority record | [authorityrecord.yaml](../../config/samples/messaging_v1alpha1_authorityrecord.yaml) |

## Queue variants

| `spec.type` | MQ object | Sample |
| --- | --- | --- |
| `local` (default) | `QLOCAL` | [queue.yaml](../../config/samples/messaging_v1alpha1_queue.yaml) |
| `alias` | `QALIAS` | [queue_alias.yaml](../../config/samples/messaging_v1alpha1_queue_alias.yaml) |
| `remote` | `QREMOTE` | [queue_remote.yaml](../../config/samples/messaging_v1alpha1_queue_remote.yaml) |

## Further reading

- Field-level install guide: [INSTALL_AND_USE.md — Resource reference](../INSTALL_AND_USE.md#resource-reference)
- Drift and attribute policy: [ATTRIBUTE_RECONCILIATION.md](../ATTRIBUTE_RECONCILIATION.md)
- Annotated samples: [config/samples/README.md](../../config/samples/README.md)
- OpenAPI golden contracts: [test/schema/README.md](../../test/schema/README.md)

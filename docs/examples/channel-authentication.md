# Channel authentication walkthrough

Configure channel authentication (`CHLAUTH`) and OAM authority records with MKurator.

## Prerequisites

- A `QueueManagerConnection` in `Ready` state (see [queue-and-connection.md](queue-and-connection.md))
- A `Channel` CR for the SVRCONN channel you want to protect

## ChannelAuthRule

`ChannelAuthRule` maps one-to-one to a `SET CHLAUTH` rule. Samples:

| Rule type | Sample |
| --- | --- |
| `ADDRESSMAP` | [channelauthrule.yaml](https://github.com/conduit-ops/MKurator/blob/main/config/samples/messaging_v1beta1_channelauthrule.yaml) |
| `BLOCKUSER` | [channelauthrule_blockuser.yaml](https://github.com/conduit-ops/MKurator/blob/main/config/samples/messaging_v1beta1_channelauthrule_blockuser.yaml) |
| `BLOCKADDR` | [channelauthrule_blockaddr.yaml](https://github.com/conduit-ops/MKurator/blob/main/config/samples/messaging_v1beta1_channelauthrule_blockaddr.yaml) |

Other `ruleType` values are accepted by the API and validated at MQ apply time. See
[PHASE5_AUTH_SKETCH.md](../PHASE5_AUTH_SKETCH.md) for the roadmap.

```bash
kubectl apply -f config/samples/messaging_v1alpha1_channelauthrule.yaml
kubectl describe channelauthrule <name>
```

## AuthorityRecord

Grant OAM authorities on a queue or channel profile:

[`config/samples/messaging_v1alpha1_authorityrecord.yaml`](https://github.com/conduit-ops/MKurator/blob/main/config/samples/messaging_v1beta1_authorityrecord.yaml)

Auth objects use GET/replace reconciliation (not the DISPLAY drift matrix used for queues).
See [ATTRIBUTE_RECONCILIATION.md](../ATTRIBUTE_RECONCILIATION.md).

## Verify

Integration and kind e2e cover `ADDRESSMAP` and `BLOCKUSER` paths — see
[README.md#what-ci-proves](https://github.com/conduit-ops/MKurator/blob/main/README.md#what-ci-proves).

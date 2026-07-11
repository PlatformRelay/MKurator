# Queue and connection walkthrough

Connect MKurator to an existing queue manager and create a local queue.

## 1. Credentials Secret

Create a `Secret` with username and password keys referenced by
`QueueManagerConnection.spec.credentialsSecretRef`. See the annotated sample:

[`config/samples/messaging_v1alpha1_queuemanagerconnection.yaml`](https://github.com/platformrelay/MKurator/blob/main/config/samples/messaging_v1beta1_queuemanagerconnection.yaml)

## 2. QueueManagerConnection

Apply a `QueueManagerConnection` with:

- `spec.endpoint` — mqweb HTTPS URL (host:port)
- `spec.queueManager` — queue manager name
- `spec.credentialsSecretRef` — name of the Secret above
- Optional TLS: `spec.tls` (CA from Secret, or dev-only `allowInsecureTLS` annotation)

Wait for `Ready=True`:

```bash
kubectl wait qmc/<name> --for=condition=Ready --timeout=120s
```

Full field reference: [INSTALL_AND_USE.md](../INSTALL_AND_USE.md#resource-reference).

## 3. Queue

Apply a `Queue` with `spec.connectionRef` pointing at the QMC name:

[`config/samples/messaging_v1alpha1_queue.yaml`](https://github.com/platformrelay/MKurator/blob/main/config/samples/messaging_v1beta1_queue.yaml)

Verify on the queue manager:

```bash
kubectl describe queue <name>
# or MQSC: DISPLAY QLOCAL(<name>)
```

## Related samples

| Variant | Sample |
| --- | --- |
| Alias queue | [queue_alias.yaml](https://github.com/platformrelay/MKurator/blob/main/config/samples/messaging_v1beta1_queue_alias.yaml) |
| Remote queue | [queue_remote.yaml](https://github.com/platformrelay/MKurator/blob/main/config/samples/messaging_v1beta1_queue_remote.yaml) |
| Topic | [topic.yaml](https://github.com/platformrelay/MKurator/blob/main/config/samples/messaging_v1beta1_topic.yaml) |
| Channel | [channel.yaml](https://github.com/platformrelay/MKurator/blob/main/config/samples/messaging_v1beta1_channel.yaml) |

Attribute drift policy: [ATTRIBUTE_RECONCILIATION.md](../ATTRIBUTE_RECONCILIATION.md).

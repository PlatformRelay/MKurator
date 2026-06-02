# Kurator documentation

| | Start here |
|---|------------|
| 🎯 | Install and manage MQ objects on your queue manager → [INSTALL_AND_USE.md](INSTALL_AND_USE.md) |
| 🛠️ | Hack on the operator locally → [DEVELOPMENT.md](DEVELOPMENT.md) |
| 🏗️ | Understand design and reconciliation → [ARCHITECTURE.md](ARCHITECTURE.md) |
| 🤖 | AI agent conventions and workflow → [../AGENTS.md](../AGENTS.md) |

## All docs

| | Document | Covers |
|---|----------|--------|
| 🎯 | [INSTALL_AND_USE.md](INSTALL_AND_USE.md) | Install, connect, CRs, troubleshooting, uninstall |
| 🔧 | [../config/samples/README.md](../config/samples/README.md) | Annotated sample Secret, Connection, Queue, Topic, Channel YAML |
| 🔧 | [../charts/kurator/README.md](../charts/kurator/README.md) | Helm chart to install the operator |
| 🛠️ | [DEVELOPMENT.md](DEVELOPMENT.md) | Prerequisites, inner loop, local platform, deploy, test tiers |
| 🛠️ | [../hack/kind-cluster/README.md](../hack/kind-cluster/README.md) | kind + Terraform + IBM MQ platform only |
| 🛠️ | [IBM_MQ_101.md](IBM_MQ_101.md) | MQ console, `runmqsc`, verify Kurator on kind |
| 🏗️ | [ARCHITECTURE.md](ARCHITECTURE.md) | Components, runtime, CRDs, reconcile flow, security |
| 🏗️ | [ATTRIBUTE_RECONCILIATION.md](ATTRIBUTE_RECONCILIATION.md) | DEFINE vs DISPLAY drift matrix per MQ object type |
| 🏗️ | [LOGGING.md](LOGGING.md) | Structured logging configuration and guidelines |
| 🏗️ | [adr/](adr/) | Architecture Decision Records |
| 📋 | [ROADMAP.md](ROADMAP.md) | Phased delivery plan |
| 📋 | [PHASE5_AUTH_SKETCH.md](PHASE5_AUTH_SKETCH.md) | Phase 5 CHLAUTH / AUTHREC CR sketch |
| 📋 | [NON_FUNCTIONAL_REQUIREMENTS.md](NON_FUNCTIONAL_REQUIREMENTS.md) | Security, reliability, observability, performance |
| 📋 | [CICD.md](CICD.md) | CI/CD pipeline and `verify` discipline |
| 🔒 | [../SECURITY.md](../SECURITY.md) | Security posture and vulnerability reporting |
| 📚 | [IBM_MQ_OBJECTS.md](IBM_MQ_OBJECTS.md) | MQSC research inventory (**not** the product API) |
| 📚 | [IBM_MQ_REST_API.md](IBM_MQ_REST_API.md) | How the `mqweb` REST API is consumed |
| 📚 | [schemas/README.md](schemas/README.md) | mqweb Swagger / MQSC JSON schemas |
| 📚 | [REFERENCES.md.example](REFERENCES.md.example) | Optional map of vendored `references/` clones (copy to `REFERENCES.md`) |

**Contract vs research:** shipped behaviour is defined by v1alpha1 CRDs and
[ATTRIBUTE_RECONCILIATION.md](ATTRIBUTE_RECONCILIATION.md), not
[IBM_MQ_OBJECTS.md](IBM_MQ_OBJECTS.md).

## Emoji key (navigation only)

| Emoji | Meaning |
|-------|---------|
| 🎯 | End-user / operator |
| 🛠️ | Developer workflow |
| 🏗️ | Architecture and design |
| 📋 | Roadmap, NFRs, CI |
| 📚 | IBM MQ reference / research |
| 🔧 | Samples, Helm, config |
| 🔒 | Security |

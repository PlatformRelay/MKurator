# IBM MQ REST API schemas

IBM does **not** ship a static OpenAPI file in the product download or on GitHub. The **complete** mqweb schema is generated at runtime by WebSphere Liberty’s API Discovery feature.

## Files in this directory

| File | Source | Purpose |
|------|--------|---------|
| `mqsc-runcommand.schema.json` | IBM documentation (stable) | Request/response shapes for the `/mqsc` admin endpoint |
| `mqweb-swagger.json` | **You generate this** | Full Swagger 2.0 from your queue manager |

## Obtain the complete Swagger schema

1. Enable API discovery in `mqwebuser.xml`:

```xml
<featureManager>
  <feature>apiDiscovery-1.0</feature>
</featureManager>
```

2. Restart mqweb / the queue manager pod.

3. Run from the repository root:

```bash
MQWEB_USER=admin MQWEB_PASS=changeme \
  ./scripts/fetch-mqweb-swagger.sh https://localhost:9443 docs/schemas/mqweb-swagger.json
```

Or against the Docker integration MQ:

```bash
task mq:integration:up
task mq:integration:wait
task mq:swagger:fetch
```

4. Commit `mqweb-swagger.json` with a note of the IBM MQ version it was exported from (e.g. `9.4.2`).

Interactive explorer: `https://host:port/ibm/api/explorer`

## Versioning

Pin the exported Swagger to your target MQ version. Re-fetch when upgrading queue managers. Target **REST API v3** for new work (`/ibmmq/rest/v3/...`).

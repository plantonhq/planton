# Pulumi Module to Deploy AlicloudLogProject

This module provisions an Alibaba Cloud Simple Log Service (SLS) project with
bundled log stores and full-text indexes. For each log store defined in the
manifest, the module creates the store and (when `enableIndex` is true) a
full-text search index.

Generated resources: `log.Project`, `log.Store` (per store), `log.StoreIndex`
(per store with indexing enabled).

## CLI Usage (OpenMCF Pulumi)

```bash
# Preview
openmcf pulumi preview \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .

# Update (apply)
openmcf pulumi update \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir . \
  --yes

# Refresh
openmcf pulumi refresh \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .

# Destroy
openmcf pulumi destroy \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .
```

## Credentials

Alibaba Cloud credentials are injected via environment variables by the runner:

- `ALIBABA_CLOUD_ACCESS_KEY_ID`
- `ALIBABA_CLOUD_ACCESS_KEY_SECRET`

The Pulumi `alicloud` provider reads these automatically. No credentials appear
in the manifest `spec`.

## Further Reading

- [`examples.md`](./examples.md) — Runnable manifest examples with CLI commands
- [`overview.md`](./overview.md) — Module architecture and design decisions
- [`../hack/manifest.yaml`](../hack/manifest.yaml) — Minimal test manifest

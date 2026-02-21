# Pulumi Module to Deploy AliCloudSaeApplication

## CLI usage (OpenMCF pulumi)

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

## Debugging

This module includes a `debug.sh` helper. To enable debugging, edit `Pulumi.yaml` and uncomment the `runtime.options.binary` line so Pulumi runs the program via the script:

```yaml
name: openmcf-alicloud-module-test
runtime:
  name: go
#  options:
#    binary: ./debug.sh
```

Then make the script executable and run your command (e.g., `preview` or `update`). See `docs/pages/docs/guide/debug-pulumi-modules.mdx` for full instructions.

```bash
chmod +x debug.sh
openmcf pulumi preview \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .
```

## Module Overview

This Pulumi module deploys an Alibaba Cloud SAE application using a single `sae.Application` resource. The module reads an `AliCloudSaeApplicationStackInput` protobuf message, initializes locals (tag merging, environment variable serialization), and provisions the application with the specified compute, networking, health check, update strategy, and logging configuration.

The module converts the `envs` map into the JSON array format that the SAE API expects, maps health check specs to the provider's `LivenessV2` and `ReadinessV2` types, and conditionally sets optional fields only when provided — avoiding zero-value overrides of provider defaults.

Both stack outputs (`app_id` and `app_name`) are exported for use by downstream components.

---

## Further Reading

- **[examples.md](./examples.md)**: Runnable manifests for common SAE application configurations.
- **[overview.md](./overview.md)**: Module architecture, file organization, and design decisions.
- **[hack/manifest.yaml](../hack/manifest.yaml)**: Minimal test manifest.

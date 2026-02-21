# Pulumi Module to Deploy AliCloudFunction

This module provisions an Alibaba Cloud Function Compute v3 function. It creates
a single `fc.V3Function` resource with conditional configuration blocks for VPC
networking, SLS logging, custom container/runtime settings, lifecycle hooks, NAS
mounts, and GPU acceleration.

Generated resources: `fc.V3Function`.

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

## Debugging

If `preview` or `update` fails:

1. Check that credentials are set (`echo $ALIBABA_CLOUD_ACCESS_KEY_ID`)
2. Verify the manifest is valid YAML and passes schema validation
3. Confirm the specified region supports FC v3
4. For VPC-attached functions, verify the VPC, VSwitch, and security group IDs exist
5. For custom-container functions, verify the image URI is accessible from FC

## Module Overview

The `module/` directory contains three files:

- `main.go` — Creates the alicloud provider and `fc.V3Function` resource with
  all optional configuration blocks (VPC, logging, container, runtime,
  lifecycle, NAS, GPU)
- `locals.go` — Initializes the locals struct, merges standard + user tags,
  provides helper functions for optional fields
- `outputs.go` — Defines output key constants (`function_id`, `function_name`,
  `function_arn`)

## Further Reading

- [`examples.md`](./examples.md) — Runnable manifest examples with CLI commands
- [`overview.md`](./overview.md) — Module architecture and design decisions
- [`../hack/manifest.yaml`](../hack/manifest.yaml) — Minimal test manifest

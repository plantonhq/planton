# Pulumi Module to Deploy AwsDynamodb

This Pulumi module provisions an AWS DynamoDB table using the Planton CLI. Use the hack manifest at `../hack/manifest.yaml` as a quick starting point.

## CLI

```bash
# Preview
planton pulumi preview \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .

# Update (apply)
planton pulumi update \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir . \
  --yes

# Refresh
planton pulumi refresh \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .

# Destroy
planton pulumi destroy \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .
```

## Examples
See `./examples.md` for sample manifests tailored to DynamoDB.

## Debugging
A helper script `debug.sh` is provided to launch the Pulumi program under Delve. To enable it, uncomment the binary option in `Pulumi.yaml`:

```yaml
runtime:
  name: go
  # options:
  #   binary: ./debug.sh
```

Then run the CLI commands above; the Pulumi engine will execute the local debug binary. For more details, see docs: docs/pages/docs/guide/debug-pulumi-modules.mdx



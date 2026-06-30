# Pulumi Module to Deploy AwsCloudFront

This Pulumi program deploys an `AwsCloudFront` distribution using the Planton CLI.

## CLI commands

```bash
# Preview
planton pulumi preview \
  --manifest ../hack/manifest.yaml \
  --stack <org>/<project>/<stack> \
  --module-dir .

# Update (apply)
planton pulumi update \
  --manifest ../hack/manifest.yaml \
  --stack <org>/<project>/<stack> \
  --module-dir . \
  --yes

# Refresh
planton pulumi refresh \
  --manifest ../hack/manifest.yaml \
  --stack <org>/<project>/<stack> \
  --module-dir .

# Destroy
planton pulumi destroy \
  --manifest ../hack/manifest.yaml \
  --stack <org>/<project>/<stack> \
  --module-dir .
```

## Examples
See `examples.md` in this directory for example manifests and flows. These mirror the root-level examples for CloudFront.

## Debugging
For local debugging, a `debug.sh` helper is provided. To enable it, uncomment the following in `Pulumi.yaml`:

```yaml
# options:
#   binary: ./debug.sh
```

Then run the preview/update commands as usual; Pulumi will execute the compiled binary under Delve.

For more details, refer to the docs page: docs/pages/docs/guide/debug-pulumi-modules.mdx



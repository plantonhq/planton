# Pulumi Module to Deploy AWSLambda

This module provisions an AWS Lambda function via the Planton CLI.

## CLI commands

```shell
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

See `examples.md` in this directory for sample manifests.

## Debugging

For local debugging, use `debug.sh` to run under Delve:

1) Uncomment the runtime binary option in `Pulumi.yaml`:

```yaml
runtime:
  name: go
  options:
    binary: ./debug.sh
```

2) Run the CLI commands above (e.g., `preview`, `update`).

`debug.sh` builds with `-gcflags "all=-N -l"` and starts `dlv` on port 2345.


For more details, see the debugging guide at `docs/pages/docs/guide/debug-pulumi-modules.mdx`.



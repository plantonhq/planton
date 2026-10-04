# Pulumi Module to Deploy AwsCloudFront

This module provisions an Amazon CloudFront distribution at the full provider
surface — origins with S3/custom/VPC arms, primary/failover origin groups,
default + path-matched ordered cache behaviors across both caching
generations, custom domains, error pages, geo restrictions, and access logs —
plus the folded satellites: per-origin Origin Access Controls (for S3 origins
that ask for one) and the CloudWatch additional-metrics monitoring
subscription.

## CLI commands

```shell
# Preview
planton pulumi preview \
  --manifest ../../e2e/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .

# Update (apply)
planton pulumi update \
  --manifest ../../e2e/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir . \
  --yes

# Refresh
planton pulumi refresh \
  --manifest ../../e2e/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .

# Destroy
planton pulumi destroy \
  --manifest ../../e2e/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .
```

Deploys propagate to every CloudFront edge location, so expect update/destroy
to take 5-15 minutes each while the distribution converges (and destroy first
disables the distribution, which is itself a propagation).

## Examples

See [`e2e/manifest.yaml`](../../e2e/manifest.yaml) and the kind's presets
for sample manifests covering private S3 origins with OAC and custom-domain
serving.

## Debugging

This module includes a `debug.sh` helper. To enable debugging, edit
`Pulumi.yaml` and uncomment the `runtime.options.binary` line so Pulumi runs
the program via the script:

```yaml
runtime:
  name: go
#  options:
#    binary: ./debug.sh
```

Then make the script executable and run your command (e.g., `preview` or
`update`). `debug.sh` builds with `-gcflags "all=-N -l"` and starts `dlv` on
port 2345. See `docs/pages/docs/guide/debug-pulumi-modules.mdx` for full
instructions.

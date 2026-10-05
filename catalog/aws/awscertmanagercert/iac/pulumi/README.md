# Pulumi Module to Deploy AwsCertManagerCert

This module provisions an AWS Certificate Manager (ACM) certificate in any of
ACM's three creation modes — requested (Amazon-issued, DNS or EMAIL validated),
imported (bring-your-own PEM material), or private (issued by an ACM-PCA
authority). For DNS-validated certificates with a managed Route53 zone it also
creates the validation CNAME records and (by default) waits for issuance.

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

## Examples

See [`e2e/manifest.yaml`](../../e2e/manifest.yaml) and the kind's presets for
sample manifests covering managed Route53 validation, external DNS, wildcard
domains, and imported certificates.

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

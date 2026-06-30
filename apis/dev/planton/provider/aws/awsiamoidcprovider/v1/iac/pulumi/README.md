# Pulumi Module to Deploy AwsIamOidcProvider

## CLI usage (Planton pulumi)

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

## Debugging

This module can be run under a debugger via a `debug.sh` helper. To enable debugging, edit `Pulumi.yaml` and uncomment the `runtime.options.binary` line so Pulumi runs the program via the script:

```yaml
name: awsiamoidcprovider-pulumi-project
runtime:
  name: go
#  options:
#    binary: ./debug.sh
```

Then make the script executable and run your command (e.g., `preview` or `update`). See `docs/pages/docs/guide/debug-pulumi-modules.mdx` for full instructions.

```bash
chmod +x debug.sh
planton pulumi preview \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .
```

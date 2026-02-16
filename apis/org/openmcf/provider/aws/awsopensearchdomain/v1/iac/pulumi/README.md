# Pulumi Module to Deploy AwsOpenSearchDomain

This Pulumi program deploys an Amazon OpenSearch Service domain using the OpenMCF API and module.

## Requirements
- OpenMCF CLI built locally
- Valid AWS credential provided via the CLI stack input (not in `spec`)

## CLI commands

Preview:

```shell
openmcf pulumi preview \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .
```

Update (apply):

```shell
openmcf pulumi update \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir . \
  --yes
```

Refresh:

```shell
openmcf pulumi refresh \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .
```

Destroy:

```shell
openmcf pulumi destroy \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir . \
  --yes
```

## Examples

See `../examples.md` for sample manifests.

## Debugging

Optionally enable debugging by setting a binary in `Pulumi.yaml` and using the `debug.sh` script.

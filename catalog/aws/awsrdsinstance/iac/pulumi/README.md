# Pulumi Module to Deploy AwsRdsInstance

This Pulumi program deploys an AWS RDS DB instance using the Planton API and module.

## Requirements
- Planton CLI built locally
- Valid AWS credential provided via the CLI IaC input (not in `spec`)

## CLI commands

Preview:

```shell
planton pulumi preview \
  --manifest ../../e2e/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .
```

Update (apply):

```shell
planton pulumi update \
  --manifest ../../e2e/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir . \
  --yes
```

Refresh:

```shell
planton pulumi refresh \
  --manifest ../../e2e/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .
```

Destroy:

```shell
planton pulumi destroy \
  --manifest ../../e2e/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir . \
  --yes
```

## Examples

See `./../../e2e/manifest.yaml` for sample manifests.

## Debugging

Use `./debug.sh` to run common commands.

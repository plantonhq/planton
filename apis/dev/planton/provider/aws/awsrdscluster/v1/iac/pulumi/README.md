# Pulumi Module to Deploy AwsRdsCluster

This Pulumi program deploys an AWS RDS Cluster (Aurora MySQL/PostgreSQL or Multi-AZ DB Cluster) using the Planton API and module.

## Requirements
- Planton CLI built locally
- Valid AWS credential provided via the CLI stack input (not in `spec`)

## CLI commands

Preview:

```shell
planton pulumi preview \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .
```

Update (apply):

```shell
planton pulumi update \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir . \
  --yes
```

Refresh:

```shell
planton pulumi refresh \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .
```

Destroy:

```shell
planton pulumi destroy \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir . \
  --yes
```

## Examples

See `../examples.md` for sample manifests.

## Debugging

Optionally enable debugging by setting a binary in `Pulumi.yaml` and using the `debug.sh` script.



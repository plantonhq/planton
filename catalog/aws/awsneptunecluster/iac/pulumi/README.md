# Pulumi Module to Deploy AwsNeptuneCluster

This Pulumi program deploys an AWS Neptune cluster (fully managed graph database supporting Gremlin and SPARQL) using the Planton API and module: the cluster, its subnet group, the managed cluster and instance parameter groups (for inline `parameters` / `instanceParameters`), per-name folded instances (pinned to the cluster's port, with `applyImmediately` and `skipFinalSnapshot` forwarded), and per-name folded custom endpoints. Send conditions match the Terraform module argument-for-argument.

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

## Presets

See `../../presets/` for sample manifests.

## Debugging

Optionally enable debugging by setting a binary in `Pulumi.yaml` and using the `debug.sh` script.

# Pulumi Module to Deploy AwsOpenSearchDomain

This Pulumi program deploys an Amazon OpenSearch Service domain using the Planton API and module: the domain (`module/domain.go`) plus its folded satellites (`module/satellites.go`) -- SAML Dashboards sign-in when `samlOptions` is configured, and one cross-account VPC endpoint grant per `authorizedVpcEndpointAccessAccounts` entry. Encryption at rest and node-to-node TLS default to ON; the encryption, software-update, and off-peak blocks are always sent explicitly (matching the Terraform module) so a `false` genuinely turns the setting off.

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

See `../../e2e/manifest.yaml` for sample manifests.

## Debugging

Optionally enable debugging by setting a binary in `Pulumi.yaml` and using the `debug.sh` script.

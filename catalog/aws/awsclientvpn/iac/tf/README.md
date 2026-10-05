# Terraform Module to Deploy AwsClientVpn

This module provisions an AWS Client VPN endpoint for secure remote access into AWS networks using OpenVPN.
It sets up the endpoint (authentication, tunnel shape, sessions, logging) plus its three folded satellites:
target network associations, authorization rules, and routes.

Generated `variables.tf` reflects the proto schema for `AwsClientVpn`.

## Usage

Use the Planton CLI (tofu) with the default local backend:

```shell
planton tofu init --manifest e2e/manifest.yaml
planton tofu plan --manifest e2e/manifest.yaml
planton tofu apply --manifest e2e/manifest.yaml --auto-approve
planton tofu destroy --manifest e2e/manifest.yaml --auto-approve
```

**Note**: Credentials are provided via IaC input (CLI), not in the manifest `spec`.

For a full-surface example, see [`e2e/manifest.yaml`](../../e2e/manifest.yaml).



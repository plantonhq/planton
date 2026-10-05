# Terraform Module to Deploy AwsIamPolicy

This module provisions a customer-managed IAM policy: a standalone, versioned
permission document with its own ARN that can be attached to many roles and
users at once (via their `managedPolicyArns` fields) or used as a permissions
boundary.

Generated `variables.tf` reflects the proto schema for `AwsIamPolicy`.

## Usage

Use the Planton CLI (tofu) with the default local backend:

```shell
planton tofu init --manifest e2e/manifest.yaml
planton tofu plan --manifest e2e/manifest.yaml
planton tofu apply --manifest e2e/manifest.yaml --auto-approve
planton tofu destroy --manifest e2e/manifest.yaml --auto-approve
```

**Note**: Credentials are provided via IaC input (CLI), not in the manifest `spec`.

For a ready-to-run fixture, see [`e2e/manifest.yaml`](../../e2e/manifest.yaml).

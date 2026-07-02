# Pulumi Module to Deploy AwsIamPolicy

This module provisions a customer-managed IAM policy: a standalone, versioned
permission document with its own ARN that can be attached to many roles and
users at once (via their `managedPolicyArns` fields) or used as a permissions
boundary.

## Usage

Use the Planton CLI with the default local backend:

```shell
planton pulumi up --manifest ../hack/manifest.yaml --module-dir .
planton pulumi destroy --manifest ../hack/manifest.yaml --module-dir .
```

**Note**: Credentials are provided via stack input (CLI), not in the manifest `spec`.

For a ready-to-run fixture, see [`hack/manifest.yaml`](../hack/manifest.yaml).

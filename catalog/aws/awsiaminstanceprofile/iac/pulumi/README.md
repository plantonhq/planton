# Pulumi Module to Deploy AwsIamInstanceProfile

This module provisions an IAM instance profile: the container that delivers an
IAM role to EC2 instances. The profile carries exactly one role (referenced
from an `AwsIamRole` component or passed as a literal role name) and is what
EC2 instances, launch templates, and Auto Scaling groups reference.

## Usage

Use the Planton CLI with the default local backend:

```shell
planton pulumi up --manifest ../../e2e/manifest.yaml --module-dir .
planton pulumi destroy --manifest ../../e2e/manifest.yaml --module-dir .
```

**Note**: Credentials are provided via IaC input (CLI), not in the manifest `spec`.

For a ready-to-run fixture, see [`e2e/manifest.yaml`](../../e2e/manifest.yaml).

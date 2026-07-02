# Terraform Module to Deploy AwsIamInstanceProfile

This module provisions an IAM instance profile: the container that delivers an
IAM role to EC2 instances. The profile carries exactly one role (referenced
from an `AwsIamRole` component or passed as a literal role name) and is what
EC2 instances, launch templates, and Auto Scaling groups reference.

Generated `variables.tf` reflects the proto schema for `AwsIamInstanceProfile`.

## Usage

Use the Planton CLI (tofu) with the default local backend:

```shell
planton tofu init --manifest hack/manifest.yaml
planton tofu plan --manifest hack/manifest.yaml
planton tofu apply --manifest hack/manifest.yaml --auto-approve
planton tofu destroy --manifest hack/manifest.yaml --auto-approve
```

**Note**: Credentials are provided via stack input (CLI), not in the manifest `spec`.

For a ready-to-run fixture, see [`hack/manifest.yaml`](../hack/manifest.yaml).

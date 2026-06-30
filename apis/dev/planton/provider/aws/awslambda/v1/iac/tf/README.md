# Terraform Module to Deploy AWSLambda

This module deploys an `AWSLambda` resource using Terraform via the Planton CLI (tofu).

## CLI

```bash
planton tofu init --manifest hack/manifest.yaml
planton tofu plan --manifest hack/manifest.yaml
planton tofu apply --manifest hack/manifest.yaml --auto-approve
planton tofu destroy --manifest hack/manifest.yaml --auto-approve
```

- Credentials are provided via the CLI stack input, not stored in the manifest `spec`.
- Example manifest: see `apis/project/planton/provider/aws/awslambda/v1/iac/hack/manifest.yaml`.

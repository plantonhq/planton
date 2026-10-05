# Terraform Module to Deploy AwsAppRunnerService

This module deploys an `AwsAppRunnerService` resource using Terraform via the Planton CLI (tofu).

## CLI

```bash
planton tofu init --manifest e2e/manifest.yaml
planton tofu plan --manifest e2e/manifest.yaml
planton tofu apply --manifest e2e/manifest.yaml --auto-approve
planton tofu destroy --manifest e2e/manifest.yaml --auto-approve
```

- Credentials are provided via the CLI IaC input, not stored in the manifest `spec`.
- Example manifest: see `catalog/aws/awsapprunnerservice/e2e/manifest.yaml`.

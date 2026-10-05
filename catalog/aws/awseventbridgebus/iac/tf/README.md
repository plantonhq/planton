## Terraform Module to Deploy AwsEventBridgeBus

Run the module via the Planton CLI (tofu) using the default local backend.

```shell
planton tofu init --manifest e2e/manifest.yaml
planton tofu plan --manifest e2e/manifest.yaml
planton tofu apply --manifest e2e/manifest.yaml --auto-approve
planton tofu destroy --manifest e2e/manifest.yaml --auto-approve
```

- Credentials are provided via IaC input (by the CLI), not in the manifest `spec`.
- Manifest file: `../../e2e/manifest.yaml`

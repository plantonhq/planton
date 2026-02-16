## Terraform Module to Deploy AwsSqsQueue

Run the module via the OpenMCF CLI (tofu) using the default local backend.

```shell
openmcf tofu init --manifest hack/manifest.yaml
openmcf tofu plan --manifest hack/manifest.yaml
openmcf tofu apply --manifest hack/manifest.yaml --auto-approve
openmcf tofu destroy --manifest hack/manifest.yaml --auto-approve
```

- Credentials are provided via stack input (by the CLI), not in the manifest `spec`.
- Manifest file: `../hack/manifest.yaml`

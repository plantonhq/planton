## Pulumi Module to Deploy AwsSnsTopic

Run the module via the Planton CLI (pulumi) using the default local backend.

```shell
planton pulumi preview --manifest hack/manifest.yaml
planton pulumi up --manifest hack/manifest.yaml --yes
planton pulumi destroy --manifest hack/manifest.yaml --yes
```

- Credentials are provided via stack input (by the CLI), not in the manifest `spec`.
- Manifest file: `../hack/manifest.yaml`

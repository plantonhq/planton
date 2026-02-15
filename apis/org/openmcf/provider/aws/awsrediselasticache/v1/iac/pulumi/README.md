## Pulumi Module to Deploy AwsRedisElasticache

Run the module via the OpenMCF CLI (pulumi) using the default local backend.

```shell
openmcf pulumi preview --manifest hack/manifest.yaml
openmcf pulumi up --manifest hack/manifest.yaml --yes
openmcf pulumi destroy --manifest hack/manifest.yaml --yes
```

- Credentials are provided via stack input (by the CLI), not in the manifest `spec`.
- Manifest file: `../hack/manifest.yaml`

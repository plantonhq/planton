## Pulumi Module to Deploy AwsRedisElasticache

Run the module via the Planton CLI (pulumi) using the default local backend.

```shell
planton pulumi preview --manifest e2e/manifest.yaml
planton pulumi up --manifest e2e/manifest.yaml --yes
planton pulumi destroy --manifest e2e/manifest.yaml --yes
```

- Credentials are provided via IaC input (by the CLI), not in the manifest `spec`.
- Manifest file: `../../e2e/manifest.yaml`

# Pulumi Module to Deploy KubernetesGoFeatureFlagFlagFile

This module renders a GO Feature Flag flag file into one ConfigMap named after the resource, under the spec's data key (default `flags.goff.yaml`). `flag_file.go` renders the flags as one JSON object keyed by flag name in GO Feature Flag's flag format (JSON is YAML, so the relay's default `yaml` file format reads it); `flag_file_test.go` pins the shape. A `KubernetesGoFeatureFlag` relay's `config_map` retriever reads the ConfigMap on every poll, so an edit reaches evaluations within one polling interval and nothing restarts.

It exports the same outputs as the Terraform module: `config_map_name`, `key` and `namespace`.

## CLI usage (Planton pulumi)

```bash
# Preview
planton pulumi preview \
  --manifest ../../e2e/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .

# Update (apply)
planton pulumi update \
  --manifest ../../e2e/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir . \
  --yes

# Destroy
planton pulumi destroy \
  --manifest ../../e2e/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .
```

Credentials (the target cluster's kubeconfig) arrive through the IaC input's `provider_config`, never in the manifest `spec`.

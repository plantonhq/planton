# Terraform Module to Deploy KubernetesGoFeatureFlagFlagFile

This module renders a GO Feature Flag flag file into `kubernetes_config_map_v1.flag_file`, named after the resource, under the spec's data key (default `flags.goff.yaml`). `locals.tf` renders the flags byte-identically to the Pulumi module (Go `json.Marshal` and OpenTofu `jsonencode` sort keys and escape alike); the flags arrive untyped because variations hold free-form values, so every field is read with `try()`.

Outputs: `config_map_name`, `key`, `namespace` -- identical to the Pulumi module.

Generated `variables.tf` reflects the proto schema for `KubernetesGoFeatureFlagFlagFile`.

## Usage

Use the Planton CLI (tofu) with the default local backend:

```shell
planton tofu init --manifest ../../e2e/manifest.yaml
planton tofu plan --manifest ../../e2e/manifest.yaml
planton tofu apply --manifest ../../e2e/manifest.yaml --auto-approve
planton tofu destroy --manifest ../../e2e/manifest.yaml --auto-approve
```

**Note**: Credentials are provided via IaC input (CLI), not in the manifest `spec`.

For a complete manifest, see [`e2e/manifest.yaml`](../../e2e/manifest.yaml).

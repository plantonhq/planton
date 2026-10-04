# KubernetesExternalSecret Terraform Module

## Module Behavior

- **Twin locals**: `locals.tf` renders the same CRD-JSON spec shape the
  Pulumi module's `spec_builder.go` produces (null-prune idiom) — identical
  field mappings, kept in lockstep, so the two engines apply
  byte-equivalent declarations.
- **`kubectl_manifest` apply**: the CR applies through the alekc/kubectl
  provider, which needs no cluster connection at plan time — an
  ExternalSecret can be PLANNED before the External Secrets Operator's CRDs
  exist, which is what lets an Infra Chart deploy the operator, its stores,
  and its syncs in one run.
- **Pinned Secret name**: the CR's `target.name` is always rendered from
  the resolved Secret name (`target.name` when set, else `metadata.name`),
  so the exported `secret_name` output can never drift from what the
  operator creates.
- **No `wait_for` block, deliberately**: the materialized Secret appears
  when the operator reaches the backend, which is not part of applying the
  resource — the never-block-on-a-controller posture.

## Usage

```bash
planton tofu apply --manifest external-secret.yaml
```

## Local Development

```bash
terraform init
terraform validate
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Inputs

See `variables.tf` for the full variable specification.

## Outputs

| Output | Description |
|--------|-------------|
| `external_secret_name` | Name of the created ExternalSecret |
| `namespace` | Namespace the ExternalSecret and its materialized Secret live in |
| `secret_name` | The materialized Secret — the handle workloads reference |

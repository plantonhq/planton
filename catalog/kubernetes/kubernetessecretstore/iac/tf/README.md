# KubernetesSecretStore Terraform Module

## Module Behavior

- **Twin locals**: `locals.tf` renders the same CRD-JSON spec shape the
  Pulumi module's shared builder produces — identical field mappings, kept
  in lockstep, so the two engines apply byte-equivalent stores.
- **`kubectl_manifest` apply**: the CR applies through the alekc/kubectl
  provider, which needs no cluster connection at plan time — a SecretStore
  can be PLANNED before the External Secrets Operator's CRDs exist, which
  is what lets an Infra Chart deploy the operator and its stores in one
  run.
- **Credential Secret materialization**: static credentials declared in the
  spec land in a `<resource-name>-credentials` Secret in the store's own
  namespace; the CR depends on it, so ESO never observes dangling
  secretRefs.
- **No `wait_for` block, deliberately**: store readiness depends on
  external reachability (the cloud secrets API, Vault) that is not part of
  applying the resource — the never-block-on-a-controller posture.

## Usage

```bash
planton tofu apply --manifest secret-store.yaml
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
| `store_name` | Name of the created SecretStore — the handle ExternalSecrets in the same namespace reference |
| `namespace` | Namespace the store and its credential Secrets live in |

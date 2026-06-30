# KubernetesReferenceGrant Terraform Module

Creates a namespaced Kubernetes Gateway API `ReferenceGrant` via the
`kubernetes_manifest` resource (apiVersion `gateway.networking.k8s.io/v1`). The
Gateway API CRDs must already be installed on the target cluster (see
`KubernetesGatewayApiCrds`), and the target namespace must exist (see
`KubernetesNamespace`). This is the "to" namespace -- the one whose resources the
grant authorizes inbound cross-namespace references to.

## Usage

```bash
planton tofu apply --manifest referencegrant.yaml
```

## Local Development

```bash
terraform init
terraform validate
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Inputs

See `variables.tf` for the full variable specification. `namespace` is a plain
string: the platform resolves its `StringValueOrRef` foreign key to a literal
before Terraform runs. The `from` and `to` entries are trust assertions about
KINDS of resources, not foreign keys to specific objects. The one genuine
cross-resource reference is `from[].namespace` (a source namespace); when it is
Planton-managed, infra-chart authors wire that DAG edge via
`metadata.relationships` (DD-009).

## Outputs

| Output | Description |
|--------|-------------|
| `reference_grant_name` | Name of the created ReferenceGrant (equals `metadata.name`) |
| `namespace` | Namespace the ReferenceGrant was created in |

# KubernetesReferenceGrant Terraform Module

Creates a namespaced Kubernetes Gateway API `ReferenceGrant` via the
`kubectl_manifest` resource (alekc/kubectl provider, apiVersion
`gateway.networking.k8s.io/v1`, server-side apply). Unlike
`kubernetes_manifest`, `kubectl_manifest` needs no cluster connection at plan
time, so the grant can be planned before the Gateway API CRDs exist -- which is
what lets an Infra Chart deploy the CRDs and its grants in a single run (and
lets offline plan proofs work).

Prerequisites at apply time: the Gateway API CRDs (`KubernetesGatewayApiCrds`)
and the target namespace (see `KubernetesNamespace`). This is the "to"
namespace -- the one whose resources the grant authorizes inbound
cross-namespace references to.

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
`metadata.relationships`.

## State Import

Existing ReferenceGrants can be adopted into state. `kubectl_manifest` uses the
composed import ID `apiVersion//kind//name//namespace`; the component's
`iac/import-map.yaml` derives each part (apiVersion and kind are constants of
this module).

## Outputs

| Output | Description |
|--------|-------------|
| `reference_grant_name` | Name of the created ReferenceGrant (equals `metadata.name`) |
| `namespace` | Namespace the ReferenceGrant was created in |

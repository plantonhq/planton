# KubernetesTcpRoute Terraform Module

Creates a namespaced Kubernetes Gateway API `TCPRoute` via the `kubectl_manifest`
resource (alekc/kubectl provider, apiVersion `gateway.networking.k8s.io/v1`,
server-side apply). Unlike `kubernetes_manifest`, `kubectl_manifest` needs no
cluster connection at plan time, so the route can be planned before the Gateway
API CRDs exist -- which is what lets an Infra Chart deploy the CRDs, a Gateway,
and its routes in a single run (and lets offline plan proofs work).

Prerequisites at apply time: the Gateway API standard-channel CRDs
(`KubernetesGatewayApiCrds`), the `Gateway` the route attaches to via
`parentRefs` with a `TCP` listener (see `KubernetesGateway`), and the target
namespace (see `KubernetesNamespace`).

## Usage

```bash
planton tofu apply --manifest tcproute.yaml
```

## Local Development

```bash
terraform init
terraform validate
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Inputs

See `variables.tf` for the full variable specification. The spec arrives from the
proto->tfvars converter already manifest-shaped (camelCase keys, null-pruned),
with every `StringValueOrRef` foreign key -- `namespace`, `parentRefs[].name`
(KubernetesGateway), `backendRefs[].name` (KubernetesService) -- resolved to a
literal string before Terraform runs.

## State Import

Existing TCPRoutes can be adopted into state. `kubectl_manifest` uses the
composed import ID `apiVersion//kind//name//namespace`; the kind's
`iac/import-map.yaml` derives each part (apiVersion and kind are constants of
this module).

## Outputs

| Output | Description |
|--------|-------------|
| `route_name` | Name of the created TCPRoute (equals `metadata.name`) |
| `namespace` | Namespace the TCPRoute was created in |

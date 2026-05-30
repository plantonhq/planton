# KubernetesTcpRoute Terraform Module

Creates a namespaced Kubernetes Gateway API `TCPRoute` via the
`kubernetes_manifest` resource (apiVersion `gateway.networking.k8s.io/v1alpha2`).
TCPRoute is an experimental-channel resource: the Gateway API **experimental**
CRDs must already be installed on the target cluster (`KubernetesGatewayApiCrds`
with `install_channel: experimental`), the `Gateway` the route attaches to via
`parentRefs` must exist with a `TCP` listener (see `KubernetesGateway`), and the
target namespace must exist (see `KubernetesNamespace`).

## Usage

```bash
openmcf tofu apply --manifest tcproute.yaml
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
before Terraform runs. `parent_refs` and `backend_refs` are plain upstream
references (matched by name), not foreign keys -- infra-chart authors wire their
DAG edges via `metadata.relationships` (DD-009).

## Outputs

| Output | Description |
|--------|-------------|
| `route_name` | Name of the created TCPRoute (equals `metadata.name`) |
| `namespace` | Namespace the TCPRoute was created in |

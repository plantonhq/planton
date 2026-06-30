# KubernetesHttpRoute Terraform Module

Creates a namespaced Kubernetes Gateway API `HTTPRoute` via the
`kubernetes_manifest` resource. The Gateway API CRDs must already be installed
on the target cluster (see `KubernetesGatewayApiCrds`), the `Gateway` the route
attaches to via `parentRefs` must exist (see `KubernetesGateway`), and the
target namespace must exist (see `KubernetesNamespace`).

## Usage

```bash
planton tofu apply --manifest httproute.yaml
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
references (matched by name), not foreign keys.

## Outputs

| Output | Description |
|--------|-------------|
| `route_name` | Name of the created HTTPRoute (equals `metadata.name`) |
| `namespace` | Namespace the HTTPRoute was created in |

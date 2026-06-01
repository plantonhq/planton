# KubernetesGateway Terraform Module

Creates a namespaced Kubernetes Gateway API `Gateway` via the
`kubernetes_manifest` resource. The Gateway API CRDs must already be installed
on the target cluster (see `KubernetesGatewayApiCrds`), a controller-backed
`GatewayClass` must exist (see `KubernetesGatewayClass`), and the target
namespace must exist (see `KubernetesNamespace`).

## Usage

```bash
openmcf tofu apply --manifest gateway.yaml
```

## Local Development

```bash
terraform init
terraform validate
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Inputs

See `variables.tf` for the full variable specification. `namespace` and
`gateway_class_name` are plain strings: the platform resolves their
`StringValueOrRef` foreign keys to literals before Terraform runs.

## Outputs

| Output | Description |
|--------|-------------|
| `gateway_name` | Name of the created Gateway (equals `metadata.name`) |
| `namespace` | Namespace the Gateway was created in |
| `gateway_class_name` | Name of the GatewayClass this Gateway belongs to |

# KubernetesGatewayClass Terraform Module

Creates a cluster-scoped Kubernetes Gateway API `GatewayClass` via the
`kubernetes_manifest` resource. The Gateway API CRDs must already be installed
on the target cluster (see the `KubernetesGatewayApiCrds` component).

## Usage

```bash
planton tofu apply --manifest gateway-class.yaml
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
| `gateway_class_name` | Name of the created GatewayClass (equals `metadata.name`) |
| `controller_name` | The controller managing this GatewayClass |

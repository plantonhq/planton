# KubernetesGatewayClass Terraform Module

Creates a cluster-scoped Kubernetes Gateway API `GatewayClass` via the
`kubectl_manifest` resource (alekc/kubectl provider, apiVersion
`gateway.networking.k8s.io/v1`, server-side apply). Unlike
`kubernetes_manifest`, `kubectl_manifest` needs no cluster connection at plan
time, so the class can be planned before the Gateway API CRDs exist -- which is
what lets an Infra Chart deploy the CRDs and the class in a single run (and lets
offline plan proofs work). At apply time the Gateway API CRDs must be installed
(see the `KubernetesGatewayApiCrds` component).

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

## State Import

Existing GatewayClasses can be adopted into state. `kubectl_manifest` uses the
composed import ID `apiVersion//kind//name` (no namespace -- GatewayClass is
cluster-scoped); the component's `iac/import-map.yaml` derives each part
(apiVersion and kind are constants of this module).

## Outputs

| Output | Description |
|--------|-------------|
| `gateway_class_name` | Name of the created GatewayClass (equals `metadata.name`) |
| `controller_name` | The controller managing this GatewayClass |

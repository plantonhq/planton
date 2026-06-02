# KubernetesServiceEntry Terraform Module

Creates a namespaced Istio `ServiceEntry` via the `kubernetes_manifest` resource. The
Istio CRDs must already be installed on the target cluster (see
`KubernetesIstioBaseCrds`), a running istiod is required to program the registry (see
`KubernetesIstio`), and the target namespace must exist (see `KubernetesNamespace`).

## Usage

```bash
openmcf tofu apply --manifest serviceentry.yaml
```

## Local Development

```bash
terraform init
terraform validate
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Inputs

See `variables.tf` for the full variable specification. `namespace` is a plain string:
the platform resolves its `StringValueOrRef` foreign key to a literal before Terraform
runs. `hosts` is required; `addresses`, `ports`, `location`, `resolution`,
`endpoints`, `export_to`, `subject_alt_names`, and `workload_selector` are optional and
null-pruned, so unset fields are omitted from the manifest and upstream defaults flow
through. `endpoints` and `workload_selector` are mutually exclusive.

## Outputs

| Output | Description |
|--------|-------------|
| `service_entry_name` | Name of the created ServiceEntry (equals `metadata.name`) |
| `namespace` | Namespace the ServiceEntry was created in |

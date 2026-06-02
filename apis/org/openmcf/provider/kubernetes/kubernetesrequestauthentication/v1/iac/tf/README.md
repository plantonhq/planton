# KubernetesRequestAuthentication Terraform Module

Creates a namespaced Istio `RequestAuthentication` via the `kubernetes_manifest`
resource. The Istio CRDs must already be installed on the target cluster (see
`KubernetesIstioBaseCrds`), a running istiod is required to enforce the policy in the
data plane (see `KubernetesIstio`), and the target namespace must exist (see
`KubernetesNamespace`).

## Usage

```bash
openmcf tofu apply --manifest requestauthentication.yaml
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
string: the platform resolves its `StringValueOrRef` foreign key to a literal before
Terraform runs. `selector.match_labels` and `target_refs` are plain references (not
foreign keys), and are mutually exclusive. `jwt_rules` and all of their nested
optional fields are null-pruned, so unset fields are omitted from the manifest and
upstream defaults flow through.

## Outputs

| Output | Description |
|--------|-------------|
| `request_authentication_name` | Name of the created RequestAuthentication (equals `metadata.name`) |
| `namespace` | Namespace the RequestAuthentication was created in |

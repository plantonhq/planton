# KubernetesPeerAuthentication Terraform Module

Creates a namespaced Istio `PeerAuthentication` via the `kubernetes_manifest`
resource. The Istio CRDs must already be installed on the target cluster (see
`KubernetesIstioBaseCrds`), a running istiod is required to enforce the policy in
the data plane (see `KubernetesIstio`), and the target namespace must exist (see
`KubernetesNamespace`).

## Usage

```bash
openmcf tofu apply --manifest peerauthentication.yaml
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
before Terraform runs. `selector.match_labels` is a plain label match (not a
foreign key). `mtls` and `port_level_mtls` are optional; omitting them lets the
policy inherit its mode from the parent namespace/mesh policy.

## Outputs

| Output | Description |
|--------|-------------|
| `peer_authentication_name` | Name of the created PeerAuthentication (equals `metadata.name`) |
| `namespace` | Namespace the PeerAuthentication was created in |

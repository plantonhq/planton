# KubernetesEnvoyFilter Terraform Module

Creates a namespaced Istio `EnvoyFilter` via the `kubernetes_manifest` resource (emitting
`networking.istio.io/v1alpha3` -- EnvoyFilter has not graduated to v1). The Istio CRDs must
already be installed on the target cluster (see `KubernetesIstioBaseCrds`), a running istiod is
required to translate the patches (see `KubernetesIstio`), and the target namespace must exist
(see `KubernetesNamespace`).

## Usage

```bash
planton tofu apply --manifest envoyfilter.yaml
```

## Local Development

```bash
terraform init
terraform validate
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Inputs

See `variables.tf` for the full variable specification. `namespace` is a plain string: the
platform resolves its `StringValueOrRef` foreign key to a literal before Terraform runs.
`workload_selector`, `target_refs`, `config_patches`, and `priority` are optional and
null-pruned, so unset fields are omitted from the manifest and upstream defaults flow through.
`workload_selector` and `target_refs` are mutually exclusive. The free-form `config_patches[].
patch.value` is typed `any` and passes through unmodified (the upstream CRD marks it
preserveUnknownFields). Snake_case spec fields are mapped to the CRD's camelCase keys.

## Outputs

| Output | Description |
|--------|-------------|
| `envoy_filter_name` | Name of the created EnvoyFilter (equals `metadata.name`) |
| `namespace` | Namespace the EnvoyFilter was created in |

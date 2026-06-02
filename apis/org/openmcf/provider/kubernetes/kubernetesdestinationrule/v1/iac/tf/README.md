# KubernetesDestinationRule Terraform Module

Creates a namespaced Istio `DestinationRule` via the `kubernetes_manifest` resource. The
Istio CRDs must already be installed on the target cluster (see `KubernetesIstioBaseCrds`),
a running istiod is required to apply the policy (see `KubernetesIstio`), and the target
namespace must exist (see `KubernetesNamespace`).

## Usage

```bash
openmcf tofu apply --manifest destinationrule.yaml
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
`host` is required; `traffic_policy`, `subsets`, `export_to`, and `workload_selector` are
optional.

`locals.tf` builds the manifest from object constructors with value-or-`null` per field --
`kubernetes_manifest` treats a null attribute as unset, so omitted fields fall through to
istiod defaults. This deliberately avoids `merge()` of uniform-type conditional fragments,
which would collapse to a `map(...)` the provider cannot morph into an object (the watch
items are the single-field `ring_hash`, `maglev`, and `port` objects). Because HCL has no
functions, the same `TrafficPolicy` shape that appears at the spec, subset, and
`port_level_settings` paths is transformed once per leaf via a path-keyed map and looked up
during assembly. Snake_case spec fields map to the CRD's camelCase (`loadBalancer`,
`connectionPool`, `outlierDetection`, `consecutive5xxErrors`, `baseEjectionTime`,
`credentialName`, `portLevelSettings`, `workloadSelector`, ...).

## Outputs

| Output | Description |
|--------|-------------|
| `destination_rule_name` | Name of the created DestinationRule (equals `metadata.name`) |
| `namespace` | Namespace the DestinationRule was created in |

# KubernetesTelemetry Terraform Module

Creates a namespaced Istio `Telemetry` resource via the `kubernetes_manifest` resource. The
Istio CRDs must already be installed on the target cluster (see `KubernetesIstioBaseCrds`),
a running istiod is required to apply the configuration (see `KubernetesIstio`), and the
target namespace must exist (see `KubernetesNamespace`).

## Usage

```bash
openmcf tofu apply --manifest telemetry.yaml
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
platform resolves its `StringValueOrRef` foreign key to a literal before Terraform runs. All
of `selector`, `target_refs`, `tracing`, `metrics`, and `access_logging` are optional.

`locals.tf` builds the manifest by pruning unset fields, with two shapes handled
differently:

- **Object-typed `oneOf` members are `merge()`-pruned** -- the `custom_tags`
  `{literal|environment|header}` are objects with required subfields, so only the chosen
  member is emitted (its required subfields seeded as the merge base). Emitting a non-chosen
  member as null would be sent as an empty `{}` and match a second `oneOf` arm.
- **Scalar `oneOf` members and uniform-type leaves are object constructors** with
  value-or-`null` -- the metrics override `match` (`metric`/`custom_metric`/`mode`) and
  `tag_overrides` values (`{operation, value}`). `kubernetes_manifest` prunes scalar nulls,
  so the metric-vs-custom_metric `oneOf` still sees only the field that was set; an
  all-conditional `merge()` of these uniform-string fields would instead collapse to a
  `map(string)` the provider cannot morph into an object.

Snake_case spec fields map to the CRD's camelCase (`randomSamplingPercentage`,
`disableSpanReporting`, `customTags`, `tagOverrides`, `reportingInterval`,
`useRequestIdForTraceSampling`, `accessLogging`, ...).

## Outputs

| Output | Description |
|--------|-------------|
| `telemetry_name` | Name of the created Telemetry resource (equals `metadata.name`) |
| `namespace` | Namespace the Telemetry resource was created in |

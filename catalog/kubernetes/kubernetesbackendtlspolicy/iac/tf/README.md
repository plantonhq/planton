# KubernetesBackendTlsPolicy Terraform Module

Creates a namespaced Kubernetes Gateway API `BackendTLSPolicy` via the
`kubectl_manifest` resource (alekc/kubectl provider, apiVersion
`gateway.networking.k8s.io/v1`, server-side apply). Unlike
`kubernetes_manifest`, `kubectl_manifest` needs no cluster connection at
plan time, so the policy can be planned before the Gateway API CRDs exist --
which is what lets an Infra Chart deploy the CRDs, a Gateway, routes, and
backend policies in a single run (and lets offline plan proofs work).

Prerequisites at apply time: the Gateway API standard-channel CRDs
(`KubernetesGatewayApiCrds`), the same-namespace backend Service(s) the
policy targets, and -- for the bring-your-own-CA arm -- the same-namespace
ConfigMap carrying the PEM CA bundle in a key named `ca.crt`.

No `wait_for` block, deliberately: the per-ancestor Accepted/ResolvedRefs
conditions appear when a Gateway controller reconciles the policy, which is
not part of applying it -- the same never-block-on-a-controller posture as
the route modules. Pulumi equivalent: a typed custom resource without await
logic.

## Rendering Notes

- The CR spec arrives from the proto->tfvars converter already
  manifest-shaped (camelCase keys, null-pruned, `StringValueOrRef` foreign
  keys resolved to literal strings); the module hands it to the engine
  verbatim minus the Planton `namespace` key, which maps to
  `metadata.namespace` rather than into the CR spec. The API server plus
  Planton protovalidate are the schema authority.
- **Null-prune, NOT empty-prune**, keeps the projection faithful: the
  `group` keys on `targetRefs` and `caCertificateRefs` are
  presence-required in the spec and carry the empty string for core-group
  referents (Service, ConfigMap), so the resolved tfvars always includes
  `group: ""` and it passes through to the CR -- the CRD rejects a missing
  key.
- The `wellKnownCACertificates` key arrives with its exact CRD casing
  (capital CA): the spec pins it via `json_name`, because the API server
  rejects the protojson-derived `wellKnownCaCertificates` as undeclared.

## Usage

```bash
planton tofu apply --manifest backendtlspolicy.yaml
```

## Local Development

```bash
terraform init
terraform validate
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Inputs

See `variables.tf` for the full variable specification. Foreign keys --
`namespace` (KubernetesNamespace), `targetRefs[].name` (KubernetesService),
`caCertificateRefs[].name` (KubernetesConfigMap) -- are resolved to literal
strings before Terraform runs.

## State Import

Existing BackendTLSPolicies can be adopted into state. `kubectl_manifest`
uses the composed import ID `apiVersion//kind//name//namespace`; the
kind's `iac/import-map.yaml` derives each part (apiVersion and kind
are constants of this module).

## Outputs

| Output | Description |
|--------|-------------|
| `policy_name` | Name of the created BackendTLSPolicy (equals `metadata.name`) |
| `namespace` | Namespace the BackendTLSPolicy was created in |

# KubernetesPeerAuthentication -- Design & Research

This document captures the deployment landscape, the modeling rationale, and the
fidelity decisions behind the `KubernetesPeerAuthentication` OpenMCF component
(kind 863).

## 1. What PeerAuthentication is

`PeerAuthentication` is an Istio security CRD (`security.istio.io`, served at both
`v1` and `v1beta1` -- identical schema) that controls **peer (service-to-service)
mutual TLS** for the workloads it selects. It answers one question: *must incoming
connections to these workloads be mTLS-tunneled?*

It is distinct from `RequestAuthentication` (which validates end-user JWTs) and
`AuthorizationPolicy` (which allows/denies requests). PeerAuthentication is purely
about the transport-level mTLS posture of the receiving proxy.

### The mTLS modes

| Mode | Behavior |
|------|----------|
| `UNSET` | Inherit from the parent policy; if there is none, behave as `PERMISSIVE`. |
| `DISABLE` | No mTLS -- connections are accepted as plaintext. |
| `PERMISSIVE` | Accept either plaintext or mTLS. The migration default. |
| `STRICT` | Require mTLS -- plaintext connections are rejected. |

### Policy hierarchy

Istio resolves the effective mode by specificity:

```
workload-level (selector present) > namespace-level (no selector) > mesh-level (root namespace, no selector)
```

`port_level_mtls` overrides the workload mode for specific ports, and is only
evaluated when a selector is present (a mesh- or namespace-wide policy cannot use
it because there is no single workload whose ports to address).

## 2. Source of truth and version

Translated proto-to-proto from the upstream `istio.io/api` clone pinned to tag
**1.26.8** (`security/v1beta1/peer_authentication.proto`). The local clone is
authoritative; no specs are pulled from the internet. The crd2pulumi
typed SDK and the CRDs installed by `KubernetesIstioBaseCrds` are likewise
generated from Istio `release-1.26`, so the proto, the typed Pulumi resource, and
the cluster CRD all agree on the schema.

The schema is small and stable: `selector`, `mtls.mode`, `port_level_mtls`. It has
**no oneofs and no `google.protobuf.Duration`** -- it exercises the full pipeline
(shared selector type, closed-set string enums, message-level CEL, typed Pulumi +
null-pruned Terraform, live E2E behind the CRDs prerequisite) without any union
discriminator machinery.

## 3. OpenMCF spec shape (fidelity decisions)

The OpenMCF spec flattens the upstream `PeerAuthentication` fields directly after
the namespaced envelope (`target_cluster`, `namespace`). There is no nested
`peer_authentication` sub-message.

| Upstream | OpenMCF | Notes |
|----------|---------|-------|
| `selector` (`type.v1beta1.WorkloadSelector`) | `selector` (`KubernetesIstioApiWorkloadSelector`) | Shared type in `istio_api.proto`, reused across the family. |
| `mtls` (`MutualTLS`) | `mtls` (`KubernetesPeerAuthenticationMutualTls`) | One field, `mode`. |
| `mtls.mode` (`Mode` enum) | `mtls.mode` (`string`) | Closed string set, not a proto enum. |
| `port_level_mtls` (`map<uint32, MutualTLS>`) | `port_level_mtls` (`map<uint32, KubernetesPeerAuthenticationMutualTls>`) | Reuses the same MutualTLS message, 1:1 with upstream. |

### Why `mtls.mode` is required-in-set (and the empty string is invalid)

Most OpenMCF enum-like fields are `optional string` with a CEL rule that also
admits `''` (treating empty as "unset, use upstream default"). PeerAuthentication
is deliberately different: the `MutualTLS` message's **only** field is `mode`, so a
present `mtls{}` block with `mode: ""` is meaningless and serializes a value the
API server rejects. The honest model is:

- Omit the entire `mtls` block to mean "inherit from parent".
- Set `mtls.mode` to a real value when you do specify it; `UNSET` is the explicit
  "inherit" value upstream provides for that purpose.

So `mode` is `required` with CEL `this in ['UNSET','DISABLE','PERMISSIVE','STRICT']`
(no empty branch). This is a conscious divergence from the gateway family idiom,
made because the empty string is not a valid wire value here.

### Enum case

The values are UPPERCASE (`STRICT`, `PERMISSIVE`, ...) because Istio is an
external standard. The field carries the
`// external standard exception` comment.

### Validation rules (each has accept + reject coverage in `spec_test.go`)

| Rule | Origin |
|------|--------|
| `peer_authentication_mtls.mode_enum` | Closed mode set. |
| `port_level_mtls` key range 1-65535 | Upstream port-key XValidation. |
| `peer_authentication.port_level_mtls_requires_selector` (message-level) | Upstream `portLevelMtls requires selector` XValidation. |
| selector `match_labels` non-empty keys / no-wildcard keys / no-wildcard values | Upstream WorkloadSelector XValidation (added to the shared `istio_api.proto` type, since PeerAuthentication is its first consumer). |

## 4. Composability

- **`namespace`** is the one true foreign key: `StringValueOrRef` ->
  `KubernetesNamespace` (`spec.name`). It can be a literal or a `valueFrom`
  reference, and creates a real DAG edge.
- **`selector.match_labels`** is a plain label match, NOT a foreign key. istiod
  resolves it against pod labels at runtime; it creates no automatic DAG edge to
  the workload it protects. Wrapping it in `StringValueOrRef` would break upstream
  fidelity (it is a multi-entry label map, not a scalar reference) and distort the
  typed CRD shape. An infra-chart author who needs the policy ordered after the
  protected workload must say so on `metadata.relationships`
  (`kind: KubernetesDeployment ... type: depends_on`). See the component README's
  "Composing in Infra Charts" section.

## 5. IaC implementation

Both engines are feature-equal and emit the same `security.istio.io/v1`
`PeerAuthentication`:

- **Pulumi** uses the typed crd2pulumi resource
  `istiosecurityv1.NewPeerAuthentication`, so field-name/structure errors
  are caught at compile time. `mtls`, `selector`, and `port_level_mtls` are only
  attached when present. The proto's `uint32` `port_level_mtls` keys are converted
  to decimal strings because the SDK models the field as `pulumi.StringMapMap`
  (`{"8080": {"mode": "STRICT"}}`).
- **Terraform** uses `kubernetes_manifest` with a fully-typed `variable "spec"` and
  a null-pruned `locals.tf` (conditional `merge`), so unset optional blocks are
  omitted from the manifest and upstream inheritance flows through. `port_level_mtls`
  is typed as a string-keyed `map(object({ mode = string }))`, matching the CRD's
  string-keyed JSON form.

## 6. E2E

- **Prerequisite**: `KubernetesIstioBaseCrds` (kind 868), declared in the registry
  (`prerequisites: [KubernetesIstioBaseCrds]`). The harness installs the istio/base
  CRDs before the scenario applies -- no full istiod is needed to prove the CR is
  accepted server-side and cleaned up on destroy.
- **Tier 4**, validated on both Pulumi and Terraform. Verification asserts the
  PeerAuthentication CR exists after apply and is gone after destroy
  (`peerauthentications.security.istio.io`), via the `ResourceExistenceVerifier`
  dispatched from `aa_e2e/verify/verifier.go`.

## 7. Future experience surface (kept open, not built here)

The spec is intentionally lossless so future Planton console experiences stay
possible without a schema change: a mesh-security posture panel that reads the
effective mode per namespace/workload, "promote PERMISSIVE -> STRICT" migration
flows, and port-level exception auditing all derive from these exact fields. The
console integration itself is a separate follow-up project.

## References

- [Istio PeerAuthentication reference](https://istio.io/latest/docs/reference/config/security/peer_authentication/)
- [Istio mutual TLS migration](https://istio.io/latest/docs/tasks/security/authentication/mtls-migration/)
- Upstream proto: `istio.io/api` `security/v1beta1/peer_authentication.proto` @ `1.26.8`

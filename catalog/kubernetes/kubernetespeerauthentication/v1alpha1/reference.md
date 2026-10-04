# KubernetesPeerAuthentication

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

KubernetesPeerAuthenticationSpec defines an Istio PeerAuthentication: a
namespaced policy that controls the mutual TLS (mTLS) requirements for
incoming connections to the workloads it selects.

100% fidelity with the upstream istio.io/api PeerAuthentication
(security/v1beta1/peer_authentication.proto, served as security.istio.io/v1),
pinned to the 1.30 line (tag 1.30.3). Upstream spec fields are flattened
directly after the Planton namespaced envelope (namespace);
there is no nested `peer_authentication` sub-message.

Scope semantics (upstream): if `selector` is omitted the policy applies to all
workloads in its namespace; in the mesh root namespace an empty selector makes
it a mesh-wide default. `port_level_mtls` only takes effect when a selector is
present (enforced below).

## Example

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesPeerAuthentication
metadata:
  name: test-peer-authentication
spec:
  namespace:
    value: test-namespace
  selector:
    match_labels:
      app: finance
  mtls:
    mode: STRICT
  port_level_mtls:
    "8080":
      mode: DISABLE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.selector` | `KubernetesIstioApiWorkloadSelector` |  |  |  |
| `spec.selector.matchLabels` | `map<string, string>` |  |  |  |
| `spec.mtls` | `KubernetesPeerAuthenticationMutualTls` |  |  |  |
| `spec.mtls.mode` | `string` | yes |  |  |
| `spec.portLevelMtls` | `map<uint32, KubernetesPeerAuthenticationMutualTls>` |  |  |  |
| `spec.portLevelMtls.*.mode` | `string` | yes |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace in which the PeerAuthentication is created. The policy's scope is
this namespace (or mesh-wide if this is the Istio root namespace and no
selector is set).

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.selector

`KubernetesIstioApiWorkloadSelector`

Selects the workloads (by pod/VM label) this policy applies to. When omitted,
the policy applies to every workload in its namespace (or, in the root
namespace, the whole mesh).

INFRA-CHART COMPOSABILITY: selector is a PLAIN label match, not an
Planton foreign key (StringValueOrRef). It is matched at runtime by istiod
against pod labels and creates NO automatic DAG edge to any workload
resource. To order this policy after the workload it protects in an infra
chart, an author MUST express the dependency via metadata.relationships, e.g.:
  metadata:
    relationships:
      - kind: KubernetesDeployment
        name: "{{ values.app }}"
        type: depends_on
See the kind's "Composing in Infra Charts" docs for the full pattern.

### spec.selector.matchLabels

`map<string, string>`

One or more labels indicating the set of pods/VMs the policy applies to.
Faithful to istio.io/api `istio.type.v1beta1.WorkloadSelector.match_labels`,
whose upstream CRD constraints are: max 4096 entries; each value <= 63 chars;
label keys must be non-empty; and wildcards ('*') are not permitted in keys or
values. The size/length bounds are expressed via the standard `map` rule; the
non-empty-key and no-wildcard constraints map to upstream's CEL XValidation
rules and are expressed here as field-level CEL.

- rule: label selector keys must not be empty
- rule: wildcard ('*') is not allowed in label selector keys
- rule: wildcard ('*') is not allowed in label selector values
- rule: {"map":{"maxPairs":"4096","values":{"string":{"maxLen":"63"}}}}

### spec.mtls

`KubernetesPeerAuthenticationMutualTls`

Mesh-TLS mode applied to the selected workloads. When omitted, the mode is
inherited from the parent (namespace-level, then mesh-level) policy. To make
inheritance explicit, set mode to UNSET rather than omitting the block.

### spec.mtls.mode

`string` · required

mTLS mode. One of:
  UNSET      — inherit from the parent policy; if none, treated as PERMISSIVE.
  DISABLE    — connection is not tunneled (plaintext only).
  PERMISSIVE — connection may be either plaintext or an mTLS tunnel.
  STRICT     — connection must be an mTLS tunnel (client cert required).
UNSET/DISABLE/PERMISSIVE/STRICT // external standard exception -- Istio MutualTLS.Mode enum

Modeled as a closed string set (not a proto enum). Unlike most
Planton enum-like fields, the empty string is NOT a valid wire value here:
MutualTLS's only field is `mode`, so an mtls block must carry a real mode.
Omitting the whole `mtls` block (or a port_level_mtls entry) is how
inheritance is expressed; UNSET is the explicit "inherit" value.

- rule: mtls mode must be one of UNSET, DISABLE, PERMISSIVE, or STRICT
- rule: {"required":true}

### spec.portLevelMtls

`map<uint32, KubernetesPeerAuthenticationMutualTls>`

Per-port mTLS overrides, keyed by the workload's port number (NOT the
Kubernetes Service port). These only take effect when `selector` is set, so a
non-empty map requires a non-empty selector (enforced below).

- rule: {"map":{"keys":{"uint32":{"lte":65535,"gte":1}}}}

### spec.portLevelMtls.*.mode

`string` · required

mTLS mode. One of:
  UNSET      — inherit from the parent policy; if none, treated as PERMISSIVE.
  DISABLE    — connection is not tunneled (plaintext only).
  PERMISSIVE — connection may be either plaintext or an mTLS tunnel.
  STRICT     — connection must be an mTLS tunnel (client cert required).
UNSET/DISABLE/PERMISSIVE/STRICT // external standard exception -- Istio MutualTLS.Mode enum

Modeled as a closed string set (not a proto enum). Unlike most
Planton enum-like fields, the empty string is NOT a valid wire value here:
MutualTLS's only field is `mode`, so an mtls block must carry a real mode.
Omitting the whole `mtls` block (or a port_level_mtls entry) is how
inheritance is expressed; UNSET is the explicit "inherit" value.

- rule: mtls mode must be one of UNSET, DISABLE, PERMISSIVE, or STRICT
- rule: {"required":true}

## Validation Rules

- `peer_authentication.port_level_mtls_requires_selector`: port_level_mtls requires a selector with at least one match label

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesPeerAuthentication, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.peer_authentication_name` | `string` | Name of the created PeerAuthentication (equals metadata.name). |
| `status.outputs.namespace` | `string` | Namespace the PeerAuthentication was created in (the resolved spec.namespace). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |

## See Also

- [Overview](../README.md)

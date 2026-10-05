# GcpSslPolicy

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpSslPolicySpec defines a Compute Engine SSL policy — the control for
which TLS versions and cipher suites a load balancer accepts from clients.
Without one, GCP's default policy applies: minimum TLS 1.0 with the
permissive COMPATIBLE cipher set. Attach a policy to a target HTTPS (or
SSL) proxy to enforce modern TLS for compliance regimes such as PCI DSS.

One kind covers both scopes. Leave `region` empty for a GLOBAL SSL policy
(used by global external load balancer proxies); set it for a REGIONAL
policy (used by regional external and internal ALB proxies). The two
scopes expose an identical configuration surface in GCP.

A policy is shared configuration: many proxies can reference one policy,
and profile / min_tls_version / custom_features all update in place — so
tightening a fleet's TLS floor is a single-resource change. Only name,
project, and description are immutable.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpSslPolicy
metadata:
  name: my-sample-ssl-policy
spec:
  # GCP project that owns the SSL policy.
  # Omit to use the provider's default project.
  projectId:
    value: my-gcp-project-123

  # Cloud-side name; omit to default to metadata.name.
  sslPolicyName: prod-tls-floor

  # Why this policy exists — shown in the GCP console.
  description: TLS 1.2 floor with MODERN ciphers for production frontends

  # Cipher-suite profile: COMPATIBLE (default), MODERN, RESTRICTED, CUSTOM,
  # or FIPS_202205 (requires minTlsVersion TLS_1_2).
  profile: MODERN

  # Minimum TLS version clients may negotiate: TLS_1_0 (default), TLS_1_1,
  # TLS_1_2, or TLS_1_3 (requires the RESTRICTED profile).
  minTlsVersion: TLS_1_2

  # Post-quantum key exchange rollout stance: DEFAULT (follow GCP's
  # timeline), ENABLED, or DEFERRED.
  postQuantumKeyExchange: ENABLED

  # What a destroy does: DELETE (default), PREVENT, or ABANDON.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.sslPolicyName` | `string` |  |  |  |
| `spec.region` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.profile` | `string` |  |  |  |
| `spec.minTlsVersion` | `string` |  |  |  |
| `spec.customFeatures` | `[]string` |  |  |  |
| `spec.postQuantumKeyExchange` | `string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project that owns the SSL policy.
Can be a literal project ID or a reference to a GcpProject resource.
If omitted, the provider's default project is used.
Immutable: changing it destroys and recreates the policy.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.sslPolicyName

`string`

Name of the SSL policy in GCP. Must be 1-63 characters: lowercase
letters, digits, and hyphens; must start with a letter and end with a
letter or digit. If not specified, defaults to metadata.name.
Immutable: changing it destroys and recreates the policy, briefly
breaking every proxy that references the old self_link.

- rule: ssl_policy_name must be RFC1035-compliant: 1-63 lowercase letters, digits, or hyphens; must start with a letter and end with a letter or digit

### spec.region

`string`

Region for a REGIONAL SSL policy (e.g. "us-central1"), used by regional
external and internal Application Load Balancer proxies. Leave empty for
a GLOBAL policy — the right scope for global external load balancers.
Immutable: a policy cannot move between scopes or regions.

- rule: region must be a valid GCP region name such as us-central1, or empty for a global SSL policy

### spec.description

`string`

Why this policy exists and which proxies should use it — write it for
the operator auditing TLS posture later. Immutable: changing it
destroys and recreates the policy (unusual for a description — a GCP
API quirk on this resource).

- rule: {"string":{"maxLen":"2048"}}

### spec.profile

`string`

The cipher-suite profile negotiated with clients (default COMPATIBLE).
COMPATIBLE allows the widest client range; MODERN drops broken ciphers
while keeping broad reach; RESTRICTED narrows to ciphers with modern
security guarantees (and is required when the TLS floor is raised beyond
what other profiles allow); CUSTOM hand-picks cipher suites via
custom_features; FIPS_202205 pins the FIPS 140-2/3 validated suite set
(and requires min_tls_version TLS_1_2 — the only floor that profile
supports). Mutable — tightening the profile applies to every proxy
referencing this policy on its next handshake.

- rule: profile must be COMPATIBLE, MODERN, RESTRICTED, CUSTOM, or FIPS_202205

### spec.minTlsVersion

`string`

The minimum TLS protocol version clients may negotiate (default
TLS_1_0). Raise to TLS_1_2 for PCI DSS and most modern compliance
regimes; TLS_1_3 is the strictest floor and requires the RESTRICTED
profile. GCP has no maximum-version control — TLS 1.3 is always
negotiable when the client supports it, whatever the floor. Mutable.

- rule: min_tls_version must be TLS_1_0, TLS_1_1, TLS_1_2, or TLS_1_3

### spec.customFeatures

`[]string`

Exact cipher suites to allow — required with (and only valid with) the
CUSTOM profile. Names are IANA-style suite identifiers from GCP's
supported set (e.g. TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256); GCP rejects
unknown names at deploy time. TLS 1.3 suites are not listable — GCP
always enables them regardless of this list. Mutable.

- rule: {"repeated":{"items":{"string":{"maxLen":"128","pattern":"^[A-Z0-9_]+$"}}}}

### spec.postQuantumKeyExchange

`string`

Post-quantum key exchange (X25519MLKEM768) posture for TLS handshakes
(default DEFAULT). GCP rolls the hybrid post-quantum group out on its
own schedule; this dial controls the rollout stance rather than a
static on/off:
  DEFAULT  -- follow GCP's rollout timeline
  ENABLED  -- allow post-quantum key exchange now
  DEFERRED -- opt out until GCP's later mandatory date
Mutable.

- rule: post_quantum_key_exchange must be DEFAULT, ENABLED, or DEFERRED

### spec.deletionPolicy

`string`

Deletion policy — what happens when this resource is destroyed:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the policy is deleted (GCP refuses while any proxy
               still references it, so destroy fails rather than
               silently loosening TLS floors)
  "PREVENT" -- destroy FAILS; protects a compliance-mandated TLS
               posture from accidental teardown
  "ABANDON" -- the policy is removed from management but left in
               GCP, still enforced by every proxy referencing it

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `custom_profile_requires_features`: the CUSTOM profile requires at least one cipher suite in custom_features
- `features_require_custom_profile`: custom_features is only valid with the CUSTOM profile — GCP rejects it on COMPATIBLE, MODERN, and RESTRICTED
- `fips_profile_requires_tls12`: the FIPS_202205 profile requires min_tls_version TLS_1_2 — the only floor GCP supports on that profile
- `tls13_floor_requires_restricted`: min_tls_version TLS_1_3 requires the RESTRICTED profile — GCP rejects a TLS 1.3 floor on any other profile

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpSslPolicy, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.self_link` | `string` | Self-link URI of the SSL policy. This is the value a target HTTPS (or SSL) proxy references in its ssl_policy field — the composition handle that hardens TLS at the load balancer. Global: https://www.googleapis.com/compute/v1/projects/{project}/global/sslPolicies/{name} Regional: https://www.googleapis.com/compute/v1/projects/{project}/regions/{region}/sslPolicies/{name} |
| `status.outputs.ssl_policy_name` | `string` | Name of the SSL policy as it exists in GCP. |
| `status.outputs.enabled_features` | `[]string` | The cipher suites the policy actually enables, as computed by GCP from the profile (or copied from custom_features on CUSTOM) — the list a compliance auditor asks for. |
| `status.outputs.region` | `string` | Region of a regional SSL policy; empty for a global one. Downstream composition can use this to confirm scope compatibility (regional proxies require an SSL policy in their own region). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpTargetHttpsProxy | `spec.sslPolicy` | `status.outputs.self_link` |

## See Also

- [Overview](../README.md)

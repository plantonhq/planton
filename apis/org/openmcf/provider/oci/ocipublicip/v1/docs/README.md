# OCI Public IP — Design & Internals

## Design Rationale

### Single-resource component

OciPublicIp manages exactly one `oci_core_public_ip` resource. Unlike OciVcn which bundles a VCN with optional gateways, a public IP has no child resources — there is a one-to-one mapping between the manifest and the OCI API object. This keeps the component minimal and composable.

### Two lifetime modes in one component

OCI public IPs have two fundamentally different lifecycle behaviors — reserved and ephemeral — but they share the same API surface (compartment, display name, private IP assignment, tagging). Rather than splitting into `OciReservedPublicIp` and `OciEphemeralPublicIp`, a single component with a `lifetime` field keeps the catalog smaller and avoids duplicating identical configuration logic.

The trade-off is a conditional validation rule: ephemeral IPs require `privateIpId`, while reserved IPs do not. This is enforced via a CEL expression (`ephemeral_requires_private_ip`) in the proto spec rather than in the Pulumi module, so invalid manifests are rejected before any infrastructure is touched.

### StringValueOrRef for all OCID fields

`compartmentId`, `privateIpId`, and `publicIpPoolId` all use the `StringValueOrRef` type. This allows:

- **Literal values** (`value: "ocid1.compartment.oc1..example"`) for standalone deployments or when referencing pre-existing resources.
- **Cross-resource references** (`valueFrom: { kind: OciCompartment, name: ..., fieldPath: ... }`) for dependency-aware multi-resource deployments.

The Pulumi module calls `.GetValue()` on each field, which resolves the reference at deployment time. The foreign key default for `compartmentId` points to `OciCompartment` / `status.outputs.compartmentId`, so `valueFrom` references to OciCompartment only need `kind` and `name`.

### Display name fallback

The `locals.go` file sets `DisplayName` to `spec.DisplayName` when provided, otherwise falls back to `metadata.Name`. This avoids requiring users to set both fields while still allowing explicit control when the OCI Console name should differ from the OpenMCF resource name.

## Tagging Strategy

Freeform tags are built in `locals.go` from a combination of fixed keys and metadata:

| Tag Key | Source | Always Present |
|---------|--------|---------------|
| `resource` | hardcoded `"true"` | yes |
| `resource_kind` | `CloudResourceKind_OciPublicIp` enum | yes |
| `resource_id` | `metadata.Id` | yes |
| `organization` | `metadata.Org` | only if set |
| `environment` | `metadata.Env` | only if set |
| *(custom)* | `metadata.Labels` | only if present |

Custom labels from `metadata.Labels` are merged last, so they can override any of the fixed keys if needed. This matches the tagging pattern used by other OCI components in the catalog.

## Immutable Fields

Two fields cannot be changed after creation:

1. **`lifetime`** — OCI does not support converting between reserved and ephemeral. Changing this in the manifest will trigger a destroy-and-recreate.
2. **`publicIpPoolId`** — the IP pool association is set at creation time. Changing pools requires a new public IP.

The Pulumi provider handles these as ForceNew fields, so `pulumi up` will show a replacement plan rather than silently failing.

## Comparison with Other Providers

| Aspect | OCI (this component) | AWS Elastic IP | GCP External Address |
|--------|---------------------|----------------|---------------------|
| Lifetime modes | `RESERVED` / `EPHEMERAL` in one resource | Elastic IPs are always reserved; ephemeral IPs are implicit on instance launch | `EXTERNAL` type with `RESERVE` subcommand; ephemeral is default on instance |
| BYOIP | `publicIpPoolId` field | `PublicIpv4Pool` parameter | `address` field with imported range |
| Assignment | `privateIpId` → private IP on VNIC | `AllocationId` → ENI or instance | `address` → instance or forwarding rule |
| Scope | Region (reserved) or AD (ephemeral) | Region | Region or global |

The main OCI-specific consideration is the explicit ephemeral lifetime mode. In AWS and GCP, ephemeral IPs are an implicit side effect of launching an instance with a public interface. In OCI, ephemeral IPs can be explicitly created and managed, which is why this component supports both modes.

## What's Deferred

The following capabilities are not included in v1 and may be added in future versions:

- **IPv6 public IPs** — OCI supports IPv6 addresses on VNICs, but the public IPv6 lifecycle differs from IPv4. This component only handles IPv4.
- **Bulk allocation** — creating multiple public IPs from a single manifest. Each IP requires its own OciPublicIp resource.
- **Automatic private IP discovery** — looking up a private IP by instance name or VNIC display name rather than requiring the OCID. This would require cross-resource queries not yet supported by the foreign key system.
- **IP address pinning** — requesting a specific IPv4 address from a BYOIP pool. OCI allocates the next available address from the pool; there is no API to request a particular address.

## Stack Output Keys

The Pulumi module exports two outputs defined as constants in `outputs.go`:

| Constant | Export Key | Proto Field | Description |
|----------|-----------|-------------|-------------|
| `OpPublicIpId` | `public_ip_id` | `publicIpId` | OCID of the created public IP |
| `OpIpAddress` | `ip_address` | `ipAddress` | Allocated IPv4 address string |

These are consumed by `OciPublicIpStackOutputs` and surfaced in `status.outputs` on the manifest after deployment.

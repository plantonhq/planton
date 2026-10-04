# GcpGlobalAddress

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpGlobalAddressSpec defines the configuration for a GCP global address reservation.

A global address reserves a static IP address at global scope. Two primary use cases:

1. External static IP — reserve a public IPv4 or IPv6 address for HTTP(S) load balancers,
   Cloud CDN, or global forwarding rules.

2. Internal IP range — reserve a private CIDR range for VPC Peering (used by Cloud SQL,
   Redis, AlloyDB for private networking) or Private Service Connect.

All fields except labels are ForceNew in the underlying GCP API — any change to the
configuration destroys and recreates the address.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpGlobalAddress
metadata:
  name: my-sample-global-address
spec:
  # GCP project that owns the reservation.
  # Omit to use the provider's default project.
  projectId:
    value: my-gcp-project-123

  # Cloud-side name of the reservation (RFC1035).
  addressName: lb-external-ip

  # Public static IP for a global load balancer frontend.
  addressType: EXTERNAL
  ipVersion: IPV4

  description: Static IP for the production HTTPS load balancer

  # User labels, merged with the platform labels (the one mutable surface).
  labels:
    team: platform

  # Ephemeral test fixture: release the reservation on destroy.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.addressName` | `string` | yes |  |  |
| `spec.address` | `string` |  |  |  |
| `spec.addressType` | `string` |  | `EXTERNAL` |  |
| `spec.description` | `string` |  |  |  |
| `spec.ipVersion` | `string` |  | `IPV4` |  |
| `spec.network` | `string \| valueFrom` |  |  | GcpVpcNetwork (`status.outputs.network_self_link`) |
| `spec.prefixLength` | `int32` |  |  |  |
| `spec.purpose` | `string` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project in which to create this global address reservation.
Can be a literal project ID or a reference to a GcpProject resource.
If omitted, the provider's default project is used.
Example: "my-prod-project-123"

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.addressName

`string` · required

Name of the global address resource in GCP.
Must be 1-63 characters, lowercase letters, numbers, or hyphens.
Must start with a lowercase letter and end with a letter or number.
Example: "lb-external-ip", "vpc-peering-range"

- rule: {"required":true,"string":{"pattern":"^[a-z]([a-z0-9-]{0,61}[a-z0-9])?$"}}

### spec.address

`string`

The static IP address to reserve. If omitted, GCP assigns an address automatically.
For EXTERNAL addresses this is a single IP. For INTERNAL VPC_PEERING addresses
this is the start of the reserved CIDR range.

### spec.addressType

`string` · optional (explicit presence)

The type of address to reserve.
EXTERNAL reserves a public IP address (default). INTERNAL reserves a private IP range
within a VPC network for purposes like VPC peering or Private Service Connect.

- default: `EXTERNAL`
- rule: address_type must be EXTERNAL or INTERNAL

### spec.description

`string`

Human-readable description of this address reservation.
Example: "Static IP for production HTTPS load balancer"

### spec.ipVersion

`string` · optional (explicit presence)

The IP version for this address. Defaults to IPV4.

- default: `IPV4`
- rule: ip_version must be IPV4 or IPV6

### spec.network

`string | valueFrom`

The VPC network for internal address reservations.
Required when address_type is INTERNAL. Accepts a network name or full self-link URL.
The reserved IP range must be in RFC1918 space and the network cannot be deleted
while reserved IP ranges refer to it.

- references: GcpVpcNetwork (`status.outputs.network_self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_self_link}} -- a bare string does not parse

### spec.prefixLength

`int32` · optional (explicit presence)

The prefix length of the IP range to reserve.
Required for VPC_PEERING purpose (e.g., 20 reserves a /20 range).
Not applicable for single IP reservations or PRIVATE_SERVICE_CONNECT addresses.
Valid range: 8 to 29.

- rule: {"int32":{"lte":29,"gte":8}}

### spec.purpose

`string`

The purpose of this address reservation. Only applicable for INTERNAL addresses,
and REQUIRED for them — the GCP API rejects an internal global reservation
without a purpose ("The field must be specified for reserving internal IP
Addresses").
VPC_PEERING — reserves a CIDR range for VPC network peering. Used by managed services
like Cloud SQL, Redis, AlloyDB, and Filestore for private networking.
PRIVATE_SERVICE_CONNECT — reserves an address for a Private Service Connect endpoint.
Leave empty for standard external address reservations.

- rule: purpose must be empty, VPC_PEERING, or PRIVATE_SERVICE_CONNECT

### spec.labels

`map<string, string>`

User labels attached to the reserved global address, merged with
Planton's platform labels (which win on key conflicts). The one mutable
surface on this resource — every other change destroys and re-reserves.

### spec.deletionPolicy

`string`

Deletion policy for the reserved global address — what happens on
destroy:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the reservation is released (GCP refuses while a global
               forwarding rule or PSA range still uses it); a released
               external anycast IP is gone for good
  "PREVENT" -- destroy FAILS; protects an IP that external DNS and
               client allow-lists may still point at
  "ABANDON" -- the reservation is removed from management but stays
               reserved in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `purpose_requires_internal`: purpose can only be set when address_type is INTERNAL
- `vpc_peering_requires_prefix_length`: prefix_length is required when purpose is VPC_PEERING
- `internal_requires_network`: network is required when address_type is INTERNAL
- `internal_requires_purpose`: purpose is required when address_type is INTERNAL (GCP rejects a purpose-less internal global reservation)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpGlobalAddress, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.address` | `string` | The reserved IP address or the start of the reserved CIDR range. For EXTERNAL addresses this is a public IP. For INTERNAL VPC_PEERING addresses this is the first address in the reserved range. |
| `status.outputs.self_link` | `string` | The self-link URL of the global address resource. Used to reference this address in other GCP resources like forwarding rules. |
| `status.outputs.creation_timestamp` | `string` | Creation timestamp in RFC3339 format. |
| `status.outputs.name` | `string` | Name of the global address resource in GCP (e.g. "vpc-peering-range"). Referenced by GcpServiceNetworkingConnection.spec.reserved_peering_ranges when composing private services access. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.network` | GcpVpcNetwork | `status.outputs.network_self_link` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpGlobalForwardingRule | `spec.ipAddress` | `status.outputs.address` |
| GcpServiceNetworkingConnection | `spec.reservedPeeringRanges` | `status.outputs.name` |
| GcpVertexAiDeployedIndex | `spec.reservedIpRanges` | `status.outputs.name` |

## See Also

- [Overview](../README.md)

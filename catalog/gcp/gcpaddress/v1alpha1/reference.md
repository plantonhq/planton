# GcpAddress

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpAddressSpec defines a regional Compute Engine address reservation
(`google_compute_address`) — a static IP at regional scope, distinct from
GcpGlobalAddress which reserves global-scope addresses.

Two primary use cases:

1. External static IP — reserve a public IPv4 or IPv6 address in a region
   for Cloud NAT, regional load balancers, or VM instances.

2. Internal regional IP — reserve a private address or CIDR range within a
   VPC network or subnetwork for GCE endpoints, internal load balancer
   VIPs, VPC peering ranges, IPsec interconnect, or DNS resolver endpoints.

All fields except labels are ForceNew in the underlying GCP API — any change
to the configuration destroys and recreates the address.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpAddress
metadata:
  name: my-sample-regional-address
spec:
  # GCP project that owns the reservation.
  # Omit to use the provider's default project.
  projectId:
    value: my-gcp-project-123

  # Cloud-side name of the reservation (RFC1035).
  addressName: nat-external-ip

  # Region for the reservation — required for regional addresses.
  region: us-central1

  # Public static IP for Cloud NAT or a regional load balancer frontend.
  addressType: EXTERNAL
  ipVersion: IPV4

  description: Static external IP for Cloud NAT in us-central1

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
| `spec.region` | `string` | yes |  |  |
| `spec.address` | `string` |  |  |  |
| `spec.addressType` | `string` |  | `EXTERNAL` |  |
| `spec.description` | `string` |  |  |  |
| `spec.ipVersion` | `string` |  | `IPV4` |  |
| `spec.network` | `string \| valueFrom` |  |  | GcpVpcNetwork (`status.outputs.network_self_link`) |
| `spec.subnetwork` | `string \| valueFrom` |  |  | GcpSubnetwork (`status.outputs.subnetwork_self_link`) |
| `spec.networkTier` | `string` |  |  |  |
| `spec.prefixLength` | `int32` |  |  |  |
| `spec.purpose` | `string` |  |  |  |
| `spec.ipv6EndpointType` | `string` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.ipCollection` | `string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project in which to create this regional address reservation.
Can be a literal project ID or a reference to a GcpProject resource.
If omitted, the provider's default project is used.
Example: "my-prod-project-123"

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.addressName

`string` · required

Name of the regional address resource in GCP.
Must be 1-63 characters, lowercase letters, numbers, or hyphens.
Must start with a lowercase letter and end with a letter or number.
Example: "nat-external-ip", "ilb-vip"

- rule: {"required":true,"string":{"pattern":"^[a-z]([a-z0-9-]{0,61}[a-z0-9])?$"}}

### spec.region

`string` · required

The GCP region in which to reserve this address (e.g. "us-central1").
Immutable: a regional address cannot move between regions.

- rule: region must be a valid GCP region name such as us-central1
- rule: {"required":true}

### spec.address

`string`

The static IP address to reserve. If omitted, GCP assigns an address
automatically. For INTERNAL VPC_PEERING addresses this is the start of
the reserved CIDR range.

### spec.addressType

`string` · optional (explicit presence)

The type of address to reserve.
EXTERNAL reserves a public IP address (default). INTERNAL reserves a
private IP within a VPC network or subnetwork.

- default: `EXTERNAL`
- rule: address_type must be EXTERNAL or INTERNAL

### spec.description

`string`

Human-readable description of this address reservation.
Example: "Static IP for Cloud NAT in us-central1"

### spec.ipVersion

`string` · optional (explicit presence)

The IP version for this address. Defaults to IPV4.

- default: `IPV4`
- rule: ip_version must be IPV4 or IPV6

### spec.network

`string | valueFrom`

The VPC network for internal address reservations with VPC_PEERING or
IPSEC_INTERCONNECT purpose. Accepts a network name or full self-link URL.

- references: GcpVpcNetwork (`status.outputs.network_self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_self_link}} -- a bare string does not parse

### spec.subnetwork

`string | valueFrom`

The subnetwork for internal address reservations with GCE_ENDPOINT or
DNS_RESOLVER purpose. Accepts a subnetwork name or full self-link URL.

- references: GcpSubnetwork (`status.outputs.subnetwork_self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSubnetwork, name: <that resource's name>, fieldPath: status.outputs.subnetwork_self_link}} -- a bare string does not parse

### spec.networkTier

`string`

The network tier for EXTERNAL addresses only. PREMIUM (default) or
STANDARD. Must not be set for INTERNAL addresses — internal traffic
always uses Premium tier.

- rule: network_tier must be empty, PREMIUM, or STANDARD

### spec.prefixLength

`int32` · optional (explicit presence)

The prefix length of the IP range to reserve (8-29).
Used for INTERNAL VPC_PEERING or IPSEC_INTERCONNECT ranges.

- rule: {"int32":{"lte":29,"gte":8}}

### spec.purpose

`string`

The purpose of this INTERNAL address reservation.
GCE_ENDPOINT — VM instances, alias IP ranges, or similar endpoints.
SHARED_LOADBALANCER_VIP — internal load balancer VIP shared across
backends.
VPC_PEERING — reserves a CIDR range for VPC network peering.
IPSEC_INTERCONNECT — reserves a range for HA VPN over Cloud Interconnect.
DNS_RESOLVER — DNS resolver endpoint address.
Leave empty for standard EXTERNAL address reservations.
PRIVATE_SERVICE_CONNECT is global-only — use GcpGlobalAddress instead.

- rule: purpose must be empty, GCE_ENDPOINT, SHARED_LOADBALANCER_VIP, VPC_PEERING, IPSEC_INTERCONNECT, or DNS_RESOLVER

### spec.ipv6EndpointType

`string`

The endpoint type for external IPv6 addresses: VM or NETLB.
Determines whether the reserved IPv6 address can be used by a VM
instance or a network load balancer after reservation.

- rule: ipv6_endpoint_type must be empty, VM, or NETLB

### spec.labels

`map<string, string>`

User labels attached to the reserved address, merged with Planton's
platform labels (which win on key conflicts). The one mutable surface on
this resource — every other change destroys and re-reserves the address.

### spec.ipCollection

`string`

Source of externally provisioned (BYOIP) addresses: a
PublicDelegatedPrefix in full or partial URL form, e.g.
"projects/{project}/regions/{region}/publicDelegatedPrefixes/{pdp-name}".
An IPv4 PDP must support enhanced IPv4 allocations; an IPv6 PDP must be
in EXTERNAL_IPV6_FORWARDING_RULE_CREATION mode. Only meaningful for
EXTERNAL addresses. Create-time only: changing it re-reserves.

- rule: {"string":{"maxLen":"2048"}}

### spec.deletionPolicy

`string`

Deletion policy for the reserved address — what happens on destroy:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the reservation is released; the IP returns to Google's
               pool (an external static IP is gone for good)
  "PREVENT" -- destroy FAILS; protects an IP that DNS records or
               allow-lists outside GCP may still point at
  "ABANDON" -- the reservation is removed from management but stays
               reserved (and billed while unattached) in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `internal_cannot_have_network_tier`: network_tier cannot be set when address_type is INTERNAL
- `purpose_requires_internal`: purpose can only be set when address_type is INTERNAL
- `peering_purpose_requires_network`: network is required when purpose is VPC_PEERING or IPSEC_INTERCONNECT
- `endpoint_purpose_requires_subnetwork`: subnetwork is required when purpose is GCE_ENDPOINT or DNS_RESOLVER
- `shared_lb_vip_requires_internal`: SHARED_LOADBALANCER_VIP purpose requires address_type INTERNAL

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpAddress, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.address` | `string` | The reserved IP address or the start of the reserved CIDR range. For EXTERNAL addresses this is a public IP. For INTERNAL addresses this is the reserved private IP or range start. |
| `status.outputs.self_link` | `string` | The self-link URL of the regional address resource. Used to reference this address in other GCP resources like forwarding rules, NAT gateways, or VM network interfaces. |
| `status.outputs.name` | `string` | Name of the regional address resource in GCP (e.g. "nat-external-ip"). |
| `status.outputs.region` | `string` | The region from the spec (e.g. "us-central1") — the plain region name, not a provider self-link, so downstream composition can confirm scope compatibility without parsing URLs. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.network` | GcpVpcNetwork | `status.outputs.network_self_link` |
| `spec.subnetwork` | GcpSubnetwork | `status.outputs.subnetwork_self_link` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpComputeInstance | `spec.networkInterfaces[].networkIp` | `status.outputs.address` |
| GcpComputeInstance | `spec.networkInterfaces[].accessConfigs[].natIp` | `status.outputs.address` |
| GcpComputeMig | `spec.perInstanceConfigs[].preservedState.externalIps[].address` | `status.outputs.address` |
| GcpComputeMig | `spec.perInstanceConfigs[].preservedState.internalIps[].address` | `status.outputs.address` |
| GcpDnsRecord | `spec.routingPolicy.wrr[].healthCheckedTargets.internalLoadBalancers[].ipAddress` | `status.outputs.address` |
| GcpDnsRecord | `spec.routingPolicy.geo[].healthCheckedTargets.internalLoadBalancers[].ipAddress` | `status.outputs.address` |
| GcpDnsRecord | `spec.routingPolicy.primaryBackup.primary.internalLoadBalancers[].ipAddress` | `status.outputs.address` |
| GcpDnsRecord | `spec.routingPolicy.primaryBackup.backupGeo[].healthCheckedTargets.internalLoadBalancers[].ipAddress` | `status.outputs.address` |
| GcpRedisClusterEndpointSet | `spec.endpoints[].connections[].address` | `status.outputs.address` |
| GcpRouterNat | `spec.natIps` | `status.outputs.self_link` |
| GcpRouterNat | `spec.drainNatIps` | `status.outputs.self_link` |
| GcpRouterNat | `spec.rules[].action.sourceNatActiveIps` | `status.outputs.self_link` |
| GcpRouterNat | `spec.rules[].action.sourceNatDrainIps` | `status.outputs.self_link` |
| GcpVertexAiNotebook | `spec.networkInterface.externalIp` | `status.outputs.address` |

## See Also

- [Overview](../README.md)

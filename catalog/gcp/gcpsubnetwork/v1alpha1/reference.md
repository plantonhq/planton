# GcpSubnetwork

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpSubnetworkSpec defines a subnetwork in a custom-mode VPC — the regional
address space workloads actually live in. Subnets carry the IP plan:
a primary IPv4 range for VM interfaces, optional secondary ranges for
alias IPs (GKE pods and services), optional IPv6, and the VPC Flow Logs
configuration for the subnet's traffic.

Most subnets keep the default PRIVATE purpose. The purpose field also
creates the special-role subnets other networking features require:
proxy-only subnets that reserve address space for Envoy-based load
balancers (REGIONAL_MANAGED_PROXY / GLOBAL_MANAGED_PROXY) and Private
Service Connect subnets (PRIVATE_SERVICE_CONNECT) that back published
services.

Sizing note: GCP reserves 4 addresses per primary range, and a subnet's
primary range can be EXPANDED later without recreation — but never shrunk.
Start with room to grow (a /20 is a common team-sized default).

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpSubnetwork
metadata:
  name: my-sample-subnetwork
spec:
  # GCP project that owns the subnetwork.
  # Omit to use the provider's default project.
  projectId:
    value: my-gcp-project-123

  # Self-link of the parent VPC (or reference a GcpVpcNetwork resource).
  vpcSelfLink:
    value: projects/my-gcp-project-123/global/networks/my-vpc

  subnetworkName: app-subnet
  region: us-central1

  # Primary range: /20 leaves room to grow (expansion is in-place;
  # shrinking recreates).
  ipCidrRange: 10.10.0.0/20

  description: Primary workload subnet for the app tier

  # Secondary ranges for GKE alias IPs.
  secondaryIpRanges:
    - rangeName: pods
      ipCidrRange: 10.16.0.0/14
    - rangeName: services
      ipCidrRange: 10.20.0.0/20

  # Private-only workloads still reach Google APIs internally.
  privateIpGoogleAccess: true

  # VPC Flow Logs at half sampling with all metadata.
  logConfig:
    aggregationInterval: INTERVAL_5_SEC
    flowSampling: 0.5
    metadata: INCLUDE_ALL_METADATA

  # Resource Manager tags for org-policy conditions and IAM scoping
  # (tagKeys/tagValues IDs, not short names). Changing them REPLACES the
  # subnetwork — plan tag changes deliberately.
  resourceManagerTags:
    tagKeys/281475123456789: tagValues/281476987654321

  # Delete the subnet on destroy (GCP's default, made explicit). GCP rejects
  # the delete while any VM, connector, or proxy still holds an address here.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.vpcSelfLink` | `string \| valueFrom` | yes |  | GcpVpcNetwork (`status.outputs.network_self_link`) |
| `spec.subnetworkName` | `string` | yes |  |  |
| `spec.region` | `string` | yes |  |  |
| `spec.ipCidrRange` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.purpose` | `string` |  | `PRIVATE` |  |
| `spec.role` | `string` |  |  |  |
| `spec.secondaryIpRanges` | `[]GcpSubnetworkSecondaryRange` |  |  |  |
| `spec.secondaryIpRanges[].rangeName` | `string` | yes |  |  |
| `spec.secondaryIpRanges[].ipCidrRange` | `string` |  |  |  |
| `spec.secondaryIpRanges[].reservedInternalRange` | `string` |  |  |  |
| `spec.privateIpGoogleAccess` | `bool` |  |  |  |
| `spec.privateIpv6GoogleAccess` | `string` |  |  |  |
| `spec.stackType` | `string` |  | `IPV4_ONLY` |  |
| `spec.ipv6AccessType` | `string` |  |  |  |
| `spec.externalIpv6Prefix` | `string` |  |  |  |
| `spec.allowSubnetCidrRoutesOverlap` | `bool` |  |  |  |
| `spec.sendSecondaryIpRangeIfEmpty` | `bool` |  |  |  |
| `spec.logConfig` | `GcpSubnetworkLogConfig` |  |  |  |
| `spec.logConfig.aggregationInterval` | `string` |  | `INTERVAL_5_SEC` |  |
| `spec.logConfig.flowSampling` | `double` |  | `0.5` |  |
| `spec.logConfig.metadata` | `string` |  | `INCLUDE_ALL_METADATA` |  |
| `spec.logConfig.metadataFields` | `[]string` |  |  |  |
| `spec.logConfig.filterExpr` | `string` |  |  |  |
| `spec.logConfig.enabled` | `bool` |  | `true` |  |
| `spec.reservedInternalRange` | `string` |  |  |  |
| `spec.internalIpv6Prefix` | `string` |  |  |  |
| `spec.ipCollection` | `string` |  |  |  |
| `spec.resourceManagerTags` | `map<string, string>` |  |  |  |
| `spec.resolveSubnetMask` | `string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project that owns this subnetwork.
Can be a literal project ID or a reference to a GcpProject resource.
If omitted, the provider's default project is used.
Immutable: changing it destroys and recreates the subnetwork.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.vpcSelfLink

`string | valueFrom` · required

The parent VPC network, by self-link.
Reference a GcpVpcNetwork resource or provide the self-link directly.
Immutable: a subnet cannot move between networks.

- references: GcpVpcNetwork (`status.outputs.network_self_link`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_self_link}} -- a bare string does not parse

### spec.subnetworkName

`string` · required

Name of the subnetwork in GCP. Must be 1-63 characters: lowercase
letters, digits, and hyphens; must start with a letter and end with a
letter or digit.
Immutable: changing it destroys and recreates the subnetwork.

- rule: {"required":true,"string":{"pattern":"^[a-z]([a-z0-9-]{0,61}[a-z0-9])?$"}}

### spec.region

`string` · required

Region the subnetwork lives in (e.g. "us-central1"). Subnets are
regional: workloads in any zone of this region can use it.
Immutable: a subnet cannot move between regions.

- rule: {"required":true,"string":{"pattern":"^[a-z]([-a-z0-9]*[a-z0-9])?$"}}

### spec.ipCidrRange

`string`

Primary IPv4 CIDR range (e.g. "10.10.0.0/20"). Must not overlap any
other range in the VPC. Mutable in ONE direction: the range can be
expanded in place (e.g. /20 → /18), but shrinking it forces destroy and
recreate — plan for growth up front. Leave empty for IPV6_ONLY subnets
(which carry no IPv4 range) or when reserved_internal_range supplies
the CIDR from a Network Connectivity internal range.

- rule: ip_cidr_range must be an IPv4 CIDR like 10.10.0.0/20

### spec.description

`string`

What this subnet carries and which workloads should use it — write it
for the operator doing IP planning later.
Immutable: description updates force recreation on this resource.

- rule: {"string":{"maxLen":"2048"}}

### spec.purpose

`string` · optional (explicit presence)

What the subnet is FOR. PRIVATE (the default) is a regular workload
subnet. REGIONAL_MANAGED_PROXY reserves address space for the region's
Envoy-based load balancers (required before creating a regional internal
or regional external Application Load Balancer in the VPC);
GLOBAL_MANAGED_PROXY is the cross-region equivalent.
PRIVATE_SERVICE_CONNECT backs published PSC services. PEER_MIGRATION
stages subnet migration between peered VPCs. PRIVATE_NAT provides source
ranges for Private NAT gateways.
Immutable in practice: choose the purpose when the subnet is created.

- default: `PRIVATE`
- rule: purpose must be one of PRIVATE, REGIONAL_MANAGED_PROXY, GLOBAL_MANAGED_PROXY, PRIVATE_SERVICE_CONNECT, PEER_MIGRATION, or PRIVATE_NAT

### spec.role

`string`

For REGIONAL_MANAGED_PROXY subnets only: ACTIVE means the region's Envoy
proxies allocate addresses from this subnet now; BACKUP holds the subnet
ready for a drain-and-swap (promote a BACKUP to ACTIVE to migrate the
proxy fleet onto fresh address space). Mutable.

- rule: role must be ACTIVE or BACKUP

### spec.secondaryIpRanges

`[]GcpSubnetworkSecondaryRange`

Secondary IPv4 ranges for alias IPs — the mechanism GKE uses for pod
and service IPs (VPC-native clusters), and any workload that needs
per-container addresses. Up to 170 per subnet; names are how consumers
(e.g. a GKE cluster's ip_allocation_policy) select a range.

- rule: each secondary range needs exactly one CIDR source — ip_cidr_range or reserved_internal_range

### spec.secondaryIpRanges[].rangeName

`string` · required

Name of this secondary range, unique within the subnet — the handle
consumers use to select it (e.g. a GKE cluster's pods range). 1-63
characters, RFC1035. Renaming a range in place is not supported by GCP.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"63","pattern":"^[a-z]([-a-z0-9]*[a-z0-9])?$"}}

### spec.secondaryIpRanges[].ipCidrRange

`string`

IPv4 CIDR of this secondary range. Must not overlap any other range in
the VPC. Size for the consumer: GKE pod ranges are commonly /14-/18,
service ranges /20. Alternative to reserved_internal_range.

- rule: ip_cidr_range must be an IPv4 CIDR like 10.16.0.0/14

### spec.secondaryIpRanges[].reservedInternalRange

`string`

Source this secondary range's CIDR from a pre-planned Network
Connectivity internal range instead of a literal ip_cidr_range, e.g.
"networkconnectivity.googleapis.com/projects/{project}/locations/global/internalRanges/{name}".
Alternative to ip_cidr_range; unlike the subnet's primary range this
is mutable in place.

### spec.privateIpGoogleAccess

`bool`

Let VMs without external IPs reach Google APIs and services over the
subnet's internal addresses. Effectively mandatory for private-only
subnets whose workloads pull from GCR/Artifact Registry or call any
Google API. Mutable.

### spec.privateIpv6GoogleAccess

`string`

IPv6 counterpart of private_ip_google_access, controlling VM-to-Google
traffic over IPv6: DISABLE_GOOGLE_ACCESS,
ENABLE_OUTBOUND_VM_ACCESS_TO_GOOGLE, or
ENABLE_BIDIRECTIONAL_ACCESS_TO_GOOGLE. Leave unset to keep GCP's
default (disabled). Mutable.

- rule: private_ipv6_google_access must be DISABLE_GOOGLE_ACCESS, ENABLE_OUTBOUND_VM_ACCESS_TO_GOOGLE, or ENABLE_BIDIRECTIONAL_ACCESS_TO_GOOGLE

### spec.stackType

`string` · optional (explicit presence)

The subnet's IP stack: IPV4_ONLY (the default), IPV4_IPV6 (dual-stack —
VMs can carry both), or IPV6_ONLY. Dual-stack and IPv6-only subnets
also need ipv6_access_type. Mutable between IPV4_ONLY and IPV4_IPV6;
moving to or from IPV6_ONLY recreates.

- default: `IPV4_ONLY`
- rule: stack_type must be IPV4_ONLY, IPV4_IPV6, or IPV6_ONLY

### spec.ipv6AccessType

`string`

Where the subnet's IPv6 addresses are reachable from: EXTERNAL
(internet-routable GUAs assigned by Google) or INTERNAL (ULAs, only
routable inside the VPC — requires the VPC to have an internal IPv6
range enabled). Required for IPV4_IPV6 and IPV6_ONLY subnets.
Immutable: the access type cannot change after creation.

- rule: ipv6_access_type must be EXTERNAL or INTERNAL

### spec.externalIpv6Prefix

`string`

For EXTERNAL IPv6 subnets: pin a specific /64 external IPv6 prefix
(must belong to a range GCP can assign). Leave empty to let Google
allocate one — the normal path. Immutable.

### spec.allowSubnetCidrRoutesOverlap

`bool`

Permit this subnet's CIDR to overlap with routes to destinations
OUTSIDE the VPC (e.g. re-using RFC1918 space that a peer or on-prem
route also claims). Subnet routes still take precedence; use only when
deliberately reclaiming address space. Mutable.

### spec.sendSecondaryIpRangeIfEmpty

`bool`

Controls what an EMPTY secondary_ip_ranges list means on update: true
sends the empty list (removing all secondary ranges); false (the
default) omits it, leaving existing secondary ranges untouched. A
safety latch against wiping GKE pod ranges with a partial manifest.

### spec.logConfig

`GcpSubnetworkLogConfig`

VPC Flow Logs for this subnet: samples of TCP/UDP flows for network
monitoring, forensics, and cost analysis, delivered to Cloud Logging.
Unset means flow logs are OFF. Logging every flow at full sampling on a
busy subnet is expensive — tune flow_sampling and filter_expr.

- rule: metadata_fields only applies when metadata is CUSTOM_METADATA

### spec.logConfig.aggregationInterval

`string` · optional (explicit presence)

How long flows are aggregated before a log entry is emitted. Longer
intervals mean fewer, larger entries (lower cost, coarser timing).

- default: `INTERVAL_5_SEC`
- rule: aggregation_interval must be one of INTERVAL_5_SEC, INTERVAL_30_SEC, INTERVAL_1_MIN, INTERVAL_5_MIN, INTERVAL_10_MIN, or INTERVAL_15_MIN

### spec.logConfig.flowSampling

`double` · optional (explicit presence)

Fraction of flows to sample, 0.0-1.0 (GCP default 0.5). 1.0 captures
everything — the right choice for security forensics, at real Logging
cost on busy subnets; lower it for trend monitoring.

- default: `0.5`
- rule: {"double":{"lte":1,"gte":0}}

### spec.logConfig.metadata

`string` · optional (explicit presence)

Which metadata joins each log entry: INCLUDE_ALL_METADATA (the
default), EXCLUDE_ALL_METADATA (smallest entries), or CUSTOM_METADATA
(only the fields listed in metadata_fields).

- default: `INCLUDE_ALL_METADATA`
- rule: metadata must be EXCLUDE_ALL_METADATA, INCLUDE_ALL_METADATA, or CUSTOM_METADATA

### spec.logConfig.metadataFields

`[]string`

Metadata fields to include when metadata is CUSTOM_METADATA (e.g.
"src_instance", "dest_vpc").

### spec.logConfig.filterExpr

`string`

CEL expression selecting which flows are logged (GCP default "true" =
all sampled flows), e.g. restricting to one port:
connection.dest_port == 443.

- rule: {"string":{"maxLen":"2048"}}

### spec.logConfig.enabled

`bool` · optional (explicit presence)

Whether flow logs are on. Unset means on: declaring the block has always
meant enabling them, and this switch lets a manifest say the opposite
out loud.

- default: `true`

### spec.reservedInternalRange

`string`

Source the primary CIDR from a pre-planned Network Connectivity
internal range instead of typing a literal ip_cidr_range: the value is
the range's resource path prefixed with the API host, e.g.
"networkconnectivity.googleapis.com/projects/{project}/locations/global/internalRanges/{name}".
The range's CIDR becomes the subnet's primary range — the enterprise
path for centrally allocated IP plans. Immutable: changing it destroys
and recreates the subnetwork.

### spec.internalIpv6Prefix

`string`

For INTERNAL IPv6 subnets: pin a specific internal IPv6 prefix (ULA)
instead of letting Google allocate one from the VPC's internal IPv6
range. Immutable: the prefix cannot change after creation.

### spec.ipCollection

`string`

BYOIP for IPv6: resource path of a PublicDelegatedPrefix (a sub-PDP in
EXTERNAL_IPV6_SUBNETWORK_CREATION or INTERNAL_IPV6_SUBNETWORK_CREATION
mode) the subnet's IPv6 space is drawn from, e.g.
"projects/{project}/regions/{region}/publicDelegatedPrefixes/{name}".
Only meaningful on dual-stack or IPv6-only subnets. Mutable.

### spec.resourceManagerTags

`map<string, string>`

Resource Manager tags bound to the subnetwork at create time, for
org-policy conditions and IAM scoping. Keys are "tagKeys/{tag_key_id}"
and values "tagValues/{tag_value_id}" (IDs, not short names). Changing
tags REPLACES the subnetwork (the provider sends them through a
create-only params block) -- plan tag changes deliberately.

### spec.resolveSubnetMask

`string`

How VMs in this subnet resolve the subnet mask in ARP responses —
relevant for appliance/NFV subnets whose workloads inspect ARP.
ARP_ALL_RANGES answers for every range in the subnet;
ARP_PRIMARY_RANGE answers only for the primary range;
the ARP_BROADCAST_* modes answer with the broadcast-domain mask
(WITH_LEARNING additionally learns from observed traffic). Leave unset
for GCP's default behavior. Immutable: changing it recreates the
subnetwork.

- rule: resolve_subnet_mask must be one of ARP_ALL_RANGES, ARP_PRIMARY_RANGE, ARP_BROADCAST_PRIMARY_RANGE, or ARP_BROADCAST_PRIMARY_RANGE_WITH_LEARNING

### spec.deletionPolicy

`string`

What happens to the subnetwork in GCP when this resource is destroyed.
  "DELETE"  -- (GCP's default when unset) the subnet is deleted; GCP
               rejects the delete while any VM, connector, or proxy
               still holds an address in it
  "PREVENT" -- destroy FAILS; protects the address space workloads
               and peers may still route to
  "ABANDON" -- the subnet is removed from management but stays in GCP
               (free at rest; its CIDR stays claimed in the VPC)

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `ipv4_range_required_unless_ipv6_only`: set ip_cidr_range or reserved_internal_range (only IPV6_ONLY subnets omit both)
- `primary_range_single_source`: ip_cidr_range and reserved_internal_range are alternative sources for the primary range — set at most one
- `internal_ipv6_prefix_requires_internal_access`: internal_ipv6_prefix only applies to subnets with ipv6_access_type INTERNAL
- `ip_collection_requires_ipv6_stack`: ip_collection (BYOIP) only applies to IPV4_IPV6 or IPV6_ONLY subnets
- `role_requires_managed_proxy_purpose`: role is only valid on REGIONAL_MANAGED_PROXY subnets — remove it or set purpose accordingly
- `ipv6_access_requires_ipv6_stack`: ipv6_access_type is required for IPV4_IPV6 and IPV6_ONLY subnets, and must be omitted for IPV4_ONLY
- `external_ipv6_prefix_requires_external_access`: external_ipv6_prefix only applies to subnets with ipv6_access_type EXTERNAL

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpSubnetwork, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.subnetwork_self_link` | `string` | Self-link URI of the subnetwork — the value GKE clusters, compute instances, and other subnet consumers reference. |
| `status.outputs.region` | `string` | The region the subnetwork lives in. |
| `status.outputs.ip_cidr_range` | `string` | The primary IPv4 CIDR range (empty for IPV6_ONLY subnets). |
| `status.outputs.secondary_ranges` | `[]GcpSubnetworkSecondaryRangeStackOutput` | Secondary (alias) ranges on this subnet, with their names and CIDRs. GKE clusters select their pod/service ranges by range_name. |
| `status.outputs.secondary_ranges[].range_name` | `string` | Name of the secondary range (unique within the subnet). |
| `status.outputs.secondary_ranges[].ip_cidr_range` | `string` | IPv4 CIDR of the secondary range. |
| `status.outputs.subnetwork_name` | `string` | Name of the subnetwork as it exists in GCP. Referenced by consumers that address subnets by name (e.g. Cloud Run Direct VPC egress). |
| `status.outputs.gateway_address` | `string` | IPv4 address of the subnet's default gateway. |
| `status.outputs.subnetwork_id` | `string` | Server-assigned numeric ID of the subnetwork. |
| `status.outputs.internal_ipv6_prefix` | `string` | The internal IPv6 prefix allocated to the subnet (INTERNAL ipv6_access_type); empty otherwise. |
| `status.outputs.external_ipv6_prefix` | `string` | The external IPv6 prefix allocated to the subnet (EXTERNAL ipv6_access_type); empty otherwise. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.vpcSelfLink` | GcpVpcNetwork | `status.outputs.network_self_link` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpAddress | `spec.subnetwork` | `status.outputs.subnetwork_self_link` |
| GcpCloudComposerEnvironment | `spec.nodeConfig.subnetwork` | `status.outputs.subnetwork_self_link` |
| GcpCloudFunction | `spec.serviceConfig.directVpcNetworkInterface.subnetwork` | `status.outputs.subnetwork_name` |
| GcpCloudRun | `spec.vpcAccess.networkInterfaces[].subnetwork` | `status.outputs.subnetwork_name` |
| GcpCloudRunJob | `spec.template.vpcAccess.networkInterfaces[].subnetwork` | `status.outputs.subnetwork_name` |
| GcpCloudRunWorkerPool | `spec.vpcAccess.networkInterfaces[].subnetwork` | `status.outputs.subnetwork_name` |
| GcpColabRuntimeTemplate | `spec.networkSpec.subnetwork` | `status.outputs.subnetwork_self_link` |
| GcpColabSchedule | `spec.notebookExecutionJob.customEnvironmentSpec.networkSpec.subnetwork` | `status.outputs.subnetwork_self_link` |
| GcpComputeInstance | `spec.networkInterfaces[].subnetwork` | `status.outputs.subnetwork_self_link` |
| GcpComputeMig | `spec.template.networkInterfaces[].subnetwork` | `status.outputs.subnetwork_self_link` |
| GcpDataprocCluster | `spec.clusterConfig.gceConfig.subnetwork` | `status.outputs.subnetwork_self_link` |
| GcpGkeCluster | `spec.subnetwork` | `status.outputs.subnetwork_self_link` |
| GcpGkeCluster | `spec.ipAllocation.clusterSecondaryRangeName` | `status.outputs.secondary_ranges.[*].range_name` |
| GcpGkeCluster | `spec.ipAllocation.servicesSecondaryRangeName` | `status.outputs.secondary_ranges.[*].range_name` |
| GcpGkeCluster | `spec.ipAllocation.additionalIpRanges[].subnetwork` | `status.outputs.subnetwork_self_link` |
| GcpGkeCluster | `spec.privateCluster.privateEndpointSubnetwork` | `status.outputs.subnetwork_self_link` |
| GcpGkeNodePool | `spec.networkConfig.subnetwork` | `status.outputs.subnetwork_self_link` |
| GcpGkeNodePool | `spec.networkConfig.additionalNodeNetworks[].subnetwork` | `status.outputs.subnetwork_self_link` |
| GcpGkeNodePool | `spec.networkConfig.additionalPodNetworks[].subnetwork` | `status.outputs.subnetwork_self_link` |
| GcpGlobalForwardingRule | `spec.subnetwork` | `status.outputs.subnetwork_self_link` |
| GcpManagedKafkaCluster | `spec.networkConfigs[].subnet` | `status.outputs.subnetwork_self_link` |
| GcpManagedKafkaConnectCluster | `spec.networkConfigs[].primarySubnet` | `status.outputs.subnetwork_self_link` |
| GcpNetworkEndpointGroup | `spec.subnetwork` | `status.outputs.subnetwork_self_link` |
| GcpPlantonRunner | `spec.vpcAccess.subnetwork` | `status.outputs.subnetwork_name` |
| GcpPscServiceAttachment | `spec.natSubnets` | `status.outputs.subnetwork_self_link` |
| GcpRegionNetworkEndpointGroup | `spec.subnetwork` | `status.outputs.subnetwork_self_link` |
| GcpRouterNat | `spec.subnetworks[].subnetwork` | `status.outputs.subnetwork_self_link` |
| GcpRouterNat | `spec.rules[].action.sourceNatActiveRanges` | `status.outputs.subnetwork_self_link` |
| GcpRouterNat | `spec.rules[].action.sourceNatDrainRanges` | `status.outputs.subnetwork_self_link` |
| GcpRouterNat | `spec.nat64Subnetworks` | `status.outputs.subnetwork_self_link` |
| GcpServerlessVpcConnector | `spec.subnet.name` | `status.outputs.subnetwork_name` |
| GcpServiceConnectionPolicy | `spec.pscConfig.subnetworks` | `status.outputs.subnetwork_self_link` |
| GcpTpuQueuedResource | `spec.nodeSpecs[].node.networkConfig.subnetwork` | `status.outputs.subnetwork_self_link` |
| GcpTpuVm | `spec.networkConfig.subnetwork` | `status.outputs.subnetwork_self_link` |
| GcpTpuVm | `spec.networkConfigs[].subnetwork` | `status.outputs.subnetwork_self_link` |
| GcpVertexAiNotebook | `spec.networkInterface.subnet` | `status.outputs.subnetwork_self_link` |

## See Also

- [Overview](../README.md)

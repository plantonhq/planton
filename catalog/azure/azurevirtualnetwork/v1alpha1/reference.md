# AzureVirtualNetwork

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**AzureVirtualNetworkSpec** defines the configuration for creating an Azure
Virtual Network (VNet): the isolated, private IP address space every
network-attached Azure workload lives inside.

The virtual network is deliberately just the network. Address planning and
network-wide policy live here; everything that partitions or extends it is
its own composable resource referencing this network's outputs:
- AzureSubnet partitions the address space into workload segments.
- AzureNatGateway provides managed outbound connectivity for a subnet.
- AzurePrivateDnsZoneVirtualNetworkLink makes a private DNS zone
  resolvable from this network.

Keeping these as separate nodes means each has its own lifecycle -- a
hub-and-spoke topology adds subnets and DNS links without ever touching
the network resource itself.

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureVirtualNetwork
metadata:
  name: test-network
  org: test-org
  env: dev
spec:
  region: eastus
  resourceGroup:
    value: test-rg
  name: test-vnet
  addressSpaces:
    - "10.0.0.0/16"
  dnsServers:
    - "10.0.0.4"
  flowTimeoutInMinutes: 10
  tags:
    cost-center: platform
    owner: network-team
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.name` | `string` | yes |  |  |
| `spec.addressSpaces` | `[]string` |  |  |  |
| `spec.ipAddressPools` | `[]AzureVirtualNetworkIpAddressPool` |  |  |  |
| `spec.ipAddressPools[].id` | `string` | yes |  |  |
| `spec.ipAddressPools[].numberOfIpAddresses` | `string` | yes |  |  |
| `spec.dnsServers` | `[]string` |  |  |  |
| `spec.bgpCommunity` | `string` |  |  |  |
| `spec.ddosProtectionPlan` | `AzureVirtualNetworkDdosProtectionPlan` |  |  |  |
| `spec.ddosProtectionPlan.id` | `string` | yes |  |  |
| `spec.ddosProtectionPlan.enable` | `bool` |  |  |  |
| `spec.encryption` | `enum` |  |  |  |
| `spec.flowTimeoutInMinutes` | `int32` |  |  |  |
| `spec.privateEndpointVnetPolicies` | `enum` |  |  |  |
| `spec.edgeZone` | `string` |  |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.region

`string` · required

The Azure region where the virtual network will be created, e.g.
"eastus", "westeurope". A virtual network is a regional resource;
changing the region replaces it (and everything inside it).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.resourceGroup

`string | valueFrom` · required

The Azure resource group the virtual network will be created in.
Can be a literal resource-group name or a reference to an
AzureResourceGroup's name output.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.name

`string` · required

The name of the virtual network. Must be unique within the resource
group; 2-64 characters (alphanumerics, underscores, periods, and
hyphens; must start with a letter or number and end with a letter,
number, or underscore). Changing the name replaces the network -- and
with it every subnet and attachment inside -- so name it durably, after
the environment or workload domain it carries ("prod-hub", "payments").

- rule: Virtual network names start with a letter or number, end with a letter, number, or underscore, and may contain alphanumerics, underscores, periods, and hyphens
- rule: {"required":true,"string":{"minLen":"2","maxLen":"64"}}

### spec.addressSpaces

`[]string`

The IPv4/IPv6 CIDR blocks that form the network's address space, e.g.
["10.0.0.0/16"]. Multiple blocks are first-class: networks routinely
grow a second range when the first fills up, and dual-stack networks
carry an IPv4 and an IPv6 block side by side. Blocks can be added or
removed in place, but a block in use by subnets cannot shrink. Every
AzureSubnet's address_prefix must fall inside one of these blocks.

Exactly one of address_spaces or ip_address_pools must be set: either
you plan addresses yourself (this field) or you delegate allocation to
an Azure Network Manager IPAM pool (ip_address_pools).

### spec.ipAddressPools

`[]AzureVirtualNetworkIpAddressPool`

Delegated address allocation from Azure Network Manager IP Address
Management (IPAM) pools -- the alternative to hand-planned
address_spaces for organizations that manage address space centrally.
At most two entries, one per IP version (IPv4 and IPv6). The actual
CIDR ranges are provisioned by the pool at deploy time and surfaced in
the network's outputs.

Exactly one of address_spaces or ip_address_pools must be set.

- rule: {"repeated":{"maxItems":"2"}}

### spec.ipAddressPools[].id

`string` · required

The ARM resource ID of the Network Manager IPAM pool to allocate from.
Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Network/networkManagers/{nm}/ipamPools/{pool}

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.ipAddressPools[].numberOfIpAddresses

`string` · required

How many IP addresses to allocate from the pool, as a positive number
in string form (IPv6 allocations can exceed integer range). The
allocation can grow in place but never shrink.

- rule: number_of_ip_addresses must be a positive number, e.g. "256"
- rule: {"required":true}

### spec.dnsServers

`[]string`

Custom DNS server IP addresses for the network. When empty, Azure's
default resolver (168.63.129.16) serves the network -- the right choice
for most deployments, and required for Azure Private DNS zone
resolution to work directly. Set custom servers only when integrating
with on-premises DNS or a self-hosted resolver; VMs pick up the change
on their next DHCP lease renewal (restart to force it).

### spec.bgpCommunity

`string`

The BGP community attribute advertised with this network's routes over
ExpressRoute, in "asn:community" notation, e.g. "12076:20010". The ASN
segment is always 12076 (Microsoft's public ASN) today. Set this when
on-premises routers filter or prefer routes by community; leave unset
otherwise. Updatable in place.

- rule: BGP community must use "asn:community" notation with both segments between 1 and 65534, e.g. "12076:20010"

### spec.ddosProtectionPlan

`AzureVirtualNetworkDdosProtectionPlan`

Attach a DDoS Protection Plan to shield the network's public IPs with
always-on traffic monitoring and adaptive mitigation. The plan is a
separate (and billed) Azure resource shared across networks; this block
attaches an existing plan by its ARM ID. Omit the block entirely for
Azure's free basic protection.

### spec.ddosProtectionPlan.id

`string` · required

The ARM resource ID of the DDoS Protection Plan to attach.
Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Network/ddosProtectionPlans/{name}

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.ddosProtectionPlan.enable

`bool`

Whether protection through the attached plan is active. Mirrors ARM's
shape exactly: a plan can stay attached with protection toggled off
(enable = false) so it can be re-activated without re-attaching.

### spec.encryption

`enum`

Virtual network encryption: encrypts traffic between Azure VMs inside
the network (VM-to-VM over the Azure backbone). Unspecified leaves
encryption off (Azure's default). When enabled, the value declares what
happens to traffic from VM sizes that cannot encrypt: ALLOW_UNENCRYPTED
lets it flow in plaintext; DROP_UNENCRYPTED discards it. Note: ARM
currently accepts only ALLOW_UNENCRYPTED (DROP_UNENCRYPTED is not yet
generally available); the value is modeled because the API defines it.

Allowed values (use exactly as shown):

- `azure_virtual_network_encryption_enforcement_unspecified` -- Not specified: virtual network encryption is not enabled.
- `ALLOW_UNENCRYPTED` -- Encryption is enabled; traffic from VMs that cannot encrypt flows unencrypted. The only enforcement mode ARM currently accepts.
- `DROP_UNENCRYPTED` -- Encryption is enabled; traffic from VMs that cannot encrypt is dropped. Defined by the API but not yet generally available in ARM.

### spec.flowTimeoutInMinutes

`int32` · optional (explicit presence)

Connection-tracking flow timeout for intra-network flows, in minutes
(4-30). Unset uses Azure's default (4 minutes). Raise it for
long-lived idle connections (database sessions, message-bus consumers)
that must not be dropped between keepalives. Updatable in place.

- rule: {"int32":{"lte":30,"gte":4}}

### spec.privateEndpointVnetPolicies

`enum`

Network-wide private endpoint policy. Unspecified applies Azure's
default (DISABLED: no network policies evaluated for private
endpoints). BASIC enables basic policy evaluation on private endpoint
traffic across the whole network. Subnet-level private endpoint
network policies (on AzureSubnet) are the more common control; this
network-wide setting exists for centrally-governed environments.

Allowed values (use exactly as shown):

- `azure_virtual_network_private_endpoint_vnet_policies_unspecified` -- Not specified: ARM's default -- no network policies are evaluated for private endpoints in this network.
- `BASIC` -- Basic policy evaluation is enabled for private endpoint traffic across the network.

### spec.edgeZone

`string`

Deploy the network into an Azure Edge Zone (a metro-local Azure
extension) instead of the main region, e.g. "losangeles". Leave unset
for the standard region -- the overwhelmingly common case. Changing
the edge zone replaces the network.

### spec.tags

`map<string, string>`

Free-form tags applied to the virtual network, merged over the
Planton-derived resource tags (organization, environment, resource id);
a user tag with the same key wins. Tags are Azure's governance surface
-- Azure Policy enforces them and Microsoft Cost Management groups by
them -- so carry your org's ownership/cost-center conventions here.
Updatable in place.

## Validation Rules

- `address_source_exactly_one`: Set exactly one address source: either address_spaces (self-managed CIDR blocks) or ip_address_pools (Azure Network Manager IPAM allocation)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureVirtualNetwork, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.virtual_network_id` | `string` | The Azure Resource Manager ID of the virtual network. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Network/virtualNetworks/{name} This is the primary output referenced by downstream resources via StringValueOrRef. |
| `status.outputs.virtual_network_name` | `string` | The name of the virtual network. Echoed from the spec for convenience -- resources created by name inside the network (subnets via azurerm's virtual_network_name argument, for example) join on this. |
| `status.outputs.guid` | `string` | The stable GUID ARM assigns the virtual network at creation. Some Azure features (BGP community advertisement, network diagnostics) identify networks by this GUID rather than the ARM ID. |
| `status.outputs.address_spaces` | `[]string` | The address space actually carried by the network. Echoes the spec's address_spaces when self-managed; when ip_address_pools delegate allocation, this surfaces the CIDR blocks the IPAM pool provisioned -- the only place those ranges are visible for downstream planning. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureBastionHost | `spec.virtualNetworkId` | `status.outputs.virtual_network_id` |
| AzureDataFactoryIntegrationRuntime | `spec.azureSsis.vnetIntegration.vnetId` | `status.outputs.virtual_network_id` |
| AzureLoadBalancer | `spec.backendPools[].virtualNetworkId` | `status.outputs.virtual_network_id` |
| AzurePrivateDnsResolver | `spec.virtualNetworkId` | `status.outputs.virtual_network_id` |
| AzurePrivateDnsResolverVirtualNetworkLink | `spec.virtualNetworkId` | `status.outputs.virtual_network_id` |
| AzurePrivateDnsZoneVirtualNetworkLink | `spec.virtualNetworkId` | `status.outputs.virtual_network_id` |
| AzureSubnet | `spec.virtualNetworkId` | `status.outputs.virtual_network_id` |
| AzureVirtualHubConnection | `spec.remoteVirtualNetworkId` | `status.outputs.virtual_network_id` |
| AzureVirtualNetworkPeering | `spec.virtualNetworkId` | `status.outputs.virtual_network_id` |
| AzureVirtualNetworkPeering | `spec.remoteVirtualNetworkId` | `status.outputs.virtual_network_id` |
| KubernetesNetworkPolicy | `spec.ingressRules[].from[].ipBlock.cidr` | `status.outputs.address_spaces` |
| KubernetesNetworkPolicy | `spec.ingressRules[].from[].ipBlock.except` | `status.outputs.address_spaces` |
| KubernetesNetworkPolicy | `spec.egressRules[].to[].ipBlock.cidr` | `status.outputs.address_spaces` |
| KubernetesNetworkPolicy | `spec.egressRules[].to[].ipBlock.except` | `status.outputs.address_spaces` |

## See Also

- [Overview](../README.md)

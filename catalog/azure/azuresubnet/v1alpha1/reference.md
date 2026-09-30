# AzureSubnet

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**AzureSubnetSpec** defines the configuration for creating an Azure Subnet:
the workload segment that partitions a virtual network's address space.

The subnet is the composition hub of Azure networking. It is where the
network's building blocks actually meet a workload:
- a route table attaches here to steer the subnet's egress (route_table_id),
- a network security group attaches here to filter its traffic
  (network_security_group_id),
- a NAT gateway attaches here to own its outbound connectivity
  (nat_gateway_id),
and downstream resources -- AKS clusters, databases, private endpoints,
load balancers, VMs -- all deploy INTO a subnet by referencing its ID.
Attachments are declared subnet-side because that is Azure's own model:
one route table, NSG, or NAT gateway serves many subnets, while a subnet
carries at most one of each.

**Note:** Subnets do not have their own region -- they inherit the region
from their parent virtual network, which is why this spec has no `region`
field. They are also not tracked ARM resources, so they carry no tags.

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureSubnet
metadata:
  name: test-subnet
spec:
  virtualNetworkId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/virtualNetworks/test-vnet
  name: test-subnet
  addressPrefixes:
    - "10.0.1.0/24"
  serviceEndpoints:
    - Microsoft.Storage
  routeTableId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/routeTables/test-rt
  networkSecurityGroupId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/networkSecurityGroups/test-nsg
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.virtualNetworkId` | `string \| valueFrom` | yes |  | AzureVirtualNetwork (`status.outputs.virtual_network_id`) |
| `spec.name` | `string` | yes |  |  |
| `spec.addressPrefixes` | `[]string` |  |  |  |
| `spec.ipAddressPool` | `AzureSubnetIpAddressPool` |  |  |  |
| `spec.ipAddressPool.id` | `string` | yes |  |  |
| `spec.ipAddressPool.numberOfIpAddresses` | `string` | yes |  |  |
| `spec.serviceEndpoints` | `[]string` |  |  |  |
| `spec.serviceEndpointPolicyIds` | `[]string` |  |  |  |
| `spec.delegations` | `[]AzureSubnetDelegation` |  |  |  |
| `spec.delegations[].name` | `string` | yes |  |  |
| `spec.delegations[].serviceName` | `string` | yes |  |  |
| `spec.delegations[].actions` | `[]string` |  |  |  |
| `spec.privateEndpointNetworkPolicies` | `enum` |  |  |  |
| `spec.privateLinkServiceNetworkPoliciesEnabled` | `bool` |  | `true` |  |
| `spec.defaultOutboundAccessEnabled` | `bool` |  | `true` |  |
| `spec.sharingScope` | `enum` |  |  |  |
| `spec.routeTableId` | `string \| valueFrom` |  |  | AzureRouteTable (`status.outputs.route_table_id`) |
| `spec.networkSecurityGroupId` | `string \| valueFrom` |  |  | AzureNetworkSecurityGroup (`status.outputs.network_security_group_id`) |
| `spec.natGatewayId` | `string \| valueFrom` |  |  | AzureNatGateway (`status.outputs.nat_gateway_id`) |

## Field Details

### spec.virtualNetworkId

`string | valueFrom` · required

The parent virtual network, by ARM ID. The subnet is an ARM child of
the network: the network's ID carries both the resource group and the
network name, and the modules derive both from it rather than modeling
redundant fields that could contradict the referenced network.
Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Network/virtualNetworks/{name}
Changing the parent replaces the subnet.

- references: AzureVirtualNetwork (`status.outputs.virtual_network_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureVirtualNetwork, name: <that resource's name>, fieldPath: status.outputs.virtual_network_id}} -- a bare string does not parse

### spec.name

`string` · required

The name of the subnet. Must be unique within the virtual network;
1-80 characters (alphanumerics, underscores, periods, and hyphens;
must start with a letter or number and end with a letter, number, or
underscore). Changing the name replaces the subnet -- and forces
everything deployed into it to move -- so name it after the workload
tier it carries ("app", "data", "gateway"), not after transient details.

- rule: Subnet names start with a letter or number, end with a letter, number, or underscore, and may contain alphanumerics, underscores, periods, and hyphens
- rule: {"required":true,"string":{"minLen":"1","maxLen":"80"}}

### spec.addressPrefixes

`[]string`

The CIDR blocks assigned to the subnet, e.g. ["10.0.1.0/24"]. Every
block must fall inside the parent network's address space and must not
overlap other subnets. Multiple blocks are first-class: a dual-stack
subnet carries an IPv4 and an IPv6 block side by side.

Sizing guidance: Azure reserves 5 IPs per subnet (the first four and
the last). /24 (251 usable) suits general workloads; Application
Gateway needs at least /27; delegated database subnets are commonly
/28. Blocks can be changed in place while the subnet is empty, but a
block in use by deployed resources cannot shrink.

Exactly one of address_prefixes or ip_address_pool must be set: either
you plan addresses yourself (this field) or you delegate allocation to
an Azure Network Manager IPAM pool (ip_address_pool).

### spec.ipAddressPool

`AzureSubnetIpAddressPool`

Delegated address allocation from an Azure Network Manager IP Address
Management (IPAM) pool -- the alternative to hand-planned
address_prefixes for organizations that manage address space centrally.
The actual CIDR is provisioned by the pool at deploy time and surfaced
in the subnet's address_prefixes output.

Exactly one of address_prefixes or ip_address_pool must be set.

### spec.ipAddressPool.id

`string` · required

The ARM resource ID of the Network Manager IPAM pool to allocate from.
Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Network/networkManagers/{nm}/ipamPools/{pool}

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.ipAddressPool.numberOfIpAddresses

`string` · required

How many IP addresses to allocate from the pool, as a positive number
in string form (IPv6 allocations can exceed integer range). The
allocation can grow in place but never shrink.

- rule: number_of_ip_addresses must be a positive number, e.g. "256"
- rule: {"required":true}

### spec.serviceEndpoints

`[]string`

Azure service endpoints to enable on this subnet. Service endpoints
route traffic to the listed Azure services over the Azure backbone
instead of the public internet, and let those services' firewalls
admit traffic by subnet identity.

Common values: "Microsoft.Storage", "Microsoft.Sql",
"Microsoft.KeyVault", "Microsoft.AzureCosmosDB", "Microsoft.ServiceBus",
"Microsoft.EventHub", "Microsoft.Web", "Microsoft.ContainerRegistry",
"Microsoft.AzureActiveDirectory". ("Microsoft.Storage.Global" extends
storage endpoints across regions but is subscription-feature-gated.)

The provider's per-endpoint network_identifier (associating a
NetworkIdentifier resource with an individual endpoint) is
deliberately not modeled: it would turn this list of service names
into a list of objects for a niche preview capability, and the
Pulumi engine's SDK cannot express it. Revisit when the capability
matures and both engines carry it.

### spec.serviceEndpointPolicyIds

`[]string`

ARM IDs of Service Endpoint Policies to associate with the subnet.
Endpoint policies narrow a service endpoint's reach to specific
resources (e.g. only your storage accounts, not all of Azure Storage).
Plain ARM IDs: service endpoint policies are not modeled as a Planton
kind, being a niche zero-trust refinement of service endpoints.
Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Network/serviceEndpointPolicies/{name}

### spec.delegations

`[]AzureSubnetDelegation`

Service delegations for this subnet. A delegation hands the subnet to
an Azure PaaS service, permitting it to inject service-managed
resources and network rules. Delegated subnets are dedicated: the
delegated service becomes the only thing that can deploy into them.

Common delegations:
- PostgreSQL Flexible Server: "Microsoft.DBforPostgreSQL/flexibleServers"
- MySQL Flexible Server: "Microsoft.DBforMySQL/flexibleServers"
- Container App Environment: "Microsoft.App/environments"
- App Service VNet integration: "Microsoft.Web/serverFarms"

Nearly every real subnet has zero or one delegation; the list form
mirrors ARM's shape, which admits multiple.

### spec.delegations[].name

`string` · required

A label for this delegation, unique within the subnet. This is a
configuration-level name (visible in the portal and IaC state), not an
Azure resource name -- "postgresql" or "container-apps" reads well.

- rule: {"required":true}

### spec.delegations[].serviceName

`string` · required

The Azure service to delegate to. Must exactly match one of Azure's
supported delegation service names, e.g.
"Microsoft.DBforPostgreSQL/flexibleServers", "Microsoft.App/environments",
"Microsoft.Web/serverFarms", "Microsoft.ContainerInstance/containerGroups",
"Microsoft.Netapp/volumes", "Microsoft.Sql/managedInstances".

- rule: {"required":true}

### spec.delegations[].actions

`[]string`

The network actions the delegated service is permitted to perform,
e.g. "Microsoft.Network/virtualNetworks/subnets/join/action". Omit to
let Azure apply the service's default action set -- correct for
virtually all delegations; the field exists for services that document
an explicit action list.

### spec.privateEndpointNetworkPolicies

`enum`

Network-policy evaluation for PRIVATE ENDPOINTS in this subnet.
Unspecified applies Azure's default (disabled): NSG rules and route
tables are NOT evaluated for private endpoint traffic, which is what
almost every private-endpoint subnet wants. Enable one of the explicit
modes only in zero-trust architectures that must filter or route
private endpoint traffic; the granular values apply just the NSG side
or just the route-table side.

Allowed values (use exactly as shown):

- `azure_subnet_private_endpoint_network_policies_unspecified` -- Not specified: ARM's default -- NSG rules and route tables are not evaluated for private endpoint traffic in this subnet.
- `ENABLED` -- Both NSG rules and route tables apply to private endpoint traffic.
- `NETWORK_SECURITY_GROUP_ENABLED` -- Only NSG rules apply to private endpoint traffic.
- `ROUTE_TABLE_ENABLED` -- Only route tables apply to private endpoint traffic.

### spec.privateLinkServiceNetworkPoliciesEnabled

`bool` · optional (explicit presence)

Whether standard network policies (NSGs, user-defined routes) apply to
PRIVATE LINK SERVICE resources in this subnet. Azure's default is
true; set false only on subnets hosting a Private Link Service, which
requires policies off. Updatable in place.

- default: `true`

### spec.defaultOutboundAccessEnabled

`bool` · optional (explicit presence)

Whether workloads in this subnet get Azure's implicit default outbound
internet access. Azure's historical default is true, but Microsoft is
retiring implicit outbound (announced for September 2025 onward for
new subnets): production subnets should set this false and route
egress explicitly through a NAT gateway (nat_gateway_id), a load
balancer's outbound rules, or a firewall via route_table_id.
Changing it requires the subnet to be empty of VMs.

- default: `true`

### spec.sharingScope

`enum`

Opt the subnet into cross-tenant sharing through Azure Virtual Network
Manager. TENANT is the only value ARM currently accepts, and it
requires default_outbound_access_enabled to be explicitly false
(enforced here, mirroring ARM's own constraint). Leave unspecified for
the overwhelmingly common unshared subnet.

Allowed values (use exactly as shown):

- `azure_subnet_sharing_scope_unspecified` -- Not specified: the subnet is not shared beyond its own tenant.
- `TENANT` -- The subnet is shareable across the tenant through Azure Virtual Network Manager. The only sharing mode ARM currently accepts.

### spec.routeTableId

`string | valueFrom`

The route table that steers this subnet's traffic, by ARM ID.
Attaching a table replaces Azure's default system routing with the
table's user-defined routes for everything in the subnet -- the
firewall-egress and forced-tunneling seam. Omit for Azure's default
routing. The attachment is declared here (not on the table) because
one table serves many subnets; detaching is just removing this field.
The modules realize this attachment (and the NSG's) through the
dedicated association resources; the provider's write-only
route_table_id_wo / network_security_group_id_wo arguments on the
subnet itself are deliberately not modeled -- they are an
alternative set-path for the same attachments, not extra capability.

- references: AzureRouteTable (`status.outputs.route_table_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureRouteTable, name: <that resource's name>, fieldPath: status.outputs.route_table_id}} -- a bare string does not parse

### spec.networkSecurityGroupId

`string | valueFrom`

The network security group that filters this subnet's traffic, by ARM
ID. The NSG's rules apply to everything deployed in the subnet --
Azure's primary network access control. Omit to leave the subnet
governed only by Azure's implicit default rules (allow VNet-internal
and load-balancer traffic, deny other inbound). One NSG typically
guards many subnets of the same tier.

- references: AzureNetworkSecurityGroup (`status.outputs.network_security_group_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureNetworkSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.network_security_group_id}} -- a bare string does not parse

### spec.natGatewayId

`string | valueFrom`

The NAT gateway that owns this subnet's outbound connectivity, by ARM
ID. Attaching a NAT gateway gives every workload in the subnet stable,
SNAT-exhaustion-resistant egress through the gateway's public IPs --
the production answer to Azure retiring implicit outbound access
(default_outbound_access_enabled). One gateway serves many subnets in
the same region.

- references: AzureNatGateway (`status.outputs.nat_gateway_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureNatGateway, name: <that resource's name>, fieldPath: status.outputs.nat_gateway_id}} -- a bare string does not parse

## Validation Rules

- `address_source_exactly_one`: Set exactly one address source: either address_prefixes (self-managed CIDR blocks) or ip_address_pool (Azure Network Manager IPAM allocation)
- `sharing_scope_requires_no_default_outbound`: sharing_scope requires default_outbound_access_enabled to be explicitly false (ARM rejects a shared subnet with implicit outbound access)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureSubnet, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.subnet_id` | `string` | The Azure Resource Manager ID of the subnet. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Network/virtualNetworks/{vnet}/subnets/{name} This is the primary output referenced by downstream resources. |
| `status.outputs.subnet_name` | `string` | The name of the subnet within its virtual network. |
| `status.outputs.address_prefixes` | `[]string` | The CIDR blocks actually assigned to the subnet. For self-managed subnets this echoes address_prefixes; for IPAM-allocated subnets it carries the ranges the Network Manager pool provisioned. Useful in NSG rules, firewall rules, and network planning downstream. |
| `status.outputs.virtual_network_name` | `string` | The name of the parent virtual network, derived from the referenced network's ARM ID -- exported so charts can compose sibling resources without re-parsing the ID. |
| `status.outputs.resource_group_name` | `string` | The name of the resource group the subnet (and its parent network) lives in, derived from the referenced network's ARM ID. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.virtualNetworkId` | AzureVirtualNetwork | `status.outputs.virtual_network_id` |
| `spec.routeTableId` | AzureRouteTable | `status.outputs.route_table_id` |
| `spec.networkSecurityGroupId` | AzureNetworkSecurityGroup | `status.outputs.network_security_group_id` |
| `spec.natGatewayId` | AzureNatGateway | `status.outputs.nat_gateway_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureAksCluster | `spec.defaultNodePool.vnetSubnetId` | `status.outputs.subnet_id` |
| AzureAksCluster | `spec.defaultNodePool.podSubnetId` | `status.outputs.subnet_id` |
| AzureAksCluster | `spec.apiServerAccessProfile.subnetId` | `status.outputs.subnet_id` |
| AzureAksCluster | `spec.ingressApplicationGateway.subnetId` | `status.outputs.subnet_id` |
| AzureAksNodePool | `spec.vnetSubnetId` | `status.outputs.subnet_id` |
| AzureAksNodePool | `spec.podSubnetId` | `status.outputs.subnet_id` |
| AzureApplicationGateway | `spec.subnetId` | `status.outputs.subnet_id` |
| AzureApplicationGateway | `spec.frontendIpConfigurations[].subnetId` | `status.outputs.subnet_id` |
| AzureApplicationGateway | `spec.privateLinkConfigurations[].ipConfigurations[].subnetId` | `status.outputs.subnet_id` |
| AzureBastionHost | `spec.ipConfiguration.subnetId` | `status.outputs.subnet_id` |
| AzureCognitiveAccount | `spec.networkAcls.virtualNetworkRules[].subnetId` | `status.outputs.subnet_id` |
| AzureCognitiveAccount | `spec.networkInjection.subnetId` | `status.outputs.subnet_id` |
| AzureContainerAppEnvironment | `spec.infrastructureSubnetId` | `status.outputs.subnet_id` |
| AzureContainerInstance | `spec.subnetId` | `status.outputs.subnet_id` |
| AzureCosmosdbAccount | `spec.virtualNetworkRules[].subnetId` | `status.outputs.subnet_id` |
| AzureDataFactoryIntegrationRuntime | `spec.azureSsis.expressVnetIntegration.subnetId` | `status.outputs.subnet_id` |
| AzureDataFactoryIntegrationRuntime | `spec.azureSsis.vnetIntegration.subnetId` | `status.outputs.subnet_id` |
| AzureEventHubNamespace | `spec.networkRuleSets.virtualNetworkRules[].subnetId` | `status.outputs.subnet_id` |
| AzureFirewall | `spec.ipConfigurations[].subnetId` | `status.outputs.subnet_id` |
| AzureFirewall | `spec.managementIpConfiguration.subnetId` | `status.outputs.subnet_id` |
| AzureFunctionApp | `spec.siteConfig.ipRestrictions[].virtualNetworkSubnetId` | `status.outputs.subnet_id` |
| AzureFunctionApp | `spec.siteConfig.scmIpRestrictions[].virtualNetworkSubnetId` | `status.outputs.subnet_id` |
| AzureFunctionApp | `spec.virtualNetworkSubnetId` | `status.outputs.subnet_id` |
| AzureFunctionAppFlexConsumption | `spec.siteConfig.ipRestrictions[].virtualNetworkSubnetId` | `status.outputs.subnet_id` |
| AzureFunctionAppFlexConsumption | `spec.siteConfig.scmIpRestrictions[].virtualNetworkSubnetId` | `status.outputs.subnet_id` |
| AzureFunctionAppFlexConsumption | `spec.virtualNetworkSubnetId` | `status.outputs.subnet_id` |
| AzureKeyVault | `spec.networkAcls.virtualNetworkSubnetIds` | `status.outputs.subnet_id` |
| AzureLinuxWebApp | `spec.siteConfig.ipRestrictions[].virtualNetworkSubnetId` | `status.outputs.subnet_id` |
| AzureLinuxWebApp | `spec.siteConfig.scmIpRestrictions[].virtualNetworkSubnetId` | `status.outputs.subnet_id` |
| AzureLinuxWebApp | `spec.virtualNetworkSubnetId` | `status.outputs.subnet_id` |
| AzureLoadBalancer | `spec.frontendIpConfigurations[].subnetId` | `status.outputs.subnet_id` |
| AzureMachineLearningComputeCluster | `spec.subnetId` | `status.outputs.subnet_id` |
| AzureMachineLearningComputeInstance | `spec.subnetId` | `status.outputs.subnet_id` |
| AzureMachineLearningWorkspace | `spec.serverlessCompute.subnetId` | `status.outputs.subnet_id` |
| AzureMssqlServer | `spec.virtualNetworkRules[].subnetId` | `status.outputs.subnet_id` |
| AzureMysqlFlexibleServer | `spec.delegatedSubnetId` | `status.outputs.subnet_id` |
| AzureNetworkInterface | `spec.ipConfigurations[].subnetId` | `status.outputs.subnet_id` |
| AzurePostgresqlFlexibleServer | `spec.delegatedSubnetId` | `status.outputs.subnet_id` |
| AzurePrivateDnsResolver | `spec.inboundEndpoints[].subnetId` | `status.outputs.subnet_id` |
| AzurePrivateDnsResolver | `spec.outboundEndpoints[].subnetId` | `status.outputs.subnet_id` |
| AzurePrivateEndpoint | `spec.subnetId` | `status.outputs.subnet_id` |
| AzurePrivateLinkService | `spec.natIpConfigurations[].subnetId` | `status.outputs.subnet_id` |
| AzureRedisCache | `spec.subnetId` | `status.outputs.subnet_id` |
| AzureServiceBusNamespace | `spec.networkRuleSet.networkRules[].subnetId` | `status.outputs.subnet_id` |
| AzureStorageAccount | `spec.networkRules.virtualNetworkSubnetIds` | `status.outputs.subnet_id` |
| AzureVirtualMachineScaleSet | `spec.networkInterfaces[].ipConfigurations[].subnetId` | `status.outputs.subnet_id` |
| AzureVirtualNetworkGateway | `spec.ipConfigurations[].subnetId` | `status.outputs.subnet_id` |
| KubernetesNetworkPolicy | `spec.ingressRules[].from[].ipBlock.cidr` | `status.outputs.address_prefixes` |
| KubernetesNetworkPolicy | `spec.ingressRules[].from[].ipBlock.except` | `status.outputs.address_prefixes` |
| KubernetesNetworkPolicy | `spec.egressRules[].to[].ipBlock.cidr` | `status.outputs.address_prefixes` |
| KubernetesNetworkPolicy | `spec.egressRules[].to[].ipBlock.except` | `status.outputs.address_prefixes` |

## See Also

- [Overview](../README.md)

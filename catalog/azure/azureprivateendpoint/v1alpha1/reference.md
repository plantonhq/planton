# AzurePrivateEndpoint

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**AzurePrivateEndpointSpec** defines an Azure Private Endpoint: a network
interface that gives a service powered by Azure Private Link a private IP
address inside your virtual network. The service -- an Azure PaaS resource
(SQL, PostgreSQL, Storage, Key Vault, Cosmos DB, ...) or a custom Private
Link Service -- becomes reachable over a private IP on the Microsoft
backbone, never the public internet.

Private endpoints deliver three things: private connectivity (traffic
stays on the backbone), data-exfiltration protection (each endpoint maps
to one sub-resource -- "blob", "vault", "sqlServer" -- not the whole
service), and a simpler network (no service endpoints, NAT, or public IPs
to reach Azure services).

The private DNS zone group is part of this resource, not a separate node:
a private endpoint without DNS registration resolves to the service's
PUBLIC IP, silently defeating the private link. Folding the zone group in
keeps that guarantee atomic. DNS zones themselves are first-class
(AzurePrivateDnsZone) and referenced here by id.

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzurePrivateEndpoint
metadata:
  name: pg-pe
  org: test-org
  env: test
spec:
  region: eastus
  resourceGroup:
    value: test-rg
  name: pg-private-endpoint
  subnetId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/virtualNetworks/test-vnet/subnets/pe-subnet
  privateServiceConnection:
    privateConnectionResourceId:
      value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.DBforPostgreSQL/flexibleServers/test-pg
    subresourceNames:
      - postgresqlServer
  privateDnsZoneIds:
    - value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/privateDnsZones/privatelink.postgres.database.azure.com
  tags:
    purpose: hack-test
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.name` | `string` | yes |  |  |
| `spec.subnetId` | `string \| valueFrom` | yes |  | AzureSubnet (`status.outputs.subnet_id`) |
| `spec.privateServiceConnection` | `AzurePrivateEndpointServiceConnection` | yes |  |  |
| `spec.privateServiceConnection.privateConnectionResourceId` | `string \| valueFrom` |  |  |  |
| `spec.privateServiceConnection.connectionAlias` | `string` |  |  |  |
| `spec.privateServiceConnection.subresourceNames` | `[]string` |  |  |  |
| `spec.privateServiceConnection.isManualConnection` | `bool` |  |  |  |
| `spec.privateServiceConnection.requestMessage` | `string` |  |  |  |
| `spec.privateDnsZoneIds` | `[]string \| valueFrom` |  |  | AzurePrivateDnsZone (`status.outputs.zone_id`) |
| `spec.ipConfigurations` | `[]AzurePrivateEndpointIpConfiguration` |  |  |  |
| `spec.ipConfigurations[].name` | `string` | yes |  |  |
| `spec.ipConfigurations[].privateIpAddress` | `string` | yes |  |  |
| `spec.ipConfigurations[].subresourceName` | `string` |  |  |  |
| `spec.ipConfigurations[].memberName` | `string` |  |  |  |
| `spec.applicationSecurityGroupIds` | `[]string \| valueFrom` |  |  | AzureApplicationSecurityGroup (`status.outputs.application_security_group_id`) |
| `spec.customNetworkInterfaceName` | `string` |  |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.region

`string` · required

The Azure region the private endpoint is created in. Must match the
region of the subnet it draws its IP from. Fixed at creation.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.resourceGroup

`string | valueFrom` · required

The Azure resource group the private endpoint is created in. Can be a
literal resource-group name or a reference to an AzureResourceGroup's
name output.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.name

`string` · required

The name of the private endpoint, unique within the resource group.
1-80 characters (alphanumerics, underscores, periods, and hyphens;
must start with a letter or number and end with a letter, number, or
underscore). Fixed at creation.

- rule: Private endpoint names start with a letter or number, end with a letter, number, or underscore, and may contain alphanumerics, underscores, periods, and hyphens
- rule: {"required":true,"string":{"minLen":"1","maxLen":"80"}}

### spec.subnetId

`string | valueFrom` · required

The subnet the private endpoint draws its private IP from. The subnet
must permit private endpoints (its private-endpoint network policies
configured accordingly). Can be a literal subnet ARM ID or a reference
to an AzureSubnet's output. Fixed at creation.

- references: AzureSubnet (`status.outputs.subnet_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.privateServiceConnection

`AzurePrivateEndpointServiceConnection` · required

The private link connection this endpoint establishes -- which service
it tunnels to and which sub-resource of that service. Required: an
endpoint with no connection has nothing to reach.

- rule: {"required":true}
- rule: Set exactly one connection target: private_connection_resource_id (an Azure resource) or connection_alias (a Private Link Service alias)
- rule: request_message is required when is_manual_connection is true, and must be empty when it is false (Azure rejects a message on an auto-approved connection)

### spec.privateServiceConnection.privateConnectionResourceId

`string | valueFrom`

The Private Link-enabled resource to connect to, by ARM ID. Polymorphic
-- any Azure resource type supporting Private Link (PostgreSQL, MySQL,
Key Vault, Storage, Cosmos DB, SQL Server, ...) -- so it carries no
default_kind; reference the service's own output id in composed
environments. Set this OR connection_alias, never both. Fixed at
creation.

Containment-exempt: the endpoint LIVES in its subnet (the subnet
reference above places it) and REACHES the service named here.
Private endpoints are how Azure PaaS joins a network, and most of
what they reach is itself a container (a storage account, a Key
Vault, a Cosmos account, a SQL server); without the exemption an
endpoint would be drawn inside the service it merely connects to,
torn out of the subnet that is its real home.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.privateServiceConnection.connectionAlias

`string`

The Private Link Service ALIAS to connect to, when the target is
exposed through an alias rather than a resource ID (typically a
partner's cross-tenant Private Link Service). Aliases always end in
".azure.privatelinkservice". Set this OR
private_connection_resource_id, never both. Fixed at creation.

- rule: connection_alias must end with .azure.privatelinkservice

### spec.privateServiceConnection.subresourceNames

`[]string`

The sub-resource (group ID) of the target service this endpoint reaches
-- the data-exfiltration boundary. Most endpoints name exactly one.
Examples: PostgreSQL "postgresqlServer"; SQL "sqlServer"; Key Vault
"vault"; Storage "blob"/"table"/"queue"/"file"; Cosmos DB "Sql"/
"MongoDB"; Redis "redisCache"; Container Registry "registry". Omit when
connecting through a Private Link Service alias (the alias already pins
the target). Fixed at creation.

### spec.privateServiceConnection.isManualConnection

`bool` · optional (explicit presence)

Whether the connection requires manual approval by the target service's
owner. False (auto-approved) is the norm and works whenever you own or
are granted access to the target. Set true for cross-tenant or
cross-subscription connections where the owner must approve; then
request_message is required. Fixed at creation.

### spec.privateServiceConnection.requestMessage

`string`

The message shown to the target owner on a manual approval request
(1-140 characters). Required when is_manual_connection is true; must be
empty otherwise.

- rule: request_message must be 1-140 characters

### spec.privateDnsZoneIds

`[]string | valueFrom`

The private DNS zones this endpoint registers its IP into, as an A
record, so the service FQDN resolves to the private IP inside the VNet.
Each entry is a private DNS zone by ARM ID, or a reference to an
AzurePrivateDnsZone's output (typically the service's privatelink zone,
e.g. privatelink.postgres.database.azure.com). When empty, no DNS zone
group is created -- use that only when DNS is managed externally, since
without registration the FQDN resolves to the PUBLIC IP.

- references: AzurePrivateDnsZone (`status.outputs.zone_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzurePrivateDnsZone, name: <that resource's name>, fieldPath: status.outputs.zone_id}} -- a bare string does not parse

### spec.ipConfigurations

`[]AzurePrivateEndpointIpConfiguration`

Static private IP assignments for the endpoint's network interface.
Leave empty for dynamic allocation from the subnet (the common case);
set one entry per sub-resource when a service must land on a fixed IP
(firewall allowlists, hard-coded DNS). Fixed at creation.

### spec.ipConfigurations[].name

`string` · required

A name for this IP configuration, unique within the endpoint. Fixed at
creation.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.ipConfigurations[].privateIpAddress

`string` · required

The static private IP to assign, which must fall within the endpoint's
subnet range. Fixed at creation.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.ipConfigurations[].subresourceName

`string`

The sub-resource (group ID) this IP applies to -- must match one of the
connection's subresource_names (e.g. "blob"). Fixed at creation.

### spec.ipConfigurations[].memberName

`string`

The member name this IP applies to. When omitted, Azure uses the
subresource_name. Some services require it explicitly. Fixed at
creation.

### spec.applicationSecurityGroupIds

`[]string | valueFrom`

Application security groups this endpoint's network interface joins, so
NSG rules can govern traffic to the endpoint by workload group. Each
entry is an application security group by ARM ID, or a reference to an
AzureApplicationSecurityGroup's output. Realized as association
resources (Azure models ASG membership from the member side).

- references: AzureApplicationSecurityGroup (`status.outputs.application_security_group_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureApplicationSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.application_security_group_id}} -- a bare string does not parse

### spec.customNetworkInterfaceName

`string`

A custom name for the network interface Azure creates for this
endpoint. Leave empty to let Azure name it. Fixed at creation.

### spec.tags

`map<string, string>`

Free-form tags applied to the endpoint, merged over the Planton-derived
resource tags (organization, environment, resource id); a user tag with
the same key wins. Updatable in place.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzurePrivateEndpoint, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.private_endpoint_id` | `string` | The Azure Resource Manager ID of the private endpoint. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Network/privateEndpoints/{name} |
| `status.outputs.private_endpoint_name` | `string` | The name of the private endpoint resource. |
| `status.outputs.private_ip_address` | `string` | The private IP address allocated to the endpoint from the subnet. This is the IP the target service's FQDN resolves to inside the VNet; when a DNS zone group is configured, it is registered as an A record in each referenced private DNS zone automatically. |
| `status.outputs.network_interface_id` | `string` | The Azure Resource Manager ID of the network interface Azure created for this endpoint and attached to the subnet. Useful for diagnostics. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.subnetId` | AzureSubnet | `status.outputs.subnet_id` |
| `spec.privateDnsZoneIds` | AzurePrivateDnsZone | `status.outputs.zone_id` |
| `spec.applicationSecurityGroupIds` | AzureApplicationSecurityGroup | `status.outputs.application_security_group_id` |

## See Also

- [Overview](../README.md)

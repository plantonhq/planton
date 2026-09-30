# AzurePrivateLinkService

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**AzurePrivateLinkServiceSpec** defines a Private Link Service -- the
PROVIDER side of Azure Private Link. It sits in front of a service you
run (behind a Standard internal load balancer, or at a fixed
destination IP) and lets consumers in OTHER virtual networks -- other
teams, other subscriptions, even other Entra tenants -- reach it through
a private endpoint, with traffic never leaving the Microsoft backbone
and no VNet peering, no overlapping-address-space negotiation, and no
public exposure.

**Exactly one traffic destination**: either
`load_balancer_frontend_ip_configuration_ids` (the classic shape -- the
service runs behind a Standard load balancer's frontend) or
`destination_ip_address` (NAT straight to one fixed private IP, no load
balancer). ARM rejects both together and neither.

**The NAT subnet contract**: each nat_ip_configuration draws a NAT
address from a subnet whose `private_link_service_network_policies_enabled`
flag is FALSE -- ARM refuses to place a Private Link Service NAT
address on a subnet that still enforces those policies. Consumer
connections are source-NATed through these addresses; add more
configurations (up to 8) to widen the NAT port budget for very high
connection counts.

**ForceNew fields**: `name`, `region`, `resource_group`, and
`load_balancer_frontend_ip_configuration_ids` -- everything else
updates in place, with two ARM-enforced exceptions inside
nat_ip_configurations (see the field comments): a NAT address, once
assigned, cannot be cleared, and the PRIMARY configuration's subnet
cannot change.

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzurePrivateLinkService
metadata:
  name: test-private-link-service
spec:
  region: eastus
  resourceGroup:
    value: test-rg
  name: orders-api
  # NAT addresses consumer traffic is source-NATed through. The subnet
  # must have privateLinkServiceNetworkPoliciesEnabled: false. Exactly
  # one configuration is primary.
  natIpConfigurations:
    - name: nat-1
      subnetId:
        value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/virtualNetworks/test-vnet/subnets/pls-nat
      primary: true
  # The classic shape: the service fronts a STANDARD internal load
  # balancer's frontend (exactly one of this or destinationIpAddress).
  loadBalancerFrontendIpConfigurationIds:
    - value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/loadBalancers/orders-lb/frontendIPConfigurations/internal
  # Discoverable by one partner subscription; their connections still
  # wait for manual approval (no autoApprovalSubscriptionIds).
  visibilitySubscriptionIds:
    - "8158df85-3d6b-4d9f-8a3c-247b63cab0a8"
  tags:
    purpose: hack-test
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.name` | `string` | yes |  |  |
| `spec.natIpConfigurations` | `[]AzurePrivateLinkServiceNatIpConfiguration` | yes |  |  |
| `spec.natIpConfigurations[].name` | `string` | yes |  |  |
| `spec.natIpConfigurations[].subnetId` | `string \| valueFrom` | yes |  | AzureSubnet (`status.outputs.subnet_id`) |
| `spec.natIpConfigurations[].privateIpAddress` | `string` |  |  |  |
| `spec.natIpConfigurations[].privateIpAddressVersion` | `string` |  | `IPv4` |  |
| `spec.natIpConfigurations[].primary` | `bool` |  |  |  |
| `spec.loadBalancerFrontendIpConfigurationIds` | `[]string \| valueFrom` |  |  | AzureLoadBalancer (`status.outputs.frontend_ip_configuration_ids`) |
| `spec.destinationIpAddress` | `string` |  |  |  |
| `spec.proxyProtocolEnabled` | `bool` |  |  |  |
| `spec.autoApprovalSubscriptionIds` | `[]string` |  |  |  |
| `spec.visibilitySubscriptionIds` | `[]string` |  |  |  |
| `spec.fqdns` | `[]string` |  |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.region

`string` · required

The Azure region the Private Link Service lives in, e.g. "eastus".
Must match the load balancer (or destination) it fronts. Changing
the region replaces the service.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.resourceGroup

`string | valueFrom` · required

The Azure resource group the service is created in. Can be a literal
resource-group name or a reference to an AzureResourceGroup's name
output. Changing it replaces the service.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.name

`string` · required

The service's name, unique within the resource group. 1-80
characters; must begin with a letter or number, end with a letter,
number, or underscore, and may contain only letters, numbers,
underscores, periods, or hyphens. Changing the name replaces the
service.

- rule: Private Link Service names start with a letter or number, end with a letter, number, or underscore, and may contain alphanumerics, underscores, periods, and hyphens
- rule: {"required":true,"string":{"minLen":"1","maxLen":"80"}}

### spec.natIpConfigurations

`[]AzurePrivateLinkServiceNatIpConfiguration` · required

The NAT IP configurations consumer traffic is source-NATed through,
1-8 entries. Every configuration's subnet must have
private_link_service_network_policies_enabled = false. Exactly one
entry is `primary`. One configuration serves most services; add more
only when NAT port exhaustion is a real risk (each address funds
~64k concurrent flows per consumer endpoint).

- rule: {"repeated":{"minItems":"1","maxItems":"8"}}

### spec.natIpConfigurations[].name

`string` · required

The configuration's name, unique on the service. 1-80 characters;
same character rules as the service name.

- rule: NAT configuration names start with a letter or number, end with a letter, number, or underscore, and may contain alphanumerics, underscores, periods, and hyphens
- rule: {"required":true,"string":{"minLen":"1","maxLen":"80"}}

### spec.natIpConfigurations[].subnetId

`string | valueFrom` · required

The subnet the NAT address is drawn from -- references an
AzureSubnet's ARM id. The subnet MUST have
private_link_service_network_policies_enabled = false (ARM rejects
the configuration otherwise). ARM-enforced: the PRIMARY
configuration's subnet cannot change after creation.

- references: AzureSubnet (`status.outputs.subnet_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.natIpConfigurations[].privateIpAddress

`string`

A STATIC private IPv4 address from the subnet's range. Leave empty
for dynamic assignment (the module sets the allocation method from
this field's presence). ARM-enforced: once a static address is
assigned it cannot be cleared back to dynamic -- only replaced.

- rule: private_ip_address must be an IPv4 address from the subnet's range (or empty for dynamic assignment)

### spec.natIpConfigurations[].privateIpAddressVersion

`string` · optional (explicit presence)

The address family of the NAT address. Azure supports only "IPv4"
today; the field exists because the API reserves room for IPv6.

- default: `IPv4`
- rule: {"string":{"in":["IPv4"]}}

### spec.natIpConfigurations[].primary

`bool`

Whether this is the service's PRIMARY NAT configuration. Exactly one
configuration is primary (see the message contract). ARM-enforced:
the primary configuration's subnet is fixed once assigned.

### spec.loadBalancerFrontendIpConfigurationIds

`[]string | valueFrom`

The Standard load balancer FRONTEND IP configurations the service
fronts -- the classic Private Link shape. Reference the load
balancer's name-keyed map output, e.g. valueFrom fieldPath
"status.outputs.frontend_ip_configuration_ids.internal". Exactly one
of this or destination_ip_address. FIXED AT CREATION -- changing the
frontend set replaces the service.

- references: AzureLoadBalancer (`status.outputs.frontend_ip_configuration_ids`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureLoadBalancer, name: <that resource's name>, fieldPath: status.outputs.frontend_ip_configuration_ids}} -- a bare string does not parse

### spec.destinationIpAddress

`string`

NAT straight to ONE fixed private IPv4 address instead of a load
balancer frontend -- for single-instance services that need Private
Link without a load balancer in front. Exactly one of this or
load_balancer_frontend_ip_configuration_ids.

- rule: destination_ip_address must be an IPv4 address

### spec.proxyProtocolEnabled

`bool`

Accept PROXY protocol v2 headers from the consumer side, giving the
backend the consumer's original source IP (the backend service must
parse the PROXY v2 header). Off by default -- Azure's default.

### spec.autoApprovalSubscriptionIds

`[]string`

Azure subscription IDs whose private-endpoint connections are
APPROVED automatically instead of waiting in the manual-approval
queue. Every entry must also be visible (see
visibility_subscription_ids).

- rule: {"repeated":{"items":{"string":{"uuid":true}}}}

### spec.visibilitySubscriptionIds

`[]string`

Azure subscription IDs allowed to DISCOVER this service and request
a connection: specific subscription UUIDs, or the single entry "*"
to make the service visible to everyone with the alias (connections
still need approval unless auto-approved). Empty means visible only
by role-based access within your own tenant.

- rule: {"repeated":{"items":{"cel":[{"id":"visibility_entry_format","message":"Each visibility entry is a subscription UUID, or the single wildcard \"*\" for public discoverability","expression":"this == '*' || this.matches('^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$')"}]}}}

### spec.fqdns

`[]string`

Fully-qualified domain names associated with the service, surfaced
to consumers on their private-endpoint connections (for TLS-name or
client-validation schemes that need a stable name).

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.tags

`map<string, string>`

Free-form tags applied to the service, merged over the
Planton-derived resource tags (organization, environment, resource
id); a user tag with the same key wins.

## Validation Rules

- `exactly_one_destination`: Point the service at exactly one destination: load_balancer_frontend_ip_configuration_ids (service behind a Standard load balancer) or destination_ip_address (NAT to one fixed IP)
- `exactly_one_primary_nat_configuration`: Mark exactly one nat_ip_configuration as primary -- ARM requires a single primary NAT address
- `nat_configuration_names_unique`: NAT IP configuration names must be unique on the service

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzurePrivateLinkService, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.private_link_service_id` | `string` | The Azure Resource Manager ID of the Private Link Service. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Network/privateLinkServices/{name} |
| `status.outputs.private_link_service_name` | `string` | The name of the Private Link Service resource. |
| `status.outputs.alias` | `string` | The service's globally unique ALIAS -- what consumers use to request a private-endpoint connection without needing your resource ID or any RBAC on your subscription. Share this string with consumers. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.natIpConfigurations[].subnetId` | AzureSubnet | `status.outputs.subnet_id` |
| `spec.loadBalancerFrontendIpConfigurationIds` | AzureLoadBalancer | `status.outputs.frontend_ip_configuration_ids` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureFrontDoorOrigin | `spec.privateLink.privateLinkTargetId` | `status.outputs.private_link_service_id` |

## See Also

- [Overview](../README.md)

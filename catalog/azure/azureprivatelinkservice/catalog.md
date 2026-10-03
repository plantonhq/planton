# Azure Private Link Service

Deploys a Private Link Service -- the PROVIDER side of Azure Private Link. Your service (behind a Standard internal load balancer, or at one fixed destination IP) becomes privately consumable from other virtual networks, subscriptions, and Entra tenants through private endpoints: traffic stays on the Microsoft backbone, address spaces never meet, and nothing is exposed publicly. Consumers connect through the service's generated alias, so the string in the outputs -- not an IP or hostname -- is what you hand to the other side.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Private Link Service** -- the ARM object carrying the destination (LB frontends or destination IP), NAT configurations, visibility/auto-approval lists, and the generated consumer-facing alias
- **Azure Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) merged with your `tags`, applied to the service

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **A subnet with Private Link Service network policies DISABLED** (`privateLinkServiceNetworkPoliciesEnabled: false` on the AzureSubnet) -- ARM refuses to place the service's NAT addresses on a subnet that still enforces those policies.
- **A Standard internal load balancer** fronting your service (the classic shape), or a fixed destination IP for single-instance services.

## Deploy

### Console

Open the deployment store, find **Azure Private Link Service**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Behind a Load Balancer** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzurePrivateLinkService
metadata:
  name: orders-api
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    value: "network-rg"
  name: orders-api
  natIpConfigurations:
    - name: nat-1
      subnetId:
        value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/network-rg/providers/Microsoft.Network/virtualNetworks/hub-vnet/subnets/pls-nat-subnet"
      primary: true
  loadBalancerFrontendIpConfigurationIds:
    - value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/network-rg/providers/Microsoft.Network/loadBalancers/orders-lb/frontendIPConfigurations/internal"
```

```shell
planton apply -f azure-private-link-service.yaml
```

This publishes the service behind the internal load balancer's frontend, with one NAT address on the policies-disabled subnet and the consumer-facing alias in the outputs. An Infra Job tracks the provisioning in real time.

### InfraChart

In a service-publication chart, the service composes onto the load balancer and subnet deployed in the same InfraPipeline by reference:

```yaml
spec:
  natIpConfigurations:
    - name: nat-1
      subnetId:
        valueFrom:
          kind: AzureSubnet
          name: pls-nat-subnet
          fieldPath: status.outputs.subnet_id
      primary: true
  loadBalancerFrontendIpConfigurationIds:
    - valueFrom:
        kind: AzureLoadBalancer
        name: orders-lb
        fieldPath: status.outputs.frontend_ip_configuration_ids.internal
```

The InfraPipeline resolves the dependency graph, deploys the network and the load balancer first, then the Private Link Service -- and the alias output is what you hand to consumers.

## Key Configuration

These are the most important decisions when configuring a Private Link Service. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Destination** -- exactly one of `loadBalancerFrontendIpConfigurationIds` (the classic shape: your service behind a Standard internal LB) or `destinationIpAddress` (NAT straight to one fixed IP, no load balancer). Spec-enforced. The frontend set is FIXED at creation.

**NAT configurations** -- 1-8 addresses consumer traffic is source-NATed through, each on a policies-disabled subnet, exactly one `primary`. One suffices for most services; each address funds ~64k concurrent flows per consumer endpoint.

**Visibility and approval** -- `visibilitySubscriptionIds` gates discovery (subscription UUIDs, or `"*"` for anyone with the alias); `autoApprovalSubscriptionIds` skips the manual connection-approval queue for trusted subscriptions.

**PROXY protocol** -- enable `proxyProtocolEnabled` only when the backend parses PROXY v2 headers; enabling it against a backend that does not breaks every connection.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureSubnet** | `natIpConfigurations[].subnetId` | `status.outputs.subnet_id` |
| **AzureLoadBalancer** | `loadBalancerFrontendIpConfigurationIds[]` | `status.outputs.frontend_ip_configuration_ids.<frontend-name>` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `private_link_service_id` | Azure Resource Manager ID of the service | A same-graph AzurePrivateEndpoint's target |
| `private_link_service_name` | Name of the service | Operational tooling |
| `alias` | The globally unique consumer-facing handle | What external consumers request connections with |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Behind a load balancer** -- The classic shape: the service fronts a Standard internal LB frontend. Start from the **Behind a Load Balancer** preset.

**Fixed destination** -- NAT to one private IP, no load balancer. Start from the **Fixed Destination IP** preset.

## Works With

- [**Azure Subnet**](/infra-catalog/azure-subnet) -- provides the policies-disabled subnet the NAT addresses draw from
- [**Azure Load Balancer**](/infra-catalog/azure-load-balancer) -- the Standard internal LB the service typically fronts
- [**Azure Private Endpoint**](/infra-catalog/azure-private-endpoint) -- the CONSUMER side that connects to this service

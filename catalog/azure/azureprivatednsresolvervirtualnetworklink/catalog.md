# Azure DNS Resolver Virtual Network Link

Deploys the virtual network link that attaches ONE virtual network to a DNS forwarding ruleset -- the switch that turns the rule book on for that network. A ruleset without links steers nobody; each link adds one network to its audience, and linked networks do not need to be peered with the resolver's network or even live in the same subscription. Links are free at rest, and everything except metadata is fixed at creation.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Virtual network link** -- a child resource of the forwarding ruleset, one per ruleset-network pair (Azure allows up to 500 per ruleset)

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription.
- **An Azure DNS Forwarding Ruleset** -- the rule book this link turns on; `dnsForwardingRulesetId` defaults to referencing its `dns_forwarding_ruleset_id` output, which also orders the deploy.
- **An Azure Virtual Network** to attach -- in the ruleset's region; peering to the resolver's network is NOT required, and cross-subscription networks are allowed.

### Azure Subscription

- **Links are free at rest** -- only the resolver's endpoint hours and query volume bill.
- **Everything except metadata is create-only** -- changing the name, network, or ruleset replaces the link, a brief forwarding gap for that network.

## Deploy

### Console

Open the deployment store, find **Azure DNS Resolver Virtual Network Link**, and click **Deploy**. The creation wizard walks you through environment and connection configuration, the ruleset reference, and the network reference. Start from the **Spoke Link** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzurePrivateDnsResolverVirtualNetworkLink
metadata:
  name: spoke-payments-link
  org: acme-corp
  env: prod
spec:
  name: spoke-payments
  dnsForwardingRulesetId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-network/providers/Microsoft.Network/dnsForwardingRulesets/hub-dns-forwarding
  virtualNetworkId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-payments/providers/Microsoft.Network/virtualNetworks/spoke-payments
  metadata:
    owner: payments-team
```

```shell
planton apply -f link.yaml
```

This attaches the `spoke-payments` network to the hub's forwarding ruleset -- the moment it lands, resources in the spoke resolve the ruleset's domains through the hub resolver's outbound endpoint. An Infra Job tracks the provisioning in real time.

### InfraChart

When the ruleset and network are Infra Components in the same chart, wire both by reference:

```yaml
spec:
  name: spoke-payments
  dnsForwardingRulesetId:
    valueFrom:
      name: hub-dns-forwarding
  virtualNetworkId:
    valueFrom:
      name: spoke-payments-vnet
```

The InfraPipeline resolves the dependency graph, provisioning the ruleset and the network before the link that joins them.

## Key Configuration

These are the most important decisions when configuring an Azure DNS Resolver Virtual Network Link. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Link the hub's own network -- it is never implicit.** The resolver's home network does not inherit its own ruleset: it needs a link like every spoke. Forgetting it produces the classic asymmetry -- spokes resolve on-premises names, the hub does not. Make the hub link part of the ruleset's own deployment story.

**One link per pair, named after the network.** Azure allows exactly one link between a given ruleset and a given network; a duplicate create fails as already-exists. Name each link after the network it attaches (`spoke-payments`) so the ruleset's link list reads as its audience roster, and use the `metadata` map to record the network's owner -- when a spoke team decommissions their network, the link to clean up is self-identifying.

**Replacement means a resolution blink for that network.** Everything except metadata replaces the link, and while the link is gone the network stops forwarding -- captured domains fall back to Azure-internal resolution until the new link lands. Sequence link replacements outside change-sensitive windows for workloads that depend on on-premises names.

**The region wall is the ruleset's, not the resolver's.** A link's network must live in the RULESET's region. Multi-region estates need a resolver-plus-ruleset pair per region -- do not try to link a west network to the east rule book; build the west stack and copy the rules instead. Cross-subscription is fine; cross-region is not.

**Let the network's owner own the link.** Which networks forward through which rule book is exactly what a security review asks, and because each link is a standalone resource, the answer is the resource list itself. Keep links in the same chart as the network they attach -- spoke teams own their links -- rather than centrally bundled with the ruleset.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|---|---|---|
| Azure DNS Forwarding Ruleset | `dnsForwardingRulesetId` | `status.outputs.dns_forwarding_ruleset_id` |
| Azure Virtual Network | `virtualNetworkId` | `status.outputs.virtual_network_id` |

### What This Kind Provides

`status.outputs` carries the link's ARM ID (`virtual_network_link_id`) and name (`virtual_network_link_name`). Nothing downstream consumes a link by reference -- it is a leaf attachment binding two other resources -- so these outputs exist for identification and import rather than composition.

## Common Patterns

**Spoke onboarding** -- One link per spoke network, deployed alongside that network: the spoke starts resolving on-premises names through the hub with no resolver, no endpoints, and no peering of its own. Start from the **Spoke Link** preset.

**The hub's self-link** -- Deploy a link for the resolver's own network as part of the hub stack, so the hub resolves the same names its spokes do from day one.

**Audience as inventory** -- Because each ruleset-network pair is one named resource with owner metadata, the ruleset's link list doubles as the audit answer to "which networks forward through this rule book."

## Works With

- [**Azure DNS Forwarding Ruleset**](/infra-catalog/azure-private-dns-resolver-forwarding-ruleset) -- the rule book this link activates; reference its `dns_forwarding_ruleset_id` output.
- [**Azure Virtual Network**](/infra-catalog/azure-virtual-network) -- the network that starts forwarding; reference its `virtual_network_id` output.
- [**Azure DNS Private Resolver**](/infra-catalog/azure-private-dns-resolver) -- owns the outbound endpoint the linked networks' queries egress through.

# Azure DNS Forwarding Ruleset

Deploys a DNS forwarding ruleset for Azure DNS Private Resolver -- the declarative rule book that sends queries for chosen domains (say, your corporate Active Directory namespace) to the DNS servers that own them, through the resolver's outbound endpoint. Each rule pairs a domain with its ordered target servers; queries for everything else resolve normally inside Azure. The ruleset takes effect in a virtual network only once that network is linked to it.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **DNS forwarding ruleset** -- bound to the resolver's outbound endpoint(s)
- **Forwarding rules** (optional, up to 1,000) -- one per captured domain, each with its ordered target DNS servers, keyed by rule name so adding or removing one never touches its siblings

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription.
- **An Azure DNS Private Resolver with an outbound endpoint** -- `outboundEndpointIds` defaults to referencing the resolver's `outbound_endpoint_id` output, which also orders the deploy.

### Azure Subscription

- **A ruleset binds at most 2 outbound endpoints, both from the SAME resolver** -- Azure enforces it at deploy time. The ruleset's region must match the resolver's.
- **Rules take effect only in LINKED networks** -- deploy an Azure DNS Resolver Virtual Network Link per network afterwards; the ruleset alone steers nothing.
- **Free at rest** -- only the resolver's endpoint hours and query volume bill.

## Deploy

### Console

Open the deployment store, find **Azure DNS Forwarding Ruleset**, and click **Deploy**. The creation wizard walks you through environment and connection configuration, the outbound endpoint reference, and the forwarding rule list. Start from the **On-Premises Domain** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzurePrivateDnsResolverForwardingRuleset
metadata:
  name: hub-dns-forwarding
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    value: prod-network
  name: hub-dns-forwarding
  outboundEndpointIds:
    - value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-network/providers/Microsoft.Network/dnsResolvers/hub-dns-resolver/outboundEndpoints/outbound
  forwardingRules:
    - name: corp-domain
      domainName: corp.acme.com.
      targetDnsServers:
        - ipAddress: 10.100.0.10
        - ipAddress: 10.100.0.11
```

```shell
planton apply -f ruleset.yaml
```

This creates a ruleset bound to the resolver's outbound endpoint with one rule forwarding `corp.acme.com.` (and everything under it) to two datacenter DNS servers, tried in order on port 53. An Infra Job tracks the provisioning in real time.

### InfraChart

When the resolver is an Infra Component in the same chart, bind its outbound endpoint by reference:

```yaml
spec:
  region: eastus
  resourceGroup:
    valueFrom:
      name: prod-network
  name: hub-dns-forwarding
  outboundEndpointIds:
    - valueFrom:
        name: hub-dns-resolver
  forwardingRules:
    - name: corp-domain
      domainName: corp.acme.com.
      targetDnsServers:
        - ipAddress: 10.100.0.10
```

The InfraPipeline resolves the dependency graph, provisioning the resolver and its outbound endpoint before the ruleset that binds it.

## Key Configuration

These are the most important decisions when configuring an Azure DNS Forwarding Ruleset. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Write domains with the trailing dot, always.** ARM stores rule domains as fully qualified names: `corp.acme.com.` -- WITH the trailing dot. A rule captures the domain and everything under it, and the most specific rule wins when domains nest. Write every domain in the spec exactly that way and spot-check the first deploy's plan.

**The ruleset steers nothing until networks are linked.** Rules on an unlinked ruleset are inert configuration. Every virtual network that should forward through them needs its own link -- including the resolver's OWN network, which is not linked implicitly. Forgetting the hub's own link is the classic "works from spokes, fails from the hub" mystery.

**Domain edits replace the rule -- plan for the blink.** Everything on a rule updates in place except `domainName`: changing it deletes and recreates that rule, a brief window where the domain resolves inside Azure instead of forwarding. For a rename, add the new rule first, then remove the old one in a second apply -- rules are keyed by name, so both operations leave siblings untouched.

**Park rules with `enabled: false`, not by deleting them.** A disabled rule keeps its configuration but forwards nothing. Staging a migration, testing a tunnel, or backing out a bad target list is a one-field flip rather than a delete-and-retype. Per-rule `metadata` is the right place to record why a rule is parked and who owns it.

**Target servers are tried in order -- put the healthy one first.** Azure walks a rule's target list in order (up to 6 per rule), and targets must be reachable FROM THE OUTBOUND ENDPOINT'S SUBNET -- over VPN or ExpressRoute for on-premises targets. A rule whose targets are unreachable fails queries slowly rather than loudly; test resolution from a linked network after every target change. Ports default to 53; a non-standard port rides the target entry.

**Do not forward what Azure already answers.** A rule capturing a domain that also exists as a linked private DNS zone hijacks those queries to the on-premises servers -- which usually cannot answer for private endpoints. Keep rulesets to genuinely external namespaces and let private zones own theirs. Azure refuses rules for its own service domains outright.

**An empty ruleset is legal plumbing.** A ruleset with no rules forwards nothing -- a fine way to stage the binding and the network links before the rules land. Azure caps a ruleset at 1,000 rules and 500 virtual-network links.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|---|---|---|
| Azure Resource Group | `resourceGroup` | `status.outputs.resource_group_name` |
| Azure DNS Private Resolver | `outboundEndpointIds` | `status.outputs.outbound_endpoint_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `dns_forwarding_ruleset_id` | The ruleset's ARM resource ID | What each Azure DNS Resolver Virtual Network Link references to attach a network to the rules |

`dns_forwarding_ruleset_name` is also exported for identification.

## Common Patterns

**One rule book on the hub** -- The corporate namespace forwarded to two datacenter DNS servers, written once on the hub's ruleset and consumed by every linked network. Start from the **On-Premises Domain** preset.

**Staged multi-domain cutover** -- The live corporate domain plus an acquired company's domain parked with `enabled: false` until its connectivity is ready; flipping one field cuts it over, with no redeploys and no rule rewrites. Start from the **Staged Multi-Domain** preset.

**Plumbing before policy** -- Deploy the empty ruleset and the network links first, then land rules as domains are agreed -- rule adds are in-place and per-rule, so the rule book grows without touching what already forwards.

## Works With

- [**Azure DNS Private Resolver**](/infra-catalog/azure-private-dns-resolver) -- owns the outbound endpoint the ruleset binds; reference its `outbound_endpoint_id` output.
- [**Azure DNS Resolver Virtual Network Link**](/infra-catalog/azure-private-dns-resolver-virtual-network-link) -- attaches each consuming network to this ruleset; without links, rules steer nothing.
- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- where the ruleset lives; reference its `resource_group_name` output.
- [**Azure Virtual Network**](/infra-catalog/azure-virtual-network) -- the networks whose DNS queries the rules steer, hub and spokes alike, each attached through a link.

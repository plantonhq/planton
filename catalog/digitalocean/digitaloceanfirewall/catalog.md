# DigitalOcean Cloud Firewall

Deploys a DigitalOcean Cloud Firewall with configurable inbound and outbound rules, applied to Droplets by reference or by tag. Firewalls filter traffic at the network edge before it reaches the Droplet, providing stateful packet inspection without consuming Droplet resources, and they deny everything not explicitly allowed. Rule sources and destinations can name other Planton resources -- Droplets, load balancers, Kubernetes clusters -- as references instead of hand-copied IDs, so firewalls compose in InfraCharts.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Cloud Firewall** -- a stateful, default-deny network firewall with the configured inbound and outbound rules, applied to the referenced Droplets and/or Droplet tags
- **Inbound Rules** -- which sources (CIDR addresses, Droplet tags, Droplets, Kubernetes clusters, load balancers) can reach the protected Droplets on which protocols and ports
- **Outbound Rules** -- which destinations the protected Droplets can reach on which protocols and ports

## Before You Deploy

### Planton Setup

- **DigitalOcean Provider Connection** -- an active connection in the Connect module with a DigitalOcean API token. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline API token authentication.

### DigitalOcean Account

- **Droplets or Droplet tags to protect** -- target Droplets by reference (`dropletIds`, up to 10) or by tag (`tags`, up to 5; any Droplet carrying the tag is protected automatically). Every tag a firewall names must already exist on the account -- DigitalOcean rejects the firewall otherwise -- and Droplets create tags implicitly when they declare them, so tag the Droplets first. Tag targeting is the production standard for anything long-lived.
- **Network planning** -- decide which protocols, ports, and sources your application needs. Firewalls deny everything not explicitly allowed, and at least one rule (in either direction) is required.

## Deploy

### Console

Open the deployment store, find **DigitalOcean Cloud Firewall**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Web-Tier Firewall** preset in the [Presets](#presets) tab for a web-facing firewall with HTTPS/HTTP inbound and restricted SSH.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanFirewall
metadata:
  name: web-firewall
  org: acme-corp
  env: prod
spec:
  firewallName: web-firewall
  tags:
    - web
  inboundRules:
    - protocol: tcp
      portRange: "443"
      sourceAddresses:
        - "0.0.0.0/0"
        - "::/0"
    - protocol: tcp
      portRange: "80"
      sourceAddresses:
        - "0.0.0.0/0"
        - "::/0"
  outboundRules:
    - protocol: tcp
      portRange: all
      destinationAddresses:
        - "0.0.0.0/0"
        - "::/0"
```

```shell
planton apply -f do-firewall.yaml
```

This creates a firewall allowing HTTPS and HTTP inbound from all addresses and all TCP outbound, applied to Droplets tagged `web`. SSH access is not included. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying a full tier in one InfraPipeline, use ValueFromRef to protect Droplets deployed alongside the firewall and to admit traffic only from their load balancer:

```yaml
spec:
  dropletIds:
    - valueFrom:
        kind: DigitalOceanDroplet
        name: web-server-1
        fieldPath: status.outputs.droplet_id
  inboundRules:
    - protocol: tcp
      portRange: "80"
      sourceLoadBalancerUids:
        - valueFrom:
            kind: DigitalOceanLoadBalancer
            name: web-lb
            fieldPath: status.outputs.load_balancer_id
```

The InfraPipeline resolves the dependency graph, deploys the Droplet and load balancer first, then provisions the firewall with the resolved IDs -- so backends accept traffic only from their balancer, with no hand-copied UUIDs.

## Key Configuration

These are the most important decisions when configuring a Cloud Firewall. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Targeting strategy** -- Use `tags` for dynamic targeting (any Droplet with the tag inherits the firewall) or `dropletIds` for explicit targeting by reference to `DigitalOceanDroplet` resources or literal numeric IDs. Tag targeting scales better where Droplets are created and destroyed by automation; the ID list caps at 10, must be edited on every membership change, and exists for firewalls that genuinely protect one known machine -- a bastion, a fixture. The API caps targeting at 5 tags per firewall.

**Inbound rules** -- Each rule specifies a `protocol` (tcp, udp, or icmp), a `portRange` (`"443"`, `"8000-9000"`, or `all`; omitted for icmp), and source filters: CIDR addresses (IPv4 and IPv6), Droplet tags, and references to Droplets, Kubernetes clusters, or load balancers. Restrict SSH (port 22) to a management CIDR rather than `0.0.0.0/0`.

**Port ranges in canonical spelling** -- Write `all`, never `1-65535`: the API reports "all ports" back as the literal string `all`, so a rule authored as `1-65535` deploys fine and then diffs forever. Omit `portRange` on icmp rules -- the provider silently drops it on read. DigitalOcean hashes each rule as a set member, so any normalized leaf makes a plan show the whole rule as removed-and-re-added; canonical spelling keeps diffs empty.

**Egress is a tier decision** -- Web tiers need open egress (package mirrors, external APIs, DNS) and get all-TCP outbound plus UDP 53. Data tiers do not: a database host whose egress is limited to DNS and HTTPS cannot stream its contents to an arbitrary endpoint, which turns a class of exfiltration into a firewall log entry.

**Multi-tier architecture** -- Use `sourceTags` in inbound rules to allow traffic only from specific tiers: a database firewall allowing port 5432 only from Droplets tagged `web` enforces segmentation with zero IP management.

**The double-firewall trap** -- A Cloud Firewall filters at DigitalOcean's edge; a host firewall (ufw, nftables) filters on the Droplet. Images that enable a host firewall by default will block traffic this firewall allows, and the symptom -- timeouts on an explicitly allowed port -- looks exactly like a Cloud Firewall bug. Check the host firewall before editing rules here.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **DigitalOceanDroplet** (optional) | `dropletIds` | `status.outputs.droplet_id` |
| **DigitalOceanDroplet** (optional, per rule) | `inboundRules[].sourceDropletIds`, `outboundRules[].destinationDropletIds` | `status.outputs.droplet_id` |
| **DigitalOceanKubernetesCluster** (optional, per rule) | `inboundRules[].sourceKubernetesIds`, `outboundRules[].destinationKubernetesIds` | `status.outputs.cluster_id` |
| **DigitalOceanLoadBalancer** (optional, per rule) | `inboundRules[].sourceLoadBalancerUids`, `outboundRules[].destinationLoadBalancerUids` | `status.outputs.load_balancer_id` |

### What This Component Provides

`status.outputs` carries a single value: `firewall_id`, the UUID DigitalOcean assigns to the firewall. No other catalog kind references a firewall, so there is no ValueFromRef story -- the ID is the handle for imports and DigitalOcean API operations. The firewall's live status and pending per-Droplet changes are deliberately not outputs: they are apply-time snapshots that go stale immediately, and verification reads them from the live API instead.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Web tier firewall** -- HTTPS/HTTP from everywhere, SSH from a management CIDR, open egress, tag-targeted. Start from the **Web-Tier Firewall** preset.

**Database tier firewall** -- inbound only from the web tier's tag on the database port, SSH from a management CIDR, egress restricted to DNS and HTTPS. Start from the **Database-Tier Firewall** preset.

## Works With

- [**DigitalOcean Droplet**](/cloud-catalog/digital-ocean-droplet) -- the Droplets this firewall protects and the rule-level source/destination Droplets
- [**DigitalOcean Load Balancer**](/cloud-catalog/digital-ocean-load-balancer) -- rule-level sources so backends accept traffic only from their balancer
- [**DigitalOcean Kubernetes Cluster**](/cloud-catalog/digital-ocean-kubernetes-cluster) -- rule-level sources/destinations for cluster-to-Droplet traffic

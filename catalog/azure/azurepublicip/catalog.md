# Azure Public IP

Deploys an Azure Public IP Address -- a static, internet-routable address that load balancers, application gateways, NAT gateways, firewalls, and virtual machines attach for inbound or outbound connectivity. The spec covers the full address contract: SKU and tier, IP version, availability zones, allocation from a reserved Public IP Prefix, the Azure-managed DNS label with its reuse scope, reverse DNS, DDoS protection stance, idle timeout, routing IP tags, and edge-zone placement. Allocation is always static, and most of the address contract -- SKU, tier, zones, prefix -- is fixed at creation, so the shape decisions come first.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Public IP Address** -- a static public IP in the specified region and resource group. Allocation is always static (the retired Basic SKU's dynamic allocation is not modeled); the address is assigned at creation and persists for the resource's lifetime. Unspecified SKU, tier, and version deploy Azure's defaults (Standard, Regional, IPv4)
- **Prefix Allocation** -- when `publicIpPrefixId` is set, the address is drawn from the reserved Public IP Prefix's contiguous, allowlistable range instead of Microsoft's general pool
- **DNS Configuration** -- when `domainNameLabel` is set, an Azure-managed name at `{label}.{region}.cloudapp.azure.com` (hashed per `domainNameLabelScope` when one is chosen); when `reverseFqdn` is set, the address's reverse-DNS (PTR) name
- **DDoS Protection Stance** -- inherit the virtual network's plan (default), opt out, or attach dedicated IP-level protection backed by `ddosProtectionPlanId`
- **Availability Zone Configuration** -- zone-redundant or zonal placement based on the `zones` field; empty leaves the address non-zonal
- **Azure Tags** -- user tags merged over the Planton-derived resource tags (organization, environment, resource ID); IP tags (routing metadata such as `RoutingPreference`) applied separately when set

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** where the Public IP will be created. Provide the name directly or reference an AzureResourceGroup Infra Component via ValueFromRef.
- **Region alignment** -- the Public IP must be in the same region as the resource it will be attached to (load balancer, application gateway, NAT gateway, firewall, or VM).
- **Availability zone support** -- verify the target region supports the desired availability zones if using zone-redundant configuration.
- **A Public IP Prefix (optional)** -- to allocate from a reserved range, the prefix must live in the same region and carry a matching SKU.
- **A DDoS protection plan (optional)** -- required in practice when `ddosProtectionMode` is `ENABLED`; referenced by plain ARM ID.

## Deploy

### Console

Open the deployment store, find **Azure Public IP**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Standard Static Public IP** preset in the [Presets](#presets) tab to pre-populate a zone-redundant configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzurePublicIp
metadata:
  name: lb-public-ip
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    value: "acme-prod-rg"
  name: lb-public-ip
  zones:
    - "1"
    - "2"
    - "3"
```

```shell
planton apply -f public-ip.yaml
```

This creates a zone-redundant public IP with static allocation; the unspecified SKU, tier, and version deploy Azure's defaults (Standard, Regional, IPv4), and the unset idle timeout leaves Azure's 4-minute default. No DNS label is configured. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the Public IP to a resource group deployed in the same InfraPipeline:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: production-rg
      fieldPath: status.outputs.resource_group_name
  publicIpPrefixId:
    valueFrom:
      kind: AzurePublicIpPrefix
      name: prod-egress
      fieldPath: status.outputs.public_ip_prefix_id
```

The InfraPipeline resolves the dependency graph, deploys the resource group (and prefix) first, then provisions the Public IP with the resolved values.

## Key Configuration

These are the most important decisions when configuring a Public IP. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**SKU and tier** -- Unspecified deploys Azure's defaults (Standard SKU, Regional tier) -- correct for virtually everything. `STANDARD_V2` exists for StandardV2 NAT gateway attachment; `GLOBAL` tier exists solely for cross-region load balancer frontends and requires the STANDARD SKU (the spec enforces the pairing). All fixed at creation.

**Availability zones** -- Set `zones` to `["1", "2", "3"]` for zone-redundant placement that survives single-zone failures (recommended for production). Use a single zone for zonal placement pinned to one AZ. Empty leaves the address non-zonal (NOT zone-redundant). Fixed at creation.

**Prefix allocation** -- Set `publicIpPrefixId` to draw the address from a reserved Public IP Prefix: every address comes from one contiguous CIDR partners allowlist once. Fixed at creation.

**DNS label and reuse scope** -- Set `domainNameLabel` to mint a stable Azure-managed name at `{label}.{region}.cloudapp.azure.com`. The classic label must be unique within the region; set `domainNameLabelScope` to hash the label (the dangling-DNS subdomain-takeover defense -- the scope requires a label and is fixed once chosen). `reverseFqdn` records the address's PTR name for outbound-email deliverability; create the forward record first.

**DDoS protection** -- Unspecified inherits the virtual network's protection plan. `ENABLED` attaches dedicated IP-level protection (pair it with `ddosProtectionPlanId` -- the spec enforces the pairing); `DISABLED` deliberately opts the address out. Updatable in place.

**Idle timeout** -- `idleTimeoutInMinutes` controls how long Azure keeps idle TCP connections open (4-30). Unset leaves Azure's 4-minute default. Increase to 15-30 minutes for long-lived connections (WebSocket, gRPC streaming, database connections). Updatable in place.

**IP tags and edge zone** -- `ipTags` attach routing metadata to the address itself (e.g. `RoutingPreference: Internet` for hot-potato transit; only Azure-permitted pairs deploy). `edgeZone` places the address in a metro-local Azure Edge Zone instead of the main region -- leave unset for the standard region. Both fixed at creation.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzurePublicIpPrefix** | `publicIpPrefixId` | `status.outputs.public_ip_prefix_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `public_ip_id` | Azure resource ID of the Public IP | Application Gateway frontend, Load Balancer frontend, NAT Gateway association, firewall IP configuration |
| `ip_address` | The allocated static IP address | DNS configuration, firewall allowlisting, external service registration |
| `fqdn` | Fully qualified domain name (empty if no `domainNameLabel` set) | Application endpoint URLs, health check targets |
| `public_ip_name` | Name of the Public IP resource | Network diagnostics, audit logging references |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Standard zone-redundant IP** -- A zone-redundant public IP spanning all three availability zones with Azure's default SKU, tier, and idle timeout. Suitable for production load balancers, application gateways, and NAT gateways that need high availability. Start from the **Standard Static Public IP** preset.

**DNS-labeled endpoint** -- A public IP with an Azure-managed hostname for services that need a stable DNS name before a custom domain exists. Start from the **DNS-Labeled Endpoint** preset.

**Allowlisted from a prefix** -- An address drawn from a reserved Public IP Prefix so partners allowlist one CIDR, forever. Start from the **Allowlisted Address from a Prefix** preset.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group where the Public IP is created
- [**Azure Public IP Prefix**](/infra-catalog/azure-public-ip-prefix) -- the reserved, allowlistable range the address can be drawn from

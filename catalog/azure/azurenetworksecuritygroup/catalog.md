# Azure Network Security Group

Deploys an Azure Network Security Group (NSG) -- the stateful firewall that filters inbound and outbound traffic for everything deployed in the subnets and NICs it guards. Each security rule is a 5-tuple filter (source, destination, port, protocol, direction) with an access decision and a priority; sources and destinations take one addressing style each -- a single prefix (CIDR, IP, service tag, or `*`), a list of CIDRs/IPs, or application security groups (identity-based addressing that follows workloads as they scale). The attachment is always the guarded side's declaration — a subnet or NIC references this NSG, so one group serves many subnets without listing them.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Network Security Group** -- a stateful firewall resource in the specified resource group and region, with a name matching your manifest
- **Security Rules** -- one rule per entry in the `securityRules` array, each realized as its own ARM rule resource under the group. An empty list is meaningful: Azure's implicit default rules then govern (allow VNet-internal traffic and load-balancer probes, deny all other inbound, allow all outbound)
- **Azure Tags** -- user tags merged over the Planton-derived resource tags (organization, environment, resource ID); a user tag with the same key wins

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** where the NSG will be created. Provide the name directly or reference an AzureResourceGroup Infra Component via ValueFromRef.
- **Subnet association** -- the attachment is the guarded side's declaration: a subnet's `networkSecurityGroupId` (or a NIC's security settings) references this NSG, so one group serves many subnets without listing them. This component creates the NSG and its rules, never the association.
- **Application Security Groups (optional)** -- rules that address workloads by role reference AzureApplicationSecurityGroup resources (up to 10 per side per rule).

## Deploy

### Console

Open the deployment store, find **Azure Network Security Group**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Web Tier NSG** preset in the [Presets](#presets) tab to pre-populate rules allowing HTTP and HTTPS inbound traffic.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureNetworkSecurityGroup
metadata:
  name: web-tier-nsg
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    value: "acme-prod-rg"
  name: web-tier
  securityRules:
    - name: allow-https-inbound
      description: Allow HTTPS from the internet
      priority: 100
      direction: INBOUND
      access: ALLOW
      protocol: TCP
      destinationPortRange: "443"
      sourceAddressPrefix: Internet
```

```shell
planton apply -f nsg.yaml
```

This creates an NSG with a single rule allowing inbound HTTPS traffic from the internet. Azure's implicit default rules (VNet-to-VNet allow, load-balancer probes allow, deny all other inbound, allow all outbound) remain in effect underneath. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the NSG to a resource group deployed in the same InfraPipeline, and address workloads by role through application security groups:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: production-rg
      fieldPath: status.outputs.resource_group_name
  securityRules:
    - name: allow-app-to-db
      priority: 200
      direction: INBOUND
      access: ALLOW
      protocol: TCP
      destinationPortRange: "5432"
      sourceApplicationSecurityGroupIds:
        - valueFrom:
            kind: AzureApplicationSecurityGroup
            name: app-servers
            fieldPath: status.outputs.application_security_group_id
```

The InfraPipeline resolves the dependency graph, deploys the referenced resources first, then provisions the NSG with the resolved values.

## Key Configuration

These are the most important decisions when configuring an NSG. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Rule priority** -- Each rule's `priority` (100-4096, unique per direction) determines evaluation order -- lower numbers are evaluated first, and the first matching rule wins. Leave gaps (100, 200, 300) so rules can be inserted later without renumbering. Azure's implicit defaults sit at 65000+ and only apply when nothing else matched.

**Direction and access** -- Every rule declares `direction` (`INBOUND` or `OUTBOUND`), `access` (`ALLOW` or `DENY`), and `protocol` (`TCP`, `UDP`, `ICMP`, `ANY`, or the IPsec pair `AH`/`ESP`). Combine deny rules at high priority numbers (e.g., 4000) with specific allow rules at low numbers for a default-deny posture.

**Source and destination addressing** -- Each side takes ONE style: `sourceAddressPrefix`/`destinationAddressPrefix` for a single CIDR, IP, service tag (`VirtualNetwork`, `Internet`, `AzureLoadBalancer`), or `*` (service tags and `*` are singular-only); the plural `sourceAddressPrefixes`/`destinationAddressPrefixes` for multi-CIDR rules; or `sourceApplicationSecurityGroupIds`/`destinationApplicationSecurityGroupIds` for identity-based addressing that follows workloads as they scale (up to 10 per side). Leaving all three unset means any.

**Port forms** -- `destinationPortRange` (a single port `"443"`, a range `"1024-65535"`, or `"*"`) or the plural `destinationPortRanges` -- exactly one of the two. Source ports are typically left unset (any), since clients use ephemeral ports; the same single-or-list choice applies when set.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureApplicationSecurityGroup** | rule `sourceApplicationSecurityGroupIds` / `destinationApplicationSecurityGroupIds` | `status.outputs.application_security_group_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `network_security_group_id` | Azure Resource Manager ID of the Network Security Group | A subnet's `networkSecurityGroupId` attachment, a NIC's security settings, diagnostic settings |

`status.outputs` also echoes `network_security_group_name` back; attachments travel by ARM ID, not name.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Web tier** -- Allows inbound HTTP (80) and HTTPS (443) from the internet using the `Internet` service tag. The standard pattern for subnets hosting load balancers, application gateways, or public-facing web servers. Start from the **Web Tier NSG** preset.

**Database tier** -- Allows only PostgreSQL (5432) and MySQL (3306) traffic from within the VNet using the `VirtualNetwork` service tag, with an explicit deny-all for internet traffic. Suitable for subnets hosting managed databases or self-hosted database servers. Start from the **Database Tier NSG** preset.

**Bastion** -- Allows SSH (22) and RDP (3389) only from trusted IP ranges with an explicit deny-all catch-all rule. Suitable for bastion or jump-host subnets requiring controlled, auditable remote access. Start from the **Bastion NSG** preset.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group where the NSG is created
- [**Azure Application Security Group**](/infra-catalog/azure-application-security-group) -- the workload-role addressing rules target instead of pinning CIDRs
- [**Azure Subnet**](/infra-catalog/azure-subnet) -- declares which NSG guards it via `networkSecurityGroupId`

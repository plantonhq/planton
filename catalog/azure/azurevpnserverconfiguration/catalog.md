# Azure VPN Server Configuration

Deploys a VPN Server Configuration -- the reusable "who may connect and how" policy for point-to-site VPN: the authentication methods remote users sign in with (Entra ID, certificate, RADIUS), the trusted and revoked certificates, the tunnel protocols offered, and optional policy groups for user segmentation. The configuration is free and deploys in seconds; a Point-to-Site VPN Gateway is born pointing at one, and many gateways can share it. Everything except the name, region, and resource group updates in place -- gateways using the configuration pick the change up, and users reconnect under the new policy.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **VPN Server Configuration** -- the authentication policy object (Entra ID / certificate / RADIUS parameters, IPsec proposal, tunnel protocols)
- **Policy Groups** (optional) -- one ARM child per `policyGroups` entry, keyed by name (the `policy_group_ids` output republishes each group's ARM ID)

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **The authentication material for each enabled type**: the Entra ID tenant/audience/issuer values for "AAD", a root certificate's public data for "Certificate", or reachable RADIUS servers and their shared secrets for "Radius".

## Deploy

### Console

Open the deployment store, find **Azure VPN Server Configuration**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Entra ID Remote Workforce** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureVpnServerConfiguration
metadata:
  name: remote-workforce
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: network-rg
      fieldPath: status.outputs.resource_group_name
  name: remote-workforce
  vpnAuthenticationTypes:
    - AAD
  aadAuthentication:
    audience: 41b23e61-6c1e-4545-b367-cd054e0ed4b4
    issuer: https://sts.windows.net/f47ac10b-58cc-4372-a567-0e02b2c3d479/
    tenant: https://login.microsoftonline.com/f47ac10b-58cc-4372-a567-0e02b2c3d479
  vpnProtocols:
    - OpenVPN
```

```shell
planton apply -f azure-vpn-server-configuration.yaml
```

This creates an Entra ID-only policy offering OpenVPN -- the audience is the Microsoft-published Azure VPN Client application ID, and the issuer and tenant URLs embed your directory (tenant) ID. The configuration is free and provisions in seconds. An Infra Job tracks the provisioning in real time.

### InfraChart

In a remote-access chart the order is: WAN → hub → **server configuration** → point-to-site gateway. Wire the resource group with ValueFromRef; the gateway then wires to this configuration the same way:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: network-rg
      fieldPath: status.outputs.resource_group_name
```

The InfraPipeline resolves the dependency graph, deploys the resource group first, then creates the configuration -- and the point-to-site gateway that consumes it references `vpn_server_configuration_id`.

## Key Configuration

These are the most important decisions when configuring the policy. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Authentication types** -- the central choice. "AAD" (Entra ID) gives managed sign-in with conditional access; "Certificate" works offline against your own root; "Radius" reuses existing enterprise auth. Each enabled type requires its block -- the spec rejects a type without its parameters upfront.

**Tunnel protocols** -- leave `vpnProtocols` empty for Azure's default. Set `["OpenVPN"]` when you use policy groups or plan multiple gateway address pools -- both require OpenVPN.

**Policy groups** -- named member-matching rules (Entra ID group ID, certificate common name, RADIUS group ID). The gateway maps groups to address pools; each group's ARM ID surfaces in `policy_group_ids` keyed by the group's name.

**Revocation without rotation** -- `clientRevokedCertificates` blocks a single lost or compromised client certificate by thumbprint, without rotating the trusted root and reissuing every user's certificate. Because the whole policy updates in place, adding a revocation entry takes effect on the shared configuration without touching any gateway.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `vpn_server_configuration_id` | ARM ID of the configuration | A point-to-site gateway's `vpnServerConfigurationId` |
| `policy_group_ids` | ARM ID of each policy group, keyed by group name | Gateway connection configurations mapping groups to address pools (`status.outputs.policy_group_ids.engineering`) |

The outputs also carry `vpn_server_configuration_name` -- gateways reference the configuration by ARM ID, so the name has no ValueFromRef consumer.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Entra ID remote workforce** -- managed sign-in, no certificate distribution. Start from the **Entra ID Remote Workforce** preset.

**Certificate auth with policy groups** -- offline-capable auth with user segmentation. Start from the **Certificate Auth with Policy Groups** preset.

## Works With

- [**Azure Point-to-Site VPN Gateway**](/infra-catalog/azure-point-to-site-vpn-gateway) -- the hub gateway that attaches this policy
- [**Azure Virtual Hub**](/infra-catalog/azure-virtual-hub) -- where the gateway lives
- [**Azure Virtual WAN**](/infra-catalog/azure-virtual-wan) -- the managed network umbrella

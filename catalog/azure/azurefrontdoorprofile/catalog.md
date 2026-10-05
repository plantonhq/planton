# Azure Front Door Profile

Deploys an Azure Front Door (Standard/Premium) profile -- the top-level container for a global content-delivery and application-acceleration deployment on Microsoft's edge network. The profile owns the SKU tier, the origin response timeout, the managed identity, access-log scrubbing, and tags; the delivery surface (endpoints, origin groups, origins, routes) composes from standalone Infra Components that reference it. The SKU tier is fixed at creation -- changing it replaces the profile and every satellite nested under it -- so choose Standard or Premium deliberately up front.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Front Door Profile** -- a global profile with the specified SKU tier (Standard or Premium) and response timeout configuration; Azure deploys it across all edge locations worldwide (no region)
- **Managed Identity Assignment** (optional) -- a system-assigned and/or user-assigned Microsoft Entra identity on the profile, used to read customer-managed TLS certificates from Key Vault for custom domains
- **Access-Log Scrubbing Rules** (optional) -- masking of the selected request parts (query-string arguments, client IP, request URI) in every access-log entry the profile writes
- **Azure Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) merged with user tags and applied to the profile

The delivery surface is deliberately NOT created here. Each piece is its own Infra Component referencing this profile, mirroring Azure's own ARM child-resource model:

- **AzureFrontDoorEndpoint** -- the public entry hostname (`*.azurefd.net`)
- **AzureFrontDoorOriginGroup** -- a load-balanced backend pool with health probing
- **AzureFrontDoorOrigin** -- one backend inside an origin group
- **AzureFrontDoorRoute** -- connects an endpoint to an origin group by URL pattern

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** where the Front Door profile will be created. Front Door is a global resource (no region), but every ARM resource belongs to a resource group. Provide the name directly or reference an AzureResourceGroup Infra Component via ValueFromRef.
- **A user-assigned managed identity** (optional) -- only when custom domains will carry bring-your-own Key Vault certificates and the identity should exist (with its Key Vault grants) before the profile does. A system-assigned identity is created with the profile instead.

## Deploy

### Console

Open the deployment store, find **Azure Front Door Profile**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Standard Delivery** preset for a global CDN anchor, **Premium with Private Origins** for locked-down backends, or **Compliance Log Scrubbing** for privacy-sensitive log handling in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureFrontDoorProfile
metadata:
  name: cdn-profile
  org: acme-corp
  env: prod
spec:
  resourceGroup:
    value: "acme-prod-rg"
  profileName: acme-edge
```

```shell
planton apply -f front-door-profile.yaml
```

This creates a Front Door profile on Azure's default tier (STANDARD -- the sku field records no opinion when omitted) with the default 120-second response timeout, no managed identity, and scrubbing disabled. Endpoints, origin groups, origins, and routes are added as their own Infra Components referencing the profile's `profile_id` output. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the Front Door profile to a resource group deployed in the same InfraPipeline:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: production-rg
      fieldPath: status.outputs.resource_group_name
  profileName: acme-edge
```

The InfraPipeline resolves the dependency graph, deploys the resource group first, then provisions the Front Door profile with the resolved value.

## Key Configuration

These are the most important decisions when configuring a Front Door profile. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**SKU tier** -- Leaving `sku` unset deploys STANDARD (global load balancing, SSL offloading, edge caching, compression, URL routing at 99.99% SLA). PREMIUM adds Private Link to origins, the managed WAF rule sets (Microsoft_DefaultRuleSet, Bot Manager), and JS challenge/CAPTCHA actions. The tier is fixed at creation -- changing it replaces the profile and every satellite nested under it, and Azure refuses a PREMIUM-to-STANDARD downgrade outright.

**Profile name** -- the ARM identity every satellite nests under. Renaming replaces the profile AND everything nested under it, including every endpoint's public hostname. 2-90 characters; letters, digits, and hyphens.

**Response timeout** -- `responseTimeoutSeconds` (16-240; Azure's default 120 applies when omitted) controls how long the edge waits for an origin response before returning a 504. Decrease for latency-sensitive APIs that should fail over quickly; increase for slow backends or large file downloads.

**Managed identity** -- assign one when custom domains will carry bring-your-own certificates: Front Door reads the Key Vault certificate (via AzureFrontDoorSecret) with this identity, keylessly. User-assigned identities can be granted Key Vault access before the profile exists; a system-assigned principal surfaces in the outputs for post-deploy grants.

**Log scrubbing** -- `logScrubbingVariables` masks the selected request parts in every access-log entry before it is written (client IP for GDPR postures, query strings for tokens riding URLs, request URIs when paths carry identifiers). Azure scrubs ALL values of each selected part.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureUserAssignedIdentity** | `identity.userAssignedIdentityIds[]` | `status.outputs.identity_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `profile_id` | Azure resource ID of the Front Door profile | The join seam: AzureFrontDoorEndpoint and AzureFrontDoorOriginGroup reference it; security policies and diagnostic settings target it |
| `profile_name` | Name of the Front Door profile | Cross-reference in the Azure portal, CLI scripts |
| `resource_guid` | Azure-assigned immutable GUID for the Front Door service | Monitoring, diagnostics |
| `identity_principal_id` | Object ID of the system-assigned identity (when one exists) | The grant target for Key Vault access (AzureRoleAssignment) so Front Door can read bring-your-own certificates |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Standard delivery anchor** -- a Standard-tier profile as the small, stable container; endpoints, origin groups, origins, and routes compose against it as their own resources, added and removed without touching the profile. Start from the **Standard Delivery** preset.

**Premium with private origins** -- a Premium-tier profile with a system-assigned managed identity, for locked-down architectures where backends disable public access entirely and Front Door reaches them over Private Link (configured on each AzureFrontDoorOrigin). Start from the **Premium with Private Origins** preset.

**Compliance log scrubbing** -- a Standard profile with access-log scrubbing turned all the way up: client IPs, request URIs, and query-string arguments masked before logs are written. Start from the **Compliance Log Scrubbing** preset.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group where the Front Door profile is created
- [**Azure User Assigned Identity**](/infra-catalog/azure-user-assigned-identity) -- pre-created, pre-granted identities for reading Key Vault certificates
- [**Azure Front Door Endpoint**](/infra-catalog/azure-front-door-endpoint) -- the public entry hostname referencing this profile
- [**Azure Front Door Origin Group**](/infra-catalog/azure-front-door-origin-group) -- a load-balanced backend pool referencing this profile
- [**Azure Front Door Origin**](/infra-catalog/azure-front-door-origin) -- one backend inside an origin group
- [**Azure Front Door Route**](/infra-catalog/azure-front-door-route) -- connects an endpoint to an origin group by URL pattern

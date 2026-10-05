# Azure Container Registry

Deploys an Azure Container Registry (ACR): the managed, private OCI registry that stores the container images and artifacts a platform's workloads run. The SKU is the registry's feature gate -- Basic and Standard carry the core push/pull surface, while Premium unlocks the enterprise surface: geo-replication, zone redundancy, network isolation, quarantine/retention policies, and customer-managed-key encryption. Spec-level validation enforces the same SKU gates ARM does, so a misconfigured manifest fails at validation, not at apply.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Container Registry** -- an Azure Container Registry in the specified region and resource group, with the chosen SKU tier, access posture (admin account, anonymous pull, public network access, export policy), and trusted-services bypass
- **Geo-Replications** -- created only on the Premium SKU when `georeplications` entries are configured; each replica serves pulls locally in its region, with per-replica zone redundancy, an explicit global-endpoint-routing choice, and its own tags
- **Network Rule Set** -- created only on the Premium SKU when `networkRuleSet` is configured; a default action plus an IPv4 CIDR allowlist for a public-but-restricted registry
- **Customer-Managed-Key Encryption** -- created only on the Premium SKU when `encryption` is configured; the registry's data is encrypted with a Key Vault key you own, unwrapped by a user-assigned identity
- **Azure Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** where the Container Registry will be created. Provide the name directly or reference an AzureResourceGroup Infra Component via ValueFromRef.
- **A globally unique registry name** (5-50 characters, lowercase letters and numbers only -- no hyphens, no uppercase). The name becomes the login server hostname: `{name}.azurecr.io`. Renaming replaces the registry and its images do not migrate.
- **For CMK encryption** (Premium): an AzureUserAssignedIdentity holding get/wrapKey/unwrapKey on the key's vault, and an AzureKeyVaultKey -- both must exist before the registry.

## Deploy

### Console

Open the deployment store, find **Azure Container Registry**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Standard Container Registry** preset in the [Presets](#presets) tab for a single-region registry with Standard tier.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureContainerRegistry
metadata:
  name: acme-registry
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    value: "acme-prod-rg"
  registryName: acmeprodregistry
  sku: STANDARD
```

```shell
planton apply -f container-registry.yaml
```

This creates a Standard-tier Container Registry with the admin account disabled and no geo-replication; authentication is handled via Microsoft Entra (service principals, managed identities, repo-scoped tokens). An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the registry to a resource group:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: production-rg
      fieldPath: status.outputs.resource_group_name
```

The InfraPipeline resolves the dependency graph, deploys the resource group first, then provisions the Container Registry with the resolved value.

## Key Configuration

These are the most important decisions when configuring a Container Registry. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**SKU tier** -- `sku` is the registry's feature gate. Unspecified deploys `STANDARD`, the production baseline (100 GiB included storage, webhooks, anonymous pull available). `BASIC` is the cost-optimized dev/test tier with the full API surface but low storage and throughput limits. `PREMIUM` adds the enterprise surface gated throughout the spec (geo-replication, zone redundancy, network isolation, artifact policies, CMK) plus the highest throughput and 500 GiB included storage. Upgrades apply in place; downgrading requires the Premium-only features to be unset first.

**Access posture** -- `adminUserEnabled` (off by default; Entra is the production path), `anonymousPullEnabled` (Standard+; makes EVERY repository publicly readable), `publicNetworkAccessEnabled` (default true; explicit false takes the registry private-endpoints-only and requires Premium), and `exportPolicyEnabled` (default true; disabling is a data-exfiltration control requiring Premium plus public access explicitly false).

**Geo-replication** -- `georeplications` replicates images to additional Azure regions (Premium only). Each entry carries its own `location`, `zoneRedundancyEnabled`, `globalEndpointRoutingEnabled` (whether the replica participates in global endpoint routing for the registry's login server), and `tags`. The list must not contain the home region -- the home replica is implicit. Pulls are served from the nearest replica; pushes propagate automatically.

**Network rules** -- `networkRuleSet` (Premium only) keeps the registry public but restricted: set `defaultAction: DENY` and enumerate IPv4 CIDR `ipRules`. `networkRuleBypassOption` decides whether trusted Azure services (ACR Tasks, Defender) skip the restrictions; `dataEndpointEnabled` gives blob traffic exact per-region hostnames for egress allowlists.

**Artifact policies** -- `retentionPolicyInDays` purges untagged manifests (0 = immediately; unset = keep forever), `quarantinePolicyEnabled` holds new pushes until scanning clears them. Both Premium-only. Docker Content Trust is deliberately not modeled: the ACR service retired it; sign images with the Notation/ORAS artifact flow instead, which needs no registry-level toggle.

**CMK encryption** -- `encryption` (Premium only, fixed at creation) encrypts with a Key Vault key you own: `keyVaultKeyId` references an AzureKeyVaultKey's versionless ID so rotation propagates, and `identityClientId` names the user-assigned identity (by CLIENT id) that unwraps it. Requires a USER_ASSIGNED flavor in `identity`.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureUserAssignedIdentity** | `identity.identityIds[]` | `status.outputs.identity_id` |
| **AzureUserAssignedIdentity** | `encryption.identityClientId` | `status.outputs.client_id` |
| **AzureKeyVaultKey** | `encryption.keyVaultKeyId` | `status.outputs.versionless_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `container_registry_id` | Azure Resource Manager ID of the registry | AKS registry attachment (AcrPull), AzurePrivateEndpoint target, role assignments, diagnostic settings |
| `container_registry_name` | The registry's name | Scripting and az CLI automation |
| `login_server` | Registry hostname (`{name}.azurecr.io`) | Image references, docker login, Container App and Function App registry configuration |
| `admin_username` | The admin account's username (only when the admin account is enabled) | Static-credential image pulls |
| `admin_password` | The admin account's password (only when the admin account is enabled) | Static-credential image pulls |
| `system_assigned_identity_principal_id` | The system-assigned identity's principal ID (when enabled) | Role assignments granting the registry access to other resources |
| `data_endpoint_host_names` | Dedicated data endpoint hostnames (when enabled) | Egress firewall allowlists |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Standard single-region registry** -- Standard SKU with the admin account disabled and no geo-replication. Suitable for most teams and production workloads in a single region. Authentication via Microsoft Entra. Start from the **Standard Container Registry** preset.

**Premium geo-replicated registry** -- Premium SKU with a zone-redundant home replica, geo-replication, and untagged-manifest retention. Low-latency pulls for multi-region AKS clusters and survival of a regional outage. Start from the **Premium Geo-Replicated Registry** preset.

**Premium network-restricted registry** -- Premium SKU, publicly addressable but DENY-by-default with a CIDR allowlist and dedicated data endpoints. The middle ground between open and fully private. Start from the **Premium Network-Restricted Registry** preset.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group where the Container Registry is created
- [**Azure User Assigned Identity**](/infra-catalog/azure-user-assigned-identity) -- attached to the registry and unwraps the CMK encryption key
- [**Azure Key Vault Key**](/infra-catalog/azure-key-vault-key) -- the customer-managed encryption key (reference its versionless ID so rotation propagates)
- [**Azure AKS Cluster**](/infra-catalog/azure-aks-cluster) -- attaches to the registry by referencing `container_registry_id` for AcrPull image access
- [**Azure Private Endpoint**](/infra-catalog/azure-private-endpoint) -- private connectivity for a registry with public network access disabled

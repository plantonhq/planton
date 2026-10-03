# Azure User Assigned Identity

Deploys an Azure user-assigned managed identity: a standalone Entra ID identity that workloads authenticate as, with no credential to store, rotate, or leak. Unlike a system-assigned identity (born and destroyed with one resource), a user-assigned identity exists independently and can be shared — the same identity can back an AKS cluster's kubelets, a Function App, a Container App, and a VM at once, and it survives all of them. That independent lifecycle is what makes it the right anchor for permissions — grants outlive any single consumer.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **User-Assigned Managed Identity** -- an Entra ID identity in the specified region and resource group, independent of the resources it is assigned to
- **Optional regional isolation** -- when `isolationScope: REGIONAL` is set, token issuance for the identity is restricted to its own region (a data-residency / blast-radius control)
- **Azure Tags** -- your governance tags merged over the Planton-derived resource tags (organization, environment, resource ID); a user tag with the same key wins

The identity is deliberately just the identity. What it may DO is granted through **AzureRoleAssignment** resources referencing its `principal_id` output; who may ACT AS it from outside Azure is declared through **AzureFederatedIdentityCredential** resources referencing its `identity_id` output. Keeping the three concerns as separate composable nodes means grants and trust rules are individually reviewable and removable without touching the identity itself.

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** where the managed identity will be created. Provide the name directly or reference an AzureResourceGroup Infra Component via ValueFromRef.

## Deploy

### Console

Open the deployment store, find **Azure User Assigned Identity**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Standard Managed Identity** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureUserAssignedIdentity
metadata:
  name: ci-deployer
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    value: "acme-prod-rg"
  name: acme-ci-deployer
```

```shell
planton apply -f identity.yaml
```

This creates a user-assigned managed identity with NO permissions -- a freshly-created identity can do nothing until AzureRoleAssignment resources grant it access. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the identity's resource group, then compose the grant and trust nodes against its outputs:

```yaml
# The identity
spec:
  region: eastus
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: production-rg
      fieldPath: status.outputs.resource_group_name
  name: payments-api
---
# A grant (separate AzureRoleAssignment resource)
spec:
  scope:
    valueFrom:
      kind: AzureKeyVault
      name: platform-secrets
      fieldPath: status.outputs.key_vault_id
  roleDefinitionName: "Key Vault Secrets User"
  principalId:
    valueFrom:
      kind: AzureUserAssignedIdentity
      name: payments-api
      fieldPath: status.outputs.principal_id
```

The InfraPipeline resolves the dependency graph, deploys the resource group and identity first, then provisions the grants with the resolved principal ID.

## Key Configuration

These are the most important decisions when configuring a user-assigned managed identity. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Durable naming** -- Changing the name replaces the identity, and with it the PRINCIPAL -- silently invalidating every role assignment and federated credential written against it. Name identities after the workload or duty they represent (`ci-deployer`, `payments-api`), not after any single consumer, because grants outlive consumers.

**Isolation scope** -- By default an identity is usable by resources in any region (its own region only anchors the ARM resource). `REGIONAL` restricts token issuance to the identity's own region -- a data-residency control some regulated environments require. Most deployments leave it unset; it updates in place, so isolation can be adopted later.

**Identity sharing** -- Unlike system-assigned identities (tied to a single resource), user-assigned identities can be attached to multiple Azure resources. Create one identity per logical application or duty, not one per resource, to keep permission auditing to one principal.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `identity_id` | Azure Resource Manager ID of the managed identity | AKS kubelet identity, VM/Function/Container App identity blocks, AzureFederatedIdentityCredential parent |
| `principal_id` | Entra ID service principal object ID | AzureRoleAssignment grants, Key Vault access policies |
| `client_id` | Client ID (application ID) for SDK authentication | Application environment variable (AZURE_CLIENT_ID) |
| `tenant_id` | Entra ID tenant ID | Cross-tenant authentication, SDK configuration |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**The identity/grant/trust triad** -- One AzureUserAssignedIdentity, one or more AzureRoleAssignment resources granting it exactly what it needs, and (for keyless CI or AKS workload identity) AzureFederatedIdentityCredential resources declaring who may act as it. Each node is individually reviewable and removable.

**Keyless CI deployment** -- An identity named for the pipeline (`ci-deployer`), a federated credential trusting the repository's OIDC token, and role assignments scoped to exactly the resources the pipeline deploys. No stored service-principal secret anywhere. Start from the **CI Deployer Identity** preset.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group where the managed identity is created
- [**Azure Role Assignment**](/infra-catalog/azure-role-assignment) -- grants the identity permissions by targeting its `principal_id` output
- [**Azure Federated Identity Credential**](/infra-catalog/azure-federated-identity-credential) -- declares keyless trust rules on the identity's `identity_id` output

# AzureUserAssignedIdentity: Research & Design Documentation

## 1. What Is Azure User-Assigned Managed Identity?

Azure User-Assigned Managed Identity is an Azure AD identity that exists as a standalone Azure resource. It provides a secure, credential-free way for Azure resources to authenticate to any Azure service that supports Azure AD authentication. The identity is managed by Azure -- there are no passwords, certificates, or keys to rotate.

### Managed Identity Types

Azure offers two types of managed identities:

**System-Assigned**: Created as part of an Azure resource (e.g., a VM or Function App). The identity shares the lifecycle of the resource -- when the resource is deleted, the identity is automatically cleaned up. Cannot be shared across resources.

**User-Assigned**: Created as a standalone Azure resource with an independent lifecycle. Can be assigned to one or more Azure resources. Persists even if all assigned resources are deleted. This is the type modeled by this component.

### Why User-Assigned Over System-Assigned?

User-assigned identities are preferred for enterprise deployments because:

1. **Independent lifecycle** -- The identity survives resource recreation, deployment slot swaps, and blue/green deployments
2. **Shared across resources** -- One identity can serve multiple Function Apps, Web Apps, or AKS clusters that need the same permissions
3. **Pre-provisioned permissions** -- Permissions can be set up before the resources that use them exist, enabling GitOps workflows
4. **Auditable** -- A standalone resource in Azure Resource Manager with its own audit trail and tagging

### How Authentication Works

When an Azure resource (AKS, Function App, Web App, VM) is configured with a user-assigned identity:

1. The resource's metadata in Azure Resource Manager includes the identity's resource ID
2. When the application code requests a token from the Azure Instance Metadata Service (IMDS), it specifies the identity's client ID
3. Azure AD issues an OAuth 2.0 access token for the requested resource (Key Vault, Storage, etc.)
4. The application uses the token to authenticate to the target service
5. No credentials are stored in code, configuration, or environment variables

This zero-credential model eliminates an entire class of security vulnerabilities (credential leaks, rotation failures, hardcoded secrets).

## 2. Deployment Landscape

### Level 0: Azure Portal

The portal provides a guided experience for creating managed identities and assigning roles. The "Identity" blade exists on most Azure resources to configure system-assigned identities.

### Level 1: Azure CLI

```bash
# Create the identity
az identity create \
  --name my-identity \
  --resource-group my-rg \
  --location eastus

# Get the principal ID
PRINCIPAL_ID=$(az identity show --name my-identity --resource-group my-rg --query principalId -o tsv)

# Assign a role
az role assignment create \
  --assignee $PRINCIPAL_ID \
  --role "Key Vault Secrets User" \
  --scope /subscriptions/.../providers/Microsoft.KeyVault/vaults/my-kv
```

### Level 2: ARM Templates / Bicep

```bicep
resource managedIdentity 'Microsoft.ManagedIdentity/userAssignedIdentities@2023-01-31' = {
  name: 'my-identity'
  location: 'eastus'
}

resource roleAssignment 'Microsoft.Authorization/roleAssignments@2022-04-01' = {
  name: guid(managedIdentity.id, keyVault.id, 'Key Vault Secrets User')
  scope: keyVault
  properties: {
    principalId: managedIdentity.properties.principalId
    roleDefinitionId: subscriptionResourceId('Microsoft.Authorization/roleDefinitions', '4633458b-17de-408a-b874-0445c86b69e6')
    principalType: 'ServicePrincipal'
  }
}
```

### Level 3: Terraform

```hcl
resource "azurerm_user_assigned_identity" "main" {
  name                = "my-identity"
  location            = "eastus"
  resource_group_name = "my-rg"
}

resource "azurerm_role_assignment" "kv" {
  scope                            = azurerm_key_vault.main.id
  role_definition_name             = "Key Vault Secrets User"
  principal_id                     = azurerm_user_assigned_identity.main.principal_id
  skip_service_principal_aad_check = true
}
```

### Level 4: Pulumi

```go
identity, _ := authorization.NewUserAssignedIdentity(ctx, "my-identity",
  &authorization.UserAssignedIdentityArgs{
    Name:              pulumi.String("my-identity"),
    Location:          pulumi.String("eastus"),
    ResourceGroupName: pulumi.String("my-rg"),
  })

authorization.NewAssignment(ctx, "kv-role",
  &authorization.AssignmentArgs{
    PrincipalId:                 identity.PrincipalId,
    Scope:                       keyVault.ID(),
    RoleDefinitionName:          pulumi.String("Key Vault Secrets User"),
    SkipServicePrincipalAadCheck: pulumi.Bool(true),
  })
```

### Level 5: OpenMCF (This Component)

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureUserAssignedIdentity
metadata:
  name: my-identity
spec:
  region: eastus
  resource_group: my-rg
  name: my-identity
  role_assignments:
    - scope:
        valueFrom:
          kind: AzureKeyVault
          name: my-kv
          fieldPath: status.outputs.key_vault_id
      role_definition_name: Key Vault Secrets User
```

## 3. Cross-Cloud Identity Comparison

### AWS IAM Role

AWS uses trust policies to define WHO can assume a role, and permission policies (managed/inline) to define WHAT the role can do. The scope is implicit in the policy resource ARNs.

**OpenMCF equivalent**: `AwsIamRole` with `trust_policy`, `managed_policy_arns`, and `inline_policies`.

### GCP Service Account

GCP uses service accounts with project-level or organization-level IAM bindings. Roles are bound at the project or org scope.

**OpenMCF equivalent**: `GcpServiceAccount` with `project_iam_roles` and `org_iam_roles`.

### Azure User-Assigned Managed Identity

Azure uses managed identities with per-resource role assignments. Each assignment specifies a scope (specific resource, resource group, or subscription) and a role definition.

**OpenMCF equivalent**: `AzureUserAssignedIdentity` with `role_assignments[].scope` and `role_assignments[].role_definition_name`.

### Key Differences

| Aspect | AWS | GCP | Azure |
|--------|-----|-----|-------|
| Identity model | IAM Role (assumed) | Service Account (impersonated) | Managed Identity (assigned) |
| Credential-free | Via STS AssumeRole | Via Workload Identity | Built-in (IMDS) |
| Scope granularity | Policy document ARNs | Project or Organization | Any Azure resource ID |
| Sharing model | Trust policy principals | IAM binding members | Identity attached to N resources |
| Key management | Optional access keys | Optional JSON keys | No keys (credential-free only) |

## 4. Design Decisions

### Why Bundle Role Assignments (DD03)

An identity without role assignments has no permissions -- it cannot access any Azure resource. Bundling role assignments with the identity ensures the component is functionally complete. Users don't have to manage a separate "role assignment" resource kind.

This mirrors how `AwsIamRole` bundles managed/inline policies and `GcpServiceAccount` bundles project/org IAM roles.

### Why StringValueOrRef for Scope

Azure role assignment scopes reference specific resource IDs. In infra charts, these resources are often created in the same chart. Making `scope` a `StringValueOrRef` enables:

```yaml
role_assignments:
  - scope:
      valueFrom:
        kind: AzureKeyVault
        name: platform-kv
        fieldPath: status.outputs.key_vault_id
    role_definition_name: Key Vault Secrets User
```

Without StringValueOrRef, users would need to hardcode Azure resource IDs, breaking the composability that infra charts depend on.

Note: `scope` has no `default_kind` annotation because it is polymorphic -- it can reference any Azure resource type.

### Why role_definition_name Over role_definition_id

Azure RBAC supports both role definition names (human-readable) and IDs (GUIDs). We chose `role_definition_name` because:

1. Names are readable: "Key Vault Secrets User" vs "4633458b-17de-408a-b874-0445c86b69e6"
2. Names are portable across subscriptions (GUIDs vary by subscription for custom roles)
3. The Terraform and Pulumi providers both support name-based lookup

### What's Excluded (80/20)

- **Federated identity credentials**: Used for OIDC-based authentication (GitHub Actions, Kubernetes pod identity). Important but complex -- a future resource or enhancement.
- **Isolation scope**: Regional isolation for managed identities. Niche enterprise feature.
- **Role assignment conditions (ABAC)**: Attribute-based access control conditions on role assignments. Advanced scenario.
- **Role definition ID**: Supported by Azure alongside name, but adds complexity for marginal value.

## 5. Azure AD Replication Considerations

Azure AD is eventually consistent. When a managed identity is created, there is a propagation delay (typically seconds, occasionally minutes) before the identity's service principal is visible across all Azure AD replicas.

This matters for role assignments: if you try to create a role assignment immediately after creating the identity, Azure may return a 403 because the principal doesn't exist yet.

### Mitigation

Both IaC implementations set `skip_service_principal_aad_check = true` (Terraform) / `SkipServicePrincipalAadCheck: true` (Pulumi) on all role assignments. This tells Azure to bypass the AAD existence check and proceed with the assignment. Azure will eventually replicate the principal and the assignment will become effective.

The Pulumi module also adds an explicit `DependsOn` on the identity resource to ensure Pulumi doesn't attempt the role assignment before the identity creation API call completes.

## 6. Common Built-In Roles Reference

| Role Name | Description | Common Use Case |
|-----------|-------------|-----------------|
| Contributor | Full resource management except RBAC | Broad access (avoid in production) |
| Reader | Read-only access | Monitoring and compliance |
| Key Vault Secrets User | Read Key Vault secrets | Application secret access |
| Key Vault Crypto User | Key Vault cryptographic operations | Encryption/decryption |
| Storage Blob Data Contributor | Read/write/delete blob data | Application storage |
| Storage Blob Data Reader | Read blob data | Read-only storage access |
| AcrPull | Pull container images from ACR | AKS, Container Apps |
| AcrPush | Push container images to ACR | CI/CD pipelines |
| Network Contributor | Manage networking resources | AKS network identity |
| Monitoring Metrics Publisher | Publish monitoring metrics | Application monitoring |
| Cosmos DB Account Reader | Read Cosmos DB account metadata | Application data access |

## 7. Infra Chart Patterns

### AKS Environment (Enhanced)

```
AzureResourceGroup (Layer 0)
  └── AzureUserAssignedIdentity (Layer 1)
        └── role: Network Contributor on AzureVpc
        └── role: AcrPull on AzureContainerRegistry
  └── AzureVpc (Layer 1)
  └── AzureContainerRegistry (Layer 1)
  └── AzureAksCluster (Layer 2, uses identity_id)
```

### Function App Environment

```
AzureResourceGroup (Layer 0)
  └── AzureUserAssignedIdentity (Layer 1)
        └── role: Key Vault Secrets User on AzureKeyVault
  └── AzureKeyVault (Layer 1)
  └── AzureServicePlan (Layer 1)
  └── AzureFunctionApp (Layer 2, uses identity_id)
```

### Web App Environment

```
AzureResourceGroup (Layer 0)
  └── AzureUserAssignedIdentity (Layer 1)
        └── role: Key Vault Secrets User on AzureKeyVault
        └── role: Storage Blob Data Contributor on AzureStorageAccount
  └── AzureKeyVault (Layer 1)
  └── AzureStorageAccount (Layer 1)
  └── AzureServicePlan (Layer 1)
  └── AzureLinuxWebApp (Layer 2, uses identity_id)
```

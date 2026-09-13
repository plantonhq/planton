# AzureUserAssignedIdentity

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**AzureUserAssignedIdentitySpec** defines the configuration for creating an
Azure user-assigned managed identity: a standalone Azure AD identity that
workloads authenticate as, with no credential to store, rotate, or leak.

Unlike a system-assigned identity (born and destroyed with one resource), a
user-assigned identity exists independently and can be shared: the same
identity can back an AKS cluster's kubelets, a Function App, a Container
App, and a VM at once, and it survives all of them. That independent
lifecycle is what makes it the right anchor for permissions -- grants
outlive any single consumer.

The identity is deliberately just the identity. What it may DO is granted
through AzureRoleAssignment resources referencing its principal_id output;
who may ACT AS it from outside Azure is declared through
AzureFederatedIdentityCredential resources referencing its identity_id
output. Keeping the three concerns as separate composable nodes means
grants and trust rules are individually reviewable and removable without
touching the identity itself.

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureUserAssignedIdentity
metadata:
  name: test-identity
  org: test-org
  env: dev
spec:
  region: eastus
  resourceGroup:
    value: test-rg
  name: test-managed-identity
  tags:
    cost-center: platform
    owner: identity-team
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.name` | `string` | yes |  |  |
| `spec.isolationScope` | `enum` |  |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.region

`string` · required

The Azure region where the managed identity will be created, e.g.
"eastus", "westeurope". The identity is a regional resource; changing
the region replaces it. Note the identity's usability is NOT limited to
its own region (see isolation_scope for the opt-in regional isolation
behavior).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.resourceGroup

`string | valueFrom` · required

The Azure resource group the managed identity will be created in.
Can be a literal resource-group name or a reference to an
AzureResourceGroup's name output.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.name

`string` · required

The name of the user-assigned managed identity. Must be unique within
the resource group; 3-128 characters (alphanumeric, hyphens, and
underscores). Changing the name replaces the identity -- and with it the
principal, invalidating existing grants -- so name it durably, after the
workload or duty it represents ("ci-deployer", "payments-api").

- rule: {"required":true,"string":{"minLen":"3","maxLen":"128"}}

### spec.isolationScope

`enum`

Opt-in regional isolation. By default (unspecified) an identity is
usable by resources in any region. REGIONAL restricts token issuance for
the identity to its own region -- a data-residency / blast-radius control
some regulated environments require. Most deployments leave this unset.
Updatable in place.

Allowed values (use exactly as shown):

- `azure_user_assigned_identity_isolation_scope_unspecified` -- Not specified: ARM's default (no isolation) -- the identity is usable by resources in any region.
- `REGIONAL` -- Token issuance for the identity is restricted to the identity's own region.

### spec.tags

`map<string, string>`

Free-form tags applied to the identity, merged over the Planton-derived
resource tags (organization, environment, resource id); a user tag with
the same key wins. Tags are Azure's governance surface -- Azure Policy
enforces them and Microsoft Cost Management groups by them -- so carry
your org's ownership/cost-center conventions here. Updatable in place.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureUserAssignedIdentity, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.identity_id` | `string` | The Azure Resource Manager ID of the User-Assigned Managed Identity. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.ManagedIdentity/userAssignedIdentities/{name} Referenced by downstream resources that accept a user-assigned identity ID. |
| `status.outputs.principal_id` | `string` | The Service Principal Object ID associated with this Managed Identity. This is the object ID in Azure AD, used internally for RBAC role assignments. Also referenced by resources that need the principal ID for access policies (e.g., Key Vault access policies that accept object IDs). |
| `status.outputs.client_id` | `string` | The Client ID (Application ID) of the Managed Identity. Used by applications to authenticate as this identity via the Azure SDK. Configured as an environment variable (AZURE_CLIENT_ID) or in SDK configuration to specify which identity to use when running on Azure services. |
| `status.outputs.tenant_id` | `string` | The Azure AD Tenant ID that the Managed Identity belongs to. Used in conjunction with client_id for cross-tenant or multi-tenant scenarios. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureAiFoundry | `spec.identity.identityIds` | `status.outputs.identity_id` |
| AzureAiFoundry | `spec.primaryUserAssignedIdentity` | `status.outputs.identity_id` |
| AzureAiFoundry | `spec.encryption.userAssignedIdentityId` | `status.outputs.identity_id` |
| AzureAiFoundryProject | `spec.identity.identityIds` | `status.outputs.identity_id` |
| AzureAiFoundryProject | `spec.primaryUserAssignedIdentity` | `status.outputs.identity_id` |
| AzureAksCluster | `spec.identity.identityIds` | `status.outputs.identity_id` |
| AzureAksCluster | `spec.kubeletIdentity.userAssignedIdentityId` | `status.outputs.identity_id` |
| AzureApplicationGateway | `spec.identity.identityIds` | `status.outputs.identity_id` |
| AzureCognitiveAccount | `spec.identity.identityIds` | `status.outputs.identity_id` |
| AzureCognitiveAccountProject | `spec.identity.identityIds` | `status.outputs.identity_id` |
| AzureContainerApp | `spec.customScaleRules[].identityId` | `status.outputs.identity_id` |
| AzureContainerApp | `spec.identity.userAssignedIdentityIds` | `status.outputs.identity_id` |
| AzureContainerAppEnvironment | `spec.identity.userAssignedIdentityIds` | `status.outputs.identity_id` |
| AzureContainerAppEnvironmentCertificate | `spec.certificateKeyVault.identity` | `status.outputs.identity_id` |
| AzureContainerAppJob | `spec.eventTrigger.scale.rules[].identityId` | `status.outputs.identity_id` |
| AzureContainerAppJob | `spec.identity.userAssignedIdentityIds` | `status.outputs.identity_id` |
| AzureContainerInstance | `spec.imageRegistryCredentials[].userAssignedIdentityId` | `status.outputs.identity_id` |
| AzureContainerInstance | `spec.identity.identityIds` | `status.outputs.identity_id` |
| AzureContainerInstance | `spec.keyVaultUserAssignedIdentityId` | `status.outputs.identity_id` |
| AzureContainerRegistry | `spec.identity.identityIds` | `status.outputs.identity_id` |
| AzureContainerRegistry | `spec.encryption.identityClientId` | `status.outputs.client_id` |
| AzureCosmosdbAccount | `spec.identity.identityIds` | `status.outputs.identity_id` |
| AzureCosmosdbAccount | `spec.defaultIdentity.userAssignedIdentityId` | `status.outputs.identity_id` |
| AzureCosmosdbSqlRoleAssignment | `spec.principalId` | `status.outputs.principal_id` |
| AzureDataFactory | `spec.identity.identityIds` | `status.outputs.identity_id` |
| AzureDataFactory | `spec.customerManagedKey.userAssignedIdentityId` | `status.outputs.identity_id` |
| AzureDataFactory | `spec.userManagedIdentityCredentials[].identityId` | `status.outputs.identity_id` |
| AzureDataProtectionBackupVault | `spec.identity.identityIds` | `status.outputs.identity_id` |
| AzureDiskEncryptionSet | `spec.identity.identityIds` | `status.outputs.identity_id` |
| AzureEventHub | `spec.captureDescription.destination.storageAuthenticationId` | `status.outputs.identity_id` |
| AzureEventHubNamespace | `spec.identity.userAssignedIdentityIds` | `status.outputs.identity_id` |
| AzureEventHubNamespaceCustomerManagedKey | `spec.userAssignedIdentityId` | `status.outputs.identity_id` |
| AzureEventgridDomain | `spec.identity.identityIds` | `status.outputs.identity_id` |
| AzureEventgridEventSubscription | `spec.deliveryIdentity.userAssignedIdentity` | `status.outputs.identity_id` |
| AzureEventgridEventSubscription | `spec.deadLetterIdentity.userAssignedIdentity` | `status.outputs.identity_id` |
| AzureEventgridNamespace | `spec.identity.identityIds` | `status.outputs.identity_id` |
| AzureEventgridSystemTopic | `spec.identity.identityIds` | `status.outputs.identity_id` |
| AzureEventgridTopic | `spec.identity.identityIds` | `status.outputs.identity_id` |
| AzureExpressRoutePort | `spec.identity.identityIds` | `status.outputs.identity_id` |
| AzureFederatedIdentityCredential | `spec.userAssignedIdentity` | `status.outputs.identity_id` |
| AzureFirewallPolicy | `spec.identity.userAssignedIdentityIds` | `status.outputs.identity_id` |
| AzureFrontDoorProfile | `spec.identity.userAssignedIdentityIds` | `status.outputs.identity_id` |
| AzureFunctionApp | `spec.identity.identityIds` | `status.outputs.identity_id` |
| AzureFunctionApp | `spec.keyVaultReferenceIdentityId` | `status.outputs.identity_id` |
| AzureFunctionAppFlexConsumption | `spec.storageUserAssignedIdentityId` | `status.outputs.identity_id` |
| AzureFunctionAppFlexConsumption | `spec.identity.identityIds` | `status.outputs.identity_id` |
| AzureKeyVault | `spec.accessPolicies[].objectId` | `status.outputs.principal_id` |
| AzureLinuxWebApp | `spec.identity.identityIds` | `status.outputs.identity_id` |
| AzureLinuxWebApp | `spec.keyVaultReferenceIdentityId` | `status.outputs.identity_id` |
| AzureLogAnalyticsWorkspace | `spec.identity.userAssignedIdentityIds` | `status.outputs.identity_id` |
| AzureMachineLearningBatchEndpoint | `spec.identity.identityIds` | `status.outputs.identity_id` |
| AzureMachineLearningComputeCluster | `spec.identity.identityIds` | `status.outputs.identity_id` |
| AzureMachineLearningComputeInstance | `spec.identity.identityIds` | `status.outputs.identity_id` |
| AzureMachineLearningOnlineEndpoint | `spec.identity.identityIds` | `status.outputs.identity_id` |
| AzureMachineLearningWorkspace | `spec.identity.identityIds` | `status.outputs.identity_id` |
| AzureMachineLearningWorkspace | `spec.primaryUserAssignedIdentity` | `status.outputs.identity_id` |
| AzureMachineLearningWorkspace | `spec.encryption.userAssignedIdentityId` | `status.outputs.identity_id` |
| AzureManagedRedis | `spec.customerManagedKey.userAssignedIdentityId` | `status.outputs.identity_id` |
| AzureManagedRedis | `spec.identity.userAssignedIdentityIds` | `status.outputs.identity_id` |
| AzureManagedRedisAccessPolicyAssignment | `spec.objectId` | `status.outputs.principal_id` |
| AzureMongoCluster | `spec.userAssignedIdentityIds` | `status.outputs.identity_id` |
| AzureMongoCluster | `spec.customerManagedKey.userAssignedIdentityId` | `status.outputs.identity_id` |
| AzureMonitorDataCollectionRule | `spec.identity.identityIds` | `status.outputs.identity_id` |
| AzureMonitorScheduledQueryAlert | `spec.identity.userAssignedIdentityIds` | `status.outputs.identity_id` |
| AzureMssqlDatabase | `spec.userAssignedIdentityIds` | `status.outputs.identity_id` |
| AzureMssqlServer | `spec.azureadAdministrator.objectId` | `status.outputs.principal_id` |
| AzureMssqlServer | `spec.identity.identityIds` | `status.outputs.identity_id` |
| AzureMssqlServer | `spec.primaryUserAssignedIdentityId` | `status.outputs.identity_id` |
| AzureMysqlFlexibleServer | `spec.userAssignedIdentityIds` | `status.outputs.identity_id` |
| AzureMysqlFlexibleServer | `spec.customerManagedKey.primaryUserAssignedIdentityId` | `status.outputs.identity_id` |
| AzureMysqlFlexibleServer | `spec.customerManagedKey.geoBackupUserAssignedIdentityId` | `status.outputs.identity_id` |
| AzureMysqlFlexibleServer | `spec.aadAdministrator.identityId` | `status.outputs.identity_id` |
| AzureMysqlFlexibleServer | `spec.aadAdministrator.objectId` | `status.outputs.client_id` |
| AzurePostgresqlFlexibleServer | `spec.aadAdministrators[].objectId` | `status.outputs.principal_id` |
| AzurePostgresqlFlexibleServer | `spec.identity.identityIds` | `status.outputs.identity_id` |
| AzurePostgresqlFlexibleServer | `spec.customerManagedKey.primaryUserAssignedIdentityId` | `status.outputs.identity_id` |
| AzurePostgresqlFlexibleServer | `spec.customerManagedKey.geoBackupUserAssignedIdentityId` | `status.outputs.identity_id` |
| AzureRecoveryServicesVault | `spec.identity.identityIds` | `status.outputs.identity_id` |
| AzureRecoveryServicesVault | `spec.encryption.userAssignedIdentityId` | `status.outputs.identity_id` |
| AzureRedisCache | `spec.identity.userAssignedIdentityIds` | `status.outputs.identity_id` |
| AzureRedisCacheAccessPolicyAssignment | `spec.objectId` | `status.outputs.principal_id` |
| AzureRoleAssignment | `spec.principalId` | `status.outputs.principal_id` |
| AzureSearchService | `spec.identity.identityIds` | `status.outputs.identity_id` |
| AzureServiceBusNamespace | `spec.identity.userAssignedIdentityIds` | `status.outputs.identity_id` |
| AzureServiceBusNamespace | `spec.customerManagedKey.userAssignedIdentityId` | `status.outputs.identity_id` |
| AzureStorageAccount | `spec.identity.identityIds` | `status.outputs.identity_id` |
| AzureStorageAccount | `spec.customerManagedKey.userAssignedIdentityId` | `status.outputs.identity_id` |
| AzureStorageDataLakeGen2Filesystem | `spec.owner` | `status.outputs.principal_id` |
| AzureStorageDataLakeGen2Filesystem | `spec.group` | `status.outputs.principal_id` |
| AzureStorageDataLakeGen2Filesystem | `spec.aces[].objectId` | `status.outputs.principal_id` |
| AzureVirtualMachine | `spec.identity.identityIds` | `status.outputs.identity_id` |
| AzureVirtualMachineScaleSet | `spec.identity.identityIds` | `status.outputs.identity_id` |
| KubernetesCertManager | `spec.workloadIdentity.aks.clientId` | `status.outputs.client_id` |
| KubernetesExternalDns | `spec.workloadIdentity.aks.clientId` | `status.outputs.client_id` |
| KubernetesExternalSecretsOperator | `spec.workloadIdentity.aks.clientId` | `status.outputs.client_id` |
| KubernetesOpenBao | `spec.backup.workloadIdentity.aks.clientId` | `status.outputs.client_id` |
| KubernetesPostgres | `spec.workloadIdentity.aks.clientId` | `status.outputs.client_id` |
| KubernetesServiceAccount | `spec.workloadIdentity.aks.clientId` | `status.outputs.client_id` |
| KubernetesVelero | `spec.backupStorage.azureBlob.workloadIdentityClientId` | `status.outputs.client_id` |

## See Also

- [Overview](../README.md)

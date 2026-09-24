# AzureDataFactory

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**AzureDataFactorySpec** defines an Azure Data Factory -- the
factory is the workspace every other Data Factory resource lives
inside: pipelines (AzureDataFactoryPipeline), data flows, linked
services, datasets, triggers, and integration runtimes are all
created against a factory's ARM ID.

The factory itself carries the workspace-level posture: its managed
identity, an optional git repository binding (GitHub or Azure
DevOps -- at most one), workspace-wide global parameters,
customer-managed-key encryption, named credentials its linked
services authenticate with, a managed virtual network, and managed
private endpoints for private egress from that network.

## Example

```yaml
# Deep-shape example for docs and offline validation: a factory with
# the combined identity, a GitHub repository binding, global
# parameters across the type vocabulary, customer-managed-key
# encryption by reference form, both credential flavors, the managed
# virtual network, and a managed private endpoint on each arm
# (subresource and Private Link Service fqdns). References are literal
# values so the manifest validates standalone.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureDataFactory
metadata:
  name: test-data-factory
  id: test-data-factory
  org: test-org
  env: test
spec:
  resourceGroup:
    value: test-rg
  name: test-org-data-factory
  region: eastus
  identity:
    type: SYSTEM_AND_USER_ASSIGNED
    identityIds:
      - value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/test-uai
  githubConfiguration:
    accountName: test-org
    branchName: main
    repositoryName: data-pipelines
    rootFolder: /
    publishingEnabled: true
  globalParameters:
    - name: environment
      type: String
      value: test
    - name: retries
      type: Int
      value: "3"
    - name: regions
      type: Array
      value: '["eastus","westus"]'
  managedVirtualNetworkEnabled: true
  publicNetworkEnabled: false
  customerManagedKey:
    keyVaultKeyId:
      value: https://test-vault.vault.azure.net/keys/df-cmk
    userAssignedIdentityId:
      value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/test-uai
  userManagedIdentityCredentials:
    - name: etl-identity
      identityId:
        value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/test-uai
      description: identity the ETL linked services run as
  servicePrincipalCredentials:
    - name: legacy-sp
      tenantId: 11111111-2222-3333-4444-555555555555
      servicePrincipalId: 22222222-3333-4444-5555-666666666666
      servicePrincipalKey:
        linkedServiceName: keyvault-ls
        secretName: legacy-sp-key
  managedPrivateEndpoints:
    - name: datalake-blob
      targetResourceId:
        value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Storage/storageAccounts/testdata
      subresourceName: blob
    - name: partner-pls
      targetResourceId:
        value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/privateLinkServices/partner-pls
      fqdns:
        - internal.partner.example
  tags:
    costCenter: platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.name` | `string` | yes |  |  |
| `spec.region` | `string` | yes |  |  |
| `spec.identity` | `AzureDataFactoryIdentity` |  |  |  |
| `spec.identity.type` | `enum` | yes |  |  |
| `spec.identity.identityIds` | `[]string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.githubConfiguration` | `AzureDataFactoryGithubConfiguration` |  |  |  |
| `spec.githubConfiguration.accountName` | `string` | yes |  |  |
| `spec.githubConfiguration.branchName` | `string` | yes |  |  |
| `spec.githubConfiguration.gitUrl` | `string` |  |  |  |
| `spec.githubConfiguration.repositoryName` | `string` | yes |  |  |
| `spec.githubConfiguration.rootFolder` | `string` | yes |  |  |
| `spec.githubConfiguration.publishingEnabled` | `bool` |  | `true` |  |
| `spec.vstsConfiguration` | `AzureDataFactoryVstsConfiguration` |  |  |  |
| `spec.vstsConfiguration.accountName` | `string` | yes |  |  |
| `spec.vstsConfiguration.branchName` | `string` | yes |  |  |
| `spec.vstsConfiguration.projectName` | `string` | yes |  |  |
| `spec.vstsConfiguration.repositoryName` | `string` | yes |  |  |
| `spec.vstsConfiguration.rootFolder` | `string` | yes |  |  |
| `spec.vstsConfiguration.tenantId` | `string` | yes |  |  |
| `spec.vstsConfiguration.publishingEnabled` | `bool` |  | `true` |  |
| `spec.globalParameters` | `[]AzureDataFactoryGlobalParameter` |  |  |  |
| `spec.globalParameters[].name` | `string` | yes |  |  |
| `spec.globalParameters[].type` | `string` | yes |  |  |
| `spec.globalParameters[].value` | `string` | yes |  |  |
| `spec.managedVirtualNetworkEnabled` | `bool` |  |  |  |
| `spec.publicNetworkEnabled` | `bool` |  | `true` |  |
| `spec.purviewId` | `string \| valueFrom` |  |  |  |
| `spec.customerManagedKey` | `AzureDataFactoryCustomerManagedKey` |  |  |  |
| `spec.customerManagedKey.keyVaultKeyId` | `string \| valueFrom` | yes |  | AzureKeyVaultKey (`status.outputs.versionless_id`) |
| `spec.customerManagedKey.userAssignedIdentityId` | `string \| valueFrom` | yes |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.userManagedIdentityCredentials` | `[]AzureDataFactoryUserManagedIdentityCredential` |  |  |  |
| `spec.userManagedIdentityCredentials[].name` | `string` | yes |  |  |
| `spec.userManagedIdentityCredentials[].identityId` | `string \| valueFrom` | yes |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.userManagedIdentityCredentials[].description` | `string` |  |  |  |
| `spec.userManagedIdentityCredentials[].annotations` | `[]string` |  |  |  |
| `spec.servicePrincipalCredentials` | `[]AzureDataFactoryServicePrincipalCredential` |  |  |  |
| `spec.servicePrincipalCredentials[].name` | `string` | yes |  |  |
| `spec.servicePrincipalCredentials[].tenantId` | `string` | yes |  |  |
| `spec.servicePrincipalCredentials[].servicePrincipalId` | `string` | yes |  |  |
| `spec.servicePrincipalCredentials[].servicePrincipalKey` | `AzureDataFactoryServicePrincipalKey` |  |  |  |
| `spec.servicePrincipalCredentials[].servicePrincipalKey.linkedServiceName` | `string` | yes |  |  |
| `spec.servicePrincipalCredentials[].servicePrincipalKey.secretName` | `string` | yes |  |  |
| `spec.servicePrincipalCredentials[].servicePrincipalKey.secretVersion` | `string` |  |  |  |
| `spec.servicePrincipalCredentials[].description` | `string` |  |  |  |
| `spec.servicePrincipalCredentials[].annotations` | `[]string` |  |  |  |
| `spec.managedPrivateEndpoints` | `[]AzureDataFactoryManagedPrivateEndpoint` |  |  |  |
| `spec.managedPrivateEndpoints[].name` | `string` | yes |  |  |
| `spec.managedPrivateEndpoints[].targetResourceId` | `string \| valueFrom` | yes |  |  |
| `spec.managedPrivateEndpoints[].subresourceName` | `string` |  |  |  |
| `spec.managedPrivateEndpoints[].fqdns` | `[]string` |  |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.resourceGroup

`string | valueFrom` · required

The Azure Resource Group the factory lives in. Can be a literal
string or a reference to an AzureResourceGroup output.

**ForceNew**: changing this destroys and recreates the factory
(and everything inside it).

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.name

`string` · required

The factory's name -- 3-63 characters; letters, numbers, and
hyphens; must start and end with a letter or number (no
consecutive-hyphen segments beyond single separators). Data
Factory names are GLOBALLY unique across Azure (the name becomes
part of the factory's URL), so prefix it with your org -- a taken
name fails at deploy time.

**ForceNew**: changing this destroys and recreates the factory.

- rule: Factory names must be 3-63 characters of letters, numbers, and hyphens, starting and ending with a letter or number
- rule: {"required":true}

### spec.region

`string` · required

The Azure region the factory is created in, e.g. "eastus".

**ForceNew**: changing this destroys and recreates the factory.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.identity

`AzureDataFactoryIdentity`

The factory's managed identity -- required for customer-managed-key
encryption (user-assigned) and for linked services that
authenticate with the factory's own identity. Omit when nothing
requires one.

- rule: identity_ids is required for USER_ASSIGNED and SYSTEM_AND_USER_ASSIGNED and must be empty for SYSTEM_ASSIGNED

### spec.identity.type

`enum` · required

Identity flavor. SYSTEM_ASSIGNED is created and rotated by Azure
with the factory; USER_ASSIGNED brings identities you manage
(grantable on data stores BEFORE the factory exists);
SYSTEM_AND_USER_ASSIGNED carries both.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_data_factory_identity_type_unspecified` -- Not specified: rejected -- an identity block requires a flavor.
- `SYSTEM_ASSIGNED` -- Azure-managed identity created with the factory. Wire value: "SystemAssigned".
- `USER_ASSIGNED` -- Identities you create and manage (AzureUserAssignedIdentity). Wire value: "UserAssigned".
- `SYSTEM_AND_USER_ASSIGNED` -- Both at once. Wire value: "SystemAssigned, UserAssigned".

### spec.identity.identityIds

`[]string | valueFrom`

For USER_ASSIGNED and SYSTEM_AND_USER_ASSIGNED: the user-assigned
identities attached to the factory, by ARM ID. Reference
AzureUserAssignedIdentity resources so data-store grants can be
composed before the factory is created.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.githubConfiguration

`AzureDataFactoryGithubConfiguration`

Bind the factory to a GitHub repository for git-integrated
authoring. At most one of github_configuration and
vsts_configuration may be set. The binding is applied through a
separate configure-repo call after the factory exists, and
REMOVING the block does not detach the repository -- detach it in
the Data Factory Studio or leave the block in place.

### spec.githubConfiguration.accountName

`string` · required

The GitHub account or organization that owns the repository, e.g.
"acme-corp".

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.githubConfiguration.branchName

`string` · required

The collaboration branch the factory authors against, e.g.
"main".

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.githubConfiguration.gitUrl

`string`

The GitHub service URL -- only for GitHub Enterprise Server, e.g.
"https://github.mycompany.com". Omit for github.com.

### spec.githubConfiguration.repositoryName

`string` · required

The repository's name (without the account prefix), e.g.
"data-pipelines".

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.githubConfiguration.rootFolder

`string` · required

The folder inside the repository the factory's resources are
stored under, e.g. "/" for the root.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.githubConfiguration.publishingEnabled

`bool` · optional (explicit presence)

Whether "Publish" from the factory's collaboration branch is
enabled. Default: true (the provider's default) -- the platform
sends the value explicitly (Azure stores the inverse
"disablePublish" flag; both engines translate).

- default: `true`

### spec.vstsConfiguration

`AzureDataFactoryVstsConfiguration`

Bind the factory to an Azure DevOps (VSTS) repository for
git-integrated authoring. At most one of github_configuration and
vsts_configuration may be set. The same detach caveat as
github_configuration applies.

### spec.vstsConfiguration.accountName

`string` · required

The Azure DevOps organization name, e.g. "acme-corp".

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.vstsConfiguration.branchName

`string` · required

The collaboration branch the factory authors against, e.g.
"main".

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.vstsConfiguration.projectName

`string` · required

The Azure DevOps project the repository lives in.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.vstsConfiguration.repositoryName

`string` · required

The repository's name.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.vstsConfiguration.rootFolder

`string` · required

The folder inside the repository the factory's resources are
stored under, e.g. "/" for the root.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.vstsConfiguration.tenantId

`string` · required

The Microsoft Entra tenant the Azure DevOps organization belongs
to, as a UUID.

- rule: {"required":true,"string":{"uuid":true}}

### spec.vstsConfiguration.publishingEnabled

`bool` · optional (explicit presence)

Whether "Publish" from the factory's collaboration branch is
enabled. Default: true (the provider's default) -- the platform
sends the value explicitly.

- default: `true`

### spec.globalParameters

`[]AzureDataFactoryGlobalParameter`

Workspace-wide parameters every pipeline in the factory can
reference. Names are the parameters' identity, so they must be
unique.

### spec.globalParameters[].name

`string` · required

The parameter's name -- its identity across the factory, so names
must be unique.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.globalParameters[].type

`string` · required

The parameter's type: "Array", "Bool", "Float", "Int", "Object",
or "String" (Azure's own vocabulary).

- rule: {"required":true,"string":{"in":["Array","Bool","Float","Int","Object","String"]}}

### spec.globalParameters[].value

`string` · required

The parameter's value, as a string. For "Array" and "Object"
parameters pass the JSON text (e.g. "[1,2]" or "{\"a\":1}") --
Azure stores the typed value; the string is how it travels.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.managedVirtualNetworkEnabled

`bool` · optional (explicit presence)

Whether the factory gets a managed virtual network -- Azure-managed
private networking its integration runtimes execute inside, and
the prerequisite for managed_private_endpoints. Enabling it on an
existing factory is an in-place update.

**ForceNew (one direction)**: DISABLING it after it is enabled
replaces the factory -- decide the posture before shipping
workloads.

### spec.publicNetworkEnabled

`bool` · optional (explicit presence)

Whether the factory's endpoints accept traffic from the public
internet. Set false to restrict access to private endpoints only.
Default: true (Azure's default) -- the platform sends the value
explicitly.

- default: `true`

### spec.purviewId

`string | valueFrom`

Connect the factory to a Microsoft Purview account for lineage
and catalog integration, by the Purview account's ARM ID. Omit
when Purview is not in use. (The catalog has no Purview kind yet,
so pass the ID as a literal or wire an explicit reference.)

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.customerManagedKey

`AzureDataFactoryCustomerManagedKey`

Customer-managed-key encryption at rest: a Key Vault key plus the
user-assigned identity Azure uses to unwrap it. Requires the
identity block to carry a user-assigned flavor, and the unwrap
identity needs get/unwrap/wrap permissions on the vault BEFORE
create -- Azure validates both at deploy time. Once enabled, the
key cannot be removed without recreating the factory (Azure has
no decrypt path).

### spec.customerManagedKey.keyVaultKeyId

`string | valueFrom` · required

The Key Vault key that encrypts the factory's data. Both
versionless and versioned key identifiers are accepted
(https://{vault}.vault.azure.net/keys/{name}[/{version}]) --
prefer versionless so rotation propagates automatically.
Reference an AzureKeyVaultKey output or pass a literal
identifier.

- references: AzureKeyVaultKey (`status.outputs.versionless_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureKeyVaultKey, name: <that resource's name>, fieldPath: status.outputs.versionless_id}} -- a bare string does not parse

### spec.customerManagedKey.userAssignedIdentityId

`string | valueFrom` · required

The user-assigned identity Azure authenticates as when unwrapping
the key. It must be attached to the factory (the identity block's
identity_ids) and hold get/unwrap/wrap permissions on the vault
before the factory is created.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.userManagedIdentityCredentials

`[]AzureDataFactoryUserManagedIdentityCredential`

Named credentials backed by user-assigned managed identities --
the factory's linked services reference these BY NAME to
authenticate as the identity. Credential names share one
namespace with service_principal_credentials, so they must be
unique across both lists.

### spec.userManagedIdentityCredentials[].name

`string` · required

The credential's name -- its identity under the factory (shared
namespace with service-principal credentials), so names must be
unique. Renaming replaces the credential.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.userManagedIdentityCredentials[].identityId

`string | valueFrom` · required

The user-assigned identity the credential wraps, by ARM ID.
Reference an AzureUserAssignedIdentity output or pass a literal
ID.

**ForceNew**: changing the identity replaces the credential.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.userManagedIdentityCredentials[].description

`string`

A human-readable description of what the credential is for.

### spec.userManagedIdentityCredentials[].annotations

`[]string`

Free-form annotation strings stored on the credential.

### spec.servicePrincipalCredentials

`[]AzureDataFactoryServicePrincipalCredential`

Named credentials backed by a service principal whose key lives
in Key Vault (referenced through a Key Vault linked service --
the secret itself never enters this spec). Credential names share
one namespace with user_managed_identity_credentials, so they
must be unique across both lists.

### spec.servicePrincipalCredentials[].name

`string` · required

The credential's name -- its identity under the factory (shared
namespace with user-managed-identity credentials), so names must
be unique. Renaming replaces the credential.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.servicePrincipalCredentials[].tenantId

`string` · required

The Microsoft Entra tenant the service principal belongs to, as a
UUID.

- rule: {"required":true,"string":{"uuid":true}}

### spec.servicePrincipalCredentials[].servicePrincipalId

`string` · required

The service principal's application (client) ID, as a UUID.

- rule: {"required":true,"string":{"uuid":true}}

### spec.servicePrincipalCredentials[].servicePrincipalKey

`AzureDataFactoryServicePrincipalKey`

Where the service principal's key lives: a secret read through a
Key Vault linked service defined in the factory. Omit for
credentials that only carry the principal's identity.

### spec.servicePrincipalCredentials[].servicePrincipalKey.linkedServiceName

`string` · required

The name of the factory's Key Vault LINKED SERVICE the secret is
read through (not the vault's own name).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.servicePrincipalCredentials[].servicePrincipalKey.secretName

`string` · required

The secret's name inside the vault.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.servicePrincipalCredentials[].servicePrincipalKey.secretVersion

`string`

The secret's version. Omit to follow the latest version.

### spec.servicePrincipalCredentials[].description

`string`

A human-readable description of what the credential is for.

### spec.servicePrincipalCredentials[].annotations

`[]string`

Free-form annotation strings stored on the credential.

### spec.managedPrivateEndpoints

`[]AzureDataFactoryManagedPrivateEndpoint`

Managed private endpoints -- private egress from the factory's
managed virtual network to data stores (a storage account, a SQL
server, a Private Link Service). Requires
managed_virtual_network_enabled. The TARGET side must approve
each endpoint's connection before traffic flows; approval happens
outside this resource.

**ForceNew (each entry)**: an endpoint cannot be changed after
create -- changing any field replaces that endpoint (siblings are
untouched; entries are keyed by name).

- rule: Exactly one of subresource_name (regular ARM targets) and fqdns (Private Link Service targets) must be set

### spec.managedPrivateEndpoints[].name

`string` · required

The endpoint's name -- its identity under the factory, so names
must be unique; 2-80 characters of letters, numbers, dots,
hyphens, and underscores, starting with a letter or number and
ending with a letter, number, or underscore.

- rule: Endpoint names must be 2-80 characters of letters, numbers, dots, hyphens, and underscores, starting with a letter or number and ending with a letter, number, or underscore
- rule: {"required":true}

### spec.managedPrivateEndpoints[].targetResourceId

`string | valueFrom` · required

The resource the endpoint connects to, by ARM ID -- a storage
account, a SQL server, or a Private Link Service. Pass a literal
ID or wire a reference to the target kind's own ID output.

Containment-exempt: a managed private endpoint is the factory's
own door to a data store it reads or writes; the factory is not
deployed into that store. Most such targets are containers, and a
factory wired to one by reference would otherwise be drawn inside
the store it merely reaches.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.managedPrivateEndpoints[].subresourceName

`string`

For regular ARM targets: which sub-resource the endpoint binds,
e.g. "blob" for a storage account or "sqlServer" for a SQL
server. Required for regular targets and FORBIDDEN when the
target is a Private Link Service (set fqdns instead) -- exactly
one of the two arms is set.

- rule: subresource_name must be 3-63 characters of letters, numbers, dots, hyphens, and underscores, starting and ending with a letter or number

### spec.managedPrivateEndpoints[].fqdns

`[]string`

For Private Link Service targets: the fully-qualified domain
names the endpoint serves. Required for Private Link Service
targets and FORBIDDEN for regular ARM targets (set
subresource_name instead) -- exactly one of the two arms is set.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.tags

`map<string, string>`

Tags to apply to the factory, merged over the Planton-derived
metadata tags (user values win on key conflicts).

## Validation Rules

- `data_factory_repo_configuration_exclusive`: At most one of github_configuration and vsts_configuration may be set -- a factory binds to one repository
- `data_factory_cmk_requires_user_assigned_identity`: customer_managed_key requires the identity block with a user-assigned flavor (USER_ASSIGNED or SYSTEM_AND_USER_ASSIGNED)
- `data_factory_global_parameter_names_unique`: global_parameters names must be unique -- the name is the parameter's identity
- `data_factory_credential_names_unique`: credential names must be unique across user_managed_identity_credentials and service_principal_credentials -- credentials share one namespace under the factory
- `data_factory_managed_private_endpoint_names_unique`: managed_private_endpoints names must be unique -- the name is the endpoint's identity
- `data_factory_managed_private_endpoints_require_managed_vnet`: managed_private_endpoints require managed_virtual_network_enabled to be true -- the endpoints live inside the managed virtual network

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureDataFactory, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.data_factory_id` | `string` | The factory's Azure Resource Manager ID -- the target an AzureDataFactoryPipeline's data_factory_id references. |
| `status.outputs.data_factory_name` | `string` | The factory's name. |
| `status.outputs.identity_principal_id` | `string` | The principal ID of the factory's system-assigned identity (empty when no system-assigned identity is configured) -- grant this on data stores the factory reads and writes with its own identity. |
| `status.outputs.credential_ids` | `map<string, string>` | The ARM IDs of the factory's named credentials, keyed by credential name ({factory_id}/credentials/{name}) -- covers both credential flavors. |
| `status.outputs.managed_private_endpoint_ids` | `map<string, string>` | The ARM IDs of the factory's managed private endpoints, keyed by endpoint name ({factory_id}/managedVirtualNetworks/default/managedPrivateEndpoints/{name}). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.identity.identityIds` | AzureUserAssignedIdentity | `status.outputs.identity_id` |
| `spec.customerManagedKey.keyVaultKeyId` | AzureKeyVaultKey | `status.outputs.versionless_id` |
| `spec.customerManagedKey.userAssignedIdentityId` | AzureUserAssignedIdentity | `status.outputs.identity_id` |
| `spec.userManagedIdentityCredentials[].identityId` | AzureUserAssignedIdentity | `status.outputs.identity_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureDataFactoryDataFlow | `spec.dataFactoryId` | `status.outputs.data_factory_id` |
| AzureDataFactoryDataset | `spec.dataFactoryId` | `status.outputs.data_factory_id` |
| AzureDataFactoryIntegrationRuntime | `spec.dataFactoryId` | `status.outputs.data_factory_id` |
| AzureDataFactoryLinkedService | `spec.dataFactoryId` | `status.outputs.data_factory_id` |
| AzureDataFactoryPipeline | `spec.dataFactoryId` | `status.outputs.data_factory_id` |
| AzureDataFactoryTrigger | `spec.dataFactoryId` | `status.outputs.data_factory_id` |

## See Also

- [Overview](../README.md)

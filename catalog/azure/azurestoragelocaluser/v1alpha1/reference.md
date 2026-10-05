# AzureStorageLocalUser

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**AzureStorageLocalUserSpec** defines the configuration for creating a
local user on an Azure Storage Account: the credential identity the
account's SFTP endpoint authenticates. Local users are how partners,
legacy pipelines, and managed file transfer tools that only speak SFTP
reach blob storage -- each user carries its own SSH credentials, a home
directory, and per-container permission scopes, so one account serves
many isolated exchange partners.

Local users are many-per-account with independent lifecycles (partners
onboard and offboard without touching the account), which is why they
are a first-class kind rather than a list folded into the account's
spec. The parent is fixed at creation: a user cannot move between
accounts.

**SFTP prerequisites live on the ACCOUNT**: sftp_enabled (with its
is_hns_enabled requirement) turns the endpoint on, and
local_user_enabled (Azure's default: true) permits users to exist.
Azure accepts a local user on an account without SFTP -- it just can't
connect -- so wire both account switches when the user is meant to log
in. Clients connect as {account-name}.{user-name} (the sftp_username
output) on port 22.

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureStorageLocalUser
metadata:
  name: test-storage-local-user
spec:
  storageAccountId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Storage/storageAccounts/plantonhacksftp
  userName: partner01
  # Exercises BOTH auth methods together: key list paired with the key
  # flag, plus the Azure-generated password.
  sshKeyEnabled: true
  sshPasswordEnabled: true
  sshAuthorizedKeys:
    - key: ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIGJ3iXFCbaYbzUZzz2i2Cv6/L7Ohtq5rM9lZ74W2mBrz partner-pipeline
      description: partner pipeline deploy key
  homeDirectory: inbound/partner01
  # Exercises both service vocabularies and the five-boolean grant
  # shape.
  permissionScopes:
    - service: BLOB
      resourceName:
        value: inbound
      read: true
      write: true
      list: true
      create: true
    - service: FILE
      resourceName:
        value: team-share
      read: true
      list: true
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.storageAccountId` | `string \| valueFrom` | yes |  | AzureStorageAccount (`status.outputs.storage_account_id`) |
| `spec.userName` | `string` | yes |  |  |
| `spec.sshKeyEnabled` | `bool` |  |  |  |
| `spec.sshPasswordEnabled` | `bool` |  |  |  |
| `spec.sshAuthorizedKeys` | `[]AzureStorageLocalUserSshAuthorizedKey` |  |  |  |
| `spec.sshAuthorizedKeys[].key` | `string` | yes |  |  |
| `spec.sshAuthorizedKeys[].description` | `string` |  |  |  |
| `spec.homeDirectory` | `string` |  |  |  |
| `spec.permissionScopes` | `[]AzureStorageLocalUserPermissionScope` |  |  |  |
| `spec.permissionScopes[].service` | `enum` |  |  |  |
| `spec.permissionScopes[].resourceName` | `string \| valueFrom` | yes |  | AzureStorageContainer (`status.outputs.container_name`) |
| `spec.permissionScopes[].read` | `bool` |  |  |  |
| `spec.permissionScopes[].write` | `bool` |  |  |  |
| `spec.permissionScopes[].delete` | `bool` |  |  |  |
| `spec.permissionScopes[].list` | `bool` |  |  |  |
| `spec.permissionScopes[].create` | `bool` |  |  |  |

## Field Details

### spec.storageAccountId

`string | valueFrom` · required

The storage account the user lives on, by ARM ID. References an
AzureStorageAccount's storage_account_id output so the account and
its users compose in one manifest set. The account needs
sftp_enabled: true (which requires is_hns_enabled) for the user to
actually connect. Fixed at creation.

- references: AzureStorageAccount (`status.outputs.storage_account_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureStorageAccount, name: <that resource's name>, fieldPath: status.outputs.storage_account_id}} -- a bare string does not parse

### spec.userName

`string` · required

The user's name: 3-64 lowercase letters and digits only (no hyphens
or underscores). Unique within the account; the SFTP login is
{account-name}.{user-name}. Changing the name replaces the user --
and regenerates its credentials.

- rule: user_name must be 3-64 lowercase letters and digits only (no hyphens or underscores)
- rule: {"required":true}

### spec.sshKeyEnabled

`bool`

Whether the user authenticates with SSH PUBLIC KEYS. Enable this
and list the keys in ssh_authorized_keys -- key auth is the posture
to prefer (no shared secret to distribute or rotate out-of-band).
At least one of ssh_key_enabled / ssh_password_enabled must be on.

### spec.sshPasswordEnabled

`bool`

Whether the user authenticates with an AZURE-GENERATED password.
Azure mints the password at creation and returns it exactly once --
it lands in the password output; there is no way to choose or
retrieve it later (flipping this off and on regenerates it). At
least one of ssh_key_enabled / ssh_password_enabled must be on.

### spec.sshAuthorizedKeys

`[]AzureStorageLocalUserSshAuthorizedKey`

The SSH public keys the user may authenticate with -- required when
(and only when) ssh_key_enabled is on. Standard OpenSSH public-key
format (the .pub file content, e.g. "ssh-ed25519 AAAA... label").

### spec.sshAuthorizedKeys[].key

`string` · required

The SSH PUBLIC key in OpenSSH format -- the .pub file content
("ssh-ed25519 AAAA..." or "ssh-rsa AAAA..."), never the private
key. The private half stays with the partner; Azure only ever sees
this public half.

- rule: key must be an OpenSSH public key, e.g. "ssh-ed25519 AAAA... label" or "ssh-rsa AAAA..."
- rule: {"required":true}

### spec.sshAuthorizedKeys[].description

`string`

A label for the key -- whose laptop, which pipeline -- so rotations
remove the right one.

### spec.homeDirectory

`string`

The directory an SFTP session lands in after login, as a path INSIDE
the account's blob namespace -- "{container}" or
"{container}/{path}" (no leading slash). Unset lands the session at
the account root, where the user sees every container its
permission scopes allow.

### spec.permissionScopes

`[]AzureStorageLocalUserPermissionScope`

What the user may do, per container or file share. Each scope
grants a set of operations on ONE named resource -- the isolation
boundary that lets one account serve many partners (each user
scoped to its own container). A user with no scopes can log in but
touch nothing.

### spec.permissionScopes[].service

`enum`

Which storage service the resource belongs to: BLOB for containers
(the SFTP case) or FILE for file shares.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `azure_storage_local_user_permission_service_unspecified` -- Not specified -- invalid; name the service the resource lives in.
- `BLOB` -- A blob container -- what the SFTP endpoint serves.
- `FILE` -- An Azure Files share.

### spec.permissionScopes[].resourceName

`string | valueFrom` · required

The container (service: BLOB) or file share (service: FILE) this
scope grants access to, by name. References an
AzureStorageContainer's container_name output by default; for a
file share, point valueFrom at an AzureStorageShare's share_name
output instead. The resource must live on the SAME account as the
user.

- references: AzureStorageContainer (`status.outputs.container_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureStorageContainer, name: <that resource's name>, fieldPath: status.outputs.container_name}} -- a bare string does not parse

### spec.permissionScopes[].read

`bool`

Whether the user may READ objects in the resource.

### spec.permissionScopes[].write

`bool`

Whether the user may WRITE (upload/overwrite) objects.

### spec.permissionScopes[].delete

`bool`

Whether the user may DELETE objects.

### spec.permissionScopes[].list

`bool`

Whether the user may LIST the resource's contents.

### spec.permissionScopes[].create

`bool`

Whether the user may CREATE new objects and directories.

## Validation Rules

- `storage_local_user_auth_method_required`: enable at least one authentication method: ssh_key_enabled and/or ssh_password_enabled
- `storage_local_user_keys_pair_with_key_auth`: ssh_authorized_keys must be set when ssh_key_enabled is on, and left empty when it is off

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureStorageLocalUser, name: <resource-name>, fieldPath: status.outputs.<output>}`. A sensitive output is a secret the resource generates: on Planton it is kept in the organization's secret store and the output holds a `$secret/` reference, so feed it only to a sensitive field.

| Output | Type | Description |
|---|---|---|
| `status.outputs.local_user_id` | `string` | The local user's Azure Resource Manager ID. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Storage/storageAccounts/{account}/localUsers/{name} |
| `status.outputs.user_name` | `string` | The user's name within the account -- the second half of the SFTP login. |
| `status.outputs.sftp_username` | `string` | The full SFTP login: {account-name}.{user-name} -- what a client passes as the username when connecting to {account-name}.blob.core.windows.net on port 22. |
| `status.outputs.sid` | `string` | The user's unique Security Identifier (SID) -- Azure generates it at creation; Azure Files NTFS-style ACLs reference principals by SID. An identifier, not a credential: it authenticates nothing on its own, so it is a public output even though the provider marks the attribute sensitive. |
| `status.outputs.password` | `string` (sensitive) | The Azure-generated SSH password -- returned EXACTLY ONCE, at the creation that enabled ssh_password_enabled (empty when password auth is off). Azure never returns it again: losing it means regenerating it (flip ssh_password_enabled off and on). SECRET. |
| `status.outputs.storage_account_name` | `string` | The name of the storage account the user lives on, parsed from the resolved account ID -- saves consumers a second reference when they need the account/user pair. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.storageAccountId` | AzureStorageAccount | `status.outputs.storage_account_id` |
| `spec.permissionScopes[].resourceName` | AzureStorageContainer | `status.outputs.container_name` |

## See Also

- [Overview](../README.md)

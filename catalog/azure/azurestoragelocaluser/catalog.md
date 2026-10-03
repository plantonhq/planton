# Azure Storage Local User

Deploys a local user on an Azure Storage Account -- the credential identity the account's SFTP endpoint authenticates. Local users are how partners, legacy pipelines, and managed file-transfer tools that only speak SFTP reach blob storage: each user carries its own SSH credentials, a home directory, and per-container permission scopes, so one account serves many isolated exchange partners. Users are many-per-account with independent lifecycles (partners onboard and offboard without touching the account), which is why they are a first-class kind rather than a list folded into the account's spec.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Storage Local User** -- a local user on the referenced storage account (by ARM ID -- the control-plane path), with SSH key and/or Azure-minted password authentication, an optional home directory, and per-resource permission scopes

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An AzureStorageAccount** the user will live on, referenced through `storageAccountId`. The parent is fixed at creation.
- **SFTP prerequisites on the ACCOUNT**: `sftpEnabled: true` (which requires `isHnsEnabled`) turns the endpoint on, and `localUserEnabled` (Azure's default) permits users to exist. Azure accepts a local user on an account without SFTP -- it just can't connect.
- **Scoped resources**: the containers (or file shares) the user's permission scopes name must live on the SAME account.

## Deploy

### Console

Open the deployment store, find **Azure Storage Local User**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Partner SFTP User with Key Authentication** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureStorageLocalUser
metadata:
  name: partner-acme-sftp
  org: acme-corp
  env: prod
spec:
  storageAccountId:
    valueFrom:
      kind: AzureStorageAccount
      name: exchange-account
      fieldPath: status.outputs.storage_account_id
  userName: partneracme
  sshKeyEnabled: true
  sshAuthorizedKeys:
    - key: ssh-ed25519 AAAAC3NzaC1lZDI1NTE5ExampleDocsOnlyPublicKey0000000000000000000 partner-laptop
      description: partner-acme laptop
  homeDirectory: partner-inbound/incoming
  permissionScopes:
    - service: BLOB
      resourceName:
        valueFrom:
          kind: AzureStorageContainer
          name: partner-inbound
          fieldPath: status.outputs.container_name
      write: true
      list: true
      create: true
```

```shell
planton apply -f local-user.yaml
```

This creates a key-authenticated partner scoped to one container -- the partner connects as `{account}.partneracme` on port 22 and lands in its own inbound directory. An Infra Job tracks the provisioning in real time.

### InfraChart

When the account, container, and user deploy in the same InfraChart, wire both references with ValueFromRef:

```yaml
spec:
  storageAccountId:
    valueFrom:
      kind: AzureStorageAccount
      name: exchange-account
      fieldPath: status.outputs.storage_account_id
  userName: partneracme
  permissionScopes:
    - service: BLOB
      resourceName:
        valueFrom:
          kind: AzureStorageContainer
          name: partner-inbound
          fieldPath: status.outputs.container_name
      write: true
      list: true
      create: true
```

The InfraPipeline resolves the dependency graph, deploys the account and container first, then provisions the user with the resolved values.

## Key Configuration

These are the most important decisions when configuring a local user. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**User name** -- `userName` is 3-64 lowercase letters and digits only. The SFTP login is `{account-name}.{user-name}`. Renaming replaces the user -- and regenerates its credentials.

**Authentication** -- at least one of `sshKeyEnabled` / `sshPasswordEnabled` must be on (the spec enforces it). Key auth is the posture to prefer: list OpenSSH public keys in `sshAuthorizedKeys` (required when -- and only when -- key auth is on). Password auth mints an AZURE-GENERATED password returned EXACTLY ONCE in the outputs; there is no way to choose or later retrieve it, and flipping the switch off and on regenerates it.

**Home directory** -- where a session lands after login, as a path inside the blob namespace (`{container}` or `{container}/{path}`, no leading slash). Blank lands at the account root.

**Permission scopes** -- each scope grants read/write/delete/list/create on ONE container (`service: BLOB`) or file share (`service: FILE`) -- the isolation boundary that lets one account serve many partners. A user with no scopes can log in but touch nothing.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureStorageAccount** | `storageAccountId` | `status.outputs.storage_account_id` |
| **AzureStorageContainer** | per-scope `resourceName` (service: BLOB) | `status.outputs.container_name` |
| **AzureStorageShare** | per-scope `resourceName` (service: FILE) | `status.outputs.share_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `sftp_username` | The full login: `{account-name}.{user-name}` | What the partner types as the SFTP username |
| `sid` | The user's Security Identifier (SECRET-bearing) | Azure Files NTFS-style ACLs reference principals by SID |
| `password` | The Azure-minted SSH password (SECRET; returned exactly once; empty when password auth is off) | Handed to the partner over a secure channel |
| `storage_account_name` | The parent account's name | Composing the SFTP host `{account}.blob.core.windows.net` |

The outputs also carry `local_user_id` (the user's ARM ID) and `user_name` (the login's second half, already embedded in `sftp_username`) -- neither has a ValueFromRef consumer.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Partner with key auth** -- one user per partner, one ed25519 key per person or pipeline, write+create+list on the partner's own inbound container. Start from the **Partner SFTP User with Key Authentication** preset.

**Password drop-box** -- for tools that cannot do keys: password auth on, the once-only password captured from the outputs and delivered securely. Start from the **Password-Authenticated Drop Box** preset.

## Works With

- [**Azure Storage Account**](/infra-catalog/azure-storage-account) -- the SFTP-enabled (HNS) parent account
- [**Azure Storage Container**](/infra-catalog/azure-storage-container) -- the per-partner containers permission scopes grant into
- [**Azure Storage Share**](/infra-catalog/azure-storage-share) -- file shares reachable through FILE-service scopes

# AzureStorageDataLakeGen2Filesystem - Pulumi Module

Pulumi implementation for the AzureStorageDataLakeGen2Filesystem
component.

## Architecture

```
storage.DataLakeGen2Filesystem (single resource)
```

## Key Design Decisions

- **A data-plane resource behind an ARM-id parent** -- the provider
  creates the filesystem through the account's dfs endpoint (with the
  account's shared key by default), so the account must be reachable
  from where the deploy runs.
- **`filesystem_id` is the CONSTRUCTED ARM container-proxy id**
  (`{account}/blobServices/default/containers/{name}`) -- ADLS
  filesystems surface in ARM as blob containers, and the proxy id is
  what data-plane role assignments scope to; the provider's own
  resource id is a dfs URL nothing management-grain can consume. Built
  identically in the Terraform module.
- **Optional inputs are presence-guarded** -- owner, group, and the
  encryption scope are Computed on the provider, so the module sends
  them only when set and Azure's server-side defaults stand.
- **ACL vocabulary maps** -- the spec's closed scope/type enums map to
  the data plane's lowercase values; an unset scope stays unset so the
  provider's own default (access) applies on both engines.
- **POSIX control requires HNS on the ACCOUNT** -- an apply-time Azure
  contract (the account arrives as a reference, so the spec cannot
  check it).
- **No Azure tags**: ARM does not support tags on this resource; the
  platform's identity tags live on the parent account.

## Provider

The Azure provider is built by the shared
`pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)` builder, which
dispatches static client-secret, keyless web-identity (OIDC), and
ambient credential chains. Never construct a provider inline.

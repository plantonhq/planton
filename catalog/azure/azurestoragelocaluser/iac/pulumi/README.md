# AzureStorageLocalUser - Pulumi Module

Pulumi implementation for the AzureStorageLocalUser deployment
component.

## Architecture

```
storage.LocalUser (single resource)
```

## Key Design Decisions

- **The parent is addressed by ARM id** (`storage_account_id`); the
  account NAME (for the `storage_account_name` output and the composed
  `sftp_username` login) is parsed from it, with a loud error if the id
  is not a storage-account ARM id.
- **The auth booleans are sent unconditionally** -- false is the
  provider default on both engines, and the spec enforces
  at-least-one-method plus the keys-iff-key-auth pairing so apply-time
  rejections become validate-time messages.
- **The permission block is a pass-through** -- the spec models the
  same five grant booleans the provider exposes; the API's `rwdlc` wire
  string stays provider-internal on both engines.
- **`password` is a secret output** -- it is returned by Azure exactly
  once (at the creation that enabled password auth) and REGENERATES when
  the flag flips off and back on. The provider also marks `sid`
  sensitive; it is an identifier, so the module unwraps it.
- **No Azure tags**: ARM does not support tags on localUsers; the
  platform's identity tags live on the parent account.

## Provider

The Azure provider is built by the shared
`pulumiazureprovider.Get(ctx, stackInput.ProviderConfig)` builder, which
dispatches static client-secret, keyless web-identity (OIDC), and
ambient credential chains. Never construct a provider inline.

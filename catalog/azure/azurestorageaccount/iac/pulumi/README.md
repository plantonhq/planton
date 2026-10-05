# AzureStorageAccount - Pulumi Module

Pulumi implementation for the AzureStorageAccount kind.

## Architecture

```
storage.Account (the account)
├── storage.ManagementPolicy      (when lifecycle_rules are declared)
└── storage.AccountStaticWebsite  (when static_website is declared)
```

## Key Design Decisions

- **Unset SKU enums materialize the spec's documented defaults**
  (StorageV2 / Standard / LRS) -- IaC inputs built from a manifest do
  NOT materialize proto defaults, and the provider requires tier and
  replication. Unset `access_tier` / `dns_endpoint_type` /
  `allowed_copy_scope` / key types are NOT sent at all, mirroring the
  Terraform module's nulls so both engines produce the same ARM payload.
- **Every optional-with-default bool is presence-guarded** to its proto
  default (`https_traffic_only_enabled`, `shared_access_key_enabled`,
  `allow_nested_items_to_be_public`, `public_network_access_enabled`,
  `local_user_enabled`, retention days).
- **The lifecycle policy is a separate `storage.ManagementPolicy`**
  parented to the account -- ARM models it as one singleton per-account
  policy document, which is why the rules fold into the account spec
  rather than being their own kind. Absent aging thresholds are simply
  not sent (the provider's -1 sentinel is an HCL-ergonomics artifact
  this module never needs).
- **The static website is the standalone `storage.AccountStaticWebsite`
  resource** -- azurerm removed the inline block in v5; the spec's
  shape is identical either way.
- **`identity_principal_id` is exported via ApplyT** on the identity
  output (empty unless the type includes SystemAssigned).

## Provider

The Azure provider is built by the shared
`pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)` builder, which
dispatches static client-secret, keyless web-identity (OIDC), and
ambient credential chains. Never construct a provider inline.

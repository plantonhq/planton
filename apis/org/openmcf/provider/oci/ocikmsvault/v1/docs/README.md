# OciKmsVault — Design Notes

## Design Rationale

OciKmsVault provisions a single KMS vault resource. Unlike more complex components, there are no sub-resources to bundle — the vault is one atomic OCI resource. The component's primary value is exporting the `managementEndpoint` and `cryptoEndpoint` outputs that OciKmsKey and other downstream resources depend on.

### Why is this a separate component from OciKmsKey?

Vaults and keys have different lifecycles. A single vault typically contains many keys (for Block Volume, Object Storage, Database, etc.), and keys rotate independently of vault creation. Separating them enables a 1:N composition: one vault manifest, many key manifests referencing its `managementEndpoint`.

### Why are there three vault types instead of separate components?

All three vault types (`default_vault`, `virtual_private`, `external`) share the same OCI API resource type (`oci_kms_vault`) with the same outputs. The only difference is the `vault_type` parameter and the conditional `externalKeyManagerMetadata` block. Separate components would duplicate 90% of the spec for a single-field difference.

### Why is `externalKeyManagerMetadata` validated with CEL?

The `external` vault type requires OAuth credentials and a private endpoint, while `default_vault` and `virtual_private` do not support those fields. CEL rules enforce both directions: `external` requires metadata, and non-external forbids it. This prevents confusing error messages from the OCI API at deploy time.

## Trade-offs

| Decision | Benefit | Cost |
|----------|---------|------|
| Single component for all vault types | One spec to learn; shared validation and outputs | CEL rules needed to enforce type-specific fields |
| Endpoint export as stack outputs | Clean composability with OciKmsKey via `valueFrom` | Consumers must know which output to reference |
| `clientAppSecret` as plaintext field | Matches OCI API model | Sensitive value in manifest; not returned after creation |

## Resource Graph

```
OciKmsVault
└── oci_kms_vault (always)
    ├── external_key_manager_metadata (if vault_type == external)
    │   └── oauth_metadata
    └── outputs: vault_id, crypto_endpoint, management_endpoint
```

## Deferred from v1

- **restore_from_file / restore_from_object_store** — operational restore workflows, not initial provisioning.
- **time_of_deletion** — deletion scheduling is an operational concern.
- **defined_tags / system_tags** — managed by platform. `freeform_tags` are auto-populated from `metadata.labels`.
- **Vault replication** — separate OCI resource with its own lifecycle, not tightly coupled to vault creation.

## Freeform Tags

The module automatically populates freeform tags from metadata:

| Tag Key | Source |
|---------|--------|
| `resource` | `"true"` (constant) |
| `resource_kind` | `OciKmsVault` |
| `resource_id` | `metadata.id` |
| `organization` | `metadata.org` (if set) |
| `environment` | `metadata.env` (if set) |
| All `metadata.labels` | Copied as-is |

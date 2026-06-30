# OciVaultSecret — Design Notes

## Design Rationale

OciVaultSecret provisions a single Vault secret resource with support for two content modes, lifecycle rules, and rotation configuration. The component is the most complex in the Security and Secrets phase because secrets have multiple orthogonal concerns (content provisioning, lifecycle, rotation) that must compose cleanly.

### Why two content modes (explicit vs auto-generation)?

OCI Vault secrets natively support both modes. Auto-generation is valuable for machine credentials where no human needs to know the initial value — OCI generates and stores it directly in the vault. Explicit content is needed for pre-existing credentials, certificates, or API keys that must be imported. Making them mutually exclusive matches the OCI API constraint and prevents ambiguous state.

### Why is `secretName` separate from `metadata.name`?

The OCI Vault secret name is immutable after creation and must be unique within the vault. The Planton metadata name is the resource identifier within the Planton control plane. They may differ because the OCI name has different constraints (e.g., no prefix requirements) and the resource may be renamed in the Planton catalog without recreating the OCI secret.

### Why use a flat discriminated model for SecretRule?

OCI's API uses a discriminated union where `rule_type` selects which fields are relevant. Rather than modeling separate proto messages for each rule type (which would require a `oneof` or repeated typed messages), a flat model with all fields and a discriminator is simpler for YAML authoring. Users set `ruleType` and the fields for that type; irrelevant fields are ignored.

### Why is rotation config a nested message instead of a separate component?

Rotation configuration is an inline property of the OCI secret resource. It configures the OCI Vault service to periodically invoke rotation against a target system. Separating it would create a resource boundary that doesn't exist in the OCI API and would complicate the lifecycle (rotation config is updated atomically with the secret).

## Trade-offs

| Decision | Benefit | Cost |
|----------|---------|------|
| Mutually exclusive content modes | Clear semantics; matches OCI API | CEL rules needed to enforce exclusivity |
| Flat discriminated SecretRule | Simple YAML authoring | Irrelevant fields visible for each rule type |
| Rotation config inline | Matches OCI API model; atomic updates | More complex spec than a standalone resource |
| `secretName` separate from `metadata.name` | OCI naming flexibility; no recreation on catalog rename | Two name fields to manage |
| `content_type` hardcoded to BASE64 | Simplifies spec (only valid value today) | Would need spec change if OCI adds new content types |

## Resource Graph

```
OciVaultSecret
└── oci_vault_secret (always)
    ├── secret_content (if explicit content mode)
    ├── secret_generation_context (if auto-generation mode)
    ├── secret_rules (0..N inline rules)
    ├── rotation_config (optional)
    │   └── target_system_details (adb or function)
    └── outputs: secret_id, current_version_number
```

## Deferred from v1

- **replication_config** — cross-region secret replication requires vault and key OCIDs from other regions, adding significant complexity. Consistent with OciKmsVault excluding vault replication.
- **defined_tags / system_tags** — managed by platform. `freeform_tags` are auto-populated from `metadata.labels`.

## Freeform Tags

The module automatically populates freeform tags from metadata:

| Tag Key | Source |
|---------|--------|
| `resource` | `"true"` (constant) |
| `resource_kind` | `OciVaultSecret` |
| `resource_id` | `metadata.id` |
| `organization` | `metadata.org` (if set) |
| `environment` | `metadata.env` (if set) |
| All `metadata.labels` | Copied as-is |

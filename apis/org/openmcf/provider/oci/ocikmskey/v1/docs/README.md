# OciKmsKey — Design Notes

## Design Rationale

OciKmsKey provisions a single KMS key resource inside a vault. The key shape (algorithm, length, curve) is immutable after creation, matching OCI's API behavior.

### Why is OciKmsKey separate from OciKmsVault?

Keys and vaults have different lifecycles and cardinality. A single vault typically contains many keys — one for Block Volume encryption, another for Object Storage, another for Database. Keys rotate independently. Separating them enables 1:N composition and avoids a monolithic vault manifest that grows with every new encryption use case.

### Why reference the vault via `managementEndpoint` instead of `vaultId`?

The OCI KMS API requires the management endpoint (a URL) to create keys — not the vault OCID. The Pulumi provider's `kms.KeyArgs` takes `ManagementEndpoint` as a required parameter. Using the endpoint directly avoids an extra lookup step and matches the API contract.

### Why are all three algorithms in one component?

AES, RSA, and ECDSA keys share the same OCI resource type (`oci_kms_key`) and the same lifecycle. The only difference is the `keyShape` configuration. Three separate components would duplicate 95% of the spec. CEL validation ensures algorithm-specific constraints (e.g., `curveId` only for ECDSA).

### Why is auto-rotation configuration inline?

Auto-rotation details are properties of the key resource in the OCI API, not a separate sub-resource. The Pulumi provider models them as `kms.KeyAutoKeyRotationDetailsArgs` within the key args. Separating them would create an abstraction that doesn't exist in the underlying API.

## Trade-offs

| Decision | Benefit | Cost |
|----------|---------|------|
| Single component for all algorithms | One spec; reduced duplication | CEL rules needed for algorithm-specific validation |
| `managementEndpoint` as reference | Matches OCI API contract exactly | Less intuitive than referencing vault by OCID |
| Immutable key shape | Matches OCI API; prevents accidental algorithm changes | Must create a new key to change algorithm or length |
| Auto-rotation inline | Matches Pulumi provider model | All rotation config updated atomically with other key properties |

## Resource Graph

```
OciKmsKey
└── oci_kms_key (always)
    ├── key_shape (algorithm + length + optional curve_id)
    ├── auto_key_rotation_details (if is_auto_rotation_enabled)
    ├── external_key_reference (if protection_mode == external)
    └── outputs: key_id, current_key_version
```

## Deferred from v1

- **oci_kms_key_version** — manual key rotation is an operational concern; auto-rotation handles this declaratively.
- **desired_state** — ENABLED/DISABLED toggle is an operational concern; keys always create as ENABLED.
- **restore_from_file / restore_from_object_store** — operational restore workflows, not initial provisioning.
- **time_of_deletion** — deletion scheduling is an operational concern.
- **defined_tags / system_tags** — managed by platform. `freeform_tags` are auto-populated from `metadata.labels`.

## Freeform Tags

The module automatically populates freeform tags from metadata:

| Tag Key | Source |
|---------|--------|
| `resource` | `"true"` (constant) |
| `resource_kind` | `OciKmsKey` |
| `resource_id` | `metadata.id` |
| `organization` | `metadata.org` (if set) |
| `environment` | `metadata.env` (if set) |
| All `metadata.labels` | Copied as-is |

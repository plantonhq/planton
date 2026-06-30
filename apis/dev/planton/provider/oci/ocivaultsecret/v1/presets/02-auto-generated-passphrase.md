# Auto-Generated Passphrase with Rotation

This preset creates an OCI Vault Secret where OCI automatically generates a 32-character passphrase and rotates it every 30 days against an Autonomous Database. The rotation process generates a new passphrase, updates the database credentials, and stores the new version in the vault -- all without human intervention. A reuse rule prevents recycling previous passphrases.

## When to Use

- Autonomous Database (ATP/ADW) admin or application credentials that should rotate automatically
- Environments where manual password rotation is operationally expensive or error-prone
- Compliance requirements mandating periodic credential rotation with audit trails
- Any database credential where eliminating human knowledge of the password improves security posture

## Key Configuration Choices

- **Auto-generation enabled** (`enableAutoGeneration: true`) -- OCI generates the secret content instead of requiring the user to provide it. This is mutually exclusive with `secretContent`; you cannot provide explicit content when auto-generation is enabled.
- **Passphrase generation** (`generationType: passphrase`, `passphraseLength: 32`) -- generates a 32-character passphrase using the `SECRET_TPL_DBAAS_DEFAULT` template, which produces passwords compatible with Oracle Database password policies. The template handles character class requirements (uppercase, lowercase, digits, special characters) automatically.
- **30-day scheduled rotation** (`rotationConfig` with `P30D`) -- every 30 days, OCI generates a new passphrase, connects to the target Autonomous Database, updates the database credential, and stores the new passphrase as a new secret version. Applications retrieving the secret always get the latest version. Adjust the interval based on your compliance policy (e.g., `P7D` for weekly, `P90D` for quarterly).
- **Autonomous Database target** (`targetSystemType: adb`) -- the rotation targets an Autonomous Database instance. OCI natively understands how to update ADB credentials during rotation. For non-database targets, use `function` target type with a custom OCI Function that handles the credential update logic.
- **Reuse prevention** (`secret_reuse_rule` with `isEnforcedOnDeletedSecretVersions: true`) -- prevents any previously used passphrase from being reused, even if the previous version has been deleted. This satisfies compliance requirements that mandate unique credentials across rotation cycles.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the secret will be created | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |
| `<secret-name>` | Unique name for the secret within the vault (e.g., `myapp-adb-admin`) | Choose a name following your naming convention |
| `<vault-ocid>` | OCID of the vault that will contain this secret | `OciKmsVault` status outputs (`vaultId`), or OCI Console > Identity & Security > Vault |
| `<encryption-key-ocid>` | OCID of the AES master encryption key within the vault | `OciKmsKey` status outputs (`keyId`), or OCI Console > Identity & Security > Vault > Keys |
| `<autonomous-database-ocid>` | OCID of the Autonomous Database whose credentials will be rotated | `OciAutonomousDatabase` status outputs (`autonomousDatabaseId`), or OCI Console > Oracle Database > Autonomous Database |

## Related Presets

- **01-explicit-credential** -- Use instead when the secret content is user-provided and rotation is handled externally

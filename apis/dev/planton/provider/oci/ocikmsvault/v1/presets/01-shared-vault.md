# Shared Vault

This preset creates an OCI KMS Vault with the default vault type, which uses a shared HSM partition. Shared vaults provide FIPS 140-2 Level 3 certified key storage at lower cost than dedicated vaults, and are the recommended starting point for the vast majority of workloads that need customer-managed encryption keys.

## When to Use

- First vault in a compartment for managing encryption keys across OCI services (Block Volume, Object Storage, Database, File Storage)
- Teams getting started with customer-managed encryption who need a cost-effective HSM-backed vault
- Environments where cryptographic operation throughput is moderate (fewer than hundreds of operations per second)
- Development, staging, and most production workloads that do not require dedicated HSM isolation

## Key Configuration Choices

- **Default vault type** (`vaultType: default_vault`) -- uses a shared HSM partition within the OCI KMS fleet. Key material is still isolated per tenant and protected by FIPS 140-2 Level 3 certified hardware. The shared model keeps costs low while providing the same cryptographic guarantees as dedicated vaults. Upgrade to `virtual_private` when throughput limits become a bottleneck.
- **Display name** (`displayName: vault`) -- a short, descriptive name. In practice, name this after the team, project, or environment it serves (e.g., `platform-prod`, `data-team-staging`).

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the vault will be created | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |

## Related Presets

- **02-dedicated-vault** -- Use instead when high-volume cryptographic operations require dedicated HSM throughput limits

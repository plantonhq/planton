# Dedicated Vault

This preset creates an OCI KMS Vault with the virtual private vault type, which allocates a dedicated HSM partition exclusively for your tenancy. Dedicated vaults provide higher cryptographic operation throughput limits and full HSM isolation, required for high-volume encryption workloads or environments with strict compliance requirements mandating dedicated hardware.

## When to Use

- High-volume encryption workloads that exceed shared vault throughput limits (e.g., encrypting thousands of objects per minute, high-throughput database TDE)
- Compliance regimes requiring dedicated HSM partitions with no shared tenancy on the hardware (PCI-DSS Level 1, HIPAA with dedicated HSM controls)
- Centralized key management hubs serving multiple compartments or applications with aggregate high throughput
- Production environments where cryptographic operation latency consistency is critical

## Key Configuration Choices

- **Virtual private vault type** (`vaultType: virtual_private`) -- allocates a dedicated HSM partition within the OCI KMS fleet. This provides higher throughput limits for cryptographic operations (encrypt, decrypt, sign, verify) and guarantees that no other tenancy shares the same HSM partition. The tradeoff is higher cost compared to `default_vault`. The vault type is immutable after creation.
- **Display name** (`displayName: dedicated-vault`) -- name this after its purpose (e.g., `platform-prod-dedicated`, `payment-processing`).

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the vault will be created | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |

## Related Presets

- **01-shared-vault** -- Use instead when throughput requirements are moderate and dedicated HSM isolation is not mandated by compliance

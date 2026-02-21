# RSA-4096 HSM Signing Key

This preset creates an RSA-4096 asymmetric key stored in an HSM for digital signing and verification. Asymmetric keys are used when the signer and verifier are different entities -- the private key signs within the HSM and the public key can be distributed to any number of verifiers without compromising key material.

## When to Use

- Container image signature verification in OKE clusters or OCI Functions applications (`imagePolicyConfig`)
- Digital document or artifact signing where recipients verify signatures independently
- API request signing for service-to-service authentication
- Code signing for CI/CD pipelines where build artifacts must be cryptographically verified before deployment

## Key Configuration Choices

- **RSA-4096** (`algorithm: rsa`, `length: 512`) -- 4096-bit RSA provides a strong security margin for long-lived signing keys. The `length` field is specified in bytes (512 bytes = 4096 bits), consistent with the OCI KMS API. RSA-2048 (length: 256) is acceptable for shorter-lived keys, but 4096 is preferred for keys that may be in service for years. The key shape is immutable after creation.
- **HSM protection** (`protectionMode: hsm`) -- the private key never leaves the HSM. All signing operations are performed inside the hardware module, preventing key extraction even by administrators. This is critical for signing use cases where private key compromise would allow forging signatures.
- **Auto-rotation disabled** (`isAutoRotationEnabled: false`) -- unlike symmetric keys where OCI transparently handles version selection, asymmetric key rotation requires all verifiers to be updated with the new public key. Enabling auto-rotation without coordinated consumer updates would cause signature verification failures. Rotate manually when needed and distribute the new public key before retiring the old version.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the key will be created | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |
| `<vault-management-endpoint>` | Management endpoint URL of the vault containing this key | `OciKmsVault` status outputs (`managementEndpoint`), or OCI Console > Identity & Security > Vault > Vault Details |

## Related Presets

- **01-aes-256-hsm-auto-rotation** -- Use instead for symmetric data-at-rest encryption (Block Volume, Object Storage, Database)

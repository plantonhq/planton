# GcpKmsKey — Pulumi Module

This directory contains the Pulumi Go implementation for the GcpKmsKey component.

## Module Structure

```
module/
  main.go       — Entry point: creates GCP provider, orchestrates resources
  locals.go     — Locals struct, GCP label computation
  kms_key.go    — Creates kms.CryptoKey with all field mappings
  outputs.go    — Output key constants

main.go         — Pulumi program entrypoint (loads stack input, calls module)
Pulumi.yaml     — Pulumi project configuration
Makefile        — Build, preview, up, destroy targets
```

## Outputs

| Key | Description |
|-----|-------------|
| `key_id` | Fully qualified crypto key path (`projects/{p}/locations/{l}/keyRings/{kr}/cryptoKeys/{name}`) |
| `key_name` | Short name of the crypto key |

## Local Development

```bash
make build      # Compile the Pulumi binary
make preview    # Preview changes
make up         # Apply changes
make destroy    # Destroy resources
```

## Notes

- Unlike GcpKmsKeyRing, crypto keys **do support GCP labels**. Framework labels are applied automatically.
- Keys **cannot be deleted** from GCP. Destroy only removes key versions and disables rotation.
- The `key_id` output is the fully qualified path used by downstream CMEK consumers (BigQuery, Spanner, GKE, etc.).

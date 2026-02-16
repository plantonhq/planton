# Global Key Ring

This preset creates a KMS key ring in the `global` location, making encryption keys accessible from any GCP region without latency penalties associated with cross-region access.

## When to Use

- Encryption keys shared across workloads in multiple GCP regions
- Global services that don't have data residency restrictions
- Shared signing keys for multi-region CI/CD pipelines
- Application-level encryption where key location is not regulated

## Key Configuration Choices

- **Global location** — keys are accessible from any GCP region. No data residency guarantees.
- **Simplicity** — no need to match key ring location with workload location.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID where the key ring will be created | GCP Console or `GcpProject` outputs |
| `<your-key-ring-name>` | Permanent name for this key ring (1-63 chars, letters/digits/hyphens/underscores) | Choose a descriptive name (e.g., `global-shared-keys`) |

## Important

Key rings **cannot be deleted** from GCP. The name you choose is permanent within the project and location.

## Related Presets

- **01-regional-key-ring** — Key ring in a specific region (data residency compliance)
- **03-multi-region-key-ring** — Key ring replicated across a continent (availability + residency)

# Short-Lived RSA Issuance

This preset creates a global issuance config that issues RSA-2048 certificates from a referenced private CA pool, valid for the 21-day minimum and renewed on day 14. Use it when some clients cannot speak ECDSA and you want keys to live as briefly as Google allows.

## When to Use

- Older clients, appliances, or libraries that only accept RSA certificates
- A security policy that prefers the shortest certificate lifetime
- A config that live certificates depend on and must not be destroyed by accident

## Key Configuration Choices

- **`RSA_2048`** — the widest client compatibility
- **21-day lifetime** — the minimum Google allows
- **Rotation at 66%** — the latest point the 21-day lifetime permits; validation refuses anything outside 34–66
- **`deletionPolicy: PREVENT`** — destroy fails rather than leaving certificates with no renewal policy

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID | `GcpProject` outputs |
| `<ca-pool-resource-name>` | The `GcpPrivateCaPool` resource the certificates are issued from | Your InfraChart or Planton resources |

## Related Presets

- **01-thirty-day-ecdsa-issuance** — P-256 keys with the maximum lifetime

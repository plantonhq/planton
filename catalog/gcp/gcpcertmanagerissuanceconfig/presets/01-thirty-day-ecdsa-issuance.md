# Thirty-Day ECDSA Issuance

This preset creates a global issuance config that issues P-256 certificates from a referenced private CA pool, valid for 30 days and renewed automatically on day 20. It is the everyday shape for internal service TLS.

## When to Use

- Internal load balancers or service-to-service TLS that must chain to your own root
- Every client speaks ECDSA (any modern client does)
- One shared policy for many Google-managed certificates

## Key Configuration Choices

- **`ECDSA_P256`** — smaller keys and faster handshakes than RSA
- **30-day lifetime** — the maximum Google allows; renewal is automatic, so there is no manual burden
- **Rotation at 66%** — renewal 10 days before expiry, leaving time to notice a failed issuance

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID | `GcpProject` outputs |
| `<ca-pool-resource-name>` | The `GcpPrivateCaPool` resource the certificates are issued from | Your InfraChart or Planton resources |

## Related Presets

- **02-short-lived-rsa-issuance** — RSA-2048 for older clients, with the minimum lifetime

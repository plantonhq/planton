# DevOps Workload TLS

## Use Case

A pool for short-lived service certificates -- a mesh or workload identity minting a certificate per workload that lives a week at most.

## When to Use

- Service-to-service mTLS where certificates rotate often and revocation is not needed
- High issuance rates (DevOps authorities issue about 3.5 times faster than Enterprise)

## What This Creates

- A DevOps-tier pool in `us-central1` issuing 7-day, P-256 leaf certificates that carry the requested SANs and server-and-client authentication

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `issuancePolicy.maximumLifetime` | `604800s` (7 days) | Shorter rotates faster; DevOps certificates cannot be revoked, so keep it short. |
| `issuancePolicy.allowedKeyTypes` | P-256 | Add `rsa` if clients cannot use EC keys. |
| `issuancePolicy.backdateDuration` | `3600s` | Absorbs clock skew on clients; at most `172800s`. |

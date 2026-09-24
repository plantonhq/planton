# mTLS Client

## Use Case

One-day client certificates for workloads that authenticate to each other with SPIFFE identities -- the subject is discarded, and every SAN must be a URI in your trust domain.

## When to Use

- Service-to-service mTLS with SPIFFE-style workload identities
- Kafka or Redis clients authenticating with certificates

## What This Creates

- A template in `us-central1` stamping CA:FALSE and client authentication, accepting only `spiffe://example.org/` URIs, and passing through only the extended key usage extension

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `identityConstraints.celExpression.expression` | `spiffe://example.org/` | Your trust domain. |
| `maximumLifetime` | `86400s` (1 day) | Workloads that re-issue less often need longer. |

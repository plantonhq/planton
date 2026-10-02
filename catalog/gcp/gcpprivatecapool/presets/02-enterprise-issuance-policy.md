# Enterprise Issuance Policy

## Use Case

A governed pool for server certificates that live up to a year: Google stores and can revoke every certificate, publishes CRLs, and a CEL rule keeps every name inside your internal domain.

## When to Use

- Internal HTTPS and gRPC servers with year-long certificates
- Compliance regimes that require revocation and at-rest encryption with your own key

## What This Creates

- An Enterprise-tier pool in `us-central1`, encrypted with the `pki-at-rest` KMS key, publishing CA certificates and CRLs, issuing RSA-2048+ or P-256 server certificates only for names under `.internal.example.com`

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `issuancePolicy.identityConstraints.celExpression.expression` | internal DNS suffix | Match your domain, or allow URIs for SPIFFE identities. |
| `issuancePolicy.maximumLifetime` | `31536000s` (1 year) | Shorter lifetimes limit exposure; public-style 90 days is `7776000s`. |
| `kmsKeyName` | `GcpKmsKey` reference | Drop it to keep Google-managed encryption. |

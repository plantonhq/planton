# GcpPrivateCaPool Guide

The judgment this guide protects: a pool is the thing relying parties trust, so it is designed once, for the lifetime of the trust relationship -- the tier and the policy outlive any single authority inside it.

## DevOps or Enterprise

Pick by what you need to do with a certificate after it is issued.

- **`DEVOPS`** -- a service mesh or workload identity minting thousands of certificates that live hours or days. Google does not store them, so there is nothing to list, describe, or revoke; you rely on short lifetimes instead of revocation. Authorities use Google-managed HSM keys only. It issues about 3.5 times faster per authority.
- **`ENTERPRISE`** -- device, user, or server certificates that live months or years, where revoking one matters. Google stores every certificate, publishes CRLs, and lets authorities use your own Cloud KMS key. It is the only tier a `GcpPrivateCaCertificate` works with, because a certificate managed as code must be readable and revocable.

The tier cannot change. Migrating means a new pool and moving relying parties to trust it.

## Policy lives on the pool

Put the rules on the pool, not on each request: `allowedKeyTypes` (for example P-256 only), `maximumLifetime`, `allowedIssuanceModes` (CSR, structured config, or both), and `identityConstraints` with a CEL expression such as "every SAN is a DNS name under `.internal.example.com`". `baselineValues` stamps key usage and constraints onto every certificate; a request cannot override them, and a template that conflicts with them fails.

## Rotation happens inside the pool

A pool issues through every enabled authority in it. To rotate: add a new authority, let relying parties pick up its certificate (the pool's CA certificates include both), enable it, then disable and later delete the old one. Clients that trust the pool never notice.

## Presence matters in `caOptions`

`isCa` and `maxIssuerPathLength` are presence-based: leaving one out omits it from the certificate, which is not the same as `false` or `0`. For a pool issuing leaf certificates, set `isCa: false`; for a root pool that signs subordinates, set `isCa: true` and `maxIssuerPathLength: 0` so the subordinates cannot sign further authorities.

## Tearing down

Google refuses to delete a pool that still holds an authority, including one in its 30-day soft delete. Destroy the authorities with `skipGracePeriod: true` first when the pool goes too.

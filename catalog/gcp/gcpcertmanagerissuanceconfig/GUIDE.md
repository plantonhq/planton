# GcpCertManagerIssuanceConfig Guide

The judgment this guide protects: an issuance config is shared policy, not a per-certificate setting. Every certificate that names it inherits its key algorithm, lifetime, and renewal point, and changing any of them means a new config. Decide the policy once, then let many certificates reuse it.

## The chain, in order

1. A `GcpPrivateCaPool`.
2. An enabled `GcpPrivateCaCertificateAuthority` in that pool.
3. The Certificate Manager service agent, `service-<project_number>@gcp-sa-certificatemanager.iam.gserviceaccount.com`, granted `roles/privateca.certificateRequester` on the pool.
4. This config, `caPool` from the pool's `name` output.
5. `GcpCertManagerCert` resources whose `managed.issuance_config` references this config's `issuance_config_id`.

The config creates without steps 2 and 3. Certificates that name it are what fail, so a missing grant shows up as a certificate stuck in provisioning, not as an error here. No catalog kind grants on a CA pool yet; make that grant outside the catalog, once per project.

## Choosing the policy

- **Key algorithm** — `ECDSA_P256` unless a client cannot speak it; `RSA_2048` for old clients.
- **Lifetime** — 21 to 30 days. Renewal is automatic, so the only cost of a shorter lifetime is more issuance from the pool.
- **Rotation window** — Google renews when this percentage of the lifetime has passed, and demands at least 7 days on each side. With 30 days the range is 24–76; with 21 days it is 34–66. Validation computes the rule for any lifetime, so an out-of-range window is refused before deploy.

## Everything but labels replaces

The pool, key algorithm, lifetime, window, and location are all immutable. A change creates a new config, and certificates still naming the old one keep serving until their own renewal. To change policy for live certificates, create a second config, point the certificates at it, then remove the first.

## Conventions and gotchas

- A global config may name a regional pool; the location here follows the certificates, not the pool.
- Google refuses to delete a config while a certificate references it. Set `deletionPolicy: PREVENT` while certificates depend on it.
- Each issued certificate counts toward the pool's Certificate Authority Service issuance billing.

## Pairs well with

- `GcpPrivateCaPool` — the pool; its `name` feeds `caPool`.
- `GcpPrivateCaCertificateAuthority` — the enabled authority that signs.
- `GcpCertManagerCert` — the consumer; it names this config by `issuance_config_id`.

# GcpCertManagerIssuanceConfig Guide

The judgment this guide protects: an issuance config is shared policy, not a per-certificate setting. Every certificate that names it inherits its key algorithm, lifetime, and renewal point, and changing any of them means a new config. Decide the policy once, then let many certificates reuse it.

## The chain, in order

1. A `GcpPrivateCaPool`.
2. An enabled `GcpPrivateCaCertificateAuthority` in that pool.
3. This config, `caPool` from the pool's `name` output.
4. `GcpCertManagerCert` resources whose `managed.issuance_config` references this config's `issuance_config_id`.

## The service agent and its grant

Google requires the Certificate Manager service agent, `service-<project_number>@gcp-sa-certificatemanager.iam.gserviceaccount.com`, to hold `roles/privateca.certificateRequester` on the pool when the config is created; without it the create fails with "Config validation failed" (measured on both engines). Google creates that agent only on request, never when the API is turned on, and a grant naming it before then is refused as an invalid member.

This block does both, in order: it creates the agent in the config's project, grants it the requester role on the pool (in the pool's own project, parsed from its name), and then creates the config. The deploying identity therefore needs `privateca.caPools.getIamPolicy` and `setIamPolicy` on the pool, beside the Certificate Manager permissions. The agent outlives the config; destroy never removes it.

The grant is a member grant, not counted per config. Two issuance configs over one pool share the one binding: destroying either removes it, and the other's certificates stop renewing until the other config is applied again. Keep pool and configs in one stack, or re-apply the survivor after removing one.

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

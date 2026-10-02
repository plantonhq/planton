# GcpPrivateCaCertificateAuthority Guide

The judgment this guide protects: an authority's certificate is immutable and trusted by things you do not control, so every choice here is a rotation plan, and losing one by accident must be hard.

## Root, subordinate, or both

Run two tiers for anything serious: a root in its own pool that signs only subordinates (`maxIssuerPathLength: 0` on the subordinates' certificates), and subordinates in the issuing pools. The root's key is used rarely; a compromised subordinate is revoked and replaced without re-trusting anything. A single self-signed authority is fine for a lab or a short-lived mesh.

- **Signed by reference** -- set `type: SUBORDINATE` and `subordinateConfig.certificateAuthority` to the root. The module has the root sign this authority's CSR and activates it on create.
- **Signed by your existing CA** -- apply once with `type: SUBORDINATE` and no activation fields; the authority waits in `AWAITING_USER_ACTIVATION`. Fetch its CSR from Google, have your CA sign it, then set `pemCaCertificate` and `subordinateConfig.pemIssuerChain` and apply again.

## The key

`algorithm` creates a Google-managed Cloud HSM key: nothing to manage, nothing to audit. `cloudKmsKeyVersion` uses a key you own -- your KMS audit trail, your rotation, your IAM -- and only Enterprise pools accept it. Prefer `EC_P384_SHA384` for a root and `EC_P256_SHA256` for subordinates unless a relying party needs RSA.

## Lifetimes

A certificate never outlives the authority that signed it. Give the root the longest lifetime (10-20 years), subordinates shorter (a few years), and rotate subordinates well before they expire: issued certificates are cut short at the subordinate's expiry.

## States

`STAGED` creates a root that is trusted (its certificate is in the pool's CA certificates) but not yet issuing -- stage a new root, distribute its certificate, then enable it. `DISABLED` stops issuance without deleting; it is refused on a new authority, so disable only after create.

## Deleting

Three guards, in order: `deletionProtection` (true until you set it false and apply), then `ignoreActiveCertificatesOnDeletion` (Google refuses while issued certificates are valid), then the 30-day soft delete, during which Google can restore the authority and the pool cannot be deleted. `skipGracePeriod` removes the last guard and is irreversible -- use it for labs and teardown pipelines, never for a production root.

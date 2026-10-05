# DNSSEC-Signed Public Zone

This preset creates a public Route53 hosted zone with DNSSEC signing enabled: Route 53 signs the zone's records with a key-signing key (KSK) backed by an asymmetric KMS key, protecting resolvers from spoofed and cache-poisoned answers. Signing the zone is HALF the chain of trust — the `ds_record` output must also be registered with the parent (your domain registrar) before any resolver validates the zone.

## When to Use

- Compliance regimes that mandate DNSSEC (many government and financial requirements)
- High-value domains where DNS spoofing is a real attack vector (payment endpoints, SSO/IdP hostnames)
- Any zone whose parent domain is already signed and expects a signed delegation

## Key Configuration Choices

- **KMS-backed KSK** — the signing key is yours, in your account. The key must live in **us-east-1** (the Route 53 DNSSEC control plane region), be an asymmetric **ECC_NIST_P256** key with **SIGN_VERIFY** usage, and its key policy must allow the `dnssec-route53.amazonaws.com` service principal (`kms:DescribeKey`, `kms:GetPublicKey`, `kms:Sign`, `kms:CreateGrant`)
- **KSK status ACTIVE** — the steady state. Flip to `INACTIVE` only while diagnosing signing problems: an inactive key stops signature refresh, so a signed zone with an inactive KSK eventually fails validation
- **Chain of trust completes at the registrar** — after deployment, take the `ds_record` output (and its `key_signing_key_tag`) to your registrar's DS-record form. Until the parent carries the DS, resolvers have no trust anchor and DNSSEC protects nobody
- **Rotation boundary** — AWS allows a second concurrent KSK per zone for zero-downtime rotation; this resource models the one steady-state KSK. Rotate with AWS tooling (create the new KSK, wait for propagation, retire the old), then reconcile this resource to the new key

## Placeholders to Replace

- `<domain-name>` — the zone's domain (e.g., `secure.example.com`)
- `<aws-region>` — provider-connection region (Route 53 itself is global)
- `<kms-key-name>` — the `AwsKmsKey` resource meeting the us-east-1 / ECC_NIST_P256 / key-policy requirements

## Related Presets

- **01-public-zone** — Use when DNSSEC is not required (the common case)
- **02-private-vpc-zone** — Private zones cannot be DNSSEC-signed (an AWS constraint)

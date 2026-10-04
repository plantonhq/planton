# AwsIamSamlProvider — Kind Guide

Authored operational judgment for the SAML provider kind: the
design decisions behind the spec's shape, and what to know before
operating federation in production.

## Design decisions

- **The metadata document is NOT a secret.** It carries the IdP's
  issuer, endpoints, and PUBLIC signing certificates — IdPs serve it
  unauthenticated at a metadata URL. It rides the spec as plain text
  (1000-10000000 characters, AWS's own bounds), never a managed-secret
  reference.
- **The name comes from metadata.name and is write-once.** AWS has no
  rename: a name change replaces the provider, mints a NEW ARN, and
  silently breaks every role trust policy naming the old one — treat
  renames as migrations.
- **One kind, one resource.** The provider stands alone; roles
  reference its ARN output in their own trust policies (the
  AwsIamRole kind's free-form trust_policy arm), so the trust edge
  lives where IAM puts it.

## Operating federation in production

- **Rotate BEFORE valid_until.** The output is derived from the
  metadata's certificate expiry; an expired document fails every
  federated sign-in at once. IdP metadata updates are in-place
  applies.
- **The trust policy pairs with the provider ARN**: a role trusts
  `{"Federated": "<provider_arn>"}` with `sts:AssumeRoleWithSAML` and
  typically the `SAML:aud` condition pinned to
  `https://signin.aws.amazon.com/saml`.
- **The IdP side needs the pair too** — SAML assertions carry
  `role_arn,provider_arn` pairs; configure the IdP's AWS app with
  both ARNs (the provider's ARN output and each role's).
- **Console sessions cap at the role's max_session_duration** — raise
  it on the role, not here.
- **The import ID is the provider ARN.**

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

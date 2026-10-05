# AwsSsmParameter — Kind Guide

Authored operational judgment for the parameter kind: the design
decisions behind the spec's shape, and what to know before operating
parameters in production.

## Design decisions

- **The name is an explicit spec field, hard-required, no fallback.**
  Parameter names are hierarchical paths (`/prod/db/url`) and slashes
  cannot live in `metadata.name`. The spec enforces AWS's
  fully-qualified rule (a name containing `/` must start with one) at
  validate time; AWS's reserved-prefix rule (names beginning `aws` or
  `ssm`) is server-side only — the field comment teaches it.
- **Two value arms, cross-mapped onto the provider's two.** The secret
  `secure_value` arm renders as the provider's sensitive `value`
  (redacted everywhere); the plain `value` arm renders as
  `insecure_value` so ordinary config stays READABLE in plans — that
  argument's whole purpose. Exactly one arm is set, and SecureString
  requires the secret arm — both CEL rules, not conventions.
- **The provider's write-only `value_wo`/`value_wo_version` pair is
  excluded**: it is Terraform state-avoidance, and the platform's
  just-in-time managed-secret resolution is the equivalent posture —
  a second secret path would be redundant surface.
- **`overwrite` renders only when true.** The provider's unset
  behavior (fail on a pre-existing foreign name at create; overwrite
  your own updates) is the safe default, and an explicit false would
  break the provider's own update path.

## Operating parameters in production

- **Advanced → Standard is a replacement, not an update.** AWS forbids
  the downgrade in place; the provider forces a new parameter. Going
  Standard → Advanced updates in place.
- **Intelligent-Tiering never persists** — AWS resolves it to Standard
  or Advanced per write, so reads (and the `tier` output) report the
  resolved tier. Expect a benign one-time reconcile if you pin
  Intelligent-Tiering and re-import.
- **`allowed_pattern` governs FUTURE writes only** — it does not
  validate the value already stored.
- **`data_type: aws:ec2:image` makes AWS verify the value is a real
  AMI ID in your account/region on every write** — a wrong-region AMI
  fails the write, not the launch that reads it. The data type forces
  replacement when changed.
- **StringList is one string** — comma-separated values, no escaping
  mechanism. Values containing commas need String, not StringList.
- **Standard parameters are free**; Advanced bills per parameter-hour
  and unlocks 8KB values plus parameter policies (expiration,
  no-change notification) — policies are an Advanced-only, out-of-band
  surface this kind does not manage today.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

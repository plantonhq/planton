# AwsIamAccountSettings — Kind Guide

Authored operational judgment for IAM's account-settings singleton:
the design decisions behind the spec's shape, and what to know before
operating account-wide IAM settings in production.

## Design decisions

- **A settings singleton for a GLOBAL service**: the identity is the
  ACCOUNT (not account+region like the SES/Bedrock members of the
  class). Deploy at most one instance per account; two instances
  fight over the same account objects. The spec's region is only the
  provider endpoint.
- **Three arms, three destroy contracts, each taught on its arm**:
  the alias truly DELETES (sign-in reverts to the bare account ID);
  the password policy RESETS to AWS defaults (reads back as "no
  custom policy" — NoSuchEntity); the STS preference is a NO-OP
  delete (the last-applied version persists; reverting is an apply
  with the other version).
- **The password policy is replaced WHOLE on every apply** — AWS's
  update is a full PUT, so an unset field means AWS's default, never
  "keep the current value". Presence-typed knobs (minimum length,
  self-service changes, age, reuse) render only when set; plain
  toggles render unconditionally (false and unset are the same
  posture at AWS).
- **Centralized root-access management is deliberately NOT here.**
  IAM's organization features are a management-account act requiring
  iam.amazonaws.com trusted access — they fold into AwsOrganization,
  whose spec already models the trusted-access list. This kind stays
  fully account-local and fully provable.

## Operating account settings in production

- **THE trap: applying the alias REPLACES the existing one.** An
  account has exactly one alias, aliases are globally unique across
  all of AWS, and everyone's bookmarked sign-in URL changes the
  moment a new alias lands. Check `aws iam list-account-aliases`
  before first apply.
- **The password policy has no partial update** — capture the current
  posture in the spec BEFORE adopting an account, or the first apply
  silently drops any setting the spec omits back to AWS defaults.
- **hard_expiry locks users out at expiry** (admin reset required) —
  pair it with allowUsersToChangePassword and a real maxPasswordAge
  or expect lockout tickets.
- **v1Token vs v2Token**: v2 tokens are larger but valid in EVERY
  region, including opt-in ones; accounts using opt-in regions should
  run v2. The setting affects the GLOBAL endpoint only — regional STS
  endpoints always issue v2.
- **Import stories differ per arm**: the alias imports by the alias
  string, the password policy by the provider's literal word
  (`iam-account-password-policy`), and the STS preference has no
  importer at the pin — adopters re-apply the version (an idempotent
  set-preferences call).

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

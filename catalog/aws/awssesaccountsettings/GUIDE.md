# AwsSesAccountSettings — Kind Guide

Authored operational judgment for the SES account-settings singleton:
the design decisions behind the spec's shape, and what to know before
managing account-wide email posture.

## Design decisions

- **A settings singleton, not a configuration-set field.** Suppression
  and VDM are attributes of the ONE SES account object per region —
  folding them into the per-set AwsSesConfigurationSet would make
  multiple sets fight over one account object (the recorded judgment
  this kind exists by).
- **Arms are optional and independent.** An omitted arm leaves the
  account's current setting untouched — omission is configuration.
  The CEL requires at least one arm: an instance managing neither is
  dead config.
- **An empty suppression list is modeled deliberately.** Present-with-
  empty-reasons means "auto-suppression explicitly OFF"; absent means
  "not managed here". The two differ and both are real postures.

## Operating account email posture in production

- **THE trap: suppression outlives destroy.** SES has no delete for
  suppression attributes — destroying this component leaves the
  last-applied reasons in effect (the same settings-retention class
  as SES's per-identity feedback forwarding). To stop suppressing,
  apply an empty reasons list BEFORE destroying.
- **BOUNCE + COMPLAINT is the reputation default.** Removing BOUNCE
  risks re-sending to hard-bounced addresses, which is how sending
  reputations die. Deviate only with a concrete reason.
- **VDM is a billing lever, not a flag.** Enabling VDM starts AWS
  charges; the dashboard/guardian sub-toggles only matter while
  enabled. Destroy DOES reset VDM to disabled (the asymmetry with
  suppression is upstream behavior, taught on the spec).
- **Suppression is account-wide.** It applies to every identity and
  configuration set in the region, including ones other teams manage
  — coordinate before turning it off.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

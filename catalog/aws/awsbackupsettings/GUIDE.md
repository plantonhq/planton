# AwsBackupSettings — Kind Guide

Authored operational judgment for the Backup settings component: the
design decisions behind the spec's shape, and what to know before
operating the settings in production.

## Design decisions

- **A settings singleton, split out of the vault kind.** Both provider
  resources have ZERO schema edges to vaults, singleton identities
  (account / region), and no-op deletes — folding them into the
  many-instances-per-region vault kind would let two vaults fight over
  one settings object (the exact defect the SES account-settings
  precedent exists to prevent).
- **Two arms, one kind.** The global arm (account) and region arm
  (region) are independent product levers with an at-least-one CEL —
  the AwsSesAccountSettings two-arm shape. The scope difference is
  taught: set the global arm in exactly ONE instance account-wide.
- **Full-map semantics are the provider's, kept honest.** Both maps
  are Required at the provider and AWS returns every supported
  key/type on read — the spec comments teach listing everything you
  manage instead of hiding the perpetual-diff behavior.

## Operating the settings in production

- **Destroy reverts NOTHING.** Both deletes are provider no-ops — the
  last-applied values stay in effect indefinitely. The revert lever is
  an apply with the desired values, never a destroy.
- **Opt-ins gate backup plans silently.** A selection covering a type
  the region has not opted in simply never backs it up — when a plan
  "misses" resources, check this component before the plan.
- **List every type deliberately.** AWS returns the complete
  preference map on read; a type absent from your map shows as a
  perpetual plan difference. Copy the full set from
  `aws backup describe-region-settings` and flip the booleans you
  mean.
- **Management preferences are one-way at AWS**: once set they can be
  flipped per type but never cleared back to unset.
- **Cross-account backup is an organizations decision** — flipping
  `isCrossAccountBackupEnabled` affects every account in the
  organization's backup posture; treat it as a change-managed control.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

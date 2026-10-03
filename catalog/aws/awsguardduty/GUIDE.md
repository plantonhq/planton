# AwsGuardDuty — Kind Guide

Authored operational judgment for the GuardDuty component: the design
decisions behind the spec's shape, and what to know before operating
threat detection in production.

## Design decisions

- **One kind for the account's posture.** The detector, its features,
  filters, lists, export, and the org/member surface are one regional
  decision keyed by one detector ID. Malware Protection for S3 split
  out — its schema carries no detector edge at all (it protects a
  bucket).
- **Features are patches, modeled as such.** Detector features,
  org-configuration features, and member features have no importer
  and no-op deletes upstream — Create and Update are the same call.
  The spec lists them with a presence-typed `enabled` (unset =
  enabled: listing a feature means you want it), and the modules
  never diff-revert what AWS keeps.
- **The org vocabulary differs on purpose.** Detector/member features
  toggle ENABLED/DISABLED; organization features auto-enable
  NEW/ALL/NONE — two different questions ("is it on here" vs "which
  members get it"), kept as two shapes.
- **Admin and member postures are mutually exclusive** (CEL-enforced):
  an account either administers members or accepts an invitation,
  never both.
- **The publishing destination is deliberately untagged** — tags are
  ForceNew on it upstream, so a tag sweep would REPLACE findings
  export mid-audit.

## Operating threat detection in production

- **The singleton collides.** A detector already present (console
  onboarding, Organizations auto-enable) fails creation with
  "detector already exists". Adopt it by import or remove it — never
  a second instance.
- **Removing a feature from the spec reverts nothing.** AWS keeps the
  last-applied state of unlisted features; to turn a plan off, list
  it with `enabled: false`. The same holds for the org configuration
  after destroy — the posture survives as last applied.
- **Findings export needs BOTH consents before create**: the bucket
  policy (PutObject for `guardduty.amazonaws.com`) and the key policy
  (`kms:GenerateDataKey`); AWS rejects the destination otherwise.
- **One ACTIVE trusted list per detector.** AWS enforces it; keep
  spare lists `activate: false`. List files must stay readable —
  GuardDuty re-reads them on activation.
- **Members inherit the admin's finding frequency.** Setting
  `finding_publishing_frequency` on a member detector fights the org
  sync forever (a perpetual-diff class the modules avoid by sending
  it only when set).
- **Runtime-monitoring agent management installs software.** The
  EKS/ECS/EC2 agent sub-toggles deploy GuardDuty agents into your
  compute — turn them on deliberately, per environment.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

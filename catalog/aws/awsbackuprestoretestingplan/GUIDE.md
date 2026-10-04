# AwsBackupRestoreTestingPlan — Kind Guide

Authored operational judgment for the restore testing kind: the
design decisions behind the spec's shape, and what to know before
operating restore tests in production.

## Design decisions

- **Its own kind, not a backup plan arm.** Restore testing has zero
  schema edges to backup plans — its selections reference the TESTING
  plan by name, and the product surface (scheduled drills over
  already-existing recovery points) is independent of how those points
  were created.
- **Names are explicit spec fields** (`plan_name`, `selections.name`):
  AWS forbids hyphens AND periods here (letters, digits, underscores
  only), stricter than metadata.name conventions — CEL rejects invalid
  names at validate time.
- **ARNs-XOR-conditions is a CEL rule** because the provider enforces
  it resource-wide (its one `ExactlyOneOf` validator in the whole
  backup service): a selection covers explicit ARNs or tag-matched
  resources, never both.
- **AWS's empty-conditions artifact is normalized away**: the API
  returns `{"StringEquals": [], "StringNotEquals": []}` for absent
  conditions, and both the provider and the spec treat
  present-but-empty as absent.

## Operating restore tests in production

- **Random beats latest.** `RANDOM_WITHIN_WINDOW` exercises older
  recovery points too — the stronger proof that your retention window
  is actually restorable, not just yesterday's snapshot.
- **The role needs restore permissions per tested type** (e.g.
  `ec2:CreateVolume` for EBS tests) on top of the
  `backup.amazonaws.com` trust — AWS's managed
  `AWSBackupServiceRolePolicyForRestores` policy is the usual grant.
- **Tests bill as restores.** Every drill creates a real temporary
  resource; size `validation_window_hours` to what validation actually
  needs (the copy is deleted when it expires).
- **Several knobs are one-way at AWS** (timezone, start window,
  selection window, validation window, exclude vaults): once set they
  keep a value — plan to flip them, never to clear them.
- **Metadata overrides use lowercase keys on read** — author them
  lowercase (`availabilityzone`, not `AvailabilityZone`) so imports
  and plans stay clean.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

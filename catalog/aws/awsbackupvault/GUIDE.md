# AwsBackupVault — Kind Guide

Authored operational judgment for the backup vault kind: the
design decisions behind the spec's shape, and what to know before
operating vaults in production.

## Design decisions

- **One kind, two arms.** AWS's own API discriminates vault types with
  `VaultType`; the spec mirrors it as an exactly-one union
  (standard / air_gapped) instead of two kinds. The satellites (lock,
  policy, notifications) live INSIDE the standard arm because the
  provider's readers reject attaching them to any other vault type —
  the union makes the invalid shape unrepresentable.
- **`force_destroy` is deploy-side behavior, not an AWS attribute.**
  AWS never reports it back; an imported vault always shows it false,
  and the first apply is a no-op. It exists to drain recovery points
  at destroy — leave it false for anything holding real backups.
- **Lock MODE is decided by `changeable_for_days` alone**, and AWS
  never reports the field back. Unset = governance mode (privileged
  principals can remove the lock later); set = compliance mode. The
  spec presence-types it so the mode choice is always explicit.
- **The KMS key is fixed at creation** on both arms, and once AWS
  fills in its default the field cannot be cleared — pick the key
  before the first recovery point lands.

## Operating vaults in production

- **Compliance mode is a one-way door.** After `changeable_for_days`
  expires (minimum 3 days), the lock is IMMUTABLE: nobody — not the
  account root, not AWS support — can remove it or delete the vault
  until every recovery point exceeds max retention. Rehearse in
  governance mode; flip to compliance only with retention bounds you
  can live with for years.
- **Empty vaults are free; recovery points are the bill.** Vault count
  is never a cost lever — retention windows in the backup plan's
  lifecycle rules are.
- **Air-gapped vaults fill through plan rules**, not directly: a
  rule's `target_logically_air_gapped_backup_vault_arn` copies
  recovery points in (the rule still needs its standard target vault).
- **A permissions failure can masquerade as a missing vault.** The
  provider maps `AccessDeniedException` to not-found on the vault's
  read path — if a vault "disappears" from plans, check IAM before
  assuming deletion.
- **Notifications need the topic's policy**, not just its ARN: SNS
  must allow `backup.amazonaws.com` to publish or events silently
  never arrive.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

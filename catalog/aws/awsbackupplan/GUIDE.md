# AwsBackupPlan — Kind Guide

Authored operational judgment for the backup plan component: the
design decisions behind the spec's shape, and what to know before
operating plans in production.

## Design decisions

- **Selections fold into the plan.** A selection references its plan's
  AWS-generated ID and cannot outlive it — a true satellite. Folding
  them kills the create/delete ordering problem: AWS refuses to delete
  a plan with live selections, and the provider retries the delete
  while the folded selections drain.
- **Lifecycle day counts are presence-typed** because the provider
  CANNOT transmit an explicit zero (zero is dropped as unset) —
  the spec makes "unset" the only honest way to say "no transition".
- **The 90-day cold-storage minimum is a CEL rule**, not a comment:
  AWS rejects `delete_after` fewer than 90 days beyond
  `cold_storage_after`, so the spec rejects it at validate time.
- **`recovery_point_tags` are metadata on recovery points**, not
  resource-identity tags — they deliberately bypass the module's
  identity tag map.
- **Selection names key everything**: the module's for_each address,
  the `selection_ids` output map, and the `{plan_id}|{selection_id}`
  import composite all use the spec's selection name.

## Operating plans in production

- **Retention is the bill.** Recovery-point storage is what AWS Backup
  charges for — set `lifecycle.delete_after_days` on every rule, and
  use cold storage for long retention (90-day minimum stay).
- **Continuous backups cap at 35 days** — keep `delete_after_days`
  within it on rules with `enable_continuous_backup: true`.
- **The selection role is trust-gated server-side**: a role that does
  not trust `backup.amazonaws.com` fails selection creation after a
  couple of minutes of provider retries (IAM eventual-consistency
  retry, then the real error).
- **Selections replace, never update.** Every selection field is
  ForceNew at the provider — a coverage change swaps the selection
  object (same name, new ID); backup history is unaffected.
- **Windows VSS applies to EC2 only** at the pinned provider, and
  malware scan actions/settings bill per scanned GB through GuardDuty
  — scope `resource_types` deliberately.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

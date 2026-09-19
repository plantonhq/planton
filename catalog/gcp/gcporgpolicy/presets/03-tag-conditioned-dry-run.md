# Tag-Conditioned Dry Run

The safe-rollout shape. The enforced policy forbids service account key
creation everywhere in the organization EXCEPT resources tagged
`environment=sandbox`; the dry-run policy audits what removing the
exemption would break, so the tightening can be watched before it bites.

## What it configures

- `policy.rules[0].enforce: true` — the unconditional rule a boolean
  constraint needs.
- `policy.rules[1]` — `enforce: false` with a `condition` on the tag: the
  exemption. Conditional rules must set the opposite of the unconditional one.
- `dryRunPolicy` — the stricter rule set in audit mode; violations appear in
  the audit log as `dryRunPolicyViolation`, nothing is blocked.

## Adjust before deploying

- **`scope.organizationId`** and the `condition.expression`'s org id — yours.
- **The tag** — an existing `GcpTagKey` / `GcpTagValue` pair; the short
  names in `resource.matchTag`, or ids via `resource.matchTagId`.
- Promote the dry-run rules into `policy` when the log is quiet.

## When to choose something else

A guardrail with no exemptions takes the **Disable Serial Port (Project)**
preset.

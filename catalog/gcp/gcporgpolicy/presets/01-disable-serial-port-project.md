# Disable Serial Port (Project)

The simplest guardrail: a boolean constraint enforced on one project. No VM
in the project can enable serial-port access. The shape that needs only
`roles/orgpolicy.policyAdmin` on the project -- no organization grant.

## What it configures

- `scope.projectId` — the project the policy applies to (omit the whole
  `scope` to use the provider's default project).
- `constraint` — a boolean constraint, so the rule uses `enforce`.
- `policy.rules[0].enforce: true` — the one unconditional rule a boolean
  constraint needs.

## Adjust before deploying

- **`scope.projectId`** — your project ID, or a `GcpProject` reference.
- **`constraint`** — any boolean constraint (`iam.disableServiceAccountKeyCreation`,
  `storage.publicAccessPrevention`).

## When to choose something else

A list constraint (regions, domains) takes the **Restrict Locations
(Folder)** preset; a rollout with exemptions takes **Tag-Conditioned Dry Run**.

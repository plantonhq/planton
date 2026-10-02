# GcpOrgPolicy Guide

The judgment this guide protects: an organization policy is a guardrail
the platform enforces on everyone, so the cost of getting it wrong is an
outage for every team beneath the scope. Roll out in audit mode, scope as
high as the rule is true, and let child scopes relax rather than parents
tighten.

## Know the constraint's type before you write the rule

Every constraint is boolean or list. A boolean constraint
(`compute.disableSerialPortAccess`, `iam.disableServiceAccountKeyCreation`)
takes a rule with `enforce`; a list constraint (`gcp.resourceLocations`,
`iam.allowedPolicyMemberDomains`) takes `values`, `allowAll`, or
`denyAll`. The type is not knowable offline, so the spec accepts every
shape and Google rejects a mismatch at apply time.
`gcloud org-policies list-constraints --project <id>` shows the type.

## Dry run first

`dryRunPolicy` evaluates the same rules in audit mode: violations are
logged (`dryRunPolicyViolation` in the audit log), nothing is blocked.
Roll out a new guardrail as a dry-run policy alone, read what it would
have broken, then add the same rules to `policy`. Keep both while a
tightening is in flight so the console shows the difference.

## Scope high, relax low

Enforce at the highest scope where the rule is universally true -- the
organization or an environment folder -- and let the exceptions be
policies at lower scopes. `enforce: false` on a boolean constraint at a
project is a real rule: it says "not here", and it wins over the parent's
`enforce: true` because the closest policy decides. For list constraints,
`inheritFromParent: true` keeps the parent's values in force beside the
child's additions; false makes the child the new root.

## One rule without a condition, then conditions

A boolean constraint needs exactly one unconditional rule; conditional
rules by tag (`resource.matchTag('123456789012/environment', 'sandbox')`)
must set `enforce` to the opposite and refine it. That is how one policy
enforces a guardrail everywhere except tagged sandboxes. The tags come
from `GcpTagKey` / `GcpTagValue` / `GcpTagBinding`.

## Custom constraints are references, not copies

An organization-defined rule is a `GcpOrgPolicyCustomConstraint`, declared
once at the organization; a policy enforces it with `customConstraint`
pointing at the constraint's `constraint` output. Three folders that need
the rule are three small policies referencing one constraint -- never
three copies of the rule.

## The name is the identity

`{scope}/policies/{constraint}` is the policy. Both halves are immutable
(a change recreates the policy), and one manifest per constraint per
scope is the rule -- a second collides with the first. `reset` is the
only way to say "back to the constraint's default here"; it stands alone,
with no rules and no inheritance flag.

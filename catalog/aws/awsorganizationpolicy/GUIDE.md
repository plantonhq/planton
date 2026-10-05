# AwsOrganizationPolicy — Kind Guide

Authored operational judgment for the organization-policy kind:
the design decisions behind the spec's shape, and what to know before
operating guardrails in production.

## Design decisions

- **Attachments fold into the policy.** An attachment is
  `{policy, target}` and nothing else — a pure edge that cannot
  outlive the policy. Each entry is its own resource keyed by target
  (half of the `{target_id}:{policy_id}` import composite); both
  leaves ForceNew, so a target change re-attaches.
- **One kind serves all thirteen policy types.** The type is a single
  immutable enum and the content is a structured document in the
  type's own syntax — the provider's shape, mirrored honestly. The
  type gate (enabled on the organization first) is AWS state; the spec
  documents it rather than pretending to validate it.
- **The name is an explicit spec field** — policy names allow spaces
  `metadata.name` cannot carry (the family convention).
- **`skip_destroy` is deliberately unmodeled** on both the policy and
  its attachments (the recorded apply-behavior exclusion class):
  destroy means detach and delete. AWS's own minimum-one-attached-
  policy rule is satisfied by FullAWSAccess, which the provider never
  manages.

## Operating guardrails in production

- **SCPs cap, they never grant** — an SCP is an outer boundary on what
  IAM in the subtree can allow. Test new guardrails on an empty OU or
  a sandbox account before attaching high.
- **Denying `organizations:LeaveOrganization` is the foundational
  guardrail** (this kind's canonical example) — without it any
  member account's root can walk out of governance.
- **Root attachments are the widest blast radius** — a wrong deny at
  the root locks the whole estate out of an API, management account
  excluded (SCPs never bind the management account). Prefer first-level
  OU attachments with the root reserved for universal invariants.
- **Content updates apply in place and propagate in seconds** — there
  is no staged rollout; pair risky edits with a narrow attachment
  first, then widen.
- **The enabled-type gate fails at ATTACH time**, not policy-create
  time — a policy of a disabled type creates fine and every attachment
  then errors. The organization component's `enabledPolicyTypes` is
  the fix, not a retry.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

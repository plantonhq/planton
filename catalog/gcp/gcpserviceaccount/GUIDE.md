# GcpServiceAccount Guide

The judgment this component protects: identity is cheap to create and expensive
to recreate. `serviceAccountId` and the project are replace-on-change — a
"rename" silently invalidates every IAM binding and Workload Identity that
referenced the old email, and the recreated account has a new `unique_id`, so
nothing transfers. Choose the ID like a hostname, not a label.

## When to use the role lists vs first-class grants

`projectIamRoles` / `orgIamRoles` are the convenience for "this identity plus
its obvious runtime roles" — they deploy atomically with the account and
subtract cleanly (additive member grants only). They deliberately carry NO IAM
conditions: a conditioned grant is a decision another team may need to see,
review, and own, so it belongs to a first-class `GcpProjectIamMember`
referencing this account's `member` output. Choosing the list where a
condition is needed fails at design time (the field does not exist), which is
intentional — not a gap.

## Keys: presence is the decision

The `userManagedKey` block's presence creates the key; omission is keyless.
Judgment earned from the flows, not the schema:

- `keepers` is the only declarative rotation you get. A key without a
  `keepers` entry will silently live forever; a key with
  `keepers: {rotation: 2026-Q3}` is rotated by editing one value — the old
  key is destroyed in the same apply, so consumers reading `key_base64` from
  outputs pick up the new key on their next resolution while anything
  that copied the key out-of-band breaks. That breakage is the rotation
  working as designed.
- `publicKeyData` (the upload flow) is the strongest posture when a key is
  unavoidable: GCP never returns private material and `key_base64` stays
  EMPTY — do not wire consumers to that output in this flow.
- `deletionPolicy: PREVENT` on the key makes every destroy of this component
  fail while the key exists. Reserve it for keys whose loss is an outage;
  it blocks the E2E-style full lifecycle by design.

## Conventions and gotchas

- `create_ignore_already_exists` adopts an existing same-email account
  instead of failing — the idempotent-bootstrap knob. The adopted account's
  destroy then DELETES it, so only adopt identities this manifest should own.
- The module wires project-role grants to the CREATED account's home project
  (not the raw spec value), so grants stay correct when `projectId` fell back
  to the ambient default project.

## On the diagram

The account is the hub node other components reference: its `member` output
feeds `iamMembers` lists (buckets, topics, secrets) and `GcpProjectIamMember`
nodes; a `GcpGkeWorkloadIdentityBinding` node renders the KSA→GSA edge.
Role-list grants render as part of THIS node — invisible as dependencies.
Grants that should be visible edges belong in first-class IAM member kinds.

## Pairs well with

- `GcpProjectIamMember` — conditioned or independently-owned grants against
  `member`.
- `GcpGkeWorkloadIdentityBinding` — keyless GKE workloads (the reason to
  leave `userManagedKey` out).
- `GcpIamCustomRole` — its `name` output is a valid entry in the role lists.

# AwsIamGroup — Kind Guide

Authored operational judgment for the IAM group kind: the design
decisions behind the spec's shape, and what to know before operating
groups in production.

## Design decisions

- **Membership is modeled ONCE, group-centric.** The provider offers
  the same edge from both sides (`aws_iam_group_membership` group-side
  authoritative, `aws_iam_user_group_membership` user-side additive);
  a declarative catalog needs one representation, and the group owning
  its members list is it. The user-side resource is recorded as
  composed coverage of the same edge.
- **The authoritative form is deliberate**: the whole users list rides
  ONE membership resource, so out-of-band additions are REMOVED on the
  next apply — the group's membership is exactly what the spec says.
  Clearing the list removes every membership.
- **Attachments key by the policy ARN, never the list index** —
  reordering `managedPolicyArns` is a no-op, not a transient
  detach/re-attach on a live group (the same idiom as AwsIamRole).
- **The exclusive-lockdown variants are deferred, not modeled.** The
  provider's `group_policies_exclusive` / `_attachments_exclusive`
  purge out-of-band policies at apply with a no-op delete —
  engine-workflow surface, and the kind already declares its full
  intended policy set (the recorded exclusive-set class, third
  application).

## Operating groups in production

- **Renames update in place** (metadata.name → UpdateGroup): members
  and policies persist, the ARN recomputes — anything pinning the old
  ARN (rare; policies usually name paths) needs a follow-up.
- **Groups are for USERS only** — roles cannot join groups; users
  assume roles individually.
- **Deletion order is handled**: IAM refuses to delete a non-empty
  group, and the module's membership/policy resources unwind first by
  dependency — a green destroy proves the whole unwind.
- **The users must exist before apply** — IAM rejects unknown names,
  and the reference wiring (AwsIamUser's user_name output) is the
  ordering guarantee in charts.
- **Import honesty**: the group, inline policies, and attachments all
  import; the membership resource ships NO importer at the pinned
  provider — adopters re-apply the declarative users list (an
  idempotent reconcile), recorded in the provider import catalog.
- **Groups are untaggable** — account-wide conventions ride the path
  (e.g. `/teams/`) instead.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

# GcpCloudIdentityGroup Guide

The judgment this guide protects: a group is the indirection that makes
access manageable. Bind roles to groups and put people in groups; the day
someone leaves, one membership edit closes every door, instead of a hunt
through every project's IAM policy.

## Groups are the unit IAM should name

Every `GcpProjectIamMember`, folder binding, and organization binding
should take `group:{groupEmail}`, not `user:{someone}`. The group's email
IS its identity -- pick it as a stable role name (`platform-admins@`,
`billing-viewers@`), not a person's or a project's, because the email is
immutable and every binding that names it breaks if the group is
recreated.

## Security groups are one-way

`security: true` adds Google's security label: the group becomes an
access-control object (no mailing-list features, stricter membership
rules) and the label can never be removed. Make the choice at creation
for any group IAM will bind; leave plain groups for discussion lists.

## Members and roles

Each membership is a user, a service account (a `GcpServiceAccount`
reference resolves to its email), or another group, by email. Every
membership carries `MEMBER`; add `MANAGER` (manages members) or `OWNER`
(full control, including deleting the group) on top. Only the `MEMBER`
role may `expireTime` -- the shape for contractors and break-glass access.
The manifest's list is the managed membership: a member removed here is
removed from the group; members added out of band (through the Admin
console or `WITH_INITIAL_OWNER`) are not touched. `initialGroupConfig:
WITH_INITIAL_OWNER` makes the deploying principal an owner so the group is
never orphaned; `EMPTY` (the default) trusts the declared list.

## The group lives under the customer

Groups belong to the Cloud Identity or Workspace customer
(`customers/{id}`), not to a project: the deploying principal needs the
Groups Admin role there, the email's domain must be verified, and the
group outlives every project that binds it.

## Deleting a group empties its bindings silently

An IAM binding to a deleted group grants nothing and complains to no one.
`deletionPolicy: PREVENT` is the guard for any group production bindings
name; `ABANDON` keeps the group while dropping management; removing
members is the surgical alternative to deleting the group.

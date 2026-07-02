# AWS IAM Instance Profile: How EC2 Gets an Identity

## What an Instance Profile Is

Most AWS services assume IAM roles directly: Lambda's execution role, an ECS
task role, an EKS pod identity all bind a role straight to the workload. EC2
is the exception. An instance is launched with an *instance profile* -- a thin
IAM container that holds exactly one role -- and the instance metadata service
(IMDS) vends that role's temporary credentials to whatever runs on the box.
The SDK default credential chain picks them up automatically; no keys ever
touch the instance.

`AwsIamInstanceProfile` models that container: a name, an optional IAM path,
and the one role it carries.

## Why a First-Class Component

The profile sits at a real seam in the identity graph, and both of its
neighbors are better off when it is its own node:

- **Roles stay universal.** A role is assumed by any AWS service its trust
  policy allows. Baking a profile into every role -- a common convenience
  hack -- would mint an EC2-only wrapper for Lambda roles, ECS roles, and
  service roles that will never see an instance. Only EC2 topologies need a
  profile, so only they create one.
- **EC2 references stay stable.** Instances, launch templates, and Auto
  Scaling groups reference the *profile*, not the role. Because the carried
  role can be swapped in place (AWS removes the old role and adds the new one
  without replacing the profile), an operator can rotate a fleet onto a new
  role without touching a single EC2-side reference -- running instances pick
  up the new credentials on their next IMDS refresh.
- **The chain is visible.** Role → profile → instance is three referenceable
  nodes in the architecture graph. When the profile is a hidden side effect of
  the role, the middle of that chain -- and the ability to own or change it --
  disappears.

## Lifecycle

- **Name and path are create-only.** Changing either replaces the profile
  (and forces dependent EC2 resources to re-reference it).
- **The role is swappable.** The one mutable part, by design -- see above.
- **One role per profile.** An AWS limit, mirrored in the spec: `role` is a
  single required reference, not a list.
- **Eventual consistency is handled.** IAM propagates slowly; attaching a
  freshly-created role can transiently fail. Both engines retry the attach
  internally, and deletion detaches the role first.

## The Role Reference

The profile takes the role by **name**, not ARN -- that is what the underlying
`AddRoleToInstanceProfile` API accepts. The spec's `role` field is a
`StringValueOrRef` that defaults to resolving an `AwsIamRole`'s `role_name`
output; a literal name works for roles that live outside Planton.

## Dual-Engine Implementation

`AwsIamInstanceProfile` ships both a Terraform/OpenTofu module and a Pulumi
(Go) module at behavioral parity. Both create the profile with the same
identity tags and export the same outputs (`instance_profile_arn`,
`instance_profile_name`, `instance_profile_id`, `role_name`). Whichever engine
a team standardizes on, the profile behaves identically.

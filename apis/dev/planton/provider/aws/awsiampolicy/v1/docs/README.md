# AWS IAM Policy: The Reusable Unit of Permissions

## What a Managed Policy Is

An IAM policy is a JSON document that grants or denies permissions. AWS
supports two ways to carry one: an *inline* policy, embedded directly in a
single role, user, or group and sharing its lifecycle; and a *managed* policy,
a standalone resource with its own ARN, its own version history, and a
many-to-many attachment model. `AwsIamPolicy` models the customer-managed
variant: a policy your account owns, defines, and attaches wherever it is
needed.

The distinction matters more than it first appears. An inline policy answers
"what can *this* principal do?" A managed policy answers "what does *this
permission set* mean?" -- and lets any number of principals adopt that meaning
by reference.

## Why a First-Class Component

Permission documents are the most duplicated artifact in a typical AWS estate:
the same "read from the artifacts bucket" or "write to the service's log
group" statements appear in role after role, drifting apart one edit at a
time. Modeling the managed policy as its own component fixes the root cause:

- **One definition, many attachments.** Roles and users reference the policy
  ARN through their `managedPolicyArns` fields (as a `valueFrom` reference or
  a literal ARN). Updating the document updates every attached principal at
  once.
- **Permissions boundaries become first-class.** A boundary is just a managed
  policy ARN applied as a ceiling on a role or user. With the policy as a
  component, the boundary an organization applies to its CI principals is
  visible, versioned, and referenced -- not a magic ARN string.
- **The architecture graph tells the truth.** When attachments are references,
  the resource graph shows exactly which principals carry which permission
  sets -- something a pile of inline documents can never show.

Inline policies still have a place -- permissions truly unique to one role stay
folded into that role's `inlinePolicies` map -- but anything reused belongs
here.

## Lifecycle and Versioning

The parts of a managed policy behave differently over time, and the component
mirrors AWS's real behavior:

- **The document is updatable.** Each change creates a new *policy version*
  and marks it default. AWS retains at most five versions per policy; both IaC
  engines prune the oldest non-default version before saving a new one, so
  continuous delivery of policy changes never hits the cap.
- **Name, path, and description are create-only.** AWS has no rename or
  re-describe API for managed policies. Changing any of them replaces the
  policy: the engines create the successor, move attachments, and delete the
  original.
- **Deletion detaches first.** A policy cannot be deleted while attached; the
  engines detach from all principals and remove non-default versions before
  deleting.

## The Path Hierarchy

The IAM `path` is an underused organizing tool: policies created under
`/service-boundaries/` can be matched in *other* IAM policies with a wildcard
resource (`arn:aws:iam::<account>:policy/service-boundaries/*`). That enables
patterns like "the platform team may attach only boundary policies" without
enumerating ARNs. The path is create-only and must begin and end with `/`.

## What This Component Deliberately Omits

- **Attachment resources.** AWS exposes standalone attachment resources
  (`aws_iam_role_policy_attachment` and friends), but a standalone attachment
  node carries no configuration of its own -- it is pure glue that would flood
  the resource graph. Attachments live as reference lists on the principals
  (`AwsIamRole.managedPolicyArns`, `AwsIamUser.managedPolicyArns`).
- **AWS-managed policies.** Amazon's own policies (e.g.
  `arn:aws:iam::aws:policy/ReadOnlyAccess`) already exist and need no
  component -- attach them by literal ARN alongside references to policies
  defined here.

## Dual-Engine Implementation

`AwsIamPolicy` ships both a Terraform/OpenTofu module and a Pulumi (Go) module
at behavioral parity. Both create the policy with the same identity tags,
encode the free-form document to the JSON string AWS expects, and export the
same outputs (`policy_arn`, `policy_id`, `policy_name`). Whichever engine a
team standardizes on, the policy behaves identically.

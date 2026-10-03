# AWS IAM Group

Creates an IAM group -- the container that grants a set of users one shared permission set, with membership declared as a list and permissions as managed-policy attachments plus inline documents. Membership is authoritative: the declared users ARE the group, and users added outside this resource are removed on the next apply. IAM is a global service, so the group exists account-wide; the spec's region is only the provider endpoint.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **IAM Group** -- the group itself, named from `metadata.name`, under the IAM path from `path` (default `/`)
- **Group Membership** -- created only when `users` is non-empty; ONE membership resource carrying the whole users list, which is what makes membership authoritative
- **Managed Policy Attachments** -- one per `managedPolicyArns` entry; each reconciles individually, so adding or removing an entry attaches or detaches just that policy
- **Inline Policies** -- one per `inlinePolicies` map entry; permission documents that live and die with the group

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with IAM permissions. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- The users in the membership list must already exist -- IAM rejects unknown names. Reference wiring to AWS IAM User resources is the ordering guarantee in charts.
- Custom managed policies referenced in `managedPolicyArns` come from AWS IAM Policy resources; AWS-managed policies (like `arn:aws:iam::aws:policy/ReadOnlyAccess`) attach as literal ARNs.

## Deploy

### Console

Open the deployment store, find **AWS IAM Group**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the spec: path, members, and policies. Start from the **Read-Only Team** preset in the [Presets](#presets) tab for the simplest useful group.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsIamGroup
metadata:
  name: auditors
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  path: /teams/
  users:
    - value: alice
    - value: bob
  managedPolicyArns:
    - value: arn:aws:iam::aws:policy/ReadOnlyAccess
```

```shell
planton apply -f aws-iam-group.yaml
```

This creates an `auditors` group under `/teams/` whose membership is exactly alice and bob, with AWS's maintained ReadOnlyAccess policy attached -- members see everything and touch nothing. An Infra Job tracks the provisioning in real time.

### InfraChart

When the group deploys alongside its users and policies in one chart, wire them via ValueFromRef:

```yaml
spec:
  region: us-east-1
  path: /teams/
  users:
    - valueFrom:
        kind: AwsIamUser
        name: alice
        fieldPath: status.outputs.user_name
  managedPolicyArns:
    - valueFrom:
        kind: AwsIamPolicy
        name: readonly-extras
        fieldPath: status.outputs.policy_arn
```

The InfraPipeline resolves the dependency graph, creates the users and policies first, then the group with its memberships and attachments.

## Key Configuration

These are the most important decisions when configuring an IAM group. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Membership is authoritative -- attachments are not** -- the whole `users` list rides one membership resource, so the group's membership is exactly what the spec says: console additions disappear on the next apply, and clearing the list removes every membership. Managed-policy attachments behave differently: each entry reconciles individually, and attachments made outside this resource are left alone. Know which contract you are relying on before you audit either.

**Managed versus inline policies** -- put anything reusable in `managedPolicyArns` (your own AWS IAM Policy resources, or AWS-maintained policies that gain new-service permissions without you editing anything). Reserve `inlinePolicies` for permissions unique to this one group -- the classic use is a narrow deny carve-out riding alongside a broad managed policy, since an explicit deny beats any allow.

**Reordering is a no-op** -- attachments key by policy ARN, never list index, so reordering `managedPolicyArns` causes no transient detach/re-attach on a live group. Adding and removing entries is safe surgery.

**Renames update in place** -- changing `metadata.name` calls IAM's rename: members and policies persist, but the ARN recomputes. Anything pinning the old ARN (rare -- policies usually match paths) needs a follow-up edit.

**Groups are for users only** -- IAM roles cannot join groups. If a permission set must serve both humans and workloads, model it as an AWS IAM Policy attached to both this group and the roles.

**Groups are untaggable** -- account-wide conventions ride the IAM `path` instead (e.g. `/teams/`), which policies can also match on.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsIamUser** | `users[]` | `status.outputs.user_name` |
| **AwsIamPolicy** | `managedPolicyArns[]` | `status.outputs.policy_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `group_name` | The group's name (also the provider's import ID) | The value IAM policies and budget actions reference the group by |
| `group_arn` | The group's ARN | Policy documents scoping `iam:*` actions to this group |

The `group_id` output (AWS's stable unique ID, which survives renames) is a record value for audit rather than a composition input.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Read-only team** -- declared members plus AWS's maintained ReadOnlyAccess policy: auditors, finance, support -- anyone who should see everything and touch nothing. The first group worth creating in any account moving off per-user policies; swap in `SecurityAudit` for a security-review shape. Start from the **Read-Only Team** preset.

**Developers with a billing fence** -- PowerUserAccess (everything except IAM and organization management) via the AWS-managed policy, plus an inline `deny-billing` document keeping billing, budgets, and cost tooling out of reach. Extend the deny list for other carve-outs (regions, services) rather than forking the managed policy. Start from the **Developers Group** preset.

**Group-per-team, policy-per-capability** -- keep groups as pure membership containers and express capabilities as AWS IAM Policy resources attached to whichever groups need them. Permissions changes then touch one policy, not N groups.

## Works With

- [**AWS IAM User**](/infra-catalog/aws-iam-user) -- the members, wired via `users[]` references to their `user_name` outputs
- [**AWS IAM Policy**](/infra-catalog/aws-iam-policy) -- reusable permission sets attached via `managedPolicyArns[]`
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) -- the workload-identity counterpart; roles cannot join groups, but share the same managed policies

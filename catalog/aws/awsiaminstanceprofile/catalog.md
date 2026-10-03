# AWS IAM Instance Profile

Deploys an IAM instance profile — the container that delivers an IAM role to EC2 instances. EC2 cannot assume a role directly: an instance can only be launched with an instance profile, which holds exactly one role, and the instance metadata service then vends that role's temporary credentials to whatever runs on the machine — no access keys on disk, credentials that rotate themselves. Everything EC2-shaped references the profile (an instance's profile field, a launch template, an Auto Scaling group), while everything else on AWS (Lambda, ECS, EKS) assumes the role directly. Modeling the profile as its own component keeps that boundary honest, and its `instance_profile_arn` output is what EC2-shaped resources reference via ValueFromRef.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Instance Profile** -- the named container under its IAM path. The profile name comes from `metadata.name`; name and path are create-only (changing them replaces the profile)
- **Role Attachment** -- the one role the profile carries, attached by NAME (that is what the AWS API takes). Swapping the role later detaches the old one and attaches the new one IN PLACE — every instance carrying the profile picks up the new credentials without any EC2 reference changing

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **An IAM role EC2 can assume** -- the role's trust policy must allow the `ec2.amazonaws.com` service principal, or launches fail with an unauthorized error. Reference an AwsIamRole Infra Component (preferred — the graph deploys it first) or name a role that exists outside Planton.

## Deploy

### Console

Open the deployment store, find **AWS IAM Instance Profile**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **EC2 Role Delivery** preset in the [Presets](#presets) tab to pre-populate the role-by-reference composition.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsIamInstanceProfile
metadata:
  name: web-server
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  role:
    valueFrom:
      kind: AwsIamRole
      name: web-server-role
      fieldPath: status.outputs.role_name
```

```shell
planton apply -f instance-profile.yaml
```

This wraps the role in a profile that EC2 instances, launch templates, and Auto Scaling groups can carry. An Infra Job tracks the provisioning in real time.

### InfraChart

When the profile deploys alongside its role in one chart, wire the role reference via ValueFromRef:

```yaml
spec:
  region: us-west-2
  role:
    valueFrom:
      kind: AwsIamRole
      name: web-server-role
      fieldPath: status.outputs.role_name
```

The InfraPipeline resolves the dependency graph, deploys the role first, then the profile — and downstream, the instances and launch templates that reference the profile's own outputs deploy last.

## Key Configuration

These are the most important decisions when configuring an instance profile. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**One role, by name** -- AWS allows exactly one role per profile, and the AddRoleToInstanceProfile API takes the role's NAME, not its ARN. A workload needing more permissions gets them on the role (more policies) — never a second profile. Reference an AwsIamRole's `role_name` output so the two components stay composed in the graph, or pass a literal name during incremental adoption while roles are still managed elsewhere.

**The role swaps in place** -- Repointing the profile at a different role updates every instance carrying it, without touching the EC2 references. That is the profile's operational superpower — and why the role is the profile's one editable lever.

**The path is the one one-way door** -- The IAM path (`/compute/`) is an ARN infix that IAM policies can match (e.g. restricting which profiles launch tooling may pass to instances). The ARN embeds it, so changing it replaces the profile. Most teams leave it as `/`.

**Region is a management detail** -- IAM is global: one profile serves instances in every region. The region only picks the API endpoint the provider calls while managing it.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsIamRole** | `role` | `status.outputs.role_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `instance_profile_arn` | Amazon Resource Name of the profile | AwsEc2Instance / AwsLaunchTemplate / AwsBatchComputeEnvironment profile fields |
| `instance_profile_name` | The friendly name (mirrors `metadata.name`) | APIs that take the profile by name rather than ARN |
| `instance_profile_id` | The stable unique ID AWS assigns (AIPA…) | Audit trails |
| `role_name` | The carried role's name, resolved from `spec.role` | Seeing the effective role without dereferencing the AwsIamRole |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**EC2 role delivery** -- Wrap an AwsIamRole by reference so web fleets, SSM-managed instances, and build runners get AWS API access with zero embedded keys. Start from the **EC2 Role Delivery** preset.

**Wrap an existing role** -- Carry a role that lives outside Planton by literal name — the incremental-adoption path while IAM is still managed elsewhere. Start from the **Wrap an Existing Role** preset.

## Works With

- [**AWS IAM Role**](/infra-catalog/aws-iam-role) -- the role this profile carries and delivers to EC2
- [**AWS IAM Policy**](/infra-catalog/aws-iam-policy) -- the permissions attached to the carried role
- [**AWS EC2 Instance**](/infra-catalog/aws-ec2-instance) -- launches with this profile as its AWS identity
- [**AWS Launch Template**](/infra-catalog/aws-launch-template) -- bakes the profile into the fleet blueprint
- [**AWS Auto Scaling Group**](/infra-catalog/aws-auto-scaling-group) -- fleets whose every instance carries the profile

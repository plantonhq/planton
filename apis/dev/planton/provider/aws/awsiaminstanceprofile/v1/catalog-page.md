# AWS IAM Instance Profile

Deploys an IAM instance profile: the container that delivers an IAM role to
EC2 instances. EC2 cannot assume a role directly -- it can only be launched
with a profile that carries one. The profile references an `AwsIamRole` and is
what EC2 instances, launch templates, and Auto Scaling groups attach.

## What Gets Created

When you deploy an AwsIamInstanceProfile resource, Planton provisions:

- **Instance profile** — an `aws_iam_instance_profile` / `iam.InstanceProfile`
  named from `metadata.name`, carrying the referenced role (by name, as the AWS
  API requires) and an optional IAM path.

The role itself is **not** created here — deploy an `AwsIamRole` component and
reference its `role_name` output.

## Prerequisites

- **AWS credentials** configured via the Planton provider config (keyless SSO/OIDC).
- **An IAM role** to carry: an `AwsIamRole` component (or the name of an
  existing role).

## Quick Start

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsIamInstanceProfile
metadata:
  name: web-server
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

## Configuration Reference

| Field | Type | Default | Description |
| --- | --- | --- | --- |
| `region` | `string` | — | AWS region for the provider's API calls. IAM is global — the profile is usable in every region. Required. |
| `role` | `StringValueOrRef` | — | The IAM role the profile carries, by **name** (not ARN). Reference an `AwsIamRole`'s `role_name` output or pass a literal name. One role per profile (AWS limit); swappable in place. Required. |
| `path` | `string` | `/` | IAM path for organizing and wildcard-matching profiles (e.g. `/compute/`). Must begin and end with `/`. Immutable. |

## Examples

### Profile for an Auto Scaling group's launch template

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsIamInstanceProfile
metadata:
  name: app-fleet
spec:
  region: us-west-2
  path: /compute/
  role:
    valueFrom:
      kind: AwsIamRole
      name: app-fleet-role
      fieldPath: status.outputs.role_name
```

### Wrapping a role that exists outside Planton

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsIamInstanceProfile
metadata:
  name: legacy-app
spec:
  region: us-west-2
  role:
    value: legacy-app-ec2-role
```

## Stack Outputs

| Output | Description |
| --- | --- |
| `instance_profile_arn` | ARN of the profile — what an EC2 instance's `iamInstanceProfileArn` references |
| `instance_profile_name` | Friendly name — launch templates take the profile by name |
| `instance_profile_id` | Stable unique ID AWS assigns to the profile (`AIPA...`) |
| `role_name` | Name of the IAM role the profile carries |

## Related Components

- [AwsIamRole](/docs/catalog/aws/iam-role) — the role this profile carries and delivers to EC2
- [AwsIamPolicy](/docs/catalog/aws/iam-policy) — the permission sets attached to that role
- [AwsEc2Instance](/docs/catalog/aws/ec2-instance) — launches with this profile to receive the role's credentials

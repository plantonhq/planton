# Wrap an Existing Role

This preset creates an instance profile carrying a role that already exists
outside Planton -- useful when adopting EC2 workloads incrementally while IAM
roles are still managed elsewhere.

## When to Use

- Migrating EC2 workloads onto Planton before their IAM roles move
- Roles owned by another team or tool that EC2 instances still need to carry
- Quick experiments against a pre-existing role

## Key Configuration Choices

- **Role by literal name** -- the AWS API attaches roles by name (not ARN), so
  the value is the bare role name, e.g. `legacy-app-ec2-role`
- **`/compute/` path** -- groups EC2-serving profiles so IAM policies can match
  them with `arn:aws:iam::<account>:instance-profile/compute/*`

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<aws-region>` | AWS region code (e.g., `us-east-1`) | Your deployment region |
| `<existing-role-name>` | Name of the pre-existing IAM role | IAM console > Roles |

## Related Presets

- **01-ec2-role-delivery** -- compose with an AwsIamRole by reference (preferred)

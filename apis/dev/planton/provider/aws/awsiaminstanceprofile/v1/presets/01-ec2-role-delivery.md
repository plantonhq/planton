# EC2 Role Delivery

This preset wraps an `AwsIamRole` in an instance profile so EC2 instances can
receive the role's temporary credentials through the instance metadata
service. Reference this profile's `instance_profile_arn` output from an
`AwsEc2Instance` (or a launch template / Auto Scaling group).

## When to Use

- Any EC2 instance that needs AWS API access without embedded keys
- SSM Session Manager access (the instance role needs the SSM managed policy,
  and the instance needs a profile to carry it)
- Fleets whose role should be swappable without touching EC2 references

## Key Configuration Choices

- **Role by reference** -- the profile resolves the `AwsIamRole`'s `role_name`
  output at deploy time, so the two components stay composed in the graph
- **Default path** -- omitted here; set `path` (e.g. `/compute/`) when IAM
  policies need to wildcard-match profiles by hierarchy

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<aws-region>` | AWS region code (e.g., `us-east-1`) | Your deployment region |
| `<role-resource-name>` | Name of the AwsIamRole resource to carry | Your AwsIamRole manifest's `metadata.name` |

## Related Presets

- **02-existing-role** -- wrap a role that exists outside Planton

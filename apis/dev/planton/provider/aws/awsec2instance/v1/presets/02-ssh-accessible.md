# SSH-Accessible Instance

This preset creates an EC2 instance accessible via traditional SSH through a bastion host or direct connection. It requires an EC2 key pair for authentication. Use this when your tooling or workflows require direct SSH access rather than AWS Systems Manager.

## When to Use

- Legacy environments or tools that require direct SSH access
- Development instances where engineers prefer `ssh` over `aws ssm start-session`
- Instances accessed through a bastion host or VPN with SSH forwarding

## Key Configuration Choices

- **Bastion connection** (`connectionMethod: BASTION`) -- Traditional SSH access via key pair; requires inbound port 22 on the security group
- **Key pair required** (`keyName`) -- EC2 key pair must be created in AWS before deployment
- **No termination protection** -- Development-oriented preset; add `disableApiTermination: true` for production
- **No IAM instance profile** -- Add `iamInstanceProfileArn` if the instance needs to access AWS services

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<aws-region>` | AWS region where the instance will be created (e.g., `us-west-2`) | AWS region list |
| `<instance-name>` | Name tag for the EC2 instance (e.g., `dev-server-01`) | Your naming convention |
| `<ami-id>` | Amazon Machine Image ID (e.g., `ami-0abcdef1234567890`) | AWS EC2 AMI catalog |
| `<private-subnet-id>` | Subnet ID where the instance will launch | AWS VPC console or `AwsVpc` status outputs |
| `<security-group-id>` | Security group ID (must allow inbound port 22) | AWS EC2 console or `AwsSecurityGroup` status outputs |
| `<ec2-key-pair-name>` | Name of the EC2 key pair for SSH authentication | AWS EC2 console > Key Pairs |

## Related Presets

- **01-ssm-managed** -- Use instead for keyless, port-less access via AWS Systems Manager (recommended for production)

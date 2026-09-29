# SSM-Managed Hardened Instance

This preset creates an EC2 instance accessible via AWS Systems Manager Session Manager, hardened to the modern baseline: IMDSv2 enforced, an encrypted gp3 root volume, and termination protection. SSM eliminates SSH keys, bastion hosts, and open inbound ports -- connections are brokered through the AWS control plane and fully audited.

## When to Use

- Production or staging instances where SSH key management overhead should be avoided
- Environments with strict security requirements that prohibit opening inbound SSH ports
- Any EC2 instance that needs secure, auditable shell access without a bastion host

## Key Configuration Choices

- **Instance profile by reference** (`instanceProfile`) -- The instance's AWS identity; the profile's role must carry `AmazonSSMManagedInstanceCore`. Reference an `AwsIamInstanceProfile`'s `instance_profile_name` output (the EC2 API takes the profile by name)
- **IMDSv2 enforced** (`metadataOptions.httpTokens: required`) -- The single most effective hardening against credential-stealing SSRF; hop limit 2 keeps containerized workloads working
- **Encrypted gp3 root** (`rootBlockDevice`) -- The current-generation volume type with encryption at rest
- **Termination protection** (`disableApiTermination: true`) -- Prevents accidental instance termination
- **t4g.small** (`instanceType`) -- Graviton price-performance; pair with an arm64 AMI

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<aws-region>` | AWS region where the instance will be created (e.g., `us-west-2`) | AWS region list |
| `ami-0123456789abcdef0` | Amazon Machine Image ID matching the instance type's architecture | AWS EC2 AMI catalog or `aws ec2 describe-images` |
| `<private-subnet-id>` | Private subnet ID where the instance will launch | `AwsSubnet` status outputs |
| `<security-group-id>` | Security group ID controlling instance traffic | `AwsSecurityGroup` status outputs |
| `<instance-profile-name>` | NAME of the IAM instance profile with SSM permissions | `AwsIamInstanceProfile` status outputs |

## Related Presets

- **02-launch-template** -- Launch from an org-wide golden template instead of inline configuration
- **03-spot-worker** -- Interruption-tolerant Spot capacity for batch and CI workloads

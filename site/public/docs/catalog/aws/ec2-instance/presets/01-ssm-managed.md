---
title: "SSM-Managed Instance"
description: "This preset creates an EC2 instance accessible via AWS Systems Manager Session Manager. SSM eliminates the need for SSH keys, bastion hosts, or open inbound ports -- connections are brokered through..."
type: "preset"
rank: "01"
presetSlug: "01-ssm-managed"
componentSlug: "ec2-instance"
componentTitle: "EC2 Instance"
provider: "aws"
icon: "package"
order: 1
---

# SSM-Managed Instance

This preset creates an EC2 instance accessible via AWS Systems Manager Session Manager. SSM eliminates the need for SSH keys, bastion hosts, or open inbound ports -- connections are brokered through the AWS control plane. This is the modern best practice for EC2 instance access.

## When to Use

- Production or staging instances where SSH key management overhead should be avoided
- Environments with strict security requirements that prohibit opening inbound SSH ports
- Any EC2 instance that needs secure, auditable shell access without a bastion host

## Key Configuration Choices

- **SSM connection** (`connectionMethod: SSM`) -- Access via `aws ssm start-session`; no SSH key or inbound port 22 needed
- **IAM instance profile required** (`iamInstanceProfileArn`) -- The instance profile must include `AmazonSSMManagedInstanceCore` policy
- **Termination protection** (`disableApiTermination: true`) -- Prevents accidental instance termination
- **t3.small** (`instanceType`) -- 2 vCPUs, 2 GiB RAM; suitable for most general-purpose workloads
- **30 GiB root volume** (`rootVolumeSizeGb: 30`) -- Default size; increase for data-intensive applications

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<aws-region>` | AWS region where the instance will be created (e.g., `us-west-2`) | AWS region list |
| `<instance-name>` | Name tag for the EC2 instance (e.g., `web-server-01`) | Your naming convention |
| `<ami-id>` | Amazon Machine Image ID (e.g., `ami-0abcdef1234567890` for Amazon Linux 2023) | AWS EC2 AMI catalog or `aws ec2 describe-images` |
| `<private-subnet-id>` | Private subnet ID where the instance will launch | AWS VPC console or `AwsVpc` status outputs |
| `<security-group-id>` | Security group ID controlling instance traffic | AWS EC2 console or `AwsSecurityGroup` status outputs |
| `<ssm-instance-profile-arn>` | ARN of IAM instance profile with SSM permissions | AWS IAM console or `AwsIamRole` status outputs |

## Related Presets

- **02-ssh-accessible** -- Use instead when SSH key-based access is required (e.g., for legacy tooling or bastion workflows)

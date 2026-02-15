---
title: "Presets"
description: "Ready-to-deploy configuration presets for EC2 Instance"
type: "preset-list"
componentSlug: "ec2-instance"
componentTitle: "EC2 Instance"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-ssm-managed"
    rank: "01"
    title: "SSM-Managed Instance"
    excerpt: "This preset creates an EC2 instance accessible via AWS Systems Manager Session Manager. SSM eliminates the need for SSH keys, bastion hosts, or open inbound ports -- connections are brokered through..."
  - slug: "02-ssh-accessible"
    rank: "02"
    title: "SSH-Accessible Instance"
    excerpt: "This preset creates an EC2 instance accessible via traditional SSH through a bastion host or direct connection. It requires an EC2 key pair for authentication. Use this when your tooling or workflows..."
---

# EC2 Instance Presets

Ready-to-deploy configuration presets for EC2 Instance. Each preset is a complete manifest you can copy, customize, and deploy.

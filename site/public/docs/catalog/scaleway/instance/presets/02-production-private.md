---
title: "Production Private Instance"
description: "This preset creates a production-grade Scaleway Instance on a Private Network with no public IP, an explicit security group, and deletion protection enabled. The instance is only reachable through..."
type: "preset"
rank: "02"
presetSlug: "02-production-private"
componentSlug: "instance"
componentTitle: "Instance"
provider: "scaleway"
icon: "package"
order: 2
---

# Production Private Instance

This preset creates a production-grade Scaleway Instance on a Private Network with no public IP, an explicit security group, and deletion protection enabled. The instance is only reachable through the Private Network -- via a Public Gateway bastion, Load Balancer, or VPN.

## When to Use

- Application servers behind a Load Balancer
- Database hosts, worker nodes, or backend services that should not be directly internet-accessible
- Any production workload requiring a hardened network posture

## Key Configuration Choices

- **PRO2-S instance** (`type: PRO2-S`) -- 2 vCPU, 8 GB RAM; production-optimized with guaranteed resources and higher baseline performance than DEV1
- **Ubuntu 22.04** (`image: ubuntu_jammy`) -- widely supported LTS release; change to your preferred OS image
- **No public IP** -- the instance is reachable only via the Private Network, reducing attack surface
- **Explicit security group** (`securityGroupId`) -- enforces specific inbound/outbound rules instead of the default allow-all
- **Private Network attached** (`privateNetworkId`) -- the instance receives a private NIC for internal communication
- **Deletion protection** (`protected: true`) -- prevents accidental termination via API or Terraform/Pulumi destroy

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<your-security-group-id>` | UUID of the security group to attach | Scaleway console or `ScalewayInstanceSecurityGroup` status outputs |
| `<your-private-network-id>` | UUID of the Private Network to attach to | Scaleway console or `ScalewayPrivateNetwork` status outputs |

## Related Presets

- **01-dev-instance** -- Use instead for quick development instances with a public IP and no security group

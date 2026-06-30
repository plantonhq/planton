# Allow SSH Ingress Rule

This preset creates a standalone security group rule that allows inbound SSH (TCP port 22) from a trusted CIDR. Use standalone rules (instead of inline rules in `OpenStackSecurityGroup`) when individual rules need to be independently managed or visible as separate DAG nodes in InfraCharts.

## When to Use

- Adding SSH access to an existing security group
- InfraCharts where rule creation depends on another resource (e.g., a security group created in a separate manifest)
- Managing rules independently from the security group lifecycle

## Key Configuration Choices

- **Ingress only** (`direction: ingress`) -- allows inbound SSH connections
- **Restricted source** (`remoteIpPrefix: <trusted-cidr>`) -- SSH is not open to the world
- **Single port** (`portRangeMin: 22`, `portRangeMax: 22`) -- only SSH, nothing else
- **ForceNew** -- all fields are immutable; changing any field recreates the rule

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<security-group-id>` | ID of the security group to add this rule to | OpenStack console or `OpenStackSecurityGroup` status outputs |
| `<trusted-cidr>` | CIDR allowed to SSH (e.g., `10.0.0.0/8` or `203.0.113.50/32`) | Your network admin or VPN configuration |

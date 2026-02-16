# Hub-and-Spoke Firewall

Transit Gateway with a centralized inspection VPC running a virtual firewall appliance (e.g., Palo Alto, Fortinet, AWS Network Firewall). The inspection VPC uses appliance mode to ensure symmetric routing for stateful packet inspection.

## When to Use

- Compliance requirements mandate centralized traffic inspection
- Running a virtual firewall appliance in a shared-services VPC
- Hub-and-spoke topology where all inter-VPC traffic passes through inspection
- Security-sensitive workloads (PCI-DSS, HIPAA, SOC2)

## Key Configuration Choices

- **applianceModeSupport: true** on the inspection VPC -- ensures return traffic routes through the same AZ as the original flow, maintaining symmetric routing for stateful firewalls
- **Multi-AZ subnets** on both VPCs -- high availability for inspection path
- **Full-mesh default routing** -- can be customized later with explicit route tables for isolation

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<inspection-vpc-id>` | Inspection/firewall VPC ID | AwsVpc status.outputs.vpc_id |
| `<inspection-subnet-az1>` | Inspection VPC subnet in AZ1 | AwsVpc status.outputs.private_subnets[0].id |
| `<inspection-subnet-az2>` | Inspection VPC subnet in AZ2 | AwsVpc status.outputs.private_subnets[1].id |
| `<workload-vpc-id>` | Workload VPC ID | AwsVpc status.outputs.vpc_id |
| `<workload-subnet-az1>` | Workload VPC subnet in AZ1 | AwsVpc status.outputs.private_subnets[0].id |
| `<workload-subnet-az2>` | Workload VPC subnet in AZ2 | AwsVpc status.outputs.private_subnets[1].id |

## Related Presets

- **01-multi-vpc-hub** -- simpler full-mesh without inspection

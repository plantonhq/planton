---
title: "VCN Flow Logs"
description: "This preset creates a log group with a service log that automatically collects VCN flow log data from a subnet. Flow logs capture metadata about every network packet accepted or rejected by security..."
type: "preset"
rank: "01"
presetSlug: "01-vcn-flow-logs"
componentSlug: "log-group"
componentTitle: "Log Group"
provider: "oci"
icon: "package"
order: 1
---

# VCN Flow Logs

This preset creates a log group with a service log that automatically collects VCN flow log data from a subnet. Flow logs capture metadata about every network packet accepted or rejected by security lists and network security groups, providing visibility into traffic patterns, security auditing, and compliance evidence. The 90-day retention supports most regulatory audit windows.

## When to Use

- Network security auditing and forensics -- understanding what traffic was allowed or denied
- Compliance requirements mandating network traffic logging (PCI-DSS, SOC 2, HIPAA)
- Troubleshooting connectivity issues by examining accepted and rejected flows
- Establishing network traffic baselines before tightening security group rules

## Key Configuration Choices

- **Service log type** (`logType: service`) -- OCI automatically collects and delivers flow log records without any agent installation. The Logging service pulls data directly from the VCN infrastructure.
- **Flow logs service** (`service: flowlogs`) -- targets the VCN flow log data source. This captures TCP, UDP, and ICMP metadata (source/destination IP, port, protocol, action, bytes, packets) for all traffic hitting security rules on the specified subnet.
- **All categories** (`category: all`) -- captures both accepted and rejected traffic. Use `accept` or `reject` individually if you only need one direction, but `all` provides the most complete audit trail.
- **90-day retention** (`retentionDuration: 90`) -- retains logs for 3 months, covering standard compliance audit windows. OCI supports 30-day increments up to 180 days. Increase to 180 for regulated industries; decrease to 30 for cost-sensitive development environments.
- **Enabled on creation** (`isEnabled: true`) -- the log starts collecting immediately. Set to `false` to create the log configuration without activating collection (useful for staged rollouts).

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the log group will be created | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |
| `<vcn-subnet-ocid>` | OCID of the subnet to collect flow logs from | OCI Console > Networking > Subnets, or `OciSubnet` status outputs (`subnetId`) |

## Related Presets

- **02-custom-application-logs** -- use instead for application-level telemetry pushed via the Logging Ingestion API

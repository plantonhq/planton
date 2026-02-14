# Static IP NAT for Specific Subnets

This preset creates a Cloud Router with NAT restricted to specific subnets using manually assigned static external IPs. Use this when you need predictable egress IP addresses -- for example, when partners require allowlisting your NAT IPs or compliance mandates stable source addresses.

## When to Use

- Egress traffic must come from known, static IP addresses (partner allowlisting, compliance)
- Only certain subnets should have NAT-based internet access
- Security policies require explicit control over which subnets can reach the internet

## Key Configuration Choices

- **Specific subnets** (`subnetworkSelfLinks`) -- only the listed subnets get NAT; others in the region are excluded
- **Static IPs** (`natIpNames`) -- manually provisioned external IPs used for NAT (predictable and allowlist-friendly)
- **Error-only logging** (`logFilter: ERRORS_ONLY`) -- monitors for port exhaustion without high log volume

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID | GCP Console or `GcpProject` outputs |
| `<vpc-network-self-link>` | Self-link of the VPC network | `GcpVpc` status outputs |
| `<gcp-region>` | GCP region (e.g., `us-central1`) | Your deployment region |
| `<your-router-name>` | Name for the Cloud Router (1-63 chars, lowercase) | Choose a descriptive name |
| `<your-nat-name>` | Name for the NAT configuration (1-63 chars, lowercase) | Choose a descriptive name |
| `<subnet-self-link>` | Self-link of the subnet to enable NAT on | `GcpSubnetwork` status outputs |
| `<static-ip-name>` | Name of a pre-provisioned static external IP address | GCP Compute Engine console or `gcloud compute addresses list` |

## Related Presets

- **01-all-subnets-auto** -- Use for simpler setups where all subnets share auto-allocated NAT IPs

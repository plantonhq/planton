# Ingress Allow SSH via IAP

This preset creates a firewall rule that allows SSH (port 22) access exclusively from Google's Identity-Aware Proxy (IAP) IP range (`35.235.240.0/20`). IAP provides authenticated, audited SSH access to VMs without requiring them to have external IP addresses or exposing port 22 to the public internet.

## When to Use

- Any environment where VMs need SSH access for administration or debugging
- Private VMs without external IPs (IAP TCP forwarding tunnels traffic through Google's edge)
- Environments that require audited SSH access (IAP logs every connection attempt in Cloud Audit Logs)
- Replacing legacy bastion host / jump box patterns with a Google-managed alternative

## Key Configuration Choices

- **Source range `35.235.240.0/20`** -- this is the only CIDR block used by Google IAP for TCP forwarding; no other source should be needed for SSH access
- **Target tag `allow-ssh`** -- restricts the rule to VMs explicitly tagged with `allow-ssh`, preventing network-wide SSH exposure; remove `targetTags` if all VMs in the network should be reachable
- **Priority 1000** -- standard application rule priority
- **No `0.0.0.0/0`** -- SSH is never exposed to the full internet; only Google's IAP infrastructure can initiate connections

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID where the firewall rule will be created | GCP Console or `GcpProject` outputs |
| `<vpc-network-name-or-self-link>` | VPC network name (e.g., `default`) or full self-link | `GcpVpc` status outputs |
| `<your-rule-name>` | Unique name for this firewall rule (1-63 chars, lowercase, hyphens) | Choose a descriptive name (e.g., `allow-ssh-from-iap`) |

## Related Presets

- **01-ingress-allow-web** -- Allow HTTP/HTTPS traffic from the internet for web servers
- **03-egress-deny-all** -- Deny all outbound traffic as a restrictive baseline

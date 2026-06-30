# Certificate-Based Split-Tunnel VPN

This preset creates an AWS Client VPN endpoint using mutual TLS certificate authentication with split-tunnel routing. Only traffic destined for the VPC flows through the VPN -- all other internet traffic stays on the client's local network. This is the most common VPN configuration for developers accessing private VPC resources from their laptops.

## When to Use

- Developers accessing private resources (databases, internal APIs) in a VPC from remote locations
- Split-tunnel is preferred when VPN is only needed for internal resources, not for general internet browsing
- Certificate-based authentication is the simplest Client VPN auth method (no Active Directory or Cognito required)

## Key Configuration Choices

- **Certificate authentication** (`authenticationType: certificate`) -- Mutual TLS; clients present a certificate signed by a trusted CA
- **Split-tunnel** (`disableSplitTunnel: false`) -- Only VPC-bound traffic routes through the VPN; internet traffic stays local for better performance
- **TCP on port 443** (`transportProtocol: tcp`, `vpnPort: 443`) -- Traverses corporate firewalls that block non-standard ports
- **VPC CIDR authorization** -- Clients can access the entire VPC; restrict to specific subnet CIDRs for tighter control
- **Client CIDR /22** (`clientCidrBlock: 10.100.0.0/22`) -- Supports up to ~1,000 concurrent VPN connections

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<vpc-id>` | VPC ID to attach the VPN endpoint to | AWS VPC console or `AwsVpc` status outputs |
| `<private-subnet-id>` | Subnet ID for VPN target network association | AWS VPC console or `AwsVpc` status outputs |
| `<server-certificate-arn>` | ACM certificate ARN for the VPN server (mutual TLS) | AWS ACM console or `AwsCertManagerCert` status outputs |
| `<vpc-cidr-block>` | VPC CIDR block to authorize for VPN clients (e.g., `10.0.0.0/16`) | Your VPC configuration |

## Related Presets

- **02-certificate-full-tunnel** -- Use instead when all client traffic must route through the VPN (security-sensitive environments)

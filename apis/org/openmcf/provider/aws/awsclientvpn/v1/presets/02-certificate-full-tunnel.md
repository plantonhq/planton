# Certificate-Based Full-Tunnel VPN

This preset creates an AWS Client VPN endpoint with full-tunnel routing, where all client traffic -- including internet traffic -- routes through the VPN. This provides complete network control and visibility over connected clients, suitable for security-sensitive environments that require traffic inspection or compliance logging.

## When to Use

- Security-sensitive environments requiring all traffic to pass through corporate network controls
- Compliance scenarios where internet traffic must be logged, filtered, or inspected
- Environments where clients should appear to originate from the VPC's IP addresses

## Key Configuration Choices

- **Full-tunnel** (`disableSplitTunnel: true`) -- All client traffic routes through the VPN, including internet-bound traffic
- **Authorize all traffic** (`cidrAuthorizationRules: [0.0.0.0/0]`) -- Clients can reach both VPC resources and the internet through the VPN
- **Certificate authentication** (`authenticationType: certificate`) -- Mutual TLS; same as split-tunnel preset
- **TCP on port 443** -- Firewall-friendly transport

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<vpc-id>` | VPC ID to attach the VPN endpoint to | AWS VPC console or `AwsVpc` status outputs |
| `<private-subnet-id>` | Subnet ID for VPN target network association | AWS VPC console or `AwsVpc` status outputs |
| `<server-certificate-arn>` | ACM certificate ARN for the VPN server | AWS ACM console or `AwsCertManagerCert` status outputs |

## Related Presets

- **01-certificate-split-tunnel** -- Use instead when only VPC traffic should route through the VPN (better performance for general internet use)

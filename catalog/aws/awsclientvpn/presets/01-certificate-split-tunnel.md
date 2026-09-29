# Certificate Split-Tunnel VPN

This preset creates a Client VPN endpoint with mutual-TLS authentication
and split-tunnel routing — the standard corp-access posture: only traffic
destined for the VPC goes through the VPN, everything else stays local to
the client.

## When to Use

- Developer and operator access to private VPC resources (databases,
  internal services) from laptops
- Teams without an external identity provider — certificate distribution
  is the only infrastructure needed

## Key Configuration Choices

- **Certificate authentication** — clients present certificates issued
  from the client CA chain (`rootCertificateChainArn`); for a self-signed
  setup the same imported ACM certificate may serve as both server
  certificate and client CA
- **Split tunnel** (`splitTunnel: true`) — only VPC-bound traffic routes
  through the VPN; internet traffic stays local for better performance
- **One authorization rule** — grants every authenticated client the VPC
  CIDR; nothing is reachable without a rule
- **Connection logging** — connect/disconnect events stream to CloudWatch
  Logs (strongly recommended; without it there is no record of who
  connected)

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<aws-region>` | AWS region code (e.g. `us-west-2`) | Your deployment region |
| `<client-ca-certificate-arn>` | ACM ARN of the client CA chain | ACM console, or an AwsCertManagerCert output |
| `<server-certificate-arn>` | ACM ARN of the server certificate | ACM console, or an AwsCertManagerCert output |
| `<vpc-id>` | The VPC whose security groups apply | Your AwsVpc output |
| `<private-subnet-id>` | Subnet to associate (bills hourly; add a second AZ for resilience) | Your AwsSubnet output |
| `10.0.0.0/16` | The VPC CIDR clients may reach (`authorizationRules[].targetNetworkCidr`) | Your VPC design |
| `<vpn-log-group-name>` | CloudWatch log group for connection events | Your AwsCloudwatchLogGroup output |

## Common Additions

- `dnsServers` pointing at the VPC resolver (VPC CIDR base + 2) so clients
  resolve private hosted zones
- A second `subnetIds` entry in another AZ
- `clientLoginBanner` for legal/acceptable-use notices

## Related Presets

- **02-certificate-full-tunnel** — route ALL client traffic through AWS
- **03-saml-sso** — SAML single sign-on with the self-service portal

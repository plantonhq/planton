# SAML SSO VPN with Self-Service Portal

This preset creates a Client VPN endpoint with SAML 2.0 single sign-on and
group-scoped authorization — users authenticate with their existing
identity provider (Okta, Entra ID, ...), download their own client
configuration from the self-service portal, and reach only the networks
their IdP group allows.

## When to Use

- Organizations with an identity provider — no certificate distribution,
  instant offboarding through the IdP
- Different access levels per team: authorization rules scope destination
  CIDRs to IdP groups

## Key Configuration Choices

- **Federated authentication** — the IAM SAML provider brokers sign-on;
  the only type that supports the self-service portal
- **Group-scoped authorization rules** (`accessGroupId`) — each rule
  grants one IdP group one destination CIDR; a client's reachable networks
  are exactly the union of its groups' rules
- **Session hygiene** — 12-hour sessions with hard disconnect at timeout
  (`disconnectOnSessionTimeout: true`) force daily re-authentication

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<aws-region>` | AWS region code (e.g. `us-west-2`) | Your deployment region |
| `arn:aws:iam::123456789012:saml-provider/okta-workforce` | The IAM SAML identity provider ARN | IAM console → Identity providers |
| `<server-certificate-arn>` | ACM ARN of the server certificate | ACM console, or an AwsCertManagerCert output |
| `<vpc-id>` | The VPC whose security groups apply | Your AwsVpc output |
| `<private-subnet-id-az1>` / `<private-subnet-id-az2>` | Subnets to associate (two AZs for resilience) | Your AwsSubnet outputs |
| `10.0.0.0/16` / `10.0.42.0/24` | Destination networks per group (the VPC, the tools subnet) | Your VPC design |
| `<idp-engineering-group-id>` / `<idp-contractors-group-id>` | The IdP group attribute values | Your identity provider |
| `<vpn-log-group-name>` | CloudWatch log group for connection events | Your AwsCloudwatchLogGroup output |

## Common Additions

- `selfServiceSamlProviderArn` on the authentication option when the
  portal needs a different SAML app than the VPN itself
- A second `authenticationOptions` entry (certificate) for break-glass
  machine access — a client passes on ANY option

## Related Presets

- **01-certificate-split-tunnel** — mutual TLS without an IdP
- **02-certificate-full-tunnel** — route ALL client traffic through AWS

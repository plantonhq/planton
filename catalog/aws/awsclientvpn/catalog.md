# AWS Client VPN

Deploys a Client VPN endpoint on AWS — the managed OpenVPN server remote users and machines connect to for secure access into AWS networks. Authentication supports certificate-based mutual TLS, Active Directory, and SAML federation (one or two options combined); the endpoint attaches to a VPC through per-subnet associations or directly to a Transit Gateway for hub-wide reach. Two things to internalize before deploying: access is deny-by-default (connected clients reach nothing until an authorization rule grants it), and authentication, the client CIDR, and the transport protocol are all fixed at create time -- changing any of them replaces the endpoint.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Client VPN Endpoint** -- an OpenVPN-compatible endpoint with configurable transport (UDP/TCP on 443 or 1194), authentication options, split- or full-tunnel routing, session controls, DNS push, and optional CloudWatch connection logging
- **Target Network Associations** -- one per subnet in `subnetIds` (VPC attachment), each activating the endpoint in that subnet's Availability Zone
- **Transit Gateway Attachment** -- when `transitGatewayConfiguration` is set instead, one attachment giving clients reach into every network the gateway routes to
- **Authorization Rules** -- one per entry in `authorizationRules`; the endpoint's network ACL — clients reach nothing until a rule grants it
- **Routes** -- one per entry in `routes`, beyond the auto-created route for each associated subnet's VPC
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **An ACM server certificate** in the same region for the endpoint's TLS identity — required for every authentication type. For a self-signed certificate-auth setup, the same imported certificate may also serve as the client CA chain.
- **An identity source per authentication option** -- an imported client CA chain in ACM (certificate auth), an AWS Directory Service directory (Active Directory), or an IAM SAML provider (federated).
- **A network to land in** -- a VPC with subnets (associate two AZs for resilience; each association bills hourly), or a Transit Gateway for hub-wide access.
- **A client CIDR block** (e.g. `10.100.0.0/22`, size /22 to /12) that does not overlap the VPC CIDR or any added route. Fixed at create time. Not used with IPv6 tunnel traffic, where AWS assigns client addressing.

## Deploy

### Console

Open the deployment store, find **AWS Client VPN**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Certificate Split-Tunnel VPN** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsClientVpn
metadata:
  name: dev-vpn
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  authenticationOptions:
    - type: certificate-authentication
      rootCertificateChainArn:
        value: "arn:aws:acm:us-west-2:123456789012:certificate/abc123"
  serverCertificateArn:
    value: "arn:aws:acm:us-west-2:123456789012:certificate/abc123"
  clientCidrBlock: "10.100.0.0/22"
  splitTunnel: true
  vpcId:
    value: "vpc-0a1b2c3d4e5f00001"
  subnetIds:
    - value: "subnet-0a1b2c3d4e5f00001"
    - value: "subnet-0a1b2c3d4e5f00002"
  authorizationRules:
    - targetNetworkCidr: "10.0.0.0/16"
      authorizeAllGroups: true
```

```shell
planton apply -f client-vpn.yaml
```

This creates a certificate-authenticated endpoint with split-tunnel routing, the AWS-default UDP transport on port 443, two AZ associations, and authorization to reach the VPC CIDR. Connection logging is not configured — add `connectionLog` in production. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the VPN to resources deployed in the same InfraPipeline:

```yaml
spec:
  vpcId:
    valueFrom:
      kind: AwsVpc
      name: production-vpc
      fieldPath: status.outputs.vpc_id
  subnetIds:
    - valueFrom:
        kind: AwsSubnet
        name: private-usw2a
        fieldPath: status.outputs.subnet_id
  serverCertificateArn:
    valueFrom:
      kind: AwsCertManagerCert
      name: vpn-server-cert
      fieldPath: status.outputs.cert_arn
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: vpn-sg
        fieldPath: status.outputs.security_group_id
  connectionLog:
    cloudwatchLogGroup:
      valueFrom:
        kind: AwsCloudwatchLogGroup
        name: vpn-logs
        fieldPath: status.outputs.log_group_name
```

The InfraPipeline resolves the dependency graph, deploys the referenced resources first, then provisions the Client VPN with the resolved values.

## Key Configuration

These are the most important decisions when configuring a Client VPN. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Authentication (create-time immutable)** -- One or two `authenticationOptions`; a client passes if it satisfies ANY one. The classic pairing is certificates for managed devices plus federated SSO for humans. Changing authentication later replaces the endpoint. Only federated authentication unlocks the self-service portal.

**VPC vs Transit Gateway attachment** -- Mutually exclusive. VPC attachment gives per-subnet associations gated by security groups; Transit Gateway attachment gives clients everything the hub routes to, with access governed entirely by authorization rules. The gateway configuration is fixed at create time.

**Split-tunnel vs full-tunnel** -- `splitTunnel: true` (the usual corp-access posture) routes only endpoint-route-table traffic through the VPN. Full tunnel (AWS's raw default) routes ALL client traffic — pair it with a `0.0.0.0/0` route AND a `0.0.0.0/0` authorization rule through a NAT-ed subnet, or clients lose internet access.

**Transport and port** -- UDP (the AWS default, lower latency) or TCP (for networks that block UDP), on port 443 or 1194. Either port works with either protocol. The protocol is fixed at create time; the port updates in place.

**Authorization rules** -- Nothing is reachable until a rule authorizes it, even with associations in place. Routes move packets; rules permit them.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsCertManagerCert** | `serverCertificateArn`, `authenticationOptions[].rootCertificateChainArn` | `status.outputs.cert_arn` |
| **AwsVpc** (VPC mode) | `vpcId` | `status.outputs.vpc_id` |
| **AwsSubnet** (VPC mode) | `subnetIds[]`, `routes[].targetSubnetId` | `status.outputs.subnet_id` |
| **AwsSecurityGroup** (VPC mode) | `securityGroupIds[]` | `status.outputs.security_group_id` |
| **AwsTransitGateway** (TGW mode) | `transitGatewayConfiguration.transitGatewayId` | `status.outputs.transit_gateway_id` |
| **AwsCloudwatchLogGroup** | `connectionLog.cloudwatchLogGroup` | `status.outputs.log_group_name` |
| **AwsLambda** | `clientConnectOptions.lambdaFunctionArn` | `status.outputs.function_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `client_vpn_endpoint_id` | AWS-assigned Client VPN endpoint ID | AWS CLI operations, exporting the client configuration file |
| `client_vpn_endpoint_arn` | ARN of the endpoint | IAM policies, cross-service permissions |
| `endpoint_dns_name` | DNS name clients connect to | VPN client configuration (the exported config prepends the required random subdomain automatically) |
| `self_service_portal_url` | Portal where federated users download their own configuration | User onboarding; empty when the portal is disabled |
| `subnet_association_ids` | Map of subnet ID → association ID | Managing specific target network associations |
| `transit_gateway_attachment_id` | The gateway attachment (TGW mode only) | Transit Gateway route table associations and routes |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Certificate split-tunnel** -- Mutual TLS with split-tunnel routing: only VPC traffic flows through the VPN. The most common configuration for developer access to private resources. Start from the **Certificate Split-Tunnel VPN** preset.

**Certificate full-tunnel** -- All client traffic routes through the VPN for complete control and inspection — compliance-sensitive environments. Start from the **Certificate Full-Tunnel VPN** preset.

**SAML SSO with self-service portal** -- Federated authentication through an IAM SAML provider (Okta, Entra ID), with the self-service portal enabled so users download their own client configuration. Start from the **SAML SSO VPN with Self-Service Portal** preset.

**Hub access via Transit Gateway** -- Attach the endpoint to a Transit Gateway so remote users reach every VPC, on-prem range, and peered network the hub routes to, with one attachment instead of per-VPC associations.

## Works With

- [**AWS ACM Certificate**](/infra-catalog/aws-cert-manager-cert) -- the endpoint's TLS identity and (for certificate auth) the client CA chain
- [**AWS VPC**](/infra-catalog/aws-vpc) -- the VPC and subnets for per-subnet attachment
- [**AWS Security Group**](/infra-catalog/aws-security-group) -- controls traffic between VPN clients and VPC resources
- [**AWS Transit Gateway**](/infra-catalog/aws-transit-gateway) -- hub-wide access without per-subnet associations
- [**AWS CloudWatch Log Group**](/infra-catalog/aws-cloudwatch-log-group) -- connection logging destination
- [**AWS Lambda**](/infra-catalog/aws-lambda) -- the per-connection allow/deny posture hook

# AwsClientVpn

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsClientVpnSpec defines the desired configuration for an AWS Client VPN
endpoint — the managed OpenVPN server remote users and machines connect to
for secure access into AWS networks.

The endpoint itself is just the front door. Access flows through three
endpoint-scoped satellites, all folded here because none has identity
outside its endpoint:

- `subnet_ids` (target network associations): each associated subnet
  attaches the endpoint to a VPC and gives clients a path into that
  subnet's Availability Zone. Associations are what you pay for hourly and
  what take minutes to attach/detach; an endpoint with zero associations
  is valid (a pre-staged front door) but routes no traffic.
- `authorization_rules`: network-ACL-style grants deciding WHICH clients
  (everyone, or one identity-provider group) may reach WHICH destination
  CIDR. No rule, no traffic — even with an association in place.
- `routes`: entries in the endpoint's route table beyond the
  auto-created route for each associated subnet's VPC — e.g. 0.0.0.0/0
  through a NAT-ed subnet for full-tunnel internet egress, or an on-prem
  CIDR through a subnet that reaches a transit gateway.

Authentication is decided at create time (all authentication options are
create-time immutable): mutual-TLS client certificates,
Active Directory user/password, or SAML federation (which also unlocks
the self-service portal where users download their own client
configuration). Up to two options may be combined (e.g. certificate +
federated) — a client passes if it satisfies any one of them.

Create-time-immutable (ForceNew) fields — changing any of these replaces
the endpoint: `authentication_options`, `client_cidr_block`,
`transport_protocol`, `endpoint_ip_address_type`,
`traffic_ip_address_type`, and `transit_gateway_configuration`.
Everything else updates in place.

Credentials, region, and deployment workflow live outside this spec in
IaC inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsClientVpn
metadata:
  name: corp-access
  id: test-clientvpn
  org: test-org
  env: dev
spec:
  region: us-west-2
  description: Full-surface Client VPN hack manifest
  authenticationOptions:
    - type: certificate-authentication
      rootCertificateChainArn:
        value: arn:aws:acm:us-west-2:123456789012:certificate/client-ca
  serverCertificateArn:
    value: arn:aws:acm:us-west-2:123456789012:certificate/server-cert
  clientCidrBlock: 10.100.0.0/22
  splitTunnel: true
  transportProtocol: udp
  vpnPort: 443
  vpcId:
    value: vpc-0123456789abcdef0
  securityGroupIds:
    - value: sg-0a1b2c3d4e5f00001
  subnetIds:
    - value: subnet-aaa111
    - value: subnet-bbb222
  authorizationRules:
    - targetNetworkCidr: 10.0.0.0/16
      authorizeAllGroups: true
      description: reach the whole VPC
  routes:
    - destinationCidrBlock: 192.168.0.0/16
      targetSubnetId:
        value: subnet-aaa111
      description: on-prem network via VPN-attached subnet
  sessionTimeoutHours: 12
  disconnectOnSessionTimeout: true
  clientRouteEnforcementEnabled: true
  clientLoginBanner:
    bannerText: Authorized use only.
  connectionLog:
    cloudwatchLogGroup:
      value: /vpn/corp-access
  dnsServers:
    - 10.0.0.2
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.authenticationOptions` | `[]AwsClientVpnAuthenticationOption` | yes |  |  |
| `spec.authenticationOptions[].type` | `string` | yes |  |  |
| `spec.authenticationOptions[].rootCertificateChainArn` | `string \| valueFrom` |  |  | AwsCertManagerCert (`status.outputs.cert_arn`) |
| `spec.authenticationOptions[].activeDirectoryId` | `string` |  |  |  |
| `spec.authenticationOptions[].samlProviderArn` | `string` |  |  |  |
| `spec.authenticationOptions[].selfServiceSamlProviderArn` | `string` |  |  |  |
| `spec.serverCertificateArn` | `string \| valueFrom` | yes |  | AwsCertManagerCert (`status.outputs.cert_arn`) |
| `spec.clientCidrBlock` | `string` |  |  |  |
| `spec.splitTunnel` | `bool` |  |  |  |
| `spec.transportProtocol` | `string` |  | `udp` |  |
| `spec.vpnPort` | `int32` |  | `443` |  |
| `spec.endpointIpAddressType` | `string` |  |  |  |
| `spec.trafficIpAddressType` | `string` |  |  |  |
| `spec.vpcId` | `string \| valueFrom` |  |  | AwsVpc (`status.outputs.vpc_id`) |
| `spec.securityGroupIds` | `[]string \| valueFrom` |  |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.subnetIds` | `[]string \| valueFrom` |  |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.transitGatewayConfiguration` | `AwsClientVpnTransitGatewayConfiguration` |  |  |  |
| `spec.transitGatewayConfiguration.transitGatewayId` | `string \| valueFrom` | yes |  | AwsTransitGateway (`status.outputs.transit_gateway_id`) |
| `spec.transitGatewayConfiguration.availabilityZones` | `[]string` |  |  |  |
| `spec.transitGatewayConfiguration.availabilityZoneIds` | `[]string` |  |  |  |
| `spec.authorizationRules` | `[]AwsClientVpnAuthorizationRule` |  |  |  |
| `spec.authorizationRules[].targetNetworkCidr` | `string` | yes |  |  |
| `spec.authorizationRules[].accessGroupId` | `string` |  |  |  |
| `spec.authorizationRules[].authorizeAllGroups` | `bool` |  |  |  |
| `spec.authorizationRules[].description` | `string` |  |  |  |
| `spec.routes` | `[]AwsClientVpnRoute` |  |  |  |
| `spec.routes[].destinationCidrBlock` | `string` | yes |  |  |
| `spec.routes[].targetSubnetId` | `string \| valueFrom` | yes |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.routes[].description` | `string` |  |  |  |
| `spec.sessionTimeoutHours` | `int32` |  | `24` |  |
| `spec.disconnectOnSessionTimeout` | `bool` |  |  |  |
| `spec.selfServicePortalEnabled` | `bool` |  |  |  |
| `spec.clientConnectOptions` | `AwsClientVpnClientConnectOptions` |  |  |  |
| `spec.clientConnectOptions.lambdaFunctionArn` | `string \| valueFrom` | yes |  | AwsLambda (`status.outputs.function_arn`) |
| `spec.clientLoginBanner` | `AwsClientVpnLoginBanner` |  |  |  |
| `spec.clientLoginBanner.bannerText` | `string` | yes |  |  |
| `spec.clientRouteEnforcementEnabled` | `bool` |  |  |  |
| `spec.dnsServers` | `[]string` |  |  |  |
| `spec.connectionLog` | `AwsClientVpnConnectionLog` |  |  |  |
| `spec.connectionLog.cloudwatchLogGroup` | `string \| valueFrom` | yes |  | AwsCloudwatchLogGroup (`status.outputs.log_group_name`) |
| `spec.connectionLog.cloudwatchLogStream` | `string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the Client VPN endpoint will be created.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.description

`string`

Human-friendly description shown in the AWS console.

### spec.authenticationOptions

`[]AwsClientVpnAuthenticationOption` · required

How clients prove who they are before a tunnel is established. One or
two options (a client passes if it satisfies ANY one); all ForceNew —
changing authentication replaces the endpoint. See
AwsClientVpnAuthenticationOption for the three types.

- rule: {"required":true,"repeated":{"minItems":"1","maxItems":"2"}}
- rule: authentication type must be 'certificate-authentication', 'directory-service-authentication', or 'federated-authentication'
- rule: certificate-authentication requires root_certificate_chain_arn — the client CA chain certificates are validated against
- rule: directory-service-authentication requires active_directory_id
- rule: federated-authentication requires saml_provider_arn
- rule: root_certificate_chain_arn is only accepted with certificate-authentication
- rule: active_directory_id is only accepted with directory-service-authentication
- rule: saml_provider_arn and self_service_saml_provider_arn are only accepted with federated-authentication

### spec.authenticationOptions[].type

`string` · required

Authentication type. Values:

- "certificate-authentication": mutual TLS — the client presents a
  certificate issued from the chain in `root_certificate_chain_arn`.
  Zero external identity infrastructure; certificate distribution and
  revocation are on you.

- "directory-service-authentication": user/password against an AWS
  Directory Service directory (`active_directory_id`) — Managed
  Microsoft AD or AD Connector to on-prem AD.

- "federated-authentication": SAML 2.0 single sign-on through an IAM
  SAML provider (`saml_provider_arn`) — Okta, Entra ID, and friends;
  the only type that supports the self-service portal.

- rule: {"required":true}

### spec.authenticationOptions[].rootCertificateChainArn

`string | valueFrom`

For "certificate-authentication": the ACM certificate whose CHAIN
client certificates must descend from — the client CA, distinct in
role from the endpoint's server certificate (for a self-signed setup
the same imported certificate may serve both roles, but never assume
it). Reference an imported CA certificate in ACM.

- references: AwsCertManagerCert (`status.outputs.cert_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsCertManagerCert, name: <that resource's name>, fieldPath: status.outputs.cert_arn}} -- a bare string does not parse

### spec.authenticationOptions[].activeDirectoryId

`string`

For "directory-service-authentication": the AWS Directory Service
directory ID (e.g. "d-1234567890"). Directories have no Planton kind —
pass the literal ID.

### spec.authenticationOptions[].samlProviderArn

`string`

For "federated-authentication": the ARN of the IAM SAML identity
provider that brokers sign-on (e.g.
"arn:aws:iam::123456789012:saml-provider/okta"). IAM SAML providers
have no Planton kind — pass the literal ARN (shape-checked here, since
no reference can supply it).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^arn:aws[a-zA-Z-]*:iam::[0-9]{12}:saml-provider/.+$"}}

### spec.authenticationOptions[].selfServiceSamlProviderArn

`string`

For "federated-authentication", optional: a second IAM SAML provider
used only by the self-service portal (when the portal needs a
different SAML app than the VPN itself).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^arn:aws[a-zA-Z-]*:iam::[0-9]{12}:saml-provider/.+$"}}

### spec.serverCertificateArn

`string | valueFrom` · required

The ACM certificate the VPN server presents to connecting clients — the
TLS identity of the endpoint itself, required for every authentication
type. Must live in ACM in the same region. Updates in place (rotating
the server certificate never replaces the endpoint).

- references: AwsCertManagerCert (`status.outputs.cert_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsCertManagerCert, name: <that resource's name>, fieldPath: status.outputs.cert_arn}} -- a bare string does not parse

### spec.clientCidrBlock

`string`

The IPv4 range, in CIDR notation, from which connecting clients are
assigned addresses. Block size between /22 and /12 (AWS's documented
bounds — enforced here so the mistake fails at manifest time, not at
apply); must not overlap the VPC CIDR or any manually added route, and
cannot change after creation. Required — except when
`traffic_ip_address_type` is "ipv6", where AWS derives client
addressing and the field must be empty. Example: "10.100.0.0/22".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^([0-9]{1,3}\\.){3}[0-9]{1,3}/(1[2-9]|2[0-2])$"}}

### spec.splitTunnel

`bool`

Split-tunnel routing. When true, only traffic destined for the
endpoint's route table goes through the VPN and everything else stays
local to the client — the usual posture for corp-access VPNs. When
false (AWS's default, full tunnel), ALL client traffic enters the VPN;
pair that with a 0.0.0.0/0 route + authorization rule through a NAT-ed
subnet or clients lose internet access. Updates in place.

### spec.transportProtocol

`string`

Transport protocol for VPN sessions. Values: "udp" (AWS default —
lower latency, the standard OpenVPN choice), "tcp" (traverses
firewalls that block UDP). ForceNew.

- default: `udp`
- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["udp","tcp"]}}

### spec.vpnPort

`int32` · optional (explicit presence)

Port the endpoint listens on. Values: 443 (default; blends with HTTPS
at the firewall) or 1194 (the traditional OpenVPN port). Either port
works with either transport protocol. Updates in place.

- default: `443`
- rule: {"int32":{"in":[443,1194]}}

### spec.endpointIpAddressType

`string`

IP address type of the ENDPOINT itself — which stacks clients can reach
it over. Values: "ipv4" (default), "ipv6", "dual-stack". ForceNew.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["ipv4","ipv6","dual-stack"]}}

### spec.trafficIpAddressType

`string`

IP address type of the traffic INSIDE the tunnel. Values: "ipv4"
(default), "ipv6", "dual-stack". With "ipv6", `client_cidr_block` must
be empty (AWS assigns client addressing). ForceNew.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["ipv4","ipv6","dual-stack"]}}

### spec.vpcId

`string | valueFrom`

The VPC whose security groups apply to this endpoint. Optional — when
omitted, AWS infers the VPC from the first associated subnet and
applies that VPC's default security group. Mutually exclusive with
`transit_gateway_configuration`. Updates in place.

- references: AwsVpc (`status.outputs.vpc_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsVpc, name: <that resource's name>, fieldPath: status.outputs.vpc_id}} -- a bare string does not parse

### spec.securityGroupIds

`[]string | valueFrom`

Security groups applied to the endpoint's network interfaces (max 5),
governing traffic between VPN clients and VPC resources. When omitted,
the VPC's default security group applies. Mutually exclusive with
`transit_gateway_configuration`. Updates in place.

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: {"repeated":{"maxItems":"5"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.subnetIds

`[]string | valueFrom`

Subnets to associate as target networks (folded
network associations, one per subnet). Each association attaches the
endpoint to that subnet's AZ; associate subnets in two AZs for
resilience (each association bills hourly). All subnets must belong to
one VPC. May be empty — a zero-association endpoint is valid but routes
no traffic until a subnet is associated. Associations add/remove in
place, but each attach/detach takes AWS several minutes. Mutually
exclusive with `transit_gateway_configuration`.

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.transitGatewayConfiguration

`AwsClientVpnTransitGatewayConfiguration`

Associate the endpoint with a transit gateway instead of a VPC — VPN
clients then reach every network the transit gateway routes to, without
per-subnet associations. Mutually exclusive with `vpc_id`,
`security_group_ids`, and `subnet_ids`. ForceNew.

- rule: availability_zones (names) and availability_zone_ids (IDs) are mutually exclusive — pick one addressing form

### spec.transitGatewayConfiguration.transitGatewayId

`string | valueFrom` · required

The transit gateway to attach to.

- references: AwsTransitGateway (`status.outputs.transit_gateway_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsTransitGateway, name: <that resource's name>, fieldPath: status.outputs.transit_gateway_id}} -- a bare string does not parse

### spec.transitGatewayConfiguration.availabilityZones

`[]string`

Availability Zone NAMES the attachment spans (e.g. "us-west-2a").
Mutually exclusive with availability_zone_ids. Leave both empty for
AWS's default AZ selection.

- rule: {"repeated":{"unique":true}}

### spec.transitGatewayConfiguration.availabilityZoneIds

`[]string`

Availability Zone IDs the attachment spans (e.g. "usw2-az1") — the
account-independent form. Mutually exclusive with availability_zones.

- rule: {"repeated":{"unique":true}}

### spec.authorizationRules

`[]AwsClientVpnAuthorizationRule`

Authorization rules — which clients may reach which destination CIDRs.
Without at least one rule, connected clients can reach NOTHING (the
endpoint authorizes no traffic by default). Rules add/remove in place.

- rule: set exactly one of access_group_id (grant one IdP group) or authorize_all_groups (grant every authenticated client)

### spec.authorizationRules[].targetNetworkCidr

`string` · required

The destination network being authorized, in CIDR notation — a VPC
CIDR, one subnet, an on-prem range, or "0.0.0.0/0" for internet access
(full-tunnel setups). Must be the canonical network address (host bits
zero: "10.0.1.0/24", never "10.0.1.5/24") — the provider rejects
non-canonical forms at plan time.

- rule: target_network_cidr must be a canonical network address — host bits must be zero (e.g. "10.0.1.0/24", not "10.0.1.5/24")
- rule: {"required":true,"string":{"pattern":"^([0-9]{1,3}\\.){3}[0-9]{1,3}/([0-9]|[1-2][0-9]|3[0-2])$"}}

### spec.authorizationRules[].accessGroupId

`string`

Grant access only to one identity-provider group: an Active Directory
group SID (directory authentication) or a SAML group attribute value
(federated authentication). Exactly one of this and
`authorize_all_groups` must be set. Certificate-only endpoints have no
group concept — use `authorize_all_groups`. Must not contain a comma —
the provider uses the comma as its internal rule-ID separator, and a
comma here corrupts the rule's identity.

- rule: {"string":{"notContains":","}}

### spec.authorizationRules[].authorizeAllGroups

`bool`

Grant access to every authenticated client. Exactly one of this and
`access_group_id` must be set.

### spec.authorizationRules[].description

`string`

Description shown in the AWS console.

### spec.routes

`[]AwsClientVpnRoute`

Additional routes in the endpoint's route table, beyond the route AWS
auto-creates for each associated subnet's VPC CIDR. The classic uses:
"0.0.0.0/0" through a NAT-ed subnet for full-tunnel internet egress, or
an on-premises CIDR through a subnet that reaches a VPN/transit
gateway. Each route's target subnet must be one of the associated
`subnet_ids`. Routes add/remove in place.

### spec.routes[].destinationCidrBlock

`string` · required

Destination network, in CIDR notation. "0.0.0.0/0" routes all client
internet traffic through the VPN (full tunnel) — the target subnet
then needs a NAT path or clients lose internet access. Must be the
canonical network address (host bits zero) — the provider rejects
non-canonical forms at plan time.

- rule: destination_cidr_block must be a canonical network address — host bits must be zero (e.g. "10.0.1.0/24", not "10.0.1.5/24")
- rule: {"required":true,"string":{"pattern":"^([0-9]{1,3}\\.){3}[0-9]{1,3}/([0-9]|[1-2][0-9]|3[0-2])$"}}

### spec.routes[].targetSubnetId

`string | valueFrom` · required

The associated subnet traffic to this destination egresses through.
Must be one of the endpoint's `subnet_ids` — AWS rejects a route whose
subnet is not (yet) associated.

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.routes[].description

`string`

Description shown in the AWS console.

### spec.sessionTimeoutHours

`int32` · optional (explicit presence)

Maximum VPN session duration in hours, after which clients must
re-authenticate. Values: 8, 10, 12, 24 (default 24). Updates in place.

- default: `24`
- rule: {"int32":{"in":[8,10,12,24]}}

### spec.disconnectOnSessionTimeout

`bool`

When true, sessions are hard-disconnected at the session timeout and
the user is prompted to reconnect; when false (AWS default), the client
attempts to reconnect automatically. Updates in place.

### spec.selfServicePortalEnabled

`bool`

Enable the self-service portal, where users download their own client
configuration and reset their certificates. Only supported with a
federated (SAML) authentication option. Updates in place.

### spec.clientConnectOptions

`AwsClientVpnClientConnectOptions`

Run a Lambda function on every new connection — allow/deny posture
checks beyond authentication (device compliance, source IP policy,
time-of-day rules). Presence enables the hook; removing the block
disables it (both engines send the explicit disabled state — AWS
otherwise keeps a once-enabled hook active forever). Updates in place.

### spec.clientConnectOptions.lambdaFunctionArn

`string | valueFrom` · required

The Lambda function AWS invokes for each connection attempt. AWS
requires the function name to start with "AWSClientVPN-". The function
receives connection metadata (user, device, source IP) and returns an
allow/deny decision, optionally with an error message shown to the
user.

- references: AwsLambda (`status.outputs.function_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsLambda, name: <that resource's name>, fieldPath: status.outputs.function_arn}} -- a bare string does not parse

### spec.clientLoginBanner

`AwsClientVpnLoginBanner`

Text banner displayed on AWS-provided clients when a session is
established (legal notices, acceptable-use reminders). Presence enables
the banner; removing the block disables it (both engines send the
explicit disabled state). Updates in place.

### spec.clientLoginBanner.bannerText

`string` · required

Banner text, up to 1400 UTF-8 characters — legal notices,
acceptable-use reminders, support contacts.

- rule: {"required":true,"string":{"maxLen":"1400"}}

### spec.clientRouteEnforcementEnabled

`bool`

Enforce administrator-defined routes on connected devices, blocking
client-side route manipulation from bypassing the tunnel (a
security-posture hardening dial for managed fleets). Both engines send
the explicit value in both states, so flipping back to false genuinely
disables enforcement. Updates in place.

### spec.dnsServers

`[]string`

Custom DNS server IPs pushed to connected clients (max 2). When empty,
clients keep their device DNS — set this to the VPC resolver (the VPC
CIDR base + 2, e.g. "10.0.0.2") so clients resolve private hosted
zones. Updates in place.

- rule: {"repeated":{"maxItems":"2","unique":true}}

### spec.connectionLog

`AwsClientVpnConnectionLog`

Stream connection events (connect, disconnect, authentication failures)
to CloudWatch Logs. Presence enables logging — strongly recommended in
production; without it there is no record of who connected. Updates in
place.

### spec.connectionLog.cloudwatchLogGroup

`string | valueFrom` · required

The CloudWatch log group connection events are written to.

- references: AwsCloudwatchLogGroup (`status.outputs.log_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsCloudwatchLogGroup, name: <that resource's name>, fieldPath: status.outputs.log_group_name}} -- a bare string does not parse

### spec.connectionLog.cloudwatchLogStream

`string`

Optional log stream within the group. When empty, AWS creates one.

## Validation Rules

- `client_cidr_required_unless_ipv6_traffic`: client_cidr_block is required — except when traffic_ip_address_type is 'ipv6', where it must be empty (AWS assigns client addressing)
- `tgw_excludes_vpc`: transit_gateway_configuration and vpc_id are mutually exclusive — the endpoint attaches to a transit gateway or to a VPC, not both
- `tgw_excludes_security_groups`: security_group_ids cannot be set together with transit_gateway_configuration — security groups belong to the VPC attachment surface
- `tgw_excludes_subnet_associations`: subnet_ids cannot be set together with transit_gateway_configuration — subnet associations belong to the VPC attachment surface
- `self_service_portal_requires_federated`: the self-service portal is only available with a federated (SAML) authentication option
- `dns_servers_format`: dns_servers must be valid IPv4 addresses (max 2)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsClientVpn, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.client_vpn_endpoint_id` | `string` | The AWS-assigned identifier for the Client VPN endpoint (e.g. "cvpn-endpoint-012345abcdeEXAMPLE"). Used for AWS CLI/API operations — including exporting the client configuration file (`aws ec2 export-client-vpn-client-configuration`). |
| `status.outputs.client_vpn_endpoint_arn` | `string` | The Amazon Resource Name of the endpoint. Used in IAM policies and cross-service permissions. |
| `status.outputs.endpoint_dns_name` | `string` | The DNS name clients connect to. Note AWS's quirk: the OpenVPN client configuration must prepend a random subdomain label to this name (e.g. "asdf.cvpn-endpoint-....amazonaws.com") — the exported client configuration handles this automatically. |
| `status.outputs.self_service_portal_url` | `string` | The URL of the self-service portal where federated users download their own client configuration. Empty when the portal is disabled. |
| `status.outputs.subnet_association_ids` | `map<string, string>` | A map of subnet ID → target network association ID (e.g. "cvpn-assoc-0abcd1234efgh5678") for each associated subnet in `spec.subnet_ids`. Useful for referencing or manually managing a specific association. |
| `status.outputs.transit_gateway_attachment_id` | `string` | The transit gateway attachment created when the endpoint is associated with a transit gateway via `spec.transit_gateway_configuration`. Empty for VPC-attached endpoints. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.authenticationOptions[].rootCertificateChainArn` | AwsCertManagerCert | `status.outputs.cert_arn` |
| `spec.serverCertificateArn` | AwsCertManagerCert | `status.outputs.cert_arn` |
| `spec.vpcId` | AwsVpc | `status.outputs.vpc_id` |
| `spec.securityGroupIds` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.subnetIds` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.transitGatewayConfiguration.transitGatewayId` | AwsTransitGateway | `status.outputs.transit_gateway_id` |
| `spec.routes[].targetSubnetId` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.clientConnectOptions.lambdaFunctionArn` | AwsLambda | `status.outputs.function_arn` |
| `spec.connectionLog.cloudwatchLogGroup` | AwsCloudwatchLogGroup | `status.outputs.log_group_name` |

## See Also

- [Overview](../README.md)

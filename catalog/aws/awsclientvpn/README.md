# Overview

The AwsClientVpn API resource creates an AWS Client VPN endpoint — the
managed OpenVPN server remote users and machines connect to for secure
access into AWS networks — together with its three endpoint-scoped
satellites: target network associations, authorization rules, and routes.

## Why We Created This API Resource

Remote access is a whole posture, not just a server:

- **The endpoint is only the front door**: without an associated subnet,
  an authorization rule, and (for anything beyond the VPC) a route, a
  connected client reaches nothing. All three fold into this one spec
  because none has identity outside its endpoint — and each is still its
  own provider resource, so membership edits apply in place.
- **Three honest authentication arms**: mutual-TLS client certificates,
  Active Directory user/password, or SAML federation (which unlocks the
  self-service portal). Up to two options combine; a client passes on ANY.
- **Composable by reference**: the server certificate and client CA chain
  reference AwsCertManagerCert, subnets reference AwsSubnet, logging
  references AwsCloudwatchLogGroup, the posture-check hook references
  AwsLambda, and the transit-gateway arm references AwsTransitGateway.

## Key Features

### Authentication (create-time immutable)

- **certificate-authentication** — mutual TLS against a client CA chain
  (its own reference, never silently the server certificate)
- **directory-service-authentication** — user/password against AWS
  Directory Service
- **federated-authentication** — SAML 2.0 SSO through an IAM SAML
  provider; the only type that supports the self-service portal

### Access control (folded satellites)

- **`subnetIds`** — target network associations, one per subnet (each
  bills hourly and takes minutes to attach/detach); a zero-association
  endpoint is a valid pre-staged front door
- **`authorizationRules`** — network-ACL-style grants: everyone
  (`authorizeAllGroups`) or one IdP group (`accessGroupId`) per
  destination CIDR
- **`routes`** — entries beyond the auto-created per-VPC route: 0.0.0.0/0
  through a NAT-ed subnet for full-tunnel egress, on-prem CIDRs through a
  gateway-connected subnet

### Tunnel shape and client experience

- Split-tunnel vs full-tunnel, UDP/TCP transport, 443/1194 port
- IPv6/dual-stack endpoint and tunnel addressing
- Session timeout (8/10/12/24h) with optional hard disconnect
- Login banner, connection posture-check Lambda, client route enforcement,
  custom DNS servers, CloudWatch connection logging — each toggles OFF as
  cleanly as on: removing the block (or flipping the flag) sends the
  explicit disabled state on both engines

### Transit-gateway attachment

Attach the endpoint to a transit gateway instead of a VPC — clients reach
every network the gateway routes to, without per-subnet associations.

## Benefits

- **Whole-posture manifests**: one document declares who may connect, what
  they may reach, and how traffic flows — reviewable security posture.
- **Honest constraints**: per-type identity sources, the
  client-CIDR-vs-IPv6 coupling, TGW-vs-VPC exclusivity, and
  exactly-one-grantee rules are CEL-enforced at validation time.
- **Consistency**: identical behavior across Terraform and Pulumi.

## Outputs

- `client_vpn_endpoint_id` / `client_vpn_endpoint_arn`: the endpoint's
  identifiers
- `endpoint_dns_name`: what clients connect to
- `self_service_portal_url`: where federated users download their client
  configuration
- `subnet_association_ids`: subnet ID → association ID map
- `transit_gateway_attachment_id`: for TGW-attached endpoints

## Deliberately Skipped (with reasons)

- **Port↔protocol pairing rules**: AWS allows 443 and 1194 with either
  transport protocol — only the port value set (443/1194) is validated.
- **A `status` output**: the provider exposes no endpoint status
  attribute; deployment state is the deploy's own success signal.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

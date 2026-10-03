# Overview

The AwsVpcEndpoint API resource creates a private connection from a VPC
to an AWS service, a PrivateLink service, or a VPC Lattice resource --
traffic stays on the AWS network instead of crossing the internet
through a NAT or internet gateway.

## Why We Created This API Resource

Private service access is the backbone of every locked-down AWS
architecture, and it deserves a first-class, composable node:

- **One endpoint, one node**: a VPC typically carries several endpoints
  (S3, ECR, STS, CloudWatch Logs, a vendor's PrivateLink service), each
  with its own policy and lifecycle; modeling them individually makes
  the architecture graph show exactly which services a VPC can reach
  privately.
- **Attach by reference**: the VPC, route tables, subnets, and security
  groups are all referenced nodes -- the endpoint never modifies a
  resource it merely references.
- **Honest type gating**: AWS's five endpoint types take different
  attachments (route tables for Gateway; subnets/security groups/private
  DNS for Interface). The spec enforces the gating with CEL so a
  misconfigured endpoint fails at validation, not mid-deploy.

## Key Features

### Endpoint Types

- **Gateway** (S3, DynamoDB): free; injects a prefix-list route into the
  route tables you attach -- the default private path for S3/DynamoDB,
  which also removes that traffic from NAT data-processing charges.
- **Interface** (most AWS services + every third-party PrivateLink
  service): an ENI per subnet, optional private DNS that overrides the
  service's public name inside the VPC, security-group controlled.
- **GatewayLoadBalancer**: fronts a GWLB appliance fleet.
- **Resource / ServiceNetwork**: VPC Lattice attachments, including
  fine-grained private-DNS domain preferences.

### Access and Addressing Controls

- **Endpoint policy**: an IAM policy document scoping which principals
  may use the endpoint to reach which resources -- e.g. "only this
  account's buckets" on an S3 gateway endpoint, a real
  data-exfiltration control. Authored as a structured document (the
  catalog's uniform policy idiom), never an embedded JSON string.
- **Private DNS is tri-state**: unset keeps AWS's default (off) AND
  keeps an existing endpoint's current setting; `true` enables; an
  EXPLICIT `false` is the only way to turn it back off once enabled
  (the provider keeps the existing value whenever the argument is
  omitted).
- **Private DNS options**: record IP type, the S3 dual-stack
  inbound-resolver-only pattern (in-VPC traffic rides the free gateway
  endpoint while on-premises clients resolve to the interface
  endpoint), and Lattice domain preferences.
- **IP address type** (`ipv4`/`dualstack`/`ipv6`) and per-subnet static
  ENI addresses for appliances that need stable endpoint IPs.
- **Cross-region endpoints**: reach a service in another region from an
  interface endpoint, with no cross-region networking of your own.

## Benefits

- **Composability**: VPC, route tables, subnets, and security groups
  attach through `valueFrom` references; the endpoint's prefix list,
  DNS name, and ENIs are outputs downstream nodes consume.
- **Honest constraints**: the service-target exactly-one rule, the
  per-type attachment gating, and the DNS preference/domain coupling
  are CEL-enforced at validation time.
- **Consistency**: identical behavior across Terraform and Pulumi.

## Outputs

- `vpc_endpoint_id`: the endpoint's id (vpce-...)
- `arn`: the endpoint's ARN
- `state`: lifecycle state after provisioning (available /
  pendingAcceptance)
- `prefix_list_id`: the service's prefix list (gateway endpoints only)
- `dns_name` / `hosted_zone_id`: the primary private DNS name and its
  Route53 zone (interface endpoints only)
- `network_interface_ids`: the endpoint's ENIs, one per attached subnet

## Deliberately Skipped (with reasons)

- **The PrivateLink provider side** (`aws_vpc_endpoint_service` and its
  allowed-principal / connection-accepter / connection-notification /
  private-DNS-verification satellites) is a separate product surface --
  publishing your own service behind an NLB/GWLB for others to consume.
  It is not part of consuming services privately and is deferred until
  real demand appears.
- **`aws_vpc_endpoint_policy`** (the standalone policy resource) is
  folded into the spec's `policy` field -- a policy belongs to exactly
  one endpoint.
- **The standalone association resources**
  (`aws_vpc_endpoint_subnet_association`,
  `aws_vpc_endpoint_route_table_association`,
  `aws_vpc_endpoint_security_group_association`) are folded into the
  spec's repeated reference fields -- glue with no independent
  lifecycle.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

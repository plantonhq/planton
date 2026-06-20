# AWS VPC

Deploys a thin AWS Virtual Private Cloud: an isolated IP address space (primary
IPv4 CIDR, optional secondary IPv4 CIDRs, and optional IPv6) with configurable
tenancy and DNS settings. Subnets, gateways, and route tables are separate,
composable components that reference this VPC.

## What Gets Created

When you deploy an AwsVpc resource, OpenMCF provisions:

- **VPC** — an `aws_vpc` / `ec2.Vpc` with the primary IPv4 CIDR (specified
  directly or allocated from an IPAM pool), tenancy, DNS, and optional IPv6.
- **Secondary IPv4 CIDR associations** — one `aws_vpc_ipv4_cidr_block_association`
  / `ec2.VpcIpv4CidrBlockAssociation` per entry in `secondaryIpv4CidrBlocks`,
  each independently associated so it can be added or removed without recreating
  the VPC.

Subnets, internet gateways, NAT gateways, and route tables are **not** created
here — deploy `AwsSubnet`, `AwsInternetGateway`, and `AwsNatGateway` components
that reference this VPC's `vpc_id` output.

## Prerequisites

- **AWS credentials** configured via the OpenMCF provider config (keyless SSO/OIDC).
- **A primary IPv4 source**: either a `cidrBlock` (e.g. `10.0.0.0/16`) or an
  `ipv4IpamPoolId`.

## Quick Start

Create a file `vpc.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsVpc
metadata:
  name: my-vpc
spec:
  region: us-west-2
  cidrBlock: "10.0.0.0/16"
  enableDnsHostnames: true
```

Deploy:

```shell
openmcf apply -f vpc.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | AWS region where the VPC is created (e.g. `us-west-2`). | Required; non-empty |
| `cidrBlock` *or* `ipv4IpamPoolId` | `string` | Primary IPv4 source: an explicit CIDR (e.g. `10.0.0.0/16`) or an IPAM pool. | Exactly one is required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `secondaryIpv4CidrBlocks` | `string[]` | `[]` | Additional IPv4 CIDRs associated with the VPC. |
| `ipv4NetmaskLength` | `int32` | — | Netmask of the primary CIDR to allocate from `ipv4IpamPoolId` (16–28). Requires `ipv4IpamPoolId`; mutually exclusive with `cidrBlock`. |
| `instanceTenancy` | `string` | `default` | `default` (shared) or `dedicated` (single-tenant hardware). |
| `enableDnsSupport` | `bool` | `true` | Amazon-provided DNS resolution within the VPC. Unset keeps it on. |
| `enableDnsHostnames` | `bool` | `false` | Public DNS hostnames for instances with public IPs. |
| `enableNetworkAddressUsageMetrics` | `bool` | `false` | CloudWatch Network Address Usage metrics. |
| `assignGeneratedIpv6CidrBlock` | `bool` | `false` | Request an Amazon-provided IPv6 /56. Mutually exclusive with the IPAM IPv6 fields. |
| `ipv6CidrBlock` | `string` | — | Explicit IPv6 CIDR to allocate from `ipv6IpamPoolId`. Requires `ipv6IpamPoolId`. |
| `ipv6CidrBlockNetworkBorderGroup` | `string` | — | Advertisement border group for an Amazon-provided IPv6 CIDR. Requires `assignGeneratedIpv6CidrBlock`. |
| `ipv6IpamPoolId` | `string` | — | IPAM pool for the IPv6 CIDR. |
| `ipv6NetmaskLength` | `int32` | — | IPv6 netmask to allocate from `ipv6IpamPoolId` (44, 48, 52, 56, or 60). Mutually exclusive with `ipv6CidrBlock`. |

## Examples

### Dual-stack VPC with a secondary CIDR

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsVpc
metadata:
  name: dual-stack-vpc
spec:
  region: us-west-2
  cidrBlock: "10.0.0.0/16"
  secondaryIpv4CidrBlocks:
    - "100.64.0.0/16"
  assignGeneratedIpv6CidrBlock: true
  enableDnsHostnames: true
```

### IPAM-allocated VPC

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsVpc
metadata:
  name: ipam-vpc
spec:
  region: us-west-2
  ipv4IpamPoolId: "ipam-pool-0abc123"
  ipv4NetmaskLength: 16
  enableDnsHostnames: true
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `vpc_id` | `string` | ID of the VPC. Referenced by subnets, gateways, and security groups. |
| `vpc_arn` | `string` | ARN of the VPC. |
| `cidr_block` | `string` | Primary IPv4 CIDR of the VPC. |
| `ipv6_cidr_block` | `string` | IPv6 CIDR of the VPC (empty when IPv4-only). |
| `owner_id` | `string` | AWS account ID that owns the VPC. |
| `main_route_table_id` | `string` | ID of the VPC's main route table. |
| `default_security_group_id` | `string` | ID of the default security group. |
| `default_network_acl_id` | `string` | ID of the default network ACL. |
| `default_route_table_id` | `string` | ID of the default route table. |
| `region` | `string` | Region the VPC was created in. |

## Related Components

- [AwsSubnet](/docs/catalog/aws/awssubnet) — a subnet within the VPC (with routing folded in)
- [AwsInternetGateway](/docs/catalog/aws/awsinternetgateway) — internet access for public subnets
- [AwsNatGateway](/docs/catalog/aws/awsnatgateway) — outbound internet access for private subnets
- [AwsSecurityGroup](/docs/catalog/aws/awssecuritygroup) — controls network traffic for resources in the VPC

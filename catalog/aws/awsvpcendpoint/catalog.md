# AWS VPC Endpoint

Deploys an Amazon VPC endpoint — a private connection from a VPC to an AWS service, a third-party PrivateLink service, or a VPC Lattice target, so traffic stays on the AWS network instead of crossing the internet through a NAT or internet gateway. Two endpoint types carry nearly all real-world use: **Gateway** (S3 and DynamoDB only — attaches by injecting a prefix-list route into your route tables, and takes that traffic off the NAT data-processing meter) and **Interface** (everything else — an ENI in each subnet you attach, with cost driven per AZ-hour and per gigabyte, and private DNS that keeps client code unchanged). The endpoint composes onto its neighbors by reference — the [AWS VPC](/infra-catalog/aws-vpc), route tables from [AWS Subnet](/infra-catalog/aws-subnet) or the VPC's own outputs, subnets, and [AWS Security Group](/infra-catalog/aws-security-group) nodes — and never modifies a resource it merely references.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **VPC Endpoint** -- gateway, interface, Gateway Load Balancer, or VPC Lattice (Resource / ServiceNetwork) type, connected to exactly one service target
- **Route Injection (gateway)** -- the service's prefix-list route in every route table you attach; traffic from subnets on those tables flows privately from that moment
- **Network Interfaces (interface/GWLB/Lattice)** -- one ENI per attached subnet, guarded on the interface type by the security groups you reference
- **Private DNS (interface)** -- an AWS-managed private hosted zone that resolves the service's public name to the endpoint inside the VPC, so SDKs need zero changes
- **Endpoint Policy (gateway/interface)** -- an optional IAM document scoping which principals may reach which resources through this path
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **The VPC first** -- deploy the [AWS VPC](/infra-catalog/aws-vpc) this endpoint lives in; the endpoint references its `vpc_id` output. For private DNS, the VPC needs BOTH DNS support and DNS hostnames enabled.

### AWS Account

- **Route tables ready (gateway)** -- know which tables carry the subnets whose traffic should go private: an [AWS Subnet](/infra-catalog/aws-subnet) with inline routes owns its table (`route_table_id` output); subnets without one ride the VPC main table (`main_route_table_id` output).
- **Security-group ingress (interface)** -- the groups you attach must allow inbound from your clients on the service's port (443 for AWS APIs); the rules live on the referenced [AWS Security Group](/infra-catalog/aws-security-group) nodes.

## Deploy

### Console

Open the deployment store, find **AWS VPC Endpoint**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields — the endpoint type reshapes the flow to exactly the decisions that type needs. Start from the **S3 Gateway Endpoint** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsVpcEndpoint
metadata:
  name: s3-private-path
  org: acme-corp
  env: dev
spec:
  region: us-west-2
  vpcId:
    valueFrom:
      kind: AwsVpc
      name: app-network
      fieldPath: status.outputs.vpc_id
  endpointType: Gateway
  serviceName: com.amazonaws.us-west-2.s3
  routeTableIds:
    - valueFrom:
        kind: AwsSubnet
        name: private-az1
        fieldPath: status.outputs.route_table_id
```

```shell
planton apply -f vpc-endpoint.yaml
```

This gives the private subnet's workloads a private path to S3, off the NAT data-processing meter. An Infra Job tracks the provisioning in real time.

### InfraChart

When the endpoint deploys alongside its VPC and subnets in one chart, wire the references via ValueFromRef:

```yaml
spec:
  region: us-west-2
  vpcId:
    valueFrom:
      kind: AwsVpc
      name: app-network
      fieldPath: status.outputs.vpc_id
  endpointType: Gateway
  serviceName: com.amazonaws.us-west-2.s3
  routeTableIds:
    - valueFrom:
        kind: AwsSubnet
        name: private-az1
        fieldPath: status.outputs.route_table_id
```

The InfraPipeline resolves the dependency graph, deploys the VPC and subnet first, then injects the endpoint's prefix-list route into the subnet's table.

## Key Configuration

These are the most important decisions when configuring a VPC endpoint. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The endpoint type** -- create-time immutable, and it decides everything downstream: Gateway attaches through `routeTableIds`, every other type through `subnetIds` ENIs; security groups and private DNS exist only on Interface; the Lattice types name their target by ARN instead of a service name. The console wizard reshapes per type so an illegal combination cannot be expressed.

**Exactly one service target** -- `serviceName` (AWS services and PrivateLink providers' `com.amazonaws.vpce.…` names), `resourceConfigurationArn` (Lattice Resource), or `serviceNetworkArn` (Lattice ServiceNetwork). The service name embeds the region — it must match the endpoint's placement.

**Private DNS (interface)** -- ON means the service's public name (e.g. `sts.us-west-2.amazonaws.com`) resolves to the endpoint's private IPs inside the VPC: zero client changes. Tri-state: leave it unset and AWS keeps its default (off) — and keeps an existing endpoint's current setting; set an EXPLICIT `false` to turn private DNS back off once enabled. The S3 dual-stack pattern combines a gateway endpoint for in-VPC traffic with an interface endpoint whose `dnsOptions.privateDnsOnlyForInboundResolverEndpoint` serves only on-premises resolver traffic (requires private DNS enabled — validated).

**The endpoint policy** -- empty means full access (the endpoint is purely a network path). A custom document turns the private path into a governance point — the classic S3 control allows only your organization's principals or your own buckets, so a compromised workload cannot exfiltrate data through your own endpoint. It restricts; it never grants what IAM has not. The policy is authored as a structured document (`policy:` with `Version`/`Statement` as YAML), matching every other policy field in the catalog.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsVpc** | `vpcId` | `status.outputs.vpc_id` |
| **AwsSubnet** (or **AwsVpc** main/default tables) | `routeTableIds[]` (gateway attachment) | `status.outputs.route_table_id` |
| **AwsSubnet** | `subnetIds[]` (ENI placement for interface/GWLB/Lattice) | `status.outputs.subnet_id` |
| **AwsSecurityGroup** | `securityGroupIds[]` (interface endpoints) | `status.outputs.security_group_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `vpc_endpoint_id` | The endpoint's id (vpce-…) | Subnet route targets (GWLB middlebox routing), AWS APIs |
| `arn` | Amazon Resource Name of the endpoint | IAM policies and audits |
| `state` | Lifecycle state after provisioning | `pendingAcceptance` signals a PrivateLink service awaiting approval |
| `prefix_list_id` | The service's prefix list (gateway only) | Security-group egress rules scoped to the service's address ranges |
| `dns_name` | The endpoint-specific private DNS name (interface only) | Client configs when private DNS is off; Route53 alias targets |
| `hosted_zone_id` | The Route53 zone of `dns_name` (interface only) | Paired with the DNS name for alias records |
| `network_interface_ids[]` | The endpoint's ENIs, one per subnet | Flow logs, firewall rules, IP lookups |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**S3 gateway endpoint** -- the default private path for S3: inject the prefix-list route into your private subnets' tables and the NAT data-processing charge for S3 traffic disappears. Start from the **S3 Gateway Endpoint** preset.

**Interface endpoint with private DNS** -- an ENI-based path to STS, ECR, CloudWatch Logs, Secrets Manager, SSM, or KMS across two AZs with private DNS on — workloads keep their default SDK endpoints. Start from the **Interface Endpoint for an AWS Service** preset. (Private ECR pulls need `ecr.api` + `ecr.dkr` + the S3 gateway.)

**Third-party PrivateLink service** -- connect to a vendor's `com.amazonaws.vpce.…` service name; cross-account services must accept the connection on their side (the `state` output shows `pendingAcceptance` until they do). Start from the **Third-Party PrivateLink Service** preset.

## Works With

- [**AWS VPC**](/infra-catalog/aws-vpc) -- the network that gets the private path (references `vpc_id`; its main/default route tables are gateway attachment points)
- [**AWS Subnet**](/infra-catalog/aws-subnet) -- route tables for gateway endpoints, ENI placement for interface endpoints — and its route rows can target this endpoint's id (GWLB middlebox routing)
- [**AWS Security Group**](/infra-catalog/aws-security-group) -- guards an interface endpoint's ENIs; its egress rules can reference a gateway endpoint's `prefix_list_id`
- [**AWS MWAA Environment**](/infra-catalog/aws-mwaa-environment) -- in CUSTOMER endpoint-management mode, you create the VPC endpoints MWAA needs — with exactly this kind

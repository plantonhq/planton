# AWS HTTP API VPC Link

Deploys an API Gateway v2 VPC link — the managed network attachment that lets HTTP APIs reach private backends inside a VPC without exposing them to the internet: an internal ALB, an NLB, or a Cloud Map service. The link is deliberately its own resource: one link is shared by any number of APIs and integrations (each private integration references the link by its ID), and the link owns its own network-attachment lifecycle — AWS provisions cross-account ENIs into the chosen subnets that persist across API create/destroy cycles. One link per VPC is typically enough.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **API Gateway v2 VPC Link** -- the managed attachment HTTP APIs route through (this is the v2 link for HTTP APIs; REST APIs use a different, NLB-only v1 link)
- **Managed Network Interfaces** -- cross-account ENIs in your chosen subnets (create-time immutable; the link can only reach targets in AZs it has an ENI in)
- **Security Group Bindings** -- the groups governing what the link's ENIs may reach (create-time immutable; empty means no link-side filtering)
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Network resources** -- the target VPC's [AWS Subnet](/infra-catalog/aws-subnet) resources (at least one; two AZs for production) and ideally an egress-scoped [AWS Security Group](/infra-catalog/aws-security-group), deployed first and referenced.

### AWS Account

- **The backend's security group must admit the link** -- allow ingress on the listener ports FROM the link's security groups (reference by group ID, never CIDR).
- **AZ alignment** -- the link only reaches targets in AZs it has an ENI in; mirror the internal load balancer's AZ spread.

## Deploy

### Console

Open the deployment store, find **AWS HTTP API VPC Link**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Private ALB Link** preset in the [Presets](#presets) tab to pre-populate the production shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsHttpApiVpcLink
metadata:
  name: private-services-link
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  subnetIds:
    - valueFrom:
        kind: AwsSubnet
        name: private-az1
        fieldPath: status.outputs.subnet_id
    - valueFrom:
        kind: AwsSubnet
        name: private-az2
        fieldPath: status.outputs.subnet_id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: vpc-link-egress
        fieldPath: status.outputs.security_group_id
```

```shell
planton apply -f http-api-vpc-link.yaml
```

This creates a two-AZ link with an egress-scoped security group — the shape every HTTP API in the VPC shares. An Infra Job tracks the provisioning in real time.

### InfraChart

When the link deploys alongside its subnets and security group in one chart, wire the references via ValueFromRef:

```yaml
spec:
  region: us-west-2
  subnetIds:
    - valueFrom:
        kind: AwsSubnet
        name: private-az1
        fieldPath: status.outputs.subnet_id
    - valueFrom:
        kind: AwsSubnet
        name: private-az2
        fieldPath: status.outputs.subnet_id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: vpc-link-egress
        fieldPath: status.outputs.security_group_id
```

The InfraPipeline resolves the dependency graph, deploys the subnets and security group first, then attaches the link to them.

## Key Configuration

These are the most important decisions when configuring a VPC link. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Subnets are the one-way door** -- AWS has no update API for the link's network attachment: changing the subnet or security-group set replaces the link, and every private integration referencing its ID re-homes with it. Spread across at least two availability zones at create time; the link cannot reach backends in AZs it has no ENI in.

**Security groups are the reach contract** -- the link's groups allow egress to the backend listener ports; the backend's group admits ingress from the link's groups. Omitting them creates the link with no link-side filtering — workable, but the set is immutable, so the production-grade group cannot be added later.

**One link, many APIs** -- the link is infrastructure, not per-API config. Deploy it once per VPC and let every HTTP API's private integrations share it by `connectionId`.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsSubnet** | `subnetIds[]` | `status.outputs.subnet_id` |
| **AwsSecurityGroup** (optional) | `securityGroupIds[]` | `status.outputs.security_group_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `vpc_link_id` | The VPC link identifier | HTTP API private integrations' `connectionId` |
| `vpc_link_arn` | Amazon Resource Name of the link | IAM policies and tag-based governance |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Private ALB frontend** -- HTTP APIs fronting an internal Application Load Balancer; two AZs, an egress-scoped security group. Start from the **Private ALB Link** preset.

**Minimal link** -- a single-subnet development link with no security groups. Start from the **Minimal Link** preset.

## Works With

- [**AWS HTTP API Gateway**](/infra-catalog/aws-http-api-gateway) -- the API whose private integrations route through this link (references `vpc_link_id`)
- [**AWS Subnet**](/infra-catalog/aws-subnet) -- where the link's ENIs land
- [**AWS Security Group**](/infra-catalog/aws-security-group) -- the egress contract to the private backend
- [**AWS ALB**](/infra-catalog/aws-alb) -- the internal load balancer the link typically fronts

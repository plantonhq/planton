# AWS App Runner VPC Connector

Deploys an App Runner VPC connector — the managed network attachment that lets [App Runner services](/infra-catalog/aws-app-runner-service) reach private resources inside a VPC (databases, caches, internal APIs) for their OUTBOUND traffic. It is deliberately its own resource: one connector is shared by any number of services, each referencing it by ARN in its egress configuration, and the connector owns the network-attachment lifecycle — AWS provisions managed ENIs into the chosen subnets that persist across service create/destroy cycles. Its subnets and security groups accept ValueFromRef wiring to AwsSubnet and AwsSecurityGroup resources, and every attribute is immutable after creation — a change replaces the connector as a new revision under the same name.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **App Runner VPC Connector** -- a named, versioned network attachment; every attribute is fixed at creation (AWS has no update API), and a change registers a new connector revision under the same name
- **Managed ENIs** -- one network interface per attached subnet, provisioned and owned by AWS; egress routes only through Availability Zones the connector has an ENI in
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Subnets and security groups** -- create the [AwsSubnet](/infra-catalog/aws-subnet) and [AwsSecurityGroup](/infra-catalog/aws-security-group) resources first (declaration before reference), or have existing ids ready as literals.

### AWS Account

- **Same-VPC subnets** -- all attached subnets must belong to one VPC; provide at least two Availability Zones for high availability.
- **The two-sided handshake** -- the connector's security groups must allow egress to the target ports, AND the targets' security groups must admit ingress from these groups. Reference by group id, never by CIDR — the ENI addresses are AWS-managed.
- **Name length** -- the connector's cloud name is `metadata.name`, and the provider validates it at 4–40 characters; keep resource names inside that range.

## Deploy

### Console

Open the deployment store, find **AWS App Runner VPC Connector**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields — the subnet and security group pickers read your connected AWS account live. Start from the **Private Backend Access** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsAppRunnerVpcConnector
metadata:
  name: private-backend-access
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  subnetIds:
    - valueFrom:
        kind: AwsSubnet
        name: private-a
        fieldPath: status.outputs.subnet_id
    - valueFrom:
        kind: AwsSubnet
        name: private-b
        fieldPath: status.outputs.subnet_id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: app-runner-egress
        fieldPath: status.outputs.security_group_id
```

```shell
planton apply -f app-runner-vpc-connector.yaml
```

This creates a two-AZ connector wearing a dedicated egress group. An Infra Job tracks the provisioning in real time.

### InfraChart

When the connector deploys alongside its network in one chart, wire the subnet and security group references via ValueFromRef:

```yaml
spec:
  subnetIds:
    - valueFrom:
        kind: AwsSubnet
        name: private-a
        fieldPath: status.outputs.subnet_id
    - valueFrom:
        kind: AwsSubnet
        name: private-b
        fieldPath: status.outputs.subnet_id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: app-runner-egress
        fieldPath: status.outputs.security_group_id
```

The InfraPipeline resolves the dependency graph, deploys the subnets and security group first, then provisions the connector — and the App Runner services that reference its `vpc_connector_arn` output deploy after it.

## Key Configuration

These are the most important decisions when configuring a VPC connector. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Egress only** -- the connector governs OUTBOUND traffic. Inbound traffic to a private App Runner service travels through a separate VPC Ingress Connection resource created against the service's ARN — a different surface.

**AZ coverage is availability** -- egress routes only through Availability Zones the connector has an ENI in. Attach subnets in at least two AZs; a single-subnet connector loses all VPC egress when its AZ degrades. Use PRIVATE subnets — the connector needs no internet path of its own.

**Everything is fixed at creation** -- AWS has no connector update API. Changing the subnets or security groups replaces the connector as a new revision under the same name; referencing services swap onto it on their next deployment. Attach the subnets you will want in a year, not just today's.

**Security groups are the egress contract** -- they govern what every referencing service can reach. Allow egress to the targets' ports here, and make the targets' groups admit ingress FROM these groups (by group id, never CIDR — ENI addresses are AWS-managed and change on replacement).

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsSubnet** | `subnetIds[]` | `status.outputs.subnet_id` |
| **AwsSecurityGroup** | `securityGroupIds[]` | `status.outputs.security_group_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `vpc_connector_arn` | Revision-carrying ARN of the connector | App Runner service `vpcConnectorArn` |
| `vpc_connector_revision` | The revision number this deployment registered | Audit and rollout tracking |
| `status` | The connector's AWS lifecycle status | Operational verification |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Private backend access** -- two private subnets across AZs plus a dedicated `app-runner-egress` security group, shared by every service that needs the VPC's databases and caches. Start from the **Private Backend Access** preset.

**One connector per VPC per region** -- connectors are shared, so a fleet needs only one per VPC; per-service connectors multiply ENIs without isolation benefit (isolation lives in the security groups).

**Dedicated egress group** -- pair the connector with a purpose-built security group referenced by the target resources' ingress rules, so granting a new backend is one rule on the target side.

## Works With

- [**AWS App Runner Service**](/infra-catalog/aws-app-runner-service) -- routes outbound traffic through this connector via `vpcConnectorArn` (consumes `vpc_connector_arn`)
- [**AWS Subnet**](/infra-catalog/aws-subnet) -- hosts the connector's managed ENIs (provides `subnet_id`)
- [**AWS Security Group**](/infra-catalog/aws-security-group) -- governs what the ENIs may reach (provides `security_group_id`)

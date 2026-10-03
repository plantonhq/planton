# AWS MSK Serverless Cluster

Deploys an Amazon MSK Serverless cluster — Apache Kafka with every capacity decision removed. AWS scales throughput and partitions automatically and bills per use: no broker counts, no instance types, no storage provisioning, no Kafka version management. The spec is exactly what remains — where the cluster lives (one or more VPC placements, each with subnets and security groups) — and all of it is create-time immutable. Authentication is fixed at SASL/IAM on port 9098 (the only scheme, always on), so producer and consumer identity is pure IAM. The cluster exports the bootstrap string and cluster ARN that downstream consumers and Lambda event source mappings wire to.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **MSK Serverless Cluster** -- an AWS-managed Kafka fleet with automatic throughput and partition scaling, SASL/IAM authentication always on
- **VPC Network Interfaces** -- ENIs in every declared VPC placement's subnets, through which that VPC's clients reach the bootstrap endpoint
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- **Subnets in at least two AZs** (production) within the target VPC — clients in an AZ with no cluster interface cross AZs on every fetch. Reference AwsSubnet Infra Components or provide subnet IDs directly.
- **Security groups with TCP-9098 ingress** (up to 5) from your producer/consumer security groups. The ingress rules live on the referenced [AWS Security Group](/infra-catalog/aws-security-group) resources — decide them BEFORE creating the cluster; the set is immutable and empty falls back to the VPC's default group.
- **IAM permissions for clients** -- every producer/consumer needs `kafka-cluster:Connect` plus topic-level `kafka-cluster:*` actions on the cluster ARN; network reachability alone is not enough.

## Deploy

### Console

Open the deployment store, find **AWS MSK Serverless Cluster**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the single networking step. Start from the **Basic IAM-Authenticated Serverless Kafka** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsMskServerlessCluster
metadata:
  name: events-serverless
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  vpcConfigs:
    - subnetIds:
        - value: subnet-0a1b2c3d4e5f00001
        - value: subnet-0a1b2c3d4e5f00002
      securityGroupIds:
        - value: sg-0123456789abcdef0
```

```shell
planton apply -f msk-serverless.yaml
```

This creates a serverless Kafka cluster reachable through two AZs, gated by one security group, with SASL/IAM authentication on port 9098. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the placement from the resource graph:

```yaml
spec:
  vpcConfigs:
    - subnetIds:
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
            name: kafka-clients
            fieldPath: status.outputs.security_group_id
```

The InfraPipeline resolves the dependency graph, deploys the subnets and security group first, then provisions the cluster with the resolved values.

## Key Configuration

These are the only decisions an MSK Serverless cluster asks for — everything else is AWS-managed. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The whole spec is a one-way door** -- region and every VPC placement are create-time immutable; the cluster is an atomic placement record (only tags change in place). Changing any of them replaces the cluster, and topics/data do not migrate — decide the placement set deliberately.

**VPC placements** -- one `vpcConfigs` entry per VPC. AWS provisions client-facing network interfaces in EACH declared VPC, so workloads in every listed VPC connect privately with no peering, transit routing, or PrivateLink setup. One entry is the common shape; add more to share the cluster across VPCs (each entry's subnets must belong to the same VPC).

**Subnets** -- where a placement's network interfaces land. Private subnets in at least two availability zones is the production shape; a single-AZ placement is a single point of failure.

**Security groups** -- up to 5 per placement, gating network access to the bootstrap endpoint from that VPC. The TCP-9098 ingress rules are composed on first-class AwsSecurityGroup resources, never embedded here. Empty attaches the VPC's default group — attach a purpose-made group at create time; it cannot be added later.

**What is deliberately absent** -- no broker/storage/version fields (AWS manages capacity), no auth fields (SASL/IAM is the only scheme and is always on), no ingress rules (composed on security groups). If a workload needs steady multi-MB/s throughput, SCRAM/mTLS auth, or version pinning, the provisioned [AWS MSK Cluster](/infra-catalog/aws-msk-cluster) is the right kind.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsSubnet** | `vpcConfigs[].subnetIds` | `status.outputs.subnet_id` |
| **AwsSecurityGroup** (optional) | `vpcConfigs[].securityGroupIds` | `status.outputs.security_group_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `bootstrap_brokers_sasl_iam` | The SASL/IAM bootstrap endpoint list (port 9098) — the only connection string serverless MSK exposes | Kafka client configuration |
| `cluster_arn` | Amazon Resource Name of the cluster | IAM `kafka-cluster:*` policies; Lambda event source mappings |
| `cluster_name` | Human-readable cluster name | Operational scripts, monitoring dashboards |
| `cluster_uuid` | Unique id extracted from the cluster ARN | Fine-grained IAM resource patterns |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Event-driven microservices** -- spiky, unpredictable throughput where paying per use beats provisioning for peak. Start from the **Basic IAM-Authenticated Serverless Kafka** preset.

**Graph-composed placement** -- subnets and security groups wired from the resource graph so the whole event fabric deploys as one InfraChart. Start from the **Composed Serverless Kafka (Full References)** preset.

**Shared cluster across VPCs** -- one Kafka fabric serving producers and consumers in separate VPCs, each through its own placement and security group. Start from the **Multi-VPC Serverless Kafka** preset.

## Works With

- [**AWS Subnet**](/infra-catalog/aws-subnet) -- provides the VPC placement for the cluster's network interfaces
- [**AWS Security Group**](/infra-catalog/aws-security-group) -- carries the TCP-9098 ingress rules gating client access
- [**AWS Lambda Event Source Mapping**](/infra-catalog/aws-lambda-event-source-mapping) -- consumes topics from this cluster via the exported cluster ARN
- [**AWS MSK Cluster**](/infra-catalog/aws-msk-cluster) -- the provisioned sibling for steady high-throughput or non-IAM auth workloads

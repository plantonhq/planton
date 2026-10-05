# AwsMskServerlessCluster

Amazon MSK Serverless cluster resource for Planton. Provisions a fully managed, capacity-managed Apache Kafka cluster on AWS: no brokers, instance types, storage volumes, or Kafka version to declare — AWS scales compute and storage automatically and bills per throughput and storage consumed. The whole declaration is where the cluster lives (one or more VPC placements); clients authenticate with AWS IAM (SASL/IAM) on port 9098.

## When to use

- You want Kafka without capacity planning: spiky, unpredictable, or low-duty-cycle streaming workloads where paying per-throughput beats paying for idle brokers.
- You want IAM-native authentication with zero credential management — SASL/IAM is the only scheme serverless MSK supports, and it is always on.
- You are starting a new event-driven architecture and want the lowest-operations entry point to Kafka on AWS.
- Choose the provisioned `AwsMskCluster` instead when you need SASL/SCRAM or mTLS auth, public access, PrivateLink multi-VPC connectivity, tiered storage, custom `server.properties`, or sustained high throughput where provisioned brokers are cheaper.

## Prerequisites

| Prerequisite | Why | Planton Resource |
|---|---|---|
| VPC with private subnets in 2+ AZs | The cluster places its network interfaces in referenced subnets | `AwsVpc` |
| Security group (optional, ≤5) | Attached to the cluster network interfaces; the ingress rule for the SASL/IAM port (9098) lives on this first-class node. AWS attaches the VPC default group when omitted. | `AwsSecurityGroup` |

## API envelope

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsMskServerlessCluster
metadata:
  name: <resource-id>
spec: { ... }
```

## Spec fields reference

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `region` | string | **yes** | — | AWS region for the cluster. |
| `vpcConfigs` | list(object) | **yes** (≥1) | — | VPC placements. AWS provisions client-facing network interfaces in EACH declared VPC, so applications in every listed VPC connect privately — no peering or PrivateLink setup. **ForceNew**. |
| `vpcConfigs[].subnetIds` | list(StringValueOrRef) | **yes** (≥1) | — | Subnets hosting this placement's network interfaces (2+ AZs recommended; all subnets in one entry must belong to the same VPC). Supports `value` or `valueFrom` (AwsSubnet). |
| `vpcConfigs[].securityGroupIds` | list(StringValueOrRef) | no (≤5) | VPC default group | Security groups attached to this placement's network interfaces. Ingress for port 9098 lives on the referenced `AwsSecurityGroup` nodes. |

The resource is effectively immutable: every field above is create-time (ForceNew) in the AWS provider — only tags change in place. SASL/IAM authentication is AWS's sole, mandatory scheme and is enabled unconditionally by both IaC modules, so it is not a spec field.

## Outputs

| Output | Description |
|---|---|
| `cluster_arn` | ARN of the cluster (also its resource identifier) — referenced in IAM policies (`kafka-cluster:*` actions) and Lambda event source mappings. |
| `cluster_name` | Human-readable cluster name. |
| `cluster_uuid` | Unique identifier extracted from the cluster ARN. |
| `bootstrap_brokers_sasl_iam` | Comma-separated SASL/IAM broker endpoint list (port 9098) — the only connection string serverless MSK exposes. |

## Deliberately omitted (with reasons)

| Provider surface | Why omitted |
|---|---|
| `client_authentication.sasl.iam.enabled` as a field | AWS requires it true and offers no alternative scheme — a field that can only ever hold one value is decoration, not configuration. Both modules hardcode it. |
| Kafka version, broker sizing, storage, tiered storage, `server.properties` | Serverless MSK has none of these — AWS manages capacity and version. Use the provisioned `AwsMskCluster` for that control. |
| SASL/SCRAM, mTLS, unauthenticated access, public access, PrivateLink connectivity | Not supported by serverless MSK — provisioned-cluster surfaces. |

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsMskServerlessCluster
metadata:
  name: events-kafka
spec:
  region: us-west-2
  vpcConfigs:
    - subnetIds:
        - valueFrom:
            kind: AwsSubnet
            name: platform-private-a
            fieldPath: status.outputs.subnet_id
        - valueFrom:
            kind: AwsSubnet
            name: platform-private-b
            fieldPath: status.outputs.subnet_id
      securityGroupIds:
        - valueFrom:
            kind: AwsSecurityGroup
            name: kafka-broker-sg
            fieldPath: status.outputs.security_group_id
```

The referenced `kafka-broker-sg` carries the ingress rule for port 9098 (e.g. from the application tier's security group), keeping network policy on composable first-class nodes.

### Multi-VPC access

Declare additional `vpcConfigs` entries to serve clients in other VPCs privately — AWS provisions network interfaces in every listed VPC, so each VPC's workloads connect without peering, transit routing, or PrivateLink endpoints. Each placement carries its own security group, keeping per-VPC network policy where that VPC's clients live (see the `03-multi-vpc-access` preset).

## Client IAM permissions

Serverless MSK authorizes every client action through IAM. A producing/consuming workload needs a policy over the cluster, topic, and group resources, for example `kafka-cluster:Connect`, `kafka-cluster:CreateTopic`, `kafka-cluster:WriteData`, `kafka-cluster:ReadData`, and `kafka-cluster:AlterGroup` scoped to this cluster's ARN (exported as `cluster_arn`).

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

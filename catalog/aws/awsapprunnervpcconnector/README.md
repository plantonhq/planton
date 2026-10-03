# AwsAppRunnerVpcConnector

Deploy and manage an AWS App Runner VPC connector using Planton -- the managed network attachment that lets App Runner services reach private resources inside a VPC (databases, caches, internal APIs) for their outbound traffic.

## Overview

A VPC connector is a set of AWS-managed elastic network interfaces provisioned into subnets you choose. App Runner services reference the connector by ARN in their egress configuration and route all outbound traffic through it into the VPC.

The connector is deliberately its own resource rather than fields on the service:

- **Shared by many services** -- one connector serves any number of App Runner services, so a fleet shares one VPC egress path tuned in one place.
- **Owns its network attachment** -- AWS provisions ENIs into the chosen subnets that persist across service create/destroy cycles.
- **Immutable attachment** -- AWS has no update API for connectors; changing subnets or security groups replaces the connector (a new revision under the same name).

The connector governs EGRESS only. Inbound traffic to a private App Runner service travels through a separate VPC Ingress Connection resource -- a different surface, referenced against the service's exported ARN.

## When to Use

- App Runner services that read from **RDS, ElastiCache, DocumentDB**, or other VPC-internal data stores.
- Reaching **internal APIs** (an internal ALB, ECS services) without exposing them publicly.
- Centralizing VPC egress: one connector per VPC, shared across every service that needs backends there.

## Prerequisites

- Subnets in the target VPC (ideally two or more availability zones -- App Runner routes egress only through AZs the connector has an ENI in). All subnets must belong to the same VPC.
- At least one security group governing what the connected services can reach; the targets' security groups must admit ingress from it.

## Minimal Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsAppRunnerVpcConnector
metadata:
  name: private-backend-access
  org: my-org
  env: prod
  id: private-backend-access-prod
spec:
  region: us-east-1
  subnetIds:
    - valueFrom:
        kind: AwsSubnet
        name: private-subnet-a
        fieldPath: status.outputs.subnet_id
    - valueFrom:
        kind: AwsSubnet
        name: private-subnet-b
        fieldPath: status.outputs.subnet_id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: apprunner-egress
        fieldPath: status.outputs.security_group_id
```

## Spec Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `region` | string | Yes | AWS region for the connector. |
| `subnetIds` | StringValueOrRef[] | Yes (min 1) | Subnets for the connector's ENIs. Immutable -- changing the set replaces the connector. Spread across at least two AZs. |
| `securityGroupIds` | StringValueOrRef[] | Yes (min 1) | Security groups on the connector's ENIs -- they govern what connected services can reach. Immutable. |

## Outputs

| Output | Description |
|--------|-------------|
| `vpc_connector_arn` | The ARN services reference via `vpcConnectorArn`. |
| `vpc_connector_revision` | The revision of this connector under its name. |
| `status` | Lifecycle status at the end of deployment (`ACTIVE` when ready). |

## Composing into a Service

```yaml
# In an AwsAppRunnerService spec:
vpcConnectorArn:
  valueFrom:
    kind: AwsAppRunnerVpcConnector
    name: private-backend-access
    fieldPath: status.outputs.vpc_connector_arn
```

## Deliberately Omitted

- **VPC Ingress Connections** (`aws_apprunner_vpc_ingress_connection`): the INBOUND private-access plane (PrivateLink into a private service) -- a separate surface that composes against the service's exported ARN; deferred until concrete pull.
- **Per-kind tags**: identity tags derive from metadata; custom user tags are a platform-wide concern.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

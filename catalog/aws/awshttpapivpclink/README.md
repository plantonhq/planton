# AwsHttpApiVpcLink

Deploy and manage an API Gateway v2 VPC link using Planton -- the managed network attachment that lets HTTP APIs reach private backends (an internal ALB, an NLB, or a Cloud Map service) inside a VPC without exposing them to the internet.

## Overview

A VPC link is a set of AWS-managed elastic network interfaces provisioned into subnets you choose. HTTP API private integrations reference the link by ID (`connection_id` with `connection_type: VPC_LINK`) and proxy HTTP traffic through it to targets inside the VPC.

The link is deliberately its own resource rather than a field on the API:

- **Shared by many APIs** -- one link serves any number of APIs and integrations, so it outlives any single API's lifecycle.
- **Owns its network attachment** -- AWS provisions ENIs into the chosen subnets that persist across API create/destroy cycles.
- **Immutable attachment** -- AWS has no update API for subnets or security groups; changing either replaces the link. Only the name mutates in place.

## When to Use

- Front an **internal ALB or NLB** (a service mesh entry point, an ECS/EKS ingress) with a public HTTP API without making the load balancer internet-facing.
- Reach a **Cloud Map service** registered by ECS service discovery.
- Centralize private connectivity: create one link per VPC and share it across every API that needs backends there.

## Prerequisites

- Subnets in the target VPC (ideally two or more availability zones -- the link can only reach targets in AZs it has an ENI in).
- (Optional) A security group governing what the link's ENIs can reach; the target's security group must admit ingress from it.

## Minimal Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsHttpApiVpcLink
metadata:
  name: private-services-link
  org: my-org
  env: prod
  id: private-services-link-prod
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
        name: vpc-link-egress
        fieldPath: status.outputs.security_group_id
```

## Spec Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `region` | string | Yes | AWS region for the VPC link. |
| `subnetIds` | StringValueOrRef[] | Yes (min 1) | Subnets for the link's ENIs. Immutable -- changing the set replaces the link. Spread across at least two AZs for high availability. |
| `securityGroupIds` | StringValueOrRef[] | No | Security groups on the link's ENIs. Immutable. When omitted, AWS applies no filtering on the link side. |

## Outputs

| Output | Description |
|--------|-------------|
| `vpc_link_id` | The VPC link ID -- what private integrations set as `connection_id`. |
| `vpc_link_arn` | ARN of the VPC link, for IAM policies and tag-based governance. |

## Composing a Private Integration

```yaml
# In an AwsHttpApiGateway spec:
routes:
  - routeKey: "$default"
    integration:
      integrationType: HTTP_PROXY
      integrationUri:
        value: arn:aws:elasticloadbalancing:us-east-1:123456789012:listener/app/internal/50dc6c495c0c9188/f2f7dc8efc522ab2
      connectionType: VPC_LINK
      connectionId:
        valueFrom:
          kind: AwsHttpApiVpcLink
          name: private-services-link
          fieldPath: status.outputs.vpc_link_id
```

## Deliberately Omitted

- **REST API (v1) VPC links** (`aws_api_gateway_vpc_link`): a different, NLB-only resource belonging to the REST API surface.
- **Per-kind tags**: identity tags derive from metadata; custom user tags are a platform-wide concern.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

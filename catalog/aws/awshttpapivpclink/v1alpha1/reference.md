# AwsHttpApiVpcLink

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsHttpApiVpcLinkSpec defines the desired configuration for an API Gateway
v2 VPC link -- the managed network attachment that lets HTTP APIs reach
private backends (an internal ALB, an NLB, or a Cloud Map service) inside a
VPC without exposing them to the internet.

A VPC link is deliberately its own resource rather than a field on the API:
one link is shared by any number of APIs and integrations (each private
integration references the link by ID), and the link owns its own network
attachment lifecycle -- AWS provisions cross-account ENIs into the chosen
subnets that persist across API create/destroy cycles.

Design notes:
- This is the API Gateway *v2* VPC link used by HTTP APIs. REST APIs (v1)
  use a different, NLB-only VPC link resource and are a separate surface.
- Both subnet_ids and security_group_ids are immutable after creation
  (AWS has no update API for them); changing either replaces the link.
  Only the name can change in place.
- The security groups govern what the link's ENIs can reach -- they must
  allow egress to the target ALB/NLB listener ports.

Credentials, region, and deployment workflow live outside this spec in
IaC inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsHttpApiVpcLink
metadata:
  name: test-vpc-link
  org: test-org
  env: dev
  id: test-vpc-link-dev
spec:
  region: us-west-2
  subnetIds:
    - value: subnet-0abc123def456
    - value: subnet-0def456abc789
  securityGroupIds:
    - value: sg-0abc123def456
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.subnetIds` | `[]string \| valueFrom` | yes |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.securityGroupIds` | `[]string \| valueFrom` |  |  | AwsSecurityGroup (`status.outputs.security_group_id`) |

## Field Details

### spec.region

`string` · required

The AWS region where the VPC link will be created.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.subnetIds

`[]string | valueFrom` · required

Subnets in which AWS provisions the VPC link's network interfaces.
Immutable after creation (changing the set replaces the link). Spread
the link across at least two availability zones for high availability --
the link can only reach targets in AZs it has an ENI in. Accepts direct
subnet IDs or references to AwsSubnet resources.

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: {"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.securityGroupIds

`[]string | valueFrom`

Security groups applied to the VPC link's network interfaces. Immutable
after creation (changing the set replaces the link). These groups control
what the link can reach inside the VPC: allow egress to the private
ALB/NLB listener ports (and make the target's security group admit
ingress from these groups). When omitted, the link is created without
security groups and AWS applies no filtering on the link side.
Accepts direct security group IDs or references to AwsSecurityGroup
resources.

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsHttpApiVpcLink, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.vpc_link_id` | `string` | The VPC link ID (e.g. "abc123"). This is the identifier private integrations set as their connection_id. |
| `status.outputs.vpc_link_arn` | `string` | The ARN of the VPC link. Useful for IAM policies and tag-based governance. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.subnetIds` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.securityGroupIds` | AwsSecurityGroup | `status.outputs.security_group_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsHttpApiGateway | `spec.routes[].integration.connectionId` | `status.outputs.vpc_link_id` |

## See Also

- [Overview](../README.md)

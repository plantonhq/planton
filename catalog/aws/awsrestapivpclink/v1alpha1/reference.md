# AwsRestApiVpcLink

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsRestApiVpcLinkSpec defines the desired configuration for an AWS
API Gateway VPC link (the API Gateway v1 form).

A REST API VPC link lets REST API integrations reach private
services: the link fronts a Network Load Balancer in your VPC, and
integrations with connection_type VPC_LINK route through it instead
of the public internet. One link is shared by many APIs and owns its
own network attachment - which is why it is its own component rather
than part of AwsRestApiGateway.

(HTTP APIs use a different link resource that attaches to subnets
directly - that is the AwsHttpApiVpcLink component. The two are not
interchangeable.)

Provisioning takes several minutes while AWS builds the network
attachment; creating the link is free - standard NLB charges apply
to the balancer it fronts.

## Example

```yaml
# Canonical AwsRestApiVpcLink example (hack/dev manifest and refgen
# Example source): a REST API VPC link fronting an internal Network
# Load Balancer. Literal ARN stands in for the composed NLB reference
# so the offline `tofu plan` renders the resource.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsRestApiVpcLink
metadata:
  name: orders-nlb-link
  id: orders-nlb-link
  org: test-org
  env: dev
spec:
  region: us-west-2
  description: Orders service NLB
  targetArn:
    value: arn:aws:elasticloadbalancing:us-west-2:123456789012:loadbalancer/net/orders/abcdef
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.targetArn` | `string \| valueFrom` | yes |  | AwsNlb (`status.outputs.load_balancer_arn`) |

## Field Details

### spec.region

`string` · required

The AWS region where the VPC link will be created.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.description

`string`

What this link reaches (e.g. "orders service NLB").

- rule: {"string":{"maxLen":"1024"}}

### spec.targetArn

`string | valueFrom` · required

The internal Network Load Balancer the link fronts. AWS accepts
exactly one balancer per link, and it cannot be changed after
creation - a different NLB means a new link.

- references: AwsNlb (`status.outputs.load_balancer_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsNlb, name: <that resource's name>, fieldPath: status.outputs.load_balancer_arn}} -- a bare string does not parse

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsRestApiVpcLink, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.vpc_link_id` | `string` | The VPC link ID (what integrations set as their connection). |
| `status.outputs.vpc_link_arn` | `string` | The VPC link ARN. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.targetArn` | AwsNlb | `status.outputs.load_balancer_arn` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsRestApiGateway | `spec.routes[].integration.vpcLinkId` | `status.outputs.vpc_link_id` |

## See Also

- [Overview](../README.md)

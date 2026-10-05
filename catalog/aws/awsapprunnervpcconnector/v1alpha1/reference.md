# AwsAppRunnerVpcConnector

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsAppRunnerVpcConnectorSpec defines the desired configuration for an AWS
App Runner VPC connector -- the managed network attachment that lets App
Runner services reach private resources inside a VPC (databases, caches,
internal APIs) for their OUTBOUND traffic.

A VPC connector is deliberately its own resource rather than a field on the
service: one connector is shared by any number of App Runner services (each
service references the connector by ARN in its egress configuration), and
the connector owns its own network-attachment lifecycle -- AWS provisions
managed ENIs into the chosen subnets that persist across service
create/destroy cycles.

Design notes:
- The connector governs EGRESS only. Inbound traffic to a private App
  Runner service travels through a separate VPC Ingress Connection
  resource, a different surface.
- Every attribute is immutable after creation (AWS has no update API for
  connectors): changing subnets or security groups replaces the connector,
  which AWS models as a new connector revision under the same name.
- The security groups govern what the connector's ENIs can reach -- they
  must allow egress to the target resources' ports, and the targets'
  security groups must admit ingress from these groups.

Credentials, region, and deployment workflow live outside this spec in
IaC inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsAppRunnerVpcConnector
metadata:
  name: test-vpc-connector
  org: test-org
  env: dev
  id: test-vpc-connector-dev
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
| `spec.securityGroupIds` | `[]string \| valueFrom` | yes |  | AwsSecurityGroup (`status.outputs.security_group_id`) |

## Field Details

### spec.region

`string` · required

The AWS region where the VPC connector will be created.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.subnetIds

`[]string | valueFrom` · required

Subnets in which AWS provisions the connector's network interfaces.
Immutable after creation (changing the set replaces the connector).
Provide subnets in at least two availability zones for high
availability -- App Runner routes egress only through AZs the connector
has an ENI in. All subnets must belong to the same VPC. Accepts direct
subnet IDs or references to AwsSubnet resources.

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: {"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.securityGroupIds

`[]string | valueFrom` · required

Security groups applied to the connector's network interfaces.
Immutable after creation (changing the set replaces the connector).
These groups control what VPC resources the connected services can
reach: allow egress to your databases' and caches' ports, and make the
targets' security groups admit ingress from these groups. AWS requires
at least one group on a connector. Accepts direct security group IDs or
references to AwsSecurityGroup resources.

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: {"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsAppRunnerVpcConnector, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.vpc_connector_arn` | `string` | The ARN of the VPC connector (e.g. "arn:aws:apprunner:us-west-2: 123456789012:vpcconnector/my-connector/1/abc123"). This is the identifier App Runner services set as their egress vpc_connector_arn. |
| `status.outputs.vpc_connector_revision` | `int64` | The revision of this connector. AWS creates a new revision (same name, incremented revision) whenever a connector with the same name is recreated with different subnets or security groups. |
| `status.outputs.status` | `string` | The connector's lifecycle status at the end of the deployment ("ACTIVE" when ready for services to attach). |

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
| AwsAppRunnerService | `spec.vpcConnectorArn` | `status.outputs.vpc_connector_arn` |

## See Also

- [Overview](../README.md)

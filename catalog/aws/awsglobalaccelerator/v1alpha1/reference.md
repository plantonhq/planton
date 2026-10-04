# AwsGlobalAccelerator

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsGlobalAcceleratorSpec defines the desired configuration for an AWS Global
Accelerator — a networking service that routes traffic through the AWS global
network to optimal endpoints based on health, geography, and routing policies.

Global Accelerator provides two static anycast IPv4 addresses (or dual-stack
IPv4 + IPv6) that serve as fixed entry points to your application. Traffic
enters the AWS network at the nearest edge location and is routed to healthy
endpoints across one or more AWS regions, providing improved availability and
performance for global users.

The resource hierarchy is: Accelerator → Listeners → Endpoint Groups → Endpoints.
All three levels are bundled because an accelerator without listeners and
endpoint groups is functionally useless (just static IPs doing nothing), and
nothing else in the resource graph references a listener or endpoint group
independently. Each listener and endpoint group still materializes as its own
provider resource, keyed by its name, so in-place updates stay surgical.

Key characteristics:
- Static anycast IPs: Two globally-advertised IP addresses that never change
- Automatic failover: Unhealthy endpoints are bypassed within seconds
- Multi-region: Endpoint groups in different regions for geographic distribution
- Protocol-agnostic: TCP and UDP at Layer 4 (not HTTP-aware like CloudFront)
- Client affinity: Optional SOURCE_IP stickiness per listener

This kind covers standard accelerators. Custom routing accelerators
(deterministic port-based routing to specific VPC subnet destinations) are a
distinct AWS resource family with its own listener and endpoint-group shapes
and are deliberately not modeled here.

Lifecycle notes the module handles for you: an accelerator must be disabled
before AWS allows deletion (the provider disables it and waits for the change
to deploy before deleting), and every accelerator/listener/endpoint-group
change is followed by a wait for the accelerator to return to the DEPLOYED
state — expect minutes, not seconds, per apply.

Credentials and deployment workflow live outside this spec in IaC inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsGlobalAccelerator
metadata:
  name: test-ga
spec:
  region: us-west-2
  enabled: true
  ipAddressType: IPV4
  listeners:
    - name: tcp-443
      protocol: TCP
      clientAffinity: NONE
      portRanges:
        - fromPort: 443
          toPort: 443
      endpointGroups:
        - name: us-east-1
          endpointGroupRegion: us-east-1
          healthCheckProtocol: TCP
          healthCheckIntervalSeconds: 30
          thresholdCount: 3
          trafficDialPercentage: 100.0
          endpoints:
            - endpointId:
                value: arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/my-alb/1234567890abcdef
              weight: 128
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.enabled` | `bool` |  | `true` |  |
| `spec.ipAddressType` | `string` |  | `IPV4` |  |
| `spec.ipAddresses` | `[]string` |  |  |  |
| `spec.flowLogs` | `AwsGlobalAcceleratorFlowLogs` |  |  |  |
| `spec.flowLogs.enabled` | `bool` |  |  |  |
| `spec.flowLogs.s3Bucket` | `string \| valueFrom` |  |  | AwsS3Bucket (`status.outputs.bucket_id`) |
| `spec.flowLogs.s3Prefix` | `string` |  |  |  |
| `spec.listeners` | `[]AwsGlobalAcceleratorListener` | yes |  |  |
| `spec.listeners[].name` | `string` | yes |  |  |
| `spec.listeners[].protocol` | `string` | yes |  |  |
| `spec.listeners[].clientAffinity` | `string` |  | `NONE` |  |
| `spec.listeners[].portRanges` | `[]AwsGlobalAcceleratorPortRange` | yes |  |  |
| `spec.listeners[].portRanges[].fromPort` | `int32` | yes |  |  |
| `spec.listeners[].portRanges[].toPort` | `int32` | yes |  |  |
| `spec.listeners[].endpointGroups` | `[]AwsGlobalAcceleratorEndpointGroup` | yes |  |  |
| `spec.listeners[].endpointGroups[].name` | `string` | yes |  |  |
| `spec.listeners[].endpointGroups[].endpointGroupRegion` | `string` |  |  |  |
| `spec.listeners[].endpointGroups[].healthCheckPort` | `int32` |  |  |  |
| `spec.listeners[].endpointGroups[].healthCheckProtocol` | `string` |  | `TCP` |  |
| `spec.listeners[].endpointGroups[].healthCheckPath` | `string` |  |  |  |
| `spec.listeners[].endpointGroups[].healthCheckIntervalSeconds` | `int32` |  | `30` |  |
| `spec.listeners[].endpointGroups[].thresholdCount` | `int32` |  | `3` |  |
| `spec.listeners[].endpointGroups[].trafficDialPercentage` | `double` |  | `100.0` |  |
| `spec.listeners[].endpointGroups[].endpoints` | `[]AwsGlobalAcceleratorEndpoint` |  |  |  |
| `spec.listeners[].endpointGroups[].endpoints[].endpointId` | `string \| valueFrom` | yes |  |  |
| `spec.listeners[].endpointGroups[].endpoints[].weight` | `int32` |  |  |  |
| `spec.listeners[].endpointGroups[].endpoints[].clientIpPreservationEnabled` | `bool` |  |  |  |
| `spec.listeners[].endpointGroups[].endpoints[].attachmentArn` | `string` |  |  |  |
| `spec.listeners[].endpointGroups[].portOverrides` | `[]AwsGlobalAcceleratorPortOverride` |  |  |  |
| `spec.listeners[].endpointGroups[].portOverrides[].listenerPort` | `int32` | yes |  |  |
| `spec.listeners[].endpointGroups[].portOverrides[].endpointPort` | `int32` | yes |  |  |

## Field Details

### spec.region

`string` · required

The AWS provider region used for deployment. Global Accelerator is a
GLOBAL service — its control-plane API is homed in us-west-2, and the
provider transparently pins API calls there regardless of this value.
This region still matters in one place: it is the default
endpoint_group_region for any endpoint group that does not set one, so
point it at the region where your primary endpoints live.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.enabled

`bool` · optional (explicit presence)

Whether the accelerator is enabled and accepting traffic. When disabled,
the accelerator's DNS name stops resolving and no traffic is routed.
Useful for temporarily disabling an accelerator during maintenance without
destroying it (the static IPs are retained while disabled).

- default: `true`

### spec.ipAddressType

`string` · optional (explicit presence)

IP address type for the accelerator.
- "IPV4": Two static IPv4 anycast addresses (default).
- "DUAL_STACK": IPv4 + IPv6 anycast addresses for clients on IPv6 networks.
  Dual-stack accelerators additionally export a dual-stack DNS name.

- default: `IPV4`

### spec.ipAddresses

`[]string`

Bring-Your-Own-IP (BYOIP) addresses to assign to the accelerator instead
of AWS-allocated anycast IPs. Provide exactly 1 or 2 IPv4 addresses from
a BYOIP address pool registered with AWS (maximum 2 — AWS hard limit).

ForceNew — changing this destroys and recreates the accelerator.
Leave empty to use AWS-allocated IPs (the default for most deployments).

- rule: {"repeated":{"maxItems":"2"}}

### spec.flowLogs

`AwsGlobalAcceleratorFlowLogs`

Optional flow log configuration for traffic analysis. When enabled,
Global Accelerator publishes flow logs to the specified S3 bucket.

Provider quirk (handled by the modules): flow-log settings ride a separate
accelerator-attributes API call after create, and changing the bucket or
prefix while flow logs are enabled requires AWS to briefly disable and
re-enable them — expect two deployment waits for that class of update.

- rule: s3_bucket is required when flow logs are enabled

### spec.flowLogs.enabled

`bool`

Enable flow log delivery. Setting this to false on an accelerator that
previously had flow logs enabled turns them off (the modules always send
the explicit disabled state — silence would leave AWS logging forever).

### spec.flowLogs.s3Bucket

`string | valueFrom`

S3 bucket name for flow log storage. Required when enabled is true.
The bucket must exist and grant Global Accelerator write permission.

The presence coupling below checks only that the reference is supplied;
whether it carries a literal value or resolves through valueFrom is
validated by the reference resolver (message-level CEL cannot dereference
StringValueOrRef sub-fields — a protovalidate-java constraint).

- references: AwsS3Bucket (`status.outputs.bucket_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_id}} -- a bare string does not parse

### spec.flowLogs.s3Prefix

`string`

S3 key prefix for flow logs. Useful for organizing logs when multiple
accelerators share a bucket. Example: "ga-logs/prod-accelerator/".
Maximum 255 characters (the provider's bound).

- rule: {"string":{"maxLen":"255"}}

### spec.listeners

`[]AwsGlobalAcceleratorListener` · required

Listeners define the ports and protocols the accelerator accepts traffic on.
Each listener routes traffic to one or more regional endpoint groups.

At least one listener is required — an accelerator without listeners
serves no purpose beyond reserving static IPs.

- rule: {"required":true,"repeated":{"minItems":"1"}}
- rule: protocol must be 'TCP' or 'UDP'
- rule: client_affinity must be 'NONE' or 'SOURCE_IP'

### spec.listeners[].name

`string` · required

User-assigned name for this listener. Used as a key in the output maps
(listener_arns, endpoint_group_arns) so downstream resources can reference
specific listener or endpoint group ARNs via valueFrom. This name is a
Planton-side key — AWS listeners have no name of their own.

Must be unique within the accelerator's listeners. Lowercase alphanumeric
and hyphens only, starting with a letter (max 63 characters).

- rule: name must start with a lowercase letter and contain only lowercase letters, numbers, and hyphens (max 63 chars)
- rule: {"required":true}

### spec.listeners[].protocol

`string` · required

Protocol for this listener. Global Accelerator operates at Layer 4.
- "TCP": For HTTP, HTTPS, WebSocket, gRPC, and other TCP workloads.
- "UDP": For DNS, gaming, IoT, and real-time media workloads.

- rule: {"required":true}

### spec.listeners[].clientAffinity

`string` · optional (explicit presence)

Client affinity setting for this listener.
- "NONE" (default): Requests from the same client may be routed to
  different endpoints. Best for stateless workloads.
- "SOURCE_IP": All requests from the same source IP address are routed
  to the same endpoint within an endpoint group. Required for stateful
  protocols (gaming, WebSocket connections, long-lived TCP sessions).

- default: `NONE`

### spec.listeners[].portRanges

`[]AwsGlobalAcceleratorPortRange` · required

Port ranges that this listener accepts traffic on. Each range defines
a from_port and to_port (inclusive). Use a single port range for most
workloads, or multiple ranges for services on different ports.

At least one port range is required. Maximum 10 ranges per listener
(AWS hard limit).

- rule: {"required":true,"repeated":{"minItems":"1","maxItems":"10"}}
- rule: to_port must be greater than or equal to from_port

### spec.listeners[].portRanges[].fromPort

`int32` · required

First port in the range (inclusive). Range: 1-65535.

- rule: {"required":true,"int32":{"lte":65535,"gte":1}}

### spec.listeners[].portRanges[].toPort

`int32` · required

Last port in the range (inclusive). Must be >= from_port. For a single
port, set from_port and to_port to the same value. Range: 1-65535.

- rule: {"required":true,"int32":{"lte":65535,"gte":1}}

### spec.listeners[].endpointGroups

`[]AwsGlobalAcceleratorEndpointGroup` · required

Endpoint groups define regional destinations for this listener's traffic.
Each endpoint group represents a set of endpoints in one AWS region.

At least one endpoint group is required — a listener without endpoint
groups drops all traffic.

- rule: {"required":true,"repeated":{"minItems":"1"}}
- rule: health_check_protocol must be 'TCP', 'HTTP', or 'HTTPS'
- rule: health_check_path is required when health_check_protocol is 'HTTP' or 'HTTPS'

### spec.listeners[].endpointGroups[].name

`string` · required

User-assigned name for this endpoint group. Used as part of the composite
key in the endpoint_group_arns output map (format: "listener_name/group_name").
This name is a Planton-side key — AWS endpoint groups have no name of
their own.

Must be unique within the parent listener's endpoint groups. Lowercase
alphanumeric and hyphens only, starting with a letter (max 63 characters).

- rule: name must start with a lowercase letter and contain only lowercase letters, numbers, and hyphens (max 63 chars)
- rule: {"required":true}

### spec.listeners[].endpointGroups[].endpointGroupRegion

`string`

AWS region for this endpoint group (e.g., "us-east-1", "eu-west-1").
When omitted, defaults to the spec's region. ForceNew — changing
the region requires replacing the endpoint group. The format is checked
here; region-name membership is validated by AWS (the region list grows).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[a-z]{2}(-[a-z]+)+-[0-9]$"}}

### spec.listeners[].endpointGroups[].healthCheckPort

`int32` · optional (explicit presence)

Port to use for health checks. When omitted, AWS uses the first port of
the listener's port ranges. Set this to check health on a dedicated
health-check port separate from the traffic port. Range: 1-65535.

- rule: {"int32":{"lte":65535,"gte":1}}

### spec.listeners[].endpointGroups[].healthCheckProtocol

`string` · optional (explicit presence)

Protocol for health checks.
- "TCP" (default): Verifies port reachability only.
- "HTTP": Sends GET request to health_check_path, expects 200 response.
- "HTTPS": Same as HTTP but over TLS.

- default: `TCP`

### spec.listeners[].endpointGroups[].healthCheckPath

`string`

Path for HTTP/HTTPS health checks. The accelerator sends a GET request
to this path (AWS defaults to "/" when omitted). Required here when
health_check_protocol is HTTP or HTTPS so intent is explicit. Ignored
for TCP health checks. Max 255 characters. Example: "/health".

- rule: {"string":{"maxLen":"255"}}

### spec.listeners[].endpointGroups[].healthCheckIntervalSeconds

`int32` · optional (explicit presence)

Seconds between health checks for each endpoint. AWS accepts exactly
two values: 10 or 30. Default: 30. (The Terraform provider's schema
validates the looser 10-30 range, but the Global Accelerator API
rejects anything except 10 or 30 at create time — this rule carries
the service's real contract so misconfigurations fail at validation,
not at deploy.)

- default: `30`
- rule: {"int32":{"in":[10,30]}}

### spec.listeners[].endpointGroups[].thresholdCount

`int32` · optional (explicit presence)

Number of consecutive health checks that must succeed (or fail) to
change an endpoint's health status. Range: 1-10. Default: 3.

- default: `3`
- rule: {"int32":{"lte":10,"gte":1}}

### spec.listeners[].endpointGroups[].trafficDialPercentage

`double` · optional (explicit presence)

Percentage of traffic to route to this endpoint group. Range: 0.0-100.0.
When omitted, AWS routes all traffic (100.0). Use values below 100 for
gradual traffic shifting between regions (blue/green, canary deployments).

Set to 0 explicitly to temporarily drain a region without removing its
endpoints — 0 is a real value here, distinct from omitting the field.

- default: `100.0`
- rule: {"double":{"lte":100,"gte":0}}

### spec.listeners[].endpointGroups[].endpoints

`[]AwsGlobalAcceleratorEndpoint`

Endpoints within this regional group. Each endpoint is a resource that
receives traffic — an ALB, NLB, Elastic IP, or EC2 instance.

Endpoints are optional. You may create the endpoint group first and
register endpoints later (e.g., when the ALB or NLB is deployed).

### spec.listeners[].endpointGroups[].endpoints[].endpointId

`string | valueFrom` · required

Resource identifier for the endpoint. Accepts:
- ALB ARN: "arn:aws:elasticloadbalancing:..."
- NLB ARN: "arn:aws:elasticloadbalancing:..."
- EIP allocation ID: "eipalloc-..."
- EC2 instance ID: "i-..."

Uses StringValueOrRef for cross-resource referencing (e.g., reference an
AwsAlb's load_balancer_arn output or an AwsElasticIp's allocation_id
output via valueFrom). No default_kind is set because the target resource
type varies.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.listeners[].endpointGroups[].endpoints[].weight

`int32` · optional (explicit presence)

Relative weight for this endpoint. Range: 0-255. When omitted, both
modules materialize AWS's documented default of 128 — the provider has no
default of its own and would otherwise transmit 0, silently draining the
endpoint. Higher weight means more traffic.

Set to 0 explicitly to temporarily stop routing traffic to this endpoint
without removing it — 0 is a real value here, distinct from omitting the
field.

- rule: {"int32":{"lte":255,"gte":0}}

### spec.listeners[].endpointGroups[].endpoints[].clientIpPreservationEnabled

`bool` · optional (explicit presence)

Preserve the client's source IP address in requests forwarded to the
endpoint. Applies to Application Load Balancer and EC2 instance
endpoints; when omitted, AWS applies its per-endpoint-type default
(false for ALB endpoints).

Operational note: when enabled, Global Accelerator creates a security
group named "GlobalAccelerator" in the endpoint's VPC that must be
deleted before that VPC can be destroyed.

### spec.listeners[].endpointGroups[].endpoints[].attachmentArn

`string`

ARN of a Global Accelerator cross-account attachment that authorizes
this endpoint when it lives in another AWS account. Create the
attachment in the endpoint-owning account (with this accelerator's
account as a principal) and supply its ARN here. Leave empty for
same-account endpoints — the common case.

- rule: attachment_arn must be a Global Accelerator attachment ARN (arn:<partition>:globalaccelerator::<account>:attachment/<id>)

### spec.listeners[].endpointGroups[].portOverrides

`[]AwsGlobalAcceleratorPortOverride`

Port overrides remap listener ports to different endpoint ports. Useful
when the listener accepts traffic on one port but the endpoint serves
on a different port. Maximum 10 overrides per endpoint group (AWS hard
limit).

Example: listener on port 443, endpoint on port 8443.

- rule: {"repeated":{"maxItems":"10"}}

### spec.listeners[].endpointGroups[].portOverrides[].listenerPort

`int32` · required

The listener port that is remapped. Must match a port within one of the
listener's port ranges.

- rule: {"required":true,"int32":{"lte":65535,"gte":1}}

### spec.listeners[].endpointGroups[].portOverrides[].endpointPort

`int32` · required

The endpoint port that traffic is forwarded to.

- rule: {"required":true,"int32":{"lte":65535,"gte":1}}

## Validation Rules

- `ip_address_type_valid`: ip_address_type must be 'IPV4' or 'DUAL_STACK'

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsGlobalAccelerator, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.accelerator_arn` | `string` | The Amazon Resource Name of the Global Accelerator. Used in IAM policies and cross-service permissions. |
| `status.outputs.accelerator_dns_name` | `string` | The DNS name assigned to the accelerator (e.g., "a1234567890abcdef.awsglobalaccelerator.com"). Clients connect to this DNS name to reach the accelerator. Create Route53 alias records pointing custom domains here. |
| `status.outputs.accelerator_dual_stack_dns_name` | `string` | The dual-stack DNS name assigned to the accelerator. Only populated when ip_address_type is DUAL_STACK. Supports both IPv4 and IPv6 clients. |
| `status.outputs.accelerator_hosted_zone_id` | `string` | The Route53 hosted zone ID for the accelerator's DNS name. Always Z2BJ6XQ5FK7U4H for Global Accelerator. Required when creating Route53 alias records that point to this accelerator. |
| `status.outputs.accelerator_ip_addresses` | `[]string` | The static anycast IP addresses assigned to the accelerator. Typically two IPv4 addresses. These never change for the lifetime of the accelerator (unless using BYOIP and the accelerator is recreated). |
| `status.outputs.listener_arns` | `map<string, string>` | Map of listener name to listener ARN. The keys correspond to the name field of each entry in spec.listeners. Downstream resources can reference specific listener ARNs via valueFrom using status.outputs.listener_arns.{name}. |
| `status.outputs.endpoint_group_arns` | `map<string, string>` | Map of composite key to endpoint group ARN. Keys use the format "listener_name/group_name" to uniquely identify each endpoint group across the accelerator. Downstream resources can reference specific endpoint group ARNs via valueFrom. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.flowLogs.s3Bucket` | AwsS3Bucket | `status.outputs.bucket_id` |

## See Also

- [Overview](../README.md)

# AwsLbTargetGroup

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsLbTargetGroupSpec defines an ELBv2 target group: the routing destination
that listeners and listener rules forward traffic to.

A target group is the composition point of AWS load balancing. It has its own
lifecycle (it exists independently of any load balancer), its own ARN, and it
is referenced from many places at once: listener default actions, listener
rule forward actions, ECS services (as their deployment target), and
auto-scaling groups (which register instances into it). One target group can
receive traffic from several listeners, and one listener can spread traffic
across several weighted target groups -- which is why it is a first-class
resource rather than a detail of the load balancer.

The same kind serves both Application and Network Load Balancers, exactly as
AWS models it: the protocol decides the family (HTTP/HTTPS for ALB;
TCP/UDP/TCP_UDP/TLS/QUIC/TCP_QUIC for NLB), and the family decides which
tuning fields apply. Field comments call out the scope of every
family-specific field.
Gateway Load Balancer (GENEVE) target groups are deliberately not modeled --
there is no gateway load balancer kind to compose them with.

The AWS name is target_group_name when set, else metadata.name, which both
IaC modules truncate deterministically to AWS's 32-character limit.
Name, port, protocol, protocol_version, vpc_id, target_type, and
ip_address_type are create-only in AWS: changing any of them replaces the
target group (and the IaC engine re-creates dependent listener attachments).

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsLbTargetGroup
metadata:
  name: api-demo
spec:
  region: us-west-2
  vpcId:
    value: vpc-0123456789abcdef0
  # The ECS/EKS shape: IP targets registered dynamically by the orchestrator,
  # with an HTTP health check on a dedicated readiness path. Exercises the
  # nested health-check and stickiness objects so the fixture proves the full
  # variable contract, not just the scalars.
  targetType: ip
  port: 8080
  protocol: HTTP
  protocolVersion: HTTP1
  deregistrationDelaySeconds: 60
  healthCheck:
    protocol: HTTP
    path: /healthz
    healthyThreshold: 3
    unhealthyThreshold: 3
    intervalSeconds: 15
    timeoutSeconds: 5
    matcher: "200-299"
  stickiness:
    type: lb_cookie
    cookieDurationSeconds: 3600
  # ALB Target Optimizer: the agent port on each target. Create-time config
  # on a group that never registers targets here -- the QUIC-family arms
  # (protocol QUIC/TCP_QUIC, flow-hash stickiness, quic_server_id) live in
  # the 04-quic-passthrough preset because protocol is one create-only field
  # and this fixture proves the ALB family.
  targetControlPort: 9999
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.targetGroupName` | `string` |  |  |  |
| `spec.vpcId` | `string \| valueFrom` |  |  | AwsVpc (`status.outputs.vpc_id`) |
| `spec.targetType` | `string` |  | `instance` |  |
| `spec.port` | `int32` |  |  |  |
| `spec.protocol` | `string` |  |  |  |
| `spec.protocolVersion` | `string` |  |  |  |
| `spec.ipAddressType` | `string` |  |  |  |
| `spec.healthCheck` | `AwsLbTargetGroupHealthCheck` |  |  |  |
| `spec.healthCheck.enabled` | `bool` |  |  |  |
| `spec.healthCheck.protocol` | `string` |  |  |  |
| `spec.healthCheck.port` | `string` |  |  |  |
| `spec.healthCheck.path` | `string` |  |  |  |
| `spec.healthCheck.healthyThreshold` | `int32` |  |  |  |
| `spec.healthCheck.unhealthyThreshold` | `int32` |  |  |  |
| `spec.healthCheck.intervalSeconds` | `int32` |  |  |  |
| `spec.healthCheck.timeoutSeconds` | `int32` |  |  |  |
| `spec.healthCheck.matcher` | `string` |  |  |  |
| `spec.stickiness` | `AwsLbTargetGroupStickiness` |  |  |  |
| `spec.stickiness.type` | `string` | yes |  |  |
| `spec.stickiness.enabled` | `bool` |  |  |  |
| `spec.stickiness.cookieDurationSeconds` | `int32` |  |  |  |
| `spec.stickiness.cookieName` | `string` |  |  |  |
| `spec.deregistrationDelaySeconds` | `int32` |  | `300` |  |
| `spec.slowStartSeconds` | `int32` |  |  |  |
| `spec.loadBalancingAlgorithmType` | `string` |  |  |  |
| `spec.loadBalancingAnomalyMitigation` | `string` |  |  |  |
| `spec.loadBalancingCrossZoneEnabled` | `string` |  |  |  |
| `spec.preserveClientIp` | `bool` |  |  |  |
| `spec.proxyProtocolV2` | `bool` |  |  |  |
| `spec.connectionTermination` | `bool` |  |  |  |
| `spec.lambdaMultiValueHeadersEnabled` | `bool` |  |  |  |
| `spec.targetGroupHealth` | `AwsLbTargetGroupHealthPolicy` |  |  |  |
| `spec.targetGroupHealth.dnsFailover` | `AwsLbTargetGroupDnsFailover` |  |  |  |
| `spec.targetGroupHealth.dnsFailover.minimumHealthyTargetsCount` | `string` |  |  |  |
| `spec.targetGroupHealth.dnsFailover.minimumHealthyTargetsPercentage` | `string` |  |  |  |
| `spec.targetGroupHealth.unhealthyStateRouting` | `AwsLbTargetGroupUnhealthyStateRouting` |  |  |  |
| `spec.targetGroupHealth.unhealthyStateRouting.minimumHealthyTargetsCount` | `int32` |  |  |  |
| `spec.targetGroupHealth.unhealthyStateRouting.minimumHealthyTargetsPercentage` | `string` |  |  |  |
| `spec.targetHealthState` | `AwsLbTargetGroupTargetHealthState` |  |  |  |
| `spec.targetHealthState.enableUnhealthyConnectionTermination` | `bool` |  |  |  |
| `spec.targetHealthState.unhealthyDrainingIntervalSeconds` | `int32` |  |  |  |
| `spec.targets` | `[]AwsLbTargetGroupTarget` |  |  |  |
| `spec.targets[].targetId` | `string \| valueFrom` | yes |  | AwsEc2Instance (`status.outputs.instance_id`) |
| `spec.targets[].port` | `int32` |  |  |  |
| `spec.targets[].availabilityZone` | `string` |  |  |  |
| `spec.targets[].quicServerId` | `string` |  |  |  |
| `spec.targetControlPort` | `int32` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the target group is created.
Must match the region of the VPC and of any load balancer that forwards
to this group. Example: "us-west-2", "eu-west-1".

- rule: {"string":{"minLen":"1"}}

### spec.targetGroupName

`string`

The explicit AWS-side target group name. Empty (the common case) means
the group is named after `metadata.name`, truncated to 32 characters. Set
it when the AWS name must differ from `metadata.name` -- for example a
target group taken over from AWS under a generated name, which keeps that
name while the component carries a readable one. AWS allows up to 32
alphanumeric characters and hyphens, not beginning or ending with a
hyphen. ForceNew: a name different from the deployed one replaces the
target group.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"32","pattern":"^[A-Za-z0-9]([A-Za-z0-9-]*[A-Za-z0-9])?$"}}

### spec.vpcId

`string | valueFrom`

The VPC the targets live in. Required for "instance", "ip", and "alb"
target types; ignored for "lambda" (a Lambda function is not addressed
through a VPC). Immutable: changing the VPC replaces the target group.

Requiredness is enforced by the IaC modules rather than a proto rule:
message-level CEL cannot inspect StringValueOrRef fields without breaking
protovalidate-java, so both engines fail fast with a clear error when the
VPC is missing for a non-lambda target type.

- references: AwsVpc (`status.outputs.vpc_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsVpc, name: <that resource's name>, fieldPath: status.outputs.vpc_id}} -- a bare string does not parse

### spec.targetType

`string`

How targets are registered into this group. Immutable.
- "instance" (default): targets are EC2 instance IDs; NLB health checks
  use the instance's primary private IP.
- "ip": targets are IP addresses from the VPC CIDR or peered/on-premises
  ranges routable through the VPC. The type ECS awsvpc tasks and most
  Kubernetes pod-IP integrations use.
- "lambda": the single target is a Lambda function; the load balancer
  invokes it directly, so port/protocol/vpc_id do not apply.
- "alb": the single target is an Application Load Balancer -- the
  NLB-in-front-of-ALB pattern that combines static Layer-4 IPs with
  Layer-7 routing.

- default: `instance`

### spec.port

`int32`

The port targets receive traffic on, 1-65535. Required for every target
type except "lambda". Individual target registrations can override it
per target. Immutable.

### spec.protocol

`string`

The protocol used between the load balancer and the targets. Decides the
load balancer family this group can attach to. Immutable.
- ALB: "HTTP", "HTTPS".
- NLB: "TCP", "UDP", "TCP_UDP", "TLS", "QUIC", "TCP_QUIC".
Required for every target type except "lambda". A TLS listener may
forward to TCP targets (the NLB terminates TLS); an HTTPS listener may
forward to HTTP targets (the ALB terminates TLS). "QUIC" carries QUIC
traffic natively; "TCP_QUIC" serves both TCP and QUIC on one group (the
HTTP/3 pattern where clients may fall back to TCP). QUIC targets are
registered with a quic_server_id (see targets).

### spec.protocolVersion

`string`

The application-layer protocol between an ALB and HTTP/HTTPS targets.
Valid values: "HTTP1" (default), "HTTP2", "GRPC". Immutable.
Choose "GRPC" for gRPC services (enables gRPC-native health-check
matchers) and "HTTP2" for end-to-end HTTP/2. Only meaningful when
protocol is HTTP or HTTPS.

### spec.ipAddressType

`string`

The address family of registered targets, for "instance" and "ip" target
types. Valid values: "ipv4" (default), "ipv6". Immutable. An IPv6 target
group requires a dualstack load balancer.

### spec.healthCheck

`AwsLbTargetGroupHealthCheck`

Health check configuration. When omitted, AWS applies protocol-appropriate
defaults (HTTP GET "/" for ALB protocols; TCP reachability on the traffic
port for NLB protocols).

- rule: protocol must be 'HTTP', 'HTTPS', or 'TCP' when set
- rule: path is not valid for TCP health checks
- rule: matcher is not valid for TCP health checks
- rule: healthy_threshold must be between 2 and 10 when set
- rule: unhealthy_threshold must be between 2 and 10 when set
- rule: interval_seconds must be between 5 and 300 when set
- rule: timeout_seconds must be between 2 and 120 when set
- rule: timeout_seconds must be smaller than interval_seconds

### spec.healthCheck.enabled

`bool` · optional (explicit presence)

Whether health checks run at all. AWS default: true. Health checks can
only be disabled for "lambda" target groups; every other type requires
them. Optional rather than plain bool so that false ("disable") is
distinguishable from unset ("keep the AWS default of true").

### spec.healthCheck.protocol

`string`

Protocol for the health check probe: "HTTP", "HTTPS", or "TCP".
AWS default: "HTTP" for ALB-protocol groups, "TCP" for NLB-protocol
groups. TCP health checks are not allowed when the traffic protocol is
HTTP/HTTPS. Not applicable to "lambda" targets (always an invocation).

### spec.healthCheck.port

`string`

Port for the health check probe: "traffic-port" (AWS default -- probe
whatever port each target receives traffic on) or a specific port number
as a string (e.g. "8081" for a dedicated health/admin port).

### spec.healthCheck.path

`string`

Destination path for HTTP/HTTPS probes. AWS default: "/". For GRPC
protocol_version, the path is a fully-qualified gRPC method name
(AWS default "/AWS.ALB/healthcheck"). Not valid for TCP probes.

### spec.healthCheck.healthyThreshold

`int32`

Consecutive successful probes before an unhealthy target is considered
healthy. Range 2-10. AWS default: 5 (ALB) / 3 (NLB).

### spec.healthCheck.unhealthyThreshold

`int32`

Consecutive failed probes before a healthy target is considered
unhealthy. Range 2-10. AWS default: 2 (ALB) / 3 (NLB).

### spec.healthCheck.intervalSeconds

`int32`

Seconds between probes of an individual target. Range 5-300.
AWS default: 30.

### spec.healthCheck.timeoutSeconds

`int32`

Seconds to wait for a probe response before counting it failed. Must be
smaller than interval_seconds. Range 2-120. AWS defaults vary by
protocol (HTTP: 5-6, TCP: 10).

### spec.healthCheck.matcher

`string`

Response codes that count as healthy, for HTTP/HTTPS probes.
HTTP matchers: a code ("200"), a range ("200-299"), or a list
("200,202"). AWS default: "200".
GRPC matchers (protocol_version = "GRPC"): gRPC status codes, e.g. "0"
or "0-99". AWS default: "12" -- gRPC UNIMPLEMENTED, so a bare health
stub counts as healthy; set "0" to require an OK response.
Not valid for TCP probes (a TCP probe has no response body to match).

### spec.stickiness

`AwsLbTargetGroupStickiness`

Session stickiness. ALB supports cookie-based stickiness ("lb_cookie",
"app_cookie"); NLB supports flow-hash stickiness ("source_ip",
"source_ip_dest_ip", "source_ip_dest_ip_proto"). When omitted,
stickiness is disabled.

- rule: type must be 'lb_cookie', 'app_cookie', 'source_ip', 'source_ip_dest_ip', or 'source_ip_dest_ip_proto'
- rule: cookie_name is required for 'app_cookie' and not valid for other types
- rule: cookie_duration_seconds only applies to 'lb_cookie' and 'app_cookie'
- rule: cookie_duration_seconds must be between 1 and 604800 when set

### spec.stickiness.type

`string` · required

The stickiness mechanism. Required.
- "lb_cookie" (ALB): the load balancer issues and manages its own
  cookie (AWSALB); duration-based.
- "app_cookie" (ALB): the application issues the cookie named in
  cookie_name; the load balancer follows it.
- "source_ip" (NLB): affinity by client source IP.
- "source_ip_dest_ip" (NLB): affinity by source and destination IP --
  for dualstack groups where one client may arrive on both families.
- "source_ip_dest_ip_proto" (NLB): affinity by source IP, destination
  IP, and protocol -- the narrowest flow-hash affinity.

- rule: {"required":true}

### spec.stickiness.enabled

`bool` · optional (explicit presence)

Whether stickiness is active. AWS default: true (configuring the block
implies enabling it). Optional rather than plain bool so that false
("configured but switched off") is distinguishable from unset.

### spec.stickiness.cookieDurationSeconds

`int32`

Seconds a "lb_cookie" or "app_cookie" association lasts before the
client is re-balanced. Range 1-604800 (7 days). AWS default: 86400
(1 day). Not applicable to "source_ip".

### spec.stickiness.cookieName

`string`

The application cookie the load balancer follows. Required for
"app_cookie", not valid otherwise. Must not begin with "AWSALB" (those
names are reserved for the load balancer's own cookies).

### spec.deregistrationDelaySeconds

`int32`

Seconds the load balancer waits before completing deregistration of a
draining target, letting in-flight requests finish. Range 0-3600.
AWS default: 300. Not supported for "lambda" targets.

- default: `300`

### spec.slowStartSeconds

`int32`

ALB only. Seconds during which a newly registered target receives a
linearly increasing share of traffic, letting caches warm before full
load arrives. 0 disables slow start (AWS default); otherwise 30-900.
Incompatible with the "least_outstanding_requests" algorithm and with
stickiness.

### spec.loadBalancingAlgorithmType

`string`

ALB only. How the load balancer picks a target for each request.
Valid values: "round_robin" (default), "least_outstanding_requests",
"weighted_random". "least_outstanding_requests" suits uneven request
costs; "weighted_random" is required for anomaly mitigation.

### spec.loadBalancingAnomalyMitigation

`string`

ALB only. Automatic anomaly mitigation: the load balancer detects targets
returning anomalous responses and reduces their traffic share. Valid
values: "on", "off" (default). Requires
load_balancing_algorithm_type = "weighted_random".

### spec.loadBalancingCrossZoneEnabled

`string`

Whether traffic may cross Availability Zones on its way to targets in
this group, overriding the load balancer's own cross-zone setting.
Valid values: "true", "false", "use_load_balancer_configuration"
(default). A string tri-state, mirroring the AWS API: the third value
means "inherit from the load balancer".

### spec.preserveClientIp

`bool` · optional (explicit presence)

NLB only. Whether targets see the original client IP in the IP header.
AWS defaults this per target type -- enabled for "instance" targets,
disabled for "ip" targets -- so leaving it unset keeps the AWS default
for whichever type is in use (the reason this field is optional rather
than a plain bool: false must be distinguishable from unset).

### spec.proxyProtocolV2

`bool`

NLB only. Send the Proxy Protocol v2 header on connections to targets,
carrying client connection metadata (source/destination address and
port, VPC endpoint ID). Targets must be configured to parse the header
-- enabling this against an unaware backend breaks the connection.

### spec.connectionTermination

`bool`

NLB only. Terminate connections to a deregistered target when the
deregistration delay expires instead of waiting for the client to close
them. Recommended for long-lived connections (WebSocket, gRPC streams,
database protocols) that would otherwise pin draining targets.

### spec.lambdaMultiValueHeadersEnabled

`bool`

Lambda targets only. Deliver multi-value HTTP headers and query
parameters to the function as arrays instead of last-value-wins strings.

### spec.targetGroupHealth

`AwsLbTargetGroupHealthPolicy`

Group-level health policy: DNS failover and unhealthy-state routing
thresholds that act on the group as a whole (ALB and NLB).

### spec.targetGroupHealth.dnsFailover

`AwsLbTargetGroupDnsFailover`

DNS failover: when healthy targets drop below the threshold, the load
balancer's DNS stops resolving to the affected Availability Zone (or to
the whole load balancer), shifting clients elsewhere.

### spec.targetGroupHealth.dnsFailover.minimumHealthyTargetsCount

`string`

Minimum number of healthy targets, as a string: a number (e.g. "2") or
"off" (AWS default) to disable the count criterion.

### spec.targetGroupHealth.dnsFailover.minimumHealthyTargetsPercentage

`string`

Minimum percentage of healthy targets, as a string: "1"-"100" or "off"
(AWS default) to disable the percentage criterion.

### spec.targetGroupHealth.unhealthyStateRouting

`AwsLbTargetGroupUnhealthyStateRouting`

Unhealthy-state routing: when healthy targets drop below the threshold,
the load balancer routes to ALL targets -- including unhealthy ones --
on the theory that a partially working target beats a rejected request
during a mass failure.

### spec.targetGroupHealth.unhealthyStateRouting.minimumHealthyTargetsCount

`int32`

Minimum number of healthy targets, 1-max. AWS default: 1.

### spec.targetGroupHealth.unhealthyStateRouting.minimumHealthyTargetsPercentage

`string`

Minimum percentage of healthy targets, as a string: "1"-"100" or "off"
(AWS default) to disable the percentage criterion.

### spec.targetHealthState

`AwsLbTargetGroupTargetHealthState`

NLB TCP/TLS only. What happens to established connections while a target
is in an unhealthy state.

- rule: unhealthy_draining_interval_seconds must be between 0 and 360000

### spec.targetHealthState.enableUnhealthyConnectionTermination

`bool`

When false, the NLB keeps established connections to a target that turns
unhealthy (AWS default behavior); when true, it terminates them. Keeping
connections suits long-lived sessions that may ride out a transient
health blip; terminating suits strict fail-fast backends.

### spec.targetHealthState.unhealthyDrainingIntervalSeconds

`int32`

Seconds an unhealthy target keeps draining established connections
before they are terminated. Only meaningful when
enable_unhealthy_connection_termination is false. Range 0-360000.
AWS default: 0.

### spec.targets

`[]AwsLbTargetGroupTarget`

Static target registrations managed with the group. Most architectures
leave this empty -- ECS services, auto-scaling groups, and Kubernetes
controllers register their own targets dynamically -- but standalone
EC2 instances, fixed IPs, a Lambda function, or an inner ALB are
registered here. Registrations are folded into this kind (not a separate
resource) because a registration is pure glue with no referenceable
identity of its own.

- rule: port must be between 1 and 65535 when set

### spec.targets[].targetId

`string | valueFrom` · required

What the target is, per the group's target_type:
- "instance": an EC2 instance ID (defaults to referencing an
  AwsEc2Instance's instance_id output).
- "ip": a literal IP address inside the VPC or a routable peered range.
- "lambda": the Lambda function ARN.
- "alb": the inner Application Load Balancer's ARN.

- references: AwsEc2Instance (`status.outputs.instance_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsEc2Instance, name: <that resource's name>, fieldPath: status.outputs.instance_id}} -- a bare string does not parse

### spec.targets[].port

`int32`

Overrides the group's port for this one target, 1-65535. Lets one group
spread traffic across heterogeneous ports (e.g. several containers on
one host).

### spec.targets[].availabilityZone

`string`

For "ip" targets outside the load balancer's VPC (peered VPC,
on-premises): the literal string "all". Leave unset for in-VPC targets.

### spec.targets[].quicServerId

`string`

QUIC / TCP_QUIC target groups only: the QUIC server ID this target
serves. QUIC routes established connections by connection ID rather
than by 5-tuple, and the server ID ties a registration to the QUIC
endpoint identity the target presents.

### spec.targetControlPort

`int32`

ALB only. The port the ALB Target Optimizer agent listens on, 1-65535.
Setting it enables Target Optimizer for this group: an agent on each
target reports its readiness for new requests over this port, and the
ALB routes accordingly. Requires the agent to be running on every
target -- enabling it without the agent marks targets unavailable.
Immutable: changing it replaces the target group.

## Validation Rules

- `target_type_valid`: target_type must be 'instance', 'ip', 'lambda', or 'alb' when set
- `protocol_valid`: protocol must be one of: HTTP, HTTPS, TCP, UDP, TCP_UDP, TLS, QUIC, TCP_QUIC
- `port_protocol_required_unless_lambda`: port and protocol are required unless target_type is 'lambda'
- `lambda_takes_no_port_or_protocol`: port, protocol, and protocol_version do not apply when target_type is 'lambda'
- `port_range`: port must be between 1 and 65535 when set
- `target_control_port_range`: target_control_port must be between 1 and 65535 when set
- `target_control_port_only_for_alb_protocols`: target_control_port only applies when protocol is 'HTTP' or 'HTTPS'
- `protocol_version_valid`: protocol_version must be 'HTTP1', 'HTTP2', or 'GRPC' when set
- `protocol_version_only_for_alb_protocols`: protocol_version only applies when protocol is 'HTTP' or 'HTTPS'
- `ip_address_type_valid`: ip_address_type must be 'ipv4' or 'ipv6' when set
- `deregistration_delay_range`: deregistration_delay_seconds must be between 0 and 3600
- `slow_start_range`: slow_start_seconds must be 0 (disabled) or between 30 and 900
- `slow_start_only_for_alb_protocols`: slow_start_seconds only applies when protocol is 'HTTP' or 'HTTPS'
- `slow_start_incompatible_with_stickiness`: slow_start_seconds cannot be combined with stickiness
- `slow_start_incompatible_with_least_outstanding_requests`: slow_start_seconds cannot be combined with the 'least_outstanding_requests' algorithm
- `load_balancing_algorithm_valid`: load_balancing_algorithm_type must be 'round_robin', 'least_outstanding_requests', or 'weighted_random' when set
- `anomaly_mitigation_valid`: load_balancing_anomaly_mitigation must be 'on' or 'off' when set
- `anomaly_mitigation_requires_weighted_random`: load_balancing_anomaly_mitigation = 'on' requires load_balancing_algorithm_type = 'weighted_random'
- `cross_zone_valid`: load_balancing_cross_zone_enabled must be 'true', 'false', or 'use_load_balancer_configuration' when set
- `proxy_protocol_v2_only_for_nlb_protocols`: proxy_protocol_v2 only applies when protocol is TCP, UDP, TCP_UDP, or TLS
- `connection_termination_only_for_nlb_protocols`: connection_termination only applies when protocol is TCP, UDP, TCP_UDP, or TLS
- `lambda_multi_value_headers_only_for_lambda`: lambda_multi_value_headers_enabled only applies when target_type is 'lambda'

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsLbTargetGroup, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.target_group_arn` | `string` | The ARN of the target group (e.g. "arn:aws:elasticloadbalancing: us-west-2:123456789012:targetgroup/api/50dc6c495c0c9188"). The primary handle other resources reference via status.outputs.target_group_arn -- listener forward actions, ECS service load-balancer wiring, and ASG attachments all take this value. |
| `status.outputs.target_group_name` | `string` | The friendly name of the target group (metadata.name, truncated to the 32-character AWS limit when necessary), for console URLs and CLI commands. |
| `status.outputs.arn_suffix` | `string` | The ARN suffix (e.g. "targetgroup/api/50dc6c495c0c9188") used as the TargetGroup dimension in CloudWatch metrics -- the handle alarms and dashboards need. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.vpcId` | AwsVpc | `status.outputs.vpc_id` |
| `spec.targets[].targetId` | AwsEc2Instance | `status.outputs.instance_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsAutoScalingGroup | `spec.targetGroups` | `status.outputs.target_group_arn` |
| AwsEcsService | `spec.loadBalancers[].targetGroupArn` | `status.outputs.target_group_arn` |
| AwsEcsService | `spec.loadBalancers[].advancedConfiguration.alternateTargetGroupArn` | `status.outputs.target_group_arn` |
| AwsEcsService | `spec.autoscaling.requestsPerTarget.targetGroupArnSuffix` | `status.outputs.arn_suffix` |
| AwsLbListener | `spec.defaultActions[].forward.targetGroups[].arn` | `status.outputs.target_group_arn` |
| AwsLbListenerRule | `spec.actions[].forward.targetGroups[].arn` | `status.outputs.target_group_arn` |

## See Also

- [Overview](../README.md)

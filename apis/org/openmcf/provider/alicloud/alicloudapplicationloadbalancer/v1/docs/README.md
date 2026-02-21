# AlicloudApplicationLoadBalancer Research Documentation

## Provider Resources

| OpenMCF Component | Terraform Resource | Pulumi Resource |
| --- | --- | --- |
| AlicloudApplicationLoadBalancer | `alicloud_alb_load_balancer` | `alb.LoadBalancer` |
| (bundled) Server Groups | `alicloud_alb_server_group` | `alb.ServerGroup` |
| (bundled) Listeners | `alicloud_alb_listener` | `alb.Listener` |

## Design Decisions

### Bundle Scope

ALB Rules (`alicloud_alb_rule`) are intentionally excluded from this bundle. ALB rules support 9 action types (ForwardGroup, Redirect, FixedResponse, Rewrite, InsertHeader, RemoveHeader, TrafficLimit, TrafficMirror, Cors) and 9 condition types (Host, Path, Header, QueryString, Method, Cookie, SourceIp, ResponseHeader, ResponseStatusCode). Including them would triple the proto spec size and produce YAML manifests that are unwieldy.

The listener's `default_actions` (ForwardGroup to a server group) covers the 80% use case. Advanced L7 routing can be managed separately.

### Billing

ALB only supports PayAsYouGo (pay-as-you-go) billing. The billing config is hardcoded in the IaC modules and not exposed in the spec.

### Server Group Membership

Server groups are created empty. Backend membership is managed externally by:
- ACK ingress controllers (for Kubernetes workloads)
- SAE bindings (for serverless applications)
- Manual attachment via console/API

This follows the same pattern as Azure LoadBalancer where "pool membership is managed outside this component."

### Zone Mappings

ALB requires a minimum of 2 zone mappings for high availability. Each zone mapping associates an availability zone with a VSwitch for IP allocation.

## Provider Field Coverage

### alicloud_alb_load_balancer

| Provider Field | Included | Notes |
| --- | --- | --- |
| load_balancer_name | Yes | |
| vpc_id | Yes | StringValueOrRef |
| address_type | Yes | Default: Internet |
| load_balancer_edition | Yes | Default: Standard |
| zone_mappings | Yes | Min 2 required |
| load_balancer_billing_config | Hardcoded | PayAsYouGo only |
| resource_group_id | Yes | Optional |
| access_log_config | Yes | Optional SLS integration |
| tags | Yes | |
| address_allocated_mode | Omitted | Advanced networking |
| address_ip_version | Omitted | IPv4 default sufficient |
| bandwidth_package_id | Omitted | Advanced networking |
| deletion_protection_config | Omitted | 80/20 scoping |
| modification_protection_config | Omitted | 80/20 scoping |
| ipv6_address_type | Omitted | IPv4 focus for v1 |

### alicloud_alb_server_group

| Provider Field | Included | Notes |
| --- | --- | --- |
| server_group_name | Yes | |
| vpc_id | Yes | From ALB spec |
| protocol | Yes | Default: HTTP |
| scheduler | Yes | Default: Wrr |
| health_check_config | Yes | All subfields |
| sticky_session_config | Yes | Optional |
| servers | Omitted | Managed externally |
| server_group_type | Omitted | Default Instance |
| connection_drain_config | Omitted | 80/20 scoping |
| slow_start_config | Omitted | 80/20 scoping |
| uch_config | Omitted | 80/20 scoping |

### alicloud_alb_listener

| Provider Field | Included | Notes |
| --- | --- | --- |
| load_balancer_id | Yes | From ALB resource |
| listener_port | Yes | |
| listener_protocol | Yes | HTTP/HTTPS/QUIC |
| default_actions | Yes | ForwardGroup |
| listener_description | Yes | Optional |
| certificates | Yes | For HTTPS |
| security_policy_id | Yes | For HTTPS |
| gzip_enabled | Yes | Default: true |
| http2_enabled | Yes | Default: true |
| idle_timeout | Yes | Default: 60 |
| request_timeout | Yes | Default: 60 |
| x_forwarded_for_config | Omitted | 80/20 scoping |
| access_log_tracing_config | Omitted | 80/20 scoping |
| quic_config | Omitted | 80/20 scoping |
| ca_certificates / ca_enabled | Omitted | Mutual TLS is niche |

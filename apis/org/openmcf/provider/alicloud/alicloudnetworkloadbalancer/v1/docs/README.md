# AliCloudNetworkLoadBalancer -- Research Documentation

## Provider Resource Mapping

| OpenMCF Concept | Terraform Resource | Pulumi Resource |
|---|---|---|
| NLB Load Balancer | `alicloud_nlb_load_balancer` | `nlb.LoadBalancer` |
| Server Group | `alicloud_nlb_server_group` | `nlb.ServerGroup` |
| Listener | `alicloud_nlb_listener` | `nlb.Listener` |

## Design Decisions

### Bundle Scope

The component bundles NLB + Server Groups + Listeners because an NLB without at least one server group and listener is non-functional. Server group server attachments are excluded -- they are managed externally (by ACK, manual, or automation).

### NLB vs ALB vs Classic SLB

Per DD03, Classic SLB is skipped. NLB handles L4 (TCP/UDP/TCPSSL), ALB handles L7 (HTTP/HTTPS/QUIC). They are sibling components with complementary roles.

### Billing

Payment type is hardcoded to `PayAsYouGo` in both IaC modules, matching the ALB convention. Subscription billing can be added via a spec field if there's demand.

### EIP Zone Binding

Unlike ALB, NLB zone mappings support an `allocation_id` field for binding a fixed EIP to each zone's NLB node. This is critical for users who need stable public IPs (e.g., database access, DNS A-records, firewall whitelisting).

### Connection Draining

NLB server groups support connection draining (10-900s timeout), which ALB server groups do not expose. This is essential for graceful backend rotation during deployments.

### Cross-Zone Load Balancing

NLB supports cross-zone load balancing (enabled by default). When disabled, traffic stays within the zone where it was received, which is useful for latency-sensitive workloads.

## Provider Field Coverage

### Load Balancer Fields

| Provider Field | Included | Reason |
|---|---|---|
| `load_balancer_name` | Yes | Optional, defaults to metadata.name |
| `vpc_id` | Yes | Required, FK to AliCloudVpc |
| `address_type` | Yes | Internet/Intranet |
| `load_balancer_type` | No | Hardcoded to "Network" |
| `payment_type` | No | Hardcoded to "PayAsYouGo" |
| `cross_zone_enabled` | Yes | Default true |
| `resource_group_id` | Yes | Optional per DD05 |
| `zone_mappings` | Yes | Required, min 2 |
| `zone_mappings.allocation_id` | Yes | Optional EIP binding |
| `tags` | Yes | User-defined + system tags |
| `address_ip_version` | No | IPv6 deferred to v2 |
| `bandwidth_package_id` | No | Advanced billing |
| `security_group_ids` | No | Separate attachment resource |
| `cps` | No | Advanced tuning |
| `deletion_protection_*` | No | Operational safeguard |
| `modification_protection_*` | No | Operational safeguard |

### Server Group Fields

| Provider Field | Included | Reason |
|---|---|---|
| `server_group_name` | Yes | Required |
| `vpc_id` | Yes | From spec.vpc_id |
| `protocol` | Yes | TCP/UDP/TCPSSL |
| `scheduler` | Yes | Wrr/Rr/Sch/Tch/Qch/Wlc |
| `connection_drain_enabled` | Yes | Graceful drain |
| `connection_drain_timeout` | Yes | 10-900s |
| `preserve_client_ip_enabled` | Yes | Default true |
| `health_check` | Yes | Full health check config |
| `any_port_enabled` | No | Edge case |
| `server_group_type` | No | Default Instance |
| `address_ip_version` | No | IPv4 default |

### Listener Fields

| Provider Field | Included | Reason |
|---|---|---|
| `listener_port` | Yes | Required |
| `listener_protocol` | Yes | TCP/UDP/TCPSSL |
| `server_group_id` | Yes | Via server_group_name reference |
| `listener_description` | Yes | Optional |
| `idle_timeout` | Yes | 1-900s, default 900 |
| `proxy_protocol_enabled` | Yes | Real client IP |
| `certificate_ids` | Yes | TCPSSL certs |
| `security_policy_id` | Yes | TCPSSL policy |
| `ca_certificate_ids` | Yes | Mutual TLS |
| `ca_enabled` | Yes | Mutual TLS toggle |
| `alpn_enabled/alpn_policy` | No | Edge case |
| `cps` | No | Advanced tuning |
| `mss` | No | Advanced tuning |
| `start_port/end_port` | No | Port range, edge case |
| `sec_sensor_enabled` | No | Security sensor |

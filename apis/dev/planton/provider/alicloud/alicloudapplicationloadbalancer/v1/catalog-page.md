# AliCloud ALB Load Balancer

Deploys an Alibaba Cloud Application Load Balancer (ALB) with bundled server groups and listeners. ALB is a modern Layer 7 load balancer for HTTP, HTTPS, and QUIC traffic, offering advanced routing, health checking, and session stickiness.

## What Gets Created

When you deploy an AliCloudApplicationLoadBalancer resource, Planton provisions:

- **ALB Load Balancer** -- an `alicloud_alb_load_balancer` spanning multiple availability zones for high availability
- **Server Groups** -- one `alicloud_alb_server_group` per entry in `serverGroups`, each with health check and optional session stickiness
- **Listeners** -- one `alicloud_alb_listener` per entry in `listeners`, forwarding traffic to a server group

Server groups are created empty. Backend membership (ECS instances, ENIs, IPs) is managed externally by ACK ingress controllers, SAE bindings, or manual attachment.

## Prerequisites

- **Alibaba Cloud credentials** configured via environment variables or Planton provider config
- **An Alibaba Cloud VPC** -- the ALB must belong to a VPC (create one with AliCloudVpc)
- **At least 2 VSwitches in different availability zones** -- ALB requires multi-AZ deployment (create with AliCloudVswitch)
- **A server certificate** (for HTTPS listeners) -- obtain from Alibaba Cloud Certificate Management Service (CAS)

## Quick Start

Create a file `alb.yaml`:

```yaml
apiVersion: alicloud.planton.dev/v1
kind: AliCloudApplicationLoadBalancer
metadata:
  name: my-alb
spec:
  region: cn-hangzhou
  vpcId:
    valueFrom:
      name: my-vpc
  zoneMappings:
    - zoneId: cn-hangzhou-a
      vswitchId:
        valueFrom:
          name: my-vswitch-a
    - zoneId: cn-hangzhou-b
      vswitchId:
        valueFrom:
          name: my-vswitch-b
  serverGroups:
    - name: web-backend
      healthCheckConfig:
        healthCheckEnabled: true
        healthCheckPath: /health
  listeners:
    - listenerPort: 80
      listenerProtocol: HTTP
      defaultActionServerGroupName: web-backend
```

Deploy:

```shell
planton apply -f alb.yaml
```

This creates an internet-facing ALB with an HTTP listener on port 80 across two availability zones.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | string | Alibaba Cloud region (e.g., `cn-hangzhou`, `cn-shanghai`) | Required; non-empty |
| `vpcId` | StringValueOrRef | VPC ID for the ALB. Can reference AliCloudVpc via `valueFrom`. | Required |
| `zoneMappings` | list | Availability zone to VSwitch mappings for HA | Minimum 2 items required |

### Zone Mapping Fields

| Field | Type | Description |
|-------|------|-------------|
| `zoneMappings[].zoneId` | string | Availability zone ID (e.g., `cn-hangzhou-a`) |
| `zoneMappings[].vswitchId` | StringValueOrRef | VSwitch in this zone. Can reference AliCloudVswitch via `valueFrom`. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `loadBalancerName` | string | metadata.name | ALB name (2-128 characters) |
| `addressType` | string | `Internet` | Network type: `Internet` or `Intranet` |
| `loadBalancerEdition` | string | `Standard` | Edition: `Basic`, `Standard`, `StandardWithWaf` |
| `resourceGroupId` | string | | Resource group for organizational grouping |
| `accessLogConfig` | object | | SLS access log shipping (see below) |
| `tags` | map | | Key-value tags for the ALB |

### Access Log Config Fields

| Field | Type | Description |
|-------|------|-------------|
| `accessLogConfig.logProject` | string | SLS log project name (must exist in the same region) |
| `accessLogConfig.logStore` | string | SLS log store name within the log project |

### Server Group Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `name` | string | *required* | Server group name (2-128 chars). Referenced by listeners via `defaultActionServerGroupName`. |
| `protocol` | string | `HTTP` | Backend protocol: `HTTP`, `HTTPS`, `GRPC` |
| `scheduler` | string | `Wrr` | Scheduling algorithm: `Wrr` (weighted round robin), `Wlc` (weighted least connections), `Sch` (source IP hash) |
| `healthCheckConfig` | object | *required* | Health check configuration (see below) |
| `stickySessionConfig` | object | | Session stickiness settings (see below) |

### Health Check Config Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `healthCheckEnabled` | bool | | Whether health checks are active. When `false`, all servers are considered healthy. |
| `healthCheckProtocol` | string | `HTTP` | Probe protocol: `HTTP`, `HTTPS`, `TCP`, `GRPC` |
| `healthCheckPath` | string | | URL path for HTTP/HTTPS probes (e.g., `/health`). Ignored for TCP. |
| `healthCheckHost` | string | | Host header for HTTP/HTTPS probes. If omitted, uses the server's IP. |
| `healthCheckMethod` | string | `HEAD` | HTTP method: `GET`, `POST`, `HEAD` |
| `healthCheckConnectPort` | int | `0` | Port for probes. `0` uses the backend server's port. (0-65535) |
| `healthCheckInterval` | int | `2` | Seconds between probes (1-50) |
| `healthCheckTimeout` | int | `5` | Probe response timeout in seconds (1-300) |
| `healthyThreshold` | int | `3` | Consecutive successes to mark healthy (2-10) |
| `unhealthyThreshold` | int | `3` | Consecutive failures to mark unhealthy (2-10) |
| `healthCheckCodes` | list | | Healthy response codes (e.g., `http_2xx`, `http_3xx`). HTTP/HTTPS only. |

### Sticky Session Config Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `stickySessionEnabled` | bool | | Whether session stickiness is enabled |
| `stickySessionType` | string | | Method: `Insert` (ALB inserts cookie) or `Server` (backend sets cookie) |
| `cookie` | string | | Cookie name when `stickySessionType` is `Server` |
| `cookieTimeout` | int | `1000` | Cookie timeout in seconds when `stickySessionType` is `Insert` (1-86400) |

### Listener Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `listenerPort` | int | *required* | Port to accept traffic (1-65535) |
| `listenerProtocol` | string | *required* | Protocol: `HTTP`, `HTTPS`, `QUIC` |
| `defaultActionServerGroupName` | string | *required* | Target server group name (must match a `serverGroups[].name`) |
| `listenerDescription` | string | | Human-readable purpose of this listener (2-256 characters) |
| `certificateId` | string | | Certificate ID from CAS (required for HTTPS and QUIC) |
| `securityPolicyId` | string | | TLS cipher policy (e.g., `tls_cipher_policy_1_2_strict`). HTTPS and QUIC only. |
| `gzipEnabled` | bool | `true` | Enable gzip compression for HTTP responses |
| `http2Enabled` | bool | `true` | Enable HTTP/2. HTTPS only. |
| `idleTimeout` | int | `60` | Connection idle timeout in seconds (1-60) |
| `requestTimeout` | int | `60` | Backend request timeout in seconds (1-180). Returns 504 if exceeded. |

## Examples

### Internet-Facing HTTP ALB

The simplest ALB: one server group, one HTTP listener, two availability zones.

```yaml
apiVersion: alicloud.planton.dev/v1
kind: AliCloudApplicationLoadBalancer
metadata:
  name: web-alb
spec:
  region: cn-hangzhou
  vpcId:
    value: vpc-abc123
  zoneMappings:
    - zoneId: cn-hangzhou-a
      vswitchId:
        value: vsw-aaa
    - zoneId: cn-hangzhou-b
      vswitchId:
        value: vsw-bbb
  serverGroups:
    - name: web-backend
      healthCheckConfig:
        healthCheckEnabled: true
        healthCheckPath: /health
  listeners:
    - listenerPort: 80
      listenerProtocol: HTTP
      defaultActionServerGroupName: web-backend
```

### HTTPS ALB with Certificate

Production ALB with TLS termination, WAF edition, and strict cipher policy.

```yaml
apiVersion: alicloud.planton.dev/v1
kind: AliCloudApplicationLoadBalancer
metadata:
  name: secure-alb
  org: acme-corp
  env: production
spec:
  region: cn-shanghai
  vpcId:
    valueFrom:
      name: prod-vpc
  loadBalancerEdition: StandardWithWaf
  zoneMappings:
    - zoneId: cn-shanghai-a
      vswitchId:
        valueFrom:
          name: prod-vswitch-a
    - zoneId: cn-shanghai-b
      vswitchId:
        valueFrom:
          name: prod-vswitch-b
  tags:
    team: platform
    cost-center: shared-infra
  serverGroups:
    - name: api-backend
      protocol: HTTPS
      healthCheckConfig:
        healthCheckEnabled: true
        healthCheckProtocol: HTTPS
        healthCheckPath: /healthz
        healthyThreshold: 5
        unhealthyThreshold: 2
  listeners:
    - listenerPort: 443
      listenerProtocol: HTTPS
      defaultActionServerGroupName: api-backend
      certificateId: cas-abc123
      securityPolicyId: tls_cipher_policy_1_2_strict
```

### Internal ALB with Multiple Server Groups

An internal ALB for service-to-service HTTP routing with two server groups and weighted least connections scheduling.

```yaml
apiVersion: alicloud.planton.dev/v1
kind: AliCloudApplicationLoadBalancer
metadata:
  name: internal-alb
spec:
  region: cn-hangzhou
  vpcId:
    value: vpc-internal
  addressType: Intranet
  zoneMappings:
    - zoneId: cn-hangzhou-a
      vswitchId:
        value: vsw-internal-a
    - zoneId: cn-hangzhou-b
      vswitchId:
        value: vsw-internal-b
  serverGroups:
    - name: api-v1
      healthCheckConfig:
        healthCheckEnabled: true
        healthCheckPath: /api/v1/health
    - name: api-v2
      scheduler: Wlc
      healthCheckConfig:
        healthCheckEnabled: true
        healthCheckPath: /api/v2/health
  listeners:
    - listenerPort: 80
      listenerProtocol: HTTP
      defaultActionServerGroupName: api-v2
      listenerDescription: Primary API listener
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `load_balancer_id` | string | ALB instance ID (e.g., `alb-xxxxx`) |
| `dns_name` | string | DNS name assigned to the ALB. For internet-facing ALBs, resolves to the public address. Use as a CNAME target for custom domains. |
| `server_group_ids` | map&lt;string, string&gt; | Map of server group names to their IDs (e.g., `{"web-backend": "sgp-xxxxx"}`) |

## Related Components

- **AliCloudVpc** -- VPC that the ALB belongs to
- **AliCloudVswitch** -- VSwitches for zone mappings (at least 2 required)
- **AliCloudSecurityGroup** -- Network security rules for backend instances
- **AliCloudDnsRecord** -- CNAME records pointing to the ALB's `dns_name`
- **AliCloudAckManagedCluster** -- Kubernetes cluster whose ingress uses the ALB

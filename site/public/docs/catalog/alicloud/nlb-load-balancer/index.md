---
title: "NLB Load Balancer"
description: "NLB Load Balancer deployment documentation"
icon: "package"
order: 100
componentName: "alicloudnetworkloadbalancer"
---

# AliCloud NLB Load Balancer

Deploys an Alibaba Cloud Network Load Balancer (NLB) with bundled server groups and listeners. NLB is a Layer 4 load balancer for TCP, UDP, and TCPSSL traffic, designed for ultra-high throughput and low latency with multi-AZ high availability.

## What Gets Created

When you deploy an AliCloudNetworkLoadBalancer resource, Planton provisions:

- **NLB Load Balancer** -- an `alicloud_nlb_load_balancer` spanning multiple availability zones, with optional per-zone EIP binding for stable public addresses
- **Server Groups** -- one `alicloud_nlb_server_group` per entry in `serverGroups`, each with health check, scheduling algorithm, and optional connection draining
- **Listeners** -- one `alicloud_nlb_listener` per entry in `listeners`, forwarding TCP, UDP, or TCPSSL traffic to a server group

Server groups are created empty. Backend membership (ECS instances, ENI IPs, etc.) is managed externally by ACK service controllers, manual attachment, or other orchestration.

## Prerequisites

- **Alibaba Cloud credentials** configured via environment variables or Planton provider config
- **An Alibaba Cloud VPC** -- the NLB must belong to a VPC (create one with AliCloudVpc)
- **At least 2 VSwitches in different availability zones** -- NLB requires multi-AZ deployment (create with AliCloudVswitch)
- **Server certificates** (for TCPSSL listeners) -- obtain from Alibaba Cloud Certificate Management Service (CAS)
- **EIP addresses** (optional) -- for fixed public IPs per zone (create with AliCloudEipAddress)

## Quick Start

Create a file `nlb.yaml`:

```yaml
apiVersion: alicloud.planton.dev/v1
kind: AliCloudNetworkLoadBalancer
metadata:
  name: my-nlb
spec:
  region: cn-hangzhou
  vpcId:
    value: vpc-abc123
  zoneMappings:
    - zoneId: cn-hangzhou-a
      vswitchId:
        value: vsw-zone-a
    - zoneId: cn-hangzhou-b
      vswitchId:
        value: vsw-zone-b
  serverGroups:
    - name: tcp-backend
      healthCheck:
        healthCheckEnabled: true
  listeners:
    - listenerPort: 80
      listenerProtocol: TCP
      serverGroupName: tcp-backend
```

Deploy:

```shell
planton apply -f nlb.yaml
```

This creates an internet-facing NLB with a TCP listener on port 80 across two availability zones.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | string | Alibaba Cloud region (e.g., `cn-hangzhou`, `us-west-1`) | Required; non-empty |
| `vpcId` | StringValueOrRef | VPC ID for the NLB. Can reference AliCloudVpc via `valueFrom`. | Required |
| `zoneMappings` | list | Availability zone to VSwitch mappings for HA | Minimum 2 items required |

### Zone Mapping Fields

| Field | Type | Description |
|-------|------|-------------|
| `zoneMappings[].zoneId` | string | Availability zone ID (e.g., `cn-hangzhou-a`) |
| `zoneMappings[].vswitchId` | StringValueOrRef | VSwitch in this zone. Can reference AliCloudVswitch via `valueFrom`. |
| `zoneMappings[].allocationId` | StringValueOrRef | EIP allocation ID for a fixed public IP in this zone. Can reference AliCloudEipAddress via `valueFrom`. Only meaningful for internet-facing NLBs. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `loadBalancerName` | string | metadata.name | NLB name (2-128 characters, starts with a letter) |
| `addressType` | string | `Internet` | Network type: `Internet` or `Intranet` |
| `resourceGroupId` | string | | Resource group for organizational grouping |
| `crossZoneEnabled` | bool | `true` | Distribute traffic across all zones. When `false`, traffic stays within the receiving zone. |
| `tags` | map | | Key-value tags applied to the NLB |

### Server Group Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `name` | string | *required* | Server group name (2-128 chars). Referenced by listeners via `serverGroupName`. |
| `protocol` | string | `TCP` | Backend protocol: `TCP`, `UDP`, `TCPSSL` |
| `scheduler` | string | `Wrr` | Scheduling algorithm: `Wrr` (weighted round robin), `Rr` (round robin), `Sch` (source IP hash), `Tch` (four-tuple hash), `Qch` (QUIC ID hash), `Wlc` (weighted least connections) |
| `connectionDrainEnabled` | bool | `false` | Allow in-flight connections to complete when a backend is removed |
| `connectionDrainTimeout` | int | `10` | Seconds to wait for draining connections (10-900). Only effective when `connectionDrainEnabled` is `true`. |
| `preserveClientIpEnabled` | bool | `true` | Forward the real client IP to backends instead of the NLB's IP |
| `healthCheck` | object | *required* | Health check configuration (see below) |

### Health Check Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `healthCheckEnabled` | bool | | Whether health checks are active. When `false`, all servers are considered healthy. |
| `healthCheckType` | string | `TCP` | Probe protocol: `TCP`, `HTTP`, `UDP` |
| `healthCheckConnectPort` | int | `0` | Port for probes. `0` uses the backend server's port. (0-65535) |
| `healthCheckConnectTimeout` | int | `5` | Probe response timeout in seconds (1-300) |
| `healthCheckInterval` | int | `10` | Seconds between probes (5-50) |
| `healthyThreshold` | int | `2` | Consecutive successes to mark healthy (2-10) |
| `unhealthyThreshold` | int | `2` | Consecutive failures to mark unhealthy (2-10) |
| `healthCheckUrl` | string | | URL path for HTTP probes (e.g., `/health`). Only for `healthCheckType: HTTP`. |
| `healthCheckDomain` | string | | Host header for HTTP probes. Only for `healthCheckType: HTTP`. |
| `httpCheckMethod` | string | `GET` | HTTP method: `GET` or `HEAD`. Only for `healthCheckType: HTTP`. |
| `healthCheckHttpCodes` | list | | Healthy response codes (e.g., `http_2xx`, `http_3xx`). Only for `healthCheckType: HTTP`. |

### Listener Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `listenerPort` | int | *required* | Port to accept traffic (1-65535) |
| `listenerProtocol` | string | *required* | Protocol: `TCP`, `UDP`, `TCPSSL` |
| `serverGroupName` | string | *required* | Target server group name (must match a `serverGroups[].name`) |
| `listenerDescription` | string | | Human-readable purpose of this listener |
| `idleTimeout` | int | `900` | Seconds before idle connections are closed (1-900). TCP and TCPSSL only. |
| `proxyProtocolEnabled` | bool | `false` | Insert Proxy Protocol header so backends see the real client IP/port |
| `certificateIds` | list | | Server certificate IDs for TCPSSL. At least one required when `listenerProtocol` is `TCPSSL`. |
| `securityPolicyId` | string | | TLS cipher policy (e.g., `tls_cipher_policy_1_2_strict`). TCPSSL only. |
| `caCertificateIds` | list | | CA certificate IDs for mutual TLS on TCPSSL listeners |
| `caEnabled` | bool | `false` | Enable client certificate verification on TCPSSL listeners. Requires `caCertificateIds`. |

## Examples

### Internet-Facing TCP

The simplest NLB: one server group, one TCP listener, two availability zones.

```yaml
apiVersion: alicloud.planton.dev/v1
kind: AliCloudNetworkLoadBalancer
metadata:
  name: dev-nlb
spec:
  region: cn-hangzhou
  vpcId:
    value: vpc-abc123
  zoneMappings:
    - zoneId: cn-hangzhou-a
      vswitchId:
        value: vsw-zone-a
    - zoneId: cn-hangzhou-b
      vswitchId:
        value: vsw-zone-b
  serverGroups:
    - name: tcp-backend
      healthCheck:
        healthCheckEnabled: true
  listeners:
    - listenerPort: 80
      listenerProtocol: TCP
      serverGroupName: tcp-backend
```

### TCPSSL with Mutual TLS and Fixed EIPs

Production NLB with TLS termination, client certificate verification, fixed public IPs per zone, and connection draining for graceful deployments.

```yaml
apiVersion: alicloud.planton.dev/v1
kind: AliCloudNetworkLoadBalancer
metadata:
  name: prod-nlb
  org: acme-corp
  env: production
spec:
  region: cn-shanghai
  vpcId:
    valueFrom:
      name: prod-vpc
  zoneMappings:
    - zoneId: cn-shanghai-a
      vswitchId:
        valueFrom:
          name: prod-vswitch-a
      allocationId:
        valueFrom:
          name: prod-eip-a
    - zoneId: cn-shanghai-b
      vswitchId:
        valueFrom:
          name: prod-vswitch-b
      allocationId:
        valueFrom:
          name: prod-eip-b
  tags:
    team: platform
    cost-center: shared-infra
  serverGroups:
    - name: api-backend
      protocol: TCPSSL
      scheduler: Wlc
      connectionDrainEnabled: true
      connectionDrainTimeout: 300
      healthCheck:
        healthCheckEnabled: true
        healthCheckType: HTTP
        healthCheckUrl: /healthz
        httpCheckMethod: GET
        healthCheckInterval: 10
        healthyThreshold: 3
        unhealthyThreshold: 2
        healthCheckHttpCodes:
          - http_2xx
  listeners:
    - listenerPort: 443
      listenerProtocol: TCPSSL
      serverGroupName: api-backend
      certificateIds:
        - cas-prod-cert
      securityPolicyId: tls_cipher_policy_1_2_strict
      caCertificateIds:
        - ca-prod-cert
      caEnabled: true
      listenerDescription: Production TCPSSL with mutual TLS
```

### Internal VPC-Private NLB

An internal NLB for service-to-service traffic with source-IP consistent hashing for session affinity, connection draining for graceful deployments, and Proxy Protocol for real client IP visibility.

```yaml
apiVersion: alicloud.planton.dev/v1
kind: AliCloudNetworkLoadBalancer
metadata:
  name: internal-nlb
spec:
  region: cn-hangzhou
  vpcId:
    value: vpc-internal
  addressType: Intranet
  crossZoneEnabled: false
  zoneMappings:
    - zoneId: cn-hangzhou-a
      vswitchId:
        value: vsw-internal-a
    - zoneId: cn-hangzhou-b
      vswitchId:
        value: vsw-internal-b
  serverGroups:
    - name: db-proxy
      scheduler: Sch
      connectionDrainEnabled: true
      connectionDrainTimeout: 60
      preserveClientIpEnabled: true
      healthCheck:
        healthCheckEnabled: true
        healthCheckType: TCP
        healthCheckConnectPort: 3306
        healthCheckInterval: 5
        healthyThreshold: 2
        unhealthyThreshold: 2
  listeners:
    - listenerPort: 3306
      listenerProtocol: TCP
      serverGroupName: db-proxy
      idleTimeout: 600
      proxyProtocolEnabled: true
      listenerDescription: MySQL proxy with connection draining
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `load_balancer_id` | string | NLB instance ID (e.g., `nlb-xxxxx`) |
| `dns_name` | string | DNS name assigned to the NLB. For internet-facing NLBs, resolves to the public address. Use as a CNAME target for custom domains. |
| `server_group_ids` | map&lt;string, string&gt; | Map of server group names to their IDs (e.g., `{"tcp-backend": "sgp-xxxxx"}`). Use for downstream backend attachment. |

## Related Components

- **AliCloudVpc** -- VPC that the NLB belongs to
- **AliCloudVswitch** -- VSwitches for zone mappings (at least 2 required)
- **AliCloudEipAddress** -- Fixed public IPs for per-zone EIP binding
- **AliCloudDnsRecord** -- DNS records pointing to the NLB's `dns_name`
- **AliCloudAckManagedCluster** -- Kubernetes cluster whose services use the NLB

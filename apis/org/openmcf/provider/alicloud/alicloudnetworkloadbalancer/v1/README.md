# AliCloudNetworkLoadBalancer

Alibaba Cloud Network Load Balancer (NLB) for high-performance Layer 4 (TCP/UDP/TCPSSL) load balancing.

## Overview

This component provisions an NLB with bundled server groups and listeners. NLB is designed for ultra-low latency and high throughput TCP/UDP workloads such as database proxying, game servers, IoT backends, and microservice-to-microservice traffic.

## What Gets Created

- **NLB Load Balancer** -- The L4 load balancer with multi-AZ zone mappings
- **Server Groups** -- Backend target groups with health checks, connection draining, and scheduling
- **Listeners** -- TCP, UDP, or TCPSSL listeners that route to server groups

Server groups are created empty. Backend membership (ECS instances, ENI IPs, etc.) is managed externally by ACK service controllers, manual attachment, or other orchestration.

## Key Features

- **Multi-AZ high availability** -- Minimum 2 zone mappings required
- **Fixed public IPs** -- Optional EIP binding per zone for stable public addresses
- **Connection draining** -- Graceful backend removal with configurable drain timeout
- **TCPSSL termination** -- TLS termination at L4 with mutual TLS support
- **Proxy Protocol** -- Real client IP/port forwarding to backends
- **Cross-zone balancing** -- Distribute traffic across all zones or keep it zone-local
- **Six scheduling algorithms** -- Wrr, Rr, Sch, Tch, Qch, Wlc

## Quick Start

```yaml
apiVersion: alicloud.openmcf.org/v1
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

See [examples.md](examples.md) for more configurations.

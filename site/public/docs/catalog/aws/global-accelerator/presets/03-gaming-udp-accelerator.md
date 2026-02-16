---
title: "Gaming UDP Accelerator"
description: "This preset creates a Global Accelerator optimized for real-time gaming workloads. It uses UDP protocol across a port range of 7000–8000, enables `SOURCE_IP` client affinity to pin each player to the..."
type: "preset"
rank: "03"
presetSlug: "03-gaming-udp-accelerator"
componentSlug: "global-accelerator"
componentTitle: "Global Accelerator"
provider: "aws"
icon: "package"
order: 3
---

# Gaming UDP Accelerator

This preset creates a Global Accelerator optimized for real-time gaming workloads. It uses UDP protocol across a port range of 7000–8000, enables `SOURCE_IP` client affinity to pin each player to the same game server for their session, and routes traffic to Elastic IP endpoints attached to dedicated game server instances. Health checks use TCP on a sidecar port since Global Accelerator does not support UDP health checks.

## When to Use

- You run multiplayer game servers that use UDP for low-latency game state synchronization
- Players need to maintain a persistent connection to the same game server throughout a session (SOURCE_IP affinity)
- You want to leverage AWS's global backbone to minimize latency and jitter for players worldwide
- Your game servers are deployed on EC2 instances with Elastic IPs, and you need static anycast IPs as the global entry point

## Key Configuration Choices

- **UDP protocol** — essential for real-time gaming traffic where low latency matters more than guaranteed delivery
- **SOURCE_IP affinity** — ensures all packets from a player's IP address reach the same game server, maintaining session state without application-level session routing
- **Port range 7000–8000** — a contiguous range for game traffic; adjust to match your game server's port allocation
- **TCP health check on port 8080** — since UDP health checks aren't supported, this relies on a sidecar HTTP/TCP health endpoint; ensure each game server exposes a TCP listener on port 8080 (or adjust the port to match your health check endpoint)
- **10-second interval, threshold 2** — detects a failed game server within 20 seconds, minimizing the window where players are routed to an unhealthy server
- **Two endpoints with equal weight (128)** — distributes players evenly across two game server instances; add more endpoints or adjust weights to match your server fleet size

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<eip-allocation-id-1>` | Elastic IP allocation ID for the first game server (e.g., `eipalloc-0abcdef1234567890`) | AWS EC2 Console → Elastic IPs, or `AwsElasticIp` status outputs |
| `<eip-allocation-id-2>` | Elastic IP allocation ID for the second game server | AWS EC2 Console → Elastic IPs, or `AwsElasticIp` status outputs |

## Related Presets

- **01-basic-tcp-accelerator** — Use for TCP-based workloads like web applications and APIs
- **02-multi-region-production** — Use for multi-region TCP deployments with HTTP health checks and flow logs

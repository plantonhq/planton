---
title: "Private Internal Load Balancer"
description: "This preset creates a load balancer attached to a Hetzner Cloud private network with its public interface disabled. It distributes HTTP traffic across explicit server targets using their private IPs,..."
type: "preset"
rank: "02"
presetSlug: "02-private-internal"
componentSlug: "hetzner-cloud-load-balancer"
componentTitle: "Hetzner Cloud Load Balancer"
provider: "hetznercloud"
icon: "package"
order: 2
---

# Private Internal Load Balancer

This preset creates a load balancer attached to a Hetzner Cloud private network with its public interface disabled. It distributes HTTP traffic across explicit server targets using their private IPs, making it invisible from the internet. Use this for internal service routing between application tiers -- API gateways distributing to microservices, application servers connecting to database pools, or any backend-to-backend communication that should never traverse the public internet.

No TLS is configured because traffic stays within the private network. If your internal security policy requires encryption in transit even on private networks, use the `01-https-web-app` preset instead and attach it to a network with `enablePublicInterface: false`.

## When to Use

- Internal load balancing between application tiers that communicate over a private network
- Services that must not be reachable from the public internet
- Backend pools where you know the exact server IDs (as opposed to dynamic label-based discovery)

## Key Configuration Choices

- **No public interface** (`enablePublicInterface: false`) -- the load balancer has no public IP and is only reachable via its private network address; this is the defining trait of an internal LB
- **Private network attachment** (`network.networkId`) -- the LB joins the specified Hetzner Cloud network and receives an auto-assigned private IP within the subnet range
- **Private IP routing** (`usePrivateIp: true`) -- traffic to each server target flows over the private network instead of the public internet; both the LB and the servers must be attached to the same network
- **Plain HTTP** (`protocol: http`) -- TLS termination is unnecessary for private network traffic; the service listens on port 80 (the protocol default) and forwards to backends on port 8080
- **Explicit server targets** -- two targets are shown to demonstrate the multi-backend pattern; add or remove entries as your backend pool changes
- **Delete protection** (`deleteProtection: true`) -- internal LBs are often invisible in monitoring until they fail; protection prevents accidental removal

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<backend-server-id-1>` | Numeric ID of the first backend server | The `status.outputs.server_id` of your HetznerCloudServer resource, or the Servers page in the Hetzner Cloud Console |
| `<backend-server-id-2>` | Numeric ID of the second backend server | Same as above for your second backend server |
| `<network-id>` | Numeric ID of the Hetzner Cloud network to attach the LB to | The `status.outputs.network_id` of your HetznerCloudNetwork resource, or the Networks page in the Hetzner Cloud Console |

## Related Presets

- **01-https-web-app** -- public HTTPS load balancer with TLS termination and label-based target discovery
- **03-tcp-pass-through** -- layer-4 TCP balancing for non-HTTP protocols

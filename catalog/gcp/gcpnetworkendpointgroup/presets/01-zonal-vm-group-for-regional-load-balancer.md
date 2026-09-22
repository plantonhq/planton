# Zonal VM Group for a Regional Load Balancer

## Use Case

Put a set of VMs behind an Application Load Balancer without an instance group: a zonal `GCE_VM_IP_PORT` group names each VM (and optionally an alias IP and port), and the backend service points at the group. Declare one group per zone and list them all on the backend service for zone-level resilience.

## When to Use

- VMs managed outside a managed instance group (hand-placed, or from another controller) that must sit behind an Application, proxy, or regional load balancer
- Container-native load balancing shapes where the endpoint is a specific IP:port on a VM
- Any regional or global external Application Load Balancer backend built from VMs by name

## What This Creates

- A zonal group `web-neg-a` in `us-central1-a` on `main-vpc` / `web-subnet`
- Type `GCE_VM_IP_PORT`, default port 8080
- Two endpoints, `web-1` and `web-2`, on their primary internal IPs and the default port, written as one set

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `zone` | `us-central1-a` | The zone the VMs live in; one group per zone. Immutable. |
| `network`, `subnetwork` | `main-vpc`, `web-subnet` | The VPC and subnet the VMs' IPs fall in. Immutable. |
| `defaultPort` | `8080` | The port most endpoints listen on; override per endpoint with `port`. |
| `endpoints[]` | two VMs | Your `GcpComputeInstance` references; add `ipAddress` to pick an alias IP, `port` to override the default. Leave the list empty when an autoscaler owns membership. |
| `networkEndpointType` | `GCE_VM_IP_PORT` | `GCE_VM_IP` (no ports) for a passthrough Network Load Balancer backend. |

The consuming `GcpBackendService` names this group's `self_link` output in `backends[].group` with a RATE balancing mode.

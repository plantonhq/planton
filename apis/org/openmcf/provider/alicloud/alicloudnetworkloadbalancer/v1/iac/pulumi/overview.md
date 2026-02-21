# Pulumi Module Overview

## Architecture

The Pulumi module creates three resource types in sequence:

1. **NLB Load Balancer** (`nlb.LoadBalancer`) -- The main L4 load balancer with multi-AZ zone mappings and optional EIP bindings.
2. **Server Groups** (`nlb.ServerGroup`) -- Backend target groups with health checks, scheduling, and connection draining. Created empty; membership is managed externally.
3. **Listeners** (`nlb.Listener`) -- TCP/UDP/TCPSSL listeners that forward traffic to server groups by ID lookup.

## File Layout

- `main.go` -- Entry point: creates provider, NLB, server groups, listeners, exports outputs.
- `locals.go` -- Initializes locals (tags, default helpers for optional fields).
- `outputs.go` -- Output key constants.
- `server_groups.go` -- Server group creation with health check configuration.
- `listeners.go` -- Listener creation with TCPSSL certificate support.

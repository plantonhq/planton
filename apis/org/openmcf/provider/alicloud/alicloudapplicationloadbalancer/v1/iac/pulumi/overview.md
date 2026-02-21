# AliCloudApplicationLoadBalancer Pulumi Module

## Architecture

The Pulumi module creates three resource types in sequence:

1. **ALB Load Balancer** (`alb.LoadBalancer`) -- the load balancer itself, spanning multiple AZs via zone mappings
2. **Server Groups** (`alb.ServerGroup`) -- one per entry in `spec.serverGroups`, each with health check and optional sticky session
3. **Listeners** (`alb.Listener`) -- one per entry in `spec.listeners`, each with a ForwardGroup default action pointing to a server group

## Resource Dependencies

```
ALB LoadBalancer
├── ServerGroup "web-backend"
├── ServerGroup "api-backend"
├── Listener :80 HTTP → ServerGroup "web-backend"
└── Listener :443 HTTPS → ServerGroup "api-backend"
```

Listeners depend on both the ALB (for `load_balancer_id`) and their target server group (for `server_group_id` in the default action). Server groups are independent of each other.

## File Structure

| File | Purpose |
| --- | --- |
| `module/main.go` | Provider setup, ALB creation, orchestration of server groups and listeners |
| `module/locals.go` | Tag computation, default value helpers |
| `module/server_groups.go` | Server group creation with health check and sticky session |
| `module/listeners.go` | Listener creation with default ForwardGroup action |
| `module/outputs.go` | Stack output key constants |

## Key Design Decisions

- **Billing hardcoded**: `PayAsYouGo` is the only billing mode ALB supports; not exposed in the spec
- **Server groups created empty**: Backend membership is managed externally (ACK ingress, SAE, manual)
- **Server group name as key**: Listeners reference server groups by name, resolved to IDs at creation time
- **Forwarding rules excluded**: ALB rules are extremely complex and managed separately if needed

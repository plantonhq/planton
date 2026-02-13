# AzureLoadBalancer Pulumi Module

Pulumi implementation for the AzureLoadBalancer deployment component.

## Architecture

The module creates:

- `lb.LoadBalancer` -- Load Balancer with Standard SKU and frontend IP config
- `lb.BackendAddressPool` -- One per backend pool in the spec
- `lb.Probe` -- One per health probe in the spec
- `lb.Rule` -- One per load balancing rule in the spec

## Package Structure

```
iac/pulumi/
├── main.go              # Entrypoint: loads stack input, calls module
├── Pulumi.yaml          # Pulumi project configuration
├── Makefile             # Build and test targets
├── debug.sh             # Delve debugger script
├── overview.md          # Architecture overview
└── module/
    ├── main.go          # Resources(): creates LB + pools + probes + rules
    ├── locals.go        # initializeLocals(): parses input, builds tags
    └── outputs.go       # Output key constants
```

## Resources Created

| Resource | Pulumi Type | Description |
|----------|-------------|-------------|
| Load Balancer | `lb.LoadBalancer` | Standard SKU LB with frontend config |
| Backend Pool | `lb.BackendAddressPool` | One per pool (membership managed externally) |
| Health Probe | `lb.Probe` | One per probe (Tcp, Http, or Https) |
| Rule | `lb.Rule` | One per rule (maps frontend to backend) |

## Local Development

```bash
make deps    # Tidy Go modules
make build   # Build module and entrypoint
make test    # Run tests
```

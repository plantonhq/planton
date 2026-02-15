# AwsMemcachedElasticache Pulumi Module

This directory contains the Pulumi IaC module for provisioning AWS ElastiCache Memcached clusters.

## Structure

```
.
├── main.go              # Entrypoint: loads stack input, calls module.Resources
├── overview.md          # Architecture documentation
├── module/
│   ├── main.go          # Orchestrator: provider setup, resource sequencing
│   ├── locals.go        # Tag computation and spec references
│   ├── outputs.go       # Output key constants
│   ├── subnet_group.go  # Conditional subnet group creation
│   ├── parameter_group.go  # Conditional parameter group creation
│   └── cluster.go       # Main Memcached cluster resource
```

## Local Development

```bash
cd module && go build ./...
```

## Testing

```bash
cd ../.. && go test ./...
```

# OpenStackContainerCluster Pulumi Module

This directory contains the Pulumi Go module for provisioning OpenStack Magnum container clusters.

## Structure

```
iac/pulumi/
├── main.go           # Entrypoint (Pulumi program)
├── Pulumi.yaml       # Pulumi project config
├── Makefile          # Build and test targets
├── module/
│   ├── main.go       # Resources() entry point
│   ├── locals.go     # Input extraction and FK resolution
│   ├── cluster.go    # Magnum cluster resource creation
│   └── outputs.go    # Output constants
└── README.md         # This file
```

## How It Works

1. The Planton CLI serializes `OpenStackContainerClusterStackInput` and passes it to Pulumi
2. `main.go` loads the stack input and calls `module.Resources()`
3. `locals.go` extracts the resolved FK values (`ClusterTemplate`, `Keypair`)
4. `cluster.go` creates `containerinfra.NewCluster()` with the extracted values
5. Outputs are exported matching `stack_outputs.proto` field names

## Sensitive Outputs

The following outputs are marked as secret using `pulumi.ToSecret()`:
- `kubeconfig_raw` — Full kubeconfig YAML
- `kubeconfig_cluster_ca_cert` — Cluster CA certificate
- `kubeconfig_client_cert` — Client certificate
- `kubeconfig_client_key` — Client private key

## Local Development

```bash
# Build
make build

# Install plugins
make install-pulumi-plugins
```

## Key Design Notes

- **Single resource**: Creates one `containerinfra.Cluster` (Magnum cluster)
- **FK extraction**: Required FK (`ClusterTemplate`) always resolved; optional FK (`Keypair`) nil-guarded
- **Sensitive kubeconfig**: Raw kubeconfig and certificate outputs are marked as Pulumi secrets
- **ForceNew fields**: Almost all fields are ForceNew; only `node_count` and `cluster_template` support updates

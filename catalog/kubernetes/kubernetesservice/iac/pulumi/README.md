# Kubernetes Service - Pulumi Module

## Overview

This Pulumi module creates and manages a standalone Kubernetes Service. It supports the complete core/v1 ServiceSpec surface: all four service types (ClusterIP, NodePort, LoadBalancer, ExternalName), headless services, traffic policies, topology-aware `trafficDistribution`, session affinity, LoadBalancer tuning knobs, and dual-stack addressing.

This is the reference engine for the component: it applies the full spec, including `traffic_distribution`, which the Terraform module cannot (the Terraform kubernetes provider does not expose the field — the Terraform module fails its plan loudly when it is set).

## Architecture

```
iac/pulumi/
├── main.go          # Entrypoint: loads IaC input, calls module
├── Pulumi.yaml      # Pulumi project configuration
├── Makefile         # Make targets for preview/up/down/refresh
└── module/
    ├── main.go      # Orchestrator: provider init, resource creation, output export
    ├── locals.go    # Derived values: labels, namespace default, enum → API string translation
    ├── service.go   # Creates kubernetes.core.v1.Service resource
    └── outputs.go   # Exports the eight outputs
```

## How It Works

1. **IaC Input Loading**: The entrypoint loads `KubernetesServiceIacInput` from Pulumi config
2. **Locals Initialization**: `locals.go` computes:
   - Standard Planton labels merged with user labels (identity keys win on conflict)
   - The target namespace (foreign-key references are pre-resolved; falls back to `default` when omitted)
   - Kubernetes API string forms of every spec enum (`load_balancer` → `LoadBalancer`, `client_ip` → `ClientIP`, `prefer_same_zone` → `PreferSameZone`, ...), resolved once so the resource and the outputs agree on wire values
3. **Provider Creation**: Kubernetes provider is initialized from `provider_config`
4. **Service Creation**: `service.go` builds the ServiceSpec, sending each optional field only when the user set it — for several fields (clusterIP, healthCheckNodePort, loadBalancerClass) an empty value is not the same as an omitted one, since they are immutable or type-gated:
   - Ports and selector are skipped for ExternalName (a pure DNS alias)
   - `headless: true` becomes `clusterIP: "None"`; otherwise a set `cluster_ip_address` is honored
   - `external_traffic_policy` is sent only for NodePort/LoadBalancer; `internal_traffic_policy` never for ExternalName
   - LoadBalancer-only knobs (source ranges, class, node-port allocation, health-check port) are sent only for LoadBalancer
   - Dual-stack families and policy are sent only when requested
5. **Output Export**: `outputs.go` exports all eight outputs unconditionally (empty when not applicable) so both engines flatten the identical field set

## Load Balancer Await Behavior

For `type: load_balancer`, Pulumi's await logic waits for the load-balancer ingress to be populated before the resource is considered created — so `load_balancer_ip` / `load_balancer_hostname` are reliably resolved in the outputs. A provider populates one or the other (GCP/Azure/MetalLB expose an IP; AWS ELB/NLB expose a hostname); both are exported independently, empty when absent.

## Outputs

| Output | Description |
|--------|-------------|
| `service_name` | Name of the Service object as created in the cluster |
| `namespace` | Namespace the Service was created in |
| `type` | Service type as deployed (ClusterIP, NodePort, LoadBalancer, ExternalName) |
| `cluster_ip` | Cluster-internal virtual IP; empty for headless and ExternalName services |
| `load_balancer_ip` | Provisioned load balancer IP; empty on hostname-based providers and non-LB types |
| `load_balancer_hostname` | Provisioned load balancer hostname; empty on IP-based providers and non-LB types |
| `kube_endpoint` | In-cluster DNS endpoint (`<name>.<namespace>.svc.cluster.local`) |
| `port_forward_command` | Ready-to-run `kubectl port-forward` command; empty for ExternalName services |

## Usage

```bash
# Preview changes
make preview manifest=../../e2e/manifest.yaml

# Deploy
make up manifest=../../e2e/manifest.yaml

# Destroy
make down manifest=../../e2e/manifest.yaml
```

## Debug

```bash
# Build the module
go build ./module/...

# Build the entrypoint
go build .
```

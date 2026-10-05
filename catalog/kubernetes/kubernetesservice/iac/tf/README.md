# Kubernetes Service - Terraform Module

## Overview

This Terraform module creates and manages a standalone Kubernetes Service. It supports the core/v1 ServiceSpec surface: all four service types (ClusterIP, NodePort, LoadBalancer, ExternalName), headless services, traffic policies, session affinity, LoadBalancer tuning knobs, and dual-stack addressing.

**Known limitation — `traffic_distribution`:** the Terraform kubernetes provider (v3.2.x) does not expose `spec.trafficDistribution`, so only the Pulumi engine can apply that field. Rather than silently dropping a set field, this module fails the plan loudly via a lifecycle precondition:

> `traffic_distribution is not supported by the Terraform kubernetes provider; deploy this manifest with the Pulumi engine or unset the field.`

Every other spec field behaves identically across both engines.

## Architecture

```
iac/tf/
├── provider.tf     # Terraform and Kubernetes provider requirements
├── variables.tf    # Input variables mirroring spec.proto
├── locals.tf       # Derived values: labels, namespace default, enum → API string translation
├── main.tf         # Creates kubernetes_service_v1 resource
├── outputs.tf      # Exports the eight outputs
└── README.md       # This file
```

## How It Works

1. **Variable Input**: The `spec` variable mirrors `spec.proto`; enum fields arrive as proto value names (`"load_balancer"`, `"client_ip"`, `"prefer_dual_stack"`), and the `StringValueOrRef` namespace arrives flattened to a plain string
2. **Enum Translation**: `locals.tf` maps proto value names to Kubernetes API strings (`load_balancer` → `LoadBalancer`, `internal_local` → `Local`, `ipv6` → `IPv6`, ...) once, so the resource and the outputs agree on wire values
3. **Namespace Default**: falls back to the `default` namespace when none is provided
4. **Label Merging**: Standard Planton identity labels are merged with user labels (identity keys win on conflict)
5. **Resource Creation**: `main.tf` creates a single `kubernetes_service_v1` resource, sending each optional field only when the user set it — for several fields (clusterIP, healthCheckNodePort, loadBalancerClass) an empty value is not the same as an omitted one, since they are immutable or type-gated:
   - Ports and selector are skipped for ExternalName (a pure DNS alias)
   - `headless: true` becomes `clusterIP: "None"`; otherwise a set `cluster_ip_address` is honored
   - `external_traffic_policy` is sent only for NodePort/LoadBalancer; `internal_traffic_policy` never for ExternalName
   - LoadBalancer-only knobs (source ranges, class, node-port allocation, health-check port) are sent only for LoadBalancer
   - Dual-stack families and policy are sent only when requested
6. **Output Export**: all eight outputs are exported unconditionally (empty when not applicable) so both engines flatten the identical field set onto the outputs proto

## Usage

```hcl
module "service" {
  source = "./iac/tf"

  metadata = {
    name = "public-web"
  }

  spec = {
    name      = "public-web"
    namespace = "production"

    type = "load_balancer"

    annotations = {
      "service.beta.kubernetes.io/aws-load-balancer-type" = "nlb"
    }

    selector = {
      app = "web"
    }

    ports = [{
      name        = "https"
      port        = 443
      target_port = "8443"
    }]

    external_traffic_policy    = "local"
    load_balancer_source_ranges = ["203.0.113.0/24"]
  }
}
```

## Inputs

| Name | Description | Type | Required |
|------|-------------|------|----------|
| `metadata` | Resource metadata (name, id, org, env) | object | yes |
| `spec` | Service specification mirroring `spec.proto` (enum values as proto value names) | object | yes |

## Outputs

| Name | Description |
|------|-------------|
| `service_name` | Name of the Service object as created in the cluster |
| `namespace` | Namespace the Service was created in |
| `type` | Service type as deployed (ClusterIP, NodePort, LoadBalancer, ExternalName) |
| `cluster_ip` | Cluster-internal virtual IP; empty for headless and ExternalName services |
| `load_balancer_ip` | Provisioned load balancer IP; empty on hostname-based providers and non-LB types |
| `load_balancer_hostname` | Provisioned load balancer hostname; empty on IP-based providers and non-LB types |
| `kube_endpoint` | In-cluster DNS endpoint (`<name>.<namespace>.svc.cluster.local`) |
| `port_forward_command` | Ready-to-run `kubectl port-forward` command; empty for ExternalName services |

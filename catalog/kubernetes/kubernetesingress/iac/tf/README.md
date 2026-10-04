# Kubernetes Ingress - Terraform Module

## Overview

This Terraform module creates and manages a Kubernetes `networking/v1` Ingress. It supports the complete IngressSpec surface: ingress class selection, a default backend, TLS termination blocks, and host/path rules with all three path types.

## Architecture

```
iac/tf/
├── provider.tf     # Terraform and Kubernetes provider requirements
├── variables.tf    # Input variables mirroring spec.proto
├── locals.tf       # Derived values: labels, namespace default, path-type map, first host
├── main.tf         # Creates kubernetes_ingress_v1 resource
├── outputs.tf      # Exports ingress_name, namespace, load_balancer_ip/hostname, first_host
└── README.md       # This file
```

## How It Works

1. **Variable Input**: The `spec` variable mirrors the protobuf schema. `StringValueOrRef` fields (namespace, backend service names, TLS secret names) arrive flattened to plain strings; `path_type` arrives as the proto enum value name (`prefix`, `exact`, `implementation_specific`)
2. **Namespace Default**: `locals.tf` falls back to the `default` namespace when none is provided
3. **Label Merging**: Standard Planton labels are merged with user labels; identity keys cannot be overridden
4. **Path Type Mapping**: Proto enum names translate to the Kubernetes API strings (`prefix` → `Prefix`, etc.); an unset path type defaults to `Prefix`
5. **Resource Creation**: `main.tf` creates a single `kubernetes_ingress_v1` resource with the class, default backend, TLS blocks, and rules
6. **Output Export**: Name, namespace, load-balancer handles, and first host are exported

## Non-Blocking Creation (wait_for_load_balancer = false)

The resource sets `wait_for_load_balancer = false`, so creation never blocks on an ingress controller claiming the object. An Ingress is valid without a controller — Infra Charts routinely deploy the workload and its exposure before the ingress controller wave — and blocking every deploy until a controller populates the load-balancer status would couple this module to cluster addon ordering. The Pulumi module's `skipAwait` annotation is the exact same choice.

Consequence: the `load_balancer_ip`/`load_balancer_hostname` outputs are `try()`-guarded reads of the object's status. They export empty on a cluster where no controller has reconciled the Ingress yet, and fill in once one has.

## Usage

```hcl
module "ingress" {
  source = "./iac/tf"

  metadata = {
    name = "web-ingress"
  }

  spec = {
    name               = "web-ingress"
    namespace          = "web"
    ingress_class_name = "nginx"

    tls = [{
      hosts       = ["app.example.com"]
      secret_name = "app-example-com-tls"
    }]

    rules = [{
      host = "app.example.com"
      paths = [{
        path      = "/"
        path_type = "prefix"
        backend = {
          service_name = "web-svc"
          port_number  = 8080
        }
      }]
    }]
  }
}
```

## Inputs

| Name | Description | Type | Required |
|------|-------------|------|----------|
| `metadata` | Resource metadata (name, org, env) | object | yes |
| `spec` | Ingress specification (class, default backend, TLS, rules) | object | yes |

## Outputs

| Name | Description |
|------|-------------|
| `ingress_name` | Name of the Ingress object as created in the cluster |
| `namespace` | Namespace the Ingress was created in |
| `load_balancer_ip` | IP the controller's load balancer exposes; empty until a controller reconciles the Ingress |
| `load_balancer_hostname` | Hostname the controller's load balancer exposes; empty until a controller reconciles the Ingress |
| `first_host` | First host declared in the rules — the primary public FQDN this Ingress serves |

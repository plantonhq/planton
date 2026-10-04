# Kubernetes Ingress - Pulumi Module

## Overview

This Pulumi module creates and manages a Kubernetes `networking/v1` Ingress. It supports the complete IngressSpec surface: ingress class selection, a default backend, TLS termination blocks, and host/path rules with all three path types.

## Architecture

```
iac/pulumi/
├── main.go          # Entrypoint: loads IaC input, calls module
├── Pulumi.yaml      # Pulumi project configuration
├── Makefile         # Make targets for preview/up/down/refresh
└── module/
    ├── main.go      # Orchestrator: provider init, resource creation, output export
    ├── locals.go    # Derived values: labels, annotations, namespace default, first host
    ├── ingress.go   # Creates kubernetes.networking.v1.Ingress resource
    └── outputs.go   # Exports ingress_name, namespace, load_balancer_ip/hostname, first_host
```

## How It Works

1. **IaC Input Loading**: The entrypoint loads `KubernetesIngressIacInput` from Pulumi config
2. **Locals Initialization**: `locals.go` computes:
   - Standard Planton labels merged with user labels (identity keys cannot be overridden)
   - User annotations (controller-specific behavior: rewrites, cert-manager issuers, ...)
   - The target namespace (foreign-key references are pre-resolved; falls back to `default` when omitted)
   - The first host declared in the rules, for the `first_host` output
3. **Provider Creation**: Kubernetes provider is initialized from `provider_config`
4. **Ingress Creation**: A single `kubernetes.networking.v1.Ingress` is created with the class, default backend, TLS blocks, and rules. `StringValueOrRef` fields (backend `service_name`, TLS `secret_name`) arrive resolved to literal names
5. **Output Export**: Name, namespace, load-balancer handles, and first host are exported as outputs

## Non-Blocking Creation (skipAwait)

The module adds the `pulumi.com/skipAwait: "true"` annotation to the created Ingress — Pulumi engine metadata, added on top of user annotations so the user-facing annotation set stays exactly what the spec declared. Without it, Pulumi's default behavior is to **wait** for the Ingress to receive a load-balancer address, which hangs forever on clusters where no ingress controller has claimed the object yet.

An Ingress object is valid without a controller — Infra Charts routinely deploy the workload and its exposure before the ingress controller wave — so creation deliberately never blocks on one. The Terraform module's `wait_for_load_balancer = false` is the exact same choice. Consequence: `load_balancer_ip`/`load_balancer_hostname` export empty until a controller reconciles the object (the status reads in `outputs.go` are nil-tolerant), and fill in on a later refresh once one has.

## Backend Handling

Each backend (default and per-path) sets exactly one of port number or port name — the spec's CEL rule guarantees it, so the two branches in `buildBackend` are exhaustive. Path types map from the proto enum to the Kubernetes API strings: `prefix` → `Prefix` (also the default for unset), `exact` → `Exact`, `implementation_specific` → `ImplementationSpecific`.

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

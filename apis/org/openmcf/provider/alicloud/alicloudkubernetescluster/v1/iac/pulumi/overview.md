# Overview

This Pulumi module deploys an Alibaba Cloud ACK Managed Kubernetes cluster from a single YAML manifest conforming to OpenMCF's Kubernetes-like resource model (`apiVersion`, `kind`, `metadata`, `spec`, `status`). The module provisions a single `cs.ManagedKubernetes` resource with support for both Flannel and Terway CNI modes, cluster addons, control plane logging, maintenance windows, and automatic version upgrades.

---

## Module Architecture

```
iac/pulumi/
├── main.go           # Entrypoint: loads stack input, calls module.Resources
├── Pulumi.yaml       # Pulumi project definition
└── module/
    ├── main.go       # Controller: creates provider, builds ManagedKubernetesArgs, provisions cluster
    ├── locals.go     # Locals: tag merging, default resolution for optional fields
    └── outputs.go    # Output constants: 11 exported values
```

**Control flow**: `main.go` deserializes the protobuf stack input → `module.Resources` initializes locals → builds the `ManagedKubernetesArgs` struct → conditionally sets optional fields → creates the `cs.ManagedKubernetes` resource → exports outputs.

## Key Components

**Controller (`module/main.go`)**: The `Resources` function is the main orchestrator. It creates an Alibaba Cloud provider scoped to the spec region, resolves the cluster name (falling back to `metadata.name`), converts `StringValueOrRef` fields to Pulumi string arrays, builds the cluster arguments, and conditionally adds networking (pod CIDR or pod VSwitches), security (security group, enterprise SG, encryption key, custom SAN), addon, logging, maintenance window, and auto-upgrade configuration.

**Locals (`module/locals.go`)**: Handles default resolution through helper functions: `clusterSpec` (default: `ack.standard`), `proxyMode` (default: `ipvs`), `nodeCidrMask` (default: 24), `newNatGateway` (default: true), `slbInternetEnabled` (default: true), `enableRrsa` (default: false), `deletionProtection` (default: false), `controlPlaneLogTtl` (default: `"30"`). Also merges resource metadata tags with user-provided tags.

**Outputs (`module/outputs.go`)**: Defines 11 output constants exported after cluster creation: `cluster_id`, `cluster_name`, `api_server_internet`, `api_server_intranet`, `vpc_id`, `security_group_id`, `nat_gateway_id`, `worker_ram_role_name`, `rrsa_oidc_issuer_url`, `ram_oidc_provider_name`, `ram_oidc_provider_arn`.

## Resource Relationships

```
StackInput (protobuf)
  └── spec
       ├── region ──────────────> alicloud.Provider
       ├── vswitchIds ──────────> ManagedKubernetesArgs.VswitchIds
       ├── podCidr ─────────────> ManagedKubernetesArgs.PodCidr (Flannel)
       ├── podVswitchIds ───────> ManagedKubernetesArgs.PodVswitchIds (Terway)
       ├── addons ──────────────> ManagedKubernetesArgs.Addons
       ├── logging ─────────────> ManagedKubernetesArgs.ControlPlaneLog* + AuditLogConfig
       ├── maintenanceWindow ───> ManagedKubernetesArgs.MaintenanceWindow
       └── autoUpgrade ─────────> ManagedKubernetesArgs.OperationPolicy
```

## Design Decisions

- **Single resource**: The module wraps exactly one `cs.ManagedKubernetes` resource. Cluster addons, logging, maintenance, and auto-upgrade are all nested within this resource, not separate resources.
- **Conditional field setting**: Optional fields are only set when the user provides a value. This avoids sending zero values to the provider, which could override server-side defaults.
- **CNI mode separation**: Flannel and Terway are mutually exclusive. The module sets `PodCidr` only when non-empty and `PodVswitchIds` only when the list is non-empty, leaving the addon list to determine the active CNI.
- **No kubeconfig export**: The kubeconfig is intentionally not exported as a stack output due to its sensitive nature. Retrieve it through `aliyun cs` CLI or the ACK console.

## Customization Guide

| Goal | File to Modify | What to Change |
|------|---------------|----------------|
| Add a new output | `module/outputs.go` + `module/main.go` | Add a constant and `ctx.Export` call |
| Change a default value | `module/locals.go` | Update the relevant helper function |
| Add a new spec field | `module/main.go` | Add conditional field mapping in `Resources` |
| Add a new resource | `module/main.go` | Create the resource after the cluster, using cluster outputs |

---

## Next Steps

- Refer to the [README.md](./README.md) for CLI flows and debugging instructions.
- Review the [examples.md](./examples.md) for runnable manifests.

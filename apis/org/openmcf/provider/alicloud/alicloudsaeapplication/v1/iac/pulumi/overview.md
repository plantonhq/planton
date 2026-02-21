# Overview

This Pulumi module deploys an Alibaba Cloud SAE application from a single YAML manifest conforming to OpenMCF's Kubernetes-like resource model (`apiVersion`, `kind`, `metadata`, `spec`, `status`). The module provisions a single `sae.Application` resource with configurable compute tiers, VPC networking, health checks, rolling update strategy, environment variables, custom host aliases, and SLS log collection.

---

## Module Architecture

```
iac/pulumi/
├── main.go           # Entrypoint: loads stack input, calls module.Resources
├── Pulumi.yaml       # Pulumi project definition
└── module/
    ├── main.go       # Controller: creates provider, builds ApplicationArgs, provisions app
    ├── locals.go     # Locals: tag merging, envs-to-JSON conversion, optional field helpers
    └── outputs.go    # Output constants: app_id, app_name
```

**Control flow**: `main.go` deserializes the protobuf stack input → `module.Resources` initializes locals (tag merging) → builds the `sae.ApplicationArgs` struct → conditionally sets optional fields (VPC, namespace, health checks, update strategy, envs, SLS) → creates the `sae.Application` resource → exports outputs.

## Key Components

**Controller (`module/main.go`)**: The `Resources` function is the main orchestrator. It creates an Alibaba Cloud provider scoped to the spec's region, builds the `ApplicationArgs` with required fields (app name, package type, replicas, CPU, memory), then conditionally sets each optional field group: networking (VPC, VSwitch, security group), deployment (image URL, package URL, command, command args, envs), runtime (JDK, JVM options, programming language, timezone, graceful shutdown), health checks (liveness_v2, readiness_v2), custom host aliases, update strategy, SLS configs, and tags.

Helper functions handle type-specific conversions:
- `healthCheckArgs` / `readinessCheckArgs` — maps the proto `AliCloudSaeApplicationHealthCheck` to the provider's `ApplicationLivenessV2Args` / `ApplicationReadinessV2Args`
- `customHostAliasArgs` — converts the proto's repeated host alias messages to `ApplicationCustomHostAliasV2Array`
- `updateStrategyArgs` — maps the proto update strategy to `ApplicationUpdateStrategyV2Args`

**Locals (`module/locals.go`)**: Handles two responsibilities:
1. **Tag merging**: Combines standard metadata tags (`resource`, `resource_name`, `resource_kind`, `resource_id`, `organization`, `environment`) with user-provided `spec.tags`
2. **Environment variable serialization**: The `envsToJSON` function converts a `map[string]string` into the JSON array format that the SAE API expects: `[{"name":"K","value":"V"},...]`
3. **Optional field helpers**: `optionalString`, `optionalInt`, `optionalBool`, `optionalStringPtr` — return `nil` when the value is zero/empty, preventing Pulumi from sending zero values that override provider defaults

**Outputs (`module/outputs.go`)**: Defines 2 output constants exported after application creation: `app_id` (the SAE-assigned application identifier) and `app_name` (mirrors the spec input for downstream reference).

## Resource Relationships

```
StackInput (protobuf)
  └── spec
       ├── region ──────────────> alicloud.Provider
       ├── appName ─────────────> sae.ApplicationArgs.AppName
       ├── packageType ─────────> sae.ApplicationArgs.PackageType
       ├── replicas ────────────> sae.ApplicationArgs.Replicas
       ├── cpu, memory ─────────> sae.ApplicationArgs.Cpu, Memory
       ├── vpcId ───────────────> sae.ApplicationArgs.VpcId
       ├── vswitchId ───────────> sae.ApplicationArgs.VswitchId
       ├── securityGroupId ─────> sae.ApplicationArgs.SecurityGroupId
       ├── imageUrl ────────────> sae.ApplicationArgs.ImageUrl
       ├── envs ────────────────> envsToJSON → sae.ApplicationArgs.Envs
       ├── liveness ────────────> healthCheckArgs → sae.ApplicationArgs.LivenessV2
       ├── readiness ───────────> readinessCheckArgs → sae.ApplicationArgs.ReadinessV2
       ├── customHostAliases ───> customHostAliasArgs → sae.ApplicationArgs.CustomHostAliasV2s
       ├── updateStrategy ──────> updateStrategyArgs → sae.ApplicationArgs.UpdateStrategyV2
       └── slsConfigs ──────────> sae.ApplicationArgs.SlsConfigs
```

## Design Decisions

- **Single resource**: The module wraps exactly one `sae.Application` resource. All configuration (health checks, update strategy, host aliases) is nested within this resource, not split across separate resources.
- **Conditional field setting**: Optional fields are only set when the user provides a non-zero value. This avoids sending empty strings or zero integers to the provider, which could override SAE's server-side defaults.
- **Envs as map**: The proto schema exposes `envs` as `map<string, string>` for a clean YAML authoring experience. The `envsToJSON` function handles the conversion to the SAE API's JSON array format internally.
- **Health check v2 only**: The module uses `LivenessV2` and `ReadinessV2` (structured blocks) exclusively. The legacy `Liveness` and `Readiness` fields (raw JSON strings) are not supported.
- **No auto-scaling**: Scaling rules are a separate SAE resource (`alicloud_sae_application_scaling_rule`) and are not included in this module. The `replicas` field sets a fixed instance count.

## Customization Guide

| Goal | File to Modify | What to Change |
|------|---------------|----------------|
| Add a new output | `module/outputs.go` + `module/main.go` | Add a constant and `ctx.Export` call |
| Change tag behavior | `module/locals.go` | Update the `initializeLocals` function |
| Add a new spec field | `module/main.go` | Add conditional field mapping in `Resources` |
| Change env var format | `module/locals.go` | Update the `envsToJSON` function |
| Add a new resource | `module/main.go` | Create the resource after the application, using app outputs |

---

## Next Steps

- Refer to the [README.md](./README.md) for CLI flows and debugging instructions.
- Review the [examples.md](./examples.md) for runnable manifests.

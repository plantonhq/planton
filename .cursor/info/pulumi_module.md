# Pulumi Module Authoring Guide

Purpose: implement the Pulumi module under `iac/pulumi/module` for a resource kind. Do not add BUILD.bazel here; `make protos` (which runs Gazelle) will handle it.

## Inputs to read
- `api.proto`, `spec.proto`, `stack_input.proto`, `stack_outputs.proto`
- Provider credential proto referenced by `stack_input.proto`

## Target directory
- `apis/project/planton/provider/<provider>/<kindfolder>/v1/iac/pulumi/module/`

## Files (typical)
- `main.go` — controller function `Resources(ctx *pulumi.Context, in *<pkg>.<Kind>StackInput) error`
- `locals.go` — `initializeLocals(ctx, in)` returning a struct with ctx, input, target, spec, derived values
- `outputs.go` — constants for `<Kind>StackOutputs` names and helpers
- `resource_*.go` — one or more resource creators split by concern (e.g., dns.go, security_group.go)

## Controller pattern
- Initialize locals; init provider from credentials; orchestrate resources in order; guard optional flows; export outputs.

## Handling Optional Fields with Defaults

When proto fields have `(org.openmcf.shared.options.default)`, they are marked as `optional` and generate pointer types in Go (`*string`, `*int32`, etc.).

### Use Getter Methods

**ALWAYS** use getter methods to access optional fields:

```go
// WRONG: Direct field access on pointer type
l.RunnerGroup = in.Target.Spec.RunnerGroup  // May panic if nil

// CORRECT: Use getter method (safe, returns zero value if nil)
l.RunnerGroup = in.Target.Spec.GetRunnerGroup()
```

### No Defensive Coding Needed

OpenMCF middleware **guarantees** that default values are applied before the input reaches IaC modules. Therefore:

```go
// WRONG: Unnecessary defensive coding
l.RunnerGroup = in.Target.Spec.GetRunnerGroup()
if l.RunnerGroup == "" {
    l.RunnerGroup = "default"  // Don't do this!
}

// CORRECT: Trust the framework
l.RunnerGroup = in.Target.Spec.GetRunnerGroup()  // Default already applied
```

### Nested Optional Fields

For nested message fields that contain optional scalars:

```go
// Safe pattern for nested optionals
if r.Image != nil {
    l.RunnerImageRepository = r.Image.GetRepository()
    l.RunnerImageTag = r.Image.GetTag()
    l.RunnerImagePullPolicy = r.Image.GetPullPolicy()
}
// No defensive defaults needed - framework handles it
```

### Why Getter Methods?

1. **Nil-safe**: `GetFieldName()` returns zero value if receiver is nil
2. **Consistent**: Same access pattern regardless of field presence
3. **Framework guarantee**: OpenMCF middleware populates defaults

### Example locals.go Pattern

```go
func initLocals(ctx *pulumi.Context, in *InputType) (*Locals, error) {
    l := &Locals{
        Context: ctx,
        Input:   in,
        Target:  in.Target,
        Spec:    in.Target.Spec,
    }
    
    // Use getters for optional fields
    l.RunnerGroup = in.Target.Spec.GetRunnerGroup()
    l.HelmChartVersion = in.Target.Spec.GetHelmChartVersion()
    
    // Nested optional fields
    if in.Target.Spec.Runner != nil && in.Target.Spec.Runner.Image != nil {
        l.ImageRepository = in.Target.Spec.Runner.Image.GetRepository()
        l.ImageTag = in.Target.Spec.Runner.Image.GetTag()
    }
    
    return l, nil
}
```

## Notes
- Use provider SDK imports matching the provider (e.g., pulumi-aws).
- Reflect outputs aligned to `<Kind>StackOutputs`.

# Auth0Action Pulumi Module — Architecture Overview

## Resource Flow

```
Auth0ActionStackInput
  ├── target (Auth0Action manifest)
  │     ├── metadata.name → Action name
  │     └── spec
  │           ├── supported_trigger → ActionSupportedTriggersArgs
  │           ├── code → Action source code
  │           ├── runtime → Node.js version (optional)
  │           ├── deploy → Auto-deploy flag
  │           ├── dependencies → ActionDependencyArray
  │           ├── secrets → ActionSecretArray
  │           └── trigger_binding → TriggerAction (conditional)
  └── provider_config → Auth0 Provider credentials
```

## Module Structure

| File | Purpose |
|---|---|
| `main.go` | Entry point: provider setup → action → trigger binding → outputs |
| `locals.go` | Extract spec fields into `Locals` struct |
| `action.go` | Create `auth0.Action` with all config |
| `trigger_binding.go` | Conditionally create `auth0.TriggerAction` |
| `outputs.go` | Export stack outputs (id, name, version_id, runtime) |

## Resource Creation Sequence

1. **Auth0 Provider** — Configured with domain/client_id/client_secret from `provider_config`, or falls back to environment variables.
2. **auth0.Action** — Created with name, code, trigger, runtime, dependencies, and secrets. If `deploy=true`, Auth0 builds and deploys the action.
3. **auth0.TriggerAction** (conditional) — If `trigger_binding` is set, binds the deployed action to its supported trigger. Depends on the action resource.

## Outputs

| Output | Source |
|---|---|
| `id` | `action.ID()` |
| `name` | `action.Name` |
| `version_id` | `action.VersionId` |
| `runtime` | `action.Runtime` |
| `metadata_name` | From locals (metadata.name) |

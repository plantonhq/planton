# Kubernetes Secret - Pulumi Module Architecture Overview

## Design Philosophy

This module follows the principle of **simplicity through type safety**. Unlike many Kubernetes components that orchestrate multiple sub-resources (namespaces with quotas, deployments with services), KubernetesSecret creates exactly **one resource** -- the Secret itself. The complexity lies in the type mapping logic, not in resource orchestration.

## Module Structure

### `main.go` (Orchestrator)

The `Resources` function follows the standard OpenMCF Pulumi module pattern:

1. Initialize locals (derived values from spec)
2. Create Kubernetes provider from credentials
3. Create the Secret resource
4. Export stack outputs

This is intentionally lean -- a single resource creation step with no conditional branches.

### `locals.go` (Type Mapping Engine)

This is the core of the module. The `computeSecretTypeAndData` function inspects the protobuf `oneof` variant and translates it into:

- A Kubernetes secret `type` string (e.g., `"Opaque"`, `"kubernetes.io/tls"`)
- A `stringData` map with the correct keys for that type

**Special handling for DockerConfigJson:**

The DockerConfigJson variant requires constructing a `.dockerconfigjson` JSON structure from the structured fields (registry_server, username, password, email). The `buildDockerConfigJSON` function:

1. Computes a base64 `auth` token from `username:password`
2. Constructs the standard Docker config JSON with the `auths` map
3. Marshals to a JSON string

This is the only non-trivial data transformation in the module.

### `secret.go` (Resource Creator)

Creates a `kubernetes.core.v1.Secret` with:
- Metadata: name, namespace, labels, annotations
- Type: computed from the oneof variant
- StringData: computed data map
- Immutable: from spec flag

Uses `stringData` (not `data`) so values are provided as plain strings. Kubernetes handles the base64 encoding automatically.

### `outputs.go` (Stack Outputs)

Exports three outputs matching `KubernetesSecretStackOutputs`:
- `secret_name`: The secret's name
- `secret_namespace`: The namespace where it was created
- `secret_type`: The Kubernetes secret type string

## Key Design Decisions

### Why `stringData` instead of `data`?

Kubernetes Secrets accept data in two forms:
- `data`: Base64-encoded values (requires encoding/decoding)
- `stringData`: Plain text values (Kubernetes encodes on write)

We use `stringData` because:
1. Values in the spec are plain strings
2. No unnecessary encoding/decoding in the IaC module
3. Cleaner code and easier debugging

### Why construct DockerConfigJson in the module?

The `.dockerconfigjson` JSON structure is an implementation detail of the `kubernetes.io/dockerconfigjson` secret type. By constructing it in the module from structured fields, we:
1. Validate the structure at build time (typed fields vs. raw JSON)
2. Prevent malformed JSON from reaching the cluster
3. Provide a clean user experience (users provide fields, not JSON)

### Why no defensive defaults?

OpenMCF middleware guarantees defaults are applied before the IaC module runs. The `namespace` field has a default of `"default"` via the proto schema. The module trusts the framework and uses `spec.GetNamespace()` without fallback logic.

## Dependencies

- `github.com/pulumi/pulumi-kubernetes/sdk/v4` - Kubernetes provider
- `github.com/pulumi/pulumi/sdk/v3` - Pulumi SDK
- `github.com/pkg/errors` - Error wrapping
- `github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule` - OpenMCF Pulumi utilities

# KubernetesTelemetry Pulumi Module

This Pulumi module creates a namespaced Istio `Telemetry` resource on a target cluster.

> Note: unlike the other Istio components, Telemetry is created via the generic
> `apiextensions.CustomResource` rather than the typed crd2pulumi SDK. crd2pulumi cannot
> faithfully type the `tracing[].customTags` map (nested object-valued `oneOf`), so the
> `spec` is built as a map from the strongly-typed proto getters instead. See
> `../../docs/README.md` section 5 for the full rationale.

## Prerequisites

- The Istio CRDs must already be installed on the cluster
  (see the `KubernetesIstioBaseCrds` component).
- A running Istio control plane (istiod) to apply the configuration
  (see the `KubernetesIstio` component). The CR applies successfully with only the
  CRDs present; it only affects telemetry where istiod and a data plane run.
- The target namespace must exist (see `KubernetesNamespace`).
- Go toolchain and the Pulumi CLI.
- Access to the target Kubernetes cluster.

## Local Development

```bash
make deps
make build
```

## Usage

### With the Planton CLI

```bash
planton pulumi up --manifest ../hack/manifest.yaml
```

### Direct Pulumi usage

The entrypoint loads the `KubernetesTelemetryStackInput` from the `STACK_INPUT_YAML_FILE`
environment variable (path to a manifest) or `STACK_INPUT_YAML` (inline YAML content):

```bash
export STACK_INPUT_YAML_FILE=../hack/manifest.yaml
pulumi up
```

## Outputs

| Output | Description |
|--------|-------------|
| `telemetry_name` | Name of the created Telemetry resource (equals `metadata.name`) |
| `namespace` | Namespace the Telemetry resource was created in |

## Module Structure

```
pulumi/
├── main.go              # Pulumi entrypoint (loads stack input)
├── Pulumi.yaml          # Pulumi project configuration
├── Makefile             # Build automation
├── README.md            # This file
├── overview.md          # Architecture overview
└── module/
    ├── main.go          # Resource creation (untyped CustomResource) + spec builders
    ├── locals.go        # Computed values + resolved foreign keys
    └── outputs.go       # Stack output constant names
```

## References

- [Istio Telemetry](https://istio.io/latest/docs/reference/config/telemetry/)
- [Pulumi Kubernetes Provider](https://www.pulumi.com/registry/packages/kubernetes/)

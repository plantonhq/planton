# KubernetesTelemetry Pulumi Module

This Pulumi module creates a namespaced Istio `Telemetry` resource on a target cluster.

> Note: unlike the other Istio kinds, Telemetry is created via the generic
> `apiextensions.CustomResource` rather than the typed crd2pulumi SDK. crd2pulumi cannot
> faithfully type the `tracing[].customTags` map (nested object-valued `oneOf`), so the
> `spec` is built as a map from the strongly-typed proto getters instead. See
> `../../GUIDE.md` section 5 for the full rationale.

## Prerequisites

- The Istio CRDs must already be installed on the cluster
  (see the `KubernetesIstioBaseCrds` kind).
- A running Istio control plane (istiod) to apply the configuration
  (see the `KubernetesIstio` kind). The CR applies successfully with only the
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
planton pulumi up --manifest ../../e2e/manifest.yaml
```

### Direct Pulumi usage

The entrypoint loads the `KubernetesTelemetryIacInput` from the `IAC_INPUT_YAML_FILE`
environment variable (path to the IaC input, with the manifest under `target`) or `IAC_INPUT_YAML` (inline YAML content). The CLI builds that input from a manifest
and runs Pulumi:

```bash
planton pulumi up --manifest ../../e2e/manifest.yaml --stack <org>/<project>/<stack>
```

## Outputs

| Output | Description |
|--------|-------------|
| `telemetry_name` | Name of the created Telemetry resource (equals `metadata.name`) |
| `namespace` | Namespace the Telemetry resource was created in |

## Module Structure

```
pulumi/
├── main.go              # Pulumi entrypoint (loads IaC input)
├── Pulumi.yaml          # Pulumi project configuration
├── Makefile             # Build automation
├── README.md            # This file
├── overview.md          # Architecture overview
└── module/
    ├── main.go          # Resource creation (untyped CustomResource) + spec builders
    ├── locals.go        # Computed values + resolved foreign keys
    └── outputs.go       # Output constant names
```

## References

- [Istio Telemetry](https://istio.io/latest/docs/reference/config/telemetry/)
- [Pulumi Kubernetes Provider](https://www.pulumi.com/registry/packages/kubernetes/)

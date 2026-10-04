# KubernetesListenerSet Pulumi Module

This Pulumi module creates a namespaced Kubernetes Gateway API `ListenerSet` on
a target cluster using the typed crd2pulumi SDK. ListenerSet is a
standard-channel resource served as `gateway.networking.k8s.io/v1` (standard
since Gateway API v1.5); it merges additional listeners into an existing
Gateway that has opted in via `allowed_listeners`.

## Prerequisites

- The Gateway API CRDs (v1.5.0 or newer) must already be installed on the
  cluster (see the `KubernetesGatewayApiCrds` kind).
- A parent `Gateway` whose `allowed_listeners` permits attachment from this
  ListenerSet's namespace (see `KubernetesGateway`).
- The target namespace must exist (see `KubernetesNamespace`).
- For HTTPS listeners, the TLS Secret(s) referenced by `certificateRefs`
  (typically produced by a `KubernetesCertificate`).
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

The entrypoint loads the `KubernetesListenerSetIacInput` from the
`IAC_INPUT_YAML_FILE` environment variable (path to the IaC input, with the manifest under `target`) or
`IAC_INPUT_YAML` (inline YAML content). The CLI builds that input from a manifest
and runs Pulumi:

```bash
planton pulumi up --manifest ../../e2e/manifest.yaml --stack <org>/<project>/<stack>
```

## Outputs

| Output | Description |
|--------|-------------|
| `listener_set_name` | Name of the created ListenerSet (equals `metadata.name`); the target of Route `parentRefs` with `kind: ListenerSet` |
| `namespace` | Namespace the ListenerSet was created in |
| `gateway_name` | Name of the parent Gateway the listeners attach to |

## Module Structure

```
pulumi/
├── main.go              # Pulumi entrypoint (loads IaC input)
├── Pulumi.yaml          # Pulumi project configuration
├── Makefile             # Build automation
├── README.md            # This file
└── module/
    ├── main.go          # Resource creation (typed NewListenerSet, v1) + parentRef mapping
    ├── locals.go        # Computed values + resolved foreign keys
    ├── outputs.go       # Output constant names
    └── listeners.go     # Listener entry + TLS + allowedRoutes mapping (shared Gateway API shapes)
```

The ListenerSet's `StringValueOrRef` foreign keys (`namespace`,
`parentRef.name`, listener `certificateRefs[].name`) arrive resolved to literal
strings in the IaC input; the module reads their final values directly. No
await/wait logic is attached: per-listener Accepted/Programmed conditions and
the parent Gateway's AttachedListenerSets count belong to the Gateway
controller's reconciliation, not to applying the resource.

## References

- [Gateway API ListenerSet](https://gateway-api.sigs.k8s.io/api-types/listenerset/)
- [Pulumi Kubernetes Provider](https://www.pulumi.com/registry/packages/kubernetes/)

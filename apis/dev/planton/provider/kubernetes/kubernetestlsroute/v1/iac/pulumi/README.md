# KubernetesTlsRoute Pulumi Module

This Pulumi module creates a namespaced Kubernetes Gateway API `TLSRoute` on a
target cluster using the typed crd2pulumi SDK (served as
`gateway.networking.k8s.io/v1`).

## Prerequisites

- The Gateway API CRDs must already be installed on the cluster
  (see the `KubernetesGatewayApiCrds` component).
- A `Gateway` the route attaches to via `parentRefs`, with a `TLS` listener
  (see `KubernetesGateway`).
- The target namespace must exist (see `KubernetesNamespace`).
- The backend Services the route forwards to.
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

The entrypoint loads the `KubernetesTlsRouteStackInput` from the
`STACK_INPUT_YAML_FILE` environment variable (path to a manifest) or
`STACK_INPUT_YAML` (inline YAML content):

```bash
export STACK_INPUT_YAML_FILE=../hack/manifest.yaml
pulumi up
```

## Debug

```bash
bash debug.sh ../hack/manifest.yaml
```

## Outputs

| Output | Description |
|--------|-------------|
| `route_name` | Name of the created TLSRoute (equals `metadata.name`) |
| `namespace` | Namespace the TLSRoute was created in |

## Module Structure

```
pulumi/
├── main.go              # Pulumi entrypoint (loads stack input)
├── Pulumi.yaml          # Pulumi project configuration
├── Makefile             # Build automation
├── README.md            # This file
├── overview.md          # Architecture overview
├── debug.sh             # Local preview helper
└── module/
    ├── main.go          # Resource creation (typed NewTLSRoute)
    ├── locals.go        # Computed values + resolved foreign keys
    ├── outputs.go       # Stack output constant names
    ├── parent_refs.go   # parentRefs (attached Gateways) mapping
    └── rules.go         # Rule + backend ref mapping (no matches/filters for TLSRoute)
```

## References

- [Gateway API TLSRoute](https://gateway-api.sigs.k8s.io/api-types/tlsroute/)
- [Pulumi Kubernetes Provider](https://www.pulumi.com/registry/packages/kubernetes/)

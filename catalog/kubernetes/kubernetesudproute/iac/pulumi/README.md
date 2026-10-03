# KubernetesUdpRoute Pulumi Module

This Pulumi module creates a namespaced Kubernetes Gateway API `UDPRoute` on a
target cluster using the typed crd2pulumi SDK. UDPRoute is a GA standard-channel
resource served as `gateway.networking.k8s.io/v1`.

## Prerequisites

- The Gateway API standard-channel CRDs must already be installed on the
  cluster (`KubernetesGatewayApiCrds`).
- A `Gateway` the route attaches to via `parentRefs`, with a `UDP` listener
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
planton pulumi up --manifest ../../e2e/manifest.yaml
```

### Direct Pulumi usage

The entrypoint loads the `KubernetesUdpRouteIacInput` from the
`IAC_INPUT_YAML_FILE` environment variable (path to a manifest) or
`IAC_INPUT_YAML` (inline YAML content):

```bash
export IAC_INPUT_YAML_FILE=../../e2e/manifest.yaml
pulumi up
```

## Outputs

| Output | Description |
|--------|-------------|
| `route_name` | Name of the created UDPRoute (equals `metadata.name`) |
| `namespace` | Namespace the UDPRoute was created in |

## Module Structure

```
pulumi/
├── main.go              # Pulumi entrypoint (loads IaC input)
├── Pulumi.yaml          # Pulumi project configuration
├── Makefile             # Build automation
├── README.md            # This file
└── module/
    ├── main.go          # Resource creation (typed NewUDPRoute, v1)
    ├── locals.go        # Computed values + resolved foreign keys
    ├── outputs.go       # Output constant names
    ├── parent_refs.go   # parentRefs (attached Gateways) mapping
    └── rules.go         # Rule + backend ref mapping (no matches/filters for UDPRoute)
```

The route's `StringValueOrRef` foreign keys (`namespace`, `parentRefs[].name`,
`backendRefs[].name`) arrive resolved to literal strings in the IaC input;
the module reads their final values directly. No await/wait logic is attached:
Accepted/ResolvedRefs conditions belong to the Gateway controller's
reconciliation, not to applying the resource.

## References

- [Gateway API UDPRoute](https://gateway-api.sigs.k8s.io/api-types/udproute/)
- [Pulumi Kubernetes Provider](https://www.pulumi.com/registry/packages/kubernetes/)

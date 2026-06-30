# KubernetesTcpRoute Pulumi Module

This Pulumi module creates a namespaced Kubernetes Gateway API `TCPRoute` on a
target cluster using the typed crd2pulumi SDK. TCPRoute is an experimental-channel
resource served as `gateway.networking.k8s.io/v1alpha2`.

## Prerequisites

- The Gateway API **experimental-channel** CRDs must already be installed on the
  cluster (`KubernetesGatewayApiCrds` with `install_channel: experimental`). The
  standard channel has no TCPRoute CRD.
- A `Gateway` the route attaches to via `parentRefs`, with a `TCP` listener
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

The entrypoint loads the `KubernetesTcpRouteStackInput` from the
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
| `route_name` | Name of the created TCPRoute (equals `metadata.name`) |
| `namespace` | Namespace the TCPRoute was created in |

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
    ├── main.go          # Resource creation (typed NewTCPRoute, v1alpha2)
    ├── locals.go        # Computed values + resolved foreign keys
    ├── outputs.go       # Stack output constant names
    ├── parent_refs.go   # parentRefs (attached Gateways) mapping
    └── rules.go         # Rule + backend ref mapping (no matches/filters for TCPRoute)
```

## References

- [Gateway API TCPRoute](https://gateway-api.sigs.k8s.io/api-types/tcproute/)
- [Pulumi Kubernetes Provider](https://www.pulumi.com/registry/packages/kubernetes/)

# KubernetesGateway Pulumi Module

This Pulumi module creates a namespaced Kubernetes Gateway API `Gateway` on a
target cluster using the typed crd2pulumi SDK.

## Prerequisites

- The Gateway API CRDs must already be installed on the cluster
  (see the `KubernetesGatewayApiCrds` component).
- A `GatewayClass` whose `controllerName` resolves to an installed controller
  (Istio, Envoy Gateway, NGINX, ...). See `KubernetesGatewayClass`.
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

The entrypoint loads the `KubernetesGatewayStackInput` from the
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
| `gateway_name` | Name of the created Gateway (equals `metadata.name`); the target of Route `parentRefs` |
| `namespace` | Namespace the Gateway was created in |
| `gateway_class_name` | Name of the GatewayClass this Gateway belongs to |

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
    ├── main.go          # Resource creation (typed NewGateway)
    ├── locals.go        # Computed values + resolved foreign keys
    ├── outputs.go       # Stack output constant names
    ├── listeners.go     # Listener + listener-TLS + allowedRoutes mapping
    ├── tls.go           # Gateway-level frontend/backend TLS mapping
    ├── infrastructure.go# Infrastructure + allowedListeners mapping
    ├── addresses.go     # Requested-address mapping
    └── selectors.go     # Namespace label-selector mapping
```

## References

- [Gateway API Gateway](https://gateway-api.sigs.k8s.io/api-types/gateway/)
- [Pulumi Kubernetes Provider](https://www.pulumi.com/registry/packages/kubernetes/)

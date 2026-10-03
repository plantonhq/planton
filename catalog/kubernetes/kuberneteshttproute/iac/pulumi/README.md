# KubernetesHttpRoute Pulumi Module

This Pulumi module creates a namespaced Kubernetes Gateway API `HTTPRoute` on a
target cluster using the typed crd2pulumi SDK.

## Prerequisites

- The Gateway API CRDs must already be installed on the cluster
  (see the `KubernetesGatewayApiCrds` component).
- A `Gateway` the route attaches to via `parentRefs` (see `KubernetesGateway`).
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

The entrypoint loads the `KubernetesHttpRouteIacInput` from the
`IAC_INPUT_YAML_FILE` environment variable (path to a manifest) or
`IAC_INPUT_YAML` (inline YAML content):

```bash
export IAC_INPUT_YAML_FILE=../../e2e/manifest.yaml
pulumi up
```

## Outputs

| Output | Description |
|--------|-------------|
| `route_name` | Name of the created HTTPRoute (equals `metadata.name`) |
| `namespace` | Namespace the HTTPRoute was created in |

## Module Structure

```
pulumi/
├── main.go              # Pulumi entrypoint (loads IaC input)
├── Pulumi.yaml          # Pulumi project configuration
├── Makefile             # Build automation
├── README.md            # This file
└── module/
    ├── main.go          # Resource creation (typed NewHTTPRoute)
    ├── locals.go        # Computed values + resolved foreign keys
    ├── outputs.go       # Output constant names
    ├── parent_refs.go   # parentRefs (attached Gateways) mapping
    ├── rules.go         # Rule + timeouts mapping
    ├── matches.go       # Path / header / query-param / method match mapping
    ├── filters.go       # Rule-level filter mapping (header/redirect/rewrite/mirror/CORS)
    └── backend_refs.go  # Backend ref + backend-level filter mapping
```

The route's `StringValueOrRef` foreign keys (`namespace`, `parentRefs[].name`,
`backendRefs[].name`) arrive resolved to literal strings in the IaC input;
the module reads their final values directly. No await/wait logic is attached:
Accepted/ResolvedRefs conditions belong to the Gateway controller's
reconciliation, not to applying the resource.

## References

- [Gateway API HTTPRoute](https://gateway-api.sigs.k8s.io/api-types/httproute/)
- [Pulumi Kubernetes Provider](https://www.pulumi.com/registry/packages/kubernetes/)

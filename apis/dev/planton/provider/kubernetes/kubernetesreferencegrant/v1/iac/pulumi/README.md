# KubernetesReferenceGrant Pulumi Module

This Pulumi module creates a namespaced Kubernetes Gateway API `ReferenceGrant` on
a target cluster using the typed crd2pulumi SDK (served as
`gateway.networking.k8s.io/v1`). A ReferenceGrant authorizes resources in other
namespaces to reference specified kinds of resources in this grant's namespace.

## Prerequisites

- The Gateway API CRDs must already be installed on the cluster
  (see the `KubernetesGatewayApiCrds` component).
- The target namespace must exist (see `KubernetesNamespace`). This is the "to"
  namespace -- the one whose resources the grant authorizes inbound references to.
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

The entrypoint loads the `KubernetesReferenceGrantStackInput` from the
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
| `reference_grant_name` | Name of the created ReferenceGrant (equals `metadata.name`) |
| `namespace` | Namespace the ReferenceGrant was created in |

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
    ├── main.go          # Resource creation (typed NewReferenceGrant)
    ├── locals.go        # Computed values + resolved foreign keys
    ├── outputs.go       # Stack output constant names
    └── references.go    # from (trusted sources) + to (referenceable targets) mapping
```

## References

- [Gateway API ReferenceGrant](https://gateway-api.sigs.k8s.io/api-types/referencegrant/)
- [Pulumi Kubernetes Provider](https://www.pulumi.com/registry/packages/kubernetes/)

# KubernetesReferenceGrant Pulumi Module

This Pulumi module creates a namespaced Kubernetes Gateway API `ReferenceGrant` on
a target cluster using the typed crd2pulumi SDK (served as
`gateway.networking.k8s.io/v1`). A ReferenceGrant authorizes resources in other
namespaces to reference specified kinds of resources in this grant's namespace.

## Prerequisites

- The Gateway API CRDs must already be installed on the cluster
  (see the `KubernetesGatewayApiCrds` kind).
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
planton pulumi up --manifest ../../e2e/manifest.yaml
```

### Direct Pulumi usage

The entrypoint loads the `KubernetesReferenceGrantIacInput` from the
`IAC_INPUT_YAML_FILE` environment variable (path to the IaC input, with the manifest under `target`) or
`IAC_INPUT_YAML` (inline YAML content). The CLI builds that input from a manifest
and runs Pulumi:

```bash
planton pulumi up --manifest ../../e2e/manifest.yaml --stack <org>/<project>/<stack>
```

## Outputs

| Output | Description |
|--------|-------------|
| `reference_grant_name` | Name of the created ReferenceGrant (equals `metadata.name`) |
| `namespace` | Namespace the ReferenceGrant was created in |

## Module Structure

```
pulumi/
├── main.go              # Pulumi entrypoint (loads IaC input)
├── Pulumi.yaml          # Pulumi project configuration
├── Makefile             # Build automation
├── README.md            # This file
└── module/
    ├── main.go          # Resource creation (typed NewReferenceGrant)
    ├── locals.go        # Computed values + resolved foreign keys
    ├── outputs.go       # Output constant names
    └── references.go    # from (trusted sources) + to (referenceable targets) mapping
```

## References

- [Gateway API ReferenceGrant](https://gateway-api.sigs.k8s.io/api-types/referencegrant/)
- [Pulumi Kubernetes Provider](https://www.pulumi.com/registry/packages/kubernetes/)

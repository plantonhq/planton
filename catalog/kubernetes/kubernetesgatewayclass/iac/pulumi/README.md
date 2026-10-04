# KubernetesGatewayClass Pulumi Module

This Pulumi module creates a cluster-scoped Kubernetes Gateway API `GatewayClass`
on a target cluster using the typed crd2pulumi SDK.

## Prerequisites

- The Gateway API CRDs must already be installed on the cluster
  (see the `KubernetesGatewayApiCrds` kind).
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

The entrypoint loads the `KubernetesGatewayClassIacInput` from the
`IAC_INPUT_YAML_FILE` environment variable (path to a manifest) or
`IAC_INPUT_YAML` (inline YAML content):

```bash
export IAC_INPUT_YAML_FILE=../../e2e/manifest.yaml
pulumi up
```

## Outputs

| Output | Description |
|--------|-------------|
| `gateway_class_name` | Name of the created GatewayClass (equals `metadata.name`) |
| `controller_name` | The controller managing this class |

## Module Structure

```
pulumi/
├── main.go           # Pulumi entrypoint (loads IaC input)
├── Pulumi.yaml       # Pulumi project configuration
├── Makefile          # Build automation
├── README.md         # This file
└── module/
    ├── main.go       # Resource creation (typed NewGatewayClass)
    ├── locals.go     # Computed values
    └── outputs.go    # Output constant names
```

No await/wait logic is attached: the Accepted condition belongs to the named
controller's reconciliation, not to applying the resource.

## References

- [Gateway API GatewayClass](https://gateway-api.sigs.k8s.io/api-types/gatewayclass/)
- [Pulumi Kubernetes Provider](https://www.pulumi.com/registry/packages/kubernetes/)

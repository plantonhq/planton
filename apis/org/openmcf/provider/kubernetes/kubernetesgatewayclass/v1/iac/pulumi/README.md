# KubernetesGatewayClass Pulumi Module

This Pulumi module creates a cluster-scoped Kubernetes Gateway API `GatewayClass`
on a target cluster using the typed crd2pulumi SDK.

## Prerequisites

- The Gateway API CRDs must already be installed on the cluster
  (see the `KubernetesGatewayApiCrds` component).
- Go toolchain and the Pulumi CLI.
- Access to the target Kubernetes cluster.

## Local Development

```bash
make deps
make build
```

## Usage

### With the OpenMCF CLI

```bash
openmcf pulumi up --manifest ../hack/manifest.yaml
```

### Direct Pulumi usage

The entrypoint loads the `KubernetesGatewayClassStackInput` from the
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
| `gateway_class_name` | Name of the created GatewayClass (equals `metadata.name`) |
| `controller_name` | The controller managing this class |

## Module Structure

```
pulumi/
├── main.go           # Pulumi entrypoint (loads stack input)
├── Pulumi.yaml       # Pulumi project configuration
├── Makefile          # Build automation
├── README.md         # This file
├── overview.md       # Architecture overview
├── debug.sh          # Local preview helper
└── module/
    ├── main.go       # Resource creation (typed NewGatewayClass)
    ├── locals.go     # Computed values
    └── outputs.go    # Stack output constant names
```

## References

- [Gateway API GatewayClass](https://gateway-api.sigs.k8s.io/api-types/gatewayclass/)
- [Pulumi Kubernetes Provider](https://www.pulumi.com/registry/packages/kubernetes/)

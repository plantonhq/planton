# KubernetesRequestAuthentication Pulumi Module

This Pulumi module creates a namespaced Istio `RequestAuthentication` on a target
cluster using the typed crd2pulumi SDK.

## Prerequisites

- The Istio CRDs must already be installed on the cluster
  (see the `KubernetesIstioBaseCrds` component).
- A running Istio control plane (istiod) to enforce the policy in the data plane
  (see the `KubernetesIstio` component). The CR applies successfully with only the
  CRDs present; enforcement requires istiod.
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

The entrypoint loads the `KubernetesRequestAuthenticationStackInput` from the
`STACK_INPUT_YAML_FILE` environment variable (path to a manifest) or
`STACK_INPUT_YAML` (inline YAML content):

```bash
export STACK_INPUT_YAML_FILE=../hack/manifest.yaml
pulumi up
```

## Outputs

| Output | Description |
|--------|-------------|
| `request_authentication_name` | Name of the created RequestAuthentication (equals `metadata.name`) |
| `namespace` | Namespace the RequestAuthentication was created in |

## Module Structure

```
pulumi/
├── main.go              # Pulumi entrypoint (loads stack input)
├── Pulumi.yaml          # Pulumi project configuration
├── Makefile             # Build automation
├── README.md            # This file
├── overview.md          # Architecture overview
└── module/
    ├── main.go          # Resource creation (typed NewRequestAuthentication)
    ├── locals.go        # Computed values + resolved foreign keys
    └── outputs.go       # Stack output constant names
```

## References

- [Istio RequestAuthentication](https://istio.io/latest/docs/reference/config/security/request_authentication/)
- [Pulumi Kubernetes Provider](https://www.pulumi.com/registry/packages/kubernetes/)

# KubernetesAuthorizationPolicy Pulumi Module

This Pulumi module creates a namespaced Istio `AuthorizationPolicy` on a target cluster
using the typed crd2pulumi SDK.

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

The entrypoint loads the `KubernetesAuthorizationPolicyStackInput` from the
`STACK_INPUT_YAML_FILE` environment variable (path to a manifest) or
`STACK_INPUT_YAML` (inline YAML content):

```bash
export STACK_INPUT_YAML_FILE=../hack/manifest.yaml
pulumi up
```

## Outputs

| Output | Description |
|--------|-------------|
| `authorization_policy_name` | Name of the created AuthorizationPolicy (equals `metadata.name`) |
| `namespace` | Namespace the AuthorizationPolicy was created in |

## Module Structure

```
pulumi/
├── main.go              # Pulumi entrypoint (loads stack input)
├── Pulumi.yaml          # Pulumi project configuration
├── Makefile             # Build automation
├── README.md            # This file
├── overview.md          # Architecture overview
└── module/
    ├── main.go          # Resource creation (typed NewAuthorizationPolicy)
    ├── locals.go        # Computed values + resolved foreign keys
    └── outputs.go       # Stack output constant names
```

## References

- [Istio AuthorizationPolicy](https://istio.io/latest/docs/reference/config/security/authorization-policy/)
- [Pulumi Kubernetes Provider](https://www.pulumi.com/registry/packages/kubernetes/)

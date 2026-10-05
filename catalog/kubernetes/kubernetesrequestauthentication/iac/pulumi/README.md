# KubernetesRequestAuthentication Pulumi Module

This Pulumi module creates a namespaced Istio `RequestAuthentication` on a target
cluster using the typed crd2pulumi SDK.

## Prerequisites

- The Istio CRDs must already be installed on the cluster
  (see the `KubernetesIstioBaseCrds` kind).
- A running Istio control plane (istiod) to enforce the policy in the data plane
  (see the `KubernetesIstio` kind). The CR applies successfully with only the
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
planton pulumi up --manifest ../../e2e/manifest.yaml
```

### Direct Pulumi usage

The entrypoint loads the `KubernetesRequestAuthenticationIacInput` from the
`IAC_INPUT_YAML_FILE` environment variable (path to the IaC input, with the manifest under `target`) or
`IAC_INPUT_YAML` (inline YAML content). The CLI builds that input from a manifest
and runs Pulumi:

```bash
planton pulumi up --manifest ../../e2e/manifest.yaml --stack <org>/<project>/<stack>
```

## Outputs

| Output | Description |
|--------|-------------|
| `request_authentication_name` | Name of the created RequestAuthentication (equals `metadata.name`) |
| `namespace` | Namespace the RequestAuthentication was created in |

## Module Structure

```
pulumi/
├── main.go              # Pulumi entrypoint (loads IaC input)
├── Pulumi.yaml          # Pulumi project configuration
├── Makefile             # Build automation
├── README.md            # This file
├── overview.md          # Architecture overview
└── module/
    ├── main.go          # Resource creation (typed NewRequestAuthentication)
    ├── locals.go        # Computed values + resolved foreign keys
    └── outputs.go       # Output constant names
```

## References

- [Istio RequestAuthentication](https://istio.io/latest/docs/reference/config/security/request_authentication/)
- [Pulumi Kubernetes Provider](https://www.pulumi.com/registry/packages/kubernetes/)

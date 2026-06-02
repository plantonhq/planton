# KubernetesDestinationRule Pulumi Module

This Pulumi module creates a namespaced Istio `DestinationRule` on a target cluster using
the typed crd2pulumi SDK.

## Prerequisites

- The Istio CRDs must already be installed on the cluster
  (see the `KubernetesIstioBaseCrds` component).
- A running Istio control plane (istiod) to apply the policy
  (see the `KubernetesIstio` component). The CR applies successfully with only the
  CRDs present; the policy only affects traffic where istiod and a data plane run.
- The target namespace must exist (see `KubernetesNamespace`).
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

The entrypoint loads the `KubernetesDestinationRuleStackInput` from the
`STACK_INPUT_YAML_FILE` environment variable (path to a manifest) or
`STACK_INPUT_YAML` (inline YAML content):

```bash
export STACK_INPUT_YAML_FILE=../hack/manifest.yaml
pulumi up
```

## Outputs

| Output | Description |
|--------|-------------|
| `destination_rule_name` | Name of the created DestinationRule (equals `metadata.name`) |
| `namespace` | Namespace the DestinationRule was created in |

## Module Structure

```
pulumi/
├── main.go              # Pulumi entrypoint (loads stack input)
├── Pulumi.yaml          # Pulumi project configuration
├── Makefile             # Build automation
├── README.md            # This file
├── overview.md          # Architecture overview
└── module/
    ├── main.go          # Resource creation (typed NewDestinationRule) + subsets
    ├── traffic_policy.go # Per-path typed builders for the traffic-policy subtree
    ├── locals.go        # Computed values + resolved foreign keys
    └── outputs.go       # Stack output constant names
```

## References

- [Istio DestinationRule](https://istio.io/latest/docs/reference/config/networking/destination-rule/)
- [Pulumi Kubernetes Provider](https://www.pulumi.com/registry/packages/kubernetes/)

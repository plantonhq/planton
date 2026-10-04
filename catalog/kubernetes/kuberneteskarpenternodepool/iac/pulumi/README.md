# KubernetesKarpenterNodePool Pulumi Module

This Pulumi module creates a cluster-scoped Karpenter `NodePool`
(`karpenter.sh/v1`) on a target cluster using the typed crd2pulumi SDK
(`karpenterv1.NewNodePool`). The typed approach catches field-name and
structure errors at compile time. The NodePool is named after
`metadata.name` — there is no namespace anywhere.

## Prerequisites

- A Karpenter installation on the cluster (`KubernetesKarpenter`), which
  brings the `karpenter.sh/v1` CRDs.
- The NodeClass the pool references through `template.nodeClassRef`
  (`KubernetesKarpenterEc2NodeClass` on AWS) — the pool provisions nothing
  until its machine template exists.
- Go toolchain and the Pulumi CLI.
- Access to the target Kubernetes cluster.

## Rendering Notes

- **Cluster-scoped**: the CR carries only `metadata.name` (equal to the
  Planton `metadata.name`) and the `planton.ai/*` identity labels.
- **`nodeClassRef.group` and `kind` always render**: the CRD requires all
  three keys with non-empty values, so the module applies the
  proto-declared AWS defaults (`karpenter.k8s.aws` / `EC2NodeClass`) when
  the manifest leaves them unset. `name` is a
  `KubernetesKarpenterEc2NodeClass` foreign key resolved to its literal
  value before the module runs.
- **`disruption.consolidateAfter` always renders when disruption is set**:
  the CRD marks it required inside the disruption object, so the module
  fills the proto-declared default (`0s`) when unset. The disruption block
  itself is omitted when absent (apiserver default: `consolidateAfter: 0s`,
  one always-active `10%` budget).
- **Presence-sensitive optionals are omitted when unset**: `weight`
  (absent = weight 0, but `0` is not an accepted literal), `replicas`
  (its presence switches the pool into ALPHA static-capacity mode; `0` is
  a valid static size and survives), requirement `minValues`,
  `expireAfter` (apiserver default `720h`), `terminationGracePeriod`,
  budget `schedule`/`duration`, taint `value`, and requirement `values`
  (Exists/DoesNotExist take an empty list).
- **No await/wait logic is attached**: the NodePool is configuration the
  Karpenter controller consumes; applying it server-side-validated is the
  whole contract (same as the Terraform twin's `kubectl_manifest`).

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

The entrypoint loads the `KubernetesKarpenterNodePoolIacInput` from the
`IAC_INPUT_YAML_FILE` environment variable (path to the IaC input, with the manifest under `target`) or
`IAC_INPUT_YAML` (inline YAML content). The CLI builds that input from a manifest
and runs Pulumi:

```bash
planton pulumi up --manifest ../../e2e/manifest.yaml --stack <org>/<project>/<stack>
```

## Outputs

| Output | Description |
|--------|-------------|
| `node_pool_name` | Name of the created NodePool (equals `metadata.name`) |

## Module Structure

```
pulumi/
├── main.go              # Pulumi entrypoint (loads IaC input)
├── Pulumi.yaml          # Pulumi project configuration
├── Makefile             # Build automation
├── README.md            # This file
└── module/
    ├── main.go          # Resource creation (typed NewNodePool); limits/weight/replicas mapped inline
    ├── locals.go        # Computed values (name, planton.ai/* labels)
    ├── outputs.go       # Output constant names
    ├── template.go      # NodeClaim template: nodeClassRef, requirements, taints, lifetime
    └── disruption.go    # Consolidation policy and disruption budgets
```

## References

- [Karpenter NodePool](https://karpenter.sh/docs/concepts/nodepools/)
- [Pulumi Kubernetes Provider](https://www.pulumi.com/registry/packages/kubernetes/)

# KubernetesTektonOperator

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**KubernetesTektonOperatorSpec** installs the Tekton Operator — the
lifecycle manager that turns a `TektonConfig` declaration into
running Tekton components (Pipelines, Triggers, Dashboard, Chains),
managing their installation, upgrades and removal through
`TektonInstallerSet` resources.

This component installs the MANAGER only. What Tekton actually runs
on the cluster — which components, in which namespace, with which
pipeline feature flags and pruner policy — is declared with a
KubernetesTekton resource (exactly one per cluster), which this
operator reconciles. The operator is installed with automatic
component installation DISABLED so that the KubernetesTekton
declaration is the single owner of the cluster's TektonConfig —
installing the operator alone deploys no Tekton components.

DISTRIBUTION AND VERSION: the operator's official distribution is
the single-file release manifest (`release.yaml` attached to each
GitHub release), which the modules fetch from the release tag they
are pinned to and apply per document. The in-repo Helm chart is
unpublished (version "devel") and is not a distribution channel.
The spec deliberately has NO version field: the installed operator
(and the TektonConfig schema the KubernetesTekton kind renders
against) is pinned to the release this catalog was designed
against — a user-selectable version would silently drift the
TektonConfig surface away from what the catalog models.

NAMESPACE IS FIXED: the release manifest installs into
`tekton-operator`, and that name is baked into the manifest's own
cross-references (webhook Service, RBAC subjects). Exactly ONE
install per cluster is the upstream contract.

CRD LIFECYCLE: the 14 `operator.tekton.dev` CRDs are documents of
the applied manifest — they install and are REMOVED with the
resource. Destroying the operator therefore deletes the CRDs, which
cascade-deletes any TektonConfig on the cluster. Always destroy the
KubernetesTekton resource FIRST (its deletion waits for the
operator to tear down the component installations cleanly); the
operator must still be running for that teardown to complete —
TektonInstallerSet resources carry finalizers only the operator can
process, and removing the operator first strands them.

## Example

```yaml
# Full-surface hack manifest for the offline plan/preview proofs: every
# typed override rendered at once (mirror images, both resource blocks,
# scheduling, pull secrets).
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesTektonOperator
metadata:
  name: tekton-operator-full
spec:
  operatorImage:
    repo: mirror.example.com/tektoncd/operator
    tag: v0.80.0
    pullSecretName: mirror-pull
  webhookImage:
    repo: mirror.example.com/tektoncd/operator-webhook
    tag: v0.80.0
  operatorResources:
    requests:
      cpu: 100m
      memory: 128Mi
    limits:
      cpu: 500m
      memory: 512Mi
  webhookResources:
    requests:
      cpu: 50m
      memory: 64Mi
  nodeSelector:
    role: platform
  tolerations:
    - key: platform
      operator: Exists
      effect: NoSchedule
  imagePullSecrets:
    - mirror-pull
    - extra-pull
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.operatorImage` | `ContainerImage` |  |  |  |
| `spec.operatorImage.repo` | `string` |  |  |  |
| `spec.operatorImage.tag` | `string` |  |  |  |
| `spec.operatorImage.pullSecretName` | `string` |  |  |  |
| `spec.webhookImage` | `ContainerImage` |  |  |  |
| `spec.webhookImage.repo` | `string` |  |  |  |
| `spec.webhookImage.tag` | `string` |  |  |  |
| `spec.webhookImage.pullSecretName` | `string` |  |  |  |
| `spec.operatorResources` | `ContainerResources` |  |  |  |
| `spec.operatorResources.limits` | `CpuMemory` |  |  |  |
| `spec.operatorResources.limits.cpu` | `string` |  |  |  |
| `spec.operatorResources.limits.memory` | `string` |  |  |  |
| `spec.operatorResources.requests` | `CpuMemory` |  |  |  |
| `spec.operatorResources.requests.cpu` | `string` |  |  |  |
| `spec.operatorResources.requests.memory` | `string` |  |  |  |
| `spec.webhookResources` | `ContainerResources` |  |  |  |
| `spec.webhookResources.limits` | `CpuMemory` |  |  |  |
| `spec.webhookResources.limits.cpu` | `string` |  |  |  |
| `spec.webhookResources.limits.memory` | `string` |  |  |  |
| `spec.webhookResources.requests` | `CpuMemory` |  |  |  |
| `spec.webhookResources.requests.cpu` | `string` |  |  |  |
| `spec.webhookResources.requests.memory` | `string` |  |  |  |
| `spec.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.tolerations[].key` | `string` |  |  |  |
| `spec.tolerations[].operator` | `string` |  |  |  |
| `spec.tolerations[].value` | `string` |  |  |  |
| `spec.tolerations[].effect` | `string` |  |  |  |
| `spec.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.imagePullSecrets` | `[]string` |  |  |  |
| `spec.imageRegistry` | `string` |  |  |  |

## Field Details

### spec.operatorImage

`ContainerImage`

Override the operator image (air-gap / private-mirror path).
Applied to both containers of the operator Deployment
(`tekton-operator-lifecycle` and
`tekton-operator-cluster-operations` share one image). Empty =
the release manifest's digest-pinned image
(`ghcr.io/tektoncd/operator/operator-*` at the release tag).

### spec.operatorImage.repo

`string`

The repository of the image (e.g., "gcr.io/project/image").

### spec.operatorImage.tag

`string`

The tag of the image (e.g., "latest" or "1.0.0").

### spec.operatorImage.pullSecretName

`string`

The name of the image pull secret for private image repositories.

### spec.webhookImage

`ContainerImage`

Override the operator webhook image (air-gap / private-mirror
path). Empty = the release manifest's digest-pinned image
(`ghcr.io/tektoncd/operator/webhook-*` at the release tag).

### spec.webhookImage.repo

`string`

The repository of the image (e.g., "gcr.io/project/image").

### spec.webhookImage.tag

`string`

The tag of the image (e.g., "latest" or "1.0.0").

### spec.webhookImage.pullSecretName

`string`

The name of the image pull secret for private image repositories.

### spec.operatorResources

`ContainerResources`

Resources for the operator Deployment's containers (applied to
both). Empty = the release manifest's defaults (none set — the
operator runs unbounded; set requests on production clusters
with resource quotas).

### spec.operatorResources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.operatorResources.limits.cpu

`string`

### spec.operatorResources.limits.memory

`string`

### spec.operatorResources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.operatorResources.requests.cpu

`string`

### spec.operatorResources.requests.memory

`string`

### spec.webhookResources

`ContainerResources`

Resources for the webhook container. Empty = the release
manifest's defaults (none set).

### spec.webhookResources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.webhookResources.limits.cpu

`string`

### spec.webhookResources.limits.memory

`string`

### spec.webhookResources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.webhookResources.requests.cpu

`string`

### spec.webhookResources.requests.memory

`string`

### spec.nodeSelector

`map<string, string>`

Node selector for the operator and webhook pods.

### spec.tolerations

`[]WorkloadToleration`

Tolerations for the operator and webhook pods.

### spec.tolerations[].key

`string`

Taint key to tolerate. Empty key with operator "Exists" tolerates every taint.

### spec.tolerations[].operator

`string`

How key/value match: "Equal" (default — value must match too) or "Exists"
(key presence alone matches).

- rule: Toleration operator must be either "Equal" or "Exists"

### spec.tolerations[].value

`string`

Taint value to match when operator is "Equal".

### spec.tolerations[].effect

`string`

Which taint effect is tolerated: "NoSchedule", "PreferNoSchedule", or
"NoExecute". Empty tolerates all effects for the key.

- rule: Toleration effect must be one of "NoSchedule", "PreferNoSchedule", or "NoExecute"

### spec.tolerations[].tolerationSeconds

`int64` · optional (explicit presence)

For "NoExecute" taints only: how many seconds already-running pods stay bound
after the taint appears. Unset means tolerate forever.

### spec.imagePullSecrets

`[]string`

Names of image-pull secrets (in the tekton-operator namespace)
for pulling the operator images from a private mirror.

### spec.imageRegistry

`string`

Registry that replaces the registry part of every image Tekton
publishes (all on ghcr.io) — the air-gap/private-mirror path. That
is the operator and its webhook, and every component the operator
installs: Pipelines, Triggers, Dashboard, Chains, Results and the
pruner, including the images every build pod starts with
(entrypoint, nop, workingdirinit, sidecarlogresults). Each image
keeps its path, tag and digest, so "mirror.example.com/ghcr" pulls
ghcr.io/tektoncd/pipeline/nop-...@sha256:... as
mirror.example.com/ghcr/tektoncd/pipeline/nop-...@sha256:...; a
mirror or pull-through cache of ghcr.io serves them all, and the
digest pins mean it can serve only the bytes the release names.
The images move with the kind's pinned Tekton release, so a catalog
upgrade never freezes one. Images Tekton does not publish keep
their registry: the shell images script steps start with
(cgr.dev, mcr.microsoft.com) and Results' bundled Postgres (Docker
Hub). `operator_image` and `webhook_image`, when set, win over this
for those two images. Empty = ghcr.io.

- rule: image_registry is a registry root such as "mirror.example.com/ghcr": no scheme and no trailing slash
- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesTektonOperator, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | Namespace the operator runs in — always `tekton-operator` (fixed by the release manifest). |

## See Also

- [Overview](../README.md)

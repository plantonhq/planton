# KubernetesAgentSandboxWarmPool

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

KubernetesAgentSandboxWarmPoolSpec defines an agent-sandbox SandboxWarmPool:
a number of sandboxes, built from one KubernetesAgentSandboxTemplate, kept
running and ready so a claim gets one in the time it takes to hand it over
instead of the time it takes to start a pod. The agent-sandbox controller
(KubernetesAgentSandboxController, extensions on) keeps the pool filled,
gives a ready member to each claim, and replaces what it hands out.

100% fidelity with the upstream SandboxWarmPool custom resource
(extensions.agents.x-k8s.io/v1beta1), pinned to agent-sandbox v1.0.7 -- the
release KubernetesAgentSandboxController installs -- and held to its served
definition by the mirror test. The upstream spec follows the Planton envelope
directly: the size, the template, and how stale members are replaced.

The pool is what a product claims from: its name and namespace are what
agent-sandbox's SDKs take (`create_sandbox(warmpool=..., namespace=...)`).
It is also the cost dial -- every member reserves its template's resources
while it waits -- and the scale target an autoscaler can own (it serves the
scale subresource; leave `replicas` unset when one does).

## Example

```yaml
# The full surface of a KubernetesAgentSandboxWarmPool, offline: every key of
# the SandboxWarmPool spec. The served-schema test renders it and holds it to
# the pinned definition. The live lanes run e2e/scenarios/.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesAgentSandboxWarmPool
metadata:
  name: coding-agent-pool
spec:
  namespace:
    value: agents
  labels:
    team: agents
  annotations:
    example.com/owner: platform
  replicas: 5
  sandboxTemplateRef:
    name:
      value: coding-agent
  updateStrategy:
    type: OnReplenish
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.annotations` | `map<string, string>` |  |  |  |
| `spec.replicas` | `int32` |  |  |  |
| `spec.sandboxTemplateRef` | `KubernetesAgentSandboxWarmPoolTemplateRef` | yes |  |  |
| `spec.sandboxTemplateRef.name` | `string \| valueFrom` | yes |  | KubernetesAgentSandboxTemplate (`status.outputs.sandbox_template_name`) |
| `spec.updateStrategy` | `KubernetesAgentSandboxWarmPoolUpdateStrategy` |  |  |  |
| `spec.updateStrategy.type` | `string` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Kubernetes namespace the SandboxWarmPool is created in; its members, the
claims made against it and its template live there too. Typically a
reference to a KubernetesNamespace resource's `spec.name`.

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.labels

`map<string, string>`

Labels on the SandboxWarmPool object itself (its metadata.labels),
merged under Planton's identity labels (a key in both keeps the identity
value). The members' pod labels come from the template.

- rule: {"map":{"keys":{"string":{"minLen":"1","maxLen":"317","pattern":"^([a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*/)?[A-Za-z0-9]([-A-Za-z0-9_.]{0,61}[A-Za-z0-9])?$"}},"values":{"string":{"maxLen":"63","pattern":"^([A-Za-z0-9]([-A-Za-z0-9_.]{0,61}[A-Za-z0-9])?)?$"}}}}

### spec.annotations

`map<string, string>`

Annotations on the SandboxWarmPool object itself (its
metadata.annotations).

### spec.replicas

`int32` · optional (explicit presence)

How many ready sandboxes the pool keeps. Each member is a Sandbox named
`<pool>-<random>` running the template's pod; a claimed member leaves the
pool and the controller starts a replacement. Zero keeps the pool empty
(every claim starts cold).

Unset, the cluster writes 1 when the pool is created, and Planton never
writes it again: leave it unset when a HorizontalPodAutoscaler or KEDA
scales the pool through its scale subresource, so each apply does not
fight the autoscaler.

- rule: {"int32":{"gte":0}}

### spec.sandboxTemplateRef

`KubernetesAgentSandboxWarmPoolTemplateRef` · required

The template the pool's members are built from, in the pool's namespace.

- rule: {"required":true}

### spec.sandboxTemplateRef.name

`string | valueFrom` · required

Name of the SandboxTemplate, in the pool's namespace. Defaults to a
KubernetesAgentSandboxTemplate reference (its sandbox_template_name
output), so the template is created first; pass the literal name with
`value:` for a template Planton does not manage.

- references: KubernetesAgentSandboxTemplate (`status.outputs.sandbox_template_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesAgentSandboxTemplate, name: <that resource's name>, fieldPath: status.outputs.sandbox_template_name}} -- a bare string does not parse

### spec.updateStrategy

`KubernetesAgentSandboxWarmPoolUpdateStrategy`

How the pool replaces members that no longer match their template (a
change to the template's pod, volume claim templates or `service`, or to
`sandbox_template_ref`). Labels, annotations and the template's claim
and network policies never make a member stale. Claimed sandboxes are
never touched: they belong to their claims.

### spec.updateStrategy.type

`string` · optional (explicit presence)

"Recreate" deletes stale unclaimed members at once, so the pool holds
only sandboxes matching the current template (a rollout with a dip in
ready members while replacements start). "OnReplenish" leaves stale
members until each is claimed or deleted, and builds every replacement
from the current template (no dip, but claims may get the old shape for
a while). Unset, the cluster writes "OnReplenish".

- rule: {"string":{"in":["Recreate","OnReplenish"]}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesAgentSandboxWarmPool, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.sandbox_warm_pool_name` | `string` | Name of the created SandboxWarmPool (equals metadata.name). |
| `status.outputs.namespace` | `string` | Namespace the SandboxWarmPool was created in (the resolved spec.namespace); claims against the pool live here. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.sandboxTemplateRef.name` | KubernetesAgentSandboxTemplate | `status.outputs.sandbox_template_name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| KubernetesAgentSandboxClaim | `spec.warmPoolRef.name` | `status.outputs.sandbox_warm_pool_name` |

## See Also

- [Overview](../README.md)

# KubernetesAgentSubstrateWorkerPool

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

KubernetesAgentSubstrateWorkerPoolSpec defines an Agent Substrate
WorkerPool: a set of worker pods actors run in. Substrate's controller
renders the pool as a Deployment of the same name, each worker registers
with the node agent on its node and then with the API server, and the API
server places each new or resumed actor on a worker with room for it. A
suspended actor leaves its worker for storage and gives its share back,
which is how a pool serves more actors than it holds at once.

100% fidelity with the upstream WorkerPool custom resource
(ate.dev/v1alpha1), pinned to the definition of the Agent Substrate release
KubernetesAgentSubstrate composes (v0.6.0) and held to it by the mirror
test. The upstream spec follows the Planton envelope directly:
the size, the worker image, the worker pods' placement and resources, and
the sandbox class the pool runs.

What a pool is for:
  - Capacity. A worker hosts many actors at once. Its resource limits are
    a budget they share: an actor is placed only on a worker whose unspent
    limits still cover the limits its actor template declares, and only
    while the worker hosts fewer actors than its actor limit (1000 at this
    release, a worker setting the pool does not expose). A worker without
    limits puts no bound on that resource.
  - Isolation. Pools are placed and selected independently: one pool per
    tenant, on its own tainted nodes, is a pool per tenant; an actor
    template's `worker_selector` names the pool by the pool's labels.
  - Runtime. A pool runs one sandbox class (gVisor or microVM) and
    resolves one SandboxConfig of that class, which the cluster then keeps
    from deletion. An actor runs in the class of the pool its template
    selects and boots the default version of the config that pool
    resolves.

A pool can live in any namespace and needs no label there; the controller
adds a NetworkPolicy that admits only Substrate's router to the workers.

## Example

```yaml
# The full surface, for offline proofs (tofu plan, pulumi preview, a server
# dry run): every field of the pool, with literal values where the lanes
# reference fixtures (the worker image the control plane publishes for
# v0.6.0, the ServiceAccount, the PriorityClass and a SandboxConfig).
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesAgentSubstrateWorkerPool
metadata:
  name: e2e-asubwp-full
spec:
  namespace:
    value: e2e-ate-workers
  labels:
    tenant: acme
  annotations:
    planton.dev/e2e-note: full-surface
  replicas: 2
  workerImage:
    value: ghcr.io/plantonhq/planton/agent-substrate/ateworker-gvisor:v0.6.0@sha256:e3f59217a1e63c280ea1cc4149dfe0686353f9ef583fe2ff1c129d4ad42f8cba
  template:
    labels:
      team: platform
    annotations:
      planton.dev/e2e-note: worker
    nodeSelector:
      kubernetes.io/os: linux
    tolerations:
      - key: tenant
        operator: Equal
        value: acme
        effect: NoSchedule
    priorityClassName:
      value: e2e-ate-workers
    nodeAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
        nodeSelectorTerms:
          - matchExpressions:
              - key: kubernetes.io/arch
                operator: In
                values:
                  - amd64
                  - arm64
    resources:
      requests:
        cpu: 100m
        memory: 256Mi
      limits:
        cpu: "1"
        memory: 1Gi
    serviceAccountName:
      value: e2e-ate-workers
  sandboxClasses:
    - name: gvisor
      configRef:
        name:
          value: e2e-gvisor-workers
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.annotations` | `map<string, string>` |  |  |  |
| `spec.replicas` | `int32` | yes |  |  |
| `spec.workerImage` | `string \| valueFrom` | yes |  | KubernetesAgentSubstrate (`status.outputs.gvisor_worker_image`) |
| `spec.template` | `KubernetesAgentSubstrateWorkerPoolPodTemplate` |  |  |  |
| `spec.template.labels` | `map<string, string>` |  |  |  |
| `spec.template.annotations` | `map<string, string>` |  |  |  |
| `spec.template.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.template.tolerations` | `[]KubernetesCoreV1ApiToleration` |  |  |  |
| `spec.template.tolerations[].key` | `string` |  |  |  |
| `spec.template.tolerations[].operator` | `string` |  |  |  |
| `spec.template.tolerations[].value` | `string` |  |  |  |
| `spec.template.tolerations[].effect` | `string` |  |  |  |
| `spec.template.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.template.priorityClassName` | `string \| valueFrom` |  |  | KubernetesPriorityClass (`status.outputs.priority_class_name`) |
| `spec.template.nodeAffinity` | `KubernetesCoreV1ApiNodeAffinity` |  |  |  |
| `spec.template.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution` | `KubernetesCoreV1ApiNodeSelector` |  |  |  |
| `spec.template.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms` | `[]KubernetesCoreV1ApiNodeSelectorTerm` | yes |  |  |
| `spec.template.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[].matchExpressions` | `[]KubernetesCoreV1ApiNodeSelectorRequirement` |  |  |  |
| `spec.template.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[].matchExpressions[].key` | `string` | yes |  |  |
| `spec.template.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[].matchExpressions[].operator` | `string` |  |  |  |
| `spec.template.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[].matchExpressions[].values` | `[]string` |  |  |  |
| `spec.template.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[].matchFields` | `[]KubernetesCoreV1ApiNodeSelectorRequirement` |  |  |  |
| `spec.template.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[].matchFields[].key` | `string` | yes |  |  |
| `spec.template.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[].matchFields[].operator` | `string` |  |  |  |
| `spec.template.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[].matchFields[].values` | `[]string` |  |  |  |
| `spec.template.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution` | `[]KubernetesCoreV1ApiPreferredSchedulingTerm` |  |  |  |
| `spec.template.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[].weight` | `int32` | yes |  |  |
| `spec.template.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[].preference` | `KubernetesCoreV1ApiNodeSelectorTerm` | yes |  |  |
| `spec.template.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[].preference.matchExpressions` | `[]KubernetesCoreV1ApiNodeSelectorRequirement` |  |  |  |
| `spec.template.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[].preference.matchExpressions[].key` | `string` | yes |  |  |
| `spec.template.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[].preference.matchExpressions[].operator` | `string` |  |  |  |
| `spec.template.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[].preference.matchExpressions[].values` | `[]string` |  |  |  |
| `spec.template.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[].preference.matchFields` | `[]KubernetesCoreV1ApiNodeSelectorRequirement` |  |  |  |
| `spec.template.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[].preference.matchFields[].key` | `string` | yes |  |  |
| `spec.template.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[].preference.matchFields[].operator` | `string` |  |  |  |
| `spec.template.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[].preference.matchFields[].values` | `[]string` |  |  |  |
| `spec.template.resources` | `KubernetesCoreV1ApiResourceRequirements` |  |  |  |
| `spec.template.resources.limits` | `map<string, string>` |  |  |  |
| `spec.template.resources.requests` | `map<string, string>` |  |  |  |
| `spec.template.resources.claims` | `[]KubernetesCoreV1ApiResourceClaim` |  |  |  |
| `spec.template.resources.claims[].name` | `string` | yes |  |  |
| `spec.template.resources.claims[].request` | `string` |  |  |  |
| `spec.template.serviceAccountName` | `string \| valueFrom` |  |  | KubernetesServiceAccount (`status.outputs.service_account_name`) |
| `spec.sandboxClasses` | `[]KubernetesAgentSubstrateWorkerPoolSandboxClass` | yes |  |  |
| `spec.sandboxClasses[].name` | `string` | yes |  |  |
| `spec.sandboxClasses[].configRef` | `KubernetesAgentSubstrateWorkerPoolSandboxConfigRef` |  |  |  |
| `spec.sandboxClasses[].configRef.name` | `string \| valueFrom` | yes |  | KubernetesAgentSubstrateSandboxConfig (`status.outputs.sandbox_config_name`) |

## Field Details

### spec.namespace

`string | valueFrom` · required

Kubernetes namespace the WorkerPool and its worker pods run in. Typically
a reference to a KubernetesNamespace resource's `spec.name`. The worker
pods run as root with broad capabilities (gVisor and microVM need them),
so a namespace that enforces the restricted or baseline Pod Security
level refuses them.

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.labels

`map<string, string>`

Labels on the WorkerPool object itself (its metadata.labels), merged
under Planton's identity labels (a key in both keeps the identity value).

These are the labels actors select the pool by: Substrate registers every
worker with the pool's labels, and an actor template's
`worker_selector.match_labels` must match them (for example
`tenant: acme`). The worker pods' own labels come from
`template.labels`.

- rule: {"map":{"keys":{"string":{"minLen":"1","maxLen":"317","pattern":"^([a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*/)?[A-Za-z0-9]([-A-Za-z0-9_.]{0,61}[A-Za-z0-9])?$"}},"values":{"string":{"maxLen":"63","pattern":"^([A-Za-z0-9]([-A-Za-z0-9_.]{0,61}[A-Za-z0-9])?)?$"}}}}

### spec.annotations

`map<string, string>`

Annotations on the WorkerPool object itself (its metadata.annotations).

### spec.replicas

`int32` · required · optional (explicit presence)

How many worker pods the pool runs. Zero keeps the pool defined with no
workers. A rollout replaces at most 10% of the workers at once and never
starts extra ones, and a stopping worker is given up to an hour to drain
its actors first, so shrinking and rolling a pool are gradual.

Required: every apply writes it. The pool serves the scale subresource,
but upstream treats worker autoscaling as a demo at this release, so an
autoscaler that resizes the pool is undone by the next apply.

- rule: {"required":true,"int32":{"gte":0}}

### spec.workerImage

`string | valueFrom` · required

The worker image the pool runs, which must match the sandbox class: the
control plane's `gvisor_worker_image` output for a gvisor pool, its
`microvm_worker_image` output for a microvm pool. Defaults to a
KubernetesAgentSubstrate reference (its gvisor_worker_image output), so
the pool runs the worker built with the installed release, with the
control plane's registry mirror already applied (after the control
plane's version moves, redeploy the pool to roll its workers onto the new
release's image); a microvm pool
points the reference at `status.outputs.microvm_worker_image`. A literal
image works too, pinned by digest.

- references: KubernetesAgentSubstrate (`status.outputs.gvisor_worker_image`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesAgentSubstrate, name: <that resource's name>, fieldPath: status.outputs.gvisor_worker_image}} -- a bare string does not parse

### spec.template

`KubernetesAgentSubstrateWorkerPoolPodTemplate`

Metadata, placement and resources for the worker pods. Unset, workers run
anywhere the scheduler puts them, with no resource requests.

### spec.template.labels

`map<string, string>`

Labels added to the worker Deployment and its pods, at most 64. Keys in
the ate.dev domain and its subdomains are Substrate's own. These do not
select the pool for actors; the pool's own labels do.

- rule: label keys in the ate.dev domain and its subdomains are reserved for Substrate's controllers
- rule: {"map":{"maxPairs":"64","keys":{"string":{"minLen":"1","maxLen":"317","pattern":"^([a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*/)?[A-Za-z0-9]([-A-Za-z0-9_.]{0,61}[A-Za-z0-9])?$"}},"values":{"string":{"maxLen":"63","pattern":"^(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?$"}}}}

### spec.template.annotations

`map<string, string>`

Annotations added to the worker Deployment and its pods, at most 64. Keys
in the ate.dev domain and its subdomains are Substrate's own.

- rule: annotation keys in the ate.dev domain and its subdomains are reserved for Substrate's controllers
- rule: {"map":{"maxPairs":"64","keys":{"string":{"minLen":"1","maxLen":"317","pattern":"^([a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*/)?[A-Za-z0-9]([-A-Za-z0-9_.]{0,61}[A-Za-z0-9])?$"}}}}

### spec.template.nodeSelector

`map<string, string>`

Node labels a worker's node must carry, for example a dedicated node
pool's label. A microvm pool needs no selector for KVM: the node agent
advertises KVM as a device and the worker requests it.

### spec.template.tolerations

`[]KubernetesCoreV1ApiToleration`

Tolerations for the worker pods, at most 16, for placing a pool on
tainted nodes (a tenant's own nodes, say). The controller always adds
`ate.dev/sandboxClass=<class>:NoSchedule`, so nodes tainted that way take
only workers of that class.

- rule: {"repeated":{"maxItems":"16"}}
- rule: A toleration with operator Exists matches any value, so it takes no value

### spec.template.tolerations[].key

`string`

`key` is the taint key that the toleration applies to. Empty means match
all taint keys. If the key is empty, operator must be Exists; this
combination means to match all values and all keys.

### spec.template.tolerations[].operator

`string` · optional (explicit presence)

`operator` represents a key's relationship to the value. Valid operators
are Exists, Equal, Lt, and Gt. Defaults to Equal. Exists is equivalent to
wildcard for value, so that a pod can tolerate all taints of a particular
category. Lt and Gt perform numeric comparisons (requires feature gate
TaintTolerationComparisonOperators).

- rule: {"string":{"in":["Exists","Equal","Lt","Gt"]}}

### spec.template.tolerations[].value

`string`

`value` is the taint value the toleration matches to. If the operator is
Exists, the value should be empty, otherwise just a regular string.

### spec.template.tolerations[].effect

`string` · optional (explicit presence)

`effect` indicates the taint effect to match. Empty means match all taint
effects. When specified, allowed values are NoSchedule, PreferNoSchedule
and NoExecute.

- rule: {"string":{"in":["NoSchedule","PreferNoSchedule","NoExecute"]}}

### spec.template.tolerations[].tolerationSeconds

`int64` · optional (explicit presence)

`tolerationSeconds` represents the period of time the toleration (which
must be of effect NoExecute, otherwise this field is ignored) tolerates the
taint. By default, it is not set, which means tolerate the taint forever
(do not evict). Zero and negative values will be treated as 0 (evict
immediately) by the system.

### spec.template.priorityClassName

`string | valueFrom`

The PriorityClass of the worker pods. Defaults to a KubernetesPriorityClass
reference (its priority_class_name output); pass the literal name with
`value:` for one Planton does not manage.

- references: KubernetesPriorityClass (`status.outputs.priority_class_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesPriorityClass, name: <that resource's name>, fieldPath: status.outputs.priority_class_name}} -- a bare string does not parse

### spec.template.nodeAffinity

`KubernetesCoreV1ApiNodeAffinity`

Node affinity rules for the worker pods, written to the pods'
`affinity.nodeAffinity`.

### spec.template.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution

`KubernetesCoreV1ApiNodeSelector`

If the affinity requirements specified by this field are not met at
scheduling time, the pod will not be scheduled onto the node. If the
affinity requirements specified by this field cease to be met at some point
during pod execution (e.g. due to an update), the system may or may not try
to eventually evict the pod from its node.

### spec.template.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms

`[]KubernetesCoreV1ApiNodeSelectorTerm` · required

Required. A list of node selector terms. The terms are ORed.

- rule: {"repeated":{"minItems":"1"}}

### spec.template.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[].matchExpressions

`[]KubernetesCoreV1ApiNodeSelectorRequirement`

A list of node selector requirements by node's labels.

- rule: In and NotIn need at least one value, Gt and Lt exactly one integer, and Exists and DoesNotExist none

### spec.template.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[].matchExpressions[].key

`string` · required

The label key that the selector applies to.

- rule: {"string":{"minLen":"1"}}

### spec.template.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[].matchExpressions[].operator

`string`

Represents a key's relationship to a set of values. Valid operators are In,
NotIn, Exists, DoesNotExist. Gt, and Lt.

- rule: {"string":{"in":["In","NotIn","Exists","DoesNotExist","Gt","Lt"]}}

### spec.template.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[].matchExpressions[].values

`[]string`

An array of string values. If the operator is In or NotIn, the values array
must be non-empty. If the operator is Exists or DoesNotExist, the values
array must be empty. If the operator is Gt or Lt, the values array must
have a single element, which will be interpreted as an integer. This array
is replaced during a strategic merge patch.

### spec.template.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[].matchFields

`[]KubernetesCoreV1ApiNodeSelectorRequirement`

A list of node selector requirements by node's fields.

- rule: In and NotIn need at least one value, Gt and Lt exactly one integer, and Exists and DoesNotExist none

### spec.template.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[].matchFields[].key

`string` · required

The label key that the selector applies to.

- rule: {"string":{"minLen":"1"}}

### spec.template.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[].matchFields[].operator

`string`

Represents a key's relationship to a set of values. Valid operators are In,
NotIn, Exists, DoesNotExist. Gt, and Lt.

- rule: {"string":{"in":["In","NotIn","Exists","DoesNotExist","Gt","Lt"]}}

### spec.template.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[].matchFields[].values

`[]string`

An array of string values. If the operator is In or NotIn, the values array
must be non-empty. If the operator is Exists or DoesNotExist, the values
array must be empty. If the operator is Gt or Lt, the values array must
have a single element, which will be interpreted as an integer. This array
is replaced during a strategic merge patch.

### spec.template.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution

`[]KubernetesCoreV1ApiPreferredSchedulingTerm`

The scheduler will prefer to schedule pods to nodes that satisfy the
affinity expressions specified by this field, but it may choose a node that
violates one or more of the expressions. The node that is most preferred is
the one with the greatest sum of weights, i.e. for each node that meets all
of the scheduling requirements (resource request, requiredDuringScheduling
affinity expressions, etc.), compute a sum by iterating through the
elements of this field and adding "weight" to the sum if the node matches
the corresponding matchExpressions; the node(s) with the highest sum are
the most preferred.

### spec.template.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[].weight

`int32` · required · optional (explicit presence)

`weight` associated with matching the corresponding nodeSelectorTerm, in
the range 1-100.

- rule: {"required":true}

### spec.template.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[].preference

`KubernetesCoreV1ApiNodeSelectorTerm` · required

A node selector term, associated with the corresponding weight.

- rule: {"required":true}

### spec.template.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[].preference.matchExpressions

`[]KubernetesCoreV1ApiNodeSelectorRequirement`

A list of node selector requirements by node's labels.

- rule: In and NotIn need at least one value, Gt and Lt exactly one integer, and Exists and DoesNotExist none

### spec.template.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[].preference.matchExpressions[].key

`string` · required

The label key that the selector applies to.

- rule: {"string":{"minLen":"1"}}

### spec.template.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[].preference.matchExpressions[].operator

`string`

Represents a key's relationship to a set of values. Valid operators are In,
NotIn, Exists, DoesNotExist. Gt, and Lt.

- rule: {"string":{"in":["In","NotIn","Exists","DoesNotExist","Gt","Lt"]}}

### spec.template.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[].preference.matchExpressions[].values

`[]string`

An array of string values. If the operator is In or NotIn, the values array
must be non-empty. If the operator is Exists or DoesNotExist, the values
array must be empty. If the operator is Gt or Lt, the values array must
have a single element, which will be interpreted as an integer. This array
is replaced during a strategic merge patch.

### spec.template.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[].preference.matchFields

`[]KubernetesCoreV1ApiNodeSelectorRequirement`

A list of node selector requirements by node's fields.

- rule: In and NotIn need at least one value, Gt and Lt exactly one integer, and Exists and DoesNotExist none

### spec.template.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[].preference.matchFields[].key

`string` · required

The label key that the selector applies to.

- rule: {"string":{"minLen":"1"}}

### spec.template.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[].preference.matchFields[].operator

`string`

Represents a key's relationship to a set of values. Valid operators are In,
NotIn, Exists, DoesNotExist. Gt, and Lt.

- rule: {"string":{"in":["In","NotIn","Exists","DoesNotExist","Gt","Lt"]}}

### spec.template.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[].preference.matchFields[].values

`[]string`

An array of string values. If the operator is In or NotIn, the values array
must be non-empty. If the operator is Exists or DoesNotExist, the values
array must be empty. If the operator is Gt or Lt, the values array must
have a single element, which will be interpreted as an integer. This array
is replaced during a strategic merge patch.

### spec.template.resources

`KubernetesCoreV1ApiResourceRequirements`

Compute resources of each worker pod. The cpu and memory limits are the
capacity the worker advertises, a budget its actors share: each placed
actor spends the limits its actor template declares, and an actor that
fits on no worker's remainder is refused for lack of capacity. Leaving a
limit unset leaves that resource unbounded for placement, so size the
limits for the actors a worker should host together. Requests size what
the Kubernetes scheduler reserves per worker. Upstream's controller
copies only requests and limits to the worker container; `claims` is
accepted and not applied.

### spec.template.resources.limits

`map<string, string>`

`limits` describes the maximum amount of compute resources allowed. More
info:
https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/

- rule: {"map":{"values":{"string":{"pattern":"^(\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))))?$"}}}}

### spec.template.resources.requests

`map<string, string>`

`requests` describes the minimum amount of compute resources required. If
Requests is omitted for a container, it defaults to Limits if that is
explicitly specified, otherwise to an implementation-defined value.
Requests cannot exceed Limits. More info:
https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/

- rule: {"map":{"values":{"string":{"pattern":"^(\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))))?$"}}}}

### spec.template.resources.claims

`[]KubernetesCoreV1ApiResourceClaim`

`claims` lists the names of resources, defined in spec.resourceClaims, that
are used by this container.

This field depends on the DynamicResourceAllocation feature gate.

This field is immutable. It can only be set for containers.

### spec.template.resources.claims[].name

`string` · required

`name` must match the name of one entry in pod.spec.resourceClaims of the
Pod where this field is used. It makes that resource available inside a
container.

- rule: {"string":{"minLen":"1"}}

### spec.template.resources.claims[].request

`string`

`request` is the name chosen for a request in the referenced claim. If
empty, everything from the claim is made available, otherwise only the
result of this request.

### spec.template.serviceAccountName

`string | valueFrom`

The ServiceAccount the worker pods run as, in the pool's namespace.
Unset, the pods run as `ateworker`, which the controller creates in the
pool's namespace. Defaults to a
KubernetesServiceAccount reference (its service_account_name output);
pass the literal name with `value:` for one Planton does not manage.

- references: KubernetesServiceAccount (`status.outputs.service_account_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesServiceAccount, name: <that resource's name>, fieldPath: status.outputs.service_account_name}} -- a bare string does not parse

### spec.sandboxClasses

`[]KubernetesAgentSubstrateWorkerPoolSandboxClass` · required

The sandbox class the pool runs, with the SandboxConfig it runs it with.
Exactly one entry at this release. Its name cannot change after the pool
is created (the cluster refuses the change, a rule a spec cannot check
because it compares with the object as it was); recreate the pool to
switch classes.

- rule: {"repeated":{"minItems":"1","maxItems":"1"}}

### spec.sandboxClasses[].name

`string` · required

The sandbox runtime family, which shapes the worker pods: "gvisor" runs
actors under gVisor inside an ordinary worker pod; "microvm" requests
the KVM device the node agent advertises (so workers land only on nodes
that have it) and mounts the host's TUN device. The worker image must be
built for the same class.

- rule: {"required":true,"string":{"in":["gvisor","microvm"]}}

### spec.sandboxClasses[].configRef

`KubernetesAgentSubstrateWorkerPoolSandboxConfigRef`

The SandboxConfig the pool resolves for this class; its class must match.
Unset, the pool resolves the newest config of the class annotated
`sandboxconfig.ate.dev/is-class-default: "true"` (for gvisor, the
control plane's `gvisor-default` unless a newer default exists). The
pool's status reports the config in use; the controller starts the
worker Deployment only once the class resolves, and a config a pool uses
cannot be deleted until the pool moves off it. Every actor on the pool
boots this config's default version, so the deletion guard covers the
runtime its actors run.

### spec.sandboxClasses[].configRef.name

`string | valueFrom` · required

Name of the SandboxConfig. Defaults to a
KubernetesAgentSubstrateSandboxConfig reference (its sandbox_config_name
output), so the config is created first; pass the literal name with
`value:` for one Planton does not manage, such as `gvisor-default`.

- references: KubernetesAgentSubstrateSandboxConfig (`status.outputs.sandbox_config_name`)
- rule: a SandboxConfig name is at most 253 characters
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesAgentSubstrateSandboxConfig, name: <that resource's name>, fieldPath: status.outputs.sandbox_config_name}} -- a bare string does not parse

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesAgentSubstrateWorkerPool, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.worker_pool_name` | `string` | Name of the created WorkerPool (equals metadata.name); the worker Deployment the controller renders carries the same name. |
| `status.outputs.namespace` | `string` | Namespace the WorkerPool and its worker pods run in (the resolved spec.namespace). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.workerImage` | KubernetesAgentSubstrate | `status.outputs.gvisor_worker_image` |
| `spec.template.priorityClassName` | KubernetesPriorityClass | `status.outputs.priority_class_name` |
| `spec.template.serviceAccountName` | KubernetesServiceAccount | `status.outputs.service_account_name` |
| `spec.sandboxClasses[].configRef.name` | KubernetesAgentSubstrateSandboxConfig | `status.outputs.sandbox_config_name` |

## See Also

- [Overview](../README.md)

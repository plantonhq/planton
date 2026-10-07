# KubernetesPlantonRunner

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**KubernetesPlantonRunnerSpec** declares a standing Planton runner on a
Kubernetes cluster — an always-on worker that receives deploy operations
from the Planton control plane and executes them from INSIDE the
cluster's network. The module installs the official `planton-runner`
Helm chart (OCI, ghcr.io/plantonhq/charts) as a real Helm release, so
the deployed runner is byte-identical to a hand-installed one.

Why it exists: some targets are reachable only from inside the network —
the canonical case is a private service or a cluster API endpoint no
hosted runner fleet can reach. Running the runner in the cluster makes
those targets deployable and operable with zero inbound exposure: the
runner only ever dials OUT to the control plane.

ENROLLMENT IS TOKEN-FIRST: the runner is born with a runner TOKEN, never
an identity. On first boot it presents the token to the control plane,
registers ITSELF, and receives its own individually revocable identity;
the identity persists on the pod's ephemeral volume, container restarts
reuse it, and pod recreation re-joins with the same token (the token's
lineage re-admits the runner it originally admitted — no other token
can). The token lives in a module-created Kubernetes Secret; it never
rides rendered chart values.

EXACTLY ONE REPLICA, by design: a runner's identity is minted for one
live instance — a second instance joining under the same name would
revoke the first's key. The chart pins replicas to 1 with a Recreate
strategy; scaling execution capacity means more runners (more resources
of this kind), never more copies of this one.

## Example

```yaml
# Minimal KubernetesPlantonRunner manifest for local module testing. The
# token value below is an obviously-fake placeholder with the right shape
# -- real deployments supply a managed-secret reference ($secret/<slug>)
# that the platform fills with a runner token before the infrastructure
# applies. The token authorizes joining and is never the runner's
# identity: the runner registers itself on first boot and receives its
# own individually revocable identity.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesPlantonRunner
metadata:
  name: kubernetesplantonrunner-demo
spec:
  namespace:
    value: planton-runner
  createNamespace: true
  token: prt_FAKE_PLACEHOLDER_VALUE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.token` | `string` (sensitive) | yes |  |  |
| `spec.runnerName` | `string` |  |  |  |
| `spec.controlPlaneEndpoint` | `string` |  |  |  |
| `spec.runnerVersion` | `string` |  | `latest` |  |
| `spec.imageRepository` | `string` |  | `ghcr.io/plantonhq/planton/runner` |  |
| `spec.chartVersion` | `string` |  |  |  |
| `spec.resources` | `ContainerResources` |  |  |  |
| `spec.resources.limits` | `CpuMemory` |  |  |  |
| `spec.resources.limits.cpu` | `string` |  |  |  |
| `spec.resources.limits.memory` | `string` |  |  |  |
| `spec.resources.requests` | `CpuMemory` |  |  |  |
| `spec.resources.requests.cpu` | `string` |  |  |  |
| `spec.resources.requests.memory` | `string` |  |  |  |
| `spec.build` | `KubernetesPlantonRunnerBuild` |  |  |  |
| `spec.build.enabled` | `bool` |  |  |  |
| `spec.build.tektonNamespace` | `string` |  |  |  |
| `spec.build.scheduling` | `KubernetesPlantonRunnerBuildScheduling` |  |  |  |
| `spec.build.scheduling.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.build.scheduling.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.build.scheduling.tolerations[].key` | `string` |  |  |  |
| `spec.build.scheduling.tolerations[].operator` | `string` |  |  |  |
| `spec.build.scheduling.tolerations[].value` | `string` |  |  |  |
| `spec.build.scheduling.tolerations[].effect` | `string` |  |  |  |
| `spec.build.scheduling.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.helmValues` | `string` |  |  |  |
| `spec.chartRepository` | `string` |  | `oci://ghcr.io/plantonhq/charts` |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

The namespace to install the runner into. Accepts a literal namespace
name or a reference to a KubernetesNamespace resource.

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.createNamespace

`bool`

When true, the namespace is created (with the standard Planton
governance labels) before the release is installed, and deleted with
the resource. When false, the namespace must already exist.

### spec.token

`string` · required · sensitive

The runner token that authorizes this runner to JOIN the control
plane. Create one with `planton runner token create` (or in the
console under Organization Settings → Runner Tokens); on Planton, the
platform mints a token and writes it at exactly the managed-secret
reference this field names before the infrastructure applies — there
is no manual credential step. The token only gates joining and is
never the runner's identity: the runner receives its own individually
revocable identity when it registers itself on arrival, and revoking
this token never touches runners it already admitted. This is a
secret: supply it as a managed-secret reference, never inline
plaintext; the module stores it in a Kubernetes Secret the chart
reads by name.

- rule: {"required":true}

### spec.runnerName

`string`

The name this runner registers itself under when it joins — how it
appears in `planton runner list` and the console, and the name deploy
operations are routed to. Defaults to "<env>-<metadata.name>" (or
metadata.name outside an environment) — the same derivation the
platform uses for records that reference this runner, so leave it
unset unless you are deliberately adopting an existing enrollment.
Re-deploying with the SAME name and the SAME token re-admits the
runner (lost-disk recovery); a different token answers a closed door.

- rule: runner name must be 1-63 characters: lowercase letters, digits, and hyphens, starting with a letter and ending with a letter or digit
- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.controlPlaneEndpoint

`string`

The control-plane endpoint the runner joins, as host:port. Leave
unset for Planton's hosted control plane (the runner's built-in
default); set it for a self-hosted instance (e.g.
"planton.example.com:443"). This is the one bootstrap coordinate the
join cannot deliver — everything else (work queue, tunnel, API
endpoints) arrives in the join response.

- rule: control plane endpoint must be host:port, e.g. "planton.example.com:443" — no scheme prefix
- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.runnerVersion

`string` · optional (explicit presence)

The runner build to deploy: an image tag of the official runner
container image. "latest" tracks the newest release; pin a specific
version tag for change control.

- default: `latest`

### spec.imageRepository

`string` · optional (explicit presence)

The container image repository the runner is pulled from. Override
only for air-gapped or mirrored registries hosting a copy of the
official image; the digest-identical mirror is your responsibility.

- default: `ghcr.io/plantonhq/planton/runner`

### spec.chartVersion

`string`

The planton-runner chart version to install. Defaults to the version
this catalog release was validated against; pin a specific version
only for change control. Versions below 0.4.0 predate token
enrollment and are refused by the module.

- rule: chart version must be an exact semver like "0.4.0" — ranges are not reproducible
- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.resources

`ContainerResources`

CPU/memory for the runner container. When omitted, the chart's own
defaults apply (requests 100m/256Mi, limits 1/2Gi). The runner runs
as many IaC operations at once as the memory limit holds -- about 1Gi
each beside its own ~768Mi, so 2Gi runs one and 4Gi three -- and
queues the rest; raise the memory limit to run more at once.

### spec.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.resources.limits.cpu

`string`

### spec.resources.limits.memory

`string`

### spec.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.resources.requests.cpu

`string`

### spec.resources.requests.memory

`string`

### spec.build

`KubernetesPlantonRunnerBuild`

Enables the runner's build worker: the runner then also executes
container-image build pipelines through Tekton on this cluster.
Requires Tekton Pipelines to be installed.

- rule: build.scheduling places build pods, so it needs build.enabled: true; without builds it would do nothing

### spec.build.enabled

`bool`

When true, the runner registers as a build worker and executes
container-image build pipelines on this cluster.

### spec.build.tektonNamespace

`string`

The namespace Tekton build pipelines run in. Defaults to the
runner's own namespace.

- rule: tekton namespace must be a valid Kubernetes namespace name: lowercase letters, digits, and hyphens, at most 63 characters
- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.build.scheduling

`KubernetesPlantonRunnerBuildScheduling`

Which nodes build pods may use. The runner puts this on every
PipelineRun's pod template, so every task pod, and the helper pod Tekton
uses to keep a run's pods together, lands only there. The usual shape is a
dedicated, tainted build node pool, so a burst of builds can never starve
the cluster's other workloads. Unset keeps builds wherever the scheduler
puts them, which is right for a single-node cluster. This places BUILD
pods; `helm_values` still places the runner itself.

### spec.build.scheduling.nodeSelector

`map<string, string>`

Every listed label must match the node (e.g.
`planton.ai/workload: build`).

### spec.build.scheduling.tolerations

`[]WorkloadToleration`

Tolerations that let build pods onto tainted nodes. A toleration only
permits; pair it with `node_selector` so builds go nowhere else.

### spec.build.scheduling.tolerations[].key

`string`

Taint key to tolerate. Empty key with operator "Exists" tolerates every taint.

### spec.build.scheduling.tolerations[].operator

`string`

How key/value match: "Equal" (default — value must match too) or "Exists"
(key presence alone matches).

- rule: Toleration operator must be either "Equal" or "Exists"

### spec.build.scheduling.tolerations[].value

`string`

Taint value to match when operator is "Equal".

### spec.build.scheduling.tolerations[].effect

`string`

Which taint effect is tolerated: "NoSchedule", "PreferNoSchedule", or
"NoExecute". Empty tolerates all effects for the key.

- rule: Toleration effect must be one of "NoSchedule", "PreferNoSchedule", or "NoExecute"

### spec.build.scheduling.tolerations[].tolerationSeconds

`int64` · optional (explicit presence)

For "NoExecute" taints only: how many seconds already-running pods stay bound
after the taint appears. Unset means tolerate forever.

### spec.helmValues

`string`

Advanced escape hatch: raw Helm values YAML merged OVER the values
this spec renders (Helm `-f` semantics: maps deep-merge with these
overrides winning, lists replace). Use it for chart knobs the spec
does not model (nodeSelector, tolerations, extra env) — never for
secret material: the enrollment token is carried by the
module-created Secret, and the enrollment block is re-pinned after
the merge so an override can never move it into rendered values.

### spec.chartRepository

`string` · optional (explicit presence)

The OCI registry path the planton-runner chart is pulled from. Defaults
to oci://ghcr.io/plantonhq/charts. Every chart release is also
published, byte for byte, to Google Artifact Registry at
oci://us-central1-docker.pkg.dev/plantonhq/charts; set that to pull
from Google, or name a mirror of your own holding the same charts.

- default: `oci://ghcr.io/plantonhq/charts`
- rule: chart repository must be an OCI path such as "oci://us-central1-docker.pkg.dev/plantonhq/charts": oci:// scheme, no trailing slash

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesPlantonRunner, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | The namespace the runner is installed in. |
| `status.outputs.release_name` | `string` | The Helm release name (metadata.name) — the handle for `helm status`/`helm get values` inspection. |
| `status.outputs.token_secret_name` | `string` | The name of the Kubernetes Secret holding the runner token. The chart's Deployment reads PLANTON_RUNNER_TOKEN from this Secret; the token authorizes joining and is never the runner's identity. |
| `status.outputs.runner_name` | `string` | The name the runner registers itself under with the control plane — the value shown by `planton runner list` the moment it joins. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |

## See Also

- [Overview](../README.md)

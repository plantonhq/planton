# AzurePlantonRunner

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AzurePlantonRunnerSpec defines a standing Planton runner appliance in
your Azure subscription: an always-on worker that receives deploy
operations from the Planton control plane and executes them from WITHIN
your network perimeter.

Why it exists: some targets are reachable only from inside the network --
the canonical case is a private AKS API server or a private-endpoint
database, which no hosted runner fleet can reach. Placing a runner
beside them makes those targets deployable and operable (day-2 updates,
destroys, CloudOps) with zero inbound exposure: the runner only ever
dials OUT to the control plane, and the deployed app exposes no ingress
at all.

ENROLLMENT IS TOKEN-FIRST: the runner is born with a runner TOKEN, never
an identity. On first boot it presents the token to the control plane,
registers ITSELF, and receives its own individually revocable identity.
Replica replacement re-joins with the same token (the token's lineage
re-admits the runner it originally admitted -- no other token can). The
token lives in the Container App's own secret store; the container reads
it as a secret-backed environment variable, never as plaintext.

The runner runs inside a Container App Environment (the referenced
prerequisite) -- the environment decides the network boundary: an
environment integrated with a VNet gives the runner reach into that
network's private endpoints. The spec models intent -- where the runner
lives (resource group/environment), how big it is (cpu/memory), which
build it runs (runner_version), and the token it joins with; the
compute substrate (a single-revision Container App pinned to exactly
one replica) is an implementation detail of the IaC modules.

## Example

```yaml
# Minimal AzurePlantonRunner manifest for local module testing. The token
# value below is an obviously-fake placeholder with the right shape --
# real deployments supply a managed-secret reference ($secret/<slug>)
# that the platform fills with a runner token before the infrastructure
# applies. The token authorizes joining and is never the runner's
# identity: the runner registers itself on first boot and receives its
# own individually revocable identity.
apiVersion: azure.planton.dev/v1alpha1
kind: AzurePlantonRunner
metadata:
  name: azureplantonrunner-demo
spec:
  resourceGroup:
    value: my-resource-group
  containerAppEnvironmentId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-resource-group/providers/Microsoft.App/managedEnvironments/my-environment
  token: prt_FAKE_PLACEHOLDER_VALUE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.containerAppEnvironmentId` | `string \| valueFrom` | yes |  | AzureContainerAppEnvironment (`status.outputs.environment_id`) |
| `spec.token` | `string` (sensitive) | yes |  |  |
| `spec.controlPlaneEndpoint` | `string` |  |  |  |
| `spec.runnerVersion` | `string` |  | `latest` |  |
| `spec.imageRepository` | `string` |  | `ghcr.io/plantonhq/planton/runner` |  |
| `spec.cpu` | `double` |  | `0.5` |  |
| `spec.memory` | `string` |  | `1Gi` |  |

## Field Details

### spec.resourceGroup

`string | valueFrom` · required

The Azure resource group the runner is created in. Accepts a literal
name or a reference to an AzureResourceGroup resource.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.containerAppEnvironmentId

`string | valueFrom` · required

The Container App Environment the runner runs in. The environment
decides the network boundary: pick (or compose) a VNet-integrated
environment when the runner must reach private endpoints. Accepts a
literal environment resource ID or a reference to an
AzureContainerAppEnvironment resource.

- references: AzureContainerAppEnvironment (`status.outputs.environment_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureContainerAppEnvironment, name: <that resource's name>, fieldPath: status.outputs.environment_id}} -- a bare string does not parse

### spec.token

`string` · required · sensitive

The runner token that authorizes this runner to JOIN the control
plane. Create one with `planton runner token create` (or in the
console under Organization Settings -> Runner Tokens); on Planton, the
platform mints a token and writes it at exactly the managed-secret
reference this field names, before the infrastructure applies -- there
is no manual credential step. The token only gates joining and is
never the runner's identity: the runner receives its own individually
revocable identity when it registers itself on arrival, and revoking
this token never touches runners it already admitted. This is a
secret: supply it as a managed-secret reference, never inline
plaintext; it reaches the runner through the Container App's secret
store, not through any plain environment variable.

- rule: {"required":true}

### spec.controlPlaneEndpoint

`string`

The control-plane endpoint the runner joins, as host:port. Leave
unset for Planton's hosted control plane (the runner's built-in
default); set it for a self-hosted instance (e.g.
"planton.example.com:443"). This is the one bootstrap coordinate the
join cannot deliver -- everything else (work queue, tunnel, API
endpoints) arrives in the join response.

- rule: control plane endpoint must be host:port, e.g. "planton.example.com:443" -- no scheme prefix
- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.runnerVersion

`string` · optional (explicit presence)

The runner build to deploy: an image tag of the official runner
container image. "latest" tracks the newest release; pin a specific
version tag for change control. New replicas pull the tag on every
(re)start.

- default: `latest`

### spec.imageRepository

`string` · optional (explicit presence)

The container image repository the runner is pulled from. Override
only for air-gapped or mirrored registries hosting a copy of the
official image; the digest-identical mirror is your responsibility.

- default: `ghcr.io/plantonhq/planton/runner`

### spec.cpu

`double` · optional (explicit presence)

CPU allocated to the runner, in vCPUs. The default 0.5 comfortably
runs the runner's control loops plus typical IaC operations; size up
for large stacks or high operation concurrency. On the Consumption
plan, cpu and memory must form one of the fixed pairings (see
memory).

- default: `0.5`

### spec.memory

`string` · optional (explicit presence)

Memory allocated to the runner, e.g. "1Gi". The default 1Gi pairs
with the default cpu of 0.5. On the Consumption plan the pairing is
fixed at memory = cpu x 2 (0.25/0.5Gi, 0.5/1Gi, 0.75/1.5Gi, 1/2Gi,
1.25/2.5Gi, 1.5/3Gi, 1.75/3.5Gi, 2/4Gi) -- an invalid pairing is
rejected by Azure only at deploy time, so the spec validates it up
front. Memory pressure shows up as failed IaC operations mid-apply;
when in doubt, size the pairing up.

- default: `1Gi`

## Validation Rules

- `cpu_valid`: cpu must be one of the Consumption-plan sizes: 0.25, 0.5, 0.75, 1, 1.25, 1.5, 1.75, or 2 vCPUs
- `cpu_memory_combination`: cpu and memory must form a valid Consumption-plan pairing -- memory is always cpu x 2: 0.25/0.5Gi, 0.5/1Gi, 0.75/1.5Gi, 1/2Gi, 1.25/2.5Gi, 1.5/3Gi, 1.75/3.5Gi, 2/4Gi

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzurePlantonRunner, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.container_app_id` | `string` | The Azure resource ID of the Container App keeping the runner running. The primary handle for inspecting the appliance with Azure tooling. |
| `status.outputs.container_app_name` | `string` | The Container App's name (metadata.name). |
| `status.outputs.token_secret_name` | `string` | The name of the Container App secret holding the runner token. The container reads PLANTON_RUNNER_TOKEN from this secret; the token authorizes joining and is never the runner's identity. |
| `status.outputs.runner_name` | `string` | The name the runner registers itself under with the control plane -- the value shown by `planton runner list` the moment it joins. |
| `status.outputs.resource_group_name` | `string` | The resource group the runner was deployed in. Echoed so downstream tooling and verifiers can target the correct group. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.containerAppEnvironmentId` | AzureContainerAppEnvironment | `status.outputs.environment_id` |

## See Also

- [Overview](../README.md)

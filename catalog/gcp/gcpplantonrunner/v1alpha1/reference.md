# GcpPlantonRunner

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpPlantonRunnerSpec defines a standing Planton runner appliance in your
GCP project: an always-on worker that receives deploy operations from
the Planton control plane and executes them from WITHIN your project's
network perimeter.

Why it exists: some targets are reachable only from inside the network --
the canonical case is a private GKE control plane or a private-IP
Cloud SQL instance, which no hosted runner fleet can reach. Placing a
runner beside them makes those targets deployable and operable (day-2
updates, destroys, CloudOps) with zero inbound exposure: the runner only
ever dials OUT to the control plane, and the deployed service accepts no
meaningful inbound traffic.

ENROLLMENT IS TOKEN-FIRST: the runner is born with a runner TOKEN, never
an identity. On first boot it presents the token to the control plane,
registers ITSELF, and receives its own individually revocable identity.
Instance replacement re-joins with the same token (the token's lineage
re-admits the runner it originally admitted -- no other token can). The
token lives in a module-created Secret Manager secret; the service reads
it as a secret-backed environment variable, never as plaintext in any
launch configuration.

The spec models intent -- where the runner lives (project/region/VPC),
how big it is (cpu/memory), which build it runs (runner_version), and
the token it joins with. The compute substrate is an implementation
detail of the IaC modules (a Cloud Run service pinned to exactly one
always-on instance); it deliberately has no representation here.

## Example

```yaml
# Minimal GcpPlantonRunner manifest for local module testing. The token
# value below is an obviously-fake placeholder with the right shape --
# real deployments supply a managed-secret reference ($secret/<slug>)
# that the platform fills with a runner token before the infrastructure
# applies. The token authorizes joining and is never the runner's
# identity: the runner registers itself on first boot and receives its
# own individually revocable identity. project_id is omitted (ambient
# project).
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpPlantonRunner
metadata:
  name: gcpplantonrunner-demo
spec:
  region: us-central1
  token: prt_FAKE_PLACEHOLDER_VALUE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.region` | `string` | yes |  |  |
| `spec.token` | `string` (sensitive) | yes |  |  |
| `spec.controlPlaneEndpoint` | `string` |  |  |  |
| `spec.runnerVersion` | `string` |  | `latest` |  |
| `spec.imageRepository` | `string` |  | `ghcr.io/plantonhq/planton/runner` |  |
| `spec.serviceAccount` | `string \| valueFrom` |  |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.vpcAccess` | `GcpPlantonRunnerVpcAccess` |  |  |  |
| `spec.vpcAccess.network` | `string \| valueFrom` | yes |  | GcpVpcNetwork (`status.outputs.network_name`) |
| `spec.vpcAccess.subnetwork` | `string \| valueFrom` | yes |  | GcpSubnetwork (`status.outputs.subnetwork_name`) |
| `spec.vpcAccess.tags` | `[]string` |  |  |  |
| `spec.cpu` | `string` |  | `1` |  |
| `spec.memory` | `string` |  | `512Mi` |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project the runner is created in. Accepts a literal project ID
or a reference to a GcpProject resource. If omitted, the provider's
default project is used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.region

`string` · required

The GCP region the runner is deployed in, e.g. "us-central1". Deploy
the runner in the same region as the private endpoints it needs to
reach.

- rule: {"required":true,"string":{"pattern":"^[a-z]+-[a-z]+[0-9]+$"}}

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
plaintext; it reaches the runner through Secret Manager, not through
any launch configuration.

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
version tag for change control. New instances pull the tag on every
(re)start.

- default: `latest`

### spec.imageRepository

`string` · optional (explicit presence)

The container image repository the runner is pulled from. Override
only for air-gapped or mirrored registries hosting a copy of the
official image; the digest-identical mirror is your responsibility.

- default: `ghcr.io/plantonhq/planton/runner`

### spec.serviceAccount

`string | valueFrom`

The service account the runner runs as -- its GCP identity when cloud
operations or IaC runs use keyless access instead of injected keys.
Accepts a literal email or a reference to a GcpServiceAccount
resource. When unset, the deployment creates a dedicated
permissionless service account so the identity seam always exists and
permissions can be granted later without replacing the runner --
deliberately never the project's Compute Engine default (which
typically carries broad project access the runner should not inherit).
Either way, the module grants exactly that account read access to the
runner's own token secret, nothing else.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.vpcAccess

`GcpPlantonRunnerVpcAccess`

Private networking for OUTBOUND traffic: routes the runner's egress
into a VPC through Direct VPC egress, which is what lets it reach
private endpoints (a private GKE control plane, a private-IP
database). Only traffic to private ranges rides the VPC; the runner's
control-plane dial-out keeps its normal internet path. Omit when the
runner only needs to reach public endpoints.

### spec.vpcAccess.network

`string | valueFrom` · required

The VPC network. Accepts a literal name or a GcpVpcNetwork reference.

- references: GcpVpcNetwork (`status.outputs.network_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_name}} -- a bare string does not parse

### spec.vpcAccess.subnetwork

`string | valueFrom` · required

The subnetwork the runner draws IPs from. Must be in the runner's
region.

- references: GcpSubnetwork (`status.outputs.subnetwork_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSubnetwork, name: <that resource's name>, fieldPath: status.outputs.subnetwork_name}} -- a bare string does not parse

### spec.vpcAccess.tags

`[]string`

Network tags applied to the runner's traffic -- how VPC firewall
rules select its egress (e.g. a rule admitting the runner to a
private GKE control plane by tag).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","repeated":{"unique":true,"items":{"string":{"pattern":"^[a-z]([-a-z0-9]*[a-z0-9])?$"}}}}

### spec.cpu

`string` · optional (explicit presence)

CPU allocated to the runner instance, in vCPUs. The default 1
comfortably runs the runner's control loops plus typical IaC
operations; size up for large stacks or high operation concurrency.
Cloud Run admits 1, 2, 4, 6, or 8 vCPUs, and each pairs with a
minimum memory (see memory).

- default: `1`

### spec.memory

`string` · optional (explicit presence)

Memory allocated to the runner instance, e.g. "512Mi" or "2Gi". The
default 512Mi pairs with the default cpu of 1. Cloud Run requires at
least 2Gi for 4 vCPUs and 4Gi for 6 or 8 vCPUs. Memory pressure shows
up as failed IaC operations mid-apply; when in doubt, size memory up
before cpu.

- default: `512Mi`

## Validation Rules

- `cpu_valid`: cpu must be one of the Cloud Run instance sizes: 1, 2, 4, 6, or 8 vCPUs
- `memory_format`: memory must be a size like "512Mi" or "2Gi"

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpPlantonRunner, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.service_name` | `string` | The fully qualified name of the Cloud Run service keeping the runner running (projects/{project}/locations/{region}/services/{name}). The primary handle for inspecting the appliance with GCP tooling. |
| `status.outputs.service_short_name` | `string` | The service's short name (metadata.name). |
| `status.outputs.service_account_email` | `string` | The email of the service account the runner runs as -- the identity the runner holds while executing work. Grant this account roles to let keyless cloud operations run through the runner (it is the referenced service_account when one was supplied, else the permissionless account created with the appliance). |
| `status.outputs.token_secret_id` | `string` | The Secret Manager secret holding the runner token (projects/{project}/secrets/{name}). The token authorizes joining and is never the runner's identity. |
| `status.outputs.runner_name` | `string` | The name the runner registers itself under with the control plane -- the value shown by `planton runner list` the moment it joins. |
| `status.outputs.project_id` | `string` | The GCP project the runner was deployed in. Echoed so downstream tooling and verifiers can target the correct project. |
| `status.outputs.region` | `string` | The GCP region the runner was deployed in. Echoed so downstream tooling and verifiers can target the correct region. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.serviceAccount` | GcpServiceAccount | `status.outputs.email` |
| `spec.vpcAccess.network` | GcpVpcNetwork | `status.outputs.network_name` |
| `spec.vpcAccess.subnetwork` | GcpSubnetwork | `status.outputs.subnetwork_name` |

## See Also

- [Overview](../README.md)

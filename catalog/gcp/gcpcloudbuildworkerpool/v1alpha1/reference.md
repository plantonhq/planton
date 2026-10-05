# GcpCloudBuildWorkerPool

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpCloudBuildWorkerPoolSpec declares a Cloud Build private worker pool
(`google_cloudbuild_worker_pool`): dedicated build machines in one
region that builds run on instead of Google's shared default pool.

A private pool is what a build needs to reach things the default pool
cannot: a private package index, an internal artifact store, a database
it migrates, a GKE control plane on a private endpoint. It also lets a
team pick larger machines or nested virtualization. Many triggers,
cloudbuild.yaml files (options.pool.name), and Cloud Deploy targets
(execution environments) share one pool, so the pool is its own block
and they reference its name output.

Networking is one of two arms, or neither:

  - network_config peers the pool's workers into one of your VPC
    networks (it needs a Service Networking connection on that network
    first: a GcpServiceNetworkingConnection);
  - private_service_connect attaches each worker to a network attachment
    in the pool's region (a PSC interface), optionally routing all
    traffic through it.

With neither, workers sit on Google's service producer network with
public egress.

Important behavioral notes:

  - location, worker_pool_id, and both network arms are create-time
    decisions: changing one replaces the pool.
  - worker_config (machine type, disk size, nested virtualization,
    public IPs) and display_name update in place.
  - A pool bills per build minute on its machine type; an idle pool
    costs nothing.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudBuildWorkerPool
metadata:
  name: private-builds
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  workerPoolId: private-builds
  displayName: Private builds
  annotations:
    owner: platform
  networkConfig:
    peeredNetwork:
      value: projects/123456789012/global/networks/ci-vpc
    peeredNetworkIpRange: /26
  workerConfig:
    machineType: e2-standard-4
    diskSizeGb: 200
    noExternalIp: true
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.location` | `string` | yes |  |  |
| `spec.workerPoolId` | `string` |  |  |  |
| `spec.displayName` | `string` |  |  |  |
| `spec.annotations` | `map<string, string>` |  |  |  |
| `spec.networkConfig` | `GcpCloudBuildWorkerPoolNetworkConfig` |  |  |  |
| `spec.networkConfig.peeredNetwork` | `string \| valueFrom` | yes |  | GcpVpcNetwork (`status.outputs.network_self_link`) |
| `spec.networkConfig.peeredNetworkIpRange` | `string` |  |  |  |
| `spec.privateServiceConnect` | `GcpCloudBuildWorkerPoolPrivateServiceConnect` |  |  |  |
| `spec.privateServiceConnect.networkAttachment` | `string` | yes |  |  |
| `spec.privateServiceConnect.routeAllTraffic` | `bool` |  |  |  |
| `spec.workerConfig` | `GcpCloudBuildWorkerPoolWorkerConfig` |  |  |  |
| `spec.workerConfig.machineType` | `string` |  |  |  |
| `spec.workerConfig.diskSizeGb` | `int32` |  |  |  |
| `spec.workerConfig.noExternalIp` | `bool` |  |  |  |
| `spec.workerConfig.enableNestedVirtualization` | `bool` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The project the pool lives in: a literal project ID or a GcpProject
reference. Empty means the provider's default project. Immutable.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.location

`string` · required

The region the pool's workers run in, e.g. "us-central1". Builds that
use the pool run here, and a trigger naming it must be in the same
region. Required. Immutable.

- rule: {"required":true}

### spec.workerPoolId

`string`

The pool's ID, unique in the project and region. Defaults to
metadata.name. Immutable.

- rule: worker_pool_id must be 1-63 lowercase letters, digits, or hyphens, starting with a letter and not ending with a hyphen

### spec.displayName

`string`

A human-readable name shown in the console, 1-63 characters.

- rule: {"string":{"maxLen":"63"}}

### spec.annotations

`map<string, string>`

Annotations on the pool (Google's AIP-128 key/value metadata; not
labels -- the pool has none). Only the keys declared here are managed.

### spec.networkConfig

`GcpCloudBuildWorkerPoolNetworkConfig`

Peer the pool's workers into one of your VPC networks. Requires a
Service Networking (private services access) connection on that
network. Mutually exclusive with private_service_connect. Immutable.

### spec.networkConfig.peeredNetwork

`string | valueFrom` · required

The VPC network the workers are peered to: a GcpVpcNetwork reference
or a literal projects/{project}/global/networks/{name}. Google requires
the project NUMBER in that path; both modules resolve a project ID to
its number (one project lookup at plan time). The network must already
have private services access configured. Required. Immutable.

- references: GcpVpcNetwork (`status.outputs.network_self_link`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_self_link}} -- a bare string does not parse

### spec.networkConfig.peeredNetworkIpRange

`string`

The subnet range within the peered network the workers take addresses
from, in CIDR notation: a prefix alone ("/26") lets Google pick the
block, an address with a prefix ("192.168.0.0/29") pins it. Empty
means "/24". Immutable.

- rule: peered_network_ip_range must be CIDR notation such as /24 or 192.168.0.0/29

### spec.privateServiceConnect

`GcpCloudBuildWorkerPoolPrivateServiceConnect`

Attach each worker to a Private Service Connect network attachment in
the pool's region. Mutually exclusive with network_config. Immutable.

### spec.privateServiceConnect.networkAttachment

`string` · required

The network attachment each worker's interface connects to, as
projects/{project}/regions/{region}/networkAttachments/{name}, in the
pool's region. Required. Immutable.

- rule: network_attachment must be projects/{project}/regions/{region}/networkAttachments/{name}
- rule: {"string":{"minLen":"1"}}

### spec.privateServiceConnect.routeAllTraffic

`bool`

Route ALL worker traffic through the PSC interface, for full control
of egress (configure Cloud NAT on the attachment's subnet to reach the
internet). False routes only private ranges (10.0.0.0/8,
172.16.0.0/12, 192.168.0.0/16) through it. Immutable.

### spec.workerConfig

`GcpCloudBuildWorkerPoolWorkerConfig`

The workers' machine shape and public addressing. Omitted fields keep
Cloud Build's defaults (n1-standard-1, a standard disk, public IPs).

### spec.workerConfig.machineType

`string`

The worker's machine type, e.g. "e2-standard-4" or "n1-highcpu-8".
Empty means n1-standard-1.

### spec.workerConfig.diskSizeGb

`int32`

The worker's disk size in GB, up to 1000. 0 means Cloud Build's
standard disk size.

- rule: {"int32":{"lte":1000,"gte":0}}

### spec.workerConfig.noExternalIp

`bool` · optional (explicit presence)

Run workers without public IP addresses, which blocks egress to public
IPs (pair with network_config or private_service_connect and Cloud NAT
when builds still need the internet). Unset keeps Cloud Build's
default (public IPs); set false explicitly to restore them.

### spec.workerConfig.enableNestedVirtualization

`bool` · optional (explicit presence)

Enable nested virtualization on the workers, for builds that run VMs
or emulators, when the machine type supports it. Unset keeps Cloud
Build's default (off).

### spec.deletionPolicy

`string`

What destroy does:
  "" / "DELETE" -- the pool is deleted
  "PREVENT"     -- destroy fails
  "ABANDON"     -- the pool leaves management and stays in Google

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `spec.one_network_arm`: set at most one of network_config or private_service_connect

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpCloudBuildWorkerPool, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name: projects/{project}/locations/{location}/workerPools/{worker_pool_id}. |
| `status.outputs.worker_pool_id` | `string` | The pool's ID. |
| `status.outputs.state` | `string` | The pool's state: CREATING, RUNNING, UPDATING, DELETING, or DELETED. |
| `status.outputs.uid` | `string` | Google's unique identifier for the pool. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.networkConfig.peeredNetwork` | GcpVpcNetwork | `status.outputs.network_self_link` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpCloudBuildTrigger | `spec.build.options.workerPool` | `status.outputs.name` |
| GcpCloudFunction | `spec.buildConfig.workerPool` | `status.outputs.name` |
| GcpCloudRun | `spec.buildConfig.workerPool` | `status.outputs.name` |
| GcpDeployTarget | `spec.executionConfigs[].workerPool` | `status.outputs.name` |
| GcpDeployTarget | `spec.executionConfigs[].privatePool.workerPool` | `status.outputs.name` |
| GcpVertexAiAgentEngine | `spec.agent.buildSpec.workerPool` | `status.outputs.name` |

## See Also

- [Overview](../README.md)

# GCP Cloud Build Worker Pool

Declares a Cloud Build private worker pool: dedicated build machines in one region that builds run on instead of Google's shared default pool. A pool can peer its workers into one of your VPC networks or attach them through Private Service Connect, so builds reach private services; it sets the machine type, disk size, nested virtualization, and whether workers have public IPs.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `cloudbuild.googleapis.com` on the pool's project (never disabled on destroy)
- **Worker pool** -- one `cloudbuild_worker_pool`

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with `roles/cloudbuild.workerPoolOwner` (or the permissions in `iac/permissions.yaml`) on the pool's project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Optional Dependencies

- **`GcpVpcNetwork`** -- the network a peered pool joins (`networkConfig.peeredNetwork`), with a **`GcpServiceNetworkingConnection`** on it first.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudBuildWorkerPool
metadata:
  name: private-builds
spec:
  projectId:
    value: acme-ci
  location: us-central1
  networkConfig:
    peeredNetwork:
      valueFrom:
        kind: GcpVpcNetwork
        name: ci-vpc
        fieldPath: status.outputs.network_self_link
    peeredNetworkIpRange: /26
  workerConfig:
    machineType: e2-standard-4
    noExternalIp: true
```

```shell
planton apply -f cloud-build-worker-pool.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `location` | `string` | The region the workers run in. Immutable. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `string` / ref | provider project | The pool's project (`GcpProject` ref). Immutable. |
| `workerPoolId` | `string` | `metadata.name` | The pool's ID. Immutable. |
| `displayName` | `string` | -- | A console name, up to 63 characters. |
| `annotations` | `map` | -- | AIP-128 annotations; only declared keys are managed. |
| `networkConfig` | object | -- | `peeredNetwork` (a `GcpVpcNetwork` ref or `projects/{p}/global/networks/{n}`; the project is resolved to its number) and `peeredNetworkIpRange` (CIDR, default `/24`). Immutable. |
| `privateServiceConnect` | object | -- | `networkAttachment` (`projects/{p}/regions/{r}/networkAttachments/{n}`, in the pool's region) and `routeAllTraffic`. Immutable. |
| `workerConfig` | object | Cloud Build defaults | `machineType` (default `n1-standard-1`), `diskSizeGb` (0-1000), `noExternalIp`, `enableNestedVirtualization`. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`. |

### Validation Rules

- At most one of `networkConfig` or `privateServiceConnect`.
- `peeredNetworkIpRange` is CIDR notation (`/26` or `192.168.0.0/29`); `networkAttachment` is a full attachment name.
- `diskSizeGb` is 0-1000; `displayName` is at most 63 characters.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `projects/{project}/locations/{location}/workerPools/{worker_pool_id}` |
| `worker_pool_id` | `string` | The pool's ID |
| `state` | `string` | `CREATING`, `RUNNING`, `UPDATING`, `DELETING`, or `DELETED` |
| `uid` | `string` | Google's unique identifier for the pool |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Networking is fixed at creation.** Changing `location`, `workerPoolId`, `networkConfig`, or `privateServiceConnect` replaces the pool; `workerConfig`, `displayName`, and `annotations` update in place.
- **The project number.** Google requires `projects/{NUMBER}/global/networks/{name}` for the peered network. A `GcpVpcNetwork` reference carries the project ID, so both modules read that project once to resolve its number.
- **Same region.** A trigger or Cloud Deploy target that names the pool must be in the pool's region.
- **Billing.** The pool has no standing charge; builds on it bill per build minute at the machine type's rate.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- **GcpCloudBuildTrigger** -- builds that run on the pool (`build.options.workerPool`)
- **GcpDeployTarget** -- Cloud Deploy jobs that run on the pool (`executionConfigs[].workerPool`)
- **GcpVpcNetwork** / **GcpServiceNetworkingConnection** -- the network a peered pool joins

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

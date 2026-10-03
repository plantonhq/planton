# GCP Planton Runner

Deploys a standing Planton runner appliance on Cloud Run -- an always-on, outbound-only worker that receives deploy operations from the control plane and executes them from within your project's network perimeter. It is the piece that makes private endpoints (most notably private GKE control planes) deployable and operable, with the project, VPC placement, and runtime service account wired through ValueFromRef. Enrollment is token-first: the runner joins with a token, registers itself, and receives its own individually revocable identity.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Cloud Run Admin API enablement** -- enables `run.googleapis.com` so a fresh project works on the first deploy; deliberately left enabled on destroy, so tearing down one runner never disables the API for everything else in the project
- **Runtime service account** -- the runner's own GCP identity, created permissionless only when `serviceAccount` does not reference an existing account; deliberately never the project's Compute Engine default
- **Secret Manager secret** (`<name>-token`) -- holds the runner token; the container reads it as a secret-backed environment variable resolved at instance start, so the token never appears in the service definition
- **Secret accessor grant** -- `roles/secretmanager.secretAccessor` granted on exactly that one secret to exactly the runtime service account, nothing else
- **Cloud Run v2 service** -- the runner container itself, pinned to exactly one always-on instance (min = max = 1), with always-allocated CPU, gen2 execution environment, and optional Direct VPC egress
- **GCP labels** -- resource labels derived from the resource's organization, environment, and name

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- the credential used to provision the appliance itself. Required.
- **Runner token** -- nothing to create by hand: the platform mints a runner token and writes it at exactly the managed-secret reference the manifest declares, before the infrastructure applies. Choose a secret slug and reference it as `$secret/<slug>` in the `token` field; never inline plaintext. The token authorizes joining and is never the runner's identity -- revoking it never touches runners it already admitted.

### GCP Project

- **Region** -- deploy the runner in the same region as the private endpoints it needs to reach.
- **VPC network and subnetwork (optional)** -- only when the runner must reach private endpoints: reference a `GcpVpcNetwork` and a `GcpSubnetwork` (in the runner's region) via `vpcAccess`. The module references them, never creates or mutates them.
- **Service account (optional)** -- an existing `GcpServiceAccount` when the runner needs GCP permissions of its own for keyless cloud access; leave unset for a created permissionless one.

## Deploy

### Console

Open the deployment store, find **GCP Planton Runner**, and click **Deploy**. The creation wizard walks you through environment and connection configuration, the region and token, VPC placement, the runtime identity, and sizing. Start from the **Regional Runner** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpPlantonRunner
metadata:
  name: gke-runner
  org: acme-corp
  env: prod
spec:
  region: us-central1
  token: $secret/gke-runner-token
```

```shell
planton apply -f runner.yaml
```

This minimal manifest deploys a single always-on worker at the default sizing (1 vCPU, 512Mi) tracking the latest runner release, in the provider's default project, with a dedicated permissionless service account and no VPC egress -- project, sizing, version pinning, VPC placement, and the runtime identity are not configured. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, the runner wires its project, network placement, and runtime identity via ValueFromRef:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: workloads
      fieldPath: status.outputs.project_id
  vpcAccess:
    network:
      valueFrom:
        kind: GcpVpcNetwork
        name: private-net
        fieldPath: status.outputs.network_name
    subnetwork:
      valueFrom:
        kind: GcpSubnetwork
        name: private-usc1
        fieldPath: status.outputs.subnetwork_name
  serviceAccount:
    valueFrom:
      kind: GcpServiceAccount
      name: runner-runtime
      fieldPath: status.outputs.email
```

The InfraPipeline resolves the dependency graph, deploys the project, network, subnetwork, and service account first, then provisions the runner with the resolved values.

## Key Configuration

These are the most important decisions when configuring the runner. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**VPC placement** -- `vpcAccess` is what gives the runner private reach: with it, private-range traffic rides Direct VPC egress into your network (the route to a private GKE control plane); without it, the runner reaches only public endpoints. Either way the control-plane dial-out keeps its normal internet path. Network `tags` are how VPC firewall rules select the runner's egress.

**Runtime identity** -- leave `serviceAccount` empty to get a dedicated permissionless account (the identity seam always exists, so permissions can be granted later without replacing the runner), or reference a `GcpServiceAccount` composed with exactly the permissions keyless cloud access needs. Never the Compute Engine default.

**Sizing** -- `cpu` must be one of Cloud Run's instance sizes (1, 2, 4, 6, or 8 vCPUs), and larger sizes carry memory minimums (2Gi at 4 vCPUs, 4Gi at 6 or 8) -- validated up front. Memory pressure shows up as failed operations mid-apply, so size memory up before cpu.

**Runner build** -- empty `runnerVersion` tracks the newest release on every instance (re)start; pin a version tag for change control. `imageRepository` is only for air-gapped or mirrored registries hosting a digest-identical copy.

**Control plane endpoint** -- leave `controlPlaneEndpoint` unset for Planton's hosted control plane; set host:port for a self-hosted instance.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** (optional) | `projectId` | `status.outputs.project_id` |
| **GcpVpcNetwork** (optional) | `vpcAccess.network` | `status.outputs.network_name` |
| **GcpSubnetwork** (optional) | `vpcAccess.subnetwork` | `status.outputs.subnetwork_name` |
| **GcpServiceAccount** (optional) | `serviceAccount` | `status.outputs.email` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `service_name` | The fully qualified Cloud Run service name | Inspecting the appliance with GCP tooling |
| `service_short_name` | The service's short name | Console and CLI lookups |
| `service_account_email` | The runner's runtime identity | Granting the runner GCP roles for keyless operations |
| `token_secret_id` | The Secret Manager secret holding the runner token | Auditing secret access; rotation tooling |
| `runner_name` | The name the runner registers itself under | Finding the runner in `planton runner list` |
| `project_id` | The GCP project the runner was deployed in | Targeting follow-up GCP operations correctly |
| `region` | The deployed region | Targeting follow-up GCP operations correctly |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Regional runner** -- the minimal appliance: one always-on worker in the region of the endpoints it operates, default sizing, latest release, permissionless identity. The right first runner for any project. Start from the **Regional Runner** preset.

**Private VPC runner** -- Direct VPC egress into the network holding private endpoints; the shape that makes a private GKE control plane deployable. The subnetwork must be in the runner's region, and the network `tags` are what firewall rules select. Start from the **Private VPC Runner** preset.

**Production hardened** -- sized up (4 vCPUs / 4Gi) with a pinned `runnerVersion` so nothing tracks `latest`; upgrades become deliberate manifest changes instead of silent restarts. Start from the **High Capacity (Production Hardened)** preset.

## Works With

- [**GCP VPC Network**](/infra-catalog/gcp-vpc-network) -- the network Direct VPC egress rides into; the placement that defines what the runner can reach
- [**GCP Subnetwork**](/infra-catalog/gcp-subnetwork) -- where the runner draws IPs when VPC egress is configured; must be in the runner's region
- [**GCP Service Account**](/infra-catalog/gcp-service-account) -- the runner's runtime identity, composed first-class with exactly the permissions its workloads need
- [**GCP GKE Cluster**](/infra-catalog/gcp-gke-cluster) -- the canonical private target: a private control plane the runner makes deployable

# GCP Cloud Run

Deploys a containerized service on Google Cloud Run v2: one or more containers per instance (sidecars are first-class), request-driven autoscaling from zero to thousands of instances, declarative traffic splitting across immutable revisions, ingress and IAM invoker controls, Direct VPC Egress, and volumes backed by Cloud SQL, Secret Manager, GCS, NFS, or scratch space. Every composition seam — the GCP project, the runtime service account, the KMS key, Cloud SQL instances, GCS buckets, VPC networks and subnetworks — accepts ValueFromRef wiring, so the service drops into an InfraChart without hand-copied identifiers.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Cloud Run v2 Service** -- a managed container service in the specified GCP project and region, configured with the provided containers (images, env vars, probes, resources), scaling bounds, concurrency, and timeout
- **Ingress Configuration** -- controls whether the service accepts traffic from all sources, internal sources only, or internal sources plus Cloud Load Balancing
- **Invoker IAM Policy** -- created when `allowUnauthenticated` is true; grants `roles/run.invoker` to `allUsers` for public access
- **VPC Access** -- created only when `vpcAccess` is configured; Direct VPC Egress network interfaces or a Serverless VPC Access connector for private connectivity
- **Volumes** -- Cloud SQL Unix sockets, Secret Manager file projections, ephemeral scratch space, GCS FUSE mounts, or NFS shares, mounted into containers by name
- **Traffic Split** -- created when the `traffic` block is populated; routes percentages and preview tags across named revisions for canary and blue/green rollouts
- **GCP Labels** -- resource metadata labels applied for tracking and billing breakdowns

Custom domains are deliberately not part of this resource -- the production-grade path is composition: a serverless network endpoint group (GcpRegionNetworkEndpointGroup) bridges the service into the load-balancer chain, with DNS managed by GcpDnsRecord.

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** where the Cloud Run service will be created. Provide the project ID directly or reference a GcpProject Cloud Resource via ValueFromRef.
- **Artifact Registry or container registry** with the container image pushed and accessible to the Cloud Run service agent.
- **Cloud Run Admin API** enabled in the target project.
- **VPC network and subnetwork** (if using Direct VPC Egress) -- the subnetwork must be in the service's region with free address space for the instance fleet.
- **Secret Manager secrets** (if referenced by env vars or volumes through `valueFromSecret`) -- the runtime service account needs `roles/secretmanager.secretAccessor` on each. A variable's `secretValue` needs neither: the component creates its secret and grants that access itself.

## Deploy

### Console

Open the deployment store, find **GCP Cloud Run**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Public API Service** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudRun
metadata:
  name: my-api
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  region: us-central1
  containers:
    - image: us-docker.pkg.dev/acme-prod-12345/registry/my-api:1.0.0
      ports:
        containerPort: 8080
      resources:
        cpu: "1"
        memory: 512Mi
        startupCpuBoost: true
  scaling:
    minInstanceCount: 0
    maxInstanceCount: 20
  allowUnauthenticated: true
```

```shell
planton apply -f cloud-run.yaml
```

This creates a publicly accessible Cloud Run service with scale-to-zero, 1 vCPU, 512Mi memory, and the Gen 2 execution environment; VPC access, volumes, and traffic splitting are not configured. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the Cloud Run service to resources deployed in the same InfraPipeline:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: production-project
      fieldPath: status.outputs.project_id
  serviceAccount:
    valueFrom:
      kind: GcpServiceAccount
      name: my-api-runtime
      fieldPath: status.outputs.email
  vpcAccess:
    networkInterfaces:
      - network:
          valueFrom:
            kind: GcpVpcNetwork
            name: production-vpc
            fieldPath: status.outputs.network_name
        subnetwork:
          valueFrom:
            kind: GcpSubnetwork
            name: production-subnet
            fieldPath: status.outputs.subnetwork_name
    egress: PRIVATE_RANGES_ONLY
```

The InfraPipeline resolves the dependency graph, deploys the project, VPC, and subnetwork first, then provisions the Cloud Run service with VPC connectivity.

## Key Configuration

These are the most important decisions when configuring a Cloud Run service. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Containers** -- The `containers` list holds the serving container plus any sidecars (log collectors, auth proxies); exactly one container may declare a `ports.containerPort` (injected as `$PORT`). Per-container `resources` set CPU/memory and the two cost levers: `cpuIdle` (request-based vs instance-based billing) and `startupCpuBoost` (faster cold starts for JIT-heavy runtimes).

**Health checking** -- Three probe types, each modeled with exactly the shape the API accepts: the startup probe (HTTP/TCP/gRPC) gates first traffic within a 240-second window, the liveness probe (HTTP/gRPC) restarts unhealthy instances, and the readiness probe (HTTP/gRPC) pulls a failing instance from serving -- and re-admits it as soon as a probe succeeds again -- without restarting it. `healthCheckDisabled` switches all probing off for workloads whose serving model breaks the probe contract.

**Deploy from source** -- `buildConfig` runs the Cloud Run functions build path: point `sourceLocation` at a Cloud Storage source archive and Cloud Build produces the serving image (optionally in a private `workerPool`, as a dedicated build `serviceAccount`, with build-time `environmentVariables`). Pair `enableAutomaticUpdates` with the containers' `baseImageUri` for managed OS/runtime patching without redeploys.

**Multi-region and front doors** -- `multiRegionSettings.regions` turns the service into one identity serving from several regions (set `region: global`). `iapEnabled` puts Google's Identity-Aware Proxy in front of the service; `defaultUriDisabled` removes the default *.run.app URL so custom-domain or load-balancer paths are the only front doors.

**Scaling and performance** -- `scaling.minInstanceCount` keeps instances warm (eliminates cold starts at idle cost) and `scaling.maxInstanceCount` caps scale-out. `maxInstanceRequestConcurrency` decides how many requests one instance serves before scale-out; `timeoutSeconds` bounds each request. `serviceScaling.scalingMode: MANUAL` pins the total instance count -- an emergency brake or load-test lever.

**Ingress and authentication** -- `ingress` controls network-level access (all traffic, internal only, or internal plus load balancer). `allowUnauthenticated` grants public access through IAM; `invokerIamDisabled` switches the IAM check off entirely (for org policies forbidding `allUsers` grants) -- set at most one. Combine internal-only ingress with authenticated access for backend microservices.

**VPC connectivity** -- Configure `vpcAccess.networkInterfaces` for Direct VPC Egress (recommended -- no connector infrastructure) or `vpcAccess.connector` for a Serverless VPC Access connector, never both. Set `egress` to `PRIVATE_RANGES_ONLY` to route only private traffic through the VPC, or `ALL_TRAFFIC` for static egress IPs via Cloud NAT.

**Traffic splitting** -- The `traffic` list makes canary and blue/green declarative: `TRAFFIC_TARGET_ALLOCATION_TYPE_REVISION` targets name a revision and a percentage (all percents must sum to 100), and a `tag` gives a target a stable `https://<tag>---<host>` preview URL. Pin `revision` names only when the traffic table routes by name.

**Execution environment** -- `EXECUTION_ENVIRONMENT_GEN2` (recommended) runs full Linux and is required for GCS/NFS volumes; `EXECUTION_ENVIRONMENT_GEN1` has faster cold starts with a gVisor-restricted syscall surface.

**Destroy semantics** -- `deletionProtection` defaults to true: destroying the resource FAILS until you set it false, because deleting a service tears down its endpoint and every revision. `deletionPolicy: ABANDON` is the softer exit — the service leaves management but keeps serving in GCP.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpServiceAccount** (optional) | `serviceAccount` | `status.outputs.email` |
| **GcpKmsKey** (optional) | `encryptionKey` | `status.outputs.key_id` |
| **GcpCloudSql** (optional) | `volumes[].cloudSqlInstance.instances` | `status.outputs.connection_name` |
| **GcpGcsBucket** (optional) | `volumes[].gcs.bucket` | `status.outputs.bucket_id` |
| **GcpServerlessVpcConnector** (optional) | `vpcAccess.connector` | `status.outputs.self_link` |
| **GcpVpcNetwork** (optional) | `vpcAccess.networkInterfaces[].network` | `status.outputs.network_name` |
| **GcpSubnetwork** (optional) | `vpcAccess.networkInterfaces[].subnetwork` | `status.outputs.subnetwork_name` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `url` | Stable serving URL of the Cloud Run service | Application configuration, API gateway routing |
| `service_name` | Name of the Cloud Run service in GCP | Serverless NEG wiring, monitoring, IAM bindings |
| `revision` | Latest ready revision name | Deployment tracking, traffic pinning |
| `location` | Region the service serves from | Regional composition (NEGs, LB backends) |
| `uid` | Server-assigned unique identifier | Audit and correlation |
| `urls` | Every serving URL, including tagged preview URLs | Canary smoke tests, preview environments |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Public API service** -- Publicly accessible with scale-to-zero, unauthenticated access, port 8080, and Gen 2 execution. Suitable for web applications, APIs, and webhooks. Start from the **Public API Service** preset.

**Private VPC-connected service** -- Internal-only ingress with IAM authentication and Direct VPC Egress for private resource access (private-IP Cloud SQL, Memorystore). Suitable for backend microservices. Start from the **Private VPC-Connected Backend** preset.

**GPU inference** -- A `nodeSelector.accelerator` (e.g. `nvidia-l4`) gives every instance a GPU for scale-to-zero model serving; containers need at least 4 CPU / 16Gi and the region needs GPU quota. Start from the **GPU Inference Service** preset.

## Works With

- [**GCP Project**](/cloud-catalog/gcp-project) -- provides the GCP project where the Cloud Run service is created
- [**GCP Service Account**](/cloud-catalog/gcp-service-account) -- provides the least-privilege runtime identity
- [**GCP Cloud SQL**](/cloud-catalog/gcp-cloud-sql) -- exposes databases as managed Unix sockets via the Cloud SQL volume
- [**GCP GCS Bucket**](/cloud-catalog/gcp-gcs-bucket) -- mounts object storage via Cloud Storage FUSE
- [**GCP VPC Network**](/cloud-catalog/gcp-vpc-network) -- provides the VPC for Direct VPC Egress connectivity
- [**GCP Subnetwork**](/cloud-catalog/gcp-subnetwork) -- provides the subnetwork instances draw IPs from
- [**GCP Region Network Endpoint Group**](/cloud-catalog/gcp-region-network-endpoint-group) -- bridges the service into the HTTPS load-balancer chain for custom domains
- [**GCP Cloud Run Job**](/cloud-catalog/gcp-cloud-run-job) -- the run-to-completion sibling for batch and scheduled work

# GCP GKE Cluster

Deploys a GKE control plane — Standard (you compose node pools) or Autopilot (GKE manages nodes) — with the full cluster-wide configuration surface: VPC-native IP allocation, private-cluster topology, master authorized networks, Dataplane V2, release channels and maintenance windows, node auto-provisioning, Workload Identity, CMEK secrets encryption, Cloud Logging/Monitoring, and the GKE-managed add-on set. On Standard clusters the compute is composed from separate `GcpGkeNodePool` resources; the cluster mode, location, and dataplane are immutable decisions made here.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **GKE Cluster** — a managed Kubernetes control plane in the specified project, at a regional location ("us-central1", replicas across three zones) or a zonal one ("us-central1-a"), on the provided VPC network and subnetwork
- **VPC-native IP Allocation** — pods and services draw from secondary ranges on the subnetwork: ranges you name, CIDRs GKE carves, or fully GKE-managed ranges
- **Private Cluster Topology** (when configured) — private nodes without public IPs, optionally a private-only control-plane endpoint, with PSC-based or peering-based (`/28`) control-plane connectivity
- **Release-Channel Upgrades** — automatic Kubernetes version management via RAPID / REGULAR / STABLE / EXTENDED, or channel NONE for self-managed versions, inside the maintenance windows and exclusions you set
- **Node Auto-Provisioning** (Standard, when enabled) — GKE creates and deletes node pools within your resource limits
- **Workload Identity Pool** — `{projectId}.svc.id.goog` for keyless KSA-to-GSA mapping (on by default)
- **Cluster Add-ons** — HTTP load balancing, HPA, and the PD CSI driver by default, plus the opt-in set (Filestore/GCS FUSE/Parallelstore/Lustre CSI, Backup for GKE, NodeLocal DNSCache, Config Connector, Stateful HA, Ray with logging/monitoring, Cloud Run, pod snapshots, agent sandbox, slice controller, Slurm)
- **Container API enablement** — `container.googleapis.com` enabled in the target project (never disabled on destroy)

On Standard clusters the default node pool is removed at create time — every node pool is a separate, first-class `GcpGkeNodePool` resource. Autopilot clusters take no node pools at all.

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** — an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A VPC network and subnetwork** — an explicit network is required (clusters on the auto-created `default` network do not compose into reviewable infrastructure). For production, plan two secondary ranges on the subnetwork — one for Pods, one for Services. Reference `GcpVpcNetwork` and `GcpSubnetwork` Infra Components via ValueFromRef.
- **Cloud NAT for private nodes** — private nodes cannot pull images from registries outside Google without it. Compose a `GcpRouterNat` on the same network.
- **For peering-based private clusters only** — a `/28` CIDR block (`privateCluster.masterIpv4CidrBlock`) that does not overlap any VPC range. PSC-based clusters (the modern default) need none.

## Deploy

### Console

Open the deployment store, find **GCP GKE Cluster**, and click **Deploy**. The creation wizard walks the decisions in the order a platform engineer makes them — cluster mode, placement, networking, control-plane access, upgrades, security, observability. Start from the **Private Production Cluster — Standard** preset in the [Presets](#presets) tab for the GCP-recommended production shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpGkeCluster
metadata:
  name: platform-cluster
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  location: us-central1
  network:
    value: "https://www.googleapis.com/compute/v1/projects/acme-prod-12345/global/networks/main-vpc"
  subnetwork:
    value: "https://www.googleapis.com/compute/v1/projects/acme-prod-12345/regions/us-central1/subnetworks/gke-subnet"
  ipAllocation:
    clusterSecondaryRangeName:
      value: pods
    servicesSecondaryRangeName:
      value: services
  datapathProvider: ADVANCED_DATAPATH
  privateCluster:
    enablePrivateNodes: true
    masterIpv4CidrBlock: "172.16.0.16/28"
  masterAuthorizedNetworks:
    cidrBlocks:
      - cidrBlock: 203.0.113.0/24
        displayName: corp-vpn
  releaseChannel: REGULAR
  maintenancePolicy:
    dailyWindow:
      startTime: "03:00"
```

```shell
planton apply -f gke-cluster.yaml
```

This creates a regional private cluster on Dataplane V2 with named secondary ranges, an API allowlist, and the REGULAR release channel. Deletion protection is on by default — a destroy plan fails until `deletionProtection: false` is set explicitly. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the cluster to infrastructure deployed in the same InfraPipeline:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: production-project
      fieldPath: status.outputs.project_id
  network:
    valueFrom:
      kind: GcpVpcNetwork
      name: main-vpc
      fieldPath: status.outputs.network_self_link
  subnetwork:
    valueFrom:
      kind: GcpSubnetwork
      name: gke-subnet
      fieldPath: status.outputs.subnetwork_self_link
  ipAllocation:
    clusterSecondaryRangeName:
      value: pods
    servicesSecondaryRangeName:
      value: services
  databaseEncryption:
    state: ENCRYPTED
    keyName:
      valueFrom:
        kind: GcpKmsKey
        name: etcd-key
        fieldPath: status.outputs.key_id
```

The InfraPipeline resolves the dependency graph, deploys the project, VPC, subnetwork, and KMS key first, then provisions the cluster with all resolved references.

## Key Configuration

These are the most important decisions when configuring a GKE cluster. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Standard vs. Autopilot** (`enableAutopilot`, immutable) — Standard gives you node pools as separate `GcpGkeNodePool` resources with full control over machines, GPUs, and Spot mixes. Autopilot hands node management to GKE and bills per pod. Converting between them means recreating the cluster.

**Regional vs. zonal** (`location`, immutable) — a region creates a regional cluster (control-plane replicas across three zones, zero-downtime upgrades); a zone creates a cheaper zonal cluster whose API pauses briefly during upgrades. Production clusters should be regional.

**Dataplane** (`datapathProvider`, immutable) — `ADVANCED_DATAPATH` (Dataplane V2, eBPF/Cilium) enforces NetworkPolicy natively and unlocks FQDN policies and transparent inter-node encryption; it is the recommendation for new clusters. The legacy dataplane needs Calico (`enableNetworkPolicy`) for policy.

**Pod/service IP ranges** (`ipAllocation`, primary ranges immutable) — name planned secondary ranges for production address governance, give CIDRs for GKE to carve, or let GKE manage the ranges for dev clusters. Underestimating the pod range is the classic GKE regret; `additionalPodRangeNames` is the post-creation escape hatch.

**Control-plane access** (`privateCluster`, `masterAuthorizedNetworks`, `controlPlaneEndpoints`) — private nodes plus a CIDR allowlist is the production posture; the DNS endpoint (`*.gke.goog`) adds IAM-authenticated access without bastions or peering.

**Release channel** (`releaseChannel`) — REGULAR (default) balances freshness and stability; RAPID for early access; STABLE/EXTENDED for conservative estates; NONE only if you own version management via `minMasterVersion`.

**Workload Identity** (`workloadIdentityEnabled`, default true) — keyless workload authentication through `{projectId}.svc.id.goog`. Disabling it forces workloads back to node scopes or exported keys — almost never right.

**Deletion protection** (`deletionProtection`, default true) — both IaC engines refuse to destroy the cluster while it is on. `deletionPolicy` layers under it: PREVENT fails any destroying plan outright, ABANDON removes the cluster from state without touching it in GCP.

**Secrets add-ons** (`enableSecretManagerCsi` + `secretManagerRotation`, `secretSync`) — mount Secret Manager secrets as CSI volumes with a rotation cadence, or sync them into Kubernetes Secret objects; two add-ons, two delivery models.

**Bring-your-own trust** (`userManagedKeys`) — for regulated estates that must own the control plane's trust chain: customer CA pools for the cluster/etcd/aggregation domains, KMS-held control-plane disk encryption, and ServiceAccount JWT signing keys with rotation-friendly verification-key lists.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpVpcNetwork** | `network` | `status.outputs.network_self_link` |
| **GcpSubnetwork** | `subnetwork` | `status.outputs.subnetwork_self_link` |
| **GcpSubnetwork** | `ipAllocation.clusterSecondaryRangeName` / `servicesSecondaryRangeName` | `status.outputs.secondary_ranges.[*].range_name` |
| **GcpSubnetwork** | `privateCluster.privateEndpointSubnetwork` | `status.outputs.subnetwork_self_link` |
| **GcpKmsKey** | `databaseEncryption.keyName`, NAP `bootDiskKmsKey` | `status.outputs.key_id` |
| **GcpServiceAccount** | NAP `autoProvisioningDefaults.serviceAccount` | `status.outputs.email` |
| **GcpPubSubTopic** | `notificationPubsub.topic` | `status.outputs.topic_id` |
| **GcpBigQueryDataset** | `resourceUsageExport.bigqueryDatasetId` | `status.outputs.dataset_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `endpoint` | Kubernetes API server endpoint (private IP for private-only clusters) | Kubernetes Provider Connections, kubectl configuration |
| `cluster_ca_certificate` | Base64-encoded cluster CA certificate (public trust material) | TLS verification in Provider Connections and CI/CD pipelines |
| `workload_identity_pool` | Workload Identity Pool (`{project}.svc.id.goog`); empty when WI is disabled | IAM bindings for KSA-to-GSA mappings |
| `cluster_id` | Fully qualified resource ID (`projects/{p}/locations/{l}/clusters/{n}`) | Dataproc on GKE references, monitoring dashboards |
| `name` | The cluster name as created in GCP | `GcpGkeNodePool.clusterName` references |
| `location` | The cluster's region or zone, as provided | `GcpGkeNodePool.location` references |
| `self_link` | Server-defined URL of the cluster resource | Automation and audit tooling |
| `master_version` | Kubernetes version running on the control plane | Version dashboards, upgrade automation |
| `fleet_membership` | The fleet membership Google created through `fleetProject`; empty without a fleet | `GcpGkeFleetScope.membershipBindings`, `GcpGkeFleetFeature.membershipConfigs` |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Private standard cluster** — regional, private nodes, Dataplane V2, named secondary ranges, an API allowlist, a daily maintenance window, and the free Security Posture tiers. The GCP-recommended production shape. Start from the **Private Production Cluster — Standard** preset.

**Autopilot cluster** — GKE manages nodes entirely and bills per pod; no node pools to size or upgrade. Start from the **Autopilot Cluster** preset.

**Dev zonal cluster** — the smallest, cheapest shape: a zonal control plane with GKE-managed IP ranges. Start from the **Development Zonal Cluster** preset.

## Works With

- [**GCP Project**](/infra-catalog/gcp-project) — provides the GCP project where the cluster is created
- [**GCP VPC Network**](/infra-catalog/gcp-vpc-network) — provides the VPC network for cluster networking
- [**GCP Subnetwork**](/infra-catalog/gcp-subnetwork) — provides the subnetwork with secondary IP ranges for Pod and Service allocation
- [**GCP GKE Node Pool**](/infra-catalog/gcp-gke-node-pool) — the compute for Standard clusters, referencing this cluster's `name` and `location` outputs
- [**GCP Router NAT**](/infra-catalog/gcp-router-nat) — provides internet egress for private nodes via Cloud NAT
- [**GCP KMS Key**](/infra-catalog/gcp-kms-key) — provides the CMEK key for Kubernetes secrets encryption

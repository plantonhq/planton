# GKE Node Pools: From One-Size-Fits-All to Production Heterogeneity

## Introduction

For years, Kubernetes clusters were homogeneous—every node identical in size and capability. This "one-size-fits-all" approach led to a painful reality: either your frontend pods wasted resources on oversized machines, or your machine learning workloads couldn't schedule because the nodes were too small. The infrastructure bin-packing problem was real, and expensive.

**GKE node pools** solved this by introducing the concept of heterogeneous compute within a single cluster. A node pool is a group of Compute Engine VMs that all share the same configuration—machine type, disk size, service account, and scheduling constraints. Multiple node pools within one cluster means you can finally match infrastructure capabilities directly to workload requirements.

This isn't just about resource efficiency. Production GKE clusters leverage multiple node pools for:
- **Cost optimization**: Dedicated pools using Spot VMs for fault-tolerant batch workloads (up to 91% cost savings)
- **Specialized hardware**: GPU/TPU pools for machine learning, isolated from general workloads
- **Security boundaries**: Separate pools with distinct service accounts and taints for multi-tenant isolation
- **Performance tiers**: Compute-optimized pools (C2 series) for CPU-bound apps, memory-optimized pools (M2/M3) for caches and databases

The key architectural decision: **node pools are separate resources from the cluster**. They reference their parent GKE cluster but have independent lifecycles. You can create, update, and destroy node pools without ever touching the cluster's control plane—a critical pattern for production stability.

This document explores the maturity spectrum of GKE node pool deployment methods, from manual click-ops to declarative GitOps, and explains Planton's approach to managing this fundamental building block of GKE infrastructure.

## The Deployment Maturity Spectrum

### Level 0: Manual Console Operations (The Discovery Phase)

**What it is**: The Google Cloud Console's "Add Node Pool" button. You navigate to your GKE cluster, click through a web form, fill in machine type, disk size, autoscaling ranges, and hit "Create."

**When it's appropriate**: Learning GKE, exploring configuration options, or performing one-off emergency scaling operations.

**Why it doesn't scale**: This is pure click-ops. It's not repeatable, not auditable, and produces configuration drift the moment someone makes a manual change. Every production incident story about "who changed that setting" starts here.

**Verdict**: Essential for learning, unacceptable for production infrastructure.

### Level 1: Imperative CLI Scripting (The First Automation)

**What it is**: Using `gcloud` commands to create and modify node pools:

```bash
gcloud container node-pools create prod-pool \
  --cluster=my-cluster \
  --machine-type=n2-standard-4 \
  --disk-size=100 \
  --disk-type=pd-ssd \
  --enable-autoscaling \
  --min-nodes=3 \
  --max-nodes=10 \
  --region=us-central1
```

**The advancement**: It's scriptable. You can wrap it in bash, version control it, run it in CI/CD pipelines.

**The limitation**: It's *imperative*, not *declarative*. The `gcloud` CLI tells GCP "do this action" but doesn't manage state or handle drift. If someone manually changes a node pool setting through the console, your script doesn't know. Idempotency is manual—you have to write logic to check "does this pool exist with these settings?" Updates require separate `gcloud container node-pools update` commands, and you're responsible for calculating the delta between current and desired state.

**Verdict**: Good for simple automation and CI/CD tasks, but lacks the state management and drift detection required for complex infrastructure.

### Level 2: Infrastructure as Code (The Production Standard)

This is where production infrastructure management begins. IaC tools are **declarative** (you define desired state), **stateful** (they track what exists), and **lifecycle-aware** (they compute diffs and execute only necessary changes).

#### Terraform

The industry standard. You define a `google_container_node_pool` resource in HCL:

```hcl
resource "google_container_node_pool" "prod_pool" {
  name       = "prod-general"
  cluster    = google_container_cluster.primary.id
  location   = "us-central1"
  
  node_config {
    machine_type = "n2-standard-4"
    disk_size_gb = 100
    disk_type    = "pd-ssd"
    
    service_account = google_service_account.gke_nodes.email
    oauth_scopes = ["https://www.googleapis.com/auth/cloud-platform"]
  }
  
  autoscaling {
    min_node_count = 3
    max_node_count = 10
    location_policy = "BALANCED"
  }
  
  management {
    auto_repair  = true
    auto_upgrade = true
  }
}
```

On `terraform apply`, Terraform compares the desired state (your HCL) against its state file and the real-world GCP resources, then executes only the necessary API calls. If you change `max_node_count` from 10 to 15, Terraform knows to call the GKE API's update method, not recreate the entire pool.

**The critical pattern**: Node pools are **separate resources**, not inline blocks within the cluster definition. Terraform documentation explicitly warns that inline node pools can cause IaC tools to "struggle with complex changes" and risk triggering unintended cluster-level operations. This validates Planton's design of `GcpGkeNodePool` as an independent resource.

#### Pulumi

Modern alternative using general-purpose languages (Python, TypeScript, Go):

```python
node_pool = gcp.container.NodePool("prod-pool",
    cluster=cluster.name,
    location="us-central1",
    node_config=gcp.container.NodePoolNodeConfigArgs(
        machine_type="n2-standard-4",
        disk_size_gb=100,
        disk_type="pd-ssd",
    ),
    autoscaling=gcp.container.NodePoolAutoscalingArgs(
        min_node_count=3,
        max_node_count=10,
        location_policy="BALANCED",
    ),
)
```

Same declarative, stateful model as Terraform, but with the power of real programming languages for complex logic, loops, and abstractions.

#### OpenTofu

Open-source Terraform fork, using identical HCL syntax. A response to HashiCorp's license changes, providing the same capabilities with community governance.

**Verdict**: This is the production standard. Terraform and Pulumi are the dominant tools for managing GCP infrastructure at scale—repeatable, auditable, state-managed, and drift-detecting.

### Level 3: Kubernetes-Native GitOps (The Unified Control Plane)

The frontier of infrastructure management: using the Kubernetes API itself to manage cloud resources.

#### Config Connector (Google's First-Party Solution)

Config Connector is a Kubernetes operator that introduces Google Cloud CRDs. You define infrastructure as Kubernetes manifests:

```yaml
apiVersion: container.cnrm.cloud.google.com/v1beta1
kind: ContainerNodePool
metadata:
  name: prod-pool
spec:
  clusterRef:
    name: my-cluster
  location: us-central1
  autoscaling:
    minNodeCount: 3
    maxNodeCount: 10
  nodeConfig:
    machineType: n2-standard-4
    diskSizeGb: 100
    diskType: pd-ssd
```

You `kubectl apply` this manifest, and the Config Connector operator running in your cluster makes the GCP API calls to create or update the node pool. Changes are tracked in Git, deployed via ArgoCD or Flux, and managed with the same tools you use for applications.

**The paradigm shift**: Your Kubernetes cluster becomes self-provisioning. Infrastructure and applications share a unified control plane.

#### Crossplane (Vendor-Neutral Multi-Cloud)

Crossplane extends the pattern with multi-cloud support and composability. Platform teams can define abstract, high-level resources (e.g., `XGkeCluster`) that compose and encapsulate lower-level resources (cluster, multiple node pools, databases), presenting a simplified API to developers.

**The strategic advantage**: GitOps workflows, unified tooling (`kubectl`, ArgoCD), and the ability to treat infrastructure as just another set of Kubernetes resources. This solves the "who provisions the cluster that provisions the infrastructure" problem by making clusters self-managing.

**Verdict**: The cutting edge. Ideal for platform engineering teams building internal developer platforms, but requires deep Kubernetes expertise and introduces operational complexity (the operator must be highly available and correctly configured).

## The Planton Choice

Planton adopts the **separate resource pattern** validated by Terraform, Pulumi, and production best practices: `GcpGkeNodePool` is an independent resource that references its parent `GcpGkeCluster`.

### Why This Design

1. **Production stability**: Node pool updates (changing machine types, disk sizes, or labels) never risk triggering unintended cluster-level operations.
2. **Lifecycle independence**: Create, modify, and delete node pools without touching the cluster's control plane.
3. **Industry alignment**: This mirrors the dominant pattern in Terraform, Pulumi, and GKE's own API design.

### The 80/20 API Philosophy

The `GcpGkeNodePoolSpec` follows an **opinionated but extensible** design, prioritized by real-world usage:

**Essential (80%)**: Fields that define a functional node pool
- Cluster reference: `cluster_project_id`, `cluster_name` (foreign key to parent cluster)
- Machine shape: `machine_type` (default: `e2-medium`), `disk_size_gb`, `disk_type` (default: `pd-standard`)
- Size: Either fixed `node_count` or `autoscaling` (min/max nodes)—enforced as mutually exclusive via `oneof`
- Cost optimization: `spot` (boolean for Spot VMs)

**Common production (15%)**: Reliability and scheduling
- Node management: `management.disable_auto_upgrade`, `management.disable_auto_repair` (both default to enabled)
- Scheduling: `node_labels` for Kubernetes `nodeSelector`
- Security: `service_account` for custom IAM service accounts
- Node OS: `image_type` (default: `COS_CONTAINERD`)

**Advanced (5%)**: Specialized use cases
- Taints for workload isolation (roadmap: not yet in current spec)
- GPU/TPU accelerators (roadmap: not yet in current spec)
- Advanced kernel tuning (sysctls, kubelet config)

This structure ensures that simple use cases are simple (3-4 required fields), while production use cases are supported without cluttering the API surface for the majority.

### Example Configurations

**Dev/Test Pool (Cost-Optimized with Spot VMs)**:
```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpGkeNodePool
metadata:
  name: dev-pool
spec:
  cluster_project_id: my-project
  cluster_name: dev-cluster
  machine_type: e2-medium
  disk_size_gb: 50
  spot: true
  autoscaling:
    min_nodes: 0      # Scale-to-zero for cost savings
    max_nodes: 3
    location_policy: ANY  # Hunt for capacity across zones
```

**Production General-Purpose Pool (High Availability)**:
```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpGkeNodePool
metadata:
  name: prod-general
spec:
  cluster_project_id: my-project
  cluster_name: prod-cluster
  machine_type: n2-standard-4
  disk_size_gb: 100
  disk_type: pd-ssd
  service_account: gke-prod-sa@my-project.iam.gserviceaccount.com
  autoscaling:
    min_nodes: 3      # Multi-zone HA baseline
    max_nodes: 10
    location_policy: BALANCED  # Even spread for availability
  management:
    disable_auto_upgrade: false  # Keep security patches current
    disable_auto_repair: false   # Auto-fix unhealthy nodes
```

**Production GPU Pool (Machine Learning Workloads)**:
```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpGkeNodePool
metadata:
  name: gpu-pool
spec:
  cluster_project_id: my-project
  cluster_name: prod-cluster
  machine_type: n1-standard-8
  disk_size_gb: 200
  disk_type: pd-ssd
  node_labels:
    workload-type: gpu
  autoscaling:
    min_nodes: 1
    max_nodes: 5
  # Note: GPU accelerator config is roadmap item
```

## Production Best Practices

### Right-Sizing: Multiple Specialized Pools vs. One Large Pool

**Anti-pattern**: A single, massive node pool with one machine type.

**Why it fails**: Resource fragmentation. Small pods leave nodes underutilized. Large pods can't schedule despite sufficient *total* cluster capacity because no single node is big enough. This is the classic "bin-packing" problem.

**Best practice**: Create **multiple, specialized node pools** aligned to workload profiles:
- **General-purpose pool**: E2 or N2 series for standard web applications
- **High-memory pool**: M2/M3 series for Redis, in-memory databases
- **Compute-optimized pool**: C2 series for CPU-bound workloads
- **Cost-optimized pool**: E2 Spot VMs for batch jobs, CI/CD runners
- **Accelerator pool**: N1 with GPUs for ML inference and training

### Autoscaling: Location Policies Matter

When enabling autoscaling in regional clusters, choose between:
- **BALANCED** (default for on-demand pools): Spreads nodes evenly across zones. Best for high-availability production workloads.
- **ANY** (recommended for Spot VMs): Adds nodes wherever capacity is available. Critical for Spot VM pools—allows the autoscaler to "hunt" for spare capacity across all zones, dramatically increasing acquisition success.

### Spot VMs: Cost Savings with Constraints

Spot VMs offer up to 91% cost reduction but come with hard limits:
- **No SLA**: Can be preempted (terminated) by Google with short notice
- **No guarantees**: Availability is not guaranteed

**Production pattern**:
1. **Mixed pools**: *Never* run a cluster solely on Spot VMs. At least one reliable, on-demand pool must exist for critical system components (kube-dns, monitoring agents) and stateful applications.
2. **Taints and tolerations**: Apply a taint to Spot VM pools (e.g., `spot=true:NoSchedule`). Only fault-tolerant workloads (batch jobs, CI runners) get the corresponding toleration, ensuring they land on cost-effective nodes while critical workloads stay on reliable ones.
3. **Graceful shutdown**: GKE 1.20+ enables kubelet graceful shutdown by default, giving Pods time to terminate cleanly when preempted.

### Auto-Upgrade and Auto-Repair: Don't Disable Them

**Anti-pattern**: Disabling auto-upgrade and auto-repair out of fear of disruption.

**Why it backfires**: Your nodes fall behind on critical security patches. You accumulate technical debt. When you *do* upgrade, the gap is large and risky.

**Best practice**: Keep both **enabled** (the GKE default) and use **Maintenance Windows** to control *when* upgrades occur. Define recurring time windows (e.g., "Sunday 2:00-4:00 AM") during which automated upgrades are permitted. This minimizes disruption while ensuring nodes stay current.

### Security: Custom Service Accounts and Workload Identity

**Default behavior**: Nodes use the Compute Engine default service account, which is overly permissive.

**Best practice**:
1. Create a **custom, minimally-privileged IAM service account** for each node pool. Grant only the specific roles required (e.g., `roles/logging.logWriter`, `roles/monitoring.metricWriter`).
2. Use **Workload Identity** (the modern approach): Bind Kubernetes Service Accounts (KSA) directly to Google Service Accounts (GSA). Pods assume the GSA's identity without relying on broad node-level permissions, following the principle of least privilege.

## Node Pool Lifecycle: Upgrades Without Downtime

### The Cordon-and-Drain Pattern

Any change to node configuration (machine type, disk type, image, taints, service account) requires node **recreation**. GKE automates this:
1. **Cordon**: Mark node unschedulable (no new Pods)
2. **Drain**: Gracefully evict existing Pods (respecting PodDisruptionBudgets)
3. **Recreate**: Terminate old VM, launch new VM with updated config
4. **Uncordon**: New node joins cluster and becomes schedulable

**Critical requirement**: PodDisruptionBudgets (PDBs) are non-negotiable for production. Without PDBs, automated operations can forcibly terminate Pods, causing outages.

### Surge Upgrades (Default Strategy)

GKE's default rolling update strategy:
- **maxSurge** (default: 1): Creates extra "surge" nodes above pool size. With `maxSurge=1`, GKE creates one new node, migrates workloads, drains one old node, repeats. Fast and minimally disruptive, but temporarily requires quota for extra VMs.
- **maxUnavailable** (default: 0): Number of nodes that can be unavailable simultaneously. If `maxSurge=0`, you must set `maxUnavailable > 0` for "in-place" upgrades (slower, more disruptive, but no extra VM quota needed).

Most users accept the default `maxSurge=1, maxUnavailable=0` balance.

### Blue-Green Upgrades (Zero-Downtime for Critical Workloads)

For disruption-intolerant production workloads, GKE offers **blue-green upgrades**:
1. **Create green pool**: Provisions entirely new nodes with new config alongside existing "blue" pool (temporarily doubles capacity and cost)
2. **Cordon blue pool**: Old nodes stop accepting new Pods
3. **Drain blue to green**: Pods migrate gracefully to ready green nodes
4. **Soak period**: Upgrade pauses for configurable duration (up to 7 days). Old blue pool remains cordoned but available.
5. **Delete blue pool**: After soak time, GKE deletes old nodes

During soak, operators can:
- **Complete upgrade early**: Skip soak and proceed to deletion
- **Rollback**: Reverse drain (green → blue) for near-instant rollback to known-good configuration

**Trade-off**: Higher temporary cost and operational complexity, but ultimate safety for zero-downtime production changes.

## Machine Type Selection Guide

| Family | Use Case | Cost Profile | Examples |
|--------|----------|--------------|----------|
| **E2** | General purpose, cost-optimized | Low | Web servers, microservices, dev/test |
| **N2** | General purpose, performance | Medium | Production apps, databases |
| **N2D** | General purpose (AMD) | Medium | High-performance web, general workloads |
| **C2** | Compute-optimized | High | HPC, gaming, CPU-bound applications |
| **M2/M3** | Memory-optimized | Very High | Redis, in-memory databases, analytics |

## Conclusion

GKE node pools transformed Kubernetes infrastructure from homogeneous to heterogeneous, solving the resource fragmentation problem and unlocking cost optimization through targeted infrastructure choices. The architectural decision to treat node pools as independent resources—separate from their parent cluster—is the foundation of production-grade automation.

Planton's `GcpGkeNodePool` embraces this pattern, validated by Terraform, Pulumi, and real-world production practices. The API's 80/20 design ensures that simple use cases remain simple while providing the depth required for complex, multi-pool production environments running diverse workload profiles.

Whether you're creating a single dev pool with Spot VMs for cost savings or orchestrating a production cluster with specialized pools for GPUs, high-memory workloads, and batch processing, the pattern is the same: define the desired state, let the controller handle the lifecycle, and trust that node pool operations never risk disrupting the cluster's control plane.

The evolution from manual console clicks to declarative GitOps reflects a broader shift in infrastructure management—from "infrastructure as a chore" to "infrastructure as software." Node pools, as the fundamental unit of heterogeneous compute in GKE, are at the heart of that transformation.


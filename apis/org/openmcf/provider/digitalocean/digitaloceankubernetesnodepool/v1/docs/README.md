# DigitalOcean Kubernetes Node Pool Deployment

## Introduction

There's a peculiar design choice in the DigitalOcean Kubernetes (DOKS) API that reveals something fundamental about infrastructure management: the "default node pool problem." When you provision a DOKS cluster using Infrastructure-as-Code, the default node pool is defined *inline* within the cluster resource. Modify any property of that default pool—change its size, adjust its node count—and your IaC tool interprets this as a change to the *parent cluster*, triggering a plan to **destroy and recreate your entire cluster**, control plane and all.

This isn't a bug. It's a structural consequence of modeling node pools as nested resources rather than independent entities. And it teaches us the first principle of production DOKS management: **treat all node pools as separate, lifecycle-independent resources from day one**.

The community-evolved workaround is elegant: create a minimal, immutable "sacrificial default pool" (one tiny node that you never touch again) and provision all real workloads on *additional* node pools managed as standalone resources. This pattern decouples cluster lifecycle from node pool lifecycle, enabling safe resizing, upgrades, and deletion without risking the control plane.

This document explores the full landscape of DOKS node pool deployment methods—from manual UI operations to Kubernetes-native control planes—and explains why OpenMCF's abstraction treats every node pool as a first-class, cluster-independent resource. We'll examine the production-ready approaches, decode the "80/20" configuration surface, and show how labels, taints, and autoscaling combine to create robust multi-pool architectures.

## The Deployment Methods Landscape

### Level 0: Manual UI Management (ClickOps)

The **DigitalOcean Control Panel** provides a wizard-style workflow for adding node pools. Navigate to your cluster, click "Add Node Pool," select a machine type (Shared CPU, Dedicated CPU, GPU), choose a Droplet size, and specify either a fixed node count or autoscaling bounds (minimum and maximum nodes).

This interface reveals a critical design choice in the underlying API: **fixed-size and autoscaling pools are mutually exclusive operational modes**. You cannot specify both a static `node_count` and autoscaling `min/max` parameters—the UI enforces this choice, and any API abstraction must mirror this constraint.

The UI also performs hidden automations. GPU-based pools automatically receive an `nvidia.com/gpu:NoSchedule` taint, preventing non-GPU workloads from landing on expensive hardware. This is a sensible default, but it's a *default*, not a hard requirement—an API client would need to specify this explicitly or opt out.

**Limitations for production:**
- No version control or audit trail of changes
- Manual changes drift from IaC definitions immediately
- Error-prone at scale (imagine 10 clusters × 3 pools each = 30 manual operations to update)
- No integration with CI/CD pipelines
- Common pitfalls: selecting undersized Droplets (&lt;2.5GB allocatable memory) for production, or renaming pools without understanding that node names only update after recycling

**Verdict:** Acceptable for learning DOKS and emergency interventions. Risky for anything beyond personal development environments. The lack of automation and audit trails makes this approach unsuitable for production operations.

### Level 1: CLI Scripting with `doctl`

The official DigitalOcean CLI, **doctl**, provides imperative commands for node pool management:

```bash
# Create a fixed-size pool
doctl kubernetes cluster node-pool create <cluster-id> \
  --name production-workers \
  --size s-4vcpu-8gb \
  --count 3

# Create an autoscaling pool
doctl kubernetes cluster node-pool create <cluster-id> \
  --name autoscale-workers \
  --size s-4vcpu-8gb \
  --auto-scale=true \
  --min-nodes=3 \
  --max-nodes=10
```

The `doctl create` command exposes the **absolute minimal API surface**: three required flags for fixed-size pools (`--name`, `--size`, `--count`) or four for autoscaling (`--name`, `--size`, `--min-nodes`, `--max-nodes`). This represents the irreducible core of what defines a node pool.

Critically, `doctl` also supports advanced isolation flags at creation time, particularly `--taint`. This confirms that taints are a "Day 1" configuration, not a post-creation patch—essential for preventing workloads from being misscheduled onto specialized pools (GPU nodes, dedicated system pools) from the moment they come online.

**Strengths:**
- Scriptable and reproducible commands
- Shell scripts can be version-controlled
- Integration into simple CI/CD workflows
- Useful building block for higher-level automation

**But it's still imperative, not declarative:**
- Scripts must track existing state to be idempotent
- No automatic drift detection or correction
- Node pool IDs must be tracked separately for updates/deletions
- Coordinating multiple resources (clusters, pools, firewalls) becomes complex

**Verdict:** Better than pure UI clicks, suitable for small teams with simple workflows. Falls short of production requirements because you're executing commands rather than declaring desired state. The CLI is excellent as a **building block** for IaC tools, not as the primary interface for production infrastructure.

### Level 2: SDK Integration (godo, pydo)

For developers building custom automation, controllers, or platform abstractions, DigitalOcean provides official SDKs: **godo** (Go) and **pydo** (Python).

The Go `godo` library (`github.com/digitalocean/godo`) defines canonical Go structs like `NodePoolCreateRequest` and `NodePoolUpdateRequest`, which serve as the definitive, type-safe reference for all available API parameters. The Python `pydo` client provides equivalent functions: `pydo.kubernetes.add_node_pool()` and `pydo.kubernetes.delete_node_pool()`.

**When to use SDKs:**
- Building custom Kubernetes operators or controllers
- Integrating DOKS management into existing platforms
- Creating CLI tools or automation frameworks
- Needing programmatic access beyond what standard IaC tools provide

**What SDKs don't provide:**
- State management (you track what exists)
- Idempotency (you implement retries and checks)
- Drift detection (you compare desired vs actual state)
- Rollback mechanisms (you orchestrate recovery)

**Verdict:** The right choice for building platforms or custom tooling, but excessive for direct infrastructure management. Most teams benefit from using tools built **on top of** these SDKs (like Terraform or Pulumi) rather than calling them directly.

### Level 3: Infrastructure-as-Code (Production Standard)

This is where node pool management becomes reliable, auditable, and scalable. The production landscape has three mature options, each with distinct philosophical approaches.

#### Terraform with Official DigitalOcean Provider

The **DigitalOcean Terraform Provider** (`digitalocean/digitalocean`) is the most established and widely-used approach. It defines two distinct resources that mirror the API's dual nature:

1. `digitalocean_kubernetes_cluster` — defines the cluster and its *inline default node pool*
2. `digitalocean_kubernetes_node_pool` — manages *additional* pools as standalone resources

```hcl
# The "sacrificial default pool" pattern
resource "digitalocean_kubernetes_cluster" "production" {
  name    = "prod-cluster"
  region  = "nyc3"
  version = "1.28.2-do.0"
  
  # Minimal, immutable default pool
  node_pool {
    name       = "system-default"
    size       = "s-1vcpu-2gb"
    node_count = 1
  }
}

# Real workload pools as separate resources
resource "digitalocean_kubernetes_node_pool" "app_workers" {
  cluster_id = digitalocean_kubernetes_cluster.production.id
  name       = "app-workers"
  size       = "s-4vcpu-8gb"
  auto_scale = true
  min_nodes  = 3
  max_nodes  = 10
  
  labels = {
    workload = "application"
    env      = "production"
  }
}

resource "digitalocean_kubernetes_node_pool" "gpu_jobs" {
  cluster_id = digitalocean_kubernetes_cluster.production.id
  name       = "gpu-workers"
  size       = "g-4vcpu-16gb"
  auto_scale = true
  min_nodes  = 0
  max_nodes  = 2
  
  labels = {
    hardware = "gpu"
  }
  
  taint {
    key    = "nvidia.com/gpu"
    value  = "true"
    effect = "NoSchedule"
  }
}
```

**Why the "sacrificial default pool" pattern is essential:**

The GitHub issue history for the Terraform provider reveals extensive community discussion about the "buzz killer" behavior: modifying the inline `node_pool` block triggers cluster replacement. The production-grade workaround is now standard practice:

1. Create a minimal default pool (1× `s-1vcpu-2gb` node) in the cluster definition
2. **Never touch this pool again**—treat it as immutable infrastructure
3. Manage all real workloads using separate `digitalocean_kubernetes_node_pool` resources
4. This decouples cluster and pool lifecycles, enabling safe independent updates

**Production patterns with Terraform:**
- Remote state with locking (S3, Terraform Cloud, DigitalOcean Spaces)
- Separate workspaces or state files per environment (dev/staging/prod)
- Module libraries for standard node pool configurations
- `terraform plan` in CI for validation before merge
- `terraform apply` automated via GitOps tools (Atlantis, Spacelift)

**Trade-offs:**
- HCL learning curve (though shallow for basic usage)
- State file management adds operational complexity
- Manual `terraform apply` runs unless automated
- Drift detection requires explicit `plan` operations

#### Pulumi with DigitalOcean Provider

**Pulumi** offers the same declarative infrastructure model but with general-purpose programming languages (TypeScript, Python, Go, C#):

```typescript
import * as digitalocean from "@pulumi/digitalocean";

// Cluster with minimal default pool
const cluster = new digitalocean.KubernetesCluster("production", {
  region: "nyc3",
  version: "1.28.2-do.0",
  nodePool: {
    name: "system-default",
    size: "s-1vcpu-2gb",
    nodeCount: 1,
  },
});

// Real workload pool with autoscaling
const appWorkers = new digitalocean.KubernetesNodePool("app-workers", {
  clusterId: cluster.id,
  size: "s-4vcpu-8gb",
  autoScale: true,
  minNodes: 3,
  maxNodes: 10,
  labels: {
    workload: "application",
    env: "production",
  },
});

// Specialized GPU pool
const gpuWorkers = new digitalocean.KubernetesNodePool("gpu-workers", {
  clusterId: cluster.id,
  size: "g-4vcpu-16gb",
  autoScale: true,
  minNodes: 0,
  maxNodes: 2,
  labels: { hardware: "gpu" },
  taints: [{
    key: "nvidia.com/gpu",
    value: "true",
    effect: "NoSchedule",
  }],
});
```

**Why Pulumi's bridged provider matters:**

The `@pulumi/digitalocean` provider is a "bridged" provider—programmatically generated from the upstream Terraform provider. This architecture guarantees:
- **1:1 API parity** with Terraform resources
- **Same-day support** for new features added to Terraform provider
- **Identical capabilities** for labels, taints, autoscaling, and all other parameters

**Advantages over Terraform:**
- Use TypeScript, Python, Go, or C# instead of HCL
- Native loops, conditionals, functions, and type checking
- Catch configuration errors at compile time (for compiled languages)
- Easier integration with existing application codebases
- Manage infrastructure and application deployment in single program

**The unified lifecycle advantage:**

Pulumi's biggest strength for DOKS is **single-program lifecycle management**. A single `pulumi up` command can:
1. Create the DOKS cluster
2. Wait for cluster provisioning to complete
3. Extract the kubeconfig as a programmatic variable
4. Pass it to a `kubernetes.Provider` instance
5. Deploy Helm charts, Kubernetes manifests, and operators
6. All in one atomic operation

Terraform requires multiple `apply` steps or complex provider dependency chains to achieve this. Pulumi makes it natural.

**When to choose Pulumi:**
- Your team already uses TypeScript, Python, or Go
- Complex logic requires programming constructs (loops, conditionals)
- You want unified infrastructure + application deployment
- Strong typing and compile-time checks are valuable

**When to choose Terraform:**
- Larger community and ecosystem (more examples, modules)
- Team already experienced with HCL
- Preference for simpler, more constrained declarative language
- Need for specific Terraform-only tooling (Terragrunt, etc.)

#### OpenTofu: The Open Source Fork

**OpenTofu** is a community-driven fork of Terraform, created after HashiCorp's license change. It maintains full compatibility with the DigitalOcean provider and all HCL configurations.

For DOKS node pool management, OpenTofu is functionally identical to Terraform—same provider, same resources, same syntax. The choice is organizational: teams preferring truly open source tooling or avoiding vendor lock-in choose OpenTofu; teams comfortable with HashiCorp's licensing stick with Terraform.

### Level 4: Kubernetes-Native Control Planes (Crossplane)

The highest abstraction level treats infrastructure as Kubernetes resources, managed via `kubectl` and reconciled by always-on controllers.

#### Crossplane with DigitalOcean Provider

**Crossplane** is a CNCF-graduated project that transforms a Kubernetes cluster into a universal control plane. The **provider-digitalocean** installs Custom Resource Definitions (CRDs) that model DigitalOcean resources as native Kubernetes objects.

Instead of running `terraform apply` or `pulumi up`, you declare infrastructure using YAML manifests and `kubectl apply`:

```yaml
apiVersion: kubernetes.digitalocean.crossplane.io/v1alpha1
kind: DOKSCluster
metadata:
  name: production-cluster
spec:
  forProvider:
    name: prod-cluster
    region: nyc3
    version: "1.28.2-do.0"
    nodePool:
      name: system-default
      size: s-1vcpu-2gb
      count: 1
  providerConfigRef:
    name: digitalocean-prod
---
apiVersion: kubernetes.digitalocean.crossplane.io/v1alpha1
kind: DOKSNodePool
metadata:
  name: app-workers
spec:
  forProvider:
    clusterId: production-cluster
    name: app-workers
    size: s-4vcpu-8gb
    autoScale: true
    minNodes: 3
    maxNodes: 10
    labels:
      workload: application
      env: production
  providerConfigRef:
    name: digitalocean-prod
```

**Why this approach exists:**

Crossplane is not just "Terraform in Kubernetes"—it's a fundamentally different operational model:

1. **Continuous reconciliation**: Controllers continuously compare desired state (CRDs) with actual state (DigitalOcean API), automatically correcting drift
2. **GitOps-native**: Infrastructure definitions live in Git, and tools like ArgoCD or Flux synchronize them to the cluster
3. **Self-service via Kubernetes RBAC**: Teams can request node pools via Git PR, subject to admission policies and resource quotas
4. **Event-driven architecture**: Other controllers can watch for infrastructure CRDs and react (e.g., deploy monitoring when a new pool appears)

**The kubeconfig-as-Secret pattern:**

When a `DOKSCluster` CR is applied, the Crossplane controller provisions the cluster and writes the resulting kubeconfig back **as a Kubernetes Secret** in the management cluster. This creates a loosely coupled, event-driven architecture. Other controllers (like ArgoCD application controllers) simply watch for this Secret to appear, then automatically begin deploying applications to the new cluster.

**Operational complexity:**
- Requires running Crossplane and provider controllers (additional infrastructure)
- Learning curve for Crossplane's Composition and resource model
- Another system to monitor, upgrade, and maintain
- Debugging failures requires understanding controller reconciliation loops

**When to choose Crossplane:**
- Building an Internal Developer Platform (IDP) with self-service infrastructure
- Already practicing GitOps for application delivery
- Want continuous drift correction, not periodic `plan` runs
- Team comfortable with Kubernetes controllers and CRDs

**When to choose Terraform/Pulumi instead:**
- Simpler operational model (stateless CLI tools)
- Team already experienced with IaC patterns
- Don't need continuous reconciliation (periodic apply is sufficient)
- Prefer explicit apply operations over always-on controllers

## Deployment Method Comparison

| Method | Type | State Management | Idempotent | Drift Detection | Primary Use Case |
|--------|------|------------------|------------|-----------------|------------------|
| **UI (Control Panel)** | Manual (ClickOps) | N/A (remote) | No | No | Learning, emergency access, simple tasks |
| **doctl (CLI)** | Imperative scripting | None | No | No | Simple scripts, CI tasks, building blocks |
| **godo/pydo (SDKs)** | Programmatic library | None | No | No | Building custom controllers, integrations |
| **Terraform** | Declarative IaC | Local/Remote state file | Yes | Yes (on plan) | Production infrastructure provisioning |
| **Pulumi** | Declarative IaC | Cloud service / self-hosted | Yes | Yes (on preview) | Unified infra + app lifecycle management |
| **OpenTofu** | Declarative IaC | Local/Remote state file | Yes | Yes (on plan) | Open source Terraform alternative |
| **Crossplane** | Declarative K8s | In-cluster (CRD status) | Yes | Yes (continuous) | GitOps, self-service IDPs, drift correction |

## Production Essentials: Sizing, Scaling, and Lifecycle Management

### Instance Sizing and Capacity Planning

**The Golden Rule: Allocatable Memory Matters**

DigitalOcean explicitly documents the production baseline: use nodes with **2.5 GB or more of allocatable memory**. Nodes with less than 2GB allocatable are acceptable only for development environments.

This distinction matters because a Droplet's advertised RAM is not the same as Kubernetes' "allocatable" memory. A significant portion is reserved for the OS, kubelet, and system daemons. For example, a `s-2vcpu-4gb` node (4GB advertised) only provides **2.5 GiB allocatable** for pods. Capacity planning must be based on this allocatable number to avoid node-pressure evictions and OOMKilled pods.

**High Availability (HA) Node Pools**

For production workloads, node pools should have a **minimum of three nodes**. This, combined with deploying applications with `replicas: 3` and pod anti-affinity rules, ensures workloads can survive single-node failures or draining during upgrades.

Note: This is separate from the DOKS High Availability control plane, which is an optional paid feature providing redundant control plane nodes. The three-node baseline applies to *worker* node pools.

**DOKS Capacity Limits**

Hard limits to be aware of:
- **512 nodes maximum per cluster**
- **110 pods maximum per node** (theoretical; practical limit is usually CPU/memory)

The 110-pod limit is rarely the constraint—allocatable CPU and memory exhaust first.

**Instance Type Selection**

Common Droplet types for DOKS:
- **Basic (s-)**: Shared CPU, cost-effective for general workloads
- **General Purpose (g-)**: Dedicated CPU, balanced compute/memory
- **CPU-Optimized (c-)**: High CPU-to-memory ratio, compute-intensive workloads
- **Memory-Optimized (m-)**: High memory-to-CPU ratio, databases, caches
- **GPU (g-, gd-)**: GPU acceleration for ML, rendering, scientific compute

**Decision framework:**
- **Development**: `s-2vcpu-4gb` (minimum production-viable size)
- **Staging**: `s-4vcpu-8gb` (mirrors production architecture at smaller scale)
- **Production baseline**: `s-8vcpu-16gb` or `g-8vcpu-16gb` (3+ nodes for HA)
- **Specialized**: `c-` for CPU-bound apps, `m-` for memory-heavy services, `g-` for GPU workloads

### Mastering the DOKS Cluster Autoscaler

DOKS provides a **managed Cluster Autoscaler (CA)** that is enabled on a per-pool basis via `auto_scale=true` and `min_nodes`/`max_nodes` configuration.

**How the CA works in symbiosis with the Horizontal Pod Autoscaler (HPA):**

1. Load increases, and the HPA scales up a Deployment, creating new pod replicas
2. New pods enter "Pending" state because existing nodes are at capacity
3. The DOKS CA detects "Pending" pods and inspects their `resources.requests`
4. CA identifies a node pool that can satisfy those requests (considering size, labels, taints)
5. CA provisions a new node in that pool (up to `max_nodes`)
6. Pods are scheduled onto the new node
7. When load decreases and nodes are underutilized, CA drains and removes nodes (down to `min_nodes`)

**The #1 autoscaling anti-pattern: missing resource requests**

The most common reason the CA "fails" to scale is deploying pods **without `resources.requests`** set. Without this information, the CA has no idea what capacity the pod needs and **will not scale up the cluster**. Always specify requests:

```yaml
resources:
  requests:
    cpu: "500m"
    memory: "512Mi"
  limits:
    cpu: "1000m"
    memory: "1Gi"
```

**Advanced CA features:**

- **Flexible node pool selection**: CA can fall back to a different pool if the primary is at capacity
- **Expanders**: `least-waste` (minimize unused capacity) or `priority` (prefer specific pools)
- **Scale-to-zero**: GPU or specialized pools can autoscale down to 0 nodes for cost savings

**Autoscaling configuration examples:**

```hcl
# General app pool: always have baseline capacity
resource "digitalocean_kubernetes_node_pool" "apps" {
  auto_scale = true
  min_nodes  = 3   # HA baseline
  max_nodes  = 10  # Cost ceiling
}

# GPU pool: scale to zero when idle
resource "digitalocean_kubernetes_node_pool" "gpu" {
  auto_scale = true
  min_nodes  = 0   # No cost when unused
  max_nodes  = 2   # Limit expensive GPU nodes
}
```

### Zero-Downtime Lifecycle Management

#### Upgrades: Surge Rolling Updates

DOKS supports **surge upgrades** for zero-downtime version updates. This is a standard rolling update process:

1. One node at a time is cordoned (marked unschedulable)
2. Node is drained of its pods, respecting Pod Disruption Budgets (PDBs)
3. Node is terminated
4. New node on the target version is provisioned
5. Pods are rescheduled onto the new node
6. Process repeats for each node in the pool

**CRITICAL WARNING: Local Disk Data Loss**

This process **guarantees** that any data stored on the node's local disk (ephemeral storage, `emptyDir`, `hostPath`) will be lost. All stateful applications **must** use PersistentVolumes backed by DigitalOcean Block Storage to prevent data loss during upgrades.

**Best practices for safe upgrades:**
- Set appropriate Pod Disruption Budgets (PDBs) to maintain availability
- Use `replicas >= 3` with pod anti-affinity for critical services
- Test upgrades in staging environment first
- Monitor pod rescheduling and application health during rollout

#### Resizing: The Blue-Green Migration Pattern

It is **not possible** to change the Droplet size (e.g., from `s-2vcpu-4gb` to `s-4vcpu-8gb`) of an existing node pool in place.

The only way to "resize" a pool is a manual blue-green migration:

1. **Create a new node pool** (`pool-v2`) with the desired Droplet size
2. **Taint the old pool** (`pool-v1`) with `NoSchedule` to prevent new pods from landing there
3. **Drain each node** in the old pool: `kubectl drain <node> --ignore-daemonsets`
4. Kubernetes reschedules pods onto the new pool automatically
5. **Delete the old pool** once fully drained and verified

This orchestration is complex and error-prone when done manually—a prime candidate for automation by higher-level tooling.

#### Deletion: Clean Removal

Deleting a node pool via IaC (`terraform destroy`, `pulumi destroy`, `kubectl delete`) triggers:
1. All nodes are cordoned and drained
2. Pods are rescheduled to other pools (if capacity exists)
3. Nodes are terminated
4. Pool is removed from cluster configuration

**Pre-deletion checklist:**
- Ensure other pools have capacity to absorb rescheduled pods
- Check for workloads with `nodeSelector` or `nodeAffinity` targeting the pool
- Verify no critical services rely exclusively on this pool

### Cost Optimization Strategies

**DOKS' #1 Cost Advantage: Free Control Plane**

Unlike AWS EKS, Google GKE, and Azure AKS, which charge hourly fees for the control plane (typically $0.10/hour = $73/month), DOKS control planes are **completely free**. Billing is based **only** on the underlying resources: Droplets, Load Balancers, and Block Storage volumes.

This architectural choice makes DOKS compelling for development and staging environments, where control plane fees can exceed compute costs for small clusters.

**Primary cost optimization levers:**

1. **Cluster Autoscaler (CA)**: The #1 tool for dynamic cost management. It scales node pools to match demand, eliminating charges for idle compute. Especially effective for specialized pools (GPU, large Droplets) that can scale to zero when not in use.

2. **Right-sizing**: Use the smallest Droplet size that meets production requirements (≥ 2.5GB allocatable memory). Avoid overprovisioning. Use multiple specialized pools (e.g., `c-` for CPU-bound, `m-` for memory-bound) so the CA can scale them independently.

3. **Multi-pool architecture**: Separate workload types onto different pools with different autoscaling policies. Example:
   - System pool: Fixed 3 nodes (`s-2vcpu-4gb`) for cluster services
   - App pool: Autoscaling 3-10 nodes (`s-4vcpu-8gb`) for web services
   - Batch pool: Autoscaling 0-5 nodes (`c-8vcpu-16gb`) for background jobs

**Critical omission: No Spot Instances**

DOKS does **not** offer Spot/Preemptible instances (like AWS EKS Spot Instances or GKE Preemptible VMs). This is a fundamental difference from major cloud providers.

This absence **elevates the importance of the Cluster Autoscaler**. On DOKS, the CA is the *primary* dynamic cost-saving mechanism, making its proper configuration essential for an efficient cluster.

**Cost monitoring with tags:**

Use the `tags` parameter to apply DigitalOcean tags to all Droplets in a pool:

```hcl
resource "digitalocean_kubernetes_node_pool" "production" {
  tags = ["env:prod", "cost-center:engineering", "team:platform"]
}
```

The DigitalOcean billing dashboard allows filtering and grouping by tags, providing granular, per-pool cost attribution.

## The 80/20 API Surface: Essential vs Advanced Configuration

A cross-reference of `doctl` flags, Terraform arguments, and Pulumi inputs reveals a converged, standard API surface. This can be divided into "Core" (80%) and "Advanced" (20%) fields.

### The 80%: Core Fields Required by All Users

Every node pool, regardless of environment or workload, requires these fundamental parameters:

| Field | Type | Purpose | Example |
|-------|------|---------|---------|
| **cluster_id** | String (reference) | Links pool to its parent cluster | `digitalocean_kubernetes_cluster.main.id` |
| **name** | String | Human-readable identifier | `"production-workers"` |
| **size** | String (Droplet slug) | Specifies Droplet type/size | `"s-4vcpu-8gb"` |
| **Scaling config** | oneof | Fixed count OR autoscaling | See below |

**The scaling configuration oneof:**

This is the fundamental, mutually exclusive choice:

**Option 1: Fixed-size pool**
```hcl
node_count = 3
```

**Option 2: Autoscaling pool**
```hcl
auto_scale = true
min_nodes  = 3
max_nodes  = 10
```

Any API abstraction must enforce this mutual exclusivity. Setting both `node_count` and `auto_scale=true` is an error.

### The 20%: Advanced Fields for Production Isolation and Scheduling

These fields are not required for basic functionality, but are **essential for production-grade multi-pool architectures**:

| Field | Type | Purpose | Example |
|-------|------|---------|---------|
| **labels** | map&lt;string, string&gt; | Kubernetes labels for node attraction (nodeSelector, nodeAffinity) | `{"workload": "web", "env": "prod"}` |
| **taints** | array&lt;Taint&gt; | Kubernetes taints for workload repulsion | `[{key: "gpu", value: "true", effect: "NoSchedule"}]` |
| **tags** | array&lt;string&gt; | DigitalOcean tags for billing and filtering | `["env:prod", "cost-center:eng"]` |

**Critical distinction: labels vs taints vs tags**

A common point of confusion:

- **labels**: **Kubernetes labels** applied to Node objects. Used by `nodeSelector` and `nodeAffinity` to *attract* pods to specific nodes.
- **taints**: **Kubernetes taints** applied to Node objects. Used to *repel* pods that don't have matching tolerations.
- **tags**: **DigitalOcean tags** applied to Droplet resources. Used *outside* Kubernetes for filtering in the DigitalOcean console and for **cost attribution** on billing reports.

All three are first-class fields in the API and serve distinct, non-overlapping purposes.

### Common Configuration Recipes

**Dev/Test Pool: Minimal, Fixed-Size**
```hcl
resource "digitalocean_kubernetes_node_pool" "dev" {
  name       = "dev-sandbox"
  size       = "s-2vcpu-4gb"  # Minimum production-viable size
  node_count = 1              # Single node for cost savings
}
```

**Staging Pool: Autoscaling, General Purpose**
```hcl
resource "digitalocean_kubernetes_node_pool" "staging" {
  name       = "staging-apps"
  size       = "s-4vcpu-8gb"
  auto_scale = true
  min_nodes  = 1              # Cost-optimized minimum
  max_nodes  = 5              # Ceiling for staging loads
  tags       = ["env:staging", "cost-center:apps"]
}
```

**Production Pool: HA, Autoscaling, Isolated**
```hcl
resource "digitalocean_kubernetes_node_pool" "production" {
  name       = "prod-web-app"
  size       = "s-8vcpu-16gb"
  auto_scale = true
  min_nodes  = 3              # HA baseline
  max_nodes  = 10             # Cost ceiling
  
  labels = {
    workload = "web-app"
    env      = "prod"
  }
  
  tags = ["env:prod", "cost-center:web"]
}
```

**Specialized GPU Pool: Scale-to-Zero**
```hcl
resource "digitalocean_kubernetes_node_pool" "gpu" {
  name       = "gpu-jobs"
  size       = "g-4vcpu-16gb"
  auto_scale = true
  min_nodes  = 0              # No cost when idle
  max_nodes  = 2              # Limit expensive GPU nodes
  
  labels = {
    hardware = "gpu"
  }
  
  taint {
    key    = "nvidia.com/gpu"
    value  = "true"
    effect = "NoSchedule"      # Repel non-GPU workloads
  }
  
  tags = ["hardware:gpu", "cost-center:ml"]
}
```

## Advanced Multi-Pool Architectures

The "20%" fields—labels and taints—are the mechanisms that enable robust, production-grade multi-pool architectures.

### Workload Isolation: The Dedicated Node Pattern

DOKS automatically applies a label, `doks.digitalocean.com/node-pool`, to all nodes in a pool. For simple attraction, you can use this label with a `nodeSelector`:

```yaml
spec:
  nodeSelector:
    doks.digitalocean.com/node-pool: app-workers
```

However, this is **insufficient for true isolation**. It attracts pods to the pool, but doesn't *repel* other pods. A runaway cronjob or batch workload could still land on your expensive GPU nodes.

**The production pattern: Bi-directional isolation with taints + labels**

1. **Repel others (taints)**: Apply a taint to the specialized pool
2. **Allow intended workloads (tolerations)**: Give specific pods tolerations for the taint
3. **Attract intended workloads (labels)**: Use `nodeSelector` or `nodeAffinity` for the pool's labels

Example: Dedicated GPU pool

```hcl
# Node pool configuration
resource "digitalocean_kubernetes_node_pool" "gpu" {
  labels = { "hardware" = "gpu" }
  
  taint {
    key    = "nvidia.com/gpu"
    value  = "true"
    effect = "NoSchedule"  # Repels all pods by default
  }
}
```

```yaml
# GPU workload pod spec
spec:
  nodeSelector:
    hardware: gpu  # Attracts to GPU pool
  
  tolerations:
  - key: "nvidia.com/gpu"
    operator: "Equal"
    value: "true"
    effect: "NoSchedule"  # Allows scheduling on tainted nodes
```

This guarantees:
- Only GPU workloads land on expensive GPU nodes
- GPU nodes are not "polluted" by general-purpose workloads
- Cluster Autoscaler only scales GPU pool for GPU-requesting pods

### The Three-Tier Production Architecture

A robust production cluster uses a multi-pool architecture with clear separation of concerns:

#### Tier 1: System Pool (Fixed, Isolated)

**Purpose**: Run critical cluster services (CoreDNS, metrics-server, Ingress controllers)

```hcl
resource "digitalocean_kubernetes_node_pool" "system" {
  name       = "system-pool"
  size       = "s-2vcpu-4gb"
  node_count = 3              # Fixed HA, no autoscaling
  
  labels = {
    node-role = "system"
  }
  
  taint {
    key    = "node-role"
    value  = "system"
    effect = "NoSchedule"
  }
}
```

**Why this matters**: Protects cluster-vital services from resource starvation. A runaway application pod cannot consume all CPU/memory and kill CoreDNS, which would break the entire cluster's networking. This is **required** if you use specialized pools (GPU, large instances), as DOKS needs a stable CPU pool to run its managed addons.

**Implementation**: Patch all `kube-system` namespace DaemonSets and Deployments with the matching toleration.

#### Tier 2: General Application Pool (Autoscaling)

**Purpose**: Run the bulk of stateless and stateful application workloads

```hcl
resource "digitalocean_kubernetes_node_pool" "apps" {
  name       = "app-workers"
  size       = "s-4vcpu-8gb"
  auto_scale = true
  min_nodes  = 3
  max_nodes  = 20
  
  labels = {
    workload = "application"
  }
}
```

**Why this matters**: The "workhorse" pool. It leverages the Cluster Autoscaler to handle elastic load, scaling up for peak traffic and down for cost savings. Most services run here by default (no special node selectors needed).

#### Tier 3: Specialized Pools (GPU, CPU-Intensive, Memory-Intensive)

**Purpose**: Expensive, specialized Droplet types for specific workloads

```hcl
resource "digitalocean_kubernetes_node_pool" "gpu" {
  name       = "gpu-workers"
  size       = "g-4vcpu-16gb"
  auto_scale = true
  min_nodes  = 0              # Scale to zero when idle
  max_nodes  = 2
  
  labels = {
    hardware = "gpu"
  }
  
  taint {
    key    = "nvidia.com/gpu"
    value  = "true"
    effect = "NoSchedule"
  }
}
```

**Why this matters**: The *only* cost-effective way to use expensive hardware. The pool scales to zero when not in use, and taints ensure only workloads that explicitly request GPUs (and have the toleration) land there.

### Integrating Node-Level Scheduling with the Cluster Autoscaler

The labels and taints fields aren't just for manual scheduling—they're the primary mechanism for **controlling the Cluster Autoscaler** in a multi-pool environment.

The CA is fully "taint-aware" and "affinity-aware". When the CA sees a "Pending" pod, it inspects its full scheduling spec. It will **only** provision a new node from a pool that matches all three requirements:

1. The pod's `nodeSelector` or `nodeAffinity` (matches pool's labels)
2. The pod's `tolerations` (matches pool's taints)
3. The pod's `resources.requests` (can be satisfied by pool's Droplet size)

**Example: On-demand GPU job**

A user submits a GPU-requiring job to the cluster. The pod spec has:
- `nodeSelector: { hardware: gpu }`
- `tolerations` for `nvidia.com/gpu:NoSchedule`
- `resources.requests.nvidia.com/gpu: 1`

The pod goes "Pending". The CA inspects its spec, sees it matches the `gpu-workers` pool (which is currently at 0 nodes), provisions a new GPU node, schedules the pod, runs the job, and then (after an idle timeout) scales the pool back to 0.

This enables fully automated, on-demand infrastructure—no manual scaling required.

## DigitalOcean-Specific Considerations

### Droplet Types and Regional Availability

The `size` field is a slug string (e.g., `s-2vcpu-4gb`) that must match a valid Droplet type. Available types include:

- **Basic (s-)**: Shared CPU, cost-effective
- **General Purpose (g-)**: Dedicated CPU, balanced
- **CPU-Optimized (c-)**: High CPU-to-memory ratio
- **Memory-Optimized (m-)**: High memory-to-CPU ratio
- **GPU (g-, gd-)**: GPU acceleration

These types, and DOKS itself, vary in regional availability. Not all regions support all Droplet types or Kubernetes versions. Consult the DigitalOcean documentation for current availability.

### Kubernetes Version Compatibility

DOKS supports the **three most recent minor versions** of Kubernetes. All node pools within a cluster **must** run the same Kubernetes version as the control plane. This is not a per-pool setting.

When the cluster is upgraded, DOKS orchestrates a rolling upgrade of each node pool, one by one, using the surge upgrade process.

### VPC Integration and Private Networking

DOKS clusters are **VPC-native**. A DOKS cluster **must** be created within a DigitalOcean VPC. Any node pool provisioned for that cluster automatically exists within the same VPC.

This is a significant architectural benefit: all node-to-node and node-to-database traffic is internal and does not count against public bandwidth limits. Private networking is the default, not an opt-in feature.

### Cost Attribution with Tags

Billing is per-second, per-Droplet, with a free control plane. The `tags` field applies DigitalOcean tags to every Droplet created by the pool.

**Why tags matter for cost management:**

The `labels` field is for Kubernetes scheduling. The `tags` field is for DigitalOcean billing. If you set `tags: ["cost-center:prod-web"]`, your DigitalOcean bill can be filtered and grouped by this tag, providing granular, per-pool cost attribution that Kubernetes labels do not.

This reinforces the need to expose `labels`, `taints`, and `tags` as three distinct, first-class fields in any API abstraction.

## OpenMCF's Abstraction Layer

OpenMCF provides a **cloud-agnostic, protobuf-defined API** for node pool management. Instead of learning DigitalOcean's specific API patterns, Terraform resource syntax, or Pulumi's programming model, you declare intent using a consistent schema:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanKubernetesNodePool
metadata:
  name: production-workers
spec:
  cluster_ref: "prod-cluster-01"
  node_pool_name: "app-workers"
  size: "s-4vcpu-8gb"
  
  # Scaling configuration (mutually exclusive)
  auto_scale_config:
    min_nodes: 3
    max_nodes: 10
  
  # Advanced isolation
  labels:
    workload: "application"
    env: "production"
  
  taints:
  - key: "workload"
    value: "application"
    effect: "PreferNoSchedule"
  
  # Cost attribution
  tags:
  - "env:production"
  - "cost-center:engineering"
```

### The 80/20 Principle in Practice

OpenMCF's `DigitalOceanKubernetesNodePoolSpec` exposes only the essential fields identified in the 80/20 analysis:

**Core (80%):**
- `cluster_ref`: Links to parent cluster (lifecycle independence)
- `node_pool_name`: Human-readable identifier
- `size`: Droplet slug
- Scaling `oneof`: `fixed_node_count` OR `auto_scale_config` (mutual exclusivity enforced at protobuf level)

**Advanced (20%):**
- `labels`: Kubernetes scheduling (attraction)
- `taints`: Kubernetes scheduling (repulsion)
- `tags`: DigitalOcean billing and filtering

**Notably absent (intentionally):**
- Network configuration (inherited from cluster's VPC)
- Kubernetes version (synchronized with cluster control plane)
- Region (inherited from cluster)
- Public IP settings (sensible defaults)

This simplification means:
- **Faster onboarding**: New users aren't overwhelmed by 50+ optional parameters
- **Less error-prone**: Fewer knobs to misconfigure
- **Consistent patterns**: Same API structure across cloud providers

### Lifecycle Independence by Design

Unlike the Terraform provider's default model (inline default pool in cluster resource), OpenMCF treats **every node pool as a first-class, standalone resource**.

The `cluster_ref` field creates a loose reference, not a nested ownership. This mirrors the "sacrificial default pool" production pattern: all pools are lifecycle-independent from the cluster.

Benefits:
- Modify, resize, or delete pools without touching the cluster
- No risk of accidental cluster replacement
- Clean separation of concerns: cluster provisioning vs workload capacity management

### Multi-Cloud Consistency

The same protobuf schema pattern works across providers. Compare:
- `DigitalOceanKubernetesNodePool`
- `GcpGkeNodePool`
- `AwsEksNodeGroup`
- `AzureAksNodePool`

The concepts map cleanly: size, scaling, labels, taints. A team managing infrastructure across clouds learns one API instead of four vendor-specific ones.

### What OpenMCF Isn't

It's not a replacement for Terraform, Pulumi, or Crossplane—it's a **higher-level orchestrator**. Under the hood, OpenMCF likely generates IaC code or calls cloud APIs directly.

The value is:
- **Opinionated defaults**: Production-ready configurations without deep cloud expertise
- **Cross-cloud portability**: Swap DigitalOcean for GCP by changing the API version
- **Policy enforcement**: Centralized control over allowed configurations
- **Simplified surface area**: 80/20 principle reduces cognitive load

For teams building platforms or managing multi-cloud infrastructure, this abstraction reduces operational complexity. For users needing full control of every DigitalOcean API parameter, Terraform/Pulumi remain better choices.

## Production Anti-Patterns to Avoid

Based on community discussions, GitHub issues, and DigitalOcean documentation, these are the most common pitfalls:

**1. The "One Giant Pool" Anti-Pattern**

Using a single node pool for all workloads (system services, apps, batch jobs, databases).

**Why it's bad:**
- Breaks workload isolation
- Cripples the autoscaler (a test pod can force an expensive scale-up)
- Makes upgrades all-or-nothing, high-risk events
- No cost optimization (can't scale different workload types independently)

**Fix:** Use multi-pool architecture (system, apps, specialized).

**2. The "Undersized Production" Anti-Pattern**

Using Droplets with &lt; 2.5GB allocatable memory for production workloads.

**Why it's bad:**
- Leads to node-pressure evictions
- OOMKilled pods and instability
- Poor resource utilization (too much overhead, not enough capacity)

**Fix:** Use `s-2vcpu-4gb` as the absolute minimum for production (provides 2.5GB allocatable).

**3. The "No-Request" Anti-Pattern**

Deploying pods **without** `resources.requests` defined.

**Why it's bad:**
- Breaks the Cluster Autoscaler (CA has no idea what capacity to provision)
- Blinds the Kubernetes scheduler (poor bin-packing decisions)
- Leads to "Pending" pods that never get scheduled

**Fix:** Always specify `resources.requests` for CPU and memory in all pod specs.

**4. The "Local Disk State" Anti-Pattern**

Using `emptyDir` or `hostPath` for stateful data (databases, file uploads, etc.).

**Why it's bad:**
- **Guaranteed 100% data loss** during any node upgrade, recycling, or failure
- Node upgrades drain nodes and terminate them, destroying local disk

**Fix:** Use PersistentVolumes backed by DigitalOcean Block Storage for all stateful data.

**5. The "Inline Default Pool Modification" Anti-Pattern**

Modifying the inline `node_pool` block in the `digitalocean_kubernetes_cluster` resource after initial creation.

**Why it's bad:**
- Triggers cluster replacement (destroy + recreate entire cluster)
- Catastrophic downtime and data loss

**Fix:** Use the "sacrificial default pool" pattern—create minimal default pool and never touch it again. Manage all real workloads via separate `digitalocean_kubernetes_node_pool` resources.

**6. The "No Pod Disruption Budget" Anti-Pattern**

Running stateful or critical services without Pod Disruption Budgets (PDBs).

**Why it's bad:**
- Node draining during upgrades can violate SLAs
- All replicas could be drained simultaneously, causing downtime

**Fix:** Define PDBs for all critical services:

```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: my-app-pdb
spec:
  minAvailable: 2
  selector:
    matchLabels:
      app: my-app
```

## Conclusion

The journey from manual UI operations to production-grade node pool management reflects the broader evolution of infrastructure: from imperative commands to declarative desired state, from ad-hoc changes to version-controlled automation, from monolithic clusters to specialized, isolated pools.

For DigitalOcean Kubernetes node pools, the production baseline is clear: **use Infrastructure-as-Code** (Terraform, Pulumi, or Crossplane), adopt the **sacrificial default pool pattern**, and embrace a **multi-pool architecture** that leverages labels, taints, and the Cluster Autoscaler for robust workload isolation and cost optimization.

The "default node pool problem" teaches a valuable lesson: infrastructure resources should be lifecycle-independent, modular, and replaceable. OpenMCF's abstraction embodies this principle by treating every node pool as a first-class resource with a loose cluster reference, not a nested child.

The 80/20 API surface—cluster reference, name, size, scaling config, plus optional labels/taints/tags—covers the vast majority of use cases while keeping the cognitive load manageable. The remaining 20% of edge cases (custom networking, specialized instance configurations) can be handled through provider-specific extensions or manual overrides.

DigitalOcean's free control plane and VPC-native architecture create a compelling foundation. When combined with the Cluster Autoscaler (the primary cost optimization tool in the absence of Spot Instances), multi-pool isolation patterns, and proper capacity planning (≥ 2.5GB allocatable memory), DOKS becomes a production-ready platform for Kubernetes workloads.

The future of node pool management isn't more parameters and complexity—it's **smarter defaults, stronger abstractions, and automation that fades into the background**. OpenMCF's cloud-agnostic API, the maturity of Terraform/Pulumi providers, and the Kubernetes-native model of Crossplane all push in that direction.

Whether you choose Terraform's widespread ecosystem, Pulumi's unified programming model, or Crossplane's continuous reconciliation, the principles remain constant: declare your intent, version it in Git, let tooling reconcile reality, and design for lifecycle independence from day one.


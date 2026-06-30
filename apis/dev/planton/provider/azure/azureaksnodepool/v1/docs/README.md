# Azure AKS Node Pool Deployment: From Manual Clicks to Production-Ready IaC

## Introduction

In the early days of Azure Kubernetes Service (AKS), the conventional wisdom was simple: spin up a cluster, use the default node pool, and call it done. If you needed more capacity, just scale up the node count. A single pool of identical VMs seemed like the path of least resistance.

But that simplicity came at a cost. Teams discovered that mixing critical system components with unpredictable application workloads on the same nodes led to resource contention and reliability issues. A misbehaving application pod could starve CoreDNS or metrics-server, destabilizing the entire cluster. The need for specialized hardware—GPU nodes for machine learning, high-memory nodes for caching, Windows nodes for legacy .NET apps—couldn't be met with one-size-fits-all infrastructure. And cost optimization strategies like Spot instances were off the table when your cluster control plane components lived on the same nodes as your batch jobs.

The paradigm shifted when Azure introduced **node pools**—independent groups of worker nodes within a single cluster, each with its own VM size, operating system, scaling policy, and availability configuration. Suddenly, you could isolate system workloads from user applications, dedicate GPU nodes to AI inference, run Windows containers alongside Linux microservices, and leverage cheap Spot instances for fault-tolerant batch processing—all within one cohesive cluster.

But with this flexibility came complexity. How do you deploy and manage multiple node pools across environments? What's the right balance between isolation and operational overhead? Which infrastructure-as-code tool handles the lifecycle of these pools most effectively?

This document explores the landscape of AKS node pool deployment methods, from anti-patterns to avoid to production-ready approaches that have been battle-tested at scale. We'll examine why Planton chose its particular abstraction and how it distills production best practices into a streamlined API that teams actually want to use.

## The Maturity Spectrum: How Teams Deploy Node Pools

### Level 0: The Azure Portal (Manual Configuration)

The Azure Portal offers a friendly graphical interface for adding node pools. You click through a wizard, select a VM size, set the node count, maybe check a box for autoscaling, and hit "Create." For learning AKS or experimenting with a proof-of-concept, this is perfectly fine.

**The Problem**: None of this is captured as code. You can't version control a series of mouse clicks. Reproducing the configuration in another region or subscription requires manually clicking through the same steps (and hoping you remember all the settings you chose). If you later adopt infrastructure-as-code, those manually-created node pools become orphans—not tracked by your IaC tool, easily deleted by accident when someone runs a deployment that doesn't know they exist.

**Common Mistakes**:
- Forgetting to enable availability zones, leaving your cluster vulnerable to zonal failures
- Not setting autoscaling limits, leading to surprise Azure bills or capacity shortages
- Mixing Spot VMs into your system node pool (AKS won't even allow this, but the portal's error messages aren't always clear)
- Inconsistent configurations across environments because "I think I set it to 3 nodes last time, but I'm not sure"

**Verdict**: The portal is fine for exploration and one-off clusters. For production infrastructure that needs to be repeatable, auditable, and safe from human error, it's a non-starter.

### Level 1: Azure CLI and PowerShell (Scripted Imperative Management)

The Azure CLI (`az aks nodepool add/delete/scale/upgrade`) and Azure PowerShell (`New-AzAksNodePool`, `Update-AzAksNodePool`) bring automation to node pool management. You can script the creation of multiple pools, store those scripts in version control, and run them in CI/CD pipelines.

**What This Solves**: Repeatability and consistency. The same script produces the same configuration every time. You can code-review changes to node pool settings. Integration with existing Azure CLI tooling is seamless.

**What It Doesn't Solve**:
- **Idempotency**: Running `az aks nodepool add` twice for the same pool name will fail the second time. Your scripts need defensive checks ("does this pool already exist?") to avoid errors.
- **State Management**: CLI commands are imperative—they tell Azure what action to take, not what the desired end state is. If someone manually changes a node pool setting, your script won't detect or correct the drift unless you explicitly query and reconcile.
- **Dependency Orchestration**: You must ensure the parent AKS cluster exists before adding pools. Order matters. Error handling for partial failures becomes your responsibility.

**Production Reality**: Teams using CLI automation often wrap commands in sophisticated shell scripts or Python programs that approximate declarative behavior—essentially building their own state management layer. This works, but it's reinventing infrastructure-as-code concepts that mature tools already provide.

**Verdict**: CLI scripting is a step up from the portal and suitable for smaller deployments or teams deeply invested in Azure CLI tooling. For complex multi-pool architectures, declarative IaC is a better fit.

### Level 2: Azure-Native IaC (ARM Templates and Bicep)

Azure Resource Manager (ARM) templates and Bicep (Microsoft's domain-specific language that compiles to ARM JSON) bring declarative infrastructure to Azure. You define the desired state of your node pools in a template, and Azure ensures the deployed resources match.

**ARM/Bicep Example**:
```bicep
resource aksCluster 'Microsoft.ContainerService/managedClusters@2023-09-01' existing = {
  name: 'my-aks-cluster'
}

resource userPool 'Microsoft.ContainerService/managedClusters/agentPools@2023-09-01' = {
  parent: aksCluster
  name: 'apppool'
  properties: {
    vmSize: 'Standard_D8s_v3'
    count: 2
    enableAutoScaling: true
    minCount: 2
    maxCount: 10
    mode: 'User'
    osType: 'Linux'
    availabilityZones: ['1', '2', '3']
    orchestratorVersion: aksCluster.properties.kubernetesVersion
  }
}
```

**Strengths**:
- **Native Azure Integration**: ARM/Bicep deployments integrate seamlessly with Azure DevOps, Azure Pipelines, and Azure's RBAC model
- **Declarative**: You describe the desired state; Azure figures out what needs to change
- **Official Support**: Microsoft maintains these formats, so new AKS features appear here quickly

**Challenges**:
- **Deployment Modes**: ARM templates can deploy in "incremental" or "complete" mode. Incremental is safer (only adds/updates resources in the template), but complete mode will delete resources not in the template—risky if you're not careful
- **State Visibility**: Unlike Terraform, ARM doesn't maintain a separate state file you can inspect. State lives in Azure's resource manager. Detecting drift requires querying Azure and comparing to your template
- **Multi-Cloud Limitations**: If you're managing infrastructure across Azure, AWS, and GCP, you'll need different tools for each. ARM/Bicep is Azure-only

**Production Adoption**: ARM templates are widely used in enterprise Azure shops, especially where governance mandates Azure-native tooling. Bicep's improved syntax over raw JSON makes it more palatable for developers.

**Verdict**: Solid choice for Azure-only deployments where deep Azure DevOps integration is valuable. The learning curve for Bicep is gentler than Terraform for teams new to IaC.

### Level 3: Multi-Cloud IaC (Terraform and Pulumi)

Terraform (with the AzureRM provider) and Pulumi (with Azure Native or Azure Classic providers) represent the multi-cloud IaC approach. Both are declarative, maintain state, and support infrastructure across cloud providers.

**Terraform Example**:
```hcl
resource "azurerm_kubernetes_cluster_node_pool" "app_pool" {
  name                  = "apppool"
  kubernetes_cluster_id = azurerm_kubernetes_cluster.main.id
  vm_size               = "Standard_D8s_v3"
  node_count            = 2
  enable_auto_scaling   = true
  min_count             = 2
  max_count             = 10
  zones                 = ["1", "2", "3"]
  mode                  = "User"
  os_type               = "Linux"
  orchestrator_version  = azurerm_kubernetes_cluster.main.kubernetes_version
}
```

**Pulumi Example (TypeScript)**:
```typescript
const appPool = new azure.containerservice.KubernetesClusterNodePool("apppool", {
    kubernetesClusterId: cluster.id,
    vmSize: "Standard_D8s_v3",
    nodeCount: 2,
    enableAutoScaling: true,
    minCount: 2,
    maxCount: 10,
    availabilityZones: ["1", "2", "3"],
    mode: "User",
    osType: "Linux",
});
```

**Why This Works in Production**:

1. **Plan-Before-Apply Workflow**: Both Terraform and Pulumi show you exactly what will change before making changes. This prevents accidental deletions or unexpected modifications
2. **State Management**: Explicit state tracking means drift detection works out of the box. If someone manually scales a node pool, your next `terraform plan` or `pulumi preview` will show the difference
3. **Lifecycle Management**: Creating, updating, and destroying node pools is handled gracefully. Want to change VM size? Terraform will show it requires replacement (destroy old pool, create new one), letting you plan a blue/green migration
4. **Module Ecosystem**: Both have rich module/component ecosystems for reusable patterns (e.g., "standard multi-pool AKS cluster" modules)
5. **Multi-Cloud Portability**: The same tool that manages your Azure node pools can manage your AWS EKS node groups and GCP GKE node pools

**Production Patterns**:
- **Separate State Per Environment**: Dev, staging, and prod clusters have separate state files, preventing cross-contamination
- **Remote State Storage**: State stored in Azure Blob Storage (Terraform) or Pulumi's service, with locking to prevent concurrent modifications
- **Autoscaling Awareness**: Configure Terraform/Pulumi to ignore the `node_count` field when autoscaling is enabled (since the cluster autoscaler will change it dynamically), focusing only on `min_count` and `max_count`
- **Blue/Green Node Pool Upgrades**: Create a new node pool with the target configuration, migrate workloads, then delete the old pool—all orchestrated through IaC

**Common Pitfalls** (and how to avoid them):
- **Accidental Deletions**: If you remove a node pool from your Terraform/Pulumi code, the next apply will delete it. Solution: Use lifecycle policies (`prevent_destroy` in Terraform) for critical pools, or require explicit deletion via targeted destroys
- **State Drift**: If someone makes out-of-band changes via the portal or CLI, your state becomes stale. Solution: Regular `terraform refresh` or `pulumi refresh`, and enforce a culture of "all changes through IaC"
- **Immutable Fields**: You can't change VM size, OS type, or certain other fields on an existing node pool. Terraform/Pulumi will try to replace the resource. Solution: Plan blue/green migrations for such changes

**Terraform vs. Pulumi**:
- **Terraform**: Mature ecosystem, widespread adoption, HCL is domain-specific and concise. State management is explicit but requires operational discipline
- **Pulumi**: Real programming languages (TypeScript, Python, Go, C#) enable loops, conditionals, functions—powerful for complex logic. State management is similar but often uses Pulumi's managed backend. Slightly steeper learning curve for non-programmers

**Verdict**: Both are production-ready. Choose Terraform if you value a massive community, extensive third-party modules, and a proven track record. Choose Pulumi if you want the expressiveness of a full programming language and appreciate modern developer tooling (like IDE autocomplete for cloud resources).

### Level 4: GitOps and Higher-Level Abstractions (Crossplane, Argo CD, Planton)

At the highest level of maturity, teams treat infrastructure as declarative Kubernetes-style resources, managed through GitOps workflows.

**Crossplane**: An open-source project that runs in Kubernetes and provisions cloud resources via Custom Resource Definitions (CRDs). You define an AKS node pool as a YAML manifest, apply it with `kubectl`, and Crossplane's controllers call Azure APIs to create it.

**Strengths**: 
- Infrastructure becomes Kubernetes-native—manage clusters and node pools the same way you manage deployments and services
- GitOps-friendly: use Argo CD or Flux to sync infrastructure from Git
- Multi-cloud composition: Crossplane can orchestrate resources across clouds with a unified API

**Challenges**:
- Crossplane's Azure provider historically lagged Azure's native features (e.g., multiple node pools weren't initially supported)
- Adds operational complexity (you're running controllers in Kubernetes to manage Kubernetes infrastructure—a meta-problem)
- Debugging can be opaque (errors surface as Kubernetes events rather than direct API feedback)

**Planton's Approach**: Sits atop Pulumi, providing a curated, opinionated API that distills production best practices into a simple resource definition. Instead of exposing every knob and lever (50+ fields in the raw AKS API), Planton focuses on the 20% of configuration that 80% of teams actually need:

- **Cluster association**: Which AKS cluster does this pool belong to?
- **VM size**: What instance type?
- **Scaling policy**: Fixed node count or autoscaling with min/max?
- **Availability zones**: High availability spread?
- **OS type**: Linux or Windows?
- **Spot instances**: Cost savings vs. reliability trade-off?

The underlying Pulumi module handles the boilerplate—subnet assignment, proper taints for system pools, orchestrator version alignment, upgrade surge settings. You get the power of Pulumi with guardrails that prevent common mistakes.

**Verdict**: For teams seeking simplicity without sacrificing production-readiness, abstractions like Planton deliver the best developer experience. For those needing fine-grained control of every Azure feature, Terraform/Pulumi's raw providers are better. Crossplane shines in Kubernetes-centric, multi-cloud environments where treating infrastructure as Kubernetes resources aligns with operational culture.

## IaC Tool Comparison: The Details

Here's a structured comparison of how Terraform, Pulumi, and ARM/Bicep handle multi-node-pool AKS clusters in production:

| Aspect | **Terraform (AzureRM)** | **Pulumi (Azure Native)** | **Bicep/ARM Templates** |
|--------|------------------------|---------------------------|-------------------------|
| **Defining Multiple Pools** | Separate `azurerm_kubernetes_cluster_node_pool` resources, each linked to the cluster ID. The default pool is defined in the cluster resource itself. | Similar: separate `KubernetesClusterNodePool` resources. Clean separation between cluster and additional pools. | Node pools are `Microsoft.ContainerService/managedClusters/agentPools` resources. Can be defined inline or as separate resources with `parent` reference. |
| **Creating Pools** | `terraform apply` creates new pools. Terraform handles dependencies automatically (pool waits for cluster). Import existing pools with `terraform import`. | `pulumi up` creates pools with preview of changes. Dependency graph ensures cluster exists first. | Deploy template via `az deployment group create`. Incremental mode adds/updates resources; complete mode replaces all (use with caution). |
| **Updating & Scaling** | Changing `min_count`/`max_count` triggers in-place update. Changing `vm_size` or `os_type` forces recreation (delete + create). Terraform plan shows which. Set lifecycle `ignore_changes` for `node_count` if autoscaling. | Same behavior: immutable fields trigger replacement. Pulumi preview clearly shows create-before-delete or delete-before-create. | ARM deployment applies changes declaratively. Changing immutable fields requires deleting the old pool resource or deploying a new one. |
| **Kubernetes Version Sync** | Must explicitly set `orchestrator_version` for each pool. No automatic sync—you manage alignment in code. Best practice: use a variable for cluster version and reference it. | Same manual alignment. Pulumi's programming model makes it easy to loop through pools and set versions programmatically. | If `orchestratorVersion` is omitted, pool stays at current version. Explicitly set to keep in sync with control plane. |
| **Autoscaling** | Enable with `enable_auto_scaling = true`, then set `min_count`/`max_count`. Cluster autoscaler (AKS component) handles actual scaling. Terraform just configures bounds. | Identical: set autoscaling properties. Pulumi doesn't scale nodes itself—AKS does. | Same: `enableAutoScaling: true` with min/max in template. Autoscaling is runtime behavior, not template-time. |
| **Deletion** | Removing resource from config and applying deletes the pool. Use `prevent_destroy` lifecycle to block accidental deletions. Pod Disruption Budgets can delay deletion (AKS drains nodes). | Removing resource and running `pulumi up` deletes. Pulumi's "protected" resources can prevent deletion. | Remove resource from template and deploy. Or use `az aks nodepool delete`. Complete mode deployment can delete resources not in template. |
| **State & Drift** | Terraform state tracks desired vs. actual. Run `terraform refresh` to sync. Detects out-of-band changes on next plan. | Pulumi state is similar. `pulumi refresh` syncs state with reality. Strong drift detection. | No separate state file. ARM compares template to current Azure resources each deployment. Drift only corrected if you redeploy. |
| **Upgrade Strategies** | Update `orchestrator_version` in code and apply. Terraform calls Azure API to trigger rolling upgrade. Can set `max_surge` to control speed/risk. | Same: update version and apply. Pulumi triggers AKS upgrade API. | Update version in template and deploy. Azure handles rolling upgrade with surge/unavailable settings. |

**Summary**: Terraform and Pulumi offer robust planning, explicit state management, and strong drift detection—ideal for complex multi-pool setups. ARM/Bicep is Azure-native, integrates deeply with Azure tooling, but requires more discipline around deployment modes and drift. All three can accomplish production-grade node pool management; the choice often comes down to existing team expertise and multi-cloud strategy.

## Production Essentials: What Actually Matters

The Azure AKS API exposes 50+ fields for configuring node pools. Most are niche. Here's what production teams actually configure, based on real-world patterns:

### The Core 20% (Essential for 80% of Use Cases)

**1. VM Size (SKU)**  
Determines CPU, memory, and network capacity. Common choices:
- **Standard_D4s_v3** (4 vCPU, 16 GB): System pools, small user pools
- **Standard_D8s_v3** (8 vCPU, 32 GB): General-purpose user pools
- **Standard_E4s_v3** (4 vCPU, 32 GB): Memory-intensive workloads (caching, databases)
- **Standard_NC6/NC12** (GPU instances): AI/ML inference and training
- **Standard_B2s/B4ms** (Burstable): Dev/test only—never for production system pools

**Gotcha**: AKS requires system pools to use VMs with ≥2 vCPU and ≥4 GB memory. Some tiny SKUs (A-series, B-series) are disallowed.

**2. Node Count and Autoscaling**  
- **Fixed Count**: Simple and predictable. Use for system pools (e.g., 3 nodes for HA) or pools with stable load.
- **Autoscaling**: Specify `min_count` and `max_count`. Cluster autoscaler adds nodes when pods are pending, removes nodes when underutilized. Enable for user pools with variable load.
  - System pool: Often fixed at 2-3 nodes (no autoscaling) for stability
  - User pool: Autoscale from 2 (baseline) to 10+ (peak load)
  - Spot pool: Can scale to 0 when idle, saving costs

**3. Availability Zones**  
Multi-zone spread (`["1", "2", "3"]`) protects against datacenter failures. If zone 1 goes down, pods reschedule to nodes in zones 2 and 3. 

**Best Practice**: Always use zones for production pools. Non-zonal deployments are single points of failure.

**Gotcha**: You can't add zones to an existing pool. If you forgot to enable zones, create a new pool and migrate workloads.

**4. OS Type**  
- **Linux**: Default, runs 90%+ of Kubernetes workloads. System pools must be Linux.
- **Windows**: Only for Windows containers (legacy .NET, certain enterprise apps). Requires Azure CNI networking. Higher resource overhead.

**5. Spot Instances (Priority)**  
Spot VMs cost up to 90% less but can be evicted with 30 seconds' notice. 

**Use Cases**:
- Batch processing, CI/CD jobs, dev/test workloads
- Overflow capacity (run critical replicas on regular nodes, scale extra replicas to Spot)

**Do NOT Use Spot For**:
- System node pools (AKS forbids this)
- Single-instance critical services (unless you enjoy outages)
- Workloads that can't tolerate abrupt restarts

**6. Mode (System vs. User)**  
- **System**: Hosts critical cluster components (CoreDNS, metrics-server, kube-proxy). Must be Linux. Every cluster needs at least one.
- **User**: Runs application workloads. Can be Linux or Windows. Can scale to zero (system pools cannot).

**Best Practice**: Isolate system pods on a dedicated system pool with a taint (`CriticalAddonsOnly=true:NoSchedule`). Run all apps on user pools.

### The Other 80% (Rarely Touched)

- **Node Labels/Taints**: Custom scheduling controls. Most teams use defaults or set these via Kubernetes (not at pool creation).
- **Max Pods per Node**: Defaults are usually fine (110 for Azure CNI). Change only for specific networking constraints.
- **Proximity Placement Groups**: Co-locate nodes for ultra-low latency (HPC, trading apps). 95% of workloads don't need this.
- **Ephemeral OS Disks**: Faster boot, lower cost. Becoming more common but still opt-in.
- **FIPS Mode, Secure Boot, TPM**: Compliance and confidential computing. Enterprise-specific.

### Example Configurations

**System Pool (Production)**:
```yaml
name: systempool
mode: System
vmSize: Standard_D4s_v3
initial_node_count: 3
availability_zones: ["1", "2", "3"]
os_type: Linux
autoscaling: null  # Fixed size
spot_enabled: false
```

**General User Pool (Production)**:
```yaml
name: apppool
mode: User
vmSize: Standard_D8s_v3
initial_node_count: 2
availability_zones: ["1", "2", "3"]
os_type: Linux
autoscaling:
  min_nodes: 2
  max_nodes: 10
spot_enabled: false
```

**Spot Pool (Cost-Optimized)**:
```yaml
name: spotpool
mode: User
vmSize: Standard_D4as_v5
initial_node_count: 0
availability_zones: ["1", "2", "3"]
os_type: Linux
autoscaling:
  min_nodes: 0
  max_nodes: 5
spot_enabled: true
```

## Upgrade and Maintenance: Keeping It Running

Node pools aren't set-and-forget. Kubernetes versions evolve, security patches land, and VM sizes change. Here's how production teams handle it.

### Kubernetes Version Upgrades

**The Rules**:
- Control plane must upgrade first
- Node pools can lag one minor version behind control plane (e.g., control plane on 1.27, pools on 1.26)
- Best practice: Keep everything on the same version to avoid subtle incompatibilities

**Strategies**:
1. **In-Place Rolling Upgrade**: AKS cordons a node, drains pods, upgrades the node OS/K8s version, uncordons. Repeat for each node. Set `max_surge` to control how many nodes upgrade concurrently (default: 1 node at a time).
2. **Blue/Green Pool Replacement**: Create a new pool with the target version, migrate workloads, delete old pool. Safest for major version jumps or when changing VM sizes.

**Automation**:
- **Auto-Upgrade Channels**: AKS can auto-upgrade control plane and pools (rapid, stable, patch channels). Hands-off but less control.
- **Manual via IaC**: Update `orchestrator_version` in Terraform/Pulumi, preview changes, apply. Gives full control and auditability.

### Node Image Patching

Microsoft releases updated node images (OS patches, security fixes) regularly. You can:
- **Auto-Update**: Enable node image auto-upgrade. AKS will roll new images weekly/bi-weekly.
- **Manual**: Run `az aks nodepool upgrade --node-image-only` on a schedule.

**Impact**: Brief pod evictions as nodes restart. Use Pod Disruption Budgets to limit concurrent disruptions.

### Handling Immutable Changes

Can't change: VM size, OS type, disk type.  
**Solution**: Create new pool, migrate, delete old pool. IaC makes this repeatable.

### Common Pitfalls

- **PDB Blocking Upgrades**: If a Pod Disruption Budget sets `maxUnavailable: 0` and you only have one replica, the node can't drain. AKS will retry for hours. Fix: Add more replicas or temporarily relax the PDB.
- **Forgetting to Upgrade All Pools**: You upgrade the control plane to 1.28 but leave a pool on 1.26. Works, but you're now managing version skew. Update your IaC to bump all pools together.

## What Planton Supports (and Why)

Planton's `AzureAksNodePool` resource distills the research and production patterns above into a focused API. Here's what we chose to include:

**Included** (The Essential 20%):
- `cluster_name`: Reference to the parent AKS cluster
- `vm_size`: Instance type (required—no sane default exists)
- `initial_node_count`: Starting node count (required, validates > 0)
- `autoscaling`: Optional; when set, enables cluster autoscaler with min/max bounds
- `availability_zones`: Optional list of zones (validates ≥2 for HA)
- `os_type`: Enum (Linux or Windows), defaults to Linux
- `spot_enabled`: Boolean, defaults to false

**Excluded** (Advanced/Niche):
- Custom node labels and taints (set via Kubernetes, not IaC)
- Max pods per node (sensible defaults)
- Proximity placement groups (HPC-specific)
- Kubelet config, sysctls (edge cases)

**Why This Abstraction Works**:

1. **Guardrails**: Validations prevent common mistakes (e.g., can't create a 0-node pool, zones require ≥2 for HA)
2. **Defaults Align with Best Practices**: Linux OS, regular (non-Spot) VMs unless you opt in
3. **Pulumi-Powered**: Underlying module handles boilerplate (subnet assignment, orchestrator version sync, surge settings)
4. **Escape Hatches Exist**: If you need to set a niche field, the Pulumi module can be extended or you can use raw Pulumi/Terraform for that specific pool

**The Philosophy**: 80% of teams need straightforward, reliable node pools with HA, autoscaling, and cost controls. The API should make that trivial. The 20% with specialized needs (FIPS mode, custom kubelets) can use lower-level tools or contribute enhancements to the Pulumi module.

**Integration with AKS Clusters**: `AzureAksNodePool` references an `AzureAksCluster` by name. The Planton orchestrator ensures the cluster exists before creating pools, handles cleanup on deletion, and can manage multiple pools declaratively from a single spec.

## Conclusion: The Paradigm Shift

The evolution from "one default node pool" to "architected multi-pool clusters" mirrors Kubernetes' own maturity. Early on, simplicity was the goal. As systems scaled and workloads diversified, that simplicity became a liability.

Today, production AKS clusters routinely run:
- A small, stable **system pool** (3 nodes, multi-zone, fixed size) for cluster components
- A **general user pool** (autoscaling 2-10 nodes, multi-zone) for stateless apps
- A **Spot pool** (autoscaling 0-5 nodes) for batch jobs and dev environments
- Specialized pools for **GPU workloads**, **Windows containers**, or **memory-intensive services**

Managing this complexity manually through the portal is untenable. Azure CLI scripts approximate repeatability but lack state management. ARM/Bicep provides Azure-native declarative IaC. Terraform and Pulumi bring battle-tested state management and multi-cloud portability.

Planton sits atop Pulumi, curating the 20% of configuration that matters most and encoding production best practices (zones, autoscaling, Spot isolation) into a simple, validated resource definition. You get the power of IaC without drowning in boilerplate.

The strategic choice is clear: **treat node pools as code**, manage them declaratively, and use tools that catch mistakes before they reach production. Whether you choose Terraform, Pulumi, ARM, or a higher-level abstraction like Planton, the anti-pattern is the same: manual, undocumented, one-off changes that can't be reviewed, tested, or rolled back.

The clusters you deploy today will run for years. Make sure the infrastructure that powers them is as robust, version-controlled, and well-architected as the applications running on top.


# Deploying Azure AKS: From Console Clicks to Production Kubernetes Infrastructure

## Introduction

For years, the conventional wisdom around managed Kubernetes was simple: "just get a cluster running." Spin up three nodes, point kubectl at it, and you're done. The reality? That's how you get a Kubernetes cluster. It's not how you get a *production* Kubernetes cluster.

The good news: the industry has converged on what "production-ready" actually means for Azure Kubernetes Service (AKS). This isn't subjective anymore. Microsoft has codified it in their baseline architectures, and the once-confusing landscape of networking plugins, identity models, and node pool configurations has crystallized into clear patterns.

The bad news: getting there still requires navigating a maze of decisions. Should you use Kubenet or Azure CNI? What's the difference between Azure CNI and Azure CNI Overlay? Do you need a User-Assigned Managed Identity or a Service Principal? Is the Free tier actually free, or just free until production breaks?

This document maps the AKS deployment landscape from anti-patterns to production-grade practices. We'll explore how the definition of a production cluster has evolved, why certain once-common patterns are now explicitly deprecated, and how Planton abstracts the complexity while preserving the power.

## The Production Baseline: What Changed (And What's Non-Negotiable Now)

Before diving into deployment methods, let's establish what we mean by "production-ready" in 2025. This isn't opinion—it's based on Microsoft's own baseline architectures and hard lessons from hundreds of cluster deployments.

### The Non-Negotiables

**1. Control Plane SKU: Standard Tier is Mandatory**

The Free tier exists for one purpose: learning and development. It has *no financially-backed uptime SLA*. The Standard tier costs approximately $0.10 per cluster per hour (~$73/month) and provides a 99.95% availability SLA for clusters using Availability Zones, or 99.9% without them.

This is not an optimization to make later. If you start on the Free tier, you're building technical debt into your infrastructure before writing a single line of application code. Production clusters begin with the Standard tier.

**2. Networking: Kubenet is Deprecated (And You Should Care)**

On March 31, 2028, Kubenet will be retired. Microsoft has been explicit: "you'll need to upgrade to Azure Container Networking Interface (CNI) overlay before that date."

Why Kubenet became obsolete: It was a basic network plugin that conserved IP addresses by placing pods on an internal network and using NAT to communicate via the node's IP. But it required complex User-Defined Routes (UDRs), introduced latency, and doesn't support Azure Network Policies or Windows node pools.

**The modern choice is between two advanced CNI modes:**

- **Azure CNI Overlay** (recommended default): Pods get IPs from a private CIDR logically separate from the VNet (e.g., 10.244.0.0/16). This combines CNI's performance with Kubenet's IP conservation, allowing massive cluster scaling without VNet IP exhaustion.

- **Azure CNI with Dynamic IP Allocation** (for VNet-integrated workloads): Pod IPs are still "real" VNet IPs but allocated dynamically from a dedicated Pod Subnet—not pre-allocated to nodes. Use this when VNet IP space is plentiful and you need true VNet integration.

The legacy Azure CNI (with pre-allocated IPs) caused severe VNet IP exhaustion—each node pre-allocated 30+ IPs, quickly consuming subnets. Overlay solves this completely.

**3. Identity: The Service Principal Era is Over**

The legacy Service Principal model—where you manage client IDs and secrets—is an anti-pattern. The modern, secure standard is:

- **Cluster-level**: Use a **User-Assigned Managed Identity** for the cluster's own operations (creating load balancers, pulling from ACR, managing DNS).
- **Workload-level**: Enable **Azure AD Workload Identity** so pods can authenticate to Azure services using Kubernetes service accounts federated with Azure Managed Identities—completely secret-less.
- **User access**: Integrate **Azure AD with Kubernetes RBAC** to control API server access using Azure AD groups, not local cluster-admin kubeconfig files.

**4. Resiliency: System and User Node Pools, Across Availability Zones**

Production node pools must:
- Be separated into **System pools** (tainted for critical add-ons like CoreDNS) and **User pools** (for application workloads)
- Deploy across at least **three Availability Zones**
- Have a minimum of 3 nodes in the system pool for high availability

**5. Observability: Azure Monitor Container Insights is the 80% Solution**

Integration with **Azure Monitor Container Insights** and a **Log Analytics Workspace** is standard. This deploys an agent to all nodes, streaming container logs (stdout/stderr), metrics, and Kubernetes events. Logs are queryable with Kusto Query Language (KQL), and it integrates with Azure Managed Prometheus and Grafana.

The 20% solution—self-hosted Prometheus/Grafana—is for teams requiring cloud-agnostic monitoring or finding Log Analytics too costly at extreme scale.

### What This Means for You

If you're starting fresh: these aren't optimizations to add later. They're the defaults. Planton's AKS API is designed to make this production baseline the *low-friction path*.

If you have existing clusters: audit them against these criteria. Clusters on Free tier, using Kubenet, or relying on Service Principals need migration plans. The good news: in-place migration from Kubenet to CNI Overlay is now possible via `az aks update`.

## The Deployment Landscape: A Methodological Survey

The tool you choose to deploy AKS profoundly impacts repeatability, maintainability, and governance. The landscape has evolved from manual operations to fully declarative, state-managed abstractions.

### Level 0: The Anti-Pattern (Portal and CLI Scripting)

#### The Azure Portal

The portal is excellent for learning—its "Cluster configuration presets" (Dev/test, Production, High availability) guide users toward appropriate settings. But it's also the vector for the most common initial configuration errors:

- **SKU selection**: Users select "Free" to avoid cost, foregoing the uptime SLA
- **Network plugin choice**: Portal defaults may lead to selecting deprecated Kubenet
- **Node pool configuration**: Minimal deployments create a single node pool, violating system/user separation
- **IP exhaustion**: When using Azure CNI, failure to provision large enough subnets leads to `insufficientSubnetSize` errors during scaling, often forcing cluster rebuilds

**Verdict:** Acceptable for learning and sandbox environments. Never for production.

#### Imperative Tooling (Azure CLI, PowerShell, SDKs)

The `az aks create` command is the foundation for scripting and CI/CD pipelines. It's the "reference implementation" for automation. But imperative tooling requires custom retry logic, waiter patterns, and error handling that IaC tools provide out of the box.

Scripts don't know if resources already exist, can't detect drift, and lack state management. They're useful for specific orchestration tasks but not as a primary deployment method.

**Verdict:** Better than portal for simple automation, but doesn't scale to complex environments.

### Level 1: Configuration Management (Ansible)

Ansible's `azure.azcollection` provides modules for AKS clusters and node pools. This is popular for teams already invested in Ansible, allowing integration of cluster creation into existing playbooks.

Ansible excels at orchestrating database provisioning *and* configuring applications that use it—provision AKS, deploy workloads, configure DNS, all in one playbook. But for pure infrastructure, dedicated IaC tools are more powerful.

**When it makes sense:** Mixed workflows (infrastructure + configuration) where you're already using Ansible.

**Limitation:** Modules lag behind Azure features. State management is "whatever Azure currently has"—no local state file for drift detection.

**Verdict:** Useful in mixed workflows, but not the strongest choice for complex AKS deployments.

### Level 2: Azure-Native IaC (ARM and Bicep)

#### ARM Templates

Azure Resource Manager (ARM) templates are the "assembly language" of the Azure platform. Verbose JSON files that represent the underlying API contract. All other deployment methods—portal, CLI, Bicep—ultimately resolve to ARM.

#### Bicep

Bicep is Microsoft's modern, domain-specific language created to replace ARM's verbosity. Bicep files transpile to ARM JSON but offer clean, modular, readable syntax.

**Key advantages:**
- **Day 0 support**: New Azure features are available immediately (third-party tools may lag)
- **Preflight validation**: Can check deployments against Azure Policies *before* executing
- **Stateless by default**: The source of truth is Azure itself, not a state file to manage

**When Bicep makes sense:** If your organization is 100% committed to Azure, Bicep is the clear, low-friction choice. The official baseline architecture samples demonstrate highly sophisticated, production-ready configurations out of the box—VNets with six dedicated subnets (system, user, pod, API server, Bastion, jump-box), codifying complex security topologies.

**Limitation:** Azure-only. Can't manage multi-cloud infrastructure.

**Verdict:** Best choice for Azure-only organizations. Day 0 feature support and native policy integration are unmatched.

### Level 3: Cloud-Agnostic IaC (Terraform, Pulumi, OpenTofu)

These tools manage multiple cloud providers with a single workflow, ideal for hybrid and multi-cloud organizations.

#### Terraform/OpenTofu

Terraform is the *de facto* industry standard for cloud-agnostic IaC. The `azurerm` provider is mature, battle-tested, and capable of managing every aspect of production AKS—complex VNet integrations, identity management, node pool configurations.

**Key advantages:**
- **State management**: Tracks infrastructure in a state file, enabling drift detection
- **Plan workflow**: `terraform plan` shows exactly what will change before applying
- **Module ecosystem**: Community modules encapsulate best practices
- **Multi-cloud**: Same tool for AWS, GCP, Azure

**Production patterns:**
- Remote state in Azure Blob Storage with state locking
- Separate state files per environment (dev, staging, prod isolation)
- Module composition for reusable AKS configurations
- Lifecycle rules to prevent accidental deletion

OpenTofu (the Terraform fork) is functionally equivalent—same HCL syntax, same providers.

#### Pulumi

Pulumi uses general-purpose programming languages (TypeScript, Python, Go) to define infrastructure. You use AWS SDK classes and can leverage the full power of the language.

**Key advantages:**
- **Familiar languages**: Developers can use skills they already have
- **Imperative logic**: Easy to express conditional resources, loops, dynamic configurations
- **Type safety**: Compile-time checks for configuration errors
- **ComponentResources**: Build high-level abstractions (e.g., a `ProductionAksCluster` component)

**When Pulumi shines:** If your infrastructure has complex logic—dynamically creating node pools based on external metrics, or tightly integrating cluster provisioning with application lifecycle—Pulumi's programming model makes this natural.

#### The Comparison

| Aspect | Terraform/OpenTofu | Pulumi | Bicep |
|--------|-------------------|--------|-------|
| **Language** | HCL (declarative DSL) | TypeScript, Python, Go | Bicep (DSL) |
| **State** | Self-hosted or Terraform Cloud | Pulumi Service or self-hosted | Azure (stateless) |
| **Azure Integration** | Very mature (azurerm) | Mature (@pulumi/azure-native) | Day 0 support |
| **Multi-cloud** | Excellent | Excellent | Azure-only |
| **Community** | Very large | Growing | Azure ecosystem |

**Verdict:** This is the production-ready approach. Terraform/OpenTofu for maximum community support and multi-cloud consistency. Pulumi when you need programming language power. Bicep for Azure-only organizations.

### Level 4: Kubernetes-Native Abstractions (Crossplane, Cluster API)

This emerging class uses the Kubernetes API itself as a control plane to manage external infrastructure. The "desired state" is a Kubernetes Custom Resource stored in etcd.

#### Cluster API for Azure (CAPZ)

CAPZ brings declarative, Kubernetes-style APIs to cluster creation. It uses a "management cluster" to provision "workload clusters" and can manage both self-managed Kubernetes nodes on Azure VMs *and* managed AKS clusters.

#### Crossplane

Crossplane extends the Kubernetes API to become a universal control plane across all major cloud providers. A platform team defines a Composition (a template of infrastructure—AKS cluster, VNet, database) and exposes it as a simple Custom Resource. Application teams provision entire stacks with `kubectl apply -f my-cluster-claim.yaml`.

**When this makes sense:** For platform engineering teams building internal developer platforms on Kubernetes, Crossplane provides a consistent interface. Developers request a "cluster" and get AKS (or GKE, or EKS) depending on context.

**Challenge:** Requires Kubernetes control plane. Provider lag means not all AKS features are immediately available. Learning curve requires understanding both Kubernetes and Azure.

**Verdict:** Powerful for platform engineering at scale. Overkill for teams just needing to deploy AKS. Best when building platforms that abstract cloud differences for application teams.

## Comparative Analysis of IaC Tools for AKS

### Scenario-Based Tool Selection

The choice is strategic, based on organizational scope and target user:

**For Azure-Only Organizations:** **Bicep** is the clear choice. Day 0 support for new features and native policy integration are unmatched. Its stateless model (Azure is the source of truth) simplifies operations.

**For Multi-Cloud Enterprises:** **Terraform** is the undisputed standard. Mature `azurerm` provider and broad adoption make it the safest choice for managing heterogeneous infrastructure with a single tool.

**For Developer-Centric Platform Teams:** **Pulumi** offers unparalleled programmatic power. If your team building the platform is more comfortable in Go/Python than HCL, Pulumi allows building robust, testable, complex abstractions.

**For Kubernetes-Native Self-Service Platforms:** **Crossplane** is the tool for building a true Kubernetes-native control plane. It's the most "Kubernetes-idiomatic" way to offer infrastructure-as-a-service to internal developers.

### Handling Production Patterns

All major tools handle the full production pattern:

**VNet Integration (Brownfield/Greenfield):**
- Terraform: Uses `data` sources to look up existing VNets, passes subnet IDs as strings
- Pulumi: Programmatic model uses implicit dependencies (pass `subnet.id` output directly)
- Bicep: Modularity excels—baseline samples deploy VNets with six dedicated subnets in a single deployable unit

**Identity and Secret Management:**
- All tools support the modern pattern: define User-Assigned Managed Identity, create RoleAssignments, attach to cluster
- This declarative, secret-less workflow is a significant security and operational improvement over Service Principals

**State Management:**
- Terraform & Pulumi: Rely on state files (managed in cloud storage with locking)
- Bicep: Stateless—Azure platform is the state
- Crossplane: Kubernetes control plane (etcd) is the state

## What Planton Supports (And Why)

Planton's philosophy is **pragmatic abstraction**: expose the 20% of configuration that 80% of users need, while making advanced options accessible.

### Default Approach: Pulumi-Backed Clusters

Planton uses **Pulumi** to provision AKS clusters. Why Pulumi?

1. **Code as configuration**: Easier to embed complex logic (conditional node pools, dynamic configurations)
2. **Type safety**: Catch configuration errors at compile time
3. **Component model**: Build reusable AKS patterns (dev cluster, prod cluster, serverless)
4. **Multi-cloud consistency**: Same approach for AWS, GCP, Azure—learn once, apply everywhere

### The 80/20 Configuration Philosophy

Research shows most AKS deployments configure:
- Region and Kubernetes version
- Control plane SKU (Standard vs Free)
- Network plugin (Azure CNI Overlay)
- VNet and subnet integration
- Node pool configuration (system and user pools)
- Azure AD RBAC integration
- Monitoring (Container Insights + Log Analytics)

Advanced configurations like custom DNS servers, BYOCNI (bring your own CNI), or exotic kubelet settings are rarely used. Planton's spec focuses on common cases while allowing advanced users to drop down to Pulumi modules when needed.

### What the AzureAksClusterSpec Provides

The protobuf API captures essential configuration:

**Essential fields (the 80%):**
- `region`: Azure region (e.g., "eastus")
- `kubernetes_version`: Pin the version (e.g., "1.30") to prevent unintended upgrades
- `vnet_subnet_id`: Primary integration point for brownfield networking
- `network_plugin`: Defaults to `azure_cni` (with Overlay mode)
- `log_analytics_workspace_id`: For monitoring integration

**Security and identity:**
- Azure AD RBAC enabled by default (`disable_azure_ad_rbac: false`)
- User-Assigned Managed Identity support (via implementation)
- Workload Identity enabled automatically

**API server access:**
- `private_cluster_enabled`: Deploy with private endpoint (no public IP)
- `authorized_ip_ranges`: CIDR allowlist for public clusters (prevents wide-open API servers)

### What Should Be Added (Based on Research)

To fully align with the production baseline, the spec should be enhanced:

**1. Control Plane SKU:**
```protobuf
enum ControlPlaneSku {
  STANDARD = 0;  // Default: financially-backed SLA
  FREE = 1;      // Explicit opt-in for dev/test
}
ControlPlaneSku control_plane_sku = 10;
```

**2. Network Plugin Mode:**
```protobuf
enum NetworkPluginMode {
  OVERLAY = 0;   // Default: modern, IP-efficient
  DYNAMIC = 1;   // For VNet-integrated workloads
}
NetworkPluginMode network_plugin_mode = 11;
```

**3. System and User Node Pools:**
```protobuf
message SystemNodePool {
  string vm_size = 1;  // e.g., "Standard_D4s_v5"
  AutoscalingConfig autoscaling = 2;  // min: 3, max: 5
  repeated string availability_zones = 3;  // ["1", "2", "3"]
}

message UserNodePool {
  string name = 1;
  string vm_size = 2;
  AutoscalingConfig autoscaling = 3;
  repeated string availability_zones = 4;
  bool spot_enabled = 5;  // For cost optimization
}

SystemNodePool system_node_pool = 12 [(buf.validate.field).required = true];
repeated UserNodePool user_node_pools = 13;
```

**4. Add-ons Configuration:**
```protobuf
message AddonsConfig {
  bool enable_container_insights = 1;  // Default: true
  bool enable_key_vault_csi_driver = 2;  // Default: true
  bool enable_azure_policy = 3;  // Default: true
  bool enable_workload_identity = 4;  // Default: true
}
AddonsConfig addons = 14;
```

**5. Advanced Networking (optional, for the 20%):**
```protobuf
message AdvancedNetworking {
  string pod_cidr = 1;
  string service_cidr = 2;
  string dns_service_ip = 3;
  repeated string custom_dns_servers = 4;
}
AdvancedNetworking advanced_networking = 15;  // Optional
```

### What's Abstracted Away

**You don't specify:**
- Low-level subnet group details (derived from VNet configuration)
- Security group minutiae (Planton creates appropriate rules)
- Default parameter configurations (sensible defaults applied)

**You do get:**
- Automatic User-Assigned Managed Identity setup
- System and user node pool separation enforced
- Availability Zone distribution for HA
- Container Insights enabled with Log Analytics integration
- Deletion protection in production environments
- Azure AD RBAC integration by default

### Example: Minimal Production AKS Cluster

In Planton, defining a production AKS cluster looks like:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureAksCluster
metadata:
  name: app-production-aks
spec:
  region: eastus
  kubernetes_version: "1.30"
  control_plane_sku: STANDARD  # 99.95% SLA with AZs
  
  network_plugin: AZURE_CNI
  network_plugin_mode: OVERLAY  # IP-efficient, modern default
  
  vnet_subnet_id: /subscriptions/.../resourceGroups/.../providers/Microsoft.Network/virtualNetworks/app-vnet/subnets/nodes-subnet
  
  private_cluster_enabled: false
  authorized_ip_ranges:
    - "203.0.113.0/24"  # Office IPs
    - "198.51.100.0/24"  # CI/CD agents
  
  system_node_pool:
    vmSize: Standard_D4s_v5
    autoscaling:
      min: 3
      max: 5
    availability_zones: ["1", "2", "3"]
  
  user_node_pools:
    - name: general
      vmSize: Standard_D8s_v5
      autoscaling:
        min: 2
        max: 10
      availability_zones: ["1", "2", "3"]
  
  addons:
    enable_container_insights: true
    enable_key_vault_csi_driver: true
    enable_azure_policy: true
    enable_workload_identity: true
  
  log_analytics_workspace_id: /subscriptions/.../resourceGroups/.../providers/Microsoft.OperationalInsights/workspaces/app-logs
```

This minimal configuration gets you:
- Standard tier cluster with 99.95% uptime SLA
- Azure CNI Overlay (modern, IP-efficient networking)
- System node pool (3-5 nodes) across three AZs
- User node pool (2-10 nodes) across three AZs with autoscaling
- Private API server accessible only from authorized IPs
- Azure AD RBAC integration for user access
- Workload Identity enabled for pod-to-Azure authentication
- Container Insights streaming to Log Analytics
- Key Vault CSI driver for secret management
- Azure Policy enforcement

For development/staging:

```yaml
spec:
  control_plane_sku: FREE  # No SLA, cost savings
  system_node_pool:
    vmSize: Standard_D2s_v3
    autoscaling:
      min: 1
      max: 2
    availability_zones: ["1"]  # Single AZ for cost
  user_node_pools:
    - name: dev
      vmSize: Standard_D4s_v5
      autoscaling:
        min: 1
        max: 3
      availability_zones: ["1"]
      spot_enabled: true  # Deep discounts, acceptable evictions
```

## Production Essentials: What You Must Get Right

Regardless of deployment method, certain configurations are **non-negotiable for production**.

### High Availability Architecture

**Minimum:**
- System node pool: 3 nodes across 3 Availability Zones
- User node pool: At least 2 nodes in separate AZs
- Cluster Autoscaler enabled to add/remove nodes based on demand

**Best practice:**
- System pool tainted with `CriticalAddonsOnly=true:NoSchedule` to isolate critical pods
- User pools sized for expected load plus buffer capacity
- Pod anti-affinity rules with `topologyKey: topology.kubernetes.io/zone` to spread replicas

**Why this matters:** If a single Azure datacenter (AZ) fails, your cluster continues operating. Without multi-AZ deployment, AZ outages cause total cluster failure.

### Networking Security

**Best practice:**
- Nodes in private subnets only (no public IPs on nodes)
- Private cluster if accessing from within VNet, or public with strict `authorized_ip_ranges`
- Network Security Groups (NSGs) allowing only necessary traffic
- Azure Network Policy or Calico for pod-level network policies

**Anti-pattern:**
- Public API server with `authorized_ip_ranges: ["0.0.0.0/0"]` (allows entire internet)
- Nodes with public IPs "for easier debugging"
- Wide-open NSG rules

### Identity and Secrets Management

**Best practice:**
- User-Assigned Managed Identity for cluster (not Service Principal)
- Azure AD Workload Identity for pods (secret-less Azure authentication)
- Azure AD integration with Kubernetes RBAC for user access
- Disable local accounts (`manage_local_accounts_disabled: true`)
- Key Vault CSI driver for application secrets

**Anti-pattern:**
- Service Principals with manually-rotated secrets
- Hardcoded credentials in environment variables or ConfigMaps
- Granting every pod the cluster's Managed Identity permissions

### Monitoring and Observability

**Essential monitoring:**
- Container Insights enabled with Log Analytics workspace
- CloudWatch equivalent: use Azure Monitor Workbooks for dashboards
- Alerts on node resource utilization, pod failures, control plane health
- Application Insights for application-level telemetry

**Best practice:**
- Performance Insights enabled (identifies slow queries in integrated Azure databases)
- Log export configured for control plane logs (audit, authenticator, controller-manager)
- Prometheus + Grafana for Kubernetes-native monitoring (Azure Managed versions available)

**Anti-pattern:**
- Setting up monitoring but never acting on alerts
- No logging aggregation (relying on `kubectl logs`)
- Ignoring control plane metrics

### Backup and Disaster Recovery

**Options:**
- **Velero**: Open-source, CNCF-graduated standard for backing up cluster resources and persistent volumes to Azure Blob Storage
- **Azure Backup for AKS**: Native service integrating with Azure Backup Center for scheduled backups of cluster state and Azure Disk PVs

**Best practice:**
- Regular automated backups (daily minimum)
- Test restore process periodically
- Document cluster configuration in code (IaC is your disaster recovery plan)
- For true DR, consider multi-region clusters with Azure Traffic Manager or Front Door

### Cost Optimization

**Strategies:**
- Cluster Autoscaler to scale down unused nodes
- Spot node pools for non-critical, stateless workloads (30-90% cost savings)
- Rightsize node VM SKUs based on actual usage (start smaller, scale up)
- Use Azure Reservations for predictable, long-term workloads (save up to 72%)
- Stop/scale down dev/staging clusters outside business hours

**Cost monitoring:**
- Azure Cost Management + Billing for cluster-level spend
- Kubecost for workload-level cost attribution (which teams/apps cost most)
- Set budget alerts in Azure Cost Management

## Common Anti-Patterns (And How to Avoid Them)

### Anti-Pattern: Free Tier in Production

**What it looks like:** "We'll save $73/month by using the Free tier. We can upgrade later if needed."

**Why it's wrong:** No uptime SLA means no Azure credits or compensation if the control plane fails. When production breaks at 3 AM, "but we saved money" isn't a comfort.

**The fix:** Use Standard tier for production. The $73/month is insurance, not waste.

### Anti-Pattern: Single-AZ Deployment

**What it looks like:** All nodes in a single Availability Zone to "reduce cross-AZ data transfer costs."

**Why it's wrong:** Cross-AZ data transfer is $0.01/GB—negligible compared to outage costs. Single-AZ means entire cluster fails if that datacenter has issues.

**The fix:** Deploy across at least three AZs. The minimal cross-AZ cost is insurance against hours of downtime.

### Anti-Pattern: Ignoring System/User Pool Separation

**What it looks like:** One node pool running both system pods (CoreDNS, metrics-server) and application workloads.

**Why it's wrong:** When application pods consume all resources, critical system pods can fail, breaking cluster functionality (DNS resolution fails, metrics unavailable).

**The fix:** Dedicated system pool with taints ensures critical infrastructure remains isolated and stable.

### Anti-Pattern: Using Service Principals

**What it looks like:** Cluster configured with `service_principal_client_id` and `service_principal_client_secret`.

**Why it's wrong:** Requires manual secret rotation, secrets stored in cluster state, higher risk of credential leakage.

**The fix:** Use User-Assigned Managed Identity. Azure handles identity lifecycle, no secrets to manage.

### Anti-Pattern: Kubenet in New Clusters

**What it looks like:** Selecting Kubenet because "it's simpler" or "we read an old guide."

**Why it's wrong:** Kubenet is deprecated and will be retired in 2028. Migration to CNI requires downtime. Kubenet also prevents Windows node pools and Azure Network Policy.

**The fix:** Use Azure CNI Overlay for new clusters. Enjoy modern networking without IP exhaustion.

### Anti-Pattern: No Deletion Protection

**What it looks like:** Skipping deletion protection because "it's annoying during testing."

**Why it's wrong:** Accidental `terraform destroy` or console click deletes production cluster. Recovery requires restoring from backups (if they exist).

**The fix:** Enable deletion protection in production. The deliberate friction for destructive operations is a feature, not a bug.

## Azure CNI Overlay: The Modern Networking Standard

Azure CNI Overlay deserves special attention because it solves the biggest historical pain point: VNet IP address exhaustion.

### The Evolution of AKS Networking

**Kubenet (deprecated):**
- Pods on internal network (10.244.0.0/16), NAT to node IP
- Conserved VNet IPs but required UDRs and added latency
- No support for Windows nodes or Azure Network Policy
- Being retired March 31, 2028

**Azure CNI (legacy):**
- Pods get "real" VNet IPs from node's subnet
- High performance, direct VNet routing
- Severe IP exhaustion: each node pre-allocates 30+ IPs
- Often led to `insufficientSubnetSize` errors during scaling

**Azure CNI Overlay (modern):**
- Pods get IPs from private, non-VNet CIDR (e.g., 10.244.0.0/16)
- High performance with overlay networking
- Solves VNet IP exhaustion completely
- Supports Azure Network Policy, Calico, and Cilium

**Azure CNI Dynamic IP (alternative modern):**
- Pods get "real" VNet IPs from dedicated Pod Subnet
- No pre-allocation (dynamic assignment)
- Best for VNet-integrated workloads needing direct pod addressability

### When to Use Each

**Use CNI Overlay when:**
- Building new clusters (it's the default recommendation)
- VNet IP space is constrained
- You need large-scale clusters (hundreds of nodes)
- Standard pod networking suffices

**Use CNI Dynamic IP when:**
- Pods need direct VNet addressability (e.g., for legacy VNet-based firewalls)
- VNet IP space is plentiful
- You have dedicated Pod Subnets available

**Migration from Kubenet:**
In-place migration is now possible: `az aks update --network-plugin-mode overlay`. Previously required cluster rebuild. Plan migration before the 2028 deadline.

## Conclusion

The journey from "clicking around in the Azure Portal" to production-ready AKS is really a journey from **hope-driven infrastructure to confidence-driven infrastructure**.

Hope-driven: "I hope this cluster is configured securely." "I hope we can recreate this in DR." "I hope the networking plugin we chose is still supported."

Confidence-driven: "Our infrastructure is versioned in Git. We've tested AZ failover. Identities are managed via Azure AD with Workload Identity. We have multi-AZ node pools. Everything uses modern CNI Overlay. Container Insights streams to Log Analytics. The cluster is protected from accidental deletion."

Infrastructure as Code—whether Terraform, Pulumi, Bicep, or Crossplane—is the foundation of that confidence. It transforms AKS deployment from art (requiring deep Azure expertise and careful portal navigation) to engineering (declarative, reviewable, testable configuration).

The production baseline is no longer subjective. Microsoft has codified it. The networking confusion has been resolved (CNI Overlay is the answer). The identity model has converged (Managed Identities everywhere). The HA pattern is clear (system and user pools across three AZs).

Planton's approach is to meet you where production infrastructure should be: sensible defaults (Standard tier, CNI Overlay, Managed Identity, multi-AZ), essential security baked in (Azure AD RBAC, private/restricted API server, deletion protection), and the flexibility to customize when your needs demand it.

AKS's promise—managed Kubernetes without operational overhead, seamless scaling, deep Azure integration—is real. But it's only realized when deployed with discipline. Use Infrastructure as Code. Enable the Standard tier and multi-AZ deployment. Use Managed Identities. Separate system and user node pools. Monitor proactively. Do these things, and AKS becomes what it was designed to be: a Kubernetes platform that gets out of your way so you can focus on your applications.

The landscape has evolved. The tooling is mature. The anti-patterns are well-documented. The production baseline is codified. There's no longer an excuse for hope-driven Kubernetes infrastructure. The path to production-ready AKS is clear—now walk it.


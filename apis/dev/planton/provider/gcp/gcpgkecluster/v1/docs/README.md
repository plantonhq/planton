# GKE Control Plane: From Pet Clusters to Cattle Infrastructure

## Introduction

For the first decade of cloud-native infrastructure, Kubernetes clusters were treated like **pets**—carefully hand-crafted, lovingly maintained, and nearly impossible to reproduce. Engineers spent hours clicking through cloud consoles, tweaking CIDR blocks, configuring VPC peering, and managing IAM policies. The result? Snowflake clusters that drifted from their documentation, required tribal knowledge to debug, and made disaster recovery a nightmare.

**Google Kubernetes Engine (GKE)** changed the conversation by moving the operational burden of Kubernetes control plane management to Google. But GKE alone doesn't solve the infrastructure-as-code problem. The cluster still needs to be provisioned, configured, and integrated with networking, security, and identity systems. Without automation, you're back to click-ops and configuration drift.

This is where **declarative infrastructure** comes in. GKE clusters should be **cattle, not pets**—defined in code, version-controlled, reproducible, and ephemeral. If a cluster fails or drifts from spec, you destroy and recreate it from the same source of truth. No tribal knowledge required.

## The GKE Control Plane Problem Space

A production-ready GKE cluster is not just "create cluster and go." It involves:

1. **Network Architecture**: VPC-native networking with IP address management for pods and services (secondary ranges), private vs public endpoints, master authorized networks
2. **Security Posture**: Private clusters (nodes without public IPs), Workload Identity (IAM for pods), service account configuration, Binary Authorization
3. **Connectivity**: Cloud NAT for outbound internet access, VPC peering for private services, firewall rules
4. **Operational Concerns**: Release channels (auto-upgrades), maintenance windows, node management policies, logging and monitoring integration
5. **Capacity Planning**: Regional vs zonal clusters, master CIDR allocation, scaling considerations

The reality: **getting all of this right is hard**. Miss a secondary IP range configuration? Your pods can't schedule. Forget Cloud NAT? Your private nodes can't pull Docker images. Misconfigure Workload Identity? Your pods can't access GCP services.

Planton's `GcpGkeCluster` API abstracts this complexity into a focused, production-ready specification. This document explains why GKE is designed this way, how deployment methods evolved, and what the 80/20 scoping decisions mean for your infrastructure.

## The Evolution: From Manual to Declarative

### Level 0: Manual Console Provisioning (The Anti-Pattern)

**What it is**: The Google Cloud Console's "Create Cluster" wizard. You click through 12+ screens, selecting region, network, IP ranges, add-ons, security options, and finally hit "Create."

**When it's used**: During initial GKE exploration, learning, or proof-of-concept work.

**Why it fails in production**:
- **Not repeatable**: Can you recreate this cluster next week with identical settings? Only if you screenshot every screen.
- **Not auditable**: No paper trail of what changed or why.
- **Not versionable**: Can't use Git, can't do code review, can't roll back.
- **Drift prone**: Manual changes accumulate, documentation becomes stale, "production" becomes a mystery box.

**Real-world consequence**: The "we need to rebuild prod but don't remember how it was configured" incident. This happens more often than engineering leaders admit.

**Verdict**: Acceptable for learning, **completely unacceptable for production**.

---

### Level 1: Scripted CLI (Imperative Automation)

**What it is**: Using `gcloud container clusters create` with a long list of flags:

```bash
gcloud container clusters create prod-cluster \
  --region=us-central1 \
  --network=my-vpc \
  --subnetwork=my-subnet \
  --cluster-secondary-range-name=pod-range \
  --services-secondary-range-name=svc-range \
  --enable-private-nodes \
  --enable-private-endpoint=false \
  --master-ipv4-cidr=172.16.0.0/28 \
  --enable-ip-alias \
  --enable-workload-identity \
  --workload-pool=my-project.svc.id.goog \
  --enable-network-policy \
  --release-channel=regular \
  --enable-autorepair \
  --enable-autoupgrade \
  --no-enable-basic-auth \
  --no-issue-client-certificate
```

**The advancement**: It's scriptable. You can version-control the bash script, run it in CI/CD, and theoretically recreate clusters.

**The limitations**:
1. **Imperative, not declarative**: The command says "do this action" but doesn't maintain state. If someone manually changes a cluster setting via the console, your script doesn't know.
2. **No drift detection**: You'd need to write custom logic to compare actual cluster config against desired state.
3. **Updates are manual**: Changing cluster settings requires separate `gcloud container clusters update` commands. You must calculate the delta between current and desired state.
4. **Idempotency is DIY**: You have to check "does this cluster already exist?" before running create.
5. **Flag explosion**: That 14-line command only covers basics. Production clusters need 30+ flags.

**Verdict**: Better than clicking, but still far from production-grade infrastructure management.

---

### Level 2: Infrastructure as Code (The Production Standard)

This is where **real infrastructure automation** begins. IaC tools are:
- **Declarative**: You define the desired end state, not the steps to get there
- **Stateful**: They track what currently exists in the cloud
- **Lifecycle-aware**: They compute diffs and execute only necessary API calls
- **Drift-correcting**: They detect when reality diverges from code and can reconcile

#### Terraform (Industry Standard)

Terraform is the dominant IaC tool for multi-cloud infrastructure. You define resources in HashiCorp Configuration Language (HCL):

```hcl
resource "google_container_cluster" "primary" {
  name     = "prod-cluster"
  location = "us-central1"
  network  = google_compute_network.vpc.id
  subnetwork = google_compute_subnetwork.primary.id

  # Remove default node pool immediately (best practice)
  remove_default_node_pool = true
  initial_node_count       = 1

  # Private cluster configuration
  private_cluster_config {
    enable_private_nodes    = true
    enable_private_endpoint = false
    master_ipv4_cidr_block  = "172.16.0.0/28"
  }

  # VPC-native networking (required for modern GKE)
  ip_allocation_policy {
    cluster_secondary_range_name  = "pod-range"
    services_secondary_range_name = "svc-range"
  }

  # Workload Identity (IAM for Pods)
  workload_identity_config {
    workload_pool = "${var.project_id}.svc.id.goog"
  }

  # Network policy enforcement (Calico)
  addons_config {
    network_policy_config {
      disabled = false
    }
  }

  # Auto-upgrade strategy
  release_channel {
    channel = "REGULAR"
  }
}
```

On `terraform apply`, Terraform:
1. Reads its state file (the last-known cloud state)
2. Queries GCP APIs (the current actual state)
3. Compares desired state (your HCL) against actual state
4. Generates an execution plan showing what will change
5. Executes only the necessary API calls (create/update/delete)

**The critical advantage**: If you change `master_ipv4_cidr_block`, Terraform knows this forces cluster recreation (GKE API limitation). If you change `release_channel`, Terraform knows this is an in-place update. You don't have to calculate the delta—Terraform does.

**Production pattern**: Terraform state is stored remotely (GCS bucket), enabling team collaboration. State locking prevents concurrent modifications. Modules enable reusability across environments (dev/staging/prod).

#### OpenTofu (Open-Source Terraform Fork)

Identical to Terraform in functionality and HCL syntax. Born from the community response to HashiCorp's 2023 license change (BSL). For organizations requiring open-source governance, OpenTofu is a drop-in Terraform replacement.

#### Pulumi (Modern IaC with General-Purpose Languages)

Pulumi brings infrastructure-as-code into the world of real programming languages (Python, TypeScript, Go, C#):

```python
import pulumi_gcp as gcp

cluster = gcp.container.Cluster("prod-cluster",
    name="prod-cluster",
    location="us-central1",
    network=vpc.id,
    subnetwork=subnet.id,
    remove_default_node_pool=True,
    initial_node_count=1,
    private_cluster_config=gcp.container.ClusterPrivateClusterConfigArgs(
        enable_private_nodes=True,
        enable_private_endpoint=False,
        master_ipv4_cidr_block="172.16.0.0/28",
    ),
    ip_allocation_policy=gcp.container.ClusterIpAllocationPolicyArgs(
        cluster_secondary_range_name="pod-range",
        services_secondary_range_name="svc-range",
    ),
    workload_identity_config=gcp.container.ClusterWorkloadIdentityConfigArgs(
        workload_pool=f"{project_id}.svc.id.goog",
    ),
    addons_config=gcp.container.ClusterAddonsConfigArgs(
        network_policy_config=gcp.container.ClusterAddonsConfigNetworkPolicyConfigArgs(
            disabled=False,
        ),
    ),
    release_channel=gcp.container.ClusterReleaseChannelArgs(
        channel="REGULAR",
    ),
)
```

**Why it's powerful**: You get the full expressiveness of a programming language—loops, conditionals, functions, types, IDE autocomplete. Complex logic that requires 200 lines of HCL templating can be 20 lines of Python.

**Trade-off**: Requires developers comfortable with the chosen language. HCL is intentionally limited; Pulumi lets you shoot yourself in the foot with complex abstractions.

**Verdict**: Terraform/OpenTofu and Pulumi are the **production standard**. They are declarative, stateful, drift-aware, and enable true infrastructure-as-code workflows.

---

### Level 3: Kubernetes-Native GitOps (The Unified Control Plane)

The next evolution: **managing infrastructure using the Kubernetes API itself**.

#### Config Connector (Google's First-Party Kubernetes Operator)

Config Connector is a Kubernetes operator that adds Google Cloud resource types as Custom Resource Definitions (CRDs). You define GCP infrastructure as Kubernetes manifests:

```yaml
apiVersion: container.cnrm.cloud.google.com/v1beta1
kind: ContainerCluster
metadata:
  name: prod-cluster
spec:
  location: us-central1
  networkRef:
    name: my-vpc
  subnetworkRef:
    name: my-subnet
  privateClusterConfig:
    enablePrivateNodes: true
    enablePrivateEndpoint: false
    masterIpv4CidrBlock: 172.16.0.0/28
  ipAllocationPolicy:
    clusterSecondaryRangeName: pod-range
    servicesSecondaryRangeName: svc-range
  workloadIdentityConfig:
    workloadPool: my-project.svc.id.goog
  addonsConfig:
    networkPolicyConfig:
      disabled: false
  releaseChannel:
    channel: REGULAR
  removeDefaultNodePool: true
  initialNodeCount: 1
```

You `kubectl apply` this manifest. The Config Connector operator, running inside your Kubernetes cluster, watches for these custom resources and makes the corresponding GCP API calls.

**The paradigm shift**: Infrastructure and applications share a **unified control plane**. The same GitOps tools (ArgoCD, Flux) that deploy your microservices also provision your GKE clusters. The same RBAC, audit logs, and observability tools work for both.

**The self-provisioning cluster**: A GKE cluster can provision its own node pools, databases, DNS zones, and load balancers. The cluster becomes "self-describing" infrastructure.

**The bootstrapping problem**: You still need *something* to create the first cluster that runs Config Connector. This is usually Terraform or Pulumi. Once bootstrapped, Config Connector can manage everything else.

#### Crossplane (Vendor-Neutral, Composable Multi-Cloud)

Crossplane extends the Kubernetes-native infrastructure pattern with:
1. **Multi-cloud**: Manage GCP, AWS, Azure resources with the same workflow
2. **Composite resources**: Platform teams define high-level abstractions (e.g., `XCluster`) that compose lower-level resources (VPC + Subnets + GKE + Node Pools + Databases)
3. **Developer self-service**: Developers request an `XCluster`, and Crossplane provisions all underlying infrastructure

**The strategic advantage**: Platform engineering teams build **internal developer platforms** that abstract cloud complexity. Developers don't need to understand VPC secondary ranges or Workload Identity—they just request a "cluster" and get production-ready infrastructure.

**The complexity cost**: Crossplane is operationally complex. The operator must be highly available. Composite resource definitions require deep Kubernetes and cloud expertise.

**Verdict**: Cutting-edge. Ideal for platform teams building self-service infrastructure, but requires investment in operator reliability and sophisticated abstractions.

---

## Planton's Approach: Guard Rails, Not Handcuffs

The `GcpGkeCluster` API resource is designed around a principle: **make the right thing easy and the wrong thing hard**.

### Core Philosophy

Production GKE clusters follow a predictable pattern:
- **Private nodes** (no public IPs) with Cloud NAT for outbound internet
- **VPC-native networking** (IP aliasing for pods and services)
- **Workload Identity** (secure IAM for pods without node-level service account keys)
- **Network Policy enforcement** (Calico for microsegmentation)
- **Release channels** (auto-upgrades with Regular channel as default)

This isn't opinion—it's the consolidated wisdom of Google's own best practices documentation, enterprise architecture patterns, and production war stories.

### What the API Includes (The 80%)

The `GcpGkeClusterSpec` focuses on **control plane configuration**—the essential, immutable decisions that define a cluster's networking, security, and upgrade strategy:

#### Core Networking (Required)
- **`project_id`**: GCP project (foreign key to `GcpProject` resource)
- **`location`**: Region (e.g., `us-central1`) or zone (e.g., `us-central1-a`)
  - Regional clusters: Multi-zonal high availability (control plane and nodes across 3+ zones)
  - Zonal clusters: Single-zone (cheaper, less resilient)
- **`subnetwork_self_link`**: VPC subnet (foreign key to `GcpSubnetwork`)
- **`cluster_secondary_range_name`**: Secondary IP range for Pod IPs
- **`services_secondary_range_name`**: Secondary IP range for Service IPs

**Why VPC-native (IP aliasing) is mandatory**: Legacy "routes-based" clusters are deprecated and lack modern features (Workload Identity, Private Service Connect, some GKE add-ons). All new clusters should be VPC-native.

#### Private Cluster Configuration (Required)
- **`master_ipv4_cidr_block`**: RFC1918 /28 CIDR for the Kubernetes API server (control plane) private endpoint
  - Must be /28 (exactly 16 IPs)
  - Must not overlap with VPC, pod, or service ranges
  - Example: `172.16.0.0/28`

**Why private by default**: Public control plane endpoints are a security liability. Private endpoints limit API access to VPC networks (or authorized networks via VPN/Interconnect).

- **`enable_public_nodes`**: Boolean to allow nodes with public IPs (default: `false`)
  - If `false`: Nodes are private (RFC1918 only), requires Cloud NAT for internet egress
  - If `true`: Nodes get external IPs (legacy pattern, higher attack surface)

#### Connectivity (Required)
- **`router_nat_name`**: Reference to Cloud NAT configuration (foreign key to `GcpRouterNat`)
  - Required for private nodes to pull images, reach APIs, install packages
  - Cloud NAT provides managed, scalable outbound internet without exposing inbound attack surface

#### Operational Configuration (Optional with Defaults)
- **`release_channel`**: Auto-upgrade strategy (default: `REGULAR`)
  - **RAPID**: Bleeding-edge (new Kubernetes versions ~2-3 months before Stable)
  - **REGULAR**: Production-recommended balance (new versions ~2-3 months after Rapid)
  - **STABLE**: Conservative (new versions ~2-3 months after Regular)
  - **NONE**: Manual upgrades (not recommended—you manage security patching)

**Best practice**: Use `REGULAR` for production. It balances access to new features with stability.

- **`disable_network_policy`**: Boolean to turn off Calico network policy enforcement (default: `false`)
  - If `false`: Network policies enabled (recommended for production security)
  - If `true`: All pods can reach all pods (simpler, less secure)

- **`disable_workload_identity`**: Boolean to turn off Workload Identity (default: `false`)
  - If `false`: Workload Identity enabled (pods use KSA→GSA binding for IAM)
  - If `true`: Legacy metadata server (pods inherit node service account—overly permissive)

**Why Workload Identity is critical**: It follows the principle of least privilege. Each pod gets exactly the IAM permissions it needs via Kubernetes Service Account annotations. No shared node-level service account keys.

### What the API Excludes (The 20%)

These are **deliberately excluded** to maintain focus on control plane configuration:

#### Node Pools (Separate Resource)
Node pools are an **independent resource** (`GcpGkeNodePool`). This aligns with GKE's API design and mirrors Terraform/Pulumi best practices. Reasons:
- **Lifecycle independence**: Change machine types, scale pools, add GPU pools—without touching the control plane
- **Avoid unintended side effects**: Inline node pool configs in cluster definitions can trigger unwanted cluster-level operations
- **Production pattern**: Validated by Terraform docs, Pulumi patterns, and real-world usage

#### Add-ons (Opinionated Defaults)
GKE add-ons (HTTP Load Balancing, Horizontal Pod Autoscaling, Cloud Logging, Cloud Monitoring) are enabled by default. Disabling them is rare. The API doesn't expose every add-on toggle to reduce surface area.

If you need granular add-on control, use the underlying Pulumi/Terraform providers directly.

#### Binary Authorization, Advanced Security Features
Features like Binary Authorization (image attestation), Shielded GKE Nodes (Secure Boot), and GKE Sandbox (gVisor isolation) are advanced security patterns. Including them would clutter the API for the 80% use case.

**Extension path**: These can be added as optional fields in future API versions without breaking existing resources.

#### Master Authorized Networks (Advanced)
Master authorized networks restrict API server access to specific CIDR ranges (e.g., corporate VPN IPs). This is important for highly regulated environments but not a universal requirement.

**Workaround**: Use VPC firewall rules and private endpoints for access control.

---

## The 80/20 Scoping Decisions

Planton's `GcpGkeCluster` is opinionated but extensible. The goal: **80% of users get production-ready defaults; 20% with advanced needs can drop to Terraform/Pulumi**.

| Feature | Included? | Why? |
|---------|-----------|------|
| VPC-native networking | ✅ Required | Routes-based is deprecated; IP aliasing is mandatory for modern features |
| Private cluster config | ✅ Required | Security best practice; public clusters are legacy anti-pattern |
| Cloud NAT reference | ✅ Required | Private nodes can't function without NAT for egress |
| Workload Identity | ✅ Enabled by default | IAM for pods; principle of least privilege |
| Network Policy (Calico) | ✅ Enabled by default | Microsegmentation; production security baseline |
| Release channel | ✅ Configurable (default: REGULAR) | Auto-upgrades are critical for security; channel choice matters |
| Node pools | ❌ Separate resource | Lifecycle independence; aligns with Terraform/Pulumi patterns |
| Master authorized networks | ❌ Excluded | Advanced use case; achievable via VPC firewall rules |
| Binary Authorization | ❌ Excluded | Advanced security; not universal requirement |
| Add-on granular control | ❌ Opinionated defaults | Defaults are production-ready; reduces API surface |

---

## Production Best Practices

### Private Clusters: The Modern Default

**Anti-pattern**: Public clusters with nodes that have external IPs.

**Why it fails**: Every node is exposed to the internet. Even with firewall rules, you've expanded the attack surface. If a node is compromised, lateral movement and data exfiltration are easier.

**Best practice**: **Private clusters with private nodes**:
- Nodes have only RFC1918 (private) IPs
- Control plane endpoint is private (accessible only from VPC or authorized networks)
- Outbound internet access via Cloud NAT (managed, stateless, no inbound ports)

**The trade-off**: SSH access to nodes requires a bastion host or IAP (Identity-Aware Proxy). This is intentional—nodes should be immutable and auto-replaceable, not SSH destinations.

---

### VPC-Native Networking: Non-Negotiable

**Legacy approach**: Routes-based clusters use Google Cloud routes for pod IP routing. This is deprecated.

**Modern approach**: VPC-native clusters use **alias IP ranges**:
- Pods get IPs from a dedicated secondary range on the subnet
- Services get IPs from another secondary range
- GCP's VPC natively understands these ranges (no custom routes)

**Why it matters**:
- **Workload Identity** requires VPC-native
- **Private Service Connect** (private access to Google APIs) requires VPC-native
- **GKE Dataplane V2** (eBPF-based networking) requires VPC-native
- **Shared VPC** patterns work better with VPC-native

**Verdict**: Always use VPC-native. Planton enforces this by requiring secondary range names in the spec.

---

### Workload Identity: IAM for Pods Without Keys

**Anti-pattern**: Nodes use the Compute Engine default service account (overly permissive). Pods inherit node-level permissions. If one pod is compromised, all pods share the same IAM blast radius.

**Modern approach**: **Workload Identity**:
1. Each Kubernetes Service Account (KSA) is annotated with a Google Service Account (GSA)
2. GKE's metadata server intercepts credential requests from pods
3. Pods receive short-lived GSA tokens scoped to the KSA's bound GSA
4. No service account keys, no node-level shared permissions

**Configuration**:

```yaml
# Kubernetes Service Account with Workload Identity binding
apiVersion: v1
kind: ServiceAccount
metadata:
  name: app-sa
  namespace: production
  annotations:
    iam.gke.io/gcp-service-account: app@my-project.iam.gserviceaccount.com
```

```bash
# Bind KSA to GSA
gcloud iam service-accounts add-iam-policy-binding \
  app@my-project.iam.gserviceaccount.com \
  --role=roles/iam.workloadIdentityUser \
  --member="serviceAccount:my-project.svc.id.goog[production/app-sa]"
```

**Result**: The `app-sa` service account in the `production` namespace gets exactly the IAM permissions of `app@my-project.iam.gserviceaccount.com`. No other pods can impersonate this GSA.

**Verdict**: Enable Workload Identity by default. Planton does this via `disable_workload_identity: false`.

---

### Release Channels: Auto-Upgrades Are Your Friend

**Anti-pattern**: Disabling auto-upgrades (`release_channel: NONE`) out of fear of disruption.

**Why it backfires**: Your cluster falls behind on security patches. When you *do* upgrade, the version gap is large and risky. You accumulate technical debt.

**Best practice**: Use **release channels** to control *when* upgrades happen, not *whether* they happen:
- **RAPID**: For dev/test environments that want bleeding-edge features
- **REGULAR**: For production (recommended)—new versions ~2-3 months after Rapid
- **STABLE**: For highly risk-averse production—new versions ~2-3 months after Regular

Combine with **maintenance windows** (define recurring time windows when upgrades are permitted, e.g., "Sunday 02:00-04:00 AM").

**Verdict**: Use `REGULAR` for production. Auto-upgrades minimize security risk.

---

### Network Policies: Microsegmentation for Kubernetes

**Anti-pattern**: Default Kubernetes allows any pod to reach any pod on any port (flat network).

**Why it's risky**: If one pod is compromised, attackers can pivot laterally to databases, internal APIs, or other services.

**Best practice**: Enable **Network Policy** enforcement (Calico in GKE):

```yaml
# Example: Only allow frontend pods to reach backend on port 8080
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: backend-policy
  namespace: production
spec:
  podSelector:
    matchLabels:
      app: backend
  policyTypes:
    - Ingress
  ingress:
    - from:
        - podSelector:
            matchLabels:
              app: frontend
      ports:
        - protocol: TCP
          port: 8080
```

**Result**: Backend pods reject traffic from anything except frontend pods on port 8080.

**Verdict**: Enable network policies by default. Planton does this via `disable_network_policy: false`.

---

### Cloud NAT: Outbound Internet for Private Nodes

Private nodes have no public IPs. How do they:
- Pull Docker images from Docker Hub, Artifact Registry?
- Reach external APIs (Stripe, Twilio, GitHub)?
- Install packages via `apt`, `yum`?

**Answer**: **Cloud NAT** (Network Address Translation)—a managed, regional NAT gateway:
- Provides outbound internet access
- No inbound ports exposed (stateless NAT)
- Auto-scales with your cluster (no VM-based NAT bottlenecks)
- Logs all outbound connections for audit trails

**Configuration**: Create a Cloud Router and Cloud NAT in the same region as your cluster. Planton's `router_nat_name` field ensures this is validated at spec definition time.

**Verdict**: Required for private clusters. Planton enforces this by making `router_nat_name` a required field.

---

## Regional vs Zonal Clusters: Availability vs Cost

### Zonal Clusters
- **Control plane**: Single zone (e.g., `us-central1-a`)
- **Nodes**: Can span multiple zones, but control plane is in one zone
- **SLA**: Lower (control plane failure = API downtime)
- **Cost**: Lower (one control plane replica)

**Use case**: Dev/test environments, cost-sensitive non-critical workloads

### Regional Clusters
- **Control plane**: Multi-zonal (replicas in 3+ zones)
- **Nodes**: Automatically spread across multiple zones
- **SLA**: Higher (control plane survives zonal outages)
- **Cost**: Higher (3+ control plane replicas)

**Use case**: Production workloads requiring high availability

**Verdict**: Use **regional clusters for production**. The control plane cost is marginal compared to node costs, and the resilience is worth it.

---

## Control Plane and Node Separation: The Architecture That Matters

One of GKE's core design principles: **control plane and nodes are separate**. The control plane (Kubernetes API server, etcd, controllers) is managed by Google. You provision the cluster (control plane config), then separately provision node pools.

This separation is intentional:
1. **Control plane upgrades** don't touch nodes
2. **Node pool changes** (machine types, autoscaling, labels) don't risk control plane disruptions
3. **Node pool lifecycles** are independent (create/update/delete pools without recreating the cluster)

Planton's API reflects this:
- `GcpGkeCluster`: Control plane configuration (networking, security, upgrade strategy)
- `GcpGkeNodePool`: Node configuration (machine types, autoscaling, taints, labels)

This mirrors Terraform and Pulumi best practices: `google_container_cluster` (control plane) and `google_container_node_pool` (nodes) are separate resources.

**Anti-pattern**: Inline node pools within cluster definitions. Terraform docs explicitly warn this can cause IaC tools to "struggle with complex changes" and risk unintended cluster recreation.

**Verdict**: Keep them separate. Planton enforces this architectural boundary.

---

## Master CIDR Planning: The /28 You Can't Change

Every GKE private cluster requires a **master CIDR block**—a /28 range (16 IPs) for the Kubernetes control plane's private endpoint.

**Critical constraints**:
1. Must be /28 (exactly 16 IPs)—no larger, no smaller
2. Must be RFC1918 (private): `10.0.0.0/8`, `172.16.0.0/12`, or `192.168.0.0/16`
3. Must not overlap with:
   - VPC primary range
   - Pod secondary range
   - Service secondary range
   - Any peered VPC ranges
   - Any VPN/Interconnect on-premises ranges
4. **Cannot be changed after cluster creation** (immutable)

**Planning strategy**: Reserve a dedicated /16 subnet for GKE master CIDRs. Example:
- `172.16.0.0/28` → Cluster 1 masters
- `172.16.0.16/28` → Cluster 2 masters
- `172.16.0.32/28` → Cluster 3 masters
- ...up to 4096 clusters in `172.16.0.0/16`

**Verdict**: Plan your master CIDR allocation strategy before creating clusters. Planton's `master_ipv4_cidr_block` field enforces /28 via regex validation.

---

## Shared VPC: Enterprise Multi-Project Networking

In enterprise GCP organizations, networking is centralized:
- **Host project**: Owns the VPC, subnets, firewall rules
- **Service projects**: Consume subnets from the host project

**Shared VPC for GKE**:
1. GKE cluster is in a service project
2. Cluster references subnets from the host project
3. Nodes and pods use IP ranges from the shared VPC
4. Firewall rules, Cloud NAT, and VPN are managed centrally

**Why it matters**: Separates network administration (platform team) from application deployment (product teams). Platform team controls IP allocation and security policies; product teams deploy clusters within those guard rails.

**Planton support**: The `subnetwork_self_link` field can reference subnets from any project (host or same project). For Shared VPC, use:
- `project_id`: Service project (where cluster is created)
- `subnetwork_self_link`: Subnet from host project

---

## The GKE Add-Ons Landscape

GKE enables several add-ons by default:
- **HTTP Load Balancing**: Provisions GCP Load Balancers for `Ingress` resources
- **Horizontal Pod Autoscaler**: Scales pods based on CPU/memory metrics
- **Cloud Logging**: Exports cluster logs to Cloud Logging
- **Cloud Monitoring**: Exports cluster metrics to Cloud Monitoring

Most production clusters keep these enabled. Planton does not expose granular add-on toggles to reduce API surface area.

**Advanced use cases**: If you need custom Ingress controllers (e.g., NGINX, Traefik) or third-party observability (e.g., Datadog, Prometheus), you can disable GCP's HTTP Load Balancing. This is rare. For most users, the defaults are production-ready.

---

## The Node Pool Lifecycle: Why Separation Matters

Node pools have independent lifecycles:
- **Create**: Add a new pool (GPU nodes, high-memory nodes) without touching the cluster
- **Update**: Change machine types, autoscaling ranges, labels—forces node recreation but cluster control plane is unaffected
- **Delete**: Remove a pool (drain nodes, terminate VMs)—no impact on other pools or cluster

**Critical pattern**: GKE's best practice is to **remove the default node pool** and create custom pools:

```hcl
resource "google_container_cluster" "primary" {
  # ...
  remove_default_node_pool = true
  initial_node_count       = 1  # Temporary; deleted after cluster creation
}

resource "google_container_node_pool" "custom" {
  cluster = google_container_cluster.primary.name
  # ... custom configuration
}
```

**Why?** The default node pool has generic settings. Production workloads need customized pools (machine types, autoscaling, labels, taints).

Planton's Pulumi implementation follows this pattern: `remove_default_node_pool = true`, then separately provision `GcpGkeNodePool` resources.

---

## Example Configurations

### Minimal Production Cluster (Private, Regional, Auto-Upgrades)

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpGkeCluster
metadata:
  name: prod-cluster
spec:
  project_id:
    value_ref:
      kind: GcpProject
      name: my-project
      path: status.outputs.project_id
  location: us-central1  # Regional for HA
  subnetwork_self_link:
    value_ref:
      kind: GcpSubnetwork
      name: prod-subnet
      path: status.outputs.self_link
  cluster_secondary_range_name:
    value_ref:
      kind: GcpSubnetwork
      name: prod-subnet
      path: status.outputs.pods_secondary_range_name
  services_secondary_range_name:
    value_ref:
      kind: GcpSubnetwork
      name: prod-subnet
      path: status.outputs.services_secondary_range_name
  master_ipv4_cidr_block: 172.16.0.0/28
  enable_public_nodes: false  # Private nodes
  release_channel: REGULAR    # Auto-upgrades
  disable_network_policy: false  # Enable Calico
  disable_workload_identity: false  # Enable Workload Identity
  router_nat_name:
    value_ref:
      kind: GcpRouterNat
      name: prod-nat
      path: metadata.name
```

**Result**: Regional private cluster with VPC-native networking, Workload Identity, network policies, and auto-upgrades on the Regular channel.

---

### Dev Cluster (Zonal, Cost-Optimized)

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpGkeCluster
metadata:
  name: dev-cluster
spec:
  project_id:
    value: dev-project-12345
  location: us-central1-a  # Zonal (cheaper)
  subnetwork_self_link:
    value: https://www.googleapis.com/compute/v1/projects/dev-project/regions/us-central1/subnetworks/dev-subnet
  cluster_secondary_range_name:
    value: dev-pods
  services_secondary_range_name:
    value: dev-services
  master_ipv4_cidr_block: 172.16.0.16/28
  enable_public_nodes: false
  release_channel: RAPID  # Bleeding-edge for dev
  disable_network_policy: true  # Simplify dev debugging
  disable_workload_identity: true  # Not needed for dev
  router_nat_name:
    value: dev-nat
```

**Result**: Zonal dev cluster with relaxed security (no network policies, no Workload Identity) for faster iteration.

---

### High-Security Prod Cluster (Stable Channel, All Security Features)

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpGkeCluster
metadata:
  name: secure-prod
spec:
  project_id:
    value: security-project-67890
  location: us-central1
  subnetwork_self_link:
    value: https://www.googleapis.com/compute/v1/projects/security-project/regions/us-central1/subnetworks/secure-subnet
  cluster_secondary_range_name:
    value: secure-pods
  services_secondary_range_name:
    value: secure-services
  master_ipv4_cidr_block: 172.16.0.32/28
  enable_public_nodes: false
  release_channel: STABLE  # Conservative upgrades
  disable_network_policy: false  # Maximum security
  disable_workload_identity: false  # IAM per pod
  router_nat_name:
    value: secure-nat
```

**Result**: Regional cluster on Stable channel with all security features enabled for risk-averse production.

---

## Conclusion: Control Plane First, Nodes Second

GKE clusters are **not monolithic**. They're composed of:
1. **Control plane** (API server, etcd, controllers)—configuration captured in `GcpGkeCluster`
2. **Node pools** (VMs running kubelet)—configuration captured in `GcpGkeNodePool`

This separation is architectural and intentional. It mirrors GKE's API design, Terraform best practices, Pulumi patterns, and real-world production stability requirements.

Planton's `GcpGkeCluster` API encodes the lessons learned from thousands of production GKE deployments:
- Private clusters with VPC-native networking by default
- Workload Identity for IAM-per-pod without shared secrets
- Network policies for microsegmentation
- Release channels for security-conscious auto-upgrades
- Cloud NAT for private node internet egress
- Master CIDR validation to prevent configuration mistakes

The goal: **production-ready GKE clusters in 10 lines of YAML, not 100 lines of HCL**.

The 80/20 philosophy ensures that simple use cases remain simple, while the remaining 20% (advanced security, add-on customization, specialized networking) can be addressed by extending the API or dropping to Terraform/Pulumi.

From pet clusters maintained by tribal knowledge to cattle infrastructure defined in Git—this is the evolution Planton enables.


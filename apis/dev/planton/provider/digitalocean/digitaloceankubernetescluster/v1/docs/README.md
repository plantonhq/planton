# DigitalOcean Kubernetes (DOKS) Deployment Methods

## Introduction: Simplicity Meets Production-Readiness

For years, managed Kubernetes was synonymous with complexity. Provisioning a cluster on AWS EKS meant navigating VPCs, subnets, security groups, IAM roles, and a labyrinth of configuration options—all before your first pod could run. DigitalOcean Kubernetes Service (DOKS) challenged that paradigm with a simple proposition: **what if managed Kubernetes could be both simple and production-ready?**

The result is a platform that provisions clusters in minutes, charges **$0 for the control plane**, and integrates seamlessly with DigitalOcean's developer-friendly ecosystem—all while delivering the full power of CNCF-certified Kubernetes.

This document explains the deployment methods available for DigitalOcean Kubernetes clusters, from manual provisioning to fully automated Infrastructure as Code (IaC) patterns, and how Planton exposes this platform through a minimal, security-first API.

## Why DigitalOcean Kubernetes Exists

### The Developer-First Positioning

DigitalOcean doesn't try to compete with AWS, GCP, or Azure on ecosystem breadth. Instead, it competes on three strategic pillars:

1. **Simplicity**: "Heroku simplicity with Kubernetes power"—minimal configuration surface, sane defaults, clean UI and CLI
2. **Cost Transparency**: Free control plane, straightforward node pricing (just Droplet costs), no hidden fees or complex billing tiers
3. **Speed**: Cluster provisioning in 3-5 minutes vs. 10-15+ minutes for GKE/EKS/AKS

This makes DOKS ideal for **startups, indie developers, cost-conscious production workloads, and development/CI environments** that value velocity and predictability over deep cloud ecosystem lock-in.

### Key Differentiators

| Feature | DOKS | GKE/EKS/AKS |
|---------|------|-------------|
| **Control Plane Cost** | Free (non-HA) or $40/month (HA, often waived for 3+ nodes) | ~$70-73/month minimum |
| **Provision Time** | 3-5 minutes | 10-15+ minutes |
| **Configuration Complexity** | Low (defaults work for 80% of cases) | High (requires VPC/IAM/networking expertise) |
| **Node Pricing** | Standard Droplet pricing (~$12-20/month for 2-4GB nodes) | Variable, often higher minimums |
| **Networking** | Cilium CNI (built-in), Hubble for observability | Varies (often requires add-ons) |
| **Storage Integration** | Automatic CSI driver for Block Storage | Automatic (cloud-specific drivers) |
| **Load Balancer Integration** | Automatic DigitalOcean LB provisioning (~$10/month flat) | Automatic (cloud-specific, pricing varies) |

### The DigitalOcean Ecosystem Integration

DOKS isn't an isolated service—it's designed to work seamlessly with DigitalOcean's broader infrastructure:

- **VPCs**: Clusters run in isolated Virtual Private Cloud networks by default
- **Load Balancers**: Creating a Service of `type: LoadBalancer` automatically provisions a managed load balancer
- **Block Storage**: PersistentVolumeClaims automatically provision DigitalOcean Volumes via the CSI driver
- **Container Registry (DOCR)**: Optional one-click integration adds registry credentials to the cluster
- **Spaces**: S3-compatible object storage for backups and static assets
- **Managed Databases**: Offload stateful services to managed PostgreSQL, MySQL, Redis, etc.

This tight integration means you can build a complete cloud stack—Kubernetes clusters, databases, object storage, networking—entirely within DigitalOcean's unified control plane.

## The Deployment Spectrum

DigitalOcean provides multiple deployment methods, all ultimately consuming the same **DigitalOcean v2 REST API**. All programmatic methods require authentication via a **Personal Access Token (PAT)**.

### Level 0: Manual Provisioning (The Anti-Pattern)

**Method**: DigitalOcean Cloud Control Panel (web UI)  
**Verdict**: ❌ Suitable for learning only, never for production

The web console provides a simple wizard: select region, name, Kubernetes version, node pool size/count, and optionally configure HA, VPC, firewall rules, and Marketplace add-ons.

This is the fastest way to *see* a DOKS cluster, but it's not repeatable, not auditable, and creates configuration drift. **Never use manual provisioning for production infrastructure.**

### Level 1: Scripted Provisioning

**Method**: DigitalOcean CLI (`doctl`)  
**Verdict**: ⚠️ Acceptable for scripts and ephemeral clusters, but not ideal for long-lived infrastructure

The official `doctl` CLI is a Go-based tool that wraps the DigitalOcean API:

```bash
doctl kubernetes cluster create my-cluster \
  --region nyc1 \
  --version 1.28.2-do.0 \
  --size s-2vcpu-2gb \
  --count 3 \
  --auto-upgrade=true \
  --ha=false \
  --tag env:dev
```

After creation, `doctl` can automatically save the kubeconfig to your local `~/.kube/config`, making the cluster immediately usable with `kubectl`.

**Authentication**: Initialize with `doctl auth init` or set the `DO_PAT` environment variable.

**Use cases**:
- Quick experimentation and debugging
- CI/CD pipelines for ephemeral test clusters
- Shell scripts for simple automation tasks

**Limitations**:
- No state management (can't detect drift)
- Imperative (must manually track what exists)
- Changes require explicit re-creation or updates

### Level 2: Direct API Integration

**Method**: DigitalOcean v2 REST API or Go SDK (`godo`)  
**Verdict**: ⚠️ Use only for custom platform building

For custom tooling, you can call the REST API (`POST https://api.civo.com/v2/kubernetes/clusters`) or use the official **Go SDK** (`github.com/digitalocean/godo`). This SDK is the foundation of both `doctl` and the Terraform provider, making it mature and production-grade.

Use this approach only if you're building a custom control plane or platform layer (e.g., multi-tenant cluster provisioning systems).

### Level 3: Infrastructure as Code (The Production Solution)

**Methods**: Terraform, Pulumi, OpenTofu, Ansible  
**Verdict**: ✅ **Recommended for all production deployments**

This is the gold standard for managing DOKS infrastructure. IaC provides:
- **Declarative configuration**: Define desired state, not imperative steps
- **Version control**: Infrastructure lives in Git alongside application code
- **State management**: Detect drift and prevent conflicting changes
- **Reusability**: Modules/stacks for multi-environment deployments

#### Terraform

DigitalOcean maintains the **official Terraform provider** (`digitalocean/digitalocean`) in the HashiCorp Registry. It's mature, actively developed, and widely used in production.

**Example: Production Cluster**

```hcl
resource "digitalocean_kubernetes_cluster" "prod_cluster" {
  name    = "prod-cluster"
  region  = "sfo3"
  version = "1.27.5-do.0"
  ha      = true
  auto_upgrade = true

  node_pool {
    name       = "default"
    size       = "s-2vcpu-4gb"
    count      = 3
    auto_scale = true
    min_nodes  = 3
    max_nodes  = 6
  }

  tags = ["env:prod"]
}
```

State is managed via remote backends (S3, GCS, or Terraform Cloud). Secrets (API tokens) are passed via environment variables.

**Key capabilities**:
- Supports all DOKS features (HA, autoscaling, multiple node pools, VPC configuration, control plane firewalls)
- Separate `digitalocean_kubernetes_node_pool` resource for additional pools
- Large community of examples and modules

**Limitations**:
- Certain changes (like resizing nodes in an existing pool) require pool recreation due to API constraints
- Feature lag: new DOKS features may take time to appear in the provider

#### Pulumi

Pulumi provides an official DigitalOcean provider, enabling cluster definitions in general-purpose languages (TypeScript, Python, Go, etc.).

**Advantages over Terraform**:
- **Programmatic IaC**: Use loops, conditionals, functions, and type checking
- **State management**: Defaults to managed Pulumi Cloud with automatic concurrency locking
- **Secret management**: Built-in encrypted configuration (`pulumi config set digitalocean:token --secret <value>`)
- **Multi-environment**: Stacks (e.g., `my-app-staging`, `my-app-prod`) with per-stack configuration

**Example: Staging Cluster (TypeScript)**

```typescript
import * as digitalocean from "@pulumi/digitalocean";

const cluster = new digitalocean.KubernetesCluster("staging-cluster", {
    name: "staging",
    region: "nyc1",
    version: "1.28.2-do.0",
    ha: false,
    autoUpgrade: true,
    nodePool: {
        name: "default",
        size: "s-2vcpu-2gb",
        nodeCount: 2,
        autoScale: false,
    },
    tags: ["env:staging"],
});

export const kubeconfig = cluster.kubeConfigs[0].rawConfig;
```

#### OpenTofu

As a compatible fork of Terraform, OpenTofu works seamlessly with the `digitalocean/digitalocean` provider. No code changes required—simply use `tofu` instead of `terraform` commands.

#### Ansible

Ansible can create and delete DOKS clusters via the `community.digitalocean.digital_ocean_kubernetes` module.

**Example Playbook Task**

```yaml
- name: Create DOKS cluster
  community.digitalocean.digital_ocean_kubernetes:
    state: present
    oauth_token: "{{ lookup('env', 'DO_API_TOKEN') }}"
    name: "mycluster"
    region: "nyc1"
    nodePools:
      - name: default
        size: s-1vcpu-2gb
        count: 3
    auto_upgrade: true
    ha: false
    wait: true
```

**Use cases**:
- Organizations already using Ansible for configuration management
- Combining cluster provisioning with application deployment in one playbook

**Limitations**:
- No state management (unlike Terraform/Pulumi)
- Weaker drift detection
- Module may lag behind new DOKS features
- Best suited for ephemeral or dev clusters, not long-lived production infrastructure

**Comparison: IaC Tools for DOKS**

| Feature | Terraform | Pulumi | Ansible |
|---------|-----------|--------|---------|
| **State Management** | Remote backend (S3, TFC) | Pulumi Cloud (default) | None (idempotent tasks) |
| **Drift Detection** | ✅ Strong | ✅ Strong | ⚠️ Weak |
| **Resource Coverage** | Complete | Complete (matches Terraform) | Good (some lag) |
| **Community Maturity** | ✅ Very high | ✅ High | ⚠️ Moderate |
| **Multi-Environment** | Workspaces / Directory layouts | Stacks | Variable files |
| **Secret Management** | External (env vars, Vault) | Built-in encrypted config | External (Ansible Vault) |
| **Production Readiness** | ✅ Excellent | ✅ Excellent | ⚠️ Moderate (best for dev/test) |

**Verdict**: Use **Terraform** for maximum maturity and ecosystem compatibility. Use **Pulumi** for type-safe, programmatic IaC with superior developer experience. Use **Ansible** only for dev/test clusters or when already standardized on Ansible.

### Level 4: Kubernetes-Native Provisioning

**Method**: Crossplane (`provider-digitalocean`)  
**Verdict**: ⚠️ Use only if you've adopted a Kubernetes-native control plane model

For organizations that manage all infrastructure as Kubernetes Custom Resources (the "GitOps-native" pattern), Crossplane is an option. DigitalOcean provides an **official Crossplane provider** that installs a `KubernetesCluster` CRD into a management cluster.

You define clusters as YAML manifests, and the Crossplane provider reconciles them by calling the DigitalOcean API.

This approach is powerful but adds significant complexity. Only use it if you're already using Crossplane for multi-cloud infrastructure orchestration.

## Production Deployment Best Practices

### High Availability Control Plane

For mission-critical workloads, enable the **highly available (HA) control plane** when creating the cluster.

**HA control plane benefits**:
- Multiple master components for redundancy
- Transparent failover during maintenance or failures
- Reduced downtime during Kubernetes version upgrades
- Production SLA on control plane uptime

**Cost**: $40/month (often waived for clusters with 3+ nodes)

**Important**: Once enabled, HA cannot be disabled. For dev/test clusters where cost matters more than uptime, a non-HA control plane is acceptable (DigitalOcean still manages control plane reliability, just without multi-master redundancy).

### Node Pools and Autoscaling

#### Multiple Node Pools for Workload Separation

Use **multiple node pools** to optimize for different workload profiles:

**Example: Production Multi-Pool Configuration**

```hcl
# Web/API workload pool
resource "digitalocean_kubernetes_node_pool" "web_pool" {
  cluster_id = digitalocean_kubernetes_cluster.prod.id
  name       = "web-pool"
  size       = "s-2vcpu-4gb"
  node_count = 3
  auto_scale = true
  min_nodes  = 3
  max_nodes  = 6

  labels = {
    role = "web"
  }
}

# Batch processing pool
resource "digitalocean_kubernetes_node_pool" "batch_pool" {
  cluster_id = digitalocean_kubernetes_cluster.prod.id
  name       = "batch-pool"
  size       = "s-4vcpu-8gb"
  node_count = 2
  auto_scale = true
  min_nodes  = 2
  max_nodes  = 4

  taint {
    key    = "workload"
    value  = "batch"
    effect = "NoSchedule"
  }

  labels = {
    role = "batch"
  }
}
```

#### Cluster Autoscaling

Enable **autoscaling** on node pools to handle variable load. DOKS integrates the standard Kubernetes Cluster Autoscaler.

**Configuration**:
- Set `auto_scale: true` on the node pool
- Define `min_nodes` and `max_nodes` boundaries
- **Note**: Scale-to-zero is not supported (min_nodes must be ≥1)

**Best practices**:
- Set min_nodes to ensure baseline capacity
- Set max_nodes to prevent runaway costs
- Use Pod Disruption Budgets to prevent disruption during scale-down
- Enable **auto-repair** (on by default) to automatically replace unhealthy nodes

### Networking and Security

#### Private Networking and VPC Isolation

Run clusters in **dedicated VPCs** to isolate network traffic. By default, DOKS nodes use private IPs within a VPC, preventing direct internet access to pod networks.

**Best practices**:
- Create a dedicated VPC per environment (dev, staging, prod)
- Use VPC peering to connect clusters if needed
- Disable public IPs on nodes entirely for maximum security (requires bastion host or VPN for access)

#### Control Plane Firewall

**Critical security setting**: Restrict Kubernetes API server access using the **control plane firewall**.

**Anti-Pattern**: Leaving the API server publicly accessible (`0.0.0.0/0`)

**Best Practice**: Whitelist only trusted IP addresses (office IPs, VPN endpoints, CI/CD runners)

**Example (Terraform)**:

```hcl
resource "digitalocean_kubernetes_cluster" "secure_prod" {
  # ... other config ...

  # Restrict API access to specific IPs
  firewall {
    inbound_rules = [
      {
        protocol         = "tcp"
        port_range       = "6443"
        source_addresses = ["203.0.113.5/32"]  # Corporate VPN
      }
    ]
  }
}
```

#### Network Policies

DOKS uses **Cilium CNI** by default, which provides eBPF-based networking with:
- Built-in support for Kubernetes **NetworkPolicy** resources
- **Hubble** for network observability (flow visualization, service maps)
- High-performance packet processing

**Production requirement**: Implement NetworkPolicy rules to restrict pod-to-pod communication on a least-privilege basis.

### Storage Best Practices

#### Block Storage Integration

DOKS includes a **CSI driver** that automatically provisions DigitalOcean Block Storage Volumes when you create PersistentVolumeClaims.

**Example PVC**:

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: postgres-data
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 100Gi
  storageClassName: do-block-storage
```

**Cost**: $0.10 per GB per month (plus $0.05/GB for snapshots)

**Best practices**:
- Right-size volumes (avoid over-allocation)
- Use application-level replication for critical data (volumes are zonal)
- Enable volume snapshots for backup
- Consider managed databases for stateful workloads instead of self-hosting

#### Container Registry Integration

Enable **DigitalOcean Container Registry (DOCR)** integration to pull private images without manual secret creation:

**Terraform Example**:

```hcl
resource "digitalocean_kubernetes_cluster" "cluster" {
  # ... other config ...
  registry_integration = true
}
```

This automatically creates the necessary `imagePullSecrets` in the cluster.

### Upgrade Strategy

#### Automatic Patch Upgrades

Enable **automatic patch upgrades** (`auto_upgrade: true`) to receive security updates within your selected Kubernetes minor version (e.g., 1.27.x).

**Configuration**:
- Upgrades occur during the configured **maintenance window**
- Use **surge upgrades** (enabled by default) to minimize downtime—adds temporary nodes during rollout to maintain capacity

**Best practices**:
- Enable auto-upgrades for production to receive security patches
- Set maintenance windows to low-traffic periods (e.g., Sunday 2:00 AM UTC)
- Test upgrades on staging clusters first

#### Manual Minor Version Upgrades

Minor version upgrades (e.g., 1.27 → 1.28) are **not automatic** and must be triggered manually via API, `doctl`, or IaC.

**Best practices**:
- Stay within N-2 versions of the latest release (DOKS typically supports the latest 3 versions)
- Plan upgrade windows and communicate to teams
- Use staging environments to validate compatibility before upgrading production

### Monitoring and Observability

#### Built-in Monitoring

DigitalOcean provides basic cluster metrics in the control panel (CPU, memory, disk usage per node).

For production, deploy a comprehensive monitoring stack:
- **Prometheus + Grafana**: Install via Helm or DigitalOcean 1-Click Apps
- **Prometheus Operator**: Full observability stack (Prometheus, Grafana, Alertmanager)
- **Hubble** (included with Cilium): Network flow observability
- **Log aggregation**: Use Loki, Elasticsearch, or external SaaS (Datadog, etc.)

#### Security Scanning

- Enable **RBAC** (on by default) and enforce least-privilege access
- Use **Pod Security Standards** to prevent privileged containers
- Scan container images (DOCR has optional security scanning, or use Trivy)
- Run **kube-bench** to audit cluster against CIS Kubernetes benchmarks

### Common Production Anti-Patterns

❌ **Exposing the API server publicly**: Always use the control plane firewall to restrict access  
❌ **Single-node clusters for production**: Run minimum 3 nodes for resilience  
❌ **Ignoring maintenance windows**: Set explicit windows to avoid upgrades during peak traffic  
❌ **No NetworkPolicy rules**: Default-allow networking is a security risk  
❌ **Resource starvation**: Always set CPU/memory requests and limits on pods  
❌ **`:latest` tags**: Pin image versions in production  
❌ **Mixing environments**: Never run dev, staging, and prod in the same cluster

## The Planton Choice

Planton's `DigitalOceanKubernetesCluster/v1` API is designed following the **80/20 principle**: expose the 20% of configuration that 80% of users need, enforce security by default, and make production-readiness the path of least resistance.

### Minimal, Opinionated API

The Planton API **mandates** the following fields:

| Field | Type | Rationale |
|-------|------|-----------|
| `cluster_name` | string (required) | Primary cluster identifier |
| `region` | string (required) | Fundamental deployment location (e.g., "nyc1", "sfo3") |
| `kubernetes_version` | string (required) | Explicit version pinning for stability |
| `node_pool.size` | string (required) | Node instance type (e.g., "s-2vcpu-4gb") |
| `node_pool.count` | uint32 (required, min: 1) | Initial node count |

### Optional But Critical Fields

| Field | Type | Default | Use Case |
|-------|------|---------|----------|
| `ha` | bool | false | Enable HA control plane for production |
| `auto_upgrade` | bool | false | Automatic patch version upgrades |
| `surge_upgrade` | bool | true | Maintain capacity during upgrades |
| `maintenance_window` | string | "any=00:00" | Scheduled maintenance time |
| `registry_integration` | bool | false | Enable DOCR integration |
| `vpc_uuid` | string | (default VPC) | Specify custom VPC |
| `control_plane_firewall_allowed_ips` | repeated string | [] | Restrict API access |
| `node_pool.auto_scale` | bool | false | Enable autoscaling |
| `node_pool.min_nodes` | uint32 | 0 | Autoscaler minimum |
| `node_pool.max_nodes` | uint32 | 0 | Autoscaler maximum |
| `tags` | repeated string | [] | Resource organization and cost allocation |

### Default Deployment Strategy

Planton uses **Pulumi** as the default IaC engine for DigitalOcean Kubernetes clusters, leveraging:

- The official `@pulumi/digitalocean` provider (complete resource coverage)
- Built-in encrypted secret management for API tokens and kubeconfig
- Stack-based multi-environment management
- Integration with Planton's unified resource lifecycle

The underlying Pulumi module is located at:

```
apis/project/planton/provider/digitalocean/digitaloceankubernetescluster/v1/iac/pulumi/
```

A reference Terraform implementation is also provided for users who prefer Terraform:

```
apis/project/planton/provider/digitalocean/digitaloceankubernetescluster/v1/iac/tf/
```

## Reference Configurations

### Dev Cluster (Minimal, Low-Cost)

**Goal**: Quick testing environment. Cost optimization is paramount.

**Configuration**:
- **Region**: nyc1 or sfo3 (nearest to developer)
- **Nodes**: 2
- **Instance Size**: `s-1vcpu-2gb` (~$12/month per node = ~$24/month total)
- **HA Control Plane**: false (saves $40/month)
- **Auto-Upgrade**: true (accept patches automatically)
- **Autoscaling**: false (fixed 2-node size)
- **Kubernetes Version**: latest stable

**Total Monthly Cost**: ~$24 (nodes only, control plane free)

**Use Case**: Local development, throwaway experiments, short-lived feature testing

### Staging Cluster (Production-Like)

**Goal**: Multi-node cluster mirroring production configuration for integration testing.

**Configuration**:
- **Region**: nyc1 or sfo3
- **Nodes**: 3 (minimum for HA simulation)
- **Instance Size**: `s-2vcpu-4gb` (~$20/month per node = ~$60/month total)
- **HA Control Plane**: false (acceptable downtime for staging)
- **Auto-Upgrade**: true
- **Autoscaling**: false (predictable capacity for testing)
- **Kubernetes Version**: Pinned to specific version (e.g., `1.27.5-do.0`)
- **Control Plane Firewall**: Optional (can restrict to office/VPN)

**Total Monthly Cost**: ~$60 (nodes only)

**Use Case**: Pre-production validation, integration testing, QA environments

### Production Cluster (HA, Secure, Scalable)

**Goal**: Robust cluster for production traffic with security hardening and autoscaling.

**Configuration**:
- **Region**: nyc1, sfo3, or multi-region strategy
- **Nodes**: 5+ (initial), autoscaling enabled
- **Instance Size**: `s-4vcpu-8gb` (~$43/month per node)
- **HA Control Plane**: **true** (critical for uptime)
- **Auto-Upgrade**: true (with maintenance window set)
- **Autoscaling**: true (min: 5, max: 10)
- **Kubernetes Version**: Pinned to validated version
- **Control Plane Firewall**: **Mandatory** (whitelist office/VPN IPs only)
- **Registry Integration**: true (DOCR)
- **Maintenance Window**: Sunday 02:00 UTC
- **VPC**: Dedicated production VPC
- **Node Pools**: Multiple pools for workload separation (web, batch, system)

**Total Monthly Cost (baseline)**:
- Control plane (HA): $40 (may be waived)
- Nodes: 5 × $43 = $215
- Load balancer: ~$10
- **Total**: ~$265/month (before autoscaling, block storage, or additional resources)

**Use Case**: Customer-facing production workloads, SaaS applications, critical services

## Cost Optimization Strategies

### Understanding the Pricing Model

**Control Plane**:
- Free (non-HA)
- $40/month (HA, often waived for clusters with 3+ nodes)

**Worker Nodes**:
- Standard Droplet pricing (per-second billing)
- Common sizes: s-1vcpu-2gb (~$12/mo), s-2vcpu-4gb (~$20/mo), s-4vcpu-8gb (~$43/mo)
- Includes bandwidth: ~1TB per node per month (pooled across cluster)

**Additional Resources**:
- Load Balancers: ~$10/month each
- Block Storage Volumes: $0.10/GB/month
- Snapshots: $0.05/GB/month
- Bandwidth overage: $0.01/GB (after included quota)

### Optimization Tactics

1. **Right-size nodes**: Monitor utilization and adjust instance types to avoid idle capacity
2. **Enable autoscaling**: Scale down during off-peak hours to reduce costs
3. **Consolidate load balancers**: Use a single Ingress controller instead of multiple LoadBalancer services
4. **Use managed databases**: Often more cost-effective than self-hosting stateful workloads
5. **Clean up unused resources**: Delete unused PersistentVolumeClaims and load balancers
6. **Automate dev cluster teardown**: Destroy and recreate dev clusters outside working hours
7. **Tag resources**: Use tags for cost attribution and tracking

**Example Cost Calculation (Production)**:
- 5 nodes × `s-4vcpu-8gb` @ $43/mo = $215
- HA control plane (waived) = $0
- 2 load balancers @ $10/mo = $20
- 500GB block storage @ $0.10/GB = $50
- **Total**: ~$285/month

Compare to AWS EKS: $73 (control plane) + 5 nodes @ ~$70/mo = $423/month minimum

## Integration Ecosystem

### DigitalOcean Marketplace

Pre-packaged applications installable during cluster creation (via `doctl --1-clicks` or Terraform):

- **Ingress**: Nginx Ingress, Traefik
- **Security**: cert-manager (Let's Encrypt TLS)
- **Observability**: Prometheus Operator, Grafana
- **CI/CD**: GitLab Runner, Argo CD

### CI/CD Integration

**GitHub Actions**: Use the DigitalOcean GitHub Action to configure `kubectl`:

```yaml
- uses: digitalocean/action-doctl@v2
  with:
    token: ${{ secrets.DIGITALOCEAN_ACCESS_TOKEN }}
- run: doctl kubernetes cluster kubeconfig save my-cluster
- run: kubectl get nodes
```

**GitLab CI**: Store the `DIGITALOCEAN_ACCESS_TOKEN` as a CI/CD variable and use `doctl` in jobs.

### Kubeconfig Management

The `kubeconfig` file is the authentication credential for `kubectl`.

**Manual (Local)**:
```bash
doctl kubernetes cluster kubeconfig save my-cluster
# Merges into ~/.kube/config
```

**IaC (Automation)**:
- The kubeconfig is an output of the IaC resource
- **Never write it to disk in CI/CD pipelines**—pass it in-memory to the Kubernetes provider

## Conclusion

DigitalOcean Kubernetes represents a compelling alternative to hyperscaler complexity: **production-grade Kubernetes without the enterprise learning curve**. By prioritizing developer experience, cost transparency, and sane defaults, DOKS delivers a managed Kubernetes service that provisions in minutes, costs significantly less than alternatives, and integrates seamlessly with modern IaC workflows.

The platform is mature, fully automatable via official Terraform and Pulumi providers, and production-ready for workloads that value **velocity, predictability, and cost efficiency** over deep cloud ecosystem lock-in.

**Planton's opinionated API design enforces security by default** (requiring network and firewall configuration for production), exposes the critical configuration surface (Kubernetes version, node pools, autoscaling, HA), and abstracts away the complexity of multi-cloud infrastructure management—making DigitalOcean Kubernetes accessible to teams of any size.

For organizations that don't need the full AWS/GCP/Azure ecosystem, DOKS provides a refreshingly simple path to production Kubernetes—backed by the reliability of a managed service and the flexibility of Infrastructure as Code.

---

**For comprehensive guides on specific implementation details, see:**

- [Node Pool Management and Autoscaling Patterns](./node-pool-guide.md) *(placeholder)*
- [Production Security Hardening for DOKS Clusters](./security-hardening.md) *(placeholder)*
- [Multi-Region HA Strategies with DigitalOcean](./multi-region-guide.md) *(placeholder)*
- [Cost Optimization and Resource Right-Sizing](./cost-optimization.md) *(placeholder)*


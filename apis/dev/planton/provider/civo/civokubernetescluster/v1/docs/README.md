# Civo Kubernetes Deployment Methods

## Introduction: The K3s Advantage

For years, managed Kubernetes meant waiting 10-15 minutes for a cluster to provision while cloud providers assembled complex control planes with distributed etcd clusters and multi-process architectures. Civo Cloud took a different approach: build a production-grade managed Kubernetes service on **K3s**, a lightweight, CNCF-certified Kubernetes distribution that consolidates the entire control plane into a single binary under 100MB.

The result? **Cluster creation in under 90 seconds**, a **free control plane**, and a developer-first experience that doesn't sacrifice production-readiness for simplicity.

This document explains the deployment methods available for Civo Kubernetes clusters, from manual provisioning to fully automated Infrastructure as Code (IaC) patterns, and how Planton exposes this powerful platform through a minimal, opinionated API.

## Why Civo Kubernetes Exists

### The K3s Foundation

Civo's managed Kubernetes service is built on **K3s**, a Rancher/SUSE project that reimagines Kubernetes for speed and resource efficiency. The architectural differences from standard Kubernetes (K8s) are fundamental:

| Feature | K3s (Civo's Base) | Standard K8s (GKE/EKS/AKS) |
|---------|------------------|---------------------------|
| **Distribution** | Single binary (<100MB) | Multiple separate control plane binaries |
| **Default Datastore** | Embedded SQLite3 | etcd (distributed consensus) |
| **Resource Footprint** | Very low (512MB RAM minimum) | High (2+ vCPU master typical) |
| **Cloud Integration** | Uses external, standard interfaces (CSI, CCM) | Historically used in-tree provider code |
| **Launch Time** | <90 seconds | 5-15+ minutes |

This isn't a fork or a toy distribution—K3s is **100% CNCF conformant** and production-proven. The speed gains come from bypassing the time-consuming bootstrap of a distributed etcd cluster by using an embedded SQLite3 datastore by default, and consolidating the control plane components (kube-apiserver, kube-controller-manager, kube-scheduler) into a single server process.

### Strategic Positioning

Civo doesn't compete with hyperscalers on ecosystem depth. Instead, it competes on:

1. **Speed**: Sub-90-second provisioning vs. 5-15+ minutes for GKE/EKS/AKS
2. **Simplicity**: Developer-focused, minimal configuration surface
3. **Cost**: Free control plane and transparent pricing with no egress fees

This makes Civo ideal for **developers, startups, CI/CD pipelines**, and **cost-sensitive production workloads** that value velocity and predictability over deep cloud ecosystem lock-in.

## The Deployment Spectrum

Civo provides a complete range of deployment methods, from manual click-ops to fully declarative, GitOps-driven automation. All programmatic methods consume the same underlying **Civo v2 REST API**.

### Level 0: Manual Provisioning (The Anti-Pattern)

**Method**: Civo web console  
**Verdict**: ❌ Suitable for learning only, not production

The Civo dashboard provides a simple wizard where you click through steps: select region, name the cluster, choose Kubernetes version, define node pools, select a network and firewall, and optionally add Marketplace applications.

This is the fastest way to *see* a Civo cluster, but it's not repeatable, not auditable, and creates configuration drift. **Never use manual provisioning for production infrastructure.**

### Level 1: Scripted Provisioning

**Method**: Civo CLI (`civo`)  
**Verdict**: ⚠️ Acceptable for scripts and CI/CD, but not ideal for complex infrastructure

The Civo CLI is an official, open-source Go tool that wraps the Civo API. It's authenticated with an API key and can create clusters imperatively:

```bash
civo kubernetes create my-cluster \
  --nodes=3 \
  --size=g4s.kube.medium \
  --region=LON1 \
  --wait
```

The CLI is excellent for:
- Quick experimentation and debugging
- CI/CD pipelines that need simple cluster creation
- Scripting one-off operations

However, it doesn't manage state, so you can't detect drift or safely update infrastructure.

### Level 2: Direct API Integration

**Method**: Civo v2 REST API or Go SDK (`civogo`)  
**Verdict**: ⚠️ Use only for custom platform building

For custom tooling or platform engineering, you can interface directly with Civo's REST API (`POST https://api.civo.com/v2/kubernetes/clusters`) or use the official **Go SDK** (`github.com/civo/civogo`). This SDK is the foundation of both the Civo CLI and the official Terraform provider, making it mature and production-grade.

Use this approach only if you're building a custom control plane or platform layer. For standard infrastructure management, use IaC tools instead.

### Level 3: Infrastructure as Code (The Production Solution)

**Methods**: Terraform, Pulumi, OpenTofu  
**Verdict**: ✅ **Recommended for all production deployments**

This is the gold standard for managing Civo Kubernetes infrastructure. IaC provides:
- **Declarative configuration**: Define desired state, not imperative steps
- **Version control**: Infrastructure lives in Git alongside application code
- **State management**: Detect drift and prevent conflicting changes
- **Reusability**: Modules/stacks for multi-environment deployments

#### Terraform

Civo maintains the **official Terraform provider** (`civo/civo`) in the HashiCorp Registry. It's mature, actively developed, includes a comprehensive acceptance test suite, and is referenced in numerous official tutorials.

**Example: Production Cluster**

```hcl
resource "civo_kubernetes_cluster" "prod_cluster" {
  name                = "planton-prod"
  region              = "LON1"
  network_id          = civo_network.prod.id
  firewall_id         = civo_firewall.prod_firewall.id
  kubernetes_version  = "1.29.0+k3s1"
  
  cni_plugin          = "cilium"  # CRITICAL: Required for NetworkPolicy
  
  applications = [
    "prometheus-operator",
    "cert-manager",
    "civo-cluster-autoscaler"
  ]
  
  pools {
    node_count = 5
    size       = "g4s.kube.large"
  }
}
```

State is managed via remote backends (S3, GCS, or Terraform Cloud). Secrets (API keys) are passed via environment variables (`TF_VAR_civo_token`).

#### Pulumi

Pulumi is also officially supported. The **`@pulumi/civo` provider** allows infrastructure definition in general-purpose programming languages (Python, Go, TypeScript, etc.).

**Important**: The Pulumi provider is a **bridged provider**, meaning it's programmatically generated from the upstream `civo/terraform-provider-civo`. This ensures 100% resource parity with Terraform, though it may introduce a slight delay in adopting new Civo features (requires an upstream Terraform provider update first).

**Advantages over Terraform**:
- **State management**: Defaults to managed Pulumi Cloud with concurrency locking out-of-the-box
- **Secret management**: Built-in encrypted configuration (`pulumi config set civo:token --secret <value>`)
- **Multi-environment**: Uses Stacks (e.g., `my-app-staging`, `my-app-prod`) with per-stack configuration files

**Example: Staging Cluster (Python)**

```python
cluster = civo.KubernetesCluster("staging-cluster",
    name="planton-staging",
    region="LON1",
    network_id=network.id,
    firewall_id=firewall.id,
    kubernetes_version="1.29.0+k3s1",
    cni_plugin="cilium",
    applications=["prometheus-operator", "cert-manager"],
    pools=[{
        "node_count": 3,
        "size": "g4s.kube.medium"
    }]
)
```

#### OpenTofu

As a compatible fork of Terraform 1.x, OpenTofu works seamlessly with the `civo/civo` provider. No code changes are required—simply use `tofu` instead of `terraform` commands.

**Comparison: Terraform vs. Pulumi for Civo**

| Feature | Terraform (civo/civo) | Pulumi (@pulumi/civo) |
|---------|----------------------|----------------------|
| **Provider Origin** | Official, Native (Go) | Bridged from Terraform provider |
| **Resource Parity** | N/A (source of truth) | 100% parity with Terraform |
| **Feature Lag** | None | Minor (waits on TF provider + bridge build) |
| **State Management** | Manual (local) or Remote (S3, TFC) | Pulumi Cloud (default) or Remote (S3, etc.) |
| **Secret Management** | Env vars or external Vault | Env vars or built-in encrypted config |
| **Multi-Environment** | Workspaces / Directory layouts | Stacks (named instances) |
| **Code Reuse** | Modules (HCL) | Functions, classes, packages (Go, TS, Py, etc.) |

**Verdict**: Both are production-ready. Choose Terraform for maximum maturity and ecosystem compatibility. Choose Pulumi for type-safe, programmatic IaC with superior secret handling and state management defaults.

### Level 4: Kubernetes-Native Provisioning

**Method**: Crossplane (`provider-civo`)  
**Verdict**: ⚠️ Use only if you've adopted a Kubernetes-native control plane model

For organizations that manage all infrastructure as Kubernetes Custom Resources (the "GitOps-native" pattern), Crossplane is an option. The **`crossplane-contrib/provider-civo`** is a community-supported provider that installs a `CivoKubernetesCluster` CRD into a management cluster.

You define clusters as YAML manifests, and the Crossplane provider reconciles them by calling the Civo API. The Civo API key is stored as a Kubernetes Secret, and the resulting kubeconfig is also written to a Secret.

This approach is powerful but adds significant complexity. Only use it if you're already using Crossplane for multi-cloud infrastructure orchestration.

## Production Deployment Best Practices

### The Critical CNI Decision

**This is the single most important production configuration decision.**

Civo allows you to select the Container Network Interface (CNI) plugin at cluster creation:

- **Flannel (default)**: Simple, lightweight, robust. **Does NOT support Kubernetes NetworkPolicy resources.**
- **Cilium**: eBPF-based, high-performance, security-focused. **Required for NetworkPolicy support.**

If you need network segmentation—a standard production security practice—**you must select `cilium` as the CNI plugin at creation time**. This cannot be easily changed post-provisioning.

**Example: Enforcing NetworkPolicy**

```hcl
cni_plugin = "cilium"  # Mandatory for network segmentation
```

### Security: Networks and Firewalls

Civo clusters are deployed into **Civo Networks** (similar to VPCs) and protected by **Civo Firewalls** (L4 stateful packet filters).

**Anti-Pattern**: Leaving the default "all-open" firewall attached or exposing the Kubernetes API server (port 6443) to the public internet (`0.0.0.0/0`).

**Best Practice**: Create a dedicated firewall for each environment, allowing only required ingress ports (80, 443) and restricting API server access to trusted IP addresses.

### High Availability

Civo markets "high availability as standard," but this refers to the **managed service promise** to ensure control plane resilience, not a multi-master, distributed control plane architecture (as in GKE/EKS).

**Application-level HA is your responsibility**: Run multiple worker nodes (minimum 3 for production) and design applications for redundancy.

### Autoscaling

Node pool autoscaling is **not enabled by default**. To enable it, install the **"Civo cluster autoscaler"** from the Marketplace. This is Civo's implementation of the standard Kubernetes `cluster-autoscaler`, which watches for pending pods and scales node pools within a defined range (e.g., 1 to 10 nodes).

### Storage and Load Balancing

Civo clusters are pre-configured with:

- **Civo CSI Driver**: Automatically provisions Civo Block Storage Volumes when you create a `PersistentVolumeClaim` using the default `civo-volume` StorageClass.
- **Civo Cloud Controller Manager (CCM)**: Automatically provisions Civo Load Balancers when you create a Service of `type: LoadBalancer`.

Both integrations work out-of-the-box with zero configuration.

### Observability

For production-grade monitoring, install the **`prometheus-operator`** from the Civo Marketplace. This deploys a full Prometheus, Grafana, and Alertmanager stack. For log aggregation, use **Grafana Loki**, also available via the Marketplace.

### Common Production Anti-Patterns

❌ **Insecure networking**: Default all-open firewall or public API access  
❌ **Assuming NetworkPolicy support**: Using the default Flannel CNI and expecting network segmentation  
❌ **Resource starvation**: Deploying pods without CPU/memory requests and limits  
❌ **`:latest` tags**: Using the `:latest` image tag in production deployments  
❌ **Environment mixing**: Running dev, staging, and prod in the same cluster or namespace  
❌ **Plain-text secrets**: Storing credentials in ConfigMaps or environment variables  

## The Planton Choice

Planton's `CivoKubernetesCluster/v1` API is designed following the **80/20 principle**: expose the 20% of configuration that 80% of users need, enforce security by default, and make production-readiness the path of least resistance.

### Minimal, Opinionated API

The Planton API **mandates** the following fields:

| Field | Type | Rationale |
|-------|------|-----------|
| `cluster_name` | string (required) | Primary cluster identifier |
| `region` | enum (required) | Fundamental deployment location |
| `kubernetes_version` | string (required) | Explicit version pinning for stability |
| `network` | reference (required) | Enforces network isolation (prevents default VPC anti-pattern) |
| `default_node_pool.size` | string (required) | Essential node pool configuration |
| `default_node_pool.node_count` | uint32 (required, min: 1) | Essential node pool configuration |

### Optional But Critical Fields

| Field | Type | Default | Use Case |
|-------|------|---------|----------|
| `highly_available` | bool | false | Enable HA control plane (when supported by Civo) |
| `auto_upgrade` | bool | false | Automatic patch version upgrades |
| `disable_surge_upgrade` | bool | false | Control upgrade strategy |
| `tags` | repeated string | [] | Resource organization and cost allocation |

**Note**: The CNI plugin selection is not exposed in the current API version. Users must configure this directly in the underlying IaC module if Cilium is required. Future API iterations may expose this as an optional field with a clear warning about the NetworkPolicy implications.

### Default Deployment Strategy

Planton uses **Pulumi** as the default IaC engine for Civo Kubernetes clusters, leveraging:

- The official `@pulumi/civo` bridged provider (100% resource parity with Terraform)
- Built-in encrypted secret management for API keys and kubeconfig
- Stack-based multi-environment management
- Integration with Planton's unified resource lifecycle

The underlying Pulumi module is located at:

```
apis/project/planton/provider/civo/civokubernetescluster/v1/iac/pulumi/
```

A reference Terraform implementation is also provided for users who prefer Terraform:

```
apis/project/planton/provider/civo/civokubernetescluster/v1/iac/tf/
```

## Reference Configurations

### Dev Cluster (Minimal, Low-Cost)

**Goal**: Single-node cluster for quick testing. Cost optimization is paramount.

**Configuration**:
- **Region**: NYC1 or LON1 (nearest to developer)
- **Nodes**: 1
- **Instance Size**: `g4s.kube.small` (1 core, 2GB RAM, ~$10.86/month)
- **CNI**: Flannel (default, acceptable for dev)
- **Kubernetes Version**: Latest stable
- **Marketplace Apps**: None

**Use Case**: Local development, throwaway experiments, CI/CD test clusters

### Staging Cluster (Production-Like)

**Goal**: Multi-node cluster mirroring production configuration with monitoring.

**Configuration**:
- **Region**: LON1 or FRA1
- **Nodes**: 3 (minimum for simulating HA patterns)
- **Instance Size**: `g4s.kube.medium` (2 cores, 4GB RAM, ~$21.73/month per node)
- **CNI**: **Cilium** (required for NetworkPolicy testing)
- **Kubernetes Version**: Pinned to specific version (e.g., `1.29.0+k3s1`)
- **Marketplace Apps**: `prometheus-operator`, `cert-manager`

**Use Case**: Pre-production validation, integration testing, QA environments

### Production Cluster (HA Applications, Secure)

**Goal**: Robust cluster for production traffic with security hardening and observability.

**Configuration**:
- **Region**: LON1, FRA1, or multi-region strategy
- **Nodes**: 5+ (allows graceful node rotation during upgrades)
- **Instance Size**: `g4s.kube.large` (4 cores, 8GB RAM, ~$43.45/month per node)
- **CNI**: **Cilium** (mandatory for network segmentation)
- **Kubernetes Version**: Pinned to specific version
- **Marketplace Apps**: `prometheus-operator`, `cert-manager`, `civo-cluster-autoscaler`
- **Firewall**: Custom firewall with strict ingress rules (only ports 80/443 public, API server restricted to office/VPN IPs)
- **Auto-Upgrade**: Disabled (manual upgrade windows preferred)

**Use Case**: Customer-facing production workloads

## Integration Ecosystem

### Civo Marketplace

The Civo Marketplace is a catalog of 80+ pre-packaged applications that can be installed declaratively during cluster creation via the `applications` parameter. Key applications include:

- **Ingress**: Traefik (installed by default), NGINX, Emissary-ingress
- **Security**: `cert-manager` (for Let's Encrypt TLS), HashiCorp Vault
- **Observability**: `prometheus-operator` (full Prometheus/Grafana/Alertmanager stack)
- **CI/CD & GitOps**: GitLab, Flux

To **remove** the default Traefik installation, use:

```hcl
applications = ["-Traefik"]
```

### CI/CD Integration

**GitHub Actions**: Use the official `civo/action-civo` GitHub Action to install the Civo CLI in runners. Store the `CIVO_TOKEN` as a GitHub Secret and run `civo kubernetes config $CLUSTER_NAME --save` to make the kubeconfig available to subsequent `kubectl` or `helm` steps.

**GitLab CI**: Civo offers deeper integration with GitLab. The official template repository (`gitlab.com/civocloud/gitlab-terraform-civo`) uses GitLab CI and Terraform to provision clusters and automatically installs the **GitLab Agent for Kubernetes**, creating a persistent, pull-based GitOps connection.

### Kubeconfig Management

The `kubeconfig` file is the authentication credential for `kubectl`. Management patterns:

- **Manual (Local)**: Download from Civo UI or use `civo kubernetes config --save` (merges into `~/.kube/config`)
- **IaC (Automation)**: The kubeconfig is an output of the IaC resource. **Never write it to disk in CI/CD pipelines**—pass it in-memory to the Kubernetes provider or tool.

## Conclusion

Civo Kubernetes represents a paradigm shift: **production-grade Kubernetes doesn't require hyperscaler complexity**. By building on K3s and prioritizing developer experience, Civo delivers a managed Kubernetes service that provisions in under 90 seconds, costs significantly less than alternatives, and integrates seamlessly with modern IaC workflows.

The platform is mature, automatable via official Terraform and Pulumi providers, and production-ready for workloads that value velocity, predictability, and cost efficiency.

**Planton's opinionated API design enforces security by default** (requiring network and firewall configuration upfront), exposes the critical configuration surface (Kubernetes version, node pools, tags), and abstracts away the complexity of multi-cloud infrastructure management—making Civo Kubernetes accessible to teams of any size.

For comprehensive guides on specific implementation details, see:

- [CNI Selection and Network Policy Configuration](./cni-guide.md) *(placeholder)*
- [Production Security Hardening for Civo Clusters](./security-hardening.md) *(placeholder)*
- [Multi-Cluster GitOps Patterns with Flux and Civo](./gitops-guide.md) *(placeholder)*


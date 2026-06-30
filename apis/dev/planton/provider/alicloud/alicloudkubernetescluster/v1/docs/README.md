# ACK Managed Kubernetes Cluster Deployment: From Console Wizards to Declarative Control Planes

## Introduction

Alibaba Cloud Container Service for Kubernetes (ACK) is the managed Kubernetes offering at the core of most production container workloads on Alibaba Cloud. An ACK managed cluster provides a fully managed control plane — etcd, kube-apiserver, kube-controller-manager, and kube-scheduler — while delegating worker node lifecycle to separately managed node pools. This separation of control plane from data plane is a defining characteristic of managed Kubernetes and allows independent scaling, patching, and failure isolation.

Despite the "managed" label, deploying an ACK cluster correctly involves dozens of interrelated decisions: VPC and VSwitch topology, CNI plugin selection (Flannel overlay vs. Terway ENI-based), service and pod CIDR allocation, addon installation, RRSA (RAM Roles for Service Accounts) configuration, control plane logging, and API server exposure settings. A single misconfiguration — overlapping CIDRs, missing NAT gateways, or disabled RRSA — can cascade into cluster-wide networking failures or security gaps that are difficult to diagnose after the fact.

This document traces the evolution of ACK cluster deployment methodologies, examines the critical architectural decisions, and explains how Planton abstracts the complexity into a declarative API that enforces best practices by default.

## The ACK Deployment Landscape

ACK cluster management spans a spectrum from manual console wizards to continuously reconciled control planes.

### Level 0: Manual Provisioning via Alibaba Cloud Console (The Anti-Pattern)

The ACK console provides a multi-step wizard for cluster creation. While suitable for learning, the wizard is **dangerous for production** because it allows silent misconfigurations:

**Common Mistakes**:

1. **Single-AZ VSwitch Selection**: The wizard allows selecting VSwitches from the same Availability Zone. The resulting cluster has no AZ-level fault tolerance — a single zone outage takes the entire cluster offline. The API allows 1–5 VSwitches, but production clusters need at least two in distinct zones.

2. **Flannel-Terway Confusion**: Choosing the wrong CNI plugin at creation time is **irreversible**. Flannel uses overlay networking (pod CIDRs detached from VPC), while Terway attaches elastic network interfaces (ENIs) directly from VPC VSwitches, providing VPC-native pod IPs. Mixing these or selecting the wrong one for the workload creates networking issues that require cluster recreation to fix.

3. **CIDR Overlap**: The wizard does not validate that `pod_cidr`, `service_cidr`, and VPC CIDR ranges do not overlap. Overlapping CIDRs cause silent routing failures where traffic destined for pods is misrouted to VPC hosts, or Kubernetes services shadow VPC IP ranges.

4. **Disabled RRSA**: The console defaults to RRSA off. Without RRSA, pods that need access to Alibaba Cloud APIs (SLS, OSS, ACR, etc.) must use static AccessKey credentials, which are a security liability: they don't rotate automatically, they grant access to any pod on any node, and they appear in environment variables visible to anyone with `kubectl exec` access.

5. **Missing NAT Gateway**: Clusters in private VSwitches need a NAT gateway for outbound internet access (pulling container images, accessing external APIs). The console's `new_nat_gateway` option creates one automatically, but users who bring their own NAT gateway must remember to set this to false — otherwise ACK creates a duplicate that wastes money and introduces routing ambiguity.

**Console Fragility**: The ACK console orchestrates multiple API calls behind the scenes. Partial failures (e.g., cluster creation succeeds but addon installation fails) leave the cluster in an inconsistent state that requires manual cleanup through the API.

**Verdict**: Acceptable for learning and experimentation. **Unacceptable for production** environments requiring reproducibility, auditability, or multi-environment consistency.

### Level 1: Scripted Provisioning with Alibaba Cloud CLI

The `aliyun cs` CLI provides imperative commands for cluster management:

```bash
aliyun cs POST /api/v2/clusters --body '{
  "name": "my-cluster",
  "cluster_type": "ManagedKubernetes",
  "kubernetes_version": "1.30.1-aliyun.1",
  "region_id": "cn-hangzhou",
  "vpcid": "vpc-xxx",
  "vswitch_ids": ["vsw-aaa", "vsw-bbb"],
  "service_cidr": "172.21.0.0/20",
  "addons": [{"name": "flannel"}, {"name": "csi-plugin"}],
  ...
}'
```

**The Complexity Problem**: The JSON payload for `CreateCluster` is deeply nested and has dozens of fields. Getting it right requires reading the API reference carefully. Unlike the console, the CLI doesn't provide defaults for most fields — you must specify everything explicitly.

**The State Problem**: The CLI creates resources but does not track state. There is no built-in mechanism to determine what has changed between the current cluster and a desired configuration. Updates require manually calculating the diff and calling the appropriate update APIs.

**Key Advantage**: Scriptable and reproducible. The JSON payload can be version-controlled, which is a significant improvement over console-based workflows.

**Verdict**: Suitable for simple automated tasks or CI/CD scripts. Not ideal for managing complex, long-lived clusters that require ongoing configuration changes.

### Level 2: Infrastructure as Code (Terraform, Pulumi)

IaC tools are the **modern standard** for ACK cluster management. They provide:
- **Declarative state**: Define the desired cluster configuration; the tool calculates and applies the diff
- **State tracking**: Know exactly what was provisioned and detect drift
- **Dependency management**: Automatically handle the creation order of VPC, VSwitches, NAT gateway, and cluster

#### Terraform/OpenTofu

The `alicloud_cs_managed_kubernetes` resource wraps the ACK managed cluster API:

```hcl
resource "alicloud_cs_managed_kubernetes" "cluster" {
  name                = "production-cluster"
  cluster_spec        = "ack.pro.small"
  vswitch_ids         = [alicloud_vswitch.a.id, alicloud_vswitch.b.id]
  service_cidr        = "172.21.0.0/20"
  new_nat_gateway     = false
  slb_internet_enabled = true
  enable_rrsa         = true

  addons {
    name = "terway-eniip"
  }
  addons {
    name = "csi-plugin"
  }
  addons {
    name = "csi-provisioner"
  }
}
```

**Resource Scope**: Unlike AWS EKS (which splits cluster, node groups, and addons into separate resources), the Alibaba Cloud provider bundles addon installation and several cluster settings into a single resource. This simplifies the basic case but means addon changes trigger a cluster update.

**Immutability Constraints**: Several fields are immutable after creation: `service_cidr`, `pod_cidr`, `pod_vswitch_ids`, `cluster_domain`, `node_cidr_mask`, and `encryption_provider_key`. Changing any of these requires cluster recreation — a critical fact that IaC tools surface clearly through their plan output.

**State Management**: Terraform's remote state (OSS backend with TableStore locking) is the standard for team-based workflows on Alibaba Cloud.

#### Pulumi

Pulumi's `cs.ManagedKubernetes` resource provides the same coverage in Go, TypeScript, Python, or C#:

```go
cluster, err := cs.NewManagedKubernetes(ctx, "cluster", &cs.ManagedKubernetesArgs{
    Name:               pulumi.String("production-cluster"),
    ClusterSpec:        pulumi.String("ack.pro.small"),
    VswitchIds:         pulumi.StringArray{vswitchA.ID(), vswitchB.ID()},
    ServiceCidr:        pulumi.String("172.21.0.0/20"),
    NewNatGateway:      pulumi.Bool(false),
    SlbInternetEnabled: pulumi.Bool(true),
    EnableRrsa:         pulumi.Bool(true),
    Addons: cs.ManagedKubernetesAddonArray{
        &cs.ManagedKubernetesAddonArgs{Name: pulumi.String("terway-eniip")},
        &cs.ManagedKubernetesAddonArgs{Name: pulumi.String("csi-plugin")},
        &cs.ManagedKubernetesAddonArgs{Name: pulumi.String("csi-provisioner")},
    },
})
```

**Key Differentiator**: Pulumi's first-class Go support maps naturally to Planton's protobuf-based stack input model, where the manifest YAML is deserialized into Go structs and passed directly to the Pulumi program.

### Level 3: Control Planes and Gitops

The most advanced deployment model uses long-running controllers that continuously reconcile desired state:

**Crossplane on Alibaba Cloud**: Crossplane's Alibaba Cloud provider can manage ACK clusters as Kubernetes Custom Resources, enabling GitOps workflows where cluster configuration lives in a Git repository and is continuously reconciled by a controller running inside an existing cluster.

**ACK Fleet Management**: Alibaba Cloud's own fleet management service provides centralized control over multiple ACK clusters across regions. This is operationally convenient but vendor-locked.

**Planton Context**: Planton's protobuf-defined APIs sit at this level — providing a specification layer that can be reconciled by any controller. The Pulumi and Terraform modules are the execution engines, but the API specification is the durable artifact.

## Production-Grade ACK Architecture

A production ACK cluster is far more than a single API call. It's the correct configuration of networking, security, observability, and lifecycle management.

### Networking: The Flannel vs. Terway Decision

This is the **most consequential architectural decision** for an ACK cluster, and it is **immutable after creation**.

**Flannel (Overlay Networking)**:
- Pods get IPs from a cluster-internal CIDR (`pod_cidr`) that is **not routable** in the VPC
- Pod-to-pod traffic is encapsulated in VXLAN tunnels
- Requires `pod_cidr` in the cluster spec (e.g., `172.20.0.0/16`)
- **Advantages**: Simpler CIDR planning (pod CIDR is independent of VPC), no ENI quotas to worry about
- **Disadvantages**: Pods cannot be directly addressed from VPC resources (databases, other services); adds encapsulation overhead
- **Best for**: Development clusters, clusters with simple networking requirements, environments where pod-to-VPC direct communication is not needed

**Terway (ENI-Based Networking)**:
- Pods get IPs directly from VPC VSwitches via elastic network interfaces
- Pod IPs are **VPC-routable** — other VPC resources can communicate with pods directly
- Requires `pod_vswitch_ids` (dedicated VSwitches for pod ENIs, separate from node VSwitches)
- **Advantages**: VPC-native pod IPs enable direct communication with RDS, SLB, and other VPC services; no encapsulation overhead; network policy enforcement at the VPC level
- **Disadvantages**: Consumes VPC IP addresses (requires careful CIDR planning); subject to ENI quota limits per ECS instance type; pod VSwitches need sufficiently large CIDR blocks
- **Best for**: Production clusters that need VPC-level connectivity, security group-based pod isolation, or low-latency communication with VPC services

**The CIDR Planning Trap**: With Terway, pod VSwitches must have enough IP addresses for all pods across all nodes. A `/24` VSwitch provides ~250 IPs, which can be exhausted quickly on a cluster with many pods. Production deployments typically use `/20` or `/18` pod VSwitches.

**Planton Decision**: The spec supports both modes through mutually exclusive fields (`pod_cidr` for Flannel, `pod_vswitch_ids` for Terway). The addon list determines which CNI is active.

### VSwitch Topology and Multi-AZ

**Multi-AZ Requirement**: Production clusters must span at least two Availability Zones. The `vswitch_ids` field accepts 1–5 VSwitches, but a single VSwitch means a single-AZ cluster with no fault tolerance.

**Node vs. Pod VSwitches (Terway)**: When using Terway, two sets of VSwitches are needed:
1. **Node VSwitches** (`vswitch_ids`): Where ECS worker nodes are placed. Typically `/24` per AZ.
2. **Pod VSwitches** (`pod_vswitch_ids`): Where pod ENIs allocate IPs. Typically `/20` or larger per AZ to accommodate pod density.

These should be in the same AZs but with separate CIDR ranges.

### Security: RRSA, Encryption, and Network Isolation

**RRSA (RAM Roles for Service Accounts)**: The most important security feature for production clusters. RRSA enables pod-level IAM by federating Kubernetes service accounts with Alibaba Cloud RAM roles via OIDC:

1. ACK creates an OIDC provider (exposed via `rrsa_oidc_issuer_url` output)
2. RAM roles are configured with trust policies that match specific Kubernetes service accounts
3. Pods with matching service accounts can assume RAM roles without static credentials

**WARNING**: Once enabled, RRSA cannot be disabled. This is by design — disabling it would break all workloads relying on pod-level IAM. Enable it from the start.

**Secrets Encryption**: The `encryption_provider_key` field enables KMS-based encryption of Kubernetes Secrets stored in etcd. Without this, Secrets are stored base64-encoded (not encrypted) in etcd. This field is immutable — you cannot add encryption after cluster creation.

**API Server Exposure**: The `slb_internet_enabled` flag controls whether the Kubernetes API server is accessible from the public internet. Setting this to `false` restricts access to the VPC-internal endpoint only, which is more secure but requires VPN or bastion access for `kubectl`.

**Enterprise Security Groups**: The `is_enterprise_security_group` flag creates advanced security groups that support up to 65,536 rules and 100,000 ENIs, compared to standard security groups (100 rules, 2,000 ENIs). Required for large clusters with Terway networking.

### Control Plane Logging and Audit

**Control Plane Logs**: ACK can send control plane component logs (apiserver, kcm, scheduler, ccm, coreDNS) to Alibaba Cloud Log Service (SLS). These are essential for debugging control plane issues but are **not enabled by default**.

**Audit Logging**: Records every API request to the Kubernetes API server. Critical for security compliance (who did what, when) and troubleshooting (which controller is generating excessive API calls).

**Log Retention**: The `control_plane_log_ttl` field controls how many days logs are retained in SLS. Default is 30 days; production environments typically use 90 or more.

**SLS Project Reference**: The `control_plane_log_project` field can reference an existing AliCloudLogProject component via foreign key, enabling centralized log management across clusters.

### Addons: The Creation-Time Constraint

ACK addons are installed during cluster creation and **cannot be changed through the cluster resource after creation**. Post-creation addon management requires the `alicloud_cs_kubernetes_addon` Terraform resource or equivalent.

**Essential Addons**:
- **Network CNI** (choose one): `flannel` or `terway-eniip`
- **Storage**: `csi-plugin` + `csi-provisioner` (required for persistent volumes)
- **DNS**: `managed-coredns` (usually installed by default)

**Recommended Addons**:
- **Logging**: `logtail-ds` (DaemonSet for log collection, configurable via JSON)
- **Monitoring**: `arms-prometheus` (Prometheus-compatible monitoring)
- **Autoscaling**: `metrics-server` (required for HPA)
- **Health**: `ack-node-problem-detector` (node health monitoring)
- **Ingress**: `nginx-ingress-controller` or `alb-ingress-controller`

### Cluster Specification Tiers

**ack.standard** (Free): Basic managed control plane with standard SLA. Suitable for development, testing, and non-critical workloads.

**ack.pro.small** (Paid): Professional managed cluster with:
- Enhanced SLA (99.95% API server availability)
- Managed node pools with auto-repair and auto-upgrade
- Topology-aware scheduling
- Cluster auditing enhancements

**Upgrade Path**: Supports in-place upgrade from `ack.standard` to `ack.pro.small`. Downgrade is not supported.

### Maintenance Windows and Auto-Upgrade

**Maintenance Windows**: Define when ACK can apply updates and patches. Without a maintenance window, ACK may apply critical patches at any time. Production clusters should configure a window during low-traffic periods.

**Auto-Upgrade**: When configured alongside a maintenance window, ACK automatically upgrades the cluster Kubernetes version. Three channels control adoption speed:
- `patch`: Only patch upgrades (1.28.3 → 1.28.5)
- `stable`: Minor upgrades after stabilization (1.28 → 1.30 after proven stable)
- `rapid`: Minor upgrades as soon as available

## Production Best Practices and Anti-Patterns

| Category | Best Practice | Common Anti-Pattern | Impact |
|----------|--------------|---------------------|--------|
| **Availability** | Deploy across **at least 2 AZs** with VSwitches in distinct zones | Single-AZ VSwitch selection | Complete cluster outage on AZ failure |
| **Networking** | Choose CNI at design time; use Terway for VPC-native connectivity | Choosing Flannel then needing VPC-routable pod IPs | Cluster recreation required (immutable) |
| **CIDR Planning** | Plan non-overlapping CIDRs for VPC, pods, and services before creation | Overlapping pod_cidr with VPC CIDR | Silent routing failures, unreachable pods |
| **Security (IAM)** | Enable **RRSA from day one** for pod-level IAM | Using static AccessKey credentials in pods | Credential leakage, no rotation, no audit trail |
| **Security (Secrets)** | Enable **KMS encryption** at creation time | Leaving Secrets unencrypted in etcd | Base64 is encoding, not encryption |
| **API Server** | Set `slb_internet_enabled: false` for internal-only clusters | Exposing API server to internet unnecessarily | Attack surface expansion |
| **Logging** | Enable **control plane + audit logging** to SLS | Skipping logging configuration | Blind to control plane issues and security events |
| **Lifecycle** | Configure **maintenance windows** for controlled patching | No maintenance window (patches applied anytime) | Unexpected disruptions during peak traffic |
| **NAT Gateway** | Set `new_nat_gateway: false` when managing your own NAT | Letting ACK auto-create a duplicate NAT gateway | Wasted cost, routing ambiguity |
| **Addons** | Install all required addons at creation time | Adding critical addons post-creation | Requires separate addon management resources |

## Advanced Features: The 20% Use Case

### Custom SANs for API Server

The `custom_san` field adds additional Subject Alternative Names to the API server TLS certificate. This is needed when:
- Accessing the API server through a custom domain (e.g., `api.example.com`)
- Routing through a private IP that isn't automatically included in the certificate

### Cluster Domain Customization

The `cluster_domain` field (default: `cluster.local`) controls the DNS domain for Kubernetes service discovery. Changing this is rare but necessary in multi-cluster environments where service DNS names must be unique across clusters.

### Resource Group Isolation

The `resource_group_id` field places the cluster in a specific Alibaba Cloud resource group for cost allocation, access control, and organizational grouping. This is an optional enterprise governance feature.

## What Planton Supports

Planton provides a declarative API for ACK managed clusters that balances the 80% production use case with extensibility for advanced scenarios.

### Design Philosophy: 80/20 API Structure

The current API (`spec.proto`) covers the production-critical settings:

**Core Fields (80% Case)**:
- `region`: Cluster region
- `vswitch_ids`: Multi-AZ VSwitch placement (validated: 1–5 items)
- `service_cidr`: Kubernetes service CIDR (required)
- `addons`: Cluster addons installed at creation time

**Networking (Flannel or Terway)**:
- `pod_cidr`: For Flannel CNI
- `pod_vswitch_ids`: For Terway CNI
- `proxy_mode`: kube-proxy mode (default: `ipvs`)
- `node_cidr_mask`: Per-node pod CIDR mask (default: 24)

**Security**:
- `enable_rrsa`: Pod-level IAM via OIDC federation
- `encryption_provider_key`: KMS encryption for Secrets
- `security_group_id` / `is_enterprise_security_group`: Network security
- `deletion_protection`: Prevent accidental deletion

**Observability**:
- `logging.control_plane_log_project`: SLS project for control plane logs
- `logging.control_plane_log_components`: Which components to log
- `logging.audit_log_enabled`: Kubernetes audit logging

**Lifecycle**:
- `maintenance_window`: Controlled patching schedule
- `auto_upgrade`: Automatic version upgrades with channel selection

### Foreign Key References

Planton's `StringValueOrRef` pattern enables declarative cross-resource dependencies:
- `vswitch_ids` → references `AliCloudVswitch.status.outputs.vswitch_id`
- `pod_vswitch_ids` → references `AliCloudVswitch.status.outputs.vswitch_id`
- `security_group_id` → references `AliCloudSecurityGroup.status.outputs.security_group_id`
- `encryption_provider_key` → references `AliCloudKmsKey.status.outputs.key_id`
- `logging.control_plane_log_project` → references `AliCloudLogProject.status.outputs.project_name`

### What's Excluded from v1 (Future Enhancements)

The following features are intentionally excluded to maintain API simplicity, following the 80/20 principle:

- **kubeconfig output**: Sensitive credential that should be retrieved through `aliyun cs` CLI or the control plane, not stored in stack outputs
- **Worker node configuration**: Managed through the separate AliCloudKubernetesNodePool component
- **Post-creation addon management**: Requires `alicloud_cs_kubernetes_addon` (separate resource, separate lifecycle)
- **Cluster autoscaler configuration**: Installed as an addon; scaling policies belong in node pool configuration
- **ACK Serverless (ASK)**: A different cluster type with its own resource and API surface

### Implementation Landscape

**Pulumi Module**: A single `cs.ManagedKubernetes` resource orchestrated through `module/main.go`. The `locals.go` file handles default resolution (cluster spec, proxy mode, node CIDR mask, NAT gateway, SLB, RRSA, deletion protection, log TTL). The `outputs.go` file exports 11 output constants.

**Terraform Module**: A single `alicloud_cs_managed_kubernetes` resource in `main.tf` with dynamic blocks for addons, maintenance windows, audit log config, and operation policy (auto-upgrade). The `variables.tf` schema mirrors the proto with HCL-native type validation.

**Structural Simplicity**: Both modules wrap a single provider resource. The complexity is in the API design (field organization, validation rules, defaults) rather than the IaC orchestration.

## Conclusion

ACK managed Kubernetes cluster deployment represents a maturity journey from manual console wizards to declarative, version-controlled infrastructure specifications.

The research makes clear that:
- **Console provisioning is an anti-pattern** that allows silent, consequential misconfigurations
- **CNI selection (Flannel vs. Terway) is the most impactful decision** and is immutable after creation
- **RRSA and KMS encryption must be enabled at creation time** — retrofitting them is either impossible or disruptive
- **IaC is the minimum acceptable approach** for production clusters
- **Control plane APIs like Planton represent the future**, providing a durable specification layer above the IaC execution engine

Planton's AliCloudKubernetesCluster component codifies these lessons: it validates multi-AZ VSwitch selection, supports both CNI modes through clean field separation, defaults to secure settings where possible, and exposes the full set of production-critical outputs (cluster endpoints, RRSA OIDC metadata, security group and NAT gateway IDs) needed by downstream components.

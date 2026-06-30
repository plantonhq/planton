# OCI Container Engine Cluster: Design Rationale and Research

## Introduction

The OciContainerEngineCluster component manages the OKE control plane — the managed Kubernetes API server, etcd, scheduler, and controller manager. This is one of the highest-leverage components in the OCI catalog: a single cluster resource anchors an entire Kubernetes platform, with node pools, workloads, services, and storage all depending on it.

The spec surface (10 top-level fields, 3 enums, 8 nested messages) reflects OKE's configuration depth. Getting the abstractions right here determines whether users can express their cluster requirements declaratively or need to escape to raw Terraform/Pulumi for missing knobs.

This document explains the design decisions that shaped the OciContainerEngineCluster component and the research that informed them.

## Why Cluster and NodePool Are Separate Resources

OKE's API models clusters and node pools as distinct lifecycle objects, and the Planton component design follows this separation. The reasons are both technical and practical:

1. **Different lifecycles.** Clusters are created once and rarely changed (the control plane is upgraded in place). Node pools are created, scaled, replaced, and deleted frequently. A single "cluster + nodes" resource would force recreation of the entire control plane when scaling workers.

2. **Different scaling concerns.** A production cluster often has multiple node pools with different shapes, sizes, and scaling policies — one for general workloads (E4.Flex), one for GPU workloads (GPU.A10), one for memory-intensive jobs (E4.Flex with high memory ratio). These are independent scaling decisions.

3. **Matches OKE's API.** The Terraform resources are `oci_containerengine_cluster` and `oci_containerengine_node_pool`. The Pulumi SDK mirrors this. Combining them would create an abstraction that fights the provider API rather than wrapping it.

4. **Matches industry precedent.** AWS EKS (cluster + managed node group), GCP GKE (cluster + node pool), and Azure AKS (cluster + agent pool) all separate the control plane from compute. Users coming from any managed Kubernetes service expect this separation.

5. **Infra-chart composability.** The OKE Environment infra chart will compose one OciContainerEngineCluster with one or more OciContainerEngineNodePool resources. Separate resources make this composition natural rather than requiring special-case handling for multi-node-pool configurations.

## Basic vs Enhanced Cluster Types

OKE offers two cluster types, exposed as the `ClusterType` enum:

| Type | Features | Cost |
|------|----------|------|
| `basic_cluster` | Standard Kubernetes control plane features | Free (no per-cluster charge) |
| `enhanced_cluster` | All basic features plus: virtual node pools, workload identity, cluster add-on lifecycle management | Per-cluster/hour charge |

### Why an Enum, Not a Boolean

A boolean (`isEnhanced`) was considered but rejected:

- OCI's API uses a string type (`BASIC_CLUSTER`, `ENHANCED_CLUSTER`). An enum is the natural proto representation.
- If OCI introduces additional cluster tiers in the future (e.g., a "premium" tier with SLA guarantees), an enum extends gracefully. A boolean would require a breaking change.
- The enum value `unspecified` (0) allows the OCI default to apply when the user doesn't explicitly choose.

### One-Way Upgrade

Basic clusters can be upgraded to enhanced, but the reverse is not possible. This is an OCI platform constraint — enhanced clusters enable capabilities (workload identity, virtual nodes) that existing workloads may depend on. Downgrading would break those dependencies.

The `type` field change from `basic_cluster` to `enhanced_cluster` is handled as an in-place update by both the Pulumi and Terraform providers. The reverse change would fail at the OCI API level.

## CNI Type: VCN-Native vs Flannel

The `CniType` enum exposes OKE's two pod networking models:

### VCN-Native IP Allocation (`oci_vcn_ip_native`)

- Every pod receives a VCN IP address from a designated pod subnet.
- Pods are directly addressable within the VCN — no overlay, no encapsulation.
- NSGs can be applied to individual pods (via pod subnet security rules).
- Kubernetes NetworkPolicy is supported (enforced at the VCN level).
- Requires additional subnet planning: pod subnets must have enough IPs for all pods across all nodes. Each node reserves IPs from the pod subnet proportional to its maximum pod count.

### Flannel Overlay (`flannel_overlay`)

- Pods use a cluster-internal overlay network (VXLAN encapsulation).
- Pods are not directly addressable from the VCN — traffic to/from pods routes through the node's IP.
- NSGs apply at the node level, not the pod level.
- Kubernetes NetworkPolicy is supported via Calico (but without VCN-native enforcement).
- Simpler setup: no pod subnet IP planning needed.

### Why Both Are Supported

VCN-native is the recommended production option, but flannel has legitimate use cases:

- **Development clusters** where network policy enforcement is unnecessary and subnet planning overhead is unwanted.
- **IP-constrained environments** where the VCN's available IP space cannot accommodate per-pod allocation across all nodes.
- **Existing clusters** — many OKE clusters were created before VCN-native was available and continue to run on flannel.

### Why CNI Is Immutable

The CNI type fundamentally determines how pods are networked — the IP address allocation scheme, routing tables, and security group association model all differ. Switching CNI would require recreating every pod in the cluster with different networking, which is effectively a cluster replacement. The immutability constraint reflects this reality.

## Endpoint Configuration Design

The `EndpointConfig` message controls how the Kubernetes API server is exposed. The design decisions here diverge from how other providers handle endpoint access.

### Why EndpointConfig Is a Separate Optional Message

Not all clusters need custom endpoint configuration. A development cluster can omit `endpointConfig` entirely and get the OCI default (public endpoint, no dedicated subnet, no NSGs). Making it an optional message means:

- The Quick Start manifest stays minimal (3 required fields only).
- Production clusters add `endpointConfig` when they need private endpoints, dedicated subnets, or NSGs.
- The field grouping makes it clear that `subnetId`, `isPublicIpEnabled`, and `nsgIds` are all part of the same concern (API server network access).

### Subnet-Based vs Access Config Model

OKE's endpoint model is architecturally different from EKS:

- **EKS** uses an "endpoint access configuration" model — toggle public access, private access, or both, then optionally restrict public access to specific CIDRs. The control plane VPC endpoints are managed by AWS.
- **OKE** places the API server endpoint in a specific subnet chosen by the user. The subnet's route table, security lists, and the endpoint's NSGs control access. Public/private is determined by IP assignment and subnet type.

The `EndpointConfig` message mirrors OKE's model directly:
- `subnetId` — which subnet hosts the endpoint
- `isPublicIpEnabled` — whether the endpoint gets a public IP
- `nsgIds` — which NSGs protect the endpoint

This is more explicit than EKS's model (you see exactly where the endpoint lives in your network) but requires more upfront planning (you need a subnet ready before creating the cluster).

### The isPublicIpEnabled Tri-State

Like the compute instance's `assignPublicIp`, the `isPublicIpEnabled` field is an `optional bool`:

| Value | Behavior |
|-------|----------|
| Unset | OCI default (public if the subnet allows; private otherwise) |
| `true` | Assign a public IP to the API endpoint |
| `false` | No public IP — the API server is private-only |

The `optional` wrapper preserves OCI's default when the user doesn't express a preference, while allowing explicit control when needed.

## Why Deprecated Features Are Omitted

The spec proto comment explicitly documents two intentionally omitted provider features:

### add_ons (Kubernetes Dashboard, Tiller)

The OKE API's `add_ons` block controlled the Kubernetes Dashboard and Tiller (Helm 2's server component). Both are removed from modern Kubernetes:

- **Kubernetes Dashboard** was removed as a default cluster add-on in Kubernetes ~1.19. The community version is installed as a standalone deployment.
- **Tiller** was the server component of Helm 2, removed in Helm 3. Modern Helm operates client-side only.

Including these fields would mislead users into thinking they do something useful on current Kubernetes versions. The OCI API still accepts them for backwards compatibility, but they are no-ops.

### admission_controller_options (Pod Security Policy)

The `admission_controller_options` block enabled Pod Security Policy (PSP), which was deprecated in Kubernetes 1.21 and removed in Kubernetes 1.25. The replacement is Pod Security Admission (PSA), which is enabled by default and configured via namespace labels, not cluster-level settings.

Including this field would lead users to enable a removed feature, resulting in confusing failures on Kubernetes >= 1.25.

## OIDC: Two Configuration Modes

The `OpenIdConnectTokenAuthenticationConfig` message supports two mutually exclusive modes:

### Inline Mode

Individual fields (`issuerUrl`, `clientId`, `caCertificate`, etc.) are set directly in the YAML manifest. Best for:

- Simple OIDC configurations with a single provider.
- GitOps workflows where the configuration is reviewed inline in the manifest.
- Environments where the OIDC parameters are stable and infrequently changed.

### Configuration File Mode

A single `configurationFile` field contains a base64-encoded Kubernetes OIDC Auth Config file. Best for:

- Complex configurations generated by CI/CD pipelines.
- Multi-provider setups where the configuration file is a versioned artifact.
- Organizations that manage OIDC configuration as a separate concern from cluster infrastructure.

### Why Mutually Exclusive

The OKE API processes these modes differently internally. When `configurationFile` is set, it takes precedence and the inline fields are ignored. Making them mutually exclusive in the documentation (and conceptually in the proto, though not enforced via `oneof` to keep YAML authoring simple) prevents the subtle bug where a user sets both and the inline fields are silently ignored.

A proto `oneof` was considered but rejected because it changes the YAML structure (adding a wrapper key) and complicates the common case (inline OIDC) for the sake of the uncommon case (config file).

## Service LB and PV Tagging

The `ServiceLbConfig` and `PersistentVolumeConfig` messages propagate tags to resources created by Kubernetes, not by the IaC module. This is a cluster-level configuration that takes effect at runtime:

- When a Kubernetes Service of type LoadBalancer is created, OKE applies the tags from `serviceLbConfig` to the OCI load balancer.
- When a PersistentVolumeClaim creates a block volume, OKE applies the tags from `persistentVolumeConfig` to the OCI block volume.

### Why This Is Part of Cluster Config

Tag propagation at the cluster level has two advantages over per-workload tagging:

1. **No tag drift.** Every load balancer and persistent volume gets the same base tags regardless of which team or namespace created them. Individual workloads don't need to know about tagging policies.

2. **Cost attribution without overhead.** Finance teams can track infrastructure costs by environment, team, or cost center without requiring every Helm chart and Kubernetes manifest to include OCI-specific tag annotations.

The `backendNsgIds` field in `ServiceLbConfig` serves a different purpose — it controls which NSGs protect the load balancer backends (worker nodes). This is a network security concern that naturally belongs alongside the LB subnet and tag configuration.

## What's Deferred

Based on the 80/20 principle, the following features are not in the initial implementation:

- **Cluster Add-On Management** — Enhanced clusters support lifecycle management for cluster add-ons (CoreDNS, kube-proxy, OCI CSI drivers). This is a complex feature with its own update/rollback lifecycle that would add significant spec surface. Deferred until the add-on management workflow is better understood.
- **Cluster Migration** — OCI supports migrating clusters between compartments via a separate API call. This is an operational concern (not a deployment concern) that doesn't fit the IaC model of "declare desired state, apply."
- **Virtual Node Pool Configuration** — Enhanced clusters support virtual nodes (serverless pods managed by OCI without customer-managed instances). Virtual node pools are a node pool variant and would be configured on OciContainerEngineNodePool, not the cluster resource.
- **Cluster Workload Identity Federation** — Enhanced clusters support OCI IAM workload identity, allowing Kubernetes service accounts to assume OCI IAM policies. The federation configuration involves IAM policies (OciIdentityPolicy) and dynamic groups (OciDynamicGroup) rather than cluster-level fields.
- **Defined Tags on the Cluster Itself** — The cluster resource uses Planton freeform tags. OCI defined tags (namespace-scoped, schema-validated tags) require a tag namespace to be created first. Defined tag support can be added later when the tag namespace pattern is established across OCI components.

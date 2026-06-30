# Overview

The **GcpGkeCluster** API resource provides a production-ready, opinionated interface for provisioning Google Kubernetes Engine (GKE) control planes on Google Cloud Platform. By focusing on essential security and networking configurations—private clusters, VPC-native networking, Workload Identity, and network policies—it enables teams to deploy secure, scalable Kubernetes clusters without navigating GKE's extensive configuration landscape.

## Purpose

Deploying a production-ready GKE cluster involves dozens of decisions: Should nodes have public IPs? Which release channel? How to handle pod and service IP allocation? What about Workload Identity? The **GcpGkeCluster** resource solves this by:

- **Enforcing Security Best Practices**: Private clusters with VPC-native networking and Workload Identity enabled by default.
- **Simplifying Network Configuration**: Abstracts VPC secondary ranges, master CIDR allocation, and Cloud NAT requirements into validated foreign key references.
- **Promoting Consistency**: Standardized cluster configurations across dev, staging, and production environments.
- **Separating Concerns**: Control plane (cluster) and compute (node pools) are independent resources, preventing unintended disruptions.

## Key Features

### Private-by-Default Architecture

- **Private Nodes**: Nodes have only RFC1918 (private) IPs, requiring Cloud NAT for outbound internet access.
- **Private Control Plane Endpoint**: Kubernetes API server is accessible only from VPC networks or authorized networks.
- **Public Endpoint Option**: For development environments, public node IPs can be enabled via `enable_public_nodes`.

### VPC-Native Networking (IP Aliasing)

- **Pod and Service Secondary Ranges**: Dedicated IP ranges for pods and services, natively understood by GCP VPC.
- **Required for Modern Features**: Workload Identity, Private Service Connect, and GKE Dataplane V2 require VPC-native networking.
- **No Route-Based Clusters**: Routes-based clusters are deprecated; VPC-native is enforced.

### Workload Identity (IAM for Pods)

- **Pod-Level IAM**: Kubernetes Service Accounts (KSA) bind to Google Service Accounts (GSA) for fine-grained IAM permissions.
- **No Shared Secrets**: Eliminates node-level service account keys; each pod gets only the IAM permissions it needs.
- **Enabled by Default**: Can be disabled for dev environments via `disable_workload_identity`.

### Network Policies (Calico)

- **Microsegmentation**: Control which pods can reach which pods and on which ports using Kubernetes Network Policies.
- **Enabled by Default**: Calico enforcement is on; can be disabled for simplified dev environments via `disable_network_policy`.

### Auto-Upgrade Strategy (Release Channels)

- **REGULAR (Default)**: Production-recommended balance—new Kubernetes versions ~2-3 months after Rapid.
- **RAPID**: Bleeding-edge for dev/test environments.
- **STABLE**: Conservative for risk-averse production.
- **NONE**: Manual upgrades (not recommended—security patches become your responsibility).

### Separate Node Pools

- **Control Plane Only**: This resource provisions the GKE control plane (API server, etcd, controllers).
- **Node Pools Are Independent**: Use the `GcpGkeNodePool` resource to provision compute nodes separately.
- **Production Pattern**: Validated by Terraform, Pulumi, and GKE best practices—prevents lifecycle conflicts.

## Benefits

- **Production-Ready in 10 Lines of YAML**: Minimal configuration for secure, private clusters with sane defaults.
- **Reduced Configuration Drift**: Standardized cluster specs across environments eliminate snowflake infrastructure.
- **Security by Default**: Private nodes, Workload Identity, and network policies reduce attack surface.
- **Simplified Networking**: Foreign key references to VPC subnets, secondary ranges, and Cloud NAT ensure validated, consistent network configurations.
- **Lifecycle Independence**: Update node pools (machine types, autoscaling, labels) without touching the control plane.

## Example Usage

Below is a minimal YAML manifest for a production-ready regional GKE cluster:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpGkeCluster
metadata:
  name: prod-cluster
  org: my-org
  env:
    id: production
spec:
  project_id:
    value_ref:
      kind: GcpProject
      name: my-project
      path: status.outputs.project_id
  location: us-central1  # Regional cluster for HA
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
  enable_public_nodes: false
  release_channel: REGULAR
  disable_network_policy: false
  disable_workload_identity: false
  router_nat_name:
    value_ref:
      kind: GcpRouterNat
      name: prod-nat
      path: metadata.name
```

### What This Creates

- **Regional GKE cluster** in `us-central1` (control plane replicas across 3+ zones for high availability)
- **Private nodes** (no public IPs) with outbound internet via Cloud NAT
- **VPC-native networking** using secondary IP ranges from the subnet
- **Workload Identity enabled** for pod-level IAM
- **Network policies enabled** for Calico enforcement
- **Auto-upgrades on REGULAR channel** for security patches

## Deploying with Planton

Once your YAML manifest is ready, deploy using the Planton CLI:

- **Using Pulumi**:
  ```bash
  planton pulumi up --manifest gcpgkecluster.yaml --stack org/project/my-stack
  ```

- **Using Terraform**:
  ```bash
  planton terraform apply --manifest gcpgkecluster.yaml --stack org/project/my-stack
  ```

Planton will:
1. Validate the manifest against the Protobuf schema
2. Provision the GKE control plane with the specified configuration
3. Export outputs (cluster endpoint, CA certificate, Workload Identity pool)
4. Make the cluster ready for node pool provisioning

## Prerequisites

Before creating a GKE cluster, ensure these resources exist:

1. **GcpProject**: The GCP project where the cluster will be created
2. **GcpVpc**: The VPC network (custom mode, not auto mode)
3. **GcpSubnetwork**: A subnet with:
   - Primary IP range for nodes
   - Secondary range for pod IPs (`cluster_secondary_range_name`)
   - Secondary range for service IPs (`services_secondary_range_name`)
4. **GcpRouterNat**: Cloud NAT configuration for private node egress

## Regional vs Zonal Clusters

### Regional Clusters (Recommended for Production)

```yaml
spec:
  location: us-central1  # Region
```

- **Control plane**: Replicated across 3+ zones
- **High availability**: Survives zonal outages
- **Higher cost**: Control plane replicas cost more
- **Use when**: Production workloads requiring HA

### Zonal Clusters (For Dev/Test)

```yaml
spec:
  location: us-central1-a  # Zone
```

- **Control plane**: Single zone
- **Lower cost**: Single control plane replica
- **No zonal HA**: Control plane failure = API downtime
- **Use when**: Dev/test environments, cost-sensitive non-critical workloads

## Release Channels Explained

| Channel | Use Case | Kubernetes Version Availability | Auto-Upgrades |
|---------|----------|--------------------------------|---------------|
| **RAPID** | Dev/test environments | New versions ~2-3 months before Stable | Yes |
| **REGULAR** | Production (recommended) | New versions ~2-3 months after Rapid | Yes |
| **STABLE** | Risk-averse production | New versions ~2-3 months after Regular | Yes |
| **NONE** | Manual upgrades (not recommended) | Manual only | No |

**Best practice**: Use `REGULAR` for production. Combine with maintenance windows (define recurring time windows when upgrades are permitted, e.g., "Sunday 02:00-04:00 AM").

## Workload Identity Setup

After creating a cluster with Workload Identity enabled:

1. **Create a Google Service Account (GSA)**:
   ```bash
   gcloud iam service-accounts create my-app-sa \
     --project=my-project \
     --display-name="My Application Service Account"
   ```

2. **Grant IAM permissions to the GSA**:
   ```bash
   gcloud projects add-iam-policy-binding my-project \
     --member="serviceAccount:my-app-sa@my-project.iam.gserviceaccount.com" \
     --role="roles/storage.objectViewer"
   ```

3. **Create a Kubernetes Service Account (KSA)** with Workload Identity annotation:
   ```yaml
   apiVersion: v1
   kind: ServiceAccount
   metadata:
     name: my-app-sa
     namespace: production
     annotations:
       iam.gke.io/gcp-service-account: my-app-sa@my-project.iam.gserviceaccount.com
   ```

4. **Bind the KSA to the GSA**:
   ```bash
   gcloud iam service-accounts add-iam-policy-binding \
     my-app-sa@my-project.iam.gserviceaccount.com \
     --role=roles/iam.workloadIdentityUser \
     --member="serviceAccount:my-project.svc.id.goog[production/my-app-sa]"
   ```

5. **Use the KSA in your pods**:
   ```yaml
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     name: my-app
   spec:
     template:
       spec:
         serviceAccountName: my-app-sa  # Pods get GSA's IAM permissions
   ```

## Network Policies for Microsegmentation

With network policies enabled (default), define which pods can reach which services:

```yaml
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

**Result**: Backend pods accept traffic only from frontend pods on port 8080. All other traffic is denied.

## Master CIDR Planning

Every GKE cluster requires a `/28` CIDR block (16 IPs) for the Kubernetes control plane's private endpoint.

**Critical constraints**:
- Must be `/28` (exactly 16 IPs)—no larger, no smaller
- Must be RFC1918 private IP range
- Cannot overlap with VPC primary range, pod range, service range, or peered VPC ranges
- **Cannot be changed after cluster creation** (immutable)

**Planning strategy**: Reserve a dedicated `/16` subnet for GKE master CIDRs:
- `172.16.0.0/28` → Cluster 1 masters
- `172.16.0.16/28` → Cluster 2 masters
- `172.16.0.32/28` → Cluster 3 masters
- ...up to 4096 clusters in `172.16.0.0/16`

## Outputs

After deployment, the following outputs are available:

- **endpoint**: The IP address of the cluster's Kubernetes API server
- **cluster_ca_certificate**: Base64 encoded public certificate for verifying the cluster's CA (sensitive)
- **workload_identity_pool**: The Workload Identity Pool for this cluster (format: `PROJECT_ID.svc.id.goog`)

Use these outputs to configure `kubectl` and deploy node pools:

```bash
gcloud container clusters get-credentials prod-cluster \
  --region=us-central1 \
  --project=my-project
```

## Next Steps

1. **Provision Node Pools**: Use the `GcpGkeNodePool` resource to add compute nodes
2. **Configure kubectl**: Run `gcloud container clusters get-credentials` to access the cluster
3. **Deploy Workloads**: Apply Kubernetes manifests for your applications
4. **Set Up Workload Identity**: Bind Kubernetes Service Accounts to Google Service Accounts for IAM
5. **Define Network Policies**: Implement microsegmentation for production security

## Best Practices

- **Use Regional Clusters for Production**: Zonal clusters lack control plane HA
- **Enable Workload Identity**: Avoid node-level service account keys
- **Enable Network Policies**: Implement microsegmentation from day one
- **Use Regular Release Channel**: Balance new features with stability
- **Plan Master CIDR Allocation**: Cannot be changed after cluster creation
- **Separate Control Plane and Node Pools**: Independent lifecycles prevent disruptions
- **Use Cloud NAT for Private Nodes**: Required for internet egress (Docker images, APIs)

---

For comprehensive design rationale, deployment patterns, and production best practices, see the [research documentation](docs/README.md).

Happy deploying! If you have questions or run into issues, feel free to open an issue on our GitHub repository or reach out through our community channels for support.


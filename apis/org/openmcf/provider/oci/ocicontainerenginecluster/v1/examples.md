# OCI Container Engine Cluster Examples

This document provides practical examples for deploying Oracle Cloud Infrastructure Container Engine for Kubernetes (OKE) clusters using the OpenMCF API. Each example demonstrates a different use case, progressing from a minimal development cluster to a fully secured production cluster with OIDC, KMS encryption, and image verification.

## Table of Contents

- [Example 1: Minimal Basic Cluster](#example-1-minimal-basic-cluster)
- [Example 2: Production Private Cluster with VCN-Native CNI](#example-2-production-private-cluster-with-vcn-native-cni)
- [Example 3: Enhanced Cluster with Inline OIDC](#example-3-enhanced-cluster-with-inline-oidc)
- [Example 4: Enhanced Cluster with OIDC Configuration File](#example-4-enhanced-cluster-with-oidc-configuration-file)
- [Example 5: KMS-Encrypted Cluster with Image Policy](#example-5-kms-encrypted-cluster-with-image-policy)
- [Example 6: Dual-Stack IPv4+IPv6 Cluster](#example-6-dual-stack-ipv4ipv6-cluster)
- [Common Operations](#common-operations)
- [Best Practices](#best-practices)

---

## Example 1: Minimal Basic Cluster

**Use Case:** A development or testing cluster with the minimum required configuration. Uses the OCI default CNI (flannel overlay) and default endpoint settings.

**Configuration:**
- **Cluster Type:** Basic (default, no per-cluster charge)
- **CNI:** OCI default (flannel overlay)
- **Endpoint:** OCI default (public, no dedicated subnet)

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciContainerEngineCluster
metadata:
  name: dev-cluster
  org: my-org
  env: dev
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  vcnId:
    value: "ocid1.vcn.oc1.iad.example"
  kubernetesVersion: "v1.28.2"
```

**Deploy with OpenMCF CLI:**

```bash
openmcf apply -f dev-cluster.yaml
```

**What happens:**
- A basic OKE cluster is created with a managed Kubernetes control plane (API server, etcd, scheduler, controller manager).
- The cluster uses flannel overlay CNI and the default pod CIDR (`10.244.0.0/16`) and service CIDR (`10.96.0.0/16`).
- No worker nodes are created — add OciContainerEngineNodePool resources to provision compute.
- The cluster ID, Kubernetes version, and API endpoint URLs are exported as stack outputs.

---

## Example 2: Production Private Cluster with VCN-Native CNI

**Use Case:** A production cluster with a private API endpoint, VCN-native pod networking, NSG-secured endpoint, dedicated service load balancer subnets, and custom network CIDRs. All infrastructure references use `valueFrom` for declarative composition with other OpenMCF resources.

**Configuration:**
- **Cluster Type:** Enhanced (workload identity, add-on management)
- **CNI:** VCN-native (pods get VCN IPs, NSGs and network policies on pods)
- **Endpoint:** Private, in a dedicated subnet, protected by NSGs
- **Service LBs:** Dedicated public subnet with backend NSGs
- **Network CIDRs:** Custom pod and service CIDRs

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciContainerEngineCluster
metadata:
  name: prod-cluster
  org: acme-corp
  env: prod
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-compartment
      fieldPath: status.outputs.compartmentId
  vcnId:
    valueFrom:
      kind: OciVcn
      name: prod-vcn
      fieldPath: status.outputs.vcnId
  kubernetesVersion: "v1.28.2"
  type: enhanced_cluster
  cniType: oci_vcn_ip_native
  endpointConfig:
    subnetId:
      valueFrom:
        kind: OciSubnet
        name: api-endpoint-subnet
        fieldPath: status.outputs.subnetId
    isPublicIpEnabled: false
    nsgIds:
      - valueFrom:
          kind: OciSecurityGroup
          name: api-endpoint-nsg
          fieldPath: status.outputs.networkSecurityGroupId
  options:
    kubernetesNetworkConfig:
      podsCidr: "10.244.0.0/16"
      servicesCidr: "10.96.0.0/16"
    serviceLbSubnetIds:
      - valueFrom:
          kind: OciSubnet
          name: public-lb-subnet-ad1
          fieldPath: status.outputs.subnetId
      - valueFrom:
          kind: OciSubnet
          name: public-lb-subnet-ad2
          fieldPath: status.outputs.subnetId
    serviceLbConfig:
      backendNsgIds:
        - valueFrom:
            kind: OciSecurityGroup
            name: worker-nsg
            fieldPath: status.outputs.networkSecurityGroupId
      freeformTags:
        environment: "production"
        managed-by: "kubernetes"
    persistentVolumeConfig:
      freeformTags:
        environment: "production"
        managed-by: "kubernetes"
```

**What happens:**
- An enhanced OKE cluster is created with VCN-native CNI — every pod receives a VCN IP address, enabling NSGs and Kubernetes network policies at the pod level.
- The Kubernetes API server endpoint is placed in a private subnet with no public IP. Access requires VPN, VCN peering, or a bastion host.
- NSGs on the API endpoint restrict which networks can reach the API server (e.g., only the corporate VPN CIDR).
- Service load balancers created by Kubernetes are placed in two public subnets across availability domains for HA. Backend NSGs control traffic between load balancers and worker nodes.
- Freeform tags are propagated to all load balancers and persistent volumes created by Kubernetes for cost tracking.

---

## Example 3: Enhanced Cluster with Inline OIDC

**Use Case:** A cluster integrated with an enterprise identity provider (Okta, Azure AD, Auth0, Keycloak) for `kubectl` and API access. Developers authenticate with their corporate credentials instead of OCI IAM tokens.

**Configuration:**
- **OIDC Mode:** Inline (individual fields)
- **Identity Provider:** Corporate OIDC provider
- **Claims:** Email as username, groups for RBAC

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciContainerEngineCluster
metadata:
  name: sso-cluster
  org: acme-corp
  env: prod
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  vcnId:
    value: "ocid1.vcn.oc1.iad.example"
  kubernetesVersion: "v1.28.2"
  type: enhanced_cluster
  cniType: oci_vcn_ip_native
  endpointConfig:
    subnetId:
      value: "ocid1.subnet.oc1.iad.example"
    isPublicIpEnabled: false
  options:
    openIdConnectTokenAuthenticationConfig:
      isOpenIdConnectAuthEnabled: true
      issuerUrl: "https://login.example.com/oauth2/default"
      clientId: "0oa1bcdef2ghijk3lmn4"
      caCertificate: "LS0tLS1CRUdJTi..."
      usernameClaim: "email"
      usernamePrefix: "oidc:"
      groupsClaim: "groups"
      groupsPrefix: "oidc:"
      signingAlgorithms:
        - "RS256"
      requiredClaims:
        - key: "iss"
          value: "https://login.example.com/oauth2/default"
    isOpenIdConnectDiscoveryEnabled: true
```

**What happens:**
- The Kubernetes API server is configured to accept OIDC tokens from the specified identity provider.
- Users authenticate with `kubectl` using their corporate SSO credentials. The `email` claim becomes the Kubernetes username (prefixed with `oidc:`), and the `groups` claim maps to Kubernetes RBAC groups.
- The `requiredClaims` field enforces that only tokens from the correct issuer are accepted — an additional security layer beyond issuer URL validation.
- OIDC Discovery is enabled, allowing external systems to discover the cluster's public signing keys at the well-known endpoint.
- The `caCertificate` field provides the IdP's certificate for TLS verification when the API server contacts the issuer URL.

---

## Example 4: Enhanced Cluster with OIDC Configuration File

**Use Case:** Same identity integration goal as Example 3, but using a base64-encoded Kubernetes OIDC Auth Config file instead of inline fields. Preferred when the OIDC configuration is managed as a file artifact in CI/CD pipelines.

**Configuration:**
- **OIDC Mode:** Configuration file (base64-encoded)

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciContainerEngineCluster
metadata:
  name: sso-cluster-v2
  org: acme-corp
  env: prod
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  vcnId:
    value: "ocid1.vcn.oc1.iad.example"
  kubernetesVersion: "v1.28.2"
  type: enhanced_cluster
  cniType: oci_vcn_ip_native
  endpointConfig:
    subnetId:
      value: "ocid1.subnet.oc1.iad.example"
    isPublicIpEnabled: false
  options:
    openIdConnectTokenAuthenticationConfig:
      isOpenIdConnectAuthEnabled: true
      configurationFile: "eyJraW5kIjoiQXV0aGVudGljYXRpb25Db25maWd1cmF0aW9uIi..."
    isOpenIdConnectDiscoveryEnabled: true
```

**What happens:**
- The OIDC configuration is provided as a single base64-encoded file instead of individual fields.
- The `configurationFile` field is mutually exclusive with the inline fields (`issuerUrl`, `clientId`, etc.). Set one or the other, not both.
- This mode is useful when the OIDC configuration is generated by a CI/CD pipeline or managed as a versioned file artifact.

**Encoding the configuration file:**

```bash
base64 -w 0 < oidc-config.json
```

---

## Example 5: KMS-Encrypted Cluster with Image Policy

**Use Case:** A compliance-oriented cluster where Kubernetes secrets are encrypted with a customer-managed KMS key and all container images must be signed with approved keys before deployment.

**Configuration:**
- **KMS:** Customer-managed key for etcd secrets encryption
- **Image Policy:** Enabled, two signing keys

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciContainerEngineCluster
metadata:
  name: secure-cluster
  org: acme-corp
  env: prod
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  vcnId:
    value: "ocid1.vcn.oc1.iad.example"
  kubernetesVersion: "v1.28.2"
  type: enhanced_cluster
  cniType: oci_vcn_ip_native
  kmsKeyId:
    value: "ocid1.key.oc1.iad.examplesecretkey"
  imagePolicyConfig:
    isPolicyEnabled: true
    keyDetails:
      - kmsKeyId:
          value: "ocid1.key.oc1.iad.examplesigningkey1"
      - kmsKeyId:
          value: "ocid1.key.oc1.iad.examplesigningkey2"
  endpointConfig:
    subnetId:
      value: "ocid1.subnet.oc1.iad.example"
    isPublicIpEnabled: false
    nsgIds:
      - value: "ocid1.networksecuritygroup.oc1.iad.example"
```

**What happens:**
- All Kubernetes secrets stored in etcd are encrypted at rest using the specified customer-managed KMS key. This provides customer control over the encryption key lifecycle (rotation, revocation) beyond OCI's default Oracle-managed encryption.
- Container image signature verification is enabled. Every image pulled into the cluster must be signed with at least one of the two specified KMS keys. Unsigned or incorrectly signed images are rejected at admission.
- Two signing keys support key rotation — sign new images with the new key while the old key remains valid for existing images.
- The `kmsKeyId` for secrets and the `keyDetails` for image signing are separate keys serving different purposes (data encryption vs signature verification).

---

## Example 6: Dual-Stack IPv4+IPv6 Cluster

**Use Case:** A cluster that supports both IPv4 and IPv6 addressing for pods and services, enabling gradual IPv6 adoption or workloads that require IPv6 connectivity.

**Configuration:**
- **IP Families:** IPv4 + IPv6 (dual-stack)
- **CNI:** VCN-native (required for dual-stack)
- **Network CIDRs:** Custom IPv4 and IPv6 CIDRs

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciContainerEngineCluster
metadata:
  name: dualstack-cluster
  org: acme-corp
  env: dev
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  vcnId:
    value: "ocid1.vcn.oc1.iad.example"
  kubernetesVersion: "v1.28.2"
  type: enhanced_cluster
  cniType: oci_vcn_ip_native
  endpointConfig:
    subnetId:
      value: "ocid1.subnet.oc1.iad.example"
  options:
    kubernetesNetworkConfig:
      podsCidr: "10.244.0.0/16"
      servicesCidr: "10.96.0.0/16"
    ipFamilies:
      - ipv4
      - ipv6
```

**What happens:**
- The cluster is created with dual-stack networking — pods and services receive both IPv4 and IPv6 addresses.
- VCN-native CNI is required for dual-stack; flannel overlay does not support IPv6.
- The `ipFamilies` field is immutable after creation. Plan your IP addressing strategy before creating the cluster.
- IPv6 pod and service CIDRs default to `fd00:eeee:eeee:0000::/96` and `fd00:eeee:eeee:0001::/108` respectively when not explicitly configured.

---

## Common Operations

### Generate Kubeconfig

After deploying a cluster, generate a kubeconfig file to interact with it using `kubectl`:

```bash
# Get the cluster ID from stack outputs
CLUSTER_ID=$(pulumi stack output cluster_id)

# Generate kubeconfig
oci ce cluster create-kubeconfig \
  --cluster-id "$CLUSTER_ID" \
  --file ~/.kube/config \
  --region us-ashburn-1 \
  --token-version 2.0.0

# Verify connectivity
kubectl get nodes
```

For Terraform deployments:

```bash
CLUSTER_ID=$(terraform output -raw cluster_id)

oci ce cluster create-kubeconfig \
  --cluster-id "$CLUSTER_ID" \
  --file ~/.kube/config \
  --region us-ashburn-1 \
  --token-version 2.0.0
```

### Get Cluster Endpoints

```bash
# Pulumi
pulumi stack output kubernetes_endpoint
pulumi stack output private_endpoint
pulumi stack output public_endpoint

# Terraform
terraform output kubernetes_endpoint
terraform output private_endpoint
terraform output public_endpoint
```

The `public_endpoint` output is empty when the cluster has a private-only endpoint (`isPublicIpEnabled: false`).

### Use Cluster ID in Downstream Resources

The `cluster_id` output is referenced by OciContainerEngineNodePool to attach worker nodes:

```yaml
spec:
  clusterId:
    valueFrom:
      kind: OciContainerEngineCluster
      name: prod-cluster
      fieldPath: status.outputs.clusterId
```

### Check Available Kubernetes Versions

```bash
oci ce cluster-options list --cluster-option-id all
```

This returns the Kubernetes versions available for new clusters and upgrades in your region.

---

## Best Practices

### Cluster Type Selection

| Scenario | Recommended Type | Rationale |
|----------|-----------------|-----------|
| Development / testing | `basic_cluster` | No per-cluster charge. Standard Kubernetes features are sufficient. |
| Production workloads | `enhanced_cluster` | Workload identity, add-on lifecycle management, virtual node pool support. |
| Multi-tenant platforms | `enhanced_cluster` | Workload identity enables fine-grained IAM integration per namespace. |

**Start with enhanced for production.** The per-cluster charge is modest compared to the node pool compute cost, and enhanced features (workload identity, add-on management) are difficult to retrofit later.

### CNI Selection

| Scenario | Recommended CNI | Rationale |
|----------|----------------|-----------|
| Production | `oci_vcn_ip_native` | Pods get VCN IPs. Enables NSGs on pods, Kubernetes network policies, and direct VCN routing to pods. |
| Development / simple workloads | `flannel_overlay` | Simpler setup. Acceptable when network policy enforcement and pod-level NSGs are not needed. |
| Dual-stack (IPv4+IPv6) | `oci_vcn_ip_native` | Required for dual-stack support. |

**VCN-native is the production default.** The additional subnet planning (pod subnets need enough IPs for all pods across all nodes) is worth the network isolation and observability benefits.

### Endpoint Security

For production clusters:

- Set `endpointConfig.isPublicIpEnabled: false` — the API server should not be reachable from the internet.
- Place the endpoint in a dedicated private subnet — separate from worker node and load balancer subnets.
- Apply NSGs to the endpoint (`endpointConfig.nsgIds`) — allow access only from corporate VPN CIDRs, bastion subnets, or CI/CD runner subnets.
- Access the private API server via VPN, VCN peering, OCI Bastion service, or a jump host in the VCN.

### Network CIDR Planning

Plan CIDRs before cluster creation — they are immutable:

| CIDR | Default | Constraint |
|------|---------|------------|
| VCN CIDR | (set on OciVcn) | Must not overlap with pod or service CIDRs |
| Pod CIDR | `10.244.0.0/16` | Must not overlap with VCN or service CIDRs. Size determines max pods per cluster. |
| Service CIDR | `10.96.0.0/16` | Must not overlap with VCN or pod CIDRs. Size determines max services. |

For VCN-native CNI, also plan pod subnet CIDRs — each node consumes IPs from the pod subnet proportional to its max pod count.

### Service LB Subnet Planning

- Designate one or two public subnets for service load balancers (`options.serviceLbSubnetIds`).
- Use subnets in different availability domains for HA.
- Size subnets to accommodate the expected number of Kubernetes Service LoadBalancers (each service creates one LB consuming one IP).
- Apply backend NSGs (`options.serviceLbConfig.backendNsgIds`) to control traffic between load balancers and worker nodes.

### Version Management

- Pin `kubernetesVersion` explicitly — don't rely on defaults.
- Check available versions with `oci ce cluster-options list` before deploying.
- Plan regular version upgrades — OKE deprecates older versions on a schedule.
- Upgrade the control plane first, then node pools. The node pool Kubernetes version must be within one minor version of the control plane.

### Tag for Cost and Compliance

Use `serviceLbConfig` and `persistentVolumeConfig` tags to track Kubernetes-created resources:

```yaml
options:
  serviceLbConfig:
    freeformTags:
      environment: "production"
      team: "platform"
      cost-center: "infrastructure"
  persistentVolumeConfig:
    freeformTags:
      environment: "production"
      team: "platform"
      cost-center: "infrastructure"
```

These tags propagate to every load balancer and block volume created by Kubernetes, eliminating the need for per-workload tagging policies.

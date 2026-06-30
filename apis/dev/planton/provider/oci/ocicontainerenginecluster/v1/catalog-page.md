# OCI Container Engine Cluster

Deploys an Oracle Cloud Infrastructure Container Engine for Kubernetes (OKE) cluster — a managed Kubernetes control plane with API server, etcd, scheduler, and controller manager. Supports basic and enhanced cluster types, VCN-native and flannel overlay CNI for pod networking, private or public API endpoints, OIDC authentication, KMS secrets encryption, and container image signature verification. Worker nodes are managed separately via OciContainerEngineNodePool.

## What Gets Created

When you deploy an OciContainerEngineCluster resource, Planton provisions:

- **OKE Cluster** — an `oci_containerengine_cluster` resource in the specified compartment and VCN. The cluster runs a managed Kubernetes control plane at the requested version. Standard Planton freeform tags are applied for resource tracking.

## Prerequisites

- **OCI credentials** configured via environment variables or Planton provider config (API Key, Instance Principal, Security Token, Resource Principal, or OKE Workload Identity)
- **A compartment OCID** where the cluster will be created — literal value or reference to an OciCompartment resource
- **A VCN OCID** where the cluster will be deployed — literal value or reference to an OciVcn resource
- **A Kubernetes version string** supported by OKE (e.g., `v1.28.2`) — run `oci ce cluster-options list` to see available versions
- **A subnet OCID** for the API server endpoint if configuring `endpointConfig` — literal value or reference to an OciSubnet resource

## Quick Start

Create a file `cluster.yaml`:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciContainerEngineCluster
metadata:
  name: my-cluster
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.OciContainerEngineCluster.my-cluster
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  vcnId:
    value: "ocid1.vcn.oc1.iad.example"
  kubernetesVersion: "v1.28.2"
```

Deploy:

```shell
planton apply -f cluster.yaml
```

This creates a basic OKE cluster with the default flannel overlay CNI in the specified VCN. OKE provisions the control plane with an API server, etcd, scheduler, and controller manager. The cluster ID, Kubernetes version, and API endpoint URLs are exported as stack outputs. Add OciContainerEngineNodePool resources to create worker nodes.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `compartmentId` | `StringValueOrRef` | OCID of the compartment where the cluster will be created. Changing this after creation forces cluster recreation. Can reference an OciCompartment resource via `valueFrom`. | Required |
| `vcnId` | `StringValueOrRef` | OCID of the VCN where the cluster will be deployed. Changing this after creation forces cluster recreation. Can reference an OciVcn resource via `valueFrom`. | Required |
| `kubernetesVersion` | `string` | Kubernetes version to install on the control plane (e.g., `v1.28.2`). Use `oci ce cluster-options list` to see available versions. | Minimum 1 character |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `name` | `string` | `metadata.name` | Human-readable name shown in the OCI Console. |
| `type` | `enum` | `unspecified` (OCI default) | Cluster type. Values: `basic_cluster` (standard features), `enhanced_cluster` (virtual node pools, workload identity, cluster add-on management). Upgrading from basic to enhanced is one-way. |
| `cniType` | `enum` | `cni_unspecified` (OCI default) | Container Network Interface plugin. Values: `flannel_overlay` (overlay network, simpler), `oci_vcn_ip_native` (pods get VCN IPs, enables network policies and NSGs on pods). Changing this after creation forces cluster recreation. |
| `endpointConfig` | `EndpointConfig` | — | Network configuration for the Kubernetes API server endpoint. See [endpointConfig fields](#endpointconfig-fields). |
| `options` | `ClusterOptions` | — | Networking, service load balancer, persistent volume, and OIDC configuration. See [options fields](#options-fields). |
| `kmsKeyId` | `StringValueOrRef` | — | OCID of a KMS key to encrypt Kubernetes secrets at rest. Requires `kubernetesVersion` >= v1.13.0. Changing this after creation forces cluster recreation. |
| `imagePolicyConfig` | `ImagePolicyConfig` | — | Container image signature verification policy. See [imagePolicyConfig fields](#imagepolicyconfig-fields). |

### endpointConfig Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `subnetId` | `StringValueOrRef` | OCID of the regional subnet hosting the Kubernetes API server endpoint. Changing this after creation forces cluster recreation. Can reference an OciSubnet resource via `valueFrom`. | Required |
| `isPublicIpEnabled` | `bool` | Whether to assign a public IP to the API server endpoint. Set to `false` for private clusters. Must be `false` when the subnet is private. When unset, uses the OCI default. | Optional |
| `nsgIds` | `StringValueOrRef[]` | OCIDs of network security groups applied to the API server endpoint. Can reference OciSecurityGroup resources via `valueFrom`. | Optional |

### options Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `kubernetesNetworkConfig` | `KubernetesNetworkConfig` | — | Pod and service CIDR configuration. Changing this after creation forces cluster recreation. See [kubernetesNetworkConfig fields](#kubernetesnetworkconfig-fields). |
| `serviceLbSubnetIds` | `StringValueOrRef[]` | — | OCIDs of subnets where Kubernetes Service load balancers will be placed. Typically one or two public subnets. Changing this after creation forces cluster recreation. Can reference OciSubnet resources via `valueFrom`. |
| `ipFamilies` | `IpFamily[]` | `[ipv4]` | IP address families for the cluster. Values: `ipv4`, `ipv6`. Use `[ipv4]` for IPv4-only or `[ipv4, ipv6]` for dual-stack. Changing this after creation forces cluster recreation. |
| `serviceLbConfig` | `ServiceLbConfig` | — | Default configuration applied to load balancers created by Kubernetes Service resources. See [serviceLbConfig fields](#servicelbconfig-fields). |
| `persistentVolumeConfig` | `PersistentVolumeConfig` | — | Default tags applied to block volumes created by Kubernetes PersistentVolumeClaim resources. See [persistentVolumeConfig fields](#persistentvolumeconfig-fields). |
| `openIdConnectTokenAuthenticationConfig` | `OpenIdConnectTokenAuthenticationConfig` | — | OIDC token authentication for the API server. See [openIdConnectTokenAuthenticationConfig fields](#openidconnecttokenauthenticationconfig-fields). |
| `isOpenIdConnectDiscoveryEnabled` | `bool` | `false` | When `true`, enables the cluster-specific OIDC Discovery endpoint, allowing external systems to discover the cluster's public signing keys. |

### kubernetesNetworkConfig Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `podsCidr` | `string` | `10.244.0.0/16` (IPv4), `fd00:eeee:eeee:0000::/96` (IPv6) | CIDR block for Kubernetes pods. Must not overlap with the VCN CIDR or the services CIDR. |
| `servicesCidr` | `string` | `10.96.0.0/16` (IPv4), `fd00:eeee:eeee:0001::/108` (IPv6) | CIDR block for Kubernetes services (ClusterIP range). Must not overlap with the VCN CIDR or the pods CIDR. |

### serviceLbConfig Fields

| Field | Type | Description |
|-------|------|-------------|
| `backendNsgIds` | `StringValueOrRef[]` | OCIDs of NSGs applied to service load balancer backends. Can reference OciSecurityGroup resources via `valueFrom`. |
| `freeformTags` | `map<string, string>` | Freeform tags applied to service load balancers created by Kubernetes. |
| `definedTags` | `map<string, string>` | Defined tags applied to service load balancers. Keys use the format `namespace.key` (e.g., `Operations.CostCenter`). |

### persistentVolumeConfig Fields

| Field | Type | Description |
|-------|------|-------------|
| `freeformTags` | `map<string, string>` | Freeform tags applied to block volumes created by Kubernetes PersistentVolumeClaim resources. |
| `definedTags` | `map<string, string>` | Defined tags applied to persistent volumes. Keys use the format `namespace.key` (e.g., `Operations.CostCenter`). |

### openIdConnectTokenAuthenticationConfig Fields

Two configuration modes are supported: **inline** (set individual fields) and **configuration file** (set `configurationFile` with a base64-encoded Kubernetes OIDC Auth Config). These modes are mutually exclusive.

| Field | Type | Description |
|-------|------|-------------|
| `isOpenIdConnectAuthEnabled` | `bool` | Whether OIDC token authentication is enabled. |
| `configurationFile` | `string` | Base64-encoded Kubernetes OIDC Auth Config file. Mutually exclusive with the inline fields below. |
| `issuerUrl` | `string` | URL of the OIDC identity provider. Must use `https://`. The API server uses this to discover signing keys. |
| `clientId` | `string` | Client ID that all tokens must be issued for. |
| `caCertificate` | `string` | Base64-encoded public RSA or ECDSA certificate of the identity provider. |
| `usernameClaim` | `string` | JWT claim to use as the Kubernetes username. Default: `sub`. |
| `usernamePrefix` | `string` | Prefix prepended to username claims to prevent collisions with existing names (e.g., `oidc:`). |
| `groupsClaim` | `string` | JWT claim to use as the user's group. Must be an array of strings in the token. |
| `groupsPrefix` | `string` | Prefix prepended to group claims. |
| `signingAlgorithms` | `string[]` | Accepted signing algorithms for tokens. Default: `["RS256"]`. |
| `requiredClaims` | `RequiredClaim[]` | Key-value pairs that must be present in the ID token. Each entry has `key` and `value` fields. If any required claim is missing or has a different value, authentication is rejected. |

### imagePolicyConfig Fields

| Field | Type | Description |
|-------|------|-------------|
| `isPolicyEnabled` | `bool` | Whether the image verification policy is enabled. |
| `keyDetails` | `ImagePolicyKeyDetail[]` | KMS keys used for image signature verification. Each entry has a `kmsKeyId` field (`StringValueOrRef`) — the OCID of the KMS key used to verify image signatures. |

## Examples

### Minimal Basic Cluster

A basic OKE cluster with only the required fields — the simplest path to a running Kubernetes control plane:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciContainerEngineCluster
metadata:
  name: dev-cluster
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.OciContainerEngineCluster.dev-cluster
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  vcnId:
    value: "ocid1.vcn.oc1.iad.example"
  kubernetesVersion: "v1.28.2"
```

### Production Private Cluster with VCN-Native CNI

An enhanced cluster with VCN-native pod networking, a private API endpoint protected by NSGs, dedicated service load balancer subnets, and custom pod/service CIDRs. All infrastructure references use `valueFrom` for declarative composition:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciContainerEngineCluster
metadata:
  name: prod-cluster
  org: acme
  env: prod
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: acme
    pulumi.planton.dev/project: oke-platform
    pulumi.planton.dev/stack.name: prod.OciContainerEngineCluster.prod-cluster
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
          name: lb-subnet
          fieldPath: status.outputs.subnetId
    serviceLbConfig:
      backendNsgIds:
        - valueFrom:
            kind: OciSecurityGroup
            name: worker-nsg
            fieldPath: status.outputs.networkSecurityGroupId
```

### Enhanced Cluster with OIDC Authentication

An enhanced cluster integrated with an external OIDC identity provider for `kubectl` and API access. Inline OIDC fields configure the issuer, client, and claim mappings directly:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciContainerEngineCluster
metadata:
  name: sso-cluster
  org: acme
  env: prod
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: acme
    pulumi.planton.dev/project: oke-platform
    pulumi.planton.dev/stack.name: prod.OciContainerEngineCluster.sso-cluster
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
      issuerUrl: "https://idp.example.com"
      clientId: "oke-cluster-client"
      usernameClaim: "email"
      usernamePrefix: "oidc:"
      groupsClaim: "groups"
      groupsPrefix: "oidc:"
      signingAlgorithms:
        - "RS256"
      requiredClaims:
        - key: "aud"
          value: "oke-cluster-client"
    isOpenIdConnectDiscoveryEnabled: true
```

### KMS-Encrypted Cluster with Image Policy

A cluster with KMS-encrypted Kubernetes secrets and container image signature verification — all images deployed to the cluster must be signed with an approved KMS key:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciContainerEngineCluster
metadata:
  name: secure-cluster
  org: acme
  env: prod
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: acme
    pulumi.planton.dev/project: oke-platform
    pulumi.planton.dev/stack.name: prod.OciContainerEngineCluster.secure-cluster
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  vcnId:
    value: "ocid1.vcn.oc1.iad.example"
  kubernetesVersion: "v1.28.2"
  type: enhanced_cluster
  kmsKeyId:
    value: "ocid1.key.oc1.iad.example"
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
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `cluster_id` | `string` | OCID of the OKE cluster. |
| `kubernetes_version` | `string` | Kubernetes version running on the cluster control plane. |
| `kubernetes_endpoint` | `string` | Kubernetes API server endpoint URL (non-native networking). |
| `private_endpoint` | `string` | Private native networking Kubernetes API server endpoint URL. |
| `public_endpoint` | `string` | Public native networking Kubernetes API server endpoint URL. Empty when the cluster endpoint is private-only. |

## Related Components

- [OciCompartment](/docs/catalog/oci/ocicompartment) — provides the compartment referenced by `compartmentId` via `valueFrom`
- [OciVcn](/docs/catalog/oci/ocivcn) — provides the VCN referenced by `vcnId` via `valueFrom`
- [OciSubnet](/docs/catalog/oci/ocisubnet) — provides subnets for the API endpoint (`endpointConfig.subnetId`) and service load balancers (`options.serviceLbSubnetIds`) via `valueFrom`
- [OciSecurityGroup](/docs/catalog/oci/ocisecuritygroup) — manages network security rules for the API endpoint (`endpointConfig.nsgIds`) and service load balancer backends (`options.serviceLbConfig.backendNsgIds`) via `valueFrom`
- [OciContainerEngineNodePool](/docs/catalog/oci/ocicontainerenginenodepool) — creates worker nodes for this cluster using the `cluster_id` output

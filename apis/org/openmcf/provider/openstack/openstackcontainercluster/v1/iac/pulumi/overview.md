# OpenStackContainerCluster Pulumi Module -- Architecture Overview

## Module Flow

```
OpenStackContainerClusterStackInput
  ├── target: OpenStackContainerCluster (api.proto)
  │   ├── metadata.name → cluster name
  │   └── spec: OpenStackContainerClusterSpec
  │       ├── cluster_template (required StringValueOrRef FK → OpenStackContainerClusterTemplate)
  │       ├── master_count (optional int32)
  │       ├── node_count (optional int32, updatable)
  │       ├── keypair (optional StringValueOrRef FK → OpenStackKeypair)
  │       ├── flavor (plain string, worker node flavor)
  │       ├── master_flavor (plain string, master node flavor)
  │       ├── docker_volume_size (optional int32)
  │       ├── labels (map<string,string>)
  │       ├── create_timeout (optional int32, minutes)
  │       ├── floating_ip_enabled (optional bool)
  │       └── region
  └── provider_config: OpenStackProviderConfig

         │
         ▼

   initializeLocals()
  ├── ClusterTemplate = spec.ClusterTemplate.GetValue()    [always]
  └── Keypair = spec.Keypair.GetValue()                     [if present]

         │
         ▼

   containerinfra.NewCluster()
  ├── Name = metadata.name
  ├── ClusterTemplateId = ClusterTemplate
  ├── MasterCount, NodeCount (if set)
  ├── Keypair (if non-empty)
  ├── Flavor, MasterFlavor (if non-empty)
  ├── DockerVolumeSize, Labels, CreateTimeout, FloatingIpEnabled
  ├── Region (if non-empty)
  └── Provider = openstackProvider

         │
         ▼

   Exports → stack_outputs.proto
  ├── cluster_id = resource ID
  ├── name = cluster name
  ├── api_address = K8s API endpoint
  ├── coe_version = container engine version
  ├── master_addresses = master node IPs
  ├── node_addresses = worker node IPs
  ├── kubeconfig_raw = full kubeconfig YAML (SECRET)
  ├── kubeconfig_host = API server URL
  ├── kubeconfig_cluster_ca_cert = CA certificate (SECRET)
  ├── kubeconfig_client_cert = client certificate (SECRET)
  ├── kubeconfig_client_key = client private key (SECRET)
  └── region = deployment region
```

## FK Resolution

| Field | Type | Resolution |
|-------|------|------------|
| `cluster_template` | Required FK | `spec.ClusterTemplate.GetValue()` -- always present |
| `keypair` | Optional FK | `spec.Keypair.GetValue()` -- nil-guarded, empty when not set |

## Resource Mapping

| Pulumi Resource | TF Equivalent | Count |
|-----------------|---------------|-------|
| `containerinfra.Cluster` | `openstack_containerinfra_cluster_v1` | 1 |

Single-resource module. No multi-resource pattern.

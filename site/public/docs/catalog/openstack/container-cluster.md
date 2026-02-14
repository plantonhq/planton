---
title: "Container Cluster"
description: "Container Cluster deployment documentation"
icon: "package"
order: 100
componentName: "openstackcontainercluster"
---

# OpenStack Container Cluster

Deploys an OpenStack Magnum container cluster that provisions a fully functional Kubernetes environment using a cluster template as a blueprint, managing master and worker nodes, networking, and kubeconfig credentials.

## What Gets Created

When you deploy an OpenStackContainerCluster resource, OpenMCF provisions:

- **Magnum Container Cluster** — a `containerinfra.Cluster` resource that creates a Kubernetes cluster using the specified cluster template. The cluster manages master and worker node instances, internal networking, and generates kubeconfig credentials for API access. Almost all fields are immutable after creation (ForceNew). Only `nodeCount` (triggers a scale operation) and `clusterTemplate` (triggers a cluster upgrade) can be updated in place.

## Prerequisites

- **OpenStack credentials** configured via environment variables or OpenMCF provider config
- **A cluster template** — either a template UUID or a reference to an OpenStackContainerClusterTemplate resource. The template defines the base image, network topology, key pair, and container orchestration engine (COE)
- **Sufficient Nova quota** for the requested number of master and worker nodes and their flavors
- **An SSH keypair** registered in OpenStack if overriding the keypair from the cluster template

## Quick Start

Create a file `container-cluster.yaml`:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackContainerCluster
metadata:
  name: my-cluster
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack.jobId: dev.OpenstackContainerCluster.my-cluster
    openmcf.org/stack.module.source: github.com/plantonhq/openmcf//apis/org/openmcf/provider/openstack/openstackcontainercluster/v1/iac/pulumi/module
spec:
  clusterTemplate: 7d4e8c2a-1f3b-4a5e-9c6d-0e8f1a2b3c4d
  nodeCount: 2
```

Deploy:

```shell
openmcf apply -f container-cluster.yaml
```

This creates a Magnum Kubernetes cluster using the specified cluster template with 2 worker nodes.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `clusterTemplate` | `StringValueOrRef` | The cluster template UUID to use as a blueprint. Changing this field triggers a cluster upgrade. Can reference an OpenStackContainerClusterTemplate resource via `valueFrom`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `masterCount` | `int32` | Magnum default (1) | The number of master nodes. Use 1 for development, 3 or more for production HA. ForceNew: changing this requires recreating the cluster. |
| `nodeCount` | `int32` | Magnum default (1) | The number of worker nodes. Updatable: changing this triggers a scale operation. |
| `keypair` | `StringValueOrRef` | template default | SSH keypair name for cluster node access. Overrides the keypair from the cluster template. Can reference an OpenStackKeypair resource via `valueFrom`. ForceNew. |
| `flavor` | `string` | template default | Nova flavor for worker nodes. Overrides the flavor from the cluster template. ForceNew. |
| `masterFlavor` | `string` | template default | Nova flavor for master nodes. Overrides the master flavor from the cluster template. ForceNew. |
| `dockerVolumeSize` | `int32` | template default | Size in GB of the Docker volume for each node. Overrides the value from the cluster template. ForceNew. |
| `labels` | `map<string, string>` | `{}` | Key-value labels for the cluster. Can extend or override the cluster template labels. Used for Kubernetes-specific settings such as `kube_tag`, `container_runtime`, and others. ForceNew. |
| `createTimeout` | `int32` | — | Timeout in minutes for cluster creation. If the cluster is not ready within this time, creation fails. ForceNew. |
| `floatingIpEnabled` | `bool` | template default | Whether to create a floating IP for every cluster node. Overrides the value from the cluster template. ForceNew. |
| `region` | `string` | provider default | Overrides the region from the provider config. ForceNew. |

## Examples

### Development Cluster

A minimal single-master, two-worker cluster for development and testing:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackContainerCluster
metadata:
  name: dev-cluster
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack.jobId: dev.OpenstackContainerCluster.dev-cluster
    openmcf.org/stack.module.source: github.com/plantonhq/openmcf//apis/org/openmcf/provider/openstack/openstackcontainercluster/v1/iac/pulumi/module
spec:
  clusterTemplate: 7d4e8c2a-1f3b-4a5e-9c6d-0e8f1a2b3c4d
  masterCount: 1
  nodeCount: 2
  flavor: m1.medium
  masterFlavor: m1.medium
  keypair: dev-keypair
```

### Production HA Cluster

A highly available cluster with 3 master nodes, dedicated flavors, larger Docker volumes, and a creation timeout:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackContainerCluster
metadata:
  name: prod-cluster
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack.jobId: prod.OpenstackContainerCluster.prod-cluster
    openmcf.org/stack.module.source: github.com/plantonhq/openmcf//apis/org/openmcf/provider/openstack/openstackcontainercluster/v1/iac/pulumi/module
spec:
  clusterTemplate: 7d4e8c2a-1f3b-4a5e-9c6d-0e8f1a2b3c4d
  masterCount: 3
  nodeCount: 5
  flavor: m1.xlarge
  masterFlavor: m1.large
  dockerVolumeSize: 100
  createTimeout: 60
  floatingIpEnabled: true
  labels:
    kube_tag: v1.28.4
    container_runtime: containerd
    auto_scaling_enabled: "true"
    min_node_count: "3"
    max_node_count: "10"
```

### Using Foreign Key References

Reference other OpenMCF-managed resources instead of hardcoding UUIDs. The `clusterTemplate` field resolves the template ID from a managed OpenStackContainerClusterTemplate, and the `keypair` field resolves the keypair name from a managed OpenStackKeypair:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackContainerCluster
metadata:
  name: ref-cluster
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack.jobId: staging.OpenstackContainerCluster.ref-cluster
    openmcf.org/stack.module.source: github.com/plantonhq/openmcf//apis/org/openmcf/provider/openstack/openstackcontainercluster/v1/iac/pulumi/module
spec:
  clusterTemplate:
    valueFrom:
      kind: OpenStackContainerClusterTemplate
      name: k8s-template
      field: status.outputs.template_id
  masterCount: 3
  nodeCount: 4
  flavor: m1.large
  masterFlavor: m1.large
  keypair:
    valueFrom:
      kind: OpenStackKeypair
      name: ops-keypair
      field: status.outputs.name
  dockerVolumeSize: 50
  floatingIpEnabled: false
```

### Cluster with Custom Kubernetes Labels

A cluster with custom labels to control Kubernetes version, runtime, and Magnum add-ons:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackContainerCluster
metadata:
  name: custom-cluster
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack.jobId: dev.OpenstackContainerCluster.custom-cluster
    openmcf.org/stack.module.source: github.com/plantonhq/openmcf//apis/org/openmcf/provider/openstack/openstackcontainercluster/v1/iac/pulumi/module
spec:
  clusterTemplate: 7d4e8c2a-1f3b-4a5e-9c6d-0e8f1a2b3c4d
  masterCount: 1
  nodeCount: 3
  flavor: m1.large
  dockerVolumeSize: 80
  createTimeout: 45
  labels:
    kube_tag: v1.29.1
    container_runtime: containerd
    cloud_provider_enabled: "true"
    cinder_csi_enabled: "true"
    monitoring_enabled: "true"
    ingress_controller: nginx
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Sensitive | Description |
|--------|------|-----------|-------------|
| `clusterId` | `string` | No | UUID of the created Magnum cluster |
| `name` | `string` | No | Name of the cluster, derived from `metadata.name` |
| `apiAddress` | `string` | No | Kubernetes API server endpoint URL (e.g., `https://10.0.0.5:6443`) |
| `coeVersion` | `string` | No | Version of the container orchestration engine (e.g., `v1.28.4`) |
| `masterAddresses` | `string[]` | No | List of IP addresses for master nodes |
| `nodeAddresses` | `string[]` | No | List of IP addresses for worker nodes |
| `kubeconfigRaw` | `string` | Yes | Full kubeconfig YAML containing all credentials needed for kubectl access |
| `kubeconfigHost` | `string` | No | Kubernetes API server URL extracted from the kubeconfig |
| `kubeconfigClusterCaCert` | `string` | Yes | Cluster CA certificate in PEM format for verifying the API server |
| `kubeconfigClientCert` | `string` | Yes | Client certificate in PEM format for API server authentication |
| `kubeconfigClientKey` | `string` | Yes | Client private key in PEM format for API server authentication |
| `region` | `string` | No | OpenStack region where the cluster was created |

## Related Components

- [OpenStack Container Cluster Template](/docs/catalog/openstack/openstackcontainerclustertemplate) — defines the blueprint (base image, network driver, COE) used by the cluster
- [OpenStack Keypair](/docs/catalog/openstack/openstackkeypair) — manages the SSH keypair for node access
- [OpenStack Network](/docs/catalog/openstack/openstacknetwork) — provides the network referenced in the cluster template
- [OpenStack Subnet](/docs/catalog/openstack/openstacksubnet) — provides the fixed subnet for cluster networking
- [OpenStack Load Balancer](/docs/catalog/openstack/openstackloadbalancer) — Magnum may provision load balancers for the Kubernetes API and services

---
title: "Container Cluster Template"
description: "Container Cluster Template deployment documentation"
icon: "package"
order: 100
componentName: "openstackcontainerclustertemplate"
---

# OpenStack Container Cluster Template

Deploys an OpenStack Magnum cluster template that serves as a reusable blueprint defining the base image, network topology, node flavors, container orchestration engine, and runtime settings for Kubernetes clusters. Cluster templates are referenced by OpenStackContainerCluster resources to create and configure clusters.

## What Gets Created

When you deploy an OpenStackContainerClusterTemplate resource, OpenMCF provisions:

- **Magnum Cluster Template** — a `containerinfra.ClusterTemplate` resource that defines the base OS image, container orchestration engine, node flavors, network driver, volume driver, and all related configuration. This template is then referenced by one or more OpenStackContainerCluster resources to create Kubernetes clusters with a consistent configuration. Almost all fields are updatable in place (PATCH-style) without recreating the template, with the exception of `region`.

## Prerequisites

- **OpenStack credentials** configured via environment variables or OpenMCF provider config
- **Magnum service** enabled and available in the target OpenStack deployment
- **A Glance image** suitable for Magnum (e.g., Fedora CoreOS, Ubuntu with Kubernetes support) registered or managed via OpenStackImage
- **A Nova flavor** for worker nodes and optionally a separate flavor for master nodes
- **An external (provider) network** if outbound connectivity is required for the cluster
- **An SSH keypair** registered in OpenStack if setting `keypair`

## Quick Start

Create a file `cluster-template.yaml`:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackContainerClusterTemplate
metadata:
  name: my-k8s-template
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack.jobId: dev.OpenstackContainerClusterTemplate.my-k8s-template
    openmcf.org/stack.module.source: github.com/plantonhq/openmcf//apis/org/openmcf/provider/openstack/openstackcontainerclustertemplate/v1/iac/pulumi/module
spec:
  coe: kubernetes
  image: fedora-coreos-39
  flavor: m1.medium
  masterFlavor: m1.large
  dnsNameserver: "8.8.8.8"
```

Deploy:

```shell
openmcf apply -f cluster-template.yaml
```

This creates a Magnum cluster template configured for Kubernetes with the `fedora-coreos-39` image, `m1.medium` worker flavor, and `m1.large` master flavor.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `coe` | `string` | Container Orchestration Engine for clusters created from this template. Magnum supports `kubernetes` (actively maintained), `swarm` (deprecated), and `mesos` (abandoned). In practice, `kubernetes` is the only production choice. | Minimum length 1 |
| `image` | `StringValueOrRef` | Base OS image for cluster nodes. Can be a literal image name or UUID (e.g., `fedora-coreos-39`). Can reference an OpenStackImage resource via `valueFrom`. | Required |

### Optional Fields

#### Node Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `flavor` | `string` | Magnum default | Nova flavor for worker nodes (e.g., `m1.medium`, `m1.xlarge`). |
| `masterFlavor` | `string` | Magnum default | Nova flavor for master (control plane) nodes (e.g., `m1.large`). |
| `dockerVolumeSize` | `int32` | Magnum default | Size in GB of the Docker volume attached to each node. Only applied when set to a value greater than 0. |
| `keypair` | `StringValueOrRef` | -- | SSH keypair name for node access. Can reference an OpenStackKeypair resource via `valueFrom`. |

#### Network Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `externalNetwork` | `StringValueOrRef` | -- | External (provider) network for outbound connectivity. Typically the pre-existing public network in the OpenStack deployment. Can reference an OpenStackNetwork resource via `valueFrom`. |
| `fixedNetwork` | `StringValueOrRef` | -- | Fixed (tenant) network for cluster nodes. Can reference an OpenStackNetwork resource via `valueFrom`. |
| `fixedSubnet` | `StringValueOrRef` | -- | Fixed subnet within the tenant network for cluster nodes. Can reference an OpenStackSubnet resource via `valueFrom`. |
| `networkDriver` | `string` | Magnum default | Container network driver. Common values: `flannel`, `calico`. |
| `volumeDriver` | `string` | Magnum default | Container volume driver. Common value: `cinder`. |
| `dnsNameserver` | `string` | -- | DNS nameserver address for cluster nodes (e.g., `8.8.8.8`). |
| `floatingIpEnabled` | `bool` | -- | When true, creates a floating IP for every cluster node. Required for external access to nodes on a private network. |

#### Cluster Behavior

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `masterLbEnabled` | `bool` | -- | When true, creates a load balancer in front of master nodes. Critical for HA Kubernetes clusters with multiple masters. |
| `tlsDisabled` | `bool` | -- | When true, disables TLS for the cluster API endpoint. Only use for testing environments; production clusters should always use TLS. |
| `labels` | `map<string, string>` | `{}` | Key-value labels passed to Magnum for Kubernetes-specific settings. Common keys: `kube_tag` (Kubernetes version, e.g., `v1.28.4`), `cloud_provider_tag` (cloud controller manager version), `container_runtime` (e.g., `containerd`), `ingress_controller` (ingress controller type). |

#### Provider Override

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `region` | `string` | provider default | Overrides the region from the provider config. ForceNew: changing this requires recreating the template. |

## Examples

### Basic Kubernetes Template

A minimal cluster template using only the required fields and a worker flavor:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackContainerClusterTemplate
metadata:
  name: basic-k8s
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack.jobId: dev.OpenstackContainerClusterTemplate.basic-k8s
    openmcf.org/stack.module.source: github.com/plantonhq/openmcf//apis/org/openmcf/provider/openstack/openstackcontainerclustertemplate/v1/iac/pulumi/module
spec:
  coe: kubernetes
  image: fedora-coreos-39
  flavor: m1.medium
  dnsNameserver: "8.8.8.8"
```

### Production-Ready HA Template

A template designed for production Kubernetes clusters with master load balancing, dedicated master and worker flavors, Calico networking, Cinder volumes, and a specific Kubernetes version:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackContainerClusterTemplate
metadata:
  name: prod-ha-k8s
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack.jobId: prod.OpenstackContainerClusterTemplate.prod-ha-k8s
    openmcf.org/stack.module.source: github.com/plantonhq/openmcf//apis/org/openmcf/provider/openstack/openstackcontainerclustertemplate/v1/iac/pulumi/module
spec:
  coe: kubernetes
  image: fedora-coreos-39
  keypair: ops-keypair
  externalNetwork: public
  fixedNetwork: 4a0e3c5b-2f1d-4e6a-8b9c-0d1e2f3a4b5c
  fixedSubnet: 7d8e9f0a-1b2c-3d4e-5f6a-7b8c9d0e1f2a
  networkDriver: calico
  volumeDriver: cinder
  dnsNameserver: "8.8.8.8"
  dockerVolumeSize: 100
  flavor: m1.xlarge
  masterFlavor: m1.large
  floatingIpEnabled: false
  masterLbEnabled: true
  tlsDisabled: false
  labels:
    kube_tag: v1.28.4
    cloud_provider_tag: v1.28.0
    container_runtime: containerd
    ingress_controller: nginx
```

### Template with Floating IPs for Development

A development template where every node gets a floating IP for direct SSH access, with TLS disabled for simpler local testing:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackContainerClusterTemplate
metadata:
  name: dev-k8s
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack.jobId: dev.OpenstackContainerClusterTemplate.dev-k8s
    openmcf.org/stack.module.source: github.com/plantonhq/openmcf//apis/org/openmcf/provider/openstack/openstackcontainerclustertemplate/v1/iac/pulumi/module
spec:
  coe: kubernetes
  image: fedora-coreos-39
  keypair: dev-keypair
  externalNetwork: public
  dnsNameserver: "8.8.8.8"
  flavor: m1.small
  masterFlavor: m1.medium
  floatingIpEnabled: true
  masterLbEnabled: false
  tlsDisabled: true
  labels:
    kube_tag: v1.28.4
    container_runtime: containerd
```

### Using Foreign Key References

Reference other OpenMCF-managed resources instead of hardcoding names and UUIDs:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackContainerClusterTemplate
metadata:
  name: ref-k8s-template
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack.jobId: prod.OpenstackContainerClusterTemplate.ref-k8s-template
    openmcf.org/stack.module.source: github.com/plantonhq/openmcf//apis/org/openmcf/provider/openstack/openstackcontainerclustertemplate/v1/iac/pulumi/module
spec:
  coe: kubernetes
  image:
    valueFrom:
      kind: OpenStackImage
      name: k8s-node-image
      field: status.outputs.image_id
  keypair:
    valueFrom:
      kind: OpenStackKeypair
      name: cluster-keypair
      field: status.outputs.name
  externalNetwork:
    valueFrom:
      kind: OpenStackNetwork
      name: public-net
      field: status.outputs.network_id
  fixedNetwork:
    valueFrom:
      kind: OpenStackNetwork
      name: cluster-network
      field: status.outputs.network_id
  fixedSubnet:
    valueFrom:
      kind: OpenStackSubnet
      name: cluster-subnet
      field: status.outputs.subnet_id
  networkDriver: calico
  volumeDriver: cinder
  dnsNameserver: "8.8.8.8"
  dockerVolumeSize: 80
  flavor: m1.xlarge
  masterFlavor: m1.large
  masterLbEnabled: true
  labels:
    kube_tag: v1.28.4
    container_runtime: containerd
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `template_id` | `string` | UUID of the created Magnum cluster template. This is the primary output used as a foreign key by OpenStackContainerCluster resources. |
| `name` | `string` | Name of the cluster template, derived from `metadata.name`. |
| `coe` | `string` | Container Orchestration Engine configured in the template (e.g., `kubernetes`). |
| `region` | `string` | OpenStack region where the template was created. |

## Related Components

- [OpenStackContainerCluster](/docs/catalog/openstack/container-cluster) — creates Kubernetes clusters using this template as a blueprint
- [OpenStackImage](/docs/catalog/openstack/image) — manages the base OS image referenced by the template
- [OpenStackKeypair](/docs/catalog/openstack/keypair) — manages the SSH keypair for node access
- [OpenStackNetwork](/docs/catalog/openstack/network) — provides the external or fixed network for cluster connectivity
- [OpenStackSubnet](/docs/catalog/openstack/subnet) — provides the fixed subnet for cluster nodes

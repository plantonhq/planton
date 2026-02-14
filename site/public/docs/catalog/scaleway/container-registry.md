---
title: "Container Registry"
description: "Container Registry deployment documentation"
icon: "package"
order: 100
componentName: "scalewaycontainerregistry"
---

# Scaleway Container Registry

Deploys a Scaleway Container Registry namespace, providing a fully managed, OCI-compliant registry for storing, managing, and deploying container images and Helm charts. Each namespace produces a Docker-compatible endpoint URL for push and pull operations.

## What Gets Created

When you deploy a ScalewayContainerRegistry resource, OpenMCF provisions:

- **Registry Namespace** — a `registry.Namespace` resource providing a dedicated OCI container image registry with a Docker endpoint at `rg.<region>.scw.cloud/<namespace-name>`

Container Registry namespaces are regional resources. The namespace name becomes part of the endpoint URL and must be unique within the Scaleway project.

## Prerequisites

- **Scaleway credentials** configured via environment variables or OpenMCF provider config
- **A unique namespace name** that is 4-63 characters, lowercase alphanumeric with hyphens, and DNS-compatible (it appears in the endpoint URL)

## Quick Start

Create a file `container-registry.yaml`:

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayContainerRegistry
metadata:
  name: my-registry
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.ScalewayContainerRegistry.my-registry
spec:
  region: fr-par
```

Deploy:

```shell
openmcf apply -f container-registry.yaml
```

This creates a private container registry namespace in the Paris region. After deployment, authenticate and push images:

```shell
docker login rg.fr-par.scw.cloud/my-registry -u nologin -p <SCW_SECRET_KEY>
docker tag myapp:latest rg.fr-par.scw.cloud/my-registry/myapp:latest
docker push rg.fr-par.scw.cloud/my-registry/myapp:latest
```

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Scaleway region for the registry namespace (e.g., `"fr-par"`, `"nl-ams"`, `"pl-waw"`). Determines the Docker endpoint URL. Cannot be changed after creation. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | `string` | `""` | Human-readable description of the registry namespace. Displayed in the Scaleway Console for identification. Has no effect on registry behavior. |
| `isPublic` | `bool` | `false` | When `true`, anyone can pull images without authentication. Pushing always requires authentication regardless of this setting. Can be changed after creation. |

## Examples

### Private Registry for Development

A minimal private registry namespace for storing development container images:

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayContainerRegistry
metadata:
  name: dev-images
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.ScalewayContainerRegistry.dev-images
spec:
  region: fr-par
  description: Development environment container images
```

### Public Registry for Open-Source Projects

A public registry namespace for distributing open-source container images and base images that external consumers can pull without credentials:

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayContainerRegistry
metadata:
  name: oss-images
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.ScalewayContainerRegistry.oss-images
spec:
  region: nl-ams
  description: Public base images and community tools
  isPublic: true
```

### Production Registry Co-Located with Kapsule Cluster

A private production registry in the same region as a Kapsule cluster for lowest image pull latency during deployments and pod scaling:

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayContainerRegistry
metadata:
  name: prod-services
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.ScalewayContainerRegistry.prod-services
spec:
  region: pl-waw
  description: Production microservices images for Warsaw Kapsule cluster
```

After deployment, create a Kubernetes image pull secret using the registry endpoint from stack outputs:

```shell
kubectl create secret docker-registry scw-registry \
  --docker-server=rg.pl-waw.scw.cloud/prod-services \
  --docker-username=nologin \
  --docker-password=<SCW_SECRET_KEY>
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `namespace_id` | `string` | Unique identifier of the registry namespace. Format: `"{region}/{uuid}"`. Used for Scaleway API and Terraform state references. |
| `endpoint` | `string` | Docker endpoint URL for login, push, and pull operations. Format: `"rg.<region>.scw.cloud/<namespace-name>"`. |
| `namespace_name` | `string` | Name of the registry namespace as it exists in Scaleway. Typically matches `metadata.name`. |
| `region` | `string` | Region where the registry namespace is deployed. Useful for downstream resources requiring region-aware configuration. |

## Related Components

- [ScalewayKapsuleCluster](/docs/catalog/scaleway/kapsule-cluster) — deploys Kubernetes clusters whose workloads pull images from this registry via imagePullSecrets
- [ScalewayPrivateNetwork](/docs/catalog/scaleway/private-network) — provides private connectivity for workloads accessing the registry

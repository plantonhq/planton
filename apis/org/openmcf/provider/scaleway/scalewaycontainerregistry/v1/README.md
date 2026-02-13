# Scaleway Container Registry

## Overview

The **ScalewayContainerRegistry** resource kind provides a declarative interface for creating and managing Container Registry namespaces on Scaleway. A registry namespace is an OCI-compliant container image store with a Docker-compatible endpoint, suitable for storing application images, base images, and Helm charts.

Scaleway Container Registry is region-scoped and produces a Docker endpoint URL that follows the pattern `rg.<region>.scw.cloud/<namespace-name>`. You can use any OCI-compatible client (Docker CLI, Podman, Buildah, Helm) to push and pull artifacts from the registry.

This is a **standalone resource** wrapping a single `scaleway_registry_namespace` Terraform resource. Image lifecycle policies and vulnerability scanning are managed via the Scaleway Console or API and can be added in future versions if Terraform support becomes available.

## Key Features

- **OCI-compliant** -- Supports Docker images, multi-architecture manifests, and Helm charts. Compatible with any OCI client.
- **Private by default** -- Images require authentication to pull. Public access is an explicit opt-in for open-source or shared images.
- **Regional deployment** -- Available in all Scaleway regions (fr-par, nl-ams, pl-waw). Choose the region closest to your build and deployment infrastructure.
- **Docker-native endpoint** -- Each namespace gets a dedicated Docker endpoint URL, ready for `docker login`, `docker push`, and `docker pull`.
- **No tags required** -- Scaleway Container Registry namespaces do not support tags in the API. The namespace name (from `metadata.name`) serves as the primary identifier. This is a Scaleway API limitation, not a design choice.

## Scaleway Terraform Resource Mapping

| OpenMCF Kind | Terraform Resource | Relationship |
|---|---|---|
| ScalewayContainerRegistry | `scaleway_registry_namespace` | 1:1 standalone |

This is a standalone (non-composite) resource wrapping a single Terraform resource.

## Spec Fields

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `region` | string | Yes | -- | Scaleway region (e.g., "fr-par", "nl-ams", "pl-waw"). Immutable after creation. |
| `description` | string | No | "" | Human-readable description of the registry namespace. |
| `is_public` | bool | No | false | Whether images can be pulled without authentication. |

## Stack Outputs

| Output | Description |
|---|---|
| `namespace_id` | Registry namespace ID (`{region}/{uuid}`). Primary cross-resource reference. |
| `endpoint` | Docker endpoint URL (`rg.<region>.scw.cloud/<namespace-name>`). Used for Docker login/push/pull. |
| `namespace_name` | Namespace name as it exists in Scaleway Container Registry. |
| `region` | Region where the registry namespace is deployed. |

## Dependencies

**Upstream:** None. ScalewayContainerRegistry is a standalone leaf resource with no `StringValueOrRef` inputs.

**Downstream:**
- **ScalewayServerlessFunction** and **ScalewayServerlessContainer** -- Reference `status.outputs.endpoint` for image source configuration.
- **ScalewayKapsuleCluster workloads** -- Use the endpoint in Kubernetes `imagePullSecrets` and Pod image references.
- **CI/CD pipelines** -- Use the endpoint as the Docker push target.

## Docker Usage Guide

### Authentication

Scaleway Container Registry uses your API secret key as the Docker password:

```bash
docker login rg.<region>.scw.cloud/<namespace-name> \
  -u nologin \
  -p <SCW_SECRET_KEY>
```

The username is always `nologin` -- Scaleway authenticates via the secret key, not a username/password pair.

### Push an Image

```bash
# Tag your local image with the registry endpoint
docker tag myapp:latest rg.fr-par.scw.cloud/my-registry/myapp:latest

# Push to Scaleway Container Registry
docker push rg.fr-par.scw.cloud/my-registry/myapp:latest
```

### Pull an Image

For private registries (default):
```bash
docker login rg.fr-par.scw.cloud/my-registry -u nologin -p <SECRET_KEY>
docker pull rg.fr-par.scw.cloud/my-registry/myapp:latest
```

For public registries:
```bash
# No authentication required
docker pull rg.fr-par.scw.cloud/my-registry/myapp:latest
```

### Kubernetes Integration

Create an `imagePullSecret` for private registries:

```bash
kubectl create secret docker-registry scw-registry \
  --docker-server=rg.fr-par.scw.cloud/my-registry \
  --docker-username=nologin \
  --docker-password=<SCW_SECRET_KEY>
```

Reference in your Pod spec:

```yaml
spec:
  imagePullSecrets:
    - name: scw-registry
  containers:
    - name: myapp
      image: rg.fr-par.scw.cloud/my-registry/myapp:latest
```

## Important Constraints

### Naming
- The namespace name becomes part of the Docker endpoint URL, so it must be DNS-compatible.
- Names must be unique within the Scaleway project.
- Changing the name requires recreating the namespace (and re-pushing all images).

### Immutability
- **Region** cannot be changed after creation. Migrating requires creating a new namespace in the target region and re-pushing images.
- **Name** changes force recreation (Scaleway treats it as a new namespace).
- **Visibility** (`is_public`) can be toggled safely after creation.

### No Cross-Region Replication
Scaleway does not support automatic cross-region registry replication. If you need images in multiple regions, create separate registries per region and push to each.

### Image Lifecycle
Image cleanup (deleting old tags, removing untagged images) is managed through the Scaleway Console or API, not through the Terraform provider. Set up retention policies in the Console for production registries.

### What's Not Included (Deferred)
- **Image lifecycle policies** -- Not available in the Terraform provider.
- **Cross-region replication** -- Not supported by Scaleway.
- **Vulnerability scanning configuration** -- Managed via Console/API.
- **Docker credentials resource** -- Authentication is handled at the client/CI level using API keys.

## Scaleway Documentation

- [Container Registry Overview](https://www.scaleway.com/en/docs/containers/container-registry/quickstart/)
- [Registry Authentication](https://www.scaleway.com/en/docs/containers/container-registry/how-to/connect-docker-cli/)
- [Terraform Resource: scaleway_registry_namespace](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/resources/registry_namespace)

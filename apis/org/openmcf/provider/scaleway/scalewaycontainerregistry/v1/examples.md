# Scaleway Container Registry Examples

This document provides YAML manifest examples for creating and managing Scaleway Container Registry namespaces using OpenMCF.

## Minimal Example: Private Registry

The simplest configuration creates a private registry namespace with just a name and region. Images require authentication to pull (default behavior).

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayContainerRegistry
metadata:
  name: my-app-registry
  org: my-org
  env: production
spec:
  region: fr-par
```

**What this creates:**
- A private registry namespace named `my-app-registry` in Paris
- Docker endpoint: `rg.fr-par.scw.cloud/my-app-registry`
- Images require authentication to pull

## Full Example: Public Registry with Description

A public registry for open-source projects or shared base images. Anyone can pull images without authentication.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayContainerRegistry
metadata:
  name: oss-images
  org: my-org
  env: shared
spec:
  region: nl-ams
  description: "Public base images and open-source project artifacts"
  is_public: true
```

**What this creates:**
- A public registry namespace named `oss-images` in Amsterdam
- Docker endpoint: `rg.nl-ams.scw.cloud/oss-images`
- Anyone can pull images without `docker login`
- Pushing still requires authentication

## Per-Environment Registries

Create isolated registries for each environment. This prevents staging images from accidentally being deployed to production.

### Staging

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayContainerRegistry
metadata:
  name: staging-registry
  org: my-org
  env: staging
spec:
  region: fr-par
  description: "Staging environment container images"
```

### Production

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayContainerRegistry
metadata:
  name: prod-registry
  org: my-org
  env: production
spec:
  region: fr-par
  description: "Production container images - promotion only"
```

## Infra Chart Integration: valueFrom References

When composing ScalewayContainerRegistry with other resources in infra charts, downstream resources can reference the registry's outputs using `valueFrom`.

### Serverless Container Using Registry Endpoint

A ScalewayServerlessContainer referencing this registry's endpoint to pull its image:

```yaml
# In an infra chart template:
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayServerlessContainer
metadata:
  name: "{{ values.service_name }}"
  org: "{{ values.org }}"
  env: "{{ values.env }}"
spec:
  region: "{{ values.region }}"
  registry_endpoint:
    valueFrom:
      kind: ScalewayContainerRegistry
      name: "{{ values.registry_name }}"
      fieldPath: status.outputs.endpoint
  image_name: "{{ values.image_name }}"
  image_tag: "{{ values.image_tag }}"
```

### CI/CD Pipeline Configuration

Use the registry endpoint output in CI/CD pipelines for Docker push targets:

```yaml
# In a CI/CD pipeline definition (pseudo-code):
steps:
  - name: docker-push
    env:
      REGISTRY_ENDPOINT:
        valueFrom:
          kind: ScalewayContainerRegistry
          name: "{{ values.registry_name }}"
          fieldPath: status.outputs.endpoint
    command: |
      docker tag $IMAGE $REGISTRY_ENDPOINT/$IMAGE
      docker push $REGISTRY_ENDPOINT/$IMAGE
```

## Multi-Region Deployment

For teams that deploy across multiple Scaleway regions and want local image pulls for faster cold starts:

```yaml
# Paris region registry
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayContainerRegistry
metadata:
  name: app-images-par
  org: my-org
  env: production
spec:
  region: fr-par
  description: "Production images - Paris region"

---

# Amsterdam region registry
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayContainerRegistry
metadata:
  name: app-images-ams
  org: my-org
  env: production
spec:
  region: nl-ams
  description: "Production images - Amsterdam region"
```

**Note:** Scaleway does not support cross-region replication. Images must be pushed to each registry independently. CI/CD pipelines should push to all region registries as part of the release process.

## After Deployment

Once the registry is created, use the outputs to configure your tools:

```bash
# Get the Docker endpoint from stack outputs
ENDPOINT=$(openmcf stack-outputs --manifest registry.yaml --field endpoint)

# Authenticate with the registry
docker login $ENDPOINT -u nologin -p $SCW_SECRET_KEY

# Tag and push an image
docker tag myapp:latest $ENDPOINT/myapp:v1.0.0
docker push $ENDPOINT/myapp:v1.0.0

# Pull the image (from any authenticated client)
docker pull $ENDPOINT/myapp:v1.0.0
```

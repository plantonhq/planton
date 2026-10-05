# Professional Container Registry

This preset creates a DigitalOcean Container Registry (DOCR) with the professional tier and 30-day push credentials. The professional tier provides the highest storage limits and is suitable for production teams pushing many images.

## When to Use

- Production teams pushing container images for Kubernetes or App Platform
- CI/CD pipelines building and pushing images frequently
- Need for larger storage than starter/basic tiers

## Key Configuration Choices

- **Professional tier** (`subscriptionTier: professional`) -- highest storage and bandwidth limits; production-ready. The tier is the one setting you can change later.
- **Push credentials** (`dockerCredentials`) -- `write: true` mints a credential your CI/CD can push with; `expirySeconds: 2592000` keeps its lifetime to 30 days (rotate by re-applying). Remove the block entirely if nothing needs a minted credential -- no block, no long-lived token.
- **Region** (`region: nyc3`) -- registry data location; choose nearest to your DOKS clusters or pipelines. Create-only.
- **Registry name** (`name`) -- globally unique across ALL DigitalOcean accounts; used in image paths (`registry.digitalocean.com/<name>/<image>:<tag>`). One registry per account.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `nyc3` | Target DigitalOcean region slug | [DigitalOcean Regions API](https://docs.digitalocean.com/reference/api/api-reference/#tag/Regions) |
| `my-registry` | Globally unique registry name (1-63 chars, lowercase, hyphens) | Choose a unique name; used in `docker push` and `imagePullSecret` |

## Related Presets

- None for this kind; consider `DigitalOceanKubernetesCluster` with `registryIntegration: true` for seamless image pull

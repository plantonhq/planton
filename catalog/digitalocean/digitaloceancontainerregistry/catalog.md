# DigitalOcean Container Registry

Deploys a private, OCI-compliant container registry on DigitalOcean for storing Docker images and Helm charts. Configures the registry name, subscription tier, and region, optionally mints Docker credentials with a controlled lifetime, and exposes the endpoint as an output for downstream workloads to pull images. A DigitalOcean account holds exactly ONE registry, and name and region are create-only -- the subscription tier is the only setting that can change after creation.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Container Registry** -- a `digitalocean_container_registry` resource with the specified name, subscription tier, and region
- **Docker Credentials** -- created only when `dockerCredentials` is set: a `digitalocean_container_registry_docker_credentials` resource minting a read-only or write credential with the configured expiry, exported through the `docker_credentials` output (a secret)

DigitalOcean restricts each account to a single container registry, and registry names are globally unique across ALL DigitalOcean accounts. Deploying a second DigitalOceanContainerRegistry resource on the same account will fail.

## Before You Deploy

### Planton Setup

- **DigitalOcean Provider Connection** -- an active connection in the Connect module with a DigitalOcean API token. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline API token authentication.

### DigitalOcean Account

- **No existing container registry** on the target DigitalOcean account. DigitalOcean allows only one registry per account.
- **A supported region** for container registry storage (e.g., `nyc3`, `sfo3`, `fra1`, `sgp1`). Choose the region nearest to your Kubernetes clusters or CI/CD pipelines for lowest pull latency.

## Deploy

### Console

Open the deployment store, find **DigitalOcean Container Registry**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Professional Container Registry** preset in the [Presets](#presets) tab for a production-ready configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanContainerRegistry
metadata:
  name: prod-registry
  org: acme-corp
  env: prod
spec:
  name: prod-registry
  subscriptionTier: professional
  region: nyc3
```

```shell
planton apply -f registry.yaml
```

This creates a professional-tier container registry in NYC3, with images addressable at `registry.digitalocean.com/prod-registry/<repository>:<tag>`. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring a container registry. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Subscription tier** -- The `subscriptionTier` field (`starter`, `basic`, or `professional`, in ascending storage capacity) is the ONLY setting that can change after creation -- everything else on the registry is create-only. Moving up a tier is live and never touches stored images; a downgrade fails if stored images exceed the smaller tier's ceiling, so garbage-collect untagged images first (an on-demand DigitalOcean action -- nothing in this manifest schedules it).

**Region** -- The `region` field determines where registry data is stored; omit it to let DigitalOcean choose (the chosen slug is reported through the `region` output). Choose the region closest to your DigitalOcean Kubernetes clusters or build pipelines to minimize image pull latency. The region cannot change after creation.

**Docker credentials** -- Set `dockerCredentials` to mint a registry credential: `write: true` for push access (default is read-only pull), and `expirySeconds` for a controlled lifetime (unset means the API maximum, roughly 50 years -- effectively forever, and revocable only by deleting the credential). No block means no long-lived token -- the secure default. Neither knob is recoverable from the API afterwards: `write` and `expirySeconds` exist only in your manifest and provisioner state, so keep the manifest authoritative and never hand-mint credentials in the control panel alongside it.

**Registry naming** -- The `name` field is globally unique across ALL DigitalOcean accounts and is used in image paths (`registry.digitalocean.com/<name>/<image>:<tag>`). Choose a stable name: it is the registry's resource identity, and changing it replaces the registry (all images gone).

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `registry_name` | Name of the created container registry (also its resource identifier) | Image path construction, Kubernetes `imagePullSecret` configuration |
| `server_url` | The registry host, always `registry.digitalocean.com` | Docker login |
| `endpoint` | Full registry endpoint (e.g., `registry.digitalocean.com/prod-registry`) | CI/CD pipeline push targets, image path construction |
| `region` | Region slug where the registry is hosted (reported by DigitalOcean) | Verifying data locality alignment with clusters |
| `docker_credentials` | Base64 Docker `config.json` (a secret); empty when credentials are not configured | Kubernetes `imagePullSecret`, CI/CD docker login |
| `credential_expiration_time` | RFC 3339 expiry of the minted credential | Rotation monitoring |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Professional registry for production** -- Professional tier with 30-day push credentials, suitable for teams pushing images frequently from CI/CD pipelines. Provides the highest storage limits. Start from the **Professional Container Registry** preset.

**One registry, many consumers** -- Because the account holds exactly one registry, treat this component as account-level shared infrastructure, like a VPC: one owner, one manifest, and everyone else consumes its outputs. Split environments through repository naming (`registry.digitalocean.com/acme/staging-api`), never through a second registry -- a second deployment on the same account fails.

**Pull/push credential split** -- Mint the `write: true` credential here for the build pipeline alone, and give everything that RUNS images read-only pull access distributed separately. DigitalOcean Kubernetes clusters can integrate with the registry natively, without this credential at all.

## Works With

- [**DigitalOcean Kubernetes Cluster**](/infra-catalog/digital-ocean-kubernetes-cluster) -- its `registryIntegration` toggle lets cluster workloads pull from this registry account-wide, with no image pull secret
- [**DigitalOcean App Platform App**](/infra-catalog/digital-ocean-app) -- services, workers, and jobs deploy images from this registry via the `docr` image source, using the account's registry access
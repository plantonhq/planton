# DigitalOcean App Platform App

Deploys a full App Platform application -- HTTP services, workers, jobs, static sites, functions, and in-app databases -- from Git or from a container image. One manifest declares every component with its source, sizing, and autoscaling, plus app-level domains, ingress rules, alerts, and VPC egress placement; App Platform builds and serves the result with automatic HTTPS on a default `ondigitalocean.app` hostname. The decisions that matter most: which source type the DigitalOcean account can actually use, and the two fields that still deploy only through Terraform at the current Pulumi SDK.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **App Platform Application** -- a `digitalocean_app` resource whose spec lists every component you declared
- **HTTP services** -- created from `spec.services`; receive traffic with automatic HTTPS
- **Workers** -- created from `spec.workers`; long-running processes without HTTP ingress
- **Jobs** -- created from `spec.jobs`; run around a deployment (`pre_deploy`, `post_deploy`, or `failed_deploy`)
- **Static sites, functions, in-app databases** -- created from the matching spec lists when present
- **Domains and ingress** -- created from `spec.domains` and `spec.ingress` when present
- **Environment variables** -- from `spec.envs` and per-component `envs`; `secret` values are stored in App Platform's secret store

## Before You Deploy

### Planton Setup

- **DigitalOcean Provider Connection** -- an active connection in the Connect module with a DigitalOcean API token. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline API token authentication.

### DigitalOcean Account

- **A source** for each component: a public Git clone URL, a linked GitHub/GitLab/Bitbucket repo, or a container image (Docker Hub, GHCR, or DigitalOcean Container Registry). The `github`/`gitlab`/`bitbucket` sources need the matching connection in the DigitalOcean control panel; `git` with a public clone URL needs none.
- **An App Platform region group** (for example `nyc`, `sfo`, `fra` -- these are datacenter groups, not droplet slugs like `nyc3`).
- **A DigitalOcean-managed DNS zone** (only for custom domains) -- `domains[].zone` expects the zone to already exist; App Platform will not create DNS for you.

## Deploy

### Console

Open the deployment store, find **DigitalOcean App Platform App**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Git Source Web App** preset in the [Presets](#presets) tab to build from a repository, or **Container Image App** to run a pre-built image with autoscaling.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanApp
metadata:
  name: my-web-app
  org: acme-corp
  env: prod
spec:
  appName: my-web-app
  region: nyc
  services:
    - name: web
      git:
        repoCloneUrl: https://github.com/digitalocean/sample-nodejs.git
        branch: main
      instanceSizeSlug: basic-xxs
      instanceCount: 1
```

```shell
planton apply -f app.yaml
```

This creates a single-instance web app in NYC3 built from the sample Node.js repository, served with automatic HTTPS on its default hostname. A Stack Job tracks the provisioning in real time.

### InfraChart

When wiring VPC egress placement, a custom domain on a DigitalOcean-managed zone, or a production database attachment, use ValueFromRef to reference resources deployed in the same InfraPipeline:

```yaml
spec:
  vpc:
    valueFrom:
      kind: DigitalOceanVpc
      name: app-network
      fieldPath: status.outputs.vpc_id
  domains:
    - name: www.example.com
      zone:
        valueFrom:
          kind: DigitalOceanDnsZone
          name: example-com
          fieldPath: status.outputs.zone_name
  databases:
    - name: app-db
      engine: pg
      production: true
      clusterName:
        valueFrom:
          kind: DigitalOceanDatabaseCluster
          name: app-postgres
          fieldPath: spec.cluster_name
```

The InfraPipeline resolves the dependency graph, deploys the VPC, DNS zone, and database cluster first, then provisions the app with the resolved values.

## Key Configuration

These are the most important decisions when configuring an App Platform application. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Components** -- add at least one of `services`, `workers`, `jobs`, `staticSites`, or `functions`; validation rejects an empty app before any provisioner runs. Jobs run around a deployment and never autoscale; static sites build from Git only (no image source).

**Source per component** -- each component sets exactly one of `git`, `github`, `gitlab`, `bitbucket`, or (services/workers/jobs) `image`. The VCS-linked sources need the matching connection in the DigitalOcean control panel -- `deployOnPush` is silently ignored without it and a missing connection fails the deploy -- so `git` with a public clone URL is the right default for unlinked accounts. For Docker Hub images, `registry` is the namespace (`library` for official images), not the string "docker-hub"; for DigitalOcean Container Registry set `registryType: docr` and leave `registry` empty.

**App name** -- `appName` is 2-32 characters matching `^[a-z][a-z0-9-]{0,30}[a-z0-9]$` (starts with a letter) and unique across every app in the account; component names follow the same rule and validation enforces it on each. Renaming updates the app in place, but the default `<name>-<hash>.ondigitalocean.app` URL changes with the name.

**Instance size and count** -- `instanceSizeSlug` is a free-form string (`basic-xxs`, `professional-s`, ...); new provider sizes work without a catalog change. The slug is the cost driver: every instance of every service, worker, and job bills by its size, so a forgotten `professional` slug on a staging worker costs real money.

**Autoscaling** -- set `autoscaling` on a service or worker and leave `instanceCount` unset; the spec rejects the combination because App Platform ignores a fixed count while autoscaling is on. Jobs cannot autoscale.

**Termination** -- `termination.drainSeconds` is a service-only HTTP connection drain; workers and jobs reject it and honor `gracePeriodSeconds` only.

**Liveness and secure headers** -- service/worker `livenessHealthCheck` is the restart probe (the service `healthCheck` is the readiness probe that gates traffic), and `ingress.secureHeader` adds one response header to every route. Both deploy on both provisioners, as does every other spec field including `vpc`, `maintenance`, ingress `authorityExact` matches, and alert destinations.

**Alert destinations** -- both engines wire email and Slack destinations, but the provider never reads them back, so a Terraform apply with destinations set re-plans and redeploys the app every time (a provider defect). Set them on Pulumi stacks, or leave them unset on Terraform and manage recipients in the control panel; recipients must be verified team members.

**Project** -- `projectId` is create-only: changing it destroys and recreates the app. Leave it unset to use the account's default project.

**In-app databases** -- a `databases[]` entry without `clusterName` is App Platform's managed dev database, not a production data store. Set `production: true` with `clusterName` referencing an existing cluster; the attachment never creates the cluster.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **DigitalOceanVpc** (optional) | `vpc` | `status.outputs.vpc_id` |
| **DigitalOceanDnsZone** (optional) | `domains[].zone` | `status.outputs.zone_name` |
| **DigitalOceanDatabaseCluster** (optional) | `databases[].clusterName` | `spec.cluster_name` |

VPC placement is wired by both provisioners.

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `app_id` | App Platform application UUID | Import, API operations |
| `default_hostname` | Default `ondigitalocean.app` hostname | DNS CNAME targets |
| `live_url` | Public URL including protocol | Application integrations |
| `live_domain` | Live hostname without scheme | Certificates, DNS |
| `active_deployment_id` | Currently live deployment UUID | Post-deploy verification that a new deployment went live |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Git source web app** -- one HTTP service built from a public repository; App Platform detects the language and builds it, no VCS connection required. The shape for getting an app live before wiring GitHub. Start from the **Git Source Web App** preset.

**Container image app** -- one HTTP service from Docker Hub, GHCR, or DOCR with CPU autoscaling; skips the build entirely, so deploys are as fast as an image pull and the running artifact is exactly what CI produced. Start from the **Container Image App** preset.

## Works With

- [**DigitalOcean VPC**](/cloud-catalog/digital-ocean-vpc) -- optional egress placement (`spec.vpc`)
- [**DigitalOcean DNS Zone**](/cloud-catalog/digital-ocean-dns-zone) -- custom domains (`spec.domains`)
- [**DigitalOcean Database Cluster**](/cloud-catalog/digital-ocean-database-cluster) -- in-app database `clusterName`
- [**DigitalOcean Function**](/cloud-catalog/digital-ocean-function) -- standalone functions app when the functions component should not share this app

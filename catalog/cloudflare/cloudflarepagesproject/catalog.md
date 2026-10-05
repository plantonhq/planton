# Cloudflare Pages Project

Provisions a Cloudflare Pages project: a managed site host that builds and serves a static site or full-stack app (static assets plus Pages Functions) from Cloudflare's edge. This manages the durable **project** -- its build configuration, optional git connection, per-environment runtime configuration (bindings, env vars, compatibility), and custom domains. Deployments themselves are produced out-of-band (a git push for connected projects, or `wrangler pages deploy` for direct upload).

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Pages Project** -- the managed site project and its `*.pages.dev` subdomain
- **Build configuration** -- how Cloudflare builds the site (for git-connected and wrangler builds)
- **Per-environment runtime config** -- bindings, variables, and secrets for preview and production
- **Custom domains** -- any hostnames you attach (Cloudflare provisions their certificates)

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** -- an active connection in the Connect module with a Cloudflare API token that has Pages edit access. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline API token authentication.

### Cloudflare Account

- **An account** -- the project is created under an existing Cloudflare account.
- **Git authorization (git-connected projects only)** -- the account must already be authorized with the git provider (the GitHub App install or GitLab OAuth), a one-time browser step in the Cloudflare dashboard.

## Deploy

### Console

Open the deployment store, find **Cloudflare Pages Project**, and click **Deploy**. The creation wizard captures the account, project name, build configuration, optional git source, the production and preview deployment configs (bindings, variables, secrets), and custom domains. Start from the **Git-Connected Site** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflarePagesProject
metadata:
  name: marketing-site
  org: acme-corp
  env: prod
spec:
  accountId: a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4
  name: marketing-site
  productionBranch: main
  buildConfig:
    buildCommand: npm run build
    destinationDir: dist
  source:
    type: github
    config:
      owner: acme-corp
      repoName: marketing-site
  deploymentConfigs:
    production:
      kvNamespaces:
        - name: CACHE
          namespaceId:
            valueFrom:
              kind: CloudflareKvNamespace
              name: prod-cache
              fieldPath: status.outputs.namespace_id
  domains:
    - www.example.com
```

```shell
planton apply -f cloudflare-pages-project.yaml
```

This creates a git-connected project that builds on every push to `main`, binds a KV namespace to production Functions, and serves on a custom domain. An Infra Job tracks the provisioning in real time.

### InfraChart

Deploy the project together with the resources its Functions bind, wiring each binding with ValueFromRef:

```yaml
spec:
  deploymentConfigs:
    production:
      kvNamespaces:
        - name: CACHE
          namespaceId:
            valueFrom:
              kind: CloudflareKvNamespace
              name: prod-cache
              fieldPath: status.outputs.namespace_id
      r2Buckets:
        - name: ASSETS
          bucketName:
            valueFrom:
              kind: CloudflareR2Bucket
              name: prod-assets
              fieldPath: status.outputs.bucket_name
```

The InfraPipeline resolves the dependency graph, provisions the KV namespace and R2 bucket first, then creates the project with the resolved binding IDs.

## Key Configuration

These are the most important decisions when configuring a Pages project. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Account (`accountId`)** and **Name (`name`)** -- both **immutable**; changing either replaces the project. The name is also the `<name>.pages.dev` subdomain label.

**Build (`buildConfig`)** -- the build command and output directory. Omit for a pre-built direct-upload project.

**Git Source (`source`)** -- connect a GitHub or GitLab repository so Cloudflare builds on every push, or leave it empty and deploy with `wrangler pages deploy`.

**Deployment Configs (`deploymentConfigs.production` / `.preview`)** -- per-environment runtime config: compatibility, environment variables, secrets, and bindings to KV namespaces, D1 databases, R2 buckets, queues, Hyperdrive configs, and other Workers. Bind preview to non-production resources to isolate test traffic.

**Domains (`domains`)** -- custom hostnames in zones on this account; the `*.pages.dev` subdomain is always available.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **CloudflareKvNamespace** | `deploymentConfigs.*.kvNamespaces[].namespaceId` | `status.outputs.namespace_id` |
| **CloudflareD1Database** | `deploymentConfigs.*.d1Databases[].databaseId` | `status.outputs.database_id` |
| **CloudflareR2Bucket** | `deploymentConfigs.*.r2Buckets[].bucketName` | `status.outputs.bucket_name` |
| **CloudflareQueue** | `deploymentConfigs.*.queueProducers[].queueName` | `status.outputs.queue_name` |
| **CloudflareHyperdriveConfig** | `deploymentConfigs.*.hyperdriveBindings[].configId` | `status.outputs.hyperdrive_id` |
| **CloudflareWorker** | `deploymentConfigs.*.services[].service` | `status.outputs.script_name` |

Secret values in `deploymentConfigs.*.secrets` reference managed secrets resolved just-in-time at deploy.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `subdomain` | The project's `*.pages.dev` subdomain | CNAME target for a CloudflareDnsRecord fronting the site |

`status.outputs` also echoes `project_name`, `domains`, and `created_on`, but those are the values you set in the spec (or provisioning metadata) rather than consumable dependencies. Per-deployment URLs are never exported -- deployments are produced out-of-band, so they do not exist at provision time.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Git-connected site** -- Cloudflare is the CI: it builds and deploys on every push to the production branch, and other branches get preview deployments. Requires the one-time git authorization in the Cloudflare dashboard. Start from the **Git-Connected Site** preset.

**Direct upload** -- deploy pre-built output with `wrangler pages deploy` from your own CI; no build config or git connection. Choose this when the build must run in a pipeline you control. Start from the **Direct-Upload Site** preset.

## Works With

- [**Cloudflare KV Namespace**](/infra-catalog/cloudflare-kv-namespace) -- bound to Functions for edge key-value storage
- [**Cloudflare D1 Database**](/infra-catalog/cloudflare-d1-database) -- bound to Functions for serverless SQL
- [**Cloudflare R2 Bucket**](/infra-catalog/cloudflare-r2-bucket) -- bound to Functions for object storage
- [**Cloudflare Queue**](/infra-catalog/cloudflare-queue) -- bound to Functions as a producer
- [**Cloudflare Hyperdrive Config**](/infra-catalog/cloudflare-hyperdrive-config) -- bound to Functions for pooled SQL access
- [**Cloudflare Worker**](/infra-catalog/cloudflare-worker) -- bound to Functions for service-to-service calls

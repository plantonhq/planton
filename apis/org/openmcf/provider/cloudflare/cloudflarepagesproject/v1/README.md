# CloudflarePagesProject

Deploy a Cloudflare Pages project — a managed host for a static site or
full-stack app (static assets + Pages Functions) served from Cloudflare's edge.

This component manages the durable **project**: its build configuration, optional
git connection, per-environment runtime configuration (bindings, env vars,
compatibility), and custom domains. The actual **deployments** (the built
versions of the site) are produced out-of-band — see "How versions are deployed".

## Two ways to deploy versions

Cloudflare Pages has no deployment resource in the Terraform/Pulumi provider, so
this component never creates a deployment. New versions land one of two ways:

- **Git-connected** (set `source`): Cloudflare connects to your repository and
  builds a new deployment on every push — Cloudflare is the CI.
- **Direct upload** (omit `source`): you build the site yourself and push it with
  `wrangler pages deploy ./dist --project-name=<name>`.

> Git-connected prerequisite: authorize Cloudflare with your git provider once in
> the Cloudflare dashboard (the GitHub App install / GitLab OAuth). The provider
> manages which repo/branch/build to use, not that one-time authorization.

## Quick start (direct upload)

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflarePagesProject
metadata:
  name: marketing-site
spec:
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  name: marketing-site
  productionBranch: main
  buildConfig:
    buildCommand: npm run build
    destinationDir: dist
```

## Git connection

```yaml
  source:
    type: github          # or gitlab
    config:
      owner: acme
      repoName: marketing-site
      productionBranch: main
      prCommentsEnabled: true
      previewDeploymentSetting: all   # all | none | custom
```

## Per-environment configuration

`deploymentConfigs.preview` and `deploymentConfigs.production` share the same
shape (bindings, env vars, compatibility, limits, placement). **If you set only
one environment, the same configuration is applied to both** — Cloudflare treats
preview and production as a paired configuration and rejects inconsistent
environments. Set both explicitly to differ them.

```yaml
  deploymentConfigs:
    production:
      compatibilityDate: "2025-01-15"
      compatibilityFlags: [nodejs_compat]
      vars:
        LOG_LEVEL: info
      secrets:
        - name: API_KEY
          value: <managed-secret-reference>
      kvNamespaces:
        - name: CONFIG
          namespaceId:
            valueFrom: { kind: CloudflareKvNamespace, name: app-config, fieldPath: status.outputs.namespace_id }
      d1Databases:
        - name: DB
          databaseId:
            valueFrom: { kind: CloudflareD1Database, name: app-db, fieldPath: status.outputs.database_id }
```

Binding groups (each accepts a literal id or a `valueFrom` reference):

| Group | Binds | Reference kind |
|---|---|---|
| `vars` | plain-text variables (map) | — |
| `secrets` | secret values (managed-secret, JIT-resolved) | — |
| `kvNamespaces` | KV namespaces | CloudflareKvNamespace |
| `d1Databases` | D1 databases | CloudflareD1Database |
| `r2Buckets` | R2 buckets (+ optional jurisdiction) | CloudflareR2Bucket |
| `queueProducers` | Queue producers | CloudflareQueue |
| `hyperdriveBindings` | Hyperdrive configs | CloudflareHyperdriveConfig |
| `services` | other Workers | CloudflareWorker |
| `durableObjectNamespaces` | Durable Object namespaces | — |
| `analyticsEngineDatasets` | Analytics Engine datasets | — |
| `vectorizeBindings` | Vectorize indexes | — |
| `aiBindings` | Constellation/AI projects | — |
| `mtlsCertificates` | mTLS certificates | — |
| `browsers` | Browser Rendering | — |

## Custom domains

```yaml
  domains:
    - www.example.com   # must be a hostname in a zone on this account
```

## Outputs

| Output | Description |
|---|---|
| `project_name` | The project name (target for service/asset references) |
| `subdomain` | The `*.pages.dev` subdomain (e.g. `marketing-site.pages.dev`) |
| `domains` | Custom domains attached to the project |
| `created_on` | Project creation timestamp |

Deployment-specific values (per-build URLs, ids) are not exported: they do not
exist until the first deployment is produced out-of-band.

## Secrets

`deploymentConfigs.*.secrets[].value` and `buildConfig.webAnalyticsToken` are
secret-by-default: provide a managed-secret reference, resolved just-in-time at
deploy. Plain configuration belongs in `vars`.

## Related components

- `CloudflareWorker` (with Static Assets) — the build-and-upload hosting model.
- `CloudflareKvNamespace`, `CloudflareD1Database`, `CloudflareR2Bucket`,
  `CloudflareQueue`, `CloudflareHyperdriveConfig`, `CloudflareDnsZone`.

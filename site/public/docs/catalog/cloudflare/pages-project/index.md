---
title: "Pages Project"
description: "Pages Project deployment documentation"
icon: "package"
order: 100
componentName: "cloudflarepagesproject"
---

# Cloudflare Pages Project

Host a static site or full-stack app (static assets + Pages Functions) on
Cloudflare's edge, with a connected git repository for automatic builds or
direct uploads of a pre-built site.

## What Gets Created

- A `cloudflare_pages_project` (the project), with its build configuration,
  optional git source, and per-environment deployment configuration (bindings,
  env vars, compatibility, limits).
- One `cloudflare_pages_domain` per attached custom domain.

## Prerequisites

- A Cloudflare account ID.
- For git-connected projects: a one-time git-provider authorization in the
  Cloudflare dashboard (GitHub App install / GitLab OAuth). The provider manages
  the source configuration, not that authorization.

## Configuration Reference

**Required**

- `accountId` — Cloudflare account ID.
- `name` — project name (lowercase, also the `*.pages.dev` subdomain; immutable).
- `productionBranch` — the production branch (e.g. `main`).

**Optional**

- `buildConfig` — `buildCommand`, `destinationDir`, `rootDir`, `buildCaching`,
  `webAnalyticsTag`, `webAnalyticsToken` (secret).
- `source` — git connection: `type` (`github`/`gitlab`) + `config`
  (`owner`, `repoName`, `productionBranch`, `prCommentsEnabled`,
  `previewDeploymentSetting`, branch/path include/exclude lists). Omit for a
  direct-upload project.
- `deploymentConfigs` — `preview` and `production` runtime config (bindings, env
  vars, compatibility, limits, placement). Set one to apply it to both, or both
  to differ them.
- `domains` — custom hostnames (each in a zone on this account).

Cross-resource bindings accept either a literal id or a `valueFrom` reference to
the producing resource (KV/D1/R2/Queue/Hyperdrive/Worker).

## How Versions Are Deployed

The provider has no deployment resource. New versions are produced out-of-band:
git push (Cloudflare builds) for git-connected projects, or
`wrangler pages deploy` for direct-upload projects.

## Stack Outputs

| Output | Description |
|---|---|
| `project_name` | The project name |
| `subdomain` | The `*.pages.dev` subdomain |
| `domains` | Attached custom domains |
| `created_on` | Project creation timestamp |

## Related Components

- CloudflareWorker (Static Assets — the build-and-upload hosting model)
- CloudflareKvNamespace, CloudflareD1Database, CloudflareR2Bucket
- CloudflareQueue, CloudflareHyperdriveConfig, CloudflareDnsZone

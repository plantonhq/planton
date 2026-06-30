# Deploying Cloudflare Pages Projects

## What Cloudflare Pages is

Cloudflare Pages is a managed host for static sites and full-stack apps. A
**project** is the durable container — a name, a production branch, build
configuration, an optional git connection, per-environment runtime configuration,
and custom domains. A **deployment** is one immutable built version of the site;
each gets its own preview URL, and production-branch deployments serve the
production domain. The project always exposes a `*.pages.dev` subdomain.

Pages overlaps with Cloudflare Workers Static Assets (a Worker that serves a
built asset directory). The two represent the two canonical hosting models:

- **Workers Static Assets** — *you build, then upload the artifact.* Deployable
  as desired state (the asset directory is a managed attribute of the Worker), so
  a new version is a normal apply. See `CloudflareWorker`.
- **Pages, git-connected** — *Cloudflare builds on git push.* Cloudflare is the
  CI; you manage the project (repo, branch, build settings, bindings, domains).
  That is what this component is for.

## The deployment model (and why this component manages only the project)

The Cloudflare Terraform/Pulumi provider exposes exactly two Pages resources —
`cloudflare_pages_project` and `cloudflare_pages_domain` — and **no deployment
resource**. Deployments are produced by the Cloudflare API/build pipeline, not by
IaC. So this component manages the project and its domains as desired state;
versions land out-of-band:

- **Git-connected** (`source` set): every push to the connected repo triggers a
  Cloudflare-side build and a new deployment.
- **Direct upload** (`source` omitted): `wrangler pages deploy ./dist
  --project-name=<name>` uploads a pre-built site.

Because no deployment exists at provision time, the stack outputs are limited to
project-level values (`project_name`, `subdomain`, `domains`, `created_on`).

## Git connection is a dashboard prerequisite

Connecting a repository requires authorizing Cloudflare with the git provider
(the GitHub App installation or GitLab OAuth). This is a one-time, browser-driven
step in the Cloudflare dashboard; the provider cannot bootstrap it. After it
exists, the `source` configuration (which repo, branch, build, preview policy) is
fully managed here.

## Per-environment configuration is paired

`deployment_configs` has a `preview` and a `production` block with an identical
shape. Cloudflare treats them as a **paired** configuration: it rejects a project
whose environments are configured inconsistently (for example, `fail_open` must
be equal across both). This component therefore mirrors a single provided
environment to both, so configuring just `production` "just works"; set both
explicitly when they must differ.

## Configuration surface

- **Build**: `build_config` (command, output dir, root dir, caching, web
  analytics).
- **Source**: git provider + repo/branch/preview/path settings.
- **Runtime** (`deployment_configs.{preview,production}`): compatibility date and
  flags, usage model, limits, placement, plain `vars`, secret `secrets`, and
  bindings to KV, D1, R2, Queues, Hyperdrive, Workers (services), Durable
  Objects, Analytics Engine, Vectorize, AI/Constellation, mTLS certs, and Browser
  Rendering. Cross-resource bindings are `valueFrom`-referenceable.
- **Domains**: custom hostnames (each in a zone on the account).

## Secrets

Secret env vars (`deployment_configs.*.secrets[].value`) and
`build_config.web_analytics_token` are secret-by-default: provide a managed-secret
reference resolved just-in-time at deploy. Plain configuration uses `vars`.
Modeling env vars as split `vars` (plain) + `secrets` (sensitive) — rather than
the provider's single typed `env_vars` map — keeps the secret annotation static
and the secret-coverage gate honest.

## Provider behaviors to know (discovered in practice)

- **Empty binding maps must be omitted, not sent as `{}`.** The provider
  normalizes an empty map to null and otherwise flags an inconsistent apply
  result; the module sends null for any empty binding group.
- **`fail_open` (and environment configs generally) must match across preview and
  production**, hence the mirroring above.
- A pure direct-upload project (no `source`) is created and destroyed cleanly;
  deployments are then pushed with wrangler.

## Composition

A Pages project is a leaf/mid-tier node: it consumes KV/D1/R2/Queue/Hyperdrive/
Worker resources via bindings and exposes `project_name`/`subdomain`/`domains`
for DNS and downstream wiring. Typical chart: backing data resources → Pages
project (binding them) → DNS records pointing at the custom domains.

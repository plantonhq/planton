# DigitalOcean App

## Overview

**DigitalOceanApp** deploys a full App Platform application: any mix of HTTP services, background workers, jobs, static sites, functions, and in-app databases, plus domains, ingress, alerts, and optional VPC placement.

App Platform runs the app, issues a default `ondigitalocean.app` hostname with HTTPS, and rolls out new deployments. You describe the app once; Terraform and Pulumi both create the same `digitalocean_app` resource.

The app name (`spec.appName`) is 2–32 characters, starts with a letter, and is unique across the account; component names follow the same rule. Component instance sizes are free-form slugs such as `basic-xxs` or `professional-s` — the provider does not publish a closed list, so new sizes work without a catalog change.

## What this kind is

This is the App Platform **application**, not a single service. An app can hold several components. A one-service app is the common starting point; a service-plus-worker app is the same kind.

Serverless functions that should stand alone use **DigitalOceanFunction** (a one-component App Platform app). Functions that should share an app with other components belong in `spec.functions` here.

## Sources

Each service, worker, and job sets exactly one source:

- **git** — public HTTPS clone URL. Use this when the DigitalOcean account has no linked GitHub/GitLab/Bitbucket connection.
- **github / gitlab / bitbucket** — `owner/repo` plus branch. `deployOnPush` needs the matching VCS connection in the control panel.
- **image** — container image. `registryType` is `docker_hub`, `docr`, or `ghcr`. For Docker Hub, `registry` is the namespace (`library` for official images). For DOCR, leave `registry` empty.

Static sites and in-app functions are git-only (no container image).

## Example

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanApp
metadata:
  name: demo-app
spec:
  appName: demo-app
  region: nyc
  services:
    - name: web
      image:
        registryType: docker_hub
        registry: library
        repository: nginx
        tag: latest
      instanceSizeSlug: basic-xxs
      instanceCount: 1
      httpPort: 80
```

## Stack outputs

| Output | Description |
|--------|-------------|
| `app_id` | App UUID. Used to import `digitalocean_app`. |
| `default_hostname` | Default `ondigitalocean.app` hostname. |
| `live_url` | Public URL including `https://`. |
| `live_domain` | Live hostname without scheme. |
| `active_deployment_id` | UUID of the currently live deployment. |

## Infrastructure as Code

- [Pulumi module](./iac/pulumi/README.md)
- [Terraform module](./iac/tf/README.md)

Operational judgment (VPC, `project.yml`, Pulumi SDK gaps, instance sizes) lives in [GUIDE.md](./GUIDE.md). Field-by-field schema is in [v1alpha1/reference.md](./v1alpha1/reference.md).

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

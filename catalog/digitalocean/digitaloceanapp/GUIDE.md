# DigitalOcean App -- Operational Guide

Judgment calls that matter when you run App Platform apps in production.

## The app name is short on purpose

`spec.appName` is 2–32 characters matching `^[a-z][a-z0-9-]{0,30}[a-z0-9]$` -- starts with a letter, ends with a letter or digit. That is the API's rule, not a Planton invention (it answers `name in body should match ...` otherwise), and every component name follows the same rule; validation enforces it on each name field so a bad name fails in seconds instead of minutes into a deploy. The app name must also be unique across every app in the account: the API's validate-only `POST /v2/apps/propose` reports a taken name as `app_name_available: false`, and the create fails.

Renaming an app is an in-place update, not a replacement -- but the default `<name>-<hash>.ondigitalocean.app` hostname carries the name, so the app's URL changes with it. Anything that calls the app by that hostname breaks until it is updated; put a custom domain in front when the URL must survive a rename.

## Regions are datacenter groups, not droplet slugs

`spec.region` takes an App Platform region group: `nyc`, `ams`, `fra`, `sfo`, `sgp`, `blr`, `tor`, `lon`, `syd`, `atl`, `ric`, `mkc` (the list is `GET /v2/apps/regions`). `nyc` covers the nyc1 and nyc3 datacenters; App Platform picks within the group. The API accepts a droplet slug such as `nyc3` but stores `nyc`, so a spec that said `nyc3` would re-plan on every apply -- which is why the field is typed with the group slugs and cannot say `nyc3`. When the app must sit beside a VPC or a database, match the group that contains their datacenter (a `nyc3` VPC goes with an `nyc` app).

## Instance sizes are free-form slugs

`instanceSizeSlug` is a string (`basic-xxs`, `professional-s`, …). App Platform does not publish a closed enum through Terraform, so new sizes work without a catalog change. Check current sizes in the DigitalOcean App Platform docs before you pick one. `basic-xxs` is the recommended default for getting started.

## Pick the source that matches the account

- **Public clone URL (`git`)** works with no VCS connection. This is the right default for accounts that have not linked GitHub.
- **`github` / `gitlab` / `bitbucket`** need the matching connection in the DigitalOcean control panel. `deployOnPush` is ignored without it, and a missing connection fails the deploy.
- **Container images** skip the build. For Docker Hub, `registry` is the namespace (`library` for official images), not the string `"docker-hub"`. For DigitalOcean Container Registry, set `registryType: docr` and leave `registry` empty.

## Autoscaling and instance_count do not combine

When `autoscaling` is set, leave `instanceCount` unset. App Platform ignores a fixed count while autoscaling is on, and the spec rejects the combination. Autoscaling is available on services and workers, not on jobs.

## Drain seconds are service-only

`termination.drainSeconds` is an HTTP connection drain. Workers and jobs reject it. They honor `gracePeriodSeconds` only.

## Terraform vs Pulumi

Both provisioners deploy the whole spec -- `vpc`, `maintenance`, ingress `authorityExact` matches, `ingress.secureHeader`, service and worker `livenessHealthCheck`, and alert destinations included. The one behavioral difference is the alert-destinations read-back below, which is a provider trait, not a wiring gap.

## Readiness versus liveness

A service's `healthCheck` is App Platform's readiness probe: while it fails, the component receives no traffic. `livenessHealthCheck` on a service or worker is the restart probe: when it fails `failureThreshold` times in a row, App Platform restarts the container. Point liveness at a cheap, dependency-free path -- a probe that touches the database restarts a healthy container every time the database hiccups. Both use the same field shape; `initialDelaySeconds` matters most on liveness, because a probe that starts before the process is listening restarts it in a loop.

## Alert destinations are write-only on the provider -- a perpetual diff on Terraform

Email and Slack destinations on app-level and component alerts are applied through a separate API call after the app spec is saved, and the provider never reads them back into state (a provider defect at v2.99.1, unchanged through v2.101.1, tracked as [digitalocean/terraform-provider-digitalocean#1606](https://github.com/digitalocean/terraform-provider-digitalocean/issues/1606); measured live: a refreshed plan proposed `+ destinations` on both alerts of a freshly applied app). Because any `spec` difference is an App update, **every Terraform apply with destinations set re-sends the spec and triggers a new deployment**. Pulumi shows the same diff only after a `pulumi refresh`; a plain `pulumi up` stays quiet. An import of an existing app never restores destinations.

Until the provider reads destinations back: set them on Pulumi stacks, or leave them unset on Terraform -- DigitalOcean then notifies the team's default address -- and manage recipients in the control panel. Whatever you set, emails must belong to verified team members; the API rejects the call otherwise.

## Secret environment variables re-plan on Terraform

`envs[].secret` values come back from the API encrypted. Terraform stores the encrypted form and compares it with your plaintext on the next plan, so every apply shows the secret as changed and redeploys the app (upstream digitalocean/terraform-provider-digitalocean#869). This is the provider's documented behaviour, not a Planton bug. Pulumi is affected only after a refresh. If you run Terraform with secret envs, expect the diff, or move the secret into a bind or a managed secret store the app reads at runtime.

## Deprecated App Spec surfaces are not modeled

Per-component `routes` / `cors` and the old top-level `domains` list are schema-deprecated on the provider. Ingress and `spec.domains` (the current domain list) replace them. Do not expect those old blocks on this kind.

## In-app databases vs DigitalOceanDatabaseCluster

`spec.databases[].clusterName` references an existing DigitalOcean database cluster (or a DigitalOceanDatabaseCluster resource). It does not create the cluster. An in-app database with no `clusterName` is App Platform's managed dev database, which is not a production data store.

## Custom domains need a zone you control

`spec.domains[].zone` can reference a DigitalOceanDnsZone. The zone must already exist; App Platform will not create DNS for you. Omit domains to use the default `ondigitalocean.app` hostname.

## `project_id` is a literal UUID, and it is create-only

Pass the project UUID as a string (a typed reference to DigitalOceanProject is recorded backlog). The provider marks it ForceNew: moving an app to another project destroys and recreates it -- new UUID, new default hostname. Leave it unset to land in the account's default project, or set it once at creation.

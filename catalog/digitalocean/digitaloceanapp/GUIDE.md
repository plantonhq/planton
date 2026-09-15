# DigitalOcean App -- Operational Guide

Judgment calls that matter when you run App Platform apps in production.

## The app name is short on purpose

`spec.appName` is 2–32 characters. That is the provider's limit, not a Planton invention. DNS-friendly names only (`a-z`, `0-9`, hyphens). Changing the name replaces the app.

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

## Terraform vs Pulumi at the current Pulumi SDK (v4.49.0)

These fields are real on the spec and Terraform wires them. Pulumi fails the apply with a loud `PARITY-EXCEPTION` if they are set, until the SDK grows the matching args:

- `spec.vpc`
- `spec.maintenance`
- service/worker `livenessHealthCheck`
- `spec.ingress.secureHeader`
- `spec.ingress.rule.match.authority` (`authorityExact`)
- alert destinations (emails / Slack) — Terraform's alert block carries them; Pulumi's alert args do not

Use Terraform for those arms, or omit them on Pulumi stacks.

## Deprecated App Spec surfaces are not modeled

Per-component `routes` / `cors` and the old top-level `domains` list are schema-deprecated on the provider. Ingress and `spec.domains` (the current domain list) replace them. Do not expect those old blocks on this kind.

## In-app databases vs DigitalOceanDatabaseCluster

`spec.databases[].clusterName` references an existing DigitalOcean database cluster (or a DigitalOceanDatabaseCluster resource). It does not create the cluster. An in-app database with no `clusterName` is App Platform's managed dev database, which is not a production data store.

## Custom domains need a zone you control

`spec.domains[].zone` can reference a DigitalOceanDnsZone. The zone must already exist; App Platform will not create DNS for you. Omit domains to use the default `ondigitalocean.app` hostname.

## `projectId` places the app in a project

`spec.projectId` can reference a DigitalOceanProject (its `project_id` output) or carry a literal project UUID. Leave it unset and the app lands in the account's default project.

## Pulling from DigitalOcean Container Registry

An image with `registryType: docr` is pulled from the account's own registry; App Platform resolves it itself, so `registry` stays empty and there is no reference to wire. To record that the app depends on a `DigitalOceanContainerRegistry` Planton manages -- so the registry deploys first and the dependency shows on the diagram -- declare it under `metadata.relationships` with type `uses`.

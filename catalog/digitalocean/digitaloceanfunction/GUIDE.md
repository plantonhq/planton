# DigitalOcean Function -- Operational Guide

Judgment calls that matter when you run Functions on App Platform.

## There is no Functions Terraform resource

Both engines create `digitalocean_app` with one functions component. `function_id` is that app's UUID. Import uses the same `digitalocean_app` id format as DigitalOceanApp (`{app_id}`), derived from the `function_id` stack output.

## Runtime does not belong on the spec

`project.yml` in the repository (at the root, or inside `sourceDirectory` when set) is the source of truth for:

- runtime (`nodejs:18`, `python:3.11`, …)
- memory and timeout
- web vs. non-web (`web: true/false`)
- cron schedules
- per-function entrypoints

App Platform reads that file at deploy time. Terraform and Pulumi cannot set those knobs on `digitalocean_app`. A spec field for them would look configurable and do nothing — so they are omitted on purpose.

To change runtime or add a schedule, edit `project.yml` and redeploy.

## `sourceDirectory` is load-bearing

It must point at the directory that contains `project.yml`. Leave it unset when `project.yml` is at the repository root -- DigitalOcean's `sample-functions-nodejs-helloworld` is laid out that way (`project.yml` at the root, functions under `packages/`), and DigitalOcean's own deploy template for it uses the root. Set it only when `project.yml` lives in a subdirectory, for example `functions/api`. Pointing it at the `packages/` tree itself is the classic mistake: the build fails minutes into the deploy with no `project.yml` found, and nothing at validation time can catch it.

## Regions are datacenter groups, not droplet slugs

`spec.region` takes an App Platform region group (`nyc`, `ams`, `fra`, `sfo`, `sgp`, `blr`, `tor`, `lon`, `syd`, `atl`, `ric`, `mkc` -- `GET /v2/apps/regions`), never a droplet slug such as `nyc3`. The API accepts `nyc3` but stores `nyc`, and a spec that said `nyc3` would re-plan on every apply; the field's type makes that impossible.

## Pick `git` unless GitHub is actually connected

`github.deployOnPush` requires the DigitalOcean account to have GitHub connected in the control panel. Without that connection the deploy fails. A public clone URL on `git` works with no extra setup and is the right default for new accounts and for E2E.

## Function name vs app name

`spec.appName` names the App Platform **app**; `spec.functionName` names the functions **component** inside it. They can differ, and both follow the API's rule `^[a-z][a-z0-9-]{0,30}[a-z0-9]$` -- 2-32 characters, starting with a letter. The app name must also be unique across every app in the DigitalOcean account; the API answers a duplicate with `app_name_available: false` on its validate-only `POST /v2/apps/propose` call and rejects the create. The app name is a spec field, never derived from `metadata.name`, because Planton names are routinely longer than 32 characters and unique only within an org and environment.

Renaming the app is an in-place update, not a replacement. The default `<name>-<hash>.ondigitalocean.app` hostname carries the name, so the function's URL changes with it -- anything that calls the endpoint by that hostname breaks until it is updated. Import and the verifier key off the app UUID (`function_id`), not either name.

## Moving the app between projects replaces it

`projectId` is create-only on the provider: changing it destroys the app and creates a new one (new UUID, new default hostname). Leave it unset to land in the account's default project, or set it once at creation.

## Environment variables

Use `envs[].plaintext` for ordinary values and `envs[].secret` for credentials. Secrets are stored in App Platform's secret store (`type = SECRET` on the provider). There is no separate `secretEnvironmentVariables` map. Secret values come back from the API encrypted, so on Terraform every plan shows them as changed and redeploys the app (upstream digitalocean/terraform-provider-digitalocean#869); Pulumi is affected only after a refresh.

## Alert destinations are write-only on the provider

Both engines wire `alerts[].destinations`, but the provider never reads destinations back into state (a provider defect at v2.99.1, measured live on the App kind; tracked as [digitalocean/terraform-provider-digitalocean#1606](https://github.com/digitalocean/terraform-provider-digitalocean/issues/1606)). On Terraform a refreshed plan proposes them again forever and every apply redeploys the app; Pulumi shows the diff only after a refresh. Set destinations on Pulumi stacks, or leave them unset and manage recipients in the control panel. Recipients must be verified team members.

## When this kind is the wrong shape

If the functions component should sit next to an HTTP service, a worker, or a static site, use DigitalOceanApp and put the functions component in `spec.functions`. DigitalOceanFunction is the one-component app.

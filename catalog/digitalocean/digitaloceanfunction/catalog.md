# DigitalOcean Function

Deploys serverless functions as an App Platform app with a single functions component. There is no standalone Functions resource; both engines create `digitalocean_app`. Runtime, memory, timeout, and schedules live in the repo's `project.yml`, not on this spec.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **App Platform Application** -- a `digitalocean_app` whose name is the Cloud Resource metadata name
- **Functions component** -- one functions component named `spec.functionName`, sourced from the Git remote you declared
- **Environment variables** -- from `spec.envs`; `secret` values are stored in App Platform's secret store

App Platform reads `project.yml` from `sourceDirectory` to set runtime, memory, timeout, entrypoint, and any cron triggers.

## Before You Deploy

### Planton Setup

- **DigitalOcean Provider Connection** -- an active connection in the Connect module with a DigitalOcean API token. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline API token authentication.

### DigitalOcean Account

- **A Git repository** in DigitalOcean Functions layout: a `project.yml` and a packages tree. Provide a public clone URL (`git`), or a linked GitHub/GitLab/Bitbucket repo.
- **A source directory** that contains `project.yml` (for DigitalOcean's hello-world sample, `packages`).
- **A supported App Platform region** (for example `nyc3`).

## Deploy

### Console

Open the deployment store, find **DigitalOcean Function**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Hello-World Function** preset in the [Presets](#presets) tab to deploy DigitalOcean's public Node.js sample.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanFunction
metadata:
  name: hello
  org: acme-corp
  env: prod
spec:
  functionName: hello
  region: nyc3
  git:
    repoCloneUrl: https://github.com/digitalocean/sample-functions-nodejs-helloworld.git
    branch: master
  sourceDirectory: packages
```

```shell
planton apply -f do-function.yaml
```

This clones the public hello-world sample and deploys it as an HTTP function; no GitHub connection is required. A Stack Job tracks the provisioning in real time.

### InfraChart

When the functions app belongs in a project Planton manages, use ValueFromRef to reference the project deployed in the same InfraPipeline:

```yaml
spec:
  projectId:
    valueFrom:
      kind: DigitalOceanProject
      name: platform
      fieldPath: status.outputs.project_id
```

## Key Configuration

These are the most important decisions when configuring a functions app. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Function name vs app name** -- `functionName` is the component name inside the app (max 32 characters); the App Platform app name is the Cloud Resource `metadata.name`. They can differ, and import and verification key off the app UUID (`function_id`), not the component name.

**Source** -- exactly one of `git`, `github`, `gitlab`, or `bitbucket`; there is no container-image source for functions. The VCS-linked sources need the matching connection in the DigitalOcean control panel -- `deployOnPush` fails without it -- so `git` with a public clone URL is the right default for new accounts.

**Source directory** -- `sourceDirectory` must point at the directory containing `project.yml` and the packages tree (for DigitalOcean's hello-world sample that is `packages`, not the repo root). A wrong directory produces a failed App Platform build, not a spec-validation error, so it surfaces minutes into the deploy rather than at apply time.

**Runtime, memory, timeout, schedules** -- edit `project.yml` in the repo. They are deliberately not fields on this spec: Terraform and Pulumi cannot set those knobs on `digitalocean_app`, so a spec field for them would look configurable and do nothing. To change runtime or add a cron schedule, edit `project.yml` and redeploy.

**Environment variables** -- use `envs[].plaintext` for ordinary values and `envs[].secret` for credentials; secrets are stored in App Platform's secret store.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **DigitalOceanProject** (optional) | `projectId` | `status.outputs.project_id` |

Sources are Git coordinates, not references; the project is the one Planton-managed resource a functions app names.

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `function_id` | App Platform app UUID hosting the functions component | Import, API operations |
| `https_endpoint` | Public HTTPS URL for invoking the function | Webhooks, API clients |
| `default_hostname` | Default `ondigitalocean.app` hostname | DNS, health checks |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Hello-world from public git** -- DigitalOcean's Node.js sample deployed from its public clone URL, no GitHub connection. The shape for proving the pipeline before pointing at your own repo. Start from the **Hello-World Function** preset.

**Linked GitHub with deploy-on-push** -- switch the source to `github` with `repo: owner/repo` and `deployOnPush: true` after connecting GitHub in the DigitalOcean control panel; every push to the branch redeploys the functions app.

## Works With

Functions that should share an app with HTTP services or workers belong on [**DigitalOcean App Platform App**](/cloud-catalog/digital-ocean-app) as `spec.functions`, not on this kind.

# DigitalOcean Function

## Overview

**DigitalOceanFunction** deploys serverless functions as an App Platform application with a single functions component. DigitalOcean's Terraform provider has no standalone Functions resource; both Pulumi and Terraform create `digitalocean_app`.

Runtime, memory, timeout, entrypoint, and cron schedules are **not** on this spec. They live in the repo's `project.yml`, which App Platform reads at deploy time from the repository root or from `sourceDirectory` when set. Putting those knobs on the spec would silently do nothing.

## Two names

`appName` names the App Platform **app** and `functionName` names the functions **component** inside it. Both follow the API's rule `^[a-z][a-z0-9-]{0,30}[a-z0-9]$` (2-32 characters, starts with a letter), and `appName` must be unique across every app in the DigitalOcean account. Renaming the app updates it in place, but its default `<name>-<hash>.ondigitalocean.app` URL changes with the name.

## When to use this kind vs DigitalOceanApp

Use **DigitalOceanFunction** when the functions component is the whole app. Use **DigitalOceanApp** `spec.functions` when functions should share an app with services, workers, or static sites.

## Sources

Set exactly one:

- **git** — public HTTPS clone URL. Use this when the DigitalOcean account has no linked GitHub/GitLab/Bitbucket connection.
- **github / gitlab / bitbucket** — `owner/repo` plus branch. `deployOnPush` needs the matching VCS connection.

Functions are git-only. There is no container-image source.

`sourceDirectory` is the directory that contains `project.yml`. Leave it unset when `project.yml` is at the repository root -- DigitalOcean's official hello-world sample is laid out that way. Set it (for example `functions/api`) only when `project.yml` lives in a subdirectory. A wrong directory fails the App Platform build minutes into the deploy, never at validation.

## Example

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanFunction
metadata:
  name: hello
spec:
  appName: hello-fn
  functionName: hello
  region: nyc
  git:
    repoCloneUrl: https://github.com/digitalocean/sample-functions-nodejs-helloworld.git
    branch: master
```

## Stack outputs

| Output | Description |
|--------|-------------|
| `function_id` | App Platform app UUID that hosts the functions component. Used to import `digitalocean_app`. |
| `https_endpoint` | Public HTTPS URL of the functions HTTP endpoint. |
| `default_hostname` | Default `ondigitalocean.app` hostname. |

## Infrastructure as Code

- [Pulumi module](./iac/pulumi/README.md)
- [Terraform module](./iac/tf/README.md)

How `project.yml` works, and why runtime is not on the spec, is in [GUIDE.md](./GUIDE.md). Field-by-field schema is in [v1alpha1/reference.md](./v1alpha1/reference.md).

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

# DigitalOceanFunction

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `digital-ocean.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

DigitalOceanFunctionSpec deploys serverless functions as an App Platform
app with a single functions component. The provider has no standalone
Functions resource; both engines create digitalocean_app.

Runtime, memory, timeout, entrypoint, and schedules are NOT on this spec.
They live in the repo's project.yml (inside source_directory), which App
Platform reads at deploy time. Putting those knobs on the spec would
silently do nothing.

## Example

```yaml
# DigitalOcean Function -- examples
#
# DigitalOceanFunction deploys serverless functions as an App Platform
# app with a single functions component. There is no standalone
# Functions resource; both engines create digitalocean_app.
#
# Runtime, memory, timeout, entrypoint, and schedules live in the
# repo's project.yml. App Platform reads it from sourceDirectory, or
# from the repository root when sourceDirectory is unset (this sample).
# Putting those knobs on this spec would silently do nothing.
#
# appName is the App Platform app's own name: 2-32 characters, starts
# with a letter, unique across every app in the DigitalOcean account.
#
# Usage:
#   planton apply -f manifest.yaml

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

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.functionName` | `string` | yes |  |  |
| `spec.region` | `enum` | yes |  |  |
| `spec.git` | `DigitalOceanAppGitSource` |  |  |  |
| `spec.git.repoCloneUrl` | `string` | yes |  |  |
| `spec.git.branch` | `string` | yes |  |  |
| `spec.github` | `DigitalOceanAppGithubSource` |  |  |  |
| `spec.github.repo` | `string` | yes |  |  |
| `spec.github.branch` | `string` | yes |  |  |
| `spec.github.deployOnPush` | `bool` |  |  |  |
| `spec.gitlab` | `DigitalOceanAppGitlabSource` |  |  |  |
| `spec.gitlab.repo` | `string` | yes |  |  |
| `spec.gitlab.branch` | `string` | yes |  |  |
| `spec.gitlab.deployOnPush` | `bool` |  |  |  |
| `spec.bitbucket` | `DigitalOceanAppBitbucketSource` |  |  |  |
| `spec.bitbucket.repo` | `string` | yes |  |  |
| `spec.bitbucket.branch` | `string` | yes |  |  |
| `spec.bitbucket.deployOnPush` | `bool` |  |  |  |
| `spec.sourceDirectory` | `string` |  |  |  |
| `spec.envs` | `[]DigitalOceanAppEnvVar` |  |  |  |
| `spec.envs[].key` | `string` | yes |  |  |
| `spec.envs[].plaintext` | `string` |  |  |  |
| `spec.envs[].secret` | `string` (sensitive) |  |  |  |
| `spec.envs[].scope` | `enum` |  |  |  |
| `spec.alerts` | `[]DigitalOceanAppComponentAlert` |  |  |  |
| `spec.alerts[].rule` | `enum` | yes |  |  |
| `spec.alerts[].operator` | `enum` | yes |  |  |
| `spec.alerts[].window` | `enum` | yes |  |  |
| `spec.alerts[].value` | `double` |  |  |  |
| `spec.alerts[].disabled` | `bool` |  |  |  |
| `spec.alerts[].destinations` | `DigitalOceanAppAlertDestinations` |  |  |  |
| `spec.alerts[].destinations.emails` | `[]string` |  |  |  |
| `spec.alerts[].destinations.slackWebhooks` | `[]DigitalOceanAppSlackWebhook` |  |  |  |
| `spec.alerts[].destinations.slackWebhooks[].channel` | `string` | yes |  |  |
| `spec.alerts[].destinations.slackWebhooks[].url` | `string` (sensitive) | yes |  |  |
| `spec.logDestinations` | `[]DigitalOceanAppLogDestination` |  |  |  |
| `spec.logDestinations[].name` | `string` | yes |  |  |
| `spec.logDestinations[].papertrail` | `DigitalOceanAppPapertrailLog` |  |  |  |
| `spec.logDestinations[].papertrail.endpoint` | `string` | yes |  |  |
| `spec.logDestinations[].datadog` | `DigitalOceanAppDatadogLog` |  |  |  |
| `spec.logDestinations[].datadog.apiKey` | `string` (sensitive) | yes |  |  |
| `spec.logDestinations[].datadog.endpoint` | `string` |  |  |  |
| `spec.logDestinations[].logtail` | `DigitalOceanAppLogtailLog` |  |  |  |
| `spec.logDestinations[].logtail.token` | `string` (sensitive) | yes |  |  |
| `spec.logDestinations[].openSearch` | `DigitalOceanAppOpenSearchLog` |  |  |  |
| `spec.logDestinations[].openSearch.endpoint` | `string` |  |  |  |
| `spec.logDestinations[].openSearch.indexName` | `string` |  |  |  |
| `spec.logDestinations[].openSearch.clusterName` | `string` |  |  |  |
| `spec.logDestinations[].openSearch.basicAuth` | `DigitalOceanAppOpenSearchBasicAuth` |  |  |  |
| `spec.logDestinations[].openSearch.basicAuth.user` | `string` |  |  |  |
| `spec.logDestinations[].openSearch.basicAuth.password` | `string` (sensitive) |  |  |  |
| `spec.projectId` | `string \| valueFrom` |  |  | DigitalOceanProject (`status.outputs.project_id`) |
| `spec.appName` | `string` | yes |  |  |

## Field Details

### spec.functionName

`string` · required

Functions component name inside the app. The App Platform API enforces
^[a-z][a-z0-9-]{0,30}[a-z0-9]$ on every component name (2-32 chars,
starts with a letter, ends with a letter or digit) and rejects the
whole app spec otherwise, so the same rule is enforced here.

- rule: {"required":true,"string":{"minLen":"2","maxLen":"32","pattern":"^[a-z][a-z0-9-]{0,30}[a-z0-9]$"}}

### spec.region

`enum` · required

App Platform region group, for example nyc (never a droplet slug such
as nyc3 -- the API would store nyc and the plan would never settle).

- rule: {"required":true}

Allowed values (use exactly as shown):

- `digital_ocean_app_region_unspecified`
- `ams` -- Amsterdam (ams3)
- `nyc` -- New York (nyc1, nyc3)
- `fra` -- Frankfurt (fra1)
- `sfo` -- San Francisco (sfo3)
- `sgp` -- Singapore (sgp1)
- `blr` -- Bangalore (blr1)
- `tor` -- Toronto (tor1)
- `lon` -- London (lon1)
- `syd` -- Sydney (syd1)
- `atl` -- Atlanta (atl1)
- `ric` -- Richmond (ric1)
- `mkc` -- Kansas City (mkc1)

### spec.git

`DigitalOceanAppGitSource`

### spec.git.repoCloneUrl

`string` · required

HTTPS or git clone URL, for example https://github.com/example/app.git

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.git.branch

`string` · required

Branch to deploy. Example: main

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.github

`DigitalOceanAppGithubSource`

### spec.github.repo

`string` · required

Repository in owner/repo form, for example plantonhq/demo

- rule: {"required":true,"string":{"pattern":"^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$"}}

### spec.github.branch

`string` · required

Branch to deploy. Example: main

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.github.deployOnPush

`bool`

Redeploy automatically when this branch is pushed.

### spec.gitlab

`DigitalOceanAppGitlabSource`

### spec.gitlab.repo

`string` · required

- rule: {"required":true,"string":{"pattern":"^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$"}}

### spec.gitlab.branch

`string` · required

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.gitlab.deployOnPush

`bool`

### spec.bitbucket

`DigitalOceanAppBitbucketSource`

### spec.bitbucket.repo

`string` · required

- rule: {"required":true,"string":{"pattern":"^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$"}}

### spec.bitbucket.branch

`string` · required

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.bitbucket.deployOnPush

`bool`

### spec.sourceDirectory

`string`

Directory inside the repo that contains project.yml (App Platform reads
runtime, memory, timeout, and schedules from it). Leave unset when
project.yml is at the repository root -- DigitalOcean's own hello-world
sample is laid out that way. Set it only when project.yml lives in a
subdirectory, for example functions/api. A wrong directory fails the
App Platform build minutes into the deploy, never at validation.

### spec.envs

`[]DigitalOceanAppEnvVar`

- rule: set either plaintext or secret for this environment variable - App Platform needs a value

### spec.envs[].key

`string` · required

Variable name, for example DATABASE_URL or NODE_ENV.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.envs[].plaintext

`string`

Non-secret value. Visible in the App Platform UI and in build logs.

### spec.envs[].secret

`string` · sensitive

Secret value (API keys, database URLs, tokens). Stored in App Platform's
secret store; the IaC modules send type=SECRET.

### spec.envs[].scope

`enum`

When the variable is injected. Omit to use run_and_build_time.

Allowed values (use exactly as shown):

- `digital_ocean_app_env_scope_unspecified`
- `run_and_build_time` -- Injected during the build and at runtime (provider default).
- `run_time` -- Injected only at runtime.
- `build_time` -- Injected only during the build.
- `unset` -- Provider UNSET - treated as run_and_build_time by the API.

### spec.alerts

`[]DigitalOceanAppComponentAlert`

### spec.alerts[].rule

`enum` · required

- rule: {"required":true}

Allowed values (use exactly as shown):

- `digital_ocean_app_component_alert_rule_unspecified`
- `cpu_utilization`
- `mem_utilization`
- `restart_count`

### spec.alerts[].operator

`enum` · required

- rule: {"required":true}

Allowed values (use exactly as shown):

- `digital_ocean_app_alert_operator_unspecified`
- `greater_than`
- `less_than`

### spec.alerts[].window

`enum` · required

- rule: {"required":true}

Allowed values (use exactly as shown):

- `digital_ocean_app_alert_window_unspecified`
- `five_minutes`
- `ten_minutes`
- `thirty_minutes`
- `one_hour`

### spec.alerts[].value

`double`

Threshold. For cpu_utilization / mem_utilization this is a percent; for
restart_count it is a count.

- rule: {"double":{"gte":0}}

### spec.alerts[].disabled

`bool`

### spec.alerts[].destinations

`DigitalOceanAppAlertDestinations`

### spec.alerts[].destinations.emails

`[]string`

### spec.alerts[].destinations.slackWebhooks

`[]DigitalOceanAppSlackWebhook`

### spec.alerts[].destinations.slackWebhooks[].channel

`string` · required

- rule: {"required":true}

### spec.alerts[].destinations.slackWebhooks[].url

`string` · required · sensitive

- rule: {"required":true}

### spec.logDestinations

`[]DigitalOceanAppLogDestination`

- rule: set exactly one sink: papertrail, datadog, logtail, or open_search

### spec.logDestinations[].name

`string` · required

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.logDestinations[].papertrail

`DigitalOceanAppPapertrailLog`

### spec.logDestinations[].papertrail.endpoint

`string` · required

Syslog endpoint, for example logs.papertrailapp.com:12345

- rule: {"required":true}

### spec.logDestinations[].datadog

`DigitalOceanAppDatadogLog`

### spec.logDestinations[].datadog.apiKey

`string` · required · sensitive

- rule: {"required":true}

### spec.logDestinations[].datadog.endpoint

`string`

Defaults to https://http-intake.logs.datadoghq.com when omitted.

### spec.logDestinations[].logtail

`DigitalOceanAppLogtailLog`

### spec.logDestinations[].logtail.token

`string` · required · sensitive

- rule: {"required":true}

### spec.logDestinations[].openSearch

`DigitalOceanAppOpenSearchLog`

### spec.logDestinations[].openSearch.endpoint

`string`

### spec.logDestinations[].openSearch.indexName

`string`

### spec.logDestinations[].openSearch.clusterName

`string`

### spec.logDestinations[].openSearch.basicAuth

`DigitalOceanAppOpenSearchBasicAuth`

The provider requires this block even when user and password are empty
(App Platform's OpenSearch integration uses it as a placeholder).

### spec.logDestinations[].openSearch.basicAuth.user

`string`

### spec.logDestinations[].openSearch.basicAuth.password

`string` · sensitive

### spec.projectId

`string | valueFrom`

(Optional) The project the functions app is created in. Reference a
DigitalOceanProject resource (the default wiring resolves its
project_id output) or pass a literal project UUID. When unset, the app
lands in the account's default project. Create-only: the provider marks
project_id ForceNew, so changing it destroys and recreates the app.

- references: DigitalOceanProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: DigitalOceanProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.appName

`string` · required

Name of the App Platform app that hosts the functions component. This is
spec.name on digitalocean_app: 2-32 chars, ^[a-z][a-z0-9-]{0,30}[a-z0-9]$,
and unique across every app in the DigitalOcean account (the API answers
"name in body should be at most 32 chars long" / "app_name_available:
false" otherwise). It is a spec field, never derived from metadata.name,
because Planton names are longer than 32 chars and unique only within an
org/env. Renaming updates the app in place; the default
<name>-<hash>.ondigitalocean.app hostname carries the name, so the URL
changes with it.

- rule: {"required":true,"string":{"minLen":"2","maxLen":"32","pattern":"^[a-z][a-z0-9-]{0,30}[a-z0-9]$"}}

## Validation Rules

- `function_one_source`: set exactly one source: git, github, gitlab, or bitbucket. Use git with a public clone URL when the DigitalOcean account has no linked GitHub/GitLab/Bitbucket connection

## Outputs

Reference an output from another manifest as `valueFrom: {kind: DigitalOceanFunction, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.function_id` | `string` | App Platform app UUID that hosts the functions component. Used to import the digitalocean_app resource. |
| `status.outputs.https_endpoint` | `string` | Public HTTPS URL of the app (the functions HTTP endpoint). |
| `status.outputs.default_hostname` | `string` | Default ondigitalocean.app hostname assigned by the platform. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | DigitalOceanProject | `status.outputs.project_id` |

## See Also

- [Overview](../README.md)

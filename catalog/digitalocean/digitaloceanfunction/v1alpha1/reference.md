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
# repo's project.yml inside sourceDirectory. Putting those knobs on
# this spec would silently do nothing.
#
# Usage:
#   planton apply -f manifest.yaml

apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanFunction
metadata:
  name: hello
spec:
  functionName: hello
  region: nyc3
  git:
    repoCloneUrl: https://github.com/digitalocean/sample-functions-nodejs-helloworld.git
    branch: master
  sourceDirectory: packages
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
| `spec.sourceDirectory` | `string` | yes |  |  |
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

## Field Details

### spec.functionName

`string` · required

Functions component name inside the app.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"32"}}

### spec.region

`enum` · required

- rule: {"required":true}

Allowed values (use exactly as shown):

- `digital_ocean_region_unspecified` -- 0: default / unspecified region
- `nyc3` -- new york 3
- `sfo3` -- san francisco 3
- `fra1` -- frankfurt 1
- `sgp1` -- singapore 1
- `lon1` -- london 1
- `tor1` -- toronto 1
- `blr1` -- bangalore 1
- `ams3` -- amsterdam 3
- `nyc1` -- new york 1
- `nyc2` -- new york 2
- `sfo2` -- san francisco 2
- `syd1` -- sydney 1
- `atl1` -- atlanta 1

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

`string` · required

Directory inside the repo that contains project.yml and the packages
tree, for example packages/api.

- rule: {"required":true,"string":{"minLen":"1"}}

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
lands in the account's default project.

- references: DigitalOceanProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: DigitalOceanProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

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

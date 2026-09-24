# DigitalOceanApp

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `digital-ocean.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

DigitalOceanAppSpec is the full App Platform application: any mix of
services, workers, jobs, static sites, functions, and in-app databases,
plus domains, ingress, alerts, and VPC placement.

The app name is spec.app_name (2-32 characters, the API's limit). Every
component name shares the API's naming rule, ^[a-z][a-z0-9-]{0,30}[a-z0-9]$,
and the API rejects the whole app spec when any one of them breaks it, so
the rule is enforced at validation on each name field below. Component
instance sizes are free-form slugs such as basic-xxs or professional-s -
the provider does not publish a closed list, so new sizes work without a
catalog change.

## Example

```yaml
# DigitalOcean App -- examples
#
# DigitalOceanApp is a full App Platform application: any mix of
# services, workers, jobs, static sites, functions, and in-app
# databases. The app name is 2-32 characters. Component instance
# sizes are free-form slugs such as basic-xxs.
#
# Usage:
#   planton apply -f manifest.yaml

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
      envs:
        - key: NODE_ENV
          plaintext: production

---
# Git-source web service. Use the generic git clone URL when the
# DigitalOcean account has no linked GitHub connection. Runtime and
# build are auto-detected from the repo.

apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanApp
metadata:
  name: sample-nodejs
spec:
  appName: sample-nodejs
  region: nyc
  services:
    - name: web
      git:
        repoCloneUrl: https://github.com/digitalocean/sample-nodejs.git
        branch: main
      instanceSizeSlug: basic-xxs
      instanceCount: 1
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.appName` | `string` | yes |  |  |
| `spec.region` | `enum` | yes |  |  |
| `spec.services` | `[]DigitalOceanAppService` |  |  |  |
| `spec.services[].name` | `string` | yes |  |  |
| `spec.services[].sourceDir` | `string` |  |  |  |
| `spec.services[].git` | `DigitalOceanAppGitSource` |  |  |  |
| `spec.services[].git.repoCloneUrl` | `string` | yes |  |  |
| `spec.services[].git.branch` | `string` | yes |  |  |
| `spec.services[].github` | `DigitalOceanAppGithubSource` |  |  |  |
| `spec.services[].github.repo` | `string` | yes |  |  |
| `spec.services[].github.branch` | `string` | yes |  |  |
| `spec.services[].github.deployOnPush` | `bool` |  |  |  |
| `spec.services[].gitlab` | `DigitalOceanAppGitlabSource` |  |  |  |
| `spec.services[].gitlab.repo` | `string` | yes |  |  |
| `spec.services[].gitlab.branch` | `string` | yes |  |  |
| `spec.services[].gitlab.deployOnPush` | `bool` |  |  |  |
| `spec.services[].bitbucket` | `DigitalOceanAppBitbucketSource` |  |  |  |
| `spec.services[].bitbucket.repo` | `string` | yes |  |  |
| `spec.services[].bitbucket.branch` | `string` | yes |  |  |
| `spec.services[].bitbucket.deployOnPush` | `bool` |  |  |  |
| `spec.services[].image` | `DigitalOceanAppImageSource` |  |  |  |
| `spec.services[].image.registryType` | `enum` | yes |  |  |
| `spec.services[].image.registry` | `string` |  |  |  |
| `spec.services[].image.repository` | `string` | yes |  |  |
| `spec.services[].image.tag` | `string` |  |  |  |
| `spec.services[].image.digest` | `string` |  |  |  |
| `spec.services[].image.registryCredentials` | `string` (sensitive) |  |  |  |
| `spec.services[].image.deployOnPush` | `bool` |  |  |  |
| `spec.services[].environmentSlug` | `string` |  |  |  |
| `spec.services[].dockerfilePath` | `string` |  |  |  |
| `spec.services[].buildCommand` | `string` |  |  |  |
| `spec.services[].runCommand` | `string` |  |  |  |
| `spec.services[].instanceSizeSlug` | `string` |  | `basic-xxs` |  |
| `spec.services[].instanceCount` | `uint32` |  | `1` |  |
| `spec.services[].httpPort` | `uint32` |  |  |  |
| `spec.services[].internalPorts` | `[]uint32` |  |  |  |
| `spec.services[].healthCheck` | `DigitalOceanAppHealthCheck` |  |  |  |
| `spec.services[].healthCheck.port` | `uint32` |  |  |  |
| `spec.services[].healthCheck.httpPath` | `string` |  |  |  |
| `spec.services[].healthCheck.initialDelaySeconds` | `uint32` |  |  |  |
| `spec.services[].healthCheck.periodSeconds` | `uint32` |  |  |  |
| `spec.services[].healthCheck.timeoutSeconds` | `uint32` |  |  |  |
| `spec.services[].healthCheck.successThreshold` | `uint32` |  |  |  |
| `spec.services[].healthCheck.failureThreshold` | `uint32` |  |  |  |
| `spec.services[].livenessHealthCheck` | `DigitalOceanAppHealthCheck` |  |  |  |
| `spec.services[].livenessHealthCheck.port` | `uint32` |  |  |  |
| `spec.services[].livenessHealthCheck.httpPath` | `string` |  |  |  |
| `spec.services[].livenessHealthCheck.initialDelaySeconds` | `uint32` |  |  |  |
| `spec.services[].livenessHealthCheck.periodSeconds` | `uint32` |  |  |  |
| `spec.services[].livenessHealthCheck.timeoutSeconds` | `uint32` |  |  |  |
| `spec.services[].livenessHealthCheck.successThreshold` | `uint32` |  |  |  |
| `spec.services[].livenessHealthCheck.failureThreshold` | `uint32` |  |  |  |
| `spec.services[].autoscaling` | `DigitalOceanAppAutoscaling` |  |  |  |
| `spec.services[].autoscaling.minInstanceCount` | `uint32` | yes |  |  |
| `spec.services[].autoscaling.maxInstanceCount` | `uint32` | yes |  |  |
| `spec.services[].autoscaling.cpuPercent` | `uint32` | yes | `80` |  |
| `spec.services[].termination` | `DigitalOceanAppTermination` |  |  |  |
| `spec.services[].termination.gracePeriodSeconds` | `uint32` |  |  |  |
| `spec.services[].termination.drainSeconds` | `uint32` |  |  |  |
| `spec.services[].envs` | `[]DigitalOceanAppEnvVar` |  |  |  |
| `spec.services[].envs[].key` | `string` | yes |  |  |
| `spec.services[].envs[].plaintext` | `string` |  |  |  |
| `spec.services[].envs[].secret` | `string` (sensitive) |  |  |  |
| `spec.services[].envs[].scope` | `enum` |  |  |  |
| `spec.services[].alerts` | `[]DigitalOceanAppComponentAlert` |  |  |  |
| `spec.services[].alerts[].rule` | `enum` | yes |  |  |
| `spec.services[].alerts[].operator` | `enum` | yes |  |  |
| `spec.services[].alerts[].window` | `enum` | yes |  |  |
| `spec.services[].alerts[].value` | `double` |  |  |  |
| `spec.services[].alerts[].disabled` | `bool` |  |  |  |
| `spec.services[].alerts[].destinations` | `DigitalOceanAppAlertDestinations` |  |  |  |
| `spec.services[].alerts[].destinations.emails` | `[]string` |  |  |  |
| `spec.services[].alerts[].destinations.slackWebhooks` | `[]DigitalOceanAppSlackWebhook` |  |  |  |
| `spec.services[].alerts[].destinations.slackWebhooks[].channel` | `string` | yes |  |  |
| `spec.services[].alerts[].destinations.slackWebhooks[].url` | `string` (sensitive) | yes |  |  |
| `spec.services[].logDestinations` | `[]DigitalOceanAppLogDestination` |  |  |  |
| `spec.services[].logDestinations[].name` | `string` | yes |  |  |
| `spec.services[].logDestinations[].papertrail` | `DigitalOceanAppPapertrailLog` |  |  |  |
| `spec.services[].logDestinations[].papertrail.endpoint` | `string` | yes |  |  |
| `spec.services[].logDestinations[].datadog` | `DigitalOceanAppDatadogLog` |  |  |  |
| `spec.services[].logDestinations[].datadog.apiKey` | `string` (sensitive) | yes |  |  |
| `spec.services[].logDestinations[].datadog.endpoint` | `string` |  |  |  |
| `spec.services[].logDestinations[].logtail` | `DigitalOceanAppLogtailLog` |  |  |  |
| `spec.services[].logDestinations[].logtail.token` | `string` (sensitive) | yes |  |  |
| `spec.services[].logDestinations[].openSearch` | `DigitalOceanAppOpenSearchLog` |  |  |  |
| `spec.services[].logDestinations[].openSearch.endpoint` | `string` |  |  |  |
| `spec.services[].logDestinations[].openSearch.indexName` | `string` |  |  |  |
| `spec.services[].logDestinations[].openSearch.clusterName` | `string` |  |  |  |
| `spec.services[].logDestinations[].openSearch.basicAuth` | `DigitalOceanAppOpenSearchBasicAuth` |  |  |  |
| `spec.services[].logDestinations[].openSearch.basicAuth.user` | `string` |  |  |  |
| `spec.services[].logDestinations[].openSearch.basicAuth.password` | `string` (sensitive) |  |  |  |
| `spec.workers` | `[]DigitalOceanAppWorker` |  |  |  |
| `spec.workers[].name` | `string` | yes |  |  |
| `spec.workers[].sourceDir` | `string` |  |  |  |
| `spec.workers[].git` | `DigitalOceanAppGitSource` |  |  |  |
| `spec.workers[].git.repoCloneUrl` | `string` | yes |  |  |
| `spec.workers[].git.branch` | `string` | yes |  |  |
| `spec.workers[].github` | `DigitalOceanAppGithubSource` |  |  |  |
| `spec.workers[].github.repo` | `string` | yes |  |  |
| `spec.workers[].github.branch` | `string` | yes |  |  |
| `spec.workers[].github.deployOnPush` | `bool` |  |  |  |
| `spec.workers[].gitlab` | `DigitalOceanAppGitlabSource` |  |  |  |
| `spec.workers[].gitlab.repo` | `string` | yes |  |  |
| `spec.workers[].gitlab.branch` | `string` | yes |  |  |
| `spec.workers[].gitlab.deployOnPush` | `bool` |  |  |  |
| `spec.workers[].bitbucket` | `DigitalOceanAppBitbucketSource` |  |  |  |
| `spec.workers[].bitbucket.repo` | `string` | yes |  |  |
| `spec.workers[].bitbucket.branch` | `string` | yes |  |  |
| `spec.workers[].bitbucket.deployOnPush` | `bool` |  |  |  |
| `spec.workers[].image` | `DigitalOceanAppImageSource` |  |  |  |
| `spec.workers[].image.registryType` | `enum` | yes |  |  |
| `spec.workers[].image.registry` | `string` |  |  |  |
| `spec.workers[].image.repository` | `string` | yes |  |  |
| `spec.workers[].image.tag` | `string` |  |  |  |
| `spec.workers[].image.digest` | `string` |  |  |  |
| `spec.workers[].image.registryCredentials` | `string` (sensitive) |  |  |  |
| `spec.workers[].image.deployOnPush` | `bool` |  |  |  |
| `spec.workers[].environmentSlug` | `string` |  |  |  |
| `spec.workers[].dockerfilePath` | `string` |  |  |  |
| `spec.workers[].buildCommand` | `string` |  |  |  |
| `spec.workers[].runCommand` | `string` |  |  |  |
| `spec.workers[].instanceSizeSlug` | `string` |  | `basic-xxs` |  |
| `spec.workers[].instanceCount` | `uint32` |  | `1` |  |
| `spec.workers[].livenessHealthCheck` | `DigitalOceanAppHealthCheck` |  |  |  |
| `spec.workers[].livenessHealthCheck.port` | `uint32` |  |  |  |
| `spec.workers[].livenessHealthCheck.httpPath` | `string` |  |  |  |
| `spec.workers[].livenessHealthCheck.initialDelaySeconds` | `uint32` |  |  |  |
| `spec.workers[].livenessHealthCheck.periodSeconds` | `uint32` |  |  |  |
| `spec.workers[].livenessHealthCheck.timeoutSeconds` | `uint32` |  |  |  |
| `spec.workers[].livenessHealthCheck.successThreshold` | `uint32` |  |  |  |
| `spec.workers[].livenessHealthCheck.failureThreshold` | `uint32` |  |  |  |
| `spec.workers[].autoscaling` | `DigitalOceanAppAutoscaling` |  |  |  |
| `spec.workers[].autoscaling.minInstanceCount` | `uint32` | yes |  |  |
| `spec.workers[].autoscaling.maxInstanceCount` | `uint32` | yes |  |  |
| `spec.workers[].autoscaling.cpuPercent` | `uint32` | yes | `80` |  |
| `spec.workers[].termination` | `DigitalOceanAppTermination` |  |  |  |
| `spec.workers[].termination.gracePeriodSeconds` | `uint32` |  |  |  |
| `spec.workers[].termination.drainSeconds` | `uint32` |  |  |  |
| `spec.workers[].envs` | `[]DigitalOceanAppEnvVar` |  |  |  |
| `spec.workers[].envs[].key` | `string` | yes |  |  |
| `spec.workers[].envs[].plaintext` | `string` |  |  |  |
| `spec.workers[].envs[].secret` | `string` (sensitive) |  |  |  |
| `spec.workers[].envs[].scope` | `enum` |  |  |  |
| `spec.workers[].alerts` | `[]DigitalOceanAppComponentAlert` |  |  |  |
| `spec.workers[].alerts[].rule` | `enum` | yes |  |  |
| `spec.workers[].alerts[].operator` | `enum` | yes |  |  |
| `spec.workers[].alerts[].window` | `enum` | yes |  |  |
| `spec.workers[].alerts[].value` | `double` |  |  |  |
| `spec.workers[].alerts[].disabled` | `bool` |  |  |  |
| `spec.workers[].alerts[].destinations` | `DigitalOceanAppAlertDestinations` |  |  |  |
| `spec.workers[].alerts[].destinations.emails` | `[]string` |  |  |  |
| `spec.workers[].alerts[].destinations.slackWebhooks` | `[]DigitalOceanAppSlackWebhook` |  |  |  |
| `spec.workers[].alerts[].destinations.slackWebhooks[].channel` | `string` | yes |  |  |
| `spec.workers[].alerts[].destinations.slackWebhooks[].url` | `string` (sensitive) | yes |  |  |
| `spec.workers[].logDestinations` | `[]DigitalOceanAppLogDestination` |  |  |  |
| `spec.workers[].logDestinations[].name` | `string` | yes |  |  |
| `spec.workers[].logDestinations[].papertrail` | `DigitalOceanAppPapertrailLog` |  |  |  |
| `spec.workers[].logDestinations[].papertrail.endpoint` | `string` | yes |  |  |
| `spec.workers[].logDestinations[].datadog` | `DigitalOceanAppDatadogLog` |  |  |  |
| `spec.workers[].logDestinations[].datadog.apiKey` | `string` (sensitive) | yes |  |  |
| `spec.workers[].logDestinations[].datadog.endpoint` | `string` |  |  |  |
| `spec.workers[].logDestinations[].logtail` | `DigitalOceanAppLogtailLog` |  |  |  |
| `spec.workers[].logDestinations[].logtail.token` | `string` (sensitive) | yes |  |  |
| `spec.workers[].logDestinations[].openSearch` | `DigitalOceanAppOpenSearchLog` |  |  |  |
| `spec.workers[].logDestinations[].openSearch.endpoint` | `string` |  |  |  |
| `spec.workers[].logDestinations[].openSearch.indexName` | `string` |  |  |  |
| `spec.workers[].logDestinations[].openSearch.clusterName` | `string` |  |  |  |
| `spec.workers[].logDestinations[].openSearch.basicAuth` | `DigitalOceanAppOpenSearchBasicAuth` |  |  |  |
| `spec.workers[].logDestinations[].openSearch.basicAuth.user` | `string` |  |  |  |
| `spec.workers[].logDestinations[].openSearch.basicAuth.password` | `string` (sensitive) |  |  |  |
| `spec.jobs` | `[]DigitalOceanAppJob` |  |  |  |
| `spec.jobs[].name` | `string` | yes |  |  |
| `spec.jobs[].sourceDir` | `string` |  |  |  |
| `spec.jobs[].git` | `DigitalOceanAppGitSource` |  |  |  |
| `spec.jobs[].git.repoCloneUrl` | `string` | yes |  |  |
| `spec.jobs[].git.branch` | `string` | yes |  |  |
| `spec.jobs[].github` | `DigitalOceanAppGithubSource` |  |  |  |
| `spec.jobs[].github.repo` | `string` | yes |  |  |
| `spec.jobs[].github.branch` | `string` | yes |  |  |
| `spec.jobs[].github.deployOnPush` | `bool` |  |  |  |
| `spec.jobs[].gitlab` | `DigitalOceanAppGitlabSource` |  |  |  |
| `spec.jobs[].gitlab.repo` | `string` | yes |  |  |
| `spec.jobs[].gitlab.branch` | `string` | yes |  |  |
| `spec.jobs[].gitlab.deployOnPush` | `bool` |  |  |  |
| `spec.jobs[].bitbucket` | `DigitalOceanAppBitbucketSource` |  |  |  |
| `spec.jobs[].bitbucket.repo` | `string` | yes |  |  |
| `spec.jobs[].bitbucket.branch` | `string` | yes |  |  |
| `spec.jobs[].bitbucket.deployOnPush` | `bool` |  |  |  |
| `spec.jobs[].image` | `DigitalOceanAppImageSource` |  |  |  |
| `spec.jobs[].image.registryType` | `enum` | yes |  |  |
| `spec.jobs[].image.registry` | `string` |  |  |  |
| `spec.jobs[].image.repository` | `string` | yes |  |  |
| `spec.jobs[].image.tag` | `string` |  |  |  |
| `spec.jobs[].image.digest` | `string` |  |  |  |
| `spec.jobs[].image.registryCredentials` | `string` (sensitive) |  |  |  |
| `spec.jobs[].image.deployOnPush` | `bool` |  |  |  |
| `spec.jobs[].environmentSlug` | `string` |  |  |  |
| `spec.jobs[].dockerfilePath` | `string` |  |  |  |
| `spec.jobs[].buildCommand` | `string` |  |  |  |
| `spec.jobs[].runCommand` | `string` |  |  |  |
| `spec.jobs[].instanceSizeSlug` | `string` |  | `basic-xxs` |  |
| `spec.jobs[].instanceCount` | `uint32` |  | `1` |  |
| `spec.jobs[].kind` | `enum` |  |  |  |
| `spec.jobs[].termination` | `DigitalOceanAppTermination` |  |  |  |
| `spec.jobs[].termination.gracePeriodSeconds` | `uint32` |  |  |  |
| `spec.jobs[].termination.drainSeconds` | `uint32` |  |  |  |
| `spec.jobs[].envs` | `[]DigitalOceanAppEnvVar` |  |  |  |
| `spec.jobs[].envs[].key` | `string` | yes |  |  |
| `spec.jobs[].envs[].plaintext` | `string` |  |  |  |
| `spec.jobs[].envs[].secret` | `string` (sensitive) |  |  |  |
| `spec.jobs[].envs[].scope` | `enum` |  |  |  |
| `spec.jobs[].alerts` | `[]DigitalOceanAppComponentAlert` |  |  |  |
| `spec.jobs[].alerts[].rule` | `enum` | yes |  |  |
| `spec.jobs[].alerts[].operator` | `enum` | yes |  |  |
| `spec.jobs[].alerts[].window` | `enum` | yes |  |  |
| `spec.jobs[].alerts[].value` | `double` |  |  |  |
| `spec.jobs[].alerts[].disabled` | `bool` |  |  |  |
| `spec.jobs[].alerts[].destinations` | `DigitalOceanAppAlertDestinations` |  |  |  |
| `spec.jobs[].alerts[].destinations.emails` | `[]string` |  |  |  |
| `spec.jobs[].alerts[].destinations.slackWebhooks` | `[]DigitalOceanAppSlackWebhook` |  |  |  |
| `spec.jobs[].alerts[].destinations.slackWebhooks[].channel` | `string` | yes |  |  |
| `spec.jobs[].alerts[].destinations.slackWebhooks[].url` | `string` (sensitive) | yes |  |  |
| `spec.jobs[].logDestinations` | `[]DigitalOceanAppLogDestination` |  |  |  |
| `spec.jobs[].logDestinations[].name` | `string` | yes |  |  |
| `spec.jobs[].logDestinations[].papertrail` | `DigitalOceanAppPapertrailLog` |  |  |  |
| `spec.jobs[].logDestinations[].papertrail.endpoint` | `string` | yes |  |  |
| `spec.jobs[].logDestinations[].datadog` | `DigitalOceanAppDatadogLog` |  |  |  |
| `spec.jobs[].logDestinations[].datadog.apiKey` | `string` (sensitive) | yes |  |  |
| `spec.jobs[].logDestinations[].datadog.endpoint` | `string` |  |  |  |
| `spec.jobs[].logDestinations[].logtail` | `DigitalOceanAppLogtailLog` |  |  |  |
| `spec.jobs[].logDestinations[].logtail.token` | `string` (sensitive) | yes |  |  |
| `spec.jobs[].logDestinations[].openSearch` | `DigitalOceanAppOpenSearchLog` |  |  |  |
| `spec.jobs[].logDestinations[].openSearch.endpoint` | `string` |  |  |  |
| `spec.jobs[].logDestinations[].openSearch.indexName` | `string` |  |  |  |
| `spec.jobs[].logDestinations[].openSearch.clusterName` | `string` |  |  |  |
| `spec.jobs[].logDestinations[].openSearch.basicAuth` | `DigitalOceanAppOpenSearchBasicAuth` |  |  |  |
| `spec.jobs[].logDestinations[].openSearch.basicAuth.user` | `string` |  |  |  |
| `spec.jobs[].logDestinations[].openSearch.basicAuth.password` | `string` (sensitive) |  |  |  |
| `spec.staticSites` | `[]DigitalOceanAppStaticSite` |  |  |  |
| `spec.staticSites[].name` | `string` | yes |  |  |
| `spec.staticSites[].sourceDir` | `string` |  |  |  |
| `spec.staticSites[].git` | `DigitalOceanAppGitSource` |  |  |  |
| `spec.staticSites[].git.repoCloneUrl` | `string` | yes |  |  |
| `spec.staticSites[].git.branch` | `string` | yes |  |  |
| `spec.staticSites[].github` | `DigitalOceanAppGithubSource` |  |  |  |
| `spec.staticSites[].github.repo` | `string` | yes |  |  |
| `spec.staticSites[].github.branch` | `string` | yes |  |  |
| `spec.staticSites[].github.deployOnPush` | `bool` |  |  |  |
| `spec.staticSites[].gitlab` | `DigitalOceanAppGitlabSource` |  |  |  |
| `spec.staticSites[].gitlab.repo` | `string` | yes |  |  |
| `spec.staticSites[].gitlab.branch` | `string` | yes |  |  |
| `spec.staticSites[].gitlab.deployOnPush` | `bool` |  |  |  |
| `spec.staticSites[].bitbucket` | `DigitalOceanAppBitbucketSource` |  |  |  |
| `spec.staticSites[].bitbucket.repo` | `string` | yes |  |  |
| `spec.staticSites[].bitbucket.branch` | `string` | yes |  |  |
| `spec.staticSites[].bitbucket.deployOnPush` | `bool` |  |  |  |
| `spec.staticSites[].environmentSlug` | `string` |  |  |  |
| `spec.staticSites[].dockerfilePath` | `string` |  |  |  |
| `spec.staticSites[].buildCommand` | `string` |  |  |  |
| `spec.staticSites[].outputDir` | `string` |  |  |  |
| `spec.staticSites[].indexDocument` | `string` |  |  |  |
| `spec.staticSites[].errorDocument` | `string` |  |  |  |
| `spec.staticSites[].catchallDocument` | `string` |  |  |  |
| `spec.staticSites[].envs` | `[]DigitalOceanAppEnvVar` |  |  |  |
| `spec.staticSites[].envs[].key` | `string` | yes |  |  |
| `spec.staticSites[].envs[].plaintext` | `string` |  |  |  |
| `spec.staticSites[].envs[].secret` | `string` (sensitive) |  |  |  |
| `spec.staticSites[].envs[].scope` | `enum` |  |  |  |
| `spec.functions` | `[]DigitalOceanAppFunctionComponent` |  |  |  |
| `spec.functions[].name` | `string` | yes |  |  |
| `spec.functions[].sourceDir` | `string` |  |  |  |
| `spec.functions[].git` | `DigitalOceanAppGitSource` |  |  |  |
| `spec.functions[].git.repoCloneUrl` | `string` | yes |  |  |
| `spec.functions[].git.branch` | `string` | yes |  |  |
| `spec.functions[].github` | `DigitalOceanAppGithubSource` |  |  |  |
| `spec.functions[].github.repo` | `string` | yes |  |  |
| `spec.functions[].github.branch` | `string` | yes |  |  |
| `spec.functions[].github.deployOnPush` | `bool` |  |  |  |
| `spec.functions[].gitlab` | `DigitalOceanAppGitlabSource` |  |  |  |
| `spec.functions[].gitlab.repo` | `string` | yes |  |  |
| `spec.functions[].gitlab.branch` | `string` | yes |  |  |
| `spec.functions[].gitlab.deployOnPush` | `bool` |  |  |  |
| `spec.functions[].bitbucket` | `DigitalOceanAppBitbucketSource` |  |  |  |
| `spec.functions[].bitbucket.repo` | `string` | yes |  |  |
| `spec.functions[].bitbucket.branch` | `string` | yes |  |  |
| `spec.functions[].bitbucket.deployOnPush` | `bool` |  |  |  |
| `spec.functions[].envs` | `[]DigitalOceanAppEnvVar` |  |  |  |
| `spec.functions[].envs[].key` | `string` | yes |  |  |
| `spec.functions[].envs[].plaintext` | `string` |  |  |  |
| `spec.functions[].envs[].secret` | `string` (sensitive) |  |  |  |
| `spec.functions[].envs[].scope` | `enum` |  |  |  |
| `spec.functions[].alerts` | `[]DigitalOceanAppComponentAlert` |  |  |  |
| `spec.functions[].alerts[].rule` | `enum` | yes |  |  |
| `spec.functions[].alerts[].operator` | `enum` | yes |  |  |
| `spec.functions[].alerts[].window` | `enum` | yes |  |  |
| `spec.functions[].alerts[].value` | `double` |  |  |  |
| `spec.functions[].alerts[].disabled` | `bool` |  |  |  |
| `spec.functions[].alerts[].destinations` | `DigitalOceanAppAlertDestinations` |  |  |  |
| `spec.functions[].alerts[].destinations.emails` | `[]string` |  |  |  |
| `spec.functions[].alerts[].destinations.slackWebhooks` | `[]DigitalOceanAppSlackWebhook` |  |  |  |
| `spec.functions[].alerts[].destinations.slackWebhooks[].channel` | `string` | yes |  |  |
| `spec.functions[].alerts[].destinations.slackWebhooks[].url` | `string` (sensitive) | yes |  |  |
| `spec.functions[].logDestinations` | `[]DigitalOceanAppLogDestination` |  |  |  |
| `spec.functions[].logDestinations[].name` | `string` | yes |  |  |
| `spec.functions[].logDestinations[].papertrail` | `DigitalOceanAppPapertrailLog` |  |  |  |
| `spec.functions[].logDestinations[].papertrail.endpoint` | `string` | yes |  |  |
| `spec.functions[].logDestinations[].datadog` | `DigitalOceanAppDatadogLog` |  |  |  |
| `spec.functions[].logDestinations[].datadog.apiKey` | `string` (sensitive) | yes |  |  |
| `spec.functions[].logDestinations[].datadog.endpoint` | `string` |  |  |  |
| `spec.functions[].logDestinations[].logtail` | `DigitalOceanAppLogtailLog` |  |  |  |
| `spec.functions[].logDestinations[].logtail.token` | `string` (sensitive) | yes |  |  |
| `spec.functions[].logDestinations[].openSearch` | `DigitalOceanAppOpenSearchLog` |  |  |  |
| `spec.functions[].logDestinations[].openSearch.endpoint` | `string` |  |  |  |
| `spec.functions[].logDestinations[].openSearch.indexName` | `string` |  |  |  |
| `spec.functions[].logDestinations[].openSearch.clusterName` | `string` |  |  |  |
| `spec.functions[].logDestinations[].openSearch.basicAuth` | `DigitalOceanAppOpenSearchBasicAuth` |  |  |  |
| `spec.functions[].logDestinations[].openSearch.basicAuth.user` | `string` |  |  |  |
| `spec.functions[].logDestinations[].openSearch.basicAuth.password` | `string` (sensitive) |  |  |  |
| `spec.databases` | `[]DigitalOceanAppDatabase` |  |  |  |
| `spec.databases[].name` | `string` |  |  |  |
| `spec.databases[].engine` | `enum` |  |  |  |
| `spec.databases[].version` | `string` |  |  |  |
| `spec.databases[].production` | `bool` |  |  |  |
| `spec.databases[].clusterName` | `string \| valueFrom` |  |  | DigitalOceanDatabaseCluster (`spec.cluster_name`) |
| `spec.databases[].dbName` | `string` |  |  |  |
| `spec.databases[].dbUser` | `string` |  |  |  |
| `spec.domains` | `[]DigitalOceanAppDomain` |  |  |  |
| `spec.domains[].name` | `string` | yes |  |  |
| `spec.domains[].wildcard` | `bool` |  |  |  |
| `spec.domains[].zone` | `string \| valueFrom` |  |  | DigitalOceanDnsZone (`status.outputs.zone_name`) |
| `spec.domains[].type` | `string` |  |  |  |
| `spec.envs` | `[]DigitalOceanAppEnvVar` |  |  |  |
| `spec.envs[].key` | `string` | yes |  |  |
| `spec.envs[].plaintext` | `string` |  |  |  |
| `spec.envs[].secret` | `string` (sensitive) |  |  |  |
| `spec.envs[].scope` | `enum` |  |  |  |
| `spec.alerts` | `[]DigitalOceanAppAlert` |  |  |  |
| `spec.alerts[].rule` | `enum` | yes |  |  |
| `spec.alerts[].disabled` | `bool` |  |  |  |
| `spec.alerts[].destinations` | `DigitalOceanAppAlertDestinations` |  |  |  |
| `spec.alerts[].destinations.emails` | `[]string` |  |  |  |
| `spec.alerts[].destinations.slackWebhooks` | `[]DigitalOceanAppSlackWebhook` |  |  |  |
| `spec.alerts[].destinations.slackWebhooks[].channel` | `string` | yes |  |  |
| `spec.alerts[].destinations.slackWebhooks[].url` | `string` (sensitive) | yes |  |  |
| `spec.ingress` | `DigitalOceanAppIngress` |  |  |  |
| `spec.ingress.rules` | `[]DigitalOceanAppIngressRule` |  |  |  |
| `spec.ingress.rules[].match` | `DigitalOceanAppIngressMatch` |  |  |  |
| `spec.ingress.rules[].match.pathPrefix` | `string` |  |  |  |
| `spec.ingress.rules[].match.authorityExact` | `string` |  |  |  |
| `spec.ingress.rules[].component` | `DigitalOceanAppIngressComponent` |  |  |  |
| `spec.ingress.rules[].component.name` | `string` | yes |  |  |
| `spec.ingress.rules[].component.preservePathPrefix` | `bool` |  |  |  |
| `spec.ingress.rules[].component.rewrite` | `string` |  |  |  |
| `spec.ingress.rules[].redirect` | `DigitalOceanAppIngressRedirect` |  |  |  |
| `spec.ingress.rules[].redirect.uri` | `string` |  |  |  |
| `spec.ingress.rules[].redirect.authority` | `string` |  |  |  |
| `spec.ingress.rules[].redirect.port` | `uint32` |  |  |  |
| `spec.ingress.rules[].redirect.scheme` | `string` |  |  |  |
| `spec.ingress.rules[].redirect.redirectCode` | `uint32` |  | `302` |  |
| `spec.ingress.rules[].cors` | `DigitalOceanAppCors` |  |  |  |
| `spec.ingress.rules[].cors.allowOrigins` | `DigitalOceanAppCorsAllowOrigins` |  |  |  |
| `spec.ingress.rules[].cors.allowOrigins.exact` | `string` |  |  |  |
| `spec.ingress.rules[].cors.allowOrigins.regex` | `string` |  |  |  |
| `spec.ingress.rules[].cors.allowMethods` | `[]string` |  |  |  |
| `spec.ingress.rules[].cors.allowHeaders` | `[]string` |  |  |  |
| `spec.ingress.rules[].cors.exposeHeaders` | `[]string` |  |  |  |
| `spec.ingress.rules[].cors.maxAge` | `string` |  |  |  |
| `spec.ingress.rules[].cors.allowCredentials` | `bool` |  |  |  |
| `spec.ingress.secureHeader` | `DigitalOceanAppSecureHeader` |  |  |  |
| `spec.ingress.secureHeader.key` | `string` | yes |  |  |
| `spec.ingress.secureHeader.value` | `string` | yes |  |  |
| `spec.egress` | `enum` |  |  |  |
| `spec.maintenance` | `DigitalOceanAppMaintenance` |  |  |  |
| `spec.maintenance.enabled` | `bool` |  |  |  |
| `spec.maintenance.archive` | `bool` |  |  |  |
| `spec.maintenance.offlinePageUrl` | `string` |  |  |  |
| `spec.vpc` | `string \| valueFrom` |  |  | DigitalOceanVpc (`status.outputs.vpc_id`) |
| `spec.features` | `[]string` |  |  |  |
| `spec.disableEdgeCache` | `bool` |  |  |  |
| `spec.disableEmailObfuscation` | `bool` |  |  |  |
| `spec.enhancedThreatControlEnabled` | `bool` |  |  |  |
| `spec.projectId` | `string \| valueFrom` |  |  | DigitalOceanProject (`status.outputs.project_id`) |

## Field Details

### spec.appName

`string` · required

App name, unique across every app in the DigitalOcean account. 2-32
characters matching ^[a-z][a-z0-9-]{0,30}[a-z0-9]$ -- starts with a
letter, ends with a letter or digit (the API's own rule; it answers
"name in body should match ..." otherwise, and reports a taken name as
app_name_available: false on its validate-only POST /v2/apps/propose).
This is spec.name in the Terraform resource. Renaming updates the app in
place; the default <name>-<hash>.ondigitalocean.app hostname carries the
name, so the app's URL changes with it.

- rule: {"required":true,"string":{"minLen":"2","maxLen":"32","pattern":"^[a-z][a-z0-9-]{0,30}[a-z0-9]$"}}

### spec.region

`enum` · required

App Platform region group, for example nyc (never a droplet slug such
as nyc3 -- the API would store nyc and the plan would never settle).
Required: the API defaults to blr when omitted, which is rarely what a
manifest author means.

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

### spec.services

`[]DigitalOceanAppService`

- rule: set exactly one source for this service: git, github, gitlab, bitbucket, or image
- rule: when autoscaling is set, leave instance_count unset - App Platform ignores a fixed count while autoscaling is on

### spec.services[].name

`string` · required

Component name: the API's rule is ^[a-z][a-z0-9-]{0,30}[a-z0-9]$ (2-32
chars, starts with a letter) and a violation rejects the whole app spec.

- rule: {"required":true,"string":{"minLen":"2","maxLen":"32","pattern":"^[a-z][a-z0-9-]{0,30}[a-z0-9]$"}}

### spec.services[].sourceDir

`string`

Working directory inside the repo for the build.

### spec.services[].git

`DigitalOceanAppGitSource`

### spec.services[].git.repoCloneUrl

`string` · required

HTTPS or git clone URL, for example https://github.com/example/app.git

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.services[].git.branch

`string` · required

Branch to deploy. Example: main

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.services[].github

`DigitalOceanAppGithubSource`

### spec.services[].github.repo

`string` · required

Repository in owner/repo form, for example plantonhq/demo

- rule: {"required":true,"string":{"pattern":"^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$"}}

### spec.services[].github.branch

`string` · required

Branch to deploy. Example: main

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.services[].github.deployOnPush

`bool`

Redeploy automatically when this branch is pushed.

### spec.services[].gitlab

`DigitalOceanAppGitlabSource`

### spec.services[].gitlab.repo

`string` · required

- rule: {"required":true,"string":{"pattern":"^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$"}}

### spec.services[].gitlab.branch

`string` · required

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.services[].gitlab.deployOnPush

`bool`

### spec.services[].bitbucket

`DigitalOceanAppBitbucketSource`

### spec.services[].bitbucket.repo

`string` · required

- rule: {"required":true,"string":{"pattern":"^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$"}}

### spec.services[].bitbucket.branch

`string` · required

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.services[].bitbucket.deployOnPush

`bool`

### spec.services[].image

`DigitalOceanAppImageSource`

- rule: set tag or digest, not both - App Platform treats them as mutually exclusive image pins
- rule: leave registry empty when registry_type is docr - DigitalOcean Container Registry does not take a registry hostname here

### spec.services[].image.registryType

`enum` · required

Which registry hosts the image. Required.

With docr the image is pulled from the account's own DigitalOcean
Container Registry; App Platform resolves it itself, so no field here
names the registry and no reference is wired. To say that the app
depends on a DigitalOceanContainerRegistry that Planton manages -- so
the registry deploys first and the dependency is drawn -- declare it
under metadata.relationships with type `uses`.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `digital_ocean_app_registry_type_unspecified`
- `docker_hub`
- `docr`
- `ghcr`

### spec.services[].image.registry

`string`

Registry hostname. Required for docker_hub and ghcr (for example ghcr.io
or a Docker Hub namespace). Must be empty for docr -- DigitalOcean's own
registry is addressed by the account, not by a hostname (see
registry_type). A hostname is not a Planton resource, so this is a
plain string, never a reference.

### spec.services[].image.repository

`string` · required

Repository name inside the registry, for example myapp/api

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.services[].image.tag

`string`

Image tag. Defaults to latest on the server when both tag and digest are
omitted. Mutually exclusive with digest.

### spec.services[].image.digest

`string`

Image digest (sha256:...). Mutually exclusive with tag.

### spec.services[].image.registryCredentials

`string` · sensitive

Credentials for a private docker_hub or ghcr registry. Not used for docr
(App Platform uses the account's DigitalOcean Container Registry access).

### spec.services[].image.deployOnPush

`bool`

Redeploy when a new image matching this repository is pushed. Only
honored for docr and ghcr.

### spec.services[].environmentSlug

`string`

Runtime identifier App Platform auto-detects when omitted, for example
node-js or go.

### spec.services[].dockerfilePath

`string`

### spec.services[].buildCommand

`string`

### spec.services[].runCommand

`string`

### spec.services[].instanceSizeSlug

`string`

Instance size slug, for example basic-xxs, basic-s, professional-xs.
Current slugs are listed in the DigitalOcean App Platform docs; the
provider does not validate the value.

- default: `basic-xxs`

### spec.services[].instanceCount

`uint32`

Fixed instance count when autoscaling is off. Default 1.

- default: `1`

### spec.services[].httpPort

`uint32` · optional (explicit presence)

- rule: {"uint32":{"lte":65535,"gte":1}}

### spec.services[].internalPorts

`[]uint32`

### spec.services[].healthCheck

`DigitalOceanAppHealthCheck`

### spec.services[].healthCheck.port

`uint32` · optional (explicit presence)

Port to probe. Range 1-65535. When omitted, App Platform uses the
component's HTTP port.

- rule: {"uint32":{"lte":65535,"gte":1}}

### spec.services[].healthCheck.httpPath

`string`

HTTP path, for example /healthz. When omitted the probe is TCP.

### spec.services[].healthCheck.initialDelaySeconds

`uint32` · optional (explicit presence)

### spec.services[].healthCheck.periodSeconds

`uint32` · optional (explicit presence)

### spec.services[].healthCheck.timeoutSeconds

`uint32` · optional (explicit presence)

### spec.services[].healthCheck.successThreshold

`uint32` · optional (explicit presence)

### spec.services[].healthCheck.failureThreshold

`uint32` · optional (explicit presence)

### spec.services[].livenessHealthCheck

`DigitalOceanAppHealthCheck`

Liveness probe: a failing check restarts the container, where the
readiness health_check only gates traffic. Both provisioners deploy it.

### spec.services[].livenessHealthCheck.port

`uint32` · optional (explicit presence)

Port to probe. Range 1-65535. When omitted, App Platform uses the
component's HTTP port.

- rule: {"uint32":{"lte":65535,"gte":1}}

### spec.services[].livenessHealthCheck.httpPath

`string`

HTTP path, for example /healthz. When omitted the probe is TCP.

### spec.services[].livenessHealthCheck.initialDelaySeconds

`uint32` · optional (explicit presence)

### spec.services[].livenessHealthCheck.periodSeconds

`uint32` · optional (explicit presence)

### spec.services[].livenessHealthCheck.timeoutSeconds

`uint32` · optional (explicit presence)

### spec.services[].livenessHealthCheck.successThreshold

`uint32` · optional (explicit presence)

### spec.services[].livenessHealthCheck.failureThreshold

`uint32` · optional (explicit presence)

### spec.services[].autoscaling

`DigitalOceanAppAutoscaling`

- rule: min_instance_count must be less than max_instance_count

### spec.services[].autoscaling.minInstanceCount

`uint32` · required

Smallest number of instances while autoscaling. At least 1.

- rule: {"required":true,"uint32":{"gte":1}}

### spec.services[].autoscaling.maxInstanceCount

`uint32` · required

Largest number of instances while autoscaling. At least 1, and greater
than min_instance_count.

- rule: {"required":true,"uint32":{"gte":1}}

### spec.services[].autoscaling.cpuPercent

`uint32` · required

Target average CPU percent (1-100) that autoscaling aims for. The
provider requires this leaf inside metrics.cpu.

- default: `80`
- rule: {"required":true,"uint32":{"lte":100,"gte":1}}

### spec.services[].termination

`DigitalOceanAppTermination`

### spec.services[].termination.gracePeriodSeconds

`uint32` · optional (explicit presence)

Seconds to wait for in-flight work after SIGTERM. Range 1-600. Server
default is 120 when omitted.

- rule: {"uint32":{"lte":600,"gte":1}}

### spec.services[].termination.drainSeconds

`uint32` · optional (explicit presence)

Seconds to drain HTTP connections on a service before SIGTERM. Range
1-110. Server default is 15. Ignored (and rejected by CEL on workers/jobs)
unless the component is a service.

- rule: {"uint32":{"lte":110,"gte":1}}

### spec.services[].envs

`[]DigitalOceanAppEnvVar`

- rule: set either plaintext or secret for this environment variable - App Platform needs a value

### spec.services[].envs[].key

`string` · required

Variable name, for example DATABASE_URL or NODE_ENV.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.services[].envs[].plaintext

`string`

Non-secret value. Visible in the App Platform UI and in build logs.

### spec.services[].envs[].secret

`string` · sensitive

Secret value (API keys, database URLs, tokens). Stored in App Platform's
secret store; the IaC modules send type=SECRET.

### spec.services[].envs[].scope

`enum`

When the variable is injected. Omit to use run_and_build_time.

Allowed values (use exactly as shown):

- `digital_ocean_app_env_scope_unspecified`
- `run_and_build_time` -- Injected during the build and at runtime (provider default).
- `run_time` -- Injected only at runtime.
- `build_time` -- Injected only during the build.
- `unset` -- Provider UNSET - treated as run_and_build_time by the API.

### spec.services[].alerts

`[]DigitalOceanAppComponentAlert`

### spec.services[].alerts[].rule

`enum` · required

- rule: {"required":true}

Allowed values (use exactly as shown):

- `digital_ocean_app_component_alert_rule_unspecified`
- `cpu_utilization`
- `mem_utilization`
- `restart_count`

### spec.services[].alerts[].operator

`enum` · required

- rule: {"required":true}

Allowed values (use exactly as shown):

- `digital_ocean_app_alert_operator_unspecified`
- `greater_than`
- `less_than`

### spec.services[].alerts[].window

`enum` · required

- rule: {"required":true}

Allowed values (use exactly as shown):

- `digital_ocean_app_alert_window_unspecified`
- `five_minutes`
- `ten_minutes`
- `thirty_minutes`
- `one_hour`

### spec.services[].alerts[].value

`double`

Threshold. For cpu_utilization / mem_utilization this is a percent; for
restart_count it is a count.

- rule: {"double":{"gte":0}}

### spec.services[].alerts[].disabled

`bool`

### spec.services[].alerts[].destinations

`DigitalOceanAppAlertDestinations`

### spec.services[].alerts[].destinations.emails

`[]string`

### spec.services[].alerts[].destinations.slackWebhooks

`[]DigitalOceanAppSlackWebhook`

### spec.services[].alerts[].destinations.slackWebhooks[].channel

`string` · required

- rule: {"required":true}

### spec.services[].alerts[].destinations.slackWebhooks[].url

`string` · required · sensitive

- rule: {"required":true}

### spec.services[].logDestinations

`[]DigitalOceanAppLogDestination`

- rule: set exactly one sink: papertrail, datadog, logtail, or open_search

### spec.services[].logDestinations[].name

`string` · required

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.services[].logDestinations[].papertrail

`DigitalOceanAppPapertrailLog`

### spec.services[].logDestinations[].papertrail.endpoint

`string` · required

Syslog endpoint, for example logs.papertrailapp.com:12345

- rule: {"required":true}

### spec.services[].logDestinations[].datadog

`DigitalOceanAppDatadogLog`

### spec.services[].logDestinations[].datadog.apiKey

`string` · required · sensitive

- rule: {"required":true}

### spec.services[].logDestinations[].datadog.endpoint

`string`

Defaults to https://http-intake.logs.datadoghq.com when omitted.

### spec.services[].logDestinations[].logtail

`DigitalOceanAppLogtailLog`

### spec.services[].logDestinations[].logtail.token

`string` · required · sensitive

- rule: {"required":true}

### spec.services[].logDestinations[].openSearch

`DigitalOceanAppOpenSearchLog`

### spec.services[].logDestinations[].openSearch.endpoint

`string`

### spec.services[].logDestinations[].openSearch.indexName

`string`

### spec.services[].logDestinations[].openSearch.clusterName

`string`

### spec.services[].logDestinations[].openSearch.basicAuth

`DigitalOceanAppOpenSearchBasicAuth`

The provider requires this block even when user and password are empty
(App Platform's OpenSearch integration uses it as a placeholder).

### spec.services[].logDestinations[].openSearch.basicAuth.user

`string`

### spec.services[].logDestinations[].openSearch.basicAuth.password

`string` · sensitive

### spec.workers

`[]DigitalOceanAppWorker`

- rule: set exactly one source for this worker: git, github, gitlab, bitbucket, or image
- rule: when autoscaling is set, leave instance_count unset - App Platform ignores a fixed count while autoscaling is on
- rule: drain_seconds is a service-only HTTP drain; workers only honor grace_period_seconds

### spec.workers[].name

`string` · required

Component name: the API's rule is ^[a-z][a-z0-9-]{0,30}[a-z0-9]$ (2-32
chars, starts with a letter) and a violation rejects the whole app spec.

- rule: {"required":true,"string":{"minLen":"2","maxLen":"32","pattern":"^[a-z][a-z0-9-]{0,30}[a-z0-9]$"}}

### spec.workers[].sourceDir

`string`

### spec.workers[].git

`DigitalOceanAppGitSource`

### spec.workers[].git.repoCloneUrl

`string` · required

HTTPS or git clone URL, for example https://github.com/example/app.git

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.workers[].git.branch

`string` · required

Branch to deploy. Example: main

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.workers[].github

`DigitalOceanAppGithubSource`

### spec.workers[].github.repo

`string` · required

Repository in owner/repo form, for example plantonhq/demo

- rule: {"required":true,"string":{"pattern":"^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$"}}

### spec.workers[].github.branch

`string` · required

Branch to deploy. Example: main

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.workers[].github.deployOnPush

`bool`

Redeploy automatically when this branch is pushed.

### spec.workers[].gitlab

`DigitalOceanAppGitlabSource`

### spec.workers[].gitlab.repo

`string` · required

- rule: {"required":true,"string":{"pattern":"^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$"}}

### spec.workers[].gitlab.branch

`string` · required

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.workers[].gitlab.deployOnPush

`bool`

### spec.workers[].bitbucket

`DigitalOceanAppBitbucketSource`

### spec.workers[].bitbucket.repo

`string` · required

- rule: {"required":true,"string":{"pattern":"^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$"}}

### spec.workers[].bitbucket.branch

`string` · required

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.workers[].bitbucket.deployOnPush

`bool`

### spec.workers[].image

`DigitalOceanAppImageSource`

- rule: set tag or digest, not both - App Platform treats them as mutually exclusive image pins
- rule: leave registry empty when registry_type is docr - DigitalOcean Container Registry does not take a registry hostname here

### spec.workers[].image.registryType

`enum` · required

Which registry hosts the image. Required.

With docr the image is pulled from the account's own DigitalOcean
Container Registry; App Platform resolves it itself, so no field here
names the registry and no reference is wired. To say that the app
depends on a DigitalOceanContainerRegistry that Planton manages -- so
the registry deploys first and the dependency is drawn -- declare it
under metadata.relationships with type `uses`.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `digital_ocean_app_registry_type_unspecified`
- `docker_hub`
- `docr`
- `ghcr`

### spec.workers[].image.registry

`string`

Registry hostname. Required for docker_hub and ghcr (for example ghcr.io
or a Docker Hub namespace). Must be empty for docr -- DigitalOcean's own
registry is addressed by the account, not by a hostname (see
registry_type). A hostname is not a Planton resource, so this is a
plain string, never a reference.

### spec.workers[].image.repository

`string` · required

Repository name inside the registry, for example myapp/api

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.workers[].image.tag

`string`

Image tag. Defaults to latest on the server when both tag and digest are
omitted. Mutually exclusive with digest.

### spec.workers[].image.digest

`string`

Image digest (sha256:...). Mutually exclusive with tag.

### spec.workers[].image.registryCredentials

`string` · sensitive

Credentials for a private docker_hub or ghcr registry. Not used for docr
(App Platform uses the account's DigitalOcean Container Registry access).

### spec.workers[].image.deployOnPush

`bool`

Redeploy when a new image matching this repository is pushed. Only
honored for docr and ghcr.

### spec.workers[].environmentSlug

`string`

### spec.workers[].dockerfilePath

`string`

### spec.workers[].buildCommand

`string`

### spec.workers[].runCommand

`string`

### spec.workers[].instanceSizeSlug

`string`

- default: `basic-xxs`

### spec.workers[].instanceCount

`uint32`

- default: `1`

### spec.workers[].livenessHealthCheck

`DigitalOceanAppHealthCheck`

Liveness probe: a failing check restarts the container. Both provisioners
deploy it.

### spec.workers[].livenessHealthCheck.port

`uint32` · optional (explicit presence)

Port to probe. Range 1-65535. When omitted, App Platform uses the
component's HTTP port.

- rule: {"uint32":{"lte":65535,"gte":1}}

### spec.workers[].livenessHealthCheck.httpPath

`string`

HTTP path, for example /healthz. When omitted the probe is TCP.

### spec.workers[].livenessHealthCheck.initialDelaySeconds

`uint32` · optional (explicit presence)

### spec.workers[].livenessHealthCheck.periodSeconds

`uint32` · optional (explicit presence)

### spec.workers[].livenessHealthCheck.timeoutSeconds

`uint32` · optional (explicit presence)

### spec.workers[].livenessHealthCheck.successThreshold

`uint32` · optional (explicit presence)

### spec.workers[].livenessHealthCheck.failureThreshold

`uint32` · optional (explicit presence)

### spec.workers[].autoscaling

`DigitalOceanAppAutoscaling`

- rule: min_instance_count must be less than max_instance_count

### spec.workers[].autoscaling.minInstanceCount

`uint32` · required

Smallest number of instances while autoscaling. At least 1.

- rule: {"required":true,"uint32":{"gte":1}}

### spec.workers[].autoscaling.maxInstanceCount

`uint32` · required

Largest number of instances while autoscaling. At least 1, and greater
than min_instance_count.

- rule: {"required":true,"uint32":{"gte":1}}

### spec.workers[].autoscaling.cpuPercent

`uint32` · required

Target average CPU percent (1-100) that autoscaling aims for. The
provider requires this leaf inside metrics.cpu.

- default: `80`
- rule: {"required":true,"uint32":{"lte":100,"gte":1}}

### spec.workers[].termination

`DigitalOceanAppTermination`

### spec.workers[].termination.gracePeriodSeconds

`uint32` · optional (explicit presence)

Seconds to wait for in-flight work after SIGTERM. Range 1-600. Server
default is 120 when omitted.

- rule: {"uint32":{"lte":600,"gte":1}}

### spec.workers[].termination.drainSeconds

`uint32` · optional (explicit presence)

Seconds to drain HTTP connections on a service before SIGTERM. Range
1-110. Server default is 15. Ignored (and rejected by CEL on workers/jobs)
unless the component is a service.

- rule: {"uint32":{"lte":110,"gte":1}}

### spec.workers[].envs

`[]DigitalOceanAppEnvVar`

- rule: set either plaintext or secret for this environment variable - App Platform needs a value

### spec.workers[].envs[].key

`string` · required

Variable name, for example DATABASE_URL or NODE_ENV.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.workers[].envs[].plaintext

`string`

Non-secret value. Visible in the App Platform UI and in build logs.

### spec.workers[].envs[].secret

`string` · sensitive

Secret value (API keys, database URLs, tokens). Stored in App Platform's
secret store; the IaC modules send type=SECRET.

### spec.workers[].envs[].scope

`enum`

When the variable is injected. Omit to use run_and_build_time.

Allowed values (use exactly as shown):

- `digital_ocean_app_env_scope_unspecified`
- `run_and_build_time` -- Injected during the build and at runtime (provider default).
- `run_time` -- Injected only at runtime.
- `build_time` -- Injected only during the build.
- `unset` -- Provider UNSET - treated as run_and_build_time by the API.

### spec.workers[].alerts

`[]DigitalOceanAppComponentAlert`

### spec.workers[].alerts[].rule

`enum` · required

- rule: {"required":true}

Allowed values (use exactly as shown):

- `digital_ocean_app_component_alert_rule_unspecified`
- `cpu_utilization`
- `mem_utilization`
- `restart_count`

### spec.workers[].alerts[].operator

`enum` · required

- rule: {"required":true}

Allowed values (use exactly as shown):

- `digital_ocean_app_alert_operator_unspecified`
- `greater_than`
- `less_than`

### spec.workers[].alerts[].window

`enum` · required

- rule: {"required":true}

Allowed values (use exactly as shown):

- `digital_ocean_app_alert_window_unspecified`
- `five_minutes`
- `ten_minutes`
- `thirty_minutes`
- `one_hour`

### spec.workers[].alerts[].value

`double`

Threshold. For cpu_utilization / mem_utilization this is a percent; for
restart_count it is a count.

- rule: {"double":{"gte":0}}

### spec.workers[].alerts[].disabled

`bool`

### spec.workers[].alerts[].destinations

`DigitalOceanAppAlertDestinations`

### spec.workers[].alerts[].destinations.emails

`[]string`

### spec.workers[].alerts[].destinations.slackWebhooks

`[]DigitalOceanAppSlackWebhook`

### spec.workers[].alerts[].destinations.slackWebhooks[].channel

`string` · required

- rule: {"required":true}

### spec.workers[].alerts[].destinations.slackWebhooks[].url

`string` · required · sensitive

- rule: {"required":true}

### spec.workers[].logDestinations

`[]DigitalOceanAppLogDestination`

- rule: set exactly one sink: papertrail, datadog, logtail, or open_search

### spec.workers[].logDestinations[].name

`string` · required

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.workers[].logDestinations[].papertrail

`DigitalOceanAppPapertrailLog`

### spec.workers[].logDestinations[].papertrail.endpoint

`string` · required

Syslog endpoint, for example logs.papertrailapp.com:12345

- rule: {"required":true}

### spec.workers[].logDestinations[].datadog

`DigitalOceanAppDatadogLog`

### spec.workers[].logDestinations[].datadog.apiKey

`string` · required · sensitive

- rule: {"required":true}

### spec.workers[].logDestinations[].datadog.endpoint

`string`

Defaults to https://http-intake.logs.datadoghq.com when omitted.

### spec.workers[].logDestinations[].logtail

`DigitalOceanAppLogtailLog`

### spec.workers[].logDestinations[].logtail.token

`string` · required · sensitive

- rule: {"required":true}

### spec.workers[].logDestinations[].openSearch

`DigitalOceanAppOpenSearchLog`

### spec.workers[].logDestinations[].openSearch.endpoint

`string`

### spec.workers[].logDestinations[].openSearch.indexName

`string`

### spec.workers[].logDestinations[].openSearch.clusterName

`string`

### spec.workers[].logDestinations[].openSearch.basicAuth

`DigitalOceanAppOpenSearchBasicAuth`

The provider requires this block even when user and password are empty
(App Platform's OpenSearch integration uses it as a placeholder).

### spec.workers[].logDestinations[].openSearch.basicAuth.user

`string`

### spec.workers[].logDestinations[].openSearch.basicAuth.password

`string` · sensitive

### spec.jobs

`[]DigitalOceanAppJob`

- rule: set exactly one source for this job: git, github, gitlab, bitbucket, or image
- rule: drain_seconds is a service-only HTTP drain; jobs only honor grace_period_seconds

### spec.jobs[].name

`string` · required

Component name: the API's rule is ^[a-z][a-z0-9-]{0,30}[a-z0-9]$ (2-32
chars, starts with a letter) and a violation rejects the whole app spec.

- rule: {"required":true,"string":{"minLen":"2","maxLen":"32","pattern":"^[a-z][a-z0-9-]{0,30}[a-z0-9]$"}}

### spec.jobs[].sourceDir

`string`

### spec.jobs[].git

`DigitalOceanAppGitSource`

### spec.jobs[].git.repoCloneUrl

`string` · required

HTTPS or git clone URL, for example https://github.com/example/app.git

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.jobs[].git.branch

`string` · required

Branch to deploy. Example: main

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.jobs[].github

`DigitalOceanAppGithubSource`

### spec.jobs[].github.repo

`string` · required

Repository in owner/repo form, for example plantonhq/demo

- rule: {"required":true,"string":{"pattern":"^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$"}}

### spec.jobs[].github.branch

`string` · required

Branch to deploy. Example: main

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.jobs[].github.deployOnPush

`bool`

Redeploy automatically when this branch is pushed.

### spec.jobs[].gitlab

`DigitalOceanAppGitlabSource`

### spec.jobs[].gitlab.repo

`string` · required

- rule: {"required":true,"string":{"pattern":"^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$"}}

### spec.jobs[].gitlab.branch

`string` · required

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.jobs[].gitlab.deployOnPush

`bool`

### spec.jobs[].bitbucket

`DigitalOceanAppBitbucketSource`

### spec.jobs[].bitbucket.repo

`string` · required

- rule: {"required":true,"string":{"pattern":"^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$"}}

### spec.jobs[].bitbucket.branch

`string` · required

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.jobs[].bitbucket.deployOnPush

`bool`

### spec.jobs[].image

`DigitalOceanAppImageSource`

- rule: set tag or digest, not both - App Platform treats them as mutually exclusive image pins
- rule: leave registry empty when registry_type is docr - DigitalOcean Container Registry does not take a registry hostname here

### spec.jobs[].image.registryType

`enum` · required

Which registry hosts the image. Required.

With docr the image is pulled from the account's own DigitalOcean
Container Registry; App Platform resolves it itself, so no field here
names the registry and no reference is wired. To say that the app
depends on a DigitalOceanContainerRegistry that Planton manages -- so
the registry deploys first and the dependency is drawn -- declare it
under metadata.relationships with type `uses`.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `digital_ocean_app_registry_type_unspecified`
- `docker_hub`
- `docr`
- `ghcr`

### spec.jobs[].image.registry

`string`

Registry hostname. Required for docker_hub and ghcr (for example ghcr.io
or a Docker Hub namespace). Must be empty for docr -- DigitalOcean's own
registry is addressed by the account, not by a hostname (see
registry_type). A hostname is not a Planton resource, so this is a
plain string, never a reference.

### spec.jobs[].image.repository

`string` · required

Repository name inside the registry, for example myapp/api

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.jobs[].image.tag

`string`

Image tag. Defaults to latest on the server when both tag and digest are
omitted. Mutually exclusive with digest.

### spec.jobs[].image.digest

`string`

Image digest (sha256:...). Mutually exclusive with tag.

### spec.jobs[].image.registryCredentials

`string` · sensitive

Credentials for a private docker_hub or ghcr registry. Not used for docr
(App Platform uses the account's DigitalOcean Container Registry access).

### spec.jobs[].image.deployOnPush

`bool`

Redeploy when a new image matching this repository is pushed. Only
honored for docr and ghcr.

### spec.jobs[].environmentSlug

`string`

### spec.jobs[].dockerfilePath

`string`

### spec.jobs[].buildCommand

`string`

### spec.jobs[].runCommand

`string`

### spec.jobs[].instanceSizeSlug

`string`

- default: `basic-xxs`

### spec.jobs[].instanceCount

`uint32`

- default: `1`

### spec.jobs[].kind

`enum`

Allowed values (use exactly as shown):

- `digital_ocean_app_job_kind_unspecified`
- `pre_deploy`
- `post_deploy`
- `failed_deploy`

### spec.jobs[].termination

`DigitalOceanAppTermination`

### spec.jobs[].termination.gracePeriodSeconds

`uint32` · optional (explicit presence)

Seconds to wait for in-flight work after SIGTERM. Range 1-600. Server
default is 120 when omitted.

- rule: {"uint32":{"lte":600,"gte":1}}

### spec.jobs[].termination.drainSeconds

`uint32` · optional (explicit presence)

Seconds to drain HTTP connections on a service before SIGTERM. Range
1-110. Server default is 15. Ignored (and rejected by CEL on workers/jobs)
unless the component is a service.

- rule: {"uint32":{"lte":110,"gte":1}}

### spec.jobs[].envs

`[]DigitalOceanAppEnvVar`

- rule: set either plaintext or secret for this environment variable - App Platform needs a value

### spec.jobs[].envs[].key

`string` · required

Variable name, for example DATABASE_URL or NODE_ENV.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.jobs[].envs[].plaintext

`string`

Non-secret value. Visible in the App Platform UI and in build logs.

### spec.jobs[].envs[].secret

`string` · sensitive

Secret value (API keys, database URLs, tokens). Stored in App Platform's
secret store; the IaC modules send type=SECRET.

### spec.jobs[].envs[].scope

`enum`

When the variable is injected. Omit to use run_and_build_time.

Allowed values (use exactly as shown):

- `digital_ocean_app_env_scope_unspecified`
- `run_and_build_time` -- Injected during the build and at runtime (provider default).
- `run_time` -- Injected only at runtime.
- `build_time` -- Injected only during the build.
- `unset` -- Provider UNSET - treated as run_and_build_time by the API.

### spec.jobs[].alerts

`[]DigitalOceanAppComponentAlert`

### spec.jobs[].alerts[].rule

`enum` · required

- rule: {"required":true}

Allowed values (use exactly as shown):

- `digital_ocean_app_component_alert_rule_unspecified`
- `cpu_utilization`
- `mem_utilization`
- `restart_count`

### spec.jobs[].alerts[].operator

`enum` · required

- rule: {"required":true}

Allowed values (use exactly as shown):

- `digital_ocean_app_alert_operator_unspecified`
- `greater_than`
- `less_than`

### spec.jobs[].alerts[].window

`enum` · required

- rule: {"required":true}

Allowed values (use exactly as shown):

- `digital_ocean_app_alert_window_unspecified`
- `five_minutes`
- `ten_minutes`
- `thirty_minutes`
- `one_hour`

### spec.jobs[].alerts[].value

`double`

Threshold. For cpu_utilization / mem_utilization this is a percent; for
restart_count it is a count.

- rule: {"double":{"gte":0}}

### spec.jobs[].alerts[].disabled

`bool`

### spec.jobs[].alerts[].destinations

`DigitalOceanAppAlertDestinations`

### spec.jobs[].alerts[].destinations.emails

`[]string`

### spec.jobs[].alerts[].destinations.slackWebhooks

`[]DigitalOceanAppSlackWebhook`

### spec.jobs[].alerts[].destinations.slackWebhooks[].channel

`string` · required

- rule: {"required":true}

### spec.jobs[].alerts[].destinations.slackWebhooks[].url

`string` · required · sensitive

- rule: {"required":true}

### spec.jobs[].logDestinations

`[]DigitalOceanAppLogDestination`

- rule: set exactly one sink: papertrail, datadog, logtail, or open_search

### spec.jobs[].logDestinations[].name

`string` · required

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.jobs[].logDestinations[].papertrail

`DigitalOceanAppPapertrailLog`

### spec.jobs[].logDestinations[].papertrail.endpoint

`string` · required

Syslog endpoint, for example logs.papertrailapp.com:12345

- rule: {"required":true}

### spec.jobs[].logDestinations[].datadog

`DigitalOceanAppDatadogLog`

### spec.jobs[].logDestinations[].datadog.apiKey

`string` · required · sensitive

- rule: {"required":true}

### spec.jobs[].logDestinations[].datadog.endpoint

`string`

Defaults to https://http-intake.logs.datadoghq.com when omitted.

### spec.jobs[].logDestinations[].logtail

`DigitalOceanAppLogtailLog`

### spec.jobs[].logDestinations[].logtail.token

`string` · required · sensitive

- rule: {"required":true}

### spec.jobs[].logDestinations[].openSearch

`DigitalOceanAppOpenSearchLog`

### spec.jobs[].logDestinations[].openSearch.endpoint

`string`

### spec.jobs[].logDestinations[].openSearch.indexName

`string`

### spec.jobs[].logDestinations[].openSearch.clusterName

`string`

### spec.jobs[].logDestinations[].openSearch.basicAuth

`DigitalOceanAppOpenSearchBasicAuth`

The provider requires this block even when user and password are empty
(App Platform's OpenSearch integration uses it as a placeholder).

### spec.jobs[].logDestinations[].openSearch.basicAuth.user

`string`

### spec.jobs[].logDestinations[].openSearch.basicAuth.password

`string` · sensitive

### spec.staticSites

`[]DigitalOceanAppStaticSite`

- rule: set exactly one source for this static site: git, github, gitlab, or bitbucket

### spec.staticSites[].name

`string` · required

Component name: the API's rule is ^[a-z][a-z0-9-]{0,30}[a-z0-9]$ (2-32
chars, starts with a letter) and a violation rejects the whole app spec.

- rule: {"required":true,"string":{"minLen":"2","maxLen":"32","pattern":"^[a-z][a-z0-9-]{0,30}[a-z0-9]$"}}

### spec.staticSites[].sourceDir

`string`

### spec.staticSites[].git

`DigitalOceanAppGitSource`

### spec.staticSites[].git.repoCloneUrl

`string` · required

HTTPS or git clone URL, for example https://github.com/example/app.git

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.staticSites[].git.branch

`string` · required

Branch to deploy. Example: main

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.staticSites[].github

`DigitalOceanAppGithubSource`

### spec.staticSites[].github.repo

`string` · required

Repository in owner/repo form, for example plantonhq/demo

- rule: {"required":true,"string":{"pattern":"^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$"}}

### spec.staticSites[].github.branch

`string` · required

Branch to deploy. Example: main

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.staticSites[].github.deployOnPush

`bool`

Redeploy automatically when this branch is pushed.

### spec.staticSites[].gitlab

`DigitalOceanAppGitlabSource`

### spec.staticSites[].gitlab.repo

`string` · required

- rule: {"required":true,"string":{"pattern":"^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$"}}

### spec.staticSites[].gitlab.branch

`string` · required

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.staticSites[].gitlab.deployOnPush

`bool`

### spec.staticSites[].bitbucket

`DigitalOceanAppBitbucketSource`

### spec.staticSites[].bitbucket.repo

`string` · required

- rule: {"required":true,"string":{"pattern":"^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$"}}

### spec.staticSites[].bitbucket.branch

`string` · required

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.staticSites[].bitbucket.deployOnPush

`bool`

### spec.staticSites[].environmentSlug

`string`

### spec.staticSites[].dockerfilePath

`string`

### spec.staticSites[].buildCommand

`string`

### spec.staticSites[].outputDir

`string`

Directory of built assets. App Platform auto-scans _static, dist, and
public when omitted.

### spec.staticSites[].indexDocument

`string`

### spec.staticSites[].errorDocument

`string`

### spec.staticSites[].catchallDocument

`string`

### spec.staticSites[].envs

`[]DigitalOceanAppEnvVar`

- rule: set either plaintext or secret for this environment variable - App Platform needs a value

### spec.staticSites[].envs[].key

`string` · required

Variable name, for example DATABASE_URL or NODE_ENV.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.staticSites[].envs[].plaintext

`string`

Non-secret value. Visible in the App Platform UI and in build logs.

### spec.staticSites[].envs[].secret

`string` · sensitive

Secret value (API keys, database URLs, tokens). Stored in App Platform's
secret store; the IaC modules send type=SECRET.

### spec.staticSites[].envs[].scope

`enum`

When the variable is injected. Omit to use run_and_build_time.

Allowed values (use exactly as shown):

- `digital_ocean_app_env_scope_unspecified`
- `run_and_build_time` -- Injected during the build and at runtime (provider default).
- `run_time` -- Injected only at runtime.
- `build_time` -- Injected only during the build.
- `unset` -- Provider UNSET - treated as run_and_build_time by the API.

### spec.functions

`[]DigitalOceanAppFunctionComponent`

- rule: set exactly one source for this functions component: git, github, gitlab, or bitbucket

### spec.functions[].name

`string` · required

Component name: the API's rule is ^[a-z][a-z0-9-]{0,30}[a-z0-9]$ (2-32
chars, starts with a letter) and a violation rejects the whole app spec.

- rule: {"required":true,"string":{"minLen":"2","maxLen":"32","pattern":"^[a-z][a-z0-9-]{0,30}[a-z0-9]$"}}

### spec.functions[].sourceDir

`string`

Directory inside the repo that contains project.yml and the packages tree.

### spec.functions[].git

`DigitalOceanAppGitSource`

### spec.functions[].git.repoCloneUrl

`string` · required

HTTPS or git clone URL, for example https://github.com/example/app.git

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.functions[].git.branch

`string` · required

Branch to deploy. Example: main

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.functions[].github

`DigitalOceanAppGithubSource`

### spec.functions[].github.repo

`string` · required

Repository in owner/repo form, for example plantonhq/demo

- rule: {"required":true,"string":{"pattern":"^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$"}}

### spec.functions[].github.branch

`string` · required

Branch to deploy. Example: main

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.functions[].github.deployOnPush

`bool`

Redeploy automatically when this branch is pushed.

### spec.functions[].gitlab

`DigitalOceanAppGitlabSource`

### spec.functions[].gitlab.repo

`string` · required

- rule: {"required":true,"string":{"pattern":"^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$"}}

### spec.functions[].gitlab.branch

`string` · required

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.functions[].gitlab.deployOnPush

`bool`

### spec.functions[].bitbucket

`DigitalOceanAppBitbucketSource`

### spec.functions[].bitbucket.repo

`string` · required

- rule: {"required":true,"string":{"pattern":"^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$"}}

### spec.functions[].bitbucket.branch

`string` · required

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.functions[].bitbucket.deployOnPush

`bool`

### spec.functions[].envs

`[]DigitalOceanAppEnvVar`

- rule: set either plaintext or secret for this environment variable - App Platform needs a value

### spec.functions[].envs[].key

`string` · required

Variable name, for example DATABASE_URL or NODE_ENV.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.functions[].envs[].plaintext

`string`

Non-secret value. Visible in the App Platform UI and in build logs.

### spec.functions[].envs[].secret

`string` · sensitive

Secret value (API keys, database URLs, tokens). Stored in App Platform's
secret store; the IaC modules send type=SECRET.

### spec.functions[].envs[].scope

`enum`

When the variable is injected. Omit to use run_and_build_time.

Allowed values (use exactly as shown):

- `digital_ocean_app_env_scope_unspecified`
- `run_and_build_time` -- Injected during the build and at runtime (provider default).
- `run_time` -- Injected only at runtime.
- `build_time` -- Injected only during the build.
- `unset` -- Provider UNSET - treated as run_and_build_time by the API.

### spec.functions[].alerts

`[]DigitalOceanAppComponentAlert`

### spec.functions[].alerts[].rule

`enum` · required

- rule: {"required":true}

Allowed values (use exactly as shown):

- `digital_ocean_app_component_alert_rule_unspecified`
- `cpu_utilization`
- `mem_utilization`
- `restart_count`

### spec.functions[].alerts[].operator

`enum` · required

- rule: {"required":true}

Allowed values (use exactly as shown):

- `digital_ocean_app_alert_operator_unspecified`
- `greater_than`
- `less_than`

### spec.functions[].alerts[].window

`enum` · required

- rule: {"required":true}

Allowed values (use exactly as shown):

- `digital_ocean_app_alert_window_unspecified`
- `five_minutes`
- `ten_minutes`
- `thirty_minutes`
- `one_hour`

### spec.functions[].alerts[].value

`double`

Threshold. For cpu_utilization / mem_utilization this is a percent; for
restart_count it is a count.

- rule: {"double":{"gte":0}}

### spec.functions[].alerts[].disabled

`bool`

### spec.functions[].alerts[].destinations

`DigitalOceanAppAlertDestinations`

### spec.functions[].alerts[].destinations.emails

`[]string`

### spec.functions[].alerts[].destinations.slackWebhooks

`[]DigitalOceanAppSlackWebhook`

### spec.functions[].alerts[].destinations.slackWebhooks[].channel

`string` · required

- rule: {"required":true}

### spec.functions[].alerts[].destinations.slackWebhooks[].url

`string` · required · sensitive

- rule: {"required":true}

### spec.functions[].logDestinations

`[]DigitalOceanAppLogDestination`

- rule: set exactly one sink: papertrail, datadog, logtail, or open_search

### spec.functions[].logDestinations[].name

`string` · required

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.functions[].logDestinations[].papertrail

`DigitalOceanAppPapertrailLog`

### spec.functions[].logDestinations[].papertrail.endpoint

`string` · required

Syslog endpoint, for example logs.papertrailapp.com:12345

- rule: {"required":true}

### spec.functions[].logDestinations[].datadog

`DigitalOceanAppDatadogLog`

### spec.functions[].logDestinations[].datadog.apiKey

`string` · required · sensitive

- rule: {"required":true}

### spec.functions[].logDestinations[].datadog.endpoint

`string`

Defaults to https://http-intake.logs.datadoghq.com when omitted.

### spec.functions[].logDestinations[].logtail

`DigitalOceanAppLogtailLog`

### spec.functions[].logDestinations[].logtail.token

`string` · required · sensitive

- rule: {"required":true}

### spec.functions[].logDestinations[].openSearch

`DigitalOceanAppOpenSearchLog`

### spec.functions[].logDestinations[].openSearch.endpoint

`string`

### spec.functions[].logDestinations[].openSearch.indexName

`string`

### spec.functions[].logDestinations[].openSearch.clusterName

`string`

### spec.functions[].logDestinations[].openSearch.basicAuth

`DigitalOceanAppOpenSearchBasicAuth`

The provider requires this block even when user and password are empty
(App Platform's OpenSearch integration uses it as a placeholder).

### spec.functions[].logDestinations[].openSearch.basicAuth.user

`string`

### spec.functions[].logDestinations[].openSearch.basicAuth.password

`string` · sensitive

### spec.databases

`[]DigitalOceanAppDatabase`

### spec.databases[].name

`string`

### spec.databases[].engine

`enum`

Allowed values (use exactly as shown):

- `digital_ocean_app_database_engine_unspecified`
- `mysql`
- `pg`
- `redis`
- `mongodb`
- `kafka`
- `opensearch`
- `valkey`

### spec.databases[].version

`string`

### spec.databases[].production

`bool`

When true, this attachment points at an existing managed cluster
(cluster_name required by the API). When false, App Platform creates a
dev database.

### spec.databases[].clusterName

`string | valueFrom`

Existing cluster name. Required by the API when production is true.
Literal cluster name, or a reference to a DigitalOceanDatabaseCluster.

Containment-exempt: an app ATTACHES an existing managed cluster as a
dependency it connects to; the app is not created inside the cluster.
On a diagram the app stays where its VPC places it, with a line to the
cluster it attaches.

- references: DigitalOceanDatabaseCluster (`spec.cluster_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: DigitalOceanDatabaseCluster, name: <that resource's name>, fieldPath: spec.cluster_name}} -- a bare string does not parse

### spec.databases[].dbName

`string`

### spec.databases[].dbUser

`string`

### spec.domains

`[]DigitalOceanAppDomain`

### spec.domains[].name

`string` · required

Hostname, for example www.example.com

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.domains[].wildcard

`bool`

### spec.domains[].zone

`string | valueFrom`

DigitalOcean-managed DNS zone that should receive the records. Omit when
the domain is managed elsewhere and you will create records yourself.

- references: DigitalOceanDnsZone (`status.outputs.zone_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: DigitalOceanDnsZone, name: <that resource's name>, fieldPath: status.outputs.zone_name}} -- a bare string does not parse

### spec.domains[].type

`string`

(Optional) The role of the domain, as the provider's own tokens.
When unset, App Platform treats the domain as DEFAULT.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["DEFAULT","PRIMARY","ALIAS"]}}

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

`[]DigitalOceanAppAlert`

### spec.alerts[].rule

`enum` · required

- rule: {"required":true}

Allowed values (use exactly as shown):

- `digital_ocean_app_alert_rule_unspecified`
- `deployment_failed`
- `deployment_live`
- `deployment_started`
- `deployment_canceled`
- `domain_failed`
- `domain_live`
- `autoscale_failed`
- `autoscale_succeeded`

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

### spec.ingress

`DigitalOceanAppIngress`

### spec.ingress.rules

`[]DigitalOceanAppIngressRule`

### spec.ingress.rules[].match

`DigitalOceanAppIngressMatch`

### spec.ingress.rules[].match.pathPrefix

`string`

Path prefix, for example /api

### spec.ingress.rules[].match.authorityExact

`string`

Exact Host header to match. Both provisioners deploy it.

### spec.ingress.rules[].component

`DigitalOceanAppIngressComponent`

### spec.ingress.rules[].component.name

`string` · required

Component name to send matching traffic to.

- rule: {"required":true}

### spec.ingress.rules[].component.preservePathPrefix

`bool`

### spec.ingress.rules[].component.rewrite

`string`

Rewrite the matched path before forwarding.

### spec.ingress.rules[].redirect

`DigitalOceanAppIngressRedirect`

### spec.ingress.rules[].redirect.uri

`string`

### spec.ingress.rules[].redirect.authority

`string`

### spec.ingress.rules[].redirect.port

`uint32` · optional (explicit presence)

### spec.ingress.rules[].redirect.scheme

`string`

### spec.ingress.rules[].redirect.redirectCode

`uint32` · optional (explicit presence)

HTTP status. Provider default is 302.

- default: `302`

### spec.ingress.rules[].cors

`DigitalOceanAppCors`

### spec.ingress.rules[].cors.allowOrigins

`DigitalOceanAppCorsAllowOrigins`

### spec.ingress.rules[].cors.allowOrigins.exact

`string`

### spec.ingress.rules[].cors.allowOrigins.regex

`string`

regex is an RE2 pattern. prefix-based matching is deprecated in the
provider and is not modeled.

### spec.ingress.rules[].cors.allowMethods

`[]string`

### spec.ingress.rules[].cors.allowHeaders

`[]string`

### spec.ingress.rules[].cors.exposeHeaders

`[]string`

### spec.ingress.rules[].cors.maxAge

`string`

Duration string, for example 5h30m

### spec.ingress.rules[].cors.allowCredentials

`bool`

### spec.ingress.secureHeader

`DigitalOceanAppSecureHeader`

One response header added to every route of this ingress. The provider
schema caps this at one header. Both provisioners deploy it.

### spec.ingress.secureHeader.key

`string` · required

- rule: {"required":true}

### spec.ingress.secureHeader.value

`string` · required

- rule: {"required":true}

### spec.egress

`enum`

Allowed values (use exactly as shown):

- `digital_ocean_app_egress_type_unspecified`
- `autoassign`
- `dedicated_ip`

### spec.maintenance

`DigitalOceanAppMaintenance`

### spec.maintenance.enabled

`bool`

### spec.maintenance.archive

`bool`

When true, the app is archived (and enabled is implied by the API).

### spec.maintenance.offlinePageUrl

`string`

### spec.vpc

`string | valueFrom`

VPC the app's egress is placed in. Optional; both provisioners deploy it.

- references: DigitalOceanVpc (`status.outputs.vpc_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: DigitalOceanVpc, name: <that resource's name>, fieldPath: status.outputs.vpc_id}} -- a bare string does not parse

### spec.features

`[]string`

Feature flags App Platform accepts as free-form strings.

### spec.disableEdgeCache

`bool`

### spec.disableEmailObfuscation

`bool`

### spec.enhancedThreatControlEnabled

`bool`

### spec.projectId

`string | valueFrom`

(Optional) The project the app is created in. Reference a
DigitalOceanProject resource (the default wiring resolves its
project_id output) or pass a literal project UUID. When unset, the app
lands in the account's default project. Create-only: the provider marks
project_id ForceNew, so changing it destroys and recreates the app (new
UUID, new default hostname).

- references: DigitalOceanProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: DigitalOceanProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

## Validation Rules

- `app_has_a_component`: add at least one component: a service, worker, job, static site, or function - an empty app cannot deploy

## Outputs

Reference an output from another manifest as `valueFrom: {kind: DigitalOceanApp, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.app_id` | `string` | App UUID. Used to import the digitalocean_app resource. |
| `status.outputs.default_hostname` | `string` | Default ondigitalocean.app hostname assigned by the platform. |
| `status.outputs.live_url` | `string` | Public URL, including https://. This is a custom domain when one is configured as PRIMARY, otherwise the default hostname. |
| `status.outputs.live_domain` | `string` | Live domain hostname without scheme. |
| `status.outputs.active_deployment_id` | `string` | UUID of the currently live deployment. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.databases[].clusterName` | DigitalOceanDatabaseCluster | `spec.cluster_name` |
| `spec.domains[].zone` | DigitalOceanDnsZone | `status.outputs.zone_name` |
| `spec.vpc` | DigitalOceanVpc | `status.outputs.vpc_id` |
| `spec.projectId` | DigitalOceanProject | `status.outputs.project_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| DigitalOceanDatabaseFirewall | `spec.appIds` | `status.outputs.app_id` |

## See Also

- [Overview](../README.md)

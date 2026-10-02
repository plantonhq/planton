# GcpCloudBuildTrigger Guide

The judgment this guide protects: a trigger is two decisions -- what event fires it and where its build comes from -- and the right pairing depends on whether the event carries a commit.

## Picking the event

Repository events carry a commit, so the trigger builds exactly what was pushed. Use `repositoryEventConfig` for any repository linked through a Cloud Build connection (`GcpCloudBuildRepository`): it is Google's current path for GitHub, GitHub Enterprise, GitLab, and Bitbucket. `github` and `bitbucketServerTriggerConfig` are the first-generation integrations (the Cloud Build GitHub App and Bitbucket Server configs), `developerConnectEventConfig` reads a Developer Connect repository link, and `triggerTemplate` watches a Cloud Source Repository. Each takes exactly one filter: a `push` filter on `branch` or `tag` (RE2 expressions, such as `^main$`), or a `pullRequest` filter on the base branch with `commentControl` deciding whether a build waits for a `/gcbrun` comment. `ignoredFiles` and `includedFiles` skip builds whose changes do not matter, like docs-only commits.

Pub/Sub, webhook, and manual triggers carry no commit. Say what to check out with `sourceToBuild` (a linked `repository` or a `uri`, plus a full `ref` such as `refs/heads/main`); without it, an inline build starts from an empty workspace, which is fine for a build that only runs commands. A Pub/Sub trigger subscribes to `pubsubConfig.topic` (Artifact Registry, Cloud Storage, and Cloud Scheduler all publish there), and `filter` keeps only the messages you want, using the substitutions the trigger binds from the payload.

## Picking the build

`filename` reads a `cloudbuild.yaml` from the triggering repository, which keeps the build next to the code; it only works when the event carries a repository. Pub/Sub, webhook, and manual triggers use `gitFileSource` instead, which names the repository and revision to read the file from. `build` declares the whole build inline -- steps, images, options -- which suits a build owned by the platform rather than the repository. The field names follow Google's Build API and `cloudbuild.yaml`: `steps`, `env`, `secretEnv`, `waitFor`, `availableSecrets.secretManager`. Two options are not offered because Google fixes them for every triggered build: dynamic substitutions are always on, and the substitution option is always `ALLOW_LOOSE`. Keep the sum of step timeouts at or below the build's `timeout` (default 600s); the provider refuses a plan that breaks that rule.

## Identity, logs, and secrets

`serviceAccount` is the identity builds run as. A dedicated account (`GcpServiceAccount`, its `name` output) with only what the build needs is the safer choice over the project's default Cloud Build account. Grant it `roles/logging.logWriter`, and give the deploying identity `roles/iam.serviceAccountUser` on it. A build running as a user-specified account must say where its logs go: `build.logsBucket` (a `GcpGcsBucket` `url`), or `build.options.logging` set to `CLOUD_LOGGING_ONLY` or `NONE`. Build secrets come from Secret Manager through `availableSecrets.secretManager`, with each variable listed in the `secretEnv` of the steps that read it; the build's account needs `roles/secretmanager.secretAccessor` on those secrets. A webhook key is read by Google's Cloud Build service agent (`service-{PROJECT_NUMBER}@gcp-sa-cloudbuild.iam.gserviceaccount.com`), so grant that agent the accessor role through the secret's `iamMembers`.

## Regions and private pools

`location` defaults to `global`. A trigger on a repository linked through a regional connection lives in that region, and so does a trigger whose build runs on a private pool. An inline build names its pool in `build.options.workerPool`, a `GcpCloudBuildWorkerPool` `name` output. That is the field the provider offers, and Google's API marks it deprecated in favor of `options.pool.name`. Builds that come from `filename` or `gitFileSource` set `options.pool.name` in their `cloudbuild.yaml`.

## Lifecycle

`location` is fixed at creation. Everything else updates in place, including `triggerName`; the generated `trigger_id` never changes. `approvalConfig` makes each build wait for someone with the Cloud Build Approver role. Destroy deletes the trigger only: builds it already started run to completion and stay in the history. In a chart, destroy the trigger before the topic, secret, repository, or pool it names.

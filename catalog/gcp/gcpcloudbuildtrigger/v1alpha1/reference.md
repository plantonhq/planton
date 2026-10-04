# GcpCloudBuildTrigger

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpCloudBuildTriggerSpec declares a Cloud Build trigger
(`google_cloudbuild_trigger`): the rule that starts a build when an
event arrives, and the build it starts.

A trigger answers two questions.

When does it fire? Set at most one event source (none makes a manual
trigger, run from the console, `gcloud builds triggers run`, or the
API):

  - repository_event_config -- pushes or pull requests on a repository
    linked through a Cloud Build connection (GcpCloudBuildRepository);
    the current way to build from GitHub, GitLab, and Bitbucket;
  - github -- pushes or pull requests through the first-generation
    Cloud Build GitHub App (or a GitHub Enterprise config);
  - bitbucket_server_trigger_config -- pushes or pull requests through a
    first-generation Bitbucket Server config;
  - developer_connect_event_config -- pushes or pull requests on a
    Developer Connect git repository link;
  - trigger_template -- pushes to a Cloud Source Repository;
  - pubsub_config -- a message on a Pub/Sub topic (Artifact Registry,
    Cloud Storage, and Cloud Scheduler all publish there);
  - webhook_config -- an HTTP POST to the trigger's webhook URL.

Google states github and trigger_template are mutually exclusive; the
spec refuses that pair. Pub/Sub, webhook, and manual triggers have no
commit of their own: source_to_build names the repository and ref they
check out.

What does it build? Exactly one build configuration:

  - build -- the build declared inline (steps, images, options);
  - filename -- a cloudbuild.yaml path in the triggering repository
    (for repository, GitHub, Bitbucket Server, and Cloud Source
    Repository events);
  - git_file_source -- a cloudbuild.yaml path in a named repository
    (what Pub/Sub, webhook, and manual triggers use).

Important behavioral notes:

  - location (default "global") is a create-time decision; everything
    else, including trigger_name, updates in place.
  - A trigger whose builds use a regional repository or a private pool
    must be in that region.
  - Builds run as service_account; leave it empty for the project's
    default Cloud Build service account. A user-specified account needs
    the build's log destination set (build.logs_bucket, or
    build.options.logging CLOUD_LOGGING_ONLY or NONE).
  - Destroy deletes the trigger; builds it already started keep running
    and their history stays.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudBuildTrigger
metadata:
  name: on-release
spec:
  projectId:
    value: my-gcp-project
  triggerName: on-release
  description: Run a build for every message on the releases topic
  tags:
    - release
  serviceAccount:
    value: projects/my-gcp-project/serviceAccounts/builder@my-gcp-project.iam.gserviceaccount.com
  pubsubConfig:
    topic:
      value: projects/my-gcp-project/topics/releases
  substitutions:
    _CHANNEL: stable
  build:
    steps:
      - name: ubuntu
        args:
          - echo
          - releasing to $_CHANNEL
    timeout: 300s
    options:
      logging: CLOUD_LOGGING_ONLY
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.location` | `string` |  |  |  |
| `spec.triggerName` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.disabled` | `bool` |  |  |  |
| `spec.tags` | `[]string` |  |  |  |
| `spec.substitutions` | `map<string, string>` |  |  |  |
| `spec.serviceAccount` | `string \| valueFrom` |  |  | GcpServiceAccount (`status.outputs.name`) |
| `spec.approvalConfig` | `GcpCloudBuildTriggerApprovalConfig` |  |  |  |
| `spec.approvalConfig.approvalRequired` | `bool` |  |  |  |
| `spec.repositoryEventConfig` | `GcpCloudBuildTriggerRepositoryEventConfig` |  |  |  |
| `spec.repositoryEventConfig.repository` | `string \| valueFrom` |  |  | GcpCloudBuildRepository (`status.outputs.name`) |
| `spec.repositoryEventConfig.pullRequest` | `GcpCloudBuildTriggerPullRequestFilter` |  |  |  |
| `spec.repositoryEventConfig.pullRequest.branch` | `string` |  |  |  |
| `spec.repositoryEventConfig.pullRequest.commentControl` | `string` |  |  |  |
| `spec.repositoryEventConfig.pullRequest.invertRegex` | `bool` |  |  |  |
| `spec.repositoryEventConfig.push` | `GcpCloudBuildTriggerPushFilter` |  |  |  |
| `spec.repositoryEventConfig.push.branch` | `string` |  |  |  |
| `spec.repositoryEventConfig.push.tag` | `string` |  |  |  |
| `spec.repositoryEventConfig.push.invertRegex` | `bool` |  |  |  |
| `spec.github` | `GcpCloudBuildTriggerGithub` |  |  |  |
| `spec.github.owner` | `string` |  |  |  |
| `spec.github.name` | `string` |  |  |  |
| `spec.github.enterpriseConfigResourceName` | `string` |  |  |  |
| `spec.github.pullRequest` | `GcpCloudBuildTriggerPullRequestFilter` |  |  |  |
| `spec.github.pullRequest.branch` | `string` |  |  |  |
| `spec.github.pullRequest.commentControl` | `string` |  |  |  |
| `spec.github.pullRequest.invertRegex` | `bool` |  |  |  |
| `spec.github.push` | `GcpCloudBuildTriggerPushFilter` |  |  |  |
| `spec.github.push.branch` | `string` |  |  |  |
| `spec.github.push.tag` | `string` |  |  |  |
| `spec.github.push.invertRegex` | `bool` |  |  |  |
| `spec.bitbucketServerTriggerConfig` | `GcpCloudBuildTriggerBitbucketServerTriggerConfig` |  |  |  |
| `spec.bitbucketServerTriggerConfig.bitbucketServerConfigResource` | `string` | yes |  |  |
| `spec.bitbucketServerTriggerConfig.projectKey` | `string` | yes |  |  |
| `spec.bitbucketServerTriggerConfig.repoSlug` | `string` | yes |  |  |
| `spec.bitbucketServerTriggerConfig.pullRequest` | `GcpCloudBuildTriggerPullRequestFilter` |  |  |  |
| `spec.bitbucketServerTriggerConfig.pullRequest.branch` | `string` |  |  |  |
| `spec.bitbucketServerTriggerConfig.pullRequest.commentControl` | `string` |  |  |  |
| `spec.bitbucketServerTriggerConfig.pullRequest.invertRegex` | `bool` |  |  |  |
| `spec.bitbucketServerTriggerConfig.push` | `GcpCloudBuildTriggerPushFilter` |  |  |  |
| `spec.bitbucketServerTriggerConfig.push.branch` | `string` |  |  |  |
| `spec.bitbucketServerTriggerConfig.push.tag` | `string` |  |  |  |
| `spec.bitbucketServerTriggerConfig.push.invertRegex` | `bool` |  |  |  |
| `spec.developerConnectEventConfig` | `GcpCloudBuildTriggerDeveloperConnectEventConfig` |  |  |  |
| `spec.developerConnectEventConfig.gitRepositoryLink` | `string` | yes |  |  |
| `spec.developerConnectEventConfig.pullRequest` | `GcpCloudBuildTriggerPullRequestFilter` |  |  |  |
| `spec.developerConnectEventConfig.pullRequest.branch` | `string` |  |  |  |
| `spec.developerConnectEventConfig.pullRequest.commentControl` | `string` |  |  |  |
| `spec.developerConnectEventConfig.pullRequest.invertRegex` | `bool` |  |  |  |
| `spec.developerConnectEventConfig.push` | `GcpCloudBuildTriggerPushFilter` |  |  |  |
| `spec.developerConnectEventConfig.push.branch` | `string` |  |  |  |
| `spec.developerConnectEventConfig.push.tag` | `string` |  |  |  |
| `spec.developerConnectEventConfig.push.invertRegex` | `bool` |  |  |  |
| `spec.triggerTemplate` | `GcpCloudBuildTriggerTriggerTemplate` |  |  |  |
| `spec.triggerTemplate.projectId` | `string` |  |  |  |
| `spec.triggerTemplate.repoName` | `string` |  |  |  |
| `spec.triggerTemplate.dir` | `string` |  |  |  |
| `spec.triggerTemplate.branchName` | `string` |  |  |  |
| `spec.triggerTemplate.tagName` | `string` |  |  |  |
| `spec.triggerTemplate.commitSha` | `string` |  |  |  |
| `spec.triggerTemplate.invertRegex` | `bool` |  |  |  |
| `spec.pubsubConfig` | `GcpCloudBuildTriggerPubsubConfig` |  |  |  |
| `spec.pubsubConfig.topic` | `string \| valueFrom` | yes |  | GcpPubSubTopic (`status.outputs.topic_id`) |
| `spec.pubsubConfig.serviceAccountEmail` | `string \| valueFrom` |  |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.webhookConfig` | `GcpCloudBuildTriggerWebhookConfig` |  |  |  |
| `spec.webhookConfig.secret` | `string \| valueFrom` | yes |  | GcpSecretManagerSecret (`status.outputs.latest_version_name`) |
| `spec.sourceToBuild` | `GcpCloudBuildTriggerSourceToBuild` |  |  |  |
| `spec.sourceToBuild.repository` | `string \| valueFrom` |  |  | GcpCloudBuildRepository (`status.outputs.name`) |
| `spec.sourceToBuild.uri` | `string` |  |  |  |
| `spec.sourceToBuild.ref` | `string` | yes |  |  |
| `spec.sourceToBuild.repoType` | `string` | yes |  |  |
| `spec.sourceToBuild.githubEnterpriseConfig` | `string` |  |  |  |
| `spec.sourceToBuild.bitbucketServerConfig` | `string` |  |  |  |
| `spec.filter` | `string` |  |  |  |
| `spec.ignoredFiles` | `[]string` |  |  |  |
| `spec.includedFiles` | `[]string` |  |  |  |
| `spec.includeBuildLogs` | `string` |  |  |  |
| `spec.build` | `GcpCloudBuildTriggerBuild` |  |  |  |
| `spec.build.steps` | `[]GcpCloudBuildTriggerBuildStep` | yes |  |  |
| `spec.build.steps[].name` | `string` | yes |  |  |
| `spec.build.steps[].id` | `string` |  |  |  |
| `spec.build.steps[].args` | `[]string` |  |  |  |
| `spec.build.steps[].entrypoint` | `string` |  |  |  |
| `spec.build.steps[].script` | `string` |  |  |  |
| `spec.build.steps[].dir` | `string` |  |  |  |
| `spec.build.steps[].env` | `[]string` |  |  |  |
| `spec.build.steps[].secretEnv` | `[]string` |  |  |  |
| `spec.build.steps[].waitFor` | `[]string` |  |  |  |
| `spec.build.steps[].timeout` | `string` |  |  |  |
| `spec.build.steps[].allowFailure` | `bool` |  |  |  |
| `spec.build.steps[].allowExitCodes` | `[]int32` |  |  |  |
| `spec.build.steps[].volumes` | `[]GcpCloudBuildTriggerBuildStepVolume` |  |  |  |
| `spec.build.steps[].volumes[].name` | `string` | yes |  |  |
| `spec.build.steps[].volumes[].path` | `string` | yes |  |  |
| `spec.build.timeout` | `string` |  |  |  |
| `spec.build.queueTtl` | `string` |  |  |  |
| `spec.build.images` | `[]string` |  |  |  |
| `spec.build.logsBucket` | `string \| valueFrom` |  |  | GcpGcsBucket (`status.outputs.url`) |
| `spec.build.substitutions` | `map<string, string>` |  |  |  |
| `spec.build.tags` | `[]string` |  |  |  |
| `spec.build.artifacts` | `GcpCloudBuildTriggerBuildArtifacts` |  |  |  |
| `spec.build.artifacts.images` | `[]string` |  |  |  |
| `spec.build.artifacts.objects` | `GcpCloudBuildTriggerBuildArtifactsObjects` |  |  |  |
| `spec.build.artifacts.objects.location` | `string` |  |  |  |
| `spec.build.artifacts.objects.paths` | `[]string` |  |  |  |
| `spec.build.artifacts.mavenArtifacts` | `[]GcpCloudBuildTriggerBuildArtifactsMavenArtifact` |  |  |  |
| `spec.build.artifacts.mavenArtifacts[].repository` | `string` |  |  |  |
| `spec.build.artifacts.mavenArtifacts[].path` | `string` |  |  |  |
| `spec.build.artifacts.mavenArtifacts[].artifactId` | `string` |  |  |  |
| `spec.build.artifacts.mavenArtifacts[].groupId` | `string` |  |  |  |
| `spec.build.artifacts.mavenArtifacts[].version` | `string` |  |  |  |
| `spec.build.artifacts.npmPackages` | `[]GcpCloudBuildTriggerBuildArtifactsNpmPackage` |  |  |  |
| `spec.build.artifacts.npmPackages[].repository` | `string` |  |  |  |
| `spec.build.artifacts.npmPackages[].packagePath` | `string` |  |  |  |
| `spec.build.artifacts.pythonPackages` | `[]GcpCloudBuildTriggerBuildArtifactsPythonPackage` |  |  |  |
| `spec.build.artifacts.pythonPackages[].repository` | `string` |  |  |  |
| `spec.build.artifacts.pythonPackages[].paths` | `[]string` |  |  |  |
| `spec.build.options` | `GcpCloudBuildTriggerBuildOptions` |  |  |  |
| `spec.build.options.machineType` | `string` |  |  |  |
| `spec.build.options.diskSizeGb` | `int64` |  |  |  |
| `spec.build.options.workerPool` | `string \| valueFrom` |  |  | GcpCloudBuildWorkerPool (`status.outputs.name`) |
| `spec.build.options.logging` | `string` |  |  |  |
| `spec.build.options.logStreamingOption` | `string` |  |  |  |
| `spec.build.options.requestedVerifyOption` | `string` |  |  |  |
| `spec.build.options.sourceProvenanceHash` | `[]string` |  |  |  |
| `spec.build.options.env` | `[]string` |  |  |  |
| `spec.build.options.secretEnv` | `[]string` |  |  |  |
| `spec.build.options.volumes` | `[]GcpCloudBuildTriggerBuildOptionsVolume` |  |  |  |
| `spec.build.options.volumes[].name` | `string` |  |  |  |
| `spec.build.options.volumes[].path` | `string` |  |  |  |
| `spec.build.source` | `GcpCloudBuildTriggerBuildSource` |  |  |  |
| `spec.build.source.repoSource` | `GcpCloudBuildTriggerBuildSourceRepoSource` |  |  |  |
| `spec.build.source.repoSource.projectId` | `string` |  |  |  |
| `spec.build.source.repoSource.repoName` | `string` | yes |  |  |
| `spec.build.source.repoSource.dir` | `string` |  |  |  |
| `spec.build.source.repoSource.branchName` | `string` |  |  |  |
| `spec.build.source.repoSource.tagName` | `string` |  |  |  |
| `spec.build.source.repoSource.commitSha` | `string` |  |  |  |
| `spec.build.source.repoSource.invertRegex` | `bool` |  |  |  |
| `spec.build.source.repoSource.substitutions` | `map<string, string>` |  |  |  |
| `spec.build.source.storageSource` | `GcpCloudBuildTriggerBuildSourceStorageSource` |  |  |  |
| `spec.build.source.storageSource.bucket` | `string` | yes |  |  |
| `spec.build.source.storageSource.object` | `string` | yes |  |  |
| `spec.build.source.storageSource.generation` | `string` |  |  |  |
| `spec.build.availableSecrets` | `GcpCloudBuildTriggerBuildAvailableSecrets` |  |  |  |
| `spec.build.availableSecrets.secretManager` | `[]GcpCloudBuildTriggerBuildSecretManagerSecret` | yes |  |  |
| `spec.build.availableSecrets.secretManager[].env` | `string` | yes |  |  |
| `spec.build.availableSecrets.secretManager[].versionName` | `string \| valueFrom` | yes |  | GcpSecretManagerSecret (`status.outputs.latest_version_name`) |
| `spec.build.secrets` | `[]GcpCloudBuildTriggerBuildSecret` |  |  |  |
| `spec.build.secrets[].kmsKeyName` | `string \| valueFrom` | yes |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.build.secrets[].secretEnv` | `map<string, string>` |  |  |  |
| `spec.filename` | `string` |  |  |  |
| `spec.gitFileSource` | `GcpCloudBuildTriggerGitFileSource` |  |  |  |
| `spec.gitFileSource.path` | `string` | yes |  |  |
| `spec.gitFileSource.repoType` | `string` | yes |  |  |
| `spec.gitFileSource.repository` | `string \| valueFrom` |  |  | GcpCloudBuildRepository (`status.outputs.name`) |
| `spec.gitFileSource.uri` | `string` |  |  |  |
| `spec.gitFileSource.revision` | `string` |  |  |  |
| `spec.gitFileSource.githubEnterpriseConfig` | `string` |  |  |  |
| `spec.gitFileSource.bitbucketServerConfig` | `string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The project the trigger lives in: a literal project ID or a GcpProject
reference. Empty means the provider's default project. Immutable.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.location

`string`

The Cloud Build location the trigger lives in: "global" (the default
when empty) or a region such as "us-central1". A trigger on a
repository linked through a regional connection, or whose builds use
a private pool, must be in that region. Immutable.

### spec.triggerName

`string`

The trigger's name, unique in the project: 1-64 letters, digits, and
dashes, starting and ending with a letter or digit (Google's rule).
Defaults to metadata.name. Renames in place.

- rule: trigger_name must be 1-64 letters, digits, or dashes, starting and ending with a letter or digit

### spec.description

`string`

A human-readable description of the trigger.

### spec.disabled

`bool`

Turn the trigger off: it never starts a build until turned back on.

### spec.tags

`[]string`

Tags on the trigger, for filtering in the console and API. (Triggers
have no labels.)

### spec.substitutions

`map<string, string>`

User substitutions every build the trigger starts receives, referenced
in the build as $_NAME or ${_NAME}. Keys start with an underscore and
use only uppercase letters, digits, and underscores (Google's rule),
e.g. "_DEPLOY_ENV".

- rule: substitution keys must match ^_[A-Z0-9_]+$ (for example _DEPLOY_ENV)

### spec.serviceAccount

`string | valueFrom`

The service account builds run as, and that Google uses for every
user-controlled operation on the trigger, as
projects/{project}/serviceAccounts/{email}: a GcpServiceAccount
reference (its name output) or a literal. Empty means the project's
default Cloud Build service account. The deploying identity needs
iam.serviceAccounts.actAs on it; the account needs
roles/logging.logWriter (and whatever the build itself touches), and
the build must set its log destination (build.logs_bucket, or
build.options.logging CLOUD_LOGGING_ONLY or NONE).

- references: GcpServiceAccount (`status.outputs.name`)
- rule: service_account must be projects/{project}/serviceAccounts/{email}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.approvalConfig

`GcpCloudBuildTriggerApprovalConfig`

Require a person with the Cloud Build Approver role to approve each
build before it runs.

### spec.approvalConfig.approvalRequired

`bool`

Builds wait in a pending state until someone with the Cloud Build
Approver role (roles/cloudbuild.builds.approver) approves them.

### spec.repositoryEventConfig

`GcpCloudBuildTriggerRepositoryEventConfig`

Fire on pushes or pull requests of a repository linked through a Cloud
Build connection (GcpCloudBuildRepository).

- rule: repository_event_config takes exactly one of pull_request or push

### spec.repositoryEventConfig.repository

`string | valueFrom`

The linked repository, as
projects/{project}/locations/{location}/connections/{connection}/repositories/{repository}:
a GcpCloudBuildRepository reference (its name output) or a literal.
The trigger must be in the repository's region.

- references: GcpCloudBuildRepository (`status.outputs.name`)
- rule: repository must be projects/{project}/locations/{location}/connections/{connection}/repositories/{repository}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpCloudBuildRepository, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.repositoryEventConfig.pullRequest

`GcpCloudBuildTriggerPullRequestFilter`

Build pull requests. Exactly one of pull_request or push.

### spec.repositoryEventConfig.pullRequest.branch

`string`

An RE2 regular expression of the pull request's BASE branch, e.g.
"^main$". Required on github and bitbucket_server_trigger_config.

### spec.repositoryEventConfig.pullRequest.commentControl

`string`

Whether a build waits for a "/gcbrun" comment:
  "COMMENTS_DISABLED" -- every pull request builds
  "COMMENTS_ENABLED"  -- builds start only on a "/gcbrun" comment
                         from an owner or collaborator
  "COMMENTS_ENABLED_FOR_EXTERNAL_CONTRIBUTORS_ONLY" -- collaborators'
                         pull requests build; others wait for the
                         comment
Empty leaves Google's default (COMMENTS_DISABLED).

- rule: comment_control must be COMMENTS_DISABLED, COMMENTS_ENABLED, or COMMENTS_ENABLED_FOR_EXTERNAL_CONTRIBUTORS_ONLY

### spec.repositoryEventConfig.pullRequest.invertRegex

`bool`

Build pull requests whose base branch does NOT match branch.

### spec.repositoryEventConfig.push

`GcpCloudBuildTriggerPushFilter`

Build pushes. Exactly one of pull_request or push.

- rule: a push filter matches exactly one of branch or tag

### spec.repositoryEventConfig.push.branch

`string`

An RE2 regular expression of the pushed branch, e.g. "^main$" or
"^release/.*". Exactly one of branch or tag.

### spec.repositoryEventConfig.push.tag

`string`

An RE2 regular expression of the pushed tag, e.g. "^v[0-9]+\\.". Exactly
one of branch or tag.

### spec.repositoryEventConfig.push.invertRegex

`bool`

Build pushes whose ref does NOT match.

### spec.github

`GcpCloudBuildTriggerGithub`

Fire on GitHub pushes or pull requests through the first-generation
Cloud Build GitHub App or a GitHub Enterprise config. Mutually
exclusive with trigger_template.

- rule: github takes exactly one of pull_request or push
- rule: github.pull_request.branch is required

### spec.github.owner

`string`

The repository's owner: the user or organization, e.g.
"googlecloudplatform" for github.com/googlecloudplatform/cloud-builders.

### spec.github.name

`string`

The repository's name, e.g. "cloud-builders".

### spec.github.enterpriseConfigResourceName

`string`

The GitHub Enterprise config the installation uses, as
projects/{project}/locations/{location}/githubEnterpriseConfigs/{id}.
Empty means github.com.

### spec.github.pullRequest

`GcpCloudBuildTriggerPullRequestFilter`

Build pull requests (branch is required). Exactly one of pull_request
or push.

### spec.github.pullRequest.branch

`string`

An RE2 regular expression of the pull request's BASE branch, e.g.
"^main$". Required on github and bitbucket_server_trigger_config.

### spec.github.pullRequest.commentControl

`string`

Whether a build waits for a "/gcbrun" comment:
  "COMMENTS_DISABLED" -- every pull request builds
  "COMMENTS_ENABLED"  -- builds start only on a "/gcbrun" comment
                         from an owner or collaborator
  "COMMENTS_ENABLED_FOR_EXTERNAL_CONTRIBUTORS_ONLY" -- collaborators'
                         pull requests build; others wait for the
                         comment
Empty leaves Google's default (COMMENTS_DISABLED).

- rule: comment_control must be COMMENTS_DISABLED, COMMENTS_ENABLED, or COMMENTS_ENABLED_FOR_EXTERNAL_CONTRIBUTORS_ONLY

### spec.github.pullRequest.invertRegex

`bool`

Build pull requests whose base branch does NOT match branch.

### spec.github.push

`GcpCloudBuildTriggerPushFilter`

Build pushes. Exactly one of pull_request or push.

- rule: a push filter matches exactly one of branch or tag

### spec.github.push.branch

`string`

An RE2 regular expression of the pushed branch, e.g. "^main$" or
"^release/.*". Exactly one of branch or tag.

### spec.github.push.tag

`string`

An RE2 regular expression of the pushed tag, e.g. "^v[0-9]+\\.". Exactly
one of branch or tag.

### spec.github.push.invertRegex

`bool`

Build pushes whose ref does NOT match.

### spec.bitbucketServerTriggerConfig

`GcpCloudBuildTriggerBitbucketServerTriggerConfig`

Fire on Bitbucket Server pushes or pull requests through a
first-generation Bitbucket Server config.

- rule: bitbucket_server_trigger_config takes exactly one of pull_request or push
- rule: bitbucket_server_trigger_config.pull_request.branch is required

### spec.bitbucketServerTriggerConfig.bitbucketServerConfigResource

`string` · required

The Bitbucket Server config the trigger uses, as
projects/{project}/locations/{location}/bitbucketServerConfigs/{id}.
Required.

- rule: {"required":true}

### spec.bitbucketServerTriggerConfig.projectKey

`string` · required

The key of the Bitbucket project the repository is in, e.g. "TEST" for
https://mybitbucket.server/projects/TEST/repos/test-repo. Required.

- rule: {"required":true}

### spec.bitbucketServerTriggerConfig.repoSlug

`string` · required

The repository's slug (its URL form), e.g. "test-repo". Required.

- rule: {"required":true}

### spec.bitbucketServerTriggerConfig.pullRequest

`GcpCloudBuildTriggerPullRequestFilter`

Build pull requests (branch is required). Exactly one of pull_request
or push.

### spec.bitbucketServerTriggerConfig.pullRequest.branch

`string`

An RE2 regular expression of the pull request's BASE branch, e.g.
"^main$". Required on github and bitbucket_server_trigger_config.

### spec.bitbucketServerTriggerConfig.pullRequest.commentControl

`string`

Whether a build waits for a "/gcbrun" comment:
  "COMMENTS_DISABLED" -- every pull request builds
  "COMMENTS_ENABLED"  -- builds start only on a "/gcbrun" comment
                         from an owner or collaborator
  "COMMENTS_ENABLED_FOR_EXTERNAL_CONTRIBUTORS_ONLY" -- collaborators'
                         pull requests build; others wait for the
                         comment
Empty leaves Google's default (COMMENTS_DISABLED).

- rule: comment_control must be COMMENTS_DISABLED, COMMENTS_ENABLED, or COMMENTS_ENABLED_FOR_EXTERNAL_CONTRIBUTORS_ONLY

### spec.bitbucketServerTriggerConfig.pullRequest.invertRegex

`bool`

Build pull requests whose base branch does NOT match branch.

### spec.bitbucketServerTriggerConfig.push

`GcpCloudBuildTriggerPushFilter`

Build pushes. Exactly one of pull_request or push.

- rule: a push filter matches exactly one of branch or tag

### spec.bitbucketServerTriggerConfig.push.branch

`string`

An RE2 regular expression of the pushed branch, e.g. "^main$" or
"^release/.*". Exactly one of branch or tag.

### spec.bitbucketServerTriggerConfig.push.tag

`string`

An RE2 regular expression of the pushed tag, e.g. "^v[0-9]+\\.". Exactly
one of branch or tag.

### spec.bitbucketServerTriggerConfig.push.invertRegex

`bool`

Build pushes whose ref does NOT match.

### spec.developerConnectEventConfig

`GcpCloudBuildTriggerDeveloperConnectEventConfig`

Fire on pushes or pull requests of a Developer Connect git repository
link.

- rule: developer_connect_event_config takes exactly one of pull_request or push

### spec.developerConnectEventConfig.gitRepositoryLink

`string` · required

The Developer Connect git repository link, as
projects/{project}/locations/{location}/connections/{connection}/gitRepositoryLinks/{link}.
Required.

- rule: {"required":true}

### spec.developerConnectEventConfig.pullRequest

`GcpCloudBuildTriggerPullRequestFilter`

Build pull requests. Exactly one of pull_request or push.

### spec.developerConnectEventConfig.pullRequest.branch

`string`

An RE2 regular expression of the pull request's BASE branch, e.g.
"^main$". Required on github and bitbucket_server_trigger_config.

### spec.developerConnectEventConfig.pullRequest.commentControl

`string`

Whether a build waits for a "/gcbrun" comment:
  "COMMENTS_DISABLED" -- every pull request builds
  "COMMENTS_ENABLED"  -- builds start only on a "/gcbrun" comment
                         from an owner or collaborator
  "COMMENTS_ENABLED_FOR_EXTERNAL_CONTRIBUTORS_ONLY" -- collaborators'
                         pull requests build; others wait for the
                         comment
Empty leaves Google's default (COMMENTS_DISABLED).

- rule: comment_control must be COMMENTS_DISABLED, COMMENTS_ENABLED, or COMMENTS_ENABLED_FOR_EXTERNAL_CONTRIBUTORS_ONLY

### spec.developerConnectEventConfig.pullRequest.invertRegex

`bool`

Build pull requests whose base branch does NOT match branch.

### spec.developerConnectEventConfig.push

`GcpCloudBuildTriggerPushFilter`

Build pushes. Exactly one of pull_request or push.

- rule: a push filter matches exactly one of branch or tag

### spec.developerConnectEventConfig.push.branch

`string`

An RE2 regular expression of the pushed branch, e.g. "^main$" or
"^release/.*". Exactly one of branch or tag.

### spec.developerConnectEventConfig.push.tag

`string`

An RE2 regular expression of the pushed tag, e.g. "^v[0-9]+\\.". Exactly
one of branch or tag.

### spec.developerConnectEventConfig.push.invertRegex

`bool`

Build pushes whose ref does NOT match.

### spec.triggerTemplate

`GcpCloudBuildTriggerTriggerTemplate`

Fire on pushes to a Cloud Source Repository. Mutually exclusive with
github.

- rule: trigger_template takes exactly one of branch_name, tag_name, or commit_sha

### spec.triggerTemplate.projectId

`string`

The project that owns the repository. Empty means the trigger's
project.

### spec.triggerTemplate.repoName

`string`

The Cloud Source Repository's name. Empty means "default".

### spec.triggerTemplate.dir

`string`

The directory, relative to the repository root, builds run in.

### spec.triggerTemplate.branchName

`string`

An RE2 regular expression of the branches to build. Exactly one of
branch_name, tag_name, or commit_sha.

### spec.triggerTemplate.tagName

`string`

An RE2 regular expression of the tags to build. Exactly one of
branch_name, tag_name, or commit_sha.

### spec.triggerTemplate.commitSha

`string`

An explicit commit SHA to build. Exactly one of branch_name,
tag_name, or commit_sha.

### spec.triggerTemplate.invertRegex

`bool`

Build revisions that do NOT match the branch or tag expression.

### spec.pubsubConfig

`GcpCloudBuildTriggerPubsubConfig`

Fire on each message published to a Pub/Sub topic. Cloud Build creates
and owns the push subscription.

### spec.pubsubConfig.topic

`string | valueFrom` · required

The topic, as projects/{project}/topics/{topic}: a GcpPubSubTopic
reference (its topic_id output) or a literal. Required.

- references: GcpPubSubTopic (`status.outputs.topic_id`)
- rule: topic must be projects/{project}/topics/{topic}
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpPubSubTopic, name: <that resource's name>, fieldPath: status.outputs.topic_id}} -- a bare string does not parse

### spec.pubsubConfig.serviceAccountEmail

`string | valueFrom`

The service account email the push subscription authenticates as: a
GcpServiceAccount reference (its email output) or a literal. Empty
lets Google choose.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.webhookConfig

`GcpCloudBuildTriggerWebhookConfig`

Fire on each HTTP POST to the trigger's webhook URL, authenticated by
a secret key.

### spec.webhookConfig.secret

`string | valueFrom` · required

The Secret Manager secret version holding the webhook's key, as
projects/{project}/secrets/{secret}/versions/{version}: a
GcpSecretManagerSecret reference (its latest_version_name output) or a
literal. Callers pass the key as the URL's "secret" parameter. Google's
Cloud Build service agent
(service-{PROJECT_NUMBER}@gcp-sa-cloudbuild.iam.gserviceaccount.com)
reads it, so grant that agent roles/secretmanager.secretAccessor on
the secret (GcpSecretManagerSecret.iam_members). Required.

- references: GcpSecretManagerSecret (`status.outputs.latest_version_name`)
- rule: webhook_config.secret must be projects/{project}/secrets/{secret}/versions/{version}
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSecretManagerSecret, name: <that resource's name>, fieldPath: status.outputs.latest_version_name}} -- a bare string does not parse

### spec.sourceToBuild

`GcpCloudBuildTriggerSourceToBuild`

The repository and ref a Pub/Sub, webhook, or manual trigger checks
out (triggers that respond to repository events build the commit that
caused the event and ignore this).

- rule: source_to_build takes exactly one of repository or uri

### spec.sourceToBuild.repository

`string | valueFrom`

A repository linked through a Cloud Build connection, as
projects/{project}/locations/{location}/connections/{connection}/repositories/{repository}:
a GcpCloudBuildRepository reference (its name output) or a literal.
Exactly one of repository or uri.

- references: GcpCloudBuildRepository (`status.outputs.name`)
- rule: source_to_build.repository must be projects/{project}/locations/{location}/connections/{connection}/repositories/{repository}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpCloudBuildRepository, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.sourceToBuild.uri

`string`

The repository's URI, e.g. "https://github.com/acme/orders" (for the
first-generation GitHub App, Bitbucket Server, or a Cloud Source
Repository). Exactly one of repository or uri.

### spec.sourceToBuild.ref

`string` · required

The branch or tag to build, as a full ref: "refs/heads/main" or
"refs/tags/v1.0.0". Required.

- rule: ref must start with refs/ (for example refs/heads/main)
- rule: {"required":true}

### spec.sourceToBuild.repoType

`string` · required

The kind of repository: "GITHUB", "BITBUCKET_SERVER",
"CLOUD_SOURCE_REPOSITORIES", or "UNKNOWN" (the values the provider
accepts; a connected GitLab or Bitbucket Cloud repository is still
named through repository). Required.

- rule: repo_type must be UNKNOWN, CLOUD_SOURCE_REPOSITORIES, GITHUB, or BITBUCKET_SERVER
- rule: {"required":true}

### spec.sourceToBuild.githubEnterpriseConfig

`string`

The GitHub Enterprise config the repository is reached through, as
projects/{project}/locations/{location}/githubEnterpriseConfigs/{id}.

### spec.sourceToBuild.bitbucketServerConfig

`string`

The Bitbucket Server config the repository is reached through, as
projects/{project}/locations/{location}/bitbucketServerConfigs/{id}.

### spec.filter

`string`

A Common Expression Language filter on the incoming event; the trigger
fires only when it evaluates true. Used with Pub/Sub and webhook
triggers, over the substitutions they bind from the payload, e.g.
"_ACTION == 'INSERT'".

### spec.ignoredFiles

`[]string`

File globs (Go filepath.Match plus "**") of changes to ignore: a
commit whose changed files all match is not built. Repository-event
triggers only.

### spec.includedFiles

`[]string`

File globs of changes that matter: when set, a commit is built only if
at least one changed file (not ignored) matches. Repository-event
triggers only.

### spec.includeBuildLogs

`string`

Show the build log link on the GitHub check run when the build
finishes: "INCLUDE_BUILD_LOGS_WITH_STATUS". Google accepts it only on
GitHub triggers (github, or repository_event_config on a GitHub
repository) and refuses it with INVALID_ARGUMENT otherwise. Empty
leaves logs off GitHub.

- rule: include_build_logs must be INCLUDE_BUILD_LOGS_UNSPECIFIED or INCLUDE_BUILD_LOGS_WITH_STATUS

### spec.build

`GcpCloudBuildTriggerBuild`

The build, declared inline. Exactly one of build, filename, or
git_file_source.

### spec.build.steps

`[]GcpCloudBuildTriggerBuildStep` · required

The build's steps, run in order (or as wait_for allows). At least one.
The sum of the steps' timeouts may not exceed the build's timeout.

- rule: {"repeated":{"minItems":"1"}}
- rule: a step with script cannot also set entrypoint or args

### spec.build.steps[].name

`string` · required

The image the step runs, e.g. "gcr.io/cloud-builders/docker",
"gcr.io/google.com/cloudsdktool/cloud-sdk", or "ubuntu". Required.

- rule: {"required":true}

### spec.build.steps[].id

`string`

The step's ID, for other steps' wait_for.

### spec.build.steps[].args

`[]string`

Arguments to the image's entrypoint (or, with no entrypoint, the
command and its arguments). Not with script.

### spec.build.steps[].entrypoint

`string`

An entrypoint replacing the image's own. Not with script.

### spec.build.steps[].script

`string`

A shell script run in the step, instead of entrypoint and args.

### spec.build.steps[].dir

`string`

The working directory, relative to the build's workspace (an absolute
path is outside it and not kept between steps unless it is a volume).

### spec.build.steps[].env

`[]string`

Environment variables, each "KEY=VALUE".

### spec.build.steps[].secretEnv

`[]string`

Names of secret environment variables the step receives, each defined
in the build's available_secrets.secret_manager (env) or secrets
(secret_env keys).

### spec.build.steps[].waitFor

`[]string`

The IDs of the steps this one waits for. Empty waits for every
earlier step; ["-"] starts it at once.

### spec.build.steps[].timeout

`string`

How long the step may run, e.g. "300s". Empty means until the build
times out.

### spec.build.steps[].allowFailure

`bool`

Let the step fail without failing the build.

### spec.build.steps[].allowExitCodes

`[]int32`

Let the step fail without failing the build only with one of these
exit codes (takes precedence over allow_failure).

### spec.build.steps[].volumes

`[]GcpCloudBuildTriggerBuildStepVolume`

Volumes mounted into the step. A named volume must be used by at
least two steps.

### spec.build.steps[].volumes[].name

`string` · required

The volume's name (a valid Docker volume name, unique in the step).
Required.

- rule: {"required":true}

### spec.build.steps[].volumes[].path

`string` · required

The absolute path the volume mounts at. Required.

- rule: {"required":true}

### spec.build.timeout

`string`

How long the build may run, in seconds with an "s" suffix, e.g.
"1200s". Empty means 600s. Must be at least the sum of the steps'
timeouts.

### spec.build.queueTtl

`string`

How long the build may wait in the queue before it expires, e.g.
"3600s". Empty means no limit.

### spec.build.images

`[]string`

Images pushed after every step succeeds, e.g.
"us-docker.pkg.dev/my-project/apps/orders:$SHORT_SHA". Their digests
land in the build's results.

### spec.build.logsBucket

`string | valueFrom`

The Cloud Storage location build logs are written to, as
gs://{bucket} or gs://{bucket}/{path}: a GcpGcsBucket reference (its
url output) or a literal. Google writes {logs_bucket}/log-{build_id}.txt.
Empty means Cloud Build's default logs bucket. A build running as a
user-specified service account needs this or options.logging set.

- references: GcpGcsBucket (`status.outputs.url`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGcsBucket, name: <that resource's name>, fieldPath: status.outputs.url}} -- a bare string does not parse

### spec.build.substitutions

`map<string, string>`

Build-level substitutions, merged with the trigger's.

### spec.build.tags

`[]string`

Tags on each build, for filtering the build history. (Not Docker tags.)

### spec.build.artifacts

`GcpCloudBuildTriggerBuildArtifacts`

Non-image artifacts uploaded after every step succeeds.

### spec.build.artifacts.images

`[]string`

Images pushed after every step succeeds (the same as the build's
images).

### spec.build.artifacts.objects

`GcpCloudBuildTriggerBuildArtifactsObjects`

Workspace files uploaded to Cloud Storage.

### spec.build.artifacts.objects.location

`string`

The destination, as gs://{bucket}/{path}/.

### spec.build.artifacts.objects.paths

`[]string`

Globs of workspace files to upload.

### spec.build.artifacts.mavenArtifacts

`[]GcpCloudBuildTriggerBuildArtifactsMavenArtifact`

Maven artifacts uploaded to Artifact Registry.

### spec.build.artifacts.mavenArtifacts[].repository

`string`

The repository, as https://{region}-maven.pkg.dev/{project}/{repository}.

### spec.build.artifacts.mavenArtifacts[].path

`string`

The artifact's path in the workspace, e.g.
"my-app/target/my-app-1.0.SNAPSHOT.jar".

### spec.build.artifacts.mavenArtifacts[].artifactId

`string`

The Maven artifactId.

### spec.build.artifacts.mavenArtifacts[].groupId

`string`

The Maven groupId.

### spec.build.artifacts.mavenArtifacts[].version

`string`

The Maven version.

### spec.build.artifacts.npmPackages

`[]GcpCloudBuildTriggerBuildArtifactsNpmPackage`

npm packages uploaded to Artifact Registry.

### spec.build.artifacts.npmPackages[].repository

`string`

The repository, as https://{region}-npm.pkg.dev/{project}/{repository}.

### spec.build.artifacts.npmPackages[].packagePath

`string`

The directory holding the package's package.json.

### spec.build.artifacts.pythonPackages

`[]GcpCloudBuildTriggerBuildArtifactsPythonPackage`

Python packages uploaded to Artifact Registry.

### spec.build.artifacts.pythonPackages[].repository

`string`

The repository, as https://{region}-python.pkg.dev/{project}/{repository}.

### spec.build.artifacts.pythonPackages[].paths

`[]string`

Globs of the files to upload, usually "dist/*".

### spec.build.options

`GcpCloudBuildTriggerBuildOptions`

Machine, disk, logging, and private-pool options.

- rule: source_provenance_hash entries must be NONE, SHA256, or MD5

### spec.build.options.machineType

`string`

The machine type builds run on in Google's default pool: "E2_MEDIUM",
"E2_HIGHCPU_8", "E2_HIGHCPU_32", "N1_HIGHCPU_8", or "N1_HIGHCPU_32".
Empty means the standard machine. A private pool's machine is set on
the pool.

### spec.build.options.diskSizeGb

`int64`

The minimum disk size in GB for the build's VM (some of it is used by
the system). 0 means the standard size.

- rule: {"int64":{"gte":"0"}}

### spec.build.options.workerPool

`string | valueFrom`

A private pool the builds run on, as
projects/{project}/locations/{location}/workerPools/{pool}: a
GcpCloudBuildWorkerPool reference (its name output) or a literal. The
trigger must be in the pool's region. Google's API marks this field
(BuildOptions.workerPool) deprecated in favor of options.pool.name,
which the provider does not expose; a filename or git_file_source
build sets options.pool.name in its cloudbuild.yaml instead.

- references: GcpCloudBuildWorkerPool (`status.outputs.name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpCloudBuildWorkerPool, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.build.options.logging

`string`

Where build logs go:
  "CLOUD_LOGGING_ONLY" -- Cloud Logging only
  "GCS_ONLY"           -- the logs bucket only
  "LEGACY"             -- both (the old default)
  "STACKDRIVER_ONLY"   -- Cloud Logging only (older name)
  "NONE"               -- nowhere
  "LOGGING_UNSPECIFIED"-- Google's default
A build running as a user-specified service account needs
CLOUD_LOGGING_ONLY, NONE, or a logs_bucket.

- rule: logging must be LOGGING_UNSPECIFIED, LEGACY, GCS_ONLY, STACKDRIVER_ONLY, CLOUD_LOGGING_ONLY, or NONE

### spec.build.options.logStreamingOption

`string`

Whether logs stream to the Cloud Storage bucket while the build runs:
"STREAM_ON", "STREAM_OFF", or "STREAM_DEFAULT".

- rule: log_streaming_option must be STREAM_DEFAULT, STREAM_ON, or STREAM_OFF

### spec.build.options.requestedVerifyOption

`string`

Request verifiable build provenance: "VERIFIED" or "NOT_VERIFIED".

- rule: requested_verify_option must be NOT_VERIFIED or VERIFIED

### spec.build.options.sourceProvenanceHash

`[]string`

Hashes of the source recorded in the build's provenance: "SHA256",
"MD5", or "NONE".

### spec.build.options.env

`[]string`

Environment variables for every step, each "KEY=VALUE"; a step's own
env wins on a collision.

### spec.build.options.secretEnv

`[]string`

Names of secret environment variables every step receives, each
defined in the build's secrets.

### spec.build.options.volumes

`[]GcpCloudBuildTriggerBuildOptionsVolume`

Volumes mounted into every step (names and paths may not collide with
a step's own volumes). Not valid on a one-step build.

### spec.build.options.volumes[].name

`string`

The volume's name (a valid Docker volume name).

### spec.build.options.volumes[].path

`string`

The absolute path the volume mounts at.

### spec.build.source

`GcpCloudBuildTriggerBuildSource`

An explicit source for the build. Usually empty: a triggered build
builds the source of its trigger.

- rule: source takes exactly one of repo_source or storage_source

### spec.build.source.repoSource

`GcpCloudBuildTriggerBuildSourceRepoSource`

A Cloud Source Repository revision.

- rule: repo_source takes exactly one of branch_name, tag_name, or commit_sha

### spec.build.source.repoSource.projectId

`string`

The project that owns the repository. Empty means the build's
project.

### spec.build.source.repoSource.repoName

`string` · required

The repository's name. Required.

- rule: {"required":true}

### spec.build.source.repoSource.dir

`string`

The directory, relative to the repository root, the build runs in.

### spec.build.source.repoSource.branchName

`string`

An RE2 regular expression of the branch to build. Exactly one of
branch_name, tag_name, or commit_sha.

### spec.build.source.repoSource.tagName

`string`

An RE2 regular expression of the tag to build. Exactly one of
branch_name, tag_name, or commit_sha.

### spec.build.source.repoSource.commitSha

`string`

An explicit commit SHA. Exactly one of branch_name, tag_name, or
commit_sha.

### spec.build.source.repoSource.invertRegex

`bool`

Build revisions that do NOT match the branch or tag expression.

### spec.build.source.repoSource.substitutions

`map<string, string>`

Substitutions for a triggered build of this source.

### spec.build.source.storageSource

`GcpCloudBuildTriggerBuildSourceStorageSource`

A gzipped tarball in Cloud Storage.

### spec.build.source.storageSource.bucket

`string` · required

The bucket's name. Required.

- rule: {"required":true}

### spec.build.source.storageSource.object

`string` · required

The object's name: a .tar.gz of the source. Required.

- rule: {"required":true}

### spec.build.source.storageSource.generation

`string`

The object's generation. Empty means the latest.

### spec.build.availableSecrets

`GcpCloudBuildTriggerBuildAvailableSecrets`

Secret Manager secrets exposed to the steps as environment variables
(listed in a step's secret_env).

### spec.build.availableSecrets.secretManager

`[]GcpCloudBuildTriggerBuildSecretManagerSecret` · required

One entry per secret environment variable. At least one.

- rule: {"repeated":{"minItems":"1"}}

### spec.build.availableSecrets.secretManager[].env

`string` · required

The environment variable's name; list it in each step's secret_env
that reads it. Unique across the build's secrets. Required.

- rule: {"required":true}

### spec.build.availableSecrets.secretManager[].versionName

`string | valueFrom` · required

The secret version, as
projects/{project}/secrets/{secret}/versions/{version}: a
GcpSecretManagerSecret reference (its latest_version_name output) or a
literal ("versions/latest" follows rotation). The build's service
account needs roles/secretmanager.secretAccessor on the secret.
Required.

- references: GcpSecretManagerSecret (`status.outputs.latest_version_name`)
- rule: version_name must be projects/{project}/secrets/{secret}/versions/{version}
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSecretManagerSecret, name: <that resource's name>, fieldPath: status.outputs.latest_version_name}} -- a bare string does not parse

### spec.build.secrets

`[]GcpCloudBuildTriggerBuildSecret`

Cloud KMS-encrypted values exposed to the steps as environment
variables (listed in a step's secret_env). Prefer available_secrets.

### spec.build.secrets[].kmsKeyName

`string | valueFrom` · required

The key that decrypts the values, as
projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}:
a GcpKmsKey reference (its key_id output) or a literal. The build's
service account needs roles/cloudkms.cryptoKeyDecrypter on it.
Required.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: kms_key_name must be projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.build.secrets[].secretEnv

`map<string, string>`

Environment variable name to its base64 ciphertext, encrypted with
kms_key_name (at most 64 KB each, 100 across the build).

### spec.filename

`string`

The path, from the repository root, of the build configuration file
(e.g. "cloudbuild.yaml") in the triggering repository. For
repository_event_config, github, bitbucket_server_trigger_config,
developer_connect_event_config, and trigger_template triggers; Pub/Sub,
webhook, and manual triggers use git_file_source. Exactly one of
build, filename, or git_file_source.

### spec.gitFileSource

`GcpCloudBuildTriggerGitFileSource`

The build configuration file read from a named repository and
revision -- what Pub/Sub, webhook, and manual triggers use. Exactly
one of build, filename, or git_file_source.

- rule: git_file_source takes at most one of repository or uri

### spec.gitFileSource.path

`string` · required

The file's path from the repository root, e.g. "cloudbuild.yaml".
Required.

- rule: {"required":true}

### spec.gitFileSource.repoType

`string` · required

The kind of repository: "GITHUB", "BITBUCKET_SERVER",
"CLOUD_SOURCE_REPOSITORIES", or "UNKNOWN". Required.

- rule: repo_type must be UNKNOWN, CLOUD_SOURCE_REPOSITORIES, GITHUB, or BITBUCKET_SERVER
- rule: {"required":true}

### spec.gitFileSource.repository

`string | valueFrom`

A repository linked through a Cloud Build connection: a
GcpCloudBuildRepository reference (its name output) or a literal
projects/{project}/locations/{location}/connections/{connection}/repositories/{repository}.
At most one of repository or uri; with neither, the file is read from
the repository the event came from.

- references: GcpCloudBuildRepository (`status.outputs.name`)
- rule: git_file_source.repository must be projects/{project}/locations/{location}/connections/{connection}/repositories/{repository}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpCloudBuildRepository, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.gitFileSource.uri

`string`

The repository's URI, e.g. "https://github.com/acme/orders". At most
one of repository or uri.

### spec.gitFileSource.revision

`string`

The branch, tag, ref, or SHA to read the file at (git revision
syntax), e.g. "refs/heads/main". Empty means the revision that
triggered the build.

### spec.gitFileSource.githubEnterpriseConfig

`string`

The GitHub Enterprise config the repository is reached through, as
projects/{project}/locations/{location}/githubEnterpriseConfigs/{id}.

### spec.gitFileSource.bitbucketServerConfig

`string`

The Bitbucket Server config the repository is reached through, as
projects/{project}/locations/{location}/bitbucketServerConfigs/{id}.

### spec.deletionPolicy

`string`

What destroy does:
  "" / "DELETE" -- the trigger is deleted
  "PREVENT"     -- destroy fails
  "ABANDON"     -- the trigger leaves management and stays in Google

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `spec.exactly_one_build_config`: set exactly one of build, filename, or git_file_source
- `spec.github_or_trigger_template`: github and trigger_template are mutually exclusive

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpCloudBuildTrigger, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.id` | `string` | The trigger's full resource ID: projects/{project}/locations/{location}/triggers/{trigger_id} for a regional trigger, projects/{project}/triggers/{trigger_id} for a global one (the form Google's legacy global API addresses). |
| `status.outputs.trigger_id` | `string` | Google's generated unique ID for the trigger -- what `gcloud builds triggers run` and the triggers.run API take. |
| `status.outputs.name` | `string` | The trigger's name (spec.trigger_name, or metadata.name by default). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.serviceAccount` | GcpServiceAccount | `status.outputs.name` |
| `spec.repositoryEventConfig.repository` | GcpCloudBuildRepository | `status.outputs.name` |
| `spec.pubsubConfig.topic` | GcpPubSubTopic | `status.outputs.topic_id` |
| `spec.pubsubConfig.serviceAccountEmail` | GcpServiceAccount | `status.outputs.email` |
| `spec.webhookConfig.secret` | GcpSecretManagerSecret | `status.outputs.latest_version_name` |
| `spec.sourceToBuild.repository` | GcpCloudBuildRepository | `status.outputs.name` |
| `spec.build.logsBucket` | GcpGcsBucket | `status.outputs.url` |
| `spec.build.options.workerPool` | GcpCloudBuildWorkerPool | `status.outputs.name` |
| `spec.build.availableSecrets.secretManager[].versionName` | GcpSecretManagerSecret | `status.outputs.latest_version_name` |
| `spec.build.secrets[].kmsKeyName` | GcpKmsKey | `status.outputs.key_id` |
| `spec.gitFileSource.repository` | GcpCloudBuildRepository | `status.outputs.name` |

## See Also

- [Overview](../README.md)

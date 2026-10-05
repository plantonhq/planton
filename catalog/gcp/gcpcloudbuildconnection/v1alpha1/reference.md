# GcpCloudBuildConnection

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpCloudBuildConnectionSpec declares a Cloud Build repository connection
(`google_cloudbuildv2_connection`): Cloud Build's authorized link to a
code host -- github.com, a GitHub Enterprise server, gitlab.com or a
GitLab Enterprise server, Bitbucket Cloud, or a Bitbucket Data Center
server.

A connection is the container its repositories live in: each
GcpCloudBuildRepository is created under one connection, and triggers,
Cloud Deploy custom target types, and other Google services then build
from those repositories.

Credentials never appear in the spec. Every token, private key, and
webhook secret is a Secret Manager secret VERSION, named here as
projects/{project}/secrets/{secret}/versions/{version} -- a
GcpSecretManagerSecret reference (its latest_version_name output) or a
literal. Google's Cloud Build service agent
(service-{PROJECT_NUMBER}@gcp-sa-cloudbuild.iam.gserviceaccount.com)
reads them, so grant it roles/secretmanager.secretAccessor on each secret
(GcpSecretManagerSecret.iam_members) before the connection is created.

GitHub connections finish in a browser: Cloud Build's GitHub App must be
installed on the account or organization, and the installation's ID set
in github_config.app_installation_id. Until then the connection reports
the next step in its installation_stage and installation_action_uri
outputs.

Important behavioral notes:

  - location and connection_id are create-time decisions.
  - Set at most one code-host block; the block you choose decides how
    Cloud Build authenticates and receives webhooks.
  - GitLab and both Bitbucket blocks treat webhook_secret_secret_version
    as immutable (changing it replaces the connection); GitHub
    Enterprise updates it in place.
  - Destroy deletes the connection; its repositories must be removed
    first (a chart orders them by the reference).

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudBuildConnection
metadata:
  name: acme-github
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  connectionId: acme-github
  annotations:
    owner: platform
  githubConfig:
    appInstallationId: 12345678
    authorizerCredential:
      oauthTokenSecretVersion:
        value: projects/my-gcp-project/secrets/github-oauth-token/versions/latest
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.location` | `string` | yes |  |  |
| `spec.connectionId` | `string` |  |  |  |
| `spec.annotations` | `map<string, string>` |  |  |  |
| `spec.disabled` | `bool` |  |  |  |
| `spec.githubConfig` | `GcpCloudBuildConnectionGithubConfig` |  |  |  |
| `spec.githubConfig.appInstallationId` | `int64` |  |  |  |
| `spec.githubConfig.authorizerCredential` | `GcpCloudBuildConnectionOauthCredential` |  |  |  |
| `spec.githubConfig.authorizerCredential.oauthTokenSecretVersion` | `string \| valueFrom` |  |  | GcpSecretManagerSecret (`status.outputs.latest_version_name`) |
| `spec.githubEnterpriseConfig` | `GcpCloudBuildConnectionGithubEnterpriseConfig` |  |  |  |
| `spec.githubEnterpriseConfig.hostUri` | `string` | yes |  |  |
| `spec.githubEnterpriseConfig.appId` | `int64` |  |  |  |
| `spec.githubEnterpriseConfig.appInstallationId` | `int64` |  |  |  |
| `spec.githubEnterpriseConfig.appSlug` | `string` |  |  |  |
| `spec.githubEnterpriseConfig.privateKeySecretVersion` | `string \| valueFrom` |  |  | GcpSecretManagerSecret (`status.outputs.latest_version_name`) |
| `spec.githubEnterpriseConfig.webhookSecretSecretVersion` | `string \| valueFrom` |  |  | GcpSecretManagerSecret (`status.outputs.latest_version_name`) |
| `spec.githubEnterpriseConfig.sslCa` | `string` |  |  |  |
| `spec.githubEnterpriseConfig.serviceDirectoryConfig` | `GcpCloudBuildConnectionServiceDirectoryConfig` |  |  |  |
| `spec.githubEnterpriseConfig.serviceDirectoryConfig.service` | `string` | yes |  |  |
| `spec.gitlabConfig` | `GcpCloudBuildConnectionGitlabConfig` |  |  |  |
| `spec.gitlabConfig.hostUri` | `string` |  |  |  |
| `spec.gitlabConfig.authorizerCredential` | `GcpCloudBuildConnectionUserTokenCredential` | yes |  |  |
| `spec.gitlabConfig.authorizerCredential.userTokenSecretVersion` | `string \| valueFrom` | yes |  | GcpSecretManagerSecret (`status.outputs.latest_version_name`) |
| `spec.gitlabConfig.readAuthorizerCredential` | `GcpCloudBuildConnectionUserTokenCredential` | yes |  |  |
| `spec.gitlabConfig.readAuthorizerCredential.userTokenSecretVersion` | `string \| valueFrom` | yes |  | GcpSecretManagerSecret (`status.outputs.latest_version_name`) |
| `spec.gitlabConfig.webhookSecretSecretVersion` | `string \| valueFrom` | yes |  | GcpSecretManagerSecret (`status.outputs.latest_version_name`) |
| `spec.gitlabConfig.sslCa` | `string` |  |  |  |
| `spec.gitlabConfig.serviceDirectoryConfig` | `GcpCloudBuildConnectionServiceDirectoryConfig` |  |  |  |
| `spec.gitlabConfig.serviceDirectoryConfig.service` | `string` | yes |  |  |
| `spec.bitbucketCloudConfig` | `GcpCloudBuildConnectionBitbucketCloudConfig` |  |  |  |
| `spec.bitbucketCloudConfig.workspace` | `string` | yes |  |  |
| `spec.bitbucketCloudConfig.authorizerCredential` | `GcpCloudBuildConnectionUserTokenCredential` | yes |  |  |
| `spec.bitbucketCloudConfig.authorizerCredential.userTokenSecretVersion` | `string \| valueFrom` | yes |  | GcpSecretManagerSecret (`status.outputs.latest_version_name`) |
| `spec.bitbucketCloudConfig.readAuthorizerCredential` | `GcpCloudBuildConnectionUserTokenCredential` | yes |  |  |
| `spec.bitbucketCloudConfig.readAuthorizerCredential.userTokenSecretVersion` | `string \| valueFrom` | yes |  | GcpSecretManagerSecret (`status.outputs.latest_version_name`) |
| `spec.bitbucketCloudConfig.webhookSecretSecretVersion` | `string \| valueFrom` | yes |  | GcpSecretManagerSecret (`status.outputs.latest_version_name`) |
| `spec.bitbucketDataCenterConfig` | `GcpCloudBuildConnectionBitbucketDataCenterConfig` |  |  |  |
| `spec.bitbucketDataCenterConfig.hostUri` | `string` | yes |  |  |
| `spec.bitbucketDataCenterConfig.authorizerCredential` | `GcpCloudBuildConnectionUserTokenCredential` | yes |  |  |
| `spec.bitbucketDataCenterConfig.authorizerCredential.userTokenSecretVersion` | `string \| valueFrom` | yes |  | GcpSecretManagerSecret (`status.outputs.latest_version_name`) |
| `spec.bitbucketDataCenterConfig.readAuthorizerCredential` | `GcpCloudBuildConnectionUserTokenCredential` | yes |  |  |
| `spec.bitbucketDataCenterConfig.readAuthorizerCredential.userTokenSecretVersion` | `string \| valueFrom` | yes |  | GcpSecretManagerSecret (`status.outputs.latest_version_name`) |
| `spec.bitbucketDataCenterConfig.webhookSecretSecretVersion` | `string \| valueFrom` | yes |  | GcpSecretManagerSecret (`status.outputs.latest_version_name`) |
| `spec.bitbucketDataCenterConfig.sslCa` | `string` |  |  |  |
| `spec.bitbucketDataCenterConfig.serviceDirectoryConfig` | `GcpCloudBuildConnectionServiceDirectoryConfig` |  |  |  |
| `spec.bitbucketDataCenterConfig.serviceDirectoryConfig.service` | `string` | yes |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The project the connection lives in: a literal project ID or a
GcpProject reference. Empty means the provider's default project.
Immutable.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.location

`string` · required

The region the connection lives in, e.g. "us-central1". Its
repositories, and the triggers that build from them, use the same
region. Required. Immutable.

- rule: {"required":true}

### spec.connectionId

`string`

The connection's ID, unique in the project and region: letters,
digits, and any of -._~%!$&'()*+,;=@ (Google's rule). Defaults to
metadata.name. Immutable.

- rule: connection_id may contain only letters, digits, and -._~%!$&'()*+,;=@

### spec.annotations

`map<string, string>`

Annotations on the connection (AIP-128 key/value metadata; a
connection has no labels). Only the keys declared here are managed.

### spec.disabled

`bool`

Turn the connection off: repository API calls and webhook processing
for every repository in it stop until it is turned back on.

### spec.githubConfig

`GcpCloudBuildConnectionGithubConfig`

A connection to github.com through Cloud Build's GitHub App.

### spec.githubConfig.appInstallationId

`int64`

The installation ID of Cloud Build's GitHub App on the GitHub account
or organization (from the app's settings URL after installing it).
0 leaves the connection waiting for the installation.

- rule: {"int64":{"gte":"0"}}

### spec.githubConfig.authorizerCredential

`GcpCloudBuildConnectionOauthCredential`

The OAuth credential of the GitHub account that authorized Cloud
Build's GitHub App -- a robot account is recommended over a person.

### spec.githubConfig.authorizerCredential.oauthTokenSecretVersion

`string | valueFrom`

The Secret Manager secret version holding the OAuth token, as
projects/{project}/secrets/{secret}/versions/{version}: a
GcpSecretManagerSecret reference (its latest_version_name output, set
when the secret declares an initial version) or a literal version name
("versions/latest" follows rotation). The token must be tied to Cloud
Build's GitHub App.

- references: GcpSecretManagerSecret (`status.outputs.latest_version_name`)
- rule: oauth_token_secret_version must be projects/{project}/secrets/{secret}/versions/{version}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSecretManagerSecret, name: <that resource's name>, fieldPath: status.outputs.latest_version_name}} -- a bare string does not parse

### spec.githubEnterpriseConfig

`GcpCloudBuildConnectionGithubEnterpriseConfig`

A connection to a GitHub Enterprise server through a GitHub App
created on that server.

### spec.githubEnterpriseConfig.hostUri

`string` · required

The server's URI, e.g. "https://github.example.com". Required.

- rule: {"required":true}

### spec.githubEnterpriseConfig.appId

`int64`

The GitHub App's ID.

- rule: {"int64":{"gte":"0"}}

### spec.githubEnterpriseConfig.appInstallationId

`int64`

The GitHub App's installation ID on the organization.

- rule: {"int64":{"gte":"0"}}

### spec.githubEnterpriseConfig.appSlug

`string`

The GitHub App's URL-friendly name.

### spec.githubEnterpriseConfig.privateKeySecretVersion

`string | valueFrom`

The Secret Manager secret version holding the GitHub App's private key
(projects/{project}/secrets/{secret}/versions/{version}): a
GcpSecretManagerSecret reference or a literal version name.

- references: GcpSecretManagerSecret (`status.outputs.latest_version_name`)
- rule: private_key_secret_version must be projects/{project}/secrets/{secret}/versions/{version}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSecretManagerSecret, name: <that resource's name>, fieldPath: status.outputs.latest_version_name}} -- a bare string does not parse

### spec.githubEnterpriseConfig.webhookSecretSecretVersion

`string | valueFrom`

The Secret Manager secret version holding the GitHub App's webhook
secret: a GcpSecretManagerSecret reference or a literal version name.
Updates in place.

- references: GcpSecretManagerSecret (`status.outputs.latest_version_name`)
- rule: webhook_secret_secret_version must be projects/{project}/secrets/{secret}/versions/{version}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSecretManagerSecret, name: <that resource's name>, fieldPath: status.outputs.latest_version_name}} -- a bare string does not parse

### spec.githubEnterpriseConfig.sslCa

`string`

The PEM CA certificate Cloud Build trusts when calling the server, for
a server with a private certificate authority.

### spec.githubEnterpriseConfig.serviceDirectoryConfig

`GcpCloudBuildConnectionServiceDirectoryConfig`

Reach an on-premises server privately through Service Directory
instead of over the internet.

### spec.githubEnterpriseConfig.serviceDirectoryConfig.service

`string` · required

The Service Directory service, as
projects/{project}/locations/{location}/namespaces/{namespace}/services/{service}.
Required.

- rule: service must be projects/{project}/locations/{location}/namespaces/{namespace}/services/{service}
- rule: {"required":true}

### spec.gitlabConfig

`GcpCloudBuildConnectionGitlabConfig`

A connection to gitlab.com or a GitLab Enterprise server.

### spec.gitlabConfig.hostUri

`string`

The GitLab server's URI. Empty means https://gitlab.com.

### spec.gitlabConfig.authorizerCredential

`GcpCloudBuildConnectionUserTokenCredential` · required

A personal access token with the "api" scope, which Cloud Build uses
to create webhooks and read repositories. Required.

- rule: {"required":true}

### spec.gitlabConfig.authorizerCredential.userTokenSecretVersion

`string | valueFrom` · required

The Secret Manager secret version holding the token, as
projects/{project}/secrets/{secret}/versions/{version}: a
GcpSecretManagerSecret reference (its latest_version_name output, set
when the secret declares an initial version) or a literal version name
("versions/latest" follows rotation). Required.

- references: GcpSecretManagerSecret (`status.outputs.latest_version_name`)
- rule: user_token_secret_version must be projects/{project}/secrets/{secret}/versions/{version}
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSecretManagerSecret, name: <that resource's name>, fieldPath: status.outputs.latest_version_name}} -- a bare string does not parse

### spec.gitlabConfig.readAuthorizerCredential

`GcpCloudBuildConnectionUserTokenCredential` · required

A personal access token with at least the "read_api" scope, which
Cloud Build uses for read-only calls. Required.

- rule: {"required":true}

### spec.gitlabConfig.readAuthorizerCredential.userTokenSecretVersion

`string | valueFrom` · required

The Secret Manager secret version holding the token, as
projects/{project}/secrets/{secret}/versions/{version}: a
GcpSecretManagerSecret reference (its latest_version_name output, set
when the secret declares an initial version) or a literal version name
("versions/latest" follows rotation). Required.

- references: GcpSecretManagerSecret (`status.outputs.latest_version_name`)
- rule: user_token_secret_version must be projects/{project}/secrets/{secret}/versions/{version}
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSecretManagerSecret, name: <that resource's name>, fieldPath: status.outputs.latest_version_name}} -- a bare string does not parse

### spec.gitlabConfig.webhookSecretSecretVersion

`string | valueFrom` · required

The Secret Manager secret version holding the webhook secret of the
GitLab project: a GcpSecretManagerSecret reference or a literal version
name. Required. Immutable (a change replaces the connection).

- references: GcpSecretManagerSecret (`status.outputs.latest_version_name`)
- rule: webhook_secret_secret_version must be projects/{project}/secrets/{secret}/versions/{version}
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSecretManagerSecret, name: <that resource's name>, fieldPath: status.outputs.latest_version_name}} -- a bare string does not parse

### spec.gitlabConfig.sslCa

`string`

The PEM CA certificate Cloud Build trusts when calling a GitLab
Enterprise server with a private certificate authority.

### spec.gitlabConfig.serviceDirectoryConfig

`GcpCloudBuildConnectionServiceDirectoryConfig`

Reach an on-premises GitLab Enterprise server privately through
Service Directory instead of over the internet.

### spec.gitlabConfig.serviceDirectoryConfig.service

`string` · required

The Service Directory service, as
projects/{project}/locations/{location}/namespaces/{namespace}/services/{service}.
Required.

- rule: service must be projects/{project}/locations/{location}/namespaces/{namespace}/services/{service}
- rule: {"required":true}

### spec.bitbucketCloudConfig

`GcpCloudBuildConnectionBitbucketCloudConfig`

A connection to a Bitbucket Cloud workspace.

### spec.bitbucketCloudConfig.workspace

`string` · required

The Bitbucket Cloud workspace ID to connect. Required.

- rule: {"required":true}

### spec.bitbucketCloudConfig.authorizerCredential

`GcpCloudBuildConnectionUserTokenCredential` · required

A workspace, project, or repository access token with the "webhook",
"repository", "repository:admin", and "pullrequest" scopes (a system
account is recommended). Required.

- rule: {"required":true}

### spec.bitbucketCloudConfig.authorizerCredential.userTokenSecretVersion

`string | valueFrom` · required

The Secret Manager secret version holding the token, as
projects/{project}/secrets/{secret}/versions/{version}: a
GcpSecretManagerSecret reference (its latest_version_name output, set
when the secret declares an initial version) or a literal version name
("versions/latest" follows rotation). Required.

- references: GcpSecretManagerSecret (`status.outputs.latest_version_name`)
- rule: user_token_secret_version must be projects/{project}/secrets/{secret}/versions/{version}
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSecretManagerSecret, name: <that resource's name>, fieldPath: status.outputs.latest_version_name}} -- a bare string does not parse

### spec.bitbucketCloudConfig.readAuthorizerCredential

`GcpCloudBuildConnectionUserTokenCredential` · required

An access token with "repository" access, for read-only calls.
Required.

- rule: {"required":true}

### spec.bitbucketCloudConfig.readAuthorizerCredential.userTokenSecretVersion

`string | valueFrom` · required

The Secret Manager secret version holding the token, as
projects/{project}/secrets/{secret}/versions/{version}: a
GcpSecretManagerSecret reference (its latest_version_name output, set
when the secret declares an initial version) or a literal version name
("versions/latest" follows rotation). Required.

- references: GcpSecretManagerSecret (`status.outputs.latest_version_name`)
- rule: user_token_secret_version must be projects/{project}/secrets/{secret}/versions/{version}
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSecretManagerSecret, name: <that resource's name>, fieldPath: status.outputs.latest_version_name}} -- a bare string does not parse

### spec.bitbucketCloudConfig.webhookSecretSecretVersion

`string | valueFrom` · required

The Secret Manager secret version holding the webhook secret Cloud
Build verifies webhook events with: a GcpSecretManagerSecret reference
or a literal version name. Required. Immutable (a change replaces the
connection).

- references: GcpSecretManagerSecret (`status.outputs.latest_version_name`)
- rule: webhook_secret_secret_version must be projects/{project}/secrets/{secret}/versions/{version}
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSecretManagerSecret, name: <that resource's name>, fieldPath: status.outputs.latest_version_name}} -- a bare string does not parse

### spec.bitbucketDataCenterConfig

`GcpCloudBuildConnectionBitbucketDataCenterConfig`

A connection to a Bitbucket Data Center server.

### spec.bitbucketDataCenterConfig.hostUri

`string` · required

The Bitbucket Data Center server's URI. Required.

- rule: {"required":true}

### spec.bitbucketDataCenterConfig.authorizerCredential

`GcpCloudBuildConnectionUserTokenCredential` · required

An HTTP access token with the "REPO_ADMIN" scope. Required.

- rule: {"required":true}

### spec.bitbucketDataCenterConfig.authorizerCredential.userTokenSecretVersion

`string | valueFrom` · required

The Secret Manager secret version holding the token, as
projects/{project}/secrets/{secret}/versions/{version}: a
GcpSecretManagerSecret reference (its latest_version_name output, set
when the secret declares an initial version) or a literal version name
("versions/latest" follows rotation). Required.

- references: GcpSecretManagerSecret (`status.outputs.latest_version_name`)
- rule: user_token_secret_version must be projects/{project}/secrets/{secret}/versions/{version}
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSecretManagerSecret, name: <that resource's name>, fieldPath: status.outputs.latest_version_name}} -- a bare string does not parse

### spec.bitbucketDataCenterConfig.readAuthorizerCredential

`GcpCloudBuildConnectionUserTokenCredential` · required

An HTTP access token with "REPO_READ" access, for read-only calls.
Required.

- rule: {"required":true}

### spec.bitbucketDataCenterConfig.readAuthorizerCredential.userTokenSecretVersion

`string | valueFrom` · required

The Secret Manager secret version holding the token, as
projects/{project}/secrets/{secret}/versions/{version}: a
GcpSecretManagerSecret reference (its latest_version_name output, set
when the secret declares an initial version) or a literal version name
("versions/latest" follows rotation). Required.

- references: GcpSecretManagerSecret (`status.outputs.latest_version_name`)
- rule: user_token_secret_version must be projects/{project}/secrets/{secret}/versions/{version}
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSecretManagerSecret, name: <that resource's name>, fieldPath: status.outputs.latest_version_name}} -- a bare string does not parse

### spec.bitbucketDataCenterConfig.webhookSecretSecretVersion

`string | valueFrom` · required

The Secret Manager secret version holding the webhook secret: a
GcpSecretManagerSecret reference or a literal version name. Required.
Immutable (a change replaces the connection).

- references: GcpSecretManagerSecret (`status.outputs.latest_version_name`)
- rule: webhook_secret_secret_version must be projects/{project}/secrets/{secret}/versions/{version}
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSecretManagerSecret, name: <that resource's name>, fieldPath: status.outputs.latest_version_name}} -- a bare string does not parse

### spec.bitbucketDataCenterConfig.sslCa

`string`

The PEM CA certificate Cloud Build trusts when calling the server.

### spec.bitbucketDataCenterConfig.serviceDirectoryConfig

`GcpCloudBuildConnectionServiceDirectoryConfig`

Reach an on-premises server privately through Service Directory
instead of over the internet.

### spec.bitbucketDataCenterConfig.serviceDirectoryConfig.service

`string` · required

The Service Directory service, as
projects/{project}/locations/{location}/namespaces/{namespace}/services/{service}.
Required.

- rule: service must be projects/{project}/locations/{location}/namespaces/{namespace}/services/{service}
- rule: {"required":true}

### spec.deletionPolicy

`string`

What destroy does:
  "" / "DELETE" -- the connection is deleted
  "PREVENT"     -- destroy fails
  "ABANDON"     -- the connection leaves management and stays in Google

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `spec.one_code_host`: set at most one of github_config, github_enterprise_config, gitlab_config, bitbucket_cloud_config, or bitbucket_data_center_config

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpCloudBuildConnection, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name: projects/{project}/locations/{location}/connections/{connection_id}. The value a GcpCloudBuildRepository's parent_connection takes. |
| `status.outputs.connection_id` | `string` | The connection's ID. |
| `status.outputs.installation_stage` | `string` | The current installation step: PENDING_CREATE_APP, PENDING_USER_OAUTH, PENDING_INSTALL_APP, or COMPLETE. |
| `status.outputs.installation_action_uri` | `string` | The link a person follows to finish the installation (for GitHub, installing Cloud Build's GitHub App); empty once complete. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.githubConfig.authorizerCredential.oauthTokenSecretVersion` | GcpSecretManagerSecret | `status.outputs.latest_version_name` |
| `spec.githubEnterpriseConfig.privateKeySecretVersion` | GcpSecretManagerSecret | `status.outputs.latest_version_name` |
| `spec.githubEnterpriseConfig.webhookSecretSecretVersion` | GcpSecretManagerSecret | `status.outputs.latest_version_name` |
| `spec.gitlabConfig.authorizerCredential.userTokenSecretVersion` | GcpSecretManagerSecret | `status.outputs.latest_version_name` |
| `spec.gitlabConfig.readAuthorizerCredential.userTokenSecretVersion` | GcpSecretManagerSecret | `status.outputs.latest_version_name` |
| `spec.gitlabConfig.webhookSecretSecretVersion` | GcpSecretManagerSecret | `status.outputs.latest_version_name` |
| `spec.bitbucketCloudConfig.authorizerCredential.userTokenSecretVersion` | GcpSecretManagerSecret | `status.outputs.latest_version_name` |
| `spec.bitbucketCloudConfig.readAuthorizerCredential.userTokenSecretVersion` | GcpSecretManagerSecret | `status.outputs.latest_version_name` |
| `spec.bitbucketCloudConfig.webhookSecretSecretVersion` | GcpSecretManagerSecret | `status.outputs.latest_version_name` |
| `spec.bitbucketDataCenterConfig.authorizerCredential.userTokenSecretVersion` | GcpSecretManagerSecret | `status.outputs.latest_version_name` |
| `spec.bitbucketDataCenterConfig.readAuthorizerCredential.userTokenSecretVersion` | GcpSecretManagerSecret | `status.outputs.latest_version_name` |
| `spec.bitbucketDataCenterConfig.webhookSecretSecretVersion` | GcpSecretManagerSecret | `status.outputs.latest_version_name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpCloudBuildRepository | `spec.parentConnection` | `status.outputs.name` |

## See Also

- [Overview](../README.md)

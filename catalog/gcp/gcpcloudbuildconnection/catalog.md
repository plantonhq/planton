# GCP Cloud Build Connection

Connects Cloud Build to your code host -- github.com, GitHub Enterprise, gitlab.com or GitLab Enterprise, Bitbucket Cloud, or Bitbucket Data Center -- so builds can be triggered by pushes and pull requests and can read your repositories. Tokens stay in Secret Manager; the connection only names the secret versions.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- the Cloud Build API on the connection's project
- **Connection** -- an authorized link to one code host, the container its linked repositories live in

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with permission to manage Cloud Build connections in the target project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Code Host Account

- **Credentials in Secret Manager** -- the host's access tokens (and webhook secret) stored as Secret Manager secrets, with Google's Cloud Build service agent granted `roles/secretmanager.secretAccessor` on each.
- **For GitHub** -- Cloud Build's GitHub App installed on your account or organization; its installation ID goes into the spec.

## Deploy

### Console

Open the deployment store, find **GCP Cloud Build Connection**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **GitHub Connection** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudBuildConnection
metadata:
  name: acme-github
  org: acme-corp
  env: prod
spec:
  projectId:
    value: acme-ci
  location: us-central1
  githubConfig:
    appInstallationId: 12345678
    authorizerCredential:
      oauthTokenSecretVersion:
        value: projects/acme-ci/secrets/github-token/versions/latest
```

```shell
planton apply -f cloud-build-connection.yaml
```

This connects Cloud Build in us-central1 to the GitHub organization where the app is installed. A Stack Job tracks the provisioning in real time.

### InfraChart

Reference a project's `status.outputs.project_id` from `projectId`, and each `GcpSecretManagerSecret`'s `status.outputs.latest_version_name` from the credential and webhook-secret fields.

## Key Configuration

These are the most important decisions when configuring a connection. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Code host** -- one block per connection: `githubConfig`, `githubEnterpriseConfig`, `gitlabConfig`, `bitbucketCloudConfig`, or `bitbucketDataCenterConfig`.

**Credentials** -- every token and secret is a Secret Manager version name; `versions/latest` follows rotation.

**Private hosts** -- `serviceDirectoryConfig` reaches an on-premises server without the internet, and `sslCa` trusts a private certificate authority.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpSecretManagerSecret** | every `*SecretVersion` field | `status.outputs.latest_version_name` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `name` | The connection's full resource name | A repository's `parentConnection` |
| `connection_id` | The connection's ID | Tooling |
| `installation_stage` | The installation step reached | Readiness checks |
| `installation_action_uri` | The link that finishes the installation | Handing a person the next step |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**GitHub** -- Cloud Build's GitHub App on your organization. Start from the **GitHub Connection** preset.

**GitLab** -- gitlab.com or a GitLab Enterprise server with personal access tokens. Start from the **GitLab Connection** preset.

## Works With

- [**GCP Cloud Build Repository**](/cloud-catalog/gcp-cloud-build-repository) -- repositories linked through the connection
- [**GCP Cloud Build Trigger**](/cloud-catalog/gcp-cloud-build-trigger) -- builds triggered by the repositories' events
- [**GCP Secret Manager Secret**](/cloud-catalog/gcp-secret-manager-secret) -- where the tokens live

# GCP Cloud Build Connection

Declares a Cloud Build repository connection: Cloud Build's authorized link to one code host -- github.com through Cloud Build's GitHub App, a GitHub Enterprise server through an app created on it, gitlab.com or GitLab Enterprise, Bitbucket Cloud, or Bitbucket Data Center. The connection is the container its linked repositories live in. Every credential is a Secret Manager secret version, never a value in the manifest.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **API enablement** -- `cloudbuild.googleapis.com` on the connection's project (never disabled on destroy)
- **Connection** -- one `cloudbuildv2_connection`

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with `roles/cloudbuild.connectionAdmin` (or the permissions in `iac/permissions.yaml`) on the connection's project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Required Setup

- **Secrets** -- each token and webhook secret stored in Secret Manager (a **`GcpSecretManagerSecret`** with an initial version), with `service-{PROJECT_NUMBER}@gcp-sa-cloudbuild.iam.gserviceaccount.com` granted `roles/secretmanager.secretAccessor` on it (the secret's `iamMembers`).
- **GitHub only** -- Cloud Build's GitHub App installed on the account or organization; the installation ID is in the app's settings URL.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudBuildConnection
metadata:
  name: acme-gitlab
spec:
  projectId:
    value: acme-ci
  location: us-central1
  gitlabConfig:
    authorizerCredential:
      userTokenSecretVersion:
        valueFrom:
          kind: GcpSecretManagerSecret
          name: gitlab-api-token
    readAuthorizerCredential:
      userTokenSecretVersion:
        valueFrom:
          kind: GcpSecretManagerSecret
          name: gitlab-read-token
    webhookSecretSecretVersion:
      valueFrom:
        kind: GcpSecretManagerSecret
        name: gitlab-webhook-secret
```

```shell
planton apply -f cloud-build-connection.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `location` | `string` | The connection's region. Immutable. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `string` / ref | provider project | The connection's project (`GcpProject` ref). Immutable. |
| `connectionId` | `string` | `metadata.name` | The connection's ID (letters, digits, `-._~%!$&'()*+,;=@`). Immutable. |
| `annotations` | `map` | -- | AIP-128 annotations; only declared keys are managed. |
| `disabled` | `bool` | `false` | Stops repository calls and webhook processing. |
| `githubConfig` | object | -- | `appInstallationId`; `authorizerCredential.oauthTokenSecretVersion`. |
| `githubEnterpriseConfig` | object | -- | `hostUri` (required), `appId`, `appInstallationId`, `appSlug`, `privateKeySecretVersion`, `webhookSecretSecretVersion`, `sslCa`, `serviceDirectoryConfig.service`. |
| `gitlabConfig` | object | -- | `hostUri` (default gitlab.com), `authorizerCredential` and `readAuthorizerCredential` (`userTokenSecretVersion`), `webhookSecretSecretVersion`, `sslCa`, `serviceDirectoryConfig.service`. |
| `bitbucketCloudConfig` | object | -- | `workspace`, `authorizerCredential`, `readAuthorizerCredential`, `webhookSecretSecretVersion`. |
| `bitbucketDataCenterConfig` | object | -- | `hostUri`, `authorizerCredential`, `readAuthorizerCredential`, `webhookSecretSecretVersion`, `sslCa`, `serviceDirectoryConfig.service`. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`. |

Every `*SecretVersion` field takes a `GcpSecretManagerSecret` reference (its `latest_version_name`) or a literal `projects/{p}/secrets/{s}/versions/{v}`.

### Validation Rules

- At most one code-host block.
- Literal secret versions are full version names; a Service Directory service is a full service name.

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `projects/{project}/locations/{location}/connections/{connection_id}` |
| `connection_id` | `string` | The connection's ID |
| `installation_stage` | `string` | `PENDING_CREATE_APP`, `PENDING_USER_OAUTH`, `PENDING_INSTALL_APP`, or `COMPLETE` |
| `installation_action_uri` | `string` | The link that finishes the installation; empty once complete |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Grant the service agent first.** Cloud Build's service agent reads every secret; without `secretAccessor` on each, the create fails.
- **GitHub finishes in a browser.** Until the app is installed and its installation ID set, `installation_stage` is not `COMPLETE` and `installation_action_uri` is the link to follow.
- **Webhook secrets.** GitLab and both Bitbucket hosts treat `webhookSecretSecretVersion` as immutable (a change replaces the connection); GitHub Enterprise updates it in place.
- **Destroy order.** Remove the connection's repositories first; a chart orders them by the repository's reference.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Kinds

- **GcpCloudBuildRepository** -- repositories linked through this connection
- **GcpCloudBuildTrigger** -- builds started by those repositories' events
- **GcpSecretManagerSecret** -- the credentials

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

# GCP IAM OAuth Client

Creates a Workforce Identity Federation OAuth client — the registration that lets an application obtain Google Cloud access tokens on behalf of workforce-federated users via OAuth 2.0. The client carries its allowed grant types, scopes, and redirect URIs, plus managed credentials whose secrets GCP generates server-side and the platform delivers to consumers through ValueFromRef.

Scope honesty: this is the ONLY kind of OAuth client Google's APIs can create programmatically. Classic consent-screen OAuth clients (end-user Google Sign-In) have no programmatic path — Google shut down the IAP OAuth Admin API that once created them (March 2026). Those clients remain a documented console step whose ID/secret feed GcpIdentityPlatformConfig's `defaultSupportedIdps` or a GcpSecretManagerSecret.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **OAuth Client** -- a `google_iam_oauth_client` (the Workforce Identity Federation OAuth registration) with grant types, scopes, redirect URIs, and confidentiality model
- **OAuth Client Credentials** -- one `google_iam_oauth_client_credential` per `credentials` entry; secrets are generated server-side by GCP
- **IAM API enablement** -- `iam.googleapis.com` enabled in the target project (never disabled on destroy)

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** where the client is created (directly or via a GcpProject reference).
- **IAM**: the deploying identity needs `roles/iam.workforcePoolAdmin` or broader on the project.

## Deploy

### Console

Open the deployment store, find **GCP IAM OAuth Client**, and click **Deploy**. The creation wizard walks you through environment and connection configuration, the grant types, scopes, and redirect URIs, and an inline builder for the managed credentials. Start from the **Web App Client** preset in the [Presets](#presets) tab for the most common shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpIamOauthClient
metadata:
  name: web-app-client
  org: acme-corp
  env: prod
spec:
  clientType: CONFIDENTIAL_CLIENT
  allowedGrantTypes:
    - AUTHORIZATION_CODE_GRANT
    - REFRESH_TOKEN_GRANT
  allowedScopes:
    - https://www.googleapis.com/auth/cloud-platform
    - openid
  allowedRedirectUris:
    - value: https://app.example.com/callback
  credentials:
    - credentialId: primary
```

```shell
planton apply -f client.yaml
```

This creates a confidential client with one managed credential; its server-generated secret is the `client_secret` output. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, the redirect URI can track a deployed service and the secret can flow into Secret Manager:

```yaml
# On this GcpIamOauthClient — the redirect URI follows the app's address:
spec:
  allowedRedirectUris:
    - valueFrom:
        kind: GcpCloudRun
        name: my-web-app
        fieldPath: status.outputs.url
```

```yaml
# On a GcpSecretManagerSecret in the same chart — the secret never leaves the graph:
spec:
  initialVersion:
    data:
      valueFrom:
        kind: GcpIamOauthClient
        name: web-app-client
        fieldPath: status.outputs.client_secret
```

The InfraPipeline deploys the app first, registers the client with the resolved URL, then stores the generated secret.

## Key Configuration

These are the most important decisions when configuring an OAuth client. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Client type** -- only `CONFIDENTIAL_CLIENT` (server-side apps — manage secrets via `credentials`) can be created: GCP's enum lists `PUBLIC_CLIENT` (mobile apps, SPAs) but the service rejects creating one ("Client type is not supported"). Immutable.

**Grant types and scopes** -- Google's API accepts exactly `AUTHORIZATION_CODE_GRANT` and `REFRESH_TOKEN_GRANT`; scopes are the OAuth scopes the client may request during flows. Grant only what the application uses.

**Credentials** -- each entry creates one managed credential whose secret GCP generates server-side. GCP requires a credential to be DISABLED before it can be deleted: disable in one apply, remove in the next. The first credential's secret is the `client_secret` output.

**Disabled** -- `disabled: true` stops new authorizations without deleting the client — the reversible kill switch.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** (optional) | `projectId` | `status.outputs.project_id` |
| **GcpCloudRun** (optional) | `allowedRedirectUris` entries | `status.outputs.url` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `client_id` | The OAuth client ID applications present in flows | Application configuration |
| `client_name` | `projects/{project}/locations/{location}/oauthClients/{id}` | Tooling that addresses the client by full name |
| `state` | Lifecycle state (`ACTIVE`, `DELETED` during soft-delete) | Health checks |
| `client_secret` | The first credential's server-generated secret | GcpSecretManagerSecret initial version — the secret never leaves the graph |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Server-side web app** -- a confidential client with both grant types and one managed credential; the secret flows to the app through Secret Manager. Start from the **Web App Client** preset.

**Single-page app** -- no workforce-federation path exists today: GCP rejects public-client creation at the API, and a confidential client's secret cannot live in a browser bundle. Until GCP ships public-client support, SPAs authenticate through a confidential backend-for-frontend that holds the credential server-side.

## Works With

- [**GCP Identity Platform Config**](/infra-catalog/gcp-identity-platform-config) -- consumes console-created consent-screen client IDs/secrets for end-user sign-in
- [**GCP Secret Manager Secret**](/infra-catalog/gcp-secret-manager-secret) -- durable home for the generated client secret
- [**GCP Cloud Run**](/infra-catalog/gcp-cloud-run) -- its URL output feeds the redirect URIs
- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project where the client is created

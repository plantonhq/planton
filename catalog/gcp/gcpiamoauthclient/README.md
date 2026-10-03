# GCP IAM OAuth Client

Creates a Workforce Identity Federation OAuth client — the registration that lets an application obtain Google Cloud access tokens on behalf of workforce-federated users via OAuth 2.0. The client carries its allowed grant types, scopes, and redirect URIs, plus managed credentials whose secrets GCP generates server-side. Redirect URIs can reference other resources' URL outputs, and the client secret flows to consumers through ValueFromRef — the OAuth registration becomes one declarative graph.

## Scope: Workforce OAuth Clients Only

This kind models **workforce** OAuth clients — the ONLY kind of OAuth client Google's APIs can create programmatically. Classic consent-screen OAuth clients (the ones behind end-user Google Sign-In) have NO programmatic path left: Google permanently shut down the IAP OAuth Admin API that once created them in March 2026. Consent-screen clients remain a documented console step; their ID and secret feed [GcpIdentityPlatformConfig](/docs/catalog/gcp/gcpidentityplatformconfig)'s `defaultSupportedIdps` or a [GcpSecretManagerSecret](/docs/catalog/gcp/gcpsecretmanagersecret). If you need Google Sign-In for consumers, create the client in the console and wire its values in — this kind will not create it for you.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **OAuth Client** -- a `google_iam_oauth_client` (the Workforce Identity Federation OAuth registration) carrying the grant types, scopes, redirect URIs, and confidentiality model
- **OAuth Client Credentials** -- one `google_iam_oauth_client_credential` per entry in `credentials`; the secret value is generated server-side by GCP, never supplied in the manifest
- **IAM API enablement** -- `iam.googleapis.com` enabled in the target project (never disabled on destroy)

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** where the client is created. Provide the project ID directly or reference a GcpProject Infra Component via ValueFromRef.
- **IAM**: the deploying principal's permissions are listed in [`iac/permissions.yaml`](iac/permissions.yaml).

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpIamOauthClient
metadata:
  name: web-app-client
spec:
  clientType: CONFIDENTIAL_CLIENT
  allowedGrantTypes:
    - AUTHORIZATION_CODE_GRANT
    - REFRESH_TOKEN_GRANT
  allowedScopes:
    - https://www.googleapis.com/auth/cloud-platform
  allowedRedirectUris:
    - value: https://app.example.com/callback
  credentials:
    - credentialId: primary
```

```shell
planton apply -f client.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `allowedGrantTypes` | `[]string` | Grant types the client may use: `AUTHORIZATION_CODE_GRANT`, `REFRESH_TOKEN_GRANT`. | Min 1 entry; only those two values |
| `allowedScopes` | `[]string` | OAuth scopes the client may request (e.g. `https://www.googleapis.com/auth/cloud-platform`, `openid`, `email`, `profile`, `groups`). | Min 1 entry |
| `allowedRedirectUris` | `[]StringValueOrRef` | Redirect URIs allowed when authorization completes. Each entry is a literal URL or a reference to another resource's URL output (e.g. a GcpCloudRun `url`). | Min 1 entry |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider default | GCP project. Can reference a GcpProject resource. |
| `location` | `string` | `global` | The client's location; `global` is the documented home for workforce OAuth clients. Immutable. |
| `oauthClientId` | `string` | `metadata.name` | The client's resource ID. Immutable: changing it destroys and recreates the client — consumers must re-register. |
| `displayName` | `string` | `""` | Name shown in consoles and consent surfaces. |
| `description` | `string` | `""` | What this client is for — the operator-facing record. |
| `disabled` | `bool` | `false` | When true, the client stops accepting new authorizations without being deleted — the reversible kill switch. |
| `clientType` | `string` | GCP default | Only `CONFIDENTIAL_CLIENT` (server-side; manage secrets via `credentials`) can be created — GCP's enum lists `PUBLIC_CLIENT` but the service rejects it ("Client type is not supported", live-verified). Immutable. |
| `credentials` | `[]object` | `[]` | Managed client secrets (CONFIDENTIAL_CLIENT only). Each entry needs a `credentialId`; `displayName` and `disabled` are optional. |
| `deletionPolicy` | `string` | `DELETE` | One switch governing the client AND its credentials: `DELETE`, `PREVENT` (refuse), or `ABANDON` (keep working, drop from management). |

### Validation Rules

- **Grant types are a closed set**: only `AUTHORIZATION_CODE_GRANT` and `REFRESH_TOKEN_GRANT` are accepted — Google's IAM API defines exactly these.
- **Client type is `CONFIDENTIAL_CLIENT` only**: GCP rejects `PUBLIC_CLIENT` creation at the API ("Client type is not supported"); the validation re-admits it if GCP ever ships support.
- **Every credential needs a `credentialId`** — it is the credential's immutable resource ID.

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `client_id` | `string` | The system-generated OAuth client ID applications present in OAuth flows (distinct from the user-chosen resource ID) |
| `client_name` | `string` | `projects/{project}/locations/{location}/oauthClients/{id}` — the full resource name |
| `state` | `string` | Lifecycle state (`ACTIVE`, or `DELETED` during the soft-delete window) |
| `client_secret` | `string` | The system-generated secret of the FIRST credential in `credentials` (empty when none are defined). Marked secret in state — feed it to consumers via ValueFromRef, never by copy-paste |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Credentials must be DISABLED before they can be deleted** — GCP refuses to delete an ENABLED credential. Removing an entry from `credentials` while it is still enabled fails at the API: set `disabled: true` in one apply, remove the entry in the next.
- **The first credential's secret is the `client_secret` output.** The single-credential case is the operating norm — rotation adds a second credential and swaps consumers over, at which point the remaining credential is again the first. Wire it via ValueFromRef (e.g. into a GcpSecretManagerSecret initial version); a copy-pasted secret is a rotation that never happens.
- **Deleted clients are soft-deleted for about 30 days** — the client ID stays reserved and cannot be reused until the window expires. Pick IDs you will not need to recreate immediately.
- **`oauthClientId`, `location`, and `clientType` are immutable** — changing any of them destroys and recreates the client, and every consumer must re-register the new client ID.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Kinds

- [GcpIdentityPlatformConfig](/docs/catalog/gcp/gcpidentityplatformconfig) — consumes console-created consent-screen client IDs/secrets in `defaultSupportedIdps`
- [GcpSecretManagerSecret](/docs/catalog/gcp/gcpsecretmanagersecret) — durable home for the `client_secret` output via ValueFromRef
- [GcpCloudRun](/docs/catalog/gcp/gcpcloudrun) — its `url` output feeds `allowedRedirectUris`, killing redirect-URI drift
- [GcpProject](/docs/catalog/gcp/gcpproject) — provides the GCP project and API enablement

## Additional Resources

- [Workforce Identity Federation OAuth Clients](https://cloud.google.com/iam/docs/workforce-oauth-app)
- [OauthClients API Reference](https://cloud.google.com/iam/docs/reference/rest/v1/projects.locations.oauthClients)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

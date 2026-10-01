# GcpIamOauthClient

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpIamOauthClientSpec defines a WORKFORCE Identity Federation OAuth
client — the registration that lets an application obtain Google Cloud
access tokens on behalf of workforce-federated users via OAuth 2.0,
with its allowed grant types, scopes, redirect URIs, and managed
credentials (client secrets).

Scope honesty: this is the only kind of OAuth client Google's APIs can
create programmatically. Classic consent-screen OAuth clients (end-user
Google Sign-In) have NO programmatic path — Google shut down the IAP
OAuth Admin API that once created them (March 2026). Those clients
remain a documented console step whose ID/secret feed
GcpIdentityPlatformConfig's default_supported_idps or a
GcpSecretManagerSecret.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpIamOauthClient
metadata:
  name: my-sample-oauth-client
spec:
  # GCP project that owns the OAuth client.
  # Omit to use the provider's default project.
  projectId:
    value: my-gcp-project-123

  # "global" is the documented home for workforce OAuth clients and the
  # default. Immutable.
  location: global

  # The client's resource ID (the last segment of its name). Defaults to
  # metadata.name when omitted. Immutable: changing it destroys and
  # recreates the client — consumers must re-register.
  oauthClientId: my-sample-oauth-client

  # Shown in consoles and consent surfaces.
  displayName: Sample workforce OAuth client

  # What this client is for — the operator-facing record.
  description: Canonical full-surface example of a workforce OAuth client

  # When true, the client stops accepting new authorizations without
  # being deleted — the reversible kill switch. Default false.
  disabled: false

  # PUBLIC_CLIENT (mobile/SPA; cannot hold secrets) or
  # CONFIDENTIAL_CLIENT (server-side; manage secrets via credentials).
  # Immutable.
  clientType: CONFIDENTIAL_CLIENT

  # Grant types the client may use — Google's API accepts exactly these
  # two values (a closed enum in the IAM REST API).
  allowedGrantTypes:
    - AUTHORIZATION_CODE_GRANT
    - REFRESH_TOKEN_GRANT

  # OAuth scopes the client may request during flows.
  allowedScopes:
    - https://www.googleapis.com/auth/cloud-platform
    - openid

  # Redirect URIs allowed when authorization completes. Each entry may
  # also be a valueFrom reference to another resource's URL output (e.g.
  # a GcpCloudRun url) — referencing kills the drift between the deployed
  # app's address and the client's registration.
  allowedRedirectUris:
    - value: https://app.example.com/callback

  # Managed client secrets (CONFIDENTIAL_CLIENT only). Each entry creates
  # one credential whose secret GCP generates server-side — the FIRST
  # credential's secret is the client_secret output. GCP refuses to
  # delete an ENABLED credential: to remove an entry, set disabled: true
  # in one apply, then delete the entry in the next.
  credentials:
    - credentialId: primary
      displayName: Primary application secret
      disabled: false

  # One switch governs the client AND its credentials:
  # DELETE (default), PREVENT, or ABANDON. Deleted clients are
  # soft-deleted ~30 days; the client ID stays reserved.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.location` | `string` |  | `global` |  |
| `spec.oauthClientId` | `string` |  |  |  |
| `spec.displayName` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.disabled` | `bool` |  |  |  |
| `spec.clientType` | `string` |  |  |  |
| `spec.allowedGrantTypes` | `[]string` | yes |  |  |
| `spec.allowedScopes` | `[]string` | yes |  |  |
| `spec.allowedRedirectUris` | `[]string \| valueFrom` | yes |  | GcpCloudRun (`status.outputs.url`) |
| `spec.credentials` | `[]GcpIamOauthClientCredential` |  |  |  |
| `spec.credentials[].credentialId` | `string` | yes |  |  |
| `spec.credentials[].displayName` | `string` |  |  |  |
| `spec.credentials[].disabled` | `bool` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project that owns the OAuth client. Can be a literal project
ID or a reference to a GcpProject resource. If omitted, the
provider's default project is used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.location

`string`

The location of the OAuth client. "global" is the documented home
for workforce OAuth clients and the default here. Immutable.

- default: `global`

### spec.oauthClientId

`string`

The client's resource ID (the last segment of its name). Defaults to
metadata.name when left empty. Immutable: changing it destroys and
recreates the client — consumers must re-register.

### spec.displayName

`string`

Human-readable name shown in consoles and consent surfaces.

### spec.description

`string`

What this client is for — the operator-facing record.

### spec.disabled

`bool`

When true, the client stops accepting new authorizations without
being deleted — the reversible kill switch.

### spec.clientType

`string`

The client's confidentiality model. Only "CONFIDENTIAL_CLIENT"
(server-side apps that can keep a secret; manage secrets via
credentials below) can be created: GCP's enum also lists
PUBLIC_CLIENT (mobile/SPA), but the service rejects creating one
with 400 "Client type is not supported" (live-verified at the raw
API — no field combination unlocks it). Re-admit PUBLIC_CLIENT here
when GCP ships support. Immutable.

- rule: client_type must be CONFIDENTIAL_CLIENT — GCP rejects PUBLIC_CLIENT creation ("Client type is not supported")

### spec.allowedGrantTypes

`[]string` · required

OAuth grant types the client may use — Google's API accepts exactly
these values (a closed enum in the IAM REST API):
  "AUTHORIZATION_CODE_GRANT" -- the standard code flow
  "REFRESH_TOKEN_GRANT"      -- long-lived sessions via refresh
                                tokens

- rule: {"repeated":{"minItems":"1","items":{"cel":[{"id":"valid_grant_type","message":"allowed_grant_types entries must be AUTHORIZATION_CODE_GRANT or REFRESH_TOKEN_GRANT","expression":"this in ['AUTHORIZATION_CODE_GRANT', 'REFRESH_TOKEN_GRANT']"}]}}}

### spec.allowedScopes

`[]string` · required

OAuth scopes the client may request during flows (e.g.
"https://www.googleapis.com/auth/cloud-platform", "openid",
"email", "profile", "groups").

- rule: {"repeated":{"minItems":"1"}}

### spec.allowedRedirectUris

`[]string | valueFrom` · required

Redirect URIs allowed when authorization completes. Each entry is a
literal URL or a reference to another resource's URL output (e.g. a
GcpCloudRun service's url) — referencing kills the drift between the
deployed app's address and the client's registration.

- references: GcpCloudRun (`status.outputs.url`)
- rule: {"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpCloudRun, name: <that resource's name>, fieldPath: status.outputs.url}} -- a bare string does not parse

### spec.credentials

`[]GcpIamOauthClientCredential`

Managed client secrets (CONFIDENTIAL_CLIENT only). Each entry
creates one credential whose secret value GCP generates server-side
— read it from this resource's client_secret output. Rotation story:
add a second credential, cut consumers over, remove the first.

### spec.credentials[].credentialId

`string` · required

The credential's resource ID (the last segment of its name), e.g.
"primary". Immutable.

- rule: {"required":true}

### spec.credentials[].displayName

`string`

Human-readable name recording what consumes this credential.

### spec.credentials[].disabled

`bool`

When true, the credential cannot be used to authenticate. GCP
requires a credential to be DISABLED before it can be deleted —
removing an entry from this list while it is still enabled fails at
the API; disable it in one apply, remove it in the next.

### spec.deletionPolicy

`string`

Deletion policy — one switch governs the client AND its credentials:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the client and its credentials are deleted; the
               client ID remains reserved briefly by GCP
  "PREVENT" -- destroy FAILS; protects a client live apps depend on
  "ABANDON" -- everything is removed from management but keeps
               working in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpIamOauthClient, name: <resource-name>, fieldPath: status.outputs.<output>}`. A sensitive output is a secret the resource generates: on Planton it is kept in the organization's secret store and the output holds a `$secret/` reference, so feed it only to a sensitive field.

| Output | Type | Description |
|---|---|---|
| `status.outputs.client_id` | `string` | The system-generated OAuth client ID applications present in OAuth flows (distinct from the user-chosen resource ID). |
| `status.outputs.client_name` | `string` | The client's full resource name: projects/{project}/locations/{location}/oauthClients/{id}. |
| `status.outputs.state` | `string` | The client's lifecycle state (e.g. "ACTIVE", "DELETED" during the 30-day soft-delete window). |
| `status.outputs.client_secret` | `string` (sensitive) | The system-generated secret of the FIRST credential in spec.credentials (empty when no credentials are defined). The single-credential case is the operating norm — rotation adds a second credential and swaps consumers over, at which point the remaining credential is again the first. A live, long-lived credential — the engines mark it secret in state; feed it to consumers via valueFrom (e.g. into a GcpSecretManagerSecret initial_version), never by copy-paste. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.allowedRedirectUris` | GcpCloudRun | `status.outputs.url` |

## See Also

- [Overview](../README.md)

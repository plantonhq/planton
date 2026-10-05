# GCP Identity Platform Config

Configures a project's Identity Platform — the sign-in methods (email/password, phone, anonymous), authorized domains, multi-factor authentication, blocking functions, SMS-region policy, quotas, and the identity providers (Google, Facebook, OIDC, SAML) end users authenticate with. One manifest takes a project from nothing to a working sign-in surface; IdP client secrets are handled as managed secrets end to end.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Identity Platform Config** -- the `identity_platform_config` PROJECT SINGLETON carrying sign-in methods, authorized domains, MFA policy, blocking functions, quotas, SMS-region policy, client permissions, request logging, and multi-tenancy settings
- **Default supported IdP configs** -- one `default_supported_idp_config` per `defaultSupportedIdps` entry (Google, Facebook, Apple, ...)
- **OIDC IdP configs** -- one `oauth_idp_config` per `oauthIdpConfigs` entry (custom OIDC providers)
- **Inbound SAML configs** -- one `inbound_saml_config` per `inboundSamlConfigs` entry (enterprise SSO)
- **Identity Toolkit API enablement** -- `identitytoolkit.googleapis.com` enabled in the target project (never disabled on destroy)

## Before You Deploy

### One-Way Initialization — Read This First

- **The first deploy PERMANENTLY initializes Identity Platform on the project.** The project must have billing enabled, and GCP provides no de-initialize. Destroying this resource ABANDONS the configuration in place — nothing is deleted, and the project keeps Identity Platform enabled forever. Choose the project deliberately.
- **The config resource carries no deletion policy** — it is undeletable by construction. `spec.deletionPolicy` governs ONLY the composed IdP configs (the `defaultSupportedIdps` / `oauthIdpConfigs` / `inboundSamlConfigs` entries).

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project with BILLING enabled** — initialization fails without it. Provide the project ID directly or reference a GcpProject Infra Component via ValueFromRef.
- **IAM**: the deploying identity needs `roles/identityplatform.admin` or broader.

### Identity Provider Credentials

- **Each IdP's `clientSecret` comes from that provider's own developer console** (Google Cloud Console for `google.com`, Meta for `facebook.com`, your OIDC provider's registration page, ...). The platform handles these as managed secrets end to end — stored as references, resolved just-in-time at deploy, never plaintext at rest.
- **Classic consent-screen OAuth clients have NO programmatic creation path anywhere** — Google shut the IAP OAuth Admin API in March 2026. Obtaining the client ID/secret pair is a documented console step, not something any resource can automate.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpIdentityPlatformConfig
metadata:
  name: app-auth
spec:
  signIn:
    email:
      enabled: true
      passwordRequired: true
```

```shell
planton apply -f config.yaml
```

## Configuration Reference

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider default | GCP project (must have billing). Can reference a GcpProject resource. |
| `signIn` | `message` | — | Email/phone/anonymous sign-in arms plus `allowDuplicateEmails`. Each arm you set is sent explicitly (including `enabled: false`); arms you omit stay unmanaged. |
| `authorizedDomains` | `list<string>` | `[]` | Domains authorized for OAuth redirects and hosted sign-in flows. localhost and the Firebase subdomains are authorized by default when empty. |
| `mfa` | `message` | — | MFA policy: `state` (`DISABLED`/`ENABLED`/`MANDATORY`), `enabledProviders` (`PHONE_SMS` only), and TOTP `providerConfigs`. |
| `blockingFunctions` | `message` | — | Cloud Functions invoked synchronously during sign-up/sign-in (`beforeCreate`/`beforeSignIn`) plus token-forwarding switches. |
| `signUpQuota` | `message` | — | Temporary sign-up ceiling — `quota` (1-1000), `quotaDuration`, and `startTime` are set together or not at all. |
| `smsRegionConfig` | `message` | — | Which regions may receive SMS — exactly one of `allowByDefault` (deny list) or `allowlistOnly` (allow list). The toll-fraud control. |
| `clientPermissions` | `message` | — | `disabledUserSignup` / `disabledUserDeletion` — restrict what client apps can do through the API directly. |
| `requestLoggingEnabled` | `bool` | unmanaged | Whether sign-in/sign-up requests reach Cloud Logging. Sent explicitly when set (true or false). |
| `multiTenant` | `message` | — | `allowTenants` must be true before any GcpIdentityPlatformTenant exists in this project; `defaultTenantLocation` sets the tenant-data location. |
| `autodeleteAnonymousUsers` | `bool` | `false` | Delete anonymous users automatically after ~30 days of inactivity. |
| `defaultSupportedIdps` | `list` | `[]` | Well-known providers (`idpId` such as `google.com`), each with the console-issued `clientId`/`clientSecret`. |
| `oauthIdpConfigs` | `list` | `[]` | Custom OIDC providers — `name` must start with `oidc.`, plus `issuer`, `clientId`, optional `clientSecret` and `responseType`. |
| `inboundSamlConfigs` | `list` | `[]` | Enterprise SAML providers — `name` matching `saml.<slug>`, `displayName`, `idpConfig`, optional `spConfig` (https `callbackUri`). |
| `deletionPolicy` | `string` | `DELETE` | Governs ONLY the composed IdP configs: `DELETE`, `PREVENT` (refuse), or `ABANDON`. The config singleton itself is always abandoned. |

### Validation Rules

- **SMS region policy**: exactly one of `allowByDefault` / `allowlistOnly`.
- **Sign-up quota**: `quota`, `quotaDuration`, and `startTime` all set or all empty; quota 1-1000.
- **MFA**: `state` in `DISABLED`/`ENABLED`/`MANDATORY`; `enabledProviders` entries only `PHONE_SMS`; TOTP `adjacentIntervals` 0-10.
- **Blocking functions**: `eventType` in `beforeCreate`/`beforeSignIn`; `functionUri` required.
- **IdP naming**: `idpId` in the ten canonical values; OIDC names start `oidc.`; SAML names match `saml.<lowercase-start slug>`; SAML `callbackUri` must be `https://`.

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `config_name` | `string` | `projects/{project}/config` — the singleton's resource name |
| `api_key` | `string` | The auto-provisioned API key client apps initialize the Identity Platform / Firebase Auth SDK with — a live credential; restrict it by domain/app in the console |
| `firebase_subdomain` | `string` | The project's Firebase subdomain (e.g. `my-project` for `my-project.firebaseapp.com`) — the default hosted sign-in domain |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Initialization is one-way AND once-only** — the first deploy permanently enables Identity Platform on the project (billing required); there is no de-initialize, and destroy abandons the config in place. Every setting remains freely updatable after initialization.
- **Already-initialized projects need `adoptExisting: true`** — GCP rejects a second initialization ("Identity Platform has already been enabled for this project"), so any project where it was ever enabled (console, Firebase Auth, a prior deploy of this kind) deploys with the adoption switch: the module imports the config singleton and applies the spec as an update.
- **`deletionPolicy` never touches the singleton** — it applies only to the composed IdP configs. `PREVENT` is the right value once real users sign in through the listed providers.
- **Sign-in `enabled` flags are sent explicitly** — setting an arm with `enabled: false` actively disables that method; omitting the arm leaves it unmanaged.
- **IdP client secrets are console artifacts** — no API creates consent-screen OAuth clients (Google closed the last programmatic path in March 2026). Budget a console step per provider.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Kinds

- [GcpIdentityPlatformTenant](/docs/catalog/gcp/gcpidentityplatformtenant) — isolated per-customer user pools; requires `multiTenant.allowTenants: true` here first
- [GcpProject](/docs/catalog/gcp/gcpproject) — provides the GCP project (with billing) that gets initialized
- [GcpCloudFunction](/docs/catalog/gcp/gcpcloudfunction) — the blocking-function endpoints referenced by `functionUri`

## Additional Resources

- [Identity Platform Documentation](https://cloud.google.com/identity-platform/docs)
- [Identity Toolkit Config API Reference](https://cloud.google.com/identity-platform/docs/reference/rest/v2/projects/updateConfig)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

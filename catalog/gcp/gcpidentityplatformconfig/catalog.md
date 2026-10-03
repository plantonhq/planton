# GCP Identity Platform Config

Configures a project's Identity Platform — the sign-in methods (email/password, phone, anonymous), authorized domains, multi-factor authentication, blocking functions, SMS-region policy, quotas, and the identity providers (Google, Facebook, OIDC, SAML) end users authenticate with. One manifest takes a project from nothing to a working sign-in surface; IdP client secrets are handled as managed secrets end to end.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Identity Platform Config** -- the `identity_platform_config` PROJECT SINGLETON with sign-in methods, MFA, blocking functions, quotas, and multi-tenancy settings
- **Identity provider configs** -- one composed resource per entry in `defaultSupportedIdps` (Google, Facebook, ...), `oauthIdpConfigs` (custom OIDC), and `inboundSamlConfigs` (enterprise SSO)
- **Identity Toolkit API enablement** -- `identitytoolkit.googleapis.com` enabled in the target project (never disabled on destroy)

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **IdP credentials** -- each provider's `clientId`/`clientSecret` comes from that provider's own developer console — consent-screen OAuth clients have no programmatic creation path. Secrets are supplied as managed secrets.

### GCP Project

- **The first deploy PERMANENTLY initializes Identity Platform on the project** (billing required). GCP has no de-initialize — destroy abandons the config in place. Choose the project deliberately.
- **Already-initialized projects deploy with `adoptExisting: true`** — GCP rejects a second initialization outright, so the adoption switch imports the project's config singleton and applies your spec as an update (a re-deploy after destroy needs it too).
- **IAM**: the deploying identity needs `roles/identityplatform.admin` or broader.

## Deploy

### Console

Open the deployment store, find **GCP Identity Platform Config**, and click **Deploy**. The creation wizard walks you through the target project, sign-in methods, authorized domains, MFA, and the identity-provider lists. Start from the **Email + Password** preset in the [Presets](#presets) tab for the most common shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpIdentityPlatformConfig
metadata:
  name: app-auth
  org: acme-corp
  env: prod
spec:
  signIn:
    email:
      enabled: true
      passwordRequired: true
  authorizedDomains:
    - app.example.com
```

```shell
planton apply -f config.yaml
```

This initializes Identity Platform on the project and enables email/password sign-in; the `api_key` output is what client apps initialize the sign-in SDK with. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, the config references its project and blocking functions via ValueFromRef:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: app-project
      fieldPath: status.outputs.project_id
  blockingFunctions:
    triggers:
      - eventType: beforeSignIn
        functionUri:
          valueFrom:
            kind: GcpCloudFunction
            name: auth-gatekeeper
            fieldPath: status.outputs.function_url
```

The InfraPipeline deploys the project and function first, then initializes Identity Platform with the resolved values.

## Key Configuration

These are the most important decisions when configuring a project's Identity Platform. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Sign-in arms** -- each of `email`, `phoneNumber`, and `anonymous` you set is sent explicitly, including `enabled: false` — a disable in the manifest actively disables the method. Arms you omit stay unmanaged.

**Identity providers** -- three lists compose the project's IdP surface: `defaultSupportedIdps` for well-known providers (by `idpId`, e.g. `google.com`), `oauthIdpConfigs` for custom OIDC (`oidc.`-prefixed names), and `inboundSamlConfigs` for enterprise SAML (`saml.`-prefixed names). Client secrets are managed secrets end to end.

**Multi-tenancy gate** -- `multiTenant.allowTenants: true` is the prerequisite for every GcpIdentityPlatformTenant in the project; tenants cannot exist without it.

**SMS region policy** -- exactly one of `allowByDefault` (deny list) or `allowlistOnly` (allow list); the allow list is the tighter toll-fraud posture.

**Deletion policy** -- governs ONLY the composed IdP configs. The config singleton itself cannot be deleted; destroy always abandons it in place.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** (optional) | `projectId` | `status.outputs.project_id` |
| **GcpCloudFunction** (optional) | `blockingFunctions.triggers[].functionUri` | `status.outputs.function_url` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `config_name` | `projects/{project}/config` | Referencing the singleton in tooling |
| `api_key` | The client SDK bootstrap credential | Initializing the Identity Platform / Firebase Auth SDK in apps (restrict by domain in console) |
| `firebase_subdomain` | Default hosted sign-in domain | Redirect/callback configuration in IdP consoles |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Email + password** -- the minimal working sign-in surface; no external credentials. Start from the **Email + Password** preset.

**Social sign-in** -- add Google (or another well-known provider) with the OAuth client from that provider's console. Start from the **Google Sign-in** preset.

**Enterprise SSO** -- inbound SAML from Okta/Azure AD/etc., typically with MFA enabled and authorized domains pinned. Start from the **Enterprise SAML** preset.

## Works With

- [**GCP Identity Platform Tenant**](/infra-catalog/gcp-identity-platform-tenant) -- isolated per-customer user pools, gated on `multiTenant.allowTenants` here
- [**GCP Project**](/infra-catalog/gcp-project) -- provides the billing-enabled project that gets initialized
- [**GCP Cloud Function**](/infra-catalog/gcp-cloud-function) -- the blocking-function endpoints invoked during sign-up/sign-in

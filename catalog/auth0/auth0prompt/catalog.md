# Auth0 Prompt

Sets how an existing Auth0 tenant's login flow behaves: Universal Login or the Classic pages, identifier-first login, and whether the device's own authenticator is offered as the first factor. One Infra Component per tenant.

## What Gets Created

Nothing new: a tenant has one set of prompt settings, and Auth0 has no way to create or delete it. When you deploy this Infra Component, the IaC module sets the prompt settings of the tenant your Auth0 connection's credential belongs to:

- **Login experience** -- `new` (Universal Login) or `classic` (the legacy Lock-based pages)
- **Identifier first** -- the email or username on a first screen, the password on a second
- **Device biometrics first** -- Touch ID, Face ID or Windows Hello offered before the password

## Before You Deploy

### Planton Setup

- **Auth0 Provider Connection** -- an active connection in the Connect module with the tenant's domain, client ID, and client secret. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Auth0 Account

- **The connection's Machine-to-Machine application** must hold `read:prompts` and `update:prompts` on the tenant's Management API (Auth0 dashboard: Applications, APIs, Auth0 Management API, Machine To Machine Applications).
- **WebAuthn with device biometrics enabled** as a multi-factor method (Auth0 dashboard: Security, Multi-factor Auth), if you offer device biometrics first. Its availability depends on your plan.

## Deploy

### Console

Open the deployment store, find **Auth0 Prompt**, and click **Deploy**. Start from the **Identifier-First Login** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0Prompt
metadata:
  name: prompt
  org: acme-corp
  env: prod
spec:
  universalLoginExperience: new
  identifierFirst: true
```

```shell
planton apply -f auth0-prompt.yaml
```

The tenant's login page asks for the email first, then the password or the company's identity provider. An Infra Job tracks the change in real time.

## Key Configuration

These are the decisions that matter. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Unset is unmanaged** -- Every field is optional; a field you leave out is never sent, and the tenant keeps its current value. Set at least one.

**Universal Login first** -- Branding, themes, custom text and partials style only the `new` experience. On `classic` they apply but nobody sees them.

**Identifier first routes enterprise users** -- With `identifierFirst`, the email's domain decides the next step: an enterprise connection's identity provider, or the password screen.

**Device biometrics need their factor** -- `webauthnPlatformFirstFactor: true` is accepted only when WebAuthn with device biometrics is enabled as a multi-factor method; pair it with `identifierFirst: true`.

**Destroy leaves the settings in place** -- Auth0 has no delete for the prompt settings, so destroying this Infra Component stops managing them and keeps their last-applied values.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies. The tenant is the one the Auth0 connection's credential belongs to.

### What This Kind Provides

After provisioning, `status.outputs` contains the settings as the tenant carries them:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `universal_login_experience` | The login experience the tenant runs | Audits, a check before custom text or partials |
| `identifier_first` | Whether the flow asks for the identifier first | Audits |
| `webauthn_platform_first_factor` | Whether device biometrics come first | Audits |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Identifier-first login** -- Universal Login asking for the email first, so enterprise users reach their identity provider. Start from the **Identifier-First Login** preset.

**Device biometrics first** -- The same flow, with the device's own authenticator offered before the password. Start from the **Device Biometrics First** preset.

## Works With

- [**Auth0 Connection**](/infra-catalog/auth0-connection) -- the enterprise connections identifier-first login routes to.
- [**Auth0 Tenant Settings**](/infra-catalog/auth0-tenant-settings) -- the name and logo the login page shows.
- [**Auth0 Prompt Custom Text**](/infra-catalog/auth0-prompt-custom-text) -- the words each prompt shows, on the `new` experience.

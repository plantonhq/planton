# Auth0Prompt

Manages how an existing [Auth0 tenant's](https://auth0.com/docs/authenticate/login/auth0-universal-login) login flow behaves: whether it runs the Universal Login experience or the Classic pages, whether it asks for the identifier first and the password on a second screen, and whether the device's own authenticator is offered as the first factor.

## When to Use

- **Enterprise sign-in by email**: with identifier-first login, a person whose email domain belongs to an enterprise connection is routed to their company's identity provider instead of being asked for a password.
- **Faster sign-in on known devices**: offer Touch ID, Face ID or Windows Hello as the first factor to people who have enrolled their device.
- **Universal Login everywhere**: the branding, theme, custom text and partials of the other Auth0 kinds reach only the `new` experience.
- **Infrastructure as code**: the login flow is version-controlled beside the connections and clients it serves.

## Quick Start

```yaml
apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0Prompt
metadata:
  name: prompt
  org: acme-corp
  env: production
spec:
  universalLoginExperience: new
  identifierFirst: true
```

## Fields

| Field | Description |
|---|---|
| `universalLoginExperience` | `new` (Universal Login) or `classic` (the legacy Lock-based pages) |
| `identifierFirst` | Ask for the email or username first and the password on a second screen |
| `webauthnPlatformFirstFactor` | Offer the device's own authenticator as the first factor |

Every field is optional; at least one is set.

## Key Behaviors

- **The tenant is the credential's**: the settings managed are those of the tenant the provider connection's credential belongs to. A tenant has one set, so one resource per tenant.
- **Unset is unmanaged**: a field left out is never sent, and the tenant keeps whatever value it already carries.
- **Nothing is ever deleted**: Auth0 has no delete for the prompt settings. Destroying this resource forgets them and leaves their last-applied values in place. To return a setting to a specific value, set that value before removing the field.
- **Device biometrics need their factor**: Auth0 accepts `webauthnPlatformFirstFactor: true` only when WebAuthn with device biometrics is enabled as a multi-factor method on the tenant. Auth0 offers the factor after the identifier screen, so pair it with `identifierFirst: true` (Auth0's "Identifier First + Biometrics" profile).
- **Classic reaches nothing else**: on `classic`, the Auth0 Branding, Auth0 Prompt Custom Text and Auth0 Prompt Screen Partials kinds apply but people never see them.
- **Plans**: the experience and identifier-first login are available on every plan. The availability of WebAuthn with device biometrics depends on the plan.
- **Permissions**: the credential needs `read:prompts` and `update:prompts` on the tenant's Management API (`iac/permissions.yaml`).

## Outputs

| Output | Description |
|---|---|
| `universal_login_experience` | The login experience the tenant runs |
| `identifier_first` | Whether the login flow asks for the identifier first |
| `webauthn_platform_first_factor` | Whether the device's own authenticator is offered as the first factor |

The outputs are the settings as the tenant carries them after the apply, managed or not.

## Auth0 Documentation

- [Identifier First authentication](https://auth0.com/docs/authenticate/login/auth0-universal-login/identifier-first)
- [WebAuthn with device biometrics](https://auth0.com/docs/authenticate/login/auth0-universal-login/passwordless-login/webauthn-device-biometrics)
- [Universal Login vs Classic Login](https://auth0.com/docs/authenticate/login/auth0-universal-login/universal-login-vs-classic-login)
- [Terraform auth0_prompt](https://registry.terraform.io/providers/auth0/auth0/latest/docs/resources/prompt)
- [Pulumi auth0.Prompt](https://www.pulumi.com/registry/packages/auth0/api-docs/prompt/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

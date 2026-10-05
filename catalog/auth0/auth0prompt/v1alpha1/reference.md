# Auth0Prompt

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `auth0.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

Auth0PromptSpec manages how the login flow of the Auth0 tenant the provider
connection's credential belongs to behaves.

Unset fields are NOT MANAGED: the module never sends them, and the tenant keeps
whatever it carries. Auth0 has no delete for the prompt settings, so destroy
leaves the last-applied values in place.

The credential needs read:prompts and update:prompts on the tenant's
Management API (iac/permissions.yaml).

https://auth0.com/docs/authenticate/login/auth0-universal-login/identifier-first
https://auth0.com/docs/authenticate/passwordless/passwordless-with-universal-login/webauthn-device-biometrics
https://registry.terraform.io/providers/auth0/auth0/latest/docs/resources/prompt

## Example

```yaml
# Auth0 Prompt Test Manifest
# This file is used for testing the Auth0Prompt kind.
#
# Applying it CHANGES the login flow of the tenant the credential belongs to:
# run it only against a test tenant nobody signs in to.
#
# Prerequisites:
# 1. Set the following environment variables:
#    - AUTH0_DOMAIN: the test tenant's domain (e.g., your-test-tenant.auth0.com)
#    - AUTH0_CLIENT_ID: M2M application client ID
#    - AUTH0_CLIENT_SECRET: M2M application client secret
#
# 2. The M2M application must have these scopes:
#    - read:prompts
#    - update:prompts

apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0Prompt
metadata:
  name: test-prompt
  org: test-org
  env: development
  labels:
    purpose: testing
spec:
  # Universal Login, the experience every branding, theme, custom text and
  # partial styles
  universalLoginExperience: new

  # The email or username first, the password on a second screen, so people
  # of an enterprise connection's email domain are routed to their identity
  # provider
  identifierFirst: true
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.universalLoginExperience` | `string` |  |  |  |
| `spec.identifierFirst` | `bool` |  |  |  |
| `spec.webauthnPlatformFirstFactor` | `bool` |  |  |  |

## Field Details

### spec.universalLoginExperience

`string`

universal_login_experience chooses the login experience: "new" (Universal
Login, which every branding, theme, custom text and partial styles) or
"classic" (the legacy Lock-based pages, which none of them reach).

- rule: {"string":{"in":["","new","classic"]}}

### spec.identifierFirst

`bool` · optional (explicit presence)

identifier_first asks for the email or username on a first screen and the
password on a second, so a person whose email domain belongs to an
enterprise connection is routed to that identity provider instead of being
asked for a password.

### spec.webauthnPlatformFirstFactor

`bool` · optional (explicit presence)

webauthn_platform_first_factor offers the device's own authenticator
(Touch ID, Face ID, Windows Hello) as the first factor, after the
identifier screen ("Identifier First + Biometrics"), so it needs
identifier_first: true. The tenant needs WebAuthn with Device Biometrics
enabled as a multi-factor factor first.

## Validation Rules

- `spec.biometrics_first_needs_identifier_first`: webauthn_platform_first_factor needs identifier_first: true -- Auth0 offers the device's authenticator only after the identifier screen has named the account
- `spec.at_least_one_setting`: configure at least one of universal_login_experience, identifier_first or webauthn_platform_first_factor -- an Auth0Prompt resource that manages nothing would deploy nothing

## Outputs

Reference an output from another manifest as `valueFrom: {kind: Auth0Prompt, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.universal_login_experience` | `string` | universal_login_experience is the login experience the tenant runs. |
| `status.outputs.identifier_first` | `bool` | identifier_first is whether the login flow asks for the identifier first. |
| `status.outputs.webauthn_platform_first_factor` | `bool` | webauthn_platform_first_factor is whether the device's authenticator is offered as the first factor. |

## See Also

- [Overview](../README.md)

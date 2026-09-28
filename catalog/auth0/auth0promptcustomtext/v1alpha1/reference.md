# Auth0PromptCustomText

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `auth0.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

Auth0PromptCustomTextSpec manages the words one Universal Login prompt shows in
one language, on the tenant the provider connection's credential belongs to.

A prompt is one step of the login flow (login, signup, reset-password, the
multi-factor challenges, ...), and each prompt has one or more screens. Auth0
stores one custom text per prompt and language, and this resource owns all of
it: every screen and key the spec leaves out shows Auth0's default words, and
a key Auth0 later adds shows its default until you declare it. The screens and
keys each prompt offers are listed in Auth0's customization reference; a key
may use Universal Login's variables (for example ${clientName} or
${companyName}).

Destroying the resource returns the prompt to Auth0's default words in that
language. Custom text needs the Universal Login experience "new"
(Auth0Prompt).

The credential needs read:prompts and update:prompts on the tenant's
Management API (iac/permissions.yaml).

https://auth0.com/docs/customize/login-pages/universal-login/customize-text-elements
https://registry.terraform.io/providers/auth0/auth0/latest/docs/resources/prompt_custom_text

## Example

```yaml
# Auth0 Prompt Custom Text Test Manifest
# This file is used for testing the Auth0PromptCustomText component.
#
# Applying it REPLACES the English words of the tenant's login prompt: run it
# only against a test tenant nobody signs in to.
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
#
# 3. The tenant runs the Universal Login experience "new" (Auth0Prompt).

apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0PromptCustomText
metadata:
  name: test-login-en
  org: test-org
  env: development
  labels:
    purpose: testing
spec:
  # The step of the login flow, and the language the words are shown in
  prompt: login
  language: en

  # Each screen of the prompt, and the words of each text key on it. A key
  # left out shows Auth0's default words.
  screens:
    login:
      texts:
        title: Welcome back
        description: Log in to Acme to continue to ${clientName}.
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.prompt` | `string` | yes |  |  |
| `spec.language` | `string` | yes |  |  |
| `spec.screens` | `map<string, Auth0PromptScreenText>` | yes |  |  |
| `spec.screens.*.texts` | `map<string, string>` | yes |  |  |

## Field Details

### spec.prompt

`string` · required

prompt is the step of the login flow the words belong to. With language it
is the custom text's identity in Auth0: treat both as fixed, and declare a
second resource for another prompt or language (a change in place would
write the new pair and leave the old pair's words as they were).

- rule: {"required":true,"string":{"in":["login","login-id","login-password","login-passwordless","login-email-verification","signup","signup-id","signup-password","phone-identifier-enrollment","phone-identifier-challenge","email-identifier-challenge","reset-password","custom-form","consent","customized-consent","logout","mfa-push","mfa-otp","mfa-voice","mfa-phone","mfa-webauthn","mfa-sms","mfa-email","mfa-recovery-code","mfa","status","device-flow","email-verification","email-otp-challenge","organizations","invitation","common","passkeys","captcha","brute-force-protection","confirmation"]}}

### spec.language

`string` · required

language is the language the words are shown in, as Auth0 names it (for
example "en", "fr-CA", "pt-BR", "zh-TW"). The tenant must list the
language among its enabled languages for Universal Login to show it, and
Auth0 checks the code against its current language list when the resource
is applied. Fixed, like prompt.

- rule: language is a language code as Auth0 names it, such as en, fr-CA, es-419 or zh-TW
- rule: {"required":true}

### spec.screens

`map<string, Auth0PromptScreenText>` · required

screens maps each screen of the prompt (for example "login" or
"login-id") to the words it shows. The screens a prompt has are named in
Auth0's customization reference.

- rule: {"map":{"minPairs":"1"}}

### spec.screens.*.texts

`map<string, string>` · required

texts maps each text key of the screen (for example "title",
"description", "buttonText", "wrong-credentials") to the words it shows.

- rule: {"map":{"minPairs":"1"}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: Auth0PromptCustomText, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.prompt` | `string` | prompt is the prompt the words belong to. |
| `status.outputs.language` | `string` | language is the language the words are shown in. |
| `status.outputs.id` | `string` | id is the custom text's identifier, "<prompt>::<language>" (for example "login::en"), which is also its import id. |

## See Also

- [Overview](../README.md)

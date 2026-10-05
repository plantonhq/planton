# Auth0PromptScreenPartials

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `auth0.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

Auth0PromptScreenPartialsSpec manages the HTML fragments one Universal Login
prompt inserts on its screens, on the tenant the provider connection's
credential belongs to.

Each screen offers named insertion points; a partial is an HTML fragment (it
may use Liquid and read the prompt's context) rendered at one of them. This
resource owns the prompt's whole set: a screen or insertion point the spec
leaves out renders nothing, and destroying the resource removes every partial
of the prompt. Form fields a partial adds are submitted with the form and
reach Actions as custom prompt fields.

Auth0 accepts partials only on a tenant with a custom domain and a page
template: without either it refuses them (403 "requires at least one custom
domain", then 403 "requires a page template"). Deploy an Auth0CustomDomain
first (it need not be verified), then an Auth0Branding that sets
universal_login_template. They render only with the Universal Login
experience "new".

The credential needs read:prompts and update:prompts on the tenant's
Management API (iac/permissions.yaml).

https://auth0.com/docs/customize/login-pages/universal-login/customize-signup-and-login-prompts
https://registry.terraform.io/providers/auth0/auth0/latest/docs/resources/prompt_screen_partials

## Example

```yaml
# Auth0 Prompt Screen Partials Test Manifest
# This file is used for testing the Auth0PromptScreenPartials kind.
#
# Applying it REPLACES every partial of the tenant's signup prompt: run it
# only against a test tenant nobody signs up through.
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
# 3. The tenant has a custom domain (Auth0CustomDomain) and a Universal Login
#    page template (Auth0Branding's universal_login_template): partials render
#    only inside a page template, and a page template needs a custom domain.

apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0PromptScreenPartials
metadata:
  name: test-signup-partials
  org: test-org
  env: development
  labels:
    purpose: testing
spec:
  # The prompt whose screens the partials extend -- the resource's identity
  promptType: signup

  # One entry per screen, each fragment at its named insertion point
  screenPartials:
    - screenName: signup
      insertionPoints:
        formContentEnd: <div class="terms">By signing up you accept the terms.</div>
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.promptType` | `string` | yes |  |  |
| `spec.screenPartials` | `[]Auth0PromptScreenPartial` | yes |  |  |
| `spec.screenPartials[].screenName` | `string` | yes |  |  |
| `spec.screenPartials[].insertionPoints` | `Auth0PromptInsertionPoints` | yes |  |  |
| `spec.screenPartials[].insertionPoints.formContent` | `string` |  |  |  |
| `spec.screenPartials[].insertionPoints.formContentStart` | `string` |  |  |  |
| `spec.screenPartials[].insertionPoints.formContentEnd` | `string` |  |  |  |
| `spec.screenPartials[].insertionPoints.formFooterStart` | `string` |  |  |  |
| `spec.screenPartials[].insertionPoints.formFooterEnd` | `string` |  |  |  |
| `spec.screenPartials[].insertionPoints.secondaryActionsStart` | `string` |  |  |  |
| `spec.screenPartials[].insertionPoints.secondaryActionsEnd` | `string` |  |  |  |

## Field Details

### spec.promptType

`string` · required

prompt_type is the prompt whose screens the partials extend. It is the
resource's identity in Auth0: changing it manages another prompt, so treat
it as fixed and declare a second resource for a second prompt.

- rule: {"required":true,"string":{"in":["login-id","login","login-password","signup","signup-id","signup-password","login-passwordless","customized-consent","passkeys","confirmation"]}}

### spec.screenPartials

`[]Auth0PromptScreenPartial` · required

screen_partials are the partials of each of the prompt's screens, one entry
per screen.

- rule: each screen appears once in screen_partials -- put all of a screen's insertion points in its one entry
- rule: {"repeated":{"minItems":"1"}}

### spec.screenPartials[].screenName

`string` · required

screen_name is the screen the partials render on (for example "login",
"signup-id"), one of the prompt's screens.

- rule: {"required":true}

### spec.screenPartials[].insertionPoints

`Auth0PromptInsertionPoints` · required

insertion_points are the fragments and where they render.

- rule: {"required":true}
- rule: set at least one insertion point -- a screen entry with no fragments renders nothing

### spec.screenPartials[].insertionPoints.formContent

`string`

form_content replaces the form's own fields area -- use only to rebuild the form.

### spec.screenPartials[].insertionPoints.formContentStart

`string`

form_content_start renders at the start of the form, above its fields.

### spec.screenPartials[].insertionPoints.formContentEnd

`string`

form_content_end renders at the end of the form, below its fields and above the submit button.

### spec.screenPartials[].insertionPoints.formFooterStart

`string`

form_footer_start renders at the start of the form's footer.

### spec.screenPartials[].insertionPoints.formFooterEnd

`string`

form_footer_end renders at the end of the form's footer.

### spec.screenPartials[].insertionPoints.secondaryActionsStart

`string`

secondary_actions_start renders before the secondary actions (the links below the form).

### spec.screenPartials[].insertionPoints.secondaryActionsEnd

`string`

secondary_actions_end renders after the secondary actions.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: Auth0PromptScreenPartials, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.prompt_type` | `string` | prompt_type is the prompt whose screens the partials extend. |

## See Also

- [Overview](../README.md)

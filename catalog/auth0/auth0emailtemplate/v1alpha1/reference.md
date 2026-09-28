# Auth0EmailTemplate

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `auth0.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

Auth0EmailTemplateSpec manages one of the emails the Auth0 tenant the provider
connection's credential belongs to sends.

The template sends through the tenant's email provider (Auth0EmailProvider);
without one, Auth0 refuses custom templates. The subject and body are Liquid
templates that read the email's context: {{ user.name }}, {{ user.email }},
{{ url }} (the action link), {{ application.name }}, {{ friendly_name }}
(the tenant's name) and the rest of Auth0's template variables.

Auth0 cannot delete a template: destroying the resource disables it, and the
tenant sends Auth0's default email again.

The credential needs read:email_templates, create:email_templates and
update:email_templates on the tenant's Management API (iac/permissions.yaml).

https://auth0.com/docs/customize/email/email-templates
https://registry.terraform.io/providers/auth0/auth0/latest/docs/resources/email_template

## Example

```yaml
# Auth0 Email Template Test Manifest
# This file is used for testing the Auth0EmailTemplate component.
#
# Prerequisites:
# 1. Set the following environment variables:
#    - AUTH0_DOMAIN: the test tenant's domain (e.g., your-test-tenant.auth0.com)
#    - AUTH0_CLIENT_ID: M2M application client ID
#    - AUTH0_CLIENT_SECRET: M2M application client secret
#
# 2. The M2M application must have these scopes:
#    - read:email_templates
#    - create:email_templates
#    - update:email_templates
#
# 3. The tenant must send through its own email provider (e.g., created via
#    the Auth0EmailProvider component): Auth0 refuses custom templates on the
#    built-in one. The from address must be on a domain that provider may
#    send for.

apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0EmailTemplate
metadata:
  name: test-verify-email
  org: test-org
  env: development
  labels:
    purpose: testing
spec:
  # The email this resource customizes
  template: verify_email

  # The sender, subject and HTML body, all Liquid templates
  from: Acme <no-reply@acme.com>
  subject: Verify your email for Acme
  body: |
    <html><body><p>Hello {{ user.name | escape }}, <a href="{{ url }}">verify your email</a>.</p></body></html>

  # How long the verification link works: one day
  urlLifetimeInSeconds: 86400
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.template` | `string` | yes |  |  |
| `spec.from` | `string` | yes |  |  |
| `spec.subject` | `string` | yes |  |  |
| `spec.body` | `string` | yes |  |  |
| `spec.syntax` | `string` |  | `liquid` |  |
| `spec.resultUrl` | `string` |  |  |  |
| `spec.urlLifetimeInSeconds` | `int32` |  |  |  |
| `spec.enabled` | `bool` |  | `true` |  |
| `spec.includeEmailInRedirect` | `bool` |  |  |  |

## Field Details

### spec.template

`string` · required

template is the email this resource customizes. It is the resource's
identity in Auth0, so treat it as fixed.
- "verify_email" / "verify_email_by_code": confirm an address, by link or
  by code
- "reset_email" / "reset_email_by_code": reset a password, by link or by
  code
- "welcome_email": sent after the first verification
- "blocked_account": sent when brute-force protection blocks an account
- "stolen_credentials": sent when breached-password detection finds a match
- "enrollment_email": invites a person to enroll in multi-factor
- "mfa_oob_code": the multi-factor code sent by email
- "user_invitation": an organization invitation
- "async_approval": an asynchronous authorization request
- "auth_email_by_code": a passwordless sign-in code
- "change_password", "password_reset": legacy reset templates, kept for
  tenants that still send them

- rule: {"required":true,"string":{"in":["verify_email","verify_email_by_code","reset_email","reset_email_by_code","welcome_email","blocked_account","stolen_credentials","enrollment_email","mfa_oob_code","user_invitation","change_password","password_reset","async_approval","auth_email_by_code"]}}

### spec.from

`string` · required

from is the sender of this email ("Acme <no-reply@acme.com>"), overriding
the provider's default_from_address.

- rule: {"required":true}

### spec.subject

`string` · required

subject is the email's subject line, a Liquid template.

- rule: {"required":true}

### spec.body

`string` · required

body is the email's HTML body, a Liquid template.

- rule: {"required":true}

### spec.syntax

`string` · optional (explicit presence)

syntax is the template language of subject and body. "liquid" is the only
language Auth0 renders.

- default: `liquid`
- rule: {"string":{"in":["liquid"]}}

### spec.resultUrl

`string`

result_url is where the person lands after acting on the email's link
(Auth0's "Redirect To"), an http(s) URL or a Liquid template of one (for
example "{{ application.callback_domain }}/welcome"). Unset, Auth0 shows its
own result page. Two limits:
- Auth0 refuses a custom value (403 "Customizations for resultUrl are not
  allowed for non-enterprise tenants") on non-Enterprise tenants created on
  or after May 5, 2026.
- Under Universal Login, reset_email ignores it; where a person lands after
  a reset is chosen by a password-reset-post-challenge Action.

- rule: result_url is an http(s) URL or a Liquid template of one, such as {{ application.callback_domain }}/welcome

### spec.urlLifetimeInSeconds

`int32` · optional (explicit presence)

url_lifetime_in_seconds is how long the email's link works, for the
templates that carry one (verification, password change, blocked account).
Unset, Auth0's default applies: 432,000 seconds (five days).

- rule: {"int32":{"gt":0}}

### spec.enabled

`bool` · optional (explicit presence)

enabled turns this template on. Disabled, the tenant sends Auth0's default
email for it.

- default: `true`

### spec.includeEmailInRedirect

`bool` · optional (explicit presence)

include_email_in_redirect appends the person's email address to result_url
as a query parameter.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: Auth0EmailTemplate, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.template` | `string` | template is the email this resource customizes. |
| `status.outputs.enabled` | `bool` | enabled is whether the template is on. |

## See Also

- [Overview](../README.md)

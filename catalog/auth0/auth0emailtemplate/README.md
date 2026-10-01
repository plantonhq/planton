# Auth0EmailTemplate

Customizes one of the [emails](https://auth0.com/docs/customize/email/email-templates) an Auth0 tenant sends -- the verification, password-reset, welcome, invitation, multi-factor or blocked-account email -- with your sender, subject and Liquid-templated HTML body, where its link leads afterwards, and how long that link lives.

## When to Use

- **Your product's voice**: people receive a verification email that looks and reads like your product, not Auth0's default.
- **Back to your application**: after verifying, people land on your page instead of Auth0's result page.
- **Short-lived links**: a reset link that works for an hour, not five days.
- **Infrastructure as code**: every email the tenant sends is reviewed and versioned beside the connections that trigger it.

## Quick Start

```yaml
apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0EmailTemplate
metadata:
  name: verify-email
  org: acme-corp
  env: production
spec:
  template: verify_email
  from: Acme <no-reply@acme.com>
  subject: Verify your email for Acme
  body: |
    <html><body>
      <p>Confirm that {{ user.email }} is your address.</p>
      <p><a href="{{ url }}">Verify email</a></p>
    </body></html>
  resultUrl: https://app.acme.com/welcome
```

## Fields

| Field | Description |
|---|---|
| `template` | The email customized (required): `verify_email`, `verify_email_by_code`, `reset_email`, `reset_email_by_code`, `welcome_email`, `blocked_account`, `stolen_credentials`, `enrollment_email`, `mfa_oob_code`, `user_invitation`, `async_approval`, `auth_email_by_code`, and the legacy `change_password` and `password_reset`. It is the resource's identity: treat it as fixed |
| `from` | The sender, overriding the email provider's `defaultFromAddress` (required) |
| `subject` | The subject line, a Liquid template (required) |
| `body` | The HTML body, a Liquid template (required) |
| `syntax` | The template language; defaults to `liquid`, the only one Auth0 renders |
| `resultUrl` | Where the person lands after acting on the link; unset, Auth0 shows its own result page |
| `urlLifetimeInSeconds` | How long the link works; unset, Auth0's default (five days) |
| `enabled` | The template is on; defaults to `true`. Off, the tenant sends Auth0's default email |
| `includeEmailInRedirect` | Appends the person's email address to `resultUrl` |

The subject and body read the email's context: `{{ user.email }}`, `{{ user.name }}`, `{{ url }}` (the action link), `{{ code }}` (in code templates), `{{ application.name }}`, `{{ friendly_name }}`, `{{ support_email }}`, and the rest of [Auth0's template variables](https://auth0.com/docs/customize/email/email-templates/supported-liquid-syntax). Escape user-supplied values (`{{ user.name | escape }}`).

## Key Behaviors

- **One per email**: a tenant carries one Auth0EmailTemplate resource for each email it customizes. Applying a template that already exists takes it over and rewrites it.
- **Needs an email provider**: Auth0 refuses custom templates on a tenant that sends through its built-in provider -- apply an `Auth0EmailProvider` first (the registry declares it as this kind's prerequisite).
- **Destroy disables**: Auth0 cannot delete a template. Destroying this resource disables it (a PATCH of `enabled: false`), and the tenant sends Auth0's default email again; the template's last content stays stored in Auth0.
- **`resultUrl` is plan-gated**: Auth0 refuses a custom `resultUrl` with 403 on non-enterprise tenants created on or after May 5, 2026; older tenants are exempt. Universal Login also ignores it after a password reset.
- **Permissions**: the credential needs `read:email_templates`, `create:email_templates` and `update:email_templates` on the tenant's Management API (`iac/permissions.yaml`). No delete scope exists or is needed.

## What the Spec Does Not Carry

Every argument of `auth0_email_template` is reachable from the spec by its own name.

## Outputs

| Output | Description |
|---|---|
| `template` | The email this resource customizes |
| `enabled` | Whether the template is on |

## Auth0 Documentation

- [Email templates](https://auth0.com/docs/customize/email/email-templates)
- [Supported Liquid syntax](https://auth0.com/docs/customize/email/email-templates/supported-liquid-syntax)
- [Terraform auth0_email_template](https://registry.terraform.io/providers/auth0/auth0/latest/docs/resources/email_template)
- [Pulumi auth0.EmailTemplate](https://www.pulumi.com/registry/packages/auth0/api-docs/emailtemplate/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

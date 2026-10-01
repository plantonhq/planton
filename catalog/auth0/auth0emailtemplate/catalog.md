# Auth0 Email Template

Customizes one of the emails an Auth0 tenant sends -- verification, password reset, welcome, invitations, multi-factor codes -- with your sender, subject and HTML body, where its link leads afterwards, and how long that link lives. One Cloud Resource per email.

## What Gets Created

When you deploy this Cloud Resource, the IaC module sets one email template of the tenant your Auth0 connection's credential belongs to:

- **Sender and subject** -- who the email comes from and what it says in the inbox
- **HTML body** -- a Liquid template reading the person, the application, the tenant, and the action link
- **Link behavior** -- where the person lands afterwards, and how long the link works

An existing template for the same email is taken over.

## Before You Deploy

### Planton Setup

- **Auth0 Provider Connection** -- an active connection in the Connect module with the tenant's domain, client ID, and client secret. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Auth0 Email Provider** -- the tenant must send through its own email provider; Auth0 refuses custom templates on its built-in one.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Auth0 Account

- **The connection's Machine-to-Machine application** must hold `read:email_templates`, `create:email_templates` and `update:email_templates` on the tenant's Management API (Auth0 dashboard: Applications, APIs, Auth0 Management API, Machine To Machine Applications).
- **An enterprise plan or an older tenant** if you set `resultUrl`: Auth0 refuses it on non-enterprise tenants created on or after May 5, 2026.

## Deploy

### Console

Open the deployment store, find **Auth0 Email Template**, and click **Deploy**. Start from the **Verify Email** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0EmailTemplate
metadata:
  name: verify-email
  org: acme-corp
  env: prod
spec:
  template: verify_email
  from: Acme <no-reply@acme.com>
  subject: Verify your email for Acme
  body: |
    <html><body><p><a href="{{ url }}">Verify {{ user.email }}</a></p></body></html>
  urlLifetimeInSeconds: 86400
```

```shell
planton apply -f auth0-email-template.yaml
```

The tenant's next verification email is yours. A Stack Job tracks the change in real time.

## Key Configuration

These are the decisions that matter. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Which email** -- `template` names it (`verify_email`, `reset_email`, `welcome_email`, ...), and it is the resource's identity: one Cloud Resource per email, never renamed.

**A sender the provider may send for** -- `from` must be on a domain your email provider has verified, or the email bounces or lands in spam. It may not contain `@auth0.com`.

**Liquid, escaped** -- The subject and body are Liquid templates. `{{ url }}` is the action link; escape user-supplied values such as `{{ user.name | escape }}`.

**Short links for resets** -- `urlLifetimeInSeconds` defaults to five days; a password-reset link acts for the person, so set an hour.

**Destroy disables** -- Auth0 cannot delete a template: destroying this Cloud Resource turns it off, and the tenant sends Auth0's default email again.

## Outputs and Dependencies

### What This Component Consumes

This component has no foreign key dependencies. It sends through the tenant's Auth0 Email Provider, which must exist first.

### What This Component Provides

After provisioning, `status.outputs` contains:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `template` | The email this resource customizes | Audits |
| `enabled` | Whether the template is on | Audits |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Verify email** -- The verification email in your voice, landing on your site afterwards. Start from the **Verify Email** preset.

**Reset password** -- A one-hour reset link in the same layout. Start from the **Reset Password** preset.

**Welcome email** -- The first email after verification, pointing at your product. Start from the **Welcome Email** preset.

## Works With

- [**Auth0 Email Provider**](/cloud-catalog/auth0-email-provider) -- the service these emails are sent through; required.
- [**Auth0 Tenant Settings**](/cloud-catalog/auth0-tenant-settings) -- the friendly name and support address the templates can show.
- [**Auth0 Connection**](/cloud-catalog/auth0-connection) -- the database connection whose sign-ups and resets send these emails.

# Auth0 Email Provider

Sets the service an Auth0 tenant sends its emails through -- verification, password reset, invitations and multi-factor codes -- so they leave from your own domain instead of Auth0's built-in test provider. One Infra Component per tenant.

## What Gets Created

When you deploy this Infra Component, the IaC module sets the email provider of the tenant your Auth0 connection's credential belongs to:

- **The sending service** -- SMTP, Amazon SES, SendGrid, SparkPost, Mailgun, Mandrill, Azure Communication Services, Microsoft 365, or a custom Action
- **Its credentials** -- sent to Auth0 once, never read back
- **The default sender** -- the address every email comes from unless a template names its own

A tenant that already has a provider is taken over.

## Before You Deploy

### Planton Setup

- **Auth0 Provider Connection** -- an active connection in the Connect module with the tenant's domain, client ID, and client secret. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Organization secrets** -- one per credential the service needs (for example `resend-api-key`), referenced from the manifest as `$secret/<name>`.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Auth0 Account

- **The connection's Machine-to-Machine application** must hold `read:email_provider`, `create:email_provider`, `update:email_provider` and `delete:email_provider` on the tenant's Management API (Auth0 dashboard: Applications, APIs, Auth0 Management API, Machine To Machine Applications).

### Sending Service

- **A verified sending domain** at the service, with its SPF and DKIM records published, and a credential allowed to send.

## Deploy

### Console

Open the deployment store, find **Auth0 Email Provider**, and click **Deploy**. Start from the **Resend Over SMTP** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0EmailProvider
metadata:
  name: email-provider
  org: acme-corp
  env: prod
spec:
  defaultFromAddress: Acme <no-reply@acme.com>
  sendgrid:
    apiKey: $secret/sendgrid-api-key
```

```shell
planton apply -f auth0-email-provider.yaml
```

The tenant's next verification email leaves from `no-reply@acme.com` through SendGrid. An Infra Job tracks the change in real time.

## Key Configuration

These are the decisions that matter. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**One service, only its settings** -- Set exactly one of `smtp`, `ses`, `sendgrid`, `sparkpost`, `mailgun`, `mandrill`, `azureCs`, `ms365` or `custom`; each carries only what that service uses.

**Resend through SMTP** -- Auth0's API accepts Resend by name, but the Terraform provider this kind deploys through (v1.58.0) does not, so use the `smtp` arm with host `smtp.resend.com`, user `resend`, port 587, and a Resend API key as the password.

**Credentials by reference** -- Every credential is sensitive. Reference an organization secret; Auth0 stores the value encrypted and never returns it.

**A sender your service may send for** -- `defaultFromAddress` must be on a domain verified at the service, or emails bounce or land in spam.

**Destroy deletes the provider** -- The tenant falls back to Auth0's built-in test provider: 10 emails per minute, from `no-reply@auth0user.net`, and no custom templates.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies. The tenant is the one the Auth0 connection's credential belongs to; credentials come from organization secrets.

### What This Kind Provides

After provisioning, `status.outputs` contains:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `name` | The service as Auth0 names it (`smtp`, `ses`, ...) | Audits |
| `default_from_address` | The sender of the tenant's emails | An Auth0 Email Template's `from`, documentation |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Resend over SMTP** -- Resend through its SMTP interface. Start from the **Resend Over SMTP** preset.

**Amazon SES** -- An IAM access key and a configuration set for delivery events. Start from the **Amazon SES** preset.

**SendGrid** -- One restricted API key. Start from the **SendGrid** preset.

## Works With

- [**Auth0 Email Template**](/infra-catalog/auth0-email-template) -- the emails sent through this provider, in your voice.
- [**Auth0 Action**](/infra-catalog/auth0-action) -- the sender for the `custom` arm, bound to the `custom-email-provider` trigger.
- [**Auth0 Tenant Settings**](/infra-catalog/auth0-tenant-settings) -- the tenant's name and support contacts the emails show.

# Auth0EmailProvider

Manages the [email provider](https://auth0.com/docs/customize/email/smtp-email-providers) an Auth0 tenant sends through: the verification, password-reset, invitation and multi-factor emails leave from your own sending domain and service, instead of Auth0's built-in test provider.

## When to Use

- **Production email**: Auth0's built-in provider sends at most 10 emails per minute, always from `no-reply@auth0user.net`, and cannot use custom templates. Any tenant real people sign in through needs its own provider.
- **Your domain as the sender**: people receive `Acme <no-reply@acme.com>`, signed for your domain, not a stranger's address that spam filters distrust.
- **Custom email templates**: an `Auth0EmailTemplate` needs a provider first -- Auth0 refuses templates without one.
- **One sending stack**: route Auth0's emails through the same service and domain as the product's own.

## Quick Start

```yaml
apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0EmailProvider
metadata:
  name: email-provider
  org: acme-corp
  env: production
spec:
  defaultFromAddress: Acme <no-reply@acme.com>
  smtp:
    host: smtp.resend.com
    port: 587
    user: resend
    password: $secret/resend-api-key
```

## Fields

| Field | Description |
|---|---|
| `defaultFromAddress` | The sender of every email, unless a template names its own (required) |
| `enabled` | Sending through this service is on; defaults to `true`. Off, the tenant falls back to the built-in provider and keeps the configuration |
| `smtp` | Any SMTP server: `host`, `port`, `user`, `password`, and optional `headers` (`xMcViewContentLink`, `xSesConfigurationSet`) |
| `ses` | Amazon SES: `accessKeyId`, `secretAccessKey`, `region`, optional `configurationSetName` |
| `sendgrid` | Twilio SendGrid: `apiKey` |
| `sparkpost` | SparkPost: `apiKey`, optional `region` (`eu`) |
| `mailgun` | Mailgun: `apiKey`, `domain`, optional `region` (`eu`) |
| `mandrill` | Mailchimp Transactional: `apiKey`, optional `viewContentLink` |
| `azureCs` | Azure Communication Services: `connectionString` |
| `ms365` | Microsoft 365 (Exchange Online): `tenantId`, `clientId`, `clientSecret` of an app registration granted `Mail.Send` |
| `custom` | No settings: an Action bound to the tenant's `custom-email-provider` trigger sends the emails |

Exactly one service is set, and it carries only what that service uses. Every credential is sensitive: declare it as a secret reference (`$secret/<name>`), never a literal.

## Key Behaviors

- **One per tenant**: the provider managed is the one of the tenant the provider connection's credential belongs to. Applying to a tenant that already has a provider takes it over and rewrites it.
- **Resend goes through SMTP**: Auth0's API accepts `resend` by name, but terraform-provider-auth0 v1.58.0, which this kind deploys through, accepts only `azure_cs`, `custom`, `mailgun`, `mandrill`, `ms365`, `sendgrid`, `ses`, `smtp` and `sparkpost` (`internal/auth0/email/resource.go:35-38`, the `name` validation). Send through Resend's SMTP interface with the `smtp` arm instead: host `smtp.resend.com`, user `resend`, an API key as the password (the Resend Over SMTP preset).
- **Credentials are write-only**: Auth0 never returns a credential, so the provider sends them only when they change, and after an import the first apply sends them again.
- **Destroy is a real delete**: the provider is deleted, and the tenant falls back to Auth0's built-in test provider -- emails keep flowing, rate-limited and from `no-reply@auth0user.net`, and custom templates stop being used.
- **Plans**: every plan, the Free plan included, can use its own provider; the sending service bills its own volume.
- **Permissions**: the credential needs `read:email_provider`, `create:email_provider`, `update:email_provider` and `delete:email_provider` on the tenant's Management API (`iac/permissions.yaml`).

## What the Spec Does Not Carry

Every argument of `auth0_email_provider` is reachable. The provider's `name` is not a field: the service arm set chooses it.

## Outputs

| Output | Description |
|---|---|
| `name` | The service the tenant sends through, as Auth0 names it (`smtp`, `ses`, `sendgrid`, ...) |
| `default_from_address` | The sender of the tenant's emails |

## Auth0 Documentation

- [Email providers](https://auth0.com/docs/customize/email/smtp-email-providers)
- [Terraform auth0_email_provider](https://registry.terraform.io/providers/auth0/auth0/latest/docs/resources/email_provider)
- [Pulumi auth0.EmailProvider](https://www.pulumi.com/registry/packages/auth0/api-docs/emailprovider/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

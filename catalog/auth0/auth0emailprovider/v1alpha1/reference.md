# Auth0EmailProvider

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `auth0.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

Auth0EmailProviderSpec manages the service the Auth0 tenant the provider
connection's credential belongs to sends its emails through.

Exactly one service arm is set, and it carries only the settings that service
uses: smtp, ses, sendgrid, sparkpost, mailgun, mandrill, azure_cs, ms365, or
custom (an Action sends the emails). Every credential is sensitive: declare it
as a secret reference, never a literal.

Resend: Auth0's API accepts Resend by name, but the Terraform provider this
kind deploys through does not; send through Resend's SMTP interface with the
smtp arm instead (the resend-over-smtp preset).

Destroying the resource deletes the provider, and the tenant falls back to
Auth0's built-in test provider, which sends only a small number of emails and
is meant for trying Auth0 out.

The credential needs read:email_provider, create:email_provider,
update:email_provider and delete:email_provider on the tenant's Management
API (iac/permissions.yaml).

https://auth0.com/docs/customize/email/smtp-email-providers
https://registry.terraform.io/providers/auth0/auth0/latest/docs/resources/email_provider

## Example

```yaml
# Auth0 Email Provider Test Manifest
# This file is used for testing the Auth0EmailProvider kind.
#
# Applying it REPLACES the email provider of the tenant the credential
# belongs to: run it only against a test tenant whose emails nobody relies
# on.
#
# Prerequisites:
# 1. Set the following environment variables:
#    - AUTH0_DOMAIN: the test tenant's domain (e.g., your-test-tenant.auth0.com)
#    - AUTH0_CLIENT_ID: M2M application client ID
#    - AUTH0_CLIENT_SECRET: M2M application client secret
#
# 2. The M2M application must have these scopes:
#    - read:email_provider
#    - create:email_provider
#    - update:email_provider
#    - delete:email_provider
#
# 3. A sending domain verified with the email service, and an organization
#    secret holding the service's API key.

apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0EmailProvider
metadata:
  name: test-email-provider
  org: test-org
  env: development
  labels:
    purpose: testing
spec:
  # The sender of every email the tenant sends, on the verified domain
  defaultFromAddress: Acme <no-reply@acme.com>

  # Resend, through its SMTP interface
  smtp:
    host: smtp.resend.com
    port: 587
    user: resend
    # A Resend API key with sending access, by secret reference
    password: $secret/resend-api-key
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.defaultFromAddress` | `string` | yes |  |  |
| `spec.enabled` | `bool` |  | `true` |  |
| `spec.smtp` | `Auth0EmailProviderSmtp` |  |  |  |
| `spec.smtp.host` | `string` | yes |  |  |
| `spec.smtp.port` | `int32` |  |  |  |
| `spec.smtp.user` | `string` | yes |  |  |
| `spec.smtp.password` | `string` (sensitive) | yes |  |  |
| `spec.smtp.headers` | `Auth0EmailProviderSmtpHeaders` |  |  |  |
| `spec.smtp.headers.xMcViewContentLink` | `string` |  |  |  |
| `spec.smtp.headers.xSesConfigurationSet` | `string` |  |  |  |
| `spec.ses` | `Auth0EmailProviderSes` |  |  |  |
| `spec.ses.accessKeyId` | `string` (sensitive) | yes |  |  |
| `spec.ses.secretAccessKey` | `string` (sensitive) | yes |  |  |
| `spec.ses.region` | `string` | yes |  |  |
| `spec.ses.configurationSetName` | `string` |  |  |  |
| `spec.sendgrid` | `Auth0EmailProviderSendgrid` |  |  |  |
| `spec.sendgrid.apiKey` | `string` (sensitive) | yes |  |  |
| `spec.sparkpost` | `Auth0EmailProviderSparkpost` |  |  |  |
| `spec.sparkpost.apiKey` | `string` (sensitive) | yes |  |  |
| `spec.sparkpost.region` | `string` |  |  |  |
| `spec.mailgun` | `Auth0EmailProviderMailgun` |  |  |  |
| `spec.mailgun.apiKey` | `string` (sensitive) | yes |  |  |
| `spec.mailgun.domain` | `string` | yes |  |  |
| `spec.mailgun.region` | `string` |  |  |  |
| `spec.mandrill` | `Auth0EmailProviderMandrill` |  |  |  |
| `spec.mandrill.apiKey` | `string` (sensitive) | yes |  |  |
| `spec.mandrill.viewContentLink` | `bool` |  |  |  |
| `spec.azureCs` | `Auth0EmailProviderAzureCs` |  |  |  |
| `spec.azureCs.connectionString` | `string` (sensitive) | yes |  |  |
| `spec.ms365` | `Auth0EmailProviderMs365` |  |  |  |
| `spec.ms365.tenantId` | `string` (sensitive) | yes |  |  |
| `spec.ms365.clientId` | `string` (sensitive) | yes |  |  |
| `spec.ms365.clientSecret` | `string` (sensitive) | yes |  |  |
| `spec.custom` | `Auth0EmailProviderCustom` |  |  |  |

## Field Details

### spec.defaultFromAddress

`string` · required

default_from_address is the sender of every email the tenant sends, unless
a template names its own (Auth0EmailTemplate.from): an address on a domain
the service is allowed to send for, optionally with a display name
("Acme <no-reply@acme.com>").

- rule: {"required":true}

### spec.enabled

`bool` · optional (explicit presence)

enabled turns sending through this service on. Disabled, the tenant falls
back to Auth0's built-in test provider while the configuration is kept.

- default: `true`

### spec.smtp

`Auth0EmailProviderSmtp`

smtp sends through any SMTP server (Resend, Postmark, Google Workspace,
your own relay).

### spec.smtp.host

`string` · required

host is the SMTP server's host name or IP address (for example
"smtp.resend.com").

- rule: smtp.host is a host name or an IP address, such as smtp.resend.com
- rule: {"required":true}

### spec.smtp.port

`int32`

port is the SMTP port, required. 587 (STARTTLS) is the one to prefer; 465
(implicit TLS) when the server requires it; 25 is often blocked.

- rule: {"int32":{"lte":65535,"gte":1}}

### spec.smtp.user

`string` · required

user is the SMTP user name (for Resend, "resend").

- rule: {"required":true}

### spec.smtp.password

`string` · required · sensitive

password is the SMTP password (for Resend, an API key with sending
access).

- rule: {"required":true}

### spec.smtp.headers

`Auth0EmailProviderSmtpHeaders`

headers are the SMTP headers Auth0 adds to every message.

### spec.smtp.headers.xMcViewContentLink

`string`

x_mc_view_content_link sets X-MC-ViewContentLink ("true" or "false"), the
Mandrill-over-SMTP switch for the "view content" link in its activity log.

- rule: {"string":{"in":["","true","false"]}}

### spec.smtp.headers.xSesConfigurationSet

`string`

x_ses_configuration_set sets X-SES-Configuration-Set, the SES
configuration set (event publishing, dedicated IPs) an SES-over-SMTP
message is sent with.

### spec.ses

`Auth0EmailProviderSes`

ses sends through Amazon Simple Email Service.

### spec.ses.accessKeyId

`string` · required · sensitive

access_key_id is the access key id of an IAM user allowed ses:SendRawEmail.

- rule: {"required":true}

### spec.ses.secretAccessKey

`string` · required · sensitive

secret_access_key is that access key's secret.

- rule: {"required":true}

### spec.ses.region

`string` · required

region is the SES region the sending domain is verified in (for example
"eu-west-1").

- rule: {"required":true}

### spec.ses.configurationSetName

`string`

configuration_set_name is the SES configuration set messages are sent
with.

### spec.sendgrid

`Auth0EmailProviderSendgrid`

sendgrid sends through Twilio SendGrid.

### spec.sendgrid.apiKey

`string` · required · sensitive

api_key is a SendGrid API key with Mail Send access.

- rule: {"required":true}

### spec.sparkpost

`Auth0EmailProviderSparkpost`

sparkpost sends through SparkPost.

### spec.sparkpost.apiKey

`string` · required · sensitive

api_key is a SparkPost API key with Transmissions access.

- rule: {"required":true}

### spec.sparkpost.region

`string`

region is the SparkPost region the account lives in: "eu" for SparkPost
EU, empty for the US.

### spec.mailgun

`Auth0EmailProviderMailgun`

mailgun sends through Mailgun.

### spec.mailgun.apiKey

`string` · required · sensitive

api_key is a Mailgun sending API key.

- rule: {"required":true}

### spec.mailgun.domain

`string` · required

domain is the Mailgun sending domain (for example "mg.acme.com").

- rule: {"required":true,"string":{"minLen":"4"}}

### spec.mailgun.region

`string`

region is the Mailgun region the domain lives in: "eu" for Mailgun EU,
empty for the US.

### spec.mandrill

`Auth0EmailProviderMandrill`

mandrill sends through Mailchimp Transactional (Mandrill).

### spec.mandrill.apiKey

`string` · required · sensitive

api_key is a Mandrill API key.

- rule: {"required":true}

### spec.mandrill.viewContentLink

`bool` · optional (explicit presence)

view_content_link shows a "view content" link for each message in
Mandrill's activity log.

### spec.azureCs

`Auth0EmailProviderAzureCs`

azure_cs sends through Azure Communication Services.

### spec.azureCs.connectionString

`string` · required · sensitive

connection_string is the Communication Services resource's connection
string.

- rule: {"required":true}

### spec.ms365

`Auth0EmailProviderMs365`

ms365 sends through Microsoft 365 (Exchange Online) with an app
registration.

### spec.ms365.tenantId

`string` · required · sensitive

tenant_id is the Microsoft Entra tenant the app registration lives in.

- rule: {"required":true}

### spec.ms365.clientId

`string` · required · sensitive

client_id is the app registration's application (client) id, granted
Mail.Send.

- rule: {"required":true}

### spec.ms365.clientSecret

`string` · required · sensitive

client_secret is the app registration's client secret.

- rule: {"required":true}

### spec.custom

`Auth0EmailProviderCustom`

custom leaves sending to an Action on the tenant's custom email provider
trigger; the Action holds its own credentials.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: Auth0EmailProvider, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | name is the service the tenant sends through, as Auth0 names it (for example "smtp", "ses", "sendgrid"). |
| `status.outputs.default_from_address` | `string` | default_from_address is the sender of the tenant's emails. |

## See Also

- [Overview](../README.md)

# Amazon SES

This preset sends the tenant's emails through Amazon Simple Email Service, from `no-reply@planton.ai`, with a configuration set that publishes their delivery, bounce and complaint events.

## When to Use

- Your product's email already runs on SES, in an AWS account you control
- You want Auth0's emails in the same bounce and complaint handling as your own

## Key Configuration Choices

- **Access key** (`ses.accessKeyId`, `ses.secretAccessKey`) -- an IAM user allowed only `ses:SendRawEmail`, both values read from organization secrets. Rotate the key by updating the secrets and applying again.
- **Region** (`ses.region`) -- the region the sending domain is verified in; SES identities are regional
- **Configuration set** (`ses.configurationSetName`) -- optional; leave it out to send without event publishing
- **Out of the sandbox** -- a new SES account sends only to verified addresses until AWS grants production access

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.defaultFromAddress` | The sender, on a domain verified in SES | SES console, Identities |
| `spec.ses.accessKeyId`, `spec.ses.secretAccessKey` | The secrets holding the IAM user's access key | IAM console, the user's Security credentials; store both as organization secrets |
| `spec.ses.region` | The region of the verified identity | SES console, the region selector |
| `spec.ses.configurationSetName` | Your configuration set | SES console, Configuration sets |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **01-resend-over-smtp** -- Resend through its SMTP interface
- **03-sendgrid** -- Twilio SendGrid with an API key

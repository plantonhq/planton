# Auth0 Prompt Custom Text Guide

## Security
## Platform Security Posture

The certifications below are Auth0's own published claims about their hosted platform (verify current status on Auth0's compliance page). They describe the vendor's service — never this catalog kind, and never your deployment: configuring this resource does not make your application certified, authorized, or compliant with any framework.

Auth0's published certifications and security standards:

- SOC 2 Type II (annual audit)
- ISO 27001, ISO 27018 (privacy controls)
- HIPAA BAA available on enterprise plans
- PCI DSS Level 1 Service Provider
- FedRAMP Authorized (moderate baseline)
- CSA STAR Level 2
- GDPR compliant with Data Processing Agreement

## Data Protection

- **Data residency**: US, EU, AU regions
- **Encryption in transit**: TLS 1.2+
- **Encryption at rest**: AES-256
- **Penetration testing**: Annual third-party assessments

## Custom-Text Security Notes

### Public by Design

Every word this component sets is shown to anyone who opens the tenant's login pages. Put nothing in it you would not publish.

### Error Messages Should Not Tell More Than Auth0's

Auth0's default errors are careful not to reveal whether an account exists ("Wrong email or password"). Keep that property when you re-word them: "No account with this email" tells an attacker which addresses are registered.

### Words That Ask for Secrets Invite Phishing

Never re-word a screen to ask for anything beyond what the screen collects, and never point people at an address that is not yours. A support address in an error message should be a monitored mailbox that knows it must never ask for a password.

## Pairing and Apply Order

### Universal Login First

Custom text reaches only the `new` experience. Apply the Auth0 Prompt Infra Component with `universalLoginExperience: new` before, or in the same install as, the custom text.

### Pick the Prompts Your Flow Uses

The prompt a person sees depends on the flow: identifier-and-password tenants sign in on `login` and sign up on `signup`; identifier-first tenants (Auth0 Prompt's `identifierFirst`) use `login-id` and `login-password`, `signup-id` and `signup-password`. Re-wording a prompt the flow never shows changes nothing people see.

### One Resource Owns a Prompt's Words in a Language

Auth0 replaces the whole custom text of a prompt and language on every write. Two resources for the same pair overwrite each other on every apply, and words changed in the dashboard are replaced at the next apply. Declare every screen of the prompt in one resource.

### Prompt and Language Are Fixed

The prompt and language are the custom text's identity. Changing either in place writes the words to the new pair and leaves the old pair's words as they were; to move, destroy the resource (which returns the old pair to Auth0's defaults) and declare a new one.

### Variables for Partials

A key named `var-<name>` (up to thirty per screen and language) defines words a screen partial reads: `var-tos` is `prompt.screen.text.varTos` in the Auth0 Prompt Screen Partials kind's HTML. Define the variable here in every language you serve, and the partial's label follows the person's language. Partials render only with Auth0 Branding's `universal_login_template`, which needs a custom domain; custom text itself needs neither.

### The Names in the Sentence

`${companyName}` is the tenant's friendly name (Auth0 Tenant Settings) and `${clientName}` the application's name (Auth0 Client). Set those rather than hard-coding a name into every prompt.

## Permissions
## Management API Scopes

Auth0 custom text requires the following Management API scopes. They must be granted to the M2M application used for infrastructure automation.

| Operation | Scope | Description |
|-----------|-------|-------------|
| Read | `read:prompts` | Read a prompt's custom text in a language |
| Update | `update:prompts` | Set a prompt's custom text in a language, and empty it on destroy |

There is no delete scope to grant: destroy writes an empty custom text.

## Compliance
## Custom-Text Compliance Notes

### The Sign-In Copy as Auditable Configuration

The words people read when they sign in -- including consent and terms language -- are version-controlled with the rest of the tenant's identity configuration, so a change to them is reviewed like any other change. Changes are also recorded in Auth0's tenant logs.

## Cost
## Pricing Model

Auth0 pricing is based on the plan tier and monthly active users, not on settings. Custom text is available on every plan, the Free plan included, in every language the tenant enables, and needs no custom domain.

## Cost Impact

Changing the words has no billing impact.

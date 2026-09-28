# Auth0 Prompt Screen Partials Guide

## Security
## Platform Security Posture

The certifications below are Auth0's own published claims about their hosted platform (verify current status on Auth0's compliance page). They describe the vendor's service — never this Planton component, and never your deployment: configuring this resource does not make your application certified, authorized, or compliant with any framework.

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

## Partials Security Notes

### A Partial Runs on Your Login Page

A fragment's HTML and JavaScript run on the page where people type their password. Review every change to a partial like a change to the sign-in itself, load scripts only from hosts you control, and never from a CDN you would not trust with a password field.

### Escape What You Echo

A partial may print Liquid variables. Pass anything that came from a person -- a name, a query parameter -- through Liquid's `escape` filter before it reaches the page, or the page renders it as HTML.

### Validate on the Server

A field's `required` or pattern attributes stop an honest browser, not a crafted request. Validate every `ulp-` field again in the Action that reads it, and refuse the sign-up or login with `api.validation.error` when it fails.

### Collect Only What You May

Auth0 asks that sensitive or regulated data be collected through partials only as your agreement with Okta permits. Collect the least the product needs.

## Pairing and Apply Order

### The Chain a Partial Needs

Auth0 accepts partials only on a tenant with both a custom domain and a page template, and refuses them otherwise (measured live): 403 "This feature requires at least one custom domain to be configured for the tenant", then 403 "This feature requires a page template to be configured for the tenant". The domain need not be verified for Auth0 to accept them, but people see the partials only on the verified domain's login pages. Apply in this order: the Auth0 Custom Domain and its verification, Auth0 Prompt with `universalLoginExperience: new`, Auth0 Branding with `universalLoginTemplate`, then this component.

### Pick the Prompts Your Flow Uses

Identifier-and-password tenants sign in on `login` and sign up on `signup`; identifier-first tenants (Auth0 Prompt's `identifierFirst`) show `login-id` and `login-password`, `signup-id` and `signup-password`. A partial on a prompt the flow never shows changes nothing people see.

### One Resource Owns a Prompt's Partials

Auth0 replaces a prompt's whole set of partials on every write. Two resources for the same prompt overwrite each other on every apply, and partials edited in the dashboard are replaced at the next apply. Declare every screen of the prompt in one resource.

### Never Mix With the Single-Screen Resource

The provider's one-screen `auth0_prompt_screen_partial` adds or replaces one screen of a prompt; its every argument is one `screenPartials` entry here, which is where it belongs. A one-screen resource on a prompt this kind manages -- in another stack, or applied by hand -- is overwritten at every apply, and overwrites this kind's set in turn.

### Translated Labels Come From Custom Text

Define `var-<name>` keys with the Auth0 Prompt Custom Text kind, per language, and read them in the partial as `prompt.screen.text.<name in camelCase>` (`var-tos` is `prompt.screen.text.varTos`). The label then follows the person's language.

### Actions Read the Fields

A `ulp-` field reaches `event.request.body` in the pre-user-registration trigger (sign-up) and the post-login trigger (login). Deploy the Action that reads it before, or with, the partial that adds it, and remove the Action's reliance before you remove the field.

## Permissions
## Management API Scopes

Auth0 prompt partials require the following Management API scopes. They must be granted to the M2M application used for infrastructure automation.

| Operation | Scope | Description |
|-----------|-------|-------------|
| Read | `read:prompts` | Read a prompt's partials |
| Update | `update:prompts` | Set a prompt's partials, and empty them on destroy |

There is no delete scope to grant: destroy writes an empty set.

## Compliance
## Partials Compliance Notes

### What the Sign-In Collects, as Auditable Configuration

The fields a sign-up or login screen adds, and the consent language beside them, are version-controlled with the rest of the tenant's identity configuration, so a change to what the product collects at sign-in is reviewed like any other change. Changes are also recorded in Auth0's tenant logs.

## Cost
## Pricing Model

Auth0 pricing is based on the plan tier and monthly active users, not on settings. Partials cost nothing themselves, but they are gated on a custom domain: they render only inside a page template, and page templates need a custom domain. Every plan, the Free plan included, has one custom domain with Auth0-managed certificates at no charge (with a card on file).

## Cost Impact

Changing the partials has no billing impact.

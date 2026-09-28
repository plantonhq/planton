# Auth0 Branding Guide

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

## Branding Security Notes

### A Page Template Runs on Your Sign-In Page

Everything `universalLoginTemplate` loads -- stylesheets, scripts, fonts, images -- runs on the page where people type their passwords. Load only from hosts you control, pin what you can, and review a template change like a code change. Never load analytics or tag managers you would not trust with a keystroke.

### Assets Are Fetched From Where You Host Them

The login pages load `logoUrl`, `faviconUrl`, `fontUrl` and the theme's image and font URLs from their hosts on every visit. Host them somewhere you control, over HTTPS, and keep the URLs answering: a moved asset breaks the page's look (never the sign-in itself).

### Look Is Trust

People decide whether a login page is real by how it looks and where it lives. Keep the branded page on your custom domain, and keep the look consistent across your tenants, so a lookalike page stands out.

## Working With Other Auth0 Kinds

### The Order to Apply

1. **Auth0 Custom Domain**, its DNS record, and **Auth0 Custom Domain Verification** -- only when you set a page template.
2. **Auth0 Branding** -- with `depends_on` the verification when it carries a page template; otherwise in any order.
3. **Auth0 Tenant Settings** -- the tenant's name and support contacts; independent of this kind.

### Which Kind Owns Which Look

- **The logo**: `logoUrl` here takes precedence over Auth0 Tenant Settings' `pictureUrl` on the login pages. The theme's `widget.logoUrl`, when set, overrides both inside the login box.
- **The font**: `fontUrl` here styles the pages; the theme's `fonts.fontUrl`, when set, styles the login box.
- **The words**: the login page's sentence ("Log in to *tenant* to continue to *application*") comes from Auth0 Tenant Settings' `friendlyName` and the Auth0 Client's `name`, never from branding.

## Traps

### A Template Before the Domain

Auth0 refuses a page template on a tenant without a custom domain, and the apply fails. Order the branding after the Auth0 Custom Domain Verification with `depends_on`.

### Both Tags, Exactly

The template must contain `{%- auth0:head -%}` inside `<head>` and `{%- auth0:widget -%}` where the login box goes, spelled exactly so; the spec refuses a template without both.

### Removing a Setting Is Not Resetting It

A logo, favicon or color removed from the spec is no longer managed, and the tenant keeps its last-applied value. To change what the page shows, set the new value; to return to Auth0's look, set Auth0's value before removing the field. The font and the page template are the exceptions: removing them returns the pages to Auth0's font and page.

### One Theme per Tenant

A tenant has one theme. Applying a theme adopts one the tenant already carries (one made in the dashboard, say) and overwrites it with the declaration; destroying the resource deletes it. Keep one Auth0 Branding resource per tenant.

### Identifiers Stick

Once applied, `theme.identifiers` can be changed but not removed. Declare it only when the tenant has the feature and you mean to keep it.

## Permissions
## Management API Scopes

Auth0 branding requires the following Management API scopes. They must be granted to the M2M application used for infrastructure automation.

| Operation | Scope | Description |
|-----------|-------|-------------|
| Read | `read:branding` | Read the branding, the page template and the theme |
| Update | `update:branding` | Change the branding, set the page template, create and update the theme |
| Delete | `delete:branding` | Remove the page template and delete the theme |
| Check for a custom domain | `read:custom_domains` | List the tenant's custom domains, which the provider does on every read, update and delete |

There is no create scope for the branding itself: a tenant has one, and it cannot be deleted.

## Compliance
## Branding Compliance Notes

### The Login Look as Auditable Configuration

The tenant's login look -- its logo, colors, page template and theme -- is version-controlled with the rest of its identity configuration, so a change to the page where people type their passwords is reviewed like any other change. Changes are also recorded in Auth0's tenant logs.

## Cost
## Pricing Model

Auth0 pricing is based on the plan tier and monthly active users, not on branding. The logo, favicon, colors, font and theme are available on every plan, the Free plan included. The page template is gated on a custom domain rather than a plan; Auth0's Free plan includes one custom domain.

## Cost Impact

Changing the branding has no billing impact.

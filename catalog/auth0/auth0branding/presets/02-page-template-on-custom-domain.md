# Page Template on a Custom Domain

This preset renders the tenant's login box inside a page of your own: your header, your footer and your styles around Auth0's widget, together with your logo, favicon and colors. Auth0 accepts a page template only on a tenant with a custom domain, so the preset composes with an Auth0 Custom Domain and its Auth0 Custom Domain Verification (start from the Auth0 Custom Domain's **Auth0-Managed Domain on Cloudflare** preset) and depends on the verification, so the template is applied after the domain serves.

## When to Use

- Your sign-in must look like the rest of your site, with the same header, footer and legal links
- The tenant already signs people in on your own domain, or is about to (the Free plan includes one custom domain)

## Key Configuration Choices

- **Page template** (`universalLoginTemplate`) -- a Liquid HTML document; it must contain `{%- auth0:head -%}` inside `<head>` (where Auth0 injects its styles and scripts) and `{%- auth0:widget -%}` where the login box renders, and the spec refuses it without both
- **Liquid variables** -- the page can read Auth0's context, such as `{{ locale }}` and `{{ prompt.screen.texts.pageTitle }}`
- **Ordered after the verification** (`relationships` `depends_on` the Auth0 Custom Domain Verification) -- Auth0 refuses the template until the tenant has a custom domain
- **Removing the template** returns the pages to Auth0's default page; destroying the resource removes it too, and keeps the logo, favicon and colors

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.logoUrl`, `spec.faviconUrl` | Public HTTPS URLs of your logo and favicon | Your brand assets host |
| `spec.colors` | Your accent color and page background | Your brand guidelines |
| `spec.universalLoginTemplate` | Your page: header, footer, styles and links | Your site's layout |
| `metadata.relationships[0].name` | Your Auth0 Custom Domain Verification resource | Your chart or manifest set |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **01-logo-and-colors** -- the logo and colors alone, on any tenant
- **03-themed-login-box** -- the login box's own look inside the page

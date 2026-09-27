# Auth0TenantSettings

Manages how an existing [Auth0 tenant](https://auth0.com/docs/get-started/tenant-settings) presents itself to the people who sign in through it: the name Universal Login shows, the logo beside it, and the support contacts its login and error pages offer.

## When to Use

- **Your product's name on the login page**: Universal Login reads "Log in to *friendly name* to continue to *application*". Without a friendly name it shows the tenant's identifier (for example `acme-prod`).
- **Your logo instead of Auth0's**: the login and consent pages show the tenant's picture.
- **A way to get help**: people who can't sign in see your support address and page.
- **Infrastructure as code**: the tenant's face is version-controlled beside the clients and connections it serves.

## Quick Start

```yaml
apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0TenantSettings
metadata:
  name: tenant-settings
  org: acme-corp
  env: production
spec:
  friendlyName: Acme
  pictureUrl: https://assets.acme.com/logo.png
  supportEmail: support@acme.com
  supportUrl: https://acme.com/support
```

## Key Behaviors

- **The tenant is the credential's**: the settings managed are those of the tenant the provider connection's credential belongs to. One resource per tenant.
- **Unset is unmanaged**: a field left out is never sent, and the tenant keeps whatever value it already carries. At least one field must be set.
- **Nothing is ever deleted**: Auth0's Management API cannot create or delete a tenant, and has no delete for its settings. Destroying this resource forgets the settings and leaves their last-applied values in place. To return a setting to a specific value, set that value before removing the field.
- **The application's half of the sentence** is each client's name (the `Auth0Client` kind's `name`, defaulting to its `metadata.name`).
- **Permissions**: the credential needs `read:tenant_settings` and `update:tenant_settings` on the tenant's Management API (`iac/permissions.yaml`).

## Outputs

| Output | Description |
|---|---|
| `friendly_name` | The tenant's name as people see it |
| `picture_url` | The URL of the tenant's logo |
| `support_email` | The support address the tenant's pages offer |
| `support_url` | The support page the tenant's pages link to |

## Auth0 Documentation

- [Tenant settings](https://auth0.com/docs/get-started/tenant-settings)
- [Terraform auth0_tenant](https://registry.terraform.io/providers/auth0/auth0/latest/docs/resources/tenant)
- [Pulumi auth0.Tenant](https://www.pulumi.com/registry/packages/auth0/api-docs/tenant/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

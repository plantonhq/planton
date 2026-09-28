# Auth0TenantSettings

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `auth0.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

Auth0TenantSettingsSpec manages how an existing Auth0 tenant presents itself
to the people who sign in through it: the name Universal Login shows ("Log in
to <friendly_name> to continue to <application>"), the logo beside it, and the
support contacts the login and error pages offer.

The tenant is the one the provider connection's credential belongs to. Auth0's
Management API cannot create or delete a tenant, so this kind never does either:
it manages the settings of the tenant that already exists.

Each field maps to one tenant setting (PATCH /api/v2/tenants/settings). A field
left unset is NOT MANAGED: the module never sends it, and whatever value the
tenant already carries stays untouched. Setting a field manages that setting;
clearing it back to unset stops managing it but does NOT revert the live value.
Auth0 has no delete for tenant settings, so destroy abandons the last-applied
values. To return a setting to a specific value, set that value explicitly
before removing the field.

The tenant's default custom domain (default_custom_domain) is managed the same
way: unset, Auth0 keeps whatever default the tenant has.

The credential needs read:tenant_settings and update:tenant_settings on the
tenant's Management API, and read:custom_domains and update:custom_domains when
default_custom_domain is set (iac/permissions.yaml).

https://auth0.com/docs/get-started/tenant-settings
https://auth0.com/docs/customize/custom-domains/multiple-custom-domains/default-domain
https://registry.terraform.io/providers/auth0/auth0/latest/docs/resources/tenant
https://www.pulumi.com/registry/packages/auth0/api-docs/tenant/

## Example

```yaml
# Auth0 Tenant Settings Test Manifest
# This file is used for testing the Auth0TenantSettings component.
#
# Applying it REWRITES the settings of the tenant the credential belongs to:
# run it only against a test tenant nobody signs in to.
#
# Prerequisites:
# 1. Set the following environment variables:
#    - AUTH0_DOMAIN: the test tenant's domain (e.g., your-test-tenant.auth0.com)
#    - AUTH0_CLIENT_ID: M2M application client ID
#    - AUTH0_CLIENT_SECRET: M2M application client secret
#
# 2. The M2M application must have these scopes:
#    - read:tenant_settings
#    - update:tenant_settings

apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0TenantSettings
metadata:
  name: test-tenant-settings
  org: test-org
  env: development
  labels:
    purpose: testing
spec:
  # The name Universal Login shows: "Log in to <friendlyName> to continue to <application>"
  friendlyName: Acme

  # The logo shown on the login and consent pages
  pictureUrl: https://assets.example.com/logo.png

  # The support contacts the login and error pages offer
  supportEmail: support@example.com
  supportUrl: https://example.com/support
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.friendlyName` | `string` |  |  |  |
| `spec.pictureUrl` | `string` |  |  |  |
| `spec.supportEmail` | `string` |  |  |  |
| `spec.supportUrl` | `string` |  |  |  |
| `spec.defaultCustomDomain` | `string \| valueFrom` |  |  | Auth0CustomDomainVerification (`status.outputs.domain`) |

## Field Details

### spec.friendlyName

`string`

friendly_name is the tenant's name as people see it: Universal Login's
"Log in to <friendly_name> to continue to <application>", and the name in
the emails Auth0 sends for the tenant. Unset, Auth0 shows the tenant's
identifier (for example "acme-prod").

### spec.pictureUrl

`string`

picture_url is the URL of the logo shown for the tenant on its login
and consent pages (Auth0 recommends 150 x 150 pixels). Unset, Auth0's own
logo is shown.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"uri":true}}

### spec.supportEmail

`string`

support_email is the address the tenant's login and error pages offer to a
person who needs help signing in.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"email":true}}

### spec.supportUrl

`string`

support_url is the page the tenant's login and error pages link to for
help.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"uri":true}}

### spec.defaultCustomDomain

`string | valueFrom`

default_custom_domain is the domain that speaks for the tenant when a
request does not say which of its domains it came through: the links in the
emails Auth0 sends (verification, password reset, invitations) and the
notifications the Management API triggers. Set it to the tenant's custom
domain so a person who signs in at id.example.com also receives links to
id.example.com, never to the tenant's canonical auth0.com domain.

Only a verified domain can be the default, so reference the
Auth0CustomDomainVerification (its status.outputs.domain): the default is
then set only after Auth0 has verified the domain. A literal is also
accepted, including the tenant's canonical domain (e.g.
"example.eu.auth0.com") to make it the default again. Unset, the tenant's
default is not managed; clearing the field stops managing it and leaves the
last-applied default in place (Auth0 has no way to unset a default).

- references: Auth0CustomDomainVerification (`status.outputs.domain`)
- rule: write as {value: <literal>} or {valueFrom: {kind: Auth0CustomDomainVerification, name: <that resource's name>, fieldPath: status.outputs.domain}} -- a bare string does not parse

## Validation Rules

- `spec.at_least_one_setting`: configure at least one tenant setting -- an Auth0TenantSettings resource that manages nothing would deploy nothing

## Outputs

Reference an output from another manifest as `valueFrom: {kind: Auth0TenantSettings, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.friendly_name` | `string` | friendly_name is the tenant's name as people see it. |
| `status.outputs.picture_url` | `string` | picture_url is the URL of the tenant's logo. |
| `status.outputs.support_email` | `string` | support_email is the support address the tenant's pages offer. |
| `status.outputs.support_url` | `string` | support_url is the support page the tenant's pages link to. |
| `status.outputs.default_custom_domain` | `string` | default_custom_domain is the tenant's default domain as set by this resource; empty when the spec leaves the default unmanaged. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.defaultCustomDomain` | Auth0CustomDomainVerification | `status.outputs.domain` |

## See Also

- [Overview](../README.md)

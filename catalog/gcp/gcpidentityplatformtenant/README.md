# GCP Identity Platform Tenant

Creates one Identity Platform TENANT — an isolated user pool with its own sign-in configuration and identity providers, inside a project whose Identity Platform allows tenants. Tenants are how one project serves multiple customer organizations (B2B SaaS) with fully separated users, providers, and policies; each tenant's IdP client secrets are handled as managed secrets end to end.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Tenant** -- an `identity_platform_tenant` with its display name, sign-up switches, kill switch, and client permissions
- **Tenant-scoped IdP configs** -- one composed resource per entry in `defaultSupportedIdps` (Google, Facebook, ...), `oauthIdpConfigs` (custom OIDC), and `inboundSamlConfigs` (enterprise SSO), all scoped to this tenant only
- **Identity Toolkit API enablement** -- `identitytoolkit.googleapis.com` enabled in the target project (never disabled on destroy)

## Before You Deploy

### Facts That Shape Everything Else

- **The tenant ID is SERVER-GENERATED.** `displayName` is the only naming input you control; GCP mints the tenant's resource ID at create time. Read it from the `tenant_id` output — client SDKs and tenant-scoped API calls need it.
- **The project must already allow tenants.** The project's GcpIdentityPlatformConfig must set `multiTenant.allowTenants: true` — the tenant API rejects creation otherwise.
- **ONE `deletionPolicy` governs the tenant AND its IdP configs.** The IdP configs have no life apart from the tenant, so a single switch covers all of them.
- **Deleting a tenant deletes ALL its users — unrecoverable.** `deletionPolicy: PREVENT` is the right posture for any tenant with real accounts.
- **Tenant-level API differences from the project level**: OIDC providers REQUIRE `displayName` and have no `responseType`; SAML providers REQUIRE `spConfig` (both `callbackUri` and `spEntityId`).

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project with Identity Platform initialized and `allowTenants` enabled** (the GcpIdentityPlatformConfig kind). Provide the project ID directly or reference a GcpProject Infra Component via ValueFromRef.
- **IAM**: the deploying identity needs `roles/identityplatform.admin` or broader.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpIdentityPlatformTenant
metadata:
  name: acme-corp-tenant
spec:
  displayName: acme-corp
  allowPasswordSignup: true
```

```shell
planton apply -f tenant.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `displayName` | `string` | Human-readable tenant name (e.g. the customer organization) — the only naming input; the resource ID is server-generated. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider default | GCP project containing the tenant. Its Identity Platform config must allow tenants. |
| `allowPasswordSignup` | `bool` | `false` | Whether users can sign up with email/password in this tenant. |
| `enableEmailLinkSignin` | `bool` | `false` | Whether users can sign in via emailed magic link (passwordless). |
| `disableAuth` | `bool` | `false` | When true, ALL authentication in this tenant is disabled — the maintenance/kill switch. Existing sessions stop refreshing. |
| `clientPermissions` | `message` | — | `disabledUserSignup` / `disabledUserDeletion` — restrict what client apps can do against this tenant through the API. |
| `defaultSupportedIdps` | `list` | `[]` | Well-known providers (`idpId` such as `google.com`) enabled for THIS tenant, with console-issued `clientId`/`clientSecret`. |
| `oauthIdpConfigs` | `list` | `[]` | Custom OIDC providers — `name` starts with `oidc.`; `displayName` is REQUIRED at tenant level; no `responseType` exists here. |
| `inboundSamlConfigs` | `list` | `[]` | Enterprise SAML providers — `name` matching `saml.<slug>`; `spConfig` with BOTH `callbackUri` (https) and `spEntityId` is REQUIRED at tenant level. |
| `deletionPolicy` | `string` | `DELETE` | One switch for the tenant AND its IdP configs: `DELETE` (removes all users — unrecoverable), `PREVENT` (refuse), or `ABANDON`. |

### Validation Rules

- **`displayName` required** — the tenant's only naming input.
- **IdP naming**: `idpId` in the ten canonical values; OIDC names start `oidc.`; SAML names match `saml.<lowercase-start slug>`.
- **Tenant-level requirements**: OIDC `displayName` required; SAML `spConfig.callbackUri` (must be `https://`) and `spConfig.spEntityId` both required.
- **`deletionPolicy`** in `DELETE`/`PREVENT`/`ABANDON`.

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `tenant_id` | `string` | The server-generated tenant ID — what client SDKs set as `tenantId` to scope sign-in to this tenant |
| `tenant_name` | `string` | `projects/{project}/tenants/{tenant_id}` — the tenant's full resource name |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **The tenant ID never appears in the spec** — it flows from GCP through the `tenant_id` output. Wire client apps and downstream references through the output, never a hardcoded guess.
- **`deletionPolicy: DELETE` destroys every user account in the tenant** with no recovery path. Use `PREVENT` for production tenants.
- **`disableAuth: true` is a full-tenant kill switch** — nothing authenticates and existing sessions stop refreshing. Useful for offboarding or incident containment.
- **Unlike the project config, tenants are fully deletable** — deploy/destroy cycles are clean.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Kinds

- [GcpIdentityPlatformConfig](/docs/catalog/gcp/gcpidentityplatformconfig) — the project singleton whose `multiTenant.allowTenants` gates tenant creation
- [GcpProject](/docs/catalog/gcp/gcpproject) — provides the GCP project containing the tenant

## Additional Resources

- [Identity Platform Multi-tenancy](https://cloud.google.com/identity-platform/docs/multi-tenancy)
- [Tenants API Reference](https://cloud.google.com/identity-platform/docs/reference/rest/v2/projects.tenants)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

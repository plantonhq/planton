# A Tenant Serves MCP and Third-Party Clients From Files

**Date**: September 28, 2026
**Type**: Feature
**Components**: Auth0TenantSettings, Auth0ResourceServer, the new Auth0ClientFromMetadataDocument, the Auth0 import catalog, the Auth0 live-test verifiers and workflow

## Summary

**An Auth0 tenant that serves MCP clients and third-party applications is declared whole.** Before this change, the settings such a tenant depends on could not be expressed: the tenant's OAuth flags, an API's subject-type access policy, the grants third-party applications get by default, and registering an application from its metadata document. They lived in a hand-run bootstrap script. With this change, the only step outside files is creating the tenant itself, which Auth0's API does not offer.

- **`Auth0TenantSettings` models every setting of `auth0_tenant`**: 66 arguments plus the tenant's default custom domain. This covers sessions and their cookie; Client ID Metadata Document registration, the `resource` parameter profile, dynamic client registration and pushed authorization requests; ACR values; mTLS; OIDC logout; the default audience and directory, as references to the API and connection that produce them; the error page; locales and phone; default token quotas; and every tenant flag. It is at total parity accounting against provider 1.58.0. One deprecated flag, which the schema marks, is left out.
- **`Auth0ResourceServer` models every argument of `auth0_resource_server`** (45 accounted) and gains `third_party_client_default_grants`. The API now has:
  - subject-type authorization for users, clients and anonymous users;
  - the access-token claims mapping, authorization details and policy;
  - proof of possession and token encryption with its public key;
  - online and ephemeral access, and the anonymous-token lifetime.

  A default grant is declared on the API it grants, one per subject type, as an `auth0_client_grant` with `default_for: third_party_clients`: every third-party application, including one registered through dynamic registration or a metadata document, gets those scopes without a grant of its own.
- **`Auth0ClientFromMetadataDocument` (kind 8016)** registers an application from the Client ID Metadata Document at its URL, the path an MCP client onboards itself by (`auth0_client_cimd`, 27 arguments, total accounting). Its prerequisite is `Auth0TenantSettings`, because Auth0 accepts the registration only while the tenant allows it.

## What Changed

- **Unset means unmanaged in all three kinds.** Every new scalar has presence, every new block's presence manages it, and the modules send an argument only when the spec declares it. Each module has a test proving that an empty spec sends nothing new. Where the provider cannot leave an argument alone on an adopted tenant or API, the field comment and the GUIDE's adoption section say to declare the live value first.
- **`Auth0ResourceServer`'s three toggles gain presence.** `allow_offline_access`, `skip_consent_for_verifiable_first_party_clients` and `enforce_policies` are now `optional` and sent only when set. OpenTofu used to send `true` for an unset skip-consent and could not express `false`, while Pulumi sent `false`, so the two engines disagreed on a live setting. All three are provider-computed, so an API that leaves them unset previews no change on either engine.
- **Import.** `aa_import/catalog.yaml` gains the formats for `auth0_tenant`, `auth0_custom_domain_default`, `auth0_resource_server`, `auth0_resource_server_scopes`, `auth0_client_grant` and `auth0_client_cimd`, and each kind carries its import map, so a live tenant adopted by create-then-import lands in files.
- **Presets.** An MCP-ready tenant; a short-idle-sessions tenant; an API an MCP server exposes; an MCP client registered from its document; one with rotating refresh tokens.
- **Live tests.**
  - The new kind has a verifier (the registered client, by its client id) and entrypoints for both engines.
  - Its lanes register from a committed document served over HTTPS (`e2e/auth0/fixtures/cimd-client.json`).
  - The `e2e-auth0` workflow picks the test tenant at dispatch and runs one tenant at a time, since the per-tenant kinds would overwrite each other.
  - It passes the custom-domain lanes their DNS zone and token.
  - The spec tests assert the rule a refusal names.
- **Catalog pages.** Tenant settings and the custom-domain verification show how an InfraChart wires them, as the page gate requires of a page whose kind consumes references.

## Verification

- Spec and module tests for all three kinds, the verifiers, the fixture-integrity and output-conformance tests, and `go vet` of the Auth0 live tests.
- `planton provider-parity --kind` at total accounting: 67 of 67 for tenant settings, 45 of 45 for the API (4 excluded with reasons), 27 of 27 for the metadata-document kind.
- `planton module verify` on both engines of all three.
- The `secret-coverage` and `validate-refs` gates, and the catalog-page, import-map, permissions, finops, compliance, logo and reference-generation tests.
- The Bazel build of every touched package.
- The live runs are the next step. Every profile stays `pending_proof` until both engines pass on the test tenant.

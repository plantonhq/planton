# Auth0User Deployment Component

**Date**: September 22, 2026
**Type**: Feature
**Components**: API Definitions, Provider Framework, IAC Stack Runner, Testing Framework

## Summary

Adds `Auth0User` to the Planton catalog -- a deployment component that manages an Auth0 User in a database or passwordless connection as code: the profile, the verification posture, the two metadata documents, the authoritative role and direct-permission sets, and the password. When no password is declared on a database connection, both modules mint one and report it once as a sensitive output; a declared password is used as given and never echoed back; a passwordless connection takes none. The user's identity-provider subject (`auth0|...`, the `sub` claim in every token) is an output other declarations reference. The kind is wired into the Auth0 provider plumbing (registry, reflection map, E2E harness) and built for 100% parity with `auth0_user`, `auth0_user_roles`, and `auth0_user_permissions` at provider 1.57.0.

## Motivation

The tenant's clients, connections, roles, APIs, actions, and event streams were all declarable; the one identity kind an operator owns rather than a person who signs up -- a staff root, a service account, a seeded test identity -- was not. Declaring it by hand meant a password typed into a dashboard and a subject copied out of one, neither recreatable from files.

## What Changed

- **`catalog/auth0/auth0user/`** -- the complete component: the four protos with validation and a 19-case spec test; both IaC modules (the Terraform `variables.tf` generated from the proto); `cost.yaml`, `controls.yaml`, `iac/permissions.yaml`; the e2e profile, manifest, and two scenarios (`minimal` declares no password so the cheapest lane proves minting; `with-roles-and-permissions` adds role and resource-server fixtures through the prerequisites annotation); three presets; README, catalog page, guide (with the hand-derived parity table -- the Auth0 provider carries no parity schema artifact), and logo.
- **Registry** -- `Auth0User = 8006` (`a0user`) with `prerequisites: [Auth0Connection]`, so the E2E framework installs the connection fixture and resolves the user's reference against it.
- **E2E harness** -- the Auth0 verifier path-escapes every id (a user's id carries `|`) and lets a kind name the output its identifier lives under (`user_id` here; `id` everywhere else), with unit tests pinning both; the two `TestAuth0User_*` entrypoints.
- **Outputs conformance** -- an `Auth0User` case pins the full output shape.
- **Forge rule** -- the module-minting law gains its SaaS form: where there is no Secret to land in, the minted credential is a sensitive output set only when generated.

## Design Notes

- `password` has three honest states because the provider does: required on a database connection, refused on a passwordless one. `passwordless` is the third state, and a validation rule refuses a password beside it.
- `verify_email` is presence-tracked (`optional bool`): unset means Auth0 decides, an explicit `false` suppresses the confirmation message. The plain bools (`email_verified`, `blocked`, `phone_verified`) are plain because their false equals the provider's omitted behaviour.
- The singular companions `auth0_user_role` and `auth0_user_permission` are deliberately not modeled: they are the non-authoritative form of the two folded sets and would let dashboard edits survive an apply.
- The Auth0 provider is not instrumented for parity measurement or import recipes (no schema artifact under `pkg/providerparity/schemas/`, no `catalog/auth0/aa_import/`), so the parity claim is "built for", accounted in the guide, and the import-map phase is skipped as the forge rule directs for an uninstrumented provider.

## Validation

Spec tests 19 of 19; `go build` on the kind's packages and the release-shaped Pulumi entrypoint; `tofu validate`; `planton validate-outputs` full population, zero unmapped; outputs conformance, anatomy, catalog-page, logo, cost, controls, permissions, secret-coverage, refcheck, `validate-refs`, crkreflect snapshot, and refgen suites green; `make protos` including the Java stub compile and the protovalidate-java conformance gate; `make e2e-build`, `make e2e-vet`. Live lanes: recorded in the e2e profile's status.

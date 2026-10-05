# Every Data Pod Names Its Environment, and an Auth0 Tenant Introduces Itself

**Date**: September 28, 2026
**Type**: Feature
**Components**: Kubernetes data modules (`kubernetestemporal`, `kubernetesopenfga`, `kubernetesvalkey`, `kubernetesneo4j`, `kubernetesopenbao`, `kubernetespostgres`), `auth0client`, `auth0tenantsettings` (new), the kind enum

## Summary

Every Kubernetes data kind now stamps its resource-identity labels, including `planton.ai/organization` and `planton.ai/environment`, on every pod it runs. Before, those labels reached only the namespace and the secrets the modules create. Temporal, OpenFGA, Valkey, Neo4j, OpenBao and PostgreSQL pods carried only their charts' own labels, so a log line, a metric or an alert from a database or a job queue couldn't say which organization's environment it came from, while every `KubernetesDeployment` pod could.

An Auth0 application can now be named for people, and an Auth0 tenant can be declared to introduce itself as the product. Universal Login reads "Log in to *tenant* to continue to *application*". The tenant's half defaulted to its identifier (for example `acme-prod`) and the application's to the resource's name (for example `console`), so a stranger's first screen named internal identifiers.

## What Changed

**Pod labels, in both engines, through each chart's documented pod-label value.** The typed values carry them, so a user's `helm_values` keeps the last word:
- **Temporal 1.6.0:** `server.podLabels`, `web.podLabels`, `admintools.podLabels` and `schema.podLabels`, merged per service by the chart's `temporal.resourceLabels`.
- **OpenFGA 0.3.10:** `podExtraLabels`. The migration runs as an init container in the same pod.
- **Valkey 0.11.0:** `podLabels`.
- **Neo4j 2026.6.0:** `neo4j.labels`, on the server pod and the chart's objects.
- **OpenBao 0.28.6:** `server.extraLabels`, plus `injector.extraLabels` when the injector is on.
- **KubernetesPostgres:** the CloudNativePG Cluster's `inheritedMetadata.labels`. The operator hands them to every object it creates.

Every chart's selectors are its own fixed labels, and none of these values reaches one. `helm template` of Temporal, OpenFGA, Valkey, Neo4j and OpenBao shows every workload's pods labeled and no selector changed.

The modules' comments that said the labels were "never injected into the chart's own resources" now state the rule: pod labels are the chart's own API, and every pod names its organization and environment.

**`Auth0Client.name` (field 30).** It's the application's name as people see it: on the login page, on consent screens, in the dashboard. It defaults to `metadata.name`, so every existing client is unchanged, and it follows the family's convention (`Auth0Role.name`, `Auth0ResourceServer.name`). The resource's identity is untouched: the Pulumi resource keeps `metadata.name` as its logical name, so renaming for people never replaces the client.

**`Auth0TenantSettings` (new kind, 8007, prefix `a0tset`).** It manages four presentation settings of the EXISTING tenant the credential belongs to: `friendly_name`, `picture_url`, `support_email` and `support_url`.
- **Unset is unmanaged:** a null is never sent, and each attribute is Optional and Computed in the provider schema.
- **At least one setting is required.**
- **Destroy is the provider's no-op:** `deleteTenant` returns nil, so the last-applied values stay, following the precedent of `CloudflareZoneSettings`.
- **Scopes:** `read:tenant_settings` and `update:tenant_settings` (`iac/permissions.yaml`).
- **Proof:** the e2e profile is `pending_proof`. A live lane would rename whatever tenant it ran against, and no throwaway tenant exists yet.

## Verification

- Scoped gates pass: `go test` over the six Kubernetes modules (including OpenBao's new `TestRender_EveryServerPodNamesItsOrganizationAndEnvironment`), `catalog/auth0/...` (the new kind's spec tests), `pkg/anatomy`, `pkg/catalogbundle`, `pkg/presetvalidity`, `pkg/finops/...`, `pkg/catalogkindreflect` (snapshot updated), `shared/...`, `pkg/explain/...`, and `e2e/framework/runner` fixture integrity.
- `terraform validate` passes for both Auth0 modules.
- `make protos`, `make generate-catalog-kind-map`, `make gazelle` and `make generate-reference` were regenerated.
- Auth0Client's `variables.tf` is edited by hand for the one new field. The generator's full rewrite would have changed every unset field's default from null to `""`, a behavior change for every live client.
- **Found on the way, left for its own change:** the Pulumi engine's identity labels spell the kind, name and id keys `planton.ai/kind`, `planton.ai/name` and `planton.ai/id` (`kuberneteslabelkeys`), where the Terraform modules write `planton.ai/resource-kind`, `-name` and `-id`. `organization` and `environment` agree.

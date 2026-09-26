# Every Secret Field Takes Only a Secret Reference

**Date**: September 26, 2026
**Type**: Fix
**Components**: Catalog (Kubernetes workload environments, KubernetesSecret, Auth0Action), Guards (secret coverage), Charts (five with a secret default), Docs and skills

## Summary

A Kubernetes workload's `env.secrets[].value`, every value in a `KubernetesSecret`, and an Auth0 action's secret now carry the `sensitive` option, so on Planton each takes only a `$secret/<slug>` reference and a plain value is refused before anything is stored. The secret-coverage gate learns to recognize a secret entry's payload by the message that declares it, so the next such field cannot ship unmarked. Five charts that defaulted a password to `change-me` now default to a named secret reference. A deploy without the platform keeps taking literals everywhere.

## What Changed

- **Marked `sensitive`:**
  - `catalog/kubernetes/container_env.proto` `SecretEnvVar.value`, with its comments rewritten to show the `$secret/` form;
  - `auth0action` `Auth0ActionSecret.value`;
  - `kubernetessecret`: `opaque.data` and `opaque.binary_data` (map-level), `tls.tls_key`, `basic_auth.password`, `ssh_auth.ssh_private_key`.
  - `binary_data`'s value pattern keeps its `$secret/` arm and drops `$var/`, because a variable names a plain value.
- **The coverage gate reads a name in its message** (`pkg/secretcoverage`): `LooksSensitive(message, field)` counts a `value`, `data` or `binary_data` field of a message named with the word `Secret`, and `tlskey` joins the compound tokens. The baseline's two deferred secret-holder gaps are gone and nothing is deferred. Two true non-secrets it found are exempt with their reasons: the AWS rotation metadata value and the External Secrets template data.
- **`shared/options/options.proto`**: the option's comment names every slot shape it applies to (a string, repeated strings, map values, a `StringValueOrRef` literal).
- **Presets:** the Auth0 action's domain allowlist and audit log, and the Kubernetes Secret's opaque and TLS presets, use the `$secret/replace-with-…` convention. The TLS preset's fields are renamed to `tlsCrt`/`tlsKey`, so it validates against its own schema and leaves the preset-validity baseline. The test kind's preset and conversion corpus use a reference too.
- **Charts:**
  - `app-data-services`, `data-analytics-platform` and `software-supply-chain` (the Valkey password), `kafka-streaming-platform` (the console password) and `identity-and-access-platform` (the OpenFGA key) keep their parameter names and default to a named secret reference. Each README says how to create it and how to name an environment's own.
  - `software-supply-chain` keeps its four coupled credentials in one key-value secret referenced by key, including a new `s3_identities_config` for the object store's identities document, and its README creates all four from one generated pair.
- **Provider credentials stay unmarked, and the provider forge rule says why:** a provider config is never a stored manifest, so no reader would act on the option.
- **Docs and skills:**
  - the secrets pages, the credentials tutorial and the CLI skill use `planton secret get --reveal` for a value;
  - the config-references skill names the newly marked fields;
  - the deployment-stage page writes `env.variables` and `env.secrets` as the lists the schema defines.
- **Regenerated:** stubs, reference pages, and the proto-docs index. Two drifts already on `main` were regenerated with them: a stale DigitalOcean database-user stub comment, and the OpenTofu module's Bazel target missing its stderr-tail files.
- **`_issues/`:** the `SecretEnvVar` literal issue is closed.

## Verification

- `go test ./pkg/secretcoverage/...`: red on the new rule before the marking, naming Auth0Action, the five workload kinds' `env.secrets.value` (app, sidecars and init containers), the Kubernetes Secret payload, and the two non-secrets; green after.
- `go test ./pkg/presetvalidity/...`: pass with the TLS preset's baseline entry removed.
- The Kubernetes Secret spec suite: a secret reference in `binary_data` is accepted, and a variable reference is refused.
- The workload kinds', Auth0 action's, External Secret's, AWS secret's and conversion certification suites: pass.
- The chart validator: 18 of 18 charts pass. `make protos` (including the protovalidate conformance gate), `make generate-reference`, and `go run ./pkg/skills/defspack`: clean.

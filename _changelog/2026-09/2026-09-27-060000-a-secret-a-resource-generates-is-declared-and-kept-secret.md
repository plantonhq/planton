# A Secret a Resource Generates Is Declared, and Both Engines Keep It Secret

**Date**: September 27, 2026
**Type**: Fix
**Components**: Catalog (outputs of 48 kinds, 38 Pulumi and 12 OpenTofu modules), Module verify, Guards (secret coverage), Output capture, Reference pages, Docs and skills

## Summary

An output that is a secret the resource generates (an Auth0 client's secret, an AWS IAM user's secret access key, a registry's admin password, a database's connection string) is now declared in its kind's schema with the `sensitive` option, on every kind that has one. `planton module verify` holds both engines to that declaration: a marked output is exported as a secret (OpenTofu `sensitive = true`, Pulumi `pulumi.ToSecret`), and no other output is, so an engine never prints a generated credential in its deploy logs or keeps it readable in state. The CLI's captured outputs read which outputs are secrets from the same schema, and the reference pages mark them. On Planton, such an output is stored in the organization's secret store and the output holds a reference.

## What Changed

- **Marked `sensitive` (68 fields in 34 kinds):**
  - Auth0: client, resource server, user;
  - AWS: IAM user, Cognito user-pool client, CodeBuild webhook;
  - Azure: 39 fields (AKS kubeconfig, ACR admin password, storage, Cosmos DB, Redis, managed Redis, Log Analytics, Cognitive, Search keys and connection strings, ExpressRoute keys, web and function app site passwords, storage local user password);
  - DigitalOcean: 16 fields (cluster kubeconfig, registry credentials, database cluster, pool, replica and user secrets, Spaces key);
  - GCP: service account key, API key, OAuth client secret, Redis auth string, Eventarc activation token;
  - Cloudflare: Zero Trust application SaaS client secret, origin CA private key.
  - The principle: generated material that on its own authenticates, authorizes or decrypts, and is not meant for untrusted clients. CA certificates, public signing certificates, identifiers, DNS validation values, and the App Insights and Firebase keys their vendors call public stay unmarked.
- **Unmarked (not secrets):** the web analytics site token and snippet (they ship inside public pages), the flex function app's site credential name (a username).
- **Engines:**
  - Pulumi exports every marked output directly through `pulumi.ToSecret` (89 exports, 33 of them on kinds already marked but exported in the clear), and unwraps the public outputs a provider marks secret with `pulumi.Unsecret`.
  - OpenTofu already declared every secret `sensitive`. Its 17 over-marked public outputs drop the flag, and the 13 whose provider attribute is sensitive unwrap with `nonsensitive()`.
- **`planton module verify`** gains the secret-outputs check in both engines, at error severity for a secret exported in the clear, and a Pulumi export-name check to match OpenTofu's. Output names resolve from literals and the module's own constants. A whole-catalog test runs both checks on every official module with no baseline, and the secret-coverage workflow runs it.
- **`pkg/secretcoverage`** pins the shape of output marks: only on a top-level `Outputs` field, no exemption reason, no value rule. The name heuristic stays on the spec, because a name cannot tell a key id from a private key.
- **Output capture reads the schema:** `outputs.SecretOutputs(kind)` is the one rule, `CaptureResult.IsSensitive` masks a secret output and any output the schema does not declare, Pulumi's second (masked) output read is gone, and a value that fails to populate is no longer logged.
- **Reference pages** mark a secret output `(sensitive)` in the Outputs table and say where it lives on Planton; `options.proto` documents the option on outputs.
- **Docs and skills:** "Secrets a Resource Generates" on the secrets page; infra-job output display; the dependencies and config-references skill pages (a secret output feeds only a sensitive field); the catalog skill's research recipes.
- **Regenerated:** stubs, reference pages, the proto-docs index, catalog schemas.

## Verification

- The whole-catalog secret-outputs test: red with 132 findings before the catalog change, green after.
- `go test` for `pkg/iac/moduleverify`, `pkg/secretcoverage`, `pkg/outputs` (red by mutation on the undeclared-key rule), `pkg/explain`, `pkg/explain/refgen`, `pkg/setdeploy`, `internal/cli/ui` and the OpenTofu capture: pass.
- Every touched kind's tests pass; every touched Pulumi module builds and vets; every touched OpenTofu module validates, with a missing `nonsensitive()` confirmed to fail validation.
- `make generate-reference`, `make build-catalog-schemas` and `make verify-catalog-schemas`, `go run ./pkg/skills/defspack`: clean.

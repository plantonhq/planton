# One Slug Rule: Lowercase Letters and Digits Joined by Single Hyphens

**Date**: September 27, 2026
**Type**: Fix
**Components**: Shared metadata (`CloudResourceMetadata.slug`), Manifest graph (slug derivation), Set deploy (node workspaces), Reference pages, Docs and skills

## Summary

A slug is lowercase letters and digits joined by single hyphens, like my-app-2. That one rule now holds everywhere a cloud resource's slug is read or made. `CloudResourceMetadata.slug` carries it as a protovalidate rule, the same id, message and expression as the platform's `ApiResourceMetadata.slug`, so a manifest with a slug like `db_main` or `example.com` is refused with that sentence before anything deploys. The CLI's name-to-slug derivation now matches the platform's byte for byte: underscores and dots become hyphens (`DB_PASSWORD` is `db-password`, `example.com` is `example-com`), and accents are dropped instead of splitting a word (`Café Résumé` is `cafe-resume`). A name stays free text; the slug is the handle. The alphabet is the one every system a slug is written into accepts: DNS labels, secret store names, cloud labels and tags, and dot-delimited identities such as the Pulumi stack `<env>.<Kind>.<slug>`.

## What Changed

- **`shared/metadata.proto`:** `slug` (field 2) gets the `metadata.slug` CEL rule `this == '' || this.matches('^[a-z0-9]+(-[a-z0-9]+)*$')`. An empty slug stays legal, because the platform derives it from the name. The field comment says what a slug is for and why it has this alphabet.
- **`manifestgraph.GenerateSlug`** is rebuilt as a mirror of the platform's `ApiRequestResourceSlugGenerator`. It NFD-normalizes, lowercases, drops combining marks, turns every run of characters outside `[a-z0-9]` into one hyphen, and trims the ends. The old version kept underscores (`a__b` stayed `a__b`), so the CLI predicted a different slug than the server stored. `slug_test.go` copies the Java test table, including the check that every generated slug is lawful.
- **`manifestgraph.IsLawfulSlug`, `SlugPattern`, `SlugRule`:** the rule as Go, spelled the same as the proto.
- **Set deploy:** a node's workspace path (`~/.planton/setdeploy/<env>/<kind>/<slug>/`) refuses a slug or env that is not lawful rather than joining it into a filesystem path.
- **Docs and skills:** a new "Slugs" section on the resource hierarchy page. The one sentence replaces "lowercase with hyphens", "lowercase alphanumeric with hyphens" and "DNS-compatible (`[a-z0-9-]`)" on the getting-started, resource-hierarchy and cloud-providers pages, the config-references skill page and the catalog's shared metadata reference. The dependencies skill page says reference names fold dots and underscores into hyphens. The example name of a secret a resource generates is updated to the platform's `<kind>-outputs-<resource>` (`$secret/@prod/auth0-client-outputs-checkout/client_secret`).
- **Regenerated:** stubs, the proto-docs index, reference pages.

## Verification

- `go test ./pkg/manifestgraph/... ./pkg/setdeploy/...`: pass. This includes the copied Java table and a new test that the set-deploy workspace refuses `../escape`, `a_b`, `example.com` and an unlawful env, each time naming the rule.
- `go build` of `./cmd/...`, `./internal/...`, `./shared/...`, `./e2e/framework/...` and the touched packages, and `go vet` on the touched packages: clean. The tests of every package that imports `manifestgraph` (`pkg/infrachart`, `e2e/framework/runner`) and of `pkg/explain/refgen` and `pkg/protodocs` pass.
- `make protos`, including the Java stub compile and the protovalidate-java conformance gate, which compiles the new rule on the platform's engine.
- `make generate-reference`, `make build-catalog-schemas` and `make verify-catalog-schemas`, `go run ./pkg/skills/defspack`: clean.
- No catalog preset, e2e fixture, chart or test manifest sets an explicit `metadata.slug` that breaks the rule.

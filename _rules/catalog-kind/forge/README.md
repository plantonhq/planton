# Forge: Create Kinds

## Overview

**Forge** is the rule system for bootstrapping **complete, production-ready kinds** in Planton. It orchestrates 20 atomic rules that create everything from proto definitions to IaC modules to documentation and presets.

**Key principle:** Forge creates kinds that match **95-100% of the ideal state** defined in `architecture/catalog-kind.md`, and whose file layout passes the machine-enforced anatomy gate (`pkg/anatomy`, CI lane `lint.kind-anatomy.yaml`).

## Kind Anatomy

A kind lives at `catalog/{provider}/{kind}/`:

- **Kind root (the living kind):** `README.md` (GitHub-facing page), `catalog.md` (the catalog page), `logo.svg`, optional `GUIDE.md` (authored operational judgment), `iac/pulumi/` + `iac/tf/` (each with a README.md; optional `iac/import-map.yaml`; a module is its directory and reads nothing outside it), `presets/` (`.yaml` + `.md` sidecar pairs), `e2e/` (manifest.yaml, profile, scenarios), optional `conversions/`.
- **Version dir `v1alpha1/` (the versioned contract ONLY):** `api.proto`, `spec.proto`, `input.proto`, `outputs.proto`, their `.pb.go` stubs, `BUILD.bazel`, `spec_test.go`, `reference.md`.

The proto FILE names are `input.proto` / `outputs.proto`; the MESSAGE names are `{Kind}IacInput` / `{Kind}Outputs` — the message names are the identity every downstream consumer keys on.

## What Forge Creates

When you run forge, you get a fully-implemented kind:

### Proto API Definitions (v1alpha1/)
- ✅ `spec.proto` - Configuration schema with field validations
- ✅ `input.proto` - Inputs to IaC modules (`{Kind}IacInput`: target + provider config)
- ✅ `outputs.proto` - Deployment outputs (`{Kind}Outputs`)
- ✅ `api.proto` - KRM wiring (apiVersion, kind, metadata, spec, status)
- ✅ Generated `.pb.go` stubs for all proto files
- ✅ `spec_test.go` - Unit tests for ALL validation rules
- ✅ **Tests execute and pass** - Validates buf.validate rules work correctly

### IaC Modules - Pulumi (iac/pulumi/)
- ✅ Module files: `module/main.go`, `module/locals.go`, `module/outputs.go`, resource-specific files
- ✅ Entrypoint: `main.go` (`package main`) and `Pulumi.yaml` at the `iac/pulumi/` root — nothing else
- ✅ Documentation: `README.md`

### IaC Modules - Terraform (iac/tf/)
- ✅ Module files: `variables.tf` (generated), `provider.tf`, `locals.tf`, `main.tf`, `outputs.tf`
- ✅ Documentation: `README.md`
- ✅ **100% behavioral parity** with the Pulumi module (Parity Mandate)

### Documentation
- ✅ Kind-root `README.md` - User-facing overview (the GitHub-facing kind page)
- ✅ Kind-root `GUIDE.md` - Authored operational judgment: recipes, design rationale, parity accounting
- ✅ Generated `reference.md` - Regenerated via `make generate-reference`, embedding the e2e manifest as its Example

### Supporting Files
- ✅ Kind-root `e2e/manifest.yaml` - The complete, protovalidate-valid example manifest; the reference page's Example block AND the E2E framework's testability marker and deployed fixture
- ✅ Kind-root `presets/` - 2-3 ready-to-deploy configuration templates with `.md` sidecars
- ✅ Enum entry in `catalog_kind.proto`
- ✅ Build validation passed
- ✅ Test validation passed

Layout conformance is machine-checked: run the anatomy gate (`go test ./pkg/anatomy/...`) instead of auditing file inventories by hand.

## When to Use Forge

Use forge when you need to:
- ✅ **Bootstrap a new kind from scratch**
- ✅ Add support for a new cloud provider resource
- ✅ Add support for a new SaaS platform resource
- ✅ Add a new Kubernetes workload or addon

**Don't use forge when:**
- ❌ Kind already exists (use **update** instead)
- ❌ You only need to fix/enhance existing kind (use **update**)
- ❌ You want to remove a kind (use **delete**)
- ❌ You want to check completion status (use **audit**)

## How to Use Forge

### Basic Usage

```
@forge-catalog-kind <KindName> --provider <provider>
```

### Examples

**Create a SaaS platform resource:**
```
@forge-catalog-kind CloudflareD1Database --provider cloudflare
```

**Create a GCP resource:**
```
@forge-catalog-kind GcpStorageBucket --provider gcp
```

**Create an AWS resource:**
```
@forge-catalog-kind AwsSqsQueue --provider aws
```

**Create a Kubernetes workload:**
```
@forge-catalog-kind PostgresKubernetes --provider kubernetes --category workload
```

**Create a Kubernetes addon:**
```
@forge-catalog-kind CertManagerKubernetes --provider kubernetes --category addon
```

### Required Information

Before running forge, have ready:
1. **Kind Name** - PascalCase (e.g., `GcpCertManagerCert`)
2. **Provider** - One of: aws, gcp, azure, kubernetes, digitalocean, cloudflare, auth0, openfga
3. **Category** - Only for Kubernetes: addon, workload, or config

### What Forge Asks You

Forge will interview you to gather:
- Kind purpose and use case
- Key configuration fields (for spec.proto)
- Expected outputs (for outputs.proto)
- Provider-specific details (project IDs, regions, etc.)
- Credential requirements
- Best practices and gotchas

## The 20-Rule Workflow

Forge orchestrates 20 rules in 9 phases. (E2E execution is handled separately via
catalog kind E2E profiles, not the forge pipeline.)

### Phase 1: Proto API Definitions
1. `001-spec-proto` - Generate spec.proto
2. `002-spec-validate` - Add validations
3. `003-spec-tests` - Generate tests
4. `004-outputs` - Generate outputs.proto
5. `005-api` - Generate api.proto
6. `006-input` - Generate input.proto

### Phase 2: Registration
7. `014-catalog-kind` - Register enum
8. `015-generate-proto-stubs` - Generate .pb.go files

### Phase 3: Documentation
9. `007-readme` - Generate the kind-root README.md

### Phase 4: Test Infrastructure
10. `008-e2e-manifest` - Generate the kind-root e2e/manifest.yaml

### Phase 5: Pulumi Implementation
11. `009-pulumi-module` - Generate module
12. `010-pulumi-entrypoint` - Generate entrypoint (main.go, Pulumi.yaml)
13. `011-pulumi-readme` - Generate docs

### Phase 6: Terraform Implementation
14. `012-terraform-module` - Generate module
15. `013-terraform-readme` - Generate docs

### Phase 7: Presets
16. `018-presets` - Generate initial presets (2-3 common configuration templates)

### Phase 8: Final Validation
17. `016-build-validation` - Compile all Go code (recursive kind build + release-equivalent entrypoint build)
18. `017-test-validation` - Run all tests

### Phase 9: Catalog Page + Guide + Reference Regeneration
19. `020-catalog-md` - Write the kind-root `catalog.md` (THE catalog page) per the catalog page standard — runs first in this phase because the guide grounds its claims partly on the catalog page
20. `019-guide` - Seed the kind-root `GUIDE.md` (authored wisdom), then run `make generate-reference` so the reference page, guide head link, and catalog indexes materialize together

### Phase 10: Logo
21. `021-logo` - Author the kind-root `logo.svg` under the logo law (one kind, one glyph: the provider's official mark only when the kind IS the product, the Planton brand mark on Planton's own kinds, a Planton-drawn glyph with its provenance `<desc>` everywhere else), then run the catalog-logo gate

Rule numbers are stable identities, not execution positions — `020` executes before `019` here, just as `014`/`015` execute in Phase 2.

## Progress Tracking

Forge provides real-time progress updates:

```
🔨 Forge: Creating CloudflareD1Database

Phase 1: Proto API Definitions
[1/21] ✅ Generated spec.proto
[2/21] ✅ Added buf.validate rules
[3/21] ✅ Generated and ran spec tests
[4/21] ✅ Generated outputs.proto
[5/21] ✅ Generated api.proto
[6/21] ✅ Generated input.proto

Phase 2: Registration
[7/21] ✅ Registered CloudflareD1Database = 7005 in catalog_kind.proto
[8/21] ✅ Generated proto stubs (make protos)

Phase 3: Documentation
[9/21] ✅ Generated kind-root README.md

Phase 4: Test Infrastructure
[10/21] ✅ Generated e2e/manifest.yaml

Phase 5: Pulumi Implementation
[11/21] ✅ Generated Pulumi module
[12/21] ✅ Generated Pulumi entrypoint
[13/21] ✅ Generated Pulumi docs

Phase 6: Terraform Implementation
[14/21] ✅ Generated Terraform module
[15/21] ✅ Generated Terraform docs

Phase 7: Presets
[16/21] ✅ Generated initial presets

Phase 8: Final Validation
[17/21] ✅ Build validation passed (go build ./catalog/<provider>/<kind>/...)
[18/21] ✅ Kind tests passed (go test -v ./catalog/<provider>/<kind>/v1alpha1/)

Phase 9: Catalog Page + Guide + Reference
[19/21] ✅ Wrote catalog.md (the catalog page)
[20/21] ✅ Seeded GUIDE.md and regenerated the reference

Phase 10: Logo
[21/21] ✅ Authored logo.svg (catalog-logo gate green, no baseline entry)

🎉 Kind creation complete!

📍 Location: catalog/cloudflare/cloudflared1database/
📊 Expected Audit Score: 95-100%

Next steps:
1. Review generated files
2. Run: @audit-catalog-kind CloudflareD1Database
3. Make any custom modifications
4. Commit and push
```

## Error Handling

### Automatic Retries
- Each rule retries up to 3 times on fixable errors
- Build errors are fixed automatically when possible
- Test failures trigger fixes and retries

### Manual Intervention
If a rule fails after 3 attempts:
1. Forge stops and shows the error
2. Fix the issue manually
3. Resume from the failed rule:
   ```
   @forge-catalog-kind CloudflareD1Database --resume-from 012
   ```

### Common Issues

**Issue: Proto build fails**
- **Cause:** Invalid protobuf syntax
- **Fix:** Forge auto-fixes and retries
- **If persists:** Check .proto file manually

**Issue: Pulumi/Terraform E2E fails**
- **Cause:** Missing credentials or invalid config
- **Fix:** Check manifest values, update and retry

**Issue: Tests fail**
- **Cause:** Validation rules too strict or test logic error
- **Fix:** Forge analyzes and fixes tests automatically

## Post-Forge Validation

After forge completes, validate the kind:

**Option 1: Manual Audit**
```bash
@audit-catalog-kind <KindName>
```
**Expected Result:** 95-100% completion score

If score is lower, the audit report shows what's missing, why it matters, and how to fix it.

**Option 2: Auto-Complete (Recommended)**
```bash
@complete-catalog-kind <KindName>
```
Automatically audits and fills any remaining gaps to reach 95%+. Useful if forge had partial failures.

**Always:** the anatomy gate (`go test ./pkg/anatomy/...`) must be green — it machine-checks the kind's file layout against the canonical anatomy.

## Customization After Forge

Forge creates a **production-ready baseline**. Common customizations:

### Add More Fields to Proto
1. Edit `spec.proto` to add fields
2. Update validations in `spec.proto`
3. Update tests in `spec_test.go`
4. Run `make protos` to regenerate stubs
5. Update Pulumi module to use new fields
6. Regenerate Terraform `variables.tf` and update the module to match
7. Update the e2e manifest and presets if the new fields belong in examples
8. Run `go build ./catalog/<provider>/<kind>/... && go test -v ./catalog/<provider>/<kind>/v1alpha1/`

### Modify IaC Implementation
1. Update Pulumi module files (`iac/pulumi/module/*.go`)
2. Update Terraform module files (`iac/tf/*.tf`)
3. Keep both engines at behavioral parity (fix whichever engine is wrong)
4. Update documentation if behavior changes

### Enhance Documentation
1. Expand the kind-root `README.md` and `catalog.md`
2. Capture new operational judgment in `GUIDE.md`
3. Add troubleshooting to `iac/pulumi/README.md` or `iac/tf/README.md`

## Comparison to Manual Creation

| Aspect | Manual Creation | Forge |
|--------|----------------|-------|
| Time | 8-16 hours | 15-30 minutes |
| Completeness | 60-80% typical | 95-100% |
| Documentation | Often skipped | Comprehensive |
| Validation | Manual | Automated |
| Consistency | Varies | Standardized |
| Best Practices | Hit or miss | Built-in |
| Error-Prone | Yes | Auto-fixed |

## Reference Documents

- **Ideal State Definition:** `architecture/catalog-kind.md`
- **Anatomy Conformance Gate:** `pkg/anatomy` (CI lane `.github/workflows/lint.kind-anatomy.yaml`)
- **Individual Flow Rules:** `_rules/catalog-kind/forge/flow/`
- **Main Orchestrator:** `_rules/catalog-kind/forge/forge-catalog-kind.mdc`

## Tips and Best Practices

### Before Running Forge

1. **Research the resource** - Understand what you're creating
2. **Check if it exists** - Run `@audit-catalog-kind` first
3. **Plan your API** - Read the provider schema at the pinned version fully and model it fully (100% of the configurable surface)
4. **Gather examples** - Have reference configurations ready

### During Forge

1. **Be specific** - Provide detailed answers to interview questions
2. **Think production** - Consider real-world use cases
3. **Include validation** - Think about what constraints make sense
4. **Document gotchas** - Share known issues and workarounds

### After Forge

1. **Review everything** - Don't blindly trust generated code
2. **Test locally** - Deploy with the e2e manifest
3. **Enhance docs** - Add your learnings to the surfaces their readers use (spec comments, GUIDE.md, presets)
4. **Run audit** - Verify 100% ideal state compliance

## Troubleshooting

### "Kind already exists"
**Error:** `Kind CloudflareD1Database already exists at ...`

**Solution:** Use `@update-catalog-kind` instead, or delete first with `@delete-catalog-kind`.

### "Provider not recognized"
**Error:** `Provider 'xyz' is not valid`

**Valid providers:** aws, gcp, azure, kubernetes, digitalocean, cloudflare, auth0, openfga

### "Build failed after 3 attempts"
**Check:**
1. Proto syntax in generated files
2. Go code compiles: `go build ./catalog/<provider>/<kind>/...` (from the repo root)
3. Import paths are correct
4. Manual fix may be needed

## Next Steps

After reading this README:
1. Review the ideal state document: `architecture/catalog-kind.md`
2. Try forge on a test kind
3. Inspect the generated code
4. Run audit to verify completion
5. Use forge for real kinds!

---

**Questions?** Check the troubleshooting section or run `@audit-catalog-kind` to see examples of complete kinds.

**Ready to create?** Run `@forge-catalog-kind <YourKindName> --provider <provider>`

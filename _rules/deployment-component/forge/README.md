# Forge: Create Deployment Components

## Overview

**Forge** is the rule system for bootstrapping **complete, production-ready deployment components** in OpenMCF. It orchestrates 21 atomic rules that create everything from proto definitions to IaC modules to comprehensive documentation.

**Key principle:** Forge creates components that match **95-100% of the ideal state** defined in `architecture/deployment-component.md`.

## What Forge Creates

When you run forge, you get a fully-implemented deployment component:

### Proto API Definitions
- ✅ `spec.proto` - Configuration schema with field validations
- ✅ `stack_input.proto` - Inputs to IaC modules (spec + credentials + context)
- ✅ `stack_outputs.proto` - Deployment outputs
- ✅ `api.proto` - KRM wiring (apiVersion, kind, metadata, spec, status)
- ✅ Generated `.pb.go` stubs for all proto files
- ✅ `spec_test.go` - Unit tests for ALL validation rules
- ✅ **Tests execute and pass** - Validates buf.validate rules work correctly

### IaC Modules - Pulumi
- ✅ Module files: `main.go`, `locals.go`, `outputs.go`, resource-specific files
- ✅ Entrypoint: `main.go`, `Pulumi.yaml`, `Makefile`
- ✅ Documentation: `README.md`, `overview.md`, `debug.sh`
- ✅ E2E tested and validated

### IaC Modules - Terraform
- ✅ Module files: `variables.tf`, `provider.tf`, `locals.tf`, `main.tf`, `outputs.tf`
- ✅ Documentation: `README.md`
- ✅ E2E tested and validated
- ✅ Feature parity with Pulumi module

### Documentation
- ✅ `v1/README.md` - User-facing overview (50-200 lines)
- ✅ `v1/examples.md` - Copy-paste ready examples (multiple use cases)
- ✅ `v1/docs/README.md` - **Comprehensive research document** (300-1000+ lines)
  - Deployment landscape analysis
  - Method comparisons
  - Best practices
  - 80/20 scoping rationale

### Supporting Files
- ✅ `iac/hack/manifest.yaml` - Test manifest
- ✅ Enum entry in `cloud_resource_kind.proto`
- ✅ Build validation passed
- ✅ Test validation passed

## When to Use Forge

Use forge when you need to:
- ✅ **Bootstrap a new deployment component from scratch**
- ✅ Add support for a new cloud provider resource
- ✅ Add support for a new SaaS platform resource
- ✅ Add a new Kubernetes workload or addon

**Don't use forge when:**
- ❌ Component already exists (use **update** instead)
- ❌ You only need to fix/enhance existing component (use **update**)
- ❌ You want to remove a component (use **delete**)
- ❌ You want to check completion status (use **audit**)

## How to Use Forge

### Basic Usage

```
@forge-openmcf-component <ComponentName> --provider <provider>
```

### Examples

**Create a SaaS platform resource:**
```
@forge-openmcf-component MongodbAtlas --provider atlas
```

**Create a GCP resource:**
```
@forge-openmcf-component GcpStorageBucket --provider gcp
```

**Create an AWS resource:**
```
@forge-openmcf-component AwsSqsQueue --provider aws
```

**Create a Kubernetes workload:**
```
@forge-openmcf-component PostgresKubernetes --provider kubernetes --category workload
```

**Create a Kubernetes addon:**
```
@forge-openmcf-component CertManagerKubernetes --provider kubernetes --category addon
```

### Required Information

Before running forge, have ready:
1. **Component Name** - PascalCase (e.g., `GcpCertManagerCert`)
2. **Provider** - One of: aws, gcp, azure, kubernetes, atlas, snowflake, confluent, digitalocean, civo, cloudflare
3. **Category** - Only for Kubernetes: addon, workload, or config

### What Forge Asks You

Forge will interview you to gather:
- Component purpose and use case
- Key configuration fields (for spec.proto)
- Expected outputs (for stack_outputs.proto)
- Provider-specific details (project IDs, regions, etc.)
- Credential requirements
- Best practices and gotchas

## The 21-Rule Workflow

Forge orchestrates 21 rules in 7 phases:

### Phase 1: Proto API Definitions
1. `001-spec-proto` - Generate spec.proto
2. `002-spec-validate` - Add validations
3. `003-spec-tests` - Generate tests
4. `004-stack-outputs` - Generate outputs proto
5. `005-api` - Generate api.proto
6. `006-stack-input` - Generate input proto

### Phase 2: Registration
7. `016-cloud-resource-kind` - Register enum
8. `017-generate-proto-stubs` - Generate .pb.go files

### Phase 3: Documentation
9. `007-docs` - Generate README and examples
10. `020-research-docs` - Generate research document

### Phase 4: Test Infrastructure
11. `008-hack-manifest` - Generate test manifest

### Phase 5: Pulumi Implementation
12. `009-pulumi-module` - Generate module
13. `010-pulumi-entrypoint` - Generate entrypoint
14. `011-pulumi-e2e` - Run E2E test
15. `012-pulumi-docs` - Generate docs
16. `021-pulumi-overview` - Generate architecture overview

### Phase 6: Terraform Implementation
17. `013-terraform-module` - Generate module
18. `014-terraform-e2e` - Run E2E test
19. `015-terraform-docs` - Generate docs

### Phase 7: Final Validation
20. `018-build-validation` - Compile all Go code
21. `019-test-validation` - Run all tests

## Progress Tracking

Forge provides real-time progress updates:

```
🔨 Forge: Creating MongodbAtlas

Phase 1: Proto API Definitions
[1/21] ✅ Generated spec.proto
[2/21] ✅ Added buf.validate rules
[3/21] ✅ Generated and ran spec tests
[4/21] ✅ Generated stack_outputs.proto
[5/21] ✅ Generated api.proto
[6/21] ✅ Generated stack_input.proto

Phase 2: Registration
[7/21] ✅ Registered MongodbAtlas = 51 in cloud_resource_kind.proto
[8/21] ✅ Generated proto stubs (make protos)

Phase 3: Documentation
[9/21] ✅ Generated v1/README.md and examples.md
[10/21] ✅ Generated v1/docs/README.md (research document)

Phase 4: Test Infrastructure
[11/21] ✅ Generated iac/hack/manifest.yaml

Phase 5: Pulumi Implementation
[12/21] ✅ Generated Pulumi module
[13/21] ✅ Generated Pulumi entrypoint
[14/21] ✅ Passed Pulumi E2E test
[15/21] ✅ Generated Pulumi docs
[16/21] ✅ Generated Pulumi overview

Phase 6: Terraform Implementation
[17/21] ✅ Generated Terraform module
[18/21] ✅ Passed Terraform E2E test
[19/21] ✅ Generated Terraform docs

Phase 7: Final Validation
[20/21] ✅ Build validation passed (go build ./apis/.../v1/...)
[21/21] ✅ Component tests passed (go test -v ./apis/.../v1/)

🎉 Component creation complete!

📍 Location: apis/org/openmcf/provider/atlas/mongodbatlas/v1/
📊 Expected Audit Score: 95-100%

Next steps:
1. Review generated files
2. Run: @audit-openmcf-component MongodbAtlas
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
   @forge-openmcf-component MongodbAtlas --resume-from 012
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

After forge completes, validate the component:

**Option 1: Manual Audit**
```bash
@audit-openmcf-component <ComponentName>
```
**Expected Result:** 95-100% completion score

If score is lower, the audit report shows what's missing, why it matters, and how to fix it.

**Option 2: Auto-Complete (Recommended)**
```bash
@complete-openmcf-component <ComponentName>
```
Automatically audits and fills any remaining gaps to reach 95%+. Useful if forge had partial failures.

## Customization After Forge

Forge creates a **production-ready baseline**. Common customizations:

### Add More Fields to Proto
1. Edit `spec.proto` to add fields
2. Update validations in `spec.proto`
3. Update tests in `spec_test.go`
4. Run `make protos` to regenerate stubs
5. Update Pulumi module to use new fields
6. Update Terraform `variables.tf` to match
7. Update examples in `examples.md`
8. Run `go build ./apis/.../v1/... && go test -v ./apis/.../v1/`

### Modify IaC Implementation
1. Update Pulumi module files (`iac/pulumi/module/*.go`)
2. Update Terraform module files (`iac/tf/*.tf`)
3. Test with `@forge-openmcf-component <Name> --test-only`
4. Update documentation if behavior changes

### Enhance Documentation
1. Add more examples to `examples.md`
2. Expand research in `docs/README.md`
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

- **Ideal State Definition:** `architecture/deployment-component.md`
- **Individual Flow Rules:** `_rules/deployment-component/forge/flow/`
- **Forge Analysis:** `_rules/deployment-component/forge/FORGE_ANALYSIS.md`
- **Main Orchestrator:** `_rules/deployment-component/forge/forge-openmcf-component.mdc`

## Tips and Best Practices

### Before Running Forge

1. **Research the resource** - Understand what you're creating
2. **Check if it exists** - Run `@audit-openmcf-component` first
3. **Plan your API** - Know which fields are essential (80/20)
4. **Gather examples** - Have reference configurations ready

### During Forge

1. **Be specific** - Provide detailed answers to interview questions
2. **Think production** - Consider real-world use cases
3. **Include validation** - Think about what constraints make sense
4. **Document gotchas** - Share known issues and workarounds

### After Forge

1. **Review everything** - Don't blindly trust generated code
2. **Test locally** - Deploy with the test manifest
3. **Enhance docs** - Add your learnings to documentation
4. **Run audit** - Verify 100% ideal state compliance

## Troubleshooting

### "Component already exists"
**Error:** `Component MongodbAtlas already exists at ...`

**Solution:** Use `@update-openmcf-component` instead, or delete first with `@delete-openmcf-component`.

### "Provider not recognized"
**Error:** `Provider 'xyz' is not valid`

**Valid providers:** aws, gcp, azure, kubernetes, atlas, snowflake, confluent, digitalocean, civo, cloudflare

### "Build failed after 3 attempts"
**Check:**
1. Proto syntax in generated files
2. Go code compiles: `cd apis/... && go build`
3. Import paths are correct
4. Manual fix may be needed

## Success Stories

**Before Forge:**
- Creating GcpCertManagerCert took 12 hours
- Documentation was incomplete
- Tests were basic
- Terraform module was added 3 months later

**After Forge:**
- Creating new components takes 20-30 minutes
- Documentation is comprehensive on day 1
- All tests pass immediately
- Both IaC modules created together

## Next Steps

After reading this README:
1. Review the ideal state document: `architecture/deployment-component.md`
2. Try forge on a test component
3. Inspect the generated code
4. Run audit to verify completion
5. Use forge for real components!

---

**Questions?** Check the troubleshooting section or run `@audit-openmcf-component` to see examples of complete components.

**Ready to create?** Run `@forge-openmcf-component <YourComponentName> --provider <provider>`


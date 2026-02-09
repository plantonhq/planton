# Terraform Module Authoring Guide

Purpose: implement the Terraform module under `iac/tf/` for a resource kind using multi-file layout.

## Inputs to read
- `api.proto`, `spec.proto`, `stack_input.proto`, `stack_outputs.proto`

## Target directory
- `apis/project/planton/provider/<provider>/<kindfolder>/v1/iac/tf/`

## Files (typical)
- `variables.tf` — generated via CLI (do not hand-edit)
- `provider.tf` — required_providers and minimal provider config
- `locals.tf` — safe_* locals, computed booleans, derived values
- `outputs.tf` — map to `<Kind>StackOutputs`
- Concern files: `security_group.tf`, `instance.tf`, `dns.tf`, `iam.tf`, `data.tf`, etc.
- Optional `main.tf` (keep minimal) and optional nested `modules/`

## Handling Optional Fields with Defaults

Fields marked `optional` with `(org.openmcf.shared.options.default)` in `spec.proto` are guaranteed to be populated by OpenMCF middleware before Terraform runs. This has important implications:

### What This Means for Terraform Modules

- The generated `variables.tf` may show `optional(string, null)` for these fields
- At runtime, the value will **never** be null for fields that have defaults in the proto schema
- You do NOT need defensive handling (`coalesce()`, `try()`, conditional expressions) for defaulted fields

### Correct Pattern

```hcl
# In locals.tf - use directly, no defensive wrapper needed
locals {
  content_type = var.spec.content_type  # Guaranteed non-null by OpenMCF middleware
}

# In resource definitions - pass through directly
resource "aws_s3_object" "this" {
  content_type = each.value.content_type  # Always populated
}
```

### Anti-Pattern

```hcl
# WRONG: Redundant defensive default - OpenMCF already handled this
locals {
  content_type = coalesce(var.spec.content_type, "application/octet-stream")
}
```

### When Defensive Handling IS Needed

Only use `coalesce()` / `try()` for fields that are truly optional without a proto-level default (i.e., fields that are plain `string` or `int32` without the `optional` keyword and default option). These fields may legitimately be empty/zero at runtime.

## Notes
- Use the CLI generator for variables.tf; derive convenience values in locals.tf.
- Split by concern; Terraform builds the graph from all files automatically.

# GcpModelArmorTemplate — Pulumi Implementation

This directory contains the Pulumi implementation for a Model Armor template from the Planton spec: one `gcp.projects.Service` (API enablement) and one `gcp.modelarmor.Template`.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `template` |
| `module/locals.go` | The `planton-ai_*` attribution labels |
| `module/template.go` | Enables the API; maps the filter configuration and metadata; exports the outputs |
| `module/outputs.go` | Output key constants (`name`, `template_id`, `location`) |

## Send Posture (parity with Terraform)

- **`TemplateId`** -- `spec.template_id`, defaulting to `metadata.name`.
- **Filter settings** -- each emitted only when its spec message is set; optional strings only when set.
- **`TemplateMetadata`** -- emitted only when set; bools as declared, codes and strings only when set; `MultiLanguageDetection` only when the lifted bool is true.
- **`DeletionPolicy`** -- sent only when set.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```

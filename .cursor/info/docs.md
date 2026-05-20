# Build and Docs Authoring Guide

Purpose: produce `README.md` for a resource after its protos exist.

## Inputs to read
- `api.proto` — envelope (api_version/kind/metadata/spec/status)
- `spec.proto` — fields and brief descriptions
- `stack_input.proto` — provisioner/IaC/target/credential
- `stack_outputs.proto` — outputs fields
- Credential proto (for provider) — name only for documentation

## README.md (outline)
- Title: `<Kind>`
- Intro: what the resource is and what it provisions
- Spec fields (80/20 summary): key fields with one-line descriptions
- Stack outputs summary
- How it works: Pulumi and Terraform modules in this repo
- References: official docs URLs

## Notes
- Keep concise and practical (80/20). Avoid secrets in examples.
- Usage examples belong in presets (v1/presets/), not in a separate examples.md file.

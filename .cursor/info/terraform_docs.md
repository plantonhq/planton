# Terraform Docs Authoring Guide

Purpose: author `README.md` for the Terraform module under `iac/tf/`.

## Inputs to read
- `iac/tf/main.tf` — resource definitions
- `iac/tf/variables.tf` — input variables
- `iac/hack/manifest.yaml` — test manifest

## Outputs to produce
- `README.md` — CLI flows using OpenMCF tofu with default local backend

## Notes
- Keep README concise — focus on how to run and test locally.
- Usage examples belong in presets (v1/presets/), not in a separate examples.md file.
- For remote/persistent state, teams can switch to an S3/Dynamo backend; not covered here.

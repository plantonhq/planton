# Pulumi Docs Authoring Guide

Purpose: author `README.md` and `debug.sh` for the Pulumi package under `iac/pulumi/`.

## Inputs to read
- `iac/pulumi/module/main.go` — controller logic
- `iac/pulumi/module/locals.go` — computed values
- `iac/hack/manifest.yaml` — test manifest

## Outputs to produce
- `README.md` — overview of the Pulumi program and how it wires the resource
- `debug.sh` — helper script to run the Pulumi program locally (optional binary mode)

## Notes
- Keep README concise — focus on how to run and test locally.
- Usage examples belong in presets (v1/presets/), not in a separate examples.md file.

# Pulumi Module: GcpVertexAiNotebook

## Overview

This Pulumi module provisions a Vertex AI Workbench instance on Google Cloud Platform. It creates a managed JupyterLab notebook environment backed by a Compute Engine VM.

## Building

```bash
make build
```

## Running

```bash
# Preview changes
make preview

# Deploy
make up

# Destroy
make destroy
```

## Debug

To debug the Pulumi program locally:

```bash
export PULUMI_CONFIG_PASSPHRASE=""
go build -o pulumi . && pulumi preview --stack dev --non-interactive
```

## Module Files

| File | Purpose |
|------|---------|
| `module/main.go` | Entry point — creates GCP provider and calls resource functions |
| `module/locals.go` | Locals struct, instance name resolution, GCP label computation |
| `module/workbench_instance.go` | Creates `workbench.NewInstance` with all configurations |
| `module/outputs.go` | Output constant names |

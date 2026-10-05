# Kubernetes Storage Class - Pulumi Module

## Overview

This Pulumi module creates and manages a Kubernetes `storage.k8s.io/v1` StorageClass. It supports the complete StorageClass surface: provisioner, provisioner-specific parameters, reclaim policy, volume binding mode, volume expansion, mount options, allowed topologies, and the default-class marker.

## Architecture

```
iac/pulumi/
├── main.go              # Entrypoint: loads IaC input, calls module
├── Pulumi.yaml          # Pulumi project configuration
├── Makefile             # Make targets for preview/up/down/refresh
└── module/
    ├── main.go          # Orchestrator: provider init, resource creation, output export
    ├── locals.go        # Derived values: labels, annotations (incl. default-class marker), resolved policies
    ├── storageclass.go  # Creates kubernetes.storage.v1.StorageClass resource
    └── outputs.go       # Exports storage_class_name, provisioner, is_default_class
```

## How It Works

1. **IaC Input Loading**: The entrypoint loads `KubernetesStorageClassIacInput` from Pulumi config
2. **Locals Initialization**: `locals.go` computes:
   - Standard Planton labels merged with user labels (identity keys cannot be overridden)
   - User annotations, plus the `storageclass.kubernetes.io/is-default-class: "true"` annotation when `is_default_class` is set
   - The resolved reclaim policy and volume binding mode: the Kubernetes API strings, with the API server's own defaults (`Delete`, `Immediate`) applied when the spec omits the optional fields
3. **Provider Creation**: Kubernetes provider is initialized from `provider_config`
4. **StorageClass Creation**: A single `storage.k8s.io/v1` StorageClass is created with the provisioner, parameters, policies, expansion flag, mount options, and topology terms
5. **Output Export**: Class name, provisioner, and the default-class flag are exported as outputs

## Semantics Preserved by the Module

- **Policies are always sent explicitly** — `reclaimPolicy`, `volumeBindingMode`, and `allowVolumeExpansion` use the resolved values from locals rather than API-server defaulting, so the Pulumi and Terraform modules submit byte-identical objects for the same manifest
- **Immutable fields force replacement, delete-before-create** — `provisioner` and `parameters` are immutable upstream; the resource is created with `DeleteBeforeReplace` so a forced replacement never collides with the cluster-unique class name
- **The default-class marker is the annotation wire form** — `is_default_class: true` renders `storageclass.kubernetes.io/is-default-class: "true"`; the spec field is the source of truth, never a hand-written annotation
- **All topology terms are sent** — the API ORs multiple `allowed_topologies` terms, and this module passes every term through (see parity note below)

## Field Mapping

| Spec Field | StorageClass Field | Notes |
|------------|-------------------|-------|
| `provisioner` | `provisioner` | Immutable; change forces replacement |
| `parameters` | `parameters` | Immutable; omitted when empty |
| `reclaim_policy` | `reclaimPolicy` | `delete`/`retain` → `"Delete"`/`"Retain"`; default `Delete` |
| `volume_binding_mode` | `volumeBindingMode` | `immediate`/`wait_for_first_consumer` → `"Immediate"`/`"WaitForFirstConsumer"`; default `Immediate` |
| `allow_volume_expansion` | `allowVolumeExpansion` | Always sent explicitly (Kubernetes default is false) |
| `mount_options` | `mountOptions` | Omitted when empty |
| `allowed_topologies` | `allowedTopologies` | Every term sent; terms OR, requirements within a term AND |
| `is_default_class` | `metadata.annotations` | Rendered as the upstream default-class annotation |

> **PARITY-EXCEPTION**: the Terraform Kubernetes provider models `allowed_topologies` as a SINGLE selector term (`max_items = 1`); its module fails the plan with a precondition when the spec lists several. This module sends every term. A spec with multiple topology terms deploys via the Pulumi provisioner only.

## Usage

```bash
# Preview changes
make preview manifest=../../e2e/manifest.yaml

# Deploy
make up manifest=../../e2e/manifest.yaml

# Destroy
make down manifest=../../e2e/manifest.yaml
```

## Debug

```bash
# Build the module
go build ./module/...

# Build the entrypoint
go build .
```

> **Note**: The class object is creatable on any cluster — parameters are consumed by the CSI driver at provision time, not validated at class creation. Claims of the class only provision when the named driver is installed. A claim under a `wait_for_first_consumer` class stays Pending until a pod uses it — correct behavior, not an error.

# Kubernetes Persistent Volume Claim - Pulumi Module

## Overview

This Pulumi module creates and manages a Kubernetes `core/v1` PersistentVolumeClaim. It supports the complete PersistentVolumeClaim surface: access modes, storage requests and limits, StorageClass selection (including the empty-vs-absent distinction), volume mode, static binding to a named volume, volume selectors, and data sources (clone a PVC or restore a VolumeSnapshot).

## Architecture

```
iac/pulumi/
├── main.go                    # Entrypoint: loads IaC input, calls module
├── Pulumi.yaml                # Pulumi project configuration
├── Makefile                   # Make targets for preview/up/down/refresh
└── module/
    ├── main.go                # Orchestrator: provider init, resource creation, output export
    ├── locals.go              # Derived values: labels, annotations (incl. skipAwait), namespace, resolved class name
    ├── persistentvolumeclaim.go # Creates kubernetes.core.v1.PersistentVolumeClaim resource
    └── outputs.go             # Exports pvc_name, namespace, storage_request
```

## How It Works

1. **IaC Input Loading**: The entrypoint loads `KubernetesPersistentVolumeClaimIacInput` from Pulumi config
2. **Locals Initialization**: `locals.go` computes:
   - Standard Planton labels merged with user labels (identity keys cannot be overridden)
   - Annotations, always including `pulumi.com/skipAwait: "true"` (see binding note below), merged with user annotations
   - The target namespace (foreign-key references are pre-resolved; falls back to `default` when omitted)
   - Access modes with the Kubernetes-default fallback (`ReadWriteOnce`) applied module-side — the API itself REQUIRES `accessModes`, so there is no server default to defer to
   - The volume mode API string (`Filesystem` default)
   - The three-valued `storageClassName`: the resolved class name, `""` when `disable_dynamic_provisioning` is set, or nil (unset — cluster default applies)
3. **Provider Creation**: Kubernetes provider is initialized from `provider_config`
4. **PersistentVolumeClaim Creation**: A single `core/v1` PersistentVolumeClaim is created with resources, static-binding fields, selector, and data source
5. **Output Export**: Claim name, namespace, and storage request are exported as outputs

## Semantics Preserved by the Module

- **The deploy NEVER waits for the claim to bind** — the `pulumi.com/skipAwait` annotation opts out of Pulumi's PVC readiness await. Under a `WaitForFirstConsumer` StorageClass a claim is correctly Pending until a pod consumes it, and awaiting would hang every such deploy. The Terraform module's `wait_until_bound = false` is the same decision on the other engine
- **Empty-vs-absent class name is preserved** — nil means absent (cluster default class); `""` means explicitly no dynamic provisioning (bind only pre-provisioned volumes). The distinction is load-bearing upstream and is why the wire value is a pointer
- **Defaults are sent explicitly** — access modes and volume mode use the resolved values from locals, so the Pulumi and Terraform modules submit identical claims for the same manifest
- **Data sources map onto TypedLocalObjectReference** — PVC clones live in the core group (apiGroup omitted); VolumeSnapshot restores name the `snapshot.storage.k8s.io` group

## Field Mapping

| Spec Field | PersistentVolumeClaim Field | Notes |
|------------|----------------------------|-------|
| `namespace` | `metadata.namespace` | Resolved reference or literal; `default` when omitted |
| `access_modes` | `spec.accessModes` | Empty → `["ReadWriteOnce"]` |
| `storage_request` / `storage_limit` | `spec.resources.requests/limits.storage` | Limit omitted when empty |
| `storage_class_name` / `disable_dynamic_provisioning` | `spec.storageClassName` | Named / `""` / absent — three distinct wire values |
| `volume_mode` | `spec.volumeMode` | `filesystem`/`block` → `"Filesystem"`/`"Block"`; always sent |
| `volume_name` | `spec.volumeName` | Omitted when empty |
| `selector` | `spec.selector` | Match labels and expressions, one-to-one |
| `data_source` | `spec.dataSource` | See parity note below |

> **PARITY-EXCEPTION**: the Terraform Kubernetes provider's PVC resource cannot express `spec.dataSource`/`dataSourceRef` (clone a PVC / restore a VolumeSnapshot); its module fails the plan with a precondition when the field is set. This module sends the data source natively — manifests with `data_source` deploy via the Pulumi provisioner only.

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

> **Note**: A claim under a `wait_for_first_consumer` StorageClass stays Pending until a pod uses it — correct behavior, not an error. The outputs deliberately avoid bind-time status (bound volume name, phase) for the same reason.

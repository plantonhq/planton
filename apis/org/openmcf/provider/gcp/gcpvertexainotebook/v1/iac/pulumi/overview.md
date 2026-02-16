# Pulumi Module: GcpVertexAiNotebook

## Architecture

This module provisions a Vertex AI Workbench instance (`workbench.NewInstance`) using the Pulumi GCP provider. The instance is a Compute Engine VM pre-configured with JupyterLab and optional GPU accelerators.

## Module Structure

```
module/
  main.go                  # Entry point: Resources()
  locals.go                # Locals struct, label computation
  workbench_instance.go    # workbench.NewInstance creation
  outputs.go               # Output constant names
```

## Resource Flow

1. `Resources()` initializes locals and sets up the GCP provider
2. `workbenchInstance()` creates the `workbench.Instance` resource with:
   - GCE setup block reconstructed from flattened spec fields
   - Boot disk and data disk with optional CMEK encryption
   - Optional GPU accelerator, network interface, service account
   - Mutually exclusive VM image or container image
   - Shielded VM configuration
   - Framework GCP labels
3. Stack outputs are exported: instance_id, instance_name, proxy_uri, state, creator, create_time

## Key Implementation Details

### Flattened Spec to Nested GCE Setup

The spec.proto flattens `gce_setup` fields to the top level for user convenience. The Pulumi module reconstructs the nested `InstanceGceSetupArgs` structure:

- `machine_type` → `gceSetup.MachineType`
- `boot_disk` → `gceSetup.BootDisk`
- `accelerator_config` → `gceSetup.AcceleratorConfigs` (single-element array)
- `network_interface` → `gceSetup.NetworkInterfaces` (single-element array)
- `service_account` → `gceSetup.ServiceAccounts` (single-element array)

### Int32-to-String Conversions

The Pulumi SDK uses `string` for `disk_size_gb` and `core_count` fields. The module converts using `fmt.Sprintf("%d", value)`.

### CMEK Encryption

If `kms_key` is set on a disk, the module automatically sets `disk_encryption = "CMEK"`. Users don't need to set the encryption mode explicitly.

### Labels

Framework GCP labels are computed in `locals.go` and applied to the instance:
- `openmcf-resource: true`
- `openmcf-resource-name: {instance_name}`
- `openmcf-resource-kind: gcpvertexainotebook`
- `openmcf-organization: {org}` (if set)
- `openmcf-environment: {env}` (if set)
- `openmcf-resource-id: {id}` (if set)

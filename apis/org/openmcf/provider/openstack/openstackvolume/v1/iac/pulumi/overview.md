# OpenStackVolume Pulumi Module Architecture

## Overview

Single-resource module that provisions an OpenStack Cinder block storage volume using `blockstorage.NewVolumeV3()`.

## Data Flow

1. **StackInput** is loaded from base64-encoded YAML (set by the Planton CLI)
2. **Locals** extracts the target resource and resolves the optional `image_id` FK from `StringValueOrRef`
3. **OpenStack provider** is constructed from the `ProviderConfig` credential
4. **Volume** is created with all spec fields mapped to `VolumeV3Args`
5. **Outputs** are exported matching `stack_outputs.proto` fields

## FK Resolution

- `image_id` is an optional `StringValueOrRef` -- resolved at runtime by the FK resolver middleware
- If `image_id` is nil (not provided), the `ImageId` local is empty string and the field is not set on the volume

## Resource Mapping

| Spec Field | Pulumi Arg | Notes |
|-----------|------------|-------|
| size | Size | `pulumi.Int()` |
| description | Description | `pulumi.StringPtr()` |
| volume_type | VolumeType | `pulumi.StringPtr()` |
| availability_zone | AvailabilityZone | ForceNew |
| snapshot_id | SnapshotId | Mutually exclusive source |
| source_vol_id | SourceVolId | Mutually exclusive source |
| image_id | ImageId | Resolved from StringValueOrRef |
| metadata | Metadata | `pulumi.StringMap{}` |
| region | Region | ForceNew |

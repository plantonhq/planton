# OpenStackVolumeAttach Pulumi Module Architecture

## Overview

Single-resource module that attaches an OpenStack Cinder volume to a compute instance using `compute.NewVolumeAttach()`.

## Data Flow

1. **StackInput** is loaded from base64-encoded YAML
2. **Locals** extracts the target and resolves both required FKs (`instance_id`, `volume_id`)
3. **OpenStack provider** is constructed from `ProviderConfig`
4. **VolumeAttach** is created with resolved FK values and optional device/region
5. **Outputs** are exported matching `stack_outputs.proto` fields

## FK Resolution

Both `instance_id` and `volume_id` are required `StringValueOrRef` fields, resolved at runtime by the FK resolver middleware.

## Resource Mapping

| Spec Field | Pulumi Arg | Notes |
|-----------|------------|-------|
| instance_id | InstanceId | Resolved from StringValueOrRef |
| volume_id | VolumeId | Resolved from StringValueOrRef |
| device | Device | Optional, e.g. "/dev/vdb" |
| region | Region | ForceNew |

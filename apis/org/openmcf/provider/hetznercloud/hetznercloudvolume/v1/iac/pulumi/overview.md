# HetznerCloudVolume Pulumi Module — Architecture Overview

## Data Flow

```
manifest.yaml
  └─> HetznerCloudVolumeStackInput (proto)
        ├── target: HetznerCloudVolume
        │     ├── metadata.name → volume name in Hetzner Cloud
        │     ├── metadata.org, env, id, labels → label computation
        │     └── spec
        │           ├── size (int32, required) → volume size in GB
        │           ├── location (string, required) → datacenter
        │           ├── format (enum, optional) → filesystem format
        │           ├── server_id (StringValueOrRef, optional) → attachment target
        │           ├── automount (bool) → auto-mount on attach
        │           └── delete_protection (bool) → prevent deletion
        └── provider_config: HetznerCloudProviderConfig
              └── hcloud_token (or HCLOUD_TOKEN env var)
```

## Module Structure

1. **main.go (entrypoint)**: Loads `HetznerCloudVolumeStackInput` from the `STACK_INPUT` environment variable (base64-encoded YAML) via `stackinput.LoadStackInput`, then calls `module.Resources`.

2. **module/main.go**: Orchestrates resource creation:
   - Initializes locals from stack input
   - Creates a Hetzner Cloud Pulumi provider via `pulumihcloudprovider.Get`
   - Calls `volume()` to create the volume, handle optional attachment, and export outputs

3. **module/locals.go**: Extracts provider config and target resource, then builds the label map:
   - Standard labels are set from metadata (`resource`, `name`, `kind`, `org`, `env`, `id`)
   - User-specified `metadata.labels` are merged in; standard labels take precedence on key conflicts

4. **module/volume.go**: The core resource file. Creates one or two resources:

   **Volume creation:** Creates `hcloud.NewVolume` with:
   - Name from `metadata.name`
   - Size, Location from spec (required fields)
   - Labels from locals (merged standard + user labels)
   - DeleteProtection from spec
   - Format (guarded): only set when `spec.Format` is not `format_unspecified`; the enum's `.String()` method provides the value directly

   **Conditional attachment:** Guarded by `if spec.ServerId != nil && spec.ServerId.GetValue() != ""`:
   - Volume ID: converted from `IDOutput` (string) to `IntOutput` via `ApplyT(strconv.Atoi)` — the volume's actual ID is only known after creation
   - Server ID: converted from string to int via `strconv.Atoi` at creation time — the value is resolved from `StringValueOrRef` during stack input loading
   - Automount: set only when `spec.Automount` is true

   **Output export:** Exports two values:
   - `volume_id` from the volume's `.ID()`
   - `linux_device` from the volume's `.LinuxDevice`

5. **module/outputs.go**: Constants for output names (`volume_id`, `linux_device`), matching the `stack_outputs.proto` field names.

## Resource Graph

```
hcloud.Volume ("volume")
  │
  ├── Name             ← metadata.name
  ├── Size             ← spec.Size (int32)
  ├── Location         ← spec.Location (string)
  ├── Labels           ← locals.Labels (merged standard + user)
  ├── DeleteProtection ← spec.DeleteProtection (bool)
  │
  ├── [if format != unspecified] Format ← spec.Format.String()
  │
  ├── [if serverId set] hcloud.VolumeAttachment ("volume-attachment")
  │     ├── VolumeId  ← volume.ID() (int-converted via ApplyT)
  │     ├── ServerId  ← spec.ServerId.GetValue() (int-converted via strconv.Atoi)
  │     └── [if automount] Automount ← true
  │
  ├── Export: "volume_id"    ← volume.ID()
  └── Export: "linux_device" ← volume.LinuxDevice
```

## Key Design Points

- **Two categories of ID type conversion**: The volume module performs two string-to-integer conversions, each using a different mechanism:
  1. `VolumeId` for the attachment — `ApplyT(strconv.Atoi)` because the volume's actual ID is only available after creation (it is a Pulumi output)
  2. `ServerId` for the attachment — plain `strconv.Atoi` because the value is known before resource creation (resolved from `StringValueOrRef`)

  This is the same pattern used in the HetznerCloudServer module for rDNS (deployment-time conversion for self-referencing ID) vs. foreign key fields (creation-time conversion for resolved values).

- **Format enum mapping**: The proto enum's `.String()` method returns `"ext4"` or `"xfs"`, which is exactly what the Hetzner Cloud provider expects. The `format_unspecified` zero value is handled by a guard check — when the format is unspecified, the `Format` field is not set on `VolumeArgs`, resulting in a raw (unformatted) volume. This is more explicit than passing an empty string, which would cause a provider error.

- **Conditional attachment, not conditional volume**: The volume is always created. The attachment is the conditional resource. This means removing `serverId` from the spec detaches the volume (destroys the attachment) without destroying the volume itself. Adding `serverId` to an unattached volume creates a new attachment. This two-resource pattern provides clean lifecycle separation.

- **Automount guard**: The `Automount` field is only set on `VolumeAttachmentArgs` when `spec.Automount` is `true`. When `false` (the default), the field is omitted entirely, letting the provider use its default. This avoids passing `Automount: false` explicitly, which would be semantically correct but adds noise to the Pulumi plan output.

- **Label merge strategy**: Same CG01 pattern as all other components. Standard labels always win over user labels. Labels are applied only to the volume resource — the attachment resource does not support labels in the Hetzner Cloud API.

- **Single resource file**: Both the volume and the conditional attachment live in `volume.go`. This is appropriate because there is only one primary resource (the volume) with one conditional dependent (the attachment). The helper pattern used in the HetznerCloudServer module (e.g., `buildPublicNet()`) is unnecessary here — the attachment logic is straightforward enough to inline.

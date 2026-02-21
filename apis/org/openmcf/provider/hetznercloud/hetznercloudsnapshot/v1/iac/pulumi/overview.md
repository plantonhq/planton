# HetznerCloudSnapshot Pulumi Module — Architecture Overview

## Data Flow

```
manifest.yaml
  └─> HetznerCloudSnapshotStackInput (proto)
        ├── target: HetznerCloudSnapshot
        │     ├── metadata.name → snapshot name (label only; the snapshot itself is unnamed in the API)
        │     ├── metadata.org, env, id, labels → label computation
        │     └── spec
        │           ├── server_id (StringValueOrRef, required) → source server
        │           └── description (string, optional) → snapshot description
        └── provider_config: HetznerCloudProviderConfig
              └── hcloud_token (or HCLOUD_TOKEN env var)
```

## Module Structure

1. **main.go (entrypoint)**: Loads `HetznerCloudSnapshotStackInput` from the `STACK_INPUT` environment variable (base64-encoded YAML) via `stackinput.LoadStackInput`, then calls `module.Resources`.

2. **module/main.go**: Orchestrates resource creation:
   - Initializes locals from stack input
   - Creates a Hetzner Cloud Pulumi provider via `pulumihcloudprovider.Get`
   - Calls `snapshot()` to create the snapshot and export outputs

3. **module/locals.go**: Extracts provider config and target resource, then builds the label map:
   - Standard labels are set from metadata (`resource`, `name`, `kind`, `org`, `env`, `id`)
   - User-specified `metadata.labels` are merged in; standard labels take precedence on key conflicts

4. **module/snapshot.go**: The sole resource file. Creates one resource:

   **Server ID conversion:** Converts `spec.ServerId.GetValue()` from string to int via `strconv.Atoi`. This is a creation-time conversion — the value is resolved from `StringValueOrRef` during stack input loading, so it is known before resource creation begins. Fails fast with a descriptive error if the value is not a valid integer.

   **Snapshot creation:** Creates `hcloud.NewSnapshot` with:
   - `ServerId` from the converted integer
   - `Labels` from locals (merged standard + user labels)
   - `Description` (guarded): only set when `spec.Description` is non-empty

   **Output export:** Exports `snapshot_id` from `createdSnapshot.ID()`.

5. **module/outputs.go**: Single constant `OpSnapshotId = "snapshot_id"`, matching the `stack_outputs.proto` field name.

## Resource Graph

```
hcloud.Snapshot ("snapshot")
  │
  ├── ServerId    ← spec.ServerId.GetValue() (int-converted via strconv.Atoi)
  ├── Labels      ← locals.Labels (merged standard + user)
  │
  ├── [if description != ""] Description ← spec.Description
  │
  └── Export: "snapshot_id" ← createdSnapshot.ID()
```

## Key Design Points

- **Creation-time ID conversion only**: Unlike components with self-referencing IDs (e.g., HetznerCloudVolume's `VolumeId` for the attachment), this module performs only a creation-time `strconv.Atoi` conversion. There is no `ApplyT` conversion because the server ID is not derived from a resource output created within this module — it comes from the resolved `StringValueOrRef` input.

- **No conditional resources**: Every apply creates exactly one `hcloud.Snapshot`. There is no conditional logic (no "if serverId is set" guard). This is the simplest module structure in the Hetzner Cloud component catalog.

- **Description guard**: The `Description` field is only set on `SnapshotArgs` when `spec.Description` is non-empty. When the field is omitted, the provider passes `nil` to the API, and the snapshot is created without a description. This avoids passing an empty string, which would set the description to `""` in the API (subtly different from "no description").

- **Snapshot ID as image ID**: The exported `snapshot_id` is the Hetzner Cloud image ID. Snapshots are stored as Images in the API, so this ID can be used anywhere an image ID is accepted — most importantly, as the `image` parameter when creating a new `HetznerCloudServer`.

- **Label merge strategy**: Same CG01 pattern as all other components. Standard labels always win over user labels. The snapshot resource supports labels in the Hetzner Cloud API, so all labels are applied directly to the snapshot.

- **Single resource file**: The entire module logic fits in `snapshot.go` because there is only one resource with no conditional dependents. No helper functions, no builders, no resource graph complexity.

# HetznerCloudPlacementGroup Pulumi Module — Architecture Overview

## Data Flow

```
manifest.yaml
  └─> HetznerCloudPlacementGroupStackInput (proto)
        ├── target: HetznerCloudPlacementGroup
        │     ├── metadata.name → placement group name in Hetzner Cloud
        │     ├── metadata.org, env, id, labels → label computation
        │     └── spec.type → placement group strategy (default: spread)
        └── provider_config: HetznerCloudProviderConfig
              └── hcloud_token (or HCLOUD_TOKEN env var)
```

## Module Structure

1. **main.go (entrypoint)**: Loads `HetznerCloudPlacementGroupStackInput` from the `STACK_INPUT` environment variable (base64-encoded YAML) via `stackinput.LoadStackInput`, then calls `module.Resources`.

2. **module/main.go**: Orchestrates resource creation:
   - Initializes locals from stack input
   - Creates a Hetzner Cloud Pulumi provider via `pulumihcloudprovider.Get`
   - Calls `placementGroup()` to create the placement group resource

3. **module/locals.go**: Extracts provider config and target resource, then builds the label map:
   - Standard labels are set from metadata (`resource`, `name`, `kind`, `organization`, `environment`, `id`)
   - User-specified `metadata.labels` are merged in; standard labels take precedence on key conflicts

4. **module/placement_group.go**: Creates the `hcloud.PlacementGroup` resource:
   - `Name` from `metadata.name`
   - `Type` from `spec.GetType().String()` — uses the proto enum's string representation, with the proto default mechanism providing `"spread"` when unset
   - `Labels` from the computed label map
   - Exports `placement_group_id` from the resource's `ID()` output

5. **module/outputs.go**: Constant for the output name (`placement_group_id`), matching the `stack_outputs.proto` field name.

## Provider Configuration

The `pulumihcloudprovider.Get()` helper maps `HetznerCloudProviderConfig` to `hcloud.ProviderArgs`. It supports:
- Explicit API token from provider config
- Fallback to `HCLOUD_TOKEN` environment variable
- Optional endpoint, poll interval, and poll function overrides

## Key Design Points

- **Single resource**: This module creates exactly one `hcloud.PlacementGroup` — no conditional or optional sub-resources.
- **Proto default for type**: The spec uses `optional Type type = 1` with a proto-level default annotation. `spec.GetType()` returns the default enum value (`spread`) when the field is unset, so the Pulumi module does not need its own defaulting logic.
- **Foundation of HA DAG**: The `placement_group_id` output is referenced by `HetznerCloudServer` via `StringValueOrRef` for anti-affinity assignment at server creation.
- **Label merge strategy**: Standard labels always win over user labels. This prevents users from accidentally overriding management metadata while still allowing custom labels for organizational needs.
- **Type change triggers replacement**: Changing the placement group type in Hetzner Cloud requires resource replacement (destroy + create). The Pulumi provider handles this transparently, but all member servers must be removed first.

# HetznerCloudSshKey Pulumi Module — Architecture Overview

## Data Flow

```
manifest.yaml
  └─> HetznerCloudSshKeyStackInput (proto)
        ├── target: HetznerCloudSshKey
        │     ├── metadata.name → SSH key name in Hetzner Cloud
        │     ├── metadata.org, env, id, labels → label computation
        │     └── spec.public_key → SSH public key content
        └── provider_config: HetznerCloudProviderConfig
              └── hcloud_token (or HCLOUD_TOKEN env var)
```

## Module Structure

1. **main.go (entrypoint)**: Loads `HetznerCloudSshKeyStackInput` from the `STACK_INPUT` environment variable (base64-encoded YAML) via `stackinput.LoadStackInput`, then calls `module.Resources`.

2. **module/main.go**: Orchestrates resource creation:
   - Initializes locals from stack input
   - Creates a Hetzner Cloud Pulumi provider via `pulumihcloudprovider.Get`
   - Calls `sshKey()` to create the SSH key resource

3. **module/locals.go**: Extracts provider config and target resource, then builds the label map:
   - Standard labels are set from metadata (`resource`, `resource_name`, `resource_kind`, `org`, `env`, `resource_id`)
   - User-specified `metadata.labels` are merged in; standard labels take precedence on key conflicts

4. **module/ssh_key.go**: Creates the `hcloud.SshKey` resource:
   - `Name` from `metadata.name`
   - `PublicKey` from `spec.public_key`
   - `Labels` from the computed label map
   - Exports `ssh_key_id` (from resource ID) and `fingerprint`

5. **module/outputs.go**: Constants for output names (`ssh_key_id`, `fingerprint`), matching `stack_outputs.proto` field names.

## Provider Configuration

The `pulumihcloudprovider.Get()` helper maps `HetznerCloudProviderConfig` to `hcloud.ProviderArgs`. It supports:
- Explicit API token from provider config
- Fallback to `HCLOUD_TOKEN` environment variable
- Optional endpoint, poll interval, and poll function overrides

## Key Design Points

- **Single resource**: This module creates exactly one `hcloud.SshKey` — no conditional or optional sub-resources.
- **Foundation of compute DAG**: The `ssh_key_id` output is referenced by `HetznerCloudServer` via `StringValueOrRef` for key injection at server creation.
- **Label merge strategy**: Standard labels always win over user labels. This prevents users from accidentally overriding management metadata while still allowing custom labels for their own organizational needs.
- **Force replacement**: Changing `public_key` triggers resource replacement because the Hetzner Cloud API does not support in-place key material updates. The Pulumi provider handles this transparently.

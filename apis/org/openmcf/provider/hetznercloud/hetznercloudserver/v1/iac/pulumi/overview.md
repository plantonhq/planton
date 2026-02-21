# HetznerCloudServer Pulumi Module — Architecture Overview

## Data Flow

```
manifest.yaml
  └─> HetznerCloudServerStackInput (proto)
        ├── target: HetznerCloudServer
        │     ├── metadata.name → server name in Hetzner Cloud
        │     ├── metadata.org, env, id, labels → label computation
        │     └── spec
        │           ├── server_type (string, required) → hardware profile
        │           ├── image (string, required) → OS image
        │           ├── location (string, required) → datacenter
        │           ├── ssh_keys (StringValueOrRef[], optional) → key injection
        │           ├── user_data (string, optional) → cloud-init
        │           ├── placement_group_id (StringValueOrRef, optional) → anti-affinity
        │           ├── firewall_ids (StringValueOrRef[], optional) → security rules
        │           ├── public_net (PublicNet, optional) → public networking config
        │           │     ├── ipv4_enabled (optional bool, default true)
        │           │     ├── ipv6_enabled (optional bool, default true)
        │           │     ├── ipv4 (StringValueOrRef) → Primary IP attachment
        │           │     └── ipv6 (StringValueOrRef) → Primary IP attachment
        │           ├── networks (NetworkAttachment[], optional) → private networking
        │           │     ├── network_id (StringValueOrRef, required)
        │           │     ├── ip (string, optional)
        │           │     └── alias_ips (string[], optional)
        │           ├── backups (bool) → daily backup toggle
        │           ├── keep_disk (bool) → prevent disk upgrade on resize
        │           ├── delete_protection (bool)
        │           ├── rebuild_protection (bool)
        │           ├── shutdown_before_deletion (bool)
        │           └── dns_ptr (string, optional) → conditional rDNS creation
        └── provider_config: HetznerCloudProviderConfig
              └── hcloud_token (or HCLOUD_TOKEN env var)
```

## Module Structure

1. **main.go (entrypoint)**: Loads `HetznerCloudServerStackInput` from the `STACK_INPUT` environment variable (base64-encoded YAML) via `stackinput.LoadStackInput`, then calls `module.Resources`.

2. **module/main.go**: Orchestrates resource creation:
   - Initializes locals from stack input
   - Creates a Hetzner Cloud Pulumi provider via `pulumihcloudprovider.Get`
   - Calls `server()` to create the server, handle optional rDNS, and export outputs

3. **module/locals.go**: Extracts provider config and target resource, then builds the label map:
   - Standard labels are set from metadata (`resource`, `name`, `kind`, `org`, `env`, `id`)
   - User-specified `metadata.labels` are merged in; standard labels take precedence on key conflicts

4. **module/server.go**: The core resource file. Creates one or two resources with extensive option wiring:

   **Server creation:** Creates `hcloud.NewServer` with:
   - Name from `metadata.name`
   - ServerType, Image, Location from spec (required fields)
   - Labels from locals (merged standard + user labels)
   - All boolean flags from spec (Backups, KeepDisk, DeleteProtection, RebuildProtection, ShutdownBeforeDeletion)

   **Optional fields (guarded):**
   - `UserData` — set only when `spec.UserData != ""`
   - `SshKeys` — resolved from `StringValueOrRef[]`, passed as string array (provider accepts names or IDs)
   - `PlacementGroupId` — converted from string to `IntPtr` via `strconv.Atoi`
   - `FirewallIds` — each entry converted from string to int, collected into `IntArray` via `toIntInputArray()` helper
   - `PublicNets` — set only when `spec.PublicNet != nil`, built by `buildPublicNet()` helper
   - `Networks` — set only when `len(spec.Networks) > 0`, built by `buildNetworkAttachments()` helper

   **Conditional rDNS creation:** Guarded by `if spec.DnsPtr != ""`. Converts the server's `IDOutput` (string) to `IntOutput` (int) via `ApplyT(strconv.Atoi)`, then creates `hcloud.NewRdns` wired to the server's IPv4 address.

   **Output export:** Exports four values:
   - `server_id` from the server's `.ID()`
   - `ipv4_address` from the server's `.Ipv4Address`
   - `ipv6_address` from the server's `.Ipv6Address`
   - `status` from the server's `.Status`

5. **module/outputs.go**: Constants for output names (`server_id`, `ipv4_address`, `ipv6_address`, `status`), matching the `stack_outputs.proto` field names.

## Resource Graph

```
hcloud.Server ("server")
  │
  ├── SshKeys        ← spec.SshKeys[] (string array, names or IDs)
  ├── PlacementGroupId ← spec.PlacementGroupId (int-converted from string)
  ├── FirewallIds     ← spec.FirewallIds[] (int array, each converted from string)
  │
  ├── [if PublicNet != nil] PublicNets
  │     ├── Ipv4Enabled ← spec.PublicNet.Ipv4Enabled (optional bool, default true)
  │     ├── Ipv6Enabled ← spec.PublicNet.Ipv6Enabled (optional bool, default true)
  │     ├── [if ipv4 set] Ipv4 ← spec.PublicNet.Ipv4 (int-converted from string)
  │     └── [if ipv6 set] Ipv6 ← spec.PublicNet.Ipv6 (int-converted from string)
  │
  ├── [for each network] Networks[]
  │     ├── NetworkId ← net.NetworkId (int-converted from string)
  │     ├── [if ip set] Ip ← net.Ip
  │     └── AliasIps  ← net.AliasIps (always passed, even empty — bridge bug workaround)
  │
  ├── [if dnsPtr != ""] hcloud.Rdns ("rdns")
  │     ├── ServerId  ← server.ID() (int-converted via ApplyT)
  │     ├── IpAddress ← server.Ipv4Address
  │     └── DnsPtr    ← spec.DnsPtr
  │
  ├── Export: "server_id"    ← server.ID()
  ├── Export: "ipv4_address" ← server.Ipv4Address
  ├── Export: "ipv6_address" ← server.Ipv6Address
  └── Export: "status"       ← server.Status
```

## Key Design Points

- **Five categories of ID type conversion**: The Pulumi hcloud SDK requires integer inputs where the spec provides strings. The module performs conversions in five places:
  1. `PlacementGroupId` — `strconv.Atoi` at creation time
  2. `FirewallIds` — loop + `strconv.Atoi` at creation time, collected via `toIntInputArray()` helper
  3. `NetworkId` per attachment — `strconv.Atoi` at creation time inside `buildNetworkAttachments()`
  4. `PublicNet.Ipv4` and `PublicNet.Ipv6` — `strconv.Atoi` at creation time inside `buildPublicNet()`
  5. `ServerId` for rDNS — `ApplyT(strconv.Atoi)` at deployment time (depends on server output)

  The first four use plain Go conversions on resolved `StringValueOrRef` values. The fifth uses Pulumi's `ApplyT` because the server's actual ID is only available after creation.

- **PublicNet nil-vs-present semantics**: When `spec.PublicNet` is `nil`, the module does not set `ServerArgs.PublicNets` at all — the provider uses its default behavior (auto-assign IPv4 + IPv6). When `spec.PublicNet` is set, `buildPublicNet()` is called, which explicitly handles the `optional bool` fields: `nil` defaults to `true`, non-nil uses the explicit value. This prevents an empty `publicNet: {}` from accidentally disabling all public networking.

- **AliasIps bridge bug workaround (#650)**: In `buildNetworkAttachments()`, `AliasIps` is always set on `ServerNetworkTypeArgs` — even when the spec's `alias_ips` is nil or empty. Without this, the Terraform bridge detects phantom drift on every `pulumi up` and attempts a network detach/reattach cycle:
  ```go
  AliasIps: pulumi.ToStringArray(net.AliasIps), // always pass, even when empty
  ```

- **SSH keys as strings, not integers**: Unlike other foreign key fields, `SshKeys` is passed as a `StringArray` (not `IntArray`). The Hetzner Cloud provider accepts SSH key names or numeric IDs as strings. This avoids an unnecessary `strconv.Atoi` conversion and lets users reference keys by human-readable name.

- **Helper functions isolate complexity**: `buildPublicNet()` encapsulates the optional-bool-with-default-true logic and Primary IP conversion. `buildNetworkAttachments()` encapsulates the per-network ID conversion and AliasIps workaround. `toIntInputArray()` converts `[]int` to `[]pulumi.IntInput` for the FirewallIds field. This keeps `server()` focused on orchestration.

- **Label merge strategy**: Standard labels always win over user labels. This prevents users from accidentally overriding management metadata while still allowing custom labels. Labels are applied only to the server resource — the rDNS resource does not support labels in the Hetzner Cloud API.

- **Single resource file**: Despite the server's complexity (15 spec fields, 2 nested messages, conditional rDNS), all resource creation lives in one file (`server.go`). This is appropriate because there is only one primary resource (the server) with one conditional dependent (rDNS). The helper functions provide modularity without file fragmentation.

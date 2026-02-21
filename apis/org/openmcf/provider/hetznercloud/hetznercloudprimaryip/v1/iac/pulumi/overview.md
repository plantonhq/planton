# HetznerCloudPrimaryIp Pulumi Module — Architecture Overview

## Data Flow

```
manifest.yaml
  └─> HetznerCloudPrimaryIpStackInput (proto)
        ├── target: HetznerCloudPrimaryIp
        │     ├── metadata.name → Primary IP name in Hetzner Cloud
        │     ├── metadata.org, env, id, labels → label computation
        │     └── spec
        │           ├── type (enum: ipv4, ipv6) → address type
        │           ├── location (string) → allocation location
        │           ├── dns_ptr (string, optional) → conditional rDNS creation
        │           └── delete_protection (bool) → API delete guard
        └── provider_config: HetznerCloudProviderConfig
              └── hcloud_token (or HCLOUD_TOKEN env var)
```

## Module Structure

1. **main.go (entrypoint)**: Loads `HetznerCloudPrimaryIpStackInput` from the `STACK_INPUT` environment variable (base64-encoded YAML) via `stackinput.LoadStackInput`, then calls `module.Resources`.

2. **module/main.go**: Orchestrates resource creation:
   - Initializes locals from stack input
   - Creates a Hetzner Cloud Pulumi provider via `pulumihcloudprovider.Get`
   - Calls `primaryIp()` to create the IP and optional rDNS

3. **module/locals.go**: Extracts provider config and target resource, then builds the label map:
   - Standard labels are set from metadata (`resource`, `name`, `kind`, `org`, `env`, `id`)
   - User-specified `metadata.labels` are merged in; standard labels take precedence on key conflicts

4. **module/primary_ip.go**: The core resource file. Creates one or two resources:

   **Primary IP creation:** Creates `hcloud.NewPrimaryIp` with:
   - Name from `metadata.name`
   - Type from spec enum (converted to string via `.String()`)
   - Location from spec
   - `assignee_type` hardcoded to `"server"`
   - `auto_delete` hardcoded to `false`
   - Labels from locals
   - Delete protection from spec

   **Conditional rDNS creation:** Guarded by `if spec.DnsPtr != ""`. When the guard passes:

   The ID type conversion is performed — `PrimaryIp.ID()` returns `IDOutput` (string), but `RdnsArgs.PrimaryIpId` expects `IntInput`:
   ```go
   primaryIpIdInt := createdPrimaryIp.ID().ApplyT(func(id pulumi.ID) (int, error) {
       return strconv.Atoi(string(id))
   }).(pulumi.IntOutput)
   ```

   Then `hcloud.NewRdns` is created with:
   - `PrimaryIpId` from the converted integer output
   - `IpAddress` from the Primary IP's allocated address output
   - `DnsPtr` from the spec

   **Output export:** Exports three values:
   - `primary_ip_id` from the Primary IP's `.ID()`
   - `ip_address` from the Primary IP's `.IpAddress`
   - `ip_network` from the Primary IP's `.IpNetwork`

5. **module/outputs.go**: Constants for output names (`primary_ip_id`, `ip_address`, `ip_network`), matching the `stack_outputs.proto` field names.

## Resource Graph

```
hcloud.PrimaryIp ("primary-ip")
  │
  ├── [if dnsPtr != ""] hcloud.Rdns ("rdns")
  │     ├── PrimaryIpId ← primary-ip.ID() (int-converted)
  │     ├── IpAddress   ← primary-ip.IpAddress
  │     └── DnsPtr      ← spec.DnsPtr
  │
  ├── Export: "primary_ip_id" ← primary-ip.ID()
  ├── Export: "ip_address"    ← primary-ip.IpAddress
  └── Export: "ip_network"    ← primary-ip.IpNetwork
```

## Key Design Points

- **Single-resource simplicity with conditional sub-resource**: Unlike multi-resource components (Network with N subnets + M routes), this module creates at most two resources. The `if` guard for rDNS keeps the code path simple — no loops, no dynamic resource naming.

- **ID type conversion**: The `ApplyT` conversion from `IDOutput` (string) to `IntOutput` is a friction point in the Pulumi hcloud SDK that every user encounters when connecting Primary IPs to rDNS resources. The module handles it once, preventing this boilerplate from leaking into user code.

- **Hardcoded safe defaults**: `auto_delete = false` and `assignee_type = "server"` are set in the resource creation call, not derived from the spec. These are not configurable because there is only one correct value for each in OpenMCF's component model.

- **Spec enum to string conversion**: The `type` field in the proto is an enum (`ipv4`, `ipv6`). The Pulumi SDK expects a string (`"ipv4"`, `"ipv6"`). The module calls `spec.Type.String()` to bridge the two. This conversion is safe because proto validation ensures only `ipv4` or `ipv6` reach the module — `ip_type_unspecified` is rejected by `buf.validate`.

- **Three outputs, not one**: The module exports `primary_ip_id` (for server assignment), `ip_address` (for DNS record configuration), and `ip_network` (for IPv6 firewall rules). The server component only needs `primary_ip_id`, but users need the other two for out-of-band configuration.

- **Label merge strategy**: Standard labels always win over user labels. This prevents users from accidentally overriding management metadata while still allowing custom labels for their own organizational needs. Labels are applied only to the Primary IP resource — the rDNS resource does not support labels in the Hetzner Cloud API.

# HetznerCloudFloatingIp Pulumi Module — Architecture Overview

## Data Flow

```
manifest.yaml
  └─> HetznerCloudFloatingIpStackInput (proto)
        ├── target: HetznerCloudFloatingIp
        │     ├── metadata.name → Floating IP name in Hetzner Cloud
        │     ├── metadata.org, env, id, labels → label computation
        │     └── spec
        │           ├── type (enum: ipv4, ipv6) → address type
        │           ├── home_location (string) → allocation location
        │           ├── description (string, optional) → human-readable description
        │           ├── server_id (StringValueOrRef, optional) → server assignment
        │           ├── dns_ptr (string, optional) → conditional rDNS creation
        │           └── delete_protection (bool) → API delete guard
        └── provider_config: HetznerCloudProviderConfig
              └── hcloud_token (or HCLOUD_TOKEN env var)
```

## Module Structure

1. **main.go (entrypoint)**: Loads `HetznerCloudFloatingIpStackInput` from the `STACK_INPUT` environment variable (base64-encoded YAML) via `stackinput.LoadStackInput`, then calls `module.Resources`.

2. **module/main.go**: Orchestrates resource creation:
   - Initializes locals from stack input
   - Creates a Hetzner Cloud Pulumi provider via `pulumihcloudprovider.Get`
   - Calls `floatingIp()` to create the IP, handle optional assignment, and create optional rDNS

3. **module/locals.go**: Extracts provider config and target resource, then builds the label map:
   - Standard labels are set from metadata (`resource`, `name`, `kind`, `org`, `env`, `id`)
   - User-specified `metadata.labels` are merged in; standard labels take precedence on key conflicts

4. **module/floating_ip.go**: The core resource file. Creates one or two resources with optional server assignment:

   **Floating IP creation:** Creates `hcloud.NewFloatingIp` with:
   - Name from `metadata.name`
   - Type from spec enum (converted to string via `.String()`)
   - HomeLocation from spec
   - Labels from locals
   - Delete protection from spec

   **Optional description:** When `spec.Description` is non-empty, `Description` is set on the Floating IP args.

   **Optional server assignment:** Guarded by `if spec.ServerId != nil && spec.ServerId.GetValue() != ""`. When the guard passes, the string server ID is converted to an integer via `strconv.Atoi` and set on `FloatingIpArgs.ServerId`. This conversion happens at creation time using the resolved string value from `StringValueOrRef`. If the value is not a valid integer, the module fails with a descriptive error.

   **Conditional rDNS creation:** Guarded by `if spec.DnsPtr != ""`. When the guard passes:

   The ID type conversion is performed — `FloatingIp.ID()` returns `IDOutput` (string), but `RdnsArgs.FloatingIpId` expects `IntInput`:
   ```go
   floatingIpIdInt := createdFloatingIp.ID().ApplyT(func(id pulumi.ID) (int, error) {
       return strconv.Atoi(string(id))
   }).(pulumi.IntOutput)
   ```

   Then `hcloud.NewRdns` is created with:
   - `FloatingIpId` from the converted integer output
   - `IpAddress` from the Floating IP's allocated address output
   - `DnsPtr` from the spec

   **Output export:** Exports three values:
   - `floating_ip_id` from the Floating IP's `.ID()`
   - `ip_address` from the Floating IP's `.IpAddress`
   - `ip_network` from the Floating IP's `.IpNetwork`

5. **module/outputs.go**: Constants for output names (`floating_ip_id`, `ip_address`, `ip_network`), matching the `stack_outputs.proto` field names.

## Resource Graph

```
hcloud.FloatingIp ("floating-ip")
  │
  ├── [if serverId != ""] ServerId ← spec.ServerId (int-converted from string)
  │
  ├── [if dnsPtr != ""] hcloud.Rdns ("rdns")
  │     ├── FloatingIpId ← floating-ip.ID() (int-converted)
  │     ├── IpAddress    ← floating-ip.IpAddress
  │     └── DnsPtr       ← spec.DnsPtr
  │
  ├── Export: "floating_ip_id" ← floating-ip.ID()
  ├── Export: "ip_address"     ← floating-ip.IpAddress
  └── Export: "ip_network"     ← floating-ip.IpNetwork
```

## Key Design Points

- **Two conditional code paths, one resource file**: The module handles optional server assignment and optional rDNS in the same file (`floating_ip.go`). Unlike multi-resource components (Network with N subnets + M routes), this module creates at most two resources and has no loops — just two `if` guards.

- **Two ID type conversions**: The Pulumi hcloud SDK requires integer inputs where Hetzner Cloud returns string IDs. The module performs two different conversions:
  1. `spec.ServerId.GetValue()` → `strconv.Atoi` at creation time (before the resource exists)
  2. `createdFloatingIp.ID()` → `ApplyT(strconv.Atoi)` at deployment time (after the resource exists, via Pulumi's output system)

  The first is a simple Go conversion. The second is a Pulumi `ApplyT` callback that runs during the deployment phase.

- **Spec enum to string conversion**: The `type` field in the proto is an enum (`ipv4`, `ipv6`). The Pulumi SDK expects a string (`"ipv4"`, `"ipv6"`). The module calls `spec.Type.String()` to bridge the two. This conversion is safe because proto validation ensures only `ipv4` or `ipv6` reach the module — `ip_type_unspecified` is rejected by `buf.validate`.

- **Label merge strategy**: Standard labels always win over user labels. This prevents users from accidentally overriding management metadata while still allowing custom labels for their own organizational needs. Labels are applied only to the Floating IP resource — the rDNS resource does not support labels in the Hetzner Cloud API.

- **Server assignment is on the Floating IP**: Unlike Primary IPs (where the server references the IP), Floating IPs own their assignment via the `server_id` attribute. This matches the Hetzner Cloud API model and means changing `serverId` in the spec triggers a Floating IP update (reassignment), not a replacement.

- **Three outputs**: The module exports `floating_ip_id` (for automation), `ip_address` (for DNS record configuration), and `ip_network` (for IPv6 firewall rules). All three are exported unconditionally — `ip_network` is simply empty for IPv4 Floating IPs.

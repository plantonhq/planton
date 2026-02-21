# HetznerCloudNetwork Pulumi Module — Architecture Overview

## Data Flow

```
manifest.yaml
  └─> HetznerCloudNetworkStackInput (proto)
        ├── target: HetznerCloudNetwork
        │     ├── metadata.name → Network name in Hetzner Cloud
        │     ├── metadata.org, env, id, labels → label computation
        │     └── spec
        │           ├── ip_range → network CIDR block
        │           ├── subnets[] → N subnet resources
        │           │     ├── type (enum) → "cloud" / "server" / "vswitch"
        │           │     ├── network_zone → "eu-central", "us-east", etc.
        │           │     ├── ip_range → subnet CIDR within network range
        │           │     └── vswitch_id → Robot vSwitch ID (vswitch type only)
        │           ├── routes[] → M route resources
        │           │     ├── destination → target CIDR
        │           │     └── gateway → next-hop IP within a subnet
        │           ├── delete_protection → API delete guard
        │           └── expose_routes_to_vswitch → route visibility for vSwitch
        └── provider_config: HetznerCloudProviderConfig
              └── hcloud_token (or HCLOUD_TOKEN env var)
```

## Module Structure

1. **main.go (entrypoint)**: Loads `HetznerCloudNetworkStackInput` from the `STACK_INPUT` environment variable (base64-encoded YAML) via `stackinput.LoadStackInput`, then calls `module.Resources`.

2. **module/main.go**: Orchestrates resource creation:
   - Initializes locals from stack input
   - Creates a Hetzner Cloud Pulumi provider via `pulumihcloudprovider.Get`
   - Calls `network()` to create all network resources

3. **module/locals.go**: Extracts provider config and target resource, then builds the label map:
   - Standard labels are set from metadata (`resource`, `name`, `kind`, `org`, `env`, `id`)
   - User-specified `metadata.labels` are merged in; standard labels take precedence on key conflicts

4. **module/network.go**: The core resource file. Creates three types of resources in sequence:

   **Network creation:** Creates a single `hcloud.NewNetwork` with name, IP range, labels, and protection settings from the spec.

   **ID type conversion:** The Pulumi hcloud SDK returns `Network.ID()` as `IDOutput` (string), but `NetworkSubnetArgs.NetworkId` and `NetworkRouteArgs.NetworkId` expect `IntInput`. The module converts once via `ApplyT`:
   ```go
   networkIdInt := createdNetwork.ID().ApplyT(func(id pulumi.ID) (int, error) {
       return strconv.Atoi(string(id))
   }).(pulumi.IntOutput)
   ```
   This shared `networkIdInt` feeds all subnet and route resources, avoiding repeated conversion.

   **Subnet loop:** Iterates over `spec.Subnets`, creating one `hcloud.NewNetworkSubnet` per entry. Each resource is named `subnet-{sanitized_cidr}` (e.g., `subnet-10-0-1-0-24`). The `vswitchId` is only set when non-zero, avoiding sending a zero value that the API would reject for non-vswitch types.

   **Route loop:** Iterates over `spec.Routes`, creating one `hcloud.NewNetworkRoute` per entry. Each resource is named `route-{sanitized_cidr}` using the destination CIDR.

   **Output export:** Exports `network_id` from the network resource's `.ID()`.

5. **module/outputs.go**: Constants for output names (`network_id`), matching the `stack_outputs.proto` field name.

## Resource Graph

```
hcloud.Network ("network")
  │
  ├── hcloud.NetworkSubnet ("subnet-10-0-1-0-24")  ─┐
  ├── hcloud.NetworkSubnet ("subnet-10-0-2-0-24")  ─┤── all reference networkIdInt
  │   ...                                           │
  ├── hcloud.NetworkRoute ("route-172-16-0-0-12")  ─┤
  │   ...                                          ─┘
  │
  └── Export: "network_id" → stack outputs
```

## Key Design Points

- **Multi-resource component**: Unlike single-resource components (SshKey, PlacementGroup, Firewall), this module creates 1 + N + M resources. The network must be created first because subnets and routes reference its ID.

- **CIDR sanitization for resource names**: Pulumi resource names cannot contain dots or slashes. The `sanitizeCidr` helper replaces `.`, `/`, and `:` with hyphens (e.g., `10.0.1.0/24` becomes `10-0-1-0-24`). This produces deterministic, human-readable resource names.

- **Keying by CIDR**: Subnets are keyed by `ip_range` and routes by `destination`. Adding or removing a subnet/route only affects that specific resource — other subnets and routes remain untouched. This is the Pulumi equivalent of Terraform's `for_each` keying.

- **Conditional vSwitch ID**: The `vswitchId` field is only set when the proto value is non-zero (`subnet.VswitchId != 0`). This prevents sending a `vswitch_id: 0` attribute to the API, which would fail for cloud and server subnet types.

- **Foundation of networking DAG**: The `network_id` output is referenced by `HetznerCloudServer` and `HetznerCloudLoadBalancer` via `StringValueOrRef`. The network has no knowledge of which resources attach to it.

- **Label merge strategy**: Standard labels always win over user labels. This prevents users from accidentally overriding management metadata while still allowing custom labels for their own organizational needs. Labels are applied only to the network resource — subnets and routes do not support labels in the Hetzner Cloud API.

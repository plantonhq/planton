# HetznerCloudFirewall Pulumi Module — Architecture Overview

## Data Flow

```
manifest.yaml
  └─> HetznerCloudFirewallStackInput (proto)
        ├── target: HetznerCloudFirewall
        │     ├── metadata.name → Firewall name in Hetzner Cloud
        │     ├── metadata.org, env, id, labels → label computation
        │     └── spec.rules[] → inline firewall rules
        │           ├── direction (enum) → "in" / "out"
        │           ├── protocol (enum) → "icmp" / "tcp" / "udp" / "esp" / "gre"
        │           ├── port → "80", "80-443", "any" (TCP/UDP only)
        │           ├── source_ips[] → CIDRs (inbound only)
        │           ├── destination_ips[] → CIDRs (outbound only)
        │           └── description → optional label
        └── provider_config: HetznerCloudProviderConfig
              └── hcloud_token (or HCLOUD_TOKEN env var)
```

## Module Structure

1. **main.go (entrypoint)**: Loads `HetznerCloudFirewallStackInput` from the `STACK_INPUT` environment variable (base64-encoded YAML) via `stackinput.LoadStackInput`, then calls `module.Resources`.

2. **module/main.go**: Orchestrates resource creation:
   - Initializes locals from stack input
   - Creates a Hetzner Cloud Pulumi provider via `pulumihcloudprovider.Get`
   - Calls `firewall()` to create the firewall resource with inline rules

3. **module/locals.go**: Extracts provider config and target resource, then builds the label map:
   - Standard labels are set from metadata (`resource`, `name`, `kind`, `org`, `env`, `id`)
   - User-specified `metadata.labels` are merged in; standard labels take precedence on key conflicts

4. **module/firewall.go**: The core resource file. Performs two operations:
   - **Rule mapping loop**: Iterates over `spec.rules`, converting each proto `Rule` message into an `hcloud.FirewallRuleArgs`. Direction and protocol enums are converted to strings via `.String()`. Optional fields (`port`, `source_ips`, `destination_ips`, `description`) are only set when non-empty, avoiding nil-pointer issues in the Pulumi SDK.
   - **Resource creation**: Creates a single `hcloud.NewFirewall` with the mapped `Name`, `Labels`, and `Rules` array. Exports `firewall_id` from the resource's `.ID()`.

5. **module/outputs.go**: Constants for output names (`firewall_id`), matching the `stack_outputs.proto` field name.

## Provider Configuration

The `pulumihcloudprovider.Get()` helper maps `HetznerCloudProviderConfig` to `hcloud.ProviderArgs`. It supports:
- Explicit API token from provider config
- Fallback to `HCLOUD_TOKEN` environment variable
- Optional endpoint, poll interval, and poll function overrides

## Key Design Points

- **Single resource with inline rules**: The module creates exactly one `hcloud.Firewall`. Rules are passed as a `FirewallRuleArray`, not as separate resources. This means adding or removing a rule triggers an in-place update of the firewall, not creation/deletion of child resources.

- **Enum-to-string conversion**: Proto enum values (`Direction.in`, `Protocol.tcp`) are converted to their string representation via `.String()`. The enum names match the Hetzner Cloud API values exactly (`"in"`, `"out"`, `"tcp"`, `"udp"`, `"icmp"`, `"esp"`, `"gre"`), so no translation table is needed.

- **Conditional field population**: The rule mapping loop checks each optional field before setting it. This is necessary because the Pulumi SDK treats `pulumi.String("")` differently from a nil/unset field — an empty port string would cause the provider to send an empty `port` attribute to the API, which would fail for ICMP/ESP/GRE rules.

- **Foundation of security DAG**: The `firewall_id` output is referenced by `HetznerCloudServer` via `StringValueOrRef` for applying firewall rules at server creation. The firewall has no knowledge of which servers reference it.

- **Label merge strategy**: Standard labels always win over user labels. This prevents users from accidentally overriding management metadata while still allowing custom labels for their own organizational needs.

- **No apply_to**: The Hetzner Cloud API supports a firewall-side `apply_to` mechanism for binding to servers. This module does not use it. Server-to-firewall binding is managed by the server component's `firewallIds` field, keeping the dependency graph unidirectional.

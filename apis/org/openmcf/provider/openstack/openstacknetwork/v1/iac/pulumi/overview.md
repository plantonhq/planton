# OpenStackNetwork Pulumi Module — Architecture Overview

## Data Flow

```
manifest.yaml
  └─> OpenStackNetworkStackInput (proto)
        ├── target: OpenStackNetwork
        │     ├── metadata.name → network name
        │     └── spec.description, spec.admin_state_up, spec.shared,
        │         spec.external, spec.mtu, spec.dns_domain,
        │         spec.port_security_enabled, spec.tags, spec.region
        └── provider_config: OpenStackProviderConfig
              └── auth_url, credentials (oneof), region, ...
```

## Module Structure

1. **main.go (entrypoint)**: Loads `OpenStackNetworkStackInput` from the `STACK_INPUT` environment variable (base64-encoded YAML), then calls `module.Resources`.

2. **module/main.go**: Orchestrates resource creation:
   - Initializes locals from stack input
   - Creates an OpenStack Pulumi provider from `OpenStackProviderConfig`
   - Calls `network()` to create the Neutron network

3. **module/locals.go**: Extracts provider config and target resource from stack input into a `Locals` struct for convenient access.

4. **module/network.go**: Creates the `networking.Network` resource:
   - Sets `Name` from `metadata.name`
   - Sets `Description` if provided
   - Sets `AdminStateUp` (default true via middleware)
   - Sets `Shared` and `External` if true
   - Sets `Mtu` if non-zero
   - Sets `DnsDomain` if provided
   - Sets `PortSecurityEnabled` if explicitly set
   - Sets `Tags` if provided
   - Sets `Region` if overridden
   - Exports outputs: network_id (from resource ID), name, region

5. **module/outputs.go**: Constants for output names, matching `stack_outputs.proto` field names.

## Provider Configuration

The `pulumiopenstackprovider.Get()` helper maps `OpenStackProviderConfig` to Pulumi provider args:
- Supports all three auth methods (password, application credential, token)
- Falls back to `OS_*` environment variables when config is nil
- Handles TLS, endpoint type, and project/domain context

## Key Design Points

- **Single resource**: This module creates exactly one `networking.Network` resource
- **Root of networking DAG**: The `network_id` output is the most referenced FK in the OpenStack component family
- **Optional bool handling**: `admin_state_up` and `port_security_enabled` use Go pointer types (`*bool`) — nil means "not set by user"
- **Middleware defaults**: `admin_state_up` defaults to `true` via OpenMCF middleware before the module runs

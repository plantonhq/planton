# OpenStackFloatingIp Pulumi Module

This directory contains the Pulumi Go module for provisioning OpenStack Neutron floating IPs.

## Structure

```
iac/pulumi/
├── main.go           # Entrypoint (Pulumi program)
├── Pulumi.yaml       # Pulumi project config
├── Makefile          # Build and test targets
├── module/
│   ├── main.go       # Resources() entry point
│   ├── locals.go     # Input extraction and FK resolution
│   ├── floating_ip.go # Floating IP resource creation
│   └── outputs.go    # Output constants
└── README.md         # This file
```

## How It Works

1. The Planton CLI serializes `OpenStackFloatingIpStackInput` and passes it to Pulumi
2. `main.go` loads the stack input and calls `module.Resources()`
3. `locals.go` extracts the resolved FK values (`FloatingNetworkId`, `PortId`)
4. `floating_ip.go` creates `networking.NewFloatingIp()` with the extracted values
5. Outputs are exported matching `stack_outputs.proto` field names

## Local Development

```bash
# Build
make build

# Install plugins
make install-pulumi-plugins
```

## Key Design Notes

- **Single resource**: Creates one `networking.FloatingIp` (no separate associate resource)
- **`Pool` mapping**: The Pulumi/TF field `Pool` maps to our `floating_network_id` FK
- **Optional association**: `port_id` is only set when the FK is present (allocation-only by default)
- **FK extraction**: Required FK (`FloatingNetworkId`) always resolved; optional FK (`PortId`) nil-guarded

# AzureApplicationGateway Pulumi Module

Pulumi IaC module for provisioning an Azure Application Gateway with backend pools,
HTTP settings, listeners, routing rules, health probes, SSL certificates, and
optional WAF configuration.

## Architecture

See [overview.md](overview.md) for the complete resource flow and design decisions.

## Quick Start

```bash
# Build the module
make build

# Run tests
make test

# Run the Pulumi program (requires stack input)
make run
```

## Debugging

```bash
# Start Delve debugger
./debug.sh
```

## Key Design Notes

- **Single resource**: Creates one `network.ApplicationGateway` with all sub-components
  as nested blocks (unlike the LB which creates separate resources)
- **Auto-derived names**: Gateway IP config, frontend IP config, and frontend port
  names are automatically derived from the resource and listener names
- **V2 SKU only**: Standard_v2 or WAF_v2 (v1 is legacy)
- **Basic routing**: Only Basic rule type is supported (no path-based routing)
- **SSL via Key Vault**: Certificates reference Key Vault secrets (no PFX upload)
- **WAF**: OWASP 3.2 rule set, Detection or Prevention mode

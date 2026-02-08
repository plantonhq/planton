# OpenStackComputeKeypair Pulumi Module

Pulumi (Go) IaC module for provisioning OpenStack compute keypairs.

## Structure

```
.
├── main.go           # Entry point: loads stack input, calls module.Resources
├── module/
│   ├── main.go       # Provider setup and orchestration
│   ├── locals.go     # Data extraction from stack input
│   ├── keypair.go    # Keypair resource creation and output exports
│   └── outputs.go    # Output name constants
├── Pulumi.yaml       # Project configuration
├── Makefile           # Build and test automation
└── debug.sh          # Local testing script
```

## Usage

```bash
# Build
make build

# Install required plugins
make install-pulumi-plugins

# Test with local manifest
make test
```

## Debug

```bash
./debug.sh ../hack/manifest.yaml
```

# OpenStackVolume Pulumi Module

Provisions an OpenStack Cinder block storage volume.

## Usage

```bash
# Build the module
make build

# Install required Pulumi plugins
make install-pulumi-plugins

# Run a preview with the test manifest
make test
```

## Module Structure

```
module/
├── main.go        # Entry point (Resources function)
├── locals.go      # Local variables and FK resolution
├── outputs.go     # Output key constants
└── volume.go      # Cinder volume creation and output exports
```

## Debug

```bash
# Set stack input from manifest
export STACK_INPUT=$(cat iac/hack/manifest.yaml | base64)

# Run Pulumi preview
pulumi preview --stack test --non-interactive
```

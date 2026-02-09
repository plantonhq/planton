# OpenStackVolumeAttach Pulumi Module

Attaches an OpenStack Cinder volume to a compute instance.

## Usage

```bash
make build
make install-pulumi-plugins
make test
```

## Module Structure

```
module/
├── main.go            # Entry point (Resources function)
├── locals.go          # Local variables and FK resolution
├── outputs.go         # Output key constants
└── volume_attach.go   # Volume attachment and output exports
```

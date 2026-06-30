# HetznerCloudVolume Pulumi Module

Pulumi (Go) IaC module for provisioning Hetzner Cloud block storage volumes with optional server attachment and automount.

## Structure

```
.
├── main.go           # Entry point: loads stack input, calls module.Resources
├── module/
│   ├── main.go       # Provider setup and orchestration
│   ├── locals.go     # Data extraction from stack input, label computation
│   ├── volume.go     # Volume creation, conditional attachment; output exports
│   └── outputs.go    # Output name constants
└── BUILD.bazel       # Bazel build configuration
```

## Resources Created

- `hcloud.Volume` — Provisions a block storage volume with the specified size, location, optional filesystem format, labels, and delete protection
- `hcloud.VolumeAttachment` (conditional) — Attaches the volume to a server with optional automount, created only when `serverId` is set in the spec

## Outputs

| Name | Description |
|------|-------------|
| `volume_id` | Hetzner Cloud numeric ID of the created volume |
| `linux_device` | Linux device path for the volume on the attached server |

## Usage

```bash
# Build
bazel build //apis/dev/planton/provider/hetznercloud/hetznercloudvolume/v1/iac/pulumi:pulumi

# Test with local manifest
export STACK_INPUT=$(cat ../hack/manifest.yaml | base64)
pulumi up
```

## Debug

```bash
# Run locally against the hack manifest
export STACK_INPUT=$(cat ../hack/manifest.yaml | base64)
export HCLOUD_TOKEN="your-api-token"
pulumi up --stack dev
```

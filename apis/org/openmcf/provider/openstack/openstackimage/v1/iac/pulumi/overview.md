# OpenStackImage Pulumi Module Overview

## Architecture

Single-resource module: creates one `images.Image` from the spec.

```
OpenStackImageStackInput
  ├── target (OpenStackImage)
  │   ├── metadata.name → image name
  │   └── spec
  │       ├── container_format → required (bare, ovf, etc.)
  │       ├── disk_format → required (qcow2, raw, etc.)
  │       ├── image_source_url → HTTP URL for Glance to download
  │       ├── min_disk_gb → minimum disk requirement
  │       ├── min_ram_mb → minimum RAM requirement
  │       ├── protected → deletion protection (default: false)
  │       ├── hidden → hide from listings (default: false)
  │       ├── tags → image tags
  │       ├── visibility → access control (default: private)
  │       └── region → region override
  └── provider_config → OpenStack credentials
```

## Outputs

| Output | Source |
|--------|--------|
| `image_id` | `createdImage.ID()` |
| `name` | `createdImage.Name` |
| `checksum` | `createdImage.Checksum` |
| `size_bytes` | `createdImage.SizeBytes` |
| `status` | `createdImage.Status` |
| `file` | `createdImage.File` |
| `region` | `createdImage.Region` |

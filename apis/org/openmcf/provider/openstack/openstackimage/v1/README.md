# OpenStackImage

An OpenStack Glance image -- a virtual disk template for compute instances and bootable volumes.

## Overview

A Glance image contains an operating system, pre-installed software, and customizations needed for workloads. Images are used to boot compute instances (`OpenStackInstance`) and to initialize bootable Cinder volumes (`OpenStackVolume`).

## When to Use

- **Custom VM images**: Upload organization-specific golden images from URLs
- **Cloud image management**: Register standard cloud images (Ubuntu, CentOS, etc.) with metadata
- **Bootable volumes**: Create image-backed Cinder volumes for production root disks

## Important Notes

- **container_format and disk_format are required**: These describe the image envelope and disk data format
- **image_source_url is the primary upload mechanism**: Point to an HTTP/HTTPS URL for Glance to download
- **ForceNew fields**: container_format, disk_format, image_source_url cannot be changed after creation
- **visibility defaults to "private"**: Only the image owner can see it; use "public" (requires admin) for shared images

## Key Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `container_format` | string | Yes | Image envelope: "bare", "ovf", "docker", etc. |
| `disk_format` | string | Yes | Disk data format: "qcow2", "raw", "vmdk", etc. |
| `image_source_url` | string | No | URL to download image data from |
| `min_disk_gb` | int32 | No | Minimum disk size required to boot |
| `min_ram_mb` | int32 | No | Minimum RAM required to boot |
| `protected` | bool | No | Prevent accidental deletion (default: false) |
| `hidden` | bool | No | Hide from default listings (default: false) |
| `tags` | list | No | Tags for filtering |
| `visibility` | string | No | Access control: "private", "shared", "community", "public" |
| `region` | string | No | Region override |

## Outputs

| Output | Description |
|--------|-------------|
| `image_id` | UUID of the image (primary FK target for Volume and Instance) |
| `name` | Image name (from metadata.name) |
| `checksum` | MD5 checksum of image data |
| `size_bytes` | Size of image data in bytes |
| `status` | Lifecycle state (active, queued, etc.) |
| `file` | URL path to image data |
| `region` | OpenStack region |

## Examples

See [examples.md](examples.md) for YAML manifests.

## Terraform Resource

`openstack_images_image_v2`

## Pulumi Resource

`openstack.images.Image`

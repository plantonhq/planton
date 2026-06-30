---
title: "Image"
description: "Image deployment documentation"
icon: "package"
order: 100
componentName: "openstackimage"
---

# OpenStack Image

Deploys an OpenStack Glance image from a URL or as a metadata entry, with configurable container and disk formats, visibility controls, minimum hardware requirements, and optional deletion protection.

## What Gets Created

When you deploy an OpenStackImage resource, Planton provisions:

- **Glance Image** — an `openstack_images_image_v2` resource registered in the OpenStack Image (Glance) service. When `imageSourceUrl` is provided, Glance downloads the image data from the specified URL. The image name is derived from `metadata.name`. Visibility, deletion protection, and minimum hardware requirements are applied as configured.

## Prerequisites

- **OpenStack credentials** configured via environment variables or Planton provider config
- **A publicly accessible URL** for the image data if using `imageSourceUrl` (Glance must be able to reach it)
- **Admin role** if setting `visibility` to `public`

## Quick Start

Create a file `image.yaml`:

```yaml
apiVersion: openstack.planton.dev/v1
kind: OpenStackImage
metadata:
  name: ubuntu-2204
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.OpenstackImage.ubuntu-2204
spec:
  containerFormat: bare
  diskFormat: qcow2
  imageSourceUrl: "https://cloud-images.ubuntu.com/jammy/current/jammy-server-cloudimg-amd64.img"
```

Deploy:

```shell
planton apply -f image.yaml
```

This uploads an Ubuntu 22.04 cloud image to Glance with `bare` container format and `qcow2` disk format. The image is private by default.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `containerFormat` | `string` | Container or envelope format of the image. ForceNew: cannot be changed after creation. | One of: `bare`, `ovf`, `aki`, `ari`, `ami`, `ova`, `docker`, `compressed` |
| `diskFormat` | `string` | Format of the disk image data. ForceNew: cannot be changed after creation. | One of: `raw`, `vhd`, `vhdx`, `vmdk`, `vdi`, `iso`, `ploop`, `qcow2`, `aki`, `ari`, `ami` |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `imageSourceUrl` | `string` | — | HTTP or HTTPS URL from which Glance downloads the image data. ForceNew: cannot be changed after creation. If omitted, the image is created as a metadata entry only and the data must be uploaded separately (e.g., via the `glance` CLI). |
| `minDiskGb` | `int` | `0` | Minimum disk size in GB required to boot this image. Zero means no minimum. |
| `minRamMb` | `int` | `0` | Minimum RAM in MB required to boot this image. Zero means no minimum. |
| `protected` | `bool` | `false` | Prevents the image from being deleted. Set to `true` for production images that should not be accidentally removed. |
| `hidden` | `bool` | `false` | Hides the image from default listing queries. Hidden images are still accessible by UUID. |
| `tags` | `string[]` | `[]` | Tags for filtering images in Glance. Example: `["ubuntu", "22.04", "cloud-init"]`. |
| `visibility` | `string` | `private` | Controls who can see and use this image. `private`: only the owner project. `shared`: owner can share with specific projects. `community`: visible to all but not in default listings. `public`: visible and usable by all (requires admin). |
| `region` | `string` | provider default | Overrides the region from the provider config for this image. |

## Examples

### Cloud Image from URL

Upload an Ubuntu 22.04 cloud image from the official Canonical repository:

```yaml
apiVersion: openstack.planton.dev/v1
kind: OpenStackImage
metadata:
  name: ubuntu-2204
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.OpenstackImage.ubuntu-2204
spec:
  containerFormat: bare
  diskFormat: qcow2
  imageSourceUrl: "https://cloud-images.ubuntu.com/jammy/current/jammy-server-cloudimg-amd64.img"
  tags:
    - ubuntu
    - "22.04"
    - cloud-init
```

### Protected Golden Image with Hardware Requirements

A production golden image with deletion protection, minimum hardware requirements, and shared visibility so it can be distributed to specific projects:

```yaml
apiVersion: openstack.planton.dev/v1
kind: OpenStackImage
metadata:
  name: golden-rhel9
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.OpenstackImage.golden-rhel9
spec:
  containerFormat: bare
  diskFormat: qcow2
  imageSourceUrl: "https://images.internal.example.com/rhel9-golden-20250101.qcow2"
  minDiskGb: 20
  minRamMb: 2048
  protected: true
  visibility: shared
  tags:
    - rhel
    - "9"
    - golden
    - hardened
```

### ISO Image for Installation Media

Register an ISO image for use as installation media, hidden from default listings:

```yaml
apiVersion: openstack.planton.dev/v1
kind: OpenStackImage
metadata:
  name: debian-12-netinst
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.OpenstackImage.debian-12-netinst
spec:
  containerFormat: bare
  diskFormat: iso
  imageSourceUrl: "https://cdimage.debian.org/debian-cd/current/amd64/iso-cd/debian-12.9.0-amd64-netinst.iso"
  hidden: true
  tags:
    - debian
    - "12"
    - installer
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `image_id` | `string` | UUID of the created Glance image. Used as a foreign key by OpenStackInstance and OpenStackVolume. |
| `name` | `string` | Name of the image, derived from `metadata.name`. |
| `checksum` | `string` | MD5 checksum of the image data, computed by Glance after upload. |
| `size_bytes` | `int64` | Size of the image data in bytes, computed by Glance after upload. |
| `status` | `string` | Current lifecycle state of the image (e.g., `active`, `queued`, `saving`). |
| `file` | `string` | URL path to the image data in the Glance store (e.g., `/v2/images/<uuid>/file`). |
| `region` | `string` | OpenStack region where the image was created. |

## Related Components

- [OpenStackInstance](/docs/catalog/openstack/instance) — boots compute instances from this image via `imageName` or `imageId`
- [OpenStackVolume](/docs/catalog/openstack/volume) — creates bootable Cinder volumes from this image

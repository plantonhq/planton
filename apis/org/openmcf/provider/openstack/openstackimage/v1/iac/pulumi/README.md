# OpenStackImage Pulumi Module

This Pulumi module provisions an OpenStack Glance image.

## Resources Created

- `openstack_images_image_v2` -- Glance image with optional URL-based upload

## Usage

This module is invoked by the OpenMCF CLI. For local development:

```bash
make build
make test
```

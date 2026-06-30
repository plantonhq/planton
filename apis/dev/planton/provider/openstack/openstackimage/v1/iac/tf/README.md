# OpenStackImage Terraform Module

This Terraform module provisions an OpenStack Glance image.

## Resources Created

- `openstack_images_image_v2` -- Glance image with optional URL-based upload

## Inputs

| Variable | Description |
|----------|-------------|
| `metadata` | Resource metadata (name, labels, etc.) |
| `spec` | OpenStackImageSpec configuration |

## Outputs

| Output | Description |
|--------|-------------|
| `image_id` | UUID of the image |
| `name` | Image name |
| `checksum` | MD5 checksum |
| `size_bytes` | Image size in bytes |
| `status` | Lifecycle status |
| `file` | URL path to image data |
| `region` | OpenStack region |

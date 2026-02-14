---
title: "OpenStackImage Research Documentation"
description: "OpenStackImage Research Documentation deployment documentation"
icon: "package"
order: 100
componentName: "openstackimage"
---

# OpenStackImage Research Documentation

## Terraform Resource: `openstack_images_image_v2`

### Provider Source

- **Package**: `openstack`
- **Resource**: `resource_openstack_images_image_v2.go`
- **Provider**: terraform-provider-openstack v3.x

### Schema (Complete)

| Field | Type | Required | Computed | ForceNew | Default | Description |
|-------|------|----------|----------|----------|---------|-------------|
| `region` | string | Optional | Yes | Yes | - | OpenStack region |
| `container_format` | string | Required | No | Yes | - | Image container format |
| `disk_format` | string | Required | No | Yes | - | Disk data format |
| `name` | string | Required | No | No | - | Image name |
| `min_disk_gb` | int | Optional | No | No | `0` | Minimum disk in GB |
| `min_ram_mb` | int | Optional | No | No | `0` | Minimum RAM in MB |
| `protected` | bool | Optional | No | No | `false` | Deletion protection |
| `hidden` | bool | Optional | No | No | `false` | Hide from listings |
| `tags` | set(string) | Optional | No | No | - | Image tags |
| `visibility` | string | Optional | No | No | `"private"` | Access control |
| `properties` | map | Optional | Yes | No | - | Glance properties |
| `image_source_url` | string | Optional | No | Yes | - | URL to download image |
| `local_file_path` | string | Optional | No | Yes | - | Local file path |
| `web_download` | bool | Optional | No | No | - | Use web-download import |
| `image_cache_path` | string | Optional | No | No | `$HOME/.terraform/image_cache` | Local cache |
| `verify_checksum` | bool | Optional | No | No | - | Verify MD5 checksum |
| `decompress` | bool | Optional | No | Yes | - | Decompress before upload |
| `image_id` | string | Optional | Yes | Yes | - | UUID override |

### Fields Excluded (80/20 Analysis)

| Field | Reason |
|-------|--------|
| `local_file_path` | Impractical for pipeline-based IaC (file doesn't exist on CI runner) |
| `web_download` | Niche; requires specific Glance backend support |
| `image_source_username/password` | Authenticated downloads are rare |
| `verify_checksum` | TF-specific operational behavior |
| `decompress` | TF-specific operational behavior |
| `image_cache_path` | TF-specific local caching |
| `image_id` | UUID override is extremely rare |
| `properties` | Computed/merged with server-side; complex to manage declaratively |

### Computed Attributes

`checksum`, `created_at`, `metadata`, `owner`, `schema`, `size_bytes`, `status`, `updated_at`, `file`, `image_id`

### Pulumi SDK

- **Package**: `github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack/images`
- **Function**: `images.NewImage()`
- **Args**: `images.ImageArgs`

### Key Behaviors

1. **Image upload**: When `image_source_url` is set, Glance downloads the image from the URL
2. **ForceNew**: container_format, disk_format, image_source_url, local_file_path, decompress, region
3. **Checksum**: Computed after upload, used for integrity verification
4. **Visibility**: "public" requires admin role; "shared" requires explicit member management
5. **Protected images**: Cannot be deleted until protection is removed

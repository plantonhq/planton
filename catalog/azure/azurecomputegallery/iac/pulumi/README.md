# AzureComputeGallery Pulumi Module

## Overview

Creates an Azure Compute Gallery -- the shared library an organization keeps its approved VM images in. Image definitions (AzureComputeGalleryImage) live inside it; VMs and scale sets deploy from their published, region-replicated versions.

## Resources Created

- `compute.SharedImageGallery` -- the gallery (description, optional sharing profile, tags)

## Outputs

- `gallery_id` -- the gallery's ARM resource ID
- `gallery_name` -- the gallery's name (what image definitions reference)
- `unique_name` -- the globally unique name Azure assigns the gallery
- `community_gallery_name` -- the public community name (empty unless Community-shared)

## Behavior Notes

- **The entire sharing tree is fixed at creation** -- changing any part of it forces replacement; only description and tags update in place.
- **Community sharing requires the community_gallery block** (the spec's CEL front-loads the provider's expand-time check); Azure generates the public name from the prefix.
- **Gallery names forbid dashes** (letters, numbers, dots, underscores; up to 80 characters) -- unusual for Azure names.
- **A gallery is free at rest** -- image versions bill for the storage their replicas consume, the gallery itself bills nothing.

## Usage

The module is executed by the Planton platform with an IaC input containing the target `AzureComputeGallery` resource and an Azure provider configuration. For a manifest example, see `../../e2e/manifest.yaml`.

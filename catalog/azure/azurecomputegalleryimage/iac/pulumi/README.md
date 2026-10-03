# AzureComputeGalleryImage Pulumi Module

## Overview

Creates a gallery image: one image definition inside an Azure Compute Gallery (marketplace-style identity, OS type, security posture, recommended sizing) plus its published versions, each replicated to its own target regions.

## Resources Created

- `compute.SharedImage` -- the image definition
- `compute.SharedImageVersion` -- one per entry in `spec.versions`, keyed by version name (the composed release artifacts VMs actually deploy)

## Outputs

- `image_id` -- the definition's ARM resource ID (deploying from it gets the latest version)
- `image_name` -- the definition's name within its gallery
- `version_ids` -- map of version name to ARM ID (pin VMs to an exact release)

## Behavior Notes

- **The security flags fire on argument PRESENCE**: the four trusted-launch/confidential-VM flags are a provider ConflictsWith clique, so each is sent ONLY when true -- an explicit false alongside another flag is provider-rejected. The spec's CEL enforces at most one.
- **Each version has exactly one source** (blob + storage account, OS disk snapshot, or managed image/VM) -- the spec's CEL mirrors the provider's ExactlyOneOf.
- **Almost the whole definition is create-only**; updatable in place: description, disk-type exclusions, recommended sizing, release notes, end-of-life, tags. On versions: target regions, exclude-from-latest, end-of-life, tags.
- **Clearing end_of_life_date forces replacement** on both the definition and a version (the provider's CustomizeDiff).
- **A target region's storage_account_type cannot be updated**: the API rejects it and the provider cannot force replacement for it (region-list membership changes in place) -- remove and re-add the region instead.
- **The classic SDK pluralizes one field name** (`diskTypesNotAlloweds` for the provider's `disk_types_not_allowed`) -- same argument underneath, an engine-shape note only.
- **The definition is free**; each version bills for the storage its regional replicas consume.

## Usage

The module is executed by the Planton platform with an IaC input containing the target `AzureComputeGalleryImage` resource and an Azure provider configuration. For a manifest example, see `../../e2e/manifest.yaml`.

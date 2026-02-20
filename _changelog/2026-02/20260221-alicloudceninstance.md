# AlicloudCenInstance

**Date**: 2026-02-21
**Type**: New Component
**Enum**: 3130
**ID Prefix**: accen

## Summary

Added `AlicloudCenInstance` -- Alibaba Cloud Cloud Enterprise Network (CEN) instance with bundled child-instance attachments. CEN is a global networking service that provides private connectivity between VPCs across regions and between VPCs and on-premises data centers.

## What's Included

- Proto API definitions (spec.proto, api.proto, stack_input.proto, stack_outputs.proto)
- Spec validation tests (21 tests covering valid/invalid inputs)
- Pulumi module (Go) using `cen.Instance` + `cen.InstanceAttachment`
- Terraform module (HCL) with `alicloud_cen_instance` + `alicloud_cen_instance_attachment`
- Documentation (README.md, examples.md, catalog-page.md, docs/README.md)
- 3 presets (multi-VPC same region, cross-region backbone, managed VPC references)
- Registered in cloud_resource_kind.proto and kind_map_gen.go

## Design Decisions

- Used `cen_instance_name` (provider-authentic) instead of deprecated `name` field
- Renamed `vpc_attachments` to `attachments` for accuracy (supports VPC, VBR, CCN)
- Added `protection_level`, `resource_group_id`, `tags` fields beyond T02 spec
- Cross-account attachment support deferred to v2
- CEN `region` field is for API routing only (CEN is a global resource)

## Milestone

This is the 30th and final Alibaba Cloud resource kind, completing the Start phase for all resources in the T02 queue.

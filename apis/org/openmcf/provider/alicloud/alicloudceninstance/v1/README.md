# AlicloudCenInstance

Manages an Alibaba Cloud Cloud Enterprise Network (CEN) instance with bundled child-instance attachments.

## Overview

CEN is a global networking service that provides high-quality, low-latency private connectivity between VPCs in different regions, or between VPCs and on-premises data centers via Virtual Border Routers (VBR). Unlike most Alibaba Cloud resources, CEN is region-agnostic -- the instance itself is not bound to a single region.

### What Gets Created

- **CEN Instance** -- a global Cloud Enterprise Network that acts as a hub for inter-network connectivity
- **Child-Instance Attachments** -- one per attachment entry, connecting a VPC, VBR, or CCN to the CEN hub

### How CEN Works

CEN creates a private backbone between attached networks. Once two VPCs in different regions are attached to the same CEN instance, they can communicate over Alibaba Cloud's internal network without traversing the public internet. This provides:

- **Cross-region connectivity** -- VPCs in cn-hangzhou and us-west-1 communicate privately
- **Hybrid cloud connectivity** -- VPCs connect to on-premises networks via VBR attachments
- **Multi-VPC architecture** -- Production, staging, and shared-services VPCs in the same region interconnect

### Region Field

The `region` field in this component is used for Alibaba Cloud API routing only. CEN is a global resource -- it does not reside in any single region. Each attachment declares its own region via `childInstanceRegionId`.

## Build and Test (Localized)

All build and test commands are scoped to this component directory. Never run project-wide `make build`.

```bash
# Proto compilation (from openmcf repo root, once after proto changes)
make protos

# Go build (Pulumi module)
go build ./apis/org/openmcf/provider/alicloud/alicloudceninstance/v1/iac/pulumi/...

# Go vet
go vet ./apis/org/openmcf/provider/alicloud/alicloudceninstance/v1/iac/pulumi/...

# Spec tests
go test ./apis/org/openmcf/provider/alicloud/alicloudceninstance/v1/...

# Terraform validation
cd apis/org/openmcf/provider/alicloud/alicloudceninstance/v1/iac/tf
terraform init -backend=false
terraform validate
```

## Configuration Reference

See [catalog-page.md](catalog-page.md) for complete field documentation, examples, and presets.

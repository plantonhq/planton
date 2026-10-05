# Overview

The AwsLaunchTemplate API resource provisions an EC2 launch template: the
versioned, reusable blueprint that describes how to launch an instance --
AMI, instance type or attribute-based requirements, storage, networking,
IAM identity, metadata-service posture, and purchase options.

## Why We Created This API Resource

The launch template is the composition anchor of EC2 fleet compute. It has
its own lifecycle and is referenced from many places at once: auto-scaling
groups (directly or through mixed-instances overrides), EKS managed node
groups, AWS Batch compute environments, and EC2 Fleet. Modeling it as a
first-class kind -- instead of burying launch configuration inside
every fleet definition -- lets you:

- **Define the golden image once**: an org-wide hardened template (IMDSv2
  enforced, encrypted gp3 volumes, detailed monitoring) that every fleet,
  node group, and batch environment references by ID.
- **Roll fleets by changing one node**: every template change creates a new
  immutable version and promotes it to the default, so groups following
  `$Default` pick it up on their next launch or instance refresh -- while
  anything pinned to a numeric version keeps exactly what it tested.
- **Describe compute by attributes, not names**: `instanceRequirements`
  expresses vCPU/memory ranges, CPU manufacturers, accelerators, and price
  protection so AWS resolves matching instance types at launch -- the
  foundation of Spot diversification.

## Key Features

### Compute Selection

- **Exact type or attribute-based**: name an `instanceType`, or describe
  requirements (memory/vCPU ranges required; generations, manufacturers,
  accelerator, local-storage, and price-protection filters) and let AWS
  pick.
- **Purchase options**: On-Demand by default; `spotOptions` turns every
  launch into a Spot request with interruption behavior and price ceiling;
  `marketType: capacity-block` consumes pre-purchased Capacity Blocks for
  ML with `capacityReservation` naming the block.
- **Reserved capacity**: `capacityReservation` targets On-Demand Capacity
  Reservations by ID or resource group, or shapes reservation use with a
  preference (`open`, `none`, `capacity-reservations-only`).
- **Dedicated Hosts**: `placement` pins BYOL workloads to a host
  (`hostId`), a License Manager host resource group, and host affinity;
  `licenseConfigurationArns` tracks the licenses launches consume.
- **Partial templates**: leave AMI and type unset when the consumer (an EKS
  node group, an ASG override) supplies them.

### Storage and Networking

- **Block device mappings**: grow or retype the root volume, encrypt with a
  KMS key reference, attach data volumes from snapshots -- pre-warmed at a
  paid `volumeInitializationRateMibps` instead of loading lazily --
  suppress AMI-baked devices.
- **Explicit network interfaces**: public-IP tri-state, static IPs, IPv6
  (including `primaryIpv6` for stable IPv6 identity), prefix delegation
  (the Kubernetes CNI pattern), EFA for HPC/ML, connection-tracking
  timeout tuning, ENA Express (SRD) with its UDP option, ENA queue counts,
  Wavelength carrier IPs, and security groups per interface referencing
  `AwsSecurityGroup` nodes.
- **Secondary interfaces**: additional interfaces attached at launch for
  multi-homed instances (e.g. a dedicated storage-subnet leg), referencing
  `AwsSubnet` nodes.
- **Bandwidth weighting**: bias baseline bandwidth toward VPC networking
  or EBS throughput on supported types -- free performance shaping.

### Security Posture

- **IMDSv2 enforcement** (`metadataOptions.httpTokens: required`) -- the
  single most effective hardening against credential-stealing SSRF.
- **Instance identity by reference**: `instanceProfile` composes with
  `AwsIamInstanceProfile`; EBS keys with `AwsKmsKey`.
- **Nitro Enclaves, hibernation, stop/termination protection, CPU
  SEV-SNP, nested virtualization** for hardened and specialized
  workloads.
- **Launch-time tagging built in**: the module renders the platform's
  identity tags onto the instances and volumes every launch creates, so
  fleet members never escape cost-allocation or orphan-cleanup queries.

## Benefits

- **Composability**: auto-scaling groups, node groups, and Batch reference
  the template through `valueFrom`, so the architecture graph shows exactly
  which fleets launch from which blueprint.
- **Safe rollouts**: immutable versions + default promotion turn AMI
  rotation into a reviewable, reversible change.
- **Consistency**: identical behavior across Terraform and Pulumi.

## Outputs

- `launch_template_id`: template ID (what ASGs, EKS node groups, and Batch reference)
- `launch_template_arn`: ARN, for IAM policies that scope ec2:RunInstances to approved templates
- `latest_version`: newest version number
- `default_version`: version consumers referencing `$Default` launch from

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

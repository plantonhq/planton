# OCI Compute Instance: Design Rationale and Research

## Introduction

The OciComputeInstance component is the fundamental workload primitive for OCI in Planton. Every application server, CI runner, container host, and database node on Oracle Cloud runs on a compute instance. This is also the most complex OCI component to date — 18 top-level spec fields, 9 nested messages, 6 enums — because OCI's compute API exposes a deep configuration surface spanning hardware selection, networking, security, and agent management.

Getting the spec shape right determines whether users can express their compute requirements declaratively without escaping to raw Terraform or Pulumi. This document explains the design decisions behind the OciComputeInstance component and the research that informed them.

## The OCI Compute Model

### Shapes: How OCI Sizes Instances

OCI uses a **shape** to define the hardware profile for an instance. Shapes fall into several categories:

| Category | Example Shapes | How Sizing Works |
|----------|---------------|------------------|
| **Flex** | VM.Standard.E4.Flex, VM.Standard.A1.Flex, VM.Standard.E5.Flex | User specifies OCPUs and memory via `shapeConfig`. The shape name selects the processor family (AMD, Arm, Intel). |
| **Burstable** | VM.Standard.E4.Flex with `baselineOcpuUtilization` | Flex shapes with a reduced baseline (1/8, 1/2, or full). The instance can burst above baseline when needed. Cheaper than full flex. |
| **Fixed Standard** | VM.Standard2.1, VM.Standard2.4 | Predetermined CPU and memory. Being phased out in favor of flex shapes. |
| **Bare Metal** | BM.Standard3.64, BM.DenseIO.E4.128 | Entire physical server. No hypervisor overhead. Used for databases, HPC, and licensing isolation. |
| **GPU** | BM.GPU4.8, VM.GPU.A10.1 | GPU-attached instances for ML/AI and graphics workloads. |
| **Dense IO** | BM.DenseIO.E4.128 | Bare metal with local NVMe SSDs for high-IOPS workloads. |

**Flex shapes are the modern default.** OCI has been moving toward flex shapes as the primary offering, with fixed shapes being legacy. The Planton spec reflects this by making `shapeConfig` a first-class field rather than burying it in a generic options map.

### Availability Domains and Fault Domains

OCI's physical topology has two layers:

1. **Availability Domains (ADs)** — Isolated data centers within a region. Equivalent to AWS AZs. Some OCI regions have 1 AD; larger regions have 3. The `availabilityDomain` field is required on every instance.

2. **Fault Domains** — Within each AD, OCI provides 3 fault domains with independent power and networking. Instances in different fault domains survive single-rack failures. The `faultDomain` field is optional; OCI auto-distributes when unspecified.

The availability domain format includes a tenancy-specific prefix (e.g., `Ixxj:US-ASHBURN-AD-1`). This prefix varies by tenancy, which means availability domain values are not portable across tenancies. This is an OCI design choice that prevents hardcoded AD references from accidentally working in the wrong tenancy.

## Why Shape and ShapeConfig Are Separate Fields

An alternative design would combine shape and sizing into a single nested message. The split into `shape` (string) and `shapeConfig` (message) reflects how the OCI API actually works:

1. **`shape` selects the hardware family.** It determines the processor type (AMD, Arm, Intel), whether the instance is virtual or bare metal, and whether flex sizing is available. This is a string because OCI adds new shapes regularly, and an enum would require proto changes for every new shape.

2. **`shapeConfig` specifies the resource allocation within flex shapes.** OCPUs, memory, baseline utilization, and NVMe count are only meaningful for flex shapes. Fixed and bare metal shapes ignore `shapeConfig`.

3. **The split maps to OCI's API.** The Terraform `oci_core_instance` resource has `shape` as a top-level string and `shape_config` as a nested block. The Pulumi SDK mirrors this. Matching the provider API shape reduces the cognitive gap for users familiar with OCI's native tools.

**Why not make `shapeConfig` required?** Because fixed and bare metal shapes don't use it. Making it optional with a documentation note ("required for flex shapes") matches the actual constraint without adding unnecessary validation complexity for non-flex shapes.

## Why SourceDetails Is Required and Inline

Every compute instance needs a boot source — there's no default image. The `sourceDetails` message is required because:

1. **No sensible default exists.** Unlike a VCN where a default CIDR might work for exploration, an instance without a boot image cannot start. The image determines the OS, kernel version, and pre-installed software. Choosing it for the user would be presumptuous.

2. **Two boot modes.** The `sourceType` enum (`image` vs `boot_volume`) covers both creating a new instance from a platform/custom image and cloning an existing boot volume. These are distinct workflows with different field requirements (`bootVolumeSizeInGbs` is only meaningful for images).

3. **Boot volume configuration is immutable after launch.** The boot volume size and VPUs-per-GB are set at instance creation and cannot be changed without creating a new boot volume. Inline configuration (rather than a separate boot volume resource) reflects this lifecycle coupling.

### Why kmsKeyId Is in SourceDetails

The KMS key for boot volume encryption is part of `sourceDetails` rather than a top-level field because it applies specifically to the boot volume created from the source. If the `sourceType` is `boot_volume` (cloning), the KMS key of the source volume is inherited, and a new `kmsKeyId` here would re-encrypt the clone.

## Why VNIC Details Are Inline

The primary VNIC is declared inline in `createVnicDetails` rather than as a separate resource. This matches OCI's architecture:

1. **The primary VNIC is created with the instance and cannot be detached.** Unlike AWS where all ENIs (including the primary) are independently attachable, OCI's primary VNIC is permanently bound to the instance. Its lifecycle is inseparable from the instance lifecycle.

2. **The primary VNIC determines the instance's network identity.** The subnet, NSGs, IP addresses, and DNS hostname are all part of the instance's network placement. Declaring them separately would split a single conceptual unit across two resources.

3. **Secondary VNICs are different.** OCI supports attaching additional VNICs to an instance, and those are independently detachable. If/when Planton adds secondary VNIC support, it would be a separate resource — matching the different lifecycle.

### Why nsgIds Has a Max of 5

OCI imposes a platform limit of 5 NSGs per VNIC. This is enforced via `buf.validate` with `max_items = 5`. The limit exists because OCI evaluates all NSG rules for every VNIC, and more NSGs means more rules to evaluate per packet. Five NSGs with focused rule sets is sufficient for most architectures (e.g., base rules, application rules, monitoring rules, environment rules, team rules).

## The assignPublicIp Tri-State

The `assignPublicIp` field is an `optional bool` in the proto, creating three states:

| Value | Behavior |
|-------|----------|
| Unset | Use subnet default — public subnets assign a public IP; private subnets do not |
| `true` | Assign a public IP regardless of subnet type |
| `false` | Do not assign a public IP regardless of subnet type |

**Why not a regular bool?** A regular bool defaults to `false`, which would change behavior for public subnets (suppressing the public IP that users expect). The `optional` wrapper preserves the OCI default behavior when the user doesn't specify a preference.

**Implementation note:** The OCI API accepts this field as a string (`"true"`, `"false"`, not set). The Pulumi module converts the optional bool to a string pointer, and the Terraform module uses `tostring()`. The proto uses `optional bool` because it provides a better user experience than asking users to write `"true"` as a string in YAML.

## Why LaunchOptions Uses Strings for bootVolumeType and networkType

The `bootVolumeType` and `networkType` fields in `LaunchOptions` are strings rather than enums. This is a deliberate trade-off caused by a protobuf constraint:

**The problem:** Both `bootVolumeType` and `networkType` include the values `VFIO` and `PARAVIRTUALIZED`. Protobuf's C++ scoping rules prevent two enums within the same enclosing message from having values with the same name. Since `LaunchOptions` is nested inside `OciComputeInstanceSpec`, the enum values would collide.

**Alternatives considered:**

| Approach | Why Rejected |
|----------|-------------|
| Prefix enum values (e.g., `boot_volume_type_vfio`) | Ugly in YAML, violates the lowercase-no-prefix convention |
| Separate the enums into different proto files | Adds complexity for two rarely-used fields |
| Move LaunchOptions to a top-level message | Breaks the nested-in-spec convention used everywhere else |
| Use strings | Simple, works, these fields are rarely set by users |

**Why strings are acceptable here:** `LaunchOptions` is an advanced configuration that most users never touch. OCI automatically selects appropriate boot volume and network types based on the image and shape. The fields exist for users who need to override defaults (e.g., forcing paravirtualized networking for maximum throughput on a shape that defaults to E1000).

## Why PlatformConfig Includes the Full Field Set

The `PlatformConfig` message includes 11 fields — security features (Secure Boot, Measured Boot, TPM, memory encryption) and bare-metal-only hardware controls (NUMA, SMT, core percentage, nested virtualization, IOMMU, ACS).

**Why not split into VMPlatformConfig and BMPlatformConfig?** The OCI API uses a single `platform_config` block with a `type` discriminator. Fields that don't apply to the chosen platform type are silently ignored by the OCI API. Mirroring this in the proto:

- Avoids a complex `oneof` that would make YAML manifests harder to write
- Matches the OCI API shape that users familiar with Terraform/Pulumi already know
- Delegates platform-type validation to the OCI API, which has the most up-to-date rules for which fields apply to which types

The `type` field is required when `platformConfig` is set. It must match the instance's shape family — `amd_vm` for AMD VM shapes, `intel_vm` for Intel VM shapes, `intel_icelake_bm` for Intel Icelake bare metal, etc. This mapping is documented in the catalog page and enforced server-side by OCI.

## Why Metadata Is a Map, Not Structured Fields

The `metadata` field is a `map<string, string>` rather than structured fields for SSH keys and cloud-init. This matches OCI's API where metadata is a free-form key-value store:

- **Extensibility:** OCI and custom images can define additional metadata keys beyond `ssh_authorized_keys` and `user_data`. A structured type would require proto changes for each new key.
- **Simplicity:** SSH keys are strings, cloud-init is a base64 string. There's no validation benefit to wrapping them in dedicated message types.
- **Familiarity:** Terraform and Pulumi users already know the `metadata` map pattern. Changing the interface would create unnecessary friction.

The commonly used keys are documented in the catalog page and examples rather than enforced in the proto schema.

## What's Deferred

Based on the 80/20 principle, the following features are not in the initial implementation:

- **Secondary VNICs** — Additional network interfaces attached to an instance. These have an independent lifecycle (can be detached/reattached) and would be a separate resource type if/when needed.
- **Extended Metadata** — OCI supports `extended_metadata` as nested JSON objects beyond the flat key-value `metadata` map. Rarely used; the flat map covers the standard use cases (SSH keys, cloud-init).
- **HPC Cluster Networking** — Cluster networking for RDMA-enabled bare metal instances. This is a specialized use case involving cluster networks and instance pools, not single-instance deployment.
- **PXE Boot and iPXE Scripts** — Network boot configuration for bare metal instances. Niche use case for data center migration and custom boot workflows.
- **Image Launch Mode Overrides** — Overriding the default launch mode (paravirtualized vs native) for specific images. The image default is correct for the vast majority of cases.
- **BYOL Licensing Configuration** — License type fields for bringing your own Windows or Oracle DB licenses. Handled at the OCI subscription level; instance-level licensing fields are rarely needed when using dedicated hosts.
- **Instance Pools and Configurations** — Auto-scaling groups of instances. These are higher-order constructs that would be separate resource types (OciInstancePool, OciInstanceConfiguration) if/when added.
- **Shielded Instance Verification** — Retrieving and verifying the Measured Boot measurements from the TPM. This is a read operation (not a deployment operation) belonging in a management plane.

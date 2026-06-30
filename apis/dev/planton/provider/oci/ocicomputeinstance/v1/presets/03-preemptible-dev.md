# Preemptible Dev Instance

This preset creates a cost-optimized preemptible (spot-like) OCI compute instance for development, testing, and CI workloads. Preemptible instances use the same shapes and images as on-demand instances but are significantly cheaper because OCI can reclaim them when capacity is needed. The boot volume is preserved on preemption so work-in-progress is not lost.

## When to Use

- Development and experimentation environments where occasional interruption is acceptable
- CI/CD build agents and test runners that can be restarted automatically
- Batch processing jobs that checkpoint progress and can resume after preemption
- Cost-sensitive workloads where saving 50%+ on compute is more important than guaranteed uptime

## Key Configuration Choices

- **Preemptible with boot volume preserved** (`preemptibleInstanceConfig.preserveBootVolume: true`) -- The instance can be terminated by OCI at any time when capacity is needed. Setting `preserveBootVolume: true` keeps the boot disk intact so you can relaunch from where you left off. Without this, both the instance and its boot volume are destroyed on preemption.
- **1 OCPU / 8 GiB memory** (`shapeConfig.ocpus: 1`, `shapeConfig.memoryInGbs: 8`) -- Minimal allocation for development. The 1:8 ratio (half of the standard 1:16) is sufficient for most dev tooling, compilers, and test suites. Scale up if your build process requires more memory.
- **Default boot volume size** (no `bootVolumeSizeInGbs` set) -- Uses the image's default minimum size, avoiding paying for unused storage on a dev instance. Override if your workflow generates large artifacts.
- **Public IP for easy SSH** (`createVnicDetails.assignPublicIp: true`) -- Development instances need quick SSH access without setting up a Bastion or VPN. Place in a public subnet with an Internet Gateway.
- **No security hardening, agent config, or availability config** -- Intentionally omitted. Dev instances do not need in-transit encryption, IMDS lockdown, or live migration preferences. This keeps the preset minimal and fast to deploy. Use preset 02-private-backend when graduating a workload to production.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the instance will be created | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |
| `<availability-domain>` | Availability domain name (e.g., `Ixxj:US-ASHBURN-AD-1`) | OCI Console > Compute > Instances > Create Instance, or `oci iam availability-domain list` |
| `<image-ocid>` | OCID of the OS image to boot from | OCI Console > Compute > Custom Images, or `oci compute image list --compartment-id <tenancy-ocid>` |
| `<subnet-ocid>` | OCID of a public subnet for the primary VNIC | OCI Console > Networking > VCNs > Subnets, or `OciSubnet` status outputs |
| `<ssh-public-key>` | SSH public key content (e.g., `ssh-rsa AAAA...`) | Your local `~/.ssh/id_rsa.pub` or equivalent |

## Related Presets

- **01-general-purpose-flex** -- Use instead for on-demand instances that are not subject to preemption
- **02-private-backend** -- Use instead for production workloads in a private subnet with security hardening

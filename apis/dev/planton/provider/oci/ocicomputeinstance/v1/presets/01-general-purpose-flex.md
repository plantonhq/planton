# General-Purpose Flex Instance

This preset creates a general-purpose OCI compute instance using the VM.Standard.E4.Flex shape (AMD EPYC). It configures 1 OCPU with 16 GiB of memory, a 50 GiB boot volume, and a public IP for direct SSH access. This is the standard starting point for the vast majority of OCI compute workloads and should be the default choice unless you have specific requirements for private networking, preemptible pricing, or security hardening.

## When to Use

- Web servers, API backends, and general application hosting
- First-time OCI compute deployments where you need SSH access and a public endpoint
- Prototyping and validating workloads before moving to a production-hardened configuration
- Any workload that needs a flexible, right-sizeable VM with straightforward public internet access

## Key Configuration Choices

- **VM.Standard.E4.Flex shape** (`shape: VM.Standard.E4.Flex`) -- AMD EPYC (Milan), the most popular general-purpose shape on OCI. Flex shapes let you pick exact OCPUs and memory instead of choosing from fixed T-shirt sizes. To switch to Arm, change to `VM.Standard.A1.Flex` (same shapeConfig works).
- **1 OCPU / 16 GiB memory** (`shapeConfig.ocpus: 1`, `shapeConfig.memoryInGbs: 16`) -- OCI's standard 1:16 OCPU-to-memory ratio for E4 Flex. Each OCPU maps to a physical core with SMT, so 1 OCPU gives 2 vCPUs. Scale up by increasing both values; E4 Flex supports up to 64 OCPUs.
- **Image boot source** (`sourceDetails.sourceType: image`) -- Boots from a platform or custom image. Use the OCI Console or CLI to find the OCID for your desired OS (e.g., Oracle Linux 8, Ubuntu 22.04) in your target region.
- **50 GiB boot volume at Balanced performance** (`bootVolumeSizeInGbs: 50`, `bootVolumeVpusPerGb: 10`) -- 50 GiB is sufficient for most OS images plus application binaries. 10 VPUs/GB is the Balanced tier; increase to 20 for Higher Performance or 30-120 for Ultra High Performance if your workload is boot-volume I/O intensive.
- **Public IP assigned** (`createVnicDetails.assignPublicIp: true`) -- Enables direct SSH and HTTP access from the internet. The instance must be in a public subnet with an Internet Gateway. Use preset 02-private-backend instead if the instance should not be directly reachable.
- **SSH key via metadata** (`metadata.ssh_authorized_keys`) -- The standard OCI mechanism for injecting SSH public keys into the instance at launch. Supports multiple keys separated by newlines.
- **Hostname label set** (`createVnicDetails.hostnameLabel: myinstance`) -- Registers a DNS hostname within the subnet's DNS domain, enabling resolution like `myinstance.<subnet-dns-label>.<vcn-dns-label>.oraclevcn.com`.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the instance will be created | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |
| `<availability-domain>` | Availability domain name (e.g., `Ixxj:US-ASHBURN-AD-1`) | OCI Console > Compute > Instances > Create Instance, or `oci iam availability-domain list` |
| `<image-ocid>` | OCID of the OS image to boot from | OCI Console > Compute > Custom Images, or `oci compute image list --compartment-id <tenancy-ocid>` |
| `<subnet-ocid>` | OCID of a public subnet for the primary VNIC | OCI Console > Networking > VCNs > Subnets, or `OciSubnet` status outputs |
| `<ssh-public-key>` | SSH public key content (e.g., `ssh-rsa AAAA...`) | Your local `~/.ssh/id_rsa.pub` or equivalent |

## Related Presets

- **02-private-backend** -- Use instead for instances that should not have a public IP and sit behind a load balancer or serve internal traffic only
- **03-preemptible-dev** -- Use instead for cost-optimized dev/test instances that tolerate preemption (spot-like pricing)

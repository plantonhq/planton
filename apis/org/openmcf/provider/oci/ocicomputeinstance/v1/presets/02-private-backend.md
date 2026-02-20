# Private Backend Instance

This preset creates a production-hardened OCI compute instance in a private subnet with no public IP. It enables in-transit encryption, disables legacy IMDS endpoints, configures live migration for zero-downtime maintenance, and associates the instance with a Network Security Group for fine-grained traffic control. This is the standard pattern for application servers, worker processes, and any backend service that sits behind a load balancer or is accessed only from within the VCN.

## When to Use

- Application servers and microservices behind an OCI Load Balancer or Network Load Balancer
- Worker processes, queue consumers, and batch jobs that need no inbound internet access
- Database clients and internal services that communicate only within the VCN
- Any production workload where direct internet exposure is unacceptable

## Key Configuration Choices

- **Private subnet with no public IP** (`createVnicDetails.assignPublicIp: false`) -- The instance is not directly reachable from the internet. Outbound internet access is available via the VCN's NAT Gateway. SSH access requires a Bastion service, VPN, or jump host.
- **NSG association** (`createVnicDetails.nsgIds`) -- Associates the VNIC with a Network Security Group for stateful, fine-grained ingress/egress rules. Use the `OciSecurityGroup` component's `02-private-backend` preset as a companion.
- **Private DNS record** (`createVnicDetails.assignPrivateDnsRecord: true`) -- Registers a DNS hostname within the VCN's private DNS, enabling service discovery like `mybackend.<subnet-dns-label>.<vcn-dns-label>.oraclevcn.com` without external DNS configuration.
- **2 OCPUs / 32 GiB memory** (`shapeConfig.ocpus: 2`, `shapeConfig.memoryInGbs: 32`) -- A production-appropriate starting point. Scale up as needed; E4 Flex supports up to 64 OCPUs with 1-64 GiB per OCPU.
- **100 GiB boot volume at Higher Performance** (`bootVolumeSizeInGbs: 100`, `bootVolumeVpusPerGb: 20`) -- More headroom for application binaries, logs, and temporary data. 20 VPUs/GB provides the Higher Performance tier with better IOPS and throughput than the default Balanced tier.
- **In-transit encryption enabled** (`isPvEncryptionInTransitEnabled: true`) -- Encrypts data between the instance and its paravirtualized boot/data volume attachments. No performance impact on modern shapes. This is an OCI security best practice for production workloads.
- **Legacy IMDS disabled** (`instanceOptions.areLegacyImdsEndpointsDisabled: true`) -- Forces applications to use the IMDSv2 token-based endpoint, preventing SSRF attacks that exploit the unauthenticated IMDSv1 endpoint. Equivalent to AWS's `HttpTokens: required` on EC2.
- **Live migration preferred** (`availabilityConfig.isLiveMigrationPreferred: true`) -- During planned infrastructure maintenance, OCI will live-migrate the instance instead of rebooting it, avoiding downtime. Not all shapes support live migration; OCI falls back to reboot when it is unavailable.
- **Automatic recovery** (`availabilityConfig.recoveryAction: restore_instance`) -- If the underlying host fails unexpectedly, OCI automatically restores the instance on healthy hardware. The alternative `stop_instance` leaves it stopped, requiring manual intervention.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the instance will be created | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |
| `<availability-domain>` | Availability domain name (e.g., `Ixxj:US-ASHBURN-AD-1`) | OCI Console > Compute > Instances > Create Instance, or `oci iam availability-domain list` |
| `<image-ocid>` | OCID of the OS image to boot from | OCI Console > Compute > Custom Images, or `oci compute image list --compartment-id <tenancy-ocid>` |
| `<private-subnet-ocid>` | OCID of a private subnet for the primary VNIC | OCI Console > Networking > VCNs > Subnets, or `OciSubnet` status outputs |
| `<nsg-ocid>` | OCID of a Network Security Group to associate with the VNIC | OCI Console > Networking > VCNs > Network Security Groups, or `OciSecurityGroup` status outputs |
| `<ssh-public-key>` | SSH public key content (e.g., `ssh-rsa AAAA...`) | Your local `~/.ssh/id_rsa.pub` or equivalent |

## Related Presets

- **01-general-purpose-flex** -- Use instead when the instance needs a public IP and you want a simpler, minimal configuration
- **03-preemptible-dev** -- Use instead for cost-optimized dev/test instances that tolerate preemption

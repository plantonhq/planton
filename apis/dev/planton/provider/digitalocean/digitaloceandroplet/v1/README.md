# DigitalOcean Droplet

## Overview

The **DigitalOcean Droplet API Resource** provides a declarative, protobuf-defined interface for provisioning and managing DigitalOcean Droplets (virtual machines) as infrastructure-as-code. This resource abstracts the complexity of Droplet lifecycle management while maintaining full control over compute resources, networking, and security configuration.

## Purpose

We developed this API resource to bring Infrastructure-as-Code best practices to Digital

Ocean Droplet management. By offering a unified, type-safe interface, it enables teams to:

- **Deploy VMs as Code**: Define Droplet configurations in version-controlled YAML manifests
- **Ensure Security by Default**: Enforce SSH key authentication and VPC networking through validation rules
- **Automate Provisioning**: Use cloud-init user data for first-boot configuration and package installation
- **Maintain Consistency**: Guarantee identical configurations across development, staging, and production environments
- **Enable Team Collaboration**: Review infrastructure changes through pull requests before deployment

## Key Features

- **Type-Safe Configuration**: Protobuf-based spec with built-in validation (hostname patterns, size slugs, image format)
- **Security-First Design**: Requires SSH keys and VPC assignment; prevents insecure password authentication
- **Cloud-Init Integration**: First-class support for cloud-config and shell scripts via `user_data` field
- **Production-Ready Defaults**: Monitoring enabled by default; backups and IPv6 available as opt-in features
- **Volume Attachments**: Reference DigitalOcean Volumes for persistent data independent of Droplet lifecycle
- **Dual IaC Support**: Deploy via Pulumi (Go) or Terraform with identical spec definitions

## Use Cases

### Development Environments

Create lightweight, disposable development servers with minimal configuration:

- 1 vCPU, 1 GB RAM Droplets for cost-effective dev environments
- Automated package installation via cloud-init
- Quick provisioning (<2 minutes from manifest to SSH-ready)

### Staging and Production Workloads

Deploy production-grade virtual machines with full observability and backup protection:

- General Purpose or CPU-Optimized Droplets for compute-intensive workloads
- Automated backups (20-30% of Droplet cost) for disaster recovery
- DigitalOcean monitoring agent for CPU, memory, and disk metrics
- VPC networking for private communication between Droplets

### Database and Stateful Services

Run self-hosted databases (PostgreSQL, MySQL, Redis) with persistent storage:

- Attach Block Storage Volumes (independent lifecycle from Droplet)
- Enable automated backups for point-in-time recovery
- Provision in VPC with Cloud Firewall rules for network isolation

### Custom Application Servers

Deploy applications that require OS-level control not available in PaaS offerings:

- Install custom system packages (compiler toolchains, native libraries)
- Configure kernel parameters or systemd services
- Run legacy software that doesn't containerize easily

## Important Considerations

### Security Best Practices

**SSH Keys (Required)**:
- The spec **requires** at least one SSH key fingerprint or ID
- DigitalOcean recommends disabling password authentication entirely
- Use SSH keys managed separately (DigitalOcean control panel or API)

**VPC Networking (Required)**:
- All Droplets must be assigned to a VPC (Virtual Private Cloud)
- VPCs provide private networking isolated from the public internet
- Use Cloud Firewalls in conjunction with VPCs for defense-in-depth

**Monitoring (Enabled by Default)**:
- DigitalOcean's monitoring agent is free and recommended for production
- Provides CPU, memory, disk, and bandwidth metrics
- Can be explicitly disabled via `disable_monitoring: true` if needed

### Resource Lifecycle

**Droplet Sizing**:
- Size can be increased after creation (more vCPU, RAM, disk)
- **Limitation**: Disk size cannot be decreased (DigitalOcean platform constraint)
- Plan size changes carefully; test with snapshots before resizing production

**Volumes**:
- Block Storage Volumes persist independently of Droplets
- **Limitation**: A Volume can only attach to one Droplet at a time (no shared storage)
- For shared storage, use NFS server or object storage (Spaces)

**Backups vs. Snapshots**:
- **Backups**: Automated (daily/weekly), 7-day or 4-week retention, percentage-based pricing
- **Snapshots**: Manual, kept indefinitely, size-based pricing ($0.06/GB-month)
- Best practice: Enable backups for production, use snapshots for golden images

### Platform Limitations

**No Auto-Scaling**:
- Droplets are individual VMs, not part of an auto-scaling group
- For auto-scaling, consider DigitalOcean Kubernetes (DOKS) or App Platform

**No Load Balancer Integration in Spec**:
- Droplets can be manually added to DigitalOcean Load Balancers via tags
- Load Balancer configuration is handled separately (not in this resource)

**Region-Specific Constraints**:
- Volumes must be in the same region as the Droplet
- IPv6 availability varies by region (most regions support it)

## Getting Started

### Basic Usage

See the `examples.md` file in this directory for practical configuration examples, including:

- Minimal development Droplet
- Standard staging configuration with monitoring
- Production setup with backups and volumes
- Cloud-init integration patterns

### Prerequisites

Before creating a Droplet, ensure you have:

1. **DigitalOcean API Token**: Generate from https://cloud.digitalocean.com/account/api/tokens
2. **SSH Key**: Upload to DigitalOcean control panel or create via API
3. **VPC**: Create a VPC in the desired region (required for Droplet)

### Deployment Workflow

1. **Define Manifest**: Create a YAML file with Droplet specification
2. **Validate**: Protobuf validation ensures required fields and correct patterns
3. **Deploy**: Use Planton CLI, Pulumi, or Terraform to provision
4. **Configure**: Cloud-init user data runs on first boot
5. **Verify**: SSH into Droplet to confirm configuration

## Architecture

### Networking Model

```
                     Public Internet
                           |
                   ┌───────┴───────┐
                   │  Load Balancer │ (optional)
                   └───────┬───────┘
                           |
      ┌────────────────────┼────────────────────┐
      │                    VPC                   │
      │  ┌─────────────────────────────────┐    │
      │  │         Droplet                 │    │
      │  │  - Public IPv4 (internet access)│    │
      │  │  - Private IPv4 (VPC only)      │    │
      │  │  - Optional IPv6                │    │
      │  └──────────┬──────────────────────┘    │
      │             │                            │
      │    ┌────────┴────────┐                  │
      │    │  Block Storage  │ (attached volume)│
      │    │   (persistent)  │                  │
      │    └─────────────────┘                  │
      └──────────────────────────────────────────┘
```

### Cloud-Init Execution Flow

```
1. Droplet created via API
2. OS boots (Ubuntu, Debian, Fedora, etc.)
3. cloud-init reads user_data from metadata service
4. Executes configuration:
   - package_update / package_upgrade
   - Install packages
   - Write files
   - Run commands
5. Droplet ready for use (SSH access via keys)
```

## Documentation

For detailed implementation guidance, refer to:

- **Research Document**: `docs/README.md` - Comprehensive analysis of Droplet deployment methods, 80/20 configuration decisions, and production best practices
- **Examples**: `examples.md` - Practical configuration patterns for common use cases
- **Pulumi Implementation**: `iac/pulumi/README.md` - Pulumi-specific deployment instructions
- **Terraform Implementation**: `iac/tf/README.md` - Terraform module usage guide

## Comparison to Alternatives

### DigitalOcean App Platform (PaaS)

**Choose Droplets when**:
- You need OS-level control (custom packages, kernel parameters)
- Running stateful services (self-hosted databases)
- Migrating from traditional hosting

**Choose App Platform when**:
- Running stateless web apps or APIs
- You want zero infrastructure management
- Rapid iteration is more important than control

### DigitalOcean Kubernetes (DOKS)

**Choose Droplets when**:
- Running a few services that don't justify Kubernetes complexity
- Team lacks Kubernetes expertise
- Simpler operational model is preferred

**Choose Kubernetes when**:
- Running containerized microservices at scale
- Need auto-scaling and self-healing capabilities
- Team has Kubernetes expertise

### AWS EC2 / GCP Compute Engine

**Choose DigitalOcean Droplets when**:
- Simplified pricing (flat monthly rate vs. complex per-second billing)
- Smaller infrastructure footprint (<100 VMs)
- DigitalOcean ecosystem integration (Spaces, Load Balancers, etc.)

**Choose AWS/GCP when**:
- Need advanced features (spot instances, custom AMIs, placement groups)
- Deep integration with AWS/GCP services
- Enterprise support requirements

## Production Best Practices

1. **Always Use VPCs**: Isolate Droplets from public internet by default
2. **Enable Monitoring**: Free and provides essential observability
3. **Automate with Cloud-Init**: Don't SSH and configure manually; use user_data
4. **Use Volumes for Data**: Separate data storage from Droplet lifecycle
5. **Enable Backups for Production**: 20-30% cost is worth disaster recovery capability
6. **Version Control Manifests**: Store configurations in Git, not just deployed state
7. **Test Snapshots**: Create snapshots before major changes (OS upgrades, config changes)

## Troubleshooting

### Common Issues

**Droplet Creation Fails**:
- Verify SSH key exists in DigitalOcean account
- Confirm VPC exists in the specified region
- Check size slug is valid for the region (some sizes are region-specific)

**Cannot SSH to Droplet**:
- Verify SSH key fingerprint matches the one in your DigitalOcean account
- Check Cloud Firewall rules allow SSH (port 22) from your IP
- Confirm Droplet has a public IPv4 address

**Cloud-Init Doesn't Run**:
- Verify user_data is valid cloud-config YAML or shell script
- SSH to Droplet and check `/var/log/cloud-init-output.log` for errors
- Ensure user_data is under 32 KiB size limit

**Volume Attachment Fails**:
- Confirm volume is in the same region as Droplet
- Verify volume is not already attached to another Droplet
- Check volume ID is correct

## Security Considerations

- **Never use password authentication**: SSH keys only
- **Restrict SSH access**: Use Cloud Firewalls to limit SSH to office IPs
- **Keep OS updated**: Enable automated security updates in cloud-init
- **Use VPCs**: Isolate Droplets from public internet where possible
- **Enable fail2ban**: Automatically ban IPs with failed SSH attempts
- **Rotate SSH keys**: Periodically update authorized_keys

## Next Steps

1. Review `examples.md` for configuration patterns
2. Read `docs/README.md` for in-depth platform analysis
3. Choose your deployment method:
   - **Pulumi**: See `iac/pulumi/README.md`
   - **Terraform**: See `iac/tf/README.md`
4. Create a test Droplet in a non-production VPC
5. Verify SSH access and cloud-init execution
6. Configure monitoring and alerting

## License

This implementation is part of the Planton monorepo and follows the same license.

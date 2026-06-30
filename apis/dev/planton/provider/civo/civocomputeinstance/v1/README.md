# CivoComputeInstance API

## Overview

The `CivoComputeInstance` API provides a declarative way to provision and manage virtual machines on Civo Cloud. Civo offers fast-booting instances (< 60 seconds), unlimited bandwidth, and simplified pricing—making it ideal for dev environments, production workloads, and cost-optimized infrastructure.

This API abstracts the complexity of instance provisioning, networking, security, and storage configuration into a simple Protobuf-based specification.

## API Structure

```protobuf
message CivoComputeInstance {
  string api_version = 1;                           // "civo.planton.dev/v1"
  string kind = 2;                                  // "CivoComputeInstance"
  CloudResourceMetadata metadata = 3;               // Name, labels, description
  CivoComputeInstanceSpec spec = 4;                 // Instance configuration
  CivoComputeInstanceStatus status = 5;             // Runtime outputs
}
```

## Specification Fields

### `CivoComputeInstanceSpec`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `instance_name` | `string` | Yes | Instance hostname (lowercase, alphanumeric + dashes/dots, ≤63 chars) |
| `region` | `CivoRegion` | Yes | Civo region (LON1, NYC1, FRA1, PHX1, MUM1) |
| `size` | `string` | Yes | Instance size slug (g3.small, g3.medium, g3.large, etc.) |
| `image` | `string` | Yes | OS image slug (ubuntu-jammy, debian-11, rocky-9, etc.) |
| `network` | `StringValueOrRef` | Yes | Network ID or reference to CivoVpc |
| `ssh_key_ids` | `repeated string` | No | SSH public key IDs for passwordless login |
| `firewall_ids` | `repeated StringValueOrRef` | No | Firewall IDs or references to CivoFirewall |
| `volume_ids` | `repeated StringValueOrRef` | No | Volume IDs or references to CivoVolume |
| `reserved_ip_id` | `StringValueOrRef` | No | Reserved IP ID or reference to CivoIpAddress |
| `tags` | `repeated string` | No | Organizational tags (must be unique) |
| `user_data` | `string` | No | Cloud-init script (≤32KB) |

#### Instance Name Validation

- **Length**: ≤63 characters
- **Characters**: Lowercase letters, numbers, dashes, dots
- **Start/End**: Must start and end with alphanumeric (no trailing dash/dot)
- **Pattern**: `^[a-z0-9]([a-z0-9\.\-]*[a-z0-9])?$`

**Valid**: `web-server`, `db01.internal`, `app-prod-01`  
**Invalid**: `Web-Server`, `test_vm`, `server-`, `.hidden`

#### Instance Sizes

Common Civo instance sizes:

| Size | vCPUs | RAM | Use Case |
|------|-------|-----|----------|
| `g3.small` | 1 | 2GB | Dev, testing, lightweight apps |
| `g3.medium` | 2 | 4GB | Small production apps, databases |
| `g3.large` | 4 | 8GB | Production web servers, API backends |
| `g3.xlarge` | 6 | 16GB | Large databases, data processing |
| `g3.2xlarge` | 8 | 32GB | Heavy workloads, in-memory caching |

**Cost**: ~$10-120/month depending on size. Check current pricing at https://www.civo.com/pricing

#### OS Images

Popular images:
- `ubuntu-jammy` - Ubuntu 22.04 LTS
- `ubuntu-focal` - Ubuntu 20.04 LTS
- `debian-11` - Debian 11
- `rocky-9` - Rocky Linux 9
- `centos-stream-9` - CentOS Stream 9

List available images: `civo diskimage list`

#### Networking

The `network` field references a Civo network (VPC). Instances in the same network can communicate privately.

**StringValueOrRef**:
- `value`: Direct network ID string
- `ref`: Reference to a `CivoVpc` resource (resolved at runtime)

#### SSH Keys

Provide SSH public key IDs for passwordless authentication. Without SSH keys, Civo sets a random password (emailed to account owner).

**Best Practice**: Always use SSH keys, disable password auth in cloud-init.

#### Firewalls

Attach firewall rules to control traffic. **Important**: Civo's default firewall allows all traffic—always create custom firewalls for production.

**Best Practice**: One firewall per role (web, database, bastion).

#### Volumes

Attach persistent block storage volumes to instances. Volumes persist beyond instance lifecycle.

**Best Practice**: Use volumes for databases and stateful applications.

#### Reserved IPs

Static public IP addresses that persist beyond instance lifecycle. Ideal for production endpoints.

**Best Practice**: Use reserved IPs for production services that need stable DNS.

#### User Data (Cloud-Init)

Shell script executed on first boot. Use for:
- Installing packages
- Configuring services
- Deploying applications
- Setting up monitoring

**Limit**: 32KB (enforced by validation)

## Status and Outputs

### `CivoComputeInstanceStackOutputs`

| Field | Type | Description |
|-------|------|-------------|
| `instance_id` | `string` | Unique identifier (UUID) |
| `public_ipv4` | `string` | Public IP address |
| `private_ipv4` | `string` | Private IP address |
| `status` | `string` | Instance state (ACTIVE, BUILDING, etc.) |
| `created_at` | `string` | Creation timestamp |

## Quick Start

### Minimal Example

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoComputeInstance
metadata:
  name: dev-vm
spec:
  instanceName: dev-vm
  region: NYC1
  size: g3.small
  image: ubuntu-jammy
  network:
    value: default-network-id
```

### Production Example

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoComputeInstance
metadata:
  name: web-server-prod
  description: Production web server
spec:
  instanceName: web-server-prod
  region: LON1
  size: g3.large
  image: ubuntu-jammy
  network:
    value: prod-network-id
  sshKeyIds:
    - team-ssh-key-id
  firewallIds:
    - value: web-firewall-id
  reservedIpId:
    value: prod-static-ip-id
  tags:
    - env:prod
    - service:web
  userData: |
    #!/bin/bash
    apt-get update
    apt-get upgrade -y
    apt-get install -y docker.io nginx
```

## Security Best Practices

1. **Always use custom firewalls** - Never rely on Civo's permissive default
2. **SSH keys only** - Disable password authentication
3. **Minimal attack surface** - Only expose necessary ports
4. **Regular updates** - Include `apt-get upgrade` in user_data
5. **Private networks** - Use isolated networks per environment
6. **Reserved IPs** - Use for production (enables zero-downtime instance replacement)

## Cost Optimization

- **Right-size instances**: Start small (g3.small), scale up as needed
- **Use reserved IPs sparingly**: Only for production endpoints
- **Leverage free bandwidth**: No egress fees (unlike AWS)
- **Shut down dev/staging** instances when not in use
- **Use snapshots**: For quick recovery vs continuous backups

## Common Use Cases

### 1. Web Application Server

```yaml
spec:
  instanceName: web-app
  size: g3.medium
  image: ubuntu-jammy
  firewallIds:
    - value: allow-http-https
  sshKeyIds: [team-key]
  userData: |
    #!/bin/bash
    apt-get update && apt-get install -y nginx
```

### 2. Database Server

```yaml
spec:
  instanceName: postgres-db
  size: g3.xlarge
  image: ubuntu-jammy
  volumeIds:
    - value: postgres-data-volume
  firewallIds:
    - value: allow-postgres-only
  reservedIpId:
    value: db-static-ip
```

### 3. Bastion Host

```yaml
spec:
  instanceName: bastion
  size: g3.small
  image: ubuntu-jammy
  firewallIds:
    - value: ssh-from-office-only
  reservedIpId:
    value: bastion-ip
```

## Troubleshooting

### Issue: Instance won't boot

**Cause**: Invalid image or network not in specified region  
**Solution**: Verify image exists in region: `civo diskimage list --region NYC1`

### Issue: Can't SSH to instance

**Cause**: Firewall blocking port 22 or wrong SSH key  
**Solution**: Check firewall rules, verify SSH key was added

### Issue: Instance has no public IP

**Cause**: Network configuration or no public IP requested  
**Solution**: Use reserved_ip_id or check network settings

## Related Documentation

- **Examples**: See [examples.md](./examples.md) for real-world scenarios
- **Research**: See [docs/README.md](./docs/README.md) for deployment methods deep dive
- **IaC Implementation**: See [iac/pulumi/](./iac/pulumi/) for Pulumi module details
- **Civo Docs**: [Civo Compute Documentation](https://www.civo.com/docs/compute)

## Version History

- **v1**: Initial release with compute instance provisioning, networking, security, and storage support

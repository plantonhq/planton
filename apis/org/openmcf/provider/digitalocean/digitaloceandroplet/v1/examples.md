# DigitalOcean Droplet Examples

This document provides practical examples of common Droplet configurations using the DigitalOcean Droplet API resource.

## Table of Contents

1. [Minimal Development Droplet](#minimal-development-droplet)
2. [Standard Staging Droplet](#standard-staging-droplet)
3. [Production Droplet with Backups](#production-droplet-with-backups)
4. [Droplet with Block Storage Volume](#droplet-with-block-storage-volume)
5. [Droplet with Cloud-Init Script](#droplet-with-cloud-init-script)
6. [Droplet with Cloud-Config YAML](#droplet-with-cloud-config-yaml)
7. [High-CPU Droplet for Compute Workloads](#high-cpu-droplet-for-compute-workloads)
8. [Droplet with IPv6 Enabled](#droplet-with-ipv6-enabled)
9. [Database Server with Volume](#database-server-with-volume)
10. [Complete Production Setup](#complete-production-setup)

---

## Minimal Development Droplet

Basic, cost-effective configuration for development and testing.

```yaml
apiVersion: code2cloud.planton.ai/v1
kind: DigitalOceanDroplet
metadata:
  name: dev-server
spec:
  digitalOceanCredentialId: do-dev-credentials
  dropletName: dev-server
  region: digital_ocean_region_nyc3
  size: s-1vcpu-1gb  # $6/month - smallest size
  image: ubuntu-22-04-x64
  sshKeys:
    - "your-ssh-key-fingerprint"
  vpc:
    value: "vpc-uuid-for-dev-environment"
  tags:
    - development
    - temporary
```

**Use Case**: Quick spin-up for testing, learning, or prototyping. Minimal cost, basic specs.

**Expected Behavior**:
- Creates 1 vCPU, 1 GB RAM Droplet in NYC3
- Monitoring enabled by default
- No backups (use snapshots manually if needed)
- Destroyable without data loss concerns

---

## Standard Staging Droplet

Balanced configuration for staging environments that mirror production settings.

```yaml
apiVersion: code2cloud.planton.ai/v1
kind: DigitalOceanDroplet
metadata:
  name: staging-app
spec:
  digitalOceanCredentialId: do-staging-credentials
  dropletName: staging-app-01
  region: digital_ocean_region_sfo3
  size: g-2vcpu-8gb  # General Purpose 2 vCPU, 8 GB RAM
  image: ubuntu-22-04-x64
  sshKeys:
    - "cicd-key-fingerprint"
    - "team-lead-key-fingerprint"
  vpc:
    value: "vpc-uuid-for-staging"
  enableBackups: true  # Enable automated backups
  tags:
    - staging
    - webapp
    - staging-firewall  # Used by Cloud Firewall for auto-assignment
  userData: |
    #!/bin/bash
    apt-get update
    apt-get install -y nginx
    systemctl enable nginx
    systemctl start nginx
```

**Use Case**: Pre-production environment with monitoring, backups, and basic package installation.

**Expected Behavior**:
- 2 vCPU, 8 GB RAM for realistic performance testing
- Automated backups (20% cost increase, ~$24/month total)
- Nginx installed automatically on first boot
- Tagged for Cloud Firewall rule application

---

## Production Droplet with Backups

Production-grade configuration with all recommended features enabled.

```yaml
apiVersion: code2cloud.planton.ai/v1
kind: DigitalOceanDroplet
metadata:
  name: prod-web-01
spec:
  digitalOceanCredentialId: do-prod-credentials
  dropletName: prod-web-01
  region: digital_ocean_region_nyc3
  size: g-4vcpu-16gb  # General Purpose 4 vCPU, 16 GB RAM
  image: ubuntu-22-04-x64
  sshKeys:
    - "prod-deploy-key-fingerprint"
    - "ops-team-key-fingerprint"
  vpc:
    valueFromResourceOutput:
      resourceIdRef:
        name: prod-vpc
      outputKey: vpc_id
  enableBackups: true
  disableMonitoring: false  # Explicitly keep monitoring enabled
  tags:
    - production
    - web-tier
    - prod-firewall
    - load-balancer-pool
```

**Use Case**: Production web server with backup protection and monitoring.

**Expected Behavior**:
- Robust specs for production workload (4 vCPU, 16 GB RAM)
- Automated daily backups with 7-day retention
- DigitalOcean monitoring agent for metrics
- VPC reference from another resource (cross-resource dependency)

---

## Droplet with Block Storage Volume

Droplet with persistent storage volume for data that outlives the VM.

```yaml
apiVersion: code2cloud.planton.ai/v1
kind: DigitalOceanDroplet
metadata:
  name: app-with-storage
spec:
  digitalOceanCredentialId: do-prod-credentials
  dropletName: app-with-storage
  region: digital_ocean_region_nyc3
  size: g-2vcpu-8gb
  image: ubuntu-22-04-x64
  sshKeys:
    - "deploy-key-fingerprint"
  vpc:
    value: "vpc-uuid-prod"
  volumeIds:
    - valueFromResourceOutput:
        resourceIdRef:
          name: app-data-volume
        outputKey: volume_id
  enableBackups: true
  tags:
    - production
    - data-tier
```

**Use Case**: Application server with separate data volume for database files or uploads.

**Expected Behavior**:
- Volume automatically attached at `/mnt/your-volume-name`
- Data persists even if Droplet is destroyed
- Volume must be in same region as Droplet

**Note**: Create the DigitalOceanVolume resource separately before referencing it.

---

## Droplet with Cloud-Init Script

Using a shell script for first-boot provisioning.

```yaml
apiVersion: code2cloud.planton.ai/v1
kind: DigitalOceanDroplet
metadata:
  name: scripted-droplet
spec:
  digitalOceanCredentialId: do-dev-credentials
  dropletName: scripted-droplet
  region: digital_ocean_region_nyc3
  size: s-2vcpu-4gb
  image: ubuntu-22-04-x64
  sshKeys:
    - "dev-key-fingerprint"
  vpc:
    value: "vpc-uuid-dev"
  userData: |
    #!/bin/bash
    set -e

    # Update package lists
    apt-get update

    # Install packages
    apt-get install -y \
      nginx \
      postgresql-client \
      fail2ban \
      ufw

    # Configure firewall
    ufw allow 22/tcp
    ufw allow 80/tcp
    ufw allow 443/tcp
    ufw --force enable

    # Start services
    systemctl enable nginx
    systemctl start nginx
    systemctl enable fail2ban
    systemctl start fail2ban

    # Write a test file
    echo "Provisioned at $(date)" > /var/www/html/index.html
```

**Use Case**: Quick package installation and service configuration.

**Expected Behavior**:
- Nginx, PostgreSQL client, fail2ban, and ufw installed
- Firewall configured to allow SSH, HTTP, HTTPS
- Services enabled and started automatically

**Debugging**: SSH to Droplet and check `/var/log/cloud-init-output.log` for script execution logs.

---

## Droplet with Cloud-Config YAML

Using declarative cloud-config for better readability and maintainability.

```yaml
apiVersion: code2cloud.planton.ai/v1
kind: DigitalOceanDroplet
metadata:
  name: cloud-config-droplet
spec:
  digitalOceanCredentialId: do-staging-credentials
  dropletName: cloud-config-droplet
  region: digital_ocean_region_sfo3
  size: g-2vcpu-8gb
  image: ubuntu-22-04-x64
  sshKeys:
    - "staging-key-fingerprint"
  vpc:
    value: "vpc-uuid-staging"
  userData: |
    #cloud-config
    package_update: true
    package_upgrade: true
    packages:
      - nginx
      - postgresql-14
      - redis-server
      - fail2ban
      - ufw

    write_files:
      - path: /etc/nginx/sites-available/default
        content: |
          server {
            listen 80 default_server;
            listen [::]:80 default_server;
            root /var/www/html;
            index index.html;
            server_name _;
            location / {
              try_files $uri $uri/ =404;
            }
          }

      - path: /etc/sysctl.d/99-custom.conf
        content: |
          # Increase TCP buffer sizes for high-throughput
          net.core.rmem_max = 134217728
          net.core.wmem_max = 134217728

    runcmd:
      - systemctl enable nginx
      - systemctl start nginx
      - systemctl enable postgresql
      - systemctl start postgresql
      - ufw allow 22/tcp
      - ufw allow 80/tcp
      - ufw allow 443/tcp
      - ufw --force enable
      - sysctl --system
```

**Use Case**: Production-style provisioning with package installation, file creation, and service management.

**Expected Behavior**:
- System packages updated before installation
- Nginx, PostgreSQL, Redis installed and configured
- Custom nginx config written
- Kernel parameters tuned for high throughput
- Firewall enabled with appropriate rules

---

## High-CPU Droplet for Compute Workloads

CPU-optimized Droplet for computational tasks (builds, video encoding, ML inference).

```yaml
apiVersion: code2cloud.planton.ai/v1
kind: DigitalOceanDroplet
metadata:
  name: compute-worker
spec:
  digitalOceanCredentialId: do-prod-credentials
  dropletName: compute-worker-01
  region: digital_ocean_region_nyc3
  size: c-8vcpu-16gb  # CPU-Optimized: 8 dedicated vCPUs, 16 GB RAM
  image: ubuntu-22-04-x64
  sshKeys:
    - "cicd-key-fingerprint"
  vpc:
    value: "vpc-uuid-prod"
  enableBackups: false  # Stateless worker, no backup needed
  tags:
    - compute-worker
    - batch-processing
```

**Use Case**: Compute-intensive workloads where CPU performance is critical.

**Expected Behavior**:
- Dedicated CPU cores (no noisy neighbors)
- Higher CPU-to-RAM ratio compared to General Purpose
- Ideal for CI/CD agents, video encoding, scientific computing

**Cost**: CPU-Optimized Droplets cost more per vCPU but provide better single-thread performance.

---

## Droplet with IPv6 Enabled

Droplet with both IPv4 and IPv6 connectivity.

```yaml
apiVersion: code2cloud.planton.ai/v1
kind: DigitalOceanDroplet
metadata:
  name: ipv6-droplet
spec:
  digitalOceanCredentialId: do-dev-credentials
  dropletName: ipv6-droplet
  region: digital_ocean_region_nyc3
  size: s-2vcpu-4gb
  image: ubuntu-22-04-x64
  sshKeys:
    - "dev-key-fingerprint"
  vpc:
    value: "vpc-uuid-dev"
  enableIpv6: true  # Enable IPv6 networking
  tags:
    - ipv6-enabled
    - development
```

**Use Case**: Services that need IPv6 connectivity (modern web apps, CDN origins).

**Expected Behavior**:
- Droplet assigned both IPv4 and IPv6 addresses
- IPv6 address routable from the internet
- Useful for testing IPv6 compatibility

**Note**: Not all regions support IPv6. Check DigitalOcean's documentation for availability.

---

## Database Server with Volume

Self-hosted PostgreSQL database with persistent storage and backups.

```yaml
apiVersion: code2cloud.planton.ai/v1
kind: DigitalOceanDroplet
metadata:
  name: postgres-db-server
spec:
  digitalOceanCredentialId: do-prod-credentials
  dropletName: postgres-db-01
  region: digital_ocean_region_nyc3
  size: g-4vcpu-16gb  # Adequate for small-to-medium database
  image: ubuntu-22-04-x64
  sshKeys:
    - "dba-team-key-fingerprint"
    - "backup-automation-key-fingerprint"
  vpc:
    value: "vpc-uuid-prod"
  volumeIds:
    - value: "volume-uuid-for-postgres-data"  # 100 GB Volume for /var/lib/postgresql
  enableBackups: true  # Automated backups (in addition to Volume snapshots)
  tags:
    - production
    - database
    - postgres
    - db-firewall  # Restrict access to app servers only
  userData: |
    #cloud-config
    package_update: true
    packages:
      - postgresql-14
      - postgresql-contrib-14

    write_files:
      - path: /etc/systemd/system/mount-postgres-data.service
        content: |
          [Unit]
          Description=Mount DigitalOcean Volume for PostgreSQL
          Before=postgresql.service

          [Service]
          Type=oneshot
          ExecStart=/bin/bash -c 'if ! mountpoint -q /var/lib/postgresql; then mount /dev/disk/by-id/scsi-0DO_Volume_postgres-data /var/lib/postgresql; fi'

          [Install]
          WantedBy=multi-user.target

    runcmd:
      - systemctl daemon-reload
      - systemctl enable mount-postgres-data.service
      - systemctl start mount-postgres-data.service
      - chown -R postgres:postgres /var/lib/postgresql
      - systemctl enable postgresql
      - systemctl start postgresql
```

**Use Case**: Production PostgreSQL database with data stored on persistent volume.

**Expected Behavior**:
- Volume mounted at `/var/lib/postgresql` for database files
- PostgreSQL installed and configured to use volume
- Data survives Droplet recreation (attach volume to new Droplet)

**Best Practice**: Also take Volume snapshots regularly for point-in-time recovery.

---

## Complete Production Setup

Comprehensive configuration with all production features.

```yaml
apiVersion: code2cloud.planton.ai/v1
kind: DigitalOceanDroplet
metadata:
  name: prod-app-complete
spec:
  digitalOceanCredentialId: do-prod-credentials
  dropletName: prod-app-01
  region: digital_ocean_region_nyc3
  size: g-8vcpu-32gb  # General Purpose 8 vCPU, 32 GB RAM
  image: ubuntu-22-04-x64
  sshKeys:
    - "ops-team-key-fingerprint"
    - "emergency-access-key-fingerprint"
    - "deploy-automation-key-fingerprint"
  vpc:
    valueFromResourceOutput:
      resourceIdRef:
        name: prod-vpc
      outputKey: vpc_id
  volumeIds:
    - valueFromResourceOutput:
        resourceIdRef:
          name: app-uploads-volume
        outputKey: volume_id
    - valueFromResourceOutput:
        resourceIdRef:
          name: app-logs-volume
        outputKey: volume_id
  enableBackups: true
  enableIpv6: true
  disableMonitoring: false
  tags:
    - production
    - app-tier
    - prod-firewall
    - load-balancer-backend
  timezone: digital_ocean_droplet_timezone_utc
  userData: |
    #cloud-config
    package_update: true
    package_upgrade: true
    packages:
      - nginx
      - nodejs
      - npm
      - postgresql-client
      - redis-tools
      - fail2ban
      - ufw
      - unattended-upgrades

    write_files:
      - path: /etc/apt/apt.conf.d/50unattended-upgrades
        content: |
          Unattended-Upgrade::Allowed-Origins {
              "${distro_id}:${distro_codename}";
              "${distro_id}:${distro_codename}-security";
          };
          Unattended-Upgrade::AutoFixInterruptedDpkg "true";
          Unattended-Upgrade::Remove-Unused-Kernel-Packages "true";
          Unattended-Upgrade::Remove-Unused-Dependencies "true";
          Unattended-Upgrade::Automatic-Reboot "true";
          Unattended-Upgrade::Automatic-Reboot-Time "03:00";

      - path: /etc/nginx/sites-available/app
        content: |
          upstream app_backend {
            server 127.0.0.1:3000;
          }

          server {
            listen 80;
            listen [::]:80;
            server_name _;

            location / {
              proxy_pass http://app_backend;
              proxy_set_header Host $host;
              proxy_set_header X-Real-IP $remote_addr;
              proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
              proxy_set_header X-Forwarded-Proto $scheme;
            }
          }

      - path: /etc/systemd/system/app.service
        content: |
          [Unit]
          Description=Production Node.js App
          After=network.target

          [Service]
          Type=simple
          User=www-data
          WorkingDirectory=/opt/app
          ExecStart=/usr/bin/node /opt/app/server.js
          Restart=always
          RestartSec=10
          StandardOutput=syslog
          StandardError=syslog
          SyslogIdentifier=app

          [Install]
          WantedBy=multi-user.target

    runcmd:
      - ln -s /etc/nginx/sites-available/app /etc/nginx/sites-enabled/
      - rm -f /etc/nginx/sites-enabled/default
      - systemctl enable nginx
      - systemctl restart nginx
      - systemctl enable fail2ban
      - systemctl start fail2ban
      - ufw allow 22/tcp
      - ufw allow 80/tcp
      - ufw allow 443/tcp
      - ufw --force enable
      - mkdir -p /opt/app
      - systemctl enable app.service
```

**Use Case**: Enterprise-grade production deployment with security hardening, auto-updates, and monitoring.

**Expected Behavior**:
- All security features enabled (fail2ban, ufw, auto-updates)
- Nginx reverse proxy to Node.js application
- systemd service for application lifecycle management
- Multiple SSH keys for team access
- Two attached volumes (uploads and logs)
- Automated security updates with scheduled reboots (3 AM)

---

## Tips and Best Practices

### SSH Key Management

- **Never hardcode fingerprints**: Use DigitalOcean API to fetch key IDs dynamically
- **Rotate keys regularly**: Update `authorized_keys` via cloud-init or configuration management
- **Use separate keys**: Different keys for CI/CD, ops team, emergency access

### VPC Best Practices

- **One VPC per environment**: dev-vpc, staging-vpc, prod-vpc
- **Private by default**: Only expose necessary Droplets to public internet
- **Use Cloud Firewalls**: Layer firewall rules on top of VPC isolation

### Cloud-Init Testing

1. Write cloud-init locally in a file (e.g., `user-data.yaml`)
2. Validate syntax: `cloud-init devel schema --config-file user-data.yaml`
3. Test in dev Droplet first before applying to staging/production
4. Check logs: `sudo cat /var/log/cloud-init-output.log`

### Volume Attachment

```yaml
# Create Volume first
apiVersion: code2cloud.planton.ai/v1
kind: DigitalOceanVolume
metadata:
  name: my-data-volume
spec:
  name: my-data
  region: digital_ocean_region_nyc3
  size: 100  # GB

---
# Then reference in Droplet
apiVersion: code2cloud.planton.ai/v1
kind: DigitalOceanDroplet
metadata:
  name: my-droplet
spec:
  # ... other fields ...
  volumeIds:
    - valueFromResourceOutput:
        resourceIdRef:
          name: my-data-volume
        outputKey: volume_id
```

### Backup Strategy

- **Development**: No backups (use snapshots manually if needed)
- **Staging**: Optional backups (depends on data criticality)
- **Production**: Always enable backups + manual snapshots before major changes

### Monitoring Integration

DigitalOcean monitoring provides:
- CPU usage
- Memory usage
- Disk usage
- Bandwidth (inbound/outbound)
- Disk I/O

Access metrics via:
- DigitalOcean Control Panel → Droplet → Graphs
- DigitalOcean API (`/v2/monitoring/metrics/droplet/...`)
- Third-party integrations (Datadog, Prometheus exporters)

### Cost Optimization

| Size | Monthly Cost | Use Case |
|------|--------------|----------|
| `s-1vcpu-1gb` | ~$6 | Dev/test |
| `s-2vcpu-4gb` | ~$18 | Staging, light production |
| `g-2vcpu-8gb` | ~$24 | Standard production |
| `g-4vcpu-16gb` | ~$48 | Databases, medium workloads |
| `c-8vcpu-16gb` | ~$96 | Compute-intensive tasks |

**Tip**: Use smaller Droplets in dev/staging, scale up only in production.

---

## Next Steps

1. Choose an example that matches your use case
2. Replace placeholder values (SSH keys, VPC IDs, credentials)
3. Test in development environment first
4. Gradually promote to staging, then production
5. Monitor Droplet health via DigitalOcean dashboard
6. Set up Cloud Firewalls for network security

## Related Resources

- **DigitalOcean VPC**: Create isolated private networks
- **DigitalOcean Volume**: Persistent block storage
- **DigitalOcean Cloud Firewall**: Network security rules
- **DigitalOcean Load Balancer**: Distribute traffic across Droplets

## Troubleshooting

For common issues and solutions, see the main README.md file in this directory.


# Preset: Production Web Server

A production-grade ECS instance with encrypted disks, public IP, deletion protection, and a RAM role.

## Use Case

- Internet-facing web applications
- Application servers requiring public IP
- Workloads needing persistent data storage with encryption
- Production environments with accidental-deletion safeguards

## Configuration

- **Instance Type**: `ecs.g7.2xlarge` (8 vCPU, 32 GiB memory)
- **Image**: Alibaba Cloud Linux 3
- **System Disk**: cloud_essd PL1, 100 GB, encrypted
- **Data Disk**: cloud_essd PL1, 200 GB, encrypted (for application data)
- **Authentication**: SSH key pair
- **Networking**: 20 Mbps public internet bandwidth (PayByTraffic)
- **Billing**: PostPaid (default)
- **Security**: Deletion protection enabled, disk encryption enabled
- **IAM**: RAM role attached for service-to-service authentication

## What's Not Included

- Spot pricing (production should be on-demand or PrePaid)
- PrePaid subscription (customize period and periodUnit if needed)
- Cloud-init user data

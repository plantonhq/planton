# GCP Cloud Router NAT - Terraform Module

This directory contains the Terraform implementation for deploying Google Cloud Router with Cloud NAT based on the `GcpRouterNat` API resource specification.

## Overview

The Terraform module provisions:
- **Google Cloud Router**: Regional router resource attached to the specified VPC network.
- **Google Compute Address (optional)**: Static external IP addresses if manual NAT IP allocation is specified.
- **Google Router NAT**: NAT gateway configuration with automatic or manual IP allocation, subnet coverage, and logging.

## Module Structure

```
iac/tf/
├── main.tf         # Resource definitions (router, static IPs, NAT)
├── variables.tf    # Input variable definitions
├── locals.tf       # Local value transformations
├── outputs.tf      # Output definitions
├── provider.tf     # GCP provider configuration
├── backend.tf      # Terraform backend configuration
└── README.md       # This file
```

## Inputs

### Required Variables

#### `metadata`
Resource metadata including name and labels.

```hcl
metadata = {
  name    = "prod-nat-uscentral1"
  id      = optional(string)
  org     = optional(string)
  env     = optional(string)
  labels  = optional(map(string))
  tags    = optional(list(string))
  version = optional(object({ id = string, message = string }))
}
```

#### `spec`
Resource specification for Cloud Router NAT.

```hcl
spec = {
  vpc_self_link          = string                            # Required: VPC network self-link
  region                 = string                            # Required: GCP region
  subnetwork_self_links  = optional(list(string), [])        # Optional: Specific subnets (default: all subnets)
  nat_ip_names           = optional(list(string), [])        # Optional: Static IP names (default: auto-allocate)
  log_filter             = optional(string, "ERRORS_ONLY")   # Optional: DISABLED, ERRORS_ONLY, or ALL
}
```

## Outputs

The module exports outputs matching `GcpRouterNatStackOutputs`:

- **`name`**: Name of the Cloud NAT gateway
- **`router_self_link`**: Self-link URL of the Cloud Router
- **`nat_ip_addresses`**: List of external IP addresses used by NAT (auto-allocated or static)

## Resource Creation

### 1. Cloud Router

```hcl
resource "google_compute_router" "router" {
  name    = var.metadata.name
  region  = var.spec.region
  network = var.spec.vpc_self_link
}
```

### 2. Static External IPs (Conditional)

Created only if `var.spec.nat_ip_names` is non-empty:

```hcl
resource "google_compute_address" "nat_ips" {
  count = length(var.spec.nat_ip_names)
  
  name         = var.spec.nat_ip_names[count.index]
  region       = var.spec.region
  address_type = "EXTERNAL"
  labels       = local.gcp_labels
}
```

### 3. Cloud NAT

```hcl
resource "google_compute_router_nat" "nat" {
  name   = var.metadata.name
  router = google_compute_router.router.name
  region = var.spec.region
  
  # IP allocation strategy
  nat_ip_allocate_option = local.nat_ip_allocate_option  # AUTO_ONLY or MANUAL_ONLY
  nat_ips                = local.nat_ip_allocate_option == "MANUAL_ONLY" ? google_compute_address.nat_ips[*].self_link : []
  
  # Subnet coverage
  source_subnetwork_ip_ranges_to_nat = local.source_subnetwork_ip_ranges_to_nat  # ALL or LIST_OF_SUBNETWORKS
  
  # Specific subnets (if LIST_OF_SUBNETWORKS)
  dynamic "subnetwork" {
    for_each = local.subnetworks
    content {
      name                    = subnetwork.value.name
      source_ip_ranges_to_nat = subnetwork.value.source_ip_ranges_to_nat
    }
  }
  
  # Logging configuration
  log_config {
    enable = local.enable_logging
    filter = local.log_filter
  }
  
  # Production defaults (GCP best practices)
  min_ports_per_vm                    = 64
  enable_endpoint_independent_mapping = true
  enable_dynamic_port_allocation      = false
  
  # Timeouts (GCP defaults)
  tcp_established_idle_timeout_sec = 1200  # 20 minutes
  tcp_transitory_idle_timeout_sec  = 30    # 30 seconds
  udp_idle_timeout_sec             = 30    # 30 seconds
  icmp_idle_timeout_sec            = 30    # 30 seconds
}
```

## Local Values

The `locals.tf` file computes derived values:

```hcl
locals {
  # Router and NAT names
  router_name = var.metadata.name
  nat_name    = var.metadata.name
  
  # NAT IP allocation strategy
  nat_ip_allocate_option = length(var.spec.nat_ip_names) > 0 ? "MANUAL_ONLY" : "AUTO_ONLY"
  
  # Subnet coverage mode
  source_subnetwork_ip_ranges_to_nat = length(var.spec.subnetwork_self_links) > 0 ? "LIST_OF_SUBNETWORKS" : "ALL_SUBNETWORKS_ALL_IP_RANGES"
  
  # Subnetworks configuration (only if specific subnets provided)
  subnetworks = length(var.spec.subnetwork_self_links) > 0 ? [
    for subnet in var.spec.subnetwork_self_links : {
      name                    = subnet
      source_ip_ranges_to_nat = ["ALL_IP_RANGES"]
    }
  ] : []
  
  # Logging configuration
  enable_logging = var.spec.log_filter != "DISABLED"
  log_filter     = var.spec.log_filter != "DISABLED" ? var.spec.log_filter : "ERRORS_ONLY"
  
  # GCP labels
  gcp_labels = merge(
    {
      "resource"      = "true"
      "resource-name" = var.metadata.name
      "resource-kind" = "gcprouternat"
    },
    var.metadata.org != null ? { "organization" = var.metadata.org } : {},
    var.metadata.env != null ? { "environment" = var.metadata.env } : {},
    var.metadata.id != null ? { "resource-id" = var.metadata.id } : {},
    var.metadata.labels != null ? var.metadata.labels : {}
  )
}
```

## Usage

### Minimal Configuration (Auto-Allocation, All Subnets)

```hcl
module "nat" {
  source = "path/to/module"
  
  metadata = {
    name = "dev-nat-uscentral1"
  }
  
  spec = {
    vpc_self_link = "https://www.googleapis.com/compute/v1/projects/my-project/global/networks/my-vpc"
    region        = "us-central1"
  }
}
```

### Manual IP Allocation for Partner Allowlisting

```hcl
module "nat" {
  source = "path/to/module"
  
  metadata = {
    name = "prod-nat-uscentral1"
    org  = "acme-corp"
    env  = "prod"
    labels = {
      cost-center = "cc-4510"
      compliance  = "pci-dss"
    }
  }
  
  spec = {
    vpc_self_link = "https://www.googleapis.com/compute/v1/projects/my-project/global/networks/prod-vpc"
    region        = "us-central1"
    nat_ip_names = [
      "prod-nat-ip-1-uscentral1",
      "prod-nat-ip-2-uscentral1"
    ]
    log_filter = "ERRORS_ONLY"
  }
}
```

### Specific Subnet Coverage

```hcl
module "nat" {
  source = "path/to/module"
  
  metadata = {
    name = "staging-nat-useast1"
  }
  
  spec = {
    vpc_self_link = "https://www.googleapis.com/compute/v1/projects/my-project/global/networks/staging-vpc"
    region        = "us-east1"
    subnetwork_self_links = [
      "https://www.googleapis.com/compute/v1/projects/my-project/regions/us-east1/subnetworks/private-subnet-1",
      "https://www.googleapis.com/compute/v1/projects/my-project/regions/us-east1/subnetworks/private-subnet-2"
    ]
    log_filter = "ERRORS_ONLY"
  }
}
```

### Accessing Outputs

```hcl
# Get NAT gateway name
output "nat_name" {
  value = module.nat.name
}

# Get Cloud Router self-link
output "router_self_link" {
  value = module.nat.router_self_link
}

# Get NAT IP addresses (for partner allowlisting)
output "nat_ip_addresses" {
  value = module.nat.nat_ip_addresses
}
```

## Implementation Details

### NAT IP Allocation Strategy

**Auto-Allocation (Default):**
- `nat_ip_names` is empty
- `nat_ip_allocate_option = "AUTO_ONLY"`
- Google automatically assigns and scales IPs
- No static IP resources created

**Manual Allocation:**
- `nat_ip_names` contains IP names
- `nat_ip_allocate_option = "MANUAL_ONLY"`
- Static IP resources created for each name
- NAT uses specified static IPs

### Subnet Coverage Strategy

**All Subnets (Default):**
- `subnetwork_self_links` is empty
- `source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"`
- All subnets in the region automatically covered

**Specific Subnets:**
- `subnetwork_self_links` contains subnet self-links
- `source_subnetwork_ip_ranges_to_nat = "LIST_OF_SUBNETWORKS"`
- Only specified subnets covered
- Each subnet configured with `source_ip_ranges_to_nat = ["ALL_IP_RANGES"]`

### Logging Configuration

**ERRORS_ONLY (Default):**
```hcl
log_config {
  enable = true
  filter = "ERRORS_ONLY"
}
```

**ALL (Full Logging):**
```hcl
log_config {
  enable = true
  filter = "ALL"
}
```

**DISABLED:**
```hcl
log_config {
  enable = false
  filter = "ERRORS_ONLY"  # Required even when disabled
}
```

## Production Defaults

The module applies production-ready defaults:

- **`min_ports_per_vm`**: 64 (sufficient for most workloads)
- **`enable_endpoint_independent_mapping`**: true (allows port reuse)
- **`enable_dynamic_port_allocation`**: false (static allocation for predictability)
- **TCP established timeout**: 1200 seconds (20 minutes)
- **TCP transitory timeout**: 30 seconds
- **UDP timeout**: 30 seconds
- **ICMP timeout**: 30 seconds

These values match GCP defaults and best practices.

## Prerequisites

### 1. Static IPs (if using manual allocation)

Create static IPs before deploying NAT:

```bash
gcloud compute addresses create prod-nat-ip-1-uscentral1 --region=us-central1
gcloud compute addresses create prod-nat-ip-2-uscentral1 --region=us-central1
```

Or use Terraform:

```hcl
resource "google_compute_address" "nat_ip_1" {
  name         = "prod-nat-ip-1-uscentral1"
  region       = "us-central1"
  address_type = "EXTERNAL"
}

resource "google_compute_address" "nat_ip_2" {
  name         = "prod-nat-ip-2-uscentral1"
  region       = "us-central1"
  address_type = "EXTERNAL"
}
```

### 2. VPC Network

Ensure the VPC network exists before deploying NAT:

```bash
gcloud compute networks describe my-vpc
```

### 3. GCP Provider Configuration

Configure the GCP provider in your Terraform configuration:

```hcl
provider "google" {
  project = "my-project"
  region  = "us-central1"
}
```

## Validation

### Terraform Validation

```bash
# Initialize Terraform
terraform init

# Validate configuration
terraform validate

# Plan changes
terraform plan

# Apply changes
terraform apply
```

### Testing

After deployment, verify NAT functionality:

```bash
# Check Cloud Router
gcloud compute routers describe <router-name> --region=<region>

# Check Cloud NAT
gcloud compute routers nats describe <nat-name> --router=<router-name> --region=<region>

# Check static IPs (if manual allocation)
gcloud compute addresses list --filter="region:<region>"

# Test egress from private VM
# SSH into private VM and check external IP:
curl ifconfig.me
# Should show NAT IP, not VM's internal IP
```

## Troubleshooting

### Error: "Network not found"
**Cause:** VPC self-link is incorrect or VPC doesn't exist.
**Solution:** Verify `spec.vpc_self_link` points to an existing VPC.

### Error: "IP address already exists"
**Cause:** Static IP name already exists in the region.
**Solution:** Use unique IP names or import existing IPs:
```bash
terraform import google_compute_address.nat_ips[0] projects/my-project/regions/us-central1/addresses/nat-ip-1
```

### Error: "Region mismatch"
**Cause:** Static IPs created in different region than NAT.
**Solution:** Ensure all resources use the same region.

### Drift Detection

If resources are modified outside Terraform (via Console or gcloud CLI):

```bash
# Detect drift
terraform plan

# Sync state
terraform refresh

# Reconcile
terraform apply
```

## Best Practices

### 1. Use Auto-Allocation Unless Allowlisting is Required

**Default (Recommended):**
```hcl
nat_ip_names = []  # or omit
```

**Manual Allocation (Only if needed):**
```hcl
nat_ip_names = ["nat-ip-1", "nat-ip-2"]
```

### 2. Cover All Subnets by Default

**Default (Recommended):**
```hcl
subnetwork_self_links = []  # or omit
```

**Specific Subnets (Only if needed):**
```hcl
subnetwork_self_links = ["subnet-1-self-link", "subnet-2-self-link"]
```

### 3. Enable ERRORS_ONLY Logging in Production

**Production (Recommended):**
```hcl
log_filter = "ERRORS_ONLY"  # or omit for default
```

**Development (Cost Optimization):**
```hcl
log_filter = "DISABLED"
```

**Security Auditing (High Cost):**
```hcl
log_filter = "ALL"
```

### 4. Use Remote State

Store Terraform state remotely for team collaboration:

```hcl
terraform {
  backend "gcs" {
    bucket = "my-terraform-state"
    prefix = "nat/uscentral1"
  }
}
```

### 5. Tag Resources with Labels

```hcl
metadata = {
  name = "prod-nat"
  org  = "acme-corp"
  env  = "prod"
  labels = {
    cost-center = "cc-4510"
    compliance  = "pci-dss"
  }
}
```

## Integration with Planton

This Terraform module is used by the Planton CLI when Terraform is selected as the IaC engine:

```bash
# Deploy NAT gateway (uses Terraform module)
planton apply -f nat-config.yaml --engine terraform

# Status
planton status gcprouternat prod-nat-uscentral1

# Destroy
planton destroy gcprouternat prod-nat-uscentral1 --engine terraform
```

## Related Documentation

- **API Specification**: [../../spec.proto](../../spec.proto)
- **Examples**: [../../examples.md](../../examples.md)
- **Overview**: [../../README.md](../../README.md)
- **Pulumi Module**: [../pulumi/README.md](../pulumi/README.md)
- **Pulumi Architecture**: [../pulumi/overview.md](../pulumi/overview.md)

## Maintenance

### Updating Provider Version

```hcl
terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"  # Update version here
    }
  }
}
```

### Adding New Features

To add new configuration options:

1. Update `spec.proto` with new fields
2. Regenerate Go stubs: `make protos`
3. Update `variables.tf` with new input variables
4. Update `locals.tf` if derived values are needed
5. Update `main.tf` resource arguments
6. Test with `terraform plan` and `terraform apply`

---

For architectural details about the Pulumi implementation, see [../pulumi/overview.md](../pulumi/overview.md).


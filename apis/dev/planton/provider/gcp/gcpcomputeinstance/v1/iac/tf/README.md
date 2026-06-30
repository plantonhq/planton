# GCP Compute Instance Terraform Module

This Terraform module deploys Google Compute Engine instances using the Planton framework.

## Overview

The module creates a Compute Engine VM instance with configurable:
- Machine type and zone
- Boot disk (image, size, type)
- Network interfaces with optional external IPs
- Service account and OAuth scopes
- Scheduling options (Spot/Preemptible VMs)
- Labels, tags, and metadata
- Startup scripts
- Additional attached disks

## Usage

### With Planton CLI

```bash
# Deploy using Terraform
planton tofu apply --manifest manifest.yaml --auto-approve

# Plan changes
planton tofu plan --manifest manifest.yaml

# Destroy resources
planton tofu destroy --manifest manifest.yaml --auto-approve
```

### Standalone Usage

1. Create a `terraform.tfvars` file:

```hcl
metadata = {
  name = "my-vm"
}

spec = {
  project_id = {
    value = "my-gcp-project"
  }
  zone         = "us-central1-a"
  machine_type = "e2-medium"

  boot_disk = {
    image   = "debian-cloud/debian-11"
    size_gb = 20
    type    = "pd-ssd"
  }

  network_interfaces = [
    {
      network = {
        value = "default"
      }
      access_configs = [
        {
          network_tier = "PREMIUM"
        }
      ]
    }
  ]
}
```

2. Run Terraform:

```bash
terraform init
terraform plan
terraform apply
```

## Requirements

| Name | Version |
|------|---------|
| terraform | >= 1.0 |
| google | >= 6.0 |

## Providers

| Name | Version |
|------|---------|
| google | >= 6.0 |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| metadata | Metadata for the resource | object | n/a | yes |
| spec | Specification for GCP Compute Instance | object | n/a | yes |

### Metadata Object

| Name | Description | Type | Required |
|------|-------------|------|:--------:|
| name | Instance name | string | yes |
| id | Resource ID | string | no |
| org | Organization | string | no |
| env | Environment | string | no |
| labels | Additional labels | map(string) | no |
| tags | Additional tags | list(string) | no |

### Spec Object

| Name | Description | Type | Required |
|------|-------------|------|:--------:|
| project_id | GCP project ID (StringValueOrRef) | object | yes |
| zone | Deployment zone | string | yes |
| machine_type | Machine type | string | yes |
| boot_disk | Boot disk configuration | object | yes |
| network_interfaces | Network interface configurations | list(object) | yes |
| attached_disks | Additional disks | list(object) | no |
| service_account | Service account configuration | object | no |
| preemptible | Use preemptible VM | bool | no |
| spot | Use Spot VM | bool | no |
| deletion_protection | Enable deletion protection | bool | no |
| metadata | Custom metadata | map(string) | no |
| labels | Instance labels | map(string) | no |
| tags | Network tags | list(string) | no |
| ssh_keys | SSH keys | list(string) | no |
| startup_script | Startup script | string | no |
| allow_stopping_for_update | Allow stopping for updates | bool | no |
| scheduling | Scheduling configuration | object | no |

## Outputs

| Name | Description |
|------|-------------|
| instance_name | Name of the instance |
| instance_id | Unique instance ID |
| self_link | Full self-link URL |
| internal_ip | Internal (private) IP |
| external_ip | External (public) IP |
| status | Instance status |
| zone | Deployment zone |
| machine_type | Machine type |
| cpu_platform | CPU platform |

## Examples

### Basic Instance

```hcl
metadata = {
  name = "basic-vm"
}

spec = {
  project_id = { value = "my-project" }
  zone       = "us-central1-a"
  machine_type = "e2-micro"
  
  boot_disk = {
    image = "debian-cloud/debian-11"
  }
  
  network_interfaces = [
    { network = { value = "default" } }
  ]
}
```

### Spot VM

```hcl
metadata = {
  name = "spot-vm"
}

spec = {
  project_id   = { value = "my-project" }
  zone         = "us-central1-a"
  machine_type = "e2-standard-4"
  spot         = true
  
  boot_disk = {
    image = "debian-cloud/debian-11"
  }
  
  network_interfaces = [
    { network = { value = "default" } }
  ]
  
  scheduling = {
    preemptible                 = true
    automatic_restart           = false
    on_host_maintenance         = "TERMINATE"
    provisioning_model          = "SPOT"
    instance_termination_action = "STOP"
  }
}
```

### Production Instance

```hcl
metadata = {
  name = "prod-server"
  org  = "my-org"
  env  = "production"
}

spec = {
  project_id   = { value = "prod-project" }
  zone         = "us-central1-a"
  machine_type = "n2-standard-4"
  
  boot_disk = {
    image   = "debian-cloud/debian-11"
    size_gb = 100
    type    = "pd-ssd"
  }
  
  network_interfaces = [
    {
      network    = { value = "projects/prod-project/global/networks/prod-vpc" }
      subnetwork = { value = "projects/prod-project/regions/us-central1/subnetworks/prod-subnet" }
      access_configs = [
        { network_tier = "PREMIUM" }
      ]
    }
  ]
  
  service_account = {
    email  = "prod-sa@prod-project.iam.gserviceaccount.com"
    scopes = ["https://www.googleapis.com/auth/cloud-platform"]
  }
  
  deletion_protection       = true
  allow_stopping_for_update = true
  
  labels = {
    app  = "webserver"
    team = "platform"
  }
  
  tags = ["web-server", "https-server"]
  
  metadata = {
    enable-oslogin = "TRUE"
  }
}
```

## Notes

- **StringValueOrRef**: Currently only literal values are supported. Reference resolution (value_from) is planned for a future release.
- **Service Account**: If not specified, the default Compute Engine service account is used.
- **External IP**: To create an instance without external IP, provide an empty `access_configs` array.
- **Spot VMs**: When using Spot VMs, `automatic_restart` must be false and `on_host_maintenance` must be "TERMINATE".

# Terraform Implementation - GCP Cert Manager Cert

This directory contains the Terraform implementation for the **GcpCertManagerCert** resource.

## Overview

The Terraform implementation provides a declarative way to deploy GCP SSL/TLS certificates using HCL. It leverages the official Google Terraform provider to create Certificate Manager certificates or Load Balancer certificates with automatic DNS validation.

## Directory Structure

```
tf/
├── main.tf          # Certificate resources
├── variables.tf     # Input variable definitions
├── outputs.tf       # Output value definitions
├── locals.tf        # Local variable definitions
├── provider.tf      # Provider configuration
├── README.md        # This file
└── examples.md      # Usage examples
```

## Files Overview

### main.tf

Contains the main resource definitions:

- **Certificate Manager Resources** (when certificateType = 0 or MANAGED):
  - `google_certificate_manager_dns_authorization`: DNS authorizations for each domain
  - `google_dns_record_set`: DNS validation records in Cloud DNS
  - `google_certificate_manager_certificate`: The certificate itself

- **Load Balancer Resources** (when certificateType = 1 or LOAD_BALANCER):
  - `google_compute_managed_ssl_certificate`: Google-managed SSL certificate

### variables.tf

Defines input variables:
- `metadata`: Resource metadata (name, id, org, env)
- `spec`: Certificate specification (domains, project, DNS zone, type)

### outputs.tf

Defines output values:
- `certificate-id`: Certificate ID
- `certificate-name`: Certificate name
- `certificate-domain-name`: Primary domain
- `certificate-status`: Certificate status

### locals.tf

Defines local variables:
- `gcp_labels`: Labels for GCP resources
- `all_domains`: Combined primary and alternate domains
- `is_managed`: Boolean for certificate type

### provider.tf

Specifies required providers:
- `google`: ~> 5.0
- `google-beta`: ~> 5.0

## Prerequisites

- Terraform 1.0+
- GCP account with appropriate permissions
- Cloud DNS managed zone for validation
- GCP credentials configured

## Required GCP Permissions

The service account needs:

- `roles/certificatemanager.editor` (for Certificate Manager)
- `roles/compute.loadBalancerAdmin` (for LB certificates)
- `roles/dns.admin` (for DNS records)

## Usage

### Direct Terraform Usage

1. **Initialize Terraform**:
   ```bash
   terraform init
   ```

2. **Create tfvars file**:
   Create `terraform.tfvars`:
   ```hcl
   metadata = {
     name = "my-cert"
     id   = "cert-001"
     org  = "my-org"
     env = {
       id = "production"
     }
   }

   spec = {
     gcp_project_id         = "my-gcp-project"
     primary_domain_name    = "example.com"
     alternate_domain_names = ["www.example.com"]
     cloud_dns_zone_id = {
       value = "example-com-zone"
     }
     certificate_type  = 0  # MANAGED
     validation_method = "DNS"
   }
   ```

3. **Plan changes**:
   ```bash
   terraform plan
   ```

4. **Apply changes**:
   ```bash
   terraform apply
   ```

5. **View outputs**:
   ```bash
   terraform output certificate-id
   ```

6. **Destroy resources**:
   ```bash
   terraform destroy
   ```

### Using Planton CLI (Recommended)

```bash
planton terraform apply --manifest cert.yaml --stack org/project/stack
```

## Certificate Types

### MANAGED (certificateType = 0)

Uses Google Certificate Manager:
- Modern certificate management
- Explicit DNS authorizations
- Works with various GCP services
- Recommended for most use cases

Resources created:
- DNS authorization per domain
- DNS validation records
- Certificate Manager certificate

### LOAD_BALANCER (certificateType = 1)

Uses Google-managed SSL certificates:
- Optimized for load balancers
- Automatic provisioning
- Simpler configuration

Resources created:
- Google-managed SSL certificate

## Input Variables

### metadata

```hcl
metadata = {
  name = "my-cert"      # Resource name
  id   = "cert-001"     # Resource ID
  org  = "my-org"       # Organization
  env = {
    id = "production"   # Environment
  }
}
```

### spec

```hcl
spec = {
  gcp_project_id         = "my-project"
  primary_domain_name    = "example.com"
  alternate_domain_names = ["www.example.com", "api.example.com"]
  cloud_dns_zone_id = {
    value = "example-zone"
  }
  certificate_type  = 0      # 0=MANAGED, 1=LOAD_BALANCER
  validation_method = "DNS"
}
```

## Outputs

After applying, these outputs are available:

```bash
terraform output certificate-id           # Certificate ID
terraform output certificate-name         # Certificate name
terraform output certificate-domain-name  # Primary domain
terraform output certificate-status       # Status
```

## Examples

### Wildcard Certificate

```hcl
spec = {
  gcp_project_id      = "my-project"
  primary_domain_name = "*.example.com"
  cloud_dns_zone_id = {
    value = "example-zone"
  }
  certificate_type = 0
}
```

### Multi-Domain Certificate

```hcl
spec = {
  gcp_project_id      = "my-project"
  primary_domain_name = "example.com"
  alternate_domain_names = [
    "www.example.com",
    "api.example.com",
    "*.services.example.com"
  ]
  cloud_dns_zone_id = {
    value = "example-zone"
  }
  certificate_type = 0
}
```

### Load Balancer Certificate

```hcl
spec = {
  gcp_project_id      = "my-project"
  primary_domain_name = "lb.example.com"
  alternate_domain_names = ["www.lb.example.com"]
  cloud_dns_zone_id = {
    value = "example-zone"
  }
  certificate_type = 1  # LOAD_BALANCER
}
```

## State Management

Terraform state can be stored:

- **Locally**: Default, state in `terraform.tfstate`
- **Remote**: Use backend configuration (GCS, S3, etc.)

Example GCS backend:

```hcl
terraform {
  backend "gcs" {
    bucket = "my-terraform-state"
    prefix = "gcp-cert"
  }
}
```

## Verification

Check resources in GCP Console:

1. Navigate to Certificate Manager or Load Balancing
2. Verify certificate exists and is active
3. Check Cloud DNS for validation records

Using gcloud:

```bash
# List Certificate Manager certificates
gcloud certificate-manager certificates list --project=my-project

# List managed SSL certificates
gcloud compute ssl-certificates list --project=my-project

# Check DNS records
gcloud dns record-sets list --zone=my-zone --project=my-project
```

## Troubleshooting

### DNS Validation Pending

- Verify DNS records created successfully
- Check domain ownership
- Wait for DNS propagation (up to 10 minutes)

### Permission Denied

Ensure service account has required roles:
```bash
gcloud projects add-iam-policy-binding my-project \
  --member="serviceAccount:sa@project.iam.gserviceaccount.com" \
  --role="roles/certificatemanager.editor"
```

### Resource Already Exists

If certificate name conflicts:
- Change metadata.name in your configuration
- Or delete existing certificate manually

## Best Practices

1. **Use Remote State**: Store state in GCS for team collaboration
2. **Version Control**: Keep Terraform files in Git
3. **Separate Environments**: Use workspaces or separate directories
4. **Plan Before Apply**: Always run `terraform plan` first
5. **Use Variables**: Keep sensitive data in tfvars files (gitignored)

## Integration with Planton

This Terraform module integrates with Planton CLI, which:

- Converts YAML to tfvars
- Manages Terraform state
- Handles credentials
- Provides consistent interface

## Next Steps

- See [examples.md](examples.md) for more usage patterns
- Check [overview.md](../overview.md) for IaC comparison
- Use Planton CLI for production workflows


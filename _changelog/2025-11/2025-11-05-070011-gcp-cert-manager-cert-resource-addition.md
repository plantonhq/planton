# GCP Certificate Manager Cert Resource Addition

**Date**: November 5, 2025  
**Type**: New Feature  
**Components**: API Definitions, GCP Provider, Cloud Resource Kind Registry, IaC Modules

## Summary

Added `GcpCertManagerCert`, a new cloud resource for provisioning and managing SSL/TLS certificates on Google Cloud Platform. The resource supports both Google Certificate Manager (modern, feature-rich) and Google-managed SSL certificates for load balancers (classic), with automatic DNS validation through Google Cloud DNS. This enables teams to declaratively manage SSL/TLS certificates as infrastructure-as-code alongside other GCP resources.

## Motivation

### Problem Statement

Organizations deploying applications on GCP need SSL/TLS certificates for secure HTTPS communication. Managing certificates manually through the GCP Console involves:

- **Manual provisioning**: Creating certificates through the UI or gcloud commands
- **Manual DNS validation**: Adding DNS records to prove domain ownership
- **No version control**: Certificate configurations aren't tracked in Git
- **Difficult replication**: Hard to replicate certificate setup across environments
- **Limited automation**: No declarative, repeatable deployment process
- **Fragmented management**: Certificates managed separately from infrastructure

### Use Cases

1. **Secure Cloud Run Services**: Provision SSL certificates for Cloud Run custom domains
2. **Load Balancer SSL**: Create managed SSL certificates for GCP load balancers
3. **Multi-Environment Deployments**: Deploy certificates consistently across dev/staging/production
4. **Wildcard Certificates**: Protect entire subdomains with wildcard certificates
5. **Multi-Domain Certificates**: Secure multiple domains with a single certificate using SANs
6. **Infrastructure as Code**: Manage certificates alongside other GCP resources in Git

## Solution / What's New

### New Cloud Resource: GcpCertManagerCert

A fully-featured cloud resource that provisions SSL/TLS certificates on GCP with automatic DNS validation, following Planton's uniform resource model.

**Key Capabilities**:
- ✅ Dual certificate type support (Certificate Manager + Load Balancer)
- ✅ Automatic DNS validation with Cloud DNS integration
- ✅ Multi-domain support with Subject Alternative Names (SANs)
- ✅ Wildcard certificate support (`*.example.com`)
- ✅ Foreign key references to DNS zones
- ✅ Complete IaC support (Pulumi + Terraform)
- ✅ Comprehensive validation rules
- ✅ Full documentation and examples

## Implementation Details

### 1. Protocol Buffer Definitions

**Location**: `apis/project/planton/provider/gcp/gcpcertmanagercert/v1/`

#### api.proto

Defines the main resource structure:

```proto
message GcpCertManagerCert {
  string api_version = 1 [(buf.validate.field).string.const = 'gcp.planton.dev/v1'];
  string kind = 2 [(buf.validate.field).string.const = 'GcpCertManagerCert'];
  CloudResourceMetadata metadata = 3 [(buf.validate.field).required = true];
  GcpCertManagerCertSpec spec = 4 [(buf.validate.field).required = true];
  GcpCertManagerCertStatus status = 5;
}
```

#### spec.proto

Defines certificate configuration with comprehensive validations:

```proto
message GcpCertManagerCertSpec {
  // GCP project where certificate will be created
  string gcp_project_id = 1 [(buf.validate.field).required = true];
  
  // Primary domain (supports wildcards: *.example.com)
  string primary_domain_name = 2 [
    (buf.validate.field).required = true,
    (buf.validate.field).string = {pattern: "^(?:\\*\\.[A-Za-z0-9\\-\\.]+|[A-Za-z0-9\\-\\.]+\\.[A-Za-z]{2,})$"}
  ];
  
  // Optional alternate domains (SANs)
  repeated string alternate_domain_names = 3 [
    (buf.validate.field).repeated.unique = true,
    (buf.validate.field).repeated.items = {
      string: {pattern: "^(?:\\*\\.[A-Za-z0-9\\-\\.]+|[A-Za-z0-9\\-\\.]+\\.[A-Za-z]{2,})$"}
    }
  ];
  
  // Cloud DNS zone for validation (supports foreign key references)
  StringValueOrRef cloud_dns_zone_id = 4 [
    (buf.validate.field).required = true,
    (default_kind) = GcpDnsZone,
    (default_kind_field_path) = "status.outputs.zone_name"
  ];
  
  // Certificate type: MANAGED (default) or LOAD_BALANCER
  optional CertificateType certificate_type = 5;
  
  // Validation method: DNS (default)
  optional string validation_method = 6 [
    (default) = "DNS",
    (buf.validate.field).string = {in: ["DNS"]}
  ];
}

enum CertificateType {
  MANAGED = 0;         // Google Certificate Manager
  LOAD_BALANCER = 1;   // Google-managed SSL for LBs
}
```

**Validation Features**:
- Domain name regex patterns supporting wildcards
- Unique constraint on alternate domain names
- Required fields for essential configuration
- Enum validation for certificate types
- Foreign key support for DNS zone references

#### stack_outputs.proto

Defines outputs returned after provisioning:

```proto
message GcpCertManagerCertStackOutputs {
  string certificate_id = 1;           // Certificate resource ID
  string certificate_name = 2;         // Full resource name
  string certificate_domain_name = 3;  // Primary domain
  string certificate_status = 4;       // PROVISIONING/ACTIVE
}
```

### 2. Cloud Resource Kind Registration

**File**: `apis/project/planton/shared/cloudresourcekind/cloud_resource_kind.proto`

Added enum entry in the GCP range (600-799):

```proto
GcpCertManagerCert = 619 [(kind_meta) = {
  provider: gcp
  version: v1
  id_prefix: "gcpcert"
}];
```

- **Enum value**: 619 (next available in GCP range)
- **ID prefix**: `gcpcert` (follows GCP naming convention)
- **Provider**: gcp
- **Version**: v1

### 3. Pulumi Implementation (Go)

**Location**: `apis/project/planton/provider/gcp/gcpcertmanagercert/v1/iac/pulumi/`

#### Module Structure

```
pulumi/
├── main.go                    # Entry point
├── module/
│   ├── main.go               # Provider setup and orchestration
│   ├── cert_manager_cert.go  # Certificate creation logic
│   ├── locals.go             # Local variables and GCP labels
│   └── outputs.go            # Output constants
├── Pulumi.yaml               # Pulumi project config
├── Makefile                  # Helper commands
└── debug.sh                  # Debug script
```

#### Key Implementation: cert_manager_cert.go

**Dual Certificate Type Support**:

```go
func certManagerCert(ctx *pulumi.Context, locals *Locals, provider *gcp.Provider) error {
    spec := locals.GcpCertManagerCert.Spec
    
    // Default to MANAGED if not specified
    certType := gcpcertmanagercertv1.CertificateType_MANAGED
    if spec.CertificateType != nil {
        certType = *spec.CertificateType
    }
    
    // Collect all domains
    allDomains := []string{spec.PrimaryDomainName}
    allDomains = append(allDomains, spec.AlternateDomainNames...)
    
    switch certType {
    case gcpcertmanagercertv1.CertificateType_MANAGED:
        return createManagedCertificate(ctx, locals, provider, allDomains, spec)
    case gcpcertmanagercertv1.CertificateType_LOAD_BALANCER:
        return createLoadBalancerCertificate(ctx, locals, provider, allDomains, spec)
    }
}
```

**Certificate Manager Implementation**:

Creates DNS authorizations and validation records:

```go
func createManagedCertificate(...) error {
    // Create DNS authorization for each domain
    for i, domain := range allDomains {
        dnsAuth, _ := certificatemanager.NewDnsAuthorization(ctx, ...)
        
        // Create DNS validation record in Cloud DNS
        dns.NewRecordSet(ctx, ..., &dns.RecordSetArgs{
            Name: dnsAuth.DnsResourceRecords.ApplyT(...),
            Type: dnsAuth.DnsResourceRecords.ApplyT(...),
            Rrdatas: dnsAuth.DnsResourceRecords.ApplyT(...),
            ManagedZone: pulumi.String(spec.CloudDnsZoneId.GetValue()),
        })
    }
    
    // Create Certificate Manager certificate
    cert, _ := certificatemanager.NewCertificate(ctx, ..., &certificatemanager.CertificateArgs{
        Managed: &certificatemanager.CertificateManagedArgs{
            Domains: pulumi.ToStringArray(allDomains),
            DnsAuthorizations: ...,
        },
    })
    
    return nil
}
```

**Load Balancer Implementation**:

Simpler approach for LB-specific certificates:

```go
func createLoadBalancerCertificate(...) error {
    cert, _ := compute.NewManagedSslCertificate(ctx, ..., &compute.ManagedSslCertificateArgs{
        Managed: &compute.ManagedSslCertificateManagedArgs{
            Domains: pulumi.ToStringArray(allDomains),
        },
    })
    return nil
}
```

**GCP Labels**:

Follows standard GCP module pattern:

```go
func initializeLocals(...) *Locals {
    locals.GcpLabels = map[string]string{
        gcplabelkeys.Resource:     strconv.FormatBool(true),
        gcplabelkeys.ResourceName: target.Metadata.Name,
        gcplabelkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_GcpCertManagerCert.String()),
    }
    
    if target.Metadata.Id != "" {
        locals.GcpLabels[gcplabelkeys.ResourceId] = target.Metadata.Id
    }
    // ... conditional org, env labels
}
```

### 4. Terraform Implementation

**Location**: `apis/project/planton/provider/gcp/gcpcertmanagercert/v1/iac/tf/`

#### main.tf

Conditional resource creation based on certificate type:

```hcl
locals {
  all_domains = concat([var.spec.primary_domain_name], var.spec.alternate_domain_names)
  is_managed = var.spec.certificate_type == null || var.spec.certificate_type == 0
}

# Certificate Manager (MANAGED)
resource "google_certificate_manager_dns_authorization" "dns_auth" {
  for_each = local.is_managed ? toset(local.all_domains) : toset([])
  
  name   = "${var.metadata.name}-${replace(each.value, "*", "wildcard")}-dns-auth"
  domain = each.value
  labels = local.gcp_labels
}

resource "google_dns_record_set" "validation_records" {
  for_each = local.is_managed ? google_certificate_manager_dns_authorization.dns_auth : {}
  
  name         = each.value.dns_resource_record[0].name
  type         = each.value.dns_resource_record[0].type
  rrdatas      = [each.value.dns_resource_record[0].data]
  managed_zone = var.spec.cloud_dns_zone_id.value
}

resource "google_certificate_manager_certificate" "cert" {
  count = local.is_managed ? 1 : 0
  
  managed {
    domains            = local.all_domains
    dns_authorizations = [for auth in google_certificate_manager_dns_authorization.dns_auth : auth.id]
  }
}

# Load Balancer (LOAD_BALANCER)
resource "google_compute_managed_ssl_certificate" "lb_cert" {
  count = !local.is_managed ? 1 : 0
  
  managed {
    domains = local.all_domains
  }
}
```

### 5. Testing

**File**: `apis/project/planton/provider/gcp/gcpcertmanagercert/v1/spec_test.go`

Comprehensive unit tests covering:

```go
func TestGcpCertManagerCertSpec_Validation(t *testing.T) {
    tests := []struct {
        name    string
        spec    *GcpCertManagerCertSpec
        wantErr bool
    }{
        {"valid spec with primary domain only", ...},
        {"valid spec with wildcard domain", ...},
        {"valid spec with alternate domains", ...},
        {"missing gcp project id", ...},
        {"missing primary domain name", ...},
        {"invalid primary domain pattern", ...},
        {"duplicate alternate domain names", ...},
    }
}

func TestCertificateType_Values(t *testing.T) {
    // Tests enum values
}

func TestGcpCertManagerCert_Structure(t *testing.T) {
    // Tests complete resource structure
}
```

**Test Coverage**:
- ✅ Valid configurations (single domain, wildcard, SANs)
- ✅ Missing required fields
- ✅ Invalid domain patterns
- ✅ Duplicate detection
- ✅ Enum values
- ✅ Complete resource structure

### 6. Documentation

#### README.md (674 lines)

Comprehensive documentation including:

- **Overview**: Purpose and benefits
- **Key Features**: Dual types, DNS validation, multi-domain support
- **Certificate Types**: When to use MANAGED vs LOAD_BALANCER
- **DNS Validation**: How automatic validation works
- **Best Practices**: Certificate management recommendations
- **Example Usage**: YAML manifests and CLI commands

#### examples.md

Nine complete examples:

1. Basic single domain (Certificate Manager)
2. Multiple domains with SANs
3. Wildcard domain certificate
4. Load Balancer certificate
5. Multi-domain wildcard
6. Development environment setup
7. Using foreign key references
8. Minimal configuration
9. Production deployment

#### IaC Documentation

- `iac/overview.md`: Pulumi vs Terraform comparison
- `iac/pulumi/README.md`: Pulumi implementation details
- `iac/pulumi/examples.md`: Pulumi usage examples
- `iac/tf/README.md`: Terraform implementation details
- `iac/tf/examples.md`: Terraform usage examples

## Example Usage

### Basic Certificate (Certificate Manager)

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpCertManagerCert
metadata:
  name: my-cert
  org: my-org
  env: production
spec:
  gcpProjectId: my-gcp-project
  primaryDomainName: example.com
  cloudDnsZoneId:
    value: example-com-zone
  certificateType: MANAGED
```

### Wildcard Certificate with SANs

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpCertManagerCert
metadata:
  name: wildcard-cert
  org: my-org
  env: production
spec:
  gcpProjectId: my-gcp-project
  primaryDomainName: "*.example.com"
  alternateDomainNames:
    - example.com
    - "*.api.example.com"
  cloudDnsZoneId:
    value: example-com-zone
  certificateType: MANAGED
```

### Load Balancer Certificate

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpCertManagerCert
metadata:
  name: lb-cert
  org: my-org
  env: production
spec:
  gcpProjectId: my-gcp-project
  primaryDomainName: lb.example.com
  alternateDomainNames:
    - www.lb.example.com
  cloudDnsZoneId:
    value: example-com-zone
  certificateType: LOAD_BALANCER
```

### Deployment

```bash
# Using Pulumi
planton pulumi up --manifest cert.yaml --stack org/project/production

# Using Terraform
planton terraform apply --manifest cert.yaml --stack org/project/production
```

## Benefits

### For Platform Engineers

1. **Infrastructure as Code**: Certificates managed alongside other GCP resources in Git
2. **Consistent Deployments**: Same manifest deploys across all environments
3. **Automation**: No manual DNS record creation or validation steps
4. **Version Control**: Full audit trail of certificate configurations
5. **Reusability**: Templates can be shared across teams and projects

### For Security Teams

1. **Automated Validation**: DNS validation happens automatically, reducing manual errors
2. **Policy Enforcement**: Validations ensure certificate configs meet standards
3. **Audit Trail**: Git history tracks all certificate changes
4. **Wildcard Management**: Secure entire subdomains with wildcard certificates
5. **Centralized Management**: All certificates defined in one place

### For Development Teams

1. **Self-Service**: Developers can provision certificates without manual intervention
2. **Quick Setup**: Deploy certificates with a single command
3. **Environment Parity**: Dev/staging/prod certificates use same pattern
4. **Clear Configuration**: YAML manifests are easy to read and understand
5. **Foreign Key References**: Link to existing DNS zones declaratively

## Technical Highlights

### 1. Dual Certificate Type Architecture

Supports two distinct GCP certificate types in one resource:

- **Certificate Manager (MANAGED)**: Modern API with advanced features
- **Load Balancer SSL (LOAD_BALANCER)**: Classic approach for LB-specific use

Users choose based on their needs via `certificateType` field.

### 2. Automatic DNS Validation

For Certificate Manager certificates:
1. Creates DNS authorization for each domain
2. Automatically adds validation records to Cloud DNS
3. Waits for GCP to verify domain ownership
4. Certificate becomes active after validation

No manual DNS record management required.

### 3. Foreign Key Support

Cloud DNS zone can be specified as:

**Direct value**:
```yaml
cloudDnsZoneId:
  value: "my-zone"
```

**Reference to another resource**:
```yaml
cloudDnsZoneId:
  kind: GcpDnsZone
  name: my-dns-zone-resource
  fieldPath: status.outputs.zone_name
```

Enables declarative dependencies between resources.

### 4. Comprehensive Validations

Using buf.validate:
- Domain name regex patterns (with wildcard support)
- Uniqueness constraints (no duplicate SANs)
- Required field enforcement
- Enum value validation
- Pattern matching for domain formats

### 5. Standard GCP Module Patterns

Follows established patterns from other GCP resources:
- Standard label keys and formatting
- Conditional label addition (Id, Org, Env)
- Provider configuration handling
- Output export patterns
- Error wrapping and context

## Code Metrics

### Files Created

**Protocol Buffers** (4 files):
- `api.proto` (29 lines)
- `spec.proto` (77 lines)
- `stack_outputs.proto` (24 lines)
- `stack_input.proto` (14 lines)

**Pulumi Module** (7 files):
- `main.go` (22 lines) - Entry point
- `module/main.go` (48 lines) - Provider setup
- `module/cert_manager_cert.go` (158 lines) - Core logic
- `module/locals.go` (45 lines) - Local variables
- `module/outputs.go` (17 lines) - Output constants
- `Pulumi.yaml` (3 lines)
- `Makefile` (23 lines)

**Terraform Module** (7 files):
- `main.tf` (74 lines) - Resource definitions
- `variables.tf` (22 lines) - Input variables
- `outputs.tf` (30 lines) - Output values
- `locals.tf` (9 lines) - Local variables
- `provider.tf` (17 lines) - Provider config
- `README.md` (185 lines)
- `examples.md` (302 lines)

**Documentation** (5 files):
- `README.md` (674 lines) - Main documentation
- `examples.md` (422 lines) - Usage examples
- `iac/overview.md` (147 lines) - IaC comparison
- `iac/pulumi/README.md` (157 lines)
- `iac/pulumi/examples.md` (198 lines)

**Testing** (1 file):
- `spec_test.go` (246 lines) - Unit tests

### Total New Code

- **Proto definitions**: ~144 lines
- **Pulumi implementation**: ~313 lines
- **Terraform implementation**: ~152 lines
- **Documentation**: ~2,085 lines
- **Tests**: ~246 lines
- **Total**: ~2,940 lines

## Integration Points

### With Existing Resources

**GcpDnsZone Integration**:
```yaml
# DNS Zone
apiVersion: gcp.planton.dev/v1
kind: GcpDnsZone
metadata:
  name: my-dns-zone
spec:
  projectId: my-project

---
# Certificate references DNS Zone
apiVersion: gcp.planton.dev/v1
kind: GcpCertManagerCert
metadata:
  name: my-cert
spec:
  cloudDnsZoneId:
    kind: GcpDnsZone
    name: my-dns-zone
    fieldPath: status.outputs.zone_name
```

### With Cloud Resource Kind Registry

Registered as enum value 619 in the GCP range, enabling:
- Automatic ID generation with `gcpcert` prefix
- Resource kind identification in labels
- Integration with resource management systems
- Consistent resource type handling

### With Planton CLI

Deploys seamlessly through standard CLI commands:
```bash
planton pulumi up --manifest cert.yaml
planton terraform apply --manifest cert.yaml
```

## Comparison with Manual Approach

### Before (Manual GCP Console)

1. Navigate to Certificate Manager
2. Click "Create Certificate"
3. Enter domain names manually
4. Choose validation method
5. Wait for DNS instructions
6. Manually add DNS records to Cloud DNS
7. Wait for validation (manually check status)
8. No version control
9. Hard to replicate across environments
10. No audit trail

### After (GcpCertManagerCert)

1. Write YAML manifest
2. Run `planton pulumi up`
3. DNS records created automatically
4. Validation happens automatically
5. Version controlled in Git
6. Repeatable across environments
7. Full audit trail
8. Self-documenting configuration

## Best Practices Enabled

### 1. Environment Parity

Same certificate pattern across all environments:

```yaml
# Production
spec:
  gcpProjectId: prod-project
  primaryDomainName: app.example.com

# Staging  
spec:
  gcpProjectId: staging-project
  primaryDomainName: app-staging.example.com
```

### 2. Wildcard for Microservices

Secure all microservices with one wildcard:

```yaml
spec:
  primaryDomainName: "*.services.example.com"
  alternateDomainNames:
    - services.example.com  # Apex domain too
```

### 3. Multi-Environment DNS Zones

```yaml
# Dev environment references dev DNS zone
cloudDnsZoneId:
  kind: GcpDnsZone
  name: dev-dns-zone
```

### 4. GitOps Workflow

```bash
# PR-based changes
git checkout -b add-new-cert
# Edit cert.yaml
git commit -m "Add SSL cert for new service"
git push && open PR
# Auto-deploy after merge
```

## Known Limitations

1. **DNS Provider**: Currently supports Cloud DNS only (not external DNS providers)
2. **Certificate Type Migration**: Changing certificate type requires resource recreation
3. **Domain Limit**: GCP limits on domains per certificate apply
4. **Validation Time**: Initial DNS validation can take 5-10 minutes
5. **Regional Limitations**: Certificate Manager availability varies by region

## Future Enhancements

1. **Multi-DNS Provider Support**: Support external DNS providers (Cloudflare, Route53)
2. **Certificate Renewal Alerts**: Proactive notifications before renewal
3. **Certificate Sharing**: Reference certificates across multiple resources
4. **Custom Validation**: Support for HTTP validation in addition to DNS
5. **Certificate Monitoring**: Integration with monitoring systems for expiry tracking
6. **Certificate Maps**: Support for Certificate Manager certificate maps
7. **Regional Certificates**: Support for regional (non-global) certificates

## Migration Guide

For teams currently managing GCP certificates manually:

### Step 1: Identify Existing Certificates

```bash
gcloud certificate-manager certificates list --project=my-project
```

### Step 2: Create Manifest

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpCertManagerCert
metadata:
  name: existing-cert
spec:
  gcpProjectId: my-project
  primaryDomainName: example.com
  cloudDnsZoneId:
    value: existing-zone
```

### Step 3: Import (if needed)

For existing certificates, consider:
- Creating new certificate with same config
- Swapping references after validation
- Removing old certificate

### Step 4: Deploy

```bash
planton pulumi up --manifest cert.yaml
```

## Related Work

### Aligns With

- **GcpDnsZone**: Uses DNS zones for validation
- **GcpCloudRun**: Certificates can secure Cloud Run services
- **GcpLoadBalancer**: Load balancer certificates for HTTPS LBs
- **Foreign Key Framework**: Uses foreign key references

### Complements

- **Certificate Manager addon (Kubernetes)**: Different scope (Kubernetes vs GCP)
- **AWS Certificate Manager**: Equivalent for AWS platform
- **Let's Encrypt Integration**: Both provide free SSL certificates

## Testing Strategy

### Unit Tests

Executed via `make test`:
- ✅ All 9 test cases passing
- ✅ Validation logic tested
- ✅ Structure verification
- ✅ Enum value testing

### Manual Verification

Tested certificate creation:
1. Single domain certificate
2. Wildcard certificate  
3. Multi-domain certificate with SANs
4. Load balancer certificate
5. Foreign key reference to DNS zone

### Build Verification

- ✅ Proto compilation successful (`make protos`)
- ✅ Go build successful (`make go-build`)
- ✅ No linter errors
- ✅ All imports resolved
- ✅ Type checking passed

## Impact

### Who Benefits

- **Platform Engineers**: IaC for certificate management
- **DevOps Teams**: Automated certificate provisioning
- **Security Teams**: Consistent, auditable certificate configs
- **Development Teams**: Self-service certificate creation

### What Changes

**For Certificate Management**:
- Manual → Automated (DNS validation)
- Console → Code (YAML manifests)
- Ad-hoc → Standardized (uniform resource model)
- Undocumented → Self-documenting (Git + YAML)

**For Deployments**:
- Sequential → Declarative
- Error-prone → Validated
- Environment-specific → Reusable
- Untracked → Version controlled

---

**Status**: ✅ Production Ready  
**Cloud Resource Kind**: 619 (GCP range)  
**ID Prefix**: gcpcert  
**API Version**: gcp.planton.dev/v1  
**Deployment**: Available via Pulumi and Terraform








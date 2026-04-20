# Pulumi Module Overview

The `GcpCertManagerCert` Pulumi module provides an intelligent abstraction for provisioning SSL/TLS certificates on Google Cloud Platform. It bridges two distinct GCP certificate services—the modern **Certificate Manager** and the classic **Google-managed SSL certificates**—through a unified API, enabling developers to choose between cost optimization and advanced features without managing the underlying complexity.

By accepting a single declarative specification, the module automatically orchestrates multi-step DNS validation workflows, creates DNS authorization records, and provisions certificates in the appropriate GCP service based on requirements. When a Cloud DNS zone is provided (`cloud_dns_zone_id`), validation records are auto-created in that zone. When omitted, the module still creates DNS authorizations but exports the required validation records as stack outputs (`dns-validation-records`) so they can be inserted manually into any DNS provider (e.g. AWS Route 53, Cloudflare). This transforms what would typically require 3-5 separate resources and explicit dependency management into a single, atomic operation.

## Key Capabilities

### Dual Certificate Type Support

The module intelligently provisions certificates based on the `certificate_type` field:

- **MANAGED (default)**: Creates certificates using Google Certificate Manager, supporting wildcard domains, DNS-based validation, regional scoping, and certificate maps for scale
- **LOAD_BALANCER**: Creates classic Google-managed SSL certificates optimized for load balancers, offering zero-cost provisioning but limited to non-wildcard, global certificates

This dual-mode design allows cost optimization for simple use cases (free classic certificates) while providing production-grade features (Certificate Manager) when needed.

### Automated DNS Validation Orchestration

For `MANAGED` certificates, the module automates the complex DNS validation workflow:

1. **DNS Authorization Creation**: Creates a `google_certificate_manager_dns_authorization` resource for each domain
2. **CNAME Record Extraction**: Extracts the unique CNAME validation record from the authorization response
3. **DNS Record Provisioning** (conditional): When `cloud_dns_zone_id` is provided, creates the CNAME record in the specified Cloud DNS zone. When omitted, this step is skipped.
4. **Validation Record Export**: Always exports all validation records as the `dns-validation-records` stack output, enabling manual insertion into any DNS provider
5. **Dependency Management**: Ensures DNS records (if created) exist before certificate creation using explicit `DependsOn` relationships
6. **Certificate Provisioning**: Creates the certificate with proper references to all DNS authorizations

This orchestration eliminates the "chicken-and-egg" validation problem and enables **zero-downtime certificate provisioning** independent of load balancer availability. The optional Cloud DNS zone support makes it possible to use GCP Certificate Manager even when the DNS zone is hosted outside GCP (e.g. AWS Route 53).

### Multi-Domain Certificate Support

The module supports both single-domain and multi-domain (Subject Alternative Names) certificates:

- Primary domain (`primary_domain_name`) serves as the main certificate identity
- Alternate domains (`alternate_domain_names`) are added as SANs
- Wildcard domains (`*.example.com`) are fully supported with Certificate Manager
- All domains are validated through a unified DNS authorization workflow

## Module Architecture

### Entry Point: `Resources` Function

The `Resources` function (`main.go`) serves as the module's entry point, following OpenMCF's standard pattern:

```go
func Resources(ctx *pulumi.Context, stackInput *gcpcertv1.GcpCertManagerCertStackInput) error
```

**Workflow:**
1. Initializes local variables from stack input (`initializeLocals`)
2. Configures the GCP provider with credentials from stack input
3. Invokes the certificate provisioning logic (`certManagerCert`)
4. Returns any errors with wrapped context for debugging

### Core Logic: `certManagerCert` Function

The `certManagerCert` function (`cert_manager_cert.go`) implements the certificate provisioning logic:

**Decision Logic:**
1. Reads the `certificate_type` from the spec (defaults to `MANAGED`)
2. Aggregates all domains (primary + alternates) into a single list
3. Routes to the appropriate provisioning function based on type:
   - `createManagedCertificate` for Certificate Manager certificates
   - `createLoadBalancerCertificate` for classic Google-managed SSL certificates

### Certificate Manager Implementation: `createManagedCertificate`

This function provisions modern Certificate Manager certificates with DNS validation:

**Resource Creation Sequence:**
1. **For each domain**:
   - Create a `certificatemanager.DnsAuthorization` with unique name and labels
   - Extract the DNS validation record (name, type, data) from the authorization
   - Create a `dns.RecordSet` in Cloud DNS with the CNAME validation record
   
2. **Certificate provisioning**:
   - Create a `certificatemanager.Certificate` with all domains
   - Reference all DNS authorizations for validation
   - Apply explicit `DependsOn` to ensure DNS records exist before certificate creation
   
3. **Output exports**:
   - `certificate-id`: The GCP resource ID
   - `certificate-name`: The certificate name
   - `certificate-domain-name`: The primary domain
   - `certificate-status`: Current provisioning status

**Key Implementation Details:**
- Uses `ApplyT` transforms to safely extract DNS record data from authorization outputs
- Iterates over all domains to create separate DNS authorizations per domain
- Aggregates all authorization IDs into a single array for certificate reference
- Applies GCP labels for resource management and cost attribution

### Load Balancer Implementation: `createLoadBalancerCertificate`

This function provisions classic Google-managed SSL certificates:

**Resource Creation:**
- Creates a `compute.ManagedSslCertificate` with all domains
- Validation happens automatically when the certificate is attached to a load balancer
- No DNS records are created (load balancer-based validation)

**Limitations:**
- Does not support wildcard domains
- Global scope only (no regional certificates)
- Validation requires active load balancer with correct DNS routing

### Local Variables: `initializeLocals`

The `initializeLocals` function (`locals.go`) prepares local variables used throughout the module:

**GCP Labels:**
- Constructs standardized GCP labels from metadata:
  - `resource`: Always `true` (marks as OpenMCF-managed)
  - `resource-name`: The certificate name
  - `resource-kind`: `gcpcertmanagercert`
  - `resource-id`: The OpenMCF resource ID
  - `organization`: The organization ID
  - `environment`: The environment name

These labels enable consistent resource management, cost tracking, and compliance across all OpenMCF-managed resources.

## Resource Dependencies

The module manages complex resource dependencies to ensure correct provisioning order:

```
DNS Authorization (per domain)
         ↓
    DNS Record (CNAME validation)  ← only when cloud_dns_zone_id is set
         ↓
    Certificate (aggregates all authorizations)
```

Pulumi's automatic dependency tracking handles most relationships, but the module explicitly declares `DependsOn` for the certificate → DNS authorizations relationship to prevent race conditions.

## Outputs and Status Tracking

The module exports outputs to the Pulumi stack for integration with other infrastructure:

| Output | Description | Example Value |
|--------|-------------|---------------|
| `certificate-id` | GCP resource ID | `projects/.../certificates/my-cert` |
| `certificate-name` | Certificate name | `my-cert` |
| `certificate-domain-name` | Primary domain | `example.com` |
| `certificate-status` | Provisioning status | `PROVISIONING` or `ACTIVE` |
| `dns-validation-records` | JSON array of CNAME records for DNS validation | `[{"record_name":"...","record_type":"CNAME","record_data":"...","domain":"example.com"}]` |

The `dns-validation-records` output is always populated for `MANAGED` certificates. When `cloud_dns_zone_id` is provided, these records are also auto-created in Cloud DNS. When omitted, users must manually insert these CNAME records into their DNS provider (e.g. AWS Route 53, Cloudflare).

These outputs are mapped to the `stack_outputs.proto` structure and returned to the OpenMCF CLI for status reporting and integration with other resources.

## Error Handling

The module employs defensive error handling throughout:

- **Provider setup errors**: Wrapped with context about provider configuration
- **DNS authorization errors**: Wrapped with the specific domain that failed
- **DNS record errors**: Wrapped with domain and zone information
- **Certificate creation errors**: Wrapped with certificate type and domain list

All errors are returned with descriptive messages to aid troubleshooting and debugging.

## Integration with OpenMCF Ecosystem

The Pulumi module integrates seamlessly with OpenMCF's multi-cloud framework:

- **Stack Input**: Receives `GcpCertManagerCertStackInput` containing the resource definition and provider credentials
- **Protobuf Validation**: All input is validated against Protobuf schema before module execution
- **Standard Metadata**: Uses shared metadata structure (name, ID, org, env) for consistent resource management
- **Foreign Key Resolution**: Resolves Cloud DNS zone references through OpenMCF's foreign key system
- **Output Mapping**: Exports outputs in a structure that maps to `stack_outputs.proto` for CLI integration

## Production Considerations

### DNS Propagation

Certificate Manager validation requires DNS records to be globally propagated. The module creates DNS records synchronously, but validation may take 5-10 minutes depending on DNS TTL and global propagation time.

### Certificate Lifecycle

All Google-managed certificates (both types) are **automatically renewed** by GCP before expiration. No user intervention is required for renewal.

### Scale Limits

- **Classic certificates**: Maximum 15 per target proxy
- **Certificate Manager (direct attachment)**: Maximum 100 per proxy
- **Certificate Manager (with maps)**: Thousands of certificates per map

For deployments requiring more than 10-15 certificates, use Certificate Manager (`MANAGED` type) to future-proof against scale limits.

### Cost Optimization

The module defaults to `MANAGED` (Certificate Manager) for maximum features, but allows explicit selection of `LOAD_BALANCER` for cost-sensitive deployments that don't require wildcard support or advanced features.

## Next Steps

- See [README.md](./README.md) for usage instructions, prerequisites, and deployment commands
- See [examples.md](./examples.md) for complete example manifests covering various scenarios
- See the parent [docs/README.md](../../docs/README.md) for research and architectural background



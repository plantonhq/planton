# Overview

The AWS Route53 Zone API resource provides a production-ready, feature-complete interface for creating and managing DNS zones and records within Amazon Route53, AWS's scalable and highly available DNS web service. This resource abstracts the complexities of DNS management while exposing Route53's powerful traffic management and high availability features through a simple, declarative API.

## Why We Created This API Resource

DNS is foundational infrastructure. When DNS works, nobody notices. When it breaks, everything breaks. Manual DNS management through the AWS console leads to configuration drift, undocumented changes, and operational risk. Traditional Infrastructure-as-Code approaches require deep Route53 expertise and verbose configuration.

We created this API resource to:

- **Make DNS Infrastructure as Code the Default**: DNS configurations should be versioned, reviewed, tested, and deployed through automated pipelines.
- **Expose Route53's Power Through Simplicity**: Access advanced features (alias records, routing policies, health checks) without memorizing AWS-specific APIs.
- **Prevent Common Mistakes**: Built-in validation prevents DNS misconfigurations like CNAME at apex, missing trailing dots, and invalid routing policy combinations.
- **Enable Production-Grade Patterns**: Support blue/green deployments, global traffic management, automatic failover, and split-horizon DNS out of the box.

## Key Features

### Core DNS Management

#### Public and Private Hosted Zones

- **Public Zones**: Resolve globally on the internet for customer-facing services
- **Private Zones**: Resolve only within associated VPCs for internal microservices and split-horizon DNS
- **VPC Associations**: Associate private zones with multiple VPCs across regions for cross-VPC DNS resolution

#### Basic DNS Records

Full support for standard DNS record types:
- **A / AAAA Records**: Map domain names to IPv4/IPv6 addresses
- **CNAME Records**: Create aliases to other domain names
- **MX Records**: Configure email routing with priorities
- **TXT Records**: Support SPF, DKIM, DMARC, domain verification, and custom text records
- All standard record types from the `DnsRecordType` enum

### Route53 Advanced Features (The 80/20)

#### Alias Records - Route53's Killer Feature

- **Zone Apex Support**: Point your apex domain (`example.com`) to AWS resources without CNAME restrictions
- **Cost Savings**: Alias queries to AWS resources are free (vs paid CNAME queries)
- **Automatic Updates**: Alias records automatically track target resource IP changes
- **Supported Targets**: CloudFront distributions, Application/Network Load Balancers, S3 website endpoints, API Gateway, Global Accelerator
- **Health Evaluation**: Optional health check evaluation for automatic failover

#### Traffic Management Routing Policies

##### Weighted Routing
- **Use Case**: Blue/green deployments, canary releases, A/B testing
- **How**: Distribute traffic across multiple resources based on assigned weights (0-255)
- **Example**: Route 90% traffic to stable version, 10% to new version

##### Latency-Based Routing
- **Use Case**: Global applications serving users from multiple regions
- **How**: Route users to the AWS region with lowest network latency
- **Benefit**: Automatic performance optimization without manual geolocation mapping

##### Failover Routing
- **Use Case**: Active-passive disaster recovery, automatic incident response
- **How**: Route to primary endpoint; automatically fail over to secondary when health check fails
- **Requirements**: Health check on primary resource, separate PRIMARY and SECONDARY records

##### Geolocation Routing
- **Use Case**: GDPR compliance, localized content, geographic restrictions
- **How**: Route based on user's geographic location (continent, country, or US state)
- **Example**: EU users to EU data centers, US users to US data centers

### Production-Grade Operations

#### Health Checks and Monitoring

- **Health Check Integration**: Attach Route53 health checks to failover records for automatic traffic shifting
- **Endpoint Monitoring**: Monitor HTTP/HTTPS endpoints, TCP connections, or CloudWatch alarms
- **Cost-Effective**: First 50 health checks on AWS resources are free (~$0.50/month per additional check)

#### Query Logging

- **CloudWatch Integration**: Send DNS query logs to CloudWatch Logs for debugging and security monitoring
- **Use Cases**: 
  - Debug resolution issues
  - Detect anomalous query patterns (security)
  - Understand query volume for cost optimization
- **Warning**: High-traffic domains generate large log volumes; set retention policies

#### DNSSEC

- **Enhanced Security**: Cryptographic signatures prevent DNS spoofing and cache poisoning attacks
- **Requirement**: Additional configuration at domain registrar level
- **Trade-off**: Adds complexity; adoption is low due to limited client support

### Validation and Compliance

#### Automatic Validation

- **DNS Name Validation**: Regex validation ensures valid domain names (supports wildcards: `*.example.com`)
- **Record Type Constraints**: Enforces DNS specifications (e.g., prevents CNAME at zone apex)
- **Routing Policy Rules**: Validates routing policy requirements (set_identifier, health_check_id)
- **Weight Boundaries**: Ensures weighted routing weights are 0-255
- **Private Zone Rules**: Requires VPC associations for private zones

#### Smart Defaults

- **Default TTL**: 300 seconds (5 minutes) - balances caching efficiency with change propagation
- **Lower TTL for Routing Policies**: Recommended 60 seconds for records using traffic management
- **Trailing Dot Handling**: Route53 automatically appends trailing dot if not provided

## Benefits

### For Platform Teams

- **Declarative DNS**: Define desired state; Pulumi handles create/update/delete operations
- **Multi-Environment Support**: Same configuration for dev/staging/prod with environment-specific overrides
- **Change Preview**: See exactly what will change before applying (Pulumi preview/Terraform plan)
- **State Management**: Automatic drift detection - manual console changes are flagged

### For Application Teams

- **Self-Service DNS**: Developers request DNS records via YAML manifests, no AWS console access needed
- **Blue/Green Deployments**: Gradually shift traffic from old to new versions with weighted routing
- **Global Traffic Management**: Automatically route users to nearest region with latency-based routing
- **High Availability**: Configure automatic failover with health checks for disaster recovery

### For Security Teams

- **Version Controlled**: All DNS changes tracked in Git with approval workflows
- **Audit Trail**: Query logging provides complete DNS query history for security analysis
- **DNSSEC Support**: Optional cryptographic signing for enhanced DNS security
- **Least Privilege**: Teams can manage DNS without full AWS Route53 access

### For Operations Teams

- **Simplified Troubleshooting**: Query logs and alias health evaluation simplify debugging
- **Proactive Monitoring**: Health checks detect issues before they impact customers
- **Cost Optimization**: Alias records reduce query costs; TTL tuning reduces query volume
- **Reduced Operational Burden**: Infrastructure as code eliminates manual console work

## Common Use Cases

### Customer-Facing Applications

```yaml
# Apex domain aliased to CloudFront, API with failover, email configured
- Alias record: example.com → CloudFront distribution (free queries, apex domain support)
- Failover routing: api.example.com → Primary/Secondary with health checks
- MX records: Email routing with priorities
- TXT records: SPF, DKIM, DMARC for email security
```

### Global Multi-Region Services

```yaml
# Latency-based routing across three regions
- global.example.com → us-east-1 (192.0.2.1)
- global.example.com → eu-west-1 (198.51.100.1)
- global.example.com → ap-southeast-1 (203.0.113.1)
Route53 automatically directs users to lowest-latency endpoint
```

### Blue/Green and Canary Deployments

```yaml
# Weighted routing for gradual rollout
- app.example.com → 90% to stable version (blue)
- app.example.com → 10% to new version (green)
Gradually increase green weight as confidence grows
```

### Microservices with Split-Horizon DNS

```yaml
# Private zone for internal services
- isPrivate: true
- VPC associations: vpc-12345678, vpc-87654321
- db.internal.example.com → 10.0.1.100
- cache.internal.example.com → 10.0.2.200
Internal services resolve within VPCs, not accessible from internet
```

### Compliance and Geolocation Routing

```yaml
# EU users to EU data centers (GDPR compliance)
- www.example.com → Europe: 198.51.100.1 (geolocation: EU)
- www.example.com → North America: 192.0.2.1 (geolocation: NA)
- www.example.com → Default: 203.0.113.1 (rest of world)
```

## Getting Started

See [examples.md](examples.md) for complete YAML manifests covering:
- Basic public zones with A, CNAME, MX, TXT records
- Alias records to CloudFront, ALB, S3 website endpoints
- Weighted routing for blue/green deployments
- Latency-based routing for global applications
- Failover routing with health checks for high availability
- Geolocation routing for compliance
- Private zones with VPC associations
- Query logging and DNSSEC configuration

For the theoretical foundation and tool comparison, see [docs/README.md](docs/README.md), which explains:
- The maturity spectrum from manual console to Infrastructure-as-Code
- Why Pulumi vs Terraform vs CloudFormation vs CDK
- The 80/20 rule: which Route53 features most teams actually need
- Production essentials and common mistakes to avoid

## Architecture Philosophy

This API resource implements the **80/20 principle** for Route53:
- 80% of production DNS use cases require only a small subset of Route53 features
- We expose those essential features with excellent ergonomics
- Advanced features (Traffic Flow policies, geoproximity routing) can be added as needed
- Simple by default, powerful when required

**Result**: Teams get production-grade DNS without Route53 expertise.

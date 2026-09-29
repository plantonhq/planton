# Production VPC Domain

This preset creates a production-grade OpenSearch domain with 3 data nodes across 3 Availability Zones, 3 dedicated master nodes, VPC deployment, fine-grained access control with internal user database, and enforced HTTPS with TLS 1.2+. This is the recommended starting point for production search and analytics workloads.

## When to Use

- Production search backends (product catalogs, document search, autocomplete)
- Application monitoring and observability pipelines
- Any workload requiring high availability, encryption, and network isolation
- Environments with compliance requirements (SOC 2, HIPAA, PCI-DSS)

## Key Configuration Choices

- **3 data nodes (r6g.large.search)** — Memory-optimized Graviton2 instances; 16 GiB RAM per node for effective filesystem caching. Scale instance type up for heavier workloads.
- **3 dedicated masters (r6g.large.search)** — Isolates cluster management from data operations for stability under load
- **Zone awareness (3 AZs)** — Distributes data and replicas across 3 Availability Zones; survives single-AZ failures
- **gp3 100 GB** — Predictable performance with baseline 3000 IOPS; increase `volumeSize` based on data volume
- **VPC with valueFrom references** — Subnet and security group IDs pulled from other Planton resources for full infrastructure composability
- **FGAC with internal user database** — Fine-grained access control with username/password authentication; add roles in OpenSearch Dashboards
- **Enforce HTTPS + TLS 1.2** — Modern encryption standards for all client-to-domain traffic
- **Auto-Tune enabled** — Automatic JVM and performance optimization

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `prod-search` | Domain name (3-28 chars, lowercase, hyphens) | Your naming convention |
| `<private-subnet-1-name>` / `<private-subnet-2-name>` / `<private-subnet-3-name>` | Names of the three private AwsSubnet resources (one per AZ) the domain spans | Your Planton subnet manifests |
| `<security-group-name>` | Name of the AwsSecurityGroup allowing HTTPS (443) | Your Planton security group manifest |
| `$secret/opensearch-master-password` | A managed-secret reference to the master user password (min 8 chars, mixed case, digit, special) — the field is sensitive, so a `$secret/<slug>` reference belongs here, never plaintext | Your org's managed secrets |

## Related Presets

- **01-single-node-dev** — Use for development and testing with minimal cost
- **03-analytics-warm-cold** — Use for analytics workloads with warm/cold storage tiers

# AWS DynamoDB Deployment: From Console Clicks to Production State Management

## Introduction

"DynamoDB is easy—just click create in the console." This statement is technically true but strategically misleading. While Amazon DynamoDB's serverless nature eliminates the operational burden of provisioning and managing servers, its distributed architecture demands rigorous design discipline. The ease of creation obscures the permanence of fundamental decisions: choose the wrong partition key and you've locked yourself into a throttled, unscalable table. Select default capacity settings without understanding their implications and you face either immediate throttling or runaway costs.

The paradox of DynamoDB is that it's simultaneously the simplest database to *create* and one of the most unforgiving databases to *design correctly*. A relational database allows you to add indexes and optimize queries after the fact. DynamoDB's key schema is immutable—there is no ALTER TABLE to fix a poor partition key design. The only remedy is data migration to a new table.

This architectural reality makes infrastructure automation not just convenient but essential. Manual provisioning through the AWS Console encourages a dangerous "create-then-configure" workflow that is fundamentally incompatible with DynamoDB's "design-for-access-patterns" requirement. Production-grade DynamoDB deployment demands declarative infrastructure as code that codifies design decisions, enforces guardrails, and makes the implicit explicit.

This document explores the full spectrum of DynamoDB deployment methods—from manual console operations to sophisticated declarative automation—and examines how Planton abstracts the complexity of the AWS API into a production-ready, protobuf-defined specification.

## The DynamoDB Deployment Spectrum

### Level 0: Manual Console Provisioning (The Learning Trap)

The AWS Management Console is where nearly every developer begins. The wizard-driven interface makes the *act* of creation trivially simple, but it systematically hides the *consequences* of configuration choices behind friendly dropdowns and default values.

**Common Pitfalls That Create Technical Debt**:

**Improper Key Schema** (The Fatal Error): A developer unfamiliar with NoSQL principles might intuitively select a low-cardinality attribute like `status` or `date` as the partition key. This design flaw creates "hot partitions"—a single physical partition receives a disproportionate volume of traffic, leading to throttling that cannot be solved by increasing capacity. The table fails to scale regardless of its provisioned throughput. This mistake is **permanent and irreversible** without migrating data to a new table.

**Wrong Billing Mode**: The console defaults to showing both billing modes with equal prominence. New users often accept the provisioned capacity defaults (historically 1-5 RCUs/WCUs), creating tables that throttle immediately under any real load. Conversely, developers "playing it safe" by over-provisioning capacity waste significant budget on unused resources.

**Missing Indexes**: Tables are frequently created to serve a single access pattern (e.g., `GetItem` by `UserID`). When a secondary pattern emerges ("find all users by email"), the lack of a Global Secondary Index forces the application into expensive, unscalable `Scan` operations that read the entire table.

**Neglected Production Safeguards**: Point-in-Time Recovery (PITR) is a single checkbox, not enabled by default. Without it, the table is vulnerable to operational disasters like accidental deletions or data corruption with no recovery path. Similarly, encryption defaults to AWS-owned keys rather than customer-managed keys (CMKs), sacrificing auditability and control.

**Verdict**: Acceptable for learning DynamoDB concepts and prototyping queries. Categorically unsuitable for any production workload or environment requiring reproducibility.

### Level 1: Scripting with AWS CLI and SDKs (Task Automation, Not State Management)

The AWS CLI and SDKs (Boto3, AWS SDK for Go, AWS SDK for Java) enable programmatic interaction with DynamoDB, but they are fundamentally **imperative tools** designed for runtime application logic, not infrastructure provisioning.

**AWS CLI**: Convenient for administrative one-off commands and simple shell scripts. The `aws dynamodb create-table` command can provision a table, but managing complex schemas requires mastering difficult `--attribute-definitions` and `--key-schema` JSON structures. There is no built-in concept of "desired state"—running the same command twice results in an error, not convergence.

**AWS SDKs**: Superior to the CLI for any programmatic interaction. These are the tools applications use to interact with DynamoDB's data plane (`PutItem`, `Query`, `Scan`) and control plane (`CreateTable`, `UpdateTable`). They are imperative—they execute a series of commands without drift detection, state tracking, or idempotency guarantees.

**Verdict**: Excellent for application code and CI/CD helper scripts. Not suitable for declarative infrastructure as code.

### Level 2: Configuration Management with Ansible (Bridging Imperative and Declarative)

Ansible represents a middle ground: declarative YAML playbooks defining a desired state, executed through imperative modules.

The `community.aws.dynamodb_table` module can create, update, and delete tables. It supports defining key schema, provisioned capacity, and secondary indexes. However, this approach highlights the challenges of using general-purpose configuration management tools for cloud provisioning.

**The Lag Problem**: Community-maintained modules can fall behind the cloud provider's API evolution. Historical documentation shows examples where Ansible modules lacked support for on-demand billing (`PAY_PER_REQUEST`) despite it being a core AWS feature, creating confusion about what is actually supported.

**Verdict**: Viable for teams already standardized on Ansible. Less robust than dedicated IaC tools for production cloud infrastructure.

### Level 3: Declarative Infrastructure as Code (The Production Standard)

This category represents the modern standard for production environments: tools that define desired end-state in code and calculate the necessary changes to achieve convergence.

#### AWS-Native: CloudFormation and AWS CDK

**AWS CloudFormation**: The foundational IaC service for AWS, using YAML or JSON templates. Its primary benefit is fully managed state—AWS maintains the stack state internally, eliminating external state file management. Its primary drawback is extreme verbosity. Defining a DynamoDB table with Global Secondary Indexes, autoscaling, and encryption can span hundreds of lines of template code.

**AWS CDK (Cloud Development Kit)**: A high-level abstraction allowing infrastructure definition in familiar programming languages (TypeScript, Python, Java). The CDK synthesizes imperative code into CloudFormation templates.

**The Game-Changer**: CDK's **L2 Constructs** (e.g., `TableV2`) provide remarkable simplification. Instead of manually defining `AttributeDefinitions` and `KeySchema` arrays, developers specify simple `partitionKey` and `sortKey` objects. The CDK *infers* the correct AWS API structures. For autoscaling, a single `billing.autoscaled()` configuration replaces the manual orchestration of three separate resources (Table, ApplicationAutoScaling Target, and two Scaling Policies).

**Trade-off**: AWS-only. Teams with multi-cloud requirements cannot use CDK.

#### Cloud-Agnostic: Terraform, Pulumi, and OpenTofu

**Terraform and OpenTofu**: The industry-dominant declarative tool using HashiCorp Configuration Language (HCL). Terraform is cloud-agnostic and manages state via an explicit state file (typically stored in S3 with DynamoDB for locking—a meta pattern where DynamoDB secures Terraform state for DynamoDB).

OpenTofu is an open-source, community-driven fork of Terraform created after HashiCorp's license change, maintaining compatibility while ensuring permanent open-source status.

**Key Characteristic**: Low-level, 1:1 mapping to the AWS API. The `aws_dynamodb_table` resource exposes the raw `attribute` blocks and `key_schema` that mirror AWS's API structure. This provides complete control but inherits AWS's complexity—developers must understand the subtle distinction that `attribute` definitions are only for keys, not all item attributes (a common source of confusion and infinite planning loops).

**Pulumi**: A direct Terraform competitor using general-purpose programming languages (Python, TypeScript, Go) for declarative infrastructure. It can leverage Terraform's provider ecosystem, inheriting the same resource schemas. The primary advantage is developer familiarity—using real code, unit tests, and familiar tooling instead of learning HCL.

**State Management**: Both tools manage explicit state files. In production, this **must** be remote (e.g., S3 backend with DynamoDB locking) to prevent corruption from concurrent runs and enable team collaboration.

**Verdict**: The gold standard for production, multi-cloud infrastructure as code. This is the foundation upon which Planton builds its abstractions.

#### Kubernetes-Native: Crossplane (Control Plane as Infrastructure)

Crossplane extends the Kubernetes API to manage external cloud resources. Instead of an external IaC tool, a control plane runs *within* the Kubernetes cluster, continuously reconciling Kubernetes Custom Resources with the AWS API.

A `Table` custom resource (e.g., `Table.dynamodb.aws.crossplane.io/v1alpha1`) maps 1:1 to the AWS API, with fields like `forProvider.attributeDefinitions` and `forProvider.billingMode`.

**The Composition Model**: Crossplane's power lies in its ability to create higher-level abstractions. A team can define a simplified `XDynamoDBTable` composition that wraps the underlying `Table` resource with opinionated defaults (on-demand billing, PITR enabled, encryption with CMK). This pattern is directly analogous to Planton's philosophy of abstraction.

**Trade-off**: Adds Kubernetes as a dependency for infrastructure management. Excellent for Kubernetes-centric organizations; additional complexity for teams not already running Kubernetes for control plane purposes.

## Comparative Analysis of Production IaC Tools

For production workloads, the choice narrows to four primary contenders. The decision hinges on trade-offs between language, state management, and abstraction level.

### The State Management Divide

**Terraform/Pulumi**: External state file management. Provides complete control and visibility but requires operational discipline:
- State files **must** be stored remotely (S3/DynamoDB for Terraform, Pulumi Service or S3 for Pulumi)
- State locking is essential to prevent corruption
- Lost or corrupted state is a critical failure scenario

**CloudFormation/CDK**: State managed implicitly by the AWS CloudFormation service. Zero operational overhead, extremely robust, but opaque. When stacks enter failed states (`UPDATE_ROLLBACK_FAILED`), recovery can be complex. Importing existing resources is notoriously painful.

### The Abstraction Spectrum

**Low-Level (Terraform/Pulumi/CloudFormation)**: 1:1 mapping to AWS API
- Complete control and flexibility
- Requires deep understanding of AWS API semantics
- Common confusion: `attribute_definitions` only for key attributes, not entire schema

**High-Level (AWS CDK)**: Opinionated constructs that hide complexity
- `partitionKey` and `sortKey` objects instead of manual `AttributeDefinitions` and `KeySchema`
- Automatic inference of correct AWS structures
- Autoscaling handled by single configuration instead of orchestrating three resources
- AWS-only limitation

### Language Philosophy

**HCL (Terraform/OpenTofu)**: Domain-Specific Language optimized for infrastructure
- Highly readable for pure infrastructure
- Limited logic capabilities (loops, conditionals are cumbersome)
- Cannot leverage general-purpose testing frameworks

**General-Purpose Languages (Pulumi/CDK)**: Python, TypeScript, Go, Java
- Familiar tools, unit testing, IDE support
- Full programming constructs (functions, loops, conditionals)
- Can abstract complex patterns into reusable libraries
- Risk: Developers may over-engineer infrastructure code

## DynamoDB Production Essentials

Understanding DynamoDB deployment tooling is only half the battle. Successful production deployments require mastery of core DynamoDB design principles.

### The Cornerstone: Key Schema Design

This is the single most critical decision in a DynamoDB table's lifecycle. The key schema dictates physical data distribution and is **immutable after creation**.

**Partition Key (HASH)**: The value is hashed to determine the physical partition for storage. For uniform load distribution, this key **must** be high-cardinality (unique or near-unique values). Examples: `UserID`, `OrderID`, `DeviceID`.

**The Cardinal Sin**: Choosing a low-cardinality partition key (e.g., `status` with values "active"/"inactive"). This creates hot partitions where a disproportionate volume of requests targets a single physical partition. Each partition has a hard throughput limit (1,000 WCUs/second, 3,000 RCUs/second). A hot partition will throttle requests *even if the table's total capacity is vastly higher*.

**Sort Key (RANGE)**: Optional. Organizes items within a single partition, enabling range queries using operators like `begins_with`, `between`, and comparison operators. Example: Partition key `UserID`, sort key `OrderDate` enables "Get all orders for user X between dates Y and Z."

**Recovery from Mistakes**: None. The only path is data migration to a new table with a corrected schema.

### Billing Modes: The Core Financial Decision

DynamoDB offers two billing modes. You can switch between them once every 24 hours.

#### On-Demand (PAY_PER_REQUEST)

**How It Works**: Pay a fixed price per read/write request. No capacity planning required. The table automatically scales to accommodate traffic up to double the previous peak.

**When to Use**: This is the **recommended default** for most workloads, especially:
- New applications with unpredictable traffic
- Spiky or sporadic workloads
- Development and testing environments

**Cost Reality**: Following a 50% price reduction in November 2024, on-demand is economically competitive with provisioned mode for a much wider range of workloads. The historical wisdom that "on-demand is vastly more expensive" is largely obsolete.

#### Provisioned Mode

**How It Works**: Specify (and pay for) a fixed number of Read Capacity Units (RCUs) and Write Capacity Units (WCUs) per hour.

**When to Use**: High-volume, stable, and predictable workloads where capacity can be accurately forecasted.

**Critical Requirement**: In production, this mode **must** be coupled with Application Auto Scaling. Manually managing RCUs/WCUs is operationally fragile. Auto Scaling automatically adjusts capacity in response to load, optimizing both performance and cost.

**Financial Optimization**: Provisioned mode's primary benefit is realized through Reserved Capacity purchases (1-year or 3-year commitments with significant discounts).

**Modern Strategy**: Start with On-Demand. Monitor usage for 1-2 months. Only migrate to Provisioned + Autoscaling + Reserved Capacity if usage patterns prove stable and predictable enough to justify the operational complexity.

### Secondary Indexes: Query Flexibility

DynamoDB's primary key supports two query patterns: exact match (`GetItem`) and range query (if a sort key exists). All other access patterns require secondary indexes. Using `Scan` operations (reading the entire table) is an anti-pattern that does not scale.

#### Global Secondary Index (GSI)

**Schema**: Can have a *different* partition key and sort key from the base table. Enables entirely new query patterns.

**Capacity**: Has its *own* read/write capacity, separate from the base table. In on-demand mode, this is automatic. In provisioned mode, each GSI requires explicit RCU/WCU configuration (and potentially separate autoscaling).

**Consistency**: Offers **eventual consistency only**. Updates are replicated from the table to the GSI asynchronously (typically milliseconds).

**Lifecycle**: Can be created, modified, or deleted after table creation (though GSI creation on large existing tables is a long-running operation).

**Cost Implication**: A GSI is effectively a second table. Writes to the base table that modify indexed attributes consume WCUs on both the base table and the GSI. Storage cost depends on projection type.

#### Local Secondary Index (LSI)

**Schema**: *Must* use the same partition key as the base table but a different sort key.

**Capacity**: *Shares* the base table's capacity.

**Consistency**: Can be queried with **strong consistency** (unlike GSIs).

**Lifecycle**: Can **only** be created at table creation time. Cannot be added later.

**Critical Limitation**: All items for a single partition key (across the base table and all LSIs) cannot exceed 10 GB total size. This is a severe scaling constraint that makes LSIs a niche feature.

**Verdict**: GSIs are the 99% solution. LSIs should only be used when strong consistency on a secondary query pattern is an absolute requirement and item collection size is guaranteed to remain small.

#### Projection Types: The Storage/Performance Trade-off

Projections define which attributes are copied from the base table into the index.

**KEYS_ONLY**: Projects only the index keys and the table's primary key. Smallest and cheapest, but querying for non-key attributes requires a second "fetch" read from the base table (consuming additional RCUs).

**INCLUDE**: Projects keys plus a specified list of non-key attributes. Often the optimal balance—include only the attributes needed by index queries.

**ALL**: Projects all attributes. Simplest to use (no fetch reads), but effectively **doubles storage cost** and doubles write cost (every item write triggers a full write to the GSI).

### Resilience: Backup and Recovery

**Point-in-Time Recovery (PITR)**: Enables continuous backups, allowing restoration to any single second within the preceding 35 days. Essential for operational recovery from accidental deletions, bad deployments, or data corruption. The cost is minimal (~20% of table storage cost). **This should be enabled by default for production tables.**

**On-Demand Backups**: Manual, full snapshots. Used for long-term archival (e.g., year-end compliance snapshots) or pre-deployment safety snapshots. Not a replacement for PITR.

**Global Tables**: Multi-region, active-active replication for disaster recovery and low-latency global applications. Complex and expensive; reserved for specific high-availability requirements.

### Security: Encryption at Rest

Encryption in transit is automatic (HTTPS). Encryption at rest is **always enabled** and cannot be disabled. The only configurable choice is the encryption key type:

1. **AWS Owned Key** (Default): Fully transparent, no cost, no management. Suitable for most applications.
2. **AWS Managed Key (aws/dynamodb)**: An AWS-managed KMS key. Incurs KMS API costs.
3. **Customer Managed Key (CMK)**: Full control over key lifecycle, rotation, and access policies. Required for strict compliance mandates. Highest cost and operational complexity.

**Recommendation**: Default to AWS Owned Key. Use CMK only when compliance or audit requirements explicitly mandate customer-controlled encryption keys.

### Streams: Change Data Capture

DynamoDB Streams provides a time-ordered log of item-level changes (INSERT, MODIFY, REMOVE). When enabled, every modification publishes an event to the stream.

**Primary Use Case**: Event-driven architecture. The stream triggers AWS Lambda functions for asynchronous processing:
- Replicate changes to search indexes (OpenSearch, Algolia)
- Invalidate application caches
- Update aggregate tables or materialized views
- Cross-region data synchronization

**View Types**: Control what data is included in stream records:
- `KEYS_ONLY`: Only key attributes
- `NEW_IMAGE`: The entire item after the change
- `OLD_IMAGE`: The entire item before the change
- `NEW_AND_OLD_IMAGES`: Both before and after states

### Observability: Monitoring and Debugging

**Amazon CloudWatch Metrics**: DynamoDB automatically publishes detailed metrics:
- `ConsumedReadCapacityUnits` / `ConsumedWriteCapacityUnits`: Actual usage
- `ThrottledRequests`: The critical alarm metric—indicates capacity exhaustion

**CloudWatch Contributor Insights**: The essential diagnostic tool for hot key detection. When `ThrottledRequests` spike, enable Contributor Insights to identify the top-N most-accessed partition keys. This pinpoints the specific items causing hot partitions.

**AWS X-Ray**: End-to-end distributed tracing. Trace a single request from API Gateway → Lambda → DynamoDB and back, identifying latency bottlenecks at each hop.

### Common Anti-Patterns to Avoid

**Design Anti-Patterns**:
- Hot partitions from low-cardinality partition keys
- Choosing provisioned capacity with default (5) RCUs/WCUs
- Over-provisioning capacity "just to be safe" instead of using on-demand or autoscaling

**Operational Anti-Patterns**:
- Using `Scan` operations at scale (reads every item in the table)
- Not enabling PITR on production tables
- Storing unbounded data without TTL (Time-to-Live) configuration
- Creating GSIs with `PROJECTION_TYPE: ALL` by default (doubles costs)

**Security Anti-Patterns**:
- Overly permissive IAM policies (e.g., `dynamodb:*` on all tables)
- Not using separate IAM roles for different application components
- Storing sensitive data without field-level encryption

## What Planton Supports

Planton provides a Kubernetes-style, protobuf-defined API for deploying AWS DynamoDB tables. The design philosophy balances completeness with usability.

### API Design: Explicit Over Implicit

Unlike higher-level abstractions (like AWS CDK's L2 constructs), Planton's `AwsDynamodbSpec` maintains a close mapping to the AWS API surface. This is a deliberate architectural choice prioritizing:

1. **Completeness**: Full access to DynamoDB features without "abstraction gaps"
2. **Predictability**: Explicit field names map directly to AWS documentation
3. **Migration Path**: Teams familiar with Terraform/CloudFormation can translate existing configurations with minimal cognitive overhead

### Core Specification

The `AwsDynamodbSpec` protobuf message includes:

**Essential Configuration**:
- `region`: The AWS region where the resource will be created (required)
- `billing_mode`: Enum for `PROVISIONED` or `PAY_PER_REQUEST`
- `attribute_definitions`: Explicit definition of key attributes (name + type: S/N/B)
- `key_schema`: Primary key structure (HASH and optional RANGE)
- `point_in_time_recovery_enabled`: Boolean (defaults should be `true` in practice)
- `deletion_protection_enabled`: Boolean safeguard for production

**Advanced Configuration**:
- `global_secondary_indexes`: Array of GSI definitions (name, key_schema, projection, optional throughput)
- `local_secondary_indexes`: Array of LSI definitions (rare; must be created with table)
- `provisioned_throughput`: RCUs/WCUs for provisioned billing mode
- `server_side_encryption`: Encryption configuration (enabled flag + optional KMS key ARN)
- `ttl`: Time-to-live configuration (enabled flag + attribute name)
- `stream_enabled` + `stream_view_type`: DynamoDB Streams configuration
- `table_class`: `STANDARD` or `STANDARD_INFREQUENT_ACCESS` (storage cost optimization)
- `contributor_insights_enabled`: Hot key detection

### Validation Guardrails

The protobuf specification includes comprehensive validations using `buf` validators:

**Structural Validations**:
- Key schema must have exactly one HASH key and at most one RANGE key
- GSI key schema must follow the same rule
- LSI must have exactly one HASH and one RANGE key (2 elements total)

**Consistency Validations**:
- If `billing_mode` is `PROVISIONED`, `provisioned_throughput` must be set with RCUs/WCUs > 0
- For `PAY_PER_REQUEST`, throughput must be unset or zero
- Each GSI's throughput must match the table's billing mode
- If `stream_enabled` is true, `stream_view_type` must be set
- If TTL is enabled, `attribute_name` must be specified

**Projection Validations**:
- If projection type is `INCLUDE`, `non_key_attributes` array must be non-empty
- Otherwise, it must be empty

These validations prevent common configuration errors at API submission time, before they reach AWS.

### Multi-Environment Pattern

Following AWS best practices, Planton deployments should use distinct AWS accounts or at minimum separate IAM roles for each environment:
- Development: Lower-cost instance sizes, on-demand billing, PITR optional
- Staging: Production-equivalent sizing, on-demand billing, PITR enabled
- Production: On-demand or provisioned with autoscaling, PITR enabled, deletion protection, CMK encryption

### Deployment Target

Planton's Pulumi-based implementation generates the AWS resources:
- DynamoDB table with specified configuration
- Optional Application Auto Scaling target and policies (for provisioned + autoscaling pattern)
- CloudWatch alarms for throttling detection (optional)
- IAM policies for application access (via separate resources)

### What's Intentionally Separate

Following the Unix philosophy of "do one thing well," certain related concerns are managed as separate resources:

**Network Security**: VPC endpoints for private DynamoDB access (managed at VPC level)

**Application Access**: IAM policies granting specific application roles read/write permissions (managed as separate IAM resources)

**Data Loading**: Initial data population or migration (handled by application-level tooling, not IaC)

This separation maintains clean boundaries and prevents the table resource from becoming a monolithic "does everything" abstraction.

## Cost Optimization Strategies

DynamoDB costs span compute (capacity), storage, backups, and data transfer.

### Billing Mode Selection: The Primary Cost Lever

**Strategy**: Default to `PAY_PER_REQUEST` for all new tables. Monitor `ConsumedReadCapacityUnits` and `ConsumedWriteCapacityUnits` for 1-2 months. If usage proves stable and predictable, evaluate migration to `PROVISIONED` with autoscaling and potentially Reserved Capacity purchases.

**Break-Even Analysis**: With the November 2024 price reduction, on-demand pricing became competitive with provisioned for a much wider range of workloads. Provisioned mode is only cost-effective when combined with Reserved Capacity (1-3 year commitments).

### Storage Optimization

**Table Class**: For tables with infrequent access patterns (e.g., archived orders, audit logs), use `STANDARD_INFREQUENT_ACCESS` table class. This offers up to 60% lower storage costs in exchange for higher read/write request costs. This is a day-two optimization, not a day-one decision.

**TTL (Time-to-Live)**: Automatically delete expired items based on a timestamp attribute. The deletion operation is **free** (does not consume WCUs). Essential for transient data like sessions, temporary tokens, or logs. Reduces storage costs and improves performance by keeping tables lean.

### Index Cost Management

**Be Frugal with GSIs**: Each GSI is effectively a second table with separate capacity and storage costs. Only create indexes for well-defined access patterns.

**Projection Optimization**: Strongly prefer `KEYS_ONLY` or `INCLUDE` over `ALL`. Projecting all attributes doubles storage cost and doubles write cost. Only project the specific attributes needed for index queries.

### Monitoring-Driven Optimization

**CloudWatch Contributor Insights**: When enabled, identifies the specific partition keys consuming the most capacity. Use this data to detect hot keys, over-provisioned indexes, or unexpected access patterns that can be optimized.

**Cost Allocation Tags**: Tag tables with environment, team, and application metadata. Use AWS Cost Explorer to attribute costs and identify optimization opportunities across the organization.

## Conclusion: Design First, Automate Second

The DynamoDB deployment landscape has matured from manual console operations to sophisticated, declarative infrastructure as code. Yet tooling sophistication cannot compensate for poor schema design. The cardinal rule remains: **design for access patterns first, automate second**.

The tooling hierarchy is clear:
- **Learning**: AWS Console for experimentation and prototyping
- **Scripting**: AWS CLI/SDKs for administrative tasks and application runtime
- **Production**: Terraform/OpenTofu/Pulumi for declarative, state-managed infrastructure
- **AWS-Centric Teams**: AWS CDK for superior developer experience with automatic CloudFormation state management
- **Kubernetes-Native Teams**: Crossplane for control-plane-driven reconciliation

Planton builds upon this foundation, positioning the `AwsDynamodb` API as a complete, explicit specification that maps closely to AWS's API surface while adding critical validation guardrails. The protobuf-defined schema prevents common misconfiguration at submission time, and the Pulumi-based implementation ensures idempotent, convergent deployments.

The paradigm shift is architectural: DynamoDB's serverless model eliminates operational database management; infrastructure as code eliminates deployment inconsistency. Combined with immutable key schema design principles—high-cardinality partition keys, sparse sort keys for range queries, intentional index design—teams can deploy DynamoDB tables with confidence that the infrastructure will scale to meet demand without the burden of server management.

Modern best practices are clear: default to on-demand billing, enable PITR for production tables, use GSIs judiciously with sparse projections, implement TTL for transient data, and monitor with Contributor Insights to detect hot partitions. These patterns, codified in infrastructure as code, transform DynamoDB from a service easy to misuse into a service engineered for production resilience.


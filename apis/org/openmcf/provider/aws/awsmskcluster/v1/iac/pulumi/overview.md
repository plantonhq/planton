# AwsMskCluster — Pulumi Module Architecture

This document describes the architecture of the Pulumi IaC module that provisions AWS MSK clusters from the `AwsMskClusterSpec`.

---

## File Structure

```
iac/pulumi/
├── Pulumi.yaml              # Pulumi project metadata
├── main.go                  # Entry point — loads stack input, calls module.Resources()
└── module/
    ├── main.go              # Orchestrator — creates resources in order, exports outputs
    ├── locals.go            # Locals struct — holds resolved AwsMskCluster + labels map
    ├── security_group.go    # Managed security group with ingress rules
    ├── configuration.go     # Inline MSK Configuration from server_properties
    ├── cluster.go           # MSK Cluster resource with all configuration blocks
    └── outputs.go           # Output key constants (15 total)
```

---

## Resource Creation Flow

The `module.Resources()` function orchestrates resource creation in a strict dependency order:

```
1. initializeLocals()          → Locals struct (labels, resolved target)
2. securityGroup()             → *ec2.SecurityGroup (conditional)
3. configuration()             → *msk.Configuration (conditional)
4. cluster()                   → *msk.Cluster (depends on SG + config)
5. ctx.Export(...)             → 15 stack outputs
```

### Step 1: Initialize Locals

`locals.go` creates a `Locals` struct containing:
- `AwsMskCluster` — the target resource from stack input.
- `Labels` — a standard tag map applied to all created resources.

### Step 2: Security Group (Conditional)

`security_group.go` creates a managed security group **only when** `securityGroupIds` or `allowedCidrBlocks` are provided in the spec.

**Created resources:**

| Resource | Condition | Description |
|----------|-----------|-------------|
| `ec2.SecurityGroup` ("cluster-sg") | `len(securityGroupIds) > 0 \|\| len(allowedCidrBlocks) > 0` | SG in the specified VPC |
| `ec2.SecurityGroupRule` ("ingress-kafka-sg-N") | Per source SG | TCP 9092-9098 from source SG |
| `ec2.SecurityGroupRule` ("ingress-zk-sg-N") | Per source SG | TCP 2181-2182 from source SG |
| `ec2.SecurityGroupRule` ("ingress-kafka-cidr") | `len(allowedCidrBlocks) > 0` | TCP 9092-9098 from CIDRs |
| `ec2.SecurityGroupRule` ("ingress-zk-cidr") | `len(allowedCidrBlocks) > 0` | TCP 2181-2182 from CIDRs |
| `ec2.SecurityGroupRule` ("egress-all") | Always (when SG created) | All outbound traffic |

**Port ranges:**
- Kafka: 9092 (plaintext), 9094 (TLS/mTLS), 9096 (SASL/SCRAM), 9098 (SASL/IAM)
- ZooKeeper: 2181 (plaintext), 2182 (TLS)

The function returns `nil` when no ingress references exist, and the cluster creation step skips adding it to the security group list.

### Step 3: MSK Configuration (Conditional)

`configuration.go` creates an inline MSK Configuration **only when** `serverProperties` is non-empty.

**Created resources:**

| Resource | Condition | Description |
|----------|-----------|-------------|
| `msk.Configuration` ("kafka-config") | `len(serverProperties) > 0` | Kafka server.properties overrides |

The configuration name follows the pattern `{metadata.id}-config`. Server properties are serialized to `key = value` format, sorted alphabetically for deterministic output.

When `serverProperties` is empty, the function returns `nil`. The cluster creation step then falls back to using `configurationArn` + `configurationRevision` if provided, or no configuration at all.

### Step 4: MSK Cluster

`cluster.go` is the main resource creation file. It constructs the `msk.ClusterArgs` by mapping every spec field to the Pulumi AWS provider's `msk.Cluster` arguments.

**Key mapping logic:**

| Spec Section | Cluster Argument | Notes |
|---|---|---|
| `subnetIds` | `BrokerNodeGroupInfo.ClientSubnets` | Resolved from `StringValueOrRef` |
| `associateSecurityGroupIds` + created SG | `BrokerNodeGroupInfo.SecurityGroups` | Combined list |
| `ebsVolumeSizeGib` / `provisionedThroughput*` | `BrokerNodeGroupInfo.StorageInfo` | Conditional block |
| `publicAccessType` | `BrokerNodeGroupInfo.ConnectivityInfo` | Conditional block |
| `storageMode` | `StorageMode` | Only set when non-empty |
| `kmsKeyArn` / `clientBrokerEncryption` / `inClusterEncryption` | `EncryptionInfo` | Conditional block |
| `authentication.*` | `ClientAuthentication` | SASL, TLS, unauthenticated |
| created config or `configurationArn` | `ConfigurationInfo` | Inline config takes precedence |
| `logging.*` | `LoggingInfo` | Delegated to `buildLogging()` |
| `enhancedMonitoring` | `EnhancedMonitoring` | Only set when non-empty |
| `jmxExporterEnabled` / `nodeExporterEnabled` | `OpenMonitoring` | Conditional Prometheus block |

The `buildLogging()` helper constructs the logging configuration from the three destination sub-messages (CloudWatch, Firehose, S3), each independently conditional.

### Step 5: Export Outputs

`main.go` exports 15 outputs using the constants defined in `outputs.go`:

| Output Key | Source | Conditional |
|---|---|---|
| `cluster_arn` | `mskCluster.Arn` | No |
| `cluster_name` | `mskCluster.ClusterName` | No |
| `cluster_uuid` | `mskCluster.ClusterUuid` | No |
| `current_version` | `mskCluster.CurrentVersion` | No |
| `bootstrap_brokers` | `mskCluster.BootstrapBrokers` | No |
| `bootstrap_brokers_tls` | `mskCluster.BootstrapBrokersTls` | No |
| `bootstrap_brokers_sasl_iam` | `mskCluster.BootstrapBrokersSaslIam` | No |
| `bootstrap_brokers_sasl_scram` | `mskCluster.BootstrapBrokersSaslScram` | No |
| `bootstrap_brokers_public_tls` | `mskCluster.BootstrapBrokersPublicTls` | No |
| `bootstrap_brokers_public_sasl_iam` | `mskCluster.BootstrapBrokersPublicSaslIam` | No |
| `bootstrap_brokers_public_sasl_scram` | `mskCluster.BootstrapBrokersPublicSaslScram` | No |
| `zookeeper_connect_string` | `mskCluster.ZookeeperConnectString` | No |
| `zookeeper_connect_string_tls` | `mskCluster.ZookeeperConnectStringTls` | No |
| `security_group_id` | `createdSg.ID()` | **Yes** — only when managed SG created |
| `configuration_arn` | `createdConfig.Arn` | **Yes** — only when inline config created |

The 13 cluster outputs are always exported (AWS returns empty strings for irrelevant endpoints). The 2 conditional outputs (`security_group_id`, `configuration_arn`) are only exported when their respective resources were created.

---

## AWS Provider Configuration

The entry point (`main.go`) loads stack input via `stackinput.LoadStackInput()` and passes it to `module.Resources()`.

The provider is configured in `module/main.go`:
- If `ProviderConfig` is nil → default AWS provider (ambient credentials).
- If `ProviderConfig` is set → explicit provider with `AccessKey`, `SecretKey`, `Region`, and optional `SessionToken`.

The provider instance is passed to every resource via `pulumi.Provider(provider)`.

---

## Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| `pulumi/pulumi/sdk` | v3 | Pulumi SDK |
| `pulumi/pulumi-aws/sdk` | v7 | AWS Classic provider |
| `pkg/errors` | — | Error wrapping |
| `openmcf/.../awsmskcluster/v1` | — | Generated protobuf types |
| `openmcf/pkg/iac/pulumi/pulumimodule/stackinput` | — | Stack input loader |

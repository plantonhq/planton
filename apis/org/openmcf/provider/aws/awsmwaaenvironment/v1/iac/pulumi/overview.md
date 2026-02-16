# AwsMwaaEnvironment — Pulumi Module Architecture

This document describes the architecture of the Pulumi IaC module that provisions AWS MWAA environments from the `AwsMwaaEnvironmentSpec`.

---

## File Structure

```
iac/pulumi/
├── Pulumi.yaml              # Pulumi project metadata
├── main.go                  # Entry point — loads stack input, calls module.Resources()
└── module/
    ├── main.go              # Orchestrator — creates resources in order, exports outputs
    ├── locals.go            # Locals struct — holds resolved AwsMwaaEnvironment + labels map
    ├── security_group.go    # Managed security group with ingress rules
    ├── environment.go       # MWAA Environment resource with all configuration blocks
    └── outputs.go           # Output key constants (8 total)
```

---

## Resource Creation Flow

The `module.Resources()` function orchestrates resource creation in a strict dependency order:

```
1. initializeLocals()          → Locals struct (labels, resolved target)
2. securityGroup()             → *ec2.SecurityGroup (conditional)
3. environment()               → *mwaa.Environment (depends on SG)
4. ctx.Export(...)             → 8 stack outputs
```

### Step 1: Initialize Locals

`locals.go` creates a `Locals` struct containing:
- `AwsMwaaEnvironment` — the target resource from stack input.
- `Labels` — a standard tag map applied to all created resources:
  - `planton.org/resource: "true"`
  - `planton.org/organization: {metadata.org}`
  - `planton.org/environment: {metadata.env}`
  - `planton.org/resource-kind: "AwsMwaaEnvironment"`
  - `planton.org/resource-id: {metadata.id}`

### Step 2: Security Group (Conditional)

`security_group.go` creates a managed security group **only when** `securityGroupIds` or `allowedCidrBlocks` are provided in the spec.

**Created resources:**

| Resource | Condition | Description |
|----------|-----------|-------------|
| `ec2.SecurityGroup` ("environment-sg") | `len(securityGroupIds) > 0 \|\| len(allowedCidrBlocks) > 0` | SG in the specified VPC |
| `ec2.SecurityGroupRule` ("ingress-self") | Always (when SG created) | Self-referencing: all traffic from itself |
| `ec2.SecurityGroupRule` ("ingress-https-sg-N") | Per source SG | TCP 443 from source SG |
| `ec2.SecurityGroupRule` ("ingress-https-cidr") | `len(allowedCidrBlocks) > 0` | TCP 443 from CIDRs |
| `ec2.SecurityGroupRule` ("egress-all") | Always (when SG created) | All outbound traffic |

**Self-referencing pattern:**

The self-referencing rule is the critical piece for MWAA networking. It creates an inbound rule where both `SourceSecurityGroupId` and `SecurityGroupId` point to the same security group. This allows all MWAA components (scheduler, workers, webserver, metadata DB) to communicate with each other through the VPC endpoints.

```go
ec2.NewSecurityGroupRule(ctx, "ingress-self", &ec2.SecurityGroupRuleArgs{
    Type:                  pulumi.String("ingress"),
    FromPort:              pulumi.Int(0),
    ToPort:                pulumi.Int(0),
    Protocol:              pulumi.String("-1"),    // all protocols
    SourceSecurityGroupId: sg.ID(),                // self
    SecurityGroupId:       sg.ID(),                // self
})
```

**Port assignment:**

Unlike MSK (which needs multiple port ranges), MWAA only requires:
- Port 443 (HTTPS) — for Airflow UI and REST API access from external clients.
- All ports (self-referencing) — for internal MWAA component communication.

The function returns `nil` when no ingress references exist, and the environment creation step skips adding it to the security group list.

### Step 3: MWAA Environment

`environment.go` is the main resource creation file. It constructs the `mwaa.EnvironmentArgs` by mapping every spec field to the Pulumi AWS provider's `mwaa.Environment` arguments.

**Security group assembly:**

The environment receives a combined security group list:
```go
sgIds := pulumi.StringArray{}
if createdSg != nil {
    sgIds = append(sgIds, createdSg.ID())    // managed SG
}
for _, sgOrRef := range spec.AssociateSecurityGroupIds {
    sgIds = append(sgIds, pulumi.String(sgOrRef.GetValue()))  // direct attachments
}
```

**Key mapping logic:**

| Spec Section | Environment Argument | Condition |
|---|---|---|
| `subnetIds` | `NetworkConfiguration.SubnetIds` | Always (required) |
| managed SG + `associateSecurityGroupIds` | `NetworkConfiguration.SecurityGroupIds` | Always (combined list) |
| `airflowVersion` | `AirflowVersion` | Non-empty |
| `airflowConfigurationOptions` | `AirflowConfigurationOptions` | Non-empty map |
| `sourceBucketArn` | `SourceBucketArn` | Always (required) |
| `dagS3Path` | `DagS3Path` | Always (required) |
| `pluginsS3Path` / `pluginsS3ObjectVersion` | `PluginsS3Path` / `PluginsS3ObjectVersion` | Non-empty |
| `requirementsS3Path` / `requirementsS3ObjectVersion` | `RequirementsS3Path` / `RequirementsS3ObjectVersion` | Non-empty |
| `startupScriptS3Path` / `startupScriptS3ObjectVersion` | `StartupScriptS3Path` / `StartupScriptS3ObjectVersion` | Non-empty |
| `executionRoleArn` | `ExecutionRoleArn` | Always (required) |
| `kmsKeyArn` | `KmsKey` | Non-nil |
| `environmentClass` | `EnvironmentClass` | Non-empty |
| `minWorkers` / `maxWorkers` | `MinWorkers` / `MaxWorkers` | > 0 |
| `minWebservers` / `maxWebservers` | `MinWebservers` / `MaxWebservers` | > 0 |
| `schedulers` | `Schedulers` | > 0 |
| `webserverAccessMode` | `WebserverAccessMode` | Non-nil (optional field) |
| `endpointManagement` | `EndpointManagement` | Non-empty |
| `loggingConfiguration` | `LoggingConfiguration` | Non-nil |
| `weeklyMaintenanceWindowStart` | `WeeklyMaintenanceWindowStart` | Non-empty |

**Note on `workerReplacementStrategy`:** This field is included in the spec but is **not yet available** in the `pulumi-aws` SDK v7. The Terraform module supports it. When the SDK is upgraded, the commented-out block in `environment.go` should be uncommented.

**Logging configuration builder:**

The logging configuration is built by `buildLoggingConfiguration()`, which delegates to 5 type-specific builder functions. Each Pulumi log module type (`DagProcessingLogs`, `SchedulerLogs`, `TaskLogs`, `WebserverLogs`, `WorkerLogs`) is a **distinct Go type** despite having identical fields (`Enabled`, `LogLevel`). This requires separate builder functions for type safety:

```go
buildLoggingModuleConfig()           → DagProcessingLogsPtrInput
buildSchedulerLoggingModuleConfig()  → SchedulerLogsPtrInput
buildTaskLoggingModuleConfig()       → TaskLogsPtrInput
buildWebserverLoggingModuleConfig()  → WebserverLogsPtrInput
buildWorkerLoggingModuleConfig()     → WorkerLogsPtrInput
```

Each builder:
1. Creates the type-specific args struct with `Enabled` set.
2. Conditionally sets `LogLevel` only when non-empty (allowing AWS to use the default `INFO`).

### Step 4: Export Outputs

`main.go` exports 8 outputs using the constants defined in `outputs.go`:

| Output Key | Source | Conditional |
|---|---|---|
| `environment_arn` | `env.Arn` | No |
| `environment_name` | `env.Name` | No |
| `webserver_url` | `env.WebserverUrl` | No |
| `airflow_version` | `env.AirflowVersion` | No |
| `service_role_arn` | `env.ServiceRoleArn` | No |
| `environment_class` | `env.EnvironmentClass` | No |
| `status` | `env.Status` | No |
| `security_group_id` | `createdSg.ID()` | **Yes** — only when managed SG created |

The 7 environment outputs are always exported. The `security_group_id` output is only exported when the managed security group was created (i.e., when `securityGroupIds` or `allowedCidrBlocks` were provided).

---

## AWS Provider Configuration

The entry point (`main.go`) loads stack input via `stackinput.LoadStackInput()` and passes it to `module.Resources()`.

The provider is configured in `module/main.go`:
- If `ProviderConfig` is nil → default AWS provider (ambient credentials from environment variables, instance profile, or `~/.aws/credentials`).
- If `ProviderConfig` is set → explicit provider with `AccessKey`, `SecretKey`, `Region`, and optional `SessionToken`.

The provider instance is passed to every resource via `pulumi.Provider(provider)`.

---

## Conditional Resource Patterns

The module uses a simple nil-check pattern for conditional resources:

```go
// securityGroup() returns nil when no ingress refs exist
createdSg, err := securityGroup(ctx, locals, provider)

// environment() checks for nil SG when building the security group list
if createdSg != nil {
    sgIds = append(sgIds, createdSg.ID())
}

// Output export checks for nil SG
if createdSg != nil {
    ctx.Export(OpSecurityGroupId, createdSg.ID())
}
```

This pattern avoids Pulumi `If` constructs and keeps the code straightforward.

---

## Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| `pulumi/pulumi/sdk` | v3 | Pulumi SDK |
| `pulumi/pulumi-aws/sdk` | v7 | AWS Classic provider |
| `pkg/errors` | — | Error wrapping |
| `openmcf/.../awsmwaaenvironment/v1` | — | Generated protobuf types |
| `openmcf/pkg/iac/pulumi/pulumimodule/stackinput` | — | Stack input loader |

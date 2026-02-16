# AwsCloudwatchLogGroup — Pulumi Module Architecture

## Overview

This module provisions a single CloudWatch Logs log group. It is one of the simplest modules in the AWS provider — a single Pulumi resource with straightforward field mapping.

## Resource Graph

```
AwsCloudwatchLogGroupStackInput
    └── cloudwatch.LogGroup (the log group)
```

## File Structure

| File | Purpose |
|------|---------|
| `module/main.go` | Entry point — creates AWS provider, invokes `logGroup()`, exports outputs |
| `module/locals.go` | Initializes `Locals` struct with tags and resource reference |
| `module/outputs.go` | Output key constants (`log_group_arn`, `log_group_name`) |
| `module/log_group.go` | Creates the `cloudwatch.NewLogGroup` resource |

## Field Mapping

| Spec Field | Pulumi Arg | Notes |
|------------|------------|-------|
| `retentionInDays` | `RetentionInDays` | Only set when > 0; nil means never expire |
| `kmsKeyId` | `KmsKeyId` | Uses `GetValue()` on StringValueOrRef |
| `logGroupClass` | `LogGroupClass` | Only set when non-empty; nil defaults to STANDARD |
| `deletionProtectionEnabled` | — | Not yet available in Pulumi AWS SDK v7 |

## Naming

The log group name is derived from `metadata.name` via the Pulumi resource name argument.

# AwsRedshiftCluster Pulumi Module

This directory contains the Pulumi Go module for provisioning an Amazon Redshift
data warehouse cluster with optional subnet group, security group, parameter
group, and audit logging.

## Quick Start

```bash
make build
pulumi preview --stack dev
pulumi up --stack dev --yes
```

## Debugging

```bash
./debug.sh
```

## Module Structure

See `overview.md` for architecture details.

## Resources Created

The module creates up to five AWS resources:

1. **`aws.redshift.Cluster`** — The core data warehouse cluster
2. **`aws.redshift.SubnetGroup`** (conditional) — When `subnetIds` are provided
3. **`aws.ec2.SecurityGroup`** (conditional) — When `securityGroupIds` or `allowedCidrBlocks` are provided
4. **`aws.redshift.ParameterGroup`** (conditional) — When inline `parameters` are provided
5. **`aws.redshift.Logging`** (conditional) — When `logging` is configured

## Outputs

| Key | Description |
|-----|-------------|
| `cluster_identifier` | Unique identifier of the Redshift cluster |
| `cluster_arn` | ARN of the cluster |
| `cluster_namespace_arn` | Namespace ARN for data sharing |
| `endpoint` | Connection endpoint (address:port) |
| `dns_name` | DNS hostname (without port) |
| `database_name` | Default database name |
| `port` | TCP port for connections |
| `subnet_group_name` | Managed subnet group name (if created) |
| `security_group_id` | Managed security group ID (if created) |
| `parameter_group_name` | Managed parameter group name (if created) |
| `master_password_secret_arn` | Secrets Manager secret ARN (if managed password) |

# AwsNeptuneCluster Pulumi Module Architecture

## Overview

This Pulumi module deploys an AWS Neptune cluster with all associated resources including subnet groups, security groups, parameter groups, and cluster instances. Neptune is a fully managed graph database supporting Gremlin (property graph) and SPARQL (RDF) query languages.

## Module Structure

```
module/
├── main.go              # Controller/orchestrator
├── locals.go            # Local values and label generation
├── outputs.go           # Output constant definitions
├── cluster.go           # Neptune cluster resource
├── instances.go          # Cluster instance creation
├── subnet_group.go      # Neptune subnet group resource
├── security_group.go    # VPC security group resource
└── parameter_group.go   # Cluster parameter group resource
```

## Resource Flow

```
┌─────────────────────────────────────────────────────────────┐
│                        main.go                               │
│                   (Resources function)                       │
└──────────────────────────┬──────────────────────────────────┘
                           │
        ┌──────────────────┼──────────────────┐
        │                  │                  │
        ▼                  ▼                  ▼
┌───────────────┐  ┌───────────────┐  ┌───────────────┐
│ Security Group│  │ Subnet Group  │  │Parameter Group│
│  (optional)   │  │  (optional)   │  │  (optional)   │
└───────┬───────┘  └───────┬───────┘  └───────┬───────┘
        │                  │                  │
        └──────────────────┼──────────────────┘
                           │
                           ▼
                  ┌─────────────────┐
                  │ Neptune         │
                  │ Cluster         │
                  └────────┬────────┘
                           │
                           ▼
                  ┌─────────────────┐
                  │ Cluster         │
                  │ Instances       │
                  └─────────────────┘
```

## Key Design Decisions

### 1. Optional Resource Creation

Resources are created conditionally based on spec configuration:

- **Subnet Group**: Created only when `subnetIds` is provided and `neptuneSubnetGroupName` is not set
- **Security Group**: Created only when `allowedCidrBlocks` or `securityGroupIds` are provided
- **Parameter Group**: Created only when `clusterParameters` are provided

### 2. Credential Handling

AWS credentials are passed through `AwsNeptuneClusterStackInput.ProviderConfig`, not embedded in the spec. Neptune does not use master username/password; access is controlled via IAM database authentication and network security.

### 3. Instance Scaling

The module creates the specified number of instances (`instanceCount`) with identical configuration. The first instance is the primary writer; additional instances are read replicas. Use `instanceClass: db.serverless` with `serverlessV2Scaling` for Neptune Serverless.

### 4. Resource Naming

Resources use the `metadata.id` from the manifest as the identifier, ensuring consistent naming across deployments.

## Outputs

The module exports the following outputs (matching `AwsNeptuneClusterStackOutputs`):

| Output | Description |
|--------|-------------|
| `cluster_endpoint` | Primary writer endpoint for Gremlin/SPARQL |
| `cluster_reader_endpoint` | Reader endpoint for load-balanced reads |
| `cluster_id` | Cluster identifier |
| `cluster_arn` | Cluster ARN |
| `cluster_resource_id` | Internal AWS resource ID |
| `cluster_port` | Connection port (default 8182) |
| `db_subnet_group_name` | Subnet group name (if created) |
| `security_group_id` | Security group ID (if created) |
| `cluster_parameter_group_name` | Parameter group name (if created) |
| `hosted_zone_id` | Route 53 hosted zone ID for endpoint |

## Dependencies

- `github.com/pulumi/pulumi-aws/sdk/v7/go/aws` - AWS provider
- `github.com/pulumi/pulumi/sdk/v3/go/pulumi` - Pulumi SDK
- `github.com/pkg/errors` - Error wrapping

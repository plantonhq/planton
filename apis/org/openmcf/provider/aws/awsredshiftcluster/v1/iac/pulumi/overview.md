# Pulumi Module: AwsRedshiftCluster

## Architecture

The Pulumi module creates up to five AWS resources depending on the spec:

- **`aws.redshift.Cluster`** — The core Redshift data warehouse cluster
- **`aws.redshift.SubnetGroup`** — Subnet group for cluster placement (conditional)
- **`aws.ec2.SecurityGroup`** + rules — Managed security group with ingress (conditional)
- **`aws.redshift.ParameterGroup`** — Custom parameter group (conditional)
- **`aws.redshift.Logging`** — Audit logging to S3 or CloudWatch (conditional)

## File Structure

```
module/
├── main.go              # Entry point: Resources()
├── locals.go            # Variable initialization and AWS tags
├── cluster.go           # Redshift cluster resource creation
├── subnet_group.go      # Conditional subnet group
├── security_group.go    # Conditional security group with rules
├── parameter_group.go   # Conditional parameter group
├── logging.go           # Conditional audit logging
└── outputs.go           # Output key constants
```

## Resource Flow

1. `main.go` initializes locals (tags, naming, conditionals) and obtains the AWS provider
2. Conditional resources are created first:
   - `subnet_group.go` — if `subnetIds` has ≥ 2 entries
   - `security_group.go` — if `securityGroupIds` or `allowedCidrBlocks` are present
   - `parameter_group.go` — if `parameters` has entries
3. `cluster.go` creates the cluster, referencing conditional resources
4. `logging.go` enables audit logging after the cluster is created
5. Outputs are exported: all 11 stack output fields

## Conditional Resource Creation

The module uses Go `if` statements to conditionally create resources:

```go
if len(spec.SubnetIds) >= 2 {
    subnetGroup, err = createSubnetGroup(ctx, ...)
}
```

This is simpler than Terraform's `count`/`for_each` pattern and produces
cleaner plans when optional resources are not needed.

## Networking Modes

The module supports three network configurations:

- **Subnet group from IDs** — Provide `subnetIds` (≥ 2) to auto-create a subnet group
- **Existing subnet group** — Provide `clusterSubnetGroupName` to use a pre-existing group
- **Managed security group** — Provide `securityGroupIds` or `allowedCidrBlocks` plus `vpcId`

Security groups from `associateSecurityGroupIds` are always attached directly.

## Key Implementation Details

- Framework AWS tags are applied to all taggable resources
- Cluster identifier is derived from `metadata.name`
- Cluster type (`single-node` vs `multi-node`) is derived from `numberOfNodes`
- Password is mutually exclusive: either `masterPassword` or `manageMasterPassword`
- The managed security group uses `name_prefix` with `create_before_destroy` lifecycle
- Logging is attached after cluster creation via a separate resource

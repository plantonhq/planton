# AwsRedshiftCluster

An Amazon Redshift provisioned cluster -- a petabyte-scale columnar data warehouse for analytical (OLAP) SQL workloads over structured and semi-structured data.

The cluster owns the warehouse brain: compute topology (node type and count), credentials, encryption, snapshots, maintenance, and restore shapes. Subnets, security groups, IAM roles, KMS keys, and Elastic IPs compose by reference -- warehouse ingress rules live on the referenced `AwsSecurityGroup` nodes, never inside the cluster. Audit logging and cross-region snapshot copy are folded in as cluster settings (AWS keys both by the cluster identifier).

## Spec highlights

- **Compute topology** -- `nodeType` (RA3 recommended: managed storage that tiers between SSD and S3) and `numberOfNodes`; resizing either is an in-place elastic/classic resize, never a replacement.
- **Credentials** -- `manageMasterPassword` (recommended) keeps the admin password in Secrets Manager, generated and rotated by AWS, with the secret's ARN exported; or supply `masterPassword` directly (sensitive).
- **Availability** -- `multiAz` (RA3 multi-node failover to a standby) XOR `availabilityZoneRelocationEnabled` (move the single cluster between zones); `elasticIp` binds a static public address by reference on public clusters.
- **Data protection** -- encryption at rest (default on, matching AWS), automated + manual snapshot retention, final-snapshot enforcement, and folded `snapshotCopy` for cross-region disaster recovery.
- **Restore shapes** -- create from a snapshot by name (`snapshotIdentifier`, with `snapshotClusterIdentifier` disambiguation) or by ARN (`snapshotArn`), including cross-account restores via `ownerAccount`.
- **Data movement** -- `iamRoles` + `defaultIamRoleArn` for COPY/UNLOAD/Spectrum, and `enhancedVpcRouting` to force warehouse data movement through the VPC.
- **Observability** -- folded `logging` delivers connection/user-activity/user audit logs to S3 or CloudWatch Logs.
- **Parameters** -- inline `parameters` (a module-managed parameter group; `parameterGroupFamily` selects `redshift-2.0` when needed, defaulting to `redshift-1.0`) or an existing `clusterParameterGroupName`.
- **Cost governance** -- `usageLimits` cap per-feature consumption (Spectrum scans, concurrency-scaling time, cross-region datasharing) with log/emit-metric/disable breach actions; `scheduledActions` pause, resume, or resize the cluster on cron/at schedules (the classic nights-and-weekends lever). The action's IAM role must TRUST `scheduler.redshift.amazonaws.com` -- AWS validates the trust at create -- and action names are unique per AWS account, so keep the cluster name in them.
- **Cross-VPC and cross-account access** -- `endpointAccesses` create Redshift-managed VPC endpoints into other subnet groups (RA3 only; per-endpoint private addresses exported); `endpointAuthorizations` grant other AWS accounts permission to create their own endpoints to this cluster (the grantor side -- the grantee's endpoint lives in their account).
- **Snapshot schedule** -- `snapshotScheduleIdentifier` associates the cluster with an existing account-level schedule (AWS keeps one per cluster), replacing the default automated-snapshot cadence.

## Outputs

`cluster_identifier`, `cluster_arn`, `cluster_namespace_arn`, `endpoint` (address:port), `dns_name`, `database_name`, `port`, `subnet_group_name`, `parameter_group_name`, `master_password_secret_arn`, `endpoint_access_addresses` (keyed by endpoint name), `usage_limit_ids` (AWS-generated, keyed by feature/limit-type/period).

## How it works

Planton provisions via the Pulumi or Terraform module in `iac/`, both implementing the same contract at full parity. The API contract is protobuf-based (`spec.proto`); stack execution is orchestrated using `AwsRedshiftClusterIacInput` (provider credentials + IaC info).

## References

- Redshift management guide: https://docs.aws.amazon.com/redshift/latest/mgmt/welcome.html
- RA3 node types and managed storage: https://docs.aws.amazon.com/redshift/latest/mgmt/working-with-clusters.html
- Managed admin passwords: https://docs.aws.amazon.com/redshift/latest/mgmt/redshift-secrets-manager-integration.html

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

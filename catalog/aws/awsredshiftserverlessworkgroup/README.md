# AwsRedshiftServerlessWorkgroup

An Amazon Redshift Serverless workgroup -- the COMPUTE plane of the serverless warehouse: Redshift Processing Unit (RPU) capacity, VPC placement, network reachability, and query-level configuration.

A workgroup computes; the data it serves lives on the `AwsRedshiftServerlessNamespace` it attaches to by name. Many workgroups can serve one namespace -- a capped dev workgroup and an autoscaling production workgroup over the same data -- and each is created and destroyed without touching the namespace. Billing follows the compute: RPU-hours accrue only while queries execute, so an idle workgroup costs nothing. Subnets and security groups compose by reference -- warehouse ingress rules live on the referenced `AwsSecurityGroup` nodes, never inside the workgroup.

## Spec highlights

- **Capacity** -- `baseCapacity` (the RPU floor each query starts from; empty keeps the AWS default 128) XOR an enabled `pricePerformanceTarget` (AWS picks the baseline against a cost/speed dial: 1 cheapest, 100 fastest); `maxCapacity` caps worst-case spend either way.
- **VPC placement** -- `subnetIds` (at least THREE subnets in three distinct AZs -- an AWS serverless requirement) and `securityGroupIds`, both by reference; empty subnets fall back to the account's default VPC.
- **Reachability** -- `publiclyAccessible` (off by default), `port` (only 5431-5455 or 8191-8215 -- the ranges Redshift Serverless accepts), `enhancedVpcRouting` to force COPY/UNLOAD data movement through the VPC.
- **Query configuration** -- `configParameters` applies engine parameters (TLS enforcement, search path, query monitoring limits) directly to the workgroup; serverless has no parameter groups.
- **Release track** -- `trackName` (`current` / `trailing` / a named track).
- **Cost governance** -- `usageLimits` cap `serverless-compute` (RPU-hours) or `cross-region-datasharing` (terabytes) per day/week/month, with log/emit-metric/deactivate breach actions (`deactivate` stops queries until the period resets -- note the serverless vocabulary differs from provisioned clusters' `disable`).
- **Cross-VPC access** -- `endpointAccesses` create VPC endpoints into other subnets (or reuse the workgroup's own); per-endpoint private addresses are exported.
- **Custom domain** -- one per workgroup (AWS's model): a branded DNS name fronted by an ACM certificate (`AwsCertManagerCert` by reference); the CNAME pointing the domain at the workgroup endpoint stays yours to manage.

## Outputs

`workgroup_name`, `workgroup_id`, `arn`, `endpoint_address`, `port`, `endpoint_access_addresses` (keyed by endpoint name), `usage_limit_ids` (AWS-generated, keyed by usage-type/period), `custom_domain_certificate_expiry_time`.

## How it works

Planton provisions via the Pulumi or Terraform module in `iac/`, both implementing the same contract at full parity. The API contract is protobuf-based (`spec.proto`); stack execution is orchestrated using `AwsRedshiftServerlessWorkgroupIacInput` (provider credentials + IaC info).

## References

- Redshift Serverless overview: https://docs.aws.amazon.com/redshift/latest/mgmt/serverless-whatis.html
- RPU capacity: https://docs.aws.amazon.com/redshift/latest/mgmt/serverless-capacity.html
- Serverless networking requirements: https://docs.aws.amazon.com/redshift/latest/mgmt/serverless-known-issues.html

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).

# AWS Redshift Serverless Workgroup

Deploys an Amazon Redshift Serverless workgroup — the compute plane of the serverless warehouse: Redshift Processing Unit (RPU) capacity, VPC placement, network reachability, and query-level configuration. A workgroup computes; the data it serves lives on the [AwsRedshiftServerlessNamespace](/infra-catalog/aws-redshift-serverless-namespace) it attaches to by name. Billing follows the compute — RPU-hours accrue only while queries execute, so an idle workgroup costs nothing. Many workgroups can serve one namespace (a capped dev endpoint and an autoscaling production endpoint over the same data), and each is created and destroyed without touching what is stored.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Redshift Serverless Workgroup** -- the compute plane whose name is the resource name (create-time immutable); SQL clients connect to its endpoint. The workgroup carries the capacity posture (a fixed RPU baseline or the price-performance dial, plus the optional hard spend ceiling), VPC placement (subnets across three AZs minimum, guarded by the referenced security groups), reachability (enhanced VPC routing, private-by-default exposure, the connection port), and query configuration (session parameters and monitoring guardrails -- serverless has no parameter groups)
- **Custom Domain Association** -- created only when `customDomain` is configured; fronts the endpoint with a branded DNS name and an ACM certificate
- **VPC Endpoint Accesses** -- one per `endpointAccesses[]` entry; private endpoints into consuming VPCs
- **Usage Limits** -- one per `usageLimits[]` entry; RPU-hour and datasharing spend caps with escalating breach actions
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **The namespace first** -- deploy the [AwsRedshiftServerlessNamespace](/infra-catalog/aws-redshift-serverless-namespace) this workgroup serves; the workgroup references its `namespace_name` output.

### AWS Account

- **Three subnets, three AZs** -- Redshift Serverless refuses a workgroup with fewer than three subnets spanning three distinct Availability Zones (leave the list empty only to use the account's default VPC). Each subnet needs free IPs in proportion to base capacity.
- **Ingress on the security groups** -- warehouse ingress rules (e.g. port 5439 from BI tooling) belong on the referenced [AwsSecurityGroup](/infra-catalog/aws-security-group) nodes, never inside the workgroup.

## Deploy

### Console

Open the deployment store, find **AWS Redshift Serverless Workgroup**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Capped Development Workgroup** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsRedshiftServerlessWorkgroup
metadata:
  name: analytics-dev
  org: acme-corp
  env: dev
spec:
  region: us-west-2
  namespaceName:
    valueFrom:
      kind: AwsRedshiftServerlessNamespace
      name: analytics-data
      fieldPath: status.outputs.namespace_name
  baseCapacity: 8
  maxCapacity: 32
  subnetIds:
    - value: subnet-0a1b2c3d4e5f60001
    - value: subnet-0a1b2c3d4e5f60002
    - value: subnet-0a1b2c3d4e5f60003
```

```shell
planton apply -f redshift-serverless-workgroup.yaml
```

This creates a cost-bounded dev workgroup over the referenced namespace. An Infra Job tracks the provisioning in real time.

### InfraChart

When the workgroup deploys alongside its namespace, subnets, and security group in one chart, wire the references via ValueFromRef:

```yaml
spec:
  region: us-west-2
  namespaceName:
    valueFrom:
      kind: AwsRedshiftServerlessNamespace
      name: analytics-data
      fieldPath: status.outputs.namespace_name
  baseCapacity: 8
  maxCapacity: 32
  subnetIds:
    - valueFrom:
        kind: AwsSubnet
        name: warehouse-az1
        fieldPath: status.outputs.subnet_id
    - valueFrom:
        kind: AwsSubnet
        name: warehouse-az2
        fieldPath: status.outputs.subnet_id
    - valueFrom:
        kind: AwsSubnet
        name: warehouse-az3
        fieldPath: status.outputs.subnet_id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: warehouse-sg
        fieldPath: status.outputs.security_group_id
```

The InfraPipeline resolves the dependency graph, deploys the namespace, subnets, and security group first, then provisions the workgroup with the resolved values.

## Key Configuration

These are the most important decisions when configuring a serverless workgroup. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The capacity model** -- mutually exclusive by design: fix the RPU baseline yourself (`baseCapacity`, 0 keeps the AWS default 128), or enable `pricePerformanceTarget` and AWS picks and adjusts the baseline against a cost/speed dial (1 cheapest, 50 balanced, 100 fastest). The console's mode selector makes the exclusivity structural. `maxCapacity` applies on BOTH models — the worst-case-spend guardrail production workgroups should always set.

**The data/compute split** -- destroying or recreating this workgroup never touches the namespace's data. Run several workgroups over one namespace when different consumers need different capacity or network postures; the attachment itself is create-time immutable.

**Network posture** -- private by default (Query Editor and in-VPC BI need no public IP). `enhancedVpcRouting` forces COPY/UNLOAD data movement through the VPC where flow logs and endpoints govern it — the usual data-governance ask. The port accepts only 5431-5455 and 8191-8215 (default 5439).

**Query guardrails** -- parameters apply directly to the workgroup from the exact list the API accepts: `require_ssl` and `max_query_execution_time` are the production classics; the `max_*` family implements query monitoring rules that cancel runaway work. `enable_user_activity_logging` pairs with the namespace's `useractivitylog` export — the workgroup produces the trail, the namespace delivers it.

**Spend caps** -- `usageLimits` bound consumption per day, week, or month: `serverless-compute` limits are RPU-hours, `cross-region-datasharing` limits are terabytes transferred. Breach actions escalate from `log` through `emit-metric` to `deactivate`, which stops queries until the period resets (data is untouched). Note the serverless vocabulary: `deactivate`, where provisioned clusters say `disable`.

**Cross-VPC access and identity** -- `endpointAccesses` create VPC endpoints into consuming VPCs' subnets (or reuse the workgroup's own); each endpoint's private address is exported in `endpoint_access_addresses` keyed by name. `customDomain` fronts the endpoint with a branded DNS name and an ACM certificate -- one per workgroup, the certificate must live in the workgroup's region and cover the domain, and the CNAME pointing the domain at the workgroup endpoint stays yours to manage. AWS serializes management operations on a workgroup, so these satellites apply in a fixed order (custom domain, then endpoint accesses, then usage limits) rather than concurrently -- declaring several simply extends the apply a little.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsRedshiftServerlessNamespace** (required) | `namespaceName` | `status.outputs.namespace_name` |
| **AwsSubnet** | `subnetIds`, `endpointAccesses[].subnetIds` | `status.outputs.subnet_id` |
| **AwsSecurityGroup** (optional) | `securityGroupIds`, `endpointAccesses[].vpcSecurityGroupIds` | `status.outputs.security_group_id` |
| **AwsCertManagerCert** (optional) | `customDomain.certificateArn` | `status.outputs.cert_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `endpoint_address` | The DNS hostname SQL clients connect to | Application/BI connection strings |
| `port` | The port the workgroup accepts connections on | Paired with the endpoint address |
| `workgroup_name` | The workgroup name (the resource name) | GetCredentials, custom domain associations |
| `workgroup_id` | The unique identifier AWS assigns | Account-level automation and audits |
| `arn` | Amazon Resource Name of the workgroup | IAM policies and usage limits |
| `endpoint_access_addresses` | Private DNS addresses of VPC endpoints, keyed by endpoint name | Connection strings for consumers in other VPCs |
| `usage_limit_ids` | AWS-generated usage-limit IDs, keyed by usage-type/period | Out-of-band CLI operations, state import |
| `custom_domain_certificate_expiry_time` | When the custom domain's certificate expires (RFC 3339) | Renewal monitoring |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Capped dev workgroup** -- the smallest practical baseline (8 RPU) with a hard 32-RPU ceiling: cheap, bounded, and safe to leave running (idle costs nothing). Start from the **Capped Development Workgroup** preset.

**Price-performance production workgroup** -- AWS owns the baseline at the balanced level, a 512-RPU cap bounds spend, enhanced VPC routing governs data movement, TLS is required, and a four-hour query limit guards runaway work. Start from the **Price-Performance Production Workgroup** preset.

**Governed workgroup with private endpoints** -- a fixed baseline with daily RPU-hour deactivation, datasharing transfer logging, a VPC endpoint for a BI-tooling VPC, and a branded TLS domain. Start from the **Governed Workgroup with Private Endpoints and a Custom Domain** preset.

## Works With

- [**AWS Redshift Serverless Namespace**](/infra-catalog/aws-redshift-serverless-namespace) -- the data plane this workgroup computes for (references `namespace_name`)
- [**AWS Subnet**](/infra-catalog/aws-subnet) -- placement for the compute and its managed VPC endpoint
- [**AWS Security Group**](/infra-catalog/aws-security-group) -- carries the warehouse's ingress rules
- [**AWS ACM Certificate**](/infra-catalog/aws-cert-manager-cert) -- provides the TLS certificate for the custom domain
- [**AWS Redshift Cluster**](/infra-catalog/aws-redshift-cluster) -- the provisioned alternative when steady, predictable load makes reserved capacity cheaper

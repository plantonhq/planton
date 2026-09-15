# AwsManagedPrometheusScraper

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

AwsManagedPrometheusScraperSpec defines one AMP scraper: AWS's
agentless Prometheus collector. It scrapes metrics from a SOURCE (an
EKS cluster, or a bare VPC placement for non-EKS endpoints) and
writes them to a DESTINATION (an AMP workspace or a CloudWatch
dataset) - which is why the scraper is its own kind: it needs no AMP
workspace to exist.

The whole SOURCE replaces on change (AWS re-provisions the
collector); alias, destination, role configuration, and the scrape
configuration update in place. Creates run long (AWS provisions
collector infrastructure - up to 30 minutes) and deletes drain
before removal (up to 20).

## Example

```yaml
# Canonical AwsManagedPrometheusScraper example (hack/dev manifest and
# refgen Example source): an EKS-sourced scraper delivering to an AMP
# workspace, with the scrape configuration left unset so the modules
# resolve AWS's published default at deploy, plus component logging.
# Literal values stand in for composed references so the offline
# `tofu plan` renders every arm (including the default-configuration
# data-source read).
apiVersion: aws.planton.dev/v1alpha1
kind: AwsManagedPrometheusScraper
metadata:
  name: platform-scraper
  id: platform-scraper
  org: test-org
  env: dev
spec:
  region: us-west-2
  alias: platform-scraper
  sourceEks:
    clusterArn:
      value: arn:aws:eks:us-west-2:123456789012:cluster/platform
    subnetIds:
      - value: subnet-0a1b2c3d4e5f60001
      - value: subnet-0a1b2c3d4e5f60002
  ampWorkspaceArn:
    value: arn:aws:aps:us-west-2:123456789012:workspace/ws-11111111-2222-3333-4444-555555555555
  logging:
    logGroupArn:
      value: arn:aws:logs:us-west-2:123456789012:log-group:/amp/scraper
    components:
      - COLLECTOR
      - EXPORTER
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.alias` | `string` |  |  |  |
| `spec.sourceEks` | `AwsManagedPrometheusScraperEksSource` |  |  |  |
| `spec.sourceEks.clusterArn` | `string \| valueFrom` | yes |  | AwsEksCluster (`status.outputs.cluster_arn`) |
| `spec.sourceEks.subnetIds` | `[]string \| valueFrom` | yes |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.sourceEks.securityGroupIds` | `[]string \| valueFrom` |  |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.sourceVpc` | `AwsManagedPrometheusScraperVpcSource` |  |  |  |
| `spec.sourceVpc.subnetIds` | `[]string \| valueFrom` | yes |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.sourceVpc.securityGroupIds` | `[]string \| valueFrom` | yes |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.ampWorkspaceArn` | `string \| valueFrom` |  |  | AwsManagedPrometheus (`status.outputs.workspace_arn`) |
| `spec.cloudwatchDatasetArn` | `string \| valueFrom` |  |  |  |
| `spec.scrapeConfiguration` | `string` |  |  |  |
| `spec.roleConfiguration` | `AwsManagedPrometheusScraperRoleConfiguration` |  |  |  |
| `spec.roleConfiguration.sourceRoleArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.roleConfiguration.targetRoleArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.logging` | `AwsManagedPrometheusScraperLogging` |  |  |  |
| `spec.logging.logGroupArn` | `string \| valueFrom` | yes |  | AwsCloudwatchLogGroup (`status.outputs.log_group_arn`) |
| `spec.logging.components` | `[]string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region the scraper runs in. Example: "us-east-1".

- rule: {"string":{"minLen":"1"}}

### spec.alias

`string`

The scraper's display alias. Updates in place.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.sourceEks

`AwsManagedPrometheusScraperEksSource`

Scrape from an EKS cluster. Changing anything here replaces the
scraper.

### spec.sourceEks.clusterArn

`string | valueFrom` · required

The EKS cluster to scrape. Reference an AwsEksCluster cluster_arn
output or pass a literal ARN.

Containment-exempt: the scraper READS the cluster's metrics; its
collectors run on AWS-managed interfaces in the subnets below, never
as workloads inside the cluster. On a diagram the scraper stands where
its subnets place it, with a line to the cluster it scrapes -- the
verdict every other reader of a room already carries.

- references: AwsEksCluster (`status.outputs.cluster_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsEksCluster, name: <that resource's name>, fieldPath: status.outputs.cluster_arn}} -- a bare string does not parse

### spec.sourceEks.subnetIds

`[]string | valueFrom` · required

Subnets the scraper's collectors place into (the cluster's
subnets). Reference AwsSubnet subnet_id outputs or pass literal
IDs. CreateScraper requires at least two subnets (server
contract: "Number of subnets must be at least 2").

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: {"repeated":{"minItems":"2"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.sourceEks.securityGroupIds

`[]string | valueFrom`

Security groups on the scraper's network interfaces. Unset lets
AWS pick the cluster's security group.

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.sourceVpc

`AwsManagedPrometheusScraperVpcSource`

Scrape from endpoints in a VPC (no EKS cluster). Changing
anything here replaces the scraper.

### spec.sourceVpc.subnetIds

`[]string | valueFrom` · required

Subnets the scraper's collectors place into. Reference AwsSubnet
subnet_id outputs or pass literal IDs. CreateScraper requires at
least two subnets (server contract: "Number of subnets must be at
least 2", live-verified on a single-subnet create).

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: {"repeated":{"minItems":"2"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.sourceVpc.securityGroupIds

`[]string | valueFrom` · required

Security groups on the scraper's network interfaces (AWS requires
at least one for VPC sources). Reference AwsSecurityGroup
security_group_id outputs or pass literal IDs.

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: {"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.ampWorkspaceArn

`string | valueFrom`

Deliver scraped metrics to an AMP workspace. Reference an
AwsManagedPrometheus workspace_arn output or pass a literal ARN.

- references: AwsManagedPrometheus (`status.outputs.workspace_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsManagedPrometheus, name: <that resource's name>, fieldPath: status.outputs.workspace_arn}} -- a bare string does not parse

### spec.cloudwatchDatasetArn

`string | valueFrom`

Deliver scraped metrics to a CloudWatch dataset (CloudWatch's
Prometheus-metrics ingestion) by dataset ARN.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.scrapeConfiguration

`string`

The Prometheus scrape configuration YAML (scrape_configs,
relabeling, ...). Optional for EKS sources - unset lets the
modules resolve AWS's published default configuration at deploy
time. Required for VPC sources. Updates in place.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.roleConfiguration

`AwsManagedPrometheusScraperRoleConfiguration`

Cross-account scraping: the role pair AWS assumes (source_role in
the scraped account, target_role in the destination account).
Both or neither.

### spec.roleConfiguration.sourceRoleArn

`string | valueFrom`

The role in the SCRAPED account. Reference an AwsIamRole role_arn
output or pass a literal ARN.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.roleConfiguration.targetRoleArn

`string | valueFrom`

The role in the DESTINATION account.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.logging

`AwsManagedPrometheusScraperLogging`

Scraper component logging to a CloudWatch log group.

### spec.logging.logGroupArn

`string | valueFrom` · required

The receiving log group. Reference an AwsCloudwatchLogGroup
log_group_arn output or pass a literal group ARN. AWS requires the
":*" wildcard suffix on this ARN; the log group resource exports
the bare ARN - both engines append ":*" when absent, so wire the
natural output and never hand-craft the suffix.

- references: AwsCloudwatchLogGroup (`status.outputs.log_group_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsCloudwatchLogGroup, name: <that resource's name>, fieldPath: status.outputs.log_group_arn}} -- a bare string does not parse

### spec.logging.components

`[]string`

Which scraper components log. Unset logs AWS's default set.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["SERVICE_DISCOVERY","COLLECTOR","EXPORTER"]}}}}

## Validation Rules

- `spec.exactly_one_source`: configure exactly one of source_eks and source_vpc
- `spec.exactly_one_destination`: configure exactly one of amp_workspace_arn and cloudwatch_dataset_arn
- `spec.vpc_source_requires_scrape_configuration`: scrape_configuration is required with source_vpc - AWS's default configuration exists only for EKS sources
- `spec.role_configuration_pair`: source_role_arn and target_role_arn must be set together

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsManagedPrometheusScraper, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.scraper_id` | `string` | The scraper's AWS-generated ID (the provider's import ID). |
| `status.outputs.scraper_arn` | `string` | The scraper's ARN. |
| `status.outputs.scraper_role_arn` | `string` | The AWS-managed role the scraper writes to its destination with - grant it remote-write on cross-account destinations. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.sourceEks.clusterArn` | AwsEksCluster | `status.outputs.cluster_arn` |
| `spec.sourceEks.subnetIds` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.sourceEks.securityGroupIds` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.sourceVpc.subnetIds` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.sourceVpc.securityGroupIds` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.ampWorkspaceArn` | AwsManagedPrometheus | `status.outputs.workspace_arn` |
| `spec.roleConfiguration.sourceRoleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.roleConfiguration.targetRoleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.logging.logGroupArn` | AwsCloudwatchLogGroup | `status.outputs.log_group_arn` |

## See Also

- [Overview](../README.md)

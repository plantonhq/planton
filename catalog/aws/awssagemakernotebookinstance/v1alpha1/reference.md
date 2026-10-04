# AwsSagemakerNotebookInstance

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsSagemakerNotebookInstanceSpec defines the desired configuration
for an Amazon SageMaker AI notebook instance - a managed EC2 instance
running Jupyter - together with its folded lifecycle configuration
(bootstrap scripts). The instance's AWS name derives from
metadata.name.

Update semantics worth knowing: SageMaker STOPS the instance to apply
most changes and restarts it afterwards (the modules ride the
provider's stop-update-start choreography - budget several minutes
per change). Growing `volume_size_gb` updates in place; SHRINKING it
replaces the instance (AWS cannot shrink a volume).

## Example

```yaml
# Canonical AwsSagemakerNotebookInstance example (hack/dev manifest and
# refgen Example source): a VPC-confined GPU notebook exercising every
# arm - private networking, KMS, code repositories, IMDSv2, root
# lockdown, and both lifecycle scripts. Literal ARNs stand in for
# composed references so the offline `tofu plan` renders every arm.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSagemakerNotebookInstance
metadata:
  name: research-notebook
  id: research-notebook
  org: test-org
  env: dev
spec:
  region: us-west-2
  instanceType: ml.g4dn.xlarge
  roleArn:
    value: arn:aws:iam::123456789012:role/sagemaker-execution
  volumeSizeGb: 100
  subnetId:
    value: subnet-0a1b2c3d4e5f60001
  securityGroupIds:
    - value: sg-0a1b2c3d4e5f60001
  kmsKeyArn:
    value: arn:aws:kms:us-west-2:123456789012:key/1234abcd-12ab-34cd-56ef-1234567890ab
  directInternetAccess: Disabled
  rootAccess: Disabled
  platformIdentifier: notebook-al2023-v1
  defaultCodeRepository: https://github.com/example/research-notebooks.git
  additionalCodeRepositories:
    - https://github.com/example/shared-utils.git
  imdsMinimumVersion: "2"
  lifecycleConfig:
    onCreate: |
      #!/bin/bash
      set -e
      pip install --quiet pandas scikit-learn
    onStart: |
      #!/bin/bash
      echo "notebook started at $(date)" >> /home/ec2-user/SageMaker/.start-log
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.instanceType` | `string` |  |  |  |
| `spec.roleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.volumeSizeGb` | `int32` |  |  |  |
| `spec.subnetId` | `string \| valueFrom` |  |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.securityGroupIds` | `[]string \| valueFrom` |  |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.kmsKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.directInternetAccess` | `string` |  |  |  |
| `spec.rootAccess` | `string` |  |  |  |
| `spec.platformIdentifier` | `string` |  |  |  |
| `spec.defaultCodeRepository` | `string` |  |  |  |
| `spec.additionalCodeRepositories` | `[]string` |  |  |  |
| `spec.imdsMinimumVersion` | `string` |  |  |  |
| `spec.lifecycleConfig` | `AwsSagemakerNotebookInstanceLifecycleConfig` |  |  |  |
| `spec.lifecycleConfig.onCreate` | `string` |  |  |  |
| `spec.lifecycleConfig.onStart` | `string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the notebook instance will be created.
Example: "us-west-2", "us-east-1"

- rule: {"string":{"minLen":"1"}}

### spec.instanceType

`string`

Compute instance type (an "ml.*" type, e.g. "ml.t3.medium" - the
cheapest current-generation choice; the instance bills hourly
whether or not a notebook is open). AWS's accepted
set grows with every release - the value passes through to the API,
which rejects unknown types.

- rule: {"string":{"pattern":"^ml\\.[a-z0-9]+([.-][a-z0-9]+)*$"}}

### spec.roleArn

`string | valueFrom` · required

IAM role the notebook assumes for AWS API calls made from it. The
role must trust sagemaker.amazonaws.com.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.volumeSizeGb

`int32` · optional (explicit presence)

ML storage volume in GB (5-16384; AWS default 5). Growing updates
in place; shrinking REPLACES the instance.

- rule: {"int32":{"lte":16384,"gte":5}}

### spec.subnetId

`string | valueFrom`

Place the notebook in a VPC subnet (private notebooks). Changing it
replaces the instance.

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.securityGroupIds

`[]string | valueFrom`

Security groups applied to the notebook's ENI (requires
`subnet_id`). Changing them replaces the instance -- the ML storage
volume is destroyed with it.

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.kmsKeyArn

`string | valueFrom`

KMS key encrypting the ML storage volume at rest. Changing it
replaces the instance -- the ML storage volume is destroyed with it.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.directInternetAccess

`string`

"Enabled" (AWS default - the notebook gets a direct internet route)
or "Disabled" (traffic flows only through your VPC - requires
subnet_id and security_group_ids, plus a NAT or endpoint path for
training/hosting calls). Changing it replaces the instance.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["Enabled","Disabled"]}}

### spec.rootAccess

`string`

"Enabled" (AWS default) or "Disabled" - whether notebook users get
root on the instance. Lifecycle scripts always run as root.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["Enabled","Disabled"]}}

### spec.platformIdentifier

`string`

Runtime platform. Omitted = AWS default ("notebook-al2-v3").
"notebook-al1-v1", "notebook-al2-v1", and "notebook-al2-v2" are
deprecated platforms AWS still accepts for existing workloads;
prefer "notebook-al2-v3" (Amazon Linux 2, JupyterLab 4) or
"notebook-al2023-v1" (Amazon Linux 2023). Changing it replaces the
instance.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["notebook-al1-v1","notebook-al2-v1","notebook-al2-v2","notebook-al2-v3","notebook-al2023-v1"]}}

### spec.defaultCodeRepository

`string`

Git repository cloned as the notebook's default working directory -
an AWS CodeCommit/SageMaker code-repository NAME or any public Git
HTTPS URL.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.additionalCodeRepositories

`[]string`

Up to three additional Git repositories cloned alongside the
default one.

- rule: {"repeated":{"maxItems":"3","unique":true,"items":{"string":{"minLen":"1"}}}}

### spec.imdsMinimumVersion

`string`

Minimum instance-metadata-service version: "1" (IMDSv1 and v2 both
allowed) or "2" (IMDSv2 only - the hardened choice). Omitted = AWS
default ("1").

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["1","2"]}}

### spec.lifecycleConfig

`AwsSagemakerNotebookInstanceLifecycleConfig`

Bootstrap scripts run on the instance (the folded lifecycle
configuration - its AWS name derives from metadata.name).

- rule: at least one of on_create and on_start must be set

### spec.lifecycleConfig.onCreate

`string`

Shell script run ONCE, when the instance is first created (max
16384 characters base64-encoded - keep the plain script well under
that).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.lifecycleConfig.onStart

`string`

Shell script run EVERY time the instance starts, including at
creation.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

## Validation Rules

- `disabled_internet_requires_vpc`: direct_internet_access Disabled requires subnet_id and security_group_ids
- `security_groups_require_subnet`: security_group_ids requires subnet_id

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsSagemakerNotebookInstance, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.notebook_instance_name` | `string` | The notebook instance name (the AWS identity). |
| `status.outputs.notebook_instance_arn` | `string` | The Amazon Resource Name of the notebook instance. |
| `status.outputs.url` | `string` | URL to open the Jupyter notebook. |
| `status.outputs.network_interface_id` | `string` | The ENI SageMaker created in your subnet (set only for VPC notebooks). |
| `status.outputs.lifecycle_config_name` | `string` | The folded lifecycle configuration's name (empty when spec.lifecycle_config is not set). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.subnetId` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.securityGroupIds` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.kmsKeyArn` | AwsKmsKey | `status.outputs.key_arn` |

## See Also

- [Overview](../README.md)

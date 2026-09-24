# AwsEksFargateProfile

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsEksFargateProfileSpec declares which Kubernetes pods of an
AwsEksCluster run on AWS Fargate -- serverless, per-pod compute with
no EC2 nodes to size, patch, or scale. Pods whose namespace (and
optionally labels) match one of the profile's selectors are scheduled
onto Fargate; everything else keeps running on the cluster's node
groups.

The profile composes onto its neighbors instead of embedding them: the
cluster attaches by reference (status.outputs.name), the pod execution
role is a referenced AwsIamRole that carries its own policies
(AmazonEKSFargatePodExecutionRolePolicy; trust
"eks-fargate-pods.amazonaws.com"), and the subnets are referenced
AwsSubnet nodes. This component never modifies a role it merely
references.

The ENTIRE profile is create-time immutable in AWS -- name, cluster,
role, subnets, and selectors. Changing anything replaces the profile;
pods keep running through the replacement window only if a second
matching profile covers them. AWS also serializes profile operations:
one profile per cluster creates or deletes at a time.

The profile name comes from metadata.name (AWS limit: 63 characters).

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsEksFargateProfile
metadata:
  name: awseksfargateprofile-demo
spec:
  region: us-west-2
  clusterName:
    value: awsekscluster-demo
  podExecutionRoleArn:
    value: arn:aws:iam::123456789012:role/EksFargatePodExecutionRole
  subnetIds:
    - value: subnet-0123456789abcdef0
    - value: subnet-0123456789abcdef1
  selectors:
    - namespace: serverless
    - namespace: batch
      labels:
        compute: fargate
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.clusterName` | `string \| valueFrom` | yes |  | AwsEksCluster (`status.outputs.name`) |
| `spec.podExecutionRoleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.subnetIds` | `[]string \| valueFrom` | yes |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.selectors` | `[]AwsEksFargateProfileSelector` | yes |  |  |
| `spec.selectors[].namespace` | `string` | yes |  |  |
| `spec.selectors[].labels` | `map<string, string>` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region the profile's cluster lives in. Must match the
cluster's region. Example: "us-west-2", "eu-west-1".

- rule: {"string":{"minLen":"1"}}

### spec.clusterName

`string | valueFrom` · required

The EKS cluster the profile attaches to. Reference an
AwsEksCluster's name output or pass a literal cluster name for a
cluster managed outside Planton. Create-only in AWS.

- references: AwsEksCluster (`status.outputs.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsEksCluster, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.podExecutionRoleArn

`string | valueFrom` · required

The IAM role Fargate uses to run the matched pods -- pulling images,
writing logs. It must trust "eks-fargate-pods.amazonaws.com" and
carry AmazonEKSFargatePodExecutionRolePolicy -- attach it on the
AwsIamRole itself; this component never modifies a role it merely
references. Reference an AwsIamRole's role_arn output or pass a
literal ARN. Create-only in AWS.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.subnetIds

`[]string | valueFrom` · required

The subnets Fargate launches the matched pods into. PRIVATE subnets
only -- AWS rejects subnets whose route table carries an internet
gateway route; give the pods outbound internet through a NAT
gateway. Reference AwsSubnet subnet_id outputs or pass literal
subnet IDs. Create-only in AWS. A Fargate profile is a member of
its cluster and lives there on a diagram; the subnets are where its
pods land, so the reference is access, not placement -- otherwise a
profile on subnets its cluster does not name would be drawn outside
the cluster it belongs to.

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: {"required":true,"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.selectors

`[]AwsEksFargateProfileSelector` · required

Which pods run on Fargate: a pod matches the profile when it matches
ANY selector (namespace, plus every label when labels are given).
AWS allows at most 5 selectors per profile. Create-only in AWS.

- rule: {"required":true,"repeated":{"minItems":"1","maxItems":"5"}}

### spec.selectors[].namespace

`string` · required

The Kubernetes namespace to match. Wildcards are allowed -- "*"
matches any sequence, "?" any single character (e.g. "prod-*").

- rule: {"required":true,"string":{"minLen":"1","maxLen":"63"}}

### spec.selectors[].labels

`map<string, string>`

Labels a pod must ALL carry to match (AND semantics within a
selector). Values may use the same "*" / "?" wildcards. Empty
matches every pod in the namespace. AWS allows at most 5 label pairs
per selector.

- rule: {"map":{"maxPairs":"5","keys":{"string":{"maxLen":"63"}},"values":{"string":{"maxLen":"63"}}}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsEksFargateProfile, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.fargate_profile_arn` | `string` | fargate_profile_arn is the Amazon Resource Name of the profile -- arn:aws:eks:<region>:<account>:fargateprofile/<cluster>/<name>/<uuid>. |
| `status.outputs.fargate_profile_name` | `string` | fargate_profile_name is the name of the profile. |
| `status.outputs.status` | `string` | status is the profile's state after provisioning -- "ACTIVE" on a successful create. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.clusterName` | AwsEksCluster | `status.outputs.name` |
| `spec.podExecutionRoleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.subnetIds` | AwsSubnet | `status.outputs.subnet_id` |

## See Also

- [Overview](../README.md)

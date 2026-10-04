# AwsIamInstanceProfile

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsIamInstanceProfileSpec defines an IAM instance profile: the container
that delivers an IAM role to EC2 instances.

EC2 cannot assume a role directly -- it can only be launched with an
instance profile, which holds exactly one role. The instance metadata
service then vends the role's temporary credentials to whatever runs on the
instance. Everything EC2-shaped (an instance's iam_instance_profile_arn, a
launch template, an Auto Scaling group) references the profile, while
everything else on AWS (Lambda, ECS, EKS, ...) assumes the role directly.
Modeling the profile as its own kind keeps that boundary honest: roles
serve every service, and the profile exists only where EC2 needs it.

The profile name comes from metadata.name. Name and path are create-only
(changing them replaces the profile); the role can be swapped in place.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsIamInstanceProfile
metadata:
  name: web-server-profile-demo
spec:
  region: us-west-2
  path: "/compute/"
  # A literal role name (a role that already exists). In composed topologies
  # this is a valueFrom reference to an AwsIamRole's role_name output.
  role:
    value: "web-server-role-demo"
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.role` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_name`) |
| `spec.path` | `string` |  | `/` |  |

## Field Details

### spec.region

`string` · required

The AWS region used by the provider while managing this profile.
IAM is a global service -- the profile is usable in every region -- but
every AWS API call is still made against a regional endpoint, so a region
is required.
Example: "us-west-2", "eu-west-1".

- rule: {"string":{"minLen":"1"}}

### spec.role

`string | valueFrom` · required

The IAM role the profile carries, by NAME (not ARN) -- that is what the
AWS API takes when adding a role to a profile. Reference an AwsIamRole's
role_name output, or pass a literal role name for a role that exists
outside Planton. AWS allows at most one role per profile, and the role can
be swapped without replacing the profile (the old role is removed and the
new one added).

- references: AwsIamRole (`status.outputs.role_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_name}} -- a bare string does not parse

### spec.path

`string`

The IAM path for the profile, used to organize and match profiles in IAM
policies. Must begin and end with "/" (e.g. "/compute/"). Defaults to "/"
when omitted. Immutable: changing the path replaces the profile.

- default: `/`

## Validation Rules

- `path_format`: path must begin and end with '/' and contain no empty segments, e.g. '/' or '/compute/'
- `path_length`: path must be at most 512 characters

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsIamInstanceProfile, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.instance_profile_arn` | `string` | The ARN of the instance profile (e.g. "arn:aws:iam::123456789012:instance-profile/web-server"). What an EC2 instance's iam_instance_profile_arn references via status.outputs.instance_profile_arn. |
| `status.outputs.instance_profile_name` | `string` | The friendly name of the instance profile (mirrors metadata.name). Launch templates and some APIs take the profile by name rather than ARN. |
| `status.outputs.instance_profile_id` | `string` | The stable unique ID AWS assigns to the profile (e.g. "AIPA..."). |
| `status.outputs.role_name` | `string` | The name of the IAM role the profile carries, resolved from spec.role -- exported for convenience so downstream consumers can see the effective role without dereferencing the AwsIamRole resource. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.role` | AwsIamRole | `status.outputs.role_name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsBatchComputeEnvironment | `spec.computeResources.instanceRole` | `status.outputs.instance_profile_arn` |
| AwsEc2Instance | `spec.instanceProfile` | `status.outputs.instance_profile_name` |
| AwsEcsCluster | `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.ec2InstanceProfileArn` | `status.outputs.instance_profile_arn` |
| AwsLaunchTemplate | `spec.instanceProfile` | `status.outputs.instance_profile_arn` |

## See Also

- [Overview](../README.md)

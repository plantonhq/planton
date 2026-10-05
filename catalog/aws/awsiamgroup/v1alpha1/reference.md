# AwsIamGroup

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsIamGroupSpec defines one IAM group: the container that grants a
set of users a shared permission set. The group's permissions are
its managed-policy attachments (see AwsIamPolicy) plus inline
policies unique to this group; its MEMBERSHIP is the declarative
users list below.

The group name comes from metadata.name (IAM group names allow
letters, digits, and +=,.@_- with no spaces; metadata.name is
always a valid subset). Renaming updates the group in place at AWS
- the ARN recomputes but the group and its members persist.

IAM is a GLOBAL service: the group exists account-wide, and the
spec's region is only the provider endpoint region.

## Example

```yaml
# Canonical AwsIamGroup example (hack/dev manifest and refgen Example
# source): a developers group with declarative membership, an
# AWS-managed policy attachment, and one inline policy. Literal user
# names stand in for composed references so the offline `tofu plan`
# renders the membership path.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsIamGroup
metadata:
  name: developers
  id: developers
  org: test-org
  env: dev
spec:
  region: us-east-1
  path: /teams/
  users:
    - value: alice
    - value: bob
  managedPolicyArns:
    - value: arn:aws:iam::aws:policy/ReadOnlyAccess
  inlinePolicies:
    deploy-artifacts:
      Version: "2012-10-17"
      Statement:
        - Effect: Allow
          Action:
            - s3:PutObject
            - s3:GetObject
          Resource: arn:aws:s3:::example-artifacts/*
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.path` | `string` |  | `/` |  |
| `spec.users` | `[]string \| valueFrom` |  |  | AwsIamUser (`status.outputs.user_name`) |
| `spec.managedPolicyArns` | `[]string \| valueFrom` |  |  | AwsIamPolicy (`status.outputs.policy_arn`) |
| `spec.inlinePolicies` | `map<string, object>` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region used by the provider while managing the group.
IAM is a global service - the group is visible account-wide - but
every AWS API call is still made against a regional endpoint.
Example: "us-east-1", "eu-west-1".

- rule: {"string":{"minLen":"1"}}

### spec.path

`string`

The IAM path for the group, used to organize and match groups in
IAM policies (e.g. "/teams/"). Unset = "/". Changing the path
updates the group in place (like a rename).

- default: `/`

### spec.users

`[]string | valueFrom`

The group's members - the DECLARATIVE membership: this list is
authoritative, so users added to the group outside this resource
are removed on the next apply. Reference an AwsIamUser's
user_name output or pass literal user names. The users must
already exist - IAM rejects unknown names.

- references: AwsIamUser (`status.outputs.user_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamUser, name: <that resource's name>, fieldPath: status.outputs.user_name}} -- a bare string does not parse

### spec.managedPolicyArns

`[]string | valueFrom`

Managed policies attached to this group. Reference an
AwsIamPolicy's policy_arn output or a literal ARN (literals are
how AWS-managed policies like arn:aws:iam::aws:policy/
ReadOnlyAccess attach). Attachments are reconciled in place:
adding or removing an entry attaches or detaches without touching
the group, and a policy attached outside this resource stays
attached (unlike `users`, which IS authoritative). Prefer managed
policies for anything reusable; use inline_policies for
permissions unique to this one group.

- references: AwsIamPolicy (`status.outputs.policy_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamPolicy, name: <that resource's name>, fieldPath: status.outputs.policy_arn}} -- a bare string does not parse

### spec.inlinePolicies

`map<string, object>`

Inline policies embedded in this group: a map of policy name to a
free-form JSON permission document. An inline policy lives and
dies with the group - use managed_policy_arns for anything shared
or reusable.

- rule: {"map":{"keys":{"string":{"maxLen":"128"}}}}

## Validation Rules

- `path_format`: path must begin and end with '/' and contain no empty segments, e.g. '/' or '/teams/'
- `path_length`: path must be at most 512 characters

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsIamGroup, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.group_arn` | `string` | The group's ARN. |
| `status.outputs.group_name` | `string` | The group's name (also the provider's import ID, and the value policies and budget actions reference the group by). |
| `status.outputs.group_id` | `string` | The group's AWS-generated stable unique ID (survives renames). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.users` | AwsIamUser | `status.outputs.user_name` |
| `spec.managedPolicyArns` | AwsIamPolicy | `status.outputs.policy_arn` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsBudget | `spec.actions[].iamActionDefinition.groups` | `status.outputs.group_name` |

## See Also

- [Overview](../README.md)

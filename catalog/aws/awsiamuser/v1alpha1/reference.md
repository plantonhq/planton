# AwsIamUser

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsIamUserSpec defines an IAM user: a long-lived identity with permanent
credentials, for the narrow set of cases temporary role credentials cannot
cover (external CI systems without OIDC federation, legacy tooling, break-
glass access).

Prefer roles wherever possible -- a role's credentials expire in hours,
while a user's access key works until rotated. When a user is genuinely
needed, this component keeps it disciplined: reusable permissions attach as
managed-policy references (see AwsIamPolicy), a permissions boundary can cap
what the user can ever do, and access-key creation is explicit.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsIamUser
metadata:
  name: cicd-user-demo
spec:
  region: us-west-2
  userName: "cicd-demo-user"
  managedPolicyArns:
    # Literal ARN (AWS-managed policy). Planton-defined policies attach via a
    # valueFrom reference to an AwsIamPolicy's policy_arn output.
    - value: "arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess"
  inlinePolicies:
    customAccess:
      Version: "2012-10-17"
      Statement:
        - Effect: "Allow"
          Action:
            - "s3:PutObject"
          Resource: "arn:aws:s3:::demo-bucket/*"
  # The rotation lever: flip to Inactive to suspend the key without deleting
  # it (id and secret survive; AWS rejects requests signed with it).
  accessKeyStatus: Active
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.userName` | `string` | yes |  |  |
| `spec.path` | `string` |  | `/` |  |
| `spec.managedPolicyArns` | `[]string \| valueFrom` |  |  | AwsIamPolicy (`status.outputs.policy_arn`) |
| `spec.inlinePolicies` | `map<string, object>` |  |  |  |
| `spec.permissionsBoundary` | `string \| valueFrom` |  |  | AwsIamPolicy (`status.outputs.policy_arn`) |
| `spec.disableAccessKeys` | `bool` |  |  |  |
| `spec.forceDestroy` | `bool` |  |  |  |
| `spec.accessKeyStatus` | `string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region used by the provider while managing this user.
IAM is a global service -- the user exists in every region -- but every
AWS API call is still made against a regional endpoint, so a region is
required.
Example: "us-west-2", "eu-west-1".

- rule: {"string":{"minLen":"1"}}

### spec.userName

`string` · required

The IAM user name. Unlike most identifiers this is mutable -- AWS renames
the user in place. 1-64 characters: letters, digits, and +=,.@_- .

- rule: {"required":true,"string":{"pattern":"^[a-zA-Z0-9+=,.@_-]{1,64}$"}}

### spec.path

`string`

The IAM path for the user, used to organize and match users in IAM
policies (e.g. scope a permission to "arn:aws:iam::<acct>:user/ci/*").
Must begin and end with "/" (e.g. "/ci/"). Defaults to "/" when omitted.
Updatable in place (users, unlike roles and policies, can move paths).

- default: `/`

### spec.managedPolicyArns

`[]string | valueFrom`

Managed policies to attach, each a reference to an AwsIamPolicy's
policy_arn output or a literal ARN (literals are how AWS-managed policies
like arn:aws:iam::aws:policy/ReadOnlyAccess attach). Attachments are
reconciled in place: adding or removing an entry attaches or detaches
without touching the user.

- references: AwsIamPolicy (`status.outputs.policy_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamPolicy, name: <that resource's name>, fieldPath: status.outputs.policy_arn}} -- a bare string does not parse

### spec.inlinePolicies

`map<string, object>`

Inline policies embedded in this user: a map of policy name to a
free-form JSON permission document. An inline policy lives and dies with
the user; anything reused across principals belongs in a first-class
AwsIamPolicy attached via managed_policy_arns.

- rule: {"map":{"keys":{"string":{"maxLen":"128"}}}}

### spec.permissionsBoundary

`string | valueFrom`

An optional permissions boundary: a managed policy whose grants cap the
maximum permissions this user can ever have -- effective permissions are
the INTERSECTION of the boundary and the user's permission policies.
Especially valuable on users, whose credentials are long-lived. Reference
an AwsIamPolicy's policy_arn output or pass a literal policy ARN.

- references: AwsIamPolicy (`status.outputs.policy_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamPolicy, name: <that resource's name>, fieldPath: status.outputs.policy_arn}} -- a bare string does not parse

### spec.disableAccessKeys

`bool`

If true, no access key is created for this user. By default one active
access key is created and its id/secret are exported as (sensitive) stack
outputs -- the usual reason a user exists. Disable for console-only or
externally-keyed users.

### spec.forceDestroy

`bool`

Whether deleting the user also deletes credentials and artifacts created
OUTSIDE this resource -- console login profiles, extra access keys, MFA
devices, SSH keys, signing certificates. Off by default: deletion fails
if such artifacts exist, surfacing them instead of silently destroying
credentials someone may depend on. Turn on for ephemeral or CI-owned
users where teardown must always succeed.

### spec.accessKeyStatus

`string`

Lifecycle state of the managed access key: "Active" (the AWS default) or
"Inactive". An Inactive key keeps its id and secret but AWS rejects every
request signed with it -- the standard rotation lever: create the
replacement credential, flip this key "Inactive" to prove nothing still
depends on it, then delete it. Updatable in place (no key re-creation, so
the secret output never changes). Empty keeps the key Active. Only
meaningful when the access key exists (disable_access_keys is false).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["Active","Inactive"]}}

## Validation Rules

- `path_format`: path must begin and end with '/' and contain no empty segments, e.g. '/' or '/ci/'
- `path_length`: path must be at most 512 characters
- `access_key_status_requires_key`: access_key_status has no effect when disable_access_keys is true -- remove one of the two

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsIamUser, name: <resource-name>, fieldPath: status.outputs.<output>}`. A sensitive output is a secret the resource generates: on Planton it is kept in the organization's secret store and the output holds a `$secret/` reference, so feed it only to a sensitive field.

| Output | Type | Description |
|---|---|---|
| `status.outputs.user_arn` | `string` | user_arn is the ARN of the created IAM user. |
| `status.outputs.access_key_id` | `string` | access_key_id is the access key ID for the user (present if an access key was created). |
| `status.outputs.secret_access_key` | `string` (sensitive) | secret_access_key is the base64-encoded secret key associated with the access key. This value is sensitive and should be handled securely. |
| `status.outputs.console_url` | `string` | console_url is the AWS console sign-in URL for this user. |
| `status.outputs.user_name` | `string` | user_name is the friendly name of the IAM user. |
| `status.outputs.user_id` | `string` | user_id is the stable unique ID of the IAM user. |
| `status.outputs.access_key_status` | `string` | access_key_status is the access key's status (Active/Inactive) as the provider reports it after apply -- the rotation-lever position, verifiable against ListAccessKeys. Empty when no access key was created. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.managedPolicyArns` | AwsIamPolicy | `status.outputs.policy_arn` |
| `spec.permissionsBoundary` | AwsIamPolicy | `status.outputs.policy_arn` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsBudget | `spec.actions[].iamActionDefinition.users` | `status.outputs.user_name` |
| AwsIamGroup | `spec.users` | `status.outputs.user_name` |

## See Also

- [Overview](../README.md)

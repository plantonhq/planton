# AwsOrganizationPolicy

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

AwsOrganizationPolicySpec defines the desired configuration for one
AWS Organizations policy (a service control policy or any of its
twelve sibling types) together with its attachments to roots,
organizational units, and member accounts.

The policy's display name is an explicit spec field - AWS allows
names with spaces metadata.name cannot carry. AWS identifies the
policy as "p-..." (the import ID); each attachment imports as
"{target_id}:{policy_id}".

Managing policies requires the ORGANIZATION'S MANAGEMENT account,
an organization with ALL features enabled, and the policy's type
enabled on the organization (AwsOrganization's
enabled_policy_types). AWS-managed policies (like FullAWSAccess)
cannot be adopted by this resource - the provider refuses to import
them.

## Example

```yaml
# Canonical AwsOrganizationPolicy example (hack/dev manifest and
# refgen Example source): a deny-guardrail SCP attached to the
# organization root and one OU. Literal target IDs stand in for
# composed references so the offline `tofu plan` renders both
# attachments.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsOrganizationPolicy
metadata:
  name: deny-leave-organization
  id: deny-leave-organization
  org: test-org
  env: dev
spec:
  region: us-west-2
  policyName: Deny Leaving The Organization
  type: SERVICE_CONTROL_POLICY
  description: Blocks member accounts from leaving the organization on their own.
  content:
    Version: "2012-10-17"
    Statement:
      - Sid: DenyLeave
        Effect: Deny
        Action: organizations:LeaveOrganization
        Resource: "*"
  attachments:
    - targetId:
        value: r-e2e1
    - targetId:
        value: ou-e2e1-workload1
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.policyName` | `string` | yes |  |  |
| `spec.type` | `string` |  |  |  |
| `spec.content` | `object` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.attachments` | `[]AwsOrganizationPolicyAttachment` |  |  |  |
| `spec.attachments[].targetId` | `string \| valueFrom` | yes |  | AwsOrganizationalUnit (`status.outputs.ou_id`) |

## Field Details

### spec.region

`string` · required

The AWS region used by the provider while managing the policy.
Organizations is a global service - the policy is the same
everywhere - but every AWS API call is still made against a
regional endpoint, so a region is required.
Example: "us-west-2", "eu-west-1".

- rule: {"string":{"minLen":"1"}}

### spec.policyName

`string` · required

The policy's display name (1-128 characters; spaces are legal -
which is why this is an explicit field rather than
metadata.name). Renames apply in place.

- rule: {"string":{"minLen":"1","maxLen":"128"}}

### spec.type

`string`

The policy type. Unset = SERVICE_CONTROL_POLICY (the provider
default). IMMUTABLE - changing the type forces replacement. The
type must be enabled on the organization before attachments
anywhere will succeed.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["SERVICE_CONTROL_POLICY","RESOURCE_CONTROL_POLICY","TAG_POLICY","BACKUP_POLICY","AISERVICES_OPT_OUT_POLICY","CHATBOT_POLICY","DECLARATIVE_POLICY_EC2","SECURITYHUB_POLICY","INSPECTOR_POLICY","UPGRADE_ROLLOUT_POLICY","BEDROCK_POLICY","S3_POLICY","NETWORK_SECURITY_DIRECTOR_POLICY"]}}

### spec.content

`object` · required

The policy document as free-form JSON in the syntax of the
policy's type (an SCP/RCP statement document, a tag policy
document, ...). Content updates apply in place.

- rule: {"required":true}

### spec.description

`string`

Human description of the policy, up to 512 characters.

- rule: {"string":{"maxLen":"512"}}

### spec.attachments

`[]AwsOrganizationPolicyAttachment`

Where the policy attaches: each entry binds the policy to one
root, OU, or member account. Both leaves of an attachment are
immutable - changing a target re-attaches (detach + attach).

### spec.attachments[].targetId

`string | valueFrom` · required

The attachment target: an AwsOrganizationalUnit reference (the
default wiring - SCPs typically govern OUs), a literal "r-..."
root ID, a literal "ou-..." ID, a literal 12-digit account ID, or
an AwsOrganizationAccount/AwsOrganization reference by field
path. IMMUTABLE.

Containment-exempt: a policy GOVERNS its targets from above; it is
not a member of the unit or organization it attaches to, and one
policy attached to several units could not live in all of them. On
a diagram the policy stands beside the tenancy tree with a line to
each target it governs.

- references: AwsOrganizationalUnit (`status.outputs.ou_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsOrganizationalUnit, name: <that resource's name>, fieldPath: status.outputs.ou_id}} -- a bare string does not parse

## Validation Rules

- `spec.attachment_target_literal_format`: a literal attachment target_id must be an organization root (r-...), an organizational unit (ou-...), or a 12-digit account ID
- `spec.attachment_targets_unique`: attachments entries must have unique targets

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsOrganizationPolicy, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.policy_id` | `string` | The policy's AWS-generated ID ("p-..." - also the provider's import ID; each folded attachment imports as "{target_id}:{policy_id}"). |
| `status.outputs.arn` | `string` | The policy's ARN. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.attachments[].targetId` | AwsOrganizationalUnit | `status.outputs.ou_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsBudget | `spec.actions[].scpActionDefinition.policyId` | `status.outputs.policy_id` |

## See Also

- [Overview](../README.md)

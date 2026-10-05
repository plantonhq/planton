# AwsOrganizationalUnit

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsOrganizationalUnitSpec defines the desired configuration for one
AWS Organizations organizational unit: a container in the
organization's OU tree that member accounts are grouped into and
organization policies attach to.

The OU's display name is an explicit spec field - AWS allows names
with spaces and arbitrary characters ("Core Services") that
metadata.name cannot carry. AWS identifies the OU as "ou-..." (the
import ID); the parent reference is immutable (moving an OU means
recreating it), while the name renames in place.

Creating OUs requires the ORGANIZATION'S MANAGEMENT account.

## Example

```yaml
# Canonical AwsOrganizationalUnit example (hack/dev manifest and
# refgen Example source): a first-level OU under the organization
# root. A literal root ID stands in for the composed AwsOrganization
# reference so the offline `tofu plan` renders.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsOrganizationalUnit
metadata:
  name: workloads-ou
  id: workloads-ou
  org: test-org
  env: dev
spec:
  region: us-west-2
  ouName: Workloads
  parentId:
    value: r-e2e1
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.ouName` | `string` | yes |  |  |
| `spec.parentId` | `string \| valueFrom` | yes |  | AwsOrganization (`status.outputs.root_id`) |

## Field Details

### spec.region

`string` · required

The AWS region used by the provider while managing the OU.
Organizations is a global service - the OU is the same
everywhere - but every AWS API call is still made against a
regional endpoint, so a region is required.
Example: "us-west-2", "eu-west-1".

- rule: {"string":{"minLen":"1"}}

### spec.ouName

`string` · required

The OU's display name (1-128 characters; spaces and arbitrary
characters are legal - which is why this is an explicit field
rather than metadata.name). Renames apply in place.

- rule: {"string":{"minLen":"1","maxLen":"128"}}

### spec.parentId

`string | valueFrom` · required

The parent this OU hangs under: the organization root for a
first-level OU (the default wiring - an AwsOrganization
reference resolves its root_id), a parent AwsOrganizationalUnit's
ou_id reference for a nested OU, or a literal "r-..."/"ou-..."
ID for a pre-existing tree. IMMUTABLE - changing the parent
forces replacement (AWS moves accounts, not OUs).

- references: AwsOrganization (`status.outputs.root_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsOrganization, name: <that resource's name>, fieldPath: status.outputs.root_id}} -- a bare string does not parse

## Validation Rules

- `spec.parent_id_literal_format`: a literal parent_id must be an organization root (r-...) or an organizational unit (ou-...)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsOrganizationalUnit, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.ou_id` | `string` | The OU's AWS-generated ID ("ou-..." - also the provider's import ID). Nested OUs, member accounts, and policy attachments reference it. |
| `status.outputs.arn` | `string` | The OU's ARN. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.parentId` | AwsOrganization | `status.outputs.root_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsOrganizationAccount | `spec.parentId` | `status.outputs.ou_id` |
| AwsOrganizationPolicy | `spec.attachments[].targetId` | `status.outputs.ou_id` |

## See Also

- [Overview](../README.md)

# An AWS Organization and its units are the rooms of the tenancy tree

## What changed

- **`AwsOrganization` and `AwsOrganizationalUnit` are container kinds.** The organization is the outer wall of the tenancy tree AWS's own console draws; organizational units are rooms inside it; member accounts stand inside those. A first-level unit names the organization through its `parent_id`, a nested unit or a member account names its unit the same way, and each is now placed inside what it names. Until now neither kind carried `container_kind: true`, so a landing zone with a dozen accounts drew as a dozen cards joined by lines rather than the tree.
- **An Organizations policy attachment is containment-exempt.** `AwsOrganizationPolicyAttachment.target_id` names the root, unit, or account a policy governs. A policy is a guard applied from above, not a member of what it attaches to -- and one policy attached to several units could not live in all of them. On a diagram it stands beside the tree with a line to each target.
- The containment-decision registry gains the unit's and the account's `parent_id` as `contained` and the attachment's `target_id` as `exempt`; nothing else moved.

## Why

The doctrine on `container_kind` asks two questions: is the kind a place or scope in the provider's own model, and would an engineer whiteboarding the system draw a box around other resources for it? An organization and its units answer yes to both -- the console's own tree view is that whiteboard. An account whose `parent_id` is left unset sits under the organization's root; the catalog cannot express that without a reference, and the picture honestly leaves it beside the wall until a manifest names the root.

## How to check

```bash
go test ./shared/cloudresourcekind/ -run TestContainmentDecisions   # green; the golden carries the two contained lines and the exempt attachment
grep -n -B1 'container_kind: true' shared/cloudresourcekind/cloud_resource_kind.proto | grep -A1 'awsorg\|awsou'
```

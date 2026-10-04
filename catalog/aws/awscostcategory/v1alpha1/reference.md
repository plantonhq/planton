# AwsCostCategory

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsCostCategorySpec defines one Cost Explorer cost category: a
named dimension YOU define that groups every line of spend into
exactly one of its values ("Team" = engineering/marketing/...,
"Environment" = prod/staging/...). Rules evaluate IN ORDER and the
first match wins; unmatched spend lands in default_value (or the
literal "Uncategorized" when unset).

Rules come in two shapes: REGULAR rules assign a value you name to
the spend a Cost Explorer expression selects; INHERITED_VALUE rules
take the value from a dimension itself (the account name, a tag's
value) - the "one rule fans out to N values" shape.

Split-charge rules re-allocate one value's costs across target
values (chargeback of shared platform costs). Cost Explorer is
account-global (served from us-east-1); the spec's region is the
provider endpoint region. The provider's rule_version argument is
module-pinned to "CostCategoryExpression.v1" - the only value the
API accepts.

## Example

```yaml
# Canonical AwsCostCategory example (hack/dev manifest and refgen
# Example source): a "Cost Center" category with an expression rule, a
# tag-inherited rule, a default value, and a proportional split of
# shared platform costs.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCostCategory
metadata:
  name: cost-center
  id: cost-center
  org: test-org
  env: dev
spec:
  region: us-east-1
  categoryName: Cost Center
  defaultValue: shared
  rules:
    # Rules take SERVICE_CODE (service codes like "AmazonEC2"), never
    # the SERVICE display-name dimension other CE surfaces accept.
    - value: platform
      rule:
        or:
          - dimension:
              key: SERVICE_CODE
              values:
                - AmazonEC2
          - tag:
              key: team
              values:
                - platform
    - type: INHERITED_VALUE
      inheritedValue:
        dimensionName: TAG
        dimensionKey: team
  splitChargeRules:
    - source: shared
      targets:
        - platform
      method: PROPORTIONAL
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.categoryName` | `string` | yes |  |  |
| `spec.defaultValue` | `string` |  |  |  |
| `spec.effectiveStart` | `string` |  |  |  |
| `spec.rules` | `[]AwsCostCategoryRule` | yes |  |  |
| `spec.rules[].type` | `string` |  |  |  |
| `spec.rules[].value` | `string` |  |  |  |
| `spec.rules[].rule` | `AwsCostCategoryExpression` |  |  |  |
| `spec.rules[].rule.dimension` | `AwsCostCategoryExpressionDimension` |  |  |  |
| `spec.rules[].rule.dimension.key` | `string` |  |  |  |
| `spec.rules[].rule.dimension.matchOptions` | `[]string` |  |  |  |
| `spec.rules[].rule.dimension.values` | `[]string` | yes |  |  |
| `spec.rules[].rule.tag` | `AwsCostCategoryExpressionTag` |  |  |  |
| `spec.rules[].rule.tag.key` | `string` |  |  |  |
| `spec.rules[].rule.tag.matchOptions` | `[]string` |  |  |  |
| `spec.rules[].rule.tag.values` | `[]string` |  |  |  |
| `spec.rules[].rule.costCategory` | `AwsCostCategoryExpressionCostCategory` |  |  |  |
| `spec.rules[].rule.costCategory.key` | `string` | yes |  |  |
| `spec.rules[].rule.costCategory.matchOptions` | `[]string` |  |  |  |
| `spec.rules[].rule.costCategory.values` | `[]string` |  |  |  |
| `spec.rules[].rule.and` | `[]AwsCostCategoryExpressionNode` |  |  |  |
| `spec.rules[].rule.and[].dimension` | `AwsCostCategoryExpressionDimension` |  |  |  |
| `spec.rules[].rule.and[].dimension.key` | `string` |  |  |  |
| `spec.rules[].rule.and[].dimension.matchOptions` | `[]string` |  |  |  |
| `spec.rules[].rule.and[].dimension.values` | `[]string` | yes |  |  |
| `spec.rules[].rule.and[].tag` | `AwsCostCategoryExpressionTag` |  |  |  |
| `spec.rules[].rule.and[].tag.key` | `string` |  |  |  |
| `spec.rules[].rule.and[].tag.matchOptions` | `[]string` |  |  |  |
| `spec.rules[].rule.and[].tag.values` | `[]string` |  |  |  |
| `spec.rules[].rule.and[].costCategory` | `AwsCostCategoryExpressionCostCategory` |  |  |  |
| `spec.rules[].rule.and[].costCategory.key` | `string` | yes |  |  |
| `spec.rules[].rule.and[].costCategory.matchOptions` | `[]string` |  |  |  |
| `spec.rules[].rule.and[].costCategory.values` | `[]string` |  |  |  |
| `spec.rules[].rule.and[].and` | `[]AwsCostCategoryExpressionLeaf` |  |  |  |
| `spec.rules[].rule.and[].and[].dimension` | `AwsCostCategoryExpressionDimension` |  |  |  |
| `spec.rules[].rule.and[].and[].dimension.key` | `string` |  |  |  |
| `spec.rules[].rule.and[].and[].dimension.matchOptions` | `[]string` |  |  |  |
| `spec.rules[].rule.and[].and[].dimension.values` | `[]string` | yes |  |  |
| `spec.rules[].rule.and[].and[].tag` | `AwsCostCategoryExpressionTag` |  |  |  |
| `spec.rules[].rule.and[].and[].tag.key` | `string` |  |  |  |
| `spec.rules[].rule.and[].and[].tag.matchOptions` | `[]string` |  |  |  |
| `spec.rules[].rule.and[].and[].tag.values` | `[]string` |  |  |  |
| `spec.rules[].rule.and[].and[].costCategory` | `AwsCostCategoryExpressionCostCategory` |  |  |  |
| `spec.rules[].rule.and[].and[].costCategory.key` | `string` | yes |  |  |
| `spec.rules[].rule.and[].and[].costCategory.matchOptions` | `[]string` |  |  |  |
| `spec.rules[].rule.and[].and[].costCategory.values` | `[]string` |  |  |  |
| `spec.rules[].rule.and[].or` | `[]AwsCostCategoryExpressionLeaf` |  |  |  |
| `spec.rules[].rule.and[].or[].dimension` | `AwsCostCategoryExpressionDimension` |  |  |  |
| `spec.rules[].rule.and[].or[].dimension.key` | `string` |  |  |  |
| `spec.rules[].rule.and[].or[].dimension.matchOptions` | `[]string` |  |  |  |
| `spec.rules[].rule.and[].or[].dimension.values` | `[]string` | yes |  |  |
| `spec.rules[].rule.and[].or[].tag` | `AwsCostCategoryExpressionTag` |  |  |  |
| `spec.rules[].rule.and[].or[].tag.key` | `string` |  |  |  |
| `spec.rules[].rule.and[].or[].tag.matchOptions` | `[]string` |  |  |  |
| `spec.rules[].rule.and[].or[].tag.values` | `[]string` |  |  |  |
| `spec.rules[].rule.and[].or[].costCategory` | `AwsCostCategoryExpressionCostCategory` |  |  |  |
| `spec.rules[].rule.and[].or[].costCategory.key` | `string` | yes |  |  |
| `spec.rules[].rule.and[].or[].costCategory.matchOptions` | `[]string` |  |  |  |
| `spec.rules[].rule.and[].or[].costCategory.values` | `[]string` |  |  |  |
| `spec.rules[].rule.and[].not` | `AwsCostCategoryExpressionLeaf` |  |  |  |
| `spec.rules[].rule.and[].not.dimension` | `AwsCostCategoryExpressionDimension` |  |  |  |
| `spec.rules[].rule.and[].not.dimension.key` | `string` |  |  |  |
| `spec.rules[].rule.and[].not.dimension.matchOptions` | `[]string` |  |  |  |
| `spec.rules[].rule.and[].not.dimension.values` | `[]string` | yes |  |  |
| `spec.rules[].rule.and[].not.tag` | `AwsCostCategoryExpressionTag` |  |  |  |
| `spec.rules[].rule.and[].not.tag.key` | `string` |  |  |  |
| `spec.rules[].rule.and[].not.tag.matchOptions` | `[]string` |  |  |  |
| `spec.rules[].rule.and[].not.tag.values` | `[]string` |  |  |  |
| `spec.rules[].rule.and[].not.costCategory` | `AwsCostCategoryExpressionCostCategory` |  |  |  |
| `spec.rules[].rule.and[].not.costCategory.key` | `string` | yes |  |  |
| `spec.rules[].rule.and[].not.costCategory.matchOptions` | `[]string` |  |  |  |
| `spec.rules[].rule.and[].not.costCategory.values` | `[]string` |  |  |  |
| `spec.rules[].rule.or` | `[]AwsCostCategoryExpressionNode` |  |  |  |
| `spec.rules[].rule.or[].dimension` | `AwsCostCategoryExpressionDimension` |  |  |  |
| `spec.rules[].rule.or[].dimension.key` | `string` |  |  |  |
| `spec.rules[].rule.or[].dimension.matchOptions` | `[]string` |  |  |  |
| `spec.rules[].rule.or[].dimension.values` | `[]string` | yes |  |  |
| `spec.rules[].rule.or[].tag` | `AwsCostCategoryExpressionTag` |  |  |  |
| `spec.rules[].rule.or[].tag.key` | `string` |  |  |  |
| `spec.rules[].rule.or[].tag.matchOptions` | `[]string` |  |  |  |
| `spec.rules[].rule.or[].tag.values` | `[]string` |  |  |  |
| `spec.rules[].rule.or[].costCategory` | `AwsCostCategoryExpressionCostCategory` |  |  |  |
| `spec.rules[].rule.or[].costCategory.key` | `string` | yes |  |  |
| `spec.rules[].rule.or[].costCategory.matchOptions` | `[]string` |  |  |  |
| `spec.rules[].rule.or[].costCategory.values` | `[]string` |  |  |  |
| `spec.rules[].rule.or[].and` | `[]AwsCostCategoryExpressionLeaf` |  |  |  |
| `spec.rules[].rule.or[].and[].dimension` | `AwsCostCategoryExpressionDimension` |  |  |  |
| `spec.rules[].rule.or[].and[].dimension.key` | `string` |  |  |  |
| `spec.rules[].rule.or[].and[].dimension.matchOptions` | `[]string` |  |  |  |
| `spec.rules[].rule.or[].and[].dimension.values` | `[]string` | yes |  |  |
| `spec.rules[].rule.or[].and[].tag` | `AwsCostCategoryExpressionTag` |  |  |  |
| `spec.rules[].rule.or[].and[].tag.key` | `string` |  |  |  |
| `spec.rules[].rule.or[].and[].tag.matchOptions` | `[]string` |  |  |  |
| `spec.rules[].rule.or[].and[].tag.values` | `[]string` |  |  |  |
| `spec.rules[].rule.or[].and[].costCategory` | `AwsCostCategoryExpressionCostCategory` |  |  |  |
| `spec.rules[].rule.or[].and[].costCategory.key` | `string` | yes |  |  |
| `spec.rules[].rule.or[].and[].costCategory.matchOptions` | `[]string` |  |  |  |
| `spec.rules[].rule.or[].and[].costCategory.values` | `[]string` |  |  |  |
| `spec.rules[].rule.or[].or` | `[]AwsCostCategoryExpressionLeaf` |  |  |  |
| `spec.rules[].rule.or[].or[].dimension` | `AwsCostCategoryExpressionDimension` |  |  |  |
| `spec.rules[].rule.or[].or[].dimension.key` | `string` |  |  |  |
| `spec.rules[].rule.or[].or[].dimension.matchOptions` | `[]string` |  |  |  |
| `spec.rules[].rule.or[].or[].dimension.values` | `[]string` | yes |  |  |
| `spec.rules[].rule.or[].or[].tag` | `AwsCostCategoryExpressionTag` |  |  |  |
| `spec.rules[].rule.or[].or[].tag.key` | `string` |  |  |  |
| `spec.rules[].rule.or[].or[].tag.matchOptions` | `[]string` |  |  |  |
| `spec.rules[].rule.or[].or[].tag.values` | `[]string` |  |  |  |
| `spec.rules[].rule.or[].or[].costCategory` | `AwsCostCategoryExpressionCostCategory` |  |  |  |
| `spec.rules[].rule.or[].or[].costCategory.key` | `string` | yes |  |  |
| `spec.rules[].rule.or[].or[].costCategory.matchOptions` | `[]string` |  |  |  |
| `spec.rules[].rule.or[].or[].costCategory.values` | `[]string` |  |  |  |
| `spec.rules[].rule.or[].not` | `AwsCostCategoryExpressionLeaf` |  |  |  |
| `spec.rules[].rule.or[].not.dimension` | `AwsCostCategoryExpressionDimension` |  |  |  |
| `spec.rules[].rule.or[].not.dimension.key` | `string` |  |  |  |
| `spec.rules[].rule.or[].not.dimension.matchOptions` | `[]string` |  |  |  |
| `spec.rules[].rule.or[].not.dimension.values` | `[]string` | yes |  |  |
| `spec.rules[].rule.or[].not.tag` | `AwsCostCategoryExpressionTag` |  |  |  |
| `spec.rules[].rule.or[].not.tag.key` | `string` |  |  |  |
| `spec.rules[].rule.or[].not.tag.matchOptions` | `[]string` |  |  |  |
| `spec.rules[].rule.or[].not.tag.values` | `[]string` |  |  |  |
| `spec.rules[].rule.or[].not.costCategory` | `AwsCostCategoryExpressionCostCategory` |  |  |  |
| `spec.rules[].rule.or[].not.costCategory.key` | `string` | yes |  |  |
| `spec.rules[].rule.or[].not.costCategory.matchOptions` | `[]string` |  |  |  |
| `spec.rules[].rule.or[].not.costCategory.values` | `[]string` |  |  |  |
| `spec.rules[].rule.not` | `AwsCostCategoryExpressionNode` |  |  |  |
| `spec.rules[].rule.not.dimension` | `AwsCostCategoryExpressionDimension` |  |  |  |
| `spec.rules[].rule.not.dimension.key` | `string` |  |  |  |
| `spec.rules[].rule.not.dimension.matchOptions` | `[]string` |  |  |  |
| `spec.rules[].rule.not.dimension.values` | `[]string` | yes |  |  |
| `spec.rules[].rule.not.tag` | `AwsCostCategoryExpressionTag` |  |  |  |
| `spec.rules[].rule.not.tag.key` | `string` |  |  |  |
| `spec.rules[].rule.not.tag.matchOptions` | `[]string` |  |  |  |
| `spec.rules[].rule.not.tag.values` | `[]string` |  |  |  |
| `spec.rules[].rule.not.costCategory` | `AwsCostCategoryExpressionCostCategory` |  |  |  |
| `spec.rules[].rule.not.costCategory.key` | `string` | yes |  |  |
| `spec.rules[].rule.not.costCategory.matchOptions` | `[]string` |  |  |  |
| `spec.rules[].rule.not.costCategory.values` | `[]string` |  |  |  |
| `spec.rules[].rule.not.and` | `[]AwsCostCategoryExpressionLeaf` |  |  |  |
| `spec.rules[].rule.not.and[].dimension` | `AwsCostCategoryExpressionDimension` |  |  |  |
| `spec.rules[].rule.not.and[].dimension.key` | `string` |  |  |  |
| `spec.rules[].rule.not.and[].dimension.matchOptions` | `[]string` |  |  |  |
| `spec.rules[].rule.not.and[].dimension.values` | `[]string` | yes |  |  |
| `spec.rules[].rule.not.and[].tag` | `AwsCostCategoryExpressionTag` |  |  |  |
| `spec.rules[].rule.not.and[].tag.key` | `string` |  |  |  |
| `spec.rules[].rule.not.and[].tag.matchOptions` | `[]string` |  |  |  |
| `spec.rules[].rule.not.and[].tag.values` | `[]string` |  |  |  |
| `spec.rules[].rule.not.and[].costCategory` | `AwsCostCategoryExpressionCostCategory` |  |  |  |
| `spec.rules[].rule.not.and[].costCategory.key` | `string` | yes |  |  |
| `spec.rules[].rule.not.and[].costCategory.matchOptions` | `[]string` |  |  |  |
| `spec.rules[].rule.not.and[].costCategory.values` | `[]string` |  |  |  |
| `spec.rules[].rule.not.or` | `[]AwsCostCategoryExpressionLeaf` |  |  |  |
| `spec.rules[].rule.not.or[].dimension` | `AwsCostCategoryExpressionDimension` |  |  |  |
| `spec.rules[].rule.not.or[].dimension.key` | `string` |  |  |  |
| `spec.rules[].rule.not.or[].dimension.matchOptions` | `[]string` |  |  |  |
| `spec.rules[].rule.not.or[].dimension.values` | `[]string` | yes |  |  |
| `spec.rules[].rule.not.or[].tag` | `AwsCostCategoryExpressionTag` |  |  |  |
| `spec.rules[].rule.not.or[].tag.key` | `string` |  |  |  |
| `spec.rules[].rule.not.or[].tag.matchOptions` | `[]string` |  |  |  |
| `spec.rules[].rule.not.or[].tag.values` | `[]string` |  |  |  |
| `spec.rules[].rule.not.or[].costCategory` | `AwsCostCategoryExpressionCostCategory` |  |  |  |
| `spec.rules[].rule.not.or[].costCategory.key` | `string` | yes |  |  |
| `spec.rules[].rule.not.or[].costCategory.matchOptions` | `[]string` |  |  |  |
| `spec.rules[].rule.not.or[].costCategory.values` | `[]string` |  |  |  |
| `spec.rules[].rule.not.not` | `AwsCostCategoryExpressionLeaf` |  |  |  |
| `spec.rules[].rule.not.not.dimension` | `AwsCostCategoryExpressionDimension` |  |  |  |
| `spec.rules[].rule.not.not.dimension.key` | `string` |  |  |  |
| `spec.rules[].rule.not.not.dimension.matchOptions` | `[]string` |  |  |  |
| `spec.rules[].rule.not.not.dimension.values` | `[]string` | yes |  |  |
| `spec.rules[].rule.not.not.tag` | `AwsCostCategoryExpressionTag` |  |  |  |
| `spec.rules[].rule.not.not.tag.key` | `string` |  |  |  |
| `spec.rules[].rule.not.not.tag.matchOptions` | `[]string` |  |  |  |
| `spec.rules[].rule.not.not.tag.values` | `[]string` |  |  |  |
| `spec.rules[].rule.not.not.costCategory` | `AwsCostCategoryExpressionCostCategory` |  |  |  |
| `spec.rules[].rule.not.not.costCategory.key` | `string` | yes |  |  |
| `spec.rules[].rule.not.not.costCategory.matchOptions` | `[]string` |  |  |  |
| `spec.rules[].rule.not.not.costCategory.values` | `[]string` |  |  |  |
| `spec.rules[].inheritedValue` | `AwsCostCategoryInheritedValue` |  |  |  |
| `spec.rules[].inheritedValue.dimensionName` | `string` |  |  |  |
| `spec.rules[].inheritedValue.dimensionKey` | `string` |  |  |  |
| `spec.splitChargeRules` | `[]AwsCostCategorySplitChargeRule` |  |  |  |
| `spec.splitChargeRules[].source` | `string` | yes |  |  |
| `spec.splitChargeRules[].targets` | `[]string` | yes |  |  |
| `spec.splitChargeRules[].method` | `string` |  |  |  |
| `spec.splitChargeRules[].parameters` | `[]AwsCostCategorySplitChargeParameter` |  |  |  |
| `spec.splitChargeRules[].parameters[].type` | `string` |  |  |  |
| `spec.splitChargeRules[].parameters[].values` | `[]string` | yes |  |  |

## Field Details

### spec.region

`string` · required

The AWS region used by the provider while managing the category.
Cost Explorer is account-global (served from us-east-1) - every
API call still needs a regional endpoint. Example: "us-east-1".

- rule: {"string":{"minLen":"1"}}

### spec.categoryName

`string` · required

The category's name - an explicit field because category names
legally carry spaces and mixed case metadata.name cannot
("Cost Center"). 1-50 characters; changing the name replaces the
category.

- rule: {"string":{"minLen":"1","maxLen":"50"}}

### spec.defaultValue

`string`

The value unmatched spend lands in (1-50 characters; the pattern
encodes the length). Unset = AWS reports such spend as
"Uncategorized".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^.{1,50}$"}}

### spec.effectiveStart

`string`

When the category starts applying, as an ISO-8601 month start
("2024-01-01T00:00:00Z" - AWS accepts month boundaries only).
Unset = AWS starts at the current month. AWS re-categorizes the
current and previous month retroactively on rule changes.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[0-9]{4}-[0-9]{2}-01T00:00:00Z$"}}

### spec.rules

`[]AwsCostCategoryRule` · required

The categorization rules, evaluated in order - first match wins.

- rule: {"repeated":{"minItems":"1"}}
- rule: a rule is exactly one shape: set rule (REGULAR) or inherited_value (INHERITED_VALUE), never both
- rule: type INHERITED_VALUE pairs with inherited_value; REGULAR (or unset) pairs with rule
- rule: REGULAR rules need value - the category value the matched spend is assigned to

### spec.rules[].type

`string`

The rule's shape. Unset = REGULAR (the provider default).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["REGULAR","INHERITED_VALUE"]}}

### spec.rules[].value

`string`

For REGULAR rules: the category value assigned to matched spend
(1-50 characters; the pattern encodes the length).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^.{1,50}$"}}

### spec.rules[].rule

`AwsCostCategoryExpression`

For REGULAR rules: the Cost Explorer expression selecting the
spend this rule matches.

### spec.rules[].rule.dimension

`AwsCostCategoryExpressionDimension`

A dimension leaf (service, region, account, ...).

### spec.rules[].rule.dimension.key

`string`

The dimension to filter on. Cost category rules accept exactly
this server-enforced subset of Cost Explorer dimensions - NOT the
full vocabulary other CE surfaces take (CreateCostCategoryDefinition
names the set in its 400: "Allowed dimension(s): USAGE_TYPE,
RECORD_TYPE, LINKED_ACCOUNT_NAME, SERVICE_CODE, LINKED_ACCOUNT,
BILLING_ENTITY, REGION"; server-verified 2026-08-25). The
by-service intent uses SERVICE_CODE with service CODES
("AmazonS3", "AmazonEC2"), never the SERVICE display-name
dimension ("Amazon Simple Storage Service") that budgets and
anomaly monitors accept.

- rule: {"string":{"in":["BILLING_ENTITY","LINKED_ACCOUNT","LINKED_ACCOUNT_NAME","RECORD_TYPE","REGION","SERVICE_CODE","USAGE_TYPE"]}}

### spec.rules[].rule.dimension.matchOptions

`[]string`

How values match. Unset = EQUALS.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.rules[].rule.dimension.values

`[]string` · required

The values to match.

- rule: {"repeated":{"minItems":"1","items":{"string":{"maxLen":"1024"}}}}

### spec.rules[].rule.tag

`AwsCostCategoryExpressionTag`

A cost-allocation-tag leaf.

### spec.rules[].rule.tag.key

`string`

The tag key.

- rule: {"string":{"maxLen":"1024"}}

### spec.rules[].rule.tag.matchOptions

`[]string`

How values match (see the dimension leaf; ABSENT matches spend
missing the tag entirely).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.rules[].rule.tag.values

`[]string`

The tag values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.rules[].rule.costCategory

`AwsCostCategoryExpressionCostCategory`

A leaf matching ANOTHER cost category's values (categories can
build on categories).

### spec.rules[].rule.costCategory.key

`string` · required

The other category's name.

- rule: {"string":{"minLen":"1","maxLen":"50"}}

### spec.rules[].rule.costCategory.matchOptions

`[]string`

How values match (see the dimension leaf).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.rules[].rule.costCategory.values

`[]string`

The category values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.rules[].rule.and

`[]AwsCostCategoryExpressionNode`

Children that must ALL match.

### spec.rules[].rule.and[].dimension

`AwsCostCategoryExpressionDimension`

A dimension leaf.

### spec.rules[].rule.and[].dimension.key

`string`

The dimension to filter on. Cost category rules accept exactly
this server-enforced subset of Cost Explorer dimensions - NOT the
full vocabulary other CE surfaces take (CreateCostCategoryDefinition
names the set in its 400: "Allowed dimension(s): USAGE_TYPE,
RECORD_TYPE, LINKED_ACCOUNT_NAME, SERVICE_CODE, LINKED_ACCOUNT,
BILLING_ENTITY, REGION"; server-verified 2026-08-25). The
by-service intent uses SERVICE_CODE with service CODES
("AmazonS3", "AmazonEC2"), never the SERVICE display-name
dimension ("Amazon Simple Storage Service") that budgets and
anomaly monitors accept.

- rule: {"string":{"in":["BILLING_ENTITY","LINKED_ACCOUNT","LINKED_ACCOUNT_NAME","RECORD_TYPE","REGION","SERVICE_CODE","USAGE_TYPE"]}}

### spec.rules[].rule.and[].dimension.matchOptions

`[]string`

How values match. Unset = EQUALS.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.rules[].rule.and[].dimension.values

`[]string` · required

The values to match.

- rule: {"repeated":{"minItems":"1","items":{"string":{"maxLen":"1024"}}}}

### spec.rules[].rule.and[].tag

`AwsCostCategoryExpressionTag`

A cost-allocation-tag leaf.

### spec.rules[].rule.and[].tag.key

`string`

The tag key.

- rule: {"string":{"maxLen":"1024"}}

### spec.rules[].rule.and[].tag.matchOptions

`[]string`

How values match (see the dimension leaf; ABSENT matches spend
missing the tag entirely).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.rules[].rule.and[].tag.values

`[]string`

The tag values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.rules[].rule.and[].costCategory

`AwsCostCategoryExpressionCostCategory`

A leaf matching another cost category's values.

### spec.rules[].rule.and[].costCategory.key

`string` · required

The other category's name.

- rule: {"string":{"minLen":"1","maxLen":"50"}}

### spec.rules[].rule.and[].costCategory.matchOptions

`[]string`

How values match (see the dimension leaf).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.rules[].rule.and[].costCategory.values

`[]string`

The category values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.rules[].rule.and[].and

`[]AwsCostCategoryExpressionLeaf`

Leaf children that must ALL match.

### spec.rules[].rule.and[].and[].dimension

`AwsCostCategoryExpressionDimension`

A dimension leaf.

### spec.rules[].rule.and[].and[].dimension.key

`string`

The dimension to filter on. Cost category rules accept exactly
this server-enforced subset of Cost Explorer dimensions - NOT the
full vocabulary other CE surfaces take (CreateCostCategoryDefinition
names the set in its 400: "Allowed dimension(s): USAGE_TYPE,
RECORD_TYPE, LINKED_ACCOUNT_NAME, SERVICE_CODE, LINKED_ACCOUNT,
BILLING_ENTITY, REGION"; server-verified 2026-08-25). The
by-service intent uses SERVICE_CODE with service CODES
("AmazonS3", "AmazonEC2"), never the SERVICE display-name
dimension ("Amazon Simple Storage Service") that budgets and
anomaly monitors accept.

- rule: {"string":{"in":["BILLING_ENTITY","LINKED_ACCOUNT","LINKED_ACCOUNT_NAME","RECORD_TYPE","REGION","SERVICE_CODE","USAGE_TYPE"]}}

### spec.rules[].rule.and[].and[].dimension.matchOptions

`[]string`

How values match. Unset = EQUALS.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.rules[].rule.and[].and[].dimension.values

`[]string` · required

The values to match.

- rule: {"repeated":{"minItems":"1","items":{"string":{"maxLen":"1024"}}}}

### spec.rules[].rule.and[].and[].tag

`AwsCostCategoryExpressionTag`

A cost-allocation-tag leaf.

### spec.rules[].rule.and[].and[].tag.key

`string`

The tag key.

- rule: {"string":{"maxLen":"1024"}}

### spec.rules[].rule.and[].and[].tag.matchOptions

`[]string`

How values match (see the dimension leaf; ABSENT matches spend
missing the tag entirely).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.rules[].rule.and[].and[].tag.values

`[]string`

The tag values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.rules[].rule.and[].and[].costCategory

`AwsCostCategoryExpressionCostCategory`

A leaf matching another cost category's values.

### spec.rules[].rule.and[].and[].costCategory.key

`string` · required

The other category's name.

- rule: {"string":{"minLen":"1","maxLen":"50"}}

### spec.rules[].rule.and[].and[].costCategory.matchOptions

`[]string`

How values match (see the dimension leaf).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.rules[].rule.and[].and[].costCategory.values

`[]string`

The category values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.rules[].rule.and[].or

`[]AwsCostCategoryExpressionLeaf`

Leaf children of which AT LEAST ONE must match.

### spec.rules[].rule.and[].or[].dimension

`AwsCostCategoryExpressionDimension`

A dimension leaf.

### spec.rules[].rule.and[].or[].dimension.key

`string`

The dimension to filter on. Cost category rules accept exactly
this server-enforced subset of Cost Explorer dimensions - NOT the
full vocabulary other CE surfaces take (CreateCostCategoryDefinition
names the set in its 400: "Allowed dimension(s): USAGE_TYPE,
RECORD_TYPE, LINKED_ACCOUNT_NAME, SERVICE_CODE, LINKED_ACCOUNT,
BILLING_ENTITY, REGION"; server-verified 2026-08-25). The
by-service intent uses SERVICE_CODE with service CODES
("AmazonS3", "AmazonEC2"), never the SERVICE display-name
dimension ("Amazon Simple Storage Service") that budgets and
anomaly monitors accept.

- rule: {"string":{"in":["BILLING_ENTITY","LINKED_ACCOUNT","LINKED_ACCOUNT_NAME","RECORD_TYPE","REGION","SERVICE_CODE","USAGE_TYPE"]}}

### spec.rules[].rule.and[].or[].dimension.matchOptions

`[]string`

How values match. Unset = EQUALS.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.rules[].rule.and[].or[].dimension.values

`[]string` · required

The values to match.

- rule: {"repeated":{"minItems":"1","items":{"string":{"maxLen":"1024"}}}}

### spec.rules[].rule.and[].or[].tag

`AwsCostCategoryExpressionTag`

A cost-allocation-tag leaf.

### spec.rules[].rule.and[].or[].tag.key

`string`

The tag key.

- rule: {"string":{"maxLen":"1024"}}

### spec.rules[].rule.and[].or[].tag.matchOptions

`[]string`

How values match (see the dimension leaf; ABSENT matches spend
missing the tag entirely).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.rules[].rule.and[].or[].tag.values

`[]string`

The tag values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.rules[].rule.and[].or[].costCategory

`AwsCostCategoryExpressionCostCategory`

A leaf matching another cost category's values.

### spec.rules[].rule.and[].or[].costCategory.key

`string` · required

The other category's name.

- rule: {"string":{"minLen":"1","maxLen":"50"}}

### spec.rules[].rule.and[].or[].costCategory.matchOptions

`[]string`

How values match (see the dimension leaf).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.rules[].rule.and[].or[].costCategory.values

`[]string`

The category values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.rules[].rule.and[].not

`AwsCostCategoryExpressionLeaf`

A leaf child that must NOT match.

### spec.rules[].rule.and[].not.dimension

`AwsCostCategoryExpressionDimension`

A dimension leaf.

### spec.rules[].rule.and[].not.dimension.key

`string`

The dimension to filter on. Cost category rules accept exactly
this server-enforced subset of Cost Explorer dimensions - NOT the
full vocabulary other CE surfaces take (CreateCostCategoryDefinition
names the set in its 400: "Allowed dimension(s): USAGE_TYPE,
RECORD_TYPE, LINKED_ACCOUNT_NAME, SERVICE_CODE, LINKED_ACCOUNT,
BILLING_ENTITY, REGION"; server-verified 2026-08-25). The
by-service intent uses SERVICE_CODE with service CODES
("AmazonS3", "AmazonEC2"), never the SERVICE display-name
dimension ("Amazon Simple Storage Service") that budgets and
anomaly monitors accept.

- rule: {"string":{"in":["BILLING_ENTITY","LINKED_ACCOUNT","LINKED_ACCOUNT_NAME","RECORD_TYPE","REGION","SERVICE_CODE","USAGE_TYPE"]}}

### spec.rules[].rule.and[].not.dimension.matchOptions

`[]string`

How values match. Unset = EQUALS.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.rules[].rule.and[].not.dimension.values

`[]string` · required

The values to match.

- rule: {"repeated":{"minItems":"1","items":{"string":{"maxLen":"1024"}}}}

### spec.rules[].rule.and[].not.tag

`AwsCostCategoryExpressionTag`

A cost-allocation-tag leaf.

### spec.rules[].rule.and[].not.tag.key

`string`

The tag key.

- rule: {"string":{"maxLen":"1024"}}

### spec.rules[].rule.and[].not.tag.matchOptions

`[]string`

How values match (see the dimension leaf; ABSENT matches spend
missing the tag entirely).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.rules[].rule.and[].not.tag.values

`[]string`

The tag values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.rules[].rule.and[].not.costCategory

`AwsCostCategoryExpressionCostCategory`

A leaf matching another cost category's values.

### spec.rules[].rule.and[].not.costCategory.key

`string` · required

The other category's name.

- rule: {"string":{"minLen":"1","maxLen":"50"}}

### spec.rules[].rule.and[].not.costCategory.matchOptions

`[]string`

How values match (see the dimension leaf).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.rules[].rule.and[].not.costCategory.values

`[]string`

The category values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.rules[].rule.or

`[]AwsCostCategoryExpressionNode`

Children of which AT LEAST ONE must match.

### spec.rules[].rule.or[].dimension

`AwsCostCategoryExpressionDimension`

A dimension leaf.

### spec.rules[].rule.or[].dimension.key

`string`

The dimension to filter on. Cost category rules accept exactly
this server-enforced subset of Cost Explorer dimensions - NOT the
full vocabulary other CE surfaces take (CreateCostCategoryDefinition
names the set in its 400: "Allowed dimension(s): USAGE_TYPE,
RECORD_TYPE, LINKED_ACCOUNT_NAME, SERVICE_CODE, LINKED_ACCOUNT,
BILLING_ENTITY, REGION"; server-verified 2026-08-25). The
by-service intent uses SERVICE_CODE with service CODES
("AmazonS3", "AmazonEC2"), never the SERVICE display-name
dimension ("Amazon Simple Storage Service") that budgets and
anomaly monitors accept.

- rule: {"string":{"in":["BILLING_ENTITY","LINKED_ACCOUNT","LINKED_ACCOUNT_NAME","RECORD_TYPE","REGION","SERVICE_CODE","USAGE_TYPE"]}}

### spec.rules[].rule.or[].dimension.matchOptions

`[]string`

How values match. Unset = EQUALS.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.rules[].rule.or[].dimension.values

`[]string` · required

The values to match.

- rule: {"repeated":{"minItems":"1","items":{"string":{"maxLen":"1024"}}}}

### spec.rules[].rule.or[].tag

`AwsCostCategoryExpressionTag`

A cost-allocation-tag leaf.

### spec.rules[].rule.or[].tag.key

`string`

The tag key.

- rule: {"string":{"maxLen":"1024"}}

### spec.rules[].rule.or[].tag.matchOptions

`[]string`

How values match (see the dimension leaf; ABSENT matches spend
missing the tag entirely).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.rules[].rule.or[].tag.values

`[]string`

The tag values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.rules[].rule.or[].costCategory

`AwsCostCategoryExpressionCostCategory`

A leaf matching another cost category's values.

### spec.rules[].rule.or[].costCategory.key

`string` · required

The other category's name.

- rule: {"string":{"minLen":"1","maxLen":"50"}}

### spec.rules[].rule.or[].costCategory.matchOptions

`[]string`

How values match (see the dimension leaf).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.rules[].rule.or[].costCategory.values

`[]string`

The category values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.rules[].rule.or[].and

`[]AwsCostCategoryExpressionLeaf`

Leaf children that must ALL match.

### spec.rules[].rule.or[].and[].dimension

`AwsCostCategoryExpressionDimension`

A dimension leaf.

### spec.rules[].rule.or[].and[].dimension.key

`string`

The dimension to filter on. Cost category rules accept exactly
this server-enforced subset of Cost Explorer dimensions - NOT the
full vocabulary other CE surfaces take (CreateCostCategoryDefinition
names the set in its 400: "Allowed dimension(s): USAGE_TYPE,
RECORD_TYPE, LINKED_ACCOUNT_NAME, SERVICE_CODE, LINKED_ACCOUNT,
BILLING_ENTITY, REGION"; server-verified 2026-08-25). The
by-service intent uses SERVICE_CODE with service CODES
("AmazonS3", "AmazonEC2"), never the SERVICE display-name
dimension ("Amazon Simple Storage Service") that budgets and
anomaly monitors accept.

- rule: {"string":{"in":["BILLING_ENTITY","LINKED_ACCOUNT","LINKED_ACCOUNT_NAME","RECORD_TYPE","REGION","SERVICE_CODE","USAGE_TYPE"]}}

### spec.rules[].rule.or[].and[].dimension.matchOptions

`[]string`

How values match. Unset = EQUALS.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.rules[].rule.or[].and[].dimension.values

`[]string` · required

The values to match.

- rule: {"repeated":{"minItems":"1","items":{"string":{"maxLen":"1024"}}}}

### spec.rules[].rule.or[].and[].tag

`AwsCostCategoryExpressionTag`

A cost-allocation-tag leaf.

### spec.rules[].rule.or[].and[].tag.key

`string`

The tag key.

- rule: {"string":{"maxLen":"1024"}}

### spec.rules[].rule.or[].and[].tag.matchOptions

`[]string`

How values match (see the dimension leaf; ABSENT matches spend
missing the tag entirely).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.rules[].rule.or[].and[].tag.values

`[]string`

The tag values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.rules[].rule.or[].and[].costCategory

`AwsCostCategoryExpressionCostCategory`

A leaf matching another cost category's values.

### spec.rules[].rule.or[].and[].costCategory.key

`string` · required

The other category's name.

- rule: {"string":{"minLen":"1","maxLen":"50"}}

### spec.rules[].rule.or[].and[].costCategory.matchOptions

`[]string`

How values match (see the dimension leaf).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.rules[].rule.or[].and[].costCategory.values

`[]string`

The category values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.rules[].rule.or[].or

`[]AwsCostCategoryExpressionLeaf`

Leaf children of which AT LEAST ONE must match.

### spec.rules[].rule.or[].or[].dimension

`AwsCostCategoryExpressionDimension`

A dimension leaf.

### spec.rules[].rule.or[].or[].dimension.key

`string`

The dimension to filter on. Cost category rules accept exactly
this server-enforced subset of Cost Explorer dimensions - NOT the
full vocabulary other CE surfaces take (CreateCostCategoryDefinition
names the set in its 400: "Allowed dimension(s): USAGE_TYPE,
RECORD_TYPE, LINKED_ACCOUNT_NAME, SERVICE_CODE, LINKED_ACCOUNT,
BILLING_ENTITY, REGION"; server-verified 2026-08-25). The
by-service intent uses SERVICE_CODE with service CODES
("AmazonS3", "AmazonEC2"), never the SERVICE display-name
dimension ("Amazon Simple Storage Service") that budgets and
anomaly monitors accept.

- rule: {"string":{"in":["BILLING_ENTITY","LINKED_ACCOUNT","LINKED_ACCOUNT_NAME","RECORD_TYPE","REGION","SERVICE_CODE","USAGE_TYPE"]}}

### spec.rules[].rule.or[].or[].dimension.matchOptions

`[]string`

How values match. Unset = EQUALS.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.rules[].rule.or[].or[].dimension.values

`[]string` · required

The values to match.

- rule: {"repeated":{"minItems":"1","items":{"string":{"maxLen":"1024"}}}}

### spec.rules[].rule.or[].or[].tag

`AwsCostCategoryExpressionTag`

A cost-allocation-tag leaf.

### spec.rules[].rule.or[].or[].tag.key

`string`

The tag key.

- rule: {"string":{"maxLen":"1024"}}

### spec.rules[].rule.or[].or[].tag.matchOptions

`[]string`

How values match (see the dimension leaf; ABSENT matches spend
missing the tag entirely).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.rules[].rule.or[].or[].tag.values

`[]string`

The tag values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.rules[].rule.or[].or[].costCategory

`AwsCostCategoryExpressionCostCategory`

A leaf matching another cost category's values.

### spec.rules[].rule.or[].or[].costCategory.key

`string` · required

The other category's name.

- rule: {"string":{"minLen":"1","maxLen":"50"}}

### spec.rules[].rule.or[].or[].costCategory.matchOptions

`[]string`

How values match (see the dimension leaf).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.rules[].rule.or[].or[].costCategory.values

`[]string`

The category values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.rules[].rule.or[].not

`AwsCostCategoryExpressionLeaf`

A leaf child that must NOT match.

### spec.rules[].rule.or[].not.dimension

`AwsCostCategoryExpressionDimension`

A dimension leaf.

### spec.rules[].rule.or[].not.dimension.key

`string`

The dimension to filter on. Cost category rules accept exactly
this server-enforced subset of Cost Explorer dimensions - NOT the
full vocabulary other CE surfaces take (CreateCostCategoryDefinition
names the set in its 400: "Allowed dimension(s): USAGE_TYPE,
RECORD_TYPE, LINKED_ACCOUNT_NAME, SERVICE_CODE, LINKED_ACCOUNT,
BILLING_ENTITY, REGION"; server-verified 2026-08-25). The
by-service intent uses SERVICE_CODE with service CODES
("AmazonS3", "AmazonEC2"), never the SERVICE display-name
dimension ("Amazon Simple Storage Service") that budgets and
anomaly monitors accept.

- rule: {"string":{"in":["BILLING_ENTITY","LINKED_ACCOUNT","LINKED_ACCOUNT_NAME","RECORD_TYPE","REGION","SERVICE_CODE","USAGE_TYPE"]}}

### spec.rules[].rule.or[].not.dimension.matchOptions

`[]string`

How values match. Unset = EQUALS.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.rules[].rule.or[].not.dimension.values

`[]string` · required

The values to match.

- rule: {"repeated":{"minItems":"1","items":{"string":{"maxLen":"1024"}}}}

### spec.rules[].rule.or[].not.tag

`AwsCostCategoryExpressionTag`

A cost-allocation-tag leaf.

### spec.rules[].rule.or[].not.tag.key

`string`

The tag key.

- rule: {"string":{"maxLen":"1024"}}

### spec.rules[].rule.or[].not.tag.matchOptions

`[]string`

How values match (see the dimension leaf; ABSENT matches spend
missing the tag entirely).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.rules[].rule.or[].not.tag.values

`[]string`

The tag values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.rules[].rule.or[].not.costCategory

`AwsCostCategoryExpressionCostCategory`

A leaf matching another cost category's values.

### spec.rules[].rule.or[].not.costCategory.key

`string` · required

The other category's name.

- rule: {"string":{"minLen":"1","maxLen":"50"}}

### spec.rules[].rule.or[].not.costCategory.matchOptions

`[]string`

How values match (see the dimension leaf).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.rules[].rule.or[].not.costCategory.values

`[]string`

The category values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.rules[].rule.not

`AwsCostCategoryExpressionNode`

A child that must NOT match.

### spec.rules[].rule.not.dimension

`AwsCostCategoryExpressionDimension`

A dimension leaf.

### spec.rules[].rule.not.dimension.key

`string`

The dimension to filter on. Cost category rules accept exactly
this server-enforced subset of Cost Explorer dimensions - NOT the
full vocabulary other CE surfaces take (CreateCostCategoryDefinition
names the set in its 400: "Allowed dimension(s): USAGE_TYPE,
RECORD_TYPE, LINKED_ACCOUNT_NAME, SERVICE_CODE, LINKED_ACCOUNT,
BILLING_ENTITY, REGION"; server-verified 2026-08-25). The
by-service intent uses SERVICE_CODE with service CODES
("AmazonS3", "AmazonEC2"), never the SERVICE display-name
dimension ("Amazon Simple Storage Service") that budgets and
anomaly monitors accept.

- rule: {"string":{"in":["BILLING_ENTITY","LINKED_ACCOUNT","LINKED_ACCOUNT_NAME","RECORD_TYPE","REGION","SERVICE_CODE","USAGE_TYPE"]}}

### spec.rules[].rule.not.dimension.matchOptions

`[]string`

How values match. Unset = EQUALS.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.rules[].rule.not.dimension.values

`[]string` · required

The values to match.

- rule: {"repeated":{"minItems":"1","items":{"string":{"maxLen":"1024"}}}}

### spec.rules[].rule.not.tag

`AwsCostCategoryExpressionTag`

A cost-allocation-tag leaf.

### spec.rules[].rule.not.tag.key

`string`

The tag key.

- rule: {"string":{"maxLen":"1024"}}

### spec.rules[].rule.not.tag.matchOptions

`[]string`

How values match (see the dimension leaf; ABSENT matches spend
missing the tag entirely).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.rules[].rule.not.tag.values

`[]string`

The tag values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.rules[].rule.not.costCategory

`AwsCostCategoryExpressionCostCategory`

A leaf matching another cost category's values.

### spec.rules[].rule.not.costCategory.key

`string` · required

The other category's name.

- rule: {"string":{"minLen":"1","maxLen":"50"}}

### spec.rules[].rule.not.costCategory.matchOptions

`[]string`

How values match (see the dimension leaf).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.rules[].rule.not.costCategory.values

`[]string`

The category values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.rules[].rule.not.and

`[]AwsCostCategoryExpressionLeaf`

Leaf children that must ALL match.

### spec.rules[].rule.not.and[].dimension

`AwsCostCategoryExpressionDimension`

A dimension leaf.

### spec.rules[].rule.not.and[].dimension.key

`string`

The dimension to filter on. Cost category rules accept exactly
this server-enforced subset of Cost Explorer dimensions - NOT the
full vocabulary other CE surfaces take (CreateCostCategoryDefinition
names the set in its 400: "Allowed dimension(s): USAGE_TYPE,
RECORD_TYPE, LINKED_ACCOUNT_NAME, SERVICE_CODE, LINKED_ACCOUNT,
BILLING_ENTITY, REGION"; server-verified 2026-08-25). The
by-service intent uses SERVICE_CODE with service CODES
("AmazonS3", "AmazonEC2"), never the SERVICE display-name
dimension ("Amazon Simple Storage Service") that budgets and
anomaly monitors accept.

- rule: {"string":{"in":["BILLING_ENTITY","LINKED_ACCOUNT","LINKED_ACCOUNT_NAME","RECORD_TYPE","REGION","SERVICE_CODE","USAGE_TYPE"]}}

### spec.rules[].rule.not.and[].dimension.matchOptions

`[]string`

How values match. Unset = EQUALS.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.rules[].rule.not.and[].dimension.values

`[]string` · required

The values to match.

- rule: {"repeated":{"minItems":"1","items":{"string":{"maxLen":"1024"}}}}

### spec.rules[].rule.not.and[].tag

`AwsCostCategoryExpressionTag`

A cost-allocation-tag leaf.

### spec.rules[].rule.not.and[].tag.key

`string`

The tag key.

- rule: {"string":{"maxLen":"1024"}}

### spec.rules[].rule.not.and[].tag.matchOptions

`[]string`

How values match (see the dimension leaf; ABSENT matches spend
missing the tag entirely).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.rules[].rule.not.and[].tag.values

`[]string`

The tag values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.rules[].rule.not.and[].costCategory

`AwsCostCategoryExpressionCostCategory`

A leaf matching another cost category's values.

### spec.rules[].rule.not.and[].costCategory.key

`string` · required

The other category's name.

- rule: {"string":{"minLen":"1","maxLen":"50"}}

### spec.rules[].rule.not.and[].costCategory.matchOptions

`[]string`

How values match (see the dimension leaf).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.rules[].rule.not.and[].costCategory.values

`[]string`

The category values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.rules[].rule.not.or

`[]AwsCostCategoryExpressionLeaf`

Leaf children of which AT LEAST ONE must match.

### spec.rules[].rule.not.or[].dimension

`AwsCostCategoryExpressionDimension`

A dimension leaf.

### spec.rules[].rule.not.or[].dimension.key

`string`

The dimension to filter on. Cost category rules accept exactly
this server-enforced subset of Cost Explorer dimensions - NOT the
full vocabulary other CE surfaces take (CreateCostCategoryDefinition
names the set in its 400: "Allowed dimension(s): USAGE_TYPE,
RECORD_TYPE, LINKED_ACCOUNT_NAME, SERVICE_CODE, LINKED_ACCOUNT,
BILLING_ENTITY, REGION"; server-verified 2026-08-25). The
by-service intent uses SERVICE_CODE with service CODES
("AmazonS3", "AmazonEC2"), never the SERVICE display-name
dimension ("Amazon Simple Storage Service") that budgets and
anomaly monitors accept.

- rule: {"string":{"in":["BILLING_ENTITY","LINKED_ACCOUNT","LINKED_ACCOUNT_NAME","RECORD_TYPE","REGION","SERVICE_CODE","USAGE_TYPE"]}}

### spec.rules[].rule.not.or[].dimension.matchOptions

`[]string`

How values match. Unset = EQUALS.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.rules[].rule.not.or[].dimension.values

`[]string` · required

The values to match.

- rule: {"repeated":{"minItems":"1","items":{"string":{"maxLen":"1024"}}}}

### spec.rules[].rule.not.or[].tag

`AwsCostCategoryExpressionTag`

A cost-allocation-tag leaf.

### spec.rules[].rule.not.or[].tag.key

`string`

The tag key.

- rule: {"string":{"maxLen":"1024"}}

### spec.rules[].rule.not.or[].tag.matchOptions

`[]string`

How values match (see the dimension leaf; ABSENT matches spend
missing the tag entirely).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.rules[].rule.not.or[].tag.values

`[]string`

The tag values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.rules[].rule.not.or[].costCategory

`AwsCostCategoryExpressionCostCategory`

A leaf matching another cost category's values.

### spec.rules[].rule.not.or[].costCategory.key

`string` · required

The other category's name.

- rule: {"string":{"minLen":"1","maxLen":"50"}}

### spec.rules[].rule.not.or[].costCategory.matchOptions

`[]string`

How values match (see the dimension leaf).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.rules[].rule.not.or[].costCategory.values

`[]string`

The category values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.rules[].rule.not.not

`AwsCostCategoryExpressionLeaf`

A leaf child that must NOT match.

### spec.rules[].rule.not.not.dimension

`AwsCostCategoryExpressionDimension`

A dimension leaf.

### spec.rules[].rule.not.not.dimension.key

`string`

The dimension to filter on. Cost category rules accept exactly
this server-enforced subset of Cost Explorer dimensions - NOT the
full vocabulary other CE surfaces take (CreateCostCategoryDefinition
names the set in its 400: "Allowed dimension(s): USAGE_TYPE,
RECORD_TYPE, LINKED_ACCOUNT_NAME, SERVICE_CODE, LINKED_ACCOUNT,
BILLING_ENTITY, REGION"; server-verified 2026-08-25). The
by-service intent uses SERVICE_CODE with service CODES
("AmazonS3", "AmazonEC2"), never the SERVICE display-name
dimension ("Amazon Simple Storage Service") that budgets and
anomaly monitors accept.

- rule: {"string":{"in":["BILLING_ENTITY","LINKED_ACCOUNT","LINKED_ACCOUNT_NAME","RECORD_TYPE","REGION","SERVICE_CODE","USAGE_TYPE"]}}

### spec.rules[].rule.not.not.dimension.matchOptions

`[]string`

How values match. Unset = EQUALS.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.rules[].rule.not.not.dimension.values

`[]string` · required

The values to match.

- rule: {"repeated":{"minItems":"1","items":{"string":{"maxLen":"1024"}}}}

### spec.rules[].rule.not.not.tag

`AwsCostCategoryExpressionTag`

A cost-allocation-tag leaf.

### spec.rules[].rule.not.not.tag.key

`string`

The tag key.

- rule: {"string":{"maxLen":"1024"}}

### spec.rules[].rule.not.not.tag.matchOptions

`[]string`

How values match (see the dimension leaf; ABSENT matches spend
missing the tag entirely).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.rules[].rule.not.not.tag.values

`[]string`

The tag values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.rules[].rule.not.not.costCategory

`AwsCostCategoryExpressionCostCategory`

A leaf matching another cost category's values.

### spec.rules[].rule.not.not.costCategory.key

`string` · required

The other category's name.

- rule: {"string":{"minLen":"1","maxLen":"50"}}

### spec.rules[].rule.not.not.costCategory.matchOptions

`[]string`

How values match (see the dimension leaf).

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EQUALS","ABSENT","STARTS_WITH","ENDS_WITH","CONTAINS","GREATER_THAN_OR_EQUAL","CASE_SENSITIVE","CASE_INSENSITIVE"]}}}}

### spec.rules[].rule.not.not.costCategory.values

`[]string`

The category values to match.

- rule: {"repeated":{"items":{"string":{"maxLen":"1024"}}}}

### spec.rules[].inheritedValue

`AwsCostCategoryInheritedValue`

For INHERITED_VALUE rules: the dimension whose own value becomes
the category value.

### spec.rules[].inheritedValue.dimensionName

`string`

Where the value comes from: the linked account's name, or a
cost-allocation tag's value.

- rule: {"string":{"in":["LINKED_ACCOUNT_NAME","TAG"]}}

### spec.rules[].inheritedValue.dimensionKey

`string`

For TAG: the tag key whose value is inherited. AWS ignores it for
LINKED_ACCOUNT_NAME.

- rule: {"string":{"maxLen":"1024"}}

### spec.splitChargeRules

`[]AwsCostCategorySplitChargeRule`

Split-charge rules: re-allocate one value's costs across target
values, proportionally, evenly, or by fixed percentages.

- rule: method FIXED needs a parameter of type ALLOCATION_PERCENTAGES with one percentage per target

### spec.splitChargeRules[].source

`string` · required

The category value whose costs are split away (e.g. a shared
"Platform" value).

- rule: {"string":{"minLen":"1","maxLen":"1024"}}

### spec.splitChargeRules[].targets

`[]string` · required

The category values receiving the split costs (1-500).

- rule: {"repeated":{"minItems":"1","maxItems":"500","items":{"string":{"maxLen":"1024"}}}}

### spec.splitChargeRules[].method

`string`

How costs split: PROPORTIONAL to each target's own costs, EVEN
across targets, or FIXED percentages (see parameters).

- rule: {"string":{"in":["FIXED","PROPORTIONAL","EVEN"]}}

### spec.splitChargeRules[].parameters

`[]AwsCostCategorySplitChargeParameter`

Split parameters - today only ALLOCATION_PERCENTAGES for the
FIXED method (one percentage per target, in target order,
summing to 100).

### spec.splitChargeRules[].parameters[].type

`string`

The parameter type ("ALLOCATION_PERCENTAGES" is the only value
the API accepts today).

- rule: {"string":{"in":["ALLOCATION_PERCENTAGES"]}}

### spec.splitChargeRules[].parameters[].values

`[]string` · required

The percentage values, one per target in target order, summing
to 100 (e.g. ["60", "40"]).

- rule: {"repeated":{"minItems":"1","maxItems":"500"}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsCostCategory, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.category_arn` | `string` | The category's ARN (also the provider's import ID). |
| `status.outputs.category_name` | `string` | The category's name (the key other expressions reference the category by). |
| `status.outputs.effective_start` | `string` | When the category's rules take effect (AWS-normalized month start). |
| `status.outputs.effective_end` | `string` | When the category stops applying (set by AWS on deletion; normally empty). |

## See Also

- [Overview](../README.md)

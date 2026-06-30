# AwsWafWebAcl Terraform Module

Terraform IaC module for deploying an AWS WAFv2 Web ACL.

## Usage

```hcl
module "waf_web_acl" {
  source = "./iac/tf"

  metadata = {
    name = "my-web-acl"
    org  = "my-org"
    env  = "production"
    id   = "awswaf-abc123"
  }

  spec = {
    scope = "REGIONAL"
    default_action = {
      type = "allow"
    }
    rules = [
      {
        name     = "aws-common-rules"
        priority = 1
        override_action = "none"
        managed_rule_group = {
          name        = "AWSManagedRulesCommonRuleSet"
          vendor_name = "AWS"
        }
      }
    ]
  }
}
```

## Inputs

| Variable | Type | Description |
|----------|------|-------------|
| `metadata` | any | Resource metadata (name, org, env, id, labels) |
| `spec` | any | AwsWafWebAclSpec configuration |

## Outputs

| Output | Description |
|--------|-------------|
| `web_acl_arn` | Web ACL ARN |
| `web_acl_id` | Web ACL ID |
| `web_acl_name` | Web ACL name |
| `capacity` | WCUs consumed |

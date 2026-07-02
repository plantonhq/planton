locals {
  # Resource-identity tags, matching the Pulumi module key-for-key.
  aws_tags = {
    "Name"                     = var.metadata.name
    "planton.ai/resource"      = "true"
    "planton.ai/organization"  = var.metadata.org
    "planton.ai/environment"   = var.metadata.env
    "planton.ai/resource-kind" = "AwsIamRole"
    "planton.ai/resource-id"   = var.metadata.id
  }

  # trust_policy is a free-form JSON object (google.protobuf.Struct); aws_iam_role
  # wants assume_role_policy as a JSON string, so encode the object here.
  trust_policy_json = jsonencode(var.spec.trust_policy)

  # inline_policies is free-form JSON (map<string, google.protobuf.Struct>), typed `any` in
  # variables.tf because its entries have heterogeneous shapes. Encode each policy document to a
  # JSON string here so the result is a homogeneous map(string): aws_iam_role_policy.for_each
  # accepts a map/set, and converting a heterogeneous object to a map would otherwise fail with
  # "all map elements must have the same type".
  inline_policies_json = {
    for k, v in var.spec.inline_policies : k => jsonencode(v)
  }
}

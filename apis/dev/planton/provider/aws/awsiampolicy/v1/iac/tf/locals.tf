locals {
  # Resource-identity tags, matching the Pulumi module key-for-key.
  aws_tags = {
    "Name"                     = var.metadata.name
    "planton.ai/resource"      = "true"
    "planton.ai/organization"  = var.metadata.org
    "planton.ai/environment"   = var.metadata.env
    "planton.ai/resource-kind" = "AwsIamPolicy"
    "planton.ai/resource-id"   = var.metadata.id
  }

  # policy_document is a free-form JSON object (google.protobuf.Struct);
  # aws_iam_policy wants the document as a JSON string, so encode it here.
  policy_document_json = jsonencode(var.spec.policy_document)
}

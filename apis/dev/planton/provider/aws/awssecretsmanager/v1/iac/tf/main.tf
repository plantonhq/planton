# AWS Secrets Manager - Secret Container Creation
# 
# This module creates secret CONTAINERS in AWS Secrets Manager following the Planton 80/20 philosophy:
# - Creates the secret infrastructure (name, encryption, tags) via IaC
# - Does NOT store actual secret values in Terraform (security best practice)
# - Secret values are populated separately via CI/CD pipelines, External Secrets Operator, or manual processes
#
# Each secret is created with:
# - AWS-managed encryption key (aws/secretsmanager) for secure storage
# - 30-day recovery window for accidental deletion protection
# - CloudTrail audit logging enabled automatically
# - Resource tags for governance and cost allocation
#
# The placeholder value seeds the secret so it exists in AWS, but is immediately ignored via lifecycle policy.
# Applications should never rely on this placeholder - real values come from secure secret management workflows.

# Create AWS Secrets Manager secrets (containers only, no actual secret values)
resource "aws_secretsmanager_secret" "secrets" {
  # Iterate over each secret name provided in the spec
  # Creates a separate AWS secret resource for each logical secret name
  for_each = {
    for name in local.secret_names :
    name => name
  }

  # Construct unique secret name: {resource-id}-{secret-name}
  # Example: "myapp-prod-secrets-abc123-DB_PASSWORD"
  # This prevents naming collisions across different AwsSecretsManager resources
  name = format("%s-%s", local.resource_id, each.value)

  # Apply Planton standard tags plus user-provided tags
  # Tags enable cost tracking, resource governance, and environment identification
  tags = local.final_tags

  # Note: We do NOT specify kms_key_id, which means AWS uses the default
  # AWS-managed key (aws/secretsmanager) for encryption. This is secure and
  # cost-effective for most use cases. For custom KMS keys, that's a 20% use case
  # handled outside this minimal implementation.

  # Note: We do NOT configure rotation here. Secret rotation requires:
  # 1. A Lambda function that knows how to rotate the specific credential type
  # 2. Testing and validation of the rotation process
  # 3. Application compatibility with rotated credentials
  # These are deployed separately after the secret containers are created.
}

# Seed secrets with placeholder values to make them valid AWS secrets
# The placeholder ensures the secret has at least one version, which AWS requires
resource "aws_secretsmanager_secret_version" "secrets" {
  # Create one secret version for each secret container
  for_each = {
    for name in local.secret_names :
    name => name
  }

  # Reference the secret created above
  secret_id = aws_secretsmanager_secret.secrets[each.key].id

  # Placeholder value - this will be REPLACED by real values from external sources
  # Common patterns for populating real values:
  # 1. CI/CD pipeline: aws secretsmanager put-secret-value after infrastructure deployment
  # 2. External Secrets Operator: Syncs from HashiCorp Vault or other secret stores
  # 3. Manual: Security engineer sets via AWS Console/CLI for high-sensitivity secrets
  secret_string = "placeholder"

  # Critical: Ignore changes to secret_string after initial creation
  # This prevents Terraform from:
  # - Overwriting real secret values with the placeholder on subsequent applies
  # - Detecting drift when values are updated outside Terraform
  # - Storing actual secret values in Terraform state (major security risk)
  #
  # Once the real value is set externally, Terraform treats this resource as immutable
  lifecycle {
    ignore_changes = [secret_string]
  }
}

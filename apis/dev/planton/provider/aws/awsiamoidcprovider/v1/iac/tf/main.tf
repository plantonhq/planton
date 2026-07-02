# An IAM OIDC provider is the trust anchor for keyless federation: it lets an
# external issuer's short-lived tokens be exchanged for AWS credentials via
# STS AssumeRoleWithWebIdentity. The URL is create-only (changing it replaces
# the provider) and AWS allows one provider per unique URL per account.
resource "aws_iam_openid_connect_provider" "this" {
  url            = var.spec.url
  client_id_list = var.spec.client_id_list

  # null (not an empty list) when unset -- see locals.thumbprint_list. AWS then
  # derives the thumbprint from its trusted CA store. Note the AWS quirk:
  # once thumbprints are set they cannot be cleared in place (the update API
  # rejects an empty list); going back to derived thumbprints replaces the provider.
  thumbprint_list = local.thumbprint_list

  tags = local.aws_tags
}

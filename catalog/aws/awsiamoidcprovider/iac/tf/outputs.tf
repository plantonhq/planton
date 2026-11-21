output "provider_arn" {
  description = "The ARN of the IAM OIDC provider (referenced as a Federated principal in IAM role trust policies)."
  value       = aws_iam_openid_connect_provider.this.arn
}

output "provider_url" {
  description = "The issuer URL AWS stored for this provider, with the https:// scheme stripped."
  # The AWS provider hands the url back as it was configured, scheme included,
  # while IAM's condition keys are the bare issuer (<host>/<path>:sub), so the
  # scheme is stripped here for a role trust that references this output.
  value = replace(aws_iam_openid_connect_provider.this.url, "/^https?:\\/\\//", "")
}

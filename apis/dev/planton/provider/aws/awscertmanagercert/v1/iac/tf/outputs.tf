output "cert_arn" {
  description = "The Amazon Resource Name (ARN) of the created ACM certificate"
  value       = aws_acm_certificate.this.arn
}

output "certificate_domain_name" {
  description = "The primary domain name for which the certificate was issued"
  value       = aws_acm_certificate.this.domain_name
}



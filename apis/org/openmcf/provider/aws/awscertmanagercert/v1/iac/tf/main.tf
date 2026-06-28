# AWS Certificate Manager certificate resource
resource "aws_acm_certificate" "this" {
  domain_name       = var.spec.primary_domain_name
  validation_method = upper(try(var.spec.validation_method, "DNS"))

  # Add alternate domain names if provided
  subject_alternative_names = try(var.spec.alternate_domain_names, [])

  # Ensure proper lifecycle management for certificate updates
  lifecycle {
    create_before_destroy = true
  }

  tags = {
    Name = var.metadata.name
    # Add other metadata tags if needed
  }
}

# DNS validation records (only when DNS validation is selected).
# A domain and its wildcard SAN (e.g. app.example.com and *.app.example.com) share
# the SAME ACM validation CNAME, so this for_each emits two entries with an identical
# record name. allow_overwrite makes each an UPSERT instead of a CREATE, so the
# duplicate is idempotent and a record left behind by a prior partial apply is adopted
# rather than colliding ("InvalidChangeBatch ... already exists").
resource "aws_route53_record" "validation" {
  for_each = local.is_dns_validation ? {
    for dvo in aws_acm_certificate.this.domain_validation_options : dvo.domain_name => {
      name   = dvo.resource_record_name
      type   = dvo.resource_record_type
      value  = dvo.resource_record_value
      zone_id = local.safe_route53_zone_id
    }
  } : {}

  allow_overwrite = true
  zone_id         = each.value.zone_id
  name            = each.value.name
  type            = each.value.type
  ttl             = 60
  records         = [each.value.value]

  depends_on = [aws_acm_certificate.this]
}

# Certificate validation (only for DNS validation)
resource "aws_acm_certificate_validation" "this" {
  count = local.is_dns_validation ? 1 : 0

  certificate_arn         = aws_acm_certificate.this.arn
  validation_record_fqdns = [for record in aws_route53_record.validation : record.fqdn]

  depends_on = [aws_route53_record.validation]
}



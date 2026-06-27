locals {
  # Route53 hosted zone ID. The generator flattens StringValueOrRef to its
  # resolved string (the orchestrator resolves any value_from before the module
  # runs), so the value is consumed directly.
  safe_route53_zone_id = var.spec.route53_hosted_zone_id

  # Boolean for DNS validation method
  is_dns_validation = upper(try(var.spec.validation_method, "DNS")) == "DNS"

  # Boolean for email validation method
  is_email_validation = upper(try(var.spec.validation_method, "DNS")) == "EMAIL"

  # Check if alternate domain names are provided
  has_alternate_domains = length(try(var.spec.alternate_domain_names, [])) > 0
}



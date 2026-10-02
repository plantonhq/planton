locals {
  # Resource-identity tags match the Pulumi module key-for-key. ACM
  # certificates have no AWS name -- metadata.name drives the Name tag and
  # consumers address the certificate through its ARN.
  aws_tags = {
    "Name"                     = var.metadata.name
    "planton.ai/resource"      = "true"
    "planton.ai/organization"  = var.metadata.org
    "planton.ai/environment"   = var.metadata.env
    "planton.ai/resource-kind" = "AwsCertManagerCert"
    "planton.ai/resource-id"   = var.metadata.id
  }

  # Creation mode. Exactly one of these is true (CEL enforces the
  # exclusivity at validation time): a domain selects requested or private
  # issuance; imported material selects import.
  is_imported  = var.spec.imported != null
  is_private   = !local.is_imported && var.spec.certificate_authority_arn != ""
  is_requested = !local.is_imported && !local.is_private

  # Requested certificates validate via DNS unless EMAIL is chosen.
  # Imported and private certificates never validate publicly.
  validation_method = local.is_requested ? coalesce(var.spec.validation_method == "" ? null : var.spec.validation_method, "DNS") : null
  is_dns_validation = local.is_requested && local.validation_method == "DNS"

  # The generator flattens StringValueOrRef to its resolved string (the
  # orchestrator resolves any value_from before the module runs). A
  # non-empty zone ID turns on managed validation: the module creates the
  # validation CNAMEs and (by default) waits for issuance.
  route53_zone_id            = var.spec.route53_hosted_zone_id
  manages_validation_records = local.is_dns_validation && local.route53_zone_id != ""

  # domain_validation_options exists only for requested certificates.
  # for_each keys must be known at plan time, so this keys by domain name
  # (config-known) -- the record name/type/value are computed and land in
  # the map VALUES, which for_each tolerates. A domain and its wildcard
  # SAN (e.g. app.example.com and *.app.example.com) share the SAME
  # validation CNAME, producing two entries with an identical record name;
  # allow_overwrite on the record resource turns the duplicate into an
  # idempotent UPSERT.
  validation_records = local.manages_validation_records ? {
    for dvo in aws_acm_certificate.this.domain_validation_options : dvo.domain_name => {
      name  = dvo.resource_record_name
      type  = dvo.resource_record_type
      value = dvo.resource_record_value
    }
  } : {}
}

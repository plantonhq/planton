# An ACM certificate in one of the three creation modes: requested
# (Amazon-issued, DNS or EMAIL validated), imported (bring-your-own
# material), or private (issued by an ACM-PCA authority). The mode is
# derived in locals.tf from which spec arm is populated; CEL keeps the
# arms mutually exclusive so this resource never sees a mixed shape.
resource "aws_acm_certificate" "this" {
  # Requested + private arms: ACM issues for these domains. Imported
  # certificates derive their domains from the certificate body, so the
  # domain arguments must stay null in that mode.
  domain_name               = local.is_imported ? null : var.spec.primary_domain_name
  subject_alternative_names = local.is_imported ? null : var.spec.alternate_domain_names

  # Only requested (public) certificates validate ownership. Private-CA
  # issuance is authorized by the CA itself; passing a validation method
  # alongside certificate_authority_arn is rejected by the provider.
  validation_method = local.is_requested ? local.validation_method : null

  # Private (ACM-PCA) arm.
  certificate_authority_arn = local.is_private ? var.spec.certificate_authority_arn : null

  # Managed early renewal -- a private-CA mechanism (ACM's
  # RenewCertificate is private-certificate-only; CEL ties the field to
  # the private arm). Durations under 60 days have no effect.
  early_renewal_duration = local.is_private && var.spec.early_renewal_duration != "" ? var.spec.early_renewal_duration : null

  # Imported arm: the PEM material. The private key is sensitive spec
  # input and never appears in outputs; re-importing new material before
  # expiry updates in place and keeps the ARN stable for consumers.
  certificate_body = local.is_imported ? var.spec.imported.certificate_body : null
  private_key      = local.is_imported ? var.spec.imported.private_key : null
  # HCL && is not short-circuiting -- never dereference imported when the
  # arm is absent (the session-008 class).
  certificate_chain = local.is_imported ? (try(var.spec.imported.certificate_chain, "") != "" ? var.spec.imported.certificate_chain : null) : null

  # Create-time immutable key algorithm for ACM-issued certificates.
  # Empty keeps the ACM default (RSA_2048).
  key_algorithm = !local.is_imported && var.spec.key_algorithm != "" ? var.spec.key_algorithm : null

  # Per-domain overrides of where the validation request is sent (e.g.
  # EMAIL-validating a subdomain at its parent domain).
  dynamic "validation_option" {
    for_each = local.is_requested ? var.spec.validation_options : []
    content {
      domain_name       = validation_option.value.domain_name
      validation_domain = validation_option.value.validation_domain
    }
  }

  # Certificate Transparency logging and exportability. Both empty-string
  # sentinels keep the ACM defaults (CT ENABLED, export DISABLED).
  dynamic "options" {
    for_each = var.spec.options != null ? [var.spec.options] : []
    content {
      certificate_transparency_logging_preference = options.value.certificate_transparency_logging_preference != "" ? options.value.certificate_transparency_logging_preference : null
      export                                      = options.value.export != "" ? options.value.export : null
    }
  }

  # A certificate swap must bring the replacement up before releasing the
  # old ARN -- consumers (listeners, CloudFront) hold a reference to it.
  lifecycle {
    create_before_destroy = true
  }

  tags = local.aws_tags
}

# Managed DNS validation: the validation CNAMEs land in the referenced
# Route53 zone. Keyed by record name -- a domain and its wildcard SAN
# share one CNAME, so this deduplicates the pair. allow_overwrite makes
# each an UPSERT instead of a CREATE, so a record left behind by a prior
# partial apply (or shared across certificates) is adopted rather than
# colliding ("InvalidChangeBatch ... already exists").
resource "aws_route53_record" "validation" {
  for_each = local.validation_records

  allow_overwrite = true
  zone_id         = local.route53_zone_id
  name            = each.value.name
  type            = each.value.type
  ttl             = 60
  records         = [each.value.value]
}

# The issuance waiter -- a read-only resource that blocks until ACM
# reports ISSUED (75-minute ceiling; DNS-validated issuance typically
# lands in minutes once the records propagate). Only created when the
# module manages the validation records AND the spec asks to wait; when
# the zone is external, the certificate rests in PENDING_VALIDATION and
# the records to create are exported instead.
resource "aws_acm_certificate_validation" "this" {
  count = local.manages_validation_records && coalesce(var.spec.wait_for_validation, true) ? 1 : 0

  certificate_arn         = aws_acm_certificate.this.arn
  validation_record_fqdns = [for record in aws_route53_record.validation : record.fqdn]
}

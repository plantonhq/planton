# -----------------------------------------------------------------------------
# App Runner service
#
# The service is the runtime: source in (image or code), HTTPS endpoint out.
# Its shared companions -- auto scaling configuration, VPC connector, and
# observability configuration -- are separate first-class resources referenced
# by ARN, never created here: each is designed by AWS to be shared across
# services, so embedding one per service would fork what should be tuned in
# one place.
# -----------------------------------------------------------------------------
resource "aws_apprunner_service" "this" {
  service_name = local.service_name

  # --- Source -----------------------------------------------------------
  # Exactly one of image_repository / code_repository is present (spec CEL).
  # Port, start command, and the env var/secret maps live at the spec top
  # level and are routed into whichever arm is active -- they configure the
  # runtime container either way.
  source_configuration {
    # Sent explicitly (never defaulted) because the honest default is
    # conditional on the source: AWS enables auto-deploy for code repos and
    # private ECR only, and REJECTS it for ECR_PUBLIC (spec CEL guards the
    # invalid combination before AWS would).
    auto_deployments_enabled = var.spec.auto_deployments_enabled

    dynamic "authentication_configuration" {
      for_each = local.needs_auth_config ? [1] : []
      content {
        # Private ECR pulls use the access role; code repositories use the
        # out-of-band App Runner connection. The two never coexist because
        # the source arms are mutually exclusive.
        access_role_arn = var.spec.image_source != null ? (
          var.spec.image_source.access_role_arn != "" ? var.spec.image_source.access_role_arn : null
        ) : null
        connection_arn = var.spec.code_source != null ? var.spec.code_source.connection_arn : null
      }
    }

    dynamic "image_repository" {
      for_each = var.spec.image_source != null ? [var.spec.image_source] : []
      content {
        image_identifier      = image_repository.value.image_identifier
        image_repository_type = image_repository.value.image_repository_type

        image_configuration {
          port                          = var.spec.port
          start_command                 = var.spec.start_command != "" ? var.spec.start_command : null
          runtime_environment_variables = length(var.spec.environment_variables) > 0 ? var.spec.environment_variables : null
          runtime_environment_secrets   = length(var.spec.environment_secrets) > 0 ? var.spec.environment_secrets : null
        }
      }
    }

    dynamic "code_repository" {
      for_each = var.spec.code_source != null ? [var.spec.code_source] : []
      content {
        repository_url   = code_repository.value.repository_url
        source_directory = code_repository.value.source_directory != "" ? code_repository.value.source_directory : null

        # BRANCH is the only source-code-version type AWS supports; the spec
        # models the branch directly rather than a one-value enum wrapper.
        source_code_version {
          type  = "BRANCH"
          value = code_repository.value.branch
        }

        code_configuration {
          configuration_source = code_repository.value.configuration_source

          # Build settings apply only in API mode; in REPOSITORY mode App
          # Runner reads apprunner.yaml from the source directory and AWS
          # rejects inline values.
          dynamic "code_configuration_values" {
            for_each = code_repository.value.configuration_source == "API" ? [1] : []
            content {
              runtime                       = code_repository.value.runtime
              build_command                 = code_repository.value.build_command != "" ? code_repository.value.build_command : null
              start_command                 = var.spec.start_command != "" ? var.spec.start_command : null
              port                          = var.spec.port
              runtime_environment_variables = length(var.spec.environment_variables) > 0 ? var.spec.environment_variables : null
              runtime_environment_secrets   = length(var.spec.environment_secrets) > 0 ? var.spec.environment_secrets : null
            }
          }
        }
      }
    }
  }

  # --- Instances ----------------------------------------------------------
  instance_configuration {
    cpu               = var.spec.cpu
    memory            = var.spec.memory
    instance_role_arn = var.spec.instance_role_arn != "" ? var.spec.instance_role_arn : null
  }

  # --- Health check -------------------------------------------------------
  # Only emitted when the spec configures it; AWS then applies TCP checks on
  # the service port with its own defaults. The path is meaningful only for
  # HTTP -- sending it alongside TCP would be silently ignored, so it is
  # nulled deliberately.
  dynamic "health_check_configuration" {
    for_each = var.spec.health_check != null ? [var.spec.health_check] : []
    content {
      protocol            = health_check_configuration.value.protocol
      path                = health_check_configuration.value.protocol == "HTTP" ? health_check_configuration.value.path : null
      interval            = health_check_configuration.value.interval
      timeout             = health_check_configuration.value.timeout
      healthy_threshold   = health_check_configuration.value.healthy_threshold
      unhealthy_threshold = health_check_configuration.value.unhealthy_threshold
    }
  }

  # --- Networking ---------------------------------------------------------
  # Always sent explicitly so the deployed shape never depends on AWS-side
  # defaults: egress routes through the referenced VPC connector when one is
  # set, ingress publicness and address family mirror the spec (middleware
  # materializes the spec defaults; the null-guards keep direct module use
  # deterministic too).
  network_configuration {
    egress_configuration {
      egress_type       = local.egress_type
      vpc_connector_arn = var.spec.vpc_connector_arn != "" ? var.spec.vpc_connector_arn : null
    }
    ingress_configuration {
      is_publicly_accessible = var.spec.is_publicly_accessible != null ? var.spec.is_publicly_accessible : true
    }
    ip_address_type = var.spec.ip_address_type
  }

  # --- Encryption (ForceNew) ------------------------------------------------
  dynamic "encryption_configuration" {
    for_each = var.spec.kms_key_arn != "" ? [1] : []
    content {
      kms_key = var.spec.kms_key_arn
    }
  }

  # --- Observability --------------------------------------------------------
  # Presence of the configuration reference IS the enable switch -- there is
  # no separate toggle to drift out of sync. One-way on a live service
  # (upstream provider gap): the provider sends NOTHING for an absent block
  # on update, so removing the reference leaves tracing running at AWS and
  # the plan never converges. Rendering a disabled block here instead is NOT
  # safe: AWS echoes NO observability member at all for a tracing-less
  # service (live DescribeService evidence, never {ObservabilityEnabled:
  # false}) and the provider flattens that nil to an empty block, so an
  # always-present disabled block perpetually diffs on every tracing-less
  # service. Disable requires service replacement until the provider gains
  # a disable path.
  dynamic "observability_configuration" {
    for_each = var.spec.observability_configuration_arn != "" ? [1] : []
    content {
      observability_enabled           = true
      observability_configuration_arn = var.spec.observability_configuration_arn
    }
  }

  # --- Auto scaling -----------------------------------------------------------
  # Null falls back to the account's default auto scaling configuration.
  auto_scaling_configuration_arn = var.spec.auto_scaling_configuration_arn != "" ? var.spec.auto_scaling_configuration_arn : null

  tags = local.aws_tags
}

# -----------------------------------------------------------------------------
# Custom domain associations
#
# One association per spec entry, keyed by domain name so entries add and
# remove independently. App Runner issues the TLS certificate; the
# per-domain validation CNAMEs surface as outputs for external DNS
# (or AwsRoute53DnsRecord composition). The association resource returns as
# soon as validation records are AVAILABLE -- it deliberately does not wait
# for the domain to go active, because that requires the DNS records this
# module does not manage.
# -----------------------------------------------------------------------------
resource "aws_apprunner_custom_domain_association" "this" {
  for_each = local.custom_domains

  domain_name          = each.value.domain_name
  service_arn          = aws_apprunner_service.this.arn
  enable_www_subdomain = each.value.enable_www_subdomain != null ? each.value.enable_www_subdomain : true
}

# -----------------------------------------------------------------------------
# WAF association
#
# The protected resource points at the web ACL (the same direction CloudFront
# models it) -- the association is glue with no identity of its own, so it
# folds here rather than existing as a kind.
# -----------------------------------------------------------------------------
resource "aws_wafv2_web_acl_association" "this" {
  count = var.spec.web_acl_arn != "" ? 1 : 0

  resource_arn = aws_apprunner_service.this.arn
  web_acl_arn  = var.spec.web_acl_arn
}

# -----------------------------------------------------------------------------
# VPC Ingress Connections
#
# One connection per spec entry, keyed by name so entries add and remove
# independently. Each publishes the service into one VPC through an
# interface VPC endpoint (AWS PrivateLink); clients inside that VPC resolve
# the per-connection domain_name exported below. Every value is create-time
# immutable by AWS design -- changing an entry replaces that one connection.
# -----------------------------------------------------------------------------
resource "aws_apprunner_vpc_ingress_connection" "this" {
  for_each = local.vpc_ingress_connections

  name        = each.value.name
  service_arn = aws_apprunner_service.this.arn

  ingress_vpc_configuration {
    vpc_id          = each.value.vpc_id
    vpc_endpoint_id = each.value.vpc_endpoint_id
  }

  tags = local.aws_tags
}

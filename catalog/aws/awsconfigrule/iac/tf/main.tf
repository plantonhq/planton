# One AWS Config rule - managed, custom-lambda, or custom-policy;
# account- or organization-scoped - plus optional auto-remediation.
#
# The spec's CELs guarantee exactly one source arm arrives here and
# that org-only / account-only surfaces never mix, so the renders
# below branch on presence without re-validating.
#
# Lifecycle facts the renders below depend on:
#   - organization rules are DIFFERENT provider resources (one per
#     source kind), deployed from the management or delegated-admin
#     account; member accounts receive them automatically;
#   - a custom-lambda rule needs config.amazonaws.com invoke
#     permission on the function BEFORE create (the consumer's
#     contract on the Lambda's policy);
#   - the remediation configuration attaches by RULE NAME - the
#     resource reference below wires create order and teardown order;
#   - AWS caps organization rule names at 64 characters (account rules
#     at 128); metadata.name carries the rule name on both engines.

locals {
  # The organization block's switch (on by default once declared) decides
  # the scope; a null switch reads as on.
  is_org = var.spec.organization != null && coalesce(var.spec.organization.enabled, true)

  # The account-scoped rule's source mapping.
  source_owner = var.spec.managed != null ? "AWS" : (var.spec.custom_lambda != null ? "CUSTOM_LAMBDA" : "CUSTOM_POLICY")
  source_identifier = var.spec.managed != null ? var.spec.managed.rule_identifier : (
    var.spec.custom_lambda != null ? var.spec.custom_lambda.function_arn : null
  )
}

# ----- account-scoped rule -----

resource "aws_config_config_rule" "this" {
  count = local.is_org ? 0 : 1

  name        = var.metadata.name
  description = var.spec.description != "" ? var.spec.description : null

  input_parameters            = var.spec.input_parameters != "" ? var.spec.input_parameters : null
  maximum_execution_frequency = var.spec.maximum_execution_frequency != "" ? var.spec.maximum_execution_frequency : null

  source {
    owner             = local.source_owner
    source_identifier = local.source_identifier

    dynamic "source_detail" {
      for_each = var.spec.custom_lambda != null ? var.spec.custom_lambda.source_details : []
      content {
        message_type                = source_detail.value.message_type
        maximum_execution_frequency = source_detail.value.maximum_execution_frequency != "" ? source_detail.value.maximum_execution_frequency : null
      }
    }

    # Guard rules evaluate on configuration changes; AWS requires the
    # trigger detail explicitly on custom-policy sources, so both
    # engines always send it (never user-facing - the derived
    # discriminator pattern).
    dynamic "source_detail" {
      for_each = var.spec.custom_policy != null ? [1] : []
      content {
        message_type = "ConfigurationItemChangeNotification"
      }
    }

    dynamic "custom_policy_details" {
      for_each = var.spec.custom_policy != null ? [var.spec.custom_policy] : []
      content {
        policy_runtime            = custom_policy_details.value.policy_runtime
        policy_text               = custom_policy_details.value.policy_text
        enable_debug_log_delivery = custom_policy_details.value.enable_debug_log_delivery
      }
    }
  }

  dynamic "scope" {
    for_each = var.spec.scope != null ? [var.spec.scope] : []
    content {
      compliance_resource_id    = scope.value.compliance_resource_id != "" ? scope.value.compliance_resource_id : null
      compliance_resource_types = length(scope.value.compliance_resource_types) > 0 ? scope.value.compliance_resource_types : null
      tag_key                   = scope.value.tag_key != "" ? scope.value.tag_key : null
      tag_value                 = scope.value.tag_value != "" ? scope.value.tag_value : null
    }
  }

  dynamic "evaluation_mode" {
    for_each = var.spec.evaluation_modes
    content {
      mode = evaluation_mode.value
    }
  }

  tags = local.aws_tags
}

# ----- organization-scoped rules (one resource per source kind) -----

resource "aws_config_organization_managed_rule" "this" {
  count = local.is_org && var.spec.managed != null ? 1 : 0

  name            = var.metadata.name
  rule_identifier = var.spec.managed.rule_identifier
  description     = var.spec.description != "" ? var.spec.description : null

  input_parameters            = var.spec.input_parameters != "" ? var.spec.input_parameters : null
  maximum_execution_frequency = var.spec.maximum_execution_frequency != "" ? var.spec.maximum_execution_frequency : null
  excluded_accounts           = length(var.spec.organization.excluded_accounts) > 0 ? var.spec.organization.excluded_accounts : null

  resource_id_scope    = var.spec.scope != null && var.spec.scope.compliance_resource_id != "" ? var.spec.scope.compliance_resource_id : null
  resource_types_scope = var.spec.scope != null && length(var.spec.scope.compliance_resource_types) > 0 ? var.spec.scope.compliance_resource_types : null
  tag_key_scope        = var.spec.scope != null && var.spec.scope.tag_key != "" ? var.spec.scope.tag_key : null
  tag_value_scope      = var.spec.scope != null && var.spec.scope.tag_value != "" ? var.spec.scope.tag_value : null
}

resource "aws_config_organization_custom_rule" "this" {
  count = local.is_org && var.spec.custom_lambda != null ? 1 : 0

  name                = var.metadata.name
  lambda_function_arn = var.spec.custom_lambda.function_arn
  trigger_types       = var.spec.organization.trigger_types
  description         = var.spec.description != "" ? var.spec.description : null

  input_parameters            = var.spec.input_parameters != "" ? var.spec.input_parameters : null
  maximum_execution_frequency = var.spec.maximum_execution_frequency != "" ? var.spec.maximum_execution_frequency : null
  excluded_accounts           = length(var.spec.organization.excluded_accounts) > 0 ? var.spec.organization.excluded_accounts : null

  resource_id_scope    = var.spec.scope != null && var.spec.scope.compliance_resource_id != "" ? var.spec.scope.compliance_resource_id : null
  resource_types_scope = var.spec.scope != null && length(var.spec.scope.compliance_resource_types) > 0 ? var.spec.scope.compliance_resource_types : null
  tag_key_scope        = var.spec.scope != null && var.spec.scope.tag_key != "" ? var.spec.scope.tag_key : null
  tag_value_scope      = var.spec.scope != null && var.spec.scope.tag_value != "" ? var.spec.scope.tag_value : null
}

resource "aws_config_organization_custom_policy_rule" "this" {
  count = local.is_org && var.spec.custom_policy != null ? 1 : 0

  name           = var.metadata.name
  policy_runtime = var.spec.custom_policy.policy_runtime
  policy_text    = var.spec.custom_policy.policy_text
  trigger_types  = var.spec.organization.trigger_types
  description    = var.spec.description != "" ? var.spec.description : null

  input_parameters            = var.spec.input_parameters != "" ? var.spec.input_parameters : null
  maximum_execution_frequency = var.spec.maximum_execution_frequency != "" ? var.spec.maximum_execution_frequency : null
  excluded_accounts           = length(var.spec.organization.excluded_accounts) > 0 ? var.spec.organization.excluded_accounts : null

  debug_log_delivery_accounts = length(var.spec.organization.debug_log_delivery_accounts) > 0 ? var.spec.organization.debug_log_delivery_accounts : null

  resource_id_scope    = var.spec.scope != null && var.spec.scope.compliance_resource_id != "" ? var.spec.scope.compliance_resource_id : null
  resource_types_scope = var.spec.scope != null && length(var.spec.scope.compliance_resource_types) > 0 ? var.spec.scope.compliance_resource_types : null
  tag_key_scope        = var.spec.scope != null && var.spec.scope.tag_key != "" ? var.spec.scope.tag_key : null
  tag_value_scope      = var.spec.scope != null && var.spec.scope.tag_value != "" ? var.spec.scope.tag_value : null
}

# ----- remediation (account-scoped rules only; the CEL guarantees it) -----

resource "aws_config_remediation_configuration" "this" {
  count = var.spec.remediation != null ? 1 : 0

  # Referencing the rule resource (not var) wires create AND destroy
  # ordering: the remediation goes first on teardown.
  config_rule_name = aws_config_config_rule.this[0].name

  target_type    = "SSM_DOCUMENT"
  target_id      = var.spec.remediation.target_id
  target_version = var.spec.remediation.target_version != "" ? var.spec.remediation.target_version : null
  resource_type  = var.spec.remediation.resource_type != "" ? var.spec.remediation.resource_type : null

  automatic                  = var.spec.remediation.automatic
  maximum_automatic_attempts = var.spec.remediation.maximum_automatic_attempts > 0 ? var.spec.remediation.maximum_automatic_attempts : null
  retry_attempt_seconds      = var.spec.remediation.retry_attempt_seconds > 0 ? var.spec.remediation.retry_attempt_seconds : null

  dynamic "parameter" {
    for_each = var.spec.remediation.parameters
    content {
      name           = parameter.value.name
      resource_value = parameter.value.resource_value != "" ? parameter.value.resource_value : null
      static_value   = parameter.value.static_value != "" ? parameter.value.static_value : null
      static_values  = length(parameter.value.static_values) > 0 ? parameter.value.static_values : null
    }
  }

  dynamic "execution_controls" {
    for_each = var.spec.remediation.concurrent_execution_rate_percentage > 0 || var.spec.remediation.error_percentage > 0 ? [1] : []
    content {
      ssm_controls {
        concurrent_execution_rate_percentage = var.spec.remediation.concurrent_execution_rate_percentage > 0 ? var.spec.remediation.concurrent_execution_rate_percentage : null
        error_percentage                     = var.spec.remediation.error_percentage > 0 ? var.spec.remediation.error_percentage : null
      }
    }
  }
}

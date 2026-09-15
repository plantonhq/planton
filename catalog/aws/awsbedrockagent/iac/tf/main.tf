# Amazon Bedrock agent: a foundation-model-powered assistant with folded
# satellites -- action groups (tools), aliases (immutable serving
# endpoints), collaborators (multi-agent delegation), and knowledge-base
# associations -- all attached to the agent's mutable DRAFT version.
#
# Lifecycle facts the renders below depend on:
#   - every satellite change re-"prepares" the agent (provider-managed;
#     the provider retries the OptLock/preparing conflicts itself);
#   - creating an alias SNAPSHOTS the draft into a new numbered version,
#     so aliases depend on every other satellite -- the snapshot must
#     capture the fully-assembled draft;
#   - prepare_agent and skip_resource_in_use_check are apply-behavior
#     knobs, not desired state: the module keeps the provider defaults
#     (prepare always, in-use check on) rather than exposing them.

resource "aws_bedrockagent_agent" "this" {
  # Create-time naming basis; doubles as the Name tag. metadata.name on
  # both engines.
  agent_name = local.agent_name

  # Required by AWS: the model that powers the agent and the role the
  # Bedrock service assumes to operate it.
  foundation_model        = var.spec.foundation_model
  agent_resource_role_arn = var.spec.agent_resource_role_arn

  # Optional+Computed at the provider: sent only when set so the module
  # never fights AWS's normalization.
  description                 = var.spec.description != "" ? var.spec.description : null
  instruction                 = var.spec.instruction != "" ? var.spec.instruction : null
  idle_session_ttl_in_seconds = var.spec.idle_session_ttl_seconds != 0 ? var.spec.idle_session_ttl_seconds : null
  customer_encryption_key_arn = var.spec.customer_encryption_key_arn != "" ? var.spec.customer_encryption_key_arn : null
  agent_collaboration         = var.spec.agent_collaboration != "" ? var.spec.agent_collaboration : null

  # Guardrail attachment (single-entry list ATTRIBUTE at the provider,
  # not a block).
  guardrail_configuration = local.has_guardrail ? [{
    guardrail_identifier = var.spec.guardrail.guardrail_id
    guardrail_version    = var.spec.guardrail.version
  }] : null

  # Session-summary memory -- SESSION_SUMMARY is the only memory type AWS
  # defines; the block's switch enables it and the module owns the
  # constant.
  memory_configuration = local.has_memory ? [{
    enabled_memory_types = ["SESSION_SUMMARY"]
    storage_days         = var.spec.memory.storage_days != 0 ? var.spec.memory.storage_days : null
    session_summary_configuration = var.spec.memory.max_recent_sessions != 0 ? [{
      max_recent_sessions = var.spec.memory.max_recent_sessions
    }] : null
  }] : null

  # Prompt-template overrides. Authoring an entry IS the override, so the
  # module marks every entry OVERRIDDEN (the provider strips non-
  # overridden AWS defaults from state -- a DEFAULT creation mode here
  # would vanish on read and drift forever).
  prompt_override_configuration = local.has_prompt_override ? [{
    override_lambda = var.spec.prompt_override.override_lambda != "" ? var.spec.prompt_override.override_lambda : null
    prompt_configurations = [for p in var.spec.prompt_override.prompt_configurations : {
      prompt_type          = p.prompt_type
      base_prompt_template = p.base_prompt_template
      prompt_creation_mode = "OVERRIDDEN"
      parser_mode          = p.parser_mode != "" ? p.parser_mode : null
      prompt_state         = p.prompt_state != "" ? p.prompt_state : null
      inference_configuration = p.inference_configuration != null ? [{
        max_length     = p.inference_configuration.max_length
        stop_sequences = length(p.inference_configuration.stop_sequences) > 0 ? p.inference_configuration.stop_sequences : null
        temperature    = p.inference_configuration.temperature
        top_k          = p.inference_configuration.top_k
        top_p          = p.inference_configuration.top_p
      }] : null
    }]
  }] : null

  tags = local.aws_tags
}

# Tools the agent can call, attached to the DRAFT version. A custom group
# carries an executor plus exactly one schema; a reserved group carries
# only its AWS signature (validated in the spec).
resource "aws_bedrockagent_agent_action_group" "this" {
  for_each = local.action_groups

  action_group_name = each.value.name
  agent_id          = aws_bedrockagent_agent.this.agent_id
  agent_version     = "DRAFT"

  description                   = each.value.description != "" ? each.value.description : null
  action_group_state            = each.value.state != "" ? each.value.state : null
  parent_action_group_signature = each.value.parent_action_group_signature != "" ? each.value.parent_action_group_signature : null

  dynamic "action_group_executor" {
    for_each = each.value.executor != null ? [each.value.executor] : []
    content {
      # RETURN_CONTROL is the only custom-control method AWS defines --
      # the spec models it as a bool and the module owns the constant.
      lambda         = action_group_executor.value.lambda != "" ? action_group_executor.value.lambda : null
      custom_control = action_group_executor.value.return_control ? "RETURN_CONTROL" : null
    }
  }

  dynamic "api_schema" {
    for_each = each.value.api_schema != null ? [each.value.api_schema] : []
    content {
      payload = api_schema.value.payload != "" ? api_schema.value.payload : null
      dynamic "s3" {
        for_each = api_schema.value.s3 != null ? [api_schema.value.s3] : []
        content {
          s3_bucket_name = s3.value.bucket_name
          s3_object_key  = s3.value.object_key
        }
      }
    }
  }

  dynamic "function_schema" {
    for_each = each.value.function_schema != null ? [each.value.function_schema] : []
    content {
      member_functions {
        dynamic "functions" {
          for_each = function_schema.value.functions
          content {
            name        = functions.value.name
            description = functions.value.description != "" ? functions.value.description : null
            dynamic "parameters" {
              for_each = functions.value.parameters
              content {
                # The provider calls the parameter name `map_block_key`
                # for its own compatibility reasons; the spec calls it
                # what it is.
                map_block_key = parameters.value.name
                type          = parameters.value.type
                description   = parameters.value.description != "" ? parameters.value.description : null
                required      = parameters.value.required
              }
            }
          }
        }
      }
    }
  }
}

# Agents this supervisor delegates to, attached to the DRAFT version.
resource "aws_bedrockagent_agent_collaborator" "this" {
  for_each = local.collaborators

  agent_id                  = aws_bedrockagent_agent.this.agent_id
  collaborator_name         = each.value.name
  collaboration_instruction = each.value.collaboration_instruction

  relay_conversation_history = each.value.relay_conversation_history != "" ? each.value.relay_conversation_history : null

  agent_descriptor {
    alias_arn = each.value.collaborator_alias_arn
  }
}

# Knowledge bases the agent queries, attached to the DRAFT version.
resource "aws_bedrockagent_agent_knowledge_base_association" "this" {
  for_each = local.kb_associations

  agent_id          = aws_bedrockagent_agent.this.agent_id
  knowledge_base_id = each.value.knowledge_base_id

  # Required by AWS -- the model reads it to decide when to retrieve.
  description = each.value.description

  # Required at the provider; the spec's omitted default is ENABLED (the
  # AWS default for new associations).
  knowledge_base_state = each.value.state != "" ? each.value.state : "ENABLED"
}

# Immutable serving endpoints. Each alias without an explicit routing
# snapshots the CURRENT draft into a new numbered version, so aliases are
# created only after every other satellite has landed on the draft --
# otherwise the snapshot captures a half-assembled agent.
resource "aws_bedrockagent_agent_alias" "this" {
  for_each = local.aliases

  agent_alias_name = each.value.name
  agent_id         = aws_bedrockagent_agent.this.agent_id

  description = each.value.description != "" ? each.value.description : null

  routing_configuration = each.value.routing != null ? [{
    agent_version          = each.value.routing.agent_version != "" ? each.value.routing.agent_version : null
    provisioned_throughput = each.value.routing.provisioned_throughput != "" ? each.value.routing.provisioned_throughput : null
  }] : null

  tags = local.aws_tags

  depends_on = [
    aws_bedrockagent_agent_action_group.this,
    aws_bedrockagent_agent_collaborator.this,
    aws_bedrockagent_agent_knowledge_base_association.this,
  ]
}

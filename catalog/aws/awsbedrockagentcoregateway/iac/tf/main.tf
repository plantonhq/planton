# Amazon Bedrock AgentCore gateway: a managed MCP front door with folded
# target satellites -- each target exposes one backend (agent runtime,
# API Gateway stage, Lambda tools, remote MCP server, OpenAPI/Smithy
# schema) as MCP tools behind one authenticated URL.
#
# Lifecycle facts the renders below depend on:
#   - targets attach to the gateway by ID; AWS deletes a gateway's
#     targets before the gateway itself at destroy (provider-managed);
#   - protocol_type has exactly one legal value (MCP) -- the provider
#     computes it; the module never sends it;
#   - the gateway's `targets` collection is `any`-typed in variables.tf
#     (heterogeneous JSON-document members defeat HCL's object-type
#     unification), so target attribute access is try()-based.

resource "aws_bedrockagentcore_gateway" "this" {
  # Create-time naming basis; doubles as the Name tag. metadata.name on
  # both engines. Changing it replaces the gateway.
  name = local.gateway_name

  # Required by AWS: the role the gateway assumes to reach its targets
  # and how inbound callers authenticate.
  role_arn        = var.spec.role_arn
  authorizer_type = var.spec.authorizer_type

  description = var.spec.description != "" ? var.spec.description : null
  kms_key_arn = var.spec.kms_key_arn != "" ? var.spec.kms_key_arn : null

  # DEBUG is the only exception level AWS defines -- the spec models it
  # as a bool and the module owns the constant.
  exception_level = var.spec.expose_debug_exceptions ? "DEBUG" : null

  # OIDC token validation -- required by AWS exactly when authorizer_type
  # is CUSTOM_JWT (spec-validated).
  dynamic "authorizer_configuration" {
    for_each = local.has_jwt ? [var.spec.custom_jwt_authorizer] : []
    content {
      custom_jwt_authorizer {
        discovery_url    = authorizer_configuration.value.discovery_url
        allowed_audience = length(authorizer_configuration.value.allowed_audience) > 0 ? authorizer_configuration.value.allowed_audience : null
        allowed_clients  = length(authorizer_configuration.value.allowed_clients) > 0 ? authorizer_configuration.value.allowed_clients : null
        allowed_scopes   = length(authorizer_configuration.value.allowed_scopes) > 0 ? authorizer_configuration.value.allowed_scopes : null

        dynamic "allowed_workload_configuration" {
          for_each = authorizer_configuration.value.allowed_workloads != null ? [authorizer_configuration.value.allowed_workloads] : []
          content {
            workload_identities = length(allowed_workload_configuration.value.workload_identities) > 0 ? allowed_workload_configuration.value.workload_identities : null
            dynamic "hosting_environment" {
              for_each = allowed_workload_configuration.value.hosting_environment_arns
              content {
                arn = hosting_environment.value
              }
            }
          }
        }

        dynamic "custom_claim" {
          for_each = authorizer_configuration.value.custom_claims
          content {
            inbound_token_claim_name       = custom_claim.value.claim_name
            inbound_token_claim_value_type = custom_claim.value.value_type
            authorizing_claim_match_value {
              claim_match_operator = custom_claim.value.match_operator
              claim_match_value {
                match_value_string      = custom_claim.value.match_value != "" ? custom_claim.value.match_value : null
                match_value_string_list = length(custom_claim.value.match_values) > 0 ? custom_claim.value.match_values : null
              }
            }
          }
        }

        dynamic "private_endpoint" {
          for_each = authorizer_configuration.value.private_endpoint != null ? [authorizer_configuration.value.private_endpoint] : []
          content {
            dynamic "managed_vpc_resource" {
              for_each = private_endpoint.value.managed_vpc != null ? [private_endpoint.value.managed_vpc] : []
              content {
                vpc_identifier           = managed_vpc_resource.value.vpc_id
                subnet_ids               = managed_vpc_resource.value.subnet_ids
                security_group_ids       = length(managed_vpc_resource.value.security_group_ids) > 0 ? managed_vpc_resource.value.security_group_ids : null
                endpoint_ip_address_type = managed_vpc_resource.value.endpoint_ip_address_type
                routing_domain           = managed_vpc_resource.value.routing_domain != "" ? managed_vpc_resource.value.routing_domain : null
                tags                     = length(managed_vpc_resource.value.tags) > 0 ? managed_vpc_resource.value.tags : null
              }
            }
            dynamic "self_managed_lattice_resource" {
              for_each = private_endpoint.value.self_managed_lattice != null ? [private_endpoint.value.self_managed_lattice] : []
              content {
                resource_configuration_identifier = self_managed_lattice_resource.value.resource_configuration_id
              }
            }
          }
        }

        dynamic "private_endpoint_overrides" {
          for_each = authorizer_configuration.value.private_endpoint_overrides
          content {
            domain = private_endpoint_overrides.value.domain
            private_endpoint {
              dynamic "managed_vpc_resource" {
                for_each = private_endpoint_overrides.value.private_endpoint.managed_vpc != null ? [private_endpoint_overrides.value.private_endpoint.managed_vpc] : []
                content {
                  vpc_identifier           = managed_vpc_resource.value.vpc_id
                  subnet_ids               = managed_vpc_resource.value.subnet_ids
                  security_group_ids       = length(managed_vpc_resource.value.security_group_ids) > 0 ? managed_vpc_resource.value.security_group_ids : null
                  endpoint_ip_address_type = managed_vpc_resource.value.endpoint_ip_address_type
                  routing_domain           = managed_vpc_resource.value.routing_domain != "" ? managed_vpc_resource.value.routing_domain : null
                  tags                     = length(managed_vpc_resource.value.tags) > 0 ? managed_vpc_resource.value.tags : null
                }
              }
              dynamic "self_managed_lattice_resource" {
                for_each = private_endpoint_overrides.value.private_endpoint.self_managed_lattice != null ? [private_endpoint_overrides.value.private_endpoint.self_managed_lattice] : []
                content {
                  resource_configuration_identifier = self_managed_lattice_resource.value.resource_configuration_id
                }
              }
            }
          }
        }
      }
    }
  }

  # Lambda interceptors in the request/response path (max 2).
  dynamic "interceptor_configuration" {
    for_each = var.spec.interceptors
    content {
      interception_points = interceptor_configuration.value.interception_points

      dynamic "input_configuration" {
        for_each = interceptor_configuration.value.pass_request_headers != null ? [interceptor_configuration.value.pass_request_headers] : []
        content {
          pass_request_headers = input_configuration.value
        }
      }

      interceptor {
        lambda {
          arn = interceptor_configuration.value.lambda_arn
        }
      }
    }
  }

  # Cedar policy-engine evaluation of every tool call.
  dynamic "policy_engine_configuration" {
    for_each = local.has_policy_engine ? [var.spec.policy_engine] : []
    content {
      arn  = policy_engine_configuration.value.policy_engine_arn
      mode = policy_engine_configuration.value.mode
    }
  }

  # MCP protocol tuning. MCP is the gateway's only protocol -- the
  # provider computes protocol_type; the module renders only the tuning
  # block. SEMANTIC is the only search type AWS defines -- the spec
  # models it as a bool and the module owns the constant.
  dynamic "protocol_configuration" {
    for_each = local.has_mcp ? [var.spec.mcp] : []
    content {
      mcp {
        instructions       = protocol_configuration.value.instructions != "" ? protocol_configuration.value.instructions : null
        search_type        = protocol_configuration.value.enable_semantic_search ? "SEMANTIC" : null
        supported_versions = length(protocol_configuration.value.supported_versions) > 0 ? protocol_configuration.value.supported_versions : null

        dynamic "session_configuration" {
          for_each = protocol_configuration.value.session_timeout_seconds != 0 ? [protocol_configuration.value.session_timeout_seconds] : []
          content {
            session_timeout_in_seconds = session_configuration.value
          }
        }

        dynamic "streaming_configuration" {
          for_each = protocol_configuration.value.enable_response_streaming ? [true] : []
          content {
            enable_response_streaming = true
          }
        }
      }
    }
  }

  tags = local.aws_tags
}

# The backends this gateway exposes as MCP tools. Exactly one backend arm
# and at most one credential arm per target (spec-validated). All
# attribute access is try()-based -- see the locals note.
resource "aws_bedrockagentcore_gateway_target" "this" {
  for_each = local.targets

  gateway_identifier = aws_bedrockagentcore_gateway.this.gateway_id
  name               = each.value.name

  description = try(each.value.description, "") != "" ? each.value.description : null

  target_configuration {
    # Front an AgentCore agent runtime over plain HTTP.
    dynamic "http" {
      for_each = try(each.value.backend.agentcore_runtime, null) != null ? [each.value.backend.agentcore_runtime] : []
      content {
        agentcore_runtime {
          arn       = http.value.agent_runtime_arn
          qualifier = try(http.value.qualifier, "") != "" ? http.value.qualifier : null
        }
      }
    }

    dynamic "mcp" {
      for_each = try(each.value.backend.agentcore_runtime, null) == null ? [each.value.backend] : []
      content {
        # Front an API Gateway REST API stage.
        dynamic "api_gateway" {
          for_each = try(mcp.value.api_gateway, null) != null ? [mcp.value.api_gateway] : []
          content {
            rest_api_id = api_gateway.value.rest_api_id
            stage       = api_gateway.value.stage

            dynamic "api_gateway_tool_configuration" {
              for_each = length(try(api_gateway.value.tool_filters, [])) > 0 || length(try(api_gateway.value.tool_overrides, [])) > 0 ? [api_gateway.value] : []
              content {
                dynamic "tool_filter" {
                  for_each = try(api_gateway_tool_configuration.value.tool_filters, [])
                  content {
                    filter_path = tool_filter.value.filter_path
                    methods     = tool_filter.value.methods
                  }
                }
                dynamic "tool_override" {
                  for_each = try(api_gateway_tool_configuration.value.tool_overrides, [])
                  content {
                    path        = tool_override.value.path
                    method      = tool_override.value.method
                    name        = tool_override.value.name
                    description = try(tool_override.value.description, "") != "" ? tool_override.value.description : null
                  }
                }
              }
            }
          }
        }

        # Front a Lambda function with explicitly-defined tools.
        dynamic "lambda" {
          for_each = try(mcp.value.lambda, null) != null ? [mcp.value.lambda] : []
          content {
            lambda_arn = lambda.value.lambda_arn

            tool_schema {
              # Inline tool definitions with three-level typed JSON
              # schemas; deeper shapes ride the raw-JSON leaves --
              # exactly where AWS's own configuration surface bottoms
              # out.
              dynamic "inline_payload" {
                for_each = try(lambda.value.tools, [])
                content {
                  name        = inline_payload.value.name
                  description = inline_payload.value.description

                  dynamic "input_schema" {
                    for_each = [inline_payload.value.input_schema]
                    content {
                      type        = input_schema.value.type
                      description = try(input_schema.value.description, "") != "" ? input_schema.value.description : null

                      dynamic "property" {
                        for_each = try(input_schema.value.properties, [])
                        content {
                          name        = property.value.name
                          type        = property.value.type
                          description = try(property.value.description, "") != "" ? property.value.description : null
                          required    = try(property.value.required, false)

                          dynamic "items" {
                            for_each = try(property.value.items, null) != null ? [property.value.items] : []
                            content {
                              type        = items.value.type
                              description = try(items.value.description, "") != "" ? items.value.description : null
                              dynamic "items" {
                                for_each = try(items.value.items, null) != null ? [items.value.items] : []
                                content {
                                  type            = items.value.type
                                  description     = try(items.value.description, "") != "" ? items.value.description : null
                                  items_json      = try(items.value.items_json, null) != null ? jsonencode(items.value.items_json) : null
                                  properties_json = try(items.value.properties_json, null) != null ? jsonencode(items.value.properties_json) : null
                                }
                              }
                              dynamic "property" {
                                for_each = try(items.value.properties, [])
                                content {
                                  name            = property.value.name
                                  type            = property.value.type
                                  description     = try(property.value.description, "") != "" ? property.value.description : null
                                  required        = try(property.value.required, false)
                                  items_json      = try(property.value.items_json, null) != null ? jsonencode(property.value.items_json) : null
                                  properties_json = try(property.value.properties_json, null) != null ? jsonencode(property.value.properties_json) : null
                                }
                              }
                            }
                          }

                          dynamic "property" {
                            for_each = try(property.value.properties, [])
                            content {
                              name            = property.value.name
                              type            = property.value.type
                              description     = try(property.value.description, "") != "" ? property.value.description : null
                              required        = try(property.value.required, false)
                              items_json      = try(property.value.items_json, null) != null ? jsonencode(property.value.items_json) : null
                              properties_json = try(property.value.properties_json, null) != null ? jsonencode(property.value.properties_json) : null
                            }
                          }
                        }
                      }

                      dynamic "items" {
                        for_each = try(input_schema.value.items, null) != null ? [input_schema.value.items] : []
                        content {
                          type        = items.value.type
                          description = try(items.value.description, "") != "" ? items.value.description : null
                          dynamic "items" {
                            for_each = try(items.value.items, null) != null ? [items.value.items] : []
                            content {
                              type            = items.value.type
                              description     = try(items.value.description, "") != "" ? items.value.description : null
                              items_json      = try(items.value.items_json, null) != null ? jsonencode(items.value.items_json) : null
                              properties_json = try(items.value.properties_json, null) != null ? jsonencode(items.value.properties_json) : null
                            }
                          }
                          dynamic "property" {
                            for_each = try(items.value.properties, [])
                            content {
                              name            = property.value.name
                              type            = property.value.type
                              description     = try(property.value.description, "") != "" ? property.value.description : null
                              required        = try(property.value.required, false)
                              items_json      = try(property.value.items_json, null) != null ? jsonencode(property.value.items_json) : null
                              properties_json = try(property.value.properties_json, null) != null ? jsonencode(property.value.properties_json) : null
                            }
                          }
                        }
                      }
                    }
                  }

                  dynamic "output_schema" {
                    for_each = try(inline_payload.value.output_schema, null) != null ? [inline_payload.value.output_schema] : []
                    content {
                      type        = output_schema.value.type
                      description = try(output_schema.value.description, "") != "" ? output_schema.value.description : null

                      dynamic "property" {
                        for_each = try(output_schema.value.properties, [])
                        content {
                          name        = property.value.name
                          type        = property.value.type
                          description = try(property.value.description, "") != "" ? property.value.description : null
                          required    = try(property.value.required, false)

                          dynamic "items" {
                            for_each = try(property.value.items, null) != null ? [property.value.items] : []
                            content {
                              type        = items.value.type
                              description = try(items.value.description, "") != "" ? items.value.description : null
                              dynamic "items" {
                                for_each = try(items.value.items, null) != null ? [items.value.items] : []
                                content {
                                  type            = items.value.type
                                  description     = try(items.value.description, "") != "" ? items.value.description : null
                                  items_json      = try(items.value.items_json, null) != null ? jsonencode(items.value.items_json) : null
                                  properties_json = try(items.value.properties_json, null) != null ? jsonencode(items.value.properties_json) : null
                                }
                              }
                              dynamic "property" {
                                for_each = try(items.value.properties, [])
                                content {
                                  name            = property.value.name
                                  type            = property.value.type
                                  description     = try(property.value.description, "") != "" ? property.value.description : null
                                  required        = try(property.value.required, false)
                                  items_json      = try(property.value.items_json, null) != null ? jsonencode(property.value.items_json) : null
                                  properties_json = try(property.value.properties_json, null) != null ? jsonencode(property.value.properties_json) : null
                                }
                              }
                            }
                          }

                          dynamic "property" {
                            for_each = try(property.value.properties, [])
                            content {
                              name            = property.value.name
                              type            = property.value.type
                              description     = try(property.value.description, "") != "" ? property.value.description : null
                              required        = try(property.value.required, false)
                              items_json      = try(property.value.items_json, null) != null ? jsonencode(property.value.items_json) : null
                              properties_json = try(property.value.properties_json, null) != null ? jsonencode(property.value.properties_json) : null
                            }
                          }
                        }
                      }

                      dynamic "items" {
                        for_each = try(output_schema.value.items, null) != null ? [output_schema.value.items] : []
                        content {
                          type        = items.value.type
                          description = try(items.value.description, "") != "" ? items.value.description : null
                          dynamic "items" {
                            for_each = try(items.value.items, null) != null ? [items.value.items] : []
                            content {
                              type            = items.value.type
                              description     = try(items.value.description, "") != "" ? items.value.description : null
                              items_json      = try(items.value.items_json, null) != null ? jsonencode(items.value.items_json) : null
                              properties_json = try(items.value.properties_json, null) != null ? jsonencode(items.value.properties_json) : null
                            }
                          }
                          dynamic "property" {
                            for_each = try(items.value.properties, [])
                            content {
                              name            = property.value.name
                              type            = property.value.type
                              description     = try(property.value.description, "") != "" ? property.value.description : null
                              required        = try(property.value.required, false)
                              items_json      = try(property.value.items_json, null) != null ? jsonencode(property.value.items_json) : null
                              properties_json = try(property.value.properties_json, null) != null ? jsonencode(property.value.properties_json) : null
                            }
                          }
                        }
                      }
                    }
                  }
                }
              }

              dynamic "s3" {
                for_each = try(each.value.backend.lambda.tools_s3, null) != null ? [each.value.backend.lambda.tools_s3] : []
                content {
                  uri                     = s3.value.uri
                  bucket_owner_account_id = try(s3.value.bucket_owner_account_id, "") != "" ? s3.value.bucket_owner_account_id : null
                }
              }
            }
          }
        }

        # Front an existing remote MCP server.
        dynamic "mcp_server" {
          for_each = try(mcp.value.mcp_server, null) != null ? [mcp.value.mcp_server] : []
          content {
            endpoint     = mcp_server.value.endpoint
            listing_mode = try(mcp_server.value.listing_mode, "") != "" ? mcp_server.value.listing_mode : null
          }
        }

        # Derive tools from an OpenAPI 3 schema.
        dynamic "open_api_schema" {
          for_each = try(mcp.value.open_api_schema, null) != null ? [mcp.value.open_api_schema] : []
          content {
            dynamic "inline_payload" {
              for_each = try(open_api_schema.value.inline_payload, "") != "" ? [open_api_schema.value.inline_payload] : []
              content {
                payload = inline_payload.value
              }
            }
            dynamic "s3" {
              for_each = try(open_api_schema.value.s3, null) != null ? [open_api_schema.value.s3] : []
              content {
                uri                     = s3.value.uri
                bucket_owner_account_id = try(s3.value.bucket_owner_account_id, "") != "" ? s3.value.bucket_owner_account_id : null
              }
            }
          }
        }

        # Derive tools from a Smithy model.
        dynamic "smithy_model" {
          for_each = try(mcp.value.smithy_model, null) != null ? [mcp.value.smithy_model] : []
          content {
            dynamic "inline_payload" {
              for_each = try(smithy_model.value.inline_payload, "") != "" ? [smithy_model.value.inline_payload] : []
              content {
                payload = inline_payload.value
              }
            }
            dynamic "s3" {
              for_each = try(smithy_model.value.s3, null) != null ? [smithy_model.value.s3] : []
              content {
                uri                     = s3.value.uri
                bucket_owner_account_id = try(s3.value.bucket_owner_account_id, "") != "" ? s3.value.bucket_owner_account_id : null
              }
            }
          }
        }
      }
    }
  }

  # How the GATEWAY authenticates to this backend: the credentials oneof
  # carries at most one arm by construction. jwt_passthrough is an empty
  # block at the provider -- choosing the arm IS the configuration.
  dynamic "credential_provider_configuration" {
    for_each = try(each.value.credentials, null) != null ? [each.value.credentials] : []
    content {
      dynamic "api_key" {
        for_each = try(credential_provider_configuration.value.api_key, null) != null ? [credential_provider_configuration.value.api_key] : []
        content {
          provider_arn              = api_key.value.provider_arn
          credential_location       = try(api_key.value.credential_location, "") != "" ? api_key.value.credential_location : null
          credential_parameter_name = try(api_key.value.credential_parameter_name, "") != "" ? api_key.value.credential_parameter_name : null
          credential_prefix         = try(api_key.value.credential_prefix, "") != "" ? api_key.value.credential_prefix : null
        }
      }
      dynamic "caller_iam_credentials" {
        for_each = try(credential_provider_configuration.value.caller_iam_credentials, null) != null ? [credential_provider_configuration.value.caller_iam_credentials] : []
        content {
          service = caller_iam_credentials.value.service
          region  = try(caller_iam_credentials.value.region, "") != "" ? caller_iam_credentials.value.region : null
        }
      }
      dynamic "gateway_iam_role" {
        for_each = try(credential_provider_configuration.value.gateway_iam_role, null) != null ? [credential_provider_configuration.value.gateway_iam_role] : []
        content {
          service = try(gateway_iam_role.value.service, "") != "" ? gateway_iam_role.value.service : null
          region  = try(gateway_iam_role.value.region, "") != "" ? gateway_iam_role.value.region : null
        }
      }
      dynamic "jwt_passthrough" {
        for_each = try(credential_provider_configuration.value.jwt_passthrough, null) != null ? [true] : []
        content {}
      }
      dynamic "oauth" {
        for_each = try(credential_provider_configuration.value.oauth, null) != null ? [credential_provider_configuration.value.oauth] : []
        content {
          provider_arn       = oauth.value.provider_arn
          scopes             = oauth.value.scopes
          grant_type         = try(oauth.value.grant_type, "") != "" ? oauth.value.grant_type : null
          default_return_url = try(oauth.value.default_return_url, "") != "" ? oauth.value.default_return_url : null
          custom_parameters  = length(try(oauth.value.custom_parameters, {})) > 0 ? oauth.value.custom_parameters : null
        }
      }
    }
  }

  # Caller metadata propagation (max 10 entries each); the block's own
  # switch (on by default once declared) decides whether it renders.
  dynamic "metadata_configuration" {
    for_each = try(each.value.metadata, null) != null && coalesce(try(each.value.metadata.enabled, null), true) ? [each.value.metadata] : []
    content {
      allowed_query_parameters = length(try(metadata_configuration.value.allowed_query_parameters, [])) > 0 ? metadata_configuration.value.allowed_query_parameters : null
      allowed_request_headers  = length(try(metadata_configuration.value.allowed_request_headers, [])) > 0 ? metadata_configuration.value.allowed_request_headers : null
      allowed_response_headers = length(try(metadata_configuration.value.allowed_response_headers, [])) > 0 ? metadata_configuration.value.allowed_response_headers : null
    }
  }

  # Reach a PRIVATE backend through your VPC.
  dynamic "private_endpoint" {
    for_each = try(each.value.private_endpoint, null) != null ? [each.value.private_endpoint] : []
    content {
      dynamic "managed_vpc_resource" {
        for_each = try(private_endpoint.value.managed_vpc, null) != null ? [private_endpoint.value.managed_vpc] : []
        content {
          vpc_identifier           = managed_vpc_resource.value.vpc_id
          subnet_ids               = managed_vpc_resource.value.subnet_ids
          security_group_ids       = length(try(managed_vpc_resource.value.security_group_ids, [])) > 0 ? managed_vpc_resource.value.security_group_ids : null
          endpoint_ip_address_type = managed_vpc_resource.value.endpoint_ip_address_type
          routing_domain           = try(managed_vpc_resource.value.routing_domain, "") != "" ? managed_vpc_resource.value.routing_domain : null
          tags                     = length(try(managed_vpc_resource.value.tags, {})) > 0 ? managed_vpc_resource.value.tags : null
        }
      }
      dynamic "self_managed_lattice_resource" {
        for_each = try(private_endpoint.value.self_managed_lattice, null) != null ? [private_endpoint.value.self_managed_lattice] : []
        content {
          resource_configuration_identifier = self_managed_lattice_resource.value.resource_configuration_id
        }
      }
    }
  }
}

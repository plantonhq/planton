variable "metadata" {
  description = "Cloud resource metadata"
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "GcpDialogflowCxAgent specification"
  type = object({
    # The GCP project the agent lives in: a literal project ID or a
    # GcpProject reference. If omitted, the provider's default project is
    # used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The Dialogflow CX location: "global" or a region such as "us-central1"
    # (data residency). A project's first regional agent needs Google's
    # one-time location settings, which only the Dialogflow console can set;
    # "global" needs none. Immutable.
    location = string

    # Human-readable name, unique within the location. Defaults to
    # metadata.name. Mutable in place.
    display_name = optional(string, "")

    # The agent's default language as a language tag, e.g. "en". Immutable.
    default_language_code = string

    # The agent's time zone from the IANA database, e.g. "America/New_York".
    time_zone = string

    # What the agent is for, at most 500 characters.
    description = optional(string, "")

    # The avatar shown in the console and the web demo integration.
    avatar_uri = optional(string, "")

    # The agent's languages other than default_language_code, e.g. ["es",
    # "fr"].
    supported_language_codes = optional(list(string), [])

    # Train one multilingual model over every supported language instead of
    # one per language.
    enable_multi_language_training = optional(bool, false)

    # Correct spelling in end-user input before matching.
    enable_spell_correction = optional(bool, false)

    # Lock the agent: Google rejects every change except a restore. Unlock
    # (set false) before changing anything else.
    locked = optional(bool, false)

    # The security settings applied to every conversation: a
    # GcpDialogflowCxSecuritySettings reference or a literal
    # projects/{project}/locations/{location}/securitySettings/{id} in the
    # agent's location.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    security_settings = optional(string, "")

    # Begin conversations in the agent's default playbook (a generative,
    # instruction-driven agent) instead of its default start flow (a
    # state-machine agent). Google allows only the default playbook as a
    # start playbook. Google does not read this back.
    start_with_default_playbook = optional(bool, false)

    # Agent-level advanced settings: audio export, touch-tone, logging,
    # speech.
    advanced_settings = optional(object({
      # Export incoming audio to Cloud Storage.
      audio_export_gcs_destination = optional(object({
        # The Cloud Storage URI, gs://bucket/object-name-or-prefix. Whether it is
        # a full object name or a prefix depends on the Dialogflow operation.
        uri = optional(string, "")
      }))

      # Touch-tone detection.
      dtmf_settings = optional(object({
        # Process incoming audio for key presses: a caller pressing "3" becomes
        # an event the agent's flows can route on.
        enabled = optional(bool, false)

        # The digit that ends a digit sequence, e.g. "#".
        finish_digit = optional(string, "")

        # The longest digit sequence collected.
        max_digits = optional(number, 0)
      }))

      # Conversation logging.
      logging_settings = optional(object({
        # Redact end-user input when the session parameter
        # $session.params.conversation-redaction is true -- consent-based
        # redaction.
        enable_consent_based_redaction = optional(bool, false)

        # Log interactions to Dialogflow's interaction history (needed for answer
        # feedback and conversation history in the console).
        enable_interaction_logging = optional(bool, false)

        # Log conversation queries to Cloud Logging.
        enable_stackdriver_logging = optional(bool, false)
      }))

      # Speech-to-text detection.
      speech_settings = optional(object({
        # How eagerly the end of speech is detected, 0 (least) to 100 (most).
        # Read as seconds of timeout when use_timeout_based_endpointing is set.
        endpointer_sensitivity = optional(number, 0)

        # The Speech-to-Text model per language, e.g. {"en": "phone_call"}.
        models = optional(map(string), {})

        # How long to wait for speech before a no-speech event: seconds with up
        # to nine fractional digits and a trailing "s", e.g. "3.5s".
        no_speech_timeout = optional(string, "")

        # Interpret endpointer_sensitivity as a timeout in seconds instead of a
        # sensitivity score.
        use_timeout_based_endpointing = optional(bool, false)
      }))
    }))

    # Let end users rate responses. Works only with interaction logging on
    # (advanced_settings.logging_settings.enable_interaction_logging). Google
    # does not read this back.
    enable_answer_feedback = optional(bool, false)

    # A custom client certificate for the agent's outgoing calls.
    client_certificate_settings = optional(object({
      # The certificate, PEM-encoded, including the BEGIN and END lines.
      ssl_certificate = string

      # The Secret Manager secret version holding the PEM private key:
      # projects/{project}/secrets/{secret}/versions/{version}. Dialogflow
      # reads the key from Secret Manager; the key never appears here.
      private_key = optional(string, "")

      # The Secret Manager secret version holding the key's passphrase, same
      # format. Leave empty when the private key is not encrypted.
      passphrase = optional(string, "")
    }))

    # The Vertex AI Search engine linked to the agent.
    gen_app_builder_settings = optional(object({
      # The engine: a GcpVertexAiSearchEngine reference or a literal
      # projects/{project}/locations/{location}/collections/{collection}/engines/{engine}.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      engine = string
    }))

    # Delete the linked Vertex AI Search engine (gen_app_builder_settings)
    # when the agent is destroyed. Google creates that engine when a data
    # store is connected to the agent, outside any declaration; this is the
    # only way it goes away with the agent. Refused when the engine is a
    # kind reference, because that block owns the engine's lifecycle.
    delete_chat_engine_on_destroy = optional(bool, false)

    # Git integration for exporting and restoring agent content.
    git_integration_settings = optional(object({
      # The GitHub integration.
      github_settings = optional(object({
        # The repository's display name in Dialogflow, unique among the agent's
        # repositories.
        display_name = optional(string, "")

        # The repository URI, e.g. https://api.github.com/repos/acme/support-bot.
        repository_uri = optional(string, "")

        # The branch Dialogflow tracks for the agent.
        tracking_branch = optional(string, "")

        # Every branch Dialogflow may use.
        branches = optional(list(string), [])

        # The GitHub access token Dialogflow authenticates with. Google returns
        # it redacted, so the declared value is always what the engines send.
        access_token = optional(string, "")
      }))
    }))

    # Default end-user metadata merged into every DetectIntent request, as a
    # JSON object string. Prefer templates over constants, e.g.
    # {"age": "$session.params.age"}.
    default_end_user_metadata = optional(string, "")

    # Use speech adaptation (phrase hints from the agent's content) in speech
    # recognition.
    enable_speech_adaptation = optional(bool, false)

    # Speech synthesis per language, as a JSON object string mapping a
    # language code to a SynthesizeSpeechConfig, e.g.
    # {"en": {"voice": {"name": "en-US-Neural2-C"}}}. Applies to the phone
    # gateway and to DetectIntent responses that request audio.
    synthesize_speech_configs = optional(string, "")

    # Webhooks, keyed by display name.
    webhooks = optional(list(object({
      # The webhook's human-readable name, unique within the agent.
      display_name = string

      # Keep the webhook but stop calling it.
      disabled = optional(bool, false)

      # How long Dialogflow waits for the webhook, e.g. "5s". Empty leaves
      # Google's default.
      timeout = optional(string, "")

      # Call an HTTPS endpoint directly.
      generic_web_service = optional(object({
        # The endpoint, https only.
        uri = optional(string, "")

        # What Dialogflow sends and expects:
        #   STANDARD -- Dialogflow's WebhookRequest / WebhookResponse over POST
        #   FLEXIBLE -- any method and body you define (http_method,
        #               request_body), with parameter_mapping lifting response
        #               fields into session parameters
        # Empty leaves Google's default (STANDARD).
        webhook_type = optional(string, "")

        # The HTTP method of a FLEXIBLE webhook; a STANDARD webhook always POSTs.
        http_method = optional(string, "")

        # The JSON request body of a FLEXIBLE webhook; session parameters may be
        # referenced as $session.params.<name>.
        request_body = optional(string, "")

        # FLEXIBLE webhooks: session parameter name -> field path in the webhook
        # response, e.g. {"order_status": "$.status"}.
        parameter_mapping = optional(map(string), {})

        # Plain request headers sent with every call. Put secrets in
        # secret_versions_for_request_headers instead -- these values are stored
        # in the agent as written.
        request_headers = optional(map(string), {})

        # Request headers whose values live in Secret Manager. A header named
        # here and in request_headers takes this value.
        secret_versions_for_request_headers = optional(list(object({
          # The header name, e.g. "X-Api-Key".
          key = string

          # The Secret Manager secret version holding the header value:
          # projects/{project}/secrets/{secret}/versions/{version}.
          secret_version = optional(string, "")
        })), [])

        # The Secret Manager secret version holding "username:password" for HTTP
        # Basic authentication: projects/{project}/secrets/{secret}/versions/{version}.
        secret_version_for_username_password = optional(string, "")

        # Authenticate with the OAuth client-credentials flow.
        oauth_config = optional(object({
          # The client id the third party issued.
          client_id = string

          # The endpoint Dialogflow exchanges the client credentials at for an
          # access token.
          token_endpoint = string

          # The client secret. Ignored when secret_version_for_client_secret is set
          # -- prefer that, so the secret stays in Secret Manager. Google never
          # returns it, so the declared value is always what the engines send.
          client_secret = optional(string, "")

          # The OAuth scopes to request.
          scopes = optional(list(string), [])

          # The Secret Manager secret version holding the client secret:
          # projects/{project}/secrets/{secret}/versions/{version}. Wins over
          # client_secret.
          secret_version_for_client_secret = optional(string, "")
        }))

        # Have the Dialogflow service agent mint a token for the Authorization
        # header:
        #   NONE         -- no token
        #   ID_TOKEN     -- a Google ID token (for Cloud Run and Cloud Functions
        #                   behind IAM)
        #   ACCESS_TOKEN -- a Google OAuth access token (for Google APIs)
        # Empty leaves Google's default (NONE).
        service_agent_auth = optional(string, "")

        # Send an access token of this service account in the Authorization
        # header: a GcpServiceAccount reference or a literal email. The Dialogflow
        # service agent needs roles/iam.serviceAccountTokenCreator on it.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        service_account = optional(string, "")

        # Custom CA certificates (DER, base64) trusted for this endpoint in place
        # of Google's default trust store. The server certificate must carry a
        # subject alternative name.
        allowed_ca_certs = optional(list(string), [])
      }))

      # Call an endpoint registered in Service Directory.
      service_directory = optional(object({
        # The service:
        # projects/{project}/locations/{location}/namespaces/{namespace}/services/{service}.
        service = optional(string, "")

        # How the endpoint behind the service is called.
        generic_web_service = optional(object({
          # The endpoint, https only.
          uri = optional(string, "")

          # What Dialogflow sends and expects:
          #   STANDARD -- Dialogflow's WebhookRequest / WebhookResponse over POST
          #   FLEXIBLE -- any method and body you define (http_method,
          #               request_body), with parameter_mapping lifting response
          #               fields into session parameters
          # Empty leaves Google's default (STANDARD).
          webhook_type = optional(string, "")

          # The HTTP method of a FLEXIBLE webhook; a STANDARD webhook always POSTs.
          http_method = optional(string, "")

          # The JSON request body of a FLEXIBLE webhook; session parameters may be
          # referenced as $session.params.<name>.
          request_body = optional(string, "")

          # FLEXIBLE webhooks: session parameter name -> field path in the webhook
          # response, e.g. {"order_status": "$.status"}.
          parameter_mapping = optional(map(string), {})

          # Plain request headers sent with every call. Put secrets in
          # secret_versions_for_request_headers instead -- these values are stored
          # in the agent as written.
          request_headers = optional(map(string), {})

          # Request headers whose values live in Secret Manager. A header named
          # here and in request_headers takes this value.
          secret_versions_for_request_headers = optional(list(object({
            # The header name, e.g. "X-Api-Key".
            key = string

            # The Secret Manager secret version holding the header value:
            # projects/{project}/secrets/{secret}/versions/{version}.
            secret_version = optional(string, "")
          })), [])

          # The Secret Manager secret version holding "username:password" for HTTP
          # Basic authentication: projects/{project}/secrets/{secret}/versions/{version}.
          secret_version_for_username_password = optional(string, "")

          # Authenticate with the OAuth client-credentials flow.
          oauth_config = optional(object({
            # The client id the third party issued.
            client_id = string

            # The endpoint Dialogflow exchanges the client credentials at for an
            # access token.
            token_endpoint = string

            # The client secret. Ignored when secret_version_for_client_secret is set
            # -- prefer that, so the secret stays in Secret Manager. Google never
            # returns it, so the declared value is always what the engines send.
            client_secret = optional(string, "")

            # The OAuth scopes to request.
            scopes = optional(list(string), [])

            # The Secret Manager secret version holding the client secret:
            # projects/{project}/secrets/{secret}/versions/{version}. Wins over
            # client_secret.
            secret_version_for_client_secret = optional(string, "")
          }))

          # Have the Dialogflow service agent mint a token for the Authorization
          # header:
          #   NONE         -- no token
          #   ID_TOKEN     -- a Google ID token (for Cloud Run and Cloud Functions
          #                   behind IAM)
          #   ACCESS_TOKEN -- a Google OAuth access token (for Google APIs)
          # Empty leaves Google's default (NONE).
          service_agent_auth = optional(string, "")

          # Send an access token of this service account in the Authorization
          # header: a GcpServiceAccount reference or a literal email. The Dialogflow
          # service agent needs roles/iam.serviceAccountTokenCreator on it.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          service_account = optional(string, "")

          # Custom CA certificates (DER, base64) trusted for this endpoint in place
          # of Google's default trust store. The server certificate must carry a
          # subject alternative name.
          allowed_ca_certs = optional(list(string), [])
        }))
      }))
    })), [])

    # Tools, keyed by display name.
    tools = optional(list(object({
      # The tool's human-readable name, unique within the agent.
      display_name = string

      # What the tool does and when to use it -- the model reads this to decide
      # when to call the tool, so write it for the model.
      description = string

      # An OpenAPI tool.
      open_api_spec = optional(object({
        # The OpenAPI 3 schema as text (YAML or JSON).
        text_schema = string

        # How calls authenticate. Omit for an unauthenticated API.
        authentication = optional(object({
          # An API key in a header or query parameter.
          api_key_config = optional(object({
            # The header or query parameter the key travels in, e.g. "X-Api-Key".
            key_name = string

            # Where the key goes: "HEADER" or "QUERY_STRING" (Google's
            # RequestLocation values).
            request_location = string

            # The key itself. Ignored when secret_version_for_api_key is set --
            # prefer that. Google never returns it.
            api_key = optional(string, "")

            # The Secret Manager secret version holding the key:
            # projects/{project}/secrets/{secret}/versions/{version}. Wins over
            # api_key.
            secret_version_for_api_key = optional(string, "")
          }))

          # A bearer token.
          bearer_token_config = optional(object({
            # The token. A session parameter reference such as
            # $session.params.user-token passes it per conversation. Ignored when
            # secret_version_for_token is set. Google never returns it.
            token = optional(string, "")

            # The Secret Manager secret version holding the token:
            # projects/{project}/secrets/{secret}/versions/{version}. Wins over
            # token.
            secret_version_for_token = optional(string, "")
          }))

          # OAuth.
          oauth_config = optional(object({
            # The client id the OAuth provider issued.
            client_id = string

            # The grant type: "CLIENT_CREDENTIAL" (Google's OauthGrantType values).
            oauth_grant_type = string

            # The provider's token endpoint.
            token_endpoint = string

            # The client secret. Ignored when secret_version_for_client_secret is
            # set -- prefer that. Google never returns it.
            client_secret = optional(string, "")

            # The OAuth scopes to request.
            scopes = optional(list(string), [])

            # The Secret Manager secret version holding the client secret:
            # projects/{project}/secrets/{secret}/versions/{version}. Wins over
            # client_secret.
            secret_version_for_client_secret = optional(string, "")
          }))

          # A token the Dialogflow service agent mints.
          service_agent_auth_config = optional(object({
            # The token type: "ID_TOKEN" or "ACCESS_TOKEN" (Google's ServiceAgentAuth
            # values). Empty leaves Google's default.
            service_agent_auth = optional(string, "")
          }))
        }))

        # Reach the server through Service Directory.
        service_directory_config = optional(object({
          # The service, in the agent's location:
          # projects/{project}/locations/{location}/namespaces/{namespace}/services/{service}.
          service = optional(string, "")
        }))

        # Trust custom CA certificates for the server.
        tls_config = optional(object({
          # The trusted certificates.
          ca_certs = list(object({
            # A name that tells the certificates apart.
            display_name = string

            # The certificate, DER and base64-encoded. It replaces Google's default
            # trust store; the server certificate must carry a subject alternative
            # name.
            cert = string
          }))
        }))
      }))

      # A data store tool.
      data_store_spec = optional(object({
        # The data stores searched.
        data_store_connections = list(object({
          # The data store: a GcpVertexAiSearchDataStore reference or a literal
          # projects/{project}/locations/{location}/collections/{collection}/dataStores/{store}
          # (or the form without collections/).
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          data_store = optional(string, "")

          # The kind of store: "PUBLIC_WEB", "UNSTRUCTURED", or "STRUCTURED"
          # (Google's DataStoreType values).
          data_store_type = optional(string, "")

          # For PUBLIC_WEB and UNSTRUCTURED stores: "DOCUMENTS" (Google's default)
          # or "CHUNKS" (Google's DocumentProcessingMode values).
          document_processing_mode = optional(string, "")
        }))
      }))

      # A client-executed function tool.
      function_spec = optional(object({
        # The JSON schema of the function's input, as a JSON object string.
        input_schema = optional(string, "")

        # The JSON schema of the function's output, as a JSON object string.
        output_schema = optional(string, "")
      }))

      # Frozen versions of this tool, keyed by display name.
      versions = optional(list(object({
        # The version's display name. Google does not require it unique; the
        # kind keys a tool's versions by it, so it must be unique within the
        # tool. A rename replaces the version.
        display_name = string

        # The frozen definition.
        tool = object({
          # The tool's display name at the time of the snapshot.
          display_name = string

          # What the tool does and when to use it -- the model reads this.
          description = string

          # An OpenAPI tool.
          open_api_spec = optional(object({
            # The OpenAPI 3 schema as text (YAML or JSON).
            text_schema = string

            # How calls authenticate. Omit for an unauthenticated API.
            authentication = optional(object({
              # An API key in a header or query parameter.
              api_key_config = optional(object({
                # The header or query parameter the key travels in, e.g. "X-Api-Key".
                key_name = string

                # Where the key goes: "HEADER" or "QUERY_STRING" (Google's
                # RequestLocation values).
                request_location = string

                # The key itself. Ignored when secret_version_for_api_key is set --
                # prefer that. Google never returns it.
                api_key = optional(string, "")

                # The Secret Manager secret version holding the key:
                # projects/{project}/secrets/{secret}/versions/{version}. Wins over
                # api_key.
                secret_version_for_api_key = optional(string, "")
              }))

              # A bearer token.
              bearer_token_config = optional(object({
                # The token. A session parameter reference such as
                # $session.params.user-token passes it per conversation. Ignored when
                # secret_version_for_token is set. Google never returns it.
                token = optional(string, "")

                # The Secret Manager secret version holding the token:
                # projects/{project}/secrets/{secret}/versions/{version}. Wins over
                # token.
                secret_version_for_token = optional(string, "")
              }))

              # OAuth.
              oauth_config = optional(object({
                # The client id the OAuth provider issued.
                client_id = string

                # The grant type: "CLIENT_CREDENTIAL" (Google's OauthGrantType values).
                oauth_grant_type = string

                # The provider's token endpoint.
                token_endpoint = string

                # The client secret. Ignored when secret_version_for_client_secret is
                # set -- prefer that. Google never returns it.
                client_secret = optional(string, "")

                # The OAuth scopes to request.
                scopes = optional(list(string), [])

                # The Secret Manager secret version holding the client secret:
                # projects/{project}/secrets/{secret}/versions/{version}. Wins over
                # client_secret.
                secret_version_for_client_secret = optional(string, "")
              }))

              # A token the Dialogflow service agent mints.
              service_agent_auth_config = optional(object({
                # The token type: "ID_TOKEN" or "ACCESS_TOKEN" (Google's ServiceAgentAuth
                # values). Empty leaves Google's default.
                service_agent_auth = optional(string, "")
              }))
            }))

            # Reach the server through Service Directory.
            service_directory_config = optional(object({
              # The service, in the agent's location:
              # projects/{project}/locations/{location}/namespaces/{namespace}/services/{service}.
              service = optional(string, "")
            }))

            # Trust custom CA certificates for the server.
            tls_config = optional(object({
              # The trusted certificates.
              ca_certs = list(object({
                # A name that tells the certificates apart.
                display_name = string

                # The certificate, DER and base64-encoded. It replaces Google's default
                # trust store; the server certificate must carry a subject alternative
                # name.
                cert = string
              }))
            }))
          }))

          # A data store tool.
          data_store_spec = optional(object({
            # The data stores searched.
            data_store_connections = list(object({
              # The data store: a GcpVertexAiSearchDataStore reference or a literal
              # projects/{project}/locations/{location}/collections/{collection}/dataStores/{store}
              # (or the form without collections/).
              # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
              data_store = optional(string, "")

              # The kind of store: "PUBLIC_WEB", "UNSTRUCTURED", or "STRUCTURED"
              # (Google's DataStoreType values).
              data_store_type = optional(string, "")

              # For PUBLIC_WEB and UNSTRUCTURED stores: "DOCUMENTS" (Google's default)
              # or "CHUNKS" (Google's DocumentProcessingMode values).
              document_processing_mode = optional(string, "")
            }))
          }))

          # A client-executed function tool.
          function_spec = optional(object({
            # The JSON schema of the function's input, as a JSON object string.
            input_schema = optional(string, "")

            # The JSON schema of the function's output, as a JSON object string.
            output_schema = optional(string, "")
          }))
        })
      })), [])
    })), [])

    # Flow versions, keyed by flow and display name.
    versions = optional(list(object({
      # The id of the flow to snapshot, the last segment of its resource name.
      # Empty is the agent's start flow (00000000-0000-0000-0000-000000000000),
      # the one flow every agent has; other flows are authored in the console
      # and named by id here, because a full flow path contains the agent's
      # id, which does not exist before the first apply.
      flow_id = optional(string, "")

      # The version's name, at most 64 characters. Google does not require it
      # unique; the kind keys a flow's versions by it, so it must be unique per
      # flow. A rename replaces the version.
      display_name = string

      # What changed in this version, at most 500 characters.
      description = optional(string, "")
    })), [])

    # Environments, keyed by display name.
    environments = optional(list(object({
      # The environment's name, unique within the agent, at most 64
      # characters.
      display_name = string

      # What the environment is for, at most 500 characters.
      description = optional(string, "")

      # The version pinned for each flow. Google requires a version for every
      # flow reachable from the start flow; a missing one fails the apply.
      version_configs = list(object({
        # The flow, by id. Empty is the agent's start flow.
        flow_id = optional(string, "")

        # The display name of a version declared in spec.versions for this flow.
        version = optional(string, "")

        # The numeric id of a version made outside this spec (in the console or
        # by another tool), the last segment of its resource name.
        version_id = optional(string, "")
      }))
    })), [])

    # Generative settings, one per language.
    generative_settings = optional(list(object({
      # The language these settings apply to -- the agent's default language
      # or one of its supported languages. Keyed by it: changing it declares
      # settings for another language and leaves the old language's settings
      # in place.
      language_code = string

      # Generative fallback.
      fallback_settings = optional(object({
        # The display name of the prompt template in use.
        selected_prompt = optional(string, "")

        # The stored prompt templates.
        prompt_templates = optional(list(object({
          # The prompt's name, e.g. "conservative" or "chatty".
          display_name = optional(string, "")

          # Freeze the prompt against edits in the console.
          frozen = optional(bool, false)

          # The prompt sent to the model on a no-match, with placeholders Google
          # fills, e.g. "Here is a conversation $conversation, a response is: ".
          prompt_text = optional(string, "")
        })), [])
      }))

      # Filters on generated text.
      generative_safety_settings = optional(object({
        # How banned phrases match: "PARTIAL_MATCH" or "WORD_MATCH" (Google's
        # PhraseMatchStrategy values). Empty leaves Google's default.
        default_banned_phrase_match_strategy = optional(string, "")

        # Phrases generated text must never contain.
        banned_phrases = optional(list(object({
          # The phrase's language code, e.g. "en".
          language_code = string

          # The phrase.
          text = string
        })), [])
      }))

      # The knowledge connector's persona and scope.
      knowledge_connector_settings = optional(object({
        # The virtual agent's name, used in the prompt. May be empty.
        agent = optional(string, "")

        # What the agent is, e.g. "virtual agent" or "AI assistant".
        agent_identity = optional(string, "")

        # Where the agent operates, e.g. "Example company website".
        agent_scope = optional(string, "")

        # The company or organization the agent represents -- used in the prompt
        # and in knowledge search.
        business = optional(string, "")

        # A description of the business, e.g. "a family company selling freshly
        # roasted coffee beans".
        business_description = optional(string, "")

        # Stop falling back to raw data store search results when the model
        # cannot pick an answer (the fallback is on by default).
        disable_data_store_fallback = optional(bool, false)
      }))

      # The model and prompt.
      llm_model_settings = optional(object({
        # The model, by the id Dialogflow lists for generative features (a
        # Gemini model). Empty leaves Google's default.
        model = optional(string, "")

        # A custom prompt for the model.
        prompt_text = optional(string, "")
      }))
    })), [])

    # What happens to the agent and everything folded into it when this
    # resource is destroyed:
    #   "" / "DELETE" -- deleted, conversation content included
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- everything leaves management and stays in GCP
    # Generative settings are never deleted in Google either way.
    deletion_policy = optional(string, "")
  })
}

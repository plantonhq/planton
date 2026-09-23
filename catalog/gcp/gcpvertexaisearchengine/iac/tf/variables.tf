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
  description = "GcpVertexAiSearchEngine specification"
  type = object({
    # The GCP project the engine lives in: a literal project ID or a
    # GcpProject reference. If omitted, the provider's default project is
    # used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Where the engine lives: "global", "us", or "eu" -- the same location as
    # its data stores. Immutable.
    location = string

    # Which app this engine is. Empty or SEARCH builds a search engine; CHAT
    # a chat engine over a Dialogflow CX agent; RECOMMENDATION a
    # recommendation engine. Immutable: a change replaces the engine.
    engine_type = optional(string, "")

    # The engine's id -- 1-63 characters, RFC 1034 (lowercase letters,
    # digits, hyphens; starts with a letter). Defaults to metadata.name.
    # Immutable.
    engine_id = optional(string, "")

    # Human-readable name shown in the console (up to 1024 characters).
    # Defaults to metadata.name. Mutable.
    display_name = optional(string, "")

    # The collection the engine and its data stores live in. Empty is
    # Google's "default_collection", where every GcpVertexAiSearchDataStore
    # lives; a GcpVertexAiSearchDataConnector reference (or its literal
    # collection id) puts the engine over the data stores a connector syncs.
    # A RECOMMENDATION engine always lives in default_collection. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    collection_id = optional(string, "")

    # The data stores the engine answers from, as GcpVertexAiSearchDataStore
    # references (their data_store_id) or literal ids, all in the engine's
    # collection and location and enrolled in its solution (SOLUTION_TYPE_CHAT
    # for a chat engine). A recommendation engine takes at most one. Mutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    data_store_ids = list(string)

    # The industry vertical, matching the data stores': GENERIC (Google's
    # default when empty), MEDIA (search or recommendation over media), or
    # HEALTHCARE_FHIR (search only). A chat engine is always GENERIC.
    # Immutable.
    industry_vertical = optional(string, "")

    # Metadata common to every engine type. Immutable.
    common_config = optional(object({
      # The company, business, or entity the engine represents; improves
      # LLM features (summaries, answers) that mention it. Immutable.
      company_name = optional(string, "")
    }))

    # The search arm's tier, add-ons, and license tier (SEARCH only). Google
    # requires the block, so the module sends it even when omitted here.
    search_engine_config = optional(object({
      # SEARCH_TIER_STANDARD (Google's default when empty) or
      # SEARCH_TIER_ENTERPRISE, which unlocks website search with extractive
      # answers, structured data search enrichment, and the LLM add-on. The
      # Enterprise tier bills a higher per-query rate. Mutable.
      search_tier = optional(string, "")

      # Add-ons: SEARCH_ADD_ON_LLM turns on generative answers and summaries
      # (requires the Enterprise tier). Sent only when set.
      search_add_ons = optional(list(string), [])

      # The Gemini Enterprise license tier a user needs to open this app.
      # Empty lets Google default. Sent only when set.
      required_subscription_tier = optional(string, "")
    }))

    # APP_TYPE_INTRANET marks a Gemini Enterprise (intranet) app; empty is a
    # standalone search app (SEARCH only). Immutable.
    app_type = optional(string, "")

    # Do not record search analytics for this engine (SEARCH only).
    disable_analytics = optional(bool, false)

    # Feature opt-ins and opt-outs by name, each "FEATURE_STATE_ON" or
    # "FEATURE_STATE_OFF" -- e.g. "agent-sharing-without-admin-approval",
    # "disable-agent-sharing", "enable-end-user-sharing-with-groups" (SEARCH
    # only). Sent only when set.
    features = optional(map(string), {})

    # Customer-managed encryption key protecting the engine (SEARCH only):
    # a GcpKmsKey reference or a literal
    # projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}.
    # Omit for Google-managed encryption. Mutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_name = optional(string, "")

    # Cloud and private knowledge graphs (SEARCH only). Sent only when set.
    knowledge_graph_config = optional(object({
      # Use Google's Cloud Knowledge Graph (public entities) for the engine.
      enable_cloud_knowledge_graph = optional(bool)

      # Build and use a private knowledge graph over the engine's data.
      enable_private_knowledge_graph = optional(bool)

      # Cloud Knowledge Graph entity types to support.
      cloud_knowledge_graph_types = optional(list(string), [])

      # Private knowledge graph features to switch off.
      feature_config = optional(object({
        # Do not use the private knowledge graph for query auto-complete.
        disable_private_kg_auto_complete = optional(bool, false)

        # Do not enrich results with private knowledge graph entities.
        disable_private_kg_enrichment = optional(bool, false)

        # Do not show private knowledge graph entities as query UI chips.
        disable_private_kg_query_ui_chips = optional(bool, false)

        # Do not use the private knowledge graph for query understanding.
        disable_private_kg_query_understanding = optional(bool, false)
      }))
    }))

    # The chat arm's agent (CHAT only; required there). Immutable.
    chat_engine_config = optional(object({
      # Create a new Dialogflow CX agent for this engine.
      agent_creation_config = optional(object({
        # The company, organization, or entity the agent represents; used in the
        # knowledge connector's LLM prompt.
        business = optional(string, "")

        # The agent's default language as a language tag, e.g. "en".
        default_language_code = string

        # The agent's IANA time zone, e.g. "America/New_York".
        time_zone = string

        # The Dialogflow location the agent is created in, e.g. "global" or
        # "us-central1". Empty lets Google pick; a location other than the
        # engine's needs allow_cross_region.
        location = optional(string, "")
      }))

      # Link an existing Dialogflow CX agent:
      # projects/{project}/locations/{location}/agents/{agent}.
      dialogflow_agent_to_link = optional(string, "")

      # Allow the agent and the engine to live in different locations
      # (Google's default requires the same location). Consumed once at
      # creation; Google does not read it back.
      allow_cross_region = optional(bool, false)
    }))

    # The recommendation arm's model (RECOMMENDATION only; MEDIA vertical).
    media_recommendation_engine_config = optional(object({
      # The recommendation model: "recommended-for-you", "others-you-may-like",
      # "more-like-this", or "most-popular-items". With optimization_objective
      # it decides how the engine trains and serves.
      type = optional(string, "")

      # What the model optimizes: "ctr" (click-through) or "cvr" (conversion,
      # e.g. watch time). Empty lets Google default by type (ctr for
      # recommended-for-you and others-you-may-like, and for more-like-this
      # and most-popular-items).
      optimization_objective = optional(string, "")

      # The threshold a cvr objective optimizes toward.
      optimization_objective_config = optional(object({
        # "watch-percentage" (a fraction in (0, 1]) or "watch-time" (seconds in
        # (0, 86400]).
        target_field = string

        # The threshold for the target field, e.g. 0.5 for half the media
        # watched, or 90 for ninety seconds.
        target_field_value_float = optional(number)
      }))

      # TRAINING (Google's default at creation) or PAUSED. Training is part of
      # the engine's cost; pause it to hold a model still. Mutable.
      training_state = optional(string, "")

      # Feature config the chosen type needs.
      engine_features_config = optional(object({
        # For type "most-popular-items".
        most_popular_config = optional(object({
          # How many days of events the engine trains and predicts on. Required
          # for most-popular-items.
          time_window_days = optional(number)
        }))

        # For type "recommended-for-you".
        recommended_for_you_config = optional(object({
          # The event the engine is queried with at prediction time: "generic"
          # (view-item, media-play, media-complete) or "view-home-page" (those
          # plus the home-page view).
          context_event_type = optional(string, "")
        }))
      }))
    }))

    # Serving controls on the engine, each keyed by control_id. A control
    # takes effect only when serving_config lists it.
    controls = optional(list(object({
      # The control's id within the engine -- 1-63 characters, RFC 1034.
      # Immutable.
      control_id = string

      # Human-readable name (up to 128 characters).
      display_name = string

      # Where the control applies: SEARCH_USE_CASE_SEARCH (queries) and/or
      # SEARCH_USE_CASE_BROWSE (empty-query browsing).
      use_cases = optional(list(string), [])

      # When the control is active; empty means always.
      conditions = optional(list(object({
        # Terms the query must contain.
        query_terms = optional(list(object({
          # The term.
          value = string

          # True requires the query to match the term exactly; false allows a
          # partial match.
          full_match = optional(bool, false)
        })), [])

        # Time windows the control is active in (any one holding satisfies the
        # condition).
        active_time_ranges = optional(list(object({
          # Start of the window, e.g. "2026-11-20T00:00:00Z".
          start_time = optional(string, "")

          # End of the window.
          end_time = optional(string, "")
        })), [])

        # A regular expression the whole query must match.
        query_regex = optional(string, "")
      })), [])

      # Reorder matching results.
      boost_action = optional(object({
        # The data store whose documents are boosted: a GcpVertexAiSearchDataStore
        # reference (its full name) or a literal
        # projects/{project}/locations/{location}/collections/{collection}/dataStores/{store}.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        data_store = string

        # The filter selecting the documents to boost, in Vertex AI Search filter
        # syntax, e.g. "(category: ANY(\"docs\"))".
        filter = string

        # A fixed boost, -1 (bury) to 1 (promote).
        fixed_boost = optional(number)

        # A curve-based boost instead of a fixed one.
        interpolation_boost_spec = optional(object({
          # The document field the curve reads.
          field_name = optional(string, "")

          # NUMERICAL (a numeric field) or FRESHNESS (a datetime field, boosted by
          # age). Sent only when set.
          attribute_type = optional(string, "")

          # LINEAR is the only interpolation Google offers. Sent only when set.
          interpolation_type = optional(string, "")

          # The curve's control point.
          control_point = optional(object({
            # The attribute value at this point (a number for NUMERICAL, a duration
            # like "7d" or "24h" for FRESHNESS).
            attribute_value = optional(string, "")

            # The boost at this point, -1 to 1.
            boost_amount = optional(number)
          }))
        }))
      }))

      # Drop non-matching results.
      filter_action = optional(object({
        # The data store the filter applies to (a GcpVertexAiSearchDataStore
        # reference or a literal full name).
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        data_store = string

        # The filter results must match, in Vertex AI Search filter syntax.
        filter = string
      }))

      # Pin a link.
      promote_action = optional(object({
        # The data store the promotion applies to (a GcpVertexAiSearchDataStore
        # reference or a literal full name).
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        data_store = string

        # The promoted link.
        search_link_promotion = object({
          # The promoted link's title.
          title = string

          # The URI to promote (website stores).
          uri = optional(string, "")

          # The document to promote (non-website stores), as a document resource
          # name.
          document = optional(string, "")

          # A description shown with the link.
          description = optional(string, "")

          # An image shown with the link.
          image_uri = optional(string, "")

          # Return the promotion on basic site search results.
          enabled = optional(bool, false)
        })
      }))

      # Redirect the query.
      redirect_action = optional(object({
        # The URI to redirect to.
        redirect_uri = string
      }))

      # Treat terms as synonyms.
      synonyms_action = optional(object({
        # The synonyms, e.g. ["laptop", "notebook"].
        synonyms = list(string)
      }))
    })), [])

    # Which controls the engine's default serving config ("default_search",
    # which Google creates with a search engine) applies. Every id must name
    # a control above with the matching action.
    serving_config = optional(object({
      # Ids of boost controls to apply.
      boost_control_ids = optional(list(string), [])

      # Ids of filter controls to apply.
      filter_control_ids = optional(list(string), [])

      # Ids of promote controls to apply.
      promote_control_ids = optional(list(string), [])

      # Ids of redirect controls to apply.
      redirect_control_ids = optional(list(string), [])

      # Ids of synonyms controls to apply.
      synonyms_control_ids = optional(list(string), [])
    }))

    # The embeddable search widget and hosted web app. Google creates a
    # search engine's default widget config with the engine; this PATCHes
    # it, and destroy leaves it in place.
    widget_config = optional(object({
      # Who may load the widget.
      access_settings = optional(object({
        # Allow unauthenticated public access to the widget.
        allow_public_access = optional(bool, false)

        # Domains allowed to embed the widget, e.g. "www.example.com".
        allowlisted_domains = optional(list(string), [])

        # Serve Google's hosted web app for the engine.
        enable_web_app = optional(bool, false)

        # UI language as a BCP 47 tag; Google defaults to "en-US".
        language_code = optional(string, "")

        # The workforce identity pool provider users sign in through:
        # locations/global/workforcePools/{pool}/providers/{provider}.
        workforce_identity_pool_provider = optional(string, "")
      }))

      # The homepage.
      homepage_setting = optional(object({
        # Shortcuts shown on the homepage.
        shortcuts = optional(list(object({
          # The shortcut's title.
          title = optional(string, "")

          # Where the shortcut goes.
          destination_uri = optional(string, "")

          # The shortcut's icon URL.
          icon_url = optional(string, "")
        })), [])
      }))

      # Branding.
      ui_branding = optional(object({
        # The logo image URL.
        logo_url = optional(string, "")
      }))

      # UI behavior.
      ui_settings = optional(object({
        # SEARCH_ONLY, SEARCH_WITH_ANSWER (a generated answer above results), or
        # SEARCH_WITH_FOLLOW_UPS (conversational). Sent only when set.
        interaction_type = optional(string, "")

        # SNIPPET or EXTRACTIVE_ANSWER under each result; empty shows none.
        result_description_type = optional(string, "")

        # The default order of results (a SearchRequest orderBy expression).
        default_search_request_order_by = optional(string, "")

        # Do not collect user events from the widget.
        disable_user_events_collection = optional(bool, false)

        # Offer query auto-complete.
        enable_autocomplete = optional(bool, false)

        # Show the "create agent" button.
        enable_create_agent_button = optional(bool, false)

        # Enable people search.
        enable_people_search = optional(bool, false)

        # Collect result-quality feedback from end users.
        enable_quality_feedback = optional(bool, false)

        # Enable safe search.
        enable_safe_search = optional(bool, false)

        # Search as the user types.
        enable_search_as_you_type = optional(bool, false)

        # Visual content summaries on applicable searches (healthcare search).
        enable_visual_content_summary = optional(bool, false)

        # Per-data-store result rendering.
        data_store_ui_configs = optional(list(object({
          # The data store: a GcpVertexAiSearchDataStore reference (its full name)
          # or a literal full name.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          name = string

          # Fields offered as facets.
          facet_fields = optional(list(object({
            # The registered field name, e.g. "category".
            field = string

            # The name end users see.
            display_name = optional(string, "")
          })), [])

          # Which document field each result component shows.
          fields_ui_components_map = optional(list(object({
            # The component: "title", "thumbnail", "url", "custom1", "custom2", or
            # "custom3".
            ui_component = string

            # The field shown in the component.
            field = string

            # Devices the component shows on: MOBILE and/or DESKTOP.
            device_visibility = optional(list(string), [])

            # A template around the value, e.g. "Price: {value}".
            display_template = optional(string, "")
          })), [])
        })), [])

        # The generated answer's settings.
        generative_answer_config = optional(object({
          # Do not suggest related questions with the answer.
          disable_related_questions = optional(bool, false)

          # Skip answering queries classified as adversarial.
          ignore_adversarial_query = optional(bool, false)

          # Skip answering when the results are not relevant to the query.
          ignore_low_relevant_content = optional(bool, false)

          # Skip answering queries that do not seek an answer (navigational or
          # browsing queries).
          ignore_non_answer_seeking_query = optional(bool, false)

          # Where answer images come from: ALL_AVAILABLE_SOURCES, CORPUS_IMAGE_ONLY,
          # or FIGURE_GENERATION_ONLY. Sent only when set.
          image_source = optional(string, "")

          # The answer's language as a BCP 47 tag (experimental).
          language_code = optional(string, "")

          # How many times the query may be rephrased, 1-5 (Google defaults to 1).
          max_rephrase_steps = optional(number)

          # Text placed before the prompt to steer the answer model.
          model_prompt_preamble = optional(string, "")

          # The answer model version.
          model_version = optional(string, "")

          # Top results the answer is generated from, up to 10.
          result_count = optional(number)
        }))
      }))
    }))

    # Gemini Enterprise assistants on the engine, keyed by assistant_id
    # (a Gemini Enterprise license on the project is Google's prerequisite).
    assistants = optional(list(object({
      # The assistant's id within the engine, e.g. "default_assistant".
      # Immutable.
      assistant_id = string

      # Human-readable name (up to 128 characters).
      display_name = string

      # Notes shown on the configuration UI, not to end users.
      description = optional(string, "")

      # WEB_GROUNDING_TYPE_DISABLED, WEB_GROUNDING_TYPE_GOOGLE_SEARCH, or
      # WEB_GROUNDING_TYPE_ENTERPRISE_WEB_SEARCH. Sent only when set.
      web_grounding_type = optional(string, "")

      # Content policy.
      customer_policy = optional(object({
        # Phrases the assistant refuses to engage with.
        banned_phrases = optional(list(object({
          # The phrase.
          phrase = string

          # SIMPLE_STRING_MATCH or WORD_BOUNDARY_STRING_MATCH. Sent only when set.
          match_type = optional(string, "")

          # Ignore accents and umlauts when matching ("cafe" matches "café").
          ignore_diacritics = optional(bool, false)
        })), [])

        # Model Armor sanitization.
        model_armor_config = optional(object({
          # The template applied to user prompts: a GcpModelArmorTemplate
          # reference or a literal
          # projects/{project}/locations/{location}/templates/{template}.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          user_prompt_template = string

          # The template applied to assistant responses: a GcpModelArmorTemplate
          # reference or a literal template name.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          response_template = string

          # FAIL_OPEN (answer anyway when sanitization fails) or FAIL_CLOSED
          # (refuse). Sent only when set.
          failure_mode = optional(string, "")
        }))
      }))

      # Answer generation.
      generation_config = optional(object({
        # Default answer language (ISO 639-1, e.g. "en"); empty auto-detects.
        default_language = optional(string, "")

        # Text appended to the default system instruction.
        additional_system_instruction = optional(string, "")
      }))
    })), [])

    # What happens to the engine, its controls, and its assistants when this
    # resource is destroyed:
    #   "" / "DELETE" -- deleted (the widget config and serving config are
    #                    Google's own and stay with the engine's deletion)
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- everything leaves management and stays in GCP
    # The data stores are separate resources and are not touched.
    deletion_policy = optional(string, "")
  })
}

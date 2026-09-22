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
  description = "GcpVertexAiAgentEngine specification"
  type = object({
    # The GCP project the agent lives in: a literal project ID or a
    # GcpProject reference. If omitted, the provider's default project is
    # used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The Vertex AI location (region) the agent runs in, e.g. "us-central1".
    # Immutable.
    location = string

    # Human-readable name of the agent. Defaults to metadata.name.
    display_name = optional(string, "")

    # Free-text description of the agent.
    description = optional(string, "")

    # Labels on the agent.
    labels = optional(map(string), {})

    # Customer-managed encryption key protecting the agent's data: a
    # GcpKmsKey reference or a literal
    # projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}.
    # Omit to use Google-managed encryption. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_name = optional(string, "")

    # The agent: its code, identity, and deployment shape.
    spec = optional(object({
      # The open-source framework the agent is built with (e.g. "google-adk",
      # "langchain", "langgraph", "llama-index", "ag2"); tells Agent Engine
      # which framework integration to load.
      agent_framework = optional(string, "")

      # Declarations of the agent object's class methods, as ONE OpenAPI JSON
      # string (write it compact; Google normalizes it). Required by Google
      # when deploying through infrastructure-as-code rather than the SDK,
      # which infers them from the object.
      class_methods = optional(string, "")

      # Which identity the agent runs as: "" or SERVICE_ACCOUNT uses
      # service_account when set and the project's Vertex AI Reasoning Engine
      # service agent otherwise; AGENT_IDENTITY gives the agent its own
      # Agent Identity (service_account must then be unset).
      identity_type = optional(string, "")

      # Custom service account the agent runs as: a GcpServiceAccount
      # reference (resolving to its email) or a literal email. Empty runs as
      # the project's Vertex AI Reasoning Engine service agent.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      service_account = optional(string, "")

      # Run a prebuilt container image. Mutually exclusive with
      # source_code_spec.
      container_spec = optional(object({
        # Container image URI in Artifact Registry, e.g.
        # us-central1-docker.pkg.dev/{project}/{repo}/agent:latest. The image
        # must implement the Agent Engine serving contract.
        image_uri = string

        # Port the container listens on (Google's default 8080). Sent only when
        # set.
        port = optional(number)
      }))

      # Build the agent from source. Mutually exclusive with container_spec.
      source_code_spec = optional(object({
        # Source uploaded inline as a base64 .tar.gz.
        inline_source = optional(object({
          # The source archive: a gzip-compressed tarball of the source root,
          # base64-encoded (`tar czf - -C agent-source . | base64`).
          source_archive = string
        }))

        # Source pulled from a Developer Connect repository.
        developer_connect_source = optional(object({
          # The repository, directory, and ref.
          config = object({
            # The Developer Connect Git repository link, as
            # projects/*/locations/*/connections/*/gitRepositoryLinks/*. A literal
            # today: no catalog block produces one yet.
            git_repository_link = string

            # Directory, relative to the repository root, that is the source root.
            dir = string

            # The Git ref to fetch: a branch, a tag, a commit SHA, or any ref.
            revision = string
          })
        }))

        # An ADK agent described by configuration.
        agent_config_source = optional(object({
          # The ADK agent config.
          adk_config = object({
            # The ADK agent config as a JSON string (write it compact; Google
            # normalizes it).
            json_config = string
          })

          # Supporting source files (tools, callbacks) the config refers to, as an
          # inline archive.
          inline_source = optional(object({
            # The source archive: a gzip-compressed tarball of the source root,
            # base64-encoded (`tar czf - -C agent-source . | base64`).
            source_archive = string
          }))
        }))

        # Build with Vertex AI's Python build (requirements + entrypoint).
        python_spec = optional(object({
          # Python version: 3.9, 3.10 (Google's default), 3.11, 3.12, or 3.13.
          version = optional(string, "")

          # Fully qualified module that defines the agent, relative to the source
          # root (which is on sys.path), e.g. "path.to.agent". Google's default
          # "agent".
          entrypoint_module = optional(string, "")

          # The callable in entrypoint_module that IS the agent. Google's default
          # "root_agent" (the ADK convention).
          entrypoint_object = optional(string, "")

          # Path of the requirements file relative to the source root. Google's
          # default "requirements.txt".
          requirements_file = optional(string, "")
        }))

        # Build from the Dockerfile at the source root.
        image_spec = optional(object({
          # Build arguments passed as --build-arg flags.
          build_args = optional(map(string), {})
        }))
      }))

      # The legacy pickled-object package.
      package_spec = optional(object({
        # Cloud Storage URI (gs://...) of the pickled Python object.
        pickle_object_gcs_uri = optional(string, "")

        # Cloud Storage URI of the dependency files, as a .tar.gz.
        dependency_files_gcs_uri = optional(string, "")

        # Cloud Storage URI of the requirements.txt.
        requirements_gcs_uri = optional(string, "")

        # Python version: 3.8 through 3.13 (Google's default 3.10).
        python_version = optional(string, "")
      }))

      # Cloud Build settings for the source build.
      build_spec = optional(object({
        # The Cloud Build private worker pool the build runs in, as
        # projects/{project}/locations/{location}/workerPools/{pool}. A literal
        # today: the catalog's Cloud Build blocks arrive with their own family.
        worker_pool = optional(string, "")
      }))

      # Instances, resources, environment, secrets, networking, gateway.
      deployment_spec = optional(object({
        # Literal environment variables.
        env = optional(list(object({
          # Variable name.
          name = string

          # Literal value. Never a credential -- use secret_env for those.
          value = string
        })), [])

        # Environment variables filled from Secret Manager.
        secret_env = optional(list(object({
          # Variable name.
          name = string

          # The secret version the value comes from.
          secret_ref = object({
            # The secret, by name in the agent's project: a GcpSecretManagerSecret
            # reference (resolving to its secret_id output) or a literal short name.
            # The agent's identity needs roles/secretmanager.secretAccessor on it.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            secret = string

            # The version to resolve: a version number or "latest" (Google's
            # default when empty).
            version = optional(string, "")
          })
        })), [])

        # Instances kept running at all times (0-10; Google's default 1). Zero
        # scales the agent to nothing between requests at the cost of cold
        # starts. Sent only when set.
        min_instances = optional(number)

        # Most instances the agent may scale to (1-1000; 1-100 when VPC Service
        # Controls or a PSC interface is on; Google's default 100). Sent only
        # when set.
        max_instances = optional(number)

        # Concurrent requests one instance handles (Google's default 9;
        # recommended 2 x cpu + 1). Sent only when set.
        container_concurrency = optional(number)

        # Per-container limits, keys "cpu" (1, 2, 4, 6, 8) and "memory" (1Gi
        # ... 32Gi); Google's default {cpu: "4", memory: "4Gi"}. Sent only when
        # set.
        resource_limits = optional(map(string), {})

        # Private Service Connect interface into a VPC.
        psc_interface_config = optional(object({
          # Bare name of a Compute Engine network attachment in the agent's region
          # and project, created beforehand.
          network_attachment = optional(string, "")

          # Private DNS zones of other projects the agent may resolve.
          dns_peering_configs = optional(list(object({
            # DNS suffix of the peered zone, ending with a dot, e.g.
            # "my-internal-domain.corp.".
            domain = string

            # The project hosting the Cloud DNS zone: a GcpProject reference or a
            # literal project ID. The Vertex AI service agent needs roles/dns.peer
            # there.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            target_project = string

            # The VPC network in target_project where the zone is visible, by bare
            # name: a GcpVpcNetwork reference (resolving to its network_name output)
            # or a literal.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            target_network = string
          })), [])
        }))

        # Route traffic through Agent Gateway.
        agent_gateway_config = optional(object({
          # Gateway for traffic targeting the agent.
          client_to_agent_config = optional(object({
            # The Agent Gateway resource name.
            agent_gateway = string
          }))

          # Gateway for traffic originating from the agent.
          agent_to_anywhere_config = optional(object({
            # The Agent Gateway resource name.
            agent_gateway = string
          }))
        }))
      }))
    }))

    # Context services -- the Memory Bank.
    context_spec = optional(object({
      # The Memory Bank.
      memory_bank_config = optional(object({
        # How memories are generated.
        generation_config = optional(object({
          # The generation model, as
          # projects/{project}/locations/{location}/publishers/google/models/{model}.
          model = string

          # When buffered events are turned into memories.
          generation_trigger_config = optional(object({
            # The active rule; omit to flush immediately.
            generation_rule = optional(object({
              # Generate when this many events have accumulated.
              event_count = optional(number)

              # Generate at a fixed interval, as a duration string with minute
              # granularity (e.g. "300s").
              fixed_interval = optional(string, "")

              # Generate when the stream has been idle this long after its last
              # event, as a duration string with minute granularity.
              idle_duration = optional(string, "")

              # Re-include the last N already-processed events in the next window
              # for context continuity.
              overlap_event_count = optional(number)
            }))
          }))
        }))

        # How memories are looked up.
        similarity_search_config = optional(object({
          # The embedding model used to find similar memories, as
          # projects/{project}/locations/{location}/publishers/google/models/{model}.
          embedding_model = string
        }))

        # Automatic expiry.
        ttl_config = optional(object({
          # Default lifetime of every memory, as a duration (e.g. "7776000s").
          default_ttl = optional(string, "")

          # Lifetimes by how the memory came to be.
          granular_ttl_config = optional(object({
            # Lifetime of memories uploaded through CreateMemory, as a duration.
            create_ttl = optional(string, "")

            # Lifetime of memories newly generated by GenerateMemories.
            generate_created_ttl = optional(string, "")

            # Lifetime reset for memories updated by GenerateMemories.
            generate_updated_ttl = optional(string, "")
          }))

          # Default lifetime of memory REVISIONS (the history behind each memory).
          memory_revision_default_ttl = optional(string, "")
        }))

        # Do not keep memory revisions (the history behind each memory).
        disable_memory_revisions = optional(bool, false)

        # Generate memories into fixed schemas.
        structured_memory_configs = optional(list(object({
          # Scope keys (e.g. "user_id") this config applies to.
          scope_keys = optional(list(string), [])

          # The schemas memories are generated into.
          schema_configs = optional(list(object({
            # The schema's identifier.
            id = string

            # The memory schema as an OpenAPI Schema Object JSON string (write it
            # compact).
            memory_schema = optional(string, "")
          })), [])
        })), [])

        # Per-scope generation tuning.
        customization_configs = optional(list(object({
          # Scope keys (e.g. "user_id", "session_id") this config applies to.
          scope_keys = optional(list(string), [])

          # Topics memories should be associated with.
          memory_topics = optional(list(object({
            # An operator-defined topic.
            custom_memory_topic = optional(object({
              # The topic's label.
              label = string

              # What the topic covers, for the generation model.
              description = optional(string, "")
            }))

            # A Google-managed topic.
            managed_memory_topic = optional(object({
              # USER_PERSONAL_INFO, USER_PREFERENCES, KEY_CONVERSATION_DETAILS, or
              # EXPLICIT_INSTRUCTIONS.
              managed_topic_enum = string
            }))
          })), [])

          # Worked examples for the generation model.
          generate_memories_examples = optional(list(object({
            # The input conversation.
            conversation_source = optional(object({
              # The conversation, in order.
              events = list(object({
                # The turn's content.
                content = object({
                  # Who produced the turn: "user" or "model" (Google's default "user").
                  role = optional(string, "")

                  # The turn's parts.
                  parts = list(object({
                    # Plain text.
                    text = optional(string, "")

                    # Marks the part as the model's reasoning rather than its answer.
                    thought = optional(bool, false)

                    # Raw media inline.
                    inline_data = optional(object({
                      # IANA MIME type of the data.
                      mime_type = string

                      # The bytes, base64-encoded.
                      data = string
                    }))

                    # Media in Cloud Storage.
                    file_data = optional(object({
                      # IANA MIME type of the file.
                      mime_type = string

                      # The file's gs:// URI.
                      file_uri = string
                    }))

                    # A tool call the model made.
                    function_call = optional(object({
                      # Correlates the call with its response.
                      id = optional(string, "")

                      # The function's name.
                      name = optional(string, "")

                      # The arguments as a JSON object string.
                      args = optional(string, "")
                    }))

                    # A tool's response.
                    function_response = optional(object({
                      # The id of the call this answers.
                      id = optional(string, "")

                      # The function's name.
                      name = string

                      # The response as a JSON object string.
                      response = optional(string, "")
                    }))

                    # Code the model produced to run.
                    executable_code = optional(object({
                      # Identifier the execution result refers back to.
                      id = optional(string, "")

                      # PYTHON or BASH.
                      language = string

                      # The code.
                      code = string
                    }))

                    # The result of running code.
                    code_execution_result = optional(object({
                      # The executable_code part this result is for.
                      id = optional(string, "")

                      # OUTCOME_OK, OUTCOME_FAILED, or OUTCOME_DEADLINE_EXCEEDED.
                      outcome = string

                      # Captured stdout (on success) or stderr (on failure).
                      output = optional(string, "")
                    }))

                    # Clip bounds when the payload is a video.
                    video_metadata = optional(object({
                      # Start offset into the video, as a duration (e.g. "3.5s").
                      start_offset = optional(string, "")

                      # End offset into the video, as a duration.
                      end_offset = optional(string, "")
                    }))
                  }))
                })
              }))
            }))

            # The memories expected from it.
            generated_memories = optional(list(object({
              # The fact the memory records.
              fact = string

              # Topics the memory belongs to.
              topics = optional(list(object({
                # The label of a custom topic declared in memory_topics.
                custom_memory_topic_label = optional(string, "")

                # A managed topic: USER_PERSONAL_INFO, USER_PREFERENCES,
                # KEY_CONVERSATION_DETAILS, or EXPLICIT_INSTRUCTIONS.
                managed_memory_topic = optional(string, "")
              })), [])
            })), [])
          })), [])

          # Revision-merging behavior.
          consolidation_config = optional(object({
            # Revisions considered per memory candidate when consolidating.
            revisions_per_candidate_count = optional(number)
          }))

          # Turn off natural-language memory generation (structured memories
          # only).
          disable_natural_language_memories = optional(bool, false)

          # Write memories in the third person ("The user prefers...").
          enable_third_person_memories = optional(bool, false)
        })), [])
      }))
    }))

    # What happens to the agent when this resource is destroyed:
    #   "" / "DELETE" -- the agent and its memories are deleted
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- the agent leaves management and keeps running
    deletion_policy = optional(string, "")
  })
}

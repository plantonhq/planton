variable "metadata" {
  description = "Catalog object metadata"
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
  description = "KubernetesStigmer specification"
  type = object({
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    namespace = string

    version = string
    operator = object({
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      namespace = string

      create_namespace   = optional(bool, false)
      image_repository   = optional(string, "")
      image_pull_policy  = optional(string, "")
      image_pull_secrets = optional(list(string), [])
      replicas           = optional(number)
      resources = optional(object({
        limits   = optional(map(string), {})
        requests = optional(map(string), {})
        claims = optional(list(object({
          name    = string
          request = optional(string, "")
        })), [])
      }))
      log_level     = optional(string, "")
      node_selector = optional(map(string), {})
      tolerations = optional(list(object({
        # `key` is the taint key that the toleration applies to. Empty means match
        # all taint keys. If the key is empty, operator must be Exists; this
        # combination means to match all values and all keys.
        key = optional(string, "")

        # `operator` represents a key's relationship to the value. Valid operators
        # are Exists, Equal, Lt, and Gt. Defaults to Equal. Exists is equivalent to
        # wildcard for value, so that a pod can tolerate all taints of a particular
        # category. Lt and Gt perform numeric comparisons (requires feature gate
        # TaintTolerationComparisonOperators).
        operator = optional(string)

        # `value` is the taint value the toleration matches to. If the operator is
        # Exists, the value should be empty, otherwise just a regular string.
        value = optional(string, "")

        # `effect` indicates the taint effect to match. Empty means match all taint
        # effects. When specified, allowed values are NoSchedule, PreferNoSchedule
        # and NoExecute.
        effect = optional(string)

        # `tolerationSeconds` represents the period of time the toleration (which
        # must be of effect NoExecute, otherwise this field is ignored) tolerates the
        # taint. By default, it is not set, which means tolerate the taint forever
        # (do not evict). Zero and negative values will be treated as 0 (evict
        # immediately) by the system.
        toleration_seconds = optional(number)
      })), [])
    })
    image = optional(object({
      registry     = optional(string, "")
      pull_policy  = optional(string, "")
      pull_secrets = optional(list(string), [])
    }))
    owner = optional(object({
      email = string
      name  = optional(string, "")
    }))
    server = optional(object({
      resources = optional(object({
        # `limits` describes the maximum amount of compute resources allowed. More
        # info:
        # https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
        limits = optional(map(string), {})

        # `requests` describes the minimum amount of compute resources required. If
        # Requests is omitted for a container, it defaults to Limits if that is
        # explicitly specified, otherwise to an implementation-defined value.
        # Requests cannot exceed Limits. More info:
        # https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
        requests = optional(map(string), {})

        # `claims` lists the names of resources, defined in spec.resourceClaims, that
        # are used by this container.
        #
        # This field depends on the DynamicResourceAllocation feature gate.
        #
        # This field is immutable. It can only be set for containers.
        claims = optional(list(object({
          # `name` must match the name of one entry in pod.spec.resourceClaims of the
          # Pod where this field is used. It makes that resource available inside a
          # container.
          name = string

          # `request` is the name chosen for a request in the referenced claim. If
          # empty, everything from the claim is made available, otherwise only the
          # result of this request.
          request = optional(string, "")
        })), [])
      }))
      storage = optional(object({
        size               = optional(string, "")
        storage_class_name = optional(string, "")
      }))
      artifacts = optional(object({
        size               = optional(string, "")
        storage_class_name = optional(string, "")
      }))
      help = optional(object({
        url   = string
        label = optional(string, "")
      }))
      internal_accounts              = optional(list(string), [])
      allow_unauthenticated_exposure = optional(bool, false)
      egress = optional(object({
        fenced      = optional(bool, false)
        allow_cidrs = optional(list(string), [])
      }))
      extra_env = optional(object({
        # Individual environment variables (non-sensitive). Their values are
        # written into the pod spec, where anyone who can read the workload sees
        # them; a secret belongs in `secrets`, which the module keeps in a
        # Kubernetes Secret the container reads by reference.
        variables = optional(list(object({
          # The environment variable name.
          # Must be a valid C_IDENTIFIER: starts with a letter or underscore,
          # followed by letters, digits, or underscores.
          name = string

          # Direct literal value.
          value = optional(string)

          # Reference to a key in a Kubernetes ConfigMap.
          config_map_key_ref = optional(object({
            # Name of the ConfigMap.
            name = string

            # Key within the ConfigMap.
            key = string

            # If true, the env var is silently skipped when the ConfigMap or key does not exist
            # (instead of blocking pod startup).
            optional = optional(bool, false)
          }))

          # Reference to a pod-level field (metadata.name, status.podIP, etc.).
          field_ref = optional(object({
            # Version of the schema. Defaults to "v1".
            api_version = optional(string, "")

            # Path of the field to select (e.g., "metadata.name", "status.podIP").
            field_path = string
          }))

          # Reference to container resource limits or requests (limits.cpu, requests.memory, etc.).
          resource_field_ref = optional(object({
            # Container name. Required for init containers; defaults to the current
            # container for regular containers.
            container_name = optional(string, "")

            # Resource to select (e.g., "limits.cpu", "requests.memory").
            resource = string

            # Specifies the output format of the exposed resource.
            # For CPU: "1" means cores. For memory: "1", "1Ki", "1Mi", "1Gi".
            divisor = optional(string, "")
          }))
        })), [])

        # Individual secret environment variables (sensitive).
        secrets = optional(list(object({
          # The environment variable name.
          name = string

          # The secret's value: on Planton a `$secret/<slug>` reference, resolved at deploy; on a
          # deploy without the platform, the literal. The module writes it into a Kubernetes Secret
          # and the environment variable references that Secret.
          value = optional(string)

          # Reference to a key within an existing Kubernetes Secret.
          secret_ref = optional(object({
            # The namespace of the Kubernetes Secret.
            # If not specified, defaults to the namespace where the component is deployed.
            # Note: Cross-namespace secret references may not be supported by all Helm charts.
            namespace = optional(string, "")

            # The name of the Kubernetes Secret.
            name = string

            # The key within the Kubernetes Secret that contains the value.
            key = string

            # If true, the env var is silently skipped when the Secret or key does not exist
            # (instead of blocking pod startup).
            optional = optional(bool, false)
          }))
        })), [])

        # Bulk import of environment variables from ConfigMaps or Secrets.
        env_from = optional(list(object({
          # Optional prefix prepended to each imported key name.
          # For example, prefix "APP_" with key "PORT" produces env var "APP_PORT".
          prefix = optional(string, "")

          # Import all keys from a ConfigMap.
          config_map_ref = optional(object({
            # Name of the ConfigMap.
            name = string

            # If true, the ConfigMap is allowed to not exist without blocking pod startup.
            optional = optional(bool, false)
          }))

          # Import all keys from a Secret.
          secret_ref = optional(object({
            # Name of the Secret.
            name = string

            # If true, the Secret is allowed to not exist without blocking pod startup.
            optional = optional(bool, false)
          }))
        })), [])
      }))
    }))
    runner = optional(object({
      model = optional(object({
        keys = optional(object({
          anthropic_api_key         = optional(string, "")
          anthropic_base_url        = optional(string, "")
          openai_api_key            = optional(string, "")
          anthropic_foundry_api_key = optional(string, "")
          aws_bearer_token_bedrock  = optional(string, "")
        }))

        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        existing_secret_name = optional(string, "")
      }))
      resources = optional(object({
        # `limits` describes the maximum amount of compute resources allowed. More
        # info:
        # https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
        limits = optional(map(string), {})

        # `requests` describes the minimum amount of compute resources required. If
        # Requests is omitted for a container, it defaults to Limits if that is
        # explicitly specified, otherwise to an implementation-defined value.
        # Requests cannot exceed Limits. More info:
        # https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
        requests = optional(map(string), {})

        # `claims` lists the names of resources, defined in spec.resourceClaims, that
        # are used by this container.
        #
        # This field depends on the DynamicResourceAllocation feature gate.
        #
        # This field is immutable. It can only be set for containers.
        claims = optional(list(object({
          # `name` must match the name of one entry in pod.spec.resourceClaims of the
          # Pod where this field is used. It makes that resource available inside a
          # container.
          name = string

          # `request` is the name chosen for a request in the referenced claim. If
          # empty, everything from the claim is made available, otherwise only the
          # result of this request.
          request = optional(string, "")
        })), [])
      }))
      storage = optional(object({
        size               = optional(string, "")
        storage_class_name = optional(string, "")
      }))
      extra_env = optional(object({
        # Individual environment variables (non-sensitive). Their values are
        # written into the pod spec, where anyone who can read the workload sees
        # them; a secret belongs in `secrets`, which the module keeps in a
        # Kubernetes Secret the container reads by reference.
        variables = optional(list(object({
          # The environment variable name.
          # Must be a valid C_IDENTIFIER: starts with a letter or underscore,
          # followed by letters, digits, or underscores.
          name = string

          # Direct literal value.
          value = optional(string)

          # Reference to a key in a Kubernetes ConfigMap.
          config_map_key_ref = optional(object({
            # Name of the ConfigMap.
            name = string

            # Key within the ConfigMap.
            key = string

            # If true, the env var is silently skipped when the ConfigMap or key does not exist
            # (instead of blocking pod startup).
            optional = optional(bool, false)
          }))

          # Reference to a pod-level field (metadata.name, status.podIP, etc.).
          field_ref = optional(object({
            # Version of the schema. Defaults to "v1".
            api_version = optional(string, "")

            # Path of the field to select (e.g., "metadata.name", "status.podIP").
            field_path = string
          }))

          # Reference to container resource limits or requests (limits.cpu, requests.memory, etc.).
          resource_field_ref = optional(object({
            # Container name. Required for init containers; defaults to the current
            # container for regular containers.
            container_name = optional(string, "")

            # Resource to select (e.g., "limits.cpu", "requests.memory").
            resource = string

            # Specifies the output format of the exposed resource.
            # For CPU: "1" means cores. For memory: "1", "1Ki", "1Mi", "1Gi".
            divisor = optional(string, "")
          }))
        })), [])

        # Individual secret environment variables (sensitive).
        secrets = optional(list(object({
          # The environment variable name.
          name = string

          # The secret's value: on Planton a `$secret/<slug>` reference, resolved at deploy; on a
          # deploy without the platform, the literal. The module writes it into a Kubernetes Secret
          # and the environment variable references that Secret.
          value = optional(string)

          # Reference to a key within an existing Kubernetes Secret.
          secret_ref = optional(object({
            # The namespace of the Kubernetes Secret.
            # If not specified, defaults to the namespace where the component is deployed.
            # Note: Cross-namespace secret references may not be supported by all Helm charts.
            namespace = optional(string, "")

            # The name of the Kubernetes Secret.
            name = string

            # The key within the Kubernetes Secret that contains the value.
            key = string

            # If true, the env var is silently skipped when the Secret or key does not exist
            # (instead of blocking pod startup).
            optional = optional(bool, false)
          }))
        })), [])

        # Bulk import of environment variables from ConfigMaps or Secrets.
        env_from = optional(list(object({
          # Optional prefix prepended to each imported key name.
          # For example, prefix "APP_" with key "PORT" produces env var "APP_PORT".
          prefix = optional(string, "")

          # Import all keys from a ConfigMap.
          config_map_ref = optional(object({
            # Name of the ConfigMap.
            name = string

            # If true, the ConfigMap is allowed to not exist without blocking pod startup.
            optional = optional(bool, false)
          }))

          # Import all keys from a Secret.
          secret_ref = optional(object({
            # Name of the Secret.
            name = string

            # If true, the Secret is allowed to not exist without blocking pod startup.
            optional = optional(bool, false)
          }))
        })), [])
      }))
    }))
    database = optional(object({
      bundled = optional(object({
        instances = optional(number)
        storage = optional(object({
          size               = optional(string, "")
          storage_class_name = optional(string, "")
        }))
        resources = optional(object({
          # `limits` describes the maximum amount of compute resources allowed. More
          # info:
          # https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
          limits = optional(map(string), {})

          # `requests` describes the minimum amount of compute resources required. If
          # Requests is omitted for a container, it defaults to Limits if that is
          # explicitly specified, otherwise to an implementation-defined value.
          # Requests cannot exceed Limits. More info:
          # https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
          requests = optional(map(string), {})

          # `claims` lists the names of resources, defined in spec.resourceClaims, that
          # are used by this container.
          #
          # This field depends on the DynamicResourceAllocation feature gate.
          #
          # This field is immutable. It can only be set for containers.
          claims = optional(list(object({
            # `name` must match the name of one entry in pod.spec.resourceClaims of the
            # Pod where this field is used. It makes that resource available inside a
            # container.
            name = string

            # `request` is the name chosen for a request in the referenced claim. If
            # empty, everything from the claim is made available, otherwise only the
            # result of this request.
            request = optional(string, "")
          })), [])
        }))
        backup = optional(object({
          schedule  = optional(string, "")
          retention = optional(string, "")
        }))
        recover_from = optional(object({
          server_name = string
          target_time = optional(string, "")
        }))
      }))
      external = optional(object({
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        host = string

        port     = optional(number)
        user     = string
        database = optional(string)
        password = object({
          value = optional(string)
          existing_secret = optional(object({
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            name = string

            key = optional(string)
          }))
        })
        ssl_mode = optional(string, "")
      }))
    }))
    temporal = optional(object({
      bundled = optional(object({
        resources = optional(object({
          # `limits` describes the maximum amount of compute resources allowed. More
          # info:
          # https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
          limits = optional(map(string), {})

          # `requests` describes the minimum amount of compute resources required. If
          # Requests is omitted for a container, it defaults to Limits if that is
          # explicitly specified, otherwise to an implementation-defined value.
          # Requests cannot exceed Limits. More info:
          # https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
          requests = optional(map(string), {})

          # `claims` lists the names of resources, defined in spec.resourceClaims, that
          # are used by this container.
          #
          # This field depends on the DynamicResourceAllocation feature gate.
          #
          # This field is immutable. It can only be set for containers.
          claims = optional(list(object({
            # `name` must match the name of one entry in pod.spec.resourceClaims of the
            # Pod where this field is used. It makes that resource available inside a
            # container.
            name = string

            # `request` is the name chosen for a request in the referenced claim. If
            # empty, everything from the claim is made available, otherwise only the
            # result of this request.
            request = optional(string, "")
          })), [])
        }))
      }))
      external = optional(object({
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        host_port = string

        namespace = optional(string)
        tls = optional(object({
          server_name = optional(string, "")
          pem = optional(object({
            ca_certificate     = optional(string, "")
            client_certificate = optional(string, "")
            client_private_key = optional(string, "")
          }))
          existing_secret = optional(object({
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            name = string

            ca_key          = optional(string, "")
            certificate_key = optional(string, "")
            private_key_key = optional(string, "")
          }))
        }))
        api_key = optional(object({
          value = optional(string)
          existing_secret = optional(object({
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            name = string

            key = string
          }))
        }))
      }))
    }))
    object_store = optional(object({
      s3 = object({
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        bucket = string

        region = optional(string)

        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        endpoint = optional(string, "")

        credentials = object({
          access_key = optional(object({
            access_key_id     = string
            secret_access_key = string
          }))

          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          existing_secret_name = optional(string, "")
        })
        runner_endpoint = optional(string, "")
      })
    }))
    ingress = optional(object({
      hostname = string
      tls = optional(object({
        secret_name = optional(string)
        issuer = optional(object({
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          name = string

          kind = optional(string, "")
        }))
      }))
      gateway = optional(object({
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        name = string

        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        namespace = optional(string, "")

        section_name = optional(string, "")
      }))
      class_name          = optional(string, "")
      annotations         = optional(map(string), {})
      service_annotations = optional(map(string), {})
    }))
    identity = optional(object({
      bundled = optional(object({
        sign_up_email_domains = optional(list(string), [])
        smtp = optional(object({
          host     = string
          port     = optional(number)
          from     = string
          starttls = optional(bool, false)
          credentials = optional(object({
            login = optional(object({
              username = string
              password = string
            }))

            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            existing_secret_name = optional(string, "")
          }))
        }))
      }))
      oidc = optional(object({
        issuer            = string
        audience          = string
        console_client_id = optional(string, "")
        native_client_id  = optional(string, "")
        discovery_url     = optional(string, "")
      }))
    }))
    enterprise = optional(object({
      license = object({
        value = optional(string)
        existing_secret = optional(object({
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          name = string

          key = string
        }))
      })
      organization_creators         = optional(list(string), [])
      access_history_retention_days = optional(number)
      outbound_egress               = optional(string, "")
    }))
    sandboxes = optional(object({
      driver       = optional(string, "")
      idle_seconds = optional(number)
      agent_sandbox = optional(object({
        egress_allow_cidrs = optional(list(string), [])
      }))
      substrate = optional(object({
        workers = optional(number)
        worker_resources = optional(object({
          # `limits` describes the maximum amount of compute resources allowed. More
          # info:
          # https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
          limits = optional(map(string), {})

          # `requests` describes the minimum amount of compute resources required. If
          # Requests is omitted for a container, it defaults to Limits if that is
          # explicitly specified, otherwise to an implementation-defined value.
          # Requests cannot exceed Limits. More info:
          # https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
          requests = optional(map(string), {})

          # `claims` lists the names of resources, defined in spec.resourceClaims, that
          # are used by this container.
          #
          # This field depends on the DynamicResourceAllocation feature gate.
          #
          # This field is immutable. It can only be set for containers.
          claims = optional(list(object({
            # `name` must match the name of one entry in pod.spec.resourceClaims of the
            # Pod where this field is used. It makes that resource available inside a
            # container.
            name = string

            # `request` is the name chosen for a request in the referenced claim. If
            # empty, everything from the claim is made available, otherwise only the
            # result of this request.
            request = optional(string, "")
          })), [])
        }))
        worker_node_selector = optional(map(string), {})
        worker_tolerations = optional(list(object({
          # `key` is the taint key that the toleration applies to. Empty means match
          # all taint keys. If the key is empty, operator must be Exists; this
          # combination means to match all values and all keys.
          key = optional(string, "")

          # `operator` represents a key's relationship to the value. Valid operators
          # are Exists, Equal, Lt, and Gt. Defaults to Equal. Exists is equivalent to
          # wildcard for value, so that a pod can tolerate all taints of a particular
          # category. Lt and Gt perform numeric comparisons (requires feature gate
          # TaintTolerationComparisonOperators).
          operator = optional(string)

          # `value` is the taint value the toleration matches to. If the operator is
          # Exists, the value should be empty, otherwise just a regular string.
          value = optional(string, "")

          # `effect` indicates the taint effect to match. Empty means match all taint
          # effects. When specified, allowed values are NoSchedule, PreferNoSchedule
          # and NoExecute.
          effect = optional(string)

          # `tolerationSeconds` represents the period of time the toleration (which
          # must be of effect NoExecute, otherwise this field is ignored) tolerates the
          # taint. By default, it is not set, which means tolerate the taint forever
          # (do not evict). Zero and negative values will be treated as 0 (evict
          # immediately) by the system.
          toleration_seconds = optional(number)
        })), [])
        sandbox_cpu_millicores = optional(number)
        sandbox_memory_mib     = optional(number)
        egress_http            = optional(list(string), [])
      }))
    }))
    prerequisites = optional(object({
      postgres_operator      = optional(string, "")
      postgres_backup_plugin = optional(string, "")
      agent_sandbox          = optional(string, "")
      substrate              = optional(string, "")
    }))
    upgrade_without_backup_to = optional(string, "")
  })
}

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
  description = "KubernetesPlantonPlatform specification"
  type = object({
    namespace        = string
    create_namespace = optional(bool, false)
    version          = string
    license = optional(object({
      key = optional(string, "")
      secret_key_ref = optional(object({
        name = string
        key  = string
      }))
    }))
    storage = optional(object({
      storage_class_name = optional(string, "")
      size               = optional(string, "")
    }))
    database = optional(object({
      postgresql = optional(object({
        replicas           = optional(number)
        storage_size       = optional(string, "")
        storage_class_name = optional(string, "")
        # backup and recover_from declare an object store in one shape. The
        # r2 arm's account_id, jurisdiction, and credentials are foreign keys
        # in the spec (CloudflareR2Bucket, CloudflareAccountApiToken); they
        # arrive here already resolved to plain strings.
        backup = optional(object({
          object_store = object({
            destination_path = string
            s3 = optional(object({
              region          = optional(string, "")
              endpoint_url    = optional(string, "")
              endpoint_ca_pem = optional(string, "")
              keyless         = optional(bool, false)
              access_keys = optional(object({
                access_key_id     = string
                secret_access_key = string
              }))
            }))
            gcs = optional(object({
              keyless                  = optional(bool, false)
              service_account_key_json = optional(string, "")
            }))
            azure_blob = optional(object({
              storage_account   = string
              keyless           = optional(bool, false)
              connection_string = optional(string, "")
            }))
            r2 = optional(object({
              account_id   = string
              jurisdiction = optional(string, "")
              credentials = object({
                access_key_id     = string
                secret_access_key = string
              })
            }))
          })
          retention_policy            = optional(string)
          schedule                    = optional(string)
          service_account_annotations = optional(map(string), {})
        }))
        recover_from = optional(object({
          object_store = object({
            destination_path = string
            s3 = optional(object({
              region          = optional(string, "")
              endpoint_url    = optional(string, "")
              endpoint_ca_pem = optional(string, "")
              keyless         = optional(bool, false)
              access_keys = optional(object({
                access_key_id     = string
                secret_access_key = string
              }))
            }))
            gcs = optional(object({
              keyless                  = optional(bool, false)
              service_account_key_json = optional(string, "")
            }))
            azure_blob = optional(object({
              storage_account   = string
              keyless           = optional(bool, false)
              connection_string = optional(string, "")
            }))
            r2 = optional(object({
              account_id   = string
              jurisdiction = optional(string, "")
              credentials = object({
                access_key_id     = string
                secret_access_key = string
              })
            }))
          })
          server_name = string
          target_time = optional(string, "")
        }))
      }))
      redis = optional(object({
        storage_size       = optional(string, "")
        storage_class_name = optional(string, "")
      }))
    }))
    ingress = optional(object({
      enabled            = optional(bool, false)
      hostname           = optional(string, "")
      ingress_class_name = optional(string, "")
      # name and namespace are KubernetesGateway foreign keys in the spec;
      # they arrive here already resolved to plain strings.
      gateway_ref = optional(object({
        name         = string
        namespace    = optional(string, "")
        section_name = optional(string, "")
      }))
      annotations = optional(map(string), {})
      tls = optional(object({
        secret_name = optional(string, "")
        issuer = optional(object({
          name = string
          kind = optional(string)
        }))
      }))
      # auto | public | private; empty rides the CRD default (auto).
      reachability = optional(string, "")
    }))
    gateway = optional(object({
      local_port = optional(number)
    }))
    identity = optional(object({
      realm       = optional(string)
      admin_email = optional(string, "")
    }))
    bootstrap = optional(object({
      organization = optional(object({
        slug = optional(string)
        name = optional(string, "")
      }))
      environment = optional(object({
        slug = optional(string)
        name = optional(string, "")
      }))
      admins          = optional(list(string), [])
      iac_provisioner = optional(string)
      secret_backend = optional(object({
        type = string
        aws_secrets_manager = optional(object({
          region      = string
          kms_key_arn = string
        }))
      }))
    }))
    runner = optional(object({
      enabled                       = optional(bool)
      storage_size                  = optional(string, "")
      storage_class_name            = optional(string, "")
      service_account_annotations   = optional(map(string), {})
      cloud_credentials_secret_name = optional(string, "")
    }))
    build = optional(object({
      enabled = optional(bool)
    }))
    remote_runners = optional(object({
      enabled = optional(bool)
    }))
    # Outbound email: exactly one of smtp | resend (the spec's CEL holds
    # it). Credentials are Secret names and Secret key references, never
    # values; port and security ride the CRD defaults (587, starttls) when
    # omitted.
    email = optional(object({
      from = object({
        address = string
        name    = optional(string)
      })
      reply_to = optional(string, "")
      smtp = optional(object({
        host = string
        port = optional(number)
        # starttls | tls | none; empty rides the CRD default (starttls).
        security                = optional(string)
        credentials_secret_name = optional(string, "")
        oauth2 = optional(object({
          user      = string
          token_url = string
          scope     = string
          client_id = string
          client_secret_ref = object({
            name = string
            key  = string
          })
        }))
        ca_bundle_secret_ref = optional(object({
          name = string
          key  = string
        }))
      }))
      resend = optional(object({
        api_key_secret_ref = object({
          name = string
          key  = string
        })
      }))
    }))
    vault = optional(object({
      enabled = optional(bool)
      # Exactly one seal arm (the spec's CEL holds it). The GCP arm's
      # project, key_ring, crypto_key, and workload_identity_service_account
      # are foreign keys in the spec; they arrive here already resolved to
      # plain strings. Credential values never reach the CR: the module
      # materializes them as the seal-credentials Secret the CR names.
      auto_unseal = optional(object({
        aws_kms = optional(object({
          region            = string
          kms_key_id        = string
          access_key_id     = optional(string, "")
          secret_access_key = optional(string, "")
        }))
        gcp_kms = optional(object({
          project                           = string
          region                            = string
          key_ring                          = string
          crypto_key                        = string
          workload_identity_service_account = optional(string, "")
        }))
        azure_key_vault = optional(object({
          vault_name    = string
          key_name      = string
          tenant_id     = string
          client_id     = optional(string, "")
          client_secret = optional(string, "")
        }))
        transit = optional(object({
          address    = string
          key_name   = string
          mount_path = optional(string)
          token      = optional(string, "")
        }))
      }))
      init_secret_name            = optional(string, "")
      service_account_annotations = optional(map(string), {})
    }))
    components = optional(object({
      graph = optional(object({
        enabled            = optional(bool, false)
        storage_size       = optional(string, "")
        storage_class_name = optional(string, "")
      }))
    }))
    prerequisites = optional(object({
      postgres_operator      = optional(string)
      tekton_pipelines       = optional(string)
      postgres_backup_plugin = optional(string)
    }))
    control_plane = optional(object({
      image = optional(object({
        repository = optional(string, "")
        tag        = optional(string, "")
      }))
      replicas                    = optional(number)
      external_config_secret_name = optional(string, "")
      service_account_annotations = optional(map(string), {})
    }))
    console = optional(object({
      image = optional(object({
        repository = optional(string, "")
        tag        = optional(string, "")
      }))
      replicas                    = optional(number)
      external_config_secret_name = optional(string, "")
    }))
  })
}
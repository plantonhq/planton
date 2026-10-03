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
  description = "AwsBedrockKnowledgeBase specification"
  type = object({
    region      = string
    description = optional(string, "")
    role_arn    = string
    vector = optional(object({
      embedding_model_arn = string
      embedding_model = optional(object({
        dimensions                 = optional(number, 0)
        embedding_data_type        = optional(string, "")
        audio_segmentation_seconds = optional(number, 0)
        video_segmentation_seconds = optional(number, 0)
      }))
      supplemental_data_s3_uri = optional(string, "")
    }))
    managed = optional(object({
      embedding_model_arn = optional(string, "")
      embedding_model = optional(object({
        dimensions                 = optional(number, 0)
        embedding_data_type        = optional(string, "")
        audio_segmentation_seconds = optional(number, 0)
        video_segmentation_seconds = optional(number, 0)
      }))
      kms_key_arn = optional(string, "")
    }))
    kendra = optional(object({
      kendra_index_arn = optional(string, "")
    }))
    sql = optional(object({
      provisioned = optional(object({
        cluster_identifier = string
        auth = object({
          type                         = optional(string, "")
          database_user                = optional(string, "")
          username_password_secret_arn = optional(string, "")
        })
      }))
      serverless = optional(object({
        workgroup_arn = string
        auth = object({
          type                         = optional(string, "")
          username_password_secret_arn = optional(string, "")
        })
      }))
      warehouse = object({
        data_catalog = optional(object({
          table_names = list(string)
        }))
        redshift = optional(object({
          database_name = string
        }))
      })
      query_generation = optional(object({
        execution_timeout_seconds = optional(number, 0)
        curated_queries = optional(list(object({
          natural_language = string
          sql              = string
        })), [])
        tables = optional(list(object({
          name        = string
          description = optional(string, "")
          inclusion   = optional(string, "")
          columns = optional(list(object({
            name        = string
            description = optional(string, "")
            inclusion   = optional(string, "")
          })), [])
        })), [])
      }))
    }))
    storage = optional(object({
      opensearch_serverless = optional(object({
        collection_arn    = string
        vector_index_name = string
        field_mapping = object({
          vector_field   = string
          text_field     = string
          metadata_field = string
        })
      }))
      opensearch_managed = optional(object({
        domain_arn        = string
        domain_endpoint   = optional(string, "")
        vector_index_name = string
        field_mapping = object({
          vector_field   = string
          text_field     = string
          metadata_field = string
        })
      }))
      s3_vectors = optional(object({
        index_arn         = optional(string, "")
        index_name        = optional(string, "")
        vector_bucket_arn = optional(string, "")
      }))
      rds = optional(object({
        resource_arn           = string
        credentials_secret_arn = string
        database_name          = string
        table_name             = string
        field_mapping = object({
          vector_field          = string
          text_field            = string
          metadata_field        = string
          primary_key_field     = string
          custom_metadata_field = optional(string, "")
        })
      }))
      pinecone = optional(object({
        connection_string      = string
        credentials_secret_arn = string
        namespace              = optional(string, "")
        field_mapping = object({
          text_field     = string
          metadata_field = string
        })
      }))
      mongodb_atlas = optional(object({
        endpoint               = string
        database_name          = string
        collection_name        = string
        vector_index_name      = string
        text_index_name        = optional(string, "")
        endpoint_service_name  = optional(string, "")
        credentials_secret_arn = string
        field_mapping = object({
          vector_field   = string
          text_field     = string
          metadata_field = string
        })
      }))
      neptune_analytics = optional(object({
        graph_arn = optional(string, "")
        field_mapping = object({
          text_field     = string
          metadata_field = string
        })
      }))
      redis_enterprise_cloud = optional(object({
        endpoint               = string
        vector_index_name      = string
        credentials_secret_arn = string
        field_mapping = optional(object({
          vector_field   = optional(string, "")
          text_field     = optional(string, "")
          metadata_field = optional(string, "")
        }))
      }))
    }))
    # Deliberately `any`: connector entries are heterogeneous (one
    # connector arm per entry, plus the managed connector's JSON-document
    # parameters whose concrete type differs per entry), and HCL cannot
    # unify `any`-typed members across list elements. The module reads
    # every optional field with try().
    # The full shape, for reference (the spec proto is the contract):
    data_sources = optional(any, [])
  })
}

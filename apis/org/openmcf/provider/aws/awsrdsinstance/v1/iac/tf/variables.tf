variable "metadata" {
  description = "metadata"
  type = object({

    # name of the resource
    name = string

    # id of the resource
    id = string

    # id of the organization to which the api-resource belongs to
    org = string

    # environment to which the resource belongs to
    env = string

    # labels for the resource
    labels = object({

      # Description for key
      key = string

      # Description for value
      value = string
    })

    # annotations for the resource
    annotations = object({

      # Description for key
      key = string

      # Description for value
      value = string
    })

    # tags for the resource
    tags = list(string)
  })
}

variable "spec" {
  description = "spec"
  type = object({

    # The AWS region where the resource will be created.
    region = string

    # List of subnet IDs for the DB. DB instance will be created in the VPC associated with the DB subnet group provisioned using the subnet IDs.
    # Specify one of `subnet_ids`, `db_subnet_group_name` or `availability_zone`
    subnet_ids = list(object({

      # Description for value
      value = string

      # Description for value_from
      value_from = object({

        # Description for kind
        kind = string

        # Description for env
        env = string

        # Description for name
        name = string

        # Description for field_path
        field_path = string
      })
    }))

    # Name of DB subnet group. DB instance will be created in the VPC associated with the DB subnet group.
    # Specify one of `subnet_ids`, `db_subnet_group_name` or `availability_zone`
    db_subnet_group_name = object({

      # Description for value
      value = string

      # Description for value_from
      value_from = object({

        # Description for kind
        kind = string

        # Description for env
        env = string

        # Description for name
        name = string

        # Description for field_path
        field_path = string
      })
    })

    # The IDs of the security groups from which to allow `ingress` traffic to the DB instance
    security_group_ids = list(object({

      # Description for value
      value = string

      # Description for value_from
      value_from = object({

        # Description for kind
        kind = string

        # Description for env
        env = string

        # Description for name
        name = string

        # Description for field_path
        field_path = string
      })
    }))

    # The database engine to use. For supported values, see the Engine parameter in [API action CreateDBInstance]
    # (https://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_CreateDBInstance.html).
    # Note that for Amazon Aurora instances the engine must match the DB cluster's engine'.
    # For information on the difference between the available Aurora MySQL engines see
    # [Comparison between Aurora MySQL 1 and Aurora MySQL 2](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/AuroraMySQL.Updates.20180206.html)
    # in the Amazon RDS User Guide.
    engine = string

    # The engine version to use. If `autoMinorVersionUpgrade` is enabled, you can provide a prefix of the version such
    # as `8.0` (for `8.0.36`). The actual engine version used is returned in the attribute `engineVersionActual`,
    # see Attribute Reference below. For supported values, see the EngineVersion parameter in
    # [API action CreateDBInstance](https://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_CreateDBInstance.html).
    # Note that for Amazon Aurora instances the engine version must match the DB cluster's engine version'.
    engine_version = string

    # The instance type of the RDS instance.
    instance_class = string

    # Description for allocated_storage_gb
    allocated_storage_gb = number

    # Specifies whether the DB instance is
    # encrypted. Note that if you are creating a cross-region read replica this field
    # is ignored and you should instead declare `kmsKeyId` with a valid ARN. The
    # default is `false` if not specified.
    storage_encrypted = bool

    # The ARN for the KMS encryption key. If creating an
    # encrypted replica, set this to the destination KMS ARN.
    kms_key_id = object({

      # Description for value
      value = string

      # Description for value_from
      value_from = object({

        # Description for kind
        kind = string

        # Description for env
        env = string

        # Description for name
        name = string

        # Description for field_path
        field_path = string
      })
    })

    # (Required unless a `snapshotIdentifier` or `replicateSourceDb` is provided)
    # Username for the master DB user. Cannot be specified for a replica.
    username = string

    # (Required unless `manageMasterUserPassword` is set to true or unless a `snapshotIdentifier` or `replicateSourceDb`
    # is provided or `manageMasterUserPassword` is set.) Password for the master DB user. Note that this may show up in
    # logs, and it will be stored in the state file. Cannot be set if `manageMasterUserPassword` is set to `true`.
    password = string

    # The port on which the DB accepts connections.
    port = number

    # Description for publicly_accessible
    publicly_accessible = bool

    # Description for multi_az
    multi_az = bool

    # Name of the DB parameter group to associate.
    parameter_group_name = string

    # Name of the DB option group to associate
    option_group_name = string
  })
}
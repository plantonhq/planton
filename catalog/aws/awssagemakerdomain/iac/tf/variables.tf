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
  description = "AwsSagemakerDomain specification"
  type = object({
    # The AWS region where the resource will be created.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # auth_mode determines how users authenticate to the SageMaker Domain.
    # "IAM": users authenticate with AWS IAM credentials. Suitable for single-account teams
    #   or programmatic access. Each user's IAM identity determines their permissions.
    # "SSO": users authenticate via AWS IAM Identity Center (formerly AWS SSO). Recommended
    #   for enterprise teams with centralized identity management, and required for
    #   trusted identity propagation.
    # ForceNew: changing auth_mode forces domain replacement.
    auth_mode = string

    # vpc_id is the VPC in which the SageMaker Domain is created.
    # All domain network interfaces (for notebooks, training, and app traffic) are placed
    # in this VPC. The VPC must have DNS resolution and DNS hostnames enabled.
    # ForceNew: changing the VPC forces domain replacement.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    vpc_id = string

    # subnet_ids are the VPC subnets where SageMaker provisions elastic network interfaces
    # for notebook and training traffic. For high availability, provide subnets in at least
    # two Availability Zones. Maximum 16 subnets.
    # ForceNew: changing subnets forces domain replacement.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    subnet_ids = list(string)

    # kms_key_id is the KMS key used to encrypt the EFS volume attached to the domain.
    # Each SageMaker Domain creates a dedicated EFS file system for user home directories.
    # If omitted, AWS uses the default aws/elasticfilesystem service key.
    # ForceNew: changing the KMS key forces domain replacement.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_id = optional(string, "")

    # app_network_access_type controls whether notebook and training traffic can reach the internet.
    # "PublicInternetOnly" (default): ENIs have internet access via AWS-managed networking.
    # "VpcOnly": all traffic stays within the VPC; internet access requires a NAT gateway.
    # "VpcOnly" is recommended for production to prevent data exfiltration and satisfy
    # compliance requirements. Docker trusted accounts only work in VpcOnly mode.
    app_network_access_type = optional(string)

    # app_security_group_management selects who manages the security groups attached to the
    # ENIs that SageMaker creates for apps.
    # "Service": SageMaker creates and manages the security groups.
    # "Customer": you manage them (required when RStudio needs to reach a license endpoint
    #   through customer-controlled networking).
    # AWS only honors this setting when RStudio Server Pro is configured on the domain
    # (r_studio_server_pro_domain_settings), so the spec requires that pairing up front
    # rather than letting the value be silently ignored at deploy time.
    app_security_group_management = optional(string)

    # tag_propagation controls whether tags on the domain automatically propagate to the
    # SageMaker resources created within it (apps, spaces, user profiles).
    # "ENABLED": in-domain resources inherit the domain's tags -- recommended for cost
    #   allocation, because per-app compute charges then carry the domain's tags.
    # "DISABLED" (default, AWS's own): in-domain resources start untagged.
    tag_propagation = optional(string)

    # home_efs_retention_policy decides what happens to the domain's auto-created EFS file
    # system (user home directories) when the domain is deleted.
    # "Retain" (default, AWS's own): the EFS file system survives domain deletion. Home
    #   directories are preserved, but the file system keeps accruing storage charges and
    #   must be deleted by hand when no longer needed.
    # "Delete": the EFS file system is deleted with the domain -- the right choice for
    #   ephemeral and test domains, and the only choice that leaves nothing behind.
    # ForceNew: the retention decision is fixed at create time.
    home_efs_retention_policy = optional(string)

    # default_user_settings defines the default configuration inherited by all user profiles
    # in the domain. These settings control the execution environment, per-IDE app
    # configurations, security boundaries, and storage defaults. Individual user profiles
    # can override these defaults.
    default_user_settings = object({
      # execution_role_arn is the IAM role assumed by SageMaker when running notebooks,
      # training jobs, and other ML workloads on behalf of users. This role determines what
      # AWS resources (S3 buckets, ECR repos, Secrets Manager, etc.) users can access from
      # their Studio sessions. The role must have a trust policy allowing
      # sagemaker.amazonaws.com to assume it.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      execution_role_arn = string

      # security_group_ids are user-level security groups controlling network access for
      # notebook instances and apps. These are attached to the ENIs created for each user's
      # apps and control inbound/outbound traffic at the user level. Maximum 5 security groups.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      security_group_ids = optional(list(string), [])

      # default_landing_uri is the URI of the default app opened when a user accesses the domain.
      # Common values:
      #   "studio::relative/JupyterLab" - opens JupyterLab (recommended for most teams)
      #   "studio::relative/JupyterServer:" - opens classic Jupyter Server
      #   "studio::" - opens SageMaker Studio home
      # If omitted, AWS uses the platform default.
      default_landing_uri = optional(string, "")

      # studio_web_portal controls whether the SageMaker Studio web portal is accessible.
      # "ENABLED" (default): users can access the full Studio web interface.
      # "DISABLED": restricts access to programmatic-only usage (API/CLI).
      studio_web_portal = optional(string)

      # auto_mount_home_efs controls whether each user's home directory on the domain's EFS
      # file system is automatically mounted into their apps.
      # "Enabled": home directories mount automatically (the classic Studio experience).
      # "Disabled": apps start without the shared home directory -- pair with space storage
      #   or custom file systems when home directories are not wanted.
      # (AWS's third value, "DefaultAsDomain", is only valid on per-user profiles -- it means
      # "inherit this domain-level setting" and is rejected here at the domain level.)
      auto_mount_home_efs = optional(string)

      # jupyter_lab_app_settings configures JupyterLab, the primary IDE for SageMaker Studio.
      # JupyterLab provides a modern notebook and code editing experience with built-in Git
      # integration, terminal access, extension support, and collaborative editing.
      jupyter_lab_app_settings = optional(object({
        # default_resource_spec sets the default compute instance type and image configuration
        # for new JupyterLab apps. Users can override the instance type at app creation time.
        default_resource_spec = optional(object({
          # instance_type is the EC2 instance type for the app's compute.
          # Common choices:
          #   "ml.t3.medium" - development/exploration (2 vCPU, 4 GB)
          #   "ml.m5.large" - general-purpose notebooks (2 vCPU, 8 GB)
          #   "ml.g4dn.xlarge" - GPU workloads (1 GPU, 4 vCPU, 16 GB)
          #   "ml.p3.2xlarge" - heavy training (1 V100 GPU, 8 vCPU, 61 GB)
          #   "system" - lightweight system-managed instance for JupyterServer
          # The shape is checked here ("system" or an "ml."-prefixed type) so a typo
          # fails at manifest time; membership in AWS's instance-type list grows with
          # every hardware launch and is validated by AWS at deploy time.
          instance_type = optional(string, "")

          # lifecycle_config_arn is the ARN of a lifecycle configuration script that runs
          # when this specific app type starts.
          lifecycle_config_arn = optional(string, "")

          # sagemaker_image_arn is the ARN of a SageMaker Image that defines the app's container.
          # Use this to specify a custom base image instead of the SageMaker-provided default.
          sagemaker_image_arn = optional(string, "")

          # sagemaker_image_version_alias is a human-readable alias for a specific image version
          # (e.g., "latest", "v2.0"). Mutually exclusive with sagemaker_image_version_arn.
          sagemaker_image_version_alias = optional(string, "")

          # sagemaker_image_version_arn is the ARN of a specific SageMaker image version.
          # Pins the app to an exact image version for reproducibility across team members.
          # Mutually exclusive with sagemaker_image_version_alias.
          sagemaker_image_version_arn = optional(string, "")
        }))

        # lifecycle_config_arns are ARNs of lifecycle configuration scripts that run when
        # JupyterLab apps start. Use lifecycle configs to install Python packages, configure
        # JupyterLab extensions, set environment variables, or mount additional storage.
        lifecycle_config_arns = optional(list(string), [])

        # built_in_lifecycle_config_arn is the ARN of an AWS-curated (built-in) lifecycle
        # configuration to run at app start, as opposed to the customer-authored scripts in
        # lifecycle_config_arns.
        built_in_lifecycle_config_arn = optional(string, "")

        # custom_images are custom Docker images available as JupyterLab kernels.
        # Each image must be registered in SageMaker via an AppImageConfig that defines
        # the kernel specification and file system layout. Maximum 200 images.
        custom_images = optional(list(object({
          # app_image_config_name is the name of the SageMaker AppImageConfig that defines how
          # the image is presented to users (kernel specifications, file system configuration).
          # The AppImageConfig must exist before referencing it here.
          app_image_config_name = string

          # image_name is the name of the SageMaker Image resource that contains this container image.
          # The Image resource must exist before referencing it here.
          image_name = string

          # image_version_number pins to a specific version of the image.
          # If omitted, the latest available version is used.
          image_version_number = optional(number)
        })), [])

        # code_repositories are Git repositories automatically cloned into JupyterLab on startup.
        # Provides immediate access to team code, shared notebooks, and documentation.
        # Maximum 10 repositories.
        code_repositories = optional(list(object({
          # repository_url is the HTTPS URL of the Git repository to clone.
          # Must be an HTTPS URL (SSH URLs are not supported by SageMaker).
          # Examples: "https://github.com/org/ml-notebooks.git"
          # Maximum length: 1024 characters.
          repository_url = string
        })), [])

        # idle_settings configures automatic shutdown of idle JupyterLab instances.
        # Critical for cost management: without idle timeout, instances run 24/7 at full
        # compute cost even when no user is interacting with them.
        idle_settings = optional(object({
          # lifecycle_management is the enforcement switch: "ENABLED" (the default when
          # this block is present) turns automatic idle shutdown on; "DISABLED" keeps
          # the block's timeout values as published guardrails users may adopt WITHOUT
          # forcing auto-shutdown on — the defined-but-disabled state. Both engines
          # send the explicit value, so flipping to "DISABLED" genuinely turns
          # enforcement off.
          lifecycle_management = optional(string)

          # idle_timeout_in_minutes is the duration of inactivity (in minutes) before an instance
          # is automatically shut down. Range: 60-525600 (1 hour to 365 days).
          # A reasonable production default is 120 (2 hours).
          idle_timeout_in_minutes = number

          # min_idle_timeout_in_minutes sets the minimum idle timeout that individual users can
          # configure for their own apps. Prevents users from setting excessively short timeouts
          # that would cause disruptive shutdowns during brief pauses. Range: 60-525600.
          min_idle_timeout_in_minutes = number

          # max_idle_timeout_in_minutes sets the maximum idle timeout that individual users can
          # configure. Prevents users from effectively disabling idle shutdown by setting
          # extremely long timeouts. Range: 60-525600.
          max_idle_timeout_in_minutes = number
        }))

        # emr_settings pre-authorizes the IAM roles JupyterLab uses to discover and connect
        # to Amazon EMR clusters for large-scale data processing directly from notebooks.
        emr_settings = optional(object({
          # assumable_role_arns are IAM roles the JupyterLab app can assume to CONNECT to EMR
          # clusters -- including clusters in other AWS accounts (cross-account access).
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          assumable_role_arns = optional(list(string), [])

          # execution_role_arns are IAM runtime roles available for EMR workloads submitted from
          # the notebook (the role the EMR job itself runs under).
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          execution_role_arns = optional(list(string), [])
        }))
      }))

      # jupyter_server_app_settings configures the classic Jupyter Server app (Studio
      # Classic). Teams still running the previous-generation Studio experience configure
      # its default resources and startup repositories here.
      jupyter_server_app_settings = optional(object({
        # default_resource_spec sets the default configuration for the Jupyter Server app.
        # Jupyter Server runs on a lightweight system-managed instance ("system" instance type).
        default_resource_spec = optional(object({
          # instance_type is the EC2 instance type for the app's compute.
          # Common choices:
          #   "ml.t3.medium" - development/exploration (2 vCPU, 4 GB)
          #   "ml.m5.large" - general-purpose notebooks (2 vCPU, 8 GB)
          #   "ml.g4dn.xlarge" - GPU workloads (1 GPU, 4 vCPU, 16 GB)
          #   "ml.p3.2xlarge" - heavy training (1 V100 GPU, 8 vCPU, 61 GB)
          #   "system" - lightweight system-managed instance for JupyterServer
          # The shape is checked here ("system" or an "ml."-prefixed type) so a typo
          # fails at manifest time; membership in AWS's instance-type list grows with
          # every hardware launch and is validated by AWS at deploy time.
          instance_type = optional(string, "")

          # lifecycle_config_arn is the ARN of a lifecycle configuration script that runs
          # when this specific app type starts.
          lifecycle_config_arn = optional(string, "")

          # sagemaker_image_arn is the ARN of a SageMaker Image that defines the app's container.
          # Use this to specify a custom base image instead of the SageMaker-provided default.
          sagemaker_image_arn = optional(string, "")

          # sagemaker_image_version_alias is a human-readable alias for a specific image version
          # (e.g., "latest", "v2.0"). Mutually exclusive with sagemaker_image_version_arn.
          sagemaker_image_version_alias = optional(string, "")

          # sagemaker_image_version_arn is the ARN of a specific SageMaker image version.
          # Pins the app to an exact image version for reproducibility across team members.
          # Mutually exclusive with sagemaker_image_version_alias.
          sagemaker_image_version_arn = optional(string, "")
        }))

        # lifecycle_config_arns are ARNs of lifecycle configuration scripts for the
        # Jupyter Server app.
        lifecycle_config_arns = optional(list(string), [])

        # code_repositories are Git repositories automatically cloned on startup.
        # Maximum 10 repositories.
        code_repositories = optional(list(object({
          # repository_url is the HTTPS URL of the Git repository to clone.
          # Must be an HTTPS URL (SSH URLs are not supported by SageMaker).
          # Examples: "https://github.com/org/ml-notebooks.git"
          # Maximum length: 1024 characters.
          repository_url = string
        })), [])
      }))

      # kernel_gateway_app_settings configures KernelGateway apps that provide custom compute
      # kernels. Use KernelGateway to bring your own Docker images with specialized ML
      # frameworks, custom libraries, or GPU-optimized environments that go beyond the
      # standard SageMaker-provided kernels.
      kernel_gateway_app_settings = optional(object({
        # default_resource_spec sets the default compute instance type and image configuration
        # for new KernelGateway apps.
        default_resource_spec = optional(object({
          # instance_type is the EC2 instance type for the app's compute.
          # Common choices:
          #   "ml.t3.medium" - development/exploration (2 vCPU, 4 GB)
          #   "ml.m5.large" - general-purpose notebooks (2 vCPU, 8 GB)
          #   "ml.g4dn.xlarge" - GPU workloads (1 GPU, 4 vCPU, 16 GB)
          #   "ml.p3.2xlarge" - heavy training (1 V100 GPU, 8 vCPU, 61 GB)
          #   "system" - lightweight system-managed instance for JupyterServer
          # The shape is checked here ("system" or an "ml."-prefixed type) so a typo
          # fails at manifest time; membership in AWS's instance-type list grows with
          # every hardware launch and is validated by AWS at deploy time.
          instance_type = optional(string, "")

          # lifecycle_config_arn is the ARN of a lifecycle configuration script that runs
          # when this specific app type starts.
          lifecycle_config_arn = optional(string, "")

          # sagemaker_image_arn is the ARN of a SageMaker Image that defines the app's container.
          # Use this to specify a custom base image instead of the SageMaker-provided default.
          sagemaker_image_arn = optional(string, "")

          # sagemaker_image_version_alias is a human-readable alias for a specific image version
          # (e.g., "latest", "v2.0"). Mutually exclusive with sagemaker_image_version_arn.
          sagemaker_image_version_alias = optional(string, "")

          # sagemaker_image_version_arn is the ARN of a specific SageMaker image version.
          # Pins the app to an exact image version for reproducibility across team members.
          # Mutually exclusive with sagemaker_image_version_alias.
          sagemaker_image_version_arn = optional(string, "")
        }))

        # lifecycle_config_arns are ARNs of lifecycle configuration scripts for KernelGateway apps.
        lifecycle_config_arns = optional(list(string), [])

        # custom_images are custom Docker images available as KernelGateway kernels.
        # Each image must be registered in SageMaker via an AppImageConfig. Maximum 200 images.
        custom_images = optional(list(object({
          # app_image_config_name is the name of the SageMaker AppImageConfig that defines how
          # the image is presented to users (kernel specifications, file system configuration).
          # The AppImageConfig must exist before referencing it here.
          app_image_config_name = string

          # image_name is the name of the SageMaker Image resource that contains this container image.
          # The Image resource must exist before referencing it here.
          image_name = string

          # image_version_number pins to a specific version of the image.
          # If omitted, the latest available version is used.
          image_version_number = optional(number)
        })), [])
      }))

      # code_editor_app_settings configures the Code Editor app -- SageMaker's VS Code
      # (Code-OSS) IDE. It carries the same resource-spec/lifecycle/custom-image/idle
      # controls as JupyterLab, for teams that prefer a full code editor over notebooks.
      code_editor_app_settings = optional(object({
        # default_resource_spec sets the default compute instance type and image configuration
        # for new Code Editor apps.
        default_resource_spec = optional(object({
          # instance_type is the EC2 instance type for the app's compute.
          # Common choices:
          #   "ml.t3.medium" - development/exploration (2 vCPU, 4 GB)
          #   "ml.m5.large" - general-purpose notebooks (2 vCPU, 8 GB)
          #   "ml.g4dn.xlarge" - GPU workloads (1 GPU, 4 vCPU, 16 GB)
          #   "ml.p3.2xlarge" - heavy training (1 V100 GPU, 8 vCPU, 61 GB)
          #   "system" - lightweight system-managed instance for JupyterServer
          # The shape is checked here ("system" or an "ml."-prefixed type) so a typo
          # fails at manifest time; membership in AWS's instance-type list grows with
          # every hardware launch and is validated by AWS at deploy time.
          instance_type = optional(string, "")

          # lifecycle_config_arn is the ARN of a lifecycle configuration script that runs
          # when this specific app type starts.
          lifecycle_config_arn = optional(string, "")

          # sagemaker_image_arn is the ARN of a SageMaker Image that defines the app's container.
          # Use this to specify a custom base image instead of the SageMaker-provided default.
          sagemaker_image_arn = optional(string, "")

          # sagemaker_image_version_alias is a human-readable alias for a specific image version
          # (e.g., "latest", "v2.0"). Mutually exclusive with sagemaker_image_version_arn.
          sagemaker_image_version_alias = optional(string, "")

          # sagemaker_image_version_arn is the ARN of a specific SageMaker image version.
          # Pins the app to an exact image version for reproducibility across team members.
          # Mutually exclusive with sagemaker_image_version_alias.
          sagemaker_image_version_arn = optional(string, "")
        }))

        # lifecycle_config_arns are ARNs of lifecycle configuration scripts for Code Editor apps.
        lifecycle_config_arns = optional(list(string), [])

        # built_in_lifecycle_config_arn is the ARN of an AWS-curated (built-in) lifecycle
        # configuration to run at app start.
        built_in_lifecycle_config_arn = optional(string, "")

        # custom_images are custom Docker images available in Code Editor. Maximum 200 images.
        custom_images = optional(list(object({
          # app_image_config_name is the name of the SageMaker AppImageConfig that defines how
          # the image is presented to users (kernel specifications, file system configuration).
          # The AppImageConfig must exist before referencing it here.
          app_image_config_name = string

          # image_name is the name of the SageMaker Image resource that contains this container image.
          # The Image resource must exist before referencing it here.
          image_name = string

          # image_version_number pins to a specific version of the image.
          # If omitted, the latest available version is used.
          image_version_number = optional(number)
        })), [])

        # idle_settings configures automatic shutdown of idle Code Editor instances -- the
        # same cost-control dial JupyterLab carries.
        idle_settings = optional(object({
          # lifecycle_management is the enforcement switch: "ENABLED" (the default when
          # this block is present) turns automatic idle shutdown on; "DISABLED" keeps
          # the block's timeout values as published guardrails users may adopt WITHOUT
          # forcing auto-shutdown on — the defined-but-disabled state. Both engines
          # send the explicit value, so flipping to "DISABLED" genuinely turns
          # enforcement off.
          lifecycle_management = optional(string)

          # idle_timeout_in_minutes is the duration of inactivity (in minutes) before an instance
          # is automatically shut down. Range: 60-525600 (1 hour to 365 days).
          # A reasonable production default is 120 (2 hours).
          idle_timeout_in_minutes = number

          # min_idle_timeout_in_minutes sets the minimum idle timeout that individual users can
          # configure for their own apps. Prevents users from setting excessively short timeouts
          # that would cause disruptive shutdowns during brief pauses. Range: 60-525600.
          min_idle_timeout_in_minutes = number

          # max_idle_timeout_in_minutes sets the maximum idle timeout that individual users can
          # configure. Prevents users from effectively disabling idle shutdown by setting
          # extremely long timeouts. Range: 60-525600.
          max_idle_timeout_in_minutes = number
        }))
      }))

      # tensor_board_app_settings configures the TensorBoard app used to visualize
      # training runs. Only the default resource spec is configurable.
      tensor_board_app_settings = optional(object({
        # default_resource_spec sets the default compute instance type and image configuration
        # for the TensorBoard app.
        default_resource_spec = optional(object({
          # instance_type is the EC2 instance type for the app's compute.
          # Common choices:
          #   "ml.t3.medium" - development/exploration (2 vCPU, 4 GB)
          #   "ml.m5.large" - general-purpose notebooks (2 vCPU, 8 GB)
          #   "ml.g4dn.xlarge" - GPU workloads (1 GPU, 4 vCPU, 16 GB)
          #   "ml.p3.2xlarge" - heavy training (1 V100 GPU, 8 vCPU, 61 GB)
          #   "system" - lightweight system-managed instance for JupyterServer
          # The shape is checked here ("system" or an "ml."-prefixed type) so a typo
          # fails at manifest time; membership in AWS's instance-type list grows with
          # every hardware launch and is validated by AWS at deploy time.
          instance_type = optional(string, "")

          # lifecycle_config_arn is the ARN of a lifecycle configuration script that runs
          # when this specific app type starts.
          lifecycle_config_arn = optional(string, "")

          # sagemaker_image_arn is the ARN of a SageMaker Image that defines the app's container.
          # Use this to specify a custom base image instead of the SageMaker-provided default.
          sagemaker_image_arn = optional(string, "")

          # sagemaker_image_version_alias is a human-readable alias for a specific image version
          # (e.g., "latest", "v2.0"). Mutually exclusive with sagemaker_image_version_arn.
          sagemaker_image_version_alias = optional(string, "")

          # sagemaker_image_version_arn is the ARN of a specific SageMaker image version.
          # Pins the app to an exact image version for reproducibility across team members.
          # Mutually exclusive with sagemaker_image_version_alias.
          sagemaker_image_version_arn = optional(string, "")
        }))
      }))

      # r_session_app_settings configures RSession apps (the R kernels backing RStudio
      # sessions). Requires RStudio to be enabled on the domain via
      # r_studio_server_pro_domain_settings.
      r_session_app_settings = optional(object({
        # default_resource_spec sets the default compute instance type and image configuration
        # for RSession apps.
        default_resource_spec = optional(object({
          # instance_type is the EC2 instance type for the app's compute.
          # Common choices:
          #   "ml.t3.medium" - development/exploration (2 vCPU, 4 GB)
          #   "ml.m5.large" - general-purpose notebooks (2 vCPU, 8 GB)
          #   "ml.g4dn.xlarge" - GPU workloads (1 GPU, 4 vCPU, 16 GB)
          #   "ml.p3.2xlarge" - heavy training (1 V100 GPU, 8 vCPU, 61 GB)
          #   "system" - lightweight system-managed instance for JupyterServer
          # The shape is checked here ("system" or an "ml."-prefixed type) so a typo
          # fails at manifest time; membership in AWS's instance-type list grows with
          # every hardware launch and is validated by AWS at deploy time.
          instance_type = optional(string, "")

          # lifecycle_config_arn is the ARN of a lifecycle configuration script that runs
          # when this specific app type starts.
          lifecycle_config_arn = optional(string, "")

          # sagemaker_image_arn is the ARN of a SageMaker Image that defines the app's container.
          # Use this to specify a custom base image instead of the SageMaker-provided default.
          sagemaker_image_arn = optional(string, "")

          # sagemaker_image_version_alias is a human-readable alias for a specific image version
          # (e.g., "latest", "v2.0"). Mutually exclusive with sagemaker_image_version_arn.
          sagemaker_image_version_alias = optional(string, "")

          # sagemaker_image_version_arn is the ARN of a specific SageMaker image version.
          # Pins the app to an exact image version for reproducibility across team members.
          # Mutually exclusive with sagemaker_image_version_alias.
          sagemaker_image_version_arn = optional(string, "")
        }))

        # custom_images are custom Docker images available as RSession kernels. Maximum 200 images.
        custom_images = optional(list(object({
          # app_image_config_name is the name of the SageMaker AppImageConfig that defines how
          # the image is presented to users (kernel specifications, file system configuration).
          # The AppImageConfig must exist before referencing it here.
          app_image_config_name = string

          # image_name is the name of the SageMaker Image resource that contains this container image.
          # The Image resource must exist before referencing it here.
          image_name = string

          # image_version_number pins to a specific version of the image.
          # If omitted, the latest available version is used.
          image_version_number = optional(number)
        })), [])
      }))

      # r_studio_server_pro_app_settings controls per-user RStudio Server Pro access.
      # Requires RStudio to be enabled on the domain via r_studio_server_pro_domain_settings.
      r_studio_server_pro_app_settings = optional(object({
        # access_status grants or denies the user access to RStudio Server Pro.
        # "ENABLED": the user sees and can launch RStudio.
        # "DISABLED": RStudio is hidden for the user.
        access_status = optional(string, "")

        # user_group assigns the RStudio authorization level.
        # "R_STUDIO_ADMIN": administrative access to the RStudio Workbench admin dashboard.
        # "R_STUDIO_USER" (AWS default): regular RStudio user.
        # Only meaningful when access_status is "ENABLED" -- AWS ignores it otherwise, so the
        # spec rejects the dead combination up front.
        user_group = optional(string, "")
      }))

      # canvas_app_settings configures SageMaker Canvas, the no-code ML workspace.
      # Each sub-block is an independent Canvas capability (direct model deployment,
      # EMR Serverless big-data processing, Bedrock generative AI, SaaS data connectors,
      # Kendra document search, cross-account model registration, time-series forecasting,
      # and the shared artifact workspace).
      canvas_app_settings = optional(object({
        # direct_deploy_status controls whether Canvas users can deploy models they build
        # directly to SageMaker real-time endpoints ("ENABLED" or "DISABLED"). Direct
        # deployment creates billable endpoints, so governance-minded teams often disable it
        # and route models through the registry instead (model_register_settings).
        direct_deploy_status = optional(string)

        # emr_serverless_settings lets Canvas run large data preparation and processing jobs
        # on EMR Serverless.
        emr_serverless_settings = optional(object({
          # execution_role_arn is the IAM role Canvas uses to submit and manage EMR Serverless
          # jobs.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          execution_role_arn = optional(string, "")

          # status enables or disables the capability ("ENABLED" or "DISABLED").
          status = optional(string, "")
        }))

        # generative_ai_bedrock_role_arn is the IAM role Canvas assumes to call Amazon Bedrock
        # for generative-AI features. Setting the role is what enables the capability.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        generative_ai_bedrock_role_arn = optional(string, "")

        # identity_provider_oauth_settings wire Canvas to external SaaS data sources
        # (Salesforce Data Cloud, Snowflake) through OAuth. Each entry names the data source
        # and the Secrets Manager secret holding its OAuth client credentials. Maximum 20.
        identity_provider_oauth_settings = optional(list(object({
          # data_source_name identifies the SaaS data source.
          # "SalesforceGenie": Salesforce Data Cloud.
          # "Snowflake": Snowflake.
          data_source_name = optional(string, "")

          # secret_arn is the Secrets Manager secret holding the data source's OAuth client
          # credentials (client ID and secret). The ARN is a reference Canvas resolves at
          # connection time -- never secret material itself.
          secret_arn = string

          # status enables or disables this connector ("ENABLED" or "DISABLED").
          status = optional(string, "")
        })), [])

        # kendra_settings_status controls whether Canvas can query Amazon Kendra indexes for
        # document search ("ENABLED" or "DISABLED").
        kendra_settings_status = optional(string)

        # model_register_settings controls whether Canvas users can register their models into
        # a SageMaker Model Registry, optionally in another AWS account.
        model_register_settings = optional(object({
          # cross_account_model_register_role_arn is the IAM role Canvas assumes to register
          # models into a Model Registry that lives in ANOTHER AWS account. Leave unset to
          # register into this account's registry.
          cross_account_model_register_role_arn = optional(string, "")

          # status enables or disables model registration ("ENABLED" or "DISABLED").
          status = optional(string, "")
        }))

        # time_series_forecasting_settings enables Canvas time-series forecasting, which uses
        # Amazon Forecast under the hood via the given IAM role.
        time_series_forecasting_settings = optional(object({
          # amazon_forecast_role_arn is the IAM role Canvas assumes to call Amazon Forecast.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          amazon_forecast_role_arn = optional(string, "")

          # status enables or disables forecasting ("ENABLED" or "DISABLED").
          status = optional(string, "")
        }))

        # workspace_settings pins the S3 location (and optional KMS key) where Canvas stores
        # its working artifacts -- datasets, intermediate results, generated models.
        workspace_settings = optional(object({
          # s3_artifact_path is the S3 URI where Canvas stores datasets, intermediate results,
          # and generated models. Example: "s3://my-canvas-workspace/artifacts/".
          s3_artifact_path = optional(string, "")

          # s3_kms_key_id is the KMS key used to encrypt Canvas artifacts in S3.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          s3_kms_key_id = optional(string, "")
        }))
      }))

      # sharing_settings controls notebook output sharing to S3. When enabled, notebook cell
      # outputs are persisted to an S3 location, allowing team members to view results
      # without running the notebook.
      sharing_settings = optional(object({
        # notebook_output_option controls whether notebook cell outputs are persisted to S3.
        # "Allowed": outputs are copied to S3 at the location specified by s3_output_path.
        # "Disabled" (default): outputs are not shared externally.
        notebook_output_option = optional(string)

        # s3_kms_key_id is the KMS key used to encrypt shared notebook outputs in S3.
        # If omitted, outputs are encrypted with the default S3 bucket encryption.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        s3_kms_key_id = optional(string, "")

        # s3_output_path is the S3 URI where shared notebook outputs are stored.
        # Required when notebook_output_option is "Allowed".
        # Example: "s3://my-team-bucket/notebook-outputs/"
        s3_output_path = optional(string, "")
      }))

      # space_storage_settings configures default EBS volume sizes for user spaces.
      # Spaces use EBS volumes for working storage beyond the shared EFS home directory.
      space_storage_settings = optional(object({
        # default_ebs_volume_size_in_gb is the default EBS volume size (in GB) assigned to new spaces.
        default_ebs_volume_size_in_gb = number

        # maximum_ebs_volume_size_in_gb is the maximum EBS volume size (in GB) that users can request
        # for their spaces. Must be >= default_ebs_volume_size_in_gb.
        maximum_ebs_volume_size_in_gb = number
      }))

      # custom_file_system_configs mount additional file systems (beyond the domain's own
      # EFS home directories) into every user's apps -- shared datasets, feature stores,
      # model artifact trees. Each entry names one file system and the path where it mounts.
      custom_file_system_configs = optional(list(object({
        # efs_file_system_config mounts an Amazon EFS file system.
        efs_file_system_config = object({
          # file_system_id is the EFS file system to mount. The file system must be reachable
          # from the domain's subnets (mount targets + security groups are the file system's
          # own configuration).
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          file_system_id = string

          # file_system_path is the path within the EFS file system to mount into apps.
          # Example: "/shared/datasets"
          file_system_path = string
        })
      })), [])

      # custom_posix_user_config sets the POSIX identity (UID/GID) that apps run as when
      # accessing the EFS home directory and custom file systems. Set this when file-system
      # permissions on shared storage must map to a specific owner instead of SageMaker's
      # default identity.
      custom_posix_user_config = optional(object({
        # uid is the POSIX user ID. Must be at least 10000.
        uid = number

        # gid is the POSIX group ID. Must be at least 1001.
        gid = number
      }))

      # studio_web_portal_settings hides parts of the Studio UI from users -- entire app
      # types, specific instance types, or ML tools. Use it to keep expensive GPU instance
      # types or unused tooling out of the picker instead of policing them after the fact.
      studio_web_portal_settings = optional(object({
        # hidden_app_types are Studio app types to hide (e.g. "JupyterServer", "KernelGateway",
        # "Canvas", "CodeEditor", "JupyterLab", "TensorBoard", "RStudioServerPro"). Values are
        # SageMaker AppType names; AWS validates them at deploy time as the set grows.
        hidden_app_types = optional(list(string), [])

        # hidden_instance_types are instance types to hide from app-creation pickers
        # (e.g. "ml.p3.2xlarge"). Values are SageMaker app instance type names.
        hidden_instance_types = optional(list(string), [])

        # hidden_ml_tools are Studio ML tools to hide (e.g. "DataWrangler", "FeatureStore",
        # "EmrClusters", "AutoMl", "Experiments", "Pipelines"). Values are SageMaker MlTools
        # names; AWS validates them at deploy time as the set grows.
        hidden_ml_tools = optional(list(string), [])
      }))
    })

    # default_space_settings defines the default configuration inherited by all shared
    # (collaborative) spaces in the domain. Spaces are workspaces multiple users can attach
    # to; they carry their own execution role and app baselines, separate from the per-user
    # plane above.
    default_space_settings = optional(object({
      # execution_role_arn is the IAM role assumed by SageMaker for workloads running in
      # shared spaces. Like the user-level role, it must trust sagemaker.amazonaws.com.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      execution_role_arn = string

      # security_group_ids are security groups attached to the ENIs of apps running in
      # shared spaces. Maximum 5 security groups.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      security_group_ids = optional(list(string), [])

      # jupyter_lab_app_settings is the JupyterLab baseline for shared spaces (same shape
      # as the user-level block).
      jupyter_lab_app_settings = optional(object({
        # default_resource_spec sets the default compute instance type and image configuration
        # for new JupyterLab apps. Users can override the instance type at app creation time.
        default_resource_spec = optional(object({
          # instance_type is the EC2 instance type for the app's compute.
          # Common choices:
          #   "ml.t3.medium" - development/exploration (2 vCPU, 4 GB)
          #   "ml.m5.large" - general-purpose notebooks (2 vCPU, 8 GB)
          #   "ml.g4dn.xlarge" - GPU workloads (1 GPU, 4 vCPU, 16 GB)
          #   "ml.p3.2xlarge" - heavy training (1 V100 GPU, 8 vCPU, 61 GB)
          #   "system" - lightweight system-managed instance for JupyterServer
          # The shape is checked here ("system" or an "ml."-prefixed type) so a typo
          # fails at manifest time; membership in AWS's instance-type list grows with
          # every hardware launch and is validated by AWS at deploy time.
          instance_type = optional(string, "")

          # lifecycle_config_arn is the ARN of a lifecycle configuration script that runs
          # when this specific app type starts.
          lifecycle_config_arn = optional(string, "")

          # sagemaker_image_arn is the ARN of a SageMaker Image that defines the app's container.
          # Use this to specify a custom base image instead of the SageMaker-provided default.
          sagemaker_image_arn = optional(string, "")

          # sagemaker_image_version_alias is a human-readable alias for a specific image version
          # (e.g., "latest", "v2.0"). Mutually exclusive with sagemaker_image_version_arn.
          sagemaker_image_version_alias = optional(string, "")

          # sagemaker_image_version_arn is the ARN of a specific SageMaker image version.
          # Pins the app to an exact image version for reproducibility across team members.
          # Mutually exclusive with sagemaker_image_version_alias.
          sagemaker_image_version_arn = optional(string, "")
        }))

        # lifecycle_config_arns are ARNs of lifecycle configuration scripts that run when
        # JupyterLab apps start. Use lifecycle configs to install Python packages, configure
        # JupyterLab extensions, set environment variables, or mount additional storage.
        lifecycle_config_arns = optional(list(string), [])

        # built_in_lifecycle_config_arn is the ARN of an AWS-curated (built-in) lifecycle
        # configuration to run at app start, as opposed to the customer-authored scripts in
        # lifecycle_config_arns.
        built_in_lifecycle_config_arn = optional(string, "")

        # custom_images are custom Docker images available as JupyterLab kernels.
        # Each image must be registered in SageMaker via an AppImageConfig that defines
        # the kernel specification and file system layout. Maximum 200 images.
        custom_images = optional(list(object({
          # app_image_config_name is the name of the SageMaker AppImageConfig that defines how
          # the image is presented to users (kernel specifications, file system configuration).
          # The AppImageConfig must exist before referencing it here.
          app_image_config_name = string

          # image_name is the name of the SageMaker Image resource that contains this container image.
          # The Image resource must exist before referencing it here.
          image_name = string

          # image_version_number pins to a specific version of the image.
          # If omitted, the latest available version is used.
          image_version_number = optional(number)
        })), [])

        # code_repositories are Git repositories automatically cloned into JupyterLab on startup.
        # Provides immediate access to team code, shared notebooks, and documentation.
        # Maximum 10 repositories.
        code_repositories = optional(list(object({
          # repository_url is the HTTPS URL of the Git repository to clone.
          # Must be an HTTPS URL (SSH URLs are not supported by SageMaker).
          # Examples: "https://github.com/org/ml-notebooks.git"
          # Maximum length: 1024 characters.
          repository_url = string
        })), [])

        # idle_settings configures automatic shutdown of idle JupyterLab instances.
        # Critical for cost management: without idle timeout, instances run 24/7 at full
        # compute cost even when no user is interacting with them.
        idle_settings = optional(object({
          # lifecycle_management is the enforcement switch: "ENABLED" (the default when
          # this block is present) turns automatic idle shutdown on; "DISABLED" keeps
          # the block's timeout values as published guardrails users may adopt WITHOUT
          # forcing auto-shutdown on — the defined-but-disabled state. Both engines
          # send the explicit value, so flipping to "DISABLED" genuinely turns
          # enforcement off.
          lifecycle_management = optional(string)

          # idle_timeout_in_minutes is the duration of inactivity (in minutes) before an instance
          # is automatically shut down. Range: 60-525600 (1 hour to 365 days).
          # A reasonable production default is 120 (2 hours).
          idle_timeout_in_minutes = number

          # min_idle_timeout_in_minutes sets the minimum idle timeout that individual users can
          # configure for their own apps. Prevents users from setting excessively short timeouts
          # that would cause disruptive shutdowns during brief pauses. Range: 60-525600.
          min_idle_timeout_in_minutes = number

          # max_idle_timeout_in_minutes sets the maximum idle timeout that individual users can
          # configure. Prevents users from effectively disabling idle shutdown by setting
          # extremely long timeouts. Range: 60-525600.
          max_idle_timeout_in_minutes = number
        }))

        # emr_settings pre-authorizes the IAM roles JupyterLab uses to discover and connect
        # to Amazon EMR clusters for large-scale data processing directly from notebooks.
        emr_settings = optional(object({
          # assumable_role_arns are IAM roles the JupyterLab app can assume to CONNECT to EMR
          # clusters -- including clusters in other AWS accounts (cross-account access).
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          assumable_role_arns = optional(list(string), [])

          # execution_role_arns are IAM runtime roles available for EMR workloads submitted from
          # the notebook (the role the EMR job itself runs under).
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          execution_role_arns = optional(list(string), [])
        }))
      }))

      # jupyter_server_app_settings is the classic Jupyter Server baseline for shared spaces.
      jupyter_server_app_settings = optional(object({
        # default_resource_spec sets the default configuration for the Jupyter Server app.
        # Jupyter Server runs on a lightweight system-managed instance ("system" instance type).
        default_resource_spec = optional(object({
          # instance_type is the EC2 instance type for the app's compute.
          # Common choices:
          #   "ml.t3.medium" - development/exploration (2 vCPU, 4 GB)
          #   "ml.m5.large" - general-purpose notebooks (2 vCPU, 8 GB)
          #   "ml.g4dn.xlarge" - GPU workloads (1 GPU, 4 vCPU, 16 GB)
          #   "ml.p3.2xlarge" - heavy training (1 V100 GPU, 8 vCPU, 61 GB)
          #   "system" - lightweight system-managed instance for JupyterServer
          # The shape is checked here ("system" or an "ml."-prefixed type) so a typo
          # fails at manifest time; membership in AWS's instance-type list grows with
          # every hardware launch and is validated by AWS at deploy time.
          instance_type = optional(string, "")

          # lifecycle_config_arn is the ARN of a lifecycle configuration script that runs
          # when this specific app type starts.
          lifecycle_config_arn = optional(string, "")

          # sagemaker_image_arn is the ARN of a SageMaker Image that defines the app's container.
          # Use this to specify a custom base image instead of the SageMaker-provided default.
          sagemaker_image_arn = optional(string, "")

          # sagemaker_image_version_alias is a human-readable alias for a specific image version
          # (e.g., "latest", "v2.0"). Mutually exclusive with sagemaker_image_version_arn.
          sagemaker_image_version_alias = optional(string, "")

          # sagemaker_image_version_arn is the ARN of a specific SageMaker image version.
          # Pins the app to an exact image version for reproducibility across team members.
          # Mutually exclusive with sagemaker_image_version_alias.
          sagemaker_image_version_arn = optional(string, "")
        }))

        # lifecycle_config_arns are ARNs of lifecycle configuration scripts for the
        # Jupyter Server app.
        lifecycle_config_arns = optional(list(string), [])

        # code_repositories are Git repositories automatically cloned on startup.
        # Maximum 10 repositories.
        code_repositories = optional(list(object({
          # repository_url is the HTTPS URL of the Git repository to clone.
          # Must be an HTTPS URL (SSH URLs are not supported by SageMaker).
          # Examples: "https://github.com/org/ml-notebooks.git"
          # Maximum length: 1024 characters.
          repository_url = string
        })), [])
      }))

      # kernel_gateway_app_settings is the KernelGateway baseline for shared spaces.
      kernel_gateway_app_settings = optional(object({
        # default_resource_spec sets the default compute instance type and image configuration
        # for new KernelGateway apps.
        default_resource_spec = optional(object({
          # instance_type is the EC2 instance type for the app's compute.
          # Common choices:
          #   "ml.t3.medium" - development/exploration (2 vCPU, 4 GB)
          #   "ml.m5.large" - general-purpose notebooks (2 vCPU, 8 GB)
          #   "ml.g4dn.xlarge" - GPU workloads (1 GPU, 4 vCPU, 16 GB)
          #   "ml.p3.2xlarge" - heavy training (1 V100 GPU, 8 vCPU, 61 GB)
          #   "system" - lightweight system-managed instance for JupyterServer
          # The shape is checked here ("system" or an "ml."-prefixed type) so a typo
          # fails at manifest time; membership in AWS's instance-type list grows with
          # every hardware launch and is validated by AWS at deploy time.
          instance_type = optional(string, "")

          # lifecycle_config_arn is the ARN of a lifecycle configuration script that runs
          # when this specific app type starts.
          lifecycle_config_arn = optional(string, "")

          # sagemaker_image_arn is the ARN of a SageMaker Image that defines the app's container.
          # Use this to specify a custom base image instead of the SageMaker-provided default.
          sagemaker_image_arn = optional(string, "")

          # sagemaker_image_version_alias is a human-readable alias for a specific image version
          # (e.g., "latest", "v2.0"). Mutually exclusive with sagemaker_image_version_arn.
          sagemaker_image_version_alias = optional(string, "")

          # sagemaker_image_version_arn is the ARN of a specific SageMaker image version.
          # Pins the app to an exact image version for reproducibility across team members.
          # Mutually exclusive with sagemaker_image_version_alias.
          sagemaker_image_version_arn = optional(string, "")
        }))

        # lifecycle_config_arns are ARNs of lifecycle configuration scripts for KernelGateway apps.
        lifecycle_config_arns = optional(list(string), [])

        # custom_images are custom Docker images available as KernelGateway kernels.
        # Each image must be registered in SageMaker via an AppImageConfig. Maximum 200 images.
        custom_images = optional(list(object({
          # app_image_config_name is the name of the SageMaker AppImageConfig that defines how
          # the image is presented to users (kernel specifications, file system configuration).
          # The AppImageConfig must exist before referencing it here.
          app_image_config_name = string

          # image_name is the name of the SageMaker Image resource that contains this container image.
          # The Image resource must exist before referencing it here.
          image_name = string

          # image_version_number pins to a specific version of the image.
          # If omitted, the latest available version is used.
          image_version_number = optional(number)
        })), [])
      }))

      # space_storage_settings configures the default and maximum EBS volume sizes for
      # shared spaces.
      space_storage_settings = optional(object({
        # default_ebs_volume_size_in_gb is the default EBS volume size (in GB) assigned to new spaces.
        default_ebs_volume_size_in_gb = number

        # maximum_ebs_volume_size_in_gb is the maximum EBS volume size (in GB) that users can request
        # for their spaces. Must be >= default_ebs_volume_size_in_gb.
        maximum_ebs_volume_size_in_gb = number
      }))

      # custom_file_system_configs mount additional file systems into every shared space's
      # apps (same shape as the user-level block).
      custom_file_system_configs = optional(list(object({
        # efs_file_system_config mounts an Amazon EFS file system.
        efs_file_system_config = object({
          # file_system_id is the EFS file system to mount. The file system must be reachable
          # from the domain's subnets (mount targets + security groups are the file system's
          # own configuration).
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          file_system_id = string

          # file_system_path is the path within the EFS file system to mount into apps.
          # Example: "/shared/datasets"
          file_system_path = string
        })
      })), [])

      # custom_posix_user_config sets the POSIX identity that shared-space apps run as.
      custom_posix_user_config = optional(object({
        # uid is the POSIX user ID. Must be at least 10000.
        uid = number

        # gid is the POSIX group ID. Must be at least 1001.
        gid = number
      }))
    }))

    # domain_security_group_ids are security groups applied at the domain level for
    # domain-scoped apps and shared resources. These are separate from user-level security
    # groups (default_user_settings.security_group_ids) and control domain-wide network
    # boundaries. Maximum 3 security groups.
    # ForceNew: changing these forces domain replacement.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    domain_security_group_ids = optional(list(string), [])

    # docker_settings controls Docker access within the SageMaker Domain.
    # When enabled, users can build and run custom Docker containers directly in their
    # notebooks and terminals -- essential for custom training containers, inference
    # endpoints, and reproducible ML pipelines.
    docker_settings = optional(object({
      # enable_docker_access controls whether users can build and run Docker containers
      # in their notebooks and terminals.
      # "ENABLED": Docker commands (build, pull, run) are available.
      # "DISABLED": Docker access is blocked.
      enable_docker_access = optional(string, "")

      # vpc_only_trusted_accounts restricts Docker image pulling to images from specified
      # AWS account IDs when app_network_access_type is "VpcOnly". This prevents users from
      # pulling arbitrary images from public registries, enforcing approved image sources.
      # Maximum 20 account IDs, each exactly 12 digits — a malformed entry would make this
      # security control silently not match the intended account.
      vpc_only_trusted_accounts = optional(list(string), [])
    }))

    # execution_role_identity_config controls how user identity appears in AWS CloudTrail
    # and in the credentials SageMaker vends to apps.
    # "USER_PROFILE_NAME": the sts:SourceIdentity of every session is set to the user
    #   profile name, so CloudTrail events and IAM policies can distinguish WHICH user
    #   acted through the shared execution role -- strongly recommended for auditability.
    # "DISABLED": sessions carry only the execution role identity.
    execution_role_identity_config = optional(string)

    # r_studio_server_pro_domain_settings enables RStudio (Posit) Workbench on the domain.
    # RStudio on SageMaker requires a Posit license purchased through AWS License Manager;
    # configuring this block is what activates the RStudio app plane for the domain.
    r_studio_server_pro_domain_settings = optional(object({
      # domain_execution_role_arn is the IAM role SageMaker assumes to run the RStudio
      # server for the domain (license validation, server lifecycle). Distinct from the
      # per-user execution role.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      domain_execution_role_arn = string

      # r_studio_connect_url is the URL of an RStudio Connect server where users publish
      # Shiny apps and R Markdown documents from their sessions.
      r_studio_connect_url = optional(string, "")

      # r_studio_package_manager_url is the URL of an RStudio Package Manager server that
      # sessions resolve R packages from (instead of public CRAN).
      r_studio_package_manager_url = optional(string, "")

      # default_resource_spec sets the compute instance type and image configuration for the
      # RStudio server app itself.
      default_resource_spec = optional(object({
        # instance_type is the EC2 instance type for the app's compute.
        # Common choices:
        #   "ml.t3.medium" - development/exploration (2 vCPU, 4 GB)
        #   "ml.m5.large" - general-purpose notebooks (2 vCPU, 8 GB)
        #   "ml.g4dn.xlarge" - GPU workloads (1 GPU, 4 vCPU, 16 GB)
        #   "ml.p3.2xlarge" - heavy training (1 V100 GPU, 8 vCPU, 61 GB)
        #   "system" - lightweight system-managed instance for JupyterServer
        # The shape is checked here ("system" or an "ml."-prefixed type) so a typo
        # fails at manifest time; membership in AWS's instance-type list grows with
        # every hardware launch and is validated by AWS at deploy time.
        instance_type = optional(string, "")

        # lifecycle_config_arn is the ARN of a lifecycle configuration script that runs
        # when this specific app type starts.
        lifecycle_config_arn = optional(string, "")

        # sagemaker_image_arn is the ARN of a SageMaker Image that defines the app's container.
        # Use this to specify a custom base image instead of the SageMaker-provided default.
        sagemaker_image_arn = optional(string, "")

        # sagemaker_image_version_alias is a human-readable alias for a specific image version
        # (e.g., "latest", "v2.0"). Mutually exclusive with sagemaker_image_version_arn.
        sagemaker_image_version_alias = optional(string, "")

        # sagemaker_image_version_arn is the ARN of a specific SageMaker image version.
        # Pins the app to an exact image version for reproducibility across team members.
        # Mutually exclusive with sagemaker_image_version_alias.
        sagemaker_image_version_arn = optional(string, "")
      }))
    }))

    # trusted_identity_propagation_status enables trusted identity propagation, which
    # forwards the IAM Identity Center user identity through SageMaker to downstream
    # AWS analytics services (Athena, Redshift, Lake Formation), so data-level permissions
    # apply per human user instead of per execution role.
    # Requires auth_mode "SSO" with ANY value: AWS rejects the setting outright on
    # IAM-auth domains ("TrustedIdentityPropagationSettings is only supported for
    # Domains with AWS IAM Identity Center enabled"), even when set to "DISABLED".
    # Leave unset on IAM domains.
    trusted_identity_propagation_status = optional(string)

    # user_profiles are the per-person workspaces inside the domain, keyed by
    # name. Each profile inherits `default_user_settings` and may override any
    # of it via its own `user_settings`. Profiles add/remove in place as the
    # list changes; removing an entry deletes that profile AND its apps and
    # data surfaces, so treat removals as destructive. Profiles created
    # outside this manifest (IAM Identity Center auto-provisioning, console)
    # are not managed or removed by it.
    user_profiles = optional(list(object({
      # The profile name — unique within the domain, and the key both IaC modules
      # use for the satellite resource. 1-63 characters: alphanumeric and hyphens,
      # starting and ending alphanumeric. ForceNew: renaming replaces the profile
      # (and its home directory association).
      user_profile_name = string

      # For SSO-auth domains: the IAM Identity Center attribute that identifies the
      # user. AWS supports exactly one identifier scheme, "UserName". Set together
      # with `single_sign_on_user_value`. ForceNew.
      single_sign_on_user_identifier = optional(string, "")

      # For SSO-auth domains: the Identity Center username this profile belongs to.
      # Set together with `single_sign_on_user_identifier`. ForceNew.
      single_sign_on_user_value = optional(string, "")

      # Per-user overrides of the domain's `default_user_settings` — the same
      # settings tree, applied on top of the domain baseline. Unset means the
      # profile inherits the baseline unchanged. `execution_role_arn` is required
      # inside this block whenever it is set (the AWS contract for UserSettings).
      user_settings = optional(object({
        # execution_role_arn is the IAM role assumed by SageMaker when running notebooks,
        # training jobs, and other ML workloads on behalf of users. This role determines what
        # AWS resources (S3 buckets, ECR repos, Secrets Manager, etc.) users can access from
        # their Studio sessions. The role must have a trust policy allowing
        # sagemaker.amazonaws.com to assume it.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        execution_role_arn = string

        # security_group_ids are user-level security groups controlling network access for
        # notebook instances and apps. These are attached to the ENIs created for each user's
        # apps and control inbound/outbound traffic at the user level. Maximum 5 security groups.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        security_group_ids = optional(list(string), [])

        # default_landing_uri is the URI of the default app opened when a user accesses the domain.
        # Common values:
        #   "studio::relative/JupyterLab" - opens JupyterLab (recommended for most teams)
        #   "studio::relative/JupyterServer:" - opens classic Jupyter Server
        #   "studio::" - opens SageMaker Studio home
        # If omitted, AWS uses the platform default.
        default_landing_uri = optional(string, "")

        # studio_web_portal controls whether the SageMaker Studio web portal is accessible.
        # "ENABLED" (default): users can access the full Studio web interface.
        # "DISABLED": restricts access to programmatic-only usage (API/CLI).
        studio_web_portal = optional(string)

        # auto_mount_home_efs controls whether each user's home directory on the domain's EFS
        # file system is automatically mounted into their apps.
        # "Enabled": home directories mount automatically (the classic Studio experience).
        # "Disabled": apps start without the shared home directory -- pair with space storage
        #   or custom file systems when home directories are not wanted.
        # (AWS's third value, "DefaultAsDomain", is only valid on per-user profiles -- it means
        # "inherit this domain-level setting" and is rejected here at the domain level.)
        auto_mount_home_efs = optional(string)

        # jupyter_lab_app_settings configures JupyterLab, the primary IDE for SageMaker Studio.
        # JupyterLab provides a modern notebook and code editing experience with built-in Git
        # integration, terminal access, extension support, and collaborative editing.
        jupyter_lab_app_settings = optional(object({
          # default_resource_spec sets the default compute instance type and image configuration
          # for new JupyterLab apps. Users can override the instance type at app creation time.
          default_resource_spec = optional(object({
            # instance_type is the EC2 instance type for the app's compute.
            # Common choices:
            #   "ml.t3.medium" - development/exploration (2 vCPU, 4 GB)
            #   "ml.m5.large" - general-purpose notebooks (2 vCPU, 8 GB)
            #   "ml.g4dn.xlarge" - GPU workloads (1 GPU, 4 vCPU, 16 GB)
            #   "ml.p3.2xlarge" - heavy training (1 V100 GPU, 8 vCPU, 61 GB)
            #   "system" - lightweight system-managed instance for JupyterServer
            # The shape is checked here ("system" or an "ml."-prefixed type) so a typo
            # fails at manifest time; membership in AWS's instance-type list grows with
            # every hardware launch and is validated by AWS at deploy time.
            instance_type = optional(string, "")

            # lifecycle_config_arn is the ARN of a lifecycle configuration script that runs
            # when this specific app type starts.
            lifecycle_config_arn = optional(string, "")

            # sagemaker_image_arn is the ARN of a SageMaker Image that defines the app's container.
            # Use this to specify a custom base image instead of the SageMaker-provided default.
            sagemaker_image_arn = optional(string, "")

            # sagemaker_image_version_alias is a human-readable alias for a specific image version
            # (e.g., "latest", "v2.0"). Mutually exclusive with sagemaker_image_version_arn.
            sagemaker_image_version_alias = optional(string, "")

            # sagemaker_image_version_arn is the ARN of a specific SageMaker image version.
            # Pins the app to an exact image version for reproducibility across team members.
            # Mutually exclusive with sagemaker_image_version_alias.
            sagemaker_image_version_arn = optional(string, "")
          }))

          # lifecycle_config_arns are ARNs of lifecycle configuration scripts that run when
          # JupyterLab apps start. Use lifecycle configs to install Python packages, configure
          # JupyterLab extensions, set environment variables, or mount additional storage.
          lifecycle_config_arns = optional(list(string), [])

          # built_in_lifecycle_config_arn is the ARN of an AWS-curated (built-in) lifecycle
          # configuration to run at app start, as opposed to the customer-authored scripts in
          # lifecycle_config_arns.
          built_in_lifecycle_config_arn = optional(string, "")

          # custom_images are custom Docker images available as JupyterLab kernels.
          # Each image must be registered in SageMaker via an AppImageConfig that defines
          # the kernel specification and file system layout. Maximum 200 images.
          custom_images = optional(list(object({
            # app_image_config_name is the name of the SageMaker AppImageConfig that defines how
            # the image is presented to users (kernel specifications, file system configuration).
            # The AppImageConfig must exist before referencing it here.
            app_image_config_name = string

            # image_name is the name of the SageMaker Image resource that contains this container image.
            # The Image resource must exist before referencing it here.
            image_name = string

            # image_version_number pins to a specific version of the image.
            # If omitted, the latest available version is used.
            image_version_number = optional(number)
          })), [])

          # code_repositories are Git repositories automatically cloned into JupyterLab on startup.
          # Provides immediate access to team code, shared notebooks, and documentation.
          # Maximum 10 repositories.
          code_repositories = optional(list(object({
            # repository_url is the HTTPS URL of the Git repository to clone.
            # Must be an HTTPS URL (SSH URLs are not supported by SageMaker).
            # Examples: "https://github.com/org/ml-notebooks.git"
            # Maximum length: 1024 characters.
            repository_url = string
          })), [])

          # idle_settings configures automatic shutdown of idle JupyterLab instances.
          # Critical for cost management: without idle timeout, instances run 24/7 at full
          # compute cost even when no user is interacting with them.
          idle_settings = optional(object({
            # lifecycle_management is the enforcement switch: "ENABLED" (the default when
            # this block is present) turns automatic idle shutdown on; "DISABLED" keeps
            # the block's timeout values as published guardrails users may adopt WITHOUT
            # forcing auto-shutdown on — the defined-but-disabled state. Both engines
            # send the explicit value, so flipping to "DISABLED" genuinely turns
            # enforcement off.
            lifecycle_management = optional(string)

            # idle_timeout_in_minutes is the duration of inactivity (in minutes) before an instance
            # is automatically shut down. Range: 60-525600 (1 hour to 365 days).
            # A reasonable production default is 120 (2 hours).
            idle_timeout_in_minutes = number

            # min_idle_timeout_in_minutes sets the minimum idle timeout that individual users can
            # configure for their own apps. Prevents users from setting excessively short timeouts
            # that would cause disruptive shutdowns during brief pauses. Range: 60-525600.
            min_idle_timeout_in_minutes = number

            # max_idle_timeout_in_minutes sets the maximum idle timeout that individual users can
            # configure. Prevents users from effectively disabling idle shutdown by setting
            # extremely long timeouts. Range: 60-525600.
            max_idle_timeout_in_minutes = number
          }))

          # emr_settings pre-authorizes the IAM roles JupyterLab uses to discover and connect
          # to Amazon EMR clusters for large-scale data processing directly from notebooks.
          emr_settings = optional(object({
            # assumable_role_arns are IAM roles the JupyterLab app can assume to CONNECT to EMR
            # clusters -- including clusters in other AWS accounts (cross-account access).
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            assumable_role_arns = optional(list(string), [])

            # execution_role_arns are IAM runtime roles available for EMR workloads submitted from
            # the notebook (the role the EMR job itself runs under).
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            execution_role_arns = optional(list(string), [])
          }))
        }))

        # jupyter_server_app_settings configures the classic Jupyter Server app (Studio
        # Classic). Teams still running the previous-generation Studio experience configure
        # its default resources and startup repositories here.
        jupyter_server_app_settings = optional(object({
          # default_resource_spec sets the default configuration for the Jupyter Server app.
          # Jupyter Server runs on a lightweight system-managed instance ("system" instance type).
          default_resource_spec = optional(object({
            # instance_type is the EC2 instance type for the app's compute.
            # Common choices:
            #   "ml.t3.medium" - development/exploration (2 vCPU, 4 GB)
            #   "ml.m5.large" - general-purpose notebooks (2 vCPU, 8 GB)
            #   "ml.g4dn.xlarge" - GPU workloads (1 GPU, 4 vCPU, 16 GB)
            #   "ml.p3.2xlarge" - heavy training (1 V100 GPU, 8 vCPU, 61 GB)
            #   "system" - lightweight system-managed instance for JupyterServer
            # The shape is checked here ("system" or an "ml."-prefixed type) so a typo
            # fails at manifest time; membership in AWS's instance-type list grows with
            # every hardware launch and is validated by AWS at deploy time.
            instance_type = optional(string, "")

            # lifecycle_config_arn is the ARN of a lifecycle configuration script that runs
            # when this specific app type starts.
            lifecycle_config_arn = optional(string, "")

            # sagemaker_image_arn is the ARN of a SageMaker Image that defines the app's container.
            # Use this to specify a custom base image instead of the SageMaker-provided default.
            sagemaker_image_arn = optional(string, "")

            # sagemaker_image_version_alias is a human-readable alias for a specific image version
            # (e.g., "latest", "v2.0"). Mutually exclusive with sagemaker_image_version_arn.
            sagemaker_image_version_alias = optional(string, "")

            # sagemaker_image_version_arn is the ARN of a specific SageMaker image version.
            # Pins the app to an exact image version for reproducibility across team members.
            # Mutually exclusive with sagemaker_image_version_alias.
            sagemaker_image_version_arn = optional(string, "")
          }))

          # lifecycle_config_arns are ARNs of lifecycle configuration scripts for the
          # Jupyter Server app.
          lifecycle_config_arns = optional(list(string), [])

          # code_repositories are Git repositories automatically cloned on startup.
          # Maximum 10 repositories.
          code_repositories = optional(list(object({
            # repository_url is the HTTPS URL of the Git repository to clone.
            # Must be an HTTPS URL (SSH URLs are not supported by SageMaker).
            # Examples: "https://github.com/org/ml-notebooks.git"
            # Maximum length: 1024 characters.
            repository_url = string
          })), [])
        }))

        # kernel_gateway_app_settings configures KernelGateway apps that provide custom compute
        # kernels. Use KernelGateway to bring your own Docker images with specialized ML
        # frameworks, custom libraries, or GPU-optimized environments that go beyond the
        # standard SageMaker-provided kernels.
        kernel_gateway_app_settings = optional(object({
          # default_resource_spec sets the default compute instance type and image configuration
          # for new KernelGateway apps.
          default_resource_spec = optional(object({
            # instance_type is the EC2 instance type for the app's compute.
            # Common choices:
            #   "ml.t3.medium" - development/exploration (2 vCPU, 4 GB)
            #   "ml.m5.large" - general-purpose notebooks (2 vCPU, 8 GB)
            #   "ml.g4dn.xlarge" - GPU workloads (1 GPU, 4 vCPU, 16 GB)
            #   "ml.p3.2xlarge" - heavy training (1 V100 GPU, 8 vCPU, 61 GB)
            #   "system" - lightweight system-managed instance for JupyterServer
            # The shape is checked here ("system" or an "ml."-prefixed type) so a typo
            # fails at manifest time; membership in AWS's instance-type list grows with
            # every hardware launch and is validated by AWS at deploy time.
            instance_type = optional(string, "")

            # lifecycle_config_arn is the ARN of a lifecycle configuration script that runs
            # when this specific app type starts.
            lifecycle_config_arn = optional(string, "")

            # sagemaker_image_arn is the ARN of a SageMaker Image that defines the app's container.
            # Use this to specify a custom base image instead of the SageMaker-provided default.
            sagemaker_image_arn = optional(string, "")

            # sagemaker_image_version_alias is a human-readable alias for a specific image version
            # (e.g., "latest", "v2.0"). Mutually exclusive with sagemaker_image_version_arn.
            sagemaker_image_version_alias = optional(string, "")

            # sagemaker_image_version_arn is the ARN of a specific SageMaker image version.
            # Pins the app to an exact image version for reproducibility across team members.
            # Mutually exclusive with sagemaker_image_version_alias.
            sagemaker_image_version_arn = optional(string, "")
          }))

          # lifecycle_config_arns are ARNs of lifecycle configuration scripts for KernelGateway apps.
          lifecycle_config_arns = optional(list(string), [])

          # custom_images are custom Docker images available as KernelGateway kernels.
          # Each image must be registered in SageMaker via an AppImageConfig. Maximum 200 images.
          custom_images = optional(list(object({
            # app_image_config_name is the name of the SageMaker AppImageConfig that defines how
            # the image is presented to users (kernel specifications, file system configuration).
            # The AppImageConfig must exist before referencing it here.
            app_image_config_name = string

            # image_name is the name of the SageMaker Image resource that contains this container image.
            # The Image resource must exist before referencing it here.
            image_name = string

            # image_version_number pins to a specific version of the image.
            # If omitted, the latest available version is used.
            image_version_number = optional(number)
          })), [])
        }))

        # code_editor_app_settings configures the Code Editor app -- SageMaker's VS Code
        # (Code-OSS) IDE. It carries the same resource-spec/lifecycle/custom-image/idle
        # controls as JupyterLab, for teams that prefer a full code editor over notebooks.
        code_editor_app_settings = optional(object({
          # default_resource_spec sets the default compute instance type and image configuration
          # for new Code Editor apps.
          default_resource_spec = optional(object({
            # instance_type is the EC2 instance type for the app's compute.
            # Common choices:
            #   "ml.t3.medium" - development/exploration (2 vCPU, 4 GB)
            #   "ml.m5.large" - general-purpose notebooks (2 vCPU, 8 GB)
            #   "ml.g4dn.xlarge" - GPU workloads (1 GPU, 4 vCPU, 16 GB)
            #   "ml.p3.2xlarge" - heavy training (1 V100 GPU, 8 vCPU, 61 GB)
            #   "system" - lightweight system-managed instance for JupyterServer
            # The shape is checked here ("system" or an "ml."-prefixed type) so a typo
            # fails at manifest time; membership in AWS's instance-type list grows with
            # every hardware launch and is validated by AWS at deploy time.
            instance_type = optional(string, "")

            # lifecycle_config_arn is the ARN of a lifecycle configuration script that runs
            # when this specific app type starts.
            lifecycle_config_arn = optional(string, "")

            # sagemaker_image_arn is the ARN of a SageMaker Image that defines the app's container.
            # Use this to specify a custom base image instead of the SageMaker-provided default.
            sagemaker_image_arn = optional(string, "")

            # sagemaker_image_version_alias is a human-readable alias for a specific image version
            # (e.g., "latest", "v2.0"). Mutually exclusive with sagemaker_image_version_arn.
            sagemaker_image_version_alias = optional(string, "")

            # sagemaker_image_version_arn is the ARN of a specific SageMaker image version.
            # Pins the app to an exact image version for reproducibility across team members.
            # Mutually exclusive with sagemaker_image_version_alias.
            sagemaker_image_version_arn = optional(string, "")
          }))

          # lifecycle_config_arns are ARNs of lifecycle configuration scripts for Code Editor apps.
          lifecycle_config_arns = optional(list(string), [])

          # built_in_lifecycle_config_arn is the ARN of an AWS-curated (built-in) lifecycle
          # configuration to run at app start.
          built_in_lifecycle_config_arn = optional(string, "")

          # custom_images are custom Docker images available in Code Editor. Maximum 200 images.
          custom_images = optional(list(object({
            # app_image_config_name is the name of the SageMaker AppImageConfig that defines how
            # the image is presented to users (kernel specifications, file system configuration).
            # The AppImageConfig must exist before referencing it here.
            app_image_config_name = string

            # image_name is the name of the SageMaker Image resource that contains this container image.
            # The Image resource must exist before referencing it here.
            image_name = string

            # image_version_number pins to a specific version of the image.
            # If omitted, the latest available version is used.
            image_version_number = optional(number)
          })), [])

          # idle_settings configures automatic shutdown of idle Code Editor instances -- the
          # same cost-control dial JupyterLab carries.
          idle_settings = optional(object({
            # lifecycle_management is the enforcement switch: "ENABLED" (the default when
            # this block is present) turns automatic idle shutdown on; "DISABLED" keeps
            # the block's timeout values as published guardrails users may adopt WITHOUT
            # forcing auto-shutdown on — the defined-but-disabled state. Both engines
            # send the explicit value, so flipping to "DISABLED" genuinely turns
            # enforcement off.
            lifecycle_management = optional(string)

            # idle_timeout_in_minutes is the duration of inactivity (in minutes) before an instance
            # is automatically shut down. Range: 60-525600 (1 hour to 365 days).
            # A reasonable production default is 120 (2 hours).
            idle_timeout_in_minutes = number

            # min_idle_timeout_in_minutes sets the minimum idle timeout that individual users can
            # configure for their own apps. Prevents users from setting excessively short timeouts
            # that would cause disruptive shutdowns during brief pauses. Range: 60-525600.
            min_idle_timeout_in_minutes = number

            # max_idle_timeout_in_minutes sets the maximum idle timeout that individual users can
            # configure. Prevents users from effectively disabling idle shutdown by setting
            # extremely long timeouts. Range: 60-525600.
            max_idle_timeout_in_minutes = number
          }))
        }))

        # tensor_board_app_settings configures the TensorBoard app used to visualize
        # training runs. Only the default resource spec is configurable.
        tensor_board_app_settings = optional(object({
          # default_resource_spec sets the default compute instance type and image configuration
          # for the TensorBoard app.
          default_resource_spec = optional(object({
            # instance_type is the EC2 instance type for the app's compute.
            # Common choices:
            #   "ml.t3.medium" - development/exploration (2 vCPU, 4 GB)
            #   "ml.m5.large" - general-purpose notebooks (2 vCPU, 8 GB)
            #   "ml.g4dn.xlarge" - GPU workloads (1 GPU, 4 vCPU, 16 GB)
            #   "ml.p3.2xlarge" - heavy training (1 V100 GPU, 8 vCPU, 61 GB)
            #   "system" - lightweight system-managed instance for JupyterServer
            # The shape is checked here ("system" or an "ml."-prefixed type) so a typo
            # fails at manifest time; membership in AWS's instance-type list grows with
            # every hardware launch and is validated by AWS at deploy time.
            instance_type = optional(string, "")

            # lifecycle_config_arn is the ARN of a lifecycle configuration script that runs
            # when this specific app type starts.
            lifecycle_config_arn = optional(string, "")

            # sagemaker_image_arn is the ARN of a SageMaker Image that defines the app's container.
            # Use this to specify a custom base image instead of the SageMaker-provided default.
            sagemaker_image_arn = optional(string, "")

            # sagemaker_image_version_alias is a human-readable alias for a specific image version
            # (e.g., "latest", "v2.0"). Mutually exclusive with sagemaker_image_version_arn.
            sagemaker_image_version_alias = optional(string, "")

            # sagemaker_image_version_arn is the ARN of a specific SageMaker image version.
            # Pins the app to an exact image version for reproducibility across team members.
            # Mutually exclusive with sagemaker_image_version_alias.
            sagemaker_image_version_arn = optional(string, "")
          }))
        }))

        # r_session_app_settings configures RSession apps (the R kernels backing RStudio
        # sessions). Requires RStudio to be enabled on the domain via
        # r_studio_server_pro_domain_settings.
        r_session_app_settings = optional(object({
          # default_resource_spec sets the default compute instance type and image configuration
          # for RSession apps.
          default_resource_spec = optional(object({
            # instance_type is the EC2 instance type for the app's compute.
            # Common choices:
            #   "ml.t3.medium" - development/exploration (2 vCPU, 4 GB)
            #   "ml.m5.large" - general-purpose notebooks (2 vCPU, 8 GB)
            #   "ml.g4dn.xlarge" - GPU workloads (1 GPU, 4 vCPU, 16 GB)
            #   "ml.p3.2xlarge" - heavy training (1 V100 GPU, 8 vCPU, 61 GB)
            #   "system" - lightweight system-managed instance for JupyterServer
            # The shape is checked here ("system" or an "ml."-prefixed type) so a typo
            # fails at manifest time; membership in AWS's instance-type list grows with
            # every hardware launch and is validated by AWS at deploy time.
            instance_type = optional(string, "")

            # lifecycle_config_arn is the ARN of a lifecycle configuration script that runs
            # when this specific app type starts.
            lifecycle_config_arn = optional(string, "")

            # sagemaker_image_arn is the ARN of a SageMaker Image that defines the app's container.
            # Use this to specify a custom base image instead of the SageMaker-provided default.
            sagemaker_image_arn = optional(string, "")

            # sagemaker_image_version_alias is a human-readable alias for a specific image version
            # (e.g., "latest", "v2.0"). Mutually exclusive with sagemaker_image_version_arn.
            sagemaker_image_version_alias = optional(string, "")

            # sagemaker_image_version_arn is the ARN of a specific SageMaker image version.
            # Pins the app to an exact image version for reproducibility across team members.
            # Mutually exclusive with sagemaker_image_version_alias.
            sagemaker_image_version_arn = optional(string, "")
          }))

          # custom_images are custom Docker images available as RSession kernels. Maximum 200 images.
          custom_images = optional(list(object({
            # app_image_config_name is the name of the SageMaker AppImageConfig that defines how
            # the image is presented to users (kernel specifications, file system configuration).
            # The AppImageConfig must exist before referencing it here.
            app_image_config_name = string

            # image_name is the name of the SageMaker Image resource that contains this container image.
            # The Image resource must exist before referencing it here.
            image_name = string

            # image_version_number pins to a specific version of the image.
            # If omitted, the latest available version is used.
            image_version_number = optional(number)
          })), [])
        }))

        # r_studio_server_pro_app_settings controls per-user RStudio Server Pro access.
        # Requires RStudio to be enabled on the domain via r_studio_server_pro_domain_settings.
        r_studio_server_pro_app_settings = optional(object({
          # access_status grants or denies the user access to RStudio Server Pro.
          # "ENABLED": the user sees and can launch RStudio.
          # "DISABLED": RStudio is hidden for the user.
          access_status = optional(string, "")

          # user_group assigns the RStudio authorization level.
          # "R_STUDIO_ADMIN": administrative access to the RStudio Workbench admin dashboard.
          # "R_STUDIO_USER" (AWS default): regular RStudio user.
          # Only meaningful when access_status is "ENABLED" -- AWS ignores it otherwise, so the
          # spec rejects the dead combination up front.
          user_group = optional(string, "")
        }))

        # canvas_app_settings configures SageMaker Canvas, the no-code ML workspace.
        # Each sub-block is an independent Canvas capability (direct model deployment,
        # EMR Serverless big-data processing, Bedrock generative AI, SaaS data connectors,
        # Kendra document search, cross-account model registration, time-series forecasting,
        # and the shared artifact workspace).
        canvas_app_settings = optional(object({
          # direct_deploy_status controls whether Canvas users can deploy models they build
          # directly to SageMaker real-time endpoints ("ENABLED" or "DISABLED"). Direct
          # deployment creates billable endpoints, so governance-minded teams often disable it
          # and route models through the registry instead (model_register_settings).
          direct_deploy_status = optional(string)

          # emr_serverless_settings lets Canvas run large data preparation and processing jobs
          # on EMR Serverless.
          emr_serverless_settings = optional(object({
            # execution_role_arn is the IAM role Canvas uses to submit and manage EMR Serverless
            # jobs.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            execution_role_arn = optional(string, "")

            # status enables or disables the capability ("ENABLED" or "DISABLED").
            status = optional(string, "")
          }))

          # generative_ai_bedrock_role_arn is the IAM role Canvas assumes to call Amazon Bedrock
          # for generative-AI features. Setting the role is what enables the capability.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          generative_ai_bedrock_role_arn = optional(string, "")

          # identity_provider_oauth_settings wire Canvas to external SaaS data sources
          # (Salesforce Data Cloud, Snowflake) through OAuth. Each entry names the data source
          # and the Secrets Manager secret holding its OAuth client credentials. Maximum 20.
          identity_provider_oauth_settings = optional(list(object({
            # data_source_name identifies the SaaS data source.
            # "SalesforceGenie": Salesforce Data Cloud.
            # "Snowflake": Snowflake.
            data_source_name = optional(string, "")

            # secret_arn is the Secrets Manager secret holding the data source's OAuth client
            # credentials (client ID and secret). The ARN is a reference Canvas resolves at
            # connection time -- never secret material itself.
            secret_arn = string

            # status enables or disables this connector ("ENABLED" or "DISABLED").
            status = optional(string, "")
          })), [])

          # kendra_settings_status controls whether Canvas can query Amazon Kendra indexes for
          # document search ("ENABLED" or "DISABLED").
          kendra_settings_status = optional(string)

          # model_register_settings controls whether Canvas users can register their models into
          # a SageMaker Model Registry, optionally in another AWS account.
          model_register_settings = optional(object({
            # cross_account_model_register_role_arn is the IAM role Canvas assumes to register
            # models into a Model Registry that lives in ANOTHER AWS account. Leave unset to
            # register into this account's registry.
            cross_account_model_register_role_arn = optional(string, "")

            # status enables or disables model registration ("ENABLED" or "DISABLED").
            status = optional(string, "")
          }))

          # time_series_forecasting_settings enables Canvas time-series forecasting, which uses
          # Amazon Forecast under the hood via the given IAM role.
          time_series_forecasting_settings = optional(object({
            # amazon_forecast_role_arn is the IAM role Canvas assumes to call Amazon Forecast.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            amazon_forecast_role_arn = optional(string, "")

            # status enables or disables forecasting ("ENABLED" or "DISABLED").
            status = optional(string, "")
          }))

          # workspace_settings pins the S3 location (and optional KMS key) where Canvas stores
          # its working artifacts -- datasets, intermediate results, generated models.
          workspace_settings = optional(object({
            # s3_artifact_path is the S3 URI where Canvas stores datasets, intermediate results,
            # and generated models. Example: "s3://my-canvas-workspace/artifacts/".
            s3_artifact_path = optional(string, "")

            # s3_kms_key_id is the KMS key used to encrypt Canvas artifacts in S3.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            s3_kms_key_id = optional(string, "")
          }))
        }))

        # sharing_settings controls notebook output sharing to S3. When enabled, notebook cell
        # outputs are persisted to an S3 location, allowing team members to view results
        # without running the notebook.
        sharing_settings = optional(object({
          # notebook_output_option controls whether notebook cell outputs are persisted to S3.
          # "Allowed": outputs are copied to S3 at the location specified by s3_output_path.
          # "Disabled" (default): outputs are not shared externally.
          notebook_output_option = optional(string)

          # s3_kms_key_id is the KMS key used to encrypt shared notebook outputs in S3.
          # If omitted, outputs are encrypted with the default S3 bucket encryption.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          s3_kms_key_id = optional(string, "")

          # s3_output_path is the S3 URI where shared notebook outputs are stored.
          # Required when notebook_output_option is "Allowed".
          # Example: "s3://my-team-bucket/notebook-outputs/"
          s3_output_path = optional(string, "")
        }))

        # space_storage_settings configures default EBS volume sizes for user spaces.
        # Spaces use EBS volumes for working storage beyond the shared EFS home directory.
        space_storage_settings = optional(object({
          # default_ebs_volume_size_in_gb is the default EBS volume size (in GB) assigned to new spaces.
          default_ebs_volume_size_in_gb = number

          # maximum_ebs_volume_size_in_gb is the maximum EBS volume size (in GB) that users can request
          # for their spaces. Must be >= default_ebs_volume_size_in_gb.
          maximum_ebs_volume_size_in_gb = number
        }))

        # custom_file_system_configs mount additional file systems (beyond the domain's own
        # EFS home directories) into every user's apps -- shared datasets, feature stores,
        # model artifact trees. Each entry names one file system and the path where it mounts.
        custom_file_system_configs = optional(list(object({
          # efs_file_system_config mounts an Amazon EFS file system.
          efs_file_system_config = object({
            # file_system_id is the EFS file system to mount. The file system must be reachable
            # from the domain's subnets (mount targets + security groups are the file system's
            # own configuration).
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            file_system_id = string

            # file_system_path is the path within the EFS file system to mount into apps.
            # Example: "/shared/datasets"
            file_system_path = string
          })
        })), [])

        # custom_posix_user_config sets the POSIX identity (UID/GID) that apps run as when
        # accessing the EFS home directory and custom file systems. Set this when file-system
        # permissions on shared storage must map to a specific owner instead of SageMaker's
        # default identity.
        custom_posix_user_config = optional(object({
          # uid is the POSIX user ID. Must be at least 10000.
          uid = number

          # gid is the POSIX group ID. Must be at least 1001.
          gid = number
        }))

        # studio_web_portal_settings hides parts of the Studio UI from users -- entire app
        # types, specific instance types, or ML tools. Use it to keep expensive GPU instance
        # types or unused tooling out of the picker instead of policing them after the fact.
        studio_web_portal_settings = optional(object({
          # hidden_app_types are Studio app types to hide (e.g. "JupyterServer", "KernelGateway",
          # "Canvas", "CodeEditor", "JupyterLab", "TensorBoard", "RStudioServerPro"). Values are
          # SageMaker AppType names; AWS validates them at deploy time as the set grows.
          hidden_app_types = optional(list(string), [])

          # hidden_instance_types are instance types to hide from app-creation pickers
          # (e.g. "ml.p3.2xlarge"). Values are SageMaker app instance type names.
          hidden_instance_types = optional(list(string), [])

          # hidden_ml_tools are Studio ML tools to hide (e.g. "DataWrangler", "FeatureStore",
          # "EmrClusters", "AutoMl", "Experiments", "Pipelines"). Values are SageMaker MlTools
          # names; AWS validates them at deploy time as the set grows.
          hidden_ml_tools = optional(list(string), [])
        }))
      }))
    })), [])

    # spaces are named shared (or private) workspaces inside the domain, keyed
    # by name — the collaboration plane, where a JupyterLab or Code Editor
    # runtime is shared rather than per-user. Spaces inherit
    # `default_space_settings` and may override via their own
    # `space_settings`. Add/remove in place; removing an entry deletes the
    # space and its EBS volume.
    spaces = optional(list(object({
      # The space name — unique within the domain, and the key both IaC modules use
      # for the satellite resource. 1-63 characters: alphanumeric and hyphens,
      # starting alphanumeric. ForceNew.
      space_name = string

      # Display name shown in Studio (the space name itself is immutable; this is
      # not). Maximum 64 characters. Updates in place.
      display_name = optional(string, "")

      # Ownership: names the user profile that owns this space. Required for
      # PRIVATE spaces and for shared spaces alike when sharing is configured —
      # AWS requires ownership and sharing to be declared together (or neither,
      # which creates a legacy unscoped space). The owner profile must exist in
      # the domain; it is usually one of `user_profiles`, but profiles provisioned
      # outside this manifest (SSO auto-provisioning) are equally valid owners.
      # Not updatable after create — the provider never sends changes to it.
      ownership_settings = optional(object({
        # The owning user profile's name (not ARN).
        owner_user_profile_name = string
      }))

      # Sharing posture: "Private" (one owner, one runtime) or "Shared"
      # (collaborative). Declared together with `ownership_settings`. Not
      # updatable after create.
      space_sharing_settings = optional(object({
        # "Private" or "Shared".
        sharing_type = string
      }))

      # The space's own settings — app type, per-app configuration, storage, and
      # mounted file systems.
      space_settings = optional(object({
        # Which app this space runs: "JupyterLab", "CodeEditor", "JupyterServer",
        # or "KernelGateway". Spaces are single-app workspaces; the matching
        # app-settings block below configures it.
        app_type = optional(string)

        # JupyterLab configuration for the space. `default_resource_spec` is
        # required here (unlike the domain baseline) — AWS's space contract.
        jupyter_lab_app_settings = optional(object({
          # The compute instance type and image for the space's JupyterLab runtime.
          default_resource_spec = object({
            # instance_type is the EC2 instance type for the app's compute.
            # Common choices:
            #   "ml.t3.medium" - development/exploration (2 vCPU, 4 GB)
            #   "ml.m5.large" - general-purpose notebooks (2 vCPU, 8 GB)
            #   "ml.g4dn.xlarge" - GPU workloads (1 GPU, 4 vCPU, 16 GB)
            #   "ml.p3.2xlarge" - heavy training (1 V100 GPU, 8 vCPU, 61 GB)
            #   "system" - lightweight system-managed instance for JupyterServer
            # The shape is checked here ("system" or an "ml."-prefixed type) so a typo
            # fails at manifest time; membership in AWS's instance-type list grows with
            # every hardware launch and is validated by AWS at deploy time.
            instance_type = optional(string, "")

            # lifecycle_config_arn is the ARN of a lifecycle configuration script that runs
            # when this specific app type starts.
            lifecycle_config_arn = optional(string, "")

            # sagemaker_image_arn is the ARN of a SageMaker Image that defines the app's container.
            # Use this to specify a custom base image instead of the SageMaker-provided default.
            sagemaker_image_arn = optional(string, "")

            # sagemaker_image_version_alias is a human-readable alias for a specific image version
            # (e.g., "latest", "v2.0"). Mutually exclusive with sagemaker_image_version_arn.
            sagemaker_image_version_alias = optional(string, "")

            # sagemaker_image_version_arn is the ARN of a specific SageMaker image version.
            # Pins the app to an exact image version for reproducibility across team members.
            # Mutually exclusive with sagemaker_image_version_alias.
            sagemaker_image_version_arn = optional(string, "")
          })

          # Git repositories cloned into the space's JupyterLab on startup. Maximum 10.
          code_repositories = optional(list(object({
            # repository_url is the HTTPS URL of the Git repository to clone.
            # Must be an HTTPS URL (SSH URLs are not supported by SageMaker).
            # Examples: "https://github.com/org/ml-notebooks.git"
            # Maximum length: 1024 characters.
            repository_url = string
          })), [])

          # Automatic shutdown of the space's JupyterLab when idle.
          idle_settings = optional(object({
            # Minutes of inactivity before the space's app shuts down. Range: 60-525600.
            idle_timeout_in_minutes = optional(number)
          }))
        }))

        # Code Editor (VS Code / Code-OSS) configuration for the space.
        code_editor_app_settings = optional(object({
          # The compute instance type and image for the space's Code Editor runtime.
          default_resource_spec = object({
            # instance_type is the EC2 instance type for the app's compute.
            # Common choices:
            #   "ml.t3.medium" - development/exploration (2 vCPU, 4 GB)
            #   "ml.m5.large" - general-purpose notebooks (2 vCPU, 8 GB)
            #   "ml.g4dn.xlarge" - GPU workloads (1 GPU, 4 vCPU, 16 GB)
            #   "ml.p3.2xlarge" - heavy training (1 V100 GPU, 8 vCPU, 61 GB)
            #   "system" - lightweight system-managed instance for JupyterServer
            # The shape is checked here ("system" or an "ml."-prefixed type) so a typo
            # fails at manifest time; membership in AWS's instance-type list grows with
            # every hardware launch and is validated by AWS at deploy time.
            instance_type = optional(string, "")

            # lifecycle_config_arn is the ARN of a lifecycle configuration script that runs
            # when this specific app type starts.
            lifecycle_config_arn = optional(string, "")

            # sagemaker_image_arn is the ARN of a SageMaker Image that defines the app's container.
            # Use this to specify a custom base image instead of the SageMaker-provided default.
            sagemaker_image_arn = optional(string, "")

            # sagemaker_image_version_alias is a human-readable alias for a specific image version
            # (e.g., "latest", "v2.0"). Mutually exclusive with sagemaker_image_version_arn.
            sagemaker_image_version_alias = optional(string, "")

            # sagemaker_image_version_arn is the ARN of a specific SageMaker image version.
            # Pins the app to an exact image version for reproducibility across team members.
            # Mutually exclusive with sagemaker_image_version_alias.
            sagemaker_image_version_arn = optional(string, "")
          })

          # Automatic shutdown of the space's Code Editor when idle.
          idle_settings = optional(object({
            # Minutes of inactivity before the space's app shuts down. Range: 60-525600.
            idle_timeout_in_minutes = optional(number)
          }))
        }))

        # Classic Jupyter Server configuration for the space. Same shape as the
        # domain baseline, but `default_resource_spec` is required here.
        jupyter_server_app_settings = optional(object({
          # default_resource_spec sets the default configuration for the Jupyter Server app.
          # Jupyter Server runs on a lightweight system-managed instance ("system" instance type).
          default_resource_spec = optional(object({
            # instance_type is the EC2 instance type for the app's compute.
            # Common choices:
            #   "ml.t3.medium" - development/exploration (2 vCPU, 4 GB)
            #   "ml.m5.large" - general-purpose notebooks (2 vCPU, 8 GB)
            #   "ml.g4dn.xlarge" - GPU workloads (1 GPU, 4 vCPU, 16 GB)
            #   "ml.p3.2xlarge" - heavy training (1 V100 GPU, 8 vCPU, 61 GB)
            #   "system" - lightweight system-managed instance for JupyterServer
            # The shape is checked here ("system" or an "ml."-prefixed type) so a typo
            # fails at manifest time; membership in AWS's instance-type list grows with
            # every hardware launch and is validated by AWS at deploy time.
            instance_type = optional(string, "")

            # lifecycle_config_arn is the ARN of a lifecycle configuration script that runs
            # when this specific app type starts.
            lifecycle_config_arn = optional(string, "")

            # sagemaker_image_arn is the ARN of a SageMaker Image that defines the app's container.
            # Use this to specify a custom base image instead of the SageMaker-provided default.
            sagemaker_image_arn = optional(string, "")

            # sagemaker_image_version_alias is a human-readable alias for a specific image version
            # (e.g., "latest", "v2.0"). Mutually exclusive with sagemaker_image_version_arn.
            sagemaker_image_version_alias = optional(string, "")

            # sagemaker_image_version_arn is the ARN of a specific SageMaker image version.
            # Pins the app to an exact image version for reproducibility across team members.
            # Mutually exclusive with sagemaker_image_version_alias.
            sagemaker_image_version_arn = optional(string, "")
          }))

          # lifecycle_config_arns are ARNs of lifecycle configuration scripts for the
          # Jupyter Server app.
          lifecycle_config_arns = optional(list(string), [])

          # code_repositories are Git repositories automatically cloned on startup.
          # Maximum 10 repositories.
          code_repositories = optional(list(object({
            # repository_url is the HTTPS URL of the Git repository to clone.
            # Must be an HTTPS URL (SSH URLs are not supported by SageMaker).
            # Examples: "https://github.com/org/ml-notebooks.git"
            # Maximum length: 1024 characters.
            repository_url = string
          })), [])
        }))

        # KernelGateway configuration for the space. Same shape as the domain
        # baseline, but `default_resource_spec` is required here.
        kernel_gateway_app_settings = optional(object({
          # default_resource_spec sets the default compute instance type and image configuration
          # for new KernelGateway apps.
          default_resource_spec = optional(object({
            # instance_type is the EC2 instance type for the app's compute.
            # Common choices:
            #   "ml.t3.medium" - development/exploration (2 vCPU, 4 GB)
            #   "ml.m5.large" - general-purpose notebooks (2 vCPU, 8 GB)
            #   "ml.g4dn.xlarge" - GPU workloads (1 GPU, 4 vCPU, 16 GB)
            #   "ml.p3.2xlarge" - heavy training (1 V100 GPU, 8 vCPU, 61 GB)
            #   "system" - lightweight system-managed instance for JupyterServer
            # The shape is checked here ("system" or an "ml."-prefixed type) so a typo
            # fails at manifest time; membership in AWS's instance-type list grows with
            # every hardware launch and is validated by AWS at deploy time.
            instance_type = optional(string, "")

            # lifecycle_config_arn is the ARN of a lifecycle configuration script that runs
            # when this specific app type starts.
            lifecycle_config_arn = optional(string, "")

            # sagemaker_image_arn is the ARN of a SageMaker Image that defines the app's container.
            # Use this to specify a custom base image instead of the SageMaker-provided default.
            sagemaker_image_arn = optional(string, "")

            # sagemaker_image_version_alias is a human-readable alias for a specific image version
            # (e.g., "latest", "v2.0"). Mutually exclusive with sagemaker_image_version_arn.
            sagemaker_image_version_alias = optional(string, "")

            # sagemaker_image_version_arn is the ARN of a specific SageMaker image version.
            # Pins the app to an exact image version for reproducibility across team members.
            # Mutually exclusive with sagemaker_image_version_alias.
            sagemaker_image_version_arn = optional(string, "")
          }))

          # lifecycle_config_arns are ARNs of lifecycle configuration scripts for KernelGateway apps.
          lifecycle_config_arns = optional(list(string), [])

          # custom_images are custom Docker images available as KernelGateway kernels.
          # Each image must be registered in SageMaker via an AppImageConfig. Maximum 200 images.
          custom_images = optional(list(object({
            # app_image_config_name is the name of the SageMaker AppImageConfig that defines how
            # the image is presented to users (kernel specifications, file system configuration).
            # The AppImageConfig must exist before referencing it here.
            app_image_config_name = string

            # image_name is the name of the SageMaker Image resource that contains this container image.
            # The Image resource must exist before referencing it here.
            image_name = string

            # image_version_number pins to a specific version of the image.
            # If omitted, the latest available version is used.
            image_version_number = optional(number)
          })), [])
        }))

        # Existing EFS file systems mounted into the space's apps (by id — the
        # space form has no per-mount path, unlike the domain baseline's config).
        custom_file_systems = optional(list(object({
          # The EFS file system to mount. The file system must have mount targets in
          # the domain's VPC.
          #
          # Containment-exempt: a space MOUNTS the file system; the domain is not
          # deployed inside it. On a diagram the domain stands in its VPC with a
          # line to the file system -- the verdict the domain-level
          # AwsSagemakerDomainEfsFileSystemConfig.file_system_id already carries.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          file_system_id = string
        })), [])

        # The space's EBS boot/work volume size, in GB (5-16384). Persists across
        # app restarts; deleted with the space.
        space_storage_settings = optional(object({
          # The EBS volume size in GB. Range: 5-16384.
          ebs_volume_size_in_gb = number
        }))
      }))
    })), [])
  })
}

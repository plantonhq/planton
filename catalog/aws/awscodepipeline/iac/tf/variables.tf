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
  description = "AwsCodePipeline specification"
  type = object({
    # The AWS region where the resource will be created.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # pipeline_type selects the pipeline version.
    #   V1: Legacy pipeline (no triggers, no variables, SUPERSEDED execution only)
    #   V2: Modern pipeline with triggers, variables, and advanced execution modes
    # Default: V2 (recommended for all new pipelines). The platform materializes
    # this default when the manifest is loaded; the AWS provider's own default
    # is V1, so a raw module invocation that bypasses manifest loading and
    # omits this field deploys a V1 pipeline.
    pipeline_type = optional(string)

    # execution_mode controls how concurrent pipeline executions are handled.
    #   SUPERSEDED: New execution supersedes any in-progress execution (default)
    #   QUEUED:     New executions queue behind the current one (V2 only)
    #   PARALLEL:   Executions run simultaneously without waiting (V2 only)
    # Default: SUPERSEDED.
    execution_mode = optional(string)

    # role_arn is the IAM role ARN that grants CodePipeline permission to
    # access source providers, invoke build/deploy actions, and manage
    # artifacts in S3. This role must have policies for every action provider
    # used in the pipeline (e.g., CodeBuild, S3, ECS, Lambda).
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    role_arn = string

    # artifact_stores define S3 buckets where pipeline artifacts are stored.
    # For single-region pipelines, provide exactly one store without a region.
    # For cross-region pipelines, provide one store per region (each with a
    # region field) so that actions in different regions have local artifact
    # access.
    artifact_stores = list(object({
      # location is the S3 bucket name for artifact storage.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      location = string

      # region is the AWS region for this artifact store. Required only for
      # cross-region pipelines. For single-region pipelines, omit this field.
      region = optional(string, "")

      # encryption_key_id is the KMS key ARN or ID used to encrypt artifacts.
      # If omitted, the default AWS-managed S3 encryption key is used.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      encryption_key_id = optional(string, "")
    }))

    # stages define the ordered sequence of pipeline stages. Each stage
    # contains one or more actions that run in parallel (same run_order)
    # or sequentially (different run_order values).
    # A pipeline requires at minimum two stages: a source stage and at
    # least one build, test, deploy, or approval stage.
    stages = list(object({
      # name is the stage name. Must be unique within the pipeline.
      # Pattern: alphanumeric, dots, at-signs, hyphens, underscores (1-100 chars).
      name = string

      # actions define the operations performed in this stage. Actions with
      # the same run_order execute in parallel; different run_order values
      # execute sequentially within the stage.
      actions = list(object({
        # name is the action name. Must be unique within the stage.
        name = string

        # category classifies the action type.
        #   Source:   Fetches source code or artifacts from a provider
        #   Build:    Compiles, tests, or transforms code
        #   Test:     Runs test suites
        #   Deploy:   Deploys artifacts to a target environment
        #   Approval: Requires manual approval before proceeding
        #   Invoke:   Invokes a Lambda function or other compute
        #   Compute:  Runs commands in a managed compute environment (V2)
        category = string

        # owner identifies who created the action type.
        #   AWS:        Built-in AWS actions (CodeBuild, S3, ECS, Lambda, etc.)
        #   ThirdParty: Third-party integrations (GitHub v1, etc.)
        #   Custom:     User-defined custom action types
        owner = string

        # provider is the service that performs the action. The valid values
        # depend on the category and owner combination. Common providers:
        #   Source:   CodeStarSourceConnection, S3, ECR, CodeCommit
        #   Build:    CodeBuild
        #   Test:     CodeBuild, DeviceFarm
        #   Deploy:   S3, CodeDeploy, CloudFormation, ECS, ElasticBeanstalk, Lambda
        #   Approval: Manual
        #   Invoke:   Lambda
        #   Compute:  EC2 (V2)
        provider = string

        # version is the action type version. Typically "1" for all built-in actions.
        version = string

        # configuration contains provider-specific key-value pairs that control
        # the action's behavior. Each provider expects different keys.
        #
        # Common examples:
        #   CodeStarSourceConnection: ConnectionArn, FullRepositoryId, BranchName,
        #                             OutputArtifactFormat, DetectChanges
        #   CodeBuild:                ProjectName, PrimarySource, EnvironmentVariables
        #   S3 (source):              S3Bucket, S3ObjectKey, PollForSourceChanges
        #   S3 (deploy):              BucketName, Extract, ObjectKey, CannedACL
        #   ECS:                      ClusterName, ServiceName, FileName
        #   Lambda:                   FunctionName, UserParameters
        #   Manual (approval):        CustomData, ExternalEntityLink, NotificationArn
        #   CloudFormation:           ActionMode, StackName, TemplatePath, RoleArn
        # Keys are limited to 50 characters (AWS's action-configuration key cap).
        configuration = optional(map(string), {})

        # input_artifacts are artifact names from previous stages or actions
        # that this action consumes. For Source actions, this is typically empty.
        input_artifacts = optional(list(string), [])

        # output_artifacts are artifact names that this action produces for
        # consumption by downstream stages or actions.
        output_artifacts = optional(list(string), [])

        # namespace defines a variable namespace for this action's output variables.
        # Other actions can reference these variables as #{namespace.VariableName}.
        # Only meaningful for actions that produce output variables (e.g., source actions).
        namespace = optional(string, "")

        # region is the AWS region where this action executes. Required for
        # cross-region actions. If omitted, the action runs in the pipeline's region.
        region = optional(string, "")

        # role_arn is an IAM role ARN that the action assumes instead of the
        # pipeline's role. Useful for cross-account deployments or when an action
        # needs different permissions than the pipeline role.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        role_arn = optional(string, "")

        # run_order controls execution order within a stage. Actions with the
        # same run_order execute in parallel. Lower values run first.
        # Range: 1-999. Default: 1 (all actions parallel).
        run_order = optional(number, 0)

        # timeout_in_minutes overrides the action type's default timeout.
        # Range: 5-86400 (60 days). AWS supports this override ONLY on Manual
        # Approval actions (category Approval, provider Manual) — every other
        # action type is rejected at pipeline creation. (The provider validates
        # only the numeric range; the Approval-only scope is AWS's own API
        # contract, mirrored here as validation.)
        timeout_in_minutes = optional(number, 0)

        # commands are shell commands run by a Compute action (category Compute,
        # provider Commands) — inline CI steps without standing up a CodeBuild
        # project. CodeBuild logs and permissions are used under the hood, and
        # running commands bills CodeBuild compute time per execution. All
        # command forms are supported except multi-line formats. Max 50 commands,
        # each 1-1000 characters. AWS ignores this field on non-Compute actions.
        commands = optional(list(string), [])

        # output_artifacts_for_compute_action declares the file artifacts a
        # Compute action exports (Compute actions use this INSTEAD of
        # output_artifacts — AWS ignores plain output artifact names on Compute
        # actions and file-based artifacts on every other category).
        output_artifacts_for_compute_action = optional(list(object({
          # name is the output artifact name, unique within the pipeline.
          name = string

          # files are the paths (relative to the compute working directory) to
          # include in the exported artifact. Max 10 paths, each 1-128 characters.
          files = optional(list(string), [])
        })), [])

        # output_variables lists variable names a Compute action exports (these
        # are the CodeBuild environment variables the commands set). Downstream
        # actions reference them through this action's namespace as
        # #{namespace.VariableName}. Max 15 names, each 1-128 characters.
        output_variables = optional(list(string), [])
      }))

      # before_entry is an entry gate: its rules must pass before the stage
      # starts (e.g., a DeploymentWindow rule that only admits executions
      # during business hours, or a CloudWatchAlarm rule that blocks deploys
      # while an alarm is firing). V2 pipelines only.
      before_entry = optional(object({
        # result is the outcome applied when the condition's rules do NOT pass.
        #   FAIL: Fail the stage (default behavior when omitted)
        #   SKIP: Skip the stage and continue the pipeline (before_entry gates)
        result = optional(string, "")

        # rules are the checks that make up this condition (1-5, AND'd together).
        rules = list(object({
          # name is the rule name, unique within the condition.
          name = string

          # rule_type_id identifies which managed rule runs.
          rule_type_id = object({
            # provider is the managed rule provider: DeploymentWindow,
            # CloudWatchAlarm, LambdaInvoke, VariableCheck, or Commands.
            provider = string

            # category is the rule category. AWS currently supports only "Rule".
            category = optional(string)

            # owner identifies who provides the rule. AWS currently supports only
            # "AWS" (managed rules).
            owner = optional(string)

            # version is the rule type version. Typically "1"; AWS resolves the
            # current version when omitted.
            version = optional(string, "")
          })

          # configuration contains rule-provider-specific key-value pairs (e.g.,
          # DeploymentWindow: Cron, TimeZone; CloudWatchAlarm: AlarmName,
          # WaitTime; LambdaInvoke: FunctionName). Values are limited to 10,000
          # characters (AWS's rule-configuration value cap).
          configuration = optional(map(string), {})

          # commands are shell commands executed by the Commands rule provider
          # (max 50, each 1-1000 characters).
          commands = optional(list(string), [])

          # input_artifacts are artifact names available to the rule (e.g., for
          # Commands rules operating on build output). Each name 1-100 characters,
          # alphanumeric with underscores and hyphens.
          input_artifacts = optional(list(string), [])

          # region is the AWS region where the rule executes. Omit to run in the
          # pipeline's region.
          region = optional(string, "")

          # role_arn is an IAM role the rule assumes instead of the pipeline's
          # role (e.g., a scoped role for the LambdaInvoke rule).
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          role_arn = optional(string, "")

          # timeout_in_minutes is the maximum time the rule can run (5-86400).
          timeout_in_minutes = optional(number, 0)
        }))
      }))

      # on_success runs verification rules after the stage's actions succeed;
      # a failing rule fails the stage despite the successful actions (e.g., a
      # post-deploy CloudWatchAlarm check). V2 pipelines only.
      on_success = optional(object({
        # result is the outcome applied when the condition's rules do NOT pass.
        #   FAIL: Fail the stage (default behavior when omitted)
        #   SKIP: Skip the stage and continue the pipeline (before_entry gates)
        result = optional(string, "")

        # rules are the checks that make up this condition (1-5, AND'd together).
        rules = list(object({
          # name is the rule name, unique within the condition.
          name = string

          # rule_type_id identifies which managed rule runs.
          rule_type_id = object({
            # provider is the managed rule provider: DeploymentWindow,
            # CloudWatchAlarm, LambdaInvoke, VariableCheck, or Commands.
            provider = string

            # category is the rule category. AWS currently supports only "Rule".
            category = optional(string)

            # owner identifies who provides the rule. AWS currently supports only
            # "AWS" (managed rules).
            owner = optional(string)

            # version is the rule type version. Typically "1"; AWS resolves the
            # current version when omitted.
            version = optional(string, "")
          })

          # configuration contains rule-provider-specific key-value pairs (e.g.,
          # DeploymentWindow: Cron, TimeZone; CloudWatchAlarm: AlarmName,
          # WaitTime; LambdaInvoke: FunctionName). Values are limited to 10,000
          # characters (AWS's rule-configuration value cap).
          configuration = optional(map(string), {})

          # commands are shell commands executed by the Commands rule provider
          # (max 50, each 1-1000 characters).
          commands = optional(list(string), [])

          # input_artifacts are artifact names available to the rule (e.g., for
          # Commands rules operating on build output). Each name 1-100 characters,
          # alphanumeric with underscores and hyphens.
          input_artifacts = optional(list(string), [])

          # region is the AWS region where the rule executes. Omit to run in the
          # pipeline's region.
          region = optional(string, "")

          # role_arn is an IAM role the rule assumes instead of the pipeline's
          # role (e.g., a scoped role for the LambdaInvoke rule).
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          role_arn = optional(string, "")

          # timeout_in_minutes is the maximum time the rule can run (5-86400).
          timeout_in_minutes = optional(number, 0)
        }))
      }))

      # on_failure controls what happens when the stage fails: automatically
      # roll back to the last successful state, retry the stage, or
      # conditionally decide via rules. V2 pipelines only.
      on_failure = optional(object({
        # result is the action taken when the stage fails.
        #   ROLLBACK: Automatically roll the stage back to its last successful
        #             execution's state
        #   RETRY:    Automatically retry per retry_configuration
        #   FAIL:     Fail the pipeline execution (default behavior when omitted)
        result = optional(string, "")

        # retry_configuration tunes automatic retry when result is RETRY.
        retry_configuration = optional(object({
          # retry_mode selects what is retried.
          #   FAILED_ACTIONS: Only the failed actions re-run (default)
          #   ALL_ACTIONS:    The whole stage re-runs from its first action
          retry_mode = string
        }))

        # condition optionally gates the failure handling behind rules — the
        # result applies only when the rules pass.
        condition = optional(object({
          # result is the outcome applied when the condition's rules do NOT pass.
          #   FAIL: Fail the stage (default behavior when omitted)
          #   SKIP: Skip the stage and continue the pipeline (before_entry gates)
          result = optional(string, "")

          # rules are the checks that make up this condition (1-5, AND'd together).
          rules = list(object({
            # name is the rule name, unique within the condition.
            name = string

            # rule_type_id identifies which managed rule runs.
            rule_type_id = object({
              # provider is the managed rule provider: DeploymentWindow,
              # CloudWatchAlarm, LambdaInvoke, VariableCheck, or Commands.
              provider = string

              # category is the rule category. AWS currently supports only "Rule".
              category = optional(string)

              # owner identifies who provides the rule. AWS currently supports only
              # "AWS" (managed rules).
              owner = optional(string)

              # version is the rule type version. Typically "1"; AWS resolves the
              # current version when omitted.
              version = optional(string, "")
            })

            # configuration contains rule-provider-specific key-value pairs (e.g.,
            # DeploymentWindow: Cron, TimeZone; CloudWatchAlarm: AlarmName,
            # WaitTime; LambdaInvoke: FunctionName). Values are limited to 10,000
            # characters (AWS's rule-configuration value cap).
            configuration = optional(map(string), {})

            # commands are shell commands executed by the Commands rule provider
            # (max 50, each 1-1000 characters).
            commands = optional(list(string), [])

            # input_artifacts are artifact names available to the rule (e.g., for
            # Commands rules operating on build output). Each name 1-100 characters,
            # alphanumeric with underscores and hyphens.
            input_artifacts = optional(list(string), [])

            # region is the AWS region where the rule executes. Omit to run in the
            # pipeline's region.
            region = optional(string, "")

            # role_arn is an IAM role the rule assumes instead of the pipeline's
            # role (e.g., a scoped role for the LambdaInvoke rule).
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            role_arn = optional(string, "")

            # timeout_in_minutes is the maximum time the rule can run (5-86400).
            timeout_in_minutes = optional(number, 0)
          }))
        }))
      }))
    }))

    # triggers define automatic pipeline execution rules based on git events.
    # V2 pipelines only. Triggers use CodeStar Connections to listen for
    # push or pull request events on source repositories with branch, tag,
    # and file path filtering.
    triggers = optional(list(object({
      # provider_type is the trigger provider. Currently only
      # CodeStarSourceConnection is supported.
      provider_type = string

      # git_configuration defines the git event filters that trigger the pipeline.
      git_configuration = object({
        # source_action_name must match the name of a Source action in the first
        # stage that uses a CodeStarSourceConnection provider.
        source_action_name = string

        # push defines filters for git push events (branch pushes, tag pushes).
        # Multiple push filters are OR'd — a push triggers the pipeline if ANY
        # filter matches.
        push = optional(list(object({
          # branches filters by branch name patterns (glob syntax).
          branches = optional(object({
            # includes are glob patterns that must match for the filter to pass.
            # At least one include pattern must match. Each pattern 1-255 characters.
            includes = optional(list(string), [])

            # excludes are glob patterns that cause the filter to reject a match.
            # Exclusions take precedence over inclusions. Each pattern 1-255 characters.
            excludes = optional(list(string), [])
          }))

          # file_paths filters by changed file path patterns (glob syntax).
          file_paths = optional(object({
            # includes are glob patterns that must match for the filter to pass.
            # At least one include pattern must match. Each pattern 1-255 characters.
            includes = optional(list(string), [])

            # excludes are glob patterns that cause the filter to reject a match.
            # Exclusions take precedence over inclusions. Each pattern 1-255 characters.
            excludes = optional(list(string), [])
          }))

          # tags filters by tag name patterns (glob syntax).
          tags = optional(object({
            # includes are glob patterns that must match for the filter to pass.
            # At least one include pattern must match. Each pattern 1-255 characters.
            includes = optional(list(string), [])

            # excludes are glob patterns that cause the filter to reject a match.
            # Exclusions take precedence over inclusions. Each pattern 1-255 characters.
            excludes = optional(list(string), [])
          }))
        })), [])

        # pull_request defines filters for pull request events (open, update, close).
        # Multiple PR filters are OR'd.
        pull_request = optional(list(object({
          # branches filters by target branch name patterns (glob syntax).
          branches = optional(object({
            # includes are glob patterns that must match for the filter to pass.
            # At least one include pattern must match. Each pattern 1-255 characters.
            includes = optional(list(string), [])

            # excludes are glob patterns that cause the filter to reject a match.
            # Exclusions take precedence over inclusions. Each pattern 1-255 characters.
            excludes = optional(list(string), [])
          }))

          # file_paths filters by changed file path patterns (glob syntax).
          file_paths = optional(object({
            # includes are glob patterns that must match for the filter to pass.
            # At least one include pattern must match. Each pattern 1-255 characters.
            includes = optional(list(string), [])

            # excludes are glob patterns that cause the filter to reject a match.
            # Exclusions take precedence over inclusions. Each pattern 1-255 characters.
            excludes = optional(list(string), [])
          }))

          # events specifies which PR lifecycle events trigger the pipeline
          # (1-3 values). Valid values: OPEN, UPDATED, CLOSED.
          events = optional(list(string), [])
        })), [])
      })
    })), [])

    # variables define pipeline-level parameters that can be referenced in
    # action configurations using #{variables.VariableName} syntax.
    # V2 pipelines only.
    variables = optional(list(object({
      # name is the variable name. Referenced in action configurations as
      # #{variables.Name}.
      name = string

      # default_value is used when no value is supplied at execution time.
      default_value = optional(string, "")

      # description is a human-readable explanation of the variable's purpose.
      description = optional(string, "")
    })), [])
  })
}

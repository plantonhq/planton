locals {
  # Identity tags -- the same keys and values the Pulumi module writes.
  # resource_kind is the CatalogKind enum name lowercased, spelled as
  # that exact literal; resource_id is added (id_tag below) only when the
  # resource has an id, never with the name as a stand-in.
  base_tags = {
    "resource"      = "true"
    "resource_kind" = "azuremachinelearningbatchdeployment"
    "resource_name" = var.metadata.name
  }

  org_tag = (
    var.metadata.org != null && var.metadata.org != ""
  ) ? { "organization" = var.metadata.org } : {}

  env_tag = (
    var.metadata.env != null && var.metadata.env != ""
  ) ? { "environment" = var.metadata.env } : {}

  id_tag = (
    var.metadata.id != null && var.metadata.id != ""
  ) ? { "resource_id" = var.metadata.id } : {}

  # Metadata-derived tags first, then the user's spec tags merged over them:
  # user tags deliberately win so an org's governance conventions (cost
  # center, owner) can override the derived values where they collide.
  final_tags = merge(local.base_tags, local.org_tag, local.env_tag, local.id_tag, var.spec.tags)

  # The model reference: the variant block set in the spec IS the ARM
  # referenceType discriminator -- the spec's exactly-one rule guarantees
  # a single arm.
  model_reference = var.spec.model == null ? {} : (
    var.spec.model.id != null ? {
      model = {
        referenceType = "Id"
        assetId       = var.spec.model.id.asset_id
      }
      } : var.spec.model.data_path != null ? {
      model = merge(
        { referenceType = "DataPath" },
        var.spec.model.data_path.datastore_id != "" ? { datastoreId = var.spec.model.data_path.datastore_id } : {},
        var.spec.model.data_path.path != "" ? { path = var.spec.model.data_path.path } : {},
      )
      } : {
      model = merge(
        { referenceType = "OutputPath" },
        var.spec.model.output_path.job_id != "" ? { jobId = var.spec.model.output_path.job_id } : {},
        var.spec.model.output_path.path != "" ? { path = var.spec.model.output_path.path } : {},
      )
    }
  )

  # Per-job compute sizing. ARM's untyped resources.properties bag is
  # deliberately not modeled (recorded exclusion on the spec message).
  resources = var.spec.resources != null ? {
    resources = merge(
      var.spec.resources.instance_count != null ? { instanceCount = var.spec.resources.instance_count } : {},
      var.spec.resources.instance_type != "" ? { instanceType = var.spec.resources.instance_type } : {},
    )
  } : {}

  retry_settings = var.spec.retry_settings != null ? {
    retrySettings = merge(
      var.spec.retry_settings.max_retries != null ? { maxRetries = var.spec.retry_settings.max_retries } : {},
      var.spec.retry_settings.timeout != "" ? { timeout = var.spec.retry_settings.timeout } : {},
    )
  } : {}

  code_configuration = var.spec.code_configuration != null ? {
    codeConfiguration = merge(
      var.spec.code_configuration.code_id != "" ? { codeId = var.spec.code_configuration.code_id } : {},
      { scoringScript = var.spec.code_configuration.scoring_script },
    )
  } : {}

  # The PipelineComponent deployment type: present only when the spec's
  # block is -- absent means ARM's default Model type (the enum's other
  # value has no concrete ARM shape of its own).
  pipeline_component = var.spec.pipeline_component != null ? {
    deploymentConfiguration = merge(
      {
        deploymentConfigurationType = "PipelineComponent"
        componentId = {
          referenceType = "Id"
          assetId       = var.spec.pipeline_component.component_id
        }
      },
      length(var.spec.pipeline_component.settings) > 0 ? { settings = var.spec.pipeline_component.settings } : {},
      length(var.spec.pipeline_component.job_tags) > 0 ? { tags = var.spec.pipeline_component.job_tags } : {},
      var.spec.pipeline_component.job_description != "" ? { description = var.spec.pipeline_component.job_description } : {},
    )
  } : {}

  # The ARM properties object, assembled key-by-key so unset optionals
  # are OMITTED (ARM applies its own defaults: miniBatchSize 10,
  # maxConcurrencyPerInstance 1, errorThreshold -1, loggingLevel Info,
  # outputAction AppendRow, outputFileName predictions.csv).
  deployment_properties = merge(
    var.spec.compute_id != "" ? { compute = var.spec.compute_id } : {},
    var.spec.environment_id != "" ? { environmentId = var.spec.environment_id } : {},
    length(var.spec.environment_variables) > 0 ? { environmentVariables = var.spec.environment_variables } : {},
    var.spec.mini_batch_size != null ? { miniBatchSize = var.spec.mini_batch_size } : {},
    var.spec.max_concurrency_per_instance != null ? { maxConcurrencyPerInstance = var.spec.max_concurrency_per_instance } : {},
    var.spec.error_threshold != null ? { errorThreshold = var.spec.error_threshold } : {},
    var.spec.output_action != "" ? { outputAction = var.spec.output_action } : {},
    var.spec.output_file_name != "" ? { outputFileName = var.spec.output_file_name } : {},
    var.spec.logging_level != "" ? { loggingLevel = var.spec.logging_level } : {},
    var.spec.description != "" ? { description = var.spec.description } : {},
    length(var.spec.properties) > 0 ? { properties = var.spec.properties } : {},
    local.model_reference,
    local.resources,
    local.retry_settings,
    local.code_configuration,
    local.pipeline_component,
  )
}

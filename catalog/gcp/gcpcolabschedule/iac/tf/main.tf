# Enable the Vertex AI API first so a fresh project works on the first
# deploy. disable_on_destroy is false: tearing down one schedule must never
# disable the API for everything else in the project.
resource "google_project_service" "aiplatform_api" {
  project = local.project_id
  service = "aiplatform.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The project NUMBER a pipeline job's peered network path needs, read only
# when that path carries a project ID; a numeric path (or no network)
# performs no read.
data "google_project" "pipeline_network" {
  count      = local.pipeline_network_needs_number ? 1 : 0
  project_id = local.pipeline_network_proj
}

# The schedule. desired_state (ACTIVE / PAUSED) is a client-side control the
# provider enforces with pause and resume calls. Exactly one request block
# is rendered; the notebook request is immutable (a change replaces the
# schedule), the pipeline request updates in place. Each request's parent
# is left to the provider, which fills projects/{project}/locations/{location}.
resource "google_colab_schedule" "this" {
  project      = local.project_id
  location     = var.spec.location
  display_name = local.display_name
  cron         = var.spec.cron

  max_concurrent_run_count        = local.max_concurrent_run_count
  allow_queueing                  = var.spec.allow_queueing
  start_time                      = local.start_time
  end_time                        = local.end_time
  max_run_count                   = local.max_run_count
  max_concurrent_active_run_count = local.max_concurrent_active_run_count
  desired_state                   = local.desired_state

  # Client-side destroy behavior: DELETE (default), PREVENT, or ABANDON.
  deletion_policy = local.deletion_policy

  dynamic "create_notebook_execution_job_request" {
    for_each = local.notebook != null ? [local.notebook] : []
    content {
      notebook_execution_job {
        display_name                            = create_notebook_execution_job_request.value.display_name != "" ? create_notebook_execution_job_request.value.display_name : local.display_name
        gcs_output_uri                          = create_notebook_execution_job_request.value.gcs_output_uri
        notebook_runtime_template_resource_name = create_notebook_execution_job_request.value.notebook_runtime_template_resource_name != "" ? create_notebook_execution_job_request.value.notebook_runtime_template_resource_name : null
        execution_user                          = create_notebook_execution_job_request.value.execution_user != "" ? create_notebook_execution_job_request.value.execution_user : null
        service_account                         = create_notebook_execution_job_request.value.service_account != "" ? create_notebook_execution_job_request.value.service_account : null
        execution_timeout                       = create_notebook_execution_job_request.value.execution_timeout != "" ? create_notebook_execution_job_request.value.execution_timeout : null
        kernel_name                             = create_notebook_execution_job_request.value.kernel_name != "" ? create_notebook_execution_job_request.value.kernel_name : null
        labels                                  = length(create_notebook_execution_job_request.value.labels) > 0 ? create_notebook_execution_job_request.value.labels : null

        dynamic "gcs_notebook_source" {
          for_each = create_notebook_execution_job_request.value.gcs_notebook_source != null ? [create_notebook_execution_job_request.value.gcs_notebook_source] : []
          content {
            uri        = gcs_notebook_source.value.uri
            generation = gcs_notebook_source.value.generation != "" ? gcs_notebook_source.value.generation : null
          }
        }

        dynamic "dataform_repository_source" {
          for_each = create_notebook_execution_job_request.value.dataform_repository_source != null ? [create_notebook_execution_job_request.value.dataform_repository_source] : []
          content {
            dataform_repository_resource_name = dataform_repository_source.value.dataform_repository_resource_name
            commit_sha                        = dataform_repository_source.value.commit_sha != "" ? dataform_repository_source.value.commit_sha : null
          }
        }

        dynamic "custom_environment_spec" {
          for_each = create_notebook_execution_job_request.value.custom_environment_spec != null ? [create_notebook_execution_job_request.value.custom_environment_spec] : []
          content {
            dynamic "machine_spec" {
              for_each = custom_environment_spec.value.machine_spec != null ? [custom_environment_spec.value.machine_spec] : []
              content {
                machine_type       = machine_spec.value.machine_type != "" ? machine_spec.value.machine_type : null
                accelerator_type   = machine_spec.value.accelerator_type != "" ? machine_spec.value.accelerator_type : null
                accelerator_count  = machine_spec.value.accelerator_count != 0 ? machine_spec.value.accelerator_count : null
                gpu_partition_size = machine_spec.value.gpu_partition_size != "" ? machine_spec.value.gpu_partition_size : null
                tpu_topology       = machine_spec.value.tpu_topology != "" ? machine_spec.value.tpu_topology : null

                dynamic "reservation_affinity" {
                  for_each = machine_spec.value.reservation_affinity != null ? [machine_spec.value.reservation_affinity] : []
                  content {
                    reservation_affinity_type = reservation_affinity.value.reservation_affinity_type
                    key                       = reservation_affinity.value.key != "" ? reservation_affinity.value.key : null
                    values                    = length(reservation_affinity.value.values) > 0 ? reservation_affinity.value.values : null
                    use_reservation_pool      = reservation_affinity.value.use_reservation_pool
                  }
                }
              }
            }

            dynamic "network_spec" {
              for_each = custom_environment_spec.value.network_spec != null ? [custom_environment_spec.value.network_spec] : []
              content {
                enable_internet_access = network_spec.value.enable_internet_access
                network                = network_spec.value.network != "" ? network_spec.value.network : null
                subnetwork             = local.notebook_subnetwork
              }
            }

            # Google types the disk size as a decimal string.
            dynamic "persistent_disk_spec" {
              for_each = custom_environment_spec.value.persistent_disk_spec != null ? [custom_environment_spec.value.persistent_disk_spec] : []
              content {
                disk_type    = persistent_disk_spec.value.disk_type != "" ? persistent_disk_spec.value.disk_type : null
                disk_size_gb = persistent_disk_spec.value.disk_size_gb != 0 ? tostring(persistent_disk_spec.value.disk_size_gb) : null
              }
            }
          }
        }

        # Google's workbench_runtime is an empty marker block; the spec
        # carries it as a bool.
        dynamic "workbench_runtime" {
          for_each = create_notebook_execution_job_request.value.workbench_runtime ? [true] : []
          content {}
        }

        dynamic "encryption_spec" {
          for_each = create_notebook_execution_job_request.value.kms_key_name != "" ? [create_notebook_execution_job_request.value.kms_key_name] : []
          content {
            kms_key_name = encryption_spec.value
          }
        }
      }
    }
  }

  dynamic "create_pipeline_job_request" {
    for_each = local.pipeline != null ? [local.pipeline] : []
    content {
      pipeline_job {
        display_name          = create_pipeline_job_request.value.display_name != "" ? create_pipeline_job_request.value.display_name : null
        pipeline_spec         = create_pipeline_job_request.value.pipeline_spec != "" ? create_pipeline_job_request.value.pipeline_spec : null
        template_uri          = create_pipeline_job_request.value.template_uri != "" ? create_pipeline_job_request.value.template_uri : null
        service_account       = create_pipeline_job_request.value.service_account != "" ? create_pipeline_job_request.value.service_account : null
        network               = local.pipeline_network
        reserved_ip_ranges    = length(create_pipeline_job_request.value.reserved_ip_ranges) > 0 ? create_pipeline_job_request.value.reserved_ip_ranges : null
        preflight_validations = create_pipeline_job_request.value.preflight_validations
        labels                = length(create_pipeline_job_request.value.labels) > 0 ? create_pipeline_job_request.value.labels : null

        dynamic "runtime_config" {
          for_each = create_pipeline_job_request.value.runtime_config != null ? [create_pipeline_job_request.value.runtime_config] : []
          content {
            gcs_output_directory = runtime_config.value.gcs_output_directory
            failure_policy       = runtime_config.value.failure_policy != "" ? runtime_config.value.failure_policy : null
            parameter_values     = length(runtime_config.value.parameter_values) > 0 ? runtime_config.value.parameter_values : null
          }
        }

        dynamic "encryption_spec" {
          for_each = create_pipeline_job_request.value.kms_key_name != "" ? [create_pipeline_job_request.value.kms_key_name] : []
          content {
            kms_key_name = encryption_spec.value
          }
        }

        dynamic "psc_interface_config" {
          for_each = create_pipeline_job_request.value.psc_interface_config != null ? [create_pipeline_job_request.value.psc_interface_config] : []
          content {
            network_attachment = psc_interface_config.value.network_attachment != "" ? psc_interface_config.value.network_attachment : null

            dynamic "dns_peering_configs" {
              for_each = psc_interface_config.value.dns_peering_configs
              content {
                domain         = dns_peering_configs.value.domain
                target_network = dns_peering_configs.value.target_network
                target_project = dns_peering_configs.value.target_project
              }
            }
          }
        }
      }
    }
  }

  depends_on = [google_project_service.aiplatform_api]
}

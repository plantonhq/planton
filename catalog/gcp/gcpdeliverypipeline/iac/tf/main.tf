# The Cloud Deploy API. disable_on_destroy is false: tearing down one
# pipeline must never disable the API for every other pipeline in the
# project.
resource "google_project_service" "clouddeploy_api" {
  project = local.project_id
  service = "clouddeploy.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The delivery pipeline. Location and ID are immutable; the stages and
# their strategies, labels, annotations, and suspension update in place.
# The provider always deletes with force=true, removing the pipeline's
# releases, rollouts, and automations with it.
resource "google_clouddeploy_delivery_pipeline" "this" {
  project         = local.project_id
  location        = var.spec.location
  name            = local.delivery_pipeline_id
  description     = var.spec.description != "" ? var.spec.description : null
  labels          = local.final_labels
  annotations     = length(var.spec.annotations) > 0 ? var.spec.annotations : null
  suspended       = var.spec.suspended ? true : null
  deletion_policy = local.deletion_policy

  dynamic "serial_pipeline" {
    for_each = var.spec.serial_pipeline != null ? [var.spec.serial_pipeline] : []
    content {
      dynamic "stages" {
        for_each = local.stages
        content {
          target_id = stages.value.target_id != "" ? stages.value.target_id : null
          profiles  = length(stages.value.profiles) > 0 ? stages.value.profiles : null

          dynamic "deploy_parameters" {
            for_each = stages.value.deploy_parameters
            content {
              values              = deploy_parameters.value.values
              match_target_labels = length(deploy_parameters.value.match_target_labels) > 0 ? deploy_parameters.value.match_target_labels : null
            }
          }

          dynamic "strategy" {
            for_each = stages.value.strategy != null ? [stages.value.strategy] : []
            content {
              # The standard strategy: one deploy, with optional jobs around
              # it. predeploy and postdeploy take actions or tasks.
              dynamic "standard" {
                for_each = strategy.value.standard != null ? [strategy.value.standard] : []
                content {
                  verify = standard.value.verify ? true : null

                  dynamic "predeploy" {
                    for_each = standard.value.predeploy != null ? [standard.value.predeploy] : []
                    content {
                      actions = length(predeploy.value.actions) > 0 ? predeploy.value.actions : null

                      dynamic "tasks" {
                        for_each = predeploy.value.tasks
                        content {
                          dynamic "container" {
                            for_each = tasks.value.container != null ? [tasks.value.container] : []
                            content {
                              image   = container.value.image
                              command = length(container.value.command) > 0 ? container.value.command : null
                              args    = length(container.value.args) > 0 ? container.value.args : null
                              env     = length(container.value.env) > 0 ? container.value.env : null
                            }
                          }
                        }
                      }
                    }
                  }

                  dynamic "postdeploy" {
                    for_each = standard.value.postdeploy != null ? [standard.value.postdeploy] : []
                    content {
                      actions = length(postdeploy.value.actions) > 0 ? postdeploy.value.actions : null

                      dynamic "tasks" {
                        for_each = postdeploy.value.tasks
                        content {
                          dynamic "container" {
                            for_each = tasks.value.container != null ? [tasks.value.container] : []
                            content {
                              image   = container.value.image
                              command = length(container.value.command) > 0 ? container.value.command : null
                              args    = length(container.value.args) > 0 ? container.value.args : null
                              env     = length(container.value.env) > 0 ? container.value.env : null
                            }
                          }
                        }
                      }
                    }
                  }

                  dynamic "verify_config" {
                    for_each = standard.value.verify_config != null ? [standard.value.verify_config] : []
                    content {
                      dynamic "tasks" {
                        for_each = verify_config.value.tasks
                        content {
                          dynamic "container" {
                            for_each = tasks.value.container != null ? [tasks.value.container] : []
                            content {
                              image   = container.value.image
                              command = length(container.value.command) > 0 ? container.value.command : null
                              args    = length(container.value.args) > 0 ? container.value.args : null
                              env     = length(container.value.env) > 0 ? container.value.env : null
                            }
                          }
                        }
                      }
                    }
                  }

                  dynamic "analysis" {
                    for_each = standard.value.analysis != null ? [standard.value.analysis] : []
                    content {
                      duration = analysis.value.duration

                      dynamic "google_cloud" {
                        for_each = analysis.value.google_cloud != null ? [analysis.value.google_cloud] : []
                        content {
                          dynamic "alert_policy_checks" {
                            for_each = google_cloud.value.alert_policy_checks
                            content {
                              id             = alert_policy_checks.value.id
                              alert_policies = alert_policy_checks.value.alert_policies
                              labels         = length(alert_policy_checks.value.labels) > 0 ? alert_policy_checks.value.labels : null
                            }
                          }
                        }
                      }

                      dynamic "custom_checks" {
                        for_each = analysis.value.custom_checks
                        content {
                          id        = custom_checks.value.id
                          frequency = custom_checks.value.frequency != "" ? custom_checks.value.frequency : null

                          dynamic "task" {
                            for_each = custom_checks.value.task != null ? [custom_checks.value.task] : []
                            content {
                              dynamic "container" {
                                for_each = task.value.container != null ? [task.value.container] : []
                                content {
                                  image   = container.value.image
                                  command = length(container.value.command) > 0 ? container.value.command : null
                                  args    = length(container.value.args) > 0 ? container.value.args : null
                                  env     = length(container.value.env) > 0 ? container.value.env : null
                                }
                              }
                            }
                          }
                        }
                      }
                    }
                  }
                }
              }

              # The canary strategy: progressive deployment on Cloud Run or
              # GKE. Its predeploy and postdeploy take actions only (the
              # provider declares no tasks on the canary paths).
              dynamic "canary" {
                for_each = strategy.value.canary != null ? [strategy.value.canary] : []
                content {
                  dynamic "canary_deployment" {
                    for_each = canary.value.canary_deployment != null ? [canary.value.canary_deployment] : []
                    content {
                      percentages = canary_deployment.value.percentages
                      verify      = canary_deployment.value.verify ? true : null

                      dynamic "predeploy" {
                        for_each = canary_deployment.value.predeploy != null ? [canary_deployment.value.predeploy] : []
                        content {
                          actions = length(predeploy.value.actions) > 0 ? predeploy.value.actions : null
                        }
                      }

                      dynamic "postdeploy" {
                        for_each = canary_deployment.value.postdeploy != null ? [canary_deployment.value.postdeploy] : []
                        content {
                          actions = length(postdeploy.value.actions) > 0 ? postdeploy.value.actions : null
                        }
                      }

                      dynamic "verify_config" {
                        for_each = canary_deployment.value.verify_config != null ? [canary_deployment.value.verify_config] : []
                        content {
                          dynamic "tasks" {
                            for_each = verify_config.value.tasks
                            content {
                              dynamic "container" {
                                for_each = tasks.value.container != null ? [tasks.value.container] : []
                                content {
                                  image   = container.value.image
                                  command = length(container.value.command) > 0 ? container.value.command : null
                                  args    = length(container.value.args) > 0 ? container.value.args : null
                                  env     = length(container.value.env) > 0 ? container.value.env : null
                                }
                              }
                            }
                          }
                        }
                      }

                      dynamic "analysis" {
                        for_each = canary_deployment.value.analysis != null ? [canary_deployment.value.analysis] : []
                        content {
                          duration = analysis.value.duration

                          dynamic "google_cloud" {
                            for_each = analysis.value.google_cloud != null ? [analysis.value.google_cloud] : []
                            content {
                              dynamic "alert_policy_checks" {
                                for_each = google_cloud.value.alert_policy_checks
                                content {
                                  id             = alert_policy_checks.value.id
                                  alert_policies = alert_policy_checks.value.alert_policies
                                  labels         = length(alert_policy_checks.value.labels) > 0 ? alert_policy_checks.value.labels : null
                                }
                              }
                            }
                          }

                          dynamic "custom_checks" {
                            for_each = analysis.value.custom_checks
                            content {
                              id        = custom_checks.value.id
                              frequency = custom_checks.value.frequency != "" ? custom_checks.value.frequency : null

                              dynamic "task" {
                                for_each = custom_checks.value.task != null ? [custom_checks.value.task] : []
                                content {
                                  dynamic "container" {
                                    for_each = task.value.container != null ? [task.value.container] : []
                                    content {
                                      image   = container.value.image
                                      command = length(container.value.command) > 0 ? container.value.command : null
                                      args    = length(container.value.args) > 0 ? container.value.args : null
                                      env     = length(container.value.env) > 0 ? container.value.env : null
                                    }
                                  }
                                }
                              }
                            }
                          }
                        }
                      }
                    }
                  }

                  dynamic "custom_canary_deployment" {
                    for_each = canary.value.custom_canary_deployment != null ? [canary.value.custom_canary_deployment] : []
                    content {
                      dynamic "phase_configs" {
                        for_each = custom_canary_deployment.value.phase_configs
                        content {
                          phase_id   = phase_configs.value.phase_id
                          percentage = phase_configs.value.percentage
                          profiles   = length(phase_configs.value.profiles) > 0 ? phase_configs.value.profiles : null
                          verify     = phase_configs.value.verify ? true : null

                          dynamic "predeploy" {
                            for_each = phase_configs.value.predeploy != null ? [phase_configs.value.predeploy] : []
                            content {
                              actions = length(predeploy.value.actions) > 0 ? predeploy.value.actions : null
                            }
                          }

                          dynamic "postdeploy" {
                            for_each = phase_configs.value.postdeploy != null ? [phase_configs.value.postdeploy] : []
                            content {
                              actions = length(postdeploy.value.actions) > 0 ? postdeploy.value.actions : null
                            }
                          }

                          dynamic "verify_config" {
                            for_each = phase_configs.value.verify_config != null ? [phase_configs.value.verify_config] : []
                            content {
                              dynamic "tasks" {
                                for_each = verify_config.value.tasks
                                content {
                                  dynamic "container" {
                                    for_each = tasks.value.container != null ? [tasks.value.container] : []
                                    content {
                                      image   = container.value.image
                                      command = length(container.value.command) > 0 ? container.value.command : null
                                      args    = length(container.value.args) > 0 ? container.value.args : null
                                      env     = length(container.value.env) > 0 ? container.value.env : null
                                    }
                                  }
                                }
                              }
                            }
                          }

                          dynamic "analysis" {
                            for_each = phase_configs.value.analysis != null ? [phase_configs.value.analysis] : []
                            content {
                              duration = analysis.value.duration

                              dynamic "google_cloud" {
                                for_each = analysis.value.google_cloud != null ? [analysis.value.google_cloud] : []
                                content {
                                  dynamic "alert_policy_checks" {
                                    for_each = google_cloud.value.alert_policy_checks
                                    content {
                                      id             = alert_policy_checks.value.id
                                      alert_policies = alert_policy_checks.value.alert_policies
                                      labels         = length(alert_policy_checks.value.labels) > 0 ? alert_policy_checks.value.labels : null
                                    }
                                  }
                                }
                              }

                              dynamic "custom_checks" {
                                for_each = analysis.value.custom_checks
                                content {
                                  id        = custom_checks.value.id
                                  frequency = custom_checks.value.frequency != "" ? custom_checks.value.frequency : null

                                  dynamic "task" {
                                    for_each = custom_checks.value.task != null ? [custom_checks.value.task] : []
                                    content {
                                      dynamic "container" {
                                        for_each = task.value.container != null ? [task.value.container] : []
                                        content {
                                          image   = container.value.image
                                          command = length(container.value.command) > 0 ? container.value.command : null
                                          args    = length(container.value.args) > 0 ? container.value.args : null
                                          env     = length(container.value.env) > 0 ? container.value.env : null
                                        }
                                      }
                                    }
                                  }
                                }
                              }
                            }
                          }
                        }
                      }
                    }
                  }

                  dynamic "runtime_config" {
                    for_each = canary.value.runtime_config != null ? [canary.value.runtime_config] : []
                    content {
                      dynamic "cloud_run" {
                        for_each = runtime_config.value.cloud_run != null ? [runtime_config.value.cloud_run] : []
                        content {
                          automatic_traffic_control = cloud_run.value.automatic_traffic_control ? true : null
                          canary_revision_tags      = length(cloud_run.value.canary_revision_tags) > 0 ? cloud_run.value.canary_revision_tags : null
                          prior_revision_tags       = length(cloud_run.value.prior_revision_tags) > 0 ? cloud_run.value.prior_revision_tags : null
                          stable_revision_tags      = length(cloud_run.value.stable_revision_tags) > 0 ? cloud_run.value.stable_revision_tags : null
                        }
                      }

                      dynamic "kubernetes" {
                        for_each = runtime_config.value.kubernetes != null ? [runtime_config.value.kubernetes] : []
                        content {
                          dynamic "gateway_service_mesh" {
                            for_each = kubernetes.value.gateway_service_mesh != null ? [kubernetes.value.gateway_service_mesh] : []
                            content {
                              http_route              = gateway_service_mesh.value.http_route
                              service                 = gateway_service_mesh.value.service
                              deployment              = gateway_service_mesh.value.deployment
                              route_update_wait_time  = gateway_service_mesh.value.route_update_wait_time != "" ? gateway_service_mesh.value.route_update_wait_time : null
                              stable_cutback_duration = gateway_service_mesh.value.stable_cutback_duration != "" ? gateway_service_mesh.value.stable_cutback_duration : null
                              pod_selector_label      = gateway_service_mesh.value.pod_selector_label != "" ? gateway_service_mesh.value.pod_selector_label : null

                              dynamic "route_destinations" {
                                for_each = gateway_service_mesh.value.route_destinations != null ? [gateway_service_mesh.value.route_destinations] : []
                                content {
                                  destination_ids   = route_destinations.value.destination_ids
                                  propagate_service = route_destinations.value.propagate_service ? true : null
                                }
                              }
                            }
                          }

                          dynamic "service_networking" {
                            for_each = kubernetes.value.service_networking != null ? [kubernetes.value.service_networking] : []
                            content {
                              service                      = service_networking.value.service
                              deployment                   = service_networking.value.deployment
                              disable_pod_overprovisioning = service_networking.value.disable_pod_overprovisioning ? true : null
                              pod_selector_label           = service_networking.value.pod_selector_label != "" ? service_networking.value.pod_selector_label : null
                            }
                          }
                        }
                      }
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }

  depends_on = [google_project_service.clouddeploy_api]
}

# The pipeline's automations, one per spec.automations[] entry, in the
# pipeline's project and region. Each rule sets exactly one rule kind;
# every repair phase exactly one of retry or rollback.
resource "google_clouddeploy_automation" "this" {
  for_each = local.automations

  project           = local.project_id
  location          = var.spec.location
  delivery_pipeline = google_clouddeploy_delivery_pipeline.this.name
  name              = each.key
  description       = each.value.description != "" ? each.value.description : null
  labels            = merge(each.value.labels, local.attribution_labels)
  annotations       = length(each.value.annotations) > 0 ? each.value.annotations : null
  suspended         = each.value.suspended ? true : null
  service_account   = each.value.service_account
  deletion_policy   = local.deletion_policy

  selector {
    dynamic "targets" {
      for_each = each.value.selector.targets
      content {
        id     = targets.value.id != "" ? targets.value.id : null
        labels = length(targets.value.labels) > 0 ? targets.value.labels : null
      }
    }
  }

  dynamic "rules" {
    for_each = each.value.rules
    content {
      dynamic "advance_rollout_rule" {
        for_each = rules.value.advance_rollout_rule != null ? [rules.value.advance_rollout_rule] : []
        content {
          id            = advance_rollout_rule.value.id
          source_phases = length(advance_rollout_rule.value.source_phases) > 0 ? advance_rollout_rule.value.source_phases : null
          wait          = advance_rollout_rule.value.wait != "" ? advance_rollout_rule.value.wait : null
        }
      }

      dynamic "promote_release_rule" {
        for_each = rules.value.promote_release_rule != null ? [rules.value.promote_release_rule] : []
        content {
          id                    = promote_release_rule.value.id
          wait                  = promote_release_rule.value.wait != "" ? promote_release_rule.value.wait : null
          destination_target_id = promote_release_rule.value.destination_target_id != "" ? promote_release_rule.value.destination_target_id : null
          destination_phase     = promote_release_rule.value.destination_phase != "" ? promote_release_rule.value.destination_phase : null
        }
      }

      dynamic "repair_rollout_rule" {
        for_each = rules.value.repair_rollout_rule != null ? [rules.value.repair_rollout_rule] : []
        content {
          id     = repair_rollout_rule.value.id
          phases = length(repair_rollout_rule.value.phases) > 0 ? repair_rollout_rule.value.phases : null
          jobs   = length(repair_rollout_rule.value.jobs) > 0 ? repair_rollout_rule.value.jobs : null

          dynamic "repair_phases" {
            for_each = repair_rollout_rule.value.repair_phases
            content {
              dynamic "retry" {
                for_each = repair_phases.value.retry != null ? [repair_phases.value.retry] : []
                content {
                  attempts     = retry.value.attempts
                  wait         = retry.value.wait != "" ? retry.value.wait : null
                  backoff_mode = retry.value.backoff_mode != "" ? retry.value.backoff_mode : null
                }
              }

              dynamic "rollback" {
                for_each = repair_phases.value.rollback != null ? [repair_phases.value.rollback] : []
                content {
                  destination_phase                   = rollback.value.destination_phase != "" ? rollback.value.destination_phase : null
                  disable_rollback_if_rollout_pending = rollback.value.disable_rollback_if_rollout_pending ? true : null
                }
              }
            }
          }
        }
      }

      dynamic "timed_promote_release_rule" {
        for_each = rules.value.timed_promote_release_rule != null ? [rules.value.timed_promote_release_rule] : []
        content {
          id                    = timed_promote_release_rule.value.id
          schedule              = timed_promote_release_rule.value.schedule
          time_zone             = timed_promote_release_rule.value.time_zone
          destination_target_id = timed_promote_release_rule.value.destination_target_id != "" ? timed_promote_release_rule.value.destination_target_id : null
          destination_phase     = timed_promote_release_rule.value.destination_phase != "" ? timed_promote_release_rule.value.destination_phase : null
        }
      }
    }
  }
}

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
  description = "GcpDeliveryPipeline specification"
  type = object({
    # The project the pipeline lives in: a literal project ID or a
    # GcpProject reference. Empty means the provider's default project. The
    # stages' targets and the automations live in the same project.
    # Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The region the pipeline lives in, e.g. "us-central1". Every stage's
    # target is looked up in this region. Required. Immutable.
    location = string

    # The pipeline's ID, unique in the project and region: 1-63 lowercase
    # letters, digits, and hyphens, starting with a letter and not ending
    # with a hyphen. Defaults to metadata.name. Immutable.
    delivery_pipeline_id = optional(string, "")

    # A description shown in the console, up to 255 characters.
    description = optional(string, "")

    # Labels on the pipeline: lowercase keys and values, at most 64 labels.
    # The platform attribution labels are added on top and win on key
    # conflicts. Deploy policies can select pipelines by label.
    labels = optional(map(string), {})

    # Annotations on the pipeline (Google's AIP-128 key/value metadata that
    # Cloud Deploy never reads). Only the keys declared here are managed.
    annotations = optional(map(string), {})

    # Suspend the pipeline: no new releases or rollouts can be created, but
    # in-progress ones complete. Updates in place.
    suspended = optional(bool, false)

    # The promotion flow: the ordered stages a release moves through.
    serial_pipeline = optional(object({
      # The stages, first to last. A release rolls out to the first stage's
      # target and is promoted to each next one.
      stages = optional(list(object({
        # The target this stage deploys to, by its bare ID ("prod", never
        # projects/.../targets/prod): a GcpDeployTarget reference (its target_id
        # output) or a literal ID. Google looks the target up in this pipeline's
        # project and region, so a referenced target must live there too.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        target_id = optional(string, "")

        # Skaffold profiles to render this stage's manifests with, e.g.
        # ["prod"] to pick the prod overlay of a skaffold.yaml.
        profiles = optional(list(string), [])

        # The rollout strategy for this stage. Empty means a standard deploy
        # without verification.
        strategy = optional(object({
          # Deploy the release in one step, optionally verifying it and running
          # jobs before and after.
          standard = optional(object({
            # Run `skaffold verify` (or verify_config's tasks) after the deploy; a
            # failed verification fails the rollout.
            verify = optional(bool, false)

            # A job that runs before the deploy.
            predeploy = optional(object({
              # Skaffold custom actions (customActions names in skaffold.yaml) to run
              # in order.
              actions = optional(list(string), [])

              # Containers to run in order, in Cloud Deploy's Cloud Build execution
              # environment.
              tasks = optional(list(object({
                # A container run in Cloud Deploy's Cloud Build execution environment.
                container = optional(object({
                  # The container image, e.g. "us-docker.pkg.dev/my-project/tools/smoke:1".
                  # Required.
                  image = string

                  # Overrides the image's entrypoint.
                  command = optional(list(string), [])

                  # Overrides the image's default arguments.
                  args = optional(list(string), [])

                  # Environment variables set in the container. Values are stored on the
                  # pipeline in plain text: never put secrets here.
                  env = optional(map(string), {})
                }))
              })), [])
            }))

            # A job that runs after the deploy.
            postdeploy = optional(object({
              # Skaffold custom actions (customActions names in skaffold.yaml) to run
              # in order.
              actions = optional(list(string), [])

              # Containers to run in order, in Cloud Deploy's Cloud Build execution
              # environment.
              tasks = optional(list(object({
                # A container run in Cloud Deploy's Cloud Build execution environment.
                container = optional(object({
                  # The container image, e.g. "us-docker.pkg.dev/my-project/tools/smoke:1".
                  # Required.
                  image = string

                  # Overrides the image's entrypoint.
                  command = optional(list(string), [])

                  # Overrides the image's default arguments.
                  args = optional(list(string), [])

                  # Environment variables set in the container. Values are stored on the
                  # pipeline in plain text: never put secrets here.
                  env = optional(map(string), {})
                }))
              })), [])
            }))

            # Container tasks that run as the verify job, instead of the release's
            # Skaffold verify configuration.
            verify_config = optional(object({
              # The containers to run in order.
              tasks = optional(list(object({
                # A container run in Cloud Deploy's Cloud Build execution environment.
                container = optional(object({
                  # The container image, e.g. "us-docker.pkg.dev/my-project/tools/smoke:1".
                  # Required.
                  image = string

                  # Overrides the image's entrypoint.
                  command = optional(list(string), [])

                  # Overrides the image's default arguments.
                  args = optional(list(string), [])

                  # Environment variables set in the container. Values are stored on the
                  # pipeline in plain text: never put secrets here.
                  env = optional(map(string), {})
                }))
              })), [])
            }))

            # An analysis job that watches the deploy for a while and fails the
            # rollout when a check fails.
            analysis = optional(object({
              # How long the analysis runs, in seconds format (e.g. "600s"). Required.
              duration = string

              # Checks against Cloud Monitoring alert policies.
              google_cloud = optional(object({
                # Alert-policy checks: the analysis fails when a listed policy fires.
                alert_policy_checks = optional(list(object({
                  # The check's ID, unique in the analysis. Required.
                  id = string

                  # The alert policies to watch, each projects/{project}/alertPolicies/{id}:
                  # GcpMonitoringAlertPolicy references (their policy_name output) or
                  # literals. Required.
                  # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
                  alert_policies = list(string)

                  # Labels that filter which incidents of those policies count.
                  labels = optional(map(string), {})
                })), [])
              }))

              # Checks that run your own container.
              custom_checks = optional(list(object({
                # The check's ID, unique in the analysis. Required.
                id = string

                # How often the check runs, in seconds format (e.g. "60s"). Empty uses
                # Cloud Deploy's default.
                frequency = optional(string, "")

                # The container the check runs.
                task = optional(object({
                  # A container run in Cloud Deploy's Cloud Build execution environment.
                  container = optional(object({
                    # The container image, e.g. "us-docker.pkg.dev/my-project/tools/smoke:1".
                    # Required.
                    image = string

                    # Overrides the image's entrypoint.
                    command = optional(list(string), [])

                    # Overrides the image's default arguments.
                    args = optional(list(string), [])

                    # Environment variables set in the container. Values are stored on the
                    # pipeline in plain text: never put secrets here.
                    env = optional(map(string), {})
                  }))
                }))
              })), [])
            }))
          }))

          # Deploy the release progressively by percentage.
          canary = optional(object({
            # The same steps for every phase: a list of percentages.
            canary_deployment = optional(object({
              # The percentages to deploy, in ascending order, e.g. [25, 50] (the
              # final 100% stable phase is implicit). Each is 0 <= n < 100; with a
              # Kubernetes Gateway API service mesh, 100 is also allowed. Required.
              percentages = list(number)

              # Run verify tests after each percentage deployment.
              verify = optional(bool, false)

              # A job that runs before the first phase.
              predeploy = optional(object({
                # Skaffold custom actions (customActions names in skaffold.yaml) to run
                # in order.
                actions = optional(list(string), [])
              }))

              # A job that runs after the last phase.
              postdeploy = optional(object({
                # Skaffold custom actions (customActions names in skaffold.yaml) to run
                # in order.
                actions = optional(list(string), [])
              }))

              # Container tasks that run as the verify job.
              verify_config = optional(object({
                # The containers to run in order.
                tasks = optional(list(object({
                  # A container run in Cloud Deploy's Cloud Build execution environment.
                  container = optional(object({
                    # The container image, e.g. "us-docker.pkg.dev/my-project/tools/smoke:1".
                    # Required.
                    image = string

                    # Overrides the image's entrypoint.
                    command = optional(list(string), [])

                    # Overrides the image's default arguments.
                    args = optional(list(string), [])

                    # Environment variables set in the container. Values are stored on the
                    # pipeline in plain text: never put secrets here.
                    env = optional(map(string), {})
                  }))
                })), [])
              }))

              # An analysis job for each phase.
              analysis = optional(object({
                # How long the analysis runs, in seconds format (e.g. "600s"). Required.
                duration = string

                # Checks against Cloud Monitoring alert policies.
                google_cloud = optional(object({
                  # Alert-policy checks: the analysis fails when a listed policy fires.
                  alert_policy_checks = optional(list(object({
                    # The check's ID, unique in the analysis. Required.
                    id = string

                    # The alert policies to watch, each projects/{project}/alertPolicies/{id}:
                    # GcpMonitoringAlertPolicy references (their policy_name output) or
                    # literals. Required.
                    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
                    alert_policies = list(string)

                    # Labels that filter which incidents of those policies count.
                    labels = optional(map(string), {})
                  })), [])
                }))

                # Checks that run your own container.
                custom_checks = optional(list(object({
                  # The check's ID, unique in the analysis. Required.
                  id = string

                  # How often the check runs, in seconds format (e.g. "60s"). Empty uses
                  # Cloud Deploy's default.
                  frequency = optional(string, "")

                  # The container the check runs.
                  task = optional(object({
                    # A container run in Cloud Deploy's Cloud Build execution environment.
                    container = optional(object({
                      # The container image, e.g. "us-docker.pkg.dev/my-project/tools/smoke:1".
                      # Required.
                      image = string

                      # Overrides the image's entrypoint.
                      command = optional(list(string), [])

                      # Overrides the image's default arguments.
                      args = optional(list(string), [])

                      # Environment variables set in the container. Values are stored on the
                      # pipeline in plain text: never put secrets here.
                      env = optional(map(string), {})
                    }))
                  }))
                })), [])
              }))
            }))

            # Phase-by-phase control: each phase its own percentage, profiles, and
            # jobs.
            custom_canary_deployment = optional(object({
              # The phases in the order they run. Required.
              phase_configs = list(object({
                # The rollout phase's ID: lowercase letters, digits, and hyphens,
                # starting with a letter, not ending with a hyphen, up to 63 characters.
                # Required.
                phase_id = string

                # The percentage this phase deploys, 0-100 (100 is the stable phase).
                # Always sent.
                percentage = optional(number, 0)

                # Skaffold profiles for this phase, in addition to the stage's profiles.
                profiles = optional(list(string), [])

                # Run verify tests after this phase's deployment.
                verify = optional(bool, false)

                # A job that runs before this phase.
                predeploy = optional(object({
                  # Skaffold custom actions (customActions names in skaffold.yaml) to run
                  # in order.
                  actions = optional(list(string), [])
                }))

                # A job that runs after this phase.
                postdeploy = optional(object({
                  # Skaffold custom actions (customActions names in skaffold.yaml) to run
                  # in order.
                  actions = optional(list(string), [])
                }))

                # Container tasks that run as this phase's verify job.
                verify_config = optional(object({
                  # The containers to run in order.
                  tasks = optional(list(object({
                    # A container run in Cloud Deploy's Cloud Build execution environment.
                    container = optional(object({
                      # The container image, e.g. "us-docker.pkg.dev/my-project/tools/smoke:1".
                      # Required.
                      image = string

                      # Overrides the image's entrypoint.
                      command = optional(list(string), [])

                      # Overrides the image's default arguments.
                      args = optional(list(string), [])

                      # Environment variables set in the container. Values are stored on the
                      # pipeline in plain text: never put secrets here.
                      env = optional(map(string), {})
                    }))
                  })), [])
                }))

                # An analysis job for this phase.
                analysis = optional(object({
                  # How long the analysis runs, in seconds format (e.g. "600s"). Required.
                  duration = string

                  # Checks against Cloud Monitoring alert policies.
                  google_cloud = optional(object({
                    # Alert-policy checks: the analysis fails when a listed policy fires.
                    alert_policy_checks = optional(list(object({
                      # The check's ID, unique in the analysis. Required.
                      id = string

                      # The alert policies to watch, each projects/{project}/alertPolicies/{id}:
                      # GcpMonitoringAlertPolicy references (their policy_name output) or
                      # literals. Required.
                      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
                      alert_policies = list(string)

                      # Labels that filter which incidents of those policies count.
                      labels = optional(map(string), {})
                    })), [])
                  }))

                  # Checks that run your own container.
                  custom_checks = optional(list(object({
                    # The check's ID, unique in the analysis. Required.
                    id = string

                    # How often the check runs, in seconds format (e.g. "60s"). Empty uses
                    # Cloud Deploy's default.
                    frequency = optional(string, "")

                    # The container the check runs.
                    task = optional(object({
                      # A container run in Cloud Deploy's Cloud Build execution environment.
                      container = optional(object({
                        # The container image, e.g. "us-docker.pkg.dev/my-project/tools/smoke:1".
                        # Required.
                        image = string

                        # Overrides the image's entrypoint.
                        command = optional(list(string), [])

                        # Overrides the image's default arguments.
                        args = optional(list(string), [])

                        # Environment variables set in the container. Values are stored on the
                        # pipeline in plain text: never put secrets here.
                        env = optional(map(string), {})
                      }))
                    }))
                  })), [])
                }))
              }))
            }))

            # How Cloud Deploy splits traffic during the canary: Cloud Run revision
            # traffic, or a Kubernetes Gateway API route or Service. Empty lets
            # Cloud Deploy pick by the target type (a GKE canary then needs one of
            # the Kubernetes arms).
            runtime_config = optional(object({
              # Cloud Run: Cloud Deploy shifts revision traffic.
              cloud_run = optional(object({
                # Let Cloud Deploy rewrite the service's traffic stanza to split traffic
                # between revisions. Required true for a canary_deployment; optional for
                # a custom_canary_deployment (where your manifests may split traffic
                # themselves).
                automatic_traffic_control = optional(bool, false)

                # Revision tags added to the canary revision while a canary phase runs.
                canary_revision_tags = optional(list(string), [])

                # Revision tags added to the prior revision while a canary phase runs.
                prior_revision_tags = optional(list(string), [])

                # Revision tags added to the new stable revision when the stable phase
                # is applied.
                stable_revision_tags = optional(list(string), [])
              }))

              # GKE: Cloud Deploy shifts traffic through a Gateway API route or a
              # Service.
              kubernetes = optional(object({
                # Split traffic with a Gateway API HTTPRoute (service-mesh or Gateway
                # setups).
                gateway_service_mesh = optional(object({
                  # The HTTPRoute whose weights Cloud Deploy adjusts. Required.
                  http_route = string

                  # The Service the route sends traffic to. Required.
                  service = string

                  # The Deployment behind the Service. Required.
                  deployment = string

                  # How long to wait for route updates to propagate, in seconds format
                  # (e.g. "60s"), at most 3 hours. Empty means no wait.
                  route_update_wait_time = optional(string, "")

                  # How long to migrate traffic back from the canary Service to the
                  # original one during the stable phase, "15s" to "3600s". Empty means
                  # no cutback time.
                  stable_cutback_duration = optional(string, "")

                  # The Pod label selecting the Deployment's and Service's Pods; it must
                  # already be on both. Empty uses Cloud Deploy's default selection.
                  pod_selector_label = optional(string, "")

                  # Deploy the HTTPRoute to additional clusters as well (multi-cluster
                  # service mesh). Empty deploys it to the target cluster only.
                  route_destinations = optional(object({
                    # The clusters: associated-entity IDs of the target (its
                    # associated_entities), and "@self" for the target cluster itself.
                    # Required.
                    destination_ids = list(string)

                    # Also deploy the Service to the destination clusters, so DNS resolves
                    # there.
                    propagate_service = optional(bool, false)
                  }))
                }))

                # Split traffic by Pod count behind a Kubernetes Service.
                service_networking = optional(object({
                  # The Service whose Pods are split. Required.
                  service = string

                  # The Deployment behind the Service. Required.
                  deployment = string

                  # Limit the total Pods used by the canary to the Deployment's current
                  # count instead of adding canary Pods on top.
                  disable_pod_overprovisioning = optional(bool, false)

                  # The Pod label selecting the Deployment's Pods; it must already be on
                  # the Deployment. Empty uses Cloud Deploy's default selection.
                  pod_selector_label = optional(string, "")
                }))
              }))
            }))
          }))
        }))

        # Deploy parameters for this stage's target: values substituted into
        # the rendered manifests wherever they reference a parameter, each set
        # optionally limited to child targets with matching labels.
        deploy_parameters = optional(list(object({
          # The parameters as key/value pairs, e.g. {"replicas": "3"}. Required.
          values = optional(map(string), {})

          # Apply these values only to targets carrying all of these labels (for a
          # multi-target, its matching child targets). Empty applies them to every
          # target of the stage, child targets included.
          match_target_labels = optional(map(string), {})
        })), [])
      })), [])
    }))

    # The pipeline's automations, each keyed by its automation_id. At most
    # 250 rules in total across a pipeline's automations (Google's limit).
    automations = optional(list(object({
      # The automation's ID, unique in the pipeline. Required. Immutable.
      automation_id = string

      # A description shown in the console, up to 255 characters.
      description = optional(string, "")

      # Labels on the automation (keys and values at most 63 characters). The
      # platform attribution labels are added on top and win on key conflicts.
      labels = optional(map(string), {})

      # Annotations on the automation (AIP-128 key/value metadata Cloud Deploy
      # never reads).
      annotations = optional(map(string), {})

      # Deactivate the automation; its rules stop firing until resumed.
      suspended = optional(bool, false)

      # The user-managed service account the automation creates releases and
      # rollouts as: its email, a GcpServiceAccount reference (its email
      # output) or a literal. The identity running the deploy needs
      # iam.serviceAccounts.actAs on it (roles/iam.serviceAccountUser), and
      # the account itself needs roles/clouddeploy.operator (or narrower
      # release and rollout permissions) plus actAs on the targets' execution
      # service accounts. Required.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      service_account = string

      # Which of the pipeline's targets the automation acts on. Required.
      selector = object({
        # The targets, at least one; a target matches when it matches any entry.
        targets = list(object({
          # A target's bare ID -- a GcpDeployTarget reference (its target_id
          # output) or a literal -- or "*" for every target in the pipeline's
          # region.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          id = optional(string, "")

          # Match targets carrying these labels. Empty keeps whatever Google
          # records.
          labels = optional(map(string), {})
        }))
      })

      # The automation's rules, at least one. Their order is not their order
      # of execution.
      rules = list(object({
        # Advance a successful canary rollout to its next phase.
        advance_rollout_rule = optional(object({
          # The rule's ID, unique in the automation: 1-63 lowercase letters,
          # digits, and hyphens, starting with a letter and not ending with a
          # hyphen. Required.
          id = string

          # Advance only from these phase IDs (e.g. "canary-25"). Empty means any
          # phase.
          source_phases = optional(list(string), [])

          # How long to wait after the phase finishes, e.g. "600s". Empty means
          # no wait.
          wait = optional(string, "")
        }))

        # Promote a release that succeeded on a selected target to the next
        # (or a named) stage.
        promote_release_rule = optional(object({
          # The rule's ID, unique in the automation (same format as every rule
          # ID). Required.
          id = string

          # How long the release waits before it is promoted, e.g. "3600s". Empty
          # means promote at once.
          wait = optional(string, "")

          # The stage to promote to: a stage target's bare ID (a GcpDeployTarget
          # reference or a literal), or "@next" for the next stage. Empty means
          # the next stage.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          destination_target_id = optional(string, "")

          # The phase the promoted rollout starts in. Empty means the first phase.
          destination_phase = optional(string, "")
        }))

        # Retry or roll back a failed rollout.
        repair_rollout_rule = optional(object({
          # The rule's ID, unique in the automation (same format as every rule
          # ID). Required.
          id = string

          # Repair failed jobs only in these phases. Empty means every phase.
          phases = optional(list(string), [])

          # Repair only these jobs (e.g. "deploy", "verify"), in the phases above.
          # Empty means every job.
          jobs = optional(list(string), [])

          # The repair steps, in order: retries, then a rollback. Google requires
          # at least one.
          repair_phases = list(object({
            # Retry the failed job.
            retry = optional(object({
              # How many retries, "1" to "10" (a string, as the provider takes it).
              # Required.
              attempts = string

              # How long to wait before the first retry, e.g. "60s", at most 14 days.
              # Empty means no wait.
              wait = optional(string, "")

              # How the wait grows between retries:
              #   "" / "BACKOFF_MODE_LINEAR" -- linear (Google's default)
              #   "BACKOFF_MODE_EXPONENTIAL" -- doubles each time
              # Ignored when wait is empty.
              backoff_mode = optional(string, "")
            }))

            # Roll the target back to its last successful release. An empty block
            # ({}) rolls back into the stable phase.
            rollback = optional(object({
              # The phase the rollback rollout starts in. Empty means the stable
              # phase.
              destination_phase = optional(string, "")

              # Abort the rollback when a rollout is already pending on the target.
              disable_rollback_if_rollout_pending = optional(bool, false)
            }))
          }))
        }))

        # Promote on a schedule (e.g. every Monday at 09:00).
        timed_promote_release_rule = optional(object({
          # The rule's ID, unique in the automation (same format as every rule
          # ID). Required.
          id = string

          # When to promote, in crontab format, e.g. "0 9 * * 1" for Mondays at
          # 09:00. Required.
          schedule = string

          # The schedule's IANA time zone, e.g. "America/New_York". Required.
          time_zone = string

          # The stage to promote to: a stage target's bare ID (a GcpDeployTarget
          # reference or a literal), or "@next". Empty means the next stage.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          destination_target_id = optional(string, "")

          # The phase the promoted rollout starts in. Empty means the first phase.
          destination_phase = optional(string, "")
        }))
      }))
    })), [])

    # What destroy does, for the pipeline and its automations:
    #   "" / "DELETE" -- deleted (the pipeline with force=true, taking its
    #                    releases and rollouts with it)
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- everything leaves management and stays in Google
    deletion_policy = optional(string, "")
  })
}

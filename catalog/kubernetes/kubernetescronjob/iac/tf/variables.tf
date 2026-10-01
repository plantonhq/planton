variable "metadata" {
  description = "Cloud resource metadata"
  type = object({
    name = string
    id = optional(string, "")
    org = optional(string, "")
    env = optional(string, "")
    labels = optional(map(string), {})
    annotations = optional(map(string), {})
    tags = optional(list(string), [])
  })
}

variable "spec" {
  description = "KubernetesCronJob specification"
  type = object({
    namespace = string
    create_namespace = optional(bool, false)
    schedule = string
    time_zone = optional(string)
    starting_deadline_seconds = optional(number)
    concurrency_policy = optional(string)
    suspend = optional(bool, false)
    successful_jobs_history_limit = optional(number)
    failed_jobs_history_limit = optional(number)
    job_template = object({
      container = object({
        app = object({
          name = optional(string, "")
          image = object({
            repo = optional(string, "")
            tag = optional(string, "")
            digest = optional(string, "")
          })
          image_pull_policy = optional(string, "")
          command = optional(list(string), [])
          args = optional(list(string), [])
          working_dir = optional(string, "")
          ports = optional(list(object({
            name = string
            container_port = number
            network_protocol = optional(string, "")
            app_protocol = optional(string, "")
            service_port = optional(number, 0)
            host_port = optional(number, 0)
          })), [])
          env = optional(object({
            variables = optional(list(object({
              name = string
              value = optional(string, "")
              config_map_key_ref = optional(object({
                name = string
                key = string
                optional = optional(bool, false)
              }))
              field_ref = optional(object({
                api_version = optional(string, "")
                field_path = string
              }))
              resource_field_ref = optional(object({
                container_name = optional(string, "")
                resource = string
                divisor = optional(string, "")
              }))
            })), [])
            secrets = optional(list(object({
              name = string
              value = optional(string, "")
              secret_ref = optional(object({
                namespace = optional(string, "")
                name = string
                key = string
                optional = optional(bool, false)
              }))
            })), [])
            env_from = optional(list(object({
              prefix = optional(string, "")
              config_map_ref = optional(object({
                name = string
                optional = optional(bool, false)
              }))
              secret_ref = optional(object({
                name = string
                optional = optional(bool, false)
              }))
            })), [])
          }))
          resources = optional(object({
            limits = optional(object({
              cpu = optional(string, "")
              memory = optional(string, "")
            }))
            requests = optional(object({
              cpu = optional(string, "")
              memory = optional(string, "")
            }))
          }))
          liveness_probe = optional(object({
            initial_delay_seconds = optional(number, 0)
            period_seconds = optional(number, 0)
            timeout_seconds = optional(number, 0)
            success_threshold = optional(number, 0)
            failure_threshold = optional(number, 0)
            http_get = optional(object({
              path = optional(string, "")
              port_number = optional(number, 0)
              port_name = optional(string, "")
              host = optional(string, "")
              scheme = optional(string, "")
              http_headers = optional(list(object({
                name = optional(string, "")
                value = optional(string, "")
              })), [])
            }))
            grpc = optional(object({
              port = optional(number, 0)
              service = optional(string, "")
            }))
            tcp_socket = optional(object({
              port_number = optional(number, 0)
              port_name = optional(string, "")
              host = optional(string, "")
            }))
            exec = optional(object({
              command = optional(list(string), [])
            }))
          }))
          readiness_probe = optional(object({
            initial_delay_seconds = optional(number, 0)
            period_seconds = optional(number, 0)
            timeout_seconds = optional(number, 0)
            success_threshold = optional(number, 0)
            failure_threshold = optional(number, 0)
            http_get = optional(object({
              path = optional(string, "")
              port_number = optional(number, 0)
              port_name = optional(string, "")
              host = optional(string, "")
              scheme = optional(string, "")
              http_headers = optional(list(object({
                name = optional(string, "")
                value = optional(string, "")
              })), [])
            }))
            grpc = optional(object({
              port = optional(number, 0)
              service = optional(string, "")
            }))
            tcp_socket = optional(object({
              port_number = optional(number, 0)
              port_name = optional(string, "")
              host = optional(string, "")
            }))
            exec = optional(object({
              command = optional(list(string), [])
            }))
          }))
          startup_probe = optional(object({
            initial_delay_seconds = optional(number, 0)
            period_seconds = optional(number, 0)
            timeout_seconds = optional(number, 0)
            success_threshold = optional(number, 0)
            failure_threshold = optional(number, 0)
            http_get = optional(object({
              path = optional(string, "")
              port_number = optional(number, 0)
              port_name = optional(string, "")
              host = optional(string, "")
              scheme = optional(string, "")
              http_headers = optional(list(object({
                name = optional(string, "")
                value = optional(string, "")
              })), [])
            }))
            grpc = optional(object({
              port = optional(number, 0)
              service = optional(string, "")
            }))
            tcp_socket = optional(object({
              port_number = optional(number, 0)
              port_name = optional(string, "")
              host = optional(string, "")
            }))
            exec = optional(object({
              command = optional(list(string), [])
            }))
          }))
          volume_mounts = optional(list(object({
            name = string
            mount_path = string
            read_only = optional(bool, false)
            sub_path = optional(string, "")
            config_map = optional(object({
              name = string
              key = optional(string, "")
              path = optional(string, "")
              default_mode = optional(number, 0)
            }))
            secret = optional(object({
              name = string
              key = optional(string, "")
              path = optional(string, "")
              default_mode = optional(number, 0)
            }))
            host_path = optional(object({
              path = string
              type = optional(string, "")
            }))
            empty_dir = optional(object({
              medium = optional(string, "")
              size_limit = optional(string, "")
            }))
            pvc = optional(object({
              claim_name = string
              read_only = optional(bool, false)
            }))
            service_account_token = optional(object({
              audience = string
              expiration_seconds = optional(number, 0)
              path = optional(string, "")
            }))
          })), [])
          lifecycle = optional(object({
            post_start = optional(object({
              exec = optional(object({
                command = optional(list(string), [])
              }))
              http_get = optional(object({
                path = optional(string, "")
                port_number = optional(number, 0)
                port_name = optional(string, "")
                host = optional(string, "")
                scheme = optional(string, "")
                http_headers = optional(list(object({
                  name = optional(string, "")
                  value = optional(string, "")
                })), [])
              }))
              tcp_socket = optional(object({
                port_number = optional(number, 0)
                port_name = optional(string, "")
                host = optional(string, "")
              }))
              sleep = optional(object({
                seconds = optional(number, 0)
              }))
            }))
            pre_stop = optional(object({
              exec = optional(object({
                command = optional(list(string), [])
              }))
              http_get = optional(object({
                path = optional(string, "")
                port_number = optional(number, 0)
                port_name = optional(string, "")
                host = optional(string, "")
                scheme = optional(string, "")
                http_headers = optional(list(object({
                  name = optional(string, "")
                  value = optional(string, "")
                })), [])
              }))
              tcp_socket = optional(object({
                port_number = optional(number, 0)
                port_name = optional(string, "")
                host = optional(string, "")
              }))
              sleep = optional(object({
                seconds = optional(number, 0)
              }))
            }))
          }))
          security_context = optional(object({
            privileged = optional(bool, false)
            run_as_user = optional(number)
            run_as_group = optional(number)
            run_as_non_root = optional(bool)
            read_only_root_filesystem = optional(bool)
            allow_privilege_escalation = optional(bool)
            capabilities = optional(object({
              add = optional(list(string), [])
              drop = optional(list(string), [])
            }))
            seccomp_profile = optional(object({
              type = string
              localhost_profile = optional(string, "")
            }))
          }))
        })
        sidecars = optional(list(object({
          name = optional(string, "")
          image = object({
            repo = optional(string, "")
            tag = optional(string, "")
            digest = optional(string, "")
          })
          image_pull_policy = optional(string, "")
          command = optional(list(string), [])
          args = optional(list(string), [])
          working_dir = optional(string, "")
          ports = optional(list(object({
            name = string
            container_port = number
            network_protocol = optional(string, "")
            app_protocol = optional(string, "")
            service_port = optional(number, 0)
            host_port = optional(number, 0)
          })), [])
          env = optional(object({
            variables = optional(list(object({
              name = string
              value = optional(string, "")
              config_map_key_ref = optional(object({
                name = string
                key = string
                optional = optional(bool, false)
              }))
              field_ref = optional(object({
                api_version = optional(string, "")
                field_path = string
              }))
              resource_field_ref = optional(object({
                container_name = optional(string, "")
                resource = string
                divisor = optional(string, "")
              }))
            })), [])
            secrets = optional(list(object({
              name = string
              value = optional(string, "")
              secret_ref = optional(object({
                namespace = optional(string, "")
                name = string
                key = string
                optional = optional(bool, false)
              }))
            })), [])
            env_from = optional(list(object({
              prefix = optional(string, "")
              config_map_ref = optional(object({
                name = string
                optional = optional(bool, false)
              }))
              secret_ref = optional(object({
                name = string
                optional = optional(bool, false)
              }))
            })), [])
          }))
          resources = optional(object({
            limits = optional(object({
              cpu = optional(string, "")
              memory = optional(string, "")
            }))
            requests = optional(object({
              cpu = optional(string, "")
              memory = optional(string, "")
            }))
          }))
          liveness_probe = optional(object({
            initial_delay_seconds = optional(number, 0)
            period_seconds = optional(number, 0)
            timeout_seconds = optional(number, 0)
            success_threshold = optional(number, 0)
            failure_threshold = optional(number, 0)
            http_get = optional(object({
              path = optional(string, "")
              port_number = optional(number, 0)
              port_name = optional(string, "")
              host = optional(string, "")
              scheme = optional(string, "")
              http_headers = optional(list(object({
                name = optional(string, "")
                value = optional(string, "")
              })), [])
            }))
            grpc = optional(object({
              port = optional(number, 0)
              service = optional(string, "")
            }))
            tcp_socket = optional(object({
              port_number = optional(number, 0)
              port_name = optional(string, "")
              host = optional(string, "")
            }))
            exec = optional(object({
              command = optional(list(string), [])
            }))
          }))
          readiness_probe = optional(object({
            initial_delay_seconds = optional(number, 0)
            period_seconds = optional(number, 0)
            timeout_seconds = optional(number, 0)
            success_threshold = optional(number, 0)
            failure_threshold = optional(number, 0)
            http_get = optional(object({
              path = optional(string, "")
              port_number = optional(number, 0)
              port_name = optional(string, "")
              host = optional(string, "")
              scheme = optional(string, "")
              http_headers = optional(list(object({
                name = optional(string, "")
                value = optional(string, "")
              })), [])
            }))
            grpc = optional(object({
              port = optional(number, 0)
              service = optional(string, "")
            }))
            tcp_socket = optional(object({
              port_number = optional(number, 0)
              port_name = optional(string, "")
              host = optional(string, "")
            }))
            exec = optional(object({
              command = optional(list(string), [])
            }))
          }))
          startup_probe = optional(object({
            initial_delay_seconds = optional(number, 0)
            period_seconds = optional(number, 0)
            timeout_seconds = optional(number, 0)
            success_threshold = optional(number, 0)
            failure_threshold = optional(number, 0)
            http_get = optional(object({
              path = optional(string, "")
              port_number = optional(number, 0)
              port_name = optional(string, "")
              host = optional(string, "")
              scheme = optional(string, "")
              http_headers = optional(list(object({
                name = optional(string, "")
                value = optional(string, "")
              })), [])
            }))
            grpc = optional(object({
              port = optional(number, 0)
              service = optional(string, "")
            }))
            tcp_socket = optional(object({
              port_number = optional(number, 0)
              port_name = optional(string, "")
              host = optional(string, "")
            }))
            exec = optional(object({
              command = optional(list(string), [])
            }))
          }))
          volume_mounts = optional(list(object({
            name = string
            mount_path = string
            read_only = optional(bool, false)
            sub_path = optional(string, "")
            config_map = optional(object({
              name = string
              key = optional(string, "")
              path = optional(string, "")
              default_mode = optional(number, 0)
            }))
            secret = optional(object({
              name = string
              key = optional(string, "")
              path = optional(string, "")
              default_mode = optional(number, 0)
            }))
            host_path = optional(object({
              path = string
              type = optional(string, "")
            }))
            empty_dir = optional(object({
              medium = optional(string, "")
              size_limit = optional(string, "")
            }))
            pvc = optional(object({
              claim_name = string
              read_only = optional(bool, false)
            }))
            service_account_token = optional(object({
              audience = string
              expiration_seconds = optional(number, 0)
              path = optional(string, "")
            }))
          })), [])
          lifecycle = optional(object({
            post_start = optional(object({
              exec = optional(object({
                command = optional(list(string), [])
              }))
              http_get = optional(object({
                path = optional(string, "")
                port_number = optional(number, 0)
                port_name = optional(string, "")
                host = optional(string, "")
                scheme = optional(string, "")
                http_headers = optional(list(object({
                  name = optional(string, "")
                  value = optional(string, "")
                })), [])
              }))
              tcp_socket = optional(object({
                port_number = optional(number, 0)
                port_name = optional(string, "")
                host = optional(string, "")
              }))
              sleep = optional(object({
                seconds = optional(number, 0)
              }))
            }))
            pre_stop = optional(object({
              exec = optional(object({
                command = optional(list(string), [])
              }))
              http_get = optional(object({
                path = optional(string, "")
                port_number = optional(number, 0)
                port_name = optional(string, "")
                host = optional(string, "")
                scheme = optional(string, "")
                http_headers = optional(list(object({
                  name = optional(string, "")
                  value = optional(string, "")
                })), [])
              }))
              tcp_socket = optional(object({
                port_number = optional(number, 0)
                port_name = optional(string, "")
                host = optional(string, "")
              }))
              sleep = optional(object({
                seconds = optional(number, 0)
              }))
            }))
          }))
          security_context = optional(object({
            privileged = optional(bool, false)
            run_as_user = optional(number)
            run_as_group = optional(number)
            run_as_non_root = optional(bool)
            read_only_root_filesystem = optional(bool)
            allow_privilege_escalation = optional(bool)
            capabilities = optional(object({
              add = optional(list(string), [])
              drop = optional(list(string), [])
            }))
            seccomp_profile = optional(object({
              type = string
              localhost_profile = optional(string, "")
            }))
          }))
        })), [])
      })
      pod = optional(object({
        service_account = optional(string, "")
        automount_service_account_token = optional(bool)
        image_pull_secrets = optional(list(string), [])
        image_registries = optional(list(object({
          server = string
          username = string
          password = string
          email = optional(string, "")
        })), [])
        init_containers = optional(list(object({
          name = optional(string, "")
          image = object({
            repo = optional(string, "")
            tag = optional(string, "")
            digest = optional(string, "")
          })
          image_pull_policy = optional(string, "")
          command = optional(list(string), [])
          args = optional(list(string), [])
          working_dir = optional(string, "")
          ports = optional(list(object({
            name = string
            container_port = number
            network_protocol = optional(string, "")
            app_protocol = optional(string, "")
            service_port = optional(number, 0)
            host_port = optional(number, 0)
          })), [])
          env = optional(object({
            variables = optional(list(object({
              name = string
              value = optional(string, "")
              config_map_key_ref = optional(object({
                name = string
                key = string
                optional = optional(bool, false)
              }))
              field_ref = optional(object({
                api_version = optional(string, "")
                field_path = string
              }))
              resource_field_ref = optional(object({
                container_name = optional(string, "")
                resource = string
                divisor = optional(string, "")
              }))
            })), [])
            secrets = optional(list(object({
              name = string
              value = optional(string, "")
              secret_ref = optional(object({
                namespace = optional(string, "")
                name = string
                key = string
                optional = optional(bool, false)
              }))
            })), [])
            env_from = optional(list(object({
              prefix = optional(string, "")
              config_map_ref = optional(object({
                name = string
                optional = optional(bool, false)
              }))
              secret_ref = optional(object({
                name = string
                optional = optional(bool, false)
              }))
            })), [])
          }))
          resources = optional(object({
            limits = optional(object({
              cpu = optional(string, "")
              memory = optional(string, "")
            }))
            requests = optional(object({
              cpu = optional(string, "")
              memory = optional(string, "")
            }))
          }))
          liveness_probe = optional(object({
            initial_delay_seconds = optional(number, 0)
            period_seconds = optional(number, 0)
            timeout_seconds = optional(number, 0)
            success_threshold = optional(number, 0)
            failure_threshold = optional(number, 0)
            http_get = optional(object({
              path = optional(string, "")
              port_number = optional(number, 0)
              port_name = optional(string, "")
              host = optional(string, "")
              scheme = optional(string, "")
              http_headers = optional(list(object({
                name = optional(string, "")
                value = optional(string, "")
              })), [])
            }))
            grpc = optional(object({
              port = optional(number, 0)
              service = optional(string, "")
            }))
            tcp_socket = optional(object({
              port_number = optional(number, 0)
              port_name = optional(string, "")
              host = optional(string, "")
            }))
            exec = optional(object({
              command = optional(list(string), [])
            }))
          }))
          readiness_probe = optional(object({
            initial_delay_seconds = optional(number, 0)
            period_seconds = optional(number, 0)
            timeout_seconds = optional(number, 0)
            success_threshold = optional(number, 0)
            failure_threshold = optional(number, 0)
            http_get = optional(object({
              path = optional(string, "")
              port_number = optional(number, 0)
              port_name = optional(string, "")
              host = optional(string, "")
              scheme = optional(string, "")
              http_headers = optional(list(object({
                name = optional(string, "")
                value = optional(string, "")
              })), [])
            }))
            grpc = optional(object({
              port = optional(number, 0)
              service = optional(string, "")
            }))
            tcp_socket = optional(object({
              port_number = optional(number, 0)
              port_name = optional(string, "")
              host = optional(string, "")
            }))
            exec = optional(object({
              command = optional(list(string), [])
            }))
          }))
          startup_probe = optional(object({
            initial_delay_seconds = optional(number, 0)
            period_seconds = optional(number, 0)
            timeout_seconds = optional(number, 0)
            success_threshold = optional(number, 0)
            failure_threshold = optional(number, 0)
            http_get = optional(object({
              path = optional(string, "")
              port_number = optional(number, 0)
              port_name = optional(string, "")
              host = optional(string, "")
              scheme = optional(string, "")
              http_headers = optional(list(object({
                name = optional(string, "")
                value = optional(string, "")
              })), [])
            }))
            grpc = optional(object({
              port = optional(number, 0)
              service = optional(string, "")
            }))
            tcp_socket = optional(object({
              port_number = optional(number, 0)
              port_name = optional(string, "")
              host = optional(string, "")
            }))
            exec = optional(object({
              command = optional(list(string), [])
            }))
          }))
          volume_mounts = optional(list(object({
            name = string
            mount_path = string
            read_only = optional(bool, false)
            sub_path = optional(string, "")
            config_map = optional(object({
              name = string
              key = optional(string, "")
              path = optional(string, "")
              default_mode = optional(number, 0)
            }))
            secret = optional(object({
              name = string
              key = optional(string, "")
              path = optional(string, "")
              default_mode = optional(number, 0)
            }))
            host_path = optional(object({
              path = string
              type = optional(string, "")
            }))
            empty_dir = optional(object({
              medium = optional(string, "")
              size_limit = optional(string, "")
            }))
            pvc = optional(object({
              claim_name = string
              read_only = optional(bool, false)
            }))
            service_account_token = optional(object({
              audience = string
              expiration_seconds = optional(number, 0)
              path = optional(string, "")
            }))
          })), [])
          lifecycle = optional(object({
            post_start = optional(object({
              exec = optional(object({
                command = optional(list(string), [])
              }))
              http_get = optional(object({
                path = optional(string, "")
                port_number = optional(number, 0)
                port_name = optional(string, "")
                host = optional(string, "")
                scheme = optional(string, "")
                http_headers = optional(list(object({
                  name = optional(string, "")
                  value = optional(string, "")
                })), [])
              }))
              tcp_socket = optional(object({
                port_number = optional(number, 0)
                port_name = optional(string, "")
                host = optional(string, "")
              }))
              sleep = optional(object({
                seconds = optional(number, 0)
              }))
            }))
            pre_stop = optional(object({
              exec = optional(object({
                command = optional(list(string), [])
              }))
              http_get = optional(object({
                path = optional(string, "")
                port_number = optional(number, 0)
                port_name = optional(string, "")
                host = optional(string, "")
                scheme = optional(string, "")
                http_headers = optional(list(object({
                  name = optional(string, "")
                  value = optional(string, "")
                })), [])
              }))
              tcp_socket = optional(object({
                port_number = optional(number, 0)
                port_name = optional(string, "")
                host = optional(string, "")
              }))
              sleep = optional(object({
                seconds = optional(number, 0)
              }))
            }))
          }))
          security_context = optional(object({
            privileged = optional(bool, false)
            run_as_user = optional(number)
            run_as_group = optional(number)
            run_as_non_root = optional(bool)
            read_only_root_filesystem = optional(bool)
            allow_privilege_escalation = optional(bool)
            capabilities = optional(object({
              add = optional(list(string), [])
              drop = optional(list(string), [])
            }))
            seccomp_profile = optional(object({
              type = string
              localhost_profile = optional(string, "")
            }))
          }))
        })), [])
        labels = optional(map(string), {})
        annotations = optional(map(string), {})
        scheduling = optional(object({
          node_selector = optional(map(string), {})
          tolerations = optional(list(object({
            key = optional(string, "")
            operator = optional(string, "")
            value = optional(string, "")
            effect = optional(string, "")
            toleration_seconds = optional(number)
          })), [])
          node_affinity = optional(object({
            required = optional(list(object({
              match_expressions = list(object({
                key = string
                operator = string
                values = optional(list(string), [])
              }))
            })), [])
            preferred = optional(list(object({
              weight = optional(number, 0)
              term = object({
                match_expressions = list(object({
                  key = string
                  operator = string
                  values = optional(list(string), [])
                }))
              })
            })), [])
          }))
          pod_affinity = optional(object({
            required = optional(list(object({
              match_labels = optional(map(string), {})
              topology_key = string
              namespaces = optional(list(string), [])
            })), [])
            preferred = optional(list(object({
              weight = optional(number, 0)
              term = object({
                match_labels = optional(map(string), {})
                topology_key = string
                namespaces = optional(list(string), [])
              })
            })), [])
          }))
          pod_anti_affinity = optional(object({
            required = optional(list(object({
              match_labels = optional(map(string), {})
              topology_key = string
              namespaces = optional(list(string), [])
            })), [])
            preferred = optional(list(object({
              weight = optional(number, 0)
              term = object({
                match_labels = optional(map(string), {})
                topology_key = string
                namespaces = optional(list(string), [])
              })
            })), [])
          }))
          topology_spread_constraints = optional(list(object({
            max_skew = optional(number, 0)
            topology_key = string
            when_unsatisfiable = string
            match_labels = optional(map(string), {})
          })), [])
          scheduler_name = optional(string, "")
        }))
        security_context = optional(object({
          run_as_user = optional(number)
          run_as_group = optional(number)
          run_as_non_root = optional(bool)
          fs_group = optional(number)
          fs_group_change_policy = optional(string, "")
          supplemental_groups = optional(list(number), [])
          sysctls = optional(list(object({
            name = string
            value = string
          })), [])
          seccomp_profile = optional(object({
            type = string
            localhost_profile = optional(string, "")
          }))
        }))
        termination_grace_period_seconds = optional(number)
        dns_policy = optional(string, "")
        dns_config = optional(object({
          nameservers = optional(list(string), [])
          searches = optional(list(string), [])
          options = optional(list(object({
            name = string
            value = optional(string, "")
          })), [])
        }))
        host_aliases = optional(list(object({
          ip = string
          hostnames = list(string)
        })), [])
        host_network = optional(bool, false)
        host_pid = optional(bool, false)
        priority_class_name = optional(string, "")
        runtime_class_name = optional(string, "")
      }))
      parallelism = optional(number)
      completions = optional(number)
      completion_mode = optional(string)
      backoff_limit = optional(number)
      backoff_limit_per_index = optional(number)
      max_failed_indexes = optional(number)
      active_deadline_seconds = optional(number)
      ttl_seconds_after_finished = optional(number)
      restart_policy = optional(string)
      pod_failure_policy = optional(object({
        rules = list(object({
          action = string
          on_exit_codes = optional(object({
            container_name = optional(string, "")
            operator = string
            values = list(number)
          }))
          on_pod_conditions = optional(list(object({
            type = string
            status = optional(string)
          })), [])
        }))
      }))
      success_policy = optional(object({
        rules = list(object({
          succeeded_indexes = optional(string, "")
          succeeded_count = optional(number)
        }))
      }))
    })
  })
}

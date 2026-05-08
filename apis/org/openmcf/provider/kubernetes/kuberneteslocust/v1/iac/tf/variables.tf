variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name = string,
    id = optional(string),
    org = optional(string),
    env = optional(string),
    labels = optional(map(string)),
    tags = optional(list(string)),
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "Specification for Kubernetes Locust deployment"
  type = object({
    # Target Kubernetes cluster (orchestrator-only, not included in generated tfvars)
    target_cluster_name = optional(string)

    # Kubernetes namespace for Locust deployment
    namespace = string

    # Flag to indicate if the namespace should be created
    create_namespace = optional(bool, true)

    # The master container specifications for the Locust cluster.
    # This defines the resource allocation and number of replicas for the master node.
    master_container = object({

      # The number of replicas for the container.
      # This determines the level of concurrency and load generation capabilities.
      replicas = number

      # The CPU and memory resources allocated to the Locust container.
      resources = object({

        # The resource limits for the container.
        # Specify the maximum amount of CPU and memory that the container can use.
        limits = object({

          # The amount of CPU allocated (e.g., "500m" for 0.5 CPU cores).
          cpu = string

          # The amount of memory allocated (e.g., "256Mi" for 256 mebibytes).
          memory = string
        })

        # The resource requests for the container.
        # Specify the minimum amount of CPU and memory that the container is guaranteed.
        requests = object({

          # The amount of CPU allocated (e.g., "500m" for 0.5 CPU cores).
          cpu = string

          # The amount of memory allocated (e.g., "256Mi" for 256 mebibytes).
          memory = string
        })
      })
    })

    # The worker container specifications for the Locust cluster.
    # This defines the resource allocation and number of replicas for the worker nodes.
    worker_container = object({

      # The number of replicas for the container.
      # This determines the level of concurrency and load generation capabilities.
      replicas = number

      # The CPU and memory resources allocated to the Locust container.
      resources = object({

        # The resource limits for the container.
        # Specify the maximum amount of CPU and memory that the container can use.
        limits = object({

          # The amount of CPU allocated (e.g., "500m" for 0.5 CPU cores).
          cpu = string

          # The amount of memory allocated (e.g., "256Mi" for 256 mebibytes).
          memory = string
        })

        # The resource requests for the container.
        # Specify the minimum amount of CPU and memory that the container is guaranteed.
        requests = object({

          # The amount of CPU allocated (e.g., "500m" for 0.5 CPU cores).
          cpu = string

          # The amount of memory allocated (e.g., "256Mi" for 256 mebibytes).
          memory = string
        })
      })
    })

    # The ingress configuration for the Locust deployment.
    ingress = optional(object({

      # A flag to enable or disable ingress.
      is_enabled = optional(bool, false)

      # The dns domain.
      dns_domain = optional(string, "")
    }))

    # The load test parameters, including the main test script, additional library files,
    # and extra Python pip packages needed for test execution.
    # This specifies how the Locust nodes will simulate traffic and interact with the target application.
    load_test = object({

      # A unique identifier or name for this particular load test specification.
      # It is used to reference or distinguish this test configuration among others within a testing suite or environment.
      name = string

      # The Python code for the main Locust test script.
      # This script defines the behavior of the simulated users and is crucial for executing the load test.
      main_py_content = string

      # A map where each entry consists of a filename and its associated Python code content.
      # These files typically contain additional classes or functions required by the main_py_content script.
      # The key of the map is the filename, and the value is the file content.
      lib_files_content = optional(map(string))

      # A list of extra Python pip packages that are required for the load test.
      # These packages will be installed in the environment where the load test is executed,
      # allowing for extended functionality or custom dependencies to be included easily.
      pip_packages = optional(list(string))
    })

    # A map of key-value pairs providing additional customization options for the Helm chart used
    # to deploy the Locust cluster. These values allow for further refinement of the deployment,
    # such as customizing resource limits, setting environment variables, or specifying version tags.
    # For detailed information on the available options, refer to the Helm chart documentation at:
    # https://github.com/deliveryhero/helm-charts/tree/master/stable/locust#values
    helm_values = optional(map(string))
  })
}

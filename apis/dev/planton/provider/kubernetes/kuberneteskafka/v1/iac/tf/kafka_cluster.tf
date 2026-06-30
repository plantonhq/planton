# This resource represents the core Strimzi Kafka custom resource.
# It will create both the Kafka and Zookeeper clusters (via the CRD).
# For simplicity, this example only demonstrates how to handle
# the internal listener plus optional external listeners when ingress is enabled.
resource "kubernetes_manifest" "kafka_cluster" {
  manifest = {
    apiVersion = "kafka.strimzi.io/v1beta2"
    kind       = "Kafka"
    metadata = {
      name      = local.resource_id
      namespace = local.namespace
      labels    = local.final_labels
    }
    spec = {
      # Entity Operator handles topic/user creation & configuration
      entityOperator = {
        topicOperator = {}
        userOperator  = {}
      }

      kafka = {
        # Configure simple authorization with an 'admin' superuser
        authorization = {
          type       = "simple"
          superUsers = [local.admin_username]
        }

        # Example default configuration. You can override or extend this.
        config = {
          "offsets.topic.replication.factor"         = 1
          "transaction.state.log.replication.factor" = 1
          "transaction.state.log.min.isr"            = 1
          "auto.create.topics.enable"                = true
        }

        # Combine internal & external (private/public) listeners
        listeners = concat(
          [
            # Internal listener
            {
              name = "int"
              port = 9094
              tls  = false
              type = "internal"
              authentication = {
                type = "scram-sha-512"
              }
            }
          ],
          local.ingress_is_enabled && local.ingress_dns_domain != "" ? [
            # Private external listener
            {
              name = "extpvt"
              port = 9093
              tls  = true
              type = "loadbalancer"
              authentication = {
                type = "scram-sha-512"
              }
              # Additional optional configuration can go here
              # configuration = {
              #   # advanced broker config, annotations, etc.
              # }
            },
            # Public external listener
            {
              name = "extpub"
              port = 9092
              tls  = true
              type = "loadbalancer"
              authentication = {
                type = "scram-sha-512"
              }
              # Additional optional configuration can go here
              # configuration = {
              #   # advanced broker config, annotations, etc.
              # }
            }
          ] : []
        )

        # Number of Kafka broker replicas
        replicas = local.broker_replicas

        # CPU/memory resources for Kafka brokers
        resources = {
          limits = {
            cpu    = try(var.spec.broker_container.resources.limits.cpu, "1000m")
            memory = try(var.spec.broker_container.resources.limits.memory, "1Gi")
          }
          requests = {
            cpu    = try(var.spec.broker_container.resources.requests.cpu, "50m")
            memory = try(var.spec.broker_container.resources.requests.memory, "100Mi")
          }
        }

        # JBOD storage configuration
        storage = {
          type = "jbod"
          volumes = [
            {
              id          = 0
              size        = try(var.spec.broker_container.disk_size, "1Gi")
              type        = "persistent-claim"
              deleteClaim = false
            }
          ]
        }
      }

      zookeeper = {
        # Number of Zookeeper replicas
        replicas = local.zookeeper_replicas

        # CPU/memory resources for Zookeeper
        resources = {
          limits = {
            cpu    = try(var.spec.zookeeper_container.resources.limits.cpu, "1000m")
            memory = try(var.spec.zookeeper_container.resources.limits.memory, "1Gi")
          }
          requests = {
            cpu    = try(var.spec.zookeeper_container.resources.requests.cpu, "50m")
            memory = try(var.spec.zookeeper_container.resources.requests.memory, "100Mi")
          }
        }

        # Persistent storage for Zookeeper
        storage = {
          type        = "persistent-claim"
          size        = try(var.spec.zookeeper_container.disk_size, "1Gi")
          deleteClaim = false
        }
      }
    }
  }

  depends_on = [
    kubernetes_namespace_v1.kafka_namespace
  ]
}

##############################################
# mongodb.tf
#
# Creates MongoDB resources using the Percona
# Server for MongoDB Operator CRD.
##############################################

# Generate random password for MongoDB root user
resource "random_password" "mongodb_root_password" {
  length           = 12
  special          = true
  numeric          = true
  upper            = true
  lower            = true
  min_special      = 3
  min_numeric      = 2
  min_upper        = 2
  min_lower        = 2
  override_special = "!@#$%^&*()-_=+[]{}:?"
}

# Create Kubernetes secret to store MongoDB password
# Percona operator expects plaintext passwords in StringData (Kubernetes auto-encodes)
resource "kubernetes_secret_v1" "mongodb_password" {
  metadata {
    name      = local.password_secret_name
    namespace = local.namespace
    labels    = local.final_labels
  }

  data = {
    "MONGODB_DATABASE_ADMIN_PASSWORD" = base64encode(random_password.mongodb_root_password.result)
  }
}

# Create PerconaServerMongoDB custom resource using the Percona operator
resource "kubernetes_manifest" "percona_server_mongodb" {
  manifest = {
    apiVersion = "psmdb.percona.com/v1"
    kind       = "PerconaServerMongoDB"

    metadata = {
      name      = var.metadata.name
      namespace = local.namespace
      labels    = local.final_labels
    }

    spec = {
      # CRD version for the Percona operator
      crVersion = "1.20.1"

      # MongoDB image version
      image = "percona/percona-server-mongodb:8.0.12-4"

      # Replica set configuration
      replsets = [
        {
          name = "rs0"
          size = var.spec.container.replicas

          # Configure container resources
          resources = {
            requests = {
              cpu    = var.spec.container.resources.requests.cpu
              memory = var.spec.container.resources.requests.memory
            }
            limits = {
              cpu    = var.spec.container.resources.limits.cpu
              memory = var.spec.container.resources.limits.memory
            }
          }

          # Configure persistence if enabled
          volumeSpec = var.spec.container.persistence_enabled ? {
            persistentVolumeClaim = {
              resources = {
                requests = {
                  storage = var.spec.container.disk_size
                }
              }
            }
          } : null
        }
      ]

      # Reference to the secret containing MongoDB credentials
      secrets = {
        users = kubernetes_secret_v1.mongodb_password.metadata[0].name
      }

      # Allow replica sets with less than 3 members for dev/test environments
      unsafeFlags = {
        replsetSize = true
      }
    }
  }

  depends_on = [
    kubernetes_secret_v1.mongodb_password
  ]
}

# Create LoadBalancer service for external access if ingress is enabled
resource "kubernetes_service_v1" "mongodb_external_lb" {
  count = local.ingress_is_enabled ? 1 : 0

  metadata {
    name      = local.external_lb_service_name
    namespace = local.namespace
    labels    = local.final_labels

    annotations = {
      "external-dns.alpha.kubernetes.io/hostname" = local.ingress_external_hostname
    }
  }

  spec {
    type = "LoadBalancer"

    port {
      name        = "tcp-mongodb"
      port        = 27017
      protocol    = "TCP"
      target_port = "mongodb"
    }

    selector = {
      "app.kubernetes.io/name"       = "percona-server-mongodb"
      "app.kubernetes.io/instance"   = var.metadata.name
      "app.kubernetes.io/managed-by" = "percona-server-mongodb-operator"
    }
  }

  depends_on = [
    kubernetes_manifest.percona_server_mongodb
  ]
}

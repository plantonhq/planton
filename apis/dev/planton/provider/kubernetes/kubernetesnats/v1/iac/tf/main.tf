##############################################
# main.tf
#
# Creates NATS resources on Kubernetes using
# the official NATS Helm chart.
#
# Deployment order (per ChatGPT guidance for avoiding race conditions):
# 1. NATS Helm release
# 2. NACK CRDs (explicit step, not via Helm)
# 3. NACK controller Helm release
# 4. Stream/Consumer custom resources
##############################################

# Create namespace for NATS deployment (conditionally based on create_namespace flag)
resource "kubernetes_namespace" "nats_namespace" {
  count = var.spec.create_namespace ? 1 : 0

  metadata {
    name   = local.namespace
    labels = local.labels
  }
}

# Generate random password for bearer token authentication
resource "random_password" "nats_bearer_token" {
  count = try(var.spec.auth.enabled, false) && try(var.spec.auth.scheme, "") == "bearer_token" ? 1 : 0

  length  = 32
  special = false
}

# Generate random password for basic auth
resource "random_password" "nats_admin_password" {
  count = try(var.spec.auth.enabled, false) && try(var.spec.auth.scheme, "") == "basic_auth" ? 1 : 0

  length  = 32
  special = false
}

# Create secret for bearer token authentication
resource "kubernetes_secret_v1" "nats_bearer_token_secret" {
  count = try(var.spec.auth.enabled, false) && try(var.spec.auth.scheme, "") == "bearer_token" ? 1 : 0

  metadata {
    name      = local.auth_secret_name
    namespace = local.namespace
    labels    = local.labels
  }

  data = {
    "token" = base64encode(random_password.nats_bearer_token[0].result)
  }

  type = "Opaque"

  depends_on = [kubernetes_namespace.nats_namespace]
}

# Create secret for basic auth admin credentials
resource "kubernetes_secret_v1" "nats_admin_secret" {
  count = try(var.spec.auth.enabled, false) && try(var.spec.auth.scheme, "") == "basic_auth" ? 1 : 0

  metadata {
    name      = local.auth_secret_name
    namespace = local.namespace
    labels    = local.labels
  }

  data = {
    "user"     = base64encode("nats")
    "password" = base64encode(random_password.nats_admin_password[0].result)
  }

  type = "Opaque"

  depends_on = [kubernetes_namespace.nats_namespace]
}

# Create secret for no-auth user (if enabled with basic auth)
resource "kubernetes_secret_v1" "nats_noauth_secret" {
  count = try(var.spec.auth.enabled, false) && try(var.spec.auth.scheme, "") == "basic_auth" && try(var.spec.auth.no_auth_user.enabled, false) ? 1 : 0

  metadata {
    name      = local.no_auth_user_secret_name
    namespace = local.namespace
    labels    = local.labels
  }

  data = {
    "user"     = base64encode("noauth")
    "password" = base64encode("nopassword")
  }

  type = "Opaque"

  depends_on = [kubernetes_namespace.nats_namespace]
}

# Generate self-signed TLS certificate if TLS is enabled
resource "tls_private_key" "nats_tls" {
  count = try(var.spec.tls_enabled, false) ? 1 : 0

  algorithm = "RSA"
  rsa_bits  = 2048
}

resource "tls_self_signed_cert" "nats_tls" {
  count = try(var.spec.tls_enabled, false) ? 1 : 0

  private_key_pem = tls_private_key.nats_tls[0].private_key_pem

  subject {
    common_name  = "${var.metadata.name}.${local.namespace}.svc.cluster.local"
    organization = "Planton"
  }

  validity_period_hours = 8760 # 1 year

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
    "client_auth",
  ]

  dns_names = [
    "${var.metadata.name}",
    "${var.metadata.name}.${local.namespace}",
    "${var.metadata.name}.${local.namespace}.svc",
    "${var.metadata.name}.${local.namespace}.svc.cluster.local",
  ]
}

# Create TLS secret
resource "kubernetes_secret_v1" "nats_tls_secret" {
  count = try(var.spec.tls_enabled, false) ? 1 : 0

  metadata {
    name      = local.tls_secret_name
    namespace = local.namespace
    labels    = local.labels
  }

  data = {
    "tls.crt" = base64encode(tls_self_signed_cert.nats_tls[0].cert_pem)
    "tls.key" = base64encode(tls_private_key.nats_tls[0].private_key_pem)
  }

  type = "kubernetes.io/tls"

  depends_on = [kubernetes_namespace.nats_namespace]
}

##############################################
# Step 1: Deploy NATS Helm chart
##############################################

resource "helm_release" "nats" {
  name       = var.metadata.name
  repository = "https://nats-io.github.io/k8s/helm/charts"
  chart      = "nats"
  version    = local.nats_helm_chart_version
  namespace  = local.namespace

  # Wait for resources to be ready
  wait    = true
  timeout = 600 # 10 minutes

  # Helm values are split into multiple yamlencode entries to avoid HCL's
  # "Inconsistent conditional result types" error. Helm deep-merges values
  # entries in order (like multiple -f flags), so each section can be
  # independently guarded without needing matching object shapes.
  values = concat(
    # Base configuration (always present, uniform shape)
    [yamlencode({
      container = {
        merge = {
          resources = {
            limits = {
              cpu    = var.spec.server_container.resources.limits.cpu
              memory = var.spec.server_container.resources.limits.memory
            }
            requests = {
              cpu    = var.spec.server_container.resources.requests.cpu
              memory = var.spec.server_container.resources.requests.memory
            }
          }
        }
      }
      config = {
        cluster = {
          enabled  = var.spec.server_container.replicas > 1
          replicas = var.spec.server_container.replicas
        }
      }
      natsbox = {
        enabled = !try(var.spec.disable_nats_box, false)
      }
    })],

    # JetStream: both branches produce string via yamlencode, so the
    # conditional is string-vs-string (always valid in HCL).
    [var.spec.disable_jet_stream
      ? yamlencode({ config = { jetstream = { enabled = false } } })
      : yamlencode({ config = { jetstream = {
        enabled   = true
        fileStore = { enabled = true, pvc = { size = var.spec.server_container.disk_size } }
      } } })
    ],

    # Bearer token auth (conditional list: list(string) vs empty list)
    try(var.spec.auth.enabled, false) && try(var.spec.auth.scheme, "") == "bearer_token" ? [yamlencode({
      auth = {
        enabled = true
        token = {
          users = [
            {
              existingSecret = {
                name = kubernetes_secret_v1.nats_bearer_token_secret[0].metadata[0].name
                key  = "token"
              }
            }
          ]
        }
      }
    })] : [],

    # Basic auth: config patch + auth flag (conditional list)
    try(var.spec.auth.enabled, false) && try(var.spec.auth.scheme, "") == "basic_auth" ? [yamlencode({
      config = {
        patch = concat(
          [
            {
              op   = "add"
              path = "/authorization"
              value = {
                users = concat(
                  [
                    {
                      username = "nats"
                      password = random_password.nats_admin_password[0].result
                    }
                  ],
                  try(var.spec.auth.no_auth_user.enabled, false) ? [
                    {
                      username = "noauth"
                      password = "nopassword"
                      permissions = {
                        publish   = try(var.spec.auth.no_auth_user.publish_subjects, [])
                        subscribe = []
                      }
                    }
                  ] : []
                )
              }
            }
          ],
          try(var.spec.auth.no_auth_user.enabled, false) ? [
            {
              op    = "add"
              path  = "/no_auth_user"
              value = "noauth"
            }
          ] : []
        )
      }
      auth = {
        enabled = true
        basic   = {}
      }
    })] : [],

    # TLS (conditional list)
    try(var.spec.tls_enabled, false) ? [yamlencode({
      tls = {
        enabled = true
        secret = {
          name = kubernetes_secret_v1.nats_tls_secret[0].metadata[0].name
        }
      }
    })] : []
  )

  depends_on = [
    kubernetes_namespace.nats_namespace,
    kubernetes_secret_v1.nats_bearer_token_secret,
    kubernetes_secret_v1.nats_admin_secret,
    kubernetes_secret_v1.nats_noauth_secret,
    kubernetes_secret_v1.nats_tls_secret
  ]
}

# Create LoadBalancer service for external access if ingress is enabled
resource "kubernetes_service_v1" "nats_external_lb" {
  count = try(var.spec.ingress.enabled, false) ? 1 : 0

  metadata {
    name      = local.external_lb_service_name
    namespace = local.namespace
    labels    = local.labels

    annotations = try(var.spec.ingress.hostname, "") != "" ? {
      "external-dns.alpha.kubernetes.io/hostname" = var.spec.ingress.hostname
    } : {}
  }

  spec {
    type = "LoadBalancer"

    port {
      name        = "client"
      port        = local.nats_client_port
      protocol    = "TCP"
      target_port = local.nats_client_port
    }

    selector = {
      "app.kubernetes.io/name"      = "nats"
      "app.kubernetes.io/component" = "nats"
      "app.kubernetes.io/instance"  = var.metadata.name
    }
  }

  depends_on = [
    helm_release.nats
  ]
}

##############################################
# Step 2: Deploy NACK CRDs
# CRDs are deployed as an explicit step (not via Helm chart) to avoid race
# conditions and preview/dry-run issues. This follows Helm's best practices
# for CRD management.
##############################################

# Fetch NACK CRDs from GitHub
data "http" "nack_crds" {
  count = local.nack_controller_enabled ? 1 : 0

  url = local.nack_crds_url
}

# Apply NACK CRDs using kubectl_manifest or kubernetes_manifest
# We use kubernetes_manifest with for_each over the decoded YAML documents
resource "kubernetes_manifest" "nack_crds" {
  for_each = local.nack_controller_enabled ? {
    for idx, doc in [
      for d in split("---", data.http.nack_crds[0].response_body) :
      yamldecode(d) if trimspace(d) != "" && can(yamldecode(d))
    ] : "${doc.metadata.name}" => doc
  } : {}

  manifest = each.value

  depends_on = [helm_release.nats]
}

##############################################
# Step 3: Deploy NACK Controller Helm release
# The NACK controller watches JetStream CRDs and reconciles them
# to actual JetStream resources in the NATS cluster.
##############################################

resource "helm_release" "nack_controller" {
  count = local.nack_controller_enabled ? 1 : 0

  name       = "${var.metadata.name}-nack"
  repository = local.nack_helm_chart_repo_url
  chart      = "nack"
  version    = local.nack_helm_chart_version
  namespace  = local.namespace

  # Wait for resources to be ready
  wait    = true
  timeout = 300 # 5 minutes

  # Skip CRDs - we install them separately for better control
  skip_crds = true

  values = [
    yamlencode({
      jetstream = merge(
        {
          enabled = true
          nats = {
            url = (
              try(var.spec.auth.enabled, false) && try(var.spec.auth.scheme, "") == "basic_auth"
              ? "nats://${local.admin_username}:${random_password.nats_admin_password[0].result}@${local.nats_service_name}.${local.namespace}.svc.cluster.local:${local.nats_client_port}"
              : local.internal_client_url
            )
          }
        },
        # Add control-loop mode if enabled
        local.nack_enable_control_loop ? {
          additionalArgs = ["--control-loop"]
        } : {}
      )
    })
  ]

  depends_on = [
    helm_release.nats,
    kubernetes_manifest.nack_crds
  ]
}

##############################################
# Step 4: Create Stream/Consumer Custom Resources
# These resources are reconciled by the NACK controller to create actual
# JetStream streams and consumers in the NATS cluster.
##############################################

# Create JetStream Stream custom resources
resource "kubernetes_manifest" "nack_streams" {
  for_each = local.nack_controller_enabled ? {
    for stream in local.streams : stream.name => stream
  } : {}

  manifest = {
    apiVersion = "jetstream.nats.io/v1beta2"
    kind       = "Stream"
    metadata = {
      name      = lower(each.value.name)
      namespace = local.namespace
      labels    = local.labels
    }
    spec = merge(
      {
        name     = each.value.name
        subjects = each.value.subjects
      },
      # Storage type
      try(each.value.storage, "") != "" ? {
        storage = each.value.storage
      } : {},
      # Replicas
      try(each.value.replicas, 0) > 0 ? {
        replicas = each.value.replicas
      } : {},
      # Retention policy
      try(each.value.retention, "") != "" ? {
        retention = each.value.retention
      } : {},
      # Max age
      try(each.value.max_age, "") != "" ? {
        maxAge = each.value.max_age
      } : {},
      # Max bytes (-1 is unlimited, omit if default)
      try(each.value.max_bytes, -1) != -1 ? {
        maxBytes = each.value.max_bytes
      } : {},
      # Max messages
      try(each.value.max_msgs, -1) != -1 ? {
        maxMsgs = each.value.max_msgs
      } : {},
      # Max message size
      try(each.value.max_msg_size, -1) != -1 ? {
        maxMsgSize = each.value.max_msg_size
      } : {},
      # Max consumers
      try(each.value.max_consumers, -1) != -1 ? {
        maxConsumers = each.value.max_consumers
      } : {},
      # Discard policy
      # Note: Convert "new_msgs" to "new" because "new" is a reserved keyword in Java.
      # Proto uses "new_msgs" but NACK CRDs expect "new".
      try(each.value.discard, "") != "" && try(each.value.discard, "") != "old" ? {
        discard = each.value.discard == "new_msgs" ? "new" : each.value.discard
      } : {},
      # Description
      try(each.value.description, "") != "" ? {
        description = each.value.description
      } : {}
    )
  }

  depends_on = [
    helm_release.nack_controller
  ]
}

# Flatten consumers from all streams for creating consumer resources
locals {
  all_consumers = local.nack_controller_enabled ? flatten([
    for stream in local.streams : [
      for consumer in try(stream.consumers, []) : {
        stream_name  = stream.name
        consumer     = consumer
        resource_key = "${stream.name}-${consumer.durable_name}"
      }
    ]
  ]) : []
}

# Create JetStream Consumer custom resources
resource "kubernetes_manifest" "nack_consumers" {
  for_each = {
    for item in local.all_consumers : item.resource_key => item
  }

  manifest = {
    apiVersion = "jetstream.nats.io/v1beta2"
    kind       = "Consumer"
    metadata = {
      name      = lower(each.value.consumer.durable_name)
      namespace = local.namespace
      labels    = local.labels
    }
    spec = merge(
      {
        streamName  = each.value.stream_name
        durableName = each.value.consumer.durable_name
      },
      # Deliver policy
      # Note: Convert "new_msgs" to "new" because "new" is a reserved keyword in Java.
      # Proto uses "new_msgs" but NACK CRDs expect "new".
      try(each.value.consumer.deliver_policy, "") != "" && try(each.value.consumer.deliver_policy, "") != "all" ? {
        deliverPolicy = each.value.consumer.deliver_policy == "new_msgs" ? "new" : each.value.consumer.deliver_policy
      } : {},
      # Ack policy
      try(each.value.consumer.ack_policy, "") != "" && try(each.value.consumer.ack_policy, "") != "none" ? {
        ackPolicy = each.value.consumer.ack_policy
      } : {},
      # Filter subject
      try(each.value.consumer.filter_subject, "") != "" ? {
        filterSubject = each.value.consumer.filter_subject
      } : {},
      # Deliver subject (for push consumers)
      try(each.value.consumer.deliver_subject, "") != "" ? {
        deliverSubject = each.value.consumer.deliver_subject
      } : {},
      # Deliver group (queue group)
      try(each.value.consumer.deliver_group, "") != "" ? {
        deliverGroup = each.value.consumer.deliver_group
      } : {},
      # Max ack pending
      try(each.value.consumer.max_ack_pending, 0) > 0 ? {
        maxAckPending = each.value.consumer.max_ack_pending
      } : {},
      # Max deliver
      try(each.value.consumer.max_deliver, -1) != -1 ? {
        maxDeliver = each.value.consumer.max_deliver
      } : {},
      # Ack wait
      try(each.value.consumer.ack_wait, "") != "" ? {
        ackWait = each.value.consumer.ack_wait
      } : {},
      # Replay policy
      try(each.value.consumer.replay_policy, "") != "" && try(each.value.consumer.replay_policy, "") != "instant" ? {
        replayPolicy = each.value.consumer.replay_policy
      } : {},
      # Description
      try(each.value.consumer.description, "") != "" ? {
        description = each.value.consumer.description
      } : {}
    )
  }

  depends_on = [
    kubernetes_manifest.nack_streams
  ]
}

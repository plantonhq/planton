# KubernetesOpenBao Terraform Examples

## Example 1: Basic Standalone Deployment

```hcl
module "openbao_standalone" {
  source = "."

  metadata = {
    name = "dev-openbao"
  }

  spec = {
    namespace        = "openbao"
    create_namespace = true

    server_container = {
      replicas          = 1
      data_storage_size = "10Gi"
      resources = {
        limits = {
          cpu    = "500m"
          memory = "256Mi"
        }
        requests = {
          cpu    = "100m"
          memory = "128Mi"
        }
      }
    }

    ui_enabled = true
  }
}
```

## Example 2: High Availability Deployment

```hcl
module "openbao_ha" {
  source = "."

  metadata = {
    name = "prod-openbao"
    org  = "my-org"
    env  = "production"
  }

  spec = {
    namespace        = "openbao-prod"
    create_namespace = true

    server_container = {
      replicas          = 3
      data_storage_size = "50Gi"
      resources = {
        limits = {
          cpu    = "1000m"
          memory = "512Mi"
        }
        requests = {
          cpu    = "250m"
          memory = "256Mi"
        }
      }
    }

    high_availability = {
      enabled  = true
      replicas = 3
    }

    ui_enabled = true
  }
}
```

## Example 3: Full Production Setup

```hcl
module "openbao_enterprise" {
  source = "."

  metadata = {
    name = "enterprise-openbao"
    org  = "my-org"
    env  = "production"
  }

  spec = {
    namespace        = "openbao-enterprise"
    create_namespace = true

    server_container = {
      replicas          = 5
      data_storage_size = "100Gi"
      resources = {
        limits = {
          cpu    = "2000m"
          memory = "1Gi"
        }
        requests = {
          cpu    = "500m"
          memory = "512Mi"
        }
      }
    }

    high_availability = {
      enabled  = true
      replicas = 5
    }

    ingress = {
      enabled            = true
      hostname           = "secrets.company.com"
      ingress_class_name = "nginx"
      tls_enabled        = true
      tls_secret_name    = "openbao-tls"
    }

    injector = {
      enabled  = true
      replicas = 2
    }

    ui_enabled  = true
    tls_enabled = true
  }
}
```

## Example 4: HA with GCP KMS Auto-Unseal

```hcl
module "openbao_auto_unseal" {
  source = "."

  metadata = {
    name = "prod-openbao"
    org  = "my-org"
    env  = "production"
  }

  spec = {
    namespace        = "openbao-prod"
    create_namespace = true

    server_container = {
      replicas          = 3
      data_storage_size = "50Gi"
      resources = {
        limits = {
          cpu    = "1000m"
          memory = "512Mi"
        }
        requests = {
          cpu    = "250m"
          memory = "256Mi"
        }
      }
    }

    high_availability = {
      enabled  = true
      replicas = 3
    }

    auto_unseal = {
      gcp_kms = {
        project                           = "my-gcp-project"
        region                            = "asia-south1"
        key_ring                          = "openbao-unseal"
        crypto_key                        = "openbao-auto-unseal-key"
        workload_identity_service_account = "openbao@my-gcp-project.iam.gserviceaccount.com"
      }
    }

    ui_enabled  = true
    tls_enabled = true
  }
}
```

## Example 5: Minimal Development Setup

```hcl
module "openbao_dev" {
  source = "."

  metadata = {
    name = "local-openbao"
  }

  spec = {
    namespace        = "openbao-dev"
    create_namespace = true

    server_container = {
      replicas          = 1
      data_storage_size = "1Gi"
      resources = {
        limits = {
          cpu    = "200m"
          memory = "128Mi"
        }
        requests = {
          cpu    = "50m"
          memory = "64Mi"
        }
      }
    }

    ui_enabled = true
  }
}
```

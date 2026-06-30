# KubernetesGatewayApiCrds Terraform Module

This Terraform module installs Kubernetes Gateway API CRDs on any Kubernetes cluster.

## Overview

The module fetches and applies the official Gateway API CRD manifests from the [kubernetes-sigs/gateway-api](https://github.com/kubernetes-sigs/gateway-api) releases to the target cluster.

## Prerequisites

- Terraform >= 1.0
- Kubernetes cluster access
- kubectl provider configured

## Providers

| Name | Version |
|------|---------|
| kubernetes | >= 2.0 |
| http | >= 3.0 |
| kubectl | >= 2.0 |

## Usage

### Basic Usage

```hcl
module "gateway_api_crds" {
  source = "./path/to/module"

  metadata = {
    name = "gateway-api-crds"
  }

  spec = {
    version = "v1.2.1"
    install_channel = {
      channel = "standard"
    }
  }
}
```

### Experimental Channel

```hcl
module "gateway_api_crds" {
  source = "./path/to/module"

  metadata = {
    name = "gateway-api-crds-experimental"
  }

  spec = {
    version = "v1.2.1"
    install_channel = {
      channel = "experimental"
    }
  }
}
```

### With Provider Configuration

```hcl
provider "kubernetes" {
  config_path = "~/.kube/config"
}

provider "kubectl" {
  config_path = "~/.kube/config"
}

module "gateway_api_crds" {
  source = "./path/to/module"

  metadata = {
    name = "gateway-api-crds"
  }
}
```

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| metadata | Resource metadata | `object({ name = string })` | n/a | yes |
| spec | KubernetesGatewayApiCrds specification | `object` | See below | no |

### Spec Object

| Field | Description | Default |
|-------|-------------|---------|
| version | Gateway API version | `"v1.2.1"` |
| install_channel.channel | `standard` or `experimental` | `"standard"` |

## Outputs

| Name | Description |
|------|-------------|
| installed_version | Gateway API version that was installed |
| installed_channel | Installation channel (standard/experimental) |
| installed_manifest_url | Full URL of the CRD bundle that was applied |

## Installed CRDs

### Standard Channel

- `gatewayclasses.gateway.networking.k8s.io`
- `gateways.gateway.networking.k8s.io`
- `httproutes.gateway.networking.k8s.io`
- `grpcroutes.gateway.networking.k8s.io`
- `tlsroutes.gateway.networking.k8s.io`
- `referencegrants.gateway.networking.k8s.io`

### Experimental Channel (includes standard)

- All standard CRDs, plus:
- `tcproutes.gateway.networking.k8s.io`
- `udproutes.gateway.networking.k8s.io`

## Troubleshooting

### CRDs Not Installing

1. Verify Kubernetes cluster access
2. Check that the version exists in Gateway API releases
3. Ensure network access to GitHub

### Permission Denied

Ensure the Terraform service account has cluster-admin or equivalent permissions to create CRDs.

### Provider Not Configured

Ensure both `kubernetes` and `kubectl` providers are configured with valid cluster credentials.

## References

- [Gateway API Releases](https://github.com/kubernetes-sigs/gateway-api/releases)
- [Gateway API Documentation](https://gateway-api.sigs.k8s.io/)
- [Terraform Kubernetes Provider](https://registry.terraform.io/providers/hashicorp/kubernetes/latest)
- [kubectl Terraform Provider](https://registry.terraform.io/providers/alekc/kubectl/latest)

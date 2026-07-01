# Azure Web App Environment

Production web application hosting on Azure App Service with monitoring.

## What This Chart Deploys

| Resource | Kind | Condition |
|----------|------|-----------|
| Resource Group | `AzureResourceGroup` | Always |
| Service Plan | `AzureServicePlan` | Always |
| Log Analytics Workspace | `AzureLogAnalyticsWorkspace` | Always |
| Application Insights | `AzureApplicationInsights` | `create_app_insights` |
| Linux Web App | `AzureLinuxWebApp` | Always |

## Architecture

The chart creates a Linux Web App on Azure App Service:

- **Service Plan** defines the compute tier. Defaults to P1v3 (Premium v3)
  for production workloads with auto-healing and deployment slots. Use B1
  for development or S1 for standard workloads.
- **Log Analytics Workspace** collects platform-level logs
- **Application Insights** (optional, default on) injects the connection
  string via `APPLICATIONINSIGHTS_CONNECTION_STRING` app setting for automatic
  request tracing and dependency tracking
- **Linux Web App** hosts the application with the selected runtime stack

## Runtime Stacks

| `runtime_stack` | Description | Example `runtime_version` |
|-----------------|-------------|---------------------------|
| `node` | Node.js | `20`, `18`, `16` |
| `python` | Python | `3.12`, `3.11`, `3.10` |
| `java` | Java | `17`, `11` |
| `dotnet` | .NET | `8.0`, `6.0` |
| `php` | PHP | `8.3`, `8.2` |
| `ruby` | Ruby | `3.2`, `3.1` |
| `go` | Go | `1.21`, `1.20` |
| `docker` | Custom Docker image | Set `docker_image` and `docker_image_tag` |

## Parameters

### Foundation

| Parameter | Description | Default |
|-----------|-------------|---------|
| `region` | Azure region | `eastus` |
| `resource_group_name` | Resource group suffix | `webapp-rg` |

### Service Plan

| Parameter | Description | Default |
|-----------|-------------|---------|
| `plan_name` | Plan name | `webapp-plan` |
| `sku_name` | SKU (B1-B3, S1-S3, P1v3-P3v3) | `P1v3` |

### Web App

| Parameter | Description | Default |
|-----------|-------------|---------|
| `web_app_name` | App name (globally unique) | `my-web-app` |
| `runtime_stack` | Runtime language | `node` |
| `runtime_version` | Language version | `20` |
| `docker_image` | Docker image (docker stack only) | (empty) |
| `docker_image_tag` | Docker tag (docker stack only) | `latest` |

### Monitoring

| Parameter | Description | Default |
|-----------|-------------|---------|
| `create_app_insights` | Enable APM | `true` |

## Example

Deploy a Python web API on Premium tier:

```yaml
params:
  region: westeurope
  resource_group_name: myapi-rg
  plan_name: myapi-plan
  sku_name: P1v3
  web_app_name: myapi-web
  runtime_stack: python
  runtime_version: "3.12"
```

Deploy a custom Docker container:

```yaml
params:
  region: eastus
  web_app_name: myapp-web
  runtime_stack: docker
  docker_image: myregistry.azurecr.io/myapp
  docker_image_tag: v1.2.3
  sku_name: P1v3
```

# AzureContainerAppEnvironmentDaprComponent - Terraform Module

Terraform implementation for the AzureContainerAppEnvironmentDaprComponent
kind.

## Resources Created

- `azurerm_container_app_environment_dapr_component.main` -- the
  pluggable Dapr backend (state store, pub/sub, secret store, or
  binding) registered on the environment

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.container_app_environment_id` | The owning environment (ForceNew) |
| `spec.component_name` / `spec.component_type` | The name apps pass to the Dapr API and the backend type (e.g. `state.azure.blobstorage`); both ForceNew |
| `spec.metadata` | Component configuration entries; each carries a literal `value` XOR a `secret_name` reference (CEL-enforced) |
| `spec.secrets` | Named secrets that metadata entries reference -- secret material never rides plain metadata |
| `spec.scopes` | The Dapr app IDs allowed to use the component; EMPTY means every app in the environment -- scope production components deliberately |

## Provider Version

`azurerm ~> 5.0`.

## Behavior Notes

- `ignore_errors` defaults to false so a broken component fails loudly
  at sidecar startup instead of silently degrading.
- `init_timeout` defaults to `5s`, sent explicitly for engine parity.
- No tags: ARM does not support tags on `daprComponents`.

## Usage

```hcl
module "dapr_state" {
  source = "./path/to/module"

  metadata = { name = "orders-state" }
  spec = {
    container_app_environment_id = "/subscriptions/.../managedEnvironments/apps-env"
    component_name               = "orders-state"
    component_type               = "state.azure.blobstorage"
    version                      = "v1"
    secrets = [{ name = "storage-key", value = var.storage_key }]
    metadata = [
      { name = "accountName", value = "ordersstore" },
      { name = "accountKey", secret_name = "storage-key" },
    ]
    scopes = ["orders-api"]
  }
}
```

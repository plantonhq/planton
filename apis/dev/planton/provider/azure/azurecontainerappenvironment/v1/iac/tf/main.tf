# Create the Azure Container App Environment.
resource "azurerm_container_app_environment" "main" {
  name                           = var.spec.name
  location                       = var.spec.region
  resource_group_name            = var.spec.resource_group
  infrastructure_subnet_id       = var.spec.infrastructure_subnet_id
  log_analytics_workspace_id     = var.spec.log_analytics_workspace_id
  logs_destination               = local.logs_destination
  internal_load_balancer_enabled = var.spec.internal_load_balancer_enabled
  zone_redundancy_enabled        = var.spec.zone_redundancy_enabled
  tags                           = local.final_tags

  dynamic "workload_profile" {
    for_each = var.spec.workload_profiles
    content {
      name                  = workload_profile.value.name
      workload_profile_type = workload_profile.value.workload_profile_type
      minimum_count         = workload_profile.value.minimum_count
      maximum_count         = workload_profile.value.maximum_count
    }
  }
}

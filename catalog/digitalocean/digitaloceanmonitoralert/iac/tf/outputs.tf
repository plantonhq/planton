# Outputs — exactly the DigitalOceanMonitorAlertOutputs
# contract, identical across both provisioners. The resource id IS the
# policy UUID; the provider's own uuid attribute is declared but never
# populated at the pinned version.

output "alert_id" {
  description = "UUID of the alert policy (the API identity, and the import id)"
  value       = digitalocean_monitor_alert.alert.id
}

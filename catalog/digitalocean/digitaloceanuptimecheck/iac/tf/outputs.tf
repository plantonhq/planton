# Stack outputs — exactly the DigitalOceanUptimeCheckStackOutputs contract,
# identical across both provisioners.

output "check_id" {
  description = "UUID of the uptime check (the API identity, and the import id)"
  value       = digitalocean_uptime_check.check.id
}

output "alert_ids" {
  description = "UUIDs of the composed alert rows keyed by the module's for_each key (<row index>-<alert name>) -- the second half of each row's {check_id},{alert_id} import id"
  value       = { for key, alert in digitalocean_uptime_alert.alerts : key => alert.id }
}

# Outputs — flattened onto KubernetesPlantonPlatformOutputs by
# the platform. Keep in lockstep with the Pulumi module's exports. All
# values derive from the declaration itself (the operator's naming is
# deterministic per platform name), so they are stable from the first
# apply.

output "namespace" {
  description = "Namespace the platform lives in"
  value       = local.namespace
}

output "platform_name" {
  description = "The PlantonPlatform CR name — the prefix of every object the operator creates for this platform"
  value       = local.platform_name
}

output "gateway_service" {
  description = "The built-in front-door gateway Service — console, API, and sign-in on one origin"
  value       = local.gateway_service
}

output "setup_code_secret" {
  description = "The Secret holding the first-run setup code the console's setup page asks for"
  value       = local.setup_code_secret
}

output "port_forward_command" {
  description = "The exact command that opens the platform's door on this machine"
  value       = local.port_forward_command
}

output "setup_code_command" {
  description = "The exact command that reads the first-run setup code"
  value       = local.setup_code_command
}

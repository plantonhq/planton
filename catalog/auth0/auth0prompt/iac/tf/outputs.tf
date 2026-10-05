# Auth0Prompt Outputs
# Maps to the Auth0PromptOutputs protobuf message: the login-flow settings
# as the tenant carries them after the apply, managed or not.

output "universal_login_experience" {
  description = "The login experience the tenant runs: new or classic"
  value       = auth0_prompt.this.universal_login_experience
}

output "identifier_first" {
  description = "Whether the login flow asks for the identifier first"
  value       = auth0_prompt.this.identifier_first
}

output "webauthn_platform_first_factor" {
  description = "Whether the device's own authenticator is offered as the first factor"
  value       = auth0_prompt.this.webauthn_platform_first_factor
}

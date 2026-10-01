# Auth0EmailTemplate Outputs
# Maps to the Auth0EmailTemplateStackOutputs protobuf message: the template
# managed.

output "template" {
  description = "The email this resource customizes (verify_email, reset_email, ...)"
  value       = auth0_email_template.this.template
}

output "enabled" {
  description = "Whether the template is on"
  value       = auth0_email_template.this.enabled
}

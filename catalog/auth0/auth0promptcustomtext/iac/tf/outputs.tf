# Auth0PromptCustomText Outputs
# Maps to the Auth0PromptCustomTextStackOutputs protobuf message: the custom
# text managed and its identifier.

output "prompt" {
  description = "The prompt the words belong to"
  value       = auth0_prompt_custom_text.this.prompt
}

output "language" {
  description = "The language the words are shown in"
  value       = auth0_prompt_custom_text.this.language
}

output "id" {
  description = "The custom text's identifier, <prompt>::<language>, which is also its import id"
  value       = auth0_prompt_custom_text.this.id
}

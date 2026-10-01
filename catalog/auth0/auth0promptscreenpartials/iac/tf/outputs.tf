# Auth0PromptScreenPartials Outputs
# Maps to the Auth0PromptScreenPartialsStackOutputs protobuf message: the
# prompt managed.

output "prompt_type" {
  description = "The prompt whose screens the partials extend"
  value       = auth0_prompt_screen_partials.this.prompt_type
}

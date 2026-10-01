# Auth0PromptCustomText Main Resources
#
# auth0_prompt_custom_text sets the words of one prompt in one language on the
# tenant the provider's credential belongs to. Auth0 stores one custom text per
# prompt and language and replaces it whole on every write, so this resource
# owns all of it: a screen or key the spec leaves out shows Auth0's default
# words. Destroy writes an empty document, which returns the prompt to Auth0's
# defaults in that language.
resource "auth0_prompt_custom_text" "this" {
  prompt   = var.spec.prompt
  language = var.spec.language
  body     = local.body
}

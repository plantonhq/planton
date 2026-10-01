# Local values for the Auth0PromptCustomText module.
#
# The provider takes the prompt's words as one JSON document, {"<screen>":
# {"<key>": "<text>"}}. jsonencode writes object and map keys in sorted order,
# so the document is the same bytes the Pulumi module's locals.go renders for
# the same spec -- keep them in lockstep.
locals {
  body = jsonencode({ for screen, text in var.spec.screens : screen => text.texts })
}

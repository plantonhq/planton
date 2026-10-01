# Local values for the Auth0ClientFromMetadataDocument module.
#
# Unset means unmanaged. An optional scalar the spec leaves unset arrives as
# null and is passed straight through, so the provider never sends it. A list
# or map the spec leaves empty arrives as [] or {} (the variable's defaults);
# each maps to null here, so an empty collection is never sent as "clear it".
# The Pulumi module's clientArgs applies the same rule -- keep them in lockstep.
locals {
  # grant_types is optional AND computed on the provider (the document seeds
  # it), so null keeps the document's grants. The other collections are
  # optional only: the provider resets them when Auth0 holds a value and the
  # configuration holds none, which is why their spec comments ask an adopter
  # to declare the live value.
  grant_types                    = length(var.spec.grant_types) > 0 ? var.spec.grant_types : null
  allowed_origins                = length(var.spec.allowed_origins) > 0 ? var.spec.allowed_origins : null
  web_origins                    = length(var.spec.web_origins) > 0 ? var.spec.web_origins : null
  organization_discovery_methods = length(var.spec.organization_discovery_methods) > 0 ? var.spec.organization_discovery_methods : null
  client_metadata                = length(var.spec.client_metadata) > 0 ? var.spec.client_metadata : null

  # Each declared block renders once; an undeclared one renders not at all.
  default_organization = var.spec.default_organization != null ? [var.spec.default_organization] : []
  jwt_configuration    = var.spec.jwt_configuration != null ? [var.spec.jwt_configuration] : []
  refresh_token        = var.spec.refresh_token != null ? [var.spec.refresh_token] : []
  token_quota          = var.spec.token_quota != null ? [var.spec.token_quota] : []
}

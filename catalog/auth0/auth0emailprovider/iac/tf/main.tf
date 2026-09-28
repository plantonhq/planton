# Auth0EmailProvider Main Resources
#
# auth0_email_provider is the ONE email provider of the tenant the provider's
# credential belongs to. The service arm the spec sets names it (local.name)
# and fills exactly the credentials and settings that service uses; the custom
# arm sends the empty credentials block the provider requires, and the Action
# bound to the tenant's custom-email-provider trigger does the sending. When
# the tenant already has a provider, the create takes it over and rewrites it.
# Auth0 never returns a credential, so the provider sends the credentials only
# when they change. Destroy deletes the provider, and the tenant falls back to
# Auth0's built-in test provider.
resource "auth0_email_provider" "this" {
  name                 = local.name
  enabled              = local.enabled
  default_from_address = var.spec.default_from_address

  credentials {
    smtp_host                  = local.credentials.smtp_host
    smtp_port                  = local.credentials.smtp_port
    smtp_user                  = local.credentials.smtp_user
    smtp_pass                  = local.credentials.smtp_pass
    access_key_id              = local.credentials.access_key_id
    secret_access_key          = local.credentials.secret_access_key
    api_key                    = local.credentials.api_key
    region                     = local.credentials.region
    domain                     = local.credentials.domain
    azure_cs_connection_string = local.credentials.azure_cs_connection_string
    ms365_tenant_id            = local.credentials.ms365_tenant_id
    ms365_client_id            = local.credentials.ms365_client_id
    ms365_client_secret        = local.credentials.ms365_client_secret
  }

  dynamic "settings" {
    for_each = local.settings == null ? [] : [local.settings]
    content {
      dynamic "headers" {
        for_each = settings.value.headers == null ? [] : [settings.value.headers]
        content {
          x_mc_view_content_link  = headers.value.x_mc_view_content_link
          x_ses_configuration_set = headers.value.x_ses_configuration_set
        }
      }

      dynamic "message" {
        for_each = settings.value.message == null ? [] : [settings.value.message]
        content {
          configuration_set_name = message.value.configuration_set_name
          view_content_link      = message.value.view_content_link
        }
      }
    }
  }
}

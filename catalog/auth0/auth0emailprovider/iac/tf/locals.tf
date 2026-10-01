# Local values for the Auth0EmailProvider module.
#
# The spec sets exactly one service arm, and the arm decides the provider's
# name and exactly which credentials and settings are sent -- the same rule the
# provider's own expander applies (internal/auth0/email/expand.go). A field the
# arm does not carry, or an optional one left unset, renders as null and is
# never sent. Empty strings are the proto's zero value for "unset", so each
# optional string maps "" to null here. The Pulumi module's locals.go applies
# the same rule -- keep them in lockstep.
locals {
  smtp      = try(var.spec.smtp, null)
  ses       = try(var.spec.ses, null)
  sendgrid  = try(var.spec.sendgrid, null)
  sparkpost = try(var.spec.sparkpost, null)
  mailgun   = try(var.spec.mailgun, null)
  mandrill  = try(var.spec.mandrill, null)
  azure_cs  = try(var.spec.azure_cs, null)
  ms365     = try(var.spec.ms365, null)
  custom    = try(var.spec.custom, null)

  # The service as Auth0 names it, from the arm the spec sets.
  name = (
    local.smtp != null ? "smtp" :
    local.ses != null ? "ses" :
    local.sendgrid != null ? "sendgrid" :
    local.sparkpost != null ? "sparkpost" :
    local.mailgun != null ? "mailgun" :
    local.mandrill != null ? "mandrill" :
    local.azure_cs != null ? "azure_cs" :
    local.ms365 != null ? "ms365" :
    local.custom != null ? "custom" : null
  )

  # Sending is on unless the spec turns it off.
  enabled = try(var.spec.enabled, null) == null ? true : var.spec.enabled

  # The credentials block. Every attribute the arm does not carry is null, so
  # the custom arm renders the empty block the provider requires.
  credentials = {
    smtp_host = local.smtp != null ? local.smtp.host : null
    smtp_port = local.smtp != null ? local.smtp.port : null
    smtp_user = local.smtp != null ? local.smtp.user : null
    smtp_pass = local.smtp != null ? local.smtp.password : null

    access_key_id     = local.ses != null ? local.ses.access_key_id : null
    secret_access_key = local.ses != null ? local.ses.secret_access_key : null

    api_key = (
      local.sendgrid != null ? local.sendgrid.api_key :
      local.sparkpost != null ? local.sparkpost.api_key :
      local.mailgun != null ? local.mailgun.api_key :
      local.mandrill != null ? local.mandrill.api_key : null
    )

    region = (
      local.ses != null ? local.ses.region :
      local.sparkpost != null ? (local.sparkpost.region != "" ? local.sparkpost.region : null) :
      local.mailgun != null ? (local.mailgun.region != "" ? local.mailgun.region : null) : null
    )

    domain = local.mailgun != null ? local.mailgun.domain : null

    azure_cs_connection_string = local.azure_cs != null ? local.azure_cs.connection_string : null

    ms365_tenant_id     = local.ms365 != null ? local.ms365.tenant_id : null
    ms365_client_id     = local.ms365 != null ? local.ms365.client_id : null
    ms365_client_secret = local.ms365 != null ? local.ms365.client_secret : null
  }

  # settings.headers -- the smtp arm's headers, declared only when at least one
  # is set.
  smtp_headers_spec = local.smtp != null ? try(local.smtp.headers, null) : null

  smtp_headers = (
    local.smtp_headers_spec == null ? null :
    local.smtp_headers_spec.x_mc_view_content_link == "" && local.smtp_headers_spec.x_ses_configuration_set == "" ? null :
    {
      x_mc_view_content_link  = local.smtp_headers_spec.x_mc_view_content_link != "" ? local.smtp_headers_spec.x_mc_view_content_link : null
      x_ses_configuration_set = local.smtp_headers_spec.x_ses_configuration_set != "" ? local.smtp_headers_spec.x_ses_configuration_set : null
    }
  )

  # settings.message -- the ses arm's configuration set or the mandrill arm's
  # view-content switch, declared only when set.
  message_configuration_set_name = local.ses != null ? (local.ses.configuration_set_name != "" ? local.ses.configuration_set_name : null) : null
  message_view_content_link      = local.mandrill != null ? try(local.mandrill.view_content_link, null) : null

  message = local.message_configuration_set_name == null && local.message_view_content_link == null ? null : {
    configuration_set_name = local.message_configuration_set_name
    view_content_link      = local.message_view_content_link
  }

  # The settings block, declared only when it carries headers or a message;
  # otherwise the provider keeps what Auth0 reports.
  settings = local.smtp_headers == null && local.message == null ? null : {
    headers = local.smtp_headers
    message = local.message
  }
}

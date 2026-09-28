# Local values for the Auth0Branding module.
#
# The tenant's branding: a setting left unset is NOT MANAGED. It renders as
# null (a block is not declared), so the provider never sends it and the tenant
# keeps whatever it already carries. Empty strings are the proto's zero value
# for "unset", so every field maps "" to null here.
locals {
  logo_url                 = var.spec.logo_url != "" ? var.spec.logo_url : null
  favicon_url              = var.spec.favicon_url != "" ? var.spec.favicon_url : null
  font_url                 = var.spec.font_url != "" ? var.spec.font_url : null
  universal_login_template = var.spec.universal_login_template != "" ? var.spec.universal_login_template : null

  # The colors block is declared when the spec declares colors; each color in it
  # is sent only when set.
  colors_declared        = var.spec.colors != null
  colors_primary         = try(var.spec.colors.primary, "") != "" ? var.spec.colors.primary : null
  colors_page_background = try(var.spec.colors.page_background, "") != "" ? var.spec.colors.page_background : null

  # The branding resource is declared only when the spec manages one of its
  # settings; a spec that declares only a theme leaves the branding untouched.
  manage_branding = (
    local.logo_url != null ||
    local.favicon_url != null ||
    local.colors_declared ||
    local.font_url != null ||
    local.universal_login_template != null
  )

  # The theme resource is declared only when the spec declares a theme.
  manage_theme = var.spec.theme != null
}

# The theme, once declared, is sent WHOLE: every block the provider requires
# (borders, colors, fonts with all six text styles, page_background, widget) is
# always sent, and every field the spec leaves unset is sent at Auth0's default.
# These defaults are the Default of each argument in the auth0/auth0 provider's
# internal/auth0/branding/resource_theme.go -- the one place this module states
# them. The Pulumi module's themeDefaults (iac/pulumi/module/locals.go) carries
# the same values -- keep them in lockstep.
locals {
  theme_defaults = {
    borders = {
      buttons_style        = "rounded"
      button_border_radius = 3
      button_border_weight = 1
      inputs_style         = "rounded"
      input_border_radius  = 3
      input_border_weight  = 1
      show_widget_shadow   = true
      widget_corner_radius = 5
      widget_border_weight = 0
    }
    colors = {
      base_focus_color          = "#635dff"
      base_hover_color          = "#000000"
      body_text                 = "#1e212a"
      captcha_widget_theme      = "auto"
      error                     = "#d03c38"
      header                    = "#1e212a"
      icons                     = "#65676e"
      input_background          = "#ffffff"
      input_border              = "#c9cace"
      input_filled_text         = "#000000"
      input_labels_placeholders = "#65676e"
      links_focused_components  = "#635dff"
      primary_button            = "#635dff"
      primary_button_label      = "#ffffff"
      secondary_button_border   = "#c9cace"
      secondary_button_label    = "#1e212a"
      success                   = "#13a688"
      widget_background         = "#ffffff"
      widget_border             = "#c9cace"
    }
    fonts = {
      font_url            = ""
      links_style         = "normal"
      reference_text_size = 16
    }
    # Each text style's own default: a size (a percentage of
    # reference_text_size) and whether it is bold.
    text_styles = {
      body_text    = { bold = false, size = 87.5 }
      buttons_text = { bold = false, size = 100 }
      input_labels = { bold = false, size = 100 }
      links        = { bold = true, size = 87.5 }
      subtitle     = { bold = false, size = 87.5 }
      title        = { bold = false, size = 150 }
    }
    page_background = {
      background_color     = "#000000"
      background_image_url = ""
      page_layout          = "center"
    }
    widget = {
      header_text_alignment = "center"
      logo_height           = 52
      logo_position         = "center"
      logo_url              = ""
      social_buttons_layout = "bottom"
    }
  }

  # Each block is its defaults overlaid with the fields the spec sets. A block
  # the spec leaves out (or a spec without a theme, when the resource is not
  # declared) is its defaults alone.
  theme_display_name = try(var.spec.theme.display_name, "") != "" ? var.spec.theme.display_name : null

  theme_borders = merge(
    local.theme_defaults.borders,
    try({ for k, v in var.spec.theme.borders : k => v if v != null }, {}),
  )
  theme_colors = merge(
    local.theme_defaults.colors,
    try({ for k, v in var.spec.theme.colors : k => v if v != null }, {}),
  )
  theme_fonts = merge(
    local.theme_defaults.fonts,
    try({ for k, v in var.spec.theme.fonts : k => v if v != null && contains(keys(local.theme_defaults.fonts), k) }, {}),
  )
  theme_text_styles = {
    for style, defaults in local.theme_defaults.text_styles : style => merge(
      defaults,
      try({ for k, v in var.spec.theme.fonts[style] : k => v if v != null }, {}),
    )
  }
  theme_page_background = merge(
    local.theme_defaults.page_background,
    try({ for k, v in var.spec.theme.page_background : k => v if v != null }, {}),
  )
  theme_widget = merge(
    local.theme_defaults.widget,
    try({ for k, v in var.spec.theme.widget : k => v if v != null }, {}),
  )

  # identifiers is the one theme block sent only when declared: Auth0 offers it
  # only on tenants with the feature enabled.
  theme_identifiers = try(var.spec.theme.identifiers, null)
}

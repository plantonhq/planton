# Auth0Branding Main Resources
#
# auth0_branding manages the branding of the EXISTING tenant the provider's
# credential belongs to -- the logo, favicon, colors and font every Universal
# Login page shares, and the page template the login box renders inside. It is
# declared only when the spec manages one of those settings, and each setting
# is sent only when the spec sets it. Auth0 has no delete for the tenant's
# branding: the provider's destroy removes the page template (on a tenant with
# a custom domain) and leaves the last-applied logo, favicon, colors and font in
# place. The provider checks for a custom domain on every read, update and
# delete, and refuses a page template on a tenant without one.
resource "auth0_branding" "this" {
  count = local.manage_branding ? 1 : 0

  logo_url    = local.logo_url
  favicon_url = local.favicon_url

  dynamic "colors" {
    for_each = local.colors_declared ? [1] : []
    content {
      primary         = local.colors_primary
      page_background = local.colors_page_background
    }
  }

  # Removing font_url after it was applied returns the pages to Auth0's font.
  dynamic "font" {
    for_each = local.font_url != null ? [local.font_url] : []
    content {
      url = font.value
    }
  }

  # Removing the template returns the pages to Auth0's default page.
  dynamic "universal_login" {
    for_each = local.universal_login_template != null ? [local.universal_login_template] : []
    content {
      body = universal_login.value
    }
  }
}

# auth0_branding_theme is the no-code look of the login box, declared only when
# the spec declares a theme and sent whole (locals.tf fills every unset field
# with Auth0's default). A tenant has one theme: the provider adopts the
# tenant's existing theme on create, and destroy deletes it, returning the login
# box to Auth0's look.
resource "auth0_branding_theme" "this" {
  count = local.manage_theme ? 1 : 0

  display_name = local.theme_display_name

  borders {
    buttons_style        = local.theme_borders.buttons_style
    button_border_radius = local.theme_borders.button_border_radius
    button_border_weight = local.theme_borders.button_border_weight
    inputs_style         = local.theme_borders.inputs_style
    input_border_radius  = local.theme_borders.input_border_radius
    input_border_weight  = local.theme_borders.input_border_weight
    show_widget_shadow   = local.theme_borders.show_widget_shadow
    widget_corner_radius = local.theme_borders.widget_corner_radius
    widget_border_weight = local.theme_borders.widget_border_weight
  }

  colors {
    base_focus_color          = local.theme_colors.base_focus_color
    base_hover_color          = local.theme_colors.base_hover_color
    body_text                 = local.theme_colors.body_text
    captcha_widget_theme      = local.theme_colors.captcha_widget_theme
    error                     = local.theme_colors.error
    header                    = local.theme_colors.header
    icons                     = local.theme_colors.icons
    input_background          = local.theme_colors.input_background
    input_border              = local.theme_colors.input_border
    input_filled_text         = local.theme_colors.input_filled_text
    input_labels_placeholders = local.theme_colors.input_labels_placeholders
    links_focused_components  = local.theme_colors.links_focused_components
    primary_button            = local.theme_colors.primary_button
    primary_button_label      = local.theme_colors.primary_button_label
    secondary_button_border   = local.theme_colors.secondary_button_border
    secondary_button_label    = local.theme_colors.secondary_button_label
    success                   = local.theme_colors.success
    widget_background         = local.theme_colors.widget_background
    widget_border             = local.theme_colors.widget_border
  }

  fonts {
    font_url            = local.theme_fonts.font_url
    links_style         = local.theme_fonts.links_style
    reference_text_size = local.theme_fonts.reference_text_size

    body_text {
      bold = local.theme_text_styles.body_text.bold
      size = local.theme_text_styles.body_text.size
    }
    buttons_text {
      bold = local.theme_text_styles.buttons_text.bold
      size = local.theme_text_styles.buttons_text.size
    }
    input_labels {
      bold = local.theme_text_styles.input_labels.bold
      size = local.theme_text_styles.input_labels.size
    }
    links {
      bold = local.theme_text_styles.links.bold
      size = local.theme_text_styles.links.size
    }
    subtitle {
      bold = local.theme_text_styles.subtitle.bold
      size = local.theme_text_styles.subtitle.size
    }
    title {
      bold = local.theme_text_styles.title.bold
      size = local.theme_text_styles.title.size
    }
  }

  page_background {
    background_color     = local.theme_page_background.background_color
    background_image_url = local.theme_page_background.background_image_url
    page_layout          = local.theme_page_background.page_layout
  }

  widget {
    header_text_alignment = local.theme_widget.header_text_alignment
    logo_height           = local.theme_widget.logo_height
    logo_position         = local.theme_widget.logo_position
    logo_url              = local.theme_widget.logo_url
    social_buttons_layout = local.theme_widget.social_buttons_layout
  }

  # Sent only when declared. Once applied, Auth0 keeps it: removing the block
  # leaves the last-applied values in place.
  dynamic "identifiers" {
    for_each = local.theme_identifiers != null ? [local.theme_identifiers] : []
    content {
      login_display    = identifiers.value.login_display
      otp_autocomplete = identifiers.value.otp_autocomplete
      phone_display {
        formatting = identifiers.value.phone_display.formatting
        masking    = identifiers.value.phone_display.masking
      }
    }
  }
}

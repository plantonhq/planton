variable "metadata" {
  description = "Cloud resource metadata"
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "Auth0Branding specification"
  type = object({
    logo_url    = optional(string, "")
    favicon_url = optional(string, "")
    colors = optional(object({
      primary         = optional(string, "")
      page_background = optional(string, "")
    }))
    font_url                 = optional(string, "")
    universal_login_template = optional(string, "")
    theme = optional(object({
      display_name = optional(string, "")
      borders = optional(object({
        buttons_style        = optional(string)
        button_border_radius = optional(number)
        button_border_weight = optional(number)
        inputs_style         = optional(string)
        input_border_radius  = optional(number)
        input_border_weight  = optional(number)
        show_widget_shadow   = optional(bool)
        widget_corner_radius = optional(number)
        widget_border_weight = optional(number)
      }))
      colors = optional(object({
        base_focus_color          = optional(string)
        base_hover_color          = optional(string)
        body_text                 = optional(string)
        captcha_widget_theme      = optional(string)
        error                     = optional(string)
        header                    = optional(string)
        icons                     = optional(string)
        input_background          = optional(string)
        input_border              = optional(string)
        input_filled_text         = optional(string)
        input_labels_placeholders = optional(string)
        links_focused_components  = optional(string)
        primary_button            = optional(string)
        primary_button_label      = optional(string)
        secondary_button_border   = optional(string)
        secondary_button_label    = optional(string)
        success                   = optional(string)
        widget_background         = optional(string)
        widget_border             = optional(string)
      }))
      fonts = optional(object({
        font_url            = optional(string, "")
        links_style         = optional(string)
        reference_text_size = optional(number)
        body_text = optional(object({
          bold = optional(bool)
          size = optional(number)
        }))
        buttons_text = optional(object({
          bold = optional(bool)
          size = optional(number)
        }))
        input_labels = optional(object({
          bold = optional(bool)
          size = optional(number)
        }))
        links = optional(object({
          bold = optional(bool)
          size = optional(number)
        }))
        subtitle = optional(object({
          bold = optional(bool)
          size = optional(number)
        }))
        title = optional(object({
          bold = optional(bool)
          size = optional(number)
        }))
      }))
      page_background = optional(object({
        background_color     = optional(string)
        background_image_url = optional(string, "")
        page_layout          = optional(string)
      }))
      widget = optional(object({
        header_text_alignment = optional(string)
        logo_height           = optional(number)
        logo_position         = optional(string)
        logo_url              = optional(string, "")
        social_buttons_layout = optional(string)
      }))
      identifiers = optional(object({
        login_display    = string
        otp_autocomplete = optional(bool, false)
        phone_display = object({
          formatting = string
          masking    = string
        })
      }))
    }))
  })
}

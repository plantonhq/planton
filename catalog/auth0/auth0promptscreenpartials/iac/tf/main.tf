# Auth0PromptScreenPartials Main Resources
#
# auth0_prompt_screen_partials sets every partial of one prompt on the tenant
# the provider's credential belongs to: the HTML fragments each of the
# prompt's screens renders at its insertion points. Auth0 replaces the
# prompt's whole set on every write, so a screen or insertion point the spec
# leaves out renders nothing. Destroy writes an empty set, removing every
# partial of the prompt.
resource "auth0_prompt_screen_partials" "this" {
  prompt_type = var.spec.prompt_type

  dynamic "screen_partials" {
    for_each = local.screen_partials
    content {
      screen_name = screen_partials.value.screen_name

      insertion_points {
        form_content            = screen_partials.value.insertion_points.form_content
        form_content_start      = screen_partials.value.insertion_points.form_content_start
        form_content_end        = screen_partials.value.insertion_points.form_content_end
        form_footer_start       = screen_partials.value.insertion_points.form_footer_start
        form_footer_end         = screen_partials.value.insertion_points.form_footer_end
        secondary_actions_start = screen_partials.value.insertion_points.secondary_actions_start
        secondary_actions_end   = screen_partials.value.insertion_points.secondary_actions_end
      }
    }
  }
}

# Local values for the Auth0PromptScreenPartials module.
#
# The provider reads the prompt's partials back ordered by screen name, so the
# screens are declared in that order too and a spec that lists them otherwise
# plans no change. Each insertion point left empty renders as null, so the
# provider never sends it and the screen renders nothing there. The Pulumi
# module's locals.go applies the same rules -- keep them in lockstep.
locals {
  screens_by_name = { for partial in var.spec.screen_partials : partial.screen_name => partial }

  screen_partials = [
    for screen_name in sort(keys(local.screens_by_name)) : {
      screen_name = screen_name
      insertion_points = {
        for point, fragment in local.screens_by_name[screen_name].insertion_points :
        point => fragment != "" ? fragment : null
      }
    }
  ]
}

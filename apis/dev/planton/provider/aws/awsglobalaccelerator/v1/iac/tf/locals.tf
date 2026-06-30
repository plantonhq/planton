locals {
  name = var.metadata.name

  tags = {
    Resource     = "true"
    Organization = var.metadata.org
    Environment  = var.metadata.env
    ResourceKind = "AwsGlobalAccelerator"
    ResourceId   = var.metadata.id
  }

  # Flatten listeners into a map keyed by listener name for for_each iteration.
  listeners_map = {
    for listener in var.spec.listeners : listener.name => listener
  }

  # Flatten all endpoint groups across all listeners into a map keyed by
  # "listener_name/group_name" for for_each iteration.
  endpoint_groups_map = {
    for pair in flatten([
      for listener in var.spec.listeners : [
        for group in listener.endpoint_groups : {
          key           = "${listener.name}/${group.name}"
          listener_name = listener.name
          group         = group
        }
      ]
    ]) : pair.key => pair
  }
}

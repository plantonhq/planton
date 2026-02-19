resource "hcloud_placement_group" "this" {
  name   = local.placement_group_name
  type   = local.placement_group_type
  labels = local.standard_labels
}

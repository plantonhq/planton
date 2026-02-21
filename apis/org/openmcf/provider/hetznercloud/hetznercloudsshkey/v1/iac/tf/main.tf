resource "hcloud_ssh_key" "this" {
  name       = local.ssh_key_name
  public_key = local.public_key
  labels     = local.standard_labels
}

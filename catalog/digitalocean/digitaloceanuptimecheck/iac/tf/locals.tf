locals {
  # One alert resource per spec row, keyed "<row index>-<alert name>". The
  # key is load-bearing in three places at once: it is this for_each key
  # (the resource address), the Pulumi module's resource name, and the key
  # of the alert_ids output -- so a blind import can find each row's id
  # from state alone. The index lets two rows share a display name without
  # colliding (the zone-records pattern); reordering rows churns addresses,
  # which is harmless for the seconds-fast alert objects.
  alerts = { for idx, alert in coalesce(var.spec.alerts, []) : "${idx}-${alert.alert_name}" => alert }

  # DigitalOcean fixes threshold and comparison for some alert types no
  # matter what is sent (measured live: a down alert created with 3 /
  # greater_than reads back as 1 / less_than; an ssl_expiry alert always
  # reads back less_than). The spec rejects those fields on those types,
  # and the module sends the API's own values so state and read-back agree
  # on every apply -- the alternative, sending nothing, reads back as a
  # perpetual diff.
  fixed_pair_types      = ["down", "down_global"]
  fixed_comparison_type = "ssl_expiry"

  alert_threshold = {
    for key, alert in local.alerts :
    key => contains(local.fixed_pair_types, alert.type) ? 1 : alert.threshold
  }
  alert_comparison = {
    for key, alert in local.alerts :
    key => contains(local.fixed_pair_types, alert.type) || alert.type == local.fixed_comparison_type ? "less_than" : alert.comparison
  }
}

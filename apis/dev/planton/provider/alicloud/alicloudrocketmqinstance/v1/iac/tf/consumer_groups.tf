resource "alicloud_rocketmq_consumer_group" "consumer_groups" {
  for_each = local.consumer_groups_map

  instance_id      = alicloud_rocketmq_instance.main.id
  consumer_group_id = each.value.consumer_group_id
  delivery_order_type = each.value.delivery_order_type != "" ? each.value.delivery_order_type : null
  remark           = each.value.remark != "" ? each.value.remark : null
  max_receive_tps  = each.value.max_receive_tps

  consume_retry_policy {
    retry_policy = (
      each.value.consume_retry_policy != null
      ? each.value.consume_retry_policy.retry_policy
      : "DefaultRetryPolicy"
    )
    max_retry_times = (
      each.value.consume_retry_policy != null
      ? each.value.consume_retry_policy.max_retry_times
      : null
    )
    dead_letter_target_topic = (
      each.value.consume_retry_policy != null && each.value.consume_retry_policy.dead_letter_target_topic != ""
      ? each.value.consume_retry_policy.dead_letter_target_topic
      : null
    )
  }
}

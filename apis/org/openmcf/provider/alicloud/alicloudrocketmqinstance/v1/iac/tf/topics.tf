resource "alicloud_rocketmq_topic" "topics" {
  for_each = local.topics_map

  instance_id  = alicloud_rocketmq_instance.main.id
  topic_name   = each.value.topic_name
  message_type = each.value.message_type
  remark       = each.value.remark != "" ? each.value.remark : null
  max_send_tps = each.value.max_send_tps
}

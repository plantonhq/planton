output "node_pool_id" {
  description = "ACK node pool ID"
  value       = alicloud_cs_kubernetes_node_pool.node_pool.node_pool_id
}

output "scaling_group_id" {
  description = "Auto Scaling group ID associated with this node pool"
  value       = alicloud_cs_kubernetes_node_pool.node_pool.scaling_group_id
}

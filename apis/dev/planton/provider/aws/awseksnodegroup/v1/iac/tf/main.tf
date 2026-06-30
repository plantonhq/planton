# AWS EKS Node Group Terraform Module
# Auto-release test: Multi-provider Terraform change (AWS component).

resource "aws_eks_node_group" "this" {
  cluster_name    = local.cluster_name
  node_role_arn   = local.node_role_arn
  subnet_ids      = local.subnet_ids
  instance_types  = [var.spec.instance_type]
  disk_size       = try(local.disk_size_gb, null)
  capacity_type   = local.capacity_type
  labels          = local.labels

  scaling_config {
    min_size     = var.spec.scaling.min_size
    max_size     = var.spec.scaling.max_size
    desired_size = var.spec.scaling.desired_size
  }

  dynamic "remote_access" {
    for_each = local.ssh_key_name != null && local.ssh_key_name != "" ? [1] : []
    content {
      ec2_ssh_key = local.ssh_key_name
    }
  }
}




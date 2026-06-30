resource "aws_eks_cluster" "this" {
  name     = local.resource_name
  role_arn = local.cluster_role_arn
  version  = local.cluster_version

  vpc_config {
    subnet_ids              = local.subnet_ids
    endpoint_private_access = local.disable_public_endpoint
    endpoint_public_access  = !local.disable_public_endpoint
    public_access_cidrs     = local.public_access_cidrs
  }

  dynamic "encryption_config" {
    for_each = local.has_kms_key ? [1] : []
    content {
      provider {
        key_arn = local.kms_key_arn
      }
      resources = ["secrets"]
    }
  }

  enabled_cluster_log_types = local.enable_control_plane_logs ? [
    "api",
    "audit", 
    "authenticator",
    "controllerManager",
    "scheduler"
  ] : []

  tags = local.tags

  depends_on = [
    aws_iam_role_policy_attachment.eks_cluster_policy
  ]
}

# Attach the required EKS cluster policy to the cluster role
resource "aws_iam_role_policy_attachment" "eks_cluster_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = local.cluster_role_name
}



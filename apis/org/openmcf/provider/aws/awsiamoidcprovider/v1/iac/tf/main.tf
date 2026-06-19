resource "aws_iam_openid_connect_provider" "this" {
  url             = local.url
  client_id_list  = local.client_id_list
  thumbprint_list = local.thumbprint_list

  tags = local.tags
}

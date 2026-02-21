output "function_id" {
  description = "The FC function ID"
  value       = alicloud_fcv3_function.main.function_id
}

output "function_name" {
  description = "The FC function name"
  value       = alicloud_fcv3_function.main.function_name
}

output "function_arn" {
  description = "The FC function ARN"
  value       = alicloud_fcv3_function.main.function_arn
}

output "compute_class_name" {
  description = "Name of the created ComputeClass (its metadata.name, from spec.name)."
  value       = var.spec.name
}

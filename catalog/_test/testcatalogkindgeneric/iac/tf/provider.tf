# Deliberately provider-free: the module provisions nothing outside the IaC
# engine's own state (terraform_data is a builtin), so the whole pipeline runs
# with zero credentials, zero network, and zero cost.
terraform {
  required_version = ">= 1.0"
}

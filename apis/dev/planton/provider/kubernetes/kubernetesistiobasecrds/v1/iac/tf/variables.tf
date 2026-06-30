##############################################
# variables.tf
#
# Input variables for the KubernetesIstioBaseCrds
# Terraform module.
##############################################

variable "metadata" {
  description = "Resource metadata including name"
  type = object({
    name = string
  })
}

# KubernetesIstioBaseCrds has no user-configurable spec fields: the installed CRD
# version is pinned to the typed SDK (see locals.tf istio_release). `any` is used so the
# module tolerates an empty spec (or any harness-supplied shape) without a strict schema.
variable "spec" {
  description = "KubernetesIstioBaseCrds specification (no configurable fields; CRD version is pinned)."
  type        = any
  default     = {}
}

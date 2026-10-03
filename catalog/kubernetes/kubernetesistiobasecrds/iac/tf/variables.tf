##############################################
# variables.tf
#
# Input variables for the KubernetesIstioBaseCrds
# Terraform module.
##############################################

variable "metadata" {
  description = "Catalog object metadata (name plus the optional Planton identity attributes the module renders as labels)."
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
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

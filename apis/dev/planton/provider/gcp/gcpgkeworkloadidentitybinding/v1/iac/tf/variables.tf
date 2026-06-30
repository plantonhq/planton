variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name = string,
    id = optional(string),
    org = optional(string),
    env = optional(string),
    labels = optional(map(string)),
    tags = optional(list(string)),
    version = optional(object({ id = string, message = string }))
  })
}

variable "project_id" {
  description = "The GCP project that hosts the GKE cluster (and therefore the Workload-Identity pool <project>.svc.id.goog)."
  type        = string
}

variable "service_account_email" {
  description = "The email of the Google Service Account to impersonate. Example: 'cert-manager@my-project.iam.gserviceaccount.com'"
  type        = string
}

variable "ksa_namespace" {
  description = "Kubernetes namespace of the ServiceAccount running in the cluster."
  type        = string
}

variable "ksa_name" {
  description = "Name of the Kubernetes ServiceAccount."
  type        = string
}

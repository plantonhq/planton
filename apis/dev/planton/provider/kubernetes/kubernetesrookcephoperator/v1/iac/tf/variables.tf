variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string,
    id      = optional(string),
    org     = optional(string),
    env     = optional(string),
    labels  = optional(map(string)),
    tags    = optional(list(string)),
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "Specification for Kubernetes Rook Ceph Operator deployment"
  type = object({
    # Target Kubernetes cluster name
    target_cluster_name = optional(string)

    # Kubernetes namespace where operator will be deployed
    namespace = optional(string, "rook-ceph")

    # Flag to indicate if the namespace should be created
    create_namespace = optional(bool, false)

    # The version of the Rook Ceph Operator Helm chart to deploy
    # https://github.com/rook/rook/releases
    operator_version = optional(string, "v1.16.6")

    # Whether the Helm chart should create and update CRDs
    crds_enabled = optional(bool, true)

    # The container specifications for the Rook Ceph Operator deployment
    container = object({
      # The CPU and memory resources allocated to the operator container
      resources = optional(object({
        # The resource limits for the container
        limits = optional(object({
          # The amount of CPU allocated (e.g., "500m" for 0.5 CPU cores)
          cpu = optional(string, "500m")
          # The amount of memory allocated (e.g., "512Mi" for 512 mebibytes)
          memory = optional(string, "512Mi")
        }))
        # The resource requests for the container
        requests = optional(object({
          # The amount of CPU allocated (e.g., "200m" for 0.2 CPU cores)
          cpu = optional(string, "200m")
          # The amount of memory allocated (e.g., "128Mi" for 128 mebibytes)
          memory = optional(string, "128Mi")
        }))
      }))
    })

    # CSI driver configuration
    csi = optional(object({
      # Enable the Ceph CSI RBD (block storage) driver
      enable_rbd_driver = optional(bool, true)
      # Enable the Ceph CSI CephFS (file storage) driver
      enable_cephfs_driver = optional(bool, true)
      # Disable the CSI driver entirely
      disable_csi_driver = optional(bool, false)
      # Enable host networking for CSI CephFS and RBD nodeplugins
      enable_csi_host_network = optional(bool, true)
      # Number of replicas for CSI provisioner deployment
      provisioner_replicas = optional(number, 2)
      # Enable CSI Addons for additional CSI functionality
      enable_csi_addons = optional(bool, false)
      # Enable NFS CSI driver for NFS storage support
      enable_nfs_driver = optional(bool, false)
    }))
  })
}

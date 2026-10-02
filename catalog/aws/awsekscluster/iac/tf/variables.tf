variable "metadata" {
  description = "Cloud resource metadata"
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

variable "spec" {
  description = "AwsEksCluster specification"
  type = object({
    # The AWS region the cluster control plane runs in. Must match the region
    # of the subnets it attaches to. Example: "us-west-2", "eu-west-1".
    region = string

    # The subnets (at least two, in distinct availability zones) where EKS
    # places the control plane's elastic network interfaces. These decide
    # which zones the API server is reachable from inside the VPC; worker
    # subnets are chosen separately on each node group. Reference AwsSubnet
    # subnet_id outputs or pass literal subnet IDs. Create-only in AWS
    # (changing the set updates in place, but the VPC itself cannot change).
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    subnet_ids = list(string)

    # The IAM role the EKS control plane assumes to manage AWS resources on
    # your behalf (ENIs, security groups, logs). The role must trust
    # eks.amazonaws.com and carry the AmazonEKSClusterPolicy managed policy --
    # attach it on the AwsIamRole itself (managed_policy_arns); this component
    # never modifies a role it merely references. Reference an AwsIamRole's
    # role_arn output or pass a literal ARN. Create-only in AWS.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    cluster_role_arn = string

    # The Kubernetes minor version of the control plane, e.g. "1.31". Leave
    # empty to let AWS pick the current default version. EKS only ever
    # upgrades one minor at a time and never downgrades -- lowering this value
    # is rejected by AWS. Node groups pin their own version, so upgrade the
    # control plane first, then roll the node groups.
    version = optional(string, "")

    # Additional security groups attached to the control plane's network
    # interfaces, on top of the cluster security group EKS always creates
    # (exported as cluster_security_group_id). Most clusters need none --
    # reach for this only for legacy rules that must ride along. Reference
    # AwsSecurityGroup security_group_id outputs or pass literal IDs.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    security_group_ids = optional(list(string), [])

    # Whether the Kubernetes API server is reachable from the internet. AWS
    # defaults this to true; set false for a fully private cluster (pair with
    # endpoint_private_access = true or the API server becomes unreachable).
    # Scope a public endpoint down with public_access_cidrs.
    endpoint_public_access = optional(bool)

    # Whether the Kubernetes API server is reachable from within the VPC
    # through private ENIs. AWS defaults this to false, which is almost never
    # what a production cluster wants: without it, in-VPC clients (nodes,
    # CI runners) reach the API server over the public endpoint.
    endpoint_private_access = optional(bool, false)

    # IPv4 CIDR blocks allowed to reach the PUBLIC API endpoint. Empty means
    # AWS's default of 0.0.0.0/0 (open to the internet -- rely on IAM/RBAC
    # only). Restricting this to office/VPN egress ranges is the single
    # cheapest hardening step for a public-endpoint cluster.
    public_access_cidrs = optional(list(string), [])

    # How control-plane-initiated traffic egresses the VPC:
    # - "AWS_MANAGED" (AWS default): EKS routes control-plane egress itself.
    # - "CUSTOMER_ROUTED": egress follows your VPC route tables (inspection/
    #   egress-firewall architectures).
    # - "CUSTOMER_ISOLATED": no control-plane egress through your VPC.
    # Reverting CUSTOMER_ROUTED back to AWS_MANAGED is not supported in place
    # -- AWS forces cluster replacement.
    control_plane_egress_mode = optional(string, "")

    # The IP address family for pod and service networking: "ipv4" (AWS
    # default) or "ipv6". Create-only: changing it replaces the cluster. IPv6
    # clusters assign pod addresses from the VPC's IPv6 CIDR and require IPv6-
    # enabled subnets.
    ip_family = optional(string, "")

    # The CIDR block Kubernetes assigns SERVICE addresses from (ipv4 clusters
    # only). Must be a /12 to /24 inside the private ranges (10/8,
    # 172.16/12, 192.168/16, 100.64/10) and must not overlap the VPC or any
    # peered/connected network -- overlap breaks routing in ways that only
    # surface later. Create-only. Empty keeps the AWS default
    # (10.100.0.0/16 or 172.20.0.0/16, whichever avoids the VPC).
    service_ipv4_cidr = optional(string, "")

    # Control-plane log types streamed to CloudWatch Logs: "api", "audit",
    # "authenticator", "controllerManager", "scheduler". Empty disables
    # control-plane logging. "audit" and "authenticator" are the two most
    # valuable in practice (who did what, and who got in); enabling all five
    # on a busy cluster carries real CloudWatch ingestion cost.
    enabled_cluster_log_types = optional(list(string), [])

    # Customer-managed KMS key for ENVELOPE ENCRYPTION of Kubernetes secrets
    # -- secrets in etcd are encrypted with a data key that this KMS key
    # wraps. Unset uses AWS-owned encryption only. One-way door: once enabled
    # it cannot be disabled or re-keyed on a live cluster (AWS forces
    # replacement). Reference an AwsKmsKey's key_arn output or pass a literal
    # ARN.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_arn = optional(string, "")

    # How identities are granted cluster access (access entries vs the legacy
    # aws-auth ConfigMap) and whether the creator gets admin. Unset keeps
    # AWS defaults (API_AND_CONFIG_MAP + creator-admin).
    access_config = optional(object({
      # The source of truth for cluster access:
      # - "API": EKS access entries only -- the modern model; IAM principals
      #   are granted access as first-class EKS resources.
      # - "API_AND_CONFIG_MAP" (AWS default): access entries plus the legacy
      #   aws-auth ConfigMap, for migration.
      # - "CONFIG_MAP": legacy aws-auth ConfigMap only.
      # Moving toward "API" is one-way on a live cluster: AWS allows
      # CONFIG_MAP -> API_AND_CONFIG_MAP -> API but never back.
      authentication_mode = optional(string, "")

      # Whether the identity creating the cluster is automatically granted
      # cluster-admin. AWS defaults this to true. Set false for clusters
      # whose admins are managed explicitly through access entries -- with no
      # other admin configured, false can lock everyone out of a fresh
      # cluster. Create-only.
      bootstrap_cluster_creator_admin_permissions = optional(bool)
    }))

    # EKS Auto Mode: AWS provisions and manages compute, block storage, and
    # load balancing for the cluster itself -- no node groups to operate.
    # The alternative to (not a companion of) explicit AwsEksNodeGroup
    # compute; in practice a cluster uses one model or the other.
    auto_mode = optional(object({
      # Turn Auto Mode on. AWS then provisions and scales EC2 capacity,
      # provisions EBS volumes, and manages load balancers for the cluster's
      # workloads without any node group.
      enabled = optional(bool, false)

      # The built-in node pools Auto Mode may launch capacity from:
      # "general-purpose" (workloads) and/or "system" (cluster-critical pods
      # on dedicated capacity). Empty enables Auto Mode with no built-in
      # pools -- capacity then comes only from custom NodePool resources
      # defined in-cluster.
      node_pools = optional(list(string), [])

      # The IAM role Auto Mode nodes assume (the node identity for launched
      # capacity). Required by AWS when node_pools is non-empty. Reference an
      # AwsIamRole's role_arn output or pass a literal ARN.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      node_role_arn = optional(string, "")
    }))

    # The upgrade support tier once this cluster's Kubernetes version leaves
    # standard support: "STANDARD" (upgrade on schedule, no surcharge) or
    # "EXTENDED" (AWS default -- stay on the version up to ~26 months at a
    # significant hourly surcharge). Choose STANDARD when you upgrade
    # promptly and want the surcharge risk gone.
    upgrade_support_type = optional(string, "")

    # Allows Amazon Application Recovery Controller to shift in-cluster
    # east-west traffic away from an impaired availability zone.
    zonal_shift_enabled = optional(bool, false)

    # Blocks cluster deletion at the EKS API until explicitly disabled --
    # the guard rail for shared/production control planes.
    deletion_protection = optional(bool, false)

    # Whether EKS installs the default self-managed add-ons (vpc-cni,
    # kube-proxy, CoreDNS) at creation. AWS defaults this to true; set false
    # for a "bring your own add-ons" cluster that manages every add-on
    # explicitly (the GitOps-friendly posture). Create-only.
    bootstrap_self_managed_addons = optional(bool)

    # Force the Kubernetes version update even if pods cannot be safely
    # drained onto the new version (pod disruption budgets that can never be
    # satisfied). Only consulted while `version` changes.
    force_update_version = optional(bool, false)

    # EKS Provisioned Control Plane: pre-provisions control-plane capacity
    # for very large or bursty clusters instead of relying on EKS's
    # reactive scaling (which can lag sudden API-server load, e.g. mass
    # node joins or operator storms). "standard" (AWS default -- reactive
    # scaling, no surcharge) or a provisioned tier of increasing capacity:
    # "tier-xl", "tier-2xl", "tier-4xl", "tier-8xl" -- each billed hourly
    # ON TOP of the cluster fee. Updates in place; empty keeps standard.
    control_plane_scaling_tier = optional(string, "")

    # EKS Hybrid Nodes: the on-premises/edge networks whose nodes and pods
    # join this cluster over your VPN or Direct Connect. Declaring the
    # ranges is free -- billing starts only when hybrid nodes register.
    # Updates in place on a live cluster.
    remote_networks = optional(object({
      # CIDR blocks of the on-premises network the hybrid NODES have their
      # addresses in. This is the range the control plane accepts kubelet
      # registrations from -- without a node's address inside one of these
      # blocks, it cannot join.
      node_cidrs = optional(list(string), [])

      # CIDR blocks the on-premises PODS have their addresses in (the CNI's
      # pod network on the hybrid nodes). Optional: needed when pods must be
      # directly routable from the cluster (e.g. webhooks running on hybrid
      # nodes); node-only setups can omit it.
      pod_cidrs = optional(list(string), [])
    }))
  })
}

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
  description = "GcpRegionNetworkEndpointGroup specification"
  type = object({
    # The GCP project that owns the network endpoint group.
    # Can be a literal project ID or a reference to a GcpProject resource.
    # If omitted, the provider's default project is used.
    # Immutable: changing it destroys and recreates the NEG.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Name of the network endpoint group in GCP. Must be 1-63 characters:
    # lowercase letters, digits, and hyphens; must start with a letter and end
    # with a letter or digit. If not specified, defaults to metadata.name.
    # Immutable: changing it destroys and recreates the NEG, briefly breaking
    # every backend service that references the old self_link.
    network_endpoint_group_name = optional(string, "")

    # Region the NEG lives in (e.g. "us-central1"). Required — a regional NEG is
    # always scoped to one region, and a serverless NEG must be in the same
    # region as the Cloud Run/Functions/App Engine workload it fronts.
    # Immutable: a NEG cannot move between regions.
    region = optional(string, "")

    # What kind of endpoints this NEG holds (default SERVERLESS). SERVERLESS
    # fronts Cloud Run/Functions/App Engine; PRIVATE_SERVICE_CONNECT fronts a
    # PSC endpoint; INTERNET_IP_PORT / INTERNET_FQDN_PORT front an external
    # origin; GCE_VM_IP_PORTMAP does PSC port mapping. Immutable.
    network_endpoint_type = optional(string)

    # What this NEG fronts and which backend service consumes it — write it for
    # the operator tracing a request path later. Immutable.
    description = optional(string, "")

    # The VPC network for PRIVATE_SERVICE_CONNECT, INTERNET, and
    # GCE_VM_IP_PORTMAP NEGs. Reference a GcpVpcNetwork or provide a network self-link
    # directly. Not used by (and rejected for) SERVERLESS NEGs — serverless
    # platforms are not attached to a VPC here. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    network = optional(string, "")

    # The subnetwork for PRIVATE_SERVICE_CONNECT and GCE_VM_IP_PORTMAP NEGs.
    # Reference a GcpSubnetwork or provide a subnetwork self-link directly.
    # Not used by serverless or internet NEGs. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    subnetwork = optional(string, "")

    # The target of a PRIVATE_SERVICE_CONNECT or INTERNET NEG. For PSC it is the
    # published service-attachment URL or a Google API bundle name (e.g.
    # "asia-northeast3-cloudkms.googleapis.com"); required for PSC NEGs.
    # Immutable.
    psc_target_service = optional(string, "")

    # Extra Private Service Connect settings for a PSC NEG. Only valid when
    # network_endpoint_type is PRIVATE_SERVICE_CONNECT.
    psc_data = optional(object({
      # The producer port a consumer PSC NEG connects to. Empty connects to the
      # first port in the producer's advertised port range. Immutable.
      producer_port = optional(string, "")
    }))

    # Front a Cloud Run service.
    cloud_run = optional(object({
      # The Cloud Run service to route to. Reference a GcpCloudRun resource or
      # provide the service name directly. GCP resolves endpoints at serving time,
      # so the service need not exist when the NEG is created. Immutable.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      service = optional(string, "")

      # Route to a specific Cloud Run revision tag for fine-grained traffic
      # splitting (e.g. "canary"). Only meaningful with service. Immutable.
      tag = optional(string, "")

      # A URL template that parses the service (and optional tag) out of each
      # request URL — for host/path-based fan-out to many services from one NEG
      # (e.g. "<service>.example.com/<tag>"). Immutable.
      url_mask = optional(string, "")
    }))

    # Front a Cloud Functions (Gen 2) function.
    cloud_function = optional(object({
      # The Cloud Functions function name to route to (case-sensitive). Accepts
      # a literal name or a reference to a GcpCloudFunction resource. GCP
      # resolves endpoints at serving time. Immutable.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      function = optional(string, "")

      # A URL template that parses the function name out of each request URL — for
      # routing many functions from one NEG. Immutable.
      url_mask = optional(string, "")
    }))

    # Front an App Engine service. The block may be empty to route to the
    # default App Engine application.
    app_engine = optional(object({
      # The App Engine service to route to. Empty routes to the default service.
      # Immutable.
      service = optional(string, "")

      # The App Engine service version to route to. Empty routes to the version
      # split configured in App Engine. Immutable.
      version = optional(string, "")

      # A URL template that parses the service (and version) out of each request
      # URL. Immutable.
      url_mask = optional(string, "")
    }))

    # Deletion policy for the NEG — what happens when this resource is
    # destroyed:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the NEG is deleted (GCP refuses while a backend
    #                service still references it — create the replacement
    #                before destroying the original)
    #   "PREVENT" -- destroy FAILS; protects the attachment a live backend
    #                service routes through
    #   "ABANDON" -- the NEG is removed from management but left in GCP
    deletion_policy = optional(string, "")
  })
}

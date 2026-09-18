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
  description = "GcpHealthCheck specification"
  type = object({
    # The GCP project that owns the health check.
    # Can be a literal project ID or a reference to a GcpProject resource.
    # If omitted, the provider's default project is used.
    # Immutable: changing it destroys and recreates the health check.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Name of the health check in GCP. Must be 1-63 characters: lowercase
    # letters, digits, and hyphens; must start with a letter and end with a
    # letter or digit. If not specified, defaults to metadata.name.
    # Immutable: changing it destroys and recreates the health check, briefly
    # breaking every backend service that references the old self_link.
    health_check_name = optional(string, "")

    # Region for a REGIONAL health check (e.g. "us-central1"), used by regional
    # backend services and regional managed instance groups. Leave empty for a
    # GLOBAL health check — the right scope for global external load balancers.
    # Immutable: a health check cannot move between scopes or regions.
    region = optional(string, "")

    # What this health check probes and which backends rely on it — write it
    # for the operator debugging failover behavior later. Mutable.
    description = optional(string, "")

    # Seconds between probe attempts from each prober (default 5). Lower values
    # detect failures faster at the cost of more probe traffic; the effective
    # detection time also depends on unhealthy_threshold. Mutable.
    check_interval_sec = optional(number)

    # Seconds to wait for a probe response before counting the attempt as a
    # failure (default 5). Must not exceed check_interval_sec — GCP rejects a
    # timeout longer than the interval. Mutable.
    timeout_sec = optional(number)

    # Consecutive successes required to mark a backend healthy again
    # (default 2). Higher values prevent flapping backends from re-entering
    # rotation too quickly. Mutable.
    healthy_threshold = optional(number)

    # Consecutive failures required to mark a backend unhealthy (default 2).
    # Failure detection time ≈ check_interval_sec × unhealthy_threshold — tune
    # both together when tightening failover. Mutable.
    unhealthy_threshold = optional(number)

    # Export a log entry on every health status change. Off by default —
    # enable it while tuning thresholds or debugging flapping backends, and
    # consider the log volume on large backend fleets. Mutable.
    enable_logging = optional(bool, false)

    # GLOBAL checks only: probe from exactly 3 specific GCP regions instead of
    # Google's default prober set, so a regional outage cannot flip global
    # health verdicts. Constraints enforced by GCP: exactly 3 regions; only
    # HTTP, HTTPS, and TCP protocols; check_interval_sec at least 30;
    # proxy_header must be NONE; TCP request payload unsupported; a check with
    # source_regions cannot be used by managed-instance-group auto-healing.
    source_regions = optional(list(string), [])

    # What happens to the health check in GCP when this resource is destroyed.
    # Applies to whichever scope the check was created in (global or regional).
    #   "DELETE"  -- (GCP's default when unset) the health check is deleted;
    #                any backend service still referencing it makes the delete
    #                fail on the API side, so tear consumers down first
    #   "PREVENT" -- destroy FAILS; protects a probe that many backend
    #                services may share
    #   "ABANDON" -- the health check is removed from management but keeps
    #                probing in GCP (free at rest; clean it up manually)
    deletion_policy = optional(string, "")

    # Probe with an HTTP GET (effective default port 80). The workhorse for
    # serverless NEGs, instance groups serving plaintext HTTP, and anything
    # behind an internal load balancer.
    http = optional(object({
      # Value of the Host header in the probe request. Empty uses the IP of the
      # backend being probed — set it when backends route by virtual host.
      host = optional(string, "")

      # TCP port for the probe when port_specification is USE_FIXED_PORT or
      # unset. Omit to use the protocol default (80). Must be 1-65535.
      port = optional(number, 0)

      # Instance-group named port to probe (only with USE_NAMED_PORT). Named
      # ports let each instance group map the logical port to its own number.
      port_name = optional(string, "")

      # How the probe port is chosen: USE_FIXED_PORT (the `port` field, the
      # default), USE_NAMED_PORT (`port_name` on the instance group), or
      # USE_SERVING_PORT (the backend's own serving port — the right choice for
      # serverless NEGs and most instance-group backends).
      port_specification = optional(string, "")

      # Prepend a PROXY protocol v1 header to the probe (PROXY_V1) so backends
      # behind a proxy see the original client info. Default NONE.
      proxy_header = optional(string, "")

      # Request path of the probe GET (default "/"). Point it at a cheap,
      # dependency-free endpoint (e.g. /healthz) — a path that touches databases
      # turns their latency into load-balancer failovers.
      request_path = optional(string, "")

      # If set, the response body must START WITH this ASCII string for the
      # probe to pass — a guard against a wrong service answering 200 on the
      # probed port.
      response = optional(string, "")
    }))

    # Probe with an HTTPS GET (effective default port 443). Requires the
    # backend to present a certificate; use when backends redirect or reject
    # plaintext.
    https = optional(object({
      # Value of the Host header in the probe request. Empty uses the IP of the
      # backend being probed — set it when backends route or present
      # certificates by virtual host.
      host = optional(string, "")

      # TCP port for the probe when port_specification is USE_FIXED_PORT or
      # unset. Omit to use the protocol default (443). Must be 1-65535.
      port = optional(number, 0)

      # Instance-group named port to probe (only with USE_NAMED_PORT).
      port_name = optional(string, "")

      # How the probe port is chosen: USE_FIXED_PORT (the `port` field, the
      # default), USE_NAMED_PORT (`port_name` on the instance group), or
      # USE_SERVING_PORT (the backend's own serving port).
      port_specification = optional(string, "")

      # Prepend a PROXY protocol v1 header to the probe (PROXY_V1). Default NONE.
      proxy_header = optional(string, "")

      # Request path of the probe GET (default "/"). Point it at a cheap,
      # dependency-free endpoint (e.g. /healthz).
      request_path = optional(string, "")

      # If set, the response body must START WITH this ASCII string for the
      # probe to pass.
      response = optional(string, "")
    }))

    # Probe with an HTTP/2 GET (effective default port 443). For backends
    # that only speak HTTP/2 over TLS.
    http2 = optional(object({
      # Value of the :authority pseudo-header in the probe request. Empty uses
      # the IP of the backend being probed.
      host = optional(string, "")

      # TCP port for the probe when port_specification is USE_FIXED_PORT or
      # unset. Omit to use the protocol default (443). Must be 1-65535.
      port = optional(number, 0)

      # Instance-group named port to probe (only with USE_NAMED_PORT).
      port_name = optional(string, "")

      # How the probe port is chosen: USE_FIXED_PORT (the `port` field, the
      # default), USE_NAMED_PORT (`port_name` on the instance group), or
      # USE_SERVING_PORT (the backend's own serving port).
      port_specification = optional(string, "")

      # Prepend a PROXY protocol v1 header to the probe (PROXY_V1). Default NONE.
      proxy_header = optional(string, "")

      # Request path of the probe GET (default "/").
      request_path = optional(string, "")

      # If set, the response body must START WITH this ASCII string for the
      # probe to pass.
      response = optional(string, "")
    }))

    # Probe by opening a TCP connection (effective default port 80),
    # optionally exchanging an ASCII request/response pair. The cheapest
    # liveness signal when no HTTP endpoint exists.
    tcp = optional(object({
      # TCP port for the probe when port_specification is USE_FIXED_PORT or
      # unset. Omit to use the protocol default (80). Must be 1-65535.
      port = optional(number, 0)

      # Instance-group named port to probe (only with USE_NAMED_PORT).
      port_name = optional(string, "")

      # How the probe port is chosen: USE_FIXED_PORT (the `port` field, the
      # default), USE_NAMED_PORT (`port_name` on the instance group), or
      # USE_SERVING_PORT (the backend's own serving port).
      port_specification = optional(string, "")

      # Prepend a PROXY protocol v1 header to the probe (PROXY_V1). Default NONE.
      proxy_header = optional(string, "")

      # ASCII string to send after the connection opens — pair with `response`
      # to verify an application-level banner instead of bare connectivity.
      request = optional(string, "")

      # If set, the first bytes the backend sends must START WITH this ASCII
      # string for the probe to pass.
      response = optional(string, "")
    }))

    # Probe by completing a TLS handshake (effective default port 443),
    # optionally exchanging an ASCII request/response pair after the
    # handshake.
    ssl = optional(object({
      # TCP port for the probe when port_specification is USE_FIXED_PORT or
      # unset. Omit to use the protocol default (443). Must be 1-65535.
      port = optional(number, 0)

      # Instance-group named port to probe (only with USE_NAMED_PORT).
      port_name = optional(string, "")

      # How the probe port is chosen: USE_FIXED_PORT (the `port` field, the
      # default), USE_NAMED_PORT (`port_name` on the instance group), or
      # USE_SERVING_PORT (the backend's own serving port).
      port_specification = optional(string, "")

      # Prepend a PROXY protocol v1 header to the probe (PROXY_V1). Default NONE.
      proxy_header = optional(string, "")

      # ASCII string to send after the TLS handshake completes.
      request = optional(string, "")

      # If set, the first bytes the backend sends must START WITH this ASCII
      # string for the probe to pass.
      response = optional(string, "")
    }))

    # Probe the standard gRPC health checking service
    # (grpc.health.v1.Health/Check) over plaintext.
    grpc = optional(object({
      # The gRPC service name passed to Health/Check. Empty probes the server's
      # OVERALL health; set it to probe one service on a multi-service server.
      # The backend's health implementation must recognize the same string.
      grpc_service_name = optional(string, "")

      # TCP port for the probe when port_specification is USE_FIXED_PORT or
      # unset. Must be 1-65535.
      port = optional(number, 0)

      # Instance-group named port to probe (only with USE_NAMED_PORT).
      port_name = optional(string, "")

      # How the probe port is chosen: USE_FIXED_PORT (the `port` field, the
      # default), USE_NAMED_PORT (`port_name` on the instance group), or
      # USE_SERVING_PORT (the backend's own serving port).
      port_specification = optional(string, "")
    }))

    # Probe the standard gRPC health checking service over TLS.
    grpc_tls = optional(object({
      # The gRPC service name passed to Health/Check. Empty probes the server's
      # OVERALL health.
      grpc_service_name = optional(string, "")

      # TCP port for the probe when port_specification is USE_FIXED_PORT or
      # unset. Must be 1-65535 when set. Optional — the provider does not
      # require it; set it explicitly for a deterministic probe target.
      port = optional(number, 0)

      # How the probe port is chosen: USE_FIXED_PORT (the `port` field, the
      # default) or USE_SERVING_PORT (the backend's own serving port).
      # USE_NAMED_PORT is not supported for gRPC-with-TLS probes.
      port_specification = optional(string, "")
    }))
  })
}

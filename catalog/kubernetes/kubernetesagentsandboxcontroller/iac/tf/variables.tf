variable "metadata" {
  description = "Catalog object metadata"
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
  description = "KubernetesAgentSandboxController specification"
  type = object({
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    namespace = string

    create_namespace = optional(bool, false)

    # The lifecycle of the four resource definitions (Sandbox, and with
    # extensions SandboxTemplate, SandboxWarmPool and SandboxClaim). Empty =
    # install them from the pinned release and keep them on destroy.
    crds = optional(object({
      # Apply the definitions from the pinned release. Set false only when
      # another tool manages them (a GitOps repository applying upstream's
      # definitions) and no other installation of agent-sandbox runs on the
      # cluster; the controller still installs and needs them.
      install = optional(bool)

      # Leave the definitions on the cluster when this resource is destroyed.
      # Deleting a definition deletes every object of that type, so false
      # deletes every Sandbox, SandboxTemplate, SandboxWarmPool and SandboxClaim
      # on the cluster, with their pods; their volumes follow their own
      # reclaim policy. Keep it true outside throwaway clusters.
      keep_on_uninstall = optional(bool)
    }))

    # The extensions: SandboxTemplate, SandboxWarmPool and SandboxClaim, and
    # the controllers that serve them. Absent = core only (Sandbox), which is
    # upstream's default and all a product that creates Sandboxes directly
    # needs. Present = on; turn it on for warm pools and claims, which is what
    # upstream's SDKs use.
    extensions = optional(object({
      # Run the extension controllers and install their definitions. False keeps
      # the settings below while turning the extensions off; their definitions
      # stay on the cluster (they are kept), so existing objects survive.
      enabled = optional(bool)

      # Label domains a SandboxClaim may set on its sandbox's pod through
      # `additionalPodMetadata.labels`. A label is allowed when its prefix is a
      # listed domain or a subdomain of one; every other prefixed label, and
      # every unprefixed one, refuses the claim. The list replaces upstream's
      # default, so include "sandbox.users.io" to keep it. The controller reads
      # the list at startup; a change restarts it. Empty = "sandbox.users.io".
      allowed_label_domains = optional(list(string), [])

      # The SandboxClaim controller.
      claim = optional(object({
        # Claims reconciled at once.
        concurrent_workers = optional(number)

        # How long a new claim waits for a warm pool member to report a pod IP
        # before it creates a cold sandbox instead (Go duration such as "2s").
        # Raise it when warm members report late and claims start cold.
        warm_candidate_grace_period = optional(string)

        # Stop the claim controller emitting Kubernetes Events, saving API writes
        # during large claim bursts.
        events_disabled = optional(bool, false)

        # Stop persisting a claim's observability annotations (when the controller
        # first saw it, its trace context), saving one API write per claim.
        # Startup-latency metrics and trace propagation keep working inside the
        # running controller; the on-object breadcrumbs are lost, and so is the
        # latency metric for claims first seen before a controller restart.
        observability_annotations_disabled = optional(bool, false)
      }))

      # The SandboxWarmPool controller.
      warm_pool = optional(object({
        # Pools reconciled at once.
        concurrent_workers = optional(number)

        # Sandboxes created or deleted in one batch. A pool fills in about
        # replicas / max_batch_size round trips to the API server; raising it
        # trades round trips for burst size.
        max_batch_size = optional(number)

        # How long refill waits after members leave the pool (claims adopting
        # them), so a claim burst gets the API server first (Go duration). The
        # wait restarts while members keep leaving. "0s" refills at once.
        replenish_delay = optional(string)

        # Most replacement sandboxes created per second per pool, which turns
        # refill into a steady stream instead of whole-deficit bursts. Combines
        # with replenish_delay: the delay decides when refill starts, the rate how
        # fast it flows. 0 leaves refill unpaced.
        max_refill_rate = optional(number)

        # How long a pool member may stay not Ready, counted from when its pod was
        # scheduled, before the controller replaces it, or holds it when its pod
        # cannot be scheduled (Go duration). Raise it for images that take long to
        # initialize.
        readiness_grace_period = optional(string)

        # How often a pool holding unschedulable members past their grace period
        # is checked again (Go duration).
        unschedulable_recheck_interval = optional(string)

        # Mark pool members' pods safe for the cluster autoscaler to evict, so an
        # idle warm pool never holds a node up. Set false to keep warm members
        # where they are.
        eviction_enabled = optional(bool)
      }))

      # The SandboxTemplate controller.
      template = optional(object({
        # Templates reconciled at once.
        concurrent_workers = optional(number)
      }))
    }))

    # The controller Deployment (`agent-sandbox-controller`): replicas,
    # resources, scheduling and every controller setting. Empty = upstream's
    # release manifest as published.
    controller = optional(object({
      # Override the controller image (a patched build, or a private mirror's
      # path). Empty = registry.k8s.io/agent-sandbox/agent-sandbox-controller at
      # the pinned release (moved to image_registry when that is set).
      image = optional(object({
        # The repository of the image (e.g., "gcr.io/project/image").
        repo = optional(string, "")

        # The tag of the image (e.g., "latest" or "1.0.0").
        tag = optional(string, "")

        # The name of the image pull secret for private image repositories.
        pull_secret_name = optional(string, "")
      }))

      # Controller pods. One leads and the rest wait as hot standbys, so more
      # than one needs leader election (on by default).
      replicas = optional(number)

      # Resources for the controller container. Empty = none set (upstream's
      # manifest sets none); set requests on clusters with resource quotas.
      resources = optional(object({
        # The resource limits for the container.
        # Specify the maximum amount of CPU and memory that the container can use.
        limits = optional(object({
          cpu    = optional(string, "")
          memory = optional(string, "")
        }))

        # The resource requests for the container.
        # Specify the minimum amount of CPU and memory that the container is guaranteed.
        requests = optional(object({
          cpu    = optional(string, "")
          memory = optional(string, "")
        }))
      }))

      # Node selector for the controller pods.
      node_selector = optional(map(string), {})

      # Tolerations for the controller pods.
      tolerations = optional(list(object({
        # Taint key to tolerate. Empty key with operator "Exists" tolerates every taint.
        key = optional(string, "")

        # How key/value match: "Equal" (default — value must match too) or "Exists"
        # (key presence alone matches).
        operator = optional(string, "")

        # Taint value to match when operator is "Equal".
        value = optional(string, "")

        # Which taint effect is tolerated: "NoSchedule", "PreferNoSchedule", or
        # "NoExecute". Empty tolerates all effects for the key.
        effect = optional(string, "")

        # For "NoExecute" taints only: how many seconds already-running pods stay bound
        # after the taint appears. Unset means tolerate forever.
        toleration_seconds = optional(number)
      })), [])

      # Leader election among controller pods.
      leader_election = optional(object({
        # Elect one active controller among the replicas. Turn it off only for a
        # single replica that must start without waiting for a lease.
        enabled = optional(bool)

        # Namespace that holds the election lease. Empty = the controller's own
        # namespace, `agent-sandbox-system`.
        namespace = optional(string, "")
      }))

      # The metrics endpoint the controller serves, and its TLS.
      metrics = optional(object({
        # Port the metrics endpoint listens on, and the port of the
        # `agent-sandbox-controller` Service in front of it. The conventional
        # port for HTTPS metrics is 8443.
        port = optional(number)

        # Publish the Kubernetes API client's own metrics (request latency, DNS
        # resolution, request and response sizes, rate limiting, retries). Set
        # false if their histograms cost too much.
        rest_client_metrics_enabled = optional(bool)

        # Serve metrics over HTTPS. Absent = plain HTTP.
        tls = optional(object({
          # Serve metrics over HTTPS. False keeps the settings below while serving
          # plain HTTP.
          enabled = optional(bool)

          # Secret in `agent-sandbox-system` holding the server certificate as
          # `tls.crt` and `tls.key`, typically written by a cert-manager
          # Certificate. Empty = the controller generates a self-signed certificate
          # at startup, which scrapers must then be told to trust or skip.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          certificate_secret = optional(string, "")

          # Oldest TLS version the metrics endpoint accepts. Empty = TLS 1.2 (Go's
          # default).
          min_version = optional(string, "")

          # Cipher suites the metrics endpoint offers, by Go name (for example
          # "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"). TLS 1.3 suites are not
          # configurable. Empty = Go's defaults.
          cipher_suites = optional(list(string), [])
        }))
      }))

      # Port of the controller's /healthz and /readyz endpoints.
      health_probe_port = optional(number)

      # The Sandbox controller.
      sandbox = optional(object({
        # Sandboxes reconciled at once.
        concurrent_workers = optional(number)

        # Stop the Sandbox controller emitting Kubernetes Events, saving API
        # writes during large warm-pool fills.
        events_disabled = optional(bool, false)

        # Window over which the Sandbox controller coalesces its recoverable,
        # metadata-only writes into one (Go duration). "0s" writes each at once.
        write_behind_window = optional(string)
      }))

      # Limit the controller's Pod and Service caches to objects carrying the
      # sandbox tracking label, cutting watch traffic, CPU and memory from the
      # whole cluster's pods to the sandboxes' own. Turn it on for shared or
      # busy clusters. Pods and Services created outside the controller and
      # adopted through the `agents.x-k8s.io/adoptable` label must then also
      # carry the tracking label, or the controller cannot see them.
      cache_label_selectors = optional(bool, false)

      # How the controller talks to the Kubernetes API server. Touch these only
      # when a large fleet makes the controller wait on its own API client.
      kube_api = optional(object({
        # Client-side limit on requests per second. Empty = no client-side limit
        # (the API server's own flow control still applies).
        qps = optional(number)

        # Requests allowed above qps in a burst.
        burst = optional(number)

        # Separate HTTP/2 connections to the API server for requests other than
        # watches. The API server caps concurrent requests per connection (100 by
        # default), so one connection caps the controller there whatever its
        # workers and qps; more connections raise the ceiling about N times.
        connections = optional(number)

        # Give the controller's watch streams their own connection, so watch
        # events never queue behind bursts of writes on a shared one.
        separate_watch_connection = optional(bool, false)
      }))

      # Send traces to an OpenTelemetry collector over OTLP gRPC. Absent = no
      # tracing.
      tracing = optional(object({
        # The collector's OTLP gRPC endpoint, for example
        # "http://otel-collector.observability.svc.cluster.local:4317". The
        # controller exits at startup when it cannot set up the exporter.
        otlp_endpoint = string

        # Connect to the collector without TLS.
        insecure = optional(bool, false)

        # Headers sent with every export, read from a key of a Secret in
        # `agent-sandbox-system` in OpenTelemetry's form
        # ("name1=value1,name2=value2"), for example a hosted collector's API key
        # ("x-honeycomb-team=..."). The other OpenTelemetry exporter settings
        # (certificates, compression, timeouts, per-signal endpoints) are not
        # modeled; the guide says why.
        headers = optional(object({
          # The Secret's name.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          secret = string

          # The key within the Secret.
          key = string
        }))
      }))

      # Profiling endpoints on the metrics server, for diagnosing a live
      # controller. Absent = off.
      profiling = optional(object({
        # Serve the CPU profile at /debug/pprof/profile.
        cpu_enabled = optional(bool, false)

        # Serve every pprof endpoint (heap, goroutine, block, mutex, trace,
        # cmdline, symbol) and fgprof, the CPU profile included. These can expose
        # sensitive process details and cost performance; use them while
        # diagnosing, then turn them off.
        debug_enabled = optional(bool, false)

        # Block-profile sampling rate when debug_enabled is on: 0 or less
        # disables it, 1 samples every blocking event, N samples one event per N
        # nanoseconds blocked (1000000 is about 1ms).
        block_profile_rate = optional(number)

        # Mutex-contention sampling when debug_enabled is on: 0 or less disables
        # it, 1 samples every event, N samples about one in N.
        mutex_profile_fraction = optional(number)
      }))

      # Log format and verbosity. Empty = JSON at info level.
      logging = optional(object({
        # Development logging: human-readable console output, debug level and
        # stack traces from warnings up, unless the settings below say otherwise.
        development = optional(bool, false)

        # Log encoding. Empty = JSON (console when development is on).
        encoder = optional(string, "")

        # Verbosity: "debug", "info", "error" or "panic", or a whole number above
        # 0 for increasingly verbose debug levels. Empty = "info" ("debug" when
        # development is on).
        level = optional(string, "")

        # Level from which log lines carry a stack trace. Empty = error (warn when
        # development is on).
        stacktrace_level = optional(string, "")

        # How timestamps are written. Empty = RFC 3339.
        time_encoding = optional(string, "")
      }))
    }))

    # The router (`sandbox-router`): the HTTP entry point that forwards a
    # request to a sandbox named in its `X-Sandbox-ID` header (or in the
    # path). Absent = no router. Present = installed; its authorization mode
    # must be chosen explicitly, because upstream's default lets every
    # request through.
    router = optional(object({
      # Install the router. False keeps the settings below while removing it.
      enabled = optional(bool)

      # Override the router image. Empty =
      # registry.k8s.io/agent-sandbox/sandbox-router-go at the pinned release
      # (moved to image_registry when that is set).
      image = optional(object({
        # The repository of the image (e.g., "gcr.io/project/image").
        repo = optional(string, "")

        # The tag of the image (e.g., "latest" or "1.0.0").
        tag = optional(string, "")

        # The name of the image pull secret for private image repositories.
        pull_secret_name = optional(string, "")
      }))

      # Router pods, spread across zones and nodes. Two is the minimum for
      # availability, not a capacity figure; the router is CPU-bound at high
      # request rates.
      replicas = optional(number)

      # Resources for the router container. Empty = upstream's: requests 250m
      # CPU and 256Mi memory, limits 2 CPU and 1Gi.
      resources = optional(object({
        # The resource limits for the container.
        # Specify the maximum amount of CPU and memory that the container can use.
        limits = optional(object({
          cpu    = optional(string, "")
          memory = optional(string, "")
        }))

        # The resource requests for the container.
        # Specify the minimum amount of CPU and memory that the container is guaranteed.
        requests = optional(object({
          cpu    = optional(string, "")
          memory = optional(string, "")
        }))
      }))

      # Node selector for the router pods.
      node_selector = optional(map(string), {})

      # Tolerations for the router pods.
      tolerations = optional(list(object({
        # Taint key to tolerate. Empty key with operator "Exists" tolerates every taint.
        key = optional(string, "")

        # How key/value match: "Equal" (default — value must match too) or "Exists"
        # (key presence alone matches).
        operator = optional(string, "")

        # Taint value to match when operator is "Equal".
        value = optional(string, "")

        # Which taint effect is tolerated: "NoSchedule", "PreferNoSchedule", or
        # "NoExecute". Empty tolerates all effects for the key.
        effect = optional(string, "")

        # For "NoExecute" taints only: how many seconds already-running pods stay bound
        # after the taint appears. Unset means tolerate forever.
        toleration_seconds = optional(number)
      })), [])

      # Router pods that must stay up during voluntary disruptions such as node
      # drains (the router's PodDisruptionBudget). It must be lower than
      # replicas, or a drain waits forever. 0 removes the budget.
      min_available = optional(number)

      # Who may reach a sandbox through the router. Required: upstream's default
      # lets every request through, so an open router is chosen, never
      # inherited.
      authorization = object({
        # Every request is let through. Anything that can reach the router can
        # reach every sandbox, so pair it with a NetworkPolicy, a Gateway that
        # authenticates, or a cluster you fully trust.
        allow_all = optional(object({}))

        # Each request carries a Kubernetes bearer token
        # (`Authorization: Bearer ...`), checked through the TokenReview API.
        # Fits callers that run in the cluster with a ServiceAccount.
        token_review = optional(object({
          # How long a review's answer is cached (Go duration). Shorter catches a
          # revoked token sooner at the cost of more API server load.
          ttl = optional(string)

          # Cached review answers kept before the oldest is evicted.
          cache_size = optional(number)

          # Refuse requests without a bearer token (401). Off lets tokenless
          # requests through, which only suits a rollout while callers start
          # sending tokens.
          require_token = optional(bool, false)

          # Audiences a token's `aud` claim must include. Required for projected
          # ServiceAccount tokens minted with an audience. Empty = no audience
          # check.
          audiences = optional(list(string), [])
        }))

        # Each request carries a token scoped to one sandbox, minted out of band
        # by the product that created it: v1 tokens are HMAC-signed with a
        # shared secret, v2 tokens are Ed25519-signed and bound to the sandbox's
        # UID.
        scoped_token = optional(object({
          # The shared HMAC-SHA256 secret that verifies v1 tokens, as a key of a
          # Secret in `agent-sandbox-system`.
          v1_secret = optional(object({
            # The Secret's name.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            secret = string

            # The key within the Secret.
            key = string
          }))

          # The JSON key set of Ed25519 public keys that verifies v2 tokens, as a
          # key of a Secret in `agent-sandbox-system`. Needs the pod-IP cache.
          verification_keys = optional(object({
            # The Secret's name.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            secret = string

            # The key within the Secret.
            key = string
          }))

          # The moment v1 tokens stop being accepted while v1 and v2 run side by
          # side (RFC 3339, for example "2026-12-31T00:00:00Z", exclusive).
          # Required when both are set.
          v1_accept_until = optional(string, "")
        }))

        # Also accept the credential from a cookie, the one channel a browser
        # attaches to every request, a WebSocket handshake included. Needs
        # token_review or scoped_token, and proxy.path_routing_prefix.
        cookie = optional(object({
          # The cookie's name: letters, digits, "-", "_", "." and "~".
          name = string

          # A URL query parameter that sets the cookie: a GET or HEAD request that
          # presents a valid credential here gets the cookie, scoped to exactly the
          # sandbox it authorized, and is redirected to the same URL without the
          # parameter, so the credential spends the least time in URLs, history and
          # Referer headers. Empty = the cookie must be set some other way.
          query_param = optional(string, "")

          # The cookie's SameSite attribute. Empty = lax, which covers same-site
          # embedding without HTTPS. strict is never sent cross-site, including the
          # navigation that sets it. none is for genuinely cross-site embedding,
          # needs a Secure cookie, and some browsers block it regardless.
          same_site = optional(string, "")

          # Drop the cookie's Secure attribute. Only for local development over
          # plain HTTP; not allowed with same_site none.
          insecure = optional(bool, false)

          # Origins (scheme://host[:port]) whose pages may make cookie-authorized
          # requests. Same-origin requests and requests without an Origin header
          # are always let through; only cookie-borne credentials are checked.
          allowed_origins = optional(list(string), [])

          # Read the request's scheme from X-Forwarded-Proto for the same-origin
          # check. Needed behind a TLS-terminating load balancer or Gateway; turn it
          # on only when the router is reachable solely through a proxy that sets
          # the header itself and strips any the client sent.
          trust_forwarded_proto = optional(bool, false)
        }))
      })

      # Serve the proxy over TLS. Absent = plain HTTP only.
      tls = optional(object({
        # Serve the proxy over TLS. False keeps the settings below while serving
        # plain HTTP only.
        enabled = optional(bool)

        # Secret in `agent-sandbox-system` holding the server certificate as
        # `tls.crt` and `tls.key`, typically written by a cert-manager
        # Certificate.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        certificate_secret = string

        # Port of the TLS listener, and of the router Service's `proxy-tls` port.
        port = optional(number)

        # Keep the plain-HTTP listener beside TLS. Set false to serve TLS only.
        http_enabled = optional(bool)

        # Client-certificate verification. Empty = off. optional verifies a
        # certificate when a client presents one; required refuses clients
        # without one. Both need client_ca.
        mtls_mode = optional(string, "")

        # The CA bundle that verifies client certificates, as a key of a Secret in
        # `agent-sandbox-system` (a cert-manager Certificate's Secret carries it
        # as `ca.crt`).
        client_ca = optional(object({
          # The Secret's name.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          secret = string

          # The key within the Secret.
          key = string
        }))

        # Oldest TLS version the listener accepts. Empty = TLS 1.2.
        min_version = optional(string, "")

        # Cipher suites the listener offers, by Go name. TLS 1.3 suites are not
        # configurable. Empty = Go's defaults.
        cipher_suites = optional(list(string), [])
      }))

      # How requests are forwarded to sandboxes.
      proxy = optional(object({
        # Total time budget for proxying one request (Go duration). Long-running
        # requests and streams need more.
        timeout = optional(string)

        # Longest wait for a sandbox's response headers (Go duration).
        response_header_timeout = optional(string)

        # Time to drain in-flight requests when a router pod stops (Go duration).
        # The pod's termination grace period is set 15 seconds longer.
        shutdown_timeout = optional(string)

        # Largest request body accepted, in bytes. 0 = no limit.
        max_request_body_bytes = optional(number, 0)

        # Accept a loopback address (127.0.0.0/8, ::1) in X-Sandbox-Pod-IP. Only
        # for a router running as a sidecar beside its sandbox, or tests against
        # localhost; link-local, multicast and unspecified addresses stay refused.
        allow_loopback_pod_ip = optional(bool, false)

        # A path prefix that also lets a caller address a sandbox by URL:
        # <prefix>/<namespace>/<id>/<port>/<rest>. Browsers need it: a WebSocket
        # handshake cannot carry custom headers, so a web IDE's terminal or a dev
        # server's live reload inside a sandbox is reachable only by path. Every
        # path under the prefix is claimed for routing, so pick one no sandbox
        # route uses. Empty = headers only.
        path_routing_prefix = optional(string, "")

        # Extra connection attempts when a sandbox cannot be dialed (DNS failure,
        # connection refused) before the router answers 502. Smooths requests to
        # a sandbox that is still starting. 0 = no retries.
        upstream_max_retries = optional(number)

        # Wait before the first retry; each later wait doubles, up to
        # upstream_retry_max_delay (Go duration).
        upstream_retry_initial_delay = optional(string)

        # Longest wait between retries (Go duration).
        upstream_retry_max_delay = optional(string)
      }))

      # The router's pod-IP cache.
      cache = optional(object({
        # Watch sandbox pods and dial a cached pod IP for requests that carry
        # X-Sandbox-UID, skipping DNS. Needs read access to pods, which the router
        # gets cluster-wide (it filters system pods out itself). Off = DNS-only
        # routing, and no pod read access.
        enabled = optional(bool)

        # Watch only this namespace's pods. Empty = every namespace.
        namespace = optional(string, "")
      }))

      # Send traces and metrics to an OpenTelemetry collector over OTLP gRPC.
      # Absent = none (the Prometheus /metrics endpoint is always served).
      telemetry = optional(object({
        # The collector's OTLP gRPC endpoint, for example
        # "http://otel-collector.observability.svc.cluster.local:4317".
        otlp_endpoint = string

        # Connect to the collector without TLS.
        insecure = optional(bool, false)

        # Send traces.
        traces_enabled = optional(bool)

        # Also push metrics (the Prometheus /metrics endpoint stays served).
        metrics_enabled = optional(bool)

        # Headers sent with every export, read from a key of a Secret in
        # `agent-sandbox-system` in OpenTelemetry's form
        # ("name1=value1,name2=value2"), for example a hosted collector's API key.
        headers = optional(object({
          # The Secret's name.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          secret = string

          # The key within the Secret.
          key = string
        }))
      }))

      # Log one structured line per proxied request (health and metrics
      # requests are skipped).
      access_log_enabled = optional(bool)

      # The router's plain-HTTP, metrics and health ports.
      ports = optional(object({
        # Port of the plain-HTTP proxy listener, and of the `sandbox-router-svc`
        # Service's `proxy` port that upstream's SDKs dial.
        http = optional(number)

        # Port of the Prometheus /metrics endpoint.
        metrics = optional(number)

        # Port of the /healthz and /readyz endpoints the pod's probes use.
        health_probe = optional(number)
      }))

      # The router's NetworkPolicy.
      network_policy = optional(object({
        # Install the router's NetworkPolicy. False leaves the router's traffic
        # unrestricted.
        enabled = optional(bool)

        # Labels of the namespaces whose pods may send requests to the router,
        # for example {"kubernetes.io/metadata.name": "gateway"} for a Gateway's
        # namespace. Empty = every namespace, which is upstream's starting point;
        # narrow it to your callers.
        client_namespace_labels = optional(map(string), {})

        # Namespace whose pods may scrape the router's metrics.
        metrics_namespace = optional(string)

        # Ports the router may open to sandboxes. A sandbox serving on any other
        # port is unreachable through the router behind this policy, so list every
        # port your sandboxes serve. Empty = 8888, upstream's default sandbox
        # port.
        sandbox_ports = optional(list(number), [])
      }))
    }))

    # The cluster's DNS suffix. The controller builds each sandbox Service's
    # fully qualified name with it, and the router resolves sandboxes with it.
    # Change it only on a cluster whose DNS domain is not the default.
    cluster_domain = optional(string)

    # Registry that replaces the registry part of every image this kind runs
    # (the controller and the router, both published on registry.k8s.io) -
    # the air-gap and private-mirror path. Each image keeps its path and tag,
    # so "mirror.example.com/k8s" pulls
    # registry.k8s.io/agent-sandbox/agent-sandbox-controller:v1.0.7 as
    # mirror.example.com/k8s/agent-sandbox/agent-sandbox-controller:v1.0.7.
    # An image override on the controller or router wins over this for that
    # image. Sandbox images are your own and never move. Empty =
    # registry.k8s.io.
    image_registry = optional(string, "")

    # Names of image-pull Secrets in `agent-sandbox-system` for pulling the
    # controller and router images from a private registry.
    image_pull_secrets = optional(list(string), [])
  })
}

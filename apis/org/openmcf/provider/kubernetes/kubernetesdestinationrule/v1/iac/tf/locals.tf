locals {
  destination_rule_name = var.metadata.name
  namespace             = var.spec.namespace

  labels = {
    "app.kubernetes.io/name"       = "destination-rule"
    "app.kubernetes.io/instance"   = var.metadata.name
    "app.kubernetes.io/managed-by" = "openmcf"
    "app.kubernetes.io/component"  = "destination-rule"
  }

  # ----------------------------------------------------------------------------------------
  # TrafficPolicy fan-out.
  #
  # The upstream CRD reuses one TrafficPolicy (and one PortTrafficPolicy / LoadBalancer /
  # ConnectionPool / OutlierDetection / ClientTLSSettings) at four reachable paths:
  #   spec.trafficPolicy, spec.subsets[].trafficPolicy, and the portLevelSettings[] under
  #   each of those. HCL has no functions, so to transform each leaf exactly ONCE we gather
  #   every instance into a flat map keyed by its path, transform the map, then look the
  #   built leaf back up during assembly. tp_ctxs keys: "main" and "subset-<i>"; pls keys add
  #   "-<j>". Leaf input maps prefix "tp:" (TrafficPolicy level) and "pls:" (PortTrafficPolicy
  #   level) so a single built map covers both.
  #
  # Null handling: every block uses merge() of conditional fragments so that UNSET fields are
  # omitted entirely from the manifest (not emitted as null). This is required because the
  # DestinationRule CRD's oneOf groups (loadBalancer simple-vs-consistentHash; consistentHash
  # hashKey/hashAlgorithm arms) reject explicit-null alternatives, and objects with required
  # subfields (warmup.duration, tunnel.target*, httpCookie.name) must not appear as empty
  # objects. Required subfields are seeded as the merge base; everything else is pruned.
  # (Note: an all-conditional merge() of fields that are ALL the same collection type would
  # collapse to a map() the provider cannot morph into an object; DestinationRule has no such
  # leaf, so merge()-pruning is safe throughout here.)
  # ----------------------------------------------------------------------------------------

  tp_ctxs = merge(
    var.spec.traffic_policy != null ? { "main" = var.spec.traffic_policy } : {},
    var.spec.subsets != null ? { for i, s in var.spec.subsets : "subset-${i}" => s.traffic_policy if s.traffic_policy != null } : {},
  )

  pls_ctxs = flatten([
    for k, tp in local.tp_ctxs : (
      tp.port_level_settings != null ? [for j, p in tp.port_level_settings : { key = "${k}-${j}", p = p }] : []
    )
  ])

  # Gather every LoadBalancer / ConnectionPool / OutlierDetection / TLS instance across all
  # paths into one map, so each transform below runs exactly once.
  lb_inputs = merge(
    { for k, tp in local.tp_ctxs : "tp:${k}" => tp.load_balancer if tp.load_balancer != null },
    { for c in local.pls_ctxs : "pls:${c.key}" => c.p.load_balancer if c.p.load_balancer != null },
  )
  cp_inputs = merge(
    { for k, tp in local.tp_ctxs : "tp:${k}" => tp.connection_pool if tp.connection_pool != null },
    { for c in local.pls_ctxs : "pls:${c.key}" => c.p.connection_pool if c.p.connection_pool != null },
  )
  od_inputs = merge(
    { for k, tp in local.tp_ctxs : "tp:${k}" => tp.outlier_detection if tp.outlier_detection != null },
    { for c in local.pls_ctxs : "pls:${c.key}" => c.p.outlier_detection if c.p.outlier_detection != null },
  )
  tls_inputs = merge(
    { for k, tp in local.tp_ctxs : "tp:${k}" => tp.tls if tp.tls != null },
    { for c in local.pls_ctxs : "pls:${c.key}" => c.p.tls if c.p.tls != null },
  )

  lb_built = { for k, lb in local.lb_inputs : k => merge(
    lb.simple != null ? { simple = lb.simple } : {},
    lb.consistent_hash != null ? { consistentHash = merge(
      lb.consistent_hash.http_header_name != null ? { httpHeaderName = lb.consistent_hash.http_header_name } : {},
      lb.consistent_hash.http_cookie != null ? { httpCookie = merge(
        { name = lb.consistent_hash.http_cookie.name },
        lb.consistent_hash.http_cookie.path != null ? { path = lb.consistent_hash.http_cookie.path } : {},
        lb.consistent_hash.http_cookie.ttl != null ? { ttl = lb.consistent_hash.http_cookie.ttl } : {},
      ) } : {},
      lb.consistent_hash.use_source_ip != null ? { useSourceIp = lb.consistent_hash.use_source_ip } : {},
      lb.consistent_hash.http_query_parameter_name != null ? { httpQueryParameterName = lb.consistent_hash.http_query_parameter_name } : {},
      lb.consistent_hash.ring_hash != null ? { ringHash = lb.consistent_hash.ring_hash.minimum_ring_size != null ? { minimumRingSize = lb.consistent_hash.ring_hash.minimum_ring_size } : {} } : {},
      lb.consistent_hash.maglev != null ? { maglev = lb.consistent_hash.maglev.table_size != null ? { tableSize = lb.consistent_hash.maglev.table_size } : {} } : {},
      lb.consistent_hash.minimum_ring_size != null ? { minimumRingSize = lb.consistent_hash.minimum_ring_size } : {},
    ) } : {},
    lb.locality_lb_setting != null ? { localityLbSetting = merge(
      lb.locality_lb_setting.distribute != null ? { distribute = [for d in lb.locality_lb_setting.distribute : merge(
        d.from != null ? { from = d.from } : {},
        d.to != null ? { to = d.to } : {},
      )] } : {},
      lb.locality_lb_setting.failover != null ? { failover = [for f in lb.locality_lb_setting.failover : merge(
        f.from != null ? { from = f.from } : {},
        f.to != null ? { to = f.to } : {},
      )] } : {},
      lb.locality_lb_setting.failover_priority != null ? { failoverPriority = lb.locality_lb_setting.failover_priority } : {},
      lb.locality_lb_setting.enabled != null ? { enabled = lb.locality_lb_setting.enabled } : {},
    ) } : {},
    lb.warmup_duration_secs != null ? { warmupDurationSecs = lb.warmup_duration_secs } : {},
    lb.warmup != null ? { warmup = merge(
      { duration = lb.warmup.duration },
      lb.warmup.minimum_percent != null ? { minimumPercent = lb.warmup.minimum_percent } : {},
      lb.warmup.aggression != null ? { aggression = lb.warmup.aggression } : {},
    ) } : {},
  ) }

  cp_built = { for k, cp in local.cp_inputs : k => merge(
    cp.tcp != null ? { tcp = merge(
      cp.tcp.max_connections != null ? { maxConnections = cp.tcp.max_connections } : {},
      cp.tcp.connect_timeout != null ? { connectTimeout = cp.tcp.connect_timeout } : {},
      cp.tcp.tcp_keepalive != null ? { tcpKeepalive = merge(
        cp.tcp.tcp_keepalive.probes != null ? { probes = cp.tcp.tcp_keepalive.probes } : {},
        cp.tcp.tcp_keepalive.time != null ? { time = cp.tcp.tcp_keepalive.time } : {},
        cp.tcp.tcp_keepalive.interval != null ? { interval = cp.tcp.tcp_keepalive.interval } : {},
      ) } : {},
      cp.tcp.max_connection_duration != null ? { maxConnectionDuration = cp.tcp.max_connection_duration } : {},
      cp.tcp.idle_timeout != null ? { idleTimeout = cp.tcp.idle_timeout } : {},
    ) } : {},
    cp.http != null ? { http = merge(
      cp.http.http1_max_pending_requests != null ? { http1MaxPendingRequests = cp.http.http1_max_pending_requests } : {},
      cp.http.http2_max_requests != null ? { http2MaxRequests = cp.http.http2_max_requests } : {},
      cp.http.max_requests_per_connection != null ? { maxRequestsPerConnection = cp.http.max_requests_per_connection } : {},
      cp.http.max_retries != null ? { maxRetries = cp.http.max_retries } : {},
      cp.http.idle_timeout != null ? { idleTimeout = cp.http.idle_timeout } : {},
      cp.http.h2_upgrade_policy != null ? { h2UpgradePolicy = cp.http.h2_upgrade_policy } : {},
      cp.http.use_client_protocol != null ? { useClientProtocol = cp.http.use_client_protocol } : {},
      cp.http.max_concurrent_streams != null ? { maxConcurrentStreams = cp.http.max_concurrent_streams } : {},
    ) } : {},
  ) }

  od_built = { for k, od in local.od_inputs : k => merge(
    od.split_external_local_origin_errors != null ? { splitExternalLocalOriginErrors = od.split_external_local_origin_errors } : {},
    od.consecutive_local_origin_failures != null ? { consecutiveLocalOriginFailures = od.consecutive_local_origin_failures } : {},
    od.consecutive_gateway_errors != null ? { consecutiveGatewayErrors = od.consecutive_gateway_errors } : {},
    od.consecutive_5xx_errors != null ? { consecutive5xxErrors = od.consecutive_5xx_errors } : {},
    od.interval != null ? { interval = od.interval } : {},
    od.base_ejection_time != null ? { baseEjectionTime = od.base_ejection_time } : {},
    od.max_ejection_percent != null ? { maxEjectionPercent = od.max_ejection_percent } : {},
    od.min_health_percent != null ? { minHealthPercent = od.min_health_percent } : {},
  ) }

  tls_built = { for k, tls in local.tls_inputs : k => merge(
    tls.mode != null ? { mode = tls.mode } : {},
    tls.client_certificate != null ? { clientCertificate = tls.client_certificate } : {},
    tls.private_key != null ? { privateKey = tls.private_key } : {},
    tls.ca_certificates != null ? { caCertificates = tls.ca_certificates } : {},
    tls.credential_name != null ? { credentialName = tls.credential_name } : {},
    tls.subject_alt_names != null ? { subjectAltNames = tls.subject_alt_names } : {},
    tls.sni != null ? { sni = tls.sni } : {},
    tls.insecure_skip_verify != null ? { insecureSkipVerify = tls.insecure_skip_verify } : {},
    tls.ca_crl != null ? { caCrl = tls.ca_crl } : {},
  ) }

  # Assemble each TrafficPolicy (main + subsets), referencing the built leaves by path key.
  # A built leaf is referenced only when it exists in the map (its input was non-null); an
  # empty built object is still valid (e.g. an empty loadBalancer satisfies the CRD oneOf).
  traffic_policy_built = { for k, tp in local.tp_ctxs : k => merge(
    contains(keys(local.lb_built), "tp:${k}") ? { loadBalancer = local.lb_built["tp:${k}"] } : {},
    contains(keys(local.cp_built), "tp:${k}") ? { connectionPool = local.cp_built["tp:${k}"] } : {},
    contains(keys(local.od_built), "tp:${k}") ? { outlierDetection = local.od_built["tp:${k}"] } : {},
    contains(keys(local.tls_built), "tp:${k}") ? { tls = local.tls_built["tp:${k}"] } : {},
    tp.tunnel != null ? { tunnel = merge(
      { targetHost = tp.tunnel.target_host, targetPort = tp.tunnel.target_port },
      tp.tunnel.protocol != null ? { protocol = tp.tunnel.protocol } : {},
    ) } : {},
    tp.proxy_protocol != null ? { proxyProtocol = merge(
      tp.proxy_protocol.version != null ? { version = tp.proxy_protocol.version } : {},
    ) } : {},
    tp.port_level_settings != null ? { portLevelSettings = [
      for j, p in tp.port_level_settings : merge(
        p.port != null ? { port = { number = p.port.number } } : {},
        contains(keys(local.lb_built), "pls:${k}-${j}") ? { loadBalancer = local.lb_built["pls:${k}-${j}"] } : {},
        contains(keys(local.cp_built), "pls:${k}-${j}") ? { connectionPool = local.cp_built["pls:${k}-${j}"] } : {},
        contains(keys(local.od_built), "pls:${k}-${j}") ? { outlierDetection = local.od_built["pls:${k}-${j}"] } : {},
        contains(keys(local.tls_built), "pls:${k}-${j}") ? { tls = local.tls_built["pls:${k}-${j}"] } : {},
      )
    ] } : {},
  ) }

  # Workload selector, mapped to its camelCase CRD form (matchLabels) and omitted entirely
  # when no labels are provided. The nested match_labels is read through a ?: conditional
  # (which only evaluates the taken branch) so a null selector never triggers an attribute
  # access on null.
  workload_selector_labels = var.spec.workload_selector != null ? var.spec.workload_selector.match_labels : null

  # Assemble the full DestinationRule spec. host is always present (required upstream); every
  # other block is pruned when unset so istiod defaults flow through.
  destination_rule_spec = merge(
    { host = var.spec.host },
    contains(keys(local.traffic_policy_built), "main") ? { trafficPolicy = local.traffic_policy_built["main"] } : {},
    var.spec.export_to != null ? { exportTo = var.spec.export_to } : {},
    local.workload_selector_labels != null ? { workloadSelector = { matchLabels = local.workload_selector_labels } } : {},
    var.spec.subsets != null ? { subsets = [
      for i, s in var.spec.subsets : merge(
        { name = s.name },
        s.labels != null ? { labels = s.labels } : {},
        contains(keys(local.traffic_policy_built), "subset-${i}") ? { trafficPolicy = local.traffic_policy_built["subset-${i}"] } : {},
      )
    ] } : {},
  )
}

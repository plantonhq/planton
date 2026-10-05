# KubernetesDestinationRule

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

KubernetesDestinationRuleSpec defines an Istio DestinationRule: a namespaced resource
that configures what happens to traffic AFTER routing has selected a destination
service — load balancing, connection-pool sizing, circuit breaking / outlier detection,
and the TLS settings the sidecar uses when originating connections upstream. It can also
carve a service into named `subsets` (e.g. versions) that route rules target.

100% fidelity with the upstream istio.io/api DestinationRule
(networking/v1alpha3/destination_rule.proto, served as networking.istio.io/v1), pinned
to the 1.30 line (tag 1.30.3). Upstream spec fields are flattened directly after the
Planton namespaced envelope (namespace); there is no nested
`destination_rule` sub-message.

Istio `oneof` modeling: DestinationRule's unions
(LoadBalancerSettings.lb_policy, ConsistentHashLB.hash_key/hash_algorithm) have NO native
discriminator field upstream and their members are distinctly-named CRD keys, so they are
modeled as optional sibling fields plus an "at most one" message-level CEL — NOT a
discriminator + value (that form is reserved for same-type collapsible unions such as the
shared KubernetesIstioApiStringMatch). This keeps the CR JSON 1:1 with upstream and avoids
inventing a required input the CRD does not have. The same idiom is used by ServiceEntry
(workload_selector vs endpoints) and EnvoyFilter (workload_selector vs target_refs), and
it matches DestinationRule's own upstream XValidation `oneof(warmupDurationSecs, warmup)`.

## Example

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesDestinationRule
metadata:
  name: test-destination-rule
spec:
  namespace:
    value: test-namespace
  host:
    value: reviews.test-namespace.svc.cluster.local
  traffic_policy:
    load_balancer:
      simple: LEAST_REQUEST
    connection_pool:
      tcp:
        max_connections: 100
        connect_timeout: 30ms
      http:
        http2_max_requests: 1000
        max_requests_per_connection: 10
        h2_upgrade_policy: UPGRADE
    outlier_detection:
      consecutive_5xx_errors: 7
      interval: 5m
      base_ejection_time: 15m
      max_ejection_percent: 50
    tls:
      mode: ISTIO_MUTUAL
    retry_budget:
      percent: 25.5
      min_retry_concurrency: 4
  subsets:
    - name: v1
      labels:
        version: v1
    - name: v2
      labels:
        version: v2
      traffic_policy:
        load_balancer:
          consistent_hash:
            http_cookie:
              name: session
              ttl: 1h
              attributes:
                - name: SameSite
                  value: Strict
                - name: Secure
        retry_budget:
          percent: 10.0
  export_to:
    - "."
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.host` | `string \| valueFrom` | yes |  | KubernetesService (`status.outputs.kube_endpoint`) |
| `spec.trafficPolicy` | `KubernetesDestinationRuleTrafficPolicy` |  |  |  |
| `spec.trafficPolicy.loadBalancer` | `KubernetesDestinationRuleLoadBalancerSettings` |  |  |  |
| `spec.trafficPolicy.loadBalancer.simple` | `string` |  |  |  |
| `spec.trafficPolicy.loadBalancer.consistentHash` | `KubernetesDestinationRuleConsistentHashLb` |  |  |  |
| `spec.trafficPolicy.loadBalancer.consistentHash.httpHeaderName` | `string` |  |  |  |
| `spec.trafficPolicy.loadBalancer.consistentHash.httpCookie` | `KubernetesDestinationRuleHttpCookie` |  |  |  |
| `spec.trafficPolicy.loadBalancer.consistentHash.httpCookie.name` | `string` | yes |  |  |
| `spec.trafficPolicy.loadBalancer.consistentHash.httpCookie.path` | `string` |  |  |  |
| `spec.trafficPolicy.loadBalancer.consistentHash.httpCookie.ttl` | `string` |  |  |  |
| `spec.trafficPolicy.loadBalancer.consistentHash.httpCookie.attributes` | `[]KubernetesDestinationRuleHttpCookieAttribute` |  |  |  |
| `spec.trafficPolicy.loadBalancer.consistentHash.httpCookie.attributes[].name` | `string` | yes |  |  |
| `spec.trafficPolicy.loadBalancer.consistentHash.httpCookie.attributes[].value` | `string` |  |  |  |
| `spec.trafficPolicy.loadBalancer.consistentHash.useSourceIp` | `bool` |  |  |  |
| `spec.trafficPolicy.loadBalancer.consistentHash.httpQueryParameterName` | `string` |  |  |  |
| `spec.trafficPolicy.loadBalancer.consistentHash.ringHash` | `KubernetesDestinationRuleRingHash` |  |  |  |
| `spec.trafficPolicy.loadBalancer.consistentHash.ringHash.minimumRingSize` | `uint64` |  |  |  |
| `spec.trafficPolicy.loadBalancer.consistentHash.maglev` | `KubernetesDestinationRuleMagLev` |  |  |  |
| `spec.trafficPolicy.loadBalancer.consistentHash.maglev.tableSize` | `uint64` |  |  |  |
| `spec.trafficPolicy.loadBalancer.consistentHash.minimumRingSize` | `uint64` |  |  |  |
| `spec.trafficPolicy.loadBalancer.localityLbSetting` | `KubernetesDestinationRuleLocalityLbSetting` |  |  |  |
| `spec.trafficPolicy.loadBalancer.localityLbSetting.distribute` | `[]KubernetesDestinationRuleLocalityDistribute` |  |  |  |
| `spec.trafficPolicy.loadBalancer.localityLbSetting.distribute[].from` | `string` |  |  |  |
| `spec.trafficPolicy.loadBalancer.localityLbSetting.distribute[].to` | `map<string, uint32>` |  |  |  |
| `spec.trafficPolicy.loadBalancer.localityLbSetting.failover` | `[]KubernetesDestinationRuleLocalityFailover` |  |  |  |
| `spec.trafficPolicy.loadBalancer.localityLbSetting.failover[].from` | `string` |  |  |  |
| `spec.trafficPolicy.loadBalancer.localityLbSetting.failover[].to` | `string` |  |  |  |
| `spec.trafficPolicy.loadBalancer.localityLbSetting.failoverPriority` | `[]string` |  |  |  |
| `spec.trafficPolicy.loadBalancer.localityLbSetting.enabled` | `bool` |  |  |  |
| `spec.trafficPolicy.loadBalancer.warmupDurationSecs` | `string` |  |  |  |
| `spec.trafficPolicy.loadBalancer.warmup` | `KubernetesDestinationRuleWarmupConfiguration` |  |  |  |
| `spec.trafficPolicy.loadBalancer.warmup.duration` | `string` | yes |  |  |
| `spec.trafficPolicy.loadBalancer.warmup.minimumPercent` | `double` |  |  |  |
| `spec.trafficPolicy.loadBalancer.warmup.aggression` | `double` |  |  |  |
| `spec.trafficPolicy.connectionPool` | `KubernetesDestinationRuleConnectionPoolSettings` |  |  |  |
| `spec.trafficPolicy.connectionPool.tcp` | `KubernetesDestinationRuleTcpSettings` |  |  |  |
| `spec.trafficPolicy.connectionPool.tcp.maxConnections` | `int32` |  |  |  |
| `spec.trafficPolicy.connectionPool.tcp.connectTimeout` | `string` |  |  |  |
| `spec.trafficPolicy.connectionPool.tcp.tcpKeepalive` | `KubernetesDestinationRuleTcpKeepalive` |  |  |  |
| `spec.trafficPolicy.connectionPool.tcp.tcpKeepalive.probes` | `uint32` |  |  |  |
| `spec.trafficPolicy.connectionPool.tcp.tcpKeepalive.time` | `string` |  |  |  |
| `spec.trafficPolicy.connectionPool.tcp.tcpKeepalive.interval` | `string` |  |  |  |
| `spec.trafficPolicy.connectionPool.tcp.maxConnectionDuration` | `string` |  |  |  |
| `spec.trafficPolicy.connectionPool.tcp.idleTimeout` | `string` |  |  |  |
| `spec.trafficPolicy.connectionPool.http` | `KubernetesDestinationRuleHttpSettings` |  |  |  |
| `spec.trafficPolicy.connectionPool.http.http1MaxPendingRequests` | `int32` |  |  |  |
| `spec.trafficPolicy.connectionPool.http.http2MaxRequests` | `int32` |  |  |  |
| `spec.trafficPolicy.connectionPool.http.maxRequestsPerConnection` | `int32` |  |  |  |
| `spec.trafficPolicy.connectionPool.http.maxRetries` | `int32` |  |  |  |
| `spec.trafficPolicy.connectionPool.http.idleTimeout` | `string` |  |  |  |
| `spec.trafficPolicy.connectionPool.http.h2UpgradePolicy` | `string` |  |  |  |
| `spec.trafficPolicy.connectionPool.http.useClientProtocol` | `bool` |  |  |  |
| `spec.trafficPolicy.connectionPool.http.maxConcurrentStreams` | `int32` |  |  |  |
| `spec.trafficPolicy.outlierDetection` | `KubernetesDestinationRuleOutlierDetection` |  |  |  |
| `spec.trafficPolicy.outlierDetection.splitExternalLocalOriginErrors` | `bool` |  |  |  |
| `spec.trafficPolicy.outlierDetection.consecutiveLocalOriginFailures` | `uint32` |  |  |  |
| `spec.trafficPolicy.outlierDetection.consecutiveGatewayErrors` | `uint32` |  |  |  |
| `spec.trafficPolicy.outlierDetection.consecutive5xxErrors` | `uint32` |  |  |  |
| `spec.trafficPolicy.outlierDetection.interval` | `string` |  |  |  |
| `spec.trafficPolicy.outlierDetection.baseEjectionTime` | `string` |  |  |  |
| `spec.trafficPolicy.outlierDetection.maxEjectionPercent` | `int32` |  |  |  |
| `spec.trafficPolicy.outlierDetection.minHealthPercent` | `int32` |  |  |  |
| `spec.trafficPolicy.tls` | `KubernetesDestinationRuleClientTlsSettings` |  |  |  |
| `spec.trafficPolicy.tls.mode` | `string` |  |  |  |
| `spec.trafficPolicy.tls.clientCertificate` | `string` |  |  |  |
| `spec.trafficPolicy.tls.privateKey` | `string` |  |  |  |
| `spec.trafficPolicy.tls.caCertificates` | `string` |  |  |  |
| `spec.trafficPolicy.tls.credentialName` | `string` |  |  |  |
| `spec.trafficPolicy.tls.subjectAltNames` | `[]string` |  |  |  |
| `spec.trafficPolicy.tls.sni` | `string` |  |  |  |
| `spec.trafficPolicy.tls.insecureSkipVerify` | `bool` |  |  |  |
| `spec.trafficPolicy.tls.caCrl` | `string` |  |  |  |
| `spec.trafficPolicy.portLevelSettings` | `[]KubernetesDestinationRulePortTrafficPolicy` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].port` | `KubernetesIstioApiPortSelector` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].port.number` | `uint32` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].loadBalancer` | `KubernetesDestinationRuleLoadBalancerSettings` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].loadBalancer.simple` | `string` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].loadBalancer.consistentHash` | `KubernetesDestinationRuleConsistentHashLb` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.httpHeaderName` | `string` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.httpCookie` | `KubernetesDestinationRuleHttpCookie` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.httpCookie.name` | `string` | yes |  |  |
| `spec.trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.httpCookie.path` | `string` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.httpCookie.ttl` | `string` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.httpCookie.attributes` | `[]KubernetesDestinationRuleHttpCookieAttribute` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.httpCookie.attributes[].name` | `string` | yes |  |  |
| `spec.trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.httpCookie.attributes[].value` | `string` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.useSourceIp` | `bool` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.httpQueryParameterName` | `string` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.ringHash` | `KubernetesDestinationRuleRingHash` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.ringHash.minimumRingSize` | `uint64` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.maglev` | `KubernetesDestinationRuleMagLev` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.maglev.tableSize` | `uint64` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.minimumRingSize` | `uint64` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].loadBalancer.localityLbSetting` | `KubernetesDestinationRuleLocalityLbSetting` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].loadBalancer.localityLbSetting.distribute` | `[]KubernetesDestinationRuleLocalityDistribute` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].loadBalancer.localityLbSetting.distribute[].from` | `string` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].loadBalancer.localityLbSetting.distribute[].to` | `map<string, uint32>` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].loadBalancer.localityLbSetting.failover` | `[]KubernetesDestinationRuleLocalityFailover` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].loadBalancer.localityLbSetting.failover[].from` | `string` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].loadBalancer.localityLbSetting.failover[].to` | `string` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].loadBalancer.localityLbSetting.failoverPriority` | `[]string` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].loadBalancer.localityLbSetting.enabled` | `bool` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].loadBalancer.warmupDurationSecs` | `string` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].loadBalancer.warmup` | `KubernetesDestinationRuleWarmupConfiguration` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].loadBalancer.warmup.duration` | `string` | yes |  |  |
| `spec.trafficPolicy.portLevelSettings[].loadBalancer.warmup.minimumPercent` | `double` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].loadBalancer.warmup.aggression` | `double` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].connectionPool` | `KubernetesDestinationRuleConnectionPoolSettings` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].connectionPool.tcp` | `KubernetesDestinationRuleTcpSettings` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].connectionPool.tcp.maxConnections` | `int32` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].connectionPool.tcp.connectTimeout` | `string` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].connectionPool.tcp.tcpKeepalive` | `KubernetesDestinationRuleTcpKeepalive` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].connectionPool.tcp.tcpKeepalive.probes` | `uint32` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].connectionPool.tcp.tcpKeepalive.time` | `string` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].connectionPool.tcp.tcpKeepalive.interval` | `string` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].connectionPool.tcp.maxConnectionDuration` | `string` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].connectionPool.tcp.idleTimeout` | `string` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].connectionPool.http` | `KubernetesDestinationRuleHttpSettings` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].connectionPool.http.http1MaxPendingRequests` | `int32` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].connectionPool.http.http2MaxRequests` | `int32` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].connectionPool.http.maxRequestsPerConnection` | `int32` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].connectionPool.http.maxRetries` | `int32` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].connectionPool.http.idleTimeout` | `string` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].connectionPool.http.h2UpgradePolicy` | `string` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].connectionPool.http.useClientProtocol` | `bool` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].connectionPool.http.maxConcurrentStreams` | `int32` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].outlierDetection` | `KubernetesDestinationRuleOutlierDetection` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].outlierDetection.splitExternalLocalOriginErrors` | `bool` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].outlierDetection.consecutiveLocalOriginFailures` | `uint32` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].outlierDetection.consecutiveGatewayErrors` | `uint32` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].outlierDetection.consecutive5xxErrors` | `uint32` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].outlierDetection.interval` | `string` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].outlierDetection.baseEjectionTime` | `string` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].outlierDetection.maxEjectionPercent` | `int32` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].outlierDetection.minHealthPercent` | `int32` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].tls` | `KubernetesDestinationRuleClientTlsSettings` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].tls.mode` | `string` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].tls.clientCertificate` | `string` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].tls.privateKey` | `string` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].tls.caCertificates` | `string` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].tls.credentialName` | `string` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].tls.subjectAltNames` | `[]string` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].tls.sni` | `string` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].tls.insecureSkipVerify` | `bool` |  |  |  |
| `spec.trafficPolicy.portLevelSettings[].tls.caCrl` | `string` |  |  |  |
| `spec.trafficPolicy.tunnel` | `KubernetesDestinationRuleTunnelSettings` |  |  |  |
| `spec.trafficPolicy.tunnel.protocol` | `string` |  |  |  |
| `spec.trafficPolicy.tunnel.targetHost` | `string` | yes |  |  |
| `spec.trafficPolicy.tunnel.targetPort` | `uint32` | yes |  |  |
| `spec.trafficPolicy.proxyProtocol` | `KubernetesDestinationRuleProxyProtocol` |  |  |  |
| `spec.trafficPolicy.proxyProtocol.version` | `string` |  |  |  |
| `spec.trafficPolicy.retryBudget` | `KubernetesDestinationRuleRetryBudget` |  |  |  |
| `spec.trafficPolicy.retryBudget.percent` | `double` |  |  |  |
| `spec.trafficPolicy.retryBudget.minRetryConcurrency` | `uint32` |  |  |  |
| `spec.subsets` | `[]KubernetesDestinationRuleSubset` |  |  |  |
| `spec.subsets[].name` | `string` | yes |  |  |
| `spec.subsets[].labels` | `map<string, string>` |  |  |  |
| `spec.subsets[].trafficPolicy` | `KubernetesDestinationRuleTrafficPolicy` |  |  |  |
| `spec.subsets[].trafficPolicy.loadBalancer` | `KubernetesDestinationRuleLoadBalancerSettings` |  |  |  |
| `spec.subsets[].trafficPolicy.loadBalancer.simple` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.loadBalancer.consistentHash` | `KubernetesDestinationRuleConsistentHashLb` |  |  |  |
| `spec.subsets[].trafficPolicy.loadBalancer.consistentHash.httpHeaderName` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.loadBalancer.consistentHash.httpCookie` | `KubernetesDestinationRuleHttpCookie` |  |  |  |
| `spec.subsets[].trafficPolicy.loadBalancer.consistentHash.httpCookie.name` | `string` | yes |  |  |
| `spec.subsets[].trafficPolicy.loadBalancer.consistentHash.httpCookie.path` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.loadBalancer.consistentHash.httpCookie.ttl` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.loadBalancer.consistentHash.httpCookie.attributes` | `[]KubernetesDestinationRuleHttpCookieAttribute` |  |  |  |
| `spec.subsets[].trafficPolicy.loadBalancer.consistentHash.httpCookie.attributes[].name` | `string` | yes |  |  |
| `spec.subsets[].trafficPolicy.loadBalancer.consistentHash.httpCookie.attributes[].value` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.loadBalancer.consistentHash.useSourceIp` | `bool` |  |  |  |
| `spec.subsets[].trafficPolicy.loadBalancer.consistentHash.httpQueryParameterName` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.loadBalancer.consistentHash.ringHash` | `KubernetesDestinationRuleRingHash` |  |  |  |
| `spec.subsets[].trafficPolicy.loadBalancer.consistentHash.ringHash.minimumRingSize` | `uint64` |  |  |  |
| `spec.subsets[].trafficPolicy.loadBalancer.consistentHash.maglev` | `KubernetesDestinationRuleMagLev` |  |  |  |
| `spec.subsets[].trafficPolicy.loadBalancer.consistentHash.maglev.tableSize` | `uint64` |  |  |  |
| `spec.subsets[].trafficPolicy.loadBalancer.consistentHash.minimumRingSize` | `uint64` |  |  |  |
| `spec.subsets[].trafficPolicy.loadBalancer.localityLbSetting` | `KubernetesDestinationRuleLocalityLbSetting` |  |  |  |
| `spec.subsets[].trafficPolicy.loadBalancer.localityLbSetting.distribute` | `[]KubernetesDestinationRuleLocalityDistribute` |  |  |  |
| `spec.subsets[].trafficPolicy.loadBalancer.localityLbSetting.distribute[].from` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.loadBalancer.localityLbSetting.distribute[].to` | `map<string, uint32>` |  |  |  |
| `spec.subsets[].trafficPolicy.loadBalancer.localityLbSetting.failover` | `[]KubernetesDestinationRuleLocalityFailover` |  |  |  |
| `spec.subsets[].trafficPolicy.loadBalancer.localityLbSetting.failover[].from` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.loadBalancer.localityLbSetting.failover[].to` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.loadBalancer.localityLbSetting.failoverPriority` | `[]string` |  |  |  |
| `spec.subsets[].trafficPolicy.loadBalancer.localityLbSetting.enabled` | `bool` |  |  |  |
| `spec.subsets[].trafficPolicy.loadBalancer.warmupDurationSecs` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.loadBalancer.warmup` | `KubernetesDestinationRuleWarmupConfiguration` |  |  |  |
| `spec.subsets[].trafficPolicy.loadBalancer.warmup.duration` | `string` | yes |  |  |
| `spec.subsets[].trafficPolicy.loadBalancer.warmup.minimumPercent` | `double` |  |  |  |
| `spec.subsets[].trafficPolicy.loadBalancer.warmup.aggression` | `double` |  |  |  |
| `spec.subsets[].trafficPolicy.connectionPool` | `KubernetesDestinationRuleConnectionPoolSettings` |  |  |  |
| `spec.subsets[].trafficPolicy.connectionPool.tcp` | `KubernetesDestinationRuleTcpSettings` |  |  |  |
| `spec.subsets[].trafficPolicy.connectionPool.tcp.maxConnections` | `int32` |  |  |  |
| `spec.subsets[].trafficPolicy.connectionPool.tcp.connectTimeout` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.connectionPool.tcp.tcpKeepalive` | `KubernetesDestinationRuleTcpKeepalive` |  |  |  |
| `spec.subsets[].trafficPolicy.connectionPool.tcp.tcpKeepalive.probes` | `uint32` |  |  |  |
| `spec.subsets[].trafficPolicy.connectionPool.tcp.tcpKeepalive.time` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.connectionPool.tcp.tcpKeepalive.interval` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.connectionPool.tcp.maxConnectionDuration` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.connectionPool.tcp.idleTimeout` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.connectionPool.http` | `KubernetesDestinationRuleHttpSettings` |  |  |  |
| `spec.subsets[].trafficPolicy.connectionPool.http.http1MaxPendingRequests` | `int32` |  |  |  |
| `spec.subsets[].trafficPolicy.connectionPool.http.http2MaxRequests` | `int32` |  |  |  |
| `spec.subsets[].trafficPolicy.connectionPool.http.maxRequestsPerConnection` | `int32` |  |  |  |
| `spec.subsets[].trafficPolicy.connectionPool.http.maxRetries` | `int32` |  |  |  |
| `spec.subsets[].trafficPolicy.connectionPool.http.idleTimeout` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.connectionPool.http.h2UpgradePolicy` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.connectionPool.http.useClientProtocol` | `bool` |  |  |  |
| `spec.subsets[].trafficPolicy.connectionPool.http.maxConcurrentStreams` | `int32` |  |  |  |
| `spec.subsets[].trafficPolicy.outlierDetection` | `KubernetesDestinationRuleOutlierDetection` |  |  |  |
| `spec.subsets[].trafficPolicy.outlierDetection.splitExternalLocalOriginErrors` | `bool` |  |  |  |
| `spec.subsets[].trafficPolicy.outlierDetection.consecutiveLocalOriginFailures` | `uint32` |  |  |  |
| `spec.subsets[].trafficPolicy.outlierDetection.consecutiveGatewayErrors` | `uint32` |  |  |  |
| `spec.subsets[].trafficPolicy.outlierDetection.consecutive5xxErrors` | `uint32` |  |  |  |
| `spec.subsets[].trafficPolicy.outlierDetection.interval` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.outlierDetection.baseEjectionTime` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.outlierDetection.maxEjectionPercent` | `int32` |  |  |  |
| `spec.subsets[].trafficPolicy.outlierDetection.minHealthPercent` | `int32` |  |  |  |
| `spec.subsets[].trafficPolicy.tls` | `KubernetesDestinationRuleClientTlsSettings` |  |  |  |
| `spec.subsets[].trafficPolicy.tls.mode` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.tls.clientCertificate` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.tls.privateKey` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.tls.caCertificates` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.tls.credentialName` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.tls.subjectAltNames` | `[]string` |  |  |  |
| `spec.subsets[].trafficPolicy.tls.sni` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.tls.insecureSkipVerify` | `bool` |  |  |  |
| `spec.subsets[].trafficPolicy.tls.caCrl` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings` | `[]KubernetesDestinationRulePortTrafficPolicy` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].port` | `KubernetesIstioApiPortSelector` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].port.number` | `uint32` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer` | `KubernetesDestinationRuleLoadBalancerSettings` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.simple` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.consistentHash` | `KubernetesDestinationRuleConsistentHashLb` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.httpHeaderName` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.httpCookie` | `KubernetesDestinationRuleHttpCookie` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.httpCookie.name` | `string` | yes |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.httpCookie.path` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.httpCookie.ttl` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.httpCookie.attributes` | `[]KubernetesDestinationRuleHttpCookieAttribute` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.httpCookie.attributes[].name` | `string` | yes |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.httpCookie.attributes[].value` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.useSourceIp` | `bool` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.httpQueryParameterName` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.ringHash` | `KubernetesDestinationRuleRingHash` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.ringHash.minimumRingSize` | `uint64` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.maglev` | `KubernetesDestinationRuleMagLev` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.maglev.tableSize` | `uint64` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.minimumRingSize` | `uint64` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.localityLbSetting` | `KubernetesDestinationRuleLocalityLbSetting` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.localityLbSetting.distribute` | `[]KubernetesDestinationRuleLocalityDistribute` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.localityLbSetting.distribute[].from` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.localityLbSetting.distribute[].to` | `map<string, uint32>` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.localityLbSetting.failover` | `[]KubernetesDestinationRuleLocalityFailover` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.localityLbSetting.failover[].from` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.localityLbSetting.failover[].to` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.localityLbSetting.failoverPriority` | `[]string` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.localityLbSetting.enabled` | `bool` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.warmupDurationSecs` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.warmup` | `KubernetesDestinationRuleWarmupConfiguration` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.warmup.duration` | `string` | yes |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.warmup.minimumPercent` | `double` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.warmup.aggression` | `double` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].connectionPool` | `KubernetesDestinationRuleConnectionPoolSettings` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].connectionPool.tcp` | `KubernetesDestinationRuleTcpSettings` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].connectionPool.tcp.maxConnections` | `int32` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].connectionPool.tcp.connectTimeout` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].connectionPool.tcp.tcpKeepalive` | `KubernetesDestinationRuleTcpKeepalive` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].connectionPool.tcp.tcpKeepalive.probes` | `uint32` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].connectionPool.tcp.tcpKeepalive.time` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].connectionPool.tcp.tcpKeepalive.interval` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].connectionPool.tcp.maxConnectionDuration` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].connectionPool.tcp.idleTimeout` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].connectionPool.http` | `KubernetesDestinationRuleHttpSettings` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].connectionPool.http.http1MaxPendingRequests` | `int32` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].connectionPool.http.http2MaxRequests` | `int32` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].connectionPool.http.maxRequestsPerConnection` | `int32` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].connectionPool.http.maxRetries` | `int32` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].connectionPool.http.idleTimeout` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].connectionPool.http.h2UpgradePolicy` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].connectionPool.http.useClientProtocol` | `bool` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].connectionPool.http.maxConcurrentStreams` | `int32` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].outlierDetection` | `KubernetesDestinationRuleOutlierDetection` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].outlierDetection.splitExternalLocalOriginErrors` | `bool` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].outlierDetection.consecutiveLocalOriginFailures` | `uint32` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].outlierDetection.consecutiveGatewayErrors` | `uint32` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].outlierDetection.consecutive5xxErrors` | `uint32` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].outlierDetection.interval` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].outlierDetection.baseEjectionTime` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].outlierDetection.maxEjectionPercent` | `int32` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].outlierDetection.minHealthPercent` | `int32` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].tls` | `KubernetesDestinationRuleClientTlsSettings` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].tls.mode` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].tls.clientCertificate` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].tls.privateKey` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].tls.caCertificates` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].tls.credentialName` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].tls.subjectAltNames` | `[]string` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].tls.sni` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].tls.insecureSkipVerify` | `bool` |  |  |  |
| `spec.subsets[].trafficPolicy.portLevelSettings[].tls.caCrl` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.tunnel` | `KubernetesDestinationRuleTunnelSettings` |  |  |  |
| `spec.subsets[].trafficPolicy.tunnel.protocol` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.tunnel.targetHost` | `string` | yes |  |  |
| `spec.subsets[].trafficPolicy.tunnel.targetPort` | `uint32` | yes |  |  |
| `spec.subsets[].trafficPolicy.proxyProtocol` | `KubernetesDestinationRuleProxyProtocol` |  |  |  |
| `spec.subsets[].trafficPolicy.proxyProtocol.version` | `string` |  |  |  |
| `spec.subsets[].trafficPolicy.retryBudget` | `KubernetesDestinationRuleRetryBudget` |  |  |  |
| `spec.subsets[].trafficPolicy.retryBudget.percent` | `double` |  |  |  |
| `spec.subsets[].trafficPolicy.retryBudget.minRetryConcurrency` | `uint32` |  |  |  |
| `spec.exportTo` | `[]string` |  |  |  |
| `spec.workloadSelector` | `KubernetesIstioApiWorkloadSelector` |  |  |  |
| `spec.workloadSelector.matchLabels` | `map<string, string>` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace in which the DestinationRule is created. For short host names, istiod
interprets the host relative to THIS namespace, not the target service's namespace —
prefer fully-qualified hosts to avoid ambiguity. `export_to` controls cross-namespace
visibility.

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.host

`string | valueFrom` · required

The name of a service from the service registry that this rule applies to. Looked up
from the platform's registry (Kubernetes services, etc.) and from hosts declared by
ServiceEntries; rules for unknown services are ignored. Applies to both HTTP and TCP.
Required. Prefer a fully-qualified name (e.g. `reviews.prod.svc.cluster.local`) over a
short name, which istiod resolves relative to this rule's namespace.

INFRA-CHART COMPOSABILITY: host is a foreign key defaulting to a KubernetesService
reference that resolves to the Service's in-cluster FQDN
(`<name>.<namespace>.svc.cluster.local`) — wiring it with valueFrom orders this
DestinationRule after the Service whose traffic it shapes and can never drift from
the Service's actual name. For hosts that are not Planton-managed Services (a
ServiceEntry host, an external FQDN, a wildcard), pass the literal host with
`value:` — istiod resolves it against the service registry at runtime either way.

- references: KubernetesService (`status.outputs.kube_endpoint`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesService, name: <that resource's name>, fieldPath: status.outputs.kube_endpoint}} -- a bare string does not parse

### spec.trafficPolicy

`KubernetesDestinationRuleTrafficPolicy`

Traffic policies (load balancing, connection pool sizes, outlier detection, TLS) applied
to traffic destined for the host, across all ports unless overridden per port.

### spec.trafficPolicy.loadBalancer

`KubernetesDestinationRuleLoadBalancerSettings`

Settings controlling the load balancer algorithm.

- rule: at most one of simple or consistent_hash may be set
- rule: at most one of warmup or warmup_duration_secs may be set

### spec.trafficPolicy.loadBalancer.simple

`string` · optional (explicit presence)

Standard load balancing algorithm requiring no tuning. One of:
  LEAST_CONN   — deprecated; use LEAST_REQUEST instead.
  RANDOM       — selects a random healthy host.
  PASSTHROUGH  — forwards to the original requested IP without load balancing.
  ROUND_ROBIN  — basic round robin; prefer LEAST_REQUEST in most cases.
  LEAST_REQUEST — favors hosts with the fewest outstanding requests (recommended).
Mutually exclusive with consistent_hash. // external standard exception -- matches Istio SimpleLB

- rule: {"string":{"in":["LEAST_CONN","RANDOM","PASSTHROUGH","ROUND_ROBIN","LEAST_REQUEST"]}}

### spec.trafficPolicy.loadBalancer.consistentHash

`KubernetesDestinationRuleConsistentHashLb`

Consistent-hash (soft session affinity) load balancing based on a request property.
Mutually exclusive with simple.

- rule: at most one of http_header_name, http_cookie, use_source_ip, or http_query_parameter_name may be set
- rule: at most one of ring_hash or maglev may be set

### spec.trafficPolicy.loadBalancer.consistentHash.httpHeaderName

`string` · optional (explicit presence)

Hash based on a specific HTTP header.

### spec.trafficPolicy.loadBalancer.consistentHash.httpCookie

`KubernetesDestinationRuleHttpCookie`

Hash based on an HTTP cookie (generated if absent).

### spec.trafficPolicy.loadBalancer.consistentHash.httpCookie.name

`string` · required

Name of the cookie. Required.

- rule: {"required":true}

### spec.trafficPolicy.loadBalancer.consistentHash.httpCookie.path

`string` · optional (explicit presence)

Path to set on the generated cookie.

### spec.trafficPolicy.loadBalancer.consistentHash.httpCookie.ttl

`string` · optional (explicit presence)

Lifetime of the cookie. If present and zero, the generated cookie is a session cookie.
Modeled as a duration string; upstream applies no minimum (duration-validation:
none), so only valid, non-negative durations are enforced.

- rule: ttl must be a valid, non-negative duration (e.g. "0s", "1h")

### spec.trafficPolicy.loadBalancer.consistentHash.httpCookie.attributes

`[]KubernetesDestinationRuleHttpCookieAttribute`

Additional attributes set on the generated cookie (e.g. SameSite=Strict, Secure).
Each attribute has a required name and an optional value.

### spec.trafficPolicy.loadBalancer.consistentHash.httpCookie.attributes[].name

`string` · required

The name of the cookie attribute. Required.

- rule: {"required":true}

### spec.trafficPolicy.loadBalancer.consistentHash.httpCookie.attributes[].value

`string` · optional (explicit presence)

The optional value of the cookie attribute (an attribute like `Secure` carries none).

### spec.trafficPolicy.loadBalancer.consistentHash.useSourceIp

`bool` · optional (explicit presence)

Hash based on the source IP address. Applicable to both TCP and HTTP connections.

### spec.trafficPolicy.loadBalancer.consistentHash.httpQueryParameterName

`string` · optional (explicit presence)

Hash based on a specific HTTP query parameter.

### spec.trafficPolicy.loadBalancer.consistentHash.ringHash

`KubernetesDestinationRuleRingHash`

The ring/modulo (ring hash) algorithm. Mutually exclusive with maglev.

### spec.trafficPolicy.loadBalancer.consistentHash.ringHash.minimumRingSize

`uint64` · optional (explicit presence)

Minimum number of virtual nodes for the hash ring (default 1024). Larger rings give more
granular distribution.

### spec.trafficPolicy.loadBalancer.consistentHash.maglev

`KubernetesDestinationRuleMagLev`

The Maglev algorithm. Mutually exclusive with ring_hash.

### spec.trafficPolicy.loadBalancer.consistentHash.maglev.tableSize

`uint64` · optional (explicit presence)

Table size for Maglev hashing (default 65537). Should be a prime number less than
5000011; a larger table reduces disruption when backends change.

### spec.trafficPolicy.loadBalancer.consistentHash.minimumRingSize

`uint64` · optional (explicit presence)

Deprecated: use `ring_hash.minimum_ring_size` instead. Minimum number of virtual nodes
in the hash ring.

### spec.trafficPolicy.loadBalancer.localityLbSetting

`KubernetesDestinationRuleLocalityLbSetting`

Locality-weighted load balancing settings; overrides mesh-wide locality settings in
their entirety (no merging).

- rule: at most one of distribute, failover, or failover_priority may be set

### spec.trafficPolicy.loadBalancer.localityLbSetting.distribute

`[]KubernetesDestinationRuleLocalityDistribute`

Explicit per-zone traffic distribution weights. Weights for a given `from` should sum to
100; any locality not listed receives no traffic.

### spec.trafficPolicy.loadBalancer.localityLbSetting.distribute[].from

`string` · optional (explicit presence)

Originating locality, '/' separated (e.g. `region/zone/sub_zone`). Terminal wildcards are
allowed on any segment (e.g. `us-west/*`).

### spec.trafficPolicy.loadBalancer.localityLbSetting.distribute[].to

`map<string, uint32>`

Map of upstream localities to traffic-distribution weights. The weights should sum to 100.

### spec.trafficPolicy.loadBalancer.localityLbSetting.failover

`[]KubernetesDestinationRuleLocalityFailover`

Region-level failover policy. Zone/sub-zone failover is automatic; specify this only to
constrain cross-region failover. Should be used with outlier detection.

### spec.trafficPolicy.loadBalancer.localityLbSetting.failover[].from

`string` · optional (explicit presence)

Originating region.

### spec.trafficPolicy.loadBalancer.localityLbSetting.failover[].to

`string` · optional (explicit presence)

Destination region traffic fails over to when endpoints in the `from` region become
unhealthy.

### spec.trafficPolicy.loadBalancer.localityLbSetting.failoverPriority

`[]string`

Ordered list of labels used to sort endpoints into priority tiers for failover. Entries
may be bare keys (`key`) or key=value pairs; special topology labels are supported. Used
with outlier detection.

### spec.trafficPolicy.loadBalancer.localityLbSetting.enabled

`bool` · optional (explicit presence)

Enable (or disable) locality load balancing for this DestinationRule, overriding mesh-wide
settings entirely.

### spec.trafficPolicy.loadBalancer.warmupDurationSecs

`string` · optional (explicit presence)

Deprecated: use `warmup` instead. Duration over which a newly added endpoint is ramped
up. Modeled as a duration string.

- rule: warmup_duration_secs must be a valid, non-negative duration (e.g. "30s")

### spec.trafficPolicy.loadBalancer.warmup

`KubernetesDestinationRuleWarmupConfiguration`

Warmup configuration: newly created endpoints receive progressively increasing traffic
over the configured window. Only effective for ROUND_ROBIN and LEAST_REQUEST.

### spec.trafficPolicy.loadBalancer.warmup.duration

`string` · required

Duration of warmup mode. Required. Modeled as a duration string.

- rule: duration must be a valid, non-negative duration (e.g. "30s")
- rule: {"required":true}

### spec.trafficPolicy.loadBalancer.warmup.minimumPercent

`double` · optional (explicit presence)

Minimum percentage of origin weight at the start of warmup. If unspecified, defaults to
10. Upstream bounds: 0–100.

- rule: {"double":{"lte":100,"gte":0}}

### spec.trafficPolicy.loadBalancer.warmup.aggression

`double` · optional (explicit presence)

Controls the speed of traffic increase over the warmup window. Defaults to 1.0 (linear);
higher values ramp up non-linearly. Upstream minimum: 1.

- rule: {"double":{"gte":1}}

### spec.trafficPolicy.connectionPool

`KubernetesDestinationRuleConnectionPoolSettings`

Settings controlling the volume of connections to an upstream service.

### spec.trafficPolicy.connectionPool.tcp

`KubernetesDestinationRuleTcpSettings`

Settings common to both HTTP and TCP upstream connections.

### spec.trafficPolicy.connectionPool.tcp.maxConnections

`int32` · optional (explicit presence)

Maximum number of HTTP1/TCP connections to a destination host. Default 2^32-1.

### spec.trafficPolicy.connectionPool.tcp.connectTimeout

`string` · optional (explicit presence)

TCP connection timeout (format: 1h/1m/1s/1ms; must be >=1ms; default 10s). Modeled as a
duration string.

- rule: connect_timeout must be a valid duration of at least 1ms (e.g. "30ms")

### spec.trafficPolicy.connectionPool.tcp.tcpKeepalive

`KubernetesDestinationRuleTcpKeepalive`

TCP keepalive settings (enables SO_KEEPALIVE on the socket).

### spec.trafficPolicy.connectionPool.tcp.tcpKeepalive.probes

`uint32` · optional (explicit presence)

Maximum number of keepalive probes to send before deciding the connection is dead.
Default: OS level (Linux 9).

### spec.trafficPolicy.connectionPool.tcp.tcpKeepalive.time

`string` · optional (explicit presence)

Idle time before keepalive probes start. Default: OS level (Linux 7200s). Modeled as a
duration string.

- rule: time must be a valid, non-negative duration (e.g. "7200s")

### spec.trafficPolicy.connectionPool.tcp.tcpKeepalive.interval

`string` · optional (explicit presence)

Interval between keepalive probes. Default: OS level (Linux 75s). Modeled as a duration
string.

- rule: interval must be a valid, non-negative duration (e.g. "75s")

### spec.trafficPolicy.connectionPool.tcp.maxConnectionDuration

`string` · optional (explicit presence)

Maximum duration of a connection, measured from when it was established. If unset, no
maximum. Must be >=1ms. Modeled as a duration string.

- rule: max_connection_duration must be a valid duration of at least 1ms (e.g. "10s")

### spec.trafficPolicy.connectionPool.tcp.idleTimeout

`string` · optional (explicit presence)

Idle timeout for TCP connections (no bytes sent/received on either side). Default 1h;
0s disables. Modeled as a duration string; upstream applies no minimum.

- rule: idle_timeout must be a valid, non-negative duration (e.g. "0s", "1h")

### spec.trafficPolicy.connectionPool.http

`KubernetesDestinationRuleHttpSettings`

HTTP-specific connection pool settings (HTTP1.1/HTTP2/gRPC).

### spec.trafficPolicy.connectionPool.http.http1MaxPendingRequests

`int32` · optional (explicit presence)

Maximum requests queued while waiting for a ready connection. Default 2^32-1. Applies to
both HTTP/1.1 and HTTP2.

### spec.trafficPolicy.connectionPool.http.http2MaxRequests

`int32` · optional (explicit presence)

Maximum active requests to a destination. Default 2^32-1. Applies to both HTTP/1.1 and
HTTP2.

### spec.trafficPolicy.connectionPool.http.maxRequestsPerConnection

`int32` · optional (explicit presence)

Maximum requests per backend connection. 1 disables keep-alive. Default 0 ("unlimited",
up to 2^29).

### spec.trafficPolicy.connectionPool.http.maxRetries

`int32` · optional (explicit presence)

Maximum retries outstanding to all hosts in a cluster at a time. Default 2^32-1.

### spec.trafficPolicy.connectionPool.http.idleTimeout

`string` · optional (explicit presence)

Idle timeout for upstream pool connections (no active requests). Default 1h. Modeled as
a duration string.

- rule: idle_timeout must be a valid, non-negative duration (e.g. "1h")

### spec.trafficPolicy.connectionPool.http.h2UpgradePolicy

`string` · optional (explicit presence)

Policy for upgrading HTTP1.1 connections to HTTP2 for this destination. One of:
  DEFAULT        — use the global default.
  DO_NOT_UPGRADE — never upgrade (overrides the default).
  UPGRADE        — always upgrade (overrides the default).
external standard exception -- matches Istio HTTPSettings.H2UpgradePolicy

- rule: {"string":{"in":["DEFAULT","DO_NOT_UPGRADE","UPGRADE"]}}

### spec.trafficPolicy.connectionPool.http.useClientProtocol

`bool` · optional (explicit presence)

If true, the client protocol is preserved when connecting to the backend. When true,
h2_upgrade_policy is ineffective.

### spec.trafficPolicy.connectionPool.http.maxConcurrentStreams

`int32` · optional (explicit presence)

Maximum concurrent streams allowed for a peer on one HTTP/2 connection. Default 2^31-1.

### spec.trafficPolicy.outlierDetection

`KubernetesDestinationRuleOutlierDetection`

Settings controlling eviction of unhealthy hosts from the load balancing pool.

### spec.trafficPolicy.outlierDetection.splitExternalLocalOriginErrors

`bool` · optional (explicit presence)

If true, locally-originated failures (connect failures, timeouts) are counted via
consecutive_local_origin_failures rather than upstream response codes. Default false.

### spec.trafficPolicy.outlierDetection.consecutiveLocalOriginFailures

`uint32` · optional (explicit presence)

Number of consecutive locally-originated failures before ejection. Default 5. Only takes
effect when split_external_local_origin_errors is true.

### spec.trafficPolicy.outlierDetection.consecutiveGatewayErrors

`uint32` · optional (explicit presence)

Number of consecutive gateway errors (502/503/504, or TCP connect failures) before
ejection. Disabled by default / when 0. Counted within consecutive_5xx_errors.

### spec.trafficPolicy.outlierDetection.consecutive5xxErrors

`uint32` · optional (explicit presence)

Number of consecutive 5xx errors before ejection. Defaults to 5; disable by setting 0.

### spec.trafficPolicy.outlierDetection.interval

`string` · optional (explicit presence)

Time interval between ejection sweep analyses (format 1h/1m/1s/1ms; must be >=1ms;
default 10s). Modeled as a duration string.

- rule: interval must be a valid duration of at least 1ms (e.g. "5m")

### spec.trafficPolicy.outlierDetection.baseEjectionTime

`string` · optional (explicit presence)

Minimum ejection duration; a host stays ejected for this value times the number of times
it has been ejected (must be >=1ms; default 30s). Modeled as a duration string.

- rule: base_ejection_time must be a valid duration of at least 1ms (e.g. "15m")

### spec.trafficPolicy.outlierDetection.maxEjectionPercent

`int32` · optional (explicit presence)

Maximum percentage of hosts in the pool that can be ejected. Default 10%.

### spec.trafficPolicy.outlierDetection.minHealthPercent

`int32` · optional (explicit presence)

Outlier detection stays enabled only while at least this percentage of hosts is healthy;
below it, all hosts (healthy and unhealthy) are load balanced. Default 0% (disabled).

### spec.trafficPolicy.tls

`KubernetesDestinationRuleClientTlsSettings`

TLS settings for connections the sidecar originates to the upstream service.

### spec.trafficPolicy.tls.mode

`string` · optional (explicit presence)

TLS mode for connections to the upstream. One of:
  DISABLE      — no TLS.
  SIMPLE       — originate TLS (one-way).
  MUTUAL       — mutual TLS presenting client certs (client_certificate + private_key,
                 or credential_name, required).
  ISTIO_MUTUAL — mutual TLS using Istio-generated certs; all other fields should be empty.
external standard exception -- matches Istio ClientTLSSettings.TLSmode

- rule: {"string":{"in":["DISABLE","SIMPLE","MUTUAL","ISTIO_MUTUAL"]}}

### spec.trafficPolicy.tls.clientCertificate

`string` · optional (explicit presence)

Path to the client-side TLS certificate file. Required for MUTUAL; empty for ISTIO_MUTUAL.

### spec.trafficPolicy.tls.privateKey

`string` · optional (explicit presence)

Path to the client's private key file. Required for MUTUAL; empty for ISTIO_MUTUAL.

### spec.trafficPolicy.tls.caCertificates

`string` · optional (explicit presence)

Path to the CA certificate file used to verify the server certificate. Empty for
ISTIO_MUTUAL; if omitted, the OS CA certificates are used.

### spec.trafficPolicy.tls.credentialName

`string` · optional (explicit presence)

Name of the Kubernetes secret (in the proxy's namespace) holding the client TLS certs
(and CA). Only one of (client_certificate + private_key + ca_certificates) OR
credential_name may be specified. At sidecars this is honored only when a
workload_selector is set; otherwise it applies at gateways only.

INFRA-CHART COMPOSABILITY: credential_name is a PLAIN reference to a Secret /
certificate by name, resolved by istiod/Envoy at runtime — NOT an Planton foreign key
(StringValueOrRef). It creates NO automatic DAG edge. To order this DestinationRule after
the secret it consumes in an infra chart, an author MUST express the dependency via
metadata.relationships (`uses` -> KubernetesCertificate / KubernetesSecret), e.g.:
  metadata:
    relationships:
      - kind: KubernetesSecret
        name: "{{ values.client_credential }}"
        type: uses
Kept plain (not StringValueOrRef) for fidelity and consistency with the sibling Gateway
family's `certificate_refs`. See the "Composing in Infra Charts" docs.

### spec.trafficPolicy.tls.subjectAltNames

`[]string`

Alternate names to verify against the server certificate's subject alt names. Overrides
the ServiceEntry's subjectAltNames. If unset, validation is based on the downstream
host/authority header.

### spec.trafficPolicy.tls.sni

`string` · optional (explicit presence)

SNI string presented to the server during the TLS handshake. If unset, derived from the
downstream host/authority header for SIMPLE and MUTUAL modes.

### spec.trafficPolicy.tls.insecureSkipVerify

`bool` · optional (explicit presence)

If true, the proxy skips verifying the CA signature and SAN of the server certificate.
Default false.

### spec.trafficPolicy.tls.caCrl

`string` · optional (explicit presence)

Path to the certificate revocation list (CRL) file used to verify the server cert. If
credential_name is set, the CRL must be supplied inside the credential, not here.

### spec.trafficPolicy.portLevelSettings

`[]KubernetesDestinationRulePortTrafficPolicy`

Traffic policies specific to individual ports. Port-level settings fully override the
destination-level settings for that port (no field-level inheritance). Upstream allows
up to 4096.

- rule: {"repeated":{"maxItems":"4096"}}

### spec.trafficPolicy.portLevelSettings[].port

`KubernetesIstioApiPortSelector`

The destination service port these settings apply to. Reuses the shared
istio.type.v1beta1 PortSelector (a single `number`).

### spec.trafficPolicy.portLevelSettings[].port.number

`uint32`

Port number (1-65535).

- rule: {"uint32":{"lte":65535,"gte":1}}

### spec.trafficPolicy.portLevelSettings[].loadBalancer

`KubernetesDestinationRuleLoadBalancerSettings`

Settings controlling the load balancer algorithm for this port.

- rule: at most one of simple or consistent_hash may be set
- rule: at most one of warmup or warmup_duration_secs may be set

### spec.trafficPolicy.portLevelSettings[].loadBalancer.simple

`string` · optional (explicit presence)

Standard load balancing algorithm requiring no tuning. One of:
  LEAST_CONN   — deprecated; use LEAST_REQUEST instead.
  RANDOM       — selects a random healthy host.
  PASSTHROUGH  — forwards to the original requested IP without load balancing.
  ROUND_ROBIN  — basic round robin; prefer LEAST_REQUEST in most cases.
  LEAST_REQUEST — favors hosts with the fewest outstanding requests (recommended).
Mutually exclusive with consistent_hash. // external standard exception -- matches Istio SimpleLB

- rule: {"string":{"in":["LEAST_CONN","RANDOM","PASSTHROUGH","ROUND_ROBIN","LEAST_REQUEST"]}}

### spec.trafficPolicy.portLevelSettings[].loadBalancer.consistentHash

`KubernetesDestinationRuleConsistentHashLb`

Consistent-hash (soft session affinity) load balancing based on a request property.
Mutually exclusive with simple.

- rule: at most one of http_header_name, http_cookie, use_source_ip, or http_query_parameter_name may be set
- rule: at most one of ring_hash or maglev may be set

### spec.trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.httpHeaderName

`string` · optional (explicit presence)

Hash based on a specific HTTP header.

### spec.trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.httpCookie

`KubernetesDestinationRuleHttpCookie`

Hash based on an HTTP cookie (generated if absent).

### spec.trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.httpCookie.name

`string` · required

Name of the cookie. Required.

- rule: {"required":true}

### spec.trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.httpCookie.path

`string` · optional (explicit presence)

Path to set on the generated cookie.

### spec.trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.httpCookie.ttl

`string` · optional (explicit presence)

Lifetime of the cookie. If present and zero, the generated cookie is a session cookie.
Modeled as a duration string; upstream applies no minimum (duration-validation:
none), so only valid, non-negative durations are enforced.

- rule: ttl must be a valid, non-negative duration (e.g. "0s", "1h")

### spec.trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.httpCookie.attributes

`[]KubernetesDestinationRuleHttpCookieAttribute`

Additional attributes set on the generated cookie (e.g. SameSite=Strict, Secure).
Each attribute has a required name and an optional value.

### spec.trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.httpCookie.attributes[].name

`string` · required

The name of the cookie attribute. Required.

- rule: {"required":true}

### spec.trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.httpCookie.attributes[].value

`string` · optional (explicit presence)

The optional value of the cookie attribute (an attribute like `Secure` carries none).

### spec.trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.useSourceIp

`bool` · optional (explicit presence)

Hash based on the source IP address. Applicable to both TCP and HTTP connections.

### spec.trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.httpQueryParameterName

`string` · optional (explicit presence)

Hash based on a specific HTTP query parameter.

### spec.trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.ringHash

`KubernetesDestinationRuleRingHash`

The ring/modulo (ring hash) algorithm. Mutually exclusive with maglev.

### spec.trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.ringHash.minimumRingSize

`uint64` · optional (explicit presence)

Minimum number of virtual nodes for the hash ring (default 1024). Larger rings give more
granular distribution.

### spec.trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.maglev

`KubernetesDestinationRuleMagLev`

The Maglev algorithm. Mutually exclusive with ring_hash.

### spec.trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.maglev.tableSize

`uint64` · optional (explicit presence)

Table size for Maglev hashing (default 65537). Should be a prime number less than
5000011; a larger table reduces disruption when backends change.

### spec.trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.minimumRingSize

`uint64` · optional (explicit presence)

Deprecated: use `ring_hash.minimum_ring_size` instead. Minimum number of virtual nodes
in the hash ring.

### spec.trafficPolicy.portLevelSettings[].loadBalancer.localityLbSetting

`KubernetesDestinationRuleLocalityLbSetting`

Locality-weighted load balancing settings; overrides mesh-wide locality settings in
their entirety (no merging).

- rule: at most one of distribute, failover, or failover_priority may be set

### spec.trafficPolicy.portLevelSettings[].loadBalancer.localityLbSetting.distribute

`[]KubernetesDestinationRuleLocalityDistribute`

Explicit per-zone traffic distribution weights. Weights for a given `from` should sum to
100; any locality not listed receives no traffic.

### spec.trafficPolicy.portLevelSettings[].loadBalancer.localityLbSetting.distribute[].from

`string` · optional (explicit presence)

Originating locality, '/' separated (e.g. `region/zone/sub_zone`). Terminal wildcards are
allowed on any segment (e.g. `us-west/*`).

### spec.trafficPolicy.portLevelSettings[].loadBalancer.localityLbSetting.distribute[].to

`map<string, uint32>`

Map of upstream localities to traffic-distribution weights. The weights should sum to 100.

### spec.trafficPolicy.portLevelSettings[].loadBalancer.localityLbSetting.failover

`[]KubernetesDestinationRuleLocalityFailover`

Region-level failover policy. Zone/sub-zone failover is automatic; specify this only to
constrain cross-region failover. Should be used with outlier detection.

### spec.trafficPolicy.portLevelSettings[].loadBalancer.localityLbSetting.failover[].from

`string` · optional (explicit presence)

Originating region.

### spec.trafficPolicy.portLevelSettings[].loadBalancer.localityLbSetting.failover[].to

`string` · optional (explicit presence)

Destination region traffic fails over to when endpoints in the `from` region become
unhealthy.

### spec.trafficPolicy.portLevelSettings[].loadBalancer.localityLbSetting.failoverPriority

`[]string`

Ordered list of labels used to sort endpoints into priority tiers for failover. Entries
may be bare keys (`key`) or key=value pairs; special topology labels are supported. Used
with outlier detection.

### spec.trafficPolicy.portLevelSettings[].loadBalancer.localityLbSetting.enabled

`bool` · optional (explicit presence)

Enable (or disable) locality load balancing for this DestinationRule, overriding mesh-wide
settings entirely.

### spec.trafficPolicy.portLevelSettings[].loadBalancer.warmupDurationSecs

`string` · optional (explicit presence)

Deprecated: use `warmup` instead. Duration over which a newly added endpoint is ramped
up. Modeled as a duration string.

- rule: warmup_duration_secs must be a valid, non-negative duration (e.g. "30s")

### spec.trafficPolicy.portLevelSettings[].loadBalancer.warmup

`KubernetesDestinationRuleWarmupConfiguration`

Warmup configuration: newly created endpoints receive progressively increasing traffic
over the configured window. Only effective for ROUND_ROBIN and LEAST_REQUEST.

### spec.trafficPolicy.portLevelSettings[].loadBalancer.warmup.duration

`string` · required

Duration of warmup mode. Required. Modeled as a duration string.

- rule: duration must be a valid, non-negative duration (e.g. "30s")
- rule: {"required":true}

### spec.trafficPolicy.portLevelSettings[].loadBalancer.warmup.minimumPercent

`double` · optional (explicit presence)

Minimum percentage of origin weight at the start of warmup. If unspecified, defaults to
10. Upstream bounds: 0–100.

- rule: {"double":{"lte":100,"gte":0}}

### spec.trafficPolicy.portLevelSettings[].loadBalancer.warmup.aggression

`double` · optional (explicit presence)

Controls the speed of traffic increase over the warmup window. Defaults to 1.0 (linear);
higher values ramp up non-linearly. Upstream minimum: 1.

- rule: {"double":{"gte":1}}

### spec.trafficPolicy.portLevelSettings[].connectionPool

`KubernetesDestinationRuleConnectionPoolSettings`

Settings controlling the volume of connections to this port.

### spec.trafficPolicy.portLevelSettings[].connectionPool.tcp

`KubernetesDestinationRuleTcpSettings`

Settings common to both HTTP and TCP upstream connections.

### spec.trafficPolicy.portLevelSettings[].connectionPool.tcp.maxConnections

`int32` · optional (explicit presence)

Maximum number of HTTP1/TCP connections to a destination host. Default 2^32-1.

### spec.trafficPolicy.portLevelSettings[].connectionPool.tcp.connectTimeout

`string` · optional (explicit presence)

TCP connection timeout (format: 1h/1m/1s/1ms; must be >=1ms; default 10s). Modeled as a
duration string.

- rule: connect_timeout must be a valid duration of at least 1ms (e.g. "30ms")

### spec.trafficPolicy.portLevelSettings[].connectionPool.tcp.tcpKeepalive

`KubernetesDestinationRuleTcpKeepalive`

TCP keepalive settings (enables SO_KEEPALIVE on the socket).

### spec.trafficPolicy.portLevelSettings[].connectionPool.tcp.tcpKeepalive.probes

`uint32` · optional (explicit presence)

Maximum number of keepalive probes to send before deciding the connection is dead.
Default: OS level (Linux 9).

### spec.trafficPolicy.portLevelSettings[].connectionPool.tcp.tcpKeepalive.time

`string` · optional (explicit presence)

Idle time before keepalive probes start. Default: OS level (Linux 7200s). Modeled as a
duration string.

- rule: time must be a valid, non-negative duration (e.g. "7200s")

### spec.trafficPolicy.portLevelSettings[].connectionPool.tcp.tcpKeepalive.interval

`string` · optional (explicit presence)

Interval between keepalive probes. Default: OS level (Linux 75s). Modeled as a duration
string.

- rule: interval must be a valid, non-negative duration (e.g. "75s")

### spec.trafficPolicy.portLevelSettings[].connectionPool.tcp.maxConnectionDuration

`string` · optional (explicit presence)

Maximum duration of a connection, measured from when it was established. If unset, no
maximum. Must be >=1ms. Modeled as a duration string.

- rule: max_connection_duration must be a valid duration of at least 1ms (e.g. "10s")

### spec.trafficPolicy.portLevelSettings[].connectionPool.tcp.idleTimeout

`string` · optional (explicit presence)

Idle timeout for TCP connections (no bytes sent/received on either side). Default 1h;
0s disables. Modeled as a duration string; upstream applies no minimum.

- rule: idle_timeout must be a valid, non-negative duration (e.g. "0s", "1h")

### spec.trafficPolicy.portLevelSettings[].connectionPool.http

`KubernetesDestinationRuleHttpSettings`

HTTP-specific connection pool settings (HTTP1.1/HTTP2/gRPC).

### spec.trafficPolicy.portLevelSettings[].connectionPool.http.http1MaxPendingRequests

`int32` · optional (explicit presence)

Maximum requests queued while waiting for a ready connection. Default 2^32-1. Applies to
both HTTP/1.1 and HTTP2.

### spec.trafficPolicy.portLevelSettings[].connectionPool.http.http2MaxRequests

`int32` · optional (explicit presence)

Maximum active requests to a destination. Default 2^32-1. Applies to both HTTP/1.1 and
HTTP2.

### spec.trafficPolicy.portLevelSettings[].connectionPool.http.maxRequestsPerConnection

`int32` · optional (explicit presence)

Maximum requests per backend connection. 1 disables keep-alive. Default 0 ("unlimited",
up to 2^29).

### spec.trafficPolicy.portLevelSettings[].connectionPool.http.maxRetries

`int32` · optional (explicit presence)

Maximum retries outstanding to all hosts in a cluster at a time. Default 2^32-1.

### spec.trafficPolicy.portLevelSettings[].connectionPool.http.idleTimeout

`string` · optional (explicit presence)

Idle timeout for upstream pool connections (no active requests). Default 1h. Modeled as
a duration string.

- rule: idle_timeout must be a valid, non-negative duration (e.g. "1h")

### spec.trafficPolicy.portLevelSettings[].connectionPool.http.h2UpgradePolicy

`string` · optional (explicit presence)

Policy for upgrading HTTP1.1 connections to HTTP2 for this destination. One of:
  DEFAULT        — use the global default.
  DO_NOT_UPGRADE — never upgrade (overrides the default).
  UPGRADE        — always upgrade (overrides the default).
external standard exception -- matches Istio HTTPSettings.H2UpgradePolicy

- rule: {"string":{"in":["DEFAULT","DO_NOT_UPGRADE","UPGRADE"]}}

### spec.trafficPolicy.portLevelSettings[].connectionPool.http.useClientProtocol

`bool` · optional (explicit presence)

If true, the client protocol is preserved when connecting to the backend. When true,
h2_upgrade_policy is ineffective.

### spec.trafficPolicy.portLevelSettings[].connectionPool.http.maxConcurrentStreams

`int32` · optional (explicit presence)

Maximum concurrent streams allowed for a peer on one HTTP/2 connection. Default 2^31-1.

### spec.trafficPolicy.portLevelSettings[].outlierDetection

`KubernetesDestinationRuleOutlierDetection`

Settings controlling eviction of unhealthy hosts for this port.

### spec.trafficPolicy.portLevelSettings[].outlierDetection.splitExternalLocalOriginErrors

`bool` · optional (explicit presence)

If true, locally-originated failures (connect failures, timeouts) are counted via
consecutive_local_origin_failures rather than upstream response codes. Default false.

### spec.trafficPolicy.portLevelSettings[].outlierDetection.consecutiveLocalOriginFailures

`uint32` · optional (explicit presence)

Number of consecutive locally-originated failures before ejection. Default 5. Only takes
effect when split_external_local_origin_errors is true.

### spec.trafficPolicy.portLevelSettings[].outlierDetection.consecutiveGatewayErrors

`uint32` · optional (explicit presence)

Number of consecutive gateway errors (502/503/504, or TCP connect failures) before
ejection. Disabled by default / when 0. Counted within consecutive_5xx_errors.

### spec.trafficPolicy.portLevelSettings[].outlierDetection.consecutive5xxErrors

`uint32` · optional (explicit presence)

Number of consecutive 5xx errors before ejection. Defaults to 5; disable by setting 0.

### spec.trafficPolicy.portLevelSettings[].outlierDetection.interval

`string` · optional (explicit presence)

Time interval between ejection sweep analyses (format 1h/1m/1s/1ms; must be >=1ms;
default 10s). Modeled as a duration string.

- rule: interval must be a valid duration of at least 1ms (e.g. "5m")

### spec.trafficPolicy.portLevelSettings[].outlierDetection.baseEjectionTime

`string` · optional (explicit presence)

Minimum ejection duration; a host stays ejected for this value times the number of times
it has been ejected (must be >=1ms; default 30s). Modeled as a duration string.

- rule: base_ejection_time must be a valid duration of at least 1ms (e.g. "15m")

### spec.trafficPolicy.portLevelSettings[].outlierDetection.maxEjectionPercent

`int32` · optional (explicit presence)

Maximum percentage of hosts in the pool that can be ejected. Default 10%.

### spec.trafficPolicy.portLevelSettings[].outlierDetection.minHealthPercent

`int32` · optional (explicit presence)

Outlier detection stays enabled only while at least this percentage of hosts is healthy;
below it, all hosts (healthy and unhealthy) are load balanced. Default 0% (disabled).

### spec.trafficPolicy.portLevelSettings[].tls

`KubernetesDestinationRuleClientTlsSettings`

TLS settings for connections the sidecar originates to this port.

### spec.trafficPolicy.portLevelSettings[].tls.mode

`string` · optional (explicit presence)

TLS mode for connections to the upstream. One of:
  DISABLE      — no TLS.
  SIMPLE       — originate TLS (one-way).
  MUTUAL       — mutual TLS presenting client certs (client_certificate + private_key,
                 or credential_name, required).
  ISTIO_MUTUAL — mutual TLS using Istio-generated certs; all other fields should be empty.
external standard exception -- matches Istio ClientTLSSettings.TLSmode

- rule: {"string":{"in":["DISABLE","SIMPLE","MUTUAL","ISTIO_MUTUAL"]}}

### spec.trafficPolicy.portLevelSettings[].tls.clientCertificate

`string` · optional (explicit presence)

Path to the client-side TLS certificate file. Required for MUTUAL; empty for ISTIO_MUTUAL.

### spec.trafficPolicy.portLevelSettings[].tls.privateKey

`string` · optional (explicit presence)

Path to the client's private key file. Required for MUTUAL; empty for ISTIO_MUTUAL.

### spec.trafficPolicy.portLevelSettings[].tls.caCertificates

`string` · optional (explicit presence)

Path to the CA certificate file used to verify the server certificate. Empty for
ISTIO_MUTUAL; if omitted, the OS CA certificates are used.

### spec.trafficPolicy.portLevelSettings[].tls.credentialName

`string` · optional (explicit presence)

Name of the Kubernetes secret (in the proxy's namespace) holding the client TLS certs
(and CA). Only one of (client_certificate + private_key + ca_certificates) OR
credential_name may be specified. At sidecars this is honored only when a
workload_selector is set; otherwise it applies at gateways only.

INFRA-CHART COMPOSABILITY: credential_name is a PLAIN reference to a Secret /
certificate by name, resolved by istiod/Envoy at runtime — NOT an Planton foreign key
(StringValueOrRef). It creates NO automatic DAG edge. To order this DestinationRule after
the secret it consumes in an infra chart, an author MUST express the dependency via
metadata.relationships (`uses` -> KubernetesCertificate / KubernetesSecret), e.g.:
  metadata:
    relationships:
      - kind: KubernetesSecret
        name: "{{ values.client_credential }}"
        type: uses
Kept plain (not StringValueOrRef) for fidelity and consistency with the sibling Gateway
family's `certificate_refs`. See the "Composing in Infra Charts" docs.

### spec.trafficPolicy.portLevelSettings[].tls.subjectAltNames

`[]string`

Alternate names to verify against the server certificate's subject alt names. Overrides
the ServiceEntry's subjectAltNames. If unset, validation is based on the downstream
host/authority header.

### spec.trafficPolicy.portLevelSettings[].tls.sni

`string` · optional (explicit presence)

SNI string presented to the server during the TLS handshake. If unset, derived from the
downstream host/authority header for SIMPLE and MUTUAL modes.

### spec.trafficPolicy.portLevelSettings[].tls.insecureSkipVerify

`bool` · optional (explicit presence)

If true, the proxy skips verifying the CA signature and SAN of the server certificate.
Default false.

### spec.trafficPolicy.portLevelSettings[].tls.caCrl

`string` · optional (explicit presence)

Path to the certificate revocation list (CRL) file used to verify the server cert. If
credential_name is set, the CRL must be supplied inside the credential, not here.

### spec.trafficPolicy.tunnel

`KubernetesDestinationRuleTunnelSettings`

Settings for tunneling TCP/TLS traffic over another transport. Applies to TCP or TLS
routes only (not HTTP).

### spec.trafficPolicy.tunnel.protocol

`string` · optional (explicit presence)

Protocol used to tunnel the downstream connection. CONNECT (HTTP CONNECT, the default
when unset) or POST (HTTP POST). The upstream proto documents this supported set; the
CRD leaves the field free-form, so this closed set is a deliberate, documented tightening.
external standard exception -- matches Istio TunnelSettings documented protocols

- rule: {"string":{"in":["CONNECT","POST"]}}

### spec.trafficPolicy.tunnel.targetHost

`string` · required

Host the downstream connection is tunneled to (FQDN or IP address). Required.

- rule: {"required":true}

### spec.trafficPolicy.tunnel.targetPort

`uint32` · required

Port the downstream connection is tunneled to (1-65535). Required.

- rule: {"required":true,"uint32":{"lte":65535,"gte":1}}

### spec.trafficPolicy.proxyProtocol

`KubernetesDestinationRuleProxyProtocol`

Upstream PROXY protocol settings.

### spec.trafficPolicy.proxyProtocol.version

`string` · optional (explicit presence)

PROXY protocol version to use. V1 (human-readable, the default) or V2 (binary).
external standard exception -- matches Istio ProxyProtocol.VERSION

- rule: {"string":{"in":["V1","V2"]}}

### spec.trafficPolicy.retryBudget

`KubernetesDestinationRuleRetryBudget`

Limits concurrent retries in relation to the number of active requests — a
relative alternative to a fixed retry cap: under low load few retries are
allowed, under high load proportionally more, without a thundering-herd
amplification. Applies at the destination and subset levels (upstream places it
on TrafficPolicy, not on port-level settings).

### spec.trafficPolicy.retryBudget.percent

`double` · optional (explicit presence)

Limit on concurrent retries as a percentage of the sum of active requests and
active pending requests (0-100). Upstream default: 20.

- rule: {"double":{"lte":100,"gte":0}}

### spec.trafficPolicy.retryBudget.minRetryConcurrency

`uint32` · optional (explicit presence)

Minimum number of concurrent retries allowed regardless of `percent` — keeps
retries possible when active-request counts are near zero. Upstream default: 3.

### spec.subsets

`[]KubernetesDestinationRuleSubset`

One or more named subsets representing individual versions of the service. Traffic
policies can be overridden per subset. A subset's policy only takes effect once a route
rule explicitly sends traffic to the subset.

### spec.subsets[].name

`string` · required

Name of the subset; used together with the service name for traffic splitting in route
rules. Required.

- rule: {"required":true}

### spec.subsets[].labels

`map<string, string>`

Labels filtering the service's endpoints that belong to this subset. May be empty when
the host represents multiple SNI hosts (e.g. an egress gateway), where the subset is
identified by its TLS settings instead.

### spec.subsets[].trafficPolicy

`KubernetesDestinationRuleTrafficPolicy`

Traffic policy applied to this subset. Subsets inherit the DestinationRule-level policy;
settings here override the corresponding DestinationRule-level settings.

### spec.subsets[].trafficPolicy.loadBalancer

`KubernetesDestinationRuleLoadBalancerSettings`

Settings controlling the load balancer algorithm.

- rule: at most one of simple or consistent_hash may be set
- rule: at most one of warmup or warmup_duration_secs may be set

### spec.subsets[].trafficPolicy.loadBalancer.simple

`string` · optional (explicit presence)

Standard load balancing algorithm requiring no tuning. One of:
  LEAST_CONN   — deprecated; use LEAST_REQUEST instead.
  RANDOM       — selects a random healthy host.
  PASSTHROUGH  — forwards to the original requested IP without load balancing.
  ROUND_ROBIN  — basic round robin; prefer LEAST_REQUEST in most cases.
  LEAST_REQUEST — favors hosts with the fewest outstanding requests (recommended).
Mutually exclusive with consistent_hash. // external standard exception -- matches Istio SimpleLB

- rule: {"string":{"in":["LEAST_CONN","RANDOM","PASSTHROUGH","ROUND_ROBIN","LEAST_REQUEST"]}}

### spec.subsets[].trafficPolicy.loadBalancer.consistentHash

`KubernetesDestinationRuleConsistentHashLb`

Consistent-hash (soft session affinity) load balancing based on a request property.
Mutually exclusive with simple.

- rule: at most one of http_header_name, http_cookie, use_source_ip, or http_query_parameter_name may be set
- rule: at most one of ring_hash or maglev may be set

### spec.subsets[].trafficPolicy.loadBalancer.consistentHash.httpHeaderName

`string` · optional (explicit presence)

Hash based on a specific HTTP header.

### spec.subsets[].trafficPolicy.loadBalancer.consistentHash.httpCookie

`KubernetesDestinationRuleHttpCookie`

Hash based on an HTTP cookie (generated if absent).

### spec.subsets[].trafficPolicy.loadBalancer.consistentHash.httpCookie.name

`string` · required

Name of the cookie. Required.

- rule: {"required":true}

### spec.subsets[].trafficPolicy.loadBalancer.consistentHash.httpCookie.path

`string` · optional (explicit presence)

Path to set on the generated cookie.

### spec.subsets[].trafficPolicy.loadBalancer.consistentHash.httpCookie.ttl

`string` · optional (explicit presence)

Lifetime of the cookie. If present and zero, the generated cookie is a session cookie.
Modeled as a duration string; upstream applies no minimum (duration-validation:
none), so only valid, non-negative durations are enforced.

- rule: ttl must be a valid, non-negative duration (e.g. "0s", "1h")

### spec.subsets[].trafficPolicy.loadBalancer.consistentHash.httpCookie.attributes

`[]KubernetesDestinationRuleHttpCookieAttribute`

Additional attributes set on the generated cookie (e.g. SameSite=Strict, Secure).
Each attribute has a required name and an optional value.

### spec.subsets[].trafficPolicy.loadBalancer.consistentHash.httpCookie.attributes[].name

`string` · required

The name of the cookie attribute. Required.

- rule: {"required":true}

### spec.subsets[].trafficPolicy.loadBalancer.consistentHash.httpCookie.attributes[].value

`string` · optional (explicit presence)

The optional value of the cookie attribute (an attribute like `Secure` carries none).

### spec.subsets[].trafficPolicy.loadBalancer.consistentHash.useSourceIp

`bool` · optional (explicit presence)

Hash based on the source IP address. Applicable to both TCP and HTTP connections.

### spec.subsets[].trafficPolicy.loadBalancer.consistentHash.httpQueryParameterName

`string` · optional (explicit presence)

Hash based on a specific HTTP query parameter.

### spec.subsets[].trafficPolicy.loadBalancer.consistentHash.ringHash

`KubernetesDestinationRuleRingHash`

The ring/modulo (ring hash) algorithm. Mutually exclusive with maglev.

### spec.subsets[].trafficPolicy.loadBalancer.consistentHash.ringHash.minimumRingSize

`uint64` · optional (explicit presence)

Minimum number of virtual nodes for the hash ring (default 1024). Larger rings give more
granular distribution.

### spec.subsets[].trafficPolicy.loadBalancer.consistentHash.maglev

`KubernetesDestinationRuleMagLev`

The Maglev algorithm. Mutually exclusive with ring_hash.

### spec.subsets[].trafficPolicy.loadBalancer.consistentHash.maglev.tableSize

`uint64` · optional (explicit presence)

Table size for Maglev hashing (default 65537). Should be a prime number less than
5000011; a larger table reduces disruption when backends change.

### spec.subsets[].trafficPolicy.loadBalancer.consistentHash.minimumRingSize

`uint64` · optional (explicit presence)

Deprecated: use `ring_hash.minimum_ring_size` instead. Minimum number of virtual nodes
in the hash ring.

### spec.subsets[].trafficPolicy.loadBalancer.localityLbSetting

`KubernetesDestinationRuleLocalityLbSetting`

Locality-weighted load balancing settings; overrides mesh-wide locality settings in
their entirety (no merging).

- rule: at most one of distribute, failover, or failover_priority may be set

### spec.subsets[].trafficPolicy.loadBalancer.localityLbSetting.distribute

`[]KubernetesDestinationRuleLocalityDistribute`

Explicit per-zone traffic distribution weights. Weights for a given `from` should sum to
100; any locality not listed receives no traffic.

### spec.subsets[].trafficPolicy.loadBalancer.localityLbSetting.distribute[].from

`string` · optional (explicit presence)

Originating locality, '/' separated (e.g. `region/zone/sub_zone`). Terminal wildcards are
allowed on any segment (e.g. `us-west/*`).

### spec.subsets[].trafficPolicy.loadBalancer.localityLbSetting.distribute[].to

`map<string, uint32>`

Map of upstream localities to traffic-distribution weights. The weights should sum to 100.

### spec.subsets[].trafficPolicy.loadBalancer.localityLbSetting.failover

`[]KubernetesDestinationRuleLocalityFailover`

Region-level failover policy. Zone/sub-zone failover is automatic; specify this only to
constrain cross-region failover. Should be used with outlier detection.

### spec.subsets[].trafficPolicy.loadBalancer.localityLbSetting.failover[].from

`string` · optional (explicit presence)

Originating region.

### spec.subsets[].trafficPolicy.loadBalancer.localityLbSetting.failover[].to

`string` · optional (explicit presence)

Destination region traffic fails over to when endpoints in the `from` region become
unhealthy.

### spec.subsets[].trafficPolicy.loadBalancer.localityLbSetting.failoverPriority

`[]string`

Ordered list of labels used to sort endpoints into priority tiers for failover. Entries
may be bare keys (`key`) or key=value pairs; special topology labels are supported. Used
with outlier detection.

### spec.subsets[].trafficPolicy.loadBalancer.localityLbSetting.enabled

`bool` · optional (explicit presence)

Enable (or disable) locality load balancing for this DestinationRule, overriding mesh-wide
settings entirely.

### spec.subsets[].trafficPolicy.loadBalancer.warmupDurationSecs

`string` · optional (explicit presence)

Deprecated: use `warmup` instead. Duration over which a newly added endpoint is ramped
up. Modeled as a duration string.

- rule: warmup_duration_secs must be a valid, non-negative duration (e.g. "30s")

### spec.subsets[].trafficPolicy.loadBalancer.warmup

`KubernetesDestinationRuleWarmupConfiguration`

Warmup configuration: newly created endpoints receive progressively increasing traffic
over the configured window. Only effective for ROUND_ROBIN and LEAST_REQUEST.

### spec.subsets[].trafficPolicy.loadBalancer.warmup.duration

`string` · required

Duration of warmup mode. Required. Modeled as a duration string.

- rule: duration must be a valid, non-negative duration (e.g. "30s")
- rule: {"required":true}

### spec.subsets[].trafficPolicy.loadBalancer.warmup.minimumPercent

`double` · optional (explicit presence)

Minimum percentage of origin weight at the start of warmup. If unspecified, defaults to
10. Upstream bounds: 0–100.

- rule: {"double":{"lte":100,"gte":0}}

### spec.subsets[].trafficPolicy.loadBalancer.warmup.aggression

`double` · optional (explicit presence)

Controls the speed of traffic increase over the warmup window. Defaults to 1.0 (linear);
higher values ramp up non-linearly. Upstream minimum: 1.

- rule: {"double":{"gte":1}}

### spec.subsets[].trafficPolicy.connectionPool

`KubernetesDestinationRuleConnectionPoolSettings`

Settings controlling the volume of connections to an upstream service.

### spec.subsets[].trafficPolicy.connectionPool.tcp

`KubernetesDestinationRuleTcpSettings`

Settings common to both HTTP and TCP upstream connections.

### spec.subsets[].trafficPolicy.connectionPool.tcp.maxConnections

`int32` · optional (explicit presence)

Maximum number of HTTP1/TCP connections to a destination host. Default 2^32-1.

### spec.subsets[].trafficPolicy.connectionPool.tcp.connectTimeout

`string` · optional (explicit presence)

TCP connection timeout (format: 1h/1m/1s/1ms; must be >=1ms; default 10s). Modeled as a
duration string.

- rule: connect_timeout must be a valid duration of at least 1ms (e.g. "30ms")

### spec.subsets[].trafficPolicy.connectionPool.tcp.tcpKeepalive

`KubernetesDestinationRuleTcpKeepalive`

TCP keepalive settings (enables SO_KEEPALIVE on the socket).

### spec.subsets[].trafficPolicy.connectionPool.tcp.tcpKeepalive.probes

`uint32` · optional (explicit presence)

Maximum number of keepalive probes to send before deciding the connection is dead.
Default: OS level (Linux 9).

### spec.subsets[].trafficPolicy.connectionPool.tcp.tcpKeepalive.time

`string` · optional (explicit presence)

Idle time before keepalive probes start. Default: OS level (Linux 7200s). Modeled as a
duration string.

- rule: time must be a valid, non-negative duration (e.g. "7200s")

### spec.subsets[].trafficPolicy.connectionPool.tcp.tcpKeepalive.interval

`string` · optional (explicit presence)

Interval between keepalive probes. Default: OS level (Linux 75s). Modeled as a duration
string.

- rule: interval must be a valid, non-negative duration (e.g. "75s")

### spec.subsets[].trafficPolicy.connectionPool.tcp.maxConnectionDuration

`string` · optional (explicit presence)

Maximum duration of a connection, measured from when it was established. If unset, no
maximum. Must be >=1ms. Modeled as a duration string.

- rule: max_connection_duration must be a valid duration of at least 1ms (e.g. "10s")

### spec.subsets[].trafficPolicy.connectionPool.tcp.idleTimeout

`string` · optional (explicit presence)

Idle timeout for TCP connections (no bytes sent/received on either side). Default 1h;
0s disables. Modeled as a duration string; upstream applies no minimum.

- rule: idle_timeout must be a valid, non-negative duration (e.g. "0s", "1h")

### spec.subsets[].trafficPolicy.connectionPool.http

`KubernetesDestinationRuleHttpSettings`

HTTP-specific connection pool settings (HTTP1.1/HTTP2/gRPC).

### spec.subsets[].trafficPolicy.connectionPool.http.http1MaxPendingRequests

`int32` · optional (explicit presence)

Maximum requests queued while waiting for a ready connection. Default 2^32-1. Applies to
both HTTP/1.1 and HTTP2.

### spec.subsets[].trafficPolicy.connectionPool.http.http2MaxRequests

`int32` · optional (explicit presence)

Maximum active requests to a destination. Default 2^32-1. Applies to both HTTP/1.1 and
HTTP2.

### spec.subsets[].trafficPolicy.connectionPool.http.maxRequestsPerConnection

`int32` · optional (explicit presence)

Maximum requests per backend connection. 1 disables keep-alive. Default 0 ("unlimited",
up to 2^29).

### spec.subsets[].trafficPolicy.connectionPool.http.maxRetries

`int32` · optional (explicit presence)

Maximum retries outstanding to all hosts in a cluster at a time. Default 2^32-1.

### spec.subsets[].trafficPolicy.connectionPool.http.idleTimeout

`string` · optional (explicit presence)

Idle timeout for upstream pool connections (no active requests). Default 1h. Modeled as
a duration string.

- rule: idle_timeout must be a valid, non-negative duration (e.g. "1h")

### spec.subsets[].trafficPolicy.connectionPool.http.h2UpgradePolicy

`string` · optional (explicit presence)

Policy for upgrading HTTP1.1 connections to HTTP2 for this destination. One of:
  DEFAULT        — use the global default.
  DO_NOT_UPGRADE — never upgrade (overrides the default).
  UPGRADE        — always upgrade (overrides the default).
external standard exception -- matches Istio HTTPSettings.H2UpgradePolicy

- rule: {"string":{"in":["DEFAULT","DO_NOT_UPGRADE","UPGRADE"]}}

### spec.subsets[].trafficPolicy.connectionPool.http.useClientProtocol

`bool` · optional (explicit presence)

If true, the client protocol is preserved when connecting to the backend. When true,
h2_upgrade_policy is ineffective.

### spec.subsets[].trafficPolicy.connectionPool.http.maxConcurrentStreams

`int32` · optional (explicit presence)

Maximum concurrent streams allowed for a peer on one HTTP/2 connection. Default 2^31-1.

### spec.subsets[].trafficPolicy.outlierDetection

`KubernetesDestinationRuleOutlierDetection`

Settings controlling eviction of unhealthy hosts from the load balancing pool.

### spec.subsets[].trafficPolicy.outlierDetection.splitExternalLocalOriginErrors

`bool` · optional (explicit presence)

If true, locally-originated failures (connect failures, timeouts) are counted via
consecutive_local_origin_failures rather than upstream response codes. Default false.

### spec.subsets[].trafficPolicy.outlierDetection.consecutiveLocalOriginFailures

`uint32` · optional (explicit presence)

Number of consecutive locally-originated failures before ejection. Default 5. Only takes
effect when split_external_local_origin_errors is true.

### spec.subsets[].trafficPolicy.outlierDetection.consecutiveGatewayErrors

`uint32` · optional (explicit presence)

Number of consecutive gateway errors (502/503/504, or TCP connect failures) before
ejection. Disabled by default / when 0. Counted within consecutive_5xx_errors.

### spec.subsets[].trafficPolicy.outlierDetection.consecutive5xxErrors

`uint32` · optional (explicit presence)

Number of consecutive 5xx errors before ejection. Defaults to 5; disable by setting 0.

### spec.subsets[].trafficPolicy.outlierDetection.interval

`string` · optional (explicit presence)

Time interval between ejection sweep analyses (format 1h/1m/1s/1ms; must be >=1ms;
default 10s). Modeled as a duration string.

- rule: interval must be a valid duration of at least 1ms (e.g. "5m")

### spec.subsets[].trafficPolicy.outlierDetection.baseEjectionTime

`string` · optional (explicit presence)

Minimum ejection duration; a host stays ejected for this value times the number of times
it has been ejected (must be >=1ms; default 30s). Modeled as a duration string.

- rule: base_ejection_time must be a valid duration of at least 1ms (e.g. "15m")

### spec.subsets[].trafficPolicy.outlierDetection.maxEjectionPercent

`int32` · optional (explicit presence)

Maximum percentage of hosts in the pool that can be ejected. Default 10%.

### spec.subsets[].trafficPolicy.outlierDetection.minHealthPercent

`int32` · optional (explicit presence)

Outlier detection stays enabled only while at least this percentage of hosts is healthy;
below it, all hosts (healthy and unhealthy) are load balanced. Default 0% (disabled).

### spec.subsets[].trafficPolicy.tls

`KubernetesDestinationRuleClientTlsSettings`

TLS settings for connections the sidecar originates to the upstream service.

### spec.subsets[].trafficPolicy.tls.mode

`string` · optional (explicit presence)

TLS mode for connections to the upstream. One of:
  DISABLE      — no TLS.
  SIMPLE       — originate TLS (one-way).
  MUTUAL       — mutual TLS presenting client certs (client_certificate + private_key,
                 or credential_name, required).
  ISTIO_MUTUAL — mutual TLS using Istio-generated certs; all other fields should be empty.
external standard exception -- matches Istio ClientTLSSettings.TLSmode

- rule: {"string":{"in":["DISABLE","SIMPLE","MUTUAL","ISTIO_MUTUAL"]}}

### spec.subsets[].trafficPolicy.tls.clientCertificate

`string` · optional (explicit presence)

Path to the client-side TLS certificate file. Required for MUTUAL; empty for ISTIO_MUTUAL.

### spec.subsets[].trafficPolicy.tls.privateKey

`string` · optional (explicit presence)

Path to the client's private key file. Required for MUTUAL; empty for ISTIO_MUTUAL.

### spec.subsets[].trafficPolicy.tls.caCertificates

`string` · optional (explicit presence)

Path to the CA certificate file used to verify the server certificate. Empty for
ISTIO_MUTUAL; if omitted, the OS CA certificates are used.

### spec.subsets[].trafficPolicy.tls.credentialName

`string` · optional (explicit presence)

Name of the Kubernetes secret (in the proxy's namespace) holding the client TLS certs
(and CA). Only one of (client_certificate + private_key + ca_certificates) OR
credential_name may be specified. At sidecars this is honored only when a
workload_selector is set; otherwise it applies at gateways only.

INFRA-CHART COMPOSABILITY: credential_name is a PLAIN reference to a Secret /
certificate by name, resolved by istiod/Envoy at runtime — NOT an Planton foreign key
(StringValueOrRef). It creates NO automatic DAG edge. To order this DestinationRule after
the secret it consumes in an infra chart, an author MUST express the dependency via
metadata.relationships (`uses` -> KubernetesCertificate / KubernetesSecret), e.g.:
  metadata:
    relationships:
      - kind: KubernetesSecret
        name: "{{ values.client_credential }}"
        type: uses
Kept plain (not StringValueOrRef) for fidelity and consistency with the sibling Gateway
family's `certificate_refs`. See the "Composing in Infra Charts" docs.

### spec.subsets[].trafficPolicy.tls.subjectAltNames

`[]string`

Alternate names to verify against the server certificate's subject alt names. Overrides
the ServiceEntry's subjectAltNames. If unset, validation is based on the downstream
host/authority header.

### spec.subsets[].trafficPolicy.tls.sni

`string` · optional (explicit presence)

SNI string presented to the server during the TLS handshake. If unset, derived from the
downstream host/authority header for SIMPLE and MUTUAL modes.

### spec.subsets[].trafficPolicy.tls.insecureSkipVerify

`bool` · optional (explicit presence)

If true, the proxy skips verifying the CA signature and SAN of the server certificate.
Default false.

### spec.subsets[].trafficPolicy.tls.caCrl

`string` · optional (explicit presence)

Path to the certificate revocation list (CRL) file used to verify the server cert. If
credential_name is set, the CRL must be supplied inside the credential, not here.

### spec.subsets[].trafficPolicy.portLevelSettings

`[]KubernetesDestinationRulePortTrafficPolicy`

Traffic policies specific to individual ports. Port-level settings fully override the
destination-level settings for that port (no field-level inheritance). Upstream allows
up to 4096.

- rule: {"repeated":{"maxItems":"4096"}}

### spec.subsets[].trafficPolicy.portLevelSettings[].port

`KubernetesIstioApiPortSelector`

The destination service port these settings apply to. Reuses the shared
istio.type.v1beta1 PortSelector (a single `number`).

### spec.subsets[].trafficPolicy.portLevelSettings[].port.number

`uint32`

Port number (1-65535).

- rule: {"uint32":{"lte":65535,"gte":1}}

### spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer

`KubernetesDestinationRuleLoadBalancerSettings`

Settings controlling the load balancer algorithm for this port.

- rule: at most one of simple or consistent_hash may be set
- rule: at most one of warmup or warmup_duration_secs may be set

### spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.simple

`string` · optional (explicit presence)

Standard load balancing algorithm requiring no tuning. One of:
  LEAST_CONN   — deprecated; use LEAST_REQUEST instead.
  RANDOM       — selects a random healthy host.
  PASSTHROUGH  — forwards to the original requested IP without load balancing.
  ROUND_ROBIN  — basic round robin; prefer LEAST_REQUEST in most cases.
  LEAST_REQUEST — favors hosts with the fewest outstanding requests (recommended).
Mutually exclusive with consistent_hash. // external standard exception -- matches Istio SimpleLB

- rule: {"string":{"in":["LEAST_CONN","RANDOM","PASSTHROUGH","ROUND_ROBIN","LEAST_REQUEST"]}}

### spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.consistentHash

`KubernetesDestinationRuleConsistentHashLb`

Consistent-hash (soft session affinity) load balancing based on a request property.
Mutually exclusive with simple.

- rule: at most one of http_header_name, http_cookie, use_source_ip, or http_query_parameter_name may be set
- rule: at most one of ring_hash or maglev may be set

### spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.httpHeaderName

`string` · optional (explicit presence)

Hash based on a specific HTTP header.

### spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.httpCookie

`KubernetesDestinationRuleHttpCookie`

Hash based on an HTTP cookie (generated if absent).

### spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.httpCookie.name

`string` · required

Name of the cookie. Required.

- rule: {"required":true}

### spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.httpCookie.path

`string` · optional (explicit presence)

Path to set on the generated cookie.

### spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.httpCookie.ttl

`string` · optional (explicit presence)

Lifetime of the cookie. If present and zero, the generated cookie is a session cookie.
Modeled as a duration string; upstream applies no minimum (duration-validation:
none), so only valid, non-negative durations are enforced.

- rule: ttl must be a valid, non-negative duration (e.g. "0s", "1h")

### spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.httpCookie.attributes

`[]KubernetesDestinationRuleHttpCookieAttribute`

Additional attributes set on the generated cookie (e.g. SameSite=Strict, Secure).
Each attribute has a required name and an optional value.

### spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.httpCookie.attributes[].name

`string` · required

The name of the cookie attribute. Required.

- rule: {"required":true}

### spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.httpCookie.attributes[].value

`string` · optional (explicit presence)

The optional value of the cookie attribute (an attribute like `Secure` carries none).

### spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.useSourceIp

`bool` · optional (explicit presence)

Hash based on the source IP address. Applicable to both TCP and HTTP connections.

### spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.httpQueryParameterName

`string` · optional (explicit presence)

Hash based on a specific HTTP query parameter.

### spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.ringHash

`KubernetesDestinationRuleRingHash`

The ring/modulo (ring hash) algorithm. Mutually exclusive with maglev.

### spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.ringHash.minimumRingSize

`uint64` · optional (explicit presence)

Minimum number of virtual nodes for the hash ring (default 1024). Larger rings give more
granular distribution.

### spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.maglev

`KubernetesDestinationRuleMagLev`

The Maglev algorithm. Mutually exclusive with ring_hash.

### spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.maglev.tableSize

`uint64` · optional (explicit presence)

Table size for Maglev hashing (default 65537). Should be a prime number less than
5000011; a larger table reduces disruption when backends change.

### spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.consistentHash.minimumRingSize

`uint64` · optional (explicit presence)

Deprecated: use `ring_hash.minimum_ring_size` instead. Minimum number of virtual nodes
in the hash ring.

### spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.localityLbSetting

`KubernetesDestinationRuleLocalityLbSetting`

Locality-weighted load balancing settings; overrides mesh-wide locality settings in
their entirety (no merging).

- rule: at most one of distribute, failover, or failover_priority may be set

### spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.localityLbSetting.distribute

`[]KubernetesDestinationRuleLocalityDistribute`

Explicit per-zone traffic distribution weights. Weights for a given `from` should sum to
100; any locality not listed receives no traffic.

### spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.localityLbSetting.distribute[].from

`string` · optional (explicit presence)

Originating locality, '/' separated (e.g. `region/zone/sub_zone`). Terminal wildcards are
allowed on any segment (e.g. `us-west/*`).

### spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.localityLbSetting.distribute[].to

`map<string, uint32>`

Map of upstream localities to traffic-distribution weights. The weights should sum to 100.

### spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.localityLbSetting.failover

`[]KubernetesDestinationRuleLocalityFailover`

Region-level failover policy. Zone/sub-zone failover is automatic; specify this only to
constrain cross-region failover. Should be used with outlier detection.

### spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.localityLbSetting.failover[].from

`string` · optional (explicit presence)

Originating region.

### spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.localityLbSetting.failover[].to

`string` · optional (explicit presence)

Destination region traffic fails over to when endpoints in the `from` region become
unhealthy.

### spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.localityLbSetting.failoverPriority

`[]string`

Ordered list of labels used to sort endpoints into priority tiers for failover. Entries
may be bare keys (`key`) or key=value pairs; special topology labels are supported. Used
with outlier detection.

### spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.localityLbSetting.enabled

`bool` · optional (explicit presence)

Enable (or disable) locality load balancing for this DestinationRule, overriding mesh-wide
settings entirely.

### spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.warmupDurationSecs

`string` · optional (explicit presence)

Deprecated: use `warmup` instead. Duration over which a newly added endpoint is ramped
up. Modeled as a duration string.

- rule: warmup_duration_secs must be a valid, non-negative duration (e.g. "30s")

### spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.warmup

`KubernetesDestinationRuleWarmupConfiguration`

Warmup configuration: newly created endpoints receive progressively increasing traffic
over the configured window. Only effective for ROUND_ROBIN and LEAST_REQUEST.

### spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.warmup.duration

`string` · required

Duration of warmup mode. Required. Modeled as a duration string.

- rule: duration must be a valid, non-negative duration (e.g. "30s")
- rule: {"required":true}

### spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.warmup.minimumPercent

`double` · optional (explicit presence)

Minimum percentage of origin weight at the start of warmup. If unspecified, defaults to
10. Upstream bounds: 0–100.

- rule: {"double":{"lte":100,"gte":0}}

### spec.subsets[].trafficPolicy.portLevelSettings[].loadBalancer.warmup.aggression

`double` · optional (explicit presence)

Controls the speed of traffic increase over the warmup window. Defaults to 1.0 (linear);
higher values ramp up non-linearly. Upstream minimum: 1.

- rule: {"double":{"gte":1}}

### spec.subsets[].trafficPolicy.portLevelSettings[].connectionPool

`KubernetesDestinationRuleConnectionPoolSettings`

Settings controlling the volume of connections to this port.

### spec.subsets[].trafficPolicy.portLevelSettings[].connectionPool.tcp

`KubernetesDestinationRuleTcpSettings`

Settings common to both HTTP and TCP upstream connections.

### spec.subsets[].trafficPolicy.portLevelSettings[].connectionPool.tcp.maxConnections

`int32` · optional (explicit presence)

Maximum number of HTTP1/TCP connections to a destination host. Default 2^32-1.

### spec.subsets[].trafficPolicy.portLevelSettings[].connectionPool.tcp.connectTimeout

`string` · optional (explicit presence)

TCP connection timeout (format: 1h/1m/1s/1ms; must be >=1ms; default 10s). Modeled as a
duration string.

- rule: connect_timeout must be a valid duration of at least 1ms (e.g. "30ms")

### spec.subsets[].trafficPolicy.portLevelSettings[].connectionPool.tcp.tcpKeepalive

`KubernetesDestinationRuleTcpKeepalive`

TCP keepalive settings (enables SO_KEEPALIVE on the socket).

### spec.subsets[].trafficPolicy.portLevelSettings[].connectionPool.tcp.tcpKeepalive.probes

`uint32` · optional (explicit presence)

Maximum number of keepalive probes to send before deciding the connection is dead.
Default: OS level (Linux 9).

### spec.subsets[].trafficPolicy.portLevelSettings[].connectionPool.tcp.tcpKeepalive.time

`string` · optional (explicit presence)

Idle time before keepalive probes start. Default: OS level (Linux 7200s). Modeled as a
duration string.

- rule: time must be a valid, non-negative duration (e.g. "7200s")

### spec.subsets[].trafficPolicy.portLevelSettings[].connectionPool.tcp.tcpKeepalive.interval

`string` · optional (explicit presence)

Interval between keepalive probes. Default: OS level (Linux 75s). Modeled as a duration
string.

- rule: interval must be a valid, non-negative duration (e.g. "75s")

### spec.subsets[].trafficPolicy.portLevelSettings[].connectionPool.tcp.maxConnectionDuration

`string` · optional (explicit presence)

Maximum duration of a connection, measured from when it was established. If unset, no
maximum. Must be >=1ms. Modeled as a duration string.

- rule: max_connection_duration must be a valid duration of at least 1ms (e.g. "10s")

### spec.subsets[].trafficPolicy.portLevelSettings[].connectionPool.tcp.idleTimeout

`string` · optional (explicit presence)

Idle timeout for TCP connections (no bytes sent/received on either side). Default 1h;
0s disables. Modeled as a duration string; upstream applies no minimum.

- rule: idle_timeout must be a valid, non-negative duration (e.g. "0s", "1h")

### spec.subsets[].trafficPolicy.portLevelSettings[].connectionPool.http

`KubernetesDestinationRuleHttpSettings`

HTTP-specific connection pool settings (HTTP1.1/HTTP2/gRPC).

### spec.subsets[].trafficPolicy.portLevelSettings[].connectionPool.http.http1MaxPendingRequests

`int32` · optional (explicit presence)

Maximum requests queued while waiting for a ready connection. Default 2^32-1. Applies to
both HTTP/1.1 and HTTP2.

### spec.subsets[].trafficPolicy.portLevelSettings[].connectionPool.http.http2MaxRequests

`int32` · optional (explicit presence)

Maximum active requests to a destination. Default 2^32-1. Applies to both HTTP/1.1 and
HTTP2.

### spec.subsets[].trafficPolicy.portLevelSettings[].connectionPool.http.maxRequestsPerConnection

`int32` · optional (explicit presence)

Maximum requests per backend connection. 1 disables keep-alive. Default 0 ("unlimited",
up to 2^29).

### spec.subsets[].trafficPolicy.portLevelSettings[].connectionPool.http.maxRetries

`int32` · optional (explicit presence)

Maximum retries outstanding to all hosts in a cluster at a time. Default 2^32-1.

### spec.subsets[].trafficPolicy.portLevelSettings[].connectionPool.http.idleTimeout

`string` · optional (explicit presence)

Idle timeout for upstream pool connections (no active requests). Default 1h. Modeled as
a duration string.

- rule: idle_timeout must be a valid, non-negative duration (e.g. "1h")

### spec.subsets[].trafficPolicy.portLevelSettings[].connectionPool.http.h2UpgradePolicy

`string` · optional (explicit presence)

Policy for upgrading HTTP1.1 connections to HTTP2 for this destination. One of:
  DEFAULT        — use the global default.
  DO_NOT_UPGRADE — never upgrade (overrides the default).
  UPGRADE        — always upgrade (overrides the default).
external standard exception -- matches Istio HTTPSettings.H2UpgradePolicy

- rule: {"string":{"in":["DEFAULT","DO_NOT_UPGRADE","UPGRADE"]}}

### spec.subsets[].trafficPolicy.portLevelSettings[].connectionPool.http.useClientProtocol

`bool` · optional (explicit presence)

If true, the client protocol is preserved when connecting to the backend. When true,
h2_upgrade_policy is ineffective.

### spec.subsets[].trafficPolicy.portLevelSettings[].connectionPool.http.maxConcurrentStreams

`int32` · optional (explicit presence)

Maximum concurrent streams allowed for a peer on one HTTP/2 connection. Default 2^31-1.

### spec.subsets[].trafficPolicy.portLevelSettings[].outlierDetection

`KubernetesDestinationRuleOutlierDetection`

Settings controlling eviction of unhealthy hosts for this port.

### spec.subsets[].trafficPolicy.portLevelSettings[].outlierDetection.splitExternalLocalOriginErrors

`bool` · optional (explicit presence)

If true, locally-originated failures (connect failures, timeouts) are counted via
consecutive_local_origin_failures rather than upstream response codes. Default false.

### spec.subsets[].trafficPolicy.portLevelSettings[].outlierDetection.consecutiveLocalOriginFailures

`uint32` · optional (explicit presence)

Number of consecutive locally-originated failures before ejection. Default 5. Only takes
effect when split_external_local_origin_errors is true.

### spec.subsets[].trafficPolicy.portLevelSettings[].outlierDetection.consecutiveGatewayErrors

`uint32` · optional (explicit presence)

Number of consecutive gateway errors (502/503/504, or TCP connect failures) before
ejection. Disabled by default / when 0. Counted within consecutive_5xx_errors.

### spec.subsets[].trafficPolicy.portLevelSettings[].outlierDetection.consecutive5xxErrors

`uint32` · optional (explicit presence)

Number of consecutive 5xx errors before ejection. Defaults to 5; disable by setting 0.

### spec.subsets[].trafficPolicy.portLevelSettings[].outlierDetection.interval

`string` · optional (explicit presence)

Time interval between ejection sweep analyses (format 1h/1m/1s/1ms; must be >=1ms;
default 10s). Modeled as a duration string.

- rule: interval must be a valid duration of at least 1ms (e.g. "5m")

### spec.subsets[].trafficPolicy.portLevelSettings[].outlierDetection.baseEjectionTime

`string` · optional (explicit presence)

Minimum ejection duration; a host stays ejected for this value times the number of times
it has been ejected (must be >=1ms; default 30s). Modeled as a duration string.

- rule: base_ejection_time must be a valid duration of at least 1ms (e.g. "15m")

### spec.subsets[].trafficPolicy.portLevelSettings[].outlierDetection.maxEjectionPercent

`int32` · optional (explicit presence)

Maximum percentage of hosts in the pool that can be ejected. Default 10%.

### spec.subsets[].trafficPolicy.portLevelSettings[].outlierDetection.minHealthPercent

`int32` · optional (explicit presence)

Outlier detection stays enabled only while at least this percentage of hosts is healthy;
below it, all hosts (healthy and unhealthy) are load balanced. Default 0% (disabled).

### spec.subsets[].trafficPolicy.portLevelSettings[].tls

`KubernetesDestinationRuleClientTlsSettings`

TLS settings for connections the sidecar originates to this port.

### spec.subsets[].trafficPolicy.portLevelSettings[].tls.mode

`string` · optional (explicit presence)

TLS mode for connections to the upstream. One of:
  DISABLE      — no TLS.
  SIMPLE       — originate TLS (one-way).
  MUTUAL       — mutual TLS presenting client certs (client_certificate + private_key,
                 or credential_name, required).
  ISTIO_MUTUAL — mutual TLS using Istio-generated certs; all other fields should be empty.
external standard exception -- matches Istio ClientTLSSettings.TLSmode

- rule: {"string":{"in":["DISABLE","SIMPLE","MUTUAL","ISTIO_MUTUAL"]}}

### spec.subsets[].trafficPolicy.portLevelSettings[].tls.clientCertificate

`string` · optional (explicit presence)

Path to the client-side TLS certificate file. Required for MUTUAL; empty for ISTIO_MUTUAL.

### spec.subsets[].trafficPolicy.portLevelSettings[].tls.privateKey

`string` · optional (explicit presence)

Path to the client's private key file. Required for MUTUAL; empty for ISTIO_MUTUAL.

### spec.subsets[].trafficPolicy.portLevelSettings[].tls.caCertificates

`string` · optional (explicit presence)

Path to the CA certificate file used to verify the server certificate. Empty for
ISTIO_MUTUAL; if omitted, the OS CA certificates are used.

### spec.subsets[].trafficPolicy.portLevelSettings[].tls.credentialName

`string` · optional (explicit presence)

Name of the Kubernetes secret (in the proxy's namespace) holding the client TLS certs
(and CA). Only one of (client_certificate + private_key + ca_certificates) OR
credential_name may be specified. At sidecars this is honored only when a
workload_selector is set; otherwise it applies at gateways only.

INFRA-CHART COMPOSABILITY: credential_name is a PLAIN reference to a Secret /
certificate by name, resolved by istiod/Envoy at runtime — NOT an Planton foreign key
(StringValueOrRef). It creates NO automatic DAG edge. To order this DestinationRule after
the secret it consumes in an infra chart, an author MUST express the dependency via
metadata.relationships (`uses` -> KubernetesCertificate / KubernetesSecret), e.g.:
  metadata:
    relationships:
      - kind: KubernetesSecret
        name: "{{ values.client_credential }}"
        type: uses
Kept plain (not StringValueOrRef) for fidelity and consistency with the sibling Gateway
family's `certificate_refs`. See the "Composing in Infra Charts" docs.

### spec.subsets[].trafficPolicy.portLevelSettings[].tls.subjectAltNames

`[]string`

Alternate names to verify against the server certificate's subject alt names. Overrides
the ServiceEntry's subjectAltNames. If unset, validation is based on the downstream
host/authority header.

### spec.subsets[].trafficPolicy.portLevelSettings[].tls.sni

`string` · optional (explicit presence)

SNI string presented to the server during the TLS handshake. If unset, derived from the
downstream host/authority header for SIMPLE and MUTUAL modes.

### spec.subsets[].trafficPolicy.portLevelSettings[].tls.insecureSkipVerify

`bool` · optional (explicit presence)

If true, the proxy skips verifying the CA signature and SAN of the server certificate.
Default false.

### spec.subsets[].trafficPolicy.portLevelSettings[].tls.caCrl

`string` · optional (explicit presence)

Path to the certificate revocation list (CRL) file used to verify the server cert. If
credential_name is set, the CRL must be supplied inside the credential, not here.

### spec.subsets[].trafficPolicy.tunnel

`KubernetesDestinationRuleTunnelSettings`

Settings for tunneling TCP/TLS traffic over another transport. Applies to TCP or TLS
routes only (not HTTP).

### spec.subsets[].trafficPolicy.tunnel.protocol

`string` · optional (explicit presence)

Protocol used to tunnel the downstream connection. CONNECT (HTTP CONNECT, the default
when unset) or POST (HTTP POST). The upstream proto documents this supported set; the
CRD leaves the field free-form, so this closed set is a deliberate, documented tightening.
external standard exception -- matches Istio TunnelSettings documented protocols

- rule: {"string":{"in":["CONNECT","POST"]}}

### spec.subsets[].trafficPolicy.tunnel.targetHost

`string` · required

Host the downstream connection is tunneled to (FQDN or IP address). Required.

- rule: {"required":true}

### spec.subsets[].trafficPolicy.tunnel.targetPort

`uint32` · required

Port the downstream connection is tunneled to (1-65535). Required.

- rule: {"required":true,"uint32":{"lte":65535,"gte":1}}

### spec.subsets[].trafficPolicy.proxyProtocol

`KubernetesDestinationRuleProxyProtocol`

Upstream PROXY protocol settings.

### spec.subsets[].trafficPolicy.proxyProtocol.version

`string` · optional (explicit presence)

PROXY protocol version to use. V1 (human-readable, the default) or V2 (binary).
external standard exception -- matches Istio ProxyProtocol.VERSION

- rule: {"string":{"in":["V1","V2"]}}

### spec.subsets[].trafficPolicy.retryBudget

`KubernetesDestinationRuleRetryBudget`

Limits concurrent retries in relation to the number of active requests — a
relative alternative to a fixed retry cap: under low load few retries are
allowed, under high load proportionally more, without a thundering-herd
amplification. Applies at the destination and subset levels (upstream places it
on TrafficPolicy, not on port-level settings).

### spec.subsets[].trafficPolicy.retryBudget.percent

`double` · optional (explicit presence)

Limit on concurrent retries as a percentage of the sum of active requests and
active pending requests (0-100). Upstream default: 20.

- rule: {"double":{"lte":100,"gte":0}}

### spec.subsets[].trafficPolicy.retryBudget.minRetryConcurrency

`uint32` · optional (explicit presence)

Minimum number of concurrent retries allowed regardless of `percent` — keeps
retries possible when active-request counts are near zero. Upstream default: 3.

### spec.exportTo

`[]string`

The namespaces to which this destination rule is exported (made visible for resolution).
If empty, the rule is exported to all namespaces. `.` exports to the declaring namespace
only; `*` exports to all namespaces.

### spec.workloadSelector

`KubernetesIstioApiWorkloadSelector`

Selects the specific pods/VMs this DestinationRule applies to, by label, within the same
namespace (selectors do not cross namespace boundaries). If omitted, the rule falls back
to its default (host-wide) behavior. Commonly used so that only certain sidecars get
egress TLS settings for an external service. Reuses the shared istio.type.v1beta1
WorkloadSelector (JSON key `matchLabels`) — the SAME selector PeerAuthentication and
AuthorizationPolicy use, NOT the networking `labels` selector (confirmed against
destination_rule.proto, which declares `istio.type.v1beta1.WorkloadSelector`).

INFRA-CHART COMPOSABILITY: workload_selector is a PLAIN label match, not an
Planton foreign key (StringValueOrRef). It is matched at runtime by istiod against pod
labels and creates NO automatic DAG edge to any workload resource. To order this
DestinationRule after the workloads it configures in an infra chart, an author MUST
express the dependency via metadata.relationships, e.g.:
  metadata:
    relationships:
      - kind: KubernetesDeployment
        name: "{{ values.app }}"
        type: depends_on
See the kind's "Composing in Infra Charts" docs for the full pattern.

### spec.workloadSelector.matchLabels

`map<string, string>`

One or more labels indicating the set of pods/VMs the policy applies to.
Faithful to istio.io/api `istio.type.v1beta1.WorkloadSelector.match_labels`,
whose upstream CRD constraints are: max 4096 entries; each value <= 63 chars;
label keys must be non-empty; and wildcards ('*') are not permitted in keys or
values. The size/length bounds are expressed via the standard `map` rule; the
non-empty-key and no-wildcard constraints map to upstream's CEL XValidation
rules and are expressed here as field-level CEL.

- rule: label selector keys must not be empty
- rule: wildcard ('*') is not allowed in label selector keys
- rule: wildcard ('*') is not allowed in label selector values
- rule: {"map":{"maxPairs":"4096","values":{"string":{"maxLen":"63"}}}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesDestinationRule, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.destination_rule_name` | `string` | Name of the created DestinationRule (equals metadata.name). |
| `status.outputs.namespace` | `string` | Namespace the DestinationRule was created in (the resolved spec.namespace). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.host` | KubernetesService | `status.outputs.kube_endpoint` |

## See Also

- [Overview](../README.md)

# AzureApplicationGateway

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**AzureApplicationGatewaySpec** defines an Azure Application Gateway --
the Layer 7 (HTTP/HTTPS) load balancer and reverse proxy that routes
traffic by host name, URI path, and other HTTP attributes, terminates
TLS (including mutual TLS), rewrites requests and responses, and
optionally enforces a Web Application Firewall policy.

The gateway bundles its sub-objects -- frontends, ports, listeners,
backend pools, backend settings, routing rules, path maps, probes,
certificates, SSL profiles, redirects, and rewrites -- because Azure
configures them as one atomic ARM resource: none has a life outside its
gateway, and they wire to each other BY NAME within the spec. What other
resources need to reach is exported as name-keyed map outputs
(`backend_address_pool_ids`, `frontend_ip_configuration_ids`), so
membership and chaining compose from the member side without splitting
the gateway apart.

**Traffic path**: a listener (frontend IP + port + protocol + optional
host names) feeds a request routing rule, which sends traffic to a
backend pool using backend HTTP settings (port, protocol, probe) --
either directly (BASIC_ROUTING) or through a URL path map
(PATH_BASED_ROUTING) that picks per-path backends, redirects, and
rewrites. TCP/TLS (layer-4) proxying uses the parallel
`listeners`/`backends`/`routing_rules` trio.

**SKUs**: BASIC (small, capped feature set), STANDARD_V2 (the general
workhorse: autoscale, zone redundancy, rewrites, mTLS), and WAF_V2
(STANDARD_V2 plus Web Application Firewall enforcement via a referenced
AzureWebApplicationFirewallPolicy). The v1 SKUs are retired in Azure and
not modeled.

**Subnet requirement**: the gateway needs a DEDICATED subnet with no
other resource types; /24 is the recommended size for production (v2
scales to 125 instances). **Deploys run 15-25 minutes** -- Azure's
slowest networking resource; plan pipelines accordingly.

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureApplicationGateway
metadata:
  name: test-agw
spec:
  region: eastus
  resourceGroup:
    value: test-rg
  name: test-agw
  subnetId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/virtualNetworks/test-vnet/subnets/agw
  # Exercises the WAF_V2 SKU, the WAF-policy FK, and autoscale.
  sku: WAF_V2
  autoscale:
    minCapacity: 2
    maxCapacity: 10
  zones: ["1", "2"]
  identity:
    type: USER_ASSIGNED
    identityIds:
      - value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/agw-identity
  firewallPolicyId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/applicationGatewayWebApplicationFirewallPolicies/test-waf
  forceFirewallPolicyAssociation: true
  frontendIpConfigurations:
    - name: public
      publicIpAddressId:
        value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/publicIPAddresses/agw-pip
    # Exercises the private frontend with static allocation.
    - name: internal
      subnetId:
        value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/virtualNetworks/test-vnet/subnets/agw
      privateIpAddress: 10.0.2.10
      privateIpAddressAllocation: STATIC
  frontendPorts:
    - name: http
      port: 80
    - name: https
      port: 443
    - name: amqps
      port: 5671
  backendAddressPools:
    - name: web
      ipAddresses:
        - 10.0.1.4
        - 10.0.1.5
    - name: api
      fqdns:
        - api.internal.contoso.com
  backendHttpSettings:
    # Exercises affinity, draining, path prefix, probe wiring, and
    # backend TLS trust.
    - name: web-settings
      port: 8080
      protocol: HTTP
      cookieBasedAffinityEnabled: true
      affinityCookieName: AppGwAffinity
      requestTimeout: 60
      probeName: web-health
      connectionDraining:
        enabled: true
        drainTimeoutSec: 60
    - name: api-settings
      port: 8443
      protocol: HTTPS
      pickHostNameFromBackendAddress: true
      trustedRootCertificateNames:
        - backend-ca
      dedicatedBackendConnectionEnabled: true
  httpListeners:
    - name: http-listener
      frontendIpConfigurationName: public
      frontendPortName: http
      protocol: HTTP
    # Exercises HTTPS + SNI + host names + per-listener WAF + custom
    # error pages + the mTLS profile.
    - name: https-listener
      frontendIpConfigurationName: public
      frontendPortName: https
      protocol: HTTPS
      sslCertificateName: tls-cert
      sslProfileName: mtls
      requireSni: true
      hostNames:
        - www.contoso.com
        - "*.contoso.com"
      firewallPolicyId:
        value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/applicationGatewayWebApplicationFirewallPolicies/listener-waf
      customErrorConfigurations:
        - statusCode: HTTP_STATUS_502
          customErrorPageUrl: https://errors.contoso.com/502.html
  requestRoutingRules:
    # Exercises the redirect arm.
    - name: http-to-https
      ruleType: BASIC_ROUTING
      httpListenerName: http-listener
      priority: 10
      redirectConfigurationName: to-https
    # Exercises path-based routing with rewrites.
    - name: https-rule
      ruleType: PATH_BASED_ROUTING
      httpListenerName: https-listener
      priority: 100
      urlPathMapName: paths
  urlPathMaps:
    - name: paths
      defaultBackendAddressPoolName: web
      defaultBackendHttpSettingsName: web-settings
      defaultRewriteRuleSetName: rewrites
      pathRules:
        - name: api
          paths:
            - /api/*
          backendAddressPoolName: api
          backendHttpSettingsName: api-settings
          firewallPolicyId:
            value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/applicationGatewayWebApplicationFirewallPolicies/route-waf
        - name: legacy-redirect
          paths:
            - /old/*
          redirectConfigurationName: to-https
  probes:
    - name: web-health
      protocol: HTTP
      path: /healthz
      pickHostNameFromBackendHttpSettings: true
      interval: 30
      timeout: 10
      unhealthyThreshold: 3
      minimumServers: 1
      match:
        statusCodes:
          - "200-399"
        body: ok
    # Exercises the TCP probe arm.
    - name: tcp-health
      protocol: TCP
      interval: 30
      timeout: 10
      unhealthyThreshold: 3
      port: 5671
      proxyProtocolHeaderEnabled: true
  sslCertificates:
    - name: tls-cert
      keyVaultSecretId:
        value: https://test-vault.vault.azure.net/secrets/tls-cert
    # Exercises the inline PFX arm.
    - name: inline-cert
      data: bWljcm9zb2Z0LXBmeA==
      password: changeit
  trustedRootCertificates:
    - name: backend-ca
      data: YmFja2VuZC1jYQ==
  trustedClientCertificates:
    - name: client-ca
      data: Y2xpZW50LWNh
  sslProfiles:
    - name: mtls
      trustedClientCertificateNames:
        - client-ca
      verifyClientCertificateIssuerDn: true
      verifyClientCertificateRevocation: OCSP
      sslPolicy:
        policyType: CUSTOM_V2
        minProtocolVersion: TLS_V1_3
        cipherSuites:
          - TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
  sslPolicy:
    policyType: PREDEFINED
    policyName: AppGwSslPolicy20220101S
  redirectConfigurations:
    - name: to-https
      redirectType: PERMANENT
      targetListenerName: https-listener
      includePath: true
      includeQueryString: true
  rewriteRuleSets:
    - name: rewrites
      rewriteRules:
        - name: forward-host
          ruleSequence: 100
          conditions:
            - variable: var_uri_path
              pattern: ^/legacy/(.*)
              ignoreCase: true
          requestHeaderConfigurations:
            - headerName: X-Forwarded-Host
              headerValue: "{var_host}"
          responseHeaderConfigurations:
            - headerName: Server
              headerValue: ""
          url:
            path: /{var_uri_path_1}
            reroute: true
  # Exercises the layer-4 trio.
  listeners:
    - name: amqps-listener
      frontendIpConfigurationName: public
      frontendPortName: amqps
      protocol: TLS
      sslCertificateName: inline-cert
  backends:
    - name: amqps-backend
      port: 5671
      protocol: TCP
      clientIpPreservationEnabled: true
      timeoutInSeconds: 120
  routingRules:
    - name: amqps-rule
      listenerName: amqps-listener
      backendAddressPoolName: web
      backendSettingsName: amqps-backend
      priority: 200
  customErrorConfigurations:
    - statusCode: HTTP_STATUS_403
      customErrorPageUrl: https://errors.contoso.com/403.html
  privateLinkConfigurations:
    - name: pl
      ipConfigurations:
        - name: nat
          subnetId:
            value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/virtualNetworks/test-vnet/subnets/agw
          privateIpAddressAllocation: DYNAMIC
          primary: true
  http2Enabled: true
  fipsEnabled: false
  globalConfiguration:
    requestBufferingEnabled: true
    responseBufferingEnabled: false
  tags:
    team: platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.name` | `string` | yes |  |  |
| `spec.subnetId` | `string \| valueFrom` | yes |  | AzureSubnet (`status.outputs.subnet_id`) |
| `spec.sku` | `enum` | yes |  |  |
| `spec.capacity` | `int32` |  |  |  |
| `spec.autoscale` | `AzureApplicationGatewayAutoscale` |  |  |  |
| `spec.autoscale.minCapacity` | `int32` | yes |  |  |
| `spec.autoscale.maxCapacity` | `int32` |  |  |  |
| `spec.zones` | `[]string` |  |  |  |
| `spec.identity` | `AzureApplicationGatewayIdentity` |  |  |  |
| `spec.identity.type` | `enum` | yes |  |  |
| `spec.identity.identityIds` | `[]string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.frontendIpConfigurations` | `[]AzureApplicationGatewayFrontendIpConfiguration` | yes |  |  |
| `spec.frontendIpConfigurations[].name` | `string` | yes |  |  |
| `spec.frontendIpConfigurations[].publicIpAddressId` | `string \| valueFrom` |  |  | AzurePublicIp (`status.outputs.public_ip_id`) |
| `spec.frontendIpConfigurations[].subnetId` | `string \| valueFrom` |  |  | AzureSubnet (`status.outputs.subnet_id`) |
| `spec.frontendIpConfigurations[].privateIpAddress` | `string` |  |  |  |
| `spec.frontendIpConfigurations[].privateIpAddressAllocation` | `enum` |  |  |  |
| `spec.frontendIpConfigurations[].privateLinkConfigurationName` | `string` |  |  |  |
| `spec.frontendPorts` | `[]AzureApplicationGatewayFrontendPort` | yes |  |  |
| `spec.frontendPorts[].name` | `string` | yes |  |  |
| `spec.frontendPorts[].port` | `int32` | yes |  |  |
| `spec.backendAddressPools` | `[]AzureApplicationGatewayBackendAddressPool` | yes |  |  |
| `spec.backendAddressPools[].name` | `string` | yes |  |  |
| `spec.backendAddressPools[].fqdns` | `[]string` |  |  |  |
| `spec.backendAddressPools[].ipAddresses` | `[]string` |  |  |  |
| `spec.backendHttpSettings` | `[]AzureApplicationGatewayBackendHttpSettings` |  |  |  |
| `spec.backendHttpSettings[].name` | `string` | yes |  |  |
| `spec.backendHttpSettings[].port` | `int32` | yes |  |  |
| `spec.backendHttpSettings[].protocol` | `enum` | yes |  |  |
| `spec.backendHttpSettings[].cookieBasedAffinityEnabled` | `bool` |  |  |  |
| `spec.backendHttpSettings[].affinityCookieName` | `string` |  |  |  |
| `spec.backendHttpSettings[].path` | `string` |  |  |  |
| `spec.backendHttpSettings[].requestTimeout` | `int32` |  | `30` |  |
| `spec.backendHttpSettings[].probeName` | `string` |  |  |  |
| `spec.backendHttpSettings[].hostName` | `string` |  |  |  |
| `spec.backendHttpSettings[].pickHostNameFromBackendAddress` | `bool` |  |  |  |
| `spec.backendHttpSettings[].trustedRootCertificateNames` | `[]string` |  |  |  |
| `spec.backendHttpSettings[].connectionDraining` | `AzureApplicationGatewayConnectionDraining` |  |  |  |
| `spec.backendHttpSettings[].connectionDraining.enabled` | `bool` |  |  |  |
| `spec.backendHttpSettings[].connectionDraining.drainTimeoutSec` | `int32` | yes |  |  |
| `spec.backendHttpSettings[].dedicatedBackendConnectionEnabled` | `bool` |  |  |  |
| `spec.httpListeners` | `[]AzureApplicationGatewayHttpListener` |  |  |  |
| `spec.httpListeners[].name` | `string` | yes |  |  |
| `spec.httpListeners[].frontendIpConfigurationName` | `string` | yes |  |  |
| `spec.httpListeners[].frontendPortName` | `string` | yes |  |  |
| `spec.httpListeners[].protocol` | `enum` | yes |  |  |
| `spec.httpListeners[].hostNames` | `[]string` |  |  |  |
| `spec.httpListeners[].sslCertificateName` | `string` |  |  |  |
| `spec.httpListeners[].requireSni` | `bool` |  |  |  |
| `spec.httpListeners[].sslProfileName` | `string` |  |  |  |
| `spec.httpListeners[].firewallPolicyId` | `string \| valueFrom` |  |  | AzureWebApplicationFirewallPolicy (`status.outputs.policy_id`) |
| `spec.httpListeners[].customErrorConfigurations` | `[]AzureApplicationGatewayCustomErrorConfiguration` |  |  |  |
| `spec.httpListeners[].customErrorConfigurations[].statusCode` | `enum` | yes |  |  |
| `spec.httpListeners[].customErrorConfigurations[].customErrorPageUrl` | `string` | yes |  |  |
| `spec.requestRoutingRules` | `[]AzureApplicationGatewayRequestRoutingRule` |  |  |  |
| `spec.requestRoutingRules[].name` | `string` | yes |  |  |
| `spec.requestRoutingRules[].ruleType` | `enum` | yes |  |  |
| `spec.requestRoutingRules[].httpListenerName` | `string` | yes |  |  |
| `spec.requestRoutingRules[].priority` | `int32` | yes |  |  |
| `spec.requestRoutingRules[].backendAddressPoolName` | `string` |  |  |  |
| `spec.requestRoutingRules[].backendHttpSettingsName` | `string` |  |  |  |
| `spec.requestRoutingRules[].urlPathMapName` | `string` |  |  |  |
| `spec.requestRoutingRules[].redirectConfigurationName` | `string` |  |  |  |
| `spec.requestRoutingRules[].rewriteRuleSetName` | `string` |  |  |  |
| `spec.urlPathMaps` | `[]AzureApplicationGatewayUrlPathMap` |  |  |  |
| `spec.urlPathMaps[].name` | `string` | yes |  |  |
| `spec.urlPathMaps[].defaultBackendAddressPoolName` | `string` |  |  |  |
| `spec.urlPathMaps[].defaultBackendHttpSettingsName` | `string` |  |  |  |
| `spec.urlPathMaps[].defaultRedirectConfigurationName` | `string` |  |  |  |
| `spec.urlPathMaps[].defaultRewriteRuleSetName` | `string` |  |  |  |
| `spec.urlPathMaps[].pathRules` | `[]AzureApplicationGatewayPathRule` | yes |  |  |
| `spec.urlPathMaps[].pathRules[].name` | `string` | yes |  |  |
| `spec.urlPathMaps[].pathRules[].paths` | `[]string` | yes |  |  |
| `spec.urlPathMaps[].pathRules[].backendAddressPoolName` | `string` |  |  |  |
| `spec.urlPathMaps[].pathRules[].backendHttpSettingsName` | `string` |  |  |  |
| `spec.urlPathMaps[].pathRules[].redirectConfigurationName` | `string` |  |  |  |
| `spec.urlPathMaps[].pathRules[].rewriteRuleSetName` | `string` |  |  |  |
| `spec.urlPathMaps[].pathRules[].firewallPolicyId` | `string \| valueFrom` |  |  | AzureWebApplicationFirewallPolicy (`status.outputs.policy_id`) |
| `spec.probes` | `[]AzureApplicationGatewayProbe` |  |  |  |
| `spec.probes[].name` | `string` | yes |  |  |
| `spec.probes[].protocol` | `enum` | yes |  |  |
| `spec.probes[].host` | `string` |  |  |  |
| `spec.probes[].pickHostNameFromBackendHttpSettings` | `bool` |  |  |  |
| `spec.probes[].path` | `string` |  |  |  |
| `spec.probes[].interval` | `int32` | yes |  |  |
| `spec.probes[].timeout` | `int32` | yes |  |  |
| `spec.probes[].unhealthyThreshold` | `int32` | yes |  |  |
| `spec.probes[].port` | `int32` |  |  |  |
| `spec.probes[].minimumServers` | `int32` |  |  |  |
| `spec.probes[].proxyProtocolHeaderEnabled` | `bool` |  |  |  |
| `spec.probes[].match` | `AzureApplicationGatewayProbeMatch` |  |  |  |
| `spec.probes[].match.statusCodes` | `[]string` | yes |  |  |
| `spec.probes[].match.body` | `string` |  |  |  |
| `spec.sslCertificates` | `[]AzureApplicationGatewaySslCertificate` |  |  |  |
| `spec.sslCertificates[].name` | `string` | yes |  |  |
| `spec.sslCertificates[].keyVaultSecretId` | `string \| valueFrom` |  |  | AzureKeyVaultCertificate (`status.outputs.versionless_secret_id`) |
| `spec.sslCertificates[].data` | `string` (sensitive) |  |  |  |
| `spec.sslCertificates[].password` | `string` (sensitive) |  |  |  |
| `spec.trustedRootCertificates` | `[]AzureApplicationGatewayTrustedRootCertificate` |  |  |  |
| `spec.trustedRootCertificates[].name` | `string` | yes |  |  |
| `spec.trustedRootCertificates[].keyVaultSecretId` | `string \| valueFrom` |  |  | AzureKeyVaultCertificate (`status.outputs.versionless_secret_id`) |
| `spec.trustedRootCertificates[].data` | `string` (sensitive) |  |  |  |
| `spec.trustedClientCertificates` | `[]AzureApplicationGatewayTrustedClientCertificate` |  |  |  |
| `spec.trustedClientCertificates[].name` | `string` | yes |  |  |
| `spec.trustedClientCertificates[].data` | `string` (sensitive) | yes |  |  |
| `spec.sslProfiles` | `[]AzureApplicationGatewaySslProfile` |  |  |  |
| `spec.sslProfiles[].name` | `string` | yes |  |  |
| `spec.sslProfiles[].trustedClientCertificateNames` | `[]string` |  |  |  |
| `spec.sslProfiles[].verifyClientCertificateIssuerDn` | `bool` |  |  |  |
| `spec.sslProfiles[].verifyClientCertificateRevocation` | `enum` |  |  |  |
| `spec.sslProfiles[].sslPolicy` | `AzureApplicationGatewaySslPolicy` |  |  |  |
| `spec.sslProfiles[].sslPolicy.policyType` | `enum` |  |  |  |
| `spec.sslProfiles[].sslPolicy.policyName` | `string` |  |  |  |
| `spec.sslProfiles[].sslPolicy.minProtocolVersion` | `enum` |  |  |  |
| `spec.sslProfiles[].sslPolicy.cipherSuites` | `[]string` |  |  |  |
| `spec.sslProfiles[].sslPolicy.disabledProtocols` | `[]enum` |  |  |  |
| `spec.sslPolicy` | `AzureApplicationGatewaySslPolicy` |  |  |  |
| `spec.sslPolicy.policyType` | `enum` |  |  |  |
| `spec.sslPolicy.policyName` | `string` |  |  |  |
| `spec.sslPolicy.minProtocolVersion` | `enum` |  |  |  |
| `spec.sslPolicy.cipherSuites` | `[]string` |  |  |  |
| `spec.sslPolicy.disabledProtocols` | `[]enum` |  |  |  |
| `spec.redirectConfigurations` | `[]AzureApplicationGatewayRedirectConfiguration` |  |  |  |
| `spec.redirectConfigurations[].name` | `string` | yes |  |  |
| `spec.redirectConfigurations[].redirectType` | `enum` | yes |  |  |
| `spec.redirectConfigurations[].targetListenerName` | `string` |  |  |  |
| `spec.redirectConfigurations[].targetUrl` | `string` |  |  |  |
| `spec.redirectConfigurations[].includePath` | `bool` |  |  |  |
| `spec.redirectConfigurations[].includeQueryString` | `bool` |  |  |  |
| `spec.rewriteRuleSets` | `[]AzureApplicationGatewayRewriteRuleSet` |  |  |  |
| `spec.rewriteRuleSets[].name` | `string` | yes |  |  |
| `spec.rewriteRuleSets[].rewriteRules` | `[]AzureApplicationGatewayRewriteRule` | yes |  |  |
| `spec.rewriteRuleSets[].rewriteRules[].name` | `string` | yes |  |  |
| `spec.rewriteRuleSets[].rewriteRules[].ruleSequence` | `int32` | yes |  |  |
| `spec.rewriteRuleSets[].rewriteRules[].conditions` | `[]AzureApplicationGatewayRewriteRuleCondition` |  |  |  |
| `spec.rewriteRuleSets[].rewriteRules[].conditions[].variable` | `string` | yes |  |  |
| `spec.rewriteRuleSets[].rewriteRules[].conditions[].pattern` | `string` | yes |  |  |
| `spec.rewriteRuleSets[].rewriteRules[].conditions[].ignoreCase` | `bool` |  |  |  |
| `spec.rewriteRuleSets[].rewriteRules[].conditions[].negate` | `bool` |  |  |  |
| `spec.rewriteRuleSets[].rewriteRules[].requestHeaderConfigurations` | `[]AzureApplicationGatewayRewriteRuleHeaderConfiguration` |  |  |  |
| `spec.rewriteRuleSets[].rewriteRules[].requestHeaderConfigurations[].headerName` | `string` | yes |  |  |
| `spec.rewriteRuleSets[].rewriteRules[].requestHeaderConfigurations[].headerValue` | `string` |  |  |  |
| `spec.rewriteRuleSets[].rewriteRules[].responseHeaderConfigurations` | `[]AzureApplicationGatewayRewriteRuleHeaderConfiguration` |  |  |  |
| `spec.rewriteRuleSets[].rewriteRules[].responseHeaderConfigurations[].headerName` | `string` | yes |  |  |
| `spec.rewriteRuleSets[].rewriteRules[].responseHeaderConfigurations[].headerValue` | `string` |  |  |  |
| `spec.rewriteRuleSets[].rewriteRules[].url` | `AzureApplicationGatewayRewriteRuleUrl` |  |  |  |
| `spec.rewriteRuleSets[].rewriteRules[].url.path` | `string` |  |  |  |
| `spec.rewriteRuleSets[].rewriteRules[].url.queryString` | `string` |  |  |  |
| `spec.rewriteRuleSets[].rewriteRules[].url.components` | `enum` |  |  |  |
| `spec.rewriteRuleSets[].rewriteRules[].url.reroute` | `bool` |  |  |  |
| `spec.listeners` | `[]AzureApplicationGatewayLayer4Listener` |  |  |  |
| `spec.listeners[].name` | `string` | yes |  |  |
| `spec.listeners[].frontendIpConfigurationName` | `string` | yes |  |  |
| `spec.listeners[].frontendPortName` | `string` | yes |  |  |
| `spec.listeners[].protocol` | `enum` | yes |  |  |
| `spec.listeners[].hostNames` | `[]string` |  |  |  |
| `spec.listeners[].sslCertificateName` | `string` |  |  |  |
| `spec.listeners[].sslProfileName` | `string` |  |  |  |
| `spec.backends` | `[]AzureApplicationGatewayLayer4BackendSettings` |  |  |  |
| `spec.backends[].name` | `string` | yes |  |  |
| `spec.backends[].port` | `int32` | yes |  |  |
| `spec.backends[].protocol` | `enum` | yes |  |  |
| `spec.backends[].clientIpPreservationEnabled` | `bool` |  |  |  |
| `spec.backends[].hostName` | `string` |  |  |  |
| `spec.backends[].probeName` | `string` |  |  |  |
| `spec.backends[].timeoutInSeconds` | `int32` |  | `30` |  |
| `spec.backends[].trustedRootCertificateNames` | `[]string` |  |  |  |
| `spec.routingRules` | `[]AzureApplicationGatewayLayer4RoutingRule` |  |  |  |
| `spec.routingRules[].name` | `string` | yes |  |  |
| `spec.routingRules[].listenerName` | `string` | yes |  |  |
| `spec.routingRules[].backendAddressPoolName` | `string` | yes |  |  |
| `spec.routingRules[].backendSettingsName` | `string` | yes |  |  |
| `spec.routingRules[].priority` | `int32` | yes |  |  |
| `spec.firewallPolicyId` | `string \| valueFrom` |  |  | AzureWebApplicationFirewallPolicy (`status.outputs.policy_id`) |
| `spec.forceFirewallPolicyAssociation` | `bool` |  |  |  |
| `spec.customErrorConfigurations` | `[]AzureApplicationGatewayCustomErrorConfiguration` |  |  |  |
| `spec.customErrorConfigurations[].statusCode` | `enum` | yes |  |  |
| `spec.customErrorConfigurations[].customErrorPageUrl` | `string` | yes |  |  |
| `spec.privateLinkConfigurations` | `[]AzureApplicationGatewayPrivateLinkConfiguration` |  |  |  |
| `spec.privateLinkConfigurations[].name` | `string` | yes |  |  |
| `spec.privateLinkConfigurations[].ipConfigurations` | `[]AzureApplicationGatewayPrivateLinkIpConfiguration` | yes |  |  |
| `spec.privateLinkConfigurations[].ipConfigurations[].name` | `string` | yes |  |  |
| `spec.privateLinkConfigurations[].ipConfigurations[].subnetId` | `string \| valueFrom` | yes |  | AzureSubnet (`status.outputs.subnet_id`) |
| `spec.privateLinkConfigurations[].ipConfigurations[].privateIpAddress` | `string` |  |  |  |
| `spec.privateLinkConfigurations[].ipConfigurations[].privateIpAddressAllocation` | `enum` | yes |  |  |
| `spec.privateLinkConfigurations[].ipConfigurations[].primary` | `bool` |  |  |  |
| `spec.http2Enabled` | `bool` |  | `false` |  |
| `spec.fipsEnabled` | `bool` |  |  |  |
| `spec.globalConfiguration` | `AzureApplicationGatewayGlobalConfiguration` |  |  |  |
| `spec.globalConfiguration.requestBufferingEnabled` | `bool` |  |  |  |
| `spec.globalConfiguration.responseBufferingEnabled` | `bool` |  |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.region

`string` · required

The Azure region the gateway lives in. Must match the subnet, public
IPs, and any WAF policy it references. Changing it replaces the
gateway.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.resourceGroup

`string | valueFrom` · required

The Azure resource group the gateway will be created in. Can be a
literal resource-group name or a reference to an AzureResourceGroup's
name output. Changing it replaces the gateway.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.name

`string` · required

The gateway's name, unique within the resource group: 1-80 letters,
digits, underscores, periods, and hyphens, starting with a letter or
digit. Changing it replaces the gateway.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"80","pattern":"^[a-zA-Z0-9][a-zA-Z0-9._-]*$"}}

### spec.subnetId

`string | valueFrom` · required

The gateway's DEDICATED subnet, by ARM ID -- no other resource types
may share it. /24 recommended for production (up to 125 v2
instances plus Azure-reserved addresses).

- references: AzureSubnet (`status.outputs.subnet_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.sku

`enum` · required

The SKU. BASIC caps capacity at 2 and drops autoscale, mTLS, and URL
rewriting; WAF_V2 is required to enforce a WAF policy.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_application_gateway_sku_unspecified` -- Not specified -- invalid; the SKU is the gateway's fundamental choice.
- `BASIC` -- Small fixed-capacity gateway (1-2 units): no autoscale, mTLS, or URL rewriting. For dev/test and small workloads.
- `STANDARD_V2` -- The general-purpose v2 platform: autoscale, zone redundancy, rewrites, mTLS, Private Link.
- `WAF_V2` -- STANDARD_V2 plus Web Application Firewall enforcement (attach an AzureWebApplicationFirewallPolicy).

### spec.capacity

`int32` · optional (explicit presence)

Fixed instance count. BASIC: 1-2 (required). STANDARD_V2/WAF_V2:
1-125, mutually exclusive with autoscale -- exactly one of the two
must be set.

- rule: {"int32":{"lte":125,"gte":1}}

### spec.autoscale

`AzureApplicationGatewayAutoscale`

Autoscale bounds (v2 SKUs only). Mutually exclusive with capacity.

### spec.autoscale.minCapacity

`int32` · required · optional (explicit presence)

The floor the gateway never scales below. 0-100; 0 lets the gateway
scale to zero capacity units at idle (billing floor, not
availability -- keep >= 2 for production). Required inside autoscale.

- rule: {"required":true,"int32":{"lte":100,"gte":0}}

### spec.autoscale.maxCapacity

`int32` · optional (explicit presence)

The ceiling the gateway never scales above. 2-125; omit for the
subscription-level maximum.

- rule: {"int32":{"lte":125,"gte":2}}

### spec.zones

`[]string`

Availability zones the gateway spans (e.g. ["1", "2", "3"] for zone
redundancy). v2 SKUs only; fixed at creation.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["1","2","3"]}}}}

### spec.identity

`AzureApplicationGatewayIdentity`

The gateway's managed identity. A user-assigned identity is required
when any certificate references Key Vault (the identity needs GET on
the vault's secrets).

- rule: identity_ids is required for USER_ASSIGNED and SYSTEM_AND_USER_ASSIGNED and must be empty for SYSTEM_ASSIGNED

### spec.identity.type

`enum` · required

Identity flavor. USER_ASSIGNED is the working grain -- Key
Vault certificate references authenticate through a user-assigned
identity that must exist (with vault access) before the gateway.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_application_gateway_identity_type_unspecified` -- Not specified: the gateway has no managed identity.
- `SYSTEM_ASSIGNED` -- Azure-managed identity created and rotated with the gateway.
- `USER_ASSIGNED` -- Bring-your-own user-assigned identities (set identity_ids) -- required for Key Vault certificate references.
- `SYSTEM_AND_USER_ASSIGNED` -- Both a system-assigned identity and the listed user-assigned ones.

### spec.identity.identityIds

`[]string | valueFrom`

The user-assigned identities attached to the gateway, by ARM ID.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.frontendIpConfigurations

`[]AzureApplicationGatewayFrontendIpConfiguration` · required

The frontend IP configurations listeners bind to. A frontend is
public (a referenced Standard public IP) or private (an address in
the gateway subnet). At least one is required; each frontend's ARM ID
is exported in the frontend_ip_configuration_ids map output.

- rule: {"repeated":{"minItems":"1"}}
- rule: a frontend is public (public_ip_address_id) or private (subnet_id) -- set exactly one
- rule: STATIC private_ip_address_allocation requires private_ip_address (and DYNAMIC forbids it); both require a private (subnet) frontend

### spec.frontendIpConfigurations[].name

`string` · required

The frontend's name, unique within the gateway. Listeners reference
it; its ARM ID is exported under this key in the
frontend_ip_configuration_ids map output.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"80"}}

### spec.frontendIpConfigurations[].publicIpAddressId

`string | valueFrom`

For a PUBLIC frontend: the Standard public IP the frontend
advertises, by ARM ID. References a first-class AzurePublicIp so the
address is visible in the resource graph and reusable.

- references: AzurePublicIp (`status.outputs.public_ip_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzurePublicIp, name: <that resource's name>, fieldPath: status.outputs.public_ip_id}} -- a bare string does not parse

### spec.frontendIpConfigurations[].subnetId

`string | valueFrom`

For a PRIVATE frontend: the subnet the frontend takes an address in
-- the gateway's own dedicated subnet. Set together with the
allocation choice below.

- references: AzureSubnet (`status.outputs.subnet_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.frontendIpConfigurations[].privateIpAddress

`string`

For a private frontend with STATIC allocation: the fixed address
inside the subnet's range.

### spec.frontendIpConfigurations[].privateIpAddressAllocation

`enum`

How the private address is assigned. Unspecified means DYNAMIC
(Azure picks); STATIC pins private_ip_address.

Allowed values (use exactly as shown):

- `azure_application_gateway_ip_allocation_unspecified` -- Not specified: DYNAMIC.
- `DYNAMIC` -- Azure assigns an address from the subnet.
- `STATIC` -- The declared private_ip_address is used.

### spec.frontendIpConfigurations[].privateLinkConfigurationName

`string`

The private_link_configurations entry (by name) that exposes this
frontend over Private Link.

### spec.frontendPorts

`[]AzureApplicationGatewayFrontendPort` · required

The frontend ports listeners bind to, declared once and referenced by
name (e.g. one "https" port 443 shared by many listeners).

- rule: {"repeated":{"minItems":"1"}}

### spec.frontendPorts[].name

`string` · required

The port's name, unique within the gateway (e.g. "http", "https").

- rule: {"required":true,"string":{"minLen":"1","maxLen":"80"}}

### spec.frontendPorts[].port

`int32` · required

The port number, 1-65535.

- rule: {"required":true,"int32":{"lte":65535,"gte":1}}

### spec.backendAddressPools

`[]AzureApplicationGatewayBackendAddressPool` · required

The backend pools traffic routes to, identified by FQDN and/or IP
address -- or joined from the member side: a NIC or scale set
references the pool's ARM ID from the backend_address_pool_ids map
output. At least one pool is required.

- rule: {"repeated":{"minItems":"1"}}

### spec.backendAddressPools[].name

`string` · required

The pool's name, unique within the gateway. Routing rules reference
it; its ARM ID is exported under this key in the
backend_address_pool_ids map output -- the seam NICs and scale sets
join from the member side.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"80"}}

### spec.backendAddressPools[].fqdns

`[]string`

Backend FQDNs the gateway resolves and routes to (e.g. an App
Service hostname).

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.backendAddressPools[].ipAddresses

`[]string`

Backend IP addresses.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.backendHttpSettings

`[]AzureApplicationGatewayBackendHttpSettings`

How the gateway talks to backends over HTTP(S): port, protocol,
affinity, timeouts, probe, host-header handling, connection draining,
and backend-TLS trust. At least one of backend_http_settings (L7) or
backends (L4) must be declared.

- rule: host_name and pick_host_name_from_backend_address are mutually exclusive

### spec.backendHttpSettings[].name

`string` · required

The settings' name, unique within the gateway; routing rules
reference it.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"80"}}

### spec.backendHttpSettings[].port

`int32` · required

The backend port, 1-65535.

- rule: {"required":true,"int32":{"lte":65535,"gte":1}}

### spec.backendHttpSettings[].protocol

`enum` · required

The backend protocol: HTTP or HTTPS (end-to-end TLS; pair with
trusted_root_certificate_names for private CAs).

- rule: backend HTTP settings use HTTP or HTTPS
- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_application_gateway_protocol_unspecified` -- Not specified -- invalid; declare the protocol.
- `HTTP` -- Plain HTTP (L7).
- `HTTPS` -- TLS-terminated HTTPS (L7).
- `TCP` -- Raw TCP proxying (L4).
- `TLS` -- TLS proxying (L4).

### spec.backendHttpSettings[].cookieBasedAffinityEnabled

`bool`

Whether requests from the same client stick to the same backend via
a gateway-managed cookie. Azure's default is off (round-robin).

### spec.backendHttpSettings[].affinityCookieName

`string`

A custom name for the affinity cookie (only meaningful with affinity
on).

### spec.backendHttpSettings[].path

`string`

A path prefix prepended to every request forwarded with these
settings (e.g. "/api/"). Must start with "/".

- rule: path must start with /

### spec.backendHttpSettings[].requestTimeout

`int32` · optional (explicit presence)

Seconds the gateway waits for a backend response before failing the
request. 1-86400; Azure's default is 30.

- default: `30`
- rule: {"int32":{"lte":86400,"gte":1}}

### spec.backendHttpSettings[].probeName

`string`

The custom probe (by name) that health-checks these backends. Omit
for Azure's default probe.

### spec.backendHttpSettings[].hostName

`string`

Override the Host header sent to backends. Mutually exclusive with
pick_host_name_from_backend_address.

### spec.backendHttpSettings[].pickHostNameFromBackendAddress

`bool`

Set the Host header to each backend's own address -- the grain for
multi-tenant PaaS backends (App Service). Azure's default is false.

### spec.backendHttpSettings[].trustedRootCertificateNames

`[]string`

The trusted_root_certificates entries (by name) validating the
backend's HTTPS certificate when it is not signed by a public CA.

### spec.backendHttpSettings[].connectionDraining

`AzureApplicationGatewayConnectionDraining`

Connection draining: how long existing connections to a backend
being removed keep flowing before termination.

### spec.backendHttpSettings[].connectionDraining.enabled

`bool`

Whether draining is active.

### spec.backendHttpSettings[].connectionDraining.drainTimeoutSec

`int32` · required

Seconds drained connections stay alive, 1-3600.

- rule: {"required":true,"int32":{"lte":3600,"gte":1}}

### spec.backendHttpSettings[].dedicatedBackendConnectionEnabled

`bool`

Whether each gateway instance opens its own dedicated connections to
backends instead of sharing a pool. Azure's default is false.

### spec.httpListeners

`[]AzureApplicationGatewayHttpListener`

The HTTP(S) entry points: frontend + port + protocol + optional host
names (with wildcards) for name-based virtual hosting, TLS
termination via ssl_certificate_name, mTLS via ssl_profile_name, and
an optional per-listener WAF policy. At least one of http_listeners
(L7) or listeners (L4) must be declared.

- rule: an HTTPS listener requires ssl_certificate_name (and an HTTP listener must not set one)

### spec.httpListeners[].name

`string` · required

The listener's name, unique within the gateway; routing rules
reference it.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"80"}}

### spec.httpListeners[].frontendIpConfigurationName

`string` · required

The frontend IP configuration (by name) the listener binds to.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.httpListeners[].frontendPortName

`string` · required

The frontend port (by name) the listener binds to.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.httpListeners[].protocol

`enum` · required

The listener protocol: HTTP or HTTPS (requires
ssl_certificate_name).

- rule: HTTP listeners use HTTP or HTTPS (TCP/TLS belong to the layer-4 listeners field)
- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_application_gateway_protocol_unspecified` -- Not specified -- invalid; declare the protocol.
- `HTTP` -- Plain HTTP (L7).
- `HTTPS` -- TLS-terminated HTTPS (L7).
- `TCP` -- Raw TCP proxying (L4).
- `TLS` -- TLS proxying (L4).

### spec.httpListeners[].hostNames

`[]string`

Host names for name-based virtual hosting, with wildcard support
(e.g. "*.contoso.com"). Only requests with a matching Host header
reach this listener; empty matches everything.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.httpListeners[].sslCertificateName

`string`

The ssl_certificates entry (by name) terminating TLS. Required for
HTTPS.

### spec.httpListeners[].requireSni

`bool`

Whether clients must send SNI -- required for multiple HTTPS
listeners sharing a frontend by host name.

### spec.httpListeners[].sslProfileName

`string`

The ssl_profiles entry (by name) applying a TLS posture (policy +
mutual TLS) to this listener.

### spec.httpListeners[].firewallPolicyId

`string | valueFrom`

A per-listener WAF policy overriding the gateway-wide one, by ARM ID
(WAF_V2 only).

- references: AzureWebApplicationFirewallPolicy (`status.outputs.policy_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureWebApplicationFirewallPolicy, name: <that resource's name>, fieldPath: status.outputs.policy_id}} -- a bare string does not parse

### spec.httpListeners[].customErrorConfigurations

`[]AzureApplicationGatewayCustomErrorConfiguration`

Custom error pages for errors this listener generates.

### spec.httpListeners[].customErrorConfigurations[].statusCode

`enum` · required

The gateway-generated status the page replaces.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_application_gateway_custom_error_status_code_unspecified` -- Not specified -- invalid; declare the status.
- `HTTP_STATUS_400` -- 400 Bad Request.
- `HTTP_STATUS_403` -- 403 Forbidden (the page WAF blocks serve).
- `HTTP_STATUS_404` -- 404 Not Found.
- `HTTP_STATUS_405` -- 405 Method Not Allowed.
- `HTTP_STATUS_408` -- 408 Request Timeout.
- `HTTP_STATUS_500` -- 500 Internal Server Error.
- `HTTP_STATUS_502` -- 502 Bad Gateway (no healthy backends).
- `HTTP_STATUS_503` -- 503 Service Unavailable.
- `HTTP_STATUS_504` -- 504 Gateway Timeout.

### spec.httpListeners[].customErrorConfigurations[].customErrorPageUrl

`string` · required

The publicly reachable HTML page URL (e.g. a blob URL).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.requestRoutingRules

`[]AzureApplicationGatewayRequestRoutingRule`

The L7 routing rules connecting listeners to backends -- directly
(BASIC_ROUTING), through a URL path map (PATH_BASED_ROUTING), or to a
redirect. At least one of request_routing_rules (L7) or routing_rules
(L4) must be declared.

- rule: a BASIC_ROUTING rule targets a backend (pool + settings) or a redirect, never both; a PATH_BASED_ROUTING rule requires url_path_map_name and carries neither

### spec.requestRoutingRules[].name

`string` · required

The rule's name, unique within the gateway.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"80"}}

### spec.requestRoutingRules[].ruleType

`enum` · required

BASIC_ROUTING sends everything from the listener to one backend (or
redirect); PATH_BASED_ROUTING consults the url_path_map.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_application_gateway_request_routing_rule_type_unspecified` -- Not specified -- invalid; declare the rule type.
- `BASIC_ROUTING` -- Everything from the listener goes to one backend or redirect (ARM: "Basic").
- `PATH_BASED_ROUTING` -- The url_path_map picks per-path destinations (ARM: "PathBasedRouting").

### spec.requestRoutingRules[].httpListenerName

`string` · required

The http_listeners entry (by name) feeding this rule.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.requestRoutingRules[].priority

`int32` · required

Evaluation order across rules: LOWER wins when listeners overlap
(e.g. a specific host listener vs a wildcard). 1-20000, unique per
gateway; required on the v2 SKUs this catalog models.

- rule: {"required":true,"int32":{"lte":20000,"gte":1}}

### spec.requestRoutingRules[].backendAddressPoolName

`string`

For BASIC_ROUTING: the backend pool (by name) traffic goes to.
Mutually exclusive with redirect_configuration_name.

### spec.requestRoutingRules[].backendHttpSettingsName

`string`

For BASIC_ROUTING: the backend HTTP settings (by name) used to reach
the pool.

### spec.requestRoutingRules[].urlPathMapName

`string`

For PATH_BASED_ROUTING: the url_path_maps entry (by name) that picks
per-path destinations.

### spec.requestRoutingRules[].redirectConfigurationName

`string`

Route to a redirect instead of a backend (e.g. the HTTP -> HTTPS
rule). Mutually exclusive with the backend pair.

### spec.requestRoutingRules[].rewriteRuleSetName

`string`

The rewrite_rule_sets entry (by name) applied to traffic on this
rule.

### spec.urlPathMaps

`[]AzureApplicationGatewayUrlPathMap`

URL path maps for path-based routing: per-path backends, redirects,
rewrites, and WAF policies, with defaults for unmatched paths.

- rule: a url path map defaults to a backend (pool + settings) or a redirect -- set exactly one

### spec.urlPathMaps[].name

`string` · required

The map's name, unique within the gateway; PATH_BASED_ROUTING rules
reference it.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"80"}}

### spec.urlPathMaps[].defaultBackendAddressPoolName

`string`

Where unmatched paths go: a default backend (pool + settings) or a
default redirect -- exactly one.

### spec.urlPathMaps[].defaultBackendHttpSettingsName

`string`

The backend HTTP settings for the default backend.

### spec.urlPathMaps[].defaultRedirectConfigurationName

`string`

The default redirect (by name) for unmatched paths. Mutually
exclusive with the default backend pair.

### spec.urlPathMaps[].defaultRewriteRuleSetName

`string`

The rewrite rule set (by name) applied to unmatched paths.

### spec.urlPathMaps[].pathRules

`[]AzureApplicationGatewayPathRule` · required

The per-path rules, first match wins.

- rule: {"repeated":{"minItems":"1"}}
- rule: a path rule targets a backend (pool + settings) or a redirect -- set exactly one

### spec.urlPathMaps[].pathRules[].name

`string` · required

The rule's name, unique within the map.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"80"}}

### spec.urlPathMaps[].pathRules[].paths

`[]string` · required

The path patterns this rule matches (e.g. "/api/*", "/images/*").

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.urlPathMaps[].pathRules[].backendAddressPoolName

`string`

The backend pool (by name) for matching paths. Mutually exclusive
with redirect_configuration_name.

### spec.urlPathMaps[].pathRules[].backendHttpSettingsName

`string`

The backend HTTP settings for this rule's pool.

### spec.urlPathMaps[].pathRules[].redirectConfigurationName

`string`

A redirect (by name) for matching paths instead of a backend.

### spec.urlPathMaps[].pathRules[].rewriteRuleSetName

`string`

The rewrite rule set (by name) applied to matching paths.

### spec.urlPathMaps[].pathRules[].firewallPolicyId

`string | valueFrom`

A per-route WAF policy overriding the listener's and gateway's, by
ARM ID (WAF_V2 only).

- references: AzureWebApplicationFirewallPolicy (`status.outputs.policy_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureWebApplicationFirewallPolicy, name: <that resource's name>, fieldPath: status.outputs.policy_id}} -- a bare string does not parse

### spec.probes

`[]AzureApplicationGatewayProbe`

Custom health probes. Backends without a custom probe get Azure's
default (GET / on the backend port).

- rule: timeout must not exceed interval
- rule: HTTP/HTTPS probes require path and exactly one of host / pick_host_name_from_backend_http_settings; TCP/TLS probes carry none of path, host, match (and only they may set proxy_protocol_header_enabled)

### spec.probes[].name

`string` · required

The probe's name, unique within the gateway; backend settings
reference it.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"80"}}

### spec.probes[].protocol

`enum` · required

The probe protocol. HTTP/HTTPS probe with a GET on `path`; TCP/TLS
probe the raw connection (for layer-4 backends).

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_application_gateway_protocol_unspecified` -- Not specified -- invalid; declare the protocol.
- `HTTP` -- Plain HTTP (L7).
- `HTTPS` -- TLS-terminated HTTPS (L7).
- `TCP` -- Raw TCP proxying (L4).
- `TLS` -- TLS proxying (L4).

### spec.probes[].host

`string`

The host header for HTTP(S) probes. Exactly one of host and
pick_host_name_from_backend_http_settings for HTTP(S); forbidden for
TCP/TLS.

### spec.probes[].pickHostNameFromBackendHttpSettings

`bool`

Use the backend settings' host name (or the backend address) as the
probe host. Azure's default is false.

### spec.probes[].path

`string`

The URI path HTTP(S) probes GET (e.g. "/healthz"). Required for
HTTP(S); forbidden for TCP/TLS.

- rule: path must start with /

### spec.probes[].interval

`int32` · required

Seconds between probe attempts, 1-86400.

- rule: {"required":true,"int32":{"lte":86400,"gte":1}}

### spec.probes[].timeout

`int32` · required

Seconds before an unanswered probe counts as failed, 1-86400. Must
not exceed interval.

- rule: {"required":true,"int32":{"lte":86400,"gte":1}}

### spec.probes[].unhealthyThreshold

`int32` · required

Consecutive failures before the backend is marked unhealthy, 1-20.

- rule: {"required":true,"int32":{"lte":20,"gte":1}}

### spec.probes[].port

`int32` · optional (explicit presence)

The probe port. Omit to probe the backend settings' port.

- rule: {"int32":{"lte":65535,"gte":1}}

### spec.probes[].minimumServers

`int32` · optional (explicit presence)

Keep the pool marked healthy while at least this many backends
respond (0 = mark unhealthy backends individually, Azure's default).

- rule: {"int32":{"gte":0}}

### spec.probes[].proxyProtocolHeaderEnabled

`bool`

Send the PROXY protocol header on TCP/TLS probes (only valid for
those protocols).

### spec.probes[].match

`AzureApplicationGatewayProbeMatch`

Healthy-response criteria for HTTP(S) probes: status-code ranges and
an optional body substring. Forbidden for TCP/TLS.

### spec.probes[].match.statusCodes

`[]string` · required

Status codes (single values or ranges) counted healthy, e.g.
"200-399", "401".

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.probes[].match.body

`string`

A substring the response body must contain. Empty matches any body.

### spec.sslCertificates

`[]AzureApplicationGatewaySslCertificate`

TLS certificates for HTTPS listeners: sourced from Key Vault (the
production grain -- renewals propagate) or inline PFX data.

- rule: a certificate is sourced from Key Vault (key_vault_secret_id) or inline PFX (data) -- set exactly one; password only accompanies inline data

### spec.sslCertificates[].name

`string` · required

The certificate's name, unique within the gateway; HTTPS listeners
reference it.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"80"}}

### spec.sslCertificates[].keyVaultSecretId

`string | valueFrom`

The Key Vault secret holding the PFX, by data-plane secret ID -- the
production grain. Defaults to referencing an
AzureKeyVaultCertificate's versionless_secret_id output so renewals
propagate to the gateway automatically. Requires a user-assigned
identity with GET on the vault's secrets. Mutually exclusive with
inline data.

- references: AzureKeyVaultCertificate (`status.outputs.versionless_secret_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureKeyVaultCertificate, name: <that resource's name>, fieldPath: status.outputs.versionless_secret_id}} -- a bare string does not parse

### spec.sslCertificates[].data

`string` · sensitive

Inline base64-encoded PFX data -- for certificates not yet in Key
Vault. Mutually exclusive with key_vault_secret_id.

### spec.sslCertificates[].password

`string` · sensitive

The PFX password (only with inline data).

### spec.trustedRootCertificates

`[]AzureApplicationGatewayTrustedRootCertificate`

CA bundles the gateway trusts when connecting to backends over HTTPS
with certificates not signed by a public CA.

- rule: a trusted root certificate is sourced from Key Vault (key_vault_secret_id) or inline data -- set exactly one

### spec.trustedRootCertificates[].name

`string` · required

The bundle's name, unique within the gateway; backend settings
reference it.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"80"}}

### spec.trustedRootCertificates[].keyVaultSecretId

`string | valueFrom`

The Key Vault secret holding the CA certificate, by data-plane
secret ID. Mutually exclusive with inline data.

- references: AzureKeyVaultCertificate (`status.outputs.versionless_secret_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureKeyVaultCertificate, name: <that resource's name>, fieldPath: status.outputs.versionless_secret_id}} -- a bare string does not parse

### spec.trustedRootCertificates[].data

`string` · sensitive

Inline base64-encoded CA certificate data. Mutually exclusive with
key_vault_secret_id.

### spec.trustedClientCertificates

`[]AzureApplicationGatewayTrustedClientCertificate`

Client-CA bundles for mutual TLS -- referenced by ssl_profiles that
verify client certificates. Not supported on the BASIC SKU.

### spec.trustedClientCertificates[].name

`string` · required

The bundle's name, unique within the gateway; SSL profiles reference
it.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"80"}}

### spec.trustedClientCertificates[].data

`string` · required · sensitive

The base64-encoded client-CA certificate (or chain) that client
certificates must chain to.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.sslProfiles

`[]AzureApplicationGatewaySslProfile`

Named TLS postures listeners opt into: per-profile TLS policy plus
mutual-TLS client verification.

### spec.sslProfiles[].name

`string` · required

The profile's name, unique within the gateway; listeners reference
it.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"80"}}

### spec.sslProfiles[].trustedClientCertificateNames

`[]string`

The client-CA bundles (by name) for mutual TLS -- clients must
present certificates chaining to one of them.

### spec.sslProfiles[].verifyClientCertificateIssuerDn

`bool`

Verify the client certificate's immediate issuer DN against the
trusted bundles (defense against sibling-CA certificates). Azure's
default is false.

### spec.sslProfiles[].verifyClientCertificateRevocation

`enum`

Check client-certificate revocation via OCSP. Unspecified skips
revocation checking (Azure's default).

Allowed values (use exactly as shown):

- `azure_application_gateway_client_revocation_check_unspecified` -- Not specified: no revocation checking (Azure's default).
- `OCSP` -- Check revocation via OCSP.

### spec.sslProfiles[].sslPolicy

`AzureApplicationGatewaySslPolicy`

A per-profile TLS policy overriding the gateway-wide ssl_policy for
listeners using this profile.

- rule: disabled_protocols is mutually exclusive with policy_type; PREDEFINED requires policy_name; CUSTOM/CUSTOM_V2 require min_protocol_version and cipher_suites

### spec.sslProfiles[].sslPolicy.policyType

`enum`

PREDEFINED applies a Microsoft-maintained policy (set policy_name);
CUSTOM picks a minimum protocol and explicit cipher suites;
CUSTOM_V2 additionally supports TLS 1.3.

Allowed values (use exactly as shown):

- `azure_application_gateway_ssl_policy_type_unspecified` -- Not specified: Azure's default TLS policy (optionally narrowed by disabled_protocols).
- `PREDEFINED` -- A Microsoft-maintained named policy (set policy_name).
- `CUSTOM` -- A custom minimum protocol + cipher-suite list (TLS 1.0-1.2).
- `CUSTOM_V2` -- A custom policy with TLS 1.3 support.

### spec.sslProfiles[].sslPolicy.policyName

`string`

For PREDEFINED: the Microsoft policy name (e.g.
"AppGwSslPolicy20220101S" -- the current strict recommendation,
"AppGwSslPolicy20220101", "AppGwSslPolicy20170401S").

### spec.sslProfiles[].sslPolicy.minProtocolVersion

`enum`

For CUSTOM/CUSTOM_V2: the minimum TLS version clients may use.
TLS_V1_3 requires CUSTOM_V2.

Allowed values (use exactly as shown):

- `azure_application_gateway_tls_protocol_unspecified` -- Not specified.
- `TLS_V1_0` -- TLS 1.0 (legacy -- disable in new deployments).
- `TLS_V1_1` -- TLS 1.1 (legacy).
- `TLS_V1_2` -- TLS 1.2 -- the standard floor.
- `TLS_V1_3` -- TLS 1.3 (CUSTOM_V2 policies only).

### spec.sslProfiles[].sslPolicy.cipherSuites

`[]string`

For CUSTOM/CUSTOM_V2: the allowed cipher suites, in Azure's exact
TLS_* names (e.g. "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384").

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.sslProfiles[].sslPolicy.disabledProtocols

`[]enum`

Ban specific TLS versions while keeping Azure's default policy
otherwise. Mutually exclusive with policy_type.

Allowed values (use exactly as shown):

- `azure_application_gateway_tls_protocol_unspecified` -- Not specified.
- `TLS_V1_0` -- TLS 1.0 (legacy -- disable in new deployments).
- `TLS_V1_1` -- TLS 1.1 (legacy).
- `TLS_V1_2` -- TLS 1.2 -- the standard floor.
- `TLS_V1_3` -- TLS 1.3 (CUSTOM_V2 policies only).

### spec.sslPolicy

`AzureApplicationGatewaySslPolicy`

The gateway-wide TLS policy (minimum protocol version, cipher
suites, or a Microsoft predefined policy). Listeners with an
ssl_profile use the profile's policy instead.

- rule: disabled_protocols is mutually exclusive with policy_type; PREDEFINED requires policy_name; CUSTOM/CUSTOM_V2 require min_protocol_version and cipher_suites

### spec.sslPolicy.policyType

`enum`

PREDEFINED applies a Microsoft-maintained policy (set policy_name);
CUSTOM picks a minimum protocol and explicit cipher suites;
CUSTOM_V2 additionally supports TLS 1.3.

Allowed values (use exactly as shown):

- `azure_application_gateway_ssl_policy_type_unspecified` -- Not specified: Azure's default TLS policy (optionally narrowed by disabled_protocols).
- `PREDEFINED` -- A Microsoft-maintained named policy (set policy_name).
- `CUSTOM` -- A custom minimum protocol + cipher-suite list (TLS 1.0-1.2).
- `CUSTOM_V2` -- A custom policy with TLS 1.3 support.

### spec.sslPolicy.policyName

`string`

For PREDEFINED: the Microsoft policy name (e.g.
"AppGwSslPolicy20220101S" -- the current strict recommendation,
"AppGwSslPolicy20220101", "AppGwSslPolicy20170401S").

### spec.sslPolicy.minProtocolVersion

`enum`

For CUSTOM/CUSTOM_V2: the minimum TLS version clients may use.
TLS_V1_3 requires CUSTOM_V2.

Allowed values (use exactly as shown):

- `azure_application_gateway_tls_protocol_unspecified` -- Not specified.
- `TLS_V1_0` -- TLS 1.0 (legacy -- disable in new deployments).
- `TLS_V1_1` -- TLS 1.1 (legacy).
- `TLS_V1_2` -- TLS 1.2 -- the standard floor.
- `TLS_V1_3` -- TLS 1.3 (CUSTOM_V2 policies only).

### spec.sslPolicy.cipherSuites

`[]string`

For CUSTOM/CUSTOM_V2: the allowed cipher suites, in Azure's exact
TLS_* names (e.g. "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384").

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.sslPolicy.disabledProtocols

`[]enum`

Ban specific TLS versions while keeping Azure's default policy
otherwise. Mutually exclusive with policy_type.

Allowed values (use exactly as shown):

- `azure_application_gateway_tls_protocol_unspecified` -- Not specified.
- `TLS_V1_0` -- TLS 1.0 (legacy -- disable in new deployments).
- `TLS_V1_1` -- TLS 1.1 (legacy).
- `TLS_V1_2` -- TLS 1.2 -- the standard floor.
- `TLS_V1_3` -- TLS 1.3 (CUSTOM_V2 policies only).

### spec.redirectConfigurations

`[]AzureApplicationGatewayRedirectConfiguration`

HTTP redirect definitions (e.g. the universal HTTP -> HTTPS
redirect), referenced by routing rules and path rules.

- rule: a redirect targets a listener (target_listener_name) or an external URL (target_url) -- set exactly one

### spec.redirectConfigurations[].name

`string` · required

The redirect's name, unique within the gateway; routing and path
rules reference it.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"80"}}

### spec.redirectConfigurations[].redirectType

`enum` · required

The HTTP redirect status: PERMANENT (301), FOUND (302), SEE_OTHER
(303), or TEMPORARY (307).

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_application_gateway_redirect_type_unspecified` -- Not specified -- invalid; declare the redirect status.
- `PERMANENT` -- 301 Moved Permanently -- cached by clients; the production HTTP -> HTTPS choice.
- `FOUND` -- 302 Found.
- `SEE_OTHER` -- 303 See Other.
- `TEMPORARY` -- 307 Temporary Redirect -- preserves the request method.

### spec.redirectConfigurations[].targetListenerName

`string`

Redirect to another listener on this gateway (the HTTP -> HTTPS
pattern), by name. Mutually exclusive with target_url.

### spec.redirectConfigurations[].targetUrl

`string`

Redirect to an external URL. Mutually exclusive with
target_listener_name.

### spec.redirectConfigurations[].includePath

`bool`

Whether the original request path is carried onto the target.
Azure's default is false.

### spec.redirectConfigurations[].includeQueryString

`bool`

Whether the original query string is carried onto the target.
Azure's default is false.

### spec.rewriteRuleSets

`[]AzureApplicationGatewayRewriteRuleSet`

Rewrite rule sets: request/response header edits and URL rewrites
driven by conditions, referenced by routing rules and path rules.

### spec.rewriteRuleSets[].name

`string` · required

The set's name, unique within the gateway; routing and path rules
reference it.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"80"}}

### spec.rewriteRuleSets[].rewriteRules

`[]AzureApplicationGatewayRewriteRule` · required

The rules, executed in rule_sequence order.

- rule: {"repeated":{"minItems":"1"}}

### spec.rewriteRuleSets[].rewriteRules[].name

`string` · required

The rule's name.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.rewriteRuleSets[].rewriteRules[].ruleSequence

`int32` · required

Execution order within the set, 1-1000 (lower first).

- rule: {"required":true,"int32":{"lte":1000,"gte":1}}

### spec.rewriteRuleSets[].rewriteRules[].conditions

`[]AzureApplicationGatewayRewriteRuleCondition`

Conditions gating the rule (all must hold). Variables use Azure's
server-variable syntax (e.g. "var_host", "http_req_Authorization").

### spec.rewriteRuleSets[].rewriteRules[].conditions[].variable

`string` · required

The server variable inspected (e.g. "var_uri_path",
"http_req_User-Agent").

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.rewriteRuleSets[].rewriteRules[].conditions[].pattern

`string` · required

The regex (or literal) the variable is matched against.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.rewriteRuleSets[].rewriteRules[].conditions[].ignoreCase

`bool`

Case-insensitive matching. Azure's default is false.

### spec.rewriteRuleSets[].rewriteRules[].conditions[].negate

`bool`

Invert the match. Azure's default is false.

### spec.rewriteRuleSets[].rewriteRules[].requestHeaderConfigurations

`[]AzureApplicationGatewayRewriteRuleHeaderConfiguration`

Request-header edits (set a header to a value; empty value removes
it).

### spec.rewriteRuleSets[].rewriteRules[].requestHeaderConfigurations[].headerName

`string` · required

The header to set (e.g. "X-Forwarded-Host").

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.rewriteRuleSets[].rewriteRules[].requestHeaderConfigurations[].headerValue

`string`

The value to set. Supports capture groups from conditions (e.g.
"{var_host}"); an empty value removes the header.

### spec.rewriteRuleSets[].rewriteRules[].responseHeaderConfigurations

`[]AzureApplicationGatewayRewriteRuleHeaderConfiguration`

Response-header edits.

### spec.rewriteRuleSets[].rewriteRules[].responseHeaderConfigurations[].headerName

`string` · required

The header to set (e.g. "X-Forwarded-Host").

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.rewriteRuleSets[].rewriteRules[].responseHeaderConfigurations[].headerValue

`string`

The value to set. Supports capture groups from conditions (e.g.
"{var_host}"); an empty value removes the header.

### spec.rewriteRuleSets[].rewriteRules[].url

`AzureApplicationGatewayRewriteRuleUrl`

URL path/query rewriting. Not supported on the BASIC SKU.

### spec.rewriteRuleSets[].rewriteRules[].url.path

`string`

The new path (supports captures). Omit to keep the original path.

### spec.rewriteRuleSets[].rewriteRules[].url.queryString

`string`

The new query string. Omit to keep the original.

### spec.rewriteRuleSets[].rewriteRules[].url.components

`enum`

Rewrite only one component. Unspecified rewrites whatever of
path/query_string is set.

Allowed values (use exactly as shown):

- `azure_application_gateway_rewrite_rule_url_component_unspecified` -- Not specified: rewrite whatever of path/query_string is set.
- `PATH_ONLY` -- Rewrite only the path (ARM: "path_only").
- `QUERY_STRING_ONLY` -- Rewrite only the query string (ARM: "query_string_only").

### spec.rewriteRuleSets[].rewriteRules[].url.reroute

`bool`

Re-evaluate the url_path_map against the REWRITTEN path, so the
rewrite can change which backend the request lands on. Azure's
default is false.

### spec.listeners

`[]AzureApplicationGatewayLayer4Listener`

Layer-4 (TCP/TLS proxy) listeners -- the non-HTTP counterpart of
http_listeners, for proxying raw TCP or TLS traffic through the
gateway.

- rule: a TLS listener requires ssl_certificate_name; a TCP listener carries neither host_names nor a certificate

### spec.listeners[].name

`string` · required

The listener's name, unique within the gateway; layer-4 routing
rules reference it.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"80"}}

### spec.listeners[].frontendIpConfigurationName

`string` · required

The frontend IP configuration (by name) the listener binds to.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.listeners[].frontendPortName

`string` · required

The frontend port (by name) the listener binds to.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.listeners[].protocol

`enum` · required

TCP proxies raw bytes; TLS terminates TLS at the gateway (requires
ssl_certificate_name).

- rule: layer-4 listeners use TCP or TLS (HTTP/HTTPS belong to http_listeners)
- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_application_gateway_protocol_unspecified` -- Not specified -- invalid; declare the protocol.
- `HTTP` -- Plain HTTP (L7).
- `HTTPS` -- TLS-terminated HTTPS (L7).
- `TCP` -- Raw TCP proxying (L4).
- `TLS` -- TLS proxying (L4).

### spec.listeners[].hostNames

`[]string`

SNI host names for TLS listeners. Forbidden for TCP.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.listeners[].sslCertificateName

`string`

The ssl_certificates entry (by name) terminating TLS. Required for
TLS, forbidden for TCP.

### spec.listeners[].sslProfileName

`string`

The ssl_profiles entry (by name) applying a TLS posture.

### spec.backends

`[]AzureApplicationGatewayLayer4BackendSettings`

Layer-4 backend settings -- the non-HTTP counterpart of
backend_http_settings.

- rule: host_name is only valid on TLS backend settings

### spec.backends[].name

`string` · required

The settings' name, unique within the gateway; layer-4 routing rules
reference it.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"80"}}

### spec.backends[].port

`int32` · required

The backend port, 1-65535.

- rule: {"required":true,"int32":{"lte":65535,"gte":1}}

### spec.backends[].protocol

`enum` · required

TCP forwards raw bytes to backends; TLS re-encrypts toward them.

- rule: layer-4 backend settings use TCP or TLS
- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_application_gateway_protocol_unspecified` -- Not specified -- invalid; declare the protocol.
- `HTTP` -- Plain HTTP (L7).
- `HTTPS` -- TLS-terminated HTTPS (L7).
- `TCP` -- Raw TCP proxying (L4).
- `TLS` -- TLS proxying (L4).

### spec.backends[].clientIpPreservationEnabled

`bool`

Preserve the client's source IP toward the backend. Azure's default
is false.

### spec.backends[].hostName

`string`

The SNI host name sent to TLS backends. Forbidden for TCP.

### spec.backends[].probeName

`string`

The custom probe (by name) health-checking these backends.

### spec.backends[].timeoutInSeconds

`int32` · optional (explicit presence)

Seconds a backend connection may idle before the gateway closes it,
1-86400. Unspecified applies Azure's default (30).

- default: `30`
- rule: {"int32":{"lte":86400,"gte":1}}

### spec.backends[].trustedRootCertificateNames

`[]string`

The trusted_root_certificates entries (by name) validating TLS
backends.

### spec.routingRules

`[]AzureApplicationGatewayLayer4RoutingRule`

Layer-4 routing rules connecting listeners to backend pools.

### spec.routingRules[].name

`string` · required

The rule's name, unique within the gateway.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"80"}}

### spec.routingRules[].listenerName

`string` · required

The layer-4 listener (by name) feeding this rule.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.routingRules[].backendAddressPoolName

`string` · required

The backend pool (by name) traffic goes to.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.routingRules[].backendSettingsName

`string` · required

The layer-4 backend settings (by name) used to reach the pool.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.routingRules[].priority

`int32` · required

Evaluation order across rules, 1-20000, unique per gateway.

- rule: {"required":true,"int32":{"lte":20000,"gte":1}}

### spec.firewallPolicyId

`string | valueFrom`

The Web Application Firewall policy the gateway enforces (WAF_V2
only), by ARM ID. Listeners and path rules can override it with their
own policies -- most specific wins.

- references: AzureWebApplicationFirewallPolicy (`status.outputs.policy_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureWebApplicationFirewallPolicy, name: <that resource's name>, fieldPath: status.outputs.policy_id}} -- a bare string does not parse

### spec.forceFirewallPolicyAssociation

`bool`

Whether the WAF policy association survives even when Azure reports
the policy in a failed state -- keeps the policy attached during
policy-side incidents.

### spec.customErrorConfigurations

`[]AzureApplicationGatewayCustomErrorConfiguration`

Gateway-wide custom error pages served when the gateway itself
generates an error response.

### spec.customErrorConfigurations[].statusCode

`enum` · required

The gateway-generated status the page replaces.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_application_gateway_custom_error_status_code_unspecified` -- Not specified -- invalid; declare the status.
- `HTTP_STATUS_400` -- 400 Bad Request.
- `HTTP_STATUS_403` -- 403 Forbidden (the page WAF blocks serve).
- `HTTP_STATUS_404` -- 404 Not Found.
- `HTTP_STATUS_405` -- 405 Method Not Allowed.
- `HTTP_STATUS_408` -- 408 Request Timeout.
- `HTTP_STATUS_500` -- 500 Internal Server Error.
- `HTTP_STATUS_502` -- 502 Bad Gateway (no healthy backends).
- `HTTP_STATUS_503` -- 503 Service Unavailable.
- `HTTP_STATUS_504` -- 504 Gateway Timeout.

### spec.customErrorConfigurations[].customErrorPageUrl

`string` · required

The publicly reachable HTML page URL (e.g. a blob URL).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.privateLinkConfigurations

`[]AzureApplicationGatewayPrivateLinkConfiguration`

Private Link configurations, enabling private endpoints in other
networks to reach the gateway's frontends without peering or public
exposure.

### spec.privateLinkConfigurations[].name

`string` · required

The configuration's name; a frontend references it via
private_link_configuration_name.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"80"}}

### spec.privateLinkConfigurations[].ipConfigurations

`[]AzureApplicationGatewayPrivateLinkIpConfiguration` · required

The NAT addresses Private Link traffic arrives through -- one or
more allocations in a subnet (typically the gateway's).

- rule: {"repeated":{"minItems":"1"}}
- rule: STATIC private_ip_address_allocation requires private_ip_address (and DYNAMIC forbids it)

### spec.privateLinkConfigurations[].ipConfigurations[].name

`string` · required

The allocation's name.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.privateLinkConfigurations[].ipConfigurations[].subnetId

`string | valueFrom` · required

The subnet the allocation takes an address in, by ARM ID.

- references: AzureSubnet (`status.outputs.subnet_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.privateLinkConfigurations[].ipConfigurations[].privateIpAddress

`string`

For STATIC allocation: the fixed address inside the subnet.

### spec.privateLinkConfigurations[].ipConfigurations[].privateIpAddressAllocation

`enum` · required

How the address is assigned.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_application_gateway_ip_allocation_unspecified` -- Not specified: DYNAMIC.
- `DYNAMIC` -- Azure assigns an address from the subnet.
- `STATIC` -- The declared private_ip_address is used.

### spec.privateLinkConfigurations[].ipConfigurations[].primary

`bool`

Whether this is the primary allocation of the configuration.

### spec.http2Enabled

`bool` · optional (explicit presence)

Whether HTTP/2 is enabled for client connections (backend
connections always use HTTP/1.1). Azure's default is false.

- default: `false`

### spec.fipsEnabled

`bool`

Whether FIPS 140-2 validated cryptographic modules process TLS --
required in some government/compliance environments.

### spec.globalConfiguration

`AzureApplicationGatewayGlobalConfiguration`

Request/response buffering. Omit for Azure's defaults (both on);
declare the block to disable buffering for streaming workloads
(both fields must then be set explicitly).

- rule: declare both request_buffering_enabled and response_buffering_enabled when the block is present

### spec.globalConfiguration.requestBufferingEnabled

`bool` · optional (explicit presence)

Whether the gateway buffers the full request before forwarding.
Azure's default (block absent) is true; false streams requests to
backends.

### spec.globalConfiguration.responseBufferingEnabled

`bool` · optional (explicit presence)

Whether the gateway buffers the full response before replying.
Azure's default (block absent) is true; false streams responses to
clients.

### spec.tags

`map<string, string>`

Free-form tags applied to the gateway, merged over the
Planton-derived resource tags (organization, environment, resource
id); a user tag with the same key wins.

## Validation Rules

- `appgw_capacity_xor_autoscale`: exactly one of capacity and autoscale must be set (BASIC supports only capacity)
- `appgw_basic_sku_gates`: the BASIC SKU supports neither autoscale, capacity above 2, trusted_client_certificates (mTLS), nor URL rewriting inside rewrite rules
- `appgw_waf_policy_requires_waf_sku`: firewall_policy_id requires the WAF_V2 SKU
- `appgw_at_least_one_listener`: declare at least one http_listeners (L7) or listeners (L4) entry
- `appgw_at_least_one_backend_settings`: declare at least one backend_http_settings (L7) or backends (L4) entry
- `appgw_at_least_one_routing_rule`: declare at least one request_routing_rules (L7) or routing_rules (L4) entry

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureApplicationGateway, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.application_gateway_id` | `string` | The Azure Resource Manager ID of the Application Gateway. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Network/applicationGateways/{name} |
| `status.outputs.application_gateway_name` | `string` | The name of the Application Gateway. |
| `status.outputs.backend_address_pool_ids` | `map<string, string>` | The ARM ID of each backend address pool, keyed by the pool's name. THE membership seam: NIC ip_configurations and scale-set network profiles reference a pool ID here to join the pool. Example valueFrom fieldPath: status.outputs.backend_address_pool_ids.web |
| `status.outputs.frontend_ip_configuration_ids` | `map<string, string>` | The ARM ID of each frontend IP configuration, keyed by the frontend's name. Example valueFrom fieldPath: status.outputs.frontend_ip_configuration_ids.public |
| `status.outputs.private_ip_address` | `string` | The first private frontend's address -- what internal DNS records point at. Empty when every frontend is public. |
| `status.outputs.private_ip_addresses` | `[]string` | The private addresses of ALL private frontends, in declaration order. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.subnetId` | AzureSubnet | `status.outputs.subnet_id` |
| `spec.identity.identityIds` | AzureUserAssignedIdentity | `status.outputs.identity_id` |
| `spec.frontendIpConfigurations[].publicIpAddressId` | AzurePublicIp | `status.outputs.public_ip_id` |
| `spec.frontendIpConfigurations[].subnetId` | AzureSubnet | `status.outputs.subnet_id` |
| `spec.httpListeners[].firewallPolicyId` | AzureWebApplicationFirewallPolicy | `status.outputs.policy_id` |
| `spec.urlPathMaps[].pathRules[].firewallPolicyId` | AzureWebApplicationFirewallPolicy | `status.outputs.policy_id` |
| `spec.sslCertificates[].keyVaultSecretId` | AzureKeyVaultCertificate | `status.outputs.versionless_secret_id` |
| `spec.trustedRootCertificates[].keyVaultSecretId` | AzureKeyVaultCertificate | `status.outputs.versionless_secret_id` |
| `spec.firewallPolicyId` | AzureWebApplicationFirewallPolicy | `status.outputs.policy_id` |
| `spec.privateLinkConfigurations[].ipConfigurations[].subnetId` | AzureSubnet | `status.outputs.subnet_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureAksCluster | `spec.ingressApplicationGateway.gatewayId` | `status.outputs.application_gateway_id` |
| AzureNetworkInterface | `spec.ipConfigurations[].applicationGatewayBackendAddressPoolIds` | `status.outputs.backend_address_pool_ids` |
| AzureVirtualMachineScaleSet | `spec.networkInterfaces[].ipConfigurations[].applicationGatewayBackendAddressPoolIds` | `status.outputs.backend_address_pool_ids` |

## See Also

- [Overview](../README.md)

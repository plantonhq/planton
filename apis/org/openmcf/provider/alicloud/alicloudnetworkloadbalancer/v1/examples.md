# AliCloudNetworkLoadBalancer Examples

## Minimal: Internet-Facing TCP

The simplest NLB configuration: one server group, one TCP listener, two availability zones.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudNetworkLoadBalancer
metadata:
  name: dev-nlb
spec:
  region: cn-hangzhou
  vpcId:
    value: vpc-abc123
  zoneMappings:
    - zoneId: cn-hangzhou-a
      vswitchId:
        value: vsw-zone-a
    - zoneId: cn-hangzhou-b
      vswitchId:
        value: vsw-zone-b
  serverGroups:
    - name: tcp-backend
      healthCheck:
        healthCheckEnabled: true
  listeners:
    - listenerPort: 80
      listenerProtocol: TCP
      serverGroupName: tcp-backend
```

## Production: TCPSSL with Mutual TLS

Internet-facing NLB with TLS termination at Layer 4, client certificate verification, and fixed EIPs for stable public addresses.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudNetworkLoadBalancer
metadata:
  name: prod-nlb
  org: acme-corp
  env: production
spec:
  region: cn-shanghai
  vpcId:
    valueFrom:
      name: prod-vpc
  zoneMappings:
    - zoneId: cn-shanghai-a
      vswitchId:
        valueFrom:
          name: prod-vswitch-a
      allocationId:
        valueFrom:
          name: prod-eip-a
    - zoneId: cn-shanghai-b
      vswitchId:
        valueFrom:
          name: prod-vswitch-b
      allocationId:
        valueFrom:
          name: prod-eip-b
  tags:
    team: platform
    cost-center: shared-infra
  serverGroups:
    - name: api-backend
      protocol: TCPSSL
      scheduler: Wlc
      connectionDrainEnabled: true
      connectionDrainTimeout: 300
      healthCheck:
        healthCheckEnabled: true
        healthCheckType: HTTP
        healthCheckUrl: /healthz
        httpCheckMethod: GET
        healthCheckInterval: 10
        healthyThreshold: 3
        unhealthyThreshold: 2
        healthCheckHttpCodes:
          - http_2xx
  listeners:
    - listenerPort: 443
      listenerProtocol: TCPSSL
      serverGroupName: api-backend
      certificateIds:
        - cas-prod-cert
      securityPolicyId: tls_cipher_policy_1_2_strict
      caCertificateIds:
        - ca-prod-cert
      caEnabled: true
      listenerDescription: Production TCPSSL with mutual TLS
```

## Internal: VPC-Private NLB with Connection Draining

An internal NLB for service-to-service TCP traffic with connection draining for graceful deployments and source-IP consistent hashing for session affinity.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudNetworkLoadBalancer
metadata:
  name: internal-nlb
spec:
  region: cn-hangzhou
  vpcId:
    value: vpc-internal
  addressType: Intranet
  crossZoneEnabled: false
  zoneMappings:
    - zoneId: cn-hangzhou-a
      vswitchId:
        value: vsw-internal-a
    - zoneId: cn-hangzhou-b
      vswitchId:
        value: vsw-internal-b
  serverGroups:
    - name: db-proxy
      scheduler: Sch
      connectionDrainEnabled: true
      connectionDrainTimeout: 60
      preserveClientIpEnabled: true
      healthCheck:
        healthCheckEnabled: true
        healthCheckType: TCP
        healthCheckConnectPort: 3306
        healthCheckInterval: 5
        healthyThreshold: 2
        unhealthyThreshold: 2
  listeners:
    - listenerPort: 3306
      listenerProtocol: TCP
      serverGroupName: db-proxy
      idleTimeout: 600
      proxyProtocolEnabled: true
      listenerDescription: MySQL proxy with connection draining
```

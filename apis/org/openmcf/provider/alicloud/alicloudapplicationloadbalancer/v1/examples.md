# AliCloudApplicationLoadBalancer Examples

## Minimal: Internet-Facing HTTP

The simplest ALB configuration: one server group, one HTTP listener, two availability zones.

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudApplicationLoadBalancer
metadata:
  name: dev-alb
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
    - name: web-backend
      healthCheckConfig:
        healthCheckEnabled: true
        healthCheckPath: /health
  listeners:
    - listenerPort: 80
      listenerProtocol: HTTP
      defaultActionServerGroupName: web-backend
```

## Production: HTTPS with WAF and Access Logs

Full production setup with HTTPS, StandardWithWaf edition, strict TLS policy, access logging to SLS, and deletion protection via tags.

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudApplicationLoadBalancer
metadata:
  name: prod-alb
  org: acme-corp
  env: production
spec:
  region: cn-shanghai
  vpcId:
    valueFrom:
      name: prod-vpc
  loadBalancerEdition: StandardWithWaf
  zoneMappings:
    - zoneId: cn-shanghai-a
      vswitchId:
        valueFrom:
          name: prod-vswitch-a
    - zoneId: cn-shanghai-b
      vswitchId:
        valueFrom:
          name: prod-vswitch-b
  accessLogConfig:
    logProject: prod-sls-project
    logStore: alb-access-log
  tags:
    team: platform
    cost-center: shared-infra
  serverGroups:
    - name: api-backend
      protocol: HTTPS
      healthCheckConfig:
        healthCheckEnabled: true
        healthCheckProtocol: HTTPS
        healthCheckPath: /healthz
        healthCheckMethod: GET
        healthCheckInterval: 5
        healthyThreshold: 5
        unhealthyThreshold: 2
      stickySessionConfig:
        stickySessionEnabled: true
        stickySessionType: Insert
        cookieTimeout: 3600
  listeners:
    - listenerPort: 443
      listenerProtocol: HTTPS
      defaultActionServerGroupName: api-backend
      certificateId: cas-prod-cert
      securityPolicyId: tls_cipher_policy_1_2_strict
      listenerDescription: Production HTTPS listener
```

## Internal: VPC-Private ALB with GRPC Backend

An internal ALB for service-to-service communication using GRPC.

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudApplicationLoadBalancer
metadata:
  name: internal-grpc-alb
spec:
  region: cn-hangzhou
  vpcId:
    value: vpc-internal
  addressType: Intranet
  loadBalancerEdition: Basic
  zoneMappings:
    - zoneId: cn-hangzhou-a
      vswitchId:
        value: vsw-internal-a
    - zoneId: cn-hangzhou-b
      vswitchId:
        value: vsw-internal-b
  serverGroups:
    - name: grpc-services
      protocol: GRPC
      scheduler: Wlc
      healthCheckConfig:
        healthCheckEnabled: true
        healthCheckProtocol: GRPC
        healthCheckInterval: 3
        healthyThreshold: 2
        unhealthyThreshold: 3
  listeners:
    - listenerPort: 443
      listenerProtocol: HTTPS
      defaultActionServerGroupName: grpc-services
      certificateId: cas-internal-cert
      http2Enabled: true
```

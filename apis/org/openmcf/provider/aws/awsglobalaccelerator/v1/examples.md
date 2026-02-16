# AWS Global Accelerator Examples

Below are several examples demonstrating how to define an AWS Global Accelerator resource in OpenMCF. After creating one of these YAML manifests, apply it with your preferred IaC engine (Pulumi or Terraform) using the OpenMCF CLI:

```shell
openmcf pulumi up --manifest <yaml-path> --stack <stack-name>
```

or

```shell
openmcf terraform apply --manifest <yaml-path> --stack <stack-name>
```

---

## 1. Minimal TCP Accelerator

The simplest useful accelerator: a single listener on port 443, one regional endpoint group, and an ALB as the target.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsGlobalAccelerator
metadata:
  name: minimal-tcp-ga
spec:
  listeners:
    - name: tcp-443
      protocol: TCP
      portRanges:
        - fromPort: 443
          toPort: 443
      endpointGroups:
        - name: primary
          endpointGroupRegion: us-east-1
          healthCheckProtocol: TCP
          endpoints:
            - endpointId: arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/my-alb/1234567890abcdef
              weight: 128
```

This example:
- Creates a Global Accelerator with two AWS-allocated static anycast IPs.
- Accepts TCP traffic on port 443 and routes it to a single ALB in `us-east-1`.
- Uses TCP health checks with 30-second intervals (defaults).
- All traffic goes to a single endpoint with weight 128 (the default midpoint).

---

## 2. Multi-Region Production Setup

A production-grade accelerator distributing traffic across two regions with HTTP health checks, flow logs, and weighted traffic distribution.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsGlobalAccelerator
metadata:
  name: prod-multi-region-ga
  labels:
    openmcf.org/provisioner: pulumi
spec:
  enabled: true
  ipAddressType: IPV4
  flowLogs:
    enabled: true
    s3Bucket: my-ga-flow-logs-bucket
    s3Prefix: ga-logs/prod/
  listeners:
    - name: tcp-443
      protocol: TCP
      clientAffinity: NONE
      portRanges:
        - fromPort: 443
          toPort: 443
      endpointGroups:
        - name: us-east-1
          endpointGroupRegion: us-east-1
          healthCheckProtocol: HTTP
          healthCheckPath: /health
          healthCheckIntervalSeconds: 10
          thresholdCount: 5
          trafficDialPercentage: 60.0
          endpoints:
            - endpointId: arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/prod-alb-us/abcdef1234567890
              weight: 128
              clientIpPreservationEnabled: true
        - name: eu-west-1
          endpointGroupRegion: eu-west-1
          healthCheckProtocol: HTTP
          healthCheckPath: /health
          healthCheckIntervalSeconds: 10
          thresholdCount: 5
          trafficDialPercentage: 40.0
          endpoints:
            - endpointId: arn:aws:elasticloadbalancing:eu-west-1:123456789012:loadbalancer/app/prod-alb-eu/fedcba0987654321
              weight: 128
              clientIpPreservationEnabled: true
```

This example:
- Routes 60% of traffic to `us-east-1` and 40% to `eu-west-1` using `trafficDialPercentage`.
- Enables HTTP health checks on `/health` with aggressive 10-second intervals and a threshold of 5 consecutive checks.
- Enables flow logs to S3 for traffic analysis and debugging.
- Preserves client IP addresses at both ALB endpoints for accurate access logging.
- If the US region becomes unhealthy, Global Accelerator automatically routes all traffic to the EU region.

---

## 3. Gaming UDP Accelerator

A UDP accelerator optimized for real-time gaming with source IP affinity and Elastic IP endpoints.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsGlobalAccelerator
metadata:
  name: game-server-ga
spec:
  listeners:
    - name: udp-game
      protocol: UDP
      clientAffinity: SOURCE_IP
      portRanges:
        - fromPort: 7000
          toPort: 8000
      endpointGroups:
        - name: us-west-2
          endpointGroupRegion: us-west-2
          healthCheckProtocol: TCP
          healthCheckPort: 8080
          healthCheckIntervalSeconds: 10
          thresholdCount: 2
          endpoints:
            - endpointId: eipalloc-0abcdef1234567890
              weight: 128
            - endpointId: eipalloc-0fedcba0987654321
              weight: 128
            - endpointId: eipalloc-0aabbccdd11223344
              weight: 64
```

This example:
- Uses UDP protocol for low-latency game traffic across port range 7000–8000.
- Enables `SOURCE_IP` affinity so each player is pinned to the same game server for the duration of their session.
- Routes to Elastic IP addresses, which are typically assigned to dedicated game server instances.
- Uses a separate TCP health check on port 8080 (a sidecar health endpoint) since UDP health checks aren't supported.
- Aggressive health checking (10-second interval, threshold of 2) detects failed game servers quickly.
- The third endpoint has a lower weight (64 vs 128), receiving approximately half the traffic of the other two.

---

## 4. Cross-Resource References with valueFrom

This example demonstrates how to wire an AwsGlobalAccelerator to other OpenMCF resources using `valueFrom` references, eliminating hardcoded ARNs.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsGlobalAccelerator
metadata:
  name: ga-with-refs
spec:
  flowLogs:
    enabled: true
    s3Bucket:
      valueFrom:
        kind: AwsS3Bucket
        name: ga-logs-bucket
        field: status.outputs.bucket_name
    s3Prefix: ga-logs/
  listeners:
    - name: tcp-443
      protocol: TCP
      portRanges:
        - fromPort: 443
          toPort: 443
      endpointGroups:
        - name: us-east-1
          endpointGroupRegion: us-east-1
          healthCheckProtocol: HTTP
          healthCheckPath: /health
          endpoints:
            - endpointId:
                valueFrom:
                  kind: AwsAlb
                  name: prod-alb-us
                  field: status.outputs.alb_arn
              weight: 128
              clientIpPreservationEnabled: true
        - name: eu-west-1
          endpointGroupRegion: eu-west-1
          healthCheckProtocol: HTTP
          healthCheckPath: /health
          endpoints:
            - endpointId:
                valueFrom:
                  kind: AwsNetworkLoadBalancer
                  name: prod-nlb-eu
                  field: status.outputs.nlb_arn
              weight: 128
```

This example demonstrates:
- **S3 bucket reference**: The flow logs bucket is resolved from an `AwsS3Bucket` resource's output, so the bucket doesn't need to be hardcoded.
- **ALB reference**: The US endpoint references an `AwsAlb` resource by name, resolving its ARN from `status.outputs.alb_arn` at deploy time.
- **NLB reference**: The EU endpoint references an `AwsNetworkLoadBalancer` resource, showing that endpoint types can mix ALBs and NLBs.
- This pattern enables composable infrastructure: changing the ALB or NLB in a separate manifest automatically propagates to the Global Accelerator on the next deployment.

---

## 5. Port Override Example

When the accelerator listens on a standard port but the backend endpoints serve on a different port.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsGlobalAccelerator
metadata:
  name: port-override-ga
spec:
  listeners:
    - name: https-override
      protocol: TCP
      portRanges:
        - fromPort: 443
          toPort: 443
        - fromPort: 80
          toPort: 80
      endpointGroups:
        - name: us-east-1
          endpointGroupRegion: us-east-1
          healthCheckProtocol: HTTP
          healthCheckPath: /health
          healthCheckPort: 8443
          endpoints:
            - endpointId: arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/my-alb/1234567890abcdef
              weight: 128
          portOverrides:
            - listenerPort: 443
              endpointPort: 8443
            - listenerPort: 80
              endpointPort: 8080
```

This example:
- Clients connect to the accelerator on ports 443 and 80 (standard HTTPS and HTTP).
- Port overrides remap: traffic arriving on port 443 is forwarded to endpoint port 8443, and port 80 is forwarded to 8080.
- Health checks target port 8443 directly, matching the actual endpoint port.
- This pattern is common when applications run on non-standard ports behind a load balancer, or during migrations where the backend port hasn't been changed yet.

---

## After Deploying

Once you've applied your manifest with OpenMCF, you can verify the accelerator:

```shell
aws globalaccelerator list-accelerators --region us-west-2
```

> **Note**: Global Accelerator is a global service but its API endpoint is in `us-west-2`. Always use `--region us-west-2` for AWS CLI commands.

To see the assigned static IPs:

```shell
aws globalaccelerator describe-accelerator \
  --accelerator-arn <accelerator-arn> \
  --region us-west-2
```

You can test connectivity by resolving the accelerator's DNS name:

```shell
dig <accelerator-dns-name>
```

The response should return the two static anycast IP addresses assigned to your accelerator.

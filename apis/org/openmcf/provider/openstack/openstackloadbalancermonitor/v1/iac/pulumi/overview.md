# OpenStackLoadBalancerMonitor Pulumi Module -- Architecture Overview

## Module Flow

```
OpenStackLoadBalancerMonitorStackInput
  |-- target: OpenStackLoadBalancerMonitor (api.proto)
  |   |-- metadata.name -> monitor name
  |   +-- spec: OpenStackLoadBalancerMonitorSpec
  |       |-- pool_id (StringValueOrRef FK -> OpenStackLoadBalancerPool)
  |       |-- type (HTTP, HTTPS, PING, TCP, TLS-HELLO, UDP-CONNECT)
  |       |-- delay (seconds between checks)
  |       |-- timeout (seconds per check)
  |       |-- max_retries (1-10, consecutive successes)
  |       |-- max_retries_down (optional, 1-10, consecutive failures)
  |       |-- url_path (HTTP/HTTPS only)
  |       |-- http_method (HTTP/HTTPS only)
  |       |-- expected_codes (HTTP/HTTPS only)
  |       |-- admin_state_up (default: true)
  |       +-- region
  +-- provider_config: OpenStackProviderConfig

         |
         v

  initializeLocals()
  |-- Resolve pool_id from StringValueOrRef -> locals.PoolId
  +-- Store references for monitor()

         |
         v

  monitor()
  |-- Map spec fields -> loadbalancer.MonitorArgs
  |-- Handle optional fields (max_retries_down, url_path, http_method, expected_codes)
  |-- Note: NO tags support (TF provider limitation)
  |-- loadbalancer.NewMonitor()
  +-- Export outputs: monitor_id, name, type, pool_id, region
```

## Resource Mapping

| Spec Field | Pulumi MonitorArgs Field | Behavior |
|---|---|---|
| `pool_id` | `PoolId` | Required. Resolved from StringValueOrRef |
| `type` | `Type` | Required. Passed directly |
| `delay` | `Delay` | Required. Passed directly |
| `timeout` | `Timeout` | Required. Passed directly |
| `max_retries` | `MaxRetries` | Required. Passed directly |
| `max_retries_down` | `MaxRetriesDown` | Set when present (optional int32) |
| `url_path` | `UrlPath` | Set when non-empty (HTTP/HTTPS only) |
| `http_method` | `HttpMethod` | Set when non-empty (HTTP/HTTPS only) |
| `expected_codes` | `ExpectedCodes` | Set when non-empty (HTTP/HTTPS only) |
| `admin_state_up` | `AdminStateUp` | Set when present (default: true via middleware) |
| `region` | `Region` | Set when non-empty |

## Outputs

All outputs match the `OpenStackLoadBalancerMonitorStackOutputs` proto message fields:

| Output Key | Source |
|---|---|
| `monitor_id` | `createdMonitor.ID()` |
| `name` | `createdMonitor.Name` |
| `type` | `createdMonitor.Type` |
| `pool_id` | `createdMonitor.PoolId` |
| `region` | `createdMonitor.Region` |

## Important Note

Health monitors do **NOT** support tags in the Terraform OpenStack provider.
This is a provider limitation, not an OpenMCF design choice.

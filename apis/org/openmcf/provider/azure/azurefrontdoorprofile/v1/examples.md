# Azure Front Door Profile - Examples

## Minimal Single Endpoint

The simplest configuration: a Standard-tier Front Door profile with one endpoint, one origin group with a single web backend, and a catch-all route. Suitable for accelerating a single web application.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureFrontDoorProfile
metadata:
  name: web-cdn
spec:
  resourceGroup:
    value: "production-rg"
  name: myapp-cdn-fd
  endpoints:
    - name: main-endpoint
  originGroups:
    - name: web-backend
      origins:
        - name: app-service
          hostName: myapp.azurewebsites.net
          originHostHeader: myapp.azurewebsites.net
  routes:
    - name: catch-all
      endpointName: main-endpoint
      originGroupName: web-backend
      patternsToMatch:
        - "/*"
      supportedProtocols:
        - Http
        - Https
```

## Multi-Origin with Health Probes

A Standard-tier profile with two origins in different regions and HTTPS health probes. Traffic is split by weight, and unhealthy origins are automatically removed from rotation.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureFrontDoorProfile
metadata:
  name: multi-region-app
  org: mycompany
  env: production
spec:
  resourceGroup:
    value: "prod-networking-rg"
  name: myapp-multiregion-fd
  endpoints:
    - name: app-endpoint
  originGroups:
    - name: regional-backends
      sessionAffinityEnabled: false
      loadBalancing:
        sampleSize: 4
        successfulSamplesRequired: 3
        additionalLatencyInMilliseconds: 50
      healthProbe:
        protocol: Https
        path: /health
        requestType: GET
        intervalInSeconds: 30
      origins:
        - name: east-us
          hostName: myapp-east.azurewebsites.net
          originHostHeader: myapp-east.azurewebsites.net
          priority: 1
          weight: 500
        - name: west-us
          hostName: myapp-west.azurewebsites.net
          originHostHeader: myapp-west.azurewebsites.net
          priority: 1
          weight: 500
  routes:
    - name: app-route
      endpointName: app-endpoint
      originGroupName: regional-backends
      patternsToMatch:
        - "/*"
      supportedProtocols:
        - Http
        - Https
      forwardingProtocol: HttpsOnly
```

## Static Site with Caching

A Standard-tier profile serving a static site from Azure Blob Storage with aggressive caching and compression enabled for common web content types.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureFrontDoorProfile
metadata:
  name: static-site-cdn
  org: mycompany
  env: production
spec:
  resourceGroup:
    value: "static-sites-rg"
  name: mysite-static-fd
  endpoints:
    - name: site-endpoint
  originGroups:
    - name: blob-storage
      sessionAffinityEnabled: false
      healthProbe:
        protocol: Https
        path: /index.html
        requestType: HEAD
        intervalInSeconds: 60
      origins:
        - name: storage-origin
          hostName: mystaticsite.blob.core.windows.net
          originHostHeader: mystaticsite.blob.core.windows.net
  routes:
    - name: static-assets
      endpointName: site-endpoint
      originGroupName: blob-storage
      patternsToMatch:
        - "/*"
      supportedProtocols:
        - Http
        - Https
      forwardingProtocol: HttpsOnly
      cache:
        queryStringCachingBehavior: IgnoreQueryString
        compressionEnabled: true
        contentTypesToCompress:
          - text/html
          - text/css
          - text/javascript
          - application/javascript
          - application/json
          - application/xml
          - image/svg+xml
          - font/woff2
```

## API Gateway with Path Routing

A Standard-tier profile with multiple routes directing different URL paths to different origin groups. API traffic goes to a container app backend, while static assets go to blob storage.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureFrontDoorProfile
metadata:
  name: api-gateway
  org: mycompany
  env: production
spec:
  resourceGroup:
    value: "prod-gateway-rg"
  name: myapp-gateway-fd
  responseTimeoutSeconds: 60
  endpoints:
    - name: gateway-endpoint
  originGroups:
    - name: api-backend
      sessionAffinityEnabled: false
      loadBalancing:
        sampleSize: 4
        successfulSamplesRequired: 3
      healthProbe:
        protocol: Https
        path: /api/healthz
        requestType: GET
        intervalInSeconds: 15
      origins:
        - name: api-service
          hostName: myapi.eastus.azurecontainerapps.io
          originHostHeader: myapi.eastus.azurecontainerapps.io
    - name: static-backend
      sessionAffinityEnabled: false
      origins:
        - name: static-storage
          hostName: myassets.blob.core.windows.net
          originHostHeader: myassets.blob.core.windows.net
  routes:
    - name: api-route
      endpointName: gateway-endpoint
      originGroupName: api-backend
      patternsToMatch:
        - "/api/*"
      supportedProtocols:
        - Http
        - Https
      forwardingProtocol: HttpsOnly
    - name: static-route
      endpointName: gateway-endpoint
      originGroupName: static-backend
      patternsToMatch:
        - "/static/*"
      supportedProtocols:
        - Http
        - Https
      forwardingProtocol: HttpsOnly
      cache:
        queryStringCachingBehavior: IgnoreQueryString
        compressionEnabled: true
        contentTypesToCompress:
          - text/css
          - text/javascript
          - application/javascript
          - image/svg+xml
```

## Premium with Private Link

A Premium-tier profile connecting to an Azure App Service origin via Private Link. The origin is not exposed to the public internet -- Front Door reaches it exclusively through Azure's backbone network.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureFrontDoorProfile
metadata:
  name: enterprise-cdn
  org: enterprise-corp
  env: production
spec:
  resourceGroup:
    value: "enterprise-networking-rg"
  name: enterprise-app-fd
  sku: Premium_AzureFrontDoor
  responseTimeoutSeconds: 120
  endpoints:
    - name: secure-endpoint
  originGroups:
    - name: private-backend
      sessionAffinityEnabled: true
      loadBalancing:
        sampleSize: 4
        successfulSamplesRequired: 3
        additionalLatencyInMilliseconds: 50
      healthProbe:
        protocol: Https
        path: /health
        requestType: GET
        intervalInSeconds: 30
      origins:
        - name: app-service
          hostName: enterprise-app.azurewebsites.net
          originHostHeader: enterprise-app.azurewebsites.net
          certificateNameCheckEnabled: true
          privateLink:
            location: eastus
            privateLinkTargetId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/enterprise-rg/providers/Microsoft.Web/sites/enterprise-app
            requestMessage: "Front Door CDN Private Link to App Service"
            targetType: sites
  routes:
    - name: secure-route
      endpointName: secure-endpoint
      originGroupName: private-backend
      patternsToMatch:
        - "/*"
      supportedProtocols:
        - Http
        - Https
      forwardingProtocol: HttpsOnly
```

## Infra Chart valueFrom Reference

When used within an infra chart, the profile references its resource group from an upstream AzureResourceGroup component via `valueFrom`.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureFrontDoorProfile
metadata:
  name: "{{ values.env }}-cdn"
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: "{{ values.env }}-rg"
      fieldPath: status.outputs.resource_group_name
  name: "{{ values.app_name }}-{{ values.env }}-fd"
  endpoints:
    - name: "{{ values.endpoint_name }}"
  originGroups:
    - name: "{{ values.origin_group_name }}"
      healthProbe:
        protocol: Https
        path: /health
        requestType: HEAD
        intervalInSeconds: 30
      origins:
        - name: "{{ values.origin_name }}"
          hostName: "{{ values.origin_hostname }}"
          originHostHeader: "{{ values.origin_hostname }}"
  routes:
    - name: "{{ values.route_name }}"
      endpointName: "{{ values.endpoint_name }}"
      originGroupName: "{{ values.origin_group_name }}"
      patternsToMatch:
        - "/*"
      supportedProtocols:
        - Http
        - Https
```

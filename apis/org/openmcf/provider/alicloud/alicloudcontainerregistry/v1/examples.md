# AlicloudContainerRegistry Examples

## Basic Development Registry

A minimal registry for development and testing with a single namespace:

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudContainerRegistry
metadata:
  name: dev-registry
spec:
  region: cn-hangzhou
  instanceName: dev-acr
  instanceType: Basic
  paymentType: PayAsYouGo
  namespaces:
    - name: dev
      autoCreate: true
      defaultVisibility: PRIVATE
```

## Standard Production Registry

A production-ready registry with multiple namespaces organized by team:

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudContainerRegistry
metadata:
  name: prod-registry
  org: acme-corp
  env: production
spec:
  region: cn-shanghai
  instanceName: prod-acr
  instanceType: Standard
  paymentType: Subscription
  period: 12
  password: "MyStr0ng!RegistryPass"
  namespaces:
    - name: platform
      autoCreate: true
      defaultVisibility: PRIVATE
    - name: backend
      autoCreate: true
      defaultVisibility: PRIVATE
    - name: frontend
      autoCreate: false
      defaultVisibility: PRIVATE
```

## Advanced Enterprise Registry with Resource Group

An enterprise-grade registry with resource group assignment and public-facing namespace:

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudContainerRegistry
metadata:
  name: enterprise-registry
  org: global-corp
  env: production
spec:
  region: cn-beijing
  instanceName: enterprise-acr
  instanceType: Advanced
  paymentType: Subscription
  period: 12
  resourceGroupId: rg-acfm1234567
  namespaces:
    - name: internal
      autoCreate: true
      defaultVisibility: PRIVATE
    - name: shared
      autoCreate: false
      defaultVisibility: PRIVATE
    - name: public-images
      autoCreate: false
      defaultVisibility: PUBLIC
```

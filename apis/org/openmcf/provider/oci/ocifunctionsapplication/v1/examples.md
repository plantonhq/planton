# OciFunctionsApplication Examples

## Minimal Application

An application with default x86 architecture in a single subnet:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciFunctionsApplication
metadata:
  name: my-app
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciFunctionsApplication.my-app
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  subnetIds:
    - value: "ocid1.subnet.oc1..example"
```

## ARM Application with Config and NSGs

An application on Ampere A1 processors with shared environment variables and network security group bindings. References infrastructure via `valueFrom`:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciFunctionsApplication
metadata:
  name: arm-app
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciFunctionsApplication.arm-app
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-compartment
      fieldPath: status.outputs.compartmentId
  subnetIds:
    - valueFrom:
        kind: OciSubnet
        name: fn-subnet
        fieldPath: status.outputs.subnetId
  shape: generic_arm
  config:
    LOG_LEVEL: "info"
    DB_ENDPOINT: "adb.us-ashburn-1.oraclecloud.com"
  networkSecurityGroupIds:
    - valueFrom:
        kind: OciSecurityGroup
        name: fn-nsg
        fieldPath: status.outputs.networkSecurityGroupId
```

## Image Signature Verification

An application that only allows deployment of images signed by a specific KMS key:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciFunctionsApplication
metadata:
  name: secure-app
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciFunctionsApplication.secure-app
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  subnetIds:
    - value: "ocid1.subnet.oc1..example"
  imagePolicyConfig:
    isPolicyEnabled: true
    keyDetails:
      - kmsKeyId:
          valueFrom:
            kind: OciKmsKey
            name: signing-key
            fieldPath: status.outputs.keyId
```

## Multi-Architecture with APM Tracing and Syslog

A production application supporting both x86 and ARM, with APM distributed tracing and syslog forwarding:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciFunctionsApplication
metadata:
  name: traced-app
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciFunctionsApplication.traced-app
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  subnetIds:
    - value: "ocid1.subnet.oc1..example-1"
    - value: "ocid1.subnet.oc1..example-2"
  shape: generic_x86_arm
  config:
    APP_ENV: "production"
    FEATURE_FLAGS_URL: "https://config.example.com/flags"
  syslogUrl: "tcp://logserver.example.com:514"
  traceConfig:
    isEnabled: true
    domainId: "ocid1.apmdomain.oc1..example"
```

## Common Operations

### Update application configuration

Modify the `config` map and re-apply. Config changes take effect on the next function invocation (functions read config at cold-start).

### Add a network security group

Append a new entry to `networkSecurityGroupIds` and re-apply. The NSG is added to the application's network configuration.

### Deploy a function to the application

After the application is created, deploy functions using the `fn` CLI:

```shell
fn deploy --app <application_id from stack outputs> --local
```

## Best Practices

1. **Use ARM for cost savings** — Ampere A1 functions are typically cheaper than x86 equivalents.
2. **Use `generic_x86_arm` for CI/CD flexibility** — allows deploying images built on either architecture.
3. **Enable image verification in production** — prevents deployment of unsigned or tampered container images.
4. **Keep config small** — the 4 KB limit means large configurations should use OCI Vault secrets or external config services.
5. **Use `valueFrom` references** for `subnetIds` and `networkSecurityGroupIds` — avoids hardcoding OCIDs and maintains dependency ordering.

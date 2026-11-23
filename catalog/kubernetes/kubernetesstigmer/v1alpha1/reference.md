# KubernetesStigmer

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**KubernetesStigmerSpec** installs Stigmer, the open-source AI agent
platform, on a Kubernetes cluster: the Stigmer operator (the
`stigmer-operator` Helm chart, oci://ghcr.io/stigmer/charts) and the one
`stigmer.ai/v1 Stigmer` resource it reconciles into a running install (the
server, the shared runner, Postgres through CloudNativePG, Temporal, the
front door, sign-in and per-session sandboxes, each only when declared).

WHY ONE KIND: Stigmer allows one operator and one Stigmer per cluster, and
the operator's version IS Stigmer's version (the resource has no version
field of its own). So the version is declared, upgraded and removed in the
one place the install is declared. A cluster already running Stigmer's
operator, or a Stigmer this kind did not declare, is refused before
anything is applied.

WHAT THE MODULES DO, in order: create the operator's namespace when asked;
install the chart; write the Secrets this spec gives values for (the model
key, the bucket's key pair, the database password, SMTP credentials, the
license, Temporal's key and certificates, extra secret settings) under the
exact key names the operator reads, and only the keys declared; check that
a restore has the encryption keys it needs; declare the Stigmer. Keys this
spec leaves unset are not rendered, so the operator's own defaults stay
authoritative.

DESTROY removes the Stigmer first, then waits, bounded, for the operator's
own cleanup (the bundled database and its volumes, the sandbox namespace,
and the prerequisites it installed: Agent Substrate, agent-sandbox, the
Barman Cloud plugin, CloudNativePG), then removes the Secrets and the
chart. If the operator cannot finish (no ready replica, or the wait runs
out), the destroy continues and prints what stays on the cluster and the
commands that remove it. The operator keeps three things on purpose: the
keys Secret `stigmer-keys` (so a backup stays restorable), the bucket's
archive, and an issued certificate. The chart keeps the `stigmers.stigmer.ai`
definition (`helm.sh/resource-policy: keep`).

KEEP THE KEYS: `stigmer-keys` holds the encryption key every stored secret
and every backup is sealed with. Deleting the Stigmer's namespace deletes
it, and with it the ability to read any backup. Save it outside the
cluster with `keys_export_command` (an output) before you delete the
namespace or the cluster.

Validated against Stigmer 0.3.2.

## Example

```yaml
# A complete install for the reference page and the e2e framework: the
# operator in its own namespace, files and backups in an S3-compatible
# bucket, a front door with a certificate cert-manager issues, the bundled
# sign-in with sign-up for one domain, and each session in a sandbox of its
# own. Every reference is written as a literal, as the e2e manifest must be.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesStigmer
metadata:
  name: stigmer
spec:
  namespace:
    value: stigmer
  version: 0.3.2
  operator:
    namespace:
      value: stigmer-system
    create_namespace: true
    resources:
      requests:
        cpu: 50m
        memory: 128Mi
      limits:
        memory: 512Mi
    log_level: info
  owner:
    email: owner@acme.example
    name: Priya
  server:
    resources:
      requests:
        cpu: 500m
        memory: 1Gi
    storage:
      size: 5Gi
    help:
      url: https://help.acme.example/stigmer
      label: Ask the platform team
    internal_accounts:
      - "*@acme.example"
  runner:
    model:
      keys:
        anthropic_api_key: sk-ant-example-not-a-real-key
    storage:
      size: 10Gi
  database:
    bundled:
      instances: 2
      storage:
        size: 20Gi
      backup:
        schedule: "0 0 2 * * *"
        retention: 30d
  object_store:
    s3:
      bucket:
        value: stigmer
      endpoint:
        value: http://seaweedfs-s3.storage.svc.cluster.local:8333
      credentials:
        access_key:
          access_key_id: stigmer
          secret_access_key: example-secret-access-key
  ingress:
    hostname: stigmer.acme.example
    class_name: nginx
    annotations:
      nginx.ingress.kubernetes.io/backend-protocol: GRPC
    tls:
      issuer:
        name:
          value: letsencrypt
        kind: ClusterIssuer
  identity:
    bundled:
      sign_up_email_domains:
        - acme.example
      smtp:
        host: smtp.acme.example
        from: stigmer@acme.example
        starttls: true
        credentials:
          login:
            username: stigmer
            password: example-smtp-password
  sandboxes:
    driver: agentSandbox
    idle_seconds: 300
    agent_sandbox:
      egress_allow_cidrs:
        - 10.20.0.0/16
  prerequisites:
    postgres_operator: auto
    postgres_backup_plugin: auto
    agent_sandbox: auto
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.version` | `string` | yes |  |  |
| `spec.operator` | `KubernetesStigmerOperator` | yes |  |  |
| `spec.operator.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.operator.createNamespace` | `bool` |  |  |  |
| `spec.operator.imageRepository` | `string` |  |  |  |
| `spec.operator.imagePullPolicy` | `string` |  |  |  |
| `spec.operator.imagePullSecrets` | `[]string` |  |  |  |
| `spec.operator.replicas` | `int32` |  |  |  |
| `spec.operator.resources` | `KubernetesStigmerOperatorResources` |  |  |  |
| `spec.operator.resources.requests` | `map<string, string>` |  |  |  |
| `spec.operator.resources.limits` | `map<string, string>` |  |  |  |
| `spec.operator.logLevel` | `string` |  |  |  |
| `spec.operator.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.operator.tolerations` | `[]KubernetesCoreV1ApiToleration` |  |  |  |
| `spec.operator.tolerations[].key` | `string` |  |  |  |
| `spec.operator.tolerations[].operator` | `string` |  |  |  |
| `spec.operator.tolerations[].value` | `string` |  |  |  |
| `spec.operator.tolerations[].effect` | `string` |  |  |  |
| `spec.operator.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.image` | `KubernetesStigmerImage` |  |  |  |
| `spec.image.registry` | `string` |  |  |  |
| `spec.image.pullPolicy` | `string` |  |  |  |
| `spec.image.pullSecrets` | `[]string` |  |  |  |
| `spec.owner` | `KubernetesStigmerOwner` |  |  |  |
| `spec.owner.email` | `string` | yes |  |  |
| `spec.owner.name` | `string` |  |  |  |
| `spec.server` | `KubernetesStigmerServer` |  |  |  |
| `spec.server.resources` | `KubernetesCoreV1ApiResourceRequirements` |  |  |  |
| `spec.server.resources.limits` | `map<string, string>` |  |  |  |
| `spec.server.resources.requests` | `map<string, string>` |  |  |  |
| `spec.server.resources.claims` | `[]KubernetesCoreV1ApiResourceClaim` |  |  |  |
| `spec.server.resources.claims[].name` | `string` | yes |  |  |
| `spec.server.resources.claims[].request` | `string` |  |  |  |
| `spec.server.storage` | `KubernetesStigmerVolume` |  |  |  |
| `spec.server.storage.size` | `string` |  |  |  |
| `spec.server.storage.storageClassName` | `string` |  |  |  |
| `spec.server.artifacts` | `KubernetesStigmerVolume` |  |  |  |
| `spec.server.artifacts.size` | `string` |  |  |  |
| `spec.server.artifacts.storageClassName` | `string` |  |  |  |
| `spec.server.help` | `KubernetesStigmerHelp` |  |  |  |
| `spec.server.help.url` | `string` | yes |  |  |
| `spec.server.help.label` | `string` |  |  |  |
| `spec.server.internalAccounts` | `[]string` |  |  |  |
| `spec.server.allowUnauthenticatedExposure` | `bool` |  |  |  |
| `spec.server.egress` | `KubernetesStigmerServerEgress` |  |  |  |
| `spec.server.egress.fenced` | `bool` |  |  |  |
| `spec.server.egress.allowCidrs` | `[]string` |  |  |  |
| `spec.server.extraEnv` | `ContainerEnv` |  |  |  |
| `spec.server.extraEnv.variables` | `[]EnvVar` (no secrets: use `secrets`) |  |  |  |
| `spec.server.extraEnv.variables[].name` | `string` | yes |  |  |
| `spec.server.extraEnv.variables[].value` | `string` |  |  |  |
| `spec.server.extraEnv.variables[].valueFrom` | `ValueFromRef` |  |  |  |
| `spec.server.extraEnv.variables[].valueFrom.kind` | `enum` |  |  |  |
| `spec.server.extraEnv.variables[].valueFrom.env` | `string` |  |  |  |
| `spec.server.extraEnv.variables[].valueFrom.name` | `string` | yes |  |  |
| `spec.server.extraEnv.variables[].valueFrom.fieldPath` | `string` |  |  |  |
| `spec.server.extraEnv.variables[].configMapKeyRef` | `ConfigMapKeyRef` |  |  |  |
| `spec.server.extraEnv.variables[].configMapKeyRef.name` | `string` | yes |  |  |
| `spec.server.extraEnv.variables[].configMapKeyRef.key` | `string` | yes |  |  |
| `spec.server.extraEnv.variables[].configMapKeyRef.optional` | `bool` |  |  |  |
| `spec.server.extraEnv.variables[].fieldRef` | `ObjectFieldRef` |  |  |  |
| `spec.server.extraEnv.variables[].fieldRef.apiVersion` | `string` |  |  |  |
| `spec.server.extraEnv.variables[].fieldRef.fieldPath` | `string` | yes |  |  |
| `spec.server.extraEnv.variables[].resourceFieldRef` | `ResourceFieldRef` |  |  |  |
| `spec.server.extraEnv.variables[].resourceFieldRef.containerName` | `string` |  |  |  |
| `spec.server.extraEnv.variables[].resourceFieldRef.resource` | `string` | yes |  |  |
| `spec.server.extraEnv.variables[].resourceFieldRef.divisor` | `string` |  |  |  |
| `spec.server.extraEnv.secrets` | `[]SecretEnvVar` |  |  |  |
| `spec.server.extraEnv.secrets[].name` | `string` | yes |  |  |
| `spec.server.extraEnv.secrets[].value` | `string` (sensitive) |  |  |  |
| `spec.server.extraEnv.secrets[].secretRef` | `KubernetesSecretKeyRef` |  |  |  |
| `spec.server.extraEnv.secrets[].secretRef.namespace` | `string` |  |  |  |
| `spec.server.extraEnv.secrets[].secretRef.name` | `string` | yes |  |  |
| `spec.server.extraEnv.secrets[].secretRef.key` | `string` | yes |  |  |
| `spec.server.extraEnv.secrets[].secretRef.optional` | `bool` |  |  |  |
| `spec.server.extraEnv.secrets[].valueFrom` | `ValueFromRef` |  |  |  |
| `spec.server.extraEnv.secrets[].valueFrom.kind` | `enum` |  |  |  |
| `spec.server.extraEnv.secrets[].valueFrom.env` | `string` |  |  |  |
| `spec.server.extraEnv.secrets[].valueFrom.name` | `string` | yes |  |  |
| `spec.server.extraEnv.secrets[].valueFrom.fieldPath` | `string` |  |  |  |
| `spec.server.extraEnv.envFrom` | `[]EnvFromSource` |  |  |  |
| `spec.server.extraEnv.envFrom[].prefix` | `string` |  |  |  |
| `spec.server.extraEnv.envFrom[].configMapRef` | `ConfigMapRef` |  |  |  |
| `spec.server.extraEnv.envFrom[].configMapRef.name` | `string` | yes |  |  |
| `spec.server.extraEnv.envFrom[].configMapRef.optional` | `bool` |  |  |  |
| `spec.server.extraEnv.envFrom[].secretRef` | `SecretRef` |  |  |  |
| `spec.server.extraEnv.envFrom[].secretRef.name` | `string` | yes |  |  |
| `spec.server.extraEnv.envFrom[].secretRef.optional` | `bool` |  |  |  |
| `spec.runner` | `KubernetesStigmerRunner` |  |  |  |
| `spec.runner.model` | `KubernetesStigmerModel` |  |  |  |
| `spec.runner.model.keys` | `KubernetesStigmerModelKeys` |  |  |  |
| `spec.runner.model.keys.anthropicApiKey` | `string` (sensitive) |  |  |  |
| `spec.runner.model.keys.anthropicBaseUrl` | `string` |  |  |  |
| `spec.runner.model.keys.openaiApiKey` | `string` (sensitive) |  |  |  |
| `spec.runner.model.keys.anthropicFoundryApiKey` | `string` (sensitive) |  |  |  |
| `spec.runner.model.keys.awsBearerTokenBedrock` | `string` (sensitive) |  |  |  |
| `spec.runner.model.existingSecretName` | `string \| valueFrom` |  |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.runner.resources` | `KubernetesCoreV1ApiResourceRequirements` |  |  |  |
| `spec.runner.resources.limits` | `map<string, string>` |  |  |  |
| `spec.runner.resources.requests` | `map<string, string>` |  |  |  |
| `spec.runner.resources.claims` | `[]KubernetesCoreV1ApiResourceClaim` |  |  |  |
| `spec.runner.resources.claims[].name` | `string` | yes |  |  |
| `spec.runner.resources.claims[].request` | `string` |  |  |  |
| `spec.runner.storage` | `KubernetesStigmerVolume` |  |  |  |
| `spec.runner.storage.size` | `string` |  |  |  |
| `spec.runner.storage.storageClassName` | `string` |  |  |  |
| `spec.runner.extraEnv` | `ContainerEnv` |  |  |  |
| `spec.runner.extraEnv.variables` | `[]EnvVar` (no secrets: use `secrets`) |  |  |  |
| `spec.runner.extraEnv.variables[].name` | `string` | yes |  |  |
| `spec.runner.extraEnv.variables[].value` | `string` |  |  |  |
| `spec.runner.extraEnv.variables[].valueFrom` | `ValueFromRef` |  |  |  |
| `spec.runner.extraEnv.variables[].valueFrom.kind` | `enum` |  |  |  |
| `spec.runner.extraEnv.variables[].valueFrom.env` | `string` |  |  |  |
| `spec.runner.extraEnv.variables[].valueFrom.name` | `string` | yes |  |  |
| `spec.runner.extraEnv.variables[].valueFrom.fieldPath` | `string` |  |  |  |
| `spec.runner.extraEnv.variables[].configMapKeyRef` | `ConfigMapKeyRef` |  |  |  |
| `spec.runner.extraEnv.variables[].configMapKeyRef.name` | `string` | yes |  |  |
| `spec.runner.extraEnv.variables[].configMapKeyRef.key` | `string` | yes |  |  |
| `spec.runner.extraEnv.variables[].configMapKeyRef.optional` | `bool` |  |  |  |
| `spec.runner.extraEnv.variables[].fieldRef` | `ObjectFieldRef` |  |  |  |
| `spec.runner.extraEnv.variables[].fieldRef.apiVersion` | `string` |  |  |  |
| `spec.runner.extraEnv.variables[].fieldRef.fieldPath` | `string` | yes |  |  |
| `spec.runner.extraEnv.variables[].resourceFieldRef` | `ResourceFieldRef` |  |  |  |
| `spec.runner.extraEnv.variables[].resourceFieldRef.containerName` | `string` |  |  |  |
| `spec.runner.extraEnv.variables[].resourceFieldRef.resource` | `string` | yes |  |  |
| `spec.runner.extraEnv.variables[].resourceFieldRef.divisor` | `string` |  |  |  |
| `spec.runner.extraEnv.secrets` | `[]SecretEnvVar` |  |  |  |
| `spec.runner.extraEnv.secrets[].name` | `string` | yes |  |  |
| `spec.runner.extraEnv.secrets[].value` | `string` (sensitive) |  |  |  |
| `spec.runner.extraEnv.secrets[].secretRef` | `KubernetesSecretKeyRef` |  |  |  |
| `spec.runner.extraEnv.secrets[].secretRef.namespace` | `string` |  |  |  |
| `spec.runner.extraEnv.secrets[].secretRef.name` | `string` | yes |  |  |
| `spec.runner.extraEnv.secrets[].secretRef.key` | `string` | yes |  |  |
| `spec.runner.extraEnv.secrets[].secretRef.optional` | `bool` |  |  |  |
| `spec.runner.extraEnv.secrets[].valueFrom` | `ValueFromRef` |  |  |  |
| `spec.runner.extraEnv.secrets[].valueFrom.kind` | `enum` |  |  |  |
| `spec.runner.extraEnv.secrets[].valueFrom.env` | `string` |  |  |  |
| `spec.runner.extraEnv.secrets[].valueFrom.name` | `string` | yes |  |  |
| `spec.runner.extraEnv.secrets[].valueFrom.fieldPath` | `string` |  |  |  |
| `spec.runner.extraEnv.envFrom` | `[]EnvFromSource` |  |  |  |
| `spec.runner.extraEnv.envFrom[].prefix` | `string` |  |  |  |
| `spec.runner.extraEnv.envFrom[].configMapRef` | `ConfigMapRef` |  |  |  |
| `spec.runner.extraEnv.envFrom[].configMapRef.name` | `string` | yes |  |  |
| `spec.runner.extraEnv.envFrom[].configMapRef.optional` | `bool` |  |  |  |
| `spec.runner.extraEnv.envFrom[].secretRef` | `SecretRef` |  |  |  |
| `spec.runner.extraEnv.envFrom[].secretRef.name` | `string` | yes |  |  |
| `spec.runner.extraEnv.envFrom[].secretRef.optional` | `bool` |  |  |  |
| `spec.database` | `KubernetesStigmerDatabase` |  |  |  |
| `spec.database.bundled` | `KubernetesStigmerBundledPostgres` |  |  |  |
| `spec.database.bundled.instances` | `int32` |  | `1` |  |
| `spec.database.bundled.storage` | `KubernetesStigmerVolume` |  |  |  |
| `spec.database.bundled.storage.size` | `string` |  |  |  |
| `spec.database.bundled.storage.storageClassName` | `string` |  |  |  |
| `spec.database.bundled.resources` | `KubernetesCoreV1ApiResourceRequirements` |  |  |  |
| `spec.database.bundled.resources.limits` | `map<string, string>` |  |  |  |
| `spec.database.bundled.resources.requests` | `map<string, string>` |  |  |  |
| `spec.database.bundled.resources.claims` | `[]KubernetesCoreV1ApiResourceClaim` |  |  |  |
| `spec.database.bundled.resources.claims[].name` | `string` | yes |  |  |
| `spec.database.bundled.resources.claims[].request` | `string` |  |  |  |
| `spec.database.bundled.backup` | `KubernetesStigmerPostgresBackup` |  |  |  |
| `spec.database.bundled.backup.schedule` | `string` |  |  |  |
| `spec.database.bundled.backup.retention` | `string` |  |  |  |
| `spec.database.bundled.recoverFrom` | `KubernetesStigmerRecoverFrom` |  |  |  |
| `spec.database.bundled.recoverFrom.serverName` | `string` | yes |  |  |
| `spec.database.bundled.recoverFrom.targetTime` | `string` |  |  |  |
| `spec.database.external` | `KubernetesStigmerExternalPostgres` |  |  |  |
| `spec.database.external.host` | `string \| valueFrom` | yes |  | KubernetesPostgres (`status.outputs.rw_service`) |
| `spec.database.external.port` | `int32` |  | `5432` |  |
| `spec.database.external.user` | `string` | yes |  |  |
| `spec.database.external.database` | `string` |  | `stigmer` |  |
| `spec.database.external.password` | `KubernetesStigmerPassword` | yes |  |  |
| `spec.database.external.password.value` | `string` (sensitive) |  |  |  |
| `spec.database.external.password.existingSecret` | `KubernetesStigmerPostgresPasswordSecret` |  |  |  |
| `spec.database.external.password.existingSecret.name` | `string \| valueFrom` | yes |  | KubernetesPostgres (`status.outputs.password_secret.name`) |
| `spec.database.external.password.existingSecret.key` | `string` |  | `password` |  |
| `spec.database.external.sslMode` | `string` |  |  |  |
| `spec.temporal` | `KubernetesStigmerTemporal` |  |  |  |
| `spec.temporal.bundled` | `KubernetesStigmerBundledTemporal` |  |  |  |
| `spec.temporal.bundled.resources` | `KubernetesCoreV1ApiResourceRequirements` |  |  |  |
| `spec.temporal.bundled.resources.limits` | `map<string, string>` |  |  |  |
| `spec.temporal.bundled.resources.requests` | `map<string, string>` |  |  |  |
| `spec.temporal.bundled.resources.claims` | `[]KubernetesCoreV1ApiResourceClaim` |  |  |  |
| `spec.temporal.bundled.resources.claims[].name` | `string` | yes |  |  |
| `spec.temporal.bundled.resources.claims[].request` | `string` |  |  |  |
| `spec.temporal.external` | `KubernetesStigmerExternalTemporal` |  |  |  |
| `spec.temporal.external.hostPort` | `string \| valueFrom` | yes |  | KubernetesTemporal (`status.outputs.frontend_endpoint`) |
| `spec.temporal.external.namespace` | `string` |  | `default` |  |
| `spec.temporal.external.tls` | `KubernetesStigmerTemporalTls` |  |  |  |
| `spec.temporal.external.tls.serverName` | `string` |  |  |  |
| `spec.temporal.external.tls.pem` | `KubernetesStigmerTemporalTlsPem` |  |  |  |
| `spec.temporal.external.tls.pem.caCertificate` | `string` |  |  |  |
| `spec.temporal.external.tls.pem.clientCertificate` | `string` |  |  |  |
| `spec.temporal.external.tls.pem.clientPrivateKey` | `string` (sensitive) |  |  |  |
| `spec.temporal.external.tls.existingSecret` | `KubernetesStigmerTemporalTlsSecret` |  |  |  |
| `spec.temporal.external.tls.existingSecret.name` | `string \| valueFrom` | yes |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.temporal.external.tls.existingSecret.caKey` | `string` |  |  |  |
| `spec.temporal.external.tls.existingSecret.certificateKey` | `string` |  |  |  |
| `spec.temporal.external.tls.existingSecret.privateKeyKey` | `string` |  |  |  |
| `spec.temporal.external.apiKey` | `KubernetesStigmerSecretValue` |  |  |  |
| `spec.temporal.external.apiKey.value` | `string` (sensitive) |  |  |  |
| `spec.temporal.external.apiKey.existingSecret` | `KubernetesStigmerSecretKeyRef` |  |  |  |
| `spec.temporal.external.apiKey.existingSecret.name` | `string \| valueFrom` | yes |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.temporal.external.apiKey.existingSecret.key` | `string` | yes |  |  |
| `spec.objectStore` | `KubernetesStigmerObjectStore` |  |  |  |
| `spec.objectStore.s3` | `KubernetesStigmerS3` | yes |  |  |
| `spec.objectStore.s3.bucket` | `string \| valueFrom` | yes |  | AwsS3Bucket (`status.outputs.bucket_id`) |
| `spec.objectStore.s3.region` | `string` |  | `us-east-1` |  |
| `spec.objectStore.s3.endpoint` | `string \| valueFrom` |  |  | KubernetesSeaweedFs (`status.outputs.s3_endpoint`) |
| `spec.objectStore.s3.credentials` | `KubernetesStigmerS3Credentials` | yes |  |  |
| `spec.objectStore.s3.credentials.accessKey` | `KubernetesStigmerS3AccessKey` |  |  |  |
| `spec.objectStore.s3.credentials.accessKey.accessKeyId` | `string` | yes |  |  |
| `spec.objectStore.s3.credentials.accessKey.secretAccessKey` | `string` (sensitive) | yes |  |  |
| `spec.objectStore.s3.credentials.existingSecretName` | `string \| valueFrom` |  |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.objectStore.s3.runnerEndpoint` | `string` |  |  |  |
| `spec.ingress` | `KubernetesStigmerIngress` |  |  |  |
| `spec.ingress.hostname` | `string` | yes |  |  |
| `spec.ingress.tls` | `KubernetesStigmerIngressTls` |  |  |  |
| `spec.ingress.tls.secretName` | `string` |  |  |  |
| `spec.ingress.tls.issuer` | `KubernetesStigmerIssuerRef` |  |  |  |
| `spec.ingress.tls.issuer.name` | `string \| valueFrom` | yes |  | KubernetesClusterIssuer (`status.outputs.cluster_issuer_name`) |
| `spec.ingress.tls.issuer.kind` | `string` |  |  |  |
| `spec.ingress.gateway` | `KubernetesStigmerGatewayRef` |  |  |  |
| `spec.ingress.gateway.name` | `string \| valueFrom` | yes |  | KubernetesGateway (`status.outputs.gateway_name`) |
| `spec.ingress.gateway.namespace` | `string \| valueFrom` |  |  | KubernetesGateway (`status.outputs.namespace`) |
| `spec.ingress.gateway.sectionName` | `string` |  |  |  |
| `spec.ingress.className` | `string` |  |  |  |
| `spec.ingress.annotations` | `map<string, string>` |  |  |  |
| `spec.ingress.serviceAnnotations` | `map<string, string>` |  |  |  |
| `spec.identity` | `KubernetesStigmerIdentity` |  |  |  |
| `spec.identity.bundled` | `KubernetesStigmerBundledIdentity` |  |  |  |
| `spec.identity.bundled.signUpEmailDomains` | `[]string` |  |  |  |
| `spec.identity.bundled.smtp` | `KubernetesStigmerSmtp` |  |  |  |
| `spec.identity.bundled.smtp.host` | `string` | yes |  |  |
| `spec.identity.bundled.smtp.port` | `int32` |  | `587` |  |
| `spec.identity.bundled.smtp.from` | `string` | yes |  |  |
| `spec.identity.bundled.smtp.starttls` | `bool` |  |  |  |
| `spec.identity.bundled.smtp.credentials` | `KubernetesStigmerSmtpCredentials` |  |  |  |
| `spec.identity.bundled.smtp.credentials.login` | `KubernetesStigmerSmtpLogin` |  |  |  |
| `spec.identity.bundled.smtp.credentials.login.username` | `string` | yes |  |  |
| `spec.identity.bundled.smtp.credentials.login.password` | `string` (sensitive) | yes |  |  |
| `spec.identity.bundled.smtp.credentials.existingSecretName` | `string \| valueFrom` |  |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.identity.oidc` | `KubernetesStigmerOidc` |  |  |  |
| `spec.identity.oidc.issuer` | `string` | yes |  |  |
| `spec.identity.oidc.audience` | `string` | yes |  |  |
| `spec.identity.oidc.consoleClientId` | `string` |  |  |  |
| `spec.identity.oidc.nativeClientId` | `string` |  |  |  |
| `spec.identity.oidc.discoveryUrl` | `string` |  |  |  |
| `spec.enterprise` | `KubernetesStigmerEnterprise` |  |  |  |
| `spec.enterprise.license` | `KubernetesStigmerSecretValue` | yes |  |  |
| `spec.enterprise.license.value` | `string` (sensitive) |  |  |  |
| `spec.enterprise.license.existingSecret` | `KubernetesStigmerSecretKeyRef` |  |  |  |
| `spec.enterprise.license.existingSecret.name` | `string \| valueFrom` | yes |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.enterprise.license.existingSecret.key` | `string` | yes |  |  |
| `spec.enterprise.organizationCreators` | `[]string` |  |  |  |
| `spec.enterprise.accessHistoryRetentionDays` | `int32` |  |  |  |
| `spec.enterprise.outboundEgress` | `string` |  |  |  |
| `spec.sandboxes` | `KubernetesStigmerSandboxes` |  |  |  |
| `spec.sandboxes.driver` | `string` |  |  |  |
| `spec.sandboxes.idleSeconds` | `int32` |  |  |  |
| `spec.sandboxes.agentSandbox` | `KubernetesStigmerAgentSandbox` |  |  |  |
| `spec.sandboxes.agentSandbox.egressAllowCidrs` | `[]string` |  |  |  |
| `spec.sandboxes.substrate` | `KubernetesStigmerSubstrate` |  |  |  |
| `spec.sandboxes.substrate.workers` | `int32` |  | `1` |  |
| `spec.sandboxes.substrate.workerResources` | `KubernetesCoreV1ApiResourceRequirements` |  |  |  |
| `spec.sandboxes.substrate.workerResources.limits` | `map<string, string>` |  |  |  |
| `spec.sandboxes.substrate.workerResources.requests` | `map<string, string>` |  |  |  |
| `spec.sandboxes.substrate.workerResources.claims` | `[]KubernetesCoreV1ApiResourceClaim` |  |  |  |
| `spec.sandboxes.substrate.workerResources.claims[].name` | `string` | yes |  |  |
| `spec.sandboxes.substrate.workerResources.claims[].request` | `string` |  |  |  |
| `spec.sandboxes.substrate.workerNodeSelector` | `map<string, string>` |  |  |  |
| `spec.sandboxes.substrate.workerTolerations` | `[]KubernetesCoreV1ApiToleration` |  |  |  |
| `spec.sandboxes.substrate.workerTolerations[].key` | `string` |  |  |  |
| `spec.sandboxes.substrate.workerTolerations[].operator` | `string` |  |  |  |
| `spec.sandboxes.substrate.workerTolerations[].value` | `string` |  |  |  |
| `spec.sandboxes.substrate.workerTolerations[].effect` | `string` |  |  |  |
| `spec.sandboxes.substrate.workerTolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.sandboxes.substrate.sandboxCpuMillicores` | `int32` |  |  |  |
| `spec.sandboxes.substrate.sandboxMemoryMib` | `int32` |  |  |  |
| `spec.sandboxes.substrate.egressHttp` | `[]string` |  |  |  |
| `spec.prerequisites` | `KubernetesStigmerPrerequisites` |  |  |  |
| `spec.prerequisites.postgresOperator` | `string` |  |  |  |
| `spec.prerequisites.postgresBackupPlugin` | `string` |  |  |  |
| `spec.prerequisites.agentSandbox` | `string` |  |  |  |
| `spec.prerequisites.substrate` | `string` |  |  |  |
| `spec.upgradeWithoutBackupTo` | `string` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

The namespace the Stigmer lives in: its server, runner, database,
Temporal and Secrets, and the keys Secret the operator keeps. It must
already exist; this kind never creates or deletes it, because deleting it
would delete the encryption keys and make every backup unreadable. Give it
a KubernetesNamespace of its own, or name one you keep.

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.version

`string` · required

The Stigmer version: the chart version, which is also the operator's and
the server's version (chart 0.3.2 runs Stigmer v0.3.2). Exact semver, at
0.3.1 or later: 0.3.0 was never published as a chart. There is no default
on purpose: a default that moved with catalog updates would upgrade a
running install on an ordinary re-apply. Moving it below the version the
server runs is refused before anything is applied, because Stigmer's
server refuses to start on a database a newer version migrated; go back by
restoring a backup instead. Released versions:
https://github.com/stigmer/stigmer/releases

- rule: version must be an exact Stigmer release such as "0.3.2": the chart, the operator and the server share it, and a range or a leading "v" names no chart
- rule: version must be 0.3.1 or later: Stigmer's operator chart starts at 0.3.1
- rule: {"required":true}

### spec.operator

`KubernetesStigmerOperator` · required

The operator: where it runs and how its chart is tuned.

- rule: {"required":true}

### spec.operator.namespace

`string | valueFrom` · required

The namespace the operator runs in ("stigmer-system" by convention).
Only the controller lives here; the Stigmer lives in `namespace`. To move
the operator to another namespace, destroy and re-create: an in-place
move is refused, because the release's cluster-wide objects and the kept
definition belong to the old one.

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.operator.createNamespace

`bool`

When true, the operator's namespace is created before the chart is
installed and deleted with this resource. When false, it must already
exist (a KubernetesNamespace of its own, or one you keep for its quotas
and policies).

### spec.operator.imageRepository

`string`

The operator image repository, for a registry mirror. Empty = the
chart's own (ghcr.io/stigmer/stigmer-operator). The tag is always
`version`: the chart has no setting this kind exposes for it, because a
different tag would run a different Stigmer from the one declared.

### spec.operator.imagePullPolicy

`string`

When the kubelet pulls the operator image. Empty = the chart's own
(IfNotPresent).

- rule: {"string":{"in":["","Always","IfNotPresent","Never"]}}

### spec.operator.imagePullSecrets

`[]string`

Names of image-pull Secrets in the operator's namespace, for a private
mirror.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.operator.replicas

`int32` · optional (explicit presence)

Operator replicas. Empty = the chart's own (1). More replicas are
leader-elected warm standbys that shorten the operator's own failover.

- rule: {"int32":{"gte":1}}

### spec.operator.resources

`KubernetesStigmerOperatorResources`

The operator container's requests and limits: cpu, memory and
ephemeral-storage. Empty = the chart's own (requests 50m and 128Mi,
memory limit 512Mi).

### spec.operator.resources.requests

`map<string, string>`

Requests, keyed by cpu, memory or ephemeral-storage.

- rule: {"map":{"keys":{"string":{"in":["cpu","memory","ephemeral-storage"]}},"values":{"string":{"pattern":"^(\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))))?$"}}}}

### spec.operator.resources.limits

`map<string, string>`

Limits, keyed by cpu, memory or ephemeral-storage.

- rule: {"map":{"keys":{"string":{"in":["cpu","memory","ephemeral-storage"]}},"values":{"string":{"pattern":"^(\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))))?$"}}}}

### spec.operator.logLevel

`string`

The operator's log level. Empty = the chart's own (info).

- rule: {"string":{"in":["","info","debug"]}}

### spec.operator.nodeSelector

`map<string, string>`

Node selector for the operator pod.

### spec.operator.tolerations

`[]KubernetesCoreV1ApiToleration`

Tolerations for the operator pod.

- rule: A toleration with operator Exists matches any value, so it takes no value

### spec.operator.tolerations[].key

`string`

`key` is the taint key that the toleration applies to. Empty means match
all taint keys. If the key is empty, operator must be Exists; this
combination means to match all values and all keys.

### spec.operator.tolerations[].operator

`string` · optional (explicit presence)

`operator` represents a key's relationship to the value. Valid operators
are Exists, Equal, Lt, and Gt. Defaults to Equal. Exists is equivalent to
wildcard for value, so that a pod can tolerate all taints of a particular
category. Lt and Gt perform numeric comparisons (requires feature gate
TaintTolerationComparisonOperators).

- rule: {"string":{"in":["Exists","Equal","Lt","Gt"]}}

### spec.operator.tolerations[].value

`string`

`value` is the taint value the toleration matches to. If the operator is
Exists, the value should be empty, otherwise just a regular string.

### spec.operator.tolerations[].effect

`string` · optional (explicit presence)

`effect` indicates the taint effect to match. Empty means match all taint
effects. When specified, allowed values are NoSchedule, PreferNoSchedule
and NoExecute.

- rule: {"string":{"in":["NoSchedule","PreferNoSchedule","NoExecute"]}}

### spec.operator.tolerations[].tolerationSeconds

`int64` · optional (explicit presence)

`tolerationSeconds` represents the period of time the toleration (which
must be of effect NoExecute, otherwise this field is ignored) tolerates the
taint. By default, it is not set, which means tolerate the taint forever
(do not evict). Zero and negative values will be treated as 0 (evict
immediately) by the system.

### spec.image

`KubernetesStigmerImage`

Where Stigmer's own images (server, runner, sandboxes) are pulled from.
The tag is always `version`.

### spec.image.registry

`string`

The registry that holds every Stigmer image under its usual name.
Empty = Stigmer's own (ghcr.io/stigmer).

### spec.image.pullPolicy

`string`

When the kubelet pulls Stigmer's images. Empty = IfNotPresent.

- rule: {"string":{"in":["","Always","IfNotPresent","Never"]}}

### spec.image.pullSecrets

`[]string`

Names of image-pull Secrets in the Stigmer's namespace.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.owner

`KubernetesStigmerOwner`

The person who runs this install. With sign-in on, they own the
organization (the bundled sign-in creates their account first); with it
off, every action is attributed to them. Required whenever the bundled
sign-in is on.

### spec.owner.email

`string` · required

The owner's sign-in address.

- rule: owner.email must be an email address
- rule: {"required":true}

### spec.owner.name

`string`

How the owner is shown.

### spec.server

`KubernetesStigmerServer`

The Stigmer server.

### spec.server.resources

`KubernetesCoreV1ApiResourceRequirements`

The server container's requests and limits. Empty = the operator's
defaults. A request above its limit, or memory below what the server
needs, is refused by the operator in the install's status.

### spec.server.resources.limits

`map<string, string>`

`limits` describes the maximum amount of compute resources allowed. More
info:
https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/

- rule: {"map":{"values":{"string":{"pattern":"^(\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))))?$"}}}}

### spec.server.resources.requests

`map<string, string>`

`requests` describes the minimum amount of compute resources required. If
Requests is omitted for a container, it defaults to Limits if that is
explicitly specified, otherwise to an implementation-defined value.
Requests cannot exceed Limits. More info:
https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/

- rule: {"map":{"values":{"string":{"pattern":"^(\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))))?$"}}}}

### spec.server.resources.claims

`[]KubernetesCoreV1ApiResourceClaim`

`claims` lists the names of resources, defined in spec.resourceClaims, that
are used by this container.

This field depends on the DynamicResourceAllocation feature gate.

This field is immutable. It can only be set for containers.

### spec.server.resources.claims[].name

`string` · required

`name` must match the name of one entry in pod.spec.resourceClaims of the
Pod where this field is used. It makes that resource available inside a
container.

- rule: {"string":{"minLen":"1"}}

### spec.server.resources.claims[].request

`string`

`request` is the name chosen for a request in the referenced claim. If
empty, everything from the claim is made available, otherwise only the
result of this request.

### spec.server.storage

`KubernetesStigmerVolume`

The server's own volume: the skill store and everything under its home
except artifacts. Empty = 5Gi on the cluster's default storage class.

### spec.server.storage.size

`string`

The claim's requested capacity, as a Kubernetes quantity (10Gi). Empty =
the operator's default for this volume.

- rule: size must be a Kubernetes quantity such as 10Gi

### spec.server.storage.storageClassName

`string`

The storage class the claim binds with. Empty = the cluster's default
class (some clusters, GKE 1.37 among them, have none: name one).

### spec.server.artifacts

`KubernetesStigmerVolume`

The volume the server writes artifacts to and the shared runner reads
them from, when no object_store is declared. Empty = 20Gi.

### spec.server.artifacts.size

`string`

The claim's requested capacity, as a Kubernetes quantity (10Gi). Empty =
the operator's default for this volume.

- rule: size must be a Kubernetes quantity such as 10Gi

### spec.server.artifacts.storageClassName

`string`

The storage class the claim binds with. Empty = the cluster's default
class (some clusters, GKE 1.37 among them, have none: name one).

### spec.server.help

`KubernetesStigmerHelp`

Where your people ask for help, shown in the console.

### spec.server.help.url

`string` · required

Where the link goes.

- rule: {"required":true}

### spec.server.help.label

`string`

The link's text.

### spec.server.internalAccounts

`[]string`

Your own team's sign-in addresses, each exact (sam@example.com) or with
one * (*@example.com), left out of the onboarding count.

### spec.server.allowUnauthenticatedExposure

`bool`

Says, on purpose, that an install with a front door signs nobody in:
everyone who reaches the hostname acts as the owner. Without it, a front
door brings the bundled sign-in.

### spec.server.egress

`KubernetesStigmerServerEgress`

Fences the server's own outbound calls. Off by default: a server reaches
an issuer or an MCP server on your private network, which a fence would
cut.

### spec.server.egress.fenced

`bool`

Lets the server reach only DNS, the Kubernetes API, what the install
runs (its database, Temporal, the bundled sign-in, Agent Substrate), the
bucket, the public internet and allow_cidrs.

### spec.server.egress.allowCidrs

`[]string`

Private ranges the fenced server may also reach, such as an external
database, Temporal, issuer or MCP server (10.20.0.0/16).

- rule: {"repeated":{"maxItems":"64","items":{"string":{"maxLen":"43","ipPrefix":true}}}}

### spec.server.extraEnv

`ContainerEnv`

Extra settings for the server container. `variables` are written into
the server's spec; `secrets` are kept in a Secret this resource writes
(or read from your own) and reach the container by reference. A name the
operator sets itself is refused, so nothing here can silently undo the
declaration; `env_from` is not accepted, because Stigmer takes single
settings only.

- rule: server.extra_env.env_from is not accepted: Stigmer takes single settings, so list each one under variables or secrets
- rule: a setting may be declared in server.extra_env.variables or server.extra_env.secrets, not both
- rule: server.extra_env names a setting Stigmer's operator sets from the declaration, and a second value would silently replace it: set it through its own field, or remove the entry. The operator sets ANTHROPIC_*, OPENAI_API_KEY, AWS_BEARER_TOKEN_BEDROCK, DATABASE_URL, PG*, S3_*, TEMPORAL_*, GRPC_PORT, ARTIFACT_*, SANDBOX_PROVISIONER_TYPE, SKILL_* and most STIGMER_* settings

### spec.server.extraEnv.variables

`[]EnvVar` · no secrets

Individual environment variables (non-sensitive). Their values are
written into the pod spec, where anyone who can read the workload sees
them; a secret belongs in `secrets`, which the module keeps in a
Kubernetes Secret the container reads by reference.

- secrets: this value is stored where anyone who can view the resource reads it, so a secret reference (`$secret/...`) here is refused -- put a secret in `secrets`, which keeps it in a secret store the workload reads by reference
### spec.server.extraEnv.variables[].name

`string` · required

The environment variable name.
Must be a valid C_IDENTIFIER: starts with a letter or underscore,
followed by letters, digits, or underscores.

- rule: Must be a valid C_IDENTIFIER (start with letter/underscore, contain only letters, digits, underscores)
- rule: {"required":true}

### spec.server.extraEnv.variables[].value

`string`

Direct literal value.

### spec.server.extraEnv.variables[].valueFrom

`ValueFromRef`

Reference to another Planton resource's field.
The orchestrator resolves this and populates the value before invoking IaC modules.

### spec.server.extraEnv.variables[].valueFrom.kind

`enum`

Allowed values (use exactly as shown):

- `unspecified` -- 0: Default/unspecified
- `TestCatalogKindGeneric` -- 1–49: Test/dev/custom
- `TestCatalogKindKubernetes`
- `AwsAlb` -- 1000–1999: AWS resources AwsSubnet is a prerequisite because an ALB requires at least two subnets in different availability zones -- the spec's subnet references must resolve before the load balancer can be created.
- `AwsCertManagerCert`
- `AwsCloudFront`
- `AwsDynamodb`
- `AwsEcrRepo`
- `AwsEcsCluster`
- `AwsEcsService` -- AwsEcsCluster, AwsEcsTaskDefinition, and AwsSubnet are prerequisites because a service schedules a referenced task-definition revision into a referenced live cluster and places task network interfaces into referenced subnets -- all three references must resolve first.
- `AwsEksCluster` -- AwsSubnet and AwsIamRole are prerequisites because the control plane attaches its network interfaces into referenced subnets and assumes a referenced cluster role that must already carry AmazonEKSClusterPolicy.
- `AwsIamRole`
- `AwsLambda`
- `AwsRdsCluster`
- `AwsRdsInstance`
- `AwsRoute53Zone`
- `AwsS3Bucket`
- `AwsLbTargetGroup` -- AwsVpc is a prerequisite because a target group's health checks and target registrations live inside one VPC -- the spec's vpc_id reference must resolve before the group can be created.
- `AwsSecurityGroup` -- AwsVpc is a prerequisite because every security group is created in a VPC; the E2E install profile resolves vpc_id against the VPC prerequisite.
- `AwsVpc`
- `AwsEksNodeGroup` -- AwsEksCluster is a prerequisite because nodes register with a live control plane; AwsIamRole and AwsSubnet back the node role and worker subnet references.
- `AwsIamUser`
- `AwsKmsKey`
- `AwsEc2Instance`
- `AwsClientVpn` -- Every Client VPN endpoint requires an ACM server certificate at create time; the imported self-signed fixture satisfies it. Subnets/VPC are optional composition (a zero-association endpoint is valid) -- composed scenarios declare them via the e2e-prerequisites annotation.
- `AwsDocumentDb`
- `AwsRoute53DnsRecord` -- AwsRoute53Zone is a prerequisite because every record lives inside a hosted zone -- the spec's zone_id reference must resolve before the record can be created.
- `AwsS3ObjectSet` -- AwsS3Bucket is a prerequisite because the object set's bucket reference is required -- objects cannot exist without the bucket that holds them.
- `AwsSqsQueue`
- `AwsSnsTopic`
- `AwsEventBridgeBus`
- `AwsEventBridgeRule`
- `AwsIamOidcProvider`
- `AwsIamPolicy`
- `AwsIamInstanceProfile` -- AwsIamRole is a prerequisite because an instance profile is a wrapper that must contain a role to be useful -- the profile's spec requires a role reference, so the role must be deployed first.
- `AwsLbListener` -- AwsAlb and AwsLbTargetGroup are prerequisites because a listener is an attachment point on a load balancer and its default action almost always forwards to a target group -- both references must resolve before the listener can be created.
- `AwsLbListenerRule` -- AwsLbListener is a prerequisite because a rule only exists as an attachment on a listener -- the listener_arn reference must resolve before the rule can be created.
- `AwsLaunchTemplate`
- `AwsAutoScalingGroup` -- AwsSubnet and AwsLaunchTemplate are prerequisites because a group cannot exist without subnets to place capacity in and a launch template to launch from -- the spec's subnets and launch_template references must resolve before the group can be created.
- `AwsEksAddon` -- AwsEksCluster is a prerequisite because an add-on installs onto a live control plane -- the spec's cluster_name reference must resolve before the add-on can be created.
- `AwsEksFargateProfile` -- AwsEksCluster, AwsIamRole, and AwsSubnet are prerequisites because a Fargate profile attaches to a live control plane, runs pods as a referenced pod-execution role, and launches them into referenced private subnets -- all three references must resolve first.
- `AwsEksAccessEntry` -- AwsEksCluster and AwsIamRole are prerequisites because an access entry grants a referenced IAM principal access to a live control plane -- both references must resolve before the entry can be created.
- `AwsEcsTaskDefinition` -- AwsIamRole is a prerequisite because the kind's default posture -- Fargate with the awslogs logging default -- is rejected by AWS at registration time without an execution role the agent can assume.
- `AwsHttpApiGateway`
- `AwsStepFunction` -- AwsIamRole is a prerequisite because a state machine cannot be created without an execution role it can assume -- the spec's role_arn reference must resolve before the CreateStateMachine call.
- `AwsHttpApiVpcLink` -- AwsSubnet is a prerequisite because a VPC link is a set of managed ENIs provisioned into referenced subnets -- the subnet references must resolve before the link can be created. Security groups are optional on the link, so they compose per-scenario rather than as a registry prerequisite.
- `AwsHttpApiDomain` -- AwsCertManagerCert is a prerequisite because a custom domain cannot be created without a TLS certificate in the same region covering the domain -- the spec's certificate_arn reference must resolve first.
- `AwsVpcEndpoint` -- AwsVpcEndpoint's composed E2E scenarios reference the AwsVpc prerequisite's outputs (vpc_id + default_route_table_id for gateway endpoints) and the AwsSubnet pair's subnet_id outputs (interface endpoints), so both are genuine deploy-order prerequisites.
- `AwsElasticacheUser`
- `AwsElasticacheUserGroup` -- AwsElasticacheUser is a genuine prerequisite: AWS refuses to create a user group that does not contain a user named "default", so a group's composed E2E scenario must resolve a deployed user's outputs.
- `AwsRedshiftServerlessNamespace`
- `AwsRedshiftServerlessWorkgroup` -- The namespace is a genuine prerequisite: a workgroup attaches to exactly one namespace by name at create time, so its composed E2E scenario must resolve a deployed namespace's outputs. AwsSubnet is a prerequisite because Redshift Serverless requires the workgroup's subnets to span three availability zones.
- `AwsRedisElasticache` -- AwsSubnet is a prerequisite because the module builds an ElastiCache subnet group from referenced subnets -- the spec's subnet references must resolve before the replication group can deploy.
- `AwsOpenSearchDomain`
- `AwsMemcachedElasticache`
- `AwsServerlessElasticache`
- `AwsNlb` -- AwsSubnet is a prerequisite because an NLB requires at least one subnet mapping -- the spec's subnet references must resolve before the load balancer can be created.
- `AwsElasticIp`
- `AwsTransitGateway`
- `AwsGlobalAccelerator`
- `AwsSubnet`
- `AwsInternetGateway`
- `AwsNatGateway` -- AwsInternetGateway is a prerequisite because a public NAT gateway can only become available once the VPC it sits in has an internet gateway attached (AWS rejects the create otherwise) -- so the gateway must be deployed first. AwsVpc is a prerequisite because a REGIONAL NAT gateway (availability_mode = regional) references the VPC directly instead of a subnet.
- `AwsEgressOnlyInternetGateway`
- `AwsElasticFileSystem` -- AwsSubnet and AwsSecurityGroup are prerequisites because mount targets (required, min 1) place the file system's NFS endpoints into subnets and attach security groups -- both references must resolve before the CreateMountTarget calls.
- `AwsEfsAccessPoint` -- AwsElasticFileSystem is a prerequisite because an access point is created INTO a file system -- the spec's required file_system_id reference must resolve before the CreateAccessPoint call.
- `AwsFsxLustreFileSystem`
- `AwsFsxOpenzfsFileSystem`
- `AwsFsxWindowsFileSystem` -- Every Windows file system must join an Active Directory domain; the directory itself is external infrastructure (AWS Managed Microsoft AD or a self-managed domain), so only the network dependency is a declarable prerequisite.
- `AwsFsxOntapFileSystem`
- `AwsFsxOntapStorageVirtualMachine`
- `AwsFsxOntapVolume`
- `AwsFsxDataRepositoryAssociation`
- `AwsCognitoUserPool`
- `AwsCognitoIdentityProvider` -- AwsCognitoUserPool is a prerequisite because an identity provider is created INTO a pool -- the spec's required user_pool_id reference must resolve before the CreateIdentityProvider call.
- `AwsCognitoUserPoolClient` -- AwsCognitoUserPool is a prerequisite because an app client is created INTO a pool -- the spec's required user_pool_id reference must resolve before the CreateUserPoolClient call.
- `AwsCognitoResourceServer` -- AwsCognitoUserPool is a prerequisite because a resource server is created INTO a pool -- the spec's required user_pool_id reference must resolve before the CreateResourceServer call.
- `AwsWafWebAcl`
- `AwsWafIpSet`
- `AwsWafRegexPatternSet`
- `AwsCloudwatchLogGroup`
- `AwsCloudwatchAlarm`
- `AwsCloudwatchCompositeAlarm`
- `AwsKinesisStream`
- `AwsKinesisFirehose` -- Every Firehose destination requires an S3 configuration (the primary target for extended_s3; the failed/all-document backup for the rest) and an IAM role Firehose assumes to write to it, so both are hard deploy prerequisites.
- `AwsKinesisStreamConsumer` -- A consumer registers against exactly one stream and cannot exist without it.
- `AwsAthenaWorkgroup`
- `AwsGlueCatalogDatabase`
- `AwsRedshiftCluster`
- `AwsSagemakerDomain` -- AI/ML A domain cannot exist without VPC subnets and a SageMaker execution role (default_user_settings.execution_role_arn is required), so both are hard deploy prerequisites.
- `AwsAppRunnerService` -- A service can run entirely on companion defaults, so the App Runner family's kinds are dependency-free leaves except the VPC connector (which cannot exist without subnets and security groups). A service's companion references (auto scaling / VPC connector / observability / WAF) are optional composition -- scenarios declare them via the e2e-prerequisites annotation.
- `AwsAppRunnerAutoScalingConfiguration`
- `AwsAppRunnerVpcConnector`
- `AwsAppRunnerObservabilityConfiguration`
- `AwsTransitGatewayVpcAttachment` -- AwsTransitGateway is a prerequisite because an attachment cannot exist without the gateway it attaches to; AwsSubnet because the attachment provisions an ENI into at least one subnet (the VPC arrives transitively through the subnet's own prerequisites).
- `AwsTransitGatewayRouteTable` -- Only the gateway is a hard prerequisite: a route table can exist empty. Associations, propagations, and routes referencing attachments are optional composition -- scenarios declare them via the e2e-prerequisites annotation.
- `AwsBatchComputeEnvironment` -- A MANAGED compute environment always launches into VPC subnets, so the subnet is a hard deploy prerequisite (security groups are required only for the Fargate types -- scenario-declared, not a registry edge).
- `AwsBatchJobQueue` -- A job queue cannot exist without at least one VALID compute environment to map onto.
- `AwsBatchSchedulingPolicy`
- `AwsBatchJobDefinition`
- `AwsCodeBuildProject` -- CI/CD
- `AwsCodePipeline`
- `AwsMwaaEnvironment` -- Workflow / Orchestration AwsSubnet and AwsSecurityGroup are prerequisites because the environment's network interfaces are placed in referenced private subnets and AWS requires at least one attached security group at creation.
- `AwsNeptuneCluster` -- Graph Database
- `AwsMemorydbCluster` -- A cluster always launches into a subnet group; the subnets are the hard deploy prerequisite. The ACL it attaches is optional composition (the built-in "open-access" ACL needs no resource) -- scenarios declare the ACL/user chain via the e2e-prerequisites annotation.
- `AwsMemorydbUser`
- `AwsMemorydbAcl` -- An empty ACL is valid (MemoryDB has no mandatory "default" member), so the user is optional composition -- the composed scenario declares it via the e2e-prerequisites annotation, never a registry edge.
- `AwsMskCluster` -- Streaming AwsSubnet and AwsSecurityGroup are prerequisites because brokers are placed in referenced subnets and AWS requires at least one attached security group at creation.
- `AwsMskServerlessCluster` -- AwsSubnet is a prerequisite because the serverless cluster's network interfaces are placed in referenced subnets (security groups are optional -- AWS attaches the VPC default group when none are referenced).
- `AwsLambdaEventSourceMapping` -- AwsLambda is a prerequisite because a mapping cannot exist without the function it invokes (a required reference). Event sources (SQS, Kinesis, DynamoDB, MSK) are optional composition -- scenarios declare them via the e2e-prerequisites annotation rather than taxing every consumer's chain.
- `AwsSnsSubscription` -- AwsSnsTopic is a prerequisite because a subscription cannot exist without the topic it subscribes to (a required reference). Endpoints (SQS queues, Lambda functions, Firehose streams) are optional composition -- scenarios declare them via the e2e-prerequisites annotation rather than taxing every consumer's chain.
- `AwsPlantonRunner` -- AwsSubnet is a prerequisite because the runner appliance places its network interfaces into referenced subnets -- the placement reference must resolve before the appliance can deploy.
- `AwsRoute53HealthCheck`
- `AwsSesConfigurationSet` -- Both SES kinds are dependency-free leaves: an identity's configuration set is optional composition (scenarios declare it via the e2e-prerequisites annotation), and a configuration set's event destinations reference other kinds only optionally.
- `AwsSesEmailIdentity`
- `AwsSecretsManagerSecret` -- A dependency-free leaf: the KMS key, rotation Lambda, and external rotation role references are all optional composition -- scenarios declare them via the e2e-prerequisites annotation, never registry edges.
- `AwsOpenSearchServerlessCollection` -- A dependency-free leaf: the collection-scoped encryption/network/ data-access/retention policies are module-rendered, and the KMS key and data-access principal references are optional composition (e2e-prerequisites annotation).
- `AwsBedrockGuardrail` -- A dependency-free leaf: the KMS key reference is optional composition (e2e-prerequisites annotation); published versions are folded satellites of the guardrail itself.
- `AwsBedrockCustomModel` -- AwsIamRole is a prerequisite because Bedrock assumes the job role to read training data and write outputs; the S3 locations and KMS key are optional composition (e2e-prerequisites annotation).
- `AwsBedrockInferenceProfile` -- A dependency-free leaf: the model source is a foundation model or an AWS system-defined cross-region profile, never a customer resource.
- `AwsBedrockProvisionedThroughput` -- A dependency-free leaf in the registry: capacity is typically bought for an AwsBedrockCustomModel (the default reference), but foundation model ARNs are equally legal, so the edge is optional composition.
- `AwsBedrockModelAccess` -- A dependency-free leaf: the agreement covers an AWS-listed foundation model, never a customer resource.
- `AwsBedrockInvocationLogging` -- Region settings singleton (one invocation-logging configuration per account+region; identity = the region). Delivery destinations are optional references (at least one of CloudWatch/S3, enforced by CEL), so prerequisites stay empty and E2E fixtures ride scenario annotations.
- `AwsBedrockAgent` -- AwsIamRole is a prerequisite because the Bedrock service assumes the agent resource role to invoke models, action-group Lambdas, and knowledge bases; the guardrail, KMS key, provisioned throughput, and collaborator/knowledge-base edges are optional composition (e2e-prerequisites annotation). Action groups, aliases, collaborators, and knowledge-base associations are folded satellites of the agent.
- `AwsBedrockKnowledgeBase` -- AwsIamRole is a prerequisite because the Bedrock service assumes the knowledge-base role to read data sources, call the embedding model, and read/write the vector store; the vector-store and data-source reference edges (OpenSearch, S3, Secrets Manager, ...) are optional composition (e2e-prerequisites annotation). Data sources are folded satellites of the knowledge base.
- `AwsBedrockFlow` -- AwsIamRole is a prerequisite because the Bedrock service assumes the flow execution role to invoke the models, agents, knowledge bases, and Lambdas its nodes reference; the node-level reference edges are optional composition (e2e-prerequisites annotation).
- `AwsBedrockPrompt` -- A dependency-free leaf: variants target AWS-listed foundation models by ID; targeting another agent's alias is optional composition (e2e-prerequisites annotation).
- `AwsBedrockAgentCoreRuntime` -- AwsIamRole is a prerequisite because the AgentCore service assumes the runtime role to pull the container image or read the S3 code bundle and to run the hosted agent; the code-bundle S3 bucket and VPC placement edges are optional composition (e2e-prerequisites annotation). Endpoints and the runtime's resource policy are folded satellites of the runtime.
- `AwsBedrockAgentCoreGateway` -- AwsIamRole is a prerequisite because the gateway assumes its role to reach targets (invoke Lambdas, sign SigV4 requests); the target and credential-provider reference edges (runtime, Lambda, Identity providers, policy engine) are optional composition (e2e-prerequisites annotation). Targets are folded satellites of the gateway - AWS deletes them before the gateway at destroy.
- `AwsBedrockAgentCoreMemory` -- A dependency-free leaf for built-in strategies: the execution role (custom strategies, Kinesis delivery), KMS key, and Kinesis stream edges are optional composition (e2e-prerequisites annotation). Strategies are folded satellites of the memory - AWS serializes their changes through the parent.
- `AwsBedrockAgentCoreIdentity` -- A dependency-free leaf: workload identities, credential providers, and the Cedar policy engine with its policies are all name-keyed arms of one identity-and-access bundle; the KMS key edge is optional composition (e2e-prerequisites annotation). The account/region token-vault CMK is deliberately NOT modeled here (settings singleton).
- `AwsBedrockAgentCoreTools` -- A dependency-free leaf in the SANDBOX/PUBLIC postures: the execution role (recordings, certificates), S3, Secrets Manager, and VPC edges are optional composition (e2e-prerequisites annotation). Browsers, profiles, and code interpreters are name-keyed arms of one tools bundle; AWS exposes no update - every field change recreates the tool.
- `AwsBedrockAgentCoreEvaluation` -- The AgentCore Evaluations bundle - evaluators (LLM-judge or Lambda scorers), harnesses (repeatable agent test benches), and online evaluation configs (continuous scoring of sampled production sessions). Deploys standalone - no arm requires an agent runtime to exist. No registry prerequisite: every arm is optional, so no dependency is required for the kind to function (scenarios compose IAM roles via annotations).
- `AwsBedrockAgentCoreTokenVault` -- Account/region settings singleton: sets the KMS key on the ONE default AgentCore token vault. The KMS reference is conditional on key_type (CEL-enforced), so prerequisites stay empty and E2E fixtures ride scenario annotations.
- `AwsSagemakerModel` -- The immutable serving definition (container image + artifacts + execution role) that endpoints deploy - one container or an inference pipeline.
- `AwsSagemakerEndpoint` -- A real-time inference endpoint WITH its folded endpoint configuration - the configuration is immutable upstream, so the modules roll name-suffixed configurations create-before-destroy and repoint the endpoint.
- `AwsSagemakerNotebookInstance` -- A managed Jupyter notebook EC2 instance with its folded lifecycle configuration (bootstrap scripts).
- `AwsSagemakerFeatureGroup` -- A Feature Store feature group - online and/or offline stores over a declared feature schema.
- `AwsSagemakerModelRegistry` -- A model registry package group with its folded resource policy - model package VERSIONS register into it imperatively (training pipelines), never declaratively.
- `AwsSagemakerPipeline` -- An ML workflow DAG (the SageMaker pipeline-definition JSON) that executions run against - free to create, billed per execution.
- `AwsSagemakerImage` -- A named registry entry exposing YOUR container images to Studio, with folded AWS-numbered versions (append-only by position).
- `AwsSagemakerMlflowServer` -- The classic hourly-billed managed MLflow tracking server (~25 min to provision; billed hourly from creation whether or not anyone is tracking). The serverless successor is AwsSagemakerMlflowApp.
- `AwsSagemakerMlflowApp` -- The serverless MLflow 3.x deployment (billed per use) - standalone, associating with SageMaker domains; NOT a tracking-server satellite.
- `AwsRestApiGateway` -- A full REST API (API Gateway v1): the resource/method tree with inline integrations (or an imported OpenAPI document), one stage with an explicit hash-triggered deployment, and the API-scoped satellites (authorizers, models, validators, gateway responses, policy, documentation, client certificate). Self-contained: a MOCK-integration API needs no other resource.
- `AwsRestApiDomain` -- A custom domain for REST APIs with base-path mappings and - for PRIVATE domains - VPC-endpoint access associations. AwsCertManagerCert is a prerequisite because the domain cannot be created without a TLS certificate covering it.
- `AwsRestApiUsagePlan` -- A usage plan metering REST API consumers - stage coverage, quota, throttles, and the API keys it admits. No registry prerequisite: a plan is valid with no stage coverage (scenarios compose the REST API via annotations).
- `AwsRestApiVpcLink` -- A REST API VPC link fronting an internal Network Load Balancer so REST integrations reach private services. AwsNlb is a prerequisite because AWS rejects link creation without the target balancer.
- `AwsApiGatewayAccountSettings` -- Region settings singleton (one API Gateway account object per account+region; identity = the region). The CloudWatch role is an optional reference (unset = the explicit no-logging posture), so prerequisites stay empty and E2E fixtures ride scenario annotations.
- `AwsCloudTrail` -- The account's API audit trail. AwsS3Bucket is a prerequisite because AWS rejects trail creation without a delivery bucket carrying the CloudTrail service-principal policy. 1240 opens the governance sub-band (1240-1249).
- `AwsConfigRecorder` -- Region singleton (one AWS Config recorder per region, named "default" by AWS; identity = the region). AwsIamRole is a prerequisite because the recorder cannot exist without its service role.
- `AwsConfigRule` -- One AWS Config compliance rule (managed, custom-lambda, or custom-policy; account- or organization-scoped) with optional auto-remediation. Managed rules need no prerequisites; the custom-lambda arm's function reference is conditional, so E2E fixtures ride scenario annotations.
- `AwsGuardDuty` -- Region singleton (AWS allows one GuardDuty detector per account+region; the detector has no name - identity = the region). Satellite references (S3 export bucket, KMS key) are conditional, so E2E fixtures ride scenario annotations.
- `AwsCloudTrailEventDataStore` -- CloudTrail Lake: a queryable, immutable event data store with its own retention and billing lifecycle - no trail required. The KMS key reference is conditional, so E2E fixtures ride scenario annotations.
- `AwsConfigAggregator` -- AWS Config cross-account/cross-region aggregation: the aggregator (collector side) and/or the reciprocal authorization grants (source-account side). Works with zero recorders; the org-source role reference is conditional, so E2E fixtures ride scenario annotations.
- `AwsConfigConformancePack` -- An AWS Config conformance pack (account- or organization-scoped): a template bundle that creates its own Config rules. Deployment requires an active Config recorder in the region (a service-side requirement, not a spec reference), so E2E fixtures ride scenario annotations.
- `AwsGuardDutyMalwareProtectionPlan` -- GuardDuty Malware Protection for S3: scans new objects in one bucket - a standalone plan protecting a bucket, not a detector satellite (its schema carries no detector reference). The execution role and the protected bucket are required references.
- `AwsBackupVault` -- An AWS Backup vault - the encrypted container recovery points live in, as either a standard vault (with its lock, access policy, and notification satellites) or a logically air-gapped vault (AWS's own VaultType discriminator). The KMS and SNS references are conditional, so E2E fixtures ride scenario annotations. 1250 opens the backup sub-band (1250-1259).
- `AwsBackupPlan` -- An AWS Backup plan: scheduled backup rules plus the resource selections that assign resources to them. AwsBackupVault is a prerequisite because every rule requires a target vault; the selections' IAM role is conditional and rides scenario annotations.
- `AwsBackupFramework` -- A Backup Audit Manager framework: compliance controls evaluating backup posture. No schema-required references (the Config recorder its evaluations need is a lane fixture, not a spec reference).
- `AwsBackupReportPlan` -- A Backup Audit Manager report plan: scheduled compliance/job reports delivered to S3. AwsS3Bucket is a prerequisite because the delivery channel's bucket is required.
- `AwsBackupRestoreTestingPlan` -- An AWS Backup restore testing plan with its folded selections: scheduled restore tests proving recovery points actually restore. Vault targeting accepts the "*" wildcard, so fixtures are conditional and ride scenario annotations.
- `AwsBackupSettings` -- Account/region settings singleton for AWS Backup: the account's global settings (cross-account backup) and the region's resource-type opt-in/management preferences. Both provider deletes are no-ops - settings persist after destroy.
- `AwsSsmParameter` -- An SSM Parameter Store entry (String/StringList/SecureString). The parameter's name is an explicit spec field - names are hierarchical paths ("/prod/db/url") metadata.name cannot carry. The KMS reference is conditional (SecureString only), so E2E fixtures ride scenario annotations. 1260 opens the SSM sub-band (1260-1269).
- `AwsSsmDocument` -- A customer-owned SSM document (Command/Automation/Session/...): reusable action definitions managed nodes and automations execute. State Manager associations are their own AwsSsmAssociation kind - an association binds ANY document (AWS-managed included), so it is not this document's satellite.
- `AwsSsmMaintenanceWindow` -- An SSM maintenance window with its folded target registrations and tasks (Run Command / Automation / Lambda / Step Functions) - the targets and tasks are true window satellites (ForceNew window_id edges). Identity is the AWS-generated "mw-..." id.
- `AwsSsmPatchBaseline` -- An SSM patch baseline with its folded patch-group registrations and the account/region default-baseline designation (delete RESTORES AWS's own predefined default for the OS). Identity is the AWS-generated "pb-..." id.
- `AwsSsmAssociation` -- A State Manager association: the binding of an SSM document to targets on a schedule. Split from the document kind because the document reference is a free string with no structural edge - associations routinely bind AWS-managed documents (AWS-RunShellScript, ...) with no user document anywhere, so no registry prerequisite either. Identity is the AWS-generated association UUID.
- `AwsOrganization` -- THE AWS Organization of the deploying account - creating it makes the caller the management account. Trusted service access, delegated administrators, the org's singleton resource policy, and centralized root-access management (IAM's organizations features - a management-account act requiring iam.amazonaws.com trusted access) fold in (none has a life of its own; the standalone service-access resource fights the org's own argument with a perpetual diff). Deleting this deletes the entire organization. 1270 opens the Organizations sub-band (1270-1279). A container kind: the organization is the outer wall of the tenancy tree AWS's own console draws -- organizational units are rooms inside it and member accounts stand inside those. A first-level unit names the organization through its parent reference and is placed inside it; a policy attached to the root names the organization too, but a policy is a guard applied from above, never a resident, and its attachment reference is containment-exempt on its spec.
- `AwsOrganizationalUnit` -- An organizational unit in the org's OU tree. The display name is an explicit spec field (OU names allow spaces metadata.name cannot carry); the parent reference (root or parent OU) is required and immutable, so the organization is a registry prerequisite. A container kind: a unit is a room of the tenancy tree -- nested units and member accounts name it through their parent reference and are placed inside it. A policy attached to a unit names it too and stands beside the room (containment-exempt on the policy's spec): one policy attaches to many units and cannot live in all of them.
- `AwsOrganizationAccount` -- A MEMBER account of the organization: creation, OU placement, and the account-level settings satellites (alternate/primary contacts, opt-in region enablement) fold onto the created account's ID. Destroy is never a clean delete (remove-from-org or ~90-day close) - taught on the spec. No registry prerequisite by the schema-required-only rule (the OU parent reference is optional).
- `AwsOrganizationPolicy` -- An Organizations policy (SCP and its twelve sibling types) with its folded attachments to roots, OUs, and member accounts. The policy type must be enabled on the organization first; AWS-managed policies are never adopted. No registry prerequisite by the schema-required-only rule (attachments are optional).
- `AwsBudget` -- A Budgets budget (COST/USAGE/RI/Savings Plans coverage and utilization) with its folded budget actions as name-keyed satellites - an action exists only on its budget and fires an IAM-policy application, an SCP attachment, or SSM instance stops when a threshold breaches. Budgets is account-global (served from us-east-1; the spec region is the provider endpoint). 1280 opens the cost-management sub-band (1280-1289).
- `AwsCostAnomalyMonitor` -- A Cost Explorer anomaly monitor (DIMENSIONAL over one dimension, or CUSTOM over a CE expression) with its folded alert subscriptions - a subscription's monitor list is the structural edge that makes it this monitor's satellite. Account-global; AWS identifies both by ARN.
- `AwsCostCategory` -- A Cost Explorer cost category: ordered rules (regular expression rules or inherited-value rules) over the recursive CE expression tree, plus split-charge rules. The account's cost-allocation-tag activation toggle is deliberately NOT folded here - it is a per-tag-key account feature with no edge to any category, so many category instances would fight over one account object.
- `AwsIamGroup` -- An IAM group with its folded declarative membership (the authoritative users list) and group policies - name-keyed inline documents plus managed-policy attachments. IAM is global; identity is the group name (renames update in place, the ARN recomputes). 1290 opens the IAM P1 sub-band (1290-1299).
- `AwsIamSamlProvider` -- An IAM SAML identity provider: the account's federation trust anchor, created from the IdP's metadata XML (a public document carrying certificates, not a secret). Identity is the provider ARN; the name is write-once.
- `AwsIamAccountSettings` -- Account settings singleton for IAM (a GLOBAL service - one object per ACCOUNT, not per region): the sign-in alias, the password policy, and the STS global-endpoint token version. Destroy contracts DIFFER per arm (each taught on its arm): the alias truly deletes, the password policy resets to AWS defaults, the STS preference is a no-op delete that persists.
- `AwsCloudwatchDashboard` -- A CloudWatch dashboard: one named dashboard whose widget layout is the dashboard-body JSON document (modeled as a typed Struct, the catalog's uniform policy-document idiom). Dashboards are untaggable at AWS. Identity is the dashboard name; every change is an in-place PutDashboard upsert. 1300 opens the CloudWatch observability P1 sub-band (1300-1309).
- `AwsCloudwatchSynthetics` -- CloudWatch Synthetics: a canary (a scheduled scripted probe running from an S3-staged code bundle under an execution role, writing run artifacts to S3) plus the grouping surface - owned groups and the canary's group associations (joins by group NAME, so shared groups are referenced, never fought over). A groups-only instance manages shared groups with no canary.
- `AwsCloudwatchLogDelivery` -- CloudWatch Logs delivery: the two ways logs leave CloudWatch. The vended-log arm pivots on a delivery SOURCE (one AWS resource whose service vends logs) with name-keyed deliveries fanning out to delivery destinations (S3 / CloudWatch Logs / Firehose / X-Ray), each created inline or referenced by ARN. The cross-account arm is the legacy Kinesis subscription destination with its access policy (whose delete is a no-op at AWS - the policy persists).
- `AwsCloudwatchLogAccountPolicy` -- A CloudWatch Logs account-level policy: one policy object per (name, type) pair per region - data protection, subscription filter, field index, transformer, or metric extraction - applied account-wide, optionally narrowed by selection criteria. Standalone account configuration, never a per-log-group satellite.
- `AwsCloudwatchLogAnomalyDetector` -- A CloudWatch Logs anomaly detector: one detector trains over a LIST of log groups (multi-parent scope - never a single group's satellite), surfacing anomalies on a chosen evaluation frequency with a bounded visibility window.
- `AwsCloudwatchLogResourcePolicy` -- A CloudWatch Logs resource policy: the account-scoped named policy (or resource-scoped policy on one log group ARN) that grants AWS services permission to write logs - Route53 query logging, EventBridge, and friends. Exactly one scope per instance.
- `AwsManagedPrometheus` -- An Amazon Managed Prometheus workspace with its folded satellites: workspace configuration (retention, label-set limits - a created-via-update singleton whose delete is a no-op at AWS), the alert manager definition (strictly one per workspace), name-keyed rule group namespaces, query logging, the workspace resource policy, and alias-keyed anomaly detectors. Scrapers are deliberately NOT folded here - a scraper can target CloudWatch with zero AMP workspaces, so it is its own kind.
- `AwsManagedPrometheusScraper` -- An Amazon Managed Prometheus scraper: the agentless collector. Source is an EKS cluster or a bare VPC placement (both replace-on-change); destination is an AMP workspace or a CloudWatch dataset. Carries its own scraper logging configuration satellite. Scrape configuration is optional on the EKS arm (AWS publishes a default, resolved at deploy) and required on the VPC arm.
- `AwsEventBridgePipe` -- An EventBridge Pipe: one point-to-point integration reading from one source (SQS, Kinesis, DynamoDB streams, MSK or self-managed Kafka, ActiveMQ/RabbitMQ), optionally filtering and enriching in-flight, and delivering to one target (ECS, Batch, Lambda, Step Functions, Kinesis, SQS, Redshift, SageMaker, CloudWatch Logs, EventBridge buses, HTTP via API destinations). The source is fixed for life (replace-on-change); the target swaps in place. 1310 opens the EventBridge extras P1 sub-band (1310-1319).
- `AwsEventBridgeScheduler` -- An EventBridge Scheduler schedule: cron/rate/one-time invocation of one target under an execution role, with flexible time windows, retry policy, and a dead-letter queue. The schedule GROUP is folded own-XOR-existing (a name-and-tags container - the provider's own update path is tags-only); unset means AWS's default group.
- `AwsEventBridgeApiDestination` -- An EventBridge API destination with its connection: the authenticated HTTP(S) endpoint rules, pipes, and schedules invoke. Two independently deployable arms - the CONNECTION (the shareable auth trust anchor: api-key, basic, or OAuth credentials that AWS stores in Secrets Manager) and the DESTINATION (endpoint + method + rate limit) whose connection is owned inline or referenced by ARN.
- `AwsVpcPeering` -- A VPC peering connection, as a request-XOR-accept mode union: the REQUEST arm creates the peering from its VPC toward a peer VPC (same-account auto-accept supported; cross-account/cross-region stays pending until accepted), the ACCEPT arm adopts and accepts a pending connection by ID from the accepter side. DNS-resolution options fold into both arms. 1320 opens the VPC networking P1 sub-band (1320-1329).
- `AwsNetworkAcl` -- A network ACL: the stateless subnet-level firewall - ordered ingress/egress rules (allow or deny, evaluated by rule number) and the subnet associations, all folded in-line as the single declarative owner (the standalone rule/association resources are the same payload and fight the in-line form).
- `AwsManagedPrefixList` -- A customer-managed prefix list: a named, versioned set of CIDR blocks that security-group rules, NACL rules, and route tables reference as one object. Entries fold in-line; max_entries is the capacity contract (referencing consumes that many rule slots regardless of how many entries exist).
- `AwsEbsVolume` -- A standalone EBS volume as a create-XOR-copy union (fresh in a zone, or cloned from another volume) with attachments managed in-line. 1330 opens the block & object storage sub-band (1330-1339).
- `AwsEbsSnapshot` -- An EBS snapshot as a three-way source union (snapshot a volume, copy a snapshot, or import a disk image) with archive tiering, fast snapshot restore, and cross-account share grants in-line.
- `AwsS3DirectoryBucket` -- An S3 directory bucket (S3 Express One Zone): single-AZ, single-digit-millisecond object storage. The modules derive the mandated "{name}--{zone_id}--x-s3" bucket name.
- `AwsS3TableBucket` -- An S3 table bucket (S3 Tables - managed Apache Iceberg storage) with its namespaces, tables, policies, and replication folded in-line as the single declarative owner.
- `AwsS3VectorBucket` -- An S3 vector bucket (AI embedding storage with similarity query) with its vector indexes folded in-line - the natural backend for Bedrock knowledge bases.
- `AwsDlmLifecyclePolicy` -- A Data Lifecycle Manager policy: account-level, tag-targeted snapshot/AMI automation (create, retain, archive, copy cross-region, share, deprecate) as a default-XOR-custom mode union. AwsIamRole is a prerequisite because DLM acts through a required execution role.
- `AwsRoute53ResolverEndpoint` -- A Route 53 Resolver endpoint (the hybrid-DNS bridge between a VPC and outside networks) with its forwarding rules and their VPC associations managed in-line. Subnets place the ENIs and security groups guard them - both schema-required. 1340 opens the DNS & service discovery sub-band (1340-1349).
- `AwsRoute53ResolverFirewall` -- A Route 53 Resolver DNS Firewall rule group with its domain lists, filtering rules, and VPC associations managed in-line - the DNS-layer block/allow policy for VPC egress queries. AwsVpc is a prerequisite because the association arm filters a referenced VPC.
- `AwsRoute53ResolverQueryLog` -- A Resolver query logging configuration (every DNS query VPCs make through the resolver, to CloudWatch Logs / S3 / Firehose) with its VPC associations managed in-line. AwsVpc is a prerequisite because the association arm logs a referenced VPC.
- `AwsCloudMapNamespace` -- An AWS Cloud Map namespace (HTTP-XOR-private-DNS-XOR-public-DNS) with its discoverable services and statically registered instances managed in-line - the service-discovery registry ECS and custom applications look each other up in. AwsVpc is a prerequisite because the private-DNS arm binds its hosted zone to a referenced VPC.
- `AwsAppSyncApi` -- An AppSync API - AWS's managed API service, as a GraphQL API (SDL schema, resolvers over data sources, caching, the MERGED federation variant) XOR an Events API (real-time pub/sub over channel namespaces) - with its data sources, resolvers, functions, types, API keys, and custom domain managed in-line. Every backend reference (data source targets, roles, the certificate) is optional, so no registry prerequisite - lanes exercise the fixture-free arms.
- `AwsLambdaLayer` -- A Lambda layer version - a shared code archive (libraries, custom runtimes) functions attach by ARN - with its cross-account and organization share grants managed in-line. The archive lives in S3 (an optional reference, so no registry prerequisite - lanes compose their own bucket fixture). 1351 sits in the app & data services sub-band (1350-1359; 1350 opens it with AwsAppSyncApi).
- `AwsRdsProxy` -- An RDS Proxy - the managed connection pool between connection-hungry applications and a database - with its connection-pool tuning, additional endpoints, and database target managed in-line. AwsIamRole is a prerequisite because the proxy assumes a required role to read database credentials from Secrets Manager; AwsSubnet because the proxy's network interfaces require at least two subnets.
- `AwsAuroraDsql` -- An Aurora DSQL cluster - serverless, PostgreSQL-compatible distributed SQL with active-active multi-region pairing managed in-line. No prerequisites: a single-region cluster deploys from defaults alone (the KMS and peer references are optional arms).
- `AwsEcrRegistrySettings` -- Region settings singleton (one private ECR registry per account+region): the registry policy, scanning configuration, replication rules, pull-through cache rules, repository creation templates, account settings, and pull-time update exclusions. Repository-scoped surface stays on AwsEcrRepo.
- `AwsPrivateCa` -- An AWS Private Certificate Authority with composed activation (a ROOT self-signs at apply; a subordinate activates from a parent AwsPrivateCa), issued certificates, the ACM renewal permission, and the resource policy managed in-line. No prerequisites: the S3 (CRL) and parent-CA references are optional arms.
- `AwsSesAccountSettings` -- Account/region settings singleton (one SES account object per account+region): the suppression list and VDM posture. 1360 opens the SES P1 sub-band (1360-1369).
- `AzureResourceGroup` -- 2000–2999: Azure resources
- `AzureAksCluster` -- AzureResourceGroup is the only required parent: the cluster is created inside a referenced resource group. Subnet is optional on the default node pool (AKS provisions managed networking when unset).
- `AzureAksNodePool` -- AzureAksCluster is a prerequisite because a node pool attaches to an existing cluster by ARM ID; the resource group chains transitively.
- `AzureContainerRegistry` -- AzureResourceGroup is a prerequisite because a container registry is created inside a resource group.
- `AzureDnsZone` -- AzureResourceGroup is a prerequisite because the DNS zone is created inside a referenced resource group that must already exist.
- `AzureKeyVault` -- AzureResourceGroup is a prerequisite because a key vault is created inside a referenced resource group in composed environments.
- `AzureVirtualNetwork` -- AzureResourceGroup is a prerequisite because a virtual network is created inside a referenced resource group in composed environments.
- `AzureNatGateway` -- AzureResourceGroup is a prerequisite because a NAT gateway is created inside a referenced resource group in composed environments.
- `AzureVirtualMachine` -- AzureNetworkInterface is a prerequisite because a virtual machine attaches at least one NIC (the subnet, network, and resource group chain transitively through the NIC's own prerequisites).
- `AzureStorageAccount` -- AzureResourceGroup is a prerequisite because a storage account is created inside a referenced resource group in composed environments.
- `AzureDnsRecord` -- AzureDnsZone is a prerequisite because a record set is created inside a referenced zone (the resource group chains transitively through the zone). Public DNS zone names are not globally unique, so a shared zone fixture is safe to recreate across scenarios.
- `AzureSubnet` -- AzureVirtualNetwork is a prerequisite because a subnet is an ARM child of a referenced network -- the network must exist before the subnet can be written. (The resource group arrives transitively through the network's own prerequisite declaration.)
- `AzureNetworkSecurityGroup` -- AzureResourceGroup is a prerequisite because a network security group is created inside a referenced resource group in composed environments.
- `AzurePublicIp` -- AzureResourceGroup is a prerequisite because a public IP is created inside a referenced resource group in composed environments.
- `AzurePrivateEndpoint` -- AzureSubnet is a prerequisite because a private endpoint draws its private IP from a referenced subnet (the virtual network and resource group chain transitively through the subnet's own prerequisite). The connection target is polymorphic and the DNS zones / ASGs are optional, so none of those are prerequisites.
- `AzurePrivateDnsZone` -- AzureResourceGroup is a prerequisite because a private DNS zone is created inside a referenced resource group in composed environments.
- `AzureApplicationGateway` -- AzureSubnet is a prerequisite because a gateway cannot exist without its dedicated gateway_ip_configuration subnet (the network and resource group chain transitively through the subnet's own prerequisites); public frontends additionally reference a public IP, but private-only gateways are legal, so it is not a registry prerequisite.
- `AzureLoadBalancer` -- AzureResourceGroup is a prerequisite because a load balancer is created inside a referenced resource group (frontends additionally reference subnets or public IPs, but neither is universally required, so they are not registry prerequisites).
- `AzureRouteTable` -- AzureResourceGroup is a prerequisite because a route table is created inside a referenced resource group in composed environments.
- `AzurePrivateDnsZoneVirtualNetworkLink` -- AzurePrivateDnsZone and AzureVirtualNetwork are prerequisites because a virtual network link is a child resource of a referenced zone and binds it to a referenced network -- both must exist before the link can be written. (The resource group arrives transitively through the zone's and network's own prerequisite declarations.)
- `AzureVirtualNetworkPeering` -- AzureVirtualNetwork is a prerequisite because a peering is an ARM child of its local network and binds it to a remote network -- the local network must exist before the peering can be written. (The resource group arrives transitively through the network's own prerequisite declaration.)
- `AzurePublicIpPrefix` -- AzureResourceGroup is a prerequisite because a public IP prefix is created inside a referenced resource group in composed environments.
- `AzureNetworkInterface` -- AzureSubnet is a prerequisite because a network interface's IP configurations deploy into a subnet (the virtual network and resource group chain transitively through the subnet's own prerequisite).
- `AzureManagedDisk` -- AzureResourceGroup is a prerequisite because a managed disk is created inside a resource group.
- `AzureVirtualMachineScaleSet` -- AzureSubnet is a prerequisite because every scale-set instance's network interface deploys into a subnet (the virtual network and resource group chain transitively through the subnet's own prerequisite).
- `AzureKeyVaultKey` -- AzureKeyVault is a prerequisite because a key is a data-plane object inside a referenced vault -- the vault must exist before the key can be written (the resource group chains transitively through the vault's own prerequisite).
- `AzureKeyVaultCertificate` -- AzureKeyVault is a prerequisite because a certificate is a data-plane object inside a referenced vault -- the vault must exist before the certificate can be enrolled or imported (the resource group chains transitively through the vault's own prerequisite).
- `AzureKeyVaultSecret` -- AzureKeyVault is a prerequisite because a secret is a data-plane object inside a referenced vault -- the vault must exist before the secret can be written (the resource group chains transitively through the vault's own prerequisite). Part of the Key Vault family (2005, 2025-2026) despite the out-of-run number -- enum numbers are pinned by the registry snapshot; never renumber.
- `AzureWebApplicationFirewallPolicy` -- AzureResourceGroup is a prerequisite because a WAF policy is created inside a referenced resource group; the Application Gateways that attach the policy reference it, never the reverse.
- `AzureApplicationSecurityGroup` -- AzureResourceGroup is a prerequisite because an application security group is created inside a referenced resource group; network interfaces, scale-set IP configurations, and NSG security rules reference the group, never the reverse.
- `AzureDiskEncryptionSet` -- AzureKeyVaultKey is a prerequisite because a disk encryption set wraps customer data with a referenced key -- the key (and its vault, which chains transitively) must exist before the set can resolve the key URL at create time.
- `AzurePostgresqlFlexibleServer` -- AzureResourceGroup is a prerequisite because a server is created inside a referenced resource group (VNet injection additionally references a delegated subnet and a private DNS zone, but neither is universally required, so they are not registry prerequisites).
- `AzureRedisCache` -- AzureResourceGroup is a prerequisite because the cache is created inside a referenced resource group (VNet injection additionally references a dedicated subnet, but only the Premium tier supports it, so it is not a registry prerequisite).
- `AzureCosmosdbAccount` -- AzureResourceGroup is a prerequisite because the account is created inside a referenced resource group.
- `AzureMssqlServer` -- AzureResourceGroup is a prerequisite because the logical server is created inside a referenced resource group.
- `AzureMysqlFlexibleServer` -- AzureResourceGroup is a prerequisite because a server is created inside a referenced resource group (VNet injection additionally references a delegated subnet and a private DNS zone, but neither is universally required, so they are not registry prerequisites).
- `AzureMssqlDatabase` -- The parent logical server is referenced via server_id, not auto-deployed: E2E scenarios declare their own server fixture (minimal-server.yaml or the pool-attach chain through AzureMssqlElasticPool) so sequential subtests never destroy and recreate the same globally unique server_name.
- `AzureMssqlElasticPool` -- AzureMssqlServer is a prerequisite because every elastic pool lives on a referenced logical server (the server's resource group is transitive).
- `AzureRedisLinkedServer` -- The target and linked caches are referenced via ARM ids, not auto-deployed: caches are the slowest-provisioning resources in the Azure catalog and their names are globally unique, so E2E scenarios declare their own cache fixtures instead of a registry prerequisite recreating a shared one per run.
- `AzureRedisCacheAccessPolicy` -- The parent cache is referenced via redis_cache_id, not auto-deployed: caches are the slowest-provisioning resources in the Azure catalog and their names are globally unique, so E2E scenarios declare their own cache fixtures instead of a registry prerequisite recreating a shared one per run.
- `AzureRedisCacheAccessPolicyAssignment` -- The parent cache is referenced via redis_cache_id, not auto-deployed: caches are the slowest-provisioning resources in the Azure catalog and their names are globally unique, so E2E scenarios declare their own cache fixtures instead of a registry prerequisite recreating a shared one per run.
- `AzureContainerAppEnvironment` -- AzureResourceGroup is a prerequisite because the environment is created inside a referenced resource group that must already exist.
- `AzureContainerApp` -- AzureContainerAppEnvironment is a prerequisite because every app runs inside a referenced environment (the resource group arrives transitively through it).
- `AzureServicePlan` -- AzureResourceGroup is a prerequisite because the plan is created inside a referenced resource group that must already exist.
- `AzureFunctionApp` -- AzureServicePlan is a prerequisite because a function app runs on a referenced plan (the resource group arrives transitively through the plan). The required storage account is deliberately NOT a registry prerequisite: storage-account names are globally unique, so scenarios bring their own scenario-local account fixtures.
- `AzureLinuxWebApp` -- AzureServicePlan is a prerequisite because a web app runs on a referenced plan (the resource group arrives transitively through the plan).
- `AzureContainerAppJob` -- AzureContainerAppEnvironment is a prerequisite because a job runs inside a referenced environment (the resource group arrives transitively through it).
- `AzureContainerAppEnvironmentStorage` -- AzureContainerAppEnvironment is a prerequisite because the storage registration lives on a referenced environment. The Azure Files share and storage account are deliberately NOT registry prerequisites: storage-account names are globally unique, so scenarios bring their own scenario-local account + share fixtures.
- `AzureContainerAppEnvironmentDaprComponent` -- AzureContainerAppEnvironment is a prerequisite because the Dapr component is registered on a referenced environment.
- `AzureContainerAppEnvironmentCertificate` -- AzureContainerAppEnvironment is a prerequisite because the certificate is stored on a referenced environment.
- `AzureContainerAppEnvironmentManagedCertificate` -- AzureContainerAppEnvironment is a prerequisite because the managed certificate is provisioned on a referenced environment.
- `AzureLogAnalyticsWorkspace` -- AzureResourceGroup is a prerequisite because the workspace is created inside a referenced resource group that must already exist.
- `AzureApplicationInsights` -- AzureLogAnalyticsWorkspace is a prerequisite because workspace-based Application Insights stores its telemetry in a referenced workspace (the resource group chains transitively through the workspace).
- `AzureMonitorDiagnosticSetting` -- AzureLogAnalyticsWorkspace is a prerequisite because the setting's scenarios route a fixture workspace's telemetry (the workspace doubles as target and destination); the target itself is polymorphic.
- `AzureMonitorActionGroup` -- AzureResourceGroup is a prerequisite because the action group is created inside a referenced resource group that must already exist.
- `AzureMonitorMetricAlert` -- AzureMonitorActionGroup is a prerequisite because a metric alert's actions fire into a referenced action group (the resource group chains transitively); alert scopes are polymorphic.
- `AzureMonitorScheduledQueryAlert` -- AzureLogAnalyticsWorkspace is a prerequisite because the rule queries a referenced workspace scope; AzureMonitorActionGroup because its action fires into a referenced action group.
- `AzureMonitorActivityLogAlert` -- AzureMonitorActionGroup is a prerequisite because an activity log alert's actions fire into a referenced action group (the resource group chains transitively). The alert itself is subscription-global and its scopes are polymorphic.
- `AzureApplicationInsightsStandardWebTest` -- AzureApplicationInsights is a prerequisite because a standard web test binds to a referenced Application Insights component (the resource group chains transitively through the kind).
- `AzureUserAssignedIdentity` -- AzureResourceGroup is a prerequisite because the identity is created inside a referenced resource group that must already exist.
- `AzureRoleAssignment` -- AzureResourceGroup and AzureUserAssignedIdentity are prerequisites because an assignment grants a role at a referenced scope (most commonly a resource group) to a referenced principal (most commonly a managed identity) -- both must exist before the grant can be written.
- `AzureRoleDefinition` -- AzureResourceGroup is a prerequisite because a custom role definition is created at a referenced scope, most commonly a resource group in composed environments -- the scope must exist before the definition can be written.
- `AzureFederatedIdentityCredential` -- AzureUserAssignedIdentity is the prerequisite because a federated identity credential is a child resource of a referenced managed identity -- the identity must exist before the credential can be written on it. (The resource group arrives transitively through the identity's own prerequisite declaration.)
- `AzureServiceBusNamespace` -- AzureResourceGroup is a prerequisite because a Service Bus namespace is created inside a referenced resource group in composed environments. The namespace is the container every Service Bus messaging entity (queue, topic, subscription, authorization rule, geo-DR pairing) nests under. The child kinds deliberately declare NO registry prerequisite: namespace names are globally unique with a post-delete name hold, so E2E composes them with scenario-local namespace fixtures instead of a shared recreate-per-scenario prerequisite.
- `AzureEventHubNamespace` -- AzureResourceGroup is a prerequisite because an Event Hub namespace is created inside a referenced resource group in composed environments. The namespace is the container every Event Hubs entity (event hub, consumer group, authorization rule, schema group, geo-DR pairing, customer-managed key) nests under. The child kinds deliberately declare NO registry prerequisite: namespace names are globally unique with a post-delete name hold, so E2E composes them with scenario-local namespace fixtures instead of a shared recreate-per-scenario prerequisite.
- `AzureServiceBusQueue`
- `AzureServiceBusTopic`
- `AzureServiceBusSubscription`
- `AzureServiceBusAuthorizationRule`
- `AzureServiceBusDisasterRecoveryConfig`
- `AzureEventHub`
- `AzureEventHubConsumerGroup`
- `AzureEventHubAuthorizationRule`
- `AzureFrontDoorProfile` -- AzureResourceGroup is a prerequisite because a Front Door profile is created inside a referenced resource group in composed environments. The profile is the container every Front Door delivery resource (endpoint, origin group, origin, route) nests under.
- `AzureFrontDoorEndpoint` -- AzureFrontDoorProfile is a prerequisite because an endpoint is an ARM child of a referenced profile -- the profile must exist before the endpoint can be written. (The resource group arrives transitively through the profile's own prerequisite declaration.)
- `AzureFrontDoorOriginGroup` -- AzureFrontDoorProfile is a prerequisite because an origin group is an ARM child of a referenced profile. A container kind: every origin is an ARM child of its origin group, so on a diagram the origins stand inside the group that load-balances them (the profile's room holds the group). A route or a rule that forwards TO an origin group lives in its endpoint or its rule set, never in the group, and those references say so with containment_exempt.
- `AzureFrontDoorOrigin` -- AzureFrontDoorOriginGroup is a prerequisite because an origin is an ARM child of a referenced origin group (the profile and resource group chain transitively).
- `AzureFrontDoorRoute` -- A route attaches to an endpoint (its ARM parent) and forwards to an origin group whose origins must exist before ARM accepts the route -- so both the endpoint and the origin chain are genuine deploy-order prerequisites.
- `AzureFrontDoorRuleSet` -- AzureFrontDoorProfile is a prerequisite because a rule set is an ARM child of a referenced profile. The rules live inside the set (they form one ordered policy document); routes attach the set by ARM ID.
- `AzureFrontDoorCustomDomain` -- AzureFrontDoorProfile is a prerequisite because a custom domain is an ARM child of a referenced profile. The DNS zone and (for bring-your-own certificates) the Front Door secret are optional references, not deploy-order prerequisites.
- `AzureFrontDoorSecret` -- AzureFrontDoorSecret is a prerequisite-light kind: only the profile (its ARM parent) must exist. The Key Vault certificate it wraps is a reference resolved before the module runs; its vault chain is exercised through scenario-local fixtures in E2E.
- `AzureFrontDoorFirewallPolicy` -- AzureResourceGroup is a prerequisite because the Front Door WAF policy is created inside a referenced resource group -- it is a GLOBAL resource, not a profile child (a different ARM type than the regional Application Gateway WAF policy). Security policies attach it to profiles; the policy itself depends on nothing else.
- `AzureFrontDoorSecurityPolicy` -- A security policy is an ARM child of a profile that associates a referenced WAF policy with referenced domains -- so the endpoint (the default-domain association target; the profile arrives transitively through it) and the WAF policy are genuine deploy-order prerequisites.
- `AzureStorageContainer` -- None of the storage data-service kinds declares a registry prerequisite on AzureStorageAccount: account names are GLOBALLY unique and Azure holds a just-deleted name, so a recreate-per-scenario fixture would hang -- their E2E scenarios declare scenario-local account fixtures instead. Deploy ordering in composed environments still flows from the storage_account_id reference itself.
- `AzureStorageShare`
- `AzureStorageQueue`
- `AzureStorageTable`
- `AzureStorageEncryptionScope`
- `AzureStorageDataLakeGen2Filesystem`
- `AzureStorageLocalUser`
- `AzureStorageObjectReplication`
- `AzureCosmosdbSqlDatabase` -- None of the Cosmos DB data-service kinds declares a registry prerequisite on AzureCosmosdbAccount: account names are GLOBALLY unique DNS labels, so a recreate-per-scenario fixture would risk name-reuse hangs -- their E2E scenarios declare scenario-local account fixtures instead. Deploy ordering in composed environments still flows from the cosmosdb_account_id / parent-database references themselves.
- `AzureCosmosdbSqlContainer`
- `AzureCosmosdbMongoDatabase`
- `AzureCosmosdbMongoCollection`
- `AzureCosmosdbSqlRoleDefinition`
- `AzureCosmosdbSqlRoleAssignment`
- `AzureManagedRedis` -- AzureResourceGroup is the cluster's only registry prerequisite: the cluster is created inside a referenced resource group. The geo-replication and access-policy-assignment children declare NO prerequisite on AzureManagedRedis: clusters are expensive, slow-provisioning parents, so their E2E scenarios declare scenario-local cluster fixtures instead of recreating a shared one per scenario. Deploy ordering in composed environments still flows from the managed_redis_id references themselves.
- `AzureManagedRedisGeoReplication`
- `AzureManagedRedisAccessPolicyAssignment`
- `AzureEventHubDisasterRecoveryConfig`
- `AzureEventHubSchemaGroup`
- `AzureEventHubCluster` -- AzureResourceGroup is a prerequisite because a dedicated Event Hubs cluster is created inside a referenced resource group in composed environments. Note: clusters cannot be deleted for 4 hours after creation (Azure's moratorium), so E2E treats this kind as offline-gated.
- `AzureEventHubNamespaceCustomerManagedKey`
- `AzureMssqlFailoverGroup` -- AzureMssqlServer is a prerequisite because a failover group is created on a referenced primary logical server and points at a partner server; the primary (and its resource group, which chains transitively) must exist before the group can be written.
- `AzureContainerAppCustomDomain` -- AzureContainerApp is a prerequisite because the domain binding lives in a referenced app's ingress configuration (the environment and resource group chain transitively through the app).
- `AzureFirewallPolicy`
- `AzureFirewallPolicyRuleCollectionGroup` -- AzureFirewallPolicy is a prerequisite because a rule collection group is a child document of a referenced policy (the resource group chains transitively through the policy).
- `AzureFirewall` -- AzureSubnet is a prerequisite because a VNet-deployed firewall's data path lives in a dedicated subnet that must be named exactly "AzureFirewallSubnet" (the virtual network and resource group chain transitively through the subnet). The E2E install profile publishes a fixture subnet with that exact name and a /26 prefix.
- `AzureIpGroup`
- `AzureVirtualNetworkGateway` -- AzureSubnet is a prerequisite because every virtual network gateway lives in a dedicated subnet named exactly "GatewaySubnet" (the virtual network and resource group chain transitively through the subnet); the subnet install profile publishes a fixture instance with that exact ARM name. AzurePublicIp is a prerequisite because a VPN-type gateway (the default shape) requires a public IP per ip configuration; the address install profile publishes a dedicated zone-redundant instance (a gateway binds its address exclusively, and the AZ gateway SKUs require zones on it).
- `AzureVirtualNetworkGatewayConnection` -- Both gateways are prerequisites: a connection joins a virtual network gateway to a far side, and the site-to-site far side is a local network gateway (the GatewaySubnet, VNet, and resource group chain transitively through the virtual network gateway).
- `AzureLocalNetworkGateway`
- `AzurePrivateLinkService` -- AzureSubnet is the sole prerequisite: every NAT ip configuration draws its address from a subnet with private-link-service network policies disabled (the subnet install profile publishes a fixture instance with that flag off). The Standard load balancer whose frontend the service typically fronts is NOT a registry prerequisite -- the spec's destination is an exactly-one-of (load balancer frontend OR fixed destination IP), so scenarios that use the load-balancer shape declare it via the planton.dev/e2e-prerequisites annotation instead.
- `AzureExpressRouteCircuit`
- `AzureExpressRouteCircuitPeering` -- The circuit is the prerequisite: a peering is an ARM child of the circuit, addressed by the circuit's name (the resource group chains transitively through the circuit).
- `AzureExpressRouteGateway` -- The hub is the prerequisite: ARM requires an ExpressRoute Gateway to be deployed INTO a Virtual WAN hub (the WAN and resource group chain transitively through the hub).
- `AzureExpressRoutePort` -- ExpressRoute Port: your own physical port pair on a Microsoft edge router (ExpressRoute Direct), from whose bandwidth circuits are carved. Self-contained -- only the resource group is required.
- `AzureVirtualWan` -- Virtual WAN: the umbrella of Azure's managed hub-and-spoke networking, under which virtual hubs and their gateways are created. Self-contained -- only the resource group is required. A container: its hubs and its branch sites are created into it, so on a diagram the WAN is the room they stand in.
- `AzureVirtualHub` -- The WAN is the prerequisite: this kind models the Virtual WAN hub (virtual_wan_id is required; standalone hubs are the legacy Route Server construction, which has its own ARM surface). The resource group chains transitively through the WAN. A container: ARM deploys the hub's VPN, ExpressRoute, and point-to-site gateways and its spoke connections INTO the hub, so on a diagram the hub is the room they stand in.
- `AzureVirtualHubConnection` -- Both sides of the attachment are prerequisites: the hub being joined and the spoke virtual network being attached.
- `AzureVpnGateway` -- The hub is the prerequisite: ARM deploys a Virtual WAN VPN gateway INTO a virtual hub (virtual_hub_id is required and immutable; the WAN and resource group chain transitively through the hub). ARM allows one VPN gateway per hub.
- `AzureVpnGatewayConnection` -- Both ends of the tunnel are prerequisites: a connection is an ARM child of the VPN gateway and pins each of its links to a specific link of the remote VPN site (the hub, WAN, and resource group chain transitively through the gateway).
- `AzureVpnSite` -- The WAN is the prerequisite: a VPN site is the Virtual WAN world's address-book entry for one branch location (virtual_wan_id is required; the classic-world sibling without a WAN is AzureLocalNetworkGateway). The resource group chains transitively through the WAN.
- `AzurePointToSiteVpnGateway` -- The hub and the server configuration are both prerequisites: a point-to-site VPN gateway deploys INTO a virtual hub (one P2S gateway per hub, a slot separate from the hub's site-to-site VPN gateway) and is born pointing at the VPN server configuration that defines how its users authenticate -- both ARM-required and fixed at creation. The WAN and resource group chain transitively through the hub.
- `AzureVpnServerConfiguration` -- Self-contained -- only the resource group is required: a VPN server configuration is the reusable "who may connect and how" authentication policy (Entra ID / certificate / RADIUS) that point-to-site VPN gateways attach to; it references no other Azure resource.
- `AzureCognitiveAccount` -- Self-contained -- only the resource group is required: an Azure AI services account (Azure OpenAI, the multi-service AIServices account, the single-service accounts) needs no other Azure resource; subnets (network rules), Key Vault keys (CMK), storage accounts and user-assigned identities are optional references. A container kind: every model deployment and every project is an ARM child created onto the account (accounts/{account}/deployments/{name}, .../projects/{name}), so a diagram draws them inside it. The subnets the account admits through its network ACLs or injects agent compute into are access, never placement (containment_exempt on those fields).
- `AzureCognitiveDeployment` -- An ARM child of its account: a model deployment (which model runs, at which throughput class) exists only on an Azure AI services account of kind "OpenAI" or "AIServices".
- `AzureCognitiveAccountProject` -- An ARM child of its account: an AI Foundry project exists only on an "AIServices"-kind account with project management enabled.
- `AzureMachineLearningWorkspace` -- The workspace REQUIRES all three companion services at creation (default storage, secrets vault, telemetry) -- genuine deploy-order prerequisites, each with its own fixture profile. A container kind: datastores, compute clusters and instances, and online and batch endpoints are ARM children of the workspace (workspaces/{ws}/datastores|computes|onlineEndpoints|batchEndpoints/{name}), so a diagram draws them inside it. The companion services it names are reaches (containment_exempt on those fields), and so is the subnet its serverless compute or a compute's nodes attach to.
- `AzureMachineLearningDatastore` -- An ARM child of its workspace. The storage target (container, filesystem or share) is scenario-declared via the e2e-prerequisites annotation -- only the blob scenario needs a container, so it is not a kind-wide prerequisite.
- `AzureMachineLearningComputeCluster` -- An ARM child of its workspace (.../computes/{name}) -- the auto-scaling pool of VMs training jobs run on.
- `AzureMachineLearningComputeInstance` -- An ARM child of its workspace (.../computes/{name}) -- a single always-on VM serving as one data scientist's cloud workstation.
- `AzureAiFoundry` -- The hub REQUIRES both companion services at creation (secrets vault, default storage) -- genuine deploy-order prerequisites, each with its own fixture profile. A container kind: every AI Foundry project is created inside its hub and deploys into the hub's resource group (the project spec carries none), so a diagram draws the projects inside the hub. The vault and storage account the hub names are reaches (containment_exempt on those fields).
- `AzureAiFoundryProject` -- Deploys into its hub's resource group (the provider derives the group from the hub reference -- the project spec carries none).
- `AzureSearchService`
- `AzureMachineLearningOnlineEndpoint` -- An ARM child of its workspace (.../onlineEndpoints/{name}) -- the stable scoring address applications call. azurerm carries no ML endpoint resources; the modules write the raw ARM shape at a pinned api-version (azapi / azure-native). A container kind in its turn: every online deployment is an ARM child created under the endpoint whose traffic map routes to it (.../onlineEndpoints/{endpoint}/deployments/{name}) and cannot exist without it, so a diagram draws the deployments inside the endpoint, inside the workspace.
- `AzureMachineLearningOnlineDeployment` -- An ARM child of its endpoint (.../deployments/{name}) -- the running copy of a model the endpoint's traffic map routes to.
- `AzureMachineLearningBatchEndpoint` -- An ARM child of its workspace (.../batchEndpoints/{name}) -- the stable address batch scoring jobs are submitted to. azurerm carries no ML endpoint resources; the modules write the raw ARM shape at a pinned api-version (azapi / azure-native). A container kind in its turn: every batch deployment is an ARM child created under the endpoint whose default-deployment pointer routes to it (.../batchEndpoints/{endpoint}/deployments/{name}), so a diagram draws the deployments inside the endpoint, inside the workspace. The compute cluster a deployment runs on is a reach, not a room.
- `AzureMachineLearningBatchDeployment` -- An ARM child of its endpoint (.../deployments/{name}) -- the job recipe (model, compute, batching behavior) the endpoint's default-deployment pointer routes submissions to.
- `AzureRecoveryServicesVault` -- The Recovery Services vault (Microsoft.RecoveryServices/vaults) -- the safe that classic Azure Backup data and Site Recovery configuration live in. Backup policies and protected items are ARM children of a vault, so the vault is a container kind: a diagram draws them inside it, as the portal's own blades do.
- `AzureBackupPolicyVm` -- An ARM child of its vault (.../backupPolicies/{name}) -- the schedule and retention rules that govern IaaS VM backups.
- `AzureBackupProtectedVm` -- An ARM child of its vault (.../protectedItems/...) -- the binding that puts one virtual machine under a backup policy's protection.
- `AzureBackupPolicyFileShare` -- An ARM child of its vault (.../backupPolicies/{name}) -- the schedule and retention rules that govern Azure Files share backups (snapshot or vaulted).
- `AzureBackupProtectedFileShare` -- An ARM child of its vault (.../protectedItems/AzureFileShare;...) -- the binding that puts one Azure Files share under a backup policy's protection. The share's storage account must already be registered with the vault (AzureBackupContainerStorageAccount). Prerequisite ORDER is load-bearing for teardown (destroy runs in reverse): the registration must list AFTER the share so it unregisters FIRST -- Azure Backup holds a DoNotDelete lock on a registered storage account, and a share delete under that lock fails ScopeLocked.
- `AzureDataProtectionBackupVault` -- The Data Protection backup vault (Microsoft.DataProtection/ backupVaults) -- the safe that MODERN Azure Backup data lives in (managed disks, blob storage, AKS clusters, MySQL/PostgreSQL flexible servers, Data Lake storage). Backup policies and backup instances are ARM children of a vault, so the vault is a container kind: a diagram draws them inside it, as the portal's own blades do.
- `AzureDataProtectionBackupPolicy` -- An ARM child of its vault (.../backupPolicies/{name}) -- the schedule and retention rules for ONE Data Protection datasource type (blob storage, disk, Kubernetes cluster, MySQL/PostgreSQL flexible server, or Data Lake storage), modeled as one kind with variant blocks.
- `AzureDataProtectionBackupInstance` -- An ARM child of its vault (.../backupInstances/{name}) -- the binding that puts ONE datasource (a managed disk, a storage account's blob services, an AKS cluster, a MySQL/PostgreSQL flexible server, or a Data Lake storage account) under a Data Protection backup policy, modeled as one kind with variant blocks. The vault's managed identity must hold the datasource roles Azure Backup requires BEFORE the instance is created.
- `AzureBastionHost` -- AzureSubnet and AzurePublicIp are prerequisites because a dedicated-infrastructure Bastion host (Basic/Standard/Premium -- the default shapes) deploys into a subnet named exactly "AzureBastionSubnet" and binds a Standard static public IP EXCLUSIVELY (the virtual network and resource group chain transitively through the subnet). The Developer SKU instead attaches to a virtual network directly and uses neither.
- `AzureNetworkWatcherFlowLog` -- AzureVirtualNetwork and AzureStorageAccount are prerequisites because a flow log records a network-scoped target (a virtual network in the common case; subnets and network interfaces chain through the network) into a referenced storage account. The regional Network Watcher parent is NOT a prerequisite: Azure auto-creates it ("NetworkWatcher_{region}" in "NetworkWatcherRG") the moment the region hosts a virtual network, and the flow log references it by name. Traffic Analytics' Log Analytics workspace is an optional arm, declared by scenarios that use it.
- `AzurePrivateDnsResolver` -- AzureVirtualNetwork and AzureSubnet are prerequisites because a DNS Private Resolver anchors to a referenced virtual network (at most ONE resolver per network -- Azure enforces it) and each of its inbound/outbound endpoints occupies its own dedicated subnet delegated to "Microsoft.Network/dnsResolvers" (the resource group chains transitively through the network and subnets).
- `AzurePrivateDnsResolverForwardingRuleset` -- AzurePrivateDnsResolver is a prerequisite because a DNS forwarding ruleset steers a resolver's OUTBOUND endpoints -- it binds their ARM ids (at most 2, same resolver) at creation. (The resource group and network chain transitively through the resolver's own prerequisite declarations.)
- `AzurePrivateDnsRecord` -- AzurePrivateDnsZone is a prerequisite because every record set is created inside a referenced private DNS zone (the resource group chains transitively through the zone's own prerequisite).
- `AzureTrafficManagerProfile` -- AzureResourceGroup is a prerequisite because a Traffic Manager profile is created inside a referenced resource group (the profile itself is a global service -- the group only holds its metadata record). A container kind: every endpoint is an ARM child of its profile, so on a diagram the endpoints stand inside the profile that steers traffic to them. A nested endpoint that points AT another profile lives in its own parent profile, never in the one it targets, and that reference says so with containment_exempt.
- `AzureTrafficManagerEndpoint` -- AzureTrafficManagerProfile is a prerequisite because every endpoint is created inside a referenced profile -- it is the destination a profile steers traffic to (the resource group chains transitively through the profile's own prerequisite).
- `AzureMonitorAutoscaleSetting` -- AzureResourceGroup is a prerequisite because an autoscale setting is created inside a referenced resource group. The scalable TARGET it controls is a no-default reference (many kinds can be scaled), so no target kind is declared here -- scenarios declare their own target fixture.
- `AzureMonitorDataCollectionRule` -- The Azure Monitor data collection rule (DCR) -- the routing table declaring what telemetry the Azure Monitor Agent collects and where it lands. AzureResourceGroup is a prerequisite because a rule is created inside a referenced resource group; AzureLogAnalyticsWorkspace because a workspace is the canonical destination a rule routes to (the smoke scenario's shape). Machines attach to a rule with AzureMonitorDataCollectionRuleAssociation resources.
- `AzureEventgridTopic` -- The Azure Event Grid custom topic -- the HTTPS endpoint an application publishes its own events to, fanned out to handlers by event subscriptions. One topic is one event stream with its own endpoint and access keys; for many streams behind one endpoint see AzureEventgridDomain. A container kind: every event subscription on the topic is an ARM child created under it ({topic_id}/providers/ Microsoft.EventGrid/eventSubscriptions/{name}) and cannot exist without it, so on a diagram the topic is the room its subscriptions stand in, as the portal's own blade lists them. A kind that merely publishes to or subscribes from a topic while living elsewhere (a Data Factory trigger, an Event Grid namespace's MQTT route) says so with containment_exempt.
- `AzureEventgridDomain` -- The Azure Event Grid domain -- ONE publishing endpoint and one pair of access keys serving many event streams (domain topics), the multi-tenant pattern. Topics inside the domain are auto-managed by Azure or declared explicitly as AzureEventgridDomainTopic resources. A container kind: every domain topic is an ARM child of its domain ({domain_id}/topics/{name}), so on a diagram the domain is the room its tenant streams stand in.
- `AzureEventgridSystemTopic` -- The Azure Event Grid system topic -- the subscription surface for events AZURE ITSELF publishes about one of your resources (a storage account's blob events, a resource group's lifecycle events). One system topic per source resource per topic type; event subscriptions attach to it to route events to handlers. A container kind: every subscription on a system topic is an ARM child created under it ({system_topic_id}/eventSubscriptions/{name}), so on a diagram the system topic is the room its subscriptions stand in.
- `AzureEventgridEventSubscription` -- The Azure Event Grid event subscription -- the delivery instruction routing events from a source (a custom topic, domain, domain topic, system topic, resource group, or subscription) to a handler (a Function, Event Hub, Service Bus queue/topic, storage queue, hybrid connection, or webhook), with filtering, retry, and dead-letter behavior.
- `AzureEventgridNamespace` -- The Azure Event Grid namespace -- the capacity-scaled hub of the newer Event Grid: hosts CloudEvents namespace topics and an optional MQTT broker behind one set of regional endpoints, sized in throughput units. A container kind: every namespace topic is an ARM child of its namespace ({namespace_id}/topics/{name}), so on a diagram the namespace is the room its streams stand in.
- `AzureDataFactory` -- The Azure Data Factory -- the workspace every other Data Factory resource lives inside: pipelines, data flows, linked services, datasets, triggers, and integration runtimes are all created against a factory's ARM ID. A container kind for exactly that reason: those six kinds are ARM children of the factory, so on a diagram the factory is the room they stand in.
- `AzureDataFactoryPipeline` -- One unit of work inside an Azure Data Factory ({factory_id}/pipelines/{name}) -- an ordered set of activities that executes as a whole when triggered.
- `AzureDataFactoryDataFlow` -- A Data Factory data flow ({factory_id}/dataflows/{name}) -- a visually-designed data transformation executed on managed Spark, or, as a flowlet, a reusable snippet other data flows embed. One kind covers both provider forms (they share one schema and one name namespace inside the factory).
- `AzureDataFactoryLinkedService` -- A Data Factory linked service ({factory_id}/linkedservices/{name}) -- a saved connection in the factory's address book: where an external system lives and how to authenticate to it. One kind covers every connection type Azure models as a first-class resource (storage, SQL family, Cosmos DB, Databricks, Key Vault, SFTP, web APIs, and more) as variants in one factory-scoped name namespace, plus the raw-JSON custom form.
- `AzureDataFactoryDataset` -- A Data Factory dataset ({factory_id}/datasets/{name}) -- a named view of data inside a system a linked service already connects to: which container and path, which table, which file format. One kind covers every data shape Azure models as a first-class dataset resource (delimited text/CSV, JSON, Parquet, binary, blob, HTTP, the SQL family, Snowflake, Cosmos DB) as variants in one factory-scoped name namespace, plus the raw-JSON custom form.
- `AzureDataFactoryTrigger` -- A Data Factory trigger ({factory_id}/triggers/{name}) -- the instruction that starts pipelines automatically: on a clock schedule, per contiguous tumbling window, on storage blob events, or on Event Grid custom events. One kind covers all four provider trigger resources as variants (one ARM namespace, one started/stopped lifecycle).
- `AzureDataFactoryIntegrationRuntime` -- A Data Factory integration runtime ({factory_id}/integrationRuntimes/{name}) -- the compute engine a factory's pipelines, data flows, and copy activities run on. One kind covers all three engine flavors as variants in one factory-scoped name namespace: the managed data-flow compute, the managed SSIS package runtime, and the self-hosted agent registration (which issues the authorization keys agents join with).
- `AzureComputeGallery` -- The Azure Compute Gallery -- the shared library an organization keeps its approved VM images in. Image definitions (AzureComputeGalleryImage) live inside it; VMs and scale sets deploy from their published, region-replicated versions. A container kind: every image definition is an ARM child of its gallery ({gallery_id}/images/{name}), so on a diagram the definitions stand inside the library that publishes them. The image ids a VM, a scale set, or a disk boots from are plain strings and never place anything inside the gallery.
- `AzureComputeGalleryImage` -- A gallery image ({gallery_id}/images/{name}) -- one image definition inside a Compute Gallery (marketplace-style identity, OS type, security posture) plus its published versions, each replicated to its own target regions. VMs deploy from a version's ARM ID or from the definition's ID to get the latest version.
- `AzureAvailabilitySet` -- The availability set -- the classic pre-zones placement grouping that spreads VMs across separate fault and update domains so one hardware failure or maintenance window cannot take them all down. VMs join the set at creation.
- `AzureDiskSnapshot` -- The managed disk snapshot -- a point-in-time copy of a disk used for backup, cloning, and as the source of gallery image versions. Incremental snapshots store only the delta since the previous snapshot of the same disk.
- `AzureContainerInstance` -- The Azure Container Instance container group -- serverless containers billed per second: one or more containers sharing a lifecycle, network, and volumes (plus one-shot init containers), with no cluster or VM to manage. Public, subnet-private, or IP-less postures.
- `AzureFunctionAppFlexConsumption` -- The Azure Function App on the Flex Consumption plan -- Azure's newest serverless Functions hosting model: per-instance memory selection, a configurable scale-out ceiling, always-ready instance pools, and explicit blob-container deployment storage. Requires an FC1-SKU service plan, which is deliberately NOT a registry prerequisite: the shared plan fixture serves the classic app tiers, and an FC1 plan is cheap to create per scenario (no idle compute cost), so scenarios bring their own plan fixture -- the same reasoning that keeps the globally-unique storage account scenario-local for AzureFunctionApp.
- `AzureMongoCluster` -- The Azure Cosmos DB for MongoDB vCore cluster -- Azure's modern managed MongoDB: a real MongoDB engine on dedicated vCore tiers with sharding, zone-redundant HA, and point-in-time restore.
- `AzureFabricCapacity` -- The Microsoft Fabric capacity -- the billing and compute anchor of Microsoft Fabric: workspaces assign themselves to a capacity, and its F-SKU sets how much compute every workload on it shares. azurerm's entire Fabric surface is this one resource (workspaces and items live in Microsoft's dedicated fabric provider).
- `AzureBackupContainerStorageAccount` -- Registers a storage account with a Recovery Services vault as a backup container (.../protectionContainers/StorageContainer;...) -- one registration per storage-account-and-vault pair, required BEFORE any of the account's file shares can be protected. Part of the backup family (2175-2179) despite the out-of-run number -- enum numbers are pinned by the registry snapshot; never renumber.
- `AzureDataProtectionResourceGuard` -- The Data Protection Resource Guard (Microsoft.DataProtection/ resourceGuards) -- the approval gate behind Multi-User Authorization: privileged vault operations (disabling soft delete, reducing retention) require an approval through a guard, which typically lives in a DIFFERENT administrator's scope. Vaults reference a guard by its ARM ID. Part of the backup family (2175-2182) despite the out-of-run number -- enum numbers are pinned by the registry snapshot; never renumber.
- `AzurePrivateDnsResolverVirtualNetworkLink` -- The attachment that makes a DNS forwarding ruleset take effect in one virtual network ({ruleset_id}/virtualNetworkLinks/{name}) -- one link per ruleset-network pair, up to 500 per ruleset, spokes joining and leaving independently (which is why the link is a standalone kind, exactly like AzurePrivateDnsZoneVirtualNetworkLink). Part of the DNS Private Resolver family (2186-2187) despite the out-of-run number -- enum numbers are pinned by the registry snapshot; never renumber.
- `AzureMonitorDataCollectionRuleAssociation` -- The attachment that puts ONE machine under an Azure Monitor data collection rule ({target_id}/providers/Microsoft.Insights/ dataCollectionRuleAssociations/{name}) -- an extension resource on the TARGET machine, many per rule, machines joining and leaving monitoring independently (which is why the association is a standalone kind, exactly like AzurePrivateDnsZoneVirtualNetworkLink). AzureVirtualMachine is a prerequisite because the smoke scenario attaches the fixture VM; the rule prerequisite chains the rule's own install manifest. Part of the Monitor family (2191-2192) despite the out-of-run number -- enum numbers are pinned by the registry snapshot; never renumber.
- `AzureEventgridDomainTopic` -- One named event stream inside an Azure Event Grid domain ({domain_id}/topics/{name}) -- the per-tenant mailbox of the multi-tenant pattern: many per domain, each with its own subscriptions and lifecycle, tenants joining and leaving without touching the domain (which is why the domain topic is a standalone kind, exactly like AzureEventHubConsumerGroup on a shared hub). Part of the Event Grid family (2193-2194) despite the out-of-run number -- enum numbers are pinned by the registry snapshot; never renumber.
- `AzureEventgridNamespaceTopic` -- One named CloudEvents stream inside an Azure Event Grid namespace ({namespace_id}/topics/{name}) -- many per namespace, publishers and teams creating and deleting their own against the shared namespace (which is why the topic is a standalone kind, exactly like AzureEventgridDomainTopic and AzureEventHubConsumerGroup). Part of the Event Grid family (2193-2197) despite the out-of-run number -- enum numbers are pinned by the registry snapshot; never renumber.
- `AzureMongoClusterUser` -- Grants one Microsoft Entra principal access to an Azure Cosmos DB for MongoDB vCore cluster ({cluster_id}/users/{object_id}) -- an access binding, not a password user: many per cluster, principals joining and leaving independently (which is why the grant is a standalone kind, the access-grant class of AzureRoleAssignment). Part of the Mongo vCore family (2211) despite the out-of-run number -- enum numbers are pinned by the registry snapshot; never renumber.
- `AzurePlantonRunner` -- AzureContainerAppEnvironment is a prerequisite because the runner appliance is a Container App, and every Container App runs inside an environment -- the environment reference must resolve before the appliance can deploy.
- `GcpArtifactRegistryRepo` -- 3000–3999: GCP resources
- `GcpTargetHttpsProxy` -- The URL map is the parent a proxy cannot exist without; the classic compute certificate kinds and the SSL policy are the fixture parents the committed scenarios attach. The Certificate Manager certificate list (certificate_manager_certificates, honored only by the cross-region internal ALB) is optional composition -- a scenario that arms it declares GcpCertManagerCert via the e2e-prerequisites annotation, never a registry edge that would tax every proxy and forwarding-rule chain.
- `GcpCloudFunction`
- `GcpCloudRun`
- `GcpCloudSql`
- `GcpDnsZone`
- `GcpGcsBucket`
- `GcpGkeCluster`
- `GcpIamCustomRole`
- `GcpProject`
- `GcpVpcNetwork`
- `GcpSubnetwork`
- `GcpRouterNat`
- `GcpGkeNodePool`
- `GcpServiceAccount`
- `GcpGkeWorkloadIdentityBinding`
- `GcpCertManagerCert`
- `GcpComputeInstance`
- `GcpDnsRecord`
- `GcpProjectIamMember`
- `GcpFirewallRule`
- `GcpGlobalAddress`
- `GcpCloudArmorPolicy`
- `GcpHealthCheck`
- `GcpBackendBucket`
- `GcpBackendService`
- `GcpRegionNetworkEndpointGroup`
- `GcpUrlMap`
- `GcpManagedSslCertificate`
- `GcpTargetHttpProxy`
- `GcpAlloydbCluster`
- `GcpRedisInstance`
- `GcpFirestoreDatabase`
- `GcpSpannerInstance`
- `GcpSpannerDatabase`
- `GcpBigtableInstance`
- `GcpMemorystoreInstance`
- `GcpCloudSqlDatabase`
- `GcpCloudSqlUser`
- `GcpAlloydbInstance`
- `GcpAlloydbUser`
- `GcpSpannerBackupSchedule`
- `GcpBigtableTable`
- `GcpFirestoreBackupSchedule`
- `GcpFirestoreIndex`
- `GcpBigQueryDataset`
- `GcpDataprocCluster`
- `GcpDataprocAutoscalingPolicy`
- `GcpBigQueryTable`
- `GcpBigQueryCapacityCommitment` -- A BigQuery capacity commitment: slots bought for a fixed term in an administration project and location, pooled across every reservation there. A purchase Google will not delete before its term ends.
- `GcpBigQueryReservationGroup` -- A BigQuery reservation group: reservations that share idle slots with each other first. Reservations reference it.
- `GcpDatastreamConnectionProfile` -- A Datastream connection profile: where one source database or destination is and how Datastream signs in. Streams reference a source and a destination profile; one profile serves many streams.
- `GcpDatastreamPrivateConnection` -- A Datastream private connection: the VPC peering or Private Service Connect interface through which Datastream reaches private databases, shared by every profile that reaches that network.
- `GcpPubSubTopic`
- `GcpPubSubSubscription`
- `GcpCloudTasksQueue`
- `GcpCloudSchedulerJob`
- `GcpPubSubSchema`
- `GcpPubSubTopicIamMember` -- One additive grant on a topic. Its own kind, not a field on the topic, because the identities that most need it (a logging sink's writer, a Security Command Center export's publisher) belong to resources that name the topic themselves; a grant on the topic that referenced them back would be a dependency cycle.
- `GcpVertexAiNotebook`
- `GcpVertexAiEndpoint`
- `GcpVertexAiIndex`
- `GcpVertexAiIndexEndpoint` -- Vector Search IndexEndpoint — distinct from the online-prediction GcpVertexAiEndpoint (671); different GCP resources, different kinds.
- `GcpVertexAiDeployedIndex`
- `GcpCloudComposerEnvironment`
- `GcpCloudComposerUserWorkloadsSecret`
- `GcpCloudComposerUserWorkloadsConfigMap`
- `GcpKmsKeyRing`
- `GcpKmsKey`
- `GcpKmsKeyIamMember`
- `GcpFilestoreInstance`
- `GcpWorkloadIdentityPool` -- 3101–3109: IAM/identity family (overflow block; the 3000–3022 foundation/security sub-band is fully allocated)
- `GcpWorkloadIdentityPoolProvider`
- `GcpServiceAccountIamMember`
- `GcpGcsBucketIamMember` -- One additive grant on a bucket, for a grantee that depends on the bucket itself (a logging sink writing into it): the bucket's own iam_members cannot reference such an identity without a cycle.
- `GcpGlobalForwardingRule` -- 3110–3119: networking/load-balancer family (overflow block; the 3023–3029 LB sub-band is fully allocated)
- `GcpSslPolicy`
- `GcpSslCertificate`
- `GcpServiceNetworkingConnection`
- `GcpAddress`
- `GcpServiceConnectionPolicy`
- `GcpCertManagerDnsAuthorization`
- `GcpCertificateMap` -- GcpCertManagerCert is a prerequisite because a map entry binds hostnames to EXISTING certificates — the canonical map references a certificate fixture's resource name.
- `GcpCertManagerTrustConfig` -- The CA certificates a load balancer validates client certificates against (mutual TLS). Its own kind: TLS policies reference it, never a certificate.
- `GcpCertManagerIssuanceConfig` -- How Google-managed certificates are issued from a private CA pool. Its own kind: many certificates share one config by name.
- `GcpCloudRunJob` -- 3120–3129: GCP serverless overflow
- `GcpServerlessVpcConnector`
- `GcpCloudRunWorkerPool` -- Cloud Run's no-ingress shape: a pool of always-running container instances (queue consumers, schedulers, background workers) that scales manually or by the owner's own signal instead of by requests. The proof deploys direct-VPC egress onto the prerequisite network.
- `GcpComputeDisk` -- 3130–3139: GCP compute overflow (the 3000–3022 foundation sub-band that holds GcpComputeInstance is fully allocated)
- `GcpComputeMig` -- GcpVpcNetwork is a prerequisite because the canonical group runs its fleet on a dedicated custom-mode VPC — a managed instance group's template must attach every VM to a network, and the default VPC is never assumed.
- `GcpMonitoringNotificationChannel` -- 3140–3149: GCP observability & log routing
- `GcpMonitoringAlertPolicy` -- GcpMonitoringNotificationChannel is a prerequisite because the policy's canonical shape references a channel to notify — a policy without a delivery endpoint measures but never pages.
- `GcpMonitoringUptimeCheck`
- `GcpLoggingSink` -- GcpGcsBucket is a prerequisite because the canonical sink exports to a Cloud Storage bucket — the cheapest destination that proves the whole writer-identity grant flow.
- `GcpMonitoringDashboard`
- `GcpMonitoringSlo`
- `GcpLogBucket`
- `GcpLogMetric`
- `GcpSecretManagerSecret` -- 3150–3159: GCP security & identity GcpServiceAccount is a prerequisite because the canonical secret grants secretAccessor to a workload service account — the access story the kind exists to model.
- `GcpIdentityPlatformConfig`
- `GcpIdentityPlatformTenant` -- GcpIdentityPlatformConfig is a prerequisite because tenants exist only in projects whose Identity Platform config enables multi_tenant.allow_tenants — a tenant without the initialized, tenant-enabled project config cannot be created at all.
- `GcpIamOauthClient`
- `GcpIamDenyPolicy`
- `GcpCloudRunDomainMapping` -- 3160–3169: GCP serverless edge GcpCloudRun is a prerequisite because a domain mapping exists only to point a verified domain at a running Cloud Run service — the route it maps must already exist for the mapping to be created at all.
- `GcpWorkflow`
- `GcpEventarcTrigger` -- GcpCloudRun is a prerequisite because the canonical trigger routes a Pub/Sub messagePublished event to a Cloud Run service — the destination story the kind exists to model.
- `GcpEventarcMessageBus`
- `GcpPlantonRunner`
- `GcpFolder` -- 3170–3179: GCP organization & governance (folders, org policies, tags, budgets, identity groups, API keys, KMS Autokey) GcpFolder is a container: the hierarchy node projects, sub-folders, policies, and tag bindings are placed inside.
- `GcpOrgPolicy`
- `GcpTagKey`
- `GcpTagValue`
- `GcpTagBinding`
- `GcpBillingBudget` -- A spending guardrail on a Cloud Billing account: amount, period, filters, thresholds, and where the alerts go. Lives on the billing account, so no project prerequisite; the proof lane needs a billing account the harness identity can administer.
- `GcpCloudIdentityGroup` -- A Google Group in Cloud Identity or Workspace with its memberships folded in -- the unit IAM bindings should name. Lives under a Cloud Identity customer, beside the service accounts and workload identity pools in the identity service group.
- `GcpApiKey`
- `GcpKmsAutokeyConfig` -- Cloud KMS Autokey switched on for a folder or a project: where the customer-managed keys GcpKmsKeyHandle requests are created. A security control, so it groups with the KMS kinds rather than the hierarchy.
- `GcpOrgPolicyCustomConstraint` -- A custom constraint is a DEFINITION the organization owns; the GcpOrgPolicy kinds that enforce it reference it by name, the way IAM bindings reference a custom role.
- `GcpSharedVpcHost` -- 3180–3189: GCP networking fabric (Shared VPC, VPC peering, HA VPN, firewall policies, PSC, network endpoint groups). HA VPN is two kinds: the gateway (with its Cloud Router) is declared once per VPC and region and referenced by every site connection, so two Google Cloud VPCs can point their connections at each other's gateway without a dependency cycle. Firewall policies are two kinds by scope: the hierarchical policy lives on the organization or a folder and is inherited by every network beneath its association; the network policy lives in a project and is attached to that project's VPC networks (globally or per region). Each folds its rules and associations -- a rule is keyed by priority inside its policy and an association is the edge that makes the policy act.
- `GcpSharedVpcServiceProject`
- `GcpVpcPeering`
- `GcpHaVpnGateway`
- `GcpHierarchicalFirewallPolicy`
- `GcpNetworkFirewallPolicy`
- `GcpHaVpnConnection`
- `GcpPscServiceAttachment` -- The producer half of Private Service Connect: publishes an internal load balancer's regional forwarding rule through NAT subnets so consumers in other VPCs reach it over a PSC endpoint (a regional GcpGlobalForwardingRule with an empty scheme targeting this attachment). The proof chain deploys the internal passthrough load balancer and the PSC NAT subnet as fixtures on the prerequisite network.
- `GcpNetworkEndpointGroup` -- Zonal (VM, hybrid, internet) and global (internet) network endpoint groups behind a `zone` selector; serverless, PSC, and regional internet groups stay in GcpRegionNetworkEndpointGroup. The zonal proof attaches a GcpComputeInstance fixture as an endpoint on the prerequisite network.
- `GcpRedisCluster` -- 3190–3199: GCP data (Memorystore for Redis Cluster, Managed Kafka, BigQuery connections and reservations, Datastream) Memorystore for Redis Cluster: the sharded, horizontally scaled Redis. Connectivity is Private Service Connect, placed by service connectivity automation through a GcpServiceConnectionPolicy for the gcp-memorystore-redis class on the network in the cluster's region -- the prerequisite; its proof pin carries that class beside the network and subnet pins.
- `GcpManagedKafkaCluster` -- Managed Service for Apache Kafka: a Google-operated broker fleet in one region, reachable from the VPC subnets it is attached to -- the prerequisite. Topics, ACLs, and Kafka Connect are their own kinds so the teams that own them declare them without editing the cluster.
- `GcpManagedKafkaTopic`
- `GcpManagedKafkaConnectCluster` -- Kafka Connect workers attached to a Kafka cluster: a separate project-and-location root that names the cluster it serves.
- `GcpBigQueryConnection` -- BigQuery's link to data outside its own storage (Cloud SQL, Spanner, AWS and Azure through Omni, Google resources through a managed service account, the Connector framework, Spark procedures).
- `GcpBigQueryReservation` -- A BigQuery slot reservation with the assignments that route projects, folders, or an organization onto it.
- `GcpDatastreamStream` -- A Datastream stream: continuous change data capture from one source database into BigQuery or Cloud Storage, through a source and a destination connection profile.
- `GcpRedisClusterEndpointSet` -- The Private Service Connect connections a consumer builds by hand (forwarding rules in other VPCs or projects) registered on a Redis Cluster, as one set: Google's resource replaces the cluster's whole user-created endpoint list in one write, so exactly one set per cluster is the honest grain. Its own kind because every connection names a forwarding rule that targets one of the cluster's service attachments -- a fold would depend on its own output.
- `GcpManagedKafkaAcl` -- The access rules for one resource pattern (a topic, a consumer group, a prefix, the cluster) on a Kafka cluster.
- `GcpManagedKafkaConnector` -- One data pipeline running on a Kafka Connect cluster.
- `GcpVertexAiAgentEngine` -- 3200–3229: GCP AI (Vertex AI agents and model deployments, RAG Engine, Vector Search, and the rest of the Vertex AI and generative-AI building blocks). The pre-existing Vertex AI kinds -- endpoint, index, index endpoint, deployed index, notebook -- live in the 3070s. Vertex AI Agent Engine: the managed runtime an AI agent runs in -- built from source or a container, hosted with its own identity and autoscaling, with an optional Memory Bank of long-term memories.
- `GcpVertexAiModelGardenDeployment` -- A Model Garden or Hugging Face model deployed to a Vertex AI endpoint in one step; every argument is immutable, so a change redeploys.
- `GcpVertexAiRagEngineConfig` -- The per-project, per-location tier of Vertex AI RAG Engine's managed vector database: a singleton Google owns, updated in place.
- `GcpVectorSearchCollection` -- 3203 is reserved for GcpVertexAiRagCorpus, forged when pulumi-gcp bridges provider 8.x (the resource has no SDK type today). A Vector Search collection -- a schema'd store of data objects with vector fields -- together with the approximate-nearest-neighbor indexes built over those fields (folded: one collection owns them).
- `GcpVertexAiSearchDataStore` -- Vertex AI Search (the Discovery Engine API behind the console's AI Applications / Gemini Enterprise): a data store is the corpus -- structured records, unstructured documents, or a public website -- with its schema, crawl patterns, and sitemaps folded in.
- `GcpVertexAiSearchEngine` -- The app over one or more data stores -- a search, chat, or recommendation engine -- with its serving controls, serving config, search widget, and assistants folded in. Its data stores are prerequisites: an engine cannot exist without one.
- `GcpVertexAiFeatureGroup` -- Vertex AI Feature Store: a feature group registers the features of some entities that live in a BigQuery table or view -- the features themselves folded in. Online stores serve them through feature views.
- `GcpVertexAiFeatureOnlineStore` -- The low-latency serving layer of Vertex AI Feature Store (Bigtable or Optimized storage), with the feature views it serves folded in.
- `GcpVertexAiDataset` -- A Vertex AI managed dataset: the registered container training, AutoML, labeling, and evaluation read their examples from.
- `GcpVertexAiTensorboard` -- A managed Vertex AI TensorBoard training jobs stream metrics into, with the experiments and runs declared in it folded in.
- `GcpVertexAiPersistentResource` -- A long-running cluster Vertex AI keeps provisioned so training jobs and Ray on Vertex AI start in seconds and scarce accelerators stay held between jobs.
- `GcpModelArmorTemplate` -- A Model Armor template: the named safety filters (prompt injection and jailbreak, Responsible AI content, sensitive data, malicious URLs) an AI application screens prompts and responses through.
- `GcpDocumentAiProcessor` -- A Document AI processor: a managed model that turns documents into structured data (OCR, forms, invoices, IDs), with its default version.
- `GcpColabRuntimeTemplate` -- A Colab Enterprise runtime template: the machine, network, image, and security settings every notebook runtime created from it gets.
- `GcpColabRuntime` -- A Colab Enterprise runtime: a notebook VM assigned to one user, built from a runtime template and started or stopped on purpose.
- `GcpColabSchedule` -- A Vertex AI schedule: a cron that launches a Colab Enterprise notebook run or a Vertex AI Pipelines run.
- `GcpTpuVm` -- A Cloud TPU VM: a slice of Google's AI accelerators with its host VMs. Beta-only in Google's provider (a recorded google-beta admission).
- `GcpDialogflowCxAgent` -- A Dialogflow CX conversational agent with the infrastructure its console-authored content uses folded in: webhooks, tools and their frozen versions, flow versions, the environments that pin them, and generative settings per language.
- `GcpVertexAiSearchDataConnector` -- A data connector is a COLLECTION of data stores Google syncs from a source (Jira, Confluence, ServiceNow, SharePoint, BigQuery, Google Drive, ...) on a schedule -- a different root from a data store, which is why it is its own kind. Engines search its stores by naming the collection.
- `GcpModelArmorFloorSetting` -- A Model Armor floor setting: the minimum safety screening a project, folder, or organization enforces on its templates and directly on Vertex AI and Google MCP server traffic. A different parent from a template (and a singleton Google never deletes), which is why it is its own kind.
- `GcpTpuQueuedResource` -- A Cloud TPU queued resource: a request that waits for TPU capacity and then provisions the nodes it describes. A different root from a TPU VM that owns many nodes, which is why it is its own kind. Beta-only in Google's provider (a recorded google-beta admission).
- `GcpDialogflowCxSecuritySettings` -- Dialogflow CX security settings: the redaction, retention, audio-export, and Insights-export policy agents in one project and location apply to their conversations. A different root from an agent, referenced by agents and shared among them, which is why it is its own kind.
- `GcpPrivateCaPool` -- 3230–3239: GCP security (Certificate Authority Service, Cloud KMS Autokey handles, Security Command Center, Binary Authorization) A Certificate Authority Service CA pool: the trust anchor and issuance policy its certificate authorities and certificates live inside.
- `GcpKmsKeyHandle` -- An Autokey key handle: asks Autokey for a customer-managed key for one resource type in one project and location; the resource it protects names the key the handle returns.
- `GcpSccNotificationConfig` -- Security Command Center streaming notifications to Pub/Sub at a project, folder, or organization.
- `GcpSccMuteConfig` -- A Security Command Center mute rule at a project, folder, or organization.
- `GcpSccBigQueryExport` -- A continuous Security Command Center findings export to a BigQuery dataset at a project, folder, or organization.
- `GcpBinaryAuthorizationPolicy` -- A project's Binary Authorization policy: which container images GKE admits, per cluster. A project singleton.
- `GcpBinaryAuthorizationAttestor` -- A Binary Authorization attestor with its Artifact Analysis note: the public keys that verify image attestations a policy requires.
- `GcpPrivateCaCertificateAuthority` -- A certificate authority in a CA pool: a self-signed root, or a subordinate signed by another authority or an outside CA. Its own kind because a pool rotates through several and a subordinate references its parent.
- `GcpPrivateCaCertificateTemplate` -- A certificate template: a reusable certificate shape in a project and location that certificates in any pool there reference.
- `GcpPrivateCaCertificate` -- A certificate issued from a CA pool for a key its owner holds; destroy revokes it.
- `GcpGkeFleet` -- 3240–3249: GCP platform engineering (GKE fleets, images, Cloud Build, Cloud Deploy) GcpGkeFleet is the container a fleet's scopes, namespaces, memberships, and features live in: the project's one fleet is the room, and a team scope or a fleet-wide feature is what is placed inside it.
- `GcpGkeFleetFeature` -- The three fleet children name GcpGkeFleet as their prerequisite: Google requires the fleet before a scope, a fleet declared after a cluster registers collides with the fleet that registration created implicitly, and a feature configures the fleet it lives in. A chart that references the fleet's project_id output orders each child after it.
- `GcpGkeFleetScope`
- `GcpGkeFleetMembership`
- `GcpComputeImage` -- A Compute Engine custom image: the golden boot image VMs, instance templates, and disks start from, rolled forward through image families.
- `GcpCloudBuildConnection` -- GcpCloudBuildConnection is the container a code host's repositories live in: the connection to GitHub, GitLab, or Bitbucket is the room, and each linked repository is what is placed inside it.
- `GcpCloudBuildRepository` -- A repository is created under its connection and cannot exist without it.
- `GcpCloudBuildTrigger` -- A Cloud Build trigger: what starts a build (a code event, a Pub/Sub message, a webhook, or a manual run) and what the build does.
- `GcpDeliveryPipeline` -- A Cloud Deploy delivery pipeline: the ordered stages a release is promoted through, with the automations that drive it.
- `GcpDeployTarget` -- A Cloud Deploy target: where a pipeline stage deploys (a GKE cluster, a Cloud Run location, a fleet cluster, several targets at once, or a custom target).
- `GcpFirebaseProject` -- 3250–3259: GCP Firebase (project enablement, app registrations, and the Firebase-adjacent products that follow) GcpFirebaseProject is the container the app registrations live in: "Firebase on this project" is the room, the Android/Apple/Web apps are what is placed inside it.
- `GcpFirebaseAndroidApp` -- The three app registrations exist only inside a Firebase-enabled project, so each names GcpFirebaseProject as its prerequisite: the E2E harness deploys the enablement first, and a chart that references the enablement's project_id output orders the registration after it.
- `GcpFirebaseAppleApp`
- `GcpFirebaseWebApp`
- `GcpCloudBuildWorkerPool` -- 3260–3269: GCP platform engineering, continued (Cloud Build and Cloud Deploy resources many pipelines share) A private Cloud Build worker pool: build machines many triggers and Cloud Deploy targets share, optionally on a private network.
- `GcpDeployPolicy` -- A Cloud Deploy deploy policy: rollout restrictions (freeze windows) that apply to every pipeline and target its selectors match.
- `GcpDeployCustomTargetType` -- A Cloud Deploy custom target type: how to render and deploy to a target Google does not deploy natively; many targets share one.
- `KubernetesNamespace` -- 4000–4999: Kubernetes resources, organized in family sub-bands (4030–4069 also hosts CNI/autoscaling/DR addons; 4130–4149 hosts analytics & ML; 4180–4199 hosts agent sandbox runtimes; 4200–4209 hosts AI agent platforms) 4000–4029: Kubernetes building blocks (core API primitives)
- `KubernetesDeployment`
- `KubernetesStatefulSet`
- `KubernetesDaemonSet`
- `KubernetesJob`
- `KubernetesCronJob`
- `KubernetesService`
- `KubernetesSecret`
- `KubernetesManifest`
- `KubernetesHelmRelease`
- `KubernetesConfigMap`
- `KubernetesServiceAccount`
- `KubernetesRbac` -- Bundles the RBAC grant grain (Role/ClusterRole + its binding) into one kind: "grant these permissions to these subjects in this scope".
- `KubernetesIngress`
- `KubernetesNetworkPolicy`
- `KubernetesPersistentVolumeClaim`
- `KubernetesStorageClass`
- `KubernetesResourceQuota` -- Manages the namespace-governance pair: the ResourceQuota plus an optional companion LimitRange (per-object defaults/bounds) — two API objects, one governance story.
- `KubernetesPriorityClass`
- `KubernetesPodDisruptionBudget`
- `KubernetesHorizontalPodAutoscaler`
- `KubernetesCertManager` -- 4030–4069: Kubernetes foundation addons (certs, DNS, secrets, ingress, Gateway API, mesh, CNI/autoscaling/DR)
- `KubernetesClusterIssuer` -- KubernetesCertManager is a prerequisite for the three cert-manager CR kinds below: ClusterIssuer/Issuer/Certificate are cert-manager custom resources — without the controller and its CRDs they cannot be applied.
- `KubernetesIssuer`
- `KubernetesCertificate`
- `KubernetesExternalDns`
- `KubernetesExternalSecretsOperator`
- `KubernetesClusterSecretStore` -- KubernetesExternalSecretsOperator is a prerequisite for the three external-secrets CR kinds below: ClusterSecretStore/SecretStore/ ExternalSecret are external-secrets custom resources — without the operator and its CRDs they cannot be applied.
- `KubernetesSecretStore`
- `KubernetesExternalSecret`
- `KubernetesIngressNginx`
- `KubernetesGatewayApiCrds`
- `KubernetesGatewayClass`
- `KubernetesGateway`
- `KubernetesListenerSet`
- `KubernetesHttpRoute`
- `KubernetesGrpcRoute`
- `KubernetesTcpRoute`
- `KubernetesUdpRoute`
- `KubernetesTlsRoute`
- `KubernetesReferenceGrant`
- `KubernetesBackendTlsPolicy`
- `KubernetesIstioBaseCrds`
- `KubernetesIstio`
- `KubernetesDestinationRule` -- Istio API kinds (mesh traffic policy, security, telemetry). The seven typed resources below (4053–4059) require the Istio CRDs on the cluster, provided by the lightweight CRDs-only KubernetesIstioBaseCrds (851) — NOT the full mesh KubernetesIstio (852).
- `KubernetesServiceEntry`
- `KubernetesPeerAuthentication`
- `KubernetesRequestAuthentication`
- `KubernetesAuthorizationPolicy`
- `KubernetesTelemetry`
- `KubernetesEnvoyFilter`
- `KubernetesMetricsServer`
- `KubernetesCilium`
- `KubernetesKeda`
- `KubernetesKarpenter`
- `KubernetesKarpenterNodePool`
- `KubernetesKarpenterEc2NodeClass`
- `KubernetesClusterAutoscaler`
- `KubernetesVelero`
- `KubernetesGkeComputeClass`
- `KubernetesKubePrometheusStack` -- 4070–4079: Kubernetes observability
- `KubernetesGrafana`
- `KubernetesSignoz` -- KubernetesClickHouse is a prerequisite because SigNoz stores every trace, metric and log in ClickHouse and deploys none of its own — the telemetry store is composed, never bundled.
- `KubernetesLoki`
- `KubernetesTempo`
- `KubernetesOtelOperator` -- The operator's admission webhooks (failurePolicy Fail) are served with a cert-manager Certificate in the default posture — cert-manager must be running before the operator installs.
- `KubernetesOtelCollector`
- `KubernetesServiceMonitor` -- KubernetesKubePrometheusStack is a prerequisite because it installs the monitoring.coreos.com CRDs the ServiceMonitor is an instance of, and the Prometheus that scrapes through it.
- `KubernetesPodMonitor` -- KubernetesKubePrometheusStack is a prerequisite because it installs the monitoring.coreos.com CRDs the PodMonitor is an instance of, and the Prometheus that scrapes through it.
- `KubernetesPrometheusRule` -- KubernetesKubePrometheusStack is a prerequisite because it installs the monitoring.coreos.com CRDs the PrometheusRule is an instance of, and the Prometheus that evaluates it.
- `KubernetesKyverno` -- 4080–4099: Kubernetes security, policy, and identity
- `KubernetesGatekeeper`
- `KubernetesKeycloak` -- Keycloak declarations compose the official Keycloak Operator (which reconciles the Keycloak CR this kind renders) and, on the recommended postgres vendor, a KubernetesPostgres database — both must resolve before the CR can converge.
- `KubernetesOpenBao`
- `KubernetesOpenFga` -- OpenFGA requires a datastore; the recommended arm composes a KubernetesPostgres database (the sandbox memory arm needs nothing, but the registry declares the shape real deployments require).
- `KubernetesKeycloakOperator`
- `KubernetesCloudNativePgOperator` -- 4100–4129: Kubernetes data platforms
- `KubernetesPostgres`
- `KubernetesValkey`
- `KubernetesPerconaMysqlOperator`
- `KubernetesMysql`
- `KubernetesPerconaMongoOperator`
- `KubernetesMongodb`
- `KubernetesStrimziKafkaOperator`
- `KubernetesKafka` -- container_kind: a Strimzi Kafka cluster is a place in the provider's own model — KafkaTopic and KafkaUser declarations BELONG to one cluster (the strimzi.io/cluster label) and are drawn inside its box. Clients that merely talk to the cluster (Connect, MirrorMaker2, UI, Karapace) carry containment_exempt on their bootstrap/trust references.
- `KubernetesKafkaTopic`
- `KubernetesKafkaUser`
- `KubernetesKafkaConnect` -- container_kind: a Connect cluster hosts the connectors deployed INTO it (KafkaConnector's strimzi.io/cluster label names its Connect cluster) — the same room shape as KubernetesKafka above.
- `KubernetesKafkaConnector`
- `KubernetesKafkaMirrorMaker2`
- `KubernetesKarapace`
- `KubernetesKafkaUi`
- `KubernetesOpenSearchOperator`
- `KubernetesOpenSearch`
- `KubernetesAltinityOperator`
- `KubernetesClickHouse`
- `KubernetesSolrOperator`
- `KubernetesSolr`
- `KubernetesNeo4j`
- `KubernetesSeaweedFs`
- `KubernetesQdrant`
- `KubernetesRabbitMqOperator` -- The RabbitMQ Cluster Operator's release manifest ships admission webhooks whose serving certificate is a cert-manager Certificate — cert-manager must be running before the operator installs.
- `KubernetesRabbitMq`
- `KubernetesCnpgBarmanCloudPlugin` -- The Barman Cloud plugin is CloudNativePG's object-store backup path, installed as its own Helm release beside the operator. It registers with the operator over CNPG-I and issues its TLS through cert-manager, so both are prerequisites: nothing to register with, or no certificates, and the release never becomes ready.
- `KubernetesAirflow` -- 4130–4149: Kubernetes analytics and ML KubernetesPostgres is a prerequisite because Airflow's metadata database composes a KubernetesPostgres by default (the spec's FK defaults resolve onto its outputs) and the migration Job needs the database reachable before the server components start.
- `KubernetesSparkOperator`
- `KubernetesKubeRayOperator`
- `KubernetesRayCluster` -- KubernetesKubeRayOperator is a prerequisite because this kind declares the RayCluster custom resource that only the operator's CRDs admit and only the operator reconciles into head and worker pods.
- `KubernetesFlinkOperator` -- KubernetesCertManager is a prerequisite because the Flink operator's chart, with its default-on admission webhook, renders cert-manager Issuer/Certificate resources and trusts the API server through cert-manager's CA injection — there is no self-signed fallback at the pinned chart, and the webhooks are fail-closed.
- `KubernetesFlinkDeployment` -- KubernetesFlinkOperator is a prerequisite because this kind declares the FlinkDeployment custom resource that only the operator's CRDs admit and only the operator reconciles into a running Flink cluster.
- `KubernetesJupyterHub` -- KubernetesPostgres is a prerequisite because JupyterHub's hub database composes a KubernetesPostgres in its external-database arm (the spec's FK defaults resolve onto its outputs) and the hub pod mounts that database's credential Secret before it can start.
- `KubernetesMlflow` -- KubernetesPostgres is a prerequisite because MLflow's backend store composes a KubernetesPostgres in its production arm (FK defaults onto its outputs; the module composes the connection URI from its credential Secret), and KubernetesSeaweedFs because the artifact store's S3-compatible arm FK-defaults onto the SeaweedFS endpoint and credential Secret.
- `KubernetesTrino` -- KubernetesPostgres is a prerequisite because Trino's postgres catalogs compose a KubernetesPostgres (the catalog host and credential FK-default onto its outputs), and the pods read that database's credential Secret to resolve catalog passwords from environment.
- `KubernetesSuperset` -- KubernetesPostgres is a prerequisite because Superset's REQUIRED metadata database composes a KubernetesPostgres (FK defaults onto its outputs; the module composes the environment Secret from its credential Secret), and KubernetesValkey because the cache/broker arm FK-defaults onto a KubernetesValkey's service and password Secret.
- `KubernetesLangfuse` -- Langfuse's four required stores each FK-default onto a catalog kind: KubernetesPostgres (users, projects, prompts, API keys - host and password Secret), KubernetesClickHouse (traces and scores - host, cluster name and auth Secret), KubernetesValkey (the ingestion queues - host, ACL user and password Secret), and KubernetesSeaweedFs (the S3-compatible event store - endpoint and credentials Secret).
- `KubernetesArgocd` -- 4150–4169: Kubernetes GitOps and CI/CD
- `KubernetesArgoWorkflows`
- `KubernetesTektonOperator`
- `KubernetesTekton` -- KubernetesTektonOperator is a prerequisite because this kind declares the TektonConfig custom resource that only the operator's CRDs admit and only the operator reconciles into running components.
- `KubernetesGhaRunnerScaleSetController`
- `KubernetesGhaRunnerScaleSet` -- KubernetesGhaRunnerScaleSetController is a prerequisite because this kind renders an AutoscalingRunnerSet custom resource that only the controller's CRDs admit and only the controller reconciles into listener and runner pods.
- `KubernetesHarbor`
- `KubernetesJenkins`
- `KubernetesTemporal` -- 4170–4179: Kubernetes app platforms KubernetesPostgres is a prerequisite because the recommended (and E2E-proven) database composition backs Temporal's default and visibility stores with a CloudNativePG cluster.
- `KubernetesNats`
- `KubernetesLocust`
- `KubernetesPlantonRunner`
- `KubernetesPlantonOperator`
- `KubernetesPlantonPlatform` -- KubernetesPlantonOperator is a prerequisite because this kind declares the PlantonPlatform custom resource that only the operator's CRD admits and only the operator reconciles into a running platform.
- `KubernetesGoFeatureFlag`
- `KubernetesFlagd`
- `KubernetesGoFeatureFlagFlagFile`
- `KubernetesFlagdFlagFile`
- `KubernetesAgentSandboxController` -- 4180–4199: agent sandbox runtimes (4180–4189 Kubernetes SIG Apps' agent-sandbox; 4190–4199 Agent Substrate)
- `KubernetesAgentSandbox` -- KubernetesAgentSandboxController is a prerequisite because it installs the agents.x-k8s.io definition the Sandbox is an instance of, and the controller that runs its pod.
- `KubernetesAgentSandboxTemplate` -- KubernetesAgentSandboxController is a prerequisite because it installs the extensions.agents.x-k8s.io definition the SandboxTemplate is an instance of, and the extension controllers that turn it into network policy and warm pool members.
- `KubernetesAgentSandboxWarmPool` -- KubernetesAgentSandboxController is a prerequisite because it installs the extensions.agents.x-k8s.io definition the SandboxWarmPool is an instance of, and the controller that keeps the pool's members ready.
- `KubernetesAgentSandboxClaim` -- KubernetesAgentSandboxController is a prerequisite because it installs the extensions.agents.x-k8s.io definition the SandboxClaim is an instance of, and the controller that binds it to a warm pool member.
- `KubernetesAgentSubstrate`
- `KubernetesAgentSubstrateWorkerPool` -- KubernetesAgentSubstrate is a prerequisite because it installs the ate.dev definition the WorkerPool is an instance of, the controller that turns it into worker pods, and the node agent and API server the workers register with.
- `KubernetesAgentSubstrateSandboxConfig` -- KubernetesAgentSubstrate is a prerequisite because it installs the ate.dev definition the SandboxConfig is an instance of, the admission policy that holds each class to its assets, and the API server and node agent that read it.
- `KubernetesAgentSubstrateCsiDriverConfig` -- KubernetesAgentSubstrate is a prerequisite because it installs the ate.dev definition the CSIDriverConfig is an instance of, and the API server and node agent that look a driver up in it.
- `KubernetesStigmer` -- 4200–4209: Kubernetes AI agent platforms KubernetesStigmer installs the Stigmer operator and declares the one Stigmer it reconciles, so it carries no prerequisite.
- `DigitalOceanApp` -- 5000–5999: DigitalOcean resources
- `DigitalOceanBucket`
- `DigitalOceanContainerRegistry`
- `DigitalOceanDatabaseCluster` -- A DigitalOcean managed database cluster -- the cluster every other managed-database resource is created on: users, logical databases, connection pools, the firewall rule set, Kafka topics, and Kafka schema subjects are all API-addressed under a cluster's id and cannot exist before it. A container kind for exactly that reason: those six kinds are children of the cluster, so on a diagram the cluster is the room they stand in. A read replica also names its primary but is a cluster of its own, a monitor alert names the clusters it watches, and an App Platform app names a cluster it attaches; those three references are containment-exempt on their specs.
- `DigitalOceanDnsZone`
- `DigitalOceanDroplet` -- No VPC prerequisite: the droplet spec's vpc reference is optional — an omitted vpc places the droplet in the region's default VPC.
- `DigitalOceanFirewall`
- `DigitalOceanFunction`
- `DigitalOceanKubernetesCluster` -- DigitalOceanVpc is a prerequisite because the cluster spec's vpc reference is required: the control plane and every node pool live inside one VPC, resolved to the DigitalOceanVpc's exported vpc_id output.
- `DigitalOceanKubernetesNodePool` -- DigitalOceanKubernetesCluster is a prerequisite because a node pool is API-addressed under its owning cluster: the spec's cluster reference is required and the pool cannot exist first.
- `DigitalOceanLoadBalancer` -- No registry prerequisite: the load balancer's vpc reference is optional (DigitalOcean places it in the region's default VPC when unset, and GLOBAL balancers take no VPC at all). Scenarios that exercise VPC placement declare it per-scenario via the e2e-prerequisites annotation.
- `DigitalOceanVolume`
- `DigitalOceanVpc`
- `DigitalOceanCertificate`
- `DigitalOceanDnsRecord` -- DigitalOceanDnsZone is a prerequisite because a record is API-addressed under its domain: the spec's domain reference is required, resolved to the DigitalOceanDnsZone's exported zone_name output.
- `DigitalOceanDatabaseUser` -- An additional user on a managed database cluster: the cluster reference is required, resolved to the DigitalOceanDatabaseCluster's exported cluster_id output.
- `DigitalOceanDatabaseDb` -- An additional logical database on a managed database cluster: the cluster reference is required.
- `DigitalOceanDatabaseConnectionPool` -- A PgBouncer connection pool on a managed PostgreSQL cluster: the cluster reference is required.
- `DigitalOceanDatabaseFirewall` -- The inbound trusted-sources rule set of a managed database cluster: the cluster reference is required.
- `DigitalOceanDatabaseReplica` -- A read-only replica of a managed database cluster: the primary cluster reference is required.
- `DigitalOceanDatabaseKafkaTopic` -- A topic on a managed Kafka cluster with the full per-topic configuration block; the owning cluster reference is required.
- `DigitalOceanDatabaseKafkaSchema` -- One schema subject registered in a managed Kafka cluster's schema registry; the owning cluster reference is required.
- `DigitalOceanProject` -- The account-level organizational container; membership is carried on the project itself as resource URNs.
- `DigitalOceanSshKey` -- An SSH public key registered on the account, referenced by droplets and droplet autoscale pools at create time.
- `DigitalOceanMonitorAlert` -- An alert policy on DigitalOcean's built-in metrics for droplets, load balancers, and managed database clusters. All entity targeting is optional, so there is no registry prerequisite.
- `DigitalOceanUptimeCheck` -- An availability/latency probe on an external endpoint with composed alert rules; the target is outside the account, so there is no registry prerequisite.
- `DigitalOceanReservedIp` -- A static public IP address (IPv4 or IPv6) reserved in a region and optionally assigned to a droplet. The droplet attachment is an optional composition seam, so there is no registry prerequisite.
- `DigitalOceanVpcPeering` -- A private-network peering connection between exactly two VPCs; both VPC references are required.
- `DigitalOceanSpacesKey` -- An access-key pair for Spaces object storage. Bucket grants are an optional composition seam, so there is no registry prerequisite.
- `DigitalOceanCdn` -- A CDN endpoint serving a Spaces bucket's content from the global edge: the origin reference is required, resolved to the DigitalOceanBucket's exported bucket_domain_name output.
- `DigitalOceanDropletAutoscalePool` -- A pool of identical droplets DigitalOcean keeps at a fixed size or scales on utilization. The template's ssh_keys reference is required (the API mandates SSH keys), resolved to the DigitalOceanSshKey's exported ssh_key_id output.
- `CloudflareDnsZone` -- 7000–7999: Cloudflare resources
- `CloudflareKvNamespace`
- `CloudflareR2Bucket`
- `CloudflareWorker`
- `CloudflareLoadBalancer` -- CloudflareDnsZone and CloudflareLoadBalancerPool are prerequisites because a load balancer is a DNS-level construct inside a zone (the spec's zone_id reference must resolve) and traffic must land somewhere (the required fallback_pool reference must resolve to a live pool).
- `CloudflareD1Database`
- `CloudflareZeroTrustAccessApplication`
- `CloudflareDnsRecord` -- CloudflareDnsZone is a prerequisite because every record lives inside a zone -- the spec's zone_id reference must resolve before the record can be created.
- `CloudflareRuleset` -- CloudflareDnsZone is a prerequisite because zone-scoped rulesets (the common case; the spec's zone_id reference defaults to the zone kind) must resolve their zone first. Account-scoped rulesets simply leave the reference unused.
- `CloudflareWorkersKvPair` -- CloudflareKvNamespace is a prerequisite because a KV pair is written into a namespace -- the spec's namespace_id reference must resolve first.
- `CloudflareHyperdriveConfig`
- `CloudflareLoadBalancerPool`
- `CloudflareLoadBalancerMonitor`
- `CloudflareZeroTrustAccessPolicy`
- `CloudflareZeroTrustAccessGroup`
- `CloudflareQueue`
- `CloudflarePagesProject`
- `CloudflareZeroTrustTunnel`
- `CloudflareZeroTrustTunnelVirtualNetwork`
- `CloudflareZeroTrustTunnelRoute` -- CloudflareZeroTrustTunnel is a prerequisite because a route steers a CIDR through an existing tunnel -- the spec's tunnel_id reference must resolve first. (The optional virtual_network_id is scenario-declared, not a registry prerequisite.)
- `CloudflareList`
- `CloudflareListItem` -- CloudflareList is a prerequisite because an item exists only inside a list -- the spec's list_id reference must resolve first.
- `CloudflareTurnstileWidget`
- `CloudflareEmailRoutingZone` -- CloudflareDnsZone is a prerequisite because email routing is enabled ON a zone -- the spec's zone_id reference must resolve first.
- `CloudflareEmailRoutingRule` -- CloudflareDnsZone is a prerequisite because a routing rule lives in a zone's email routing configuration -- the spec's zone_id reference must resolve first. CloudflareEmailRoutingZone is deliberately NOT a prerequisite: the API accepts rule creation on a zone whose Email Routing was never enabled (measured live 2026-08-26 -- POST zones/{id}/email/routing/rules answered 200 on a fresh zone with routing status "unconfigured"). Enablement gates mail FLOW, not rule lifecycle, so it is an operational concern the kind docs teach, not a create-time dependency. (Forward destinations reference CloudflareEmailRoutingAddress only for forward-type rules, so that edge is scenario-declared.)
- `CloudflareEmailRoutingAddress`
- `CloudflareOriginCaCertificate`
- `CloudflareCertificatePack` -- CloudflareDnsZone is a prerequisite because a certificate pack is ordered for a zone's hostnames -- the spec's zone_id reference must resolve first.
- `CloudflareCustomHostname` -- CloudflareDnsZone is a prerequisite because a custom hostname (SSL for SaaS) is provisioned inside a zone -- the spec's zone_id reference must resolve first.
- `CloudflareCustomHostnameFallbackOrigin` -- CloudflareDnsZone is a prerequisite because the fallback origin is a zone-level SSL-for-SaaS setting -- the spec's zone_id reference must resolve first.
- `CloudflareZeroTrustAccessIdentityProvider` -- No prerequisites: identity providers are account-scoped in the canonical case (the optional zone scope is a per-manifest choice, not a structural dependency).
- `CloudflareZeroTrustAccessServiceToken` -- No prerequisites: service tokens are account-scoped in the canonical case (the optional zone scope is a per-manifest choice, not a structural dependency).
- `CloudflareZeroTrustOrganization` -- No prerequisites: the organization is an account-scoped configuration singleton (the optional zone scope is a per-manifest choice, not a structural dependency).
- `CloudflareZeroTrustAccessInfrastructureTarget` -- No prerequisites: targets are account-scoped, and the virtual-network reference is an optional per-manifest edge (omitted = the account's default virtual network).
- `CloudflareZeroTrustMcpPortal` -- No prerequisites: portals are account-scoped, and the servers[] rows' MCP-server references are optional per-manifest edges.
- `CloudflareZeroTrustMcpServer` -- No prerequisites: MCP server registrations are account-scoped and self-contained -- portals reference them, not the reverse.
- `CloudflareZeroTrustGatewayPolicy` -- No prerequisites: Gateway policies are account-scoped, and their list / virtual-network references are optional per-manifest edges.
- `CloudflareZeroTrustList` -- No prerequisites: Zero Trust lists are account-scoped and self-contained.
- `CloudflareZeroTrustGatewaySettings` -- No prerequisites: the Gateway configuration is an account-scoped singleton, and its certificate reference is an optional per-manifest edge.
- `CloudflareZeroTrustDnsLocation` -- No prerequisites: DNS locations are account-scoped and self-contained.
- `CloudflareZeroTrustDeviceDefaultProfile` -- No prerequisites: the default device profile is an account-scoped configuration singleton; its virtual-network and zone-certificate references are optional per-manifest edges.
- `CloudflareZeroTrustDeviceCustomProfile` -- No prerequisites: custom device profiles are account-scoped, and the virtual-network reference is an optional per-manifest edge.
- `CloudflareZeroTrustDevicePostureRule` -- No prerequisites: posture rules are account-scoped and self-contained (list and integration references are literal UUIDs today).
- `CloudflareZoneTlsSettings` -- CloudflareDnsZone is a prerequisite because TLS settings are zone-scoped configuration -- the spec's zone_id reference must resolve first.
- `CloudflareCustomSslCertificate` -- CloudflareDnsZone is a prerequisite because a custom certificate is uploaded to an existing zone -- the spec's zone_id reference must resolve first.
- `CloudflareMtlsCertificate` -- No prerequisites: mTLS certificates are account-scoped uploads and self-contained -- consumers (zone TLS CA associations, Authenticated Origin Pulls rows, Workers mTLS bindings) reference them, not the reverse.
- `CloudflareAuthenticatedOriginPulls` -- CloudflareDnsZone is a prerequisite because Authenticated Origin Pulls enablement configures an existing zone -- the spec's zone_id reference must resolve first. The per-hostname certificate edge is optional and scenario-declared, never a registry prerequisite.
- `CloudflareAuthenticatedOriginPullsCertificate` -- CloudflareDnsZone is a prerequisite because the client certificate is uploaded to an existing zone -- the spec's zone_id reference must resolve first.
- `CloudflareZoneSettings` -- CloudflareDnsZone is a prerequisite because zone settings configure an existing zone -- the spec's zone_id reference must resolve first.
- `CloudflareCacheSettings` -- CloudflareDnsZone is a prerequisite because cache settings configure an existing zone -- the spec's zone_id reference must resolve first.
- `CloudflareIpAccessRule` -- No prerequisites: IP Access rules are account-scoped in the canonical case (the zone scope is a per-manifest choice, not a structural dependency).
- `CloudflareBotManagement` -- CloudflareDnsZone is a prerequisite because Bot Management is zone-singleton configuration -- the spec's zone_id reference must resolve first.
- `CloudflareSnippet` -- CloudflareDnsZone is a prerequisite because snippets deploy to a zone -- the spec's zone_id reference must resolve first.
- `CloudflareSnippetRules` -- CloudflareDnsZone and CloudflareSnippet are prerequisites: the rules table is zone-scoped and every rule invokes a snippet by name.
- `CloudflareWaitingRoom` -- CloudflareDnsZone is a prerequisite because waiting rooms sit on a zone's host+path -- the spec's zone_id reference must resolve first.
- `CloudflareWaitingRoomEvent` -- CloudflareWaitingRoom is a prerequisite because events run on a room (and the room's own chain brings the zone).
- `CloudflareLogpushJob` -- No prerequisites: logpush jobs are dual-scope (account or zone) and the zone reference is optional -- zone-scoped lanes declare the edge at the scenario level.
- `CloudflareNotificationPolicy` -- No prerequisites: every delivery mechanism (email, PagerDuty, webhook) is optional -- policies referencing a webhook declare the edge at the scenario level.
- `CloudflareNotificationWebhook` -- No prerequisites: a webhook destination is account-scoped and self-contained -- notification policies reference it, not the reverse.
- `CloudflareWebAnalyticsSite` -- No prerequisites: a site is identified by host OR zone, and the zone reference is optional -- zone-measured lanes declare the edge at the scenario level.
- `CloudflareWorkflow` -- CloudflareWorker is a prerequisite because a workflow registers a class exported by a DEPLOYED Worker script -- the spec's script_name reference must resolve first.
- `CloudflareSecretsStore` -- No prerequisites: the Secrets Store is an account-scoped container and self-contained -- consumers (store secrets, Worker bindings, AI Gateway authentication) reference it, not the reverse.
- `CloudflareSecretsStoreSecret` -- CloudflareSecretsStore is a prerequisite because every secret lives inside a store -- the spec's store_id reference must resolve first.
- `CloudflareAiGateway` -- No prerequisites: the gateway is account-scoped and self-contained; its optional Secrets Store link (BYO provider keys) is a scenario-level composition, not a structural requirement.
- `CloudflareAccountApiToken` -- No prerequisites: an account-owned API token is self-contained; the resources its policies cover are identifiers, not references.
- `CloudflareHealthcheck` -- CloudflareDnsZone is a prerequisite because standalone health checks are zone-scoped -- the spec's zone_id reference must resolve first.
- `Auth0Connection` -- 8000–8999: Auth0 resources
- `Auth0Client`
- `Auth0EventStream`
- `Auth0ResourceServer`
- `Auth0Action`
- `Auth0Role`
- `Auth0User`
- `Auth0TenantSettings`
- `Auth0CustomDomain`
- `Auth0CustomDomainVerification`
- `Auth0Branding` -- 8010–8015: how the tenant's Universal Login looks, reads and mails -- branding and theme, the login flow, each prompt's words and screen partials, the email service and each email.
- `Auth0Prompt`
- `Auth0PromptCustomText`
- `Auth0PromptScreenPartials`
- `Auth0EmailProvider`
- `Auth0EmailTemplate`
- `Auth0ClientFromMetadataDocument`
- `OpenFgaStore` -- 9000–9999: OpenFGA resources OpenFGA publishes a Terraform provider and no Pulumi provider, so its kinds ship one HCL module and declare the engines that run it; the CLI and the platform refuse Pulumi for them before anything runs.
- `OpenFgaAuthorizationModel`
- `OpenFgaRelationshipTuple`
- `StripeWebhookEndpoint` -- 10000–10999: Stripe resources Stripe publishes a Terraform provider and no Pulumi provider, and its kinds are proven on OpenTofu alone, so each declares ["tofu"]: one HCL module, and every other engine refused before anything runs. Family sub-bands: 10000 account and delivery, 10100 catalog and pricing.
- `StripeEventDestination`
- `StripeBillingPortalConfiguration`
- `StripePaymentMethodConfiguration`
- `StripePaymentMethodDomain`
- `StripeRadarValueList`
- `StripeProduct`
- `StripePrice` -- A price always belongs to a product (its product field is required), so the product deploys first.
- `StripeCoupon`
- `StripePromotionCode` -- A promotion code always redeems a coupon (its coupon field is required), so the coupon deploys first.
- `StripeEntitlementFeature`
- `StripeShippingRate`
- `StripeTaxRate` -- Tax rate and tax registration share the initials "tr", so each takes Stripe's own id prefix with its vowels dropped (txr_, taxreg_).
- `StripeTaxRegistration`
- `StripeBillingMeter`
- `StripePaymentLink` -- A payment link always sells at least one price (line_items is required), so the price deploys first.

### spec.server.extraEnv.variables[].valueFrom.env

`string`

### spec.server.extraEnv.variables[].valueFrom.name

`string` · required

- rule: {"required":true}

### spec.server.extraEnv.variables[].valueFrom.fieldPath

`string`

### spec.server.extraEnv.variables[].configMapKeyRef

`ConfigMapKeyRef`

Reference to a key in a Kubernetes ConfigMap.

### spec.server.extraEnv.variables[].configMapKeyRef.name

`string` · required

Name of the ConfigMap.

- rule: {"required":true}

### spec.server.extraEnv.variables[].configMapKeyRef.key

`string` · required

Key within the ConfigMap.

- rule: {"required":true}

### spec.server.extraEnv.variables[].configMapKeyRef.optional

`bool`

If true, the env var is silently skipped when the ConfigMap or key does not exist
(instead of blocking pod startup).

### spec.server.extraEnv.variables[].fieldRef

`ObjectFieldRef`

Reference to a pod-level field (metadata.name, status.podIP, etc.).

### spec.server.extraEnv.variables[].fieldRef.apiVersion

`string`

Version of the schema. Defaults to "v1".

### spec.server.extraEnv.variables[].fieldRef.fieldPath

`string` · required

Path of the field to select (e.g., "metadata.name", "status.podIP").

- rule: {"required":true}

### spec.server.extraEnv.variables[].resourceFieldRef

`ResourceFieldRef`

Reference to container resource limits or requests (limits.cpu, requests.memory, etc.).

### spec.server.extraEnv.variables[].resourceFieldRef.containerName

`string`

Container name. Required for init containers; defaults to the current
container for regular containers.

### spec.server.extraEnv.variables[].resourceFieldRef.resource

`string` · required

Resource to select (e.g., "limits.cpu", "requests.memory").

- rule: {"required":true}

### spec.server.extraEnv.variables[].resourceFieldRef.divisor

`string`

Specifies the output format of the exposed resource.
For CPU: "1" means cores. For memory: "1", "1Ki", "1Mi", "1Gi".

### spec.server.extraEnv.secrets

`[]SecretEnvVar`

Individual secret environment variables (sensitive).

### spec.server.extraEnv.secrets[].name

`string` · required

The environment variable name.

- rule: Must be a valid C_IDENTIFIER (start with letter/underscore, contain only letters, digits, underscores)
- rule: {"required":true}

### spec.server.extraEnv.secrets[].value

`string` · sensitive

The secret's value: on Planton a `$secret/<slug>` reference, resolved at deploy; on a
deploy without the platform, the literal. The module writes it into a Kubernetes Secret
and the environment variable references that Secret.

### spec.server.extraEnv.secrets[].secretRef

`KubernetesSecretKeyRef`

Reference to a key within an existing Kubernetes Secret.

### spec.server.extraEnv.secrets[].secretRef.namespace

`string`

The namespace of the Kubernetes Secret.
If not specified, defaults to the namespace where the component is deployed.
Note: Cross-namespace secret references may not be supported by all Helm charts.

### spec.server.extraEnv.secrets[].secretRef.name

`string` · required

The name of the Kubernetes Secret.

- rule: {"required":true}

### spec.server.extraEnv.secrets[].secretRef.key

`string` · required

The key within the Kubernetes Secret that contains the value.

- rule: {"required":true}

### spec.server.extraEnv.secrets[].secretRef.optional

`bool`

If true, the env var is silently skipped when the Secret or key does not exist
(instead of blocking pod startup).

### spec.server.extraEnv.secrets[].valueFrom

`ValueFromRef`

Reference to another Planton resource's secret output field.
The orchestrator resolves this before invoking IaC modules.

### spec.server.extraEnv.secrets[].valueFrom.kind

`enum`

Allowed values (use exactly as shown):

- `unspecified` -- 0: Default/unspecified
- `TestCatalogKindGeneric` -- 1–49: Test/dev/custom
- `TestCatalogKindKubernetes`
- `AwsAlb` -- 1000–1999: AWS resources AwsSubnet is a prerequisite because an ALB requires at least two subnets in different availability zones -- the spec's subnet references must resolve before the load balancer can be created.
- `AwsCertManagerCert`
- `AwsCloudFront`
- `AwsDynamodb`
- `AwsEcrRepo`
- `AwsEcsCluster`
- `AwsEcsService` -- AwsEcsCluster, AwsEcsTaskDefinition, and AwsSubnet are prerequisites because a service schedules a referenced task-definition revision into a referenced live cluster and places task network interfaces into referenced subnets -- all three references must resolve first.
- `AwsEksCluster` -- AwsSubnet and AwsIamRole are prerequisites because the control plane attaches its network interfaces into referenced subnets and assumes a referenced cluster role that must already carry AmazonEKSClusterPolicy.
- `AwsIamRole`
- `AwsLambda`
- `AwsRdsCluster`
- `AwsRdsInstance`
- `AwsRoute53Zone`
- `AwsS3Bucket`
- `AwsLbTargetGroup` -- AwsVpc is a prerequisite because a target group's health checks and target registrations live inside one VPC -- the spec's vpc_id reference must resolve before the group can be created.
- `AwsSecurityGroup` -- AwsVpc is a prerequisite because every security group is created in a VPC; the E2E install profile resolves vpc_id against the VPC prerequisite.
- `AwsVpc`
- `AwsEksNodeGroup` -- AwsEksCluster is a prerequisite because nodes register with a live control plane; AwsIamRole and AwsSubnet back the node role and worker subnet references.
- `AwsIamUser`
- `AwsKmsKey`
- `AwsEc2Instance`
- `AwsClientVpn` -- Every Client VPN endpoint requires an ACM server certificate at create time; the imported self-signed fixture satisfies it. Subnets/VPC are optional composition (a zero-association endpoint is valid) -- composed scenarios declare them via the e2e-prerequisites annotation.
- `AwsDocumentDb`
- `AwsRoute53DnsRecord` -- AwsRoute53Zone is a prerequisite because every record lives inside a hosted zone -- the spec's zone_id reference must resolve before the record can be created.
- `AwsS3ObjectSet` -- AwsS3Bucket is a prerequisite because the object set's bucket reference is required -- objects cannot exist without the bucket that holds them.
- `AwsSqsQueue`
- `AwsSnsTopic`
- `AwsEventBridgeBus`
- `AwsEventBridgeRule`
- `AwsIamOidcProvider`
- `AwsIamPolicy`
- `AwsIamInstanceProfile` -- AwsIamRole is a prerequisite because an instance profile is a wrapper that must contain a role to be useful -- the profile's spec requires a role reference, so the role must be deployed first.
- `AwsLbListener` -- AwsAlb and AwsLbTargetGroup are prerequisites because a listener is an attachment point on a load balancer and its default action almost always forwards to a target group -- both references must resolve before the listener can be created.
- `AwsLbListenerRule` -- AwsLbListener is a prerequisite because a rule only exists as an attachment on a listener -- the listener_arn reference must resolve before the rule can be created.
- `AwsLaunchTemplate`
- `AwsAutoScalingGroup` -- AwsSubnet and AwsLaunchTemplate are prerequisites because a group cannot exist without subnets to place capacity in and a launch template to launch from -- the spec's subnets and launch_template references must resolve before the group can be created.
- `AwsEksAddon` -- AwsEksCluster is a prerequisite because an add-on installs onto a live control plane -- the spec's cluster_name reference must resolve before the add-on can be created.
- `AwsEksFargateProfile` -- AwsEksCluster, AwsIamRole, and AwsSubnet are prerequisites because a Fargate profile attaches to a live control plane, runs pods as a referenced pod-execution role, and launches them into referenced private subnets -- all three references must resolve first.
- `AwsEksAccessEntry` -- AwsEksCluster and AwsIamRole are prerequisites because an access entry grants a referenced IAM principal access to a live control plane -- both references must resolve before the entry can be created.
- `AwsEcsTaskDefinition` -- AwsIamRole is a prerequisite because the kind's default posture -- Fargate with the awslogs logging default -- is rejected by AWS at registration time without an execution role the agent can assume.
- `AwsHttpApiGateway`
- `AwsStepFunction` -- AwsIamRole is a prerequisite because a state machine cannot be created without an execution role it can assume -- the spec's role_arn reference must resolve before the CreateStateMachine call.
- `AwsHttpApiVpcLink` -- AwsSubnet is a prerequisite because a VPC link is a set of managed ENIs provisioned into referenced subnets -- the subnet references must resolve before the link can be created. Security groups are optional on the link, so they compose per-scenario rather than as a registry prerequisite.
- `AwsHttpApiDomain` -- AwsCertManagerCert is a prerequisite because a custom domain cannot be created without a TLS certificate in the same region covering the domain -- the spec's certificate_arn reference must resolve first.
- `AwsVpcEndpoint` -- AwsVpcEndpoint's composed E2E scenarios reference the AwsVpc prerequisite's outputs (vpc_id + default_route_table_id for gateway endpoints) and the AwsSubnet pair's subnet_id outputs (interface endpoints), so both are genuine deploy-order prerequisites.
- `AwsElasticacheUser`
- `AwsElasticacheUserGroup` -- AwsElasticacheUser is a genuine prerequisite: AWS refuses to create a user group that does not contain a user named "default", so a group's composed E2E scenario must resolve a deployed user's outputs.
- `AwsRedshiftServerlessNamespace`
- `AwsRedshiftServerlessWorkgroup` -- The namespace is a genuine prerequisite: a workgroup attaches to exactly one namespace by name at create time, so its composed E2E scenario must resolve a deployed namespace's outputs. AwsSubnet is a prerequisite because Redshift Serverless requires the workgroup's subnets to span three availability zones.
- `AwsRedisElasticache` -- AwsSubnet is a prerequisite because the module builds an ElastiCache subnet group from referenced subnets -- the spec's subnet references must resolve before the replication group can deploy.
- `AwsOpenSearchDomain`
- `AwsMemcachedElasticache`
- `AwsServerlessElasticache`
- `AwsNlb` -- AwsSubnet is a prerequisite because an NLB requires at least one subnet mapping -- the spec's subnet references must resolve before the load balancer can be created.
- `AwsElasticIp`
- `AwsTransitGateway`
- `AwsGlobalAccelerator`
- `AwsSubnet`
- `AwsInternetGateway`
- `AwsNatGateway` -- AwsInternetGateway is a prerequisite because a public NAT gateway can only become available once the VPC it sits in has an internet gateway attached (AWS rejects the create otherwise) -- so the gateway must be deployed first. AwsVpc is a prerequisite because a REGIONAL NAT gateway (availability_mode = regional) references the VPC directly instead of a subnet.
- `AwsEgressOnlyInternetGateway`
- `AwsElasticFileSystem` -- AwsSubnet and AwsSecurityGroup are prerequisites because mount targets (required, min 1) place the file system's NFS endpoints into subnets and attach security groups -- both references must resolve before the CreateMountTarget calls.
- `AwsEfsAccessPoint` -- AwsElasticFileSystem is a prerequisite because an access point is created INTO a file system -- the spec's required file_system_id reference must resolve before the CreateAccessPoint call.
- `AwsFsxLustreFileSystem`
- `AwsFsxOpenzfsFileSystem`
- `AwsFsxWindowsFileSystem` -- Every Windows file system must join an Active Directory domain; the directory itself is external infrastructure (AWS Managed Microsoft AD or a self-managed domain), so only the network dependency is a declarable prerequisite.
- `AwsFsxOntapFileSystem`
- `AwsFsxOntapStorageVirtualMachine`
- `AwsFsxOntapVolume`
- `AwsFsxDataRepositoryAssociation`
- `AwsCognitoUserPool`
- `AwsCognitoIdentityProvider` -- AwsCognitoUserPool is a prerequisite because an identity provider is created INTO a pool -- the spec's required user_pool_id reference must resolve before the CreateIdentityProvider call.
- `AwsCognitoUserPoolClient` -- AwsCognitoUserPool is a prerequisite because an app client is created INTO a pool -- the spec's required user_pool_id reference must resolve before the CreateUserPoolClient call.
- `AwsCognitoResourceServer` -- AwsCognitoUserPool is a prerequisite because a resource server is created INTO a pool -- the spec's required user_pool_id reference must resolve before the CreateResourceServer call.
- `AwsWafWebAcl`
- `AwsWafIpSet`
- `AwsWafRegexPatternSet`
- `AwsCloudwatchLogGroup`
- `AwsCloudwatchAlarm`
- `AwsCloudwatchCompositeAlarm`
- `AwsKinesisStream`
- `AwsKinesisFirehose` -- Every Firehose destination requires an S3 configuration (the primary target for extended_s3; the failed/all-document backup for the rest) and an IAM role Firehose assumes to write to it, so both are hard deploy prerequisites.
- `AwsKinesisStreamConsumer` -- A consumer registers against exactly one stream and cannot exist without it.
- `AwsAthenaWorkgroup`
- `AwsGlueCatalogDatabase`
- `AwsRedshiftCluster`
- `AwsSagemakerDomain` -- AI/ML A domain cannot exist without VPC subnets and a SageMaker execution role (default_user_settings.execution_role_arn is required), so both are hard deploy prerequisites.
- `AwsAppRunnerService` -- A service can run entirely on companion defaults, so the App Runner family's kinds are dependency-free leaves except the VPC connector (which cannot exist without subnets and security groups). A service's companion references (auto scaling / VPC connector / observability / WAF) are optional composition -- scenarios declare them via the e2e-prerequisites annotation.
- `AwsAppRunnerAutoScalingConfiguration`
- `AwsAppRunnerVpcConnector`
- `AwsAppRunnerObservabilityConfiguration`
- `AwsTransitGatewayVpcAttachment` -- AwsTransitGateway is a prerequisite because an attachment cannot exist without the gateway it attaches to; AwsSubnet because the attachment provisions an ENI into at least one subnet (the VPC arrives transitively through the subnet's own prerequisites).
- `AwsTransitGatewayRouteTable` -- Only the gateway is a hard prerequisite: a route table can exist empty. Associations, propagations, and routes referencing attachments are optional composition -- scenarios declare them via the e2e-prerequisites annotation.
- `AwsBatchComputeEnvironment` -- A MANAGED compute environment always launches into VPC subnets, so the subnet is a hard deploy prerequisite (security groups are required only for the Fargate types -- scenario-declared, not a registry edge).
- `AwsBatchJobQueue` -- A job queue cannot exist without at least one VALID compute environment to map onto.
- `AwsBatchSchedulingPolicy`
- `AwsBatchJobDefinition`
- `AwsCodeBuildProject` -- CI/CD
- `AwsCodePipeline`
- `AwsMwaaEnvironment` -- Workflow / Orchestration AwsSubnet and AwsSecurityGroup are prerequisites because the environment's network interfaces are placed in referenced private subnets and AWS requires at least one attached security group at creation.
- `AwsNeptuneCluster` -- Graph Database
- `AwsMemorydbCluster` -- A cluster always launches into a subnet group; the subnets are the hard deploy prerequisite. The ACL it attaches is optional composition (the built-in "open-access" ACL needs no resource) -- scenarios declare the ACL/user chain via the e2e-prerequisites annotation.
- `AwsMemorydbUser`
- `AwsMemorydbAcl` -- An empty ACL is valid (MemoryDB has no mandatory "default" member), so the user is optional composition -- the composed scenario declares it via the e2e-prerequisites annotation, never a registry edge.
- `AwsMskCluster` -- Streaming AwsSubnet and AwsSecurityGroup are prerequisites because brokers are placed in referenced subnets and AWS requires at least one attached security group at creation.
- `AwsMskServerlessCluster` -- AwsSubnet is a prerequisite because the serverless cluster's network interfaces are placed in referenced subnets (security groups are optional -- AWS attaches the VPC default group when none are referenced).
- `AwsLambdaEventSourceMapping` -- AwsLambda is a prerequisite because a mapping cannot exist without the function it invokes (a required reference). Event sources (SQS, Kinesis, DynamoDB, MSK) are optional composition -- scenarios declare them via the e2e-prerequisites annotation rather than taxing every consumer's chain.
- `AwsSnsSubscription` -- AwsSnsTopic is a prerequisite because a subscription cannot exist without the topic it subscribes to (a required reference). Endpoints (SQS queues, Lambda functions, Firehose streams) are optional composition -- scenarios declare them via the e2e-prerequisites annotation rather than taxing every consumer's chain.
- `AwsPlantonRunner` -- AwsSubnet is a prerequisite because the runner appliance places its network interfaces into referenced subnets -- the placement reference must resolve before the appliance can deploy.
- `AwsRoute53HealthCheck`
- `AwsSesConfigurationSet` -- Both SES kinds are dependency-free leaves: an identity's configuration set is optional composition (scenarios declare it via the e2e-prerequisites annotation), and a configuration set's event destinations reference other kinds only optionally.
- `AwsSesEmailIdentity`
- `AwsSecretsManagerSecret` -- A dependency-free leaf: the KMS key, rotation Lambda, and external rotation role references are all optional composition -- scenarios declare them via the e2e-prerequisites annotation, never registry edges.
- `AwsOpenSearchServerlessCollection` -- A dependency-free leaf: the collection-scoped encryption/network/ data-access/retention policies are module-rendered, and the KMS key and data-access principal references are optional composition (e2e-prerequisites annotation).
- `AwsBedrockGuardrail` -- A dependency-free leaf: the KMS key reference is optional composition (e2e-prerequisites annotation); published versions are folded satellites of the guardrail itself.
- `AwsBedrockCustomModel` -- AwsIamRole is a prerequisite because Bedrock assumes the job role to read training data and write outputs; the S3 locations and KMS key are optional composition (e2e-prerequisites annotation).
- `AwsBedrockInferenceProfile` -- A dependency-free leaf: the model source is a foundation model or an AWS system-defined cross-region profile, never a customer resource.
- `AwsBedrockProvisionedThroughput` -- A dependency-free leaf in the registry: capacity is typically bought for an AwsBedrockCustomModel (the default reference), but foundation model ARNs are equally legal, so the edge is optional composition.
- `AwsBedrockModelAccess` -- A dependency-free leaf: the agreement covers an AWS-listed foundation model, never a customer resource.
- `AwsBedrockInvocationLogging` -- Region settings singleton (one invocation-logging configuration per account+region; identity = the region). Delivery destinations are optional references (at least one of CloudWatch/S3, enforced by CEL), so prerequisites stay empty and E2E fixtures ride scenario annotations.
- `AwsBedrockAgent` -- AwsIamRole is a prerequisite because the Bedrock service assumes the agent resource role to invoke models, action-group Lambdas, and knowledge bases; the guardrail, KMS key, provisioned throughput, and collaborator/knowledge-base edges are optional composition (e2e-prerequisites annotation). Action groups, aliases, collaborators, and knowledge-base associations are folded satellites of the agent.
- `AwsBedrockKnowledgeBase` -- AwsIamRole is a prerequisite because the Bedrock service assumes the knowledge-base role to read data sources, call the embedding model, and read/write the vector store; the vector-store and data-source reference edges (OpenSearch, S3, Secrets Manager, ...) are optional composition (e2e-prerequisites annotation). Data sources are folded satellites of the knowledge base.
- `AwsBedrockFlow` -- AwsIamRole is a prerequisite because the Bedrock service assumes the flow execution role to invoke the models, agents, knowledge bases, and Lambdas its nodes reference; the node-level reference edges are optional composition (e2e-prerequisites annotation).
- `AwsBedrockPrompt` -- A dependency-free leaf: variants target AWS-listed foundation models by ID; targeting another agent's alias is optional composition (e2e-prerequisites annotation).
- `AwsBedrockAgentCoreRuntime` -- AwsIamRole is a prerequisite because the AgentCore service assumes the runtime role to pull the container image or read the S3 code bundle and to run the hosted agent; the code-bundle S3 bucket and VPC placement edges are optional composition (e2e-prerequisites annotation). Endpoints and the runtime's resource policy are folded satellites of the runtime.
- `AwsBedrockAgentCoreGateway` -- AwsIamRole is a prerequisite because the gateway assumes its role to reach targets (invoke Lambdas, sign SigV4 requests); the target and credential-provider reference edges (runtime, Lambda, Identity providers, policy engine) are optional composition (e2e-prerequisites annotation). Targets are folded satellites of the gateway - AWS deletes them before the gateway at destroy.
- `AwsBedrockAgentCoreMemory` -- A dependency-free leaf for built-in strategies: the execution role (custom strategies, Kinesis delivery), KMS key, and Kinesis stream edges are optional composition (e2e-prerequisites annotation). Strategies are folded satellites of the memory - AWS serializes their changes through the parent.
- `AwsBedrockAgentCoreIdentity` -- A dependency-free leaf: workload identities, credential providers, and the Cedar policy engine with its policies are all name-keyed arms of one identity-and-access bundle; the KMS key edge is optional composition (e2e-prerequisites annotation). The account/region token-vault CMK is deliberately NOT modeled here (settings singleton).
- `AwsBedrockAgentCoreTools` -- A dependency-free leaf in the SANDBOX/PUBLIC postures: the execution role (recordings, certificates), S3, Secrets Manager, and VPC edges are optional composition (e2e-prerequisites annotation). Browsers, profiles, and code interpreters are name-keyed arms of one tools bundle; AWS exposes no update - every field change recreates the tool.
- `AwsBedrockAgentCoreEvaluation` -- The AgentCore Evaluations bundle - evaluators (LLM-judge or Lambda scorers), harnesses (repeatable agent test benches), and online evaluation configs (continuous scoring of sampled production sessions). Deploys standalone - no arm requires an agent runtime to exist. No registry prerequisite: every arm is optional, so no dependency is required for the kind to function (scenarios compose IAM roles via annotations).
- `AwsBedrockAgentCoreTokenVault` -- Account/region settings singleton: sets the KMS key on the ONE default AgentCore token vault. The KMS reference is conditional on key_type (CEL-enforced), so prerequisites stay empty and E2E fixtures ride scenario annotations.
- `AwsSagemakerModel` -- The immutable serving definition (container image + artifacts + execution role) that endpoints deploy - one container or an inference pipeline.
- `AwsSagemakerEndpoint` -- A real-time inference endpoint WITH its folded endpoint configuration - the configuration is immutable upstream, so the modules roll name-suffixed configurations create-before-destroy and repoint the endpoint.
- `AwsSagemakerNotebookInstance` -- A managed Jupyter notebook EC2 instance with its folded lifecycle configuration (bootstrap scripts).
- `AwsSagemakerFeatureGroup` -- A Feature Store feature group - online and/or offline stores over a declared feature schema.
- `AwsSagemakerModelRegistry` -- A model registry package group with its folded resource policy - model package VERSIONS register into it imperatively (training pipelines), never declaratively.
- `AwsSagemakerPipeline` -- An ML workflow DAG (the SageMaker pipeline-definition JSON) that executions run against - free to create, billed per execution.
- `AwsSagemakerImage` -- A named registry entry exposing YOUR container images to Studio, with folded AWS-numbered versions (append-only by position).
- `AwsSagemakerMlflowServer` -- The classic hourly-billed managed MLflow tracking server (~25 min to provision; billed hourly from creation whether or not anyone is tracking). The serverless successor is AwsSagemakerMlflowApp.
- `AwsSagemakerMlflowApp` -- The serverless MLflow 3.x deployment (billed per use) - standalone, associating with SageMaker domains; NOT a tracking-server satellite.
- `AwsRestApiGateway` -- A full REST API (API Gateway v1): the resource/method tree with inline integrations (or an imported OpenAPI document), one stage with an explicit hash-triggered deployment, and the API-scoped satellites (authorizers, models, validators, gateway responses, policy, documentation, client certificate). Self-contained: a MOCK-integration API needs no other resource.
- `AwsRestApiDomain` -- A custom domain for REST APIs with base-path mappings and - for PRIVATE domains - VPC-endpoint access associations. AwsCertManagerCert is a prerequisite because the domain cannot be created without a TLS certificate covering it.
- `AwsRestApiUsagePlan` -- A usage plan metering REST API consumers - stage coverage, quota, throttles, and the API keys it admits. No registry prerequisite: a plan is valid with no stage coverage (scenarios compose the REST API via annotations).
- `AwsRestApiVpcLink` -- A REST API VPC link fronting an internal Network Load Balancer so REST integrations reach private services. AwsNlb is a prerequisite because AWS rejects link creation without the target balancer.
- `AwsApiGatewayAccountSettings` -- Region settings singleton (one API Gateway account object per account+region; identity = the region). The CloudWatch role is an optional reference (unset = the explicit no-logging posture), so prerequisites stay empty and E2E fixtures ride scenario annotations.
- `AwsCloudTrail` -- The account's API audit trail. AwsS3Bucket is a prerequisite because AWS rejects trail creation without a delivery bucket carrying the CloudTrail service-principal policy. 1240 opens the governance sub-band (1240-1249).
- `AwsConfigRecorder` -- Region singleton (one AWS Config recorder per region, named "default" by AWS; identity = the region). AwsIamRole is a prerequisite because the recorder cannot exist without its service role.
- `AwsConfigRule` -- One AWS Config compliance rule (managed, custom-lambda, or custom-policy; account- or organization-scoped) with optional auto-remediation. Managed rules need no prerequisites; the custom-lambda arm's function reference is conditional, so E2E fixtures ride scenario annotations.
- `AwsGuardDuty` -- Region singleton (AWS allows one GuardDuty detector per account+region; the detector has no name - identity = the region). Satellite references (S3 export bucket, KMS key) are conditional, so E2E fixtures ride scenario annotations.
- `AwsCloudTrailEventDataStore` -- CloudTrail Lake: a queryable, immutable event data store with its own retention and billing lifecycle - no trail required. The KMS key reference is conditional, so E2E fixtures ride scenario annotations.
- `AwsConfigAggregator` -- AWS Config cross-account/cross-region aggregation: the aggregator (collector side) and/or the reciprocal authorization grants (source-account side). Works with zero recorders; the org-source role reference is conditional, so E2E fixtures ride scenario annotations.
- `AwsConfigConformancePack` -- An AWS Config conformance pack (account- or organization-scoped): a template bundle that creates its own Config rules. Deployment requires an active Config recorder in the region (a service-side requirement, not a spec reference), so E2E fixtures ride scenario annotations.
- `AwsGuardDutyMalwareProtectionPlan` -- GuardDuty Malware Protection for S3: scans new objects in one bucket - a standalone plan protecting a bucket, not a detector satellite (its schema carries no detector reference). The execution role and the protected bucket are required references.
- `AwsBackupVault` -- An AWS Backup vault - the encrypted container recovery points live in, as either a standard vault (with its lock, access policy, and notification satellites) or a logically air-gapped vault (AWS's own VaultType discriminator). The KMS and SNS references are conditional, so E2E fixtures ride scenario annotations. 1250 opens the backup sub-band (1250-1259).
- `AwsBackupPlan` -- An AWS Backup plan: scheduled backup rules plus the resource selections that assign resources to them. AwsBackupVault is a prerequisite because every rule requires a target vault; the selections' IAM role is conditional and rides scenario annotations.
- `AwsBackupFramework` -- A Backup Audit Manager framework: compliance controls evaluating backup posture. No schema-required references (the Config recorder its evaluations need is a lane fixture, not a spec reference).
- `AwsBackupReportPlan` -- A Backup Audit Manager report plan: scheduled compliance/job reports delivered to S3. AwsS3Bucket is a prerequisite because the delivery channel's bucket is required.
- `AwsBackupRestoreTestingPlan` -- An AWS Backup restore testing plan with its folded selections: scheduled restore tests proving recovery points actually restore. Vault targeting accepts the "*" wildcard, so fixtures are conditional and ride scenario annotations.
- `AwsBackupSettings` -- Account/region settings singleton for AWS Backup: the account's global settings (cross-account backup) and the region's resource-type opt-in/management preferences. Both provider deletes are no-ops - settings persist after destroy.
- `AwsSsmParameter` -- An SSM Parameter Store entry (String/StringList/SecureString). The parameter's name is an explicit spec field - names are hierarchical paths ("/prod/db/url") metadata.name cannot carry. The KMS reference is conditional (SecureString only), so E2E fixtures ride scenario annotations. 1260 opens the SSM sub-band (1260-1269).
- `AwsSsmDocument` -- A customer-owned SSM document (Command/Automation/Session/...): reusable action definitions managed nodes and automations execute. State Manager associations are their own AwsSsmAssociation kind - an association binds ANY document (AWS-managed included), so it is not this document's satellite.
- `AwsSsmMaintenanceWindow` -- An SSM maintenance window with its folded target registrations and tasks (Run Command / Automation / Lambda / Step Functions) - the targets and tasks are true window satellites (ForceNew window_id edges). Identity is the AWS-generated "mw-..." id.
- `AwsSsmPatchBaseline` -- An SSM patch baseline with its folded patch-group registrations and the account/region default-baseline designation (delete RESTORES AWS's own predefined default for the OS). Identity is the AWS-generated "pb-..." id.
- `AwsSsmAssociation` -- A State Manager association: the binding of an SSM document to targets on a schedule. Split from the document kind because the document reference is a free string with no structural edge - associations routinely bind AWS-managed documents (AWS-RunShellScript, ...) with no user document anywhere, so no registry prerequisite either. Identity is the AWS-generated association UUID.
- `AwsOrganization` -- THE AWS Organization of the deploying account - creating it makes the caller the management account. Trusted service access, delegated administrators, the org's singleton resource policy, and centralized root-access management (IAM's organizations features - a management-account act requiring iam.amazonaws.com trusted access) fold in (none has a life of its own; the standalone service-access resource fights the org's own argument with a perpetual diff). Deleting this deletes the entire organization. 1270 opens the Organizations sub-band (1270-1279). A container kind: the organization is the outer wall of the tenancy tree AWS's own console draws -- organizational units are rooms inside it and member accounts stand inside those. A first-level unit names the organization through its parent reference and is placed inside it; a policy attached to the root names the organization too, but a policy is a guard applied from above, never a resident, and its attachment reference is containment-exempt on its spec.
- `AwsOrganizationalUnit` -- An organizational unit in the org's OU tree. The display name is an explicit spec field (OU names allow spaces metadata.name cannot carry); the parent reference (root or parent OU) is required and immutable, so the organization is a registry prerequisite. A container kind: a unit is a room of the tenancy tree -- nested units and member accounts name it through their parent reference and are placed inside it. A policy attached to a unit names it too and stands beside the room (containment-exempt on the policy's spec): one policy attaches to many units and cannot live in all of them.
- `AwsOrganizationAccount` -- A MEMBER account of the organization: creation, OU placement, and the account-level settings satellites (alternate/primary contacts, opt-in region enablement) fold onto the created account's ID. Destroy is never a clean delete (remove-from-org or ~90-day close) - taught on the spec. No registry prerequisite by the schema-required-only rule (the OU parent reference is optional).
- `AwsOrganizationPolicy` -- An Organizations policy (SCP and its twelve sibling types) with its folded attachments to roots, OUs, and member accounts. The policy type must be enabled on the organization first; AWS-managed policies are never adopted. No registry prerequisite by the schema-required-only rule (attachments are optional).
- `AwsBudget` -- A Budgets budget (COST/USAGE/RI/Savings Plans coverage and utilization) with its folded budget actions as name-keyed satellites - an action exists only on its budget and fires an IAM-policy application, an SCP attachment, or SSM instance stops when a threshold breaches. Budgets is account-global (served from us-east-1; the spec region is the provider endpoint). 1280 opens the cost-management sub-band (1280-1289).
- `AwsCostAnomalyMonitor` -- A Cost Explorer anomaly monitor (DIMENSIONAL over one dimension, or CUSTOM over a CE expression) with its folded alert subscriptions - a subscription's monitor list is the structural edge that makes it this monitor's satellite. Account-global; AWS identifies both by ARN.
- `AwsCostCategory` -- A Cost Explorer cost category: ordered rules (regular expression rules or inherited-value rules) over the recursive CE expression tree, plus split-charge rules. The account's cost-allocation-tag activation toggle is deliberately NOT folded here - it is a per-tag-key account feature with no edge to any category, so many category instances would fight over one account object.
- `AwsIamGroup` -- An IAM group with its folded declarative membership (the authoritative users list) and group policies - name-keyed inline documents plus managed-policy attachments. IAM is global; identity is the group name (renames update in place, the ARN recomputes). 1290 opens the IAM P1 sub-band (1290-1299).
- `AwsIamSamlProvider` -- An IAM SAML identity provider: the account's federation trust anchor, created from the IdP's metadata XML (a public document carrying certificates, not a secret). Identity is the provider ARN; the name is write-once.
- `AwsIamAccountSettings` -- Account settings singleton for IAM (a GLOBAL service - one object per ACCOUNT, not per region): the sign-in alias, the password policy, and the STS global-endpoint token version. Destroy contracts DIFFER per arm (each taught on its arm): the alias truly deletes, the password policy resets to AWS defaults, the STS preference is a no-op delete that persists.
- `AwsCloudwatchDashboard` -- A CloudWatch dashboard: one named dashboard whose widget layout is the dashboard-body JSON document (modeled as a typed Struct, the catalog's uniform policy-document idiom). Dashboards are untaggable at AWS. Identity is the dashboard name; every change is an in-place PutDashboard upsert. 1300 opens the CloudWatch observability P1 sub-band (1300-1309).
- `AwsCloudwatchSynthetics` -- CloudWatch Synthetics: a canary (a scheduled scripted probe running from an S3-staged code bundle under an execution role, writing run artifacts to S3) plus the grouping surface - owned groups and the canary's group associations (joins by group NAME, so shared groups are referenced, never fought over). A groups-only instance manages shared groups with no canary.
- `AwsCloudwatchLogDelivery` -- CloudWatch Logs delivery: the two ways logs leave CloudWatch. The vended-log arm pivots on a delivery SOURCE (one AWS resource whose service vends logs) with name-keyed deliveries fanning out to delivery destinations (S3 / CloudWatch Logs / Firehose / X-Ray), each created inline or referenced by ARN. The cross-account arm is the legacy Kinesis subscription destination with its access policy (whose delete is a no-op at AWS - the policy persists).
- `AwsCloudwatchLogAccountPolicy` -- A CloudWatch Logs account-level policy: one policy object per (name, type) pair per region - data protection, subscription filter, field index, transformer, or metric extraction - applied account-wide, optionally narrowed by selection criteria. Standalone account configuration, never a per-log-group satellite.
- `AwsCloudwatchLogAnomalyDetector` -- A CloudWatch Logs anomaly detector: one detector trains over a LIST of log groups (multi-parent scope - never a single group's satellite), surfacing anomalies on a chosen evaluation frequency with a bounded visibility window.
- `AwsCloudwatchLogResourcePolicy` -- A CloudWatch Logs resource policy: the account-scoped named policy (or resource-scoped policy on one log group ARN) that grants AWS services permission to write logs - Route53 query logging, EventBridge, and friends. Exactly one scope per instance.
- `AwsManagedPrometheus` -- An Amazon Managed Prometheus workspace with its folded satellites: workspace configuration (retention, label-set limits - a created-via-update singleton whose delete is a no-op at AWS), the alert manager definition (strictly one per workspace), name-keyed rule group namespaces, query logging, the workspace resource policy, and alias-keyed anomaly detectors. Scrapers are deliberately NOT folded here - a scraper can target CloudWatch with zero AMP workspaces, so it is its own kind.
- `AwsManagedPrometheusScraper` -- An Amazon Managed Prometheus scraper: the agentless collector. Source is an EKS cluster or a bare VPC placement (both replace-on-change); destination is an AMP workspace or a CloudWatch dataset. Carries its own scraper logging configuration satellite. Scrape configuration is optional on the EKS arm (AWS publishes a default, resolved at deploy) and required on the VPC arm.
- `AwsEventBridgePipe` -- An EventBridge Pipe: one point-to-point integration reading from one source (SQS, Kinesis, DynamoDB streams, MSK or self-managed Kafka, ActiveMQ/RabbitMQ), optionally filtering and enriching in-flight, and delivering to one target (ECS, Batch, Lambda, Step Functions, Kinesis, SQS, Redshift, SageMaker, CloudWatch Logs, EventBridge buses, HTTP via API destinations). The source is fixed for life (replace-on-change); the target swaps in place. 1310 opens the EventBridge extras P1 sub-band (1310-1319).
- `AwsEventBridgeScheduler` -- An EventBridge Scheduler schedule: cron/rate/one-time invocation of one target under an execution role, with flexible time windows, retry policy, and a dead-letter queue. The schedule GROUP is folded own-XOR-existing (a name-and-tags container - the provider's own update path is tags-only); unset means AWS's default group.
- `AwsEventBridgeApiDestination` -- An EventBridge API destination with its connection: the authenticated HTTP(S) endpoint rules, pipes, and schedules invoke. Two independently deployable arms - the CONNECTION (the shareable auth trust anchor: api-key, basic, or OAuth credentials that AWS stores in Secrets Manager) and the DESTINATION (endpoint + method + rate limit) whose connection is owned inline or referenced by ARN.
- `AwsVpcPeering` -- A VPC peering connection, as a request-XOR-accept mode union: the REQUEST arm creates the peering from its VPC toward a peer VPC (same-account auto-accept supported; cross-account/cross-region stays pending until accepted), the ACCEPT arm adopts and accepts a pending connection by ID from the accepter side. DNS-resolution options fold into both arms. 1320 opens the VPC networking P1 sub-band (1320-1329).
- `AwsNetworkAcl` -- A network ACL: the stateless subnet-level firewall - ordered ingress/egress rules (allow or deny, evaluated by rule number) and the subnet associations, all folded in-line as the single declarative owner (the standalone rule/association resources are the same payload and fight the in-line form).
- `AwsManagedPrefixList` -- A customer-managed prefix list: a named, versioned set of CIDR blocks that security-group rules, NACL rules, and route tables reference as one object. Entries fold in-line; max_entries is the capacity contract (referencing consumes that many rule slots regardless of how many entries exist).
- `AwsEbsVolume` -- A standalone EBS volume as a create-XOR-copy union (fresh in a zone, or cloned from another volume) with attachments managed in-line. 1330 opens the block & object storage sub-band (1330-1339).
- `AwsEbsSnapshot` -- An EBS snapshot as a three-way source union (snapshot a volume, copy a snapshot, or import a disk image) with archive tiering, fast snapshot restore, and cross-account share grants in-line.
- `AwsS3DirectoryBucket` -- An S3 directory bucket (S3 Express One Zone): single-AZ, single-digit-millisecond object storage. The modules derive the mandated "{name}--{zone_id}--x-s3" bucket name.
- `AwsS3TableBucket` -- An S3 table bucket (S3 Tables - managed Apache Iceberg storage) with its namespaces, tables, policies, and replication folded in-line as the single declarative owner.
- `AwsS3VectorBucket` -- An S3 vector bucket (AI embedding storage with similarity query) with its vector indexes folded in-line - the natural backend for Bedrock knowledge bases.
- `AwsDlmLifecyclePolicy` -- A Data Lifecycle Manager policy: account-level, tag-targeted snapshot/AMI automation (create, retain, archive, copy cross-region, share, deprecate) as a default-XOR-custom mode union. AwsIamRole is a prerequisite because DLM acts through a required execution role.
- `AwsRoute53ResolverEndpoint` -- A Route 53 Resolver endpoint (the hybrid-DNS bridge between a VPC and outside networks) with its forwarding rules and their VPC associations managed in-line. Subnets place the ENIs and security groups guard them - both schema-required. 1340 opens the DNS & service discovery sub-band (1340-1349).
- `AwsRoute53ResolverFirewall` -- A Route 53 Resolver DNS Firewall rule group with its domain lists, filtering rules, and VPC associations managed in-line - the DNS-layer block/allow policy for VPC egress queries. AwsVpc is a prerequisite because the association arm filters a referenced VPC.
- `AwsRoute53ResolverQueryLog` -- A Resolver query logging configuration (every DNS query VPCs make through the resolver, to CloudWatch Logs / S3 / Firehose) with its VPC associations managed in-line. AwsVpc is a prerequisite because the association arm logs a referenced VPC.
- `AwsCloudMapNamespace` -- An AWS Cloud Map namespace (HTTP-XOR-private-DNS-XOR-public-DNS) with its discoverable services and statically registered instances managed in-line - the service-discovery registry ECS and custom applications look each other up in. AwsVpc is a prerequisite because the private-DNS arm binds its hosted zone to a referenced VPC.
- `AwsAppSyncApi` -- An AppSync API - AWS's managed API service, as a GraphQL API (SDL schema, resolvers over data sources, caching, the MERGED federation variant) XOR an Events API (real-time pub/sub over channel namespaces) - with its data sources, resolvers, functions, types, API keys, and custom domain managed in-line. Every backend reference (data source targets, roles, the certificate) is optional, so no registry prerequisite - lanes exercise the fixture-free arms.
- `AwsLambdaLayer` -- A Lambda layer version - a shared code archive (libraries, custom runtimes) functions attach by ARN - with its cross-account and organization share grants managed in-line. The archive lives in S3 (an optional reference, so no registry prerequisite - lanes compose their own bucket fixture). 1351 sits in the app & data services sub-band (1350-1359; 1350 opens it with AwsAppSyncApi).
- `AwsRdsProxy` -- An RDS Proxy - the managed connection pool between connection-hungry applications and a database - with its connection-pool tuning, additional endpoints, and database target managed in-line. AwsIamRole is a prerequisite because the proxy assumes a required role to read database credentials from Secrets Manager; AwsSubnet because the proxy's network interfaces require at least two subnets.
- `AwsAuroraDsql` -- An Aurora DSQL cluster - serverless, PostgreSQL-compatible distributed SQL with active-active multi-region pairing managed in-line. No prerequisites: a single-region cluster deploys from defaults alone (the KMS and peer references are optional arms).
- `AwsEcrRegistrySettings` -- Region settings singleton (one private ECR registry per account+region): the registry policy, scanning configuration, replication rules, pull-through cache rules, repository creation templates, account settings, and pull-time update exclusions. Repository-scoped surface stays on AwsEcrRepo.
- `AwsPrivateCa` -- An AWS Private Certificate Authority with composed activation (a ROOT self-signs at apply; a subordinate activates from a parent AwsPrivateCa), issued certificates, the ACM renewal permission, and the resource policy managed in-line. No prerequisites: the S3 (CRL) and parent-CA references are optional arms.
- `AwsSesAccountSettings` -- Account/region settings singleton (one SES account object per account+region): the suppression list and VDM posture. 1360 opens the SES P1 sub-band (1360-1369).
- `AzureResourceGroup` -- 2000–2999: Azure resources
- `AzureAksCluster` -- AzureResourceGroup is the only required parent: the cluster is created inside a referenced resource group. Subnet is optional on the default node pool (AKS provisions managed networking when unset).
- `AzureAksNodePool` -- AzureAksCluster is a prerequisite because a node pool attaches to an existing cluster by ARM ID; the resource group chains transitively.
- `AzureContainerRegistry` -- AzureResourceGroup is a prerequisite because a container registry is created inside a resource group.
- `AzureDnsZone` -- AzureResourceGroup is a prerequisite because the DNS zone is created inside a referenced resource group that must already exist.
- `AzureKeyVault` -- AzureResourceGroup is a prerequisite because a key vault is created inside a referenced resource group in composed environments.
- `AzureVirtualNetwork` -- AzureResourceGroup is a prerequisite because a virtual network is created inside a referenced resource group in composed environments.
- `AzureNatGateway` -- AzureResourceGroup is a prerequisite because a NAT gateway is created inside a referenced resource group in composed environments.
- `AzureVirtualMachine` -- AzureNetworkInterface is a prerequisite because a virtual machine attaches at least one NIC (the subnet, network, and resource group chain transitively through the NIC's own prerequisites).
- `AzureStorageAccount` -- AzureResourceGroup is a prerequisite because a storage account is created inside a referenced resource group in composed environments.
- `AzureDnsRecord` -- AzureDnsZone is a prerequisite because a record set is created inside a referenced zone (the resource group chains transitively through the zone). Public DNS zone names are not globally unique, so a shared zone fixture is safe to recreate across scenarios.
- `AzureSubnet` -- AzureVirtualNetwork is a prerequisite because a subnet is an ARM child of a referenced network -- the network must exist before the subnet can be written. (The resource group arrives transitively through the network's own prerequisite declaration.)
- `AzureNetworkSecurityGroup` -- AzureResourceGroup is a prerequisite because a network security group is created inside a referenced resource group in composed environments.
- `AzurePublicIp` -- AzureResourceGroup is a prerequisite because a public IP is created inside a referenced resource group in composed environments.
- `AzurePrivateEndpoint` -- AzureSubnet is a prerequisite because a private endpoint draws its private IP from a referenced subnet (the virtual network and resource group chain transitively through the subnet's own prerequisite). The connection target is polymorphic and the DNS zones / ASGs are optional, so none of those are prerequisites.
- `AzurePrivateDnsZone` -- AzureResourceGroup is a prerequisite because a private DNS zone is created inside a referenced resource group in composed environments.
- `AzureApplicationGateway` -- AzureSubnet is a prerequisite because a gateway cannot exist without its dedicated gateway_ip_configuration subnet (the network and resource group chain transitively through the subnet's own prerequisites); public frontends additionally reference a public IP, but private-only gateways are legal, so it is not a registry prerequisite.
- `AzureLoadBalancer` -- AzureResourceGroup is a prerequisite because a load balancer is created inside a referenced resource group (frontends additionally reference subnets or public IPs, but neither is universally required, so they are not registry prerequisites).
- `AzureRouteTable` -- AzureResourceGroup is a prerequisite because a route table is created inside a referenced resource group in composed environments.
- `AzurePrivateDnsZoneVirtualNetworkLink` -- AzurePrivateDnsZone and AzureVirtualNetwork are prerequisites because a virtual network link is a child resource of a referenced zone and binds it to a referenced network -- both must exist before the link can be written. (The resource group arrives transitively through the zone's and network's own prerequisite declarations.)
- `AzureVirtualNetworkPeering` -- AzureVirtualNetwork is a prerequisite because a peering is an ARM child of its local network and binds it to a remote network -- the local network must exist before the peering can be written. (The resource group arrives transitively through the network's own prerequisite declaration.)
- `AzurePublicIpPrefix` -- AzureResourceGroup is a prerequisite because a public IP prefix is created inside a referenced resource group in composed environments.
- `AzureNetworkInterface` -- AzureSubnet is a prerequisite because a network interface's IP configurations deploy into a subnet (the virtual network and resource group chain transitively through the subnet's own prerequisite).
- `AzureManagedDisk` -- AzureResourceGroup is a prerequisite because a managed disk is created inside a resource group.
- `AzureVirtualMachineScaleSet` -- AzureSubnet is a prerequisite because every scale-set instance's network interface deploys into a subnet (the virtual network and resource group chain transitively through the subnet's own prerequisite).
- `AzureKeyVaultKey` -- AzureKeyVault is a prerequisite because a key is a data-plane object inside a referenced vault -- the vault must exist before the key can be written (the resource group chains transitively through the vault's own prerequisite).
- `AzureKeyVaultCertificate` -- AzureKeyVault is a prerequisite because a certificate is a data-plane object inside a referenced vault -- the vault must exist before the certificate can be enrolled or imported (the resource group chains transitively through the vault's own prerequisite).
- `AzureKeyVaultSecret` -- AzureKeyVault is a prerequisite because a secret is a data-plane object inside a referenced vault -- the vault must exist before the secret can be written (the resource group chains transitively through the vault's own prerequisite). Part of the Key Vault family (2005, 2025-2026) despite the out-of-run number -- enum numbers are pinned by the registry snapshot; never renumber.
- `AzureWebApplicationFirewallPolicy` -- AzureResourceGroup is a prerequisite because a WAF policy is created inside a referenced resource group; the Application Gateways that attach the policy reference it, never the reverse.
- `AzureApplicationSecurityGroup` -- AzureResourceGroup is a prerequisite because an application security group is created inside a referenced resource group; network interfaces, scale-set IP configurations, and NSG security rules reference the group, never the reverse.
- `AzureDiskEncryptionSet` -- AzureKeyVaultKey is a prerequisite because a disk encryption set wraps customer data with a referenced key -- the key (and its vault, which chains transitively) must exist before the set can resolve the key URL at create time.
- `AzurePostgresqlFlexibleServer` -- AzureResourceGroup is a prerequisite because a server is created inside a referenced resource group (VNet injection additionally references a delegated subnet and a private DNS zone, but neither is universally required, so they are not registry prerequisites).
- `AzureRedisCache` -- AzureResourceGroup is a prerequisite because the cache is created inside a referenced resource group (VNet injection additionally references a dedicated subnet, but only the Premium tier supports it, so it is not a registry prerequisite).
- `AzureCosmosdbAccount` -- AzureResourceGroup is a prerequisite because the account is created inside a referenced resource group.
- `AzureMssqlServer` -- AzureResourceGroup is a prerequisite because the logical server is created inside a referenced resource group.
- `AzureMysqlFlexibleServer` -- AzureResourceGroup is a prerequisite because a server is created inside a referenced resource group (VNet injection additionally references a delegated subnet and a private DNS zone, but neither is universally required, so they are not registry prerequisites).
- `AzureMssqlDatabase` -- The parent logical server is referenced via server_id, not auto-deployed: E2E scenarios declare their own server fixture (minimal-server.yaml or the pool-attach chain through AzureMssqlElasticPool) so sequential subtests never destroy and recreate the same globally unique server_name.
- `AzureMssqlElasticPool` -- AzureMssqlServer is a prerequisite because every elastic pool lives on a referenced logical server (the server's resource group is transitive).
- `AzureRedisLinkedServer` -- The target and linked caches are referenced via ARM ids, not auto-deployed: caches are the slowest-provisioning resources in the Azure catalog and their names are globally unique, so E2E scenarios declare their own cache fixtures instead of a registry prerequisite recreating a shared one per run.
- `AzureRedisCacheAccessPolicy` -- The parent cache is referenced via redis_cache_id, not auto-deployed: caches are the slowest-provisioning resources in the Azure catalog and their names are globally unique, so E2E scenarios declare their own cache fixtures instead of a registry prerequisite recreating a shared one per run.
- `AzureRedisCacheAccessPolicyAssignment` -- The parent cache is referenced via redis_cache_id, not auto-deployed: caches are the slowest-provisioning resources in the Azure catalog and their names are globally unique, so E2E scenarios declare their own cache fixtures instead of a registry prerequisite recreating a shared one per run.
- `AzureContainerAppEnvironment` -- AzureResourceGroup is a prerequisite because the environment is created inside a referenced resource group that must already exist.
- `AzureContainerApp` -- AzureContainerAppEnvironment is a prerequisite because every app runs inside a referenced environment (the resource group arrives transitively through it).
- `AzureServicePlan` -- AzureResourceGroup is a prerequisite because the plan is created inside a referenced resource group that must already exist.
- `AzureFunctionApp` -- AzureServicePlan is a prerequisite because a function app runs on a referenced plan (the resource group arrives transitively through the plan). The required storage account is deliberately NOT a registry prerequisite: storage-account names are globally unique, so scenarios bring their own scenario-local account fixtures.
- `AzureLinuxWebApp` -- AzureServicePlan is a prerequisite because a web app runs on a referenced plan (the resource group arrives transitively through the plan).
- `AzureContainerAppJob` -- AzureContainerAppEnvironment is a prerequisite because a job runs inside a referenced environment (the resource group arrives transitively through it).
- `AzureContainerAppEnvironmentStorage` -- AzureContainerAppEnvironment is a prerequisite because the storage registration lives on a referenced environment. The Azure Files share and storage account are deliberately NOT registry prerequisites: storage-account names are globally unique, so scenarios bring their own scenario-local account + share fixtures.
- `AzureContainerAppEnvironmentDaprComponent` -- AzureContainerAppEnvironment is a prerequisite because the Dapr component is registered on a referenced environment.
- `AzureContainerAppEnvironmentCertificate` -- AzureContainerAppEnvironment is a prerequisite because the certificate is stored on a referenced environment.
- `AzureContainerAppEnvironmentManagedCertificate` -- AzureContainerAppEnvironment is a prerequisite because the managed certificate is provisioned on a referenced environment.
- `AzureLogAnalyticsWorkspace` -- AzureResourceGroup is a prerequisite because the workspace is created inside a referenced resource group that must already exist.
- `AzureApplicationInsights` -- AzureLogAnalyticsWorkspace is a prerequisite because workspace-based Application Insights stores its telemetry in a referenced workspace (the resource group chains transitively through the workspace).
- `AzureMonitorDiagnosticSetting` -- AzureLogAnalyticsWorkspace is a prerequisite because the setting's scenarios route a fixture workspace's telemetry (the workspace doubles as target and destination); the target itself is polymorphic.
- `AzureMonitorActionGroup` -- AzureResourceGroup is a prerequisite because the action group is created inside a referenced resource group that must already exist.
- `AzureMonitorMetricAlert` -- AzureMonitorActionGroup is a prerequisite because a metric alert's actions fire into a referenced action group (the resource group chains transitively); alert scopes are polymorphic.
- `AzureMonitorScheduledQueryAlert` -- AzureLogAnalyticsWorkspace is a prerequisite because the rule queries a referenced workspace scope; AzureMonitorActionGroup because its action fires into a referenced action group.
- `AzureMonitorActivityLogAlert` -- AzureMonitorActionGroup is a prerequisite because an activity log alert's actions fire into a referenced action group (the resource group chains transitively). The alert itself is subscription-global and its scopes are polymorphic.
- `AzureApplicationInsightsStandardWebTest` -- AzureApplicationInsights is a prerequisite because a standard web test binds to a referenced Application Insights component (the resource group chains transitively through the kind).
- `AzureUserAssignedIdentity` -- AzureResourceGroup is a prerequisite because the identity is created inside a referenced resource group that must already exist.
- `AzureRoleAssignment` -- AzureResourceGroup and AzureUserAssignedIdentity are prerequisites because an assignment grants a role at a referenced scope (most commonly a resource group) to a referenced principal (most commonly a managed identity) -- both must exist before the grant can be written.
- `AzureRoleDefinition` -- AzureResourceGroup is a prerequisite because a custom role definition is created at a referenced scope, most commonly a resource group in composed environments -- the scope must exist before the definition can be written.
- `AzureFederatedIdentityCredential` -- AzureUserAssignedIdentity is the prerequisite because a federated identity credential is a child resource of a referenced managed identity -- the identity must exist before the credential can be written on it. (The resource group arrives transitively through the identity's own prerequisite declaration.)
- `AzureServiceBusNamespace` -- AzureResourceGroup is a prerequisite because a Service Bus namespace is created inside a referenced resource group in composed environments. The namespace is the container every Service Bus messaging entity (queue, topic, subscription, authorization rule, geo-DR pairing) nests under. The child kinds deliberately declare NO registry prerequisite: namespace names are globally unique with a post-delete name hold, so E2E composes them with scenario-local namespace fixtures instead of a shared recreate-per-scenario prerequisite.
- `AzureEventHubNamespace` -- AzureResourceGroup is a prerequisite because an Event Hub namespace is created inside a referenced resource group in composed environments. The namespace is the container every Event Hubs entity (event hub, consumer group, authorization rule, schema group, geo-DR pairing, customer-managed key) nests under. The child kinds deliberately declare NO registry prerequisite: namespace names are globally unique with a post-delete name hold, so E2E composes them with scenario-local namespace fixtures instead of a shared recreate-per-scenario prerequisite.
- `AzureServiceBusQueue`
- `AzureServiceBusTopic`
- `AzureServiceBusSubscription`
- `AzureServiceBusAuthorizationRule`
- `AzureServiceBusDisasterRecoveryConfig`
- `AzureEventHub`
- `AzureEventHubConsumerGroup`
- `AzureEventHubAuthorizationRule`
- `AzureFrontDoorProfile` -- AzureResourceGroup is a prerequisite because a Front Door profile is created inside a referenced resource group in composed environments. The profile is the container every Front Door delivery resource (endpoint, origin group, origin, route) nests under.
- `AzureFrontDoorEndpoint` -- AzureFrontDoorProfile is a prerequisite because an endpoint is an ARM child of a referenced profile -- the profile must exist before the endpoint can be written. (The resource group arrives transitively through the profile's own prerequisite declaration.)
- `AzureFrontDoorOriginGroup` -- AzureFrontDoorProfile is a prerequisite because an origin group is an ARM child of a referenced profile. A container kind: every origin is an ARM child of its origin group, so on a diagram the origins stand inside the group that load-balances them (the profile's room holds the group). A route or a rule that forwards TO an origin group lives in its endpoint or its rule set, never in the group, and those references say so with containment_exempt.
- `AzureFrontDoorOrigin` -- AzureFrontDoorOriginGroup is a prerequisite because an origin is an ARM child of a referenced origin group (the profile and resource group chain transitively).
- `AzureFrontDoorRoute` -- A route attaches to an endpoint (its ARM parent) and forwards to an origin group whose origins must exist before ARM accepts the route -- so both the endpoint and the origin chain are genuine deploy-order prerequisites.
- `AzureFrontDoorRuleSet` -- AzureFrontDoorProfile is a prerequisite because a rule set is an ARM child of a referenced profile. The rules live inside the set (they form one ordered policy document); routes attach the set by ARM ID.
- `AzureFrontDoorCustomDomain` -- AzureFrontDoorProfile is a prerequisite because a custom domain is an ARM child of a referenced profile. The DNS zone and (for bring-your-own certificates) the Front Door secret are optional references, not deploy-order prerequisites.
- `AzureFrontDoorSecret` -- AzureFrontDoorSecret is a prerequisite-light kind: only the profile (its ARM parent) must exist. The Key Vault certificate it wraps is a reference resolved before the module runs; its vault chain is exercised through scenario-local fixtures in E2E.
- `AzureFrontDoorFirewallPolicy` -- AzureResourceGroup is a prerequisite because the Front Door WAF policy is created inside a referenced resource group -- it is a GLOBAL resource, not a profile child (a different ARM type than the regional Application Gateway WAF policy). Security policies attach it to profiles; the policy itself depends on nothing else.
- `AzureFrontDoorSecurityPolicy` -- A security policy is an ARM child of a profile that associates a referenced WAF policy with referenced domains -- so the endpoint (the default-domain association target; the profile arrives transitively through it) and the WAF policy are genuine deploy-order prerequisites.
- `AzureStorageContainer` -- None of the storage data-service kinds declares a registry prerequisite on AzureStorageAccount: account names are GLOBALLY unique and Azure holds a just-deleted name, so a recreate-per-scenario fixture would hang -- their E2E scenarios declare scenario-local account fixtures instead. Deploy ordering in composed environments still flows from the storage_account_id reference itself.
- `AzureStorageShare`
- `AzureStorageQueue`
- `AzureStorageTable`
- `AzureStorageEncryptionScope`
- `AzureStorageDataLakeGen2Filesystem`
- `AzureStorageLocalUser`
- `AzureStorageObjectReplication`
- `AzureCosmosdbSqlDatabase` -- None of the Cosmos DB data-service kinds declares a registry prerequisite on AzureCosmosdbAccount: account names are GLOBALLY unique DNS labels, so a recreate-per-scenario fixture would risk name-reuse hangs -- their E2E scenarios declare scenario-local account fixtures instead. Deploy ordering in composed environments still flows from the cosmosdb_account_id / parent-database references themselves.
- `AzureCosmosdbSqlContainer`
- `AzureCosmosdbMongoDatabase`
- `AzureCosmosdbMongoCollection`
- `AzureCosmosdbSqlRoleDefinition`
- `AzureCosmosdbSqlRoleAssignment`
- `AzureManagedRedis` -- AzureResourceGroup is the cluster's only registry prerequisite: the cluster is created inside a referenced resource group. The geo-replication and access-policy-assignment children declare NO prerequisite on AzureManagedRedis: clusters are expensive, slow-provisioning parents, so their E2E scenarios declare scenario-local cluster fixtures instead of recreating a shared one per scenario. Deploy ordering in composed environments still flows from the managed_redis_id references themselves.
- `AzureManagedRedisGeoReplication`
- `AzureManagedRedisAccessPolicyAssignment`
- `AzureEventHubDisasterRecoveryConfig`
- `AzureEventHubSchemaGroup`
- `AzureEventHubCluster` -- AzureResourceGroup is a prerequisite because a dedicated Event Hubs cluster is created inside a referenced resource group in composed environments. Note: clusters cannot be deleted for 4 hours after creation (Azure's moratorium), so E2E treats this kind as offline-gated.
- `AzureEventHubNamespaceCustomerManagedKey`
- `AzureMssqlFailoverGroup` -- AzureMssqlServer is a prerequisite because a failover group is created on a referenced primary logical server and points at a partner server; the primary (and its resource group, which chains transitively) must exist before the group can be written.
- `AzureContainerAppCustomDomain` -- AzureContainerApp is a prerequisite because the domain binding lives in a referenced app's ingress configuration (the environment and resource group chain transitively through the app).
- `AzureFirewallPolicy`
- `AzureFirewallPolicyRuleCollectionGroup` -- AzureFirewallPolicy is a prerequisite because a rule collection group is a child document of a referenced policy (the resource group chains transitively through the policy).
- `AzureFirewall` -- AzureSubnet is a prerequisite because a VNet-deployed firewall's data path lives in a dedicated subnet that must be named exactly "AzureFirewallSubnet" (the virtual network and resource group chain transitively through the subnet). The E2E install profile publishes a fixture subnet with that exact name and a /26 prefix.
- `AzureIpGroup`
- `AzureVirtualNetworkGateway` -- AzureSubnet is a prerequisite because every virtual network gateway lives in a dedicated subnet named exactly "GatewaySubnet" (the virtual network and resource group chain transitively through the subnet); the subnet install profile publishes a fixture instance with that exact ARM name. AzurePublicIp is a prerequisite because a VPN-type gateway (the default shape) requires a public IP per ip configuration; the address install profile publishes a dedicated zone-redundant instance (a gateway binds its address exclusively, and the AZ gateway SKUs require zones on it).
- `AzureVirtualNetworkGatewayConnection` -- Both gateways are prerequisites: a connection joins a virtual network gateway to a far side, and the site-to-site far side is a local network gateway (the GatewaySubnet, VNet, and resource group chain transitively through the virtual network gateway).
- `AzureLocalNetworkGateway`
- `AzurePrivateLinkService` -- AzureSubnet is the sole prerequisite: every NAT ip configuration draws its address from a subnet with private-link-service network policies disabled (the subnet install profile publishes a fixture instance with that flag off). The Standard load balancer whose frontend the service typically fronts is NOT a registry prerequisite -- the spec's destination is an exactly-one-of (load balancer frontend OR fixed destination IP), so scenarios that use the load-balancer shape declare it via the planton.dev/e2e-prerequisites annotation instead.
- `AzureExpressRouteCircuit`
- `AzureExpressRouteCircuitPeering` -- The circuit is the prerequisite: a peering is an ARM child of the circuit, addressed by the circuit's name (the resource group chains transitively through the circuit).
- `AzureExpressRouteGateway` -- The hub is the prerequisite: ARM requires an ExpressRoute Gateway to be deployed INTO a Virtual WAN hub (the WAN and resource group chain transitively through the hub).
- `AzureExpressRoutePort` -- ExpressRoute Port: your own physical port pair on a Microsoft edge router (ExpressRoute Direct), from whose bandwidth circuits are carved. Self-contained -- only the resource group is required.
- `AzureVirtualWan` -- Virtual WAN: the umbrella of Azure's managed hub-and-spoke networking, under which virtual hubs and their gateways are created. Self-contained -- only the resource group is required. A container: its hubs and its branch sites are created into it, so on a diagram the WAN is the room they stand in.
- `AzureVirtualHub` -- The WAN is the prerequisite: this kind models the Virtual WAN hub (virtual_wan_id is required; standalone hubs are the legacy Route Server construction, which has its own ARM surface). The resource group chains transitively through the WAN. A container: ARM deploys the hub's VPN, ExpressRoute, and point-to-site gateways and its spoke connections INTO the hub, so on a diagram the hub is the room they stand in.
- `AzureVirtualHubConnection` -- Both sides of the attachment are prerequisites: the hub being joined and the spoke virtual network being attached.
- `AzureVpnGateway` -- The hub is the prerequisite: ARM deploys a Virtual WAN VPN gateway INTO a virtual hub (virtual_hub_id is required and immutable; the WAN and resource group chain transitively through the hub). ARM allows one VPN gateway per hub.
- `AzureVpnGatewayConnection` -- Both ends of the tunnel are prerequisites: a connection is an ARM child of the VPN gateway and pins each of its links to a specific link of the remote VPN site (the hub, WAN, and resource group chain transitively through the gateway).
- `AzureVpnSite` -- The WAN is the prerequisite: a VPN site is the Virtual WAN world's address-book entry for one branch location (virtual_wan_id is required; the classic-world sibling without a WAN is AzureLocalNetworkGateway). The resource group chains transitively through the WAN.
- `AzurePointToSiteVpnGateway` -- The hub and the server configuration are both prerequisites: a point-to-site VPN gateway deploys INTO a virtual hub (one P2S gateway per hub, a slot separate from the hub's site-to-site VPN gateway) and is born pointing at the VPN server configuration that defines how its users authenticate -- both ARM-required and fixed at creation. The WAN and resource group chain transitively through the hub.
- `AzureVpnServerConfiguration` -- Self-contained -- only the resource group is required: a VPN server configuration is the reusable "who may connect and how" authentication policy (Entra ID / certificate / RADIUS) that point-to-site VPN gateways attach to; it references no other Azure resource.
- `AzureCognitiveAccount` -- Self-contained -- only the resource group is required: an Azure AI services account (Azure OpenAI, the multi-service AIServices account, the single-service accounts) needs no other Azure resource; subnets (network rules), Key Vault keys (CMK), storage accounts and user-assigned identities are optional references. A container kind: every model deployment and every project is an ARM child created onto the account (accounts/{account}/deployments/{name}, .../projects/{name}), so a diagram draws them inside it. The subnets the account admits through its network ACLs or injects agent compute into are access, never placement (containment_exempt on those fields).
- `AzureCognitiveDeployment` -- An ARM child of its account: a model deployment (which model runs, at which throughput class) exists only on an Azure AI services account of kind "OpenAI" or "AIServices".
- `AzureCognitiveAccountProject` -- An ARM child of its account: an AI Foundry project exists only on an "AIServices"-kind account with project management enabled.
- `AzureMachineLearningWorkspace` -- The workspace REQUIRES all three companion services at creation (default storage, secrets vault, telemetry) -- genuine deploy-order prerequisites, each with its own fixture profile. A container kind: datastores, compute clusters and instances, and online and batch endpoints are ARM children of the workspace (workspaces/{ws}/datastores|computes|onlineEndpoints|batchEndpoints/{name}), so a diagram draws them inside it. The companion services it names are reaches (containment_exempt on those fields), and so is the subnet its serverless compute or a compute's nodes attach to.
- `AzureMachineLearningDatastore` -- An ARM child of its workspace. The storage target (container, filesystem or share) is scenario-declared via the e2e-prerequisites annotation -- only the blob scenario needs a container, so it is not a kind-wide prerequisite.
- `AzureMachineLearningComputeCluster` -- An ARM child of its workspace (.../computes/{name}) -- the auto-scaling pool of VMs training jobs run on.
- `AzureMachineLearningComputeInstance` -- An ARM child of its workspace (.../computes/{name}) -- a single always-on VM serving as one data scientist's cloud workstation.
- `AzureAiFoundry` -- The hub REQUIRES both companion services at creation (secrets vault, default storage) -- genuine deploy-order prerequisites, each with its own fixture profile. A container kind: every AI Foundry project is created inside its hub and deploys into the hub's resource group (the project spec carries none), so a diagram draws the projects inside the hub. The vault and storage account the hub names are reaches (containment_exempt on those fields).
- `AzureAiFoundryProject` -- Deploys into its hub's resource group (the provider derives the group from the hub reference -- the project spec carries none).
- `AzureSearchService`
- `AzureMachineLearningOnlineEndpoint` -- An ARM child of its workspace (.../onlineEndpoints/{name}) -- the stable scoring address applications call. azurerm carries no ML endpoint resources; the modules write the raw ARM shape at a pinned api-version (azapi / azure-native). A container kind in its turn: every online deployment is an ARM child created under the endpoint whose traffic map routes to it (.../onlineEndpoints/{endpoint}/deployments/{name}) and cannot exist without it, so a diagram draws the deployments inside the endpoint, inside the workspace.
- `AzureMachineLearningOnlineDeployment` -- An ARM child of its endpoint (.../deployments/{name}) -- the running copy of a model the endpoint's traffic map routes to.
- `AzureMachineLearningBatchEndpoint` -- An ARM child of its workspace (.../batchEndpoints/{name}) -- the stable address batch scoring jobs are submitted to. azurerm carries no ML endpoint resources; the modules write the raw ARM shape at a pinned api-version (azapi / azure-native). A container kind in its turn: every batch deployment is an ARM child created under the endpoint whose default-deployment pointer routes to it (.../batchEndpoints/{endpoint}/deployments/{name}), so a diagram draws the deployments inside the endpoint, inside the workspace. The compute cluster a deployment runs on is a reach, not a room.
- `AzureMachineLearningBatchDeployment` -- An ARM child of its endpoint (.../deployments/{name}) -- the job recipe (model, compute, batching behavior) the endpoint's default-deployment pointer routes submissions to.
- `AzureRecoveryServicesVault` -- The Recovery Services vault (Microsoft.RecoveryServices/vaults) -- the safe that classic Azure Backup data and Site Recovery configuration live in. Backup policies and protected items are ARM children of a vault, so the vault is a container kind: a diagram draws them inside it, as the portal's own blades do.
- `AzureBackupPolicyVm` -- An ARM child of its vault (.../backupPolicies/{name}) -- the schedule and retention rules that govern IaaS VM backups.
- `AzureBackupProtectedVm` -- An ARM child of its vault (.../protectedItems/...) -- the binding that puts one virtual machine under a backup policy's protection.
- `AzureBackupPolicyFileShare` -- An ARM child of its vault (.../backupPolicies/{name}) -- the schedule and retention rules that govern Azure Files share backups (snapshot or vaulted).
- `AzureBackupProtectedFileShare` -- An ARM child of its vault (.../protectedItems/AzureFileShare;...) -- the binding that puts one Azure Files share under a backup policy's protection. The share's storage account must already be registered with the vault (AzureBackupContainerStorageAccount). Prerequisite ORDER is load-bearing for teardown (destroy runs in reverse): the registration must list AFTER the share so it unregisters FIRST -- Azure Backup holds a DoNotDelete lock on a registered storage account, and a share delete under that lock fails ScopeLocked.
- `AzureDataProtectionBackupVault` -- The Data Protection backup vault (Microsoft.DataProtection/ backupVaults) -- the safe that MODERN Azure Backup data lives in (managed disks, blob storage, AKS clusters, MySQL/PostgreSQL flexible servers, Data Lake storage). Backup policies and backup instances are ARM children of a vault, so the vault is a container kind: a diagram draws them inside it, as the portal's own blades do.
- `AzureDataProtectionBackupPolicy` -- An ARM child of its vault (.../backupPolicies/{name}) -- the schedule and retention rules for ONE Data Protection datasource type (blob storage, disk, Kubernetes cluster, MySQL/PostgreSQL flexible server, or Data Lake storage), modeled as one kind with variant blocks.
- `AzureDataProtectionBackupInstance` -- An ARM child of its vault (.../backupInstances/{name}) -- the binding that puts ONE datasource (a managed disk, a storage account's blob services, an AKS cluster, a MySQL/PostgreSQL flexible server, or a Data Lake storage account) under a Data Protection backup policy, modeled as one kind with variant blocks. The vault's managed identity must hold the datasource roles Azure Backup requires BEFORE the instance is created.
- `AzureBastionHost` -- AzureSubnet and AzurePublicIp are prerequisites because a dedicated-infrastructure Bastion host (Basic/Standard/Premium -- the default shapes) deploys into a subnet named exactly "AzureBastionSubnet" and binds a Standard static public IP EXCLUSIVELY (the virtual network and resource group chain transitively through the subnet). The Developer SKU instead attaches to a virtual network directly and uses neither.
- `AzureNetworkWatcherFlowLog` -- AzureVirtualNetwork and AzureStorageAccount are prerequisites because a flow log records a network-scoped target (a virtual network in the common case; subnets and network interfaces chain through the network) into a referenced storage account. The regional Network Watcher parent is NOT a prerequisite: Azure auto-creates it ("NetworkWatcher_{region}" in "NetworkWatcherRG") the moment the region hosts a virtual network, and the flow log references it by name. Traffic Analytics' Log Analytics workspace is an optional arm, declared by scenarios that use it.
- `AzurePrivateDnsResolver` -- AzureVirtualNetwork and AzureSubnet are prerequisites because a DNS Private Resolver anchors to a referenced virtual network (at most ONE resolver per network -- Azure enforces it) and each of its inbound/outbound endpoints occupies its own dedicated subnet delegated to "Microsoft.Network/dnsResolvers" (the resource group chains transitively through the network and subnets).
- `AzurePrivateDnsResolverForwardingRuleset` -- AzurePrivateDnsResolver is a prerequisite because a DNS forwarding ruleset steers a resolver's OUTBOUND endpoints -- it binds their ARM ids (at most 2, same resolver) at creation. (The resource group and network chain transitively through the resolver's own prerequisite declarations.)
- `AzurePrivateDnsRecord` -- AzurePrivateDnsZone is a prerequisite because every record set is created inside a referenced private DNS zone (the resource group chains transitively through the zone's own prerequisite).
- `AzureTrafficManagerProfile` -- AzureResourceGroup is a prerequisite because a Traffic Manager profile is created inside a referenced resource group (the profile itself is a global service -- the group only holds its metadata record). A container kind: every endpoint is an ARM child of its profile, so on a diagram the endpoints stand inside the profile that steers traffic to them. A nested endpoint that points AT another profile lives in its own parent profile, never in the one it targets, and that reference says so with containment_exempt.
- `AzureTrafficManagerEndpoint` -- AzureTrafficManagerProfile is a prerequisite because every endpoint is created inside a referenced profile -- it is the destination a profile steers traffic to (the resource group chains transitively through the profile's own prerequisite).
- `AzureMonitorAutoscaleSetting` -- AzureResourceGroup is a prerequisite because an autoscale setting is created inside a referenced resource group. The scalable TARGET it controls is a no-default reference (many kinds can be scaled), so no target kind is declared here -- scenarios declare their own target fixture.
- `AzureMonitorDataCollectionRule` -- The Azure Monitor data collection rule (DCR) -- the routing table declaring what telemetry the Azure Monitor Agent collects and where it lands. AzureResourceGroup is a prerequisite because a rule is created inside a referenced resource group; AzureLogAnalyticsWorkspace because a workspace is the canonical destination a rule routes to (the smoke scenario's shape). Machines attach to a rule with AzureMonitorDataCollectionRuleAssociation resources.
- `AzureEventgridTopic` -- The Azure Event Grid custom topic -- the HTTPS endpoint an application publishes its own events to, fanned out to handlers by event subscriptions. One topic is one event stream with its own endpoint and access keys; for many streams behind one endpoint see AzureEventgridDomain. A container kind: every event subscription on the topic is an ARM child created under it ({topic_id}/providers/ Microsoft.EventGrid/eventSubscriptions/{name}) and cannot exist without it, so on a diagram the topic is the room its subscriptions stand in, as the portal's own blade lists them. A kind that merely publishes to or subscribes from a topic while living elsewhere (a Data Factory trigger, an Event Grid namespace's MQTT route) says so with containment_exempt.
- `AzureEventgridDomain` -- The Azure Event Grid domain -- ONE publishing endpoint and one pair of access keys serving many event streams (domain topics), the multi-tenant pattern. Topics inside the domain are auto-managed by Azure or declared explicitly as AzureEventgridDomainTopic resources. A container kind: every domain topic is an ARM child of its domain ({domain_id}/topics/{name}), so on a diagram the domain is the room its tenant streams stand in.
- `AzureEventgridSystemTopic` -- The Azure Event Grid system topic -- the subscription surface for events AZURE ITSELF publishes about one of your resources (a storage account's blob events, a resource group's lifecycle events). One system topic per source resource per topic type; event subscriptions attach to it to route events to handlers. A container kind: every subscription on a system topic is an ARM child created under it ({system_topic_id}/eventSubscriptions/{name}), so on a diagram the system topic is the room its subscriptions stand in.
- `AzureEventgridEventSubscription` -- The Azure Event Grid event subscription -- the delivery instruction routing events from a source (a custom topic, domain, domain topic, system topic, resource group, or subscription) to a handler (a Function, Event Hub, Service Bus queue/topic, storage queue, hybrid connection, or webhook), with filtering, retry, and dead-letter behavior.
- `AzureEventgridNamespace` -- The Azure Event Grid namespace -- the capacity-scaled hub of the newer Event Grid: hosts CloudEvents namespace topics and an optional MQTT broker behind one set of regional endpoints, sized in throughput units. A container kind: every namespace topic is an ARM child of its namespace ({namespace_id}/topics/{name}), so on a diagram the namespace is the room its streams stand in.
- `AzureDataFactory` -- The Azure Data Factory -- the workspace every other Data Factory resource lives inside: pipelines, data flows, linked services, datasets, triggers, and integration runtimes are all created against a factory's ARM ID. A container kind for exactly that reason: those six kinds are ARM children of the factory, so on a diagram the factory is the room they stand in.
- `AzureDataFactoryPipeline` -- One unit of work inside an Azure Data Factory ({factory_id}/pipelines/{name}) -- an ordered set of activities that executes as a whole when triggered.
- `AzureDataFactoryDataFlow` -- A Data Factory data flow ({factory_id}/dataflows/{name}) -- a visually-designed data transformation executed on managed Spark, or, as a flowlet, a reusable snippet other data flows embed. One kind covers both provider forms (they share one schema and one name namespace inside the factory).
- `AzureDataFactoryLinkedService` -- A Data Factory linked service ({factory_id}/linkedservices/{name}) -- a saved connection in the factory's address book: where an external system lives and how to authenticate to it. One kind covers every connection type Azure models as a first-class resource (storage, SQL family, Cosmos DB, Databricks, Key Vault, SFTP, web APIs, and more) as variants in one factory-scoped name namespace, plus the raw-JSON custom form.
- `AzureDataFactoryDataset` -- A Data Factory dataset ({factory_id}/datasets/{name}) -- a named view of data inside a system a linked service already connects to: which container and path, which table, which file format. One kind covers every data shape Azure models as a first-class dataset resource (delimited text/CSV, JSON, Parquet, binary, blob, HTTP, the SQL family, Snowflake, Cosmos DB) as variants in one factory-scoped name namespace, plus the raw-JSON custom form.
- `AzureDataFactoryTrigger` -- A Data Factory trigger ({factory_id}/triggers/{name}) -- the instruction that starts pipelines automatically: on a clock schedule, per contiguous tumbling window, on storage blob events, or on Event Grid custom events. One kind covers all four provider trigger resources as variants (one ARM namespace, one started/stopped lifecycle).
- `AzureDataFactoryIntegrationRuntime` -- A Data Factory integration runtime ({factory_id}/integrationRuntimes/{name}) -- the compute engine a factory's pipelines, data flows, and copy activities run on. One kind covers all three engine flavors as variants in one factory-scoped name namespace: the managed data-flow compute, the managed SSIS package runtime, and the self-hosted agent registration (which issues the authorization keys agents join with).
- `AzureComputeGallery` -- The Azure Compute Gallery -- the shared library an organization keeps its approved VM images in. Image definitions (AzureComputeGalleryImage) live inside it; VMs and scale sets deploy from their published, region-replicated versions. A container kind: every image definition is an ARM child of its gallery ({gallery_id}/images/{name}), so on a diagram the definitions stand inside the library that publishes them. The image ids a VM, a scale set, or a disk boots from are plain strings and never place anything inside the gallery.
- `AzureComputeGalleryImage` -- A gallery image ({gallery_id}/images/{name}) -- one image definition inside a Compute Gallery (marketplace-style identity, OS type, security posture) plus its published versions, each replicated to its own target regions. VMs deploy from a version's ARM ID or from the definition's ID to get the latest version.
- `AzureAvailabilitySet` -- The availability set -- the classic pre-zones placement grouping that spreads VMs across separate fault and update domains so one hardware failure or maintenance window cannot take them all down. VMs join the set at creation.
- `AzureDiskSnapshot` -- The managed disk snapshot -- a point-in-time copy of a disk used for backup, cloning, and as the source of gallery image versions. Incremental snapshots store only the delta since the previous snapshot of the same disk.
- `AzureContainerInstance` -- The Azure Container Instance container group -- serverless containers billed per second: one or more containers sharing a lifecycle, network, and volumes (plus one-shot init containers), with no cluster or VM to manage. Public, subnet-private, or IP-less postures.
- `AzureFunctionAppFlexConsumption` -- The Azure Function App on the Flex Consumption plan -- Azure's newest serverless Functions hosting model: per-instance memory selection, a configurable scale-out ceiling, always-ready instance pools, and explicit blob-container deployment storage. Requires an FC1-SKU service plan, which is deliberately NOT a registry prerequisite: the shared plan fixture serves the classic app tiers, and an FC1 plan is cheap to create per scenario (no idle compute cost), so scenarios bring their own plan fixture -- the same reasoning that keeps the globally-unique storage account scenario-local for AzureFunctionApp.
- `AzureMongoCluster` -- The Azure Cosmos DB for MongoDB vCore cluster -- Azure's modern managed MongoDB: a real MongoDB engine on dedicated vCore tiers with sharding, zone-redundant HA, and point-in-time restore.
- `AzureFabricCapacity` -- The Microsoft Fabric capacity -- the billing and compute anchor of Microsoft Fabric: workspaces assign themselves to a capacity, and its F-SKU sets how much compute every workload on it shares. azurerm's entire Fabric surface is this one resource (workspaces and items live in Microsoft's dedicated fabric provider).
- `AzureBackupContainerStorageAccount` -- Registers a storage account with a Recovery Services vault as a backup container (.../protectionContainers/StorageContainer;...) -- one registration per storage-account-and-vault pair, required BEFORE any of the account's file shares can be protected. Part of the backup family (2175-2179) despite the out-of-run number -- enum numbers are pinned by the registry snapshot; never renumber.
- `AzureDataProtectionResourceGuard` -- The Data Protection Resource Guard (Microsoft.DataProtection/ resourceGuards) -- the approval gate behind Multi-User Authorization: privileged vault operations (disabling soft delete, reducing retention) require an approval through a guard, which typically lives in a DIFFERENT administrator's scope. Vaults reference a guard by its ARM ID. Part of the backup family (2175-2182) despite the out-of-run number -- enum numbers are pinned by the registry snapshot; never renumber.
- `AzurePrivateDnsResolverVirtualNetworkLink` -- The attachment that makes a DNS forwarding ruleset take effect in one virtual network ({ruleset_id}/virtualNetworkLinks/{name}) -- one link per ruleset-network pair, up to 500 per ruleset, spokes joining and leaving independently (which is why the link is a standalone kind, exactly like AzurePrivateDnsZoneVirtualNetworkLink). Part of the DNS Private Resolver family (2186-2187) despite the out-of-run number -- enum numbers are pinned by the registry snapshot; never renumber.
- `AzureMonitorDataCollectionRuleAssociation` -- The attachment that puts ONE machine under an Azure Monitor data collection rule ({target_id}/providers/Microsoft.Insights/ dataCollectionRuleAssociations/{name}) -- an extension resource on the TARGET machine, many per rule, machines joining and leaving monitoring independently (which is why the association is a standalone kind, exactly like AzurePrivateDnsZoneVirtualNetworkLink). AzureVirtualMachine is a prerequisite because the smoke scenario attaches the fixture VM; the rule prerequisite chains the rule's own install manifest. Part of the Monitor family (2191-2192) despite the out-of-run number -- enum numbers are pinned by the registry snapshot; never renumber.
- `AzureEventgridDomainTopic` -- One named event stream inside an Azure Event Grid domain ({domain_id}/topics/{name}) -- the per-tenant mailbox of the multi-tenant pattern: many per domain, each with its own subscriptions and lifecycle, tenants joining and leaving without touching the domain (which is why the domain topic is a standalone kind, exactly like AzureEventHubConsumerGroup on a shared hub). Part of the Event Grid family (2193-2194) despite the out-of-run number -- enum numbers are pinned by the registry snapshot; never renumber.
- `AzureEventgridNamespaceTopic` -- One named CloudEvents stream inside an Azure Event Grid namespace ({namespace_id}/topics/{name}) -- many per namespace, publishers and teams creating and deleting their own against the shared namespace (which is why the topic is a standalone kind, exactly like AzureEventgridDomainTopic and AzureEventHubConsumerGroup). Part of the Event Grid family (2193-2197) despite the out-of-run number -- enum numbers are pinned by the registry snapshot; never renumber.
- `AzureMongoClusterUser` -- Grants one Microsoft Entra principal access to an Azure Cosmos DB for MongoDB vCore cluster ({cluster_id}/users/{object_id}) -- an access binding, not a password user: many per cluster, principals joining and leaving independently (which is why the grant is a standalone kind, the access-grant class of AzureRoleAssignment). Part of the Mongo vCore family (2211) despite the out-of-run number -- enum numbers are pinned by the registry snapshot; never renumber.
- `AzurePlantonRunner` -- AzureContainerAppEnvironment is a prerequisite because the runner appliance is a Container App, and every Container App runs inside an environment -- the environment reference must resolve before the appliance can deploy.
- `GcpArtifactRegistryRepo` -- 3000–3999: GCP resources
- `GcpTargetHttpsProxy` -- The URL map is the parent a proxy cannot exist without; the classic compute certificate kinds and the SSL policy are the fixture parents the committed scenarios attach. The Certificate Manager certificate list (certificate_manager_certificates, honored only by the cross-region internal ALB) is optional composition -- a scenario that arms it declares GcpCertManagerCert via the e2e-prerequisites annotation, never a registry edge that would tax every proxy and forwarding-rule chain.
- `GcpCloudFunction`
- `GcpCloudRun`
- `GcpCloudSql`
- `GcpDnsZone`
- `GcpGcsBucket`
- `GcpGkeCluster`
- `GcpIamCustomRole`
- `GcpProject`
- `GcpVpcNetwork`
- `GcpSubnetwork`
- `GcpRouterNat`
- `GcpGkeNodePool`
- `GcpServiceAccount`
- `GcpGkeWorkloadIdentityBinding`
- `GcpCertManagerCert`
- `GcpComputeInstance`
- `GcpDnsRecord`
- `GcpProjectIamMember`
- `GcpFirewallRule`
- `GcpGlobalAddress`
- `GcpCloudArmorPolicy`
- `GcpHealthCheck`
- `GcpBackendBucket`
- `GcpBackendService`
- `GcpRegionNetworkEndpointGroup`
- `GcpUrlMap`
- `GcpManagedSslCertificate`
- `GcpTargetHttpProxy`
- `GcpAlloydbCluster`
- `GcpRedisInstance`
- `GcpFirestoreDatabase`
- `GcpSpannerInstance`
- `GcpSpannerDatabase`
- `GcpBigtableInstance`
- `GcpMemorystoreInstance`
- `GcpCloudSqlDatabase`
- `GcpCloudSqlUser`
- `GcpAlloydbInstance`
- `GcpAlloydbUser`
- `GcpSpannerBackupSchedule`
- `GcpBigtableTable`
- `GcpFirestoreBackupSchedule`
- `GcpFirestoreIndex`
- `GcpBigQueryDataset`
- `GcpDataprocCluster`
- `GcpDataprocAutoscalingPolicy`
- `GcpBigQueryTable`
- `GcpBigQueryCapacityCommitment` -- A BigQuery capacity commitment: slots bought for a fixed term in an administration project and location, pooled across every reservation there. A purchase Google will not delete before its term ends.
- `GcpBigQueryReservationGroup` -- A BigQuery reservation group: reservations that share idle slots with each other first. Reservations reference it.
- `GcpDatastreamConnectionProfile` -- A Datastream connection profile: where one source database or destination is and how Datastream signs in. Streams reference a source and a destination profile; one profile serves many streams.
- `GcpDatastreamPrivateConnection` -- A Datastream private connection: the VPC peering or Private Service Connect interface through which Datastream reaches private databases, shared by every profile that reaches that network.
- `GcpPubSubTopic`
- `GcpPubSubSubscription`
- `GcpCloudTasksQueue`
- `GcpCloudSchedulerJob`
- `GcpPubSubSchema`
- `GcpPubSubTopicIamMember` -- One additive grant on a topic. Its own kind, not a field on the topic, because the identities that most need it (a logging sink's writer, a Security Command Center export's publisher) belong to resources that name the topic themselves; a grant on the topic that referenced them back would be a dependency cycle.
- `GcpVertexAiNotebook`
- `GcpVertexAiEndpoint`
- `GcpVertexAiIndex`
- `GcpVertexAiIndexEndpoint` -- Vector Search IndexEndpoint — distinct from the online-prediction GcpVertexAiEndpoint (671); different GCP resources, different kinds.
- `GcpVertexAiDeployedIndex`
- `GcpCloudComposerEnvironment`
- `GcpCloudComposerUserWorkloadsSecret`
- `GcpCloudComposerUserWorkloadsConfigMap`
- `GcpKmsKeyRing`
- `GcpKmsKey`
- `GcpKmsKeyIamMember`
- `GcpFilestoreInstance`
- `GcpWorkloadIdentityPool` -- 3101–3109: IAM/identity family (overflow block; the 3000–3022 foundation/security sub-band is fully allocated)
- `GcpWorkloadIdentityPoolProvider`
- `GcpServiceAccountIamMember`
- `GcpGcsBucketIamMember` -- One additive grant on a bucket, for a grantee that depends on the bucket itself (a logging sink writing into it): the bucket's own iam_members cannot reference such an identity without a cycle.
- `GcpGlobalForwardingRule` -- 3110–3119: networking/load-balancer family (overflow block; the 3023–3029 LB sub-band is fully allocated)
- `GcpSslPolicy`
- `GcpSslCertificate`
- `GcpServiceNetworkingConnection`
- `GcpAddress`
- `GcpServiceConnectionPolicy`
- `GcpCertManagerDnsAuthorization`
- `GcpCertificateMap` -- GcpCertManagerCert is a prerequisite because a map entry binds hostnames to EXISTING certificates — the canonical map references a certificate fixture's resource name.
- `GcpCertManagerTrustConfig` -- The CA certificates a load balancer validates client certificates against (mutual TLS). Its own kind: TLS policies reference it, never a certificate.
- `GcpCertManagerIssuanceConfig` -- How Google-managed certificates are issued from a private CA pool. Its own kind: many certificates share one config by name.
- `GcpCloudRunJob` -- 3120–3129: GCP serverless overflow
- `GcpServerlessVpcConnector`
- `GcpCloudRunWorkerPool` -- Cloud Run's no-ingress shape: a pool of always-running container instances (queue consumers, schedulers, background workers) that scales manually or by the owner's own signal instead of by requests. The proof deploys direct-VPC egress onto the prerequisite network.
- `GcpComputeDisk` -- 3130–3139: GCP compute overflow (the 3000–3022 foundation sub-band that holds GcpComputeInstance is fully allocated)
- `GcpComputeMig` -- GcpVpcNetwork is a prerequisite because the canonical group runs its fleet on a dedicated custom-mode VPC — a managed instance group's template must attach every VM to a network, and the default VPC is never assumed.
- `GcpMonitoringNotificationChannel` -- 3140–3149: GCP observability & log routing
- `GcpMonitoringAlertPolicy` -- GcpMonitoringNotificationChannel is a prerequisite because the policy's canonical shape references a channel to notify — a policy without a delivery endpoint measures but never pages.
- `GcpMonitoringUptimeCheck`
- `GcpLoggingSink` -- GcpGcsBucket is a prerequisite because the canonical sink exports to a Cloud Storage bucket — the cheapest destination that proves the whole writer-identity grant flow.
- `GcpMonitoringDashboard`
- `GcpMonitoringSlo`
- `GcpLogBucket`
- `GcpLogMetric`
- `GcpSecretManagerSecret` -- 3150–3159: GCP security & identity GcpServiceAccount is a prerequisite because the canonical secret grants secretAccessor to a workload service account — the access story the kind exists to model.
- `GcpIdentityPlatformConfig`
- `GcpIdentityPlatformTenant` -- GcpIdentityPlatformConfig is a prerequisite because tenants exist only in projects whose Identity Platform config enables multi_tenant.allow_tenants — a tenant without the initialized, tenant-enabled project config cannot be created at all.
- `GcpIamOauthClient`
- `GcpIamDenyPolicy`
- `GcpCloudRunDomainMapping` -- 3160–3169: GCP serverless edge GcpCloudRun is a prerequisite because a domain mapping exists only to point a verified domain at a running Cloud Run service — the route it maps must already exist for the mapping to be created at all.
- `GcpWorkflow`
- `GcpEventarcTrigger` -- GcpCloudRun is a prerequisite because the canonical trigger routes a Pub/Sub messagePublished event to a Cloud Run service — the destination story the kind exists to model.
- `GcpEventarcMessageBus`
- `GcpPlantonRunner`
- `GcpFolder` -- 3170–3179: GCP organization & governance (folders, org policies, tags, budgets, identity groups, API keys, KMS Autokey) GcpFolder is a container: the hierarchy node projects, sub-folders, policies, and tag bindings are placed inside.
- `GcpOrgPolicy`
- `GcpTagKey`
- `GcpTagValue`
- `GcpTagBinding`
- `GcpBillingBudget` -- A spending guardrail on a Cloud Billing account: amount, period, filters, thresholds, and where the alerts go. Lives on the billing account, so no project prerequisite; the proof lane needs a billing account the harness identity can administer.
- `GcpCloudIdentityGroup` -- A Google Group in Cloud Identity or Workspace with its memberships folded in -- the unit IAM bindings should name. Lives under a Cloud Identity customer, beside the service accounts and workload identity pools in the identity service group.
- `GcpApiKey`
- `GcpKmsAutokeyConfig` -- Cloud KMS Autokey switched on for a folder or a project: where the customer-managed keys GcpKmsKeyHandle requests are created. A security control, so it groups with the KMS kinds rather than the hierarchy.
- `GcpOrgPolicyCustomConstraint` -- A custom constraint is a DEFINITION the organization owns; the GcpOrgPolicy kinds that enforce it reference it by name, the way IAM bindings reference a custom role.
- `GcpSharedVpcHost` -- 3180–3189: GCP networking fabric (Shared VPC, VPC peering, HA VPN, firewall policies, PSC, network endpoint groups). HA VPN is two kinds: the gateway (with its Cloud Router) is declared once per VPC and region and referenced by every site connection, so two Google Cloud VPCs can point their connections at each other's gateway without a dependency cycle. Firewall policies are two kinds by scope: the hierarchical policy lives on the organization or a folder and is inherited by every network beneath its association; the network policy lives in a project and is attached to that project's VPC networks (globally or per region). Each folds its rules and associations -- a rule is keyed by priority inside its policy and an association is the edge that makes the policy act.
- `GcpSharedVpcServiceProject`
- `GcpVpcPeering`
- `GcpHaVpnGateway`
- `GcpHierarchicalFirewallPolicy`
- `GcpNetworkFirewallPolicy`
- `GcpHaVpnConnection`
- `GcpPscServiceAttachment` -- The producer half of Private Service Connect: publishes an internal load balancer's regional forwarding rule through NAT subnets so consumers in other VPCs reach it over a PSC endpoint (a regional GcpGlobalForwardingRule with an empty scheme targeting this attachment). The proof chain deploys the internal passthrough load balancer and the PSC NAT subnet as fixtures on the prerequisite network.
- `GcpNetworkEndpointGroup` -- Zonal (VM, hybrid, internet) and global (internet) network endpoint groups behind a `zone` selector; serverless, PSC, and regional internet groups stay in GcpRegionNetworkEndpointGroup. The zonal proof attaches a GcpComputeInstance fixture as an endpoint on the prerequisite network.
- `GcpRedisCluster` -- 3190–3199: GCP data (Memorystore for Redis Cluster, Managed Kafka, BigQuery connections and reservations, Datastream) Memorystore for Redis Cluster: the sharded, horizontally scaled Redis. Connectivity is Private Service Connect, placed by service connectivity automation through a GcpServiceConnectionPolicy for the gcp-memorystore-redis class on the network in the cluster's region -- the prerequisite; its proof pin carries that class beside the network and subnet pins.
- `GcpManagedKafkaCluster` -- Managed Service for Apache Kafka: a Google-operated broker fleet in one region, reachable from the VPC subnets it is attached to -- the prerequisite. Topics, ACLs, and Kafka Connect are their own kinds so the teams that own them declare them without editing the cluster.
- `GcpManagedKafkaTopic`
- `GcpManagedKafkaConnectCluster` -- Kafka Connect workers attached to a Kafka cluster: a separate project-and-location root that names the cluster it serves.
- `GcpBigQueryConnection` -- BigQuery's link to data outside its own storage (Cloud SQL, Spanner, AWS and Azure through Omni, Google resources through a managed service account, the Connector framework, Spark procedures).
- `GcpBigQueryReservation` -- A BigQuery slot reservation with the assignments that route projects, folders, or an organization onto it.
- `GcpDatastreamStream` -- A Datastream stream: continuous change data capture from one source database into BigQuery or Cloud Storage, through a source and a destination connection profile.
- `GcpRedisClusterEndpointSet` -- The Private Service Connect connections a consumer builds by hand (forwarding rules in other VPCs or projects) registered on a Redis Cluster, as one set: Google's resource replaces the cluster's whole user-created endpoint list in one write, so exactly one set per cluster is the honest grain. Its own kind because every connection names a forwarding rule that targets one of the cluster's service attachments -- a fold would depend on its own output.
- `GcpManagedKafkaAcl` -- The access rules for one resource pattern (a topic, a consumer group, a prefix, the cluster) on a Kafka cluster.
- `GcpManagedKafkaConnector` -- One data pipeline running on a Kafka Connect cluster.
- `GcpVertexAiAgentEngine` -- 3200–3229: GCP AI (Vertex AI agents and model deployments, RAG Engine, Vector Search, and the rest of the Vertex AI and generative-AI building blocks). The pre-existing Vertex AI kinds -- endpoint, index, index endpoint, deployed index, notebook -- live in the 3070s. Vertex AI Agent Engine: the managed runtime an AI agent runs in -- built from source or a container, hosted with its own identity and autoscaling, with an optional Memory Bank of long-term memories.
- `GcpVertexAiModelGardenDeployment` -- A Model Garden or Hugging Face model deployed to a Vertex AI endpoint in one step; every argument is immutable, so a change redeploys.
- `GcpVertexAiRagEngineConfig` -- The per-project, per-location tier of Vertex AI RAG Engine's managed vector database: a singleton Google owns, updated in place.
- `GcpVectorSearchCollection` -- 3203 is reserved for GcpVertexAiRagCorpus, forged when pulumi-gcp bridges provider 8.x (the resource has no SDK type today). A Vector Search collection -- a schema'd store of data objects with vector fields -- together with the approximate-nearest-neighbor indexes built over those fields (folded: one collection owns them).
- `GcpVertexAiSearchDataStore` -- Vertex AI Search (the Discovery Engine API behind the console's AI Applications / Gemini Enterprise): a data store is the corpus -- structured records, unstructured documents, or a public website -- with its schema, crawl patterns, and sitemaps folded in.
- `GcpVertexAiSearchEngine` -- The app over one or more data stores -- a search, chat, or recommendation engine -- with its serving controls, serving config, search widget, and assistants folded in. Its data stores are prerequisites: an engine cannot exist without one.
- `GcpVertexAiFeatureGroup` -- Vertex AI Feature Store: a feature group registers the features of some entities that live in a BigQuery table or view -- the features themselves folded in. Online stores serve them through feature views.
- `GcpVertexAiFeatureOnlineStore` -- The low-latency serving layer of Vertex AI Feature Store (Bigtable or Optimized storage), with the feature views it serves folded in.
- `GcpVertexAiDataset` -- A Vertex AI managed dataset: the registered container training, AutoML, labeling, and evaluation read their examples from.
- `GcpVertexAiTensorboard` -- A managed Vertex AI TensorBoard training jobs stream metrics into, with the experiments and runs declared in it folded in.
- `GcpVertexAiPersistentResource` -- A long-running cluster Vertex AI keeps provisioned so training jobs and Ray on Vertex AI start in seconds and scarce accelerators stay held between jobs.
- `GcpModelArmorTemplate` -- A Model Armor template: the named safety filters (prompt injection and jailbreak, Responsible AI content, sensitive data, malicious URLs) an AI application screens prompts and responses through.
- `GcpDocumentAiProcessor` -- A Document AI processor: a managed model that turns documents into structured data (OCR, forms, invoices, IDs), with its default version.
- `GcpColabRuntimeTemplate` -- A Colab Enterprise runtime template: the machine, network, image, and security settings every notebook runtime created from it gets.
- `GcpColabRuntime` -- A Colab Enterprise runtime: a notebook VM assigned to one user, built from a runtime template and started or stopped on purpose.
- `GcpColabSchedule` -- A Vertex AI schedule: a cron that launches a Colab Enterprise notebook run or a Vertex AI Pipelines run.
- `GcpTpuVm` -- A Cloud TPU VM: a slice of Google's AI accelerators with its host VMs. Beta-only in Google's provider (a recorded google-beta admission).
- `GcpDialogflowCxAgent` -- A Dialogflow CX conversational agent with the infrastructure its console-authored content uses folded in: webhooks, tools and their frozen versions, flow versions, the environments that pin them, and generative settings per language.
- `GcpVertexAiSearchDataConnector` -- A data connector is a COLLECTION of data stores Google syncs from a source (Jira, Confluence, ServiceNow, SharePoint, BigQuery, Google Drive, ...) on a schedule -- a different root from a data store, which is why it is its own kind. Engines search its stores by naming the collection.
- `GcpModelArmorFloorSetting` -- A Model Armor floor setting: the minimum safety screening a project, folder, or organization enforces on its templates and directly on Vertex AI and Google MCP server traffic. A different parent from a template (and a singleton Google never deletes), which is why it is its own kind.
- `GcpTpuQueuedResource` -- A Cloud TPU queued resource: a request that waits for TPU capacity and then provisions the nodes it describes. A different root from a TPU VM that owns many nodes, which is why it is its own kind. Beta-only in Google's provider (a recorded google-beta admission).
- `GcpDialogflowCxSecuritySettings` -- Dialogflow CX security settings: the redaction, retention, audio-export, and Insights-export policy agents in one project and location apply to their conversations. A different root from an agent, referenced by agents and shared among them, which is why it is its own kind.
- `GcpPrivateCaPool` -- 3230–3239: GCP security (Certificate Authority Service, Cloud KMS Autokey handles, Security Command Center, Binary Authorization) A Certificate Authority Service CA pool: the trust anchor and issuance policy its certificate authorities and certificates live inside.
- `GcpKmsKeyHandle` -- An Autokey key handle: asks Autokey for a customer-managed key for one resource type in one project and location; the resource it protects names the key the handle returns.
- `GcpSccNotificationConfig` -- Security Command Center streaming notifications to Pub/Sub at a project, folder, or organization.
- `GcpSccMuteConfig` -- A Security Command Center mute rule at a project, folder, or organization.
- `GcpSccBigQueryExport` -- A continuous Security Command Center findings export to a BigQuery dataset at a project, folder, or organization.
- `GcpBinaryAuthorizationPolicy` -- A project's Binary Authorization policy: which container images GKE admits, per cluster. A project singleton.
- `GcpBinaryAuthorizationAttestor` -- A Binary Authorization attestor with its Artifact Analysis note: the public keys that verify image attestations a policy requires.
- `GcpPrivateCaCertificateAuthority` -- A certificate authority in a CA pool: a self-signed root, or a subordinate signed by another authority or an outside CA. Its own kind because a pool rotates through several and a subordinate references its parent.
- `GcpPrivateCaCertificateTemplate` -- A certificate template: a reusable certificate shape in a project and location that certificates in any pool there reference.
- `GcpPrivateCaCertificate` -- A certificate issued from a CA pool for a key its owner holds; destroy revokes it.
- `GcpGkeFleet` -- 3240–3249: GCP platform engineering (GKE fleets, images, Cloud Build, Cloud Deploy) GcpGkeFleet is the container a fleet's scopes, namespaces, memberships, and features live in: the project's one fleet is the room, and a team scope or a fleet-wide feature is what is placed inside it.
- `GcpGkeFleetFeature` -- The three fleet children name GcpGkeFleet as their prerequisite: Google requires the fleet before a scope, a fleet declared after a cluster registers collides with the fleet that registration created implicitly, and a feature configures the fleet it lives in. A chart that references the fleet's project_id output orders each child after it.
- `GcpGkeFleetScope`
- `GcpGkeFleetMembership`
- `GcpComputeImage` -- A Compute Engine custom image: the golden boot image VMs, instance templates, and disks start from, rolled forward through image families.
- `GcpCloudBuildConnection` -- GcpCloudBuildConnection is the container a code host's repositories live in: the connection to GitHub, GitLab, or Bitbucket is the room, and each linked repository is what is placed inside it.
- `GcpCloudBuildRepository` -- A repository is created under its connection and cannot exist without it.
- `GcpCloudBuildTrigger` -- A Cloud Build trigger: what starts a build (a code event, a Pub/Sub message, a webhook, or a manual run) and what the build does.
- `GcpDeliveryPipeline` -- A Cloud Deploy delivery pipeline: the ordered stages a release is promoted through, with the automations that drive it.
- `GcpDeployTarget` -- A Cloud Deploy target: where a pipeline stage deploys (a GKE cluster, a Cloud Run location, a fleet cluster, several targets at once, or a custom target).
- `GcpFirebaseProject` -- 3250–3259: GCP Firebase (project enablement, app registrations, and the Firebase-adjacent products that follow) GcpFirebaseProject is the container the app registrations live in: "Firebase on this project" is the room, the Android/Apple/Web apps are what is placed inside it.
- `GcpFirebaseAndroidApp` -- The three app registrations exist only inside a Firebase-enabled project, so each names GcpFirebaseProject as its prerequisite: the E2E harness deploys the enablement first, and a chart that references the enablement's project_id output orders the registration after it.
- `GcpFirebaseAppleApp`
- `GcpFirebaseWebApp`
- `GcpCloudBuildWorkerPool` -- 3260–3269: GCP platform engineering, continued (Cloud Build and Cloud Deploy resources many pipelines share) A private Cloud Build worker pool: build machines many triggers and Cloud Deploy targets share, optionally on a private network.
- `GcpDeployPolicy` -- A Cloud Deploy deploy policy: rollout restrictions (freeze windows) that apply to every pipeline and target its selectors match.
- `GcpDeployCustomTargetType` -- A Cloud Deploy custom target type: how to render and deploy to a target Google does not deploy natively; many targets share one.
- `KubernetesNamespace` -- 4000–4999: Kubernetes resources, organized in family sub-bands (4030–4069 also hosts CNI/autoscaling/DR addons; 4130–4149 hosts analytics & ML; 4180–4199 hosts agent sandbox runtimes; 4200–4209 hosts AI agent platforms) 4000–4029: Kubernetes building blocks (core API primitives)
- `KubernetesDeployment`
- `KubernetesStatefulSet`
- `KubernetesDaemonSet`
- `KubernetesJob`
- `KubernetesCronJob`
- `KubernetesService`
- `KubernetesSecret`
- `KubernetesManifest`
- `KubernetesHelmRelease`
- `KubernetesConfigMap`
- `KubernetesServiceAccount`
- `KubernetesRbac` -- Bundles the RBAC grant grain (Role/ClusterRole + its binding) into one kind: "grant these permissions to these subjects in this scope".
- `KubernetesIngress`
- `KubernetesNetworkPolicy`
- `KubernetesPersistentVolumeClaim`
- `KubernetesStorageClass`
- `KubernetesResourceQuota` -- Manages the namespace-governance pair: the ResourceQuota plus an optional companion LimitRange (per-object defaults/bounds) — two API objects, one governance story.
- `KubernetesPriorityClass`
- `KubernetesPodDisruptionBudget`
- `KubernetesHorizontalPodAutoscaler`
- `KubernetesCertManager` -- 4030–4069: Kubernetes foundation addons (certs, DNS, secrets, ingress, Gateway API, mesh, CNI/autoscaling/DR)
- `KubernetesClusterIssuer` -- KubernetesCertManager is a prerequisite for the three cert-manager CR kinds below: ClusterIssuer/Issuer/Certificate are cert-manager custom resources — without the controller and its CRDs they cannot be applied.
- `KubernetesIssuer`
- `KubernetesCertificate`
- `KubernetesExternalDns`
- `KubernetesExternalSecretsOperator`
- `KubernetesClusterSecretStore` -- KubernetesExternalSecretsOperator is a prerequisite for the three external-secrets CR kinds below: ClusterSecretStore/SecretStore/ ExternalSecret are external-secrets custom resources — without the operator and its CRDs they cannot be applied.
- `KubernetesSecretStore`
- `KubernetesExternalSecret`
- `KubernetesIngressNginx`
- `KubernetesGatewayApiCrds`
- `KubernetesGatewayClass`
- `KubernetesGateway`
- `KubernetesListenerSet`
- `KubernetesHttpRoute`
- `KubernetesGrpcRoute`
- `KubernetesTcpRoute`
- `KubernetesUdpRoute`
- `KubernetesTlsRoute`
- `KubernetesReferenceGrant`
- `KubernetesBackendTlsPolicy`
- `KubernetesIstioBaseCrds`
- `KubernetesIstio`
- `KubernetesDestinationRule` -- Istio API kinds (mesh traffic policy, security, telemetry). The seven typed resources below (4053–4059) require the Istio CRDs on the cluster, provided by the lightweight CRDs-only KubernetesIstioBaseCrds (851) — NOT the full mesh KubernetesIstio (852).
- `KubernetesServiceEntry`
- `KubernetesPeerAuthentication`
- `KubernetesRequestAuthentication`
- `KubernetesAuthorizationPolicy`
- `KubernetesTelemetry`
- `KubernetesEnvoyFilter`
- `KubernetesMetricsServer`
- `KubernetesCilium`
- `KubernetesKeda`
- `KubernetesKarpenter`
- `KubernetesKarpenterNodePool`
- `KubernetesKarpenterEc2NodeClass`
- `KubernetesClusterAutoscaler`
- `KubernetesVelero`
- `KubernetesGkeComputeClass`
- `KubernetesKubePrometheusStack` -- 4070–4079: Kubernetes observability
- `KubernetesGrafana`
- `KubernetesSignoz` -- KubernetesClickHouse is a prerequisite because SigNoz stores every trace, metric and log in ClickHouse and deploys none of its own — the telemetry store is composed, never bundled.
- `KubernetesLoki`
- `KubernetesTempo`
- `KubernetesOtelOperator` -- The operator's admission webhooks (failurePolicy Fail) are served with a cert-manager Certificate in the default posture — cert-manager must be running before the operator installs.
- `KubernetesOtelCollector`
- `KubernetesServiceMonitor` -- KubernetesKubePrometheusStack is a prerequisite because it installs the monitoring.coreos.com CRDs the ServiceMonitor is an instance of, and the Prometheus that scrapes through it.
- `KubernetesPodMonitor` -- KubernetesKubePrometheusStack is a prerequisite because it installs the monitoring.coreos.com CRDs the PodMonitor is an instance of, and the Prometheus that scrapes through it.
- `KubernetesPrometheusRule` -- KubernetesKubePrometheusStack is a prerequisite because it installs the monitoring.coreos.com CRDs the PrometheusRule is an instance of, and the Prometheus that evaluates it.
- `KubernetesKyverno` -- 4080–4099: Kubernetes security, policy, and identity
- `KubernetesGatekeeper`
- `KubernetesKeycloak` -- Keycloak declarations compose the official Keycloak Operator (which reconciles the Keycloak CR this kind renders) and, on the recommended postgres vendor, a KubernetesPostgres database — both must resolve before the CR can converge.
- `KubernetesOpenBao`
- `KubernetesOpenFga` -- OpenFGA requires a datastore; the recommended arm composes a KubernetesPostgres database (the sandbox memory arm needs nothing, but the registry declares the shape real deployments require).
- `KubernetesKeycloakOperator`
- `KubernetesCloudNativePgOperator` -- 4100–4129: Kubernetes data platforms
- `KubernetesPostgres`
- `KubernetesValkey`
- `KubernetesPerconaMysqlOperator`
- `KubernetesMysql`
- `KubernetesPerconaMongoOperator`
- `KubernetesMongodb`
- `KubernetesStrimziKafkaOperator`
- `KubernetesKafka` -- container_kind: a Strimzi Kafka cluster is a place in the provider's own model — KafkaTopic and KafkaUser declarations BELONG to one cluster (the strimzi.io/cluster label) and are drawn inside its box. Clients that merely talk to the cluster (Connect, MirrorMaker2, UI, Karapace) carry containment_exempt on their bootstrap/trust references.
- `KubernetesKafkaTopic`
- `KubernetesKafkaUser`
- `KubernetesKafkaConnect` -- container_kind: a Connect cluster hosts the connectors deployed INTO it (KafkaConnector's strimzi.io/cluster label names its Connect cluster) — the same room shape as KubernetesKafka above.
- `KubernetesKafkaConnector`
- `KubernetesKafkaMirrorMaker2`
- `KubernetesKarapace`
- `KubernetesKafkaUi`
- `KubernetesOpenSearchOperator`
- `KubernetesOpenSearch`
- `KubernetesAltinityOperator`
- `KubernetesClickHouse`
- `KubernetesSolrOperator`
- `KubernetesSolr`
- `KubernetesNeo4j`
- `KubernetesSeaweedFs`
- `KubernetesQdrant`
- `KubernetesRabbitMqOperator` -- The RabbitMQ Cluster Operator's release manifest ships admission webhooks whose serving certificate is a cert-manager Certificate — cert-manager must be running before the operator installs.
- `KubernetesRabbitMq`
- `KubernetesCnpgBarmanCloudPlugin` -- The Barman Cloud plugin is CloudNativePG's object-store backup path, installed as its own Helm release beside the operator. It registers with the operator over CNPG-I and issues its TLS through cert-manager, so both are prerequisites: nothing to register with, or no certificates, and the release never becomes ready.
- `KubernetesAirflow` -- 4130–4149: Kubernetes analytics and ML KubernetesPostgres is a prerequisite because Airflow's metadata database composes a KubernetesPostgres by default (the spec's FK defaults resolve onto its outputs) and the migration Job needs the database reachable before the server components start.
- `KubernetesSparkOperator`
- `KubernetesKubeRayOperator`
- `KubernetesRayCluster` -- KubernetesKubeRayOperator is a prerequisite because this kind declares the RayCluster custom resource that only the operator's CRDs admit and only the operator reconciles into head and worker pods.
- `KubernetesFlinkOperator` -- KubernetesCertManager is a prerequisite because the Flink operator's chart, with its default-on admission webhook, renders cert-manager Issuer/Certificate resources and trusts the API server through cert-manager's CA injection — there is no self-signed fallback at the pinned chart, and the webhooks are fail-closed.
- `KubernetesFlinkDeployment` -- KubernetesFlinkOperator is a prerequisite because this kind declares the FlinkDeployment custom resource that only the operator's CRDs admit and only the operator reconciles into a running Flink cluster.
- `KubernetesJupyterHub` -- KubernetesPostgres is a prerequisite because JupyterHub's hub database composes a KubernetesPostgres in its external-database arm (the spec's FK defaults resolve onto its outputs) and the hub pod mounts that database's credential Secret before it can start.
- `KubernetesMlflow` -- KubernetesPostgres is a prerequisite because MLflow's backend store composes a KubernetesPostgres in its production arm (FK defaults onto its outputs; the module composes the connection URI from its credential Secret), and KubernetesSeaweedFs because the artifact store's S3-compatible arm FK-defaults onto the SeaweedFS endpoint and credential Secret.
- `KubernetesTrino` -- KubernetesPostgres is a prerequisite because Trino's postgres catalogs compose a KubernetesPostgres (the catalog host and credential FK-default onto its outputs), and the pods read that database's credential Secret to resolve catalog passwords from environment.
- `KubernetesSuperset` -- KubernetesPostgres is a prerequisite because Superset's REQUIRED metadata database composes a KubernetesPostgres (FK defaults onto its outputs; the module composes the environment Secret from its credential Secret), and KubernetesValkey because the cache/broker arm FK-defaults onto a KubernetesValkey's service and password Secret.
- `KubernetesLangfuse` -- Langfuse's four required stores each FK-default onto a catalog kind: KubernetesPostgres (users, projects, prompts, API keys - host and password Secret), KubernetesClickHouse (traces and scores - host, cluster name and auth Secret), KubernetesValkey (the ingestion queues - host, ACL user and password Secret), and KubernetesSeaweedFs (the S3-compatible event store - endpoint and credentials Secret).
- `KubernetesArgocd` -- 4150–4169: Kubernetes GitOps and CI/CD
- `KubernetesArgoWorkflows`
- `KubernetesTektonOperator`
- `KubernetesTekton` -- KubernetesTektonOperator is a prerequisite because this kind declares the TektonConfig custom resource that only the operator's CRDs admit and only the operator reconciles into running components.
- `KubernetesGhaRunnerScaleSetController`
- `KubernetesGhaRunnerScaleSet` -- KubernetesGhaRunnerScaleSetController is a prerequisite because this kind renders an AutoscalingRunnerSet custom resource that only the controller's CRDs admit and only the controller reconciles into listener and runner pods.
- `KubernetesHarbor`
- `KubernetesJenkins`
- `KubernetesTemporal` -- 4170–4179: Kubernetes app platforms KubernetesPostgres is a prerequisite because the recommended (and E2E-proven) database composition backs Temporal's default and visibility stores with a CloudNativePG cluster.
- `KubernetesNats`
- `KubernetesLocust`
- `KubernetesPlantonRunner`
- `KubernetesPlantonOperator`
- `KubernetesPlantonPlatform` -- KubernetesPlantonOperator is a prerequisite because this kind declares the PlantonPlatform custom resource that only the operator's CRD admits and only the operator reconciles into a running platform.
- `KubernetesGoFeatureFlag`
- `KubernetesFlagd`
- `KubernetesGoFeatureFlagFlagFile`
- `KubernetesFlagdFlagFile`
- `KubernetesAgentSandboxController` -- 4180–4199: agent sandbox runtimes (4180–4189 Kubernetes SIG Apps' agent-sandbox; 4190–4199 Agent Substrate)
- `KubernetesAgentSandbox` -- KubernetesAgentSandboxController is a prerequisite because it installs the agents.x-k8s.io definition the Sandbox is an instance of, and the controller that runs its pod.
- `KubernetesAgentSandboxTemplate` -- KubernetesAgentSandboxController is a prerequisite because it installs the extensions.agents.x-k8s.io definition the SandboxTemplate is an instance of, and the extension controllers that turn it into network policy and warm pool members.
- `KubernetesAgentSandboxWarmPool` -- KubernetesAgentSandboxController is a prerequisite because it installs the extensions.agents.x-k8s.io definition the SandboxWarmPool is an instance of, and the controller that keeps the pool's members ready.
- `KubernetesAgentSandboxClaim` -- KubernetesAgentSandboxController is a prerequisite because it installs the extensions.agents.x-k8s.io definition the SandboxClaim is an instance of, and the controller that binds it to a warm pool member.
- `KubernetesAgentSubstrate`
- `KubernetesAgentSubstrateWorkerPool` -- KubernetesAgentSubstrate is a prerequisite because it installs the ate.dev definition the WorkerPool is an instance of, the controller that turns it into worker pods, and the node agent and API server the workers register with.
- `KubernetesAgentSubstrateSandboxConfig` -- KubernetesAgentSubstrate is a prerequisite because it installs the ate.dev definition the SandboxConfig is an instance of, the admission policy that holds each class to its assets, and the API server and node agent that read it.
- `KubernetesAgentSubstrateCsiDriverConfig` -- KubernetesAgentSubstrate is a prerequisite because it installs the ate.dev definition the CSIDriverConfig is an instance of, and the API server and node agent that look a driver up in it.
- `KubernetesStigmer` -- 4200–4209: Kubernetes AI agent platforms KubernetesStigmer installs the Stigmer operator and declares the one Stigmer it reconciles, so it carries no prerequisite.
- `DigitalOceanApp` -- 5000–5999: DigitalOcean resources
- `DigitalOceanBucket`
- `DigitalOceanContainerRegistry`
- `DigitalOceanDatabaseCluster` -- A DigitalOcean managed database cluster -- the cluster every other managed-database resource is created on: users, logical databases, connection pools, the firewall rule set, Kafka topics, and Kafka schema subjects are all API-addressed under a cluster's id and cannot exist before it. A container kind for exactly that reason: those six kinds are children of the cluster, so on a diagram the cluster is the room they stand in. A read replica also names its primary but is a cluster of its own, a monitor alert names the clusters it watches, and an App Platform app names a cluster it attaches; those three references are containment-exempt on their specs.
- `DigitalOceanDnsZone`
- `DigitalOceanDroplet` -- No VPC prerequisite: the droplet spec's vpc reference is optional — an omitted vpc places the droplet in the region's default VPC.
- `DigitalOceanFirewall`
- `DigitalOceanFunction`
- `DigitalOceanKubernetesCluster` -- DigitalOceanVpc is a prerequisite because the cluster spec's vpc reference is required: the control plane and every node pool live inside one VPC, resolved to the DigitalOceanVpc's exported vpc_id output.
- `DigitalOceanKubernetesNodePool` -- DigitalOceanKubernetesCluster is a prerequisite because a node pool is API-addressed under its owning cluster: the spec's cluster reference is required and the pool cannot exist first.
- `DigitalOceanLoadBalancer` -- No registry prerequisite: the load balancer's vpc reference is optional (DigitalOcean places it in the region's default VPC when unset, and GLOBAL balancers take no VPC at all). Scenarios that exercise VPC placement declare it per-scenario via the e2e-prerequisites annotation.
- `DigitalOceanVolume`
- `DigitalOceanVpc`
- `DigitalOceanCertificate`
- `DigitalOceanDnsRecord` -- DigitalOceanDnsZone is a prerequisite because a record is API-addressed under its domain: the spec's domain reference is required, resolved to the DigitalOceanDnsZone's exported zone_name output.
- `DigitalOceanDatabaseUser` -- An additional user on a managed database cluster: the cluster reference is required, resolved to the DigitalOceanDatabaseCluster's exported cluster_id output.
- `DigitalOceanDatabaseDb` -- An additional logical database on a managed database cluster: the cluster reference is required.
- `DigitalOceanDatabaseConnectionPool` -- A PgBouncer connection pool on a managed PostgreSQL cluster: the cluster reference is required.
- `DigitalOceanDatabaseFirewall` -- The inbound trusted-sources rule set of a managed database cluster: the cluster reference is required.
- `DigitalOceanDatabaseReplica` -- A read-only replica of a managed database cluster: the primary cluster reference is required.
- `DigitalOceanDatabaseKafkaTopic` -- A topic on a managed Kafka cluster with the full per-topic configuration block; the owning cluster reference is required.
- `DigitalOceanDatabaseKafkaSchema` -- One schema subject registered in a managed Kafka cluster's schema registry; the owning cluster reference is required.
- `DigitalOceanProject` -- The account-level organizational container; membership is carried on the project itself as resource URNs.
- `DigitalOceanSshKey` -- An SSH public key registered on the account, referenced by droplets and droplet autoscale pools at create time.
- `DigitalOceanMonitorAlert` -- An alert policy on DigitalOcean's built-in metrics for droplets, load balancers, and managed database clusters. All entity targeting is optional, so there is no registry prerequisite.
- `DigitalOceanUptimeCheck` -- An availability/latency probe on an external endpoint with composed alert rules; the target is outside the account, so there is no registry prerequisite.
- `DigitalOceanReservedIp` -- A static public IP address (IPv4 or IPv6) reserved in a region and optionally assigned to a droplet. The droplet attachment is an optional composition seam, so there is no registry prerequisite.
- `DigitalOceanVpcPeering` -- A private-network peering connection between exactly two VPCs; both VPC references are required.
- `DigitalOceanSpacesKey` -- An access-key pair for Spaces object storage. Bucket grants are an optional composition seam, so there is no registry prerequisite.
- `DigitalOceanCdn` -- A CDN endpoint serving a Spaces bucket's content from the global edge: the origin reference is required, resolved to the DigitalOceanBucket's exported bucket_domain_name output.
- `DigitalOceanDropletAutoscalePool` -- A pool of identical droplets DigitalOcean keeps at a fixed size or scales on utilization. The template's ssh_keys reference is required (the API mandates SSH keys), resolved to the DigitalOceanSshKey's exported ssh_key_id output.
- `CloudflareDnsZone` -- 7000–7999: Cloudflare resources
- `CloudflareKvNamespace`
- `CloudflareR2Bucket`
- `CloudflareWorker`
- `CloudflareLoadBalancer` -- CloudflareDnsZone and CloudflareLoadBalancerPool are prerequisites because a load balancer is a DNS-level construct inside a zone (the spec's zone_id reference must resolve) and traffic must land somewhere (the required fallback_pool reference must resolve to a live pool).
- `CloudflareD1Database`
- `CloudflareZeroTrustAccessApplication`
- `CloudflareDnsRecord` -- CloudflareDnsZone is a prerequisite because every record lives inside a zone -- the spec's zone_id reference must resolve before the record can be created.
- `CloudflareRuleset` -- CloudflareDnsZone is a prerequisite because zone-scoped rulesets (the common case; the spec's zone_id reference defaults to the zone kind) must resolve their zone first. Account-scoped rulesets simply leave the reference unused.
- `CloudflareWorkersKvPair` -- CloudflareKvNamespace is a prerequisite because a KV pair is written into a namespace -- the spec's namespace_id reference must resolve first.
- `CloudflareHyperdriveConfig`
- `CloudflareLoadBalancerPool`
- `CloudflareLoadBalancerMonitor`
- `CloudflareZeroTrustAccessPolicy`
- `CloudflareZeroTrustAccessGroup`
- `CloudflareQueue`
- `CloudflarePagesProject`
- `CloudflareZeroTrustTunnel`
- `CloudflareZeroTrustTunnelVirtualNetwork`
- `CloudflareZeroTrustTunnelRoute` -- CloudflareZeroTrustTunnel is a prerequisite because a route steers a CIDR through an existing tunnel -- the spec's tunnel_id reference must resolve first. (The optional virtual_network_id is scenario-declared, not a registry prerequisite.)
- `CloudflareList`
- `CloudflareListItem` -- CloudflareList is a prerequisite because an item exists only inside a list -- the spec's list_id reference must resolve first.
- `CloudflareTurnstileWidget`
- `CloudflareEmailRoutingZone` -- CloudflareDnsZone is a prerequisite because email routing is enabled ON a zone -- the spec's zone_id reference must resolve first.
- `CloudflareEmailRoutingRule` -- CloudflareDnsZone is a prerequisite because a routing rule lives in a zone's email routing configuration -- the spec's zone_id reference must resolve first. CloudflareEmailRoutingZone is deliberately NOT a prerequisite: the API accepts rule creation on a zone whose Email Routing was never enabled (measured live 2026-08-26 -- POST zones/{id}/email/routing/rules answered 200 on a fresh zone with routing status "unconfigured"). Enablement gates mail FLOW, not rule lifecycle, so it is an operational concern the kind docs teach, not a create-time dependency. (Forward destinations reference CloudflareEmailRoutingAddress only for forward-type rules, so that edge is scenario-declared.)
- `CloudflareEmailRoutingAddress`
- `CloudflareOriginCaCertificate`
- `CloudflareCertificatePack` -- CloudflareDnsZone is a prerequisite because a certificate pack is ordered for a zone's hostnames -- the spec's zone_id reference must resolve first.
- `CloudflareCustomHostname` -- CloudflareDnsZone is a prerequisite because a custom hostname (SSL for SaaS) is provisioned inside a zone -- the spec's zone_id reference must resolve first.
- `CloudflareCustomHostnameFallbackOrigin` -- CloudflareDnsZone is a prerequisite because the fallback origin is a zone-level SSL-for-SaaS setting -- the spec's zone_id reference must resolve first.
- `CloudflareZeroTrustAccessIdentityProvider` -- No prerequisites: identity providers are account-scoped in the canonical case (the optional zone scope is a per-manifest choice, not a structural dependency).
- `CloudflareZeroTrustAccessServiceToken` -- No prerequisites: service tokens are account-scoped in the canonical case (the optional zone scope is a per-manifest choice, not a structural dependency).
- `CloudflareZeroTrustOrganization` -- No prerequisites: the organization is an account-scoped configuration singleton (the optional zone scope is a per-manifest choice, not a structural dependency).
- `CloudflareZeroTrustAccessInfrastructureTarget` -- No prerequisites: targets are account-scoped, and the virtual-network reference is an optional per-manifest edge (omitted = the account's default virtual network).
- `CloudflareZeroTrustMcpPortal` -- No prerequisites: portals are account-scoped, and the servers[] rows' MCP-server references are optional per-manifest edges.
- `CloudflareZeroTrustMcpServer` -- No prerequisites: MCP server registrations are account-scoped and self-contained -- portals reference them, not the reverse.
- `CloudflareZeroTrustGatewayPolicy` -- No prerequisites: Gateway policies are account-scoped, and their list / virtual-network references are optional per-manifest edges.
- `CloudflareZeroTrustList` -- No prerequisites: Zero Trust lists are account-scoped and self-contained.
- `CloudflareZeroTrustGatewaySettings` -- No prerequisites: the Gateway configuration is an account-scoped singleton, and its certificate reference is an optional per-manifest edge.
- `CloudflareZeroTrustDnsLocation` -- No prerequisites: DNS locations are account-scoped and self-contained.
- `CloudflareZeroTrustDeviceDefaultProfile` -- No prerequisites: the default device profile is an account-scoped configuration singleton; its virtual-network and zone-certificate references are optional per-manifest edges.
- `CloudflareZeroTrustDeviceCustomProfile` -- No prerequisites: custom device profiles are account-scoped, and the virtual-network reference is an optional per-manifest edge.
- `CloudflareZeroTrustDevicePostureRule` -- No prerequisites: posture rules are account-scoped and self-contained (list and integration references are literal UUIDs today).
- `CloudflareZoneTlsSettings` -- CloudflareDnsZone is a prerequisite because TLS settings are zone-scoped configuration -- the spec's zone_id reference must resolve first.
- `CloudflareCustomSslCertificate` -- CloudflareDnsZone is a prerequisite because a custom certificate is uploaded to an existing zone -- the spec's zone_id reference must resolve first.
- `CloudflareMtlsCertificate` -- No prerequisites: mTLS certificates are account-scoped uploads and self-contained -- consumers (zone TLS CA associations, Authenticated Origin Pulls rows, Workers mTLS bindings) reference them, not the reverse.
- `CloudflareAuthenticatedOriginPulls` -- CloudflareDnsZone is a prerequisite because Authenticated Origin Pulls enablement configures an existing zone -- the spec's zone_id reference must resolve first. The per-hostname certificate edge is optional and scenario-declared, never a registry prerequisite.
- `CloudflareAuthenticatedOriginPullsCertificate` -- CloudflareDnsZone is a prerequisite because the client certificate is uploaded to an existing zone -- the spec's zone_id reference must resolve first.
- `CloudflareZoneSettings` -- CloudflareDnsZone is a prerequisite because zone settings configure an existing zone -- the spec's zone_id reference must resolve first.
- `CloudflareCacheSettings` -- CloudflareDnsZone is a prerequisite because cache settings configure an existing zone -- the spec's zone_id reference must resolve first.
- `CloudflareIpAccessRule` -- No prerequisites: IP Access rules are account-scoped in the canonical case (the zone scope is a per-manifest choice, not a structural dependency).
- `CloudflareBotManagement` -- CloudflareDnsZone is a prerequisite because Bot Management is zone-singleton configuration -- the spec's zone_id reference must resolve first.
- `CloudflareSnippet` -- CloudflareDnsZone is a prerequisite because snippets deploy to a zone -- the spec's zone_id reference must resolve first.
- `CloudflareSnippetRules` -- CloudflareDnsZone and CloudflareSnippet are prerequisites: the rules table is zone-scoped and every rule invokes a snippet by name.
- `CloudflareWaitingRoom` -- CloudflareDnsZone is a prerequisite because waiting rooms sit on a zone's host+path -- the spec's zone_id reference must resolve first.
- `CloudflareWaitingRoomEvent` -- CloudflareWaitingRoom is a prerequisite because events run on a room (and the room's own chain brings the zone).
- `CloudflareLogpushJob` -- No prerequisites: logpush jobs are dual-scope (account or zone) and the zone reference is optional -- zone-scoped lanes declare the edge at the scenario level.
- `CloudflareNotificationPolicy` -- No prerequisites: every delivery mechanism (email, PagerDuty, webhook) is optional -- policies referencing a webhook declare the edge at the scenario level.
- `CloudflareNotificationWebhook` -- No prerequisites: a webhook destination is account-scoped and self-contained -- notification policies reference it, not the reverse.
- `CloudflareWebAnalyticsSite` -- No prerequisites: a site is identified by host OR zone, and the zone reference is optional -- zone-measured lanes declare the edge at the scenario level.
- `CloudflareWorkflow` -- CloudflareWorker is a prerequisite because a workflow registers a class exported by a DEPLOYED Worker script -- the spec's script_name reference must resolve first.
- `CloudflareSecretsStore` -- No prerequisites: the Secrets Store is an account-scoped container and self-contained -- consumers (store secrets, Worker bindings, AI Gateway authentication) reference it, not the reverse.
- `CloudflareSecretsStoreSecret` -- CloudflareSecretsStore is a prerequisite because every secret lives inside a store -- the spec's store_id reference must resolve first.
- `CloudflareAiGateway` -- No prerequisites: the gateway is account-scoped and self-contained; its optional Secrets Store link (BYO provider keys) is a scenario-level composition, not a structural requirement.
- `CloudflareAccountApiToken` -- No prerequisites: an account-owned API token is self-contained; the resources its policies cover are identifiers, not references.
- `CloudflareHealthcheck` -- CloudflareDnsZone is a prerequisite because standalone health checks are zone-scoped -- the spec's zone_id reference must resolve first.
- `Auth0Connection` -- 8000–8999: Auth0 resources
- `Auth0Client`
- `Auth0EventStream`
- `Auth0ResourceServer`
- `Auth0Action`
- `Auth0Role`
- `Auth0User`
- `Auth0TenantSettings`
- `Auth0CustomDomain`
- `Auth0CustomDomainVerification`
- `Auth0Branding` -- 8010–8015: how the tenant's Universal Login looks, reads and mails -- branding and theme, the login flow, each prompt's words and screen partials, the email service and each email.
- `Auth0Prompt`
- `Auth0PromptCustomText`
- `Auth0PromptScreenPartials`
- `Auth0EmailProvider`
- `Auth0EmailTemplate`
- `Auth0ClientFromMetadataDocument`
- `OpenFgaStore` -- 9000–9999: OpenFGA resources OpenFGA publishes a Terraform provider and no Pulumi provider, so its kinds ship one HCL module and declare the engines that run it; the CLI and the platform refuse Pulumi for them before anything runs.
- `OpenFgaAuthorizationModel`
- `OpenFgaRelationshipTuple`
- `StripeWebhookEndpoint` -- 10000–10999: Stripe resources Stripe publishes a Terraform provider and no Pulumi provider, and its kinds are proven on OpenTofu alone, so each declares ["tofu"]: one HCL module, and every other engine refused before anything runs. Family sub-bands: 10000 account and delivery, 10100 catalog and pricing.
- `StripeEventDestination`
- `StripeBillingPortalConfiguration`
- `StripePaymentMethodConfiguration`
- `StripePaymentMethodDomain`
- `StripeRadarValueList`
- `StripeProduct`
- `StripePrice` -- A price always belongs to a product (its product field is required), so the product deploys first.
- `StripeCoupon`
- `StripePromotionCode` -- A promotion code always redeems a coupon (its coupon field is required), so the coupon deploys first.
- `StripeEntitlementFeature`
- `StripeShippingRate`
- `StripeTaxRate` -- Tax rate and tax registration share the initials "tr", so each takes Stripe's own id prefix with its vowels dropped (txr_, taxreg_).
- `StripeTaxRegistration`
- `StripeBillingMeter`
- `StripePaymentLink` -- A payment link always sells at least one price (line_items is required), so the price deploys first.

### spec.server.extraEnv.secrets[].valueFrom.env

`string`

### spec.server.extraEnv.secrets[].valueFrom.name

`string` · required

- rule: {"required":true}

### spec.server.extraEnv.secrets[].valueFrom.fieldPath

`string`

### spec.server.extraEnv.envFrom

`[]EnvFromSource`

Bulk import of environment variables from ConfigMaps or Secrets.

### spec.server.extraEnv.envFrom[].prefix

`string`

Optional prefix prepended to each imported key name.
For example, prefix "APP_" with key "PORT" produces env var "APP_PORT".

### spec.server.extraEnv.envFrom[].configMapRef

`ConfigMapRef`

Import all keys from a ConfigMap.

### spec.server.extraEnv.envFrom[].configMapRef.name

`string` · required

Name of the ConfigMap.

- rule: {"required":true}

### spec.server.extraEnv.envFrom[].configMapRef.optional

`bool`

If true, the ConfigMap is allowed to not exist without blocking pod startup.

### spec.server.extraEnv.envFrom[].secretRef

`SecretRef`

Import all keys from a Secret.

### spec.server.extraEnv.envFrom[].secretRef.name

`string` · required

Name of the Secret.

- rule: {"required":true}

### spec.server.extraEnv.envFrom[].secretRef.optional

`bool`

If true, the Secret is allowed to not exist without blocking pod startup.

### spec.runner

`KubernetesStigmerRunner`

The shared runner: the worker that runs agents beside the server, and
the model key agents call their model with.

### spec.runner.model

`KubernetesStigmerModel`

The key agents call their model with. Agents cannot run without one.

### spec.runner.model.keys

`KubernetesStigmerModelKeys`

The model provider's key, and optionally a gateway address.

- rule: runner.model.keys needs a key: anthropic_api_key, openai_api_key, anthropic_foundry_api_key or aws_bearer_token_bedrock (a gateway address alone signs no request)

### spec.runner.model.keys.anthropicApiKey

`string` · sensitive

An Anthropic API key (ANTHROPIC_API_KEY).

### spec.runner.model.keys.anthropicBaseUrl

`string`

A model gateway's address (ANTHROPIC_BASE_URL), for a proxy or gateway
that speaks Anthropic's API, such as http://models.ai.svc.cluster.local:8080.
It needs a key beside it.

- rule: anthropic_base_url must be an http or https URL

### spec.runner.model.keys.openaiApiKey

`string` · sensitive

An OpenAI API key (OPENAI_API_KEY).

### spec.runner.model.keys.anthropicFoundryApiKey

`string` · sensitive

An Anthropic-on-Microsoft-Foundry key (ANTHROPIC_FOUNDRY_API_KEY).

### spec.runner.model.keys.awsBearerTokenBedrock

`string` · sensitive

An Amazon Bedrock bearer token (AWS_BEARER_TOKEN_BEDROCK).

### spec.runner.model.existingSecretName

`string | valueFrom`

A Secret in the Stigmer's namespace holding the keys by Stigmer's own
names: ANTHROPIC_API_KEY, ANTHROPIC_BASE_URL, OPENAI_API_KEY,
ANTHROPIC_FOUNDRY_API_KEY or AWS_BEARER_TOKEN_BEDROCK.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.runner.resources

`KubernetesCoreV1ApiResourceRequirements`

The runner container's requests and limits. Empty = the operator's
defaults.

### spec.runner.resources.limits

`map<string, string>`

`limits` describes the maximum amount of compute resources allowed. More
info:
https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/

- rule: {"map":{"values":{"string":{"pattern":"^(\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))))?$"}}}}

### spec.runner.resources.requests

`map<string, string>`

`requests` describes the minimum amount of compute resources required. If
Requests is omitted for a container, it defaults to Limits if that is
explicitly specified, otherwise to an implementation-defined value.
Requests cannot exceed Limits. More info:
https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/

- rule: {"map":{"values":{"string":{"pattern":"^(\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))))?$"}}}}

### spec.runner.resources.claims

`[]KubernetesCoreV1ApiResourceClaim`

`claims` lists the names of resources, defined in spec.resourceClaims, that
are used by this container.

This field depends on the DynamicResourceAllocation feature gate.

This field is immutable. It can only be set for containers.

### spec.runner.resources.claims[].name

`string` · required

`name` must match the name of one entry in pod.spec.resourceClaims of the
Pod where this field is used. It makes that resource available inside a
container.

- rule: {"string":{"minLen":"1"}}

### spec.runner.resources.claims[].request

`string`

`request` is the name chosen for a request in the referenced claim. If
empty, everything from the claim is made available, otherwise only the
result of this request.

### spec.runner.storage

`KubernetesStigmerVolume`

The runner's own volume: paused sessions, their approvals and agent
workspaces. Empty = 10Gi.

### spec.runner.storage.size

`string`

The claim's requested capacity, as a Kubernetes quantity (10Gi). Empty =
the operator's default for this volume.

- rule: size must be a Kubernetes quantity such as 10Gi

### spec.runner.storage.storageClassName

`string`

The storage class the claim binds with. Empty = the cluster's default
class (some clusters, GKE 1.37 among them, have none: name one).

### spec.runner.extraEnv

`ContainerEnv`

Extra settings for the runner container, as for the server. A name the
operator sets itself is refused; `env_from` is not accepted.

- rule: runner.extra_env.env_from is not accepted: Stigmer takes single settings, so list each one under variables or secrets
- rule: a setting may be declared in runner.extra_env.variables or runner.extra_env.secrets, not both
- rule: runner.extra_env names a setting Stigmer's operator sets from the declaration, and a second value would silently replace it: set it through its own field, or remove the entry. The operator sets ANTHROPIC_*, OPENAI_API_KEY, AWS_BEARER_TOKEN_BEDROCK, ARTIFACT_STORAGE_TYPE, LOCAL_ARTIFACT_PATH, MODE, TEMPORAL_NAMESPACE, TEMPORAL_SERVICE_ADDRESS, WORKSPACE_ROOT_DIR and the runner's STIGMER_* settings

### spec.runner.extraEnv.variables

`[]EnvVar` · no secrets

Individual environment variables (non-sensitive). Their values are
written into the pod spec, where anyone who can read the workload sees
them; a secret belongs in `secrets`, which the module keeps in a
Kubernetes Secret the container reads by reference.

- secrets: this value is stored where anyone who can view the resource reads it, so a secret reference (`$secret/...`) here is refused -- put a secret in `secrets`, which keeps it in a secret store the workload reads by reference
### spec.runner.extraEnv.variables[].name

`string` · required

The environment variable name.
Must be a valid C_IDENTIFIER: starts with a letter or underscore,
followed by letters, digits, or underscores.

- rule: Must be a valid C_IDENTIFIER (start with letter/underscore, contain only letters, digits, underscores)
- rule: {"required":true}

### spec.runner.extraEnv.variables[].value

`string`

Direct literal value.

### spec.runner.extraEnv.variables[].valueFrom

`ValueFromRef`

Reference to another Planton resource's field.
The orchestrator resolves this and populates the value before invoking IaC modules.

### spec.runner.extraEnv.variables[].valueFrom.kind

`enum`

Allowed values (use exactly as shown):

- `unspecified` -- 0: Default/unspecified
- `TestCatalogKindGeneric` -- 1–49: Test/dev/custom
- `TestCatalogKindKubernetes`
- `AwsAlb` -- 1000–1999: AWS resources AwsSubnet is a prerequisite because an ALB requires at least two subnets in different availability zones -- the spec's subnet references must resolve before the load balancer can be created.
- `AwsCertManagerCert`
- `AwsCloudFront`
- `AwsDynamodb`
- `AwsEcrRepo`
- `AwsEcsCluster`
- `AwsEcsService` -- AwsEcsCluster, AwsEcsTaskDefinition, and AwsSubnet are prerequisites because a service schedules a referenced task-definition revision into a referenced live cluster and places task network interfaces into referenced subnets -- all three references must resolve first.
- `AwsEksCluster` -- AwsSubnet and AwsIamRole are prerequisites because the control plane attaches its network interfaces into referenced subnets and assumes a referenced cluster role that must already carry AmazonEKSClusterPolicy.
- `AwsIamRole`
- `AwsLambda`
- `AwsRdsCluster`
- `AwsRdsInstance`
- `AwsRoute53Zone`
- `AwsS3Bucket`
- `AwsLbTargetGroup` -- AwsVpc is a prerequisite because a target group's health checks and target registrations live inside one VPC -- the spec's vpc_id reference must resolve before the group can be created.
- `AwsSecurityGroup` -- AwsVpc is a prerequisite because every security group is created in a VPC; the E2E install profile resolves vpc_id against the VPC prerequisite.
- `AwsVpc`
- `AwsEksNodeGroup` -- AwsEksCluster is a prerequisite because nodes register with a live control plane; AwsIamRole and AwsSubnet back the node role and worker subnet references.
- `AwsIamUser`
- `AwsKmsKey`
- `AwsEc2Instance`
- `AwsClientVpn` -- Every Client VPN endpoint requires an ACM server certificate at create time; the imported self-signed fixture satisfies it. Subnets/VPC are optional composition (a zero-association endpoint is valid) -- composed scenarios declare them via the e2e-prerequisites annotation.
- `AwsDocumentDb`
- `AwsRoute53DnsRecord` -- AwsRoute53Zone is a prerequisite because every record lives inside a hosted zone -- the spec's zone_id reference must resolve before the record can be created.
- `AwsS3ObjectSet` -- AwsS3Bucket is a prerequisite because the object set's bucket reference is required -- objects cannot exist without the bucket that holds them.
- `AwsSqsQueue`
- `AwsSnsTopic`
- `AwsEventBridgeBus`
- `AwsEventBridgeRule`
- `AwsIamOidcProvider`
- `AwsIamPolicy`
- `AwsIamInstanceProfile` -- AwsIamRole is a prerequisite because an instance profile is a wrapper that must contain a role to be useful -- the profile's spec requires a role reference, so the role must be deployed first.
- `AwsLbListener` -- AwsAlb and AwsLbTargetGroup are prerequisites because a listener is an attachment point on a load balancer and its default action almost always forwards to a target group -- both references must resolve before the listener can be created.
- `AwsLbListenerRule` -- AwsLbListener is a prerequisite because a rule only exists as an attachment on a listener -- the listener_arn reference must resolve before the rule can be created.
- `AwsLaunchTemplate`
- `AwsAutoScalingGroup` -- AwsSubnet and AwsLaunchTemplate are prerequisites because a group cannot exist without subnets to place capacity in and a launch template to launch from -- the spec's subnets and launch_template references must resolve before the group can be created.
- `AwsEksAddon` -- AwsEksCluster is a prerequisite because an add-on installs onto a live control plane -- the spec's cluster_name reference must resolve before the add-on can be created.
- `AwsEksFargateProfile` -- AwsEksCluster, AwsIamRole, and AwsSubnet are prerequisites because a Fargate profile attaches to a live control plane, runs pods as a referenced pod-execution role, and launches them into referenced private subnets -- all three references must resolve first.
- `AwsEksAccessEntry` -- AwsEksCluster and AwsIamRole are prerequisites because an access entry grants a referenced IAM principal access to a live control plane -- both references must resolve before the entry can be created.
- `AwsEcsTaskDefinition` -- AwsIamRole is a prerequisite because the kind's default posture -- Fargate with the awslogs logging default -- is rejected by AWS at registration time without an execution role the agent can assume.
- `AwsHttpApiGateway`
- `AwsStepFunction` -- AwsIamRole is a prerequisite because a state machine cannot be created without an execution role it can assume -- the spec's role_arn reference must resolve before the CreateStateMachine call.
- `AwsHttpApiVpcLink` -- AwsSubnet is a prerequisite because a VPC link is a set of managed ENIs provisioned into referenced subnets -- the subnet references must resolve before the link can be created. Security groups are optional on the link, so they compose per-scenario rather than as a registry prerequisite.
- `AwsHttpApiDomain` -- AwsCertManagerCert is a prerequisite because a custom domain cannot be created without a TLS certificate in the same region covering the domain -- the spec's certificate_arn reference must resolve first.
- `AwsVpcEndpoint` -- AwsVpcEndpoint's composed E2E scenarios reference the AwsVpc prerequisite's outputs (vpc_id + default_route_table_id for gateway endpoints) and the AwsSubnet pair's subnet_id outputs (interface endpoints), so both are genuine deploy-order prerequisites.
- `AwsElasticacheUser`
- `AwsElasticacheUserGroup` -- AwsElasticacheUser is a genuine prerequisite: AWS refuses to create a user group that does not contain a user named "default", so a group's composed E2E scenario must resolve a deployed user's outputs.
- `AwsRedshiftServerlessNamespace`
- `AwsRedshiftServerlessWorkgroup` -- The namespace is a genuine prerequisite: a workgroup attaches to exactly one namespace by name at create time, so its composed E2E scenario must resolve a deployed namespace's outputs. AwsSubnet is a prerequisite because Redshift Serverless requires the workgroup's subnets to span three availability zones.
- `AwsRedisElasticache` -- AwsSubnet is a prerequisite because the module builds an ElastiCache subnet group from referenced subnets -- the spec's subnet references must resolve before the replication group can deploy.
- `AwsOpenSearchDomain`
- `AwsMemcachedElasticache`
- `AwsServerlessElasticache`
- `AwsNlb` -- AwsSubnet is a prerequisite because an NLB requires at least one subnet mapping -- the spec's subnet references must resolve before the load balancer can be created.
- `AwsElasticIp`
- `AwsTransitGateway`
- `AwsGlobalAccelerator`
- `AwsSubnet`
- `AwsInternetGateway`
- `AwsNatGateway` -- AwsInternetGateway is a prerequisite because a public NAT gateway can only become available once the VPC it sits in has an internet gateway attached (AWS rejects the create otherwise) -- so the gateway must be deployed first. AwsVpc is a prerequisite because a REGIONAL NAT gateway (availability_mode = regional) references the VPC directly instead of a subnet.
- `AwsEgressOnlyInternetGateway`
- `AwsElasticFileSystem` -- AwsSubnet and AwsSecurityGroup are prerequisites because mount targets (required, min 1) place the file system's NFS endpoints into subnets and attach security groups -- both references must resolve before the CreateMountTarget calls.
- `AwsEfsAccessPoint` -- AwsElasticFileSystem is a prerequisite because an access point is created INTO a file system -- the spec's required file_system_id reference must resolve before the CreateAccessPoint call.
- `AwsFsxLustreFileSystem`
- `AwsFsxOpenzfsFileSystem`
- `AwsFsxWindowsFileSystem` -- Every Windows file system must join an Active Directory domain; the directory itself is external infrastructure (AWS Managed Microsoft AD or a self-managed domain), so only the network dependency is a declarable prerequisite.
- `AwsFsxOntapFileSystem`
- `AwsFsxOntapStorageVirtualMachine`
- `AwsFsxOntapVolume`
- `AwsFsxDataRepositoryAssociation`
- `AwsCognitoUserPool`
- `AwsCognitoIdentityProvider` -- AwsCognitoUserPool is a prerequisite because an identity provider is created INTO a pool -- the spec's required user_pool_id reference must resolve before the CreateIdentityProvider call.
- `AwsCognitoUserPoolClient` -- AwsCognitoUserPool is a prerequisite because an app client is created INTO a pool -- the spec's required user_pool_id reference must resolve before the CreateUserPoolClient call.
- `AwsCognitoResourceServer` -- AwsCognitoUserPool is a prerequisite because a resource server is created INTO a pool -- the spec's required user_pool_id reference must resolve before the CreateResourceServer call.
- `AwsWafWebAcl`
- `AwsWafIpSet`
- `AwsWafRegexPatternSet`
- `AwsCloudwatchLogGroup`
- `AwsCloudwatchAlarm`
- `AwsCloudwatchCompositeAlarm`
- `AwsKinesisStream`
- `AwsKinesisFirehose` -- Every Firehose destination requires an S3 configuration (the primary target for extended_s3; the failed/all-document backup for the rest) and an IAM role Firehose assumes to write to it, so both are hard deploy prerequisites.
- `AwsKinesisStreamConsumer` -- A consumer registers against exactly one stream and cannot exist without it.
- `AwsAthenaWorkgroup`
- `AwsGlueCatalogDatabase`
- `AwsRedshiftCluster`
- `AwsSagemakerDomain` -- AI/ML A domain cannot exist without VPC subnets and a SageMaker execution role (default_user_settings.execution_role_arn is required), so both are hard deploy prerequisites.
- `AwsAppRunnerService` -- A service can run entirely on companion defaults, so the App Runner family's kinds are dependency-free leaves except the VPC connector (which cannot exist without subnets and security groups). A service's companion references (auto scaling / VPC connector / observability / WAF) are optional composition -- scenarios declare them via the e2e-prerequisites annotation.
- `AwsAppRunnerAutoScalingConfiguration`
- `AwsAppRunnerVpcConnector`
- `AwsAppRunnerObservabilityConfiguration`
- `AwsTransitGatewayVpcAttachment` -- AwsTransitGateway is a prerequisite because an attachment cannot exist without the gateway it attaches to; AwsSubnet because the attachment provisions an ENI into at least one subnet (the VPC arrives transitively through the subnet's own prerequisites).
- `AwsTransitGatewayRouteTable` -- Only the gateway is a hard prerequisite: a route table can exist empty. Associations, propagations, and routes referencing attachments are optional composition -- scenarios declare them via the e2e-prerequisites annotation.
- `AwsBatchComputeEnvironment` -- A MANAGED compute environment always launches into VPC subnets, so the subnet is a hard deploy prerequisite (security groups are required only for the Fargate types -- scenario-declared, not a registry edge).
- `AwsBatchJobQueue` -- A job queue cannot exist without at least one VALID compute environment to map onto.
- `AwsBatchSchedulingPolicy`
- `AwsBatchJobDefinition`
- `AwsCodeBuildProject` -- CI/CD
- `AwsCodePipeline`
- `AwsMwaaEnvironment` -- Workflow / Orchestration AwsSubnet and AwsSecurityGroup are prerequisites because the environment's network interfaces are placed in referenced private subnets and AWS requires at least one attached security group at creation.
- `AwsNeptuneCluster` -- Graph Database
- `AwsMemorydbCluster` -- A cluster always launches into a subnet group; the subnets are the hard deploy prerequisite. The ACL it attaches is optional composition (the built-in "open-access" ACL needs no resource) -- scenarios declare the ACL/user chain via the e2e-prerequisites annotation.
- `AwsMemorydbUser`
- `AwsMemorydbAcl` -- An empty ACL is valid (MemoryDB has no mandatory "default" member), so the user is optional composition -- the composed scenario declares it via the e2e-prerequisites annotation, never a registry edge.
- `AwsMskCluster` -- Streaming AwsSubnet and AwsSecurityGroup are prerequisites because brokers are placed in referenced subnets and AWS requires at least one attached security group at creation.
- `AwsMskServerlessCluster` -- AwsSubnet is a prerequisite because the serverless cluster's network interfaces are placed in referenced subnets (security groups are optional -- AWS attaches the VPC default group when none are referenced).
- `AwsLambdaEventSourceMapping` -- AwsLambda is a prerequisite because a mapping cannot exist without the function it invokes (a required reference). Event sources (SQS, Kinesis, DynamoDB, MSK) are optional composition -- scenarios declare them via the e2e-prerequisites annotation rather than taxing every consumer's chain.
- `AwsSnsSubscription` -- AwsSnsTopic is a prerequisite because a subscription cannot exist without the topic it subscribes to (a required reference). Endpoints (SQS queues, Lambda functions, Firehose streams) are optional composition -- scenarios declare them via the e2e-prerequisites annotation rather than taxing every consumer's chain.
- `AwsPlantonRunner` -- AwsSubnet is a prerequisite because the runner appliance places its network interfaces into referenced subnets -- the placement reference must resolve before the appliance can deploy.
- `AwsRoute53HealthCheck`
- `AwsSesConfigurationSet` -- Both SES kinds are dependency-free leaves: an identity's configuration set is optional composition (scenarios declare it via the e2e-prerequisites annotation), and a configuration set's event destinations reference other kinds only optionally.
- `AwsSesEmailIdentity`
- `AwsSecretsManagerSecret` -- A dependency-free leaf: the KMS key, rotation Lambda, and external rotation role references are all optional composition -- scenarios declare them via the e2e-prerequisites annotation, never registry edges.
- `AwsOpenSearchServerlessCollection` -- A dependency-free leaf: the collection-scoped encryption/network/ data-access/retention policies are module-rendered, and the KMS key and data-access principal references are optional composition (e2e-prerequisites annotation).
- `AwsBedrockGuardrail` -- A dependency-free leaf: the KMS key reference is optional composition (e2e-prerequisites annotation); published versions are folded satellites of the guardrail itself.
- `AwsBedrockCustomModel` -- AwsIamRole is a prerequisite because Bedrock assumes the job role to read training data and write outputs; the S3 locations and KMS key are optional composition (e2e-prerequisites annotation).
- `AwsBedrockInferenceProfile` -- A dependency-free leaf: the model source is a foundation model or an AWS system-defined cross-region profile, never a customer resource.
- `AwsBedrockProvisionedThroughput` -- A dependency-free leaf in the registry: capacity is typically bought for an AwsBedrockCustomModel (the default reference), but foundation model ARNs are equally legal, so the edge is optional composition.
- `AwsBedrockModelAccess` -- A dependency-free leaf: the agreement covers an AWS-listed foundation model, never a customer resource.
- `AwsBedrockInvocationLogging` -- Region settings singleton (one invocation-logging configuration per account+region; identity = the region). Delivery destinations are optional references (at least one of CloudWatch/S3, enforced by CEL), so prerequisites stay empty and E2E fixtures ride scenario annotations.
- `AwsBedrockAgent` -- AwsIamRole is a prerequisite because the Bedrock service assumes the agent resource role to invoke models, action-group Lambdas, and knowledge bases; the guardrail, KMS key, provisioned throughput, and collaborator/knowledge-base edges are optional composition (e2e-prerequisites annotation). Action groups, aliases, collaborators, and knowledge-base associations are folded satellites of the agent.
- `AwsBedrockKnowledgeBase` -- AwsIamRole is a prerequisite because the Bedrock service assumes the knowledge-base role to read data sources, call the embedding model, and read/write the vector store; the vector-store and data-source reference edges (OpenSearch, S3, Secrets Manager, ...) are optional composition (e2e-prerequisites annotation). Data sources are folded satellites of the knowledge base.
- `AwsBedrockFlow` -- AwsIamRole is a prerequisite because the Bedrock service assumes the flow execution role to invoke the models, agents, knowledge bases, and Lambdas its nodes reference; the node-level reference edges are optional composition (e2e-prerequisites annotation).
- `AwsBedrockPrompt` -- A dependency-free leaf: variants target AWS-listed foundation models by ID; targeting another agent's alias is optional composition (e2e-prerequisites annotation).
- `AwsBedrockAgentCoreRuntime` -- AwsIamRole is a prerequisite because the AgentCore service assumes the runtime role to pull the container image or read the S3 code bundle and to run the hosted agent; the code-bundle S3 bucket and VPC placement edges are optional composition (e2e-prerequisites annotation). Endpoints and the runtime's resource policy are folded satellites of the runtime.
- `AwsBedrockAgentCoreGateway` -- AwsIamRole is a prerequisite because the gateway assumes its role to reach targets (invoke Lambdas, sign SigV4 requests); the target and credential-provider reference edges (runtime, Lambda, Identity providers, policy engine) are optional composition (e2e-prerequisites annotation). Targets are folded satellites of the gateway - AWS deletes them before the gateway at destroy.
- `AwsBedrockAgentCoreMemory` -- A dependency-free leaf for built-in strategies: the execution role (custom strategies, Kinesis delivery), KMS key, and Kinesis stream edges are optional composition (e2e-prerequisites annotation). Strategies are folded satellites of the memory - AWS serializes their changes through the parent.
- `AwsBedrockAgentCoreIdentity` -- A dependency-free leaf: workload identities, credential providers, and the Cedar policy engine with its policies are all name-keyed arms of one identity-and-access bundle; the KMS key edge is optional composition (e2e-prerequisites annotation). The account/region token-vault CMK is deliberately NOT modeled here (settings singleton).
- `AwsBedrockAgentCoreTools` -- A dependency-free leaf in the SANDBOX/PUBLIC postures: the execution role (recordings, certificates), S3, Secrets Manager, and VPC edges are optional composition (e2e-prerequisites annotation). Browsers, profiles, and code interpreters are name-keyed arms of one tools bundle; AWS exposes no update - every field change recreates the tool.
- `AwsBedrockAgentCoreEvaluation` -- The AgentCore Evaluations bundle - evaluators (LLM-judge or Lambda scorers), harnesses (repeatable agent test benches), and online evaluation configs (continuous scoring of sampled production sessions). Deploys standalone - no arm requires an agent runtime to exist. No registry prerequisite: every arm is optional, so no dependency is required for the kind to function (scenarios compose IAM roles via annotations).
- `AwsBedrockAgentCoreTokenVault` -- Account/region settings singleton: sets the KMS key on the ONE default AgentCore token vault. The KMS reference is conditional on key_type (CEL-enforced), so prerequisites stay empty and E2E fixtures ride scenario annotations.
- `AwsSagemakerModel` -- The immutable serving definition (container image + artifacts + execution role) that endpoints deploy - one container or an inference pipeline.
- `AwsSagemakerEndpoint` -- A real-time inference endpoint WITH its folded endpoint configuration - the configuration is immutable upstream, so the modules roll name-suffixed configurations create-before-destroy and repoint the endpoint.
- `AwsSagemakerNotebookInstance` -- A managed Jupyter notebook EC2 instance with its folded lifecycle configuration (bootstrap scripts).
- `AwsSagemakerFeatureGroup` -- A Feature Store feature group - online and/or offline stores over a declared feature schema.
- `AwsSagemakerModelRegistry` -- A model registry package group with its folded resource policy - model package VERSIONS register into it imperatively (training pipelines), never declaratively.
- `AwsSagemakerPipeline` -- An ML workflow DAG (the SageMaker pipeline-definition JSON) that executions run against - free to create, billed per execution.
- `AwsSagemakerImage` -- A named registry entry exposing YOUR container images to Studio, with folded AWS-numbered versions (append-only by position).
- `AwsSagemakerMlflowServer` -- The classic hourly-billed managed MLflow tracking server (~25 min to provision; billed hourly from creation whether or not anyone is tracking). The serverless successor is AwsSagemakerMlflowApp.
- `AwsSagemakerMlflowApp` -- The serverless MLflow 3.x deployment (billed per use) - standalone, associating with SageMaker domains; NOT a tracking-server satellite.
- `AwsRestApiGateway` -- A full REST API (API Gateway v1): the resource/method tree with inline integrations (or an imported OpenAPI document), one stage with an explicit hash-triggered deployment, and the API-scoped satellites (authorizers, models, validators, gateway responses, policy, documentation, client certificate). Self-contained: a MOCK-integration API needs no other resource.
- `AwsRestApiDomain` -- A custom domain for REST APIs with base-path mappings and - for PRIVATE domains - VPC-endpoint access associations. AwsCertManagerCert is a prerequisite because the domain cannot be created without a TLS certificate covering it.
- `AwsRestApiUsagePlan` -- A usage plan metering REST API consumers - stage coverage, quota, throttles, and the API keys it admits. No registry prerequisite: a plan is valid with no stage coverage (scenarios compose the REST API via annotations).
- `AwsRestApiVpcLink` -- A REST API VPC link fronting an internal Network Load Balancer so REST integrations reach private services. AwsNlb is a prerequisite because AWS rejects link creation without the target balancer.
- `AwsApiGatewayAccountSettings` -- Region settings singleton (one API Gateway account object per account+region; identity = the region). The CloudWatch role is an optional reference (unset = the explicit no-logging posture), so prerequisites stay empty and E2E fixtures ride scenario annotations.
- `AwsCloudTrail` -- The account's API audit trail. AwsS3Bucket is a prerequisite because AWS rejects trail creation without a delivery bucket carrying the CloudTrail service-principal policy. 1240 opens the governance sub-band (1240-1249).
- `AwsConfigRecorder` -- Region singleton (one AWS Config recorder per region, named "default" by AWS; identity = the region). AwsIamRole is a prerequisite because the recorder cannot exist without its service role.
- `AwsConfigRule` -- One AWS Config compliance rule (managed, custom-lambda, or custom-policy; account- or organization-scoped) with optional auto-remediation. Managed rules need no prerequisites; the custom-lambda arm's function reference is conditional, so E2E fixtures ride scenario annotations.
- `AwsGuardDuty` -- Region singleton (AWS allows one GuardDuty detector per account+region; the detector has no name - identity = the region). Satellite references (S3 export bucket, KMS key) are conditional, so E2E fixtures ride scenario annotations.
- `AwsCloudTrailEventDataStore` -- CloudTrail Lake: a queryable, immutable event data store with its own retention and billing lifecycle - no trail required. The KMS key reference is conditional, so E2E fixtures ride scenario annotations.
- `AwsConfigAggregator` -- AWS Config cross-account/cross-region aggregation: the aggregator (collector side) and/or the reciprocal authorization grants (source-account side). Works with zero recorders; the org-source role reference is conditional, so E2E fixtures ride scenario annotations.
- `AwsConfigConformancePack` -- An AWS Config conformance pack (account- or organization-scoped): a template bundle that creates its own Config rules. Deployment requires an active Config recorder in the region (a service-side requirement, not a spec reference), so E2E fixtures ride scenario annotations.
- `AwsGuardDutyMalwareProtectionPlan` -- GuardDuty Malware Protection for S3: scans new objects in one bucket - a standalone plan protecting a bucket, not a detector satellite (its schema carries no detector reference). The execution role and the protected bucket are required references.
- `AwsBackupVault` -- An AWS Backup vault - the encrypted container recovery points live in, as either a standard vault (with its lock, access policy, and notification satellites) or a logically air-gapped vault (AWS's own VaultType discriminator). The KMS and SNS references are conditional, so E2E fixtures ride scenario annotations. 1250 opens the backup sub-band (1250-1259).
- `AwsBackupPlan` -- An AWS Backup plan: scheduled backup rules plus the resource selections that assign resources to them. AwsBackupVault is a prerequisite because every rule requires a target vault; the selections' IAM role is conditional and rides scenario annotations.
- `AwsBackupFramework` -- A Backup Audit Manager framework: compliance controls evaluating backup posture. No schema-required references (the Config recorder its evaluations need is a lane fixture, not a spec reference).
- `AwsBackupReportPlan` -- A Backup Audit Manager report plan: scheduled compliance/job reports delivered to S3. AwsS3Bucket is a prerequisite because the delivery channel's bucket is required.
- `AwsBackupRestoreTestingPlan` -- An AWS Backup restore testing plan with its folded selections: scheduled restore tests proving recovery points actually restore. Vault targeting accepts the "*" wildcard, so fixtures are conditional and ride scenario annotations.
- `AwsBackupSettings` -- Account/region settings singleton for AWS Backup: the account's global settings (cross-account backup) and the region's resource-type opt-in/management preferences. Both provider deletes are no-ops - settings persist after destroy.
- `AwsSsmParameter` -- An SSM Parameter Store entry (String/StringList/SecureString). The parameter's name is an explicit spec field - names are hierarchical paths ("/prod/db/url") metadata.name cannot carry. The KMS reference is conditional (SecureString only), so E2E fixtures ride scenario annotations. 1260 opens the SSM sub-band (1260-1269).
- `AwsSsmDocument` -- A customer-owned SSM document (Command/Automation/Session/...): reusable action definitions managed nodes and automations execute. State Manager associations are their own AwsSsmAssociation kind - an association binds ANY document (AWS-managed included), so it is not this document's satellite.
- `AwsSsmMaintenanceWindow` -- An SSM maintenance window with its folded target registrations and tasks (Run Command / Automation / Lambda / Step Functions) - the targets and tasks are true window satellites (ForceNew window_id edges). Identity is the AWS-generated "mw-..." id.
- `AwsSsmPatchBaseline` -- An SSM patch baseline with its folded patch-group registrations and the account/region default-baseline designation (delete RESTORES AWS's own predefined default for the OS). Identity is the AWS-generated "pb-..." id.
- `AwsSsmAssociation` -- A State Manager association: the binding of an SSM document to targets on a schedule. Split from the document kind because the document reference is a free string with no structural edge - associations routinely bind AWS-managed documents (AWS-RunShellScript, ...) with no user document anywhere, so no registry prerequisite either. Identity is the AWS-generated association UUID.
- `AwsOrganization` -- THE AWS Organization of the deploying account - creating it makes the caller the management account. Trusted service access, delegated administrators, the org's singleton resource policy, and centralized root-access management (IAM's organizations features - a management-account act requiring iam.amazonaws.com trusted access) fold in (none has a life of its own; the standalone service-access resource fights the org's own argument with a perpetual diff). Deleting this deletes the entire organization. 1270 opens the Organizations sub-band (1270-1279). A container kind: the organization is the outer wall of the tenancy tree AWS's own console draws -- organizational units are rooms inside it and member accounts stand inside those. A first-level unit names the organization through its parent reference and is placed inside it; a policy attached to the root names the organization too, but a policy is a guard applied from above, never a resident, and its attachment reference is containment-exempt on its spec.
- `AwsOrganizationalUnit` -- An organizational unit in the org's OU tree. The display name is an explicit spec field (OU names allow spaces metadata.name cannot carry); the parent reference (root or parent OU) is required and immutable, so the organization is a registry prerequisite. A container kind: a unit is a room of the tenancy tree -- nested units and member accounts name it through their parent reference and are placed inside it. A policy attached to a unit names it too and stands beside the room (containment-exempt on the policy's spec): one policy attaches to many units and cannot live in all of them.
- `AwsOrganizationAccount` -- A MEMBER account of the organization: creation, OU placement, and the account-level settings satellites (alternate/primary contacts, opt-in region enablement) fold onto the created account's ID. Destroy is never a clean delete (remove-from-org or ~90-day close) - taught on the spec. No registry prerequisite by the schema-required-only rule (the OU parent reference is optional).
- `AwsOrganizationPolicy` -- An Organizations policy (SCP and its twelve sibling types) with its folded attachments to roots, OUs, and member accounts. The policy type must be enabled on the organization first; AWS-managed policies are never adopted. No registry prerequisite by the schema-required-only rule (attachments are optional).
- `AwsBudget` -- A Budgets budget (COST/USAGE/RI/Savings Plans coverage and utilization) with its folded budget actions as name-keyed satellites - an action exists only on its budget and fires an IAM-policy application, an SCP attachment, or SSM instance stops when a threshold breaches. Budgets is account-global (served from us-east-1; the spec region is the provider endpoint). 1280 opens the cost-management sub-band (1280-1289).
- `AwsCostAnomalyMonitor` -- A Cost Explorer anomaly monitor (DIMENSIONAL over one dimension, or CUSTOM over a CE expression) with its folded alert subscriptions - a subscription's monitor list is the structural edge that makes it this monitor's satellite. Account-global; AWS identifies both by ARN.
- `AwsCostCategory` -- A Cost Explorer cost category: ordered rules (regular expression rules or inherited-value rules) over the recursive CE expression tree, plus split-charge rules. The account's cost-allocation-tag activation toggle is deliberately NOT folded here - it is a per-tag-key account feature with no edge to any category, so many category instances would fight over one account object.
- `AwsIamGroup` -- An IAM group with its folded declarative membership (the authoritative users list) and group policies - name-keyed inline documents plus managed-policy attachments. IAM is global; identity is the group name (renames update in place, the ARN recomputes). 1290 opens the IAM P1 sub-band (1290-1299).
- `AwsIamSamlProvider` -- An IAM SAML identity provider: the account's federation trust anchor, created from the IdP's metadata XML (a public document carrying certificates, not a secret). Identity is the provider ARN; the name is write-once.
- `AwsIamAccountSettings` -- Account settings singleton for IAM (a GLOBAL service - one object per ACCOUNT, not per region): the sign-in alias, the password policy, and the STS global-endpoint token version. Destroy contracts DIFFER per arm (each taught on its arm): the alias truly deletes, the password policy resets to AWS defaults, the STS preference is a no-op delete that persists.
- `AwsCloudwatchDashboard` -- A CloudWatch dashboard: one named dashboard whose widget layout is the dashboard-body JSON document (modeled as a typed Struct, the catalog's uniform policy-document idiom). Dashboards are untaggable at AWS. Identity is the dashboard name; every change is an in-place PutDashboard upsert. 1300 opens the CloudWatch observability P1 sub-band (1300-1309).
- `AwsCloudwatchSynthetics` -- CloudWatch Synthetics: a canary (a scheduled scripted probe running from an S3-staged code bundle under an execution role, writing run artifacts to S3) plus the grouping surface - owned groups and the canary's group associations (joins by group NAME, so shared groups are referenced, never fought over). A groups-only instance manages shared groups with no canary.
- `AwsCloudwatchLogDelivery` -- CloudWatch Logs delivery: the two ways logs leave CloudWatch. The vended-log arm pivots on a delivery SOURCE (one AWS resource whose service vends logs) with name-keyed deliveries fanning out to delivery destinations (S3 / CloudWatch Logs / Firehose / X-Ray), each created inline or referenced by ARN. The cross-account arm is the legacy Kinesis subscription destination with its access policy (whose delete is a no-op at AWS - the policy persists).
- `AwsCloudwatchLogAccountPolicy` -- A CloudWatch Logs account-level policy: one policy object per (name, type) pair per region - data protection, subscription filter, field index, transformer, or metric extraction - applied account-wide, optionally narrowed by selection criteria. Standalone account configuration, never a per-log-group satellite.
- `AwsCloudwatchLogAnomalyDetector` -- A CloudWatch Logs anomaly detector: one detector trains over a LIST of log groups (multi-parent scope - never a single group's satellite), surfacing anomalies on a chosen evaluation frequency with a bounded visibility window.
- `AwsCloudwatchLogResourcePolicy` -- A CloudWatch Logs resource policy: the account-scoped named policy (or resource-scoped policy on one log group ARN) that grants AWS services permission to write logs - Route53 query logging, EventBridge, and friends. Exactly one scope per instance.
- `AwsManagedPrometheus` -- An Amazon Managed Prometheus workspace with its folded satellites: workspace configuration (retention, label-set limits - a created-via-update singleton whose delete is a no-op at AWS), the alert manager definition (strictly one per workspace), name-keyed rule group namespaces, query logging, the workspace resource policy, and alias-keyed anomaly detectors. Scrapers are deliberately NOT folded here - a scraper can target CloudWatch with zero AMP workspaces, so it is its own kind.
- `AwsManagedPrometheusScraper` -- An Amazon Managed Prometheus scraper: the agentless collector. Source is an EKS cluster or a bare VPC placement (both replace-on-change); destination is an AMP workspace or a CloudWatch dataset. Carries its own scraper logging configuration satellite. Scrape configuration is optional on the EKS arm (AWS publishes a default, resolved at deploy) and required on the VPC arm.
- `AwsEventBridgePipe` -- An EventBridge Pipe: one point-to-point integration reading from one source (SQS, Kinesis, DynamoDB streams, MSK or self-managed Kafka, ActiveMQ/RabbitMQ), optionally filtering and enriching in-flight, and delivering to one target (ECS, Batch, Lambda, Step Functions, Kinesis, SQS, Redshift, SageMaker, CloudWatch Logs, EventBridge buses, HTTP via API destinations). The source is fixed for life (replace-on-change); the target swaps in place. 1310 opens the EventBridge extras P1 sub-band (1310-1319).
- `AwsEventBridgeScheduler` -- An EventBridge Scheduler schedule: cron/rate/one-time invocation of one target under an execution role, with flexible time windows, retry policy, and a dead-letter queue. The schedule GROUP is folded own-XOR-existing (a name-and-tags container - the provider's own update path is tags-only); unset means AWS's default group.
- `AwsEventBridgeApiDestination` -- An EventBridge API destination with its connection: the authenticated HTTP(S) endpoint rules, pipes, and schedules invoke. Two independently deployable arms - the CONNECTION (the shareable auth trust anchor: api-key, basic, or OAuth credentials that AWS stores in Secrets Manager) and the DESTINATION (endpoint + method + rate limit) whose connection is owned inline or referenced by ARN.
- `AwsVpcPeering` -- A VPC peering connection, as a request-XOR-accept mode union: the REQUEST arm creates the peering from its VPC toward a peer VPC (same-account auto-accept supported; cross-account/cross-region stays pending until accepted), the ACCEPT arm adopts and accepts a pending connection by ID from the accepter side. DNS-resolution options fold into both arms. 1320 opens the VPC networking P1 sub-band (1320-1329).
- `AwsNetworkAcl` -- A network ACL: the stateless subnet-level firewall - ordered ingress/egress rules (allow or deny, evaluated by rule number) and the subnet associations, all folded in-line as the single declarative owner (the standalone rule/association resources are the same payload and fight the in-line form).
- `AwsManagedPrefixList` -- A customer-managed prefix list: a named, versioned set of CIDR blocks that security-group rules, NACL rules, and route tables reference as one object. Entries fold in-line; max_entries is the capacity contract (referencing consumes that many rule slots regardless of how many entries exist).
- `AwsEbsVolume` -- A standalone EBS volume as a create-XOR-copy union (fresh in a zone, or cloned from another volume) with attachments managed in-line. 1330 opens the block & object storage sub-band (1330-1339).
- `AwsEbsSnapshot` -- An EBS snapshot as a three-way source union (snapshot a volume, copy a snapshot, or import a disk image) with archive tiering, fast snapshot restore, and cross-account share grants in-line.
- `AwsS3DirectoryBucket` -- An S3 directory bucket (S3 Express One Zone): single-AZ, single-digit-millisecond object storage. The modules derive the mandated "{name}--{zone_id}--x-s3" bucket name.
- `AwsS3TableBucket` -- An S3 table bucket (S3 Tables - managed Apache Iceberg storage) with its namespaces, tables, policies, and replication folded in-line as the single declarative owner.
- `AwsS3VectorBucket` -- An S3 vector bucket (AI embedding storage with similarity query) with its vector indexes folded in-line - the natural backend for Bedrock knowledge bases.
- `AwsDlmLifecyclePolicy` -- A Data Lifecycle Manager policy: account-level, tag-targeted snapshot/AMI automation (create, retain, archive, copy cross-region, share, deprecate) as a default-XOR-custom mode union. AwsIamRole is a prerequisite because DLM acts through a required execution role.
- `AwsRoute53ResolverEndpoint` -- A Route 53 Resolver endpoint (the hybrid-DNS bridge between a VPC and outside networks) with its forwarding rules and their VPC associations managed in-line. Subnets place the ENIs and security groups guard them - both schema-required. 1340 opens the DNS & service discovery sub-band (1340-1349).
- `AwsRoute53ResolverFirewall` -- A Route 53 Resolver DNS Firewall rule group with its domain lists, filtering rules, and VPC associations managed in-line - the DNS-layer block/allow policy for VPC egress queries. AwsVpc is a prerequisite because the association arm filters a referenced VPC.
- `AwsRoute53ResolverQueryLog` -- A Resolver query logging configuration (every DNS query VPCs make through the resolver, to CloudWatch Logs / S3 / Firehose) with its VPC associations managed in-line. AwsVpc is a prerequisite because the association arm logs a referenced VPC.
- `AwsCloudMapNamespace` -- An AWS Cloud Map namespace (HTTP-XOR-private-DNS-XOR-public-DNS) with its discoverable services and statically registered instances managed in-line - the service-discovery registry ECS and custom applications look each other up in. AwsVpc is a prerequisite because the private-DNS arm binds its hosted zone to a referenced VPC.
- `AwsAppSyncApi` -- An AppSync API - AWS's managed API service, as a GraphQL API (SDL schema, resolvers over data sources, caching, the MERGED federation variant) XOR an Events API (real-time pub/sub over channel namespaces) - with its data sources, resolvers, functions, types, API keys, and custom domain managed in-line. Every backend reference (data source targets, roles, the certificate) is optional, so no registry prerequisite - lanes exercise the fixture-free arms.
- `AwsLambdaLayer` -- A Lambda layer version - a shared code archive (libraries, custom runtimes) functions attach by ARN - with its cross-account and organization share grants managed in-line. The archive lives in S3 (an optional reference, so no registry prerequisite - lanes compose their own bucket fixture). 1351 sits in the app & data services sub-band (1350-1359; 1350 opens it with AwsAppSyncApi).
- `AwsRdsProxy` -- An RDS Proxy - the managed connection pool between connection-hungry applications and a database - with its connection-pool tuning, additional endpoints, and database target managed in-line. AwsIamRole is a prerequisite because the proxy assumes a required role to read database credentials from Secrets Manager; AwsSubnet because the proxy's network interfaces require at least two subnets.
- `AwsAuroraDsql` -- An Aurora DSQL cluster - serverless, PostgreSQL-compatible distributed SQL with active-active multi-region pairing managed in-line. No prerequisites: a single-region cluster deploys from defaults alone (the KMS and peer references are optional arms).
- `AwsEcrRegistrySettings` -- Region settings singleton (one private ECR registry per account+region): the registry policy, scanning configuration, replication rules, pull-through cache rules, repository creation templates, account settings, and pull-time update exclusions. Repository-scoped surface stays on AwsEcrRepo.
- `AwsPrivateCa` -- An AWS Private Certificate Authority with composed activation (a ROOT self-signs at apply; a subordinate activates from a parent AwsPrivateCa), issued certificates, the ACM renewal permission, and the resource policy managed in-line. No prerequisites: the S3 (CRL) and parent-CA references are optional arms.
- `AwsSesAccountSettings` -- Account/region settings singleton (one SES account object per account+region): the suppression list and VDM posture. 1360 opens the SES P1 sub-band (1360-1369).
- `AzureResourceGroup` -- 2000–2999: Azure resources
- `AzureAksCluster` -- AzureResourceGroup is the only required parent: the cluster is created inside a referenced resource group. Subnet is optional on the default node pool (AKS provisions managed networking when unset).
- `AzureAksNodePool` -- AzureAksCluster is a prerequisite because a node pool attaches to an existing cluster by ARM ID; the resource group chains transitively.
- `AzureContainerRegistry` -- AzureResourceGroup is a prerequisite because a container registry is created inside a resource group.
- `AzureDnsZone` -- AzureResourceGroup is a prerequisite because the DNS zone is created inside a referenced resource group that must already exist.
- `AzureKeyVault` -- AzureResourceGroup is a prerequisite because a key vault is created inside a referenced resource group in composed environments.
- `AzureVirtualNetwork` -- AzureResourceGroup is a prerequisite because a virtual network is created inside a referenced resource group in composed environments.
- `AzureNatGateway` -- AzureResourceGroup is a prerequisite because a NAT gateway is created inside a referenced resource group in composed environments.
- `AzureVirtualMachine` -- AzureNetworkInterface is a prerequisite because a virtual machine attaches at least one NIC (the subnet, network, and resource group chain transitively through the NIC's own prerequisites).
- `AzureStorageAccount` -- AzureResourceGroup is a prerequisite because a storage account is created inside a referenced resource group in composed environments.
- `AzureDnsRecord` -- AzureDnsZone is a prerequisite because a record set is created inside a referenced zone (the resource group chains transitively through the zone). Public DNS zone names are not globally unique, so a shared zone fixture is safe to recreate across scenarios.
- `AzureSubnet` -- AzureVirtualNetwork is a prerequisite because a subnet is an ARM child of a referenced network -- the network must exist before the subnet can be written. (The resource group arrives transitively through the network's own prerequisite declaration.)
- `AzureNetworkSecurityGroup` -- AzureResourceGroup is a prerequisite because a network security group is created inside a referenced resource group in composed environments.
- `AzurePublicIp` -- AzureResourceGroup is a prerequisite because a public IP is created inside a referenced resource group in composed environments.
- `AzurePrivateEndpoint` -- AzureSubnet is a prerequisite because a private endpoint draws its private IP from a referenced subnet (the virtual network and resource group chain transitively through the subnet's own prerequisite). The connection target is polymorphic and the DNS zones / ASGs are optional, so none of those are prerequisites.
- `AzurePrivateDnsZone` -- AzureResourceGroup is a prerequisite because a private DNS zone is created inside a referenced resource group in composed environments.
- `AzureApplicationGateway` -- AzureSubnet is a prerequisite because a gateway cannot exist without its dedicated gateway_ip_configuration subnet (the network and resource group chain transitively through the subnet's own prerequisites); public frontends additionally reference a public IP, but private-only gateways are legal, so it is not a registry prerequisite.
- `AzureLoadBalancer` -- AzureResourceGroup is a prerequisite because a load balancer is created inside a referenced resource group (frontends additionally reference subnets or public IPs, but neither is universally required, so they are not registry prerequisites).
- `AzureRouteTable` -- AzureResourceGroup is a prerequisite because a route table is created inside a referenced resource group in composed environments.
- `AzurePrivateDnsZoneVirtualNetworkLink` -- AzurePrivateDnsZone and AzureVirtualNetwork are prerequisites because a virtual network link is a child resource of a referenced zone and binds it to a referenced network -- both must exist before the link can be written. (The resource group arrives transitively through the zone's and network's own prerequisite declarations.)
- `AzureVirtualNetworkPeering` -- AzureVirtualNetwork is a prerequisite because a peering is an ARM child of its local network and binds it to a remote network -- the local network must exist before the peering can be written. (The resource group arrives transitively through the network's own prerequisite declaration.)
- `AzurePublicIpPrefix` -- AzureResourceGroup is a prerequisite because a public IP prefix is created inside a referenced resource group in composed environments.
- `AzureNetworkInterface` -- AzureSubnet is a prerequisite because a network interface's IP configurations deploy into a subnet (the virtual network and resource group chain transitively through the subnet's own prerequisite).
- `AzureManagedDisk` -- AzureResourceGroup is a prerequisite because a managed disk is created inside a resource group.
- `AzureVirtualMachineScaleSet` -- AzureSubnet is a prerequisite because every scale-set instance's network interface deploys into a subnet (the virtual network and resource group chain transitively through the subnet's own prerequisite).
- `AzureKeyVaultKey` -- AzureKeyVault is a prerequisite because a key is a data-plane object inside a referenced vault -- the vault must exist before the key can be written (the resource group chains transitively through the vault's own prerequisite).
- `AzureKeyVaultCertificate` -- AzureKeyVault is a prerequisite because a certificate is a data-plane object inside a referenced vault -- the vault must exist before the certificate can be enrolled or imported (the resource group chains transitively through the vault's own prerequisite).
- `AzureKeyVaultSecret` -- AzureKeyVault is a prerequisite because a secret is a data-plane object inside a referenced vault -- the vault must exist before the secret can be written (the resource group chains transitively through the vault's own prerequisite). Part of the Key Vault family (2005, 2025-2026) despite the out-of-run number -- enum numbers are pinned by the registry snapshot; never renumber.
- `AzureWebApplicationFirewallPolicy` -- AzureResourceGroup is a prerequisite because a WAF policy is created inside a referenced resource group; the Application Gateways that attach the policy reference it, never the reverse.
- `AzureApplicationSecurityGroup` -- AzureResourceGroup is a prerequisite because an application security group is created inside a referenced resource group; network interfaces, scale-set IP configurations, and NSG security rules reference the group, never the reverse.
- `AzureDiskEncryptionSet` -- AzureKeyVaultKey is a prerequisite because a disk encryption set wraps customer data with a referenced key -- the key (and its vault, which chains transitively) must exist before the set can resolve the key URL at create time.
- `AzurePostgresqlFlexibleServer` -- AzureResourceGroup is a prerequisite because a server is created inside a referenced resource group (VNet injection additionally references a delegated subnet and a private DNS zone, but neither is universally required, so they are not registry prerequisites).
- `AzureRedisCache` -- AzureResourceGroup is a prerequisite because the cache is created inside a referenced resource group (VNet injection additionally references a dedicated subnet, but only the Premium tier supports it, so it is not a registry prerequisite).
- `AzureCosmosdbAccount` -- AzureResourceGroup is a prerequisite because the account is created inside a referenced resource group.
- `AzureMssqlServer` -- AzureResourceGroup is a prerequisite because the logical server is created inside a referenced resource group.
- `AzureMysqlFlexibleServer` -- AzureResourceGroup is a prerequisite because a server is created inside a referenced resource group (VNet injection additionally references a delegated subnet and a private DNS zone, but neither is universally required, so they are not registry prerequisites).
- `AzureMssqlDatabase` -- The parent logical server is referenced via server_id, not auto-deployed: E2E scenarios declare their own server fixture (minimal-server.yaml or the pool-attach chain through AzureMssqlElasticPool) so sequential subtests never destroy and recreate the same globally unique server_name.
- `AzureMssqlElasticPool` -- AzureMssqlServer is a prerequisite because every elastic pool lives on a referenced logical server (the server's resource group is transitive).
- `AzureRedisLinkedServer` -- The target and linked caches are referenced via ARM ids, not auto-deployed: caches are the slowest-provisioning resources in the Azure catalog and their names are globally unique, so E2E scenarios declare their own cache fixtures instead of a registry prerequisite recreating a shared one per run.
- `AzureRedisCacheAccessPolicy` -- The parent cache is referenced via redis_cache_id, not auto-deployed: caches are the slowest-provisioning resources in the Azure catalog and their names are globally unique, so E2E scenarios declare their own cache fixtures instead of a registry prerequisite recreating a shared one per run.
- `AzureRedisCacheAccessPolicyAssignment` -- The parent cache is referenced via redis_cache_id, not auto-deployed: caches are the slowest-provisioning resources in the Azure catalog and their names are globally unique, so E2E scenarios declare their own cache fixtures instead of a registry prerequisite recreating a shared one per run.
- `AzureContainerAppEnvironment` -- AzureResourceGroup is a prerequisite because the environment is created inside a referenced resource group that must already exist.
- `AzureContainerApp` -- AzureContainerAppEnvironment is a prerequisite because every app runs inside a referenced environment (the resource group arrives transitively through it).
- `AzureServicePlan` -- AzureResourceGroup is a prerequisite because the plan is created inside a referenced resource group that must already exist.
- `AzureFunctionApp` -- AzureServicePlan is a prerequisite because a function app runs on a referenced plan (the resource group arrives transitively through the plan). The required storage account is deliberately NOT a registry prerequisite: storage-account names are globally unique, so scenarios bring their own scenario-local account fixtures.
- `AzureLinuxWebApp` -- AzureServicePlan is a prerequisite because a web app runs on a referenced plan (the resource group arrives transitively through the plan).
- `AzureContainerAppJob` -- AzureContainerAppEnvironment is a prerequisite because a job runs inside a referenced environment (the resource group arrives transitively through it).
- `AzureContainerAppEnvironmentStorage` -- AzureContainerAppEnvironment is a prerequisite because the storage registration lives on a referenced environment. The Azure Files share and storage account are deliberately NOT registry prerequisites: storage-account names are globally unique, so scenarios bring their own scenario-local account + share fixtures.
- `AzureContainerAppEnvironmentDaprComponent` -- AzureContainerAppEnvironment is a prerequisite because the Dapr component is registered on a referenced environment.
- `AzureContainerAppEnvironmentCertificate` -- AzureContainerAppEnvironment is a prerequisite because the certificate is stored on a referenced environment.
- `AzureContainerAppEnvironmentManagedCertificate` -- AzureContainerAppEnvironment is a prerequisite because the managed certificate is provisioned on a referenced environment.
- `AzureLogAnalyticsWorkspace` -- AzureResourceGroup is a prerequisite because the workspace is created inside a referenced resource group that must already exist.
- `AzureApplicationInsights` -- AzureLogAnalyticsWorkspace is a prerequisite because workspace-based Application Insights stores its telemetry in a referenced workspace (the resource group chains transitively through the workspace).
- `AzureMonitorDiagnosticSetting` -- AzureLogAnalyticsWorkspace is a prerequisite because the setting's scenarios route a fixture workspace's telemetry (the workspace doubles as target and destination); the target itself is polymorphic.
- `AzureMonitorActionGroup` -- AzureResourceGroup is a prerequisite because the action group is created inside a referenced resource group that must already exist.
- `AzureMonitorMetricAlert` -- AzureMonitorActionGroup is a prerequisite because a metric alert's actions fire into a referenced action group (the resource group chains transitively); alert scopes are polymorphic.
- `AzureMonitorScheduledQueryAlert` -- AzureLogAnalyticsWorkspace is a prerequisite because the rule queries a referenced workspace scope; AzureMonitorActionGroup because its action fires into a referenced action group.
- `AzureMonitorActivityLogAlert` -- AzureMonitorActionGroup is a prerequisite because an activity log alert's actions fire into a referenced action group (the resource group chains transitively). The alert itself is subscription-global and its scopes are polymorphic.
- `AzureApplicationInsightsStandardWebTest` -- AzureApplicationInsights is a prerequisite because a standard web test binds to a referenced Application Insights component (the resource group chains transitively through the kind).
- `AzureUserAssignedIdentity` -- AzureResourceGroup is a prerequisite because the identity is created inside a referenced resource group that must already exist.
- `AzureRoleAssignment` -- AzureResourceGroup and AzureUserAssignedIdentity are prerequisites because an assignment grants a role at a referenced scope (most commonly a resource group) to a referenced principal (most commonly a managed identity) -- both must exist before the grant can be written.
- `AzureRoleDefinition` -- AzureResourceGroup is a prerequisite because a custom role definition is created at a referenced scope, most commonly a resource group in composed environments -- the scope must exist before the definition can be written.
- `AzureFederatedIdentityCredential` -- AzureUserAssignedIdentity is the prerequisite because a federated identity credential is a child resource of a referenced managed identity -- the identity must exist before the credential can be written on it. (The resource group arrives transitively through the identity's own prerequisite declaration.)
- `AzureServiceBusNamespace` -- AzureResourceGroup is a prerequisite because a Service Bus namespace is created inside a referenced resource group in composed environments. The namespace is the container every Service Bus messaging entity (queue, topic, subscription, authorization rule, geo-DR pairing) nests under. The child kinds deliberately declare NO registry prerequisite: namespace names are globally unique with a post-delete name hold, so E2E composes them with scenario-local namespace fixtures instead of a shared recreate-per-scenario prerequisite.
- `AzureEventHubNamespace` -- AzureResourceGroup is a prerequisite because an Event Hub namespace is created inside a referenced resource group in composed environments. The namespace is the container every Event Hubs entity (event hub, consumer group, authorization rule, schema group, geo-DR pairing, customer-managed key) nests under. The child kinds deliberately declare NO registry prerequisite: namespace names are globally unique with a post-delete name hold, so E2E composes them with scenario-local namespace fixtures instead of a shared recreate-per-scenario prerequisite.
- `AzureServiceBusQueue`
- `AzureServiceBusTopic`
- `AzureServiceBusSubscription`
- `AzureServiceBusAuthorizationRule`
- `AzureServiceBusDisasterRecoveryConfig`
- `AzureEventHub`
- `AzureEventHubConsumerGroup`
- `AzureEventHubAuthorizationRule`
- `AzureFrontDoorProfile` -- AzureResourceGroup is a prerequisite because a Front Door profile is created inside a referenced resource group in composed environments. The profile is the container every Front Door delivery resource (endpoint, origin group, origin, route) nests under.
- `AzureFrontDoorEndpoint` -- AzureFrontDoorProfile is a prerequisite because an endpoint is an ARM child of a referenced profile -- the profile must exist before the endpoint can be written. (The resource group arrives transitively through the profile's own prerequisite declaration.)
- `AzureFrontDoorOriginGroup` -- AzureFrontDoorProfile is a prerequisite because an origin group is an ARM child of a referenced profile. A container kind: every origin is an ARM child of its origin group, so on a diagram the origins stand inside the group that load-balances them (the profile's room holds the group). A route or a rule that forwards TO an origin group lives in its endpoint or its rule set, never in the group, and those references say so with containment_exempt.
- `AzureFrontDoorOrigin` -- AzureFrontDoorOriginGroup is a prerequisite because an origin is an ARM child of a referenced origin group (the profile and resource group chain transitively).
- `AzureFrontDoorRoute` -- A route attaches to an endpoint (its ARM parent) and forwards to an origin group whose origins must exist before ARM accepts the route -- so both the endpoint and the origin chain are genuine deploy-order prerequisites.
- `AzureFrontDoorRuleSet` -- AzureFrontDoorProfile is a prerequisite because a rule set is an ARM child of a referenced profile. The rules live inside the set (they form one ordered policy document); routes attach the set by ARM ID.
- `AzureFrontDoorCustomDomain` -- AzureFrontDoorProfile is a prerequisite because a custom domain is an ARM child of a referenced profile. The DNS zone and (for bring-your-own certificates) the Front Door secret are optional references, not deploy-order prerequisites.
- `AzureFrontDoorSecret` -- AzureFrontDoorSecret is a prerequisite-light kind: only the profile (its ARM parent) must exist. The Key Vault certificate it wraps is a reference resolved before the module runs; its vault chain is exercised through scenario-local fixtures in E2E.
- `AzureFrontDoorFirewallPolicy` -- AzureResourceGroup is a prerequisite because the Front Door WAF policy is created inside a referenced resource group -- it is a GLOBAL resource, not a profile child (a different ARM type than the regional Application Gateway WAF policy). Security policies attach it to profiles; the policy itself depends on nothing else.
- `AzureFrontDoorSecurityPolicy` -- A security policy is an ARM child of a profile that associates a referenced WAF policy with referenced domains -- so the endpoint (the default-domain association target; the profile arrives transitively through it) and the WAF policy are genuine deploy-order prerequisites.
- `AzureStorageContainer` -- None of the storage data-service kinds declares a registry prerequisite on AzureStorageAccount: account names are GLOBALLY unique and Azure holds a just-deleted name, so a recreate-per-scenario fixture would hang -- their E2E scenarios declare scenario-local account fixtures instead. Deploy ordering in composed environments still flows from the storage_account_id reference itself.
- `AzureStorageShare`
- `AzureStorageQueue`
- `AzureStorageTable`
- `AzureStorageEncryptionScope`
- `AzureStorageDataLakeGen2Filesystem`
- `AzureStorageLocalUser`
- `AzureStorageObjectReplication`
- `AzureCosmosdbSqlDatabase` -- None of the Cosmos DB data-service kinds declares a registry prerequisite on AzureCosmosdbAccount: account names are GLOBALLY unique DNS labels, so a recreate-per-scenario fixture would risk name-reuse hangs -- their E2E scenarios declare scenario-local account fixtures instead. Deploy ordering in composed environments still flows from the cosmosdb_account_id / parent-database references themselves.
- `AzureCosmosdbSqlContainer`
- `AzureCosmosdbMongoDatabase`
- `AzureCosmosdbMongoCollection`
- `AzureCosmosdbSqlRoleDefinition`
- `AzureCosmosdbSqlRoleAssignment`
- `AzureManagedRedis` -- AzureResourceGroup is the cluster's only registry prerequisite: the cluster is created inside a referenced resource group. The geo-replication and access-policy-assignment children declare NO prerequisite on AzureManagedRedis: clusters are expensive, slow-provisioning parents, so their E2E scenarios declare scenario-local cluster fixtures instead of recreating a shared one per scenario. Deploy ordering in composed environments still flows from the managed_redis_id references themselves.
- `AzureManagedRedisGeoReplication`
- `AzureManagedRedisAccessPolicyAssignment`
- `AzureEventHubDisasterRecoveryConfig`
- `AzureEventHubSchemaGroup`
- `AzureEventHubCluster` -- AzureResourceGroup is a prerequisite because a dedicated Event Hubs cluster is created inside a referenced resource group in composed environments. Note: clusters cannot be deleted for 4 hours after creation (Azure's moratorium), so E2E treats this kind as offline-gated.
- `AzureEventHubNamespaceCustomerManagedKey`
- `AzureMssqlFailoverGroup` -- AzureMssqlServer is a prerequisite because a failover group is created on a referenced primary logical server and points at a partner server; the primary (and its resource group, which chains transitively) must exist before the group can be written.
- `AzureContainerAppCustomDomain` -- AzureContainerApp is a prerequisite because the domain binding lives in a referenced app's ingress configuration (the environment and resource group chain transitively through the app).
- `AzureFirewallPolicy`
- `AzureFirewallPolicyRuleCollectionGroup` -- AzureFirewallPolicy is a prerequisite because a rule collection group is a child document of a referenced policy (the resource group chains transitively through the policy).
- `AzureFirewall` -- AzureSubnet is a prerequisite because a VNet-deployed firewall's data path lives in a dedicated subnet that must be named exactly "AzureFirewallSubnet" (the virtual network and resource group chain transitively through the subnet). The E2E install profile publishes a fixture subnet with that exact name and a /26 prefix.
- `AzureIpGroup`
- `AzureVirtualNetworkGateway` -- AzureSubnet is a prerequisite because every virtual network gateway lives in a dedicated subnet named exactly "GatewaySubnet" (the virtual network and resource group chain transitively through the subnet); the subnet install profile publishes a fixture instance with that exact ARM name. AzurePublicIp is a prerequisite because a VPN-type gateway (the default shape) requires a public IP per ip configuration; the address install profile publishes a dedicated zone-redundant instance (a gateway binds its address exclusively, and the AZ gateway SKUs require zones on it).
- `AzureVirtualNetworkGatewayConnection` -- Both gateways are prerequisites: a connection joins a virtual network gateway to a far side, and the site-to-site far side is a local network gateway (the GatewaySubnet, VNet, and resource group chain transitively through the virtual network gateway).
- `AzureLocalNetworkGateway`
- `AzurePrivateLinkService` -- AzureSubnet is the sole prerequisite: every NAT ip configuration draws its address from a subnet with private-link-service network policies disabled (the subnet install profile publishes a fixture instance with that flag off). The Standard load balancer whose frontend the service typically fronts is NOT a registry prerequisite -- the spec's destination is an exactly-one-of (load balancer frontend OR fixed destination IP), so scenarios that use the load-balancer shape declare it via the planton.dev/e2e-prerequisites annotation instead.
- `AzureExpressRouteCircuit`
- `AzureExpressRouteCircuitPeering` -- The circuit is the prerequisite: a peering is an ARM child of the circuit, addressed by the circuit's name (the resource group chains transitively through the circuit).
- `AzureExpressRouteGateway` -- The hub is the prerequisite: ARM requires an ExpressRoute Gateway to be deployed INTO a Virtual WAN hub (the WAN and resource group chain transitively through the hub).
- `AzureExpressRoutePort` -- ExpressRoute Port: your own physical port pair on a Microsoft edge router (ExpressRoute Direct), from whose bandwidth circuits are carved. Self-contained -- only the resource group is required.
- `AzureVirtualWan` -- Virtual WAN: the umbrella of Azure's managed hub-and-spoke networking, under which virtual hubs and their gateways are created. Self-contained -- only the resource group is required. A container: its hubs and its branch sites are created into it, so on a diagram the WAN is the room they stand in.
- `AzureVirtualHub` -- The WAN is the prerequisite: this kind models the Virtual WAN hub (virtual_wan_id is required; standalone hubs are the legacy Route Server construction, which has its own ARM surface). The resource group chains transitively through the WAN. A container: ARM deploys the hub's VPN, ExpressRoute, and point-to-site gateways and its spoke connections INTO the hub, so on a diagram the hub is the room they stand in.
- `AzureVirtualHubConnection` -- Both sides of the attachment are prerequisites: the hub being joined and the spoke virtual network being attached.
- `AzureVpnGateway` -- The hub is the prerequisite: ARM deploys a Virtual WAN VPN gateway INTO a virtual hub (virtual_hub_id is required and immutable; the WAN and resource group chain transitively through the hub). ARM allows one VPN gateway per hub.
- `AzureVpnGatewayConnection` -- Both ends of the tunnel are prerequisites: a connection is an ARM child of the VPN gateway and pins each of its links to a specific link of the remote VPN site (the hub, WAN, and resource group chain transitively through the gateway).
- `AzureVpnSite` -- The WAN is the prerequisite: a VPN site is the Virtual WAN world's address-book entry for one branch location (virtual_wan_id is required; the classic-world sibling without a WAN is AzureLocalNetworkGateway). The resource group chains transitively through the WAN.
- `AzurePointToSiteVpnGateway` -- The hub and the server configuration are both prerequisites: a point-to-site VPN gateway deploys INTO a virtual hub (one P2S gateway per hub, a slot separate from the hub's site-to-site VPN gateway) and is born pointing at the VPN server configuration that defines how its users authenticate -- both ARM-required and fixed at creation. The WAN and resource group chain transitively through the hub.
- `AzureVpnServerConfiguration` -- Self-contained -- only the resource group is required: a VPN server configuration is the reusable "who may connect and how" authentication policy (Entra ID / certificate / RADIUS) that point-to-site VPN gateways attach to; it references no other Azure resource.
- `AzureCognitiveAccount` -- Self-contained -- only the resource group is required: an Azure AI services account (Azure OpenAI, the multi-service AIServices account, the single-service accounts) needs no other Azure resource; subnets (network rules), Key Vault keys (CMK), storage accounts and user-assigned identities are optional references. A container kind: every model deployment and every project is an ARM child created onto the account (accounts/{account}/deployments/{name}, .../projects/{name}), so a diagram draws them inside it. The subnets the account admits through its network ACLs or injects agent compute into are access, never placement (containment_exempt on those fields).
- `AzureCognitiveDeployment` -- An ARM child of its account: a model deployment (which model runs, at which throughput class) exists only on an Azure AI services account of kind "OpenAI" or "AIServices".
- `AzureCognitiveAccountProject` -- An ARM child of its account: an AI Foundry project exists only on an "AIServices"-kind account with project management enabled.
- `AzureMachineLearningWorkspace` -- The workspace REQUIRES all three companion services at creation (default storage, secrets vault, telemetry) -- genuine deploy-order prerequisites, each with its own fixture profile. A container kind: datastores, compute clusters and instances, and online and batch endpoints are ARM children of the workspace (workspaces/{ws}/datastores|computes|onlineEndpoints|batchEndpoints/{name}), so a diagram draws them inside it. The companion services it names are reaches (containment_exempt on those fields), and so is the subnet its serverless compute or a compute's nodes attach to.
- `AzureMachineLearningDatastore` -- An ARM child of its workspace. The storage target (container, filesystem or share) is scenario-declared via the e2e-prerequisites annotation -- only the blob scenario needs a container, so it is not a kind-wide prerequisite.
- `AzureMachineLearningComputeCluster` -- An ARM child of its workspace (.../computes/{name}) -- the auto-scaling pool of VMs training jobs run on.
- `AzureMachineLearningComputeInstance` -- An ARM child of its workspace (.../computes/{name}) -- a single always-on VM serving as one data scientist's cloud workstation.
- `AzureAiFoundry` -- The hub REQUIRES both companion services at creation (secrets vault, default storage) -- genuine deploy-order prerequisites, each with its own fixture profile. A container kind: every AI Foundry project is created inside its hub and deploys into the hub's resource group (the project spec carries none), so a diagram draws the projects inside the hub. The vault and storage account the hub names are reaches (containment_exempt on those fields).
- `AzureAiFoundryProject` -- Deploys into its hub's resource group (the provider derives the group from the hub reference -- the project spec carries none).
- `AzureSearchService`
- `AzureMachineLearningOnlineEndpoint` -- An ARM child of its workspace (.../onlineEndpoints/{name}) -- the stable scoring address applications call. azurerm carries no ML endpoint resources; the modules write the raw ARM shape at a pinned api-version (azapi / azure-native). A container kind in its turn: every online deployment is an ARM child created under the endpoint whose traffic map routes to it (.../onlineEndpoints/{endpoint}/deployments/{name}) and cannot exist without it, so a diagram draws the deployments inside the endpoint, inside the workspace.
- `AzureMachineLearningOnlineDeployment` -- An ARM child of its endpoint (.../deployments/{name}) -- the running copy of a model the endpoint's traffic map routes to.
- `AzureMachineLearningBatchEndpoint` -- An ARM child of its workspace (.../batchEndpoints/{name}) -- the stable address batch scoring jobs are submitted to. azurerm carries no ML endpoint resources; the modules write the raw ARM shape at a pinned api-version (azapi / azure-native). A container kind in its turn: every batch deployment is an ARM child created under the endpoint whose default-deployment pointer routes to it (.../batchEndpoints/{endpoint}/deployments/{name}), so a diagram draws the deployments inside the endpoint, inside the workspace. The compute cluster a deployment runs on is a reach, not a room.
- `AzureMachineLearningBatchDeployment` -- An ARM child of its endpoint (.../deployments/{name}) -- the job recipe (model, compute, batching behavior) the endpoint's default-deployment pointer routes submissions to.
- `AzureRecoveryServicesVault` -- The Recovery Services vault (Microsoft.RecoveryServices/vaults) -- the safe that classic Azure Backup data and Site Recovery configuration live in. Backup policies and protected items are ARM children of a vault, so the vault is a container kind: a diagram draws them inside it, as the portal's own blades do.
- `AzureBackupPolicyVm` -- An ARM child of its vault (.../backupPolicies/{name}) -- the schedule and retention rules that govern IaaS VM backups.
- `AzureBackupProtectedVm` -- An ARM child of its vault (.../protectedItems/...) -- the binding that puts one virtual machine under a backup policy's protection.
- `AzureBackupPolicyFileShare` -- An ARM child of its vault (.../backupPolicies/{name}) -- the schedule and retention rules that govern Azure Files share backups (snapshot or vaulted).
- `AzureBackupProtectedFileShare` -- An ARM child of its vault (.../protectedItems/AzureFileShare;...) -- the binding that puts one Azure Files share under a backup policy's protection. The share's storage account must already be registered with the vault (AzureBackupContainerStorageAccount). Prerequisite ORDER is load-bearing for teardown (destroy runs in reverse): the registration must list AFTER the share so it unregisters FIRST -- Azure Backup holds a DoNotDelete lock on a registered storage account, and a share delete under that lock fails ScopeLocked.
- `AzureDataProtectionBackupVault` -- The Data Protection backup vault (Microsoft.DataProtection/ backupVaults) -- the safe that MODERN Azure Backup data lives in (managed disks, blob storage, AKS clusters, MySQL/PostgreSQL flexible servers, Data Lake storage). Backup policies and backup instances are ARM children of a vault, so the vault is a container kind: a diagram draws them inside it, as the portal's own blades do.
- `AzureDataProtectionBackupPolicy` -- An ARM child of its vault (.../backupPolicies/{name}) -- the schedule and retention rules for ONE Data Protection datasource type (blob storage, disk, Kubernetes cluster, MySQL/PostgreSQL flexible server, or Data Lake storage), modeled as one kind with variant blocks.
- `AzureDataProtectionBackupInstance` -- An ARM child of its vault (.../backupInstances/{name}) -- the binding that puts ONE datasource (a managed disk, a storage account's blob services, an AKS cluster, a MySQL/PostgreSQL flexible server, or a Data Lake storage account) under a Data Protection backup policy, modeled as one kind with variant blocks. The vault's managed identity must hold the datasource roles Azure Backup requires BEFORE the instance is created.
- `AzureBastionHost` -- AzureSubnet and AzurePublicIp are prerequisites because a dedicated-infrastructure Bastion host (Basic/Standard/Premium -- the default shapes) deploys into a subnet named exactly "AzureBastionSubnet" and binds a Standard static public IP EXCLUSIVELY (the virtual network and resource group chain transitively through the subnet). The Developer SKU instead attaches to a virtual network directly and uses neither.
- `AzureNetworkWatcherFlowLog` -- AzureVirtualNetwork and AzureStorageAccount are prerequisites because a flow log records a network-scoped target (a virtual network in the common case; subnets and network interfaces chain through the network) into a referenced storage account. The regional Network Watcher parent is NOT a prerequisite: Azure auto-creates it ("NetworkWatcher_{region}" in "NetworkWatcherRG") the moment the region hosts a virtual network, and the flow log references it by name. Traffic Analytics' Log Analytics workspace is an optional arm, declared by scenarios that use it.
- `AzurePrivateDnsResolver` -- AzureVirtualNetwork and AzureSubnet are prerequisites because a DNS Private Resolver anchors to a referenced virtual network (at most ONE resolver per network -- Azure enforces it) and each of its inbound/outbound endpoints occupies its own dedicated subnet delegated to "Microsoft.Network/dnsResolvers" (the resource group chains transitively through the network and subnets).
- `AzurePrivateDnsResolverForwardingRuleset` -- AzurePrivateDnsResolver is a prerequisite because a DNS forwarding ruleset steers a resolver's OUTBOUND endpoints -- it binds their ARM ids (at most 2, same resolver) at creation. (The resource group and network chain transitively through the resolver's own prerequisite declarations.)
- `AzurePrivateDnsRecord` -- AzurePrivateDnsZone is a prerequisite because every record set is created inside a referenced private DNS zone (the resource group chains transitively through the zone's own prerequisite).
- `AzureTrafficManagerProfile` -- AzureResourceGroup is a prerequisite because a Traffic Manager profile is created inside a referenced resource group (the profile itself is a global service -- the group only holds its metadata record). A container kind: every endpoint is an ARM child of its profile, so on a diagram the endpoints stand inside the profile that steers traffic to them. A nested endpoint that points AT another profile lives in its own parent profile, never in the one it targets, and that reference says so with containment_exempt.
- `AzureTrafficManagerEndpoint` -- AzureTrafficManagerProfile is a prerequisite because every endpoint is created inside a referenced profile -- it is the destination a profile steers traffic to (the resource group chains transitively through the profile's own prerequisite).
- `AzureMonitorAutoscaleSetting` -- AzureResourceGroup is a prerequisite because an autoscale setting is created inside a referenced resource group. The scalable TARGET it controls is a no-default reference (many kinds can be scaled), so no target kind is declared here -- scenarios declare their own target fixture.
- `AzureMonitorDataCollectionRule` -- The Azure Monitor data collection rule (DCR) -- the routing table declaring what telemetry the Azure Monitor Agent collects and where it lands. AzureResourceGroup is a prerequisite because a rule is created inside a referenced resource group; AzureLogAnalyticsWorkspace because a workspace is the canonical destination a rule routes to (the smoke scenario's shape). Machines attach to a rule with AzureMonitorDataCollectionRuleAssociation resources.
- `AzureEventgridTopic` -- The Azure Event Grid custom topic -- the HTTPS endpoint an application publishes its own events to, fanned out to handlers by event subscriptions. One topic is one event stream with its own endpoint and access keys; for many streams behind one endpoint see AzureEventgridDomain. A container kind: every event subscription on the topic is an ARM child created under it ({topic_id}/providers/ Microsoft.EventGrid/eventSubscriptions/{name}) and cannot exist without it, so on a diagram the topic is the room its subscriptions stand in, as the portal's own blade lists them. A kind that merely publishes to or subscribes from a topic while living elsewhere (a Data Factory trigger, an Event Grid namespace's MQTT route) says so with containment_exempt.
- `AzureEventgridDomain` -- The Azure Event Grid domain -- ONE publishing endpoint and one pair of access keys serving many event streams (domain topics), the multi-tenant pattern. Topics inside the domain are auto-managed by Azure or declared explicitly as AzureEventgridDomainTopic resources. A container kind: every domain topic is an ARM child of its domain ({domain_id}/topics/{name}), so on a diagram the domain is the room its tenant streams stand in.
- `AzureEventgridSystemTopic` -- The Azure Event Grid system topic -- the subscription surface for events AZURE ITSELF publishes about one of your resources (a storage account's blob events, a resource group's lifecycle events). One system topic per source resource per topic type; event subscriptions attach to it to route events to handlers. A container kind: every subscription on a system topic is an ARM child created under it ({system_topic_id}/eventSubscriptions/{name}), so on a diagram the system topic is the room its subscriptions stand in.
- `AzureEventgridEventSubscription` -- The Azure Event Grid event subscription -- the delivery instruction routing events from a source (a custom topic, domain, domain topic, system topic, resource group, or subscription) to a handler (a Function, Event Hub, Service Bus queue/topic, storage queue, hybrid connection, or webhook), with filtering, retry, and dead-letter behavior.
- `AzureEventgridNamespace` -- The Azure Event Grid namespace -- the capacity-scaled hub of the newer Event Grid: hosts CloudEvents namespace topics and an optional MQTT broker behind one set of regional endpoints, sized in throughput units. A container kind: every namespace topic is an ARM child of its namespace ({namespace_id}/topics/{name}), so on a diagram the namespace is the room its streams stand in.
- `AzureDataFactory` -- The Azure Data Factory -- the workspace every other Data Factory resource lives inside: pipelines, data flows, linked services, datasets, triggers, and integration runtimes are all created against a factory's ARM ID. A container kind for exactly that reason: those six kinds are ARM children of the factory, so on a diagram the factory is the room they stand in.
- `AzureDataFactoryPipeline` -- One unit of work inside an Azure Data Factory ({factory_id}/pipelines/{name}) -- an ordered set of activities that executes as a whole when triggered.
- `AzureDataFactoryDataFlow` -- A Data Factory data flow ({factory_id}/dataflows/{name}) -- a visually-designed data transformation executed on managed Spark, or, as a flowlet, a reusable snippet other data flows embed. One kind covers both provider forms (they share one schema and one name namespace inside the factory).
- `AzureDataFactoryLinkedService` -- A Data Factory linked service ({factory_id}/linkedservices/{name}) -- a saved connection in the factory's address book: where an external system lives and how to authenticate to it. One kind covers every connection type Azure models as a first-class resource (storage, SQL family, Cosmos DB, Databricks, Key Vault, SFTP, web APIs, and more) as variants in one factory-scoped name namespace, plus the raw-JSON custom form.
- `AzureDataFactoryDataset` -- A Data Factory dataset ({factory_id}/datasets/{name}) -- a named view of data inside a system a linked service already connects to: which container and path, which table, which file format. One kind covers every data shape Azure models as a first-class dataset resource (delimited text/CSV, JSON, Parquet, binary, blob, HTTP, the SQL family, Snowflake, Cosmos DB) as variants in one factory-scoped name namespace, plus the raw-JSON custom form.
- `AzureDataFactoryTrigger` -- A Data Factory trigger ({factory_id}/triggers/{name}) -- the instruction that starts pipelines automatically: on a clock schedule, per contiguous tumbling window, on storage blob events, or on Event Grid custom events. One kind covers all four provider trigger resources as variants (one ARM namespace, one started/stopped lifecycle).
- `AzureDataFactoryIntegrationRuntime` -- A Data Factory integration runtime ({factory_id}/integrationRuntimes/{name}) -- the compute engine a factory's pipelines, data flows, and copy activities run on. One kind covers all three engine flavors as variants in one factory-scoped name namespace: the managed data-flow compute, the managed SSIS package runtime, and the self-hosted agent registration (which issues the authorization keys agents join with).
- `AzureComputeGallery` -- The Azure Compute Gallery -- the shared library an organization keeps its approved VM images in. Image definitions (AzureComputeGalleryImage) live inside it; VMs and scale sets deploy from their published, region-replicated versions. A container kind: every image definition is an ARM child of its gallery ({gallery_id}/images/{name}), so on a diagram the definitions stand inside the library that publishes them. The image ids a VM, a scale set, or a disk boots from are plain strings and never place anything inside the gallery.
- `AzureComputeGalleryImage` -- A gallery image ({gallery_id}/images/{name}) -- one image definition inside a Compute Gallery (marketplace-style identity, OS type, security posture) plus its published versions, each replicated to its own target regions. VMs deploy from a version's ARM ID or from the definition's ID to get the latest version.
- `AzureAvailabilitySet` -- The availability set -- the classic pre-zones placement grouping that spreads VMs across separate fault and update domains so one hardware failure or maintenance window cannot take them all down. VMs join the set at creation.
- `AzureDiskSnapshot` -- The managed disk snapshot -- a point-in-time copy of a disk used for backup, cloning, and as the source of gallery image versions. Incremental snapshots store only the delta since the previous snapshot of the same disk.
- `AzureContainerInstance` -- The Azure Container Instance container group -- serverless containers billed per second: one or more containers sharing a lifecycle, network, and volumes (plus one-shot init containers), with no cluster or VM to manage. Public, subnet-private, or IP-less postures.
- `AzureFunctionAppFlexConsumption` -- The Azure Function App on the Flex Consumption plan -- Azure's newest serverless Functions hosting model: per-instance memory selection, a configurable scale-out ceiling, always-ready instance pools, and explicit blob-container deployment storage. Requires an FC1-SKU service plan, which is deliberately NOT a registry prerequisite: the shared plan fixture serves the classic app tiers, and an FC1 plan is cheap to create per scenario (no idle compute cost), so scenarios bring their own plan fixture -- the same reasoning that keeps the globally-unique storage account scenario-local for AzureFunctionApp.
- `AzureMongoCluster` -- The Azure Cosmos DB for MongoDB vCore cluster -- Azure's modern managed MongoDB: a real MongoDB engine on dedicated vCore tiers with sharding, zone-redundant HA, and point-in-time restore.
- `AzureFabricCapacity` -- The Microsoft Fabric capacity -- the billing and compute anchor of Microsoft Fabric: workspaces assign themselves to a capacity, and its F-SKU sets how much compute every workload on it shares. azurerm's entire Fabric surface is this one resource (workspaces and items live in Microsoft's dedicated fabric provider).
- `AzureBackupContainerStorageAccount` -- Registers a storage account with a Recovery Services vault as a backup container (.../protectionContainers/StorageContainer;...) -- one registration per storage-account-and-vault pair, required BEFORE any of the account's file shares can be protected. Part of the backup family (2175-2179) despite the out-of-run number -- enum numbers are pinned by the registry snapshot; never renumber.
- `AzureDataProtectionResourceGuard` -- The Data Protection Resource Guard (Microsoft.DataProtection/ resourceGuards) -- the approval gate behind Multi-User Authorization: privileged vault operations (disabling soft delete, reducing retention) require an approval through a guard, which typically lives in a DIFFERENT administrator's scope. Vaults reference a guard by its ARM ID. Part of the backup family (2175-2182) despite the out-of-run number -- enum numbers are pinned by the registry snapshot; never renumber.
- `AzurePrivateDnsResolverVirtualNetworkLink` -- The attachment that makes a DNS forwarding ruleset take effect in one virtual network ({ruleset_id}/virtualNetworkLinks/{name}) -- one link per ruleset-network pair, up to 500 per ruleset, spokes joining and leaving independently (which is why the link is a standalone kind, exactly like AzurePrivateDnsZoneVirtualNetworkLink). Part of the DNS Private Resolver family (2186-2187) despite the out-of-run number -- enum numbers are pinned by the registry snapshot; never renumber.
- `AzureMonitorDataCollectionRuleAssociation` -- The attachment that puts ONE machine under an Azure Monitor data collection rule ({target_id}/providers/Microsoft.Insights/ dataCollectionRuleAssociations/{name}) -- an extension resource on the TARGET machine, many per rule, machines joining and leaving monitoring independently (which is why the association is a standalone kind, exactly like AzurePrivateDnsZoneVirtualNetworkLink). AzureVirtualMachine is a prerequisite because the smoke scenario attaches the fixture VM; the rule prerequisite chains the rule's own install manifest. Part of the Monitor family (2191-2192) despite the out-of-run number -- enum numbers are pinned by the registry snapshot; never renumber.
- `AzureEventgridDomainTopic` -- One named event stream inside an Azure Event Grid domain ({domain_id}/topics/{name}) -- the per-tenant mailbox of the multi-tenant pattern: many per domain, each with its own subscriptions and lifecycle, tenants joining and leaving without touching the domain (which is why the domain topic is a standalone kind, exactly like AzureEventHubConsumerGroup on a shared hub). Part of the Event Grid family (2193-2194) despite the out-of-run number -- enum numbers are pinned by the registry snapshot; never renumber.
- `AzureEventgridNamespaceTopic` -- One named CloudEvents stream inside an Azure Event Grid namespace ({namespace_id}/topics/{name}) -- many per namespace, publishers and teams creating and deleting their own against the shared namespace (which is why the topic is a standalone kind, exactly like AzureEventgridDomainTopic and AzureEventHubConsumerGroup). Part of the Event Grid family (2193-2197) despite the out-of-run number -- enum numbers are pinned by the registry snapshot; never renumber.
- `AzureMongoClusterUser` -- Grants one Microsoft Entra principal access to an Azure Cosmos DB for MongoDB vCore cluster ({cluster_id}/users/{object_id}) -- an access binding, not a password user: many per cluster, principals joining and leaving independently (which is why the grant is a standalone kind, the access-grant class of AzureRoleAssignment). Part of the Mongo vCore family (2211) despite the out-of-run number -- enum numbers are pinned by the registry snapshot; never renumber.
- `AzurePlantonRunner` -- AzureContainerAppEnvironment is a prerequisite because the runner appliance is a Container App, and every Container App runs inside an environment -- the environment reference must resolve before the appliance can deploy.
- `GcpArtifactRegistryRepo` -- 3000–3999: GCP resources
- `GcpTargetHttpsProxy` -- The URL map is the parent a proxy cannot exist without; the classic compute certificate kinds and the SSL policy are the fixture parents the committed scenarios attach. The Certificate Manager certificate list (certificate_manager_certificates, honored only by the cross-region internal ALB) is optional composition -- a scenario that arms it declares GcpCertManagerCert via the e2e-prerequisites annotation, never a registry edge that would tax every proxy and forwarding-rule chain.
- `GcpCloudFunction`
- `GcpCloudRun`
- `GcpCloudSql`
- `GcpDnsZone`
- `GcpGcsBucket`
- `GcpGkeCluster`
- `GcpIamCustomRole`
- `GcpProject`
- `GcpVpcNetwork`
- `GcpSubnetwork`
- `GcpRouterNat`
- `GcpGkeNodePool`
- `GcpServiceAccount`
- `GcpGkeWorkloadIdentityBinding`
- `GcpCertManagerCert`
- `GcpComputeInstance`
- `GcpDnsRecord`
- `GcpProjectIamMember`
- `GcpFirewallRule`
- `GcpGlobalAddress`
- `GcpCloudArmorPolicy`
- `GcpHealthCheck`
- `GcpBackendBucket`
- `GcpBackendService`
- `GcpRegionNetworkEndpointGroup`
- `GcpUrlMap`
- `GcpManagedSslCertificate`
- `GcpTargetHttpProxy`
- `GcpAlloydbCluster`
- `GcpRedisInstance`
- `GcpFirestoreDatabase`
- `GcpSpannerInstance`
- `GcpSpannerDatabase`
- `GcpBigtableInstance`
- `GcpMemorystoreInstance`
- `GcpCloudSqlDatabase`
- `GcpCloudSqlUser`
- `GcpAlloydbInstance`
- `GcpAlloydbUser`
- `GcpSpannerBackupSchedule`
- `GcpBigtableTable`
- `GcpFirestoreBackupSchedule`
- `GcpFirestoreIndex`
- `GcpBigQueryDataset`
- `GcpDataprocCluster`
- `GcpDataprocAutoscalingPolicy`
- `GcpBigQueryTable`
- `GcpBigQueryCapacityCommitment` -- A BigQuery capacity commitment: slots bought for a fixed term in an administration project and location, pooled across every reservation there. A purchase Google will not delete before its term ends.
- `GcpBigQueryReservationGroup` -- A BigQuery reservation group: reservations that share idle slots with each other first. Reservations reference it.
- `GcpDatastreamConnectionProfile` -- A Datastream connection profile: where one source database or destination is and how Datastream signs in. Streams reference a source and a destination profile; one profile serves many streams.
- `GcpDatastreamPrivateConnection` -- A Datastream private connection: the VPC peering or Private Service Connect interface through which Datastream reaches private databases, shared by every profile that reaches that network.
- `GcpPubSubTopic`
- `GcpPubSubSubscription`
- `GcpCloudTasksQueue`
- `GcpCloudSchedulerJob`
- `GcpPubSubSchema`
- `GcpPubSubTopicIamMember` -- One additive grant on a topic. Its own kind, not a field on the topic, because the identities that most need it (a logging sink's writer, a Security Command Center export's publisher) belong to resources that name the topic themselves; a grant on the topic that referenced them back would be a dependency cycle.
- `GcpVertexAiNotebook`
- `GcpVertexAiEndpoint`
- `GcpVertexAiIndex`
- `GcpVertexAiIndexEndpoint` -- Vector Search IndexEndpoint — distinct from the online-prediction GcpVertexAiEndpoint (671); different GCP resources, different kinds.
- `GcpVertexAiDeployedIndex`
- `GcpCloudComposerEnvironment`
- `GcpCloudComposerUserWorkloadsSecret`
- `GcpCloudComposerUserWorkloadsConfigMap`
- `GcpKmsKeyRing`
- `GcpKmsKey`
- `GcpKmsKeyIamMember`
- `GcpFilestoreInstance`
- `GcpWorkloadIdentityPool` -- 3101–3109: IAM/identity family (overflow block; the 3000–3022 foundation/security sub-band is fully allocated)
- `GcpWorkloadIdentityPoolProvider`
- `GcpServiceAccountIamMember`
- `GcpGcsBucketIamMember` -- One additive grant on a bucket, for a grantee that depends on the bucket itself (a logging sink writing into it): the bucket's own iam_members cannot reference such an identity without a cycle.
- `GcpGlobalForwardingRule` -- 3110–3119: networking/load-balancer family (overflow block; the 3023–3029 LB sub-band is fully allocated)
- `GcpSslPolicy`
- `GcpSslCertificate`
- `GcpServiceNetworkingConnection`
- `GcpAddress`
- `GcpServiceConnectionPolicy`
- `GcpCertManagerDnsAuthorization`
- `GcpCertificateMap` -- GcpCertManagerCert is a prerequisite because a map entry binds hostnames to EXISTING certificates — the canonical map references a certificate fixture's resource name.
- `GcpCertManagerTrustConfig` -- The CA certificates a load balancer validates client certificates against (mutual TLS). Its own kind: TLS policies reference it, never a certificate.
- `GcpCertManagerIssuanceConfig` -- How Google-managed certificates are issued from a private CA pool. Its own kind: many certificates share one config by name.
- `GcpCloudRunJob` -- 3120–3129: GCP serverless overflow
- `GcpServerlessVpcConnector`
- `GcpCloudRunWorkerPool` -- Cloud Run's no-ingress shape: a pool of always-running container instances (queue consumers, schedulers, background workers) that scales manually or by the owner's own signal instead of by requests. The proof deploys direct-VPC egress onto the prerequisite network.
- `GcpComputeDisk` -- 3130–3139: GCP compute overflow (the 3000–3022 foundation sub-band that holds GcpComputeInstance is fully allocated)
- `GcpComputeMig` -- GcpVpcNetwork is a prerequisite because the canonical group runs its fleet on a dedicated custom-mode VPC — a managed instance group's template must attach every VM to a network, and the default VPC is never assumed.
- `GcpMonitoringNotificationChannel` -- 3140–3149: GCP observability & log routing
- `GcpMonitoringAlertPolicy` -- GcpMonitoringNotificationChannel is a prerequisite because the policy's canonical shape references a channel to notify — a policy without a delivery endpoint measures but never pages.
- `GcpMonitoringUptimeCheck`
- `GcpLoggingSink` -- GcpGcsBucket is a prerequisite because the canonical sink exports to a Cloud Storage bucket — the cheapest destination that proves the whole writer-identity grant flow.
- `GcpMonitoringDashboard`
- `GcpMonitoringSlo`
- `GcpLogBucket`
- `GcpLogMetric`
- `GcpSecretManagerSecret` -- 3150–3159: GCP security & identity GcpServiceAccount is a prerequisite because the canonical secret grants secretAccessor to a workload service account — the access story the kind exists to model.
- `GcpIdentityPlatformConfig`
- `GcpIdentityPlatformTenant` -- GcpIdentityPlatformConfig is a prerequisite because tenants exist only in projects whose Identity Platform config enables multi_tenant.allow_tenants — a tenant without the initialized, tenant-enabled project config cannot be created at all.
- `GcpIamOauthClient`
- `GcpIamDenyPolicy`
- `GcpCloudRunDomainMapping` -- 3160–3169: GCP serverless edge GcpCloudRun is a prerequisite because a domain mapping exists only to point a verified domain at a running Cloud Run service — the route it maps must already exist for the mapping to be created at all.
- `GcpWorkflow`
- `GcpEventarcTrigger` -- GcpCloudRun is a prerequisite because the canonical trigger routes a Pub/Sub messagePublished event to a Cloud Run service — the destination story the kind exists to model.
- `GcpEventarcMessageBus`
- `GcpPlantonRunner`
- `GcpFolder` -- 3170–3179: GCP organization & governance (folders, org policies, tags, budgets, identity groups, API keys, KMS Autokey) GcpFolder is a container: the hierarchy node projects, sub-folders, policies, and tag bindings are placed inside.
- `GcpOrgPolicy`
- `GcpTagKey`
- `GcpTagValue`
- `GcpTagBinding`
- `GcpBillingBudget` -- A spending guardrail on a Cloud Billing account: amount, period, filters, thresholds, and where the alerts go. Lives on the billing account, so no project prerequisite; the proof lane needs a billing account the harness identity can administer.
- `GcpCloudIdentityGroup` -- A Google Group in Cloud Identity or Workspace with its memberships folded in -- the unit IAM bindings should name. Lives under a Cloud Identity customer, beside the service accounts and workload identity pools in the identity service group.
- `GcpApiKey`
- `GcpKmsAutokeyConfig` -- Cloud KMS Autokey switched on for a folder or a project: where the customer-managed keys GcpKmsKeyHandle requests are created. A security control, so it groups with the KMS kinds rather than the hierarchy.
- `GcpOrgPolicyCustomConstraint` -- A custom constraint is a DEFINITION the organization owns; the GcpOrgPolicy kinds that enforce it reference it by name, the way IAM bindings reference a custom role.
- `GcpSharedVpcHost` -- 3180–3189: GCP networking fabric (Shared VPC, VPC peering, HA VPN, firewall policies, PSC, network endpoint groups). HA VPN is two kinds: the gateway (with its Cloud Router) is declared once per VPC and region and referenced by every site connection, so two Google Cloud VPCs can point their connections at each other's gateway without a dependency cycle. Firewall policies are two kinds by scope: the hierarchical policy lives on the organization or a folder and is inherited by every network beneath its association; the network policy lives in a project and is attached to that project's VPC networks (globally or per region). Each folds its rules and associations -- a rule is keyed by priority inside its policy and an association is the edge that makes the policy act.
- `GcpSharedVpcServiceProject`
- `GcpVpcPeering`
- `GcpHaVpnGateway`
- `GcpHierarchicalFirewallPolicy`
- `GcpNetworkFirewallPolicy`
- `GcpHaVpnConnection`
- `GcpPscServiceAttachment` -- The producer half of Private Service Connect: publishes an internal load balancer's regional forwarding rule through NAT subnets so consumers in other VPCs reach it over a PSC endpoint (a regional GcpGlobalForwardingRule with an empty scheme targeting this attachment). The proof chain deploys the internal passthrough load balancer and the PSC NAT subnet as fixtures on the prerequisite network.
- `GcpNetworkEndpointGroup` -- Zonal (VM, hybrid, internet) and global (internet) network endpoint groups behind a `zone` selector; serverless, PSC, and regional internet groups stay in GcpRegionNetworkEndpointGroup. The zonal proof attaches a GcpComputeInstance fixture as an endpoint on the prerequisite network.
- `GcpRedisCluster` -- 3190–3199: GCP data (Memorystore for Redis Cluster, Managed Kafka, BigQuery connections and reservations, Datastream) Memorystore for Redis Cluster: the sharded, horizontally scaled Redis. Connectivity is Private Service Connect, placed by service connectivity automation through a GcpServiceConnectionPolicy for the gcp-memorystore-redis class on the network in the cluster's region -- the prerequisite; its proof pin carries that class beside the network and subnet pins.
- `GcpManagedKafkaCluster` -- Managed Service for Apache Kafka: a Google-operated broker fleet in one region, reachable from the VPC subnets it is attached to -- the prerequisite. Topics, ACLs, and Kafka Connect are their own kinds so the teams that own them declare them without editing the cluster.
- `GcpManagedKafkaTopic`
- `GcpManagedKafkaConnectCluster` -- Kafka Connect workers attached to a Kafka cluster: a separate project-and-location root that names the cluster it serves.
- `GcpBigQueryConnection` -- BigQuery's link to data outside its own storage (Cloud SQL, Spanner, AWS and Azure through Omni, Google resources through a managed service account, the Connector framework, Spark procedures).
- `GcpBigQueryReservation` -- A BigQuery slot reservation with the assignments that route projects, folders, or an organization onto it.
- `GcpDatastreamStream` -- A Datastream stream: continuous change data capture from one source database into BigQuery or Cloud Storage, through a source and a destination connection profile.
- `GcpRedisClusterEndpointSet` -- The Private Service Connect connections a consumer builds by hand (forwarding rules in other VPCs or projects) registered on a Redis Cluster, as one set: Google's resource replaces the cluster's whole user-created endpoint list in one write, so exactly one set per cluster is the honest grain. Its own kind because every connection names a forwarding rule that targets one of the cluster's service attachments -- a fold would depend on its own output.
- `GcpManagedKafkaAcl` -- The access rules for one resource pattern (a topic, a consumer group, a prefix, the cluster) on a Kafka cluster.
- `GcpManagedKafkaConnector` -- One data pipeline running on a Kafka Connect cluster.
- `GcpVertexAiAgentEngine` -- 3200–3229: GCP AI (Vertex AI agents and model deployments, RAG Engine, Vector Search, and the rest of the Vertex AI and generative-AI building blocks). The pre-existing Vertex AI kinds -- endpoint, index, index endpoint, deployed index, notebook -- live in the 3070s. Vertex AI Agent Engine: the managed runtime an AI agent runs in -- built from source or a container, hosted with its own identity and autoscaling, with an optional Memory Bank of long-term memories.
- `GcpVertexAiModelGardenDeployment` -- A Model Garden or Hugging Face model deployed to a Vertex AI endpoint in one step; every argument is immutable, so a change redeploys.
- `GcpVertexAiRagEngineConfig` -- The per-project, per-location tier of Vertex AI RAG Engine's managed vector database: a singleton Google owns, updated in place.
- `GcpVectorSearchCollection` -- 3203 is reserved for GcpVertexAiRagCorpus, forged when pulumi-gcp bridges provider 8.x (the resource has no SDK type today). A Vector Search collection -- a schema'd store of data objects with vector fields -- together with the approximate-nearest-neighbor indexes built over those fields (folded: one collection owns them).
- `GcpVertexAiSearchDataStore` -- Vertex AI Search (the Discovery Engine API behind the console's AI Applications / Gemini Enterprise): a data store is the corpus -- structured records, unstructured documents, or a public website -- with its schema, crawl patterns, and sitemaps folded in.
- `GcpVertexAiSearchEngine` -- The app over one or more data stores -- a search, chat, or recommendation engine -- with its serving controls, serving config, search widget, and assistants folded in. Its data stores are prerequisites: an engine cannot exist without one.
- `GcpVertexAiFeatureGroup` -- Vertex AI Feature Store: a feature group registers the features of some entities that live in a BigQuery table or view -- the features themselves folded in. Online stores serve them through feature views.
- `GcpVertexAiFeatureOnlineStore` -- The low-latency serving layer of Vertex AI Feature Store (Bigtable or Optimized storage), with the feature views it serves folded in.
- `GcpVertexAiDataset` -- A Vertex AI managed dataset: the registered container training, AutoML, labeling, and evaluation read their examples from.
- `GcpVertexAiTensorboard` -- A managed Vertex AI TensorBoard training jobs stream metrics into, with the experiments and runs declared in it folded in.
- `GcpVertexAiPersistentResource` -- A long-running cluster Vertex AI keeps provisioned so training jobs and Ray on Vertex AI start in seconds and scarce accelerators stay held between jobs.
- `GcpModelArmorTemplate` -- A Model Armor template: the named safety filters (prompt injection and jailbreak, Responsible AI content, sensitive data, malicious URLs) an AI application screens prompts and responses through.
- `GcpDocumentAiProcessor` -- A Document AI processor: a managed model that turns documents into structured data (OCR, forms, invoices, IDs), with its default version.
- `GcpColabRuntimeTemplate` -- A Colab Enterprise runtime template: the machine, network, image, and security settings every notebook runtime created from it gets.
- `GcpColabRuntime` -- A Colab Enterprise runtime: a notebook VM assigned to one user, built from a runtime template and started or stopped on purpose.
- `GcpColabSchedule` -- A Vertex AI schedule: a cron that launches a Colab Enterprise notebook run or a Vertex AI Pipelines run.
- `GcpTpuVm` -- A Cloud TPU VM: a slice of Google's AI accelerators with its host VMs. Beta-only in Google's provider (a recorded google-beta admission).
- `GcpDialogflowCxAgent` -- A Dialogflow CX conversational agent with the infrastructure its console-authored content uses folded in: webhooks, tools and their frozen versions, flow versions, the environments that pin them, and generative settings per language.
- `GcpVertexAiSearchDataConnector` -- A data connector is a COLLECTION of data stores Google syncs from a source (Jira, Confluence, ServiceNow, SharePoint, BigQuery, Google Drive, ...) on a schedule -- a different root from a data store, which is why it is its own kind. Engines search its stores by naming the collection.
- `GcpModelArmorFloorSetting` -- A Model Armor floor setting: the minimum safety screening a project, folder, or organization enforces on its templates and directly on Vertex AI and Google MCP server traffic. A different parent from a template (and a singleton Google never deletes), which is why it is its own kind.
- `GcpTpuQueuedResource` -- A Cloud TPU queued resource: a request that waits for TPU capacity and then provisions the nodes it describes. A different root from a TPU VM that owns many nodes, which is why it is its own kind. Beta-only in Google's provider (a recorded google-beta admission).
- `GcpDialogflowCxSecuritySettings` -- Dialogflow CX security settings: the redaction, retention, audio-export, and Insights-export policy agents in one project and location apply to their conversations. A different root from an agent, referenced by agents and shared among them, which is why it is its own kind.
- `GcpPrivateCaPool` -- 3230–3239: GCP security (Certificate Authority Service, Cloud KMS Autokey handles, Security Command Center, Binary Authorization) A Certificate Authority Service CA pool: the trust anchor and issuance policy its certificate authorities and certificates live inside.
- `GcpKmsKeyHandle` -- An Autokey key handle: asks Autokey for a customer-managed key for one resource type in one project and location; the resource it protects names the key the handle returns.
- `GcpSccNotificationConfig` -- Security Command Center streaming notifications to Pub/Sub at a project, folder, or organization.
- `GcpSccMuteConfig` -- A Security Command Center mute rule at a project, folder, or organization.
- `GcpSccBigQueryExport` -- A continuous Security Command Center findings export to a BigQuery dataset at a project, folder, or organization.
- `GcpBinaryAuthorizationPolicy` -- A project's Binary Authorization policy: which container images GKE admits, per cluster. A project singleton.
- `GcpBinaryAuthorizationAttestor` -- A Binary Authorization attestor with its Artifact Analysis note: the public keys that verify image attestations a policy requires.
- `GcpPrivateCaCertificateAuthority` -- A certificate authority in a CA pool: a self-signed root, or a subordinate signed by another authority or an outside CA. Its own kind because a pool rotates through several and a subordinate references its parent.
- `GcpPrivateCaCertificateTemplate` -- A certificate template: a reusable certificate shape in a project and location that certificates in any pool there reference.
- `GcpPrivateCaCertificate` -- A certificate issued from a CA pool for a key its owner holds; destroy revokes it.
- `GcpGkeFleet` -- 3240–3249: GCP platform engineering (GKE fleets, images, Cloud Build, Cloud Deploy) GcpGkeFleet is the container a fleet's scopes, namespaces, memberships, and features live in: the project's one fleet is the room, and a team scope or a fleet-wide feature is what is placed inside it.
- `GcpGkeFleetFeature` -- The three fleet children name GcpGkeFleet as their prerequisite: Google requires the fleet before a scope, a fleet declared after a cluster registers collides with the fleet that registration created implicitly, and a feature configures the fleet it lives in. A chart that references the fleet's project_id output orders each child after it.
- `GcpGkeFleetScope`
- `GcpGkeFleetMembership`
- `GcpComputeImage` -- A Compute Engine custom image: the golden boot image VMs, instance templates, and disks start from, rolled forward through image families.
- `GcpCloudBuildConnection` -- GcpCloudBuildConnection is the container a code host's repositories live in: the connection to GitHub, GitLab, or Bitbucket is the room, and each linked repository is what is placed inside it.
- `GcpCloudBuildRepository` -- A repository is created under its connection and cannot exist without it.
- `GcpCloudBuildTrigger` -- A Cloud Build trigger: what starts a build (a code event, a Pub/Sub message, a webhook, or a manual run) and what the build does.
- `GcpDeliveryPipeline` -- A Cloud Deploy delivery pipeline: the ordered stages a release is promoted through, with the automations that drive it.
- `GcpDeployTarget` -- A Cloud Deploy target: where a pipeline stage deploys (a GKE cluster, a Cloud Run location, a fleet cluster, several targets at once, or a custom target).
- `GcpFirebaseProject` -- 3250–3259: GCP Firebase (project enablement, app registrations, and the Firebase-adjacent products that follow) GcpFirebaseProject is the container the app registrations live in: "Firebase on this project" is the room, the Android/Apple/Web apps are what is placed inside it.
- `GcpFirebaseAndroidApp` -- The three app registrations exist only inside a Firebase-enabled project, so each names GcpFirebaseProject as its prerequisite: the E2E harness deploys the enablement first, and a chart that references the enablement's project_id output orders the registration after it.
- `GcpFirebaseAppleApp`
- `GcpFirebaseWebApp`
- `GcpCloudBuildWorkerPool` -- 3260–3269: GCP platform engineering, continued (Cloud Build and Cloud Deploy resources many pipelines share) A private Cloud Build worker pool: build machines many triggers and Cloud Deploy targets share, optionally on a private network.
- `GcpDeployPolicy` -- A Cloud Deploy deploy policy: rollout restrictions (freeze windows) that apply to every pipeline and target its selectors match.
- `GcpDeployCustomTargetType` -- A Cloud Deploy custom target type: how to render and deploy to a target Google does not deploy natively; many targets share one.
- `KubernetesNamespace` -- 4000–4999: Kubernetes resources, organized in family sub-bands (4030–4069 also hosts CNI/autoscaling/DR addons; 4130–4149 hosts analytics & ML; 4180–4199 hosts agent sandbox runtimes; 4200–4209 hosts AI agent platforms) 4000–4029: Kubernetes building blocks (core API primitives)
- `KubernetesDeployment`
- `KubernetesStatefulSet`
- `KubernetesDaemonSet`
- `KubernetesJob`
- `KubernetesCronJob`
- `KubernetesService`
- `KubernetesSecret`
- `KubernetesManifest`
- `KubernetesHelmRelease`
- `KubernetesConfigMap`
- `KubernetesServiceAccount`
- `KubernetesRbac` -- Bundles the RBAC grant grain (Role/ClusterRole + its binding) into one kind: "grant these permissions to these subjects in this scope".
- `KubernetesIngress`
- `KubernetesNetworkPolicy`
- `KubernetesPersistentVolumeClaim`
- `KubernetesStorageClass`
- `KubernetesResourceQuota` -- Manages the namespace-governance pair: the ResourceQuota plus an optional companion LimitRange (per-object defaults/bounds) — two API objects, one governance story.
- `KubernetesPriorityClass`
- `KubernetesPodDisruptionBudget`
- `KubernetesHorizontalPodAutoscaler`
- `KubernetesCertManager` -- 4030–4069: Kubernetes foundation addons (certs, DNS, secrets, ingress, Gateway API, mesh, CNI/autoscaling/DR)
- `KubernetesClusterIssuer` -- KubernetesCertManager is a prerequisite for the three cert-manager CR kinds below: ClusterIssuer/Issuer/Certificate are cert-manager custom resources — without the controller and its CRDs they cannot be applied.
- `KubernetesIssuer`
- `KubernetesCertificate`
- `KubernetesExternalDns`
- `KubernetesExternalSecretsOperator`
- `KubernetesClusterSecretStore` -- KubernetesExternalSecretsOperator is a prerequisite for the three external-secrets CR kinds below: ClusterSecretStore/SecretStore/ ExternalSecret are external-secrets custom resources — without the operator and its CRDs they cannot be applied.
- `KubernetesSecretStore`
- `KubernetesExternalSecret`
- `KubernetesIngressNginx`
- `KubernetesGatewayApiCrds`
- `KubernetesGatewayClass`
- `KubernetesGateway`
- `KubernetesListenerSet`
- `KubernetesHttpRoute`
- `KubernetesGrpcRoute`
- `KubernetesTcpRoute`
- `KubernetesUdpRoute`
- `KubernetesTlsRoute`
- `KubernetesReferenceGrant`
- `KubernetesBackendTlsPolicy`
- `KubernetesIstioBaseCrds`
- `KubernetesIstio`
- `KubernetesDestinationRule` -- Istio API kinds (mesh traffic policy, security, telemetry). The seven typed resources below (4053–4059) require the Istio CRDs on the cluster, provided by the lightweight CRDs-only KubernetesIstioBaseCrds (851) — NOT the full mesh KubernetesIstio (852).
- `KubernetesServiceEntry`
- `KubernetesPeerAuthentication`
- `KubernetesRequestAuthentication`
- `KubernetesAuthorizationPolicy`
- `KubernetesTelemetry`
- `KubernetesEnvoyFilter`
- `KubernetesMetricsServer`
- `KubernetesCilium`
- `KubernetesKeda`
- `KubernetesKarpenter`
- `KubernetesKarpenterNodePool`
- `KubernetesKarpenterEc2NodeClass`
- `KubernetesClusterAutoscaler`
- `KubernetesVelero`
- `KubernetesGkeComputeClass`
- `KubernetesKubePrometheusStack` -- 4070–4079: Kubernetes observability
- `KubernetesGrafana`
- `KubernetesSignoz` -- KubernetesClickHouse is a prerequisite because SigNoz stores every trace, metric and log in ClickHouse and deploys none of its own — the telemetry store is composed, never bundled.
- `KubernetesLoki`
- `KubernetesTempo`
- `KubernetesOtelOperator` -- The operator's admission webhooks (failurePolicy Fail) are served with a cert-manager Certificate in the default posture — cert-manager must be running before the operator installs.
- `KubernetesOtelCollector`
- `KubernetesServiceMonitor` -- KubernetesKubePrometheusStack is a prerequisite because it installs the monitoring.coreos.com CRDs the ServiceMonitor is an instance of, and the Prometheus that scrapes through it.
- `KubernetesPodMonitor` -- KubernetesKubePrometheusStack is a prerequisite because it installs the monitoring.coreos.com CRDs the PodMonitor is an instance of, and the Prometheus that scrapes through it.
- `KubernetesPrometheusRule` -- KubernetesKubePrometheusStack is a prerequisite because it installs the monitoring.coreos.com CRDs the PrometheusRule is an instance of, and the Prometheus that evaluates it.
- `KubernetesKyverno` -- 4080–4099: Kubernetes security, policy, and identity
- `KubernetesGatekeeper`
- `KubernetesKeycloak` -- Keycloak declarations compose the official Keycloak Operator (which reconciles the Keycloak CR this kind renders) and, on the recommended postgres vendor, a KubernetesPostgres database — both must resolve before the CR can converge.
- `KubernetesOpenBao`
- `KubernetesOpenFga` -- OpenFGA requires a datastore; the recommended arm composes a KubernetesPostgres database (the sandbox memory arm needs nothing, but the registry declares the shape real deployments require).
- `KubernetesKeycloakOperator`
- `KubernetesCloudNativePgOperator` -- 4100–4129: Kubernetes data platforms
- `KubernetesPostgres`
- `KubernetesValkey`
- `KubernetesPerconaMysqlOperator`
- `KubernetesMysql`
- `KubernetesPerconaMongoOperator`
- `KubernetesMongodb`
- `KubernetesStrimziKafkaOperator`
- `KubernetesKafka` -- container_kind: a Strimzi Kafka cluster is a place in the provider's own model — KafkaTopic and KafkaUser declarations BELONG to one cluster (the strimzi.io/cluster label) and are drawn inside its box. Clients that merely talk to the cluster (Connect, MirrorMaker2, UI, Karapace) carry containment_exempt on their bootstrap/trust references.
- `KubernetesKafkaTopic`
- `KubernetesKafkaUser`
- `KubernetesKafkaConnect` -- container_kind: a Connect cluster hosts the connectors deployed INTO it (KafkaConnector's strimzi.io/cluster label names its Connect cluster) — the same room shape as KubernetesKafka above.
- `KubernetesKafkaConnector`
- `KubernetesKafkaMirrorMaker2`
- `KubernetesKarapace`
- `KubernetesKafkaUi`
- `KubernetesOpenSearchOperator`
- `KubernetesOpenSearch`
- `KubernetesAltinityOperator`
- `KubernetesClickHouse`
- `KubernetesSolrOperator`
- `KubernetesSolr`
- `KubernetesNeo4j`
- `KubernetesSeaweedFs`
- `KubernetesQdrant`
- `KubernetesRabbitMqOperator` -- The RabbitMQ Cluster Operator's release manifest ships admission webhooks whose serving certificate is a cert-manager Certificate — cert-manager must be running before the operator installs.
- `KubernetesRabbitMq`
- `KubernetesCnpgBarmanCloudPlugin` -- The Barman Cloud plugin is CloudNativePG's object-store backup path, installed as its own Helm release beside the operator. It registers with the operator over CNPG-I and issues its TLS through cert-manager, so both are prerequisites: nothing to register with, or no certificates, and the release never becomes ready.
- `KubernetesAirflow` -- 4130–4149: Kubernetes analytics and ML KubernetesPostgres is a prerequisite because Airflow's metadata database composes a KubernetesPostgres by default (the spec's FK defaults resolve onto its outputs) and the migration Job needs the database reachable before the server components start.
- `KubernetesSparkOperator`
- `KubernetesKubeRayOperator`
- `KubernetesRayCluster` -- KubernetesKubeRayOperator is a prerequisite because this kind declares the RayCluster custom resource that only the operator's CRDs admit and only the operator reconciles into head and worker pods.
- `KubernetesFlinkOperator` -- KubernetesCertManager is a prerequisite because the Flink operator's chart, with its default-on admission webhook, renders cert-manager Issuer/Certificate resources and trusts the API server through cert-manager's CA injection — there is no self-signed fallback at the pinned chart, and the webhooks are fail-closed.
- `KubernetesFlinkDeployment` -- KubernetesFlinkOperator is a prerequisite because this kind declares the FlinkDeployment custom resource that only the operator's CRDs admit and only the operator reconciles into a running Flink cluster.
- `KubernetesJupyterHub` -- KubernetesPostgres is a prerequisite because JupyterHub's hub database composes a KubernetesPostgres in its external-database arm (the spec's FK defaults resolve onto its outputs) and the hub pod mounts that database's credential Secret before it can start.
- `KubernetesMlflow` -- KubernetesPostgres is a prerequisite because MLflow's backend store composes a KubernetesPostgres in its production arm (FK defaults onto its outputs; the module composes the connection URI from its credential Secret), and KubernetesSeaweedFs because the artifact store's S3-compatible arm FK-defaults onto the SeaweedFS endpoint and credential Secret.
- `KubernetesTrino` -- KubernetesPostgres is a prerequisite because Trino's postgres catalogs compose a KubernetesPostgres (the catalog host and credential FK-default onto its outputs), and the pods read that database's credential Secret to resolve catalog passwords from environment.
- `KubernetesSuperset` -- KubernetesPostgres is a prerequisite because Superset's REQUIRED metadata database composes a KubernetesPostgres (FK defaults onto its outputs; the module composes the environment Secret from its credential Secret), and KubernetesValkey because the cache/broker arm FK-defaults onto a KubernetesValkey's service and password Secret.
- `KubernetesLangfuse` -- Langfuse's four required stores each FK-default onto a catalog kind: KubernetesPostgres (users, projects, prompts, API keys - host and password Secret), KubernetesClickHouse (traces and scores - host, cluster name and auth Secret), KubernetesValkey (the ingestion queues - host, ACL user and password Secret), and KubernetesSeaweedFs (the S3-compatible event store - endpoint and credentials Secret).
- `KubernetesArgocd` -- 4150–4169: Kubernetes GitOps and CI/CD
- `KubernetesArgoWorkflows`
- `KubernetesTektonOperator`
- `KubernetesTekton` -- KubernetesTektonOperator is a prerequisite because this kind declares the TektonConfig custom resource that only the operator's CRDs admit and only the operator reconciles into running components.
- `KubernetesGhaRunnerScaleSetController`
- `KubernetesGhaRunnerScaleSet` -- KubernetesGhaRunnerScaleSetController is a prerequisite because this kind renders an AutoscalingRunnerSet custom resource that only the controller's CRDs admit and only the controller reconciles into listener and runner pods.
- `KubernetesHarbor`
- `KubernetesJenkins`
- `KubernetesTemporal` -- 4170–4179: Kubernetes app platforms KubernetesPostgres is a prerequisite because the recommended (and E2E-proven) database composition backs Temporal's default and visibility stores with a CloudNativePG cluster.
- `KubernetesNats`
- `KubernetesLocust`
- `KubernetesPlantonRunner`
- `KubernetesPlantonOperator`
- `KubernetesPlantonPlatform` -- KubernetesPlantonOperator is a prerequisite because this kind declares the PlantonPlatform custom resource that only the operator's CRD admits and only the operator reconciles into a running platform.
- `KubernetesGoFeatureFlag`
- `KubernetesFlagd`
- `KubernetesGoFeatureFlagFlagFile`
- `KubernetesFlagdFlagFile`
- `KubernetesAgentSandboxController` -- 4180–4199: agent sandbox runtimes (4180–4189 Kubernetes SIG Apps' agent-sandbox; 4190–4199 Agent Substrate)
- `KubernetesAgentSandbox` -- KubernetesAgentSandboxController is a prerequisite because it installs the agents.x-k8s.io definition the Sandbox is an instance of, and the controller that runs its pod.
- `KubernetesAgentSandboxTemplate` -- KubernetesAgentSandboxController is a prerequisite because it installs the extensions.agents.x-k8s.io definition the SandboxTemplate is an instance of, and the extension controllers that turn it into network policy and warm pool members.
- `KubernetesAgentSandboxWarmPool` -- KubernetesAgentSandboxController is a prerequisite because it installs the extensions.agents.x-k8s.io definition the SandboxWarmPool is an instance of, and the controller that keeps the pool's members ready.
- `KubernetesAgentSandboxClaim` -- KubernetesAgentSandboxController is a prerequisite because it installs the extensions.agents.x-k8s.io definition the SandboxClaim is an instance of, and the controller that binds it to a warm pool member.
- `KubernetesAgentSubstrate`
- `KubernetesAgentSubstrateWorkerPool` -- KubernetesAgentSubstrate is a prerequisite because it installs the ate.dev definition the WorkerPool is an instance of, the controller that turns it into worker pods, and the node agent and API server the workers register with.
- `KubernetesAgentSubstrateSandboxConfig` -- KubernetesAgentSubstrate is a prerequisite because it installs the ate.dev definition the SandboxConfig is an instance of, the admission policy that holds each class to its assets, and the API server and node agent that read it.
- `KubernetesAgentSubstrateCsiDriverConfig` -- KubernetesAgentSubstrate is a prerequisite because it installs the ate.dev definition the CSIDriverConfig is an instance of, and the API server and node agent that look a driver up in it.
- `KubernetesStigmer` -- 4200–4209: Kubernetes AI agent platforms KubernetesStigmer installs the Stigmer operator and declares the one Stigmer it reconciles, so it carries no prerequisite.
- `DigitalOceanApp` -- 5000–5999: DigitalOcean resources
- `DigitalOceanBucket`
- `DigitalOceanContainerRegistry`
- `DigitalOceanDatabaseCluster` -- A DigitalOcean managed database cluster -- the cluster every other managed-database resource is created on: users, logical databases, connection pools, the firewall rule set, Kafka topics, and Kafka schema subjects are all API-addressed under a cluster's id and cannot exist before it. A container kind for exactly that reason: those six kinds are children of the cluster, so on a diagram the cluster is the room they stand in. A read replica also names its primary but is a cluster of its own, a monitor alert names the clusters it watches, and an App Platform app names a cluster it attaches; those three references are containment-exempt on their specs.
- `DigitalOceanDnsZone`
- `DigitalOceanDroplet` -- No VPC prerequisite: the droplet spec's vpc reference is optional — an omitted vpc places the droplet in the region's default VPC.
- `DigitalOceanFirewall`
- `DigitalOceanFunction`
- `DigitalOceanKubernetesCluster` -- DigitalOceanVpc is a prerequisite because the cluster spec's vpc reference is required: the control plane and every node pool live inside one VPC, resolved to the DigitalOceanVpc's exported vpc_id output.
- `DigitalOceanKubernetesNodePool` -- DigitalOceanKubernetesCluster is a prerequisite because a node pool is API-addressed under its owning cluster: the spec's cluster reference is required and the pool cannot exist first.
- `DigitalOceanLoadBalancer` -- No registry prerequisite: the load balancer's vpc reference is optional (DigitalOcean places it in the region's default VPC when unset, and GLOBAL balancers take no VPC at all). Scenarios that exercise VPC placement declare it per-scenario via the e2e-prerequisites annotation.
- `DigitalOceanVolume`
- `DigitalOceanVpc`
- `DigitalOceanCertificate`
- `DigitalOceanDnsRecord` -- DigitalOceanDnsZone is a prerequisite because a record is API-addressed under its domain: the spec's domain reference is required, resolved to the DigitalOceanDnsZone's exported zone_name output.
- `DigitalOceanDatabaseUser` -- An additional user on a managed database cluster: the cluster reference is required, resolved to the DigitalOceanDatabaseCluster's exported cluster_id output.
- `DigitalOceanDatabaseDb` -- An additional logical database on a managed database cluster: the cluster reference is required.
- `DigitalOceanDatabaseConnectionPool` -- A PgBouncer connection pool on a managed PostgreSQL cluster: the cluster reference is required.
- `DigitalOceanDatabaseFirewall` -- The inbound trusted-sources rule set of a managed database cluster: the cluster reference is required.
- `DigitalOceanDatabaseReplica` -- A read-only replica of a managed database cluster: the primary cluster reference is required.
- `DigitalOceanDatabaseKafkaTopic` -- A topic on a managed Kafka cluster with the full per-topic configuration block; the owning cluster reference is required.
- `DigitalOceanDatabaseKafkaSchema` -- One schema subject registered in a managed Kafka cluster's schema registry; the owning cluster reference is required.
- `DigitalOceanProject` -- The account-level organizational container; membership is carried on the project itself as resource URNs.
- `DigitalOceanSshKey` -- An SSH public key registered on the account, referenced by droplets and droplet autoscale pools at create time.
- `DigitalOceanMonitorAlert` -- An alert policy on DigitalOcean's built-in metrics for droplets, load balancers, and managed database clusters. All entity targeting is optional, so there is no registry prerequisite.
- `DigitalOceanUptimeCheck` -- An availability/latency probe on an external endpoint with composed alert rules; the target is outside the account, so there is no registry prerequisite.
- `DigitalOceanReservedIp` -- A static public IP address (IPv4 or IPv6) reserved in a region and optionally assigned to a droplet. The droplet attachment is an optional composition seam, so there is no registry prerequisite.
- `DigitalOceanVpcPeering` -- A private-network peering connection between exactly two VPCs; both VPC references are required.
- `DigitalOceanSpacesKey` -- An access-key pair for Spaces object storage. Bucket grants are an optional composition seam, so there is no registry prerequisite.
- `DigitalOceanCdn` -- A CDN endpoint serving a Spaces bucket's content from the global edge: the origin reference is required, resolved to the DigitalOceanBucket's exported bucket_domain_name output.
- `DigitalOceanDropletAutoscalePool` -- A pool of identical droplets DigitalOcean keeps at a fixed size or scales on utilization. The template's ssh_keys reference is required (the API mandates SSH keys), resolved to the DigitalOceanSshKey's exported ssh_key_id output.
- `CloudflareDnsZone` -- 7000–7999: Cloudflare resources
- `CloudflareKvNamespace`
- `CloudflareR2Bucket`
- `CloudflareWorker`
- `CloudflareLoadBalancer` -- CloudflareDnsZone and CloudflareLoadBalancerPool are prerequisites because a load balancer is a DNS-level construct inside a zone (the spec's zone_id reference must resolve) and traffic must land somewhere (the required fallback_pool reference must resolve to a live pool).
- `CloudflareD1Database`
- `CloudflareZeroTrustAccessApplication`
- `CloudflareDnsRecord` -- CloudflareDnsZone is a prerequisite because every record lives inside a zone -- the spec's zone_id reference must resolve before the record can be created.
- `CloudflareRuleset` -- CloudflareDnsZone is a prerequisite because zone-scoped rulesets (the common case; the spec's zone_id reference defaults to the zone kind) must resolve their zone first. Account-scoped rulesets simply leave the reference unused.
- `CloudflareWorkersKvPair` -- CloudflareKvNamespace is a prerequisite because a KV pair is written into a namespace -- the spec's namespace_id reference must resolve first.
- `CloudflareHyperdriveConfig`
- `CloudflareLoadBalancerPool`
- `CloudflareLoadBalancerMonitor`
- `CloudflareZeroTrustAccessPolicy`
- `CloudflareZeroTrustAccessGroup`
- `CloudflareQueue`
- `CloudflarePagesProject`
- `CloudflareZeroTrustTunnel`
- `CloudflareZeroTrustTunnelVirtualNetwork`
- `CloudflareZeroTrustTunnelRoute` -- CloudflareZeroTrustTunnel is a prerequisite because a route steers a CIDR through an existing tunnel -- the spec's tunnel_id reference must resolve first. (The optional virtual_network_id is scenario-declared, not a registry prerequisite.)
- `CloudflareList`
- `CloudflareListItem` -- CloudflareList is a prerequisite because an item exists only inside a list -- the spec's list_id reference must resolve first.
- `CloudflareTurnstileWidget`
- `CloudflareEmailRoutingZone` -- CloudflareDnsZone is a prerequisite because email routing is enabled ON a zone -- the spec's zone_id reference must resolve first.
- `CloudflareEmailRoutingRule` -- CloudflareDnsZone is a prerequisite because a routing rule lives in a zone's email routing configuration -- the spec's zone_id reference must resolve first. CloudflareEmailRoutingZone is deliberately NOT a prerequisite: the API accepts rule creation on a zone whose Email Routing was never enabled (measured live 2026-08-26 -- POST zones/{id}/email/routing/rules answered 200 on a fresh zone with routing status "unconfigured"). Enablement gates mail FLOW, not rule lifecycle, so it is an operational concern the kind docs teach, not a create-time dependency. (Forward destinations reference CloudflareEmailRoutingAddress only for forward-type rules, so that edge is scenario-declared.)
- `CloudflareEmailRoutingAddress`
- `CloudflareOriginCaCertificate`
- `CloudflareCertificatePack` -- CloudflareDnsZone is a prerequisite because a certificate pack is ordered for a zone's hostnames -- the spec's zone_id reference must resolve first.
- `CloudflareCustomHostname` -- CloudflareDnsZone is a prerequisite because a custom hostname (SSL for SaaS) is provisioned inside a zone -- the spec's zone_id reference must resolve first.
- `CloudflareCustomHostnameFallbackOrigin` -- CloudflareDnsZone is a prerequisite because the fallback origin is a zone-level SSL-for-SaaS setting -- the spec's zone_id reference must resolve first.
- `CloudflareZeroTrustAccessIdentityProvider` -- No prerequisites: identity providers are account-scoped in the canonical case (the optional zone scope is a per-manifest choice, not a structural dependency).
- `CloudflareZeroTrustAccessServiceToken` -- No prerequisites: service tokens are account-scoped in the canonical case (the optional zone scope is a per-manifest choice, not a structural dependency).
- `CloudflareZeroTrustOrganization` -- No prerequisites: the organization is an account-scoped configuration singleton (the optional zone scope is a per-manifest choice, not a structural dependency).
- `CloudflareZeroTrustAccessInfrastructureTarget` -- No prerequisites: targets are account-scoped, and the virtual-network reference is an optional per-manifest edge (omitted = the account's default virtual network).
- `CloudflareZeroTrustMcpPortal` -- No prerequisites: portals are account-scoped, and the servers[] rows' MCP-server references are optional per-manifest edges.
- `CloudflareZeroTrustMcpServer` -- No prerequisites: MCP server registrations are account-scoped and self-contained -- portals reference them, not the reverse.
- `CloudflareZeroTrustGatewayPolicy` -- No prerequisites: Gateway policies are account-scoped, and their list / virtual-network references are optional per-manifest edges.
- `CloudflareZeroTrustList` -- No prerequisites: Zero Trust lists are account-scoped and self-contained.
- `CloudflareZeroTrustGatewaySettings` -- No prerequisites: the Gateway configuration is an account-scoped singleton, and its certificate reference is an optional per-manifest edge.
- `CloudflareZeroTrustDnsLocation` -- No prerequisites: DNS locations are account-scoped and self-contained.
- `CloudflareZeroTrustDeviceDefaultProfile` -- No prerequisites: the default device profile is an account-scoped configuration singleton; its virtual-network and zone-certificate references are optional per-manifest edges.
- `CloudflareZeroTrustDeviceCustomProfile` -- No prerequisites: custom device profiles are account-scoped, and the virtual-network reference is an optional per-manifest edge.
- `CloudflareZeroTrustDevicePostureRule` -- No prerequisites: posture rules are account-scoped and self-contained (list and integration references are literal UUIDs today).
- `CloudflareZoneTlsSettings` -- CloudflareDnsZone is a prerequisite because TLS settings are zone-scoped configuration -- the spec's zone_id reference must resolve first.
- `CloudflareCustomSslCertificate` -- CloudflareDnsZone is a prerequisite because a custom certificate is uploaded to an existing zone -- the spec's zone_id reference must resolve first.
- `CloudflareMtlsCertificate` -- No prerequisites: mTLS certificates are account-scoped uploads and self-contained -- consumers (zone TLS CA associations, Authenticated Origin Pulls rows, Workers mTLS bindings) reference them, not the reverse.
- `CloudflareAuthenticatedOriginPulls` -- CloudflareDnsZone is a prerequisite because Authenticated Origin Pulls enablement configures an existing zone -- the spec's zone_id reference must resolve first. The per-hostname certificate edge is optional and scenario-declared, never a registry prerequisite.
- `CloudflareAuthenticatedOriginPullsCertificate` -- CloudflareDnsZone is a prerequisite because the client certificate is uploaded to an existing zone -- the spec's zone_id reference must resolve first.
- `CloudflareZoneSettings` -- CloudflareDnsZone is a prerequisite because zone settings configure an existing zone -- the spec's zone_id reference must resolve first.
- `CloudflareCacheSettings` -- CloudflareDnsZone is a prerequisite because cache settings configure an existing zone -- the spec's zone_id reference must resolve first.
- `CloudflareIpAccessRule` -- No prerequisites: IP Access rules are account-scoped in the canonical case (the zone scope is a per-manifest choice, not a structural dependency).
- `CloudflareBotManagement` -- CloudflareDnsZone is a prerequisite because Bot Management is zone-singleton configuration -- the spec's zone_id reference must resolve first.
- `CloudflareSnippet` -- CloudflareDnsZone is a prerequisite because snippets deploy to a zone -- the spec's zone_id reference must resolve first.
- `CloudflareSnippetRules` -- CloudflareDnsZone and CloudflareSnippet are prerequisites: the rules table is zone-scoped and every rule invokes a snippet by name.
- `CloudflareWaitingRoom` -- CloudflareDnsZone is a prerequisite because waiting rooms sit on a zone's host+path -- the spec's zone_id reference must resolve first.
- `CloudflareWaitingRoomEvent` -- CloudflareWaitingRoom is a prerequisite because events run on a room (and the room's own chain brings the zone).
- `CloudflareLogpushJob` -- No prerequisites: logpush jobs are dual-scope (account or zone) and the zone reference is optional -- zone-scoped lanes declare the edge at the scenario level.
- `CloudflareNotificationPolicy` -- No prerequisites: every delivery mechanism (email, PagerDuty, webhook) is optional -- policies referencing a webhook declare the edge at the scenario level.
- `CloudflareNotificationWebhook` -- No prerequisites: a webhook destination is account-scoped and self-contained -- notification policies reference it, not the reverse.
- `CloudflareWebAnalyticsSite` -- No prerequisites: a site is identified by host OR zone, and the zone reference is optional -- zone-measured lanes declare the edge at the scenario level.
- `CloudflareWorkflow` -- CloudflareWorker is a prerequisite because a workflow registers a class exported by a DEPLOYED Worker script -- the spec's script_name reference must resolve first.
- `CloudflareSecretsStore` -- No prerequisites: the Secrets Store is an account-scoped container and self-contained -- consumers (store secrets, Worker bindings, AI Gateway authentication) reference it, not the reverse.
- `CloudflareSecretsStoreSecret` -- CloudflareSecretsStore is a prerequisite because every secret lives inside a store -- the spec's store_id reference must resolve first.
- `CloudflareAiGateway` -- No prerequisites: the gateway is account-scoped and self-contained; its optional Secrets Store link (BYO provider keys) is a scenario-level composition, not a structural requirement.
- `CloudflareAccountApiToken` -- No prerequisites: an account-owned API token is self-contained; the resources its policies cover are identifiers, not references.
- `CloudflareHealthcheck` -- CloudflareDnsZone is a prerequisite because standalone health checks are zone-scoped -- the spec's zone_id reference must resolve first.
- `Auth0Connection` -- 8000–8999: Auth0 resources
- `Auth0Client`
- `Auth0EventStream`
- `Auth0ResourceServer`
- `Auth0Action`
- `Auth0Role`
- `Auth0User`
- `Auth0TenantSettings`
- `Auth0CustomDomain`
- `Auth0CustomDomainVerification`
- `Auth0Branding` -- 8010–8015: how the tenant's Universal Login looks, reads and mails -- branding and theme, the login flow, each prompt's words and screen partials, the email service and each email.
- `Auth0Prompt`
- `Auth0PromptCustomText`
- `Auth0PromptScreenPartials`
- `Auth0EmailProvider`
- `Auth0EmailTemplate`
- `Auth0ClientFromMetadataDocument`
- `OpenFgaStore` -- 9000–9999: OpenFGA resources OpenFGA publishes a Terraform provider and no Pulumi provider, so its kinds ship one HCL module and declare the engines that run it; the CLI and the platform refuse Pulumi for them before anything runs.
- `OpenFgaAuthorizationModel`
- `OpenFgaRelationshipTuple`
- `StripeWebhookEndpoint` -- 10000–10999: Stripe resources Stripe publishes a Terraform provider and no Pulumi provider, and its kinds are proven on OpenTofu alone, so each declares ["tofu"]: one HCL module, and every other engine refused before anything runs. Family sub-bands: 10000 account and delivery, 10100 catalog and pricing.
- `StripeEventDestination`
- `StripeBillingPortalConfiguration`
- `StripePaymentMethodConfiguration`
- `StripePaymentMethodDomain`
- `StripeRadarValueList`
- `StripeProduct`
- `StripePrice` -- A price always belongs to a product (its product field is required), so the product deploys first.
- `StripeCoupon`
- `StripePromotionCode` -- A promotion code always redeems a coupon (its coupon field is required), so the coupon deploys first.
- `StripeEntitlementFeature`
- `StripeShippingRate`
- `StripeTaxRate` -- Tax rate and tax registration share the initials "tr", so each takes Stripe's own id prefix with its vowels dropped (txr_, taxreg_).
- `StripeTaxRegistration`
- `StripeBillingMeter`
- `StripePaymentLink` -- A payment link always sells at least one price (line_items is required), so the price deploys first.

### spec.runner.extraEnv.variables[].valueFrom.env

`string`

### spec.runner.extraEnv.variables[].valueFrom.name

`string` · required

- rule: {"required":true}

### spec.runner.extraEnv.variables[].valueFrom.fieldPath

`string`

### spec.runner.extraEnv.variables[].configMapKeyRef

`ConfigMapKeyRef`

Reference to a key in a Kubernetes ConfigMap.

### spec.runner.extraEnv.variables[].configMapKeyRef.name

`string` · required

Name of the ConfigMap.

- rule: {"required":true}

### spec.runner.extraEnv.variables[].configMapKeyRef.key

`string` · required

Key within the ConfigMap.

- rule: {"required":true}

### spec.runner.extraEnv.variables[].configMapKeyRef.optional

`bool`

If true, the env var is silently skipped when the ConfigMap or key does not exist
(instead of blocking pod startup).

### spec.runner.extraEnv.variables[].fieldRef

`ObjectFieldRef`

Reference to a pod-level field (metadata.name, status.podIP, etc.).

### spec.runner.extraEnv.variables[].fieldRef.apiVersion

`string`

Version of the schema. Defaults to "v1".

### spec.runner.extraEnv.variables[].fieldRef.fieldPath

`string` · required

Path of the field to select (e.g., "metadata.name", "status.podIP").

- rule: {"required":true}

### spec.runner.extraEnv.variables[].resourceFieldRef

`ResourceFieldRef`

Reference to container resource limits or requests (limits.cpu, requests.memory, etc.).

### spec.runner.extraEnv.variables[].resourceFieldRef.containerName

`string`

Container name. Required for init containers; defaults to the current
container for regular containers.

### spec.runner.extraEnv.variables[].resourceFieldRef.resource

`string` · required

Resource to select (e.g., "limits.cpu", "requests.memory").

- rule: {"required":true}

### spec.runner.extraEnv.variables[].resourceFieldRef.divisor

`string`

Specifies the output format of the exposed resource.
For CPU: "1" means cores. For memory: "1", "1Ki", "1Mi", "1Gi".

### spec.runner.extraEnv.secrets

`[]SecretEnvVar`

Individual secret environment variables (sensitive).

### spec.runner.extraEnv.secrets[].name

`string` · required

The environment variable name.

- rule: Must be a valid C_IDENTIFIER (start with letter/underscore, contain only letters, digits, underscores)
- rule: {"required":true}

### spec.runner.extraEnv.secrets[].value

`string` · sensitive

The secret's value: on Planton a `$secret/<slug>` reference, resolved at deploy; on a
deploy without the platform, the literal. The module writes it into a Kubernetes Secret
and the environment variable references that Secret.

### spec.runner.extraEnv.secrets[].secretRef

`KubernetesSecretKeyRef`

Reference to a key within an existing Kubernetes Secret.

### spec.runner.extraEnv.secrets[].secretRef.namespace

`string`

The namespace of the Kubernetes Secret.
If not specified, defaults to the namespace where the component is deployed.
Note: Cross-namespace secret references may not be supported by all Helm charts.

### spec.runner.extraEnv.secrets[].secretRef.name

`string` · required

The name of the Kubernetes Secret.

- rule: {"required":true}

### spec.runner.extraEnv.secrets[].secretRef.key

`string` · required

The key within the Kubernetes Secret that contains the value.

- rule: {"required":true}

### spec.runner.extraEnv.secrets[].secretRef.optional

`bool`

If true, the env var is silently skipped when the Secret or key does not exist
(instead of blocking pod startup).

### spec.runner.extraEnv.secrets[].valueFrom

`ValueFromRef`

Reference to another Planton resource's secret output field.
The orchestrator resolves this before invoking IaC modules.

### spec.runner.extraEnv.secrets[].valueFrom.kind

`enum`

Allowed values (use exactly as shown):

- `unspecified` -- 0: Default/unspecified
- `TestCatalogKindGeneric` -- 1–49: Test/dev/custom
- `TestCatalogKindKubernetes`
- `AwsAlb` -- 1000–1999: AWS resources AwsSubnet is a prerequisite because an ALB requires at least two subnets in different availability zones -- the spec's subnet references must resolve before the load balancer can be created.
- `AwsCertManagerCert`
- `AwsCloudFront`
- `AwsDynamodb`
- `AwsEcrRepo`
- `AwsEcsCluster`
- `AwsEcsService` -- AwsEcsCluster, AwsEcsTaskDefinition, and AwsSubnet are prerequisites because a service schedules a referenced task-definition revision into a referenced live cluster and places task network interfaces into referenced subnets -- all three references must resolve first.
- `AwsEksCluster` -- AwsSubnet and AwsIamRole are prerequisites because the control plane attaches its network interfaces into referenced subnets and assumes a referenced cluster role that must already carry AmazonEKSClusterPolicy.
- `AwsIamRole`
- `AwsLambda`
- `AwsRdsCluster`
- `AwsRdsInstance`
- `AwsRoute53Zone`
- `AwsS3Bucket`
- `AwsLbTargetGroup` -- AwsVpc is a prerequisite because a target group's health checks and target registrations live inside one VPC -- the spec's vpc_id reference must resolve before the group can be created.
- `AwsSecurityGroup` -- AwsVpc is a prerequisite because every security group is created in a VPC; the E2E install profile resolves vpc_id against the VPC prerequisite.
- `AwsVpc`
- `AwsEksNodeGroup` -- AwsEksCluster is a prerequisite because nodes register with a live control plane; AwsIamRole and AwsSubnet back the node role and worker subnet references.
- `AwsIamUser`
- `AwsKmsKey`
- `AwsEc2Instance`
- `AwsClientVpn` -- Every Client VPN endpoint requires an ACM server certificate at create time; the imported self-signed fixture satisfies it. Subnets/VPC are optional composition (a zero-association endpoint is valid) -- composed scenarios declare them via the e2e-prerequisites annotation.
- `AwsDocumentDb`
- `AwsRoute53DnsRecord` -- AwsRoute53Zone is a prerequisite because every record lives inside a hosted zone -- the spec's zone_id reference must resolve before the record can be created.
- `AwsS3ObjectSet` -- AwsS3Bucket is a prerequisite because the object set's bucket reference is required -- objects cannot exist without the bucket that holds them.
- `AwsSqsQueue`
- `AwsSnsTopic`
- `AwsEventBridgeBus`
- `AwsEventBridgeRule`
- `AwsIamOidcProvider`
- `AwsIamPolicy`
- `AwsIamInstanceProfile` -- AwsIamRole is a prerequisite because an instance profile is a wrapper that must contain a role to be useful -- the profile's spec requires a role reference, so the role must be deployed first.
- `AwsLbListener` -- AwsAlb and AwsLbTargetGroup are prerequisites because a listener is an attachment point on a load balancer and its default action almost always forwards to a target group -- both references must resolve before the listener can be created.
- `AwsLbListenerRule` -- AwsLbListener is a prerequisite because a rule only exists as an attachment on a listener -- the listener_arn reference must resolve before the rule can be created.
- `AwsLaunchTemplate`
- `AwsAutoScalingGroup` -- AwsSubnet and AwsLaunchTemplate are prerequisites because a group cannot exist without subnets to place capacity in and a launch template to launch from -- the spec's subnets and launch_template references must resolve before the group can be created.
- `AwsEksAddon` -- AwsEksCluster is a prerequisite because an add-on installs onto a live control plane -- the spec's cluster_name reference must resolve before the add-on can be created.
- `AwsEksFargateProfile` -- AwsEksCluster, AwsIamRole, and AwsSubnet are prerequisites because a Fargate profile attaches to a live control plane, runs pods as a referenced pod-execution role, and launches them into referenced private subnets -- all three references must resolve first.
- `AwsEksAccessEntry` -- AwsEksCluster and AwsIamRole are prerequisites because an access entry grants a referenced IAM principal access to a live control plane -- both references must resolve before the entry can be created.
- `AwsEcsTaskDefinition` -- AwsIamRole is a prerequisite because the kind's default posture -- Fargate with the awslogs logging default -- is rejected by AWS at registration time without an execution role the agent can assume.
- `AwsHttpApiGateway`
- `AwsStepFunction` -- AwsIamRole is a prerequisite because a state machine cannot be created without an execution role it can assume -- the spec's role_arn reference must resolve before the CreateStateMachine call.
- `AwsHttpApiVpcLink` -- AwsSubnet is a prerequisite because a VPC link is a set of managed ENIs provisioned into referenced subnets -- the subnet references must resolve before the link can be created. Security groups are optional on the link, so they compose per-scenario rather than as a registry prerequisite.
- `AwsHttpApiDomain` -- AwsCertManagerCert is a prerequisite because a custom domain cannot be created without a TLS certificate in the same region covering the domain -- the spec's certificate_arn reference must resolve first.
- `AwsVpcEndpoint` -- AwsVpcEndpoint's composed E2E scenarios reference the AwsVpc prerequisite's outputs (vpc_id + default_route_table_id for gateway endpoints) and the AwsSubnet pair's subnet_id outputs (interface endpoints), so both are genuine deploy-order prerequisites.
- `AwsElasticacheUser`
- `AwsElasticacheUserGroup` -- AwsElasticacheUser is a genuine prerequisite: AWS refuses to create a user group that does not contain a user named "default", so a group's composed E2E scenario must resolve a deployed user's outputs.
- `AwsRedshiftServerlessNamespace`
- `AwsRedshiftServerlessWorkgroup` -- The namespace is a genuine prerequisite: a workgroup attaches to exactly one namespace by name at create time, so its composed E2E scenario must resolve a deployed namespace's outputs. AwsSubnet is a prerequisite because Redshift Serverless requires the workgroup's subnets to span three availability zones.
- `AwsRedisElasticache` -- AwsSubnet is a prerequisite because the module builds an ElastiCache subnet group from referenced subnets -- the spec's subnet references must resolve before the replication group can deploy.
- `AwsOpenSearchDomain`
- `AwsMemcachedElasticache`
- `AwsServerlessElasticache`
- `AwsNlb` -- AwsSubnet is a prerequisite because an NLB requires at least one subnet mapping -- the spec's subnet references must resolve before the load balancer can be created.
- `AwsElasticIp`
- `AwsTransitGateway`
- `AwsGlobalAccelerator`
- `AwsSubnet`
- `AwsInternetGateway`
- `AwsNatGateway` -- AwsInternetGateway is a prerequisite because a public NAT gateway can only become available once the VPC it sits in has an internet gateway attached (AWS rejects the create otherwise) -- so the gateway must be deployed first. AwsVpc is a prerequisite because a REGIONAL NAT gateway (availability_mode = regional) references the VPC directly instead of a subnet.
- `AwsEgressOnlyInternetGateway`
- `AwsElasticFileSystem` -- AwsSubnet and AwsSecurityGroup are prerequisites because mount targets (required, min 1) place the file system's NFS endpoints into subnets and attach security groups -- both references must resolve before the CreateMountTarget calls.
- `AwsEfsAccessPoint` -- AwsElasticFileSystem is a prerequisite because an access point is created INTO a file system -- the spec's required file_system_id reference must resolve before the CreateAccessPoint call.
- `AwsFsxLustreFileSystem`
- `AwsFsxOpenzfsFileSystem`
- `AwsFsxWindowsFileSystem` -- Every Windows file system must join an Active Directory domain; the directory itself is external infrastructure (AWS Managed Microsoft AD or a self-managed domain), so only the network dependency is a declarable prerequisite.
- `AwsFsxOntapFileSystem`
- `AwsFsxOntapStorageVirtualMachine`
- `AwsFsxOntapVolume`
- `AwsFsxDataRepositoryAssociation`
- `AwsCognitoUserPool`
- `AwsCognitoIdentityProvider` -- AwsCognitoUserPool is a prerequisite because an identity provider is created INTO a pool -- the spec's required user_pool_id reference must resolve before the CreateIdentityProvider call.
- `AwsCognitoUserPoolClient` -- AwsCognitoUserPool is a prerequisite because an app client is created INTO a pool -- the spec's required user_pool_id reference must resolve before the CreateUserPoolClient call.
- `AwsCognitoResourceServer` -- AwsCognitoUserPool is a prerequisite because a resource server is created INTO a pool -- the spec's required user_pool_id reference must resolve before the CreateResourceServer call.
- `AwsWafWebAcl`
- `AwsWafIpSet`
- `AwsWafRegexPatternSet`
- `AwsCloudwatchLogGroup`
- `AwsCloudwatchAlarm`
- `AwsCloudwatchCompositeAlarm`
- `AwsKinesisStream`
- `AwsKinesisFirehose` -- Every Firehose destination requires an S3 configuration (the primary target for extended_s3; the failed/all-document backup for the rest) and an IAM role Firehose assumes to write to it, so both are hard deploy prerequisites.
- `AwsKinesisStreamConsumer` -- A consumer registers against exactly one stream and cannot exist without it.
- `AwsAthenaWorkgroup`
- `AwsGlueCatalogDatabase`
- `AwsRedshiftCluster`
- `AwsSagemakerDomain` -- AI/ML A domain cannot exist without VPC subnets and a SageMaker execution role (default_user_settings.execution_role_arn is required), so both are hard deploy prerequisites.
- `AwsAppRunnerService` -- A service can run entirely on companion defaults, so the App Runner family's kinds are dependency-free leaves except the VPC connector (which cannot exist without subnets and security groups). A service's companion references (auto scaling / VPC connector / observability / WAF) are optional composition -- scenarios declare them via the e2e-prerequisites annotation.
- `AwsAppRunnerAutoScalingConfiguration`
- `AwsAppRunnerVpcConnector`
- `AwsAppRunnerObservabilityConfiguration`
- `AwsTransitGatewayVpcAttachment` -- AwsTransitGateway is a prerequisite because an attachment cannot exist without the gateway it attaches to; AwsSubnet because the attachment provisions an ENI into at least one subnet (the VPC arrives transitively through the subnet's own prerequisites).
- `AwsTransitGatewayRouteTable` -- Only the gateway is a hard prerequisite: a route table can exist empty. Associations, propagations, and routes referencing attachments are optional composition -- scenarios declare them via the e2e-prerequisites annotation.
- `AwsBatchComputeEnvironment` -- A MANAGED compute environment always launches into VPC subnets, so the subnet is a hard deploy prerequisite (security groups are required only for the Fargate types -- scenario-declared, not a registry edge).
- `AwsBatchJobQueue` -- A job queue cannot exist without at least one VALID compute environment to map onto.
- `AwsBatchSchedulingPolicy`
- `AwsBatchJobDefinition`
- `AwsCodeBuildProject` -- CI/CD
- `AwsCodePipeline`
- `AwsMwaaEnvironment` -- Workflow / Orchestration AwsSubnet and AwsSecurityGroup are prerequisites because the environment's network interfaces are placed in referenced private subnets and AWS requires at least one attached security group at creation.
- `AwsNeptuneCluster` -- Graph Database
- `AwsMemorydbCluster` -- A cluster always launches into a subnet group; the subnets are the hard deploy prerequisite. The ACL it attaches is optional composition (the built-in "open-access" ACL needs no resource) -- scenarios declare the ACL/user chain via the e2e-prerequisites annotation.
- `AwsMemorydbUser`
- `AwsMemorydbAcl` -- An empty ACL is valid (MemoryDB has no mandatory "default" member), so the user is optional composition -- the composed scenario declares it via the e2e-prerequisites annotation, never a registry edge.
- `AwsMskCluster` -- Streaming AwsSubnet and AwsSecurityGroup are prerequisites because brokers are placed in referenced subnets and AWS requires at least one attached security group at creation.
- `AwsMskServerlessCluster` -- AwsSubnet is a prerequisite because the serverless cluster's network interfaces are placed in referenced subnets (security groups are optional -- AWS attaches the VPC default group when none are referenced).
- `AwsLambdaEventSourceMapping` -- AwsLambda is a prerequisite because a mapping cannot exist without the function it invokes (a required reference). Event sources (SQS, Kinesis, DynamoDB, MSK) are optional composition -- scenarios declare them via the e2e-prerequisites annotation rather than taxing every consumer's chain.
- `AwsSnsSubscription` -- AwsSnsTopic is a prerequisite because a subscription cannot exist without the topic it subscribes to (a required reference). Endpoints (SQS queues, Lambda functions, Firehose streams) are optional composition -- scenarios declare them via the e2e-prerequisites annotation rather than taxing every consumer's chain.
- `AwsPlantonRunner` -- AwsSubnet is a prerequisite because the runner appliance places its network interfaces into referenced subnets -- the placement reference must resolve before the appliance can deploy.
- `AwsRoute53HealthCheck`
- `AwsSesConfigurationSet` -- Both SES kinds are dependency-free leaves: an identity's configuration set is optional composition (scenarios declare it via the e2e-prerequisites annotation), and a configuration set's event destinations reference other kinds only optionally.
- `AwsSesEmailIdentity`
- `AwsSecretsManagerSecret` -- A dependency-free leaf: the KMS key, rotation Lambda, and external rotation role references are all optional composition -- scenarios declare them via the e2e-prerequisites annotation, never registry edges.
- `AwsOpenSearchServerlessCollection` -- A dependency-free leaf: the collection-scoped encryption/network/ data-access/retention policies are module-rendered, and the KMS key and data-access principal references are optional composition (e2e-prerequisites annotation).
- `AwsBedrockGuardrail` -- A dependency-free leaf: the KMS key reference is optional composition (e2e-prerequisites annotation); published versions are folded satellites of the guardrail itself.
- `AwsBedrockCustomModel` -- AwsIamRole is a prerequisite because Bedrock assumes the job role to read training data and write outputs; the S3 locations and KMS key are optional composition (e2e-prerequisites annotation).
- `AwsBedrockInferenceProfile` -- A dependency-free leaf: the model source is a foundation model or an AWS system-defined cross-region profile, never a customer resource.
- `AwsBedrockProvisionedThroughput` -- A dependency-free leaf in the registry: capacity is typically bought for an AwsBedrockCustomModel (the default reference), but foundation model ARNs are equally legal, so the edge is optional composition.
- `AwsBedrockModelAccess` -- A dependency-free leaf: the agreement covers an AWS-listed foundation model, never a customer resource.
- `AwsBedrockInvocationLogging` -- Region settings singleton (one invocation-logging configuration per account+region; identity = the region). Delivery destinations are optional references (at least one of CloudWatch/S3, enforced by CEL), so prerequisites stay empty and E2E fixtures ride scenario annotations.
- `AwsBedrockAgent` -- AwsIamRole is a prerequisite because the Bedrock service assumes the agent resource role to invoke models, action-group Lambdas, and knowledge bases; the guardrail, KMS key, provisioned throughput, and collaborator/knowledge-base edges are optional composition (e2e-prerequisites annotation). Action groups, aliases, collaborators, and knowledge-base associations are folded satellites of the agent.
- `AwsBedrockKnowledgeBase` -- AwsIamRole is a prerequisite because the Bedrock service assumes the knowledge-base role to read data sources, call the embedding model, and read/write the vector store; the vector-store and data-source reference edges (OpenSearch, S3, Secrets Manager, ...) are optional composition (e2e-prerequisites annotation). Data sources are folded satellites of the knowledge base.
- `AwsBedrockFlow` -- AwsIamRole is a prerequisite because the Bedrock service assumes the flow execution role to invoke the models, agents, knowledge bases, and Lambdas its nodes reference; the node-level reference edges are optional composition (e2e-prerequisites annotation).
- `AwsBedrockPrompt` -- A dependency-free leaf: variants target AWS-listed foundation models by ID; targeting another agent's alias is optional composition (e2e-prerequisites annotation).
- `AwsBedrockAgentCoreRuntime` -- AwsIamRole is a prerequisite because the AgentCore service assumes the runtime role to pull the container image or read the S3 code bundle and to run the hosted agent; the code-bundle S3 bucket and VPC placement edges are optional composition (e2e-prerequisites annotation). Endpoints and the runtime's resource policy are folded satellites of the runtime.
- `AwsBedrockAgentCoreGateway` -- AwsIamRole is a prerequisite because the gateway assumes its role to reach targets (invoke Lambdas, sign SigV4 requests); the target and credential-provider reference edges (runtime, Lambda, Identity providers, policy engine) are optional composition (e2e-prerequisites annotation). Targets are folded satellites of the gateway - AWS deletes them before the gateway at destroy.
- `AwsBedrockAgentCoreMemory` -- A dependency-free leaf for built-in strategies: the execution role (custom strategies, Kinesis delivery), KMS key, and Kinesis stream edges are optional composition (e2e-prerequisites annotation). Strategies are folded satellites of the memory - AWS serializes their changes through the parent.
- `AwsBedrockAgentCoreIdentity` -- A dependency-free leaf: workload identities, credential providers, and the Cedar policy engine with its policies are all name-keyed arms of one identity-and-access bundle; the KMS key edge is optional composition (e2e-prerequisites annotation). The account/region token-vault CMK is deliberately NOT modeled here (settings singleton).
- `AwsBedrockAgentCoreTools` -- A dependency-free leaf in the SANDBOX/PUBLIC postures: the execution role (recordings, certificates), S3, Secrets Manager, and VPC edges are optional composition (e2e-prerequisites annotation). Browsers, profiles, and code interpreters are name-keyed arms of one tools bundle; AWS exposes no update - every field change recreates the tool.
- `AwsBedrockAgentCoreEvaluation` -- The AgentCore Evaluations bundle - evaluators (LLM-judge or Lambda scorers), harnesses (repeatable agent test benches), and online evaluation configs (continuous scoring of sampled production sessions). Deploys standalone - no arm requires an agent runtime to exist. No registry prerequisite: every arm is optional, so no dependency is required for the kind to function (scenarios compose IAM roles via annotations).
- `AwsBedrockAgentCoreTokenVault` -- Account/region settings singleton: sets the KMS key on the ONE default AgentCore token vault. The KMS reference is conditional on key_type (CEL-enforced), so prerequisites stay empty and E2E fixtures ride scenario annotations.
- `AwsSagemakerModel` -- The immutable serving definition (container image + artifacts + execution role) that endpoints deploy - one container or an inference pipeline.
- `AwsSagemakerEndpoint` -- A real-time inference endpoint WITH its folded endpoint configuration - the configuration is immutable upstream, so the modules roll name-suffixed configurations create-before-destroy and repoint the endpoint.
- `AwsSagemakerNotebookInstance` -- A managed Jupyter notebook EC2 instance with its folded lifecycle configuration (bootstrap scripts).
- `AwsSagemakerFeatureGroup` -- A Feature Store feature group - online and/or offline stores over a declared feature schema.
- `AwsSagemakerModelRegistry` -- A model registry package group with its folded resource policy - model package VERSIONS register into it imperatively (training pipelines), never declaratively.
- `AwsSagemakerPipeline` -- An ML workflow DAG (the SageMaker pipeline-definition JSON) that executions run against - free to create, billed per execution.
- `AwsSagemakerImage` -- A named registry entry exposing YOUR container images to Studio, with folded AWS-numbered versions (append-only by position).
- `AwsSagemakerMlflowServer` -- The classic hourly-billed managed MLflow tracking server (~25 min to provision; billed hourly from creation whether or not anyone is tracking). The serverless successor is AwsSagemakerMlflowApp.
- `AwsSagemakerMlflowApp` -- The serverless MLflow 3.x deployment (billed per use) - standalone, associating with SageMaker domains; NOT a tracking-server satellite.
- `AwsRestApiGateway` -- A full REST API (API Gateway v1): the resource/method tree with inline integrations (or an imported OpenAPI document), one stage with an explicit hash-triggered deployment, and the API-scoped satellites (authorizers, models, validators, gateway responses, policy, documentation, client certificate). Self-contained: a MOCK-integration API needs no other resource.
- `AwsRestApiDomain` -- A custom domain for REST APIs with base-path mappings and - for PRIVATE domains - VPC-endpoint access associations. AwsCertManagerCert is a prerequisite because the domain cannot be created without a TLS certificate covering it.
- `AwsRestApiUsagePlan` -- A usage plan metering REST API consumers - stage coverage, quota, throttles, and the API keys it admits. No registry prerequisite: a plan is valid with no stage coverage (scenarios compose the REST API via annotations).
- `AwsRestApiVpcLink` -- A REST API VPC link fronting an internal Network Load Balancer so REST integrations reach private services. AwsNlb is a prerequisite because AWS rejects link creation without the target balancer.
- `AwsApiGatewayAccountSettings` -- Region settings singleton (one API Gateway account object per account+region; identity = the region). The CloudWatch role is an optional reference (unset = the explicit no-logging posture), so prerequisites stay empty and E2E fixtures ride scenario annotations.
- `AwsCloudTrail` -- The account's API audit trail. AwsS3Bucket is a prerequisite because AWS rejects trail creation without a delivery bucket carrying the CloudTrail service-principal policy. 1240 opens the governance sub-band (1240-1249).
- `AwsConfigRecorder` -- Region singleton (one AWS Config recorder per region, named "default" by AWS; identity = the region). AwsIamRole is a prerequisite because the recorder cannot exist without its service role.
- `AwsConfigRule` -- One AWS Config compliance rule (managed, custom-lambda, or custom-policy; account- or organization-scoped) with optional auto-remediation. Managed rules need no prerequisites; the custom-lambda arm's function reference is conditional, so E2E fixtures ride scenario annotations.
- `AwsGuardDuty` -- Region singleton (AWS allows one GuardDuty detector per account+region; the detector has no name - identity = the region). Satellite references (S3 export bucket, KMS key) are conditional, so E2E fixtures ride scenario annotations.
- `AwsCloudTrailEventDataStore` -- CloudTrail Lake: a queryable, immutable event data store with its own retention and billing lifecycle - no trail required. The KMS key reference is conditional, so E2E fixtures ride scenario annotations.
- `AwsConfigAggregator` -- AWS Config cross-account/cross-region aggregation: the aggregator (collector side) and/or the reciprocal authorization grants (source-account side). Works with zero recorders; the org-source role reference is conditional, so E2E fixtures ride scenario annotations.
- `AwsConfigConformancePack` -- An AWS Config conformance pack (account- or organization-scoped): a template bundle that creates its own Config rules. Deployment requires an active Config recorder in the region (a service-side requirement, not a spec reference), so E2E fixtures ride scenario annotations.
- `AwsGuardDutyMalwareProtectionPlan` -- GuardDuty Malware Protection for S3: scans new objects in one bucket - a standalone plan protecting a bucket, not a detector satellite (its schema carries no detector reference). The execution role and the protected bucket are required references.
- `AwsBackupVault` -- An AWS Backup vault - the encrypted container recovery points live in, as either a standard vault (with its lock, access policy, and notification satellites) or a logically air-gapped vault (AWS's own VaultType discriminator). The KMS and SNS references are conditional, so E2E fixtures ride scenario annotations. 1250 opens the backup sub-band (1250-1259).
- `AwsBackupPlan` -- An AWS Backup plan: scheduled backup rules plus the resource selections that assign resources to them. AwsBackupVault is a prerequisite because every rule requires a target vault; the selections' IAM role is conditional and rides scenario annotations.
- `AwsBackupFramework` -- A Backup Audit Manager framework: compliance controls evaluating backup posture. No schema-required references (the Config recorder its evaluations need is a lane fixture, not a spec reference).
- `AwsBackupReportPlan` -- A Backup Audit Manager report plan: scheduled compliance/job reports delivered to S3. AwsS3Bucket is a prerequisite because the delivery channel's bucket is required.
- `AwsBackupRestoreTestingPlan` -- An AWS Backup restore testing plan with its folded selections: scheduled restore tests proving recovery points actually restore. Vault targeting accepts the "*" wildcard, so fixtures are conditional and ride scenario annotations.
- `AwsBackupSettings` -- Account/region settings singleton for AWS Backup: the account's global settings (cross-account backup) and the region's resource-type opt-in/management preferences. Both provider deletes are no-ops - settings persist after destroy.
- `AwsSsmParameter` -- An SSM Parameter Store entry (String/StringList/SecureString). The parameter's name is an explicit spec field - names are hierarchical paths ("/prod/db/url") metadata.name cannot carry. The KMS reference is conditional (SecureString only), so E2E fixtures ride scenario annotations. 1260 opens the SSM sub-band (1260-1269).
- `AwsSsmDocument` -- A customer-owned SSM document (Command/Automation/Session/...): reusable action definitions managed nodes and automations execute. State Manager associations are their own AwsSsmAssociation kind - an association binds ANY document (AWS-managed included), so it is not this document's satellite.
- `AwsSsmMaintenanceWindow` -- An SSM maintenance window with its folded target registrations and tasks (Run Command / Automation / Lambda / Step Functions) - the targets and tasks are true window satellites (ForceNew window_id edges). Identity is the AWS-generated "mw-..." id.
- `AwsSsmPatchBaseline` -- An SSM patch baseline with its folded patch-group registrations and the account/region default-baseline designation (delete RESTORES AWS's own predefined default for the OS). Identity is the AWS-generated "pb-..." id.
- `AwsSsmAssociation` -- A State Manager association: the binding of an SSM document to targets on a schedule. Split from the document kind because the document reference is a free string with no structural edge - associations routinely bind AWS-managed documents (AWS-RunShellScript, ...) with no user document anywhere, so no registry prerequisite either. Identity is the AWS-generated association UUID.
- `AwsOrganization` -- THE AWS Organization of the deploying account - creating it makes the caller the management account. Trusted service access, delegated administrators, the org's singleton resource policy, and centralized root-access management (IAM's organizations features - a management-account act requiring iam.amazonaws.com trusted access) fold in (none has a life of its own; the standalone service-access resource fights the org's own argument with a perpetual diff). Deleting this deletes the entire organization. 1270 opens the Organizations sub-band (1270-1279). A container kind: the organization is the outer wall of the tenancy tree AWS's own console draws -- organizational units are rooms inside it and member accounts stand inside those. A first-level unit names the organization through its parent reference and is placed inside it; a policy attached to the root names the organization too, but a policy is a guard applied from above, never a resident, and its attachment reference is containment-exempt on its spec.
- `AwsOrganizationalUnit` -- An organizational unit in the org's OU tree. The display name is an explicit spec field (OU names allow spaces metadata.name cannot carry); the parent reference (root or parent OU) is required and immutable, so the organization is a registry prerequisite. A container kind: a unit is a room of the tenancy tree -- nested units and member accounts name it through their parent reference and are placed inside it. A policy attached to a unit names it too and stands beside the room (containment-exempt on the policy's spec): one policy attaches to many units and cannot live in all of them.
- `AwsOrganizationAccount` -- A MEMBER account of the organization: creation, OU placement, and the account-level settings satellites (alternate/primary contacts, opt-in region enablement) fold onto the created account's ID. Destroy is never a clean delete (remove-from-org or ~90-day close) - taught on the spec. No registry prerequisite by the schema-required-only rule (the OU parent reference is optional).
- `AwsOrganizationPolicy` -- An Organizations policy (SCP and its twelve sibling types) with its folded attachments to roots, OUs, and member accounts. The policy type must be enabled on the organization first; AWS-managed policies are never adopted. No registry prerequisite by the schema-required-only rule (attachments are optional).
- `AwsBudget` -- A Budgets budget (COST/USAGE/RI/Savings Plans coverage and utilization) with its folded budget actions as name-keyed satellites - an action exists only on its budget and fires an IAM-policy application, an SCP attachment, or SSM instance stops when a threshold breaches. Budgets is account-global (served from us-east-1; the spec region is the provider endpoint). 1280 opens the cost-management sub-band (1280-1289).
- `AwsCostAnomalyMonitor` -- A Cost Explorer anomaly monitor (DIMENSIONAL over one dimension, or CUSTOM over a CE expression) with its folded alert subscriptions - a subscription's monitor list is the structural edge that makes it this monitor's satellite. Account-global; AWS identifies both by ARN.
- `AwsCostCategory` -- A Cost Explorer cost category: ordered rules (regular expression rules or inherited-value rules) over the recursive CE expression tree, plus split-charge rules. The account's cost-allocation-tag activation toggle is deliberately NOT folded here - it is a per-tag-key account feature with no edge to any category, so many category instances would fight over one account object.
- `AwsIamGroup` -- An IAM group with its folded declarative membership (the authoritative users list) and group policies - name-keyed inline documents plus managed-policy attachments. IAM is global; identity is the group name (renames update in place, the ARN recomputes). 1290 opens the IAM P1 sub-band (1290-1299).
- `AwsIamSamlProvider` -- An IAM SAML identity provider: the account's federation trust anchor, created from the IdP's metadata XML (a public document carrying certificates, not a secret). Identity is the provider ARN; the name is write-once.
- `AwsIamAccountSettings` -- Account settings singleton for IAM (a GLOBAL service - one object per ACCOUNT, not per region): the sign-in alias, the password policy, and the STS global-endpoint token version. Destroy contracts DIFFER per arm (each taught on its arm): the alias truly deletes, the password policy resets to AWS defaults, the STS preference is a no-op delete that persists.
- `AwsCloudwatchDashboard` -- A CloudWatch dashboard: one named dashboard whose widget layout is the dashboard-body JSON document (modeled as a typed Struct, the catalog's uniform policy-document idiom). Dashboards are untaggable at AWS. Identity is the dashboard name; every change is an in-place PutDashboard upsert. 1300 opens the CloudWatch observability P1 sub-band (1300-1309).
- `AwsCloudwatchSynthetics` -- CloudWatch Synthetics: a canary (a scheduled scripted probe running from an S3-staged code bundle under an execution role, writing run artifacts to S3) plus the grouping surface - owned groups and the canary's group associations (joins by group NAME, so shared groups are referenced, never fought over). A groups-only instance manages shared groups with no canary.
- `AwsCloudwatchLogDelivery` -- CloudWatch Logs delivery: the two ways logs leave CloudWatch. The vended-log arm pivots on a delivery SOURCE (one AWS resource whose service vends logs) with name-keyed deliveries fanning out to delivery destinations (S3 / CloudWatch Logs / Firehose / X-Ray), each created inline or referenced by ARN. The cross-account arm is the legacy Kinesis subscription destination with its access policy (whose delete is a no-op at AWS - the policy persists).
- `AwsCloudwatchLogAccountPolicy` -- A CloudWatch Logs account-level policy: one policy object per (name, type) pair per region - data protection, subscription filter, field index, transformer, or metric extraction - applied account-wide, optionally narrowed by selection criteria. Standalone account configuration, never a per-log-group satellite.
- `AwsCloudwatchLogAnomalyDetector` -- A CloudWatch Logs anomaly detector: one detector trains over a LIST of log groups (multi-parent scope - never a single group's satellite), surfacing anomalies on a chosen evaluation frequency with a bounded visibility window.
- `AwsCloudwatchLogResourcePolicy` -- A CloudWatch Logs resource policy: the account-scoped named policy (or resource-scoped policy on one log group ARN) that grants AWS services permission to write logs - Route53 query logging, EventBridge, and friends. Exactly one scope per instance.
- `AwsManagedPrometheus` -- An Amazon Managed Prometheus workspace with its folded satellites: workspace configuration (retention, label-set limits - a created-via-update singleton whose delete is a no-op at AWS), the alert manager definition (strictly one per workspace), name-keyed rule group namespaces, query logging, the workspace resource policy, and alias-keyed anomaly detectors. Scrapers are deliberately NOT folded here - a scraper can target CloudWatch with zero AMP workspaces, so it is its own kind.
- `AwsManagedPrometheusScraper` -- An Amazon Managed Prometheus scraper: the agentless collector. Source is an EKS cluster or a bare VPC placement (both replace-on-change); destination is an AMP workspace or a CloudWatch dataset. Carries its own scraper logging configuration satellite. Scrape configuration is optional on the EKS arm (AWS publishes a default, resolved at deploy) and required on the VPC arm.
- `AwsEventBridgePipe` -- An EventBridge Pipe: one point-to-point integration reading from one source (SQS, Kinesis, DynamoDB streams, MSK or self-managed Kafka, ActiveMQ/RabbitMQ), optionally filtering and enriching in-flight, and delivering to one target (ECS, Batch, Lambda, Step Functions, Kinesis, SQS, Redshift, SageMaker, CloudWatch Logs, EventBridge buses, HTTP via API destinations). The source is fixed for life (replace-on-change); the target swaps in place. 1310 opens the EventBridge extras P1 sub-band (1310-1319).
- `AwsEventBridgeScheduler` -- An EventBridge Scheduler schedule: cron/rate/one-time invocation of one target under an execution role, with flexible time windows, retry policy, and a dead-letter queue. The schedule GROUP is folded own-XOR-existing (a name-and-tags container - the provider's own update path is tags-only); unset means AWS's default group.
- `AwsEventBridgeApiDestination` -- An EventBridge API destination with its connection: the authenticated HTTP(S) endpoint rules, pipes, and schedules invoke. Two independently deployable arms - the CONNECTION (the shareable auth trust anchor: api-key, basic, or OAuth credentials that AWS stores in Secrets Manager) and the DESTINATION (endpoint + method + rate limit) whose connection is owned inline or referenced by ARN.
- `AwsVpcPeering` -- A VPC peering connection, as a request-XOR-accept mode union: the REQUEST arm creates the peering from its VPC toward a peer VPC (same-account auto-accept supported; cross-account/cross-region stays pending until accepted), the ACCEPT arm adopts and accepts a pending connection by ID from the accepter side. DNS-resolution options fold into both arms. 1320 opens the VPC networking P1 sub-band (1320-1329).
- `AwsNetworkAcl` -- A network ACL: the stateless subnet-level firewall - ordered ingress/egress rules (allow or deny, evaluated by rule number) and the subnet associations, all folded in-line as the single declarative owner (the standalone rule/association resources are the same payload and fight the in-line form).
- `AwsManagedPrefixList` -- A customer-managed prefix list: a named, versioned set of CIDR blocks that security-group rules, NACL rules, and route tables reference as one object. Entries fold in-line; max_entries is the capacity contract (referencing consumes that many rule slots regardless of how many entries exist).
- `AwsEbsVolume` -- A standalone EBS volume as a create-XOR-copy union (fresh in a zone, or cloned from another volume) with attachments managed in-line. 1330 opens the block & object storage sub-band (1330-1339).
- `AwsEbsSnapshot` -- An EBS snapshot as a three-way source union (snapshot a volume, copy a snapshot, or import a disk image) with archive tiering, fast snapshot restore, and cross-account share grants in-line.
- `AwsS3DirectoryBucket` -- An S3 directory bucket (S3 Express One Zone): single-AZ, single-digit-millisecond object storage. The modules derive the mandated "{name}--{zone_id}--x-s3" bucket name.
- `AwsS3TableBucket` -- An S3 table bucket (S3 Tables - managed Apache Iceberg storage) with its namespaces, tables, policies, and replication folded in-line as the single declarative owner.
- `AwsS3VectorBucket` -- An S3 vector bucket (AI embedding storage with similarity query) with its vector indexes folded in-line - the natural backend for Bedrock knowledge bases.
- `AwsDlmLifecyclePolicy` -- A Data Lifecycle Manager policy: account-level, tag-targeted snapshot/AMI automation (create, retain, archive, copy cross-region, share, deprecate) as a default-XOR-custom mode union. AwsIamRole is a prerequisite because DLM acts through a required execution role.
- `AwsRoute53ResolverEndpoint` -- A Route 53 Resolver endpoint (the hybrid-DNS bridge between a VPC and outside networks) with its forwarding rules and their VPC associations managed in-line. Subnets place the ENIs and security groups guard them - both schema-required. 1340 opens the DNS & service discovery sub-band (1340-1349).
- `AwsRoute53ResolverFirewall` -- A Route 53 Resolver DNS Firewall rule group with its domain lists, filtering rules, and VPC associations managed in-line - the DNS-layer block/allow policy for VPC egress queries. AwsVpc is a prerequisite because the association arm filters a referenced VPC.
- `AwsRoute53ResolverQueryLog` -- A Resolver query logging configuration (every DNS query VPCs make through the resolver, to CloudWatch Logs / S3 / Firehose) with its VPC associations managed in-line. AwsVpc is a prerequisite because the association arm logs a referenced VPC.
- `AwsCloudMapNamespace` -- An AWS Cloud Map namespace (HTTP-XOR-private-DNS-XOR-public-DNS) with its discoverable services and statically registered instances managed in-line - the service-discovery registry ECS and custom applications look each other up in. AwsVpc is a prerequisite because the private-DNS arm binds its hosted zone to a referenced VPC.
- `AwsAppSyncApi` -- An AppSync API - AWS's managed API service, as a GraphQL API (SDL schema, resolvers over data sources, caching, the MERGED federation variant) XOR an Events API (real-time pub/sub over channel namespaces) - with its data sources, resolvers, functions, types, API keys, and custom domain managed in-line. Every backend reference (data source targets, roles, the certificate) is optional, so no registry prerequisite - lanes exercise the fixture-free arms.
- `AwsLambdaLayer` -- A Lambda layer version - a shared code archive (libraries, custom runtimes) functions attach by ARN - with its cross-account and organization share grants managed in-line. The archive lives in S3 (an optional reference, so no registry prerequisite - lanes compose their own bucket fixture). 1351 sits in the app & data services sub-band (1350-1359; 1350 opens it with AwsAppSyncApi).
- `AwsRdsProxy` -- An RDS Proxy - the managed connection pool between connection-hungry applications and a database - with its connection-pool tuning, additional endpoints, and database target managed in-line. AwsIamRole is a prerequisite because the proxy assumes a required role to read database credentials from Secrets Manager; AwsSubnet because the proxy's network interfaces require at least two subnets.
- `AwsAuroraDsql` -- An Aurora DSQL cluster - serverless, PostgreSQL-compatible distributed SQL with active-active multi-region pairing managed in-line. No prerequisites: a single-region cluster deploys from defaults alone (the KMS and peer references are optional arms).
- `AwsEcrRegistrySettings` -- Region settings singleton (one private ECR registry per account+region): the registry policy, scanning configuration, replication rules, pull-through cache rules, repository creation templates, account settings, and pull-time update exclusions. Repository-scoped surface stays on AwsEcrRepo.
- `AwsPrivateCa` -- An AWS Private Certificate Authority with composed activation (a ROOT self-signs at apply; a subordinate activates from a parent AwsPrivateCa), issued certificates, the ACM renewal permission, and the resource policy managed in-line. No prerequisites: the S3 (CRL) and parent-CA references are optional arms.
- `AwsSesAccountSettings` -- Account/region settings singleton (one SES account object per account+region): the suppression list and VDM posture. 1360 opens the SES P1 sub-band (1360-1369).
- `AzureResourceGroup` -- 2000–2999: Azure resources
- `AzureAksCluster` -- AzureResourceGroup is the only required parent: the cluster is created inside a referenced resource group. Subnet is optional on the default node pool (AKS provisions managed networking when unset).
- `AzureAksNodePool` -- AzureAksCluster is a prerequisite because a node pool attaches to an existing cluster by ARM ID; the resource group chains transitively.
- `AzureContainerRegistry` -- AzureResourceGroup is a prerequisite because a container registry is created inside a resource group.
- `AzureDnsZone` -- AzureResourceGroup is a prerequisite because the DNS zone is created inside a referenced resource group that must already exist.
- `AzureKeyVault` -- AzureResourceGroup is a prerequisite because a key vault is created inside a referenced resource group in composed environments.
- `AzureVirtualNetwork` -- AzureResourceGroup is a prerequisite because a virtual network is created inside a referenced resource group in composed environments.
- `AzureNatGateway` -- AzureResourceGroup is a prerequisite because a NAT gateway is created inside a referenced resource group in composed environments.
- `AzureVirtualMachine` -- AzureNetworkInterface is a prerequisite because a virtual machine attaches at least one NIC (the subnet, network, and resource group chain transitively through the NIC's own prerequisites).
- `AzureStorageAccount` -- AzureResourceGroup is a prerequisite because a storage account is created inside a referenced resource group in composed environments.
- `AzureDnsRecord` -- AzureDnsZone is a prerequisite because a record set is created inside a referenced zone (the resource group chains transitively through the zone). Public DNS zone names are not globally unique, so a shared zone fixture is safe to recreate across scenarios.
- `AzureSubnet` -- AzureVirtualNetwork is a prerequisite because a subnet is an ARM child of a referenced network -- the network must exist before the subnet can be written. (The resource group arrives transitively through the network's own prerequisite declaration.)
- `AzureNetworkSecurityGroup` -- AzureResourceGroup is a prerequisite because a network security group is created inside a referenced resource group in composed environments.
- `AzurePublicIp` -- AzureResourceGroup is a prerequisite because a public IP is created inside a referenced resource group in composed environments.
- `AzurePrivateEndpoint` -- AzureSubnet is a prerequisite because a private endpoint draws its private IP from a referenced subnet (the virtual network and resource group chain transitively through the subnet's own prerequisite). The connection target is polymorphic and the DNS zones / ASGs are optional, so none of those are prerequisites.
- `AzurePrivateDnsZone` -- AzureResourceGroup is a prerequisite because a private DNS zone is created inside a referenced resource group in composed environments.
- `AzureApplicationGateway` -- AzureSubnet is a prerequisite because a gateway cannot exist without its dedicated gateway_ip_configuration subnet (the network and resource group chain transitively through the subnet's own prerequisites); public frontends additionally reference a public IP, but private-only gateways are legal, so it is not a registry prerequisite.
- `AzureLoadBalancer` -- AzureResourceGroup is a prerequisite because a load balancer is created inside a referenced resource group (frontends additionally reference subnets or public IPs, but neither is universally required, so they are not registry prerequisites).
- `AzureRouteTable` -- AzureResourceGroup is a prerequisite because a route table is created inside a referenced resource group in composed environments.
- `AzurePrivateDnsZoneVirtualNetworkLink` -- AzurePrivateDnsZone and AzureVirtualNetwork are prerequisites because a virtual network link is a child resource of a referenced zone and binds it to a referenced network -- both must exist before the link can be written. (The resource group arrives transitively through the zone's and network's own prerequisite declarations.)
- `AzureVirtualNetworkPeering` -- AzureVirtualNetwork is a prerequisite because a peering is an ARM child of its local network and binds it to a remote network -- the local network must exist before the peering can be written. (The resource group arrives transitively through the network's own prerequisite declaration.)
- `AzurePublicIpPrefix` -- AzureResourceGroup is a prerequisite because a public IP prefix is created inside a referenced resource group in composed environments.
- `AzureNetworkInterface` -- AzureSubnet is a prerequisite because a network interface's IP configurations deploy into a subnet (the virtual network and resource group chain transitively through the subnet's own prerequisite).
- `AzureManagedDisk` -- AzureResourceGroup is a prerequisite because a managed disk is created inside a resource group.
- `AzureVirtualMachineScaleSet` -- AzureSubnet is a prerequisite because every scale-set instance's network interface deploys into a subnet (the virtual network and resource group chain transitively through the subnet's own prerequisite).
- `AzureKeyVaultKey` -- AzureKeyVault is a prerequisite because a key is a data-plane object inside a referenced vault -- the vault must exist before the key can be written (the resource group chains transitively through the vault's own prerequisite).
- `AzureKeyVaultCertificate` -- AzureKeyVault is a prerequisite because a certificate is a data-plane object inside a referenced vault -- the vault must exist before the certificate can be enrolled or imported (the resource group chains transitively through the vault's own prerequisite).
- `AzureKeyVaultSecret` -- AzureKeyVault is a prerequisite because a secret is a data-plane object inside a referenced vault -- the vault must exist before the secret can be written (the resource group chains transitively through the vault's own prerequisite). Part of the Key Vault family (2005, 2025-2026) despite the out-of-run number -- enum numbers are pinned by the registry snapshot; never renumber.
- `AzureWebApplicationFirewallPolicy` -- AzureResourceGroup is a prerequisite because a WAF policy is created inside a referenced resource group; the Application Gateways that attach the policy reference it, never the reverse.
- `AzureApplicationSecurityGroup` -- AzureResourceGroup is a prerequisite because an application security group is created inside a referenced resource group; network interfaces, scale-set IP configurations, and NSG security rules reference the group, never the reverse.
- `AzureDiskEncryptionSet` -- AzureKeyVaultKey is a prerequisite because a disk encryption set wraps customer data with a referenced key -- the key (and its vault, which chains transitively) must exist before the set can resolve the key URL at create time.
- `AzurePostgresqlFlexibleServer` -- AzureResourceGroup is a prerequisite because a server is created inside a referenced resource group (VNet injection additionally references a delegated subnet and a private DNS zone, but neither is universally required, so they are not registry prerequisites).
- `AzureRedisCache` -- AzureResourceGroup is a prerequisite because the cache is created inside a referenced resource group (VNet injection additionally references a dedicated subnet, but only the Premium tier supports it, so it is not a registry prerequisite).
- `AzureCosmosdbAccount` -- AzureResourceGroup is a prerequisite because the account is created inside a referenced resource group.
- `AzureMssqlServer` -- AzureResourceGroup is a prerequisite because the logical server is created inside a referenced resource group.
- `AzureMysqlFlexibleServer` -- AzureResourceGroup is a prerequisite because a server is created inside a referenced resource group (VNet injection additionally references a delegated subnet and a private DNS zone, but neither is universally required, so they are not registry prerequisites).
- `AzureMssqlDatabase` -- The parent logical server is referenced via server_id, not auto-deployed: E2E scenarios declare their own server fixture (minimal-server.yaml or the pool-attach chain through AzureMssqlElasticPool) so sequential subtests never destroy and recreate the same globally unique server_name.
- `AzureMssqlElasticPool` -- AzureMssqlServer is a prerequisite because every elastic pool lives on a referenced logical server (the server's resource group is transitive).
- `AzureRedisLinkedServer` -- The target and linked caches are referenced via ARM ids, not auto-deployed: caches are the slowest-provisioning resources in the Azure catalog and their names are globally unique, so E2E scenarios declare their own cache fixtures instead of a registry prerequisite recreating a shared one per run.
- `AzureRedisCacheAccessPolicy` -- The parent cache is referenced via redis_cache_id, not auto-deployed: caches are the slowest-provisioning resources in the Azure catalog and their names are globally unique, so E2E scenarios declare their own cache fixtures instead of a registry prerequisite recreating a shared one per run.
- `AzureRedisCacheAccessPolicyAssignment` -- The parent cache is referenced via redis_cache_id, not auto-deployed: caches are the slowest-provisioning resources in the Azure catalog and their names are globally unique, so E2E scenarios declare their own cache fixtures instead of a registry prerequisite recreating a shared one per run.
- `AzureContainerAppEnvironment` -- AzureResourceGroup is a prerequisite because the environment is created inside a referenced resource group that must already exist.
- `AzureContainerApp` -- AzureContainerAppEnvironment is a prerequisite because every app runs inside a referenced environment (the resource group arrives transitively through it).
- `AzureServicePlan` -- AzureResourceGroup is a prerequisite because the plan is created inside a referenced resource group that must already exist.
- `AzureFunctionApp` -- AzureServicePlan is a prerequisite because a function app runs on a referenced plan (the resource group arrives transitively through the plan). The required storage account is deliberately NOT a registry prerequisite: storage-account names are globally unique, so scenarios bring their own scenario-local account fixtures.
- `AzureLinuxWebApp` -- AzureServicePlan is a prerequisite because a web app runs on a referenced plan (the resource group arrives transitively through the plan).
- `AzureContainerAppJob` -- AzureContainerAppEnvironment is a prerequisite because a job runs inside a referenced environment (the resource group arrives transitively through it).
- `AzureContainerAppEnvironmentStorage` -- AzureContainerAppEnvironment is a prerequisite because the storage registration lives on a referenced environment. The Azure Files share and storage account are deliberately NOT registry prerequisites: storage-account names are globally unique, so scenarios bring their own scenario-local account + share fixtures.
- `AzureContainerAppEnvironmentDaprComponent` -- AzureContainerAppEnvironment is a prerequisite because the Dapr component is registered on a referenced environment.
- `AzureContainerAppEnvironmentCertificate` -- AzureContainerAppEnvironment is a prerequisite because the certificate is stored on a referenced environment.
- `AzureContainerAppEnvironmentManagedCertificate` -- AzureContainerAppEnvironment is a prerequisite because the managed certificate is provisioned on a referenced environment.
- `AzureLogAnalyticsWorkspace` -- AzureResourceGroup is a prerequisite because the workspace is created inside a referenced resource group that must already exist.
- `AzureApplicationInsights` -- AzureLogAnalyticsWorkspace is a prerequisite because workspace-based Application Insights stores its telemetry in a referenced workspace (the resource group chains transitively through the workspace).
- `AzureMonitorDiagnosticSetting` -- AzureLogAnalyticsWorkspace is a prerequisite because the setting's scenarios route a fixture workspace's telemetry (the workspace doubles as target and destination); the target itself is polymorphic.
- `AzureMonitorActionGroup` -- AzureResourceGroup is a prerequisite because the action group is created inside a referenced resource group that must already exist.
- `AzureMonitorMetricAlert` -- AzureMonitorActionGroup is a prerequisite because a metric alert's actions fire into a referenced action group (the resource group chains transitively); alert scopes are polymorphic.
- `AzureMonitorScheduledQueryAlert` -- AzureLogAnalyticsWorkspace is a prerequisite because the rule queries a referenced workspace scope; AzureMonitorActionGroup because its action fires into a referenced action group.
- `AzureMonitorActivityLogAlert` -- AzureMonitorActionGroup is a prerequisite because an activity log alert's actions fire into a referenced action group (the resource group chains transitively). The alert itself is subscription-global and its scopes are polymorphic.
- `AzureApplicationInsightsStandardWebTest` -- AzureApplicationInsights is a prerequisite because a standard web test binds to a referenced Application Insights component (the resource group chains transitively through the kind).
- `AzureUserAssignedIdentity` -- AzureResourceGroup is a prerequisite because the identity is created inside a referenced resource group that must already exist.
- `AzureRoleAssignment` -- AzureResourceGroup and AzureUserAssignedIdentity are prerequisites because an assignment grants a role at a referenced scope (most commonly a resource group) to a referenced principal (most commonly a managed identity) -- both must exist before the grant can be written.
- `AzureRoleDefinition` -- AzureResourceGroup is a prerequisite because a custom role definition is created at a referenced scope, most commonly a resource group in composed environments -- the scope must exist before the definition can be written.
- `AzureFederatedIdentityCredential` -- AzureUserAssignedIdentity is the prerequisite because a federated identity credential is a child resource of a referenced managed identity -- the identity must exist before the credential can be written on it. (The resource group arrives transitively through the identity's own prerequisite declaration.)
- `AzureServiceBusNamespace` -- AzureResourceGroup is a prerequisite because a Service Bus namespace is created inside a referenced resource group in composed environments. The namespace is the container every Service Bus messaging entity (queue, topic, subscription, authorization rule, geo-DR pairing) nests under. The child kinds deliberately declare NO registry prerequisite: namespace names are globally unique with a post-delete name hold, so E2E composes them with scenario-local namespace fixtures instead of a shared recreate-per-scenario prerequisite.
- `AzureEventHubNamespace` -- AzureResourceGroup is a prerequisite because an Event Hub namespace is created inside a referenced resource group in composed environments. The namespace is the container every Event Hubs entity (event hub, consumer group, authorization rule, schema group, geo-DR pairing, customer-managed key) nests under. The child kinds deliberately declare NO registry prerequisite: namespace names are globally unique with a post-delete name hold, so E2E composes them with scenario-local namespace fixtures instead of a shared recreate-per-scenario prerequisite.
- `AzureServiceBusQueue`
- `AzureServiceBusTopic`
- `AzureServiceBusSubscription`
- `AzureServiceBusAuthorizationRule`
- `AzureServiceBusDisasterRecoveryConfig`
- `AzureEventHub`
- `AzureEventHubConsumerGroup`
- `AzureEventHubAuthorizationRule`
- `AzureFrontDoorProfile` -- AzureResourceGroup is a prerequisite because a Front Door profile is created inside a referenced resource group in composed environments. The profile is the container every Front Door delivery resource (endpoint, origin group, origin, route) nests under.
- `AzureFrontDoorEndpoint` -- AzureFrontDoorProfile is a prerequisite because an endpoint is an ARM child of a referenced profile -- the profile must exist before the endpoint can be written. (The resource group arrives transitively through the profile's own prerequisite declaration.)
- `AzureFrontDoorOriginGroup` -- AzureFrontDoorProfile is a prerequisite because an origin group is an ARM child of a referenced profile. A container kind: every origin is an ARM child of its origin group, so on a diagram the origins stand inside the group that load-balances them (the profile's room holds the group). A route or a rule that forwards TO an origin group lives in its endpoint or its rule set, never in the group, and those references say so with containment_exempt.
- `AzureFrontDoorOrigin` -- AzureFrontDoorOriginGroup is a prerequisite because an origin is an ARM child of a referenced origin group (the profile and resource group chain transitively).
- `AzureFrontDoorRoute` -- A route attaches to an endpoint (its ARM parent) and forwards to an origin group whose origins must exist before ARM accepts the route -- so both the endpoint and the origin chain are genuine deploy-order prerequisites.
- `AzureFrontDoorRuleSet` -- AzureFrontDoorProfile is a prerequisite because a rule set is an ARM child of a referenced profile. The rules live inside the set (they form one ordered policy document); routes attach the set by ARM ID.
- `AzureFrontDoorCustomDomain` -- AzureFrontDoorProfile is a prerequisite because a custom domain is an ARM child of a referenced profile. The DNS zone and (for bring-your-own certificates) the Front Door secret are optional references, not deploy-order prerequisites.
- `AzureFrontDoorSecret` -- AzureFrontDoorSecret is a prerequisite-light kind: only the profile (its ARM parent) must exist. The Key Vault certificate it wraps is a reference resolved before the module runs; its vault chain is exercised through scenario-local fixtures in E2E.
- `AzureFrontDoorFirewallPolicy` -- AzureResourceGroup is a prerequisite because the Front Door WAF policy is created inside a referenced resource group -- it is a GLOBAL resource, not a profile child (a different ARM type than the regional Application Gateway WAF policy). Security policies attach it to profiles; the policy itself depends on nothing else.
- `AzureFrontDoorSecurityPolicy` -- A security policy is an ARM child of a profile that associates a referenced WAF policy with referenced domains -- so the endpoint (the default-domain association target; the profile arrives transitively through it) and the WAF policy are genuine deploy-order prerequisites.
- `AzureStorageContainer` -- None of the storage data-service kinds declares a registry prerequisite on AzureStorageAccount: account names are GLOBALLY unique and Azure holds a just-deleted name, so a recreate-per-scenario fixture would hang -- their E2E scenarios declare scenario-local account fixtures instead. Deploy ordering in composed environments still flows from the storage_account_id reference itself.
- `AzureStorageShare`
- `AzureStorageQueue`
- `AzureStorageTable`
- `AzureStorageEncryptionScope`
- `AzureStorageDataLakeGen2Filesystem`
- `AzureStorageLocalUser`
- `AzureStorageObjectReplication`
- `AzureCosmosdbSqlDatabase` -- None of the Cosmos DB data-service kinds declares a registry prerequisite on AzureCosmosdbAccount: account names are GLOBALLY unique DNS labels, so a recreate-per-scenario fixture would risk name-reuse hangs -- their E2E scenarios declare scenario-local account fixtures instead. Deploy ordering in composed environments still flows from the cosmosdb_account_id / parent-database references themselves.
- `AzureCosmosdbSqlContainer`
- `AzureCosmosdbMongoDatabase`
- `AzureCosmosdbMongoCollection`
- `AzureCosmosdbSqlRoleDefinition`
- `AzureCosmosdbSqlRoleAssignment`
- `AzureManagedRedis` -- AzureResourceGroup is the cluster's only registry prerequisite: the cluster is created inside a referenced resource group. The geo-replication and access-policy-assignment children declare NO prerequisite on AzureManagedRedis: clusters are expensive, slow-provisioning parents, so their E2E scenarios declare scenario-local cluster fixtures instead of recreating a shared one per scenario. Deploy ordering in composed environments still flows from the managed_redis_id references themselves.
- `AzureManagedRedisGeoReplication`
- `AzureManagedRedisAccessPolicyAssignment`
- `AzureEventHubDisasterRecoveryConfig`
- `AzureEventHubSchemaGroup`
- `AzureEventHubCluster` -- AzureResourceGroup is a prerequisite because a dedicated Event Hubs cluster is created inside a referenced resource group in composed environments. Note: clusters cannot be deleted for 4 hours after creation (Azure's moratorium), so E2E treats this kind as offline-gated.
- `AzureEventHubNamespaceCustomerManagedKey`
- `AzureMssqlFailoverGroup` -- AzureMssqlServer is a prerequisite because a failover group is created on a referenced primary logical server and points at a partner server; the primary (and its resource group, which chains transitively) must exist before the group can be written.
- `AzureContainerAppCustomDomain` -- AzureContainerApp is a prerequisite because the domain binding lives in a referenced app's ingress configuration (the environment and resource group chain transitively through the app).
- `AzureFirewallPolicy`
- `AzureFirewallPolicyRuleCollectionGroup` -- AzureFirewallPolicy is a prerequisite because a rule collection group is a child document of a referenced policy (the resource group chains transitively through the policy).
- `AzureFirewall` -- AzureSubnet is a prerequisite because a VNet-deployed firewall's data path lives in a dedicated subnet that must be named exactly "AzureFirewallSubnet" (the virtual network and resource group chain transitively through the subnet). The E2E install profile publishes a fixture subnet with that exact name and a /26 prefix.
- `AzureIpGroup`
- `AzureVirtualNetworkGateway` -- AzureSubnet is a prerequisite because every virtual network gateway lives in a dedicated subnet named exactly "GatewaySubnet" (the virtual network and resource group chain transitively through the subnet); the subnet install profile publishes a fixture instance with that exact ARM name. AzurePublicIp is a prerequisite because a VPN-type gateway (the default shape) requires a public IP per ip configuration; the address install profile publishes a dedicated zone-redundant instance (a gateway binds its address exclusively, and the AZ gateway SKUs require zones on it).
- `AzureVirtualNetworkGatewayConnection` -- Both gateways are prerequisites: a connection joins a virtual network gateway to a far side, and the site-to-site far side is a local network gateway (the GatewaySubnet, VNet, and resource group chain transitively through the virtual network gateway).
- `AzureLocalNetworkGateway`
- `AzurePrivateLinkService` -- AzureSubnet is the sole prerequisite: every NAT ip configuration draws its address from a subnet with private-link-service network policies disabled (the subnet install profile publishes a fixture instance with that flag off). The Standard load balancer whose frontend the service typically fronts is NOT a registry prerequisite -- the spec's destination is an exactly-one-of (load balancer frontend OR fixed destination IP), so scenarios that use the load-balancer shape declare it via the planton.dev/e2e-prerequisites annotation instead.
- `AzureExpressRouteCircuit`
- `AzureExpressRouteCircuitPeering` -- The circuit is the prerequisite: a peering is an ARM child of the circuit, addressed by the circuit's name (the resource group chains transitively through the circuit).
- `AzureExpressRouteGateway` -- The hub is the prerequisite: ARM requires an ExpressRoute Gateway to be deployed INTO a Virtual WAN hub (the WAN and resource group chain transitively through the hub).
- `AzureExpressRoutePort` -- ExpressRoute Port: your own physical port pair on a Microsoft edge router (ExpressRoute Direct), from whose bandwidth circuits are carved. Self-contained -- only the resource group is required.
- `AzureVirtualWan` -- Virtual WAN: the umbrella of Azure's managed hub-and-spoke networking, under which virtual hubs and their gateways are created. Self-contained -- only the resource group is required. A container: its hubs and its branch sites are created into it, so on a diagram the WAN is the room they stand in.
- `AzureVirtualHub` -- The WAN is the prerequisite: this kind models the Virtual WAN hub (virtual_wan_id is required; standalone hubs are the legacy Route Server construction, which has its own ARM surface). The resource group chains transitively through the WAN. A container: ARM deploys the hub's VPN, ExpressRoute, and point-to-site gateways and its spoke connections INTO the hub, so on a diagram the hub is the room they stand in.
- `AzureVirtualHubConnection` -- Both sides of the attachment are prerequisites: the hub being joined and the spoke virtual network being attached.
- `AzureVpnGateway` -- The hub is the prerequisite: ARM deploys a Virtual WAN VPN gateway INTO a virtual hub (virtual_hub_id is required and immutable; the WAN and resource group chain transitively through the hub). ARM allows one VPN gateway per hub.
- `AzureVpnGatewayConnection` -- Both ends of the tunnel are prerequisites: a connection is an ARM child of the VPN gateway and pins each of its links to a specific link of the remote VPN site (the hub, WAN, and resource group chain transitively through the gateway).
- `AzureVpnSite` -- The WAN is the prerequisite: a VPN site is the Virtual WAN world's address-book entry for one branch location (virtual_wan_id is required; the classic-world sibling without a WAN is AzureLocalNetworkGateway). The resource group chains transitively through the WAN.
- `AzurePointToSiteVpnGateway` -- The hub and the server configuration are both prerequisites: a point-to-site VPN gateway deploys INTO a virtual hub (one P2S gateway per hub, a slot separate from the hub's site-to-site VPN gateway) and is born pointing at the VPN server configuration that defines how its users authenticate -- both ARM-required and fixed at creation. The WAN and resource group chain transitively through the hub.
- `AzureVpnServerConfiguration` -- Self-contained -- only the resource group is required: a VPN server configuration is the reusable "who may connect and how" authentication policy (Entra ID / certificate / RADIUS) that point-to-site VPN gateways attach to; it references no other Azure resource.
- `AzureCognitiveAccount` -- Self-contained -- only the resource group is required: an Azure AI services account (Azure OpenAI, the multi-service AIServices account, the single-service accounts) needs no other Azure resource; subnets (network rules), Key Vault keys (CMK), storage accounts and user-assigned identities are optional references. A container kind: every model deployment and every project is an ARM child created onto the account (accounts/{account}/deployments/{name}, .../projects/{name}), so a diagram draws them inside it. The subnets the account admits through its network ACLs or injects agent compute into are access, never placement (containment_exempt on those fields).
- `AzureCognitiveDeployment` -- An ARM child of its account: a model deployment (which model runs, at which throughput class) exists only on an Azure AI services account of kind "OpenAI" or "AIServices".
- `AzureCognitiveAccountProject` -- An ARM child of its account: an AI Foundry project exists only on an "AIServices"-kind account with project management enabled.
- `AzureMachineLearningWorkspace` -- The workspace REQUIRES all three companion services at creation (default storage, secrets vault, telemetry) -- genuine deploy-order prerequisites, each with its own fixture profile. A container kind: datastores, compute clusters and instances, and online and batch endpoints are ARM children of the workspace (workspaces/{ws}/datastores|computes|onlineEndpoints|batchEndpoints/{name}), so a diagram draws them inside it. The companion services it names are reaches (containment_exempt on those fields), and so is the subnet its serverless compute or a compute's nodes attach to.
- `AzureMachineLearningDatastore` -- An ARM child of its workspace. The storage target (container, filesystem or share) is scenario-declared via the e2e-prerequisites annotation -- only the blob scenario needs a container, so it is not a kind-wide prerequisite.
- `AzureMachineLearningComputeCluster` -- An ARM child of its workspace (.../computes/{name}) -- the auto-scaling pool of VMs training jobs run on.
- `AzureMachineLearningComputeInstance` -- An ARM child of its workspace (.../computes/{name}) -- a single always-on VM serving as one data scientist's cloud workstation.
- `AzureAiFoundry` -- The hub REQUIRES both companion services at creation (secrets vault, default storage) -- genuine deploy-order prerequisites, each with its own fixture profile. A container kind: every AI Foundry project is created inside its hub and deploys into the hub's resource group (the project spec carries none), so a diagram draws the projects inside the hub. The vault and storage account the hub names are reaches (containment_exempt on those fields).
- `AzureAiFoundryProject` -- Deploys into its hub's resource group (the provider derives the group from the hub reference -- the project spec carries none).
- `AzureSearchService`
- `AzureMachineLearningOnlineEndpoint` -- An ARM child of its workspace (.../onlineEndpoints/{name}) -- the stable scoring address applications call. azurerm carries no ML endpoint resources; the modules write the raw ARM shape at a pinned api-version (azapi / azure-native). A container kind in its turn: every online deployment is an ARM child created under the endpoint whose traffic map routes to it (.../onlineEndpoints/{endpoint}/deployments/{name}) and cannot exist without it, so a diagram draws the deployments inside the endpoint, inside the workspace.
- `AzureMachineLearningOnlineDeployment` -- An ARM child of its endpoint (.../deployments/{name}) -- the running copy of a model the endpoint's traffic map routes to.
- `AzureMachineLearningBatchEndpoint` -- An ARM child of its workspace (.../batchEndpoints/{name}) -- the stable address batch scoring jobs are submitted to. azurerm carries no ML endpoint resources; the modules write the raw ARM shape at a pinned api-version (azapi / azure-native). A container kind in its turn: every batch deployment is an ARM child created under the endpoint whose default-deployment pointer routes to it (.../batchEndpoints/{endpoint}/deployments/{name}), so a diagram draws the deployments inside the endpoint, inside the workspace. The compute cluster a deployment runs on is a reach, not a room.
- `AzureMachineLearningBatchDeployment` -- An ARM child of its endpoint (.../deployments/{name}) -- the job recipe (model, compute, batching behavior) the endpoint's default-deployment pointer routes submissions to.
- `AzureRecoveryServicesVault` -- The Recovery Services vault (Microsoft.RecoveryServices/vaults) -- the safe that classic Azure Backup data and Site Recovery configuration live in. Backup policies and protected items are ARM children of a vault, so the vault is a container kind: a diagram draws them inside it, as the portal's own blades do.
- `AzureBackupPolicyVm` -- An ARM child of its vault (.../backupPolicies/{name}) -- the schedule and retention rules that govern IaaS VM backups.
- `AzureBackupProtectedVm` -- An ARM child of its vault (.../protectedItems/...) -- the binding that puts one virtual machine under a backup policy's protection.
- `AzureBackupPolicyFileShare` -- An ARM child of its vault (.../backupPolicies/{name}) -- the schedule and retention rules that govern Azure Files share backups (snapshot or vaulted).
- `AzureBackupProtectedFileShare` -- An ARM child of its vault (.../protectedItems/AzureFileShare;...) -- the binding that puts one Azure Files share under a backup policy's protection. The share's storage account must already be registered with the vault (AzureBackupContainerStorageAccount). Prerequisite ORDER is load-bearing for teardown (destroy runs in reverse): the registration must list AFTER the share so it unregisters FIRST -- Azure Backup holds a DoNotDelete lock on a registered storage account, and a share delete under that lock fails ScopeLocked.
- `AzureDataProtectionBackupVault` -- The Data Protection backup vault (Microsoft.DataProtection/ backupVaults) -- the safe that MODERN Azure Backup data lives in (managed disks, blob storage, AKS clusters, MySQL/PostgreSQL flexible servers, Data Lake storage). Backup policies and backup instances are ARM children of a vault, so the vault is a container kind: a diagram draws them inside it, as the portal's own blades do.
- `AzureDataProtectionBackupPolicy` -- An ARM child of its vault (.../backupPolicies/{name}) -- the schedule and retention rules for ONE Data Protection datasource type (blob storage, disk, Kubernetes cluster, MySQL/PostgreSQL flexible server, or Data Lake storage), modeled as one kind with variant blocks.
- `AzureDataProtectionBackupInstance` -- An ARM child of its vault (.../backupInstances/{name}) -- the binding that puts ONE datasource (a managed disk, a storage account's blob services, an AKS cluster, a MySQL/PostgreSQL flexible server, or a Data Lake storage account) under a Data Protection backup policy, modeled as one kind with variant blocks. The vault's managed identity must hold the datasource roles Azure Backup requires BEFORE the instance is created.
- `AzureBastionHost` -- AzureSubnet and AzurePublicIp are prerequisites because a dedicated-infrastructure Bastion host (Basic/Standard/Premium -- the default shapes) deploys into a subnet named exactly "AzureBastionSubnet" and binds a Standard static public IP EXCLUSIVELY (the virtual network and resource group chain transitively through the subnet). The Developer SKU instead attaches to a virtual network directly and uses neither.
- `AzureNetworkWatcherFlowLog` -- AzureVirtualNetwork and AzureStorageAccount are prerequisites because a flow log records a network-scoped target (a virtual network in the common case; subnets and network interfaces chain through the network) into a referenced storage account. The regional Network Watcher parent is NOT a prerequisite: Azure auto-creates it ("NetworkWatcher_{region}" in "NetworkWatcherRG") the moment the region hosts a virtual network, and the flow log references it by name. Traffic Analytics' Log Analytics workspace is an optional arm, declared by scenarios that use it.
- `AzurePrivateDnsResolver` -- AzureVirtualNetwork and AzureSubnet are prerequisites because a DNS Private Resolver anchors to a referenced virtual network (at most ONE resolver per network -- Azure enforces it) and each of its inbound/outbound endpoints occupies its own dedicated subnet delegated to "Microsoft.Network/dnsResolvers" (the resource group chains transitively through the network and subnets).
- `AzurePrivateDnsResolverForwardingRuleset` -- AzurePrivateDnsResolver is a prerequisite because a DNS forwarding ruleset steers a resolver's OUTBOUND endpoints -- it binds their ARM ids (at most 2, same resolver) at creation. (The resource group and network chain transitively through the resolver's own prerequisite declarations.)
- `AzurePrivateDnsRecord` -- AzurePrivateDnsZone is a prerequisite because every record set is created inside a referenced private DNS zone (the resource group chains transitively through the zone's own prerequisite).
- `AzureTrafficManagerProfile` -- AzureResourceGroup is a prerequisite because a Traffic Manager profile is created inside a referenced resource group (the profile itself is a global service -- the group only holds its metadata record). A container kind: every endpoint is an ARM child of its profile, so on a diagram the endpoints stand inside the profile that steers traffic to them. A nested endpoint that points AT another profile lives in its own parent profile, never in the one it targets, and that reference says so with containment_exempt.
- `AzureTrafficManagerEndpoint` -- AzureTrafficManagerProfile is a prerequisite because every endpoint is created inside a referenced profile -- it is the destination a profile steers traffic to (the resource group chains transitively through the profile's own prerequisite).
- `AzureMonitorAutoscaleSetting` -- AzureResourceGroup is a prerequisite because an autoscale setting is created inside a referenced resource group. The scalable TARGET it controls is a no-default reference (many kinds can be scaled), so no target kind is declared here -- scenarios declare their own target fixture.
- `AzureMonitorDataCollectionRule` -- The Azure Monitor data collection rule (DCR) -- the routing table declaring what telemetry the Azure Monitor Agent collects and where it lands. AzureResourceGroup is a prerequisite because a rule is created inside a referenced resource group; AzureLogAnalyticsWorkspace because a workspace is the canonical destination a rule routes to (the smoke scenario's shape). Machines attach to a rule with AzureMonitorDataCollectionRuleAssociation resources.
- `AzureEventgridTopic` -- The Azure Event Grid custom topic -- the HTTPS endpoint an application publishes its own events to, fanned out to handlers by event subscriptions. One topic is one event stream with its own endpoint and access keys; for many streams behind one endpoint see AzureEventgridDomain. A container kind: every event subscription on the topic is an ARM child created under it ({topic_id}/providers/ Microsoft.EventGrid/eventSubscriptions/{name}) and cannot exist without it, so on a diagram the topic is the room its subscriptions stand in, as the portal's own blade lists them. A kind that merely publishes to or subscribes from a topic while living elsewhere (a Data Factory trigger, an Event Grid namespace's MQTT route) says so with containment_exempt.
- `AzureEventgridDomain` -- The Azure Event Grid domain -- ONE publishing endpoint and one pair of access keys serving many event streams (domain topics), the multi-tenant pattern. Topics inside the domain are auto-managed by Azure or declared explicitly as AzureEventgridDomainTopic resources. A container kind: every domain topic is an ARM child of its domain ({domain_id}/topics/{name}), so on a diagram the domain is the room its tenant streams stand in.
- `AzureEventgridSystemTopic` -- The Azure Event Grid system topic -- the subscription surface for events AZURE ITSELF publishes about one of your resources (a storage account's blob events, a resource group's lifecycle events). One system topic per source resource per topic type; event subscriptions attach to it to route events to handlers. A container kind: every subscription on a system topic is an ARM child created under it ({system_topic_id}/eventSubscriptions/{name}), so on a diagram the system topic is the room its subscriptions stand in.
- `AzureEventgridEventSubscription` -- The Azure Event Grid event subscription -- the delivery instruction routing events from a source (a custom topic, domain, domain topic, system topic, resource group, or subscription) to a handler (a Function, Event Hub, Service Bus queue/topic, storage queue, hybrid connection, or webhook), with filtering, retry, and dead-letter behavior.
- `AzureEventgridNamespace` -- The Azure Event Grid namespace -- the capacity-scaled hub of the newer Event Grid: hosts CloudEvents namespace topics and an optional MQTT broker behind one set of regional endpoints, sized in throughput units. A container kind: every namespace topic is an ARM child of its namespace ({namespace_id}/topics/{name}), so on a diagram the namespace is the room its streams stand in.
- `AzureDataFactory` -- The Azure Data Factory -- the workspace every other Data Factory resource lives inside: pipelines, data flows, linked services, datasets, triggers, and integration runtimes are all created against a factory's ARM ID. A container kind for exactly that reason: those six kinds are ARM children of the factory, so on a diagram the factory is the room they stand in.
- `AzureDataFactoryPipeline` -- One unit of work inside an Azure Data Factory ({factory_id}/pipelines/{name}) -- an ordered set of activities that executes as a whole when triggered.
- `AzureDataFactoryDataFlow` -- A Data Factory data flow ({factory_id}/dataflows/{name}) -- a visually-designed data transformation executed on managed Spark, or, as a flowlet, a reusable snippet other data flows embed. One kind covers both provider forms (they share one schema and one name namespace inside the factory).
- `AzureDataFactoryLinkedService` -- A Data Factory linked service ({factory_id}/linkedservices/{name}) -- a saved connection in the factory's address book: where an external system lives and how to authenticate to it. One kind covers every connection type Azure models as a first-class resource (storage, SQL family, Cosmos DB, Databricks, Key Vault, SFTP, web APIs, and more) as variants in one factory-scoped name namespace, plus the raw-JSON custom form.
- `AzureDataFactoryDataset` -- A Data Factory dataset ({factory_id}/datasets/{name}) -- a named view of data inside a system a linked service already connects to: which container and path, which table, which file format. One kind covers every data shape Azure models as a first-class dataset resource (delimited text/CSV, JSON, Parquet, binary, blob, HTTP, the SQL family, Snowflake, Cosmos DB) as variants in one factory-scoped name namespace, plus the raw-JSON custom form.
- `AzureDataFactoryTrigger` -- A Data Factory trigger ({factory_id}/triggers/{name}) -- the instruction that starts pipelines automatically: on a clock schedule, per contiguous tumbling window, on storage blob events, or on Event Grid custom events. One kind covers all four provider trigger resources as variants (one ARM namespace, one started/stopped lifecycle).
- `AzureDataFactoryIntegrationRuntime` -- A Data Factory integration runtime ({factory_id}/integrationRuntimes/{name}) -- the compute engine a factory's pipelines, data flows, and copy activities run on. One kind covers all three engine flavors as variants in one factory-scoped name namespace: the managed data-flow compute, the managed SSIS package runtime, and the self-hosted agent registration (which issues the authorization keys agents join with).
- `AzureComputeGallery` -- The Azure Compute Gallery -- the shared library an organization keeps its approved VM images in. Image definitions (AzureComputeGalleryImage) live inside it; VMs and scale sets deploy from their published, region-replicated versions. A container kind: every image definition is an ARM child of its gallery ({gallery_id}/images/{name}), so on a diagram the definitions stand inside the library that publishes them. The image ids a VM, a scale set, or a disk boots from are plain strings and never place anything inside the gallery.
- `AzureComputeGalleryImage` -- A gallery image ({gallery_id}/images/{name}) -- one image definition inside a Compute Gallery (marketplace-style identity, OS type, security posture) plus its published versions, each replicated to its own target regions. VMs deploy from a version's ARM ID or from the definition's ID to get the latest version.
- `AzureAvailabilitySet` -- The availability set -- the classic pre-zones placement grouping that spreads VMs across separate fault and update domains so one hardware failure or maintenance window cannot take them all down. VMs join the set at creation.
- `AzureDiskSnapshot` -- The managed disk snapshot -- a point-in-time copy of a disk used for backup, cloning, and as the source of gallery image versions. Incremental snapshots store only the delta since the previous snapshot of the same disk.
- `AzureContainerInstance` -- The Azure Container Instance container group -- serverless containers billed per second: one or more containers sharing a lifecycle, network, and volumes (plus one-shot init containers), with no cluster or VM to manage. Public, subnet-private, or IP-less postures.
- `AzureFunctionAppFlexConsumption` -- The Azure Function App on the Flex Consumption plan -- Azure's newest serverless Functions hosting model: per-instance memory selection, a configurable scale-out ceiling, always-ready instance pools, and explicit blob-container deployment storage. Requires an FC1-SKU service plan, which is deliberately NOT a registry prerequisite: the shared plan fixture serves the classic app tiers, and an FC1 plan is cheap to create per scenario (no idle compute cost), so scenarios bring their own plan fixture -- the same reasoning that keeps the globally-unique storage account scenario-local for AzureFunctionApp.
- `AzureMongoCluster` -- The Azure Cosmos DB for MongoDB vCore cluster -- Azure's modern managed MongoDB: a real MongoDB engine on dedicated vCore tiers with sharding, zone-redundant HA, and point-in-time restore.
- `AzureFabricCapacity` -- The Microsoft Fabric capacity -- the billing and compute anchor of Microsoft Fabric: workspaces assign themselves to a capacity, and its F-SKU sets how much compute every workload on it shares. azurerm's entire Fabric surface is this one resource (workspaces and items live in Microsoft's dedicated fabric provider).
- `AzureBackupContainerStorageAccount` -- Registers a storage account with a Recovery Services vault as a backup container (.../protectionContainers/StorageContainer;...) -- one registration per storage-account-and-vault pair, required BEFORE any of the account's file shares can be protected. Part of the backup family (2175-2179) despite the out-of-run number -- enum numbers are pinned by the registry snapshot; never renumber.
- `AzureDataProtectionResourceGuard` -- The Data Protection Resource Guard (Microsoft.DataProtection/ resourceGuards) -- the approval gate behind Multi-User Authorization: privileged vault operations (disabling soft delete, reducing retention) require an approval through a guard, which typically lives in a DIFFERENT administrator's scope. Vaults reference a guard by its ARM ID. Part of the backup family (2175-2182) despite the out-of-run number -- enum numbers are pinned by the registry snapshot; never renumber.
- `AzurePrivateDnsResolverVirtualNetworkLink` -- The attachment that makes a DNS forwarding ruleset take effect in one virtual network ({ruleset_id}/virtualNetworkLinks/{name}) -- one link per ruleset-network pair, up to 500 per ruleset, spokes joining and leaving independently (which is why the link is a standalone kind, exactly like AzurePrivateDnsZoneVirtualNetworkLink). Part of the DNS Private Resolver family (2186-2187) despite the out-of-run number -- enum numbers are pinned by the registry snapshot; never renumber.
- `AzureMonitorDataCollectionRuleAssociation` -- The attachment that puts ONE machine under an Azure Monitor data collection rule ({target_id}/providers/Microsoft.Insights/ dataCollectionRuleAssociations/{name}) -- an extension resource on the TARGET machine, many per rule, machines joining and leaving monitoring independently (which is why the association is a standalone kind, exactly like AzurePrivateDnsZoneVirtualNetworkLink). AzureVirtualMachine is a prerequisite because the smoke scenario attaches the fixture VM; the rule prerequisite chains the rule's own install manifest. Part of the Monitor family (2191-2192) despite the out-of-run number -- enum numbers are pinned by the registry snapshot; never renumber.
- `AzureEventgridDomainTopic` -- One named event stream inside an Azure Event Grid domain ({domain_id}/topics/{name}) -- the per-tenant mailbox of the multi-tenant pattern: many per domain, each with its own subscriptions and lifecycle, tenants joining and leaving without touching the domain (which is why the domain topic is a standalone kind, exactly like AzureEventHubConsumerGroup on a shared hub). Part of the Event Grid family (2193-2194) despite the out-of-run number -- enum numbers are pinned by the registry snapshot; never renumber.
- `AzureEventgridNamespaceTopic` -- One named CloudEvents stream inside an Azure Event Grid namespace ({namespace_id}/topics/{name}) -- many per namespace, publishers and teams creating and deleting their own against the shared namespace (which is why the topic is a standalone kind, exactly like AzureEventgridDomainTopic and AzureEventHubConsumerGroup). Part of the Event Grid family (2193-2197) despite the out-of-run number -- enum numbers are pinned by the registry snapshot; never renumber.
- `AzureMongoClusterUser` -- Grants one Microsoft Entra principal access to an Azure Cosmos DB for MongoDB vCore cluster ({cluster_id}/users/{object_id}) -- an access binding, not a password user: many per cluster, principals joining and leaving independently (which is why the grant is a standalone kind, the access-grant class of AzureRoleAssignment). Part of the Mongo vCore family (2211) despite the out-of-run number -- enum numbers are pinned by the registry snapshot; never renumber.
- `AzurePlantonRunner` -- AzureContainerAppEnvironment is a prerequisite because the runner appliance is a Container App, and every Container App runs inside an environment -- the environment reference must resolve before the appliance can deploy.
- `GcpArtifactRegistryRepo` -- 3000–3999: GCP resources
- `GcpTargetHttpsProxy` -- The URL map is the parent a proxy cannot exist without; the classic compute certificate kinds and the SSL policy are the fixture parents the committed scenarios attach. The Certificate Manager certificate list (certificate_manager_certificates, honored only by the cross-region internal ALB) is optional composition -- a scenario that arms it declares GcpCertManagerCert via the e2e-prerequisites annotation, never a registry edge that would tax every proxy and forwarding-rule chain.
- `GcpCloudFunction`
- `GcpCloudRun`
- `GcpCloudSql`
- `GcpDnsZone`
- `GcpGcsBucket`
- `GcpGkeCluster`
- `GcpIamCustomRole`
- `GcpProject`
- `GcpVpcNetwork`
- `GcpSubnetwork`
- `GcpRouterNat`
- `GcpGkeNodePool`
- `GcpServiceAccount`
- `GcpGkeWorkloadIdentityBinding`
- `GcpCertManagerCert`
- `GcpComputeInstance`
- `GcpDnsRecord`
- `GcpProjectIamMember`
- `GcpFirewallRule`
- `GcpGlobalAddress`
- `GcpCloudArmorPolicy`
- `GcpHealthCheck`
- `GcpBackendBucket`
- `GcpBackendService`
- `GcpRegionNetworkEndpointGroup`
- `GcpUrlMap`
- `GcpManagedSslCertificate`
- `GcpTargetHttpProxy`
- `GcpAlloydbCluster`
- `GcpRedisInstance`
- `GcpFirestoreDatabase`
- `GcpSpannerInstance`
- `GcpSpannerDatabase`
- `GcpBigtableInstance`
- `GcpMemorystoreInstance`
- `GcpCloudSqlDatabase`
- `GcpCloudSqlUser`
- `GcpAlloydbInstance`
- `GcpAlloydbUser`
- `GcpSpannerBackupSchedule`
- `GcpBigtableTable`
- `GcpFirestoreBackupSchedule`
- `GcpFirestoreIndex`
- `GcpBigQueryDataset`
- `GcpDataprocCluster`
- `GcpDataprocAutoscalingPolicy`
- `GcpBigQueryTable`
- `GcpBigQueryCapacityCommitment` -- A BigQuery capacity commitment: slots bought for a fixed term in an administration project and location, pooled across every reservation there. A purchase Google will not delete before its term ends.
- `GcpBigQueryReservationGroup` -- A BigQuery reservation group: reservations that share idle slots with each other first. Reservations reference it.
- `GcpDatastreamConnectionProfile` -- A Datastream connection profile: where one source database or destination is and how Datastream signs in. Streams reference a source and a destination profile; one profile serves many streams.
- `GcpDatastreamPrivateConnection` -- A Datastream private connection: the VPC peering or Private Service Connect interface through which Datastream reaches private databases, shared by every profile that reaches that network.
- `GcpPubSubTopic`
- `GcpPubSubSubscription`
- `GcpCloudTasksQueue`
- `GcpCloudSchedulerJob`
- `GcpPubSubSchema`
- `GcpPubSubTopicIamMember` -- One additive grant on a topic. Its own kind, not a field on the topic, because the identities that most need it (a logging sink's writer, a Security Command Center export's publisher) belong to resources that name the topic themselves; a grant on the topic that referenced them back would be a dependency cycle.
- `GcpVertexAiNotebook`
- `GcpVertexAiEndpoint`
- `GcpVertexAiIndex`
- `GcpVertexAiIndexEndpoint` -- Vector Search IndexEndpoint — distinct from the online-prediction GcpVertexAiEndpoint (671); different GCP resources, different kinds.
- `GcpVertexAiDeployedIndex`
- `GcpCloudComposerEnvironment`
- `GcpCloudComposerUserWorkloadsSecret`
- `GcpCloudComposerUserWorkloadsConfigMap`
- `GcpKmsKeyRing`
- `GcpKmsKey`
- `GcpKmsKeyIamMember`
- `GcpFilestoreInstance`
- `GcpWorkloadIdentityPool` -- 3101–3109: IAM/identity family (overflow block; the 3000–3022 foundation/security sub-band is fully allocated)
- `GcpWorkloadIdentityPoolProvider`
- `GcpServiceAccountIamMember`
- `GcpGcsBucketIamMember` -- One additive grant on a bucket, for a grantee that depends on the bucket itself (a logging sink writing into it): the bucket's own iam_members cannot reference such an identity without a cycle.
- `GcpGlobalForwardingRule` -- 3110–3119: networking/load-balancer family (overflow block; the 3023–3029 LB sub-band is fully allocated)
- `GcpSslPolicy`
- `GcpSslCertificate`
- `GcpServiceNetworkingConnection`
- `GcpAddress`
- `GcpServiceConnectionPolicy`
- `GcpCertManagerDnsAuthorization`
- `GcpCertificateMap` -- GcpCertManagerCert is a prerequisite because a map entry binds hostnames to EXISTING certificates — the canonical map references a certificate fixture's resource name.
- `GcpCertManagerTrustConfig` -- The CA certificates a load balancer validates client certificates against (mutual TLS). Its own kind: TLS policies reference it, never a certificate.
- `GcpCertManagerIssuanceConfig` -- How Google-managed certificates are issued from a private CA pool. Its own kind: many certificates share one config by name.
- `GcpCloudRunJob` -- 3120–3129: GCP serverless overflow
- `GcpServerlessVpcConnector`
- `GcpCloudRunWorkerPool` -- Cloud Run's no-ingress shape: a pool of always-running container instances (queue consumers, schedulers, background workers) that scales manually or by the owner's own signal instead of by requests. The proof deploys direct-VPC egress onto the prerequisite network.
- `GcpComputeDisk` -- 3130–3139: GCP compute overflow (the 3000–3022 foundation sub-band that holds GcpComputeInstance is fully allocated)
- `GcpComputeMig` -- GcpVpcNetwork is a prerequisite because the canonical group runs its fleet on a dedicated custom-mode VPC — a managed instance group's template must attach every VM to a network, and the default VPC is never assumed.
- `GcpMonitoringNotificationChannel` -- 3140–3149: GCP observability & log routing
- `GcpMonitoringAlertPolicy` -- GcpMonitoringNotificationChannel is a prerequisite because the policy's canonical shape references a channel to notify — a policy without a delivery endpoint measures but never pages.
- `GcpMonitoringUptimeCheck`
- `GcpLoggingSink` -- GcpGcsBucket is a prerequisite because the canonical sink exports to a Cloud Storage bucket — the cheapest destination that proves the whole writer-identity grant flow.
- `GcpMonitoringDashboard`
- `GcpMonitoringSlo`
- `GcpLogBucket`
- `GcpLogMetric`
- `GcpSecretManagerSecret` -- 3150–3159: GCP security & identity GcpServiceAccount is a prerequisite because the canonical secret grants secretAccessor to a workload service account — the access story the kind exists to model.
- `GcpIdentityPlatformConfig`
- `GcpIdentityPlatformTenant` -- GcpIdentityPlatformConfig is a prerequisite because tenants exist only in projects whose Identity Platform config enables multi_tenant.allow_tenants — a tenant without the initialized, tenant-enabled project config cannot be created at all.
- `GcpIamOauthClient`
- `GcpIamDenyPolicy`
- `GcpCloudRunDomainMapping` -- 3160–3169: GCP serverless edge GcpCloudRun is a prerequisite because a domain mapping exists only to point a verified domain at a running Cloud Run service — the route it maps must already exist for the mapping to be created at all.
- `GcpWorkflow`
- `GcpEventarcTrigger` -- GcpCloudRun is a prerequisite because the canonical trigger routes a Pub/Sub messagePublished event to a Cloud Run service — the destination story the kind exists to model.
- `GcpEventarcMessageBus`
- `GcpPlantonRunner`
- `GcpFolder` -- 3170–3179: GCP organization & governance (folders, org policies, tags, budgets, identity groups, API keys, KMS Autokey) GcpFolder is a container: the hierarchy node projects, sub-folders, policies, and tag bindings are placed inside.
- `GcpOrgPolicy`
- `GcpTagKey`
- `GcpTagValue`
- `GcpTagBinding`
- `GcpBillingBudget` -- A spending guardrail on a Cloud Billing account: amount, period, filters, thresholds, and where the alerts go. Lives on the billing account, so no project prerequisite; the proof lane needs a billing account the harness identity can administer.
- `GcpCloudIdentityGroup` -- A Google Group in Cloud Identity or Workspace with its memberships folded in -- the unit IAM bindings should name. Lives under a Cloud Identity customer, beside the service accounts and workload identity pools in the identity service group.
- `GcpApiKey`
- `GcpKmsAutokeyConfig` -- Cloud KMS Autokey switched on for a folder or a project: where the customer-managed keys GcpKmsKeyHandle requests are created. A security control, so it groups with the KMS kinds rather than the hierarchy.
- `GcpOrgPolicyCustomConstraint` -- A custom constraint is a DEFINITION the organization owns; the GcpOrgPolicy kinds that enforce it reference it by name, the way IAM bindings reference a custom role.
- `GcpSharedVpcHost` -- 3180–3189: GCP networking fabric (Shared VPC, VPC peering, HA VPN, firewall policies, PSC, network endpoint groups). HA VPN is two kinds: the gateway (with its Cloud Router) is declared once per VPC and region and referenced by every site connection, so two Google Cloud VPCs can point their connections at each other's gateway without a dependency cycle. Firewall policies are two kinds by scope: the hierarchical policy lives on the organization or a folder and is inherited by every network beneath its association; the network policy lives in a project and is attached to that project's VPC networks (globally or per region). Each folds its rules and associations -- a rule is keyed by priority inside its policy and an association is the edge that makes the policy act.
- `GcpSharedVpcServiceProject`
- `GcpVpcPeering`
- `GcpHaVpnGateway`
- `GcpHierarchicalFirewallPolicy`
- `GcpNetworkFirewallPolicy`
- `GcpHaVpnConnection`
- `GcpPscServiceAttachment` -- The producer half of Private Service Connect: publishes an internal load balancer's regional forwarding rule through NAT subnets so consumers in other VPCs reach it over a PSC endpoint (a regional GcpGlobalForwardingRule with an empty scheme targeting this attachment). The proof chain deploys the internal passthrough load balancer and the PSC NAT subnet as fixtures on the prerequisite network.
- `GcpNetworkEndpointGroup` -- Zonal (VM, hybrid, internet) and global (internet) network endpoint groups behind a `zone` selector; serverless, PSC, and regional internet groups stay in GcpRegionNetworkEndpointGroup. The zonal proof attaches a GcpComputeInstance fixture as an endpoint on the prerequisite network.
- `GcpRedisCluster` -- 3190–3199: GCP data (Memorystore for Redis Cluster, Managed Kafka, BigQuery connections and reservations, Datastream) Memorystore for Redis Cluster: the sharded, horizontally scaled Redis. Connectivity is Private Service Connect, placed by service connectivity automation through a GcpServiceConnectionPolicy for the gcp-memorystore-redis class on the network in the cluster's region -- the prerequisite; its proof pin carries that class beside the network and subnet pins.
- `GcpManagedKafkaCluster` -- Managed Service for Apache Kafka: a Google-operated broker fleet in one region, reachable from the VPC subnets it is attached to -- the prerequisite. Topics, ACLs, and Kafka Connect are their own kinds so the teams that own them declare them without editing the cluster.
- `GcpManagedKafkaTopic`
- `GcpManagedKafkaConnectCluster` -- Kafka Connect workers attached to a Kafka cluster: a separate project-and-location root that names the cluster it serves.
- `GcpBigQueryConnection` -- BigQuery's link to data outside its own storage (Cloud SQL, Spanner, AWS and Azure through Omni, Google resources through a managed service account, the Connector framework, Spark procedures).
- `GcpBigQueryReservation` -- A BigQuery slot reservation with the assignments that route projects, folders, or an organization onto it.
- `GcpDatastreamStream` -- A Datastream stream: continuous change data capture from one source database into BigQuery or Cloud Storage, through a source and a destination connection profile.
- `GcpRedisClusterEndpointSet` -- The Private Service Connect connections a consumer builds by hand (forwarding rules in other VPCs or projects) registered on a Redis Cluster, as one set: Google's resource replaces the cluster's whole user-created endpoint list in one write, so exactly one set per cluster is the honest grain. Its own kind because every connection names a forwarding rule that targets one of the cluster's service attachments -- a fold would depend on its own output.
- `GcpManagedKafkaAcl` -- The access rules for one resource pattern (a topic, a consumer group, a prefix, the cluster) on a Kafka cluster.
- `GcpManagedKafkaConnector` -- One data pipeline running on a Kafka Connect cluster.
- `GcpVertexAiAgentEngine` -- 3200–3229: GCP AI (Vertex AI agents and model deployments, RAG Engine, Vector Search, and the rest of the Vertex AI and generative-AI building blocks). The pre-existing Vertex AI kinds -- endpoint, index, index endpoint, deployed index, notebook -- live in the 3070s. Vertex AI Agent Engine: the managed runtime an AI agent runs in -- built from source or a container, hosted with its own identity and autoscaling, with an optional Memory Bank of long-term memories.
- `GcpVertexAiModelGardenDeployment` -- A Model Garden or Hugging Face model deployed to a Vertex AI endpoint in one step; every argument is immutable, so a change redeploys.
- `GcpVertexAiRagEngineConfig` -- The per-project, per-location tier of Vertex AI RAG Engine's managed vector database: a singleton Google owns, updated in place.
- `GcpVectorSearchCollection` -- 3203 is reserved for GcpVertexAiRagCorpus, forged when pulumi-gcp bridges provider 8.x (the resource has no SDK type today). A Vector Search collection -- a schema'd store of data objects with vector fields -- together with the approximate-nearest-neighbor indexes built over those fields (folded: one collection owns them).
- `GcpVertexAiSearchDataStore` -- Vertex AI Search (the Discovery Engine API behind the console's AI Applications / Gemini Enterprise): a data store is the corpus -- structured records, unstructured documents, or a public website -- with its schema, crawl patterns, and sitemaps folded in.
- `GcpVertexAiSearchEngine` -- The app over one or more data stores -- a search, chat, or recommendation engine -- with its serving controls, serving config, search widget, and assistants folded in. Its data stores are prerequisites: an engine cannot exist without one.
- `GcpVertexAiFeatureGroup` -- Vertex AI Feature Store: a feature group registers the features of some entities that live in a BigQuery table or view -- the features themselves folded in. Online stores serve them through feature views.
- `GcpVertexAiFeatureOnlineStore` -- The low-latency serving layer of Vertex AI Feature Store (Bigtable or Optimized storage), with the feature views it serves folded in.
- `GcpVertexAiDataset` -- A Vertex AI managed dataset: the registered container training, AutoML, labeling, and evaluation read their examples from.
- `GcpVertexAiTensorboard` -- A managed Vertex AI TensorBoard training jobs stream metrics into, with the experiments and runs declared in it folded in.
- `GcpVertexAiPersistentResource` -- A long-running cluster Vertex AI keeps provisioned so training jobs and Ray on Vertex AI start in seconds and scarce accelerators stay held between jobs.
- `GcpModelArmorTemplate` -- A Model Armor template: the named safety filters (prompt injection and jailbreak, Responsible AI content, sensitive data, malicious URLs) an AI application screens prompts and responses through.
- `GcpDocumentAiProcessor` -- A Document AI processor: a managed model that turns documents into structured data (OCR, forms, invoices, IDs), with its default version.
- `GcpColabRuntimeTemplate` -- A Colab Enterprise runtime template: the machine, network, image, and security settings every notebook runtime created from it gets.
- `GcpColabRuntime` -- A Colab Enterprise runtime: a notebook VM assigned to one user, built from a runtime template and started or stopped on purpose.
- `GcpColabSchedule` -- A Vertex AI schedule: a cron that launches a Colab Enterprise notebook run or a Vertex AI Pipelines run.
- `GcpTpuVm` -- A Cloud TPU VM: a slice of Google's AI accelerators with its host VMs. Beta-only in Google's provider (a recorded google-beta admission).
- `GcpDialogflowCxAgent` -- A Dialogflow CX conversational agent with the infrastructure its console-authored content uses folded in: webhooks, tools and their frozen versions, flow versions, the environments that pin them, and generative settings per language.
- `GcpVertexAiSearchDataConnector` -- A data connector is a COLLECTION of data stores Google syncs from a source (Jira, Confluence, ServiceNow, SharePoint, BigQuery, Google Drive, ...) on a schedule -- a different root from a data store, which is why it is its own kind. Engines search its stores by naming the collection.
- `GcpModelArmorFloorSetting` -- A Model Armor floor setting: the minimum safety screening a project, folder, or organization enforces on its templates and directly on Vertex AI and Google MCP server traffic. A different parent from a template (and a singleton Google never deletes), which is why it is its own kind.
- `GcpTpuQueuedResource` -- A Cloud TPU queued resource: a request that waits for TPU capacity and then provisions the nodes it describes. A different root from a TPU VM that owns many nodes, which is why it is its own kind. Beta-only in Google's provider (a recorded google-beta admission).
- `GcpDialogflowCxSecuritySettings` -- Dialogflow CX security settings: the redaction, retention, audio-export, and Insights-export policy agents in one project and location apply to their conversations. A different root from an agent, referenced by agents and shared among them, which is why it is its own kind.
- `GcpPrivateCaPool` -- 3230–3239: GCP security (Certificate Authority Service, Cloud KMS Autokey handles, Security Command Center, Binary Authorization) A Certificate Authority Service CA pool: the trust anchor and issuance policy its certificate authorities and certificates live inside.
- `GcpKmsKeyHandle` -- An Autokey key handle: asks Autokey for a customer-managed key for one resource type in one project and location; the resource it protects names the key the handle returns.
- `GcpSccNotificationConfig` -- Security Command Center streaming notifications to Pub/Sub at a project, folder, or organization.
- `GcpSccMuteConfig` -- A Security Command Center mute rule at a project, folder, or organization.
- `GcpSccBigQueryExport` -- A continuous Security Command Center findings export to a BigQuery dataset at a project, folder, or organization.
- `GcpBinaryAuthorizationPolicy` -- A project's Binary Authorization policy: which container images GKE admits, per cluster. A project singleton.
- `GcpBinaryAuthorizationAttestor` -- A Binary Authorization attestor with its Artifact Analysis note: the public keys that verify image attestations a policy requires.
- `GcpPrivateCaCertificateAuthority` -- A certificate authority in a CA pool: a self-signed root, or a subordinate signed by another authority or an outside CA. Its own kind because a pool rotates through several and a subordinate references its parent.
- `GcpPrivateCaCertificateTemplate` -- A certificate template: a reusable certificate shape in a project and location that certificates in any pool there reference.
- `GcpPrivateCaCertificate` -- A certificate issued from a CA pool for a key its owner holds; destroy revokes it.
- `GcpGkeFleet` -- 3240–3249: GCP platform engineering (GKE fleets, images, Cloud Build, Cloud Deploy) GcpGkeFleet is the container a fleet's scopes, namespaces, memberships, and features live in: the project's one fleet is the room, and a team scope or a fleet-wide feature is what is placed inside it.
- `GcpGkeFleetFeature` -- The three fleet children name GcpGkeFleet as their prerequisite: Google requires the fleet before a scope, a fleet declared after a cluster registers collides with the fleet that registration created implicitly, and a feature configures the fleet it lives in. A chart that references the fleet's project_id output orders each child after it.
- `GcpGkeFleetScope`
- `GcpGkeFleetMembership`
- `GcpComputeImage` -- A Compute Engine custom image: the golden boot image VMs, instance templates, and disks start from, rolled forward through image families.
- `GcpCloudBuildConnection` -- GcpCloudBuildConnection is the container a code host's repositories live in: the connection to GitHub, GitLab, or Bitbucket is the room, and each linked repository is what is placed inside it.
- `GcpCloudBuildRepository` -- A repository is created under its connection and cannot exist without it.
- `GcpCloudBuildTrigger` -- A Cloud Build trigger: what starts a build (a code event, a Pub/Sub message, a webhook, or a manual run) and what the build does.
- `GcpDeliveryPipeline` -- A Cloud Deploy delivery pipeline: the ordered stages a release is promoted through, with the automations that drive it.
- `GcpDeployTarget` -- A Cloud Deploy target: where a pipeline stage deploys (a GKE cluster, a Cloud Run location, a fleet cluster, several targets at once, or a custom target).
- `GcpFirebaseProject` -- 3250–3259: GCP Firebase (project enablement, app registrations, and the Firebase-adjacent products that follow) GcpFirebaseProject is the container the app registrations live in: "Firebase on this project" is the room, the Android/Apple/Web apps are what is placed inside it.
- `GcpFirebaseAndroidApp` -- The three app registrations exist only inside a Firebase-enabled project, so each names GcpFirebaseProject as its prerequisite: the E2E harness deploys the enablement first, and a chart that references the enablement's project_id output orders the registration after it.
- `GcpFirebaseAppleApp`
- `GcpFirebaseWebApp`
- `GcpCloudBuildWorkerPool` -- 3260–3269: GCP platform engineering, continued (Cloud Build and Cloud Deploy resources many pipelines share) A private Cloud Build worker pool: build machines many triggers and Cloud Deploy targets share, optionally on a private network.
- `GcpDeployPolicy` -- A Cloud Deploy deploy policy: rollout restrictions (freeze windows) that apply to every pipeline and target its selectors match.
- `GcpDeployCustomTargetType` -- A Cloud Deploy custom target type: how to render and deploy to a target Google does not deploy natively; many targets share one.
- `KubernetesNamespace` -- 4000–4999: Kubernetes resources, organized in family sub-bands (4030–4069 also hosts CNI/autoscaling/DR addons; 4130–4149 hosts analytics & ML; 4180–4199 hosts agent sandbox runtimes; 4200–4209 hosts AI agent platforms) 4000–4029: Kubernetes building blocks (core API primitives)
- `KubernetesDeployment`
- `KubernetesStatefulSet`
- `KubernetesDaemonSet`
- `KubernetesJob`
- `KubernetesCronJob`
- `KubernetesService`
- `KubernetesSecret`
- `KubernetesManifest`
- `KubernetesHelmRelease`
- `KubernetesConfigMap`
- `KubernetesServiceAccount`
- `KubernetesRbac` -- Bundles the RBAC grant grain (Role/ClusterRole + its binding) into one kind: "grant these permissions to these subjects in this scope".
- `KubernetesIngress`
- `KubernetesNetworkPolicy`
- `KubernetesPersistentVolumeClaim`
- `KubernetesStorageClass`
- `KubernetesResourceQuota` -- Manages the namespace-governance pair: the ResourceQuota plus an optional companion LimitRange (per-object defaults/bounds) — two API objects, one governance story.
- `KubernetesPriorityClass`
- `KubernetesPodDisruptionBudget`
- `KubernetesHorizontalPodAutoscaler`
- `KubernetesCertManager` -- 4030–4069: Kubernetes foundation addons (certs, DNS, secrets, ingress, Gateway API, mesh, CNI/autoscaling/DR)
- `KubernetesClusterIssuer` -- KubernetesCertManager is a prerequisite for the three cert-manager CR kinds below: ClusterIssuer/Issuer/Certificate are cert-manager custom resources — without the controller and its CRDs they cannot be applied.
- `KubernetesIssuer`
- `KubernetesCertificate`
- `KubernetesExternalDns`
- `KubernetesExternalSecretsOperator`
- `KubernetesClusterSecretStore` -- KubernetesExternalSecretsOperator is a prerequisite for the three external-secrets CR kinds below: ClusterSecretStore/SecretStore/ ExternalSecret are external-secrets custom resources — without the operator and its CRDs they cannot be applied.
- `KubernetesSecretStore`
- `KubernetesExternalSecret`
- `KubernetesIngressNginx`
- `KubernetesGatewayApiCrds`
- `KubernetesGatewayClass`
- `KubernetesGateway`
- `KubernetesListenerSet`
- `KubernetesHttpRoute`
- `KubernetesGrpcRoute`
- `KubernetesTcpRoute`
- `KubernetesUdpRoute`
- `KubernetesTlsRoute`
- `KubernetesReferenceGrant`
- `KubernetesBackendTlsPolicy`
- `KubernetesIstioBaseCrds`
- `KubernetesIstio`
- `KubernetesDestinationRule` -- Istio API kinds (mesh traffic policy, security, telemetry). The seven typed resources below (4053–4059) require the Istio CRDs on the cluster, provided by the lightweight CRDs-only KubernetesIstioBaseCrds (851) — NOT the full mesh KubernetesIstio (852).
- `KubernetesServiceEntry`
- `KubernetesPeerAuthentication`
- `KubernetesRequestAuthentication`
- `KubernetesAuthorizationPolicy`
- `KubernetesTelemetry`
- `KubernetesEnvoyFilter`
- `KubernetesMetricsServer`
- `KubernetesCilium`
- `KubernetesKeda`
- `KubernetesKarpenter`
- `KubernetesKarpenterNodePool`
- `KubernetesKarpenterEc2NodeClass`
- `KubernetesClusterAutoscaler`
- `KubernetesVelero`
- `KubernetesGkeComputeClass`
- `KubernetesKubePrometheusStack` -- 4070–4079: Kubernetes observability
- `KubernetesGrafana`
- `KubernetesSignoz` -- KubernetesClickHouse is a prerequisite because SigNoz stores every trace, metric and log in ClickHouse and deploys none of its own — the telemetry store is composed, never bundled.
- `KubernetesLoki`
- `KubernetesTempo`
- `KubernetesOtelOperator` -- The operator's admission webhooks (failurePolicy Fail) are served with a cert-manager Certificate in the default posture — cert-manager must be running before the operator installs.
- `KubernetesOtelCollector`
- `KubernetesServiceMonitor` -- KubernetesKubePrometheusStack is a prerequisite because it installs the monitoring.coreos.com CRDs the ServiceMonitor is an instance of, and the Prometheus that scrapes through it.
- `KubernetesPodMonitor` -- KubernetesKubePrometheusStack is a prerequisite because it installs the monitoring.coreos.com CRDs the PodMonitor is an instance of, and the Prometheus that scrapes through it.
- `KubernetesPrometheusRule` -- KubernetesKubePrometheusStack is a prerequisite because it installs the monitoring.coreos.com CRDs the PrometheusRule is an instance of, and the Prometheus that evaluates it.
- `KubernetesKyverno` -- 4080–4099: Kubernetes security, policy, and identity
- `KubernetesGatekeeper`
- `KubernetesKeycloak` -- Keycloak declarations compose the official Keycloak Operator (which reconciles the Keycloak CR this kind renders) and, on the recommended postgres vendor, a KubernetesPostgres database — both must resolve before the CR can converge.
- `KubernetesOpenBao`
- `KubernetesOpenFga` -- OpenFGA requires a datastore; the recommended arm composes a KubernetesPostgres database (the sandbox memory arm needs nothing, but the registry declares the shape real deployments require).
- `KubernetesKeycloakOperator`
- `KubernetesCloudNativePgOperator` -- 4100–4129: Kubernetes data platforms
- `KubernetesPostgres`
- `KubernetesValkey`
- `KubernetesPerconaMysqlOperator`
- `KubernetesMysql`
- `KubernetesPerconaMongoOperator`
- `KubernetesMongodb`
- `KubernetesStrimziKafkaOperator`
- `KubernetesKafka` -- container_kind: a Strimzi Kafka cluster is a place in the provider's own model — KafkaTopic and KafkaUser declarations BELONG to one cluster (the strimzi.io/cluster label) and are drawn inside its box. Clients that merely talk to the cluster (Connect, MirrorMaker2, UI, Karapace) carry containment_exempt on their bootstrap/trust references.
- `KubernetesKafkaTopic`
- `KubernetesKafkaUser`
- `KubernetesKafkaConnect` -- container_kind: a Connect cluster hosts the connectors deployed INTO it (KafkaConnector's strimzi.io/cluster label names its Connect cluster) — the same room shape as KubernetesKafka above.
- `KubernetesKafkaConnector`
- `KubernetesKafkaMirrorMaker2`
- `KubernetesKarapace`
- `KubernetesKafkaUi`
- `KubernetesOpenSearchOperator`
- `KubernetesOpenSearch`
- `KubernetesAltinityOperator`
- `KubernetesClickHouse`
- `KubernetesSolrOperator`
- `KubernetesSolr`
- `KubernetesNeo4j`
- `KubernetesSeaweedFs`
- `KubernetesQdrant`
- `KubernetesRabbitMqOperator` -- The RabbitMQ Cluster Operator's release manifest ships admission webhooks whose serving certificate is a cert-manager Certificate — cert-manager must be running before the operator installs.
- `KubernetesRabbitMq`
- `KubernetesCnpgBarmanCloudPlugin` -- The Barman Cloud plugin is CloudNativePG's object-store backup path, installed as its own Helm release beside the operator. It registers with the operator over CNPG-I and issues its TLS through cert-manager, so both are prerequisites: nothing to register with, or no certificates, and the release never becomes ready.
- `KubernetesAirflow` -- 4130–4149: Kubernetes analytics and ML KubernetesPostgres is a prerequisite because Airflow's metadata database composes a KubernetesPostgres by default (the spec's FK defaults resolve onto its outputs) and the migration Job needs the database reachable before the server components start.
- `KubernetesSparkOperator`
- `KubernetesKubeRayOperator`
- `KubernetesRayCluster` -- KubernetesKubeRayOperator is a prerequisite because this kind declares the RayCluster custom resource that only the operator's CRDs admit and only the operator reconciles into head and worker pods.
- `KubernetesFlinkOperator` -- KubernetesCertManager is a prerequisite because the Flink operator's chart, with its default-on admission webhook, renders cert-manager Issuer/Certificate resources and trusts the API server through cert-manager's CA injection — there is no self-signed fallback at the pinned chart, and the webhooks are fail-closed.
- `KubernetesFlinkDeployment` -- KubernetesFlinkOperator is a prerequisite because this kind declares the FlinkDeployment custom resource that only the operator's CRDs admit and only the operator reconciles into a running Flink cluster.
- `KubernetesJupyterHub` -- KubernetesPostgres is a prerequisite because JupyterHub's hub database composes a KubernetesPostgres in its external-database arm (the spec's FK defaults resolve onto its outputs) and the hub pod mounts that database's credential Secret before it can start.
- `KubernetesMlflow` -- KubernetesPostgres is a prerequisite because MLflow's backend store composes a KubernetesPostgres in its production arm (FK defaults onto its outputs; the module composes the connection URI from its credential Secret), and KubernetesSeaweedFs because the artifact store's S3-compatible arm FK-defaults onto the SeaweedFS endpoint and credential Secret.
- `KubernetesTrino` -- KubernetesPostgres is a prerequisite because Trino's postgres catalogs compose a KubernetesPostgres (the catalog host and credential FK-default onto its outputs), and the pods read that database's credential Secret to resolve catalog passwords from environment.
- `KubernetesSuperset` -- KubernetesPostgres is a prerequisite because Superset's REQUIRED metadata database composes a KubernetesPostgres (FK defaults onto its outputs; the module composes the environment Secret from its credential Secret), and KubernetesValkey because the cache/broker arm FK-defaults onto a KubernetesValkey's service and password Secret.
- `KubernetesLangfuse` -- Langfuse's four required stores each FK-default onto a catalog kind: KubernetesPostgres (users, projects, prompts, API keys - host and password Secret), KubernetesClickHouse (traces and scores - host, cluster name and auth Secret), KubernetesValkey (the ingestion queues - host, ACL user and password Secret), and KubernetesSeaweedFs (the S3-compatible event store - endpoint and credentials Secret).
- `KubernetesArgocd` -- 4150–4169: Kubernetes GitOps and CI/CD
- `KubernetesArgoWorkflows`
- `KubernetesTektonOperator`
- `KubernetesTekton` -- KubernetesTektonOperator is a prerequisite because this kind declares the TektonConfig custom resource that only the operator's CRDs admit and only the operator reconciles into running components.
- `KubernetesGhaRunnerScaleSetController`
- `KubernetesGhaRunnerScaleSet` -- KubernetesGhaRunnerScaleSetController is a prerequisite because this kind renders an AutoscalingRunnerSet custom resource that only the controller's CRDs admit and only the controller reconciles into listener and runner pods.
- `KubernetesHarbor`
- `KubernetesJenkins`
- `KubernetesTemporal` -- 4170–4179: Kubernetes app platforms KubernetesPostgres is a prerequisite because the recommended (and E2E-proven) database composition backs Temporal's default and visibility stores with a CloudNativePG cluster.
- `KubernetesNats`
- `KubernetesLocust`
- `KubernetesPlantonRunner`
- `KubernetesPlantonOperator`
- `KubernetesPlantonPlatform` -- KubernetesPlantonOperator is a prerequisite because this kind declares the PlantonPlatform custom resource that only the operator's CRD admits and only the operator reconciles into a running platform.
- `KubernetesGoFeatureFlag`
- `KubernetesFlagd`
- `KubernetesGoFeatureFlagFlagFile`
- `KubernetesFlagdFlagFile`
- `KubernetesAgentSandboxController` -- 4180–4199: agent sandbox runtimes (4180–4189 Kubernetes SIG Apps' agent-sandbox; 4190–4199 Agent Substrate)
- `KubernetesAgentSandbox` -- KubernetesAgentSandboxController is a prerequisite because it installs the agents.x-k8s.io definition the Sandbox is an instance of, and the controller that runs its pod.
- `KubernetesAgentSandboxTemplate` -- KubernetesAgentSandboxController is a prerequisite because it installs the extensions.agents.x-k8s.io definition the SandboxTemplate is an instance of, and the extension controllers that turn it into network policy and warm pool members.
- `KubernetesAgentSandboxWarmPool` -- KubernetesAgentSandboxController is a prerequisite because it installs the extensions.agents.x-k8s.io definition the SandboxWarmPool is an instance of, and the controller that keeps the pool's members ready.
- `KubernetesAgentSandboxClaim` -- KubernetesAgentSandboxController is a prerequisite because it installs the extensions.agents.x-k8s.io definition the SandboxClaim is an instance of, and the controller that binds it to a warm pool member.
- `KubernetesAgentSubstrate`
- `KubernetesAgentSubstrateWorkerPool` -- KubernetesAgentSubstrate is a prerequisite because it installs the ate.dev definition the WorkerPool is an instance of, the controller that turns it into worker pods, and the node agent and API server the workers register with.
- `KubernetesAgentSubstrateSandboxConfig` -- KubernetesAgentSubstrate is a prerequisite because it installs the ate.dev definition the SandboxConfig is an instance of, the admission policy that holds each class to its assets, and the API server and node agent that read it.
- `KubernetesAgentSubstrateCsiDriverConfig` -- KubernetesAgentSubstrate is a prerequisite because it installs the ate.dev definition the CSIDriverConfig is an instance of, and the API server and node agent that look a driver up in it.
- `KubernetesStigmer` -- 4200–4209: Kubernetes AI agent platforms KubernetesStigmer installs the Stigmer operator and declares the one Stigmer it reconciles, so it carries no prerequisite.
- `DigitalOceanApp` -- 5000–5999: DigitalOcean resources
- `DigitalOceanBucket`
- `DigitalOceanContainerRegistry`
- `DigitalOceanDatabaseCluster` -- A DigitalOcean managed database cluster -- the cluster every other managed-database resource is created on: users, logical databases, connection pools, the firewall rule set, Kafka topics, and Kafka schema subjects are all API-addressed under a cluster's id and cannot exist before it. A container kind for exactly that reason: those six kinds are children of the cluster, so on a diagram the cluster is the room they stand in. A read replica also names its primary but is a cluster of its own, a monitor alert names the clusters it watches, and an App Platform app names a cluster it attaches; those three references are containment-exempt on their specs.
- `DigitalOceanDnsZone`
- `DigitalOceanDroplet` -- No VPC prerequisite: the droplet spec's vpc reference is optional — an omitted vpc places the droplet in the region's default VPC.
- `DigitalOceanFirewall`
- `DigitalOceanFunction`
- `DigitalOceanKubernetesCluster` -- DigitalOceanVpc is a prerequisite because the cluster spec's vpc reference is required: the control plane and every node pool live inside one VPC, resolved to the DigitalOceanVpc's exported vpc_id output.
- `DigitalOceanKubernetesNodePool` -- DigitalOceanKubernetesCluster is a prerequisite because a node pool is API-addressed under its owning cluster: the spec's cluster reference is required and the pool cannot exist first.
- `DigitalOceanLoadBalancer` -- No registry prerequisite: the load balancer's vpc reference is optional (DigitalOcean places it in the region's default VPC when unset, and GLOBAL balancers take no VPC at all). Scenarios that exercise VPC placement declare it per-scenario via the e2e-prerequisites annotation.
- `DigitalOceanVolume`
- `DigitalOceanVpc`
- `DigitalOceanCertificate`
- `DigitalOceanDnsRecord` -- DigitalOceanDnsZone is a prerequisite because a record is API-addressed under its domain: the spec's domain reference is required, resolved to the DigitalOceanDnsZone's exported zone_name output.
- `DigitalOceanDatabaseUser` -- An additional user on a managed database cluster: the cluster reference is required, resolved to the DigitalOceanDatabaseCluster's exported cluster_id output.
- `DigitalOceanDatabaseDb` -- An additional logical database on a managed database cluster: the cluster reference is required.
- `DigitalOceanDatabaseConnectionPool` -- A PgBouncer connection pool on a managed PostgreSQL cluster: the cluster reference is required.
- `DigitalOceanDatabaseFirewall` -- The inbound trusted-sources rule set of a managed database cluster: the cluster reference is required.
- `DigitalOceanDatabaseReplica` -- A read-only replica of a managed database cluster: the primary cluster reference is required.
- `DigitalOceanDatabaseKafkaTopic` -- A topic on a managed Kafka cluster with the full per-topic configuration block; the owning cluster reference is required.
- `DigitalOceanDatabaseKafkaSchema` -- One schema subject registered in a managed Kafka cluster's schema registry; the owning cluster reference is required.
- `DigitalOceanProject` -- The account-level organizational container; membership is carried on the project itself as resource URNs.
- `DigitalOceanSshKey` -- An SSH public key registered on the account, referenced by droplets and droplet autoscale pools at create time.
- `DigitalOceanMonitorAlert` -- An alert policy on DigitalOcean's built-in metrics for droplets, load balancers, and managed database clusters. All entity targeting is optional, so there is no registry prerequisite.
- `DigitalOceanUptimeCheck` -- An availability/latency probe on an external endpoint with composed alert rules; the target is outside the account, so there is no registry prerequisite.
- `DigitalOceanReservedIp` -- A static public IP address (IPv4 or IPv6) reserved in a region and optionally assigned to a droplet. The droplet attachment is an optional composition seam, so there is no registry prerequisite.
- `DigitalOceanVpcPeering` -- A private-network peering connection between exactly two VPCs; both VPC references are required.
- `DigitalOceanSpacesKey` -- An access-key pair for Spaces object storage. Bucket grants are an optional composition seam, so there is no registry prerequisite.
- `DigitalOceanCdn` -- A CDN endpoint serving a Spaces bucket's content from the global edge: the origin reference is required, resolved to the DigitalOceanBucket's exported bucket_domain_name output.
- `DigitalOceanDropletAutoscalePool` -- A pool of identical droplets DigitalOcean keeps at a fixed size or scales on utilization. The template's ssh_keys reference is required (the API mandates SSH keys), resolved to the DigitalOceanSshKey's exported ssh_key_id output.
- `CloudflareDnsZone` -- 7000–7999: Cloudflare resources
- `CloudflareKvNamespace`
- `CloudflareR2Bucket`
- `CloudflareWorker`
- `CloudflareLoadBalancer` -- CloudflareDnsZone and CloudflareLoadBalancerPool are prerequisites because a load balancer is a DNS-level construct inside a zone (the spec's zone_id reference must resolve) and traffic must land somewhere (the required fallback_pool reference must resolve to a live pool).
- `CloudflareD1Database`
- `CloudflareZeroTrustAccessApplication`
- `CloudflareDnsRecord` -- CloudflareDnsZone is a prerequisite because every record lives inside a zone -- the spec's zone_id reference must resolve before the record can be created.
- `CloudflareRuleset` -- CloudflareDnsZone is a prerequisite because zone-scoped rulesets (the common case; the spec's zone_id reference defaults to the zone kind) must resolve their zone first. Account-scoped rulesets simply leave the reference unused.
- `CloudflareWorkersKvPair` -- CloudflareKvNamespace is a prerequisite because a KV pair is written into a namespace -- the spec's namespace_id reference must resolve first.
- `CloudflareHyperdriveConfig`
- `CloudflareLoadBalancerPool`
- `CloudflareLoadBalancerMonitor`
- `CloudflareZeroTrustAccessPolicy`
- `CloudflareZeroTrustAccessGroup`
- `CloudflareQueue`
- `CloudflarePagesProject`
- `CloudflareZeroTrustTunnel`
- `CloudflareZeroTrustTunnelVirtualNetwork`
- `CloudflareZeroTrustTunnelRoute` -- CloudflareZeroTrustTunnel is a prerequisite because a route steers a CIDR through an existing tunnel -- the spec's tunnel_id reference must resolve first. (The optional virtual_network_id is scenario-declared, not a registry prerequisite.)
- `CloudflareList`
- `CloudflareListItem` -- CloudflareList is a prerequisite because an item exists only inside a list -- the spec's list_id reference must resolve first.
- `CloudflareTurnstileWidget`
- `CloudflareEmailRoutingZone` -- CloudflareDnsZone is a prerequisite because email routing is enabled ON a zone -- the spec's zone_id reference must resolve first.
- `CloudflareEmailRoutingRule` -- CloudflareDnsZone is a prerequisite because a routing rule lives in a zone's email routing configuration -- the spec's zone_id reference must resolve first. CloudflareEmailRoutingZone is deliberately NOT a prerequisite: the API accepts rule creation on a zone whose Email Routing was never enabled (measured live 2026-08-26 -- POST zones/{id}/email/routing/rules answered 200 on a fresh zone with routing status "unconfigured"). Enablement gates mail FLOW, not rule lifecycle, so it is an operational concern the kind docs teach, not a create-time dependency. (Forward destinations reference CloudflareEmailRoutingAddress only for forward-type rules, so that edge is scenario-declared.)
- `CloudflareEmailRoutingAddress`
- `CloudflareOriginCaCertificate`
- `CloudflareCertificatePack` -- CloudflareDnsZone is a prerequisite because a certificate pack is ordered for a zone's hostnames -- the spec's zone_id reference must resolve first.
- `CloudflareCustomHostname` -- CloudflareDnsZone is a prerequisite because a custom hostname (SSL for SaaS) is provisioned inside a zone -- the spec's zone_id reference must resolve first.
- `CloudflareCustomHostnameFallbackOrigin` -- CloudflareDnsZone is a prerequisite because the fallback origin is a zone-level SSL-for-SaaS setting -- the spec's zone_id reference must resolve first.
- `CloudflareZeroTrustAccessIdentityProvider` -- No prerequisites: identity providers are account-scoped in the canonical case (the optional zone scope is a per-manifest choice, not a structural dependency).
- `CloudflareZeroTrustAccessServiceToken` -- No prerequisites: service tokens are account-scoped in the canonical case (the optional zone scope is a per-manifest choice, not a structural dependency).
- `CloudflareZeroTrustOrganization` -- No prerequisites: the organization is an account-scoped configuration singleton (the optional zone scope is a per-manifest choice, not a structural dependency).
- `CloudflareZeroTrustAccessInfrastructureTarget` -- No prerequisites: targets are account-scoped, and the virtual-network reference is an optional per-manifest edge (omitted = the account's default virtual network).
- `CloudflareZeroTrustMcpPortal` -- No prerequisites: portals are account-scoped, and the servers[] rows' MCP-server references are optional per-manifest edges.
- `CloudflareZeroTrustMcpServer` -- No prerequisites: MCP server registrations are account-scoped and self-contained -- portals reference them, not the reverse.
- `CloudflareZeroTrustGatewayPolicy` -- No prerequisites: Gateway policies are account-scoped, and their list / virtual-network references are optional per-manifest edges.
- `CloudflareZeroTrustList` -- No prerequisites: Zero Trust lists are account-scoped and self-contained.
- `CloudflareZeroTrustGatewaySettings` -- No prerequisites: the Gateway configuration is an account-scoped singleton, and its certificate reference is an optional per-manifest edge.
- `CloudflareZeroTrustDnsLocation` -- No prerequisites: DNS locations are account-scoped and self-contained.
- `CloudflareZeroTrustDeviceDefaultProfile` -- No prerequisites: the default device profile is an account-scoped configuration singleton; its virtual-network and zone-certificate references are optional per-manifest edges.
- `CloudflareZeroTrustDeviceCustomProfile` -- No prerequisites: custom device profiles are account-scoped, and the virtual-network reference is an optional per-manifest edge.
- `CloudflareZeroTrustDevicePostureRule` -- No prerequisites: posture rules are account-scoped and self-contained (list and integration references are literal UUIDs today).
- `CloudflareZoneTlsSettings` -- CloudflareDnsZone is a prerequisite because TLS settings are zone-scoped configuration -- the spec's zone_id reference must resolve first.
- `CloudflareCustomSslCertificate` -- CloudflareDnsZone is a prerequisite because a custom certificate is uploaded to an existing zone -- the spec's zone_id reference must resolve first.
- `CloudflareMtlsCertificate` -- No prerequisites: mTLS certificates are account-scoped uploads and self-contained -- consumers (zone TLS CA associations, Authenticated Origin Pulls rows, Workers mTLS bindings) reference them, not the reverse.
- `CloudflareAuthenticatedOriginPulls` -- CloudflareDnsZone is a prerequisite because Authenticated Origin Pulls enablement configures an existing zone -- the spec's zone_id reference must resolve first. The per-hostname certificate edge is optional and scenario-declared, never a registry prerequisite.
- `CloudflareAuthenticatedOriginPullsCertificate` -- CloudflareDnsZone is a prerequisite because the client certificate is uploaded to an existing zone -- the spec's zone_id reference must resolve first.
- `CloudflareZoneSettings` -- CloudflareDnsZone is a prerequisite because zone settings configure an existing zone -- the spec's zone_id reference must resolve first.
- `CloudflareCacheSettings` -- CloudflareDnsZone is a prerequisite because cache settings configure an existing zone -- the spec's zone_id reference must resolve first.
- `CloudflareIpAccessRule` -- No prerequisites: IP Access rules are account-scoped in the canonical case (the zone scope is a per-manifest choice, not a structural dependency).
- `CloudflareBotManagement` -- CloudflareDnsZone is a prerequisite because Bot Management is zone-singleton configuration -- the spec's zone_id reference must resolve first.
- `CloudflareSnippet` -- CloudflareDnsZone is a prerequisite because snippets deploy to a zone -- the spec's zone_id reference must resolve first.
- `CloudflareSnippetRules` -- CloudflareDnsZone and CloudflareSnippet are prerequisites: the rules table is zone-scoped and every rule invokes a snippet by name.
- `CloudflareWaitingRoom` -- CloudflareDnsZone is a prerequisite because waiting rooms sit on a zone's host+path -- the spec's zone_id reference must resolve first.
- `CloudflareWaitingRoomEvent` -- CloudflareWaitingRoom is a prerequisite because events run on a room (and the room's own chain brings the zone).
- `CloudflareLogpushJob` -- No prerequisites: logpush jobs are dual-scope (account or zone) and the zone reference is optional -- zone-scoped lanes declare the edge at the scenario level.
- `CloudflareNotificationPolicy` -- No prerequisites: every delivery mechanism (email, PagerDuty, webhook) is optional -- policies referencing a webhook declare the edge at the scenario level.
- `CloudflareNotificationWebhook` -- No prerequisites: a webhook destination is account-scoped and self-contained -- notification policies reference it, not the reverse.
- `CloudflareWebAnalyticsSite` -- No prerequisites: a site is identified by host OR zone, and the zone reference is optional -- zone-measured lanes declare the edge at the scenario level.
- `CloudflareWorkflow` -- CloudflareWorker is a prerequisite because a workflow registers a class exported by a DEPLOYED Worker script -- the spec's script_name reference must resolve first.
- `CloudflareSecretsStore` -- No prerequisites: the Secrets Store is an account-scoped container and self-contained -- consumers (store secrets, Worker bindings, AI Gateway authentication) reference it, not the reverse.
- `CloudflareSecretsStoreSecret` -- CloudflareSecretsStore is a prerequisite because every secret lives inside a store -- the spec's store_id reference must resolve first.
- `CloudflareAiGateway` -- No prerequisites: the gateway is account-scoped and self-contained; its optional Secrets Store link (BYO provider keys) is a scenario-level composition, not a structural requirement.
- `CloudflareAccountApiToken` -- No prerequisites: an account-owned API token is self-contained; the resources its policies cover are identifiers, not references.
- `CloudflareHealthcheck` -- CloudflareDnsZone is a prerequisite because standalone health checks are zone-scoped -- the spec's zone_id reference must resolve first.
- `Auth0Connection` -- 8000–8999: Auth0 resources
- `Auth0Client`
- `Auth0EventStream`
- `Auth0ResourceServer`
- `Auth0Action`
- `Auth0Role`
- `Auth0User`
- `Auth0TenantSettings`
- `Auth0CustomDomain`
- `Auth0CustomDomainVerification`
- `Auth0Branding` -- 8010–8015: how the tenant's Universal Login looks, reads and mails -- branding and theme, the login flow, each prompt's words and screen partials, the email service and each email.
- `Auth0Prompt`
- `Auth0PromptCustomText`
- `Auth0PromptScreenPartials`
- `Auth0EmailProvider`
- `Auth0EmailTemplate`
- `Auth0ClientFromMetadataDocument`
- `OpenFgaStore` -- 9000–9999: OpenFGA resources OpenFGA publishes a Terraform provider and no Pulumi provider, so its kinds ship one HCL module and declare the engines that run it; the CLI and the platform refuse Pulumi for them before anything runs.
- `OpenFgaAuthorizationModel`
- `OpenFgaRelationshipTuple`
- `StripeWebhookEndpoint` -- 10000–10999: Stripe resources Stripe publishes a Terraform provider and no Pulumi provider, and its kinds are proven on OpenTofu alone, so each declares ["tofu"]: one HCL module, and every other engine refused before anything runs. Family sub-bands: 10000 account and delivery, 10100 catalog and pricing.
- `StripeEventDestination`
- `StripeBillingPortalConfiguration`
- `StripePaymentMethodConfiguration`
- `StripePaymentMethodDomain`
- `StripeRadarValueList`
- `StripeProduct`
- `StripePrice` -- A price always belongs to a product (its product field is required), so the product deploys first.
- `StripeCoupon`
- `StripePromotionCode` -- A promotion code always redeems a coupon (its coupon field is required), so the coupon deploys first.
- `StripeEntitlementFeature`
- `StripeShippingRate`
- `StripeTaxRate` -- Tax rate and tax registration share the initials "tr", so each takes Stripe's own id prefix with its vowels dropped (txr_, taxreg_).
- `StripeTaxRegistration`
- `StripeBillingMeter`
- `StripePaymentLink` -- A payment link always sells at least one price (line_items is required), so the price deploys first.

### spec.runner.extraEnv.secrets[].valueFrom.env

`string`

### spec.runner.extraEnv.secrets[].valueFrom.name

`string` · required

- rule: {"required":true}

### spec.runner.extraEnv.secrets[].valueFrom.fieldPath

`string`

### spec.runner.extraEnv.envFrom

`[]EnvFromSource`

Bulk import of environment variables from ConfigMaps or Secrets.

### spec.runner.extraEnv.envFrom[].prefix

`string`

Optional prefix prepended to each imported key name.
For example, prefix "APP_" with key "PORT" produces env var "APP_PORT".

### spec.runner.extraEnv.envFrom[].configMapRef

`ConfigMapRef`

Import all keys from a ConfigMap.

### spec.runner.extraEnv.envFrom[].configMapRef.name

`string` · required

Name of the ConfigMap.

- rule: {"required":true}

### spec.runner.extraEnv.envFrom[].configMapRef.optional

`bool`

If true, the ConfigMap is allowed to not exist without blocking pod startup.

### spec.runner.extraEnv.envFrom[].secretRef

`SecretRef`

Import all keys from a Secret.

### spec.runner.extraEnv.envFrom[].secretRef.name

`string` · required

Name of the Secret.

- rule: {"required":true}

### spec.runner.extraEnv.envFrom[].secretRef.optional

`bool`

If true, the Secret is allowed to not exist without blocking pod startup.

### spec.database

`KubernetesStigmerDatabase`

Where Stigmer keeps its records. Unset: the operator runs Postgres for
you (through CloudNativePG, which it installs when missing).

### spec.database.bundled

`KubernetesStigmerBundledPostgres`

The Postgres the operator runs, through CloudNativePG.

### spec.database.bundled.instances

`int32` · optional (explicit presence)

Postgres instances: one primary, the rest standbys.

- default: `1`
- rule: {"int32":{"lte":5,"gte":1}}

### spec.database.bundled.storage

`KubernetesStigmerVolume`

Each instance's volume. Empty = 10Gi.

### spec.database.bundled.storage.size

`string`

The claim's requested capacity, as a Kubernetes quantity (10Gi). Empty =
the operator's default for this volume.

- rule: size must be a Kubernetes quantity such as 10Gi

### spec.database.bundled.storage.storageClassName

`string`

The storage class the claim binds with. Empty = the cluster's default
class (some clusters, GKE 1.37 among them, have none: name one).

### spec.database.bundled.resources

`KubernetesCoreV1ApiResourceRequirements`

Each instance's requests and limits. Empty = the operator's defaults.

### spec.database.bundled.resources.limits

`map<string, string>`

`limits` describes the maximum amount of compute resources allowed. More
info:
https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/

- rule: {"map":{"values":{"string":{"pattern":"^(\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))))?$"}}}}

### spec.database.bundled.resources.requests

`map<string, string>`

`requests` describes the minimum amount of compute resources required. If
Requests is omitted for a container, it defaults to Limits if that is
explicitly specified, otherwise to an implementation-defined value.
Requests cannot exceed Limits. More info:
https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/

- rule: {"map":{"values":{"string":{"pattern":"^(\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))))?$"}}}}

### spec.database.bundled.resources.claims

`[]KubernetesCoreV1ApiResourceClaim`

`claims` lists the names of resources, defined in spec.resourceClaims, that
are used by this container.

This field depends on the DynamicResourceAllocation feature gate.

This field is immutable. It can only be set for containers.

### spec.database.bundled.resources.claims[].name

`string` · required

`name` must match the name of one entry in pod.spec.resourceClaims of the
Pod where this field is used. It makes that resource available inside a
container.

- rule: {"string":{"minLen":"1"}}

### spec.database.bundled.resources.claims[].request

`string`

`request` is the name chosen for a request in the referenced claim. If
empty, everything from the claim is made available, otherwise only the
result of this request.

### spec.database.bundled.backup

`KubernetesStigmerPostgresBackup`

Backups. With object_store declared, the database archives every change
to the bucket from its first minute and takes a full backup on this
schedule; these settings tune that. Needs object_store, and cert-manager
on the cluster (the backup plugin needs it; nothing here installs it).

### spec.database.bundled.backup.schedule

`string`

When a full backup is taken: a six-field cron with seconds first, in
UTC. Empty = daily at 02:00 ("0 0 2 * * *").

- rule: backup.schedule is a six-field cron with seconds first, such as "0 0 2 * * *"

### spec.database.bundled.backup.retention

`string`

How long backups and the archive are kept: a number and d, w or m.
Empty = 30d.

- rule: backup.retention is a number and d, w or m, such as 30d

### spec.database.bundled.recoverFrom

`KubernetesStigmerRecoverFrom`

Creates the database from an archive in the bucket instead of empty.
Honored only when the database is created; on a running install it
changes nothing. A restore also needs the source install's keys Secret
in this namespace (see keys_export_command): without it the encryption
key would be minted fresh and the restored records unreadable, so this
resource refuses to declare the Stigmer until the keys are there.

### spec.database.bundled.recoverFrom.serverName

`string` · required

The archive's name in the bucket: the source install's backup server
name (its status.backup.serverName).

- rule: {"required":true}

### spec.database.bundled.recoverFrom.targetTime

`string`

The moment to restore to, in RFC 3339 (2026-11-20T08:00:00Z). Empty =
the newest moment the archive holds.

- rule: recover_from.target_time must be an RFC 3339 time such as 2026-11-20T08:00:00Z

### spec.database.external

`KubernetesStigmerExternalPostgres`

A Postgres you run. It needs your own Temporal (temporal.external) and
your own sign-in issuer when the install has a front door.

### spec.database.external.host

`string | valueFrom` · required

The server's address. Composes a KubernetesPostgres's read-write
Service; any reachable Postgres works as a literal.

- references: KubernetesPostgres (`status.outputs.rw_service`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesPostgres, name: <that resource's name>, fieldPath: status.outputs.rw_service}} -- a bare string does not parse

### spec.database.external.port

`int32` · optional (explicit presence)

The server's port.

- default: `5432`
- rule: {"int32":{"lte":65535,"gte":1}}

### spec.database.external.user

`string` · required

The role Stigmer signs in as. It must own `database`, where Stigmer
creates its schemas.

- rule: {"required":true}

### spec.database.external.database

`string` · optional (explicit presence)

The database Stigmer uses.

- default: `stigmer`

### spec.database.external.password

`KubernetesStigmerPassword` · required

The role's password: given here (this resource writes it into a Secret),
or read from a Secret in the Stigmer's namespace.

- rule: {"required":true}

### spec.database.external.password.value

`string` · sensitive

The password. This resource writes it into a Secret in the Stigmer's
namespace.

### spec.database.external.password.existingSecret

`KubernetesStigmerPostgresPasswordSecret`

A Secret in the Stigmer's namespace that holds it. Composes a
KubernetesPostgres's application credential Secret when the database
runs in the same namespace.

### spec.database.external.password.existingSecret.name

`string | valueFrom` · required

The Secret's name.

- references: KubernetesPostgres (`status.outputs.password_secret.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesPostgres, name: <that resource's name>, fieldPath: status.outputs.password_secret.name}} -- a bare string does not parse

### spec.database.external.password.existingSecret.key

`string` · optional (explicit presence)

The key holding the password.

- default: `password`

### spec.database.external.sslMode

`string`

libpq's sslmode for the connection. Empty = libpq's default (prefer).

- rule: {"string":{"in":["","disable","allow","prefer","require","verify-ca","verify-full"]}}

### spec.temporal

`KubernetesStigmerTemporal`

The workflow engine Stigmer's runs execute on. Unset: the operator runs
Temporal for you on the bundled database.

### spec.temporal.bundled

`KubernetesStigmerBundledTemporal`

The Temporal the operator runs on the bundled database.

### spec.temporal.bundled.resources

`KubernetesCoreV1ApiResourceRequirements`

Requests and limits for each Temporal service (frontend, history,
matching, worker). Empty = the operator's defaults.

### spec.temporal.bundled.resources.limits

`map<string, string>`

`limits` describes the maximum amount of compute resources allowed. More
info:
https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/

- rule: {"map":{"values":{"string":{"pattern":"^(\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))))?$"}}}}

### spec.temporal.bundled.resources.requests

`map<string, string>`

`requests` describes the minimum amount of compute resources required. If
Requests is omitted for a container, it defaults to Limits if that is
explicitly specified, otherwise to an implementation-defined value.
Requests cannot exceed Limits. More info:
https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/

- rule: {"map":{"values":{"string":{"pattern":"^(\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))))?$"}}}}

### spec.temporal.bundled.resources.claims

`[]KubernetesCoreV1ApiResourceClaim`

`claims` lists the names of resources, defined in spec.resourceClaims, that
are used by this container.

This field depends on the DynamicResourceAllocation feature gate.

This field is immutable. It can only be set for containers.

### spec.temporal.bundled.resources.claims[].name

`string` · required

`name` must match the name of one entry in pod.spec.resourceClaims of the
Pod where this field is used. It makes that resource available inside a
container.

- rule: {"string":{"minLen":"1"}}

### spec.temporal.bundled.resources.claims[].request

`string`

`request` is the name chosen for a request in the referenced claim. If
empty, everything from the claim is made available, otherwise only the
result of this request.

### spec.temporal.external

`KubernetesStigmerExternalTemporal`

A Temporal frontend you run.

### spec.temporal.external.hostPort

`string | valueFrom` · required

The frontend's gRPC address, host:port (an IPv6 address in brackets).
Composes a KubernetesTemporal's frontend endpoint.

- references: KubernetesTemporal (`status.outputs.frontend_endpoint`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesTemporal, name: <that resource's name>, fieldPath: status.outputs.frontend_endpoint}} -- a bare string does not parse

### spec.temporal.external.namespace

`string` · optional (explicit presence)

The Temporal namespace Stigmer's runs use.

- default: `default`

### spec.temporal.external.tls

`KubernetesStigmerTemporalTls`

TLS to the frontend. Without certificates named, the frontend's
certificate is checked against the system's trusted roots.

### spec.temporal.external.tls.serverName

`string`

The name the frontend's certificate is checked against, when it differs
from the host.

### spec.temporal.external.tls.pem

`KubernetesStigmerTemporalTlsPem`

The certificates, given here as PEM. This resource writes them into a
Secret in the Stigmer's namespace.

- rule: temporal.external.tls.pem: client_certificate and client_private_key go together; mutual TLS needs the certificate and its key

### spec.temporal.external.tls.pem.caCertificate

`string`

The CA that signed the frontend's certificate. Empty = the system's
trusted roots.

### spec.temporal.external.tls.pem.clientCertificate

`string`

The client certificate, for mutual TLS.

### spec.temporal.external.tls.pem.clientPrivateKey

`string` · sensitive

The client certificate's private key.

### spec.temporal.external.tls.existingSecret

`KubernetesStigmerTemporalTlsSecret`

A Secret in the Stigmer's namespace holding the certificates as PEM,
read by key.

- rule: temporal.external.tls.existing_secret: certificate_key and private_key_key go together; mutual TLS needs the certificate and its key

### spec.temporal.external.tls.existingSecret.name

`string | valueFrom` · required

The Secret's name.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.temporal.external.tls.existingSecret.caKey

`string`

The key of the CA that signed the frontend's certificate.

### spec.temporal.external.tls.existingSecret.certificateKey

`string`

The key of the client certificate, for mutual TLS.

### spec.temporal.external.tls.existingSecret.privateKeyKey

`string`

The key of the client certificate's private key.

### spec.temporal.external.apiKey

`KubernetesStigmerSecretValue`

The frontend's API key (Temporal Cloud, or a frontend with a JWT
authorizer). An API key implies TLS.

### spec.temporal.external.apiKey.value

`string` · sensitive

The value.

### spec.temporal.external.apiKey.existingSecret

`KubernetesStigmerSecretKeyRef`

A key in an existing Secret in the Stigmer's namespace.

### spec.temporal.external.apiKey.existingSecret.name

`string | valueFrom` · required

The Secret's name.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.temporal.external.apiKey.existingSecret.key

`string` · required

The key within the Secret.

- rule: {"required":true}

### spec.objectStore

`KubernetesStigmerObjectStore`

The bucket Stigmer keeps its files in: what a run stores beside its
record, Skills, and the database's archive. Declared, the server and the
runner share no disk, runners read and write files through short-lived
URLs without ever holding the bucket's keys, and the bundled database
archives to the bucket from its first minute. Unset: files live on a
volume the server and the shared runner share. Backups, restores and
per-session sandboxes all need it.

### spec.objectStore.s3

`KubernetesStigmerS3` · required

Any S3-compatible bucket: AWS S3, Cloudflare R2, SeaweedFS, MinIO, or
Google Cloud Storage through its S3 API (Agent Substrate's snapshots
need a store other than Cloud Storage, whose S3 API refuses them).

- rule: {"required":true}

### spec.objectStore.s3.bucket

`string | valueFrom` · required

The bucket's name (3 to 63 characters). Composes an AwsS3Bucket.

- references: AwsS3Bucket (`status.outputs.bucket_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_id}} -- a bare string does not parse

### spec.objectStore.s3.region

`string` · optional (explicit presence)

The bucket's region; R2 also accepts us-east-1.

- default: `us-east-1`

### spec.objectStore.s3.endpoint

`string | valueFrom`

The store's URL, such as https://<account>.r2.cloudflarestorage.com or
http://seaweedfs-s3.storage.svc.cluster.local:8333. Empty = AWS S3
itself. Composes a KubernetesSeaweedFs's S3 endpoint.

- references: KubernetesSeaweedFs (`status.outputs.s3_endpoint`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSeaweedFs, name: <that resource's name>, fieldPath: status.outputs.s3_endpoint}} -- a bare string does not parse

### spec.objectStore.s3.credentials

`KubernetesStigmerS3Credentials` · required

The bucket's access key, which the server and the database's backups
use; no runner ever holds it.

- rule: {"required":true}

### spec.objectStore.s3.credentials.accessKey

`KubernetesStigmerS3AccessKey`

The key pair. This resource writes it into a Secret under the names
Stigmer reads (ACCESS_KEY_ID, SECRET_ACCESS_KEY).

### spec.objectStore.s3.credentials.accessKey.accessKeyId

`string` · required

The access key id.

- rule: {"required":true}

### spec.objectStore.s3.credentials.accessKey.secretAccessKey

`string` · required · sensitive

The secret access key.

- rule: {"required":true}

### spec.objectStore.s3.credentials.existingSecretName

`string | valueFrom`

A Secret in the Stigmer's namespace that already holds the pair as
ACCESS_KEY_ID and SECRET_ACCESS_KEY.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.objectStore.s3.runnerEndpoint

`string`

The address runners reach the store at, when it is not `endpoint`: URLs
handed to runners are signed for it, since a signed URL's host is part
of its signature.

- rule: object_store.s3.runner_endpoint must be an http or https URL

### spec.ingress

`KubernetesStigmerIngress`

The front door: the one hostname people reach Stigmer at, for the
console, the API, sign-in and files. Unset: the install is reached by
port-forward (`port_forward_command`).

### spec.ingress.hostname

`string` · required

The one name people reach Stigmer at, on port 443 (the CLI and the
desktop app speak TLS to a server on 443).

- rule: {"required":true,"string":{"pattern":"^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)+$"}}

### spec.ingress.tls

`KubernetesStigmerIngressTls`

An Ingress the cluster's controller serves, with its certificate.

### spec.ingress.tls.secretName

`string`

A TLS Secret in the Stigmer's namespace that you keep current.

### spec.ingress.tls.issuer

`KubernetesStigmerIssuerRef`

A cert-manager issuer that issues the certificate. cert-manager must
run on the cluster.

### spec.ingress.tls.issuer.name

`string | valueFrom` · required

The issuer's name. Composes a KubernetesClusterIssuer.

- references: KubernetesClusterIssuer (`status.outputs.cluster_issuer_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesClusterIssuer, name: <that resource's name>, fieldPath: status.outputs.cluster_issuer_name}} -- a bare string does not parse

### spec.ingress.tls.issuer.kind

`string`

ClusterIssuer (the default) or Issuer (in the Stigmer's namespace).

- rule: {"string":{"in":["","ClusterIssuer","Issuer"]}}

### spec.ingress.gateway

`KubernetesStigmerGatewayRef`

Routes on a Gateway the cluster runs, whose HTTPS listener for the
hostname holds the certificate. Stigmer never edits the Gateway.

### spec.ingress.gateway.name

`string | valueFrom` · required

The Gateway's name. Composes a KubernetesGateway.

- references: KubernetesGateway (`status.outputs.gateway_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesGateway, name: <that resource's name>, fieldPath: status.outputs.gateway_name}} -- a bare string does not parse

### spec.ingress.gateway.namespace

`string | valueFrom`

The Gateway's namespace. Empty = the Stigmer's namespace.

- references: KubernetesGateway (`status.outputs.namespace`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesGateway, name: <that resource's name>, fieldPath: status.outputs.namespace}} -- a bare string does not parse

### spec.ingress.gateway.sectionName

`string`

The listener the routes attach to.

### spec.ingress.className

`string`

The IngressClass that serves the front door (with `tls`).

### spec.ingress.annotations

`map<string, string>`

Annotations on the Ingress, for a controller that needs a hint to speak
gRPC to the server.

### spec.ingress.serviceAnnotations

`map<string, string>`

Annotations on the server's Service, for a controller that reads its
gRPC hint there (Traefik's
traefik.ingress.kubernetes.io/service.serversscheme: h2c).

### spec.identity

`KubernetesStigmerIdentity`

How people sign in. Unset on an install with a front door: the operator
runs the bundled sign-in. Unset on one without: the server signs nobody
in and every caller acts as the owner, which is safe only while the
install is reachable from inside the cluster.

### spec.identity.bundled

`KubernetesStigmerBundledIdentity`

The sign-in the operator runs beside Stigmer, at <hostname>/idp. The
owner's account is created with a one-time password
(`owner_password_command`).

- rule: sign-up for your email domains needs mail: without a verification email anyone could claim an address in the domain. Set identity.bundled.smtp

### spec.identity.bundled.signUpEmailDomains

`[]string`

The domains whose addresses may sign up once the owner has signed in,
such as acme.dev. Empty = only the owner signs in. Needs smtp, because
without a verification email anyone could claim an address.

- rule: {"repeated":{"items":{"string":{"pattern":"^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)+$"}}}}

### spec.identity.bundled.smtp

`KubernetesStigmerSmtp`

The mail server the sign-in sends verification and password mail
through.

### spec.identity.bundled.smtp.host

`string` · required

The mail server's address.

- rule: {"required":true}

### spec.identity.bundled.smtp.port

`int32` · optional (explicit presence)

The mail server's port.

- default: `587`
- rule: {"int32":{"lte":65535,"gte":1}}

### spec.identity.bundled.smtp.from

`string` · required

The address mail is sent from.

- rule: identity.bundled.smtp.from must be an email address
- rule: {"required":true}

### spec.identity.bundled.smtp.starttls

`bool`

Upgrade the connection to TLS.

### spec.identity.bundled.smtp.credentials

`KubernetesStigmerSmtpCredentials`

The user name and password, when the mail server asks for them.

### spec.identity.bundled.smtp.credentials.login

`KubernetesStigmerSmtpLogin`

The user name and password. This resource writes them into a Secret
under the names Stigmer reads (username, password).

### spec.identity.bundled.smtp.credentials.login.username

`string` · required

The user name.

- rule: {"required":true}

### spec.identity.bundled.smtp.credentials.login.password

`string` · required · sensitive

The password.

- rule: {"required":true}

### spec.identity.bundled.smtp.credentials.existingSecretName

`string | valueFrom`

A Secret in the Stigmer's namespace with the keys username and
password.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.identity.oidc

`KubernetesStigmerOidc`

Your own OpenID Connect issuer. Its realm or tenant, the console and
native clients, the audience, the owner's account and sign-up are
yours to set up.

### spec.identity.oidc.issuer

`string` · required

The issuer's URL, exactly as its discovery document names it: https, or
plain http only for a loopback issuer.

- rule: identity.oidc.issuer must be an https URL (plain http is allowed only for a loopback issuer)
- rule: {"required":true}

### spec.identity.oidc.audience

`string` · required

The audience the issuer puts in Stigmer's tokens.

- rule: {"required":true}

### spec.identity.oidc.consoleClientId

`string`

The web console's client.

### spec.identity.oidc.nativeClientId

`string`

The desktop app's and the CLI's client.

### spec.identity.oidc.discoveryUrl

`string`

Where the server fetches the issuer's discovery document when it cannot
reach the issuer's public name from inside the cluster. The document
must still name `issuer`.

### spec.enterprise

`KubernetesStigmerEnterprise`

Stigmer Enterprise, under a license.

### spec.enterprise.license

`KubernetesStigmerSecretValue` · required

The license key you were sent. The server runs the Enterprise image and
verifies it at boot.

- rule: {"required":true}

### spec.enterprise.license.value

`string` · sensitive

The value.

### spec.enterprise.license.existingSecret

`KubernetesStigmerSecretKeyRef`

A key in an existing Secret in the Stigmer's namespace.

### spec.enterprise.license.existingSecret.name

`string | valueFrom` · required

The Secret's name.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.enterprise.license.existingSecret.key

`string` · required

The key within the Secret.

- rule: {"required":true}

### spec.enterprise.organizationCreators

`[]string`

The account ids (ida_ and 26 characters, as `stigmer auth whoami`
prints them) who may create an organization. Empty = anyone signed in.

### spec.enterprise.accessHistoryRetentionDays

`int32` · optional (explicit presence)

How long the access history is kept, in days. Empty = every entry is
kept.

- rule: {"int32":{"gte":1}}

### spec.enterprise.outboundEgress

`string`

Which addresses the server dials when it reaches a URL a person wrote.
Empty = strict.

- rule: {"string":{"in":["","strict","relaxed"]}}

### spec.sandboxes

`KubernetesStigmerSandboxes`

Where engineers' agents run: on the shared runner beside the server, or
each session in a sandbox of its own.

- rule: sandboxes.agent_sandbox tunes the agentSandbox driver: set sandboxes.driver to agentSandbox, or remove it
- rule: sandboxes.substrate tunes the substrate driver: set sandboxes.driver to substrate, or remove it

### spec.sandboxes.driver

`string`

Where each session runs: shared (every session on the shared runner
beside the server, one pod and one disk), agentSandbox (each session in a
pod of its own through agent-sandbox, agents.x-k8s.io) or substrate (each
session in an Agent Substrate actor, a gVisor sandbox on a pool of
workers, snapshotted to the bucket while it sleeps). Empty = shared. Both
per-session drivers need object_store, and the operator installs what
they run on.

- rule: {"string":{"in":["","shared","agentSandbox","substrate"]}}

### spec.sandboxes.idleSeconds

`int32` · optional (explicit presence)

How long a quiet sandbox stays awake before it sleeps; it wakes with its
files on the session's next message. Empty = the driver's own default.

- rule: {"int32":{"gte":30}}

### spec.sandboxes.agentSandbox

`KubernetesStigmerAgentSandbox`

Tunes the agentSandbox driver.

### spec.sandboxes.agentSandbox.egressAllowCidrs

`[]string`

Private ranges a sandbox may reach besides the server and the bucket,
such as a model gateway or an MCP server on your network. Sandboxes
always reach the server, the bucket, DNS and the public internet.

- rule: {"repeated":{"maxItems":"64","items":{"string":{"maxLen":"43","ipPrefix":true}}}}

### spec.sandboxes.substrate

`KubernetesStigmerSubstrate`

Tunes the substrate driver.

### spec.sandboxes.substrate.workers

`int32` · optional (explicit presence)

How many gVisor workers the pool keeps. Sessions beyond what they hold
wait for a free seat.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.sandboxes.substrate.workerResources

`KubernetesCoreV1ApiResourceRequirements`

Each worker pod's requests and limits; the limits are the room every
sandbox on it shares.

### spec.sandboxes.substrate.workerResources.limits

`map<string, string>`

`limits` describes the maximum amount of compute resources allowed. More
info:
https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/

- rule: {"map":{"values":{"string":{"pattern":"^(\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))))?$"}}}}

### spec.sandboxes.substrate.workerResources.requests

`map<string, string>`

`requests` describes the minimum amount of compute resources required. If
Requests is omitted for a container, it defaults to Limits if that is
explicitly specified, otherwise to an implementation-defined value.
Requests cannot exceed Limits. More info:
https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/

- rule: {"map":{"values":{"string":{"pattern":"^(\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))))?$"}}}}

### spec.sandboxes.substrate.workerResources.claims

`[]KubernetesCoreV1ApiResourceClaim`

`claims` lists the names of resources, defined in spec.resourceClaims, that
are used by this container.

This field depends on the DynamicResourceAllocation feature gate.

This field is immutable. It can only be set for containers.

### spec.sandboxes.substrate.workerResources.claims[].name

`string` · required

`name` must match the name of one entry in pod.spec.resourceClaims of the
Pod where this field is used. It makes that resource available inside a
container.

- rule: {"string":{"minLen":"1"}}

### spec.sandboxes.substrate.workerResources.claims[].request

`string`

`request` is the name chosen for a request in the referenced claim. If
empty, everything from the claim is made available, otherwise only the
result of this request.

### spec.sandboxes.substrate.workerNodeSelector

`map<string, string>`

Places the workers on matching nodes.

### spec.sandboxes.substrate.workerTolerations

`[]KubernetesCoreV1ApiToleration`

Lets the workers run on tainted nodes.

- rule: A toleration with operator Exists matches any value, so it takes no value

### spec.sandboxes.substrate.workerTolerations[].key

`string`

`key` is the taint key that the toleration applies to. Empty means match
all taint keys. If the key is empty, operator must be Exists; this
combination means to match all values and all keys.

### spec.sandboxes.substrate.workerTolerations[].operator

`string` · optional (explicit presence)

`operator` represents a key's relationship to the value. Valid operators
are Exists, Equal, Lt, and Gt. Defaults to Equal. Exists is equivalent to
wildcard for value, so that a pod can tolerate all taints of a particular
category. Lt and Gt perform numeric comparisons (requires feature gate
TaintTolerationComparisonOperators).

- rule: {"string":{"in":["Exists","Equal","Lt","Gt"]}}

### spec.sandboxes.substrate.workerTolerations[].value

`string`

`value` is the taint value the toleration matches to. If the operator is
Exists, the value should be empty, otherwise just a regular string.

### spec.sandboxes.substrate.workerTolerations[].effect

`string` · optional (explicit presence)

`effect` indicates the taint effect to match. Empty means match all taint
effects. When specified, allowed values are NoSchedule, PreferNoSchedule
and NoExecute.

- rule: {"string":{"in":["NoSchedule","PreferNoSchedule","NoExecute"]}}

### spec.sandboxes.substrate.workerTolerations[].tolerationSeconds

`int64` · optional (explicit presence)

`tolerationSeconds` represents the period of time the toleration (which
must be of effect NoExecute, otherwise this field is ignored) tolerates the
taint. By default, it is not set, which means tolerate the taint forever
(do not evict). Zero and negative values will be treated as 0 (evict
immediately) by the system.

### spec.sandboxes.substrate.sandboxCpuMillicores

`int32` · optional (explicit presence)

Each sandbox's processor ceiling, in millicores. Empty = 2000.

- rule: {"int32":{"lte":999999,"gte":1}}

### spec.sandboxes.substrate.sandboxMemoryMib

`int32` · optional (explicit presence)

Each sandbox's memory ceiling, in MiB. Empty = 2048.

- rule: {"int32":{"gte":1}}

### spec.sandboxes.substrate.egressHttp

`[]string`

host:port addresses a sandbox may reach over plain HTTP, such as an
in-cluster model gateway (models.ai.svc.cluster.local:8080). Each host is
a DNS name of two or more labels, never an address. The server and the
bucket are added by the operator.

- rule: sandboxes.substrate.egress_http holds host:port addresses whose host is a DNS name of two or more labels, such as models.ai.svc.cluster.local:8080
- rule: {"repeated":{"maxItems":"64","items":{"string":{"maxLen":"259"}}}}

### spec.prerequisites

`KubernetesStigmerPrerequisites`

Which cluster-level operators Stigmer's operator may install when they
are missing. A copy someone else installed is always left alone.

### spec.prerequisites.postgresOperator

`string`

CloudNativePG, which runs the bundled Postgres.

- rule: {"string":{"in":["","auto","skip"]}}

### spec.prerequisites.postgresBackupPlugin

`string`

CloudNativePG's Barman Cloud plugin, which backs the bundled Postgres up
to the bucket. It needs cert-manager, which nothing here installs.

- rule: {"string":{"in":["","auto","skip"]}}

### spec.prerequisites.agentSandbox

`string`

agent-sandbox's controller, which runs the agentSandbox driver's pods.

- rule: {"string":{"in":["","auto","skip"]}}

### spec.prerequisites.substrate

`string`

Agent Substrate's control plane, which runs the substrate driver's
actors.

- rule: {"string":{"in":["","auto","skip"]}}

### spec.upgradeWithoutBackupTo

`string`

Lets one upgrade proceed without the backup the operator takes first.
With backups on, the operator backs the database up before it rolls the
server to a new version, and holds the upgrade while that backup fails.
Set this to the same value as `version` to release that one held upgrade
without a backup (it becomes the `stigmer.ai/upgrade-without-backup`
annotation). It must equal `version`: remove it after the upgrade, so the
next upgrade is backed up again.

- rule: upgrade_without_backup_to must be an exact Stigmer release such as "0.3.2"

## Validation Rules

- `spec.upgrade_without_backup_to.matches_version`: upgrade_without_backup_to releases the upgrade to version without a backup, so it must equal version; remove it once the upgrade is done, so the next one is backed up
- `spec.operator_namespace.not_created_in_stigmer_namespace`: operator.create_namespace creates operator.namespace and deletes it on destroy, and namespace names the same namespace, where the encryption keys live: deleting it would make every backup unreadable. Install the operator in a namespace of its own (stigmer-system by convention), or turn operator.create_namespace off
- `spec.database.external_needs_temporal_external`: database.external names your own Postgres, and Stigmer's operator does not create Temporal's databases on a Postgres it does not run: bring your own Temporal too (temporal.external), or let the operator run the database
- `spec.enterprise.needs_oidc`: Stigmer Enterprise signs every caller in, and a licensed server refuses to boot without sign-in: set identity.oidc
- `spec.enterprise.temporal_must_authenticate`: under a Stigmer Enterprise license, temporal.external must authenticate: whoever reaches a plaintext Temporal can start work on the runner's queue, and a licensed install serves many organizations. Set temporal.external.tls or temporal.external.api_key
- `spec.enterprise.needs_per_session_sandboxes`: a licensed Stigmer runs each session in a sandbox of its own, and refuses to run sessions side by side on the shared runner: set sandboxes.driver to agentSandbox or substrate
- `spec.database.backup_needs_object_store`: the database's backups and restores live in the bucket: declare object_store to use database.bundled.backup or database.bundled.recover_from
- `spec.server.artifacts_without_object_store`: server.artifacts sizes the artifacts volume, and an install with object_store keeps its artifacts in the bucket instead: remove server.artifacts
- `spec.identity.bundled_needs_ingress`: the bundled sign-in serves people's browsers at the front door: declare ingress with the hostname they reach Stigmer at
- `spec.owner.required_with_bundled_sign_in`: the bundled sign-in creates the owner's account first, so an exposed install needs owner.email; or name your own issuer in identity.oidc
- `spec.database.external_without_bundled_sign_in`: the bundled sign-in keeps its accounts on the bundled database, and database.external names your own: sign people in with your own issuer (identity.oidc), or let the operator run the database
- `spec.server.unauthenticated_exposure_without_identity`: server.allow_unauthenticated_exposure exposes the install with sign-in off, and identity turns sign-in on: keep one
- `spec.sandboxes.per_session_needs_object_store`: a sandbox on any node reads and writes a run's files through the bucket, never through a shared disk: declare object_store to run agents in per-session sandboxes

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesStigmer, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | The namespace the Stigmer lives in. |
| `status.outputs.operator_namespace` | `string` | The namespace the operator runs in. |
| `status.outputs.operator_release` | `string` | The operator's Helm release ("stigmer-operator"). |
| `status.outputs.server_endpoint` | `string` | The server's in-cluster gRPC and console address, http://stigmer.<namespace>.svc.cluster.local:7234: what an in-cluster client or a gateway points at. |
| `status.outputs.artifacts_endpoint` | `string` | The server's in-cluster files address (port 7235), for artifacts kept on a volume. |
| `status.outputs.url` | `string` | Where people open Stigmer: https://<hostname> with a front door, or http://localhost:7234 through port_forward_command without one. |
| `status.outputs.port_forward_command` | `string` | The command that opens the server on this machine (ports 7234 and 7235). |
| `status.outputs.status_command` | `string` | The command that shows the install's phase, version and the one sentence that explains it. |
| `status.outputs.owner_password_command` | `string` | The command that reads the owner's one-time password, when the bundled sign-in is on; empty otherwise. |
| `status.outputs.keys_secret` | `string` | The Secret holding the install's encryption keys ("stigmer-keys"). Every backup is sealed with them; the operator keeps the Secret when the Stigmer is deleted. |
| `status.outputs.keys_export_command` | `string` | The command that saves the keys Secret to a file. Keep that file outside the cluster: a restore into a new cluster applies it before the Stigmer is declared. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.operator.namespace` | KubernetesNamespace | `spec.name` |
| `spec.runner.model.existingSecretName` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.database.external.host` | KubernetesPostgres | `status.outputs.rw_service` |
| `spec.database.external.password.existingSecret.name` | KubernetesPostgres | `status.outputs.password_secret.name` |
| `spec.temporal.external.hostPort` | KubernetesTemporal | `status.outputs.frontend_endpoint` |
| `spec.temporal.external.tls.existingSecret.name` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.temporal.external.apiKey.existingSecret.name` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.objectStore.s3.bucket` | AwsS3Bucket | `status.outputs.bucket_id` |
| `spec.objectStore.s3.endpoint` | KubernetesSeaweedFs | `status.outputs.s3_endpoint` |
| `spec.objectStore.s3.credentials.existingSecretName` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.ingress.tls.issuer.name` | KubernetesClusterIssuer | `status.outputs.cluster_issuer_name` |
| `spec.ingress.gateway.name` | KubernetesGateway | `status.outputs.gateway_name` |
| `spec.ingress.gateway.namespace` | KubernetesGateway | `status.outputs.namespace` |
| `spec.identity.bundled.smtp.credentials.existingSecretName` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.enterprise.license.existingSecret.name` | KubernetesSecret | `status.outputs.secret_name` |

## See Also

- [Overview](../README.md)

# KubernetesAgentSandbox

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

KubernetesAgentSandboxSpec defines an agent-sandbox Sandbox: one isolated,
stateful pod with a stable name, its own volumes and, optionally, a headless
Service -- the unit an AI agent's code, a notebook or a dev environment runs
in. The agent-sandbox controller (KubernetesAgentSandboxController) creates
the pod, keeps it running, suspends it on request and tears it down when the
sandbox expires.

100% fidelity with the upstream Sandbox custom resource
(agents.x-k8s.io/v1beta1), pinned to agent-sandbox v1.0.7 -- the release
KubernetesAgentSandboxController installs -- and held to its served
definition by the mirror test. The upstream spec follows the Planton envelope
directly: the pod template, the volume claim templates and `service` (the
blueprint a SandboxTemplate shares), the lifecycle, and the operating mode.

Declare a Sandbox for a sandbox that should exist as long as the declaration
does: an always-on agent, a long-lived dev box, a service an agent hosts.
Sandboxes a product hands out on demand come from a warm pool through a
claim (KubernetesAgentSandboxWarmPool, KubernetesAgentSandboxClaim), usually
created at runtime by agent-sandbox's SDKs.

Isolation is the pod's: set `pod_template.spec.runtime_class_name` to a
sandboxing runtime the cluster offers (`gvisor` on GKE Sandbox nodes, a Kata
class such as `kata-qemu`) and schedule onto those nodes. Port-forwarding
does not reach a gVisor or Kata pod; callers reach it through the
controller's router or the headless Service.

The API server also refuses a Sandbox whose name is longer than 63
characters when `service` is true (the Service is named after it). That rule
reads the object's metadata, which no spec rule can see.

## Example

```yaml
# The full surface of a KubernetesAgentSandbox, offline: every top-level key
# of the Sandbox spec and of the pod specification, every container setting,
# every probe and hook handler, and a volume of each common source -- including
# what kind cannot run (a gVisor runtime class, feature-gated fields). The
# served-schema test renders it and holds it to the pinned definition: nothing
# pruned, nothing refused. The live lanes run e2e/scenarios/.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesAgentSandbox
metadata:
  name: agent-workbench
spec:
  namespace:
    value: agents
  labels:
    team: agents
  annotations:
    example.com/owner: platform
  podTemplate:
    metadata:
      labels:
        app: agent-workbench
      annotations:
        example.com/purpose: coding-agent
    spec:
      runtimeClassName: gvisor
      nodeSelector:
        sandbox.gke.io/runtime: gvisor
      tolerations:
        - key: sandbox.gke.io/runtime
          operator: Equal
          value: gvisor
          effect: NoSchedule
        - key: node.kubernetes.io/not-ready
          operator: Exists
          effect: NoExecute
          tolerationSeconds: 300
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
              - matchExpressions:
                  - key: kubernetes.io/arch
                    operator: In
                    values: ["amd64"]
                matchFields:
                  - key: metadata.name
                    operator: NotIn
                    values: ["drained-node"]
          preferredDuringSchedulingIgnoredDuringExecution:
            - weight: 50
              preference:
                matchExpressions:
                  - key: example.com/cpu-generation
                    operator: Gt
                    values: ["3"]
        podAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
            - weight: 20
              podAffinityTerm:
                topologyKey: topology.kubernetes.io/zone
                labelSelector:
                  matchLabels:
                    app: agent-cache
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            - topologyKey: kubernetes.io/hostname
              labelSelector:
                matchExpressions:
                  - key: app
                    operator: In
                    values: ["agent-workbench"]
              namespaceSelector: {}
              matchLabelKeys: ["pod-template-hash"]
      topologySpreadConstraints:
        - maxSkew: 1
          minDomains: 2
          topologyKey: topology.kubernetes.io/zone
          whenUnsatisfiable: DoNotSchedule
          labelSelector:
            matchLabels:
              app: agent-workbench
          nodeAffinityPolicy: Honor
          nodeTaintsPolicy: Ignore
      priorityClassName:
        value: agent-standard
      priority: 1000
      preemptionPolicy: Never
      schedulerName: default-scheduler
      serviceAccountName:
        value: agent-workbench
      automountServiceAccountToken: false
      enableServiceLinks: false
      shareProcessNamespace: false
      hostname: workbench
      subdomain: agents
      setHostnameAsFQDN: false
      hostAliases:
        - ip: 10.0.0.10
          hostnames: ["internal-registry.example.com"]
      dnsPolicy: None
      dnsConfig:
        nameservers: ["10.96.0.10"]
        searches: ["agents.svc.cluster.local"]
        options:
          - name: ndots
            value: "2"
      imagePullSecrets:
        - name:
            value: registry-credentials
      restartPolicy: Always
      terminationGracePeriodSeconds: 30
      activeDeadlineSeconds: 86400
      readinessGates:
        - conditionType: example.com/agent-ready
      os:
        name: linux
      hostUsers: false
      resources:
        requests:
          cpu: "1"
          memory: 2Gi
        limits:
          memory: 4Gi
      overhead:
        cpu: 250m
        memory: 120Mi
      schedulingGates:
        - name: example.com/quota-approved
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        runAsGroup: 1000
        fsGroup: 1000
        fsGroupChangePolicy: OnRootMismatch
        supplementalGroups: [2000]
        supplementalGroupsPolicy: Strict
        seccompProfile:
          type: RuntimeDefault
        appArmorProfile:
          type: RuntimeDefault
        seLinuxChangePolicy: MountOption
        sysctls:
          - name: net.ipv4.ip_unprivileged_port_start
            value: "0"
      initContainers:
        - name: clone
          image: alpine/git:2.47.2
          args: ["clone", "--depth=1", "https://github.com/example/repo", "/workspace/repo"]
          volumeMounts:
            - name: workspace
              mountPath: /workspace
        - name: log-shipper
          image: fluent/fluent-bit:4.0.3
          restartPolicy: Always
          resources:
            requests:
              cpu: 50m
              memory: 64Mi
      containers:
        - name: agent
          image: us-docker.pkg.dev/example/agents/workbench:1.4.2
          imagePullPolicy: IfNotPresent
          command: ["/usr/local/bin/workbench"]
          args: ["serve", "--port=8888"]
          workingDir: /workspace
          ports:
            - name: http
              containerPort: 8888
              protocol: TCP
            - name: grpc
              containerPort: 9090
          env:
            - name: AGENT_MODE
              value: interactive
            - name: API_KEY
              valueFrom:
                secretKeyRef:
                  name:
                    value: agent-credentials
                  key: api-key
            - name: MODEL
              valueFrom:
                configMapKeyRef:
                  name:
                    value: agent-settings
                  key: model
                  optional: true
            - name: POD_IP
              valueFrom:
                fieldRef:
                  fieldPath: status.podIP
            - name: MEMORY_LIMIT
              valueFrom:
                resourceFieldRef:
                  containerName: agent
                  resource: limits.memory
                  divisor: 1Mi
          envFrom:
            - prefix: AGENT_
              configMapRef:
                name:
                  value: agent-settings
            - secretRef:
                name:
                  value: agent-credentials
                optional: true
          resources:
            requests:
              cpu: 500m
              memory: 1Gi
            limits:
              cpu: "2"
              memory: 3Gi
          resizePolicy:
            - resourceName: cpu
              restartPolicy: NotRequired
            - resourceName: memory
              restartPolicy: RestartContainer
          restartPolicy: Always
          restartPolicyRules:
            - action: Restart
              exitCodes:
                operator: In
                values: [42]
          volumeMounts:
            - name: workspace
              mountPath: /workspace
            - name: tmp
              mountPath: /tmp
            - name: settings
              mountPath: /etc/agent
              readOnly: true
              recursiveReadOnly: IfPossible
            - name: identity
              mountPath: /var/run/identity
              readOnly: true
            - name: tools
              mountPath: /opt/tools
              subPath: bin
              mountPropagation: None
          readinessProbe:
            httpGet:
              path: /healthz
              port: http
              scheme: HTTP
              httpHeaders:
                - name: X-Probe
                  value: readiness
            initialDelaySeconds: 2
            periodSeconds: 5
            timeoutSeconds: 2
            successThreshold: 1
            failureThreshold: 3
          livenessProbe:
            grpc:
              port: 9090
              service: workbench.v1.Health
            periodSeconds: 10
            terminationGracePeriodSeconds: 15
          startupProbe:
            tcpSocket:
              port: "8888"
            failureThreshold: 60
            periodSeconds: 2
          lifecycle:
            postStart:
              exec:
                command: ["/usr/local/bin/workbench", "warm-cache"]
            preStop:
              httpGet:
                path: /drain
                port: "8888"
            stopSignal: SIGTERM
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: FallbackToLogsOnError
          securityContext:
            allowPrivilegeEscalation: false
            readOnlyRootFilesystem: true
            runAsNonRoot: true
            capabilities:
              drop: ["ALL"]
            seccompProfile:
              type: RuntimeDefault
            procMount: Default
          stdin: false
          tty: false
      volumes:
        - name: tmp
          emptyDir:
            medium: Memory
            sizeLimit: 256Mi
        - name: settings
          configMap:
            name:
              value: agent-settings
            defaultMode: 420
            items:
              - key: workbench.yaml
                path: workbench.yaml
        - name: credentials
          secret:
            secretName:
              value: agent-credentials
            defaultMode: 256
        - name: identity
          projected:
            defaultMode: 288
            sources:
              - serviceAccountToken:
                  audience: https://api.example.com
                  expirationSeconds: 3600
                  path: token
              - downwardAPI:
                  items:
                    - path: labels
                      fieldRef:
                        fieldPath: metadata.labels
              - configMap:
                  name:
                    value: agent-settings
                  items:
                    - key: ca.crt
                      path: ca.crt
              - secret:
                  name:
                    value: agent-credentials
                  items:
                    - key: api-key
                      path: api-key
        - name: tools
          image:
            reference: us-docker.pkg.dev/example/agents/tools:2.1.0
            pullPolicy: IfNotPresent
        - name: shared-cache
          persistentVolumeClaim:
            claimName:
              value: agent-cache
            readOnly: true
        - name: scratch
          ephemeral:
            volumeClaimTemplate:
              spec:
                accessModes: ["ReadWriteOnce"]
                storageClassName:
                  value: standard-rwo
                resources:
                  requests:
                    storage: 10Gi
        - name: secrets-store
          csi:
            driver: secrets-store.csi.k8s.io
            readOnly: true
            volumeAttributes:
              secretProviderClass: agent-secrets
        - name: pod-info
          downwardAPI:
            items:
              - path: annotations
                fieldRef:
                  fieldPath: metadata.annotations
  volumeClaimTemplates:
    - metadata:
        name: workspace
        labels:
          example.com/data: workspace
        annotations:
          example.com/backup: daily
      spec:
        accessModes: ["ReadWriteOnce"]
        storageClassName:
          value: standard-rwo
        volumeMode: Filesystem
        resources:
          requests:
            storage: 20Gi
  service: true
  shutdownTime: "2099-01-01T00:00:00Z"
  shutdownPolicy: Retain
  operatingMode: Running
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.annotations` | `map<string, string>` |  |  |  |
| `spec.podTemplate` | `KubernetesAgentSandboxApiPodTemplate` | yes |  |  |
| `spec.podTemplate.spec` | `KubernetesCoreV1ApiPodSpec` | yes |  |  |
| `spec.podTemplate.spec.volumes` | `[]KubernetesCoreV1ApiVolume` |  |  |  |
| `spec.podTemplate.spec.volumes[].name` | `string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].hostPath` | `KubernetesCoreV1ApiHostPathVolumeSource` |  |  |  |
| `spec.podTemplate.spec.volumes[].hostPath.path` | `string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].hostPath.type` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].emptyDir` | `KubernetesCoreV1ApiEmptyDirVolumeSource` |  |  |  |
| `spec.podTemplate.spec.volumes[].emptyDir.medium` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].emptyDir.sizeLimit` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].emptyDir.mode` | `int32` |  |  |  |
| `spec.podTemplate.spec.volumes[].gcePersistentDisk` | `KubernetesCoreV1ApiGCEPersistentDiskVolumeSource` |  |  |  |
| `spec.podTemplate.spec.volumes[].gcePersistentDisk.pdName` | `string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].gcePersistentDisk.fsType` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].gcePersistentDisk.partition` | `int32` |  |  |  |
| `spec.podTemplate.spec.volumes[].gcePersistentDisk.readOnly` | `bool` |  |  |  |
| `spec.podTemplate.spec.volumes[].awsElasticBlockStore` | `KubernetesCoreV1ApiAWSElasticBlockStoreVolumeSource` |  |  |  |
| `spec.podTemplate.spec.volumes[].awsElasticBlockStore.volumeID` | `string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].awsElasticBlockStore.fsType` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].awsElasticBlockStore.partition` | `int32` |  |  |  |
| `spec.podTemplate.spec.volumes[].awsElasticBlockStore.readOnly` | `bool` |  |  |  |
| `spec.podTemplate.spec.volumes[].gitRepo` | `KubernetesCoreV1ApiGitRepoVolumeSource` |  |  |  |
| `spec.podTemplate.spec.volumes[].gitRepo.repository` | `string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].gitRepo.revision` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].gitRepo.directory` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].secret` | `KubernetesCoreV1ApiSecretVolumeSource` |  |  |  |
| `spec.podTemplate.spec.volumes[].secret.secretName` | `string \| valueFrom` |  |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.podTemplate.spec.volumes[].secret.items` | `[]KubernetesCoreV1ApiKeyToPath` |  |  |  |
| `spec.podTemplate.spec.volumes[].secret.items[].key` | `string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].secret.items[].path` | `string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].secret.items[].mode` | `int32` |  |  |  |
| `spec.podTemplate.spec.volumes[].secret.items[].user` | `int64` |  |  |  |
| `spec.podTemplate.spec.volumes[].secret.defaultMode` | `int32` |  |  |  |
| `spec.podTemplate.spec.volumes[].secret.optional` | `bool` |  |  |  |
| `spec.podTemplate.spec.volumes[].secret.defaultUser` | `int64` |  |  |  |
| `spec.podTemplate.spec.volumes[].nfs` | `KubernetesCoreV1ApiNFSVolumeSource` |  |  |  |
| `spec.podTemplate.spec.volumes[].nfs.server` | `string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].nfs.path` | `string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].nfs.readOnly` | `bool` |  |  |  |
| `spec.podTemplate.spec.volumes[].iscsi` | `KubernetesCoreV1ApiISCSIVolumeSource` |  |  |  |
| `spec.podTemplate.spec.volumes[].iscsi.targetPortal` | `string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].iscsi.iqn` | `string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].iscsi.lun` | `int32` | yes |  |  |
| `spec.podTemplate.spec.volumes[].iscsi.iscsiInterface` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].iscsi.fsType` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].iscsi.readOnly` | `bool` |  |  |  |
| `spec.podTemplate.spec.volumes[].iscsi.portals` | `[]string` |  |  |  |
| `spec.podTemplate.spec.volumes[].iscsi.chapAuthDiscovery` | `bool` |  |  |  |
| `spec.podTemplate.spec.volumes[].iscsi.chapAuthSession` | `bool` |  |  |  |
| `spec.podTemplate.spec.volumes[].iscsi.secretRef` | `KubernetesCoreV1ApiLocalObjectReference` |  |  |  |
| `spec.podTemplate.spec.volumes[].iscsi.secretRef.name` | `string \| valueFrom` | yes |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.podTemplate.spec.volumes[].iscsi.initiatorName` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].glusterfs` | `KubernetesCoreV1ApiGlusterfsVolumeSource` |  |  |  |
| `spec.podTemplate.spec.volumes[].glusterfs.endpoints` | `string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].glusterfs.path` | `string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].glusterfs.readOnly` | `bool` |  |  |  |
| `spec.podTemplate.spec.volumes[].persistentVolumeClaim` | `KubernetesCoreV1ApiPersistentVolumeClaimVolumeSource` |  |  |  |
| `spec.podTemplate.spec.volumes[].persistentVolumeClaim.claimName` | `string \| valueFrom` | yes |  | KubernetesPersistentVolumeClaim (`status.outputs.pvc_name`) |
| `spec.podTemplate.spec.volumes[].persistentVolumeClaim.readOnly` | `bool` |  |  |  |
| `spec.podTemplate.spec.volumes[].rbd` | `KubernetesCoreV1ApiRBDVolumeSource` |  |  |  |
| `spec.podTemplate.spec.volumes[].rbd.monitors` | `[]string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].rbd.image` | `string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].rbd.fsType` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].rbd.pool` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].rbd.user` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].rbd.keyring` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].rbd.secretRef` | `KubernetesCoreV1ApiLocalObjectReference` |  |  |  |
| `spec.podTemplate.spec.volumes[].rbd.secretRef.name` | `string \| valueFrom` | yes |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.podTemplate.spec.volumes[].rbd.readOnly` | `bool` |  |  |  |
| `spec.podTemplate.spec.volumes[].flexVolume` | `KubernetesCoreV1ApiFlexVolumeSource` |  |  |  |
| `spec.podTemplate.spec.volumes[].flexVolume.driver` | `string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].flexVolume.fsType` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].flexVolume.secretRef` | `KubernetesCoreV1ApiLocalObjectReference` |  |  |  |
| `spec.podTemplate.spec.volumes[].flexVolume.secretRef.name` | `string \| valueFrom` | yes |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.podTemplate.spec.volumes[].flexVolume.readOnly` | `bool` |  |  |  |
| `spec.podTemplate.spec.volumes[].flexVolume.options` | `map<string, string>` |  |  |  |
| `spec.podTemplate.spec.volumes[].cinder` | `KubernetesCoreV1ApiCinderVolumeSource` |  |  |  |
| `spec.podTemplate.spec.volumes[].cinder.volumeID` | `string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].cinder.fsType` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].cinder.readOnly` | `bool` |  |  |  |
| `spec.podTemplate.spec.volumes[].cinder.secretRef` | `KubernetesCoreV1ApiLocalObjectReference` |  |  |  |
| `spec.podTemplate.spec.volumes[].cinder.secretRef.name` | `string \| valueFrom` | yes |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.podTemplate.spec.volumes[].cephfs` | `KubernetesCoreV1ApiCephFSVolumeSource` |  |  |  |
| `spec.podTemplate.spec.volumes[].cephfs.monitors` | `[]string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].cephfs.path` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].cephfs.user` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].cephfs.secretFile` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].cephfs.secretRef` | `KubernetesCoreV1ApiLocalObjectReference` |  |  |  |
| `spec.podTemplate.spec.volumes[].cephfs.secretRef.name` | `string \| valueFrom` | yes |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.podTemplate.spec.volumes[].cephfs.readOnly` | `bool` |  |  |  |
| `spec.podTemplate.spec.volumes[].flocker` | `KubernetesCoreV1ApiFlockerVolumeSource` |  |  |  |
| `spec.podTemplate.spec.volumes[].flocker.datasetName` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].flocker.datasetUUID` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].downwardAPI` | `KubernetesCoreV1ApiDownwardAPIVolumeSource` |  |  |  |
| `spec.podTemplate.spec.volumes[].downwardAPI.items` | `[]KubernetesCoreV1ApiDownwardAPIVolumeFile` |  |  |  |
| `spec.podTemplate.spec.volumes[].downwardAPI.items[].path` | `string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].downwardAPI.items[].fieldRef` | `KubernetesCoreV1ApiObjectFieldSelector` |  |  |  |
| `spec.podTemplate.spec.volumes[].downwardAPI.items[].fieldRef.apiVersion` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].downwardAPI.items[].fieldRef.fieldPath` | `string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].downwardAPI.items[].resourceFieldRef` | `KubernetesCoreV1ApiResourceFieldSelector` |  |  |  |
| `spec.podTemplate.spec.volumes[].downwardAPI.items[].resourceFieldRef.containerName` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].downwardAPI.items[].resourceFieldRef.resource` | `string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].downwardAPI.items[].resourceFieldRef.divisor` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].downwardAPI.items[].mode` | `int32` |  |  |  |
| `spec.podTemplate.spec.volumes[].downwardAPI.items[].user` | `int64` |  |  |  |
| `spec.podTemplate.spec.volumes[].downwardAPI.defaultMode` | `int32` |  |  |  |
| `spec.podTemplate.spec.volumes[].downwardAPI.defaultUser` | `int64` |  |  |  |
| `spec.podTemplate.spec.volumes[].fc` | `KubernetesCoreV1ApiFCVolumeSource` |  |  |  |
| `spec.podTemplate.spec.volumes[].fc.targetWWNs` | `[]string` |  |  |  |
| `spec.podTemplate.spec.volumes[].fc.lun` | `int32` |  |  |  |
| `spec.podTemplate.spec.volumes[].fc.fsType` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].fc.readOnly` | `bool` |  |  |  |
| `spec.podTemplate.spec.volumes[].fc.wwids` | `[]string` |  |  |  |
| `spec.podTemplate.spec.volumes[].azureFile` | `KubernetesCoreV1ApiAzureFileVolumeSource` |  |  |  |
| `spec.podTemplate.spec.volumes[].azureFile.secretName` | `string \| valueFrom` | yes |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.podTemplate.spec.volumes[].azureFile.shareName` | `string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].azureFile.readOnly` | `bool` |  |  |  |
| `spec.podTemplate.spec.volumes[].configMap` | `KubernetesCoreV1ApiConfigMapVolumeSource` |  |  |  |
| `spec.podTemplate.spec.volumes[].configMap.name` | `string \| valueFrom` | yes |  | KubernetesConfigMap (`status.outputs.configmap_name`) |
| `spec.podTemplate.spec.volumes[].configMap.items` | `[]KubernetesCoreV1ApiKeyToPath` |  |  |  |
| `spec.podTemplate.spec.volumes[].configMap.items[].key` | `string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].configMap.items[].path` | `string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].configMap.items[].mode` | `int32` |  |  |  |
| `spec.podTemplate.spec.volumes[].configMap.items[].user` | `int64` |  |  |  |
| `spec.podTemplate.spec.volumes[].configMap.defaultMode` | `int32` |  |  |  |
| `spec.podTemplate.spec.volumes[].configMap.optional` | `bool` |  |  |  |
| `spec.podTemplate.spec.volumes[].configMap.defaultUser` | `int64` |  |  |  |
| `spec.podTemplate.spec.volumes[].vsphereVolume` | `KubernetesCoreV1ApiVsphereVirtualDiskVolumeSource` |  |  |  |
| `spec.podTemplate.spec.volumes[].vsphereVolume.volumePath` | `string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].vsphereVolume.fsType` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].vsphereVolume.storagePolicyName` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].vsphereVolume.storagePolicyID` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].quobyte` | `KubernetesCoreV1ApiQuobyteVolumeSource` |  |  |  |
| `spec.podTemplate.spec.volumes[].quobyte.registry` | `string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].quobyte.volume` | `string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].quobyte.readOnly` | `bool` |  |  |  |
| `spec.podTemplate.spec.volumes[].quobyte.user` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].quobyte.group` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].quobyte.tenant` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].azureDisk` | `KubernetesCoreV1ApiAzureDiskVolumeSource` |  |  |  |
| `spec.podTemplate.spec.volumes[].azureDisk.diskName` | `string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].azureDisk.diskURI` | `string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].azureDisk.cachingMode` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].azureDisk.fsType` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].azureDisk.readOnly` | `bool` |  |  |  |
| `spec.podTemplate.spec.volumes[].azureDisk.kind` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].photonPersistentDisk` | `KubernetesCoreV1ApiPhotonPersistentDiskVolumeSource` |  |  |  |
| `spec.podTemplate.spec.volumes[].photonPersistentDisk.pdID` | `string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].photonPersistentDisk.fsType` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].projected` | `KubernetesCoreV1ApiProjectedVolumeSource` |  |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources` | `[]KubernetesCoreV1ApiVolumeProjection` |  |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].secret` | `KubernetesCoreV1ApiSecretProjection` |  |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].secret.name` | `string \| valueFrom` | yes |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.podTemplate.spec.volumes[].projected.sources[].secret.items` | `[]KubernetesCoreV1ApiKeyToPath` |  |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].secret.items[].key` | `string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].secret.items[].path` | `string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].secret.items[].mode` | `int32` |  |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].secret.items[].user` | `int64` |  |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].secret.optional` | `bool` |  |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].downwardAPI` | `KubernetesCoreV1ApiDownwardAPIProjection` |  |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].downwardAPI.items` | `[]KubernetesCoreV1ApiDownwardAPIVolumeFile` |  |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].downwardAPI.items[].path` | `string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].downwardAPI.items[].fieldRef` | `KubernetesCoreV1ApiObjectFieldSelector` |  |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].downwardAPI.items[].fieldRef.apiVersion` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].downwardAPI.items[].fieldRef.fieldPath` | `string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].downwardAPI.items[].resourceFieldRef` | `KubernetesCoreV1ApiResourceFieldSelector` |  |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].downwardAPI.items[].resourceFieldRef.containerName` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].downwardAPI.items[].resourceFieldRef.resource` | `string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].downwardAPI.items[].resourceFieldRef.divisor` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].downwardAPI.items[].mode` | `int32` |  |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].downwardAPI.items[].user` | `int64` |  |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].configMap` | `KubernetesCoreV1ApiConfigMapProjection` |  |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].configMap.name` | `string \| valueFrom` | yes |  | KubernetesConfigMap (`status.outputs.configmap_name`) |
| `spec.podTemplate.spec.volumes[].projected.sources[].configMap.items` | `[]KubernetesCoreV1ApiKeyToPath` |  |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].configMap.items[].key` | `string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].configMap.items[].path` | `string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].configMap.items[].mode` | `int32` |  |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].configMap.items[].user` | `int64` |  |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].configMap.optional` | `bool` |  |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].serviceAccountToken` | `KubernetesCoreV1ApiServiceAccountTokenProjection` |  |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].serviceAccountToken.audience` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].serviceAccountToken.expirationSeconds` | `int64` |  |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].serviceAccountToken.path` | `string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].serviceAccountToken.user` | `int64` |  |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].clusterTrustBundle` | `KubernetesCoreV1ApiClusterTrustBundleProjection` |  |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].clusterTrustBundle.name` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].clusterTrustBundle.signerName` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].clusterTrustBundle.labelSelector` | `KubernetesCoreV1ApiLabelSelector` |  |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].clusterTrustBundle.labelSelector.matchLabels` | `map<string, string>` |  |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].clusterTrustBundle.labelSelector.matchExpressions` | `[]KubernetesCoreV1ApiLabelSelectorRequirement` |  |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].clusterTrustBundle.labelSelector.matchExpressions[].key` | `string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].clusterTrustBundle.labelSelector.matchExpressions[].operator` | `string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].clusterTrustBundle.labelSelector.matchExpressions[].values` | `[]string` |  |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].clusterTrustBundle.optional` | `bool` |  |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].clusterTrustBundle.path` | `string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].clusterTrustBundle.user` | `int64` |  |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].podCertificate` | `KubernetesCoreV1ApiPodCertificateProjection` |  |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].podCertificate.signerName` | `string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].podCertificate.keyType` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].podCertificate.maxExpirationSeconds` | `int32` |  |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].podCertificate.credentialBundlePath` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].podCertificate.keyPath` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].podCertificate.certificateChainPath` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].podCertificate.userAnnotations` | `map<string, string>` |  |  |  |
| `spec.podTemplate.spec.volumes[].projected.sources[].podCertificate.user` | `int64` |  |  |  |
| `spec.podTemplate.spec.volumes[].projected.defaultMode` | `int32` |  |  |  |
| `spec.podTemplate.spec.volumes[].projected.defaultUser` | `int64` |  |  |  |
| `spec.podTemplate.spec.volumes[].portworxVolume` | `KubernetesCoreV1ApiPortworxVolumeSource` |  |  |  |
| `spec.podTemplate.spec.volumes[].portworxVolume.volumeID` | `string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].portworxVolume.fsType` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].portworxVolume.readOnly` | `bool` |  |  |  |
| `spec.podTemplate.spec.volumes[].scaleIO` | `KubernetesCoreV1ApiScaleIOVolumeSource` |  |  |  |
| `spec.podTemplate.spec.volumes[].scaleIO.gateway` | `string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].scaleIO.system` | `string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].scaleIO.secretRef` | `KubernetesCoreV1ApiLocalObjectReference` | yes |  |  |
| `spec.podTemplate.spec.volumes[].scaleIO.secretRef.name` | `string \| valueFrom` | yes |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.podTemplate.spec.volumes[].scaleIO.sslEnabled` | `bool` |  |  |  |
| `spec.podTemplate.spec.volumes[].scaleIO.protectionDomain` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].scaleIO.storagePool` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].scaleIO.storageMode` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].scaleIO.volumeName` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].scaleIO.fsType` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].scaleIO.readOnly` | `bool` |  |  |  |
| `spec.podTemplate.spec.volumes[].storageos` | `KubernetesCoreV1ApiStorageOSVolumeSource` |  |  |  |
| `spec.podTemplate.spec.volumes[].storageos.volumeName` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].storageos.volumeNamespace` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].storageos.fsType` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].storageos.readOnly` | `bool` |  |  |  |
| `spec.podTemplate.spec.volumes[].storageos.secretRef` | `KubernetesCoreV1ApiLocalObjectReference` |  |  |  |
| `spec.podTemplate.spec.volumes[].storageos.secretRef.name` | `string \| valueFrom` | yes |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.podTemplate.spec.volumes[].csi` | `KubernetesCoreV1ApiCSIVolumeSource` |  |  |  |
| `spec.podTemplate.spec.volumes[].csi.driver` | `string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].csi.readOnly` | `bool` |  |  |  |
| `spec.podTemplate.spec.volumes[].csi.fsType` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].csi.volumeAttributes` | `map<string, string>` |  |  |  |
| `spec.podTemplate.spec.volumes[].csi.nodePublishSecretRef` | `KubernetesCoreV1ApiLocalObjectReference` |  |  |  |
| `spec.podTemplate.spec.volumes[].csi.nodePublishSecretRef.name` | `string \| valueFrom` | yes |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.podTemplate.spec.volumes[].ephemeral` | `KubernetesCoreV1ApiEphemeralVolumeSource` |  |  |  |
| `spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate` | `KubernetesCoreV1ApiPersistentVolumeClaimTemplate` | yes |  |  |
| `spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.metadata` | `KubernetesCoreV1ApiObjectMeta` |  |  |  |
| `spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.spec` | `KubernetesCoreV1ApiPersistentVolumeClaimSpec` | yes |  |  |
| `spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.spec.accessModes` | `[]string` |  |  |  |
| `spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.spec.selector` | `KubernetesCoreV1ApiLabelSelector` |  |  |  |
| `spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.spec.selector.matchLabels` | `map<string, string>` |  |  |  |
| `spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.spec.selector.matchExpressions` | `[]KubernetesCoreV1ApiLabelSelectorRequirement` |  |  |  |
| `spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.spec.selector.matchExpressions[].key` | `string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.spec.selector.matchExpressions[].operator` | `string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.spec.selector.matchExpressions[].values` | `[]string` |  |  |  |
| `spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.spec.resources` | `KubernetesCoreV1ApiVolumeResourceRequirements` |  |  |  |
| `spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.spec.resources.limits` | `map<string, string>` |  |  |  |
| `spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.spec.resources.requests` | `map<string, string>` |  |  |  |
| `spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.spec.volumeName` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.spec.storageClassName` | `string \| valueFrom` |  |  | KubernetesStorageClass (`status.outputs.storage_class_name`) |
| `spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.spec.volumeMode` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.spec.dataSource` | `KubernetesCoreV1ApiTypedLocalObjectReference` |  |  |  |
| `spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.spec.dataSource.apiGroup` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.spec.dataSource.kind` | `string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.spec.dataSource.name` | `string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.spec.dataSourceRef` | `KubernetesCoreV1ApiTypedObjectReference` |  |  |  |
| `spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.spec.dataSourceRef.apiGroup` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.spec.dataSourceRef.kind` | `string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.spec.dataSourceRef.name` | `string` | yes |  |  |
| `spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.spec.dataSourceRef.namespace` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.spec.volumeAttributesClassName` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].image` | `KubernetesCoreV1ApiImageVolumeSource` |  |  |  |
| `spec.podTemplate.spec.volumes[].image.reference` | `string` |  |  |  |
| `spec.podTemplate.spec.volumes[].image.pullPolicy` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers` | `[]KubernetesCoreV1ApiContainer` |  |  |  |
| `spec.podTemplate.spec.initContainers[].name` | `string` | yes |  |  |
| `spec.podTemplate.spec.initContainers[].image` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].command` | `[]string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].args` | `[]string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].workingDir` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].ports` | `[]KubernetesCoreV1ApiContainerPort` |  |  |  |
| `spec.podTemplate.spec.initContainers[].ports[].name` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].ports[].hostPort` | `int32` |  |  |  |
| `spec.podTemplate.spec.initContainers[].ports[].containerPort` | `int32` | yes |  |  |
| `spec.podTemplate.spec.initContainers[].ports[].protocol` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].ports[].hostIP` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].envFrom` | `[]KubernetesCoreV1ApiEnvFromSource` |  |  |  |
| `spec.podTemplate.spec.initContainers[].envFrom[].prefix` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].envFrom[].configMapRef` | `KubernetesCoreV1ApiConfigMapEnvSource` |  |  |  |
| `spec.podTemplate.spec.initContainers[].envFrom[].configMapRef.name` | `string \| valueFrom` | yes |  | KubernetesConfigMap (`status.outputs.configmap_name`) |
| `spec.podTemplate.spec.initContainers[].envFrom[].configMapRef.optional` | `bool` |  |  |  |
| `spec.podTemplate.spec.initContainers[].envFrom[].secretRef` | `KubernetesCoreV1ApiSecretEnvSource` |  |  |  |
| `spec.podTemplate.spec.initContainers[].envFrom[].secretRef.name` | `string \| valueFrom` | yes |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.podTemplate.spec.initContainers[].envFrom[].secretRef.optional` | `bool` |  |  |  |
| `spec.podTemplate.spec.initContainers[].env` | `[]KubernetesCoreV1ApiEnvVar` |  |  |  |
| `spec.podTemplate.spec.initContainers[].env[].name` | `string` | yes |  |  |
| `spec.podTemplate.spec.initContainers[].env[].value` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].env[].valueFrom` | `KubernetesCoreV1ApiEnvVarSource` |  |  |  |
| `spec.podTemplate.spec.initContainers[].env[].valueFrom.fieldRef` | `KubernetesCoreV1ApiObjectFieldSelector` |  |  |  |
| `spec.podTemplate.spec.initContainers[].env[].valueFrom.fieldRef.apiVersion` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].env[].valueFrom.fieldRef.fieldPath` | `string` | yes |  |  |
| `spec.podTemplate.spec.initContainers[].env[].valueFrom.resourceFieldRef` | `KubernetesCoreV1ApiResourceFieldSelector` |  |  |  |
| `spec.podTemplate.spec.initContainers[].env[].valueFrom.resourceFieldRef.containerName` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].env[].valueFrom.resourceFieldRef.resource` | `string` | yes |  |  |
| `spec.podTemplate.spec.initContainers[].env[].valueFrom.resourceFieldRef.divisor` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].env[].valueFrom.configMapKeyRef` | `KubernetesCoreV1ApiConfigMapKeySelector` |  |  |  |
| `spec.podTemplate.spec.initContainers[].env[].valueFrom.configMapKeyRef.name` | `string \| valueFrom` | yes |  | KubernetesConfigMap (`status.outputs.configmap_name`) |
| `spec.podTemplate.spec.initContainers[].env[].valueFrom.configMapKeyRef.key` | `string` | yes |  |  |
| `spec.podTemplate.spec.initContainers[].env[].valueFrom.configMapKeyRef.optional` | `bool` |  |  |  |
| `spec.podTemplate.spec.initContainers[].env[].valueFrom.secretKeyRef` | `KubernetesCoreV1ApiSecretKeySelector` |  |  |  |
| `spec.podTemplate.spec.initContainers[].env[].valueFrom.secretKeyRef.name` | `string \| valueFrom` | yes |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.podTemplate.spec.initContainers[].env[].valueFrom.secretKeyRef.key` | `string` | yes |  |  |
| `spec.podTemplate.spec.initContainers[].env[].valueFrom.secretKeyRef.optional` | `bool` |  |  |  |
| `spec.podTemplate.spec.initContainers[].env[].valueFrom.fileKeyRef` | `KubernetesCoreV1ApiFileKeySelector` |  |  |  |
| `spec.podTemplate.spec.initContainers[].env[].valueFrom.fileKeyRef.volumeName` | `string` | yes |  |  |
| `spec.podTemplate.spec.initContainers[].env[].valueFrom.fileKeyRef.path` | `string` | yes |  |  |
| `spec.podTemplate.spec.initContainers[].env[].valueFrom.fileKeyRef.key` | `string` | yes |  |  |
| `spec.podTemplate.spec.initContainers[].env[].valueFrom.fileKeyRef.optional` | `bool` |  |  |  |
| `spec.podTemplate.spec.initContainers[].resources` | `KubernetesCoreV1ApiResourceRequirements` |  |  |  |
| `spec.podTemplate.spec.initContainers[].resources.limits` | `map<string, string>` |  |  |  |
| `spec.podTemplate.spec.initContainers[].resources.requests` | `map<string, string>` |  |  |  |
| `spec.podTemplate.spec.initContainers[].resources.claims` | `[]KubernetesCoreV1ApiResourceClaim` |  |  |  |
| `spec.podTemplate.spec.initContainers[].resources.claims[].name` | `string` | yes |  |  |
| `spec.podTemplate.spec.initContainers[].resources.claims[].request` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].resizePolicy` | `[]KubernetesCoreV1ApiContainerResizePolicy` |  |  |  |
| `spec.podTemplate.spec.initContainers[].resizePolicy[].resourceName` | `string` | yes |  |  |
| `spec.podTemplate.spec.initContainers[].resizePolicy[].restartPolicy` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].restartPolicy` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].restartPolicyRules` | `[]KubernetesCoreV1ApiContainerRestartRule` |  |  |  |
| `spec.podTemplate.spec.initContainers[].restartPolicyRules[].action` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].restartPolicyRules[].exitCodes` | `KubernetesCoreV1ApiContainerRestartRuleOnExitCodes` |  |  |  |
| `spec.podTemplate.spec.initContainers[].restartPolicyRules[].exitCodes.operator` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].restartPolicyRules[].exitCodes.values` | `[]int32` |  |  |  |
| `spec.podTemplate.spec.initContainers[].volumeMounts` | `[]KubernetesCoreV1ApiVolumeMount` |  |  |  |
| `spec.podTemplate.spec.initContainers[].volumeMounts[].name` | `string` | yes |  |  |
| `spec.podTemplate.spec.initContainers[].volumeMounts[].readOnly` | `bool` |  |  |  |
| `spec.podTemplate.spec.initContainers[].volumeMounts[].recursiveReadOnly` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].volumeMounts[].mountPath` | `string` | yes |  |  |
| `spec.podTemplate.spec.initContainers[].volumeMounts[].subPath` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].volumeMounts[].mountPropagation` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].volumeMounts[].subPathExpr` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].volumeMounts[].bindMountOptions` | `[]string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].volumeDevices` | `[]KubernetesCoreV1ApiVolumeDevice` |  |  |  |
| `spec.podTemplate.spec.initContainers[].volumeDevices[].name` | `string` | yes |  |  |
| `spec.podTemplate.spec.initContainers[].volumeDevices[].devicePath` | `string` | yes |  |  |
| `spec.podTemplate.spec.initContainers[].livenessProbe` | `KubernetesCoreV1ApiProbe` |  |  |  |
| `spec.podTemplate.spec.initContainers[].livenessProbe.exec` | `KubernetesCoreV1ApiExecAction` |  |  |  |
| `spec.podTemplate.spec.initContainers[].livenessProbe.exec.command` | `[]string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].livenessProbe.httpGet` | `KubernetesCoreV1ApiHTTPGetAction` |  |  |  |
| `spec.podTemplate.spec.initContainers[].livenessProbe.httpGet.path` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].livenessProbe.httpGet.port` | `string` | yes |  |  |
| `spec.podTemplate.spec.initContainers[].livenessProbe.httpGet.host` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].livenessProbe.httpGet.scheme` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].livenessProbe.httpGet.httpHeaders` | `[]KubernetesCoreV1ApiHTTPHeader` |  |  |  |
| `spec.podTemplate.spec.initContainers[].livenessProbe.httpGet.httpHeaders[].name` | `string` | yes |  |  |
| `spec.podTemplate.spec.initContainers[].livenessProbe.httpGet.httpHeaders[].value` | `string` | yes |  |  |
| `spec.podTemplate.spec.initContainers[].livenessProbe.httpGet.protocol` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].livenessProbe.tcpSocket` | `KubernetesCoreV1ApiTCPSocketAction` |  |  |  |
| `spec.podTemplate.spec.initContainers[].livenessProbe.tcpSocket.port` | `string` | yes |  |  |
| `spec.podTemplate.spec.initContainers[].livenessProbe.tcpSocket.host` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].livenessProbe.grpc` | `KubernetesCoreV1ApiGRPCAction` |  |  |  |
| `spec.podTemplate.spec.initContainers[].livenessProbe.grpc.port` | `int32` | yes |  |  |
| `spec.podTemplate.spec.initContainers[].livenessProbe.grpc.service` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].livenessProbe.grpc.mode` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].livenessProbe.initialDelaySeconds` | `int32` |  |  |  |
| `spec.podTemplate.spec.initContainers[].livenessProbe.timeoutSeconds` | `int32` |  |  |  |
| `spec.podTemplate.spec.initContainers[].livenessProbe.periodSeconds` | `int32` |  |  |  |
| `spec.podTemplate.spec.initContainers[].livenessProbe.successThreshold` | `int32` |  |  |  |
| `spec.podTemplate.spec.initContainers[].livenessProbe.failureThreshold` | `int32` |  |  |  |
| `spec.podTemplate.spec.initContainers[].livenessProbe.terminationGracePeriodSeconds` | `int64` |  |  |  |
| `spec.podTemplate.spec.initContainers[].readinessProbe` | `KubernetesCoreV1ApiProbe` |  |  |  |
| `spec.podTemplate.spec.initContainers[].readinessProbe.exec` | `KubernetesCoreV1ApiExecAction` |  |  |  |
| `spec.podTemplate.spec.initContainers[].readinessProbe.exec.command` | `[]string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].readinessProbe.httpGet` | `KubernetesCoreV1ApiHTTPGetAction` |  |  |  |
| `spec.podTemplate.spec.initContainers[].readinessProbe.httpGet.path` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].readinessProbe.httpGet.port` | `string` | yes |  |  |
| `spec.podTemplate.spec.initContainers[].readinessProbe.httpGet.host` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].readinessProbe.httpGet.scheme` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].readinessProbe.httpGet.httpHeaders` | `[]KubernetesCoreV1ApiHTTPHeader` |  |  |  |
| `spec.podTemplate.spec.initContainers[].readinessProbe.httpGet.httpHeaders[].name` | `string` | yes |  |  |
| `spec.podTemplate.spec.initContainers[].readinessProbe.httpGet.httpHeaders[].value` | `string` | yes |  |  |
| `spec.podTemplate.spec.initContainers[].readinessProbe.httpGet.protocol` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].readinessProbe.tcpSocket` | `KubernetesCoreV1ApiTCPSocketAction` |  |  |  |
| `spec.podTemplate.spec.initContainers[].readinessProbe.tcpSocket.port` | `string` | yes |  |  |
| `spec.podTemplate.spec.initContainers[].readinessProbe.tcpSocket.host` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].readinessProbe.grpc` | `KubernetesCoreV1ApiGRPCAction` |  |  |  |
| `spec.podTemplate.spec.initContainers[].readinessProbe.grpc.port` | `int32` | yes |  |  |
| `spec.podTemplate.spec.initContainers[].readinessProbe.grpc.service` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].readinessProbe.grpc.mode` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].readinessProbe.initialDelaySeconds` | `int32` |  |  |  |
| `spec.podTemplate.spec.initContainers[].readinessProbe.timeoutSeconds` | `int32` |  |  |  |
| `spec.podTemplate.spec.initContainers[].readinessProbe.periodSeconds` | `int32` |  |  |  |
| `spec.podTemplate.spec.initContainers[].readinessProbe.successThreshold` | `int32` |  |  |  |
| `spec.podTemplate.spec.initContainers[].readinessProbe.failureThreshold` | `int32` |  |  |  |
| `spec.podTemplate.spec.initContainers[].readinessProbe.terminationGracePeriodSeconds` | `int64` |  |  |  |
| `spec.podTemplate.spec.initContainers[].startupProbe` | `KubernetesCoreV1ApiProbe` |  |  |  |
| `spec.podTemplate.spec.initContainers[].startupProbe.exec` | `KubernetesCoreV1ApiExecAction` |  |  |  |
| `spec.podTemplate.spec.initContainers[].startupProbe.exec.command` | `[]string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].startupProbe.httpGet` | `KubernetesCoreV1ApiHTTPGetAction` |  |  |  |
| `spec.podTemplate.spec.initContainers[].startupProbe.httpGet.path` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].startupProbe.httpGet.port` | `string` | yes |  |  |
| `spec.podTemplate.spec.initContainers[].startupProbe.httpGet.host` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].startupProbe.httpGet.scheme` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].startupProbe.httpGet.httpHeaders` | `[]KubernetesCoreV1ApiHTTPHeader` |  |  |  |
| `spec.podTemplate.spec.initContainers[].startupProbe.httpGet.httpHeaders[].name` | `string` | yes |  |  |
| `spec.podTemplate.spec.initContainers[].startupProbe.httpGet.httpHeaders[].value` | `string` | yes |  |  |
| `spec.podTemplate.spec.initContainers[].startupProbe.httpGet.protocol` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].startupProbe.tcpSocket` | `KubernetesCoreV1ApiTCPSocketAction` |  |  |  |
| `spec.podTemplate.spec.initContainers[].startupProbe.tcpSocket.port` | `string` | yes |  |  |
| `spec.podTemplate.spec.initContainers[].startupProbe.tcpSocket.host` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].startupProbe.grpc` | `KubernetesCoreV1ApiGRPCAction` |  |  |  |
| `spec.podTemplate.spec.initContainers[].startupProbe.grpc.port` | `int32` | yes |  |  |
| `spec.podTemplate.spec.initContainers[].startupProbe.grpc.service` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].startupProbe.grpc.mode` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].startupProbe.initialDelaySeconds` | `int32` |  |  |  |
| `spec.podTemplate.spec.initContainers[].startupProbe.timeoutSeconds` | `int32` |  |  |  |
| `spec.podTemplate.spec.initContainers[].startupProbe.periodSeconds` | `int32` |  |  |  |
| `spec.podTemplate.spec.initContainers[].startupProbe.successThreshold` | `int32` |  |  |  |
| `spec.podTemplate.spec.initContainers[].startupProbe.failureThreshold` | `int32` |  |  |  |
| `spec.podTemplate.spec.initContainers[].startupProbe.terminationGracePeriodSeconds` | `int64` |  |  |  |
| `spec.podTemplate.spec.initContainers[].lifecycle` | `KubernetesCoreV1ApiLifecycle` |  |  |  |
| `spec.podTemplate.spec.initContainers[].lifecycle.postStart` | `KubernetesCoreV1ApiLifecycleHandler` |  |  |  |
| `spec.podTemplate.spec.initContainers[].lifecycle.postStart.exec` | `KubernetesCoreV1ApiExecAction` |  |  |  |
| `spec.podTemplate.spec.initContainers[].lifecycle.postStart.exec.command` | `[]string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].lifecycle.postStart.httpGet` | `KubernetesCoreV1ApiHTTPGetAction` |  |  |  |
| `spec.podTemplate.spec.initContainers[].lifecycle.postStart.httpGet.path` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].lifecycle.postStart.httpGet.port` | `string` | yes |  |  |
| `spec.podTemplate.spec.initContainers[].lifecycle.postStart.httpGet.host` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].lifecycle.postStart.httpGet.scheme` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].lifecycle.postStart.httpGet.httpHeaders` | `[]KubernetesCoreV1ApiHTTPHeader` |  |  |  |
| `spec.podTemplate.spec.initContainers[].lifecycle.postStart.httpGet.httpHeaders[].name` | `string` | yes |  |  |
| `spec.podTemplate.spec.initContainers[].lifecycle.postStart.httpGet.httpHeaders[].value` | `string` | yes |  |  |
| `spec.podTemplate.spec.initContainers[].lifecycle.postStart.httpGet.protocol` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].lifecycle.postStart.tcpSocket` | `KubernetesCoreV1ApiTCPSocketAction` |  |  |  |
| `spec.podTemplate.spec.initContainers[].lifecycle.postStart.tcpSocket.port` | `string` | yes |  |  |
| `spec.podTemplate.spec.initContainers[].lifecycle.postStart.tcpSocket.host` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].lifecycle.postStart.sleep` | `KubernetesCoreV1ApiSleepAction` |  |  |  |
| `spec.podTemplate.spec.initContainers[].lifecycle.postStart.sleep.seconds` | `int64` | yes |  |  |
| `spec.podTemplate.spec.initContainers[].lifecycle.preStop` | `KubernetesCoreV1ApiLifecycleHandler` |  |  |  |
| `spec.podTemplate.spec.initContainers[].lifecycle.preStop.exec` | `KubernetesCoreV1ApiExecAction` |  |  |  |
| `spec.podTemplate.spec.initContainers[].lifecycle.preStop.exec.command` | `[]string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].lifecycle.preStop.httpGet` | `KubernetesCoreV1ApiHTTPGetAction` |  |  |  |
| `spec.podTemplate.spec.initContainers[].lifecycle.preStop.httpGet.path` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].lifecycle.preStop.httpGet.port` | `string` | yes |  |  |
| `spec.podTemplate.spec.initContainers[].lifecycle.preStop.httpGet.host` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].lifecycle.preStop.httpGet.scheme` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].lifecycle.preStop.httpGet.httpHeaders` | `[]KubernetesCoreV1ApiHTTPHeader` |  |  |  |
| `spec.podTemplate.spec.initContainers[].lifecycle.preStop.httpGet.httpHeaders[].name` | `string` | yes |  |  |
| `spec.podTemplate.spec.initContainers[].lifecycle.preStop.httpGet.httpHeaders[].value` | `string` | yes |  |  |
| `spec.podTemplate.spec.initContainers[].lifecycle.preStop.httpGet.protocol` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].lifecycle.preStop.tcpSocket` | `KubernetesCoreV1ApiTCPSocketAction` |  |  |  |
| `spec.podTemplate.spec.initContainers[].lifecycle.preStop.tcpSocket.port` | `string` | yes |  |  |
| `spec.podTemplate.spec.initContainers[].lifecycle.preStop.tcpSocket.host` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].lifecycle.preStop.sleep` | `KubernetesCoreV1ApiSleepAction` |  |  |  |
| `spec.podTemplate.spec.initContainers[].lifecycle.preStop.sleep.seconds` | `int64` | yes |  |  |
| `spec.podTemplate.spec.initContainers[].lifecycle.stopSignal` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].terminationMessagePath` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].terminationMessagePolicy` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].imagePullPolicy` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].securityContext` | `KubernetesCoreV1ApiSecurityContext` |  |  |  |
| `spec.podTemplate.spec.initContainers[].securityContext.capabilities` | `KubernetesCoreV1ApiCapabilities` |  |  |  |
| `spec.podTemplate.spec.initContainers[].securityContext.capabilities.add` | `[]string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].securityContext.capabilities.drop` | `[]string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].securityContext.privileged` | `bool` |  |  |  |
| `spec.podTemplate.spec.initContainers[].securityContext.seLinuxOptions` | `KubernetesCoreV1ApiSELinuxOptions` |  |  |  |
| `spec.podTemplate.spec.initContainers[].securityContext.seLinuxOptions.user` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].securityContext.seLinuxOptions.role` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].securityContext.seLinuxOptions.type` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].securityContext.seLinuxOptions.level` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].securityContext.windowsOptions` | `KubernetesCoreV1ApiWindowsSecurityContextOptions` |  |  |  |
| `spec.podTemplate.spec.initContainers[].securityContext.windowsOptions.gmsaCredentialSpecName` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].securityContext.windowsOptions.gmsaCredentialSpec` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].securityContext.windowsOptions.runAsUserName` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].securityContext.windowsOptions.hostProcess` | `bool` |  |  |  |
| `spec.podTemplate.spec.initContainers[].securityContext.runAsUser` | `int64` |  |  |  |
| `spec.podTemplate.spec.initContainers[].securityContext.runAsGroup` | `int64` |  |  |  |
| `spec.podTemplate.spec.initContainers[].securityContext.runAsNonRoot` | `bool` |  |  |  |
| `spec.podTemplate.spec.initContainers[].securityContext.readOnlyRootFilesystem` | `bool` |  |  |  |
| `spec.podTemplate.spec.initContainers[].securityContext.allowPrivilegeEscalation` | `bool` |  |  |  |
| `spec.podTemplate.spec.initContainers[].securityContext.procMount` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].securityContext.seccompProfile` | `KubernetesCoreV1ApiSeccompProfile` |  |  |  |
| `spec.podTemplate.spec.initContainers[].securityContext.seccompProfile.type` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].securityContext.seccompProfile.localhostProfile` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].securityContext.appArmorProfile` | `KubernetesCoreV1ApiAppArmorProfile` |  |  |  |
| `spec.podTemplate.spec.initContainers[].securityContext.appArmorProfile.type` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].securityContext.appArmorProfile.localhostProfile` | `string` |  |  |  |
| `spec.podTemplate.spec.initContainers[].stdin` | `bool` |  |  |  |
| `spec.podTemplate.spec.initContainers[].stdinOnce` | `bool` |  |  |  |
| `spec.podTemplate.spec.initContainers[].tty` | `bool` |  |  |  |
| `spec.podTemplate.spec.containers` | `[]KubernetesCoreV1ApiContainer` | yes |  |  |
| `spec.podTemplate.spec.containers[].name` | `string` | yes |  |  |
| `spec.podTemplate.spec.containers[].image` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].command` | `[]string` |  |  |  |
| `spec.podTemplate.spec.containers[].args` | `[]string` |  |  |  |
| `spec.podTemplate.spec.containers[].workingDir` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].ports` | `[]KubernetesCoreV1ApiContainerPort` |  |  |  |
| `spec.podTemplate.spec.containers[].ports[].name` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].ports[].hostPort` | `int32` |  |  |  |
| `spec.podTemplate.spec.containers[].ports[].containerPort` | `int32` | yes |  |  |
| `spec.podTemplate.spec.containers[].ports[].protocol` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].ports[].hostIP` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].envFrom` | `[]KubernetesCoreV1ApiEnvFromSource` |  |  |  |
| `spec.podTemplate.spec.containers[].envFrom[].prefix` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].envFrom[].configMapRef` | `KubernetesCoreV1ApiConfigMapEnvSource` |  |  |  |
| `spec.podTemplate.spec.containers[].envFrom[].configMapRef.name` | `string \| valueFrom` | yes |  | KubernetesConfigMap (`status.outputs.configmap_name`) |
| `spec.podTemplate.spec.containers[].envFrom[].configMapRef.optional` | `bool` |  |  |  |
| `spec.podTemplate.spec.containers[].envFrom[].secretRef` | `KubernetesCoreV1ApiSecretEnvSource` |  |  |  |
| `spec.podTemplate.spec.containers[].envFrom[].secretRef.name` | `string \| valueFrom` | yes |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.podTemplate.spec.containers[].envFrom[].secretRef.optional` | `bool` |  |  |  |
| `spec.podTemplate.spec.containers[].env` | `[]KubernetesCoreV1ApiEnvVar` |  |  |  |
| `spec.podTemplate.spec.containers[].env[].name` | `string` | yes |  |  |
| `spec.podTemplate.spec.containers[].env[].value` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].env[].valueFrom` | `KubernetesCoreV1ApiEnvVarSource` |  |  |  |
| `spec.podTemplate.spec.containers[].env[].valueFrom.fieldRef` | `KubernetesCoreV1ApiObjectFieldSelector` |  |  |  |
| `spec.podTemplate.spec.containers[].env[].valueFrom.fieldRef.apiVersion` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].env[].valueFrom.fieldRef.fieldPath` | `string` | yes |  |  |
| `spec.podTemplate.spec.containers[].env[].valueFrom.resourceFieldRef` | `KubernetesCoreV1ApiResourceFieldSelector` |  |  |  |
| `spec.podTemplate.spec.containers[].env[].valueFrom.resourceFieldRef.containerName` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].env[].valueFrom.resourceFieldRef.resource` | `string` | yes |  |  |
| `spec.podTemplate.spec.containers[].env[].valueFrom.resourceFieldRef.divisor` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].env[].valueFrom.configMapKeyRef` | `KubernetesCoreV1ApiConfigMapKeySelector` |  |  |  |
| `spec.podTemplate.spec.containers[].env[].valueFrom.configMapKeyRef.name` | `string \| valueFrom` | yes |  | KubernetesConfigMap (`status.outputs.configmap_name`) |
| `spec.podTemplate.spec.containers[].env[].valueFrom.configMapKeyRef.key` | `string` | yes |  |  |
| `spec.podTemplate.spec.containers[].env[].valueFrom.configMapKeyRef.optional` | `bool` |  |  |  |
| `spec.podTemplate.spec.containers[].env[].valueFrom.secretKeyRef` | `KubernetesCoreV1ApiSecretKeySelector` |  |  |  |
| `spec.podTemplate.spec.containers[].env[].valueFrom.secretKeyRef.name` | `string \| valueFrom` | yes |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.podTemplate.spec.containers[].env[].valueFrom.secretKeyRef.key` | `string` | yes |  |  |
| `spec.podTemplate.spec.containers[].env[].valueFrom.secretKeyRef.optional` | `bool` |  |  |  |
| `spec.podTemplate.spec.containers[].env[].valueFrom.fileKeyRef` | `KubernetesCoreV1ApiFileKeySelector` |  |  |  |
| `spec.podTemplate.spec.containers[].env[].valueFrom.fileKeyRef.volumeName` | `string` | yes |  |  |
| `spec.podTemplate.spec.containers[].env[].valueFrom.fileKeyRef.path` | `string` | yes |  |  |
| `spec.podTemplate.spec.containers[].env[].valueFrom.fileKeyRef.key` | `string` | yes |  |  |
| `spec.podTemplate.spec.containers[].env[].valueFrom.fileKeyRef.optional` | `bool` |  |  |  |
| `spec.podTemplate.spec.containers[].resources` | `KubernetesCoreV1ApiResourceRequirements` |  |  |  |
| `spec.podTemplate.spec.containers[].resources.limits` | `map<string, string>` |  |  |  |
| `spec.podTemplate.spec.containers[].resources.requests` | `map<string, string>` |  |  |  |
| `spec.podTemplate.spec.containers[].resources.claims` | `[]KubernetesCoreV1ApiResourceClaim` |  |  |  |
| `spec.podTemplate.spec.containers[].resources.claims[].name` | `string` | yes |  |  |
| `spec.podTemplate.spec.containers[].resources.claims[].request` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].resizePolicy` | `[]KubernetesCoreV1ApiContainerResizePolicy` |  |  |  |
| `spec.podTemplate.spec.containers[].resizePolicy[].resourceName` | `string` | yes |  |  |
| `spec.podTemplate.spec.containers[].resizePolicy[].restartPolicy` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].restartPolicy` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].restartPolicyRules` | `[]KubernetesCoreV1ApiContainerRestartRule` |  |  |  |
| `spec.podTemplate.spec.containers[].restartPolicyRules[].action` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].restartPolicyRules[].exitCodes` | `KubernetesCoreV1ApiContainerRestartRuleOnExitCodes` |  |  |  |
| `spec.podTemplate.spec.containers[].restartPolicyRules[].exitCodes.operator` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].restartPolicyRules[].exitCodes.values` | `[]int32` |  |  |  |
| `spec.podTemplate.spec.containers[].volumeMounts` | `[]KubernetesCoreV1ApiVolumeMount` |  |  |  |
| `spec.podTemplate.spec.containers[].volumeMounts[].name` | `string` | yes |  |  |
| `spec.podTemplate.spec.containers[].volumeMounts[].readOnly` | `bool` |  |  |  |
| `spec.podTemplate.spec.containers[].volumeMounts[].recursiveReadOnly` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].volumeMounts[].mountPath` | `string` | yes |  |  |
| `spec.podTemplate.spec.containers[].volumeMounts[].subPath` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].volumeMounts[].mountPropagation` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].volumeMounts[].subPathExpr` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].volumeMounts[].bindMountOptions` | `[]string` |  |  |  |
| `spec.podTemplate.spec.containers[].volumeDevices` | `[]KubernetesCoreV1ApiVolumeDevice` |  |  |  |
| `spec.podTemplate.spec.containers[].volumeDevices[].name` | `string` | yes |  |  |
| `spec.podTemplate.spec.containers[].volumeDevices[].devicePath` | `string` | yes |  |  |
| `spec.podTemplate.spec.containers[].livenessProbe` | `KubernetesCoreV1ApiProbe` |  |  |  |
| `spec.podTemplate.spec.containers[].livenessProbe.exec` | `KubernetesCoreV1ApiExecAction` |  |  |  |
| `spec.podTemplate.spec.containers[].livenessProbe.exec.command` | `[]string` |  |  |  |
| `spec.podTemplate.spec.containers[].livenessProbe.httpGet` | `KubernetesCoreV1ApiHTTPGetAction` |  |  |  |
| `spec.podTemplate.spec.containers[].livenessProbe.httpGet.path` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].livenessProbe.httpGet.port` | `string` | yes |  |  |
| `spec.podTemplate.spec.containers[].livenessProbe.httpGet.host` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].livenessProbe.httpGet.scheme` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].livenessProbe.httpGet.httpHeaders` | `[]KubernetesCoreV1ApiHTTPHeader` |  |  |  |
| `spec.podTemplate.spec.containers[].livenessProbe.httpGet.httpHeaders[].name` | `string` | yes |  |  |
| `spec.podTemplate.spec.containers[].livenessProbe.httpGet.httpHeaders[].value` | `string` | yes |  |  |
| `spec.podTemplate.spec.containers[].livenessProbe.httpGet.protocol` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].livenessProbe.tcpSocket` | `KubernetesCoreV1ApiTCPSocketAction` |  |  |  |
| `spec.podTemplate.spec.containers[].livenessProbe.tcpSocket.port` | `string` | yes |  |  |
| `spec.podTemplate.spec.containers[].livenessProbe.tcpSocket.host` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].livenessProbe.grpc` | `KubernetesCoreV1ApiGRPCAction` |  |  |  |
| `spec.podTemplate.spec.containers[].livenessProbe.grpc.port` | `int32` | yes |  |  |
| `spec.podTemplate.spec.containers[].livenessProbe.grpc.service` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].livenessProbe.grpc.mode` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].livenessProbe.initialDelaySeconds` | `int32` |  |  |  |
| `spec.podTemplate.spec.containers[].livenessProbe.timeoutSeconds` | `int32` |  |  |  |
| `spec.podTemplate.spec.containers[].livenessProbe.periodSeconds` | `int32` |  |  |  |
| `spec.podTemplate.spec.containers[].livenessProbe.successThreshold` | `int32` |  |  |  |
| `spec.podTemplate.spec.containers[].livenessProbe.failureThreshold` | `int32` |  |  |  |
| `spec.podTemplate.spec.containers[].livenessProbe.terminationGracePeriodSeconds` | `int64` |  |  |  |
| `spec.podTemplate.spec.containers[].readinessProbe` | `KubernetesCoreV1ApiProbe` |  |  |  |
| `spec.podTemplate.spec.containers[].readinessProbe.exec` | `KubernetesCoreV1ApiExecAction` |  |  |  |
| `spec.podTemplate.spec.containers[].readinessProbe.exec.command` | `[]string` |  |  |  |
| `spec.podTemplate.spec.containers[].readinessProbe.httpGet` | `KubernetesCoreV1ApiHTTPGetAction` |  |  |  |
| `spec.podTemplate.spec.containers[].readinessProbe.httpGet.path` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].readinessProbe.httpGet.port` | `string` | yes |  |  |
| `spec.podTemplate.spec.containers[].readinessProbe.httpGet.host` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].readinessProbe.httpGet.scheme` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].readinessProbe.httpGet.httpHeaders` | `[]KubernetesCoreV1ApiHTTPHeader` |  |  |  |
| `spec.podTemplate.spec.containers[].readinessProbe.httpGet.httpHeaders[].name` | `string` | yes |  |  |
| `spec.podTemplate.spec.containers[].readinessProbe.httpGet.httpHeaders[].value` | `string` | yes |  |  |
| `spec.podTemplate.spec.containers[].readinessProbe.httpGet.protocol` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].readinessProbe.tcpSocket` | `KubernetesCoreV1ApiTCPSocketAction` |  |  |  |
| `spec.podTemplate.spec.containers[].readinessProbe.tcpSocket.port` | `string` | yes |  |  |
| `spec.podTemplate.spec.containers[].readinessProbe.tcpSocket.host` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].readinessProbe.grpc` | `KubernetesCoreV1ApiGRPCAction` |  |  |  |
| `spec.podTemplate.spec.containers[].readinessProbe.grpc.port` | `int32` | yes |  |  |
| `spec.podTemplate.spec.containers[].readinessProbe.grpc.service` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].readinessProbe.grpc.mode` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].readinessProbe.initialDelaySeconds` | `int32` |  |  |  |
| `spec.podTemplate.spec.containers[].readinessProbe.timeoutSeconds` | `int32` |  |  |  |
| `spec.podTemplate.spec.containers[].readinessProbe.periodSeconds` | `int32` |  |  |  |
| `spec.podTemplate.spec.containers[].readinessProbe.successThreshold` | `int32` |  |  |  |
| `spec.podTemplate.spec.containers[].readinessProbe.failureThreshold` | `int32` |  |  |  |
| `spec.podTemplate.spec.containers[].readinessProbe.terminationGracePeriodSeconds` | `int64` |  |  |  |
| `spec.podTemplate.spec.containers[].startupProbe` | `KubernetesCoreV1ApiProbe` |  |  |  |
| `spec.podTemplate.spec.containers[].startupProbe.exec` | `KubernetesCoreV1ApiExecAction` |  |  |  |
| `spec.podTemplate.spec.containers[].startupProbe.exec.command` | `[]string` |  |  |  |
| `spec.podTemplate.spec.containers[].startupProbe.httpGet` | `KubernetesCoreV1ApiHTTPGetAction` |  |  |  |
| `spec.podTemplate.spec.containers[].startupProbe.httpGet.path` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].startupProbe.httpGet.port` | `string` | yes |  |  |
| `spec.podTemplate.spec.containers[].startupProbe.httpGet.host` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].startupProbe.httpGet.scheme` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].startupProbe.httpGet.httpHeaders` | `[]KubernetesCoreV1ApiHTTPHeader` |  |  |  |
| `spec.podTemplate.spec.containers[].startupProbe.httpGet.httpHeaders[].name` | `string` | yes |  |  |
| `spec.podTemplate.spec.containers[].startupProbe.httpGet.httpHeaders[].value` | `string` | yes |  |  |
| `spec.podTemplate.spec.containers[].startupProbe.httpGet.protocol` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].startupProbe.tcpSocket` | `KubernetesCoreV1ApiTCPSocketAction` |  |  |  |
| `spec.podTemplate.spec.containers[].startupProbe.tcpSocket.port` | `string` | yes |  |  |
| `spec.podTemplate.spec.containers[].startupProbe.tcpSocket.host` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].startupProbe.grpc` | `KubernetesCoreV1ApiGRPCAction` |  |  |  |
| `spec.podTemplate.spec.containers[].startupProbe.grpc.port` | `int32` | yes |  |  |
| `spec.podTemplate.spec.containers[].startupProbe.grpc.service` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].startupProbe.grpc.mode` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].startupProbe.initialDelaySeconds` | `int32` |  |  |  |
| `spec.podTemplate.spec.containers[].startupProbe.timeoutSeconds` | `int32` |  |  |  |
| `spec.podTemplate.spec.containers[].startupProbe.periodSeconds` | `int32` |  |  |  |
| `spec.podTemplate.spec.containers[].startupProbe.successThreshold` | `int32` |  |  |  |
| `spec.podTemplate.spec.containers[].startupProbe.failureThreshold` | `int32` |  |  |  |
| `spec.podTemplate.spec.containers[].startupProbe.terminationGracePeriodSeconds` | `int64` |  |  |  |
| `spec.podTemplate.spec.containers[].lifecycle` | `KubernetesCoreV1ApiLifecycle` |  |  |  |
| `spec.podTemplate.spec.containers[].lifecycle.postStart` | `KubernetesCoreV1ApiLifecycleHandler` |  |  |  |
| `spec.podTemplate.spec.containers[].lifecycle.postStart.exec` | `KubernetesCoreV1ApiExecAction` |  |  |  |
| `spec.podTemplate.spec.containers[].lifecycle.postStart.exec.command` | `[]string` |  |  |  |
| `spec.podTemplate.spec.containers[].lifecycle.postStart.httpGet` | `KubernetesCoreV1ApiHTTPGetAction` |  |  |  |
| `spec.podTemplate.spec.containers[].lifecycle.postStart.httpGet.path` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].lifecycle.postStart.httpGet.port` | `string` | yes |  |  |
| `spec.podTemplate.spec.containers[].lifecycle.postStart.httpGet.host` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].lifecycle.postStart.httpGet.scheme` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].lifecycle.postStart.httpGet.httpHeaders` | `[]KubernetesCoreV1ApiHTTPHeader` |  |  |  |
| `spec.podTemplate.spec.containers[].lifecycle.postStart.httpGet.httpHeaders[].name` | `string` | yes |  |  |
| `spec.podTemplate.spec.containers[].lifecycle.postStart.httpGet.httpHeaders[].value` | `string` | yes |  |  |
| `spec.podTemplate.spec.containers[].lifecycle.postStart.httpGet.protocol` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].lifecycle.postStart.tcpSocket` | `KubernetesCoreV1ApiTCPSocketAction` |  |  |  |
| `spec.podTemplate.spec.containers[].lifecycle.postStart.tcpSocket.port` | `string` | yes |  |  |
| `spec.podTemplate.spec.containers[].lifecycle.postStart.tcpSocket.host` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].lifecycle.postStart.sleep` | `KubernetesCoreV1ApiSleepAction` |  |  |  |
| `spec.podTemplate.spec.containers[].lifecycle.postStart.sleep.seconds` | `int64` | yes |  |  |
| `spec.podTemplate.spec.containers[].lifecycle.preStop` | `KubernetesCoreV1ApiLifecycleHandler` |  |  |  |
| `spec.podTemplate.spec.containers[].lifecycle.preStop.exec` | `KubernetesCoreV1ApiExecAction` |  |  |  |
| `spec.podTemplate.spec.containers[].lifecycle.preStop.exec.command` | `[]string` |  |  |  |
| `spec.podTemplate.spec.containers[].lifecycle.preStop.httpGet` | `KubernetesCoreV1ApiHTTPGetAction` |  |  |  |
| `spec.podTemplate.spec.containers[].lifecycle.preStop.httpGet.path` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].lifecycle.preStop.httpGet.port` | `string` | yes |  |  |
| `spec.podTemplate.spec.containers[].lifecycle.preStop.httpGet.host` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].lifecycle.preStop.httpGet.scheme` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].lifecycle.preStop.httpGet.httpHeaders` | `[]KubernetesCoreV1ApiHTTPHeader` |  |  |  |
| `spec.podTemplate.spec.containers[].lifecycle.preStop.httpGet.httpHeaders[].name` | `string` | yes |  |  |
| `spec.podTemplate.spec.containers[].lifecycle.preStop.httpGet.httpHeaders[].value` | `string` | yes |  |  |
| `spec.podTemplate.spec.containers[].lifecycle.preStop.httpGet.protocol` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].lifecycle.preStop.tcpSocket` | `KubernetesCoreV1ApiTCPSocketAction` |  |  |  |
| `spec.podTemplate.spec.containers[].lifecycle.preStop.tcpSocket.port` | `string` | yes |  |  |
| `spec.podTemplate.spec.containers[].lifecycle.preStop.tcpSocket.host` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].lifecycle.preStop.sleep` | `KubernetesCoreV1ApiSleepAction` |  |  |  |
| `spec.podTemplate.spec.containers[].lifecycle.preStop.sleep.seconds` | `int64` | yes |  |  |
| `spec.podTemplate.spec.containers[].lifecycle.stopSignal` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].terminationMessagePath` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].terminationMessagePolicy` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].imagePullPolicy` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].securityContext` | `KubernetesCoreV1ApiSecurityContext` |  |  |  |
| `spec.podTemplate.spec.containers[].securityContext.capabilities` | `KubernetesCoreV1ApiCapabilities` |  |  |  |
| `spec.podTemplate.spec.containers[].securityContext.capabilities.add` | `[]string` |  |  |  |
| `spec.podTemplate.spec.containers[].securityContext.capabilities.drop` | `[]string` |  |  |  |
| `spec.podTemplate.spec.containers[].securityContext.privileged` | `bool` |  |  |  |
| `spec.podTemplate.spec.containers[].securityContext.seLinuxOptions` | `KubernetesCoreV1ApiSELinuxOptions` |  |  |  |
| `spec.podTemplate.spec.containers[].securityContext.seLinuxOptions.user` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].securityContext.seLinuxOptions.role` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].securityContext.seLinuxOptions.type` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].securityContext.seLinuxOptions.level` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].securityContext.windowsOptions` | `KubernetesCoreV1ApiWindowsSecurityContextOptions` |  |  |  |
| `spec.podTemplate.spec.containers[].securityContext.windowsOptions.gmsaCredentialSpecName` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].securityContext.windowsOptions.gmsaCredentialSpec` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].securityContext.windowsOptions.runAsUserName` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].securityContext.windowsOptions.hostProcess` | `bool` |  |  |  |
| `spec.podTemplate.spec.containers[].securityContext.runAsUser` | `int64` |  |  |  |
| `spec.podTemplate.spec.containers[].securityContext.runAsGroup` | `int64` |  |  |  |
| `spec.podTemplate.spec.containers[].securityContext.runAsNonRoot` | `bool` |  |  |  |
| `spec.podTemplate.spec.containers[].securityContext.readOnlyRootFilesystem` | `bool` |  |  |  |
| `spec.podTemplate.spec.containers[].securityContext.allowPrivilegeEscalation` | `bool` |  |  |  |
| `spec.podTemplate.spec.containers[].securityContext.procMount` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].securityContext.seccompProfile` | `KubernetesCoreV1ApiSeccompProfile` |  |  |  |
| `spec.podTemplate.spec.containers[].securityContext.seccompProfile.type` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].securityContext.seccompProfile.localhostProfile` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].securityContext.appArmorProfile` | `KubernetesCoreV1ApiAppArmorProfile` |  |  |  |
| `spec.podTemplate.spec.containers[].securityContext.appArmorProfile.type` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].securityContext.appArmorProfile.localhostProfile` | `string` |  |  |  |
| `spec.podTemplate.spec.containers[].stdin` | `bool` |  |  |  |
| `spec.podTemplate.spec.containers[].stdinOnce` | `bool` |  |  |  |
| `spec.podTemplate.spec.containers[].tty` | `bool` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers` | `[]KubernetesCoreV1ApiEphemeralContainer` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].name` | `string` | yes |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].image` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].command` | `[]string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].args` | `[]string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].workingDir` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].ports` | `[]KubernetesCoreV1ApiContainerPort` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].ports[].name` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].ports[].hostPort` | `int32` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].ports[].containerPort` | `int32` | yes |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].ports[].protocol` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].ports[].hostIP` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].envFrom` | `[]KubernetesCoreV1ApiEnvFromSource` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].envFrom[].prefix` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].envFrom[].configMapRef` | `KubernetesCoreV1ApiConfigMapEnvSource` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].envFrom[].configMapRef.name` | `string \| valueFrom` | yes |  | KubernetesConfigMap (`status.outputs.configmap_name`) |
| `spec.podTemplate.spec.ephemeralContainers[].envFrom[].configMapRef.optional` | `bool` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].envFrom[].secretRef` | `KubernetesCoreV1ApiSecretEnvSource` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].envFrom[].secretRef.name` | `string \| valueFrom` | yes |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.podTemplate.spec.ephemeralContainers[].envFrom[].secretRef.optional` | `bool` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].env` | `[]KubernetesCoreV1ApiEnvVar` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].env[].name` | `string` | yes |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].env[].value` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].env[].valueFrom` | `KubernetesCoreV1ApiEnvVarSource` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].env[].valueFrom.fieldRef` | `KubernetesCoreV1ApiObjectFieldSelector` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].env[].valueFrom.fieldRef.apiVersion` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].env[].valueFrom.fieldRef.fieldPath` | `string` | yes |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].env[].valueFrom.resourceFieldRef` | `KubernetesCoreV1ApiResourceFieldSelector` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].env[].valueFrom.resourceFieldRef.containerName` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].env[].valueFrom.resourceFieldRef.resource` | `string` | yes |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].env[].valueFrom.resourceFieldRef.divisor` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].env[].valueFrom.configMapKeyRef` | `KubernetesCoreV1ApiConfigMapKeySelector` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].env[].valueFrom.configMapKeyRef.name` | `string \| valueFrom` | yes |  | KubernetesConfigMap (`status.outputs.configmap_name`) |
| `spec.podTemplate.spec.ephemeralContainers[].env[].valueFrom.configMapKeyRef.key` | `string` | yes |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].env[].valueFrom.configMapKeyRef.optional` | `bool` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].env[].valueFrom.secretKeyRef` | `KubernetesCoreV1ApiSecretKeySelector` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].env[].valueFrom.secretKeyRef.name` | `string \| valueFrom` | yes |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.podTemplate.spec.ephemeralContainers[].env[].valueFrom.secretKeyRef.key` | `string` | yes |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].env[].valueFrom.secretKeyRef.optional` | `bool` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].env[].valueFrom.fileKeyRef` | `KubernetesCoreV1ApiFileKeySelector` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].env[].valueFrom.fileKeyRef.volumeName` | `string` | yes |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].env[].valueFrom.fileKeyRef.path` | `string` | yes |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].env[].valueFrom.fileKeyRef.key` | `string` | yes |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].env[].valueFrom.fileKeyRef.optional` | `bool` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].resources` | `KubernetesCoreV1ApiResourceRequirements` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].resources.limits` | `map<string, string>` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].resources.requests` | `map<string, string>` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].resources.claims` | `[]KubernetesCoreV1ApiResourceClaim` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].resources.claims[].name` | `string` | yes |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].resources.claims[].request` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].resizePolicy` | `[]KubernetesCoreV1ApiContainerResizePolicy` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].resizePolicy[].resourceName` | `string` | yes |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].resizePolicy[].restartPolicy` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].restartPolicy` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].restartPolicyRules` | `[]KubernetesCoreV1ApiContainerRestartRule` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].restartPolicyRules[].action` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].restartPolicyRules[].exitCodes` | `KubernetesCoreV1ApiContainerRestartRuleOnExitCodes` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].restartPolicyRules[].exitCodes.operator` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].restartPolicyRules[].exitCodes.values` | `[]int32` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].volumeMounts` | `[]KubernetesCoreV1ApiVolumeMount` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].volumeMounts[].name` | `string` | yes |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].volumeMounts[].readOnly` | `bool` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].volumeMounts[].recursiveReadOnly` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].volumeMounts[].mountPath` | `string` | yes |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].volumeMounts[].subPath` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].volumeMounts[].mountPropagation` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].volumeMounts[].subPathExpr` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].volumeMounts[].bindMountOptions` | `[]string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].volumeDevices` | `[]KubernetesCoreV1ApiVolumeDevice` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].volumeDevices[].name` | `string` | yes |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].volumeDevices[].devicePath` | `string` | yes |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].livenessProbe` | `KubernetesCoreV1ApiProbe` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].livenessProbe.exec` | `KubernetesCoreV1ApiExecAction` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].livenessProbe.exec.command` | `[]string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].livenessProbe.httpGet` | `KubernetesCoreV1ApiHTTPGetAction` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].livenessProbe.httpGet.path` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].livenessProbe.httpGet.port` | `string` | yes |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].livenessProbe.httpGet.host` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].livenessProbe.httpGet.scheme` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].livenessProbe.httpGet.httpHeaders` | `[]KubernetesCoreV1ApiHTTPHeader` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].livenessProbe.httpGet.httpHeaders[].name` | `string` | yes |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].livenessProbe.httpGet.httpHeaders[].value` | `string` | yes |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].livenessProbe.httpGet.protocol` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].livenessProbe.tcpSocket` | `KubernetesCoreV1ApiTCPSocketAction` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].livenessProbe.tcpSocket.port` | `string` | yes |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].livenessProbe.tcpSocket.host` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].livenessProbe.grpc` | `KubernetesCoreV1ApiGRPCAction` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].livenessProbe.grpc.port` | `int32` | yes |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].livenessProbe.grpc.service` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].livenessProbe.grpc.mode` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].livenessProbe.initialDelaySeconds` | `int32` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].livenessProbe.timeoutSeconds` | `int32` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].livenessProbe.periodSeconds` | `int32` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].livenessProbe.successThreshold` | `int32` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].livenessProbe.failureThreshold` | `int32` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].livenessProbe.terminationGracePeriodSeconds` | `int64` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].readinessProbe` | `KubernetesCoreV1ApiProbe` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].readinessProbe.exec` | `KubernetesCoreV1ApiExecAction` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].readinessProbe.exec.command` | `[]string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].readinessProbe.httpGet` | `KubernetesCoreV1ApiHTTPGetAction` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].readinessProbe.httpGet.path` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].readinessProbe.httpGet.port` | `string` | yes |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].readinessProbe.httpGet.host` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].readinessProbe.httpGet.scheme` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].readinessProbe.httpGet.httpHeaders` | `[]KubernetesCoreV1ApiHTTPHeader` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].readinessProbe.httpGet.httpHeaders[].name` | `string` | yes |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].readinessProbe.httpGet.httpHeaders[].value` | `string` | yes |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].readinessProbe.httpGet.protocol` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].readinessProbe.tcpSocket` | `KubernetesCoreV1ApiTCPSocketAction` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].readinessProbe.tcpSocket.port` | `string` | yes |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].readinessProbe.tcpSocket.host` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].readinessProbe.grpc` | `KubernetesCoreV1ApiGRPCAction` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].readinessProbe.grpc.port` | `int32` | yes |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].readinessProbe.grpc.service` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].readinessProbe.grpc.mode` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].readinessProbe.initialDelaySeconds` | `int32` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].readinessProbe.timeoutSeconds` | `int32` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].readinessProbe.periodSeconds` | `int32` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].readinessProbe.successThreshold` | `int32` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].readinessProbe.failureThreshold` | `int32` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].readinessProbe.terminationGracePeriodSeconds` | `int64` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].startupProbe` | `KubernetesCoreV1ApiProbe` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].startupProbe.exec` | `KubernetesCoreV1ApiExecAction` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].startupProbe.exec.command` | `[]string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].startupProbe.httpGet` | `KubernetesCoreV1ApiHTTPGetAction` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].startupProbe.httpGet.path` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].startupProbe.httpGet.port` | `string` | yes |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].startupProbe.httpGet.host` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].startupProbe.httpGet.scheme` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].startupProbe.httpGet.httpHeaders` | `[]KubernetesCoreV1ApiHTTPHeader` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].startupProbe.httpGet.httpHeaders[].name` | `string` | yes |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].startupProbe.httpGet.httpHeaders[].value` | `string` | yes |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].startupProbe.httpGet.protocol` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].startupProbe.tcpSocket` | `KubernetesCoreV1ApiTCPSocketAction` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].startupProbe.tcpSocket.port` | `string` | yes |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].startupProbe.tcpSocket.host` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].startupProbe.grpc` | `KubernetesCoreV1ApiGRPCAction` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].startupProbe.grpc.port` | `int32` | yes |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].startupProbe.grpc.service` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].startupProbe.grpc.mode` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].startupProbe.initialDelaySeconds` | `int32` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].startupProbe.timeoutSeconds` | `int32` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].startupProbe.periodSeconds` | `int32` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].startupProbe.successThreshold` | `int32` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].startupProbe.failureThreshold` | `int32` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].startupProbe.terminationGracePeriodSeconds` | `int64` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].lifecycle` | `KubernetesCoreV1ApiLifecycle` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].lifecycle.postStart` | `KubernetesCoreV1ApiLifecycleHandler` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].lifecycle.postStart.exec` | `KubernetesCoreV1ApiExecAction` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].lifecycle.postStart.exec.command` | `[]string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].lifecycle.postStart.httpGet` | `KubernetesCoreV1ApiHTTPGetAction` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].lifecycle.postStart.httpGet.path` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].lifecycle.postStart.httpGet.port` | `string` | yes |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].lifecycle.postStart.httpGet.host` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].lifecycle.postStart.httpGet.scheme` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].lifecycle.postStart.httpGet.httpHeaders` | `[]KubernetesCoreV1ApiHTTPHeader` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].lifecycle.postStart.httpGet.httpHeaders[].name` | `string` | yes |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].lifecycle.postStart.httpGet.httpHeaders[].value` | `string` | yes |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].lifecycle.postStart.httpGet.protocol` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].lifecycle.postStart.tcpSocket` | `KubernetesCoreV1ApiTCPSocketAction` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].lifecycle.postStart.tcpSocket.port` | `string` | yes |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].lifecycle.postStart.tcpSocket.host` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].lifecycle.postStart.sleep` | `KubernetesCoreV1ApiSleepAction` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].lifecycle.postStart.sleep.seconds` | `int64` | yes |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].lifecycle.preStop` | `KubernetesCoreV1ApiLifecycleHandler` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].lifecycle.preStop.exec` | `KubernetesCoreV1ApiExecAction` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].lifecycle.preStop.exec.command` | `[]string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].lifecycle.preStop.httpGet` | `KubernetesCoreV1ApiHTTPGetAction` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].lifecycle.preStop.httpGet.path` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].lifecycle.preStop.httpGet.port` | `string` | yes |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].lifecycle.preStop.httpGet.host` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].lifecycle.preStop.httpGet.scheme` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].lifecycle.preStop.httpGet.httpHeaders` | `[]KubernetesCoreV1ApiHTTPHeader` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].lifecycle.preStop.httpGet.httpHeaders[].name` | `string` | yes |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].lifecycle.preStop.httpGet.httpHeaders[].value` | `string` | yes |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].lifecycle.preStop.httpGet.protocol` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].lifecycle.preStop.tcpSocket` | `KubernetesCoreV1ApiTCPSocketAction` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].lifecycle.preStop.tcpSocket.port` | `string` | yes |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].lifecycle.preStop.tcpSocket.host` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].lifecycle.preStop.sleep` | `KubernetesCoreV1ApiSleepAction` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].lifecycle.preStop.sleep.seconds` | `int64` | yes |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].lifecycle.stopSignal` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].terminationMessagePath` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].terminationMessagePolicy` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].imagePullPolicy` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].securityContext` | `KubernetesCoreV1ApiSecurityContext` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].securityContext.capabilities` | `KubernetesCoreV1ApiCapabilities` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].securityContext.capabilities.add` | `[]string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].securityContext.capabilities.drop` | `[]string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].securityContext.privileged` | `bool` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].securityContext.seLinuxOptions` | `KubernetesCoreV1ApiSELinuxOptions` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].securityContext.seLinuxOptions.user` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].securityContext.seLinuxOptions.role` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].securityContext.seLinuxOptions.type` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].securityContext.seLinuxOptions.level` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].securityContext.windowsOptions` | `KubernetesCoreV1ApiWindowsSecurityContextOptions` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].securityContext.windowsOptions.gmsaCredentialSpecName` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].securityContext.windowsOptions.gmsaCredentialSpec` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].securityContext.windowsOptions.runAsUserName` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].securityContext.windowsOptions.hostProcess` | `bool` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].securityContext.runAsUser` | `int64` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].securityContext.runAsGroup` | `int64` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].securityContext.runAsNonRoot` | `bool` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].securityContext.readOnlyRootFilesystem` | `bool` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].securityContext.allowPrivilegeEscalation` | `bool` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].securityContext.procMount` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].securityContext.seccompProfile` | `KubernetesCoreV1ApiSeccompProfile` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].securityContext.seccompProfile.type` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].securityContext.seccompProfile.localhostProfile` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].securityContext.appArmorProfile` | `KubernetesCoreV1ApiAppArmorProfile` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].securityContext.appArmorProfile.type` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].securityContext.appArmorProfile.localhostProfile` | `string` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].stdin` | `bool` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].stdinOnce` | `bool` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].tty` | `bool` |  |  |  |
| `spec.podTemplate.spec.ephemeralContainers[].targetContainerName` | `string` |  |  |  |
| `spec.podTemplate.spec.restartPolicy` | `string` |  |  |  |
| `spec.podTemplate.spec.terminationGracePeriodSeconds` | `int64` |  |  |  |
| `spec.podTemplate.spec.activeDeadlineSeconds` | `int64` |  |  |  |
| `spec.podTemplate.spec.dnsPolicy` | `string` |  |  |  |
| `spec.podTemplate.spec.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.podTemplate.spec.serviceAccountName` | `string \| valueFrom` |  |  | KubernetesServiceAccount (`status.outputs.service_account_name`) |
| `spec.podTemplate.spec.serviceAccount` | `string` |  |  |  |
| `spec.podTemplate.spec.automountServiceAccountToken` | `bool` |  |  |  |
| `spec.podTemplate.spec.nodeName` | `string` |  |  |  |
| `spec.podTemplate.spec.hostNetwork` | `bool` |  |  |  |
| `spec.podTemplate.spec.hostPID` | `bool` |  |  |  |
| `spec.podTemplate.spec.hostIPC` | `bool` |  |  |  |
| `spec.podTemplate.spec.shareProcessNamespace` | `bool` |  |  |  |
| `spec.podTemplate.spec.securityContext` | `KubernetesCoreV1ApiPodSecurityContext` |  |  |  |
| `spec.podTemplate.spec.securityContext.seLinuxOptions` | `KubernetesCoreV1ApiSELinuxOptions` |  |  |  |
| `spec.podTemplate.spec.securityContext.seLinuxOptions.user` | `string` |  |  |  |
| `spec.podTemplate.spec.securityContext.seLinuxOptions.role` | `string` |  |  |  |
| `spec.podTemplate.spec.securityContext.seLinuxOptions.type` | `string` |  |  |  |
| `spec.podTemplate.spec.securityContext.seLinuxOptions.level` | `string` |  |  |  |
| `spec.podTemplate.spec.securityContext.windowsOptions` | `KubernetesCoreV1ApiWindowsSecurityContextOptions` |  |  |  |
| `spec.podTemplate.spec.securityContext.windowsOptions.gmsaCredentialSpecName` | `string` |  |  |  |
| `spec.podTemplate.spec.securityContext.windowsOptions.gmsaCredentialSpec` | `string` |  |  |  |
| `spec.podTemplate.spec.securityContext.windowsOptions.runAsUserName` | `string` |  |  |  |
| `spec.podTemplate.spec.securityContext.windowsOptions.hostProcess` | `bool` |  |  |  |
| `spec.podTemplate.spec.securityContext.runAsUser` | `int64` |  |  |  |
| `spec.podTemplate.spec.securityContext.runAsGroup` | `int64` |  |  |  |
| `spec.podTemplate.spec.securityContext.runAsNonRoot` | `bool` |  |  |  |
| `spec.podTemplate.spec.securityContext.supplementalGroups` | `[]int64` |  |  |  |
| `spec.podTemplate.spec.securityContext.supplementalGroupsPolicy` | `string` |  |  |  |
| `spec.podTemplate.spec.securityContext.fsGroup` | `int64` |  |  |  |
| `spec.podTemplate.spec.securityContext.sysctls` | `[]KubernetesCoreV1ApiSysctl` |  |  |  |
| `spec.podTemplate.spec.securityContext.sysctls[].name` | `string` | yes |  |  |
| `spec.podTemplate.spec.securityContext.sysctls[].value` | `string` | yes |  |  |
| `spec.podTemplate.spec.securityContext.fsGroupChangePolicy` | `string` |  |  |  |
| `spec.podTemplate.spec.securityContext.seccompProfile` | `KubernetesCoreV1ApiSeccompProfile` |  |  |  |
| `spec.podTemplate.spec.securityContext.seccompProfile.type` | `string` |  |  |  |
| `spec.podTemplate.spec.securityContext.seccompProfile.localhostProfile` | `string` |  |  |  |
| `spec.podTemplate.spec.securityContext.appArmorProfile` | `KubernetesCoreV1ApiAppArmorProfile` |  |  |  |
| `spec.podTemplate.spec.securityContext.appArmorProfile.type` | `string` |  |  |  |
| `spec.podTemplate.spec.securityContext.appArmorProfile.localhostProfile` | `string` |  |  |  |
| `spec.podTemplate.spec.securityContext.seLinuxChangePolicy` | `string` |  |  |  |
| `spec.podTemplate.spec.imagePullSecrets` | `[]KubernetesCoreV1ApiLocalObjectReference` |  |  |  |
| `spec.podTemplate.spec.imagePullSecrets[].name` | `string \| valueFrom` | yes |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.podTemplate.spec.hostname` | `string` |  |  |  |
| `spec.podTemplate.spec.subdomain` | `string` |  |  |  |
| `spec.podTemplate.spec.affinity` | `KubernetesCoreV1ApiAffinity` |  |  |  |
| `spec.podTemplate.spec.affinity.nodeAffinity` | `KubernetesCoreV1ApiNodeAffinity` |  |  |  |
| `spec.podTemplate.spec.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution` | `KubernetesCoreV1ApiNodeSelector` |  |  |  |
| `spec.podTemplate.spec.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms` | `[]KubernetesCoreV1ApiNodeSelectorTerm` | yes |  |  |
| `spec.podTemplate.spec.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[].matchExpressions` | `[]KubernetesCoreV1ApiNodeSelectorRequirement` |  |  |  |
| `spec.podTemplate.spec.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[].matchExpressions[].key` | `string` | yes |  |  |
| `spec.podTemplate.spec.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[].matchExpressions[].operator` | `string` |  |  |  |
| `spec.podTemplate.spec.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[].matchExpressions[].values` | `[]string` |  |  |  |
| `spec.podTemplate.spec.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[].matchFields` | `[]KubernetesCoreV1ApiNodeSelectorRequirement` |  |  |  |
| `spec.podTemplate.spec.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[].matchFields[].key` | `string` | yes |  |  |
| `spec.podTemplate.spec.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[].matchFields[].operator` | `string` |  |  |  |
| `spec.podTemplate.spec.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[].matchFields[].values` | `[]string` |  |  |  |
| `spec.podTemplate.spec.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution` | `[]KubernetesCoreV1ApiPreferredSchedulingTerm` |  |  |  |
| `spec.podTemplate.spec.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[].weight` | `int32` | yes |  |  |
| `spec.podTemplate.spec.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[].preference` | `KubernetesCoreV1ApiNodeSelectorTerm` | yes |  |  |
| `spec.podTemplate.spec.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[].preference.matchExpressions` | `[]KubernetesCoreV1ApiNodeSelectorRequirement` |  |  |  |
| `spec.podTemplate.spec.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[].preference.matchExpressions[].key` | `string` | yes |  |  |
| `spec.podTemplate.spec.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[].preference.matchExpressions[].operator` | `string` |  |  |  |
| `spec.podTemplate.spec.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[].preference.matchExpressions[].values` | `[]string` |  |  |  |
| `spec.podTemplate.spec.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[].preference.matchFields` | `[]KubernetesCoreV1ApiNodeSelectorRequirement` |  |  |  |
| `spec.podTemplate.spec.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[].preference.matchFields[].key` | `string` | yes |  |  |
| `spec.podTemplate.spec.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[].preference.matchFields[].operator` | `string` |  |  |  |
| `spec.podTemplate.spec.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[].preference.matchFields[].values` | `[]string` |  |  |  |
| `spec.podTemplate.spec.affinity.podAffinity` | `KubernetesCoreV1ApiPodAffinity` |  |  |  |
| `spec.podTemplate.spec.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution` | `[]KubernetesCoreV1ApiPodAffinityTerm` |  |  |  |
| `spec.podTemplate.spec.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[].labelSelector` | `KubernetesCoreV1ApiLabelSelector` |  |  |  |
| `spec.podTemplate.spec.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[].labelSelector.matchLabels` | `map<string, string>` |  |  |  |
| `spec.podTemplate.spec.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[].labelSelector.matchExpressions` | `[]KubernetesCoreV1ApiLabelSelectorRequirement` |  |  |  |
| `spec.podTemplate.spec.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[].labelSelector.matchExpressions[].key` | `string` | yes |  |  |
| `spec.podTemplate.spec.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[].labelSelector.matchExpressions[].operator` | `string` | yes |  |  |
| `spec.podTemplate.spec.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[].labelSelector.matchExpressions[].values` | `[]string` |  |  |  |
| `spec.podTemplate.spec.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[].namespaces` | `[]string` |  |  |  |
| `spec.podTemplate.spec.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[].topologyKey` | `string` | yes |  |  |
| `spec.podTemplate.spec.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[].namespaceSelector` | `KubernetesCoreV1ApiLabelSelector` |  |  |  |
| `spec.podTemplate.spec.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[].namespaceSelector.matchLabels` | `map<string, string>` |  |  |  |
| `spec.podTemplate.spec.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[].namespaceSelector.matchExpressions` | `[]KubernetesCoreV1ApiLabelSelectorRequirement` |  |  |  |
| `spec.podTemplate.spec.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[].namespaceSelector.matchExpressions[].key` | `string` | yes |  |  |
| `spec.podTemplate.spec.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[].namespaceSelector.matchExpressions[].operator` | `string` | yes |  |  |
| `spec.podTemplate.spec.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[].namespaceSelector.matchExpressions[].values` | `[]string` |  |  |  |
| `spec.podTemplate.spec.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[].matchLabelKeys` | `[]string` |  |  |  |
| `spec.podTemplate.spec.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[].mismatchLabelKeys` | `[]string` |  |  |  |
| `spec.podTemplate.spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution` | `[]KubernetesCoreV1ApiWeightedPodAffinityTerm` |  |  |  |
| `spec.podTemplate.spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution[].weight` | `int32` | yes |  |  |
| `spec.podTemplate.spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm` | `KubernetesCoreV1ApiPodAffinityTerm` | yes |  |  |
| `spec.podTemplate.spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.labelSelector` | `KubernetesCoreV1ApiLabelSelector` |  |  |  |
| `spec.podTemplate.spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.labelSelector.matchLabels` | `map<string, string>` |  |  |  |
| `spec.podTemplate.spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.labelSelector.matchExpressions` | `[]KubernetesCoreV1ApiLabelSelectorRequirement` |  |  |  |
| `spec.podTemplate.spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.labelSelector.matchExpressions[].key` | `string` | yes |  |  |
| `spec.podTemplate.spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.labelSelector.matchExpressions[].operator` | `string` | yes |  |  |
| `spec.podTemplate.spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.labelSelector.matchExpressions[].values` | `[]string` |  |  |  |
| `spec.podTemplate.spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.namespaces` | `[]string` |  |  |  |
| `spec.podTemplate.spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.topologyKey` | `string` | yes |  |  |
| `spec.podTemplate.spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.namespaceSelector` | `KubernetesCoreV1ApiLabelSelector` |  |  |  |
| `spec.podTemplate.spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.namespaceSelector.matchLabels` | `map<string, string>` |  |  |  |
| `spec.podTemplate.spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.namespaceSelector.matchExpressions` | `[]KubernetesCoreV1ApiLabelSelectorRequirement` |  |  |  |
| `spec.podTemplate.spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.namespaceSelector.matchExpressions[].key` | `string` | yes |  |  |
| `spec.podTemplate.spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.namespaceSelector.matchExpressions[].operator` | `string` | yes |  |  |
| `spec.podTemplate.spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.namespaceSelector.matchExpressions[].values` | `[]string` |  |  |  |
| `spec.podTemplate.spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.matchLabelKeys` | `[]string` |  |  |  |
| `spec.podTemplate.spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.mismatchLabelKeys` | `[]string` |  |  |  |
| `spec.podTemplate.spec.affinity.podAntiAffinity` | `KubernetesCoreV1ApiPodAntiAffinity` |  |  |  |
| `spec.podTemplate.spec.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution` | `[]KubernetesCoreV1ApiPodAffinityTerm` |  |  |  |
| `spec.podTemplate.spec.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[].labelSelector` | `KubernetesCoreV1ApiLabelSelector` |  |  |  |
| `spec.podTemplate.spec.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[].labelSelector.matchLabels` | `map<string, string>` |  |  |  |
| `spec.podTemplate.spec.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[].labelSelector.matchExpressions` | `[]KubernetesCoreV1ApiLabelSelectorRequirement` |  |  |  |
| `spec.podTemplate.spec.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[].labelSelector.matchExpressions[].key` | `string` | yes |  |  |
| `spec.podTemplate.spec.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[].labelSelector.matchExpressions[].operator` | `string` | yes |  |  |
| `spec.podTemplate.spec.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[].labelSelector.matchExpressions[].values` | `[]string` |  |  |  |
| `spec.podTemplate.spec.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[].namespaces` | `[]string` |  |  |  |
| `spec.podTemplate.spec.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[].topologyKey` | `string` | yes |  |  |
| `spec.podTemplate.spec.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[].namespaceSelector` | `KubernetesCoreV1ApiLabelSelector` |  |  |  |
| `spec.podTemplate.spec.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[].namespaceSelector.matchLabels` | `map<string, string>` |  |  |  |
| `spec.podTemplate.spec.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[].namespaceSelector.matchExpressions` | `[]KubernetesCoreV1ApiLabelSelectorRequirement` |  |  |  |
| `spec.podTemplate.spec.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[].namespaceSelector.matchExpressions[].key` | `string` | yes |  |  |
| `spec.podTemplate.spec.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[].namespaceSelector.matchExpressions[].operator` | `string` | yes |  |  |
| `spec.podTemplate.spec.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[].namespaceSelector.matchExpressions[].values` | `[]string` |  |  |  |
| `spec.podTemplate.spec.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[].matchLabelKeys` | `[]string` |  |  |  |
| `spec.podTemplate.spec.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[].mismatchLabelKeys` | `[]string` |  |  |  |
| `spec.podTemplate.spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution` | `[]KubernetesCoreV1ApiWeightedPodAffinityTerm` |  |  |  |
| `spec.podTemplate.spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[].weight` | `int32` | yes |  |  |
| `spec.podTemplate.spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm` | `KubernetesCoreV1ApiPodAffinityTerm` | yes |  |  |
| `spec.podTemplate.spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.labelSelector` | `KubernetesCoreV1ApiLabelSelector` |  |  |  |
| `spec.podTemplate.spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.labelSelector.matchLabels` | `map<string, string>` |  |  |  |
| `spec.podTemplate.spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.labelSelector.matchExpressions` | `[]KubernetesCoreV1ApiLabelSelectorRequirement` |  |  |  |
| `spec.podTemplate.spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.labelSelector.matchExpressions[].key` | `string` | yes |  |  |
| `spec.podTemplate.spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.labelSelector.matchExpressions[].operator` | `string` | yes |  |  |
| `spec.podTemplate.spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.labelSelector.matchExpressions[].values` | `[]string` |  |  |  |
| `spec.podTemplate.spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.namespaces` | `[]string` |  |  |  |
| `spec.podTemplate.spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.topologyKey` | `string` | yes |  |  |
| `spec.podTemplate.spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.namespaceSelector` | `KubernetesCoreV1ApiLabelSelector` |  |  |  |
| `spec.podTemplate.spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.namespaceSelector.matchLabels` | `map<string, string>` |  |  |  |
| `spec.podTemplate.spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.namespaceSelector.matchExpressions` | `[]KubernetesCoreV1ApiLabelSelectorRequirement` |  |  |  |
| `spec.podTemplate.spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.namespaceSelector.matchExpressions[].key` | `string` | yes |  |  |
| `spec.podTemplate.spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.namespaceSelector.matchExpressions[].operator` | `string` | yes |  |  |
| `spec.podTemplate.spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.namespaceSelector.matchExpressions[].values` | `[]string` |  |  |  |
| `spec.podTemplate.spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.matchLabelKeys` | `[]string` |  |  |  |
| `spec.podTemplate.spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.mismatchLabelKeys` | `[]string` |  |  |  |
| `spec.podTemplate.spec.schedulerName` | `string` |  |  |  |
| `spec.podTemplate.spec.tolerations` | `[]KubernetesCoreV1ApiToleration` |  |  |  |
| `spec.podTemplate.spec.tolerations[].key` | `string` |  |  |  |
| `spec.podTemplate.spec.tolerations[].operator` | `string` |  |  |  |
| `spec.podTemplate.spec.tolerations[].value` | `string` |  |  |  |
| `spec.podTemplate.spec.tolerations[].effect` | `string` |  |  |  |
| `spec.podTemplate.spec.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.podTemplate.spec.hostAliases` | `[]KubernetesCoreV1ApiHostAlias` |  |  |  |
| `spec.podTemplate.spec.hostAliases[].ip` | `string` | yes |  |  |
| `spec.podTemplate.spec.hostAliases[].hostnames` | `[]string` |  |  |  |
| `spec.podTemplate.spec.priorityClassName` | `string \| valueFrom` |  |  | KubernetesPriorityClass (`status.outputs.priority_class_name`) |
| `spec.podTemplate.spec.priority` | `int32` |  |  |  |
| `spec.podTemplate.spec.dnsConfig` | `KubernetesCoreV1ApiPodDNSConfig` |  |  |  |
| `spec.podTemplate.spec.dnsConfig.nameservers` | `[]string` |  |  |  |
| `spec.podTemplate.spec.dnsConfig.searches` | `[]string` |  |  |  |
| `spec.podTemplate.spec.dnsConfig.options` | `[]KubernetesCoreV1ApiPodDNSConfigOption` |  |  |  |
| `spec.podTemplate.spec.dnsConfig.options[].name` | `string` |  |  |  |
| `spec.podTemplate.spec.dnsConfig.options[].value` | `string` |  |  |  |
| `spec.podTemplate.spec.readinessGates` | `[]KubernetesCoreV1ApiPodReadinessGate` |  |  |  |
| `spec.podTemplate.spec.readinessGates[].conditionType` | `string` | yes |  |  |
| `spec.podTemplate.spec.runtimeClassName` | `string` |  |  |  |
| `spec.podTemplate.spec.enableServiceLinks` | `bool` |  |  |  |
| `spec.podTemplate.spec.preemptionPolicy` | `string` |  |  |  |
| `spec.podTemplate.spec.overhead` | `map<string, string>` |  |  |  |
| `spec.podTemplate.spec.topologySpreadConstraints` | `[]KubernetesCoreV1ApiTopologySpreadConstraint` |  |  |  |
| `spec.podTemplate.spec.topologySpreadConstraints[].maxSkew` | `int32` | yes |  |  |
| `spec.podTemplate.spec.topologySpreadConstraints[].topologyKey` | `string` | yes |  |  |
| `spec.podTemplate.spec.topologySpreadConstraints[].whenUnsatisfiable` | `string` |  |  |  |
| `spec.podTemplate.spec.topologySpreadConstraints[].labelSelector` | `KubernetesCoreV1ApiLabelSelector` |  |  |  |
| `spec.podTemplate.spec.topologySpreadConstraints[].labelSelector.matchLabels` | `map<string, string>` |  |  |  |
| `spec.podTemplate.spec.topologySpreadConstraints[].labelSelector.matchExpressions` | `[]KubernetesCoreV1ApiLabelSelectorRequirement` |  |  |  |
| `spec.podTemplate.spec.topologySpreadConstraints[].labelSelector.matchExpressions[].key` | `string` | yes |  |  |
| `spec.podTemplate.spec.topologySpreadConstraints[].labelSelector.matchExpressions[].operator` | `string` | yes |  |  |
| `spec.podTemplate.spec.topologySpreadConstraints[].labelSelector.matchExpressions[].values` | `[]string` |  |  |  |
| `spec.podTemplate.spec.topologySpreadConstraints[].minDomains` | `int32` |  |  |  |
| `spec.podTemplate.spec.topologySpreadConstraints[].nodeAffinityPolicy` | `string` |  |  |  |
| `spec.podTemplate.spec.topologySpreadConstraints[].nodeTaintsPolicy` | `string` |  |  |  |
| `spec.podTemplate.spec.topologySpreadConstraints[].matchLabelKeys` | `[]string` |  |  |  |
| `spec.podTemplate.spec.setHostnameAsFQDN` | `bool` |  |  |  |
| `spec.podTemplate.spec.os` | `KubernetesCoreV1ApiPodOS` |  |  |  |
| `spec.podTemplate.spec.os.name` | `string` | yes |  |  |
| `spec.podTemplate.spec.hostUsers` | `bool` |  |  |  |
| `spec.podTemplate.spec.schedulingGates` | `[]KubernetesCoreV1ApiPodSchedulingGate` |  |  |  |
| `spec.podTemplate.spec.schedulingGates[].name` | `string` | yes |  |  |
| `spec.podTemplate.spec.resourceClaims` | `[]KubernetesCoreV1ApiPodResourceClaim` |  |  |  |
| `spec.podTemplate.spec.resourceClaims[].name` | `string` | yes |  |  |
| `spec.podTemplate.spec.resourceClaims[].resourceClaimName` | `string` |  |  |  |
| `spec.podTemplate.spec.resourceClaims[].resourceClaimTemplateName` | `string` |  |  |  |
| `spec.podTemplate.spec.resources` | `KubernetesCoreV1ApiResourceRequirements` |  |  |  |
| `spec.podTemplate.spec.resources.limits` | `map<string, string>` |  |  |  |
| `spec.podTemplate.spec.resources.requests` | `map<string, string>` |  |  |  |
| `spec.podTemplate.spec.resources.claims` | `[]KubernetesCoreV1ApiResourceClaim` |  |  |  |
| `spec.podTemplate.spec.resources.claims[].name` | `string` | yes |  |  |
| `spec.podTemplate.spec.resources.claims[].request` | `string` |  |  |  |
| `spec.podTemplate.spec.hostnameOverride` | `string` |  |  |  |
| `spec.podTemplate.spec.schedulingGroup` | `KubernetesCoreV1ApiPodSchedulingGroup` |  |  |  |
| `spec.podTemplate.spec.schedulingGroup.podGroupName` | `string` |  |  |  |
| `spec.podTemplate.spec.evictionResponders` | `[]KubernetesCoreV1ApiEvictionResponder` |  |  |  |
| `spec.podTemplate.spec.evictionResponders[].name` | `string` | yes |  |  |
| `spec.podTemplate.spec.evictionResponders[].priority` | `int32` | yes |  |  |
| `spec.podTemplate.metadata` | `KubernetesAgentSandboxApiPodMetadata` |  |  |  |
| `spec.podTemplate.metadata.labels` | `map<string, string>` |  |  |  |
| `spec.podTemplate.metadata.annotations` | `map<string, string>` |  |  |  |
| `spec.volumeClaimTemplates` | `[]KubernetesAgentSandboxApiPersistentVolumeClaimTemplate` |  |  |  |
| `spec.volumeClaimTemplates[].metadata` | `KubernetesAgentSandboxApiEmbeddedObjectMetadata` | yes |  |  |
| `spec.volumeClaimTemplates[].metadata.name` | `string` | yes |  |  |
| `spec.volumeClaimTemplates[].metadata.labels` | `map<string, string>` |  |  |  |
| `spec.volumeClaimTemplates[].metadata.annotations` | `map<string, string>` |  |  |  |
| `spec.volumeClaimTemplates[].spec` | `KubernetesCoreV1ApiPersistentVolumeClaimSpec` | yes |  |  |
| `spec.volumeClaimTemplates[].spec.accessModes` | `[]string` |  |  |  |
| `spec.volumeClaimTemplates[].spec.selector` | `KubernetesCoreV1ApiLabelSelector` |  |  |  |
| `spec.volumeClaimTemplates[].spec.selector.matchLabels` | `map<string, string>` |  |  |  |
| `spec.volumeClaimTemplates[].spec.selector.matchExpressions` | `[]KubernetesCoreV1ApiLabelSelectorRequirement` |  |  |  |
| `spec.volumeClaimTemplates[].spec.selector.matchExpressions[].key` | `string` | yes |  |  |
| `spec.volumeClaimTemplates[].spec.selector.matchExpressions[].operator` | `string` | yes |  |  |
| `spec.volumeClaimTemplates[].spec.selector.matchExpressions[].values` | `[]string` |  |  |  |
| `spec.volumeClaimTemplates[].spec.resources` | `KubernetesCoreV1ApiVolumeResourceRequirements` |  |  |  |
| `spec.volumeClaimTemplates[].spec.resources.limits` | `map<string, string>` |  |  |  |
| `spec.volumeClaimTemplates[].spec.resources.requests` | `map<string, string>` |  |  |  |
| `spec.volumeClaimTemplates[].spec.volumeName` | `string` |  |  |  |
| `spec.volumeClaimTemplates[].spec.storageClassName` | `string \| valueFrom` |  |  | KubernetesStorageClass (`status.outputs.storage_class_name`) |
| `spec.volumeClaimTemplates[].spec.volumeMode` | `string` |  |  |  |
| `spec.volumeClaimTemplates[].spec.dataSource` | `KubernetesCoreV1ApiTypedLocalObjectReference` |  |  |  |
| `spec.volumeClaimTemplates[].spec.dataSource.apiGroup` | `string` |  |  |  |
| `spec.volumeClaimTemplates[].spec.dataSource.kind` | `string` | yes |  |  |
| `spec.volumeClaimTemplates[].spec.dataSource.name` | `string` | yes |  |  |
| `spec.volumeClaimTemplates[].spec.dataSourceRef` | `KubernetesCoreV1ApiTypedObjectReference` |  |  |  |
| `spec.volumeClaimTemplates[].spec.dataSourceRef.apiGroup` | `string` |  |  |  |
| `spec.volumeClaimTemplates[].spec.dataSourceRef.kind` | `string` | yes |  |  |
| `spec.volumeClaimTemplates[].spec.dataSourceRef.name` | `string` | yes |  |  |
| `spec.volumeClaimTemplates[].spec.dataSourceRef.namespace` | `string` |  |  |  |
| `spec.volumeClaimTemplates[].spec.volumeAttributesClassName` | `string` |  |  |  |
| `spec.service` | `bool` |  |  |  |
| `spec.shutdownTime` | `string` |  |  |  |
| `spec.shutdownPolicy` | `string` |  |  |  |
| `spec.operatingMode` | `string` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Kubernetes namespace the Sandbox is created in; its pod, Service and
claims live there too. Typically a reference to a KubernetesNamespace
resource's `spec.name`.

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.labels

`map<string, string>`

Labels on the Sandbox object itself (its metadata.labels), merged under
Planton's identity labels (a key in both keeps the identity value). The
pod's labels are `pod_template.metadata.labels`.

- rule: {"map":{"keys":{"string":{"minLen":"1","maxLen":"317","pattern":"^([a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*/)?[A-Za-z0-9]([-A-Za-z0-9_.]{0,61}[A-Za-z0-9])?$"}},"values":{"string":{"maxLen":"63","pattern":"^([A-Za-z0-9]([-A-Za-z0-9_.]{0,61}[A-Za-z0-9])?)?$"}}}}

### spec.annotations

`map<string, string>`

Annotations on the Sandbox object itself (its metadata.annotations).

### spec.podTemplate

`KubernetesAgentSandboxApiPodTemplate` · required

The pod the sandbox runs: Kubernetes' full pod specification, and the
labels and annotations the pod carries. The controller names the pod
after the sandbox and recreates it if it is deleted while the sandbox is
Running.

- rule: {"required":true}

### spec.podTemplate.spec

`KubernetesCoreV1ApiPodSpec` · required

The pod's specification: containers, volumes, security, scheduling, the
runtime class that isolates it (gVisor's `gvisor`, Kata's `kata-qemu`).

When the sandbox comes from a template (a warm pool member or a claimed
sandbox) and `automountServiceAccountToken` is unset, the controller sets
it to false, so a sandbox holds no Kubernetes credentials unless asked.
An agent's code rarely needs them; give it a ServiceAccount of its own
when it does.

- rule: {"required":true}

### spec.podTemplate.spec.volumes

`[]KubernetesCoreV1ApiVolume`

List of volumes that can be mounted by containers belonging to the pod.
More info: https://kubernetes.io/docs/concepts/storage/volumes

- rule: Set at most one of awsElasticBlockStore, azureDisk, azureFile, cephfs, cinder, configMap, csi, downwardAPI, emptyDir, ephemeral, fc, flexVolume, flocker, gcePersistentDisk, gitRepo, glusterfs, hostPath, image, iscsi, nfs, persistentVolumeClaim, photonPersistentDisk, portworxVolume, projected, quobyte, rbd, scaleIO, secret, storageos, vsphereVolume

### spec.podTemplate.spec.volumes[].name

`string` · required

name of the volume. Must be a DNS_LABEL and unique within the pod. More
info:
https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.volumes[].hostPath

`KubernetesCoreV1ApiHostPathVolumeSource`

hostPath represents a pre-existing file or directory on the host machine
that is directly exposed to the container. This is generally used for
system agents or other privileged things that are allowed to see the host
machine. Most containers will NOT need this. More info:
https://kubernetes.io/docs/concepts/storage/volumes#hostpath mount host
directories as read/write.

### spec.podTemplate.spec.volumes[].hostPath.path

`string` · required

path of the directory on the host. If the path is a symlink, it will follow
the link to the real path. More info:
https://kubernetes.io/docs/concepts/storage/volumes#hostpath

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.volumes[].hostPath.type

`string` · optional (explicit presence)

type for HostPath Volume Defaults to "" More info:
https://kubernetes.io/docs/concepts/storage/volumes#hostpath

- rule: {"string":{"in":["","DirectoryOrCreate","Directory","FileOrCreate","File","Socket","CharDevice","BlockDevice"]}}

### spec.podTemplate.spec.volumes[].emptyDir

`KubernetesCoreV1ApiEmptyDirVolumeSource`

emptyDir represents a temporary directory that shares a pod's lifetime.
More info: https://kubernetes.io/docs/concepts/storage/volumes#emptydir

### spec.podTemplate.spec.volumes[].emptyDir.medium

`string` · optional (explicit presence)

medium represents what type of storage medium should back this directory.
The default is "" which means to use the node's default medium. Must be an
empty string (default) or Memory. More info:
https://kubernetes.io/docs/concepts/storage/volumes#emptydir

- rule: {"string":{"pattern":"^(Memory|HugePages(-[0-9]+[KMGTPE]i)?)?$"}}

### spec.podTemplate.spec.volumes[].emptyDir.sizeLimit

`string` · optional (explicit presence)

sizeLimit is the total amount of local storage required for this EmptyDir
volume. The size limit is also applicable for memory medium. The maximum
usage on memory medium EmptyDir would be the minimum value between the
SizeLimit specified here and the sum of memory limits of all containers in
a pod. The default is nil which means that the limit is undefined. More
info: https://kubernetes.io/docs/concepts/storage/volumes#emptydir

- rule: {"string":{"pattern":"^(\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))))?$"}}

### spec.podTemplate.spec.volumes[].emptyDir.mode

`int32` · optional (explicit presence)

mode specifies the permission bits for the emptyDir directory, in numeric
notation (e.g., 0755, 01777). Must be a value between 0000 and 01777. If
not specified, defaults to 0777. This might be in conflict with other
options that affect the file mode, like fsGroup. If fsGroup is specified,
the fsGroup permissions will override the mode specified here. This field
has no effect on Windows. This field is alpha and requires
EmptyDirVolumeMode featuregate to be enabled.

### spec.podTemplate.spec.volumes[].gcePersistentDisk

`KubernetesCoreV1ApiGCEPersistentDiskVolumeSource`

gcePersistentDisk represents a GCE Disk resource that is attached to a
kubelet's host machine and then exposed to the pod. Deprecated:
GCEPersistentDisk is deprecated. All operations for the in-tree
gcePersistentDisk type are redirected to the pd.csi.storage.gke.io CSI
driver. More info:
https://kubernetes.io/docs/concepts/storage/volumes#gcepersistentdisk

### spec.podTemplate.spec.volumes[].gcePersistentDisk.pdName

`string` · required

pdName is unique name of the PD resource in GCE. Used to identify the disk
in GCE. More info:
https://kubernetes.io/docs/concepts/storage/volumes#gcepersistentdisk

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.volumes[].gcePersistentDisk.fsType

`string`

fsType is filesystem type of the volume that you want to mount. Tip: Ensure
that the filesystem type is supported by the host operating system.
Examples: "ext4", "xfs", "ntfs". Implicitly inferred to be "ext4" if
unspecified. More info:
https://kubernetes.io/docs/concepts/storage/volumes#gcepersistentdisk

### spec.podTemplate.spec.volumes[].gcePersistentDisk.partition

`int32`

partition is the partition in the volume that you want to mount. If
omitted, the default is to mount by volume name. Examples: For volume
/dev/sda1, you specify the partition as "1". Similarly, the volume
partition for /dev/sda is "0" (or you can leave the property empty). More
info: https://kubernetes.io/docs/concepts/storage/volumes#gcepersistentdisk

### spec.podTemplate.spec.volumes[].gcePersistentDisk.readOnly

`bool`

readOnly here will force the ReadOnly setting in VolumeMounts. Defaults to
false. More info:
https://kubernetes.io/docs/concepts/storage/volumes#gcepersistentdisk

### spec.podTemplate.spec.volumes[].awsElasticBlockStore

`KubernetesCoreV1ApiAWSElasticBlockStoreVolumeSource`

awsElasticBlockStore represents an AWS Disk resource that is attached to a
kubelet's host machine and then exposed to the pod. Deprecated:
AWSElasticBlockStore is deprecated. All operations for the in-tree
awsElasticBlockStore type are redirected to the ebs.csi.aws.com CSI driver.
More info:
https://kubernetes.io/docs/concepts/storage/volumes#awselasticblockstore

### spec.podTemplate.spec.volumes[].awsElasticBlockStore.volumeID

`string` · required

volumeID is unique ID of the persistent disk resource in AWS (Amazon EBS
volume). More info:
https://kubernetes.io/docs/concepts/storage/volumes#awselasticblockstore

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.volumes[].awsElasticBlockStore.fsType

`string`

fsType is the filesystem type of the volume that you want to mount. Tip:
Ensure that the filesystem type is supported by the host operating system.
Examples: "ext4", "xfs", "ntfs". Implicitly inferred to be "ext4" if
unspecified. More info:
https://kubernetes.io/docs/concepts/storage/volumes#awselasticblockstore

### spec.podTemplate.spec.volumes[].awsElasticBlockStore.partition

`int32`

partition is the partition in the volume that you want to mount. If
omitted, the default is to mount by volume name. Examples: For volume
/dev/sda1, you specify the partition as "1". Similarly, the volume
partition for /dev/sda is "0" (or you can leave the property empty).

### spec.podTemplate.spec.volumes[].awsElasticBlockStore.readOnly

`bool`

readOnly value true will force the readOnly setting in VolumeMounts. More
info:
https://kubernetes.io/docs/concepts/storage/volumes#awselasticblockstore

### spec.podTemplate.spec.volumes[].gitRepo

`KubernetesCoreV1ApiGitRepoVolumeSource`

gitRepo represents a git repository at a particular revision. Deprecated:
GitRepo is deprecated. To provision a container with a git repo, mount an
EmptyDir into an InitContainer that clones the repo using git, then mount
the EmptyDir into the Pod's container.

### spec.podTemplate.spec.volumes[].gitRepo.repository

`string` · required

repository is the URL.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.volumes[].gitRepo.revision

`string`

revision is the commit hash for the specified revision.

### spec.podTemplate.spec.volumes[].gitRepo.directory

`string`

directory is the target directory name. Must not contain or start with
'..'.  If '.' is supplied, the volume directory will be the git repository.
Otherwise, if specified, the volume will contain the git repository in the
subdirectory with the given name.

### spec.podTemplate.spec.volumes[].secret

`KubernetesCoreV1ApiSecretVolumeSource`

secret represents a secret that should populate this volume. More info:
https://kubernetes.io/docs/concepts/storage/volumes#secret

### spec.podTemplate.spec.volumes[].secret.secretName

`string | valueFrom`

secretName is the name of the secret in the pod's namespace to use. More
info: https://kubernetes.io/docs/concepts/storage/volumes#secret

Defaults to a KubernetesSecret reference (its secret_name output); pass the
literal name with `value:` for one Planton does not manage.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.podTemplate.spec.volumes[].secret.items

`[]KubernetesCoreV1ApiKeyToPath`

items If unspecified, each key-value pair in the Data field of the
referenced Secret will be projected into the volume as a file whose name is
the key and content is the value. If specified, the listed keys will be
projected into the specified paths, and unlisted keys will not be present.
If a key is specified which is not present in the Secret, the volume setup
will error unless it is marked optional. Paths must be relative and may not
contain the '..' path or start with '..'.

### spec.podTemplate.spec.volumes[].secret.items[].key

`string` · required

key is the key to project.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.volumes[].secret.items[].path

`string` · required

path is the relative path of the file to map the key to. May not be an
absolute path. May not contain the path element '..'. May not start with
the string '..'.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.volumes[].secret.items[].mode

`int32` · optional (explicit presence)

mode is Optional: mode bits used to set permissions on this file. Must be
an octal value between 0000 and 0777 or a decimal value between 0 and 511.
YAML accepts both octal and decimal values, JSON requires decimal values
for mode bits. If not specified, the volume defaultMode will be used. This
might be in conflict with other options that affect the file mode, like
fsGroup, and the result can be other mode bits set.

### spec.podTemplate.spec.volumes[].secret.items[].user

`int64` · optional (explicit presence)

user is Optional: The owner UID of the created file. If specified, the
item-level user field takes precedence over defaultUser. (Alpha) This field
requires the AtomicWriteVolumeUserFields feature gate to be enabled.

### spec.podTemplate.spec.volumes[].secret.defaultMode

`int32` · optional (explicit presence)

defaultMode is Optional: mode bits used to set permissions on created files
by default. Must be an octal value between 0000 and 0777 or a decimal value
between 0 and 511. YAML accepts both octal and decimal values, JSON
requires decimal values for mode bits. Defaults to 0644. Directories within
the path are not affected by this setting. This might be in conflict with
other options that affect the file mode, like fsGroup, and the result can
be other mode bits set.

### spec.podTemplate.spec.volumes[].secret.optional

`bool` · optional (explicit presence)

optional field specify whether the Secret or its keys must be defined.

### spec.podTemplate.spec.volumes[].secret.defaultUser

`int64` · optional (explicit presence)

defaultUser is Optional: The owner UID of the created files by default. The
defaultUser field is only used as a fallback when the item-level user field
is unset. (Alpha) This field requires the AtomicWriteVolumeUserFields
feature gate to be enabled.

### spec.podTemplate.spec.volumes[].nfs

`KubernetesCoreV1ApiNFSVolumeSource`

nfs represents an NFS mount on the host that shares a pod's lifetime More
info: https://kubernetes.io/docs/concepts/storage/volumes#nfs

### spec.podTemplate.spec.volumes[].nfs.server

`string` · required

server is the hostname or IP address of the NFS server. More info:
https://kubernetes.io/docs/concepts/storage/volumes#nfs

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.volumes[].nfs.path

`string` · required

path that is exported by the NFS server. More info:
https://kubernetes.io/docs/concepts/storage/volumes#nfs

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.volumes[].nfs.readOnly

`bool`

readOnly here will force the NFS export to be mounted with read-only
permissions. Defaults to false. More info:
https://kubernetes.io/docs/concepts/storage/volumes#nfs

### spec.podTemplate.spec.volumes[].iscsi

`KubernetesCoreV1ApiISCSIVolumeSource`

iscsi represents an ISCSI Disk resource that is attached to a kubelet's
host machine and then exposed to the pod. More info:
https://kubernetes.io/docs/concepts/storage/volumes/#iscsi

### spec.podTemplate.spec.volumes[].iscsi.targetPortal

`string` · required

targetPortal is iSCSI Target Portal. The Portal is either an IP or
ip_addr:port if the port is other than default (typically TCP ports 860 and
3260).

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.volumes[].iscsi.iqn

`string` · required

iqn is the target iSCSI Qualified Name.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.volumes[].iscsi.lun

`int32` · required · optional (explicit presence)

lun represents iSCSI Target Lun number.

- rule: {"required":true}

### spec.podTemplate.spec.volumes[].iscsi.iscsiInterface

`string`

iscsiInterface is the interface Name that uses an iSCSI transport. Defaults
to 'default' (tcp).

Unset, the cluster writes "default".

### spec.podTemplate.spec.volumes[].iscsi.fsType

`string`

fsType is the filesystem type of the volume that you want to mount. Tip:
Ensure that the filesystem type is supported by the host operating system.
Examples: "ext4", "xfs", "ntfs". Implicitly inferred to be "ext4" if
unspecified. More info:
https://kubernetes.io/docs/concepts/storage/volumes#iscsi

### spec.podTemplate.spec.volumes[].iscsi.readOnly

`bool`

readOnly here will force the ReadOnly setting in VolumeMounts. Defaults to
false.

### spec.podTemplate.spec.volumes[].iscsi.portals

`[]string`

portals is the iSCSI Target Portal List. The portal is either an IP or
ip_addr:port if the port is other than default (typically TCP ports 860 and
3260).

### spec.podTemplate.spec.volumes[].iscsi.chapAuthDiscovery

`bool`

chapAuthDiscovery defines whether support iSCSI Discovery CHAP
authentication.

### spec.podTemplate.spec.volumes[].iscsi.chapAuthSession

`bool`

chapAuthSession defines whether support iSCSI Session CHAP authentication.

### spec.podTemplate.spec.volumes[].iscsi.secretRef

`KubernetesCoreV1ApiLocalObjectReference`

secretRef is the CHAP Secret for iSCSI target and initiator authentication.

### spec.podTemplate.spec.volumes[].iscsi.secretRef.name

`string | valueFrom` · required

`name` of the referent. This field is effectively required, but due to
backwards compatibility is allowed to be empty. Instances of this type with
an empty value here are almost certainly wrong. More info:
https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names

Defaults to a KubernetesSecret reference (its secret_name output); pass the
literal name with `value:` for one Planton does not manage.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.podTemplate.spec.volumes[].iscsi.initiatorName

`string` · optional (explicit presence)

initiatorName is the custom iSCSI Initiator Name. If initiatorName is
specified with iscsiInterface simultaneously, new iSCSI interface <target
portal>:<volume name> will be created for the connection.

### spec.podTemplate.spec.volumes[].glusterfs

`KubernetesCoreV1ApiGlusterfsVolumeSource`

glusterfs represents a Glusterfs mount on the host that shares a pod's
lifetime. Deprecated: Glusterfs is deprecated and the in-tree glusterfs
type is no longer supported.

### spec.podTemplate.spec.volumes[].glusterfs.endpoints

`string` · required

endpoints is the endpoint name that details Glusterfs topology.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.volumes[].glusterfs.path

`string` · required

path is the Glusterfs volume path. More info:
https://examples.k8s.io/volumes/glusterfs/README.md#create-a-pod

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.volumes[].glusterfs.readOnly

`bool`

readOnly here will force the Glusterfs volume to be mounted with read-only
permissions. Defaults to false. More info:
https://examples.k8s.io/volumes/glusterfs/README.md#create-a-pod

### spec.podTemplate.spec.volumes[].persistentVolumeClaim

`KubernetesCoreV1ApiPersistentVolumeClaimVolumeSource`

persistentVolumeClaimVolumeSource represents a reference to a
PersistentVolumeClaim in the same namespace. More info:
https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims

### spec.podTemplate.spec.volumes[].persistentVolumeClaim.claimName

`string | valueFrom` · required

claimName is the name of a PersistentVolumeClaim in the same namespace as
the pod using this volume. More info:
https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims

Defaults to a KubernetesPersistentVolumeClaim reference (its pvc_name
output); pass the literal name with `value:` for one Planton does not
manage.

- references: KubernetesPersistentVolumeClaim (`status.outputs.pvc_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesPersistentVolumeClaim, name: <that resource's name>, fieldPath: status.outputs.pvc_name}} -- a bare string does not parse

### spec.podTemplate.spec.volumes[].persistentVolumeClaim.readOnly

`bool`

readOnly Will force the ReadOnly setting in VolumeMounts. Default false.

### spec.podTemplate.spec.volumes[].rbd

`KubernetesCoreV1ApiRBDVolumeSource`

rbd represents a Rados Block Device mount on the host that shares a pod's
lifetime. Deprecated: RBD is deprecated and the in-tree rbd type is no
longer supported.

### spec.podTemplate.spec.volumes[].rbd.monitors

`[]string` · required

monitors is a collection of Ceph monitors. More info:
https://examples.k8s.io/volumes/rbd/README.md#how-to-use-it

- rule: {"repeated":{"minItems":"1"}}

### spec.podTemplate.spec.volumes[].rbd.image

`string` · required

image is the rados image name. More info:
https://examples.k8s.io/volumes/rbd/README.md#how-to-use-it

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.volumes[].rbd.fsType

`string`

fsType is the filesystem type of the volume that you want to mount. Tip:
Ensure that the filesystem type is supported by the host operating system.
Examples: "ext4", "xfs", "ntfs". Implicitly inferred to be "ext4" if
unspecified. More info:
https://kubernetes.io/docs/concepts/storage/volumes#rbd

### spec.podTemplate.spec.volumes[].rbd.pool

`string`

pool is the rados pool name. Default is rbd. More info:
https://examples.k8s.io/volumes/rbd/README.md#how-to-use-it

Unset, the cluster writes "rbd".

### spec.podTemplate.spec.volumes[].rbd.user

`string`

user is the rados user name. Default is admin. More info:
https://examples.k8s.io/volumes/rbd/README.md#how-to-use-it

Unset, the cluster writes "admin".

### spec.podTemplate.spec.volumes[].rbd.keyring

`string`

keyring is the path to key ring for RBDUser. Default is /etc/ceph/keyring.
More info: https://examples.k8s.io/volumes/rbd/README.md#how-to-use-it

Unset, the cluster writes "/etc/ceph/keyring".

### spec.podTemplate.spec.volumes[].rbd.secretRef

`KubernetesCoreV1ApiLocalObjectReference`

secretRef is name of the authentication secret for RBDUser. If provided
overrides keyring. Default is nil. More info:
https://examples.k8s.io/volumes/rbd/README.md#how-to-use-it

### spec.podTemplate.spec.volumes[].rbd.secretRef.name

`string | valueFrom` · required

`name` of the referent. This field is effectively required, but due to
backwards compatibility is allowed to be empty. Instances of this type with
an empty value here are almost certainly wrong. More info:
https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names

Defaults to a KubernetesSecret reference (its secret_name output); pass the
literal name with `value:` for one Planton does not manage.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.podTemplate.spec.volumes[].rbd.readOnly

`bool`

readOnly here will force the ReadOnly setting in VolumeMounts. Defaults to
false. More info:
https://examples.k8s.io/volumes/rbd/README.md#how-to-use-it

### spec.podTemplate.spec.volumes[].flexVolume

`KubernetesCoreV1ApiFlexVolumeSource`

flexVolume represents a generic volume resource that is
provisioned/attached using an exec based plugin. Deprecated: FlexVolume is
deprecated. Consider using a CSIDriver instead.

### spec.podTemplate.spec.volumes[].flexVolume.driver

`string` · required

driver is the name of the driver to use for this volume.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.volumes[].flexVolume.fsType

`string`

fsType is the filesystem type to mount. Must be a filesystem type supported
by the host operating system. Ex. "ext4", "xfs", "ntfs". The default
filesystem depends on FlexVolume script.

### spec.podTemplate.spec.volumes[].flexVolume.secretRef

`KubernetesCoreV1ApiLocalObjectReference`

secretRef is Optional: secretRef is reference to the secret object
containing sensitive information to pass to the plugin scripts. This may be
empty if no secret object is specified. If the secret object contains more
than one secret, all secrets are passed to the plugin scripts.

### spec.podTemplate.spec.volumes[].flexVolume.secretRef.name

`string | valueFrom` · required

`name` of the referent. This field is effectively required, but due to
backwards compatibility is allowed to be empty. Instances of this type with
an empty value here are almost certainly wrong. More info:
https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names

Defaults to a KubernetesSecret reference (its secret_name output); pass the
literal name with `value:` for one Planton does not manage.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.podTemplate.spec.volumes[].flexVolume.readOnly

`bool`

readOnly is Optional: defaults to false (read/write). ReadOnly here will
force the ReadOnly setting in VolumeMounts.

### spec.podTemplate.spec.volumes[].flexVolume.options

`map<string, string>`

options is Optional: this field holds extra command options if any.

### spec.podTemplate.spec.volumes[].cinder

`KubernetesCoreV1ApiCinderVolumeSource`

cinder represents a cinder volume attached and mounted on kubelets host
machine. Deprecated: Cinder is deprecated. All operations for the in-tree
cinder type are redirected to the cinder.csi.openstack.org CSI driver. More
info: https://examples.k8s.io/mysql-cinder-pd/README.md

### spec.podTemplate.spec.volumes[].cinder.volumeID

`string` · required

volumeID used to identify the volume in cinder. More info:
https://examples.k8s.io/mysql-cinder-pd/README.md

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.volumes[].cinder.fsType

`string`

fsType is the filesystem type to mount. Must be a filesystem type supported
by the host operating system. Examples: "ext4", "xfs", "ntfs". Implicitly
inferred to be "ext4" if unspecified. More info:
https://examples.k8s.io/mysql-cinder-pd/README.md

### spec.podTemplate.spec.volumes[].cinder.readOnly

`bool`

readOnly defaults to false (read/write). ReadOnly here will force the
ReadOnly setting in VolumeMounts. More info:
https://examples.k8s.io/mysql-cinder-pd/README.md

### spec.podTemplate.spec.volumes[].cinder.secretRef

`KubernetesCoreV1ApiLocalObjectReference`

secretRef is optional: points to a secret object containing parameters used
to connect to OpenStack.

### spec.podTemplate.spec.volumes[].cinder.secretRef.name

`string | valueFrom` · required

`name` of the referent. This field is effectively required, but due to
backwards compatibility is allowed to be empty. Instances of this type with
an empty value here are almost certainly wrong. More info:
https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names

Defaults to a KubernetesSecret reference (its secret_name output); pass the
literal name with `value:` for one Planton does not manage.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.podTemplate.spec.volumes[].cephfs

`KubernetesCoreV1ApiCephFSVolumeSource`

cephFS represents a Ceph FS mount on the host that shares a pod's lifetime.
Deprecated: CephFS is deprecated and the in-tree cephfs type is no longer
supported.

### spec.podTemplate.spec.volumes[].cephfs.monitors

`[]string` · required

monitors is Required: Monitors is a collection of Ceph monitors More info:
https://examples.k8s.io/volumes/cephfs/README.md#how-to-use-it

- rule: {"repeated":{"minItems":"1"}}

### spec.podTemplate.spec.volumes[].cephfs.path

`string`

path is Optional: Used as the mounted root, rather than the full Ceph tree,
default is /.

### spec.podTemplate.spec.volumes[].cephfs.user

`string`

user is optional: User is the rados user name, default is admin More info:
https://examples.k8s.io/volumes/cephfs/README.md#how-to-use-it

### spec.podTemplate.spec.volumes[].cephfs.secretFile

`string`

secretFile is Optional: SecretFile is the path to key ring for User,
default is /etc/ceph/user.secret More info:
https://examples.k8s.io/volumes/cephfs/README.md#how-to-use-it

### spec.podTemplate.spec.volumes[].cephfs.secretRef

`KubernetesCoreV1ApiLocalObjectReference`

secretRef is Optional: SecretRef is reference to the authentication secret
for User, default is empty. More info:
https://examples.k8s.io/volumes/cephfs/README.md#how-to-use-it

### spec.podTemplate.spec.volumes[].cephfs.secretRef.name

`string | valueFrom` · required

`name` of the referent. This field is effectively required, but due to
backwards compatibility is allowed to be empty. Instances of this type with
an empty value here are almost certainly wrong. More info:
https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names

Defaults to a KubernetesSecret reference (its secret_name output); pass the
literal name with `value:` for one Planton does not manage.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.podTemplate.spec.volumes[].cephfs.readOnly

`bool`

readOnly is Optional: Defaults to false (read/write). ReadOnly here will
force the ReadOnly setting in VolumeMounts. More info:
https://examples.k8s.io/volumes/cephfs/README.md#how-to-use-it

### spec.podTemplate.spec.volumes[].flocker

`KubernetesCoreV1ApiFlockerVolumeSource`

flocker represents a Flocker volume attached to a kubelet's host machine.
This depends on the Flocker control service being running. Deprecated:
Flocker is deprecated and the in-tree flocker type is no longer supported.

### spec.podTemplate.spec.volumes[].flocker.datasetName

`string`

datasetName is Name of the dataset stored as metadata -> name on the
dataset for Flocker should be considered as deprecated.

### spec.podTemplate.spec.volumes[].flocker.datasetUUID

`string`

datasetUUID is the UUID of the dataset. This is unique identifier of a
Flocker dataset.

### spec.podTemplate.spec.volumes[].downwardAPI

`KubernetesCoreV1ApiDownwardAPIVolumeSource`

downwardAPI represents downward API about the pod that should populate this
volume.

### spec.podTemplate.spec.volumes[].downwardAPI.items

`[]KubernetesCoreV1ApiDownwardAPIVolumeFile`

`items` is a list of downward API volume file.

### spec.podTemplate.spec.volumes[].downwardAPI.items[].path

`string` · required

Required: Path is  the relative path name of the file to be created. Must
not be absolute or contain the '..' path. Must be utf-8 encoded. The first
item of the relative path must not start with '..'.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.volumes[].downwardAPI.items[].fieldRef

`KubernetesCoreV1ApiObjectFieldSelector`

Required: Selects a field of the pod: only annotations, labels, name,
namespace and uid are supported.

### spec.podTemplate.spec.volumes[].downwardAPI.items[].fieldRef.apiVersion

`string`

Version of the schema the FieldPath is written in terms of, defaults to
"v1".

### spec.podTemplate.spec.volumes[].downwardAPI.items[].fieldRef.fieldPath

`string` · required

Path of the field to select in the specified API version.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.volumes[].downwardAPI.items[].resourceFieldRef

`KubernetesCoreV1ApiResourceFieldSelector`

Selects a resource of the container: only resources limits and requests
(limits.cpu, limits.memory, requests.cpu and requests.memory) are currently
supported.

### spec.podTemplate.spec.volumes[].downwardAPI.items[].resourceFieldRef.containerName

`string`

Container name: required for volumes, optional for env vars.

### spec.podTemplate.spec.volumes[].downwardAPI.items[].resourceFieldRef.resource

`string` · required

Required: resource to select.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.volumes[].downwardAPI.items[].resourceFieldRef.divisor

`string` · optional (explicit presence)

Specifies the output format of the exposed resources, defaults to "1"

- rule: {"string":{"pattern":"^(\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))))?$"}}

### spec.podTemplate.spec.volumes[].downwardAPI.items[].mode

`int32` · optional (explicit presence)

Optional: mode bits used to set permissions on this file, must be an octal
value between 0000 and 0777 or a decimal value between 0 and 511. YAML
accepts both octal and decimal values, JSON requires decimal values for
mode bits. If not specified, the volume defaultMode will be used. This
might be in conflict with other options that affect the file mode, like
fsGroup, and the result can be other mode bits set.

### spec.podTemplate.spec.volumes[].downwardAPI.items[].user

`int64` · optional (explicit presence)

user is Optional: The owner UID of the created file. If specified, the
item-level user field takes precedence over defaultUser. (Alpha) This field
requires the AtomicWriteVolumeUserFields feature gate to be enabled.

### spec.podTemplate.spec.volumes[].downwardAPI.defaultMode

`int32` · optional (explicit presence)

Optional: mode bits to use on created files by default. Must be a Optional:
mode bits used to set permissions on created files by default. Must be an
octal value between 0000 and 0777 or a decimal value between 0 and 511.
YAML accepts both octal and decimal values, JSON requires decimal values
for mode bits. Defaults to 0644. Directories within the path are not
affected by this setting. This might be in conflict with other options that
affect the file mode, like fsGroup, and the result can be other mode bits
set.

### spec.podTemplate.spec.volumes[].downwardAPI.defaultUser

`int64` · optional (explicit presence)

defaultUser is Optional: The owner UID of the created files by default. The
defaultUser field is only used as a fallback when the item-level user field
is unset. (Alpha) This field requires the AtomicWriteVolumeUserFields
feature gate to be enabled.

### spec.podTemplate.spec.volumes[].fc

`KubernetesCoreV1ApiFCVolumeSource`

fc represents a Fibre Channel resource that is attached to a kubelet's host
machine and then exposed to the pod.

### spec.podTemplate.spec.volumes[].fc.targetWWNs

`[]string`

targetWWNs is Optional: FC target worldwide names (WWNs)

### spec.podTemplate.spec.volumes[].fc.lun

`int32` · optional (explicit presence)

lun is Optional: FC target lun number.

### spec.podTemplate.spec.volumes[].fc.fsType

`string`

fsType is the filesystem type to mount. Must be a filesystem type supported
by the host operating system. Ex. "ext4", "xfs", "ntfs". Implicitly
inferred to be "ext4" if unspecified.

### spec.podTemplate.spec.volumes[].fc.readOnly

`bool`

readOnly is Optional: Defaults to false (read/write). ReadOnly here will
force the ReadOnly setting in VolumeMounts.

### spec.podTemplate.spec.volumes[].fc.wwids

`[]string`

wwids Optional: FC volume world wide identifiers (wwids) Either wwids or
combination of targetWWNs and lun must be set, but not both simultaneously.

### spec.podTemplate.spec.volumes[].azureFile

`KubernetesCoreV1ApiAzureFileVolumeSource`

azureFile represents an Azure File Service mount on the host and bind mount
to the pod. Deprecated: AzureFile is deprecated. All operations for the
in-tree azureFile type are redirected to the file.csi.azure.com CSI driver.

### spec.podTemplate.spec.volumes[].azureFile.secretName

`string | valueFrom` · required

secretName is the  name of secret that contains Azure Storage Account Name
and Key.

Defaults to a KubernetesSecret reference (its secret_name output); pass the
literal name with `value:` for one Planton does not manage.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.podTemplate.spec.volumes[].azureFile.shareName

`string` · required

shareName is the azure share Name.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.volumes[].azureFile.readOnly

`bool`

readOnly defaults to false (read/write). ReadOnly here will force the
ReadOnly setting in VolumeMounts.

### spec.podTemplate.spec.volumes[].configMap

`KubernetesCoreV1ApiConfigMapVolumeSource`

configMap represents a configMap that should populate this volume.

### spec.podTemplate.spec.volumes[].configMap.name

`string | valueFrom` · required

`name` of the referent. This field is effectively required, but due to
backwards compatibility is allowed to be empty. Instances of this type with
an empty value here are almost certainly wrong. More info:
https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names

Defaults to a KubernetesConfigMap reference (its configmap_name output);
pass the literal name with `value:` for one Planton does not manage.

- references: KubernetesConfigMap (`status.outputs.configmap_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesConfigMap, name: <that resource's name>, fieldPath: status.outputs.configmap_name}} -- a bare string does not parse

### spec.podTemplate.spec.volumes[].configMap.items

`[]KubernetesCoreV1ApiKeyToPath`

items if unspecified, each key-value pair in the Data field of the
referenced ConfigMap will be projected into the volume as a file whose name
is the key and content is the value. If specified, the listed keys will be
projected into the specified paths, and unlisted keys will not be present.
If a key is specified which is not present in the ConfigMap, the volume
setup will error unless it is marked optional. Paths must be relative and
may not contain the '..' path or start with '..'.

### spec.podTemplate.spec.volumes[].configMap.items[].key

`string` · required

key is the key to project.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.volumes[].configMap.items[].path

`string` · required

path is the relative path of the file to map the key to. May not be an
absolute path. May not contain the path element '..'. May not start with
the string '..'.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.volumes[].configMap.items[].mode

`int32` · optional (explicit presence)

mode is Optional: mode bits used to set permissions on this file. Must be
an octal value between 0000 and 0777 or a decimal value between 0 and 511.
YAML accepts both octal and decimal values, JSON requires decimal values
for mode bits. If not specified, the volume defaultMode will be used. This
might be in conflict with other options that affect the file mode, like
fsGroup, and the result can be other mode bits set.

### spec.podTemplate.spec.volumes[].configMap.items[].user

`int64` · optional (explicit presence)

user is Optional: The owner UID of the created file. If specified, the
item-level user field takes precedence over defaultUser. (Alpha) This field
requires the AtomicWriteVolumeUserFields feature gate to be enabled.

### spec.podTemplate.spec.volumes[].configMap.defaultMode

`int32` · optional (explicit presence)

defaultMode is optional: mode bits used to set permissions on created files
by default. Must be an octal value between 0000 and 0777 or a decimal value
between 0 and 511. YAML accepts both octal and decimal values, JSON
requires decimal values for mode bits. Defaults to 0644. Directories within
the path are not affected by this setting. This might be in conflict with
other options that affect the file mode, like fsGroup, and the result can
be other mode bits set.

### spec.podTemplate.spec.volumes[].configMap.optional

`bool` · optional (explicit presence)

optional specify whether the ConfigMap or its keys must be defined.

### spec.podTemplate.spec.volumes[].configMap.defaultUser

`int64` · optional (explicit presence)

defaultUser is Optional: The owner UID of the created files by default. The
defaultUser field is only used as a fallback when the item-level user field
is unset. (Alpha) This field requires the AtomicWriteVolumeUserFields
feature gate to be enabled.

### spec.podTemplate.spec.volumes[].vsphereVolume

`KubernetesCoreV1ApiVsphereVirtualDiskVolumeSource`

vsphereVolume represents a vSphere volume attached and mounted on kubelets
host machine. Deprecated: VsphereVolume is deprecated. All operations for
the in-tree vsphereVolume type are redirected to the csi.vsphere.vmware.com
CSI driver.

### spec.podTemplate.spec.volumes[].vsphereVolume.volumePath

`string` · required

volumePath is the path that identifies vSphere volume vmdk.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.volumes[].vsphereVolume.fsType

`string`

fsType is filesystem type to mount. Must be a filesystem type supported by
the host operating system. Ex. "ext4", "xfs", "ntfs". Implicitly inferred
to be "ext4" if unspecified.

### spec.podTemplate.spec.volumes[].vsphereVolume.storagePolicyName

`string`

storagePolicyName is the storage Policy Based Management (SPBM) profile
name.

### spec.podTemplate.spec.volumes[].vsphereVolume.storagePolicyID

`string`

storagePolicyID is the storage Policy Based Management (SPBM) profile ID
associated with the StoragePolicyName.

### spec.podTemplate.spec.volumes[].quobyte

`KubernetesCoreV1ApiQuobyteVolumeSource`

quobyte represents a Quobyte mount on the host that shares a pod's
lifetime. Deprecated: Quobyte is deprecated and the in-tree quobyte type is
no longer supported.

### spec.podTemplate.spec.volumes[].quobyte.registry

`string` · required

registry represents a single or multiple Quobyte Registry services
specified as a string as host:port pair (multiple entries are separated
with commas) which acts as the central registry for volumes.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.volumes[].quobyte.volume

`string` · required

volume is a string that references an already created Quobyte volume by
name.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.volumes[].quobyte.readOnly

`bool`

readOnly here will force the Quobyte volume to be mounted with read-only
permissions. Defaults to false.

### spec.podTemplate.spec.volumes[].quobyte.user

`string`

user to map volume access to Defaults to serivceaccount user.

### spec.podTemplate.spec.volumes[].quobyte.group

`string`

group to map volume access to Default is no group.

### spec.podTemplate.spec.volumes[].quobyte.tenant

`string`

tenant owning the given Quobyte volume in the Backend Used with dynamically
provisioned Quobyte volumes, value is set by the plugin.

### spec.podTemplate.spec.volumes[].azureDisk

`KubernetesCoreV1ApiAzureDiskVolumeSource`

azureDisk represents an Azure Data Disk mount on the host and bind mount to
the pod. Deprecated: AzureDisk is deprecated. All operations for the
in-tree azureDisk type are redirected to the disk.csi.azure.com CSI driver.

### spec.podTemplate.spec.volumes[].azureDisk.diskName

`string` · required

diskName is the Name of the data disk in the blob storage.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.volumes[].azureDisk.diskURI

`string` · required

diskURI is the URI of data disk in the blob storage.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.volumes[].azureDisk.cachingMode

`string` · optional (explicit presence)

cachingMode is the Host Caching mode: None, Read Only, Read Write.

- rule: {"string":{"in":["None","ReadOnly","ReadWrite"]}}

### spec.podTemplate.spec.volumes[].azureDisk.fsType

`string` · optional (explicit presence)

fsType is Filesystem type to mount. Must be a filesystem type supported by
the host operating system. Ex. "ext4", "xfs", "ntfs". Implicitly inferred
to be "ext4" if unspecified.

Unset, the cluster writes "ext4".

### spec.podTemplate.spec.volumes[].azureDisk.readOnly

`bool` · optional (explicit presence)

readOnly Defaults to false (read/write). ReadOnly here will force the
ReadOnly setting in VolumeMounts.

Unset, the cluster writes false.

### spec.podTemplate.spec.volumes[].azureDisk.kind

`string` · optional (explicit presence)

kind expected values are Shared: multiple blob disks per storage account
Dedicated: single blob disk per storage account  Managed: azure managed
data disk (only in managed availability set). defaults to shared.

- rule: {"string":{"in":["Shared","Dedicated","Managed"]}}

### spec.podTemplate.spec.volumes[].photonPersistentDisk

`KubernetesCoreV1ApiPhotonPersistentDiskVolumeSource`

photonPersistentDisk represents a PhotonController persistent disk attached
and mounted on kubelets host machine. Deprecated: PhotonPersistentDisk is
deprecated and the in-tree photonPersistentDisk type is no longer
supported.

### spec.podTemplate.spec.volumes[].photonPersistentDisk.pdID

`string` · required

pdID is the ID that identifies Photon Controller persistent disk.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.volumes[].photonPersistentDisk.fsType

`string`

fsType is the filesystem type to mount. Must be a filesystem type supported
by the host operating system. Ex. "ext4", "xfs", "ntfs". Implicitly
inferred to be "ext4" if unspecified.

### spec.podTemplate.spec.volumes[].projected

`KubernetesCoreV1ApiProjectedVolumeSource`

projected items for all in one resources secrets, configmaps, and downward
API.

### spec.podTemplate.spec.volumes[].projected.sources

`[]KubernetesCoreV1ApiVolumeProjection`

sources is the list of volume projections. Each entry in this list handles
one source.

- rule: Set exactly one of clusterTrustBundle, configMap, downwardAPI, podCertificate, secret, serviceAccountToken

### spec.podTemplate.spec.volumes[].projected.sources[].secret

`KubernetesCoreV1ApiSecretProjection`

secret information about the secret data to project.

### spec.podTemplate.spec.volumes[].projected.sources[].secret.name

`string | valueFrom` · required

`name` of the referent. This field is effectively required, but due to
backwards compatibility is allowed to be empty. Instances of this type with
an empty value here are almost certainly wrong. More info:
https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names

Defaults to a KubernetesSecret reference (its secret_name output); pass the
literal name with `value:` for one Planton does not manage.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.podTemplate.spec.volumes[].projected.sources[].secret.items

`[]KubernetesCoreV1ApiKeyToPath`

items if unspecified, each key-value pair in the Data field of the
referenced Secret will be projected into the volume as a file whose name is
the key and content is the value. If specified, the listed keys will be
projected into the specified paths, and unlisted keys will not be present.
If a key is specified which is not present in the Secret, the volume setup
will error unless it is marked optional. Paths must be relative and may not
contain the '..' path or start with '..'.

### spec.podTemplate.spec.volumes[].projected.sources[].secret.items[].key

`string` · required

key is the key to project.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.volumes[].projected.sources[].secret.items[].path

`string` · required

path is the relative path of the file to map the key to. May not be an
absolute path. May not contain the path element '..'. May not start with
the string '..'.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.volumes[].projected.sources[].secret.items[].mode

`int32` · optional (explicit presence)

mode is Optional: mode bits used to set permissions on this file. Must be
an octal value between 0000 and 0777 or a decimal value between 0 and 511.
YAML accepts both octal and decimal values, JSON requires decimal values
for mode bits. If not specified, the volume defaultMode will be used. This
might be in conflict with other options that affect the file mode, like
fsGroup, and the result can be other mode bits set.

### spec.podTemplate.spec.volumes[].projected.sources[].secret.items[].user

`int64` · optional (explicit presence)

user is Optional: The owner UID of the created file. If specified, the
item-level user field takes precedence over defaultUser. (Alpha) This field
requires the AtomicWriteVolumeUserFields feature gate to be enabled.

### spec.podTemplate.spec.volumes[].projected.sources[].secret.optional

`bool` · optional (explicit presence)

optional field specify whether the Secret or its key must be defined.

### spec.podTemplate.spec.volumes[].projected.sources[].downwardAPI

`KubernetesCoreV1ApiDownwardAPIProjection`

downwardAPI information about the downwardAPI data to project.

### spec.podTemplate.spec.volumes[].projected.sources[].downwardAPI.items

`[]KubernetesCoreV1ApiDownwardAPIVolumeFile`

`items` is a list of DownwardAPIVolume file.

### spec.podTemplate.spec.volumes[].projected.sources[].downwardAPI.items[].path

`string` · required

Required: Path is  the relative path name of the file to be created. Must
not be absolute or contain the '..' path. Must be utf-8 encoded. The first
item of the relative path must not start with '..'.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.volumes[].projected.sources[].downwardAPI.items[].fieldRef

`KubernetesCoreV1ApiObjectFieldSelector`

Required: Selects a field of the pod: only annotations, labels, name,
namespace and uid are supported.

### spec.podTemplate.spec.volumes[].projected.sources[].downwardAPI.items[].fieldRef.apiVersion

`string`

Version of the schema the FieldPath is written in terms of, defaults to
"v1".

### spec.podTemplate.spec.volumes[].projected.sources[].downwardAPI.items[].fieldRef.fieldPath

`string` · required

Path of the field to select in the specified API version.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.volumes[].projected.sources[].downwardAPI.items[].resourceFieldRef

`KubernetesCoreV1ApiResourceFieldSelector`

Selects a resource of the container: only resources limits and requests
(limits.cpu, limits.memory, requests.cpu and requests.memory) are currently
supported.

### spec.podTemplate.spec.volumes[].projected.sources[].downwardAPI.items[].resourceFieldRef.containerName

`string`

Container name: required for volumes, optional for env vars.

### spec.podTemplate.spec.volumes[].projected.sources[].downwardAPI.items[].resourceFieldRef.resource

`string` · required

Required: resource to select.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.volumes[].projected.sources[].downwardAPI.items[].resourceFieldRef.divisor

`string` · optional (explicit presence)

Specifies the output format of the exposed resources, defaults to "1"

- rule: {"string":{"pattern":"^(\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))))?$"}}

### spec.podTemplate.spec.volumes[].projected.sources[].downwardAPI.items[].mode

`int32` · optional (explicit presence)

Optional: mode bits used to set permissions on this file, must be an octal
value between 0000 and 0777 or a decimal value between 0 and 511. YAML
accepts both octal and decimal values, JSON requires decimal values for
mode bits. If not specified, the volume defaultMode will be used. This
might be in conflict with other options that affect the file mode, like
fsGroup, and the result can be other mode bits set.

### spec.podTemplate.spec.volumes[].projected.sources[].downwardAPI.items[].user

`int64` · optional (explicit presence)

user is Optional: The owner UID of the created file. If specified, the
item-level user field takes precedence over defaultUser. (Alpha) This field
requires the AtomicWriteVolumeUserFields feature gate to be enabled.

### spec.podTemplate.spec.volumes[].projected.sources[].configMap

`KubernetesCoreV1ApiConfigMapProjection`

configMap information about the configMap data to project.

### spec.podTemplate.spec.volumes[].projected.sources[].configMap.name

`string | valueFrom` · required

`name` of the referent. This field is effectively required, but due to
backwards compatibility is allowed to be empty. Instances of this type with
an empty value here are almost certainly wrong. More info:
https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names

Defaults to a KubernetesConfigMap reference (its configmap_name output);
pass the literal name with `value:` for one Planton does not manage.

- references: KubernetesConfigMap (`status.outputs.configmap_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesConfigMap, name: <that resource's name>, fieldPath: status.outputs.configmap_name}} -- a bare string does not parse

### spec.podTemplate.spec.volumes[].projected.sources[].configMap.items

`[]KubernetesCoreV1ApiKeyToPath`

items if unspecified, each key-value pair in the Data field of the
referenced ConfigMap will be projected into the volume as a file whose name
is the key and content is the value. If specified, the listed keys will be
projected into the specified paths, and unlisted keys will not be present.
If a key is specified which is not present in the ConfigMap, the volume
setup will error unless it is marked optional. Paths must be relative and
may not contain the '..' path or start with '..'.

### spec.podTemplate.spec.volumes[].projected.sources[].configMap.items[].key

`string` · required

key is the key to project.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.volumes[].projected.sources[].configMap.items[].path

`string` · required

path is the relative path of the file to map the key to. May not be an
absolute path. May not contain the path element '..'. May not start with
the string '..'.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.volumes[].projected.sources[].configMap.items[].mode

`int32` · optional (explicit presence)

mode is Optional: mode bits used to set permissions on this file. Must be
an octal value between 0000 and 0777 or a decimal value between 0 and 511.
YAML accepts both octal and decimal values, JSON requires decimal values
for mode bits. If not specified, the volume defaultMode will be used. This
might be in conflict with other options that affect the file mode, like
fsGroup, and the result can be other mode bits set.

### spec.podTemplate.spec.volumes[].projected.sources[].configMap.items[].user

`int64` · optional (explicit presence)

user is Optional: The owner UID of the created file. If specified, the
item-level user field takes precedence over defaultUser. (Alpha) This field
requires the AtomicWriteVolumeUserFields feature gate to be enabled.

### spec.podTemplate.spec.volumes[].projected.sources[].configMap.optional

`bool` · optional (explicit presence)

optional specify whether the ConfigMap or its keys must be defined.

### spec.podTemplate.spec.volumes[].projected.sources[].serviceAccountToken

`KubernetesCoreV1ApiServiceAccountTokenProjection`

serviceAccountToken is information about the serviceAccountToken data to
project.

### spec.podTemplate.spec.volumes[].projected.sources[].serviceAccountToken.audience

`string`

audience is the intended audience of the token. A recipient of a token must
identify itself with an identifier specified in the audience of the token,
and otherwise should reject the token. The audience defaults to the
identifier of the apiserver.

### spec.podTemplate.spec.volumes[].projected.sources[].serviceAccountToken.expirationSeconds

`int64` · optional (explicit presence)

expirationSeconds is the requested duration of validity of the service
account token. As the token approaches expiration, the kubelet volume
plugin will proactively rotate the service account token. The kubelet will
start trying to rotate the token if the token is older than 80 percent of
its time to live or if the token is older than 24 hours.Defaults to 1 hour
and must be at least 10 minutes.

### spec.podTemplate.spec.volumes[].projected.sources[].serviceAccountToken.path

`string` · required

path is the path relative to the mount point of the file to project the
token into.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.volumes[].projected.sources[].serviceAccountToken.user

`int64` · optional (explicit presence)

user is Optional: The owner UID of the created file. If specified, the
item-level user field takes precedence over defaultUser. (Alpha) This field
requires the AtomicWriteVolumeUserFields feature gate to be enabled.

### spec.podTemplate.spec.volumes[].projected.sources[].clusterTrustBundle

`KubernetesCoreV1ApiClusterTrustBundleProjection`

`clusterTrustBundle` allows a pod to access the `.spec.trustBundle` field
of ClusterTrustBundle objects in an auto-updating file.

Alpha, gated by the ClusterTrustBundleProjection feature gate.

ClusterTrustBundle objects can either be selected by name, or by the
combination of signer name and a label selector.

Kubelet performs aggressive normalization of the PEM contents written into
the pod filesystem.  Esoteric PEM features such as inter-block comments and
block headers are stripped.  Certificates are deduplicated. The ordering of
certificates within the file is arbitrary, and Kubelet may change the order
over time.

### spec.podTemplate.spec.volumes[].projected.sources[].clusterTrustBundle.name

`string` · optional (explicit presence)

Select a single ClusterTrustBundle by object name.  Mutually-exclusive with
signerName and labelSelector.

### spec.podTemplate.spec.volumes[].projected.sources[].clusterTrustBundle.signerName

`string` · optional (explicit presence)

Select all ClusterTrustBundles that match this signer name.
Mutually-exclusive with name.  The contents of all selected
ClusterTrustBundles will be unified and deduplicated.

### spec.podTemplate.spec.volumes[].projected.sources[].clusterTrustBundle.labelSelector

`KubernetesCoreV1ApiLabelSelector`

Select all ClusterTrustBundles that match this label selector.  Only has
effect if signerName is set.  Mutually-exclusive with name.  If unset,
interpreted as "match nothing".  If set but empty, interpreted as "match
everything".

### spec.podTemplate.spec.volumes[].projected.sources[].clusterTrustBundle.labelSelector.matchLabels

`map<string, string>`

matchLabels is a map of {key,value} pairs. A single {key,value} in the
matchLabels map is equivalent to an element of matchExpressions, whose key
field is "key", the operator is "In", and the values array contains only
"value". The requirements are ANDed.

### spec.podTemplate.spec.volumes[].projected.sources[].clusterTrustBundle.labelSelector.matchExpressions

`[]KubernetesCoreV1ApiLabelSelectorRequirement`

matchExpressions is a list of label selector requirements. The requirements
are ANDed.

- rule: In and NotIn need at least one value; Exists and DoesNotExist take none

### spec.podTemplate.spec.volumes[].projected.sources[].clusterTrustBundle.labelSelector.matchExpressions[].key

`string` · required

key is the label key that the selector applies to.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.volumes[].projected.sources[].clusterTrustBundle.labelSelector.matchExpressions[].operator

`string` · required

operator represents a key's relationship to a set of values. Valid
operators are In, NotIn, Exists and DoesNotExist.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.volumes[].projected.sources[].clusterTrustBundle.labelSelector.matchExpressions[].values

`[]string`

values is an array of string values. If the operator is In or NotIn, the
values array must be non-empty. If the operator is Exists or DoesNotExist,
the values array must be empty. This array is replaced during a strategic
merge patch.

### spec.podTemplate.spec.volumes[].projected.sources[].clusterTrustBundle.optional

`bool` · optional (explicit presence)

If true, don't block pod startup if the referenced ClusterTrustBundle(s)
aren't available.  If using name, then the named ClusterTrustBundle is
allowed not to exist.  If using signerName, then the combination of
signerName and labelSelector is allowed to match zero ClusterTrustBundles.

### spec.podTemplate.spec.volumes[].projected.sources[].clusterTrustBundle.path

`string` · required

Relative path from the volume root to write the bundle.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.volumes[].projected.sources[].clusterTrustBundle.user

`int64` · optional (explicit presence)

user is Optional: The owner UID of the created file. If specified, the
item-level user field takes precedence over defaultUser. (Alpha) This field
requires the AtomicWriteVolumeUserFields feature gate to be enabled.

### spec.podTemplate.spec.volumes[].projected.sources[].podCertificate

`KubernetesCoreV1ApiPodCertificateProjection`

Projects an auto-rotating credential bundle (private key and certificate
chain) that the pod can use either as a TLS client or server.

Kubelet generates a private key and uses it to send a PodCertificateRequest
to the named signer.  Once the signer approves the request and issues a
certificate chain, Kubelet writes the key and certificate chain to the pod
filesystem.  The pod does not start until certificates have been issued for
each podCertificate projected volume source in its spec.

Kubelet will begin trying to rotate the certificate at the time indicated
by the signer using the PodCertificateRequest.Status.BeginRefreshAt
timestamp.

Kubelet can write a single file, indicated by the credentialBundlePath
field, or separate files, indicated by the keyPath and certificateChainPath
fields.

The credential bundle is a single file in PEM format.  The first PEM entry
is the private key (in PKCS#8 format), and the remaining PEM entries are
the certificate chain issued by the signer (typically, signers will return
their certificate chain in leaf-to-root order).

Prefer using the credential bundle format, since your application code can
read it atomically.  If you use keyPath and certificateChainPath, your
application must make two separate file reads. If these coincide with a
certificate rotation, it is possible that the private key and leaf
certificate you read may not correspond to each other.  Your application
will need to check for this condition, and re-read until they are
consistent.

The named signer controls chooses the format of the certificate it issues;
consult the signer implementation's documentation to learn how to use the
certificates it issues.

Requires the PodCertificateProjection feature gate on the cluster; a
cluster without it drops this field from the pod.

### spec.podTemplate.spec.volumes[].projected.sources[].podCertificate.signerName

`string` · required

Kubelet's generated CSRs will be addressed to this signer.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.volumes[].projected.sources[].podCertificate.keyType

`string`

The type of keypair Kubelet will generate for the pod.

Valid values are "RSA3072", "RSA4096", "ECDSAP256", "ECDSAP384",
"ECDSAP521", and "ED25519".

- rule: {"string":{"in":["RSA3072","RSA4096","ECDSAP256","ECDSAP384","ECDSAP521","ED25519"]}}

### spec.podTemplate.spec.volumes[].projected.sources[].podCertificate.maxExpirationSeconds

`int32` · optional (explicit presence)

maxExpirationSeconds is the maximum lifetime permitted for the certificate.

Kubelet copies this value verbatim into the PodCertificateRequests it
generates for this projection.

If omitted, kube-apiserver will set it to 86400(24 hours). kube-apiserver
will reject values shorter than 3600 (1 hour).  The maximum allowable value
is 7862400 (91 days).

The signer implementation is then free to issue a certificate with any
lifetime *shorter* than MaxExpirationSeconds, but no shorter than 3600
seconds (1 hour).  This constraint is enforced by kube-apiserver.
`kubernetes.io` signers will never issue certificates with a lifetime
longer than 24 hours.

### spec.podTemplate.spec.volumes[].projected.sources[].podCertificate.credentialBundlePath

`string`

Write the credential bundle at this path in the projected volume.

The credential bundle is a single file that contains multiple PEM blocks.
The first PEM block is a PRIVATE KEY block, containing a PKCS#8 private
key.

The remaining blocks are CERTIFICATE blocks, containing the issued
certificate chain from the signer (leaf and any intermediates).

Using credentialBundlePath lets your Pod's application code make a single
atomic read that retrieves a consistent key and certificate chain.  If you
project them to separate files, your application code will need to
additionally check that the leaf certificate was issued to the key.

### spec.podTemplate.spec.volumes[].projected.sources[].podCertificate.keyPath

`string`

Write the key at this path in the projected volume.

Most applications should use credentialBundlePath.  When using keyPath and
certificateChainPath, your application needs to check that the key and leaf
certificate are consistent, because it is possible to read the files
mid-rotation.

### spec.podTemplate.spec.volumes[].projected.sources[].podCertificate.certificateChainPath

`string`

Write the certificate chain at this path in the projected volume.

Most applications should use credentialBundlePath.  When using keyPath and
certificateChainPath, your application needs to check that the key and leaf
certificate are consistent, because it is possible to read the files
mid-rotation.

### spec.podTemplate.spec.volumes[].projected.sources[].podCertificate.userAnnotations

`map<string, string>`

userAnnotations allow pod authors to pass additional information to the
signer implementation.  Kubernetes does not restrict or validate this
metadata in any way.

These values are copied verbatim into the `spec.unverifiedUserAnnotations`
field of the PodCertificateRequest objects that Kubelet creates.

Entries are subject to the same validation as object metadata annotations,
with the addition that all keys must be domain-prefixed. No restrictions
are placed on values, except an overall size limitation on the entire
field.

Signers should document the keys and values they support. Signers should
deny requests that contain keys they do not recognize.

### spec.podTemplate.spec.volumes[].projected.sources[].podCertificate.user

`int64` · optional (explicit presence)

user is Optional: The owner UID of the created file. If specified, the
item-level user field takes precedence over defaultUser. (Alpha) This field
requires the AtomicWriteVolumeUserFields feature gate to be enabled.

### spec.podTemplate.spec.volumes[].projected.defaultMode

`int32` · optional (explicit presence)

defaultMode are the mode bits used to set permissions on created files by
default. Must be an octal value between 0000 and 0777 or a decimal value
between 0 and 511. YAML accepts both octal and decimal values, JSON
requires decimal values for mode bits. Directories within the path are not
affected by this setting. This might be in conflict with other options that
affect the file mode, like fsGroup, and the result can be other mode bits
set.

### spec.podTemplate.spec.volumes[].projected.defaultUser

`int64` · optional (explicit presence)

defaultUser is Optional: The owner UID of the created files by default. The
defaultUser field is only used as a fallback when the item-level user field
is unset. (Alpha) This field requires the AtomicWriteVolumeUserFields
feature gate to be enabled.

### spec.podTemplate.spec.volumes[].portworxVolume

`KubernetesCoreV1ApiPortworxVolumeSource`

portworxVolume represents a portworx volume attached and mounted on
kubelets host machine. Deprecated: PortworxVolume is deprecated. All
operations for the in-tree portworxVolume type are redirected to the
pxd.portworx.com CSI driver.

### spec.podTemplate.spec.volumes[].portworxVolume.volumeID

`string` · required

volumeID uniquely identifies a Portworx volume.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.volumes[].portworxVolume.fsType

`string`

fSType represents the filesystem type to mount Must be a filesystem type
supported by the host operating system. Ex. "ext4", "xfs". Implicitly
inferred to be "ext4" if unspecified.

### spec.podTemplate.spec.volumes[].portworxVolume.readOnly

`bool`

readOnly defaults to false (read/write). ReadOnly here will force the
ReadOnly setting in VolumeMounts.

### spec.podTemplate.spec.volumes[].scaleIO

`KubernetesCoreV1ApiScaleIOVolumeSource`

scaleIO represents a ScaleIO persistent volume attached and mounted on
Kubernetes nodes. Deprecated: ScaleIO is deprecated and the in-tree scaleIO
type is no longer supported.

### spec.podTemplate.spec.volumes[].scaleIO.gateway

`string` · required

gateway is the host address of the ScaleIO API Gateway.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.volumes[].scaleIO.system

`string` · required

system is the name of the storage system as configured in ScaleIO.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.volumes[].scaleIO.secretRef

`KubernetesCoreV1ApiLocalObjectReference` · required

secretRef references to the secret for ScaleIO user and other sensitive
information. If this is not provided, Login operation will fail.

- rule: {"required":true}

### spec.podTemplate.spec.volumes[].scaleIO.secretRef.name

`string | valueFrom` · required

`name` of the referent. This field is effectively required, but due to
backwards compatibility is allowed to be empty. Instances of this type with
an empty value here are almost certainly wrong. More info:
https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names

Defaults to a KubernetesSecret reference (its secret_name output); pass the
literal name with `value:` for one Planton does not manage.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.podTemplate.spec.volumes[].scaleIO.sslEnabled

`bool`

sslEnabled Flag enable/disable SSL communication with Gateway, default
false.

### spec.podTemplate.spec.volumes[].scaleIO.protectionDomain

`string`

protectionDomain is the name of the ScaleIO Protection Domain for the
configured storage.

### spec.podTemplate.spec.volumes[].scaleIO.storagePool

`string`

storagePool is the ScaleIO Storage Pool associated with the protection
domain.

### spec.podTemplate.spec.volumes[].scaleIO.storageMode

`string`

storageMode indicates whether the storage for a volume should be
ThickProvisioned or ThinProvisioned. Default is ThinProvisioned.

Unset, the cluster writes "ThinProvisioned".

### spec.podTemplate.spec.volumes[].scaleIO.volumeName

`string`

volumeName is the name of a volume already created in the ScaleIO system
that is associated with this volume source.

### spec.podTemplate.spec.volumes[].scaleIO.fsType

`string`

fsType is the filesystem type to mount. Must be a filesystem type supported
by the host operating system. Ex. "ext4", "xfs", "ntfs". Default is "xfs".

Unset, the cluster writes "xfs".

### spec.podTemplate.spec.volumes[].scaleIO.readOnly

`bool`

readOnly Defaults to false (read/write). ReadOnly here will force the
ReadOnly setting in VolumeMounts.

### spec.podTemplate.spec.volumes[].storageos

`KubernetesCoreV1ApiStorageOSVolumeSource`

storageOS represents a StorageOS volume attached and mounted on Kubernetes
nodes. Deprecated: StorageOS is deprecated and the in-tree storageos type
is no longer supported.

### spec.podTemplate.spec.volumes[].storageos.volumeName

`string`

volumeName is the human-readable name of the StorageOS volume.  Volume
names are only unique within a namespace.

### spec.podTemplate.spec.volumes[].storageos.volumeNamespace

`string`

volumeNamespace specifies the scope of the volume within StorageOS.  If no
namespace is specified then the Pod's namespace will be used.  This allows
the Kubernetes name scoping to be mirrored within StorageOS for tighter
integration. Set VolumeName to any name to override the default behaviour.
Set to "default" if you are not using namespaces within StorageOS.
Namespaces that do not pre-exist within StorageOS will be created.

### spec.podTemplate.spec.volumes[].storageos.fsType

`string`

fsType is the filesystem type to mount. Must be a filesystem type supported
by the host operating system. Ex. "ext4", "xfs", "ntfs". Implicitly
inferred to be "ext4" if unspecified.

### spec.podTemplate.spec.volumes[].storageos.readOnly

`bool`

readOnly defaults to false (read/write). ReadOnly here will force the
ReadOnly setting in VolumeMounts.

### spec.podTemplate.spec.volumes[].storageos.secretRef

`KubernetesCoreV1ApiLocalObjectReference`

secretRef specifies the secret to use for obtaining the StorageOS API
credentials.  If not specified, default values will be attempted.

### spec.podTemplate.spec.volumes[].storageos.secretRef.name

`string | valueFrom` · required

`name` of the referent. This field is effectively required, but due to
backwards compatibility is allowed to be empty. Instances of this type with
an empty value here are almost certainly wrong. More info:
https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names

Defaults to a KubernetesSecret reference (its secret_name output); pass the
literal name with `value:` for one Planton does not manage.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.podTemplate.spec.volumes[].csi

`KubernetesCoreV1ApiCSIVolumeSource`

csi (Container Storage Interface) represents ephemeral storage that is
handled by certain external CSI drivers.

### spec.podTemplate.spec.volumes[].csi.driver

`string` · required

driver is the name of the CSI driver that handles this volume. Consult with
your admin for the correct name as registered in the cluster.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.volumes[].csi.readOnly

`bool` · optional (explicit presence)

readOnly specifies a read-only configuration for the volume. Defaults to
false (read/write).

### spec.podTemplate.spec.volumes[].csi.fsType

`string` · optional (explicit presence)

fsType to mount. Ex. "ext4", "xfs", "ntfs". If not provided, the empty
value is passed to the associated CSI driver which will determine the
default filesystem to apply.

### spec.podTemplate.spec.volumes[].csi.volumeAttributes

`map<string, string>`

volumeAttributes stores driver-specific properties that are passed to the
CSI driver. Consult your driver's documentation for supported values.

### spec.podTemplate.spec.volumes[].csi.nodePublishSecretRef

`KubernetesCoreV1ApiLocalObjectReference`

nodePublishSecretRef is a reference to the secret object containing
sensitive information to pass to the CSI driver to complete the CSI
NodePublishVolume and NodeUnpublishVolume calls. This field is optional,
and  may be empty if no secret is required. If the secret object contains
more than one secret, all secret references are passed.

### spec.podTemplate.spec.volumes[].csi.nodePublishSecretRef.name

`string | valueFrom` · required

`name` of the referent. This field is effectively required, but due to
backwards compatibility is allowed to be empty. Instances of this type with
an empty value here are almost certainly wrong. More info:
https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names

Defaults to a KubernetesSecret reference (its secret_name output); pass the
literal name with `value:` for one Planton does not manage.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.podTemplate.spec.volumes[].ephemeral

`KubernetesCoreV1ApiEphemeralVolumeSource`

ephemeral represents a volume that is handled by a cluster storage driver.
The volume's lifecycle is tied to the pod that defines it - it will be
created before the pod starts, and deleted when the pod is removed.

Use this if: a) the volume is only needed while the pod runs, b) features
of normal volumes like restoring from snapshot or capacity tracking are
needed, c) the storage driver is specified through a storage class, and d)
the storage driver supports dynamic volume provisioning through a
PersistentVolumeClaim (see EphemeralVolumeSource for more information on
the connection between this volume type and PersistentVolumeClaim).

Use PersistentVolumeClaim or one of the vendor-specific APIs for volumes
that persist for longer than the lifecycle of an individual pod.

Use CSI for light-weight local ephemeral volumes if the CSI driver is meant
to be used that way - see the documentation of the driver for more
information.

A pod can use both types of ephemeral volumes and persistent volumes at the
same time.

### spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate

`KubernetesCoreV1ApiPersistentVolumeClaimTemplate` · required

Will be used to create a stand-alone PVC to provision the volume. The pod
in which this EphemeralVolumeSource is embedded will be the owner of the
PVC, i.e. the PVC will be deleted together with the pod.  The name of the
PVC will be `<pod name>-<volume name>` where `<volume name>` is the name
from the `PodSpec.Volumes` array entry. Pod validation will reject the pod
if the concatenated name is not valid for a PVC (for example, too long).

An existing PVC with that name that is not owned by the pod will *not* be
used for the pod to avoid using an unrelated volume by mistake. Starting
the pod is then blocked until the unrelated PVC is removed. If such a
pre-created PVC is meant to be used by the pod, the PVC has to updated with
an owner reference to the pod once the pod exists. Normally this should not
be necessary, but it may be useful when manually reconstructing a broken
cluster.

This field is read-only and no changes will be made by Kubernetes to the
PVC after it has been created.

Required, must not be nil.

- rule: {"required":true}

### spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.metadata

`KubernetesCoreV1ApiObjectMeta`

Upstream copies labels and annotations from here onto the PVC, but the
definition agent-sandbox serves declares this object with no properties, so
the API server prunes anything set inside it. It is kept, empty, so the
shape stays upstream's; label the PVC through the claim template instead
when you need to.

### spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.spec

`KubernetesCoreV1ApiPersistentVolumeClaimSpec` · required

The specification for the PersistentVolumeClaim. The entire content is
copied unchanged into the PVC that gets created from this template. The
same fields as in a PersistentVolumeClaim are also valid here.

- rule: {"required":true}

### spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.spec.accessModes

`[]string`

accessModes contains the desired access modes the volume should have. More
info:
https://kubernetes.io/docs/concepts/storage/persistent-volumes#access-modes-1

- rule: {"repeated":{"items":{"string":{"in":["ReadWriteOnce","ReadOnlyMany","ReadWriteMany","ReadWriteOncePod"]}}}}

### spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.spec.selector

`KubernetesCoreV1ApiLabelSelector`

selector is a label query over volumes to consider for binding.

### spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.spec.selector.matchLabels

`map<string, string>`

matchLabels is a map of {key,value} pairs. A single {key,value} in the
matchLabels map is equivalent to an element of matchExpressions, whose key
field is "key", the operator is "In", and the values array contains only
"value". The requirements are ANDed.

### spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.spec.selector.matchExpressions

`[]KubernetesCoreV1ApiLabelSelectorRequirement`

matchExpressions is a list of label selector requirements. The requirements
are ANDed.

- rule: In and NotIn need at least one value; Exists and DoesNotExist take none

### spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.spec.selector.matchExpressions[].key

`string` · required

key is the label key that the selector applies to.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.spec.selector.matchExpressions[].operator

`string` · required

operator represents a key's relationship to a set of values. Valid
operators are In, NotIn, Exists and DoesNotExist.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.spec.selector.matchExpressions[].values

`[]string`

values is an array of string values. If the operator is In or NotIn, the
values array must be non-empty. If the operator is Exists or DoesNotExist,
the values array must be empty. This array is replaced during a strategic
merge patch.

### spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.spec.resources

`KubernetesCoreV1ApiVolumeResourceRequirements`

resources represents the minimum resources the volume should have. Users
are allowed to specify resource requirements that are lower than previous
value but must still be higher than capacity recorded in the status field
of the claim. More info:
https://kubernetes.io/docs/concepts/storage/persistent-volumes#resources

### spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.spec.resources.limits

`map<string, string>`

`limits` describes the maximum amount of compute resources allowed. More
info:
https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/

- rule: {"map":{"values":{"string":{"pattern":"^(\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))))?$"}}}}

### spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.spec.resources.requests

`map<string, string>`

`requests` describes the minimum amount of compute resources required. If
Requests is omitted for a container, it defaults to Limits if that is
explicitly specified, otherwise to an implementation-defined value.
Requests cannot exceed Limits. More info:
https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/

- rule: {"map":{"values":{"string":{"pattern":"^(\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))))?$"}}}}

### spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.spec.volumeName

`string`

volumeName is the binding reference to the PersistentVolume backing this
claim.

### spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.spec.storageClassName

`string | valueFrom`

storageClassName is the name of the StorageClass required by the claim.
More info:
https://kubernetes.io/docs/concepts/storage/persistent-volumes#class-1

Defaults to a KubernetesStorageClass reference (its storage_class_name
output); pass the literal name with `value:` for one Planton does not
manage.

- references: KubernetesStorageClass (`status.outputs.storage_class_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesStorageClass, name: <that resource's name>, fieldPath: status.outputs.storage_class_name}} -- a bare string does not parse

### spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.spec.volumeMode

`string` · optional (explicit presence)

volumeMode defines what type of volume is required by the claim. Value of
Filesystem is implied when not included in claim spec.

- rule: {"string":{"in":["Block","Filesystem"]}}

### spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.spec.dataSource

`KubernetesCoreV1ApiTypedLocalObjectReference`

dataSource field can be used to specify either: * An existing
VolumeSnapshot object (snapshot.storage.k8s.io/VolumeSnapshot) * An
existing PVC (PersistentVolumeClaim) If the provisioner or an external
controller can support the specified data source, it will create a new
volume based on the contents of the specified data source. dataSource
contents will be copied to dataSourceRef, and dataSourceRef contents will
be copied to dataSource when dataSourceRef.namespace is not specified. If
the namespace is specified, then dataSourceRef will not be copied to
dataSource.

### spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.spec.dataSource.apiGroup

`string` · optional (explicit presence)

`apiGroup` is the group for the resource being referenced. If APIGroup is
not specified, the specified Kind must be in the core API group. For any
other third-party types, APIGroup is required.

### spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.spec.dataSource.kind

`string` · required

`kind` is the type of resource being referenced.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.spec.dataSource.name

`string` · required

`name` is the name of resource being referenced.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.spec.dataSourceRef

`KubernetesCoreV1ApiTypedObjectReference`

dataSourceRef specifies the object from which to populate the volume with
data, if a non-empty volume is desired. This may be any object from a
non-empty API group (non core object) or a PersistentVolumeClaim object.
When this field is specified, volume binding will only succeed if the type
of the specified object matches some installed volume populator or dynamic
provisioner. This field will replace the functionality of the dataSource
field and as such if both fields are non-empty, they must have the same
value. For backwards compatibility, when namespace isn't specified in
dataSourceRef, both fields (dataSource and dataSourceRef) will be set to
the same value automatically if one of them is empty and the other is
non-empty. When namespace is specified in dataSourceRef, dataSource isn't
set to the same value and must be empty. There are three important
differences between dataSource and dataSourceRef: * While dataSource only
allows two specific types of objects, dataSourceRef allows any non-core
object, as well as PersistentVolumeClaim objects. * While dataSource
ignores disallowed values (dropping them), dataSourceRef preserves all
values, and generates an error if a disallowed value is specified. * While
dataSource only allows local objects, dataSourceRef allows objects in any
namespaces. (Alpha) Using the namespace field of dataSourceRef requires the
CrossNamespaceVolumeDataSource feature gate to be enabled.

### spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.spec.dataSourceRef.apiGroup

`string` · optional (explicit presence)

`apiGroup` is the group for the resource being referenced. If APIGroup is
not specified, the specified Kind must be in the core API group. For any
other third-party types, APIGroup is required.

### spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.spec.dataSourceRef.kind

`string` · required

`kind` is the type of resource being referenced.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.spec.dataSourceRef.name

`string` · required

`name` is the name of resource being referenced.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.spec.dataSourceRef.namespace

`string` · optional (explicit presence)

`namespace` is the namespace of resource being referenced Note that when a
namespace is specified, a gateway.networking.k8s.io/ReferenceGrant object
is required in the referent namespace to allow that namespace's owner to
accept the reference. See the ReferenceGrant documentation for details.
(Alpha) This field requires the CrossNamespaceVolumeDataSource feature gate
to be enabled.

### spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.spec.volumeAttributesClassName

`string` · optional (explicit presence)

volumeAttributesClassName may be used to set the VolumeAttributesClass used
by this claim. If specified, the CSI driver will create or update the
volume with the attributes defined in the corresponding
VolumeAttributesClass. This has a different purpose than storageClassName,
it can be changed after the claim is created. An empty string or nil value
indicates that no VolumeAttributesClass will be applied to the claim. If
the claim enters an Infeasible error state, this field can be reset to its
previous value (including nil) to cancel the modification. If the resource
referred to by volumeAttributesClass does not exist, this
PersistentVolumeClaim will be set to a Pending state, as reflected by the
modifyVolumeStatus field, until such as a resource exists. More info:
https://kubernetes.io/docs/concepts/storage/volume-attributes-classes/

### spec.podTemplate.spec.volumes[].image

`KubernetesCoreV1ApiImageVolumeSource`

image represents an OCI object (a container image or artifact) pulled and
mounted on the kubelet's host machine. The volume is resolved at pod
startup depending on which PullPolicy value is provided:

- Always: the kubelet always attempts to pull the reference. Container
creation will fail If the pull fails. - Never: the kubelet never pulls the
reference and only uses a local image or artifact. Container creation will
fail if the reference isn't present. - IfNotPresent: the kubelet pulls if
the reference isn't already present on disk. Container creation will fail
if the reference isn't present and the pull fails.

The volume gets re-resolved if the pod gets deleted and recreated, which
means that new remote content will become available on pod recreation. A
failure to resolve or pull the image during pod startup will block
containers from starting and may add significant latency. Failures will be
retried using normal volume backoff and will be reported on the pod reason
and message. The types of objects that may be mounted by this volume are
defined by the container runtime implementation on a host machine and at
minimum must include all valid types supported by the container image
field. The OCI object gets mounted in a single directory
(spec.containers[*].volumeMounts.mountPath) by merging the manifest layers
in the same way as for container images. The volume will be mounted
read-only (ro). Sub path mounts for containers are not supported
(spec.containers[*].volumeMounts.subpath) before 1.33. The field
spec.securityContext.fsGroupChangePolicy has no effect on this volume type.

### spec.podTemplate.spec.volumes[].image.reference

`string`

Required: Image or artifact reference to be used. Behaves in the same way
as pod.spec.containers[*].image. Pull secrets will be assembled in the same
way as for the container image by looking up node credentials, SA image
pull secrets, and pod spec image pull secrets. More info:
https://kubernetes.io/docs/concepts/containers/images This field is
optional to allow higher level config management to default or override
container images in workload controllers like Deployments and StatefulSets.

### spec.podTemplate.spec.volumes[].image.pullPolicy

`string` · optional (explicit presence)

Policy for pulling OCI objects. Possible values are: Always: the kubelet
always attempts to pull the reference. Container creation will fail If the
pull fails. Never: the kubelet never pulls the reference and only uses a
local image or artifact. Container creation will fail if the reference
isn't present. IfNotPresent: the kubelet pulls if the reference isn't
already present on disk. Container creation will fail if the reference
isn't present and the pull fails. Defaults to Always if :latest tag is
specified, or IfNotPresent otherwise.

- rule: {"string":{"in":["Always","Never","IfNotPresent"]}}

### spec.podTemplate.spec.initContainers

`[]KubernetesCoreV1ApiContainer`

List of initialization containers belonging to the pod. Init containers are
executed in order prior to containers being started. If any init container
fails, the pod is considered to have failed and is handled according to its
restartPolicy. The name for an init container or normal container must be
unique among all containers. Init containers may not have Lifecycle
actions, Readiness probes, Liveness probes, or Startup probes. The
resourceRequirements of an init container are taken into account during
scheduling by finding the highest request/limit for each resource type, and
then using the max of that value or the sum of the normal containers.
Limits are applied to init containers in a similar fashion. Init containers
cannot currently be added or removed. Cannot be updated. More info:
https://kubernetes.io/docs/concepts/workloads/pods/init-containers/

### spec.podTemplate.spec.initContainers[].name

`string` · required

`name` of the container specified as a DNS_LABEL. Each container in a pod
must have a unique name (DNS_LABEL). Cannot be updated.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.initContainers[].image

`string`

Container image name. More info:
https://kubernetes.io/docs/concepts/containers/images This field is
optional to allow higher level config management to default or override
container images in workload controllers like Deployments and StatefulSets.

### spec.podTemplate.spec.initContainers[].command

`[]string`

Entrypoint array. Not executed within a shell. The container image's
ENTRYPOINT is used if this is not provided. Variable references $(VAR_NAME)
are expanded using the container's environment. If a variable cannot be
resolved, the reference in the input string will be unchanged. Double $$
are reduced to a single $, which allows for escaping the $(VAR_NAME)
syntax: i.e. "$$(VAR_NAME)" will produce the string literal "$(VAR_NAME)".
Escaped references will never be expanded, regardless of whether the
variable exists or not. Cannot be updated. More info:
https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell

### spec.podTemplate.spec.initContainers[].args

`[]string`

Arguments to the entrypoint. The container image's CMD is used if this is
not provided. Variable references $(VAR_NAME) are expanded using the
container's environment. If a variable cannot be resolved, the reference in
the input string will be unchanged. Double $$ are reduced to a single $,
which allows for escaping the $(VAR_NAME) syntax: i.e. "$$(VAR_NAME)" will
produce the string literal "$(VAR_NAME)". Escaped references will never be
expanded, regardless of whether the variable exists or not. Cannot be
updated. More info:
https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell

### spec.podTemplate.spec.initContainers[].workingDir

`string`

Container's working directory. If not specified, the container runtime's
default will be used, which might be configured in the container image.
Cannot be updated.

### spec.podTemplate.spec.initContainers[].ports

`[]KubernetesCoreV1ApiContainerPort`

List of ports to expose from the container. Not specifying a port here DOES
NOT prevent that port from being exposed. Any port which is listening on
the default "0.0.0.0" address inside a container will be accessible from
the network. Modifying this array with strategic merge patch may corrupt
the data. For more information See
https://github.com/kubernetes/kubernetes/issues/108255. Cannot be updated.

### spec.podTemplate.spec.initContainers[].ports[].name

`string`

If specified, this must be an IANA_SVC_NAME and unique within the pod. Each
named port in a pod must have a unique name. Name for the port that can be
referred to by services.

### spec.podTemplate.spec.initContainers[].ports[].hostPort

`int32`

Number of port to expose on the host. If specified, this must be a valid
port number, 0 < x < 65536. If HostNetwork is specified, this must match
ContainerPort. Most containers do not need this.

### spec.podTemplate.spec.initContainers[].ports[].containerPort

`int32` · required · optional (explicit presence)

Number of port to expose on the pod's IP address. This must be a valid port
number, 0 < x < 65536.

- rule: {"required":true}

### spec.podTemplate.spec.initContainers[].ports[].protocol

`string` · optional (explicit presence)

`protocol` for port. Must be UDP, TCP, or SCTP. Defaults to "TCP".

Unset, the cluster writes "TCP".

- rule: {"string":{"in":["TCP","UDP","SCTP"]}}

### spec.podTemplate.spec.initContainers[].ports[].hostIP

`string`

What host IP to bind the external port to.

### spec.podTemplate.spec.initContainers[].envFrom

`[]KubernetesCoreV1ApiEnvFromSource`

List of sources to populate environment variables in the container. The
keys defined within a source may consist of any printable ASCII characters
except '='. When a key exists in multiple sources, the value associated
with the last source will take precedence. Values defined by an Env with a
duplicate key will take precedence. Cannot be updated.

- rule: Set exactly one of configMapRef, secretRef

### spec.podTemplate.spec.initContainers[].envFrom[].prefix

`string`

Optional text to prepend to the name of each environment variable. May
consist of any printable ASCII characters except '='.

### spec.podTemplate.spec.initContainers[].envFrom[].configMapRef

`KubernetesCoreV1ApiConfigMapEnvSource`

The ConfigMap to select from.

### spec.podTemplate.spec.initContainers[].envFrom[].configMapRef.name

`string | valueFrom` · required

`name` of the referent. This field is effectively required, but due to
backwards compatibility is allowed to be empty. Instances of this type with
an empty value here are almost certainly wrong. More info:
https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names

Defaults to a KubernetesConfigMap reference (its configmap_name output);
pass the literal name with `value:` for one Planton does not manage.

- references: KubernetesConfigMap (`status.outputs.configmap_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesConfigMap, name: <that resource's name>, fieldPath: status.outputs.configmap_name}} -- a bare string does not parse

### spec.podTemplate.spec.initContainers[].envFrom[].configMapRef.optional

`bool` · optional (explicit presence)

Specify whether the ConfigMap must be defined.

### spec.podTemplate.spec.initContainers[].envFrom[].secretRef

`KubernetesCoreV1ApiSecretEnvSource`

The Secret to select from.

### spec.podTemplate.spec.initContainers[].envFrom[].secretRef.name

`string | valueFrom` · required

`name` of the referent. This field is effectively required, but due to
backwards compatibility is allowed to be empty. Instances of this type with
an empty value here are almost certainly wrong. More info:
https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names

Defaults to a KubernetesSecret reference (its secret_name output); pass the
literal name with `value:` for one Planton does not manage.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.podTemplate.spec.initContainers[].envFrom[].secretRef.optional

`bool` · optional (explicit presence)

Specify whether the Secret must be defined.

### spec.podTemplate.spec.initContainers[].env

`[]KubernetesCoreV1ApiEnvVar`

List of environment variables to set in the container. Cannot be updated.

### spec.podTemplate.spec.initContainers[].env[].name

`string` · required

`name` of the environment variable. May consist of any printable ASCII
characters except '='.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.initContainers[].env[].value

`string`

Variable references $(VAR_NAME) are expanded using the previously defined
environment variables in the container and any service environment
variables. If a variable cannot be resolved, the reference in the input
string will be unchanged. Double $$ are reduced to a single $, which allows
for escaping the $(VAR_NAME) syntax: i.e. "$$(VAR_NAME)" will produce the
string literal "$(VAR_NAME)". Escaped references will never be expanded,
regardless of whether the variable exists or not. Defaults to "".

### spec.podTemplate.spec.initContainers[].env[].valueFrom

`KubernetesCoreV1ApiEnvVarSource`

Source for the environment variable's value. Cannot be used if value is not
empty.

- rule: Set exactly one of configMapKeyRef, fieldRef, fileKeyRef, resourceFieldRef, secretKeyRef

### spec.podTemplate.spec.initContainers[].env[].valueFrom.fieldRef

`KubernetesCoreV1ApiObjectFieldSelector`

Selects a field of the pod: supports metadata.name, metadata.namespace,
`metadata.labels['<KEY>']`, `metadata.annotations['<KEY>']`, spec.nodeName,
spec.serviceAccountName, status.hostIP, status.podIP, status.podIPs.

### spec.podTemplate.spec.initContainers[].env[].valueFrom.fieldRef.apiVersion

`string`

Version of the schema the FieldPath is written in terms of, defaults to
"v1".

### spec.podTemplate.spec.initContainers[].env[].valueFrom.fieldRef.fieldPath

`string` · required

Path of the field to select in the specified API version.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.initContainers[].env[].valueFrom.resourceFieldRef

`KubernetesCoreV1ApiResourceFieldSelector`

Selects a resource of the container: only resources limits and requests
(limits.cpu, limits.memory, limits.ephemeral-storage, requests.cpu,
requests.memory and requests.ephemeral-storage) are currently supported.

### spec.podTemplate.spec.initContainers[].env[].valueFrom.resourceFieldRef.containerName

`string`

Container name: required for volumes, optional for env vars.

### spec.podTemplate.spec.initContainers[].env[].valueFrom.resourceFieldRef.resource

`string` · required

Required: resource to select.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.initContainers[].env[].valueFrom.resourceFieldRef.divisor

`string` · optional (explicit presence)

Specifies the output format of the exposed resources, defaults to "1"

- rule: {"string":{"pattern":"^(\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))))?$"}}

### spec.podTemplate.spec.initContainers[].env[].valueFrom.configMapKeyRef

`KubernetesCoreV1ApiConfigMapKeySelector`

Selects a key of a ConfigMap.

### spec.podTemplate.spec.initContainers[].env[].valueFrom.configMapKeyRef.name

`string | valueFrom` · required

`name` of the referent. This field is effectively required, but due to
backwards compatibility is allowed to be empty. Instances of this type with
an empty value here are almost certainly wrong. More info:
https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names

Defaults to a KubernetesConfigMap reference (its configmap_name output);
pass the literal name with `value:` for one Planton does not manage.

- references: KubernetesConfigMap (`status.outputs.configmap_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesConfigMap, name: <that resource's name>, fieldPath: status.outputs.configmap_name}} -- a bare string does not parse

### spec.podTemplate.spec.initContainers[].env[].valueFrom.configMapKeyRef.key

`string` · required

The key to select from the ConfigMap's Data field. Keys in the BinaryData
field are not currently propagated to container env vars.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.initContainers[].env[].valueFrom.configMapKeyRef.optional

`bool` · optional (explicit presence)

Specify whether the ConfigMap or its key must be defined.

### spec.podTemplate.spec.initContainers[].env[].valueFrom.secretKeyRef

`KubernetesCoreV1ApiSecretKeySelector`

Selects a key of a secret in the pod's namespace.

### spec.podTemplate.spec.initContainers[].env[].valueFrom.secretKeyRef.name

`string | valueFrom` · required

`name` of the referent. This field is effectively required, but due to
backwards compatibility is allowed to be empty. Instances of this type with
an empty value here are almost certainly wrong. More info:
https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names

Defaults to a KubernetesSecret reference (its secret_name output); pass the
literal name with `value:` for one Planton does not manage.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.podTemplate.spec.initContainers[].env[].valueFrom.secretKeyRef.key

`string` · required

The key of the secret to select from.  Must be a valid secret key.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.initContainers[].env[].valueFrom.secretKeyRef.optional

`bool` · optional (explicit presence)

Specify whether the Secret or its key must be defined.

### spec.podTemplate.spec.initContainers[].env[].valueFrom.fileKeyRef

`KubernetesCoreV1ApiFileKeySelector`

`fileKeyRef` selects a key of the env file. Requires the EnvFiles feature
gate to be enabled.

### spec.podTemplate.spec.initContainers[].env[].valueFrom.fileKeyRef.volumeName

`string` · required

The name of the volume mount containing the env file.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.initContainers[].env[].valueFrom.fileKeyRef.path

`string` · required

The path within the volume from which to select the file. Must be relative
and may not contain the '..' path or start with '..'.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.initContainers[].env[].valueFrom.fileKeyRef.key

`string` · required

The key within the env file. An invalid key will prevent the pod from
starting. The keys defined within a source may consist of any printable
ASCII characters except '='. During Alpha stage of the EnvFiles feature
gate, the key size is limited to 128 characters.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.initContainers[].env[].valueFrom.fileKeyRef.optional

`bool` · optional (explicit presence)

Specify whether the file or its key must be defined. If the file or key
does not exist, then the env var is not published. If optional is set to
true and the specified key does not exist, the environment variable will
not be set in the Pod's containers.

If optional is set to false and the specified key does not exist, an error
will be returned during Pod creation.

Unset, the cluster writes false.

### spec.podTemplate.spec.initContainers[].resources

`KubernetesCoreV1ApiResourceRequirements`

Compute Resources required by this container. Cannot be updated. More info:
https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/

### spec.podTemplate.spec.initContainers[].resources.limits

`map<string, string>`

`limits` describes the maximum amount of compute resources allowed. More
info:
https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/

- rule: {"map":{"values":{"string":{"pattern":"^(\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))))?$"}}}}

### spec.podTemplate.spec.initContainers[].resources.requests

`map<string, string>`

`requests` describes the minimum amount of compute resources required. If
Requests is omitted for a container, it defaults to Limits if that is
explicitly specified, otherwise to an implementation-defined value.
Requests cannot exceed Limits. More info:
https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/

- rule: {"map":{"values":{"string":{"pattern":"^(\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))))?$"}}}}

### spec.podTemplate.spec.initContainers[].resources.claims

`[]KubernetesCoreV1ApiResourceClaim`

`claims` lists the names of resources, defined in spec.resourceClaims, that
are used by this container.

This field depends on the DynamicResourceAllocation feature gate.

This field is immutable. It can only be set for containers.

### spec.podTemplate.spec.initContainers[].resources.claims[].name

`string` · required

`name` must match the name of one entry in pod.spec.resourceClaims of the
Pod where this field is used. It makes that resource available inside a
container.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.initContainers[].resources.claims[].request

`string`

`request` is the name chosen for a request in the referenced claim. If
empty, everything from the claim is made available, otherwise only the
result of this request.

### spec.podTemplate.spec.initContainers[].resizePolicy

`[]KubernetesCoreV1ApiContainerResizePolicy`

Resources resize policy for the container. This field cannot be set on
ephemeral containers.

Requires the InPlacePodVerticalScaling feature gate on the cluster; a
cluster without it drops this field from the pod.

### spec.podTemplate.spec.initContainers[].resizePolicy[].resourceName

`string` · required

Name of the resource to which this resource resize policy applies.
Supported values: cpu, memory.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.initContainers[].resizePolicy[].restartPolicy

`string`

Restart policy to apply when specified resource is resized. If not
specified, it defaults to NotRequired.

- rule: {"string":{"in":["NotRequired","RestartContainer"]}}

### spec.podTemplate.spec.initContainers[].restartPolicy

`string` · optional (explicit presence)

`restartPolicy` defines the restart behavior of individual containers in a
pod. This overrides the pod-level restart policy. When this field is not
specified, the restart behavior is defined by the Pod's restart policy and
the container type. Additionally, setting the RestartPolicy as "Always" for
the init container will have the following effect: this init container will
be continually restarted on exit until all regular containers have
terminated. Once all regular containers have completed, all init containers
with restartPolicy "Always" will be shut down. This lifecycle differs from
normal init containers and is often referred to as a "sidecar" container.
Although this init container still starts in the init container sequence,
it does not wait for the container to complete before proceeding to the
next init container. Instead, the next init container starts immediately
after this init container is started, or after any startupProbe has
successfully completed.

- rule: {"string":{"in":["Always","Never","OnFailure"]}}

### spec.podTemplate.spec.initContainers[].restartPolicyRules

`[]KubernetesCoreV1ApiContainerRestartRule`

Represents a list of rules to be checked to determine if the container
should be restarted on exit. The rules are evaluated in order. Once a rule
matches a container exit condition, the remaining rules are ignored. If no
rule matches the container exit condition, the Container-level restart
policy determines the whether the container is restarted or not.
Constraints on the rules: - At most 20 rules are allowed. - Rules can have
the same action. - Identical rules are not forbidden in validations. When
rules are specified, container MUST set RestartPolicy explicitly even it if
matches the Pod's RestartPolicy.

Requires the ContainerRestartRules feature gate on the cluster; a cluster
without it drops this field from the pod.

### spec.podTemplate.spec.initContainers[].restartPolicyRules[].action

`string`

Specifies the action taken on a container exit if the requirements are
satisfied. The only possible value is "Restart" to restart the container.

- rule: {"string":{"in":["Restart","RestartAllContainers"]}}

### spec.podTemplate.spec.initContainers[].restartPolicyRules[].exitCodes

`KubernetesCoreV1ApiContainerRestartRuleOnExitCodes`

Represents the exit codes to check on container exits.

### spec.podTemplate.spec.initContainers[].restartPolicyRules[].exitCodes.operator

`string`

Represents the relationship between the container exit code(s) and the
specified values. Possible values are: - In: the requirement is satisfied
if the container exit code is in the set of specified values. - NotIn: the
requirement is satisfied if the container exit code is not in the set of
specified values.

- rule: {"string":{"in":["In","NotIn"]}}

### spec.podTemplate.spec.initContainers[].restartPolicyRules[].exitCodes.values

`[]int32`

Specifies the set of values to check for container exit codes. At most 255
elements are allowed.

### spec.podTemplate.spec.initContainers[].volumeMounts

`[]KubernetesCoreV1ApiVolumeMount`

Pod volumes to mount into the container's filesystem. Cannot be updated.

### spec.podTemplate.spec.initContainers[].volumeMounts[].name

`string` · required

This must match the Name of a Volume.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.initContainers[].volumeMounts[].readOnly

`bool`

Mounted read-only if true, read-write otherwise (false or unspecified).
Defaults to false.

### spec.podTemplate.spec.initContainers[].volumeMounts[].recursiveReadOnly

`string` · optional (explicit presence)

`recursiveReadOnly` specifies whether read-only mounts should be handled
recursively.

If ReadOnly is false, this field has no meaning and must be unspecified.

If ReadOnly is true, and this field is set to Disabled, the mount is not
made recursively read-only.  If this field is set to IfPossible, the mount
is made recursively read-only, if it is supported by the container runtime.
If this field is set to Enabled, the mount is made recursively read-only if
it is supported by the container runtime, otherwise the pod will not be
started and an error will be generated to indicate the reason.

If this field is set to IfPossible or Enabled, MountPropagation must be set
to None (or be unspecified, which defaults to None).

If this field is not specified, it is treated as an equivalent of Disabled.

- rule: {"string":{"in":["Disabled","IfPossible","Enabled"]}}

### spec.podTemplate.spec.initContainers[].volumeMounts[].mountPath

`string` · required

Path within the container at which the volume should be mounted.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.initContainers[].volumeMounts[].subPath

`string`

Path within the volume from which the container's volume should be mounted.
Defaults to "" (volume's root).

### spec.podTemplate.spec.initContainers[].volumeMounts[].mountPropagation

`string` · optional (explicit presence)

mountPropagation determines how mounts are propagated from the host to
container and the other way around. When not set, MountPropagationNone is
used. This field is beta in 1.10. When RecursiveReadOnly is set to
IfPossible or to Enabled, MountPropagation must be None or unspecified
(which defaults to None).

- rule: {"string":{"in":["None","HostToContainer","Bidirectional"]}}

### spec.podTemplate.spec.initContainers[].volumeMounts[].subPathExpr

`string`

Expanded path within the volume from which the container's volume should be
mounted. Behaves similarly to SubPath but environment variable references
$(VAR_NAME) are expanded using the container's environment. Defaults to ""
(volume's root). SubPathExpr and SubPath are mutually exclusive.

### spec.podTemplate.spec.initContainers[].volumeMounts[].bindMountOptions

`[]string`

bindMountOptions is the list of additional bind mount options to apply when
mounting this volume into the container. Allowed values are noexec, nodev,
and nosuid. These are Linux mount options and have no effect on Windows
nodes. This field is not supported with image volumes. This is an alpha
field and requires enabling the VolumeBindMountOptions feature gate.

### spec.podTemplate.spec.initContainers[].volumeDevices

`[]KubernetesCoreV1ApiVolumeDevice`

volumeDevices is the list of block devices to be used by the container.

### spec.podTemplate.spec.initContainers[].volumeDevices[].name

`string` · required

name must match the name of a persistentVolumeClaim in the pod.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.initContainers[].volumeDevices[].devicePath

`string` · required

devicePath is the path inside of the container that the device will be
mapped to.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.initContainers[].livenessProbe

`KubernetesCoreV1ApiProbe`

Periodic probe of container liveness. Container will be restarted if the
probe fails. Cannot be updated. More info:
https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes

- rule: Set at most one of exec, httpGet, tcpSocket, grpc

### spec.podTemplate.spec.initContainers[].livenessProbe.exec

`KubernetesCoreV1ApiExecAction`

`exec` specifies a command to execute in the container.

### spec.podTemplate.spec.initContainers[].livenessProbe.exec.command

`[]string`

`command` is the command line to execute inside the container, the working
directory for the command  is root ('/') in the container's filesystem. The
command is simply exec'd, it is not run inside a shell, so traditional
shell instructions ('|', etc) won't work. To use a shell, you need to
explicitly call out to that shell. Exit status of 0 is treated as
live/healthy and non-zero is unhealthy.

### spec.podTemplate.spec.initContainers[].livenessProbe.httpGet

`KubernetesCoreV1ApiHTTPGetAction`

`httpGet` specifies an HTTP GET request to perform.

### spec.podTemplate.spec.initContainers[].livenessProbe.httpGet.path

`string`

`path` to access on the HTTP server.

### spec.podTemplate.spec.initContainers[].livenessProbe.httpGet.port

`string` · required

Name or number of the port to access on the container. Number must be in
the range 1 to 65535. Name must be an IANA_SVC_NAME.

- rule: port is a port number from 1 to 65535, or a port name (up to 15 lowercase letters, digits and hyphens, with at least one letter)
- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.initContainers[].livenessProbe.httpGet.host

`string`

`host` name to connect to, defaults to the pod IP. You probably want to set
"Host" in httpHeaders instead.

### spec.podTemplate.spec.initContainers[].livenessProbe.httpGet.scheme

`string` · optional (explicit presence)

`scheme` to use for connecting to the host. Defaults to HTTP.

- rule: {"string":{"in":["HTTP","HTTPS"]}}

### spec.podTemplate.spec.initContainers[].livenessProbe.httpGet.httpHeaders

`[]KubernetesCoreV1ApiHTTPHeader`

Custom headers to set in the request. HTTP allows repeated headers.

### spec.podTemplate.spec.initContainers[].livenessProbe.httpGet.httpHeaders[].name

`string` · required

The header field name. This will be canonicalized upon output, so
case-variant names will be understood as the same header.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.initContainers[].livenessProbe.httpGet.httpHeaders[].value

`string` · required

The header field value.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.initContainers[].livenessProbe.httpGet.protocol

`string` · optional (explicit presence)

`protocol` selects the wire protocol for the probe connection. Nil defaults
to HTTP/1.1.

Requires the H2CContainerProbe feature gate on the cluster; a cluster
without it drops this field from the pod.

- rule: {"string":{"in":["HTTP1","HTTP2"]}}

### spec.podTemplate.spec.initContainers[].livenessProbe.tcpSocket

`KubernetesCoreV1ApiTCPSocketAction`

`tcpSocket` specifies a connection to a TCP port.

### spec.podTemplate.spec.initContainers[].livenessProbe.tcpSocket.port

`string` · required

Number or name of the port to access on the container. Number must be in
the range 1 to 65535. Name must be an IANA_SVC_NAME.

- rule: port is a port number from 1 to 65535, or a port name (up to 15 lowercase letters, digits and hyphens, with at least one letter)
- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.initContainers[].livenessProbe.tcpSocket.host

`string`

Optional: Host name to connect to, defaults to the pod IP.

### spec.podTemplate.spec.initContainers[].livenessProbe.grpc

`KubernetesCoreV1ApiGRPCAction`

`grpc` specifies a GRPC HealthCheckRequest.

### spec.podTemplate.spec.initContainers[].livenessProbe.grpc.port

`int32` · required · optional (explicit presence)

`port` number of the gRPC service. Number must be in the range 1 to 65535.

- rule: {"required":true}

### spec.podTemplate.spec.initContainers[].livenessProbe.grpc.service

`string` · optional (explicit presence)

`service` is the name of the service to place in the gRPC
HealthCheckRequest (see
https://github.com/grpc/grpc/blob/master/doc/health-checking.md).

If this is not specified, the default behavior is defined by gRPC.

### spec.podTemplate.spec.initContainers[].livenessProbe.grpc.mode

`string` · optional (explicit presence)

mode specifies the connection mode for the gRPC health probe. Set to "TLS"
to use TLS without certificate verification. Set to "Plaintext" to use a
plaintext (insecure) connection explicitly. If not specified, the probe
uses a plaintext (insecure) connection.

Requires the GRPCContainerProbeTLS feature gate on the cluster; a cluster
without it drops this field from the pod.

- rule: {"string":{"in":["Plaintext","TLS"]}}

### spec.podTemplate.spec.initContainers[].livenessProbe.initialDelaySeconds

`int32`

Number of seconds after the container has started before liveness probes
are initiated. More info:
https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes

### spec.podTemplate.spec.initContainers[].livenessProbe.timeoutSeconds

`int32`

Number of seconds after which the probe times out. Defaults to 1 second.
Minimum value is 1. More info:
https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes

### spec.podTemplate.spec.initContainers[].livenessProbe.periodSeconds

`int32`

How often (in seconds) to perform the probe. Default to 10 seconds. Minimum
value is 1.

### spec.podTemplate.spec.initContainers[].livenessProbe.successThreshold

`int32`

Minimum consecutive successes for the probe to be considered successful
after having failed. Defaults to 1. Must be 1 for liveness and startup.
Minimum value is 1.

### spec.podTemplate.spec.initContainers[].livenessProbe.failureThreshold

`int32`

Minimum consecutive failures for the probe to be considered failed after
having succeeded. Defaults to 3. Minimum value is 1.

### spec.podTemplate.spec.initContainers[].livenessProbe.terminationGracePeriodSeconds

`int64` · optional (explicit presence)

Optional duration in seconds the pod needs to terminate gracefully upon
probe failure. The grace period is the duration in seconds after the
processes running in the pod are sent a termination signal and the time
when the processes are forcibly halted with a kill signal. Set this value
longer than the expected cleanup time for your process. If this value is
nil, the pod's terminationGracePeriodSeconds will be used. Otherwise, this
value overrides the value provided by the pod spec. Value must be
non-negative integer. The value zero indicates stop immediately via the
kill signal (no opportunity to shut down). This is a beta field and
requires enabling ProbeTerminationGracePeriod feature gate. Minimum value
is 1. spec.terminationGracePeriodSeconds is used if unset.

### spec.podTemplate.spec.initContainers[].readinessProbe

`KubernetesCoreV1ApiProbe`

Periodic probe of container service readiness. Container will be removed
from service endpoints if the probe fails. Cannot be updated. More info:
https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes

- rule: Set at most one of exec, httpGet, tcpSocket, grpc

### spec.podTemplate.spec.initContainers[].readinessProbe.exec

`KubernetesCoreV1ApiExecAction`

`exec` specifies a command to execute in the container.

### spec.podTemplate.spec.initContainers[].readinessProbe.exec.command

`[]string`

`command` is the command line to execute inside the container, the working
directory for the command  is root ('/') in the container's filesystem. The
command is simply exec'd, it is not run inside a shell, so traditional
shell instructions ('|', etc) won't work. To use a shell, you need to
explicitly call out to that shell. Exit status of 0 is treated as
live/healthy and non-zero is unhealthy.

### spec.podTemplate.spec.initContainers[].readinessProbe.httpGet

`KubernetesCoreV1ApiHTTPGetAction`

`httpGet` specifies an HTTP GET request to perform.

### spec.podTemplate.spec.initContainers[].readinessProbe.httpGet.path

`string`

`path` to access on the HTTP server.

### spec.podTemplate.spec.initContainers[].readinessProbe.httpGet.port

`string` · required

Name or number of the port to access on the container. Number must be in
the range 1 to 65535. Name must be an IANA_SVC_NAME.

- rule: port is a port number from 1 to 65535, or a port name (up to 15 lowercase letters, digits and hyphens, with at least one letter)
- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.initContainers[].readinessProbe.httpGet.host

`string`

`host` name to connect to, defaults to the pod IP. You probably want to set
"Host" in httpHeaders instead.

### spec.podTemplate.spec.initContainers[].readinessProbe.httpGet.scheme

`string` · optional (explicit presence)

`scheme` to use for connecting to the host. Defaults to HTTP.

- rule: {"string":{"in":["HTTP","HTTPS"]}}

### spec.podTemplate.spec.initContainers[].readinessProbe.httpGet.httpHeaders

`[]KubernetesCoreV1ApiHTTPHeader`

Custom headers to set in the request. HTTP allows repeated headers.

### spec.podTemplate.spec.initContainers[].readinessProbe.httpGet.httpHeaders[].name

`string` · required

The header field name. This will be canonicalized upon output, so
case-variant names will be understood as the same header.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.initContainers[].readinessProbe.httpGet.httpHeaders[].value

`string` · required

The header field value.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.initContainers[].readinessProbe.httpGet.protocol

`string` · optional (explicit presence)

`protocol` selects the wire protocol for the probe connection. Nil defaults
to HTTP/1.1.

Requires the H2CContainerProbe feature gate on the cluster; a cluster
without it drops this field from the pod.

- rule: {"string":{"in":["HTTP1","HTTP2"]}}

### spec.podTemplate.spec.initContainers[].readinessProbe.tcpSocket

`KubernetesCoreV1ApiTCPSocketAction`

`tcpSocket` specifies a connection to a TCP port.

### spec.podTemplate.spec.initContainers[].readinessProbe.tcpSocket.port

`string` · required

Number or name of the port to access on the container. Number must be in
the range 1 to 65535. Name must be an IANA_SVC_NAME.

- rule: port is a port number from 1 to 65535, or a port name (up to 15 lowercase letters, digits and hyphens, with at least one letter)
- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.initContainers[].readinessProbe.tcpSocket.host

`string`

Optional: Host name to connect to, defaults to the pod IP.

### spec.podTemplate.spec.initContainers[].readinessProbe.grpc

`KubernetesCoreV1ApiGRPCAction`

`grpc` specifies a GRPC HealthCheckRequest.

### spec.podTemplate.spec.initContainers[].readinessProbe.grpc.port

`int32` · required · optional (explicit presence)

`port` number of the gRPC service. Number must be in the range 1 to 65535.

- rule: {"required":true}

### spec.podTemplate.spec.initContainers[].readinessProbe.grpc.service

`string` · optional (explicit presence)

`service` is the name of the service to place in the gRPC
HealthCheckRequest (see
https://github.com/grpc/grpc/blob/master/doc/health-checking.md).

If this is not specified, the default behavior is defined by gRPC.

### spec.podTemplate.spec.initContainers[].readinessProbe.grpc.mode

`string` · optional (explicit presence)

mode specifies the connection mode for the gRPC health probe. Set to "TLS"
to use TLS without certificate verification. Set to "Plaintext" to use a
plaintext (insecure) connection explicitly. If not specified, the probe
uses a plaintext (insecure) connection.

Requires the GRPCContainerProbeTLS feature gate on the cluster; a cluster
without it drops this field from the pod.

- rule: {"string":{"in":["Plaintext","TLS"]}}

### spec.podTemplate.spec.initContainers[].readinessProbe.initialDelaySeconds

`int32`

Number of seconds after the container has started before liveness probes
are initiated. More info:
https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes

### spec.podTemplate.spec.initContainers[].readinessProbe.timeoutSeconds

`int32`

Number of seconds after which the probe times out. Defaults to 1 second.
Minimum value is 1. More info:
https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes

### spec.podTemplate.spec.initContainers[].readinessProbe.periodSeconds

`int32`

How often (in seconds) to perform the probe. Default to 10 seconds. Minimum
value is 1.

### spec.podTemplate.spec.initContainers[].readinessProbe.successThreshold

`int32`

Minimum consecutive successes for the probe to be considered successful
after having failed. Defaults to 1. Must be 1 for liveness and startup.
Minimum value is 1.

### spec.podTemplate.spec.initContainers[].readinessProbe.failureThreshold

`int32`

Minimum consecutive failures for the probe to be considered failed after
having succeeded. Defaults to 3. Minimum value is 1.

### spec.podTemplate.spec.initContainers[].readinessProbe.terminationGracePeriodSeconds

`int64` · optional (explicit presence)

Optional duration in seconds the pod needs to terminate gracefully upon
probe failure. The grace period is the duration in seconds after the
processes running in the pod are sent a termination signal and the time
when the processes are forcibly halted with a kill signal. Set this value
longer than the expected cleanup time for your process. If this value is
nil, the pod's terminationGracePeriodSeconds will be used. Otherwise, this
value overrides the value provided by the pod spec. Value must be
non-negative integer. The value zero indicates stop immediately via the
kill signal (no opportunity to shut down). This is a beta field and
requires enabling ProbeTerminationGracePeriod feature gate. Minimum value
is 1. spec.terminationGracePeriodSeconds is used if unset.

### spec.podTemplate.spec.initContainers[].startupProbe

`KubernetesCoreV1ApiProbe`

`startupProbe` indicates that the Pod has successfully initialized. If
specified, no other probes are executed until this completes successfully.
If this probe fails, the Pod will be restarted, just as if the
livenessProbe failed. This can be used to provide different probe
parameters at the beginning of a Pod's lifecycle, when it might take a long
time to load data or warm a cache, than during steady-state operation. This
cannot be updated. More info:
https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes

- rule: Set at most one of exec, httpGet, tcpSocket, grpc

### spec.podTemplate.spec.initContainers[].startupProbe.exec

`KubernetesCoreV1ApiExecAction`

`exec` specifies a command to execute in the container.

### spec.podTemplate.spec.initContainers[].startupProbe.exec.command

`[]string`

`command` is the command line to execute inside the container, the working
directory for the command  is root ('/') in the container's filesystem. The
command is simply exec'd, it is not run inside a shell, so traditional
shell instructions ('|', etc) won't work. To use a shell, you need to
explicitly call out to that shell. Exit status of 0 is treated as
live/healthy and non-zero is unhealthy.

### spec.podTemplate.spec.initContainers[].startupProbe.httpGet

`KubernetesCoreV1ApiHTTPGetAction`

`httpGet` specifies an HTTP GET request to perform.

### spec.podTemplate.spec.initContainers[].startupProbe.httpGet.path

`string`

`path` to access on the HTTP server.

### spec.podTemplate.spec.initContainers[].startupProbe.httpGet.port

`string` · required

Name or number of the port to access on the container. Number must be in
the range 1 to 65535. Name must be an IANA_SVC_NAME.

- rule: port is a port number from 1 to 65535, or a port name (up to 15 lowercase letters, digits and hyphens, with at least one letter)
- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.initContainers[].startupProbe.httpGet.host

`string`

`host` name to connect to, defaults to the pod IP. You probably want to set
"Host" in httpHeaders instead.

### spec.podTemplate.spec.initContainers[].startupProbe.httpGet.scheme

`string` · optional (explicit presence)

`scheme` to use for connecting to the host. Defaults to HTTP.

- rule: {"string":{"in":["HTTP","HTTPS"]}}

### spec.podTemplate.spec.initContainers[].startupProbe.httpGet.httpHeaders

`[]KubernetesCoreV1ApiHTTPHeader`

Custom headers to set in the request. HTTP allows repeated headers.

### spec.podTemplate.spec.initContainers[].startupProbe.httpGet.httpHeaders[].name

`string` · required

The header field name. This will be canonicalized upon output, so
case-variant names will be understood as the same header.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.initContainers[].startupProbe.httpGet.httpHeaders[].value

`string` · required

The header field value.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.initContainers[].startupProbe.httpGet.protocol

`string` · optional (explicit presence)

`protocol` selects the wire protocol for the probe connection. Nil defaults
to HTTP/1.1.

Requires the H2CContainerProbe feature gate on the cluster; a cluster
without it drops this field from the pod.

- rule: {"string":{"in":["HTTP1","HTTP2"]}}

### spec.podTemplate.spec.initContainers[].startupProbe.tcpSocket

`KubernetesCoreV1ApiTCPSocketAction`

`tcpSocket` specifies a connection to a TCP port.

### spec.podTemplate.spec.initContainers[].startupProbe.tcpSocket.port

`string` · required

Number or name of the port to access on the container. Number must be in
the range 1 to 65535. Name must be an IANA_SVC_NAME.

- rule: port is a port number from 1 to 65535, or a port name (up to 15 lowercase letters, digits and hyphens, with at least one letter)
- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.initContainers[].startupProbe.tcpSocket.host

`string`

Optional: Host name to connect to, defaults to the pod IP.

### spec.podTemplate.spec.initContainers[].startupProbe.grpc

`KubernetesCoreV1ApiGRPCAction`

`grpc` specifies a GRPC HealthCheckRequest.

### spec.podTemplate.spec.initContainers[].startupProbe.grpc.port

`int32` · required · optional (explicit presence)

`port` number of the gRPC service. Number must be in the range 1 to 65535.

- rule: {"required":true}

### spec.podTemplate.spec.initContainers[].startupProbe.grpc.service

`string` · optional (explicit presence)

`service` is the name of the service to place in the gRPC
HealthCheckRequest (see
https://github.com/grpc/grpc/blob/master/doc/health-checking.md).

If this is not specified, the default behavior is defined by gRPC.

### spec.podTemplate.spec.initContainers[].startupProbe.grpc.mode

`string` · optional (explicit presence)

mode specifies the connection mode for the gRPC health probe. Set to "TLS"
to use TLS without certificate verification. Set to "Plaintext" to use a
plaintext (insecure) connection explicitly. If not specified, the probe
uses a plaintext (insecure) connection.

Requires the GRPCContainerProbeTLS feature gate on the cluster; a cluster
without it drops this field from the pod.

- rule: {"string":{"in":["Plaintext","TLS"]}}

### spec.podTemplate.spec.initContainers[].startupProbe.initialDelaySeconds

`int32`

Number of seconds after the container has started before liveness probes
are initiated. More info:
https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes

### spec.podTemplate.spec.initContainers[].startupProbe.timeoutSeconds

`int32`

Number of seconds after which the probe times out. Defaults to 1 second.
Minimum value is 1. More info:
https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes

### spec.podTemplate.spec.initContainers[].startupProbe.periodSeconds

`int32`

How often (in seconds) to perform the probe. Default to 10 seconds. Minimum
value is 1.

### spec.podTemplate.spec.initContainers[].startupProbe.successThreshold

`int32`

Minimum consecutive successes for the probe to be considered successful
after having failed. Defaults to 1. Must be 1 for liveness and startup.
Minimum value is 1.

### spec.podTemplate.spec.initContainers[].startupProbe.failureThreshold

`int32`

Minimum consecutive failures for the probe to be considered failed after
having succeeded. Defaults to 3. Minimum value is 1.

### spec.podTemplate.spec.initContainers[].startupProbe.terminationGracePeriodSeconds

`int64` · optional (explicit presence)

Optional duration in seconds the pod needs to terminate gracefully upon
probe failure. The grace period is the duration in seconds after the
processes running in the pod are sent a termination signal and the time
when the processes are forcibly halted with a kill signal. Set this value
longer than the expected cleanup time for your process. If this value is
nil, the pod's terminationGracePeriodSeconds will be used. Otherwise, this
value overrides the value provided by the pod spec. Value must be
non-negative integer. The value zero indicates stop immediately via the
kill signal (no opportunity to shut down). This is a beta field and
requires enabling ProbeTerminationGracePeriod feature gate. Minimum value
is 1. spec.terminationGracePeriodSeconds is used if unset.

### spec.podTemplate.spec.initContainers[].lifecycle

`KubernetesCoreV1ApiLifecycle`

Actions that the management system should take in response to container
lifecycle events. Cannot be updated.

### spec.podTemplate.spec.initContainers[].lifecycle.postStart

`KubernetesCoreV1ApiLifecycleHandler`

`postStart` is called immediately after a container is created. If the
handler fails, the container is terminated and restarted according to its
restart policy. Other management of the container blocks until the hook
completes. More info:
https://kubernetes.io/docs/concepts/containers/container-lifecycle-hooks/#container-hooks

- rule: Set exactly one of exec, httpGet, sleep, tcpSocket

### spec.podTemplate.spec.initContainers[].lifecycle.postStart.exec

`KubernetesCoreV1ApiExecAction`

`exec` specifies a command to execute in the container.

### spec.podTemplate.spec.initContainers[].lifecycle.postStart.exec.command

`[]string`

`command` is the command line to execute inside the container, the working
directory for the command  is root ('/') in the container's filesystem. The
command is simply exec'd, it is not run inside a shell, so traditional
shell instructions ('|', etc) won't work. To use a shell, you need to
explicitly call out to that shell. Exit status of 0 is treated as
live/healthy and non-zero is unhealthy.

### spec.podTemplate.spec.initContainers[].lifecycle.postStart.httpGet

`KubernetesCoreV1ApiHTTPGetAction`

`httpGet` specifies an HTTP GET request to perform.

### spec.podTemplate.spec.initContainers[].lifecycle.postStart.httpGet.path

`string`

`path` to access on the HTTP server.

### spec.podTemplate.spec.initContainers[].lifecycle.postStart.httpGet.port

`string` · required

Name or number of the port to access on the container. Number must be in
the range 1 to 65535. Name must be an IANA_SVC_NAME.

- rule: port is a port number from 1 to 65535, or a port name (up to 15 lowercase letters, digits and hyphens, with at least one letter)
- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.initContainers[].lifecycle.postStart.httpGet.host

`string`

`host` name to connect to, defaults to the pod IP. You probably want to set
"Host" in httpHeaders instead.

### spec.podTemplate.spec.initContainers[].lifecycle.postStart.httpGet.scheme

`string` · optional (explicit presence)

`scheme` to use for connecting to the host. Defaults to HTTP.

- rule: {"string":{"in":["HTTP","HTTPS"]}}

### spec.podTemplate.spec.initContainers[].lifecycle.postStart.httpGet.httpHeaders

`[]KubernetesCoreV1ApiHTTPHeader`

Custom headers to set in the request. HTTP allows repeated headers.

### spec.podTemplate.spec.initContainers[].lifecycle.postStart.httpGet.httpHeaders[].name

`string` · required

The header field name. This will be canonicalized upon output, so
case-variant names will be understood as the same header.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.initContainers[].lifecycle.postStart.httpGet.httpHeaders[].value

`string` · required

The header field value.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.initContainers[].lifecycle.postStart.httpGet.protocol

`string` · optional (explicit presence)

`protocol` selects the wire protocol for the probe connection. Nil defaults
to HTTP/1.1.

Requires the H2CContainerProbe feature gate on the cluster; a cluster
without it drops this field from the pod.

- rule: {"string":{"in":["HTTP1","HTTP2"]}}

### spec.podTemplate.spec.initContainers[].lifecycle.postStart.tcpSocket

`KubernetesCoreV1ApiTCPSocketAction`

Deprecated. TCPSocket is NOT supported as a LifecycleHandler and kept for
backward compatibility. There is no validation of this field and lifecycle
hooks will fail at runtime when it is specified.

### spec.podTemplate.spec.initContainers[].lifecycle.postStart.tcpSocket.port

`string` · required

Number or name of the port to access on the container. Number must be in
the range 1 to 65535. Name must be an IANA_SVC_NAME.

- rule: port is a port number from 1 to 65535, or a port name (up to 15 lowercase letters, digits and hyphens, with at least one letter)
- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.initContainers[].lifecycle.postStart.tcpSocket.host

`string`

Optional: Host name to connect to, defaults to the pod IP.

### spec.podTemplate.spec.initContainers[].lifecycle.postStart.sleep

`KubernetesCoreV1ApiSleepAction`

`sleep` represents a duration that the container should sleep.

### spec.podTemplate.spec.initContainers[].lifecycle.postStart.sleep.seconds

`int64` · required · optional (explicit presence)

`seconds` is the number of seconds to sleep.

- rule: {"required":true}

### spec.podTemplate.spec.initContainers[].lifecycle.preStop

`KubernetesCoreV1ApiLifecycleHandler`

`preStop` is called immediately before a container is terminated due to an
API request or management event such as liveness/startup probe failure,
preemption, resource contention, etc. The handler is not called if the
container crashes or exits. The Pod's termination grace period countdown
begins before the PreStop hook is executed. Regardless of the outcome of
the handler, the container will eventually terminate within the Pod's
termination grace period (unless delayed by finalizers). Other management
of the container blocks until the hook completes or until the termination
grace period is reached. More info:
https://kubernetes.io/docs/concepts/containers/container-lifecycle-hooks/#container-hooks

- rule: Set exactly one of exec, httpGet, sleep, tcpSocket

### spec.podTemplate.spec.initContainers[].lifecycle.preStop.exec

`KubernetesCoreV1ApiExecAction`

`exec` specifies a command to execute in the container.

### spec.podTemplate.spec.initContainers[].lifecycle.preStop.exec.command

`[]string`

`command` is the command line to execute inside the container, the working
directory for the command  is root ('/') in the container's filesystem. The
command is simply exec'd, it is not run inside a shell, so traditional
shell instructions ('|', etc) won't work. To use a shell, you need to
explicitly call out to that shell. Exit status of 0 is treated as
live/healthy and non-zero is unhealthy.

### spec.podTemplate.spec.initContainers[].lifecycle.preStop.httpGet

`KubernetesCoreV1ApiHTTPGetAction`

`httpGet` specifies an HTTP GET request to perform.

### spec.podTemplate.spec.initContainers[].lifecycle.preStop.httpGet.path

`string`

`path` to access on the HTTP server.

### spec.podTemplate.spec.initContainers[].lifecycle.preStop.httpGet.port

`string` · required

Name or number of the port to access on the container. Number must be in
the range 1 to 65535. Name must be an IANA_SVC_NAME.

- rule: port is a port number from 1 to 65535, or a port name (up to 15 lowercase letters, digits and hyphens, with at least one letter)
- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.initContainers[].lifecycle.preStop.httpGet.host

`string`

`host` name to connect to, defaults to the pod IP. You probably want to set
"Host" in httpHeaders instead.

### spec.podTemplate.spec.initContainers[].lifecycle.preStop.httpGet.scheme

`string` · optional (explicit presence)

`scheme` to use for connecting to the host. Defaults to HTTP.

- rule: {"string":{"in":["HTTP","HTTPS"]}}

### spec.podTemplate.spec.initContainers[].lifecycle.preStop.httpGet.httpHeaders

`[]KubernetesCoreV1ApiHTTPHeader`

Custom headers to set in the request. HTTP allows repeated headers.

### spec.podTemplate.spec.initContainers[].lifecycle.preStop.httpGet.httpHeaders[].name

`string` · required

The header field name. This will be canonicalized upon output, so
case-variant names will be understood as the same header.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.initContainers[].lifecycle.preStop.httpGet.httpHeaders[].value

`string` · required

The header field value.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.initContainers[].lifecycle.preStop.httpGet.protocol

`string` · optional (explicit presence)

`protocol` selects the wire protocol for the probe connection. Nil defaults
to HTTP/1.1.

Requires the H2CContainerProbe feature gate on the cluster; a cluster
without it drops this field from the pod.

- rule: {"string":{"in":["HTTP1","HTTP2"]}}

### spec.podTemplate.spec.initContainers[].lifecycle.preStop.tcpSocket

`KubernetesCoreV1ApiTCPSocketAction`

Deprecated. TCPSocket is NOT supported as a LifecycleHandler and kept for
backward compatibility. There is no validation of this field and lifecycle
hooks will fail at runtime when it is specified.

### spec.podTemplate.spec.initContainers[].lifecycle.preStop.tcpSocket.port

`string` · required

Number or name of the port to access on the container. Number must be in
the range 1 to 65535. Name must be an IANA_SVC_NAME.

- rule: port is a port number from 1 to 65535, or a port name (up to 15 lowercase letters, digits and hyphens, with at least one letter)
- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.initContainers[].lifecycle.preStop.tcpSocket.host

`string`

Optional: Host name to connect to, defaults to the pod IP.

### spec.podTemplate.spec.initContainers[].lifecycle.preStop.sleep

`KubernetesCoreV1ApiSleepAction`

`sleep` represents a duration that the container should sleep.

### spec.podTemplate.spec.initContainers[].lifecycle.preStop.sleep.seconds

`int64` · required · optional (explicit presence)

`seconds` is the number of seconds to sleep.

- rule: {"required":true}

### spec.podTemplate.spec.initContainers[].lifecycle.stopSignal

`string` · optional (explicit presence)

`stopSignal` defines which signal will be sent to a container when it is
being stopped. If not specified, the default is defined by the container
runtime in use. StopSignal can only be set for Pods with a non-empty
.spec.os.name.

- rule: {"string":{"in":["SIGABRT","SIGALRM","SIGBUS","SIGCHLD","SIGCLD","SIGCONT","SIGFPE","SIGHUP","SIGILL","SIGINT","SIGIO","SIGIOT","SIGKILL","SIGPIPE","SIGPOLL","SIGPROF","SIGPWR","SIGQUIT","SIGSEGV","SIGSTKFLT","SIGSTOP","SIGSYS","SIGTERM","SIGTRAP","SIGTSTP","SIGTTIN","SIGTTOU","SIGURG","SIGUSR1","SIGUSR2","SIGVTALRM","SIGWINCH","SIGXCPU","SIGXFSZ","SIGRTMIN","SIGRTMIN+1","SIGRTMIN+2","SIGRTMIN+3","SIGRTMIN+4","SIGRTMIN+5","SIGRTMIN+6","SIGRTMIN+7","SIGRTMIN+8","SIGRTMIN+9","SIGRTMIN+10","SIGRTMIN+11","SIGRTMIN+12","SIGRTMIN+13","SIGRTMIN+14","SIGRTMIN+15","SIGRTMAX-14","SIGRTMAX-13","SIGRTMAX-12","SIGRTMAX-11","SIGRTMAX-10","SIGRTMAX-9","SIGRTMAX-8","SIGRTMAX-7","SIGRTMAX-6","SIGRTMAX-5","SIGRTMAX-4","SIGRTMAX-3","SIGRTMAX-2","SIGRTMAX-1","SIGRTMAX"]}}

### spec.podTemplate.spec.initContainers[].terminationMessagePath

`string`

Optional: Path at which the file to which the container's termination
message will be written is mounted into the container's filesystem. Message
written is intended to be brief final status, such as an assertion failure
message. Will be truncated by the node if greater than 4096 bytes. The
total message length across all containers will be limited to 12kb.
Defaults to /dev/termination-log. Cannot be updated.

### spec.podTemplate.spec.initContainers[].terminationMessagePolicy

`string` · optional (explicit presence)

Indicate how the termination message should be populated. File will use the
contents of terminationMessagePath to populate the container status message
on both success and failure. FallbackToLogsOnError will use the last chunk
of container log output if the termination message file is empty and the
container exited with an error. The log output is limited to 2048 bytes or
80 lines, whichever is smaller. Defaults to File. Cannot be updated.

- rule: {"string":{"in":["File","FallbackToLogsOnError"]}}

### spec.podTemplate.spec.initContainers[].imagePullPolicy

`string` · optional (explicit presence)

Image pull policy. One of Always, Never, IfNotPresent. Defaults to Always
if :latest tag is specified, or IfNotPresent otherwise. Cannot be updated.
More info:
https://kubernetes.io/docs/concepts/containers/images#updating-images

- rule: {"string":{"in":["Always","Never","IfNotPresent"]}}

### spec.podTemplate.spec.initContainers[].securityContext

`KubernetesCoreV1ApiSecurityContext`

`securityContext` defines the security options the container should be run
with. If set, the fields of SecurityContext override the equivalent fields
of PodSecurityContext. More info:
https://kubernetes.io/docs/tasks/configure-pod-container/security-context/

### spec.podTemplate.spec.initContainers[].securityContext.capabilities

`KubernetesCoreV1ApiCapabilities`

The capabilities to add/drop when running containers. Defaults to the
default set of capabilities granted by the container runtime. Note that
this field cannot be set when spec.os.name is windows.

### spec.podTemplate.spec.initContainers[].securityContext.capabilities.add

`[]string`

Added capabilities.

### spec.podTemplate.spec.initContainers[].securityContext.capabilities.drop

`[]string`

Removed capabilities.

### spec.podTemplate.spec.initContainers[].securityContext.privileged

`bool` · optional (explicit presence)

Run container in privileged mode. Processes in privileged containers are
essentially equivalent to root on the host. Defaults to false. Note that
this field cannot be set when spec.os.name is windows.

### spec.podTemplate.spec.initContainers[].securityContext.seLinuxOptions

`KubernetesCoreV1ApiSELinuxOptions`

The SELinux context to be applied to the container. If unspecified, the
container runtime will allocate a random SELinux context for each
container.  May also be set in PodSecurityContext.  If set in both
SecurityContext and PodSecurityContext, the value specified in
SecurityContext takes precedence. Note that this field cannot be set when
spec.os.name is windows.

### spec.podTemplate.spec.initContainers[].securityContext.seLinuxOptions.user

`string`

`user` is a SELinux user label that applies to the container.

### spec.podTemplate.spec.initContainers[].securityContext.seLinuxOptions.role

`string`

`role` is a SELinux role label that applies to the container.

### spec.podTemplate.spec.initContainers[].securityContext.seLinuxOptions.type

`string`

`type` is a SELinux type label that applies to the container.

### spec.podTemplate.spec.initContainers[].securityContext.seLinuxOptions.level

`string`

`level` is SELinux level label that applies to the container.

### spec.podTemplate.spec.initContainers[].securityContext.windowsOptions

`KubernetesCoreV1ApiWindowsSecurityContextOptions`

The Windows specific settings applied to all containers. If unspecified,
the options from the PodSecurityContext will be used. If set in both
SecurityContext and PodSecurityContext, the value specified in
SecurityContext takes precedence. Note that this field cannot be set when
spec.os.name is linux.

### spec.podTemplate.spec.initContainers[].securityContext.windowsOptions.gmsaCredentialSpecName

`string` · optional (explicit presence)

`gmsaCredentialSpecName` is the name of the GMSA credential spec to use.

### spec.podTemplate.spec.initContainers[].securityContext.windowsOptions.gmsaCredentialSpec

`string` · optional (explicit presence)

`gmsaCredentialSpec` is where the GMSA admission webhook
(https://github.com/kubernetes-sigs/windows-gmsa) inlines the contents of
the GMSA credential spec named by the GMSACredentialSpecName field.

### spec.podTemplate.spec.initContainers[].securityContext.windowsOptions.runAsUserName

`string` · optional (explicit presence)

The UserName in Windows to run the entrypoint of the container process.
Defaults to the user specified in image metadata if unspecified. May also
be set in PodSecurityContext. If set in both SecurityContext and
PodSecurityContext, the value specified in SecurityContext takes
precedence.

### spec.podTemplate.spec.initContainers[].securityContext.windowsOptions.hostProcess

`bool` · optional (explicit presence)

`hostProcess` determines if a container should be run as a 'Host Process'
container. All of a Pod's containers must have the same effective
HostProcess value (it is not allowed to have a mix of HostProcess
containers and non-HostProcess containers). In addition, if HostProcess is
true then HostNetwork must also be set to true.

### spec.podTemplate.spec.initContainers[].securityContext.runAsUser

`int64` · optional (explicit presence)

The UID to run the entrypoint of the container process. Defaults to user
specified in image metadata if unspecified. May also be set in
PodSecurityContext.  If set in both SecurityContext and PodSecurityContext,
the value specified in SecurityContext takes precedence. Note that this
field cannot be set when spec.os.name is windows.

### spec.podTemplate.spec.initContainers[].securityContext.runAsGroup

`int64` · optional (explicit presence)

The GID to run the entrypoint of the container process. Uses runtime
default if unset. May also be set in PodSecurityContext.  If set in both
SecurityContext and PodSecurityContext, the value specified in
SecurityContext takes precedence. Note that this field cannot be set when
spec.os.name is windows.

### spec.podTemplate.spec.initContainers[].securityContext.runAsNonRoot

`bool` · optional (explicit presence)

Indicates that the container must run as a non-root user. If true, the
Kubelet will validate the image at runtime to ensure that it does not run
as UID 0 (root) and fail to start the container if it does. If unset or
false, no such validation will be performed. May also be set in
PodSecurityContext.  If set in both SecurityContext and PodSecurityContext,
the value specified in SecurityContext takes precedence.

### spec.podTemplate.spec.initContainers[].securityContext.readOnlyRootFilesystem

`bool` · optional (explicit presence)

Whether this container has a read-only root filesystem. Default is false.
Note that this field cannot be set when spec.os.name is windows.

### spec.podTemplate.spec.initContainers[].securityContext.allowPrivilegeEscalation

`bool` · optional (explicit presence)

`allowPrivilegeEscalation` controls whether a process can gain more
privileges than its parent process. This bool directly controls if the
no_new_privs flag will be set on the container process.
AllowPrivilegeEscalation is true always when the container is: 1) run as
Privileged 2) has CAP_SYS_ADMIN Note that this field cannot be set when
spec.os.name is windows.

### spec.podTemplate.spec.initContainers[].securityContext.procMount

`string` · optional (explicit presence)

procMount denotes the type of proc mount to use for the containers. The
default value is Default which uses the container runtime defaults for
readonly paths and masked paths. Note that this field cannot be set when
spec.os.name is windows.

- rule: {"string":{"in":["Default","Unmasked"]}}

### spec.podTemplate.spec.initContainers[].securityContext.seccompProfile

`KubernetesCoreV1ApiSeccompProfile`

The seccomp options to use by this container. If seccomp options are
provided at both the pod & container level, the container options override
the pod options. Note that this field cannot be set when spec.os.name is
windows.

### spec.podTemplate.spec.initContainers[].securityContext.seccompProfile.type

`string`

type indicates which kind of seccomp profile will be applied. Valid options
are:

Localhost - a profile defined in a file on the node should be used.
RuntimeDefault - the container runtime default profile should be used.
Unconfined - no profile should be applied.

- rule: {"string":{"in":["Unconfined","RuntimeDefault","Localhost"]}}

### spec.podTemplate.spec.initContainers[].securityContext.seccompProfile.localhostProfile

`string` · optional (explicit presence)

localhostProfile indicates a profile defined in a file on the node should
be used. The profile must be preconfigured on the node to work. Must be a
descending path, relative to the kubelet's configured seccomp profile
location. Must be set if type is "Localhost". Must NOT be set for any other
type.

### spec.podTemplate.spec.initContainers[].securityContext.appArmorProfile

`KubernetesCoreV1ApiAppArmorProfile`

appArmorProfile is the AppArmor options to use by this container. If set,
this profile overrides the pod's appArmorProfile. Note that this field
cannot be set when spec.os.name is windows.

### spec.podTemplate.spec.initContainers[].securityContext.appArmorProfile.type

`string`

type indicates which kind of AppArmor profile will be applied. Valid
options are: Localhost - a profile pre-loaded on the node. RuntimeDefault -
the container runtime's default profile. Unconfined - no AppArmor
enforcement.

- rule: {"string":{"in":["Unconfined","RuntimeDefault","Localhost"]}}

### spec.podTemplate.spec.initContainers[].securityContext.appArmorProfile.localhostProfile

`string` · optional (explicit presence)

localhostProfile indicates a profile loaded on the node that should be
used. The profile must be preconfigured on the node to work. Must match the
loaded name of the profile. Must be set if and only if type is "Localhost".

### spec.podTemplate.spec.initContainers[].stdin

`bool`

Whether this container should allocate a buffer for stdin in the container
runtime. If this is not set, reads from stdin in the container will always
result in EOF. Default is false.

### spec.podTemplate.spec.initContainers[].stdinOnce

`bool`

Whether the container runtime should close the stdin channel after it has
been opened by a single attach. When stdin is true the stdin stream will
remain open across multiple attach sessions. If stdinOnce is set to true,
stdin is opened on container start, is empty until the first client
attaches to stdin, and then remains open and accepts data until the client
disconnects, at which time stdin is closed and remains closed until the
container is restarted. If this flag is false, a container processes that
reads from stdin will never receive an EOF. Default is false.

### spec.podTemplate.spec.initContainers[].tty

`bool`

Whether this container should allocate a TTY for itself, also requires
'stdin' to be true. Default is false.

### spec.podTemplate.spec.containers

`[]KubernetesCoreV1ApiContainer` · required

List of containers belonging to the pod. Containers cannot currently be
added or removed. There must be at least one container in a Pod. Cannot be
updated.

- rule: {"repeated":{"minItems":"1"}}

### spec.podTemplate.spec.containers[].name

`string` · required

`name` of the container specified as a DNS_LABEL. Each container in a pod
must have a unique name (DNS_LABEL). Cannot be updated.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.containers[].image

`string`

Container image name. More info:
https://kubernetes.io/docs/concepts/containers/images This field is
optional to allow higher level config management to default or override
container images in workload controllers like Deployments and StatefulSets.

### spec.podTemplate.spec.containers[].command

`[]string`

Entrypoint array. Not executed within a shell. The container image's
ENTRYPOINT is used if this is not provided. Variable references $(VAR_NAME)
are expanded using the container's environment. If a variable cannot be
resolved, the reference in the input string will be unchanged. Double $$
are reduced to a single $, which allows for escaping the $(VAR_NAME)
syntax: i.e. "$$(VAR_NAME)" will produce the string literal "$(VAR_NAME)".
Escaped references will never be expanded, regardless of whether the
variable exists or not. Cannot be updated. More info:
https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell

### spec.podTemplate.spec.containers[].args

`[]string`

Arguments to the entrypoint. The container image's CMD is used if this is
not provided. Variable references $(VAR_NAME) are expanded using the
container's environment. If a variable cannot be resolved, the reference in
the input string will be unchanged. Double $$ are reduced to a single $,
which allows for escaping the $(VAR_NAME) syntax: i.e. "$$(VAR_NAME)" will
produce the string literal "$(VAR_NAME)". Escaped references will never be
expanded, regardless of whether the variable exists or not. Cannot be
updated. More info:
https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell

### spec.podTemplate.spec.containers[].workingDir

`string`

Container's working directory. If not specified, the container runtime's
default will be used, which might be configured in the container image.
Cannot be updated.

### spec.podTemplate.spec.containers[].ports

`[]KubernetesCoreV1ApiContainerPort`

List of ports to expose from the container. Not specifying a port here DOES
NOT prevent that port from being exposed. Any port which is listening on
the default "0.0.0.0" address inside a container will be accessible from
the network. Modifying this array with strategic merge patch may corrupt
the data. For more information See
https://github.com/kubernetes/kubernetes/issues/108255. Cannot be updated.

### spec.podTemplate.spec.containers[].ports[].name

`string`

If specified, this must be an IANA_SVC_NAME and unique within the pod. Each
named port in a pod must have a unique name. Name for the port that can be
referred to by services.

### spec.podTemplate.spec.containers[].ports[].hostPort

`int32`

Number of port to expose on the host. If specified, this must be a valid
port number, 0 < x < 65536. If HostNetwork is specified, this must match
ContainerPort. Most containers do not need this.

### spec.podTemplate.spec.containers[].ports[].containerPort

`int32` · required · optional (explicit presence)

Number of port to expose on the pod's IP address. This must be a valid port
number, 0 < x < 65536.

- rule: {"required":true}

### spec.podTemplate.spec.containers[].ports[].protocol

`string` · optional (explicit presence)

`protocol` for port. Must be UDP, TCP, or SCTP. Defaults to "TCP".

Unset, the cluster writes "TCP".

- rule: {"string":{"in":["TCP","UDP","SCTP"]}}

### spec.podTemplate.spec.containers[].ports[].hostIP

`string`

What host IP to bind the external port to.

### spec.podTemplate.spec.containers[].envFrom

`[]KubernetesCoreV1ApiEnvFromSource`

List of sources to populate environment variables in the container. The
keys defined within a source may consist of any printable ASCII characters
except '='. When a key exists in multiple sources, the value associated
with the last source will take precedence. Values defined by an Env with a
duplicate key will take precedence. Cannot be updated.

- rule: Set exactly one of configMapRef, secretRef

### spec.podTemplate.spec.containers[].envFrom[].prefix

`string`

Optional text to prepend to the name of each environment variable. May
consist of any printable ASCII characters except '='.

### spec.podTemplate.spec.containers[].envFrom[].configMapRef

`KubernetesCoreV1ApiConfigMapEnvSource`

The ConfigMap to select from.

### spec.podTemplate.spec.containers[].envFrom[].configMapRef.name

`string | valueFrom` · required

`name` of the referent. This field is effectively required, but due to
backwards compatibility is allowed to be empty. Instances of this type with
an empty value here are almost certainly wrong. More info:
https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names

Defaults to a KubernetesConfigMap reference (its configmap_name output);
pass the literal name with `value:` for one Planton does not manage.

- references: KubernetesConfigMap (`status.outputs.configmap_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesConfigMap, name: <that resource's name>, fieldPath: status.outputs.configmap_name}} -- a bare string does not parse

### spec.podTemplate.spec.containers[].envFrom[].configMapRef.optional

`bool` · optional (explicit presence)

Specify whether the ConfigMap must be defined.

### spec.podTemplate.spec.containers[].envFrom[].secretRef

`KubernetesCoreV1ApiSecretEnvSource`

The Secret to select from.

### spec.podTemplate.spec.containers[].envFrom[].secretRef.name

`string | valueFrom` · required

`name` of the referent. This field is effectively required, but due to
backwards compatibility is allowed to be empty. Instances of this type with
an empty value here are almost certainly wrong. More info:
https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names

Defaults to a KubernetesSecret reference (its secret_name output); pass the
literal name with `value:` for one Planton does not manage.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.podTemplate.spec.containers[].envFrom[].secretRef.optional

`bool` · optional (explicit presence)

Specify whether the Secret must be defined.

### spec.podTemplate.spec.containers[].env

`[]KubernetesCoreV1ApiEnvVar`

List of environment variables to set in the container. Cannot be updated.

### spec.podTemplate.spec.containers[].env[].name

`string` · required

`name` of the environment variable. May consist of any printable ASCII
characters except '='.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.containers[].env[].value

`string`

Variable references $(VAR_NAME) are expanded using the previously defined
environment variables in the container and any service environment
variables. If a variable cannot be resolved, the reference in the input
string will be unchanged. Double $$ are reduced to a single $, which allows
for escaping the $(VAR_NAME) syntax: i.e. "$$(VAR_NAME)" will produce the
string literal "$(VAR_NAME)". Escaped references will never be expanded,
regardless of whether the variable exists or not. Defaults to "".

### spec.podTemplate.spec.containers[].env[].valueFrom

`KubernetesCoreV1ApiEnvVarSource`

Source for the environment variable's value. Cannot be used if value is not
empty.

- rule: Set exactly one of configMapKeyRef, fieldRef, fileKeyRef, resourceFieldRef, secretKeyRef

### spec.podTemplate.spec.containers[].env[].valueFrom.fieldRef

`KubernetesCoreV1ApiObjectFieldSelector`

Selects a field of the pod: supports metadata.name, metadata.namespace,
`metadata.labels['<KEY>']`, `metadata.annotations['<KEY>']`, spec.nodeName,
spec.serviceAccountName, status.hostIP, status.podIP, status.podIPs.

### spec.podTemplate.spec.containers[].env[].valueFrom.fieldRef.apiVersion

`string`

Version of the schema the FieldPath is written in terms of, defaults to
"v1".

### spec.podTemplate.spec.containers[].env[].valueFrom.fieldRef.fieldPath

`string` · required

Path of the field to select in the specified API version.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.containers[].env[].valueFrom.resourceFieldRef

`KubernetesCoreV1ApiResourceFieldSelector`

Selects a resource of the container: only resources limits and requests
(limits.cpu, limits.memory, limits.ephemeral-storage, requests.cpu,
requests.memory and requests.ephemeral-storage) are currently supported.

### spec.podTemplate.spec.containers[].env[].valueFrom.resourceFieldRef.containerName

`string`

Container name: required for volumes, optional for env vars.

### spec.podTemplate.spec.containers[].env[].valueFrom.resourceFieldRef.resource

`string` · required

Required: resource to select.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.containers[].env[].valueFrom.resourceFieldRef.divisor

`string` · optional (explicit presence)

Specifies the output format of the exposed resources, defaults to "1"

- rule: {"string":{"pattern":"^(\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))))?$"}}

### spec.podTemplate.spec.containers[].env[].valueFrom.configMapKeyRef

`KubernetesCoreV1ApiConfigMapKeySelector`

Selects a key of a ConfigMap.

### spec.podTemplate.spec.containers[].env[].valueFrom.configMapKeyRef.name

`string | valueFrom` · required

`name` of the referent. This field is effectively required, but due to
backwards compatibility is allowed to be empty. Instances of this type with
an empty value here are almost certainly wrong. More info:
https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names

Defaults to a KubernetesConfigMap reference (its configmap_name output);
pass the literal name with `value:` for one Planton does not manage.

- references: KubernetesConfigMap (`status.outputs.configmap_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesConfigMap, name: <that resource's name>, fieldPath: status.outputs.configmap_name}} -- a bare string does not parse

### spec.podTemplate.spec.containers[].env[].valueFrom.configMapKeyRef.key

`string` · required

The key to select from the ConfigMap's Data field. Keys in the BinaryData
field are not currently propagated to container env vars.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.containers[].env[].valueFrom.configMapKeyRef.optional

`bool` · optional (explicit presence)

Specify whether the ConfigMap or its key must be defined.

### spec.podTemplate.spec.containers[].env[].valueFrom.secretKeyRef

`KubernetesCoreV1ApiSecretKeySelector`

Selects a key of a secret in the pod's namespace.

### spec.podTemplate.spec.containers[].env[].valueFrom.secretKeyRef.name

`string | valueFrom` · required

`name` of the referent. This field is effectively required, but due to
backwards compatibility is allowed to be empty. Instances of this type with
an empty value here are almost certainly wrong. More info:
https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names

Defaults to a KubernetesSecret reference (its secret_name output); pass the
literal name with `value:` for one Planton does not manage.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.podTemplate.spec.containers[].env[].valueFrom.secretKeyRef.key

`string` · required

The key of the secret to select from.  Must be a valid secret key.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.containers[].env[].valueFrom.secretKeyRef.optional

`bool` · optional (explicit presence)

Specify whether the Secret or its key must be defined.

### spec.podTemplate.spec.containers[].env[].valueFrom.fileKeyRef

`KubernetesCoreV1ApiFileKeySelector`

`fileKeyRef` selects a key of the env file. Requires the EnvFiles feature
gate to be enabled.

### spec.podTemplate.spec.containers[].env[].valueFrom.fileKeyRef.volumeName

`string` · required

The name of the volume mount containing the env file.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.containers[].env[].valueFrom.fileKeyRef.path

`string` · required

The path within the volume from which to select the file. Must be relative
and may not contain the '..' path or start with '..'.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.containers[].env[].valueFrom.fileKeyRef.key

`string` · required

The key within the env file. An invalid key will prevent the pod from
starting. The keys defined within a source may consist of any printable
ASCII characters except '='. During Alpha stage of the EnvFiles feature
gate, the key size is limited to 128 characters.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.containers[].env[].valueFrom.fileKeyRef.optional

`bool` · optional (explicit presence)

Specify whether the file or its key must be defined. If the file or key
does not exist, then the env var is not published. If optional is set to
true and the specified key does not exist, the environment variable will
not be set in the Pod's containers.

If optional is set to false and the specified key does not exist, an error
will be returned during Pod creation.

Unset, the cluster writes false.

### spec.podTemplate.spec.containers[].resources

`KubernetesCoreV1ApiResourceRequirements`

Compute Resources required by this container. Cannot be updated. More info:
https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/

### spec.podTemplate.spec.containers[].resources.limits

`map<string, string>`

`limits` describes the maximum amount of compute resources allowed. More
info:
https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/

- rule: {"map":{"values":{"string":{"pattern":"^(\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))))?$"}}}}

### spec.podTemplate.spec.containers[].resources.requests

`map<string, string>`

`requests` describes the minimum amount of compute resources required. If
Requests is omitted for a container, it defaults to Limits if that is
explicitly specified, otherwise to an implementation-defined value.
Requests cannot exceed Limits. More info:
https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/

- rule: {"map":{"values":{"string":{"pattern":"^(\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))))?$"}}}}

### spec.podTemplate.spec.containers[].resources.claims

`[]KubernetesCoreV1ApiResourceClaim`

`claims` lists the names of resources, defined in spec.resourceClaims, that
are used by this container.

This field depends on the DynamicResourceAllocation feature gate.

This field is immutable. It can only be set for containers.

### spec.podTemplate.spec.containers[].resources.claims[].name

`string` · required

`name` must match the name of one entry in pod.spec.resourceClaims of the
Pod where this field is used. It makes that resource available inside a
container.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.containers[].resources.claims[].request

`string`

`request` is the name chosen for a request in the referenced claim. If
empty, everything from the claim is made available, otherwise only the
result of this request.

### spec.podTemplate.spec.containers[].resizePolicy

`[]KubernetesCoreV1ApiContainerResizePolicy`

Resources resize policy for the container. This field cannot be set on
ephemeral containers.

Requires the InPlacePodVerticalScaling feature gate on the cluster; a
cluster without it drops this field from the pod.

### spec.podTemplate.spec.containers[].resizePolicy[].resourceName

`string` · required

Name of the resource to which this resource resize policy applies.
Supported values: cpu, memory.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.containers[].resizePolicy[].restartPolicy

`string`

Restart policy to apply when specified resource is resized. If not
specified, it defaults to NotRequired.

- rule: {"string":{"in":["NotRequired","RestartContainer"]}}

### spec.podTemplate.spec.containers[].restartPolicy

`string` · optional (explicit presence)

`restartPolicy` defines the restart behavior of individual containers in a
pod. This overrides the pod-level restart policy. When this field is not
specified, the restart behavior is defined by the Pod's restart policy and
the container type. Additionally, setting the RestartPolicy as "Always" for
the init container will have the following effect: this init container will
be continually restarted on exit until all regular containers have
terminated. Once all regular containers have completed, all init containers
with restartPolicy "Always" will be shut down. This lifecycle differs from
normal init containers and is often referred to as a "sidecar" container.
Although this init container still starts in the init container sequence,
it does not wait for the container to complete before proceeding to the
next init container. Instead, the next init container starts immediately
after this init container is started, or after any startupProbe has
successfully completed.

- rule: {"string":{"in":["Always","Never","OnFailure"]}}

### spec.podTemplate.spec.containers[].restartPolicyRules

`[]KubernetesCoreV1ApiContainerRestartRule`

Represents a list of rules to be checked to determine if the container
should be restarted on exit. The rules are evaluated in order. Once a rule
matches a container exit condition, the remaining rules are ignored. If no
rule matches the container exit condition, the Container-level restart
policy determines the whether the container is restarted or not.
Constraints on the rules: - At most 20 rules are allowed. - Rules can have
the same action. - Identical rules are not forbidden in validations. When
rules are specified, container MUST set RestartPolicy explicitly even it if
matches the Pod's RestartPolicy.

Requires the ContainerRestartRules feature gate on the cluster; a cluster
without it drops this field from the pod.

### spec.podTemplate.spec.containers[].restartPolicyRules[].action

`string`

Specifies the action taken on a container exit if the requirements are
satisfied. The only possible value is "Restart" to restart the container.

- rule: {"string":{"in":["Restart","RestartAllContainers"]}}

### spec.podTemplate.spec.containers[].restartPolicyRules[].exitCodes

`KubernetesCoreV1ApiContainerRestartRuleOnExitCodes`

Represents the exit codes to check on container exits.

### spec.podTemplate.spec.containers[].restartPolicyRules[].exitCodes.operator

`string`

Represents the relationship between the container exit code(s) and the
specified values. Possible values are: - In: the requirement is satisfied
if the container exit code is in the set of specified values. - NotIn: the
requirement is satisfied if the container exit code is not in the set of
specified values.

- rule: {"string":{"in":["In","NotIn"]}}

### spec.podTemplate.spec.containers[].restartPolicyRules[].exitCodes.values

`[]int32`

Specifies the set of values to check for container exit codes. At most 255
elements are allowed.

### spec.podTemplate.spec.containers[].volumeMounts

`[]KubernetesCoreV1ApiVolumeMount`

Pod volumes to mount into the container's filesystem. Cannot be updated.

### spec.podTemplate.spec.containers[].volumeMounts[].name

`string` · required

This must match the Name of a Volume.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.containers[].volumeMounts[].readOnly

`bool`

Mounted read-only if true, read-write otherwise (false or unspecified).
Defaults to false.

### spec.podTemplate.spec.containers[].volumeMounts[].recursiveReadOnly

`string` · optional (explicit presence)

`recursiveReadOnly` specifies whether read-only mounts should be handled
recursively.

If ReadOnly is false, this field has no meaning and must be unspecified.

If ReadOnly is true, and this field is set to Disabled, the mount is not
made recursively read-only.  If this field is set to IfPossible, the mount
is made recursively read-only, if it is supported by the container runtime.
If this field is set to Enabled, the mount is made recursively read-only if
it is supported by the container runtime, otherwise the pod will not be
started and an error will be generated to indicate the reason.

If this field is set to IfPossible or Enabled, MountPropagation must be set
to None (or be unspecified, which defaults to None).

If this field is not specified, it is treated as an equivalent of Disabled.

- rule: {"string":{"in":["Disabled","IfPossible","Enabled"]}}

### spec.podTemplate.spec.containers[].volumeMounts[].mountPath

`string` · required

Path within the container at which the volume should be mounted.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.containers[].volumeMounts[].subPath

`string`

Path within the volume from which the container's volume should be mounted.
Defaults to "" (volume's root).

### spec.podTemplate.spec.containers[].volumeMounts[].mountPropagation

`string` · optional (explicit presence)

mountPropagation determines how mounts are propagated from the host to
container and the other way around. When not set, MountPropagationNone is
used. This field is beta in 1.10. When RecursiveReadOnly is set to
IfPossible or to Enabled, MountPropagation must be None or unspecified
(which defaults to None).

- rule: {"string":{"in":["None","HostToContainer","Bidirectional"]}}

### spec.podTemplate.spec.containers[].volumeMounts[].subPathExpr

`string`

Expanded path within the volume from which the container's volume should be
mounted. Behaves similarly to SubPath but environment variable references
$(VAR_NAME) are expanded using the container's environment. Defaults to ""
(volume's root). SubPathExpr and SubPath are mutually exclusive.

### spec.podTemplate.spec.containers[].volumeMounts[].bindMountOptions

`[]string`

bindMountOptions is the list of additional bind mount options to apply when
mounting this volume into the container. Allowed values are noexec, nodev,
and nosuid. These are Linux mount options and have no effect on Windows
nodes. This field is not supported with image volumes. This is an alpha
field and requires enabling the VolumeBindMountOptions feature gate.

### spec.podTemplate.spec.containers[].volumeDevices

`[]KubernetesCoreV1ApiVolumeDevice`

volumeDevices is the list of block devices to be used by the container.

### spec.podTemplate.spec.containers[].volumeDevices[].name

`string` · required

name must match the name of a persistentVolumeClaim in the pod.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.containers[].volumeDevices[].devicePath

`string` · required

devicePath is the path inside of the container that the device will be
mapped to.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.containers[].livenessProbe

`KubernetesCoreV1ApiProbe`

Periodic probe of container liveness. Container will be restarted if the
probe fails. Cannot be updated. More info:
https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes

- rule: Set at most one of exec, httpGet, tcpSocket, grpc

### spec.podTemplate.spec.containers[].livenessProbe.exec

`KubernetesCoreV1ApiExecAction`

`exec` specifies a command to execute in the container.

### spec.podTemplate.spec.containers[].livenessProbe.exec.command

`[]string`

`command` is the command line to execute inside the container, the working
directory for the command  is root ('/') in the container's filesystem. The
command is simply exec'd, it is not run inside a shell, so traditional
shell instructions ('|', etc) won't work. To use a shell, you need to
explicitly call out to that shell. Exit status of 0 is treated as
live/healthy and non-zero is unhealthy.

### spec.podTemplate.spec.containers[].livenessProbe.httpGet

`KubernetesCoreV1ApiHTTPGetAction`

`httpGet` specifies an HTTP GET request to perform.

### spec.podTemplate.spec.containers[].livenessProbe.httpGet.path

`string`

`path` to access on the HTTP server.

### spec.podTemplate.spec.containers[].livenessProbe.httpGet.port

`string` · required

Name or number of the port to access on the container. Number must be in
the range 1 to 65535. Name must be an IANA_SVC_NAME.

- rule: port is a port number from 1 to 65535, or a port name (up to 15 lowercase letters, digits and hyphens, with at least one letter)
- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.containers[].livenessProbe.httpGet.host

`string`

`host` name to connect to, defaults to the pod IP. You probably want to set
"Host" in httpHeaders instead.

### spec.podTemplate.spec.containers[].livenessProbe.httpGet.scheme

`string` · optional (explicit presence)

`scheme` to use for connecting to the host. Defaults to HTTP.

- rule: {"string":{"in":["HTTP","HTTPS"]}}

### spec.podTemplate.spec.containers[].livenessProbe.httpGet.httpHeaders

`[]KubernetesCoreV1ApiHTTPHeader`

Custom headers to set in the request. HTTP allows repeated headers.

### spec.podTemplate.spec.containers[].livenessProbe.httpGet.httpHeaders[].name

`string` · required

The header field name. This will be canonicalized upon output, so
case-variant names will be understood as the same header.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.containers[].livenessProbe.httpGet.httpHeaders[].value

`string` · required

The header field value.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.containers[].livenessProbe.httpGet.protocol

`string` · optional (explicit presence)

`protocol` selects the wire protocol for the probe connection. Nil defaults
to HTTP/1.1.

Requires the H2CContainerProbe feature gate on the cluster; a cluster
without it drops this field from the pod.

- rule: {"string":{"in":["HTTP1","HTTP2"]}}

### spec.podTemplate.spec.containers[].livenessProbe.tcpSocket

`KubernetesCoreV1ApiTCPSocketAction`

`tcpSocket` specifies a connection to a TCP port.

### spec.podTemplate.spec.containers[].livenessProbe.tcpSocket.port

`string` · required

Number or name of the port to access on the container. Number must be in
the range 1 to 65535. Name must be an IANA_SVC_NAME.

- rule: port is a port number from 1 to 65535, or a port name (up to 15 lowercase letters, digits and hyphens, with at least one letter)
- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.containers[].livenessProbe.tcpSocket.host

`string`

Optional: Host name to connect to, defaults to the pod IP.

### spec.podTemplate.spec.containers[].livenessProbe.grpc

`KubernetesCoreV1ApiGRPCAction`

`grpc` specifies a GRPC HealthCheckRequest.

### spec.podTemplate.spec.containers[].livenessProbe.grpc.port

`int32` · required · optional (explicit presence)

`port` number of the gRPC service. Number must be in the range 1 to 65535.

- rule: {"required":true}

### spec.podTemplate.spec.containers[].livenessProbe.grpc.service

`string` · optional (explicit presence)

`service` is the name of the service to place in the gRPC
HealthCheckRequest (see
https://github.com/grpc/grpc/blob/master/doc/health-checking.md).

If this is not specified, the default behavior is defined by gRPC.

### spec.podTemplate.spec.containers[].livenessProbe.grpc.mode

`string` · optional (explicit presence)

mode specifies the connection mode for the gRPC health probe. Set to "TLS"
to use TLS without certificate verification. Set to "Plaintext" to use a
plaintext (insecure) connection explicitly. If not specified, the probe
uses a plaintext (insecure) connection.

Requires the GRPCContainerProbeTLS feature gate on the cluster; a cluster
without it drops this field from the pod.

- rule: {"string":{"in":["Plaintext","TLS"]}}

### spec.podTemplate.spec.containers[].livenessProbe.initialDelaySeconds

`int32`

Number of seconds after the container has started before liveness probes
are initiated. More info:
https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes

### spec.podTemplate.spec.containers[].livenessProbe.timeoutSeconds

`int32`

Number of seconds after which the probe times out. Defaults to 1 second.
Minimum value is 1. More info:
https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes

### spec.podTemplate.spec.containers[].livenessProbe.periodSeconds

`int32`

How often (in seconds) to perform the probe. Default to 10 seconds. Minimum
value is 1.

### spec.podTemplate.spec.containers[].livenessProbe.successThreshold

`int32`

Minimum consecutive successes for the probe to be considered successful
after having failed. Defaults to 1. Must be 1 for liveness and startup.
Minimum value is 1.

### spec.podTemplate.spec.containers[].livenessProbe.failureThreshold

`int32`

Minimum consecutive failures for the probe to be considered failed after
having succeeded. Defaults to 3. Minimum value is 1.

### spec.podTemplate.spec.containers[].livenessProbe.terminationGracePeriodSeconds

`int64` · optional (explicit presence)

Optional duration in seconds the pod needs to terminate gracefully upon
probe failure. The grace period is the duration in seconds after the
processes running in the pod are sent a termination signal and the time
when the processes are forcibly halted with a kill signal. Set this value
longer than the expected cleanup time for your process. If this value is
nil, the pod's terminationGracePeriodSeconds will be used. Otherwise, this
value overrides the value provided by the pod spec. Value must be
non-negative integer. The value zero indicates stop immediately via the
kill signal (no opportunity to shut down). This is a beta field and
requires enabling ProbeTerminationGracePeriod feature gate. Minimum value
is 1. spec.terminationGracePeriodSeconds is used if unset.

### spec.podTemplate.spec.containers[].readinessProbe

`KubernetesCoreV1ApiProbe`

Periodic probe of container service readiness. Container will be removed
from service endpoints if the probe fails. Cannot be updated. More info:
https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes

- rule: Set at most one of exec, httpGet, tcpSocket, grpc

### spec.podTemplate.spec.containers[].readinessProbe.exec

`KubernetesCoreV1ApiExecAction`

`exec` specifies a command to execute in the container.

### spec.podTemplate.spec.containers[].readinessProbe.exec.command

`[]string`

`command` is the command line to execute inside the container, the working
directory for the command  is root ('/') in the container's filesystem. The
command is simply exec'd, it is not run inside a shell, so traditional
shell instructions ('|', etc) won't work. To use a shell, you need to
explicitly call out to that shell. Exit status of 0 is treated as
live/healthy and non-zero is unhealthy.

### spec.podTemplate.spec.containers[].readinessProbe.httpGet

`KubernetesCoreV1ApiHTTPGetAction`

`httpGet` specifies an HTTP GET request to perform.

### spec.podTemplate.spec.containers[].readinessProbe.httpGet.path

`string`

`path` to access on the HTTP server.

### spec.podTemplate.spec.containers[].readinessProbe.httpGet.port

`string` · required

Name or number of the port to access on the container. Number must be in
the range 1 to 65535. Name must be an IANA_SVC_NAME.

- rule: port is a port number from 1 to 65535, or a port name (up to 15 lowercase letters, digits and hyphens, with at least one letter)
- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.containers[].readinessProbe.httpGet.host

`string`

`host` name to connect to, defaults to the pod IP. You probably want to set
"Host" in httpHeaders instead.

### spec.podTemplate.spec.containers[].readinessProbe.httpGet.scheme

`string` · optional (explicit presence)

`scheme` to use for connecting to the host. Defaults to HTTP.

- rule: {"string":{"in":["HTTP","HTTPS"]}}

### spec.podTemplate.spec.containers[].readinessProbe.httpGet.httpHeaders

`[]KubernetesCoreV1ApiHTTPHeader`

Custom headers to set in the request. HTTP allows repeated headers.

### spec.podTemplate.spec.containers[].readinessProbe.httpGet.httpHeaders[].name

`string` · required

The header field name. This will be canonicalized upon output, so
case-variant names will be understood as the same header.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.containers[].readinessProbe.httpGet.httpHeaders[].value

`string` · required

The header field value.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.containers[].readinessProbe.httpGet.protocol

`string` · optional (explicit presence)

`protocol` selects the wire protocol for the probe connection. Nil defaults
to HTTP/1.1.

Requires the H2CContainerProbe feature gate on the cluster; a cluster
without it drops this field from the pod.

- rule: {"string":{"in":["HTTP1","HTTP2"]}}

### spec.podTemplate.spec.containers[].readinessProbe.tcpSocket

`KubernetesCoreV1ApiTCPSocketAction`

`tcpSocket` specifies a connection to a TCP port.

### spec.podTemplate.spec.containers[].readinessProbe.tcpSocket.port

`string` · required

Number or name of the port to access on the container. Number must be in
the range 1 to 65535. Name must be an IANA_SVC_NAME.

- rule: port is a port number from 1 to 65535, or a port name (up to 15 lowercase letters, digits and hyphens, with at least one letter)
- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.containers[].readinessProbe.tcpSocket.host

`string`

Optional: Host name to connect to, defaults to the pod IP.

### spec.podTemplate.spec.containers[].readinessProbe.grpc

`KubernetesCoreV1ApiGRPCAction`

`grpc` specifies a GRPC HealthCheckRequest.

### spec.podTemplate.spec.containers[].readinessProbe.grpc.port

`int32` · required · optional (explicit presence)

`port` number of the gRPC service. Number must be in the range 1 to 65535.

- rule: {"required":true}

### spec.podTemplate.spec.containers[].readinessProbe.grpc.service

`string` · optional (explicit presence)

`service` is the name of the service to place in the gRPC
HealthCheckRequest (see
https://github.com/grpc/grpc/blob/master/doc/health-checking.md).

If this is not specified, the default behavior is defined by gRPC.

### spec.podTemplate.spec.containers[].readinessProbe.grpc.mode

`string` · optional (explicit presence)

mode specifies the connection mode for the gRPC health probe. Set to "TLS"
to use TLS without certificate verification. Set to "Plaintext" to use a
plaintext (insecure) connection explicitly. If not specified, the probe
uses a plaintext (insecure) connection.

Requires the GRPCContainerProbeTLS feature gate on the cluster; a cluster
without it drops this field from the pod.

- rule: {"string":{"in":["Plaintext","TLS"]}}

### spec.podTemplate.spec.containers[].readinessProbe.initialDelaySeconds

`int32`

Number of seconds after the container has started before liveness probes
are initiated. More info:
https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes

### spec.podTemplate.spec.containers[].readinessProbe.timeoutSeconds

`int32`

Number of seconds after which the probe times out. Defaults to 1 second.
Minimum value is 1. More info:
https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes

### spec.podTemplate.spec.containers[].readinessProbe.periodSeconds

`int32`

How often (in seconds) to perform the probe. Default to 10 seconds. Minimum
value is 1.

### spec.podTemplate.spec.containers[].readinessProbe.successThreshold

`int32`

Minimum consecutive successes for the probe to be considered successful
after having failed. Defaults to 1. Must be 1 for liveness and startup.
Minimum value is 1.

### spec.podTemplate.spec.containers[].readinessProbe.failureThreshold

`int32`

Minimum consecutive failures for the probe to be considered failed after
having succeeded. Defaults to 3. Minimum value is 1.

### spec.podTemplate.spec.containers[].readinessProbe.terminationGracePeriodSeconds

`int64` · optional (explicit presence)

Optional duration in seconds the pod needs to terminate gracefully upon
probe failure. The grace period is the duration in seconds after the
processes running in the pod are sent a termination signal and the time
when the processes are forcibly halted with a kill signal. Set this value
longer than the expected cleanup time for your process. If this value is
nil, the pod's terminationGracePeriodSeconds will be used. Otherwise, this
value overrides the value provided by the pod spec. Value must be
non-negative integer. The value zero indicates stop immediately via the
kill signal (no opportunity to shut down). This is a beta field and
requires enabling ProbeTerminationGracePeriod feature gate. Minimum value
is 1. spec.terminationGracePeriodSeconds is used if unset.

### spec.podTemplate.spec.containers[].startupProbe

`KubernetesCoreV1ApiProbe`

`startupProbe` indicates that the Pod has successfully initialized. If
specified, no other probes are executed until this completes successfully.
If this probe fails, the Pod will be restarted, just as if the
livenessProbe failed. This can be used to provide different probe
parameters at the beginning of a Pod's lifecycle, when it might take a long
time to load data or warm a cache, than during steady-state operation. This
cannot be updated. More info:
https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes

- rule: Set at most one of exec, httpGet, tcpSocket, grpc

### spec.podTemplate.spec.containers[].startupProbe.exec

`KubernetesCoreV1ApiExecAction`

`exec` specifies a command to execute in the container.

### spec.podTemplate.spec.containers[].startupProbe.exec.command

`[]string`

`command` is the command line to execute inside the container, the working
directory for the command  is root ('/') in the container's filesystem. The
command is simply exec'd, it is not run inside a shell, so traditional
shell instructions ('|', etc) won't work. To use a shell, you need to
explicitly call out to that shell. Exit status of 0 is treated as
live/healthy and non-zero is unhealthy.

### spec.podTemplate.spec.containers[].startupProbe.httpGet

`KubernetesCoreV1ApiHTTPGetAction`

`httpGet` specifies an HTTP GET request to perform.

### spec.podTemplate.spec.containers[].startupProbe.httpGet.path

`string`

`path` to access on the HTTP server.

### spec.podTemplate.spec.containers[].startupProbe.httpGet.port

`string` · required

Name or number of the port to access on the container. Number must be in
the range 1 to 65535. Name must be an IANA_SVC_NAME.

- rule: port is a port number from 1 to 65535, or a port name (up to 15 lowercase letters, digits and hyphens, with at least one letter)
- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.containers[].startupProbe.httpGet.host

`string`

`host` name to connect to, defaults to the pod IP. You probably want to set
"Host" in httpHeaders instead.

### spec.podTemplate.spec.containers[].startupProbe.httpGet.scheme

`string` · optional (explicit presence)

`scheme` to use for connecting to the host. Defaults to HTTP.

- rule: {"string":{"in":["HTTP","HTTPS"]}}

### spec.podTemplate.spec.containers[].startupProbe.httpGet.httpHeaders

`[]KubernetesCoreV1ApiHTTPHeader`

Custom headers to set in the request. HTTP allows repeated headers.

### spec.podTemplate.spec.containers[].startupProbe.httpGet.httpHeaders[].name

`string` · required

The header field name. This will be canonicalized upon output, so
case-variant names will be understood as the same header.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.containers[].startupProbe.httpGet.httpHeaders[].value

`string` · required

The header field value.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.containers[].startupProbe.httpGet.protocol

`string` · optional (explicit presence)

`protocol` selects the wire protocol for the probe connection. Nil defaults
to HTTP/1.1.

Requires the H2CContainerProbe feature gate on the cluster; a cluster
without it drops this field from the pod.

- rule: {"string":{"in":["HTTP1","HTTP2"]}}

### spec.podTemplate.spec.containers[].startupProbe.tcpSocket

`KubernetesCoreV1ApiTCPSocketAction`

`tcpSocket` specifies a connection to a TCP port.

### spec.podTemplate.spec.containers[].startupProbe.tcpSocket.port

`string` · required

Number or name of the port to access on the container. Number must be in
the range 1 to 65535. Name must be an IANA_SVC_NAME.

- rule: port is a port number from 1 to 65535, or a port name (up to 15 lowercase letters, digits and hyphens, with at least one letter)
- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.containers[].startupProbe.tcpSocket.host

`string`

Optional: Host name to connect to, defaults to the pod IP.

### spec.podTemplate.spec.containers[].startupProbe.grpc

`KubernetesCoreV1ApiGRPCAction`

`grpc` specifies a GRPC HealthCheckRequest.

### spec.podTemplate.spec.containers[].startupProbe.grpc.port

`int32` · required · optional (explicit presence)

`port` number of the gRPC service. Number must be in the range 1 to 65535.

- rule: {"required":true}

### spec.podTemplate.spec.containers[].startupProbe.grpc.service

`string` · optional (explicit presence)

`service` is the name of the service to place in the gRPC
HealthCheckRequest (see
https://github.com/grpc/grpc/blob/master/doc/health-checking.md).

If this is not specified, the default behavior is defined by gRPC.

### spec.podTemplate.spec.containers[].startupProbe.grpc.mode

`string` · optional (explicit presence)

mode specifies the connection mode for the gRPC health probe. Set to "TLS"
to use TLS without certificate verification. Set to "Plaintext" to use a
plaintext (insecure) connection explicitly. If not specified, the probe
uses a plaintext (insecure) connection.

Requires the GRPCContainerProbeTLS feature gate on the cluster; a cluster
without it drops this field from the pod.

- rule: {"string":{"in":["Plaintext","TLS"]}}

### spec.podTemplate.spec.containers[].startupProbe.initialDelaySeconds

`int32`

Number of seconds after the container has started before liveness probes
are initiated. More info:
https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes

### spec.podTemplate.spec.containers[].startupProbe.timeoutSeconds

`int32`

Number of seconds after which the probe times out. Defaults to 1 second.
Minimum value is 1. More info:
https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes

### spec.podTemplate.spec.containers[].startupProbe.periodSeconds

`int32`

How often (in seconds) to perform the probe. Default to 10 seconds. Minimum
value is 1.

### spec.podTemplate.spec.containers[].startupProbe.successThreshold

`int32`

Minimum consecutive successes for the probe to be considered successful
after having failed. Defaults to 1. Must be 1 for liveness and startup.
Minimum value is 1.

### spec.podTemplate.spec.containers[].startupProbe.failureThreshold

`int32`

Minimum consecutive failures for the probe to be considered failed after
having succeeded. Defaults to 3. Minimum value is 1.

### spec.podTemplate.spec.containers[].startupProbe.terminationGracePeriodSeconds

`int64` · optional (explicit presence)

Optional duration in seconds the pod needs to terminate gracefully upon
probe failure. The grace period is the duration in seconds after the
processes running in the pod are sent a termination signal and the time
when the processes are forcibly halted with a kill signal. Set this value
longer than the expected cleanup time for your process. If this value is
nil, the pod's terminationGracePeriodSeconds will be used. Otherwise, this
value overrides the value provided by the pod spec. Value must be
non-negative integer. The value zero indicates stop immediately via the
kill signal (no opportunity to shut down). This is a beta field and
requires enabling ProbeTerminationGracePeriod feature gate. Minimum value
is 1. spec.terminationGracePeriodSeconds is used if unset.

### spec.podTemplate.spec.containers[].lifecycle

`KubernetesCoreV1ApiLifecycle`

Actions that the management system should take in response to container
lifecycle events. Cannot be updated.

### spec.podTemplate.spec.containers[].lifecycle.postStart

`KubernetesCoreV1ApiLifecycleHandler`

`postStart` is called immediately after a container is created. If the
handler fails, the container is terminated and restarted according to its
restart policy. Other management of the container blocks until the hook
completes. More info:
https://kubernetes.io/docs/concepts/containers/container-lifecycle-hooks/#container-hooks

- rule: Set exactly one of exec, httpGet, sleep, tcpSocket

### spec.podTemplate.spec.containers[].lifecycle.postStart.exec

`KubernetesCoreV1ApiExecAction`

`exec` specifies a command to execute in the container.

### spec.podTemplate.spec.containers[].lifecycle.postStart.exec.command

`[]string`

`command` is the command line to execute inside the container, the working
directory for the command  is root ('/') in the container's filesystem. The
command is simply exec'd, it is not run inside a shell, so traditional
shell instructions ('|', etc) won't work. To use a shell, you need to
explicitly call out to that shell. Exit status of 0 is treated as
live/healthy and non-zero is unhealthy.

### spec.podTemplate.spec.containers[].lifecycle.postStart.httpGet

`KubernetesCoreV1ApiHTTPGetAction`

`httpGet` specifies an HTTP GET request to perform.

### spec.podTemplate.spec.containers[].lifecycle.postStart.httpGet.path

`string`

`path` to access on the HTTP server.

### spec.podTemplate.spec.containers[].lifecycle.postStart.httpGet.port

`string` · required

Name or number of the port to access on the container. Number must be in
the range 1 to 65535. Name must be an IANA_SVC_NAME.

- rule: port is a port number from 1 to 65535, or a port name (up to 15 lowercase letters, digits and hyphens, with at least one letter)
- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.containers[].lifecycle.postStart.httpGet.host

`string`

`host` name to connect to, defaults to the pod IP. You probably want to set
"Host" in httpHeaders instead.

### spec.podTemplate.spec.containers[].lifecycle.postStart.httpGet.scheme

`string` · optional (explicit presence)

`scheme` to use for connecting to the host. Defaults to HTTP.

- rule: {"string":{"in":["HTTP","HTTPS"]}}

### spec.podTemplate.spec.containers[].lifecycle.postStart.httpGet.httpHeaders

`[]KubernetesCoreV1ApiHTTPHeader`

Custom headers to set in the request. HTTP allows repeated headers.

### spec.podTemplate.spec.containers[].lifecycle.postStart.httpGet.httpHeaders[].name

`string` · required

The header field name. This will be canonicalized upon output, so
case-variant names will be understood as the same header.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.containers[].lifecycle.postStart.httpGet.httpHeaders[].value

`string` · required

The header field value.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.containers[].lifecycle.postStart.httpGet.protocol

`string` · optional (explicit presence)

`protocol` selects the wire protocol for the probe connection. Nil defaults
to HTTP/1.1.

Requires the H2CContainerProbe feature gate on the cluster; a cluster
without it drops this field from the pod.

- rule: {"string":{"in":["HTTP1","HTTP2"]}}

### spec.podTemplate.spec.containers[].lifecycle.postStart.tcpSocket

`KubernetesCoreV1ApiTCPSocketAction`

Deprecated. TCPSocket is NOT supported as a LifecycleHandler and kept for
backward compatibility. There is no validation of this field and lifecycle
hooks will fail at runtime when it is specified.

### spec.podTemplate.spec.containers[].lifecycle.postStart.tcpSocket.port

`string` · required

Number or name of the port to access on the container. Number must be in
the range 1 to 65535. Name must be an IANA_SVC_NAME.

- rule: port is a port number from 1 to 65535, or a port name (up to 15 lowercase letters, digits and hyphens, with at least one letter)
- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.containers[].lifecycle.postStart.tcpSocket.host

`string`

Optional: Host name to connect to, defaults to the pod IP.

### spec.podTemplate.spec.containers[].lifecycle.postStart.sleep

`KubernetesCoreV1ApiSleepAction`

`sleep` represents a duration that the container should sleep.

### spec.podTemplate.spec.containers[].lifecycle.postStart.sleep.seconds

`int64` · required · optional (explicit presence)

`seconds` is the number of seconds to sleep.

- rule: {"required":true}

### spec.podTemplate.spec.containers[].lifecycle.preStop

`KubernetesCoreV1ApiLifecycleHandler`

`preStop` is called immediately before a container is terminated due to an
API request or management event such as liveness/startup probe failure,
preemption, resource contention, etc. The handler is not called if the
container crashes or exits. The Pod's termination grace period countdown
begins before the PreStop hook is executed. Regardless of the outcome of
the handler, the container will eventually terminate within the Pod's
termination grace period (unless delayed by finalizers). Other management
of the container blocks until the hook completes or until the termination
grace period is reached. More info:
https://kubernetes.io/docs/concepts/containers/container-lifecycle-hooks/#container-hooks

- rule: Set exactly one of exec, httpGet, sleep, tcpSocket

### spec.podTemplate.spec.containers[].lifecycle.preStop.exec

`KubernetesCoreV1ApiExecAction`

`exec` specifies a command to execute in the container.

### spec.podTemplate.spec.containers[].lifecycle.preStop.exec.command

`[]string`

`command` is the command line to execute inside the container, the working
directory for the command  is root ('/') in the container's filesystem. The
command is simply exec'd, it is not run inside a shell, so traditional
shell instructions ('|', etc) won't work. To use a shell, you need to
explicitly call out to that shell. Exit status of 0 is treated as
live/healthy and non-zero is unhealthy.

### spec.podTemplate.spec.containers[].lifecycle.preStop.httpGet

`KubernetesCoreV1ApiHTTPGetAction`

`httpGet` specifies an HTTP GET request to perform.

### spec.podTemplate.spec.containers[].lifecycle.preStop.httpGet.path

`string`

`path` to access on the HTTP server.

### spec.podTemplate.spec.containers[].lifecycle.preStop.httpGet.port

`string` · required

Name or number of the port to access on the container. Number must be in
the range 1 to 65535. Name must be an IANA_SVC_NAME.

- rule: port is a port number from 1 to 65535, or a port name (up to 15 lowercase letters, digits and hyphens, with at least one letter)
- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.containers[].lifecycle.preStop.httpGet.host

`string`

`host` name to connect to, defaults to the pod IP. You probably want to set
"Host" in httpHeaders instead.

### spec.podTemplate.spec.containers[].lifecycle.preStop.httpGet.scheme

`string` · optional (explicit presence)

`scheme` to use for connecting to the host. Defaults to HTTP.

- rule: {"string":{"in":["HTTP","HTTPS"]}}

### spec.podTemplate.spec.containers[].lifecycle.preStop.httpGet.httpHeaders

`[]KubernetesCoreV1ApiHTTPHeader`

Custom headers to set in the request. HTTP allows repeated headers.

### spec.podTemplate.spec.containers[].lifecycle.preStop.httpGet.httpHeaders[].name

`string` · required

The header field name. This will be canonicalized upon output, so
case-variant names will be understood as the same header.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.containers[].lifecycle.preStop.httpGet.httpHeaders[].value

`string` · required

The header field value.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.containers[].lifecycle.preStop.httpGet.protocol

`string` · optional (explicit presence)

`protocol` selects the wire protocol for the probe connection. Nil defaults
to HTTP/1.1.

Requires the H2CContainerProbe feature gate on the cluster; a cluster
without it drops this field from the pod.

- rule: {"string":{"in":["HTTP1","HTTP2"]}}

### spec.podTemplate.spec.containers[].lifecycle.preStop.tcpSocket

`KubernetesCoreV1ApiTCPSocketAction`

Deprecated. TCPSocket is NOT supported as a LifecycleHandler and kept for
backward compatibility. There is no validation of this field and lifecycle
hooks will fail at runtime when it is specified.

### spec.podTemplate.spec.containers[].lifecycle.preStop.tcpSocket.port

`string` · required

Number or name of the port to access on the container. Number must be in
the range 1 to 65535. Name must be an IANA_SVC_NAME.

- rule: port is a port number from 1 to 65535, or a port name (up to 15 lowercase letters, digits and hyphens, with at least one letter)
- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.containers[].lifecycle.preStop.tcpSocket.host

`string`

Optional: Host name to connect to, defaults to the pod IP.

### spec.podTemplate.spec.containers[].lifecycle.preStop.sleep

`KubernetesCoreV1ApiSleepAction`

`sleep` represents a duration that the container should sleep.

### spec.podTemplate.spec.containers[].lifecycle.preStop.sleep.seconds

`int64` · required · optional (explicit presence)

`seconds` is the number of seconds to sleep.

- rule: {"required":true}

### spec.podTemplate.spec.containers[].lifecycle.stopSignal

`string` · optional (explicit presence)

`stopSignal` defines which signal will be sent to a container when it is
being stopped. If not specified, the default is defined by the container
runtime in use. StopSignal can only be set for Pods with a non-empty
.spec.os.name.

- rule: {"string":{"in":["SIGABRT","SIGALRM","SIGBUS","SIGCHLD","SIGCLD","SIGCONT","SIGFPE","SIGHUP","SIGILL","SIGINT","SIGIO","SIGIOT","SIGKILL","SIGPIPE","SIGPOLL","SIGPROF","SIGPWR","SIGQUIT","SIGSEGV","SIGSTKFLT","SIGSTOP","SIGSYS","SIGTERM","SIGTRAP","SIGTSTP","SIGTTIN","SIGTTOU","SIGURG","SIGUSR1","SIGUSR2","SIGVTALRM","SIGWINCH","SIGXCPU","SIGXFSZ","SIGRTMIN","SIGRTMIN+1","SIGRTMIN+2","SIGRTMIN+3","SIGRTMIN+4","SIGRTMIN+5","SIGRTMIN+6","SIGRTMIN+7","SIGRTMIN+8","SIGRTMIN+9","SIGRTMIN+10","SIGRTMIN+11","SIGRTMIN+12","SIGRTMIN+13","SIGRTMIN+14","SIGRTMIN+15","SIGRTMAX-14","SIGRTMAX-13","SIGRTMAX-12","SIGRTMAX-11","SIGRTMAX-10","SIGRTMAX-9","SIGRTMAX-8","SIGRTMAX-7","SIGRTMAX-6","SIGRTMAX-5","SIGRTMAX-4","SIGRTMAX-3","SIGRTMAX-2","SIGRTMAX-1","SIGRTMAX"]}}

### spec.podTemplate.spec.containers[].terminationMessagePath

`string`

Optional: Path at which the file to which the container's termination
message will be written is mounted into the container's filesystem. Message
written is intended to be brief final status, such as an assertion failure
message. Will be truncated by the node if greater than 4096 bytes. The
total message length across all containers will be limited to 12kb.
Defaults to /dev/termination-log. Cannot be updated.

### spec.podTemplate.spec.containers[].terminationMessagePolicy

`string` · optional (explicit presence)

Indicate how the termination message should be populated. File will use the
contents of terminationMessagePath to populate the container status message
on both success and failure. FallbackToLogsOnError will use the last chunk
of container log output if the termination message file is empty and the
container exited with an error. The log output is limited to 2048 bytes or
80 lines, whichever is smaller. Defaults to File. Cannot be updated.

- rule: {"string":{"in":["File","FallbackToLogsOnError"]}}

### spec.podTemplate.spec.containers[].imagePullPolicy

`string` · optional (explicit presence)

Image pull policy. One of Always, Never, IfNotPresent. Defaults to Always
if :latest tag is specified, or IfNotPresent otherwise. Cannot be updated.
More info:
https://kubernetes.io/docs/concepts/containers/images#updating-images

- rule: {"string":{"in":["Always","Never","IfNotPresent"]}}

### spec.podTemplate.spec.containers[].securityContext

`KubernetesCoreV1ApiSecurityContext`

`securityContext` defines the security options the container should be run
with. If set, the fields of SecurityContext override the equivalent fields
of PodSecurityContext. More info:
https://kubernetes.io/docs/tasks/configure-pod-container/security-context/

### spec.podTemplate.spec.containers[].securityContext.capabilities

`KubernetesCoreV1ApiCapabilities`

The capabilities to add/drop when running containers. Defaults to the
default set of capabilities granted by the container runtime. Note that
this field cannot be set when spec.os.name is windows.

### spec.podTemplate.spec.containers[].securityContext.capabilities.add

`[]string`

Added capabilities.

### spec.podTemplate.spec.containers[].securityContext.capabilities.drop

`[]string`

Removed capabilities.

### spec.podTemplate.spec.containers[].securityContext.privileged

`bool` · optional (explicit presence)

Run container in privileged mode. Processes in privileged containers are
essentially equivalent to root on the host. Defaults to false. Note that
this field cannot be set when spec.os.name is windows.

### spec.podTemplate.spec.containers[].securityContext.seLinuxOptions

`KubernetesCoreV1ApiSELinuxOptions`

The SELinux context to be applied to the container. If unspecified, the
container runtime will allocate a random SELinux context for each
container.  May also be set in PodSecurityContext.  If set in both
SecurityContext and PodSecurityContext, the value specified in
SecurityContext takes precedence. Note that this field cannot be set when
spec.os.name is windows.

### spec.podTemplate.spec.containers[].securityContext.seLinuxOptions.user

`string`

`user` is a SELinux user label that applies to the container.

### spec.podTemplate.spec.containers[].securityContext.seLinuxOptions.role

`string`

`role` is a SELinux role label that applies to the container.

### spec.podTemplate.spec.containers[].securityContext.seLinuxOptions.type

`string`

`type` is a SELinux type label that applies to the container.

### spec.podTemplate.spec.containers[].securityContext.seLinuxOptions.level

`string`

`level` is SELinux level label that applies to the container.

### spec.podTemplate.spec.containers[].securityContext.windowsOptions

`KubernetesCoreV1ApiWindowsSecurityContextOptions`

The Windows specific settings applied to all containers. If unspecified,
the options from the PodSecurityContext will be used. If set in both
SecurityContext and PodSecurityContext, the value specified in
SecurityContext takes precedence. Note that this field cannot be set when
spec.os.name is linux.

### spec.podTemplate.spec.containers[].securityContext.windowsOptions.gmsaCredentialSpecName

`string` · optional (explicit presence)

`gmsaCredentialSpecName` is the name of the GMSA credential spec to use.

### spec.podTemplate.spec.containers[].securityContext.windowsOptions.gmsaCredentialSpec

`string` · optional (explicit presence)

`gmsaCredentialSpec` is where the GMSA admission webhook
(https://github.com/kubernetes-sigs/windows-gmsa) inlines the contents of
the GMSA credential spec named by the GMSACredentialSpecName field.

### spec.podTemplate.spec.containers[].securityContext.windowsOptions.runAsUserName

`string` · optional (explicit presence)

The UserName in Windows to run the entrypoint of the container process.
Defaults to the user specified in image metadata if unspecified. May also
be set in PodSecurityContext. If set in both SecurityContext and
PodSecurityContext, the value specified in SecurityContext takes
precedence.

### spec.podTemplate.spec.containers[].securityContext.windowsOptions.hostProcess

`bool` · optional (explicit presence)

`hostProcess` determines if a container should be run as a 'Host Process'
container. All of a Pod's containers must have the same effective
HostProcess value (it is not allowed to have a mix of HostProcess
containers and non-HostProcess containers). In addition, if HostProcess is
true then HostNetwork must also be set to true.

### spec.podTemplate.spec.containers[].securityContext.runAsUser

`int64` · optional (explicit presence)

The UID to run the entrypoint of the container process. Defaults to user
specified in image metadata if unspecified. May also be set in
PodSecurityContext.  If set in both SecurityContext and PodSecurityContext,
the value specified in SecurityContext takes precedence. Note that this
field cannot be set when spec.os.name is windows.

### spec.podTemplate.spec.containers[].securityContext.runAsGroup

`int64` · optional (explicit presence)

The GID to run the entrypoint of the container process. Uses runtime
default if unset. May also be set in PodSecurityContext.  If set in both
SecurityContext and PodSecurityContext, the value specified in
SecurityContext takes precedence. Note that this field cannot be set when
spec.os.name is windows.

### spec.podTemplate.spec.containers[].securityContext.runAsNonRoot

`bool` · optional (explicit presence)

Indicates that the container must run as a non-root user. If true, the
Kubelet will validate the image at runtime to ensure that it does not run
as UID 0 (root) and fail to start the container if it does. If unset or
false, no such validation will be performed. May also be set in
PodSecurityContext.  If set in both SecurityContext and PodSecurityContext,
the value specified in SecurityContext takes precedence.

### spec.podTemplate.spec.containers[].securityContext.readOnlyRootFilesystem

`bool` · optional (explicit presence)

Whether this container has a read-only root filesystem. Default is false.
Note that this field cannot be set when spec.os.name is windows.

### spec.podTemplate.spec.containers[].securityContext.allowPrivilegeEscalation

`bool` · optional (explicit presence)

`allowPrivilegeEscalation` controls whether a process can gain more
privileges than its parent process. This bool directly controls if the
no_new_privs flag will be set on the container process.
AllowPrivilegeEscalation is true always when the container is: 1) run as
Privileged 2) has CAP_SYS_ADMIN Note that this field cannot be set when
spec.os.name is windows.

### spec.podTemplate.spec.containers[].securityContext.procMount

`string` · optional (explicit presence)

procMount denotes the type of proc mount to use for the containers. The
default value is Default which uses the container runtime defaults for
readonly paths and masked paths. Note that this field cannot be set when
spec.os.name is windows.

- rule: {"string":{"in":["Default","Unmasked"]}}

### spec.podTemplate.spec.containers[].securityContext.seccompProfile

`KubernetesCoreV1ApiSeccompProfile`

The seccomp options to use by this container. If seccomp options are
provided at both the pod & container level, the container options override
the pod options. Note that this field cannot be set when spec.os.name is
windows.

### spec.podTemplate.spec.containers[].securityContext.seccompProfile.type

`string`

type indicates which kind of seccomp profile will be applied. Valid options
are:

Localhost - a profile defined in a file on the node should be used.
RuntimeDefault - the container runtime default profile should be used.
Unconfined - no profile should be applied.

- rule: {"string":{"in":["Unconfined","RuntimeDefault","Localhost"]}}

### spec.podTemplate.spec.containers[].securityContext.seccompProfile.localhostProfile

`string` · optional (explicit presence)

localhostProfile indicates a profile defined in a file on the node should
be used. The profile must be preconfigured on the node to work. Must be a
descending path, relative to the kubelet's configured seccomp profile
location. Must be set if type is "Localhost". Must NOT be set for any other
type.

### spec.podTemplate.spec.containers[].securityContext.appArmorProfile

`KubernetesCoreV1ApiAppArmorProfile`

appArmorProfile is the AppArmor options to use by this container. If set,
this profile overrides the pod's appArmorProfile. Note that this field
cannot be set when spec.os.name is windows.

### spec.podTemplate.spec.containers[].securityContext.appArmorProfile.type

`string`

type indicates which kind of AppArmor profile will be applied. Valid
options are: Localhost - a profile pre-loaded on the node. RuntimeDefault -
the container runtime's default profile. Unconfined - no AppArmor
enforcement.

- rule: {"string":{"in":["Unconfined","RuntimeDefault","Localhost"]}}

### spec.podTemplate.spec.containers[].securityContext.appArmorProfile.localhostProfile

`string` · optional (explicit presence)

localhostProfile indicates a profile loaded on the node that should be
used. The profile must be preconfigured on the node to work. Must match the
loaded name of the profile. Must be set if and only if type is "Localhost".

### spec.podTemplate.spec.containers[].stdin

`bool`

Whether this container should allocate a buffer for stdin in the container
runtime. If this is not set, reads from stdin in the container will always
result in EOF. Default is false.

### spec.podTemplate.spec.containers[].stdinOnce

`bool`

Whether the container runtime should close the stdin channel after it has
been opened by a single attach. When stdin is true the stdin stream will
remain open across multiple attach sessions. If stdinOnce is set to true,
stdin is opened on container start, is empty until the first client
attaches to stdin, and then remains open and accepts data until the client
disconnects, at which time stdin is closed and remains closed until the
container is restarted. If this flag is false, a container processes that
reads from stdin will never receive an EOF. Default is false.

### spec.podTemplate.spec.containers[].tty

`bool`

Whether this container should allocate a TTY for itself, also requires
'stdin' to be true. Default is false.

### spec.podTemplate.spec.ephemeralContainers

`[]KubernetesCoreV1ApiEphemeralContainer`

List of ephemeral containers run in this pod. Ephemeral containers may be
run in an existing pod to perform user-initiated actions such as debugging.
This list cannot be specified when creating a pod, and it cannot be
modified by updating the pod spec. In order to add an ephemeral container
to an existing pod, use the pod's ephemeralcontainers subresource.

### spec.podTemplate.spec.ephemeralContainers[].name

`string` · required

`name` of the ephemeral container specified as a DNS_LABEL. This name must
be unique among all containers, init containers and ephemeral containers.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.ephemeralContainers[].image

`string`

Container image name. More info:
https://kubernetes.io/docs/concepts/containers/images

### spec.podTemplate.spec.ephemeralContainers[].command

`[]string`

Entrypoint array. Not executed within a shell. The image's ENTRYPOINT is
used if this is not provided. Variable references $(VAR_NAME) are expanded
using the container's environment. If a variable cannot be resolved, the
reference in the input string will be unchanged. Double $$ are reduced to a
single $, which allows for escaping the $(VAR_NAME) syntax: i.e.
"$$(VAR_NAME)" will produce the string literal "$(VAR_NAME)". Escaped
references will never be expanded, regardless of whether the variable
exists or not. Cannot be updated. More info:
https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell

### spec.podTemplate.spec.ephemeralContainers[].args

`[]string`

Arguments to the entrypoint. The image's CMD is used if this is not
provided. Variable references $(VAR_NAME) are expanded using the
container's environment. If a variable cannot be resolved, the reference in
the input string will be unchanged. Double $$ are reduced to a single $,
which allows for escaping the $(VAR_NAME) syntax: i.e. "$$(VAR_NAME)" will
produce the string literal "$(VAR_NAME)". Escaped references will never be
expanded, regardless of whether the variable exists or not. Cannot be
updated. More info:
https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell

### spec.podTemplate.spec.ephemeralContainers[].workingDir

`string`

Container's working directory. If not specified, the container runtime's
default will be used, which might be configured in the container image.
Cannot be updated.

### spec.podTemplate.spec.ephemeralContainers[].ports

`[]KubernetesCoreV1ApiContainerPort`

`ports` are not allowed for ephemeral containers.

### spec.podTemplate.spec.ephemeralContainers[].ports[].name

`string`

If specified, this must be an IANA_SVC_NAME and unique within the pod. Each
named port in a pod must have a unique name. Name for the port that can be
referred to by services.

### spec.podTemplate.spec.ephemeralContainers[].ports[].hostPort

`int32`

Number of port to expose on the host. If specified, this must be a valid
port number, 0 < x < 65536. If HostNetwork is specified, this must match
ContainerPort. Most containers do not need this.

### spec.podTemplate.spec.ephemeralContainers[].ports[].containerPort

`int32` · required · optional (explicit presence)

Number of port to expose on the pod's IP address. This must be a valid port
number, 0 < x < 65536.

- rule: {"required":true}

### spec.podTemplate.spec.ephemeralContainers[].ports[].protocol

`string` · optional (explicit presence)

`protocol` for port. Must be UDP, TCP, or SCTP. Defaults to "TCP".

Unset, the cluster writes "TCP".

- rule: {"string":{"in":["TCP","UDP","SCTP"]}}

### spec.podTemplate.spec.ephemeralContainers[].ports[].hostIP

`string`

What host IP to bind the external port to.

### spec.podTemplate.spec.ephemeralContainers[].envFrom

`[]KubernetesCoreV1ApiEnvFromSource`

List of sources to populate environment variables in the container. The
keys defined within a source may consist of any printable ASCII characters
except '='. When a key exists in multiple sources, the value associated
with the last source will take precedence. Values defined by an Env with a
duplicate key will take precedence. Cannot be updated.

- rule: Set exactly one of configMapRef, secretRef

### spec.podTemplate.spec.ephemeralContainers[].envFrom[].prefix

`string`

Optional text to prepend to the name of each environment variable. May
consist of any printable ASCII characters except '='.

### spec.podTemplate.spec.ephemeralContainers[].envFrom[].configMapRef

`KubernetesCoreV1ApiConfigMapEnvSource`

The ConfigMap to select from.

### spec.podTemplate.spec.ephemeralContainers[].envFrom[].configMapRef.name

`string | valueFrom` · required

`name` of the referent. This field is effectively required, but due to
backwards compatibility is allowed to be empty. Instances of this type with
an empty value here are almost certainly wrong. More info:
https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names

Defaults to a KubernetesConfigMap reference (its configmap_name output);
pass the literal name with `value:` for one Planton does not manage.

- references: KubernetesConfigMap (`status.outputs.configmap_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesConfigMap, name: <that resource's name>, fieldPath: status.outputs.configmap_name}} -- a bare string does not parse

### spec.podTemplate.spec.ephemeralContainers[].envFrom[].configMapRef.optional

`bool` · optional (explicit presence)

Specify whether the ConfigMap must be defined.

### spec.podTemplate.spec.ephemeralContainers[].envFrom[].secretRef

`KubernetesCoreV1ApiSecretEnvSource`

The Secret to select from.

### spec.podTemplate.spec.ephemeralContainers[].envFrom[].secretRef.name

`string | valueFrom` · required

`name` of the referent. This field is effectively required, but due to
backwards compatibility is allowed to be empty. Instances of this type with
an empty value here are almost certainly wrong. More info:
https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names

Defaults to a KubernetesSecret reference (its secret_name output); pass the
literal name with `value:` for one Planton does not manage.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.podTemplate.spec.ephemeralContainers[].envFrom[].secretRef.optional

`bool` · optional (explicit presence)

Specify whether the Secret must be defined.

### spec.podTemplate.spec.ephemeralContainers[].env

`[]KubernetesCoreV1ApiEnvVar`

List of environment variables to set in the container. Cannot be updated.

### spec.podTemplate.spec.ephemeralContainers[].env[].name

`string` · required

`name` of the environment variable. May consist of any printable ASCII
characters except '='.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.ephemeralContainers[].env[].value

`string`

Variable references $(VAR_NAME) are expanded using the previously defined
environment variables in the container and any service environment
variables. If a variable cannot be resolved, the reference in the input
string will be unchanged. Double $$ are reduced to a single $, which allows
for escaping the $(VAR_NAME) syntax: i.e. "$$(VAR_NAME)" will produce the
string literal "$(VAR_NAME)". Escaped references will never be expanded,
regardless of whether the variable exists or not. Defaults to "".

### spec.podTemplate.spec.ephemeralContainers[].env[].valueFrom

`KubernetesCoreV1ApiEnvVarSource`

Source for the environment variable's value. Cannot be used if value is not
empty.

- rule: Set exactly one of configMapKeyRef, fieldRef, fileKeyRef, resourceFieldRef, secretKeyRef

### spec.podTemplate.spec.ephemeralContainers[].env[].valueFrom.fieldRef

`KubernetesCoreV1ApiObjectFieldSelector`

Selects a field of the pod: supports metadata.name, metadata.namespace,
`metadata.labels['<KEY>']`, `metadata.annotations['<KEY>']`, spec.nodeName,
spec.serviceAccountName, status.hostIP, status.podIP, status.podIPs.

### spec.podTemplate.spec.ephemeralContainers[].env[].valueFrom.fieldRef.apiVersion

`string`

Version of the schema the FieldPath is written in terms of, defaults to
"v1".

### spec.podTemplate.spec.ephemeralContainers[].env[].valueFrom.fieldRef.fieldPath

`string` · required

Path of the field to select in the specified API version.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.ephemeralContainers[].env[].valueFrom.resourceFieldRef

`KubernetesCoreV1ApiResourceFieldSelector`

Selects a resource of the container: only resources limits and requests
(limits.cpu, limits.memory, limits.ephemeral-storage, requests.cpu,
requests.memory and requests.ephemeral-storage) are currently supported.

### spec.podTemplate.spec.ephemeralContainers[].env[].valueFrom.resourceFieldRef.containerName

`string`

Container name: required for volumes, optional for env vars.

### spec.podTemplate.spec.ephemeralContainers[].env[].valueFrom.resourceFieldRef.resource

`string` · required

Required: resource to select.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.ephemeralContainers[].env[].valueFrom.resourceFieldRef.divisor

`string` · optional (explicit presence)

Specifies the output format of the exposed resources, defaults to "1"

- rule: {"string":{"pattern":"^(\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))))?$"}}

### spec.podTemplate.spec.ephemeralContainers[].env[].valueFrom.configMapKeyRef

`KubernetesCoreV1ApiConfigMapKeySelector`

Selects a key of a ConfigMap.

### spec.podTemplate.spec.ephemeralContainers[].env[].valueFrom.configMapKeyRef.name

`string | valueFrom` · required

`name` of the referent. This field is effectively required, but due to
backwards compatibility is allowed to be empty. Instances of this type with
an empty value here are almost certainly wrong. More info:
https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names

Defaults to a KubernetesConfigMap reference (its configmap_name output);
pass the literal name with `value:` for one Planton does not manage.

- references: KubernetesConfigMap (`status.outputs.configmap_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesConfigMap, name: <that resource's name>, fieldPath: status.outputs.configmap_name}} -- a bare string does not parse

### spec.podTemplate.spec.ephemeralContainers[].env[].valueFrom.configMapKeyRef.key

`string` · required

The key to select from the ConfigMap's Data field. Keys in the BinaryData
field are not currently propagated to container env vars.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.ephemeralContainers[].env[].valueFrom.configMapKeyRef.optional

`bool` · optional (explicit presence)

Specify whether the ConfigMap or its key must be defined.

### spec.podTemplate.spec.ephemeralContainers[].env[].valueFrom.secretKeyRef

`KubernetesCoreV1ApiSecretKeySelector`

Selects a key of a secret in the pod's namespace.

### spec.podTemplate.spec.ephemeralContainers[].env[].valueFrom.secretKeyRef.name

`string | valueFrom` · required

`name` of the referent. This field is effectively required, but due to
backwards compatibility is allowed to be empty. Instances of this type with
an empty value here are almost certainly wrong. More info:
https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names

Defaults to a KubernetesSecret reference (its secret_name output); pass the
literal name with `value:` for one Planton does not manage.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.podTemplate.spec.ephemeralContainers[].env[].valueFrom.secretKeyRef.key

`string` · required

The key of the secret to select from.  Must be a valid secret key.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.ephemeralContainers[].env[].valueFrom.secretKeyRef.optional

`bool` · optional (explicit presence)

Specify whether the Secret or its key must be defined.

### spec.podTemplate.spec.ephemeralContainers[].env[].valueFrom.fileKeyRef

`KubernetesCoreV1ApiFileKeySelector`

`fileKeyRef` selects a key of the env file. Requires the EnvFiles feature
gate to be enabled.

### spec.podTemplate.spec.ephemeralContainers[].env[].valueFrom.fileKeyRef.volumeName

`string` · required

The name of the volume mount containing the env file.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.ephemeralContainers[].env[].valueFrom.fileKeyRef.path

`string` · required

The path within the volume from which to select the file. Must be relative
and may not contain the '..' path or start with '..'.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.ephemeralContainers[].env[].valueFrom.fileKeyRef.key

`string` · required

The key within the env file. An invalid key will prevent the pod from
starting. The keys defined within a source may consist of any printable
ASCII characters except '='. During Alpha stage of the EnvFiles feature
gate, the key size is limited to 128 characters.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.ephemeralContainers[].env[].valueFrom.fileKeyRef.optional

`bool` · optional (explicit presence)

Specify whether the file or its key must be defined. If the file or key
does not exist, then the env var is not published. If optional is set to
true and the specified key does not exist, the environment variable will
not be set in the Pod's containers.

If optional is set to false and the specified key does not exist, an error
will be returned during Pod creation.

Unset, the cluster writes false.

### spec.podTemplate.spec.ephemeralContainers[].resources

`KubernetesCoreV1ApiResourceRequirements`

`resources` are not allowed for ephemeral containers. Ephemeral containers
use spare resources already allocated to the pod.

### spec.podTemplate.spec.ephemeralContainers[].resources.limits

`map<string, string>`

`limits` describes the maximum amount of compute resources allowed. More
info:
https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/

- rule: {"map":{"values":{"string":{"pattern":"^(\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))))?$"}}}}

### spec.podTemplate.spec.ephemeralContainers[].resources.requests

`map<string, string>`

`requests` describes the minimum amount of compute resources required. If
Requests is omitted for a container, it defaults to Limits if that is
explicitly specified, otherwise to an implementation-defined value.
Requests cannot exceed Limits. More info:
https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/

- rule: {"map":{"values":{"string":{"pattern":"^(\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))))?$"}}}}

### spec.podTemplate.spec.ephemeralContainers[].resources.claims

`[]KubernetesCoreV1ApiResourceClaim`

`claims` lists the names of resources, defined in spec.resourceClaims, that
are used by this container.

This field depends on the DynamicResourceAllocation feature gate.

This field is immutable. It can only be set for containers.

### spec.podTemplate.spec.ephemeralContainers[].resources.claims[].name

`string` · required

`name` must match the name of one entry in pod.spec.resourceClaims of the
Pod where this field is used. It makes that resource available inside a
container.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.ephemeralContainers[].resources.claims[].request

`string`

`request` is the name chosen for a request in the referenced claim. If
empty, everything from the claim is made available, otherwise only the
result of this request.

### spec.podTemplate.spec.ephemeralContainers[].resizePolicy

`[]KubernetesCoreV1ApiContainerResizePolicy`

Resources resize policy for the container.

Requires the InPlacePodVerticalScaling feature gate on the cluster; a
cluster without it drops this field from the pod.

### spec.podTemplate.spec.ephemeralContainers[].resizePolicy[].resourceName

`string` · required

Name of the resource to which this resource resize policy applies.
Supported values: cpu, memory.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.ephemeralContainers[].resizePolicy[].restartPolicy

`string`

Restart policy to apply when specified resource is resized. If not
specified, it defaults to NotRequired.

- rule: {"string":{"in":["NotRequired","RestartContainer"]}}

### spec.podTemplate.spec.ephemeralContainers[].restartPolicy

`string` · optional (explicit presence)

Restart policy for the container to manage the restart behavior of each
container within a pod. You cannot set this field on ephemeral containers.

- rule: {"string":{"in":["Always","Never","OnFailure"]}}

### spec.podTemplate.spec.ephemeralContainers[].restartPolicyRules

`[]KubernetesCoreV1ApiContainerRestartRule`

Represents a list of rules to be checked to determine if the container
should be restarted on exit. You cannot set this field on ephemeral
containers.

Requires the ContainerRestartRules feature gate on the cluster; a cluster
without it drops this field from the pod.

### spec.podTemplate.spec.ephemeralContainers[].restartPolicyRules[].action

`string`

Specifies the action taken on a container exit if the requirements are
satisfied. The only possible value is "Restart" to restart the container.

- rule: {"string":{"in":["Restart","RestartAllContainers"]}}

### spec.podTemplate.spec.ephemeralContainers[].restartPolicyRules[].exitCodes

`KubernetesCoreV1ApiContainerRestartRuleOnExitCodes`

Represents the exit codes to check on container exits.

### spec.podTemplate.spec.ephemeralContainers[].restartPolicyRules[].exitCodes.operator

`string`

Represents the relationship between the container exit code(s) and the
specified values. Possible values are: - In: the requirement is satisfied
if the container exit code is in the set of specified values. - NotIn: the
requirement is satisfied if the container exit code is not in the set of
specified values.

- rule: {"string":{"in":["In","NotIn"]}}

### spec.podTemplate.spec.ephemeralContainers[].restartPolicyRules[].exitCodes.values

`[]int32`

Specifies the set of values to check for container exit codes. At most 255
elements are allowed.

### spec.podTemplate.spec.ephemeralContainers[].volumeMounts

`[]KubernetesCoreV1ApiVolumeMount`

Pod volumes to mount into the container's filesystem. Subpath mounts are
not allowed for ephemeral containers. Cannot be updated.

### spec.podTemplate.spec.ephemeralContainers[].volumeMounts[].name

`string` · required

This must match the Name of a Volume.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.ephemeralContainers[].volumeMounts[].readOnly

`bool`

Mounted read-only if true, read-write otherwise (false or unspecified).
Defaults to false.

### spec.podTemplate.spec.ephemeralContainers[].volumeMounts[].recursiveReadOnly

`string` · optional (explicit presence)

`recursiveReadOnly` specifies whether read-only mounts should be handled
recursively.

If ReadOnly is false, this field has no meaning and must be unspecified.

If ReadOnly is true, and this field is set to Disabled, the mount is not
made recursively read-only.  If this field is set to IfPossible, the mount
is made recursively read-only, if it is supported by the container runtime.
If this field is set to Enabled, the mount is made recursively read-only if
it is supported by the container runtime, otherwise the pod will not be
started and an error will be generated to indicate the reason.

If this field is set to IfPossible or Enabled, MountPropagation must be set
to None (or be unspecified, which defaults to None).

If this field is not specified, it is treated as an equivalent of Disabled.

- rule: {"string":{"in":["Disabled","IfPossible","Enabled"]}}

### spec.podTemplate.spec.ephemeralContainers[].volumeMounts[].mountPath

`string` · required

Path within the container at which the volume should be mounted.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.ephemeralContainers[].volumeMounts[].subPath

`string`

Path within the volume from which the container's volume should be mounted.
Defaults to "" (volume's root).

### spec.podTemplate.spec.ephemeralContainers[].volumeMounts[].mountPropagation

`string` · optional (explicit presence)

mountPropagation determines how mounts are propagated from the host to
container and the other way around. When not set, MountPropagationNone is
used. This field is beta in 1.10. When RecursiveReadOnly is set to
IfPossible or to Enabled, MountPropagation must be None or unspecified
(which defaults to None).

- rule: {"string":{"in":["None","HostToContainer","Bidirectional"]}}

### spec.podTemplate.spec.ephemeralContainers[].volumeMounts[].subPathExpr

`string`

Expanded path within the volume from which the container's volume should be
mounted. Behaves similarly to SubPath but environment variable references
$(VAR_NAME) are expanded using the container's environment. Defaults to ""
(volume's root). SubPathExpr and SubPath are mutually exclusive.

### spec.podTemplate.spec.ephemeralContainers[].volumeMounts[].bindMountOptions

`[]string`

bindMountOptions is the list of additional bind mount options to apply when
mounting this volume into the container. Allowed values are noexec, nodev,
and nosuid. These are Linux mount options and have no effect on Windows
nodes. This field is not supported with image volumes. This is an alpha
field and requires enabling the VolumeBindMountOptions feature gate.

### spec.podTemplate.spec.ephemeralContainers[].volumeDevices

`[]KubernetesCoreV1ApiVolumeDevice`

volumeDevices is the list of block devices to be used by the container.

### spec.podTemplate.spec.ephemeralContainers[].volumeDevices[].name

`string` · required

name must match the name of a persistentVolumeClaim in the pod.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.ephemeralContainers[].volumeDevices[].devicePath

`string` · required

devicePath is the path inside of the container that the device will be
mapped to.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.ephemeralContainers[].livenessProbe

`KubernetesCoreV1ApiProbe`

Probes are not allowed for ephemeral containers.

- rule: Set at most one of exec, httpGet, tcpSocket, grpc

### spec.podTemplate.spec.ephemeralContainers[].livenessProbe.exec

`KubernetesCoreV1ApiExecAction`

`exec` specifies a command to execute in the container.

### spec.podTemplate.spec.ephemeralContainers[].livenessProbe.exec.command

`[]string`

`command` is the command line to execute inside the container, the working
directory for the command  is root ('/') in the container's filesystem. The
command is simply exec'd, it is not run inside a shell, so traditional
shell instructions ('|', etc) won't work. To use a shell, you need to
explicitly call out to that shell. Exit status of 0 is treated as
live/healthy and non-zero is unhealthy.

### spec.podTemplate.spec.ephemeralContainers[].livenessProbe.httpGet

`KubernetesCoreV1ApiHTTPGetAction`

`httpGet` specifies an HTTP GET request to perform.

### spec.podTemplate.spec.ephemeralContainers[].livenessProbe.httpGet.path

`string`

`path` to access on the HTTP server.

### spec.podTemplate.spec.ephemeralContainers[].livenessProbe.httpGet.port

`string` · required

Name or number of the port to access on the container. Number must be in
the range 1 to 65535. Name must be an IANA_SVC_NAME.

- rule: port is a port number from 1 to 65535, or a port name (up to 15 lowercase letters, digits and hyphens, with at least one letter)
- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.ephemeralContainers[].livenessProbe.httpGet.host

`string`

`host` name to connect to, defaults to the pod IP. You probably want to set
"Host" in httpHeaders instead.

### spec.podTemplate.spec.ephemeralContainers[].livenessProbe.httpGet.scheme

`string` · optional (explicit presence)

`scheme` to use for connecting to the host. Defaults to HTTP.

- rule: {"string":{"in":["HTTP","HTTPS"]}}

### spec.podTemplate.spec.ephemeralContainers[].livenessProbe.httpGet.httpHeaders

`[]KubernetesCoreV1ApiHTTPHeader`

Custom headers to set in the request. HTTP allows repeated headers.

### spec.podTemplate.spec.ephemeralContainers[].livenessProbe.httpGet.httpHeaders[].name

`string` · required

The header field name. This will be canonicalized upon output, so
case-variant names will be understood as the same header.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.ephemeralContainers[].livenessProbe.httpGet.httpHeaders[].value

`string` · required

The header field value.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.ephemeralContainers[].livenessProbe.httpGet.protocol

`string` · optional (explicit presence)

`protocol` selects the wire protocol for the probe connection. Nil defaults
to HTTP/1.1.

Requires the H2CContainerProbe feature gate on the cluster; a cluster
without it drops this field from the pod.

- rule: {"string":{"in":["HTTP1","HTTP2"]}}

### spec.podTemplate.spec.ephemeralContainers[].livenessProbe.tcpSocket

`KubernetesCoreV1ApiTCPSocketAction`

`tcpSocket` specifies a connection to a TCP port.

### spec.podTemplate.spec.ephemeralContainers[].livenessProbe.tcpSocket.port

`string` · required

Number or name of the port to access on the container. Number must be in
the range 1 to 65535. Name must be an IANA_SVC_NAME.

- rule: port is a port number from 1 to 65535, or a port name (up to 15 lowercase letters, digits and hyphens, with at least one letter)
- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.ephemeralContainers[].livenessProbe.tcpSocket.host

`string`

Optional: Host name to connect to, defaults to the pod IP.

### spec.podTemplate.spec.ephemeralContainers[].livenessProbe.grpc

`KubernetesCoreV1ApiGRPCAction`

`grpc` specifies a GRPC HealthCheckRequest.

### spec.podTemplate.spec.ephemeralContainers[].livenessProbe.grpc.port

`int32` · required · optional (explicit presence)

`port` number of the gRPC service. Number must be in the range 1 to 65535.

- rule: {"required":true}

### spec.podTemplate.spec.ephemeralContainers[].livenessProbe.grpc.service

`string` · optional (explicit presence)

`service` is the name of the service to place in the gRPC
HealthCheckRequest (see
https://github.com/grpc/grpc/blob/master/doc/health-checking.md).

If this is not specified, the default behavior is defined by gRPC.

### spec.podTemplate.spec.ephemeralContainers[].livenessProbe.grpc.mode

`string` · optional (explicit presence)

mode specifies the connection mode for the gRPC health probe. Set to "TLS"
to use TLS without certificate verification. Set to "Plaintext" to use a
plaintext (insecure) connection explicitly. If not specified, the probe
uses a plaintext (insecure) connection.

Requires the GRPCContainerProbeTLS feature gate on the cluster; a cluster
without it drops this field from the pod.

- rule: {"string":{"in":["Plaintext","TLS"]}}

### spec.podTemplate.spec.ephemeralContainers[].livenessProbe.initialDelaySeconds

`int32`

Number of seconds after the container has started before liveness probes
are initiated. More info:
https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes

### spec.podTemplate.spec.ephemeralContainers[].livenessProbe.timeoutSeconds

`int32`

Number of seconds after which the probe times out. Defaults to 1 second.
Minimum value is 1. More info:
https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes

### spec.podTemplate.spec.ephemeralContainers[].livenessProbe.periodSeconds

`int32`

How often (in seconds) to perform the probe. Default to 10 seconds. Minimum
value is 1.

### spec.podTemplate.spec.ephemeralContainers[].livenessProbe.successThreshold

`int32`

Minimum consecutive successes for the probe to be considered successful
after having failed. Defaults to 1. Must be 1 for liveness and startup.
Minimum value is 1.

### spec.podTemplate.spec.ephemeralContainers[].livenessProbe.failureThreshold

`int32`

Minimum consecutive failures for the probe to be considered failed after
having succeeded. Defaults to 3. Minimum value is 1.

### spec.podTemplate.spec.ephemeralContainers[].livenessProbe.terminationGracePeriodSeconds

`int64` · optional (explicit presence)

Optional duration in seconds the pod needs to terminate gracefully upon
probe failure. The grace period is the duration in seconds after the
processes running in the pod are sent a termination signal and the time
when the processes are forcibly halted with a kill signal. Set this value
longer than the expected cleanup time for your process. If this value is
nil, the pod's terminationGracePeriodSeconds will be used. Otherwise, this
value overrides the value provided by the pod spec. Value must be
non-negative integer. The value zero indicates stop immediately via the
kill signal (no opportunity to shut down). This is a beta field and
requires enabling ProbeTerminationGracePeriod feature gate. Minimum value
is 1. spec.terminationGracePeriodSeconds is used if unset.

### spec.podTemplate.spec.ephemeralContainers[].readinessProbe

`KubernetesCoreV1ApiProbe`

Probes are not allowed for ephemeral containers.

- rule: Set at most one of exec, httpGet, tcpSocket, grpc

### spec.podTemplate.spec.ephemeralContainers[].readinessProbe.exec

`KubernetesCoreV1ApiExecAction`

`exec` specifies a command to execute in the container.

### spec.podTemplate.spec.ephemeralContainers[].readinessProbe.exec.command

`[]string`

`command` is the command line to execute inside the container, the working
directory for the command  is root ('/') in the container's filesystem. The
command is simply exec'd, it is not run inside a shell, so traditional
shell instructions ('|', etc) won't work. To use a shell, you need to
explicitly call out to that shell. Exit status of 0 is treated as
live/healthy and non-zero is unhealthy.

### spec.podTemplate.spec.ephemeralContainers[].readinessProbe.httpGet

`KubernetesCoreV1ApiHTTPGetAction`

`httpGet` specifies an HTTP GET request to perform.

### spec.podTemplate.spec.ephemeralContainers[].readinessProbe.httpGet.path

`string`

`path` to access on the HTTP server.

### spec.podTemplate.spec.ephemeralContainers[].readinessProbe.httpGet.port

`string` · required

Name or number of the port to access on the container. Number must be in
the range 1 to 65535. Name must be an IANA_SVC_NAME.

- rule: port is a port number from 1 to 65535, or a port name (up to 15 lowercase letters, digits and hyphens, with at least one letter)
- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.ephemeralContainers[].readinessProbe.httpGet.host

`string`

`host` name to connect to, defaults to the pod IP. You probably want to set
"Host" in httpHeaders instead.

### spec.podTemplate.spec.ephemeralContainers[].readinessProbe.httpGet.scheme

`string` · optional (explicit presence)

`scheme` to use for connecting to the host. Defaults to HTTP.

- rule: {"string":{"in":["HTTP","HTTPS"]}}

### spec.podTemplate.spec.ephemeralContainers[].readinessProbe.httpGet.httpHeaders

`[]KubernetesCoreV1ApiHTTPHeader`

Custom headers to set in the request. HTTP allows repeated headers.

### spec.podTemplate.spec.ephemeralContainers[].readinessProbe.httpGet.httpHeaders[].name

`string` · required

The header field name. This will be canonicalized upon output, so
case-variant names will be understood as the same header.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.ephemeralContainers[].readinessProbe.httpGet.httpHeaders[].value

`string` · required

The header field value.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.ephemeralContainers[].readinessProbe.httpGet.protocol

`string` · optional (explicit presence)

`protocol` selects the wire protocol for the probe connection. Nil defaults
to HTTP/1.1.

Requires the H2CContainerProbe feature gate on the cluster; a cluster
without it drops this field from the pod.

- rule: {"string":{"in":["HTTP1","HTTP2"]}}

### spec.podTemplate.spec.ephemeralContainers[].readinessProbe.tcpSocket

`KubernetesCoreV1ApiTCPSocketAction`

`tcpSocket` specifies a connection to a TCP port.

### spec.podTemplate.spec.ephemeralContainers[].readinessProbe.tcpSocket.port

`string` · required

Number or name of the port to access on the container. Number must be in
the range 1 to 65535. Name must be an IANA_SVC_NAME.

- rule: port is a port number from 1 to 65535, or a port name (up to 15 lowercase letters, digits and hyphens, with at least one letter)
- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.ephemeralContainers[].readinessProbe.tcpSocket.host

`string`

Optional: Host name to connect to, defaults to the pod IP.

### spec.podTemplate.spec.ephemeralContainers[].readinessProbe.grpc

`KubernetesCoreV1ApiGRPCAction`

`grpc` specifies a GRPC HealthCheckRequest.

### spec.podTemplate.spec.ephemeralContainers[].readinessProbe.grpc.port

`int32` · required · optional (explicit presence)

`port` number of the gRPC service. Number must be in the range 1 to 65535.

- rule: {"required":true}

### spec.podTemplate.spec.ephemeralContainers[].readinessProbe.grpc.service

`string` · optional (explicit presence)

`service` is the name of the service to place in the gRPC
HealthCheckRequest (see
https://github.com/grpc/grpc/blob/master/doc/health-checking.md).

If this is not specified, the default behavior is defined by gRPC.

### spec.podTemplate.spec.ephemeralContainers[].readinessProbe.grpc.mode

`string` · optional (explicit presence)

mode specifies the connection mode for the gRPC health probe. Set to "TLS"
to use TLS without certificate verification. Set to "Plaintext" to use a
plaintext (insecure) connection explicitly. If not specified, the probe
uses a plaintext (insecure) connection.

Requires the GRPCContainerProbeTLS feature gate on the cluster; a cluster
without it drops this field from the pod.

- rule: {"string":{"in":["Plaintext","TLS"]}}

### spec.podTemplate.spec.ephemeralContainers[].readinessProbe.initialDelaySeconds

`int32`

Number of seconds after the container has started before liveness probes
are initiated. More info:
https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes

### spec.podTemplate.spec.ephemeralContainers[].readinessProbe.timeoutSeconds

`int32`

Number of seconds after which the probe times out. Defaults to 1 second.
Minimum value is 1. More info:
https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes

### spec.podTemplate.spec.ephemeralContainers[].readinessProbe.periodSeconds

`int32`

How often (in seconds) to perform the probe. Default to 10 seconds. Minimum
value is 1.

### spec.podTemplate.spec.ephemeralContainers[].readinessProbe.successThreshold

`int32`

Minimum consecutive successes for the probe to be considered successful
after having failed. Defaults to 1. Must be 1 for liveness and startup.
Minimum value is 1.

### spec.podTemplate.spec.ephemeralContainers[].readinessProbe.failureThreshold

`int32`

Minimum consecutive failures for the probe to be considered failed after
having succeeded. Defaults to 3. Minimum value is 1.

### spec.podTemplate.spec.ephemeralContainers[].readinessProbe.terminationGracePeriodSeconds

`int64` · optional (explicit presence)

Optional duration in seconds the pod needs to terminate gracefully upon
probe failure. The grace period is the duration in seconds after the
processes running in the pod are sent a termination signal and the time
when the processes are forcibly halted with a kill signal. Set this value
longer than the expected cleanup time for your process. If this value is
nil, the pod's terminationGracePeriodSeconds will be used. Otherwise, this
value overrides the value provided by the pod spec. Value must be
non-negative integer. The value zero indicates stop immediately via the
kill signal (no opportunity to shut down). This is a beta field and
requires enabling ProbeTerminationGracePeriod feature gate. Minimum value
is 1. spec.terminationGracePeriodSeconds is used if unset.

### spec.podTemplate.spec.ephemeralContainers[].startupProbe

`KubernetesCoreV1ApiProbe`

Probes are not allowed for ephemeral containers.

- rule: Set at most one of exec, httpGet, tcpSocket, grpc

### spec.podTemplate.spec.ephemeralContainers[].startupProbe.exec

`KubernetesCoreV1ApiExecAction`

`exec` specifies a command to execute in the container.

### spec.podTemplate.spec.ephemeralContainers[].startupProbe.exec.command

`[]string`

`command` is the command line to execute inside the container, the working
directory for the command  is root ('/') in the container's filesystem. The
command is simply exec'd, it is not run inside a shell, so traditional
shell instructions ('|', etc) won't work. To use a shell, you need to
explicitly call out to that shell. Exit status of 0 is treated as
live/healthy and non-zero is unhealthy.

### spec.podTemplate.spec.ephemeralContainers[].startupProbe.httpGet

`KubernetesCoreV1ApiHTTPGetAction`

`httpGet` specifies an HTTP GET request to perform.

### spec.podTemplate.spec.ephemeralContainers[].startupProbe.httpGet.path

`string`

`path` to access on the HTTP server.

### spec.podTemplate.spec.ephemeralContainers[].startupProbe.httpGet.port

`string` · required

Name or number of the port to access on the container. Number must be in
the range 1 to 65535. Name must be an IANA_SVC_NAME.

- rule: port is a port number from 1 to 65535, or a port name (up to 15 lowercase letters, digits and hyphens, with at least one letter)
- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.ephemeralContainers[].startupProbe.httpGet.host

`string`

`host` name to connect to, defaults to the pod IP. You probably want to set
"Host" in httpHeaders instead.

### spec.podTemplate.spec.ephemeralContainers[].startupProbe.httpGet.scheme

`string` · optional (explicit presence)

`scheme` to use for connecting to the host. Defaults to HTTP.

- rule: {"string":{"in":["HTTP","HTTPS"]}}

### spec.podTemplate.spec.ephemeralContainers[].startupProbe.httpGet.httpHeaders

`[]KubernetesCoreV1ApiHTTPHeader`

Custom headers to set in the request. HTTP allows repeated headers.

### spec.podTemplate.spec.ephemeralContainers[].startupProbe.httpGet.httpHeaders[].name

`string` · required

The header field name. This will be canonicalized upon output, so
case-variant names will be understood as the same header.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.ephemeralContainers[].startupProbe.httpGet.httpHeaders[].value

`string` · required

The header field value.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.ephemeralContainers[].startupProbe.httpGet.protocol

`string` · optional (explicit presence)

`protocol` selects the wire protocol for the probe connection. Nil defaults
to HTTP/1.1.

Requires the H2CContainerProbe feature gate on the cluster; a cluster
without it drops this field from the pod.

- rule: {"string":{"in":["HTTP1","HTTP2"]}}

### spec.podTemplate.spec.ephemeralContainers[].startupProbe.tcpSocket

`KubernetesCoreV1ApiTCPSocketAction`

`tcpSocket` specifies a connection to a TCP port.

### spec.podTemplate.spec.ephemeralContainers[].startupProbe.tcpSocket.port

`string` · required

Number or name of the port to access on the container. Number must be in
the range 1 to 65535. Name must be an IANA_SVC_NAME.

- rule: port is a port number from 1 to 65535, or a port name (up to 15 lowercase letters, digits and hyphens, with at least one letter)
- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.ephemeralContainers[].startupProbe.tcpSocket.host

`string`

Optional: Host name to connect to, defaults to the pod IP.

### spec.podTemplate.spec.ephemeralContainers[].startupProbe.grpc

`KubernetesCoreV1ApiGRPCAction`

`grpc` specifies a GRPC HealthCheckRequest.

### spec.podTemplate.spec.ephemeralContainers[].startupProbe.grpc.port

`int32` · required · optional (explicit presence)

`port` number of the gRPC service. Number must be in the range 1 to 65535.

- rule: {"required":true}

### spec.podTemplate.spec.ephemeralContainers[].startupProbe.grpc.service

`string` · optional (explicit presence)

`service` is the name of the service to place in the gRPC
HealthCheckRequest (see
https://github.com/grpc/grpc/blob/master/doc/health-checking.md).

If this is not specified, the default behavior is defined by gRPC.

### spec.podTemplate.spec.ephemeralContainers[].startupProbe.grpc.mode

`string` · optional (explicit presence)

mode specifies the connection mode for the gRPC health probe. Set to "TLS"
to use TLS without certificate verification. Set to "Plaintext" to use a
plaintext (insecure) connection explicitly. If not specified, the probe
uses a plaintext (insecure) connection.

Requires the GRPCContainerProbeTLS feature gate on the cluster; a cluster
without it drops this field from the pod.

- rule: {"string":{"in":["Plaintext","TLS"]}}

### spec.podTemplate.spec.ephemeralContainers[].startupProbe.initialDelaySeconds

`int32`

Number of seconds after the container has started before liveness probes
are initiated. More info:
https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes

### spec.podTemplate.spec.ephemeralContainers[].startupProbe.timeoutSeconds

`int32`

Number of seconds after which the probe times out. Defaults to 1 second.
Minimum value is 1. More info:
https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes

### spec.podTemplate.spec.ephemeralContainers[].startupProbe.periodSeconds

`int32`

How often (in seconds) to perform the probe. Default to 10 seconds. Minimum
value is 1.

### spec.podTemplate.spec.ephemeralContainers[].startupProbe.successThreshold

`int32`

Minimum consecutive successes for the probe to be considered successful
after having failed. Defaults to 1. Must be 1 for liveness and startup.
Minimum value is 1.

### spec.podTemplate.spec.ephemeralContainers[].startupProbe.failureThreshold

`int32`

Minimum consecutive failures for the probe to be considered failed after
having succeeded. Defaults to 3. Minimum value is 1.

### spec.podTemplate.spec.ephemeralContainers[].startupProbe.terminationGracePeriodSeconds

`int64` · optional (explicit presence)

Optional duration in seconds the pod needs to terminate gracefully upon
probe failure. The grace period is the duration in seconds after the
processes running in the pod are sent a termination signal and the time
when the processes are forcibly halted with a kill signal. Set this value
longer than the expected cleanup time for your process. If this value is
nil, the pod's terminationGracePeriodSeconds will be used. Otherwise, this
value overrides the value provided by the pod spec. Value must be
non-negative integer. The value zero indicates stop immediately via the
kill signal (no opportunity to shut down). This is a beta field and
requires enabling ProbeTerminationGracePeriod feature gate. Minimum value
is 1. spec.terminationGracePeriodSeconds is used if unset.

### spec.podTemplate.spec.ephemeralContainers[].lifecycle

`KubernetesCoreV1ApiLifecycle`

`lifecycle` is not allowed for ephemeral containers.

### spec.podTemplate.spec.ephemeralContainers[].lifecycle.postStart

`KubernetesCoreV1ApiLifecycleHandler`

`postStart` is called immediately after a container is created. If the
handler fails, the container is terminated and restarted according to its
restart policy. Other management of the container blocks until the hook
completes. More info:
https://kubernetes.io/docs/concepts/containers/container-lifecycle-hooks/#container-hooks

- rule: Set exactly one of exec, httpGet, sleep, tcpSocket

### spec.podTemplate.spec.ephemeralContainers[].lifecycle.postStart.exec

`KubernetesCoreV1ApiExecAction`

`exec` specifies a command to execute in the container.

### spec.podTemplate.spec.ephemeralContainers[].lifecycle.postStart.exec.command

`[]string`

`command` is the command line to execute inside the container, the working
directory for the command  is root ('/') in the container's filesystem. The
command is simply exec'd, it is not run inside a shell, so traditional
shell instructions ('|', etc) won't work. To use a shell, you need to
explicitly call out to that shell. Exit status of 0 is treated as
live/healthy and non-zero is unhealthy.

### spec.podTemplate.spec.ephemeralContainers[].lifecycle.postStart.httpGet

`KubernetesCoreV1ApiHTTPGetAction`

`httpGet` specifies an HTTP GET request to perform.

### spec.podTemplate.spec.ephemeralContainers[].lifecycle.postStart.httpGet.path

`string`

`path` to access on the HTTP server.

### spec.podTemplate.spec.ephemeralContainers[].lifecycle.postStart.httpGet.port

`string` · required

Name or number of the port to access on the container. Number must be in
the range 1 to 65535. Name must be an IANA_SVC_NAME.

- rule: port is a port number from 1 to 65535, or a port name (up to 15 lowercase letters, digits and hyphens, with at least one letter)
- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.ephemeralContainers[].lifecycle.postStart.httpGet.host

`string`

`host` name to connect to, defaults to the pod IP. You probably want to set
"Host" in httpHeaders instead.

### spec.podTemplate.spec.ephemeralContainers[].lifecycle.postStart.httpGet.scheme

`string` · optional (explicit presence)

`scheme` to use for connecting to the host. Defaults to HTTP.

- rule: {"string":{"in":["HTTP","HTTPS"]}}

### spec.podTemplate.spec.ephemeralContainers[].lifecycle.postStart.httpGet.httpHeaders

`[]KubernetesCoreV1ApiHTTPHeader`

Custom headers to set in the request. HTTP allows repeated headers.

### spec.podTemplate.spec.ephemeralContainers[].lifecycle.postStart.httpGet.httpHeaders[].name

`string` · required

The header field name. This will be canonicalized upon output, so
case-variant names will be understood as the same header.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.ephemeralContainers[].lifecycle.postStart.httpGet.httpHeaders[].value

`string` · required

The header field value.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.ephemeralContainers[].lifecycle.postStart.httpGet.protocol

`string` · optional (explicit presence)

`protocol` selects the wire protocol for the probe connection. Nil defaults
to HTTP/1.1.

Requires the H2CContainerProbe feature gate on the cluster; a cluster
without it drops this field from the pod.

- rule: {"string":{"in":["HTTP1","HTTP2"]}}

### spec.podTemplate.spec.ephemeralContainers[].lifecycle.postStart.tcpSocket

`KubernetesCoreV1ApiTCPSocketAction`

Deprecated. TCPSocket is NOT supported as a LifecycleHandler and kept for
backward compatibility. There is no validation of this field and lifecycle
hooks will fail at runtime when it is specified.

### spec.podTemplate.spec.ephemeralContainers[].lifecycle.postStart.tcpSocket.port

`string` · required

Number or name of the port to access on the container. Number must be in
the range 1 to 65535. Name must be an IANA_SVC_NAME.

- rule: port is a port number from 1 to 65535, or a port name (up to 15 lowercase letters, digits and hyphens, with at least one letter)
- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.ephemeralContainers[].lifecycle.postStart.tcpSocket.host

`string`

Optional: Host name to connect to, defaults to the pod IP.

### spec.podTemplate.spec.ephemeralContainers[].lifecycle.postStart.sleep

`KubernetesCoreV1ApiSleepAction`

`sleep` represents a duration that the container should sleep.

### spec.podTemplate.spec.ephemeralContainers[].lifecycle.postStart.sleep.seconds

`int64` · required · optional (explicit presence)

`seconds` is the number of seconds to sleep.

- rule: {"required":true}

### spec.podTemplate.spec.ephemeralContainers[].lifecycle.preStop

`KubernetesCoreV1ApiLifecycleHandler`

`preStop` is called immediately before a container is terminated due to an
API request or management event such as liveness/startup probe failure,
preemption, resource contention, etc. The handler is not called if the
container crashes or exits. The Pod's termination grace period countdown
begins before the PreStop hook is executed. Regardless of the outcome of
the handler, the container will eventually terminate within the Pod's
termination grace period (unless delayed by finalizers). Other management
of the container blocks until the hook completes or until the termination
grace period is reached. More info:
https://kubernetes.io/docs/concepts/containers/container-lifecycle-hooks/#container-hooks

- rule: Set exactly one of exec, httpGet, sleep, tcpSocket

### spec.podTemplate.spec.ephemeralContainers[].lifecycle.preStop.exec

`KubernetesCoreV1ApiExecAction`

`exec` specifies a command to execute in the container.

### spec.podTemplate.spec.ephemeralContainers[].lifecycle.preStop.exec.command

`[]string`

`command` is the command line to execute inside the container, the working
directory for the command  is root ('/') in the container's filesystem. The
command is simply exec'd, it is not run inside a shell, so traditional
shell instructions ('|', etc) won't work. To use a shell, you need to
explicitly call out to that shell. Exit status of 0 is treated as
live/healthy and non-zero is unhealthy.

### spec.podTemplate.spec.ephemeralContainers[].lifecycle.preStop.httpGet

`KubernetesCoreV1ApiHTTPGetAction`

`httpGet` specifies an HTTP GET request to perform.

### spec.podTemplate.spec.ephemeralContainers[].lifecycle.preStop.httpGet.path

`string`

`path` to access on the HTTP server.

### spec.podTemplate.spec.ephemeralContainers[].lifecycle.preStop.httpGet.port

`string` · required

Name or number of the port to access on the container. Number must be in
the range 1 to 65535. Name must be an IANA_SVC_NAME.

- rule: port is a port number from 1 to 65535, or a port name (up to 15 lowercase letters, digits and hyphens, with at least one letter)
- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.ephemeralContainers[].lifecycle.preStop.httpGet.host

`string`

`host` name to connect to, defaults to the pod IP. You probably want to set
"Host" in httpHeaders instead.

### spec.podTemplate.spec.ephemeralContainers[].lifecycle.preStop.httpGet.scheme

`string` · optional (explicit presence)

`scheme` to use for connecting to the host. Defaults to HTTP.

- rule: {"string":{"in":["HTTP","HTTPS"]}}

### spec.podTemplate.spec.ephemeralContainers[].lifecycle.preStop.httpGet.httpHeaders

`[]KubernetesCoreV1ApiHTTPHeader`

Custom headers to set in the request. HTTP allows repeated headers.

### spec.podTemplate.spec.ephemeralContainers[].lifecycle.preStop.httpGet.httpHeaders[].name

`string` · required

The header field name. This will be canonicalized upon output, so
case-variant names will be understood as the same header.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.ephemeralContainers[].lifecycle.preStop.httpGet.httpHeaders[].value

`string` · required

The header field value.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.ephemeralContainers[].lifecycle.preStop.httpGet.protocol

`string` · optional (explicit presence)

`protocol` selects the wire protocol for the probe connection. Nil defaults
to HTTP/1.1.

Requires the H2CContainerProbe feature gate on the cluster; a cluster
without it drops this field from the pod.

- rule: {"string":{"in":["HTTP1","HTTP2"]}}

### spec.podTemplate.spec.ephemeralContainers[].lifecycle.preStop.tcpSocket

`KubernetesCoreV1ApiTCPSocketAction`

Deprecated. TCPSocket is NOT supported as a LifecycleHandler and kept for
backward compatibility. There is no validation of this field and lifecycle
hooks will fail at runtime when it is specified.

### spec.podTemplate.spec.ephemeralContainers[].lifecycle.preStop.tcpSocket.port

`string` · required

Number or name of the port to access on the container. Number must be in
the range 1 to 65535. Name must be an IANA_SVC_NAME.

- rule: port is a port number from 1 to 65535, or a port name (up to 15 lowercase letters, digits and hyphens, with at least one letter)
- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.ephemeralContainers[].lifecycle.preStop.tcpSocket.host

`string`

Optional: Host name to connect to, defaults to the pod IP.

### spec.podTemplate.spec.ephemeralContainers[].lifecycle.preStop.sleep

`KubernetesCoreV1ApiSleepAction`

`sleep` represents a duration that the container should sleep.

### spec.podTemplate.spec.ephemeralContainers[].lifecycle.preStop.sleep.seconds

`int64` · required · optional (explicit presence)

`seconds` is the number of seconds to sleep.

- rule: {"required":true}

### spec.podTemplate.spec.ephemeralContainers[].lifecycle.stopSignal

`string` · optional (explicit presence)

`stopSignal` defines which signal will be sent to a container when it is
being stopped. If not specified, the default is defined by the container
runtime in use. StopSignal can only be set for Pods with a non-empty
.spec.os.name.

- rule: {"string":{"in":["SIGABRT","SIGALRM","SIGBUS","SIGCHLD","SIGCLD","SIGCONT","SIGFPE","SIGHUP","SIGILL","SIGINT","SIGIO","SIGIOT","SIGKILL","SIGPIPE","SIGPOLL","SIGPROF","SIGPWR","SIGQUIT","SIGSEGV","SIGSTKFLT","SIGSTOP","SIGSYS","SIGTERM","SIGTRAP","SIGTSTP","SIGTTIN","SIGTTOU","SIGURG","SIGUSR1","SIGUSR2","SIGVTALRM","SIGWINCH","SIGXCPU","SIGXFSZ","SIGRTMIN","SIGRTMIN+1","SIGRTMIN+2","SIGRTMIN+3","SIGRTMIN+4","SIGRTMIN+5","SIGRTMIN+6","SIGRTMIN+7","SIGRTMIN+8","SIGRTMIN+9","SIGRTMIN+10","SIGRTMIN+11","SIGRTMIN+12","SIGRTMIN+13","SIGRTMIN+14","SIGRTMIN+15","SIGRTMAX-14","SIGRTMAX-13","SIGRTMAX-12","SIGRTMAX-11","SIGRTMAX-10","SIGRTMAX-9","SIGRTMAX-8","SIGRTMAX-7","SIGRTMAX-6","SIGRTMAX-5","SIGRTMAX-4","SIGRTMAX-3","SIGRTMAX-2","SIGRTMAX-1","SIGRTMAX"]}}

### spec.podTemplate.spec.ephemeralContainers[].terminationMessagePath

`string`

Optional: Path at which the file to which the container's termination
message will be written is mounted into the container's filesystem. Message
written is intended to be brief final status, such as an assertion failure
message. Will be truncated by the node if greater than 4096 bytes. The
total message length across all containers will be limited to 12kb.
Defaults to /dev/termination-log. Cannot be updated.

### spec.podTemplate.spec.ephemeralContainers[].terminationMessagePolicy

`string` · optional (explicit presence)

Indicate how the termination message should be populated. File will use the
contents of terminationMessagePath to populate the container status message
on both success and failure. FallbackToLogsOnError will use the last chunk
of container log output if the termination message file is empty and the
container exited with an error. The log output is limited to 2048 bytes or
80 lines, whichever is smaller. Defaults to File. Cannot be updated.

- rule: {"string":{"in":["File","FallbackToLogsOnError"]}}

### spec.podTemplate.spec.ephemeralContainers[].imagePullPolicy

`string` · optional (explicit presence)

Image pull policy. One of Always, Never, IfNotPresent. Defaults to Always
if :latest tag is specified, or IfNotPresent otherwise. Cannot be updated.
More info:
https://kubernetes.io/docs/concepts/containers/images#updating-images

- rule: {"string":{"in":["Always","Never","IfNotPresent"]}}

### spec.podTemplate.spec.ephemeralContainers[].securityContext

`KubernetesCoreV1ApiSecurityContext`

Optional: SecurityContext defines the security options the ephemeral
container should be run with. If set, the fields of SecurityContext
override the equivalent fields of PodSecurityContext.

### spec.podTemplate.spec.ephemeralContainers[].securityContext.capabilities

`KubernetesCoreV1ApiCapabilities`

The capabilities to add/drop when running containers. Defaults to the
default set of capabilities granted by the container runtime. Note that
this field cannot be set when spec.os.name is windows.

### spec.podTemplate.spec.ephemeralContainers[].securityContext.capabilities.add

`[]string`

Added capabilities.

### spec.podTemplate.spec.ephemeralContainers[].securityContext.capabilities.drop

`[]string`

Removed capabilities.

### spec.podTemplate.spec.ephemeralContainers[].securityContext.privileged

`bool` · optional (explicit presence)

Run container in privileged mode. Processes in privileged containers are
essentially equivalent to root on the host. Defaults to false. Note that
this field cannot be set when spec.os.name is windows.

### spec.podTemplate.spec.ephemeralContainers[].securityContext.seLinuxOptions

`KubernetesCoreV1ApiSELinuxOptions`

The SELinux context to be applied to the container. If unspecified, the
container runtime will allocate a random SELinux context for each
container.  May also be set in PodSecurityContext.  If set in both
SecurityContext and PodSecurityContext, the value specified in
SecurityContext takes precedence. Note that this field cannot be set when
spec.os.name is windows.

### spec.podTemplate.spec.ephemeralContainers[].securityContext.seLinuxOptions.user

`string`

`user` is a SELinux user label that applies to the container.

### spec.podTemplate.spec.ephemeralContainers[].securityContext.seLinuxOptions.role

`string`

`role` is a SELinux role label that applies to the container.

### spec.podTemplate.spec.ephemeralContainers[].securityContext.seLinuxOptions.type

`string`

`type` is a SELinux type label that applies to the container.

### spec.podTemplate.spec.ephemeralContainers[].securityContext.seLinuxOptions.level

`string`

`level` is SELinux level label that applies to the container.

### spec.podTemplate.spec.ephemeralContainers[].securityContext.windowsOptions

`KubernetesCoreV1ApiWindowsSecurityContextOptions`

The Windows specific settings applied to all containers. If unspecified,
the options from the PodSecurityContext will be used. If set in both
SecurityContext and PodSecurityContext, the value specified in
SecurityContext takes precedence. Note that this field cannot be set when
spec.os.name is linux.

### spec.podTemplate.spec.ephemeralContainers[].securityContext.windowsOptions.gmsaCredentialSpecName

`string` · optional (explicit presence)

`gmsaCredentialSpecName` is the name of the GMSA credential spec to use.

### spec.podTemplate.spec.ephemeralContainers[].securityContext.windowsOptions.gmsaCredentialSpec

`string` · optional (explicit presence)

`gmsaCredentialSpec` is where the GMSA admission webhook
(https://github.com/kubernetes-sigs/windows-gmsa) inlines the contents of
the GMSA credential spec named by the GMSACredentialSpecName field.

### spec.podTemplate.spec.ephemeralContainers[].securityContext.windowsOptions.runAsUserName

`string` · optional (explicit presence)

The UserName in Windows to run the entrypoint of the container process.
Defaults to the user specified in image metadata if unspecified. May also
be set in PodSecurityContext. If set in both SecurityContext and
PodSecurityContext, the value specified in SecurityContext takes
precedence.

### spec.podTemplate.spec.ephemeralContainers[].securityContext.windowsOptions.hostProcess

`bool` · optional (explicit presence)

`hostProcess` determines if a container should be run as a 'Host Process'
container. All of a Pod's containers must have the same effective
HostProcess value (it is not allowed to have a mix of HostProcess
containers and non-HostProcess containers). In addition, if HostProcess is
true then HostNetwork must also be set to true.

### spec.podTemplate.spec.ephemeralContainers[].securityContext.runAsUser

`int64` · optional (explicit presence)

The UID to run the entrypoint of the container process. Defaults to user
specified in image metadata if unspecified. May also be set in
PodSecurityContext.  If set in both SecurityContext and PodSecurityContext,
the value specified in SecurityContext takes precedence. Note that this
field cannot be set when spec.os.name is windows.

### spec.podTemplate.spec.ephemeralContainers[].securityContext.runAsGroup

`int64` · optional (explicit presence)

The GID to run the entrypoint of the container process. Uses runtime
default if unset. May also be set in PodSecurityContext.  If set in both
SecurityContext and PodSecurityContext, the value specified in
SecurityContext takes precedence. Note that this field cannot be set when
spec.os.name is windows.

### spec.podTemplate.spec.ephemeralContainers[].securityContext.runAsNonRoot

`bool` · optional (explicit presence)

Indicates that the container must run as a non-root user. If true, the
Kubelet will validate the image at runtime to ensure that it does not run
as UID 0 (root) and fail to start the container if it does. If unset or
false, no such validation will be performed. May also be set in
PodSecurityContext.  If set in both SecurityContext and PodSecurityContext,
the value specified in SecurityContext takes precedence.

### spec.podTemplate.spec.ephemeralContainers[].securityContext.readOnlyRootFilesystem

`bool` · optional (explicit presence)

Whether this container has a read-only root filesystem. Default is false.
Note that this field cannot be set when spec.os.name is windows.

### spec.podTemplate.spec.ephemeralContainers[].securityContext.allowPrivilegeEscalation

`bool` · optional (explicit presence)

`allowPrivilegeEscalation` controls whether a process can gain more
privileges than its parent process. This bool directly controls if the
no_new_privs flag will be set on the container process.
AllowPrivilegeEscalation is true always when the container is: 1) run as
Privileged 2) has CAP_SYS_ADMIN Note that this field cannot be set when
spec.os.name is windows.

### spec.podTemplate.spec.ephemeralContainers[].securityContext.procMount

`string` · optional (explicit presence)

procMount denotes the type of proc mount to use for the containers. The
default value is Default which uses the container runtime defaults for
readonly paths and masked paths. Note that this field cannot be set when
spec.os.name is windows.

- rule: {"string":{"in":["Default","Unmasked"]}}

### spec.podTemplate.spec.ephemeralContainers[].securityContext.seccompProfile

`KubernetesCoreV1ApiSeccompProfile`

The seccomp options to use by this container. If seccomp options are
provided at both the pod & container level, the container options override
the pod options. Note that this field cannot be set when spec.os.name is
windows.

### spec.podTemplate.spec.ephemeralContainers[].securityContext.seccompProfile.type

`string`

type indicates which kind of seccomp profile will be applied. Valid options
are:

Localhost - a profile defined in a file on the node should be used.
RuntimeDefault - the container runtime default profile should be used.
Unconfined - no profile should be applied.

- rule: {"string":{"in":["Unconfined","RuntimeDefault","Localhost"]}}

### spec.podTemplate.spec.ephemeralContainers[].securityContext.seccompProfile.localhostProfile

`string` · optional (explicit presence)

localhostProfile indicates a profile defined in a file on the node should
be used. The profile must be preconfigured on the node to work. Must be a
descending path, relative to the kubelet's configured seccomp profile
location. Must be set if type is "Localhost". Must NOT be set for any other
type.

### spec.podTemplate.spec.ephemeralContainers[].securityContext.appArmorProfile

`KubernetesCoreV1ApiAppArmorProfile`

appArmorProfile is the AppArmor options to use by this container. If set,
this profile overrides the pod's appArmorProfile. Note that this field
cannot be set when spec.os.name is windows.

### spec.podTemplate.spec.ephemeralContainers[].securityContext.appArmorProfile.type

`string`

type indicates which kind of AppArmor profile will be applied. Valid
options are: Localhost - a profile pre-loaded on the node. RuntimeDefault -
the container runtime's default profile. Unconfined - no AppArmor
enforcement.

- rule: {"string":{"in":["Unconfined","RuntimeDefault","Localhost"]}}

### spec.podTemplate.spec.ephemeralContainers[].securityContext.appArmorProfile.localhostProfile

`string` · optional (explicit presence)

localhostProfile indicates a profile loaded on the node that should be
used. The profile must be preconfigured on the node to work. Must match the
loaded name of the profile. Must be set if and only if type is "Localhost".

### spec.podTemplate.spec.ephemeralContainers[].stdin

`bool`

Whether this container should allocate a buffer for stdin in the container
runtime. If this is not set, reads from stdin in the container will always
result in EOF. Default is false.

### spec.podTemplate.spec.ephemeralContainers[].stdinOnce

`bool`

Whether the container runtime should close the stdin channel after it has
been opened by a single attach. When stdin is true the stdin stream will
remain open across multiple attach sessions. If stdinOnce is set to true,
stdin is opened on container start, is empty until the first client
attaches to stdin, and then remains open and accepts data until the client
disconnects, at which time stdin is closed and remains closed until the
container is restarted. If this flag is false, a container processes that
reads from stdin will never receive an EOF. Default is false.

### spec.podTemplate.spec.ephemeralContainers[].tty

`bool`

Whether this container should allocate a TTY for itself, also requires
'stdin' to be true. Default is false.

### spec.podTemplate.spec.ephemeralContainers[].targetContainerName

`string`

If set, the name of the container from PodSpec that this ephemeral
container targets. The ephemeral container will be run in the namespaces
(IPC, PID, etc) of this container. If not set then the ephemeral container
uses the namespaces configured in the Pod spec.

The container runtime must implement support for this feature. If the
runtime does not support namespace targeting then the result of setting
this field is undefined.

### spec.podTemplate.spec.restartPolicy

`string` · optional (explicit presence)

Restart policy for all containers within the pod. One of Always, OnFailure,
Never. In some contexts, only a subset of those values may be permitted.
Default to Always. More info:
https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#restart-policy

- rule: {"string":{"in":["Always","OnFailure","Never"]}}

### spec.podTemplate.spec.terminationGracePeriodSeconds

`int64` · optional (explicit presence)

Optional duration in seconds the pod needs to terminate gracefully. May be
decreased in delete request. Value must be non-negative integer. The value
zero indicates stop immediately via the kill signal (no opportunity to shut
down). If this value is nil, the default grace period will be used instead.
The grace period is the duration in seconds after the processes running in
the pod are sent a termination signal and the time when the processes are
forcibly halted with a kill signal. Set this value longer than the expected
cleanup time for your process. Defaults to 30 seconds.

### spec.podTemplate.spec.activeDeadlineSeconds

`int64` · optional (explicit presence)

Optional duration in seconds the pod may be active on the node relative to
StartTime before the system will actively try to mark it failed and kill
associated containers. Value must be a positive integer.

### spec.podTemplate.spec.dnsPolicy

`string` · optional (explicit presence)

Set DNS policy for the pod. Defaults to "ClusterFirst". Valid values are
'ClusterFirstWithHostNet', 'ClusterFirst', 'Default' or 'None'. DNS
parameters given in DNSConfig will be merged with the policy selected with
DNSPolicy. To have DNS options set along with hostNetwork, you have to
specify DNS policy explicitly to 'ClusterFirstWithHostNet'.

- rule: {"string":{"in":["ClusterFirstWithHostNet","ClusterFirst","Default","None"]}}

### spec.podTemplate.spec.nodeSelector

`map<string, string>`

`nodeSelector` is a selector which must be true for the pod to fit on a
node. Selector which must match a node's labels for the pod to be scheduled
on that node. More info:
https://kubernetes.io/docs/concepts/configuration/assign-pod-node/

### spec.podTemplate.spec.serviceAccountName

`string | valueFrom`

`serviceAccountName` is the name of the ServiceAccount to use to run this
pod. More info:
https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/

Defaults to a KubernetesServiceAccount reference (its service_account_name
output); pass the literal name with `value:` for one Planton does not
manage.

- references: KubernetesServiceAccount (`status.outputs.service_account_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesServiceAccount, name: <that resource's name>, fieldPath: status.outputs.service_account_name}} -- a bare string does not parse

### spec.podTemplate.spec.serviceAccount

`string`

`serviceAccount` is a deprecated alias for ServiceAccountName. Deprecated:
Use serviceAccountName instead.

### spec.podTemplate.spec.automountServiceAccountToken

`bool` · optional (explicit presence)

`automountServiceAccountToken` indicates whether a service account token
should be automatically mounted.

### spec.podTemplate.spec.nodeName

`string`

`nodeName` indicates in which node this pod is scheduled. If empty, this
pod is a candidate for scheduling by the scheduler defined in
schedulerName. Once this field is set, the kubelet for this node becomes
responsible for the lifecycle of this pod. This field should not be used to
express a desire for the pod to be scheduled on a specific node.
https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#nodename

### spec.podTemplate.spec.hostNetwork

`bool`

Host networking requested for this pod. Use the host's network namespace.
When using HostNetwork you should specify ports so the scheduler is aware.
When `hostNetwork` is true, specified `hostPort` fields in port definitions
must match `containerPort`, and unspecified `hostPort` fields in port
definitions are defaulted to match `containerPort`. Default to false.

### spec.podTemplate.spec.hostPID

`bool`

Use the host's pid namespace. Optional: Default to false.

### spec.podTemplate.spec.hostIPC

`bool`

Use the host's ipc namespace. Optional: Default to false.

### spec.podTemplate.spec.shareProcessNamespace

`bool` · optional (explicit presence)

Share a single process namespace between all of the containers in a pod.
When this is set containers will be able to view and signal processes from
other containers in the same pod, and the first process in each container
will not be assigned PID 1. HostPID and ShareProcessNamespace cannot both
be set. Optional: Default to false.

### spec.podTemplate.spec.securityContext

`KubernetesCoreV1ApiPodSecurityContext`

`securityContext` holds pod-level security attributes and common container
settings. Optional: Defaults to empty.  See type description for default
values of each field.

### spec.podTemplate.spec.securityContext.seLinuxOptions

`KubernetesCoreV1ApiSELinuxOptions`

The SELinux context to be applied to all containers. If unspecified, the
container runtime will allocate a random SELinux context for each
container.  May also be set in SecurityContext.  If set in both
SecurityContext and PodSecurityContext, the value specified in
SecurityContext takes precedence for that container. Note that this field
cannot be set when spec.os.name is windows.

### spec.podTemplate.spec.securityContext.seLinuxOptions.user

`string`

`user` is a SELinux user label that applies to the container.

### spec.podTemplate.spec.securityContext.seLinuxOptions.role

`string`

`role` is a SELinux role label that applies to the container.

### spec.podTemplate.spec.securityContext.seLinuxOptions.type

`string`

`type` is a SELinux type label that applies to the container.

### spec.podTemplate.spec.securityContext.seLinuxOptions.level

`string`

`level` is SELinux level label that applies to the container.

### spec.podTemplate.spec.securityContext.windowsOptions

`KubernetesCoreV1ApiWindowsSecurityContextOptions`

The Windows specific settings applied to all containers. If unspecified,
the options within a container's SecurityContext will be used. If set in
both SecurityContext and PodSecurityContext, the value specified in
SecurityContext takes precedence. Note that this field cannot be set when
spec.os.name is linux.

### spec.podTemplate.spec.securityContext.windowsOptions.gmsaCredentialSpecName

`string` · optional (explicit presence)

`gmsaCredentialSpecName` is the name of the GMSA credential spec to use.

### spec.podTemplate.spec.securityContext.windowsOptions.gmsaCredentialSpec

`string` · optional (explicit presence)

`gmsaCredentialSpec` is where the GMSA admission webhook
(https://github.com/kubernetes-sigs/windows-gmsa) inlines the contents of
the GMSA credential spec named by the GMSACredentialSpecName field.

### spec.podTemplate.spec.securityContext.windowsOptions.runAsUserName

`string` · optional (explicit presence)

The UserName in Windows to run the entrypoint of the container process.
Defaults to the user specified in image metadata if unspecified. May also
be set in PodSecurityContext. If set in both SecurityContext and
PodSecurityContext, the value specified in SecurityContext takes
precedence.

### spec.podTemplate.spec.securityContext.windowsOptions.hostProcess

`bool` · optional (explicit presence)

`hostProcess` determines if a container should be run as a 'Host Process'
container. All of a Pod's containers must have the same effective
HostProcess value (it is not allowed to have a mix of HostProcess
containers and non-HostProcess containers). In addition, if HostProcess is
true then HostNetwork must also be set to true.

### spec.podTemplate.spec.securityContext.runAsUser

`int64` · optional (explicit presence)

The UID to run the entrypoint of the container process. Defaults to user
specified in image metadata if unspecified. May also be set in
SecurityContext.  If set in both SecurityContext and PodSecurityContext,
the value specified in SecurityContext takes precedence for that container.
Note that this field cannot be set when spec.os.name is windows.

### spec.podTemplate.spec.securityContext.runAsGroup

`int64` · optional (explicit presence)

The GID to run the entrypoint of the container process. Uses runtime
default if unset. May also be set in SecurityContext.  If set in both
SecurityContext and PodSecurityContext, the value specified in
SecurityContext takes precedence for that container. Note that this field
cannot be set when spec.os.name is windows.

### spec.podTemplate.spec.securityContext.runAsNonRoot

`bool` · optional (explicit presence)

Indicates that the container must run as a non-root user. If true, the
Kubelet will validate the image at runtime to ensure that it does not run
as UID 0 (root) and fail to start the container if it does. If unset or
false, no such validation will be performed. May also be set in
SecurityContext.  If set in both SecurityContext and PodSecurityContext,
the value specified in SecurityContext takes precedence.

### spec.podTemplate.spec.securityContext.supplementalGroups

`[]int64`

A list of groups applied to the first process run in each container, in
addition to the container's primary GID and fsGroup (if specified).  If the
SupplementalGroupsPolicy feature is enabled, the supplementalGroupsPolicy
field determines whether these are in addition to or instead of any group
memberships defined in the container image. If unspecified, no additional
groups are added, though group memberships defined in the container image
may still be used, depending on the supplementalGroupsPolicy field. Note
that this field cannot be set when spec.os.name is windows.

### spec.podTemplate.spec.securityContext.supplementalGroupsPolicy

`string` · optional (explicit presence)

Defines how supplemental groups of the first container processes are
calculated. Valid values are "Merge" and "Strict". If not specified,
"Merge" is used. (Alpha) Using the field requires the
SupplementalGroupsPolicy feature gate to be enabled and the container
runtime must implement support for this feature. Note that this field
cannot be set when spec.os.name is windows.

- rule: {"string":{"in":["Merge","Strict"]}}

### spec.podTemplate.spec.securityContext.fsGroup

`int64` · optional (explicit presence)

A special supplemental group that applies to all containers in a pod. Some
volume types allow the Kubelet to change the ownership of that volume to be
owned by the pod:

1. The owning GID will be the FSGroup 2. The setgid bit is set (new files
created in the volume will be owned by FSGroup) 3. The permission bits are
OR'd with rw-rw----.

If unset, the Kubelet will not modify the ownership and permissions of any
volume. Note that this field cannot be set when spec.os.name is windows.

### spec.podTemplate.spec.securityContext.sysctls

`[]KubernetesCoreV1ApiSysctl`

`sysctls` hold a list of namespaced sysctls used for the pod. Pods with
unsupported sysctls (by the container runtime) might fail to launch. Note
that this field cannot be set when spec.os.name is windows.

### spec.podTemplate.spec.securityContext.sysctls[].name

`string` · required

`name` of a property to set.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.securityContext.sysctls[].value

`string` · required

`value` of a property to set.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.securityContext.fsGroupChangePolicy

`string` · optional (explicit presence)

fsGroupChangePolicy defines behavior of changing ownership and permission
of the volume before being exposed inside Pod. This field will only apply
to volume types which support fsGroup based ownership(and permissions). It
will have no effect on ephemeral volume types such as: secret, configmaps
and emptydir. Valid values are "OnRootMismatch" and "Always". If not
specified, "Always" is used. Note that this field cannot be set when
spec.os.name is windows.

- rule: {"string":{"in":["OnRootMismatch","Always"]}}

### spec.podTemplate.spec.securityContext.seccompProfile

`KubernetesCoreV1ApiSeccompProfile`

The seccomp options to use by the containers in this pod. Note that this
field cannot be set when spec.os.name is windows.

### spec.podTemplate.spec.securityContext.seccompProfile.type

`string`

type indicates which kind of seccomp profile will be applied. Valid options
are:

Localhost - a profile defined in a file on the node should be used.
RuntimeDefault - the container runtime default profile should be used.
Unconfined - no profile should be applied.

- rule: {"string":{"in":["Unconfined","RuntimeDefault","Localhost"]}}

### spec.podTemplate.spec.securityContext.seccompProfile.localhostProfile

`string` · optional (explicit presence)

localhostProfile indicates a profile defined in a file on the node should
be used. The profile must be preconfigured on the node to work. Must be a
descending path, relative to the kubelet's configured seccomp profile
location. Must be set if type is "Localhost". Must NOT be set for any other
type.

### spec.podTemplate.spec.securityContext.appArmorProfile

`KubernetesCoreV1ApiAppArmorProfile`

appArmorProfile is the AppArmor options to use by the containers in this
pod. Note that this field cannot be set when spec.os.name is windows.

### spec.podTemplate.spec.securityContext.appArmorProfile.type

`string`

type indicates which kind of AppArmor profile will be applied. Valid
options are: Localhost - a profile pre-loaded on the node. RuntimeDefault -
the container runtime's default profile. Unconfined - no AppArmor
enforcement.

- rule: {"string":{"in":["Unconfined","RuntimeDefault","Localhost"]}}

### spec.podTemplate.spec.securityContext.appArmorProfile.localhostProfile

`string` · optional (explicit presence)

localhostProfile indicates a profile loaded on the node that should be
used. The profile must be preconfigured on the node to work. Must match the
loaded name of the profile. Must be set if and only if type is "Localhost".

### spec.podTemplate.spec.securityContext.seLinuxChangePolicy

`string` · optional (explicit presence)

seLinuxChangePolicy defines how the container's SELinux label is applied to
all volumes used by the Pod. It has no effect on nodes that do not support
SELinux or to volumes does not support SELinux. Valid values are
"MountOption" and "Recursive".

"Recursive" means relabeling of all files on all Pod volumes by the
container runtime. This may be slow for large volumes, but allows mixing
privileged and unprivileged Pods sharing the same volume on the same node.

"MountOption" mounts all eligible Pod volumes with `-o context` mount
option. This requires all Pods that share the same volume to use the same
SELinux label. It is not possible to share the same volume among privileged
and unprivileged Pods. Eligible volumes are in-tree FibreChannel and iSCSI
volumes, and all CSI volumes whose CSI driver announces SELinux support by
setting spec.seLinuxMount: true in their CSIDriver instance. Other volumes
are always re-labelled recursively.

If not specified, "MountOption" is used.

This field affects only Pods that have SELinux label set, either in
PodSecurityContext or in SecurityContext of all containers.

All Pods that use the same volume should use the same seLinuxChangePolicy,
otherwise some pods can get stuck in ContainerCreating state. Note that
this field cannot be set when spec.os.name is windows.

Requires the SELinuxChangePolicy feature gate on the cluster; a cluster
without it drops this field from the pod.

- rule: {"string":{"in":["Recursive","MountOption"]}}

### spec.podTemplate.spec.imagePullSecrets

`[]KubernetesCoreV1ApiLocalObjectReference`

`imagePullSecrets` is an optional list of references to secrets in the same
namespace to use for pulling any of the images used by this PodSpec. If
specified, these secrets will be passed to individual puller
implementations for them to use. More info:
https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod

### spec.podTemplate.spec.imagePullSecrets[].name

`string | valueFrom` · required

`name` of the referent. This field is effectively required, but due to
backwards compatibility is allowed to be empty. Instances of this type with
an empty value here are almost certainly wrong. More info:
https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names

Defaults to a KubernetesSecret reference (its secret_name output); pass the
literal name with `value:` for one Planton does not manage.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.podTemplate.spec.hostname

`string`

Specifies the hostname of the Pod If not specified, the pod's hostname will
be set to a system-defined value.

### spec.podTemplate.spec.subdomain

`string`

If specified, the fully qualified Pod hostname will be
"<hostname>.<subdomain>.<pod namespace>.svc.<cluster domain>". If not
specified, the pod will not have a domainname at all.

### spec.podTemplate.spec.affinity

`KubernetesCoreV1ApiAffinity`

If specified, the pod's scheduling constraints.

### spec.podTemplate.spec.affinity.nodeAffinity

`KubernetesCoreV1ApiNodeAffinity`

Describes node affinity scheduling rules for the pod.

### spec.podTemplate.spec.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution

`KubernetesCoreV1ApiNodeSelector`

If the affinity requirements specified by this field are not met at
scheduling time, the pod will not be scheduled onto the node. If the
affinity requirements specified by this field cease to be met at some point
during pod execution (e.g. due to an update), the system may or may not try
to eventually evict the pod from its node.

### spec.podTemplate.spec.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms

`[]KubernetesCoreV1ApiNodeSelectorTerm` · required

Required. A list of node selector terms. The terms are ORed.

- rule: {"repeated":{"minItems":"1"}}

### spec.podTemplate.spec.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[].matchExpressions

`[]KubernetesCoreV1ApiNodeSelectorRequirement`

A list of node selector requirements by node's labels.

- rule: In and NotIn need at least one value, Gt and Lt exactly one integer, and Exists and DoesNotExist none

### spec.podTemplate.spec.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[].matchExpressions[].key

`string` · required

The label key that the selector applies to.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[].matchExpressions[].operator

`string`

Represents a key's relationship to a set of values. Valid operators are In,
NotIn, Exists, DoesNotExist. Gt, and Lt.

- rule: {"string":{"in":["In","NotIn","Exists","DoesNotExist","Gt","Lt"]}}

### spec.podTemplate.spec.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[].matchExpressions[].values

`[]string`

An array of string values. If the operator is In or NotIn, the values array
must be non-empty. If the operator is Exists or DoesNotExist, the values
array must be empty. If the operator is Gt or Lt, the values array must
have a single element, which will be interpreted as an integer. This array
is replaced during a strategic merge patch.

### spec.podTemplate.spec.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[].matchFields

`[]KubernetesCoreV1ApiNodeSelectorRequirement`

A list of node selector requirements by node's fields.

- rule: In and NotIn need at least one value, Gt and Lt exactly one integer, and Exists and DoesNotExist none

### spec.podTemplate.spec.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[].matchFields[].key

`string` · required

The label key that the selector applies to.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[].matchFields[].operator

`string`

Represents a key's relationship to a set of values. Valid operators are In,
NotIn, Exists, DoesNotExist. Gt, and Lt.

- rule: {"string":{"in":["In","NotIn","Exists","DoesNotExist","Gt","Lt"]}}

### spec.podTemplate.spec.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[].matchFields[].values

`[]string`

An array of string values. If the operator is In or NotIn, the values array
must be non-empty. If the operator is Exists or DoesNotExist, the values
array must be empty. If the operator is Gt or Lt, the values array must
have a single element, which will be interpreted as an integer. This array
is replaced during a strategic merge patch.

### spec.podTemplate.spec.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution

`[]KubernetesCoreV1ApiPreferredSchedulingTerm`

The scheduler will prefer to schedule pods to nodes that satisfy the
affinity expressions specified by this field, but it may choose a node that
violates one or more of the expressions. The node that is most preferred is
the one with the greatest sum of weights, i.e. for each node that meets all
of the scheduling requirements (resource request, requiredDuringScheduling
affinity expressions, etc.), compute a sum by iterating through the
elements of this field and adding "weight" to the sum if the node matches
the corresponding matchExpressions; the node(s) with the highest sum are
the most preferred.

### spec.podTemplate.spec.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[].weight

`int32` · required · optional (explicit presence)

`weight` associated with matching the corresponding nodeSelectorTerm, in
the range 1-100.

- rule: {"required":true}

### spec.podTemplate.spec.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[].preference

`KubernetesCoreV1ApiNodeSelectorTerm` · required

A node selector term, associated with the corresponding weight.

- rule: {"required":true}

### spec.podTemplate.spec.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[].preference.matchExpressions

`[]KubernetesCoreV1ApiNodeSelectorRequirement`

A list of node selector requirements by node's labels.

- rule: In and NotIn need at least one value, Gt and Lt exactly one integer, and Exists and DoesNotExist none

### spec.podTemplate.spec.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[].preference.matchExpressions[].key

`string` · required

The label key that the selector applies to.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[].preference.matchExpressions[].operator

`string`

Represents a key's relationship to a set of values. Valid operators are In,
NotIn, Exists, DoesNotExist. Gt, and Lt.

- rule: {"string":{"in":["In","NotIn","Exists","DoesNotExist","Gt","Lt"]}}

### spec.podTemplate.spec.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[].preference.matchExpressions[].values

`[]string`

An array of string values. If the operator is In or NotIn, the values array
must be non-empty. If the operator is Exists or DoesNotExist, the values
array must be empty. If the operator is Gt or Lt, the values array must
have a single element, which will be interpreted as an integer. This array
is replaced during a strategic merge patch.

### spec.podTemplate.spec.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[].preference.matchFields

`[]KubernetesCoreV1ApiNodeSelectorRequirement`

A list of node selector requirements by node's fields.

- rule: In and NotIn need at least one value, Gt and Lt exactly one integer, and Exists and DoesNotExist none

### spec.podTemplate.spec.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[].preference.matchFields[].key

`string` · required

The label key that the selector applies to.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[].preference.matchFields[].operator

`string`

Represents a key's relationship to a set of values. Valid operators are In,
NotIn, Exists, DoesNotExist. Gt, and Lt.

- rule: {"string":{"in":["In","NotIn","Exists","DoesNotExist","Gt","Lt"]}}

### spec.podTemplate.spec.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[].preference.matchFields[].values

`[]string`

An array of string values. If the operator is In or NotIn, the values array
must be non-empty. If the operator is Exists or DoesNotExist, the values
array must be empty. If the operator is Gt or Lt, the values array must
have a single element, which will be interpreted as an integer. This array
is replaced during a strategic merge patch.

### spec.podTemplate.spec.affinity.podAffinity

`KubernetesCoreV1ApiPodAffinity`

Describes pod affinity scheduling rules (e.g. co-locate this pod in the
same node, zone, etc. as some other pod(s)).

### spec.podTemplate.spec.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution

`[]KubernetesCoreV1ApiPodAffinityTerm`

If the affinity requirements specified by this field are not met at
scheduling time, the pod will not be scheduled onto the node. If the
affinity requirements specified by this field cease to be met at some point
during pod execution (e.g. due to a pod label update), the system may or
may not try to eventually evict the pod from its node. When there are
multiple elements, the lists of nodes corresponding to each podAffinityTerm
are intersected, i.e. all terms must be satisfied.

### spec.podTemplate.spec.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[].labelSelector

`KubernetesCoreV1ApiLabelSelector`

A label query over a set of resources, in this case pods. If it's null,
this PodAffinityTerm matches with no Pods.

### spec.podTemplate.spec.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[].labelSelector.matchLabels

`map<string, string>`

matchLabels is a map of {key,value} pairs. A single {key,value} in the
matchLabels map is equivalent to an element of matchExpressions, whose key
field is "key", the operator is "In", and the values array contains only
"value". The requirements are ANDed.

### spec.podTemplate.spec.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[].labelSelector.matchExpressions

`[]KubernetesCoreV1ApiLabelSelectorRequirement`

matchExpressions is a list of label selector requirements. The requirements
are ANDed.

- rule: In and NotIn need at least one value; Exists and DoesNotExist take none

### spec.podTemplate.spec.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[].labelSelector.matchExpressions[].key

`string` · required

key is the label key that the selector applies to.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[].labelSelector.matchExpressions[].operator

`string` · required

operator represents a key's relationship to a set of values. Valid
operators are In, NotIn, Exists and DoesNotExist.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[].labelSelector.matchExpressions[].values

`[]string`

values is an array of string values. If the operator is In or NotIn, the
values array must be non-empty. If the operator is Exists or DoesNotExist,
the values array must be empty. This array is replaced during a strategic
merge patch.

### spec.podTemplate.spec.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[].namespaces

`[]string`

namespaces specifies a static list of namespace names that the term applies
to. The term is applied to the union of the namespaces listed in this field
and the ones selected by namespaceSelector. null or empty namespaces list
and null namespaceSelector means "this pod's namespace".

### spec.podTemplate.spec.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[].topologyKey

`string` · required

This pod should be co-located (affinity) or not co-located (anti-affinity)
with the pods matching the labelSelector in the specified namespaces, where
co-located is defined as running on a node whose value of the label with
key topologyKey matches that of any node on which any of the selected pods
is running. Empty topologyKey is not allowed.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[].namespaceSelector

`KubernetesCoreV1ApiLabelSelector`

A label query over the set of namespaces that the term applies to. The term
is applied to the union of the namespaces selected by this field and the
ones listed in the namespaces field. null selector and null or empty
namespaces list means "this pod's namespace". An empty selector ({})
matches all namespaces.

### spec.podTemplate.spec.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[].namespaceSelector.matchLabels

`map<string, string>`

matchLabels is a map of {key,value} pairs. A single {key,value} in the
matchLabels map is equivalent to an element of matchExpressions, whose key
field is "key", the operator is "In", and the values array contains only
"value". The requirements are ANDed.

### spec.podTemplate.spec.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[].namespaceSelector.matchExpressions

`[]KubernetesCoreV1ApiLabelSelectorRequirement`

matchExpressions is a list of label selector requirements. The requirements
are ANDed.

- rule: In and NotIn need at least one value; Exists and DoesNotExist take none

### spec.podTemplate.spec.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[].namespaceSelector.matchExpressions[].key

`string` · required

key is the label key that the selector applies to.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[].namespaceSelector.matchExpressions[].operator

`string` · required

operator represents a key's relationship to a set of values. Valid
operators are In, NotIn, Exists and DoesNotExist.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[].namespaceSelector.matchExpressions[].values

`[]string`

values is an array of string values. If the operator is In or NotIn, the
values array must be non-empty. If the operator is Exists or DoesNotExist,
the values array must be empty. This array is replaced during a strategic
merge patch.

### spec.podTemplate.spec.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[].matchLabelKeys

`[]string`

`matchLabelKeys` is a set of pod label keys to select which pods will be
taken into consideration. The keys are used to lookup values from the
incoming pod labels, those key-value labels are merged with `labelSelector`
as `key in (value)` to select the group of existing pods which pods will be
taken into consideration for the incoming pod's pod (anti) affinity. Keys
that don't exist in the incoming pod labels will be ignored. The default
value is empty. The same key is forbidden to exist in both matchLabelKeys
and labelSelector. Also, matchLabelKeys cannot be set when labelSelector
isn't set.

### spec.podTemplate.spec.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[].mismatchLabelKeys

`[]string`

`mismatchLabelKeys` is a set of pod label keys to select which pods will be
taken into consideration. The keys are used to lookup values from the
incoming pod labels, those key-value labels are merged with `labelSelector`
as `key notin (value)` to select the group of existing pods which pods will
be taken into consideration for the incoming pod's pod (anti) affinity.
Keys that don't exist in the incoming pod labels will be ignored. The
default value is empty. The same key is forbidden to exist in both
mismatchLabelKeys and labelSelector. Also, mismatchLabelKeys cannot be set
when labelSelector isn't set.

### spec.podTemplate.spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution

`[]KubernetesCoreV1ApiWeightedPodAffinityTerm`

The scheduler will prefer to schedule pods to nodes that satisfy the
affinity expressions specified by this field, but it may choose a node that
violates one or more of the expressions. The node that is most preferred is
the one with the greatest sum of weights, i.e. for each node that meets all
of the scheduling requirements (resource request, requiredDuringScheduling
affinity expressions, etc.), compute a sum by iterating through the
elements of this field and adding "weight" to the sum if the node has pods
which matches the corresponding podAffinityTerm; the node(s) with the
highest sum are the most preferred.

### spec.podTemplate.spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution[].weight

`int32` · required · optional (explicit presence)

weight associated with matching the corresponding podAffinityTerm, in the
range 1-100.

- rule: {"required":true}

### spec.podTemplate.spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm

`KubernetesCoreV1ApiPodAffinityTerm` · required

Required. A pod affinity term, associated with the corresponding weight.

- rule: {"required":true}

### spec.podTemplate.spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.labelSelector

`KubernetesCoreV1ApiLabelSelector`

A label query over a set of resources, in this case pods. If it's null,
this PodAffinityTerm matches with no Pods.

### spec.podTemplate.spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.labelSelector.matchLabels

`map<string, string>`

matchLabels is a map of {key,value} pairs. A single {key,value} in the
matchLabels map is equivalent to an element of matchExpressions, whose key
field is "key", the operator is "In", and the values array contains only
"value". The requirements are ANDed.

### spec.podTemplate.spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.labelSelector.matchExpressions

`[]KubernetesCoreV1ApiLabelSelectorRequirement`

matchExpressions is a list of label selector requirements. The requirements
are ANDed.

- rule: In and NotIn need at least one value; Exists and DoesNotExist take none

### spec.podTemplate.spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.labelSelector.matchExpressions[].key

`string` · required

key is the label key that the selector applies to.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.labelSelector.matchExpressions[].operator

`string` · required

operator represents a key's relationship to a set of values. Valid
operators are In, NotIn, Exists and DoesNotExist.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.labelSelector.matchExpressions[].values

`[]string`

values is an array of string values. If the operator is In or NotIn, the
values array must be non-empty. If the operator is Exists or DoesNotExist,
the values array must be empty. This array is replaced during a strategic
merge patch.

### spec.podTemplate.spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.namespaces

`[]string`

namespaces specifies a static list of namespace names that the term applies
to. The term is applied to the union of the namespaces listed in this field
and the ones selected by namespaceSelector. null or empty namespaces list
and null namespaceSelector means "this pod's namespace".

### spec.podTemplate.spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.topologyKey

`string` · required

This pod should be co-located (affinity) or not co-located (anti-affinity)
with the pods matching the labelSelector in the specified namespaces, where
co-located is defined as running on a node whose value of the label with
key topologyKey matches that of any node on which any of the selected pods
is running. Empty topologyKey is not allowed.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.namespaceSelector

`KubernetesCoreV1ApiLabelSelector`

A label query over the set of namespaces that the term applies to. The term
is applied to the union of the namespaces selected by this field and the
ones listed in the namespaces field. null selector and null or empty
namespaces list means "this pod's namespace". An empty selector ({})
matches all namespaces.

### spec.podTemplate.spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.namespaceSelector.matchLabels

`map<string, string>`

matchLabels is a map of {key,value} pairs. A single {key,value} in the
matchLabels map is equivalent to an element of matchExpressions, whose key
field is "key", the operator is "In", and the values array contains only
"value". The requirements are ANDed.

### spec.podTemplate.spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.namespaceSelector.matchExpressions

`[]KubernetesCoreV1ApiLabelSelectorRequirement`

matchExpressions is a list of label selector requirements. The requirements
are ANDed.

- rule: In and NotIn need at least one value; Exists and DoesNotExist take none

### spec.podTemplate.spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.namespaceSelector.matchExpressions[].key

`string` · required

key is the label key that the selector applies to.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.namespaceSelector.matchExpressions[].operator

`string` · required

operator represents a key's relationship to a set of values. Valid
operators are In, NotIn, Exists and DoesNotExist.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.namespaceSelector.matchExpressions[].values

`[]string`

values is an array of string values. If the operator is In or NotIn, the
values array must be non-empty. If the operator is Exists or DoesNotExist,
the values array must be empty. This array is replaced during a strategic
merge patch.

### spec.podTemplate.spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.matchLabelKeys

`[]string`

`matchLabelKeys` is a set of pod label keys to select which pods will be
taken into consideration. The keys are used to lookup values from the
incoming pod labels, those key-value labels are merged with `labelSelector`
as `key in (value)` to select the group of existing pods which pods will be
taken into consideration for the incoming pod's pod (anti) affinity. Keys
that don't exist in the incoming pod labels will be ignored. The default
value is empty. The same key is forbidden to exist in both matchLabelKeys
and labelSelector. Also, matchLabelKeys cannot be set when labelSelector
isn't set.

### spec.podTemplate.spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.mismatchLabelKeys

`[]string`

`mismatchLabelKeys` is a set of pod label keys to select which pods will be
taken into consideration. The keys are used to lookup values from the
incoming pod labels, those key-value labels are merged with `labelSelector`
as `key notin (value)` to select the group of existing pods which pods will
be taken into consideration for the incoming pod's pod (anti) affinity.
Keys that don't exist in the incoming pod labels will be ignored. The
default value is empty. The same key is forbidden to exist in both
mismatchLabelKeys and labelSelector. Also, mismatchLabelKeys cannot be set
when labelSelector isn't set.

### spec.podTemplate.spec.affinity.podAntiAffinity

`KubernetesCoreV1ApiPodAntiAffinity`

Describes pod anti-affinity scheduling rules (e.g. avoid putting this pod
in the same node, zone, etc. as some other pod(s)).

### spec.podTemplate.spec.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution

`[]KubernetesCoreV1ApiPodAffinityTerm`

If the anti-affinity requirements specified by this field are not met at
scheduling time, the pod will not be scheduled onto the node. If the
anti-affinity requirements specified by this field cease to be met at some
point during pod execution (e.g. due to a pod label update), the system may
or may not try to eventually evict the pod from its node. When there are
multiple elements, the lists of nodes corresponding to each podAffinityTerm
are intersected, i.e. all terms must be satisfied.

### spec.podTemplate.spec.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[].labelSelector

`KubernetesCoreV1ApiLabelSelector`

A label query over a set of resources, in this case pods. If it's null,
this PodAffinityTerm matches with no Pods.

### spec.podTemplate.spec.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[].labelSelector.matchLabels

`map<string, string>`

matchLabels is a map of {key,value} pairs. A single {key,value} in the
matchLabels map is equivalent to an element of matchExpressions, whose key
field is "key", the operator is "In", and the values array contains only
"value". The requirements are ANDed.

### spec.podTemplate.spec.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[].labelSelector.matchExpressions

`[]KubernetesCoreV1ApiLabelSelectorRequirement`

matchExpressions is a list of label selector requirements. The requirements
are ANDed.

- rule: In and NotIn need at least one value; Exists and DoesNotExist take none

### spec.podTemplate.spec.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[].labelSelector.matchExpressions[].key

`string` · required

key is the label key that the selector applies to.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[].labelSelector.matchExpressions[].operator

`string` · required

operator represents a key's relationship to a set of values. Valid
operators are In, NotIn, Exists and DoesNotExist.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[].labelSelector.matchExpressions[].values

`[]string`

values is an array of string values. If the operator is In or NotIn, the
values array must be non-empty. If the operator is Exists or DoesNotExist,
the values array must be empty. This array is replaced during a strategic
merge patch.

### spec.podTemplate.spec.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[].namespaces

`[]string`

namespaces specifies a static list of namespace names that the term applies
to. The term is applied to the union of the namespaces listed in this field
and the ones selected by namespaceSelector. null or empty namespaces list
and null namespaceSelector means "this pod's namespace".

### spec.podTemplate.spec.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[].topologyKey

`string` · required

This pod should be co-located (affinity) or not co-located (anti-affinity)
with the pods matching the labelSelector in the specified namespaces, where
co-located is defined as running on a node whose value of the label with
key topologyKey matches that of any node on which any of the selected pods
is running. Empty topologyKey is not allowed.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[].namespaceSelector

`KubernetesCoreV1ApiLabelSelector`

A label query over the set of namespaces that the term applies to. The term
is applied to the union of the namespaces selected by this field and the
ones listed in the namespaces field. null selector and null or empty
namespaces list means "this pod's namespace". An empty selector ({})
matches all namespaces.

### spec.podTemplate.spec.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[].namespaceSelector.matchLabels

`map<string, string>`

matchLabels is a map of {key,value} pairs. A single {key,value} in the
matchLabels map is equivalent to an element of matchExpressions, whose key
field is "key", the operator is "In", and the values array contains only
"value". The requirements are ANDed.

### spec.podTemplate.spec.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[].namespaceSelector.matchExpressions

`[]KubernetesCoreV1ApiLabelSelectorRequirement`

matchExpressions is a list of label selector requirements. The requirements
are ANDed.

- rule: In and NotIn need at least one value; Exists and DoesNotExist take none

### spec.podTemplate.spec.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[].namespaceSelector.matchExpressions[].key

`string` · required

key is the label key that the selector applies to.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[].namespaceSelector.matchExpressions[].operator

`string` · required

operator represents a key's relationship to a set of values. Valid
operators are In, NotIn, Exists and DoesNotExist.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[].namespaceSelector.matchExpressions[].values

`[]string`

values is an array of string values. If the operator is In or NotIn, the
values array must be non-empty. If the operator is Exists or DoesNotExist,
the values array must be empty. This array is replaced during a strategic
merge patch.

### spec.podTemplate.spec.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[].matchLabelKeys

`[]string`

`matchLabelKeys` is a set of pod label keys to select which pods will be
taken into consideration. The keys are used to lookup values from the
incoming pod labels, those key-value labels are merged with `labelSelector`
as `key in (value)` to select the group of existing pods which pods will be
taken into consideration for the incoming pod's pod (anti) affinity. Keys
that don't exist in the incoming pod labels will be ignored. The default
value is empty. The same key is forbidden to exist in both matchLabelKeys
and labelSelector. Also, matchLabelKeys cannot be set when labelSelector
isn't set.

### spec.podTemplate.spec.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[].mismatchLabelKeys

`[]string`

`mismatchLabelKeys` is a set of pod label keys to select which pods will be
taken into consideration. The keys are used to lookup values from the
incoming pod labels, those key-value labels are merged with `labelSelector`
as `key notin (value)` to select the group of existing pods which pods will
be taken into consideration for the incoming pod's pod (anti) affinity.
Keys that don't exist in the incoming pod labels will be ignored. The
default value is empty. The same key is forbidden to exist in both
mismatchLabelKeys and labelSelector. Also, mismatchLabelKeys cannot be set
when labelSelector isn't set.

### spec.podTemplate.spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution

`[]KubernetesCoreV1ApiWeightedPodAffinityTerm`

The scheduler will prefer to schedule pods to nodes that satisfy the
anti-affinity expressions specified by this field, but it may choose a node
that violates one or more of the expressions. The node that is most
preferred is the one with the greatest sum of weights, i.e. for each node
that meets all of the scheduling requirements (resource request,
requiredDuringScheduling anti-affinity expressions, etc.), compute a sum by
iterating through the elements of this field and subtracting "weight" from
the sum if the node has pods which matches the corresponding
podAffinityTerm; the node(s) with the highest sum are the most preferred.

### spec.podTemplate.spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[].weight

`int32` · required · optional (explicit presence)

weight associated with matching the corresponding podAffinityTerm, in the
range 1-100.

- rule: {"required":true}

### spec.podTemplate.spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm

`KubernetesCoreV1ApiPodAffinityTerm` · required

Required. A pod affinity term, associated with the corresponding weight.

- rule: {"required":true}

### spec.podTemplate.spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.labelSelector

`KubernetesCoreV1ApiLabelSelector`

A label query over a set of resources, in this case pods. If it's null,
this PodAffinityTerm matches with no Pods.

### spec.podTemplate.spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.labelSelector.matchLabels

`map<string, string>`

matchLabels is a map of {key,value} pairs. A single {key,value} in the
matchLabels map is equivalent to an element of matchExpressions, whose key
field is "key", the operator is "In", and the values array contains only
"value". The requirements are ANDed.

### spec.podTemplate.spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.labelSelector.matchExpressions

`[]KubernetesCoreV1ApiLabelSelectorRequirement`

matchExpressions is a list of label selector requirements. The requirements
are ANDed.

- rule: In and NotIn need at least one value; Exists and DoesNotExist take none

### spec.podTemplate.spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.labelSelector.matchExpressions[].key

`string` · required

key is the label key that the selector applies to.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.labelSelector.matchExpressions[].operator

`string` · required

operator represents a key's relationship to a set of values. Valid
operators are In, NotIn, Exists and DoesNotExist.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.labelSelector.matchExpressions[].values

`[]string`

values is an array of string values. If the operator is In or NotIn, the
values array must be non-empty. If the operator is Exists or DoesNotExist,
the values array must be empty. This array is replaced during a strategic
merge patch.

### spec.podTemplate.spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.namespaces

`[]string`

namespaces specifies a static list of namespace names that the term applies
to. The term is applied to the union of the namespaces listed in this field
and the ones selected by namespaceSelector. null or empty namespaces list
and null namespaceSelector means "this pod's namespace".

### spec.podTemplate.spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.topologyKey

`string` · required

This pod should be co-located (affinity) or not co-located (anti-affinity)
with the pods matching the labelSelector in the specified namespaces, where
co-located is defined as running on a node whose value of the label with
key topologyKey matches that of any node on which any of the selected pods
is running. Empty topologyKey is not allowed.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.namespaceSelector

`KubernetesCoreV1ApiLabelSelector`

A label query over the set of namespaces that the term applies to. The term
is applied to the union of the namespaces selected by this field and the
ones listed in the namespaces field. null selector and null or empty
namespaces list means "this pod's namespace". An empty selector ({})
matches all namespaces.

### spec.podTemplate.spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.namespaceSelector.matchLabels

`map<string, string>`

matchLabels is a map of {key,value} pairs. A single {key,value} in the
matchLabels map is equivalent to an element of matchExpressions, whose key
field is "key", the operator is "In", and the values array contains only
"value". The requirements are ANDed.

### spec.podTemplate.spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.namespaceSelector.matchExpressions

`[]KubernetesCoreV1ApiLabelSelectorRequirement`

matchExpressions is a list of label selector requirements. The requirements
are ANDed.

- rule: In and NotIn need at least one value; Exists and DoesNotExist take none

### spec.podTemplate.spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.namespaceSelector.matchExpressions[].key

`string` · required

key is the label key that the selector applies to.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.namespaceSelector.matchExpressions[].operator

`string` · required

operator represents a key's relationship to a set of values. Valid
operators are In, NotIn, Exists and DoesNotExist.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.namespaceSelector.matchExpressions[].values

`[]string`

values is an array of string values. If the operator is In or NotIn, the
values array must be non-empty. If the operator is Exists or DoesNotExist,
the values array must be empty. This array is replaced during a strategic
merge patch.

### spec.podTemplate.spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.matchLabelKeys

`[]string`

`matchLabelKeys` is a set of pod label keys to select which pods will be
taken into consideration. The keys are used to lookup values from the
incoming pod labels, those key-value labels are merged with `labelSelector`
as `key in (value)` to select the group of existing pods which pods will be
taken into consideration for the incoming pod's pod (anti) affinity. Keys
that don't exist in the incoming pod labels will be ignored. The default
value is empty. The same key is forbidden to exist in both matchLabelKeys
and labelSelector. Also, matchLabelKeys cannot be set when labelSelector
isn't set.

### spec.podTemplate.spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[].podAffinityTerm.mismatchLabelKeys

`[]string`

`mismatchLabelKeys` is a set of pod label keys to select which pods will be
taken into consideration. The keys are used to lookup values from the
incoming pod labels, those key-value labels are merged with `labelSelector`
as `key notin (value)` to select the group of existing pods which pods will
be taken into consideration for the incoming pod's pod (anti) affinity.
Keys that don't exist in the incoming pod labels will be ignored. The
default value is empty. The same key is forbidden to exist in both
mismatchLabelKeys and labelSelector. Also, mismatchLabelKeys cannot be set
when labelSelector isn't set.

### spec.podTemplate.spec.schedulerName

`string`

If specified, the pod will be dispatched by specified scheduler. If not
specified, the pod will be dispatched by default scheduler.

### spec.podTemplate.spec.tolerations

`[]KubernetesCoreV1ApiToleration`

If specified, the pod's tolerations.

- rule: A toleration with operator Exists matches any value, so it takes no value

### spec.podTemplate.spec.tolerations[].key

`string`

`key` is the taint key that the toleration applies to. Empty means match
all taint keys. If the key is empty, operator must be Exists; this
combination means to match all values and all keys.

### spec.podTemplate.spec.tolerations[].operator

`string` · optional (explicit presence)

`operator` represents a key's relationship to the value. Valid operators
are Exists, Equal, Lt, and Gt. Defaults to Equal. Exists is equivalent to
wildcard for value, so that a pod can tolerate all taints of a particular
category. Lt and Gt perform numeric comparisons (requires feature gate
TaintTolerationComparisonOperators).

- rule: {"string":{"in":["Exists","Equal","Lt","Gt"]}}

### spec.podTemplate.spec.tolerations[].value

`string`

`value` is the taint value the toleration matches to. If the operator is
Exists, the value should be empty, otherwise just a regular string.

### spec.podTemplate.spec.tolerations[].effect

`string` · optional (explicit presence)

`effect` indicates the taint effect to match. Empty means match all taint
effects. When specified, allowed values are NoSchedule, PreferNoSchedule
and NoExecute.

- rule: {"string":{"in":["NoSchedule","PreferNoSchedule","NoExecute"]}}

### spec.podTemplate.spec.tolerations[].tolerationSeconds

`int64` · optional (explicit presence)

`tolerationSeconds` represents the period of time the toleration (which
must be of effect NoExecute, otherwise this field is ignored) tolerates the
taint. By default, it is not set, which means tolerate the taint forever
(do not evict). Zero and negative values will be treated as 0 (evict
immediately) by the system.

### spec.podTemplate.spec.hostAliases

`[]KubernetesCoreV1ApiHostAlias`

`hostAliases` is an optional list of hosts and IPs that will be injected
into the pod's hosts file if specified.

### spec.podTemplate.spec.hostAliases[].ip

`string` · required

`ip` address of the host file entry.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.hostAliases[].hostnames

`[]string`

`hostnames` for the above IP address.

### spec.podTemplate.spec.priorityClassName

`string | valueFrom`

If specified, indicates the pod's priority. "system-node-critical" and
"system-cluster-critical" are two special keywords which indicate the
highest priorities with the former being the highest priority. Any other
name must be defined by creating a PriorityClass object with that name. If
not specified, the pod priority will be default or zero if there is no
default.

Defaults to a KubernetesPriorityClass reference (its priority_class_name
output); pass the literal name with `value:` for one Planton does not
manage.

- references: KubernetesPriorityClass (`status.outputs.priority_class_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesPriorityClass, name: <that resource's name>, fieldPath: status.outputs.priority_class_name}} -- a bare string does not parse

### spec.podTemplate.spec.priority

`int32` · optional (explicit presence)

The priority value. Various system components use this field to find the
priority of the pod. When Priority Admission Controller is enabled, it
prevents users from setting this field. The admission controller populates
this field from PriorityClassName. The higher the value, the higher the
priority.

### spec.podTemplate.spec.dnsConfig

`KubernetesCoreV1ApiPodDNSConfig`

Specifies the DNS parameters of a pod. Parameters specified here will be
merged to the generated DNS configuration based on DNSPolicy.

### spec.podTemplate.spec.dnsConfig.nameservers

`[]string`

A list of DNS name server IP addresses. This will be appended to the base
nameservers generated from DNSPolicy. Duplicated nameservers will be
removed.

### spec.podTemplate.spec.dnsConfig.searches

`[]string`

A list of DNS search domains for host-name lookup. This will be appended to
the base search paths generated from DNSPolicy. Duplicated search paths
will be removed.

### spec.podTemplate.spec.dnsConfig.options

`[]KubernetesCoreV1ApiPodDNSConfigOption`

A list of DNS resolver options. This will be merged with the base options
generated from DNSPolicy. Duplicated entries will be removed. Resolution
options given in Options will override those that appear in the base
DNSPolicy.

### spec.podTemplate.spec.dnsConfig.options[].name

`string`

`name` is this DNS resolver option's name. Required.

### spec.podTemplate.spec.dnsConfig.options[].value

`string` · optional (explicit presence)

`value` is this DNS resolver option's value.

### spec.podTemplate.spec.readinessGates

`[]KubernetesCoreV1ApiPodReadinessGate`

If specified, all readiness gates will be evaluated for pod readiness. A
pod is ready when all its containers are ready AND all conditions specified
in the readiness gates have status equal to "True" More info:
https://git.k8s.io/enhancements/keps/sig-network/580-pod-readiness-gates

### spec.podTemplate.spec.readinessGates[].conditionType

`string` · required

`conditionType` refers to a condition in the pod's condition list with
matching type.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.runtimeClassName

`string` · optional (explicit presence)

`runtimeClassName` refers to a RuntimeClass object in the node.k8s.io
group, which should be used to run this pod.  If no RuntimeClass resource
matches the named class, the pod will not be run. If unset or empty, the
"legacy" RuntimeClass will be used, which is an implicit class with an
empty definition that uses the default runtime handler. More info:
https://git.k8s.io/enhancements/keps/sig-node/585-runtime-class

### spec.podTemplate.spec.enableServiceLinks

`bool` · optional (explicit presence)

`enableServiceLinks` indicates whether information about services should be
injected into pod's environment variables, matching the syntax of Docker
links. Optional: Defaults to true.

### spec.podTemplate.spec.preemptionPolicy

`string` · optional (explicit presence)

`preemptionPolicy` is the Policy for preempting pods with lower priority.
One of Never, PreemptLowerPriority. When Priority Admission Controller is
enabled, it prevents users from setting this field. The admission
controller populates this field from PriorityClassName. Defaults to
PreemptLowerPriority if unset.

- rule: {"string":{"in":["PreemptLowerPriority","Never"]}}

### spec.podTemplate.spec.overhead

`map<string, string>`

`overhead` represents the resource overhead associated with running a pod
for a given RuntimeClass. This field will be autopopulated at admission
time by the RuntimeClass admission controller. If the RuntimeClass
admission controller is enabled, overhead must not be set in Pod create
requests. The RuntimeClass admission controller will reject Pod create
requests which have the overhead already set. If RuntimeClass is configured
and selected in the PodSpec, Overhead will be set to the value defined in
the corresponding RuntimeClass, otherwise it will remain unset and treated
as zero. More info:
https://git.k8s.io/enhancements/keps/sig-node/688-pod-overhead/README.md

- rule: {"map":{"values":{"string":{"pattern":"^(\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))))?$"}}}}

### spec.podTemplate.spec.topologySpreadConstraints

`[]KubernetesCoreV1ApiTopologySpreadConstraint`

`topologySpreadConstraints` describes how a group of pods ought to spread
across topology domains. Scheduler will schedule pods in a way which abides
by the constraints. All topologySpreadConstraints are ANDed.

### spec.podTemplate.spec.topologySpreadConstraints[].maxSkew

`int32` · required · optional (explicit presence)

`maxSkew` describes the degree to which pods may be unevenly distributed.
When `whenUnsatisfiable=DoNotSchedule`, it is the maximum permitted
difference between the number of matching pods in the target topology and
the global minimum. The global minimum is the minimum number of matching
pods in an eligible domain or zero if the number of eligible domains is
less than MinDomains. For example, in a 3-zone cluster, MaxSkew is set to
1, and pods with the same labelSelector spread as 2/2/1: In this case, the
global minimum is 1. | zone1 | zone2 | zone3 | |  P P  |  P P  |   P   | -
if MaxSkew is 1, incoming pod can only be scheduled to zone3 to become
2/2/2; scheduling it onto zone1(zone2) would make the ActualSkew(3-1) on
zone1(zone2) violate MaxSkew(1). - if MaxSkew is 2, incoming pod can be
scheduled onto any zone. When `whenUnsatisfiable=ScheduleAnyway`, it is
used to give higher precedence to topologies that satisfy it. It's a
required field. Default value is 1 and 0 is not allowed.

- rule: {"required":true}

### spec.podTemplate.spec.topologySpreadConstraints[].topologyKey

`string` · required

`topologyKey` is the key of node labels. Nodes that have a label with this
key and identical values are considered to be in the same topology. We
consider each <key, value> as a "bucket", and try to put balanced number of
pods into each bucket. We define a domain as a particular instance of a
topology. Also, we define an eligible domain as a domain whose nodes meet
the requirements of nodeAffinityPolicy and nodeTaintsPolicy. e.g. If
TopologyKey is "kubernetes.io/hostname", each Node is a domain of that
topology. And, if TopologyKey is "topology.kubernetes.io/zone", each zone
is a domain of that topology. It's a required field.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.topologySpreadConstraints[].whenUnsatisfiable

`string`

`whenUnsatisfiable` indicates how to deal with a pod if it doesn't satisfy
the spread constraint. - DoNotSchedule (default) tells the scheduler not to
schedule it. - ScheduleAnyway tells the scheduler to schedule the pod in
any location, but giving higher precedence to topologies that would help
reduce the skew. A constraint is considered "Unsatisfiable" for an incoming
pod if and only if every possible node assignment for that pod would
violate "MaxSkew" on some topology. For example, in a 3-zone cluster,
MaxSkew is set to 1, and pods with the same labelSelector spread as 3/1/1:
| zone1 | zone2 | zone3 | | P P P |   P   |   P   | If WhenUnsatisfiable is
set to DoNotSchedule, incoming pod can only be scheduled to zone2(zone3) to
become 3/2/1(3/1/2) as ActualSkew(2-1) on zone2(zone3) satisfies
MaxSkew(1). In other words, the cluster can still be imbalanced, but
scheduler won't make it *more* imbalanced. It's a required field.

- rule: {"string":{"in":["DoNotSchedule","ScheduleAnyway"]}}

### spec.podTemplate.spec.topologySpreadConstraints[].labelSelector

`KubernetesCoreV1ApiLabelSelector`

`labelSelector` is used to find matching pods. Pods that match this label
selector are counted to determine the number of pods in their corresponding
topology domain.

### spec.podTemplate.spec.topologySpreadConstraints[].labelSelector.matchLabels

`map<string, string>`

matchLabels is a map of {key,value} pairs. A single {key,value} in the
matchLabels map is equivalent to an element of matchExpressions, whose key
field is "key", the operator is "In", and the values array contains only
"value". The requirements are ANDed.

### spec.podTemplate.spec.topologySpreadConstraints[].labelSelector.matchExpressions

`[]KubernetesCoreV1ApiLabelSelectorRequirement`

matchExpressions is a list of label selector requirements. The requirements
are ANDed.

- rule: In and NotIn need at least one value; Exists and DoesNotExist take none

### spec.podTemplate.spec.topologySpreadConstraints[].labelSelector.matchExpressions[].key

`string` · required

key is the label key that the selector applies to.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.topologySpreadConstraints[].labelSelector.matchExpressions[].operator

`string` · required

operator represents a key's relationship to a set of values. Valid
operators are In, NotIn, Exists and DoesNotExist.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.topologySpreadConstraints[].labelSelector.matchExpressions[].values

`[]string`

values is an array of string values. If the operator is In or NotIn, the
values array must be non-empty. If the operator is Exists or DoesNotExist,
the values array must be empty. This array is replaced during a strategic
merge patch.

### spec.podTemplate.spec.topologySpreadConstraints[].minDomains

`int32` · optional (explicit presence)

`minDomains` indicates a minimum number of eligible domains. When the
number of eligible domains with matching topology keys is less than
minDomains, Pod Topology Spread treats "global minimum" as 0, and then the
calculation of Skew is performed. And when the number of eligible domains
with matching topology keys equals or greater than minDomains, this value
has no effect on scheduling. As a result, when the number of eligible
domains is less than minDomains, scheduler won't schedule more than maxSkew
Pods to those domains. If value is nil, the constraint behaves as if
MinDomains is equal to 1. Valid values are integers greater than 0. When
value is not nil, WhenUnsatisfiable must be DoNotSchedule.

For example, in a 3-zone cluster, MaxSkew is set to 2, MinDomains is set to
5 and pods with the same labelSelector spread as 2/2/2: | zone1 | zone2 |
zone3 | |  P P  |  P P  |  P P  | The number of domains is less than
5(MinDomains), so "global minimum" is treated as 0. In this situation, new
pod with the same labelSelector cannot be scheduled, because computed skew
will be 3(3 - 0) if new Pod is scheduled to any of the three zones, it will
violate MaxSkew.

### spec.podTemplate.spec.topologySpreadConstraints[].nodeAffinityPolicy

`string` · optional (explicit presence)

`nodeAffinityPolicy` indicates how we will treat Pod's
nodeAffinity/nodeSelector when calculating pod topology spread skew.
Options are: - Honor: only nodes matching nodeAffinity/nodeSelector are
included in the calculations. - Ignore: nodeAffinity/nodeSelector are
ignored. All nodes are included in the calculations.

If this value is nil, the behavior is equivalent to the Honor policy.

- rule: {"string":{"in":["Ignore","Honor"]}}

### spec.podTemplate.spec.topologySpreadConstraints[].nodeTaintsPolicy

`string` · optional (explicit presence)

`nodeTaintsPolicy` indicates how we will treat node taints when calculating
pod topology spread skew. Options are: - Honor: nodes without taints, along
with tainted nodes for which the incoming pod has a toleration, are
included. - Ignore: node taints are ignored. All nodes are included.

If this value is nil, the behavior is equivalent to the Ignore policy.

- rule: {"string":{"in":["Ignore","Honor"]}}

### spec.podTemplate.spec.topologySpreadConstraints[].matchLabelKeys

`[]string`

`matchLabelKeys` is a set of pod label keys to select the pods over which
spreading will be calculated. The keys are used to lookup values from the
incoming pod labels, those key-value labels are ANDed with labelSelector to
select the group of existing pods over which spreading will be calculated
for the incoming pod. The same key is forbidden to exist in both
MatchLabelKeys and LabelSelector. MatchLabelKeys cannot be set when
LabelSelector isn't set. Keys that don't exist in the incoming pod labels
will be ignored. A null or empty list means only match against
labelSelector.

This is a beta field and requires the MatchLabelKeysInPodTopologySpread
feature gate to be enabled (enabled by default).

### spec.podTemplate.spec.setHostnameAsFQDN

`bool` · optional (explicit presence)

If true the pod's hostname will be configured as the pod's FQDN, rather
than the leaf name (the default). In Linux containers, this means setting
the FQDN in the hostname field of the kernel (the nodename field of struct
utsname). In Windows containers, this means setting the registry value of
hostname for the registry key
HKEY_LOCAL_MACHINE\\SYSTEM\\CurrentControlSet\\Services\\Tcpip\\Parameters
to FQDN. If a pod does not have FQDN, this has no effect. Default to false.

### spec.podTemplate.spec.os

`KubernetesCoreV1ApiPodOS`

Specifies the OS of the containers in the pod. Some pod and container
fields are restricted if this is set.

If the OS field is set to linux, the following fields must be unset:
-securityContext.windowsOptions.

If the OS field is set to windows, following fields must be unset: -
spec.hostPID - spec.hostIPC - spec.hostUsers - spec.resources -
spec.securityContext.appArmorProfile - spec.securityContext.seLinuxOptions
- spec.securityContext.seccompProfile - spec.securityContext.fsGroup -
spec.securityContext.fsGroupChangePolicy - spec.securityContext.sysctls -
spec.shareProcessNamespace - spec.securityContext.runAsUser -
spec.securityContext.runAsGroup - spec.securityContext.supplementalGroups -
spec.securityContext.supplementalGroupsPolicy -
spec.containers[*].securityContext.appArmorProfile -
spec.containers[*].securityContext.seLinuxOptions -
spec.containers[*].securityContext.seccompProfile -
spec.containers[*].securityContext.capabilities -
spec.containers[*].securityContext.readOnlyRootFilesystem -
spec.containers[*].securityContext.privileged -
spec.containers[*].securityContext.allowPrivilegeEscalation -
spec.containers[*].securityContext.procMount -
spec.containers[*].securityContext.runAsUser -
spec.containers[*].securityContext.runAsGroup.

### spec.podTemplate.spec.os.name

`string` · required

`name` is the name of the operating system. The currently supported values
are linux and windows. Additional value may be defined in future and can be
one of:
https://github.com/opencontainers/runtime-spec/blob/master/config.md#platform-specific-configuration
Clients should expect to handle additional values and treat unrecognized
values in this field as os: null.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.hostUsers

`bool` · optional (explicit presence)

Use the host's user namespace. Optional: Default to true. If set to true or
not present, the pod will be run in the host user namespace, useful for
when the pod needs a feature only available to the host user namespace,
such as loading a kernel module with CAP_SYS_MODULE. When set to false, a
new userns is created for the pod. Setting false is useful for mitigating
container breakout vulnerabilities even allowing users to run their
containers as root without actually having root privileges on the host.

### spec.podTemplate.spec.schedulingGates

`[]KubernetesCoreV1ApiPodSchedulingGate`

`schedulingGates` is an opaque list of values that if specified will block
scheduling the pod. If schedulingGates is not empty, the pod will stay in
the SchedulingGated state and the scheduler will not attempt to schedule
the pod.

SchedulingGates can only be set at pod creation time, and be removed only
afterwards.

### spec.podTemplate.spec.schedulingGates[].name

`string` · required

`name` of the scheduling gate. Each scheduling gate must have a unique name
field.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.resourceClaims

`[]KubernetesCoreV1ApiPodResourceClaim`

`resourceClaims` defines which ResourceClaims must be allocated and
reserved before the Pod is allowed to start. The resources will be made
available to those containers which consume them by name.

This is a stable field but requires that the DynamicResourceAllocation
feature gate is enabled.

This field is immutable.

### spec.podTemplate.spec.resourceClaims[].name

`string` · required

`name` uniquely identifies this resource claim inside the pod. This must be
a DNS_LABEL.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.resourceClaims[].resourceClaimName

`string` · optional (explicit presence)

`resourceClaimName` is the name of a ResourceClaim object in the same
namespace as this pod.

Exactly one of ResourceClaimName and ResourceClaimTemplateName must be set.

### spec.podTemplate.spec.resourceClaims[].resourceClaimTemplateName

`string` · optional (explicit presence)

`resourceClaimTemplateName` is the name of a ResourceClaimTemplate object
in the same namespace as this pod.

The template will be used to create a new ResourceClaim, which will be
bound to this pod. When this pod is deleted, the ResourceClaim will also be
deleted. The pod name and resource name, along with a generated component,
will be used to form a unique name for the ResourceClaim, which will be
recorded in pod.status.resourceClaimStatuses.

When the DRAWorkloadResourceClaims feature gate is enabled and the pod
belongs to a PodGroup that defines a PodGroupResourceClaim with the same
Name and ResourceClaimTemplateName, this PodResourceClaim resolves to the
ResourceClaim generated for the PodGroup. All pods in the group that define
an equivalent PodResourceClaim matching the PodGroupResourceClaim's Name
and ResourceClaimTemplateName share the same generated ResourceClaim.
ResourceClaims generated for a PodGroup are owned by the PodGroup and their
lifecycles are tied to the PodGroup instead of any individual pod.

This field is immutable and no changes will be made to the corresponding
ResourceClaim by the control plane after creating the ResourceClaim.

Exactly one of ResourceClaimName and ResourceClaimTemplateName must be set.

### spec.podTemplate.spec.resources

`KubernetesCoreV1ApiResourceRequirements`

`resources` is the total amount of CPU and Memory resources required by all
containers in the pod. It supports specifying Requests and Limits for
"cpu", "memory" and "hugepages-" resource names only. ResourceClaims are
not supported.

This field enables fine-grained control over resource allocation for the
entire pod, allowing resource sharing among containers in a pod.

This is an alpha field and requires enabling the PodLevelResources feature
gate.

### spec.podTemplate.spec.resources.limits

`map<string, string>`

`limits` describes the maximum amount of compute resources allowed. More
info:
https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/

- rule: {"map":{"values":{"string":{"pattern":"^(\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))))?$"}}}}

### spec.podTemplate.spec.resources.requests

`map<string, string>`

`requests` describes the minimum amount of compute resources required. If
Requests is omitted for a container, it defaults to Limits if that is
explicitly specified, otherwise to an implementation-defined value.
Requests cannot exceed Limits. More info:
https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/

- rule: {"map":{"values":{"string":{"pattern":"^(\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))))?$"}}}}

### spec.podTemplate.spec.resources.claims

`[]KubernetesCoreV1ApiResourceClaim`

`claims` lists the names of resources, defined in spec.resourceClaims, that
are used by this container.

This field depends on the DynamicResourceAllocation feature gate.

This field is immutable. It can only be set for containers.

### spec.podTemplate.spec.resources.claims[].name

`string` · required

`name` must match the name of one entry in pod.spec.resourceClaims of the
Pod where this field is used. It makes that resource available inside a
container.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.resources.claims[].request

`string`

`request` is the name chosen for a request in the referenced claim. If
empty, everything from the claim is made available, otherwise only the
result of this request.

### spec.podTemplate.spec.hostnameOverride

`string` · optional (explicit presence)

`hostnameOverride` specifies an explicit override for the pod's hostname as
perceived by the pod. This field only specifies the pod's hostname and does
not affect its DNS records. When this field is set to a non-empty string: -
It takes precedence over the values set in `hostname` and `subdomain`. -
The Pod's hostname will be set to this value. - `setHostnameAsFQDN` must be
nil or set to false. - `hostNetwork` must be set to false.

This field must be a valid DNS subdomain as defined in RFC 1123 and contain
at most 64 characters.

Requires the HostnameOverride feature gate on the cluster; a cluster
without it drops this field from the pod.

### spec.podTemplate.spec.schedulingGroup

`KubernetesCoreV1ApiPodSchedulingGroup`

`schedulingGroup` provides a reference to the immediate scheduling runtime
grouping object that this Pod belongs to. This field is used by the
scheduler to identify the group and apply the correct group scheduling
policies. The association with a group also impacts other lifecycle aspects
of a Pod that are relevant in a wider context of scheduling like
preemption, resource attachment, etc. If not specified, the Pod is treated
as a single unit in all of these aspects. The group object referenced by
this field may not exist at the time the Pod is created. This field is
immutable, but a group object with the same name may be recreated with
different policies. Doing this during pod scheduling may result in the
placement not conforming to the expected policies.

Requires the GenericWorkload feature gate on the cluster; a cluster without
it drops this field from the pod.

### spec.podTemplate.spec.schedulingGroup.podGroupName

`string` · optional (explicit presence)

`podGroupName` specifies the name of the standalone PodGroup object that
represents the runtime instance of this group. Must be a DNS subdomain.

### spec.podTemplate.spec.evictionResponders

`[]KubernetesCoreV1ApiEvictionResponder`

evictionResponders reference responders that react to Evictions based on
EvictionRequests. Responders should observe and communicate through the
Eviction Resource API to help with the graceful termination of a pod. The
responders are selected sequentially, according to their specified
priority.

Responders should periodically report on an eviction progress by updating
the .status.responders[].heartbeatTime field of the Eviction object. If
this field is not updated within the heartbeat deadline defined by the
Eviction API (currently 20 minutes), the eviction is passed over to the
next responder with a lower priority. If there is no other responder, the
last default imperative-eviction.k8s.io/evictor responder with a priority
of 100 will evict the pod using the imperative Eviction API
(pods/<name>/eviction subresource).

The maximum length of the responders list is 10. Responders are not
supported when the pod is part of a PodGroup (.spec.schedulingGroup is
set). This field can only be set on creation and is immutable afterwards.

Requires the EvictionRequestAPI feature gate on the cluster; a cluster
without it drops this field from the pod.

### spec.podTemplate.spec.evictionResponders[].name

`string` · required

name allows you to identify the responder responding to the Eviction.

It must be a valid domain-prefixed key (such as "acme.io/foo"). Domain
names *.k8s.io and *.kubernetes.io are reserved. This field must be unique
for each responder. This field is required.

- rule: {"string":{"minLen":"1"}}

### spec.podTemplate.spec.evictionResponders[].priority

`int32` · required · optional (explicit presence)

priority for this responder. Higher priorities are selected first by the
evictionrequest-controller. If there are responders with the same priority,
the responder whose domain name comes first in the alphabetical higher
domain order, will be picked. This means that the top domain labels are
compared alphabetically first, followed by the lower domain labels. The key
is compared last.

The responder that is the managing controller of the pod should set the
value of this field to 10000 to allow both for preemption or fallback
registration by other responders.

The minimum value is 0 and the maximum value is 100000. The interval 0-999
is reserved for responders with *.k8s.io suffix. This field is required.

- rule: {"required":true}

### spec.podTemplate.metadata

`KubernetesAgentSandboxApiPodMetadata`

The labels and annotations the pod carries. Only these two are used;
the controller names the pod after the sandbox.

### spec.podTemplate.metadata.labels

`map<string, string>`

Labels on the pod, which Services, NetworkPolicies and monitors select it
by. More info:
https://kubernetes.io/docs/concepts/overview/working-with-objects/labels

- rule: {"map":{"keys":{"string":{"minLen":"1","maxLen":"317","pattern":"^([a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*/)?[A-Za-z0-9]([-A-Za-z0-9_.]{0,61}[A-Za-z0-9])?$"}},"values":{"string":{"maxLen":"63","pattern":"^([A-Za-z0-9]([-A-Za-z0-9_.]{0,61}[A-Za-z0-9])?)?$"}}}}

### spec.podTemplate.metadata.annotations

`map<string, string>`

Annotations on the pod: metadata for tools, such as a service mesh's
sidecar switch or a scrape hint. More info:
https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations

### spec.volumeClaimTemplates

`[]KubernetesAgentSandboxApiPersistentVolumeClaimTemplate`

Volumes the sandbox owns. The controller creates one
PersistentVolumeClaim per template, named `<template name>-<sandbox
name>`, and the pod mounts it by the template's name (a volume with that
name is added to the pod; mount it with a `volumeMounts` entry). The
claims belong to the sandbox: they outlive its pod, so a suspended or
restarted sandbox keeps its files, and they go when the sandbox is
deleted.

Fixed once the sandbox exists: the API server refuses any change to this
list (add, remove or edit), so changing a sandbox's volumes means
replacing the sandbox.

- rule: Each volume claim template needs a unique name: the pod mounts each claim by its template's name

### spec.volumeClaimTemplates[].metadata

`KubernetesAgentSandboxApiEmbeddedObjectMetadata` · required

The claim's name, labels and annotations.

- rule: {"required":true}

### spec.volumeClaimTemplates[].metadata.name

`string` · required

The template's name: the pod's volume name (a `volumeMounts` entry names
it) and the first half of the claim's name. Upstream marks it optional,
but the controller names the pod volume and the claim after it, so a
template without one produces no usable volume; it is required here.

- rule: {"string":{"minLen":"1","maxLen":"63","pattern":"^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"}}

### spec.volumeClaimTemplates[].metadata.labels

`map<string, string>`

Labels on the claim. More info:
https://kubernetes.io/docs/concepts/overview/working-with-objects/labels

- rule: {"map":{"keys":{"string":{"minLen":"1","maxLen":"317","pattern":"^([a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*/)?[A-Za-z0-9]([-A-Za-z0-9_.]{0,61}[A-Za-z0-9])?$"}},"values":{"string":{"maxLen":"63","pattern":"^([A-Za-z0-9]([-A-Za-z0-9_.]{0,61}[A-Za-z0-9])?)?$"}}}}

### spec.volumeClaimTemplates[].metadata.annotations

`map<string, string>`

Annotations on the claim, such as a storage provisioner's options. More
info:
https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations

### spec.volumeClaimTemplates[].spec

`KubernetesCoreV1ApiPersistentVolumeClaimSpec` · required

The claim's specification: access modes, size, storage class.

- rule: {"required":true}

### spec.volumeClaimTemplates[].spec.accessModes

`[]string`

accessModes contains the desired access modes the volume should have. More
info:
https://kubernetes.io/docs/concepts/storage/persistent-volumes#access-modes-1

- rule: {"repeated":{"items":{"string":{"in":["ReadWriteOnce","ReadOnlyMany","ReadWriteMany","ReadWriteOncePod"]}}}}

### spec.volumeClaimTemplates[].spec.selector

`KubernetesCoreV1ApiLabelSelector`

selector is a label query over volumes to consider for binding.

### spec.volumeClaimTemplates[].spec.selector.matchLabels

`map<string, string>`

matchLabels is a map of {key,value} pairs. A single {key,value} in the
matchLabels map is equivalent to an element of matchExpressions, whose key
field is "key", the operator is "In", and the values array contains only
"value". The requirements are ANDed.

### spec.volumeClaimTemplates[].spec.selector.matchExpressions

`[]KubernetesCoreV1ApiLabelSelectorRequirement`

matchExpressions is a list of label selector requirements. The requirements
are ANDed.

- rule: In and NotIn need at least one value; Exists and DoesNotExist take none

### spec.volumeClaimTemplates[].spec.selector.matchExpressions[].key

`string` · required

key is the label key that the selector applies to.

- rule: {"string":{"minLen":"1"}}

### spec.volumeClaimTemplates[].spec.selector.matchExpressions[].operator

`string` · required

operator represents a key's relationship to a set of values. Valid
operators are In, NotIn, Exists and DoesNotExist.

- rule: {"string":{"minLen":"1"}}

### spec.volumeClaimTemplates[].spec.selector.matchExpressions[].values

`[]string`

values is an array of string values. If the operator is In or NotIn, the
values array must be non-empty. If the operator is Exists or DoesNotExist,
the values array must be empty. This array is replaced during a strategic
merge patch.

### spec.volumeClaimTemplates[].spec.resources

`KubernetesCoreV1ApiVolumeResourceRequirements`

resources represents the minimum resources the volume should have. Users
are allowed to specify resource requirements that are lower than previous
value but must still be higher than capacity recorded in the status field
of the claim. More info:
https://kubernetes.io/docs/concepts/storage/persistent-volumes#resources

### spec.volumeClaimTemplates[].spec.resources.limits

`map<string, string>`

`limits` describes the maximum amount of compute resources allowed. More
info:
https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/

- rule: {"map":{"values":{"string":{"pattern":"^(\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))))?$"}}}}

### spec.volumeClaimTemplates[].spec.resources.requests

`map<string, string>`

`requests` describes the minimum amount of compute resources required. If
Requests is omitted for a container, it defaults to Limits if that is
explicitly specified, otherwise to an implementation-defined value.
Requests cannot exceed Limits. More info:
https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/

- rule: {"map":{"values":{"string":{"pattern":"^(\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\\+|-)?(([0-9]+(\\.[0-9]*)?)|(\\.[0-9]+))))?$"}}}}

### spec.volumeClaimTemplates[].spec.volumeName

`string`

volumeName is the binding reference to the PersistentVolume backing this
claim.

### spec.volumeClaimTemplates[].spec.storageClassName

`string | valueFrom`

storageClassName is the name of the StorageClass required by the claim.
More info:
https://kubernetes.io/docs/concepts/storage/persistent-volumes#class-1

Defaults to a KubernetesStorageClass reference (its storage_class_name
output); pass the literal name with `value:` for one Planton does not
manage.

- references: KubernetesStorageClass (`status.outputs.storage_class_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesStorageClass, name: <that resource's name>, fieldPath: status.outputs.storage_class_name}} -- a bare string does not parse

### spec.volumeClaimTemplates[].spec.volumeMode

`string` · optional (explicit presence)

volumeMode defines what type of volume is required by the claim. Value of
Filesystem is implied when not included in claim spec.

- rule: {"string":{"in":["Block","Filesystem"]}}

### spec.volumeClaimTemplates[].spec.dataSource

`KubernetesCoreV1ApiTypedLocalObjectReference`

dataSource field can be used to specify either: * An existing
VolumeSnapshot object (snapshot.storage.k8s.io/VolumeSnapshot) * An
existing PVC (PersistentVolumeClaim) If the provisioner or an external
controller can support the specified data source, it will create a new
volume based on the contents of the specified data source. dataSource
contents will be copied to dataSourceRef, and dataSourceRef contents will
be copied to dataSource when dataSourceRef.namespace is not specified. If
the namespace is specified, then dataSourceRef will not be copied to
dataSource.

### spec.volumeClaimTemplates[].spec.dataSource.apiGroup

`string` · optional (explicit presence)

`apiGroup` is the group for the resource being referenced. If APIGroup is
not specified, the specified Kind must be in the core API group. For any
other third-party types, APIGroup is required.

### spec.volumeClaimTemplates[].spec.dataSource.kind

`string` · required

`kind` is the type of resource being referenced.

- rule: {"string":{"minLen":"1"}}

### spec.volumeClaimTemplates[].spec.dataSource.name

`string` · required

`name` is the name of resource being referenced.

- rule: {"string":{"minLen":"1"}}

### spec.volumeClaimTemplates[].spec.dataSourceRef

`KubernetesCoreV1ApiTypedObjectReference`

dataSourceRef specifies the object from which to populate the volume with
data, if a non-empty volume is desired. This may be any object from a
non-empty API group (non core object) or a PersistentVolumeClaim object.
When this field is specified, volume binding will only succeed if the type
of the specified object matches some installed volume populator or dynamic
provisioner. This field will replace the functionality of the dataSource
field and as such if both fields are non-empty, they must have the same
value. For backwards compatibility, when namespace isn't specified in
dataSourceRef, both fields (dataSource and dataSourceRef) will be set to
the same value automatically if one of them is empty and the other is
non-empty. When namespace is specified in dataSourceRef, dataSource isn't
set to the same value and must be empty. There are three important
differences between dataSource and dataSourceRef: * While dataSource only
allows two specific types of objects, dataSourceRef allows any non-core
object, as well as PersistentVolumeClaim objects. * While dataSource
ignores disallowed values (dropping them), dataSourceRef preserves all
values, and generates an error if a disallowed value is specified. * While
dataSource only allows local objects, dataSourceRef allows objects in any
namespaces. (Alpha) Using the namespace field of dataSourceRef requires the
CrossNamespaceVolumeDataSource feature gate to be enabled.

### spec.volumeClaimTemplates[].spec.dataSourceRef.apiGroup

`string` · optional (explicit presence)

`apiGroup` is the group for the resource being referenced. If APIGroup is
not specified, the specified Kind must be in the core API group. For any
other third-party types, APIGroup is required.

### spec.volumeClaimTemplates[].spec.dataSourceRef.kind

`string` · required

`kind` is the type of resource being referenced.

- rule: {"string":{"minLen":"1"}}

### spec.volumeClaimTemplates[].spec.dataSourceRef.name

`string` · required

`name` is the name of resource being referenced.

- rule: {"string":{"minLen":"1"}}

### spec.volumeClaimTemplates[].spec.dataSourceRef.namespace

`string` · optional (explicit presence)

`namespace` is the namespace of resource being referenced Note that when a
namespace is specified, a gateway.networking.k8s.io/ReferenceGrant object
is required in the referent namespace to allow that namespace's owner to
accept the reference. See the ReferenceGrant documentation for details.
(Alpha) This field requires the CrossNamespaceVolumeDataSource feature gate
to be enabled.

### spec.volumeClaimTemplates[].spec.volumeAttributesClassName

`string` · optional (explicit presence)

volumeAttributesClassName may be used to set the VolumeAttributesClass used
by this claim. If specified, the CSI driver will create or update the
volume with the attributes defined in the corresponding
VolumeAttributesClass. This has a different purpose than storageClassName,
it can be changed after the claim is created. An empty string or nil value
indicates that no VolumeAttributesClass will be applied to the claim. If
the claim enters an Infeasible error state, this field can be reset to its
previous value (including nil) to cancel the modification. If the resource
referred to by volumeAttributesClass does not exist, this
PersistentVolumeClaim will be set to a Pending state, as reflected by the
modifyVolumeStatus field, until such as a resource exists. More info:
https://kubernetes.io/docs/concepts/storage/volume-attributes-classes/

### spec.service

`bool` · optional (explicit presence)

Whether the controller keeps a headless Service named after the sandbox,
which gives the pod the stable DNS name `<sandbox>.<namespace>.svc` (the
router falls back to it when it has no pod IP cached). True creates it,
false removes it, and unset keeps a Service that already exists without
creating one.

### spec.shutdownTime

`string` · optional (explicit presence)

When the sandbox expires, as an RFC 3339 time ("2026-12-31T18:00:00Z").
At that moment the controller deletes the pod and the Service and then
applies `shutdown_policy` to the Sandbox object. Unset, the sandbox never
expires. The time is absolute: a declaration whose time has passed
creates a sandbox that is torn down at once.

- rule: {"string":{"pattern":"^[0-9]{4}-[0-9]{2}-[0-9]{2}[Tt][0-9]{2}:[0-9]{2}:[0-9]{2}(\\.[0-9]+)?([Zz]|[+-][0-9]{2}:[0-9]{2})$"}}

### spec.shutdownPolicy

`string` · optional (explicit presence)

What happens to the Sandbox object when it expires (the pod and Service
are deleted either way): "Retain" keeps the object, with a Ready=False
condition whose reason is SandboxExpired, and "Delete" deletes it. Unset,
the cluster writes "Retain". "Delete" removes an object Planton still
declares, so the next apply recreates it; pair it with a sandbox created
for a bounded job.

- rule: {"string":{"in":["Delete","Retain"]}}

### spec.operatingMode

`string` · optional (explicit presence)

The state the sandbox should be in: "Running" keeps its pod running, and
"Suspended" deletes the pod while keeping the Sandbox object, its claims
(and so its files) and its Service, so a later "Running" brings it back
where it left off. The Ready and Suspended conditions report what the
controller has reached. Unset, the cluster writes "Running".

- rule: {"string":{"in":["Running","Suspended"]}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesAgentSandbox, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.sandbox_name` | `string` | Name of the created Sandbox (equals metadata.name); its pod and headless Service carry the same name. |
| `status.outputs.namespace` | `string` | Namespace the Sandbox was created in (the resolved spec.namespace). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.podTemplate.spec.volumes[].secret.secretName` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.podTemplate.spec.volumes[].iscsi.secretRef.name` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.podTemplate.spec.volumes[].persistentVolumeClaim.claimName` | KubernetesPersistentVolumeClaim | `status.outputs.pvc_name` |
| `spec.podTemplate.spec.volumes[].rbd.secretRef.name` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.podTemplate.spec.volumes[].flexVolume.secretRef.name` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.podTemplate.spec.volumes[].cinder.secretRef.name` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.podTemplate.spec.volumes[].cephfs.secretRef.name` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.podTemplate.spec.volumes[].azureFile.secretName` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.podTemplate.spec.volumes[].configMap.name` | KubernetesConfigMap | `status.outputs.configmap_name` |
| `spec.podTemplate.spec.volumes[].projected.sources[].secret.name` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.podTemplate.spec.volumes[].projected.sources[].configMap.name` | KubernetesConfigMap | `status.outputs.configmap_name` |
| `spec.podTemplate.spec.volumes[].scaleIO.secretRef.name` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.podTemplate.spec.volumes[].storageos.secretRef.name` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.podTemplate.spec.volumes[].csi.nodePublishSecretRef.name` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.podTemplate.spec.volumes[].ephemeral.volumeClaimTemplate.spec.storageClassName` | KubernetesStorageClass | `status.outputs.storage_class_name` |
| `spec.podTemplate.spec.initContainers[].envFrom[].configMapRef.name` | KubernetesConfigMap | `status.outputs.configmap_name` |
| `spec.podTemplate.spec.initContainers[].envFrom[].secretRef.name` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.podTemplate.spec.initContainers[].env[].valueFrom.configMapKeyRef.name` | KubernetesConfigMap | `status.outputs.configmap_name` |
| `spec.podTemplate.spec.initContainers[].env[].valueFrom.secretKeyRef.name` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.podTemplate.spec.containers[].envFrom[].configMapRef.name` | KubernetesConfigMap | `status.outputs.configmap_name` |
| `spec.podTemplate.spec.containers[].envFrom[].secretRef.name` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.podTemplate.spec.containers[].env[].valueFrom.configMapKeyRef.name` | KubernetesConfigMap | `status.outputs.configmap_name` |
| `spec.podTemplate.spec.containers[].env[].valueFrom.secretKeyRef.name` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.podTemplate.spec.ephemeralContainers[].envFrom[].configMapRef.name` | KubernetesConfigMap | `status.outputs.configmap_name` |
| `spec.podTemplate.spec.ephemeralContainers[].envFrom[].secretRef.name` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.podTemplate.spec.ephemeralContainers[].env[].valueFrom.configMapKeyRef.name` | KubernetesConfigMap | `status.outputs.configmap_name` |
| `spec.podTemplate.spec.ephemeralContainers[].env[].valueFrom.secretKeyRef.name` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.podTemplate.spec.serviceAccountName` | KubernetesServiceAccount | `status.outputs.service_account_name` |
| `spec.podTemplate.spec.imagePullSecrets[].name` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.podTemplate.spec.priorityClassName` | KubernetesPriorityClass | `status.outputs.priority_class_name` |
| `spec.volumeClaimTemplates[].spec.storageClassName` | KubernetesStorageClass | `status.outputs.storage_class_name` |

## See Also

- [Overview](../README.md)

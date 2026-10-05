# KubernetesFlagd Pulumi Module

Runs flagd as module-owned manifests (flagd publishes no Helm chart): the optional namespace, the `<name>-sources` Secret, the ServiceAccount, the FeatureFlag reader Roles, the Deployment and Service, and the optional HorizontalPodAutoscaler, PodDisruptionBudget and ServiceMonitor. Pod scheduling and security render through the platform's shared workload pod builder, the same code every Kubernetes workload kind uses.

## What the Module Creates

1. **Namespace** (optional) -- created with the standard governance labels when `create_namespace` is true (`namespace.go`)
2. **`<name>-sources` Secret** -- flagd's SourceConfig array under the data key `sources`, marked secret in Pulumi state (`identity.go`)
3. **ServiceAccount `<name>`** -- unless `service_account.existing_name` names one (`identity.go`)
4. **`<name>-flag-reader` Role and RoleBinding** -- one pair per namespace a `feature_flag` source reads, granting get/list/watch on `featureflags.core.openfeature.dev` (`identity.go`)
5. **Deployment `<name>`** -- flagd with its arguments, `FLAGD_SOURCES` from the Secret, directory-mounted ConfigMap sources, `/healthz` and `/readyz` probes on the management port (`deployment.go`)
6. **Service `<name>`** -- the evaluation, management, sync and OFREP ports (`deployment.go`)
7. **HorizontalPodAutoscaler and PodDisruptionBudget** (optional, `scaling.go`)
8. **ServiceMonitor `<name>-metrics`** (optional, `service_monitor.go`)

## Rendering Notes

- **One renderer for arguments and sources** -- `flagd_config.go` builds the `flagd start` arguments and the sources document; the Terraform twin's `locals.tf` produces the same arguments and a byte-identical document (Go `json.Marshal` and OpenTofu `jsonencode` sort keys and escape alike), so the `checksum/sources` pod annotation matches across engines.
- **A source change rolls the pods** -- a variable read from a Secret is fixed at container start, so the SHA-256 of the sources document is stamped on the pod template.
- **ConfigMap sources are directories** -- each mounts at `/etc/flagd/sources/<index>` and flagd reads `<mount>/<key>`; a `subPath` mount would never see an edit.
- **The name budget fails loudly** -- `Resources()` rejects `metadata.name` past 63 characters: the Service is named after the resource, and a Service name is a 63-character DNS label.
- **A header declared both plain and sensitive is refused** before anything is created.
- **Empty topology-spread selectors self-spread** on flagd's own selector labels (`app.kubernetes.io/name: flagd`, `app.kubernetes.io/instance: <name>`).

## Inputs

`KubernetesFlagdIacInput`: `target` (the KubernetesFlagd manifest) and `provider_config` (the Kubernetes provider configuration).

## Outputs

`namespace`, `service`, `evaluation_endpoint`, `sync_endpoint`, `ofrep_endpoint`, `management_endpoint`, `port_forward_command`.

## Local Development

```shell
planton pulumi preview --manifest ../../e2e/manifest.yaml --module-dir .
planton pulumi up --manifest ../../e2e/manifest.yaml --module-dir .
planton pulumi destroy --manifest ../../e2e/manifest.yaml --module-dir .
```

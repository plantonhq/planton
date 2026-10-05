# GcpDeliveryPipeline

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpDeliveryPipelineSpec declares a Cloud Deploy delivery pipeline
(`google_clouddeploy_delivery_pipeline`) together with its automations
(`google_clouddeploy_automation`): the ordered list of targets a release
is promoted through, how each rollout is carried out, and which
promotions, advances, and repairs Cloud Deploy performs on its own.

A pipeline is a serial promotion flow: a release is created against the
pipeline, rolls out to the first stage's target, and is promoted stage
by stage (dev, then staging, then prod). Each stage names a target -- a
GcpDeployTarget in this pipeline's project and region, by its bare ID --
and optionally a rollout strategy: standard (one deploy, optional
verification and pre/post jobs) or canary (progressive percentages on
Cloud Run or GKE).

Automations are children by path (deliveryPipelines/*/automations/*):
nothing else references them and destroying the pipeline removes them,
so they live here. Each one runs as a service account you name and acts
on the targets its selector picks.

Important behavioral notes:

  - location and delivery_pipeline_id are create-time decisions;
    everything else, including the stages and their strategies, updates
    in place.
  - Releases and rollouts are created outside this block (gcloud deploy
    releases create, Cloud Build, or an automation's promotions). Destroy
    deletes the pipeline with force=true, which ALSO deletes every
    release, rollout, and automation under it -- the provider always
    sends force, there is no option to refuse. Use deletion_policy
    PREVENT to guard a pipeline whose release history matters.
  - The targets need not exist when the pipeline is created (Google
    reports missing ones in the pipeline's condition), but a release
    cannot roll out to a stage whose target is missing.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpDeliveryPipeline
metadata:
  name: web
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  deliveryPipelineId: web
  description: Web service from dev to prod on Cloud Run
  labels:
    team: web
  annotations:
    owner: platform
  serialPipeline:
    stages:
      - targetId:
          value: dev
        profiles:
          - dev
        strategy:
          standard:
            verify: false
      - targetId:
          value: prod
        profiles:
          - prod
        strategy:
          standard:
            verify: false
        deployParameters:
          - values:
              min-instances: "2"
  automations:
    - automationId: promote-to-prod
      description: Promote every release that succeeds on dev to prod
      labels:
        team: web
      serviceAccount:
        value: deploy-automation@my-gcp-project.iam.gserviceaccount.com
      selector:
        targets:
          - id:
              value: dev
      rules:
        - promoteReleaseRule:
            id: promote-to-prod
            wait: 600s
            destinationTargetId:
              value: "@next"
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.location` | `string` | yes |  |  |
| `spec.deliveryPipelineId` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.annotations` | `map<string, string>` |  |  |  |
| `spec.suspended` | `bool` |  |  |  |
| `spec.serialPipeline` | `GcpDeliveryPipelineSerialPipeline` |  |  |  |
| `spec.serialPipeline.stages` | `[]GcpDeliveryPipelineStage` |  |  |  |
| `spec.serialPipeline.stages[].targetId` | `string \| valueFrom` |  |  | GcpDeployTarget (`status.outputs.target_id`) |
| `spec.serialPipeline.stages[].profiles` | `[]string` |  |  |  |
| `spec.serialPipeline.stages[].strategy` | `GcpDeliveryPipelineStrategy` |  |  |  |
| `spec.serialPipeline.stages[].strategy.standard` | `GcpDeliveryPipelineStandard` |  |  |  |
| `spec.serialPipeline.stages[].strategy.standard.verify` | `bool` |  |  |  |
| `spec.serialPipeline.stages[].strategy.standard.predeploy` | `GcpDeliveryPipelineStandardDeployHook` |  |  |  |
| `spec.serialPipeline.stages[].strategy.standard.predeploy.actions` | `[]string` |  |  |  |
| `spec.serialPipeline.stages[].strategy.standard.predeploy.tasks` | `[]GcpDeliveryPipelineTask` |  |  |  |
| `spec.serialPipeline.stages[].strategy.standard.predeploy.tasks[].container` | `GcpDeliveryPipelineContainerTask` |  |  |  |
| `spec.serialPipeline.stages[].strategy.standard.predeploy.tasks[].container.image` | `string` | yes |  |  |
| `spec.serialPipeline.stages[].strategy.standard.predeploy.tasks[].container.command` | `[]string` |  |  |  |
| `spec.serialPipeline.stages[].strategy.standard.predeploy.tasks[].container.args` | `[]string` |  |  |  |
| `spec.serialPipeline.stages[].strategy.standard.predeploy.tasks[].container.env` | `map<string, string>` |  |  |  |
| `spec.serialPipeline.stages[].strategy.standard.postdeploy` | `GcpDeliveryPipelineStandardDeployHook` |  |  |  |
| `spec.serialPipeline.stages[].strategy.standard.postdeploy.actions` | `[]string` |  |  |  |
| `spec.serialPipeline.stages[].strategy.standard.postdeploy.tasks` | `[]GcpDeliveryPipelineTask` |  |  |  |
| `spec.serialPipeline.stages[].strategy.standard.postdeploy.tasks[].container` | `GcpDeliveryPipelineContainerTask` |  |  |  |
| `spec.serialPipeline.stages[].strategy.standard.postdeploy.tasks[].container.image` | `string` | yes |  |  |
| `spec.serialPipeline.stages[].strategy.standard.postdeploy.tasks[].container.command` | `[]string` |  |  |  |
| `spec.serialPipeline.stages[].strategy.standard.postdeploy.tasks[].container.args` | `[]string` |  |  |  |
| `spec.serialPipeline.stages[].strategy.standard.postdeploy.tasks[].container.env` | `map<string, string>` |  |  |  |
| `spec.serialPipeline.stages[].strategy.standard.verifyConfig` | `GcpDeliveryPipelineVerifyConfig` |  |  |  |
| `spec.serialPipeline.stages[].strategy.standard.verifyConfig.tasks` | `[]GcpDeliveryPipelineTask` |  |  |  |
| `spec.serialPipeline.stages[].strategy.standard.verifyConfig.tasks[].container` | `GcpDeliveryPipelineContainerTask` |  |  |  |
| `spec.serialPipeline.stages[].strategy.standard.verifyConfig.tasks[].container.image` | `string` | yes |  |  |
| `spec.serialPipeline.stages[].strategy.standard.verifyConfig.tasks[].container.command` | `[]string` |  |  |  |
| `spec.serialPipeline.stages[].strategy.standard.verifyConfig.tasks[].container.args` | `[]string` |  |  |  |
| `spec.serialPipeline.stages[].strategy.standard.verifyConfig.tasks[].container.env` | `map<string, string>` |  |  |  |
| `spec.serialPipeline.stages[].strategy.standard.analysis` | `GcpDeliveryPipelineAnalysis` |  |  |  |
| `spec.serialPipeline.stages[].strategy.standard.analysis.duration` | `string` | yes |  |  |
| `spec.serialPipeline.stages[].strategy.standard.analysis.googleCloud` | `GcpDeliveryPipelineGoogleCloudAnalysis` |  |  |  |
| `spec.serialPipeline.stages[].strategy.standard.analysis.googleCloud.alertPolicyChecks` | `[]GcpDeliveryPipelineAlertPolicyCheck` |  |  |  |
| `spec.serialPipeline.stages[].strategy.standard.analysis.googleCloud.alertPolicyChecks[].id` | `string` | yes |  |  |
| `spec.serialPipeline.stages[].strategy.standard.analysis.googleCloud.alertPolicyChecks[].alertPolicies` | `[]string \| valueFrom` | yes |  | GcpMonitoringAlertPolicy (`status.outputs.policy_name`) |
| `spec.serialPipeline.stages[].strategy.standard.analysis.googleCloud.alertPolicyChecks[].labels` | `map<string, string>` |  |  |  |
| `spec.serialPipeline.stages[].strategy.standard.analysis.customChecks` | `[]GcpDeliveryPipelineCustomCheck` |  |  |  |
| `spec.serialPipeline.stages[].strategy.standard.analysis.customChecks[].id` | `string` | yes |  |  |
| `spec.serialPipeline.stages[].strategy.standard.analysis.customChecks[].frequency` | `string` |  |  |  |
| `spec.serialPipeline.stages[].strategy.standard.analysis.customChecks[].task` | `GcpDeliveryPipelineTask` |  |  |  |
| `spec.serialPipeline.stages[].strategy.standard.analysis.customChecks[].task.container` | `GcpDeliveryPipelineContainerTask` |  |  |  |
| `spec.serialPipeline.stages[].strategy.standard.analysis.customChecks[].task.container.image` | `string` | yes |  |  |
| `spec.serialPipeline.stages[].strategy.standard.analysis.customChecks[].task.container.command` | `[]string` |  |  |  |
| `spec.serialPipeline.stages[].strategy.standard.analysis.customChecks[].task.container.args` | `[]string` |  |  |  |
| `spec.serialPipeline.stages[].strategy.standard.analysis.customChecks[].task.container.env` | `map<string, string>` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary` | `GcpDeliveryPipelineCanary` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.canaryDeployment` | `GcpDeliveryPipelineCanaryDeployment` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.canaryDeployment.percentages` | `[]int32` | yes |  |  |
| `spec.serialPipeline.stages[].strategy.canary.canaryDeployment.verify` | `bool` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.canaryDeployment.predeploy` | `GcpDeliveryPipelineCanaryDeployHook` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.canaryDeployment.predeploy.actions` | `[]string` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.canaryDeployment.postdeploy` | `GcpDeliveryPipelineCanaryDeployHook` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.canaryDeployment.postdeploy.actions` | `[]string` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.canaryDeployment.verifyConfig` | `GcpDeliveryPipelineVerifyConfig` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.canaryDeployment.verifyConfig.tasks` | `[]GcpDeliveryPipelineTask` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.canaryDeployment.verifyConfig.tasks[].container` | `GcpDeliveryPipelineContainerTask` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.canaryDeployment.verifyConfig.tasks[].container.image` | `string` | yes |  |  |
| `spec.serialPipeline.stages[].strategy.canary.canaryDeployment.verifyConfig.tasks[].container.command` | `[]string` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.canaryDeployment.verifyConfig.tasks[].container.args` | `[]string` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.canaryDeployment.verifyConfig.tasks[].container.env` | `map<string, string>` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.canaryDeployment.analysis` | `GcpDeliveryPipelineAnalysis` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.canaryDeployment.analysis.duration` | `string` | yes |  |  |
| `spec.serialPipeline.stages[].strategy.canary.canaryDeployment.analysis.googleCloud` | `GcpDeliveryPipelineGoogleCloudAnalysis` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.canaryDeployment.analysis.googleCloud.alertPolicyChecks` | `[]GcpDeliveryPipelineAlertPolicyCheck` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.canaryDeployment.analysis.googleCloud.alertPolicyChecks[].id` | `string` | yes |  |  |
| `spec.serialPipeline.stages[].strategy.canary.canaryDeployment.analysis.googleCloud.alertPolicyChecks[].alertPolicies` | `[]string \| valueFrom` | yes |  | GcpMonitoringAlertPolicy (`status.outputs.policy_name`) |
| `spec.serialPipeline.stages[].strategy.canary.canaryDeployment.analysis.googleCloud.alertPolicyChecks[].labels` | `map<string, string>` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.canaryDeployment.analysis.customChecks` | `[]GcpDeliveryPipelineCustomCheck` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.canaryDeployment.analysis.customChecks[].id` | `string` | yes |  |  |
| `spec.serialPipeline.stages[].strategy.canary.canaryDeployment.analysis.customChecks[].frequency` | `string` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.canaryDeployment.analysis.customChecks[].task` | `GcpDeliveryPipelineTask` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.canaryDeployment.analysis.customChecks[].task.container` | `GcpDeliveryPipelineContainerTask` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.canaryDeployment.analysis.customChecks[].task.container.image` | `string` | yes |  |  |
| `spec.serialPipeline.stages[].strategy.canary.canaryDeployment.analysis.customChecks[].task.container.command` | `[]string` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.canaryDeployment.analysis.customChecks[].task.container.args` | `[]string` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.canaryDeployment.analysis.customChecks[].task.container.env` | `map<string, string>` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment` | `GcpDeliveryPipelineCustomCanaryDeployment` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs` | `[]GcpDeliveryPipelinePhaseConfig` | yes |  |  |
| `spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].phaseId` | `string` | yes |  |  |
| `spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].percentage` | `int32` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].profiles` | `[]string` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].verify` | `bool` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].predeploy` | `GcpDeliveryPipelineCanaryDeployHook` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].predeploy.actions` | `[]string` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].postdeploy` | `GcpDeliveryPipelineCanaryDeployHook` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].postdeploy.actions` | `[]string` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].verifyConfig` | `GcpDeliveryPipelineVerifyConfig` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].verifyConfig.tasks` | `[]GcpDeliveryPipelineTask` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].verifyConfig.tasks[].container` | `GcpDeliveryPipelineContainerTask` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].verifyConfig.tasks[].container.image` | `string` | yes |  |  |
| `spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].verifyConfig.tasks[].container.command` | `[]string` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].verifyConfig.tasks[].container.args` | `[]string` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].verifyConfig.tasks[].container.env` | `map<string, string>` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].analysis` | `GcpDeliveryPipelineAnalysis` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].analysis.duration` | `string` | yes |  |  |
| `spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].analysis.googleCloud` | `GcpDeliveryPipelineGoogleCloudAnalysis` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].analysis.googleCloud.alertPolicyChecks` | `[]GcpDeliveryPipelineAlertPolicyCheck` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].analysis.googleCloud.alertPolicyChecks[].id` | `string` | yes |  |  |
| `spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].analysis.googleCloud.alertPolicyChecks[].alertPolicies` | `[]string \| valueFrom` | yes |  | GcpMonitoringAlertPolicy (`status.outputs.policy_name`) |
| `spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].analysis.googleCloud.alertPolicyChecks[].labels` | `map<string, string>` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].analysis.customChecks` | `[]GcpDeliveryPipelineCustomCheck` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].analysis.customChecks[].id` | `string` | yes |  |  |
| `spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].analysis.customChecks[].frequency` | `string` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].analysis.customChecks[].task` | `GcpDeliveryPipelineTask` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].analysis.customChecks[].task.container` | `GcpDeliveryPipelineContainerTask` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].analysis.customChecks[].task.container.image` | `string` | yes |  |  |
| `spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].analysis.customChecks[].task.container.command` | `[]string` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].analysis.customChecks[].task.container.args` | `[]string` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].analysis.customChecks[].task.container.env` | `map<string, string>` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.runtimeConfig` | `GcpDeliveryPipelineRuntimeConfig` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.runtimeConfig.cloudRun` | `GcpDeliveryPipelineCloudRunConfig` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.runtimeConfig.cloudRun.automaticTrafficControl` | `bool` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.runtimeConfig.cloudRun.canaryRevisionTags` | `[]string` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.runtimeConfig.cloudRun.priorRevisionTags` | `[]string` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.runtimeConfig.cloudRun.stableRevisionTags` | `[]string` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.runtimeConfig.kubernetes` | `GcpDeliveryPipelineKubernetesConfig` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.runtimeConfig.kubernetes.gatewayServiceMesh` | `GcpDeliveryPipelineGatewayServiceMesh` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.runtimeConfig.kubernetes.gatewayServiceMesh.httpRoute` | `string` | yes |  |  |
| `spec.serialPipeline.stages[].strategy.canary.runtimeConfig.kubernetes.gatewayServiceMesh.service` | `string` | yes |  |  |
| `spec.serialPipeline.stages[].strategy.canary.runtimeConfig.kubernetes.gatewayServiceMesh.deployment` | `string` | yes |  |  |
| `spec.serialPipeline.stages[].strategy.canary.runtimeConfig.kubernetes.gatewayServiceMesh.routeUpdateWaitTime` | `string` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.runtimeConfig.kubernetes.gatewayServiceMesh.stableCutbackDuration` | `string` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.runtimeConfig.kubernetes.gatewayServiceMesh.podSelectorLabel` | `string` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.runtimeConfig.kubernetes.gatewayServiceMesh.routeDestinations` | `GcpDeliveryPipelineRouteDestinations` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.runtimeConfig.kubernetes.gatewayServiceMesh.routeDestinations.destinationIds` | `[]string` | yes |  |  |
| `spec.serialPipeline.stages[].strategy.canary.runtimeConfig.kubernetes.gatewayServiceMesh.routeDestinations.propagateService` | `bool` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.runtimeConfig.kubernetes.serviceNetworking` | `GcpDeliveryPipelineServiceNetworking` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.runtimeConfig.kubernetes.serviceNetworking.service` | `string` | yes |  |  |
| `spec.serialPipeline.stages[].strategy.canary.runtimeConfig.kubernetes.serviceNetworking.deployment` | `string` | yes |  |  |
| `spec.serialPipeline.stages[].strategy.canary.runtimeConfig.kubernetes.serviceNetworking.disablePodOverprovisioning` | `bool` |  |  |  |
| `spec.serialPipeline.stages[].strategy.canary.runtimeConfig.kubernetes.serviceNetworking.podSelectorLabel` | `string` |  |  |  |
| `spec.serialPipeline.stages[].deployParameters` | `[]GcpDeliveryPipelineDeployParameters` |  |  |  |
| `spec.serialPipeline.stages[].deployParameters[].values` | `map<string, string>` | yes |  |  |
| `spec.serialPipeline.stages[].deployParameters[].matchTargetLabels` | `map<string, string>` |  |  |  |
| `spec.automations` | `[]GcpDeliveryPipelineAutomation` |  |  |  |
| `spec.automations[].automationId` | `string` | yes |  |  |
| `spec.automations[].description` | `string` |  |  |  |
| `spec.automations[].labels` | `map<string, string>` |  |  |  |
| `spec.automations[].annotations` | `map<string, string>` |  |  |  |
| `spec.automations[].suspended` | `bool` |  |  |  |
| `spec.automations[].serviceAccount` | `string \| valueFrom` | yes |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.automations[].selector` | `GcpDeliveryPipelineAutomationSelector` | yes |  |  |
| `spec.automations[].selector.targets` | `[]GcpDeliveryPipelineAutomationTarget` | yes |  |  |
| `spec.automations[].selector.targets[].id` | `string \| valueFrom` |  |  | GcpDeployTarget (`status.outputs.target_id`) |
| `spec.automations[].selector.targets[].labels` | `map<string, string>` |  |  |  |
| `spec.automations[].rules` | `[]GcpDeliveryPipelineAutomationRule` | yes |  |  |
| `spec.automations[].rules[].advanceRolloutRule` | `GcpDeliveryPipelineAdvanceRolloutRule` |  |  |  |
| `spec.automations[].rules[].advanceRolloutRule.id` | `string` | yes |  |  |
| `spec.automations[].rules[].advanceRolloutRule.sourcePhases` | `[]string` |  |  |  |
| `spec.automations[].rules[].advanceRolloutRule.wait` | `string` |  |  |  |
| `spec.automations[].rules[].promoteReleaseRule` | `GcpDeliveryPipelinePromoteReleaseRule` |  |  |  |
| `spec.automations[].rules[].promoteReleaseRule.id` | `string` | yes |  |  |
| `spec.automations[].rules[].promoteReleaseRule.wait` | `string` |  |  |  |
| `spec.automations[].rules[].promoteReleaseRule.destinationTargetId` | `string \| valueFrom` |  |  | GcpDeployTarget (`status.outputs.target_id`) |
| `spec.automations[].rules[].promoteReleaseRule.destinationPhase` | `string` |  |  |  |
| `spec.automations[].rules[].repairRolloutRule` | `GcpDeliveryPipelineRepairRolloutRule` |  |  |  |
| `spec.automations[].rules[].repairRolloutRule.id` | `string` | yes |  |  |
| `spec.automations[].rules[].repairRolloutRule.phases` | `[]string` |  |  |  |
| `spec.automations[].rules[].repairRolloutRule.jobs` | `[]string` |  |  |  |
| `spec.automations[].rules[].repairRolloutRule.repairPhases` | `[]GcpDeliveryPipelineRepairPhase` | yes |  |  |
| `spec.automations[].rules[].repairRolloutRule.repairPhases[].retry` | `GcpDeliveryPipelineRetry` |  |  |  |
| `spec.automations[].rules[].repairRolloutRule.repairPhases[].retry.attempts` | `string` | yes |  |  |
| `spec.automations[].rules[].repairRolloutRule.repairPhases[].retry.wait` | `string` |  |  |  |
| `spec.automations[].rules[].repairRolloutRule.repairPhases[].retry.backoffMode` | `string` |  |  |  |
| `spec.automations[].rules[].repairRolloutRule.repairPhases[].rollback` | `GcpDeliveryPipelineRollback` |  |  |  |
| `spec.automations[].rules[].repairRolloutRule.repairPhases[].rollback.destinationPhase` | `string` |  |  |  |
| `spec.automations[].rules[].repairRolloutRule.repairPhases[].rollback.disableRollbackIfRolloutPending` | `bool` |  |  |  |
| `spec.automations[].rules[].timedPromoteReleaseRule` | `GcpDeliveryPipelineTimedPromoteReleaseRule` |  |  |  |
| `spec.automations[].rules[].timedPromoteReleaseRule.id` | `string` | yes |  |  |
| `spec.automations[].rules[].timedPromoteReleaseRule.schedule` | `string` | yes |  |  |
| `spec.automations[].rules[].timedPromoteReleaseRule.timeZone` | `string` | yes |  |  |
| `spec.automations[].rules[].timedPromoteReleaseRule.destinationTargetId` | `string \| valueFrom` |  |  | GcpDeployTarget (`status.outputs.target_id`) |
| `spec.automations[].rules[].timedPromoteReleaseRule.destinationPhase` | `string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The project the pipeline lives in: a literal project ID or a
GcpProject reference. Empty means the provider's default project. The
stages' targets and the automations live in the same project.
Immutable.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.location

`string` · required

The region the pipeline lives in, e.g. "us-central1". Every stage's
target is looked up in this region. Required. Immutable.

- rule: {"required":true}

### spec.deliveryPipelineId

`string`

The pipeline's ID, unique in the project and region: 1-63 lowercase
letters, digits, and hyphens, starting with a letter and not ending
with a hyphen. Defaults to metadata.name. Immutable.

- rule: delivery_pipeline_id must be 1-63 lowercase letters, digits, or hyphens, starting with a letter and not ending with a hyphen

### spec.description

`string`

A description shown in the console, up to 255 characters.

- rule: {"string":{"maxLen":"255"}}

### spec.labels

`map<string, string>`

Labels on the pipeline: lowercase keys and values, at most 64 labels.
The platform attribution labels are added on top and win on key
conflicts. Deploy policies can select pipelines by label.

### spec.annotations

`map<string, string>`

Annotations on the pipeline (Google's AIP-128 key/value metadata that
Cloud Deploy never reads). Only the keys declared here are managed.

### spec.suspended

`bool`

Suspend the pipeline: no new releases or rollouts can be created, but
in-progress ones complete. Updates in place.

### spec.serialPipeline

`GcpDeliveryPipelineSerialPipeline`

The promotion flow: the ordered stages a release moves through.

### spec.serialPipeline.stages

`[]GcpDeliveryPipelineStage`

The stages, first to last. A release rolls out to the first stage's
target and is promoted to each next one.

### spec.serialPipeline.stages[].targetId

`string | valueFrom`

The target this stage deploys to, by its bare ID ("prod", never
projects/.../targets/prod): a GcpDeployTarget reference (its target_id
output) or a literal ID. Google looks the target up in this pipeline's
project and region, so a referenced target must live there too.

- references: GcpDeployTarget (`status.outputs.target_id`)
- rule: target_id is the target's bare ID (the last segment of its name), not a full resource name
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpDeployTarget, name: <that resource's name>, fieldPath: status.outputs.target_id}} -- a bare string does not parse

### spec.serialPipeline.stages[].profiles

`[]string`

Skaffold profiles to render this stage's manifests with, e.g.
["prod"] to pick the prod overlay of a skaffold.yaml.

### spec.serialPipeline.stages[].strategy

`GcpDeliveryPipelineStrategy`

The rollout strategy for this stage. Empty means a standard deploy
without verification.

### spec.serialPipeline.stages[].strategy.standard

`GcpDeliveryPipelineStandard`

Deploy the release in one step, optionally verifying it and running
jobs before and after.

### spec.serialPipeline.stages[].strategy.standard.verify

`bool`

Run `skaffold verify` (or verify_config's tasks) after the deploy; a
failed verification fails the rollout.

### spec.serialPipeline.stages[].strategy.standard.predeploy

`GcpDeliveryPipelineStandardDeployHook`

A job that runs before the deploy.

- rule: set at most one of actions or tasks

### spec.serialPipeline.stages[].strategy.standard.predeploy.actions

`[]string`

Skaffold custom actions (customActions names in skaffold.yaml) to run
in order.

### spec.serialPipeline.stages[].strategy.standard.predeploy.tasks

`[]GcpDeliveryPipelineTask`

Containers to run in order, in Cloud Deploy's Cloud Build execution
environment.

### spec.serialPipeline.stages[].strategy.standard.predeploy.tasks[].container

`GcpDeliveryPipelineContainerTask`

A container run in Cloud Deploy's Cloud Build execution environment.

### spec.serialPipeline.stages[].strategy.standard.predeploy.tasks[].container.image

`string` · required

The container image, e.g. "us-docker.pkg.dev/my-project/tools/smoke:1".
Required.

- rule: {"required":true}

### spec.serialPipeline.stages[].strategy.standard.predeploy.tasks[].container.command

`[]string`

Overrides the image's entrypoint.

### spec.serialPipeline.stages[].strategy.standard.predeploy.tasks[].container.args

`[]string`

Overrides the image's default arguments.

### spec.serialPipeline.stages[].strategy.standard.predeploy.tasks[].container.env

`map<string, string>`

Environment variables set in the container. Values are stored on the
pipeline in plain text: never put secrets here.

### spec.serialPipeline.stages[].strategy.standard.postdeploy

`GcpDeliveryPipelineStandardDeployHook`

A job that runs after the deploy.

- rule: set at most one of actions or tasks

### spec.serialPipeline.stages[].strategy.standard.postdeploy.actions

`[]string`

Skaffold custom actions (customActions names in skaffold.yaml) to run
in order.

### spec.serialPipeline.stages[].strategy.standard.postdeploy.tasks

`[]GcpDeliveryPipelineTask`

Containers to run in order, in Cloud Deploy's Cloud Build execution
environment.

### spec.serialPipeline.stages[].strategy.standard.postdeploy.tasks[].container

`GcpDeliveryPipelineContainerTask`

A container run in Cloud Deploy's Cloud Build execution environment.

### spec.serialPipeline.stages[].strategy.standard.postdeploy.tasks[].container.image

`string` · required

The container image, e.g. "us-docker.pkg.dev/my-project/tools/smoke:1".
Required.

- rule: {"required":true}

### spec.serialPipeline.stages[].strategy.standard.postdeploy.tasks[].container.command

`[]string`

Overrides the image's entrypoint.

### spec.serialPipeline.stages[].strategy.standard.postdeploy.tasks[].container.args

`[]string`

Overrides the image's default arguments.

### spec.serialPipeline.stages[].strategy.standard.postdeploy.tasks[].container.env

`map<string, string>`

Environment variables set in the container. Values are stored on the
pipeline in plain text: never put secrets here.

### spec.serialPipeline.stages[].strategy.standard.verifyConfig

`GcpDeliveryPipelineVerifyConfig`

Container tasks that run as the verify job, instead of the release's
Skaffold verify configuration.

### spec.serialPipeline.stages[].strategy.standard.verifyConfig.tasks

`[]GcpDeliveryPipelineTask`

The containers to run in order.

### spec.serialPipeline.stages[].strategy.standard.verifyConfig.tasks[].container

`GcpDeliveryPipelineContainerTask`

A container run in Cloud Deploy's Cloud Build execution environment.

### spec.serialPipeline.stages[].strategy.standard.verifyConfig.tasks[].container.image

`string` · required

The container image, e.g. "us-docker.pkg.dev/my-project/tools/smoke:1".
Required.

- rule: {"required":true}

### spec.serialPipeline.stages[].strategy.standard.verifyConfig.tasks[].container.command

`[]string`

Overrides the image's entrypoint.

### spec.serialPipeline.stages[].strategy.standard.verifyConfig.tasks[].container.args

`[]string`

Overrides the image's default arguments.

### spec.serialPipeline.stages[].strategy.standard.verifyConfig.tasks[].container.env

`map<string, string>`

Environment variables set in the container. Values are stored on the
pipeline in plain text: never put secrets here.

### spec.serialPipeline.stages[].strategy.standard.analysis

`GcpDeliveryPipelineAnalysis`

An analysis job that watches the deploy for a while and fails the
rollout when a check fails.

### spec.serialPipeline.stages[].strategy.standard.analysis.duration

`string` · required

How long the analysis runs, in seconds format (e.g. "600s"). Required.

- rule: {"required":true}

### spec.serialPipeline.stages[].strategy.standard.analysis.googleCloud

`GcpDeliveryPipelineGoogleCloudAnalysis`

Checks against Cloud Monitoring alert policies.

### spec.serialPipeline.stages[].strategy.standard.analysis.googleCloud.alertPolicyChecks

`[]GcpDeliveryPipelineAlertPolicyCheck`

Alert-policy checks: the analysis fails when a listed policy fires.

### spec.serialPipeline.stages[].strategy.standard.analysis.googleCloud.alertPolicyChecks[].id

`string` · required

The check's ID, unique in the analysis. Required.

- rule: {"required":true}

### spec.serialPipeline.stages[].strategy.standard.analysis.googleCloud.alertPolicyChecks[].alertPolicies

`[]string | valueFrom` · required

The alert policies to watch, each projects/{project}/alertPolicies/{id}:
GcpMonitoringAlertPolicy references (their policy_name output) or
literals. Required.

- references: GcpMonitoringAlertPolicy (`status.outputs.policy_name`)
- rule: {"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpMonitoringAlertPolicy, name: <that resource's name>, fieldPath: status.outputs.policy_name}} -- a bare string does not parse

### spec.serialPipeline.stages[].strategy.standard.analysis.googleCloud.alertPolicyChecks[].labels

`map<string, string>`

Labels that filter which incidents of those policies count.

### spec.serialPipeline.stages[].strategy.standard.analysis.customChecks

`[]GcpDeliveryPipelineCustomCheck`

Checks that run your own container.

### spec.serialPipeline.stages[].strategy.standard.analysis.customChecks[].id

`string` · required

The check's ID, unique in the analysis. Required.

- rule: {"required":true}

### spec.serialPipeline.stages[].strategy.standard.analysis.customChecks[].frequency

`string`

How often the check runs, in seconds format (e.g. "60s"). Empty uses
Cloud Deploy's default.

### spec.serialPipeline.stages[].strategy.standard.analysis.customChecks[].task

`GcpDeliveryPipelineTask`

The container the check runs.

### spec.serialPipeline.stages[].strategy.standard.analysis.customChecks[].task.container

`GcpDeliveryPipelineContainerTask`

A container run in Cloud Deploy's Cloud Build execution environment.

### spec.serialPipeline.stages[].strategy.standard.analysis.customChecks[].task.container.image

`string` · required

The container image, e.g. "us-docker.pkg.dev/my-project/tools/smoke:1".
Required.

- rule: {"required":true}

### spec.serialPipeline.stages[].strategy.standard.analysis.customChecks[].task.container.command

`[]string`

Overrides the image's entrypoint.

### spec.serialPipeline.stages[].strategy.standard.analysis.customChecks[].task.container.args

`[]string`

Overrides the image's default arguments.

### spec.serialPipeline.stages[].strategy.standard.analysis.customChecks[].task.container.env

`map<string, string>`

Environment variables set in the container. Values are stored on the
pipeline in plain text: never put secrets here.

### spec.serialPipeline.stages[].strategy.canary

`GcpDeliveryPipelineCanary`

Deploy the release progressively by percentage.

- rule: set at most one of canary_deployment or custom_canary_deployment
- rule: a canary_deployment on Cloud Run requires runtime_config.cloud_run.automatic_traffic_control to be true

### spec.serialPipeline.stages[].strategy.canary.canaryDeployment

`GcpDeliveryPipelineCanaryDeployment`

The same steps for every phase: a list of percentages.

### spec.serialPipeline.stages[].strategy.canary.canaryDeployment.percentages

`[]int32` · required

The percentages to deploy, in ascending order, e.g. [25, 50] (the
final 100% stable phase is implicit). Each is 0 <= n < 100; with a
Kubernetes Gateway API service mesh, 100 is also allowed. Required.

- rule: {"repeated":{"minItems":"1","items":{"int32":{"lte":100,"gte":0}}}}

### spec.serialPipeline.stages[].strategy.canary.canaryDeployment.verify

`bool`

Run verify tests after each percentage deployment.

### spec.serialPipeline.stages[].strategy.canary.canaryDeployment.predeploy

`GcpDeliveryPipelineCanaryDeployHook`

A job that runs before the first phase.

### spec.serialPipeline.stages[].strategy.canary.canaryDeployment.predeploy.actions

`[]string`

Skaffold custom actions (customActions names in skaffold.yaml) to run
in order.

### spec.serialPipeline.stages[].strategy.canary.canaryDeployment.postdeploy

`GcpDeliveryPipelineCanaryDeployHook`

A job that runs after the last phase.

### spec.serialPipeline.stages[].strategy.canary.canaryDeployment.postdeploy.actions

`[]string`

Skaffold custom actions (customActions names in skaffold.yaml) to run
in order.

### spec.serialPipeline.stages[].strategy.canary.canaryDeployment.verifyConfig

`GcpDeliveryPipelineVerifyConfig`

Container tasks that run as the verify job.

### spec.serialPipeline.stages[].strategy.canary.canaryDeployment.verifyConfig.tasks

`[]GcpDeliveryPipelineTask`

The containers to run in order.

### spec.serialPipeline.stages[].strategy.canary.canaryDeployment.verifyConfig.tasks[].container

`GcpDeliveryPipelineContainerTask`

A container run in Cloud Deploy's Cloud Build execution environment.

### spec.serialPipeline.stages[].strategy.canary.canaryDeployment.verifyConfig.tasks[].container.image

`string` · required

The container image, e.g. "us-docker.pkg.dev/my-project/tools/smoke:1".
Required.

- rule: {"required":true}

### spec.serialPipeline.stages[].strategy.canary.canaryDeployment.verifyConfig.tasks[].container.command

`[]string`

Overrides the image's entrypoint.

### spec.serialPipeline.stages[].strategy.canary.canaryDeployment.verifyConfig.tasks[].container.args

`[]string`

Overrides the image's default arguments.

### spec.serialPipeline.stages[].strategy.canary.canaryDeployment.verifyConfig.tasks[].container.env

`map<string, string>`

Environment variables set in the container. Values are stored on the
pipeline in plain text: never put secrets here.

### spec.serialPipeline.stages[].strategy.canary.canaryDeployment.analysis

`GcpDeliveryPipelineAnalysis`

An analysis job for each phase.

### spec.serialPipeline.stages[].strategy.canary.canaryDeployment.analysis.duration

`string` · required

How long the analysis runs, in seconds format (e.g. "600s"). Required.

- rule: {"required":true}

### spec.serialPipeline.stages[].strategy.canary.canaryDeployment.analysis.googleCloud

`GcpDeliveryPipelineGoogleCloudAnalysis`

Checks against Cloud Monitoring alert policies.

### spec.serialPipeline.stages[].strategy.canary.canaryDeployment.analysis.googleCloud.alertPolicyChecks

`[]GcpDeliveryPipelineAlertPolicyCheck`

Alert-policy checks: the analysis fails when a listed policy fires.

### spec.serialPipeline.stages[].strategy.canary.canaryDeployment.analysis.googleCloud.alertPolicyChecks[].id

`string` · required

The check's ID, unique in the analysis. Required.

- rule: {"required":true}

### spec.serialPipeline.stages[].strategy.canary.canaryDeployment.analysis.googleCloud.alertPolicyChecks[].alertPolicies

`[]string | valueFrom` · required

The alert policies to watch, each projects/{project}/alertPolicies/{id}:
GcpMonitoringAlertPolicy references (their policy_name output) or
literals. Required.

- references: GcpMonitoringAlertPolicy (`status.outputs.policy_name`)
- rule: {"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpMonitoringAlertPolicy, name: <that resource's name>, fieldPath: status.outputs.policy_name}} -- a bare string does not parse

### spec.serialPipeline.stages[].strategy.canary.canaryDeployment.analysis.googleCloud.alertPolicyChecks[].labels

`map<string, string>`

Labels that filter which incidents of those policies count.

### spec.serialPipeline.stages[].strategy.canary.canaryDeployment.analysis.customChecks

`[]GcpDeliveryPipelineCustomCheck`

Checks that run your own container.

### spec.serialPipeline.stages[].strategy.canary.canaryDeployment.analysis.customChecks[].id

`string` · required

The check's ID, unique in the analysis. Required.

- rule: {"required":true}

### spec.serialPipeline.stages[].strategy.canary.canaryDeployment.analysis.customChecks[].frequency

`string`

How often the check runs, in seconds format (e.g. "60s"). Empty uses
Cloud Deploy's default.

### spec.serialPipeline.stages[].strategy.canary.canaryDeployment.analysis.customChecks[].task

`GcpDeliveryPipelineTask`

The container the check runs.

### spec.serialPipeline.stages[].strategy.canary.canaryDeployment.analysis.customChecks[].task.container

`GcpDeliveryPipelineContainerTask`

A container run in Cloud Deploy's Cloud Build execution environment.

### spec.serialPipeline.stages[].strategy.canary.canaryDeployment.analysis.customChecks[].task.container.image

`string` · required

The container image, e.g. "us-docker.pkg.dev/my-project/tools/smoke:1".
Required.

- rule: {"required":true}

### spec.serialPipeline.stages[].strategy.canary.canaryDeployment.analysis.customChecks[].task.container.command

`[]string`

Overrides the image's entrypoint.

### spec.serialPipeline.stages[].strategy.canary.canaryDeployment.analysis.customChecks[].task.container.args

`[]string`

Overrides the image's default arguments.

### spec.serialPipeline.stages[].strategy.canary.canaryDeployment.analysis.customChecks[].task.container.env

`map<string, string>`

Environment variables set in the container. Values are stored on the
pipeline in plain text: never put secrets here.

### spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment

`GcpDeliveryPipelineCustomCanaryDeployment`

Phase-by-phase control: each phase its own percentage, profiles, and
jobs.

- rule: phase_id must be unique within the canary

### spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs

`[]GcpDeliveryPipelinePhaseConfig` · required

The phases in the order they run. Required.

- rule: {"repeated":{"minItems":"1"}}

### spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].phaseId

`string` · required

The rollout phase's ID: lowercase letters, digits, and hyphens,
starting with a letter, not ending with a hyphen, up to 63 characters.
Required.

- rule: phase_id must be 1-63 lowercase letters, digits, or hyphens, starting with a letter and not ending with a hyphen
- rule: {"required":true}

### spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].percentage

`int32`

The percentage this phase deploys, 0-100 (100 is the stable phase).
Always sent.

- rule: {"int32":{"lte":100,"gte":0}}

### spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].profiles

`[]string`

Skaffold profiles for this phase, in addition to the stage's profiles.

### spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].verify

`bool`

Run verify tests after this phase's deployment.

### spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].predeploy

`GcpDeliveryPipelineCanaryDeployHook`

A job that runs before this phase.

### spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].predeploy.actions

`[]string`

Skaffold custom actions (customActions names in skaffold.yaml) to run
in order.

### spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].postdeploy

`GcpDeliveryPipelineCanaryDeployHook`

A job that runs after this phase.

### spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].postdeploy.actions

`[]string`

Skaffold custom actions (customActions names in skaffold.yaml) to run
in order.

### spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].verifyConfig

`GcpDeliveryPipelineVerifyConfig`

Container tasks that run as this phase's verify job.

### spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].verifyConfig.tasks

`[]GcpDeliveryPipelineTask`

The containers to run in order.

### spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].verifyConfig.tasks[].container

`GcpDeliveryPipelineContainerTask`

A container run in Cloud Deploy's Cloud Build execution environment.

### spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].verifyConfig.tasks[].container.image

`string` · required

The container image, e.g. "us-docker.pkg.dev/my-project/tools/smoke:1".
Required.

- rule: {"required":true}

### spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].verifyConfig.tasks[].container.command

`[]string`

Overrides the image's entrypoint.

### spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].verifyConfig.tasks[].container.args

`[]string`

Overrides the image's default arguments.

### spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].verifyConfig.tasks[].container.env

`map<string, string>`

Environment variables set in the container. Values are stored on the
pipeline in plain text: never put secrets here.

### spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].analysis

`GcpDeliveryPipelineAnalysis`

An analysis job for this phase.

### spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].analysis.duration

`string` · required

How long the analysis runs, in seconds format (e.g. "600s"). Required.

- rule: {"required":true}

### spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].analysis.googleCloud

`GcpDeliveryPipelineGoogleCloudAnalysis`

Checks against Cloud Monitoring alert policies.

### spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].analysis.googleCloud.alertPolicyChecks

`[]GcpDeliveryPipelineAlertPolicyCheck`

Alert-policy checks: the analysis fails when a listed policy fires.

### spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].analysis.googleCloud.alertPolicyChecks[].id

`string` · required

The check's ID, unique in the analysis. Required.

- rule: {"required":true}

### spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].analysis.googleCloud.alertPolicyChecks[].alertPolicies

`[]string | valueFrom` · required

The alert policies to watch, each projects/{project}/alertPolicies/{id}:
GcpMonitoringAlertPolicy references (their policy_name output) or
literals. Required.

- references: GcpMonitoringAlertPolicy (`status.outputs.policy_name`)
- rule: {"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpMonitoringAlertPolicy, name: <that resource's name>, fieldPath: status.outputs.policy_name}} -- a bare string does not parse

### spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].analysis.googleCloud.alertPolicyChecks[].labels

`map<string, string>`

Labels that filter which incidents of those policies count.

### spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].analysis.customChecks

`[]GcpDeliveryPipelineCustomCheck`

Checks that run your own container.

### spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].analysis.customChecks[].id

`string` · required

The check's ID, unique in the analysis. Required.

- rule: {"required":true}

### spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].analysis.customChecks[].frequency

`string`

How often the check runs, in seconds format (e.g. "60s"). Empty uses
Cloud Deploy's default.

### spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].analysis.customChecks[].task

`GcpDeliveryPipelineTask`

The container the check runs.

### spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].analysis.customChecks[].task.container

`GcpDeliveryPipelineContainerTask`

A container run in Cloud Deploy's Cloud Build execution environment.

### spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].analysis.customChecks[].task.container.image

`string` · required

The container image, e.g. "us-docker.pkg.dev/my-project/tools/smoke:1".
Required.

- rule: {"required":true}

### spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].analysis.customChecks[].task.container.command

`[]string`

Overrides the image's entrypoint.

### spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].analysis.customChecks[].task.container.args

`[]string`

Overrides the image's default arguments.

### spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].analysis.customChecks[].task.container.env

`map<string, string>`

Environment variables set in the container. Values are stored on the
pipeline in plain text: never put secrets here.

### spec.serialPipeline.stages[].strategy.canary.runtimeConfig

`GcpDeliveryPipelineRuntimeConfig`

How Cloud Deploy splits traffic during the canary: Cloud Run revision
traffic, or a Kubernetes Gateway API route or Service. Empty lets
Cloud Deploy pick by the target type (a GKE canary then needs one of
the Kubernetes arms).

- rule: set at most one of cloud_run or kubernetes

### spec.serialPipeline.stages[].strategy.canary.runtimeConfig.cloudRun

`GcpDeliveryPipelineCloudRunConfig`

Cloud Run: Cloud Deploy shifts revision traffic.

### spec.serialPipeline.stages[].strategy.canary.runtimeConfig.cloudRun.automaticTrafficControl

`bool`

Let Cloud Deploy rewrite the service's traffic stanza to split traffic
between revisions. Required true for a canary_deployment; optional for
a custom_canary_deployment (where your manifests may split traffic
themselves).

### spec.serialPipeline.stages[].strategy.canary.runtimeConfig.cloudRun.canaryRevisionTags

`[]string`

Revision tags added to the canary revision while a canary phase runs.

### spec.serialPipeline.stages[].strategy.canary.runtimeConfig.cloudRun.priorRevisionTags

`[]string`

Revision tags added to the prior revision while a canary phase runs.

### spec.serialPipeline.stages[].strategy.canary.runtimeConfig.cloudRun.stableRevisionTags

`[]string`

Revision tags added to the new stable revision when the stable phase
is applied.

### spec.serialPipeline.stages[].strategy.canary.runtimeConfig.kubernetes

`GcpDeliveryPipelineKubernetesConfig`

GKE: Cloud Deploy shifts traffic through a Gateway API route or a
Service.

- rule: set at most one of gateway_service_mesh or service_networking

### spec.serialPipeline.stages[].strategy.canary.runtimeConfig.kubernetes.gatewayServiceMesh

`GcpDeliveryPipelineGatewayServiceMesh`

Split traffic with a Gateway API HTTPRoute (service-mesh or Gateway
setups).

### spec.serialPipeline.stages[].strategy.canary.runtimeConfig.kubernetes.gatewayServiceMesh.httpRoute

`string` · required

The HTTPRoute whose weights Cloud Deploy adjusts. Required.

- rule: {"required":true}

### spec.serialPipeline.stages[].strategy.canary.runtimeConfig.kubernetes.gatewayServiceMesh.service

`string` · required

The Service the route sends traffic to. Required.

- rule: {"required":true}

### spec.serialPipeline.stages[].strategy.canary.runtimeConfig.kubernetes.gatewayServiceMesh.deployment

`string` · required

The Deployment behind the Service. Required.

- rule: {"required":true}

### spec.serialPipeline.stages[].strategy.canary.runtimeConfig.kubernetes.gatewayServiceMesh.routeUpdateWaitTime

`string`

How long to wait for route updates to propagate, in seconds format
(e.g. "60s"), at most 3 hours. Empty means no wait.

### spec.serialPipeline.stages[].strategy.canary.runtimeConfig.kubernetes.gatewayServiceMesh.stableCutbackDuration

`string`

How long to migrate traffic back from the canary Service to the
original one during the stable phase, "15s" to "3600s". Empty means
no cutback time.

### spec.serialPipeline.stages[].strategy.canary.runtimeConfig.kubernetes.gatewayServiceMesh.podSelectorLabel

`string`

The Pod label selecting the Deployment's and Service's Pods; it must
already be on both. Empty uses Cloud Deploy's default selection.

### spec.serialPipeline.stages[].strategy.canary.runtimeConfig.kubernetes.gatewayServiceMesh.routeDestinations

`GcpDeliveryPipelineRouteDestinations`

Deploy the HTTPRoute to additional clusters as well (multi-cluster
service mesh). Empty deploys it to the target cluster only.

### spec.serialPipeline.stages[].strategy.canary.runtimeConfig.kubernetes.gatewayServiceMesh.routeDestinations.destinationIds

`[]string` · required

The clusters: associated-entity IDs of the target (its
associated_entities), and "@self" for the target cluster itself.
Required.

- rule: {"repeated":{"minItems":"1"}}

### spec.serialPipeline.stages[].strategy.canary.runtimeConfig.kubernetes.gatewayServiceMesh.routeDestinations.propagateService

`bool`

Also deploy the Service to the destination clusters, so DNS resolves
there.

### spec.serialPipeline.stages[].strategy.canary.runtimeConfig.kubernetes.serviceNetworking

`GcpDeliveryPipelineServiceNetworking`

Split traffic by Pod count behind a Kubernetes Service.

### spec.serialPipeline.stages[].strategy.canary.runtimeConfig.kubernetes.serviceNetworking.service

`string` · required

The Service whose Pods are split. Required.

- rule: {"required":true}

### spec.serialPipeline.stages[].strategy.canary.runtimeConfig.kubernetes.serviceNetworking.deployment

`string` · required

The Deployment behind the Service. Required.

- rule: {"required":true}

### spec.serialPipeline.stages[].strategy.canary.runtimeConfig.kubernetes.serviceNetworking.disablePodOverprovisioning

`bool`

Limit the total Pods used by the canary to the Deployment's current
count instead of adding canary Pods on top.

### spec.serialPipeline.stages[].strategy.canary.runtimeConfig.kubernetes.serviceNetworking.podSelectorLabel

`string`

The Pod label selecting the Deployment's Pods; it must already be on
the Deployment. Empty uses Cloud Deploy's default selection.

### spec.serialPipeline.stages[].deployParameters

`[]GcpDeliveryPipelineDeployParameters`

Deploy parameters for this stage's target: values substituted into
the rendered manifests wherever they reference a parameter, each set
optionally limited to child targets with matching labels.

### spec.serialPipeline.stages[].deployParameters[].values

`map<string, string>` · required

The parameters as key/value pairs, e.g. {"replicas": "3"}. Required.

- rule: {"map":{"minPairs":"1"}}

### spec.serialPipeline.stages[].deployParameters[].matchTargetLabels

`map<string, string>`

Apply these values only to targets carrying all of these labels (for a
multi-target, its matching child targets). Empty applies them to every
target of the stage, child targets included.

### spec.automations

`[]GcpDeliveryPipelineAutomation`

The pipeline's automations, each keyed by its automation_id. At most
250 rules in total across a pipeline's automations (Google's limit).

- rule: rule ids must be unique within the automation

### spec.automations[].automationId

`string` · required

The automation's ID, unique in the pipeline. Required. Immutable.

- rule: {"required":true}

### spec.automations[].description

`string`

A description shown in the console, up to 255 characters.

- rule: {"string":{"maxLen":"255"}}

### spec.automations[].labels

`map<string, string>`

Labels on the automation (keys and values at most 63 characters). The
platform attribution labels are added on top and win on key conflicts.

### spec.automations[].annotations

`map<string, string>`

Annotations on the automation (AIP-128 key/value metadata Cloud Deploy
never reads).

### spec.automations[].suspended

`bool`

Deactivate the automation; its rules stop firing until resumed.

### spec.automations[].serviceAccount

`string | valueFrom` · required

The user-managed service account the automation creates releases and
rollouts as: its email, a GcpServiceAccount reference (its email
output) or a literal. The identity running the deploy needs
iam.serviceAccounts.actAs on it (roles/iam.serviceAccountUser), and
the account itself needs roles/clouddeploy.operator (or narrower
release and rollout permissions) plus actAs on the targets' execution
service accounts. Required.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.automations[].selector

`GcpDeliveryPipelineAutomationSelector` · required

Which of the pipeline's targets the automation acts on. Required.

- rule: {"required":true}

### spec.automations[].selector.targets

`[]GcpDeliveryPipelineAutomationTarget` · required

The targets, at least one; a target matches when it matches any entry.

- rule: {"repeated":{"minItems":"1"}}

### spec.automations[].selector.targets[].id

`string | valueFrom`

A target's bare ID -- a GcpDeployTarget reference (its target_id
output) or a literal -- or "*" for every target in the pipeline's
region.

- references: GcpDeployTarget (`status.outputs.target_id`)
- rule: id is a target's bare ID or "*", not a full resource name
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpDeployTarget, name: <that resource's name>, fieldPath: status.outputs.target_id}} -- a bare string does not parse

### spec.automations[].selector.targets[].labels

`map<string, string>`

Match targets carrying these labels. Empty keeps whatever Google
records.

### spec.automations[].rules

`[]GcpDeliveryPipelineAutomationRule` · required

The automation's rules, at least one. Their order is not their order
of execution.

- rule: {"repeated":{"minItems":"1"}}
- rule: a rule sets exactly one of advance_rollout_rule, promote_release_rule, repair_rollout_rule, or timed_promote_release_rule

### spec.automations[].rules[].advanceRolloutRule

`GcpDeliveryPipelineAdvanceRolloutRule`

Advance a successful canary rollout to its next phase.

### spec.automations[].rules[].advanceRolloutRule.id

`string` · required

The rule's ID, unique in the automation: 1-63 lowercase letters,
digits, and hyphens, starting with a letter and not ending with a
hyphen. Required.

- rule: a rule id must be 1-63 lowercase letters, digits, or hyphens, starting with a letter and not ending with a hyphen
- rule: {"required":true}

### spec.automations[].rules[].advanceRolloutRule.sourcePhases

`[]string`

Advance only from these phase IDs (e.g. "canary-25"). Empty means any
phase.

### spec.automations[].rules[].advanceRolloutRule.wait

`string`

How long to wait after the phase finishes, e.g. "600s". Empty means
no wait.

### spec.automations[].rules[].promoteReleaseRule

`GcpDeliveryPipelinePromoteReleaseRule`

Promote a release that succeeded on a selected target to the next
(or a named) stage.

### spec.automations[].rules[].promoteReleaseRule.id

`string` · required

The rule's ID, unique in the automation (same format as every rule
ID). Required.

- rule: a rule id must be 1-63 lowercase letters, digits, or hyphens, starting with a letter and not ending with a hyphen
- rule: {"required":true}

### spec.automations[].rules[].promoteReleaseRule.wait

`string`

How long the release waits before it is promoted, e.g. "3600s". Empty
means promote at once.

### spec.automations[].rules[].promoteReleaseRule.destinationTargetId

`string | valueFrom`

The stage to promote to: a stage target's bare ID (a GcpDeployTarget
reference or a literal), or "@next" for the next stage. Empty means
the next stage.

- references: GcpDeployTarget (`status.outputs.target_id`)
- rule: destination_target_id is a target's bare ID or "@next", not a full resource name
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpDeployTarget, name: <that resource's name>, fieldPath: status.outputs.target_id}} -- a bare string does not parse

### spec.automations[].rules[].promoteReleaseRule.destinationPhase

`string`

The phase the promoted rollout starts in. Empty means the first phase.

### spec.automations[].rules[].repairRolloutRule

`GcpDeliveryPipelineRepairRolloutRule`

Retry or roll back a failed rollout.

### spec.automations[].rules[].repairRolloutRule.id

`string` · required

The rule's ID, unique in the automation (same format as every rule
ID). Required.

- rule: a rule id must be 1-63 lowercase letters, digits, or hyphens, starting with a letter and not ending with a hyphen
- rule: {"required":true}

### spec.automations[].rules[].repairRolloutRule.phases

`[]string`

Repair failed jobs only in these phases. Empty means every phase.

### spec.automations[].rules[].repairRolloutRule.jobs

`[]string`

Repair only these jobs (e.g. "deploy", "verify"), in the phases above.
Empty means every job.

### spec.automations[].rules[].repairRolloutRule.repairPhases

`[]GcpDeliveryPipelineRepairPhase` · required

The repair steps, in order: retries, then a rollback. Google requires
at least one.

- rule: {"repeated":{"minItems":"1"}}
- rule: a repair phase sets exactly one of retry or rollback

### spec.automations[].rules[].repairRolloutRule.repairPhases[].retry

`GcpDeliveryPipelineRetry`

Retry the failed job.

### spec.automations[].rules[].repairRolloutRule.repairPhases[].retry.attempts

`string` · required

How many retries, "1" to "10" (a string, as the provider takes it).
Required.

- rule: attempts must be a whole number such as "3"
- rule: {"required":true}

### spec.automations[].rules[].repairRolloutRule.repairPhases[].retry.wait

`string`

How long to wait before the first retry, e.g. "60s", at most 14 days.
Empty means no wait.

### spec.automations[].rules[].repairRolloutRule.repairPhases[].retry.backoffMode

`string`

How the wait grows between retries:
  "" / "BACKOFF_MODE_LINEAR" -- linear (Google's default)
  "BACKOFF_MODE_EXPONENTIAL" -- doubles each time
Ignored when wait is empty.

- rule: backoff_mode must be BACKOFF_MODE_UNSPECIFIED, BACKOFF_MODE_LINEAR, or BACKOFF_MODE_EXPONENTIAL

### spec.automations[].rules[].repairRolloutRule.repairPhases[].rollback

`GcpDeliveryPipelineRollback`

Roll the target back to its last successful release. An empty block
({}) rolls back into the stable phase.

### spec.automations[].rules[].repairRolloutRule.repairPhases[].rollback.destinationPhase

`string`

The phase the rollback rollout starts in. Empty means the stable
phase.

### spec.automations[].rules[].repairRolloutRule.repairPhases[].rollback.disableRollbackIfRolloutPending

`bool`

Abort the rollback when a rollout is already pending on the target.

### spec.automations[].rules[].timedPromoteReleaseRule

`GcpDeliveryPipelineTimedPromoteReleaseRule`

Promote on a schedule (e.g. every Monday at 09:00).

### spec.automations[].rules[].timedPromoteReleaseRule.id

`string` · required

The rule's ID, unique in the automation (same format as every rule
ID). Required.

- rule: a rule id must be 1-63 lowercase letters, digits, or hyphens, starting with a letter and not ending with a hyphen
- rule: {"required":true}

### spec.automations[].rules[].timedPromoteReleaseRule.schedule

`string` · required

When to promote, in crontab format, e.g. "0 9 * * 1" for Mondays at
09:00. Required.

- rule: {"required":true}

### spec.automations[].rules[].timedPromoteReleaseRule.timeZone

`string` · required

The schedule's IANA time zone, e.g. "America/New_York". Required.

- rule: {"required":true}

### spec.automations[].rules[].timedPromoteReleaseRule.destinationTargetId

`string | valueFrom`

The stage to promote to: a stage target's bare ID (a GcpDeployTarget
reference or a literal), or "@next". Empty means the next stage.

- references: GcpDeployTarget (`status.outputs.target_id`)
- rule: destination_target_id is a target's bare ID or "@next", not a full resource name
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpDeployTarget, name: <that resource's name>, fieldPath: status.outputs.target_id}} -- a bare string does not parse

### spec.automations[].rules[].timedPromoteReleaseRule.destinationPhase

`string`

The phase the promoted rollout starts in. Empty means the first phase.

### spec.deletionPolicy

`string`

What destroy does, for the pipeline and its automations:
  "" / "DELETE" -- deleted (the pipeline with force=true, taking its
                   releases and rollouts with it)
  "PREVENT"     -- destroy fails
  "ABANDON"     -- everything leaves management and stays in Google

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `spec.unique_automation_ids`: automation_id must be unique within the pipeline

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpDeliveryPipeline, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name: projects/{project}/locations/{location}/deliveryPipelines/{delivery_pipeline_id}. |
| `status.outputs.delivery_pipeline_id` | `string` | The pipeline's ID -- the value `gcloud deploy releases create --delivery-pipeline` and a deploy policy's pipeline selector take. |
| `status.outputs.uid` | `string` | Google's unique identifier for the pipeline. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.serialPipeline.stages[].targetId` | GcpDeployTarget | `status.outputs.target_id` |
| `spec.serialPipeline.stages[].strategy.standard.analysis.googleCloud.alertPolicyChecks[].alertPolicies` | GcpMonitoringAlertPolicy | `status.outputs.policy_name` |
| `spec.serialPipeline.stages[].strategy.canary.canaryDeployment.analysis.googleCloud.alertPolicyChecks[].alertPolicies` | GcpMonitoringAlertPolicy | `status.outputs.policy_name` |
| `spec.serialPipeline.stages[].strategy.canary.customCanaryDeployment.phaseConfigs[].analysis.googleCloud.alertPolicyChecks[].alertPolicies` | GcpMonitoringAlertPolicy | `status.outputs.policy_name` |
| `spec.automations[].serviceAccount` | GcpServiceAccount | `status.outputs.email` |
| `spec.automations[].selector.targets[].id` | GcpDeployTarget | `status.outputs.target_id` |
| `spec.automations[].rules[].promoteReleaseRule.destinationTargetId` | GcpDeployTarget | `status.outputs.target_id` |
| `spec.automations[].rules[].timedPromoteReleaseRule.destinationTargetId` | GcpDeployTarget | `status.outputs.target_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpDeployPolicy | `spec.selectors[].deliveryPipeline.id` | `status.outputs.delivery_pipeline_id` |

## See Also

- [Overview](../README.md)

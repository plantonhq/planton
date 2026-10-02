package module

import (
	"fmt"

	"github.com/pkg/errors"
	kubernetesprovider "github.com/plantonhq/planton/catalog/kubernetes"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"sigs.k8s.io/yaml"
)

// buildHelmValues renders the typed spec into the chart's values map, then
// merges the spec's helm_values escape hatch over it with Helm `-f`
// semantics (maps deep-merge with the later document winning, lists
// replace).
//
// PARITY: the Terraform module reaches the same result natively — its
// helm_release passes values = [yamlencode(typed values), helm_values] and
// the provider merges the documents in exactly this order. Keep every typed
// mapping below in lockstep with the Terraform module's locals.
func buildHelmValues(locals *Locals) (map[string]interface{}, error) {
	spec := locals.Spec
	values := map[string]interface{}{}

	// Pin the chart's fullname to the release name (= metadata.name): every
	// chart object (Deployment, Service, RBAC) then carries a deterministic,
	// manifest-derived name — what verification, imports, and multi-instance
	// coexistence all key off.
	values["fullnameOverride"] = locals.ReleaseName

	// The chart creates the controller ServiceAccount; the module pins its
	// name (deterministic identity subject) and rides workload-identity
	// annotations on it.
	serviceAccount := map[string]interface{}{"name": locals.ServiceAccountName}
	if annotations := workloadIdentityAnnotations(spec.GetWorkloadIdentity()); len(annotations) > 0 {
		serviceAccount["annotations"] = annotations
	}
	values["serviceAccount"] = serviceAccount
	if spec.GetWorkloadIdentity().GetAks() != nil {
		// The azure-workload-identity webhook only injects the federated
		// token volume into pods carrying this label — the SA annotation
		// alone does nothing.
		values["podLabels"] = map[string]interface{}{"azure.workload.identity/use": "true"}
	}

	// ---- watching / sync behavior --------------------------------------
	if len(spec.GetSources()) > 0 {
		values["sources"] = toInterfaceSlice(spec.GetSources())
	}
	if spec.GetGatewayListenerSets() {
		values["enableGatewayListenerSets"] = true
	}
	if spec.Policy != nil {
		values["policy"] = spec.GetPolicy()
	}
	if spec.Registry != nil {
		values["registry"] = spec.GetRegistry()
	}
	if spec.GetTxtOwnerId() != "" {
		values["txtOwnerId"] = spec.GetTxtOwnerId()
	}
	if spec.GetTxtPrefix() != "" {
		values["txtPrefix"] = spec.GetTxtPrefix()
	}
	if spec.GetTxtSuffix() != "" {
		values["txtSuffix"] = spec.GetTxtSuffix()
	}
	if len(spec.GetDomainFilters()) > 0 {
		values["domainFilters"] = toInterfaceSlice(spec.GetDomainFilters())
	}
	if len(spec.GetExcludeDomains()) > 0 {
		values["excludeDomains"] = toInterfaceSlice(spec.GetExcludeDomains())
	}
	if spec.GetAnnotationFilter() != "" {
		values["annotationFilter"] = spec.GetAnnotationFilter()
	}
	if spec.GetLabelFilter() != "" {
		values["labelFilter"] = spec.GetLabelFilter()
	}
	if len(spec.GetManagedRecordTypes()) > 0 {
		values["managedRecordTypes"] = toInterfaceSlice(spec.GetManagedRecordTypes())
	}
	if spec.Interval != nil {
		values["interval"] = spec.GetInterval()
	}
	if spec.GetTriggerLoopOnEvent() {
		values["triggerLoopOnEvent"] = true
	}
	if spec.GetNamespaced() {
		values["namespaced"] = true
	}
	if spec.LogLevel != nil {
		values["logLevel"] = spec.GetLogLevel()
	}
	if spec.LogFormat != nil {
		values["logFormat"] = spec.GetLogFormat()
	}

	// ---- pod placement / sizing ------------------------------------------
	if r := resourcesMap(spec.GetResources()); r != nil {
		values["resources"] = r
	}
	if len(spec.GetNodeSelector()) > 0 {
		values["nodeSelector"] = stringMapToInterface(spec.GetNodeSelector())
	}
	if len(spec.GetTolerations()) > 0 {
		values["tolerations"] = tolerationsSlice(spec.GetTolerations())
	}
	if spec.GetPriorityClassName() != "" {
		values["priorityClassName"] = spec.GetPriorityClassName()
	}

	// ---- observability ------------------------------------------------------
	if p := spec.GetPrometheus(); p != nil && p.GetServiceMonitor() {
		sm := map[string]interface{}{"enabled": true}
		if p.GetServiceMonitorInterval() != "" {
			sm["interval"] = p.GetServiceMonitorInterval()
		}
		if len(p.GetServiceMonitorLabels()) > 0 {
			sm["additionalLabels"] = stringMapToInterface(p.GetServiceMonitorLabels())
		}
		values["serviceMonitor"] = sm
	}

	// ---- image ----------------------------------------------------------------
	image := map[string]interface{}{}
	if spec.GetImageRepository() != "" {
		image["repository"] = spec.GetImageRepository()
	}
	if spec.GetImageTag() != "" {
		image["tag"] = spec.GetImageTag()
	}
	if len(image) > 0 {
		values["image"] = image
	}

	// ---- provider + its env/args/volumes ------------------------------------
	if err := applyProvider(locals, values); err != nil {
		return nil, err
	}

	// ---- escape hatch (merged LAST, helm -f semantics) --------------------------
	if spec.GetHelmValues() != "" {
		overrides := map[string]interface{}{}
		if err := yaml.Unmarshal([]byte(spec.GetHelmValues()), &overrides); err != nil {
			return nil, errors.Wrap(err, "failed to parse helm_values as a YAML document")
		}
		values = mergeMaps(values, overrides)
	}

	return values, nil
}

// applyProvider renders the selected dns_provider arm: the chart's
// provider.name, the provider-specific CLI flags (extraArgs, assembled in a
// FIXED order so both engines render byte-identical lists), the credential
// env/volume wiring against the Secrets credentialSecrets materializes, and
// (webhook arm) the sidecar block.
func applyProvider(locals *Locals, values map[string]interface{}) error {
	spec := locals.Spec

	var args []interface{}
	var env []interface{}

	// Generic zone filtering: every arm's typed zone references render as
	// repeated --zone-id-filter flags.
	zoneFilterArgs := func(refs []*foreignkeyv1.StringValueOrRef) {
		for _, ref := range refs {
			if ref.GetValue() != "" {
				args = append(args, fmt.Sprintf("--zone-id-filter=%s", ref.GetValue()))
			}
		}
	}

	switch {
	case spec.GetAwsRoute53() != nil:
		aws := spec.GetAwsRoute53()
		values["provider"] = map[string]interface{}{"name": "aws"}
		zoneFilterArgs(aws.GetZoneIdFilters())
		if aws.ZoneType != nil && aws.GetZoneType() != "" {
			args = append(args, fmt.Sprintf("--aws-zone-type=%s", aws.GetZoneType()))
		}
		if aws.GetAssumeRole().GetValue() != "" {
			args = append(args, fmt.Sprintf("--aws-assume-role=%s", aws.GetAssumeRole().GetValue()))
		}
		if aws.GetAssumeRoleExternalId() != "" {
			args = append(args, fmt.Sprintf("--aws-assume-role-external-id=%s", aws.GetAssumeRoleExternalId()))
		}
		if aws.GetRegion() != "" {
			env = append(env, map[string]interface{}{"name": "AWS_DEFAULT_REGION", "value": aws.GetRegion()})
		}
		if aws.GetAccessKeyId() != "" {
			env = append(env,
				secretEnv("AWS_ACCESS_KEY_ID", locals.AwsSecretName, "access-key-id"),
				secretEnv("AWS_SECRET_ACCESS_KEY", locals.AwsSecretName, "secret-access-key"))
		}
		// dynamodb registry settings are AWS-scoped flags.
		if spec.GetDynamodbRegion() != "" {
			args = append(args, fmt.Sprintf("--dynamodb-region=%s", spec.GetDynamodbRegion()))
		}
		if spec.GetDynamodbTable() != "" {
			args = append(args, fmt.Sprintf("--dynamodb-table=%s", spec.GetDynamodbTable()))
		}

	case spec.GetGoogleCloudDns() != nil:
		gcp := spec.GetGoogleCloudDns()
		values["provider"] = map[string]interface{}{"name": "google"}
		args = append(args, fmt.Sprintf("--google-project=%s", gcp.GetProject().GetValue()))
		zoneFilterArgs(gcp.GetZoneIdFilters())
		if gcp.ZoneVisibility != nil && gcp.GetZoneVisibility() != "" {
			args = append(args, fmt.Sprintf("--google-zone-visibility=%s", gcp.GetZoneVisibility()))
		}
		if gcp.GetServiceAccountKeyJson() != "" {
			// ADC reads the key from a file: mount the Secret and point
			// GOOGLE_APPLICATION_CREDENTIALS at it.
			env = append(env, map[string]interface{}{
				"name":  "GOOGLE_APPLICATION_CREDENTIALS",
				"value": "/etc/kubernetes/gcp/credentials.json",
			})
			values["extraVolumes"] = []interface{}{map[string]interface{}{
				"name":   "gcp-credentials",
				"secret": map[string]interface{}{"secretName": locals.GcpSecretName},
			}}
			values["extraVolumeMounts"] = []interface{}{map[string]interface{}{
				"name":      "gcp-credentials",
				"mountPath": "/etc/kubernetes/gcp",
				"readOnly":  true,
			}}
		}

	case spec.GetAzureDns() != nil:
		az := spec.GetAzureDns()
		providerName := "azure"
		if az.GetPrivateZones() {
			providerName = "azure-private-dns"
		}
		values["provider"] = map[string]interface{}{"name": providerName}
		zoneFilterArgs(az.GetZoneIdFilters())
		// The controller reads identity + subscription + resource group
		// from azure.json, mounted at its DEFAULT config path — so no
		// --azure-config-file override is needed.
		values["extraVolumes"] = []interface{}{map[string]interface{}{
			"name":   "azure-config",
			"secret": map[string]interface{}{"secretName": locals.AzureSecretName},
		}}
		values["extraVolumeMounts"] = []interface{}{map[string]interface{}{
			"name":      "azure-config",
			"mountPath": "/etc/kubernetes/azure.json",
			"subPath":   "azure.json",
			"readOnly":  true,
		}}

	case spec.GetCloudflare() != nil:
		cf := spec.GetCloudflare()
		values["provider"] = map[string]interface{}{"name": "cloudflare"}
		zoneFilterArgs(cf.GetZoneIdFilters())
		if cf.GetProxied() {
			args = append(args, "--cloudflare-proxied")
		}
		if cf.DnsRecordsPerPage != nil {
			args = append(args, fmt.Sprintf("--cloudflare-dns-records-per-page=%d", cf.GetDnsRecordsPerPage()))
		}
		env = append(env, secretEnv("CF_API_TOKEN", locals.CloudflareSecretName, "api-token"))

	case spec.GetWebhook() != nil:
		wh := spec.GetWebhook()
		// provider.name=webhook makes the chart run the provider image as a
		// sidecar next to the controller, talking over localhost.
		webhookImage := map[string]interface{}{"repository": wh.GetImageRepository()}
		if wh.GetImageTag() != "" {
			webhookImage["tag"] = wh.GetImageTag()
		}
		webhook := map[string]interface{}{"image": webhookImage}
		if len(wh.GetArgs()) > 0 {
			webhook["args"] = toInterfaceSlice(wh.GetArgs())
		}
		if len(wh.GetEnv()) > 0 {
			var webhookEnv []interface{}
			for _, name := range sortedKeys(wh.GetEnv()) {
				webhookEnv = append(webhookEnv, map[string]interface{}{"name": name, "value": wh.GetEnv()[name]})
			}
			webhook["env"] = webhookEnv
		}
		if r := resourcesMap(wh.GetResources()); r != nil {
			webhook["resources"] = r
		}
		values["provider"] = map[string]interface{}{"name": "webhook", "webhook": webhook}

	case spec.GetInMemory() != nil:
		values["provider"] = map[string]interface{}{"name": "inmemory"}
		for _, zone := range spec.GetInMemory().GetZones() {
			args = append(args, fmt.Sprintf("--inmemory-zone=%s", zone))
		}
	}

	if len(args) > 0 {
		values["extraArgs"] = args
	}
	if len(env) > 0 {
		values["env"] = env
	}
	return nil
}

// secretEnv renders an env var sourced from a Kubernetes Secret key.
func secretEnv(name, secretName, key string) map[string]interface{} {
	return map[string]interface{}{
		"name": name,
		"valueFrom": map[string]interface{}{
			"secretKeyRef": map[string]interface{}{
				"name": secretName,
				"key":  key,
			},
		},
	}
}

// workloadIdentityAnnotations renders the shared workload-identity oneof
// into the per-cloud ServiceAccount annotations.
func workloadIdentityAnnotations(wi *kubernetesprovider.KubernetesWorkloadIdentity) map[string]interface{} {
	if wi == nil {
		return nil
	}
	annotations := map[string]interface{}{}
	if gke := wi.GetGke(); gke != nil {
		annotations["iam.gke.io/gcp-service-account"] = gke.GetServiceAccountEmail().GetValue()
	}
	if eks := wi.GetEks(); eks != nil {
		annotations["eks.amazonaws.com/role-arn"] = eks.GetRoleArn().GetValue()
	}
	if aks := wi.GetAks(); aks != nil {
		annotations["azure.workload.identity/client-id"] = aks.GetClientId().GetValue()
		if aks.TenantId != nil {
			annotations["azure.workload.identity/tenant-id"] = aks.GetTenantId()
		}
	}
	return annotations
}

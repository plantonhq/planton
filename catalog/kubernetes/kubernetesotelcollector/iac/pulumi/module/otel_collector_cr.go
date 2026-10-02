package module

import (
	"sort"

	"github.com/pkg/errors"
	kubernetesprovider "github.com/plantonhq/planton/catalog/kubernetes"
	kubernetesotelcollectorv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesotelcollector/v1alpha1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apiextensions"
	kubernetesmeta "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"sigs.k8s.io/yaml"
)

// otelCollectorCR renders the single opentelemetry.io/v1beta1
// OpenTelemetryCollector CR the OpenTelemetry Operator (a
// KubernetesOtelOperator install, the prerequisite) reconciles into the
// collector workload (Deployment/DaemonSet/StatefulSet per mode, or
// sidecar registration), the `<name>-collector` Service with
// receiver-derived ports, the headless and monitoring Services, and the
// rendered config ConfigMap.
//
// UNTYPED CR: the body assembles in ONE place (collectorSpecBody) through
// apiextensions.CustomResource with an untyped spec map, in byte lockstep
// with the Terraform twin's locals (locals.tf `collector_spec`). Every
// key renders ONLY when the spec declares it, so the operator's
// defaulting stays authoritative for everything the manifest leaves
// unsaid.
//
// NO WAIT on the CR, deliberately: collector readiness depends on the
// operator (webhook admission, image injection, workload rollout) that
// is not part of applying the resource — the verifier owns readiness,
// the same never-block-on-a-controller posture as the sibling
// operator-CR modules.
func otelCollectorCR(ctx *pulumi.Context, locals *Locals,
	kubernetesProvider pulumi.ProviderResource,
	dependencies []pulumi.ResourceOption,
) error {
	specBody, err := collectorSpecBody(locals)
	if err != nil {
		return err
	}

	_, err = apiextensions.NewCustomResource(ctx, locals.ResourceName,
		&apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String(vars.ApiVersion),
			Kind:       pulumi.String(vars.Kind),
			Metadata: &kubernetesmeta.ObjectMetaArgs{
				Name:      pulumi.String(locals.ResourceName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
				// BACKGROUND deletion, explicitly: the OPERATOR owns the
				// collector CR's cascade — its ownership references tear
				// down the workload, Services and ConfigMap. Foreground
				// propagation would block the delete on children the
				// operator keeps reconciling. Terraform twin:
				// delete_cascade = "Background" on kubectl_manifest.
				Annotations: pulumi.StringMap{
					"pulumi.com/deletionPropagationPolicy": pulumi.String("background"),
				},
			},
			OtherFields: map[string]interface{}{
				"spec": specBody,
			},
		},
		append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, dependencies...)...)
	if err != nil {
		return errors.Wrap(err, "failed to render the OpenTelemetryCollector custom resource")
	}
	return nil
}

// collectorSpecBody renders the OpenTelemetryCollector CR spec from the
// typed fields. Field names are the CRD's own JSON keys (verified against
// the pinned v1beta1 API types at operator 0.156.0). The Terraform twin
// (locals.tf `collector_spec`) renders the identical body — keep them in
// byte lockstep.
func collectorSpecBody(locals *Locals) (map[string]interface{}, error) {
	spec := locals.Spec

	body := map[string]interface{}{}

	// The CRD defaults an absent mode to deployment; rendered on
	// declaration (an explicit "deployment" re-states the default
	// harmlessly). The proto enum's value names ARE the CR's mode
	// strings.
	if spec.Mode != nil && spec.GetMode() != kubernetesotelcollectorv1alpha1.KubernetesOtelCollectorMode_kubernetes_otel_collector_mode_unspecified {
		body["mode"] = spec.GetMode().String()
	}

	// THE PIPELINE IS THE PRODUCT: config_yaml carries the collector's
	// own configuration document on its own open contract. The v1beta1
	// CR's `config` is a STRUCTURED object (not a string — that was
	// v1alpha1), so the document is parsed here and embedded as an
	// object. An unparseable document fails loudly before anything
	// applies. Twin: the Terraform module's yamldecode at plan.
	config := map[string]interface{}{}
	if err := yaml.Unmarshal([]byte(spec.GetConfigYaml()), &config); err != nil {
		return nil, errors.Wrap(err, "failed to parse config_yaml as a YAML document")
	}
	body["config"] = config

	// Workload modes only; never alongside the autoscaler (it manages
	// the count); the middleware-defaulted 1 in daemonset/sidecar modes
	// is deliberately ignored (the spec CEL's expressibility tolerance).
	if locals.isWorkloadMode() && spec.Autoscaler == nil && spec.Replicas != nil {
		body["replicas"] = int(spec.GetReplicas())
	}

	if locals.isWorkloadMode() && spec.Autoscaler != nil {
		autoscaler := map[string]interface{}{
			"maxReplicas": int(spec.GetAutoscaler().GetMaxReplicas()),
		}
		if spec.GetAutoscaler().MinReplicas != nil {
			autoscaler["minReplicas"] = int(spec.GetAutoscaler().GetMinReplicas())
		}
		if spec.GetAutoscaler().TargetCpuUtilization != nil {
			autoscaler["targetCPUUtilization"] = int(spec.GetAutoscaler().GetTargetCpuUtilization())
		}
		if spec.GetAutoscaler().TargetMemoryUtilization != nil {
			autoscaler["targetMemoryUtilization"] = int(spec.GetAutoscaler().GetTargetMemoryUtilization())
		}
		body["autoscaler"] = autoscaler
	}

	// Empty = the operator injects its default collector image
	// (fleet-wide override on the operator kind's
	// default_collector_image).
	if spec.GetImage() != "" {
		body["image"] = spec.GetImage()
	}

	// Empty = the operator creates a default ServiceAccount. Set when the
	// pipeline reads cluster state (see the spec's PERMISSIONS note).
	if spec.GetServiceAccount() != "" {
		body["serviceAccount"] = spec.GetServiceAccount()
	}

	// Plain env vars, sorted by name — deterministic CR bodies on both
	// engines (Go maps iterate randomly; the sort makes the contract
	// explicit). Twin: the Terraform module's sort(keys(...)).
	if len(spec.GetEnv()) > 0 {
		keys := make([]string, 0, len(spec.GetEnv()))
		for k := range spec.GetEnv() {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		env := make([]interface{}, 0, len(keys))
		for _, k := range keys {
			env = append(env, map[string]interface{}{"name": k, "value": spec.GetEnv()[k]})
		}
		body["env"] = env
	}

	// Secrets loaded whole as environment variables (the credential path
	// — referenced in config as ${env:VAR_NAME}; nothing secret-bearing
	// ever lands in the rendered config document).
	if len(spec.GetEnvFromSecrets()) > 0 {
		envFrom := make([]interface{}, 0, len(spec.GetEnvFromSecrets()))
		for _, s := range spec.GetEnvFromSecrets() {
			envFrom = append(envFrom, map[string]interface{}{
				"secretRef": map[string]interface{}{"name": s},
			})
		}
		body["envFrom"] = envFrom
	}

	// The shared VolumeMount message carries BOTH halves the CR splits:
	// `volumes` (pod volume sources) and `volumeMounts` (container
	// mounts).
	if len(spec.GetVolumes()) > 0 {
		volumes := make([]interface{}, 0, len(spec.GetVolumes()))
		mounts := make([]interface{}, 0, len(spec.GetVolumes()))
		for _, v := range spec.GetVolumes() {
			volumes = append(volumes, volumeBody(v))
			mounts = append(mounts, volumeMountBody(v))
		}
		body["volumes"] = volumes
		body["volumeMounts"] = mounts
	}

	// Extra Service ports — only for receivers the operator cannot infer
	// (it derives the standard components' ports from the config itself).
	if len(spec.GetAdditionalPorts()) > 0 {
		ports := make([]interface{}, 0, len(spec.GetAdditionalPorts()))
		for _, p := range spec.GetAdditionalPorts() {
			port := map[string]interface{}{
				"name": p.GetName(),
				"port": int(p.GetPort()),
			}
			if p.Protocol != nil && p.GetProtocol() != "" {
				port["protocol"] = p.GetProtocol()
			}
			ports = append(ports, port)
		}
		body["ports"] = ports
	}

	if resources := resourcesBody(spec.GetResources()); len(resources) > 0 {
		body["resources"] = resources
	}

	if sched := spec.GetScheduling(); sched != nil {
		if len(sched.GetNodeSelector()) > 0 {
			nodeSelector := map[string]interface{}{}
			for k, v := range sched.GetNodeSelector() {
				nodeSelector[k] = v
			}
			body["nodeSelector"] = nodeSelector
		}
		if len(sched.GetTolerations()) > 0 {
			body["tolerations"] = tolerationsBody(sched.GetTolerations())
		}
		if sched.GetPriorityClassName() != "" {
			body["priorityClassName"] = sched.GetPriorityClassName()
		}
	}

	if psc := podSecurityContextBody(spec.GetPodSecurityContext()); len(psc) > 0 {
		body["podSecurityContext"] = psc
	}

	// The operator then creates the monitor itself (a ServiceMonitor, or a
	// PodMonitor in sidecar mode); absent, it creates none.
	if spec.GetServiceMonitorEnabled() {
		body["observability"] = map[string]interface{}{
			"metrics": map[string]interface{}{"enableMetrics": true},
		}
	}

	return body, nil
}

// podSecurityContextBody renders the shared WorkloadPodSecurityContext
// message into the CR's podSecurityContext shape. The daemonset
// log-collection pattern typically needs runAsUser 0 — container
// runtimes write pod log files readable only by root (taught on the
// spec field). Twin: the Terraform module's pod_security_context.
func podSecurityContextBody(p *kubernetesprovider.WorkloadPodSecurityContext) map[string]interface{} {
	out := map[string]interface{}{}
	if p == nil {
		return out
	}
	if p.RunAsUser != nil {
		out["runAsUser"] = p.GetRunAsUser()
	}
	if p.RunAsGroup != nil {
		out["runAsGroup"] = p.GetRunAsGroup()
	}
	if p.RunAsNonRoot != nil {
		out["runAsNonRoot"] = p.GetRunAsNonRoot()
	}
	if p.FsGroup != nil {
		out["fsGroup"] = p.GetFsGroup()
	}
	if p.GetFsGroupChangePolicy() != "" {
		out["fsGroupChangePolicy"] = p.GetFsGroupChangePolicy()
	}
	if len(p.GetSupplementalGroups()) > 0 {
		groups := make([]interface{}, 0, len(p.GetSupplementalGroups()))
		for _, g := range p.GetSupplementalGroups() {
			groups = append(groups, g)
		}
		out["supplementalGroups"] = groups
	}
	if len(p.GetSysctls()) > 0 {
		sysctls := make([]interface{}, 0, len(p.GetSysctls()))
		for _, s := range p.GetSysctls() {
			sysctls = append(sysctls, map[string]interface{}{
				"name":  s.GetName(),
				"value": s.GetValue(),
			})
		}
		out["sysctls"] = sysctls
	}
	if sp := p.GetSeccompProfile(); sp != nil {
		seccomp := map[string]interface{}{}
		if sp.GetType() != "" {
			seccomp["type"] = sp.GetType()
		}
		if sp.GetLocalhostProfile() != "" {
			seccomp["localhostProfile"] = sp.GetLocalhostProfile()
		}
		if len(seccomp) > 0 {
			out["seccompProfile"] = seccomp
		}
	}
	return out
}

// volumeBody renders one shared-VolumeMount entry's VOLUME half (the pod
// volume source). Exactly one source block per entry (spec-documented);
// single-key config_map/secret entries project just that key via items.
// Twin: the Terraform module's cr_volumes.
func volumeBody(v *kubernetesprovider.VolumeMount) map[string]interface{} {
	volume := map[string]interface{}{"name": v.GetName()}
	switch {
	case v.GetConfigMap() != nil:
		cm := map[string]interface{}{"name": v.GetConfigMap().GetName()}
		if v.GetConfigMap().GetDefaultMode() != 0 {
			cm["defaultMode"] = int(v.GetConfigMap().GetDefaultMode())
		}
		if v.GetConfigMap().GetKey() != "" {
			path := v.GetConfigMap().GetPath()
			if path == "" {
				path = v.GetConfigMap().GetKey()
			}
			cm["items"] = []interface{}{map[string]interface{}{
				"key":  v.GetConfigMap().GetKey(),
				"path": path,
			}}
		}
		volume["configMap"] = cm
	case v.GetSecret() != nil:
		secret := map[string]interface{}{"secretName": v.GetSecret().GetName()}
		if v.GetSecret().GetDefaultMode() != 0 {
			secret["defaultMode"] = int(v.GetSecret().GetDefaultMode())
		}
		if v.GetSecret().GetKey() != "" {
			path := v.GetSecret().GetPath()
			if path == "" {
				path = v.GetSecret().GetKey()
			}
			secret["items"] = []interface{}{map[string]interface{}{
				"key":  v.GetSecret().GetKey(),
				"path": path,
			}}
		}
		volume["secret"] = secret
	case v.GetHostPath() != nil:
		hostPath := map[string]interface{}{"path": v.GetHostPath().GetPath()}
		if v.GetHostPath().GetType() != "" {
			hostPath["type"] = v.GetHostPath().GetType()
		}
		volume["hostPath"] = hostPath
	case v.GetEmptyDir() != nil:
		emptyDir := map[string]interface{}{}
		if v.GetEmptyDir().GetMedium() != "" {
			emptyDir["medium"] = v.GetEmptyDir().GetMedium()
		}
		if v.GetEmptyDir().GetSizeLimit() != "" {
			emptyDir["sizeLimit"] = v.GetEmptyDir().GetSizeLimit()
		}
		volume["emptyDir"] = emptyDir
	case v.GetPvc() != nil:
		pvc := map[string]interface{}{"claimName": v.GetPvc().GetClaimName()}
		if v.GetPvc().GetReadOnly() {
			pvc["readOnly"] = true
		}
		volume["persistentVolumeClaim"] = pvc
	case v.GetServiceAccountToken() != nil:
		// Kubelet-rotated, audience-bound ServiceAccount tokens ride the
		// "projected" volume type in the Kubernetes API.
		path := v.GetServiceAccountToken().GetPath()
		if path == "" {
			path = "token"
		}
		token := map[string]interface{}{
			"audience": v.GetServiceAccountToken().GetAudience(),
			"path":     path,
		}
		if v.GetServiceAccountToken().GetExpirationSeconds() > 0 {
			token["expirationSeconds"] = int(v.GetServiceAccountToken().GetExpirationSeconds())
		}
		volume["projected"] = map[string]interface{}{
			"sources": []interface{}{map[string]interface{}{
				"serviceAccountToken": token,
			}},
		}
	}
	return volume
}

// volumeMountBody renders one shared-VolumeMount entry's MOUNT half (the
// container volumeMount). Twin: the Terraform module's cr_volume_mounts.
func volumeMountBody(v *kubernetesprovider.VolumeMount) map[string]interface{} {
	mount := map[string]interface{}{
		"name":      v.GetName(),
		"mountPath": v.GetMountPath(),
	}
	if v.GetReadOnly() {
		mount["readOnly"] = true
	}
	if v.GetSubPath() != "" {
		mount["subPath"] = v.GetSubPath()
	}
	return mount
}

// resourcesBody renders the shared ContainerResources message into the
// CR's resources shape. Returns an empty map when nothing is set.
func resourcesBody(r *kubernetesprovider.ContainerResources) map[string]interface{} {
	out := map[string]interface{}{}
	if r == nil {
		return out
	}
	if l := r.GetLimits(); l != nil && (l.GetCpu() != "" || l.GetMemory() != "") {
		limits := map[string]interface{}{}
		if l.GetCpu() != "" {
			limits["cpu"] = l.GetCpu()
		}
		if l.GetMemory() != "" {
			limits["memory"] = l.GetMemory()
		}
		out["limits"] = limits
	}
	if q := r.GetRequests(); q != nil && (q.GetCpu() != "" || q.GetMemory() != "") {
		requests := map[string]interface{}{}
		if q.GetCpu() != "" {
			requests["cpu"] = q.GetCpu()
		}
		if q.GetMemory() != "" {
			requests["memory"] = q.GetMemory()
		}
		out["requests"] = requests
	}
	return out
}

// tolerationsBody renders the shared WorkloadToleration list into the
// CR's tolerations shape.
func tolerationsBody(tolerations []*kubernetesprovider.WorkloadToleration) []interface{} {
	out := make([]interface{}, 0, len(tolerations))
	for _, t := range tolerations {
		tol := map[string]interface{}{}
		if t.GetKey() != "" {
			tol["key"] = t.GetKey()
		}
		if t.GetOperator() != "" {
			tol["operator"] = t.GetOperator()
		}
		if t.GetValue() != "" {
			tol["value"] = t.GetValue()
		}
		if t.GetEffect() != "" {
			tol["effect"] = t.GetEffect()
		}
		if t.TolerationSeconds != nil {
			tol["tolerationSeconds"] = t.GetTolerationSeconds()
		}
		out = append(out, tol)
	}
	return out
}

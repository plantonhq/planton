package module

import (
	kuberneteslangfusev1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kuberneteslangfuse/v1alpha1"
)

// buildTypedValues renders the typed spec into the chart's values. The
// helm_values escape hatch and the re-pins are merged over it in
// helm_release.go.
//
// PARITY: the Terraform module renders the same document as
// local.typed_helm_values. Keep every mapping below in lockstep.
//
// CHART FACTS (verified at chart 2.1.5) the mapping relies on:
//   - the four bundled stores are switched off (`deploy: false`) and every
//     store is wired by host and existingSecret;
//   - each s3.<uploadType> key falls back to the top-level s3.* key through
//     Helm's `default`, which reads false as unset: the top-level region and
//     endpoint are pinned empty and forcePathStyle false, and every upload
//     type renders its own values explicitly;
//   - media uploads have no off switch: the chart always renders the media
//     bucket;
//   - salt, encryptionKey, nextauth.secret and licenseKey take
//     {secretKeyRef}; given one, the chart's `<release>-app` Secret holds no
//     generated value;
//   - additionalEnv lists are concatenated (global first, then the
//     component's own) into each container's env.
func buildTypedValues(locals *Locals) map[string]interface{} {
	spec := locals.Spec
	instanceRef := func(key string) map[string]interface{} {
		return map[string]interface{}{"secretKeyRef": map[string]interface{}{"name": locals.InstanceSecretName, "key": key}}
	}

	// ---- langfuse (application settings shared by both workloads) ----------
	features := map[string]interface{}{
		"telemetryEnabled": spec.GetFeatures() == nil || spec.GetFeatures().TelemetryEnabled == nil || spec.GetFeatures().GetTelemetryEnabled(),
		"signUpDisabled":   spec.GetAuth().GetDisableSignUp(),
	}
	app := map[string]interface{}{
		"salt":          instanceRef(keySalt),
		"encryptionKey": instanceRef(keyEncryptionKey),
		"nextauth": map[string]interface{}{
			"url":    locals.ExternalUrl,
			"secret": instanceRef(keyNextauthSecret),
		},
		"features":       features,
		"serviceAccount": serviceAccountValues(spec.GetServiceAccount()),
		"web":            webValues(locals),
		"worker":         workerValues(locals),
	}
	logging := map[string]interface{}{}
	if lvl := spec.GetObservability().GetLogLevel(); lvl != kuberneteslangfusev1alpha1.KubernetesLangfuseObservability_log_level_unspecified {
		logging["level"] = lvl.String()
	}
	if f := spec.GetObservability().GetLogFormat(); f != kuberneteslangfusev1alpha1.KubernetesLangfuseObservability_log_format_unspecified {
		logging["format"] = f.String()
	}
	if len(logging) > 0 {
		app["logging"] = logging
	}
	if locals.DeclaredSecrets[keyLicenseKey] != "" {
		app["licenseKey"] = instanceRef(keyLicenseKey)
	}
	if creators := spec.GetFeatures().GetAllowedOrganizationCreators(); len(creators) > 0 {
		app["allowedOrganizationCreators"] = stringSlice(creators)
	}
	if spec.GetAuth().GetDisablePasswordSignIn() {
		app["auth"] = map[string]interface{}{"disableUsernamePassword": true}
	}
	if psc := podSecurityContextBody(spec.GetPodSecurityContext()); len(psc) > 0 {
		app["podSecurityContext"] = psc
	}
	if csc := containerSecurityContextBody(spec.GetContainerSecurityContext()); len(csc) > 0 {
		app["securityContext"] = csc
	}
	if dns := dnsConfigBody(spec.GetDnsConfig()); dns != nil {
		app["dnsConfig"] = dns
	}
	if len(locals.Volumes) > 0 {
		app["extraVolumes"] = locals.Volumes
		app["extraVolumeMounts"] = locals.VolumeMounts
	}
	if env := envBodies(locals.CommonEnv, locals.InstanceSecretName); len(env) > 0 {
		app["additionalEnv"] = env
	}

	return map[string]interface{}{
		"langfuse":   app,
		"postgresql": postgresValues(spec.GetPostgres()),
		"redis":      redisValues(locals),
		"clickhouse": clickhouseValues(spec.GetClickhouse()),
		"s3":         blobValues(locals),
	}
}

func serviceAccountValues(sa *kuberneteslangfusev1alpha1.KubernetesLangfuseServiceAccount) map[string]interface{} {
	if sa == nil {
		sa = &kuberneteslangfusev1alpha1.KubernetesLangfuseServiceAccount{}
	}
	out := map[string]interface{}{
		"automountServiceAccountToken": sa.AutomountToken == nil || sa.GetAutomountToken(),
	}
	if sa.GetExistingName() != "" {
		out["create"] = false
		out["name"] = sa.GetExistingName()
		return out
	}
	annotations := stringMap(sa.GetAnnotations())
	wi := sa.GetWorkloadIdentity()
	switch {
	case wi.GetGke() != nil:
		annotations["iam.gke.io/gcp-service-account"] = wi.GetGke().GetServiceAccountEmail().GetValue()
	case wi.GetEks() != nil:
		annotations["eks.amazonaws.com/role-arn"] = wi.GetEks().GetRoleArn().GetValue()
	case wi.GetAks() != nil:
		annotations["azure.workload.identity/client-id"] = wi.GetAks().GetClientId().GetValue()
		if wi.GetAks().GetTenantId() != "" {
			annotations["azure.workload.identity/tenant-id"] = wi.GetAks().GetTenantId()
		}
	}
	if len(annotations) > 0 {
		out["annotations"] = annotations
	}
	return out
}

func postgresValues(pg *kuberneteslangfusev1alpha1.KubernetesLangfusePostgres) map[string]interface{} {
	port := 5432
	if pg.Port != nil {
		port = int(pg.GetPort())
	}
	out := map[string]interface{}{
		"deploy": false,
		"host":   pg.GetHost().GetValue(),
		"port":   port,
		"auth": map[string]interface{}{
			"username":       pg.GetUsername(),
			"existingSecret": pg.GetPasswordSecret().GetSecretName().GetValue(),
			"secretKeys": map[string]interface{}{
				"userPasswordKey": stringOr(pg.GetPasswordSecret().SecretKey, "password"),
			},
			"database": stringOr(pg.Database, "langfuse"),
		},
		"migration": map[string]interface{}{
			"autoMigrate": pg.AutoMigrate == nil || pg.GetAutoMigrate(),
		},
	}
	setString(out, "args", pg.GetArgs())
	return out
}

func clickhouseValues(ch *kuberneteslangfusev1alpha1.KubernetesLangfuseClickHouse) map[string]interface{} {
	// The chart selects TLS for queries from an https:// prefix on the host
	// and strips it again for the migration URL.
	host := ch.GetHost().GetValue()
	if ch.GetSecure() {
		host = "https://" + host
	}
	httpPort, nativePort := 8123, 9000
	if ch.HttpPort != nil {
		httpPort = int(ch.GetHttpPort())
	}
	if ch.NativePort != nil {
		nativePort = int(ch.GetNativePort())
	}
	secretKey := ch.GetPasswordSecret().GetSecretKey()
	if secretKey == "" {
		secretKey = ch.GetUsername()
	}
	return map[string]interface{}{
		"deploy":     false,
		"host":       host,
		"httpPort":   httpPort,
		"nativePort": nativePort,
		"database":   stringOr(ch.Database, "default"),
		"auth": map[string]interface{}{
			"username":          ch.GetUsername(),
			"existingSecret":    ch.GetPasswordSecret().GetSecretName().GetValue(),
			"existingSecretKey": secretKey,
		},
		"migration": map[string]interface{}{
			"ssl":         ch.GetSecure(),
			"autoMigrate": ch.AutoMigrate == nil || ch.GetAutoMigrate(),
		},
		"cluster": map[string]interface{}{
			"enabled": ch.GetCluster() == nil || ch.GetCluster().Enabled == nil || ch.GetCluster().GetEnabled(),
		},
	}
}

func redisValues(locals *Locals) map[string]interface{} {
	redis := locals.Spec.GetRedis()
	auth := map[string]interface{}{
		// The chart builds the connection string with `<username>:` only
		// when a username is set; null removes the chart's "default".
		"username":                  nil,
		"existingSecret":            "",
		"existingSecretPasswordKey": "",
		"database":                  int(redis.GetStandalone().GetDatabase()),
	}
	if u := redis.GetUsername().GetValue(); u != "" {
		auth["username"] = u
	}
	if ps := redis.GetPasswordSecret(); ps != nil {
		auth["existingSecret"] = ps.GetSecretName().GetValue()
		auth["existingSecretPasswordKey"] = ps.GetSecretKey().GetValue()
	}
	tls := map[string]interface{}{"enabled": redis.GetTls().GetEnabled()}
	setString(tls, "caPath", locals.RedisTlsCaPath)
	setString(tls, "certPath", locals.RedisTlsCertPath)
	setString(tls, "keyPath", locals.RedisTlsKeyPath)

	out := map[string]interface{}{
		"deploy": false,
		"auth":   auth,
		"tls":    tls,
		"cluster": map[string]interface{}{
			"enabled": redis.GetCluster() != nil,
			"nodes":   stringSlice(redis.GetCluster().GetNodes()),
		},
		"sentinel": map[string]interface{}{"enabled": false},
	}
	if s := redis.GetStandalone(); s != nil {
		port := 6379
		if s.Port != nil {
			port = int(s.GetPort())
		}
		out["host"] = s.GetHost().GetValue()
		out["port"] = port
	}
	if s := redis.GetSentinel(); s != nil {
		sentinel := map[string]interface{}{
			"enabled":    true,
			"masterName": s.GetMasterName(),
			"nodes":      csv(s.GetNodes()),
		}
		setString(sentinel, "username", s.GetUsername())
		if ps := s.GetPasswordSecret(); ps != nil {
			sentinel["existingSecret"] = ps.GetName()
			sentinel["existingSecretPasswordKey"] = ps.GetKey()
		}
		out["sentinel"] = sentinel
	}
	return out
}

func blobValues(locals *Locals) map[string]interface{} {
	storage := locals.Spec.GetBlobStorage()
	event := storage.GetEventUpload()
	if event == nil {
		event = &kuberneteslangfusev1alpha1.KubernetesLangfuseEventUpload{}
	}
	mediaUpload := storage.GetMediaUpload()
	if mediaUpload == nil {
		mediaUpload = &kuberneteslangfusev1alpha1.KubernetesLangfuseMediaUpload{}
	}
	batchExport := storage.GetBatchExport()
	if batchExport == nil {
		batchExport = &kuberneteslangfusev1alpha1.KubernetesLangfuseBatchExport{}
	}
	reads, writes := 50, 50
	if storage.ConcurrentReads != nil {
		reads = int(storage.GetConcurrentReads())
	}
	if storage.ConcurrentWrites != nil {
		writes = int(storage.GetConcurrentWrites())
	}

	media := locals.MediaTarget.values(stringOr(mediaUpload.Prefix, "media/"))
	if pe := mediaUpload.GetPublicEndpoint(); pe != "" {
		// Browsers and SDKs get presigned URLs for the public endpoint;
		// Langfuse reads media through the target's own endpoint
		// (LANGFUSE_S3_MEDIA_UPLOAD_INTERNAL_ENDPOINT in the environment).
		media["endpoint"] = pe
	}
	media["maxContentLength"] = int64OrDefault(mediaUpload.MaxContentLengthBytes, 1000000000)
	media["downloadUrlExpirySeconds"] = intOrDefault(mediaUpload.DownloadUrlExpirySeconds, 3600)

	batch := locals.BatchTarget.values(stringOr(batchExport.Prefix, "exports/"))
	batch["enabled"] = batchExport.Enabled == nil || batchExport.GetEnabled()

	out := map[string]interface{}{
		"deploy":          false,
		"storageProvider": locals.EventTarget.Family,
		"bucket":          locals.EventTarget.Bucket,
		"region":          "",
		"endpoint":        "",
		"forcePathStyle":  false,
		"concurrency":     map[string]interface{}{"reads": reads, "writes": writes},
		"eventUpload":     locals.EventTarget.values(stringOr(event.Prefix, "events/")),
		"mediaUpload":     media,
		"batchExport":     batch,
	}
	if locals.DeclaredSecrets[keyGcsCredentials] != "" {
		out["gcs"] = map[string]interface{}{"credentials": map[string]interface{}{
			"secretKeyRef": map[string]interface{}{"name": locals.InstanceSecretName, "key": keyGcsCredentials},
		}}
	}
	return out
}

// componentValues renders what the web server and the worker share in
// shape: image, replicas, resources, autoscaling, disruption budget, pod
// placement and metadata, rollout, and the component's own environment.
func componentValues(locals *Locals, component string,
	image *kuberneteslangfusev1alpha1.KubernetesLangfuseImage,
	replicas *int32,
	resources map[string]interface{},
	autoscaling *kuberneteslangfusev1alpha1.KubernetesLangfuseAutoscaling,
	pdb *kuberneteslangfusev1alpha1.KubernetesLangfusePdb,
	scheduling *kuberneteslangfusev1alpha1.KubernetesLangfuseScheduling,
	workload *kuberneteslangfusev1alpha1.KubernetesLangfuseWorkloadShape,
	env []envVar,
) map[string]interface{} {
	if workload == nil {
		workload = &kuberneteslangfusev1alpha1.KubernetesLangfuseWorkloadShape{}
	}
	out := map[string]interface{}{
		"replicas": intOrDefault(replicas, 1),
		"pdb":      pdbValues(pdb),
	}
	if img := imageValues(image); len(img) > 0 {
		out["image"] = img
	}
	if resources != nil {
		out["resources"] = resources
	}
	if hpa := autoscaling.GetHpa(); hpa != nil {
		h := map[string]interface{}{
			"enabled":                        true,
			"minReplicas":                    intOrDefault(hpa.MinReplicas, 1),
			"maxReplicas":                    intOrDefault(hpa.MaxReplicas, 2),
			"targetCPUUtilizationPercentage": intOrDefault(hpa.TargetCpuUtilizationPercent, 50),
		}
		setInt(h, "targetMemoryUtilizationPercentage", hpa.TargetMemoryUtilizationPercent)
		if len(hpa.GetAnnotations()) > 0 {
			h["annotations"] = stringMap(hpa.GetAnnotations())
		}
		out["hpa"] = h
	}
	if keda := autoscaling.GetKeda(); keda != nil {
		k := map[string]interface{}{
			"enabled":         true,
			"minReplicas":     intOrDefault(keda.MinReplicas, 1),
			"maxReplicas":     intOrDefault(keda.MaxReplicas, 2),
			"pollingInterval": intOrDefault(keda.PollingIntervalSeconds, 30),
			"triggerType":     "cpu",
			"metricType":      "Utilization",
			"value":           stringOr(keda.TargetValue, "50"),
		}
		if keda.GetTrigger() == kuberneteslangfusev1alpha1.KubernetesLangfuseKeda_memory {
			k["triggerType"] = "memory"
		}
		if keda.GetMetricType() == kuberneteslangfusev1alpha1.KubernetesLangfuseKeda_AverageValue {
			k["metricType"] = "AverageValue"
		}
		setString(k, "containerName", keda.GetContainerName())
		out["keda"] = k
	}
	if vpa := autoscaling.GetVpa(); vpa != nil {
		v := map[string]interface{}{
			"enabled":      true,
			"updatePolicy": map[string]interface{}{"updateMode": stringOr(vpa.UpdateMode, "Auto")},
		}
		if len(vpa.GetControlledResources()) > 0 {
			v["controlledResources"] = stringSlice(vpa.GetControlledResources())
		}
		if m := cpuMemory(vpa.GetMaxAllowed()); m != nil {
			v["maxAllowed"] = m
		}
		if m := cpuMemory(vpa.GetMinAllowed()); m != nil {
			v["minAllowed"] = m
		}
		out["vpa"] = v
	}

	// Pod metadata: the component's own labels and annotations, the
	// Planton identity labels (they win), and the instance Secret checksum.
	podLabels := stringMap(workload.GetPodLabels())
	for k, v := range locals.Labels {
		podLabels[k] = v
	}
	podAnnotations := stringMap(workload.GetPodAnnotations())
	podAnnotations[vars.InstanceSecretChecksumAnnotation] = locals.SecretChecksum
	pod := map[string]interface{}{
		"labels":      podLabels,
		"annotations": podAnnotations,
	}
	if len(scheduling.GetNodeSelector()) > 0 {
		pod["nodeSelector"] = stringMap(scheduling.GetNodeSelector())
	}
	if len(scheduling.GetTolerations()) > 0 {
		pod["tolerations"] = tolerationsBody(scheduling.GetTolerations())
	}
	if aff := affinityBody(scheduling.GetNodeAffinity(), scheduling.GetPodAffinity(), scheduling.GetPodAntiAffinity()); aff != nil {
		pod["affinity"] = aff
	}
	if len(scheduling.GetTopologySpreadConstraints()) > 0 {
		pod["topologySpreadConstraints"] = topologySpreadBody(scheduling.GetTopologySpreadConstraints(), map[string]string{
			"app.kubernetes.io/name":     vars.HelmChartName,
			"app.kubernetes.io/instance": locals.ReleaseName,
			"app":                        component,
		})
	}
	setString(pod, "priorityClassName", scheduling.GetPriorityClassName())
	if e := envBodies(env, locals.InstanceSecretName); len(e) > 0 {
		pod["additionalEnv"] = e
	}
	out["pod"] = pod

	deployment := map[string]interface{}{}
	if len(workload.GetDeploymentLabels()) > 0 {
		deployment["additionalLabels"] = stringMap(workload.GetDeploymentLabels())
	}
	if len(workload.GetDeploymentAnnotations()) > 0 {
		deployment["annotations"] = stringMap(workload.GetDeploymentAnnotations())
	}
	if s := strategyBody(workload.GetStrategy()); s != nil {
		deployment["strategy"] = s
	}
	if len(deployment) > 0 {
		out["deployment"] = deployment
	}
	setInt(out, "revisionHistoryLimit", workload.RevisionHistoryLimit)
	if len(workload.GetHostAliases()) > 0 {
		out["hostAliases"] = hostAliasesBody(workload.GetHostAliases())
	}
	return out
}

func webValues(locals *Locals) map[string]interface{} {
	web := locals.Spec.GetWeb()
	if web == nil {
		web = &kuberneteslangfusev1alpha1.KubernetesLangfuseWeb{}
	}
	out := componentValues(locals, "web", web.GetImage(), web.Replicas, resourcesBody(web.GetResources()),
		web.GetAutoscaling(), web.GetPdb(), web.GetScheduling(), web.GetWorkload(), locals.WebEnv)
	if p := probeValues(web.GetLivenessProbe()); len(p) > 0 {
		out["livenessProbe"] = p
	}
	if p := probeValues(web.GetReadinessProbe()); len(p) > 0 {
		out["readinessProbe"] = p
	}
	svc := web.GetService()
	if svc == nil {
		svc = &kuberneteslangfusev1alpha1.KubernetesLangfuseWebService{}
	}
	service := map[string]interface{}{
		"type": stringOr(svc.Type, "ClusterIP"),
		"port": locals.WebPort,
	}
	setInt(service, "externalPort", svc.ExternalPort)
	setInt(service, "nodePort", svc.NodePort)
	if len(svc.GetAnnotations()) > 0 {
		service["annotations"] = stringMap(svc.GetAnnotations())
	}
	if len(svc.GetLabels()) > 0 {
		service["additionalLabels"] = stringMap(svc.GetLabels())
	}
	out["service"] = service
	return out
}

func workerValues(locals *Locals) map[string]interface{} {
	worker := locals.Spec.GetWorker()
	if worker == nil {
		worker = &kuberneteslangfusev1alpha1.KubernetesLangfuseWorker{}
	}
	out := componentValues(locals, "worker", worker.GetImage(), worker.Replicas, resourcesBody(worker.GetResources()),
		worker.GetAutoscaling(), worker.GetPdb(), worker.GetScheduling(), worker.GetWorkload(), locals.WorkerEnv)
	probe := probeValues(worker.GetLivenessProbe())
	if worker.GetFailLivenessIfQueueConsumptionStuck() {
		probe["path"] = vars.WorkerLivenessStuckPath
	}
	if len(probe) > 0 {
		out["livenessProbe"] = probe
	}
	return out
}

func imageValues(img *kuberneteslangfusev1alpha1.KubernetesLangfuseImage) map[string]interface{} {
	out := map[string]interface{}{}
	setString(out, "repository", img.GetRepository())
	setString(out, "tag", img.GetTag())
	setString(out, "pullPolicy", img.GetPullPolicy())
	if len(img.GetPullSecretNames()) > 0 {
		secrets := make([]interface{}, 0, len(img.GetPullSecretNames()))
		for _, n := range img.GetPullSecretNames() {
			secrets = append(secrets, map[string]interface{}{"name": n})
		}
		out["pullSecrets"] = secrets
	}
	return out
}

func pdbValues(pdb *kuberneteslangfusev1alpha1.KubernetesLangfusePdb) map[string]interface{} {
	if pdb == nil {
		pdb = &kuberneteslangfusev1alpha1.KubernetesLangfusePdb{}
	}
	out := map[string]interface{}{"create": pdb.Enabled == nil || pdb.GetEnabled()}
	setString(out, "minAvailable", pdb.GetMinAvailable())
	setString(out, "maxUnavailable", pdb.GetMaxUnavailable())
	return out
}

func probeValues(p *kuberneteslangfusev1alpha1.KubernetesLangfuseProbe) map[string]interface{} {
	out := map[string]interface{}{}
	if p == nil {
		return out
	}
	setInt(out, "initialDelaySeconds", p.InitialDelaySeconds)
	setInt(out, "periodSeconds", p.PeriodSeconds)
	setInt(out, "timeoutSeconds", p.TimeoutSeconds)
	setInt(out, "failureThreshold", p.FailureThreshold)
	setInt(out, "successThreshold", p.SuccessThreshold)
	return out
}

func strategyBody(s *kuberneteslangfusev1alpha1.KubernetesLangfuseRolloutStrategy) map[string]interface{} {
	if s == nil || (s.GetType() == "" && s.GetMaxSurge() == "" && s.GetMaxUnavailable() == "") {
		return nil
	}
	out := map[string]interface{}{}
	setString(out, "type", s.GetType())
	rolling := map[string]interface{}{}
	if s.GetMaxSurge() != "" {
		rolling["maxSurge"] = intOrString(s.GetMaxSurge())
	}
	if s.GetMaxUnavailable() != "" {
		rolling["maxUnavailable"] = intOrString(s.GetMaxUnavailable())
	}
	if len(rolling) > 0 {
		out["rollingUpdate"] = rolling
	}
	return out
}

func envBodies(env []envVar, instanceSecretName string) []interface{} {
	out := make([]interface{}, 0, len(env))
	for _, v := range env {
		out = append(out, v.body(instanceSecretName))
	}
	return out
}

func intOrDefault(value *int32, fallback int) int {
	if value == nil {
		return fallback
	}
	return int(*value)
}

func int64OrDefault(value *int64, fallback int64) int64 {
	if value == nil {
		return fallback
	}
	return *value
}

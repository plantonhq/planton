package module

import (
	"github.com/pkg/errors"
	kubernetessignozv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetessignoz/v1alpha1"
	"sigs.k8s.io/yaml"
)

// buildHelmValues renders the typed spec into the chart-values map, merges
// the helm_values escape hatch over it with Helm -f semantics, and
// re-pins the fullname LAST — the exact semantic twin of the Terraform
// module's values = [typed, helm_values, re-pin] document list.
//
// SECRET DISCIPLINE (load-bearing): the ClickHouse password is NEVER
// declared or rendered — the chart reads it from the referenced Secret
// (existingSecret → secretKeyRef), so it appears in no values map and no
// preview diff. The SMTP password rides a valueFrom secretKeyRef env
// entry — a reference, never material.
func buildHelmValues(locals *Locals) (map[string]interface{}, error) {
	spec := locals.Spec

	values := map[string]interface{}{
		"fullnameOverride": locals.ReleaseName,
	}

	// ---- globals ---------------------------------------------------------
	// Every chart image here is the SPLIT registry+repository form and its
	// registry key defers to global.imageRegistry — one override reaches
	// the SigNoz server, the collector and the schema migrator.
	global := map[string]interface{}{}
	if spec.ImageRegistry != "" {
		global["imageRegistry"] = spec.ImageRegistry
	}
	if spec.ClusterName != "" {
		global["clusterName"] = spec.ClusterName
		values["clusterName"] = spec.ClusterName
	}
	if len(spec.ImagePullSecrets) > 0 {
		pullSecrets := make([]interface{}, 0, len(spec.ImagePullSecrets))
		for _, s := range spec.ImagePullSecrets {
			pullSecrets = append(pullSecrets, s)
		}
		global["imagePullSecrets"] = pullSecrets
	}
	if len(global) > 0 {
		values["global"] = global
	}

	// ---- the clickhouse connection (composed, never bundled) ---------------
	// `clickhouse.enabled: false` is a CONSTANT of this kind's
	// design: nothing ClickHouse-related ever installs — the telemetry
	// store is the composed KubernetesClickHouse the connection points at.
	values["clickhouse"] = map[string]interface{}{"enabled": false}
	values["externalClickhouse"] = clickHouseConnectionValues(spec.GetClickhouse())

	// ---- scheduling (server + collector + migrator) -------------------------
	var nodeSelector map[string]interface{}
	var tolerations []interface{}
	priorityClass := ""
	if scheduling := spec.GetScheduling(); scheduling != nil {
		if len(scheduling.NodeSelector) > 0 {
			nodeSelector = stringMapToInterface(scheduling.NodeSelector)
		}
		if len(scheduling.Tolerations) > 0 {
			tolerations = tolerationsSlice(scheduling.Tolerations)
		}
		priorityClass = scheduling.PriorityClassName
	}

	// ---- the signoz server ---------------------------------------------------
	values["signoz"] = serverValues(locals, nodeSelector, tolerations, priorityClass)

	// ---- the ingestion collector ----------------------------------------------
	values["otelCollector"] = collectorValues(spec.GetOtelCollector(), nodeSelector, tolerations, priorityClass)

	// ---- the schema migrator ----------------------------------------------------
	migrator := map[string]interface{}{}
	if nodeSelector != nil {
		migrator["nodeSelector"] = nodeSelector
	}
	if tolerations != nil {
		migrator["tolerations"] = tolerations
	}
	if len(migrator) > 0 {
		values["telemetryStoreMigrator"] = migrator
	}

	// ---- escape hatch + re-pin -----------------------------------------------
	if spec.HelmValues != "" {
		overrides := map[string]interface{}{}
		if err := yaml.Unmarshal([]byte(spec.HelmValues), &overrides); err != nil {
			return nil, errors.Wrap(err, "helm_values is not a valid YAML document")
		}
		values = mergeMaps(values, overrides)
	}

	// Re-pin the fullname LAST — the one deliberate exception to the
	// escape hatch's last-word contract. Every child name — and the
	// exported outputs built from them — derives from this fullname;
	// letting an override move it would break every output.
	values["fullnameOverride"] = locals.ReleaseName

	return values, nil
}

// clickHouseConnectionValues renders the composed telemetry-store
// connection — fully secret-native: the chart wires the referenced
// Secret as a secretKeyRef.
func clickHouseConnectionValues(clickhouse *kubernetessignozv1alpha1.KubernetesSignozClickHouse) map[string]interface{} {
	clusterName := clickhouse.ClusterName.GetValue()
	if clusterName == "" {
		clusterName = "cluster"
	}
	tcpPort := int32(9000)
	if clickhouse.TcpPort != nil {
		tcpPort = clickhouse.GetTcpPort()
	}
	httpPort := int32(8123)
	if clickhouse.HttpPort != nil {
		httpPort = clickhouse.GetHttpPort()
	}

	out := map[string]interface{}{
		"host":                      clickhouse.Host.GetValue(),
		"cluster":                   clusterName,
		"tcpPort":                   tcpPort,
		"httpPort":                  httpPort,
		"user":                      clickhouse.Username,
		"existingSecret":            clickhouse.PasswordSecret.SecretName.GetValue(),
		"existingSecretPasswordKey": clickhouse.PasswordSecret.SecretKey,
	}
	if clickhouse.Secure {
		out["secure"] = true
	}
	if clickhouse.Verify {
		out["verify"] = true
	}
	return out
}

// serverValues renders the SigNoz server tier: state volume, resources,
// and the env map carrying the typed alerting/SMTP configuration. Typed
// entries WIN over the user's advanced env map (the spec's documented
// contract). Env keys follow SigNoz's own derivation
// (signoz_<section>_<key>, embedded underscores doubled).
func serverValues(locals *Locals, nodeSelector map[string]interface{}, tolerations []interface{}, priorityClass string) map[string]interface{} {
	spec := locals.Spec
	server := spec.GetServer()

	diskSize := vars.DefaultServerDiskSize
	persistence := map[string]interface{}{
		"enabled": true,
	}

	out := map[string]interface{}{}

	env := map[string]interface{}{}
	if server != nil {
		if server.GetDiskSize() != "" {
			diskSize = server.GetDiskSize()
		}
		if server.StorageClass.GetValue() != "" {
			persistence["storageClass"] = server.StorageClass.GetValue()
		}
		if resources := resourcesMap(server.GetResources()); resources != nil {
			out["resources"] = resources
		}
		for k, v := range server.Env {
			env[k] = v
		}
		if server.ExternalUrl != "" {
			env["signoz_alertmanager_signoz_external__url"] = server.ExternalUrl
		}
		if smtp := server.GetSmtp(); smtp != nil {
			env["signoz_emailing_enabled"] = "true"
			env["signoz_emailing_smtp_address"] = smtp.Address
			env["signoz_emailing_smtp_from"] = smtp.From
			if smtp.Username != "" {
				env["signoz_emailing_smtp_auth_username"] = smtp.Username
			}
			if smtp.TlsEnabled {
				env["signoz_emailing_smtp_tls_enabled"] = "true"
			}
			if passwordSecret := smtp.GetPasswordSecret(); passwordSecret != nil {
				// The chart's flexible env structure renders this as a
				// proper env-from-secret — a reference, never material.
				env["signoz_emailing_smtp_auth_password"] = map[string]interface{}{
					"valueFrom": map[string]interface{}{
						"secretKeyRef": map[string]interface{}{
							"name": passwordSecret.Name,
							"key":  passwordSecret.Key,
						},
					},
				}
			}
		}
	}

	persistence["size"] = diskSize
	out["persistence"] = persistence
	if len(env) > 0 {
		out["env"] = env
	}
	if nodeSelector != nil {
		out["nodeSelector"] = nodeSelector
	}
	if tolerations != nil {
		out["tolerations"] = tolerations
	}
	if priorityClass != "" {
		out["priorityClassName"] = priorityClass
	}
	return out
}

// collectorValues renders the ingestion collector tier. The pipeline
// receiver LISTS are always rendered from the toggles (lists replace
// under Helm merge — rendering them from one derivation is what keeps the
// Service ports and the collector pipelines in agreement by
// construction).
func collectorValues(collector *kubernetessignozv1alpha1.KubernetesSignozOtelCollector, nodeSelector map[string]interface{}, tolerations []interface{}, priorityClass string) map[string]interface{} {
	replicas := int32(1)
	jaegerEnabled := true
	zipkinEnabled := false
	httpLogsEnabled := true

	out := map[string]interface{}{}

	if collector != nil {
		if collector.Replicas != nil {
			replicas = collector.GetReplicas()
		}
		if collector.JaegerReceiverEnabled != nil {
			jaegerEnabled = collector.GetJaegerReceiverEnabled()
		}
		zipkinEnabled = collector.ZipkinReceiverEnabled
		if collector.HttpLogsReceiversEnabled != nil {
			httpLogsEnabled = collector.GetHttpLogsReceiversEnabled()
		}
		if resources := resourcesMap(collector.GetResources()); resources != nil {
			out["resources"] = resources
		}
		if collector.LowCardinalityExceptionGrouping {
			out["lowCardinalityExceptionGrouping"] = true
		}
		if autoscaling := collector.GetAutoscaling(); autoscaling != nil && autoscaling.Enabled {
			minReplicas := int32(1)
			if autoscaling.MinReplicas != nil {
				minReplicas = autoscaling.GetMinReplicas()
			}
			maxReplicas := int32(11)
			if autoscaling.MaxReplicas != nil {
				maxReplicas = autoscaling.GetMaxReplicas()
			}
			hpa := map[string]interface{}{
				"enabled":     true,
				"minReplicas": minReplicas,
				"maxReplicas": maxReplicas,
			}
			if autoscaling.TargetCpuUtilizationPercent != nil {
				hpa["targetCPUUtilizationPercentage"] = autoscaling.GetTargetCpuUtilizationPercent()
			}
			if autoscaling.TargetMemoryUtilizationPercent != nil {
				hpa["targetMemoryUtilizationPercentage"] = autoscaling.GetTargetMemoryUtilizationPercent()
			}
			out["autoscaling"] = hpa
		}
	}

	out["replicaCount"] = replicas

	// Receiver toggles: jaeger + http-logs default ON (the chart's
	// grain), zipkin defaults OFF. Only diverging Service-port toggles
	// render; the pipeline lists always render.
	ports := map[string]interface{}{}
	if zipkinEnabled {
		ports["zipkin"] = map[string]interface{}{"enabled": true}
	}
	if !jaegerEnabled {
		ports["jaeger-thrift"] = map[string]interface{}{"enabled": false}
		ports["jaeger-grpc"] = map[string]interface{}{"enabled": false}
	}
	if !httpLogsEnabled {
		ports["logsheroku"] = map[string]interface{}{"enabled": false}
		ports["logsjson"] = map[string]interface{}{"enabled": false}
	}
	if len(ports) > 0 {
		out["ports"] = ports
	}

	tracesReceivers := []interface{}{"otlp"}
	if jaegerEnabled {
		tracesReceivers = append(tracesReceivers, "jaeger")
	}
	if zipkinEnabled {
		tracesReceivers = append(tracesReceivers, "zipkin")
	}
	logsReceivers := []interface{}{"otlp"}
	if httpLogsEnabled {
		logsReceivers = append(logsReceivers, "httplogreceiver/heroku", "httplogreceiver/json")
	}

	config := map[string]interface{}{
		"service": map[string]interface{}{
			"pipelines": map[string]interface{}{
				"traces": map[string]interface{}{"receivers": tracesReceivers},
				"logs":   map[string]interface{}{"receivers": logsReceivers},
			},
		},
	}
	if zipkinEnabled {
		config["receivers"] = map[string]interface{}{
			"zipkin": map[string]interface{}{"endpoint": "0.0.0.0:9411"},
		}
	}
	out["config"] = config

	if nodeSelector != nil {
		out["nodeSelector"] = nodeSelector
	}
	if tolerations != nil {
		out["tolerations"] = tolerations
	}
	if priorityClass != "" {
		out["priorityClassName"] = priorityClass
	}
	return out
}

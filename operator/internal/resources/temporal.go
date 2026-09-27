package resources

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	TemporalHelmChartVersion = "0.62.0"
	TemporalFrontendGRPCPort = 7233
	TemporalFrontendHTTPPort = 7243
	TemporalWebUIPort        = 8080
	TemporalDefaultDB        = "temporal"
	TemporalVisibilityDB     = "temporal_visibility"
	TemporalPostgresDriver   = "postgres12"

	// The size of Temporal's auxiliary workloads -- the web UI, the admin
	// tools, and the one-shot schema Jobs -- fixed rather than registered
	// (sizing.go): a few Mi each live, none of them carrying the platform's
	// load, so one small size in the house pattern serves them all. The four
	// server services are sized from the registry, each on its own.
	temporalAuxiliaryCPURequest    = "10m"
	temporalAuxiliaryMemoryRequest = "32Mi"
	temporalAuxiliaryMemoryLimit   = "256Mi"
)

// TemporalHelmOptions is everything the chart's values are rendered from.
type TemporalHelmOptions struct {
	CRName    string
	Namespace string

	// The four server services' effective sizing (SizingTemporalFrontend,
	// -History, -Matching, and -Worker in the registry, each merged with its
	// own override by the component).
	Frontend corev1.ResourceRequirements
	History  corev1.ResourceRequirements
	Matching corev1.ResourceRequirements
	Worker   corev1.ResourceRequirements
}

// TemporalHelmValues builds the Helm values map for rendering the Temporal
// chart. The PostgreSQL connection details come from the shared connection
// seam; Temporal's own schema job creates its two databases (the
// createDatabase toggle below), which is why the connection user must be able
// to CREATE DATABASE.
func TemporalHelmValues(opts TemporalHelmOptions) map[string]any {
	crName := opts.CRName
	conn := PostgreSQLConnection(crName, opts.Namespace)

	sqlConfig := func(database string) map[string]any {
		return map[string]any{
			"driver": "sql",
			"sql": map[string]any{
				"driver":         TemporalPostgresDriver,
				"host":           conn.Host,
				"port":           conn.Port,
				"database":       database,
				"user":           conn.User,
				"existingSecret": conn.SecretName,
				"secretKey":      conn.PassKey,
			},
		}
	}

	return map[string]any{
		"fullnameOverride": fmt.Sprintf("%s-temporal", crName),

		"cassandra":  map[string]any{"enabled": false},
		"mysql":      map[string]any{"enabled": false},
		"postgresql": map[string]any{"enabled": false},

		"server": map[string]any{
			"config": map[string]any{
				"persistence": map[string]any{
					"driver":     "sql",
					"default":    sqlConfig(TemporalDefaultDB),
					"visibility": sqlConfig(TemporalVisibilityDB),
				},
			},
			// Every service names its own resources: the chart's fallback,
			// server.resources, REPLACES a service's block rather than
			// merging with it, so a shared block could not size one service
			// without restating it for the rest.
			"frontend": map[string]any{"resources": helmResourceValues(mustBeSized(SizingTemporalFrontend, opts.Frontend))},
			"history":  map[string]any{"resources": helmResourceValues(mustBeSized(SizingTemporalHistory, opts.History))},
			"matching": map[string]any{"resources": helmResourceValues(mustBeSized(SizingTemporalMatching, opts.Matching))},
			"worker":   map[string]any{"resources": helmResourceValues(mustBeSized(SizingTemporalWorker, opts.Worker))},
		},

		"schema": map[string]any{
			"createDatabase": map[string]any{"enabled": true},
			"setup":          map[string]any{"enabled": true},
			"update":         map[string]any{"enabled": true},
			"resources":      helmResourceValues(temporalAuxiliaryResources()),
		},
		"admintools": map[string]any{
			"resources": helmResourceValues(temporalAuxiliaryResources()),
		},
		"web": map[string]any{
			"resources": helmResourceValues(temporalAuxiliaryResources()),
		},

		"prometheus":          map[string]any{"enabled": false},
		"grafana":             map[string]any{"enabled": false},
		"kubePrometheusStack": map[string]any{"enabled": false},
		"elasticsearch":       map[string]any{"enabled": false},
	}
}

// TemporalFrontendServiceName returns the Kubernetes Service name for the
// Temporal frontend, which is the primary gRPC endpoint applications connect to.
func TemporalFrontendServiceName(crName string) string {
	return fmt.Sprintf("%s-temporal-frontend", crName)
}

// TemporalWebUIServiceName returns the Kubernetes Service name for the
// Temporal web UI.
func TemporalWebUIServiceName(crName string) string {
	return fmt.Sprintf("%s-temporal-web", crName)
}

// TemporalFrontendEndpoint returns the in-cluster FQDN for the Temporal
// frontend gRPC service.
func TemporalFrontendEndpoint(crName, namespace string) string {
	return fmt.Sprintf("%s.%s.svc.cluster.local:%d",
		TemporalFrontendServiceName(crName), namespace, TemporalFrontendGRPCPort)
}

// temporalAuxiliaryResources is the auxiliary workloads' fixed size (the
// constants above carry the reasoning).
func temporalAuxiliaryResources() corev1.ResourceRequirements {
	return corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(temporalAuxiliaryCPURequest),
			corev1.ResourceMemory: resource.MustParse(temporalAuxiliaryMemoryRequest),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceMemory: resource.MustParse(temporalAuxiliaryMemoryLimit),
		},
	}
}

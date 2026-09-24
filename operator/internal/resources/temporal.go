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

	// Temporal's sizing, chosen here rather than left to the chart (which
	// ships none for any of its six workloads). Read live on a one-node
	// install: history ~375Mi (it holds the mutable state and its caches),
	// matching ~145Mi, frontend ~90Mi, worker ~60Mi, the web UI and the
	// admin tools a few Mi each. One default for the four server services
	// with history sized above it; the web UI, the admin tools, and the
	// one-shot schema Jobs small. No CPU limit anywhere: a workflow burst is
	// never throttled (requests-only, the house pattern).
	temporalServerCPURequest       = "50m"
	temporalServerMemoryRequest    = "128Mi"
	temporalServerMemoryLimit      = "512Mi"
	temporalHistoryCPURequest      = "100m"
	temporalHistoryMemoryRequest   = "256Mi"
	temporalHistoryMemoryLimit     = "1Gi"
	temporalAuxiliaryCPURequest    = "10m"
	temporalAuxiliaryMemoryRequest = "32Mi"
	temporalAuxiliaryMemoryLimit   = "256Mi"
)

// TemporalHelmValues builds the Helm values map for rendering the Temporal
// chart. The PostgreSQL connection details come from the shared connection
// seam; Temporal's own schema job creates its two databases (the
// createDatabase toggle below), which is why the connection user must be able
// to CREATE DATABASE.
func TemporalHelmValues(crName, namespace string) map[string]any {
	conn := PostgreSQLConnection(crName, namespace)

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
			// The chart applies server.resources to every service that does
			// not name its own; history names its own, above the default.
			"resources": helmResourceValues(temporalServerResources()),
			"history": map[string]any{
				"resources": helmResourceValues(temporalHistoryResources()),
			},
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

// temporalServerResources is the container sizing every install gets (the constants
// above carry the reasoning).
func temporalServerResources() corev1.ResourceRequirements {
	return corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(temporalServerCPURequest),
			corev1.ResourceMemory: resource.MustParse(temporalServerMemoryRequest),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceMemory: resource.MustParse(temporalServerMemoryLimit),
		},
	}
}

// temporalHistoryResources is the container sizing every install gets (the constants
// above carry the reasoning).
func temporalHistoryResources() corev1.ResourceRequirements {
	return corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(temporalHistoryCPURequest),
			corev1.ResourceMemory: resource.MustParse(temporalHistoryMemoryRequest),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceMemory: resource.MustParse(temporalHistoryMemoryLimit),
		},
	}
}

// temporalAuxiliaryResources is the container sizing every install gets (the constants
// above carry the reasoning).
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

package resources

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	OpenFGAHelmChartVersion = "0.3.13"
	OpenFGAHTTPPort         = 8080
	OpenFGAGRPCPort         = 8081
	OpenFGADatastoreEngine  = "postgres"
	OpenFGAStoreName        = "planton"

	// The authorization engine's sizing, chosen here rather than left to the
	// chart (which ships none). A Go server answering one control plane's
	// checks is small and steady: ~30Mi resident live. A request so it
	// schedules honestly, a memory limit so a leak cannot take the node, no
	// CPU limit so a burst of checks is never throttled (requests-only, the
	// house pattern).
	openFGACPURequest    = "50m"
	openFGAMemoryRequest = "64Mi"
	openFGAMemoryLimit   = "256Mi"
)

// OpenFGAHelmValues builds the Helm values map for rendering the OpenFGA
// chart. The PostgreSQL connection details come from the shared connection
// seam. The openfga database itself is born with the platform cluster (see
// the Cluster builder's postInitSQL): OpenFGA's migrate job applies schema
// but cannot create its database, so enabling authorization at any later
// point must find the database waiting.
func OpenFGAHelmValues(crName, namespace string) map[string]any {
	conn := PostgreSQLConnection(crName, namespace)

	datastoreURI := fmt.Sprintf(
		"postgres://%s:$(%s)@%s:%d/%s?sslmode=disable",
		conn.User,
		"OPENFGA_DATASTORE_PASSWORD",
		conn.Host,
		conn.Port,
		DBOpenFGA,
	)

	return map[string]any{
		"fullnameOverride": fmt.Sprintf("%s-openfga", crName),
		"replicaCount":     1,
		"resources":        helmResourceValues(openFGAResources()),
		"datastore": map[string]any{
			"engine":          OpenFGADatastoreEngine,
			"uri":             datastoreURI,
			"applyMigrations": true,
		},
		"extraEnvVars": []any{
			map[string]any{
				"name": "OPENFGA_DATASTORE_PASSWORD",
				"valueFrom": map[string]any{
					"secretKeyRef": map[string]any{
						"name": conn.SecretName,
						"key":  conn.PassKey,
					},
				},
			},
		},
	}
}

// OpenFGAServiceName returns the Kubernetes Service name for the OpenFGA
// server deployed by this operator.
func OpenFGAServiceName(crName string) string {
	return fmt.Sprintf("%s-openfga", crName)
}

// OpenFGAServiceFQDN returns the in-cluster fully-qualified domain name for
// the OpenFGA HTTP API.
func OpenFGAServiceFQDN(crName, namespace string) string {
	return fmt.Sprintf("%s.%s.svc.cluster.local", OpenFGAServiceName(crName), namespace)
}

// OpenFGAHTTPURL returns the full HTTP URL for the OpenFGA API, usable from
// any pod in the cluster.
func OpenFGAHTTPURL(crName, namespace string) string {
	return fmt.Sprintf("http://%s:%d", OpenFGAServiceFQDN(crName, namespace), OpenFGAHTTPPort)
}

// openFGAResources is the container sizing every install gets (the constants
// above carry the reasoning).
func openFGAResources() corev1.ResourceRequirements {
	return corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(openFGACPURequest),
			corev1.ResourceMemory: resource.MustParse(openFGAMemoryRequest),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceMemory: resource.MustParse(openFGAMemoryLimit),
		},
	}
}

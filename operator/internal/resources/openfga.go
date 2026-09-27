package resources

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

const (
	OpenFGAHelmChartVersion = "0.3.13"
	OpenFGAHTTPPort         = 8080
	OpenFGAGRPCPort         = 8081
	OpenFGADatastoreEngine  = "postgres"
	OpenFGAStoreName        = "planton"
)

// OpenFGAHelmOptions is everything the chart's values are rendered from.
type OpenFGAHelmOptions struct {
	CRName    string
	Namespace string

	// Resources is the server container's effective sizing (SizingOpenFGA in
	// the registry, merged with the spec's override by the component).
	Resources corev1.ResourceRequirements
}

// OpenFGAHelmValues builds the Helm values map for rendering the OpenFGA
// chart. The PostgreSQL connection details come from the shared connection
// seam. The openfga database itself is born with the platform cluster (see
// the Cluster builder's postInitSQL): OpenFGA's migrate job applies schema
// but cannot create its database, so enabling authorization at any later
// point must find the database waiting.
func OpenFGAHelmValues(opts OpenFGAHelmOptions) map[string]any {
	crName := opts.CRName
	conn := PostgreSQLConnection(crName, opts.Namespace)

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
		"resources":        helmResourceValues(mustBeSized(SizingOpenFGA, opts.Resources)),
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

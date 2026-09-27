package resources

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

const (
	Neo4jHelmChartVersion = "2026.1.4"
	Neo4jBoltPort         = 7687
	Neo4jHTTPPort         = 7474

	// Neo4jPasswordSecretKey is the data key inside the chart-generated auth
	// Secret that holds the Neo4j password. The Neo4j Helm chart creates this
	// Secret automatically as "{releaseName}-auth".
	Neo4jPasswordSecretKey = "neo4j-password"

	// Neo4jDefaultUser is the default admin username for Neo4j Community Edition.
	Neo4jDefaultUser = "neo4j"
)

// Neo4jHelmValues builds the Helm values map for rendering the official Neo4j
// Helm chart. The chart deploys a standalone Community Edition instance.
//
// The Neo4j chart auto-generates its own password in a Secret
// ("{releaseName}-auth"), so no operator-generated credential is needed.
//
// Reference: Planton KubernetesNeo4j module (neo4j/neo4j).
//
// The volume shape follows the chart's _volumeTemplate.tpl contract: the size
// must live at volumes.data.<mode>.requests.storage (a bare volumes.data.size
// key is silently ignored), and a StorageClass may only be named in "dynamic"
// mode -- the chart REJECTS a storageClassName under "defaultStorageClass"
// mode. So the mode switches on whether a class is pinned.
//
// res is the server's effective sizing (SizingNeo4j in the registry, merged
// with the spec's override by the component). The chart takes a requests and
// limits pair whenever either key is present, and otherwise copies a bare
// cpu/memory pair into both; the requests/limits form is what lets the house
// pattern drop the CPU limit.
func Neo4jHelmValues(crName, storageSize, storageClass string, res corev1.ResourceRequirements) map[string]any {
	data := map[string]any{}
	if storageClass != "" {
		data["mode"] = "dynamic"
		data["dynamic"] = map[string]any{
			"storageClassName": storageClass,
			"accessModes":      []any{"ReadWriteOnce"},
			"requests":         map[string]any{"storage": storageSize},
		}
	} else {
		data["mode"] = "defaultStorageClass"
		data["defaultStorageClass"] = map[string]any{
			"accessModes": []any{"ReadWriteOnce"},
			"requests":    map[string]any{"storage": storageSize},
		}
	}
	return map[string]any{
		"neo4j": map[string]any{
			"name":                   crName,
			"acceptLicenseAgreement": "yes",
			"resources":              helmResourceValues(mustBeSized(SizingNeo4j, res)),
		},
		"volumes": map[string]any{
			"data": data,
		},
	}
}

// neo4jReleaseName returns the Helm release name: "{crName}-neo4j".
func neo4jReleaseName(crName string) string {
	return fmt.Sprintf("%s-neo4j", crName)
}

// Neo4jAuthSecretName returns the chart-generated auth Secret name:
// "{crName}-neo4j-auth". This Secret is created by the Neo4j Helm chart
// and contains the auto-generated password.
func Neo4jAuthSecretName(crName string) string {
	return fmt.Sprintf("%s-neo4j-auth", crName)
}

// Neo4jServiceHost returns the in-cluster DNS hostname for the Neo4j
// Service: "{crName}-neo4j.{namespace}.svc.cluster.local".
func Neo4jServiceHost(crName, namespace string) string {
	return fmt.Sprintf("%s-neo4j.%s.svc.cluster.local", crName, namespace)
}

// Neo4jBoltURI returns the bolt:// connection URI for Neo4j:
// "bolt://{crName}-neo4j.{namespace}.svc.cluster.local:7687".
func Neo4jBoltURI(crName, namespace string) string {
	return fmt.Sprintf("bolt://%s:%d", Neo4jServiceHost(crName, namespace), Neo4jBoltPort)
}

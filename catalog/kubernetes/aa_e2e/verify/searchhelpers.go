package verify

// Manifest-parsing helpers for the search-wave kinds. Scenario manifests
// are written in the protos' snake_case form; the helpers tolerate the
// camelCase twins the same way the Kafka helpers do.

// opensearchFirstPool returns the first declared node pool's name and
// the SUM of every pool's replicas (the cluster's availableNodes
// target).
func opensearchFirstPool(spec map[string]interface{}) (string, int) {
	pools, _ := spec["node_pools"].([]interface{})
	if pools == nil {
		pools, _ = spec["nodePools"].([]interface{})
	}
	firstName := ""
	total := 0
	for i, p := range pools {
		pool, _ := p.(map[string]interface{})
		if pool == nil {
			continue
		}
		if i == 0 {
			firstName, _ = pool["name"].(string)
		}
		switch r := pool["replicas"].(type) {
		case int:
			total += r
		case float64:
			total += int(r)
		}
	}
	return firstName, total
}

func opensearchDashboardsEnabled(spec map[string]interface{}) bool {
	d, _ := spec["dashboards"].(map[string]interface{})
	if d == nil {
		return false
	}
	enabled, _ := d["enabled"].(bool)
	return enabled
}

// solrZookeeperOperatorInstalled reports whether the scenario keeps the
// bundled zookeeper-operator (the chart default when the block is absent
// or install is unset/true).
func solrZookeeperOperatorInstalled(spec map[string]interface{}) bool {
	zk, _ := spec["zookeeper_operator"].(map[string]interface{})
	if zk == nil {
		zk, _ = spec["zookeeperOperator"].(map[string]interface{})
	}
	if zk == nil {
		return true
	}
	if install, ok := zk["install"].(bool); ok {
		return install
	}
	return true
}

func solrSecurityEnabled(spec map[string]interface{}) bool {
	sec, _ := spec["security"].(map[string]interface{})
	if sec == nil {
		return false
	}
	authType, _ := sec["authentication_type"].(string)
	if authType == "" {
		authType, _ = sec["authenticationType"].(string)
	}
	return authType == "basic"
}

// neo4jAuthSecretName resolves the credentials Secret the verifier
// reads: the referenced Secret for the existing-secret arm, otherwise the
// module-materialized `<name>-auth` — which the module creates for every
// other arm, carrying the declared password or the one it generated when
// auth was left empty. Never empty: an absent auth block is the
// generated-credential proof, not a skip.
func neo4jAuthSecretName(spec map[string]interface{}, resourceName string) string {
	if auth, _ := spec["auth"].(map[string]interface{}); auth != nil {
		if existing, _ := auth["existing_secret"].(string); existing != "" {
			return existing
		}
		if existing, _ := auth["existingSecret"].(string); existing != "" {
			return existing
		}
	}
	return resourceName + "-auth"
}

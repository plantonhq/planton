package verify

import "gopkg.in/yaml.v3"

// Spec-map helpers for the observability kinds' verifier dispatch. Like
// every manifest helper here, each reader tolerates both the proto
// snake_case and the JSON camelCase key forms — scenario manifests are
// authored in either.

// kpsReplicas resolves a stack half's declared replica count (default 1
// — the proto default on both the prometheus and alertmanager blocks).
func kpsReplicas(spec map[string]interface{}, half string) int {
	block, _ := spec[half].(map[string]interface{})
	if block == nil {
		return 1
	}
	if replicas, ok := block["replicas"].(float64); ok && replicas >= 1 {
		return int(replicas)
	}
	if replicas, ok := block["replicas"].(int); ok && replicas >= 1 {
		return replicas
	}
	return 1
}

// kpsHalfEnabled reports whether the alertmanager / grafana half deploys:
// the proto optional-bool defaults TRUE, so only an explicit false
// disables it.
func kpsHalfEnabled(spec map[string]interface{}, half string) bool {
	block, _ := spec[half].(map[string]interface{})
	if block == nil {
		return true
	}
	if enabled, ok := block["enabled"].(bool); ok {
		return enabled
	}
	return true
}

// kpsOperatorEnabled reports whether the stack runs its own
// prometheus-operator. A receiver beside another stack turns it off through
// helm_values (prometheusOperator.enabled: false), because the other stack's
// operator already reconciles every Prometheus on the cluster.
func kpsOperatorEnabled(spec map[string]interface{}) bool {
	raw, _ := spec["helm_values"].(string)
	if raw == "" {
		raw, _ = spec["helmValues"].(string)
	}
	if raw == "" {
		return true
	}
	var values struct {
		PrometheusOperator struct {
			Enabled *bool `yaml:"enabled"`
		} `yaml:"prometheusOperator"`
	}
	if err := yaml.Unmarshal([]byte(raw), &values); err != nil || values.PrometheusOperator.Enabled == nil {
		return true
	}
	return *values.PrometheusOperator.Enabled
}

// kpsKubeStateMetricsEnabled reports whether the stack deploys
// kube-state-metrics: exporters.kube_state_metrics_enabled defaults true.
func kpsKubeStateMetricsEnabled(spec map[string]interface{}) bool {
	block, _ := spec["exporters"].(map[string]interface{})
	if block == nil {
		return true
	}
	for _, key := range []string{"kube_state_metrics_enabled", "kubeStateMetricsEnabled"} {
		if enabled, ok := block[key].(bool); ok {
			return enabled
		}
	}
	return true
}

// grafanaAdminSecretName resolves the Secret carrying the admin
// credentials: the referenced existing Secret when declared, else the
// chart-owned Secret named after the release (= the resource name, via
// the pinned fullname).
func grafanaAdminSecretName(spec map[string]interface{}, resourceName string) string {
	admin := grafanaAdminSecretMap(spec)
	if admin == nil {
		return resourceName
	}
	if name, _ := admin["name"].(string); name != "" {
		return name
	}
	return resourceName
}

// grafanaAdminSecretKey resolves one key-name override of the admin
// Secret ("" = the chart's admin-user / admin-password defaults, applied
// by the verifier).
func grafanaAdminSecretKey(spec map[string]interface{}, snakeKey string) string {
	admin := grafanaAdminSecretMap(spec)
	if admin == nil {
		return ""
	}
	if v, _ := admin[snakeKey].(string); v != "" {
		return v
	}
	camel := map[string]string{
		"user_key":     "userKey",
		"password_key": "passwordKey",
	}[snakeKey]
	if v, _ := admin[camel].(string); v != "" {
		return v
	}
	return ""
}

func grafanaAdminSecretMap(spec map[string]interface{}) map[string]interface{} {
	admin, _ := spec["admin_secret"].(map[string]interface{})
	if admin == nil {
		admin, _ = spec["adminSecret"].(map[string]interface{})
	}
	return admin
}

// grafanaDatasourceNames lists the declared datasource names the verifier
// must find provisioned.
// grafanaAgentReaderOf reads spec.agent_reader (snake or camel case) with
// the proto's defaults applied; nil when the block is not declared.
func grafanaAgentReaderOf(spec map[string]interface{}) *grafanaAgentReader {
	raw, declared := spec["agent_reader"]
	if !declared {
		raw, declared = spec["agentReader"]
	}
	if !declared {
		return nil
	}
	block, _ := raw.(map[string]interface{})
	field := func(snake, camel string) interface{} {
		if v, ok := block[snake]; ok {
			return v
		}
		return block[camel]
	}
	reader := &grafanaAgentReader{ServiceAccountName: "agent-reader", TokenGeneration: 1}
	if name, _ := field("service_account_name", "serviceAccountName").(string); name != "" {
		reader.ServiceAccountName = name
	}
	switch generation := field("token_generation", "tokenGeneration").(type) {
	case int:
		reader.TokenGeneration = generation
	case float64:
		reader.TokenGeneration = int(generation)
	}
	reader.Disabled, _ = field("disabled", "disabled").(bool)
	return reader
}

func grafanaDatasourceNames(spec map[string]interface{}) []string {
	raw, _ := spec["datasources"].([]interface{})
	names := make([]string, 0, len(raw))
	for _, entry := range raw {
		ds, _ := entry.(map[string]interface{})
		if name, _ := ds["name"].(string); name != "" {
			names = append(names, name)
		}
	}
	return names
}

// lokiGatewayEnabled reports whether the nginx gateway deploys: the proto
// optional-bool defaults TRUE, so only an explicit false disables it. The
// exported endpoints and the push→query proof route through the gateway.
func lokiGatewayEnabled(spec map[string]interface{}) bool {
	gw, _ := spec["gateway"].(map[string]interface{})
	if gw == nil {
		return true
	}
	if enabled, ok := gw["enabled"].(bool); ok {
		return enabled
	}
	return true
}

// lokiE2eTenantPassword is the plaintext half of the multi-tenant
// scenarios' credential pairing: tenant passwords in the spec are one-way
// bcrypt hashes (that is the product's security posture — no plaintext in
// manifests), so a scenario that declares a tenant commits to hashing
// EXACTLY this constant (`htpasswd -nbBC10 <tenant> e2e-password`) and the
// verifier authenticates with it. Change either side only with the other.
const lokiE2eTenantPassword = "e2e-password"

// lokiFirstTenantUser returns the first declared tenant name when
// multi-tenancy is enabled with an inline tenant list, else "". With
// tenants declared, the chart guards the WHOLE gateway server with
// auth_basic and injects X-Scope-OrgID from the authenticated username —
// there is no unauthenticated path — so the proof must run AS a tenant.
func lokiFirstTenantUser(spec map[string]interface{}) string {
	mt, _ := spec["multi_tenancy"].(map[string]interface{})
	if mt == nil {
		mt, _ = spec["multiTenancy"].(map[string]interface{})
	}
	if mt == nil {
		return ""
	}
	if enabled, ok := mt["enabled"].(bool); !ok || !enabled {
		return ""
	}
	tenants, _ := mt["tenants"].([]interface{})
	if len(tenants) == 0 {
		return ""
	}
	first, _ := tenants[0].(map[string]interface{})
	name, _ := first["name"].(string)
	return name
}

// kpsAlertOverride is one curated alert's declared hold (`for`, as a
// Prometheus duration) and severity; empty means the chart's own.
type kpsAlertOverride struct {
	For      string
	Severity string
}

// kpsRuleTuning reads the manifest's per-alert tuning of the curated rules
// (spec.default_rules.disabled_alerts and alert_overrides, snake_case or
// protojson camelCase keys).
func kpsRuleTuning(spec map[string]interface{}) ([]string, map[string]kpsAlertOverride) {
	block, _ := spec["default_rules"].(map[string]interface{})
	if block == nil {
		block, _ = spec["defaultRules"].(map[string]interface{})
	}
	if block == nil {
		return nil, nil
	}
	var disabled []string
	list, _ := block["disabled_alerts"].([]interface{})
	if list == nil {
		list, _ = block["disabledAlerts"].([]interface{})
	}
	for _, name := range list {
		if s, ok := name.(string); ok {
			disabled = append(disabled, s)
		}
	}
	overrides := map[string]kpsAlertOverride{}
	entries, _ := block["alert_overrides"].([]interface{})
	if entries == nil {
		entries, _ = block["alertOverrides"].([]interface{})
	}
	for _, entry := range entries {
		o, _ := entry.(map[string]interface{})
		alert, _ := o["alert"].(string)
		if alert == "" {
			continue
		}
		hold, _ := o["for"].(string)
		severity, _ := o["severity"].(string)
		overrides[alert] = kpsAlertOverride{For: hold, Severity: severity}
	}
	return disabled, overrides
}

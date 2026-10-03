package module

import (
	"fmt"
	"strings"

	kuberneteskubeprometheusstackv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kuberneteskubeprometheusstack/v1alpha1"
)

// Typed alert delivery (spec.alertmanager.notifications) renders into three
// pieces the chart already knows how to carry:
//   - the alertmanager.config map: the routing tree and the receivers. Helm
//     merges it over the chart's default config, so the chart's inhibit
//     rules, resolve timeout and template glob stay exactly the chart's;
//   - alertmanager.templateFiles: the one notification template, loaded by
//     that glob from /etc/alertmanager/config/;
//   - the module-owned credentials Secret, listed in
//     alertmanagerSpec.secrets so the operator mounts it at
//     /etc/alertmanager/secrets/<secret>/, and read by the receivers'
//     `_file` fields. No credential lands in chart values or in the
//     rendered configuration.
//
// PARITY: the Terraform module's locals.tf renders the same keys, paths,
// routes and template text byte for byte; the live notifications scenario
// asserts both engines deliver the identical message.

const (
	// notificationsTemplateFile is the templateFiles key; the chart's
	// default template glob (/etc/alertmanager/config/*.tmpl) loads it.
	notificationsTemplateFile = "planton-notifications.tmpl"
	// alertmanagerSecretsRoot is where the operator mounts each Secret
	// listed in alertmanagerSpec.secrets.
	alertmanagerSecretsRoot = "/etc/alertmanager/secrets"
	// heartbeatReceiver and discardReceiver are the module-owned receivers;
	// the spec refuses declared receivers with these names.
	heartbeatReceiver = "heartbeat"
	discardReceiver   = "discard"
	// heartbeatTokenKey is the heartbeat's bearer token inside the Secret.
	heartbeatTokenKey = "heartbeat-bearer-token"

	notificationTitle = `{{ template "planton.title" . }}`
	notificationText  = `{{ template "planton.text" . }}`
)

// notifications is the rendered form of spec.alertmanager.notifications.
type notifications struct {
	// Config is the alertmanager.config value.
	Config map[string]interface{}
	// TemplateFiles is the alertmanager.templateFiles value.
	TemplateFiles map[string]interface{}
	// SecretData is the module-owned Secret's content, keyed by the
	// deterministic keys below. Empty when no integration carries a
	// credential (then no Secret is created or mounted).
	SecretData map[string]string
}

// buildNotifications renders the typed notifications block, or returns nil
// when the manifest declares none (or Alertmanager is disabled).
func buildNotifications(locals *Locals) *notifications {
	n := locals.Spec.GetAlertmanager().GetNotifications()
	if !locals.AlertmanagerEnabled || n == nil {
		return nil
	}
	envLabel, componentLabel := notificationLabels(n)
	secretFile := func(key string) string {
		return alertmanagerSecretsRoot + "/" + locals.NotificationsSecretName + "/" + key
	}
	data := map[string]string{}

	receivers := []interface{}{map[string]interface{}{"name": discardReceiver}}

	heartbeat := n.GetHeartbeat()
	if heartbeat != nil {
		hook := map[string]interface{}{
			"url":           heartbeat.GetUrl(),
			"send_resolved": false,
			"max_alerts":    1,
		}
		if token := heartbeat.GetBearerToken().GetValue(); token != "" {
			data[heartbeatTokenKey] = token
			hook["http_config"] = bearerFile(secretFile(heartbeatTokenKey))
		}
		receivers = append(receivers, map[string]interface{}{
			"name":            heartbeatReceiver,
			"webhook_configs": []interface{}{hook},
		})
	}

	for _, r := range n.GetReceivers() {
		receiver := map[string]interface{}{"name": r.GetName()}
		if len(r.GetDiscord()) > 0 {
			configs := make([]interface{}, 0, len(r.GetDiscord()))
			for j, d := range r.GetDiscord() {
				key := fmt.Sprintf("%s-discord-%d-webhook-url", r.GetName(), j)
				data[key] = d.GetWebhookUrl().GetValue()
				configs = append(configs, map[string]interface{}{
					"webhook_url_file": secretFile(key),
					"title":            notificationTitle,
					"message":          notificationText,
				})
			}
			receiver["discord_configs"] = configs
		}
		if len(r.GetPushover()) > 0 {
			configs := make([]interface{}, 0, len(r.GetPushover()))
			for j, p := range r.GetPushover() {
				tokenKey := fmt.Sprintf("%s-pushover-%d-token", r.GetName(), j)
				userKey := fmt.Sprintf("%s-pushover-%d-user-key", r.GetName(), j)
				data[tokenKey] = p.GetToken().GetValue()
				data[userKey] = p.GetUserKey().GetValue()
				configs = append(configs, map[string]interface{}{
					"token_file":    secretFile(tokenKey),
					"user_key_file": secretFile(userKey),
					"title":         notificationTitle,
					"message":       notificationText,
					"url":           `{{ (index .Alerts 0).Annotations.runbook_url }}`,
					"url_title":     "Runbook",
					"priority":      pushoverPriority(p),
				})
			}
			receiver["pushover_configs"] = configs
		}
		if len(r.GetWebhook()) > 0 {
			configs := make([]interface{}, 0, len(r.GetWebhook()))
			for j, w := range r.GetWebhook() {
				hook := map[string]interface{}{"url": w.GetUrl()}
				if token := w.GetBearerToken().GetValue(); token != "" {
					key := fmt.Sprintf("%s-webhook-%d-bearer-token", r.GetName(), j)
					data[key] = token
					hook["http_config"] = bearerFile(secretFile(key))
				}
				configs = append(configs, hook)
			}
			receiver["webhook_configs"] = configs
		}
		receivers = append(receivers, receiver)
	}

	return &notifications{
		Config: map[string]interface{}{
			"route":     buildNotificationRoute(n, envLabel, componentLabel),
			"receivers": receivers,
		},
		TemplateFiles: map[string]interface{}{
			notificationsTemplateFile: notificationTemplate(envLabel, componentLabel),
		},
		SecretData: data,
	}
}

// buildNotificationRoute renders the routing tree: the module-owned Watchdog
// and InfoInhibitor routes first, so declared routes never see either, then
// the declared child routes in order.
func buildNotificationRoute(n *kuberneteskubeprometheusstackv1alpha1.KubernetesKubePrometheusStackAlertNotifications, envLabel, componentLabel string) map[string]interface{} {
	declared := n.GetRoute()
	groupBy := []interface{}{"alertname", envLabel, componentLabel}
	if len(declared.GetGroupBy()) > 0 {
		groupBy = stringsToInterfaces(declared.GetGroupBy())
	}
	route := map[string]interface{}{
		"receiver": declared.GetReceiver(),
		"group_by": groupBy,
	}
	if declared.GetGroupWait() != "" {
		route["group_wait"] = declared.GetGroupWait()
	}
	if declared.GetGroupInterval() != "" {
		route["group_interval"] = declared.GetGroupInterval()
	}
	if declared.GetRepeatInterval() != "" {
		route["repeat_interval"] = declared.GetRepeatInterval()
	}

	watchdog := map[string]interface{}{
		"receiver": discardReceiver,
		"matchers": []interface{}{`alertname="Watchdog"`},
	}
	if hb := n.GetHeartbeat(); hb != nil {
		// An always-firing alert is re-sent only every repeat_interval, so
		// the heartbeat route repeats at its own interval, well inside the
		// outside monitor's staleness window.
		interval := hb.GetInterval()
		if interval == "" {
			interval = vars.DefaultHeartbeatInterval
		}
		watchdog["receiver"] = heartbeatReceiver
		watchdog["group_wait"] = "0s"
		watchdog["group_interval"] = interval
		watchdog["repeat_interval"] = interval
	}
	routes := []interface{}{
		watchdog,
		map[string]interface{}{
			"receiver": discardReceiver,
			"matchers": []interface{}{`alertname="InfoInhibitor"`},
		},
	}
	for _, child := range declared.GetRoutes() {
		matchers := make([]interface{}, 0, len(child.GetMatchers()))
		for _, m := range child.GetMatchers() {
			matchers = append(matchers, renderMatcher(m))
		}
		entry := map[string]interface{}{
			"receiver": child.GetReceiver(),
			"matchers": matchers,
		}
		if child.GetContinueMatching() {
			entry["continue"] = true
		}
		if child.GetRepeatInterval() != "" {
			entry["repeat_interval"] = child.GetRepeatInterval()
		}
		routes = append(routes, entry)
	}
	route["routes"] = routes
	return route
}

// notificationLabels resolves the message's environment and component
// label names to their defaults.
func notificationLabels(n *kuberneteskubeprometheusstackv1alpha1.KubernetesKubePrometheusStackAlertNotifications) (string, string) {
	envLabel := n.GetMessage().GetEnvironmentLabel()
	if envLabel == "" {
		envLabel = vars.DefaultEnvironmentLabel
	}
	componentLabel := n.GetMessage().GetComponentLabel()
	if componentLabel == "" {
		componentLabel = vars.DefaultComponentLabel
	}
	return envLabel, componentLabel
}

// notificationTemplate is the one message shape every Discord and Pushover
// notification carries. The title leads with environment and component (an
// alert with no component is titled by its scrape job, and one with neither,
// such as an overcommit sum, by "cluster": it is about the whole cluster);
// the body is the first alert's customer_impact (else summary, else name), a
// count when the group holds more, and the runbook. It renders ONLY those
// labels and annotations: never description, namespace or pod, because on a
// shared cluster a namespace can name a customer and upstream rule
// descriptions interpolate it. The label names are validated to
// [a-zA-Z_][a-zA-Z0-9_]*, so interpolating them is safe. The OpenTofu
// module's notification_template is its byte-for-byte twin
// (template_twin_test.go).
func notificationTemplate(envLabel, componentLabel string) string {
	return `{{ define "planton.title" }}{{ if eq .Status "resolved" }}[RESOLVED] {{ end }}[{{ or (index .CommonLabels "` + envLabel + `") "unknown" }}] {{ or (index .CommonLabels "` + componentLabel + `") .CommonLabels.job "cluster" }}: {{ .CommonLabels.alertname }}{{ end }}
{{ define "planton.text" }}{{ with index .Alerts 0 }}{{ or .Annotations.customer_impact .Annotations.summary .Labels.alertname }}{{ end }}{{ if gt (len .Alerts) 1 }} ({{ len .Alerts }} alerts){{ end }}{{ with (index .Alerts 0).Annotations.runbook_url }}
Runbook: {{ . }}{{ end }}{{ end }}
`
}

// pushoverPriority renders the priority as a template: a firing alert
// pushes at the declared priority, the resolved follow-up at normal, so a
// recovery never pages.
func pushoverPriority(p *kuberneteskubeprometheusstackv1alpha1.KubernetesKubePrometheusStackAlertPushover) string {
	switch p.GetPriority() {
	case kuberneteskubeprometheusstackv1alpha1.KubernetesKubePrometheusStackPushoverPriority_normal:
		return "0"
	case kuberneteskubeprometheusstackv1alpha1.KubernetesKubePrometheusStackPushoverPriority_high:
		return `{{ if eq .Status "firing" }}1{{ else }}0{{ end }}`
	default:
		return `{{ if eq .Status "firing" }}2{{ else }}0{{ end }}`
	}
}

// renderMatcher renders one typed matcher in Alertmanager's matcher syntax,
// escaping the value for its double quotes.
func renderMatcher(m *kuberneteskubeprometheusstackv1alpha1.KubernetesKubePrometheusStackAlertMatcher) string {
	op := "="
	switch m.GetOperator() {
	case kuberneteskubeprometheusstackv1alpha1.KubernetesKubePrometheusStackAlertMatchOperator_not_equals:
		op = "!="
	case kuberneteskubeprometheusstackv1alpha1.KubernetesKubePrometheusStackAlertMatchOperator_matches_regex:
		op = "=~"
	case kuberneteskubeprometheusstackv1alpha1.KubernetesKubePrometheusStackAlertMatchOperator_not_matches_regex:
		op = "!~"
	}
	value := strings.ReplaceAll(strings.ReplaceAll(m.GetValue(), `\`, `\\`), `"`, `\"`)
	return m.GetLabel() + op + `"` + value + `"`
}

// bearerFile is the http_config that sends the file's content as a bearer
// token.
func bearerFile(path string) map[string]interface{} {
	return map[string]interface{}{
		"authorization": map[string]interface{}{
			"type":             "Bearer",
			"credentials_file": path,
		},
	}
}

func stringsToInterfaces(in []string) []interface{} {
	out := make([]interface{}, 0, len(in))
	for _, s := range in {
		out = append(out, s)
	}
	return out
}

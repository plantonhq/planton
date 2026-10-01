package module

import (
	"strings"
	"testing"

	kuberneteskubeprometheusstackv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kuberneteskubeprometheusstack/v1alpha1"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

// These tests pin what typed notifications render into the chart values:
// the credentials leave the configuration for one mounted Secret, the
// module-owned Watchdog and InfoInhibitor routes come before every declared
// route, a page-class route reaches the pager and renders Alertmanager's
// continue, and the message template never renders a label that can name a
// customer.
// The OpenTofu twin is held to the same shapes by the live notifications
// scenario, which asserts both engines deliver the identical message.

func literal(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v}}
}

// productionShapedLocals is a stack whose delivery mirrors a production
// cluster: channel by default, severity=page in environment prod to the
// pager (continuing to any route after it), and a heartbeat.
func productionShapedLocals(t *testing.T) *Locals {
	t.Helper()
	emergency := kuberneteskubeprometheusstackv1alpha1.KubernetesKubePrometheusStackPushoverPriority_emergency
	stackInput := &kuberneteskubeprometheusstackv1alpha1.KubernetesKubePrometheusStackStackInput{
		Target: &kuberneteskubeprometheusstackv1alpha1.KubernetesKubePrometheusStack{
			Metadata: &shared.CloudResourceMetadata{Name: "prod-metrics"},
			Spec: &kuberneteskubeprometheusstackv1alpha1.KubernetesKubePrometheusStackSpec{
				Namespace: literal("observability"),
				Alertmanager: &kuberneteskubeprometheusstackv1alpha1.KubernetesKubePrometheusStackAlertmanager{
					Notifications: &kuberneteskubeprometheusstackv1alpha1.KubernetesKubePrometheusStackAlertNotifications{
						Receivers: []*kuberneteskubeprometheusstackv1alpha1.KubernetesKubePrometheusStackAlertReceiver{
							{
								Name: "channel",
								Discord: []*kuberneteskubeprometheusstackv1alpha1.KubernetesKubePrometheusStackAlertDiscord{
									{WebhookUrl: literal("https://discord.example/webhook")},
								},
							},
							{
								Name: "pager",
								Pushover: []*kuberneteskubeprometheusstackv1alpha1.KubernetesKubePrometheusStackAlertPushover{
									{Token: literal("uQiRzpo4DXghDmr9QzzfQu27cmVRsG"), UserKey: literal("gznej3rKEVAvPUxu9vvNnqpmZpokzF"), Priority: &emergency},
								},
							},
						},
						Route: &kuberneteskubeprometheusstackv1alpha1.KubernetesKubePrometheusStackAlertRoute{
							Receiver: "channel",
							Routes: []*kuberneteskubeprometheusstackv1alpha1.KubernetesKubePrometheusStackAlertChildRoute{{
								Matchers: []*kuberneteskubeprometheusstackv1alpha1.KubernetesKubePrometheusStackAlertMatcher{
									{Label: "severity", Value: "page"},
									{Label: "environment", Value: `pr"od`},
								},
								Receiver:         "pager",
								ContinueMatching: true,
							}},
						},
						Heartbeat: &kuberneteskubeprometheusstackv1alpha1.KubernetesKubePrometheusStackAlertHeartbeat{
							Url:         "https://watcher.example/heartbeat/prod",
							BearerToken: literal("hb-6f1c0e2b9a7d4f58"),
						},
					},
				},
			},
		},
	}
	return initializeLocals(nil, stackInput)
}

func TestNotificationsKeepCredentialsInOneMountedSecret(t *testing.T) {
	locals := productionShapedLocals(t)
	n := locals.Notifications
	if n == nil {
		t.Fatal("declared notifications rendered nothing")
	}
	if locals.NotificationsSecretName != "prod-metrics-alertmanager-notifications" {
		t.Fatalf("secret name = %q", locals.NotificationsSecretName)
	}
	want := map[string]string{
		"channel-discord-0-webhook-url": "https://discord.example/webhook",
		"pager-pushover-0-token":        "uQiRzpo4DXghDmr9QzzfQu27cmVRsG",
		"pager-pushover-0-user-key":     "gznej3rKEVAvPUxu9vvNnqpmZpokzF",
		"heartbeat-bearer-token":        "hb-6f1c0e2b9a7d4f58",
	}
	if len(n.SecretData) != len(want) {
		t.Fatalf("secret data has %d keys, want %d: %v", len(n.SecretData), len(want), n.SecretData)
	}
	for k, v := range want {
		if n.SecretData[k] != v {
			t.Errorf("secret key %q = %q, want %q", k, n.SecretData[k], v)
		}
	}

	values, err := buildHelmValues(locals)
	if err != nil {
		t.Fatal(err)
	}
	alertmanager := values["alertmanager"].(map[string]interface{})
	secrets := alertmanager["alertmanagerSpec"].(map[string]interface{})["secrets"].([]interface{})
	if len(secrets) != 1 || secrets[0] != locals.NotificationsSecretName {
		t.Errorf("alertmanagerSpec.secrets = %v, want the one notifications Secret", secrets)
	}

	// Not one credential may appear anywhere in the rendered values.
	rendered := flatten(values)
	for _, v := range want {
		if strings.Contains(rendered, v) {
			t.Errorf("credential %q leaked into the chart values", v)
		}
	}
	if !strings.Contains(rendered, "/etc/alertmanager/secrets/prod-metrics-alertmanager-notifications/pager-pushover-0-token") {
		t.Error("the pushover token is not read from the mounted Secret")
	}
}

func TestNotificationsRouteModuleOwnedAlertsFirstAndPagesContinue(t *testing.T) {
	route := productionShapedLocals(t).Notifications.Config["route"].(map[string]interface{})
	if route["receiver"] != "channel" {
		t.Errorf("root receiver = %v", route["receiver"])
	}
	groupBy := route["group_by"].([]interface{})
	if len(groupBy) != 3 || groupBy[0] != "alertname" || groupBy[1] != "environment" || groupBy[2] != "component" {
		t.Errorf("default group_by = %v", groupBy)
	}
	routes := route["routes"].([]interface{})
	if len(routes) != 3 {
		t.Fatalf("want watchdog, info-inhibitor and the declared route, got %d routes", len(routes))
	}
	watchdog := routes[0].(map[string]interface{})
	if watchdog["receiver"] != heartbeatReceiver || watchdog["repeat_interval"] != "1m" || watchdog["group_wait"] != "0s" {
		t.Errorf("watchdog route = %v", watchdog)
	}
	if routes[1].(map[string]interface{})["receiver"] != discardReceiver {
		t.Errorf("info-inhibitor route = %v", routes[1])
	}
	page := routes[2].(map[string]interface{})
	matchers := page["matchers"].([]interface{})
	if page["receiver"] != "pager" || page["continue"] != true ||
		matchers[0] != `severity="page"` || matchers[1] != `environment="pr\"od"` {
		t.Errorf("page route = %v", page)
	}
}

func TestNotificationsWithoutHeartbeatDropTheWatchdog(t *testing.T) {
	locals := productionShapedLocals(t)
	locals.Spec.GetAlertmanager().GetNotifications().Heartbeat = nil
	n := buildNotifications(locals)
	watchdog := n.Config["route"].(map[string]interface{})["routes"].([]interface{})[0].(map[string]interface{})
	if watchdog["receiver"] != discardReceiver {
		t.Errorf("without a heartbeat the Watchdog must be dropped, got %v", watchdog)
	}
	if _, ok := n.SecretData[heartbeatTokenKey]; ok {
		t.Error("no heartbeat, yet its token was written")
	}
}

func TestNotificationTemplateRendersNoCustomerLabel(t *testing.T) {
	tmpl := notificationTemplate("environment", "component")
	for _, forbidden := range []string{"description", "namespace", ".pod", "SortedPairs", ".Labels.Values", "CommonAnnotations"} {
		if strings.Contains(tmpl, forbidden) {
			t.Errorf("the message template reads %q, which can carry a customer's name", forbidden)
		}
	}
	if !strings.Contains(tmpl, `index .CommonLabels "environment"`) || !strings.Contains(tmpl, `index .CommonLabels "component"`) {
		t.Error("the title must lead with the environment and component labels")
	}
}

func TestPushoverPriorityNeverPagesARecovery(t *testing.T) {
	for _, p := range []kuberneteskubeprometheusstackv1alpha1.KubernetesKubePrometheusStackPushoverPriority{
		kuberneteskubeprometheusstackv1alpha1.KubernetesKubePrometheusStackPushoverPriority_kubernetes_kube_prometheus_stack_pushover_priority_unspecified,
		kuberneteskubeprometheusstackv1alpha1.KubernetesKubePrometheusStackPushoverPriority_high,
		kuberneteskubeprometheusstackv1alpha1.KubernetesKubePrometheusStackPushoverPriority_emergency,
	} {
		priority := pushoverPriority(&kuberneteskubeprometheusstackv1alpha1.KubernetesKubePrometheusStackAlertPushover{Priority: &p})
		if !strings.Contains(priority, `{{ else }}0{{ end }}`) {
			t.Errorf("priority %v = %q: a resolved alert must push at normal priority", p, priority)
		}
	}
}

// flatten renders every string leaf of a values tree into one string for
// leak checks.
func flatten(v interface{}) string {
	var b strings.Builder
	var walk func(interface{})
	walk = func(v interface{}) {
		switch t := v.(type) {
		case map[string]interface{}:
			for _, child := range t {
				walk(child)
			}
		case []interface{}:
			for _, child := range t {
				walk(child)
			}
		case string:
			b.WriteString(t)
			b.WriteString("\n")
		}
	}
	walk(v)
	return b.String()
}

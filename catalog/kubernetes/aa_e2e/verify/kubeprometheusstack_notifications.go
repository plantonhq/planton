package verify

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// The behavioral-notifications scenario's alert-delivery proof. Typed
// notifications are only worth anything if a notification reaches a
// person, so this proof watches deliveries arrive at an in-cluster sink
// that stands in for Discord's webhook endpoint and an outside heartbeat
// monitor, rather than trusting the rendered configuration:
//   - the Watchdog heartbeat arrives with its bearer token (the
//     dead-man's-switch leg, read from the module-owned Secret);
//   - amtool accepts the configuration Alertmanager actually loaded;
//   - a synthetic page-class alert carrying customer-like labels reaches
//     the pager webhook with its token and the Discord channel through the
//     continuing route, titled with environment and component, with no
//     trace of the customer (the message template's allowlist);
//   - an alert with neither a component nor a job label (the shape of a
//     cluster-wide sum) is titled by its subject, the cluster;
//   - a rotated credential is picked up with no restart (Alertmanager reads
//     `_file` credentials per notification and the kubelet refreshes the
//     mounted Secret).
// Every expected value below is a literal of the scenario file
// (e2e/scenarios/behavioral-notifications.yaml); both engines must
// satisfy the same exact strings, which is the parity proof for the
// message template the two modules each carry.

const (
	notificationSinkName    = "notification-sink"
	notificationSinkPort    = 8080
	notifyHeartbeatPath     = "/heartbeat/e2e"
	notifyHeartbeatToken    = "e2e-heartbeat-token"
	notifyHeartbeatRotated  = "e2e-heartbeat-token-rotated"
	notifyPagerPath         = "/pager"
	notifyPagerToken        = "e2e-pager-token"
	notifyDiscordPath       = "/discord"
	notifyProofAlert        = "E2ENotificationProof"
	notifyExpectedTitle     = "[e2e] notification-proof: E2ENotificationProof"
	notifyExpectedSummary   = "Synthetic notification proof."
	notifyCustomerMarker    = "acme"
	notifyClusterWideAlert  = "E2EClusterWideProof"
	notifyClusterWideTitle  = "[e2e] cluster: E2EClusterWideProof"
	notifyAlertmanagerURL   = "http://localhost:9093"
	notifyLoadedConfigPath  = "/etc/alertmanager/config_out/alertmanager.env.yaml"
	notifyCredentialsSuffix = "-alertmanager-notifications"
)

// notificationSinkScript records every POST as one JSON line (path, the
// Authorization header, the body) and answers 200, which is all Discord's
// notifier and a webhook need to count a delivery.
const notificationSinkScript = `import http.server, json
class Sink(http.server.BaseHTTPRequestHandler):
    def do_POST(self):
        body = self.rfile.read(int(self.headers.get("Content-Length", 0))).decode()
        print(json.dumps({"path": self.path, "authorization": self.headers.get("Authorization", ""), "body": body}), flush=True)
        self.send_response(200)
        self.end_headers()
    def log_message(self, *args):
        pass
http.server.HTTPServer(("", 8080), Sink).serve_forever()
`

// sinkDelivery is one POST the sink recorded.
type sinkDelivery struct {
	Path          string `json:"path"`
	Authorization string `json:"authorization"`
	Body          string `json:"body"`
}

func (v *KubePrometheusStackVerifier) proveNotifications(ctx context.Context, kubeconfig string) error {
	if err := v.deployNotificationSink(ctx, kubeconfig); err != nil {
		return err
	}
	alertmanagerPod := "alertmanager-" + v.Name + "-alertmanager-0"

	if err := v.awaitDelivery(ctx, kubeconfig, 4*time.Minute, func(d sinkDelivery) bool {
		return d.Path == notifyHeartbeatPath && d.Authorization == "Bearer "+notifyHeartbeatToken
	}); err != nil {
		return errors.Wrap(err, "NOTIFICATIONS: the Watchdog heartbeat never arrived with its bearer token")
	}
	fmt.Printf("  [verify] NOTIFICATIONS: the Watchdog heartbeat arrived with its bearer token\n")

	if out, err := v.amtool(ctx, kubeconfig, alertmanagerPod, "check-config", notifyLoadedConfigPath); err != nil {
		return errors.Wrapf(err, "NOTIFICATIONS: amtool rejected the loaded configuration: %s", out)
	}
	fmt.Printf("  [verify] NOTIFICATIONS: amtool accepts the configuration Alertmanager loaded\n")

	if out, err := v.amtool(ctx, kubeconfig, alertmanagerPod, "alert", "add",
		"alertname="+notifyProofAlert, "environment=e2e", "component=notification-proof",
		"severity=page", "namespace=acme-corp", "pod=acme-corp-checkout-0",
		"--annotation=summary="+notifyExpectedSummary,
		"--annotation=description=acme-corp checkout is failing for every acme-corp customer",
		"--annotation=runbook_url=https://runbooks.example.com/e2e-notification-proof",
		"--alertmanager.url="+notifyAlertmanagerURL); err != nil {
		return errors.Wrapf(err, "NOTIFICATIONS: firing the synthetic alert failed: %s", out)
	}

	if err := v.awaitDelivery(ctx, kubeconfig, 3*time.Minute, func(d sinkDelivery) bool {
		return d.Path == notifyPagerPath && d.Authorization == "Bearer "+notifyPagerToken && strings.Contains(d.Body, notifyProofAlert)
	}); err != nil {
		return errors.Wrap(err, "NOTIFICATIONS: the page-class alert never reached the pager webhook with its token")
	}
	fmt.Printf("  [verify] NOTIFICATIONS: the page-class alert reached the pager with its token\n")

	var discord sinkDelivery
	if err := v.awaitDelivery(ctx, kubeconfig, 3*time.Minute, func(d sinkDelivery) bool {
		if d.Path == notifyDiscordPath && strings.Contains(d.Body, notifyProofAlert) {
			discord = d
			return true
		}
		return false
	}); err != nil {
		return errors.Wrap(err, "NOTIFICATIONS: the continuing route never delivered the alert to the Discord channel")
	}
	if !strings.Contains(discord.Body, notifyExpectedTitle) || !strings.Contains(discord.Body, notifyExpectedSummary) {
		return errors.Errorf("NOTIFICATIONS: the Discord message does not lead with environment and component and the summary: %s", discord.Body)
	}
	if strings.Contains(strings.ToLower(discord.Body), notifyCustomerMarker) {
		return errors.Errorf("NOTIFICATIONS: the Discord message carries a customer's name: %s", discord.Body)
	}
	fmt.Printf("  [verify] NOTIFICATIONS: the Discord message reads %q with no customer label: %s\n", notifyExpectedTitle, discord.Body)

	if out, err := v.amtool(ctx, kubeconfig, alertmanagerPod, "alert", "add",
		"alertname="+notifyClusterWideAlert, "environment=e2e", "severity=warning",
		"--annotation=summary=Synthetic cluster-wide proof.",
		"--alertmanager.url="+notifyAlertmanagerURL); err != nil {
		return errors.Wrapf(err, "NOTIFICATIONS: firing the cluster-wide alert failed: %s", out)
	}
	if err := v.awaitDelivery(ctx, kubeconfig, 3*time.Minute, func(d sinkDelivery) bool {
		return d.Path == notifyDiscordPath && strings.Contains(d.Body, notifyClusterWideTitle)
	}); err != nil {
		return errors.Wrapf(err, "NOTIFICATIONS: an alert with neither a component nor a job was not titled %q", notifyClusterWideTitle)
	}
	fmt.Printf("  [verify] NOTIFICATIONS: an alert with neither a component nor a job reads %q\n", notifyClusterWideTitle)

	// Rotation: rewrite the heartbeat token in the module-owned Secret and
	// expect the next heartbeats to present it with no pod restart.
	patch := fmt.Sprintf(`{"stringData":{"heartbeat-bearer-token":%q}}`, notifyHeartbeatRotated)
	if out, err := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig, "patch", "secret",
		v.Name+notifyCredentialsSuffix, "-n", v.Namespace, "--type", "merge", "-p", patch).CombinedOutput(); err != nil {
		return errors.Wrapf(err, "NOTIFICATIONS: patching the credentials Secret failed: %s", strings.TrimSpace(string(out)))
	}
	if err := v.awaitDelivery(ctx, kubeconfig, 5*time.Minute, func(d sinkDelivery) bool {
		return d.Path == notifyHeartbeatPath && d.Authorization == "Bearer "+notifyHeartbeatRotated
	}); err != nil {
		return errors.Wrap(err, "NOTIFICATIONS: a rotated credential was not picked up without a restart")
	}
	fmt.Printf("  [verify] NOTIFICATIONS: a rotated credential was picked up with no restart\n")
	return nil
}

// deployNotificationSink starts the sink Pod and its Service in the stack's
// namespace and waits for it to be Ready. Until it is, Alertmanager's
// deliveries fail and retry, which is the behaviour being proven anyway.
func (v *KubePrometheusStackVerifier) deployNotificationSink(ctx context.Context, kubeconfig string) error {
	// The script is a block scalar under a list item indented 8 spaces, so
	// its lines sit at 12.
	indented := "            " + strings.ReplaceAll(strings.TrimRight(notificationSinkScript, "\n"), "\n", "\n            ")
	manifest := fmt.Sprintf(`apiVersion: v1
kind: Pod
metadata:
  name: %[1]s
  namespace: %[2]s
  labels:
    app: %[1]s
spec:
  containers:
    - name: sink
      image: python:3.12-alpine
      command:
        - python
        - -u
        - -c
        - |
%[4]s
      ports:
        - containerPort: %[3]d
---
apiVersion: v1
kind: Service
metadata:
  name: %[1]s
  namespace: %[2]s
spec:
  selector:
    app: %[1]s
  ports:
    - port: %[3]d
      targetPort: %[3]d
`, notificationSinkName, v.Namespace, notificationSinkPort, indented)
	if err := kubectlApplyStdin(ctx, kubeconfig, manifest); err != nil {
		return errors.Wrap(err, "NOTIFICATIONS: deploying the notification sink failed")
	}
	if err := kubectlWait(ctx, kubeconfig, "pod", notificationSinkName, v.Namespace, "condition=Ready", 3*time.Minute); err != nil {
		return errors.Wrap(err, "NOTIFICATIONS: the notification sink never became ready")
	}
	return nil
}

// awaitDelivery polls the sink's log until a recorded delivery satisfies
// match or the budget runs out.
func (v *KubePrometheusStackVerifier) awaitDelivery(ctx context.Context, kubeconfig string, budget time.Duration, match func(sinkDelivery) bool) error {
	deadline := time.Now().Add(budget)
	seen := 0
	for time.Now().Before(deadline) {
		out, err := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig,
			"logs", notificationSinkName, "-n", v.Namespace).Output()
		if err == nil {
			seen = 0
			for _, line := range strings.Split(string(out), "\n") {
				var d sinkDelivery
				if json.Unmarshal([]byte(line), &d) != nil {
					continue
				}
				seen++
				if match(d) {
					return nil
				}
			}
		}
		time.Sleep(10 * time.Second)
	}
	return errors.Errorf("no matching delivery among the %d the sink recorded within %s", seen, budget)
}

// amtool runs amtool inside the Alertmanager container, which ships it at
// /bin/amtool.
func (v *KubePrometheusStackVerifier) amtool(ctx context.Context, kubeconfig, pod string, args ...string) (string, error) {
	cmdArgs := append([]string{"--kubeconfig", kubeconfig, "exec", pod, "-n", v.Namespace,
		"-c", "alertmanager", "--", "amtool"}, args...)
	out, err := exec.CommandContext(ctx, "kubectl", cmdArgs...).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

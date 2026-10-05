package verify

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// The feature-flag verifiers prove a flag engine to the point an application
// could ship behind it: the engine rolled out, a flag evaluates BOTH ways
// through OFREP (the OpenFeature remote evaluation protocol every OpenFeature
// SDK speaks) - the targeted organization gets the flag ON, an untargeted one
// gets it OFF - and THE LIVE FLIP: the flag file's ConfigMap is edited and the
// untargeted organization's answer turns ON within the engine's reload
// window, with zero pod restarts. A flag engine that cannot say no, or that
// needs a restart to change its mind, is not a release mechanism.
//
// The fixtures (each engine kind's consumer-scoped flag-file prerequisite)
// declare a boolean flag `assistant` targeting org "planton".

const (
	flagUnderTest      = "assistant"
	targetedOrg        = "planton"
	untargetedOrg      = "acme"
	flagSinkName       = "flag-change-sink"
	flagSinkPort       = 8080
	flagSinkToken      = "e2e-sink-token"
	goffFlipBudget     = 60 * time.Second
	flagdFlipBudget    = 4 * time.Minute
	flagEngineRollout  = 5 * time.Minute
	ofrepAnswerTimeout = 2 * time.Minute
)

// GoFeatureFlagVerifier proves a KubernetesGoFeatureFlag relay.
type GoFeatureFlagVerifier struct {
	Namespace string
	Name      string
	// Port is the evaluation port; MonitoringPort the health/metrics port.
	Port           string
	MonitoringPort string
	// Key is an evaluation key the relay must accept ("" = unauthenticated
	// posture); OtherKey, when set, is a second flag set's key that must
	// also resolve the flag.
	Key      string
	OtherKey string
	// FlagConfigMap / FlagNamespace / FlagKey locate the flag file the
	// first retriever reads.
	FlagConfigMap string
	FlagNamespace string
	FlagKey       string
	// ExpectNotification asks for a webhook delivery to the in-cluster sink
	// on the flip (the scenario wires a webhook notifier at it).
	ExpectNotification bool
	// PdbEnabled asks for the PodDisruptionBudget to select every replica.
	PdbEnabled bool
}

func (v *GoFeatureFlagVerifier) VerifyExists(ctx context.Context, kubeconfig string) error {
	fmt.Printf("  [verify] go-feature-flag relay %q in namespace %q (keyed: %v)\n", v.Name, v.Namespace, v.Key != "")

	if err := kubectlRolloutStatus(ctx, kubeconfig, "deployment/"+v.Name, v.Namespace, flagEngineRollout); err != nil {
		return errors.Wrap(err, "the relay deployment never rolled out")
	}
	ports, err := kubectlGetJSONPath(ctx, kubeconfig, "service", v.Name, v.Namespace, "{.spec.ports[*].name}")
	if err != nil {
		return errors.Wrap(err, "the relay service not found")
	}
	if !strings.Contains(ports, "monitoring") {
		return errors.Errorf("the relay service exposes %q - the `monitoring` port is missing, so the chart did not discover the monitoring port from the rendered configuration", ports)
	}
	fmt.Printf("  [verify] the service exposes the monitoring port\n")

	if v.PdbEnabled {
		if err := assertPdbCoversReplicas(ctx, kubeconfig, v.Name, v.Namespace); err != nil {
			return err
		}
	}

	if v.ExpectNotification {
		if err := deployFlagChangeSink(ctx, kubeconfig, v.Namespace); err != nil {
			return err
		}
	}

	closeTunnel, err := openServiceTunnel(ctx, kubeconfig, v.Namespace, v.Name, "18031", v.Port)
	if err != nil {
		return err
	}
	defer closeTunnel()
	base := "http://127.0.0.1:18031"

	if v.Key != "" {
		if _, code, err := ofrepEvaluate(ctx, base, "e2e-not-a-key", untargetedOrg); err != nil || code != http.StatusUnauthorized {
			return errors.Errorf("AUTH GATE: an evaluation with an unknown key answered %d (want 401): %v", code, err)
		}
		fmt.Printf("  [verify] AUTH GATE: an unknown key is refused with 401\n")
	}

	restarts, err := podRestartCount(ctx, kubeconfig, v.Namespace, v.Name)
	if err != nil {
		return err
	}
	if err := proveBothWaysAndFlip(ctx, kubeconfig, base, v.Key, v.FlagConfigMap, v.FlagNamespace, v.FlagKey, goffFlipBudget, flipGoFeatureFlagFile); err != nil {
		return err
	}
	if v.OtherKey != "" {
		value, code, err := ofrepEvaluate(ctx, base, v.OtherKey, targetedOrg)
		if err != nil || code != http.StatusOK || value != true {
			return errors.Errorf("FLAG SETS: the second flag set's key resolved %v (HTTP %d): %v", value, code, err)
		}
		fmt.Printf("  [verify] FLAG SETS: the second flag set's key resolves the flag too\n")
	}
	after, err := podRestartCount(ctx, kubeconfig, v.Namespace, v.Name)
	if err != nil {
		return err
	}
	if after != restarts {
		return errors.Errorf("THE LIVE FLIP restarted the relay (restarts %d -> %d); a flip must not restart anything", restarts, after)
	}

	if v.ExpectNotification {
		if err := awaitFlagChangeDelivery(ctx, kubeconfig, v.Namespace, v.Name); err != nil {
			return err
		}
	}
	return nil
}

func (v *GoFeatureFlagVerifier) VerifyAbsent(ctx context.Context, kubeconfig string) error {
	return KubectlResourceAbsent(ctx, kubeconfig, "deployment", v.Name, v.Namespace)
}

// FlagdVerifier proves a KubernetesFlagd daemon.
type FlagdVerifier struct {
	Namespace      string
	Name           string
	ManagementPort string
	OfrepPort      string
	FlagConfigMap  string
	FlagKey        string
	PdbEnabled     bool
}

func (v *FlagdVerifier) VerifyExists(ctx context.Context, kubeconfig string) error {
	fmt.Printf("  [verify] flagd %q in namespace %q\n", v.Name, v.Namespace)

	// flagd reports ready only once every source has synced, so a rolled
	// out Deployment already proves the mounted flag file was read.
	if err := kubectlRolloutStatus(ctx, kubeconfig, "deployment/"+v.Name, v.Namespace, flagEngineRollout); err != nil {
		return errors.Wrap(err, "the flagd deployment never rolled out (readiness waits on every source syncing)")
	}
	if v.PdbEnabled {
		if err := assertPdbCoversReplicas(ctx, kubeconfig, v.Name, v.Namespace); err != nil {
			return err
		}
	}

	closeMgmt, err := openServiceTunnel(ctx, kubeconfig, v.Namespace, v.Name, "18014", v.ManagementPort)
	if err != nil {
		return err
	}
	ready, err := httpStatus(ctx, "http://127.0.0.1:18014/readyz")
	closeMgmt()
	if err != nil || ready != http.StatusOK {
		return errors.Errorf("/readyz answered %d: %v", ready, err)
	}
	fmt.Printf("  [verify] /readyz answers 200 (every source synced)\n")

	closeTunnel, err := openServiceTunnel(ctx, kubeconfig, v.Namespace, v.Name, "18016", v.OfrepPort)
	if err != nil {
		return err
	}
	defer closeTunnel()

	restarts, err := podRestartCount(ctx, kubeconfig, v.Namespace, v.Name)
	if err != nil {
		return err
	}
	if err := proveBothWaysAndFlip(ctx, kubeconfig, "http://127.0.0.1:18016", "", v.FlagConfigMap, v.Namespace, v.FlagKey, flagdFlipBudget, flipFlagdFile); err != nil {
		return err
	}
	after, err := podRestartCount(ctx, kubeconfig, v.Namespace, v.Name)
	if err != nil {
		return err
	}
	if after != restarts {
		return errors.Errorf("THE LIVE FLIP restarted flagd (restarts %d -> %d); a flip must not restart anything", restarts, after)
	}
	return nil
}

func (v *FlagdVerifier) VerifyAbsent(ctx context.Context, kubeconfig string) error {
	return KubectlResourceAbsent(ctx, kubeconfig, "deployment", v.Name, v.Namespace)
}

// FlagFileVerifier proves a flag-file kind rendered its ConfigMap: the key
// holds a JSON document that declares the expected flag.
type FlagFileVerifier struct {
	Namespace string
	Name      string
	Key       string
	// FlagsPath is the path to the flag map inside the document ("" = the
	// document root, as GO Feature Flag files are; "flags" for flagd).
	FlagsPath string
	Flag      string
}

func (v *FlagFileVerifier) VerifyExists(ctx context.Context, kubeconfig string) error {
	fmt.Printf("  [verify] flag file %q (key %q) in namespace %q\n", v.Name, v.Key, v.Namespace)
	doc, err := readConfigMapKey(ctx, kubeconfig, v.Namespace, v.Name, v.Key)
	if err != nil {
		return err
	}
	parsed := map[string]interface{}{}
	if err := json.Unmarshal([]byte(doc), &parsed); err != nil {
		return errors.Wrapf(err, "the rendered flag file is not JSON: %s", firstLines(doc, 3))
	}
	flags := parsed
	if v.FlagsPath != "" {
		inner, ok := parsed[v.FlagsPath].(map[string]interface{})
		if !ok {
			return errors.Errorf("the rendered flag file has no %q object", v.FlagsPath)
		}
		flags = inner
	}
	if _, ok := flags[v.Flag]; !ok {
		return errors.Errorf("the rendered flag file does not declare flag %q", v.Flag)
	}
	fmt.Printf("  [verify] the ConfigMap declares flag %q\n", v.Flag)
	return nil
}

func (v *FlagFileVerifier) VerifyAbsent(ctx context.Context, kubeconfig string) error {
	return KubectlResourceAbsent(ctx, kubeconfig, "configmap", v.Name, v.Namespace)
}

// proveBothWaysAndFlip evaluates the flag for the targeted and the
// untargeted organization, flips the flag file so everyone is ON, waits for
// the untargeted answer to turn, then restores the original file.
func proveBothWaysAndFlip(ctx context.Context, kubeconfig, base, key, configMap, namespace, dataKey string, budget time.Duration,
	flip func(doc map[string]interface{}) error) error {
	if err := awaitOfrep(ctx, base, key, targetedOrg, true, ofrepAnswerTimeout); err != nil {
		return errors.Wrap(err, "BOTH WAYS: the targeted organization")
	}
	if err := awaitOfrep(ctx, base, key, untargetedOrg, false, ofrepAnswerTimeout); err != nil {
		return errors.Wrap(err, "BOTH WAYS: the untargeted organization")
	}
	fmt.Printf("  [verify] BOTH WAYS: %s gets %s ON, %s gets it OFF\n", targetedOrg, flagUnderTest, untargetedOrg)

	original, err := readConfigMapKey(ctx, kubeconfig, namespace, configMap, dataKey)
	if err != nil {
		return err
	}
	doc := map[string]interface{}{}
	if err := json.Unmarshal([]byte(original), &doc); err != nil {
		return errors.Wrap(err, "the flag file is not JSON")
	}
	if err := flip(doc); err != nil {
		return err
	}
	flipped, _ := json.Marshal(doc)
	if err := patchConfigMapKey(ctx, kubeconfig, namespace, configMap, dataKey, string(flipped)); err != nil {
		return err
	}
	defer func() {
		_ = patchConfigMapKey(context.Background(), kubeconfig, namespace, configMap, dataKey, original)
	}()

	started := time.Now()
	if err := awaitOfrep(ctx, base, key, untargetedOrg, true, budget); err != nil {
		return errors.Wrap(err, "THE LIVE FLIP: the edited flag file never reached evaluations")
	}
	fmt.Printf("  [verify] THE LIVE FLIP: %s turned ON %s after the edit\n", untargetedOrg, time.Since(started).Round(time.Second))
	return nil
}

// flipGoFeatureFlagFile turns the flag on for everyone (default rule ->
// the enabled variation).
func flipGoFeatureFlagFile(doc map[string]interface{}) error {
	flag, ok := doc[flagUnderTest].(map[string]interface{})
	if !ok {
		return errors.Errorf("the flag file does not declare %q", flagUnderTest)
	}
	flag["defaultRule"] = map[string]interface{}{"variation": "enabled"}
	return nil
}

// flipFlagdFile turns the flag on for everyone (default variant -> on, no
// targeting).
func flipFlagdFile(doc map[string]interface{}) error {
	flags, ok := doc["flags"].(map[string]interface{})
	if !ok {
		return errors.New("the flag file has no flags object")
	}
	flag, ok := flags[flagUnderTest].(map[string]interface{})
	if !ok {
		return errors.Errorf("the flag file does not declare %q", flagUnderTest)
	}
	flag["defaultVariant"] = "on"
	delete(flag, "targeting")
	return nil
}

// ofrepEvaluate evaluates the flag for an organization through OFREP and
// returns its value and the HTTP status.
func ofrepEvaluate(ctx context.Context, base, key, org string) (interface{}, int, error) {
	body := fmt.Sprintf(`{"context":{"targetingKey":%q,"org":%q}}`, org, org)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/ofrep/v1/evaluate/flags/"+flagUnderTest, strings.NewReader(body))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	if key != "" {
		req.Header.Set("Authorization", "Bearer "+key)
	}
	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, nil
	}
	answer := struct {
		Value interface{} `json:"value"`
	}{}
	if err := json.Unmarshal(raw, &answer); err != nil {
		return nil, resp.StatusCode, errors.Wrapf(err, "decoding the OFREP answer %s", string(raw))
	}
	return answer.Value, resp.StatusCode, nil
}

// awaitOfrep polls until the organization's answer equals want.
func awaitOfrep(ctx context.Context, base, key, org string, want bool, budget time.Duration) error {
	deadline := time.Now().Add(budget)
	var last string
	for time.Now().Before(deadline) {
		value, code, err := ofrepEvaluate(ctx, base, key, org)
		if err == nil && code == http.StatusOK && value == want {
			return nil
		}
		last = fmt.Sprintf("value=%v http=%d err=%v", value, code, err)
		time.Sleep(time.Second)
	}
	return errors.Errorf("%s never evaluated to %v within %s (last: %s)", org, want, budget, last)
}

func httpStatus(ctx context.Context, url string) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}
	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		return 0, err
	}
	resp.Body.Close()
	return resp.StatusCode, nil
}

func readConfigMapKey(ctx context.Context, kubeconfig, namespace, name, key string) (string, error) {
	out, err := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig, "get", "configmap", name, "-n", namespace, "-o", "json").Output()
	if err != nil {
		return "", errors.Wrapf(err, "reading configmap %s/%s", namespace, name)
	}
	cm := struct {
		Data map[string]string `json:"data"`
	}{}
	if err := json.Unmarshal(out, &cm); err != nil {
		return "", errors.Wrap(err, "decoding the configmap")
	}
	doc, ok := cm.Data[key]
	if !ok {
		return "", errors.Errorf("configmap %s/%s has no key %q", namespace, name, key)
	}
	return doc, nil
}

func patchConfigMapKey(ctx context.Context, kubeconfig, namespace, name, key, value string) error {
	patch, _ := json.Marshal(map[string]interface{}{"data": map[string]string{key: value}})
	if out, err := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig, "patch", "configmap", name, "-n", namespace,
		"--type", "merge", "-p", string(patch)).CombinedOutput(); err != nil {
		return errors.Wrapf(err, "patching configmap %s/%s: %s", namespace, name, string(out))
	}
	return nil
}

// podRestartCount sums container restarts across the workload's pods.
func podRestartCount(ctx context.Context, kubeconfig, namespace, instance string) (int, error) {
	out, err := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig, "get", "pods", "-n", namespace,
		"-l", "app.kubernetes.io/instance="+instance, "-o", "jsonpath={.items[*].status.containerStatuses[*].restartCount}").Output()
	if err != nil {
		return 0, errors.Wrap(err, "reading pod restart counts")
	}
	total := 0
	for _, f := range strings.Fields(string(out)) {
		var n int
		fmt.Sscanf(f, "%d", &n)
		total += n
	}
	return total, nil
}

// assertPdbCoversReplicas proves the PodDisruptionBudget selects every
// replica (expectedPods == the Deployment's ready replicas).
func assertPdbCoversReplicas(ctx context.Context, kubeconfig, name, namespace string) error {
	deadline := time.Now().Add(2 * time.Minute)
	for time.Now().Before(deadline) {
		expected, err1 := kubectlGetJSONPath(ctx, kubeconfig, "pdb", name, namespace, "{.status.expectedPods}")
		replicas, err2 := kubectlGetJSONPath(ctx, kubeconfig, "deployment", name, namespace, "{.status.readyReplicas}")
		if err1 == nil && err2 == nil && expected != "" && expected != "0" && expected == replicas {
			fmt.Printf("  [verify] the PodDisruptionBudget covers all %s ready replicas\n", replicas)
			return nil
		}
		time.Sleep(3 * time.Second)
	}
	return errors.Errorf("the PodDisruptionBudget %s/%s does not select the relay's pods (expectedPods never matched the ready replicas)", namespace, name)
}

// flagChangeSinkScript records every POST (path, the credential header, the
// signature header, the body) as one JSON line and answers 200. It reads
// chunked bodies too - the relay's webhook notifier sends its payload
// without a Content-Length.
const flagChangeSinkScript = `import http.server, json
class Sink(http.server.BaseHTTPRequestHandler):
    protocol_version = "HTTP/1.1"
    def read_body(self):
        if self.headers.get("Transfer-Encoding", "").lower() == "chunked":
            data = b""
            while True:
                size = int(self.rfile.readline().strip() or b"0", 16)
                if size == 0:
                    self.rfile.readline()
                    return data
                data += self.rfile.read(size)
                self.rfile.readline()
        return self.rfile.read(int(self.headers.get("Content-Length", 0)))
    def do_POST(self):
        body = self.read_body().decode()
        print(json.dumps({"path": self.path, "token": self.headers.get("X-Sink-Token", ""), "signature": self.headers.get("X-Hub-Signature-256", ""), "body": body}), flush=True)
        self.send_response(200)
        self.send_header("Content-Length", "0")
        self.end_headers()
    def log_message(self, *args):
        pass
http.server.HTTPServer(("", 8080), Sink).serve_forever()
`

func deployFlagChangeSink(ctx context.Context, kubeconfig, namespace string) error {
	indented := "            " + strings.ReplaceAll(strings.TrimRight(flagChangeSinkScript, "\n"), "\n", "\n            ")
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
`, flagSinkName, namespace, flagSinkPort, indented)
	if err := kubectlApplyStdin(ctx, kubeconfig, manifest); err != nil {
		return errors.Wrap(err, "NOTIFICATIONS: deploying the flag-change sink failed")
	}
	if err := kubectlWait(ctx, kubeconfig, "pod", flagSinkName, namespace, "condition=Ready", 3*time.Minute); err != nil {
		return errors.Wrap(err, "NOTIFICATIONS: the flag-change sink never became ready")
	}
	// The relay notifies on its initial load, before the sink exists, so the
	// cluster DNS has cached "no such host" for the sink's name. Wait out the
	// negative-cache window (CoreDNS caches denials for up to 30s) before
	// the flip, or the flip's notification resolves against the cached
	// denial and never leaves the relay.
	time.Sleep(35 * time.Second)
	return nil
}

// awaitFlagChangeDelivery polls the sink for a delivery carrying the
// credential header and an HMAC signature - proving the notifier's secret
// values reached the relay through the module-owned env Secret.
func awaitFlagChangeDelivery(ctx context.Context, kubeconfig, namespace, relay string) error {
	deadline := time.Now().Add(2 * time.Minute)
	var last string
	for time.Now().Before(deadline) {
		out, err := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig, "logs", flagSinkName, "-n", namespace).Output()
		if err == nil {
			last = string(out)
			for _, line := range strings.Split(last, "\n") {
				d := struct {
					Token     string `json:"token"`
					Signature string `json:"signature"`
					Body      string `json:"body"`
				}{}
				if json.Unmarshal([]byte(line), &d) == nil && d.Token == flagSinkToken && strings.HasPrefix(d.Signature, "sha256=") && strings.Contains(d.Body, flagUnderTest) {
					fmt.Printf("  [verify] NOTIFICATIONS: the flag change reached the webhook, signed, with the credential header\n")
					_ = exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig, "delete", "pod,service", flagSinkName, "-n", namespace, "--wait=false").Run()
					return nil
				}
			}
		}
		time.Sleep(3 * time.Second)
	}
	relayLog, _ := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig, "logs", "deployment/"+relay, "-n", namespace, "--tail=15").CombinedOutput()
	return errors.Errorf("NOTIFICATIONS: no signed delivery with the credential header reached the sink within 2m (sink log: %s; relay log: %s)", firstLines(last, 5), string(relayLog))
}

// newGoFeatureFlagVerifier reads what the relay verifier needs from the
// scenario manifest: ports, the evaluation keys (flag_source mode: the
// first evaluation key; flag_sets mode: each of the first two sets' first
// key), the first ConfigMap retriever, and the webhook notifier.
func newGoFeatureFlagVerifier(namespace, name string, spec map[string]interface{}) *GoFeatureFlagVerifier {
	v := &GoFeatureFlagVerifier{Namespace: namespace, Name: name, Port: "1031", MonitoringPort: "1032"}
	server := specNestedMap(spec, "server")
	if p := specNumberString(server, "port"); p != "" {
		v.Port = p
	}
	if p := specNumberString(server, "monitoringPort"); p != "" {
		v.MonitoringPort = p
	}
	v.PdbEnabled = specBool(specNestedMap(spec, "pdb"), "enabled")

	var sources []map[string]interface{}
	if fs := specNestedMap(spec, "flagSource"); fs != nil {
		sources = append(sources, fs)
		v.Key = firstString(specNestedMap(spec, "authorizedKeys"), "evaluation")
	}
	if sets := specNestedMap(spec, "flagSets"); sets != nil {
		items, _ := sets["items"].([]interface{})
		for i, item := range items {
			set, _ := item.(map[string]interface{})
			if set == nil {
				continue
			}
			source := specNestedMap(set, "source")
			sources = append(sources, source)
			switch i {
			case 0:
				v.Key = firstString(set, "apiKeys")
			case 1:
				v.OtherKey = firstString(set, "apiKeys")
			}
		}
	}
	for _, source := range sources {
		retrievers, _ := source["retrievers"].([]interface{})
		for _, r := range retrievers {
			cm := specNestedMap(asMap(r), "configMap")
			if cm != nil && v.FlagConfigMap == "" {
				v.FlagConfigMap = valueOrRefName(cm["configMapName"])
				v.FlagKey = valueOrRefKey(cm["key"], "flags.goff.yaml")
				v.FlagNamespace = valueOrRefName(cm["namespace"])
				if v.FlagNamespace == "" {
					v.FlagNamespace = namespace
				}
			}
		}
		notifiers, _ := source["notifiers"].([]interface{})
		for _, n := range notifiers {
			if specNestedMap(asMap(n), "webhook") != nil {
				v.ExpectNotification = true
			}
		}
	}
	return v
}

// newFlagdVerifier reads the ports and the first ConfigMap source.
func newFlagdVerifier(namespace, name string, spec map[string]interface{}) *FlagdVerifier {
	v := &FlagdVerifier{Namespace: namespace, Name: name, ManagementPort: "8014", OfrepPort: "8016"}
	server := specNestedMap(spec, "server")
	if p := specNumberString(server, "managementPort"); p != "" {
		v.ManagementPort = p
	}
	if p := specNumberString(server, "ofrepPort"); p != "" {
		v.OfrepPort = p
	}
	v.PdbEnabled = specBool(specNestedMap(spec, "pdb"), "enabled")
	sources, _ := spec["sources"].([]interface{})
	for _, s := range sources {
		if cm := specNestedMap(asMap(s), "configMap"); cm != nil {
			v.FlagConfigMap = valueOrRefName(cm["configMapName"])
			v.FlagKey = valueOrRefKey(cm["key"], "flags.flagd.json")
			break
		}
	}
	return v
}

// newFlagFileVerifier reads the rendered ConfigMap's key (the spec default
// when unset).
func newFlagFileVerifier(namespace, name string, spec map[string]interface{}, defaultKey, flagsPath string) *FlagFileVerifier {
	key, _ := spec["key"].(string)
	if key == "" {
		key = defaultKey
	}
	return &FlagFileVerifier{Namespace: namespace, Name: name, Key: key, FlagsPath: flagsPath, Flag: flagUnderTest}
}

func asMap(v interface{}) map[string]interface{} {
	m, _ := v.(map[string]interface{})
	return m
}

func specBool(m map[string]interface{}, key string) bool {
	b, _ := m[key].(bool)
	return b
}

func specNumberString(m map[string]interface{}, key string) string {
	switch n := m[key].(type) {
	case float64:
		return fmt.Sprintf("%d", int(n))
	case int:
		return fmt.Sprintf("%d", n)
	case int64:
		return fmt.Sprintf("%d", n)
	case string:
		return n
	}
	return ""
}

func firstString(m map[string]interface{}, key string) string {
	list, _ := m[key].([]interface{})
	if len(list) == 0 {
		return ""
	}
	s, _ := list[0].(string)
	return s
}

// valueOrRefName returns a value-or-ref's literal, or the referenced
// resource's name (the flag-file kinds render their ConfigMap under their
// own name).
func valueOrRefName(v interface{}) string {
	m := asMap(v)
	if m == nil {
		s, _ := v.(string)
		return s
	}
	if s, ok := m["value"].(string); ok {
		return s
	}
	if ref := asMap(m["valueFrom"]); ref != nil {
		s, _ := ref["name"].(string)
		return s
	}
	return ""
}

// valueOrRefKey returns a key's literal, or the referenced flag file's key:
// the scenarios' flag-file prerequisites leave the key at its spec default.
func valueOrRefKey(v interface{}, flagFileDefault string) string {
	m := asMap(v)
	if m == nil {
		s, _ := v.(string)
		return s
	}
	if s, ok := m["value"].(string); ok {
		return s
	}
	if asMap(m["valueFrom"]) != nil {
		return flagFileDefault
	}
	return ""
}

package verify

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// The agent-sandbox controller's fixed names and pinned release. The module
// installs upstream's own manifests, so these are upstream's names; the
// release mirrors the module's pin (the CRD stamps carry it).
const (
	agentSandboxKind              = "KubernetesAgentSandboxController"
	agentSandboxRelease           = "v1.0.7"
	agentSandboxBundle            = "agent-sandbox"
	agentSandboxNamespace         = "agent-sandbox-system"
	agentSandboxControllerName    = "agent-sandbox-controller"
	agentSandboxRouterName        = "sandbox-router"
	agentSandboxRouterServiceName = "sandbox-router-svc"
	agentSandboxOwnerLabel        = "planton.ai/kind"

	// The probe the verifier creates to prove the controller serves: a
	// namespace of its own, one Sandbox, and with extensions a template, a
	// one-member warm pool and a claim. agnhost's netexec answers /hostname
	// on the router's default sandbox port, so a request through the router
	// proves the route end to end.
	agentSandboxProbeNamespace = "e2e-asbx-probe"
	agentSandboxProbeSandbox   = "e2e-asbx-sandbox"
	agentSandboxProbeClaim     = "e2e-asbx-claim"
	agentSandboxProbeClient    = "e2e-asbx-client"
	agentSandboxProbePort      = 8888
	agentSandboxProbeImage     = "registry.k8s.io/e2e-test-images/agnhost:2.53"
	agentSandboxClientImage    = "curlimages/curl:8.10.1"
)

var (
	agentSandboxCoreCrds      = []string{"sandboxes.agents.x-k8s.io"}
	agentSandboxExtensionCrds = []string{
		"sandboxtemplates.extensions.agents.x-k8s.io",
		"sandboxwarmpools.extensions.agents.x-k8s.io",
		"sandboxclaims.extensions.agents.x-k8s.io",
	}
)

// The refusals this kind adds to the CRD primitive's: a controller or router
// another installation created, and the namespace's owner (missing while
// another resource owns it, or created by someone else while this resource
// would create it). Each is asserted in its three parts.
const (
	failureWorkloadOwnedElsewhere  = "workload-owned-elsewhere"
	failureNamespaceMissing        = "namespace-missing"
	failureNamespaceOwnedElsewhere = "namespace-owned-elsewhere"
)

// AgentSandboxControllerVerifier checks an agent-sandbox install by what it
// promises a user, not only by what runs: a Sandbox becomes Ready; with
// extensions, a claim binds a warm pool member (carrying a pod label from the
// declared label domains); with a router, a request reaches the sandbox by
// name, and with TokenReview a tokenless request is refused while a
// ServiceAccount token passes. Destroy asserts the kept-definition and
// namespace-ownership promises: with the definitions kept, the probe Sandbox
// and its pod outlive the controller.
type AgentSandboxControllerVerifier struct {
	Extensions bool
	// LabelDomain is the domain the probe claim's pod label uses: the first
	// declared allowed domain, or upstream's default.
	LabelDomain     string
	CreateNamespace bool
	Router          *agentSandboxRouter
	helmCRDLifecycle
}

type agentSandboxRouter struct {
	URL          string
	TokenReview  bool
	RequireToken bool
	// ClientNamespaceLabels go on the probe namespace so the router's
	// NetworkPolicy admits the probe client on a cluster that enforces it.
	ClientNamespaceLabels map[string]string
}

func newAgentSandboxControllerVerifier(manifestPath string) *AgentSandboxControllerVerifier {
	spec := manifestSpecMap(manifestPath)
	v := &AgentSandboxControllerVerifier{LabelDomain: "sandbox.users.io"}
	v.CreateNamespace, _ = specField(spec, "createNamespace").(bool)

	if ext, ok := specField(spec, "extensions").(map[string]interface{}); ok {
		v.Extensions = boolOr(specField(ext, "enabled"), true)
		if domains, ok := specField(ext, "allowedLabelDomains").([]interface{}); ok && len(domains) > 0 {
			if first, ok := domains[0].(string); ok {
				v.LabelDomain = first
			}
		}
	}

	crds := append([]string{}, agentSandboxCoreCrds...)
	if v.Extensions {
		crds = append(crds, agentSandboxExtensionCrds...)
	}
	// A refusal lane names the definitions it planted; the refusal must name
	// those, not the whole set.
	if expected := manifestAnnotation(manifestPath, "planton.dev/e2e-expect-crds"); expected != "" {
		crds = strings.Split(expected, ",")
	}
	v.helmCRDLifecycle = readHelmCRDLifecycle(manifestPath, agentSandboxRelease, agentSandboxBundle, crds)

	if router, ok := specField(spec, "router").(map[string]interface{}); ok && boolOr(specField(router, "enabled"), true) {
		v.Router = readAgentSandboxRouter(router, stringOr(specField(spec, "clusterDomain"), "cluster.local"))
	}
	return v
}

func readAgentSandboxRouter(router map[string]interface{}, clusterDomain string) *agentSandboxRouter {
	r := &agentSandboxRouter{ClientNamespaceLabels: map[string]string{}}
	if authz, ok := specField(router, "authorization").(map[string]interface{}); ok {
		if tr, ok := specField(authz, "tokenReview").(map[string]interface{}); ok {
			r.TokenReview = true
			r.RequireToken, _ = specField(tr, "requireToken").(bool)
		}
	}
	if np, ok := specField(router, "networkPolicy").(map[string]interface{}); ok {
		if labels, ok := specField(np, "clientNamespaceLabels").(map[string]interface{}); ok {
			for key, value := range labels {
				r.ClientNamespaceLabels[key], _ = value.(string)
			}
		}
	}
	host := fmt.Sprintf("%s.%s.svc.%s", agentSandboxRouterServiceName, agentSandboxNamespace, clusterDomain)
	httpPort, httpEnabled := int64(8080), true
	if ports, ok := specField(router, "ports").(map[string]interface{}); ok {
		httpPort = intOr(specField(ports, "http"), httpPort)
	}
	if tls, ok := specField(router, "tls").(map[string]interface{}); ok && boolOr(specField(tls, "enabled"), true) {
		httpEnabled = boolOr(specField(tls, "httpEnabled"), true)
		if !httpEnabled {
			r.URL = fmt.Sprintf("https://%s:%d", host, intOr(specField(tls, "port"), 8443))
			return r
		}
	}
	r.URL = fmt.Sprintf("http://%s:%d", host, httpPort)
	return r
}

func (v *AgentSandboxControllerVerifier) VerifyExists(ctx context.Context, kubeconfig string) error {
	fmt.Printf("  [verify] agent-sandbox controller in %q (extensions=%t, router=%t)\n", agentSandboxNamespace, v.Extensions, v.Router != nil)

	if err := kubectlWait(ctx, kubeconfig, "deployment", agentSandboxControllerName, agentSandboxNamespace,
		"condition=Available", 5*time.Minute); err != nil {
		return errors.Wrap(err, "the controller Deployment never became Available")
	}
	args, err := kubectlGetJSONPath(ctx, kubeconfig, "deployment", agentSandboxControllerName, agentSandboxNamespace,
		"{.spec.template.spec.containers[0].args}")
	if err != nil {
		return errors.Wrap(err, "reading the controller's arguments")
	}
	if strings.Contains(args, "--extensions") != v.Extensions {
		return errors.Errorf("the controller runs with arguments %s, but the manifest sets extensions=%t -- the wrong release asset was installed", args, v.Extensions)
	}
	if err := v.verifyEstablishedAndStamped(ctx, kubeconfig); err != nil {
		return err
	}
	if !v.Extensions && !v.KeepOnUninstall && v.InstallCrds {
		for _, crd := range agentSandboxExtensionCrds {
			if err := KubectlResourceAbsent(ctx, kubeconfig, "crd", crd, ""); err != nil {
				return errors.Wrapf(err, "extensions are off and the definitions are not kept, so %s must be gone", crd)
			}
		}
	}
	if err := v.verifyNamespaceOwner(ctx, kubeconfig); err != nil {
		return err
	}

	if err := v.applyProbe(ctx, kubeconfig); err != nil {
		return err
	}
	if err := kubectlWait(ctx, kubeconfig, "sandboxes.agents.x-k8s.io", agentSandboxProbeSandbox, agentSandboxProbeNamespace,
		"condition=Ready", 4*time.Minute); err != nil {
		return errors.Wrap(err, "the probe Sandbox never became Ready -- the controller runs but does not serve")
	}
	fmt.Printf("  [verify] a Sandbox became Ready: the controller serves\n")

	if v.Extensions {
		if err := kubectlWait(ctx, kubeconfig, "sandboxclaims.extensions.agents.x-k8s.io", agentSandboxProbeClaim, agentSandboxProbeNamespace,
			"condition=Ready", 5*time.Minute); err != nil {
			return errors.Wrap(err, "the probe SandboxClaim never became Ready -- the extension controllers do not bind claims")
		}
		fmt.Printf("  [verify] a SandboxClaim bound a warm pool member, with a %s pod label\n", v.LabelDomain)
	}

	if v.Router != nil {
		return v.verifyRouter(ctx, kubeconfig)
	}
	return nil
}

func (v *AgentSandboxControllerVerifier) verifyNamespaceOwner(ctx context.Context, kubeconfig string) error {
	owner, err := kubectlGetJSONPath(ctx, kubeconfig, "namespace", agentSandboxNamespace, "",
		"{.metadata.labels.planton\\.ai/kind}")
	if err != nil {
		return errors.Wrapf(err, "reading namespace %s", agentSandboxNamespace)
	}
	owner = strings.TrimSpace(owner)
	if v.CreateNamespace && owner != agentSandboxKind {
		return errors.Errorf("namespace %s carries %s=%q, but this resource created it", agentSandboxNamespace, agentSandboxOwnerLabel, owner)
	}
	if !v.CreateNamespace && owner == agentSandboxKind {
		return errors.Errorf("namespace %s carries this kind's mark, but another resource owns it", agentSandboxNamespace)
	}
	return nil
}

// applyProbe creates the probe objects; applying them again on a reinstall is
// a no-op.
func (v *AgentSandboxControllerVerifier) applyProbe(ctx context.Context, kubeconfig string) error {
	labels := ""
	if v.Router != nil {
		for key, value := range v.Router.ClientNamespaceLabels {
			labels += fmt.Sprintf("    %q: %q\n", key, value)
		}
	}
	if labels != "" {
		labels = "  labels:\n" + labels
	}
	container := fmt.Sprintf(`        - name: app
          image: %s
          args: ["netexec", "--http-port=%d", "--udp-port=-1"]
          ports:
            - containerPort: %d
          readinessProbe:
            httpGet:
              path: /hostname
              port: %d`, agentSandboxProbeImage, agentSandboxProbePort, agentSandboxProbePort, agentSandboxProbePort)
	manifest := fmt.Sprintf(`apiVersion: v1
kind: Namespace
metadata:
  name: %[1]s
%[2]s---
apiVersion: agents.x-k8s.io/v1beta1
kind: Sandbox
metadata:
  name: %[3]s
  namespace: %[1]s
spec:
  podTemplate:
    spec:
      containers:
%[4]s
`, agentSandboxProbeNamespace, labels, agentSandboxProbeSandbox, container)
	if v.Extensions {
		manifest += fmt.Sprintf(`---
apiVersion: extensions.agents.x-k8s.io/v1beta1
kind: SandboxTemplate
metadata:
  name: e2e-asbx-template
  namespace: %[1]s
spec:
  podTemplate:
    spec:
      containers:
%[2]s
---
apiVersion: extensions.agents.x-k8s.io/v1beta1
kind: SandboxWarmPool
metadata:
  name: e2e-asbx-pool
  namespace: %[1]s
spec:
  replicas: 1
  sandboxTemplateRef:
    name: e2e-asbx-template
---
apiVersion: extensions.agents.x-k8s.io/v1beta1
kind: SandboxClaim
metadata:
  name: %[3]s
  namespace: %[1]s
spec:
  warmPoolRef:
    name: e2e-asbx-pool
  additionalPodMetadata:
    labels:
      %[4]s/e2e: probe
`, agentSandboxProbeNamespace, container, agentSandboxProbeClaim, v.LabelDomain)
	}
	if v.Router != nil {
		manifest += fmt.Sprintf(`---
apiVersion: v1
kind: Pod
metadata:
  name: %s
  namespace: %s
spec:
  containers:
    - name: curl
      image: %s
      command: ["sleep", "3600"]
`, agentSandboxProbeClient, agentSandboxProbeNamespace, agentSandboxClientImage)
	}
	if err := kubectlApplyStdin(ctx, kubeconfig, manifest); err != nil {
		return errors.Wrap(err, "creating the probe sandbox")
	}
	return nil
}

// verifyRouter sends requests through the router to the probe Sandbox by
// name, as upstream's SDKs do.
func (v *AgentSandboxControllerVerifier) verifyRouter(ctx context.Context, kubeconfig string) error {
	if err := kubectlWait(ctx, kubeconfig, "deployment", agentSandboxRouterName, agentSandboxNamespace,
		"condition=Available", 4*time.Minute); err != nil {
		return errors.Wrap(err, "the router Deployment never became Available")
	}
	if err := kubectlWait(ctx, kubeconfig, "pod", agentSandboxProbeClient, agentSandboxProbeNamespace,
		"condition=Ready", 3*time.Minute); err != nil {
		return errors.Wrap(err, "the probe client never became Ready")
	}
	if v.Router.TokenReview && v.Router.RequireToken {
		if err := v.pollRouter(ctx, kubeconfig, "", "401", 2*time.Minute,
			"a request without a token was not refused by TokenReview authorization"); err != nil {
			return err
		}
		out, err := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig,
			"create", "token", "default", "-n", agentSandboxProbeNamespace, "--duration=10m").CombinedOutput()
		if err != nil {
			return errors.Wrapf(err, "minting a ServiceAccount token for the probe: %s", firstLines(string(out), 2))
		}
		return v.pollRouter(ctx, kubeconfig, strings.TrimSpace(string(out)), "200", 2*time.Minute,
			"a request with a valid ServiceAccount token did not reach the sandbox")
	}
	return v.pollRouter(ctx, kubeconfig, "", "200", 3*time.Minute,
		"a request through the router never reached the sandbox")
}

func (v *AgentSandboxControllerVerifier) pollRouter(ctx context.Context, kubeconfig, token, want string, timeout time.Duration, failMsg string) error {
	args := []string{"--kubeconfig", kubeconfig, "exec", agentSandboxProbeClient, "-n", agentSandboxProbeNamespace, "--",
		"curl", "-sk", "-o", "/dev/null", "-w", "%{http_code}", "--max-time", "5",
		"-H", "X-Sandbox-ID: " + agentSandboxProbeSandbox,
		"-H", "X-Sandbox-Namespace: " + agentSandboxProbeNamespace,
		"-H", "X-Sandbox-Port: " + strconv.Itoa(agentSandboxProbePort)}
	if token != "" {
		args = append(args, "-H", "Authorization: Bearer "+token)
	}
	args = append(args, v.Router.URL+"/hostname")
	deadline := time.Now().Add(timeout)
	var last string
	for time.Now().Before(deadline) {
		out, err := exec.CommandContext(ctx, "kubectl", args...).CombinedOutput()
		status := strings.TrimSpace(string(out))
		if err == nil && status == want {
			fmt.Printf("  [verify] a request through the router to the sandbox by name returned HTTP %s, as required\n", status)
			return nil
		}
		last = fmt.Sprintf("status=%q err=%v", status, err)
		time.Sleep(5 * time.Second)
	}
	return errors.Errorf("%s (%s, last probe: %s)", failMsg, v.Router.URL, last)
}

func (v *AgentSandboxControllerVerifier) VerifyAbsent(ctx context.Context, kubeconfig string) error {
	for _, name := range []string{agentSandboxControllerName, agentSandboxRouterName} {
		if err := KubectlResourceAbsent(ctx, kubeconfig, "deployment", name, agentSandboxNamespace); err != nil {
			return err
		}
	}
	if !v.DeployRefused {
		if v.CreateNamespace {
			if err := KubectlResourceAbsent(ctx, kubeconfig, "namespace", agentSandboxNamespace, ""); err != nil {
				return errors.Wrap(err, "this resource created the namespace, so destroy must delete it")
			}
		} else if err := KubectlResourceExists(ctx, kubeconfig, "namespace", agentSandboxNamespace, ""); err != nil {
			return errors.Wrap(err, "another resource owns the namespace, so destroying the controller must leave it")
		}
	}
	if err := v.verifyDestroyPosture(ctx, kubeconfig); err != nil {
		return err
	}
	// The product promise behind keeping the definitions: removing the
	// controller never deletes a sandbox or its pod.
	if !v.DeployRefused && v.InstallCrds && v.KeepOnUninstall {
		if err := KubectlResourceExists(ctx, kubeconfig, "sandboxes.agents.x-k8s.io", agentSandboxProbeSandbox, agentSandboxProbeNamespace); err != nil {
			return errors.Wrap(err, "the probe Sandbox was deleted with the controller")
		}
		if err := KubectlResourceExists(ctx, kubeconfig, "pod", agentSandboxProbeSandbox, agentSandboxProbeNamespace); err != nil {
			return errors.Wrap(err, "the probe Sandbox's pod was deleted with the controller")
		}
		fmt.Printf("  [verify] DESTROY: the Sandbox and its pod outlived the controller\n")
	}
	// The probe is the verifier's own; nothing else removes it.
	if out, err := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig, "delete", "namespace",
		agentSandboxProbeNamespace, "--ignore-not-found", "--wait=true", "--timeout=3m").CombinedOutput(); err != nil {
		return errors.Wrapf(err, "removing the probe namespace: %s", firstLines(string(out), 3))
	}
	return nil
}

// VerifyExpectedDeployFailure pins a refused deploy to its three-part text:
// this kind's own refusals here, the CRD primitive's in helmCRDLifecycle.
func (v *AgentSandboxControllerVerifier) VerifyExpectedDeployFailure(ctx context.Context, kubeconfig, expectation string, deployErr error) error {
	text := strings.Join(strings.Fields(strings.ReplaceAll(deployErr.Error(), "|", " ")), " ")
	var wants []string
	switch expectation {
	case failureWorkloadOwnedElsewhere:
		wants = []string{
			"Deployment " + agentSandboxNamespace + "/" + agentSandboxControllerName + " already runs on the cluster and was not created by this " + agentSandboxKind,
			"another installation of the same software runs here",
			"label deployment " + agentSandboxControllerName,
		}
	case failureNamespaceMissing:
		wants = []string{
			"namespace " + agentSandboxNamespace + " does not exist on the cluster, and this " + agentSandboxKind + " does not create it",
			"its owner must create it first",
			"set spec.create_namespace to true",
		}
	case failureNamespaceOwnedElsewhere:
		wants = []string{
			"namespace " + agentSandboxNamespace + " already exists on the cluster and was not created by this " + agentSandboxKind,
			"would take the namespace over",
			"set spec.create_namespace to false",
		}
	default:
		return v.verifyRefusal(ctx, kubeconfig, expectation, deployErr)
	}
	for _, part := range []string{"observed:", "meaning:", "next step:"} {
		if !strings.Contains(text, part) {
			return errors.Errorf("the refusal lacks its %q part; got: %s", part, firstLines(text, 12))
		}
	}
	for _, want := range wants {
		if !strings.Contains(text, want) {
			return errors.Errorf("expected the %s refusal saying %q; got: %s", expectation, want, firstLines(text, 12))
		}
	}
	// Nothing was taken over: the other installation's controller still
	// lacks this kind's mark.
	if expectation == failureWorkloadOwnedElsewhere {
		owner, err := kubectlGetJSONPath(ctx, kubeconfig, "deployment", agentSandboxControllerName, agentSandboxNamespace,
			"{.metadata.labels.planton\\.ai/kind}")
		if err != nil {
			return errors.Wrap(err, "reading the other installation's controller after the refusal")
		}
		if strings.TrimSpace(owner) == agentSandboxKind {
			return errors.New("the other installation's controller gained this kind's mark despite the refusal")
		}
	}
	fmt.Printf("  [verify] REFUSED as expected (%s), with what was observed, what it means and the next step\n", expectation)
	return nil
}

func boolOr(value interface{}, fallback bool) bool {
	if b, ok := value.(bool); ok {
		return b
	}
	return fallback
}

func stringOr(value interface{}, fallback string) string {
	if s, ok := value.(string); ok && s != "" {
		return s
	}
	return fallback
}

func intOr(value interface{}, fallback int64) int64 {
	switch n := value.(type) {
	case int:
		return int64(n)
	case int64:
		return n
	case float64:
		return int64(n)
	}
	return fallback
}

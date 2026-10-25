package verify

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/catalog/kubernetes/aa_e2e/verify/internal/ateapipb"
	"github.com/plantonhq/planton/e2e/framework/provider"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// How every Agent Substrate verifier talks to the control plane's API: as a
// product would, from the control plane's outputs alone. The control plane's
// own verifier receives those outputs directly; the verifiers of the resources
// that run on it (worker pools, sandbox configurations) read them from the
// prerequisite the runner deployed in front of the scenario. Either way the
// API is dialed at the output address under the output server name, trusted
// through the output trust bundle, and called with a token minted for the
// output audience by a service account in the probe namespace, the namespace
// the API admits because it carries the output access label.

// agentSubstrateOutputsFromPrerequisite returns the outputs of the one
// KubernetesAgentSubstrate deployed in front of the scenario under test.
func agentSubstrateOutputsFromPrerequisite(ctx context.Context) (map[string]interface{}, error) {
	instances := provider.DependencyOutputsFrom(ctx)[agentSubstrateKind]
	if len(instances) != 1 {
		return nil, errors.Errorf("the proof calls the Agent Substrate API through the control plane deployed in front of the scenario, and %d %s instances were deployed (want exactly one)", len(instances), agentSubstrateKind)
	}
	for _, outputs := range instances {
		return outputs, nil
	}
	return nil, nil
}

// applySubstrateProbe creates the probe namespace carrying the API-access
// label, its two service accounts (a client and a stranger), a small shell
// pod for in-cluster calls, and a router caller as each service account:
// the router accepts a client by the SPIFFE ID of its pod certificate, so
// only an in-cluster workload holding one can call an actor, as a product's
// does. With outsider set it also creates a namespace without the label,
// with its own shell, for the network-policy check.
func applySubstrateProbe(ctx context.Context, kubeconfig string, outputs map[string]interface{}, outsider bool) error {
	accessLabel := fmt.Sprint(outputs["api_access_namespace_label"])
	key, value, ok := strings.Cut(accessLabel, "=")
	if !ok {
		return errors.Errorf("the api_access_namespace_label output %q is not key=value", accessLabel)
	}
	shell := func(namespace string) string {
		return fmt.Sprintf(`---
apiVersion: v1
kind: Pod
metadata:
  name: %s
  namespace: %s
spec:
  containers:
    - name: shell
      image: %s
      command: ["sleep", "3600"]
`, agentSubstrateProbePod, namespace, agentSubstrateProbeImage)
	}
	manifest := fmt.Sprintf(`apiVersion: v1
kind: Namespace
metadata:
  name: %[1]s
  labels:
    %[2]q: %[3]q
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: %[4]s
  namespace: %[1]s
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: %[5]s
  namespace: %[1]s
`, agentSubstrateProbeNamespace, key, value, agentSubstrateProbeClient, agentSubstrateProbeStranger)
	manifest += shell(agentSubstrateProbeNamespace)
	for pod, account := range routerCallers {
		manifest += routerCaller(pod, account, fmt.Sprint(outputs["router_client_signer_name"]), fmt.Sprint(outputs["api_trust_bundle_name"]))
	}
	namespaces := []string{agentSubstrateProbeNamespace}
	if outsider {
		manifest += fmt.Sprintf("---\napiVersion: v1\nkind: Namespace\nmetadata:\n  name: %s\n", agentSubstrateOutsiderNamespace)
		manifest += shell(agentSubstrateOutsiderNamespace)
		namespaces = append(namespaces, agentSubstrateOutsiderNamespace)
	}
	if err := kubectlApplyStdin(ctx, kubeconfig, manifest); err != nil {
		return errors.Wrap(err, "creating the API probe")
	}
	for _, namespace := range namespaces {
		if err := kubectlWait(ctx, kubeconfig, "pod", agentSubstrateProbePod, namespace, "condition=Ready", 3*time.Minute); err != nil {
			return errors.Wrapf(err, "the probe shell in %s never became Ready", namespace)
		}
	}
	for pod := range routerCallers {
		if err := kubectlWait(ctx, kubeconfig, "pod", pod, agentSubstrateProbeNamespace, "condition=Ready", 3*time.Minute); err != nil {
			return errors.Wrapf(err, "the router caller %s never became Ready (its pod certificate is projected before it starts)", pod)
		}
	}
	return nil
}

// routerCallers are the probe's router callers, by pod, each running as one
// of the probe's service accounts: the client, which the lanes' manifests
// list among the router's allowed clients, and the stranger, which they do
// not.
var routerCallers = map[string]string{
	agentSubstrateRouterCaller:   agentSubstrateProbeClient,
	agentSubstrateRouterStranger: agentSubstrateProbeStranger,
}

// routerCaller is a curl pod holding what a product's workload holds to call
// actors: a pod certificate from the router's client signer (the
// router_client_signer_name output), whose SPIFFE ID names its service
// account, and the trust bundle the router's serving certificate chains to
// (the api_trust_bundle_name output; every in-cluster Substrate service is
// signed by that signer).
func routerCaller(pod, serviceAccount, signer, bundle string) string {
	return fmt.Sprintf(`---
apiVersion: v1
kind: Pod
metadata:
  name: %s
  namespace: %s
spec:
  serviceAccountName: %s
  containers:
    - name: curl
      image: %s
      command: ["sleep", "3600"]
      volumeMounts:
        - {name: identity, mountPath: /run/router-client, readOnly: true}
        - {name: trust, mountPath: /run/router-trust, readOnly: true}
  volumes:
    - name: identity
      projected:
        sources:
          - podCertificate: {signerName: %q, keyType: ECDSAP256, credentialBundlePath: credential-bundle.pem}
    - name: trust
      projected:
        sources:
          - clusterTrustBundle: {name: %q, path: trust-bundle.pem}
`, pod, agentSubstrateProbeNamespace, serviceAccount, agentSandboxClientImage, signer, bundle)
}

// callRouter makes one request a product makes to an actor: a POST to the
// router with `ate-target-actor: <atespace>/<actor>`, from the given router
// caller. Over TLS (the Envoy router) the caller presents its pod
// certificate and trusts only the Substrate bundle; over plaintext
// (agentgateway's router_endpoint) it sends the request as is. It returns
// the HTTP status, 0 when the router answered no HTTP at all (a refused TLS
// handshake), and the body.
func callRouter(ctx context.Context, kubeconfig, caller, endpoint, target string) (int, string, error) {
	args := []string{"--kubeconfig", kubeconfig, "exec", caller, "-n", agentSubstrateProbeNamespace, "--",
		"curl", "-sS", "-m", "20", "-X", "POST", "-H", agentSubstrateHeader + ": " + target, "-w", "\n%{http_code}"}
	if strings.HasPrefix(endpoint, "https://") {
		args = append(args, "--cert", "/run/router-client/credential-bundle.pem", "--key", "/run/router-client/credential-bundle.pem",
			"--cacert", "/run/router-trust/trust-bundle.pem")
	}
	args = append(args, strings.TrimSuffix(endpoint, "/")+"/")
	out, err := exec.CommandContext(ctx, "kubectl", args...).CombinedOutput()
	text := strings.TrimRight(string(out), "\n")
	cut := strings.LastIndex(text, "\n")
	code, convErr := strconv.Atoi(strings.TrimSpace(text[cut+1:]))
	if convErr != nil {
		return 0, text, errors.Errorf("no HTTP answer from %s as %s (%v): %s", endpoint, caller, err, firstLines(text, 3))
	}
	body := ""
	if cut > 0 {
		body = text[:cut]
	}
	return code, body, nil
}

// removeSubstrateProbe deletes every namespace applySubstrateProbe can
// create, the outsider included. The control plane's verifier probes with the
// outsider and also runs when the control plane is a resource lane's
// prerequisite, where nothing calls its destroy check; so whichever verifier
// tears the probe down removes both, and a lane leaves no namespace behind on
// a reused cluster.
func removeSubstrateProbe(ctx context.Context, kubeconfig string) error {
	for _, namespace := range []string{agentSubstrateProbeNamespace, agentSubstrateOutsiderNamespace} {
		if out, err := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig, "delete", "namespace",
			namespace, "--ignore-not-found", "--wait=true", "--timeout=3m").CombinedOutput(); err != nil {
			return errors.Wrapf(err, "removing the probe namespace %s: %s", namespace, firstLines(string(out), 3))
		}
	}
	return nil
}

// substrateAPI is a connection to the control plane's API, made from its
// outputs, with tokens minted per probe service account.
type substrateAPI struct {
	control    ateapipb.ControlClient
	kubeconfig string
	address    string
	serverName string
	audience   string
	bundle     string
	close      func()

	mu     sync.Mutex
	tokens map[string]string
}

// dialSubstrateAPI forwards a local port to the Service the api_address
// output names and dials it with TLS: the server name is the api_server_name
// output and the only trusted roots are the ClusterTrustBundle the
// api_trust_bundle_name output names.
func dialSubstrateAPI(ctx context.Context, kubeconfig string, outputs map[string]interface{}) (*substrateAPI, error) {
	a := &substrateAPI{
		kubeconfig: kubeconfig,
		address:    fmt.Sprint(outputs["api_address"]),
		serverName: fmt.Sprint(outputs["api_server_name"]),
		audience:   fmt.Sprint(outputs["api_token_audience"]),
		bundle:     fmt.Sprint(outputs["api_trust_bundle_name"]),
		tokens:     map[string]string{},
	}
	pem, err := kubectlGetJSONPath(ctx, kubeconfig, "clustertrustbundle", a.bundle, "", "{.spec.trustBundle}")
	if err != nil {
		return nil, errors.Wrapf(err, "reading the API's trust bundle %s (the api_trust_bundle_name output)", a.bundle)
	}
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM([]byte(pem)) {
		return nil, errors.Errorf("ClusterTrustBundle %s holds no certificate", a.bundle)
	}

	host, port, err := net.SplitHostPort(a.address)
	if err != nil {
		return nil, errors.Wrapf(err, "the api_address output %q is not host:port", a.address)
	}
	labels := strings.Split(host, ".")
	if len(labels) < 2 {
		return nil, errors.Errorf("the api_address output %q does not name a Service", a.address)
	}
	localPort, err := freeLocalPort()
	if err != nil {
		return nil, err
	}
	stop, err := startPortForward(ctx, kubeconfig, "svc/"+labels[0], labels[1], fmt.Sprintf("%d:%s", localPort, port))
	if err != nil {
		return nil, errors.Wrap(err, "forwarding a local port to the API")
	}
	conn, err := grpc.NewClient(fmt.Sprintf("127.0.0.1:%d", localPort),
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{RootCAs: roots, ServerName: a.serverName, MinVersion: tls.VersionTLS12})))
	if err != nil {
		stop()
		return nil, errors.Wrap(err, "dialing the API")
	}
	a.control = ateapipb.NewControlClient(conn)
	a.close = func() {
		_ = conn.Close()
		stop()
	}
	return a, nil
}

// Close ends the connection and its port-forward.
func (a *substrateAPI) Close() { a.close() }

// token returns a token for a probe service account, minted for the output
// audience once per connection.
func (a *substrateAPI) token(ctx context.Context, serviceAccount string) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if t, ok := a.tokens[serviceAccount]; ok {
		return t, nil
	}
	t, err := mintToken(ctx, a.kubeconfig, serviceAccount, a.audience)
	if err != nil {
		return "", err
	}
	a.tokens[serviceAccount] = t
	return t, nil
}

// as returns a context for one call as the probe service account, bounded by
// timeout.
func (a *substrateAPI) as(ctx context.Context, serviceAccount string, timeout time.Duration) (context.Context, context.CancelFunc, error) {
	t, err := a.token(ctx, serviceAccount)
	if err != nil {
		return nil, nil, err
	}
	c, cancel := context.WithTimeout(ctx, timeout)
	return withBearer(c, t), cancel, nil
}

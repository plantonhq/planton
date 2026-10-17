package module

import (
	"crypto/sha256"
	"encoding/hex"
	"math"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/failure"
	"github.com/plantonhq/planton/shared/catalogkind"
)

// The patches applied onto upstream's own objects. Each one names the
// object it expects by upstream's fixed name and refuses, rather than
// guesses, when the pinned files no longer have that shape. The Terraform
// twin applies the same patches in locals.tf.

// kindName is the ownership label's value on every Deployment the module
// applies (failure.WorkloadOwnerLabel).
var kindName = catalogkind.CatalogKind_KubernetesAgentSandboxController.String()

// resolveImage picks a container's image: an explicit override wins, then
// the registry mirror moves an upstream image, else upstream's stays.
func resolveImage(upstream, override, registry string) string {
	if override != "" {
		return override
	}
	if registry != "" && strings.HasPrefix(upstream, vars.UpstreamRegistry+"/") {
		return registry + strings.TrimPrefix(upstream, vars.UpstreamRegistry)
	}
	return upstream
}

// pullSecrets is the spec's pull secrets plus an image override's own, each
// once, in that order.
func pullSecrets(names []string, overrideSecret string) []interface{} {
	seen := map[string]bool{}
	var out []interface{}
	for _, name := range append(append([]string{}, names...), overrideSecret) {
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		out = append(out, document{"name": name})
	}
	return out
}

// telemetryEnv is the OpenTelemetry exporter environment both binaries
// read (the SDK's own variables).
func telemetryEnv(t *telemetrySettings) []interface{} {
	if t == nil {
		return nil
	}
	env := []interface{}{document{"name": "OTEL_EXPORTER_OTLP_ENDPOINT", "value": t.endpoint}}
	if t.insecure {
		env = append(env, document{"name": "OTEL_EXPORTER_OTLP_INSECURE", "value": "true"})
	}
	if t.headersSecret != "" {
		env = append(env, document{
			"name": "OTEL_EXPORTER_OTLP_HEADERS",
			"valueFrom": document{"secretKeyRef": document{
				"name": t.headersSecret,
				"key":  t.headersKey,
			}},
		})
	}
	return env
}

// applyPodPlacement writes replicas and the pod-level scheduling settings.
func applyPodPlacement(deploymentSpec, podSpec document, replicas int32, nodeSelector map[string]string, tolerations []interface{}, secrets []interface{}) {
	deploymentSpec["replicas"] = int64(replicas)
	if len(nodeSelector) > 0 {
		selector := document{}
		for k, v := range nodeSelector {
			selector[k] = v
		}
		podSpec["nodeSelector"] = selector
	}
	if len(tolerations) > 0 {
		podSpec["tolerations"] = tolerations
	}
	if len(secrets) > 0 {
		podSpec["imagePullSecrets"] = secrets
	}
}

func markOwned(object document) {
	child(child(object, "metadata"), "labels")[failure.WorkloadOwnerLabel] = kindName
}

func appendTo(object document, key string, items ...interface{}) {
	existing, _ := object[key].([]interface{})
	object[key] = append(existing, items...)
}

func secretVolume(name, secret string) document {
	return document{"name": name, "secret": document{"secretName": secret}}
}

func readOnlyMount(name, path string) document {
	return document{"name": name, "mountPath": path, "readOnly": true}
}

// configDigest is what the controller's pod template carries so a change to
// the allowed label domains restarts it.
func configDigest(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}

// configMapContent is the allowed-label-domains value, one domain per line.
func configMapContent(domains []string) string {
	return strings.Join(domains, "\n")
}

// configMap is the controller's runtime configuration, applied only when the
// extensions run and the spec lists label domains (upstream's default
// otherwise).
func configMap(domains []string) document {
	return document{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata":   document{"name": vars.ConfigMapName, "namespace": vars.Namespace},
		"data":       document{vars.ConfigMapKey: configMapContent(domains)},
	}
}

// patchControllerDeployment applies the controller settings onto upstream's
// agent-sandbox-controller Deployment.
func patchControllerDeployment(d document, s *settings) error {
	c := s.controller
	markOwned(d)
	deploymentSpec := child(d, "spec")
	template := child(deploymentSpec, "template")
	podSpec := child(template, "spec")
	container, err := containerNamed(podSpec, vars.ControllerContainerName)
	if err != nil {
		return err
	}

	container["image"] = resolveImage(str(container, "image"), c.image, s.imageRegistry)
	container["args"] = toInterfaces(applyFlags(stringList(container["args"]), controllerFlags(s)))
	if c.resources != nil {
		container["resources"] = c.resources
	}
	if env := telemetryEnv(c.tracing); env != nil {
		container["env"] = env
	}
	ports, _ := container["ports"].([]interface{})
	for _, p := range ports {
		containerPort, _ := p.(map[string]interface{})
		switch containerPort["name"] {
		case "metrics":
			containerPort["containerPort"] = int64(c.metricsPort)
		case "healthz":
			containerPort["containerPort"] = int64(c.healthProbePort)
		}
	}
	// Upstream's manifest wires no probes to the controller's own health
	// endpoints; without a readiness probe a finished rollout only means
	// the container started. The timings are controller-runtime's
	// scaffolding defaults.
	container["livenessProbe"] = document{
		"httpGet":             document{"path": "/healthz", "port": "healthz"},
		"initialDelaySeconds": int64(15),
		"periodSeconds":       int64(20),
	}
	container["readinessProbe"] = document{
		"httpGet":             document{"path": "/readyz", "port": "healthz"},
		"initialDelaySeconds": int64(5),
		"periodSeconds":       int64(10),
	}
	if c.metricsTLS != nil && c.metricsTLS.certificateSecret != "" {
		appendTo(podSpec, "volumes", secretVolume("metrics-tls", c.metricsTLS.certificateSecret))
		appendTo(container, "volumeMounts", readOnlyMount("metrics-tls", vars.MetricsTLSMountPath))
	}
	if s.extensions && len(s.allowedLabelDomains) > 0 {
		child(child(template, "metadata"), "annotations")[vars.ConfigChecksumAnnotation] = configDigest(configMapContent(s.allowedLabelDomains))
	}
	applyPodPlacement(deploymentSpec, podSpec, c.replicas, c.nodeSelector, c.tolerations, pullSecrets(s.imagePullSecrets, c.imagePullSecret))
	return nil
}

// patchControllerService moves the metrics Service's port with the metrics
// endpoint.
func patchControllerService(svc document, s *settings) {
	ports, _ := lookup(svc, "spec", "ports").([]interface{})
	for _, p := range ports {
		if servicePort, _ := p.(map[string]interface{}); servicePort["name"] == "metrics" {
			servicePort["port"] = int64(s.controller.metricsPort)
		}
	}
}

// patchRouterDeployment applies the router settings onto upstream's
// sandbox-router Deployment, pinning its image to the release (the file
// names :latest).
func patchRouterDeployment(d document, s *settings) error {
	r := s.router
	markOwned(d)
	deploymentSpec := child(d, "spec")
	podSpec := child(child(deploymentSpec, "template"), "spec")
	container, err := containerNamed(podSpec, vars.RouterContainerName)
	if err != nil {
		return err
	}
	upstreamImage := str(container, "image")
	if repo := strings.SplitN(upstreamImage, ":", 2)[0]; repo != vars.RouterImageRepo {
		return errors.Errorf("upstream's router deployment at %s runs %s, not %s; the module's pin needs a catalog fix", vars.Release, upstreamImage, vars.RouterImageRepo)
	}
	container["image"] = resolveImage(vars.RouterImageRepo+":"+vars.Release, r.image, s.imageRegistry)
	container["args"] = toInterfaces(applyFlags(stringList(container["args"]), routerFlags(s)))
	if r.resources != nil {
		container["resources"] = r.resources
	}
	if env := telemetryEnv(r.telemetry); env != nil {
		container["env"] = env
	}

	var ports []interface{}
	upstreamPorts, _ := container["ports"].([]interface{})
	for _, p := range upstreamPorts {
		containerPort, _ := p.(map[string]interface{})
		switch containerPort["name"] {
		case "proxy":
			if !r.httpEnabled {
				continue
			}
			containerPort["containerPort"] = int64(r.httpPort)
			ports = append(ports, containerPort)
			if r.tls != nil {
				ports = append(ports, document{"name": "proxy-tls", "containerPort": int64(r.tls.port), "protocol": "TCP"})
			}
		case "metrics":
			containerPort["containerPort"] = int64(r.metricsPort)
			ports = append(ports, containerPort)
		case "healthz":
			containerPort["containerPort"] = int64(r.healthProbePort)
			ports = append(ports, containerPort)
		default:
			ports = append(ports, containerPort)
		}
	}
	if !r.httpEnabled && r.tls != nil {
		ports = append([]interface{}{document{"name": "proxy-tls", "containerPort": int64(r.tls.port), "protocol": "TCP"}}, ports...)
	}
	container["ports"] = ports

	mount := func(volume, secret, path string) {
		appendTo(podSpec, "volumes", secretVolume(volume, secret))
		appendTo(container, "volumeMounts", readOnlyMount(volume, path))
	}
	if t := r.tls; t != nil {
		mount("tls-certs", t.certificateSecret, vars.RouterTLSMountPath)
		if t.clientCA != nil {
			mount("client-ca", t.clientCA.secret, vars.RouterClientCAMountPath)
		}
	}
	if r.scopedTokenV1 != nil {
		mount("scoped-token-v1", r.scopedTokenV1.secret, vars.RouterScopedTokenV1Mount)
	}
	if r.scopedTokenV2 != nil {
		mount("scoped-token-v2", r.scopedTokenV2.secret, vars.RouterScopedTokenV2Mount)
	}

	if r.shutdownTimeout != "30s" {
		grace, err := terminationGraceSeconds(r.shutdownTimeout)
		if err != nil {
			return err
		}
		podSpec["terminationGracePeriodSeconds"] = grace
	}
	applyPodPlacement(deploymentSpec, podSpec, r.replicas, r.nodeSelector, r.tolerations, pullSecrets(s.imagePullSecrets, r.imagePullSecret))
	return nil
}

// terminationGraceSeconds gives in-flight requests the router's drain budget
// plus upstream's margin before the kubelet kills the pod (upstream pairs
// its 30s default with 45s).
func terminationGraceSeconds(shutdownTimeout string) (int64, error) {
	d, err := time.ParseDuration(shutdownTimeout)
	if err != nil {
		return 0, errors.Wrapf(err, "router shutdown timeout %q", shutdownTimeout)
	}
	return int64(math.Ceil(d.Seconds())) + vars.RouterTerminationGraceExtra, nil
}

// patchRouterService publishes the listeners the router serves.
func patchRouterService(svc document, s *settings) {
	r := s.router
	var ports []interface{}
	upstreamPorts, _ := lookup(svc, "spec", "ports").([]interface{})
	for _, p := range upstreamPorts {
		servicePort, _ := p.(map[string]interface{})
		if servicePort["name"] == "proxy" {
			if r.httpEnabled {
				servicePort["port"] = int64(r.httpPort)
				ports = append(ports, servicePort)
			}
			continue
		}
		ports = append(ports, servicePort)
	}
	if r.tls != nil {
		ports = append(ports, document{"name": "proxy-tls", "port": int64(r.tls.port), "targetPort": "proxy-tls", "protocol": "TCP"})
	}
	child(svc, "spec")["ports"] = ports
}

// patchRouterPDB sets the disruption budget.
func patchRouterPDB(pdb document, s *settings) {
	child(pdb, "spec")["minAvailable"] = int64(s.router.minAvailable)
}

// firstPort is a NetworkPolicy rule's first port number, 0 when it has none.
func firstPort(rule interface{}) float64 {
	m, _ := rule.(map[string]interface{})
	ports, _ := m["ports"].([]interface{})
	if len(ports) == 0 {
		return 0
	}
	p, _ := ports[0].(map[string]interface{})
	n, _ := p["port"].(float64)
	return n
}

// patchRouterNetworkPolicy fills the placeholders upstream's policy leaves
// for each cluster: who may call the router, which namespace scrapes it,
// which sandbox ports it may open. The rules it does not name (DNS, the
// OpenTelemetry collector, the API server) stay upstream's.
func patchRouterNetworkPolicy(np document, s *settings) error {
	r := s.router
	spec := child(np, "spec")
	ingress, _ := spec["ingress"].([]interface{})
	egress, _ := spec["egress"].([]interface{})
	if len(ingress) != 3 || len(egress) != 4 ||
		firstPort(ingress[0]) != 8080 || firstPort(ingress[1]) != 9090 || firstPort(ingress[2]) != 8081 || firstPort(egress[1]) != defaultSandboxPort {
		return errors.Errorf("upstream's router NetworkPolicy at %s no longer has the proxy, metrics and probe ingress rules and the sandbox egress rule where the module expects them; the module's pin needs a catalog fix", vars.Release)
	}

	proxyRule, _ := ingress[0].(map[string]interface{})
	selector := document{}
	if len(r.networkPolicy.clientNamespaceLabels) > 0 {
		labels := document{}
		for k, v := range r.networkPolicy.clientNamespaceLabels {
			labels[k] = v
		}
		selector["matchLabels"] = labels
	}
	proxyRule["from"] = []interface{}{document{"namespaceSelector": selector}}
	var proxyPorts []interface{}
	if r.httpEnabled {
		proxyPorts = append(proxyPorts, document{"protocol": "TCP", "port": int64(r.httpPort)})
	}
	if r.tls != nil {
		proxyPorts = append(proxyPorts, document{"protocol": "TCP", "port": int64(r.tls.port)})
	}
	proxyRule["ports"] = proxyPorts

	metricsRule, _ := ingress[1].(map[string]interface{})
	metricsRule["from"] = []interface{}{document{"namespaceSelector": document{"matchLabels": document{
		"kubernetes.io/metadata.name": r.networkPolicy.metricsNamespace,
	}}}}
	metricsRule["ports"] = []interface{}{document{"protocol": "TCP", "port": int64(r.metricsPort)}}

	probeRule, _ := ingress[2].(map[string]interface{})
	probeRule["ports"] = []interface{}{document{"protocol": "TCP", "port": int64(r.healthProbePort)}}

	sandboxRule, _ := egress[1].(map[string]interface{})
	var sandboxPorts []interface{}
	for _, p := range r.networkPolicy.sandboxPorts {
		sandboxPorts = append(sandboxPorts, document{"protocol": "TCP", "port": int64(p)})
	}
	sandboxRule["ports"] = sandboxPorts
	return nil
}

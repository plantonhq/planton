package module

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	kuberneteslocustv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kuberneteslocust/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kuberneteslocustv1alpha1.KubernetesLocustSpec

	// Resource-identity labels stamped on the module-created satellites
	// (the namespace, the script ConfigMaps and the auth Secret — never
	// injected into the chart's own resources; Helm owns those).
	Labels map[string]string

	// Namespace Locust installs into (resolved literal from the spec's
	// value-or-ref).
	Namespace string

	// Helm release name = metadata.name. The module PINS
	// fullnameOverride to the same value, so child names are
	// deterministic: `<name>-master`, `<name>-worker`, and the master
	// Service is the bare `<name>`.
	ReleaseName string

	// The load-test name — labels every chart resource
	// (`load_test: <name>` rides the Deployments' IMMUTABLE selector
	// labels). Empty in the spec = the release name.
	LoadTestName string

	// Script-delivery resolution: the ConfigMap the locustfile mounts
	// from, the locustfile's filename inside it, and the lib ConfigMap
	// ("" = no lib mount). On the inline arm the module renders the
	// ConfigMaps itself (scripts.go).
	LocustfileConfigMap string
	LocustfileName      string
	LibConfigMap        string

	// Web-UI login resolution. WebLoginEnabled is the secured default
	// (true unless explicitly disabled) AND requires a web UI to exist
	// — headless runs never start one, so the login machinery is
	// skipped entirely there.
	WebLoginEnabled bool
	WebUsername     string
	AuthSecretName  string
	WebAuthCodeName string

	// The content hash stamped onto both pod templates — module-owned
	// script/credential-shape changes must roll the pods (the chart
	// checksums only its own ConfigMaps).
	ConfigChecksum string

	// Output handles.
	MasterService      string
	WebEndpoint        string
	MasterBindEndpoint string
	PortForwardCommand string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, iacInput *kuberneteslocustv1alpha1.KubernetesLocustIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesLocust.String(),
	}
	if target.Metadata.Id != "" {
		labels[kuberneteslabelkeys.ResourceId] = target.Metadata.Id
	}
	if target.Metadata.Org != "" {
		labels[kuberneteslabelkeys.Organization] = target.Metadata.Org
	}
	if target.Metadata.Env != "" {
		labels[kuberneteslabelkeys.Environment] = target.Metadata.Env
	}

	locals := &Locals{
		Spec:        spec,
		Labels:      labels,
		Namespace:   spec.Namespace.GetValue(),
		ReleaseName: target.Metadata.Name,
	}

	loadTest := spec.GetLoadTest()

	locals.LoadTestName = loadTest.GetName()
	if locals.LoadTestName == "" {
		locals.LoadTestName = locals.ReleaseName
	}

	// --------------------------- script delivery --------------------------
	// Inline scripts render into module-owned ConfigMaps; the
	// existing-ConfigMaps arm mounts the user's own. Either way the
	// chart values name the ConfigMaps EXPLICITLY — the chart's
	// bundled-example defaults are a fragile literal-string coupling
	// (they render empty the moment loadtest.name changes) and are
	// never engaged.
	if inline := loadTest.GetInline(); inline != nil {
		locals.LocustfileConfigMap = locals.ReleaseName + vars.LocustfileSuffix
		locals.LocustfileName = "main.py"
		if len(inline.GetLibFiles()) > 0 {
			locals.LibConfigMap = locals.ReleaseName + vars.LibSuffix
		}
	} else if existing := loadTest.GetExistingConfigMaps(); existing != nil {
		locals.LocustfileConfigMap = existing.GetLocustfileConfigMap()
		locals.LocustfileName = existing.GetLocustfileName()
		if locals.LocustfileName == "" {
			locals.LocustfileName = "main.py"
		}
		locals.LibConfigMap = existing.GetLibConfigMap()
	}

	// ---------------------------- web-UI login ----------------------------
	// The secured default: an ABSENT web_ui_auth block means the login
	// is ON with a module-generated credential — the chart's own
	// default (an open UI that can fire load at any reachable host)
	// never ships. Headless runs start no web UI, so there is nothing
	// to protect and the login machinery is skipped.
	locals.WebLoginEnabled = !loadTest.GetHeadless()
	if auth := spec.GetWebUiAuth(); auth != nil && auth.Enabled != nil && !auth.GetEnabled() {
		locals.WebLoginEnabled = false
	}
	if locals.WebLoginEnabled {
		locals.WebUsername = spec.GetWebUiAuth().GetUsername()
		if locals.WebUsername == "" {
			locals.WebUsername = "locust"
		}
		locals.AuthSecretName = locals.ReleaseName + vars.AuthSecretSuffix
		locals.WebAuthCodeName = locals.ReleaseName + vars.WebAuthCodeSuffix
	}

	locals.ConfigChecksum = configChecksum(loadTest, locals)

	// ------------------------------- outputs ------------------------------
	locals.MasterService = locals.ReleaseName
	locals.WebEndpoint = fmt.Sprintf("http://%s.%s.svc.cluster.local:%d",
		locals.MasterService, locals.Namespace, vars.WebPort)
	locals.MasterBindEndpoint = fmt.Sprintf("%s.%s.svc.cluster.local:%d",
		locals.MasterService, locals.Namespace, vars.MasterBindPort)
	locals.PortForwardCommand = fmt.Sprintf("kubectl port-forward svc/%s -n %s %d:%d",
		locals.MasterService, locals.Namespace, vars.WebPort, vars.WebPort)

	return locals
}

// configChecksum hashes every module-owned input that reaches the pods
// OUTSIDE chart-rendered resources — inline script content, the lib
// files, the login-backend code and the credential wiring shape. The
// chart's own checksum annotations cover only ITS ConfigMaps/Secret, so
// without this hash a locustfile edit would update the ConfigMap and
// roll nothing. Terraform twin: the sha256 over the same ordered parts.
func configChecksum(loadTest *kuberneteslocustv1alpha1.KubernetesLocustLoadTest, locals *Locals) string {
	parts := []string{
		"locustfile-configmap=" + locals.LocustfileConfigMap,
		"locustfile-name=" + locals.LocustfileName,
		"lib-configmap=" + locals.LibConfigMap,
		"web-login=" + strconv.FormatBool(locals.WebLoginEnabled),
	}
	if inline := loadTest.GetInline(); inline != nil {
		parts = append(parts, "locustfile-content="+inline.GetLocustfileContent())
		libNames := make([]string, 0, len(inline.GetLibFiles()))
		for name := range inline.GetLibFiles() {
			libNames = append(libNames, name)
		}
		sort.Strings(libNames)
		for _, name := range libNames {
			parts = append(parts, "lib:"+name+"="+inline.GetLibFiles()[name])
		}
	}
	if locals.WebLoginEnabled {
		parts = append(parts, "web-auth-code="+webAuthBackendPy, "web-username="+locals.WebUsername)
	}
	// The record-separator join is expressible in BOTH engines (HCL
	// cannot carry NUL) — the hashes must stay byte-identical across
	// the twins.
	sum := sha256.Sum256([]byte(strings.Join(parts, "\x1e")))
	return hex.EncodeToString(sum[:])
}

// imageTagAllowsWebLogin enforces the login floor: the chart renders the
// modern `--web-login` flag only for tags >= 2.21.0 and otherwise falls
// onto rendering credentials as a LITERAL POD ARGUMENT — a path this
// module refuses. Non-numeric tags (e.g. "latest") cannot prove the
// floor, so they fail too.
func imageTagAllowsWebLogin(tag string) bool {
	matches := regexp.MustCompile(`^v?(\d+)\.(\d+)`).FindStringSubmatch(tag)
	if matches == nil {
		return false
	}
	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	if major != vars.AuthMinMajor {
		return major > vars.AuthMinMajor
	}
	return minor >= vars.AuthMinMinor
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

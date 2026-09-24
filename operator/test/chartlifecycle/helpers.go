//go:build e2e

package chartlifecycle

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2" //nolint:revive,staticcheck
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"sigs.k8s.io/yaml"

	"github.com/plantonhq/planton/operator/internal/plantonregistry"
	"github.com/plantonhq/planton/operator/test/utils"
)

const (
	chartRegistry = plantonregistry.DefaultChartRepository

	// lastInstallOnceOperatorChart is the newest published operator chart that
	// shipped its definitions through Helm's install-once crds/ directory. An
	// install of it is the operator-chart upgrade path this suite rehearses.
	lastInstallOnceOperatorChart = "0.7.1"

	// lastBundlingPlatformChart is the newest published planton chart that
	// bundled the operator as a subchart. An install of it is the other
	// upgrade path this suite rehearses.
	lastBundlingPlatformChart = "0.3.1"

	platformCRD         = "plantonplatforms.planton.ai"
	identityProviderCRD = "plantonidentityproviders.planton.ai"
)

// projectDir is the operator module root; utils.Run changes the process's
// working directory to it, so every path here is resolved from it explicitly.
func projectDir() string {
	dir, err := utils.GetProjectDir()
	if err != nil {
		Fail(fmt.Sprintf("locating the operator directory: %v", err))
	}
	return dir
}

// workingTreeChart loads a chart from this checkout's helm/ directory.
func workingTreeChart(name string) *chart.Chart {
	chrt, err := loader.Load(filepath.Join(projectDir(), "..", "helm", name))
	if err != nil {
		Fail(fmt.Sprintf("loading working-tree chart %s: %v", name, err))
	}
	return chrt
}

// publishedChart pulls a chart at a pinned version from the public registry,
// the way the CLI installer resolves the chart it installs.
func publishedChart(name, version string) *chart.Chart {
	registryClient, err := registry.NewClient()
	if err != nil {
		Fail(fmt.Sprintf("creating the registry client: %v", err))
	}
	locate := action.NewInstall(&action.Configuration{})
	locate.SetRegistryClient(registryClient)
	locate.ChartPathOptions.Version = version
	path, err := locate.ChartPathOptions.LocateChart(chartRegistry+"/"+name, cli.New())
	if err != nil {
		Fail(fmt.Sprintf("pulling %s %s: %v", name, version, err))
	}
	chrt, err := loader.Load(path)
	if err != nil {
		Fail(fmt.Sprintf("loading %s %s: %v", name, version, err))
	}
	return chrt
}

// helm drives releases in one namespace through the Helm SDK: the same
// action package the helm CLI, the Terraform provider, Pulumi's Helm release,
// and the Planton CLI installer all run.
type helm struct {
	namespace string
	cfg       *action.Configuration
}

func newHelm(namespace string) *helm {
	flags := genericclioptions.NewConfigFlags(false)
	flags.Namespace = &namespace
	cfg := new(action.Configuration)
	if err := cfg.Init(flags, namespace, "", func(format string, v ...any) {
		_, _ = fmt.Fprintf(GinkgoWriter, "helm: "+format+"\n", v...)
	}); err != nil {
		Fail(fmt.Sprintf("initializing helm for namespace %s: %v", namespace, err))
	}
	registryClient, err := registry.NewClient()
	if err != nil {
		Fail(fmt.Sprintf("creating the registry client: %v", err))
	}
	cfg.RegistryClient = registryClient
	return &helm{namespace: namespace, cfg: cfg}
}

func (h *helm) install(release string, chrt *chart.Chart, values map[string]any) error {
	install := action.NewInstall(h.cfg)
	install.ReleaseName = release
	install.Namespace = h.namespace
	install.CreateNamespace = true
	_, err := install.Run(chrt, values)
	return err
}

func (h *helm) upgrade(release string, chrt *chart.Chart, values map[string]any) error {
	upgrade := action.NewUpgrade(h.cfg)
	upgrade.Namespace = h.namespace
	_, err := upgrade.Run(release, chrt, values)
	return err
}

// uninstall removes a release; a release that does not exist is not an error,
// so cleanup can call it unconditionally.
func (h *helm) uninstall(release string) error {
	_, err := action.NewUninstall(h.cfg).Run(release)
	if err != nil && strings.Contains(err.Error(), driver.ErrReleaseNotFound.Error()) {
		return nil
	}
	return err
}

func kubectl(args ...string) (string, error) {
	return utils.Run(exec.Command("kubectl", args...))
}

func crdExists(name string) bool {
	_, err := kubectl("get", "crd", name)
	return err == nil
}

func crdAnnotation(name, key string) string {
	out, err := kubectl("get", "crd", name, "-o", fmt.Sprintf("jsonpath={.metadata.annotations['%s']}", strings.ReplaceAll(key, ".", "\\.")))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

func crdLabel(name, key string) string {
	out, err := kubectl("get", "crd", name, "-o", fmt.Sprintf("jsonpath={.metadata.labels['%s']}", strings.ReplaceAll(key, ".", "\\.")))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

// ownedBy reports whether the definition carries the three pieces of Helm
// ownership metadata for the given release.
func ownedBy(crdName, release, namespace string) bool {
	return crdLabel(crdName, "app.kubernetes.io/managed-by") == "Helm" &&
		crdAnnotation(crdName, "meta.helm.sh/release-name") == release &&
		crdAnnotation(crdName, "meta.helm.sh/release-namespace") == namespace
}

// clusterSchema is the definition's served schema as the API server holds it.
func clusterSchema(crdName string) any {
	out, err := kubectl("get", "crd", crdName, "-o", "jsonpath={.spec.versions[0].schema.openAPIV3Schema}")
	if err != nil {
		Fail(fmt.Sprintf("reading the schema of %s: %v", crdName, err))
	}
	var schema any
	if err := json.Unmarshal([]byte(out), &schema); err != nil {
		Fail(fmt.Sprintf("parsing the schema of %s: %v", crdName, err))
	}
	return schema
}

// controllerGenSchema is the schema controller-gen derived from the working
// tree's Go types, which is what the working-tree chart installs (the chart
// render test holds the two byte-identical).
func controllerGenSchema(crdFile string) any {
	data, err := os.ReadFile(filepath.Join(projectDir(), "config", "crd", "bases", crdFile))
	if err != nil {
		Fail(fmt.Sprintf("reading %s: %v", crdFile, err))
	}
	var crd struct {
		Spec struct {
			Versions []struct {
				Schema struct {
					OpenAPIV3Schema any `json:"openAPIV3Schema"`
				} `json:"schema"`
			} `json:"versions"`
		} `json:"spec"`
	}
	if err := yaml.Unmarshal(data, &crd); err != nil {
		Fail(fmt.Sprintf("parsing %s: %v", crdFile, err))
	}
	return crd.Spec.Versions[0].Schema.OpenAPIV3Schema
}

// runKubectlLinesOf executes, verbatim, every line of a failure message that
// is a kubectl command. The preflight's promise is that its next step can be
// followed as printed; running the printed text is the proof.
func runKubectlLinesOf(message string) int {
	ran := 0
	for _, line := range strings.Split(message, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "kubectl ") {
			continue
		}
		if _, err := kubectl(strings.Fields(line)[1:]...); err != nil {
			Fail(fmt.Sprintf("running the printed command %q: %v", line, err))
		}
		ran++
	}
	return ran
}

// clean removes every definition and namespace a scenario may have left,
// so the scenarios stay independent of one another.
func clean(namespaces ...string) {
	_, _ = kubectl("delete", "crd", platformCRD, identityProviderCRD, "--ignore-not-found", "--wait=true")
	for _, ns := range namespaces {
		_, _ = kubectl("delete", "namespace", ns, "--ignore-not-found", "--wait=true")
	}
}

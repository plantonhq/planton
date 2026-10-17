package verify

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"testing"
)

// The agent-sandbox lanes are only as honest as the verifier's reading of
// their scenarios: which definitions it expects, who owns the namespace,
// where and how it calls the router. Each committed scenario is read here
// offline, so a scenario edit that changes what a lane asserts shows up as a
// failing test, not as a lane that quietly proves something else. The release
// the stamps must carry is held equal to both modules' pin.

const agentSandboxScenarios = "../../kubernetesagentsandboxcontroller/e2e/scenarios"

func TestAgentSandboxVerifierReadsTheScenarios(t *testing.T) {
	core := agentSandboxCoreCrds
	all := append(append([]string{}, agentSandboxCoreCrds...), agentSandboxExtensionCrds...)
	cases := map[string]struct {
		extensions, createNamespace, keep, refused bool
		labelDomain                                string
		crds                                       []string
		router                                     *agentSandboxRouter
	}{
		"minimal.yaml": {createNamespace: true, keep: true, labelDomain: "sandbox.users.io", crds: core},
		"sdk-ready.yaml": {extensions: true, createNamespace: true, labelDomain: "sandbox.users.io", crds: all,
			router: &agentSandboxRouter{URL: "http://sandbox-router-svc.agent-sandbox-system.svc.cluster.local:8080", ClientNamespaceLabels: map[string]string{}}},
		"full-surface.yaml": {extensions: true, labelDomain: "example.com", crds: all,
			router: &agentSandboxRouter{URL: "https://sandbox-router-svc.agent-sandbox-system.svc.cluster.local:8443", TokenReview: true, RequireToken: true, ClientNamespaceLabels: map[string]string{}}},
		"upgrade.yaml": {extensions: true, createNamespace: true, labelDomain: "example.com", crds: all},
		"upgrade.to.yaml": {createNamespace: true, labelDomain: "example.com", crds: core,
			router: &agentSandboxRouter{URL: "http://sandbox-router-svc.agent-sandbox-system.svc.cluster.local:8080", ClientNamespaceLabels: map[string]string{}}},
		"crds-owned-elsewhere.yaml":       {extensions: true, createNamespace: true, refused: true, labelDomain: "sandbox.users.io", crds: []string{"sandboxclaims.extensions.agents.x-k8s.io"}},
		"controller-owned-elsewhere.yaml": {keep: true, refused: true, labelDomain: "sandbox.users.io", crds: core},
		"namespace-missing.yaml":          {keep: true, refused: true, labelDomain: "sandbox.users.io", crds: core},
		"namespace-owned-elsewhere.yaml":  {createNamespace: true, keep: true, refused: true, labelDomain: "sandbox.users.io", crds: core},
	}
	files, err := filepath.Glob(filepath.Join(agentSandboxScenarios, "*.yaml"))
	if err != nil || len(files) != len(cases) {
		t.Fatalf("every scenario needs a case here: found %d files, %d cases (%v)", len(files), len(cases), err)
	}
	for name, want := range cases {
		t.Run(name, func(t *testing.T) {
			v := newAgentSandboxControllerVerifier(filepath.Join(agentSandboxScenarios, name))
			if v.Extensions != want.extensions || v.CreateNamespace != want.createNamespace || v.LabelDomain != want.labelDomain {
				t.Errorf("extensions=%t createNamespace=%t labelDomain=%q, want %t %t %q", v.Extensions, v.CreateNamespace, v.LabelDomain, want.extensions, want.createNamespace, want.labelDomain)
			}
			if v.KeepOnUninstall != want.keep || v.DeployRefused != want.refused || v.ChartVersion != agentSandboxRelease || v.SourceLabel != agentSandboxBundle {
				t.Errorf("keep=%t refused=%t version=%q source=%q", v.KeepOnUninstall, v.DeployRefused, v.ChartVersion, v.SourceLabel)
			}
			if !reflect.DeepEqual(v.CRDs, want.crds) {
				t.Errorf("CRDs %v, want %v", v.CRDs, want.crds)
			}
			if !reflect.DeepEqual(v.Router, want.router) {
				t.Errorf("router %+v, want %+v", v.Router, want.router)
			}
		})
	}
}

func TestAgentSandboxVerifierPinsTheModulesRelease(t *testing.T) {
	for path, pattern := range map[string]*regexp.Regexp{
		"../../kubernetesagentsandboxcontroller/iac/pulumi/module/vars.go": regexp.MustCompile(`Release:\s+"([^"]+)"`),
		"../../kubernetesagentsandboxcontroller/iac/tf/locals.tf":          regexp.MustCompile(`release\s*=\s*"([^"]+)"`),
	} {
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		match := pattern.FindSubmatch(src)
		if match == nil || string(match[1]) != agentSandboxRelease {
			t.Errorf("%s pins %q; the verifier expects the definitions stamped with %s", path, match, agentSandboxRelease)
		}
	}
}

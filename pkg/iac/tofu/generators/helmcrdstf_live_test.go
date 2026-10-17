package generators

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestHelmCRDsTFRefusals drives the generated block directly, with no kind
// around it, and asserts the three-part text of every refusal the block
// phrases itself: one scratch module per case, `tofu init` and `tofu plan`
// against a real cluster (the render reads the server version). It is the
// Terraform twin of helmcrds's fixture tests and the one place the block's
// own words are proven independently of any catalog kind; the refusals a
// kind's lane also proves (a downgrade, a foreign owner, a denied right)
// live in the kinds' e2e scenarios.
//
// Live by nature, so it runs only when asked: HELM_CRDS_TF_LIVE=1 with a
// kubeconfig in KUBE_CONFIG_PATH and `tofu` on PATH. The cases reach public
// chart repositories, so they need egress.
func TestHelmCRDsTFRefusals(t *testing.T) {
	if os.Getenv("HELM_CRDS_TF_LIVE") != "1" {
		t.Skip("set HELM_CRDS_TF_LIVE=1 (with KUBE_CONFIG_PATH and tofu on PATH) to drive the block against a cluster")
	}
	if os.Getenv("KUBE_CONFIG_PATH") == "" {
		t.Fatal("KUBE_CONFIG_PATH must point at the cluster the render reads its Kubernetes version from")
	}
	if _, err := exec.LookPath("tofu"); err != nil {
		t.Fatal("tofu is not on PATH")
	}

	cases := []struct {
		name        string
		locals      string
		observedHas string
		nextStepHas string
	}{
		{
			name:        "version not in the repository index",
			locals:      contractLocals(`expect_crds = false`, "99.99.99", "spec.chart_version", chartBranch("https://stefanprodan.github.io/podinfo", "podinfo", `""`, `[]`), "null"),
			observedHas: "chart podinfo 99.99.99 is not in the index of https://stefanprodan.github.io/podinfo",
			nextStepHas: "index.yaml",
		},
		{
			name:        "typed kind whose render yields no CRD",
			locals:      contractLocals(`expect_crds = true`, "6.7.1", "spec.chart_version", chartBranch("https://stefanprodan.github.io/podinfo", "podinfo", `"crds:\n  create: true\n"`, `[]`), "null"),
			observedHas: "produced no CustomResourceDefinition documents",
			nextStepHas: "helm show values podinfo",
		},
		{
			name:        "helm-managed CRDs without the keep mark, dial off",
			locals:      contractLocals(`expect_crds = false`, "2.8.0", "spec.chart_version", chartBranch("https://opensearch-project.github.io/opensearch-k8s-operator/", "opensearch-operator", `""`, `["installCRDs: true\n"]`), "null"),
			observedHas: "templates 10 CustomResourceDefinition(s) as release resources without the helm.sh/resource-policy: keep annotation",
			nextStepHas: "allow_helm_managed",
		},
		{
			name:        "bundle URL answers 404 at a version a user chose",
			locals:      contractLocals(`expect_crds = true`, "0.0.1", "spec.chart_version", "null", bundleBranch("solr-operator", "https://archive.apache.org/dist/solr/solr-operator/v{{version}}/crds/all-with-dependencies.yaml")),
			observedHas: "answered HTTP 404",
			nextStepHas: "confirm version 0.0.1 exists at the upstream download page; set spec.chart_version to a published version",
		},
		{
			name:        "bundle URL answers 404 at the release the module pins",
			locals:      contractLocals(`expect_crds = true`, "v0.0.1", "", "null", bundleBranch("agent-sandbox", "https://github.com/kubernetes-sigs/agent-sandbox/releases/download/{{version}}/sandbox.yaml")),
			observedHas: "answered HTTP 404",
			nextStepHas: "confirm what upstream publishes for agent-sandbox release v0.0.1",
		},
		{
			name:        "bundle URL serves no CRD",
			locals:      contractLocals(`expect_crds = true`, "6.7.1", "", "null", bundleBranch("podinfo", "https://raw.githubusercontent.com/stefanprodan/podinfo/{{version}}/charts/podinfo/Chart.yaml")),
			observedHas: "contained no CustomResourceDefinition documents",
			nextStepHas: "correct the module's bundle URL template",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			write := func(name, content string) {
				if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			write("provider.tf", scratchProviders)
			write("locals.tf", tc.locals)
			write(HelmCRDsTFFileName, HelmCRDsTF())

			run := func(args ...string) (string, error) {
				cmd := exec.Command("tofu", append(args, "-no-color", "-input=false")...)
				cmd.Dir = root
				cmd.Env = append(os.Environ(), helmIsolation(t.TempDir())...)
				out, err := cmd.CombinedOutput()
				return string(out), err
			}
			if out, err := run("init"); err != nil {
				t.Fatalf("tofu init: %v\n%s", err, out)
			}
			out, err := run("plan")
			if err == nil {
				t.Fatalf("the plan must be refused; it succeeded:\n%s", tail(out, 20))
			}
			text := strings.Join(strings.Fields(out), " ")
			for _, part := range []string{"observed:", "meaning:", "next step:"} {
				if !strings.Contains(text, part) {
					t.Fatalf("the refusal lacks its %q part:\n%s", part, tail(out, 40))
				}
			}
			if !strings.Contains(text, tc.observedHas) {
				t.Fatalf("observation lacks %q:\n%s", tc.observedHas, tail(out, 40))
			}
			if !strings.Contains(text, tc.nextStepHas) {
				t.Fatalf("next step lacks %q:\n%s", tc.nextStepHas, tail(out, 40))
			}
		})
	}
}

// contractLocals is a module's locals.tf as the block's header specifies it:
// the policy knob a case varies, the version and the field that chooses it
// ("" when the module pins it), and exactly one of the two branches as HCL
// (the other "null").
func contractLocals(expectCRDs, version, versionField, chart, bundle string) string {
	return fmt.Sprintf(`locals {
  helm_crds_args = {
    install           = true
    keep_on_uninstall = true
    %s
    version           = %q
    version_field     = %q
    chart             = %s
    bundle            = %s
  }
}
`, expectCRDs, version, versionField, chart, bundle)
}

// chartBranch is the contract's chart object for a public chart rendered
// under a scratch identity.
func chartBranch(repository, name, renderOverride, values string) string {
	return fmt.Sprintf(`{
      repository          = %q
      name                = %q
      release_name        = "refusal"
      namespace           = "refusal"
      values              = %s
      render_override     = %s
      api_versions        = []
      allow_helm_managed  = false
      repository_username = ""
      repository_password = ""
      set                 = []
      set_sensitive       = []
    }`, repository, name, values, renderOverride)
}

// bundleBranch is the contract's bundle object.
func bundleBranch(name, url string) string {
	return fmt.Sprintf(`{
      name = %q
      url  = %q
    }`, name, url)
}

// scratchProviders pins the same four providers a derive-branch module pins,
// configured from the environment (KUBE_CONFIG_PATH) as the modules are.
const scratchProviders = `terraform {
  required_providers {
    kubernetes = { source = "hashicorp/kubernetes", version = ">= 2.20" }
    helm       = { source = "hashicorp/helm", version = ">= 3.1" }
    kubectl    = { source = "alekc/kubectl", version = ">= 2.0" }
    http       = { source = "hashicorp/http", version = "~> 3.4" }
  }
}
provider "kubernetes" {}
provider "helm" {}
provider "kubectl" {}
`

// helmIsolation points Helm's repository configuration at an empty directory,
// as the runner and the e2e harness do, so a stale entry on this machine
// cannot fail a render.
func helmIsolation(dir string) []string {
	return []string{
		"HELM_REPOSITORY_CONFIG=" + filepath.Join(dir, "repositories.yaml"),
		"HELM_REPOSITORY_CACHE=" + filepath.Join(dir, "cache"),
		"HELM_REGISTRY_CONFIG=" + filepath.Join(dir, "registry.json"),
	}
}

func tail(s string, lines int) string {
	all := strings.Split(strings.TrimSpace(s), "\n")
	if len(all) > lines {
		all = all[len(all)-lines:]
	}
	return strings.Join(all, "\n")
}

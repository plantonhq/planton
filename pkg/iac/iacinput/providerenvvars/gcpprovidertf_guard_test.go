package providerenvvars

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// canonicalGcpProviderProject is the lines every Google OpenTofu module whose resources take a
// project from the spec carries in each `provider "google"` and `provider "google-beta"` block.
// Google's provider fills an imported resource's project from the provider's default when the
// import ID carries none (a bucket imported by its name); with no default the state records an
// empty project and the next preview replaces the resource. Handing the provider the component's
// own project makes every import form record it. Credentials stay injected by the runtime as
// environment variables (see loadGcpEnvVars), and the connection never names a project.
// local.project_id is the spec's project or null, never a google_client_config fallback: a
// provider cannot depend on its own data source. If you are adding a new GCP kind, copy these
// lines verbatim into its iac/tf/provider.tf, and pass the same project to the Pulumi builder.
const canonicalGcpProviderProject = `  # The project this component's spec names (null when it names none). An
  # import records it, so a taken-over resource is never planned for
  # replacement. Credentials are injected by the runtime; the connection
  # never names a project.
  project = local.project_id
`

// gcpKindsWithoutOwnProject are the Google kinds whose resources take no project from the spec:
// their project, if any, is carried by the path of the resource they attach to, or they live
// above projects. Their provider blocks stay empty and their Pulumi modules pass "". A kind that
// gains a spec project leaves this list and takes the canonical lines.
var gcpKindsWithoutOwnProject = map[string]string{
	"gcpbillingbudget":              "a budget belongs to a billing account",
	"gcpcloudidentitygroup":         "a group belongs to a Cloud Identity customer",
	"gcpfolder":                     "a folder lives under an organization or folder",
	"gcpgcsbucketiammember":         "the grant names its bucket; it has no project argument",
	"gcpgkeworkloadidentitybinding": "the grant names its service account by path; the pool project only shapes the member",
	"gcphierarchicalfirewallpolicy": "the policy attaches to an organization or folder",
	"gcpiamdenypolicy":              "the policy attaches to its parent by path; it has no project argument",
	"gcpkmskey":                     "the key's project is carried by its key ring's path",
	"gcpkmskeyiammember":            "the grant names its key by path",
	"gcporgpolicy":                  "the policy attaches to its parent by path",
	"gcporgpolicycustomconstraint":  "the constraint belongs to an organization",
	"gcpproject":                    "it creates the project itself",
	"gcppubsubtopiciammember":       "the grant names its topic by path",
	"gcpserviceaccountiammember":    "the grant names its service account by path",
	"gcpsharedvpcserviceproject":    "it joins two projects, neither of them the provider's",
	"gcptagbinding":                 "the binding names its resource by full name",
	"gcptagkey":                     "the key attaches to its parent by path",
	"gcptagvalue":                   "the value belongs to its tag key",
	"gcpvertexaideployedindex":      "the deployment inherits its index endpoint's project",
	"gcpvpcpeering":                 "the peering spans two networks, each named by self link",
}

var (
	gcpProviderBlock = regexp.MustCompile(`(?s)provider "google(?:-beta)?" \{.*?\n?\}`)
	// Credentials and project wiring that must never be written into the HCL: credentials flow
	// via env injection, and a connection-level project is out by design.
	gcpProviderForbidden = regexp.MustCompile(`(?m)^\s*(credentials|access_token|impersonate_service_account)\s*=|var\.provider_config`)
	// The project argument a Google Pulumi module hands its provider builder.
	gcpPulumiBuilderProject = regexp.MustCompile(
		`pulumigoogleprovider\.Get\w*\(\s*ctx,\s*iacInput\.ProviderConfig,\s*([^,)]+)`)
)

// TestGcpProviderTfConvergence enforces that every Google tofu module either carries the
// canonical project lines in each google / google-beta provider block, or is a kind without a
// project of its own and leaves its blocks without them -- so no import can lose the project,
// and the exemption list cannot go stale.
func TestGcpProviderTfConvergence(t *testing.T) {
	// This guard reads the catalog/ source tree, which is not present in the Bazel test sandbox
	// (Bazel sets TEST_SRCDIR). It runs under `go test` / `make test` and CI go-test instead.
	if os.Getenv("TEST_SRCDIR") != "" {
		t.Skip("convergence guard reads the catalog source tree; skipped in the bazel sandbox")
	}

	root := repoRoot(t)
	matches, err := filepath.Glob(filepath.Join(root,
		"catalog", "gcp", "*", "iac", "tf", "provider.tf"))
	require.NoError(t, err)

	// Sized assertion: a new GCP tofu kind must take a side (bump this with intent).
	assert.Len(t, matches, 186, "unexpected number of GCP tofu provider.tf files")

	seenExempt := 0
	for _, path := range matches {
		kind := filepath.Base(filepath.Dir(filepath.Dir(filepath.Dir(path))))
		b, err := os.ReadFile(path)
		require.NoError(t, err)
		content := string(b)

		assert.Falsef(t, gcpProviderForbidden.MatchString(content),
			"%s wires credentials or the provider config into HCL; they flow via env injection", rel(root, path))

		blocks := gcpProviderBlock.FindAllString(content, -1)
		require.NotEmptyf(t, blocks, "%s declares no google provider block", rel(root, path))

		reason, exempt := gcpKindsWithoutOwnProject[kind]
		if exempt {
			seenExempt++
			assert.NotContainsf(t, content, "project = local.project_id",
				"%s names a project but its kind is listed without one (%s): remove it from the list",
				rel(root, path), reason)
			continue
		}
		for _, block := range blocks {
			assert.Containsf(t, block, canonicalGcpProviderProject,
				"%s has a provider block without the component's project:\n%s", rel(root, path), block)
		}
	}
	assert.Equal(t, len(gcpKindsWithoutOwnProject), seenExempt,
		"a kind listed without a project of its own has no tofu module")
}

// TestGcpPulumiProviderProjectConvergence is the Pulumi half: every Google Pulumi module hands its
// provider builder the component's project, and a kind without a project of its own hands it "".
func TestGcpPulumiProviderProjectConvergence(t *testing.T) {
	if os.Getenv("TEST_SRCDIR") != "" {
		t.Skip("convergence guard reads the catalog source tree; skipped in the bazel sandbox")
	}

	root := repoRoot(t)
	modules, err := filepath.Glob(filepath.Join(root, "catalog", "gcp", "*", "iac", "pulumi", "module"))
	require.NoError(t, err)
	assert.Len(t, modules, 186, "unexpected number of GCP pulumi modules")

	for _, module := range modules {
		kind := filepath.Base(filepath.Dir(filepath.Dir(filepath.Dir(module))))
		sources, err := filepath.Glob(filepath.Join(module, "*.go"))
		require.NoError(t, err)
		var calls [][]string
		for _, source := range sources {
			if strings.HasSuffix(source, "_test.go") {
				continue
			}
			b, err := os.ReadFile(source)
			require.NoError(t, err)
			calls = append(calls, gcpPulumiBuilderProject.FindAllStringSubmatch(string(b), -1)...)
		}
		require.NotEmptyf(t, calls, "%s never builds its google provider through pulumigoogleprovider", rel(root, module))

		_, exempt := gcpKindsWithoutOwnProject[kind]
		for _, call := range calls {
			project := strings.TrimSpace(call[1])
			if exempt {
				assert.Equalf(t, `""`, project, "%s is listed without a project of its own", rel(root, module))
			} else {
				assert.NotEqualf(t, `""`, project,
					"%s hands its provider no project; pass the component's own", rel(root, module))
			}
		}
	}
}

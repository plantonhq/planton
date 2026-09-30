package tofumodule

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	cloudflareworkerv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflareworker/v1alpha1"
	"github.com/plantonhq/planton/internal/manifest"
	"github.com/plantonhq/planton/pkg/fileutil"
	"github.com/plantonhq/planton/pkg/iac/stackinput"
	"github.com/plantonhq/planton/pkg/iac/stackinput/stackinputproviderconfig"
	"github.com/plantonhq/planton/pkg/iac/tofu/generators"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

// Credential-isolated plans of the CloudflareWorker OpenTofu module, driven through the same
// seam the CLI and the platform runner use: a stack input built from a manifest and a provider
// config, handed to GetProviderConfigEnvVars, then `init` + `plan` with ONLY those variables and
// PATH in the environment -- no AWS_*, an empty HOME, no ~/.aws. `tofu validate` never configures
// a provider, so a module that configures a provider it does not use on the default path (and
// that provider demands a credential the connection never supplies) passes validate and every
// live lane whose machine happens to hold that credential, and fails only for a customer whose
// runner does not. Only a plan run like this one catches that class.
//
// Two cases:
//   - inline: the committed e2e manifest (inline content). Nothing but the Cloudflare token may be
//     needed to plan it.
//   - r2 bundle: the connection carries an R2 key pair and an endpoint override pointing at a
//     local fake S3 server that serves the bundle as application/javascript (a type the storage
//     provider's plain `body` attribute leaves empty). The plan must sign with exactly the
//     connection's key, and the bundle's bytes must land in the planned script content.
//
// The providers are downloaded from the registry on a cold plugin cache (pass TF_PLUGIN_CACHE_DIR
// to reuse one), so this runs only when asked: PLANTON_TF_PLAN_GUARDS=1 with `tofu` on PATH.
// `terraform` is exercised too when it is on PATH -- the module serves both engines. The guard
// reads the catalog source tree, which the Bazel sandbox does not carry.

const planGuardsEnv = "PLANTON_TF_PLAN_GUARDS"

const (
	isolatedPlanToken     = "cf-isolated-plan-token-0123456789abcdef"
	isolatedPlanAccountID = "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
	isolatedPlanAccessKey = "r2-isolated-plan-access-key"
	isolatedPlanSecretKey = "r2-isolated-plan-secret-access-key-0123456789"
	isolatedPlanBucket    = "worker-builds"
	isolatedPlanKey       = "api/index.js"
	isolatedPlanBundle    = "export default { async fetch() { return new Response(\"from-r2-bundle\"); } };\n"
)

func TestCloudflareWorkerPlansWithOnlyItsConnection_Inline(t *testing.T) {
	root, binaries := isolatedPlanPreconditions(t)
	manifestObject, err := manifest.LoadManifest(filepath.Join(root,
		"catalog", "cloudflare", "cloudflareworker", "e2e", "manifest.yaml"))
	if err != nil {
		t.Fatalf("load the Worker's e2e manifest: %v", err)
	}

	for _, binary := range binaries {
		t.Run(binary, func(t *testing.T) {
			workDir := isolatedWorkerModule(t, root)
			providerConfig := writeCloudflareProviderConfig(t, "")
			out, err := isolatedPlan(t, binary, workDir, manifestObject, providerConfig)
			if err != nil {
				t.Fatalf("an inline-content Worker must plan with nothing but its Cloudflare connection:\n%s", out)
			}
		})
	}
}

func TestCloudflareWorkerPlansWithOnlyItsConnection_R2Bundle(t *testing.T) {
	root, binaries := isolatedPlanPreconditions(t)

	for _, binary := range binaries {
		t.Run(binary, func(t *testing.T) {
			s3 := newFakeR2(t)
			workDir := isolatedWorkerModule(t, root)
			providerConfig := writeCloudflareProviderConfig(t, fmt.Sprintf(`r2:
  accessKeyId: %s
  secretAccessKey: %s
  endpoint: %s
`, isolatedPlanAccessKey, isolatedPlanSecretKey, s3.URL))

			out, err := isolatedPlan(t, binary, workDir, bundleWorker(), providerConfig, "-out=tfplan")
			if err != nil {
				t.Fatalf("an R2-bundle Worker must plan with the connection's R2 pair alone:\n%s", out)
			}

			if !s3.servedObjectTo(isolatedPlanAccessKey) {
				t.Errorf("the bundle was never read with the connection's R2 access key; requests seen: %v", s3.seen())
			}
			if got := plannedScriptContent(t, binary, workDir); got != isolatedPlanBundle {
				t.Errorf("planned script content = %q, want the bundle's bytes %q", got, isolatedPlanBundle)
			}
		})
	}
}

// isolatedPlanPreconditions gates the guard and returns the repo root and the engines to run.
func isolatedPlanPreconditions(t *testing.T) (string, []string) {
	t.Helper()
	if os.Getenv(planGuardsEnv) != "1" {
		t.Skipf("set %s=1 (with tofu on PATH and registry egress) to run the credential-isolated plan guards", planGuardsEnv)
	}
	if os.Getenv("TEST_SRCDIR") != "" {
		t.Skip("the guard reads the catalog source tree; skipped in the bazel sandbox")
	}
	if _, err := exec.LookPath("tofu"); err != nil {
		t.Fatal("tofu is not on PATH")
	}
	binaries := []string{"tofu"}
	if _, err := exec.LookPath("terraform"); err == nil {
		binaries = append(binaries, "terraform")
	}
	return isolatedPlanRepoRoot(t), binaries
}

// isolatedWorkerModule copies the module into a disposable directory: init writes into it.
func isolatedWorkerModule(t *testing.T, root string) string {
	t.Helper()
	workDir := t.TempDir()
	src := filepath.Join(root, "catalog", "cloudflare", "cloudflareworker", "iac", "tf")
	if err := fileutil.CopyDir(src, workDir); err != nil {
		t.Fatalf("copy the Worker module: %v", err)
	}
	for _, residue := range []string{".terraform", ".terraform.lock.hcl"} {
		_ = os.RemoveAll(filepath.Join(workDir, residue))
	}
	return workDir
}

// writeCloudflareProviderConfig writes the provider config a Cloudflare connection resolves to:
// the API token, plus the given extra YAML (the R2 block).
func writeCloudflareProviderConfig(t *testing.T, extra string) *stackinputproviderconfig.ProviderConfig {
	t.Helper()
	path := filepath.Join(t.TempDir(), "cloudflare-provider-config.yaml")
	content := fmt.Sprintf("authScheme: api_token\napiToken: %s\n%s", isolatedPlanToken, extra)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write provider config: %v", err)
	}
	return &stackinputproviderconfig.ProviderConfig{
		Path:     path,
		Provider: cloudresourcekind.CloudResourceProvider_cloudflare,
	}
}

// isolatedPlan renders the tfvars, derives the provider environment exactly as a deploy does,
// and runs init + plan with that environment and nothing else of the machine's.
func isolatedPlan(t *testing.T, binary, workDir string, manifestObject proto.Message,
	providerConfig *stackinputproviderconfig.ProviderConfig, planArgs ...string) (string, error) {
	t.Helper()
	if err := generators.WriteVarFile(manifestObject, filepath.Join(workDir, "terraform.tfvars")); err != nil {
		t.Fatalf("render tfvars: %v", err)
	}
	stackInputYaml, err := stackinput.BuildStackInputYaml(manifestObject, providerConfig)
	if err != nil {
		t.Fatalf("build stack input: %v", err)
	}
	providerEnv, err := GetProviderConfigEnvVars(stackInputYaml, workDir, "")
	if err != nil {
		t.Fatalf("derive provider env: %v", err)
	}

	env := append([]string{
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + t.TempDir(),
		"TF_IN_AUTOMATION=1",
	}, providerEnv...)
	if cache := os.Getenv("TF_PLUGIN_CACHE_DIR"); cache != "" {
		env = append(env, "TF_PLUGIN_CACHE_DIR="+cache)
	}
	for _, kv := range env {
		if strings.HasPrefix(kv, "AWS_") {
			t.Fatalf("the isolated environment must hold no AWS_* variable, found %s", strings.SplitN(kv, "=", 2)[0])
		}
	}

	var combined strings.Builder
	for _, args := range [][]string{
		{"init", "-input=false", "-no-color"},
		append([]string{"plan", "-input=false", "-no-color", "-refresh=false"}, planArgs...),
	} {
		cmd := exec.Command(binary, args...)
		cmd.Dir = workDir
		cmd.Env = env
		out, err := cmd.CombinedOutput()
		fmt.Fprintf(&combined, "$ %s %s\n%s\n", binary, strings.Join(args, " "), out)
		if err != nil {
			return combined.String(), err
		}
	}
	return combined.String(), nil
}

// plannedScriptContent reads the planned content of the Worker script from the saved plan.
func plannedScriptContent(t *testing.T, binary, workDir string) string {
	t.Helper()
	cmd := exec.Command(binary, "show", "-json", "tfplan")
	cmd.Dir = workDir
	cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "HOME=" + t.TempDir()}
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("%s show -json: %v", binary, err)
	}
	var plan struct {
		ResourceChanges []struct {
			Address string `json:"address"`
			Change  struct {
				After map[string]any `json:"after"`
			} `json:"change"`
		} `json:"resource_changes"`
	}
	if err := json.Unmarshal(out, &plan); err != nil {
		t.Fatalf("parse plan json: %v", err)
	}
	for _, rc := range plan.ResourceChanges {
		if rc.Address == "cloudflare_workers_script.main" {
			content, _ := rc.Change.After["content"].(string)
			return content
		}
	}
	t.Fatal("the plan carries no cloudflare_workers_script.main")
	return ""
}

// bundleWorker is the smallest Worker whose code is an R2 bundle.
func bundleWorker() *cloudflareworkerv1alpha1.CloudflareWorker {
	return &cloudflareworkerv1alpha1.CloudflareWorker{
		ApiVersion: "cloudflare.planton.dev/v1alpha1",
		Kind:       "CloudflareWorker",
		Metadata:   &shared.CloudResourceMetadata{Name: "bundled-api"},
		Spec: &cloudflareworkerv1alpha1.CloudflareWorkerSpec{
			AccountId:         isolatedPlanAccountID,
			WorkerName:        "bundled-api",
			CompatibilityDate: "2025-01-15",
			Source: &cloudflareworkerv1alpha1.CloudflareWorkerSpec_R2Bundle{
				R2Bundle: &cloudflareworkerv1alpha1.CloudflareWorkerScriptBundle{
					Bucket: &foreignkeyv1.StringValueOrRef{
						LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: isolatedPlanBucket},
					},
					Path: isolatedPlanKey,
				},
			},
		},
	}
}

// fakeR2 is a path-style S3 endpoint holding one object, served the way R2 serves a bundle an
// upload tool typed as application/javascript. It records which access key signed each request.
type fakeR2 struct {
	*httptest.Server
	mu       sync.Mutex
	requests []string
}

func newFakeR2(t *testing.T) *fakeR2 {
	t.Helper()
	f := &fakeR2{}
	objectPath := "/" + isolatedPlanBucket + "/" + isolatedPlanKey
	f.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessKey := signingAccessKey(r.Header.Get("Authorization"))
		f.mu.Lock()
		f.requests = append(f.requests, fmt.Sprintf("%s %s?%s key=%s", r.Method, r.URL.Path, r.URL.RawQuery, accessKey))
		f.mu.Unlock()

		if r.URL.Path != objectPath {
			http.Error(w, "NoSuchKey", http.StatusNotFound)
			return
		}
		if _, ok := r.URL.Query()["tagging"]; ok {
			w.Header().Set("Content-Type", "application/xml")
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><Tagging><TagSet></TagSet></Tagging>`))
			return
		}
		w.Header().Set("Content-Type", "application/javascript")
		w.Header().Set("Content-Length", fmt.Sprint(len(isolatedPlanBundle)))
		w.Header().Set("ETag", `"0123456789abcdef0123456789abcdef"`)
		w.Header().Set("Last-Modified", "Mon, 02 Jan 2006 15:04:05 GMT")
		if r.Method == http.MethodHead {
			return
		}
		_, _ = w.Write([]byte(isolatedPlanBundle))
	}))
	t.Cleanup(f.Close)
	return f
}

// servedObjectTo reports whether the object body was served to a request signed with accessKey.
func (f *fakeR2) servedObjectTo(accessKey string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	want := "GET /" + isolatedPlanBucket + "/" + isolatedPlanKey + "?"
	for _, r := range f.requests {
		if strings.HasPrefix(r, want) && strings.HasSuffix(r, " key="+accessKey) {
			return true
		}
	}
	return false
}

func (f *fakeR2) seen() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.requests...)
}

// signingAccessKey extracts the access key ID from a SigV4 Authorization header
// ("AWS4-HMAC-SHA256 Credential=<key>/<date>/<region>/s3/aws4_request, ...").
func signingAccessKey(authorization string) string {
	_, rest, ok := strings.Cut(authorization, "Credential=")
	if !ok {
		return ""
	}
	key, _, _ := strings.Cut(rest, "/")
	return key
}

// isolatedPlanRepoRoot walks up from this test file to the directory holding go.mod.
func isolatedPlanRepoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	dir := filepath.Dir(thisFile)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not locate the repo root (go.mod)")
		}
		dir = parent
	}
}

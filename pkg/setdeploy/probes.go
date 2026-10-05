package setdeploy

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/plantonhq/planton/internal/cli/version"
	"github.com/plantonhq/planton/pkg/iac/provisioner"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumibinary"
	"github.com/plantonhq/planton/pkg/iac/tofu/backendconfig"
	"github.com/plantonhq/planton/pkg/iac/tofu/tofuzip"
	"github.com/plantonhq/planton/shared/catalogkind"
	"golang.org/x/oauth2/google"
	goyaml "gopkg.in/yaml.v3"
)

// ProbeResult is one probe's honest outcome, exactly one arm set:
// Verified — the probe confirmed the fact and the line says what it proved;
// Assumed — the probe cannot verify this class and the line states the
// assumption the deploy proceeds on; Refused — the probe failed and the line
// names what, where, why, and what fixes it.
type ProbeResult struct {
	Verified string
	Assumed  string
	Refused  string
}

func verified(format string, args ...any) ProbeResult {
	return ProbeResult{Verified: fmt.Sprintf(format, args...)}
}

func assumed(format string, args ...any) ProbeResult {
	return ProbeResult{Assumed: fmt.Sprintf(format, args...)}
}

func refused(format string, args ...any) ProbeResult {
	return ProbeResult{Refused: fmt.Sprintf(format, args...)}
}

// Probes is the wall's window on the world. Every check that touches the
// machine or the network goes through this interface so the wall's logic —
// which checks run for which nodes, how outcomes aggregate into the report —
// is exhaustively testable with fakes, and so the platform CLI can swap probe
// implementations without forking the wall.
type Probes interface {
	// HclBinary verifies the tofu/terraform binary is executable on PATH.
	HclBinary(binaryName string) ProbeResult

	// PulumiBinary verifies the pulumi CLI is executable on PATH.
	PulumiBinary() ProbeResult

	// ModulePublished verifies a catalog module exists for the kind: the kind
	// registry answers structurally (an unknown kind is a refusal), and the
	// published artifact is HEAD-probed when a release version is pinned (a
	// missing artifact is a warning — the engine falls back to a source
	// checkout — surfaced here so version skew is loud instead of a silently
	// slower deploy).
	ModulePublished(kindName string, prov provisioner.ProvisionerType, moduleVersion string) ProbeResult

	// TofuBackend verifies the state backend is reachable with the
	// credentials in hand (bucket-level probes for s3/S3-compatible and gcs;
	// azurerm carries no account identity in its config, so it degrades to a
	// stated assumption).
	TofuBackend(cfg *backendconfig.TofuBackendConfig) ProbeResult

	// PulumiBackend verifies the pulumi backend URL's store is reachable for
	// object-store URLs; service backends (Pulumi Cloud) and ambient logins
	// degrade to stated assumptions.
	PulumiBackend(url string) ProbeResult

	// ProviderCredentials verifies the ambient credentials for one provider
	// actually authenticate (a live identity call, not a file check).
	// Providers without a probe degrade to stated assumptions.
	ProviderCredentials(provider catalogkind.CatalogProvider) ProbeResult

	// KubeContext verifies a named kubectl context exists in the local
	// kubeconfig.
	KubeContext(name string) ProbeResult
}

// LiveProbes is the production Probes implementation: real PATH lookups, real
// bucket-level reachability calls, real credential identity calls. Every
// probe is bounded by a short timeout — a wall that hangs is worse than a
// wall that states an assumption.
type LiveProbes struct{}

const probeTimeout = 15 * time.Second

func (LiveProbes) HclBinary(binaryName string) ProbeResult {
	if _, err := exec.LookPath(binaryName); err != nil {
		return refused("%s is not on PATH — install it (https://opentofu.org/docs/intro/install) or switch the manifest's provisioner", binaryName)
	}
	return verified("%s binary on PATH", binaryName)
}

func (LiveProbes) PulumiBinary() ProbeResult {
	if _, err := exec.LookPath("pulumi"); err != nil {
		return refused("pulumi is not on PATH — install it (https://www.pulumi.com/docs/install) or switch the manifest's provisioner")
	}
	return verified("pulumi binary on PATH")
}

func (LiveProbes) ModulePublished(kindName string, prov provisioner.ProvisionerType, moduleVersion string) ProbeResult {
	// The same version resolution the module resolvers apply: an explicit
	// --module-version wins, else the CLI's own release version; a dev build
	// has no pinned artifact and resolves from a source checkout.
	targetVersion := moduleVersion
	if targetVersion == "" && version.Version != "" && version.Version != version.DefaultVersion {
		targetVersion = version.Version
	}

	var url string
	var err error
	switch prov {
	case provisioner.ProvisionerTypePulumi:
		url, err = pulumibinary.BuildDownloadURL(kindName, targetVersion)
	default:
		url, err = tofuzip.BuildDownloadURL(kindName, targetVersion)
	}
	if err != nil {
		// The registry does not know the kind — no module can exist for it.
		return refused("no %s module exists for kind %s in the catalog — check the kind name, or upgrade the CLI (`planton upgrade`)", prov, kindName)
	}

	if targetVersion == "" {
		return assumed("kind %s resolves in the module catalog; dev build — the module comes from a source checkout, not a published artifact", kindName)
	}

	client := &http.Client{Timeout: probeTimeout}
	resp, headErr := client.Head(url)
	if headErr != nil {
		return assumed("kind %s resolves in the module catalog; artifact availability not verified (%v)", kindName, headErr)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		// Not a refusal: the engine falls back to a source checkout — slower,
		// not broken. The Assumed arm carries it; the wall renders module
		// assumptions as warnings so version skew is loud.
		return assumed("module artifact for %s at %s is not published (HTTP %d) — the deploy falls back to a source checkout, which is slower", kindName, targetVersion, resp.StatusCode)
	}
	return verified("%s module for %s published at %s", prov, kindName, targetVersion)
}

func (LiveProbes) TofuBackend(cfg *backendconfig.TofuBackendConfig) ProbeResult {
	ctx, cancel := context.WithTimeout(context.Background(), probeTimeout)
	defer cancel()

	switch cfg.BackendType {
	case "s3":
		return probeS3Bucket(ctx, cfg.BackendBucket, cfg.BackendRegion, cfg.BackendEndpoint)
	case "gcs":
		return probeGcsBucket(ctx, cfg.BackendBucket)
	case "azurerm":
		// The azurerm backend addresses its storage account through raw
		// backend-config arguments this config does not carry, so there is
		// nothing addressable to probe from here.
		return assumed("azurerm state container %q not probed (the storage account is addressed at init) — `tofu init` verifies it at handoff", cfg.BackendBucket)
	case "remote":
		return assumed("remote (TFE-protocol) backend not probed — `tofu init` verifies it at handoff")
	default:
		return assumed("backend type %q has no reachability probe — `tofu init` verifies it at handoff", cfg.BackendType)
	}
}

func probeS3Bucket(ctx context.Context, bucket, region, endpoint string) ProbeResult {
	loadOpts := []func(*awsconfig.LoadOptions) error{}
	if region != "" && region != "auto" {
		loadOpts = append(loadOpts, awsconfig.WithRegion(region))
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, loadOpts...)
	if err != nil {
		return refused("state bucket %q not verified: no usable AWS credentials in the environment (%v) — configure credentials (or the runner's OIDC step) and retry", bucket, err)
	}
	client := awss3.NewFromConfig(awsCfg, func(o *awss3.Options) {
		if endpoint != "" {
			o.BaseEndpoint = &endpoint
			// S3-compatible stores (R2, MinIO) speak path-style addressing.
			o.UsePathStyle = true
		}
		if region == "auto" {
			o.Region = "auto"
		}
	})
	if _, err := client.HeadBucket(ctx, &awss3.HeadBucketInput{Bucket: &bucket}); err != nil {
		return refused("state bucket %q is not reachable with the credentials in hand (%v) — create the bucket, fix its name, or fix the credentials", bucket, err)
	}
	return verified("s3 state bucket %q reachable", bucket)
}

func probeGcsBucket(ctx context.Context, bucket string) ProbeResult {
	creds, err := google.FindDefaultCredentials(ctx, "https://www.googleapis.com/auth/devstorage.read_only")
	if err != nil {
		return refused("state bucket %q not verified: no usable Google credentials in the environment (%v) — configure ADC (or the runner's OIDC step) and retry", bucket, err)
	}
	token, err := creds.TokenSource.Token()
	if err != nil {
		return refused("state bucket %q not verified: Google credentials present but no token could be minted (%v)", bucket, err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://storage.googleapis.com/storage/v1/b/"+bucket+"?fields=name", nil)
	if err != nil {
		return assumed("gcs state bucket %q not probed (%v) — `tofu init` verifies it at handoff", bucket, err)
	}
	token.SetAuthHeader(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return assumed("gcs state bucket %q not probed (%v) — `tofu init` verifies it at handoff", bucket, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return refused("gcs state bucket %q is not reachable with the credentials in hand (HTTP %d) — create the bucket, fix its name, or fix the credentials", bucket, resp.StatusCode)
	}
	return verified("gcs state bucket %q reachable", bucket)
}

func (p LiveProbes) PulumiBackend(url string) ProbeResult {
	ctx, cancel := context.WithTimeout(context.Background(), probeTimeout)
	defer cancel()

	switch {
	case url == "":
		return assumed("no pulumi backend URL pinned — the machine's ambient `pulumi login` decides where state lives")
	case strings.HasPrefix(url, "s3://"):
		bucket := bucketFromURL(url, "s3://")
		return probeS3Bucket(ctx, bucket, "", "")
	case strings.HasPrefix(url, "gs://"):
		return probeGcsBucket(ctx, bucketFromURL(url, "gs://"))
	case strings.HasPrefix(url, "file://"):
		dir := strings.TrimPrefix(url, "file://")
		if dir == "" || dir == "~" {
			return assumed("pulumi file backend %q — pulumi resolves it at handoff", url)
		}
		if info, err := os.Stat(expandHome(dir)); err != nil || !info.IsDir() {
			return refused("pulumi file backend %q does not exist or is not a directory — create it or fix the URL", url)
		}
		return verified("pulumi file backend %q exists", url)
	default:
		return assumed("pulumi backend %q not probed (service backends authenticate at handoff)", url)
	}
}

// bucketFromURL extracts the bucket segment of an object-store backend URL
// ("s3://bucket/prefix" -> "bucket").
func bucketFromURL(url, scheme string) string {
	rest := strings.TrimPrefix(url, scheme)
	if i := strings.IndexByte(rest, '/'); i >= 0 {
		return rest[:i]
	}
	return rest
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

func (LiveProbes) ProviderCredentials(provider catalogkind.CatalogProvider) ProbeResult {
	ctx, cancel := context.WithTimeout(context.Background(), probeTimeout)
	defer cancel()

	switch provider {
	case catalogkind.CatalogProvider_aws:
		awsCfg, err := awsconfig.LoadDefaultConfig(ctx)
		if err != nil {
			return refused("aws credentials: none usable in the environment (%v) — configure credentials (or the runner's OIDC step)", err)
		}
		out, err := sts.NewFromConfig(awsCfg).GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
		if err != nil {
			return refused("aws credentials do not authenticate (%v) — fix the credentials (or the runner's OIDC step)", err)
		}
		return verified("aws credentials authenticate (account %s)", safeDeref(out.Account))
	case catalogkind.CatalogProvider_gcp:
		creds, err := google.FindDefaultCredentials(ctx, "https://www.googleapis.com/auth/cloud-platform")
		if err != nil {
			return refused("gcp credentials: no application default credentials found (%v) — run `gcloud auth application-default login` (or the runner's OIDC step)", err)
		}
		if _, err := creds.TokenSource.Token(); err != nil {
			return refused("gcp credentials do not authenticate (%v)", err)
		}
		return verified("gcp credentials authenticate")
	case catalogkind.CatalogProvider_azure:
		cred, err := azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			return refused("azure credentials: no usable credential chain (%v) — run `az login` (or the runner's OIDC step)", err)
		}
		if _, err := cred.GetToken(ctx, policy.TokenRequestOptions{Scopes: []string{"https://management.azure.com/.default"}}); err != nil {
			return refused("azure credentials do not authenticate (%v) — run `az login` (or the runner's OIDC step)", err)
		}
		return verified("azure credentials authenticate")
	case catalogkind.CatalogProvider_kubernetes:
		// Cluster reachability is per kube context, verified by KubeContext;
		// there is no provider-level ambient identity to probe.
		return assumed("kubernetes access is verified per kube context")
	default:
		return assumed("%s credentials have no preflight probe — the provider verifies them at handoff", provider)
	}
}

func safeDeref(s *string) string {
	if s == nil {
		return "unknown"
	}
	return *s
}

// KubeContext verifies the named context exists in the local kubeconfig. The
// read is deliberately minimal — context names only — so no kubernetes client
// dependency enters this package for what is a presence check.
func (LiveProbes) KubeContext(name string) ProbeResult {
	kubeconfigPath := os.Getenv("KUBECONFIG")
	if kubeconfigPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return assumed("kube context %q not probed (no home directory resolved)", name)
		}
		kubeconfigPath = filepath.Join(home, ".kube", "config")
	}
	raw, err := os.ReadFile(kubeconfigPath)
	if err != nil {
		return refused("kube context %q not verified: kubeconfig %s is not readable (%v)", name, kubeconfigPath, err)
	}
	var kubeconfig struct {
		Contexts []struct {
			Name string `yaml:"name"`
		} `yaml:"contexts"`
	}
	if err := goyaml.Unmarshal(raw, &kubeconfig); err != nil {
		return refused("kube context %q not verified: kubeconfig %s does not parse (%v)", name, kubeconfigPath, err)
	}
	for _, c := range kubeconfig.Contexts {
		if c.Name == name {
			return verified("kube context %q exists in %s", name, kubeconfigPath)
		}
	}
	return refused("kube context %q does not exist in %s — check `kubectl config get-contexts`", name, kubeconfigPath)
}

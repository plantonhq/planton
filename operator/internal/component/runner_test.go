package component

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	v1 "github.com/plantonhq/planton/operator/api/v1"
	"github.com/plantonhq/planton/operator/internal/resources"
)

// wantTofu is the zero-config default provisioner the seeds and the binding
// must both land on.
const wantTofu = "tofu"

// The runner is deployed by default: an install that cannot deploy
// infrastructure is a browsing UI. Opting out is the explicit act.
func TestRunner_IsEnabledByDefault(t *testing.T) {
	r := &Runner{}
	if !r.IsEnabled(ingressPlatform(false)) {
		t.Error("runner must be enabled with no spec.runner at all")
	}

	p := ingressPlatform(false)
	p.Spec.Runner = &v1.RunnerSpec{}
	if !r.IsEnabled(p) {
		t.Error("runner must be enabled with an empty spec.runner")
	}

	off := false
	p.Spec.Runner = &v1.RunnerSpec{Enabled: &off}
	if r.IsEnabled(p) {
		t.Error("explicit enabled=false must disable the runner")
	}
}

// The controlplane dependency (not just temporal) makes first boot clean: by
// the time the runner pod starts, its registration, credential hash, and
// deploy defaults have been seeded.
func TestRunner_Dependencies(t *testing.T) {
	deps := (&Runner{}).Dependencies(ingressPlatform(false))
	want := map[string]bool{"temporal": true, "controlplane": true}
	if len(deps) != len(want) {
		t.Fatalf("deps = %v, want temporal + controlplane", deps)
	}
	for _, dep := range deps {
		if !want[dep] {
			t.Errorf("unexpected dependency %s", dep)
		}
	}
}

func TestRunnerConfig_Defaults(t *testing.T) {
	p := ingressPlatform(false)
	cfg := runnerConfig(p, nil)

	if cfg.OrgSlug != "default" {
		t.Errorf("orgSlug = %s, want the bootstrap default", cfg.OrgSlug)
	}
	if cfg.Version != "v1.0.0" {
		t.Errorf("version = %s, want the platform version", cfg.Version)
	}
	if cfg.StorageSize.String() != resources.RunnerDefaultStorageSize {
		t.Errorf("storage = %s, want the resolved default %s (resolution lives here so spec.storage participates)",
			cfg.StorageSize.String(), resources.RunnerDefaultStorageSize)
	}
	if cfg.StorageClassName != "" {
		t.Errorf("storageClassName = %q, want empty when nothing pins a class", cfg.StorageClassName)
	}
}

// The platform-wide spec.storage block reaches the runner volume like every
// other volume; the runner's own fields still win.
func TestRunnerConfig_GlobalStorage(t *testing.T) {
	p := ingressPlatform(false)
	p.Spec.Storage = &v1.StorageSpec{
		StorageClassName: "trident",
		Size:             resource.MustParse("800Gi"),
	}
	cfg := runnerConfig(p, nil)
	if cfg.StorageSize.String() != "800Gi" || cfg.StorageClassName != "trident" {
		t.Errorf("global storage must reach the runner: got %s/%q, want 800Gi/trident",
			cfg.StorageSize.String(), cfg.StorageClassName)
	}

	p.Spec.Runner = &v1.RunnerSpec{
		StorageSize:      resource.MustParse("50Gi"),
		StorageClassName: "fast-ssd",
	}
	cfg = runnerConfig(p, nil)
	if cfg.StorageSize.String() != "50Gi" || cfg.StorageClassName != "fast-ssd" {
		t.Errorf("runner fields must beat the global block: got %s/%q, want 50Gi/fast-ssd",
			cfg.StorageSize.String(), cfg.StorageClassName)
	}
}

// The platform's registry root moves the runner and the control plane together;
// the runner's own image override still wins over it.
func TestImageRegistry_MovesEveryPlantonImage(t *testing.T) {
	const mirror = "asia-south1-docker.pkg.dev/plantonhq/planton"
	p := ingressPlatform(false)
	p.Spec.ImageRegistry = mirror

	if got := runnerConfig(p, nil).ImageRepository; got != mirror+"/runner" {
		t.Errorf("runner image = %s, want the mirror root", got)
	}
	if got := (&ControlPlane{}).buildConfig(p, nil).ImageRepository; got != mirror+"/control-plane" {
		t.Errorf("control-plane image = %s, want the mirror root", got)
	}

	p.Spec.Runner = &v1.RunnerSpec{Image: &v1.ImageSpec{Repository: "example.com/runner"}}
	if got := runnerConfig(p, nil).ImageRepository; got != "example.com/runner" {
		t.Errorf("runner image = %s, want the declared override over the root", got)
	}
}

func TestRunnerConfig_SpecWins(t *testing.T) {
	p := ingressPlatform(false)
	p.Spec.Bootstrap = &v1.BootstrapSpec{
		Organization: &v1.BootstrapOrganizationSpec{Slug: "acme"},
	}
	p.Spec.Runner = &v1.RunnerSpec{
		Image:                      &v1.ImageSpec{Repository: "example.com/runner", Tag: "custom"},
		StorageSize:                resource.MustParse("10Gi"),
		ServiceAccountAnnotations:  map[string]string{"eks.amazonaws.com/role-arn": "arn:aws:iam::1:role/r"},
		CloudCredentialsSecretName: "my-keys",
	}
	cfg := runnerConfig(p, nil)

	if cfg.OrgSlug != "acme" {
		t.Errorf("orgSlug = %s, want acme (follows the bootstrap org)", cfg.OrgSlug)
	}
	if cfg.ImageRepository != "example.com/runner" || cfg.ImageTag != "custom" {
		t.Errorf("image = %s:%s, want the declared override", cfg.ImageRepository, cfg.ImageTag)
	}
	if cfg.StorageSize.String() != "10Gi" {
		t.Errorf("storage = %s, want 10Gi", cfg.StorageSize.String())
	}
	if cfg.CloudCredentialsSecretName != "my-keys" {
		t.Errorf("cloudCredentialsSecretName = %s, want my-keys", cfg.CloudCredentialsSecretName)
	}
}

// Defaulting lives in code as well as the CRD marker: a nil spec.bootstrap
// gets no marker defaults.
func TestEffectiveIacProvisioner(t *testing.T) {
	p := ingressPlatform(false)
	if got := effectiveIacProvisioner(p); got != wantTofu {
		t.Errorf("provisioner = %s, want tofu with no bootstrap spec", got)
	}

	p.Spec.Bootstrap = &v1.BootstrapSpec{}
	if got := effectiveIacProvisioner(p); got != wantTofu {
		t.Errorf("provisioner = %s, want tofu with an empty bootstrap spec", got)
	}

	p.Spec.Bootstrap.IacProvisioner = "terraform"
	if got := effectiveIacProvisioner(p); got != "terraform" {
		t.Errorf("provisioner = %s, want the declared terraform", got)
	}
}

// The controlplane config carries the runner binding exactly when the runner
// is enabled -- the Java seeders activate on the slug property's presence, so
// a disabled runner must leave the binding nil.
func TestBuildConfig_RunnerBindingFollowsToggle(t *testing.T) {
	cp := &ControlPlane{}

	cfg := cp.buildConfig(ingressPlatform(true), nil)
	if cfg.Runner == nil {
		t.Fatal("expected a runner binding by default")
	}
	if cfg.Runner.CloudOpsSecretName == "" || cfg.Runner.Provisioner != wantTofu {
		t.Errorf("binding = %+v, want the CloudOps Secret name and tofu", cfg.Runner)
	}

	p := ingressPlatform(true)
	off := false
	p.Spec.Runner = &v1.RunnerSpec{Enabled: &off}
	if cfg := cp.buildConfig(p, nil); cfg.Runner != nil {
		t.Error("a disabled runner must leave the binding nil (seeds stay inert)")
	}
}

// Substrate detection changes only the GUIDANCE we surface, so it needs to be
// right on exactly the axis that changes guidance: EKS (IRSA) vs AWS-not-EKS
// (instance profile) vs the other managed clouds vs everything else.
func TestDetectSubstrate(t *testing.T) {
	node := func(providerID, kubelet string, labels map[string]string) corev1.Node {
		n := corev1.Node{}
		n.Spec.ProviderID = providerID
		n.Status.NodeInfo.KubeletVersion = kubelet
		n.Labels = labels
		return n
	}

	cases := []struct {
		name  string
		nodes []corev1.Node
		want  substrate
	}{
		{"eks by kubelet suffix",
			[]corev1.Node{node("aws:///us-west-2a/i-0abc", "v1.31.0-eks-a737599", nil)},
			substrateEKS},
		{"eks by nodegroup label",
			[]corev1.Node{node("aws:///us-west-2a/i-0abc", "v1.31.0",
				map[string]string{"eks.amazonaws.com/nodegroup": "ng-1"})},
			substrateEKS},
		{"aws without eks markers (kOps, RKE2-on-EC2)",
			[]corev1.Node{node("aws:///us-west-2a/i-0abc", "v1.31.0", nil)},
			substrateAWS},
		{"gke",
			[]corev1.Node{node("gce://proj/us-central1-a/node-1", "v1.31.1-gke.1846000", nil)},
			substrateGKE},
		{"aks",
			[]corev1.Node{node("azure:///subscriptions/x/vm-1", "v1.31.0",
				map[string]string{"kubernetes.azure.com/agentpool": "pool1"})},
			substrateAKS},
		{"kind / bare metal",
			[]corev1.Node{node("kind://docker/kind/kind-control-plane", "v1.31.0", nil)},
			substrateUnknown},
		{"no nodes", nil, substrateUnknown},
	}

	for _, tc := range cases {
		if got := detectSubstrate(tc.nodes); got != tc.want {
			t.Errorf("%s: substrate = %s, want %s", tc.name, got, tc.want)
		}
	}
}

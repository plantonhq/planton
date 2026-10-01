package artifactslot

import (
	"strings"
	"testing"

	awsecstaskdefinitionv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsecstaskdefinition/v1alpha1"
	gcpcloudrunv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcloudrun/v1alpha1"
	"github.com/plantonhq/planton/catalog/kubernetes"
	kubernetesdeploymentv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesdeployment/v1alpha1"
	sharedoptions "github.com/plantonhq/planton/shared/options"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// The blank-fill law on a repeated slot: a blank image receives the
// reference, an authored image (a sidecar) is untouched, and the
// injections report names exactly what was written.
func TestInject_CloudRunBlankFill(t *testing.T) {
	manifest := &gcpcloudrunv1alpha1.GcpCloudRun{
		Spec: &gcpcloudrunv1alpha1.GcpCloudRunSpec{
			Containers: []*gcpcloudrunv1alpha1.GcpCloudRunContainer{
				{Name: "app"},
				{Name: "collector", Image: "otel/collector:0.99"},
			},
		},
	}
	injections, err := Inject(manifest, "ghcr.io/acme/storefront:abc123", "")
	if err != nil {
		t.Fatalf("inject: %v", err)
	}
	if len(injections) != 1 || injections[0].FieldPath != "spec.containers[0].image" {
		t.Fatalf("expected exactly the blank container injected, got %+v", injections)
	}
	if got := manifest.Spec.Containers[0].Image; got != "ghcr.io/acme/storefront:abc123" {
		t.Fatalf("blank container image = %q", got)
	}
	if got := manifest.Spec.Containers[1].Image; got != "otel/collector:0.99" {
		t.Fatalf("authored sidecar image must be untouched, got %q", got)
	}
}

// A fully-authored container list takes zero injections — the caller
// hears that honestly instead of a silent overwrite.
func TestInject_CloudRunAllAuthoredInjectsNothing(t *testing.T) {
	manifest := &gcpcloudrunv1alpha1.GcpCloudRun{
		Spec: &gcpcloudrunv1alpha1.GcpCloudRunSpec{
			Containers: []*gcpcloudrunv1alpha1.GcpCloudRunContainer{
				{Name: "app", Image: "authored:v1"},
			},
		},
	}
	injections, err := Inject(manifest, "ghcr.io/acme/x:1", "")
	if err != nil || len(injections) != 0 {
		t.Fatalf("expected no injections, got %v / %v", injections, err)
	}
}

func TestInject_EcsTaskDefinitionBlankFill(t *testing.T) {
	manifest := &awsecstaskdefinitionv1alpha1.AwsEcsTaskDefinition{
		Spec: &awsecstaskdefinitionv1alpha1.AwsEcsTaskDefinitionSpec{
			Containers: []*awsecstaskdefinitionv1alpha1.AwsEcsTaskDefinitionContainer{
				{Name: "api"},
			},
		},
	}
	injections, err := Inject(manifest, "123.dkr.ecr.us-west-2.amazonaws.com/api:1.4.2", "")
	if err != nil || len(injections) != 1 {
		t.Fatalf("expected one injection, got %v / %v", injections, err)
	}
	if got := manifest.Spec.Containers[0].Image; got != "123.dkr.ecr.us-west-2.amazonaws.com/api:1.4.2" {
		t.Fatalf("image = %q", got)
	}
}

// The Kubernetes shape: unconditional injection (the kind models one app
// container), the reference SPLIT into repo+tag, and the version slot
// stamped from the sanitized branch.
func TestInject_KubernetesDeploymentSplitAndVersion(t *testing.T) {
	manifest := &kubernetesdeploymentv1alpha1.KubernetesDeployment{
		Spec: &kubernetesdeploymentv1alpha1.KubernetesDeploymentSpec{
			Container: &kubernetesdeploymentv1alpha1.KubernetesDeploymentContainer{
				App: &kubernetes.WorkloadContainer{
					Image: &kubernetes.WorkloadContainerImage{Repo: "authored", Tag: "old", Digest: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
				},
			},
		},
	}
	injections, err := Inject(manifest, "ghcr.io/acme/storefront:abc123", "feature/PR_42")
	if err != nil {
		t.Fatalf("inject: %v", err)
	}
	if len(injections) != 2 {
		t.Fatalf("expected image + version injections, got %+v", injections)
	}
	image := manifest.Spec.Container.App.Image
	if image.Repo != "ghcr.io/acme/storefront" || image.Tag != "abc123" {
		t.Fatalf("split = %q / %q", image.Repo, image.Tag)
	}
	if got := manifest.Spec.GetVersion(); got != "feature-pr-42" {
		t.Fatalf("version = %q", got)
	}
}

// A digest reference writes its digest into the digest leaf, never into the
// tag: "repo:tag@sha256:…" keeps the tag as the build's readable name and pins
// the digest, and a bare "repo@sha256:…" leaves the tag empty. A reference
// without a digest clears an authored one, so the new build is what runs.
func TestInject_KubernetesDigestReference(t *testing.T) {
	const digest = "sha256:b5e2a1c0d9f8e7a6b5c4d3e2f1a0b9c8d7e6f5a4b3c2d1e0f9a8b7c6d5e4f3a2"
	for reference, want := range map[string]*kubernetes.WorkloadContainerImage{
		"ghcr.io/acme/x:a743940@" + digest:       {Repo: "ghcr.io/acme/x", Tag: "a743940", Digest: digest},
		"ghcr.io/acme/x@" + digest:               {Repo: "ghcr.io/acme/x", Digest: digest},
		"registry:5000/acme/x:a743940@" + digest: {Repo: "registry:5000/acme/x", Tag: "a743940", Digest: digest},
		"ghcr.io/acme/x:a743940":                 {Repo: "ghcr.io/acme/x", Tag: "a743940"},
	} {
		manifest := &kubernetesdeploymentv1alpha1.KubernetesDeployment{
			Spec: &kubernetesdeploymentv1alpha1.KubernetesDeploymentSpec{
				Container: &kubernetesdeploymentv1alpha1.KubernetesDeploymentContainer{
					App: &kubernetes.WorkloadContainer{Image: &kubernetes.WorkloadContainerImage{Digest: "sha256:" + strings.Repeat("0", 64)}},
				},
			},
		}
		if _, err := Inject(manifest, reference, ""); err != nil {
			t.Fatalf("inject %s: %v", reference, err)
		}
		got := manifest.Spec.Container.App.Image
		if got.Repo != want.Repo || got.Tag != want.Tag || got.Digest != want.Digest {
			t.Errorf("%s: split = %q / %q / %q, want %q / %q / %q", reference, got.Repo, got.Tag, got.Digest, want.Repo, want.Tag, want.Digest)
		}
		if manifest.Spec.GetVersion() != "" {
			t.Fatal("no branch means no version stamp")
		}
	}
}

func TestParseReference_ReadsTagAndDigest(t *testing.T) {
	const digest = "sha256:" + "c0ffee"
	for reference, want := range map[string]Reference{
		"ghcr.io/acme/x:a743940@" + digest: {Repo: "ghcr.io/acme/x", Tag: "a743940", Digest: digest},
		"ghcr.io/acme/x@" + digest:         {Repo: "ghcr.io/acme/x", Digest: digest},
		"localhost:5000/x":                 {Repo: "localhost:5000/x"},
		"localhost:5000/x:1":               {Repo: "localhost:5000/x", Tag: "1"},
	} {
		got := ParseReference(reference)
		if got.Repo != want.Repo || got.Tag != want.Tag || got.Digest != want.Digest || got.Raw != reference {
			t.Errorf("%s: got %+v", reference, got)
		}
	}
}

// A kind with no annotated slot injects nothing and errors nothing — the
// caller owns the refusal.
func TestInject_NoSlotKindIsHonestNoOp(t *testing.T) {
	manifest := &gcpcloudrunv1alpha1.GcpCloudRun{} // nil spec: nothing to walk into
	injections, err := Inject(manifest, "ghcr.io/x:1", "")
	if err != nil {
		t.Fatalf("inject: %v", err)
	}
	// With a nil spec the walk still visits the annotated field over an
	// empty list — zero injections, honestly reported.
	if len(injections) != 0 {
		t.Fatalf("expected none, got %+v", injections)
	}
}

func TestParseReference(t *testing.T) {
	cases := []struct{ in, repo, tag, digest string }{
		{"ghcr.io/acme/app:1.2.3", "ghcr.io/acme/app", "1.2.3", ""},
		{"ghcr.io/acme/app@sha256:abc", "ghcr.io/acme/app", "", "sha256:abc"},
		{"localhost:5000/app:v1", "localhost:5000/app", "v1", ""},
		{"localhost:5000/app", "localhost:5000/app", "", ""},
		{"nginx", "nginx", "", ""},
	}
	for _, tc := range cases {
		got := ParseReference(tc.in)
		if got.Repo != tc.repo || got.Tag != tc.tag || got.Digest != tc.digest {
			t.Errorf("%q -> %q/%q/%q, want %q/%q/%q", tc.in, got.Repo, got.Tag, got.Digest, tc.repo, tc.tag, tc.digest)
		}
	}
}

// The conformance pin: the artifact-receiver kinds carry their slot
// annotations with exactly these shapes. The hosted control plane's
// injector registry is held to the same set by its own agreement test, so
// a change on either side that forgets the other fails a build. (A
// catalog-wide every-workload-kind ratchet is deliberately not built for
// three kinds; adopt the secret-coverage gate pattern when the receiver
// set grows.)
func TestReceiverKindsCarryTheirSlots(t *testing.T) {
	slot := func(m interface {
		ProtoReflect() protoreflect.Message
	}, field string) string {
		fd := m.ProtoReflect().Descriptor().Fields().ByName("spec")
		specField := fd.Message().Fields().ByName(protoreflect.Name(field))
		if specField == nil {
			t.Fatalf("no spec.%s field", field)
		}
		subpath, _ := proto.GetExtension(specField.Options(), sharedoptions.E_ArtifactImageSlot).(string)
		return subpath
	}
	if got := slot(&gcpcloudrunv1alpha1.GcpCloudRun{}, "containers"); got != "image" {
		t.Fatalf("GcpCloudRun slot = %q", got)
	}
	if got := slot(&awsecstaskdefinitionv1alpha1.AwsEcsTaskDefinition{}, "containers"); got != "image" {
		t.Fatalf("AwsEcsTaskDefinition slot = %q", got)
	}
	if got := slot(&kubernetesdeploymentv1alpha1.KubernetesDeployment{}, "container"); got != "app.image" {
		t.Fatalf("KubernetesDeployment slot = %q", got)
	}
	versionField := (&kubernetesdeploymentv1alpha1.KubernetesDeployment{}).ProtoReflect().Descriptor().
		Fields().ByName("spec").Message().Fields().ByName("version")
	if isSlot, _ := proto.GetExtension(versionField.Options(), sharedoptions.E_ArtifactVersionSlot).(bool); !isSlot {
		t.Fatal("KubernetesDeployment spec.version must carry the version slot")
	}
}

func TestSanitizeVersion(t *testing.T) {
	cases := []struct{ in, want string }{
		{"feature/PR_42", "feature-pr-42"},
		{"main", "main"},
		{"-weird--Branch-", "weird-branch"},
		{"a-very-long-branch-name-that-exceeds-the-ceiling", "a-very-long-branch-name-that-e"},
	}
	for _, tc := range cases {
		if got := SanitizeVersion(tc.in); got != tc.want {
			t.Errorf("%q -> %q, want %q", tc.in, got, tc.want)
		}
	}
}

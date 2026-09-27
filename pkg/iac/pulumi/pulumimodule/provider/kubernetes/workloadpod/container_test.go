package workloadpod

import (
	"testing"

	kubernetesv1 "github.com/plantonhq/planton/catalog/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// A digest pins the exact build: Kubernetes pulls repo:tag@digest by the
// digest and keeps the tag as the readable name, so re-pushing the tag never
// changes what the container runs.
func TestBuildContainer_TheImageReferenceCarriesTheDigest(t *testing.T) {
	const digest = "sha256:b5e2a1c0d9f8e7a6b5c4d3e2f1a0b9c8d7e6f5a4b3c2d1e0f9a8b7c6d5e4f3a2"
	for want, image := range map[string]*kubernetesv1.WorkloadContainerImage{
		"ghcr.io/acme/checkout:a743940":           {Repo: "ghcr.io/acme/checkout", Tag: "a743940"},
		"ghcr.io/acme/checkout:a743940@" + digest: {Repo: "ghcr.io/acme/checkout", Tag: "a743940", Digest: digest},
		"ghcr.io/acme/checkout@" + digest:         {Repo: "ghcr.io/acme/checkout", Digest: digest},
	} {
		args := BuildContainer(&kubernetesv1.WorkloadContainer{Image: image}, "app", "")
		if got := string(args.Image.(pulumi.String)); got != want {
			t.Errorf("image %+v: got %q, want %q", image, got, want)
		}
	}
}

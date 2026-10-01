package module

import (
	"reflect"
	"sort"
	"strings"
	"testing"

	kubernetes "github.com/plantonhq/planton/catalog/kubernetes"
	kubernetestektonoperatorv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetestektonoperator/v1alpha1"
	"gopkg.in/yaml.v3"
)

// The image_registry contract, locked here because nothing else reads the
// rendered Deployments for it offline: empty renders the release manifest
// untouched; set, every image Tekton publishes moves to the registry at the
// same path and digest (the manifest's own and, through the lifecycle
// container's IMAGE_* variables, every component's), images on other hosts
// stay, an explicit image override wins, and the table belongs to the
// pinned operator release. The Terraform module renders the same env from
// the same table text (iac/tf/images.tf, parity-guarded).

const mirror = "mirror.example.com/ghcr"

// The two Deployments in the shape the pinned release manifest ships them:
// the operator's two containers share one image and only the lifecycle
// container carries IMAGE_* variables.
const operatorDeployment = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: tekton-operator
  namespace: tekton-operator
spec:
  template:
    spec:
      containers:
        - name: tekton-operator-lifecycle
          image: ghcr.io/tektoncd/operator/operator-303303c315a48490ba6517859ef65b77:v0.80.0@sha256:49f9f258920f77ca9db8031f8ee93b6abef0ff6a1a23254d937fd347601f4853
          env:
            - name: SYSTEM_NAMESPACE
              value: tekton-operator
            - name: IMAGE_PIPELINES_PROXY
              value: ghcr.io/tektoncd/operator/proxy-webhook-f6167da7bc41b96a27c5529f850e63d1:v0.80.0@sha256:2260f9bdf1699253a47e5d2664f70465f1a8c59ad7651cede2469a7586774341
            - name: IMAGE_JOB_PRUNER_TKN
              value: ghcr.io/tektoncd/plumbing/tkn@sha256:233de6c8b8583a34c2379fa98d42dba739146c9336e8d41b66030484357481ed
        - name: tekton-operator-cluster-operations
          image: ghcr.io/tektoncd/operator/operator-303303c315a48490ba6517859ef65b77:v0.80.0@sha256:49f9f258920f77ca9db8031f8ee93b6abef0ff6a1a23254d937fd347601f4853
          env:
            - name: PROFILING_PORT
              value: "9009"
`

const webhookDeployment = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: tekton-operator-webhook
  namespace: tekton-operator
spec:
  template:
    spec:
      containers:
        - name: tekton-operator-webhook
          image: ghcr.io/tektoncd/operator/webhook-f2bb711aa8f0c0892856a4cbf6d9ddd8:v0.80.0@sha256:e4ec282861616a364dd8ad02cadec0fdabe26087605852ec85de3b7c084e5c50
`

func document(t *testing.T, text string) map[string]interface{} {
	t.Helper()
	state := map[string]interface{}{}
	if err := yaml.Unmarshal([]byte(text), &state); err != nil {
		t.Fatalf("fixture does not parse: %v", err)
	}
	return state
}

// render applies deploymentTransformation the way Resources does: the table
// is loaded only when image_registry is set.
func render(t *testing.T, spec *kubernetestektonoperatorv1alpha1.KubernetesTektonOperatorSpec, text string) map[string]interface{} {
	t.Helper()
	var images *tektonImages
	if spec.GetImageRegistry() != "" {
		loaded, err := loadImageTable()
		if err != nil {
			t.Fatalf("loadImageTable: %v", err)
		}
		images = loaded
	}
	state := document(t, text)
	deploymentTransformation(spec, images)(state)
	return state
}

func containers(state map[string]interface{}) []map[string]interface{} {
	podSpec := state["spec"].(map[string]interface{})["template"].(map[string]interface{})["spec"].(map[string]interface{})
	out := []map[string]interface{}{}
	for _, raw := range podSpec["containers"].([]interface{}) {
		out = append(out, raw.(map[string]interface{}))
	}
	return out
}

func envOf(container map[string]interface{}) map[string]string {
	out := map[string]string{}
	entries, _ := container["env"].([]interface{})
	for _, raw := range entries {
		entry := raw.(map[string]interface{})
		out[entry["name"].(string)] = entry["value"].(string)
	}
	return out
}

func TestImageTableBelongsToThePinnedOperatorRelease(t *testing.T) {
	table, err := loadImageTable()
	if err != nil {
		t.Fatalf("the image table must be rebuilt with every OperatorRelease move: %v", err)
	}
	names := []string{}
	for _, image := range table.Images {
		if !strings.HasPrefix(image.Name, "IMAGE_") {
			t.Errorf("%s: an operator image variable starts with IMAGE_", image.Name)
		}
		if !strings.HasPrefix(image.Image, upstreamRegistry+"/") || !strings.Contains(image.Image, "@sha256:") {
			t.Errorf("%s: %q must be a digest-pinned image Tekton publishes on %s", image.Name, image.Image, upstreamRegistry)
		}
		names = append(names, image.Name)
	}
	if !sort.StringsAreSorted(names) {
		t.Errorf("table entries must be sorted by name so both engines render one env order: %v", names)
	}
}

func TestAnEmptyRegistryRendersTheManifestUntouched(t *testing.T) {
	spec := &kubernetestektonoperatorv1alpha1.KubernetesTektonOperatorSpec{}
	for _, text := range []string{operatorDeployment, webhookDeployment} {
		if got, want := render(t, spec, text), document(t, text); !reflect.DeepEqual(got, want) {
			t.Errorf("an empty image_registry changed the manifest:\n got %#v\nwant %#v", got, want)
		}
	}
}

func TestARegistryMovesEveryTektonImageAtItsDigest(t *testing.T) {
	spec := &kubernetestektonoperatorv1alpha1.KubernetesTektonOperatorSpec{ImageRegistry: mirror}
	table, err := loadImageTable()
	if err != nil {
		t.Fatal(err)
	}

	operator := containers(render(t, spec, operatorDeployment))
	for _, container := range operator {
		if image := container["image"].(string); !strings.HasPrefix(image, mirror+"/tektoncd/operator/operator-") || !strings.HasSuffix(image, "@sha256:49f9f258920f77ca9db8031f8ee93b6abef0ff6a1a23254d937fd347601f4853") {
			t.Errorf("%s: image %q not moved at its digest", container["name"], image)
		}
	}

	lifecycle := envOf(operator[0])
	if lifecycle["SYSTEM_NAMESPACE"] != "tekton-operator" {
		t.Errorf("a variable that names no image changed: %q", lifecycle["SYSTEM_NAMESPACE"])
	}
	if got := lifecycle["IMAGE_PIPELINES_PROXY"]; got != mirror+"/tektoncd/operator/proxy-webhook-f6167da7bc41b96a27c5529f850e63d1:v0.80.0@sha256:2260f9bdf1699253a47e5d2664f70465f1a8c59ad7651cede2469a7586774341" {
		t.Errorf("the manifest's own IMAGE_PIPELINES_PROXY not moved: %q", got)
	}
	if got := lifecycle["IMAGE_JOB_PRUNER_TKN"]; got != mirror+"/tektoncd/plumbing/tkn@sha256:233de6c8b8583a34c2379fa98d42dba739146c9336e8d41b66030484357481ed" {
		t.Errorf("the manifest's own IMAGE_JOB_PRUNER_TKN not moved: %q", got)
	}
	for _, image := range table.Images {
		if got, want := lifecycle[image.Name], mirror+strings.TrimPrefix(image.Image, upstreamRegistry); got != want {
			t.Errorf("%s = %q, want %q", image.Name, got, want)
		}
	}
	if got, want := len(operator[0]["env"].([]interface{})), 3+len(table.Images); got != want {
		t.Errorf("lifecycle env has %d entries, want %d (no duplicates)", got, want)
	}

	if env := envOf(operator[1]); len(env) != 1 || env["PROFILING_PORT"] != "9009" {
		t.Errorf("the cluster-operations container reads no image variables and keeps its env: %v", env)
	}

	webhook := containers(render(t, spec, webhookDeployment))[0]["image"].(string)
	if webhook != mirror+"/tektoncd/operator/webhook-f2bb711aa8f0c0892856a4cbf6d9ddd8:v0.80.0@sha256:e4ec282861616a364dd8ad02cadec0fdabe26087605852ec85de3b7c084e5c50" {
		t.Errorf("webhook image not moved at its digest: %q", webhook)
	}
}

func TestImagesTektonDoesNotPublishKeepTheirRegistry(t *testing.T) {
	for _, image := range []string{
		"cgr.dev/chainguard/busybox@sha256:19f02276bf8dbdd62f069b922f10c65262cc34b710eea26ff928129a736be791",
		"bitnami/postgresql@sha256:ac8dd0d6512c4c5fb146c16b1c5f05862bd5f600d73348506ab4252587e7fcc6",
		"ghcr.io.example.com/tektoncd/pipeline/nop:v1",
	} {
		if got := mirroredImage(image, mirror); got != image {
			t.Errorf("mirroredImage(%q) = %q, want it unchanged", image, got)
		}
	}
}

func TestAnExplicitImageWinsOverTheRegistry(t *testing.T) {
	spec := &kubernetestektonoperatorv1alpha1.KubernetesTektonOperatorSpec{
		ImageRegistry: mirror,
		OperatorImage: &kubernetes.ContainerImage{Repo: "registry.internal/tekton/operator", Tag: "v0.80.0"},
	}
	for _, container := range containers(render(t, spec, operatorDeployment)) {
		if got := container["image"]; got != "registry.internal/tekton/operator:v0.80.0" {
			t.Errorf("%s: image %q, want the explicit operator_image", container["name"], got)
		}
	}
}

package module

import (
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// upstreamRegistry is the one registry Tekton publishes its images to. An
// image_registry replaces exactly this part of every image below and of the
// release manifest's own images; an image on any other host is not Tekton's
// and keeps its registry.
const upstreamRegistry = "ghcr.io"

// imageTable names every image the pinned operator installs that Tekton
// publishes, by the operator variable that redirects it. The operator
// rewrites a component's image only when its IMAGE_<COMPONENT>_<key>
// variable names it, keyed by container name or `arg_<flag>`, lowercased with
// `-` as `_` (pkg/reconciler/common/transformers.go at the operator tag); its
// own TEKTON_REGISTRY_OVERRIDE is not used because it rewrites every host,
// including the shell images on cgr.dev and mcr.microsoft.com and Results'
// bundled Postgres, which a mirror of Tekton's images does not carry.
//
// The entries are read from the component manifests bundled in the operator
// image itself (/var/run/ko/<component>/<version>/ in the image at
// OperatorRelease), never from the operator repository's components.yaml:
// at v0.80.0 that file names Results v0.19.0 while the image bundles, and
// installs, v0.18.0. Rebuild the table on every OperatorRelease move by
// exporting the operator image and listing each bundled Deployment's
// container images and `-*-image` args for Pipelines, Triggers, Dashboard,
// Chains, Results and the pruner. The manifest's own IMAGE_PIPELINES_PROXY
// and IMAGE_JOB_PRUNER_TKN are rewritten in place and are not repeated
// here. Entries are sorted by name so both engines render one env order.
//
// PARITY: the Terraform module carries this exact text (iac/tf/images.tf,
// under a `# parity:` marker the repository's cross-engine parity guard
// checks), so the two engines can never mirror different image sets.
const imageTable = `operatorRelease: v0.80.0
images:
  - name: IMAGE_CHAINS_TEKTON_CHAINS_CONTROLLER
    image: ghcr.io/tektoncd/chains/github.com/tektoncd/chains/cmd/controller:v0.27.1@sha256:2011424841784c67b7fe9f5a103304859ebd9bb809b3326de6698528b548be17
  - name: IMAGE_DASHBOARD_TEKTON_DASHBOARD
    image: ghcr.io/tektoncd/dashboard/dashboard-9623576a202fe86c8b7d1bc489905f86:v0.68.0@sha256:69eec56980bc586fdeeac282d2648ed79253a587a80610bb1bdee5aa109a814c
  - name: IMAGE_PIPELINES_ARG__ENTRYPOINT_IMAGE
    image: ghcr.io/tektoncd/pipeline/entrypoint-bff0a22da108bc2f16c818c97641a296:v1.12.0@sha256:3ec960b07abd85604242e146092e72f11be8787452c2a20d12a37fdee4a666e2
  - name: IMAGE_PIPELINES_ARG__NOP_IMAGE
    image: ghcr.io/tektoncd/pipeline/nop-8eac7c133edad5df719dc37b36b62482:v1.12.0@sha256:f89fb760b05fdef6895290e524d992b66ede72546622d5028b0406a9bea36d2f
  - name: IMAGE_PIPELINES_ARG__SIDECARLOGRESULTS_IMAGE
    image: ghcr.io/tektoncd/pipeline/sidecarlogresults-7501c6a20d741631510a448b48ab098f:v1.12.0@sha256:8b61bdcad62a99e7b15f9dc92690ea39f49be2c5a1c9f428a0aac712f349543d
  - name: IMAGE_PIPELINES_ARG__WORKINGDIRINIT_IMAGE
    image: ghcr.io/tektoncd/pipeline/workingdirinit-0c558922ec6a1b739e550e349f2d5fc1:v1.12.0@sha256:11031cbed2b8ddbb5af0947a5e0c997ba7cd3da5f309ddfa76e58f5418868d71
  - name: IMAGE_PIPELINES_CONTROLLER
    image: ghcr.io/tektoncd/pipeline/resolvers-ff86b24f130c42b88983d3c13993056d:v1.12.0@sha256:b1f06197342ed946b23efd1ce124d7c17d50678b461733a6cc0912b07a4943e7
  - name: IMAGE_PIPELINES_TEKTON_EVENTS_CONTROLLER
    image: ghcr.io/tektoncd/pipeline/events-a9042f7efb0cbade2a868a1ee5ddd52c:v1.12.0@sha256:d5d5869215a58fea01ce627d605786090949a324969102ab1fd25716adde10e9
  - name: IMAGE_PIPELINES_TEKTON_PIPELINES_CONTROLLER
    image: ghcr.io/tektoncd/pipeline/controller-10a3e32792f33651396d02b6855a6e36:v1.12.0@sha256:8d5f900677386b8fe5371429e9c1461b25de47e5e34736c7f99e82e2c279ebbc
  - name: IMAGE_PIPELINES_WEBHOOK
    image: ghcr.io/tektoncd/pipeline/webhook-d4749e605405422fd87700164e31b2d1:v1.12.0@sha256:1e4eded1773f6fa8ec67c9348f84bed0d39666a9698ebd1ffb4eaed88b38f7d4
  - name: IMAGE_PRUNER_CONTROLLER
    image: ghcr.io/tektoncd/pruner/controller-fb9e5cbe6aa14ed1aa64e966d9e4d295:v0.4.0@sha256:ba988fb989996c1bd67b8905656ee974a0efe3fc49b5605ecf12683a7371e9aa
  - name: IMAGE_PRUNER_WEBHOOK
    image: ghcr.io/tektoncd/pruner/webhook-19748cdf3f89caca6b5e27d42f93fa7f:v0.4.0@sha256:037dae99b235a333b89747eef944e3b9296a5329ab18588557c216111f7059a7
  - name: IMAGE_RESULTS_API
    image: ghcr.io/tektoncd/results/api-b1b7ffa9ba32f7c3020c3b68830b30a8:v0.18.0@sha256:72833d31749615fe45d4f7b7a3d4964e5da1cc2966e983fd31f67720d5c22720
  - name: IMAGE_RESULTS_RETENTION_POLICY_AGENT
    image: ghcr.io/tektoncd/results/retention-policy-agent-07427b345034d96a9a27896ebb138518:v0.18.0@sha256:b523e3d2bc70fbf8089b8b5ee42b506ccd335f10c633e7f975ec0d250e0d2aea
  - name: IMAGE_RESULTS_WATCHER
    image: ghcr.io/tektoncd/results/watcher-83f971ea227fb24157c0c699b824a628:v0.18.0@sha256:4b50291457cb63028e5f9e2acb921b852d9ad0f265562ea0fa83bafc8a58b367
  - name: IMAGE_TRIGGERS_ARG__EL_IMAGE
    image: ghcr.io/tektoncd/triggers/eventlistenersink-7ad1faa98cddbcb0c24990303b220bb8:v0.36.0@sha256:ca044dc8b45cb5f5089faf553b4cbd5f0f8e05fcf8d66db7a703c88a40163a73
  - name: IMAGE_TRIGGERS_TEKTON_TRIGGERS_CONTROLLER
    image: ghcr.io/tektoncd/triggers/controller-f656ca31de179ab913fa76abc255c315:v0.36.0@sha256:03f192f4aebeb3471db9d390d0532ab4d49e5aeeb70327df91be445f3629e859
  - name: IMAGE_TRIGGERS_TEKTON_TRIGGERS_CORE_INTERCEPTORS
    image: ghcr.io/tektoncd/triggers/interceptors-3176d6a3f314c3655b30bfd36e421dd5:v0.36.0@sha256:35e9444ce57aed35f378b18d75e71483d81c41dc7fb6bf44119c5730d617ce54
  - name: IMAGE_TRIGGERS_WEBHOOK
    image: ghcr.io/tektoncd/triggers/webhook-dd1edc925ee1772a9f76e2c1bc291ef6:v0.36.0@sha256:02d25a89d4b58e7cf341ea32c44d4938eaa3a43b6eafb61d88b9abb0d8a6f67d
`

// tektonImages is imageTable parsed.
type tektonImages struct {
	OperatorRelease string        `yaml:"operatorRelease"`
	Images          []tektonImage `yaml:"images"`
}

// tektonImage is one operator variable and the upstream image it names.
type tektonImage struct {
	Name  string `yaml:"name"`
	Image string `yaml:"image"`
}

// loadImageTable parses imageTable and refuses a table built for another
// operator release: handing a new operator last release's component images
// would run a version split the operator never tested, silently.
func loadImageTable() (*tektonImages, error) {
	table := &tektonImages{}
	if err := yaml.Unmarshal([]byte(imageTable), table); err != nil {
		return nil, errors.Wrap(err, "reading the Tekton image table")
	}
	if table.OperatorRelease != vars.OperatorRelease {
		return nil, errors.Errorf(
			"image_registry cannot be honored: the module's Tekton image table was read from operator %s, but the module installs operator %s. "+
				"Rebuild the table from the component manifests bundled in the %s operator image (see images.go), or leave image_registry empty to pull from ghcr.io",
			table.OperatorRelease, vars.OperatorRelease, vars.OperatorRelease)
	}
	return table, nil
}

// mirroredImage moves an image Tekton publishes to registry at the same path,
// tag and digest; an empty registry, or an image on another host, is
// returned unchanged. The digest pin travels with the path, so a mirror can
// serve only the bytes the release names.
func mirroredImage(image string, registry string) string {
	if registry == "" || !strings.HasPrefix(image, upstreamRegistry+"/") {
		return image
	}
	return registry + strings.TrimPrefix(image, upstreamRegistry)
}

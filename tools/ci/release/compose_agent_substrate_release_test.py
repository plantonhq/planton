#!/usr/bin/env python3
"""Tests for compose_agent_substrate_release.py: run with
`cd tools/ci/release && python3 -m unittest compose_agent_substrate_release_test.py`."""
import hashlib
import json
import os
import shutil
import tempfile
import unittest

from compose_agent_substrate_release import (
    ENVOY_PLACEHOLDER,
    CompositionError,
    check,
    check_shape,
    compose,
    read_upstream,
)

REPOSITORY = "registry.example.com/agent-substrate"
VERSION = "v0.6.0"
PATCH = "v0.6.1"
COMMIT = "e8d40c33f96f4b7a4984bfea7334a36ab5b45b7b"
MODULE = "ko://go.ate.dev/substrate"

COMPONENTS_GO = """package images

const ModulePath = "go.ate.dev/substrate"

var Components = []string{
	"cmd/ateapi",
	"cmd/atelet",
	"cmd/benchmarking/isolate/fake-atelet",
	"demos/multi-template/fspersist",
}

var released = []string{
	"cmd/ateapi",
	"cmd/atelet",
	"demos/multi-template/fspersist",
}

var EnvoyDataplane = DockerfileImage{Name: "envoy-dataplane", Context: "cmd/dataplane/envoy"}

var releasedDockerfiles = []DockerfileImage{EnvoyDataplane}
"""

API_MAIN_GO = """package main

var (
	snapshotLocation = pflag.String("snapshot-location", "", "Base object-storage URI.")
	previewFlags     = pflag.StringSlice("preview", nil, "Preview gates to enable.")
)

func init() {
	fs.StringVar(&cfg.AllowedIDs, "ingress-allowed-spiffe-ids", nil, "SPIFFE IDs.")
	cacheFlags = filecachegc.NewFlags(pflag.CommandLine, "snapshot-cache", "snapshot cache")
	slog.String("not-a-flag", "logged")
	loadFromEnv(snapshotLocation, "ATE_API_SNAPSHOT_LOCATION")
}
"""

PREVIEW_GO = """package preview

const (
	GateExternalVolumes Gate = "ExternalVolumes"

	GatePerTemplateSnapshotLocation Gate = "PerTemplateSnapshotLocation"
)
"""

APP_YAML = f"""# The installer resolves every ko:// image reference below.
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
spec:
  template:
    spec:
      containers:
      - name: api
        image: {MODULE}/cmd/ateapi
      - name: node
        image: "{MODULE}/cmd/atelet"
      - name: envoy
        image: ${{ENVOY_DATAPLANE_IMAGE}}
"""

CRD_YAML = """# Copyright upstream.
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: workerpools.ate.dev
spec:
  group: ate.dev
"""

TEMPLATE = "assets: ${BUCKET_SCHEME}://${BUCKET_NAME}/kata\n"

# An opt-in component naming an image no release publishes, as upstream's
# benchmarking fakes are.
FAKE_COMPONENT = f"""apiVersion: kustomize.config.k8s.io/v1alpha1
kind: Component
patches:
  - patch: |
      - op: replace
        path: /spec/template/spec/containers/0/image
        value: {MODULE}/cmd/benchmarking/isolate/fake-atelet
"""


def write(path, text):
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, "w", encoding="utf-8") as f:
        f.write(text)


def write_index(index_manifests, name, platforms=("linux/amd64", "linux/arm64"), media_type="application/vnd.oci.image.index.v1+json", build=""):
    """Save one image's index as its registry serves it, and return its digest."""
    index = json.dumps({
        "schemaVersion": 2,
        "mediaType": media_type,
        "annotations": {"build": build},
        "manifests": [
            {"platform": {"os": p.split("/")[0], "architecture": p.split("/")[1]}, "digest": "sha256:" + "0" * 64}
            for p in platforms
        ],
    }).encode()
    os.makedirs(index_manifests, exist_ok=True)
    with open(os.path.join(index_manifests, f"{name}.json"), "wb") as f:
        f.write(index)
    return "sha256:" + hashlib.sha256(index).hexdigest()


class Fixture:
    """An upstream checkout at a tag and the indexes its registry serves."""

    def __init__(self, root, version=VERSION):
        self.source = os.path.join(root, "source")
        self.indexes = os.path.join(root, "indexes")
        self.root = root
        self.version = version
        write(os.path.join(self.source, "cmd/ate-setup/internal/images/images.go"), COMPONENTS_GO)
        write(os.path.join(self.source, "cmd/ate-setup/internal/config/config.go"), "package config\n")
        write(os.path.join(self.source, "internal/installdefaults/installdefaults.go"), "package installdefaults\n")
        write(os.path.join(self.source, "cmd/atelet/internal/filecachegc/flags.go"), "package filecachegc\n")
        write(os.path.join(self.source, "pkg/postgressetup/setup.sql"), "CREATE SCHEMA substrate;\n")
        write(os.path.join(self.source, "cmd/ate-setup/internal/steps/overlay.go"), "package steps\n")
        write(os.path.join(self.source, "cmd/ateapi/main.go"), API_MAIN_GO)
        write(os.path.join(self.source, "internal/preview/preview.go"), PREVIEW_GO)
        write(os.path.join(self.source, "manifests/ate-install/app.yaml"), APP_YAML)
        write(os.path.join(self.source, "manifests/ate-install/generated/ate.dev_workerpools.yaml"), CRD_YAML)
        write(os.path.join(self.source, "manifests/ate-install/base/kustomization.yaml"),
              "apiVersion: kustomize.config.k8s.io/v1beta1\nkind: Kustomization\nresources:\n  - ../app.yaml\n")
        write(os.path.join(self.source, "manifests/ate-install/components/pin/kustomization.yaml"),
              "apiVersion: kustomize.config.k8s.io/v1alpha1\nkind: Component\n")
        write(os.path.join(self.source, "manifests/ate-install/components/fake-data-plane/kustomization.yaml"), FAKE_COMPONENT)
        write(os.path.join(self.source, "manifests/microvm/sandboxconfig.yaml.tmpl"), TEMPLATE)
        write(os.path.join(self.source, "LICENSE"), "Apache License 2.0\n")
        write(os.path.join(self.source, "_LICENSES/github.com/x/y/LICENSE"), "MIT\n")
        self.digests = {
            name: write_index(self.indexes, name, build=version)
            for name in ("ateapi", "atelet", "fspersist", "envoy-dataplane")
        }

    def compose(self, out_name="out", shape=None):
        out = os.path.join(self.root, out_name)
        os.makedirs(out)
        return out, compose(self.source, self.indexes, REPOSITORY, self.version, shape or self.version, COMMIT, out)

    def pinned(self, name):
        return f"{REPOSITORY}/{name}:{self.version}@{self.digests[name]}"


class ComposeTest(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.mkdtemp()
        self.fixture = Fixture(self.tmp)

    def tearDown(self):
        shutil.rmtree(self.tmp)

    def read(self, out, relative):
        with open(os.path.join(out, relative), encoding="utf-8") as f:
            return f.read()

    def test_references_become_tag_and_digest_and_prose_is_left_alone(self):
        out, _ = self.fixture.compose()
        app = self.read(out, "manifests/ate-install/app.yaml")
        self.assertIn(f"image: {self.fixture.pinned('ateapi')}\n", app)
        self.assertIn(f'image: "{self.fixture.pinned("atelet")}"\n', app)
        self.assertIn(f"image: {self.fixture.pinned('envoy-dataplane')}\n", app)
        self.assertIn("# The installer resolves every ko:// image reference below.", app)

    def test_templates_keep_their_own_placeholders(self):
        out, _ = self.fixture.compose()
        self.assertEqual(self.read(out, "manifests/microvm/sandboxconfig.yaml.tmpl"), TEMPLATE)

    def test_a_reference_extending_a_published_one_fails_instead_of_rewriting_its_prefix(self):
        write(os.path.join(self.fixture.source, "manifests/ate-install/sidecar.yaml"),
              f"image: {MODULE}/cmd/atelet-sidecar\n")
        with self.assertRaises(CompositionError) as raised:
            self.fixture.compose()
        self.assertIn(f"{MODULE}/cmd/atelet-sidecar (in manifests/ate-install/sidecar.yaml) has no published image", str(raised.exception))

    def test_an_image_without_both_platforms_fails(self):
        write_index(self.fixture.indexes, "atelet", platforms=("linux/amd64",))
        with self.assertRaises(CompositionError) as raised:
            self.fixture.compose()
        self.assertIn("lacks linux/arm64", str(raised.exception))

    def test_a_single_platform_image_fails(self):
        write_index(self.fixture.indexes, "atelet", media_type="application/vnd.oci.image.manifest.v1+json")
        with self.assertRaises(CompositionError):
            self.fixture.compose()

    def test_the_fetched_set_must_equal_upstreams_published_set(self):
        os.remove(os.path.join(self.fixture.indexes, "fspersist.json"))
        write_index(self.fixture.indexes, "fake-atelet")
        with self.assertRaises(CompositionError) as raised:
            self.fixture.compose()
        message = str(raised.exception)
        self.assertIn("no index fetched for fspersist", message)
        self.assertIn("fetched images fake-atelet are not in upstream's published set", message)

    def test_the_digest_is_the_hash_of_the_served_index(self):
        _, references = self.fixture.compose()
        with open(os.path.join(self.fixture.indexes, "ateapi.json"), "rb") as f:
            self.assertEqual(references["ateapi"], f"{REPOSITORY}/ateapi:{VERSION}@sha256:{hashlib.sha256(f.read()).hexdigest()}")

    def test_a_component_naming_an_unreleased_image_is_left_out_whole(self):
        out, references = self.fixture.compose()
        self.assertFalse(os.path.exists(os.path.join(out, "manifests/ate-install/components/fake-data-plane")))
        self.assertTrue(os.path.exists(os.path.join(out, "manifests/ate-install/components/pin/kustomization.yaml")))
        self.assertNotIn("fake-atelet", references)

    def test_an_install_file_naming_an_unreleased_image_fails(self):
        write(os.path.join(self.fixture.source, "manifests/ate-install/bench.yaml"),
              f"image: {MODULE}/cmd/benchmarking/isolate/fake-atelet\n")
        with self.assertRaises(CompositionError) as raised:
            self.fixture.compose()
        self.assertIn("manifests/ate-install/bench.yaml names", str(raised.exception))
        self.assertIn("which no release publishes", str(raised.exception))

    def test_crds_yaml_holds_only_the_crds(self):
        out, _ = self.fixture.compose()
        crds = self.read(out, "crds.yaml")
        self.assertEqual(crds.count("---\n"), 1)
        self.assertIn("name: workerpools.ate.dev", crds)
        self.assertNotIn("Copyright", crds)
        self.assertNotIn("kind: Deployment", crds)

    def test_images_json_names_the_source_and_every_pinned_image(self):
        out, references = self.fixture.compose()
        images = json.loads(self.read(out, "images.json"))
        self.assertEqual(images["version"], VERSION)
        self.assertEqual(images["shape"], VERSION)
        self.assertEqual(images["source"], {"repository": "https://github.com/agent-substrate/substrate", "tag": VERSION, "commit": COMMIT})
        self.assertEqual(images["images"], references)
        self.assertEqual(sorted(references), ["ateapi", "atelet", "envoy-dataplane", "fspersist"])

    def test_checksums_cover_every_file_and_match(self):
        out, _ = self.fixture.compose()
        listed = {}
        for line in self.read(out, "SHA256SUMS").splitlines():
            digest, relative = line.split("  ", 1)
            listed[relative] = digest
        on_disk = set()
        for root, _, files in os.walk(out):
            for name in files:
                relative = os.path.relpath(os.path.join(root, name), out)
                if relative != "SHA256SUMS":
                    on_disk.add(relative)
                    with open(os.path.join(out, relative), "rb") as f:
                        self.assertEqual(listed[relative], hashlib.sha256(f.read()).hexdigest(), relative)
        self.assertEqual(set(listed), on_disk)
        self.assertIn("third-party-licenses.tar.gz", on_disk)
        self.assertIn("LICENSE", on_disk)
        self.assertIn("setup.sql", on_disk)

    def test_composing_twice_writes_the_same_bytes(self):
        first, _ = self.fixture.compose("first")
        second, _ = self.fixture.compose("second")
        self.assertEqual(self.read(first, "SHA256SUMS"), self.read(second, "SHA256SUMS"))

    @unittest.skipUnless(shutil.which("kustomize"), "kustomize is not installed")
    def test_check_builds_every_kustomization_and_accepts_the_composed_tree(self):
        out, references = self.fixture.compose()
        check(out, REPOSITORY, references, shutil.which("kustomize"))

    @unittest.skipUnless(shutil.which("kustomize"), "kustomize is not installed")
    def test_check_rejects_a_reference_the_release_does_not_publish(self):
        out, references = self.fixture.compose()
        write(os.path.join(out, "manifests/ate-install/stray.yaml"), f"image: {REPOSITORY}/atelet:v0.4.0\n")
        with self.assertRaises(CompositionError) as raised:
            check(out, REPOSITORY, references, shutil.which("kustomize"))
        self.assertIn(f"names {REPOSITORY}/atelet:v0.4.0", str(raised.exception))


class ShapeTest(unittest.TestCase):
    """A patch claiming the shape of the release before it."""

    def setUp(self):
        self.tmp = tempfile.mkdtemp()
        self.shape = Fixture(os.path.join(self.tmp, "shape"))
        self.shape_release, _ = self.shape.compose()
        self.patch = Fixture(os.path.join(self.tmp, "patch"), version=PATCH)

    def tearDown(self):
        shutil.rmtree(self.tmp)

    def claim(self):
        out, _ = self.patch.compose(shape=VERSION)
        check_shape(out, REPOSITORY, PATCH, VERSION, self.shape_release, self.patch.source, self.shape.source)
        return out

    def test_a_release_that_differs_only_in_its_images_keeps_the_shape(self):
        out = self.claim()
        with open(os.path.join(out, "images.json"), encoding="utf-8") as f:
            images = json.load(f)
        self.assertEqual((images["version"], images["shape"]), (PATCH, VERSION))
        self.assertNotEqual(self.patch.digests, self.shape.digests)

    def test_a_changed_manifest_refuses_the_claim(self):
        with open(os.path.join(self.patch.source, "manifests/ate-install/app.yaml"), "a", encoding="utf-8") as f:
            f.write("        args: [--new-required-flag]\n")
        with self.assertRaises(CompositionError) as raised:
            self.claim()
        self.assertIn("changed: manifests/ate-install/app.yaml", str(raised.exception))
        self.assertIn(f"declare {PATCH} as its own shape", str(raised.exception))

    def test_a_new_manifest_file_refuses_the_claim(self):
        write(os.path.join(self.patch.source, "manifests/ate-install/new-component.yaml"), "kind: Service\n")
        with self.assertRaises(CompositionError) as raised:
            self.claim()
        self.assertIn("present in one release only: manifests/ate-install/new-component.yaml", str(raised.exception))

    def test_a_changed_definition_refuses_the_claim(self):
        with open(os.path.join(self.patch.source, "manifests/ate-install/generated/ate.dev_workerpools.yaml"), "a", encoding="utf-8") as f:
            f.write("  names:\n    kind: WorkerPool\n")
        with self.assertRaises(CompositionError) as raised:
            self.claim()
        self.assertIn("crds.yaml", str(raised.exception))

    def test_a_changed_setup_script_refuses_the_claim(self):
        write(os.path.join(self.patch.source, "pkg/postgressetup/setup.sql"), "CREATE SCHEMA substrate_v2;\n")
        with self.assertRaises(CompositionError) as raised:
            self.claim()
        self.assertIn("changed: setup.sql", str(raised.exception))

    def test_a_changed_installer_refuses_the_claim(self):
        write(os.path.join(self.patch.source, "cmd/ate-setup/internal/steps/apienv.go"), "package steps\n")
        with self.assertRaises(CompositionError) as raised:
            self.claim()
        self.assertIn("upstream changed how an install is assembled", str(raised.exception))
        self.assertIn("cmd/ate-setup/internal/steps/apienv.go", str(raised.exception))

    def test_a_removed_flag_refuses_the_claim(self):
        write(os.path.join(self.patch.source, "cmd/ateapi/main.go"), API_MAIN_GO.replace('"snapshot-location"', '"snapshot-uri"'))
        with self.assertRaises(CompositionError) as raised:
            self.claim()
        self.assertIn(f"{PATCH} no longer accepts settings {VERSION} accepted: cmd/ateapi: --snapshot-location", str(raised.exception))
        self.assertNotIn("not-a-flag", str(raised.exception))

    def test_a_removed_flag_prefix_refuses_the_claim(self):
        write(os.path.join(self.patch.source, "cmd/ateapi/main.go"), API_MAIN_GO.replace('"snapshot-cache"', '"snapshots"'))
        with self.assertRaises(CompositionError) as raised:
            self.claim()
        self.assertIn("cmd/ateapi: --snapshot-cache-*", str(raised.exception))

    def test_a_removed_environment_variable_refuses_the_claim(self):
        write(os.path.join(self.patch.source, "cmd/ateapi/main.go"), API_MAIN_GO.replace('"ATE_API_SNAPSHOT_LOCATION"', '"ATE_API_SNAPSHOT_URI"'))
        with self.assertRaises(CompositionError) as raised:
            self.claim()
        self.assertIn("cmd/ateapi: $ATE_API_SNAPSHOT_LOCATION", str(raised.exception))

    def test_a_removed_preview_gate_refuses_the_claim(self):
        write(os.path.join(self.patch.source, "internal/preview/preview.go"), PREVIEW_GO.replace('GateExternalVolumes Gate = "ExternalVolumes"\n', ""))
        with self.assertRaises(CompositionError) as raised:
            self.claim()
        self.assertIn("preview gate ExternalVolumes", str(raised.exception))

    def test_a_moved_install_deciding_directory_refuses_the_claim(self):
        shutil.rmtree(os.path.join(self.patch.source, "internal/installdefaults"))
        with self.assertRaises(CompositionError) as raised:
            self.claim()
        self.assertIn("has no internal/installdefaults", str(raised.exception))

    def test_an_added_flag_keeps_the_shape(self):
        write(os.path.join(self.patch.source, "cmd/atelet/flags.go"), 'package main\n\nvar budget = pflag.Int("upload-budget", 0, "New.")\n')
        self.claim()


class ReadUpstreamTest(unittest.TestCase):
    def read(self, images_go):
        tmp = tempfile.mkdtemp()
        self.addCleanup(shutil.rmtree, tmp)
        write(os.path.join(tmp, "cmd/ate-setup/internal/images/images.go"), images_go)
        return read_upstream(tmp)

    def test_a_missing_module_path_fails_with_the_file_named(self):
        with self.assertRaises(CompositionError) as raised:
            self.read(COMPONENTS_GO.replace('const ModulePath = "go.ate.dev/substrate"\n', ""))
        self.assertIn("ModulePath", str(raised.exception))

    def test_a_moved_component_list_fails_with_the_file_named(self):
        with self.assertRaises(CompositionError) as raised:
            self.read("package images\n")
        self.assertIn("cmd/ate-setup/internal/images/images.go", str(raised.exception))

    def test_a_released_package_outside_components_fails(self):
        with self.assertRaises(CompositionError) as raised:
            self.read(COMPONENTS_GO.replace('var released = []string{\n', 'var released = []string{\n\t"cmd/elsewhere",\n'))
        self.assertIn("`var released", str(raised.exception))

    def test_an_envoy_image_no_release_publishes_fails(self):
        with self.assertRaises(CompositionError) as raised:
            self.read(COMPONENTS_GO.replace("[]DockerfileImage{EnvoyDataplane}", "[]DockerfileImage{}"))
        self.assertIn(ENVOY_PLACEHOLDER, str(raised.exception))

    def test_image_names_are_the_released_package_base_names(self):
        upstream = self.read(COMPONENTS_GO)
        self.assertEqual(upstream.image_names, {"ateapi", "atelet", "fspersist", "envoy-dataplane"})
        self.assertEqual(upstream.image_by_reference[f"{MODULE}/demos/multi-template/fspersist"], "fspersist")
        self.assertEqual(upstream.image_by_reference[f"{MODULE}/cmd/benchmarking/isolate/fake-atelet"], "fake-atelet")


if __name__ == "__main__":
    unittest.main()

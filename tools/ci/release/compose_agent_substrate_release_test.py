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
    CompositionError,
    check,
    compose,
    read_upstream,
)

REPOSITORY = "registry.example.com/agent-substrate"
VERSION = "v0.5.0"
COMMIT = "e8d40c33f96f4b7a4984bfea7334a36ab5b45b7b"
MODULE = "ko://github.com/agent-substrate/substrate"

COMPONENTS_GO = """package images

var Components = []string{
	"cmd/ateapi",
	"cmd/atelet",
	"demos/multi-template/fspersist",
}
"""

OVERLAY_GO = """package steps

const (
	envoyDataplaneImage = "envoy-dataplane"
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


def write(path, text):
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, "w", encoding="utf-8") as f:
        f.write(text)


def write_layout(layouts, name, platforms=("linux/amd64", "linux/arm64"), media_type="application/vnd.oci.image.index.v1+json"):
    index = json.dumps({
        "schemaVersion": 2,
        "mediaType": media_type,
        "manifests": [
            {"platform": {"os": p.split("/")[0], "architecture": p.split("/")[1]}, "digest": "sha256:" + "0" * 64}
            for p in platforms
        ],
    }).encode()
    digest = hashlib.sha256(index).hexdigest()
    blob = os.path.join(layouts, name, "blobs", "sha256", digest)
    os.makedirs(os.path.dirname(blob))
    with open(blob, "wb") as f:
        f.write(index)
    write(os.path.join(layouts, name, "index.json"), json.dumps({"manifests": [{"digest": "sha256:" + digest}]}))
    return "sha256:" + digest


class Fixture:
    """An upstream checkout at a tag and the layouts its build exported."""

    def __init__(self, root):
        self.source = os.path.join(root, "source")
        self.layouts = os.path.join(root, "layouts")
        self.root = root
        write(os.path.join(self.source, "cmd/ate-setup/internal/images/images.go"), COMPONENTS_GO)
        write(os.path.join(self.source, "cmd/ate-setup/internal/steps/overlay.go"), OVERLAY_GO)
        write(os.path.join(self.source, "manifests/ate-install/app.yaml"), APP_YAML)
        write(os.path.join(self.source, "manifests/ate-install/generated/ate.dev_workerpools.yaml"), CRD_YAML)
        write(os.path.join(self.source, "manifests/ate-install/base/kustomization.yaml"),
              "apiVersion: kustomize.config.k8s.io/v1beta1\nkind: Kustomization\nresources:\n  - ../app.yaml\n")
        write(os.path.join(self.source, "manifests/ate-install/components/pin/kustomization.yaml"),
              "apiVersion: kustomize.config.k8s.io/v1alpha1\nkind: Component\n")
        write(os.path.join(self.source, "manifests/microvm/sandboxconfig.yaml.tmpl"), TEMPLATE)
        write(os.path.join(self.source, "LICENSE"), "Apache License 2.0\n")
        write(os.path.join(self.source, "_LICENSES/github.com/x/y/LICENSE"), "MIT\n")
        self.digests = {name: write_layout(self.layouts, name) for name in ("ateapi", "atelet", "fspersist", "envoy-dataplane")}

    def compose(self, out_name="out"):
        out = os.path.join(self.root, out_name)
        os.makedirs(out)
        return out, compose(self.source, self.layouts, REPOSITORY, VERSION, COMMIT, out)

    def pinned(self, name):
        return f"{REPOSITORY}/{name}:{VERSION}@{self.digests[name]}"


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
        shutil.rmtree(os.path.join(self.fixture.layouts, "atelet"))
        write_layout(self.fixture.layouts, "atelet", platforms=("linux/amd64",))
        with self.assertRaises(CompositionError) as raised:
            self.fixture.compose()
        self.assertIn("lacks linux/arm64", str(raised.exception))

    def test_a_single_platform_image_fails(self):
        shutil.rmtree(os.path.join(self.fixture.layouts, "atelet"))
        write_layout(self.fixture.layouts, "atelet", media_type="application/vnd.oci.image.manifest.v1+json")
        with self.assertRaises(CompositionError):
            self.fixture.compose()

    def test_the_built_set_must_equal_upstreams_published_set(self):
        shutil.rmtree(os.path.join(self.fixture.layouts, "fspersist"))
        write_layout(self.fixture.layouts, "fake-extra")
        with self.assertRaises(CompositionError) as raised:
            self.fixture.compose()
        message = str(raised.exception)
        self.assertIn("no built image for fspersist", message)
        self.assertIn("built images fake-extra are not in upstream's published set", message)

    def test_a_tampered_archive_fails(self):
        layout = os.path.join(self.fixture.layouts, "ateapi")
        digest = self.fixture.digests["ateapi"].split(":", 1)[1]
        with open(os.path.join(layout, "blobs", "sha256", digest), "ab") as f:
            f.write(b" ")
        with self.assertRaises(CompositionError) as raised:
            self.fixture.compose()
        self.assertIn("does not hash to", str(raised.exception))

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


class ReadUpstreamTest(unittest.TestCase):
    def test_a_moved_component_list_fails_with_the_file_named(self):
        tmp = tempfile.mkdtemp()
        try:
            write(os.path.join(tmp, "cmd/ate-setup/internal/steps/overlay.go"), OVERLAY_GO)
            with self.assertRaises(CompositionError) as raised:
                read_upstream(tmp)
            self.assertIn("cmd/ate-setup/internal/images/images.go", str(raised.exception))
        finally:
            shutil.rmtree(tmp)

    def test_image_names_are_the_package_base_names(self):
        tmp = tempfile.mkdtemp()
        try:
            write(os.path.join(tmp, "cmd/ate-setup/internal/images/images.go"), COMPONENTS_GO)
            write(os.path.join(tmp, "cmd/ate-setup/internal/steps/overlay.go"), OVERLAY_GO)
            upstream = read_upstream(tmp)
            self.assertEqual(upstream.image_names, {"ateapi", "atelet", "fspersist", "envoy-dataplane"})
            self.assertEqual(upstream.image_by_reference[f"{MODULE}/demos/multi-template/fspersist"], "fspersist")
        finally:
            shutil.rmtree(tmp)


if __name__ == "__main__":
    unittest.main()

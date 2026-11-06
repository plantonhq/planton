#!/usr/bin/env python3
"""Tests for publish_stub_sdks.sh: run with `python3 -m unittest tools/ci/release/publish_stub_sdks_test.py`.

The stub zips are not byte-reproducible, so the lane seals the release with its
manifest and refuses to describe bytes it did not build. A fake `aws` stands in
for R2 (publish_once_test.py's), a fake `make` builds the stub trees with bytes
the test chooses, and a fake `buf` reports its version.
"""
import hashlib
import os
import subprocess
import tempfile
import unittest

from publish_once_test import FAKE_AWS as ONCE_FAKE_AWS

# publish_once_test's fake keeps an object's checksum metadata in
# "<object>.sha256", which is also where this lane's own .sha256 objects
# would land; a suffix no key ends with keeps the two apart.
FAKE_AWS = ONCE_FAKE_AWS.replace('.sha256"', '.metadata"')

HERE = os.path.dirname(os.path.abspath(__file__))
SCRIPT = os.path.join(HERE, "publish_stub_sdks.sh")
TAG = "v9.9.9"

# `make build-stub-sdks tag=<tag>`: write the five artifacts the real target
# writes, from $STUB_BYTES, and count the builds.
FAKE_MAKE = r'''#!/usr/bin/env bash
set -euo pipefail
echo built >> "${FAKE_BUILDS}"
out=generated/stub-sdks
mkdir -p "${out}"
printf 'java %s' "${STUB_BYTES}" > "${out}/stubs-java.zip"
printf 'ts %s' "${STUB_BYTES}" > "${out}/stubs-ts.zip"
for zip in stubs-java.zip stubs-ts.zip; do
  printf '%s  %s\n' "$(sha256sum "${out}/${zip}" | cut -d' ' -f1)" "${zip}" > "${out}/${zip}.sha256"
done
printf '{"releaseTag": "%s"}\n' "${2#tag=}" > "${out}/stub-sdks-manifest.json"
'''

FAKE_BUF = '#!/usr/bin/env bash\necho "${FAKE_BUF_VERSION:-1.65.0}"\n'


def sha256(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


class PublishStubSDKsTest(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.store = os.path.join(self.tmp.name, "r2")
        self.work = os.path.join(self.tmp.name, "repo")
        bindir = os.path.join(self.tmp.name, "bin")
        for d in (self.store, self.work, bindir):
            os.mkdir(d)
        for name, body in (("aws", FAKE_AWS), ("make", FAKE_MAKE), ("buf", FAKE_BUF)):
            path = os.path.join(bindir, name)
            with open(path, "w") as f:
                f.write(body)
            os.chmod(path, 0o755)
        self.builds = os.path.join(self.tmp.name, "builds")
        self.env = dict(os.environ, PATH=bindir + os.pathsep + os.environ["PATH"], FAKE_R2=self.store,
                        FAKE_BUILDS=self.builds, R2_BUCKET="bucket", AWS_ENDPOINT_URL="https://r2", STUB_BYTES="one")

    def tearDown(self):
        self.tmp.cleanup()

    def run_lane(self, **env):
        return subprocess.run(["bash", SCRIPT, TAG], cwd=self.work, env=dict(self.env, **env), capture_output=True, text=True)

    def published(self, name):
        path = os.path.join(self.store, f"releases/{TAG}/stubs/{name}".replace("/", "_"))
        if not os.path.exists(path):
            return None
        with open(path, "rb") as f:
            return f.read()

    def builds_run(self):
        if not os.path.exists(self.builds):
            return 0
        with open(self.builds) as f:
            return len(f.read().split())

    def test_a_first_run_publishes_each_zip_its_checksum_and_seals_with_the_manifest(self):
        result = self.run_lane()
        self.assertEqual(result.returncode, 0, result.stderr)
        for zip, content in (("stubs-java.zip", b"java one"), ("stubs-ts.zip", b"ts one")):
            self.assertEqual(self.published(zip), content)
            self.assertEqual(self.published(zip + ".sha256").decode(), f"{sha256(content)}  {zip}\n")
        self.assertIn(TAG, self.published("stub-sdks-manifest.json").decode())

    def test_a_sealed_release_is_left_alone_and_nothing_is_rebuilt(self):
        self.run_lane()
        result = self.run_lane(STUB_BYTES="two")
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIn("published and sealed", result.stdout)
        self.assertEqual(self.builds_run(), 1)
        self.assertEqual(self.published("stubs-java.zip"), b"java one")

    def test_an_unsealed_earlier_run_with_other_bytes_stops_before_the_manifest(self):
        self.run_lane()
        os.remove(os.path.join(self.store, f"releases/{TAG}/stubs/stub-sdks-manifest.json".replace("/", "_")))
        result = self.run_lane(STUB_BYTES="two")
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("already holds other bytes", result.stderr)
        self.assertIsNone(self.published("stub-sdks-manifest.json"))

    def test_a_buf_other_than_the_pin_is_refused_before_building(self):
        result = self.run_lane(FAKE_BUF_VERSION="1.70.0")
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("pins 1.65.0", result.stderr)
        self.assertEqual(self.builds_run(), 0)


if __name__ == "__main__":
    unittest.main()

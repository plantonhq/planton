#!/usr/bin/env python3
"""Compose a published Agent Substrate release from upstream's published images.

Agent Substrate publishes its release images itself
(ghcr.io/agent-substrate/substrate/<image>:<tag>, from v1.0.0) but no
pre-built install. This script turns upstream's tag and those images into the
release set downloads.planton.ai serves under releases/agent-substrate/<version>/:

    manifests/...                upstream's manifests/ tree, file for file, with
                                 only the image references filled in, less the
                                 overlays that name images upstream never
                                 releases
    crds.yaml                    every CustomResourceDefinition in that tree, in
                                 file order: the bundle the kept-CRD mechanism
                                 fetches (pkg/kubernetes/helmcrds, BundleURL)
    images.json                  the version, its install shape, the upstream
                                 commit, and every image's digest-pinned
                                 reference
    setup.sql                    upstream's database setup script
                                 (pkg/postgressetup/setup.sql), which the
                                 control-plane module runs in the bundled
                                 database
    LICENSE                      upstream's license
    third-party-licenses.tar.gz  upstream's _LICENSES/ tree: the notices of every
                                 dependency compiled into the images
    SHA256SUMS                   every file above; publishing it seals the version

Why the whole tree and not one install file: upstream's installer assembles
the install from about a dozen settings (kustomize overlays for the gateway,
cluster size, Cloud SQL and node pinning; text markers it fills in for the
credential provider and extra egress filters). One flattened file would fix
every one of those choices. The tree keeps them all, and a module fetches the
files it composes by URL, the way the catalog's release-manifest kinds do.

The rewrite is upstream's own rule (cmd/ate-setup/internal/images/prebuilt.go,
`ate-setup --image-repo`), read from the checkout rather than restated:

  - the manifests name images as ko://<module>/<package> references, and the
    packages that have an image are upstream's closed `Components` list; a
    reference outside it fails rather than being guessed at;
  - a release publishes only the packages in `released` and the Dockerfile
    images in `releasedDockerfiles` (Envoy); an overlay that names any other
    Component (the benchmarking fakes) cannot be installed from a release, so
    its directory is left out of the set, as upstream's installer would fail
    it at install time;
  - each reference becomes <repository>/<image>:<version>@<digest>, the
    tag-and-digest form upstream's pre-built install writes;
  - ${ENVOY_DATAPLANE_IMAGE} becomes the Dockerfile-built Envoy image, pinned
    the same way.

The digests are never taken on trust: the lane fetches each image's
multi-platform index from upstream's registry (`crane manifest`), and the
digest is the sha256 of those bytes, the same value every registry the image
is copied to answers with. The index must cover both platforms upstream builds.

Every output is a pure function of its inputs (sorted walks, fixed archive
timestamps), so a re-run of a version composes the same bytes and the publish
step can prove what it serves equals what this run composed.

The install shape. Planton's control-plane module composes one install: the
files, objects, flags and settings of one release, its shape. A release whose
install is the shape's own apart from its images installs with that module
unchanged, so a customer adopts it by setting the kind's version, with no
catalog release. images.json records the shape every release belongs to, and
the module refuses a release of any other shape before it applies anything.
A release is its own shape (--shape equal to --version) or claims an earlier
one, and the claim is proven here, never trusted: the composed tree must equal
the shape's published tree once image references are reduced to their names,
file for file, and the upstream code that decides how an install is assembled
(the installer's steps and settings, the binaries' install defaults) must be
the same at both tags. Any difference refuses the claim; the release is then
its own shape and waits for the module to compose it.

Usage:
    compose_agent_substrate_release.py images --source <upstream checkout at the tag>
        prints the images a release of the tag publishes, one per line
    compose_agent_substrate_release.py --source <upstream checkout at the tag>
        --index-manifests <directory of <image>.json, each image's index as its registry serves it>
        --repository <registry path the images are published under>
        --version vX.Y.Z --shape vX.Y.Z --commit <sha> --out <directory>
        [--shape-release <the shape's published tree>
         --shape-source <upstream checkout at the shape's tag>]
        [--kustomize <kustomize binary>]

With --kustomize, every kustomization in the composed tree is built, and every
image reference under the repository, raw or rendered, must be one this
release publishes.
"""
import argparse
import gzip
import hashlib
import io
import json
import os
import re
import shutil
import subprocess
import sys
import tarfile
from dataclasses import dataclass

UPSTREAM_REPOSITORY = "https://github.com/agent-substrate/substrate"

# Where upstream keeps the facts the rewrite needs: its module path (a ko://
# reference is it joined with a package), the closed package list, the
# packages and Dockerfile images a release publishes, and the Envoy image's
# name. A release that moves any of them fails composition with the file
# named, never silently.
COMPONENTS_FILE = "cmd/ate-setup/internal/images/images.go"

# Upstream's database setup script, published beside the manifests so the
# seal covers it.
SETUP_SQL_FILE = "pkg/postgressetup/setup.sql"

# The upstream code that decides how an install is assembled beyond the
# manifests: the installer's steps (the objects and flags it adds, the markers
# it fills), its settings, and the Google Cloud identities and grants an
# install needs. A release claiming an earlier shape must not change them.
SHAPE_SOURCE_PATHS = (
    "cmd/ate-setup/internal/steps",
    "cmd/ate-setup/internal/config",
    # The namespace and Service names every binary defaults to, matching the
    # canonical install layout.
    "internal/installdefaults",
    # Builds the node agent's cache flags from a prefix (--<prefix>-gc-period
    # and its siblings), which FLAG_PREFIX records.
    "cmd/atelet/internal/filecachegc",
    # The Google Cloud setup (tools/setup-gcp) left upstream's repository at
    # v1.0.0 for ai-on-gke/substrate-gke, so no shape comparison covers it;
    # the kind's own GKE keyless wiring is what installs on Google Cloud.
)

# The settings the module passes by name: every flag an installed binary
# registers (a call on a flag set whose first string argument is the flag's
# name, as pflag.Duration("x", ...) or fs.StringVar(&v, "x", ...)), every
# ATE_* environment variable it reads, and the preview gates its --preview
# accepts. A binary refuses a flag or gate it does not know and ignores a
# variable it no longer reads, so a release claiming an earlier shape must keep
# every one; one it adds is never set by the module and changes nothing an
# install does.
FLAG_REGISTRATION = re.compile(
    r'\b(?:pflag|flag|fs|flags|\w+\.(?:Persistent)?Flags\(\))\.\w+\(\s*(?:&[\w.\[\]]+\s*,\s*)?"([a-z0-9][a-z0-9-]*)"'
)
# The environment variables a binary reads its settings from.
ENVIRONMENT_VARIABLE = re.compile(r'"(ATE_[A-Z0-9_]+)"')
# A helper handed a flag set and a prefix, registering the flags under it.
FLAG_PREFIX = re.compile(r'\w+\(\s*(?:pflag\.CommandLine|fs|flags)\s*,\s*"([a-z0-9][a-z0-9-]*)"')
PREVIEW_GATES_FILE = "internal/preview/preview.go"
PREVIEW_GATE = re.compile(r'^\s*Gate\w+\s+Gate\s*=\s*"([^"]+)"', re.MULTILINE)

# The files of a published tree a shape is compared over: everything an
# install reads. images.json and SHA256SUMS describe the release itself, and
# the licenses are not installed.
SHAPE_EXEMPT_FILES = ("images.json", "SHA256SUMS", "LICENSE", "third-party-licenses.tar.gz")

# Upstream's own pattern for a whole ko:// reference (images/prebuilt.go,
# koRef): matching the whole reference is what keeps ko://.../cmd/atelet from
# rewriting a prefix of ko://.../cmd/atelet-sidecar, and requiring a character
# after the prefix leaves prose such as "# ko:// image reference" alone.
KO_REFERENCE = re.compile(r"""ko://[^\s"',\]}]+""")

ENVOY_PLACEHOLDER = "${ENVOY_DATAPLANE_IMAGE}"

# The platforms upstream's ko config and its Envoy build both target
# (.ko.yaml defaultPlatforms). An image missing one would schedule on some
# nodes and fail to pull on the rest.
REQUIRED_PLATFORMS = ("linux/amd64", "linux/arm64")

INDEX_MEDIA_TYPES = (
    "application/vnd.oci.image.index.v1+json",
    "application/vnd.docker.distribution.manifest.list.v2+json",
)

DIGEST = re.compile(r"^sha256:[0-9a-f]{64}$")

# Fixed archive timestamp: the archive's bytes depend only on its contents.
ARCHIVE_MTIME = 0


class CompositionError(Exception):
    """One or more findings, each stated as observed, meaning, next step."""

    def __init__(self, findings):
        super().__init__("\n\n".join(findings))
        self.findings = findings


def finding(observed: str, means: str, next_step: str) -> str:
    return f"Observed: {observed}\nMeans:    {means}\nNext:     {next_step}"


@dataclass(frozen=True)
class Upstream:
    """What the upstream checkout says about its images."""

    # ko:// reference -> image name, for every package in Components: the
    # rewrite table, which names unreleased packages too.
    image_by_reference: dict
    # The image names of the packages a release publishes (`released`).
    released_images: frozenset
    # The name the Dockerfile-built Envoy image is published under.
    envoy_image: str
    # The packages a release publishes, relative to the checkout: the binaries
    # an install runs.
    packages: tuple

    @property
    def image_names(self) -> set:
        """Every image a release publishes: the ko-built ones and Envoy."""
        return set(self.released_images) | {self.envoy_image}


def _string_list(text: str, declaration: str) -> list:
    """The quoted entries of one `var <name> = []string{...}` block, or []."""
    block = re.search(re.escape(declaration) + r" = \[\]string\{(.*?)\n\}", text, re.DOTALL)
    return re.findall(r'^\s*"([^"]+)",', block.group(1), re.MULTILINE) if block else []


def read_upstream(source: str) -> Upstream:
    """Read the module path, the component lists and the Envoy image name from the tag."""
    path = os.path.join(source, COMPONENTS_FILE)
    findings = []
    try:
        with open(path, encoding="utf-8") as f:
            text = f.read()
    except OSError:
        text = ""

    match = re.search(r'^const ModulePath = "([^"]+)"', text, re.MULTILINE)
    module_path = match.group(1) if match else ""
    components = _string_list(text, "var Components")
    released = _string_list(text, "var released")
    # The Envoy image is a DockerfileImage, and a release publishes it only
    # when releasedDockerfiles names the variable that declares it.
    envoy = re.search(r'^var (\w+) = DockerfileImage\{Name: "([^"]+)"', text, re.MULTILINE)
    dockerfiles = re.search(r"^var releasedDockerfiles = \[\]DockerfileImage\{([^}]*)\}", text, re.MULTILINE)
    envoy_image = envoy.group(2) if envoy and dockerfiles and envoy.group(1) in re.findall(r"\w+", dockerfiles.group(1)) else ""

    if not module_path:
        findings.append(finding(
            f"no `const ModulePath = \"...\"` in {COMPONENTS_FILE}",
            "upstream moved the module path its ko:// references start with, so no "
            "reference can be matched to an image",
            f"read {COMPONENTS_FILE} at this tag and update read_upstream to the new shape "
            "in the change that adopts this version",
        ))
    if not components:
        findings.append(finding(
            f"no `var Components = []string{{...}}` package list in {COMPONENTS_FILE}",
            "upstream moved or reshaped the list of packages that have images, "
            "so the composer cannot tell a known reference from an unknown one",
            f"read {COMPONENTS_FILE} and images/prebuilt.go at this tag and update "
            "read_upstream to the new shape in the change that adopts this version",
        ))
    if not released or not set(released) <= set(components):
        findings.append(finding(
            f"no `var released = []string{{...}}` allowlist within Components in {COMPONENTS_FILE}",
            "upstream moved or reshaped the list of packages a release publishes, "
            "so the composer cannot tell which images exist in its registry",
            f"read {COMPONENTS_FILE} at this tag and update read_upstream in the change "
            "that adopts this version",
        ))
    if not envoy_image:
        findings.append(finding(
            f"no released `DockerfileImage{{Name: \"...\"}}` in {COMPONENTS_FILE}",
            f"upstream moved the image its Envoy dataplane is published under, "
            f"so {ENVOY_PLACEHOLDER} cannot be filled",
            f"read {COMPONENTS_FILE} at this tag and update read_upstream in the change "
            "that adopts this version",
        ))
    if findings:
        raise CompositionError(findings)
    # images.ImageName: the last element of the import path, which is what
    # ko's --base-import-paths publishes it as.
    image_of = {pkg: pkg.rsplit("/", 1)[-1].lower() for pkg in components}
    return Upstream(
        image_by_reference={f"ko://{module_path}/{pkg}": image_of[pkg] for pkg in components},
        released_images=frozenset(image_of[pkg] for pkg in released),
        envoy_image=envoy_image,
        packages=tuple(released),
    )


def read_index_digests(index_manifests: str) -> dict:
    """Map each image name to the digest of its multi-platform index.

    The directory holds one <image>.json per image: the index exactly as the
    registry serves it for the release's tag. The digest is the sha256 of those
    bytes, so it is what the registry, and every registry the image is copied
    to, answers for the image; the index must name both platforms upstream
    builds.
    """
    digests = {}
    findings = []
    for name in sorted(os.listdir(index_manifests)):
        if not name.endswith(".json"):
            continue
        image = name[: -len(".json")]
        with open(os.path.join(index_manifests, name), "rb") as f:
            blob = f.read()
        try:
            manifest = json.loads(blob)
        except ValueError as err:
            findings.append(finding(
                f"the index fetched for {image} is not JSON ({err})",
                "the fetch saved something other than the image's index",
                f"fetch it again with `crane manifest <repository>/{image}:<version>`",
            ))
            continue
        platforms = {
            f"{m.get('platform', {}).get('os')}/{m.get('platform', {}).get('architecture')}"
            for m in manifest.get("manifests", [])
        }
        missing = [p for p in REQUIRED_PLATFORMS if p not in platforms]
        if manifest.get("mediaType") not in INDEX_MEDIA_TYPES or missing:
            findings.append(finding(
                f"{image} is {manifest.get('mediaType')} with platforms {sorted(platforms)}",
                f"the image lacks {', '.join(missing) or 'a multi-platform index'}, so it cannot "
                "run on every node an adopter's cluster may schedule it to",
                "check upstream's release workflow for this image; Planton publishes no release "
                "whose images miss a platform",
            ))
            continue
        digests[image] = "sha256:" + hashlib.sha256(blob).hexdigest()
    if findings:
        raise CompositionError(findings)
    return digests


def pinned_references(upstream: Upstream, digests: dict, repository: str, version: str) -> dict:
    """Every published image's reference, after checking the set is exactly upstream's."""
    expected = upstream.image_names
    built = set(digests)
    findings = []
    if expected - built:
        findings.append(finding(
            f"no index fetched for {', '.join(sorted(expected - built))}",
            "upstream's release publishes these images, and an install pulls them",
            "fetch each one's index from the repository at the release's tag",
        ))
    if built - expected:
        findings.append(finding(
            f"fetched images {', '.join(sorted(built - expected))} are not in upstream's published set",
            "the fetch read an image a release does not publish, so the set would name it",
            f"fetch exactly the images {COMPONENTS_FILE} releases (`released` and `releasedDockerfiles`)",
        ))
    if findings:
        raise CompositionError(findings)
    return {name: f"{repository}/{name}:{version}@{digests[name]}" for name in sorted(expected)}


def rewrite(text: str, upstream: Upstream, references: dict, path: str, unknown: dict) -> str:
    """Fill every ko:// reference and the Envoy placeholder in one file's text."""

    def replace(match):
        image = upstream.image_by_reference.get(match.group(0))
        if image is None:
            unknown.setdefault(match.group(0), []).append(path)
            return match.group(0)
        return references[image]

    return KO_REFERENCE.sub(replace, text).replace(ENVOY_PLACEHOLDER, references[upstream.envoy_image])


def split_documents(text: str) -> list:
    """Split a YAML stream on its document separators, keeping each document's text."""
    documents, current = [], []
    for line in text.splitlines(keepends=True):
        if re.match(r"^---\s*$", line):
            documents.append("".join(current))
            current = []
        else:
            current.append(line)
    documents.append("".join(current))
    return [d for d in documents if d.strip()]


def is_crd(document: str) -> bool:
    return re.search(r"^kind:\s*CustomResourceDefinition\s*$", document, re.MULTILINE) is not None


def deterministic_tar_gz(source_dir: str, arcname: str, destination: str) -> None:
    """Archive a directory so the bytes depend only on the files' names and contents."""
    buffer = io.BytesIO()
    with tarfile.open(fileobj=buffer, mode="w", format=tarfile.PAX_FORMAT) as tar:
        paths = [source_dir]
        for root, dirs, files in os.walk(source_dir):
            dirs.sort()
            paths.extend(os.path.join(root, d) for d in dirs)
            paths.extend(os.path.join(root, f) for f in sorted(files))
        for path in sorted(paths):
            relative = os.path.relpath(path, source_dir)
            info = tar.gettarinfo(path, arcname=arcname if relative == "." else f"{arcname}/{relative}")
            info.mtime = ARCHIVE_MTIME
            info.uid = info.gid = 0
            info.uname = info.gname = ""
            info.mode = 0o755 if info.isdir() else 0o644
            if info.isfile():
                with open(path, "rb") as f:
                    tar.addfile(info, f)
            else:
                tar.addfile(info)
    with open(destination, "wb") as f:
        with gzip.GzipFile(fileobj=f, mode="wb", mtime=ARCHIVE_MTIME, filename="") as gz:
            gz.write(buffer.getvalue())


def unreleased_references(text: str, upstream: Upstream) -> list:
    """The ko:// references in text that name a Component no release publishes."""
    return sorted({
        reference for reference in KO_REFERENCE.findall(text)
        if reference in upstream.image_by_reference
        and upstream.image_by_reference[reference] not in upstream.released_images
    })


def compose(source: str, index_manifests: str, repository: str, version: str, shape: str, commit: str, out: str) -> dict:
    """Write the release set to out and return the published image references."""
    upstream = read_upstream(source)
    references = pinned_references(upstream, read_index_digests(index_manifests), repository, version)

    manifests_source = os.path.join(source, "manifests")
    manifests_out = os.path.join(out, "manifests")
    texts = {}
    for root, dirs, files in os.walk(manifests_source):
        dirs.sort()
        for name in sorted(files):
            source_path = os.path.join(root, name)
            with open(source_path, encoding="utf-8") as f:
                texts[os.path.relpath(source_path, manifests_source)] = f.read()

    # A kustomize component that names an image no release publishes is left
    # out whole (in v1.0.0, the benchmarking fakes): none of it installs from a
    # release, and upstream's installer layers it only when asked. A kept
    # kustomization that layers one fails the check's build. Any other file
    # naming such an image is a finding, so an install file is never dropped.
    def component_of(relative: str) -> str:
        directory = os.path.dirname(relative)
        text = texts.get(os.path.join(directory, "kustomization.yaml"), "")
        return directory if re.search(r"^kind:\s*Component\s*$", text, re.MULTILINE) else ""

    naming_unreleased = {relative: unreleased_references(text, upstream) for relative, text in texts.items()}
    omitted = sorted({component_of(relative) for relative, refs in naming_unreleased.items() if refs} - {""})

    def kept(relative: str) -> bool:
        return not any(relative.startswith(d + os.sep) for d in omitted)

    stray = {relative: refs for relative, refs in naming_unreleased.items() if refs and kept(relative)}
    if stray:
        raise CompositionError([
            finding(
                f"manifests/{relative} names {', '.join(refs)}, which no release publishes",
                "the file is part of the install, not an optional component, so an install from a "
                "release would fail to pull these images",
                "read the file at this tag: if upstream means it as optional, it belongs in a kustomize "
                "component; otherwise report the reference upstream",
            )
            for relative, refs in sorted(stray.items())
        ])

    unknown = {}
    crds = []
    for relative, text in texts.items():
        if not kept(relative):
            continue
        destination = os.path.join(manifests_out, relative)
        os.makedirs(os.path.dirname(destination), exist_ok=True)
        text = rewrite(text, upstream, references, f"manifests/{relative}", unknown)
        with open(destination, "w", encoding="utf-8") as f:
            f.write(text)
        if relative.endswith((".yaml", ".yml")):
            crds.extend(d for d in split_documents(text) if is_crd(d))
    for directory in omitted:
        print(f"left out manifests/{directory}: it names images no release publishes", file=sys.stderr)

    findings = [
        finding(
            f"{reference} (in {', '.join(paths)}) has no published image",
            f"upstream has images only for the packages in {COMPONENTS_FILE}'s Components; "
            "this reference names one outside it, so no install could pull it",
            "if the manifest is templated, it is rendered before install and must be left out of "
            "the rewrite; otherwise upstream's tag is broken, so report it upstream",
        )
        for reference, paths in sorted(unknown.items())
    ]
    if not crds:
        findings.append(finding(
            "no CustomResourceDefinition in the composed manifests",
            "upstream moved its CRDs out of manifests/, so the kept-CRD bundle would be empty",
            "find where this tag keeps its CRDs and compose crds.yaml from there",
        ))
    if findings:
        raise CompositionError(findings)

    with open(os.path.join(out, "crds.yaml"), "w", encoding="utf-8") as f:
        f.write("".join("---\n" + (d if d.endswith("\n") else d + "\n") for d in crds))

    with open(os.path.join(out, "images.json"), "w", encoding="utf-8") as f:
        json.dump(
            {
                "version": version,
                "shape": shape,
                "source": {"repository": UPSTREAM_REPOSITORY, "tag": version, "commit": commit},
                "images": references,
            },
            f,
            indent=2,
            sort_keys=True,
        )
        f.write("\n")

    shutil.copyfile(os.path.join(source, SETUP_SQL_FILE), os.path.join(out, "setup.sql"))
    shutil.copyfile(os.path.join(source, "LICENSE"), os.path.join(out, "LICENSE"))
    deterministic_tar_gz(os.path.join(source, "_LICENSES"), "_LICENSES", os.path.join(out, "third-party-licenses.tar.gz"))

    write_checksums(out)
    return references


def write_checksums(out: str) -> None:
    """SHA256SUMS over every file in the set, in `sha256sum -c` format."""
    lines = []
    for root, dirs, files in os.walk(out):
        dirs.sort()
        for name in sorted(files):
            path = os.path.join(root, name)
            relative = os.path.relpath(path, out)
            if relative == "SHA256SUMS":
                continue
            with open(path, "rb") as f:
                lines.append(f"{hashlib.sha256(f.read()).hexdigest()}  {relative}\n")
    with open(os.path.join(out, "SHA256SUMS"), "w", encoding="utf-8") as f:
        f.writelines(sorted(lines, key=lambda line: line.split("  ", 1)[1]))


def check(out: str, repository: str, references: dict, kustomize: str) -> None:
    """Prove the composed tree still works as upstream's installer uses it.

    No ko:// reference or Envoy placeholder survives; every kustomization
    builds (components are layered by the installer, never built alone); and
    every reference under the repository, in a file or in a build's output, is
    exactly one this release publishes.
    """
    published = set(references.values())
    ours = re.compile(re.escape(repository) + r"""/[^\s"',\]}]+""")
    findings = []

    def scan(text: str, where: str) -> None:
        if KO_REFERENCE.search(text):
            findings.append(finding(
                f"{where} still names {KO_REFERENCE.search(text).group(0)}",
                "a ko:// reference survived the rewrite, so an install would fail to pull it",
                "fix the composer's rewrite before publishing",
            ))
        if ENVOY_PLACEHOLDER in text:
            findings.append(finding(
                f"{where} still carries {ENVOY_PLACEHOLDER}",
                "a workload that runs Envoy would deploy with no image",
                "fix the composer's rewrite before publishing",
            ))
        for reference in sorted(set(ours.findall(text)) - published):
            findings.append(finding(
                f"{where} names {reference}",
                "that reference is not one of the images this release publishes",
                "fix the composer's rewrite before publishing",
            ))

    manifests = os.path.join(out, "manifests")
    kustomizations = []
    for root, dirs, files in os.walk(manifests):
        dirs.sort()
        for name in sorted(files):
            path = os.path.join(root, name)
            with open(path, encoding="utf-8") as f:
                text = f.read()
            scan(text, os.path.relpath(path, out))
            if name == "kustomization.yaml" and re.search(r"^kind:\s*Kustomization\s*$", text, re.MULTILINE):
                kustomizations.append(root)

    for directory in kustomizations:
        where = os.path.relpath(directory, out)
        # Upstream's overlays name files beside their directory (../atelet.yaml),
        # which kustomize refuses by default; upstream's installer builds them
        # with load restrictions off (cmd/ate-setup/internal/kustomize), and so
        # does this check.
        result = subprocess.run(
            [kustomize, "build", "--load-restrictor", "LoadRestrictionsNone", directory],
            capture_output=True,
            text=True,
        )
        if result.returncode != 0:
            findings.append(finding(
                f"`kustomize build {where}` failed: {result.stderr.strip()}",
                "upstream's installer could not render this overlay from the published tree",
                "compare the composed tree with upstream's at this tag; the rewrite must change "
                "nothing but image references",
            ))
            continue
        scan(result.stdout, f"`kustomize build {where}`")

    if not kustomizations:
        findings.append(finding(
            "no kustomization in the composed manifests",
            "upstream reshaped its install tree, so nothing proved the overlays still build",
            "read manifests/ at this tag and update the check",
        ))
    if findings:
        raise CompositionError(findings)


def installed_files(tree: str) -> dict:
    """Every file of a published tree an install reads, by relative path."""
    files = {}
    for root, dirs, names in os.walk(tree):
        dirs.sort()
        for name in sorted(names):
            path = os.path.join(root, name)
            relative = os.path.relpath(path, tree)
            if relative not in SHAPE_EXEMPT_FILES:
                files[relative] = path
    return files


def source_files(source: str) -> dict:
    """The upstream files under SHAPE_SOURCE_PATHS, by relative path.

    A path the checkout lacks is a finding rather than an empty comparison:
    an install-deciding directory that moved would otherwise pass unseen.
    """
    missing = [top for top in SHAPE_SOURCE_PATHS if not os.path.isdir(os.path.join(source, top))]
    if missing:
        raise CompositionError([finding(
            f"the upstream checkout at {source} has no {', '.join(missing)}",
            "upstream moved code that decides how an install is assembled, so the shape "
            "comparison would skip it",
            "find where it moved at this tag and update SHAPE_SOURCE_PATHS in the change that adopts it",
        )])
    files = {}
    for top in SHAPE_SOURCE_PATHS:
        for root, dirs, names in os.walk(os.path.join(source, top)):
            dirs.sort()
            for name in sorted(names):
                path = os.path.join(root, name)
                files[os.path.relpath(path, source)] = path
    return files


def settings(source: str) -> set:
    """The flags and environment variables the installed binaries read, and the preview gates."""
    found = set()
    for package in read_upstream(source).packages:
        for root, dirs, names in os.walk(os.path.join(source, package)):
            for name in names:
                if name.endswith(".go") and not name.endswith("_test.go"):
                    with open(os.path.join(root, name), encoding="utf-8") as f:
                        text = f.read()
                    found |= {f"{package}: --{flag}" for flag in FLAG_REGISTRATION.findall(text)}
                    found |= {f"{package}: --{prefix}-*" for prefix in FLAG_PREFIX.findall(text)}
                    found |= {f"{package}: ${variable}" for variable in ENVIRONMENT_VARIABLE.findall(text)}
    try:
        with open(os.path.join(source, PREVIEW_GATES_FILE), encoding="utf-8") as f:
            found |= {f"preview gate {gate}" for gate in PREVIEW_GATE.findall(f.read())}
    except OSError:
        pass
    return found


def differing(ours: dict, theirs: dict, read) -> tuple:
    """The paths only one side has, and the shared paths whose contents differ."""
    only = sorted(set(ours) ^ set(theirs))
    changed = [path for path in sorted(set(ours) & set(theirs)) if read(ours[path]) != read(theirs[path])]
    return only, changed


def check_shape(out: str, repository: str, version: str, shape: str, shape_release: str, source: str, shape_source: str) -> None:
    """Prove a release claiming an earlier shape installs exactly as that shape does.

    The composed tree must equal the shape's published tree file for file once
    every image reference is reduced to its name (the digests and the tag are
    the release's own), the upstream code under SHAPE_SOURCE_PATHS must be
    byte-identical at both tags, and every setting the shape's binaries accept
    must still be accepted.
    """
    pinned = re.compile(re.escape(repository) + r"""/([^\s:@"',\]}]+):[^\s@"',\]}]+@sha256:[0-9a-f]{64}""")

    def reduced(path: str) -> str:
        with open(path, encoding="utf-8") as f:
            return pinned.sub(lambda m: f"{repository}/{m.group(1)}", f.read())

    def raw(path: str) -> bytes:
        with open(path, "rb") as f:
            return f.read()

    findings = []
    only, changed = differing(installed_files(out), installed_files(shape_release), reduced)
    if only or changed:
        findings.append(finding(
            f"{version}'s install differs from {shape}'s beyond its images: "
            + "; ".join(filter(None, [
                only and f"present in one release only: {', '.join(only)}",
                changed and f"changed: {', '.join(changed)}",
            ])),
            f"the control-plane module composes {shape}'s install, so it cannot be trusted to install {version}",
            f"declare {version} as its own shape (shape: {version} in tools/ci/release/agent-substrate.yaml) "
            "and adapt the module to it before customers can set it",
        ))
    only, changed = differing(source_files(source), source_files(shape_source), raw)
    if only or changed:
        findings.append(finding(
            f"upstream changed how an install is assembled between {shape} and {version}: "
            + ", ".join(only + changed),
            "the installer's steps, settings or Google Cloud setup moved, which the module mirrors and the "
            "published tree does not show",
            f"read the change; then declare {version} as its own shape and adapt the module to it",
        ))
    removed = sorted(settings(shape_source) - settings(source))
    if removed:
        findings.append(finding(
            f"{version} no longer accepts settings {shape} accepted: {', '.join(removed)}",
            "the module passes upstream's flags and preview gates by name, and a binary refuses one it "
            "does not know, so the module's install of this release would not start",
            f"declare {version} as its own shape and adapt the module to the settings it accepts",
        ))
    if findings:
        raise CompositionError(findings)


def main() -> None:
    if sys.argv[1:2] == ["images"]:
        # `images --source <checkout>`: the names of the images a release of
        # this tag publishes, one per line, for the lane to fetch the indexes of.
        parser = argparse.ArgumentParser(prog="compose_agent_substrate_release.py images")
        parser.add_argument("--source", required=True, help="upstream checkout at the release tag")
        args = parser.parse_args(sys.argv[2:])
        try:
            print("\n".join(sorted(read_upstream(args.source).image_names)))
        except CompositionError as err:
            print(err, file=sys.stderr)
            sys.exit(1)
        return

    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument("--source", required=True, help="upstream checkout at the release tag")
    parser.add_argument("--index-manifests", required=True, help="directory holding <image>.json: each image's multi-platform index as its registry serves it")
    parser.add_argument("--repository", required=True, help="registry path the images are published under")
    parser.add_argument("--version", required=True, help="the upstream release tag, vX.Y.Z")
    parser.add_argument("--shape", required=True, help="the release whose install this one shares; equal to --version for a new shape")
    parser.add_argument("--shape-release", help="the shape's published tree; required when --shape differs from --version")
    parser.add_argument("--shape-source", help="upstream checkout at the shape's tag; required when --shape differs from --version")
    parser.add_argument("--commit", required=True, help="the commit the tag names")
    parser.add_argument("--out", required=True, help="empty directory to write the release set into")
    parser.add_argument("--kustomize", help="kustomize binary; when set, the composed tree is checked")
    args = parser.parse_args()

    if args.kustomize is not None and not args.kustomize:
        print("Error: --kustomize names no binary; pass the kustomize to check the composed tree with, or leave the flag out", file=sys.stderr)
        sys.exit(2)
    if os.path.exists(args.out) and os.listdir(args.out):
        print(f"Error: --out {args.out} is not empty; compose into a fresh directory", file=sys.stderr)
        sys.exit(2)
    claims_earlier_shape = args.shape != args.version
    if claims_earlier_shape and not (args.shape_release and args.shape_source):
        print(f"Error: {args.version} claims shape {args.shape}; pass --shape-release and --shape-source to prove it", file=sys.stderr)
        sys.exit(2)
    os.makedirs(args.out, exist_ok=True)
    try:
        references = compose(args.source, args.index_manifests, args.repository, args.version, args.shape, args.commit, args.out)
        if args.kustomize:
            check(args.out, args.repository, references, args.kustomize)
        if claims_earlier_shape:
            check_shape(args.out, args.repository, args.version, args.shape, args.shape_release, args.source, args.shape_source)
    except CompositionError as err:
        print(f"Composing Agent Substrate {args.version} failed:\n", file=sys.stderr)
        print(err, file=sys.stderr)
        sys.exit(1)
    for name, reference in references.items():
        print(f"{name}: {reference}")


if __name__ == "__main__":
    main()

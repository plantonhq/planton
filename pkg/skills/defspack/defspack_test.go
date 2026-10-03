package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestCommittedTreeValidates is the lint gate's core: the real committed
// skills/ and agents/ trees must satisfy the structure contract at all
// times. A truncated reference, an orphaned file, or a frontmatter/slug
// mismatch in a pull request fails here with the exact repair message.
func TestCommittedTreeValidates(t *testing.T) {
	root := repoRoot(t)
	if _, err := os.Stat(filepath.Join(root, "skills")); err != nil {
		// Under Bazel the sandbox carries only declared inputs, not the
		// repository tree; the enforcing lane for this gate is `go test`
		// from the repo root (the skills lint workflow), same as the
		// reference drift gate.
		t.Skipf("skills tree not present at %s (sandboxed run)", root)
	}
	tree, err := LoadTree(root)
	if err != nil {
		t.Fatalf("loading committed tree: %v", err)
	}
	for _, err := range Validate(tree) {
		t.Errorf("committed tree invalid: %v", err)
	}
	if len(tree.Skills) == 0 {
		t.Fatal("no skills loaded from the committed tree")
	}
}

// TestValidateCatchesBrokenTrees proves every guarded failure class is
// actually caught (a validator that passes everything would also pass the
// committed tree). Each case plants exactly one defect in an otherwise
// valid tree and asserts the violation is reported.
func TestValidateCatchesBrokenTrees(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(files map[string]string)
		wantErr string
	}{
		{
			name: "cited reference file is empty",
			mutate: func(files map[string]string) {
				files["skills/demo/references/topic.md"] = ""
			},
			wantErr: "is empty",
		},
		{
			name: "cited reference file is missing",
			mutate: func(files map[string]string) {
				delete(files, "skills/demo/references/topic.md")
			},
			wantErr: "does not exist",
		},
		{
			name: "reference file exists but is never cited",
			mutate: func(files map[string]string) {
				files["skills/demo/references/orphan.md"] = "never cited"
			},
			wantErr: "never cites it",
		},
		{
			name: "frontmatter name does not match directory slug",
			mutate: func(files map[string]string) {
				files["skills/demo/SKILL.md"] = strings.Replace(
					files["skills/demo/SKILL.md"], "name: demo", "name: other", 1)
			},
			wantErr: "must equal the directory slug",
		},
		{
			name: "agent instructions are empty",
			mutate: func(files map[string]string) {
				files["agents/helper/instructions.md"] = ""
			},
			wantErr: "instructions.md is missing or empty",
		},
		{
			name: "compat floor is not a version",
			mutate: func(files map[string]string) {
				files["skills/compat.yaml"] = "minimum_daemon_version: soon\nminimum_cli_version: v0.0.0\n"
			},
			wantErr: "is not a vX.Y.Z version",
		},
		{
			name: "automation document slug does not match its file name",
			mutate: func(files map[string]string) {
				files["automations/investigate.yaml"] = "slug: other\ndisplayName: Investigate\n"
			},
			wantErr: "must equal the file name",
		},
		{
			name: "automation file is not valid yaml",
			mutate: func(files map[string]string) {
				files["automations/investigate.yaml"] = "slug: [unclosed\n"
			},
			wantErr: "not valid YAML",
		},
		{
			name: "automation file is empty",
			mutate: func(files map[string]string) {
				files["automations/investigate.yaml"] = ""
			},
			wantErr: "file is empty",
		},
		{
			name: "description exceeds the spec's character cap",
			mutate: func(files map[string]string) {
				files["skills/demo/SKILL.md"] = strings.Replace(
					files["skills/demo/SKILL.md"],
					"description: A demo skill.",
					"description: "+strings.Repeat("wordy ", 200), 1)
			},
			wantErr: "the Agent Skills spec caps it at 1024",
		},
		{
			name: "frontmatter carries a key outside the spec's field set",
			mutate: func(files map[string]string) {
				files["skills/demo/SKILL.md"] = strings.Replace(
					files["skills/demo/SKILL.md"],
					"name: demo\n",
					"name: demo\nowner: someone\n", 1)
			},
			wantErr: "outside the Agent Skills spec's field set",
		},
		{
			name: "SKILL.md body exceeds the line ceiling",
			mutate: func(files map[string]string) {
				files["skills/demo/SKILL.md"] += strings.Repeat("filler line\n", 501)
			},
			wantErr: "the spec's ceiling is 500",
		},
		{
			name: "a directory under references is refused, never silently dropped",
			mutate: func(files map[string]string) {
				files["skills/demo/references/nested/deep.md"] = "invisible to packaging\n"
			},
			wantErr: "references are FLAT files",
		},
		{
			name: "reference filename breaks the grammar",
			mutate: func(files map[string]string) {
				files["skills/demo/references/Bad_Name.md"] = "content\n"
				files["skills/demo/SKILL.md"] = strings.Replace(
					files["skills/demo/SKILL.md"],
					"Read `references/topic.md` first.",
					"Read `references/topic.md` and `references/Bad_Name.md` first.", 1)
			},
			wantErr: "breaks the reference-name grammar",
		},
		{
			name: "skill directory slug breaks the spec's name grammar",
			mutate: func(files map[string]string) {
				files["skills/Demo-Two/SKILL.md"] = "---\nname: Demo-Two\ndescription: Bad name.\n---\n\n# Bad\n"
			},
			wantErr: "breaks the Agent Skills spec's grammar",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			files := validFixtureTree()
			tc.mutate(files)
			root := writeTree(t, files)

			tree, err := LoadTree(root)
			if err != nil {
				t.Fatalf("loading fixture tree: %v", err)
			}
			errs := Validate(tree)
			if len(errs) == 0 {
				t.Fatalf("planted defect was not caught")
			}
			for _, e := range errs {
				if strings.Contains(e.Error(), tc.wantErr) {
					return
				}
			}
			t.Fatalf("no violation mentions %q; got: %v", tc.wantErr, errs)
		})
	}

	// The unmutated fixture must be clean, or the cases above prove nothing.
	t.Run("unmutated fixture is valid", func(t *testing.T) {
		tree, err := LoadTree(writeTree(t, validFixtureTree()))
		if err != nil {
			t.Fatalf("loading fixture tree: %v", err)
		}
		if errs := Validate(tree); len(errs) > 0 {
			t.Fatalf("fixture tree should be valid, got: %v", errs)
		}
	})
}

// TestBuildSkillArchiveDeterministic holds the archive to its contract:
// byte-identical across builds, stored entries only, descriptor-free
// (CreateRaw), SKILL.md at the archive root.
func TestBuildSkillArchiveDeterministic(t *testing.T) {
	tree, err := LoadTree(writeTree(t, validFixtureTree()))
	if err != nil {
		t.Fatalf("loading fixture tree: %v", err)
	}
	skill := tree.Skills[0]

	first, err := BuildSkillArchive(skill)
	if err != nil {
		t.Fatalf("first build: %v", err)
	}
	second, err := BuildSkillArchive(skill)
	if err != nil {
		t.Fatalf("second build: %v", err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("two builds of identical content produced different bytes")
	}

	reader, err := zip.NewReader(bytes.NewReader(first), int64(len(first)))
	if err != nil {
		t.Fatalf("reading archive back: %v", err)
	}
	var names []string
	for _, f := range reader.File {
		names = append(names, f.Name)
		if f.Method != zip.Store {
			t.Errorf("%s: entry is not stored", f.Name)
		}
		if f.Flags&0x8 != 0 {
			t.Errorf("%s: entry carries a data descriptor", f.Name)
		}
	}
	if names[0] != "SKILL.md" {
		t.Errorf("first entry is %s, want SKILL.md at the archive root", names[0])
	}
}

// TestPackageReleaseManifest packages the fixture tree and verifies the
// manifest's integrity claims against the files actually written: every
// artifact exists, its size matches, and its SHA-256 matches.
func TestPackageReleaseManifest(t *testing.T) {
	tree, err := LoadTree(writeTree(t, validFixtureTree()))
	if err != nil {
		t.Fatalf("loading fixture tree: %v", err)
	}
	outDir := t.TempDir()

	manifest, err := PackageRelease(tree, "v9.9.9", outDir)
	if err != nil {
		t.Fatalf("packaging: %v", err)
	}
	if manifest.Version != "v9.9.9" {
		t.Errorf("manifest version = %q", manifest.Version)
	}
	if manifest.Floor.MinimumDaemonVersion != "v0.0.0" {
		t.Errorf("manifest floor = %+v, want the fixture's compat.yaml values", manifest.Floor)
	}

	var reread Manifest
	raw, err := os.ReadFile(filepath.Join(outDir, ManifestFileName))
	if err != nil {
		t.Fatalf("manifest not written: %v", err)
	}
	if err := json.Unmarshal(raw, &reread); err != nil {
		t.Fatalf("manifest does not parse: %v", err)
	}

	artifacts := append(append(reread.Skills, reread.Agents...), reread.Automations...)
	for _, artifact := range artifacts {
		content, err := os.ReadFile(filepath.Join(outDir, artifact.File))
		if err != nil {
			t.Errorf("%s: manifest lists a file that was not written: %v", artifact.File, err)
			continue
		}
		if int64(len(content)) != artifact.SizeBytes {
			t.Errorf("%s: size %d does not match manifest's %d", artifact.File, len(content), artifact.SizeBytes)
		}
		sum := sha256.Sum256(content)
		if hex.EncodeToString(sum[:]) != artifact.Sha256 {
			t.Errorf("%s: checksum does not match manifest", artifact.File)
		}
	}
	if len(reread.Skills) != 1 || len(reread.Agents) != 1 || len(reread.Automations) != 1 {
		t.Errorf("manifest lists %d skill(s), %d agent(s), %d automation(s), want 1 of each",
			len(reread.Skills), len(reread.Agents), len(reread.Automations))
	}
	if reread.Automations[0].File != "automation-investigate.yaml" {
		t.Errorf("automation artifact file = %q, want the automation-<slug>.yaml contract", reread.Automations[0].File)
	}
}

// TestManifestOmitsEmptyAutomations pins the compatibility posture: a
// release carrying zero automations writes a manifest with NO automations
// field at all (older-shaped, byte-compatible), never an empty array.
func TestManifestOmitsEmptyAutomations(t *testing.T) {
	files := validFixtureTree()
	delete(files, "automations/investigate.yaml")
	delete(files, "automations/README.md")
	tree, err := LoadTree(writeTree(t, files))
	if err != nil {
		t.Fatalf("loading fixture tree: %v", err)
	}
	outDir := t.TempDir()
	if _, err := PackageRelease(tree, "v9.9.9", outDir); err != nil {
		t.Fatalf("packaging: %v", err)
	}
	raw, err := os.ReadFile(filepath.Join(outDir, ManifestFileName))
	if err != nil {
		t.Fatalf("manifest not written: %v", err)
	}
	if strings.Contains(string(raw), "automations") {
		t.Fatalf("manifest of an automation-less release must omit the automations field, got:\n%s", raw)
	}
}

// TestBrowseManifestMatchesArchives holds the browse manifest to its one
// contract: it describes exactly the bytes the release ships. Both
// directions are asserted against the packaged output -- every archive
// entry is listed with its exact checksum and size (an exploded file IS
// the archive entry, which is what lets the release workflow derive the
// exploded tree by unzipping), no listed file is absent from its archive,
// and the agent/automation entries carry the integrity manifest's own
// checksums (they describe the same release-root files).
func TestBrowseManifestMatchesArchives(t *testing.T) {
	// The catalog fixture exercises nested pack paths, not just the flat
	// references/ layout -- the browse manifest must describe both.
	tree, err := LoadTree(writeTree(t, catalogFixtureTree()))
	if err != nil {
		t.Fatalf("loading fixture tree: %v", err)
	}
	outDir := t.TempDir()
	manifest, err := PackageRelease(tree, "v9.9.9", outDir)
	if err != nil {
		t.Fatalf("packaging: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(outDir, BrowseManifestFileName))
	if err != nil {
		t.Fatalf("browse manifest not written: %v", err)
	}
	var browse BrowseManifest
	if err := json.Unmarshal(raw, &browse); err != nil {
		t.Fatalf("browse manifest does not parse: %v", err)
	}
	if browse.Version != "v9.9.9" {
		t.Errorf("browse manifest version = %q", browse.Version)
	}
	if len(browse.Skills) != len(manifest.Skills) {
		t.Fatalf("browse manifest lists %d skill(s), integrity manifest %d", len(browse.Skills), len(manifest.Skills))
	}

	for _, skillBrowse := range browse.Skills {
		archiveBytes, err := os.ReadFile(filepath.Join(outDir, "skill-"+skillBrowse.Slug+".zip"))
		if err != nil {
			t.Fatalf("reading archive for %s: %v", skillBrowse.Slug, err)
		}
		reader, err := zip.NewReader(bytes.NewReader(archiveBytes), int64(len(archiveBytes)))
		if err != nil {
			t.Fatalf("opening archive for %s: %v", skillBrowse.Slug, err)
		}
		archiveEntries := map[string][]byte{}
		for _, f := range reader.File {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("%s/%s: opening entry: %v", skillBrowse.Slug, f.Name, err)
			}
			content, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				t.Fatalf("%s/%s: reading entry: %v", skillBrowse.Slug, f.Name, err)
			}
			archiveEntries[f.Name] = content
		}

		if len(skillBrowse.Files) != len(archiveEntries) {
			t.Errorf("%s: browse manifest lists %d file(s), archive carries %d",
				skillBrowse.Slug, len(skillBrowse.Files), len(archiveEntries))
		}
		for _, file := range skillBrowse.Files {
			content, ok := archiveEntries[file.Path]
			if !ok {
				t.Errorf("%s: browse manifest lists %s which the archive does not carry", skillBrowse.Slug, file.Path)
				continue
			}
			if int64(len(content)) != file.SizeBytes {
				t.Errorf("%s/%s: size %d does not match browse manifest's %d", skillBrowse.Slug, file.Path, len(content), file.SizeBytes)
			}
			sum := sha256.Sum256(content)
			if hex.EncodeToString(sum[:]) != file.Sha256 {
				t.Errorf("%s/%s: checksum does not match browse manifest", skillBrowse.Slug, file.Path)
			}
		}
	}

	// Agents and automations already ship individually at the release root;
	// the browse manifest points at those exact files, so its entries must
	// agree with the integrity manifest's checksums for them.
	assertRootFilesAgree := func(kind string, browseFiles []BrowseFile, artifacts []Artifact) {
		if len(browseFiles) != len(artifacts) {
			t.Errorf("%s: browse manifest lists %d, integrity manifest %d", kind, len(browseFiles), len(artifacts))
			return
		}
		for i, file := range browseFiles {
			if file.Path != artifacts[i].File {
				t.Errorf("%s[%d]: browse path %q != integrity file %q", kind, i, file.Path, artifacts[i].File)
			}
			if file.Sha256 != artifacts[i].Sha256 {
				t.Errorf("%s[%d]: browse checksum disagrees with the integrity manifest", kind, i)
			}
		}
	}
	assertRootFilesAgree("agents", browse.Agents, manifest.Agents)
	assertRootFilesAgree("automations", browse.Automations, manifest.Automations)
}

// TestBrowseManifestDeterministic: like the archives, the browse
// manifest's bytes are a pure function of the tree's content.
func TestBrowseManifestDeterministic(t *testing.T) {
	tree, err := LoadTree(writeTree(t, catalogFixtureTree()))
	if err != nil {
		t.Fatalf("loading fixture tree: %v", err)
	}
	first, err := EncodeBrowseManifest(BuildBrowseManifest(tree, "v9.9.9"))
	if err != nil {
		t.Fatalf("first build: %v", err)
	}
	second, err := EncodeBrowseManifest(BuildBrowseManifest(tree, "v9.9.9"))
	if err != nil {
		t.Fatalf("second build: %v", err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("two builds of identical content produced different browse manifest bytes")
	}
}

// TestBrowseManifestOmitsEmptyAutomations mirrors the integrity
// manifest's compatibility posture for the browse document.
func TestBrowseManifestOmitsEmptyAutomations(t *testing.T) {
	files := validFixtureTree()
	delete(files, "automations/investigate.yaml")
	delete(files, "automations/README.md")
	tree, err := LoadTree(writeTree(t, files))
	if err != nil {
		t.Fatalf("loading fixture tree: %v", err)
	}
	raw, err := EncodeBrowseManifest(BuildBrowseManifest(tree, "v9.9.9"))
	if err != nil {
		t.Fatalf("encoding: %v", err)
	}
	if strings.Contains(string(raw), "automations") {
		t.Fatalf("browse manifest of an automation-less release must omit the automations field, got:\n%s", raw)
	}
}

// TestExplodedLaneMirroredInReleaseWorkflow keeps this package and the
// release workflow in lockstep, in TestPackSelectionMirroredInReleaseLane's
// exact idiom: the workflow derives and verifies the exploded layout this
// package's browse manifest describes, so a shape change on either side
// without the other turns the release red here first.
func TestExplodedLaneMirroredInReleaseWorkflow(t *testing.T) {
	root := repoRoot(t)
	wfPath := filepath.Join(root, ".github", "workflows", "release.definitions.yaml")
	wf, err := os.ReadFile(wfPath)
	if err != nil {
		// Same sandbox posture as TestCommittedTreeValidates: the
		// enforcing lane is `go test` from the repo root.
		t.Skipf("release workflow not present at %s (sandboxed run)", wfPath)
	}
	text := string(wf)
	for want, teach := range map[string]string{
		"definitions-manifest.json|definitions-browse.json|": "the discover step's straggler whitelist must accept the browse manifest or every release fails at discovery",
		"\n  exploded:\n":   "the exploded job derives and uploads the browsable layout this package's browse manifest describes",
		"sha256sum --check": "the exploded job must verify every unzipped file against the browse manifest before uploading",
		"needs: [skill, agent, automation, exploded]": "publish-manifest must wait on the exploded upload, or the stable pointer could name a release whose browsable layout is missing",
		"definitions/releases/index.json":             "the releases index is the browse surfaces' release picker; it lives behind the same DAG guarantee",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("release.definitions.yaml is missing %q -- %s", want, teach)
		}
	}
}

// TestCatalogSkillAssemblesPack proves the catalog skill's archive is
// self-contained: the pack is collected from catalog/ by the frozen name
// contract (reference pages, guides, indexes, graph, commons, patterns),
// test fixtures are excluded, unrelated files are not swept in, and the
// assembled entries land under kinds/ with catalog-relative paths.
func TestCatalogSkillAssemblesPack(t *testing.T) {
	tree, err := LoadTree(writeTree(t, catalogFixtureTree()))
	if err != nil {
		t.Fatalf("loading fixture tree: %v", err)
	}
	if errs := Validate(tree); len(errs) > 0 {
		t.Fatalf("fixture should be valid, got: %v", errs)
	}

	var catalogSkill *Skill
	for i := range tree.Skills {
		if tree.Skills[i].Slug == catalogPackSkillSlug {
			catalogSkill = &tree.Skills[i]
		}
	}
	if catalogSkill == nil {
		t.Fatal("catalog skill not loaded")
	}

	wantPresent := []string{
		"kinds/_docs/reference-commons.md",
		"kinds/_docs/reference-index.md",
		"kinds/_docs/reference-graph.yaml",
		"kinds/_docs/GUIDE.md",
		"kinds/_patterns/observability.md",
		"kinds/aws/reference-index.md",
		"kinds/aws/awsvpc/v1alpha1/reference.md",
		"kinds/aws/awsvpc/GUIDE.md",
		// The fact-sheet layer: per-kind sidecars by name, the
		// central estimates and compliance trees by path.
		"kinds/aws/awsvpc/cost.yaml",
		"kinds/aws/awsvpc/controls.yaml",
		"kinds/aws/awsvpc/iac/permissions.yaml",
		"kinds/_pricing/estimates/awsvpc.yaml",
		"kinds/_compliance/controls-catalog.yaml",
		"kinds/_compliance/frameworks/cis-aws.yaml",
	}
	for _, path := range wantPresent {
		if _, ok := catalogSkill.PackFiles[path]; !ok {
			t.Errorf("pack is missing %s", path)
		}
	}
	wantAbsent := []string{
		"kinds/aws/awsvpc/v1alpha1/spec.proto", // not a pack file name
		"kinds/_test/fake/v1alpha1/reference.md",
		"kinds/aws/awsvpc/README.md",
		// Pricing-pipeline machinery never ships: agents read estimates,
		// engines read models/derivations/books.
		"kinds/_pricing/models/awsvpc.yaml",
		"kinds/_pricing/pricebook/aws.yaml",
		"kinds/_pricing/derivations/awsvpc.yaml",
		"kinds/_compliance/README.md", // central trees ship .yaml documents only
	}
	for _, path := range wantAbsent {
		if _, ok := catalogSkill.PackFiles[path]; ok {
			t.Errorf("pack wrongly includes %s", path)
		}
	}

	archive, err := BuildSkillArchive(*catalogSkill)
	if err != nil {
		t.Fatalf("building archive: %v", err)
	}
	reader, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		t.Fatalf("reading archive back: %v", err)
	}
	entries := map[string]bool{}
	for _, f := range reader.File {
		entries[f.Name] = true
	}
	for _, path := range append([]string{"SKILL.md", "references/pack-layout.md"}, wantPresent...) {
		if !entries[path] {
			t.Errorf("archive is missing entry %s", path)
		}
	}

	// Determinism must hold with the pack included.
	second, err := BuildSkillArchive(*catalogSkill)
	if err != nil {
		t.Fatalf("second build: %v", err)
	}
	if !bytes.Equal(archive, second) {
		t.Fatal("two builds of identical pack content produced different bytes")
	}
}

// TestStripPackFilesGatesPackaging pins the shipping gate: packaging
// without -embed-catalog-pack (the release lanes pass it) produces a
// catalog archive with no kinds/ entries, while the assembled build
// carries them. Validation always sees the pack either way (the strip
// happens after Validate).
func TestStripPackFilesGatesPackaging(t *testing.T) {
	tree, err := LoadTree(writeTree(t, catalogFixtureTree()))
	if err != nil {
		t.Fatalf("loading fixture tree: %v", err)
	}
	if errs := Validate(tree); len(errs) > 0 {
		t.Fatalf("fixture should be valid, got: %v", errs)
	}

	StripPackFiles(tree)
	for _, skill := range tree.Skills {
		if len(skill.PackFiles) != 0 {
			t.Fatalf("skill %s still carries %d pack files after the strip", skill.Slug, len(skill.PackFiles))
		}
	}
	var catalogSkill *Skill
	for i := range tree.Skills {
		if tree.Skills[i].Slug == catalogPackSkillSlug {
			catalogSkill = &tree.Skills[i]
		}
	}
	archive, err := BuildSkillArchive(*catalogSkill)
	if err != nil {
		t.Fatalf("building stripped archive: %v", err)
	}
	reader, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		t.Fatalf("reading archive back: %v", err)
	}
	for _, f := range reader.File {
		if strings.HasPrefix(f.Name, packDirName+"/") {
			t.Fatalf("stripped archive still carries pack entry %s", f.Name)
		}
	}
}

// TestValidateCatchesBrokenPacks plants one pack defect at a time in the
// catalog fixture and asserts the violation is reported -- an assembly
// that silently ships empty or misrooted would put a research skill with
// nothing to research on every engine.
func TestValidateCatchesBrokenPacks(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(files map[string]string)
		wantErr string
	}{
		{
			name: "catalog tree missing entirely",
			mutate: func(files map[string]string) {
				for path := range files {
					if strings.HasPrefix(path, "catalog/") {
						delete(files, path)
					}
				}
			},
			wantErr: "catalog pack is empty",
		},
		{
			name: "commons root marker missing",
			mutate: func(files map[string]string) {
				delete(files, "catalog/_docs/reference-commons.md")
			},
			wantErr: "missing its root marker",
		},
		{
			name: "no kind reference pages",
			mutate: func(files map[string]string) {
				delete(files, "catalog/aws/awsvpc/v1alpha1/reference.md")
			},
			wantErr: "no kind reference pages",
		},
		{
			name: "pack file is empty",
			mutate: func(files map[string]string) {
				files["catalog/aws/awsvpc/GUIDE.md"] = ""
			},
			wantErr: "is empty",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			files := catalogFixtureTree()
			tc.mutate(files)
			tree, err := LoadTree(writeTree(t, files))
			if err != nil {
				t.Fatalf("loading fixture tree: %v", err)
			}
			errs := Validate(tree)
			if len(errs) == 0 {
				t.Fatalf("planted pack defect was not caught")
			}
			for _, e := range errs {
				if strings.Contains(e.Error(), tc.wantErr) {
					return
				}
			}
			t.Fatalf("no violation mentions %q; got: %v", tc.wantErr, errs)
		})
	}
}

// TestPackSelectionMirroredInReleaseLane is the two-homes tripwire: the
// self-contained skill's pack assembly (this package) and the release
// content lane's reference-pack.zip (tools/ci/release/package_content.sh)
// are the same pack on two transports, so every selection token this
// package uses must appear in the script's find clause. This cannot parse
// shell, so it asserts token presence -- enough to make "extended one home,
// forgot the other" fail loudly with the file to fix named.
func TestPackSelectionMirroredInReleaseLane(t *testing.T) {
	root := repoRoot(t)
	scriptPath := filepath.Join(root, "tools", "ci", "release", "package_content.sh")
	script, err := os.ReadFile(scriptPath)
	if err != nil {
		// Same sandbox posture as TestCommittedTreeValidates: the
		// enforcing lane is `go test` from the repo root.
		t.Skipf("release packaging script not present at %s (sandboxed run)", scriptPath)
	}
	text := string(script)
	for name := range packFileNames {
		if !strings.Contains(text, "'"+name+"'") {
			t.Errorf("pack file name %q is selected here but missing from %s -- extend the reference-pack find clause to keep the two homes mirrored", name, scriptPath)
		}
	}
	for _, prefix := range packPathPrefixes {
		if !strings.Contains(text, prefix) {
			t.Errorf("pack path prefix %q is selected here but missing from %s -- extend the reference-pack find clause to keep the two homes mirrored", prefix, scriptPath)
		}
	}
}

// validFixtureTree is a minimal tree satisfying the whole structure
// contract; broken-tree cases mutate exactly one thing at a time.
func validFixtureTree() map[string]string {
	return map[string]string{
		"skills/demo/SKILL.md":            "---\nname: demo\ndescription: A demo skill.\n---\n\n# Demo\n\nRead `references/topic.md` first.\n",
		"skills/demo/references/topic.md": "The topic, explained.\n",
		"skills/compat.yaml":              "minimum_daemon_version: v0.0.0\nminimum_cli_version: v0.0.0\n",
		"agents/helper/instructions.md":   "You are the helper.\n",
		"automations/investigate.yaml":    "slug: investigate\ndisplayName: Investigate\n",
		// README.md must be ignored by the loader -- only *.yaml files are
		// definitions.
		"automations/README.md": "The automations tree.\n",
	}
}

// catalogFixtureTree is validFixtureTree plus a minimal catalog skill and
// a miniature catalog/ tree exercising every selection rule: the _docs
// root files, a patterns page, a provider index, a kind reference
// page with authored wisdom, the kind's fact-sheet sidecars, the
// central estimates and compliance documents, the pricing-machinery
// siblings that must be ignored, a non-pack file that must be ignored,
// and a _test fixture that must be excluded.
func catalogFixtureTree() map[string]string {
	files := validFixtureTree()
	files["skills/multi-cloud-catalog/SKILL.md"] = "---\nname: multi-cloud-catalog\ndescription: Research the catalog pack.\n---\n\n# Catalog\n\nRead `references/pack-layout.md` first.\n"
	files["skills/multi-cloud-catalog/references/pack-layout.md"] = "The pack lives in kinds/.\n"
	files["catalog/_docs/reference-commons.md"] = "The manifest grammar.\n"
	files["catalog/_docs/reference-index.md"] = "| provider | kinds |\n"
	files["catalog/_docs/reference-graph.yaml"] = "edges: []\n"
	files["catalog/_docs/GUIDE.md"] = "Catalog-level wisdom.\n"
	files["catalog/_patterns/observability.md"] = "The observability pattern.\n"
	files["catalog/aws/reference-index.md"] = "| kind | purpose |\n"
	files["catalog/aws/awsvpc/v1alpha1/reference.md"] = "# AwsVpc\n"
	files["catalog/aws/awsvpc/GUIDE.md"] = "AwsVpc wisdom.\n"
	files["catalog/aws/awsvpc/v1alpha1/spec.proto"] = "syntax = \"proto3\";\n"
	files["catalog/aws/awsvpc/README.md"] = "Not part of the pack.\n"
	files["catalog/aws/awsvpc/cost.yaml"] = "billingModel: usage_based\n"
	files["catalog/aws/awsvpc/controls.yaml"] = "controls: []\n"
	files["catalog/aws/awsvpc/iac/permissions.yaml"] = "providers: {}\n"
	files["catalog/_pricing/estimates/awsvpc.yaml"] = "presets: []\n"
	files["catalog/_pricing/models/awsvpc.yaml"] = "machinery, never ships\n"
	files["catalog/_pricing/pricebook/aws.yaml"] = "machinery, never ships\n"
	files["catalog/_pricing/derivations/awsvpc.yaml"] = "machinery, never ships\n"
	files["catalog/_compliance/controls-catalog.yaml"] = "controls: []\n"
	files["catalog/_compliance/frameworks/cis-aws.yaml"] = "mappings: []\n"
	files["catalog/_compliance/README.md"] = "Not a pack document.\n"
	files["catalog/_test/fake/v1alpha1/reference.md"] = "# Fake\n"
	return files
}

func writeTree(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for path, content := range files {
		full := filepath.Join(root, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// Validate requires the agents/ dir to exist even in skill-only cases.
	if err := os.MkdirAll(filepath.Join(root, "agents"), 0o755); err != nil {
		t.Fatal(err)
	}
	return root
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate test file")
	}
	return filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(thisFile))))
}

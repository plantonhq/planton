package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
)

// BrowseManifestFileName is the browse manifest's filename inside a
// definitions release. It is deliberately a SEPARATE document from
// definitions-manifest.json: the integrity manifest is the install-and-
// verify contract (whole-artifact checksums consumers refuse on), while
// the browse manifest describes the release's per-file tree for reading
// surfaces. Two single-purpose documents keep the integrity manifest's
// shape frozen for its existing consumers.
const BrowseManifestFileName = "definitions-browse.json"

// ExplodedDirName is the directory under releases/{tag}/definitions/ that
// carries every skill's content as individually fetchable files
// (exploded/<slug>/<archive-path>). The release workflow derives it by
// unzipping the skill archives -- never from a second packaging pass -- and
// verifies every file against this manifest before uploading, so the
// exploded tree and the archives structurally cannot diverge.
const ExplodedDirName = "exploded"

// BrowseManifest describes one definitions release as a browsable file
// tree: every skill file individually addressable under the exploded
// layout, plus the agent and automation files that already ship
// individually at the release root. One fetch of this document is enough
// to render the whole release's tree; each file is then fetched only when
// opened.
type BrowseManifest struct {
	Version string        `json:"version"`
	Skills  []SkillBrowse `json:"skills"`
	Agents  []BrowseFile  `json:"agents"`
	// omitempty mirrors the integrity manifest's posture: a release
	// carrying no automations omits the field rather than writing an
	// empty array.
	Automations []BrowseFile `json:"automations,omitempty"`
}

// SkillBrowse is one skill's file tree. Paths are archive-relative
// (SKILL.md at the root, references/ beside it, the catalog pack under
// kinds/); the downloadable URL for each is
// releases/{version}/definitions/exploded/{slug}/{path}.
type SkillBrowse struct {
	Slug  string       `json:"slug"`
	Files []BrowseFile `json:"files"`
}

// BrowseFile is one fetchable file: its path (archive-relative for skill
// files; the release-root filename for agents and automations), size, and
// content checksum -- enough for a reader to verify what it fetched and
// for the release workflow to verify the exploded tree before upload.
type BrowseFile struct {
	Path      string `json:"path"`
	SizeBytes int64  `json:"sizeBytes"`
	Sha256    string `json:"sha256"`
}

// BuildBrowseManifest composes the browse manifest from the same
// skillEntries seam the archive builder packs, so the two carriers of a
// skill's content agree by construction. Output ordering is fully
// deterministic (skills in tree order -- ReadDir is lexical -- and files
// sorted by path) because the manifest's bytes, like the archives', must
// be a pure function of the tree's content.
func BuildBrowseManifest(tree *Tree, version string) *BrowseManifest {
	manifest := &BrowseManifest{Version: version}

	for _, skill := range tree.Skills {
		entries := skillEntries(skill)
		var paths []string
		for path := range entries {
			paths = append(paths, path)
		}
		sort.Strings(paths)
		browse := SkillBrowse{Slug: skill.Slug}
		for _, path := range paths {
			browse.Files = append(browse.Files, describeBrowseFile(path, entries[path]))
		}
		manifest.Skills = append(manifest.Skills, browse)
	}

	for _, agent := range tree.Agents {
		fileName := "agent-" + agent.Slug + "-instructions.md"
		manifest.Agents = append(manifest.Agents, describeBrowseFile(fileName, agent.Instructions))
	}

	for _, automation := range tree.Automations {
		fileName := "automation-" + automation.Slug + ".yaml"
		manifest.Automations = append(manifest.Automations, describeBrowseFile(fileName, automation.Content))
	}

	return manifest
}

// EncodeBrowseManifest renders the manifest in the same style as the
// integrity manifest (indented JSON, trailing newline).
func EncodeBrowseManifest(manifest *BrowseManifest) ([]byte, error) {
	encoded, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(encoded, '\n'), nil
}

func describeBrowseFile(path string, content []byte) BrowseFile {
	sum := sha256.Sum256(content)
	return BrowseFile{
		Path:      path,
		SizeBytes: int64(len(content)),
		Sha256:    hex.EncodeToString(sum[:]),
	}
}

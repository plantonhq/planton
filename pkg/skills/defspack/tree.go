package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Tree is the loaded definitions content: every skill under skills/, every
// agent under agents/, every automation under automations/, and the
// compatibility floor from skills/compat.yaml.
type Tree struct {
	Root        string
	Skills      []Skill
	Agents      []Agent
	Automations []Automation
	Floor       CompatFloor
}

// Skill is one skills/<slug>/ directory: SKILL.md plus its reference files.
type Skill struct {
	Slug            string
	Dir             string
	FrontmatterName string
	Description     string
	// FrontmatterKeys lists every top-level key the frontmatter carries,
	// so validation can hold the metadata surface to the Agent Skills
	// specification's field set.
	FrontmatterKeys []string
	SkillMD         []byte
	// ReferenceFiles maps a references/-relative name (e.g. "chart-format.md")
	// to its content, for every file present on disk.
	ReferenceFiles map[string][]byte
	// ReferenceDirs lists directories found under references/. The contract
	// is FLAT files: a directory's contents are invisible to packaging AND
	// to the bidirectional-citation check (which runs over ReferenceFiles),
	// so a directory here is content an author believes exists and no agent
	// will ever receive. Recorded so Validate refuses it loudly instead of
	// the loader dropping it silently.
	ReferenceDirs []string
	// CitedReferences lists the references/<name>.md tokens SKILL.md cites,
	// deduplicated and sorted.
	CitedReferences []string
	// PackFiles carries content assembled from OUTSIDE the skill directory
	// at package time, keyed by archive path. Today only the catalog skill
	// uses it (the kind reference pack under kinds/ -- see
	// catalogpack.go); the git tree never duplicates these files.
	PackFiles map[string][]byte
}

// Agent is one agents/<slug>/ directory carrying the agent's instructions.
type Agent struct {
	Slug         string
	Instructions []byte
}

// Automation is one automations/<slug>.yaml file: a published definition of
// something the Planton Assistant can do unattended. The file's content is
// carried verbatim into the release; the deep schema laws (budget, pinned
// model, consent class) are enforced by the platform's publish lane, which
// owns the schema -- this repo's validation is structural (parseable YAML,
// slug matching the filename), so a contributor's typo fails the PR gate
// while the schema's law stays defined in exactly one place.
type Automation struct {
	Slug    string
	Content []byte
}

// CompatFloor is the minimum consumer version set the current content
// assumes (see skills/compat.yaml for the raising discipline).
type CompatFloor struct {
	MinimumDaemonVersion string `yaml:"minimum_daemon_version" json:"minimumDaemonVersion"`
	MinimumCliVersion    string `yaml:"minimum_cli_version" json:"minimumCliVersion"`
}

// citationPattern matches the reference citations SKILL.md uses in prose
// (e.g. `references/chart-format.md`). The path shape is the contract: a
// reference an agent can load must be citable this way, so the same token
// is what validation keys on.
var citationPattern = regexp.MustCompile(`references/([A-Za-z0-9._-]+\.md)`)

var semverPattern = regexp.MustCompile(`^v\d+\.\d+\.\d+$`)

// The Agent Skills specification (agentskills.io) is this tree's binding
// standard: skills authored here must load correctly in any conforming
// runtime, and the spec's sizing guidance exists because the always-loaded
// surfaces (name + description at startup, the SKILL.md body on activation)
// are paid for on every conversation. The limits below are the spec's own
// numbers; the ones the spec states as guidance (the 500-line body) are
// adopted here as law so drift fails a pull request instead of accumulating.
const (
	// maxSkillNameChars is the spec's ceiling for the frontmatter name.
	maxSkillNameChars = 64
	// maxDescriptionChars is the spec's ceiling for the description -- the
	// one surface every agent loads at startup for every skill.
	maxDescriptionChars = 1024
	// maxSkillMDLines adopts the spec's "keep your main SKILL.md under 500
	// lines" guidance as an enforced ceiling: detail belongs in references,
	// which load on demand.
	maxSkillMDLines = 500
)

// skillNamePattern is the spec's name grammar: lowercase alphanumerics and
// single hyphens, never leading, trailing, or consecutive.
var skillNamePattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// referenceNamePattern is the house grammar for reference filenames:
// lowercase alphanumeric segments (hyphens within a segment) separated by
// dots, ending in .md. Dots organize a large skill's references by domain
// (e.g. service.offline-deploy.md beside infra.chart-format.md) while the
// directory stays flat per the spec's own keep-references-one-level-deep
// guidance; a small skill's single-segment names are equally valid.
var referenceNamePattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*(\.[a-z0-9]+(-[a-z0-9]+)*)*\.md$`)

// specFrontmatterFields is the Agent Skills specification's complete
// frontmatter field set. Anything else belongs under the spec's own
// `metadata` map, so conforming runtimes never meet unknown top-level keys.
var specFrontmatterFields = map[string]bool{
	"name":          true,
	"description":   true,
	"license":       true,
	"compatibility": true,
	"metadata":      true,
	"allowed-tools": true,
}

// LoadTree reads skills/, agents/, and skills/compat.yaml under root.
// Loading is shape-tolerant (missing pieces surface in Validate, not here)
// but real I/O errors fail immediately.
func LoadTree(root string) (*Tree, error) {
	tree := &Tree{Root: root}

	skillsDir := filepath.Join(root, "skills")
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return nil, fmt.Errorf("reading skills tree %s: %w", skillsDir, err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		skill, err := loadSkill(filepath.Join(skillsDir, entry.Name()), entry.Name())
		if err != nil {
			return nil, err
		}
		if skill.Slug == catalogPackSkillSlug {
			pack, err := loadCatalogPack(root)
			if err != nil {
				return nil, err
			}
			skill.PackFiles = pack
		}
		tree.Skills = append(tree.Skills, *skill)
	}

	agentsDir := filepath.Join(root, "agents")
	agentEntries, err := os.ReadDir(agentsDir)
	if err != nil {
		return nil, fmt.Errorf("reading agents tree %s: %w", agentsDir, err)
	}
	for _, entry := range agentEntries {
		if !entry.IsDir() {
			continue
		}
		instructions, err := os.ReadFile(filepath.Join(agentsDir, entry.Name(), "instructions.md"))
		if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("reading agent instructions for %s: %w", entry.Name(), err)
		}
		tree.Agents = append(tree.Agents, Agent{Slug: entry.Name(), Instructions: instructions})
	}

	// Absent-tolerant by design, unlike skills/ and agents/: automations are
	// the youngest definitions class, and consumers (the platform's publish
	// lane) treat a release without them as publishing zero automations.
	automationsDir := filepath.Join(root, "automations")
	automationEntries, err := os.ReadDir(automationsDir)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("reading automations tree %s: %w", automationsDir, err)
	}
	for _, entry := range automationEntries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		content, err := os.ReadFile(filepath.Join(automationsDir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("reading automation %s: %w", entry.Name(), err)
		}
		tree.Automations = append(tree.Automations, Automation{
			Slug:    strings.TrimSuffix(entry.Name(), ".yaml"),
			Content: content,
		})
	}

	compatRaw, err := os.ReadFile(filepath.Join(skillsDir, "compat.yaml"))
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("reading compat floor: %w", err)
		}
	} else if err := yaml.Unmarshal(compatRaw, &tree.Floor); err != nil {
		return nil, fmt.Errorf("parsing skills/compat.yaml: %w", err)
	}

	return tree, nil
}

func loadSkill(dir, slug string) (*Skill, error) {
	skill := &Skill{Slug: slug, Dir: dir, ReferenceFiles: map[string][]byte{}}

	skillMD, err := os.ReadFile(filepath.Join(dir, "SKILL.md"))
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("reading SKILL.md for %s: %w", slug, err)
	}
	skill.SkillMD = skillMD
	skill.FrontmatterName, skill.Description, skill.FrontmatterKeys = parseFrontmatter(skillMD)

	cited := map[string]bool{}
	for _, match := range citationPattern.FindAllStringSubmatch(string(skillMD), -1) {
		cited[match[1]] = true
	}
	for name := range cited {
		skill.CitedReferences = append(skill.CitedReferences, name)
	}
	sort.Strings(skill.CitedReferences)

	refsDir := filepath.Join(dir, "references")
	refEntries, err := os.ReadDir(refsDir)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("reading references for %s: %w", slug, err)
	}
	for _, ref := range refEntries {
		if ref.IsDir() {
			// Recorded, never loaded: Validate refuses directories here
			// (see Skill.ReferenceDirs for why silence would be worse).
			skill.ReferenceDirs = append(skill.ReferenceDirs, ref.Name())
			continue
		}
		content, err := os.ReadFile(filepath.Join(refsDir, ref.Name()))
		if err != nil {
			return nil, fmt.Errorf("reading reference %s/%s: %w", slug, ref.Name(), err)
		}
		skill.ReferenceFiles[ref.Name()] = content
	}
	sort.Strings(skill.ReferenceDirs)

	return skill, nil
}

// validatePack holds the catalog skill's assembled pack to its contract:
// present and recognizably a pack (the commons root marker plus at least
// one kind reference page -- an empty or misrooted assembly would
// ship a research skill with nothing to research), every file non-empty,
// and the whole archive comfortably inside the serving engine's push
// ceilings so catalog growth fails HERE with a clear message instead of
// failing the engine push at release time.
func validatePack(skill Skill, report func(format string, args ...any)) {
	if skill.Slug != catalogPackSkillSlug {
		return
	}
	if len(skill.PackFiles) == 0 {
		report("skills/%s: catalog pack is empty -- no catalog/ tree was found to assemble", skill.Slug)
		return
	}
	if _, ok := skill.PackFiles[packDirName+"/_docs/reference-commons.md"]; !ok {
		report("skills/%s: assembled pack is missing its root marker (%s/_docs/reference-commons.md)", skill.Slug, packDirName)
	}
	referencePages := 0
	totalBytes := len(skill.SkillMD)
	for _, content := range skill.ReferenceFiles {
		totalBytes += len(content)
	}
	for path, content := range skill.PackFiles {
		if len(content) == 0 {
			report("skills/%s: pack file %s is empty", skill.Slug, path)
		}
		if strings.HasSuffix(path, "/reference.md") {
			referencePages++
		}
		totalBytes += len(content)
	}
	if referencePages == 0 {
		report("skills/%s: assembled pack carries no kind reference pages", skill.Slug)
	}
	totalFiles := 1 + len(skill.ReferenceFiles) + len(skill.PackFiles)
	if totalFiles > maxSkillArchiveFiles {
		report("skills/%s: archive would carry %d files (limit %d, engine ceiling 10000) -- the pack has outgrown its margins", skill.Slug, totalFiles, maxSkillArchiveFiles)
	}
	if totalBytes > maxSkillArchiveBytes {
		report("skills/%s: archive would carry %d bytes (limit %d, engine ceiling 100MB compressed) -- the pack has outgrown its margins", skill.Slug, totalBytes, maxSkillArchiveBytes)
	}
}

// parseFrontmatter extracts the name, description, and the full top-level
// key set from SKILL.md's leading YAML frontmatter block. Absent or
// malformed frontmatter yields empty values, which Validate reports with a
// precise message.
func parseFrontmatter(skillMD []byte) (name, description string, keys []string) {
	text := string(skillMD)
	if !strings.HasPrefix(text, "---\n") {
		return "", "", nil
	}
	rest := text[len("---\n"):]
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return "", "", nil
	}
	var fm struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
	}
	if err := yaml.Unmarshal([]byte(rest[:end]), &fm); err != nil {
		return "", "", nil
	}
	var raw map[string]any
	if err := yaml.Unmarshal([]byte(rest[:end]), &raw); err == nil {
		for key := range raw {
			keys = append(keys, key)
		}
		sort.Strings(keys)
	}
	return fm.Name, fm.Description, keys
}

// Validate enforces the structure contract documented in skills/README.md.
// It returns every violation rather than stopping at the first, so a PR
// author sees the complete repair list in one run.
func Validate(tree *Tree) []error {
	var errs []error
	report := func(format string, args ...any) {
		errs = append(errs, fmt.Errorf(format, args...))
	}

	if len(tree.Skills) == 0 {
		report("skills/: no skill directories found")
	}
	for _, skill := range tree.Skills {
		if len(skill.SkillMD) == 0 {
			report("skills/%s: SKILL.md is missing or empty", skill.Slug)
			continue
		}
		if skill.FrontmatterName == "" {
			report("skills/%s: SKILL.md frontmatter is missing a name", skill.Slug)
		} else if skill.FrontmatterName != skill.Slug {
			// The serving engine derives the skill's identity from the
			// frontmatter name; a mismatch would publish content under a
			// different slug than its directory advertises.
			report("skills/%s: frontmatter name %q must equal the directory slug", skill.Slug, skill.FrontmatterName)
		}
		if !skillNamePattern.MatchString(skill.Slug) || len(skill.Slug) > maxSkillNameChars {
			report("skills/%s: the name breaks the Agent Skills spec's grammar -- lowercase alphanumerics and single hyphens, at most %d characters", skill.Slug, maxSkillNameChars)
		}
		if skill.Description == "" {
			report("skills/%s: SKILL.md frontmatter is missing a description", skill.Slug)
		} else if runeCount := len([]rune(skill.Description)); runeCount > maxDescriptionChars {
			report("skills/%s: the description is %d characters -- the Agent Skills spec caps it at %d, and it is loaded for every agent at startup; move capability detail into the body or references", skill.Slug, runeCount, maxDescriptionChars)
		}
		for _, key := range skill.FrontmatterKeys {
			if !specFrontmatterFields[key] {
				report("skills/%s: frontmatter key %q is outside the Agent Skills spec's field set -- custom properties belong under the spec's `metadata` map", skill.Slug, key)
			}
		}
		if lines := bytes.Count(skill.SkillMD, []byte("\n")); lines > maxSkillMDLines {
			report("skills/%s: SKILL.md is %d lines -- the spec's ceiling is %d; the body loads whole on every activation, so move detail into references (they load on demand)", skill.Slug, lines, maxSkillMDLines)
		}
		for _, dir := range skill.ReferenceDirs {
			report("skills/%s: references/%s/ is a directory -- references are FLAT files (dot-separated names organize domains, e.g. service.offline-deploy.md); a directory's contents are silently invisible to packaging and to this very gate, so nothing inside it would ever reach an agent", skill.Slug, dir)
		}
		for name := range skill.ReferenceFiles {
			if !referenceNamePattern.MatchString(name) {
				report("skills/%s: references/%s breaks the reference-name grammar -- lowercase alphanumeric segments (hyphens within a segment) separated by dots, ending in .md", skill.Slug, name)
			}
		}

		for _, cited := range skill.CitedReferences {
			content, exists := skill.ReferenceFiles[cited]
			if !exists {
				report("skills/%s: SKILL.md cites references/%s which does not exist", skill.Slug, cited)
			} else if len(content) == 0 {
				report("skills/%s: references/%s is empty -- an agent would load a blank document", skill.Slug, cited)
			}
		}
		citedSet := map[string]bool{}
		for _, cited := range skill.CitedReferences {
			citedSet[cited] = true
		}
		for name := range skill.ReferenceFiles {
			if !citedSet[name] {
				report("skills/%s: references/%s exists but SKILL.md never cites it -- orphaned content rots", skill.Slug, name)
			}
			if len(skill.ReferenceFiles[name]) == 0 && citedSet[name] {
				continue // already reported above
			}
			if len(skill.ReferenceFiles[name]) == 0 {
				report("skills/%s: references/%s is empty", skill.Slug, name)
			}
		}

		validatePack(skill, report)
	}

	if len(tree.Agents) == 0 {
		report("agents/: no agent directories found")
	}
	for _, agent := range tree.Agents {
		if len(agent.Instructions) == 0 {
			report("agents/%s: instructions.md is missing or empty", agent.Slug)
		}
	}

	for _, automation := range tree.Automations {
		if len(automation.Content) == 0 {
			report("automations/%s.yaml: file is empty", automation.Slug)
			continue
		}
		var doc struct {
			Slug string `yaml:"slug"`
		}
		if err := yaml.Unmarshal(automation.Content, &doc); err != nil {
			report("automations/%s.yaml: not valid YAML: %v", automation.Slug, err)
			continue
		}
		if doc.Slug != automation.Slug {
			// Slugs are permanent join keys (org switch records adopt by
			// slug), so the filename and the document must agree before the
			// content can enter a release.
			report("automations/%s.yaml: document slug %q must equal the file name", automation.Slug, doc.Slug)
		}
	}

	if !semverPattern.MatchString(tree.Floor.MinimumDaemonVersion) {
		report("skills/compat.yaml: minimum_daemon_version %q is not a vX.Y.Z version", tree.Floor.MinimumDaemonVersion)
	}
	if !semverPattern.MatchString(tree.Floor.MinimumCliVersion) {
		report("skills/compat.yaml: minimum_cli_version %q is not a vX.Y.Z version", tree.Floor.MinimumCliVersion)
	}

	return errs
}

package skills

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The planton skill's body and its references are reachable, rooted at the
// skill's own directory -- the shape a host walks to check the skill's
// commands.
func TestPlantonSkillIsEmbedded(t *testing.T) {
	skill, ok := Skill("planton")
	if !ok {
		t.Fatal("Skill(\"planton\") reported no such skill")
	}
	body, err := fs.ReadFile(skill, "SKILL.md")
	if err != nil {
		t.Fatalf("reading SKILL.md: %v", err)
	}
	if !bytes.HasPrefix(body, []byte("---\nname: planton\n")) {
		t.Errorf("SKILL.md does not open with the planton frontmatter: %q", firstLine(body))
	}
	if _, err := fs.ReadFile(skill, "references/craft.planton-cli.md"); err != nil {
		t.Errorf("reading the CLI command map reference: %v", err)
	}
	if _, ok := Skill("no-such-skill"); ok {
		t.Error("Skill(\"no-such-skill\") reported a skill")
	}
}

// Every authored file on disk is embedded, byte for byte, and nothing else
// is: the package's promise is that its text IS the skill an agent loads.
func TestDocumentsAreTheFilesOnDisk(t *testing.T) {
	docs, err := Documents()
	if err != nil {
		t.Fatalf("Documents: %v", err)
	}
	embedded := map[string][]byte{}
	for _, doc := range docs {
		embedded[doc.Path] = doc.Content
	}

	onDisk := map[string]bool{}
	for _, slug := range []string{"planton", "multi-cloud-catalog"} {
		onDisk[slug+"/SKILL.md"] = true
		refs, err := os.ReadDir(filepath.Join(slug, "references"))
		if err != nil {
			t.Fatalf("listing %s references: %v", slug, err)
		}
		for _, ref := range refs {
			if !ref.IsDir() && strings.HasSuffix(ref.Name(), ".md") {
				onDisk[slug+"/references/"+ref.Name()] = true
			}
		}
	}

	for path := range onDisk {
		want, err := os.ReadFile(filepath.FromSlash(path))
		if err != nil {
			t.Fatalf("reading %s: %v", path, err)
		}
		got, ok := embedded[path]
		if !ok {
			t.Errorf("%s is on disk but not embedded", path)
			continue
		}
		if !bytes.Equal(got, want) {
			t.Errorf("%s differs from the file on disk", path)
		}
	}
	for path := range embedded {
		if !onDisk[path] {
			t.Errorf("%s is embedded but is not an authored skill file", path)
		}
	}
}

func firstLine(b []byte) string {
	line, _, _ := bytes.Cut(b, []byte("\n"))
	return string(line)
}

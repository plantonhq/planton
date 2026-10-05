// Package skills carries the authored text of the skills in this directory
// -- each skill's SKILL.md and its references -- as Go, for hosts that must
// check what a skill claims against their own surface.
//
// A skill teaches an agent to run commands and read records that live
// elsewhere: the planton skill names the Planton CLI's commands, and a host
// that ships that CLI can prove every command the skill names resolves in
// its real command tree. That proof needs the skill's words at test time,
// and under Bazel a test reads only what a package exposes. This package is
// that exposure and nothing more: the embedded files ARE the skill, byte for
// byte the files beside this one, so a check that passes here passes against
// what agents load.
//
// Only authored text is embedded. The multi-cloud-catalog skill's kind
// pack is assembled from catalog/ at release time (pkg/skills/defspack) and
// is not part of this package.
//
// This file does not ship inside any skill. The skills validator loads only
// directories under skills/, and releases package what it built, so a Go
// file here is invisible to every skill consumer.
package skills

import (
	"embed"
	"io/fs"
)

//go:embed planton/SKILL.md planton/references/*.md
//go:embed multi-cloud-catalog/SKILL.md multi-cloud-catalog/references/*.md
var files embed.FS

// Document is one authored skill file: its path under skills/ (for example
// "planton/references/craft.planton-cli.md") and its bytes.
type Document struct {
	Path    string
	Content []byte
}

// FS is the embedded tree, rooted at skills/: "<slug>/SKILL.md" and
// "<slug>/references/<name>.md" for every skill this package carries.
func FS() fs.FS {
	return files
}

// Skill is one skill's files, rooted at its own directory: "SKILL.md" and
// "references/<name>.md". ok is false for a slug this package does not
// carry.
func Skill(slug string) (skill fs.FS, ok bool) {
	if _, err := fs.Stat(files, slug+"/SKILL.md"); err != nil {
		return nil, false
	}
	sub, err := fs.Sub(files, slug)
	if err != nil {
		return nil, false
	}
	return sub, true
}

// Documents returns every embedded file in path order -- the shape a check
// that walks the text line by line wants, so a failure can name the file
// and line an author must fix.
func Documents() ([]Document, error) {
	var docs []Document
	err := fs.WalkDir(files, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		content, readErr := files.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		docs = append(docs, Document{Path: path, Content: content})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return docs, nil
}

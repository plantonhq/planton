// Package priceprovenance is the machine-enforced guarantee that the
// catalog's PROSE never quotes a cloud price. Verified dollar figures have
// exactly one home -- the pinned, source-dated price books and the generated
// estimates under catalog/_pricing/ -- and every other dollar figure in a
// kind's documentation is a second, unverified source of truth that
// contradicts the verified one the day the provider reprices. This gate makes
// writing one a CI failure instead of a review hope.
//
// What documentation teaches instead of a rate: the cost DRIVERS (which
// dimensions bill, which choices are the cost cliffs) and, when a magnitude
// comparison is essential, a structural statement ("roughly 4x the smaller
// tier", "billed hourly whether idle or not") -- never dollars.
//
// Two kinds of dollar figures are legitimate and live in the baseline's
// `allowed` list, each with its reason recorded beside it:
//
//   - a user-chosen dollar VALUE illustrating a dollar-typed configuration
//     field (a budget limit, a cost-anomaly alert threshold) -- that is the
//     user's number, not a provider rate, and banning it would break honest
//     teaching of dollar-typed kinds;
//   - a non-price token the scanner cannot distinguish (a regex
//     backreference like "/v2/$1" in a rewrite example).
//
// The scanned surfaces are every markdown file under catalog/ EXCEPT
// catalog/_pricing/ (the one legitimate dollar home), plus every markdown
// file under charts/. Generated reference.md pages are deliberately IN scope:
// their dollar figures originate in proto comments, and the fix belongs at
// the proto source followed by `make generate-reference`.
//
// The descriptor of accepted gaps lives in baseline.yaml -- a burn-down list
// mirroring pkg/secretcoverage's and pkg/anatomy's baselines (a reader who
// knows one knows all three). The CI lane is
// .github/workflows/lint.price-provenance.yaml.
package priceprovenance

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// pricePattern matches a dollar token: `$` followed by digits, optionally
// with thousands separators and decimals ($3, $0.65, $9,042.50). The scanner
// deliberately matches the TOKEN, not surrounding units: a bare "$100" in
// prose is either a price (a violation), a user-chosen config example (an
// allowed entry with its reason), or a scanner false positive like a regex
// backreference (also an allowed entry) -- the baseline's allowed list is
// where that judgment is recorded once, permanently.
var pricePattern = regexp.MustCompile(`\$\d[\d,]*(?:\.\d+)?`)

// Finding is one dollar token in one scanned file.
type Finding struct {
	// Path is repo-root-relative, forward-slash form.
	Path string
	// Token is the matched dollar token exactly as written (e.g. "$0.65").
	Token string
	// Line is 1-based, for human-facing reports only -- it is deliberately
	// NOT part of the finding's identity, so unrelated edits above a token
	// never churn the baseline.
	Line int
}

// ID is the finding's stable identity: "<path>:<token>". One allowed or
// baselined entry covers every occurrence of that token in that file --
// coarse on purpose: identities that survive unrelated edits keep the
// baseline honest, and a file legitimately using "$100" twice should not
// need two entries.
func (f Finding) ID() string {
	return f.Path + ":" + f.Token
}

// scanRoots are the surfaces this gate holds, relative to the repo root.
// catalog/_pricing/ is excluded by scan(), not omitted here, so the
// exclusion is visible in one place beside its reason.
var scanRoots = []string{"catalog", "charts"}

// Scan walks the scanned surfaces under repoRoot and returns every dollar
// token in every markdown file, sorted by path then token.
func Scan(repoRoot string) ([]Finding, error) {
	var findings []Finding
	for _, root := range scanRoots {
		rootPath := filepath.Join(repoRoot, root)
		err := filepath.WalkDir(rootPath, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				// catalog/_pricing/ is the verified-data home -- the ONE
				// place dollar figures belong, with source URLs and
				// retrieval dates. Everything else is prose.
				if d.Name() == "_pricing" {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(d.Name(), ".md") {
				return nil
			}
			rel, relErr := filepath.Rel(repoRoot, path)
			if relErr != nil {
				return relErr
			}
			fileFindings, scanErr := scanFile(path, filepath.ToSlash(rel))
			if scanErr != nil {
				return scanErr
			}
			findings = append(findings, fileFindings...)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Path != findings[j].Path {
			return findings[i].Path < findings[j].Path
		}
		return findings[i].Token < findings[j].Token
	})
	return findings, nil
}

func scanFile(path, rel string) ([]Finding, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var findings []Finding
	seen := map[string]bool{}
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		for _, token := range pricePattern.FindAllString(scanner.Text(), -1) {
			finding := Finding{Path: rel, Token: token, Line: lineNo}
			if seen[finding.ID()] {
				continue
			}
			seen[finding.ID()] = true
			findings = append(findings, finding)
		}
	}
	return findings, scanner.Err()
}

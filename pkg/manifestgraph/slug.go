package manifestgraph

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/plantonhq/planton/shared"
	"golang.org/x/text/unicode/norm"
)

// SlugPattern is the one rule every resource slug obeys: lowercase letters
// and digits joined by single hyphens (my-app-2). It is spelled exactly as
// CloudResourceMetadata.slug's protovalidate rule and the platform's
// ApiResourceSlugLaw.PATTERN spell it. The alphabet is the one every system a
// slug is written into accepts: DNS labels, secret store names, cloud labels
// and tags, and dot-delimited identities (the Pulumi stack
// <env>.<Kind>.<slug>), which is why a slug carries no dot and no underscore.
const SlugPattern = `^[a-z0-9]+(-[a-z0-9]+)*$`

// SlugRule is the one sentence a refusal of an unlawful slug says, everywhere.
const SlugRule = "A slug is lowercase letters and digits joined by single hyphens, like my-app-2."

var (
	lawfulSlug    = regexp.MustCompile(SlugPattern)
	outsideTheLaw = regexp.MustCompile(`[^a-z0-9]+`)
)

// IsLawfulSlug reports whether slug obeys SlugPattern. The empty string is
// not a lawful slug (an empty metadata.slug is legal only because the slug is
// then derived from the name).
func IsLawfulSlug(slug string) bool {
	return lawfulSlug.MatchString(slug)
}

// GenerateSlug derives a resource's identity slug from its human-readable
// name: accents dropped ("Résumé" becomes "resume"), lowercased, and every
// run of other characters -- spaces, dots, underscores, apostrophes,
// symbols -- folded into one hyphen, trimmed at both ends. "Sandeep's AWS" is
// "sandeep-s-aws", "example.com" is "example-com", "DB_PASSWORD" is
// "db-password". A blank name, or one that keeps no letter or digit, yields
// "". The result is always lawful (IsLawfulSlug) or empty.
//
// This mirrors the platform's ApiRequestResourceSlugGenerator (the source of
// truth) byte for byte, and slug_test.go copies its test table: node and
// edge-target identity both derive through it, so a manifest named
// "My Shared VPC" and a reference naming "My Shared VPC" land on the slug the
// server will store -- the join that makes them one graph node.
func GenerateSlug(name string) string {
	if strings.TrimSpace(name) == "" {
		return ""
	}
	// Lowercase before dropping accents: lowercasing can itself produce a
	// combining mark ("İ" becomes "i" plus a combining dot), and every mirror
	// must agree on "istanbul".
	lowered := strings.ToLower(norm.NFD.String(name))
	withoutAccents := strings.Map(func(r rune) rune {
		if unicode.Is(unicode.M, r) {
			return -1
		}
		return r
	}, lowered)
	folded := outsideTheLaw.ReplaceAllString(withoutAccents, "-")
	return strings.Trim(folded, "-")
}

// ResolveSlug resolves a manifest's identity slug: an explicit metadata.slug
// passes through untouched (it IS the platform identity when present),
// otherwise the slug generates from metadata.name.
func ResolveSlug(meta *shared.CloudResourceMetadata) string {
	if meta == nil {
		return ""
	}
	if meta.GetSlug() != "" {
		return meta.GetSlug()
	}
	return GenerateSlug(meta.GetName())
}

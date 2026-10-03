package module

const (
	// OpSiteTag is the exported output containing the
	// Cloudflare-assigned site tag (the site's identity in every RUM API
	// path).
	OpSiteTag = "site_tag"

	// OpSiteToken is the exported output containing the site's
	// measurement token (secret-marked -- the beacon credential).
	OpSiteToken = "site_token"

	// OpSnippet is the exported output containing the ready-to-embed
	// JavaScript snippet (secret-marked: it carries the site token).
	OpSnippet = "snippet"

	// OpRulesetId is the exported output containing the site's
	// ruleset ID (the parent object the include/exclude rules live under).
	OpRulesetId = "ruleset_id"
)

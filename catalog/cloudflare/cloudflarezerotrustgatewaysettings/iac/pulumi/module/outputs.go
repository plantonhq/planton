package module

const (
	// OpAccountId is the exported output containing the account the
	// Gateway configuration was applied to (the singleton's identity -- the
	// harness and import recipes key on it).
	OpAccountId = "account_id"
	// OpPacfileIds is the exported map of PAC-file Cloudflare ids, keyed by
	// each file's name (the same key the tofu module uses for for_each) --
	// import recipes derive per-file import IDs from it because PAC-file ids
	// are server-assigned at creation.
	OpPacfileIds = "pacfile_ids"
)

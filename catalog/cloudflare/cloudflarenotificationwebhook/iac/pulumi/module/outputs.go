package module

const (
	// OpWebhookId is the exported output containing the
	// Cloudflare-assigned UUID of the webhook destination (what
	// notification policies reference).
	OpWebhookId = "webhook_id"

	// OpType is the exported output containing the destination type
	// Cloudflare inferred from the URL (datadog, discord, feishu, gchat,
	// generic, opsgenie, slack, splunk).
	OpType = "type"
)

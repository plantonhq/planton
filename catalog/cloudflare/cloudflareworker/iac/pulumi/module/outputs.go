package module

const (
	// OpScriptId is the exported output containing the Worker script ID.
	OpScriptId = "script_id"
	// OpScriptName is the exported output containing the Worker script name.
	OpScriptName = "script_name"
	// OpCustomDomainHostnames lists the custom-domain hostnames attached to the Worker.
	OpCustomDomainHostnames = "custom_domain_hostnames"
	// OpRoutePatterns lists the route patterns mapped to the Worker.
	OpRoutePatterns = "route_patterns"
	// OpCustomDomainIds is hostname → Cloudflare domain id (import half).
	OpCustomDomainIds = "custom_domain_ids"
	// OpRouteIds is list-index → Cloudflare route id (import half).
	OpRouteIds = "route_ids"
	// OpRouteZoneIds is list-index → zone id (the other import half).
	OpRouteZoneIds = "route_zone_ids"
)

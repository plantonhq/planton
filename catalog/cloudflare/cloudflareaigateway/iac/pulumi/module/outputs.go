package module

const (
	// OpGatewayId is the exported output containing the gateway's id
	// (URL slug) -- the segment clients put in the gateway endpoint URL.
	OpGatewayId = "gateway_id"
	// OpDynamicRouteIds is the exported output mapping each managed
	// dynamic route's name to its id -- what the per-route import identity
	// derives from.
	OpDynamicRouteIds = "dynamic_route_ids"
)

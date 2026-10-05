package module

import (
	"crypto/sha1"
	"encoding/hex"
	"sort"
	"strconv"
	"strings"

	awsrestapigatewayv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsrestapigateway/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	Target *awsrestapigatewayv1alpha1.AwsRestApiGateway
	Spec   *awsrestapigatewayv1alpha1.AwsRestApiGatewaySpec

	// StageName is the resolved stage name ("prod" when the stage is
	// omitted or unnamed).
	StageName string

	// ResourcePaths is every tree node the route paths imply
	// ("/users/{id}" implies "/users" and "/users/{id}"), sorted
	// shallow-first so parents always exist before children.
	ResourcePaths []string

	// DeploymentTriggerHash fingerprints the API DEFINITION (everything
	// except the stage and documentation): any change redeploys - the
	// declarative behavior REST APIs' explicit-snapshot model would
	// otherwise lose.
	DeploymentTriggerHash string

	AwsTags map[string]string
}

func initializeLocals(_ *pulumi.Context, in *awsrestapigatewayv1alpha1.AwsRestApiGatewayIacInput) (*Locals, error) {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	locals.StageName = "prod"
	if locals.Spec.Stage != nil && locals.Spec.Stage.Name != "" {
		locals.StageName = locals.Spec.Stage.Name
	}

	locals.ResourcePaths = derivedResourcePaths(locals.Spec.Routes)

	hash, err := definitionHash(locals.Spec)
	if err != nil {
		return nil, err
	}
	locals.DeploymentTriggerHash = hash

	metadata := in.Target.Metadata

	// Resource-identity tags match the Terraform module key-for-key.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsRestApiGateway.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals, nil
}

// derivedResourcePaths expands route paths into the full set of tree
// nodes, sorted shallow-first then lexically (deterministic previews;
// parents before children).
func derivedResourcePaths(routes []*awsrestapigatewayv1alpha1.AwsRestApiGatewayRoute) []string {
	seen := map[string]bool{}
	for _, r := range routes {
		if r.Path == "/" {
			continue
		}
		segments := strings.Split(strings.TrimPrefix(r.Path, "/"), "/")
		for i := 1; i <= len(segments); i++ {
			seen["/"+strings.Join(segments[:i], "/")] = true
		}
	}
	paths := make([]string, 0, len(seen))
	for p := range seen {
		paths = append(paths, p)
	}
	sort.Slice(paths, func(i, j int) bool {
		di, dj := strings.Count(paths[i], "/"), strings.Count(paths[j], "/")
		if di != dj {
			return di < dj
		}
		return paths[i] < paths[j]
	})
	return paths
}

// definitionHash fingerprints the API definition. The stage,
// documentation, description, and region are excluded: changing them
// must not roll a new deployment (the same exclusion set the Terraform
// module's explicit hash list encodes). Each engine hashes its own
// canonical rendering - the values differ across engines by design;
// redeploy-on-change behavior is what parity requires.
func definitionHash(spec *awsrestapigatewayv1alpha1.AwsRestApiGatewaySpec) (string, error) {
	clone := proto.Clone(spec).(*awsrestapigatewayv1alpha1.AwsRestApiGatewaySpec)
	clone.Stage = nil
	clone.Documentation = nil
	clone.Description = ""
	clone.Region = ""
	raw, err := protojson.MarshalOptions{UseProtoNames: true}.Marshal(clone)
	if err != nil {
		return "", err
	}
	sum := sha1.Sum(raw)
	return hex.EncodeToString(sum[:]), nil
}

// parentPath returns the tree parent of a resource path ("" for
// first-level segments, which attach under the root resource).
func parentPath(path string) string {
	idx := strings.LastIndex(path, "/")
	if idx <= 0 {
		return ""
	}
	return path[:idx]
}

// lastSegment returns the path's final segment (the API Gateway
// path_part).
func lastSegment(path string) string {
	idx := strings.LastIndex(path, "/")
	return path[idx+1:]
}

// routeKey is the stable per-route key both engines share ("GET
// /users/{id}") - resource names and output map keys derive from it.
func routeKey(r *awsrestapigatewayv1alpha1.AwsRestApiGatewayRoute) string {
	return r.Method + " " + r.Path
}

// sortedRoutes returns routes ordered by their route key for
// deterministic previews.
func sortedRoutes(in []*awsrestapigatewayv1alpha1.AwsRestApiGatewayRoute) []*awsrestapigatewayv1alpha1.AwsRestApiGatewayRoute {
	out := append([]*awsrestapigatewayv1alpha1.AwsRestApiGatewayRoute{}, in...)
	sort.Slice(out, func(i, j int) bool { return routeKey(out[i]) < routeKey(out[j]) })
	return out
}

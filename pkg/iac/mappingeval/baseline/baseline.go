// Package baseline is the deterministic reference proposer: it maps a
// read-only account scan to Planton kind manifests using only plain
// code and the scan's own facts -- no model, no judgment. It exists for two
// reasons, both about keeping the eval harness honest:
//
//   - it proves the whole examination end to end (seed, scan, propose,
//     score) with a proposer whose behavior is exactly reproducible, and is
//     pinned to a PERFECT score on the seeded suites -- any score drop is a
//     harness or recipe regression, never model variance;
//   - it sets the floor an AI mapper must beat: on a clean, well-signaled
//     account, free deterministic code already maps the staples correctly,
//     so the AI earns its place only on the genuinely fuzzy remainder
//     (ambiguous grouping, messy accounts, judgment calls).
//
// The mappers here are deliberately BOUNDED to the kinds the seeded suites
// exercise. This package is not on a path to cover the whole catalog --
// generalizing per-kind mapping code is exactly the infeasible-by-hand work
// the AI proposer exists to replace.
package baseline

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	proposalv1 "github.com/plantonhq/planton/iac/importmappingproposal/v1"
	"github.com/plantonhq/planton/pkg/catalogkindreflect"
	"github.com/plantonhq/planton/pkg/iac/envpartition"
	"github.com/plantonhq/planton/pkg/iac/envpartition/awsscan"
	"github.com/plantonhq/planton/pkg/iac/mappingeval"
	"google.golang.org/protobuf/types/known/structpb"
)

// Cloud Control type names the baseline maps.
const (
	typeVPC             = "AWS::EC2::VPC"
	typeSubnet          = "AWS::EC2::Subnet"
	typeRouteTable      = "AWS::EC2::RouteTable"
	typeRTBAssociation  = "AWS::EC2::SubnetRouteTableAssociation"
	typeInternetGateway = "AWS::EC2::InternetGateway"
	typeBucket          = "AWS::S3::Bucket"
	typeBucketPolicy    = "AWS::S3::BucketPolicy"
	typeQueue           = "AWS::SQS::Queue"
	typeTopic           = "AWS::SNS::Topic"
)

// Propose maps the scan to an ImportMappingProposal. Deterministic: the
// same scan always yields the same proposal, byte for byte.
func Propose(scan *mappingeval.Scan) (*proposalv1.ImportMappingProposal, error) {
	b := &builder{
		scan:      scan,
		byType:    map[string][]mappingeval.ScannedResource{},
		claimed:   map[mappingeval.AccountResourceRef]bool{},
		nameByRef: map[mappingeval.AccountResourceRef]string{},
		envByRef:  partitionScan(scan),
	}
	for _, r := range scan.Resources {
		b.byType[r.TypeName] = append(b.byType[r.TypeName], r)
	}
	for _, resources := range b.byType {
		sort.Slice(resources, func(i, j int) bool { return resources[i].Identifier < resources[j].Identifier })
	}

	// Producers before consumers, so reference targets have names: VPCs,
	// then gateways, then subnets (which reference both), then the
	// standalone service resources.
	b.mapVPCs()
	b.mapInternetGateways()
	if err := b.mapSubnets(); err != nil {
		return nil, err
	}
	b.mapBuckets()
	b.mapQueues()
	b.mapTopics()
	b.recordUnmapped()

	return &proposalv1.ImportMappingProposal{
		ApiVersion: "iac.planton.dev/v1",
		Kind:       "ImportMappingProposal",
		Spec: &proposalv1.ImportMappingProposalSpec{
			Resources: b.resources,
			Unmapped:  b.unmapped,
		},
	}, nil
}

type builder struct {
	scan      *mappingeval.Scan
	byType    map[string][]mappingeval.ScannedResource
	resources []*proposalv1.ProposedResource
	unmapped  []*proposalv1.UnmappedAccountResource
	claimed   map[mappingeval.AccountResourceRef]bool
	// nameByRef records each claimed resource's proposed instance name, so
	// later mappers can reference producers (a subnet's vpc_id -> the VPC
	// instance that claims that vpc id).
	nameByRef map[mappingeval.AccountResourceRef]string
	// envByRef is the deterministic partition of the scan under the
	// UNTAUGHT default rule ("" = honestly unassigned). Resources are
	// partitioned; manifests aggregate (see instanceEnv).
	envByRef map[mappingeval.AccountResourceRef]string
}

// partitionScan runs the environment partition engine over the whole scan
// with the untaught default rule. The baseline stamps metadata.env only
// from these assignments -- it never guesses an environment, exactly as it
// never guesses a grouping.
func partitionScan(scan *mappingeval.Scan) map[mappingeval.AccountResourceRef]string {
	resources := make([]envpartition.Resource, 0, len(scan.Resources))
	for _, r := range scan.Resources {
		resources = append(resources, awsscan.Adapt(r.TypeName, r.Identifier, r.Properties))
	}
	result := envpartition.Partition(envpartition.DefaultRule(), resources)
	envByRef := make(map[mappingeval.AccountResourceRef]string, len(result.Assignments))
	for _, a := range result.Assignments {
		envByRef[mappingeval.AccountResourceRef{TypeName: a.TypeName, Identifier: a.Identifier}] = a.Environment
	}
	return envByRef
}

// instanceEnv aggregates the claimed resources' assignments into the
// instance's environment: the unique environment the assigned claims agree
// on. Unassigned claims carry no vote (no signal is not a veto); claims
// assigned to DIFFERENT environments mean the instance's environment is
// genuinely ambiguous, so the honest answer is none at all.
func (b *builder) instanceEnv(claims []mappingeval.AccountResourceRef) string {
	env := ""
	for _, claim := range claims {
		claimEnv := b.envByRef[claim]
		if claimEnv == "" {
			continue
		}
		if env != "" && env != claimEnv {
			return ""
		}
		env = claimEnv
	}
	return env
}

func (b *builder) mapVPCs() {
	for _, vpc := range b.byType[typeVPC] {
		spec := map[string]any{"region": b.scan.Region}
		setString(spec, "cidrBlock", vpc.Properties["CidrBlock"])
		setBool(spec, "enableDnsSupport", vpc.Properties["EnableDnsSupport"])
		setBool(spec, "enableDnsHostnames", vpc.Properties["EnableDnsHostnames"])
		// InstanceTenancy "default" is AWS's default; materializing it
		// would be noise, not information.
		if tenancy, _ := vpc.Properties["InstanceTenancy"].(string); tenancy != "" && tenancy != "default" {
			spec["instanceTenancy"] = tenancy
		}
		b.emit("AwsVpc", b.displayName(vpc), spec,
			"a VPC is its own component instance; grouping signal: the vpc id",
			vpc.Ref())
	}
}

func (b *builder) mapInternetGateways() {
	for _, gateway := range b.byType[typeInternetGateway] {
		spec := map[string]any{"region": b.scan.Region}
		if vpcID := firstAttachmentVpcID(gateway.Properties); vpcID != "" {
			spec["vpcId"] = b.refOrLiteral(typeVPC, vpcID, "AwsVpc", "status.outputs.vpc_id")
		}
		b.emit("AwsInternetGateway", b.displayName(gateway), spec,
			"an internet gateway is its own component instance; its VPC comes from the attachment",
			gateway.Ref())
	}
}

// mapSubnets groups each subnet with the route table explicitly associated
// to it (and the association resource itself) -- the ownership signal is
// the association's SubnetId. Main route tables are AWS-implicit and are
// deliberately NOT grouped into any subnet.
func (b *builder) mapSubnets() error {
	routeTables := map[string]mappingeval.ScannedResource{}
	for _, table := range b.byType[typeRouteTable] {
		routeTables[table.Identifier] = table
	}
	associationsBySubnet := map[string]mappingeval.ScannedResource{}
	for _, association := range b.byType[typeRTBAssociation] {
		if subnetID, _ := association.Properties["SubnetId"].(string); subnetID != "" {
			associationsBySubnet[subnetID] = association
		}
	}

	for _, subnet := range b.byType[typeSubnet] {
		spec := map[string]any{"region": b.scan.Region}
		if vpcID, _ := subnet.Properties["VpcId"].(string); vpcID != "" {
			spec["vpcId"] = b.refOrLiteral(typeVPC, vpcID, "AwsVpc", "status.outputs.vpc_id")
		}
		setString(spec, "availabilityZone", subnet.Properties["AvailabilityZone"])
		setString(spec, "cidrBlock", subnet.Properties["CidrBlock"])
		setBool(spec, "mapPublicIpOnLaunch", subnet.Properties["MapPublicIpOnLaunch"])

		claims := []mappingeval.AccountResourceRef{subnet.Ref()}
		rationale := "a subnet is its own component instance"

		if association, associated := associationsBySubnet[subnet.Identifier]; associated {
			routeTableID, _ := association.Properties["RouteTableId"].(string)
			table, known := routeTables[routeTableID]
			if !known {
				return errors.Errorf("association %s references route table %s, which the scan did not surface", association.Identifier, routeTableID)
			}
			routes, err := b.subnetRoutes(table)
			if err != nil {
				return errors.Wrapf(err, "routes of table %s (subnet %s)", routeTableID, subnet.Identifier)
			}
			if len(routes) > 0 {
				spec["routes"] = routes
			}
			claims = append(claims, table.Ref(), association.Ref())
			rationale += "; the explicitly associated route table and the association are the subnet's inline-routing surface (grouping signal: the association's SubnetId)"
		}

		b.emit("AwsSubnet", b.displayName(subnet), spec, rationale, claims...)
	}
	return nil
}

// subnetRoutes converts a route table's enriched Routes into the subnet
// spec's inline routes, skipping the VPC-local route AWS materializes in
// every table.
func (b *builder) subnetRoutes(table mappingeval.ScannedResource) ([]any, error) {
	rawRoutes, _ := table.Properties["Routes"].([]any)
	var routes []any
	for _, raw := range rawRoutes {
		route, _ := raw.(map[string]any)
		if route == nil {
			continue
		}
		if gatewayID, _ := route["GatewayId"].(string); gatewayID == "local" {
			continue
		}
		entry := map[string]any{}
		setString(entry, "destinationCidrBlock", route["DestinationCidrBlock"])
		setString(entry, "destinationIpv6CidrBlock", route["DestinationIpv6CidrBlock"])
		setString(entry, "destinationPrefixListId", route["DestinationPrefixListId"])

		targetType, targetID, err := routeTarget(route)
		if err != nil {
			return nil, err
		}
		entry["targetType"] = targetType
		if targetType == "internet_gateway" {
			entry["targetId"] = b.refOrLiteral(typeInternetGateway, targetID, "AwsInternetGateway", "status.outputs.internet_gateway_id")
		} else {
			entry["targetId"] = map[string]any{"value": targetID}
		}
		routes = append(routes, entry)
	}
	return routes, nil
}

// routeTarget reads which target attribute the route carries and maps it to
// the spec's target_type vocabulary.
func routeTarget(route map[string]any) (string, string, error) {
	targets := []struct{ property, targetType string }{
		{"GatewayId", "internet_gateway"},
		{"NatGatewayId", "nat_gateway"},
		{"TransitGatewayId", "transit_gateway"},
		{"VpcPeeringConnectionId", "vpc_peering_connection"},
		{"NetworkInterfaceId", "network_interface"},
		{"EgressOnlyInternetGatewayId", "egress_only_internet_gateway"},
	}
	for _, t := range targets {
		if id, _ := route[t.property].(string); id != "" {
			return t.targetType, id, nil
		}
	}
	return "", "", errors.Errorf("route %v carries no recognized target", route)
}

func (b *builder) mapBuckets() {
	policiesByBucket := map[string]mappingeval.ScannedResource{}
	for _, policy := range b.byType[typeBucketPolicy] {
		policiesByBucket[policy.Identifier] = policy
	}
	for _, bucket := range b.byType[typeBucket] {
		spec := map[string]any{"region": b.scan.Region}
		claims := []mappingeval.AccountResourceRef{bucket.Ref()}
		rationale := "a bucket is its own component instance; metadata.name must be the bucket name (the import recipe derives the import id from it)"
		if policy, exists := policiesByBucket[bucket.Identifier]; exists {
			if document, _ := policy.Properties["PolicyDocument"].(map[string]any); document != nil {
				spec["policy"] = document
			}
			claims = append(claims, policy.Ref())
			rationale += "; the bucket policy is a facet of the bucket"
		}
		// The name is NOT free here: the S3 import recipe derives the
		// import id from metadata.name, so the manifest name must be the
		// bucket name itself.
		b.emit("AwsS3Bucket", bucket.Identifier, spec, rationale, claims...)
	}
}

func (b *builder) mapQueues() {
	for _, queue := range b.byType[typeQueue] {
		spec := map[string]any{"region": b.scan.Region}
		setBool(spec, "fifoQueue", queue.Properties["FifoQueue"])
		setBool(spec, "contentBasedDeduplication", queue.Properties["ContentBasedDeduplication"])
		setString(spec, "deduplicationScope", queue.Properties["DeduplicationScope"])
		setString(spec, "fifoThroughputLimit", queue.Properties["FifoThroughputLimit"])
		setBool(spec, "sqsManagedSseEnabled", queue.Properties["SqsManagedSseEnabled"])
		setNumber(spec, "visibilityTimeoutSeconds", queue.Properties["VisibilityTimeout"])
		setNumber(spec, "messageRetentionSeconds", queue.Properties["MessageRetentionPeriod"])
		setNumber(spec, "receiveWaitTimeSeconds", queue.Properties["ReceiveMessageWaitTimeSeconds"])
		setNumber(spec, "delaySeconds", queue.Properties["DelaySeconds"])
		setNumber(spec, "maxMessageSizeBytes", queue.Properties["MaximumMessageSize"])
		b.emit("AwsSqsQueue", queueDisplayName(queue), spec,
			"a queue is its own component instance", queue.Ref())
	}
}

func (b *builder) mapTopics() {
	for _, topic := range b.byType[typeTopic] {
		spec := map[string]any{"region": b.scan.Region}
		setString(spec, "displayName", topic.Properties["DisplayName"])
		setString(spec, "tracingConfig", topic.Properties["TracingConfig"])
		setBool(spec, "fifoTopic", topic.Properties["FifoTopic"])
		setBool(spec, "contentBasedDeduplication", topic.Properties["ContentBasedDeduplication"])
		if version, _ := topic.Properties["SignatureVersion"].(string); version != "" {
			if parsed, err := strconv.Atoi(version); err == nil && parsed != 0 {
				spec["signatureVersion"] = float64(parsed)
			}
		}
		if policy, _ := topic.Properties["Policy"].(map[string]any); policy != nil {
			spec["policy"] = policy
		}
		b.emit("AwsSnsTopic", topicDisplayName(topic), spec,
			"a topic is its own component instance", topic.Ref())
	}
}

// recordUnmapped accounts for every scanned resource no instance claimed.
func (b *builder) recordUnmapped() {
	for _, r := range b.scan.Resources {
		if b.claimed[r.Ref()] {
			continue
		}
		b.unmapped = append(b.unmapped, &proposalv1.UnmappedAccountResource{
			TypeName:   r.TypeName,
			Identifier: r.Identifier,
			Reason:     unmappedReason(r),
		})
	}
	sort.Slice(b.unmapped, func(i, j int) bool {
		if b.unmapped[i].TypeName != b.unmapped[j].TypeName {
			return b.unmapped[i].TypeName < b.unmapped[j].TypeName
		}
		return b.unmapped[i].Identifier < b.unmapped[j].Identifier
	})
}

func unmappedReason(r mappingeval.ScannedResource) string {
	switch r.TypeName {
	case typeRouteTable:
		return "route table with no explicit subnet association (a VPC's main route table is implicitly created by AWS and not modeled as a kind)"
	case typeRTBAssociation:
		return "main route table association -- implicitly created by AWS, not modeled"
	case typeBucketPolicy:
		return "bucket policy whose bucket is outside this scan's region"
	default:
		return "no deterministic mapper covers this resource"
	}
}

// emit records one proposed instance and indexes its claims. metadata.env
// rides only when the partition engine assigned the claimed resources an
// environment (see instanceEnv) -- an unpartitioned instance honestly
// carries none.
func (b *builder) emit(kind, name string, spec map[string]any, rationale string, claims ...mappingeval.AccountResourceRef) {
	metadata := map[string]any{"name": name}
	if env := b.instanceEnv(claims); env != "" {
		metadata["env"] = env
	}
	// The apiVersion follows the kind's registry metadata, so proposals stay
	// stamped with the version the kind actually serves.
	apiVersion := catalogkindreflect.GroupVersion(catalogkindreflect.KindFromString(kind))
	if apiVersion == "" {
		// Every kind emitted here is one of this file's own mappers' AWS
		// kinds; an unresolvable name is a programming error.
		panic(fmt.Sprintf("baseline manifest kind %q does not resolve in the kind registry", kind))
	}
	manifest, err := structpb.NewStruct(map[string]any{
		"apiVersion": apiVersion,
		"kind":       kind,
		"metadata":   metadata,
		"spec":       spec,
	})
	if err != nil {
		// structpb.NewStruct only fails on non-JSON-representable values;
		// every value here originates from decoded JSON or literals.
		panic(fmt.Sprintf("baseline manifest for %s %q not Struct-representable: %v", kind, name, err))
	}
	proposed := &proposalv1.ProposedResource{Manifest: manifest, Rationale: rationale}
	for _, claim := range claims {
		proposed.Claims = append(proposed.Claims, &proposalv1.AccountResourceClaim{
			TypeName:   claim.TypeName,
			Identifier: claim.Identifier,
		})
		b.claimed[claim] = true
		b.nameByRef[claim] = name
	}
	b.resources = append(b.resources, proposed)
}

// refOrLiteral emits a value_from reference when the identified producer is
// itself a proposed instance, and a literal otherwise -- the same rule the
// production contract enforces (a ref may only target something in the
// mapped set; anything else stays a literal, visible in review).
func (b *builder) refOrLiteral(typeName, identifier, kind, fieldPath string) map[string]any {
	ref := mappingeval.AccountResourceRef{TypeName: typeName, Identifier: identifier}
	if producerName, proposed := b.nameByRef[ref]; proposed {
		return map[string]any{"valueFrom": map[string]any{
			"kind":      kind,
			"name":      producerName,
			"fieldPath": fieldPath,
		}}
	}
	return map[string]any{"value": identifier}
}

// displayName names an instance after its Name tag when the account's
// owners tagged it, falling back to the cloud identifier. Names carry no
// score weight (instances match by claims), but a proposal a human reviews
// should read like their infrastructure.
func (b *builder) displayName(r mappingeval.ScannedResource) string {
	if tags, ok := r.Properties["Tags"].([]any); ok {
		for _, raw := range tags {
			tag, _ := raw.(map[string]any)
			if tag == nil {
				continue
			}
			if key, _ := tag["Key"].(string); key == "Name" {
				if value, _ := tag["Value"].(string); value != "" {
					return value
				}
			}
		}
	}
	return r.Identifier
}

func queueDisplayName(queue mappingeval.ScannedResource) string {
	name, _ := queue.Properties["QueueName"].(string)
	if name == "" {
		segments := strings.Split(queue.Identifier, "/")
		name = segments[len(segments)-1]
	}
	return strings.TrimSuffix(name, ".fifo")
}

func topicDisplayName(topic mappingeval.ScannedResource) string {
	if name, _ := topic.Properties["TopicName"].(string); name != "" {
		return name
	}
	segments := strings.Split(topic.Identifier, ":")
	return segments[len(segments)-1]
}

// firstAttachmentVpcID reads the VPC id from an internet gateway's enriched
// Attachments (an internet gateway attaches to at most one VPC).
func firstAttachmentVpcID(properties map[string]any) string {
	attachments, _ := properties["Attachments"].([]any)
	for _, raw := range attachments {
		attachment, _ := raw.(map[string]any)
		if attachment == nil {
			continue
		}
		if vpcID, _ := attachment["VpcId"].(string); vpcID != "" {
			return vpcID
		}
	}
	return ""
}

func setString(spec map[string]any, key string, value any) {
	if s, _ := value.(string); s != "" {
		spec[key] = s
	}
}

func setBool(spec map[string]any, key string, value any) {
	if v, ok := value.(bool); ok && v {
		spec[key] = true
	}
}

func setNumber(spec map[string]any, key string, value any) {
	if v, ok := value.(float64); ok && v != 0 {
		spec[key] = v
	}
}

//go:build !codegen

// Package mappingeval_test proves the harness end to end, offline: the
// recorded-shape scan fixture drives the deterministic baseline, whose
// proposal must score PERFECT against the ground truth assembled from the
// real network-staples suite manifests -- and hand-mutated proposals prove
// every scoring axis DISCRIMINATES (an exam nothing can fail is not an
// exam). Runs creds-free in `make test`; the live e2e lane re-proves the
// same chain against a real account.
package mappingeval_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	proposalv1 "github.com/plantonhq/planton/iac/importmappingproposal/v1"
	"github.com/plantonhq/planton/pkg/iac/mappingeval"
	"github.com/plantonhq/planton/pkg/iac/mappingeval/baseline"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// Fixture identifiers -- must match testdata/network-staples-scan.json.
const (
	fixtureVpcID      = "vpc-0aaa111fixture0001"
	fixtureIgwID      = "igw-0bbb222fixture0001"
	fixtureSubnetID   = "subnet-0ccc333fixture01"
	fixtureRtbID      = "rtb-0ddd444fixture0001"
	fixtureRtbAssocID = "rtbassoc-0eee555fix01"
	fixtureBucketName = "planton-oss-e2e-awss3bucket-smoke"
	fixtureQueueURL   = "https://sqs.us-west-2.amazonaws.com/123456789012/planton-oss-e2e-awssqsqueue-fifo-full-surface.fifo"
	fixtureTopicArn   = "arn:aws:sns:us-west-2:123456789012:planton-oss-e2e-awssnstopic-standard-topic"
)

func TestNetworkStaplesSuiteLoads(t *testing.T) {
	root := repoRoot(t)
	suite := loadSuite(t, root)
	if len(suite.Members) != 6 {
		t.Fatalf("network-staples suite has %d members, want 6", len(suite.Members))
	}
}

func TestBaselineScoresPerfectOnFixture(t *testing.T) {
	root := repoRoot(t)
	gt := buildGroundTruth(t, root)
	report := scoreBaseline(t, root, gt, nil)

	if !report.Perfect() {
		reportJSON, _ := json.MarshalIndent(report, "", "  ")
		t.Fatalf("baseline must score PERFECT on the fixture; it did not.\n%s\n\nfull report:\n%s",
			report.Summary(), string(reportJSON))
	}
	// The fixture deliberately carries out-of-universe noise (a default
	// VPC, main route tables); a perfect score must coexist with it being
	// REPORTED, proving the scored-universe rule filters rather than hides.
	if len(report.Coverage.OutOfUniverseClaims) == 0 {
		t.Error("expected the default VPC's claim to be reported as out-of-universe")
	}
	// This suite does NOT declare partition grading (its members' env is
	// e2e seeding bookkeeping with no scan-visible signal -- unbeatable
	// debt, not an answer key), so the axis must stay vacuously ungraded
	// and the PERFECT pin must hold without it.
	if report.Partition.Graded || report.Partition.GroundTruthInstances != 0 {
		t.Fatalf("network-staples must not grade partition: %+v", report.Partition)
	}
}

func TestScorerDetectsMisgrouping(t *testing.T) {
	root := repoRoot(t)
	gt := buildGroundTruth(t, root)
	report := scoreBaseline(t, root, gt, func(p *proposalv1.ImportMappingProposal) {
		// Move the subnet's route-table claim onto the VPC instance: same
		// resources accounted for, wrong grouping.
		moveClaim(t, p, fixtureRtbID, fixtureVpcID)
	})
	if report.Perfect() {
		t.Fatal("misgrouped proposal must not score perfect")
	}
	if len(report.Grouping.Misassigned) != 1 {
		t.Fatalf("want exactly 1 misassigned resource, got %d (%v)", len(report.Grouping.Misassigned), report.Grouping.Misassigned)
	}
	if report.Grouping.Misassigned[0].Resource.Identifier != fixtureRtbID {
		t.Fatalf("misassigned the wrong resource: %v", report.Grouping.Misassigned[0])
	}
}

func TestScorerDetectsLiteralWhereRefExpected(t *testing.T) {
	root := repoRoot(t)
	gt := buildGroundTruth(t, root)
	report := scoreBaseline(t, root, gt, func(p *proposalv1.ImportMappingProposal) {
		// Freeze the subnet's vpc reference into a literal -- the classic
		// import mistake that silently severs the graph.
		subnet := resourceClaiming(t, p, fixtureSubnetID)
		spec := subnet.GetManifest().GetFields()["spec"].GetStructValue()
		spec.Fields["vpcId"] = structpb.NewStructValue(&structpb.Struct{Fields: map[string]*structpb.Value{
			"value": structpb.NewStringValue(fixtureVpcID),
		}})
	})
	if report.Perfect() {
		t.Fatal("literal-instead-of-ref must not score perfect")
	}
	if len(report.Refs.MissingEdges) != 1 || report.Refs.MissingEdges[0].FieldPath != "vpc_id" {
		t.Fatalf("want exactly the vpc_id edge missing, got %v", report.Refs.MissingEdges)
	}
}

func TestScorerDetectsWrongSpecValue(t *testing.T) {
	root := repoRoot(t)
	gt := buildGroundTruth(t, root)
	report := scoreBaseline(t, root, gt, func(p *proposalv1.ImportMappingProposal) {
		subnet := resourceClaiming(t, p, fixtureSubnetID)
		spec := subnet.GetManifest().GetFields()["spec"].GetStructValue()
		spec.Fields["cidrBlock"] = structpb.NewStringValue("10.0.200.0/24")
	})
	if report.Perfect() {
		t.Fatal("wrong spec value must not score perfect")
	}
	if len(report.Spec.Mismatched) != 1 || report.Spec.Mismatched[0].FieldPath != "cidr_block" {
		t.Fatalf("want exactly cidr_block mismatched, got %v", report.Spec.Mismatched)
	}
}

func TestScorerDetectsDuplicateClaim(t *testing.T) {
	root := repoRoot(t)
	gt := buildGroundTruth(t, root)
	report := scoreBaseline(t, root, gt, func(p *proposalv1.ImportMappingProposal) {
		vpc := resourceClaiming(t, p, fixtureVpcID)
		vpc.Claims = append(vpc.Claims, &proposalv1.AccountResourceClaim{
			TypeName:   "AWS::EC2::Subnet",
			Identifier: fixtureSubnetID,
		})
	})
	if report.Perfect() {
		t.Fatal("duplicate claim must not score perfect")
	}
	if len(report.Grouping.DuplicateClaims) != 1 || report.Grouping.DuplicateClaims[0].Identifier != fixtureSubnetID {
		t.Fatalf("want exactly the subnet double-claimed, got %v", report.Grouping.DuplicateClaims)
	}
}

func TestScorerDetectsUnaccountedResource(t *testing.T) {
	root := repoRoot(t)
	gt := buildGroundTruth(t, root)
	report := scoreBaseline(t, root, gt, func(p *proposalv1.ImportMappingProposal) {
		// Drop the topic instance without declaring it unmapped -- the
		// silent gap, the worst coverage class.
		kept := p.GetSpec().GetResources()[:0]
		for _, r := range p.GetSpec().GetResources() {
			if !claims(r, fixtureTopicArn) {
				kept = append(kept, r)
			}
		}
		p.Spec.Resources = kept
	})
	if report.Perfect() {
		t.Fatal("an unaccounted universe resource must not score perfect")
	}
	if len(report.Coverage.Unaccounted) != 1 || report.Coverage.Unaccounted[0].Identifier != fixtureTopicArn {
		t.Fatalf("want exactly the topic unaccounted, got %v", report.Coverage.Unaccounted)
	}
	if len(report.Grouping.Unclaimed) != 1 {
		t.Fatalf("want exactly the topic unclaimed in grouping, got %v", report.Grouping.Unclaimed)
	}
}

func TestScorerCountsUnproposedInstanceEdges(t *testing.T) {
	root := repoRoot(t)
	gt := buildGroundTruth(t, root)
	report := scoreBaseline(t, root, gt, func(p *proposalv1.ImportMappingProposal) {
		// Drop the subnet instance entirely -- it carries two of the
		// suite's three value_from edges (its vpc and its route's gateway
		// target). Skipping a resource must never let a proposer escape
		// its ref debt: the spec axis already counts an unproposed
		// instance's leaves as missing, and the refs axis must mirror it
		// rather than silently shrinking the denominator.
		kept := p.GetSpec().GetResources()[:0]
		for _, r := range p.GetSpec().GetResources() {
			if !claims(r, fixtureSubnetID) {
				kept = append(kept, r)
			}
		}
		p.Spec.Resources = kept
	})
	if report.Perfect() {
		t.Fatal("dropping a ref-carrying instance must not score perfect")
	}
	if report.Refs.GroundTruthEdges != 3 {
		t.Fatalf("unproposed instance's edges must stay in the denominator: want 3 ground-truth edges, got %d", report.Refs.GroundTruthEdges)
	}
	if len(report.Refs.MissingEdges) != 2 {
		t.Fatalf("want the dropped subnet's 2 edges missing, got %v", report.Refs.MissingEdges)
	}
}

func TestScorerFlagsNameDerivabilityBreak(t *testing.T) {
	root := repoRoot(t)
	gt := buildGroundTruth(t, root)
	report := scoreBaseline(t, root, gt, func(p *proposalv1.ImportMappingProposal) {
		// Rename the bucket instance: every other axis still passes, but
		// the S3 import recipe derives the import id from metadata.name,
		// so this break must surface.
		bucket := resourceClaiming(t, p, fixtureBucketName)
		metadata := bucket.GetManifest().GetFields()["metadata"].GetStructValue()
		metadata.Fields["name"] = structpb.NewStringValue("my-nicely-renamed-bucket")
	})
	if report.Perfect() {
		t.Fatal("a name-derivability break must not score perfect")
	}
	if len(report.NameDerivability) != 1 || report.NameDerivability[0].Identifier != fixtureBucketName {
		t.Fatalf("want exactly the bucket flagged, got %v", report.NameDerivability)
	}
}

func TestProposalContractRejectsDanglingRef(t *testing.T) {
	root := repoRoot(t)
	proposal := proposeFromFixture(t, root)
	// Point the subnet's vpc reference at an instance the proposal never
	// proposes.
	subnet := resourceClaiming(t, proposal, fixtureSubnetID)
	spec := subnet.GetManifest().GetFields()["spec"].GetStructValue()
	spec.Fields["vpcId"].GetStructValue().
		Fields["valueFrom"].GetStructValue().
		Fields["name"] = structpb.NewStringValue("a-vpc-nobody-proposed")

	if _, err := mappingeval.ParseProposal(proposal); err == nil {
		t.Fatal("a dangling value_from reference must fail the contract")
	}
}

func TestProposalContractRejectsClaimlessInstance(t *testing.T) {
	root := repoRoot(t)
	proposal := proposeFromFixture(t, root)
	resourceClaiming(t, proposal, fixtureBucketName).Claims = nil

	if _, err := mappingeval.ParseProposal(proposal); err == nil {
		t.Fatal("an instance claiming nothing must fail the contract")
	}
}

func TestSuiteRejectsForwardReference(t *testing.T) {
	root := repoRoot(t)
	// The routed subnet references the VPC and internet-gateway fixtures;
	// listing it FIRST makes those references forward, which must fail --
	// deploy order is list order.
	suiteYAML := `apiVersion: qa.planton.dev/v1
kind: MappingEvalSuite
metadata:
  name: forward-ref
spec:
  members:
    - kind_dir: awssubnet
      manifest_path: catalog/aws/awssubnet/v1alpha1/e2e/scenarios/routed.yaml
    - kind_dir: awsvpc
      manifest_path: catalog/aws/awsvpc/v1alpha1/e2e/prerequisite.yaml
  scan_scope:
    region: us-west-2
`
	suitePath := filepath.Join(t.TempDir(), "forward-ref.yaml")
	if err := os.WriteFile(suitePath, []byte(suiteYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := mappingeval.LoadSuite(root, suitePath); err == nil {
		t.Fatal("a suite whose member references a LATER member must fail to load")
	}
}

// --- helpers ---------------------------------------------------------------

// loadFixtureScan reads the recorded-shape scan (shapes verified against
// live Cloud Control responses; identifiers synthetic so the fixture is
// stable forever).
func loadFixtureScan(t *testing.T) *mappingeval.Scan {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(thisDir(t), "testdata", "network-staples-scan.json"))
	if err != nil {
		t.Fatal(err)
	}
	scan := &mappingeval.Scan{}
	if err := json.Unmarshal(raw, scan); err != nil {
		t.Fatal(err)
	}
	return scan
}

func loadSuite(t *testing.T, root string) *mappingeval.LoadedSuite {
	t.Helper()
	suite, err := mappingeval.LoadSuite(root, filepath.Join(root, "catalog/aws/aa_eval/suites/network-staples.yaml"))
	if err != nil {
		t.Fatalf("loading network-staples suite: %v", err)
	}
	return suite
}

// buildGroundTruth assembles the answer key exactly as the live deployer
// does -- the suite's real manifests -- with the claims a deploy of the
// fixture identifiers would record.
func buildGroundTruth(t *testing.T, root string) *mappingeval.GroundTruth {
	t.Helper()
	claimsByKind := map[string][]mappingeval.AccountResourceRef{
		"awsvpc":             {{TypeName: "AWS::EC2::VPC", Identifier: fixtureVpcID}},
		"awsinternetgateway": {{TypeName: "AWS::EC2::InternetGateway", Identifier: fixtureIgwID}},
		"awssubnet": {
			{TypeName: "AWS::EC2::Subnet", Identifier: fixtureSubnetID},
			{TypeName: "AWS::EC2::RouteTable", Identifier: fixtureRtbID},
			{TypeName: "AWS::EC2::SubnetRouteTableAssociation", Identifier: fixtureRtbAssocID},
		},
		"awss3bucket": {{TypeName: "AWS::S3::Bucket", Identifier: fixtureBucketName}},
		"awssqsqueue": {{TypeName: "AWS::SQS::Queue", Identifier: fixtureQueueURL}},
		"awssnstopic": {{TypeName: "AWS::SNS::Topic", Identifier: fixtureTopicArn}},
	}
	gt := &mappingeval.GroundTruth{}
	for _, member := range loadSuite(t, root).Members {
		claims, known := claimsByKind[member.KindDir]
		if !known {
			t.Fatalf("suite member %s has no fixture claims -- update the test fixture alongside the suite", member.KindDir)
		}
		gt.Instances = append(gt.Instances, mappingeval.GroundTruthInstance{
			KindDir:  member.KindDir,
			Kind:     member.Kind,
			Name:     member.Name,
			Manifest: member.Manifest,
			Claims:   claims,
		})
	}
	return gt
}

func proposeFromFixture(t *testing.T, root string) *proposalv1.ImportMappingProposal {
	t.Helper()
	scan := loadFixtureScan(t)
	// The offline pipeline mirrors the live lane exactly: fingerprints are
	// redacted between scan and proposer, so the exam's honesty is a
	// property of the pipeline, not of the recorded fixture.
	mappingeval.RedactSeedFingerprints(scan)
	proposal, err := baseline.Propose(scan)
	if err != nil {
		t.Fatalf("baseline proposer: %v", err)
	}
	return proposal
}

// scoreBaseline runs the full offline chain: baseline proposal (optionally
// mutated to inject a defect), contract validation, scoring.
func scoreBaseline(t *testing.T, root string, gt *mappingeval.GroundTruth, mutate func(*proposalv1.ImportMappingProposal)) *mappingeval.Report {
	t.Helper()
	proposal := proposeFromFixture(t, root)
	if mutate != nil {
		mutate(proposal)
	}
	loaded, err := mappingeval.ParseProposal(proposal)
	if err != nil {
		t.Fatalf("proposal violates the contract: %v", err)
	}
	kindDirs := []string{"awsvpc", "awsinternetgateway", "awssubnet", "awss3bucket", "awssqsqueue", "awssnstopic"}
	opts, err := mappingeval.ScoreOptionsFromCatalog(root, "aws", kindDirs)
	if err != nil {
		t.Fatalf("score options: %v", err)
	}
	return mappingeval.Score(gt, loaded, opts)
}

// resourceClaiming finds the proposed resource claiming an identifier.
func resourceClaiming(t *testing.T, p *proposalv1.ImportMappingProposal, identifier string) *proposalv1.ProposedResource {
	t.Helper()
	for _, r := range p.GetSpec().GetResources() {
		if claims(r, identifier) {
			return r
		}
	}
	t.Fatalf("no proposed resource claims %q", identifier)
	return nil
}

func claims(r *proposalv1.ProposedResource, identifier string) bool {
	for _, claim := range r.GetClaims() {
		if claim.GetIdentifier() == identifier {
			return true
		}
	}
	return false
}

// moveClaim moves one claim between the instances claiming from/to anchors.
func moveClaim(t *testing.T, p *proposalv1.ImportMappingProposal, moveIdentifier, ontoInstanceClaiming string) {
	t.Helper()
	source := resourceClaiming(t, p, moveIdentifier)
	target := resourceClaiming(t, p, ontoInstanceClaiming)
	var moved *proposalv1.AccountResourceClaim
	kept := source.Claims[:0]
	for _, claim := range source.Claims {
		if claim.GetIdentifier() == moveIdentifier {
			moved = claim
			continue
		}
		kept = append(kept, claim)
	}
	source.Claims = kept
	target.Claims = append(target.Claims, proto.Clone(moved).(*proposalv1.AccountResourceClaim))
}

func thisDir(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Dir(thisFile)
}

func repoRoot(t *testing.T) string {
	t.Helper()
	dir := thisDir(t)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found walking up from test file")
		}
		dir = parent
	}
}

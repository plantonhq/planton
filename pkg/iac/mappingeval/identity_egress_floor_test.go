//go:build !codegen

package mappingeval_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	proposalv1 "github.com/plantonhq/planton/iac/importmappingproposal/v1"
	"github.com/plantonhq/planton/pkg/iac/mappingeval"
	"github.com/plantonhq/planton/pkg/iac/mappingeval/baseline"
	"google.golang.org/protobuf/types/known/structpb"
)

// The identity-and-egress suite is the COVERAGE exam: with it, every kind
// that carries an import recipe appears in at least one exam. Like the
// messy-account floor, every number asserted here is load-bearing -- a
// drift is a harness, baseline, or suite regression. What is distinct
// here:
//
//   - the NAT gateway and its nat_gateway route target (the first
//     non-internet-gateway route target in any exam) are baseline blind
//     spots by design, so both of their edges are owed at the floor;
//   - the IAM role is the first GLOBAL-service member -- the fixture
//     carries service-linked and foreign roles as noise, proving
//     universe-only scoring holds on a tier where noise is the norm.
//
// Fixture identifiers -- must match testdata/identity-and-egress-scan.json.
const (
	hubVpcID       = "vpc-0hub111fixture0001"
	hubNatSubnetID = "subnet-0hub222natfix001"
	hubNatID       = "nat-0hub333egress0001"
	hubAppSubnetID = "subnet-0hub444appfix001"
	hubRtbID       = "rtb-0hub555appfix001"
	hubRtbAssocID  = "rtbassoc-0hub666afx01"
	hubRoleName    = "hub-ci-deployer"
)

// identityEgressKinds mirrors the suite's member kinds
// (deduplicated) for ScoreOptionsFromCatalog.
var identityEgressKinds = []string{
	"awsvpc", "awssubnet", "awsnatgateway", "awsiamrole",
}

func TestIdentityAndEgressSuiteLoads(t *testing.T) {
	root := repoRoot(t)
	suite := loadIdentityEgressSuite(t, root)
	if len(suite.Members) != 5 {
		t.Fatalf("identity-and-egress suite has %d members, want 5", len(suite.Members))
	}
}

// TestBaselineFloorOnIdentityAndEgress pins the deterministic baseline's
// imperfect score on the identity-and-egress fixture -- the recorded floor.
func TestBaselineFloorOnIdentityAndEgress(t *testing.T) {
	root := repoRoot(t)
	gt := buildIdentityEgressGroundTruth(t, root)
	report := scoreIdentityEgressBaseline(t, root, gt)

	if report.Perfect() {
		t.Fatal("the identity-and-egress suite carries kinds the baseline deliberately does not map; a perfect score means the exam lost its headroom")
	}

	// Grouping: the network tier (VPC + both subnets, incl. the routed
	// subnet's table and association) lands correctly; the NAT gateway and
	// the role -- the two kinds this suite exists to examine -- are
	// unclaimed. Nothing is misassigned or double-claimed.
	if report.Grouping.UniverseSize != 7 {
		t.Fatalf("universe size drifted: want 7, got %d", report.Grouping.UniverseSize)
	}
	if report.Grouping.Correct != 5 || report.Grouping.InUniverseClaims != 5 {
		t.Fatalf("covered-kind grouping floor drifted: want 5 correct / 5 in-universe claims, got %d / %d",
			report.Grouping.Correct, report.Grouping.InUniverseClaims)
	}
	if len(report.Grouping.Misassigned) != 0 || len(report.Grouping.DuplicateClaims) != 0 {
		t.Fatalf("the baseline must never misassign or double-claim: %v / %v",
			report.Grouping.Misassigned, report.Grouping.DuplicateClaims)
	}
	uncovered := []mappingeval.AccountResourceRef{
		{TypeName: "AWS::EC2::NatGateway", Identifier: hubNatID},
		{TypeName: "AWS::IAM::Role", Identifier: hubRoleName},
	}
	assertRefsEqual(t, "unclaimed", report.Grouping.Unclaimed, uncovered)

	// Coverage: the two uncovered resources are DECLARED unmapped, never
	// silently dropped. The global-tier noise (service-linked and foreign
	// roles) plus the default network are out-of-universe: the baseline
	// proposes the default VPC and its subnet (they are its covered kinds)
	// and declares every foreign role unmapped -- all informational, none
	// scored. Zero unaccounted is part of the floor.
	if len(report.Coverage.Unaccounted) != 0 {
		t.Fatalf("the baseline accounts for everything it sees; unaccounted: %v", report.Coverage.Unaccounted)
	}
	assertRefsEqual(t, "unmapped-in-universe", report.Coverage.UnmappedInUniverse, uncovered)
	if len(report.Coverage.OutOfUniverseClaims) != 2 {
		t.Fatalf("want the default VPC + its subnet reported as out-of-universe claims, got %v",
			report.Coverage.OutOfUniverseClaims)
	}

	// Refs: the two vpcId edges inside covered territory hold; the two
	// edges that touch the NAT gateway -- its own subnet_id and the app
	// subnet's nat_gateway route target (the baseline wires only
	// internet-gateway targets) -- are owed. Nothing is wired WRONG.
	if report.Refs.GroundTruthEdges != 4 || report.Refs.CorrectEdges != 2 {
		t.Fatalf("refs floor drifted: want 2/4 edges, got %d/%d",
			report.Refs.CorrectEdges, report.Refs.GroundTruthEdges)
	}
	wantMissing := map[string]string{
		"hub-egress-nat": "subnet_id",
		"hub-app-subnet": "routes[0].target_id",
	}
	if len(report.Refs.MissingEdges) != len(wantMissing) {
		t.Fatalf("want %d missing edges, got %v", len(wantMissing), report.Refs.MissingEdges)
	}
	for _, finding := range report.Refs.MissingEdges {
		if wantMissing[finding.Instance] != finding.FieldPath {
			t.Fatalf("unexpected missing edge %s on %s (want %v)", finding.FieldPath, finding.Instance, wantMissing)
		}
	}
	if len(report.Refs.WrongTargetEdges) != 0 || len(report.Refs.UnexpectedEdges) != 0 {
		t.Fatalf("the baseline never wires a WRONG edge: %v / %v",
			report.Refs.WrongTargetEdges, report.Refs.UnexpectedEdges)
	}

	// Spec: zero mismatches; every missing leaf belongs to the two
	// instances the baseline never proposed. The exact leaf counts are
	// pinned so silent spec drift in members, kinds, or scorer surfaces
	// here.
	if len(report.Spec.Mismatched) != 0 {
		t.Fatalf("the baseline never writes a wrong value: %v", report.Spec.Mismatched)
	}
	unproposed := map[string]bool{
		"hub-egress-nat":  true,
		"hub-ci-deployer": true,
	}
	for _, finding := range report.Spec.Missing {
		if !unproposed[finding.Instance] {
			t.Fatalf("a PROPOSED instance is missing leaf %s on %s -- the floor only tolerates whole-instance gaps",
				finding.FieldPath, finding.Instance)
		}
	}
	if report.Spec.GroundTruthLeaves != 21 || report.Spec.Matched != 12 {
		t.Fatalf("spec floor drifted: want 12/21 leaves matched, got %d/%d\nmissing: %v",
			report.Spec.Matched, report.Spec.GroundTruthLeaves, report.Spec.Missing)
	}

	// The role is a name-derived-identity kind (like S3), but the baseline
	// never proposes it, and the check applies to matched instances only
	// -- zero findings is part of the floor.
	if len(report.NameDerivability) != 0 {
		t.Fatalf("no matched instance breaks name-derived identity: %v", report.NameDerivability)
	}

	// This suite does NOT declare partition grading: its members' env
	// ("ops") leaves no scan-visible trace once seeding fingerprints are
	// redacted -- an answer key no proposer could recover is not a fair
	// exam. The axis must stay vacuously ungraded and the floor must hold
	// without it.
	if report.Partition.Graded || report.Partition.GroundTruthInstances != 0 {
		t.Fatalf("identity-and-egress must not grade partition: %+v", report.Partition)
	}
}

// TestScorerComparesRepeatedRefLiterals proves the repeated-reference leaf
// discriminates: a proposed role whose managed policy list disagrees with
// the ground truth must produce a mismatch finding, not silently pass.
// (The floor test only exercises the MISSING arm of this leaf -- the
// baseline never proposes roles -- so without this mutation the wrong-value
// arm would be unproven.)
func TestScorerComparesRepeatedRefLiterals(t *testing.T) {
	root := repoRoot(t)
	gt := buildIdentityEgressGroundTruth(t, root)

	scan := loadIdentityEgressFixtureScan(t)
	mappingeval.RedactSeedFingerprints(scan)
	proposal, err := baseline.Propose(scan)
	if err != nil {
		t.Fatalf("baseline proposer: %v", err)
	}
	// Hand the baseline the role mapping it cannot produce itself, with the
	// WRONG managed policy attached.
	manifest, err := structpb.NewStruct(map[string]any{
		"apiVersion": "aws.planton.dev/v1alpha1",
		"kind":       "AwsIamRole",
		"metadata":   map[string]any{"name": hubRoleName},
		"spec": map[string]any{
			"region": "us-west-2",
			"managedPolicyArns": []any{
				map[string]any{"value": "arn:aws:iam::aws:policy/AdministratorAccess"},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	proposal.Spec.Resources = append(proposal.Spec.Resources, &proposalv1.ProposedResource{
		Manifest:  manifest,
		Rationale: "test mutation: role proposed with the wrong managed policy",
		Claims: []*proposalv1.AccountResourceClaim{
			{TypeName: "AWS::IAM::Role", Identifier: hubRoleName},
		},
	})
	// The mutated proposal no longer declares the role unmapped.
	kept := proposal.Spec.Unmapped[:0]
	for _, u := range proposal.Spec.Unmapped {
		if u.GetTypeName() == "AWS::IAM::Role" && u.GetIdentifier() == hubRoleName {
			continue
		}
		kept = append(kept, u)
	}
	proposal.Spec.Unmapped = kept

	loaded, err := mappingeval.ParseProposal(proposal)
	if err != nil {
		t.Fatalf("mutated proposal violates the contract: %v", err)
	}
	opts, err := mappingeval.ScoreOptionsFromCatalog(root, "aws", identityEgressKinds)
	if err != nil {
		t.Fatalf("score options: %v", err)
	}
	report := mappingeval.Score(gt, loaded, opts)

	for _, finding := range report.Spec.Mismatched {
		if finding.Instance == hubRoleName && finding.FieldPath == "managed_policy_arns" {
			return
		}
	}
	t.Fatalf("a wrong managed-policy list must surface as a spec mismatch; mismatched: %v", report.Spec.Mismatched)
}

// --- identity-and-egress helpers ---------------------------------------------

func loadIdentityEgressSuite(t *testing.T, root string) *mappingeval.LoadedSuite {
	t.Helper()
	suite, err := mappingeval.LoadSuite(root, filepath.Join(root, "catalog/aws/aa_eval/suites/identity-and-egress.yaml"))
	if err != nil {
		t.Fatalf("loading identity-and-egress suite: %v", err)
	}
	return suite
}

func loadIdentityEgressFixtureScan(t *testing.T) *mappingeval.Scan {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(thisDir(t), "testdata", "identity-and-egress-scan.json"))
	if err != nil {
		t.Fatal(err)
	}
	scan := &mappingeval.Scan{}
	if err := json.Unmarshal(raw, scan); err != nil {
		t.Fatal(err)
	}
	return scan
}

// buildIdentityEgressGroundTruth assembles the answer key from the suite's
// real manifests with the claims a deploy of the fixture identifiers would
// record. Keyed by member NAME (the suite carries two subnets).
func buildIdentityEgressGroundTruth(t *testing.T, root string) *mappingeval.GroundTruth {
	t.Helper()
	claimsByName := map[string][]mappingeval.AccountResourceRef{
		"hub-vpc":        {{TypeName: "AWS::EC2::VPC", Identifier: hubVpcID}},
		"hub-nat-subnet": {{TypeName: "AWS::EC2::Subnet", Identifier: hubNatSubnetID}},
		"hub-egress-nat": {{TypeName: "AWS::EC2::NatGateway", Identifier: hubNatID}},
		"hub-app-subnet": {
			{TypeName: "AWS::EC2::Subnet", Identifier: hubAppSubnetID},
			{TypeName: "AWS::EC2::RouteTable", Identifier: hubRtbID},
			{TypeName: "AWS::EC2::SubnetRouteTableAssociation", Identifier: hubRtbAssocID},
		},
		hubRoleName: {{TypeName: "AWS::IAM::Role", Identifier: hubRoleName}},
	}
	gt := &mappingeval.GroundTruth{}
	for _, member := range loadIdentityEgressSuite(t, root).Members {
		claims, known := claimsByName[member.Name]
		if !known {
			t.Fatalf("suite member %q has no fixture claims -- update the test fixture alongside the suite", member.Name)
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

func scoreIdentityEgressBaseline(t *testing.T, root string, gt *mappingeval.GroundTruth) *mappingeval.Report {
	t.Helper()
	scan := loadIdentityEgressFixtureScan(t)
	mappingeval.RedactSeedFingerprints(scan)
	proposal, err := baseline.Propose(scan)
	if err != nil {
		t.Fatalf("baseline proposer: %v", err)
	}
	loaded, err := mappingeval.ParseProposal(proposal)
	if err != nil {
		t.Fatalf("baseline proposal violates the contract: %v", err)
	}
	opts, err := mappingeval.ScoreOptionsFromCatalog(root, "aws", identityEgressKinds)
	if err != nil {
		t.Fatalf("score options: %v", err)
	}
	return mappingeval.Score(gt, loaded, opts)
}

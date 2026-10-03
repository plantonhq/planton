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

// The messy-account suite is the exam with HEADROOM: where network-staples
// pins the baseline to a PERFECT score (proving the instrument), this suite
// pins the baseline to a specific IMPERFECT score -- the floor any smarter
// proposer must beat. Every number asserted here is load-bearing: a drift
// in the floor is a harness, baseline, or suite regression, exactly like a
// drop from perfect on the staples.
//
// Fixture identifiers -- must match testdata/messy-account-scan.json.
const (
	messyProdVpcID    = "vpc-0aaa111prodfix0001"
	messyStgVpcID     = "vpc-0bbb222stgfix00001"
	messyIgwID        = "igw-0ccc333prodfix001"
	messyProdSubnetID = "subnet-0ddd444prodfix01"
	messyStgSubnetID  = "subnet-0eee555stgfix01"
	messyRtbID        = "rtb-0ggg888prodfix001"
	messyRtbAssocID   = "rtbassoc-0hhh999pfx01"
	messySgID         = "sg-0iii000prodfix0001"
	messyKmsKeyID     = "1a2b3c4d-f1x7-4e5f-8a9b-000000000001"
	messyKmsAliasID   = "alias/orders-prod-data-key"
	messyTableName    = "orders-prod-table"
	messyQueueURL     = "https://sqs.us-west-2.amazonaws.com/123456789012/orders-prod-events"
	messyEcrName      = "orders-api"
	messyBucketName   = "orders-archive-planton-eval"
)

// messyKinds mirrors the suite's member kinds (deduplicated) for
// ScoreOptionsFromCatalog.
var messyKinds = []string{
	"awsvpc", "awsinternetgateway", "awssubnet", "awskmskey",
	"awssecuritygroup", "awsdynamodb", "awssqsqueue", "awsecrrepo",
	"awss3bucket",
}

func TestMessyAccountSuiteLoads(t *testing.T) {
	root := repoRoot(t)
	suite := loadMessySuite(t, root)
	if len(suite.Members) != 11 {
		t.Fatalf("messy-account suite has %d members, want 11", len(suite.Members))
	}
}

// TestBaselineFloorOnMessyAccount pins the deterministic baseline's
// imperfect score on the messy-account fixture -- the recorded floor.
func TestBaselineFloorOnMessyAccount(t *testing.T) {
	root := repoRoot(t)
	gt := buildMessyGroundTruth(t, root)
	report := scoreMessyBaseline(t, root, gt, nil)

	if report.Perfect() {
		t.Fatal("the messy-account suite exists BECAUSE the baseline cannot ace it; a perfect score means the exam lost its headroom")
	}

	// Grouping: every resource of a covered kind lands correctly; every
	// resource of an uncovered kind is unclaimed -- and NOTHING is
	// misassigned or double-claimed. The baseline is honest even where it
	// is blind.
	if report.Grouping.UniverseSize != 14 {
		t.Fatalf("universe size drifted: want 14, got %d", report.Grouping.UniverseSize)
	}
	if report.Grouping.Correct != 9 || report.Grouping.InUniverseClaims != 9 {
		t.Fatalf("covered-kind grouping floor drifted: want 9 correct / 9 in-universe claims, got %d / %d",
			report.Grouping.Correct, report.Grouping.InUniverseClaims)
	}
	if len(report.Grouping.Misassigned) != 0 || len(report.Grouping.DuplicateClaims) != 0 {
		t.Fatalf("the baseline must never misassign or double-claim: %v / %v",
			report.Grouping.Misassigned, report.Grouping.DuplicateClaims)
	}
	uncovered := []mappingeval.AccountResourceRef{
		{TypeName: "AWS::DynamoDB::Table", Identifier: messyTableName},
		{TypeName: "AWS::EC2::SecurityGroup", Identifier: messySgID},
		{TypeName: "AWS::ECR::Repository", Identifier: messyEcrName},
		{TypeName: "AWS::KMS::Alias", Identifier: messyKmsAliasID},
		{TypeName: "AWS::KMS::Key", Identifier: messyKmsKeyID},
	}
	assertRefsEqual(t, "unclaimed", report.Grouping.Unclaimed, uncovered)

	// Coverage: the uncovered tier is DECLARED unmapped, never silently
	// dropped -- zero unaccounted is part of the floor. The out-of-universe
	// noise (default VPC, its subnet) is reported, not scored.
	if len(report.Coverage.Unaccounted) != 0 {
		t.Fatalf("the baseline accounts for everything it sees; unaccounted: %v", report.Coverage.Unaccounted)
	}
	assertRefsEqual(t, "unmapped-in-universe", report.Coverage.UnmappedInUniverse, uncovered)
	if len(report.Coverage.OutOfUniverseClaims) != 2 {
		t.Fatalf("want the default VPC + its subnet reported as out-of-universe claims, got %v",
			report.Coverage.OutOfUniverseClaims)
	}

	// Refs: the wiring inside the baseline's covered territory holds (both
	// subnets to the RIGHT look-alike VPC, the route to the gateway), and
	// exactly three edges are beyond it -- the queue's kms edge on a
	// MATCHED instance, plus the two edges owed by instances it never
	// proposed.
	if report.Refs.GroundTruthEdges != 7 || report.Refs.CorrectEdges != 4 {
		t.Fatalf("refs floor drifted: want 4/7 edges, got %d/%d",
			report.Refs.CorrectEdges, report.Refs.GroundTruthEdges)
	}
	wantMissing := map[string]string{
		"orders-prod-events": "kms_key_id",
		"orders-prod-app-sg": "vpc_id",
		"orders-prod-table":  "server_side_encryption.kms_key_arn",
	}
	if len(report.Refs.MissingEdges) != len(wantMissing) {
		t.Fatalf("want %d missing edges, got %v", len(wantMissing), report.Refs.MissingEdges)
	}
	for _, edge := range report.Refs.MissingEdges {
		if wantMissing[edge.Instance] != edge.FieldPath {
			t.Fatalf("unexpected missing edge %s at %s (full: %v)", edge.Instance, edge.FieldPath, report.Refs.MissingEdges)
		}
	}
	if len(report.Refs.WrongTargetEdges) != 0 || len(report.Refs.UnexpectedEdges) != 0 {
		t.Fatalf("the baseline never wires a WRONG edge: %v / %v",
			report.Refs.WrongTargetEdges, report.Refs.UnexpectedEdges)
	}

	// Spec: zero mismatches (what the baseline reconstructs, it
	// reconstructs correctly); every missing leaf belongs to an instance
	// it never proposed. The exact leaf counts are pinned so silent spec
	// drift in members, kinds, or scorer surfaces here.
	if len(report.Spec.Mismatched) != 0 {
		t.Fatalf("the baseline never writes a wrong value: %v", report.Spec.Mismatched)
	}
	unproposed := map[string]bool{
		"orders-prod-data-key": true,
		"orders-prod-app-sg":   true,
		"orders-prod-table":    true,
		"orders-api":           true,
	}
	for _, finding := range report.Spec.Missing {
		if !unproposed[finding.Instance] {
			t.Fatalf("a PROPOSED instance is missing leaf %s on %s -- the floor only tolerates whole-instance gaps",
				finding.FieldPath, finding.Instance)
		}
	}
	// Floor re-pinned 21/72 -> 21/69 (2026-08-18): three ground-truth leaves
	// left member kinds' spec surfaces as their schemas evolved; the matched
	// count held and every missing leaf remains a whole-instance gap.
	if report.Spec.GroundTruthLeaves != 69 || report.Spec.Matched != 21 {
		t.Fatalf("spec floor drifted: want 21/69 leaves matched, got %d/%d\nmissing: %v",
			report.Spec.Matched, report.Spec.GroundTruthLeaves, report.Spec.Missing)
	}

	if len(report.NameDerivability) != 0 {
		t.Fatalf("the bucket keeps its name-derived identity: %v", report.NameDerivability)
	}

	// Partition: this is the ONE suite that grades environment assignment.
	// The untaught baseline recovers prod/staging from the member names'
	// env tokens (orders-prod-*, orders-stg-* -- stg normalizes to
	// staging); the token-less bucket plus the four unproposed instances
	// stay honestly owed. Zero wrong is part of the floor: the baseline
	// never assigns a WRONG environment, exactly as it never wires a
	// wrong edge.
	if !report.Partition.Graded {
		t.Fatal("the messy suite declares partition grading; the axis must be graded")
	}
	if report.Partition.GroundTruthInstances != 11 || report.Partition.Correct != 6 {
		t.Fatalf("partition floor drifted: want 6/11 environments assigned, got %d/%d\nmissing: %v\nwrong: %v",
			report.Partition.Correct, report.Partition.GroundTruthInstances,
			report.Partition.MissingEnv, report.Partition.WrongEnv)
	}
	if len(report.Partition.WrongEnv) != 0 || len(report.Partition.ExtraEnv) != 0 {
		t.Fatalf("the baseline must never assign a wrong or extra environment: %v / %v",
			report.Partition.WrongEnv, report.Partition.ExtraEnv)
	}
	owedEnv := map[string]bool{
		messyBucketName:        true, // no env token in its name
		"orders-prod-data-key": true, // unproposed (uncovered tier)
		"orders-prod-app-sg":   true,
		"orders-prod-table":    true,
		"orders-api":           true,
	}
	if len(report.Partition.MissingEnv) != len(owedEnv) {
		t.Fatalf("want %d owed environments, got %v", len(owedEnv), report.Partition.MissingEnv)
	}
	for _, finding := range report.Partition.MissingEnv {
		if !owedEnv[finding.Instance] {
			t.Fatalf("unexpected owed environment on %q (full: %v)", finding.Instance, report.Partition.MissingEnv)
		}
	}
}

// TestPartitionAxisDiscriminates proves the new axis catches each defect
// class specifically -- an exam nothing can fail is not an exam.
func TestPartitionAxisDiscriminates(t *testing.T) {
	root := repoRoot(t)
	gt := buildMessyGroundTruth(t, root)

	// Wrong environment: cross-file the prod VPC into staging. The worst
	// class -- it must surface as wrong-env, not blend into missing.
	report := scoreMessyBaseline(t, root, gt, func(p *proposalv1.ImportMappingProposal) {
		setManifestEnv(t, p, messyProdVpcID, "staging")
	})
	if len(report.Partition.WrongEnv) != 1 || report.Partition.WrongEnv[0].Instance != "orders-prod-vpc" {
		t.Fatalf("want exactly the prod VPC wrong-env, got %v", report.Partition.WrongEnv)
	}
	if report.Partition.Correct != 5 {
		t.Fatalf("correct must drop to 5, got %d", report.Partition.Correct)
	}

	// Missing environment: strip the staging VPC's env. Honest
	// unassignment where the answer was recoverable is owed, not wrong.
	report = scoreMessyBaseline(t, root, gt, func(p *proposalv1.ImportMappingProposal) {
		setManifestEnv(t, p, messyStgVpcID, "")
	})
	if len(report.Partition.WrongEnv) != 0 {
		t.Fatalf("a missing env must not read as wrong: %v", report.Partition.WrongEnv)
	}
	if len(report.Partition.MissingEnv) != 6 || report.Partition.Correct != 5 {
		t.Fatalf("want 6 missing / 5 correct after stripping the staging VPC's env, got %d missing / %d correct",
			len(report.Partition.MissingEnv), report.Partition.Correct)
	}
}

// setManifestEnv rewrites (or removes) metadata.env on the proposed
// instance claiming the identifier.
func setManifestEnv(t *testing.T, p *proposalv1.ImportMappingProposal, claimIdentifier, env string) {
	t.Helper()
	metadata := resourceClaiming(t, p, claimIdentifier).GetManifest().GetFields()["metadata"].GetStructValue()
	if env == "" {
		delete(metadata.Fields, "env")
		return
	}
	metadata.Fields["env"] = structpb.NewStringValue(env)
}

// --- messy-suite helpers ----------------------------------------------------

func loadMessySuite(t *testing.T, root string) *mappingeval.LoadedSuite {
	t.Helper()
	suite, err := mappingeval.LoadSuite(root, filepath.Join(root, "catalog/aws/aa_eval/suites/messy-account.yaml"))
	if err != nil {
		t.Fatalf("loading messy-account suite: %v", err)
	}
	return suite
}

func loadMessyFixtureScan(t *testing.T) *mappingeval.Scan {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(thisDir(t), "testdata", "messy-account-scan.json"))
	if err != nil {
		t.Fatal(err)
	}
	scan := &mappingeval.Scan{}
	if err := json.Unmarshal(raw, scan); err != nil {
		t.Fatal(err)
	}
	return scan
}

// buildMessyGroundTruth assembles the answer key from the suite's real
// manifests with the claims a deploy of the fixture identifiers would
// record. Keyed by member NAME (not kind -- this suite deliberately
// carries two VPCs and two subnets).
func buildMessyGroundTruth(t *testing.T, root string) *mappingeval.GroundTruth {
	t.Helper()
	claimsByName := map[string][]mappingeval.AccountResourceRef{
		"orders-prod-vpc": {{TypeName: "AWS::EC2::VPC", Identifier: messyProdVpcID}},
		"orders-prod-igw": {{TypeName: "AWS::EC2::InternetGateway", Identifier: messyIgwID}},
		"orders-prod-app-subnet": {
			{TypeName: "AWS::EC2::Subnet", Identifier: messyProdSubnetID},
			{TypeName: "AWS::EC2::RouteTable", Identifier: messyRtbID},
			{TypeName: "AWS::EC2::SubnetRouteTableAssociation", Identifier: messyRtbAssocID},
		},
		"orders-stg-vpc":        {{TypeName: "AWS::EC2::VPC", Identifier: messyStgVpcID}},
		"orders-stg-app-subnet": {{TypeName: "AWS::EC2::Subnet", Identifier: messyStgSubnetID}},
		"orders-prod-data-key": {
			{TypeName: "AWS::KMS::Key", Identifier: messyKmsKeyID},
			{TypeName: "AWS::KMS::Alias", Identifier: messyKmsAliasID},
		},
		"orders-prod-app-sg": {{TypeName: "AWS::EC2::SecurityGroup", Identifier: messySgID}},
		"orders-prod-table":  {{TypeName: "AWS::DynamoDB::Table", Identifier: messyTableName}},
		"orders-prod-events": {{TypeName: "AWS::SQS::Queue", Identifier: messyQueueURL}},
		"orders-api":         {{TypeName: "AWS::ECR::Repository", Identifier: messyEcrName}},
		messyBucketName:      {{TypeName: "AWS::S3::Bucket", Identifier: messyBucketName}},
	}
	gt := &mappingeval.GroundTruth{}
	for _, member := range loadMessySuite(t, root).Members {
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

func scoreMessyBaseline(t *testing.T, root string, gt *mappingeval.GroundTruth, mutate func(*proposalv1.ImportMappingProposal)) *mappingeval.Report {
	t.Helper()
	scan := loadMessyFixtureScan(t)
	mappingeval.RedactSeedFingerprints(scan)
	proposal, err := baseline.Propose(scan)
	if err != nil {
		t.Fatalf("baseline proposer: %v", err)
	}
	if mutate != nil {
		mutate(proposal)
	}
	loaded, err := mappingeval.ParseProposal(proposal)
	if err != nil {
		t.Fatalf("baseline proposal violates the contract: %v", err)
	}
	opts, err := mappingeval.ScoreOptionsFromCatalog(root, "aws", messyKinds)
	if err != nil {
		t.Fatalf("score options: %v", err)
	}
	// The partition axis is suite-declared, never assumed: the flag comes
	// off the loaded suite exactly as the live lane reads it.
	opts.GradeEnvironmentPartition = loadMessySuite(t, root).Suite.GetSpec().GetGradeEnvironmentPartition()
	return mappingeval.Score(gt, loaded, opts)
}

// assertRefsEqual compares a report's (sorted) ref list against the
// expected SET of refs.
func assertRefsEqual(t *testing.T, label string, got, want []mappingeval.AccountResourceRef) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s: want %d refs %v, got %v", label, len(want), want, got)
	}
	wantSet := map[mappingeval.AccountResourceRef]bool{}
	for _, ref := range want {
		wantSet[ref] = true
	}
	for _, ref := range got {
		if !wantSet[ref] {
			t.Fatalf("%s: unexpected ref %v (want %v)", label, ref, want)
		}
	}
}

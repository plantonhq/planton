//go:build e2e

package aws

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	evalrunner "github.com/plantonhq/planton/e2e/framework/mappingeval"
	"github.com/plantonhq/planton/pkg/iac/mappingeval"
	"github.com/plantonhq/planton/pkg/iac/mappingeval/baseline"
	"github.com/plantonhq/planton/pkg/iac/mappingeval/inventory"
	"google.golang.org/protobuf/encoding/protojson"
)

// MappingEvalEnvVar opts the mapping-eval lane in. Opt-in like the import
// round-trip: it deploys whole multi-kind suites (create-and-destroy)
// and scans the account, which is a deliberate, scheduled spend -- not part
// of every e2e invocation.
const MappingEvalEnvVar = "PLANTON_E2E_MAPPING_EVAL"

// TestMappingEval_NetworkStaples is the live proof of the mapping eval
// harness: seed the account from the network-staples suite (the answer
// key), scan it back blind through the read-only scanner, have the
// deterministic baseline propose the mapping, and machine-score the
// proposal. The baseline is PINNED to a perfect score on this suite -- any
// drop is a harness, recipe, or scanner regression, never model variance.
//
// Artifacts (scan, proposal, report) are written to
// PLANTON_E2E_MAPPING_EVAL_ARTIFACTS when set, so runs leave durable
// evidence.
func TestMappingEval_NetworkStaples(t *testing.T) {
	report := runMappingEvalChain(t, "network-staples")

	if !report.Perfect() {
		reportJSON, _ := json.MarshalIndent(report, "", "  ")
		t.Fatalf("baseline must score PERFECT on the seeded suite; it did not.\n%s\n\nfull report:\n%s",
			report.Summary(), string(reportJSON))
	}
}

// TestMappingEval_MessyAccount is the live proof of the exam WITH HEADROOM:
// the messy-account suite is designed so the deterministic baseline cannot
// ace it, and this test pins the baseline's imperfect score -- the floor a
// smarter proposer must beat. The pins mirror the offline floor test
// (TestBaselineFloorOnMessyAccount) by class and count: everything here
// derives from the suite's fixed members, never from the shared account's
// unrelated contents (out-of-universe noise stays informational).
func TestMappingEval_MessyAccount(t *testing.T) {
	report := runMappingEvalChain(t, "messy-account")

	if report.Perfect() {
		t.Fatal("the messy-account suite exists BECAUSE the baseline cannot ace it; a perfect score means the exam lost its headroom")
	}
	// Grouping: the 9 covered-kind resources land correctly; the 5
	// uncovered-tier resources (security group, KMS key + alias, DynamoDB
	// table, ECR repository) are unclaimed -- and nothing is misassigned
	// or double-claimed.
	if report.Grouping.UniverseSize != 14 || report.Grouping.Correct != 9 || report.Grouping.InUniverseClaims != 9 {
		t.Fatalf("grouping floor drifted: want 9/14 correct with 9 in-universe claims, got %d/%d with %d\n%s",
			report.Grouping.Correct, report.Grouping.UniverseSize, report.Grouping.InUniverseClaims, report.Summary())
	}
	if len(report.Grouping.Misassigned) != 0 || len(report.Grouping.DuplicateClaims) != 0 {
		t.Fatalf("the baseline must never misassign or double-claim: %v / %v",
			report.Grouping.Misassigned, report.Grouping.DuplicateClaims)
	}
	if len(report.Grouping.Unclaimed) != 5 {
		t.Fatalf("want the 5 uncovered-tier resources unclaimed, got %v", report.Grouping.Unclaimed)
	}
	// Coverage: the uncovered tier is DECLARED unmapped; zero unaccounted
	// is part of the floor (honesty even where blind).
	if len(report.Coverage.Unaccounted) != 0 {
		t.Fatalf("the baseline accounts for everything it sees; unaccounted: %v", report.Coverage.Unaccounted)
	}
	if len(report.Coverage.UnmappedInUniverse) != 5 {
		t.Fatalf("want the 5 uncovered-tier resources declared unmapped, got %v", report.Coverage.UnmappedInUniverse)
	}
	// Refs: covered wiring holds (4 edges); the queue's kms edge and the
	// two unproposed instances' edges are missing; nothing is wired WRONG.
	if report.Refs.GroundTruthEdges != 7 || report.Refs.CorrectEdges != 4 || len(report.Refs.MissingEdges) != 3 {
		t.Fatalf("refs floor drifted: want 4/7 edges with 3 missing, got %d/%d with %d\n%v",
			report.Refs.CorrectEdges, report.Refs.GroundTruthEdges, len(report.Refs.MissingEdges), report.Refs.MissingEdges)
	}
	if len(report.Refs.WrongTargetEdges) != 0 || len(report.Refs.UnexpectedEdges) != 0 {
		t.Fatalf("the baseline never wires a WRONG edge: %v / %v",
			report.Refs.WrongTargetEdges, report.Refs.UnexpectedEdges)
	}
	// Spec: what the baseline reconstructs, it reconstructs correctly; the
	// missing leaves belong to the four instances it never proposed. The
	// exact pin matches the offline floor -- the leaves derive from the
	// suite's fixed manifests, so live and offline must agree.
	if len(report.Spec.Mismatched) != 0 {
		t.Fatalf("the baseline never writes a wrong value: %v", report.Spec.Mismatched)
	}
	if report.Spec.GroundTruthLeaves != 64 || report.Spec.Matched != 21 {
		t.Fatalf("spec floor drifted: want 21/64 leaves matched, got %d/%d\nmissing: %v\nmismatched: %v",
			report.Spec.Matched, report.Spec.GroundTruthLeaves, report.Spec.Missing, report.Spec.Mismatched)
	}
	if len(report.NameDerivability) != 0 {
		t.Fatalf("the bucket keeps its name-derived identity: %v", report.NameDerivability)
	}
	// Partition: the ONE graded suite. The untaught baseline recovers
	// prod/staging from the member names' env tokens; the token-less
	// bucket and the four unproposed instances are owed; nothing is
	// assigned WRONG. Mirrors the offline pin -- the denominator derives
	// from the suite's fixed members, so live and offline must agree.
	if !report.Partition.Graded {
		t.Fatal("the messy suite declares partition grading; the axis must be graded")
	}
	if report.Partition.GroundTruthInstances != 11 || report.Partition.Correct != 6 || len(report.Partition.MissingEnv) != 5 {
		t.Fatalf("partition floor drifted: want 6/11 assigned with 5 missing, got %d/%d with %d\nmissing: %v\nwrong: %v",
			report.Partition.Correct, report.Partition.GroundTruthInstances, len(report.Partition.MissingEnv),
			report.Partition.MissingEnv, report.Partition.WrongEnv)
	}
	if len(report.Partition.WrongEnv) != 0 || len(report.Partition.ExtraEnv) != 0 {
		t.Fatalf("the baseline must never assign a wrong or extra environment: %v / %v",
			report.Partition.WrongEnv, report.Partition.ExtraEnv)
	}
}

// TestMappingEval_IdentityAndEgress is the live proof of the COVERAGE exam:
// the identity-and-egress suite brings the last two import-recipe kinds
// (IAM role, NAT gateway) under examination, so every recipe kind appears
// in at least one exam. Like the messy lane, this pins the baseline's
// imperfect floor by class and count -- universe-derived numbers only. The
// IAM scan is the first GLOBAL-service scan a live lane performs: it lists
// every role in the account, and all of that noise must stay informational.
func TestMappingEval_IdentityAndEgress(t *testing.T) {
	report := runMappingEvalChain(t, "identity-and-egress")

	if report.Perfect() {
		t.Fatal("the identity-and-egress suite carries kinds the baseline deliberately does not map; a perfect score means the exam lost its headroom")
	}
	// Grouping: the network tier (VPC + both subnets incl. the routed
	// subnet's table and association) lands correctly; the NAT gateway and
	// the role are unclaimed. Nothing is misassigned or double-claimed.
	if report.Grouping.UniverseSize != 7 || report.Grouping.Correct != 5 || report.Grouping.InUniverseClaims != 5 {
		t.Fatalf("grouping floor drifted: want 5/7 correct with 5 in-universe claims, got %d/%d with %d\n%s",
			report.Grouping.Correct, report.Grouping.UniverseSize, report.Grouping.InUniverseClaims, report.Summary())
	}
	if len(report.Grouping.Misassigned) != 0 || len(report.Grouping.DuplicateClaims) != 0 {
		t.Fatalf("the baseline must never misassign or double-claim: %v / %v",
			report.Grouping.Misassigned, report.Grouping.DuplicateClaims)
	}
	if len(report.Grouping.Unclaimed) != 2 {
		t.Fatalf("want the NAT gateway + role unclaimed, got %v", report.Grouping.Unclaimed)
	}
	// Coverage: both uncovered resources are DECLARED unmapped; zero
	// unaccounted is part of the floor (honesty even where blind).
	if len(report.Coverage.Unaccounted) != 0 {
		t.Fatalf("the baseline accounts for everything it sees; unaccounted: %v", report.Coverage.Unaccounted)
	}
	if len(report.Coverage.UnmappedInUniverse) != 2 {
		t.Fatalf("want the NAT gateway + role declared unmapped, got %v", report.Coverage.UnmappedInUniverse)
	}
	// Refs: the two vpcId edges hold; the two edges touching the NAT
	// gateway (its subnet_id, the route's nat_gateway target) are owed;
	// nothing is wired WRONG.
	if report.Refs.GroundTruthEdges != 4 || report.Refs.CorrectEdges != 2 || len(report.Refs.MissingEdges) != 2 {
		t.Fatalf("refs floor drifted: want 2/4 edges with 2 missing, got %d/%d with %d\n%v",
			report.Refs.CorrectEdges, report.Refs.GroundTruthEdges, len(report.Refs.MissingEdges), report.Refs.MissingEdges)
	}
	if len(report.Refs.WrongTargetEdges) != 0 || len(report.Refs.UnexpectedEdges) != 0 {
		t.Fatalf("the baseline never wires a WRONG edge: %v / %v",
			report.Refs.WrongTargetEdges, report.Refs.UnexpectedEdges)
	}
	// Spec: what the baseline reconstructs, it reconstructs correctly; the
	// missing leaves belong to the two instances it never proposed. The
	// leaf counts derive from the suite's fixed manifests, so live and
	// offline must agree.
	if len(report.Spec.Mismatched) != 0 {
		t.Fatalf("the baseline never writes a wrong value: %v", report.Spec.Mismatched)
	}
	if report.Spec.GroundTruthLeaves != 21 || report.Spec.Matched != 12 {
		t.Fatalf("spec floor drifted: want 12/21 leaves matched, got %d/%d\nmissing: %v\nmismatched: %v",
			report.Spec.Matched, report.Spec.GroundTruthLeaves, report.Spec.Missing, report.Spec.Mismatched)
	}
	if len(report.NameDerivability) != 0 {
		t.Fatalf("no matched instance breaks name-derived identity: %v", report.NameDerivability)
	}
}

// runMappingEvalChain runs the full live chain for one suite: deploy the
// answer key (teardown ALWAYS runs and its failure fails the test -- leaked
// fixtures are a real cost), scan the region blind over the derived type
// allowlist, redact the deploy machinery's own fingerprints, propose via
// the baseline, validate against the contract, and score.
func runMappingEvalChain(t *testing.T, suiteName string) *mappingeval.Report {
	if os.Getenv(MappingEvalEnvVar) != "1" {
		t.Skipf("mapping eval lane is opt-in; set %s=1", MappingEvalEnvVar)
	}
	ctx := context.Background()

	suitePath := filepath.Join(repoRoot, "catalog", "aws", "aa_eval", "suites", suiteName+".yaml")
	suite, err := mappingeval.LoadSuite(repoRoot, suitePath)
	if err != nil {
		t.Fatalf("loading suite: %v", err)
	}
	region := suite.Suite.GetSpec().GetScanScope().GetRegion()

	deployed, groundTruth, deployErr := evalrunner.DeploySuite(t, repoRoot, "aws", suite)
	defer func() {
		if err := evalrunner.TeardownSuite(t, deployed); err != nil {
			t.Errorf("suite teardown failed -- fixtures may be leaking: %v", err)
		}
	}()
	if deployErr != nil {
		t.Fatalf("deploying suite: %v", deployErr)
	}

	scanTypes, err := evalrunner.ScanTypeNames(repoRoot, "aws", suite)
	if err != nil {
		t.Fatalf("deriving scan type allowlist: %v", err)
	}
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		t.Fatalf("loading AWS config: %v", err)
	}
	scan, err := inventory.NewScanner(cfg).Scan(ctx, region, scanTypes)
	if err != nil {
		t.Fatalf("scanning account: %v", err)
	}
	// The exam presents the seeded account as a stranger's: the deploy
	// machinery's own identity tags would otherwise hand any tag-reading
	// proposer the answer key.
	mappingeval.RedactSeedFingerprints(scan)
	writeArtifact(t, suiteName+"-scan.json", func() ([]byte, error) { return json.MarshalIndent(scan, "", "  ") })

	proposal, err := baseline.Propose(scan)
	if err != nil {
		t.Fatalf("baseline proposer: %v", err)
	}
	writeArtifact(t, suiteName+"-proposal.json", func() ([]byte, error) {
		return protojson.MarshalOptions{Multiline: true, Indent: "  "}.Marshal(proposal)
	})
	loaded, err := mappingeval.ParseProposal(proposal)
	if err != nil {
		t.Fatalf("baseline proposal violates the contract: %v", err)
	}

	kindDirs := make([]string, 0, len(suite.Members))
	for _, member := range suite.Members {
		kindDirs = append(kindDirs, member.KindDir)
	}
	scoreOptions, err := mappingeval.ScoreOptionsFromCatalog(repoRoot, "aws", kindDirs)
	if err != nil {
		t.Fatalf("deriving score options: %v", err)
	}
	// The partition axis is suite-declared (exam fairness: only suites
	// whose member envs are recoverable answer keys grade it).
	scoreOptions.GradeEnvironmentPartition = suite.Suite.GetSpec().GetGradeEnvironmentPartition()
	report := mappingeval.Score(groundTruth, loaded, scoreOptions)
	writeArtifact(t, suiteName+"-report.json", func() ([]byte, error) { return json.MarshalIndent(report, "", "  ") })
	t.Logf("mapping eval report (%s):\n%s", suiteName, report.Summary())
	return report
}

// writeArtifact persists one run artifact when an artifacts dir is set.
func writeArtifact(t *testing.T, name string, render func() ([]byte, error)) {
	dir := os.Getenv("PLANTON_E2E_MAPPING_EVAL_ARTIFACTS")
	if dir == "" {
		return
	}
	payload, err := render()
	if err != nil {
		t.Logf("rendering artifact %s failed: %v", name, err)
		return
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Logf("creating artifacts dir failed: %v", err)
		return
	}
	if err := os.WriteFile(filepath.Join(dir, name), payload, 0o644); err != nil {
		t.Logf("writing artifact %s failed: %v", name, err)
	}
}

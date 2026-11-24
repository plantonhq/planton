package verify

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/catalog/kubernetes/aa_e2e/fakemodel"
	"github.com/plantonhq/planton/e2e/framework/provider"
	stigmer "go.stigmer.ai/sdk"
	agentv1 "go.stigmer.ai/sdk/proto/ai/stigmer/agentic/agent/v1"
	runv1 "go.stigmer.ai/sdk/proto/ai/stigmer/agentic/run/v1"
	sessionv1 "go.stigmer.ai/sdk/proto/ai/stigmer/agentic/session/v1"
	"go.stigmer.ai/sdk/proto/ai/stigmer/commons/apiresource/apiresourcekind"
	organizationv1 "go.stigmer.ai/sdk/proto/ai/stigmer/tenancy/organization/v1"
)

// Stigmer's fixed names, as the kind's module writes them (iac/tf/locals.tf,
// iac/pulumi/module/vars.go) and Stigmer's operator derives them
// (deploy/operator/internal/render/names.go, internal/manifests in
// stigmer/stigmer).
const (
	stigmerName            = "stigmer"
	stigmerResource        = "stigmers.stigmer.ai"
	stigmerOperatorRelease = "stigmer-operator"
	stigmerKeysSecret      = "stigmer-keys"
	stigmerServerPort      = 7234
	stigmerPostgresCluster = "stigmer-postgres"

	// What the operator installs on its own when a prerequisite is missing,
	// labelled app.kubernetes.io/managed-by=stigmer-operator.
	stigmerManagedByLabel       = "app.kubernetes.io/managed-by"
	stigmerManagedByValue       = "stigmer-operator"
	stigmerCloudNativePgNs      = "cnpg-system"
	stigmerCloudNativePgCrd     = "clusters.postgresql.cnpg.io"
	stigmerAgentSandboxNs       = "agent-sandbox-system"
	stigmerAgentSandboxCrd      = "sandboxes.agents.x-k8s.io"
	stigmerBarmanObjectStoreCrd = "objectstores.barmancloud.cnpg.io"
	stigmerSubstrateNs          = "ate-system"
	stigmerSubstrateCrd         = "workerpools.ate.dev"

	// The agent every lane talks to: one with no tools, which answers in one
	// turn. A fixed slug, so a later act (a restore, an upgrade) finds the
	// runs the first act made.
	stigmerProbeAgent = "e2e-probe"

	stigmerReadyBudget  = 20 * time.Minute
	stigmerRunBudget    = 10 * time.Minute
	stigmerBackupBudget = 10 * time.Minute
)

// The refusal classes a Stigmer scenario may expect, each pinned to the words
// both engines write (iac/tf/preflight.tf and scripts.tf, and their Pulumi
// twins in iac/pulumi/module/preflight.go and scripts.go).
var stigmerRefusals = map[string][]string{
	"stigmer-downgrade":                {"records Stigmer v", "does not move from v"},
	"stigmer-keys-missing":             {"database.bundled.recover_from restores archive", "is not in namespace"},
	"stigmer-operator-owned-elsewhere": {"the Stigmer definition " + stigmerResource + " on the cluster belongs to"},
	"stigmer-prerequisite-orphaned":    {"is on the cluster, installed by the Helm release", "and no controller runs it"},
}

// StigmerVerifier checks a Stigmer install by what it promises a platform
// engineer: the install reaches Ready at the declared version, the operator
// runs, and an agent answers a message, driven through Stigmer's own Go SDK
// against the fake model the lane serves (catalog/kubernetes/aa_e2e/fakemodel).
// A lane that declares backups also proves an on-demand backup completes; a
// restore act proves the restored install holds the run the first act made.
// Destroy asserts the posture the README teaches: the Stigmer, the operator
// and what the operator itself installed are gone, while the namespace's
// encryption keys and the Stigmer definition stay.
type StigmerVerifier struct {
	ManifestPath      string
	Version           string
	Namespace         string
	OperatorNamespace string
	BundledDatabase   bool
	Backups           bool
	RecoverFrom       string
	SandboxDriver     string
	// Exposed marks an install behind a hostname with Stigmer's bundled
	// sign-in; its proof is that the front door answers and refuses an
	// anonymous call, not an anonymous agent turn.
	Exposed  bool
	Hostname string
	// Gateway names the Gateway an exposed install routes through, with its
	// namespace; empty for an Ingress.
	Gateway, GatewayNamespace string
	// Refused marks a lane whose deploy must be refused before anything is
	// installed, so its cleanup proves only that nothing was.
	Refused bool
	// CatalogCloudNativePg marks a lane whose CloudNativePG is the catalog's
	// own (a KubernetesCloudNativePgOperator fixture), which Stigmer's `auto`
	// mode must leave in place.
	CatalogCloudNativePg bool
}

func newStigmerVerifier(manifestPath string) *StigmerVerifier {
	spec := manifestSpecMap(manifestPath)
	v := &StigmerVerifier{
		ManifestPath:      manifestPath,
		Version:           stringOr(specField(spec, "version"), ""),
		Namespace:         stringOr(specField(specFieldMap(spec, "namespace"), "value"), stigmerName),
		OperatorNamespace: stringOr(specField(specFieldMap(specFieldMap(spec, "operator"), "namespace"), "value"), "stigmer-system"),
	}
	database := specFieldMap(spec, "database")
	if _, external := specField(database, "external").(map[string]interface{}); !external {
		v.BundledDatabase = true
		bundled := specFieldMap(database, "bundled")
		_, v.Backups = specField(bundled, "backup").(map[string]interface{})
		v.RecoverFrom = stringOr(specField(specFieldMap(bundled, "recoverFrom"), "serverName"), "")
	}
	v.SandboxDriver = stringOr(specField(specFieldMap(spec, "sandboxes"), "driver"), "")
	if ingress := specFieldMap(spec, "ingress"); len(ingress) > 0 {
		v.Exposed = true
		v.Hostname = stringOr(specField(ingress, "hostname"), "")
		if gateway := specFieldMap(ingress, "gateway"); len(gateway) > 0 {
			v.Gateway = stringOr(specField(specFieldMap(gateway, "name"), "value"), "")
			v.GatewayNamespace = stringOr(specField(specFieldMap(gateway, "namespace"), "value"), v.Namespace)
		}
	}
	v.CatalogCloudNativePg = manifestHasPrerequisite(manifestPath, "KubernetesCloudNativePgOperator")
	v.Refused = manifestAnnotation(manifestPath, "planton.dev/e2e-expect-deploy-failure") != ""
	return v
}

// VerifyExists has no outputs to read; every Stigmer lane is verified through
// VerifyExistsWithOutputs.
func (v *StigmerVerifier) VerifyExists(ctx context.Context, kubeconfig string) error {
	return v.VerifyExistsWithOutputs(ctx, kubeconfig, nil)
}

// VerifyExistsWithOutputs proves the install from the kind's outputs: the
// namespace and the server's address a consumer composes against.
func (v *StigmerVerifier) VerifyExistsWithOutputs(ctx context.Context, kubeconfig string, outputs map[string]interface{}) error {
	namespace := v.Namespace
	if out, ok := outputs["namespace"].(string); ok && out != "" {
		namespace = out
	}
	status, err := v.waitReady(ctx, kubeconfig, namespace)
	if err != nil {
		return err
	}
	fmt.Printf("  [verify] Stigmer %s/%s is Ready at %s\n", namespace, stigmerName, status.Version)

	if err := kubectlRun(ctx, kubeconfig, "-n", v.OperatorNamespace, "wait", "deployment/"+stigmerOperatorRelease,
		"--for=condition=Available", "--timeout=120s"); err != nil {
		return errors.Wrapf(err, "the operator Deployment %s/%s is not Available", v.OperatorNamespace, stigmerOperatorRelease)
	}

	if v.RecoverFrom != "" {
		if err := v.verifyRestored(ctx, kubeconfig, namespace, status); err != nil {
			return err
		}
	}
	upgradedFrom := v.upgradedFrom(ctx)
	if upgradedFrom != "" {
		if err := v.verifyUpgradeRecorded(ctx, kubeconfig, namespace, upgradedFrom); err != nil {
			return err
		}
	}
	if v.Exposed {
		// An exposed install signs people in, so its proof is the front door
		// and the refusal of a caller who has not signed in, never an
		// anonymous agent turn.
		if err := v.verifyFrontDoor(ctx, kubeconfig); err != nil {
			return err
		}
		if err := v.withClient(ctx, kubeconfig, namespace, func(client *stigmer.Client) error { return verifySignInEnforced(ctx, client) }); err != nil {
			return err
		}
	} else if err := v.withClient(ctx, kubeconfig, namespace, func(client *stigmer.Client) error {
		if v.RecoverFrom != "" || upgradedFrom != "" {
			if err := v.verifyEarlierRunSurvived(ctx, client); err != nil {
				return err
			}
		}
		if err := v.verifyAgentTurn(ctx, client); err != nil {
			return err
		}
		if v.SandboxDriver == "agentSandbox" {
			return v.verifyRanInSandbox(ctx, kubeconfig, namespace)
		}
		return nil
	}); err != nil {
		return err
	}
	// The backup comes after the turn, so the archive holds the run: a
	// restore brings back the newest moment the archive holds, and changes
	// reach the archive only when the database moves to its next WAL file.
	if v.Backups {
		return v.verifyBackup(ctx, kubeconfig, namespace)
	}
	return nil
}

// verifyRanInSandbox proves a per-session install ran the turn in a sandbox
// of its own: a Sandbox in the namespace the operator keeps them in.
func (v *StigmerVerifier) verifyRanInSandbox(ctx context.Context, kubeconfig, namespace string) error {
	sandboxes := namespace + "-sandboxes"
	names, err := kubectlGetJSONPath(ctx, kubeconfig, "sandboxes.agents.x-k8s.io", "", sandboxes, "{.items[*].metadata.name}")
	if err != nil || strings.TrimSpace(names) == "" {
		return errors.Errorf("the run completed, but no Sandbox is in %s: a per-session install must run each session in its own (%v)", sandboxes, err)
	}
	fmt.Printf("  [verify] the session ran in Sandbox %s/%s\n", sandboxes, strings.Fields(names)[0])
	return nil
}

// VerifyAbsent asserts the destroy posture.
func (v *StigmerVerifier) VerifyAbsent(ctx context.Context, kubeconfig string) error {
	if err := KubectlResourceAbsent(ctx, kubeconfig, "stigmer", stigmerName, v.Namespace); err != nil {
		return errors.Wrap(err, "the Stigmer outlived the destroy")
	}
	if err := KubectlResourceAbsent(ctx, kubeconfig, "deployment", stigmerOperatorRelease, v.OperatorNamespace); err != nil {
		return errors.Wrap(err, "the operator outlived the destroy")
	}
	if v.Refused {
		fmt.Printf("  [verify] the refused deploy installed nothing\n")
		return nil
	}
	if v.BundledDatabase {
		if v.CatalogCloudNativePg {
			if err := v.requireNotStigmers(ctx, kubeconfig, stigmerCloudNativePgCrd); err != nil {
				return errors.Wrap(err, "the catalog's CloudNativePG must stay after Stigmer's destroy")
			}
		} else if err := v.requireStigmersGone(ctx, kubeconfig, stigmerCloudNativePgCrd, stigmerCloudNativePgNs); err != nil {
			return err
		}
	}
	if v.SandboxDriver == "agentSandbox" {
		if err := v.requireStigmersGone(ctx, kubeconfig, stigmerAgentSandboxCrd, stigmerAgentSandboxNs); err != nil {
			return err
		}
		if err := KubectlResourceAbsent(ctx, kubeconfig, "namespace", v.Namespace+"-sandboxes", ""); err != nil {
			return errors.Wrap(err, "the operator's cleanup left the sandbox namespace")
		}
	}
	if v.SandboxDriver == "substrate" {
		if err := v.requireStigmersGone(ctx, kubeconfig, stigmerSubstrateCrd, stigmerSubstrateNs); err != nil {
			return err
		}
	}
	if err := KubectlResourceExists(ctx, kubeconfig, "secret", stigmerKeysSecret, v.Namespace); err != nil {
		return errors.Wrapf(err, "the encryption keys (%s/%s) must outlive the destroy, or every backup becomes unreadable", v.Namespace, stigmerKeysSecret)
	}
	if err := KubectlResourceExists(ctx, kubeconfig, "crd", stigmerResource, ""); err != nil {
		return errors.Wrapf(err, "the Stigmer definition %s is kept on uninstall by Stigmer's chart", stigmerResource)
	}
	fmt.Printf("  [verify] destroy left the keys %s/%s and the definition %s, and removed the rest\n", v.Namespace, stigmerKeysSecret, stigmerResource)
	return nil
}

// VerifyExpectedDeployFailure pins a refusal to its class's words and the
// three parts every refusal carries.
func (v *StigmerVerifier) VerifyExpectedDeployFailure(ctx context.Context, kubeconfig, expectation string, deployErr error) error {
	return verifyStigmerRefusal(expectation, deployErr)
}

func verifyStigmerRefusal(expectation string, deployErr error) error {
	phrases, ok := stigmerRefusals[expectation]
	if !ok {
		return errors.Errorf("unknown KubernetesStigmer refusal class %q", expectation)
	}
	if deployErr == nil {
		return errors.Errorf("the deploy was expected to be refused (%s) and succeeded", expectation)
	}
	// OpenTofu wraps a precondition's message at 80 columns, so words are
	// compared with every run of whitespace as one space.
	text := strings.Join(strings.Fields(deployErr.Error()), " ")
	for _, want := range append([]string{"observed:", "meaning:", "next step:"}, phrases...) {
		if !strings.Contains(text, want) {
			return errors.Errorf("the deploy failed, but not with the %s refusal: %q is missing from %s", expectation, want, lastLines(deployErr.Error(), 12))
		}
	}
	fmt.Printf("  [verify] refused as %s, in three parts\n", expectation)
	return nil
}

// ---- Ready -------------------------------------------------------------------

type stigmerStatus struct {
	Phase      string `json:"phase"`
	Version    string `json:"version"`
	Conditions []struct {
		Type    string `json:"type"`
		Status  string `json:"status"`
		Reason  string `json:"reason"`
		Message string `json:"message"`
	} `json:"conditions"`
	Backup *struct {
		ServerName   string `json:"serverName"`
		RestoredFrom string `json:"restoredFrom"`
	} `json:"backup"`
}

func (s stigmerStatus) condition(kind string) (string, string) {
	for _, c := range s.Conditions {
		if c.Type == kind {
			return c.Status, c.Reason + ": " + c.Message
		}
	}
	return "", "not reported yet"
}

func readStigmerStatus(ctx context.Context, kubeconfig, namespace string) (stigmerStatus, error) {
	out, err := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig, "-n", namespace,
		"get", "stigmer", stigmerName, "-o", "jsonpath={.status}").CombinedOutput()
	if err != nil {
		return stigmerStatus{}, errors.Errorf("reading Stigmer %s/%s: %s", namespace, stigmerName, strings.TrimSpace(string(out)))
	}
	var status stigmerStatus
	if strings.TrimSpace(string(out)) != "" {
		if err := json.Unmarshal(out, &status); err != nil {
			return stigmerStatus{}, errors.Wrap(err, "parsing the Stigmer's status")
		}
	}
	return status, nil
}

// waitReady waits for phase Ready at the manifest's version; a Refused or
// Error phase ends the wait at once with the operator's own sentence.
func (v *StigmerVerifier) waitReady(ctx context.Context, kubeconfig, namespace string) (stigmerStatus, error) {
	deadline := time.Now().Add(stigmerReadyBudget)
	var last stigmerStatus
	var lastErr error
	for time.Now().Before(deadline) {
		last, lastErr = readStigmerStatus(ctx, kubeconfig, namespace)
		if lastErr == nil {
			switch last.Phase {
			case "Ready":
				// The status spells the version with a leading "v"; the spec
				// does not.
				if v.Version == "" || strings.TrimPrefix(last.Version, "v") == v.Version {
					return last, nil
				}
			case "Refused", "Error":
				_, why := last.condition("Ready")
				return last, errors.Errorf("Stigmer %s/%s reads %s: %s", namespace, stigmerName, last.Phase, why)
			}
		}
		time.Sleep(10 * time.Second)
	}
	if lastErr != nil {
		return last, errors.Wrapf(lastErr, "Stigmer %s/%s did not become Ready within %s", namespace, stigmerName, stigmerReadyBudget)
	}
	_, why := last.condition("Ready")
	return last, errors.Errorf("Stigmer %s/%s did not become Ready at v%s within %s: phase %q, version %q, Ready %s",
		namespace, stigmerName, v.Version, stigmerReadyBudget, last.Phase, last.Version, why)
}

// ---- backups and restores ----------------------------------------------------

// verifyBackup waits until the database archives (BackupHealthy), then takes
// an on-demand backup through the barman-cloud plugin and waits for it, as a
// platform engineer does before a risky change.
func (v *StigmerVerifier) verifyBackup(ctx context.Context, kubeconfig, namespace string) error {
	deadline := time.Now().Add(stigmerBackupBudget)
	var why string
	for {
		status, err := readStigmerStatus(ctx, kubeconfig, namespace)
		if err == nil {
			var healthy string
			healthy, why = status.condition("BackupHealthy")
			if healthy == "True" {
				break
			}
		}
		if time.Now().After(deadline) {
			return errors.Errorf("the database never reported BackupHealthy within %s: %s", stigmerBackupBudget, why)
		}
		time.Sleep(10 * time.Second)
	}
	name := "e2e-" + strconv.FormatInt(time.Now().Unix(), 36)
	backup := fmt.Sprintf(`{"apiVersion":"postgresql.cnpg.io/v1","kind":"Backup","metadata":{"name":%q,"namespace":%q},"spec":{"cluster":{"name":%q},"method":"plugin","pluginConfiguration":{"name":"barman-cloud.cloudnative-pg.io"}}}`,
		name, namespace, stigmerPostgresCluster)
	apply := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig, "apply", "-f", "-")
	apply.Stdin = strings.NewReader(backup)
	if out, err := apply.CombinedOutput(); err != nil {
		return errors.Errorf("taking an on-demand backup: %s", strings.TrimSpace(string(out)))
	}
	deadline = time.Now().Add(stigmerBackupBudget)
	for time.Now().Before(deadline) {
		phase, _ := kubectlGetJSONPath(ctx, kubeconfig, "backups.postgresql.cnpg.io", name, namespace, "{.status.phase}")
		switch phase {
		case "completed":
			fmt.Printf("  [verify] on-demand backup %s completed\n", name)
			return nil
		case "failed":
			why, _ := kubectlGetJSONPath(ctx, kubeconfig, "backups.postgresql.cnpg.io", name, namespace, "{.status.error}")
			return errors.Errorf("on-demand backup %s failed: %s", name, why)
		}
		time.Sleep(10 * time.Second)
	}
	return errors.Errorf("on-demand backup %s did not complete within %s", name, stigmerBackupBudget)
}

// verifyRestored proves the database was restored from the declared archive
// and archives under a name of its own, so it never writes into the archive
// it came from.
func (v *StigmerVerifier) verifyRestored(ctx context.Context, kubeconfig, namespace string, status stigmerStatus) error {
	if status.Backup == nil || status.Backup.RestoredFrom != v.RecoverFrom {
		got := ""
		if status.Backup != nil {
			got = status.Backup.RestoredFrom
		}
		return errors.Errorf("status.backup.restoredFrom reads %q, want %q", got, v.RecoverFrom)
	}
	if status.Backup.ServerName == v.RecoverFrom {
		return errors.Errorf("the restored database archives into %q, the archive it was restored from", v.RecoverFrom)
	}
	fmt.Printf("  [verify] restored from %s; now archiving as %s\n", v.RecoverFrom, status.Backup.ServerName)
	return nil
}

// upgradedFrom is the first act's version when this act moved the install to
// another one in place (an upgrade lane's second act), else "".
func (v *StigmerVerifier) upgradedFrom(ctx context.Context) string {
	firstAct, _ := ctx.Value(provider.FirstActManifestPathKey{}).(string)
	if firstAct == "" || v.RecoverFrom != "" {
		return ""
	}
	from := stringOr(specField(manifestSpecMap(firstAct), "version"), "")
	if from == v.Version {
		return ""
	}
	return from
}

// verifyUpgradeRecorded proves the operator upgraded as Stigmer promises: the
// server's Deployment records the version it came from and, with backups on,
// the backup the operator took before rolling it.
func (v *StigmerVerifier) verifyUpgradeRecorded(ctx context.Context, kubeconfig, namespace, from string) error {
	recorded, err := kubectlGetJSONPath(ctx, kubeconfig, "deployment", stigmerName, namespace, `{.metadata.annotations.stigmer\.ai/upgraded-from}`)
	if err != nil {
		return errors.Wrap(err, "reading what the server's Deployment records about the upgrade")
	}
	if strings.TrimPrefix(recorded, "v") != from {
		return errors.Errorf("the server's Deployment records stigmer.ai/upgraded-from %q, want %s", recorded, from)
	}
	if v.Backups {
		backup, err := kubectlGetJSONPath(ctx, kubeconfig, "deployment", stigmerName, namespace, `{.metadata.annotations.stigmer\.ai/upgrade-backup}`)
		if err != nil || backup == "" {
			return errors.Errorf("the server's Deployment records no stigmer.ai/upgrade-backup: the operator must back the database up before an upgrade (%v)", err)
		}
		phase, _ := kubectlGetJSONPath(ctx, kubeconfig, "backups.postgresql.cnpg.io", backup, namespace, "{.status.phase}")
		if phase != "completed" {
			return errors.Errorf("the backup %s taken before the upgrade reads %q, want completed", backup, phase)
		}
		fmt.Printf("  [verify] upgraded from %s after backup %s\n", from, backup)
		return nil
	}
	fmt.Printf("  [verify] upgraded from %s\n", from)
	return nil
}

// ---- the agent turn ------------------------------------------------------------

// withClient opens a port-forward to the server and hands fn a client on
// it. The install has no hostname, so every caller is its owner.
func (v *StigmerVerifier) withClient(ctx context.Context, kubeconfig, namespace string, fn func(*stigmer.Client) error) error {
	port, err := freeLocalPort()
	if err != nil {
		return err
	}
	stop, err := startPortForward(ctx, kubeconfig, "svc/"+stigmerName, namespace, fmt.Sprintf("%d:%d", port, stigmerServerPort))
	if err != nil {
		return errors.Wrap(err, "port-forwarding to the Stigmer server")
	}
	defer stop()
	client, err := stigmer.NewClient(stigmer.WithBaseURL(fmt.Sprintf("127.0.0.1:%d", port)), stigmer.WithInsecure())
	if err != nil {
		return errors.Wrap(err, "opening a Stigmer client")
	}
	defer client.Close()
	return fn(client)
}

// probeOrganization is the organization the probe works in: the one a
// single-organization install holds, else one the probe creates.
func probeOrganization(ctx context.Context, client *stigmer.Client) (string, error) {
	mine, err := client.Organization.ListMine(ctx, &organizationv1.ListMyOrganizationsInput{})
	if err != nil {
		return "", errors.Wrap(err, "listing the caller's organizations")
	}
	if entries := mine.GetEntries(); len(entries) > 0 {
		return entries[0].GetMetadata().GetId(), nil
	}
	org, err := client.Organization.Create(ctx, &stigmer.OrganizationInput{Name: "e2e"})
	if err != nil {
		return "", errors.Wrap(err, "creating the probe's organization")
	}
	return org.GetMetadata().GetId(), nil
}

// probeAgent returns the probe agent, creating it on first use.
func probeAgent(ctx context.Context, client *stigmer.Client, org string) (*agentv1.Agent, error) {
	ref := stigmer.ResourceRef{Org: org, Slug: stigmerProbeAgent, Kind: apiresourcekind.ApiResourceKind_agent}
	if agent, err := client.Agent.GetByReference(ctx, ref); err == nil {
		return agent, nil
	}
	agent, err := client.Agent.Create(ctx, &stigmer.AgentInput{
		Name:         stigmerProbeAgent,
		Org:          org,
		Description:  "Planton e2e probe",
		Instructions: "Answer the message in one short sentence.",
	})
	if err != nil {
		return nil, errors.Wrap(err, "creating the probe agent")
	}
	return agent, nil
}

// verifyAgentTurn sends the probe agent one message and requires the run to
// complete with the fake model's reply, and the fake to have been called.
func (v *StigmerVerifier) verifyAgentTurn(ctx context.Context, client *stigmer.Client) error {
	model := fakemodel.Running()
	if model == nil {
		return errors.New("no fake model is running in this process: the KubernetesStigmer entrypoints start one before the lanes")
	}
	calls := model.Messages()
	org, err := probeOrganization(ctx, client)
	if err != nil {
		return err
	}
	agent, err := probeAgent(ctx, client, org)
	if err != nil {
		return err
	}
	run, err := client.Run.Create(ctx, &stigmer.RunInput{
		Name: "e2e-turn-" + strconv.FormatInt(time.Now().UnixNano(), 36),
		Org:  org,
		SessionSpec: &stigmer.SessionSpecInput{
			AgentRef: stigmer.ResourceRef{Org: org, Slug: agent.GetMetadata().GetSlug(), Kind: apiresourcekind.ApiResourceKind_agent},
		},
		Message: "Say hello.",
	})
	if err != nil {
		return errors.Wrap(err, "starting a run")
	}
	done, err := waitRunEnded(ctx, client, run.GetMetadata().GetId())
	if err != nil {
		return err
	}
	if done.GetStatus().GetPhase() != runv1.RunPhase_RUN_COMPLETED {
		return errors.Errorf("run %s ended %s: %s", done.GetMetadata().GetId(), done.GetStatus().GetPhase(), done.GetStatus().GetError())
	}
	reply, err := runReply(ctx, client, done)
	if err != nil {
		return err
	}
	if !strings.Contains(reply, fakemodel.ReplyText) {
		return errors.Errorf("run %s completed, but its reply is not the fake model's: %q", done.GetMetadata().GetId(), reply)
	}
	if model.Messages() <= calls {
		return errors.Errorf("run %s completed with the fake's words, but the fake model counted no call", done.GetMetadata().GetId())
	}
	fmt.Printf("  [verify] agent turn: run %s completed with the model's reply (%d model calls)\n",
		done.GetMetadata().GetId(), model.Messages()-calls)
	return nil
}

// verifyEarlierRunSurvived proves a restored or upgraded install holds the
// completed run the first act made.
func (v *StigmerVerifier) verifyEarlierRunSurvived(ctx context.Context, client *stigmer.Client) error {
	org, err := probeOrganization(ctx, client)
	if err != nil {
		return err
	}
	agent, err := client.Agent.GetByReference(ctx, stigmer.ResourceRef{Org: org, Slug: stigmerProbeAgent, Kind: apiresourcekind.ApiResourceKind_agent})
	if err != nil {
		return errors.Wrap(err, "the probe agent the first act made is missing")
	}
	runs, err := client.Run.ListByAgent(ctx, &runv1.ListRunsByAgentRequest{AgentId: agent.GetMetadata().GetId()})
	if err != nil {
		return errors.Wrap(err, "listing the probe agent's runs")
	}
	for _, run := range runs.GetEntries() {
		if run.GetStatus().GetPhase() == runv1.RunPhase_RUN_COMPLETED {
			fmt.Printf("  [verify] the run %s the first act made is there, completed\n", run.GetMetadata().GetId())
			return nil
		}
	}
	return errors.Errorf("the probe agent has %d runs after the restore, none completed: the first act's run did not survive", len(runs.GetEntries()))
}

func waitRunEnded(ctx context.Context, client *stigmer.Client, id string) (*runv1.Run, error) {
	deadline := time.Now().Add(stigmerRunBudget)
	var phase runv1.RunPhase
	for time.Now().Before(deadline) {
		run, err := client.Run.Get(ctx, id)
		if err == nil {
			phase = run.GetStatus().GetPhase()
			switch phase {
			case runv1.RunPhase_RUN_COMPLETED, runv1.RunPhase_RUN_FAILED, runv1.RunPhase_RUN_CANCELLED, runv1.RunPhase_RUN_TERMINATED:
				return run, nil
			}
		}
		time.Sleep(3 * time.Second)
	}
	return nil, errors.Errorf("run %s did not end within %s (last phase %s)", id, stigmerRunBudget, phase)
}

// runReply is the text of the run's newest top-level agent message.
func runReply(ctx context.Context, client *stigmer.Client, run *runv1.Run) (string, error) {
	events, err := client.Session.ListEvents(ctx, &sessionv1.ListSessionEventsRequest{
		SessionId: run.GetSpec().GetSessionId(),
		RunId:     run.GetMetadata().GetId(),
		Types:     []string{"agent.message"},
		Order:     "desc",
		PageSize:  100,
	})
	if err != nil {
		return "", errors.Wrapf(err, "reading run %s's messages", run.GetMetadata().GetId())
	}
	for _, event := range events.GetEvents() {
		if event.GetThreadId() != "" || event.GetAgentMessage() == nil {
			continue
		}
		var text strings.Builder
		for _, block := range event.GetAgentMessage().GetContent() {
			if block.GetType() == "text" {
				text.WriteString(block.GetText())
			}
		}
		return text.String(), nil
	}
	return "", errors.Errorf("run %s completed with no agent message", run.GetMetadata().GetId())
}

// ---- the front door ----------------------------------------------------------------

// verifyFrontDoor proves the hostname serves Stigmer over HTTPS through the
// cluster's front door (the ingress controller, or the Gateway's own
// Service): the certificate it presents names the hostname, and the bundled
// sign-in answers its discovery document at the path the operator routes.
func (v *StigmerVerifier) verifyFrontDoor(ctx context.Context, kubeconfig string) error {
	service, namespace, err := v.frontDoorService(ctx, kubeconfig)
	if err != nil {
		return err
	}
	port, err := freeLocalPort()
	if err != nil {
		return err
	}
	stop, err := startPortForward(ctx, kubeconfig, "svc/"+service, namespace, fmt.Sprintf("%d:443", port))
	if err != nil {
		return errors.Wrapf(err, "port-forwarding to the front door %s/%s", namespace, service)
	}
	defer stop()

	client := &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{TLSClientConfig: &tls.Config{
			ServerName: v.Hostname,
			// The lane's issuer is self-signed, so the chain is not judged;
			// the presented certificate must still name the hostname.
			InsecureSkipVerify: true,
		}},
	}
	discovery := "/idp/realms/stigmer/.well-known/openid-configuration"
	deadline := time.Now().Add(stigmerReadyBudget)
	var lastErr error
	for time.Now().Before(deadline) {
		lastErr = func() error {
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://127.0.0.1:%d%s", port, discovery), nil)
			if err != nil {
				return err
			}
			req.Host = v.Hostname
			resp, err := client.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			if resp.TLS == nil || len(resp.TLS.PeerCertificates) == 0 {
				return errors.New("the front door presented no certificate")
			}
			if err := resp.TLS.PeerCertificates[0].VerifyHostname(v.Hostname); err != nil {
				return errors.Wrapf(err, "the front door's certificate does not name %s", v.Hostname)
			}
			if resp.StatusCode != http.StatusOK {
				return errors.Errorf("https://%s%s answered HTTP %d", v.Hostname, discovery, resp.StatusCode)
			}
			var doc struct {
				Issuer string `json:"issuer"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
				return errors.Wrap(err, "the sign-in's discovery document is not JSON")
			}
			if !strings.Contains(doc.Issuer, v.Hostname) {
				return errors.Errorf("the sign-in's issuer is %q, not at %s", doc.Issuer, v.Hostname)
			}
			return nil
		}()
		if lastErr == nil {
			fmt.Printf("  [verify] https://%s serves the bundled sign-in through %s/%s, with a certificate for the hostname\n", v.Hostname, namespace, service)
			return nil
		}
		time.Sleep(10 * time.Second)
	}
	return errors.Wrapf(lastErr, "https://%s never served the bundled sign-in through %s/%s", v.Hostname, namespace, service)
}

// frontDoorService is the Service traffic for the hostname enters through:
// the Gateway's own (Istio names it <gateway>-istio), or the ingress-nginx
// controller's.
func (v *StigmerVerifier) frontDoorService(ctx context.Context, kubeconfig string) (string, string, error) {
	if v.Gateway != "" {
		return v.Gateway + "-istio", v.GatewayNamespace, nil
	}
	out, err := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig, "get", "svc", "-A",
		"-l", "app.kubernetes.io/name=ingress-nginx,app.kubernetes.io/component=controller",
		"-o", `jsonpath={range .items[*]}{.metadata.namespace}/{.metadata.name}{"\n"}{end}`).CombinedOutput()
	if err != nil {
		return "", "", errors.Errorf("finding the ingress controller's Service: %s", strings.TrimSpace(string(out)))
	}
	for _, line := range strings.Fields(string(out)) {
		if namespace, name, ok := strings.Cut(line, "/"); ok && !strings.HasSuffix(name, "-admission") {
			return name, namespace, nil
		}
	}
	return "", "", errors.New("no ingress-nginx controller Service is on the cluster")
}

// verifySignInEnforced proves an exposed install refuses a caller who has
// not signed in.
func verifySignInEnforced(ctx context.Context, client *stigmer.Client) error {
	_, err := client.Organization.ListMine(ctx, &organizationv1.ListMyOrganizationsInput{})
	// The SDK wraps the server's gRPC status in its own error type, so the
	// code is read through the SDK's predicate, never off the status or text.
	if stigmer.IsUnauthenticated(err) {
		fmt.Printf("  [verify] a caller who has not signed in is refused\n")
		return nil
	}
	return errors.Errorf("an exposed install with sign-in answered a caller who has not signed in: %v", err)
}

// ---- what the operator installed -------------------------------------------------

// requireStigmersGone asserts the operator removed what it installed itself:
// the prerequisite's namespace, and its definition unless someone else owns
// it.
func (v *StigmerVerifier) requireStigmersGone(ctx context.Context, kubeconfig, crd, namespace string) error {
	if err := KubectlResourceAbsent(ctx, kubeconfig, "namespace", namespace, ""); err != nil {
		return errors.Wrapf(err, "the operator's cleanup left namespace %s", namespace)
	}
	owner, err := kubectlGetJSONPath(ctx, kubeconfig, "crd", crd, "", "{.metadata.labels."+strings.ReplaceAll(stigmerManagedByLabel, ".", `\.`)+"}")
	if err == nil && owner == stigmerManagedByValue {
		return errors.Errorf("the operator's cleanup left its definition %s", crd)
	}
	return nil
}

// requireNotStigmers asserts a prerequisite someone else installed is still
// there and still theirs.
func (v *StigmerVerifier) requireNotStigmers(ctx context.Context, kubeconfig, crd string) error {
	owner, err := kubectlGetJSONPath(ctx, kubeconfig, "crd", crd, "", "{.metadata.labels."+strings.ReplaceAll(stigmerManagedByLabel, ".", `\.`)+"}")
	if err != nil {
		return errors.Wrapf(err, "definition %s is gone", crd)
	}
	if owner == stigmerManagedByValue {
		return errors.Errorf("definition %s is labelled as Stigmer's own", crd)
	}
	return nil
}

// ---- helpers ---------------------------------------------------------------------

func kubectlRun(ctx context.Context, kubeconfig string, args ...string) error {
	out, err := exec.CommandContext(ctx, "kubectl", append([]string{"--kubeconfig", kubeconfig}, args...)...).CombinedOutput()
	if err != nil {
		return errors.New(strings.TrimSpace(string(out)))
	}
	return nil
}

func lastLines(text string, n int) string {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "\n")
}

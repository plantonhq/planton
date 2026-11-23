package module

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	kubernetesmetav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// keysGate declares the ConfigMap whose creation runs the restore check
// (keysCheckScript). Terraform twin: kubectl_manifest.keys_gate in
// iac/tf/main.tf. It is a BeforeCreate hook, as the Tekton operator's
// readiness gate is: a failing hook fails the step before the ConfigMap is
// recorded, so the next deploy checks again. Its data holds what the check
// is about (the namespace and the archive), and ReplaceOnChanges("*") with
// DeleteBeforeReplace re-creates it, and so re-checks, whenever either moves;
// that is exactly when a new database could be created from an archive. The
// Stigmer depends on it, so the check runs before the Stigmer is declared.
// The hook never runs on a destroy or a preview.
func keysGate(ctx *pulumi.Context, locals *Locals, provider pulumi.ProviderResource, dependsOn []pulumi.Resource) (pulumi.Resource, error) {
	recoverFrom := locals.Spec.GetDatabase().GetBundled().GetRecoverFrom().GetServerName()
	hook, err := ctx.RegisterResourceHook(vars.KeysGate, func(*pulumi.ResourceHookArgs) error {
		return runScript("the restore check refused the deploy", keysCheckScript, []string{
			"STIGMER_NAMESPACE=" + locals.Namespace,
			"STIGMER_NAME=" + vars.StigmerName,
			"STIGMER_KEYS_SECRET=" + vars.KeysSecret,
			"STIGMER_RECOVER_FROM=" + recoverFrom,
		}, nil)
	}, nil)
	if err != nil {
		return nil, errors.Wrap(err, "registering the restore check")
	}
	return kubernetescorev1.NewConfigMap(ctx, vars.KeysGate, &kubernetescorev1.ConfigMapArgs{
		Metadata: &kubernetesmetav1.ObjectMetaArgs{
			Name:      pulumi.String(vars.KeysGate),
			Namespace: pulumi.String(locals.Namespace),
			Labels:    pulumi.ToStringMap(locals.Labels),
		},
		Data: pulumi.ToStringMap(keysGateData(locals.Namespace, recoverFrom)),
	},
		pulumi.Provider(provider),
		pulumi.DependsOn(dependsOn),
		pulumi.ReplaceOnChanges([]string{"*"}),
		pulumi.DeleteBeforeReplace(true),
		pulumi.ResourceHooks(&pulumi.ResourceHookBinding{BeforeCreate: []*pulumi.ResourceHook{hook}}))
}

// keysGateData is the restore check's data, identical to the Terraform
// twin's keys_gate_data.
func keysGateData(namespace, recoverFrom string) map[string]string {
	return map[string]string{
		"purpose":     "Records that the restore check passed for this namespace and archive: a restore found the keys Secret it needs. Deleting it changes nothing.",
		"namespace":   namespace,
		"recoverFrom": recoverFrom,
	}
}

// cleanupMarker declares the ConfigMap whose deletion waits for the
// operator's own cleanup (cleanupWaitScript). It depends on the chart's
// release, every Secret the module writes and the operator's namespace, and
// the Stigmer depends on it, so a destroy deletes the Stigmer first, then
// runs this marker's BeforeDelete hook while the operator and the bucket's
// credentials are still there, then removes the rest. The hook's code lives
// in this program, never in the cluster, and delete hooks need
// `pulumi destroy --run-program`, which Planton's stack runner passes.
//
// PARITY-EXCEPTION: the Terraform twin carries the wait on a state-only
// terraform_data (iac/tf/main.tf, terraform_data.cleanup_wait), because a
// destroy-time step there can read only its own state, and state is the one
// place nobody with access to the cluster can change the script the runner
// executes. So this ConfigMap is the one object the two engines do not share;
// the cross-engine test leaves it out.
func cleanupMarker(ctx *pulumi.Context, locals *Locals, provider pulumi.ProviderResource, dependsOn []pulumi.Resource) (pulumi.Resource, error) {
	hook, err := ctx.RegisterResourceHook(vars.WaitMarker, func(*pulumi.ResourceHookArgs) error {
		return runScript("", cleanupWaitScript, []string{
			"STIGMER_NAMESPACE=" + locals.Namespace,
			"STIGMER_NAME=" + vars.StigmerName,
			"STIGMER_OPERATOR_NAMESPACE=" + locals.OperatorNamespace,
			"STIGMER_OPERATOR_RELEASE=" + vars.OperatorRelease,
			"STIGMER_WAIT_SECONDS=" + strconv.Itoa(int(vars.CleanupWait.Seconds())),
		}, func(line string) { _ = ctx.Log.Info(line, nil) })
	}, nil)
	if err != nil {
		return nil, errors.Wrap(err, "registering the cleanup wait")
	}
	return kubernetescorev1.NewConfigMap(ctx, vars.WaitMarker, &kubernetescorev1.ConfigMapArgs{
		Metadata: &kubernetesmetav1.ObjectMetaArgs{
			Name:      pulumi.String(vars.WaitMarker),
			Namespace: pulumi.String(locals.Namespace),
			Labels:    pulumi.ToStringMap(locals.Labels),
		},
		Data: pulumi.StringMap{
			"purpose": pulumi.String("Marks where a destroy waits for Stigmer's operator to remove what it installed. Deleting it changes nothing."),
		},
	},
		pulumi.Provider(provider),
		pulumi.DependsOn(dependsOn),
		pulumi.ResourceHooks(&pulumi.ResourceHookBinding{BeforeDelete: []*pulumi.ResourceHook{hook}}))
}

// runScript runs one of the module's shell scripts with the deploy's
// environment plus env. With report set (the cleanup wait, which always
// succeeds), each line goes to the engine's log as it is written, so a long
// wait is never a silent step. Without it (the restore check), a failure
// returns the script's own explanation, bounded so a chatty kubectl never
// floods the engine's diagnostics, after what refused.
func runScript(refusal, script string, env []string, report func(string)) error {
	cmd := exec.Command("/bin/sh", "-c", script)
	cmd.Env = append(os.Environ(), env...)
	if report != nil {
		reader, writer := io.Pipe()
		cmd.Stdout = writer
		cmd.Stderr = writer
		done := make(chan struct{})
		go func() {
			scanner := bufio.NewScanner(reader)
			for scanner.Scan() {
				report(scanner.Text())
			}
			close(done)
		}()
		err := cmd.Run()
		_ = writer.Close()
		<-done
		return err
	}
	out, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}
	text := strings.TrimSpace(string(out))
	if len(text) > 4096 {
		text = "..." + text[len(text)-4096:]
	}
	return fmt.Errorf("%s (%v):\n%s", refusal, err, text)
}

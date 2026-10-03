package verify

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/pkg/errors"
)

func agentCoreClient(cfg aws.Config, region string) *bedrockagentcorecontrol.Client {
	return bedrockagentcorecontrol.NewFromConfig(cfg, func(o *bedrockagentcorecontrol.Options) {
		if region != "" {
			o.Region = region
		}
	})
}

// isAgentCoreNotFound matches the bedrock-agentcore control plane's
// resource-not-found class.
func isAgentCoreNotFound(err error) bool {
	var nf *awstypes.ResourceNotFoundException
	return errors.As(err, &nf) || strings.Contains(err.Error(), "ResourceNotFoundException")
}

// agentCoreRuntimeVerifier verifies AwsBedrockAgentCoreRuntime
// components. A healthy deploy lands the runtime READY; endpoints and
// the resource policy are folded satellites the lanes' output maps
// cover.
type agentCoreRuntimeVerifier struct{}

func (*agentCoreRuntimeVerifier) IDOutputKey() string { return "agent_runtime_id" }

func (*agentCoreRuntimeVerifier) VerifyExists(ctx context.Context, cfg aws.Config, id, region string) error {
	out, err := agentCoreClient(cfg, region).GetAgentRuntime(ctx, &bedrockagentcorecontrol.GetAgentRuntimeInput{
		AgentRuntimeId: aws.String(id),
	})
	if err != nil {
		return errors.Wrapf(err, "GetAgentRuntime(%s)", id)
	}
	status := out.Status
	if status != awstypes.AgentRuntimeStatusReady {
		return errors.Errorf("agent runtime %s is %s, expected READY", id, status)
	}
	return nil
}

func (*agentCoreRuntimeVerifier) VerifyAbsent(ctx context.Context, cfg aws.Config, id, region string) error {
	out, err := agentCoreClient(cfg, region).GetAgentRuntime(ctx, &bedrockagentcorecontrol.GetAgentRuntimeInput{
		AgentRuntimeId: aws.String(id),
	})
	if err == nil {
		// The control plane deletes asynchronously -- DELETING counts as
		// gone (the runner polls until fully absent regardless).
		if out.Status == awstypes.AgentRuntimeStatusDeleting {
			return nil
		}
		return errors.Errorf("agent runtime %s still exists (status %s)", id, out.Status)
	}
	if isAgentCoreNotFound(err) {
		return nil
	}
	return errors.Wrapf(err, "GetAgentRuntime(%s) during absence check", id)
}

// agentCoreGatewayVerifier verifies AwsBedrockAgentCoreGateway
// components. AWS deletes a gateway's targets before the gateway itself
// at destroy, so gateway absence implies target absence.
type agentCoreGatewayVerifier struct{}

func (*agentCoreGatewayVerifier) IDOutputKey() string { return "gateway_id" }

func (*agentCoreGatewayVerifier) VerifyExists(ctx context.Context, cfg aws.Config, id, region string) error {
	out, err := agentCoreClient(cfg, region).GetGateway(ctx, &bedrockagentcorecontrol.GetGatewayInput{
		GatewayIdentifier: aws.String(id),
	})
	if err != nil {
		return errors.Wrapf(err, "GetGateway(%s)", id)
	}
	status := out.Status
	if status == awstypes.GatewayStatusFailed || status == awstypes.GatewayStatusDeleting {
		return errors.Errorf("gateway %s is in status %s", id, status)
	}
	if status != awstypes.GatewayStatusReady {
		return errors.Errorf("gateway %s is %s, expected READY", id, status)
	}
	if aws.ToString(out.GatewayUrl) == "" {
		return errors.Errorf("gateway %s is READY but reports no gateway URL", id)
	}
	return nil
}

func (*agentCoreGatewayVerifier) VerifyAbsent(ctx context.Context, cfg aws.Config, id, region string) error {
	out, err := agentCoreClient(cfg, region).GetGateway(ctx, &bedrockagentcorecontrol.GetGatewayInput{
		GatewayIdentifier: aws.String(id),
	})
	if err == nil {
		if out.Status == awstypes.GatewayStatusDeleting {
			return nil
		}
		return errors.Errorf("gateway %s still exists (status %s)", id, out.Status)
	}
	if isAgentCoreNotFound(err) {
		return nil
	}
	return errors.Wrapf(err, "GetGateway(%s) during absence check", id)
}

// agentCoreMemoryVerifier verifies AwsBedrockAgentCoreMemory components.
// Strategy attach/detach serializes through the parent memory, so a
// healthy deploy lands the memory ACTIVE with every strategy attached.
type agentCoreMemoryVerifier struct{}

func (*agentCoreMemoryVerifier) IDOutputKey() string { return "memory_id" }

func (*agentCoreMemoryVerifier) VerifyExists(ctx context.Context, cfg aws.Config, id, region string) error {
	out, err := agentCoreClient(cfg, region).GetMemory(ctx, &bedrockagentcorecontrol.GetMemoryInput{
		MemoryId: aws.String(id),
	})
	if err != nil {
		return errors.Wrapf(err, "GetMemory(%s)", id)
	}
	status := out.Memory.Status
	if status != awstypes.MemoryStatusActive {
		return errors.Errorf("memory %s is %s, expected ACTIVE", id, status)
	}
	return nil
}

func (*agentCoreMemoryVerifier) VerifyAbsent(ctx context.Context, cfg aws.Config, id, region string) error {
	out, err := agentCoreClient(cfg, region).GetMemory(ctx, &bedrockagentcorecontrol.GetMemoryInput{
		MemoryId: aws.String(id),
	})
	if err == nil {
		if out.Memory.Status == awstypes.MemoryStatusDeleting {
			return nil
		}
		return errors.Errorf("memory %s still exists (status %s)", id, out.Memory.Status)
	}
	if isAgentCoreNotFound(err) {
		return nil
	}
	return errors.Wrapf(err, "GetMemory(%s) during absence check", id)
}

// agentCoreIdentityVerifier verifies AwsBedrockAgentCoreIdentity
// components -- an identity-and-access bundle whose arms are all
// name/id-keyed output maps, so verification walks the full outputs
// (the OutputsVerifier path).
type agentCoreIdentityVerifier struct{}

func (*agentCoreIdentityVerifier) IDOutputKey() string { return "policy_engine_id" }

func (*agentCoreIdentityVerifier) VerifyExists(ctx context.Context, cfg aws.Config, id, region string) error {
	return errors.New("awsbedrockagentcoreidentity verify-exists requires full outputs (name-keyed arm maps); use OutputsVerifier path")
}

func (*agentCoreIdentityVerifier) VerifyAbsent(ctx context.Context, cfg aws.Config, id, region string) error {
	return errors.New("awsbedrockagentcoreidentity verify-absent requires full outputs (name-keyed arm maps); use OutputsVerifier path")
}

func (*agentCoreIdentityVerifier) VerifyExistsFromOutputs(ctx context.Context, cfg aws.Config, outputs map[string]interface{}, region string) error {
	client := agentCoreClient(cfg, region)
	verified := 0
	for name := range outputKeys(outputs, "workload_identity_arns") {
		if _, err := client.GetWorkloadIdentity(ctx, &bedrockagentcorecontrol.GetWorkloadIdentityInput{Name: aws.String(name)}); err != nil {
			return errors.Wrapf(err, "GetWorkloadIdentity(%s)", name)
		}
		verified++
	}
	for name := range outputKeys(outputs, "api_key_provider_arns") {
		if _, err := client.GetApiKeyCredentialProvider(ctx, &bedrockagentcorecontrol.GetApiKeyCredentialProviderInput{Name: aws.String(name)}); err != nil {
			return errors.Wrapf(err, "GetApiKeyCredentialProvider(%s)", name)
		}
		verified++
	}
	for name := range outputKeys(outputs, "oauth2_provider_arns") {
		if _, err := client.GetOauth2CredentialProvider(ctx, &bedrockagentcorecontrol.GetOauth2CredentialProviderInput{Name: aws.String(name)}); err != nil {
			return errors.Wrapf(err, "GetOauth2CredentialProvider(%s)", name)
		}
		verified++
	}
	if engineId := stringOutputMap(outputs, "policy_engine_id"); engineId != "" {
		out, err := client.GetPolicyEngine(ctx, &bedrockagentcorecontrol.GetPolicyEngineInput{PolicyEngineId: aws.String(engineId)})
		if err != nil {
			return errors.Wrapf(err, "GetPolicyEngine(%s)", engineId)
		}
		if out.Status != awstypes.PolicyEngineStatusActive {
			return errors.Errorf("policy engine %s is %s, expected ACTIVE", engineId, out.Status)
		}
		verified++
	}
	if verified == 0 {
		return errors.New("awsbedrockagentcoreidentity outputs carry no arm to verify")
	}
	return nil
}

func (*agentCoreIdentityVerifier) VerifyAbsentFromOutputs(ctx context.Context, cfg aws.Config, outputs map[string]interface{}, region string) error {
	client := agentCoreClient(cfg, region)
	for name := range outputKeys(outputs, "workload_identity_arns") {
		if _, err := client.GetWorkloadIdentity(ctx, &bedrockagentcorecontrol.GetWorkloadIdentityInput{Name: aws.String(name)}); err == nil {
			return errors.Errorf("workload identity %s still exists", name)
		} else if !isAgentCoreNotFound(err) {
			return errors.Wrapf(err, "GetWorkloadIdentity(%s) during absence check", name)
		}
	}
	for name := range outputKeys(outputs, "api_key_provider_arns") {
		if _, err := client.GetApiKeyCredentialProvider(ctx, &bedrockagentcorecontrol.GetApiKeyCredentialProviderInput{Name: aws.String(name)}); err == nil {
			return errors.Errorf("api key credential provider %s still exists", name)
		} else if !isAgentCoreNotFound(err) {
			return errors.Wrapf(err, "GetApiKeyCredentialProvider(%s) during absence check", name)
		}
	}
	for name := range outputKeys(outputs, "oauth2_provider_arns") {
		if _, err := client.GetOauth2CredentialProvider(ctx, &bedrockagentcorecontrol.GetOauth2CredentialProviderInput{Name: aws.String(name)}); err == nil {
			return errors.Errorf("oauth2 credential provider %s still exists", name)
		} else if !isAgentCoreNotFound(err) {
			return errors.Wrapf(err, "GetOauth2CredentialProvider(%s) during absence check", name)
		}
	}
	if engineId := stringOutputMap(outputs, "policy_engine_id"); engineId != "" {
		if out, err := client.GetPolicyEngine(ctx, &bedrockagentcorecontrol.GetPolicyEngineInput{PolicyEngineId: aws.String(engineId)}); err == nil {
			if out.Status != awstypes.PolicyEngineStatusDeleting {
				return errors.Errorf("policy engine %s still exists (status %s)", engineId, out.Status)
			}
		} else if !isAgentCoreNotFound(err) {
			return errors.Wrapf(err, "GetPolicyEngine(%s) during absence check", engineId)
		}
	}
	return nil
}

// agentCoreToolsVerifier verifies AwsBedrockAgentCoreTools components --
// a tools bundle whose arms are all id-keyed output maps, so
// verification walks the full outputs (the OutputsVerifier path). The
// browser/profile/interpreter status vocabularies include DELETED as a
// terminal state, so absence accepts it alongside not-found.
type agentCoreToolsVerifier struct{}

func (*agentCoreToolsVerifier) IDOutputKey() string { return "browser_ids" }

func (*agentCoreToolsVerifier) VerifyExists(ctx context.Context, cfg aws.Config, id, region string) error {
	return errors.New("awsbedrockagentcoretools verify-exists requires full outputs (id-keyed arm maps); use OutputsVerifier path")
}

func (*agentCoreToolsVerifier) VerifyAbsent(ctx context.Context, cfg aws.Config, id, region string) error {
	return errors.New("awsbedrockagentcoretools verify-absent requires full outputs (id-keyed arm maps); use OutputsVerifier path")
}

func (*agentCoreToolsVerifier) VerifyExistsFromOutputs(ctx context.Context, cfg aws.Config, outputs map[string]interface{}, region string) error {
	client := agentCoreClient(cfg, region)
	verified := 0
	for _, browserId := range outputKeys(outputs, "browser_ids") {
		out, err := client.GetBrowser(ctx, &bedrockagentcorecontrol.GetBrowserInput{BrowserId: aws.String(browserId)})
		if err != nil {
			return errors.Wrapf(err, "GetBrowser(%s)", browserId)
		}
		if out.Status != awstypes.BrowserStatusReady {
			return errors.Errorf("browser %s is %s, expected READY", browserId, out.Status)
		}
		verified++
	}
	for _, profileId := range outputKeys(outputs, "browser_profile_ids") {
		out, err := client.GetBrowserProfile(ctx, &bedrockagentcorecontrol.GetBrowserProfileInput{ProfileId: aws.String(profileId)})
		if err != nil {
			return errors.Wrapf(err, "GetBrowserProfile(%s)", profileId)
		}
		if out.Status != awstypes.BrowserProfileStatusReady {
			return errors.Errorf("browser profile %s is %s, expected READY", profileId, out.Status)
		}
		verified++
	}
	for _, interpreterId := range outputKeys(outputs, "code_interpreter_ids") {
		out, err := client.GetCodeInterpreter(ctx, &bedrockagentcorecontrol.GetCodeInterpreterInput{CodeInterpreterId: aws.String(interpreterId)})
		if err != nil {
			return errors.Wrapf(err, "GetCodeInterpreter(%s)", interpreterId)
		}
		if out.Status != awstypes.CodeInterpreterStatusReady {
			return errors.Errorf("code interpreter %s is %s, expected READY", interpreterId, out.Status)
		}
		verified++
	}
	if verified == 0 {
		return errors.New("awsbedrockagentcoretools outputs carry no arm to verify")
	}
	return nil
}

func (*agentCoreToolsVerifier) VerifyAbsentFromOutputs(ctx context.Context, cfg aws.Config, outputs map[string]interface{}, region string) error {
	client := agentCoreClient(cfg, region)
	for _, browserId := range outputKeys(outputs, "browser_ids") {
		if out, err := client.GetBrowser(ctx, &bedrockagentcorecontrol.GetBrowserInput{BrowserId: aws.String(browserId)}); err == nil {
			if out.Status != awstypes.BrowserStatusDeleted && out.Status != awstypes.BrowserStatusDeleting {
				return errors.Errorf("browser %s still exists (status %s)", browserId, out.Status)
			}
		} else if !isAgentCoreNotFound(err) {
			return errors.Wrapf(err, "GetBrowser(%s) during absence check", browserId)
		}
	}
	for _, profileId := range outputKeys(outputs, "browser_profile_ids") {
		if out, err := client.GetBrowserProfile(ctx, &bedrockagentcorecontrol.GetBrowserProfileInput{ProfileId: aws.String(profileId)}); err == nil {
			if out.Status != awstypes.BrowserProfileStatusDeleted && out.Status != awstypes.BrowserProfileStatusDeleting {
				return errors.Errorf("browser profile %s still exists (status %s)", profileId, out.Status)
			}
		} else if !isAgentCoreNotFound(err) {
			return errors.Wrapf(err, "GetBrowserProfile(%s) during absence check", profileId)
		}
	}
	for _, interpreterId := range outputKeys(outputs, "code_interpreter_ids") {
		if out, err := client.GetCodeInterpreter(ctx, &bedrockagentcorecontrol.GetCodeInterpreterInput{CodeInterpreterId: aws.String(interpreterId)}); err == nil {
			if out.Status != awstypes.CodeInterpreterStatusDeleted && out.Status != awstypes.CodeInterpreterStatusDeleting {
				return errors.Errorf("code interpreter %s still exists (status %s)", interpreterId, out.Status)
			}
		} else if !isAgentCoreNotFound(err) {
			return errors.Wrapf(err, "GetCodeInterpreter(%s) during absence check", interpreterId)
		}
	}
	return nil
}

// agentCoreEvaluationVerifier verifies AwsBedrockAgentCoreEvaluation
// components -- a bundle whose arms are all id-keyed output maps, so
// verification walks the full outputs (the OutputsVerifier path).
// Evaluators and online configs land ACTIVE; harnesses land READY
// (the harness vocabulary differs from the other two). Absence is
// not-found (or DELETING on the harness).
type agentCoreEvaluationVerifier struct{}

func (*agentCoreEvaluationVerifier) IDOutputKey() string { return "evaluator_ids" }

func (*agentCoreEvaluationVerifier) VerifyExists(ctx context.Context, cfg aws.Config, id, region string) error {
	return errors.New("awsbedrockagentcoreevaluation verify-exists requires full outputs (id-keyed arm maps); use OutputsVerifier path")
}

func (*agentCoreEvaluationVerifier) VerifyAbsent(ctx context.Context, cfg aws.Config, id, region string) error {
	return errors.New("awsbedrockagentcoreevaluation verify-absent requires full outputs (id-keyed arm maps); use OutputsVerifier path")
}

func (*agentCoreEvaluationVerifier) VerifyExistsFromOutputs(ctx context.Context, cfg aws.Config, outputs map[string]interface{}, region string) error {
	client := agentCoreClient(cfg, region)
	verified := 0
	for _, evaluatorId := range outputKeys(outputs, "evaluator_ids") {
		out, err := client.GetEvaluator(ctx, &bedrockagentcorecontrol.GetEvaluatorInput{EvaluatorId: aws.String(evaluatorId)})
		if err != nil {
			return errors.Wrapf(err, "GetEvaluator(%s)", evaluatorId)
		}
		if out.Status != awstypes.EvaluatorStatusActive {
			return errors.Errorf("evaluator %s is %s, expected ACTIVE", evaluatorId, out.Status)
		}
		verified++
	}
	for _, harnessId := range outputKeys(outputs, "harness_ids") {
		out, err := client.GetHarness(ctx, &bedrockagentcorecontrol.GetHarnessInput{HarnessId: aws.String(harnessId)})
		if err != nil {
			return errors.Wrapf(err, "GetHarness(%s)", harnessId)
		}
		// GetHarnessOutput wraps the harness (unlike the evaluator and
		// online-config outputs, which carry Status at the top level).
		if out.Harness.Status != awstypes.HarnessStatusReady {
			return errors.Errorf("harness %s is %s, expected READY", harnessId, out.Harness.Status)
		}
		verified++
	}
	for _, configId := range outputKeys(outputs, "online_evaluation_config_ids") {
		out, err := client.GetOnlineEvaluationConfig(ctx, &bedrockagentcorecontrol.GetOnlineEvaluationConfigInput{
			OnlineEvaluationConfigId: aws.String(configId),
		})
		if err != nil {
			return errors.Wrapf(err, "GetOnlineEvaluationConfig(%s)", configId)
		}
		if out.Status != awstypes.OnlineEvaluationConfigStatusActive {
			return errors.Errorf("online evaluation config %s is %s, expected ACTIVE", configId, out.Status)
		}
		verified++
	}
	if verified == 0 {
		return errors.New("awsbedrockagentcoreevaluation outputs carry no arm to verify")
	}
	return nil
}

func (*agentCoreEvaluationVerifier) VerifyAbsentFromOutputs(ctx context.Context, cfg aws.Config, outputs map[string]interface{}, region string) error {
	client := agentCoreClient(cfg, region)
	for _, evaluatorId := range outputKeys(outputs, "evaluator_ids") {
		if _, err := client.GetEvaluator(ctx, &bedrockagentcorecontrol.GetEvaluatorInput{EvaluatorId: aws.String(evaluatorId)}); err == nil {
			return errors.Errorf("evaluator %s still exists", evaluatorId)
		} else if !isAgentCoreNotFound(err) {
			return errors.Wrapf(err, "GetEvaluator(%s) during absence check", evaluatorId)
		}
	}
	for _, harnessId := range outputKeys(outputs, "harness_ids") {
		if out, err := client.GetHarness(ctx, &bedrockagentcorecontrol.GetHarnessInput{HarnessId: aws.String(harnessId)}); err == nil {
			if out.Harness.Status != awstypes.HarnessStatusDeleting {
				return errors.Errorf("harness %s still exists (status %s)", harnessId, out.Harness.Status)
			}
		} else if !isAgentCoreNotFound(err) {
			return errors.Wrapf(err, "GetHarness(%s) during absence check", harnessId)
		}
	}
	for _, configId := range outputKeys(outputs, "online_evaluation_config_ids") {
		if _, err := client.GetOnlineEvaluationConfig(ctx, &bedrockagentcorecontrol.GetOnlineEvaluationConfigInput{
			OnlineEvaluationConfigId: aws.String(configId),
		}); err == nil {
			return errors.Errorf("online evaluation config %s still exists", configId)
		} else if !isAgentCoreNotFound(err) {
			return errors.Wrapf(err, "GetOnlineEvaluationConfig(%s) during absence check", configId)
		}
	}
	return nil
}

// outputKeys returns an output's map entries as key -> string value
// (empty when the output is absent or not a map).
func outputKeys(outputs map[string]interface{}, key string) map[string]string {
	result := map[string]string{}
	raw, ok := outputs[key]
	if !ok || raw == nil {
		return result
	}
	switch m := raw.(type) {
	case map[string]interface{}:
		for k, v := range m {
			if s, ok := v.(string); ok {
				result[k] = s
			}
		}
	case map[string]string:
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

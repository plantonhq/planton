package aa_e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/e2e/framework/provider"
	"github.com/plantonhq/planton/e2e/framework/runner"
	"google.golang.org/api/cloudkms/v1"
	"sigs.k8s.io/yaml"
)

// The classes of provider-created objects a GCP scenario may name with
// planton.dev/e2e-release-dependents (provider.DependentsReleaser).
const (
	// releaseAutokeyKey: a key handle makes Autokey create one Cloud HSM key
	// for it. Google keeps both the handle and the key (Cloud KMS deletes
	// neither), and an enabled HSM key version bills every month, so the
	// release schedules the lane's key's versions for destruction -- the
	// end state Cloud KMS allows, and the one the sweep accepts.
	releaseAutokeyKey = "autokey-key"

	// releaseChatEngineAgent: a CHAT engine that creates its agent
	// (agentCreationConfig) makes Dialogflow create a CX agent named after the
	// engine, and deleting the engine leaves that agent behind. The release
	// deletes the agent this lane's engine made, found by the engine's
	// run-scoped display name. Cloud Asset Inventory does not index
	// Dialogflow agents, so the sweep would never see one left over.
	releaseChatEngineAgent = "chat-engine-agent"

	// releaseDataStoreToolChatEngines: a Dialogflow agent's data store tool
	// makes Discovery Engine create chat engines named
	// <agent-id>-chat-<n>, sometimes more than one; the agent records only
	// the newest, which delete_chat_engine_on_destroy removes, and the rest
	// keep holding the store. The release deletes the chat engines over this
	// lane's own data stores (their ids carry the lane's run id).
	releaseDataStoreToolChatEngines = "data-store-tool-chat-engines"

	// releaseModelGardenModels: a Model Garden deployment makes Vertex AI
	// upload the model into the project's Model Registry; destroying the
	// deployment undeploys it and deletes the endpoint, but no provider call
	// deletes the uploaded model. Google's service agent does the upload, so
	// the lane's identity never appears on it in the audit log. Model Garden
	// scenarios take the gcp-vertex-gpu-serving class, so one runs at a time:
	// every undeployed model the project gained in the lane's window is the
	// lane's, and the release deletes those.
	releaseModelGardenModels = "model-garden-models"
)

// ReleaseDependents implements provider.DependentsReleaser.
func (h *Harness) ReleaseDependents(ctx context.Context, tc *provider.KindTestContext, class string) error {
	manifestPath, _ := ctx.Value(provider.ManifestPathKey{}).(string)
	if manifestPath == "" {
		return errors.New("no manifest on the context; the runner passes the first act's")
	}
	switch class {
	case releaseAutokeyKey:
		return h.releaseAutokeyKey(ctx, tc, manifestPath)
	case releaseChatEngineAgent:
		return h.releaseChatEngineAgent(ctx, tc, manifestPath)
	case releaseDataStoreToolChatEngines:
		return h.releaseDataStoreToolChatEngines(ctx, tc)
	case releaseModelGardenModels:
		return h.releaseModelGardenModels(ctx, tc, manifestPath)
	default:
		return errors.Errorf("unknown release class %q: the GCP harness releases %q, %q, %q and %q",
			class, releaseAutokeyKey, releaseChatEngineAgent, releaseDataStoreToolChatEngines, releaseModelGardenModels)
	}
}

// releaseAutokeyKey finds the key Autokey made for THIS lane's handle -- by
// the handle's run-scoped name in the expanded manifest, never by listing the
// project's shared autokey key ring, so a lane running beside it keeps its
// key -- and schedules every version still able to encrypt for destruction.
// A handle that never got created (a deploy that failed first) means Autokey
// made nothing.
func (h *Harness) releaseAutokeyKey(ctx context.Context, tc *provider.KindTestContext, manifestPath string) error {
	raw, err := os.ReadFile(manifestPath)
	if err != nil {
		return errors.Wrap(err, "reading the key handle manifest")
	}
	var manifest struct {
		Spec struct {
			Location      string `json:"location"`
			KeyHandleName string `json:"keyHandleName"`
		} `json:"spec"`
	}
	if err := yaml.Unmarshal(raw, &manifest); err != nil {
		return errors.Wrap(err, "parsing the key handle manifest")
	}
	if manifest.Spec.Location == "" || manifest.Spec.KeyHandleName == "" {
		return errors.New("the manifest names no key handle (spec.location and spec.keyHandleName)")
	}
	handleName := fmt.Sprintf("projects/%s/locations/%s/keyHandles/%s", h.services.Project, manifest.Spec.Location, manifest.Spec.KeyHandleName)

	var handle struct {
		KmsKey string `json:"kmsKey"`
	}
	status, err := h.restGet(ctx, "https://cloudkms.googleapis.com/v1/"+handleName, &handle)
	if status == http.StatusNotFound {
		lanef(tc, "  [release] key handle %s was never created; Autokey made no key", handleName)
		return nil
	}
	if err != nil {
		return errors.Wrapf(err, "reading the key handle %s", handleName)
	}
	if handle.KmsKey == "" {
		return errors.Errorf("key handle %s names no key", handleName)
	}

	versions := h.services.CloudKms.Projects.Locations.KeyRings.CryptoKeys.CryptoKeyVersions
	var destroyed []string
	err = versions.List(handle.KmsKey).Pages(ctx, func(page *cloudkms.ListCryptoKeyVersionsResponse) error {
		for _, version := range page.CryptoKeyVersions {
			if !keyVersionCanStillBeUsed(version.State) {
				continue
			}
			result, err := versions.Destroy(version.Name, &cloudkms.DestroyCryptoKeyVersionRequest{}).Context(ctx).Do()
			if err != nil {
				return errors.Wrapf(err, "scheduling %s for destruction", version.Name)
			}
			if keyVersionCanStillBeUsed(result.State) {
				return errors.Errorf("%s reads %s after its destroy was requested", version.Name, result.State)
			}
			destroyed = append(destroyed, version.Name)
		}
		return nil
	})
	if err != nil {
		return err
	}
	lanef(tc, "  [release] Autokey's key %s for %s: %d version(s) scheduled for destruction", handle.KmsKey, handleName, len(destroyed))
	return nil
}

// keyVersionCanStillBeUsed reports a Cloud KMS key version state that is not
// on its way out: anything but DESTROY_SCHEDULED, DESTROYED and the failed
// import or generation states, which hold no usable key material.
func keyVersionCanStillBeUsed(state string) bool {
	switch state {
	case "DESTROY_SCHEDULED", "DESTROYED", "IMPORT_FAILED", "GENERATION_FAILED":
		return false
	}
	return true
}

// restGet reads one Google REST resource with the harness's credential,
// for APIs the pinned typed clients do not cover. It returns the HTTP status
// so a caller can tell "absent" from a failure.
func (h *Harness) restGet(ctx context.Context, url string, out interface{}) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}
	resp, err := h.services.RestClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, errors.Errorf("GET %s: HTTP %d", url, resp.StatusCode)
	}
	return resp.StatusCode, json.NewDecoder(resp.Body).Decode(out)
}

// releaseChatEngineAgent deletes the Dialogflow CX agents in the engine's
// Dialogflow location whose display name is this lane's engine's -- the
// agent Google created for it. The display name carries the lane's run id,
// so another lane's agent never matches. None found means Google made none
// (the deploy failed first) or already removed it.
func (h *Harness) releaseChatEngineAgent(ctx context.Context, tc *provider.KindTestContext, manifestPath string) error {
	raw, err := os.ReadFile(manifestPath)
	if err != nil {
		return errors.Wrap(err, "reading the chat engine manifest")
	}
	var manifest struct {
		Spec struct {
			DisplayName      string `json:"displayName"`
			ChatEngineConfig struct {
				AgentCreationConfig struct {
					Location string `json:"location"`
				} `json:"agentCreationConfig"`
			} `json:"chatEngineConfig"`
		} `json:"spec"`
	}
	if err := yaml.Unmarshal(raw, &manifest); err != nil {
		return errors.Wrap(err, "parsing the chat engine manifest")
	}
	displayName := manifest.Spec.DisplayName
	if displayName == "" || !strings.Contains(displayName, tc.RunID) {
		return errors.Errorf("the chat engine's displayName %q must carry the lane's run id, so the release finds this lane's agent and no other", displayName)
	}
	location := manifest.Spec.ChatEngineConfig.AgentCreationConfig.Location
	if location == "" {
		location = "global"
	}
	base := fmt.Sprintf("https://dialogflow.googleapis.com/v3/projects/%s/locations/%s/agents", h.services.Project, location)
	var listing struct {
		Agents []struct {
			Name        string `json:"name"`
			DisplayName string `json:"displayName"`
		} `json:"agents"`
	}
	if _, err := h.restGet(ctx, base+"?pageSize=1000", &listing); err != nil {
		return errors.Wrap(err, "listing the Dialogflow agents")
	}
	released := 0
	for _, agent := range listing.Agents {
		if agent.DisplayName != displayName {
			continue
		}
		if err := h.restDelete(ctx, "https://dialogflow.googleapis.com/v3/"+agent.Name); err != nil {
			return errors.Wrapf(err, "deleting the agent %s the chat engine left", agent.Name)
		}
		released++
	}
	lanef(tc, "  [release] the chat engine %q left %d Dialogflow agent(s); deleted", displayName, released)
	return nil
}

// restDelete deletes one Google REST resource with the harness's credential;
// an absent resource is already released.
func (h *Harness) restDelete(ctx context.Context, url string) error {
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	resp, err := h.services.RestClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		return errors.Errorf("DELETE %s: HTTP %d", url, resp.StatusCode)
	}
	return nil
}

// releaseDataStoreToolChatEngines deletes the chat engines in the default
// collection whose data stores are this lane's (a store id carrying the
// lane's engine-scoped run id) -- the engines Google made for the agent's
// data store tool -- so the fixture stores can be deleted after them.
func (h *Harness) releaseDataStoreToolChatEngines(ctx context.Context, tc *provider.KindTestContext) error {
	runID := runner.EngineScopedRunID(tc.RunID, tc.Engine)
	base := fmt.Sprintf("https://discoveryengine.googleapis.com/v1/projects/%s/locations/global/collections/default_collection/engines", h.services.Project)
	var listing struct {
		Engines []struct {
			Name         string   `json:"name"`
			DataStoreIds []string `json:"dataStoreIds"`
			SolutionType string   `json:"solutionType"`
		} `json:"engines"`
	}
	if _, err := h.restGet(ctx, base+"?pageSize=1000", &listing); err != nil {
		return errors.Wrap(err, "listing the Discovery Engine engines")
	}
	released := 0
	for _, engine := range listing.Engines {
		if engine.SolutionType != "SOLUTION_TYPE_CHAT" || !strings.Contains(engine.Name, "-chat-") {
			continue
		}
		ours := false
		for _, store := range engine.DataStoreIds {
			if strings.Contains(store, runID) {
				ours = true
			}
		}
		if !ours {
			continue
		}
		if err := h.restDelete(ctx, "https://discoveryengine.googleapis.com/v1/"+engine.Name); err != nil {
			return errors.Wrapf(err, "deleting the chat engine %s the data store tool left", engine.Name)
		}
		released++
	}
	lanef(tc, "  [release] the data store tool left %d chat engine(s) over this lane's stores; deleted", released)
	return nil
}

// releaseModelGardenModels deletes the models Vertex AI uploaded for this
// lane's deployments: every model in the deployment's location created since
// the lane's identity was provisioned that no endpoint serves any more. The
// scenario's exclusive class keeps any other Model Garden lane out of that
// window. A lane with no recorded window (its identity never provisioned)
// made nothing.
func (h *Harness) releaseModelGardenModels(ctx context.Context, tc *provider.KindTestContext, manifestPath string) error {
	h.mu.Lock()
	audit := h.laneAudits[manifestPath]
	h.mu.Unlock()
	if audit == nil {
		lanef(tc, "  [release] no lane window recorded, so no Model Garden model to release")
		return nil
	}
	raw, err := os.ReadFile(manifestPath)
	if err != nil {
		return errors.Wrap(err, "reading the Model Garden manifest")
	}
	var manifest struct {
		Spec struct {
			Location string `json:"location"`
		} `json:"spec"`
	}
	if err := yaml.Unmarshal(raw, &manifest); err != nil {
		return errors.Wrap(err, "parsing the Model Garden manifest")
	}
	location := manifest.Spec.Location
	if location == "" {
		return errors.New("the Model Garden manifest names no location")
	}
	base := fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/models", location, h.services.Project, location)
	type model struct {
		Name           string          `json:"name"`
		CreateTime     time.Time       `json:"createTime"`
		DeployedModels json.RawMessage `json:"deployedModels"`
	}
	// Vertex AI pages its model list at 100 at most.
	var models []model
	for pageToken := ""; ; {
		var page struct {
			Models        []model `json:"models"`
			NextPageToken string  `json:"nextPageToken"`
		}
		if _, err := h.restGet(ctx, base+"?pageSize=100&pageToken="+pageToken, &page); err != nil {
			return errors.Wrap(err, "listing the Vertex AI models")
		}
		models = append(models, page.Models...)
		if pageToken = page.NextPageToken; pageToken == "" {
			break
		}
	}
	released := 0
	for _, model := range models {
		if model.CreateTime.Before(audit.start) || len(model.DeployedModels) > 0 {
			continue
		}
		if err := h.restDelete(ctx, fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1/%s", location, model.Name)); err != nil {
			return errors.Wrapf(err, "deleting the model %s the deployment left", model.Name)
		}
		released++
	}
	lanef(tc, "  [release] the deployment left %d uploaded model(s) in the Model Registry; deleted", released)
	return nil
}

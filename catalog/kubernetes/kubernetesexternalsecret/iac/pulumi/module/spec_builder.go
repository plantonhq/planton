package module

import (
	kubernetesexternalsecretv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesexternalsecret/v1alpha1"
)

// buildExternalSecretSpec renders the typed spec into the
// external-secrets.io/v1 ExternalSecret CRD spec shape. The Terraform
// module renders the same shape through its locals (null-prune idiom) —
// keep field mappings in lockstep.
func buildExternalSecretSpec(locals *Locals) map[string]interface{} {
	spec := locals.Spec
	rendered := map[string]interface{}{}

	// ---- store reference --------------------------------------------------
	// The arm is the grain: a cluster store renders its kind, a namespaced
	// store renders only its name (SecretStore is the CRD's default).
	storeRef := map[string]interface{}{}
	switch store := spec.GetStoreRef().GetStoreType().(type) {
	case *kubernetesexternalsecretv1alpha1.KubernetesExternalSecretStoreRef_ClusterSecretStore:
		storeRef["name"] = store.ClusterSecretStore.GetName().GetValue()
		storeRef["kind"] = "ClusterSecretStore"
	case *kubernetesexternalsecretv1alpha1.KubernetesExternalSecretStoreRef_SecretStore:
		storeRef["name"] = store.SecretStore.GetName().GetValue()
	}
	rendered["secretStoreRef"] = storeRef

	// ---- refresh behavior ---------------------------------------------------
	if spec.RefreshInterval != nil {
		rendered["refreshInterval"] = spec.GetRefreshInterval()
	}
	if spec.RefreshPolicy != nil && spec.GetRefreshPolicy() != "" {
		rendered["refreshPolicy"] = spec.GetRefreshPolicy()
	}

	// ---- target Secret ---------------------------------------------------------
	// Always rendered: the materialized Secret's name is pinned to the
	// resolved locals.SecretName so the exported handle can never drift
	// from what the operator creates.
	target := map[string]interface{}{"name": locals.SecretName}
	if t := spec.GetTarget(); t != nil {
		if t.CreationPolicy != nil {
			target["creationPolicy"] = t.GetCreationPolicy()
		}
		if t.DeletionPolicy != nil {
			target["deletionPolicy"] = t.GetDeletionPolicy()
		}
		if t.GetImmutable() {
			target["immutable"] = true
		}
		if template := buildTemplate(t.GetTemplate()); template != nil {
			target["template"] = template
		}
	}
	rendered["target"] = target

	// ---- explicit entries ----------------------------------------------------------
	if len(spec.GetData()) > 0 {
		data := make([]interface{}, 0, len(spec.GetData()))
		for _, entry := range spec.GetData() {
			data = append(data, map[string]interface{}{
				"secretKey": entry.GetSecretKey(),
				"remoteRef": buildRemoteRef(entry.GetRemoteRef()),
			})
		}
		rendered["data"] = data
	}

	// ---- bulk pulls -------------------------------------------------------------------
	if len(spec.GetDataFrom()) > 0 {
		dataFrom := make([]interface{}, 0, len(spec.GetDataFrom()))
		for _, pull := range spec.GetDataFrom() {
			entry := map[string]interface{}{}
			if extract := pull.GetExtract(); extract != nil {
				entry["extract"] = buildRemoteRef(extract)
			}
			if find := pull.GetFind(); find != nil {
				findSpec := map[string]interface{}{}
				if find.GetPath() != "" {
					findSpec["path"] = find.GetPath()
				}
				if find.GetNameRegexp() != "" {
					findSpec["name"] = map[string]interface{}{"regexp": find.GetNameRegexp()}
				}
				if len(find.GetTags()) > 0 {
					tags := map[string]interface{}{}
					for k, v := range find.GetTags() {
						tags[k] = v
					}
					findSpec["tags"] = tags
				}
				entry["find"] = findSpec
			}
			if len(pull.GetRewrite()) > 0 {
				rewrites := make([]interface{}, 0, len(pull.GetRewrite()))
				for _, rewrite := range pull.GetRewrite() {
					rewrites = append(rewrites, map[string]interface{}{
						"regexp": map[string]interface{}{
							"source": rewrite.GetSource(),
							"target": rewrite.GetTarget(),
						},
					})
				}
				entry["rewrite"] = rewrites
			}
			dataFrom = append(dataFrom, entry)
		}
		rendered["dataFrom"] = dataFrom
	}

	return rendered
}

// buildRemoteRef renders one backend-entry address.
func buildRemoteRef(ref *kubernetesexternalsecretv1alpha1.KubernetesExternalSecretRemoteRef) map[string]interface{} {
	rendered := map[string]interface{}{"key": ref.GetKey()}
	if ref.GetProperty() != "" {
		rendered["property"] = ref.GetProperty()
	}
	if ref.GetVersion() != "" {
		rendered["version"] = ref.GetVersion()
	}
	if ref.DecodingStrategy != nil && ref.GetDecodingStrategy() != "None" {
		rendered["decodingStrategy"] = ref.GetDecodingStrategy()
	}
	return rendered
}

// buildTemplate renders the materialized-Secret template. Returns nil when
// nothing is set.
func buildTemplate(template *kubernetesexternalsecretv1alpha1.KubernetesExternalSecretTemplate) map[string]interface{} {
	if template == nil {
		return nil
	}
	rendered := map[string]interface{}{}
	if template.GetType() != "" {
		rendered["type"] = template.GetType()
	}
	if template.MergePolicy != nil && template.GetMergePolicy() != "Replace" {
		rendered["mergePolicy"] = template.GetMergePolicy()
	}
	metadata := map[string]interface{}{}
	if len(template.GetLabels()) > 0 {
		labels := map[string]interface{}{}
		for k, v := range template.GetLabels() {
			labels[k] = v
		}
		metadata["labels"] = labels
	}
	if len(template.GetAnnotations()) > 0 {
		annotations := map[string]interface{}{}
		for k, v := range template.GetAnnotations() {
			annotations[k] = v
		}
		metadata["annotations"] = annotations
	}
	if len(metadata) > 0 {
		rendered["metadata"] = metadata
	}
	if len(template.GetData()) > 0 {
		data := map[string]interface{}{}
		for k, v := range template.GetData() {
			data[k] = v
		}
		rendered["data"] = data
	}
	if len(rendered) == 0 {
		return nil
	}
	return rendered
}

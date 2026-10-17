package module

import (
	"github.com/pkg/errors"
)

// upstreamFiles are the pinned files the composition reads: the release
// asset the extensions setting selects, and the router's deployment files
// (only when a router is installed).
type upstreamFiles struct {
	release string
	router  map[string]string
}

// composition is the document set the module applies besides the kept
// definitions: the Namespace when this resource owns it (nil otherwise),
// applied first and deleted last, and everything else, applied once the
// definitions exist. Every object is upstream's own with the typed settings
// applied, plus the module's ConfigMap.
type composition struct {
	namespace document
	workloads []document
}

// routerFilesFor names the router files the settings apply, in apply order.
func routerFilesFor(s *settings) []string {
	if s.router == nil {
		return nil
	}
	var files []string
	for _, file := range routerFiles {
		switch {
		case file == routerFileRBAC && !s.router.cacheEnabled:
		case file == routerFileTokenReview && s.router.authzMode != "tokenreview":
		case file == routerFilePDB && s.router.minAvailable == 0:
		case file == routerFileNetworkPolicy && s.router.networkPolicy == nil:
		default:
			files = append(files, file)
		}
	}
	return files
}

// compose builds the document set from the pinned files. The definitions are
// dropped here: the kept-CRD primitive applies them from the same release
// asset, kept and ownership-checked.
func compose(s *settings, files upstreamFiles) (*composition, error) {
	releaseDocuments, err := splitDocuments(files.release)
	if err != nil {
		return nil, err
	}
	result := &composition{}
	var namespace, controllerDeployment, controllerService bool
	for _, d := range releaseDocuments {
		switch str(d, "kind") {
		case "CustomResourceDefinition":
			continue
		case "Namespace":
			namespace = true
			if s.createNamespace {
				labels := child(child(d, "metadata"), "labels")
				for key, value := range s.identityLabels {
					labels[key] = value
				}
				result.namespace = d
			}
			continue
		case "Deployment":
			if str(d, "metadata", "name") == vars.ControllerDeploymentName {
				if err := patchControllerDeployment(d, s); err != nil {
					return nil, err
				}
				controllerDeployment = true
			}
		case "Service":
			if str(d, "metadata", "name") == vars.ControllerServiceName {
				patchControllerService(d, s)
				controllerService = true
			}
		}
		result.workloads = append(result.workloads, d)
	}
	if !namespace || !controllerDeployment || !controllerService {
		return nil, errors.Errorf("upstream's release asset at %s no longer carries the %s Namespace, Deployment and Service the module patches; the module's pin needs a catalog fix", vars.Release, vars.Namespace)
	}

	if s.extensions && len(s.allowedLabelDomains) > 0 {
		result.workloads = append(result.workloads, configMap(s.allowedLabelDomains))
	}

	for _, file := range routerFilesFor(s) {
		text, ok := files.router[file]
		if !ok {
			return nil, errors.Errorf("the router file %s was not fetched", file)
		}
		documents, err := splitDocuments(text)
		if err != nil {
			return nil, err
		}
		for _, d := range documents {
			if err := patchRouterDocument(d, file, s); err != nil {
				return nil, err
			}
			result.workloads = append(result.workloads, d)
		}
	}
	return result, nil
}

// patchRouterDocument applies the settings that belong to one router file.
func patchRouterDocument(d document, file string, s *settings) error {
	switch file {
	case routerFileDeployment:
		return patchRouterDeployment(d, s)
	case routerFileService:
		patchRouterService(d, s)
	case routerFilePDB:
		patchRouterPDB(d, s)
	case routerFileNetworkPolicy:
		return patchRouterNetworkPolicy(d, s)
	}
	return nil
}

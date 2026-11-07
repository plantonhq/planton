package helmcrds

import (
	"context"
	"io"
	"net/http"
	"time"
)

// bundleFetchTimeout bounds one bundle download. Upstream CRD bundles run to a
// few megabytes (Solr's is ~2 MB); a minute is generous without letting a
// hung host stall a plan indefinitely.
const bundleFetchTimeout = 60 * time.Second

// fetchBundle downloads the pinned upstream CRD bundle, every URL in order,
// and splits it into documents. The bundle branch exists for upstreams that
// publish their CRDs outside a chart: beside a chart that ships none, in an
// install manifest, or one file per CRD.
func fetchBundle(ctx context.Context, src Source) ([]string, error) {
	var documents []string
	for _, url := range src.ResolvedBundleURLs() {
		fetched, err := fetchBundleURL(ctx, src, url)
		if err != nil {
			return nil, err
		}
		documents = append(documents, fetched...)
	}
	return documents, nil
}

// fetchBundleURL downloads one of the bundle's manifests.
func fetchBundleURL(ctx context.Context, src Source, url string) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, bundleFetchTimeout)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, bundleFetchFailure(src, url, 0, err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, bundleFetchFailure(src, url, 0, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, bundleFetchFailure(src, url, response.StatusCode, nil)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, bundleFetchFailure(src, url, 0, err)
	}
	return splitDocuments(string(body)), nil
}
